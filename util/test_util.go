package util

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
)

var TestDB *sql.DB

func SetupTestDB() {
	var err error
	TestDB, err = sql.Open("postgres", "postgres://postgres:password@localhost:5432/bankapitest?sslmode=disable")
	if err != nil {
		log.Fatal("cannot connect to test db:", err)
	}

	_, err = TestDB.Exec("TRUNCATE TABLE accounts, transfers, entries, referral_codes, referral_history RESTART IDENTITY CASCADE;")
	if err != nil {
		log.Fatal("failed to clean up test db before tests:", err)
	}
}

func CleanupTestDB() {
	_, err := TestDB.Exec("TRUNCATE TABLE accounts, transfers, entries, referral_codes, referral_history RESTART IDENTITY CASCADE;")
	if err != nil {
		log.Fatal("failed to clean up test db after tests:", err)
	}
}
