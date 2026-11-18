//go:build (esp32 && !qemu) || esp32c3 || esp32c6 || esp32h2 || esp32p4

package runtime

func exit(code int) {
	abort()
}
