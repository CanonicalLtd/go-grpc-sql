package grpcsql

// Result is the result of a query execution.
type Result struct {
}

// LastInsertId returns the database's auto-generated ID.
func (s *Result) LastInsertId() (int64, error) {
	return 0, nil
}

// RowsAffected returns the number of rows affected by the
// query.
func (s *Result) RowsAffected() (int64, error) {
	return 0, nil
}
