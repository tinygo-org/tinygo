//go:build wasip2

package runtime

import (
	"unsafe"

	"internal/wasi/cli/v0.2.0/environment"
	wasiclirun "internal/wasi/cli/v0.2.0/run"
	monotonicclock "internal/wasi/clocks/v0.2.0/monotonic-clock"

	"internal/cm"
)

func init() {
	wasiclirun.Exports.Run = func() cm.BoolResult {
		callMain()
		return false
	}
}

var args []string

//go:linkname os_runtime_args os.runtime_args
func os_runtime_args() []string {
	if args == nil {
		args = environment.GetArguments().Slice()
	}
	return args
}

//export cabi_realloc
func cabi_realloc(ptr unsafe.Pointer, oldSize, align, newSize uintptr) unsafe.Pointer {
	if newSize == 0 {
		return nil
	}
	newPtr := realloc(ptr, newSize)
	if ptr != nil {
		for i := range wasmAllocs {
			if wasmAllocs[i] == ptr {
				wasmAllocs[i] = newPtr
				return newPtr
			}
		}
	}
	wasmAllocs = append(wasmAllocs, newPtr)
	return newPtr
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks)
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns)
}

func sleepTicks(d timeUnit) {
	p := monotonicclock.SubscribeDuration(monotonicclock.Duration(d))
	p.Block()
}

func ticks() timeUnit {
	return timeUnit(monotonicclock.Now())
}
