package runtime

// This file implements syscall.Syscall and the like.

//go:linkname syscall_Syscall syscall.Syscall
func syscall_Syscall(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err uintptr) {
	panic("syscall")
}

//go:linkname syscall_Syscall6 syscall.Syscall6
func syscall_Syscall6(trap, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err uintptr) {
	panic("syscall6")
}
