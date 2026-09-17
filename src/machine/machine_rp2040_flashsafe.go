//go:build tinygo && rp2040

// "Flash safe" is RP2040/Pico SDK terminology: flash operations must run
// while the other core is not executing from XIP flash.

package machine

import (
	"runtime/interrupt"
	_ "unsafe"
)

// The flash-safe hooks are implemented in package runtime and accessed via
// linkname to avoid an import cycle.

//go:linkname rp2040EnterFlashSafeSection runtime.rp2040EnterFlashSafeSection
func rp2040EnterFlashSafeSection() (interrupt.State, bool)

//go:linkname rp2040ExitFlashSafeSection runtime.rp2040ExitFlashSafeSection
func rp2040ExitFlashSafeSection(state interrupt.State, multicore bool)
