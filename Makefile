.PHONY: install dev down dev-web dev-api migrate seed backup restore-verify media-reconcile-dry test e2e typecheck lint format format-check build check compose-config

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

backup:
	docker compose --env-file .env --profile tools run --rm backup

restore-verify:
	docker compose --env-file .env --profile restore up -d restore-postgres
	docker compose --env-file .env --profile restore run --rm restore

media-reconcile-dry:
	docker compose --env-file .env --profile tools run --rm media-reconcile

test:
	npm --prefix frontend run test
	cd backend && go test ./...

e2e:
	npm --prefix frontend run test:e2e

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
	cd backend && go build ./cmd/api ./cmd/migrate ./cmd/seed ./cmd/worker ./cmd/media-reconcile ./cmd/media-init

check: typecheck lint format-check test build

compose-config:
	docker compose --env-file .env.example config --quiet
