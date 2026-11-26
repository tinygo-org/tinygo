//go:build machine.psram

package runtime

import (
	"runtime/volatile"
	"unsafe"
)

var (
	psramReady      uint8
	psramInitFailed uint8
	psramInitErr    string
)

//go:extern _spsram
var _spsram [0]byte

//go:extern _epsram
var _epsram [0]byte

//go:extern _psram_start
var _psram_start [0]byte

//go:extern _psram_end
var _psram_end [0]byte

// startPSRAM runs the PSRAM driver, which maps the __psram_size bytes the
// target declares. Without .psram variables a failure only leaves
// machine.PSRAM nil, otherwise checkPSRAM stops the program.
func startPSRAM() {
	err := initPSRAM()
	if err == "" {
		err = zeroInitPSRAM()
	}
	if err == "" {
		volatile.StoreUint8(&psramReady, 1)
		return
	}
	if uintptr(unsafe.Pointer(&_epsram)) == uintptr(unsafe.Pointer(&_spsram)) {
		return
	}
	psramInitErr = err
	volatile.StoreUint8(&psramInitFailed, 1)
}

// checkPSRAM reports a startPSRAM failure once the serial port is up.
// The volatile load stops interp from folding it before startPSRAM runs.
func checkPSRAM() {
	if volatile.LoadUint8(&psramInitFailed) != 0 {
		psramFailed(psramInitErr)
	}
}

// machinePSRAM returns the PSRAM left after the .psram section, for machine.PSRAM.
func machinePSRAM() (uintptr, int) {
	if volatile.LoadUint8(&psramReady) == 0 {
		return 0, 0
	}
	start := uintptr(unsafe.Pointer(&_epsram))
	return start, int(uintptr(unsafe.Pointer(&_psram_end)) - start)
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
