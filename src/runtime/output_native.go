// +build output.native

package runtime

// The "native" output is useful on devices with a native stdout, usually real
// operating systems (Linux, etc).

func initOutput() {
}

func putchar(c byte) {
	nativePutchar(c)
}
