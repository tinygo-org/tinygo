// +build darwin linux,!baremetal,!wasi freebsd,!baremetal
// +build !nintendoswitch

// +build !gc.none,!gc.extalloc

package runtime

const heapSize = 1 * 1024 * 1024 // 1MB to start

var heapStart, heapEnd uintptr

func preinit() {
	heapStart = uintptr(malloc(heapSize))
	heapEnd = heapStart + heapSize
}
