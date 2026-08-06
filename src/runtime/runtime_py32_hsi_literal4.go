//go:build py32 && py32_hsi_fs_literal4

package runtime

import "device/py32"

// F001 lacks its referenced device header. F002C's header shifts its HSI_FS
// bits by the unrelated LSI_TRIM position. In both cases Puya's official system
// sources map encoding 4 to the runtime's 24 MHz clock.
func configureHSI() {
	py32.RCC.SetICSCR_HSI_FS(4)
}
