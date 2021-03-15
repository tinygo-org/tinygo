// +build wasm,wasi

package runtime

import (
	"unsafe"
)

type timeUnit int64

//export _start
func _start() {
	// These need to be initialized early so that the heap can be initialized.
	heapStart = uintptr(unsafe.Pointer(&heapStartSymbol))
	heapEnd = uintptr(wasm_memory_size(0) * wasmPageSize)
	run()
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks)
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns)
}

const (
	asyncScheduler           = false
	timePrecisionNanoseconds = 1000 // TODO: how can we determine the appropriate `precision`?
)

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

// Implementations of wasi_unstable APIs

//go:wasm-module wasi_snapshot_preview1
//export clock_time_get
func clock_time_get(clockid uint32, precision uint64, time *uint64) (errno uint16)

//go:wasm-module wasi_snapshot_preview1
//export poll_oneoff
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
