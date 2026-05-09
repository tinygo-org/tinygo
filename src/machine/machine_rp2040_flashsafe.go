//go:build tinygo && rp2040

package machine

import (
	"runtime/interrupt"
	_ "unsafe"
)

//go:linkname rp2040EnterFlashSafeSection runtime.rp2040EnterFlashSafeSection
func rp2040EnterFlashSafeSection() interrupt.State

//go:linkname rp2040ExitFlashSafeSection runtime.rp2040ExitFlashSafeSection
func rp2040ExitFlashSafeSection(state interrupt.State)
