//go:build py32 && py32_hsi_fs_literal3

package runtime

import "device/py32"

// The F032 header names this field HSI_FS_CR while its SVD names it HSI_FS, so
// no matching header-derived value constants are generated. Puya's official
// system source maps encoding 3 to 24 MHz.
func configureHSI() {
	py32.RCC.SetICSCR_HSI_FS(3)
}
