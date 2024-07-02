package main

import (
	"bank-api/api"
	"bank-api/db/sqlc"
	"database/sql"
	_ "github.com/lib/pq"
	"log"
)

const (
	serverAddress = ":8080"
)

func main() {
	conn, err := sql.Open("postgres", "postgres://postgres:password@postgres:5432/bankapi?sslmode=disable")
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
