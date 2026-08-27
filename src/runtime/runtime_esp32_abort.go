//go:build esp32 && !qemu

package runtime

import "device"

func abort() {
	for {
		device.Asm("waiti 0")
	}
}
