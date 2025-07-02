package runtime

import "unsafe"

//export abort
func abort()

//export exit
func libc_exit(code int)

//export putchar
func libc_putchar(c int) int

//export VirtualAlloc
func _VirtualAlloc(lpAddress unsafe.Pointer, dwSize uintptr, flAllocationType, flProtect uint32) unsafe.Pointer

//export QueryUnbiasedInterruptTime
func _QueryUnbiasedInterruptTime(UnbiasedTime *uint64) bool

// The parameter is really a LPFILETIME, but *uint64 should be compatible.
//
//export GetSystemTimeAsFileTime
func _GetSystemTimeAsFileTime(lpSystemTimeAsFileTime *uint64)

//export LoadLibraryExW
func _LoadLibraryExW(lpLibFileName *uint16, hFile uintptr, dwFlags uint32) uintptr

//export Sleep
func _Sleep(milliseconds uint32)

const _LOAD_LIBRARY_SEARCH_SYSTEM32 = 0x00000800

//export GetProcAddress
func getProcAddress(handle uintptr, procname *byte) uintptr

//export _configure_narrow_argv
func _configure_narrow_argv(int32) int32

//export __p___argc
func __p___argc() *int32

//export __p___argv
func __p___argv() **unsafe.Pointer

type startupInfo struct {
	newMode int32
}

//export __getmainargs
func __getmainargs(argc *int32, argv, env **unsafe.Pointer, doWildcard int, startInfo *startupInfo) int32

var performanceFrequency int64

//export mainCRTStartup
func mainCRTStartup() int {
	preinit()

	// Obtain the (constant) performance frequency when needed.
	if GOARCH == "386" {
		_QueryPerformanceFrequency(&performanceFrequency)
	}

	// Obtain the initial stack pointer right before calling the run() function.
	// The run function has been moved to a separate (non-inlined) function so
	// that the correct stack pointer is read.
	stackTop = getCurrentStackPointer()
	runMain()

	// Exit via exit(0) instead of returning. This matches
	// mingw-w64-crt/crt/crtexe.c, which exits using exit(0) instead of
	// returning the return value.
	// Exiting this way (instead of returning) also fixes an issue where not all
	// output would be sent to stdout before exit.
	// See: https://github.com/tinygo-org/tinygo/pull/4589
	libc_exit(0)

	// Unreachable, since we've already exited. But we need to return something
	// here to make this valid Go code.
	return 0
}

// Must be a separate function to get the correct stack pointer.
//
//go:noinline
func runMain() {
	run()
}

var args []string

//go:linkname os_runtime_args os.runtime_args
func os_runtime_args() []string {
	if args == nil {
		// Obtain argc/argv from the environment.
		var argc int32
		var argv *unsafe.Pointer
		if GOARCH == "386" {
			// MSVCRT.DLL
			var env *unsafe.Pointer
			startInfo := startupInfo{newMode: 0}
			__getmainargs(&argc, &argv, &env, 1, &startInfo)
		} else {
			// UCRT
			_configure_narrow_argv(2)
			argc = *__p___argc()
			argv = *__p___argv()
		}

		// Make args slice big enough so that it can store all command line
		// arguments.
		args = make([]string, argc)

		// Initialize command line parameters.
		for i := 0; i < int(argc); i++ {
			// Convert the C string to a Go string.
			length := strlen(*argv)
			arg := (*_string)(unsafe.Pointer(&args[i]))
			arg.length = length
			arg.ptr = (*byte)(*argv)
			// This is the Go equivalent of "argv++" in C.
			argv = (*unsafe.Pointer)(unsafe.Add(unsafe.Pointer(argv), unsafe.Sizeof(argv)))
		}
	}
	return args
}

func putchar(c byte) {
	libc_putchar(int(c))
}

var heapSize uintptr = 128 * 1024 // small amount to start
var heapMaxSize uintptr

var heapStart, heapEnd uintptr

func preinit() {
	// Allocate a large chunk of virtual memory. Because it is virtual, it won't
	// really be allocated in RAM. Memory will only be allocated when it is
	// first touched.
	heapMaxSize = 1 * 1024 * 1024 * 1024 // 1GB for the entire heap
	const (
		MEM_COMMIT     = 0x00001000
		MEM_RESERVE    = 0x00002000
		PAGE_READWRITE = 0x04
	)
	heapStart = uintptr(_VirtualAlloc(nil, heapMaxSize, MEM_COMMIT|MEM_RESERVE, PAGE_READWRITE))
	heapEnd = heapStart + heapSize
}

var stackTop uintptr

func ticksToNanoseconds(ticks timeUnit) int64 {
	// Interrupt time count works in units of 100 nanoseconds.
	return int64(ticks) * 100
}

