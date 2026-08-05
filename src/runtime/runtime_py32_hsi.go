//go:build py32 && !py32_hsi_fs_op && !py32_no_hsi_fs

package runtime

import "device/py32"

// Puya's raw SVDs define HSI_FS but omit its enumerated values. Value 4 selects
// the 24 MHz clock used by the common PY32 runtime.
func configureHSI() {
	py32.RCC.SetICSCR_HSI_FS(4)
}
