#!/usr/bin/env sh
# Frontend dev server with API/video proxied to a remote vidvault instance.
# Usage: ./scripts/dev-remote.sh [backend-url]
# Default backend: http://192.168.1.184:8766
set -e
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
export VIDVAULT_BACKEND="${1:-http://192.168.1.184:8766}"
echo "VIDVAULT_BACKEND=$VIDVAULT_BACKEND"
exec npm run dev --prefix "$ROOT/frontend"
