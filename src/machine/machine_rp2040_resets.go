//go:build rp2040
// +build rp2040

package machine

import (
	"device/rp"
	"runtime/volatile"
	"unsafe"
)

// RESETS_RESET_Msk is bitmask to reset all peripherals
//
// TODO: This field is not available in the device file.
const RESETS_RESET_Msk = 0x01ffffff

type resetsType struct {
	reset     volatile.Register32
	wdSel     volatile.Register32
	resetDone volatile.Register32
}

var resets = (*resetsType)(unsafe.Pointer(rp.RESETS))

// resetBlock resets hardware blocks specified
// by the bit pattern in bits.
func resetBlock(bits uint32) {
	resets.reset.SetBits(bits)
}

// unresetBlock brings hardware blocks specified by the
// bit pattern in bits out of reset.
func unresetBlock(bits uint32) {
	resets.reset.ClearBits(bits)
}

// unresetBlockWait brings specified hardware blocks
// specified by the bit pattern in bits
// out of reset and wait for completion.
func unresetBlockWait(bits uint32) {
	unresetBlock(bits)
	for !resets.resetDone.HasBits(bits) {
	}
}
