//go:build (esp32 || esp32s3) && machine.psram

// PSRAM support shared by the ESP32 Quad driver (machine_esp32_psram_quad.go)
// and the ESP32-S3 Octal driver (machine_esp32s3_psram_octal.go): the .psram
// section, its zero-init and what is left for machine.PSRAM.
package runtime

import (
	"runtime/volatile"
	"unsafe"
)

const hasPSRAM = true

var (
	psramFreeStart uintptr
	psramFreeLen   int
)

// machinePSRAM returns the PSRAM left after the .psram section, for machine.PSRAM.
func machinePSRAM() (uintptr, int) {
	return uintptr(volatile.LoadUint32((*uint32)(unsafe.Pointer(&psramFreeStart)))),
		int(volatile.LoadUint32((*uint32)(unsafe.Pointer(&psramFreeLen))))
}

//go:extern _spsram
var _spsram [0]byte

//go:extern _epsram
var _epsram [0]byte

//go:extern _psram_start
var _psram_start [0]byte

//go:extern _psram_end
var _psram_end [0]byte

// psramWindow returns the linker-provided PSRAM vaddr window
// (targets/esp32s3.ld) and its capacity in 64KB MMU pages.
func psramWindow() (start, end uintptr, pageCap int) {
	start = uintptr(unsafe.Pointer(&_psram_start))
	end = uintptr(unsafe.Pointer(&_psram_end))
	pageCap = int((end - start) / 65536)
	return
}

// zeroInitPSRAM zero-initializes .psram BSS after PSRAM has been mapped
// into the MMU, verifying with a known pattern first that writes actually
// reach the chip (a hardware or timing issue could otherwise silently
// drop them).
func zeroInitPSRAM() string {
	spsram := uintptr(unsafe.Pointer(&_spsram))
	epsram := uintptr(unsafe.Pointer(&_epsram))
	if epsram <= spsram {
		return ""
	}

	*(*uint32)(unsafe.Pointer(spsram)) = 0x5A5A5A5A
	if *(*uint32)(unsafe.Pointer(spsram)) != 0x5A5A5A5A {
		return "PSRAM readback test failed"
	}
	*(*uint32)(unsafe.Pointer(spsram)) = 0
	if *(*uint32)(unsafe.Pointer(spsram)) != 0 {
		return "PSRAM readback test failed"
	}

	for addr := spsram + 4; addr < epsram; addr += 4 {
		*(*uint32)(unsafe.Pointer(addr)) = 0
	}
	return ""
}
