//go:build wasip3

package runtime

import (
	"unsafe"
)

// Types from wasi-libc's wasi/__generated_wasip3.h.
type wasip3String struct {
	ptr *byte
	len uintptr
}

type wasip3ListString struct {
	ptr *wasip3String
	len uintptr
}

// void environment_get_arguments(wasip3_list_string_t *ret);
//
//export environment_get_arguments
func environment_get_arguments(ret *wasip3ListString)

// void wasip3_list_string_free(wasip3_list_string_t *ptr);
//
//export wasip3_list_string_free
func wasip3_list_string_free(ptr *wasip3ListString)

// Entry point for the wasi:cli/run export. The result is true for an error.
//
//go:wasmexport wasi:cli/run@0.3.0#run
func wasiCliRun() bool {
	initAll()
	callMain()
	return false
}

var args []string

//go:linkname os_runtime_args os.runtime_args
func os_runtime_args() []string {
	if args == nil {
		var list wasip3ListString
		environment_get_arguments(&list)
		args = make([]string, list.len)
		for i := range args {
			s := (*wasip3String)(unsafe.Add(unsafe.Pointer(list.ptr), uintptr(i)*unsafe.Sizeof(wasip3String{})))
			args[i] = string(unsafe.Slice(s.ptr, s.len))
		}
		wasip3_list_string_free(&list)
	}
	return args
}

//export cabi_realloc
func cabi_realloc(ptr, oldsize, align, newsize unsafe.Pointer) unsafe.Pointer {
	// Use libc_realloc (not the GC-internal realloc) so this allocation is
	// tracked in the same allocs map as malloc/free: wasi-libc's own C code
	// takes ownership of buffers allocated here during canonical-ABI lifting
	// and later frees them with a plain free().
	return libc_realloc(ptr, uintptr(newsize))
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks)
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns)
}

type libcTimespec struct {
	sec  int64
	nsec int32
}

// int nanosleep(const struct timespec *req, struct timespec *rem);
//
//export nanosleep
func libc_nanosleep(req, rem *libcTimespec) int32

func sleepTicks(d timeUnit) {
	req := libcTimespec{
		sec:  int64(d) / 1e9,
		nsec: int32(int64(d) % 1e9),
	}
	libc_nanosleep(&req, nil)
}

// monotonic_clock_mark_t monotonic_clock_now(void);
//
//export monotonic_clock_now
func monotonic_clock_now() uint64

func ticks() timeUnit {
	return timeUnit(monotonic_clock_now())
}
