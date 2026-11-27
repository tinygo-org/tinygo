//go:build darwin || (linux && !baremetal && !nintendoswitch && !wasm_unknown && !wasip2)

package runtime

var (
	speckLock   randLock
	speckState  speckRand
	speckSeeded bool
)

// hardwareRand returns values from a Speck CTR generator keyed once from
// osRand, so most calls avoid a syscall. See issue #6531.
func hardwareRand() (n uint64, ok bool) {
	speckLock.Lock()
	if !speckSeeded {
		k0, ok0 := osRand()
		k1, ok1 := osRand()
		if ok0 && ok1 {
			speckState.init(k0, k1)
			speckSeeded = true
		}
	}
	if speckSeeded {
		n, ok = speckState.next(), true
	}
	speckLock.Unlock()
	return
}
