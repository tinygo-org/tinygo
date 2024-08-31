//go:build wasm_unknown

package runtime

// TODO: this is essentially reactor mode wasm. So we might want to support
// -buildmode=c-shared (and default to it).

type timeUnit int64

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

// with the wasm32-unknown-unknown target there is no way to determine any `precision`
const timePrecisionNanoseconds = 1000

func sleepTicks(d timeUnit) {
}

func ticks() timeUnit {
	return timeUnit(0)
}

func beforeExit() {
}
