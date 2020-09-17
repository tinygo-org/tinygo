// +build mimxrt1062

package runtime

import (
	"device/arm"
	"device/nxp"
	"runtime/volatile"
	"unsafe"
)

type MPU_Type struct {
	TYPE    volatile.Register32 // 0x000 (R/ ) - MPU Type Register
	CTRL    volatile.Register32 // 0x004 (R/W) - MPU Control Register
	RNR     volatile.Register32 // 0x008 (R/W) - MPU Region RNRber Register
	RBAR    volatile.Register32 // 0x00C (R/W) - MPU Region Base Address Register
	RASR    volatile.Register32 // 0x010 (R/W) - MPU Region Attribute and Size Register
	RBAR_A1 volatile.Register32 // 0x014 (R/W) - MPU Alias 1 Region Base Address Register
	RASR_A1 volatile.Register32 // 0x018 (R/W) - MPU Alias 1 Region Attribute and Size Register
	RBAR_A2 volatile.Register32 // 0x01C (R/W) - MPU Alias 2 Region Base Address Register
	RASR_A2 volatile.Register32 // 0x020 (R/W) - MPU Alias 2 Region Attribute and Size Register
	RBAR_A3 volatile.Register32 // 0x024 (R/W) - MPU Alias 3 Region Base Address Register
	RASR_A3 volatile.Register32 // 0x028 (R/W) - MPU Alias 3 Region Attribute and Size Register
}

var MPU = (*MPU_Type)(unsafe.Pointer(uintptr(0xe000ed90)))

func initCache() {
	MPU.initialize()
}

func (mpu *MPU_Type) initialize() {

	mpu.enable(false)

	// -------------------------------------------------------- OVERLAY REGIONS --

	// add Default [0] region to deny access to whole address space to workaround
	// speculative prefetch. Refer to Arm errata 1013783-B for more details.

	// [0] Default {OVERLAY}:
	//          4 GiB, -access, @device, -exec, -share, -cache, -buffer, -subregion
	mpu.setRBAR(0, 0x00000000)
	mpu.setRASR(rs4GB, apNone, exDevice, false, false, false, false, false)

	// [1] Peripherals {OVERLAY}:
	//          64 MiB, +ACCESS, @device, +EXEC, -share, -cache, -buffer, -subregion
	mpu.setRBAR(1, 0x40000000)
	mpu.setRASR(rs64MB, apFull, exDevice, true, false, false, false, false)

	// [2] RAM {OVERLAY}:
	//          1 GiB, +ACCESS, @device, +EXEC, -share, -cache, -buffer, -subregion
	mpu.setRBAR(2, 0x00000000)
	mpu.setRASR(rs1GB, apFull, exDevice, true, false, false, false, false)

	// ----------------------------------------------------- PERIPHERAL REGIONS --

	// [3] ITCM:
	//          512 KiB, +ACCESS, #NORMAL, +EXEC, -share, -cache, -buffer, -subregion
	mpu.setRBAR(3, 0x00000000)
	mpu.setRASR(rs512KB, apFull, exNormal, true, false, false, false, false)

	// [4] DTCM:
	//          512 KiB, +ACCESS, #NORMAL, +EXEC, -share, -cache, -buffer, -subregion
	mpu.setRBAR(4, 0x20000000)
	mpu.setRASR(rs512KB, apFull, exNormal, true, false, false, false, false)

	// [5] RAM (AXI):
	//          512 KiB, +ACCESS, #NORMAL, +EXEC, -share, +CACHE, +BUFFER, -subregion
	mpu.setRBAR(5, 0x20200000)
	mpu.setRASR(rs512KB, apFull, exNormal, true, false, true, true, false)

	// [6] FlexSPI:
	//          512 MiB, +ACCESS, #NORMAL, +EXEC, -share, +CACHE, +BUFFER, -subregion
	mpu.setRBAR(6, 0x70000000)
	mpu.setRASR(rs512MB, apFull, exNormal, true, false, true, true, false)

	// [7] QSPI flash:
	//          2 MiB, +ACCESS, #NORMAL, +EXEC, -share, +CACHE, +BUFFER, -subregion
	mpu.setRBAR(7, 0x60000000)
	mpu.setRASR(rs2MB, apFull, exNormal, true, false, true, true, false)

	mpu.enable(true)
}

