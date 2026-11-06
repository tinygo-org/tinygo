//go:build esp32s3

package machine

import (
	"runtime/interrupt"
	_ "unsafe"
)

// Implemented in package runtime, which parks the other core in IRAM.

//go:linkname flashSafeEnter runtime.esp32s3EnterFlashSafeSection
func flashSafeEnter() (interrupt.State, bool)

//go:linkname flashSafeExit runtime.esp32s3ExitFlashSafeSection
func flashSafeExit(state interrupt.State, multicore bool)
