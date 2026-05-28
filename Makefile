.PHONY: run down migrate-up migrate-down test test-coverage

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

test-coverage:
	go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out
