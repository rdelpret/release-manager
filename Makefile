.PHONY: dev dev-backend dev-frontend test test-backend test-frontend test-e2e build clean migrate deploy setup

dev:
	@echo "Starting backend and frontend..."
	@make dev-backend & make dev-frontend & wait

dev-backend:
	cd backend && $(HOME)/go/bin/air

dev-frontend:
	cd frontend && npm run dev

test: test-backend test-frontend

test-backend:
	cd backend && go test ./... -v -count=1

test-frontend:
	cd frontend && npm test

test-e2e:
	./scripts/test-e2e.sh

build:
	cd backend && go build -o ../bin/server ./cmd/server/main.go
	cd frontend && npm run build

clean:
	rm -rf bin/
	rm -rf frontend/.next

migrate:
	@echo "Run: psql $$DATABASE_URL -f backend/migrations/001_initial.sql"

deploy:
	cd frontend && npm run build
	npm run deploy

setup:
	cp scripts/pre-commit .git/hooks/pre-commit
	chmod +x .git/hooks/pre-commit
	@echo "Pre-commit hook installed"
