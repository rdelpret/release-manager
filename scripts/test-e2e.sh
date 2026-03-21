#!/bin/bash
set -e

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

# Test database config
export DATABASE_URL="postgresql://test:test@localhost:5555/mrp_test?sslmode=disable"
export SESSION_SECRET="test-secret-for-e2e-tests-only"
export ALLOWED_EMAILS="dev@subwave.music"
export FRONTEND_URL="http://localhost:3000"
export ENV="development"
export PORT="8080"

cleanup() {
  echo "Cleaning up..."
  kill $BACKEND_PID 2>/dev/null || true
  docker compose -f docker-compose.test.yml down -v 2>/dev/null || true
}
trap cleanup EXIT

# Start test database
echo "Starting test database..."
docker compose -f docker-compose.test.yml up -d --wait

# Run migrations
echo "Running migrations..."
PSQL="/opt/homebrew/Cellar/libpq/18.3/bin/psql"
if [ ! -f "$PSQL" ]; then
  PSQL="psql"
fi
$PSQL "$DATABASE_URL" -f backend/migrations/001_initial.sql

# Build and start backend
echo "Starting backend..."
cd backend
go build -o ../bin/test-server ./cmd/server/main.go
cd ..
./bin/test-server &
BACKEND_PID=$!
sleep 2

# Verify backend is up
curl -sf http://localhost:8080/api/health > /dev/null || { echo "Backend failed to start"; exit 1; }
echo "Backend running (PID $BACKEND_PID)"

# Run Playwright tests
echo "Running Playwright tests..."
cd frontend
npx playwright test "$@"
TEST_EXIT=$?

exit $TEST_EXIT
