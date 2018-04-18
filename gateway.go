package grpcsql

import (
	"database/sql/driver"
	"fmt"
	"io"
	"sync"

	"github.com/CanonicalLtd/go-grpc-sql/internal/legacy"
	"github.com/CanonicalLtd/go-sqlite3"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

// Gateway mapping gRPC requests to SQL queries.
type Gateway struct {
	driver driver.Driver // Underlying SQL driver.
}

// NewGateway creates a new gRPC gateway executing requests against the given
// SQL driver.
func NewGateway(drv driver.Driver) *Gateway {
	return &Gateway{
		driver: drv,
	}
}

// Conn creates a new database connection using the underlying driver, and
// start accepting requests for it.
func (s *Gateway) Conn(stream legacy.SQL_ConnServer) error {
	var conn *gatewayConn
	for {
		request, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return errors.Wrapf(err, "failed to receive request")
		}

		if conn == nil {
			conn = &gatewayConn{
				driver: s.driver,
			}
			defer conn.Close()
		}

		response, err := conn.handle(request)
		if err != nil {
			// TODO: add support for more driver-specific errors.
			switch err := err.(type) {
			case sqlite3.Error:
				response = legacy.NewResponseSQLiteError(
					err.Code, err.ExtendedCode, err.Error())
			case sqlite3.ErrNo:
				response = legacy.NewResponseSQLiteError(
					err, 0, err.Error())
			default:
				return errors.Wrapf(err, "failed to handle %s request", request.Code)
			}
		}

		if err := stream.Send(response); err != nil {
			return errors.Wrapf(err, "failed to send response")
		}
	}
}

// Track a single driver connection
type gatewayConn struct {
	driver     driver.Driver
	driverConn driver.Conn
	stmts      map[int64]driver.Stmt
	txs        map[int64]driver.Tx
	rows       map[int64]driver.Rows
	serial     int64
	mu         sync.Mutex
}

// Handle a single gRPC request for this connection.
func (c *gatewayConn) handle(request *legacy.Request) (*legacy.Response, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var message proto.Message

	switch request.Code {
	case legacy.RequestCode_OPEN:
		message = &legacy.RequestOpen{}
	case legacy.RequestCode_PREPARE:
		message = &legacy.RequestPrepare{}
	case legacy.RequestCode_EXEC:
		message = &legacy.RequestExec{}
	case legacy.RequestCode_QUERY:
		message = &legacy.RequestQuery{}
	case legacy.RequestCode_NEXT:
		message = &legacy.RequestNext{}
	case legacy.RequestCode_COLUMN_TYPE_SCAN_TYPE:
		message = &legacy.RequestColumnTypeScanType{}
	case legacy.RequestCode_COLUMN_TYPE_DATABASE_TYPE_NAME:
		message = &legacy.RequestColumnTypeDatabaseTypeName{}
	case legacy.RequestCode_ROWS_CLOSE:
		message = &legacy.RequestRowsClose{}
	case legacy.RequestCode_STMT_CLOSE:
		message = &legacy.RequestStmtClose{}
	case legacy.RequestCode_BEGIN:
		message = &legacy.RequestBegin{}
	case legacy.RequestCode_COMMIT:
		message = &legacy.RequestCommit{}
	case legacy.RequestCode_ROLLBACK:
		message = &legacy.RequestRollback{}
	case legacy.RequestCode_CLOSE:
		message = &legacy.RequestClose{}
	case legacy.RequestCode_CONN_EXEC:
		message = &legacy.RequestConnExec{}
	default:
		return nil, fmt.Errorf("invalid request code %d", request.Code)
	}

	if err := proto.Unmarshal(request.Data, message); err != nil {
		return nil, errors.Wrapf(err, "request parse error")
	}

	// The very first request must be an OPEN one.
	if c.driverConn == nil && request.Code != legacy.RequestCode_OPEN {
		return nil, fmt.Errorf("expected OPEN request, got %s", request.Code)
	}

	switch r := message.(type) {
	case *legacy.RequestOpen:
		return c.handleOpen(r)
	case *legacy.RequestPrepare:
		return c.handlePrepare(r)
	case *legacy.RequestExec:
		return c.handleExec(r)
	case *legacy.RequestQuery:
		return c.handleQuery(r)
	case *legacy.RequestNext:
		return c.handleNext(r)
	case *legacy.RequestColumnTypeScanType:
		return c.handleColumnTypeScanType(r)
	case *legacy.RequestColumnTypeDatabaseTypeName:
		return c.handleColumnTypeDatabaseTypeName(r)
	case *legacy.RequestRowsClose:
		return c.handleRowsClose(r)
	case *legacy.RequestStmtClose:
		return c.handleStmtClose(r)
	case *legacy.RequestBegin:
		return c.handleBegin(r)
	case *legacy.RequestCommit:
		return c.handleCommit(r)
	case *legacy.RequestRollback:
		return c.handleRollback(r)
	case *legacy.RequestClose:
		return c.handleClose(r)
	case *legacy.RequestConnExec:
		return c.handleConnExec(r)
	default:
		panic("unhandled request payload type")
	}
}

