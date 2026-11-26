//go:build (gc.conservative || gc.custom || gc.precise || gc.boehm) && tinygo.wasm

package runtime

import (
	"runtime/volatile"
	"unsafe"
)

//go:extern runtime.stackChainStart
var stackChainStart *stackChainObject

type stackChainObject struct {
	parent   *stackChainObject
	numSlots uintptr
}

// trackPointer is a stub function call inserted by the compiler during IR
// construction. Calls to it are later replaced with regular stack bookkeeping
// code.
func trackPointer(ptr, alloca unsafe.Pointer)

// swapStackChain swaps the stack chain.
// This is called from internal/task when switching goroutines.
func swapStackChain(dst **stackChainObject) {
	*dst, stackChainStart = stackChainStart, *dst
}

// keepStackChain forces LLVM to consider stackChainStart to be live. Without
// this, loads and stores may be considered dead and objects on the stack might
// not be tracked.
func keepStackChain() {
	volatile.LoadUint32((*uint32)(unsafe.Pointer(&stackChainStart)))
}
