package protocol

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"time"

	"github.com/golang/protobuf/proto"
)

// NewRequestOpen creates a new Request of type RequestOpen.
func NewRequestOpen(name string) *Request {
	return newRequest(&RequestOpen{Name: name})
}

// NewRequestPrepare creates a new Request of type RequestPrepare.
func NewRequestPrepare(query string) *Request {
	return newRequest(&RequestPrepare{Query: query})
}

// NewRequestExec creates a new Request of type RequestExec.
func NewRequestExec(id int64, args []*Value) *Request {
	return newRequest(&RequestExec{Id: id, Args: args})
}

// NewRequestStmtClose creates a new Request of type RequestStmtClose.
func NewRequestStmtClose(id int64) *Request {
	return newRequest(&RequestStmtClose{Id: id})
}

// NewRequestClose creates a new Request of type RequestClose.
func NewRequestClose() *Request {
	return newRequest(&RequestClose{})
}

// Create a new Request with the given payload.
func newRequest(message proto.Message) *Request {
	var code RequestCode
	switch message.(type) {
	case *RequestOpen:
		code = RequestCode_OPEN
	case *RequestPrepare:
		code = RequestCode_PREPARE
	case *RequestExec:
		code = RequestCode_EXEC
	case *RequestStmtClose:
		code = RequestCode_STMT_CLOSE
	case *RequestClose:
		code = RequestCode_CLOSE
	default:
		panic(fmt.Errorf("invalid message type: %s", reflect.TypeOf(message).Kind()))
	}

	data, err := proto.Marshal(message)
	if err != nil {
		panic(fmt.Errorf("cannot marshal %s request", code))
	}

	request := &Request{
		Code: code,
		Data: data,
	}

	return request
}

// NewResponseOpen creates a new Response of type ResponseOpen.
func NewResponseOpen() *Response {
	return newResponse(&ResponseOpen{})
}

// NewResponsePrepare creates a new Response of type ResponsePrepare.
func NewResponsePrepare(id int64, numInput int) *Response {
	return newResponse(&ResponsePrepare{Id: id, NumInput: int64(numInput)})
}

// NewResponseExec creates a new Response of type ResponseExec.
func NewResponseExec(lastInsertID, rowsAffected int64) *Response {
	return newResponse(&ResponseExec{
		LastInsertId: lastInsertID,
		RowsAffected: rowsAffected,
	})
}

// NewResponseStmtClose creates a new Response of type ResponseStmtClose.
func NewResponseStmtClose() *Response {
	return newResponse(&ResponseStmtClose{})
}

// NewResponseClose creates a new Response of type ResponseClose.
func NewResponseClose() *Response {
	return newResponse(&ResponseClose{})
}

// Prepare returns a ResponsePrepare payload.
func (r *Response) Prepare() *ResponsePrepare {
	message := &ResponsePrepare{}
	r.unmarshal(message)
	return message
}

// Exec returns a ResponseExec payload.
func (r *Response) Exec() *ResponseExec {
	message := &ResponseExec{}
	r.unmarshal(message)
	return message
}

func (r *Response) unmarshal(message proto.Message) {
	if err := proto.Unmarshal(r.Data, message); err != nil {
		panic(fmt.Errorf("failed to unmarshal response: %v", err))
	}
}

// Create a new Response with the given payload.
func newResponse(message proto.Message) *Response {
	var code RequestCode
	switch message.(type) {
	case *ResponseOpen:
		code = RequestCode_OPEN
	case *ResponsePrepare:
		code = RequestCode_PREPARE
	case *ResponseExec:
		code = RequestCode_EXEC
	case *ResponseStmtClose:
		code = RequestCode_STMT_CLOSE
	case *ResponseClose:
		code = RequestCode_CLOSE
	default:
		panic(fmt.Errorf("invalid message type"))
	}

	data, err := proto.Marshal(message)
	if err != nil {
		panic(fmt.Errorf("cannot marshal %s response", code))
	}

	response := &Response{
		Code: code,
		Data: data,
	}

	return response
}

// FromDriverValues converts a slice of Go driver.Value objects of supported
// types to a slice of protobuf Value objects.
func FromDriverValues(objects []driver.Value) ([]*Value, error) {
	values := make([]*Value, len(objects))
	for i, object := range objects {
		value, err := toValue(object)
		if err != nil {
			return nil, fmt.Errorf("cannot marshal object %d (%v): %s", i, object, err)
		}
		values[i] = value
	}
	return values, nil
}

