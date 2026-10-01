#!/usr/bin/env bash
# Build the browser version and publish it to https://ruzzoli.de/games/contractor/play/
set -euo pipefail
cd "$(dirname "$0")"
./build_web.sh
rsync -az --delete /Users/felix/Projects/fx-games/site/contractor/play/ ruzzoli.de:/var/www/ruzzoli.de/games/contractor/play/
echo "published: https://ruzzoli.de/games/contractor/play/"
