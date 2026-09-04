//go:build darwin || (linux && !baremetal && !wasip1 && !wasm_unknown && !wasip2 && !nintendoswitch)

package runtime

import (
	"internal/futex"
	"internal/task"
	"math/bits"
	"sync/atomic"
	"tinygo"
	"unsafe"
)

//export write
func libc_write(fd int32, buf unsafe.Pointer, count uint) int

//export usleep
func usleep(usec uint) int

//export pause
func pause() int32

// void *mmap(void *addr, size_t length, int prot, int flags, int fd, off_t offset);
// Note: off_t is defined as int64 because:
//   - musl (used on Linux) always defines it as int64
//   - darwin is practically always 64-bit anyway
//
//export mmap
func mmap(addr unsafe.Pointer, length uintptr, prot, flags, fd int, offset int64) unsafe.Pointer

//export abort
func abort()

//export exit
func exit(code int)

//export raise
func raise(sig int32)

//export clock_gettime
func libc_clock_gettime(clk_id int32, ts *timespec)

//export __clock_gettime64
func libc_clock_gettime64(clk_id int32, ts *timespec)

// Portable (64-bit) variant of clock_gettime.
func clock_gettime(clk_id int32, ts *timespec) {
	if TargetBits == 32 {
		// This is a 32-bit architecture (386, arm, etc).
		// We would like to use the 64-bit version of this function so that
		// binaries will continue to run after Y2038.
		// For more information:
		//   - https://musl.libc.org/time64.html
		//   - https://sourceware.org/glibc/wiki/Y2038ProofnessDesign
		libc_clock_gettime64(clk_id, ts)
	} else {
		// This is a 64-bit architecture (amd64, arm64, etc).
		// Use the regular variant, because it already fixes the Y2038 problem
		// by using 64-bit integer types.
		libc_clock_gettime(clk_id, ts)
	}
}

// Note: tv_sec and tv_nsec normally vary in size by platform. However, we're
// using the time64 variant (see clock_gettime above), so the formats are the
// same between 32-bit and 64-bit architectures.
// There is one issue though: on big-endian systems, tv_nsec would be incorrect.
// But we don't support big-endian systems yet (as of 2021) so this is fine.
type timespec struct {
	tv_sec  int64 // time_t with time64 support (always 64-bit)
	tv_nsec int64 // unsigned 64-bit integer on all time64 platforms
}

// Highest address of the stack of the main thread.
var stackTop uintptr

// Entry point for Go. Initialize all packages and call main.main().
//
//export main
func main(argc int32, argv *unsafe.Pointer) int {
	if needsStaticHeap {
		// Allocate area for the heap if the GC needs it.
		allocateHeap()
	}

	// Store argc and argv for later use.
	main_argc = argc
	main_argv = argv

	// Register some fatal signals, so that we can print slightly better error
	// messages.
	tinygo_register_fatal_signals()

	// Obtain the initial stack pointer right before calling the run() function.
	// The run function has been moved to a separate (non-inlined) function so
	// that the correct stack pointer is read.
	stackTop = getCurrentStackPointer()
	runMain()

	// For libc compatibility.
	return 0
}

var (
	main_argc int32
	main_argv *unsafe.Pointer
	args      []string
)

