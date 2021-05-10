// Hand created file. DO NOT DELETE.
// Type definitions, fields, and constants associated with the MPU peripheral of
// the STM32H7x7 family of dual-core MCUs.
// These definitions are applicable to both the Cortex-M7 and Cortex-M4 cores.

// +build stm32h7x7_cm4 stm32h7x7_cm7

package stm32

import "device/arm"

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

func (mpu *MPU_Type) Enable(enable bool) bool {
	if enable {
		// Enable the MPU with privileged access to default memory
		mpu.MPU_CTRL.Set(MPU_MPU_CTRL_PRIVDEFENA_Msk | MPU_MPU_CTRL_ENABLE_Msk)
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
		mpu.MPU_CTRL.Set(0)
	}
	return true
}

func (mpu *MPU_Type) Configure(region MPU_REGION_Type) bool {
	// Set the Region number
	mpu.MPU_RNR.Set(uint32(region.Number))
	if 0 != region.Enable {
		mpu.MPU_RBAR.Set(region.BaseAddress)
		mpu.MPU_RASR.Set((uint32(region.DisableExec) << MPU_MPU_RASR_XN_Pos) |
			(uint32(region.AccessPermission) << MPU_MPU_RASR_AP_Pos) |
			(uint32(region.TypeExtField) << MPU_MPU_RASR_TEX_Pos) |
			(uint32(region.IsShareable) << MPU_MPU_RASR_S_Pos) |
			(uint32(region.IsCacheable) << MPU_MPU_RASR_C_Pos) |
			(uint32(region.IsBufferable) << MPU_MPU_RASR_B_Pos) |
			(uint32(region.SubRegionDisable) << MPU_MPU_RASR_SRD_Pos) |
			(uint32(region.Size) << MPU_MPU_RASR_SIZE_Pos) |
			(uint32(region.Enable) << MPU_MPU_RASR_ENABLE_Pos))
	} else {
		mpu.MPU_RBAR.Set(0)
		mpu.MPU_RASR.Set(0)
	}
	return true
}
