package driver

import (
	"context"
	"database/sql/driver"
	"io"
	"reflect"

	"github.com/CanonicalLtd/go-grpc-sql/internal/sql"
)

// Rows is an iterator over an executed query's results.
type Rows struct {
	client sql.GatewayClient // gRPC SQL gateway client to use.
	conn   *sql.Conn         // Info about the connection opened on the gateway.
	rows   *sql.Rows         // Info about the result set on the gateway.
}

func newRows(client sql.GatewayClient, conn *sql.Conn, rows *sql.Rows) *Rows {
	return &Rows{
		client: client,
		conn:   conn,
		rows:   rows,
	}
}

// Columns returns the names of the columns. The number of
// columns of the result is inferred from the length of the
// slice. If a particular column name isn't known, an empty
// string should be returned for that entry.
func (r *Rows) Columns() []string {
	return r.rows.Columns
}

// Close closes the rows iterator.
func (r *Rows) Close() error {
	_, err := r.client.CloseRows(context.Background(), r.rows)
	if err != nil {
		return newErrorFromMethod(err)
	}

	return nil
}

// Next is called to populate the next row of data into
// the provided slice. The provided slice will be the same
// size as the Columns() are wide.
//
// Next should return io.EOF when there are no more rows.
func (r *Rows) Next(dest []driver.Value) error {
	row, err := r.client.Next(context.Background(), r.rows)
	if err != nil {
		return newErrorFromMethod(err)
	}

	columns, err := row.DriverColumns()
	if err != nil {
		return newErrorMisuse(err)
	}

	if columns == nil {
		// This means that there are no more rows.
		return io.EOF
	}

	for i, value := range columns {
		if i == len(dest) {
			break
		}
		dest[i] = value
	}

	return nil
}

// ColumnTypeScanType implements RowsColumnTypeScanType.
func (r *Rows) ColumnTypeScanType(i int) reflect.Type {
	column := sql.NewColumn(r.rows, i)

	typ, err := r.client.ColumnTypeScanType(context.Background(), column)
	if err != nil {
		return nil
	}

	return typ.DriverType()
}

// ColumnTypeDatabaseTypeName implements RowsColumnTypeDatabaseTypeName.
func (r *Rows) ColumnTypeDatabaseTypeName(i int) string {
	column := sql.NewColumn(r.rows, i)

	typeName, err := r.client.ColumnTypeDatabaseTypeName(context.Background(), column)
	if err != nil {
		return ""
	}

	return typeName.Value
}
