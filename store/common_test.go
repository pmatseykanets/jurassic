//go:build integration
// +build integration

package store

import (
	"database/sql"

	_ "github.com/lib/pq"
)

var (
	testDBConnString = "postgres://jurassic_test:secret@localhost:5432/jurassic_test?sslmode=disable"
	testDB           *sql.DB
)

func init() {
	var err error
	testDB, err = sql.Open("postgres", testDBConnString)
	if err != nil {
		panic(err)
	}
}
