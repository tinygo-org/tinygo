//go:build (esp8266 || esp32 || esp32s3) && !qemu

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
