package sql

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"time"

	"github.com/golang/protobuf/proto"
)

// ConnID identifies a connection in a gRPC SQL gateway service.
type ConnID uint32

// StmtID identifies a prepared statement in a connection.
type StmtID uint32

// TxID identifies a transaction in a connection.
type TxID uint32

// RowsID identifies a result set in a connection.
type RowsID uint32

// WithConn is implemented by messages that target a specific connection.
type WithConn interface {
	GetConnId() ConnID
}

// NewClient creates a new Client object with the given ID.
func NewClient(id string) *Client {
	return &Client{
		Id: id,
	}
}

// NewDuration creates a new Duration object with the given value.
func NewDuration(duration time.Duration) *Duration {
	return &Duration{
		Value: int64(duration),
	}
}

// NewServer creates a new Server object with the given address.
func NewServer(address string) *Server {
	return &Server{
		Address: address,
	}
}

// NewCluster creates a new Cluster formed by the given servers.
func NewCluster(servers []*Server) *Cluster {
	return &Cluster{
		Servers: servers,
	}
}

// NewDatabase creates a new Database object.
func NewDatabase(clientID string, name string) *Database {
	return &Database{
		ClientId: clientID,
		Name:     name,
	}
}

// NewConn creates a new Conn object.
func NewConn(id ConnID) *Conn {
	return &Conn{
		Id: id,
	}
}

// GetConnId is added to have conn implement the WithConn interface.
func (c *Conn) GetConnId() ConnID {
	return c.Id
}

// NewSQL creates a new SQL object.
func NewSQL(conn *Conn, text string) *SQL {
	return &SQL{
		ConnId: conn.Id,
		Text:   text,
	}
}

// NewStmt creates a new Stmt object.
func NewStmt(connID ConnID, id StmtID, numInput int32) *Stmt {
	return &Stmt{
		ConnId:   connID,
		Id:       id,
		NumInput: numInput,
	}
}

// NewTx creates a new Tx object.
func NewTx(connID ConnID, id TxID) *Tx {
	return &Tx{
		ConnId: connID,
		Id:     id,
	}
}

// NewBoundStmt creates a new BoundStmt object.
func NewBoundStmt(stmt *Stmt, driverNamedValues []driver.NamedValue) (*BoundStmt, error) {
	namedValues, err := fromDriverNamedValueSlice(driverNamedValues)
	if err != nil {
		return nil, err
	}

	boundStmt := &BoundStmt{
		ConnId:      stmt.ConnId,
		StmtId:      stmt.Id,
		NamedValues: namedValues,
	}

	return boundStmt, nil
}

// DriverNamedValues converts the internal protobuf NamedValue slice into a
// driver.NamedValue slice.
func (s *BoundStmt) DriverNamedValues() ([]driver.NamedValue, error) {
	return toDriverNamedValueSlice(s.NamedValues)
}

// NewBoundSQL creates a new BoundSQL object.
func NewBoundSQL(conn *Conn, text string, driverNamedValues []driver.NamedValue) (*BoundSQL, error) {
	namedValues, err := fromDriverNamedValueSlice(driverNamedValues)
	if err != nil {
		return nil, err
	}

	boundSQL := &BoundSQL{
		ConnId:      conn.Id,
		Text:        text,
		NamedValues: namedValues,
	}

	return boundSQL, nil
}

// DriverNamedValues converts the internal protobuf NamedValue slice into a
// driver.NamedValue slice.
func (s *BoundSQL) DriverNamedValues() ([]driver.NamedValue, error) {
	return toDriverNamedValueSlice(s.NamedValues)
}

// NewResult creates a new Result object.
func NewResult(lastInsertedID int64, rowsAffected int64) *Result {
	return &Result{
		LastInsertId: lastInsertedID,
		RowsAffected: rowsAffected,
	}
}

// NewRows creates a new Rows object.
func NewRows(connID ConnID, id RowsID, columns []string) *Rows {
	return &Rows{
		ConnId:  connID,
		Id:      id,
		Columns: columns,
	}
}