// MPU Type Register Definitions
const (
	MPU_TYPE_IREGION_Pos  = 16                           // MPU TYPE: IREGION Position
	MPU_TYPE_IREGION_Msk  = 0xFF << MPU_TYPE_IREGION_Pos // MPU TYPE: IREGION Mask
	MPU_TYPE_DREGION_Pos  = 8                            // MPU TYPE: DREGION Position
	MPU_TYPE_DREGION_Msk  = 0xFF << MPU_TYPE_DREGION_Pos // MPU TYPE: DREGION Mask
	MPU_TYPE_SEPARATE_Pos = 0                            // MPU TYPE: SEPARATE Position
	MPU_TYPE_SEPARATE_Msk = 1                            // MPU TYPE: SEPARATE Mask
)

// MPU Control Register Definitions
const (
	MPU_CTRL_PRIVDEFENA_Pos = 2                            // MPU CTRL: PRIVDEFENA Position
	MPU_CTRL_PRIVDEFENA_Msk = 1 << MPU_CTRL_PRIVDEFENA_Pos // MPU CTRL: PRIVDEFENA Mask
	MPU_CTRL_HFNMIENA_Pos   = 1                            // MPU CTRL: HFNMIENA Position
	MPU_CTRL_HFNMIENA_Msk   = 1 << MPU_CTRL_HFNMIENA_Pos   // MPU CTRL: HFNMIENA Mask
	MPU_CTRL_ENABLE_Pos     = 0                            // MPU CTRL: ENABLE Position
	MPU_CTRL_ENABLE_Msk     = 1                            // MPU CTRL: ENABLE Mask
)

// MPU Region Number Register Definitions
const (
	MPU_RNR_REGION_Pos = 0    // MPU RNR: REGION Position
	MPU_RNR_REGION_Msk = 0xFF // MPU RNR: REGION Mask
)

// MPU Region Base Address Register Definitions
const (
	MPU_RBAR_ADDR_Pos   = 5                              // MPU RBAR: ADDR Position
	MPU_RBAR_ADDR_Msk   = 0x7FFFFFF << MPU_RBAR_ADDR_Pos // MPU RBAR: ADDR Mask
	MPU_RBAR_VALID_Pos  = 4                              // MPU RBAR: VALID Position
	MPU_RBAR_VALID_Msk  = 1 << MPU_RBAR_VALID_Pos        // MPU RBAR: VALID Mask
	MPU_RBAR_REGION_Pos = 0                              // MPU RBAR: REGION Position
	MPU_RBAR_REGION_Msk = 0xF                            // MPU RBAR: REGION Mask
)

