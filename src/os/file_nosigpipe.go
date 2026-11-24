//go:build !(darwin || (linux && !baremetal && !tinygo.wasm && !nintendoswitch))

package os

func epipecheck(file *File, e error) {}
