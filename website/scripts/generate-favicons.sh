#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
# Optional asset-maintenance command; normal builds use the committed images.
magick -background none static/img/favicon.svg -resize 180x180 static/img/apple-touch-icon.png
magick -background none static/img/favicon.svg -resize 32x32 static/img/favicon-32.png
magick -background none static/img/favicon.svg -define icon:auto-resize=64,48,32,16 static/img/favicon.ico
