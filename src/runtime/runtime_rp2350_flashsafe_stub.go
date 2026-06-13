//go:build rp2350

package runtime

func rp2FlashSafeInterruptHandler(uint32) {
	// No-op on RP2350. RP2350 flash-safe handling is intentionally unchanged.
}
