//go:build stm32f4 && (stm32f429 || stm32f427 || stm32f411 || stm32f407 || stm32f405 || stm32f401)

package machine

import "device/stm32"

// initOTGFSPHY enables the STM32F4 OTG FS PHY and bypasses VBUS sensing.
// GCCFG.PWRDWN deactivates the PHY power-down; NOVBUSSENS skips the VBUS pin
// check so boards without PA9 connected to VBUS still enumerate.
func initOTGFSPHY() {
	stm32.OTG_FS_GLOBAL.GCCFG.Set(
		stm32.USB_OTG_FS_GCCFG_PWRDWN | // enable FS PHY
			stm32.USB_OTG_FS_GCCFG_NOVBUSSENS, // bypass VBUS sensing
	)
}