// Close the underlying driver connection, if any.
func (c *gatewayConn) Close() error {
	if c.driverConn != nil {
		return c.driverConn.Close()
	}
	return nil
}

// Handle a request of type OPEN.
func (c *gatewayConn) handleOpen(request *legacy.RequestOpen) (*legacy.Response, error) {
	driverConn, err := c.driver.Open(request.Name)
	if err != nil {
		return nil, fmt.Errorf("could not open driver connection: %v", err)
	}

	c.driverConn = driverConn
	c.stmts = make(map[int64]driver.Stmt)
	c.txs = make(map[int64]driver.Tx)
	c.rows = make(map[int64]driver.Rows)

	response := legacy.NewResponseOpen()
	return response, nil
}

// Handle a request of type PREPARE.
func (c *gatewayConn) handlePrepare(request *legacy.RequestPrepare) (*legacy.Response, error) {
	driverStmt, err := c.driverConn.Prepare(request.Query)
	if err != nil {
		return nil, err
	}
	c.serial++
	c.stmts[c.serial] = driverStmt
	return legacy.NewResponsePrepare(c.serial, driverStmt.NumInput()), nil
}

// Handle a request of type EXEC.
func (c *gatewayConn) handleExec(request *legacy.RequestExec) (*legacy.Response, error) {
	driverStmt, ok := c.stmts[request.Id]
	if !ok {
		return nil, fmt.Errorf("no prepared statement with ID %d", request.Id)
	}

	args, err := legacy.ToDriverValues(request.Args)
	if err != nil {
		return nil, err
	}

	result, err := driverStmt.Exec(args)
	if err != nil {
		return nil, err
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	response := legacy.NewResponseExec(lastInsertID, rowsAffected)

	return response, nil
}

// Handle a request of type QUERY.
func (c *gatewayConn) handleQuery(request *legacy.RequestQuery) (*legacy.Response, error) {
	driverStmt, ok := c.stmts[request.Id]
	if !ok {
		return nil, fmt.Errorf("no prepared statement with ID %d", request.Id)
	}

	args, err := legacy.ToDriverValues(request.Args)
	if err != nil {
		return nil, err
	}

	driverRows, err := driverStmt.Query(args)
	if err != nil {
		return nil, err
	}

	c.serial++
	c.rows[c.serial] = driverRows

	return legacy.NewResponseQuery(c.serial, driverRows.Columns()), nil
}

// Handle a request of type NEXT.
func (c *gatewayConn) handleNext(request *legacy.RequestNext) (*legacy.Response, error) {
	driverRows, ok := c.rows[request.Id]
	if !ok {
		return nil, fmt.Errorf("no rows with ID %d", request.Id)
	}

	dest := make([]driver.Value, int(request.Len))
	err := driverRows.Next(dest)
	if err == io.EOF {
		return legacy.NewResponseNext(true, nil), nil
	}

	if err != nil {
		return nil, err
	}

	values, err := legacy.FromDriverValues(dest)
	if err != nil {
		return nil, err
	}

	return legacy.NewResponseNext(false, values), nil
}

// Handle a request of type COLUMN_TYPE_SCAN_TYPE.
func (c *gatewayConn) handleColumnTypeScanType(request *legacy.RequestColumnTypeScanType) (*legacy.Response, error) {
	driverRows, ok := c.rows[request.Id]
	if !ok {
		return nil, fmt.Errorf("no rows with ID %d", request.Id)
	}

	code := legacy.ValueCode_BYTES
	typeScanner, ok := driverRows.(driver.RowsColumnTypeScanType)
	if ok {
		typ := typeScanner.ColumnTypeScanType(int(request.Column))
		code = legacy.ToValueCode(typ)
	}

	return legacy.NewResponseColumnTypeScanType(code), nil
}

// Handle a request of type COLUMN_TYPE_DATABASE_TYPE_NAME.
func (c *gatewayConn) handleColumnTypeDatabaseTypeName(request *legacy.RequestColumnTypeDatabaseTypeName) (*legacy.Response, error) {
	driverRows, ok := c.rows[request.Id]
	if !ok {
		return nil, fmt.Errorf("no rows with ID %d", request.Id)
	}

	name := ""
	nameScanner, ok := driverRows.(driver.RowsColumnTypeDatabaseTypeName)
	if ok {
		name = nameScanner.ColumnTypeDatabaseTypeName(int(request.Column))
	}

	return legacy.NewResponseColumnTypeDatabaseTypeName(name), nil
}

// Handle a request of type ROWS_CLOSE.
func (c *gatewayConn) handleRowsClose(request *legacy.RequestRowsClose) (*legacy.Response, error) {
	driverRows, ok := c.rows[request.Id]
	if !ok {
		return nil, fmt.Errorf("no rows with ID %d", request.Id)
	}
	delete(c.rows, request.Id)

	if err := driverRows.Close(); err != nil {
		return nil, err
	}

	response := legacy.NewResponseRowsClose()
	return response, nil
}

// Handle a request of type STMT_CLOSE.
func (c *gatewayConn) handleStmtClose(request *legacy.RequestStmtClose) (*legacy.Response, error) {
	driverStmt, ok := c.stmts[request.Id]
	if !ok {
		return nil, fmt.Errorf("no prepared statement with ID %d", request.Id)
	}
	delete(c.stmts, request.Id)

	if err := driverStmt.Close(); err != nil {
		return nil, err
	}

	response := legacy.NewResponseStmtClose()
	return response, nil
}

// Handle a request of type BEGIN.
func (c *gatewayConn) handleBegin(request *legacy.RequestBegin) (*legacy.Response, error) {
	driverTx, err := c.driverConn.Begin()
	if err != nil {
		return nil, err
	}
	c.serial++
	c.txs[c.serial] = driverTx
	return legacy.NewResponseBegin(c.serial), nil
}

// Handle a request of type COMMIT.
func (c *gatewayConn) handleCommit(request *legacy.RequestCommit) (*legacy.Response, error) {
	driverTx, ok := c.txs[request.Id]
	if !ok {
		return nil, fmt.Errorf("no transaction with ID %d", request.Id)
	}
	delete(c.txs, request.Id)

	if err := driverTx.Commit(); err != nil {
		return nil, err
	}

	response := legacy.NewResponseCommit()
	return response, nil
}

// Handle a request of type ROLLBACK.
func (c *gatewayConn) handleRollback(request *legacy.RequestRollback) (*legacy.Response, error) {
	driverTx, ok := c.txs[request.Id]
	if !ok {
		return nil, fmt.Errorf("no transaction with ID %d", request.Id)
	}
	delete(c.txs, request.Id)

	if err := driverTx.Rollback(); err != nil {
		return nil, err
	}

	response := legacy.NewResponseRollback()
	return response, nil
}

// Handle a request of type CLOSE.
func (c *gatewayConn) handleClose(request *legacy.RequestClose) (*legacy.Response, error) {
	if err := c.driverConn.Close(); err != nil {
		return nil, err
	}

	response := legacy.NewResponseClose()
	return response, nil
}

// Handle a request of type CONN_EXEC.
func (c *gatewayConn) handleConnExec(request *legacy.RequestConnExec) (*legacy.Response, error) {
	args, err := legacy.ToDriverValues(request.Args)
	if err != nil {
		return nil, err
	}

	execer, ok := c.driverConn.(driver.Execer)
	if !ok {
		return nil, fmt.Errorf("backend driver does not implement driver.Execer")
	}
	result, err := execer.Exec(request.Query, args)
	if err != nil {
		return nil, err
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	response := legacy.NewResponseExec(lastInsertID, rowsAffected)

	return response, nil
}
