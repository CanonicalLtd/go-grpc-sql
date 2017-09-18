package grpcsql_test

import (
	"database/sql"
	"fmt"
	"log"
)

func Example() {
	db, err := sql.Open("grpc", "test?certfile=testdata/clientcert.pem&keyfile=testdata/clientkey.pem")
	if err != nil {
		log.Fatalf("failed to create grpc database: %s", err)
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("failed to create grpc transaction: %s", err)
	}
	defer tx.Rollback()

	result, err := tx.Exec("CREATE TABLE test (n INT)")
	if err != nil {
		log.Fatalf("failed to execute create table statement over grpc: %s", err)
	}

	rows, err := tx.Query("SELECT n FROM test")
	if err != nil {
		log.Fatalf("failed to select rows over grpc: %s", err)
	}
	defer rows.Close()

	// Output:
	// 0 <nil>
	// 0 <nil>
	// false
	fmt.Println(result.LastInsertId())
	fmt.Println(result.RowsAffected())
	fmt.Println(rows.Next())
}
