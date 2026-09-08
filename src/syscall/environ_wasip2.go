//go:build wasip2

package syscall

import "unsafe"

func Environ() []string {
	// __wasilibc_get_environ (rather than referencing the `environ` symbol
	// directly, as environ_libc.go does for wasip1) triggers lazy
	// initialization of the environment instead of eager,
	// constructor-based initialization (see wasi-libc's environ.c vs.
	// __wasilibc_environ.c). Eager initialization runs from a global
	// constructor, which under the wasip2 component model can end up
	// running while cabi_realloc is being invoked reentrantly to service an
	// unrelated, still in-flight host->guest call -- wasmtime rejects that
	// ("cannot leave component instance"). Lazy initialization only runs
	// when Environ() is actually called by user code, well after module
	// instantiation has completed, so it doesn't hit that restriction.
	return environFromPointer(libc_get_environ())
}

//export __wasilibc_get_environ
func libc_get_environ() *unsafe.Pointer
