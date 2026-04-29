#!/usr/bin/env bash
# Build an amd64 Docker image (linux/amd64 by default). Use on x86_64 hosts
# (e.g. ZimaBlade / ZimaOS, Intel NUC, typical Linux servers).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

IMAGE="${IMAGE:-alexthegreat1941/vidvault:amd64}"
PLATFORM="${PLATFORM:-linux/amd64}"
PUSH="${PUSH:-0}"
LOAD="${LOAD:-1}"
PROVENANCE="${PROVENANCE:-false}"

echo "Building image for amd64..."
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
