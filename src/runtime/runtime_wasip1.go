//go:build wasip1

package runtime

import (
	"unsafe"
)

type timeUnit int64

// libc constructors
//
//export __wasm_call_ctors
func __wasm_call_ctors()

//export _start
func _start() {
	// These need to be initialized early so that the heap can be initialized.
	heapStart = uintptr(unsafe.Pointer(&heapStartSymbol))
	heapEnd = uintptr(wasm_memory_size(0) * wasmPageSize)
	run()
	__stdio_exit()
}

// Read the command line arguments from WASI.
// For example, they can be passed to a program with wasmtime like this:
//
//	wasmtime run ./program.wasm arg1 arg2
func init() {
	__wasm_call_ctors()
}

var args []string

//go:linkname os_runtime_args os.runtime_args
func os_runtime_args() []string {
	if args == nil {
		// Read the number of args (argc) and the buffer size required to store
		// all these args (argv).
		var argc, argv_buf_size uint32
		args_sizes_get(&argc, &argv_buf_size)
		if argc == 0 {
			return nil
		}

		// Obtain the command line arguments
		argsSlice := make([]unsafe.Pointer, argc)
		buf := make([]byte, argv_buf_size)
		args_get(&argsSlice[0], unsafe.Pointer(&buf[0]))

		// Convert the array of C strings to an array of Go strings.
		args = make([]string, argc)
		for i, cstr := range argsSlice {
			length := strlen(cstr)
			argString := _string{
				length: length,
				ptr:    (*byte)(cstr),
			}
			args[i] = *(*string)(unsafe.Pointer(&argString))
		}
	}
	return args
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks)
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns)
}

const timePrecisionNanoseconds = 1000 // TODO: how can we determine the appropriate `precision`?

var (
	sleepTicksSubscription = __wasi_subscription_t{
		userData: 0,
		u: __wasi_subscription_u_t{
			tag: __wasi_eventtype_t_clock,
			u: __wasi_subscription_clock_t{
				id:        0,
				timeout:   0,
				precision: timePrecisionNanoseconds,
				flags:     0,
			},
		},
	}
	sleepTicksResult  = __wasi_event_t{}
	sleepTicksNEvents uint32
)

func sleepTicks(d timeUnit) {
	sleepTicksSubscription.u.u.timeout = uint64(d)
	poll_oneoff(&sleepTicksSubscription, &sleepTicksResult, 1, &sleepTicksNEvents)
}

func ticks() timeUnit {
	var nano uint64
	clock_time_get(0, timePrecisionNanoseconds, &nano)
	return timeUnit(nano)
}

// Implementations of WASI APIs

//go:wasmimport wasi_snapshot_preview1 args_get
func args_get(argv *unsafe.Pointer, argv_buf unsafe.Pointer) (errno uint16)

//go:wasmimport wasi_snapshot_preview1 args_sizes_get
func args_sizes_get(argc *uint32, argv_buf_size *uint32) (errno uint16)

//go:wasmimport wasi_snapshot_preview1 clock_time_get
func clock_time_get(clockid uint32, precision uint64, time *uint64) (errno uint16)

//go:wasmimport wasi_snapshot_preview1 poll_oneoff
func poll_oneoff(in *__wasi_subscription_t, out *__wasi_event_t, nsubscriptions uint32, nevents *uint32) (errno uint16)

type __wasi_eventtype_t = uint8

const (
	__wasi_eventtype_t_clock __wasi_eventtype_t = 0
	// TODO: __wasi_eventtype_t_fd_read  __wasi_eventtype_t = 1
	// TODO: __wasi_eventtype_t_fd_write __wasi_eventtype_t = 2
)

type (
	// https://github.com/WebAssembly/WASI/blob/main/phases/snapshot/docs.md#-subscription-record
	__wasi_subscription_t struct {
		userData uint64
		u        __wasi_subscription_u_t
	}

	__wasi_subscription_u_t struct {
		tag __wasi_eventtype_t

		// TODO: support fd_read/fd_write event
		u __wasi_subscription_clock_t
	}

	// https://github.com/WebAssembly/WASI/blob/main/phases/snapshot/docs.md#-subscription_clock-record
	__wasi_subscription_clock_t struct {
		id        uint32
		timeout   uint64
		precision uint64
		flags     uint16
	}
)

type (
	// https://github.com/WebAssembly/WASI/blob/main/phases/snapshot/docs.md#-event-record
	__wasi_event_t struct {
		userData  uint64
		errno     uint16
		eventType __wasi_eventtype_t

		// only used for fd_read or fd_write events
		// TODO: support fd_read/fd_write event
		_ struct {
			nBytes uint64
			flags  uint16
		}
	}
)
