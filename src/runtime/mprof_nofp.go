//go:build (gc.conservative || gc.precise) && !((darwin || linux) && !baremetal && (amd64 || arm64) && scheduler.threads) && !(tinygo.wasm && tinygo.shadowstack)

package runtime

// memProfileCallers cannot walk the stack on this target, so the memory
// profile has no stacks.
func memProfileCallers(stk []uintptr) int {
	return 0
}
