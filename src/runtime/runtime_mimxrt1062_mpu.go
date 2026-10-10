//go:build mimxrt1062

package runtime

import (
	"device/nxp"
	"unsafe"
)

//go:extern _usb_dma_start
var _usb_dma_start [0]byte

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

	// [3] ITCM: 512 KiB, +ACCESS, #NORMAL (non-cacheable), +EXEC, -share, -subregion
	// TEX 0b001 with C=0 B=0 is Normal non cacheable memory. TEX 0 is Strongly
	// Ordered and unaligned accesses fault there. See Arm DDI 0403, PMSAv7.
	nxp.MPU.SetRBAR(3, 0x00000000)
	nxp.MPU.SetRASR(nxp.RGNSZ_512KB, nxp.PERM_FULL, nxp.Extension(1), true, false, false, false, false)

	// [4] DTCM: 512 KiB, +ACCESS, #NORMAL (non-cacheable), +EXEC, -share, -subregion
	nxp.MPU.SetRBAR(4, 0x20000000)
	nxp.MPU.SetRASR(nxp.RGNSZ_512KB, nxp.PERM_FULL, nxp.Extension(1), true, false, false, false, false)

	// [5] RAM (AXI): 512 KiB, +ACCESS, #NORMAL, +EXEC, -share, +CACHE, +BUFFER, -subregion
	nxp.MPU.SetRBAR(5, 0x20200000)
	nxp.MPU.SetRASR(nxp.RGNSZ_512KB, nxp.PERM_FULL, nxp.EXTN_NORMAL, true, false, true, true, false)

	// [6] FlexSPI: 512 MiB, +ACCESS, #NORMAL, +EXEC, -share, +CACHE, +BUFFER, -subregion
	nxp.MPU.SetRBAR(6, 0x70000000)
	nxp.MPU.SetRASR(nxp.RGNSZ_512MB, nxp.PERM_FULL, nxp.EXTN_NORMAL, true, false, true, true, false)

	// [7] QSPI flash: 2 or 8 MiB, +ACCESS, #NORMAL, +EXEC, -share, +CACHE, +BUFFER, -subregion
	nxp.MPU.SetRBAR(7, 0x60000000)
	nxp.MPU.SetRASR(qspiFlashMPUSize, nxp.PERM_FULL, nxp.EXTN_NORMAL, true, false, true, true, false)

	// [8] USB DMA region, top 4 KiB of OCRAM, #NORMAL non cacheable, -EXEC.
	// It holds the USB dQH and dTD descriptors and the endpoint buffers.
	nxp.MPU.SetRBAR(8, uint32(uintptr(unsafe.Pointer(&_usb_dma_start))))
	nxp.MPU.SetRASR(nxp.RGNSZ_4KB, nxp.PERM_FULL, nxp.Extension(1), false, false, false, false, false)

	nxp.MPU.Enable(true)
}