// NewRow creates a new Row object with the given column values.
func NewRow(driverColumns []driver.Value) (*Row, error) {
	var columns []*Value
	if driverColumns != nil {
		var err error
		columns, err = fromDriverValueSlice(driverColumns)
		if err != nil {
			return nil, err
		}
	}

	row := &Row{
		Columns: columns,
	}

	return row, nil
}

// NewColumn creates a new Column referencing the given Rows.
func NewColumn(rows *Rows, index int) *Column {
	return &Column{
		ConnId: rows.ConnId,
		RowsId: rows.Id,
		Index:  int32(index),
	}
}

// DriverColumns converts the internal protobuf Value slice into a
// driver.Value slice.
func (r *Row) DriverColumns() ([]driver.Value, error) {
	if r.Columns == nil {
		return nil, nil
	}
	return toDriverValueSlice(r.Columns)
}

// NewType creates a new Type with a code matching the given Go type object.
func NewType(t reflect.Type) *Type {
	return &Type{
		Code: toTypeCode(t),
	}
}

// NewTypeName creates a new TypeName with the given value.
func NewTypeName(value string) *TypeName {
	return &TypeName{
		Value: value,
	}
}

// DriverType converts the internal protobuf Type into a Go reflect.Type
// object.
func (t *Type) DriverType() reflect.Type {
	return fromTypeCode(t.Code)
}

// NewToken creates a new Token object with the given value.
func NewToken(value uint64) *Token {
	return &Token{
		Value: value,
	}
}

// NewEmpty creates a new Empty response object.
func NewEmpty() *Empty {
	return &Empty{}
}

// NewError creates a new Error object.
func NewError(code int64, extendedCode int64, message string) *Error {
	return &Error{
		Code:         code,
		ExtendedCode: extendedCode,
		Message:      message,
	}
}

// Converts Go driver.NamedValue slice of supported types to a protobuf
// NamedValue slice.
func fromDriverNamedValueSlice(driverNamedValues []driver.NamedValue) ([]*NamedValue, error) {
	namedValues := make([]*NamedValue, len(driverNamedValues))
	for i, driverNamedValue := range driverNamedValues {
		namedValue, err := fromDriverNamedValue(driverNamedValue)
		if err != nil {
			return nil, fmt.Errorf("named value %d (%v): %s", i, driverNamedValue.Value, err)
		}
		namedValues[i] = namedValue
	}
	return namedValues, nil
}

// Converts Go driver.Value slice of supported types to a protobuf Value slice.
func fromDriverValueSlice(driverValues []driver.Value) ([]*Value, error) {
	values := make([]*Value, len(driverValues))
	for i, driverValue := range driverValues {
		value, err := fromDriverValue(driverValue)
		if err != nil {
			return nil, fmt.Errorf(" value %d (%v): %s", i, driverValue, err)
		}
		values[i] = value
	}
	return values, nil
}

// Convert a Go driver.NamedValue of a supported type to a protobuf NamedValue.
func fromDriverNamedValue(driverNamedValue driver.NamedValue) (*NamedValue, error) {
	namedValue := &NamedValue{Ordinal: int32(driverNamedValue.Ordinal)}

	if driverNamedValue.Name != "" {
		namedValue.Name = driverNamedValue.Name
	}

	var err error
	namedValue.Value, err = fromDriverValue(driverNamedValue.Value)
	if err != nil {
		return nil, err
	}

	return namedValue, nil
}

// Convert a Go driver.Value of a supported type to a protobuf Value.
func fromDriverValue(driverValue driver.Value) (*Value, error) {
	value := &Value{}

	switch dv := driverValue.(type) {
	case int64:
		value.Type = &Value_Int64{Int64: &Int64{Value: dv}}
	case float64:
		value.Type = &Value_Float64{Float64: &Float64{Value: dv}}
	case bool:
		value.Type = &Value_Bool{Bool: &Bool{Value: dv}}
	case []byte:
		value.Type = &Value_Bytes{Bytes: &Bytes{Value: dv}}
	case string:
		value.Type = &Value_String_{String_: &String{Value: dv}}
	case time.Time:
		value.Type = &Value_Time{Time: &Time{Value: dv.Unix()}}
	default:
		if dv != nil {
			return nil, fmt.Errorf("invalid type %s", reflect.TypeOf(dv).Kind())
		}
		value.Type = &Value_Null{Null: &Null{}}
	}

	return value, nil
}

