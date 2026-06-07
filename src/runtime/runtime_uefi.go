//go:build uefi

package runtime

import "device/uefi"

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
var consoleInEx *uefi.EFI_SIMPLE_TEXT_INPUT_EX_PROTOCOL
var consoleIn *uefi.EFI_SIMPLE_TEXT_INPUT_PROTOCOL
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
	if protoEx, status := uefi.SimpleTextInExProtocol(); status == uefi.EFI_SUCCESS {
		consoleInEx = protoEx
	}
	if proto, status := uefi.SimpleTextInProtocol(); status == uefi.EFI_SUCCESS {
		consoleIn = proto
	}
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
	if consoleInEx != nil {
		if uefi.BS().CheckEvent(consoleInEx.WaitForKeyEx) == uefi.EFI_SUCCESS {
			return 1
		}
		return 0
	}
	if consoleIn != nil {
		if uefi.BS().CheckEvent(consoleIn.WaitForKey) == uefi.EFI_SUCCESS {
			return 1
		}
		return 0
	}
	return 0
}

func getchar() byte {
	for {
		if consoleInEx != nil {
			key, status := consoleInEx.GetKey()
			if status == uefi.EFI_SUCCESS && key.Key.UnicodeChar != 0 {
				return byte(key.Key.UnicodeChar)
			}
			if status != uefi.EFI_SUCCESS && status != uefi.EFI_NOT_READY && consoleIn == nil {
				return 0
			}
			if status == uefi.EFI_SUCCESS || status == uefi.EFI_NOT_READY {
				continue
			}
		}
		if consoleIn != nil {
			key, status := consoleIn.GetKey()
			if status == uefi.EFI_SUCCESS && key.UnicodeChar != 0 {
				return byte(key.UnicodeChar)
			}
			if status != uefi.EFI_SUCCESS && status != uefi.EFI_NOT_READY {
				return 0
			}
			continue
		}
		return 0
	}
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
