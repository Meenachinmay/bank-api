package sqlc

import (
	"bank-api/util"
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

var testQueries *Queries

var testDB *sql.DB

func TestMain(m *testing.M) {
	//var err error
	//
	//testDB, err = sql.Open("postgres", "postgres://postgres:password@localhost:5432/bankapitest?sslmode=disable")
	//if err != nil {
	//	log.Fatal("cannot connect to db:", err)
	//

	util.SetupTestDB()
	testDB = util.TestDB
	testQueries = New(util.TestDB)

	code := m.Run()
	util.CleanupTestDB()
	//// Clean up the test database after tests
	//_, err = testDB.Exec("TRUNCATE TABLE accounts, transfers, entries, referral_codes, referral_history RESTART IDENTITY CASCADE;")
	//if err != nil {
	//	log.Fatal("failed to clean up test db:[TestMain-sqlc]", err)
	//}

	os.Exit(code)
}
