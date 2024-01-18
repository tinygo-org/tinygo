//go:build nintendoswitch

package runtime

import "unsafe"

type timeUnit int64

const (
	// Handles
	infoTypeTotalMemorySize = 6          // Total amount of memory available for process.
	infoTypeUsedMemorySize  = 7          // Amount of memory currently used by process.
	currentProcessHandle    = 0xFFFF8001 // Pseudo handle for the current process.

	// Types of config Entry
	envEntryTypeEndOfList        = 0 // Entry list terminator.
	envEntryTypeMainThreadHandle = 1 // Provides the handle to the main thread.
	envEntryTypeOverrideHeap     = 3 // Provides heap override information.

	// Default heap size allocated by libnx
	defaultHeapSize = 0x2000000 * 16

	debugInit = false
)

//go:extern _saved_return_address
var savedReturnAddress uintptr

//export __stack_top
var stackTop uintptr

//go:extern _context
var context uintptr

//go:extern _main_thread
var mainThread uintptr

var (
	heapStart = uintptr(0)
	heapEnd   = uintptr(0)
	usedRam   = uint64(0)
	totalRam  = uint64(0)
	totalHeap = uint64(0)
)

func preinit() {
	// Unsafe to use heap here
	setupEnv()
	setupHeap()
}

// Entry point for Go. Initialize all packages and call main.main().
//
//export main
func main() {
	preinit()
	run()

	// Call exit to correctly finish the program
	// Without this, the application crashes at start, not sure why
	for {
		exit(0)
	}
}

// sleepTicks sleeps for the specified system ticks
func sleepTicks(d timeUnit) {
	svcSleepThread(uint64(ticksToNanoseconds(d)))
}

// armTicksToNs converts cpu ticks to nanoseconds
// Nintendo Switch CPU ticks has a fixed rate at 19200000
// It is basically 52 ns per tick
// The formula 625 / 12 is equivalent to 1e9 / 19200000
func ticksToNanoseconds(tick timeUnit) int64 {
	return int64(tick * 625 / 12)
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(12 * ns / 625)
}

func ticks() timeUnit {
	return timeUnit(ticksToNanoseconds(timeUnit(getArmSystemTick())))
}

var stdoutBuffer = make([]byte, 120)
var position = 0

func putchar(c byte) {
	if c == '\n' || position >= len(stdoutBuffer) {
		svcOutputDebugString(&stdoutBuffer[0], uint64(position))
		position = 0
		return
	}

	stdoutBuffer[position] = c
	position++
}

func abort() {
	for {
		exit(1)
	}
}

//export write
func libc_write(fd int32, buf *byte, count int) int {
	// TODO: Proper handling write
	for i := 0; i < count; i++ {
		putchar(*buf)
		buf = (*byte)(unsafe.Add(unsafe.Pointer(buf), 1))
	}
	return count
}

// exit checks if a savedReturnAddress were provided by the launcher
// if so, calls the nxExit which restores the stack and returns to launcher
// otherwise just calls systemcall exit
func exit(code int) {
	if savedReturnAddress == 0 {
		svcExitProcess(code)
		return
	}

	nxExit(code, stackTop, savedReturnAddress)
}

type configEntry struct {
	Key   uint32
	Flags uint32
	Value [2]uint64
}

func setupEnv() {
	if debugInit {
		println("Saved Return Address:", savedReturnAddress)
		println("Context:", context)
		println("Main Thread Handle:", mainThread)
	}

	// See https://switchbrew.org/w/index.php?title=Homebrew_ABI
	// Here we parse only the required configs for initializing
	if context != 0 {
		ptr := context
		entry := (*configEntry)(unsafe.Pointer(ptr))
		for entry.Key != envEntryTypeEndOfList {
			switch entry.Key {
			case envEntryTypeOverrideHeap:
				if debugInit {
					println("Got heap override")
				}
				heapStart = uintptr(entry.Value[0])
				heapEnd = heapStart + uintptr(entry.Value[1])
			case envEntryTypeMainThreadHandle:
				mainThread = uintptr(entry.Value[0])
			default:
				if entry.Flags&1 > 0 {
					// Mandatory but not parsed
					runtimePanic("mandatory config entry not parsed")
				}
			}
			ptr += unsafe.Sizeof(configEntry{})
			entry = (*configEntry)(unsafe.Pointer(ptr))
		}
	}
	// Fetch used / total RAM for allocating HEAP
	svcGetInfo(&totalRam, infoTypeTotalMemorySize, currentProcessHandle, 0)
	svcGetInfo(&usedRam, infoTypeUsedMemorySize, currentProcessHandle, 0)
}

