#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
go build -o vidvault ./src
exec ./vidvault \
    -p 5001 \
    -disable-browser \
    -d \
    -pin 1234 \
    ./data