func nanosecondsToTicks(ns int64) timeUnit {
	// Interrupt time count works in units of 100 nanoseconds.
	return timeUnit(ns) / 100
}

func sleepTicks(d timeUnit) {
	// Calculate milliseconds from ticks (which have a resolution of 100ns),
	// rounding up.
	milliseconds := int64(d+9_999) / 10_000
	for milliseconds != 0 {
		duration := uint32(milliseconds)
		_Sleep(duration)
		milliseconds -= int64(duration)
	}
}

//export QueryPerformanceFrequency
func _QueryPerformanceFrequency(*int64) bool

//export QueryPerformanceCounter
func _QueryPerformanceCounter(*int64) bool

func ticks() timeUnit {
	if GOARCH == "386" {
		// Unfortunately QueryUnbiasedInterruptTime is only available starting
		// with Windows 7.

		// Obtain counter (that runs at a fixed frequency).
		var counter int64
		_QueryPerformanceCounter(&counter)
		// Convert this counter to ticks of 100ns (just like
		// QueryUnbiasedInterruptTime).
		// (We could also change the definition of ticks on GOOS=386 but that
		// seems messy).
		return timeUnit((counter * 10000000) / performanceFrequency)
	} else {
		var unbiasedTime uint64
		_QueryUnbiasedInterruptTime(&unbiasedTime)
		return timeUnit(unbiasedTime)
	}
}

//go:linkname now time.now
func now() (sec int64, nsec int32, mono int64) {
	// Get the current time in Windows "file time" format.
	var time uint64
	_GetSystemTimeAsFileTime(&time)

	// Convert file time to Unix time.
	// According to the documentation:
	// > Contains a 64-bit value representing the number of 100-nanosecond
	// > intervals since January 1, 1601 (UTC).
	// We'll convert it to 100 nanosecond intervals starting at 1970.
	const (
		// number of 100-nanosecond intervals in a second
		intervalsPerSecond = 10_000_000
		secondsPerDay      = 60 * 60 * 24
		// Number of days between the Windows epoch (1 january 1601) and the
		// Unix epoch (1 january 1970). Source:
		// https://www.wolframalpha.com/input/?i=days+between+1+january+1601+and+1+january+1970
		days = 134774
	)
	time -= days * secondsPerDay * intervalsPerSecond

	// Convert the time (in 100ns units) to sec/nsec/mono as expected by the
	// time package.
	sec = int64(time / intervalsPerSecond)
	nsec = int32((time - (uint64(sec) * intervalsPerSecond)) * 100)
	mono = ticksToNanoseconds(ticks())
	return
}

//go:linkname syscall_Exit syscall.Exit
func syscall_Exit(code int) {
	libc_exit(code)
}

func growHeap() bool {
	if heapSize == heapMaxSize {
		// Already at the max. If we run out of memory, we should consider
		// increasing heapMaxSize..
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

//go:linkname syscall_loadsystemlibrary syscall.loadsystemlibrary
func syscall_loadsystemlibrary(filename *uint16, absoluteFilepath *uint16) (handle, err uintptr) {
	handle = _LoadLibraryExW(filename, 0, _LOAD_LIBRARY_SEARCH_SYSTEM32)
	if handle == 0 {
		panic("todo: get error")
	}
	return
}

//go:linkname syscall_loadlibrary syscall.loadlibrary
func syscall_loadlibrary(filename *uint16) (handle, err uintptr) {
	panic("todo: syscall.loadlibrary")
}

//go:linkname syscall_getprocaddress syscall.getprocaddress
func syscall_getprocaddress(handle uintptr, procname *byte) (outhandle, err uintptr) {
	outhandle = getProcAddress(handle, procname)
	if outhandle == 0 {
		panic("todo: get error")
	}
	return
}

// TinyGo does not yet support any form of parallelism on Windows, so these can
// be left empty.

//go:linkname procPin sync/atomic.runtime_procPin
func procPin() {
}

//go:linkname procUnpin sync/atomic.runtime_procUnpin
func procUnpin() {
}

func hardwareRand() (n uint64, ok bool) {
	// Use the old RtlGenRandom, introduced in Windows XP.
	// See the rationale in src/crypto/rand/rand_windows.go for why we use this
	// one.
	ok = _RtlGenRandom(unsafe.Pointer(&n), 8)
	return
}

// This function is part of advapi32.dll, and is called SystemFunction036 for
// some reason. It's available on Windows XP and newer.
// See: https://learn.microsoft.com/en-us/windows/win32/api/ntsecapi/nf-ntsecapi-rtlgenrandom
//
//export SystemFunction036
func _RtlGenRandom(buf unsafe.Pointer, len int) bool
