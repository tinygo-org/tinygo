//go:build windows && !go1.26

package runtime

// For Go < 1.26, SyscallN is declared without a body in
// syscall/dll_windows.go. The standard Go runtime provides its implementation
// via //go:linkname. TinyGo must provide one as well.
//
// Note: The TinyGo compiler cannot correctly handle SyscallN as a builtin
// because it uses variadic arguments represented as a slice in SSA, which
// the fixed-argument builtin mechanism cannot process. This implementation
// is provided to satisfy the linker for code paths that reference SyscallN
// (e.g., syscall.Proc.Call).

//go:linkname syscall_SyscallN syscall.SyscallN
func syscall_SyscallN(trap uintptr, args ...uintptr) (r1, r2, err uintptr) {
	panic("syscall.SyscallN is not yet supported in TinyGo for Go < 1.26")
}
