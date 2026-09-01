//go:build stm32 && stm32h7 && !stm32h723 && !stm32h757_cm7

package runtime

import (
	_ "machine/usb/cdc"
)
