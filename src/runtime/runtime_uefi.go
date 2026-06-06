//go:build uefi

package runtime

import (
	"machine/uefi"
	"unsafe"
)

var timerFrequency uint64 // Timer frequency in ticks per microsecond

// peImageBase is the base address of the loaded PE image in memory.
// __ImageBase is a synthetic symbol provided by LLD (the PE/COFF linker)
// that points to the start of the PE image headers (the DOS "MZ" stub).
// It is always available in PE/COFF binaries and requires no linker script.
//
//go:extern __ImageBase
var peImageBase [0]byte

func preinit() {
	// Fix stackTop: baremetal.go sets it to the ADDRESS of _stack_top,
	// but _stack_top is a variable storing the initial RSP value.
	stackTop = *(*uintptr)(unsafe.Pointer(&stackTopSymbol))

	base := uintptr(unsafe.Pointer(&peImageBase))

	// Set globalsStart/globalsEnd by parsing PE/COFF headers.
	findGlobalsFromPE((*dosHeader)(unsafe.Pointer(base)), func(start, end uintptr) {
		if globalsStart == 0 {
			globalsStart = start
			globalsEnd = end
		}
	})

	// Allocate heap dynamically via UEFI Boot Services
	allocateUEFIHeap()
	// Calibrate timer frequency
	timerFrequency = uefi.CalibrateTimerFrequency()
}

// variables defined as globals and used before the heap is allocated
var (
	// memMapBuffer is a static byte buffer for GetMemoryMap (used before heap).
	// 256 * 48 = 12288 bytes, enough for ~256 descriptors at typical OVMF descSize=48.
	memMapBuffer    [256 * 48]byte
	memDescSize     uintptr
	memMapSize      uintptr = uintptr(len(memMapBuffer))
	allocatedMemory uintptr
)

// allocateUEFIHeap allocates heap memory using UEFI AllocatePages.
// It queries the memory map to determine available memory, allocates the
// largest region found.
func allocateUEFIHeap() {
	count := uefi.GetMemoryMap(memMapBuffer[:], &memMapSize, &memDescSize)
	if count == 0 {
		runtimePanic("failed to get UEFI memory map")
	}

	var pages uintptr
	var heapAllocatedPages uintptr

	// Find the largest contiguous EfiConventionalMemory region
	for i := 0; i < count; i++ {
		desc := uefi.MemMapEntry(memMapBuffer[:], i, memDescSize)
		if uefi.MemoryType(desc.Type) == uefi.EfiConventionalMemory {
			n := uintptr(desc.NumberOfPages)
			if n > pages {
				pages = n
				allocatedMemory = uintptr(desc.PhysicalStart)
			}
		}
	}

	if pages == 0 || allocatedMemory == 0 {
		runtimePanic("no conventional memory region found for heap allocation")
	} else if pages > 0 {
		heapAllocatedPages = pages
	}

	// Try to allocate at the known region address first
	if allocatedMemory != 0 {
		heapStart = uefi.AllocatePages(
			uefi.AllocateAddress,
			uefi.EfiLoaderData,
			heapAllocatedPages,
			&allocatedMemory,
		)
	}

	// Fall back to any-pages allocation, halving on failure
	for heapStart == 0 {
		heapStart = uefi.AllocatePages(
			uefi.AllocateAnyPages,
			uefi.EfiLoaderData,
			heapAllocatedPages,
			&allocatedMemory,
		)
		if heapStart == 0 {
			heapAllocatedPages /= 2
		}
	}

	if heapStart == 0 {
		runtimePanic("failed to allocate heap via UEFI AllocatePages")
	}

	heapEnd = heapStart + (heapAllocatedPages * uefi.PageSize)
}

// called from scheduler by initAll
func init() {
	// Disable UEFI watchdog timer
	uefi.DisableWatchdog()
	// Initialize clock time
	initClockTime()
}

//export main
func main() {
	preinit()
	run()
	exit(0)
}

// putchar outputs a single character via UEFI ConOut.
func putchar(c byte) {
	var buf [2]uint16

	buf[0] = uint16(c)
	buf[1] = 0

	uefi.OutputString(&buf[0])
}

// getchar reads a single byte from the keyboard.
func getchar() byte {
	for {
		if b, ok := uefi.KeyBufferPop(); ok {
			return b
		}
		if uefi.IsKeyPressed() {
			uefi.ReadKey()
		}
		Gosched()
	}
}

func buffered() int {
	if uefi.IsKeyPressed() {
		uefi.ReadKey()
	}
	return uefi.KeyBufferAvailable()
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	if timerFrequency == 0 {
		return int64(ticks)
	}
	return int64(ticks) * 1000 / int64(timerFrequency)
}

func nanosecondsToTicks(ns int64) timeUnit {
	if timerFrequency == 0 {
		return timeUnit(ns)
	}
	return timeUnit(ns / 1000 * int64(timerFrequency))
}

func initClockTime() {
	mono := nanotime()

	var t uefi.Time
	var sec int64
	var nsec int64

	status := uefi.GetTime(&t)
	if status != 0 {
		sec = mono / (1000 * 1000 * 1000)
		nsec = mono - sec*(1000*1000*1000)
	} else {
		sec = t.Timestamp()
		nsec = int64(t.Nanosecond)
	}

	timeOffset.Store(sec*1000000000 + nsec - mono)
}

func ticks() timeUnit {
	return timeUnit(uefi.Ticks())
}

func sleepTicks(d timeUnit) {
	if d <= 0 {
		return
	}

	target := ticks() + d
	for ticks() < target {
		uefi.CpuPause()
	}
}

func exit(code int) {
	uefi.Exit(code)
}

func abort() {
	exit(1)
}

func hardwareRand() (uint64, bool) {
	if !uefi.HasRNGSupport() {
		return 0, false
	}
	return uefi.ReadRandom()
}

// TinyGo does not support any form of parallelism on UEFI, so these can
// be left empty.

//go:linkname procPin sync/atomic.runtime_procPin
func procPin() {
}

//go:linkname procUnpin sync/atomic.runtime_procUnpin
func procUnpin() {
}
