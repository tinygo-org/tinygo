//go:build (gc.conservative || gc.custom || gc.precise || gc.boehm) && tinygo.wasm

package runtime

import (
	"internal/task"
	"runtime/volatile"
	"unsafe"
)

func gcMarkReachable() {
	markStack()
	findGlobals(markRoots)
}

//go:extern runtime.stackChainStart
var stackChainStart *stackChainObject

type stackChainObject struct {
	parent   *stackChainObject
	numSlots uintptr
}

// markStack marks all root pointers found on the stack.
//
//   - Goroutine stacks are heap allocated and always reachable in some way
//     (for example through internal/task.currentTask) so they will always be
//     scanned.
//   - The system stack (aka startup stack) is not heap allocated, so even
//     though it may be referenced it will not be scanned by default.
//
// The compiler also inserts code to store all globals in a chain via
// stackChainStart. Luckily we don't need to scan these, as these globals are
// stored on the goroutine stack and are therefore already getting scanned.
func markStack() {
	// Hack to force LLVM to consider stackChainStart to be live.
	// Without this hack, loads and stores may be considered dead and objects on
	// the stack might not be correctly tracked. With this volatile load, LLVM
	// is forced to consider stackChainStart (and everything it points to) as
	// live.
	volatile.LoadUint32((*uint32)(unsafe.Pointer(&stackChainStart)))

	// Scan the system stack.
	var sysSP uintptr
	if task.OnSystemStack() {
		// We are on the system stack.
		// Use the current stack pointer.
		sysSP = getCurrentStackPointer()
	} else {
		// We are in a goroutine.
		// Use the saved stack pointer.
		sysSP = savedStackPointer
	}
	markRoots(sysSP, stackTop)
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

func gcResumeWorld() {
	// Nothing to do here (single threaded).
}
