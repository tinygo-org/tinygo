//go:build darwin || windows || wasip1 || (linux && !baremetal && !wasm_unknown && !wasip2 && !nintendoswitch)

package runtime

import "unsafe"

//export getenv
func libc_getenv(name *byte) *byte

func startupEnv(key string) string {
	name := make([]byte, len(key)+1)
	copy(name, key)
	value := libc_getenv(&name[0])
	if value == nil {
		return ""
	}
	length := strlen(unsafe.Pointer(value))
	return string(unsafe.Slice(value, length))
}

func init() {
	if value := startupEnv("GOROOT"); value != "" {
		goroot = value
	}
}
