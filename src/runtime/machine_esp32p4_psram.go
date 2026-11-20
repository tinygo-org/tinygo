//go:build esp32p4 && machine.psram

package runtime

import (
	"machine"
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

//go:extern _psram_end
var _psram_end [0]byte

// initPSRAM brings up the P4 external PSRAM and maps the full window into the
// CPU address space. The P4 MSPI/MMU setup lives in the machine package.
func initPSRAM() string {
	if !machine.InitPSRAM80MHz() {
		return "PSRAM init failed"
	}
	if err := zeroInitPSRAM(); err != "" {
		return err
	}

	psramEnd := uintptr(unsafe.Pointer(&_psram_end))
	epsram := uintptr(unsafe.Pointer(&_epsram))
	if psramEnd > epsram {
		psramFreeStart = epsram
		psramFreeLen = int(psramEnd - epsram)
	}
	return ""
}

// zeroInitPSRAM zero-initializes the static .psram section after the PSRAM
// window has been mapped, verifying that writes reach the device.
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
