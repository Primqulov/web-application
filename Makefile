.PHONY: up down logs seed lint test be-run fe-run bot-run

up:
	docker compose up --build -d

down:
	docker compose down

logs:
	docker compose logs -f --tail=200

seed:
	cd backend && go run ./seed

be-run:
	cd backend && go run ./cmd/api

fe-run:
	cd frontend && npm run dev

bot-run:
	cd bot && go run ./cmd/bot

lint:
	cd backend && go vet ./... && (command -v golangci-lint && golangci-lint run ./... || true)
	cd frontend && npm run lint || true

test:
	cd backend && go test ./...