// MPU Region Attribute and Size Register Definitions
const (
	MPU_RASR_ATTRS_Pos  = 16                           // MPU RASR: MPU Region Attribute field Position
	MPU_RASR_ATTRS_Msk  = 0xFFFF << MPU_RASR_ATTRS_Pos // MPU RASR: MPU Region Attribute field Mask
	MPU_RASR_XN_Pos     = 28                           // MPU RASR: ATTRS.XN Position
	MPU_RASR_XN_Msk     = 1 << MPU_RASR_XN_Pos         // MPU RASR: ATTRS.XN Mask
	MPU_RASR_AP_Pos     = 24                           // MPU RASR: ATTRS.AP Position
	MPU_RASR_AP_Msk     = 0x7 << MPU_RASR_AP_Pos       // MPU RASR: ATTRS.AP Mask
	MPU_RASR_TEX_Pos    = 19                           // MPU RASR: ATTRS.TEX Position
	MPU_RASR_TEX_Msk    = 0x7 << MPU_RASR_TEX_Pos      // MPU RASR: ATTRS.TEX Mask
	MPU_RASR_S_Pos      = 18                           // MPU RASR: ATTRS.S Position
	MPU_RASR_S_Msk      = 1 << MPU_RASR_S_Pos          // MPU RASR: ATTRS.S Mask
	MPU_RASR_C_Pos      = 17                           // MPU RASR: ATTRS.C Position
	MPU_RASR_C_Msk      = 1 << MPU_RASR_C_Pos          // MPU RASR: ATTRS.C Mask
	MPU_RASR_B_Pos      = 16                           // MPU RASR: ATTRS.B Position
	MPU_RASR_B_Msk      = 1 << MPU_RASR_B_Pos          // MPU RASR: ATTRS.B Mask
	MPU_RASR_SRD_Pos    = 8                            // MPU RASR: Sub-Region Disable Position
	MPU_RASR_SRD_Msk    = 0xFF << MPU_RASR_SRD_Pos     // MPU RASR: Sub-Region Disable Mask
	MPU_RASR_SIZE_Pos   = 1                            // MPU RASR: Region Size Field Position
	MPU_RASR_SIZE_Msk   = 0x1F << MPU_RASR_SIZE_Pos    // MPU RASR: Region Size Field Mask
	MPU_RASR_ENABLE_Pos = 0                            // MPU RASR: Region enable bit Position
	MPU_RASR_ENABLE_Msk = 1                            // MPU RASR: Region enable bit Disable Mask
)

const (
	SCB_DCISW_WAY_Pos = 30                         // SCB DCISW: Way Position
	SCB_DCISW_WAY_Msk = 3 << SCB_DCISW_WAY_Pos     // SCB DCISW: Way Mask
	SCB_DCISW_SET_Pos = 5                          // SCB DCISW: Set Position
	SCB_DCISW_SET_Msk = 0x1FF << SCB_DCISW_SET_Pos // SCB DCISW: Set Mask
)

const (
	SCB_DCCISW_WAY_Pos = 30                          // SCB DCCISW: Way Position
	SCB_DCCISW_WAY_Msk = 3 << SCB_DCCISW_WAY_Pos     // SCB DCCISW: Way Mask
	SCB_DCCISW_SET_Pos = 5                           // SCB DCCISW: Set Position
	SCB_DCCISW_SET_Msk = 0x1FF << SCB_DCCISW_SET_Pos // SCB DCCISW: Set Mask
)

type regionSize uint32

const (
	rs32B   regionSize = 0x04 // MPU Region Size 32 Bytes
	rs64B   regionSize = 0x05 // MPU Region Size 64 Bytes
	rs128B  regionSize = 0x06 // MPU Region Size 128 Bytes
	rs256B  regionSize = 0x07 // MPU Region Size 256 Bytes
	rs512B  regionSize = 0x08 // MPU Region Size 512 Bytes
	rs1KB   regionSize = 0x09 // MPU Region Size 1 KByte
	rs2KB   regionSize = 0x0A // MPU Region Size 2 KBytes
	rs4KB   regionSize = 0x0B // MPU Region Size 4 KBytes
	rs8KB   regionSize = 0x0C // MPU Region Size 8 KBytes
	rs16KB  regionSize = 0x0D // MPU Region Size 16 KBytes
	rs32KB  regionSize = 0x0E // MPU Region Size 32 KBytes
	rs64KB  regionSize = 0x0F // MPU Region Size 64 KBytes
	rs128KB regionSize = 0x10 // MPU Region Size 128 KBytes
	rs256KB regionSize = 0x11 // MPU Region Size 256 KBytes
	rs512KB regionSize = 0x12 // MPU Region Size 512 KBytes
	rs1MB   regionSize = 0x13 // MPU Region Size 1 MByte
	rs2MB   regionSize = 0x14 // MPU Region Size 2 MBytes
	rs4MB   regionSize = 0x15 // MPU Region Size 4 MBytes
	rs8MB   regionSize = 0x16 // MPU Region Size 8 MBytes
	rs16MB  regionSize = 0x17 // MPU Region Size 16 MBytes
	rs32MB  regionSize = 0x18 // MPU Region Size 32 MBytes
	rs64MB  regionSize = 0x19 // MPU Region Size 64 MBytes
	rs128MB regionSize = 0x1A // MPU Region Size 128 MBytes
	rs256MB regionSize = 0x1B // MPU Region Size 256 MBytes
	rs512MB regionSize = 0x1C // MPU Region Size 512 MBytes
	rs1GB   regionSize = 0x1D // MPU Region Size 1 GByte
	rs2GB   regionSize = 0x1E // MPU Region Size 2 GBytes
	rs4GB   regionSize = 0x1F // MPU Region Size 4 GBytes
)

