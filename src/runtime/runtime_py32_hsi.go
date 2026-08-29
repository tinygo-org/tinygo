//go:build py32 && !py32_hsi_fs_op && !py32_no_hsi_fs

package runtime

import "device/py32"

func configureHSI() {
	py32.RCC.SetICSCR_HSI_FS(py32.RCC_ICSCR_HSI_FS_Freq24MHz)
}
