//go:build uefi

package runtime

import "machine/uefi"

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

func ticks() timeUnit {
	return timeUnit(uefi.Ticks())
}

func nanosecondsToTicks(ns int64) timeUnit {
	frequency := int64(uefi.TicksFrequency())
	if frequency == 0 {
		return timeUnit(ns)
	}
	return timeUnit(ns * frequency / 1000000000)
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
		runtimePanic("could not allocate initial UEFI heap")
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

//go:noinline
func runMain() {
	run()
}

//export efi_main
func main(imageHandle uintptr, systemTable uintptr) uintptr {
	uefi.Init(imageHandle, systemTable)
	preinit()
	stackTop = getCurrentStackPointer()
	runMain()
	uefi.BS().Exit(uefi.GetImageHandle(), 0, 0, nil)
	return 0
}
