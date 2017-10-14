package grpcsql

import (
	"database/sql/driver"
	"io"
	"reflect"

	"github.com/CanonicalLtd/go-grpc-sql/internal/protocol"
)

// Rows is an iterator over an executed query's results.
type Rows struct {
	conn    *Conn
	id      int64
	columns []string
}

// Columns returns the names of the columns. The number of
// columns of the result is inferred from the length of the
// slice. If a particular column name isn't known, an empty
// string should be returned for that entry.
func (r *Rows) Columns() []string {
	return r.columns
}

// Close closes the rows iterator.
func (r *Rows) Close() error {
	_, err := r.conn.exec(protocol.NewRequestRowsClose(r.id))
	return err
}

// Next is called to populate the next row of data into
// the provided slice. The provided slice will be the same
// size as the Columns() are wide.
//
// Next should return io.EOF when there are no more rows.
func (r *Rows) Next(dest []driver.Value) error {
	response, err := r.conn.exec(protocol.NewRequestNext(r.id, int64(len(dest))))
	if err != nil {
		return err
	}

	if response.Next().Eof {
		return io.EOF
	}

	values, err := protocol.ToDriverValues(response.Next().Values)
	if err != nil {
		return err
	}
	for i, value := range values {
		dest[i] = value
	}
	return nil
}

// ColumnTypeScanType implements RowsColumnTypeScanType.
func (r *Rows) ColumnTypeScanType(i int) reflect.Type {
	response, err := r.conn.exec(protocol.NewRequestColumnTypeScanType(r.id, int64(i)))
	if err != nil {
		return reflect.Type(nil)
	}
	return protocol.FromValueCode(response.ColumnTypeScanType().Code)
}

// ColumnTypeDatabaseTypeName implements RowsColumnTypeDatabaseTypeName.
func (r *Rows) ColumnTypeDatabaseTypeName(i int) string {
	response, err := r.conn.exec(protocol.NewRequestColumnTypeDatabaseTypeName(r.id, int64(i)))
	if err != nil {
		return ""
	}
	return response.ColumnTypeDatabaseTypeName().Name
}
