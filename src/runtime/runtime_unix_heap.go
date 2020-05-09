// +build darwin linux,!baremetal freebsd,!baremetal
// +build !nintendoswitch

// +build gc.conservative gc.leaking

package runtime

const heapSize = 1 * 1024 * 1024 // 1MB to start

var heapStart, heapEnd uintptr

func preinit() {
	heapStart = uintptr(malloc(heapSize))
	heapEnd = heapStart + heapSize
}
