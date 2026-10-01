#!/usr/bin/env bash
# Browser (js/wasm) build. Ebiten frontend only; data_atom is embedded (data_js.go).
set -euo pipefail
cd "$(dirname "$0")"
OUT=${OUT:-/Users/felix/Projects/fx-games/site/contractor/play}
WORK=builds/web
mkdir -p "$OUT" "$WORK"

# tcell v2.8.1's wscreen.go (js only) lacks Get/SetClipboard: build against a patched copy via -modfile.
TCELL=$(go list -m -f '{{.Dir}}' github.com/gdamore/tcell/v2)
rm -rf "$WORK/tcell" && cp -R "$TCELL" "$WORK/tcell" && chmod -R u+w "$WORK/tcell"
printf 'package tcell\n\nfunc (t *wScreen) SetClipboard(_ []byte) {}\nfunc (t *wScreen) GetClipboard()         {}\n' > "$WORK/tcell/wscreen_clip.go"
cp go.mod "$WORK/go.mod"; cp go.sum "$WORK/go.sum"
go mod edit -modfile="$WORK/go.mod" -replace "github.com/gdamore/tcell/v2=./$WORK/tcell"

GOOS=js GOARCH=wasm go build -modfile="$WORK/go.mod" -tags ebitensinglethread,ebiten \
  -trimpath -ldflags '-s -w' -o "$OUT/contractor.wasm" .
gzip -9 -k -f "$OUT/contractor.wasm"
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" "$OUT/"
ls -lh "$OUT"/contractor.wasm*
