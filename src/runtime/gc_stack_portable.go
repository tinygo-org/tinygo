//go:build (gc.conservative || gc.custom) && tinygo.wasm
// +build gc.conservative gc.custom
// +build tinygo.wasm

package runtime

import (
	"internal/task"
	"runtime/volatile"
	"unsafe"
)

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
// Therefore, we only need to scan the system stack.
// It is relatively easy to scan the system stack while we're on it: we can
// simply read __stack_pointer and __global_base and scan the area inbetween.
// Unfortunately, it's hard to get the system stack pointer while we're on a
// goroutine stack. But when we're on a goroutine stack, the system stack is in
// the scheduler which means there shouldn't be anything on the system stack
// anyway.
// ...I hope this assumption holds, otherwise we will need to store the system
// stack in a global or something.
//
// The compiler also inserts code to store all globals in a chain via
// stackChainStart. Luckily we don't need to scan these, as these globals are
// stored on the goroutine stack and are therefore already getting scanned.
func markStack() {
	loadStackChain()
	if task.OnSystemStack() {
		markRoots(getCurrentStackPointer(), stackTop)
	}
}

// trackPointer is a stub function call inserted by the compiler during IR
// construction. Calls to it are later replaced with regular stack bookkeeping
// code.
func trackPointer(ptr unsafe.Pointer)

// swapStackChain swaps the stack chain.
// This is called from internal/task when switching goroutines.
func swapStackChain(dst **stackChainObject) {
	*dst, stackChainStart = stackChainStart, *dst
}

func init() {
	// At least currently, it is enough to load the stack chain once in the app to make
	// sure stack tracking is emitted by LLVM, which will mostly be important when using
	// gc=custom. In the future, this may change if LLVM optimizes away the stack chain
	// because the load happens before the writes and we may need to revisit.
	loadStackChain()
}

func loadStackChain() {
	// Hack to force LLVM to consider stackChainStart to be live.
	// Without this hack, loads and stores may be considered dead and objects on
	// the stack might not be correctly tracked. With this volatile load, LLVM
	// is forced to consider stackChainStart (and everything it points to) as
	// live.
	volatile.LoadUint32((*uint32)(unsafe.Pointer(&stackChainStart)))
}
