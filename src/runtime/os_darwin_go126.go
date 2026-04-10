//go:build darwin && go1.26

package runtime

// Go 1.26 replaced the individual Darwin syscall functions (syscall, syscallX,
// syscallPtr, syscall6, syscall6X) with two variadic entry points: syscalln and
// rawsyscalln. All the old wrappers now have Go bodies that call these and then
// interpret the result (errno, errnoX, errnoPtr). Because the callers decide
// how to check for errors, we must:
//   - always return the full register-width result (use call_syscallX /
//     call_syscall6X which return uintptr, not call_syscall / call_syscall6
//     which truncate to int32), and
//   - always read errno so that the Go wrappers have the value available when
//     the error condition is met.

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
		r1 = call_syscallX(fn, a1, a2, a3)
		err = uintptr(*libc_errno_location())
		return

	case 6:
		a6 = args[5]
		fallthrough
	case 5:
		a5 = args[4]
		fallthrough
	case 4:
		a4 = args[3]

		a1, a2, a3 = args[0], args[1], args[2]
		r1 = call_syscall6X(fn, a1, a2, a3, a4, a5, a6)
		err = uintptr(*libc_errno_location())
		return
	}

	panic("syscall args not handled")
}
