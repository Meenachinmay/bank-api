#DB_SOURCE_TEST=postgres://postgres:password@localhost:5432/bankapitest?sslmode=disable
#DB_SOURCE=postgres://postgres:password@localhost:5432/bankapi?sslmode=disable

startdb:
	docker run --name bank-postgres -p 5432:5432 -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=password -d postgres:16-alpine

stopdb:
	docker stop bank-postgres && docker rm bank-postgres

restartdb: stopdb postgres

logs:
	docker logs -f bank-postgres

psql:
	docker exec -it bank-postgres psql -U postgres

createdb:
	@echo "Checking if test database exists..."
	psql $(DB_SOURCE) -tc "SELECT 1 FROM pg_database WHERE datname = 'bankapitest'" | grep -q 1 || psql $(DB_SOURCE) -c 'CREATE DATABASE bankapitest;'

dbmigrate:
	cd sql && cd schema && goose postgres "${DB_SOURCE}" up

dbmigratedown:
	cd sql && cd schema && goose postgres "${DB_SOURCE}" down

dbmigrate-test:
	cd sql && cd schema && goose postgres "${DB_SOURCE_TEST}" up

dbmigratedown-test:
	cd sql && cd schema && goose postgres "${DB_SOURCE_TEST}" down

sqlc:
	sqlc generate

test:
	DB_SOURCE=${DB_SOURCE_TEST} go test -v -cover ./...

server:
	go run cmd/main.go