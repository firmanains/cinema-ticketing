.PHONY: run down migrate-up migrate-down test

run:
	docker compose up -d postgres redis
	docker compose run --rm migrate up
	docker compose up --build app

migrate-up:
	docker compose run --rm migrate up

migrate-down:
	docker compose run --rm migrate down

down:
	docker compose down

test:
	go test ./...
