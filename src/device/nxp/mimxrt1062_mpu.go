// Hand created file. DO NOT DELETE.
// Type definitions, fields, and constants associated with the MPU peripheral
// of the NXP MIMXRT1062.

//go:build nxp && mimxrt1062

package nxp

import (
	"device/arm"
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

type (
	RegionSize  uint32
	AccessPerms uint32
	Extension   uint32
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

const (
	RGNSZ_32B   RegionSize = 0x04 // MPU Region Size 32 Bytes
	RGNSZ_64B   RegionSize = 0x05 // MPU Region Size 64 Bytes
	RGNSZ_128B  RegionSize = 0x06 // MPU Region Size 128 Bytes
	RGNSZ_256B  RegionSize = 0x07 // MPU Region Size 256 Bytes
	RGNSZ_512B  RegionSize = 0x08 // MPU Region Size 512 Bytes
	RGNSZ_1KB   RegionSize = 0x09 // MPU Region Size 1 KByte
	RGNSZ_2KB   RegionSize = 0x0A // MPU Region Size 2 KBytes
	RGNSZ_4KB   RegionSize = 0x0B // MPU Region Size 4 KBytes
	RGNSZ_8KB   RegionSize = 0x0C // MPU Region Size 8 KBytes
	RGNSZ_16KB  RegionSize = 0x0D // MPU Region Size 16 KBytes
	RGNSZ_32KB  RegionSize = 0x0E // MPU Region Size 32 KBytes
	RGNSZ_64KB  RegionSize = 0x0F // MPU Region Size 64 KBytes
	RGNSZ_128KB RegionSize = 0x10 // MPU Region Size 128 KBytes
	RGNSZ_256KB RegionSize = 0x11 // MPU Region Size 256 KBytes
	RGNSZ_512KB RegionSize = 0x12 // MPU Region Size 512 KBytes
	RGNSZ_1MB   RegionSize = 0x13 // MPU Region Size 1 MByte
	RGNSZ_2MB   RegionSize = 0x14 // MPU Region Size 2 MBytes
	RGNSZ_4MB   RegionSize = 0x15 // MPU Region Size 4 MBytes
	RGNSZ_8MB   RegionSize = 0x16 // MPU Region Size 8 MBytes
	RGNSZ_16MB  RegionSize = 0x17 // MPU Region Size 16 MBytes
	RGNSZ_32MB  RegionSize = 0x18 // MPU Region Size 32 MBytes
	RGNSZ_64MB  RegionSize = 0x19 // MPU Region Size 64 MBytes
	RGNSZ_128MB RegionSize = 0x1A // MPU Region Size 128 MBytes
	RGNSZ_256MB RegionSize = 0x1B // MPU Region Size 256 MBytes
	RGNSZ_512MB RegionSize = 0x1C // MPU Region Size 512 MBytes
	RGNSZ_1GB   RegionSize = 0x1D // MPU Region Size 1 GByte
	RGNSZ_2GB   RegionSize = 0x1E // MPU Region Size 2 GBytes
	RGNSZ_4GB   RegionSize = 0x1F // MPU Region Size 4 GBytes
)

const (
	PERM_NONE AccessPerms = 0 // MPU Access Permission no access
	PERM_PRIV AccessPerms = 1 // MPU Access Permission privileged access only
	PERM_URO  AccessPerms = 2 // MPU Access Permission unprivileged access read-only
	PERM_FULL AccessPerms = 3 // MPU Access Permission full access
	PERM_PRO  AccessPerms = 5 // MPU Access Permission privileged access read-only
	PERM_RO   AccessPerms = 6 // MPU Access Permission read-only access
)

const (
	EXTN_NORMAL Extension = 0
	EXTN_DEVICE Extension = 2
)

func (mpu *MPU_Type) Enable(enable bool) {
	if enable {
		mpu.CTRL.Set(MPU_CTRL_PRIVDEFENA_Msk | MPU_CTRL_ENABLE_Msk)
		SystemControl.SHCSR.SetBits(SCB_SHCSR_MEMFAULTENA_Msk)
		arm.Asm("dsb 0xF")
		arm.Asm("isb 0xF")
		enableDcache(true)
		enableIcache(true)
	} else {
		enableIcache(false)
		enableDcache(false)
		arm.Asm("dmb 0xF")
		SystemControl.SHCSR.ClearBits(SCB_SHCSR_MEMFAULTENA_Msk)
		mpu.CTRL.ClearBits(MPU_CTRL_ENABLE_Msk)
	}
}

// MPU Region Base Address Register value
func (mpu *MPU_Type) SetRBAR(region uint32, baseAddress uint32) {
	mpu.RBAR.Set((baseAddress & MPU_RBAR_ADDR_Msk) |
		(region & MPU_RBAR_REGION_Msk) | MPU_RBAR_VALID_Msk)
}

// MPU Region Attribute and Size Register value
func (mpu *MPU_Type) SetRASR(size RegionSize, access AccessPerms, ext Extension, exec, share, cache, buffer, disable bool) {
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
	if enable != SystemControl.CCR.HasBits(SCB_CCR_IC_Msk) {
		if enable {
			arm.Asm("dsb 0xF")
			arm.Asm("isb 0xF")
			SystemControl.ICIALLU.Set(0)
			arm.Asm("dsb 0xF")
			arm.Asm("isb 0xF")
			SystemControl.CCR.SetBits(SCB_CCR_IC_Msk)
			arm.Asm("dsb 0xF")
			arm.Asm("isb 0xF")
		} else {
			arm.Asm("dsb 0xF")
			arm.Asm("isb 0xF")
			SystemControl.CCR.ClearBits(SCB_CCR_IC_Msk)
			SystemControl.ICIALLU.Set(0)
			arm.Asm("dsb 0xF")
			arm.Asm("isb 0xF")
		}
	}
}

var (
	dcacheCcsidr volatile.Register32
	dcacheSets   volatile.Register32
	dcacheWays   volatile.Register32
)

func enableDcache(enable bool) {
	if enable != SystemControl.CCR.HasBits(SCB_CCR_DC_Msk) {
		if enable {
			SystemControl.CSSELR.Set(0)
			arm.Asm("dsb 0xF")
			ccsidr := SystemControl.CCSIDR.Get()
			sets := (ccsidr & SCB_CCSIDR_NUMSETS_Msk) >> SCB_CCSIDR_NUMSETS_Pos
			for sets != 0 {
				ways := (ccsidr & SCB_CCSIDR_ASSOCIATIVITY_Msk) >> SCB_CCSIDR_ASSOCIATIVITY_Pos
				for ways != 0 {
					SystemControl.DCISW.Set(
						((sets << SCB_DCISW_SET_Pos) & SCB_DCISW_SET_Msk) |
							((ways << SCB_DCISW_WAY_Pos) & SCB_DCISW_WAY_Msk))
					ways--
				}
				sets--
			}
			arm.Asm("dsb 0xF")
			SystemControl.CCR.SetBits(SCB_CCR_DC_Msk)
			arm.Asm("dsb 0xF")
			arm.Asm("isb 0xF")
		} else {
			SystemControl.CSSELR.Set(0)
			arm.Asm("dsb 0xF")
			SystemControl.CCR.ClearBits(SCB_CCR_DC_Msk)
			arm.Asm("dsb 0xF")
			dcacheCcsidr.Set(SystemControl.CCSIDR.Get())
			dcacheSets.Set((dcacheCcsidr.Get() & SCB_CCSIDR_NUMSETS_Msk) >> SCB_CCSIDR_NUMSETS_Pos)
			for dcacheSets.Get() != 0 {
				dcacheWays.Set((dcacheCcsidr.Get() & SCB_CCSIDR_ASSOCIATIVITY_Msk) >> SCB_CCSIDR_ASSOCIATIVITY_Pos)
				for dcacheWays.Get() != 0 {
					SystemControl.DCCISW.Set(
						((dcacheSets.Get() << SCB_DCCISW_SET_Pos) & SCB_DCCISW_SET_Msk) |
							((dcacheWays.Get() << SCB_DCCISW_WAY_Pos) & SCB_DCCISW_WAY_Msk))
					dcacheWays.Set(dcacheWays.Get() - 1)
				}
				dcacheSets.Set(dcacheSets.Get() - 1)
			}
			arm.Asm("dsb 0xF")
			arm.Asm("isb 0xF")
		}
	}
}
