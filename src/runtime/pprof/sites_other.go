//go:build !(tinygo.wasm && tinygo.shadowstack)

package pprof

// The stacks hold return addresses, so the call is the instruction before.
const returnAddressAdjust = 1

func siteFrame(site uintptr, i int) (name, file string, line int, ok bool) {
	return "", "", 0, false
}
