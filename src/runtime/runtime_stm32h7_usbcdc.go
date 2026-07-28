//go:build stm32 && stm32h7 && !stm32h723

package runtime

import (
	_ "machine/usb/cdc"
)
