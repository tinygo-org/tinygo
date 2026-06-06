//go:build uefi

package uefi

// callAsm is the single assembly stub for all UEFI calls.
// It takes a function pointer, a pointer to an argument array, and the count.
//
//export uefiCall
func callAsm(fn uintptr, args *uintptr, nargs uintptr) Status

// Call invokes a UEFI function with the given arguments via the MS x64 ABI.
func Call(fn uintptr, args ...uintptr) Status {
	if len(args) == 0 {
		return callAsm(fn, nil, 0)
	}
	return callAsm(fn, &args[0], uintptr(len(args)))
}
