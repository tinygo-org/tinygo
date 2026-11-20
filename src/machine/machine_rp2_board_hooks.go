//go:build (rp2040 || rp2350) && boardpackage

package machine

// Globals mutated by boardInit0 or boardInit1 must be zero-initialized package
// vars populated inside the hook, because runtime.initAll runs afterwards.

// boardXTAL returns the board crystal frequency in Hz. Only 12 MHz is
// supported today and other values panic in socInit0 before serial is ready.
func boardXTAL() uint32
