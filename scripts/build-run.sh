#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
npm run build:frontend
go build -o vidvault ./src
exec ./vidvault \
    -p 5001 \
    -disable-browser \
    -d \
    ./data
