#!/bin/bash
set -e
cd "$(dirname "$0")"

DB_PATH="server_data/runtime/db.sqlite"
SEED_SQL="testing/seed_core.sql"

echo "==> Cleaning up old server data..."
rm -rf server_data

echo "==> Generating problem files..."
bash testing/generate_server_data.sh

echo "==> Pulling gcc:latest (ensures Docker image is cached)..."
docker pull gcc:latest

echo "==> Starting server to initialise DB (migrations)..."
go run main.go &
SERVER_PID=$!

echo -n "    Waiting for DB..."
until [ -f "$DB_PATH" ]; do sleep 0.5; echo -n "."; done
sleep 1  # let migrations finish
echo " ready."

echo "==> Seeding contest data..."
sqlite3 "$DB_PATH" < "$SEED_SQL"

echo "==> Registering test teams..."
REG1=$(curl -s -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Team Alpha"}')
REG2=$(curl -s -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Team Beta"}')

echo "    Team Alpha → $(echo $REG1 | grep -o '"login_code":"[^"]*"')"
echo "    Team Beta  → $(echo $REG2 | grep -o '"login_code":"[^"]*"')"

echo "==> Stopping init server..."
kill $SERVER_PID
wait $SERVER_PID 2>/dev/null

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Setup done. Starting server — logs below."
echo "  Frontend:  http://localhost:5173"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
go run main.go
