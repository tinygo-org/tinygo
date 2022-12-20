//go:build mimxrt1062

package runtime

import (
	"device/nxp"
)

func initCache() {

	nxp.MPU.Enable(false)

	// add Default [0] region to deny access to whole address space to workaround
	// speculative prefetch. Refer to Arm errata 1013783-B for more details.

	// [0] Default {OVERLAY}: 4 GiB, -access, @device, -exec, -share, -cache, -buffer, -subregion
	nxp.MPU.SetRBAR(0, 0x00000000)
	nxp.MPU.SetRASR(nxp.RGNSZ_4GB, nxp.PERM_NONE, nxp.EXTN_DEVICE, false, false, false, false, false)

	// [1] Peripherals {OVERLAY}: 64 MiB, +ACCESS, @device, +EXEC, -share, -cache, -buffer, -subregion
	nxp.MPU.SetRBAR(1, 0x40000000)
	nxp.MPU.SetRASR(nxp.RGNSZ_64MB, nxp.PERM_FULL, nxp.EXTN_DEVICE, true, false, false, false, false)

	// [2] RAM {OVERLAY}: 1 GiB, +ACCESS, @device, +EXEC, -share, -cache, -buffer, -subregion
	nxp.MPU.SetRBAR(2, 0x00000000)
	nxp.MPU.SetRASR(nxp.RGNSZ_1GB, nxp.PERM_FULL, nxp.EXTN_DEVICE, true, false, false, false, false)

	// [3] ITCM: 512 KiB, +ACCESS, #NORMAL, +EXEC, -share, -cache, -buffer, -subregion
	nxp.MPU.SetRBAR(3, 0x00000000)
	nxp.MPU.SetRASR(nxp.RGNSZ_512KB, nxp.PERM_FULL, nxp.EXTN_NORMAL, true, false, false, false, false)

	// [4] DTCM: 512 KiB, +ACCESS, #NORMAL, +EXEC, -share, -cache, -buffer, -subregion
	nxp.MPU.SetRBAR(4, 0x20000000)
	nxp.MPU.SetRASR(nxp.RGNSZ_512KB, nxp.PERM_FULL, nxp.EXTN_NORMAL, true, false, false, false, false)

	// [5] RAM (AXI): 512 KiB, +ACCESS, #NORMAL, +EXEC, -share, +CACHE, +BUFFER, -subregion
	nxp.MPU.SetRBAR(5, 0x20200000)
	nxp.MPU.SetRASR(nxp.RGNSZ_512KB, nxp.PERM_FULL, nxp.EXTN_NORMAL, true, false, true, true, false)

	// [6] FlexSPI: 512 MiB, +ACCESS, #NORMAL, +EXEC, -share, +CACHE, +BUFFER, -subregion
	nxp.MPU.SetRBAR(6, 0x70000000)
	nxp.MPU.SetRASR(nxp.RGNSZ_512MB, nxp.PERM_FULL, nxp.EXTN_NORMAL, true, false, true, true, false)

	// [7] QSPI flash: 2 MiB, +ACCESS, #NORMAL, +EXEC, -share, +CACHE, +BUFFER, -subregion
	nxp.MPU.SetRBAR(7, 0x60000000)
	nxp.MPU.SetRASR(nxp.RGNSZ_2MB, nxp.PERM_FULL, nxp.EXTN_NORMAL, true, false, true, true, false)

	nxp.MPU.Enable(true)
}
