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
var testStore *Store

func TestMain(m *testing.M) {

	util.SetupTestDB()
	testDB = util.TestDB
	testQueries = New(util.TestDB)
	testStore = NewStore(testDB)

	code := m.Run()
	util.CleanupTestDB()

	os.Exit(code)
}
