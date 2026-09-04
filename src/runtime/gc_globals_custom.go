//go:build gc.custom

package runtime

func markGlobals() {
	findGlobals(markRoots)
}
