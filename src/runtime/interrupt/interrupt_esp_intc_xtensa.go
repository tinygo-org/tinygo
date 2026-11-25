//go:build esp32 || esp32s3

package interrupt

import "device/xtensa"

// The ESP32 (Xtensa LX6) and ESP32-S3 (Xtensa LX7) interrupt model:
//
//  1. The **interrupt matrix** (DPORT on the ESP32, INTERRUPT_CORE0 on the
//     ESP32-S3) maps each peripheral source to one of 32 CPU interrupt lines
//     via a 5-bit mapping register.
//  2. The CPU's INTENABLE special register (SR 228) enables/disables each of
//     the 32 CPU interrupt lines independently.
//  3. When an enabled CPU interrupt fires, the processor vectors to the
//     level-1 exception vector (offset 0x340 from VECBASE).
//  4. The INTERRUPT special register (SR 226) shows which CPU interrupts are
//     currently pending.
//
// We allocate CPU interrupt lines 1..30 for use by peripherals via
// interrupt.New().  Line 0 is reserved and line 31 is avoided because some
// hardware treats it specially.

const (
	// First / last allocatable CPU interrupt for peripherals.
	firstCPUInt = 1
	lastCPUInt  = 30
)

const maxCores = 2 // the maximum number of cores supported

// inInterrupt is set while we're inside the interrupt handler so that
// interrupt.In() returns the correct value.
var inInterrupt [maxCores]bool

// enableLine enables CPU interrupt line num on the calling core.
func enableLine(num int) {
	enableLocal(1 << uint(num))
}

// enableLocal sets the given bits in INTENABLE of the calling core.
func enableLocal(mask uint32) {
	xtensa.SetINTENABLEBits(mask)
}

// In returns whether the CPU is currently inside an interrupt handler.
func In() bool {
	return inInterrupt[currentCPU()%maxCores]
}

// handleInterrupt is called from the assembly vector code in
// targets/esp32-interrupts.S and targets/esp32s3-interrupts.S.
// It determines which CPU interrupt(s) fired and dispatches to the
// registered Go handlers.
//
//export handleInterrupt
func handleInterrupt() {
	core := currentCPU() % maxCores
	inInterrupt[core] = true

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

	inInterrupt[core] = false
}

// readINTENABLE reads the INTENABLE special register (SR 228).
func readINTENABLE() uint32 {
	return xtensa.ReadINTENABLE()
}

// readINTERRUPT reads the INTERRUPT special register (SR 226), which
// reflects the currently pending CPU interrupts.
func readINTERRUPT() uint32 {
	return xtensa.ReadINTERRUPT()
}

// writeINTCLEAR writes the INTCLEAR special register (SR 227).
// Setting bit N clears CPU interrupt N if it is edge-triggered or
// software-triggered.  Bits corresponding to level-triggered interrupts
// are ignored by hardware.
func writeINTCLEAR(val uint32) {
	xtensa.WriteINTCLEAR(val)
}
