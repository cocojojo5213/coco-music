#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"

(
  cd "$ROOT/backend"
  ADDR="${ADDR:-:18280}" DATA_DIR="$ROOT/backend/data" go run ./cmd/server
) &
API_PID=$!

cleanup() {
  kill "$API_PID" 2>/dev/null || true
}
trap cleanup EXIT

cd "$ROOT/frontend"
npm run dev -- --host 0.0.0.0 --port 5173
