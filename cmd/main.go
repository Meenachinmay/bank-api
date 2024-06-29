package main

import (
	"bank-api/api"
	"bank-api/db/sqlc"
	"database/sql"
	_ "github.com/lib/pq"
	"log"
)

const (
	dbDriver      = "postgres"
	dbSource      = "postgres://postgres:password@localhost:5432/bankapi?sslmode=disable"
	serverAddress = ":4000"
)

func main() {
	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	store := sqlc.NewStore(conn)
	server := api.NewServer(store)

	err = server.Start(serverAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
