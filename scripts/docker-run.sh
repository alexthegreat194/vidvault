#!/usr/bin/env bash
# Build the image and run vidvault with a host directory mounted at /data.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

IMAGE="${IMAGE:-vidvault:local}"
PORT="${PORT:-8765}"
DATA_DIR="${DATA_DIR:-"$ROOT/data"}"

mkdir -p "$DATA_DIR"

echo "Building $IMAGE …"
docker build -t "$IMAGE" .

echo "Starting container (host $DATA_DIR → /data, http://127.0.0.1:$PORT/) …"
exec docker run --rm -p "${PORT}:8765" -v "${DATA_DIR}:/data" "$IMAGE" -p 8765 /data
