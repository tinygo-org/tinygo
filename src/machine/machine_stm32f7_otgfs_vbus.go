//go:build stm32f7

package machine

import (
	"device/arm"
	"device/stm32"
)

// initOTGFSPHY enables the STM32F7 OTG FS PHY and overrides B-session VBUS
// detection via GOTGCTL so boards without VBUS sensing still enumerate.
func initOTGFSPHY() {
	// Enable USB voltage regulator (specific to F72x/F73x).
	// Bit 14 in PWR_CR2 is USBREGEN.
	stm32.PWR.CR2.SetBits(1 << 14)

	// Stabilization delay for the regulator (~100us is plenty).
	for i := 0; i < 10000; i++ {
		arm.Asm("nop")
	}

	// Enable FS PHY.
	stm32.OTG_FS_GLOBAL.GCCFG.SetBits(stm32.USB_OTG_FS_GCCFG_PWRDWN)

	// Disable hardware VBUS detection (F7 uses VBDEN, opposite polarity to F4's NOVBUSSENS).
	// Clearing this prevents the peripheral from gating enumeration on PA9 VBUS level.
	stm32.OTG_FS_GLOBAL.GCCFG.ClearBits(stm32.USB_OTG_FS_GCCFG_VBDEN)

	// Override B-session valid so GOTGCTL-based detection reports device connected.
	stm32.OTG_FS_GLOBAL.GOTGCTL.SetBits(
		stm32.USB_OTG_FS_GOTGCTL_BVALOEN | // enable B-valid override
			stm32.USB_OTG_FS_GOTGCTL_BVALOVAL, // set B-valid = 1
	)
}
