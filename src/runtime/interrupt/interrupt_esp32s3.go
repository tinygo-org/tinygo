//go:build esp32s3

package interrupt

import (
	"device"
	"device/esp"
	"runtime/volatile"
	"unsafe"
)

// State represents the previous global interrupt state.
type State uintptr

// Disable disables all interrupts and returns the previous interrupt state. It
// can be used in a critical section like this:
//
//	state := interrupt.Disable()
//	// critical section
//	interrupt.Restore(state)
//
// Critical sections can be nested. Make sure to call Restore in the same order
// as you called Disable (this happens naturally with the pattern above).
func Disable() (state State) {
	return State(device.AsmFull("rsil {}, 15", nil))
}

// Restore restores interrupts to what they were before. Give the previous state
// returned by Disable as a parameter. If interrupts were disabled before
// calling Disable, this will not re-enable interrupts, allowing for nested
// critical sections.
func Restore(state State) {
	device.AsmFull("wsr {state}, PS", map[string]interface{}{
		"state": state,
	})
}

// The ESP32-S3 (Xtensa LX7) interrupt model:
//
//  1. The **interrupt matrix** (INTERRUPT_CORE0) maps each peripheral source
//     (0-98) to one of 32 CPU interrupt lines via a 5-bit mapping register.
//  2. The CPU's INTENABLE special register (SR 228) enables/disables each of
//     the 32 CPU interrupt lines independently.
//  3. When an enabled CPU interrupt fires, the processor vectors to the
//     level-1 exception vector (offset 0x180 from VECBASE).
//  4. The INTERRUPT special register (SR 226) shows which CPU interrupts are
//     currently pending.
//
// We allocate CPU interrupt lines 6..30 for use by peripherals via
// interrupt.New().  Lines 0-5 are reserved (timer, software, etc.) and
// line 31 is avoided because some hardware treats it specially.

const (
	// First / last allocatable CPU interrupt for peripherals.
	firstCPUInt = 6
	lastCPUInt  = 30
)

// cpuIntUsed tracks which CPU interrupt lines have been allocated.
var cpuIntUsed [32]bool

// cpuIntToPeripheral maps CPU interrupt number → peripheral IRQ source,
// so that handleInterrupt can dispatch to the correct Go handler.
var cpuIntToPeripheral [32]int

// inInterrupt is set while we're inside the interrupt handler so that
// interrupt.In() returns the correct value.
var inInterrupt bool

// Enable enables a CPU interrupt for the ESP32-S3.  The caller must first
// map the peripheral to a CPU interrupt line using the interrupt matrix,
// e.g.:
//
//	esp.INTERRUPT_CORE0.SetGPIO_INTERRUPT_PRO_MAP(cpuInt)
//	interrupt.New(cpuInt, handler).Enable()
func (i Interrupt) Enable() error {
	if i.num < firstCPUInt || i.num > lastCPUInt {
		return errInterruptRange
	}

	// Mark as used.
	cpuIntUsed[i.num] = true

	// Read current INTENABLE, set the bit for this CPU interrupt.
	cur := readINTENABLE()
	cur |= 1 << uint(i.num)
	writeINTENABLE(cur)

	return nil
}

// In returns whether the CPU is currently inside an interrupt handler.
func In() bool {
	return inInterrupt
}

// handleInterrupt is called from the assembly vector code in esp32s3.S.
// It determines which CPU interrupt(s) fired and dispatches to the
// registered Go handlers.
//
//export handleInterrupt
func handleInterrupt() {
	inInterrupt = true

	// INTERRUPT register shows pending + enabled CPU interrupts.
	pending := readINTERRUPT()
	enabled := readINTENABLE()
	active := pending & enabled

	// Clear edge-triggered pending bits before dispatching handlers so that
	// new edges arriving during handler execution are not lost.  Writing to
	// INTCLEAR is a no-op for level-triggered lines, so this is safe for all
	// interrupt types.
	writeINTCLEAR(active)

	for i := firstCPUInt; i <= lastCPUInt; i++ {
		if active&(1<<uint(i)) != 0 {
			// callHandlers requires a compile-time constant, so we
			// dispatch through a switch.
			callHandler(i)
		}
	}

	// Signal to sleepTicks that an interrupt has occurred.
	signalInterrupt()

	inInterrupt = false
}

