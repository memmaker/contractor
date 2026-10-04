#!/usr/bin/env bash
# Browser (js/wasm) build. Ebiten frontend only; data_atom is embedded (data_js.go).
set -euo pipefail
cd "$(dirname "$0")"
OUT=${OUT:-/Users/felix/Projects/fx-games/site/contractor/play}
WORK=builds/web
mkdir -p "$OUT" "$WORK"

# kelindar/binary v1.2.3's js build lacks ToString: build against a patched copy via -modfile.
BINARY=$(go list -m -f '{{.Dir}}' github.com/kelindar/binary)
rm -rf "$WORK/binary" && cp -R "$BINARY" "$WORK/binary" && chmod -R u+w "$WORK/binary"
printf '//go:build js\n\npackage binary\n\nfunc ToString(b *[]byte) string { return string(*b) }\n' > "$WORK/binary/convert_js_tostring.go"
cp go.mod "$WORK/go.mod"; cp go.sum "$WORK/go.sum"
go mod edit -modfile="$WORK/go.mod" -replace "github.com/kelindar/binary=./$WORK/binary"

GOOS=js GOARCH=wasm go build -modfile="$WORK/go.mod" -tags ebitensinglethread,ebiten \
  -trimpath -ldflags '-s -w' -o "$OUT/contractor.wasm" .
gzip -9 -k -f "$OUT/contractor.wasm"
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" "$OUT/"
ls -lh "$OUT"/contractor.wasm*