type accessPerms uint32

const (
	apNone accessPerms = 0 // MPU Access Permission no access
	apPriv accessPerms = 1 // MPU Access Permission privileged access only
	apURO  accessPerms = 2 // MPU Access Permission unprivileged access read-only
	apFull accessPerms = 3 // MPU Access Permission full access
	apPRO  accessPerms = 5 // MPU Access Permission privileged access read-only
	apRO   accessPerms = 6 // MPU Access Permission read-only access
)

type extension uint32

const (
	exNormal extension = 0
	exDevice extension = 2
)

func (mpu *MPU_Type) enable(enable bool) {
	if enable {
		mpu.CTRL.Set(MPU_CTRL_PRIVDEFENA_Msk | MPU_CTRL_ENABLE_Msk)
		nxp.SystemControl.SHCSR.SetBits(nxp.SCB_SHCSR_MEMFAULTENA_Msk)
		arm.AsmFull(`
			dsb 0xF
			isb 0xF
		`, nil)
		enableDcache(true)
		enableIcache(true)
	} else {
		enableIcache(false)
		enableDcache(false)
		arm.AsmFull(`
			dmb 0xF
		`, nil)
		nxp.SystemControl.SHCSR.ClearBits(nxp.SCB_SHCSR_MEMFAULTENA_Msk)
		mpu.CTRL.ClearBits(MPU_CTRL_ENABLE_Msk)
	}
}

// MPU Region Base Address Register value
func (mpu *MPU_Type) setRBAR(region uint32, baseAddress uint32) {
	mpu.RBAR.Set((baseAddress & MPU_RBAR_ADDR_Msk) |
		(region & MPU_RBAR_REGION_Msk) | MPU_RBAR_VALID_Msk)
}

// MPU Region Attribute and Size Register value
func (mpu *MPU_Type) setRASR(size regionSize, access accessPerms, ext extension, exec, share, cache, buffer, disable bool) {
	boolBit := func(b bool) uint32 {
		if b {
			return 1
		}
		return 0
	}
	attr := ((uint32(ext) << MPU_RASR_TEX_Pos) & MPU_RASR_TEX_Msk) |
		((boolBit(share) << MPU_RASR_S_Pos) & MPU_RASR_S_Msk) |
		((boolBit(cache) << MPU_RASR_C_Pos) & MPU_RASR_C_Msk) |
		((boolBit(buffer) << MPU_RASR_B_Pos) & MPU_RASR_B_Msk)
	mpu.RASR.Set(((boolBit(!exec) << MPU_RASR_XN_Pos) & MPU_RASR_XN_Msk) |
		((uint32(access) << MPU_RASR_AP_Pos) & MPU_RASR_AP_Msk) |
		(attr & (MPU_RASR_TEX_Msk | MPU_RASR_S_Msk | MPU_RASR_C_Msk | MPU_RASR_B_Msk)) |
		((boolBit(disable) << MPU_RASR_SRD_Pos) & MPU_RASR_SRD_Msk) |
		((uint32(size) << MPU_RASR_SIZE_Pos) & MPU_RASR_SIZE_Msk) |
		MPU_RASR_ENABLE_Msk)
}

