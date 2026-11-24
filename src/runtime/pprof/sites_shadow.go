//go:build tinygo.wasm && tinygo.shadowstack

package pprof

// The stacks hold call site IDs from the shadow stack, which are not
// addresses, so they are used as they are.
const returnAddressAdjust = 0

// siteFrame returns frame i of a call site, innermost first.
func siteFrame(site uintptr, i int) (name, file string, line int, ok bool)
