//go:build py32 && py32_hsi_fs_literal3

package runtime

import "device/py32"

// The F032 header does not provide value constants for this field.
func configureHSI() {
	py32.RCC.SetICSCR_HSI_FS(3)
}
