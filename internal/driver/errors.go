package driver

import (
	"database/sql/driver"
	"io"
	"reflect"
	"unsafe"

	"github.com/CanonicalLtd/go-grpc-sql/internal/gateway"
	"github.com/CanonicalLtd/go-grpc-sql/internal/sql"
	"github.com/CanonicalLtd/go-sqlite3"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Create a new instance of sqlite3.Error, the only backend currently
// supported.
func newError(code sqlite3.ErrNo, extendedCode sqlite3.ErrNoExtended, err string) sqlite3.Error {
	e := sqlite3.Error{
		Code:         code,
		ExtendedCode: extendedCode,
	}

	// FIXME: unfortunately the err attribute of sqlite3.Error is private,
	// so it's not possible to instantiate a sqlite3.Error with a custom
	// error string, so we need to use reflect to set the private field.
	//
	// See https://play.golang.org/p/DXXHuIvRLI
	rs := reflect.ValueOf(&e).Elem() // e, but writable

	// Assumes that e.err is the third field of sqlite3.Error.
	rf := rs.Field(2) // e.err

	ri := reflect.ValueOf(&err).Elem() // err, but writeable

	// Cheat
	rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
	rf.Set(ri)

	return e
}

// Convenience to return a SQLITE_MISUSE error
func newErrorMisuse(err error) sqlite3.Error {
	return newError(sqlite3.ErrMisuse, 0, err.Error())
}

// Convert an error occurred upon invoking gRPC method into either
// sqlite3.Error (the only backend currently supported) or a driver.ErrBadConn
// if it's a fatal error.
func newErrorFromMethod(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		// This is unexpected because the gateway should always return
		// proper status errors.
		if errors.Cause(err) == io.EOF {
			// FIXME: This behavior is pasted here from v1 of the
			//        gRPC SQL client. We should check if it still
			//        makes sense.
			return driver.ErrBadConn
		}

		// Fallback to a generic error.
		return newError(sqlite3.ErrError, 0, err.Error())
	}
	switch st.Code() {
	case gateway.DriverFailed:
		// The underlying gateway driver returned an error.
		details := st.Details()
		if len(details) == 0 {
			// This is unexpected because the gateway should always
			// return details about a driver error. Fallback to a
			// generic error.
			return newError(sqlite3.ErrError, 0, err.Error())
		}
		detail, ok := details[0].(*sql.Error)
		if !ok {
			// This is unexpected because the gateway should always
			// return a sql.Error as first and only status
			// detail. Fallback to a generic error.
			return newError(sqlite3.ErrError, 0, err.Error())
		}

		// If the failure is due to the driver not being the leader
		// anymore, return ErrBadConn.
		switch sqlite3.ErrNoExtended(detail.ExtendedCode) {
		case sqlite3.ErrIoErrNotLeader:
			fallthrough
		case sqlite3.ErrIoErrLeadershipLost:
			return driver.ErrBadConn
		}

		return newError(
			sqlite3.ErrNo(detail.Code),
			sqlite3.ErrNoExtended(detail.ExtendedCode),
			detail.Message,
		)
	case codes.Canceled:
		// FIXME: Investiage a way to:
		//
		// 1) Soundly detect if Canceled can only be result of the
		//    context of an gRPC request being cancelled.
		//
		// 2) Make the cancelling caller block until the gateway can
		//    handle the cancellation, run sqlite3_interrupt and return
		//    a proper error.
		//
		// For now we just return SQLITE_INTERRUPT
		return newError(sqlite3.ErrInterrupt, 0, err.Error())
	case codes.Unavailable:
		// The gateway is down.
		return driver.ErrBadConn
	case codes.Internal:
		// FIXME: Investigate under which conditions the grpc package
		//        might return this code, since it should never be
		//        returned by the gateway service.
		if st.Message() == grpcNoErrorMessage {
			// FIXME: This look like a spurious error which gets generated
			//        only by the http test server of Go >= 1.9. Still, we
			//        handle it in this ad-hoc way because when it happens
			//        it simply means that the other end is down.
			return driver.ErrBadConn
		}
		return newError(sqlite3.ErrInternal, 0, err.Error())
	case codes.ResourceExhausted:
		// To many connections, transactions, statements or result sets
		// are open.
		return newError(sqlite3.ErrMisuse, 0, err.Error())
	case codes.Unimplemented:
		// This should never be returned by a dqlite driver, since it
		// implements Leader()/Recover(), but we include this case for
		// good measure.
		return newErrorMisuse(err)
	default:
		// Fallback to a generic error.
		return newError(sqlite3.ErrError, 0, err.Error())
	}
}

const grpcNoErrorMessage = "stream terminated by RST_STREAM with error code: NO_ERROR"
