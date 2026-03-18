//go:build darwin && !go1.26

package runtime

// Call "system calls" (actually: libc functions) in a special way.
//   - Most calls calls return a C int (which is 32-bits), and -1 on failure.
//   - syscallX* is for functions that return a 64-bit integer (and also return
//     -1 on failure).
//   - syscallPtr is for functions that return a pointer on success or NULL on
//     failure.
//   - rawSyscall seems to avoid some stack modifications, which isn't relevant
//     to TinyGo.

//go:linkname syscall_syscall syscall.syscall
func syscall_syscall(fn, a1, a2, a3 uintptr) (r1, r2, err uintptr) {
	// For TinyGo we don't need to do anything special to call C functions.
	return syscall_rawSyscall(fn, a1, a2, a3)
}

//go:linkname syscall_rawSyscall syscall.rawSyscall
func syscall_rawSyscall(fn, a1, a2, a3 uintptr) (r1, r2, err uintptr) {
	result := call_syscall(fn, a1, a2, a3)
	r1 = uintptr(result)
	if result == -1 {
		// Syscall returns -1 on failure.
		err = uintptr(*libc_errno_location())
	}
	return
}

//go:linkname syscall_syscallX syscall.syscallX
func syscall_syscallX(fn, a1, a2, a3 uintptr) (r1, r2, err uintptr) {
	r1 = call_syscallX(fn, a1, a2, a3)
	if int64(r1) == -1 {
		// Syscall returns -1 on failure.
		err = uintptr(*libc_errno_location())
	}
	return
}

//go:linkname syscall_syscallPtr syscall.syscallPtr
func syscall_syscallPtr(fn, a1, a2, a3 uintptr) (r1, r2, err uintptr) {
	r1 = call_syscallX(fn, a1, a2, a3)
	if r1 == 0 {
		// Syscall returns a pointer on success, or NULL on failure.
		err = uintptr(*libc_errno_location())
	}
	return
}

//go:linkname syscall_syscall6 syscall.syscall6
func syscall_syscall6(fn, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2, err uintptr) {
	result := call_syscall6(fn, a1, a2, a3, a4, a5, a6)
	r1 = uintptr(result)
	if result == -1 {
		// Syscall returns -1 on failure.
		err = uintptr(*libc_errno_location())
	}
	return
}

//go:linkname syscall_syscall6X syscall.syscall6X
func syscall_syscall6X(fn, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2, err uintptr) {
	r1 = call_syscall6X(fn, a1, a2, a3, a4, a5, a6)
	if int64(r1) == -1 {
		// Syscall returns -1 on failure.
		err = uintptr(*libc_errno_location())
	}
	return
}
