//go:build tinygo && (rp2040 || rp2350)

// "Flash safe" is RP2040/Pico SDK terminology: flash operations must run
// while the other core is not executing from XIP flash.

package machine

import (
	"runtime/interrupt"
	_ "unsafe"
)

// The flash-safe hooks are implemented in package runtime and accessed via
// linkname to avoid an import cycle.

//go:linkname rp2EnterFlashSafeSection runtime.rp2EnterFlashSafeSection
func rp2EnterFlashSafeSection() (interrupt.State, bool)

//go:linkname rp2ExitFlashSafeSection runtime.rp2ExitFlashSafeSection
func rp2ExitFlashSafeSection(state interrupt.State, multicore bool)
