docker-build:
	docker build -t app .

docker-rerun:
	docker rm -f app || true
	docker run --name=app -p 8080:8080 -e APP_PG__HOST=host.docker.internal app

run-infra:
	docker compose up postgres

run-all:
	docker compose up --build

install:
	go mod download

run:
	go run .
