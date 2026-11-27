//go:build (!linux && !darwin) || baremetal || nintendoswitch || tinygo.wasi || wasm_unknown

package runtime

func startupRand() (uint64, bool) {
	return hardwareRand()
}
