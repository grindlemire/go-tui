#!/usr/bin/env bash
set -euo pipefail

DIR="$(cd "$(dirname "$0")/../docs/public" && pwd)"

if ! command -v rsvg-convert &>/dev/null; then
  echo "rsvg-convert not found. Install with: brew install librsvg"
  exit 1
fi

rsvg-convert -w 1200 -h 630 "$DIR/og-image-dark.svg"  -o "$DIR/og-image-dark.png"
rsvg-convert -w 1200 -h 630 "$DIR/og-image-light.svg" -o "$DIR/og-image-light.png"

echo "Generated og-image-dark.png and og-image-light.png"
