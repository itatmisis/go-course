export PG_DSN=postgres://tracker:tracker@localhost:5432/tracker?sslmode=disable
export MIGRATIONS_PATH=./migrations

local-db-up:
	docker run --name task-tracker-db \
      -e POSTGRES_USER=tracker \
      -e POSTGRES_PASSWORD=tracker \
      -e POSTGRES_DB=tracker \
      -p 127.0.0.1:5432:5432 \
      -d postgres:17;

local-db-down:
	docker stop task-tracker-db;
	docker rm task-tracker-db

add-migration:
	goose create fix_me sql

migrate:
	 goose -dir ./migrations postgres ${PG_DSN} up