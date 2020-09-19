// +build wasm,wasi

package runtime

import (
	"unsafe"

	"github.com/tinygo-org/tinygo/src/syscall/wasi"
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
	sleepTicksSubscription = wasi.Subscription_t{
		UserData: 0,
		U: wasi.Subscription_u_t{
			Tag: wasi.Eventtype_t_clock,
			U: wasi.Subscription_clock_t{
				UserData:  0,
				ID:        0,
				Timeout:   0,
				Precision: timePrecisionNanoseconds,
				Flags:     0,
			},
		},
	}
	sleepTicksResult  = wasi.Event_t{}
	sleepTicksNEvents uint32
)

func sleepTicks(d timeUnit) {
	sleepTicksSubscription.U.U.Timeout = int64(d)
	wasi.Poll_oneoff(&sleepTicksSubscription, &sleepTicksResult, 1, &sleepTicksNEvents)
}

func ticks() timeUnit {
	var nano int64
	wasi.Clock_time_get(0, timePrecisionNanoseconds, &nano)
	return timeUnit(nano)
}
