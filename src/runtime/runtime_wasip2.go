//go:build wasip2

package runtime

import (
	"unsafe"

	"internal/wasi/cli/v0.2.12/environment"
	wasiclirun "internal/wasi/cli/v0.2.12/run"
	monotonicclock "internal/wasi/clocks/v0.2.12/monotonic-clock"

	"internal/cm"
)

func init() {
	if gorootEnvEnabled() {
		if value := startupEnv("GOROOT"); value != "" {
			goroot = value
		}
	}
	if godebugEnvEnabled() {
		godebugStartup = startupEnv("GODEBUG")
	}

	wasiclirun.Exports.Run = func() cm.BoolResult {
		callMain()
		return false
	}
}

func startupEnv(key string) string {
	for _, entry := range environment.GetEnvironment().Slice() {
		if entry[0] == key {
			return entry[1]
		}
	}
	return ""
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
func cabi_realloc(ptr, oldsize, align, newsize unsafe.Pointer) unsafe.Pointer {
	// The C code in wasi-libc frees these buffers with free().
	return libc_realloc(ptr, uintptr(newsize))
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
