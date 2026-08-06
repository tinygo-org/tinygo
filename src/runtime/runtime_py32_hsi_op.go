//go:build py32 && py32_hsi_fs_op

package runtime

import "device/py32"

// The official L090/T09x headers provide HSI_FS_OP position and mask helpers,
// but no value definitions. Encoding 4 selects the runtime's 24 MHz clock.
func configureHSI() {
	py32.RCC.SetICSCR_HSI_FS_OP(4)
}
