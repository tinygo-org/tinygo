// Hand created file. DO NOT DELETE.
// Type definitions, fields, and constants associated with the MPU peripheral of
// the STM32H7x7 family of dual-core MCUs.
// These definitions are applicable to the Cortex-M7 core only.

// +build stm32h7x7_cm7

package stm32

import "device/arm"

func EnableMPU(enable bool) bool {
	if enable {
		// Enable the MPU with privileged access to default memory
		MPU.MPU_CTRL.Set(MPU_MPU_CTRL_PRIVDEFENA_Msk | MPU_MPU_CTRL_ENABLE_Msk)
		// Enable fault exceptions
		SCB_SHCSR.SetBits(SCB_SHCSR_MEMFAULTENA_Msk)
		// Ensure MPU setting take effects
		arm.AsmFull(`
				dsb 0xF
				isb 0xF
			`, nil)
	} else {
		// Make sure outstanding transfers are done
		arm.AsmFull(`
				dmb 0xF
			`, nil)
		// Disable fault exceptions
		SCB_SHCSR.ClearBits(SCB_SHCSR_MEMFAULTENA_Msk)
		// Disable the MPU and clear the control register
		MPU.MPU_CTRL.Set(0)
	}
	return true
}

type MPU_REGION_Type struct {
	Enable           uint8  // Specifies the status of the region.
	Number           uint8  // Specifies the number of the region to protect.
	BaseAddress      uint32 // Specifies the base address of the region to protect.
	Size             uint8  // Specifies the size of the region to protect.
	SubRegionDisable uint8  // Specifies the number of the subregion protection to disable.
	TypeExtField     uint8  // Specifies the TEX field level.
	AccessPermission uint8  // Specifies the region access permission type.
	DisableExec      uint8  // Specifies the instruction access status.
	IsShareable      uint8  // Specifies the shareability status of the protected region.
	IsCacheable      uint8  // Specifies the cacheable status of the region protected.
	IsBufferable     uint8  // Specifies the bufferable status of the protected region.
}

func (r MPU_REGION_Type) Set() bool {
	// Set the Region number
	MPU.MPU_RNR.Set(uint32(r.Number))
	if 0 != r.Enable {
		MPU.MPU_RBAR.Set(r.BaseAddress)
		MPU.MPU_RASR.Set((uint32(r.DisableExec) << MPU_MPU_RASR_XN_Pos) |
			(uint32(r.AccessPermission) << MPU_MPU_RASR_AP_Pos) |
			(uint32(r.TypeExtField) << MPU_MPU_RASR_TEX_Pos) |
			(uint32(r.IsShareable) << MPU_MPU_RASR_S_Pos) |
			(uint32(r.IsCacheable) << MPU_MPU_RASR_C_Pos) |
			(uint32(r.IsBufferable) << MPU_MPU_RASR_B_Pos) |
			(uint32(r.SubRegionDisable) << MPU_MPU_RASR_SRD_Pos) |
			(uint32(r.Size) << MPU_MPU_RASR_SIZE_Pos) |
			(uint32(r.Enable) << MPU_MPU_RASR_ENABLE_Pos))
	} else {
		MPU.MPU_RBAR.Set(0)
		MPU.MPU_RASR.Set(0)
	}
	return true
}
