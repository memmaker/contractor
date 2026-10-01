//go:build js

package main

import (
	"embed"
	"io/fs"
	"syscall/js"
)

//go:embed data_atom
var dataAtom embed.FS

// Hands data_atom to the in-memory globalThis.fs shim in index.html, so all os.* reads work unchanged.
func init() {
	put := js.Global().Get("__fsPut")
	fs.WalkDir(dataAtom, "data_atom", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, _ := dataAtom.ReadFile(p)
		u := js.Global().Get("Uint8Array").New(len(b))
		js.CopyBytesToJS(u, b)
		put.Invoke(p, u)
		return nil
	})
}