// Convert a protobuf NamedValue slice to a Go driver.NamedValue slice.
func toDriverNamedValueSlice(namedValues []*NamedValue) ([]driver.NamedValue, error) {
	driverNamedValues := make([]driver.NamedValue, len(namedValues))
	for i, namedValue := range namedValues {
		driverNamedValue, err := toDriverNamedValue(namedValue)
		if err != nil {
			return nil, fmt.Errorf("invalid value %d: %v", i, err)
		}
		driverNamedValues[i] = driverNamedValue
	}
	return driverNamedValues, nil
}

// Convert a protobuf Value slice to a Go driver.Value slice.
func toDriverValueSlice(values []*Value) ([]driver.Value, error) {
	driverValues := make([]driver.Value, len(values))
	for i, value := range values {
		driverValue, err := toDriverValue(value)
		if err != nil {
			return nil, fmt.Errorf("invalid value %d: %v", i, err)
		}
		driverValues[i] = driverValue
	}
	return driverValues, nil
}

// Convert a protobuf NamedValue to a Go driver.NamedValue.
func toDriverNamedValue(namedValue *NamedValue) (driver.NamedValue, error) {
	driverNamedValue := driver.NamedValue{
		Name:    namedValue.Name,
		Ordinal: int(namedValue.Ordinal),
	}

	var err error
	driverNamedValue.Value, err = toDriverValue(namedValue.Value)

	return driverNamedValue, err
}

// Convert a protobuf Value to a Go driver.NamedValue.
func toDriverValue(value *Value) (driver.Value, error) {
	var driverValue driver.Value
	var err error
	switch t := value.Type.(type) {
	case *Value_Null:
		driverValue = nil
	case *Value_Int64:
		driverValue = t.Int64.Value
	case *Value_Float64:
		driverValue = t.Float64.Value
	case *Value_Bool:
		driverValue = t.Bool.Value
	case *Value_Bytes:
		driverValue = t.Bytes.Value
	case *Value_String_:
		driverValue = t.String_.Value
	case *Value_Time:
		driverValue = time.Unix(t.Time.Value, 0)
	default:
		err = fmt.Errorf("invalid type %#v", t)
	}
	return driverValue, err
}

// Convert a Go type object into its associated Type_Code number.
func toTypeCode(t reflect.Type) Type_Code {
	var code Type_Code
	switch t {
	case reflect.TypeOf(int64(0)):
		code = INT64
	case reflect.TypeOf(float64(0)):
		code = FLOAT64
	case reflect.TypeOf(false):
		code = BOOL
	case reflect.TypeOf(byte(0)):
		code = BYTES
	case reflect.TypeOf(""):
		code = STRING
	case reflect.TypeOf(time.Time{}):
		code = TIME
	case reflect.TypeOf(nil):
		code = NULL
	default:
		code = BYTES
	}
	return code
}

// Convert a Type_Code number into its associated Go type object.
func fromTypeCode(code Type_Code) reflect.Type {
	var t reflect.Type
	switch code {
	case INT64:
		t = reflect.TypeOf(int64(0))
	case FLOAT64:
		t = reflect.TypeOf(float64(0))
	case BOOL:
		t = reflect.TypeOf(false)
	case BYTES:
		t = reflect.TypeOf(byte(0))
	case STRING:
		t = reflect.TypeOf("")
	case TIME:
		t = reflect.TypeOf(time.Time{})
	case NULL:
		t = reflect.TypeOf(nil)
	default:
		t = reflect.TypeOf(byte(0))
	}
	return t
}

func init() {
	// Register the type against golang/protobuf/proto too, otherwise
	// grpc.Status.WithDetails won't be able to resolve it.
	proto.RegisterType((*Error)(nil), "sql.Error")
}
