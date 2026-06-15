.PHONY: dev up down logs test build

dev:
	docker compose up --build

up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f

test:
	cd services/api && go test ./...
	cd apps/web && npm run typecheck

build:
	docker compose build
