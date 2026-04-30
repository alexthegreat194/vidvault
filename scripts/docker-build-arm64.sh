#!/usr/bin/env bash
# Build a arm64-ready Docker image (linux/arm64 by default).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

IMAGE="${IMAGE:-alexthegreat1941/vidvault:arm64}"
PLATFORM="${PLATFORM:-linux/arm64}"
PUSH="${PUSH:-1}"
LOAD="${LOAD:-1}"
PROVENANCE="${PROVENANCE:-false}"

echo "Building image for arm64..."
echo "  image:    $IMAGE"
echo "  platform: $PLATFORM"

cmd=(docker buildx build --platform "$PLATFORM" -t "$IMAGE" --provenance="$PROVENANCE")

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
