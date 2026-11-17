package llvmutil

/*
extern int LLVMGetThreadLocalMode(void *global);
extern void LLVMSetThreadLocalMode(void *global, int mode);
*/
import "C"

import (
	"unsafe"

	"tinygo.org/x/go-llvm"
)

func readThreadLocalMode(global llvm.Value) int {
	return int(C.LLVMGetThreadLocalMode(unsafe.Pointer(global.C)))
}

func setThreadLocalMode(global llvm.Value, mode int) {
	C.LLVMSetThreadLocalMode(unsafe.Pointer(global.C), C.int(mode))
}
