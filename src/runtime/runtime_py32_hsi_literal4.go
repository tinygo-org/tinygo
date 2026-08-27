//go:build py32 && py32_hsi_fs_literal4

package runtime

import "device/py32"

// The affected vendor headers do not provide usable value constants.
func configureHSI() {
	py32.RCC.SetICSCR_HSI_FS(4)
}
