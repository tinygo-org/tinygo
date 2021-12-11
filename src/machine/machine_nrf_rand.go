//go:build nrf && !softdevice
// +build nrf,!softdevice

package machine

// GetRNG returns 32 bits of cryptographically secure random data
func GetRNG() (uint32, error) {
	return getRNG()
}
