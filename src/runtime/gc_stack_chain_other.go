//go:build !(gc.conservative || gc.custom || gc.precise || gc.boehm) || !tinygo.wasm

package runtime

func keepStackChain() {}