// ToValueSlice converts a slice of Go objects of supported types to a slice of
// protobuf Value objects.
func ToValueSlice(objects []interface{}) ([]*Value, error) {
	values := make([]*Value, len(objects))
	for i, object := range objects {
		value, err := toValue(object)
		if err != nil {
			return nil, fmt.Errorf("cannot marshal object %d (%v): %s", i, object, err)
		}
		values[i] = value
	}
	return values, nil
}

// Convert a Go object of a supported type to a protobuf Value object.
func toValue(value interface{}) (*Value, error) {
	var code ValueCode
	var message proto.Message

	switch v := value.(type) {
	case int64:
		code = ValueCode_INT64
		message = &ValueInt64{Value: v}
	case float64:
		code = ValueCode_FLOAT64
		message = &ValueFloat64{Value: v}
	case bool:
		code = ValueCode_BOOL
		message = &ValueBool{Value: v}
	case []byte:
		code = ValueCode_BYTES
		message = &ValueBytes{Value: v}
	case string:
		code = ValueCode_STRING
		message = &ValueString{Value: v}
	case time.Time:
		code = ValueCode_TIME
		message = &ValueTime{Value: v.Unix()}
	default:
		if value != nil {
			return nil, fmt.Errorf("invalid type %s", reflect.TypeOf(value).Kind())
		}
		code = ValueCode_NULL
		message = &ValueNull{}
	}

	data, err := proto.Marshal(message)
	if err != nil {
		return nil, err
	}
	return &Value{Code: code, Data: data}, nil
}

// ToDriverValues converts a slice of protobuf Value objects to a slice of Go
// driver.Value objects.
func ToDriverValues(values []*Value) ([]driver.Value, error) {
	args, err := FromValueSlice(values)
	if err != nil {
		return nil, err
	}
	a := make([]driver.Value, len(args))
	for i, arg := range args {
		a[i] = arg
	}
	return a, nil
}

// FromValueSlice converts a slice of protobuf Value objects to a slice of Go
// interface{} objects.
func FromValueSlice(values []*Value) ([]interface{}, error) {
	objects := make([]interface{}, len(values))
	for i, value := range values {
		object, err := fromValue(value)
		if err != nil {
			return nil, fmt.Errorf("cannot unmarshal value %d: %s", i, err)
		}
		objects[i] = object
	}
	return objects, nil
}

// Convert a protobuf Value object to a Go interface object.
func fromValue(value *Value) (interface{}, error) {
	var message valueMessage
	switch value.Code {
	case ValueCode_INT64:
		message = &ValueInt64{}
	case ValueCode_FLOAT64:
		message = &ValueFloat64{}
	case ValueCode_BOOL:
		message = &ValueBool{}
	case ValueCode_BYTES:
		message = &ValueBytes{}
	case ValueCode_STRING:
		message = &ValueString{}
	case ValueCode_TIME:
		message = &ValueTime{}
	case ValueCode_NULL:
		message = &ValueNull{}
	default:
		return nil, fmt.Errorf("invalid value type code %d", value.Code)
	}

	err := proto.Unmarshal(value.Data, message)
	if err != nil {
		return nil, err
	}

	return message.Interface(), nil
}

// Interface implemented by the various ValueXXX objects, that returns the
// underlying value as interface{}.
type valueMessage interface {
	proto.Message
	Interface() interface{}
}

// Interface implements valueMessage.
func (v *ValueInt64) Interface() interface{} {
	return v.Value
}

// Interface implements valueMessage.
func (v *ValueFloat64) Interface() interface{} {
	return v.Value
}

// Interface implements valueMessage.
func (v *ValueBool) Interface() interface{} {
	return v.Value
}

// Interface implements valueMessage.
func (v *ValueBytes) Interface() interface{} {
	return v.Value
}

// Interface implements valueMessage.
func (v *ValueString) Interface() interface{} {
	return v.Value
}

// Interface implements valueMessage.
func (v *ValueTime) Interface() interface{} {
	return time.Unix(v.Value, 0)
}

// Interface implements valueMessage.
func (v *ValueNull) Interface() interface{} {
	return nil
}
