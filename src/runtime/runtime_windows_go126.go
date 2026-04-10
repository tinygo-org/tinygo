//go:build windows && go1.26

package runtime

// Starting with Go 1.26, the syscall package on Windows defines function bodies
// for Syscall, SyscallN, etc., that all call syscalln (lowercase). In standard
// Go, syscalln is provided by the runtime via //go:linkname. TinyGo's compiler
// intercepts calls to syscall.syscalln and replaces them with inline LLVM IR
// (see compiler/syscall.go createSyscalln), so this function body is never
// actually called at runtime. However, the compiled function bodies in the
// syscall package still reference it, so we must provide a definition to
// satisfy the linker.

//go:linkname syscall_syscalln syscall.syscalln
func syscall_syscalln(fn, n uintptr, args ...uintptr) (r1, r2, err uintptr) {
	panic("unreachable: syscall.syscalln should be handled by the compiler")
}
