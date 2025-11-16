//go:build nrf && !softdevice

package machine

// GetRNG returns 32 bits of non-deterministic random data based on internal thermal noise.
// According to Nordic's documentation, the random output is suitable for cryptographic purposes.
func GetRNG() (ret uint32, err error) {
	return getRNG()
}

func isSoftDeviceEnabled() bool {
	return false
}
