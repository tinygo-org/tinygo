//go:build esp && xtensa && !qemu

package runtime

import "device"

func exit(code int) {
	abort()
}

func abort() {
	for {
		device.Asm("waiti 0")
	}
}