func setupHeap() {
	if heapStart != 0 {
		if debugInit {
			print("Heap already overrided by hblauncher")
		}
		// Already overrided
		return
	}

	if debugInit {
		print("No heap override. Using normal initialization")
	}

	size := uint32(defaultHeapSize)

	if totalRam > usedRam+0x200000 {
		// Get maximum possible heap
		size = uint32(totalRam-usedRam-0x200000) & ^uint32(0x1FFFFF)
	}

	if size < defaultHeapSize {
		size = defaultHeapSize
	}

	if debugInit {
		println("Trying to allocate", size, "bytes of heap")
	}

	svcSetHeapSize(&heapStart, uint64(size))

	if heapStart == 0 {
		runtimePanic("failed to allocate heap")
	}

	totalHeap = uint64(size)

	heapEnd = heapStart + uintptr(size)

	if debugInit {
		println("Heap Start", heapStart)
		println("Heap End  ", heapEnd)
		println("Total Heap", totalHeap)
	}
}

// growHeap tries to grow the heap size. It returns true if it succeeds, false
// otherwise.
func growHeap() bool {
	// Growing the heap is unimplemented.
	return false
}

// getHeapBase returns the start address of the heap
// this is externally linked by gonx
func getHeapBase() uintptr {
	return heapStart
}

// getHeapEnd returns the end address of the heap
// this is externally linked by gonx
func getHeapEnd() uintptr {
	return heapEnd
}

//go:extern __data_start
var dataStartSymbol [0]byte

//go:extern __data_end
var dataEndSymbol [0]byte

//go:extern __bss_start
var bssStartSymbol [0]byte

//go:extern __bss_end
var bssEndSymbol [0]byte

// Find global variables.
// The linker script provides __*_start and __*_end symbols that can be used to
// scan the given sections. They are already aligned so don't need to be
// manually aligned here.
func findGlobals(found func(start, end uintptr)) {
	dataStart := uintptr(unsafe.Pointer(&dataStartSymbol))
	dataEnd := uintptr(unsafe.Pointer(&dataEndSymbol))
	found(dataStart, dataEnd)
	bssStart := uintptr(unsafe.Pointer(&bssStartSymbol))
	bssEnd := uintptr(unsafe.Pointer(&bssEndSymbol))
	found(bssStart, bssEnd)
}

// getContextPtr returns the hblauncher context
// this is externally linked by gonx
func getContextPtr() uintptr {
	return context
}

// getMainThreadHandle returns the main thread handler if any
// this is externally linked by gonx
func getMainThreadHandle() uintptr {
	return mainThread
}

//export armGetSystemTick
func getArmSystemTick() int64

// nxExit exits the program to homebrew launcher
//
//export __nx_exit
func nxExit(code int, stackTop uintptr, exitFunction uintptr)

// Horizon System Calls
// svcSetHeapSize Set the process heap to a given size. It can both extend and shrink the heap.
// svc 0x01
//
//export svcSetHeapSize
func svcSetHeapSize(addr *uintptr, length uint64) uint64

// svcExitProcess Exits the current process.
// svc 0x07
//
//export svcExitProcess
func svcExitProcess(code int)

// svcSleepThread Sleeps the current thread for the specified amount of time.
// svc 0x0B
//
//export svcSleepThread
func svcSleepThread(nanos uint64)

// svcOutputDebugString Outputs debug text, if used during debugging.
// svc 0x27
//
//export svcOutputDebugString
func svcOutputDebugString(str *uint8, size uint64) uint64

// svcGetInfo Retrieves information about the system, or a certain kernel object.
// svc 0x29
//
//export svcGetInfo
func svcGetInfo(output *uint64, id0 uint32, handle uint32, id1 uint64) uint64

func hardwareRand() (n uint64, ok bool) {
	// TODO: see whether there is a RNG and use it.
	return 0, false
}
