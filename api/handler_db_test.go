package api

import (
	"bank-api/db/sqlc"
	"bank-api/util"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

var testStore *sqlc.Store

func TestMain(m *testing.M) {
	util.SetupTestDB()
	testStore = sqlc.NewStore(util.TestDB)

	code := m.Run()
	//util.CleanupTestDB()

	os.Exit(code)
}
