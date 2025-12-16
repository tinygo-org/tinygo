//go:build rp2040 || rp2350

package runtime

import "machine"

var hardwareRandValue uint64

func hardwareRand() (n uint64, ok bool) {
	if hardwareRandValue == 0 {
		n1, _ := machine.GetRNG()
		n2, _ := machine.GetRNG()
		hardwareRandValue = uint64(n1)<<32 | uint64(n2)
	}

	// Return ok=false to keep using fastrand64(),
	// with hardwareRandVal used only as its initial random state.

	return hardwareRandValue, false
}
