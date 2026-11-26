//go:build uefi

package runtime

import "device/uefi"

func putchar(c byte) {
	if c == '\n' {
		buf := [2]uefi.CHAR16{uefi.CHAR16('\r'), 0}
		uefi.ST().ConOut.OutputString(&buf[0])
	}
	buf := [2]uefi.CHAR16{uefi.CHAR16(c), 0}
	uefi.ST().ConOut.OutputString(&buf[0])
}

func getchar() byte {
	return 0
}

func buffered() int {
	return 0
}
