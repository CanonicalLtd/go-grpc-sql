package protocol

import (
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

// Create a new Request with the given payload.
func newRequest(message proto.Message) *Request {
	var code RequestCode
	switch message.(type) {
	case *RequestOpen:
		code = RequestCode_OPEN
	case *RequestPrepare:
		code = RequestCode_PREPARE
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

// Create a new Response with the given payload.
func newResponse(message proto.Message) *Response {
	var code RequestCode
	switch message.(type) {
	case *ResponseOpen:
		code = RequestCode_OPEN
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

// NewStatement creates a new Statement object with the given SQL text and
// arguments.
//
// It returns an error if the given arguments contain one or more values of
// unsupported type.
func NewStatement(text string, args []interface{}) (*Statement, error) {
	values, err := toValueSlice(args)
	if err != nil {
		return nil, err
	}

	stmt := &Statement{
		Text: text,
		Args: values,
	}
	return stmt, nil
}

// UnmarshalArgs deserializes the arguments of a SQL statement.
func (s *Statement) UnmarshalArgs() ([]interface{}, error) {
	return fromValueSlice(s.Args)
}

// Convert a slice of Go objects of supported types to a slice of protobuf
// Value objects.
func toValueSlice(objects []interface{}) ([]*Value, error) {
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

// Convert a slice of protobuf Value objects to a slice of Go interface{} objects.
func fromValueSlice(values []*Value) ([]interface{}, error) {
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
