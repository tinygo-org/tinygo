//go:build wasm_unknown

package runtime

// TODO: this is essentially reactor mode wasm. So we might want to support
// -buildmode=c-shared (and default to it).

// libc constructors
//
//export __wasm_call_ctors
func __wasm_call_ctors()

func init() {
	__wasm_call_ctors()
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks)
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns)
}

func sleepTicks(d timeUnit) {
}

func ticks() timeUnit {
	return timeUnit(0)
}

func mainReturnExit() {
	// Don't exit explicitly here. We can't (there is no environment with an
	// exit call) but also it's not needed. We can just let _start and main.main
	// return to the caller.
}