func enableIcache(enable bool) {
	if enable != nxp.SystemControl.CCR.HasBits(nxp.SCB_CCR_IC_Msk) {
		if enable {
			arm.AsmFull(`
				dsb 0xF
				isb 0xF
			`, nil)
			nxp.SystemControl.ICIALLU.Set(0)
			arm.AsmFull(`
				dsb 0xF
				isb 0xF
			`, nil)
			nxp.SystemControl.CCR.SetBits(nxp.SCB_CCR_IC_Msk)
			arm.AsmFull(`
				dsb 0xF
				isb 0xF
			`, nil)
		} else {
			arm.AsmFull(`
				dsb 0xF
				isb 0xF
			`, nil)
			nxp.SystemControl.CCR.ClearBits(nxp.SCB_CCR_IC_Msk)
			nxp.SystemControl.ICIALLU.Set(0)
			arm.AsmFull(`
				dsb 0xF
				isb 0xF
			`, nil)
		}
	}
}

func enableDcache(enable bool) {
	if enable != nxp.SystemControl.CCR.HasBits(nxp.SCB_CCR_DC_Msk) {
		if enable {
			nxp.SystemControl.CSSELR.Set(0)
			arm.AsmFull(`
				dsb 0xF
			`, nil)
			ccsidr := nxp.SystemControl.CCSIDR.Get()
			sets := (ccsidr & nxp.SCB_CCSIDR_NUMSETS_Msk) >> nxp.SCB_CCSIDR_NUMSETS_Pos
			for sets != 0 {
				ways := (ccsidr & nxp.SCB_CCSIDR_ASSOCIATIVITY_Msk) >> nxp.SCB_CCSIDR_ASSOCIATIVITY_Pos
				for ways != 0 {
					nxp.SystemControl.DCISW.Set(
						((sets << SCB_DCISW_SET_Pos) & SCB_DCISW_SET_Msk) |
							((ways << SCB_DCISW_WAY_Pos) & SCB_DCISW_WAY_Msk))
					ways--
				}
				sets--
			}
			arm.AsmFull(`
				dsb 0xF
			`, nil)
			nxp.SystemControl.CCR.SetBits(nxp.SCB_CCR_DC_Msk)
			arm.AsmFull(`
				dsb 0xF
				isb 0xF
			`, nil)
		} else {
			var (
				ccsidr volatile.Register32
				sets   volatile.Register32
				ways   volatile.Register32
			)
			nxp.SystemControl.CSSELR.Set(0)
			arm.AsmFull(`
				dsb 0xF
			`, nil)
			nxp.SystemControl.CCR.ClearBits(nxp.SCB_CCR_DC_Msk)
			arm.AsmFull(`
				dsb 0xF
			`, nil)
			ccsidr.Set(nxp.SystemControl.CCSIDR.Get())
			sets.Set((ccsidr.Get() & nxp.SCB_CCSIDR_NUMSETS_Msk) >> nxp.SCB_CCSIDR_NUMSETS_Pos)
			for sets.Get() != 0 {
				ways.Set((ccsidr.Get() & nxp.SCB_CCSIDR_ASSOCIATIVITY_Msk) >> nxp.SCB_CCSIDR_ASSOCIATIVITY_Pos)
				for ways.Get() != 0 {
					nxp.SystemControl.DCCISW.Set(
						((sets.Get() << SCB_DCCISW_SET_Pos) & SCB_DCCISW_SET_Msk) |
							((ways.Get() << SCB_DCCISW_WAY_Pos) & SCB_DCCISW_WAY_Msk))
					ways.Set(ways.Get() - 1)
				}
				sets.Set(sets.Get() - 1)
			}
			arm.AsmFull(`
				dsb 0xF
				isb 0xF
			`, nil)
		}
	}
}
