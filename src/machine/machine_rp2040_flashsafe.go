//go:build tinygo && rp2040

// "Flash safe" follows the RP2040/Pico SDK terminology: flash operations
// must run while the other core is not executing from XIP flash.
//
// Use linkname to call runtime hooks from package machine without creating
// an import cycle.

package machine

import (
	"runtime/interrupt"
	_ "unsafe"
)

//go:linkname rp2040EnterFlashSafeSection runtime.rp2040EnterFlashSafeSection
func rp2040EnterFlashSafeSection() interrupt.State

//go:linkname rp2040ExitFlashSafeSection runtime.rp2040ExitFlashSafeSection
func rp2040ExitFlashSafeSection(state interrupt.State)
