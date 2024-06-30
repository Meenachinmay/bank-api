package sqlc

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

var testQueries *Queries
var testDB *sql.DB

func TestMain(m *testing.M) {
	var err error

	testDB, err = sql.Open("postgres", "postgres://postgres:password@localhost:5432/bankapitest?sslmode=disable")
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	testQueries = New(testDB)

	code := m.Run()

	// Clean up the test database after tests
	_, err = testDB.Exec("TRUNCATE TABLE accounts, transfers, entries, referral_codes, referral_history RESTART IDENTITY CASCADE;")
	if err != nil {
		log.Fatal("failed to clean up test db:", err)
	}

	os.Exit(code)
}
