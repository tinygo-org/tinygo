//go:build uefi

package uefi

const microsecondsCalibration = 10000 // 10 milliseconds

// CalibrateTimerFrequency calibrates the timer frequency by measuring
// ticks over a known interval. Returns the timer frequency in ticks
// per microsecond.
func CalibrateTimerFrequency() uint64 {
	// Not the most accurate method, but should be good enough for EFI.
	start := Ticks()
	Stall(microsecondsCalibration)
	end := Ticks()

	frequency := (end - start) / microsecondsCalibration // ticks per microsecond
	if frequency == 0 {
		frequency = 1000 // fallback
	}
	return frequency
}