//go:linkname os_runtime_args os.runtime_args
func os_runtime_args() []string {
	if args == nil {
		// Make args slice big enough so that it can store all command line
		// arguments.
		args = make([]string, main_argc)

		// Initialize command line parameters.
		argv := main_argv
		for i := 0; i < int(main_argc); i++ {
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

// Must be a separate function to get the correct stack pointer.
//
//go:noinline
func runMain() {
	run()
}

//export tinygo_register_fatal_signals
func tinygo_register_fatal_signals()

//go:extern tinygo_caught_signal
var tinygo_caught_signal int32

// tinygo_sigpanic is called when a signal (SIGSEGV, SIGFPE, etc.) is caught
// and the signal handler has redirected execution here. It turns the signal
// into a Go panic that can be recovered with recover().
//
//export tinygo_sigpanic
func tinygo_sigpanic() {
	sig := tinygo_caught_signal
	switch sig {
	case sig_SIGSEGV, sig_SIGBUS:
		runtimePanic("nil pointer dereference")
	case sig_SIGFPE:
		runtimePanic("divide by zero")
	default:
		runtimeFatal("signal")
	}
}

// Print fatal errors when they happen, including the instruction location.
// With the particular formatting below, `tinygo run` can extract the location
// where the signal happened and try to show the source location based on DWARF
// information.
//
//export tinygo_handle_fatal_signal
func tinygo_handle_fatal_signal(sig int32, addr uintptr) {
	if panicStrategy() == tinygo.PanicStrategyTrap {
		trap()
	}

	// Print signal including the faulting instruction.
	if addr != 0 {
		printstring("panic: runtime error at ")
		printptr(addr)
	} else {
		printstring("panic: runtime error")
	}
	printstring(": caught signal ")
	switch sig {
	case sig_SIGBUS:
		println("SIGBUS")
	case sig_SIGILL:
		println("SIGILL")
	case sig_SIGSEGV:
		println("SIGSEGV")
	default:
		println(sig)
	}

	// TODO: it might be interesting to also print the invalid address for
	// SIGSEGV and SIGBUS.

	// Do *not* abort here, instead raise the same signal again. The signal is
	// registered with SA_RESETHAND which means it executes only once. So when
	// we raise the signal again below, the signal isn't handled specially but
	// is handled in the default way (probably exiting the process, maybe with a
	// core dump).
	raise(sig)
}

//go:extern environ
var environ *unsafe.Pointer

//go:linkname syscall_runtime_envs syscall.runtime_envs
func syscall_runtime_envs() []string {
	// Count how many environment variables there are.
	env := environ
	numEnvs := 0
	for *env != nil {
		numEnvs++
		env = (*unsafe.Pointer)(unsafe.Add(unsafe.Pointer(env), unsafe.Sizeof(environ)))
	}

	// Create a string slice of all environment variables.
	// This requires just a single heap allocation.
	env = environ
	envs := make([]string, 0, numEnvs)
	for *env != nil {
		ptr := *env
		length := strlen(ptr)
		s := _string{
			ptr:    (*byte)(ptr),
			length: length,
		}
		envs = append(envs, *(*string)(unsafe.Pointer(&s)))
		env = (*unsafe.Pointer)(unsafe.Add(unsafe.Pointer(env), unsafe.Sizeof(environ)))
	}

	return envs
}

func putchar(c byte) {
	buf := [1]byte{c}
	libc_write(1, unsafe.Pointer(&buf[0]), 1)
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	// The OS API works in nanoseconds so no conversion necessary.
	return int64(ticks)
}

func nanosecondsToTicks(ns int64) timeUnit {
	// The OS API works in nanoseconds so no conversion necessary.
	return timeUnit(ns)
}

func sleepTicks(d timeUnit) {
	until := ticks() + d

	for {
		// Sleep for the given amount of time.
		// If a signal arrived before going to sleep, or during the sleep, the
		// sleep will exit early.
		signalFutex.WaitUntil(0, uint64(ticksToNanoseconds(d)))

		// Check whether there was a signal before or during the call to
		// WaitUntil.
		if signalFutex.Swap(0) != 0 {
			if checkSignals() && hasScheduler {
				// We got a signal, so return to the scheduler.
				// (If there is no scheduler, there is no other goroutine that
				// might need to run now).
				return
			}
		}

		// Set duration (in next loop iteration) to the remaining time.
		d = until - ticks()
		if d <= 0 {
			return
		}
	}
}

func getTime(clock int32) uint64 {
	ts := timespec{}
	clock_gettime(clock, &ts)
	return uint64(ts.tv_sec)*1000*1000*1000 + uint64(ts.tv_nsec)
}

// Return monotonic time in nanoseconds.
func monotime() uint64 {
	return getTime(clock_MONOTONIC_RAW)
}

func ticks() timeUnit {
	return timeUnit(monotime())
}

//go:linkname now time.now
func now() (sec int64, nsec int32, mono int64) {
	ts := timespec{}
	clock_gettime(clock_REALTIME, &ts)
	sec = int64(ts.tv_sec)
	nsec = int32(ts.tv_nsec)
	mono = nanotime()
	return
}

//go:linkname syscall_Exit syscall.Exit
func syscall_Exit(code int) {
	exit(code)
}

// TinyGo does not yet support any form of parallelism on an OS, so these can be
// left empty.

//go:linkname procPin sync/atomic.runtime_procPin
func procPin() {
}

//go:linkname procUnpin sync/atomic.runtime_procUnpin
func procUnpin() {
}

var heapSize uintptr = 128 * 1024 // small amount to start
var heapMaxSize uintptr

var heapStart, heapEnd uintptr

func allocateHeap() {
	// Allocate a large chunk of virtual memory. Because it is virtual, it won't
	// really be allocated in RAM. Memory will only be allocated when it is
	// first touched.
	heapMaxSize = 1 * 1024 * 1024 * 1024 // 1GB for the entire heap
	for {
		addr := mmap(nil, heapMaxSize, flag_PROT_READ|flag_PROT_WRITE, flag_MAP_PRIVATE|flag_MAP_ANONYMOUS, -1, 0)
		if addr == unsafe.Pointer(^uintptr(0)) {
			// Heap was too big to be mapped by mmap. Reduce the maximum size.
			// We might want to make this a bit smarter than simply halving the
			// heap size.
			// This can happen on 32-bit systems.
			heapMaxSize /= 2
			if heapMaxSize < 4096 {
				runtimeFatal("cannot allocate heap memory")
			}
			continue
		}
		heapStart = uintptr(addr)
		heapEnd = heapStart + heapSize
		break
	}
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

// Indicate whether signals have been registered.
var hasSignals bool

// Futex for the signal handler.
// The value is 0 when there are no new signals, or 1 when there are unhandled
// signals and the main thread doesn't know about it yet.
// When a signal arrives, the futex value is changed to 1 and if it was 0
// before, all waiters are awoken.
// When a wait exits, the value is changed to 0 and if it wasn't 0 before, the
// signals are checked.
var signalFutex futex.Futex

// Mask of signals that have been received. The signal handler atomically ORs
// signals into this value.
var receivedSignals atomic.Uint32

//go:linkname signal_enable os/signal.signal_enable
func signal_enable(s uint32) {
	if s >= 32 {
		// TODO: to support higher signal numbers, we need to turn
		// receivedSignals into a uint32 array.
		runtimePanicAt(returnAddress(0), "unsupported signal number")
	}

	// This is intentonally a non-atomic store. This is safe, since hasSignals
	// is only used in waitForEvents which is only called when there's a
	// scheduler (and therefore there is no parallelism).
	hasSignals = true

	// Under the threads scheduler there is no scheduler idle loop to notice
	// signals: checkSignals() is only reached from sleepTicks(), i.e. while
	// some goroutine happens to be inside time.Sleep. A program blocked purely
	// on I/O or channels would otherwise never observe a signal (Ctrl+C would
	// be ignored). Start a dedicated watcher thread to cover that case. This is
	// a no-op for every other scheduler.
	startSignalWatcher(s)

	// It's easier to implement this function in C.
	tinygo_signal_enable(s)
}

// signalWatcherStarted guards the start of signalWatcher. signal_enable is
// serialized by os/signal's handlers lock, but this stays defensive.
var signalWatcherStarted atomic.Uint32

// signalWatcherStop asks signalWatcher to return. The watcher reads it after
// waking, so stopping it means setting this and then waking the futex.
var signalWatcherStop atomic.Uint32

// enabledSignals is the set of signals os/signal currently wants delivered. The
// watcher thread exists only to serve them, so it runs exactly while this is
// non-zero: the last disable/ignore stops it, and a later enable starts a fresh
// one.
var enabledSignals atomic.Uint32

// startSignalWatcher starts the signal watcher thread when the first signal is
// enabled, but only under the threads scheduler (!hasScheduler && hasParallelism
// is true only there). The cooperative and multicore schedulers process signals
// from their idle loop (waitForEvents), and the "none" scheduler has no
// goroutines, so none of them need this.
func startSignalWatcher(s uint32) {
	if hasScheduler || !hasParallelism {
		return
	}
	enabledSignals.Or(uint32(1) << s)
	signalWatcherStop.Store(0)
	if signalWatcherStarted.Swap(1) == 0 {
		go signalWatcher()
	}
}

// stopSignalWatcher lets the watcher thread return once the signal it was
// serving is the last one to go away.
//
// Without this the thread is unstoppable by construction: it blocks on a futex
// forever, so a program that finishes with signals keeps a thread parked on one
// for the rest of its life. Nothing observable breaks — the thread is idle and
// the process exit tears it down — but "no exit condition" is a property worth
// not having, and it costs a flag and a wake to avoid.
func stopSignalWatcher(s uint32) {
	if hasScheduler || !hasParallelism {
		return
	}
	// And returns the value BEFORE the mask was applied, so clear the bit from
	// it to get what is left enabled.
	bit := uint32(1) << s
	if enabledSignals.And(^bit)&^bit != 0 {
		return // still serving other signals
	}
	if signalWatcherStarted.Swap(0) == 0 {
		return // not running
	}
	signalWatcherStop.Store(1)
	// Wake it so it can observe the flag. The value bump matters as much as the
	// wake: Wait(0) returns immediately if the futex is already non-zero, which
	// closes the window between the store above and a watcher about to sleep.
	//
	// WakeAll rather than Wake because the watcher is not the only thing that
	// sleeps on this futex: sleepTicks waits on it too, and so does
	// waitForEvents. Waking one waiter could wake a sleeping goroutine instead
	// — which would consume the value with its own Swap and leave the watcher
	// asleep on a futex that is 0 again, never seeing the flag it was told to
	// look at. The signal handler wakes this futex the same way.
	signalFutex.Store(1)
	signalFutex.WakeAll()
}

// signalWatcher runs on its own thread under the threads scheduler. It blocks on
// signalFutex and resumes the signal-receiving goroutine (signal_recv) whenever
// a signal arrives, decoupling signal delivery from sleepTicks(). It mirrors the
// signal half of waitForEvents(), which the threads scheduler never calls.
//
// It returns when stopSignalWatcher says the last enabled signal has gone away.
func signalWatcher() {
	for {
		// Block until the signal handler bumps the futex from 0 to 1.
		signalFutex.Wait(0)
		if signalWatcherStop.Load() != 0 {
			// Leave the futex as we found it for whoever runs next: a later
			// signal_enable starts a new watcher, and it must be able to sleep.
			signalFutex.Store(0)
			return
		}
		if signalFutex.Swap(0) != 0 {
			checkSignals()
		}
	}
}

//go:linkname signal_ignore os/signal.signal_ignore
func signal_ignore(s uint32) {
	if s >= 32 {
		// TODO: to support higher signal numbers, we need to turn
		// receivedSignals into a uint32 array.
		runtimePanicAt(returnAddress(0), "unsupported signal number")
	}
	stopSignalWatcher(s)
	tinygo_signal_ignore(s)
}

//go:linkname signal_disable os/signal.signal_disable
func signal_disable(s uint32) {
	if s >= 32 {
		// TODO: to support higher signal numbers, we need to turn
		// receivedSignals into a uint32 array.
		runtimePanicAt(returnAddress(0), "unsupported signal number")
	}
	stopSignalWatcher(s)
	tinygo_signal_disable(s)
}

//go:linkname signal_waitUntilIdle os/signal.signalWaitUntilIdle
func signal_waitUntilIdle() {
	// Wait until signal_recv has processed all signals.
	for receivedSignals.Load() != 0 {
		// TODO: this becomes a busy loop when using threads.
		// We might want to pause until signal_recv has no more incoming signals
		// to process.
		Gosched()
	}
}

//export tinygo_signal_enable
func tinygo_signal_enable(s uint32)

//export tinygo_signal_ignore
func tinygo_signal_ignore(s uint32)

//export tinygo_signal_disable
func tinygo_signal_disable(s uint32)

// void tinygo_signal_handler(int sig);
//
//export tinygo_signal_handler
func tinygo_signal_handler(s int32) {
	receivedSignals.Or(uint32(1) << uint32(s))

	// Notify the main thread that there was a signal.
	// This will exit the call to Wait or WaitUntil early.
	if signalFutex.Swap(1) == 0 {
		// Changed from 0 to 1, so there may have been a waiting goroutine.
		// This could be optimized to avoid a syscall when there are no waiting
		// goroutines.
		signalFutex.WakeAll()
	}
}

// Task waiting for a signal to arrive, or nil if it is running or there are no
// signals.
var signalRecvWaiter atomic.Pointer[task.Task]

//go:linkname signal_recv os/signal.signal_recv
func signal_recv() uint32 {
	// Function called from os/signal to get the next received signal.
	for {
		val := receivedSignals.Load()
		if val == 0 {
			// There are no signals to receive. Sleep until there are.
			if signalRecvWaiter.Swap(task.Current()) != nil {
				// We expect only a single goroutine to call signal_recv.
				runtimeFatal("signal_recv called concurrently")
			}
			task.Pause()
			continue
		}

		// Extract the lowest numbered signal number from receivedSignals.
		num := uint32(bits.TrailingZeros32(val))

		// Atomically clear the signal number from receivedSignals.
		receivedSignals.And(^(uint32(1) << num))

		return num
	}
}

// Reactivate the goroutine waiting for signals, if there are any.
// Return true if it was reactivated (and therefore the scheduler should run
// again), and false otherwise.
func checkSignals() bool {
	if receivedSignals.Load() != 0 {
		if waiter := signalRecvWaiter.Swap(nil); waiter != nil {
			scheduleTask(waiter)
			return true
		}
	}
	return false
}

func waitForEvents() {
	if hasSignals {
		// Wait as long as the futex value is 0.
		// This can happen either before or during the call to Wait.
		// This can be optimized: if the value is nonzero we don't need to do a
		// futex wait syscall and can instead immediately call checkSignals.
		signalFutex.Wait(0)

		// Check for signals that arrived before or during the call to Wait.
		// If there are any signals, the value is 0.
		if signalFutex.Swap(0) != 0 {
			checkSignals()
		}
	} else {
		// The program doesn't use signals, so this is a deadlock.
		runtimeFatal("deadlocked: no event source")
	}
}
