#!/bin/sh
set -e

# Maps VIDVAULT_* env vars to vidvault flags (see Dockerfile comments).
truthy() {
	_v=$(printf %s "${1:-}" | tr '[:upper:]' '[:lower:]' | tr -d ' \t\r\n')
	case "$_v" in 1|true|yes|on) return 0 ;; *) return 1 ;; esac
}

extra=""
if truthy "${VIDVAULT_DEBUG:-}"; then extra="$extra -d"; fi
if truthy "${VIDVAULT_DISABLE_BROWSER:-}"; then extra="$extra -disable-browser"; fi

pin_trim=$(printf %s "${VIDVAULT_PIN:-}" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
if [ -n "$pin_trim" ]; then
	exec /vidvault $extra -pin "$VIDVAULT_PIN" "$@"
else
	exec /vidvault $extra "$@"
fi
