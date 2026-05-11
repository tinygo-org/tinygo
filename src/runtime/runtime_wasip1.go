//go:build wasip1

package runtime

import (
	"unsafe"
)

// libc constructors
//
//export __wasm_call_ctors
func __wasm_call_ctors()

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
		args_get(&argsSlice[0], unsafe.Pointer(unsafe.SliceData(buf)))

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
	__wasi_eventtype_t_clock __wasi_eventtype_t = iota
	__wasi_eventtype_t_fd_read
	__wasi_eventtype_t_fd_write
)

type (
	// https://github.com/WebAssembly/WASI/blob/main/phases/snapshot/docs.md#-subscription-record
	__wasi_subscription_t struct {
		userData uint64
		u        __wasi_subscription_u_t
	}

	// The union payload is sized by the largest variant (clock, 32 bytes after
	// the tag and its 7-byte alignment pad). FD read/write subscriptions reuse
	// the same memory via setFDReadWrite.
	__wasi_subscription_u_t struct {
		tag __wasi_eventtype_t

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

// __wasi_subscription_fd_readwrite_t is the FD variant of the subscription
// union payload. It overlays the first 4 bytes of the clock variant.
type __wasi_subscription_fd_readwrite_t struct {
	fd uint32
}

func (s *__wasi_subscription_u_t) setClock(id uint32, timeoutNs, precision uint64, flags uint16) {
	s.tag = __wasi_eventtype_t_clock
	s.u = __wasi_subscription_clock_t{
		id:        id,
		timeout:   timeoutNs,
		precision: precision,
		flags:     flags,
	}
}

func (s *__wasi_subscription_u_t) setFDReadWrite(eventType __wasi_eventtype_t, fd uint32) {
	s.tag = eventType
	s.u = __wasi_subscription_clock_t{}
	(*__wasi_subscription_fd_readwrite_t)(unsafe.Pointer(&s.u)).fd = fd
}

type (
	// https://github.com/WebAssembly/WASI/blob/main/phases/snapshot/docs.md#-event-record
	__wasi_event_t struct {
		userData  uint64
		errno     uint16
		eventType __wasi_eventtype_t

		// fdReadWrite is populated by poll_oneoff for fd_read / fd_write events.
		// For clock events the field is zero. Reading nBytes/flags after a
		// clock event is meaningless but not unsafe.
		fdReadWrite struct {
			nBytes uint64
			flags  uint16
		}
	}
)

// Compile-time size assertions for the wasip1 ABI. If these fail to compile
// the struct layout drifted from the spec and poll_oneoff would corrupt
// memory.
var _ [0]byte = [48 - unsafe.Sizeof(__wasi_subscription_t{})]byte{}
var _ [0]byte = [32 - unsafe.Sizeof(__wasi_event_t{})]byte{}
