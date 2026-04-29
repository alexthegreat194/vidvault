#!/usr/bin/env bash
# Build a generic Docker image from the repo Dockerfile (host platform unless PLATFORM is set).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

IMAGE="${IMAGE:-alexthegreat1941/vidvault:latest}"
PLATFORM="${PLATFORM:-}"

echo "Building $IMAGE …"
if [[ -n "$PLATFORM" ]]; then
	docker build --platform "$PLATFORM" -t "$IMAGE" .
else
	docker build -t "$IMAGE" .
fi

echo "Done: $IMAGE"