//go:inline
func callHandler(n int) {
	switch n {
	case 6:
		callHandlers(6)
	case 7:
		callHandlers(7)
	case 8:
		callHandlers(8)
	case 9:
		callHandlers(9)
	case 10:
		callHandlers(10)
	case 11:
		callHandlers(11)
	case 12:
		callHandlers(12)
	case 13:
		callHandlers(13)
	case 14:
		callHandlers(14)
	case 15:
		callHandlers(15)
	case 16:
		callHandlers(16)
	case 17:
		callHandlers(17)
	case 18:
		callHandlers(18)
	case 19:
		callHandlers(19)
	case 20:
		callHandlers(20)
	case 21:
		callHandlers(21)
	case 22:
		callHandlers(22)
	case 23:
		callHandlers(23)
	case 24:
		callHandlers(24)
	case 25:
		callHandlers(25)
	case 26:
		callHandlers(26)
	case 27:
		callHandlers(27)
	case 28:
		callHandlers(28)
	case 29:
		callHandlers(29)
	case 30:
		callHandlers(30)
	}
}

// callHandlers dispatches to registered interrupt handlers for a given
// interrupt number.
//
//go:linkname callHandlers runtime/interrupt.callHandlers
func callHandlers(num int)

//go:linkname signalInterrupt runtime.signalInterrupt
func signalInterrupt()

var errInterruptRange = constError("interrupt for ESP32-S3 must be in range 6 through 30")

type constError string

func (e constError) Error() string {
	return string(e)
}

// readINTENABLE reads the INTENABLE special register (SR 228).
func readINTENABLE() uint32 {
	return uint32(device.AsmFull("rsr {}, INTENABLE", nil))
}

// writeINTENABLE writes the INTENABLE special register (SR 228).
func writeINTENABLE(val uint32) {
	device.AsmFull("wsr {val}, INTENABLE", map[string]interface{}{
		"val": val,
	})
}

// readINTERRUPT reads the INTERRUPT special register (SR 226), which
// reflects the currently pending CPU interrupts.
func readINTERRUPT() uint32 {
	return uint32(device.AsmFull("rsr {}, INTERRUPT", nil))
}

// writeINTCLEAR writes the INTCLEAR special register (SR 227).
// Setting bit N clears CPU interrupt N if it is edge-triggered or
// software-triggered.  Bits corresponding to level-triggered interrupts
// are ignored by hardware.
func writeINTCLEAR(val uint32) {
	device.AsmFull("wsr {val}, INTCLEAR", map[string]interface{}{
		"val": val,
	})
}

// -- Interrupt matrix helpers -----------------------------------------------
// The ESP32-S3 interrupt matrix has one mapping register per peripheral
// source.  These are memory-mapped in the INTERRUPT_CORE0 peripheral.
// The mapping register for peripheral source N is at:
//   base + N*4  (where base = &INTERRUPT_CORE0.PRO_MAC_INTR_MAP)
//
// We provide helpers to set/get the mapping for any source number.

// mapPeripheralToInt routes peripheral IRQ source `src` to CPU interrupt
// `cpuInt` via the interrupt matrix.
func mapPeripheralToInt(src int, cpuInt int) {
	base := unsafe.Pointer(&esp.INTERRUPT_CORE0.PRO_MAC_INTR_MAP)
	reg := (*volatile.Register32)(unsafe.Add(base, uintptr(src)*4))
	reg.Set(uint32(cpuInt))
}
