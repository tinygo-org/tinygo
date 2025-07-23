//go:build uefi

package runtime

import (
	"machine/uefi"
	"unsafe"
)

type WaitForEvents func()

// ticks returns the number of ticks (microseconds) elapsed since power up.
func ticks() timeUnit {
	t := uefi.Ticks()
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

// Try for 256MiB, which should be reasonable for an amd64 system running UEFI
// We use growHeap() to do the initial allocation which tries for
// heapSize*4/3, thus 192 * 4 / 3 = 256.
var heapSize uintptr = 192 * 1024 * 1024

var heapStart, heapEnd uintptr

var stackTop uintptr

var allocatePagesAddress uefi.EFI_PHYSICAL_ADDRESS

func preinit() {
	// always disable watchdog; if the user wants it they can turn it back on
	uefi.ST().BootServices.SetWatchdogTimer(0, 0, 0, nil)

	// first time allocating heap
	if !growHeap() {
		panic("couldn't allocate heap")
	}
}

// growHeap tries to grow the heap size. It returns true if it succeeds, false
// otherwise.
//
// Current implementation is flawed in that once we've allocated more than half the memory
// we cannot grow any farther as a big enough contiguous chunk is no longer available.
// Additionally, UEFI on real hardware, in contrast to a qemu virtual machine running
// EDK2 Tiano, seems to have much more memory fragmentation. In some cases, allocating
// just a quarter of total available RAM will fail.
//
// TODO: Consider using GetMemoryMap to locate the largest chunk of EfiConventionalMemory.
func growHeap() bool {
	// try a 33% bigger heap, page aligned
	newHeapSize := ((heapSize * 4) / 3) &^ 4095

	bs := uefi.BS()
	var status uefi.EFI_STATUS
	for newHeapSize >= heapSize {
		pages := newHeapSize / 4096
		status = bs.AllocatePages(
			uefi.AllocateAnyPages,
			uefi.EfiLoaderData,
			uefi.UINTN(pages),
			&allocatePagesAddress)
		if status == uefi.EFI_SUCCESS {
			heapStart = uintptr(allocatePagesAddress)
			heapSize = newHeapSize
			setHeapEnd(heapStart + heapSize)
			return true
		}
		if status != uefi.EFI_OUT_OF_RESOURCES {
			uefi.DebugPrint("AllocatePages failed", uint64(status))
			return false
		}
		newHeapSize /= 2
	}

	return false
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

	if heapStart != 0 {
		uefi.ST().BootServices.FreePool((*uefi.VOID)(unsafe.Pointer(heapStart)))
	}

	uefi.BS().Exit(uefi.GetImageHandle(), 0, 0, nil)

	// For libc compatibility.
	return 0
}

// lie and say we have a byte ready
func buffered() int { return 1 }

// getchar blocks trying to get a KeyStroke event
func getchar() byte {
	conIn := uefi.ST().ConIn
	var key uefi.EFI_INPUT_KEY
	for {
		if conIn.ReadKeyStroke(&key) == uefi.EFI_SUCCESS {
			if key.UnicodeChar != 0 {
				return byte(key.UnicodeChar)
			}
		}
		Gosched()
	}
}
