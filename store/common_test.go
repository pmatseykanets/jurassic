//go:build integration
// +build integration

package store

import (
	"database/sql"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

var (
	testDBConnString = "postgres://jurassic_test:secret@localhost:5432/jurassic_test?sslmode=disable"
	testDB           *sql.DB
)

func setUpTestDB(t *testing.T) {
	if testDB != nil {
		return
	}

	if s := os.Getenv("JURASSIC_TEST_DB_CONN"); s != "" {
		testDBConnString = s
	}

	migrations, err := migrate.New("file://../db/migrations", testDBConnString)
	if err != nil {
		t.Fatal(err)
	}

	err = migrations.Up()
	if err != nil && err != migrate.ErrNoChange {
		t.Fatal(err)
	}

	testDB, err = sql.Open("postgres", testDBConnString)
	if err != nil {
		t.Fatal(err)
	}
}
