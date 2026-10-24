//go:build nrf && !softdevice

package machine

// RandomUint32 returns 32 bits of non-deterministic random data based on internal thermal noise.
// According to Nordic's documentation, the random output is suitable for cryptographic purposes.
func RandomUint32() (ret uint32, err error) {
	return getRNG()
}

// GetRNG returns a random uint32. See RandomUint32 for details.
//
// Deprecated: use RandomUint32 instead.
//
//go:fix inline
func GetRNG() (uint32, error) {
	return RandomUint32()
}

func isSoftDeviceEnabled() bool {
	return false
}
