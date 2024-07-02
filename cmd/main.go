package main

import (
	"bank-api/api"
	"bank-api/db/sqlc"
	"database/sql"
	"errors"
	_ "github.com/lib/pq"
	"log"
	"os"
	"time"
)

const (
	serverAddress = ":8080"
)

var counts int64

func main() {
	// connect to database
	conn := connectToDB()
	defer conn.Close()

	store := sqlc.NewStore(conn)
	server := api.NewServer(store)

	err := server.Start(serverAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}

func openDB() (*sql.DB, error) {
	dbURL := os.Getenv("DB_SOURCE_PROD")
	if dbURL == "" {
		dbURL = os.Getenv("DB_SOURCE")
	}
	if dbURL == "" {
		return nil, errors.New("missing DATABASE_URL")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func connectToDB() *sql.DB {
	for {
		connection, err := openDB()
		if err != nil {
			log.Println("Could not connect to database, Postgres is not ready...")
			counts += 1
		} else {
			log.Println("Connected to database...")
			return connection
		}

		if counts > 10 {
			log.Println(err)
			return nil
		}

		log.Println("Waiting for database to become ready...")
		time.Sleep(2 * time.Second)
		continue
	}
}
