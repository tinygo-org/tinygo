//go:build gc.custom && tinygo.wasm

package runtime

func markGlobals() {
	findGlobals(markRoots)
}
