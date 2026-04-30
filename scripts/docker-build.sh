#!/usr/bin/env bash
# Build a Docker image from the repo Dockerfile (host platform unless PLATFORM is set).
# Same env vars as docker-build-amd64.sh / docker-build-arm64.sh.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

IMAGE="${IMAGE:-alexthegreat1941/vidvault:latest}"
PLATFORM="${PLATFORM:-}"
PUSH="${PUSH:-1}"
LOAD="${LOAD:-1}"
PROVENANCE="${PROVENANCE:-false}"

echo "Building image..."
echo "  image:    $IMAGE"
if [[ -n "$PLATFORM" ]]; then
  echo "  platform: $PLATFORM"
else
  echo "  platform: (default / host)"
fi

cmd=(docker buildx build -t "$IMAGE" --provenance="$PROVENANCE")
if [[ -n "$PLATFORM" ]]; then
  cmd+=(--platform "$PLATFORM")
fi

if [[ "$PUSH" == "1" ]]; then
  cmd+=(--push)
elif [[ "$LOAD" == "1" ]]; then
  cmd+=(--load)
fi

cmd+=(.)
"${cmd[@]}"

echo
echo "Build complete: $IMAGE"
echo "Run example:"
echo "  docker run --rm -p 8765:8765 -v /path/to/videos:/data $IMAGE -p 8765 /data"
