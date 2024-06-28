startdb:
	docker run --name bank-postgres -p 5432:5432 -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=password -d postgres:16-alpine

stopdb:
	docker stop bank-postgres && docker rm bank-postgres

restartdb: stopdb postgres

logs:
	docker logs -f bank-postgres

psql:
	docker exec -it bank-postgres psql -U postgres

dbmigrate:
	cd sql && cd schema && goose postgres "postgres://postgres:password@localhost:5432/bankapi?sslmode=disable" up

dbmigratedown:
	cd sql && cd schema && goose postgres "postgres://postgres:password@localhost:5432/bankapi?sslmode=disable" down

sqlc:
	sqlc generate