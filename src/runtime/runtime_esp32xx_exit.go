//go:build (esp32 && !qemu) || esp32c3 || esp32c6

package runtime

func exit(code int) {
	abort()
}
