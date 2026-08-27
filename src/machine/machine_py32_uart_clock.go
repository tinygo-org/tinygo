//go:build py32 && !py32_uart_type && !py32_no_hsi_fs

package machine

import "device/py32"

func enableUSART1Clock() {
	py32.RCC.APBENR2.SetBits(1 << 14)
}
