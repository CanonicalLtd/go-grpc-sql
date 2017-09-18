package grpcsql_test

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"log"
	"net/http/httptest"
	"time"

	"github.com/CanonicalLtd/go-grpc-sql"
	"github.com/mattn/go-sqlite3"
)

func Example() {
	driver := &sqlite3.SQLiteDriver{}
	server := httptest.NewUnstartedServer(grpcsql.NewServer(driver))
	server.TLS = &tls.Config{NextProtos: []string{"h2"}}
	server.StartTLS()
	defer server.Close()

	options := &grpcsql.DialOptions{
		Address:    server.Listener.Addr().String(),
		Name:       ":memory:",
		Timeout:    5 * time.Second,
		CertFile:   "testdata/clientcert.pem",
		KeyFile:    "testdata/clientkey.pem",
		SkipVerify: true,
	}
	db, err := sql.Open("grpc", options.String())
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
