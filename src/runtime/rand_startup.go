//go:build !linux || baremetal || nintendoswitch || wasip1 || wasm_unknown || wasip2

package runtime

func startupRand() (uint64, bool) {
	return hardwareRand()
}
