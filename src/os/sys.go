//go:build !(linux && !baremetal && !tinygo.wasm && !nintendoswitch)

package os

func Hostname() (name string, err error) {
	return "", ErrNotImplemented
}
