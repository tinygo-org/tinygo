//go:build darwin && go1.26

package runtime

// syscall_syscalln is a wrapper around the libc call with variable arguments.
//
//go:nosplit
//go:linkname syscall_syscalln syscall.syscalln
func syscall_syscalln(fn uintptr, args ...uintptr) (r1, r2, err uintptr) {
	r1, r2, err = syscall_rawsyscalln(fn, args...)
	return r1, r2, err
}

// syscall_rawsyscalln is a wrapper around the libc call with variable arguments.
//
//go:linkname syscall_rawsyscalln syscall.rawsyscalln
//go:nosplit
func syscall_rawsyscalln(fn uintptr, args ...uintptr) (r1, r2, err uintptr) {

	var a1, a2, a3, a4, a5, a6 uintptr

	switch len(args) {
	case 3:
		a3 = args[2]
		fallthrough
	case 2:
		a2 = args[1]
		fallthrough
	case 1:
		a1 = args[0]
		fallthrough
	case 0:
		return runtime_syscall(fn, a1, a2, a3)

	case 6:
		a6 = args[5]
		fallthrough
	case 5:
		a5 = args[4]
		fallthrough
	case 4:
		a3 = args[3]

		a1, a2, a3 = args[0], args[1], args[2]
		return runtime_syscall6(fn, a1, a2, a3, a4, a5, a6)
	}

	panic("syscall args not handled")
}

func runtime_syscall(fn, a1, a2, a3 uintptr) (r1, r2, err uintptr) {
	result := call_syscall(fn, a1, a2, a3)
	r1 = uintptr(result)
	if result == -1 {
		// Syscall returns -1 on failure.
		err = uintptr(*libc_errno_location())
	}
	return
}

func runtime_syscall6(fn, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2, err uintptr) {
	result := call_syscall6(fn, a1, a2, a3, a4, a5, a6)
	r1 = uintptr(result)
	if result == -1 {
		// Syscall returns -1 on failure.
		err = uintptr(*libc_errno_location())
	}
	return
}
