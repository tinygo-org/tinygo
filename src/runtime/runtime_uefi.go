//go:build uefi

package runtime

import (
	"machine/uefi"
)

type WaitForEvents func()

func machineTicks() uint64 {
	return uefi.Ticks()
}

type timeUnit int64

// ticks returns the number of ticks (microseconds) elapsed since power up.
func ticks() timeUnit {
	t := machineTicks()
	return timeUnit(t)
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns * int64(uefi.TicksFrequency()) / 1000000000)
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	frequency := timeUnit(uefi.TicksFrequency())

	//          Ticks
	// Time = --------- x 1,000,000,000
	//        Frequency
	nanoSeconds := (ticks / frequency) * 1000000000
	remainder := ticks % frequency

	// Ensure (Remainder * 1,000,000,000) will not overflow 64-bit.
	// Since 2^29 < 1,000,000,000 = 0x3B9ACA00 < 2^30, Remainder should < 2^(64-30) = 2^34,
	// i.e. highest bit set in Remainder should <= 33.
	//
	shift := highBitSet64(uint64(remainder)) - 32
	if shift < 0 {
		shift = 0
	}
	remainder = remainder >> shift
	frequency = frequency >> shift
	nanoSeconds += remainder * 1000000000 / frequency

	return int64(nanoSeconds)
}

func highBitSet64(operand uint64) int {
	if operand == (operand & 0xffffffff) {
		return highBitSet32(uint32(operand))
	}
	return highBitSet32(uint32(operand>>32)) + 32
}

func highBitSet32(operand uint32) int {
	if operand == 0 {
		return -1
	}
	bitIndex := 32
	for operand > 0 {
		bitIndex--
		operand <<= 1
	}
	return bitIndex
}

func sleepTicks(d timeUnit) {
	if d == 0 {
		return
	}

	sleepUntil := ticks() + d
	for ticks() < sleepUntil {
		uefi.CpuPause()
	}
}

func putchar(c byte) {
	buf := [2]uefi.CHAR16{uefi.CHAR16(c), 0}
	st := uefi.ST()
	st.ConOut.OutputString(&buf[0])
}

func exit(code int) {
	uefi.BS().Exit(uefi.GetImageHandle(), uefi.EFI_STATUS(code), 0, nil)
}

func abort() {
	uefi.BS().Exit(uefi.GetImageHandle(), uefi.EFI_ABORTED, 0, nil)
}

//go:linkname procPin sync/atomic.runtime_procPin
func procPin() {
}

//go:linkname procUnpin sync/atomic.runtime_procUnpin
func procUnpin() {
}

var heapSize uintptr = 128 * 1024 // small amount to start
var heapMaxSize uintptr

var heapStart, heapEnd uintptr

var stackTop uintptr

var allocatePagesAddress uefi.EFI_PHYSICAL_ADDRESS

func preinit() {
	// status is must be register
	var status uefi.EFI_STATUS

	heapMaxSize = 1024 * 1024 * 1024 // 1G for the entire heap

	bs := uefi.BS()
	for heapMaxSize > 16*1024*1024 {
		status = bs.AllocatePages(uefi.AllocateAnyPages, uefi.EfiLoaderData, uefi.UINTN(heapMaxSize)/4096, &allocatePagesAddress)
		if status != uefi.EFI_OUT_OF_RESOURCES {
			heapStart = uintptr(allocatePagesAddress)
			break
		}
		heapMaxSize /= 2
	}
	if status != 0 {
		uefi.DebugPrint("AllocatePool failed", uint64(status))
		return
	}
	heapEnd = heapStart + heapSize
}

// growHeap tries to grow the heap size. It returns true if it succeeds, false
// otherwise.
func growHeap() bool {
	if heapSize == heapMaxSize {
		// Already at the max. If we run out of memory, we should consider
		// increasing heapMaxSize on 64-bit systems.
		return false
	}
	// Grow the heap size used by the program.
	heapSize = (heapSize * 4 / 3) &^ 4095 // grow by around 33%
	if heapSize > heapMaxSize {
		heapSize = heapMaxSize
	}
	setHeapEnd(heapStart + heapSize)
	return true
}

func init() {
	var efiTime uefi.EFI_TIME

	mono := nanotime()
	efiTime, status := uefi.GetTime()
	if status == uefi.EFI_SUCCESS {
		sec, nsec := efiTime.GetEpoch()
		timeOffset = sec*1000000000 + int64(nsec) - mono
	}
}

var waitForEventsFunction WaitForEvents = func() {
	uefi.CpuPause()
}

// SetWaitForEvents
// You can implement your own event-loop with BS.CheckEvent or BS.WaitForEvents.
func SetWaitForEvents(f WaitForEvents) {
	waitForEventsFunction = f
}

func waitForEvents() {
	waitForEventsFunction()
}

// Must be a separate function to get the correct stack pointer.
//
//go:noinline
func runMain() {
	run()
}

//export efi_main
func main(imageHandle uintptr, systemTable uintptr) uintptr {
	uefi.Init(imageHandle, systemTable)

	preinit()

	// Obtain the initial stack pointer right before calling the run() function.
	// The run function has been moved to a separate (non-inlined) function so
	// that the correct stack pointer is read.
	stackTop = getCurrentStackPointer()
	runMain()

	// For libc compatibility.
	return 0
}
