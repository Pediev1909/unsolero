.PHONY: install dev down dev-web dev-api migrate seed test typecheck lint format format-check build check compose-config

install:
	npm --prefix frontend ci
	cd backend && go mod download

dev:
	docker compose up --build

down:
	docker compose down

dev-web:
	npm --prefix frontend run dev

dev-api:
	cd backend && go run ./cmd/api

migrate:
	cd backend && go run ./cmd/migrate

seed:
	cd backend && go run ./cmd/seed

test:
	npm --prefix frontend run test
	cd backend && go test ./...

typecheck:
	npm --prefix frontend run typecheck

lint:
	npm --prefix frontend run lint
	cd backend && go vet ./...

format:
	npm --prefix frontend run format
	cd backend && gofmt -w .

format-check:
	npm --prefix frontend run format:check
	@test -z "$$(gofmt -l backend)"

build:
	npm --prefix frontend run build
	cd backend && go build ./cmd/api ./cmd/migrate ./cmd/seed

check: typecheck lint format-check test build

compose-config:
	docker compose --env-file .env.example config --quiet
