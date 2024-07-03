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
	psql $(DB_SOURCE_TEST) -tc "SELECT 1 FROM pg_database WHERE datname = 'bankapitest'" | grep -q 1 || psql $(DB_SOURCE) -c 'CREATE DATABASE bankapitest;'

dbmigrate:
	cd sql && cd schema && goose postgres "${DB_SOURCE_PROD}" up

dbmigratedown:
	cd sql && cd schema && goose postgres "${DB_SOURCE_PROD}" down

dbreset:
	dbmigratedown && dbmigratedown

dbmigrate-test:
	cd sql && cd schema && goose postgres "${DB_SOURCE_TEST}" up

dbmigratedown-test:
	cd sql && cd schema && goose postgres "${DB_SOURCE_TEST}" down

dbresettest:
	dbmigrate-test && dbmigratedown-test

sqlc:
	sqlc generate

test:
	DB_SOURCE=${DB_SOURCE_TEST} go test -v -count=1 -cover ./...

store-test:
	DB_SOURCE=${DB_SOURCE_TEST}	go test -tags=storetest -v -count=1 ./db/sqlc

handler-test:
	DB_SOURCE=${DB_SOURCE_TEST}	go test -tags=handlertest -v -count=1 ./api

server:
	go run cmd/main.go


## up: starts all containers in the background without forcing build
up:
	@echo "Starting Docker images..."
	docker-compose up -d
	@echo "Docker images started!"

## up_build: stops docker-compose (if running), builds all projects and starts docker compose
up_build:
	@echo "Stopping docker images (if running...)"
	docker-compose down
	@echo "Building (when required) and starting docker images..."
	docker-compose up --build -d
	@echo "Docker images built and started!"

## down: stop docker compose
down:
	@echo "Stopping docker compose..."
	docker-compose down
	@echo "Done!"


se:
	export $(grep -v '^#' .env | xargs)