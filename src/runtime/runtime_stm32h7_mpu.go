//go:build stm32 && stm32h7

package runtime

import (
	"device/arm"
	"runtime/volatile"
	"unsafe"
)

// Cortex-M7 cache size registers (ARM TRM Table 4-2, within SCB address space).
var (
	scbCCSIDR = (*volatile.Register32)(unsafe.Pointer(uintptr(0xE000ED80))) // Cache Size ID Register (R)
	scbCSELR  = (*volatile.Register32)(unsafe.Pointer(uintptr(0xE000ED84))) // Cache Size Selection Register (R/W)
)

// Cortex-M7 cache maintenance registers (ARMv7-M Architecture Ref Manual Table B3-7).
var (
	scbICIALLU = (*volatile.Register32)(unsafe.Pointer(uintptr(0xE000EF50))) // Invalidate all I-cache (W)
	scbDCISW   = (*volatile.Register32)(unsafe.Pointer(uintptr(0xE000EF60))) // Invalidate D-cache by set/way (W)
)

// RASR region attribute presets for this chip's memory map (ARMv7-M
// Architecture Ref Manual §B3.5.5). The register layout itself
// (arm.MPU_Type, arm.MPU_RASR_* field positions) is generic to any
// ARMv7-M core (M3/M4/M7) and lives in device/arm; only these specific
// size/type/permission combinations are STM32H7-specific.
const (
	// SIZE field bits[5:1]: value = log2(region_bytes) - 1.
	mpuRASRSize2MB   = 20 << arm.MPU_RASR_SIZE_Pos // 2MB  = 2^21, field=20
	mpuRASRSize512KB = 18 << arm.MPU_RASR_SIZE_Pos // 512KB = 2^19, field=18
	mpuRASRSize512MB = 28 << arm.MPU_RASR_SIZE_Pos // 512MB = 2^29, field=28

	// AP field bits[26:24].
	mpuRASRAPReadOnly   = 0x6 << arm.MPU_RASR_AP_Pos // Privileged and unprivileged read-only
	mpuRASRAPFullAccess = 0x3 << arm.MPU_RASR_AP_Pos // Full access (privileged and unprivileged)

	// Memory type encodings: TEX bits[21:19], S bit[18], C bit[17], B bit[16].
	// Normal, Write-Through, No Write-Allocate (TEX=000, C=1, B=0, S=0).
	mpuRASRNormalWT = arm.MPU_RASR_C
	// Normal, Write-Back, Write-Allocate (TEX=001, C=1, B=1, S=0).
	mpuRASRNormalWBWA = (1 << arm.MPU_RASR_TEX_Pos) | arm.MPU_RASR_C | arm.MPU_RASR_B
	// Shared Device memory (TEX=000, C=0, B=1, S=1).
	mpuRASRDevice = arm.MPU_RASR_S | arm.MPU_RASR_B
)

// initMPU configures the Cortex-M7 MPU, then enables L1 instruction and data
// caches. Must be called after initCLK() and before any peripheral access.
//
// Memory map configured:
//
//	Region 0: Flash     0x08000000  2MB    Normal WT,   RO, executable
//	Region 1: AXI SRAM  0x24000000  512KB  Normal WBWA, RW, no-execute
//	Region 2: Peripherals 0x40000000 512MB  Shared Device, RW, no-execute
//
// Unmapped regions fall back to the ARMv7-M default privileged map via
// PRIVDEFENA, keeping NVIC/SCB and other PPB accesses strongly-ordered.
func initMPU() {
	// Disable MPU before reconfiguring regions.
	arm.MPU.CTRL.Set(0)
	arm.Asm("dsb 0xF")
	arm.Asm("isb 0xF")

	// Region 0: Flash — Normal, Write-Through, read-only, executable.
	arm.MPU.RNR.Set(0)
	arm.MPU.RBAR.Set(0x08000000)
	arm.MPU.RASR.Set(mpuRASRNormalWT | mpuRASRAPReadOnly | mpuRASRSize2MB | arm.MPU_RASR_ENABLE)

	// Region 1: AXI SRAM — Normal, Write-Back Write-Allocate, full access, no-execute.
	arm.MPU.RNR.Set(1)
	arm.MPU.RBAR.Set(0x24000000)
	arm.MPU.RASR.Set(arm.MPU_RASR_XN | mpuRASRNormalWBWA | mpuRASRAPFullAccess | mpuRASRSize512KB | arm.MPU_RASR_ENABLE)

	// Region 2: Peripherals — Shared Device, full access, no-execute.
	arm.MPU.RNR.Set(2)
	arm.MPU.RBAR.Set(0x40000000)
	arm.MPU.RASR.Set(arm.MPU_RASR_XN | mpuRASRDevice | mpuRASRAPFullAccess | mpuRASRSize512MB | arm.MPU_RASR_ENABLE)

	// Enable MemManage fault so MPU violations raise a MemFault rather than
	// hard-faulting directly.
	arm.SCB.SHCSR.SetBits(arm.SCB_SHCSR_MEMFAULTENA)

	// Enable MPU with privileged default background map.
	arm.MPU.CTRL.Set(arm.MPU_CTRL_ENABLE | arm.MPU_CTRL_PRIVDEFENA)
	arm.Asm("dsb 0xF")
	arm.Asm("isb 0xF")

	// Enable L1 caches now that the MPU defines cacheability for each region.
	initICache()
	initDCache()
}

// initICache invalidates then enables the L1 instruction cache.
func initICache() {
	arm.Asm("dsb 0xF")
	arm.Asm("isb 0xF")
	scbICIALLU.Set(0) // Invalidate all I-cache lines.
	arm.Asm("dsb 0xF")
	arm.Asm("isb 0xF")
	arm.SCB.CCR.SetBits(arm.SCB_CCR_IC)
	arm.Asm("dsb 0xF")
	arm.Asm("isb 0xF")
}

// initDCache invalidates all D-cache lines by set/way then enables the cache.
// Iterates over sets and ways read from CCSIDR so it works for any M7 cache
// size (8–64 KB, always 4-way, 32-byte lines on STM32H743).
func initDCache() {
	scbCSELR.Set(0) // Select L1 D-cache.
	arm.Asm("dsb 0xF")

	ccsidr := scbCCSIDR.Get()
	numSets := (ccsidr >> 13) & 0x7FFF // NUMSETS field (value = sets-1)
	assoc := (ccsidr >> 3) & 0x3FF     // ASSOCIATIVITY field (value = ways-1)

	// Invalidate every set/way. For a 4-way cache the way index occupies
	// bits[31:30] of DCISW; the set index starts at bit 5 (32-byte line = 2^5).
	for set := uint32(0); set <= numSets; set++ {
		for way := uint32(0); way <= assoc; way++ {
			scbDCISW.Set((way << 30) | (set << 5))
		}
	}
	arm.Asm("dsb 0xF")

	arm.SCB.CCR.SetBits(arm.SCB_CCR_DC)
	arm.Asm("dsb 0xF")
	arm.Asm("isb 0xF")
}
