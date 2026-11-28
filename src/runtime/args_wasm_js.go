//go:build wasm && js && !wasip1

package runtime

import "unsafe"

// The page sets go.argv and go.env on the loader, as with Go's wasm_exec.js.
// They are read once at start, before package os takes its copy.

//go:wasmimport gojs runtime.argvString
func jsArgvString(buf unsafe.Pointer, n uint32) uint32

//go:wasmimport gojs runtime.envString
func jsEnvString(buf unsafe.Pointer, n uint32) uint32

func argvString(buf unsafe.Pointer, n uint32) uint32 { return jsArgvString(buf, n) }

func envString(buf unsafe.Pointer, n uint32) uint32 { return jsEnvString(buf, n) }

func init() {
	if a := jsStrings(argvString); len(a) > 0 {
		args = a
	}
	if e := jsStrings(envString); len(e) > 0 {
		env = e
	}
}

func jsStrings(get func(unsafe.Pointer, uint32) uint32) []string {
	n := get(nil, 0)
	if n == 0 {
		return nil
	}
	buf := make([]byte, n)
	get(unsafe.Pointer(&buf[0]), n)
	var out []string
	start := 0
	for i := 0; i <= len(buf); i++ {
		if i == len(buf) || buf[i] == 0 {
			out = append(out, string(buf[start:i]))
			start = i + 1
		}
	}
	return out
}
