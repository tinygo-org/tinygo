//go:build uefi

package runtime

import "device/uefi"

const zeroSizeAllocPtr uintptr = 16

//go:linkname procPin sync/atomic.runtime_procPin
func procPin() {
}

//go:linkname procUnpin sync/atomic.runtime_procUnpin
func procUnpin() {
}

var heapSize uintptr = 64 * 1024 * 1024
var heapStart, heapEnd uintptr
var stackTop uintptr
var allocatePagesAddress uefi.EFI_PHYSICAL_ADDRESS
var waitForEventsFunction = func() {
	uefi.CpuPause()
}

func ticks() timeUnit {
	return timeUnit(uefi.Ticks())
}

func nanosecondsToTicks(ns int64) timeUnit {
	frequency := int64(uefi.TicksFrequency())
	if frequency == 0 {
		return timeUnit(ns)
	}
	seconds := ns / 1000000000
	remainder := ns % 1000000000
	return timeUnit(seconds*frequency + (remainder*frequency)/1000000000)
}

func ticksToNanoseconds(t timeUnit) int64 {
	frequency := int64(uefi.TicksFrequency())
	if frequency == 0 {
		return int64(t)
	}
	return int64(t) * 1000000000 / frequency
}

func sleepTicks(d timeUnit) {
	if d == 0 {
		return
	}
	end := ticks() + d
	for ticks() < end {
		uefi.CpuPause()
	}
}

func putchar(c byte) {
	if c == '\n' {
		buf := [2]uefi.CHAR16{uefi.CHAR16('\r'), 0}
		uefi.ST().ConOut.OutputString(&buf[0])
	}
	buf := [2]uefi.CHAR16{uefi.CHAR16(c), 0}
	uefi.ST().ConOut.OutputString(&buf[0])
}

func exit(code int) {
	uefi.BS().Exit(uefi.GetImageHandle(), uefi.EFI_STATUS(code), 0, nil)
}

func abort() {
	uefi.BS().Exit(uefi.GetImageHandle(), uefi.EFI_ABORTED, 0, nil)
}

func preinit() {
	uefi.BS().SetWatchdogTimer(0, 0, 0, nil)
	if !growHeap() {
		runtimeFatal("could not allocate initial UEFI heap")
	}
}

func growHeap() bool {
	newHeapSize := ((heapSize * 4) / 3) &^ 4095
	for newHeapSize >= heapSize {
		pages := newHeapSize / 4096
		status := uefi.BS().AllocatePages(
			uefi.AllocateAnyPages,
			uefi.EfiLoaderData,
			uefi.UINTN(pages),
			&allocatePagesAddress,
		)
		if status == uefi.EFI_SUCCESS {
			heapStart = uintptr(allocatePagesAddress)
			heapSize = newHeapSize
			setHeapEnd(heapStart + heapSize)
			return true
		}
		if status != uefi.EFI_OUT_OF_RESOURCES {
			return false
		}
		newHeapSize /= 2
	}
	return false
}

func init() {
	mono := nanotime()
	efiTime, status := uefi.GetTime()
	if status == uefi.EFI_SUCCESS {
		sec, nsec := efiTime.GetEpoch()
		timeOffset.Store(sec*1000000000 + int64(nsec) - mono)
	}
}

func SetWaitForEvents(f func()) {
	waitForEventsFunction = f
}

func waitForEvents() {
	waitForEventsFunction()
}

//go:noinline
func runMain() {
	run()
}

func buffered() int {
	return 0
}

func getchar() byte {
	return 0
}

//export efi_main
func main(imageHandle uintptr, systemTable uintptr) uintptr {
	uefi.Init(imageHandle, systemTable)
	preinit()
	stackTop = getCurrentStackPointer()
	runMain()

	if heapStart != 0 {
		uefi.BS().FreePages(uefi.EFI_PHYSICAL_ADDRESS(heapStart), uefi.UINTN(heapSize/4096))
	}

	return 0
}
