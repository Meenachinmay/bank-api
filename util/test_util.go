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

	cleanupDB()
}

func CleanupTestDB() {
	cleanupDB()
	if TestDB != nil {
		err := TestDB.Close()
		if err != nil {
			log.Printf("failed to close test db: %v", err)
		}
	}
}

func cleanupDB() {
	_, err := TestDB.Exec("DISCARD ALL;")
	if err != nil {
		log.Printf("failed to discard all: %v", err)
	}

	_, err = TestDB.Exec("TRUNCATE TABLE accounts, transfers, entries, referral_codes, referral_history RESTART IDENTITY CASCADE;")
	if err != nil {
		log.Fatalf("failed to clean up test db: %v", err)
	}
}
