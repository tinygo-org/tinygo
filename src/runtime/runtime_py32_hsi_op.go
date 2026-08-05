//go:build py32 && py32_hsi_fs_op

package runtime

import "device/py32"

func configureHSI() {
	py32.RCC.SetICSCR_HSI_FS_OP(4)
}
