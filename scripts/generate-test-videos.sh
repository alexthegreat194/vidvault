#!/usr/bin/env bash
# Generate small browser-playable MP4s for local gallery testing.
# Requires ffmpeg on the host (e.g. brew install ffmpeg).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

COUNT=50
OUT_DIR="$ROOT/data"
FOLDERS=0
DUPLICATES=0
DURATION="10"
WIDTH=320
HEIGHT=180
PREFIX="test"
CLEAN=0

usage() {
	cat <<'EOF'
Usage: generate-test-videos.sh [options]

Generate small MP4 test clips for filling the vidvault gallery.

Options:
  -n, --count N       Number of videos to create (default: 50)
  -o, --out DIR       Output directory (default: <repo>/data)
      --folders N     Spread files across folder-01 … folder-N (default: 0 = root only)
      --duplicates N  Create N extra copies of the first generated file (duplicate testing)
      --duration SEC  Clip length in seconds (default: 10)
      --width PX      Video width (default: 320)
      --height PX     Video height (default: 180)
      --prefix STR    Filename prefix (default: test)
      --clean         Remove existing PREFIX-*.mp4 under output before generating
  -h, --help          Show this help

Examples:
  ./scripts/generate-test-videos.sh -n 100
  ./scripts/generate-test-videos.sh -n 150 --folders 4
  npm run seed:videos -- -n 200 --folders 5
EOF
}

while [[ $# -gt 0 ]]; do
	case "$1" in
	-n | --count)
		COUNT="$2"
		shift 2
		;;
	-o | --out)
		OUT_DIR="$2"
		shift 2
		;;
	--folders)
		FOLDERS="$2"
		shift 2
		;;
	--duplicates)
		DUPLICATES="$2"
		shift 2
		;;
	--duration)
		DURATION="$2"
		shift 2
		;;
	--width)
		WIDTH="$2"
		shift 2
		;;
	--height)
		HEIGHT="$2"
		shift 2
		;;
	--prefix)
		PREFIX="$2"
		shift 2
		;;
	--clean)
		CLEAN=1
		shift
		;;
	-h | --help)
		usage
		exit 0
		;;
	*)
		echo "Unknown option: $1" >&2
		usage >&2
		exit 1
		;;
	esac
done

if ! command -v ffmpeg >/dev/null 2>&1; then
	echo "ffmpeg is required but not found on PATH." >&2
	echo "Install it, e.g.: brew install ffmpeg" >&2
	exit 1
fi

if ! [[ "$COUNT" =~ ^[0-9]+$ ]] || [[ "$COUNT" -lt 1 ]]; then
	echo "--count must be a positive integer" >&2
	exit 1
fi
if ! [[ "$FOLDERS" =~ ^[0-9]+$ ]]; then
	echo "--folders must be a non-negative integer" >&2
	exit 1
fi
if ! [[ "$DUPLICATES" =~ ^[0-9]+$ ]]; then
	echo "--duplicates must be a non-negative integer" >&2
	exit 1
fi

mkdir -p "$OUT_DIR"

if [[ "$CLEAN" -eq 1 ]]; then
	echo "Removing existing ${PREFIX}-*.mp4 under $OUT_DIR …"
	find "$OUT_DIR" -type f -name "${PREFIX}-*.mp4" -delete
fi

generate_one() {
	local i="$1"
	local dest_dir="$2"
	local color freq outfile

	color=$(printf '%06x' $(( (i * 7919) % 0xFFFFFF )))
	freq=$((440 + (i % 200)))
	outfile="${dest_dir}/${PREFIX}-$(printf '%04d' "$i").mp4"

	mkdir -p "$dest_dir"
	ffmpeg -hide_banner -loglevel error \
		-f lavfi -i "color=c=0x${color}:s=${WIDTH}x${HEIGHT}:d=${DURATION}" \
		-f lavfi -i "sine=frequency=${freq}:duration=${DURATION}" \
		-shortest \
		-c:v libx264 -pix_fmt yuv420p -crf 30 -preset ultrafast \
		-c:a aac -b:a 48k -movflags +faststart \
		-y "$outfile"
	printf '%s\n' "$outfile"
}

echo "Generating $COUNT test video(s) in $OUT_DIR …"

first_file=""
for ((i = 1; i <= COUNT; i++)); do
	dest_dir="$OUT_DIR"
	if [[ "$FOLDERS" -gt 0 ]]; then
		folder_idx=$(( (i - 1) % FOLDERS + 1 ))
		dest_dir="${OUT_DIR}/folder-$(printf '%02d' "$folder_idx")"
	fi
	outfile=$(generate_one "$i" "$dest_dir")
	if [[ -z "$first_file" ]]; then
		first_file="$outfile"
	fi
	printf '\r  %d / %d' "$i" "$COUNT"
done
echo

if [[ "$DUPLICATES" -gt 0 ]]; then
	if [[ -z "$first_file" || ! -f "$first_file" ]]; then
		echo "No source file available for duplicates" >&2
		exit 1
	fi
	echo "Creating $DUPLICATES duplicate(s) of $(basename "$first_file") …"
	for ((d = 1; d <= DUPLICATES; d++)); do
		dup_path="${OUT_DIR}/${PREFIX}-dup-$(printf '%04d' "$d").mp4"
		cp "$first_file" "$dup_path"
	done
fi

total_files=$((COUNT + DUPLICATES))
total_bytes=0
while IFS= read -r -d '' f; do
	size=$(stat -f%z "$f" 2>/dev/null || stat -c%s "$f")
	total_bytes=$((total_bytes + size))
done < <(find "$OUT_DIR" -type f -name "${PREFIX}*.mp4" -print0)

if command -v numfmt >/dev/null 2>&1; then
	size_human=$(numfmt --to=iec-i --suffix=B "$total_bytes")
else
	size_human="${total_bytes} bytes"
fi

echo "Done: $total_files file(s) under $OUT_DIR ($size_human total)"
