//go:build cortexm

package arm

import (
	"runtime/volatile"
	"unsafe"
)

/*
The struct below has been created using gen-device-svd
on cortex_m/armv7em.yaml from stm32-rs.
The MPU type has been included as some stm32 svd files
do not contain MPU definitions.
*/

const MPU_BASE = SCS_BASE + 0x0D90

// Memory Protection Unit
type MPU_Type struct {
	TYPE    volatile.Register32 // 0x0
	CTRL    volatile.Register32 // 0x4
	RNR     volatile.Register32 // 0x8
	RBAR    volatile.Register32 // 0xC
	RASR    volatile.Register32 // 0x10
	RBAR_A1 volatile.Register32 // 0x14
	RASR_A1 volatile.Register32 // 0x18
	RBAR_A2 volatile.Register32 // 0x1C
	RASR_A2 volatile.Register32 // 0x20
	RBAR_A3 volatile.Register32 // 0x24
	RASR_A3 volatile.Register32 // 0x28
}

var MPU = (*MPU_Type)(unsafe.Pointer(uintptr(MPU_BASE)))

// Constants for MPU: Memory protection unit
const (
	// TYPER: MPU type register
	// Position of SEPARATE field.
	MPU_TYPER_SEPARATE_Pos = 0x0
	// Bit mask of SEPARATE field.
	MPU_TYPER_SEPARATE_Msk = 0x1
	// Bit SEPARATE.
	MPU_TYPER_SEPARATE = 0x1
	// Position of DREGION field.
	MPU_TYPER_DREGION_Pos = 0x8
	// Bit mask of DREGION field.
	MPU_TYPER_DREGION_Msk = 0xff00
	// Position of IREGION field.
	MPU_TYPER_IREGION_Pos = 0x10
	// Bit mask of IREGION field.
	MPU_TYPER_IREGION_Msk = 0xff0000

	// CTRL: MPU control register
	// Position of ENABLE field.
	MPU_CTRL_ENABLE_Pos = 0x0
	// Bit mask of ENABLE field.
	MPU_CTRL_ENABLE_Msk = 0x1
	// Bit ENABLE.
	MPU_CTRL_ENABLE = 0x1
	// Position of HFNMIENA field.
	MPU_CTRL_HFNMIENA_Pos = 0x1
	// Bit mask of HFNMIENA field.
	MPU_CTRL_HFNMIENA_Msk = 0x2
	// Bit HFNMIENA.
	MPU_CTRL_HFNMIENA = 0x2
	// Position of PRIVDEFENA field.
	MPU_CTRL_PRIVDEFENA_Pos = 0x2
	// Bit mask of PRIVDEFENA field.
	MPU_CTRL_PRIVDEFENA_Msk = 0x4
	// Bit PRIVDEFENA.
	MPU_CTRL_PRIVDEFENA = 0x4

	// RNR: MPU region number register
	// Position of REGION field.
	MPU_RNR_REGION_Pos = 0x0
	// Bit mask of REGION field.
	MPU_RNR_REGION_Msk = 0xff

	// RBAR: MPU region base address register
	// Position of REGION field.
	MPU_RBAR_REGION_Pos = 0x0
	// Bit mask of REGION field.
	MPU_RBAR_REGION_Msk = 0xf
	// Position of VALID field.
	MPU_RBAR_VALID_Pos = 0x4
	// Bit mask of VALID field.
	MPU_RBAR_VALID_Msk = 0x10
	// Bit VALID.
	MPU_RBAR_VALID = 0x10
	// Position of ADDR field.
	MPU_RBAR_ADDR_Pos = 0x5
	// Bit mask of ADDR field.
	MPU_RBAR_ADDR_Msk = 0xffffffe0

	// RASR: MPU region attribute and size register
	// Position of ENABLE field.
	MPU_RASR_ENABLE_Pos = 0x0
	// Bit mask of ENABLE field.
	MPU_RASR_ENABLE_Msk = 0x1
	// Bit ENABLE.
	MPU_RASR_ENABLE = 0x1
	// Position of SIZE field.
	MPU_RASR_SIZE_Pos = 0x1
	// Bit mask of SIZE field.
	MPU_RASR_SIZE_Msk = 0x3e
	// Position of SRD field.
	MPU_RASR_SRD_Pos = 0x8
	// Bit mask of SRD field.
	MPU_RASR_SRD_Msk = 0xff00
	// Position of B field.
	MPU_RASR_B_Pos = 0x10
	// Bit mask of B field.
	MPU_RASR_B_Msk = 0x10000
	// Bit B.
	MPU_RASR_B = 0x10000
	// Position of C field.
	MPU_RASR_C_Pos = 0x11
	// Bit mask of C field.
	MPU_RASR_C_Msk = 0x20000
	// Bit C.
	MPU_RASR_C = 0x20000
	// Position of S field.
	MPU_RASR_S_Pos = 0x12
	// Bit mask of S field.
	MPU_RASR_S_Msk = 0x40000
	// Bit S.
	MPU_RASR_S = 0x40000
	// Position of TEX field.
	MPU_RASR_TEX_Pos = 0x13
	// Bit mask of TEX field.
	MPU_RASR_TEX_Msk = 0x380000
	// Position of AP field.
	MPU_RASR_AP_Pos = 0x18
	// Bit mask of AP field.
	MPU_RASR_AP_Msk = 0x7000000
	// Position of XN field.
	MPU_RASR_XN_Pos = 0x1c
	// Bit mask of XN field.
	MPU_RASR_XN_Msk = 0x10000000
	// Bit XN.
	MPU_RASR_XN = 0x10000000
)
