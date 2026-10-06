//go:build !(darwin || (linux && !baremetal && !wasip1 && !wasm_unknown && !wasip2 && !nintendoswitch))

package os

func epipecheck(file *File, e error) {}
