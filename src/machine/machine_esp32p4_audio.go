//go:build esp32p4

// Shared audio helpers for ESP32-P4 boards.

package machine

// SetAmplifier configures an amplifier enable pin and drives it on or off.
// activeHigh selects the polarity: true drives the pin high to enable the
// amplifier, false drives it low.  A NoPin is ignored.
func SetAmplifier(pin Pin, activeHigh, enable bool) {
	if pin == NoPin {
		return
	}
	pin.Configure(PinConfig{Mode: PinOutput})
	if activeHigh {
		pin.Set(enable)
	} else {
		pin.Set(!enable)
	}
}
