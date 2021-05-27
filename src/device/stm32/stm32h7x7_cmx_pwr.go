// Hand created file. DO NOT DELETE.
// Type definitions, fields, and constants associated with the PWR peripheral of
// the STM32H7x7 family of dual-core MCUs.
// These definitions are applicable to both the Cortex-M7 and Cortex-M4 cores.

// +build stm32h7x7

package stm32

import (
	"runtime/volatile"
	"unsafe"
)

// SYSCFG power control register offset: 0x2C
var SYSCFG_PWRCR = (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(SYSCFG)) + 0x02C)))

const (
	SYSCFG_PWRCR_ODEN_Pos = 0
	SYSCFG_PWRCR_ODEN_Msk = 0x1 << SYSCFG_PWRCR_ODEN_Pos // 0x00000001
	SYSCFG_PWRCR_ODEN     = SYSCFG_PWRCR_ODEN_Msk        // PWR overdrive enable
)

var (
	// PWR CPU1 control register offset: 0x10
	PWR_CPU1CR = (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(PWR)) + 0x010)))
	// PWR CPU2 control register offset: 0x14
	PWR_CPU2CR = (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(PWR)) + 0x014)))
)

const pwrFlagTimeoutMs = 1000 // 1 second

func SetPowerSupply(supplySource, voltageScale uint32) bool {

	if (PWR_PWR_CR3_SDEN | PWR_PWR_CR3_LDOEN) != (PWR.PWR_CR3.Get() & (PWR_PWR_CR3_SDEN | PWR_PWR_CR3_LDOEN | PWR_PWR_CR3_BYPASS)) {
		// Check supply configuration
		if supplySource != (PWR.PWR_CR3.Get() & PWR_SUPPLY_CONFIG_Msk) {
			// Supply configuration update locked, can't apply a new supply config
			return false
		} else {
			// Supply configuration update locked, but new supply configuration
			// matches old supply configuration; nothing to do.
			return true
		}
	}

	// Set the power supply configuration
	PWR.PWR_CR3.ReplaceBits(supplySource, PWR_SUPPLY_CONFIG_Msk, 0)

	// Wait until voltage level flag is set
	start := ticks()
	for !PWR_FLAG_ACTVOSRDY.Get() {
		if ticks()-start > pwrFlagTimeoutMs {
			return false // timeout
		}
	}

	/* When the SMPS supplies external circuits verify that SDEXTRDY flag is set */
	if (supplySource == PWR_SMPS_1V8_SUPPLIES_EXT_AND_LDO) ||
		(supplySource == PWR_SMPS_2V5_SUPPLIES_EXT_AND_LDO) ||
		(supplySource == PWR_SMPS_1V8_SUPPLIES_EXT) ||
		(supplySource == PWR_SMPS_2V5_SUPPLIES_EXT) {

		// Wait till SMPS external supply ready flag is set
		start = ticks()
		for !PWR_FLAG_SMPSEXTRDY.Get() {
			if ticks()-start > pwrFlagTimeoutMs {
				return false // timeout
			}
		}
	}

	switch voltageScale {
	case PWR_REGULATOR_VOLTAGE_SCALE0:
		// Configure the voltage scaling 1
		PWR.PWR_D3CR.ReplaceBits(PWR_REGULATOR_VOLTAGE_SCALE1, PWR_PWR_D3CR_VOS_Msk, 0)
		// Enable the PWR overdrive
		SYSCFG_PWRCR.SetBits(SYSCFG_PWRCR_ODEN)
	default:
		// Disable the PWR overdrive
		SYSCFG_PWRCR.ClearBits(SYSCFG_PWRCR_ODEN)
		// Configure the voltage scaling x
		PWR.PWR_D3CR.ReplaceBits(voltageScale, PWR_PWR_D3CR_VOS_Msk, 0)
	}

	return true
}

const (
	PWR_LDO_SUPPLY = PWR_PWR_CR3_LDOEN // Core domains are suppplied from the LDO

	PWR_DIRECT_SMPS_SUPPLY            = PWR_PWR_CR3_SDEN                                                                      // Core domains are suppplied from the SMPS only
	PWR_SMPS_1V8_SUPPLIES_LDO         = (PWR_PWR_CR3_SDLEVEL_0 | PWR_PWR_CR3_SDEN | PWR_PWR_CR3_LDOEN)                        // The SMPS 1.8V output supplies the LDO which supplies the Core domains
	PWR_SMPS_2V5_SUPPLIES_LDO         = (PWR_PWR_CR3_SDLEVEL_1 | PWR_PWR_CR3_SDEN | PWR_PWR_CR3_LDOEN)                        // The SMPS 2.5V output supplies the LDO which supplies the Core domains
	PWR_SMPS_1V8_SUPPLIES_EXT_AND_LDO = (PWR_PWR_CR3_SDLEVEL_0 | PWR_PWR_CR3_SDEXTHP | PWR_PWR_CR3_SDEN | PWR_PWR_CR3_LDOEN)  // The SMPS 1.8V output supplies an external circuits and the LDO. The Core domains are suppplied from the LDO
	PWR_SMPS_2V5_SUPPLIES_EXT_AND_LDO = (PWR_PWR_CR3_SDLEVEL_1 | PWR_PWR_CR3_SDEXTHP | PWR_PWR_CR3_SDEN | PWR_PWR_CR3_LDOEN)  // The SMPS 2.5V output supplies an external circuits and the LDO. The Core domains are suppplied from the LDO
	PWR_SMPS_1V8_SUPPLIES_EXT         = (PWR_PWR_CR3_SDLEVEL_0 | PWR_PWR_CR3_SDEXTHP | PWR_PWR_CR3_SDEN | PWR_PWR_CR3_BYPASS) // The SMPS 1.8V output supplies an external source which supplies the Core domains
	PWR_SMPS_2V5_SUPPLIES_EXT         = (PWR_PWR_CR3_SDLEVEL_1 | PWR_PWR_CR3_SDEXTHP | PWR_PWR_CR3_SDEN | PWR_PWR_CR3_BYPASS) // The SMPS 2.5V output supplies an external source which supplies the Core domains

	PWR_EXTERNAL_SOURCE_SUPPLY = PWR_PWR_CR3_BYPASS // The SMPS disabled and the LDO Bypass. The Core domains are supplied from an external source

	PWR_SUPPLY_CONFIG_Msk = (PWR_PWR_CR3_SDLEVEL_Msk | PWR_PWR_CR3_SDEXTHP | PWR_PWR_CR3_SDEN | PWR_PWR_CR3_LDOEN | PWR_PWR_CR3_BYPASS)

	PWR_REGULATOR_VOLTAGE_SCALE0 = 0
	PWR_REGULATOR_VOLTAGE_SCALE1 = PWR_PWR_D3CR_VOS_1 | PWR_PWR_D3CR_VOS_0
	PWR_REGULATOR_VOLTAGE_SCALE2 = PWR_PWR_D3CR_VOS_1
	PWR_REGULATOR_VOLTAGE_SCALE3 = PWR_PWR_D3CR_VOS_0
)

const (
	PWR_PWR_D3CR_VOS_0 = 0x1 << PWR_PWR_D3CR_VOS_Pos // 0x00004000
	PWR_PWR_D3CR_VOS_1 = 0x2 << PWR_PWR_D3CR_VOS_Pos // 0x00008000

	PWR_PWR_CR3_SDLEVEL_0 = (0x1 << PWR_PWR_CR3_SDLEVEL_Pos) // 0x00000010
	PWR_PWR_CR3_SDLEVEL_1 = (0x2 << PWR_PWR_CR3_SDLEVEL_Pos) // 0x00000020

	PWR_PWR_CPUxCR_HOLDyF_Pos = 4
	PWR_PWR_CPUxCR_HOLDyF_Msk = 0x1 << PWR_PWR_CPUxCR_HOLDyF_Pos // 0x00000010
	PWR_PWR_CPUxCR_HOLDyF     = PWR_PWR_CPUxCR_HOLDyF_Msk        // CPU1 in hold wakeup flag

	PWR_PWR_WKUPCR_WKUPC6_Pos = 5
	PWR_PWR_WKUPCR_WKUPC6_Msk = 0x1 << PWR_PWR_WKUPCR_WKUPC6_Pos // 0x00000020
	PWR_PWR_WKUPCR_WKUPC6     = PWR_PWR_WKUPCR_WKUPC6_Msk        // Clear Wakeup Pin Flag 6
	PWR_PWR_WKUPCR_WKUPC5_Pos = 4
	PWR_PWR_WKUPCR_WKUPC5_Msk = 0x1 << PWR_PWR_WKUPCR_WKUPC5_Pos // 0x00000010
	PWR_PWR_WKUPCR_WKUPC5     = PWR_PWR_WKUPCR_WKUPC5_Msk        // Clear Wakeup Pin Flag 5
	PWR_PWR_WKUPCR_WKUPC4_Pos = 3
	PWR_PWR_WKUPCR_WKUPC4_Msk = 0x1 << PWR_PWR_WKUPCR_WKUPC4_Pos // 0x00000008
	PWR_PWR_WKUPCR_WKUPC4     = PWR_PWR_WKUPCR_WKUPC4_Msk        // Clear Wakeup Pin Flag 4
	PWR_PWR_WKUPCR_WKUPC3_Pos = 2
	PWR_PWR_WKUPCR_WKUPC3_Msk = 0x1 << PWR_PWR_WKUPCR_WKUPC3_Pos // 0x00000004
	PWR_PWR_WKUPCR_WKUPC3     = PWR_PWR_WKUPCR_WKUPC3_Msk        // Clear Wakeup Pin Flag 3
	PWR_PWR_WKUPCR_WKUPC2_Pos = 1
	PWR_PWR_WKUPCR_WKUPC2_Msk = 0x1 << PWR_PWR_WKUPCR_WKUPC2_Pos // 0x00000002
	PWR_PWR_WKUPCR_WKUPC2     = PWR_PWR_WKUPCR_WKUPC2_Msk        // Clear Wakeup Pin Flag 2
	PWR_PWR_WKUPCR_WKUPC1_Pos = 0
	PWR_PWR_WKUPCR_WKUPC1_Msk = 0x1 << PWR_PWR_WKUPCR_WKUPC1_Pos // 0x00000001
	PWR_PWR_WKUPCR_WKUPC1     = PWR_PWR_WKUPCR_WKUPC1_Msk        // Clear Wakeup Pin Flag 1
)

type PWR_FLAG_Type uint8

const (
	PWR_FLAG_STOP       PWR_FLAG_Type = 0x01
	PWR_FLAG_SB_D1      PWR_FLAG_Type = 0x02
	PWR_FLAG_SB_D2      PWR_FLAG_Type = 0x03
	PWR_FLAG_SB         PWR_FLAG_Type = 0x04
	PWR_FLAG_CPU1_HOLD  PWR_FLAG_Type = 0x05
	PWR_FLAG_CPU2_HOLD  PWR_FLAG_Type = 0x06
	PWR_FLAG2_STOP      PWR_FLAG_Type = 0x07
	PWR_FLAG2_SB_D1     PWR_FLAG_Type = 0x08
	PWR_FLAG2_SB_D2     PWR_FLAG_Type = 0x09
	PWR_FLAG2_SB        PWR_FLAG_Type = 0x0A
	PWR_FLAG_PVDO       PWR_FLAG_Type = 0x0B
	PWR_FLAG_AVDO       PWR_FLAG_Type = 0x0C
	PWR_FLAG_ACTVOSRDY  PWR_FLAG_Type = 0x0D
	PWR_FLAG_ACTVOS     PWR_FLAG_Type = 0x0E
	PWR_FLAG_BRR        PWR_FLAG_Type = 0x0F
	PWR_FLAG_VOSRDY     PWR_FLAG_Type = 0x10
	PWR_FLAG_SMPSEXTRDY PWR_FLAG_Type = 0x11
	PWR_FLAG_MMCVDO     PWR_FLAG_Type = 0x12
	PWR_FLAG_USB33RDY   PWR_FLAG_Type = 0x13
	PWR_FLAG_TEMPH      PWR_FLAG_Type = 0x14
	PWR_FLAG_TEMPL      PWR_FLAG_Type = 0x15
	PWR_FLAG_VBATH      PWR_FLAG_Type = 0x16
	PWR_FLAG_VBATL      PWR_FLAG_Type = 0x17

	PWR_FLAG_WKUP1 PWR_FLAG_Type = PWR_PWR_WKUPCR_WKUPC1_Msk
	PWR_FLAG_WKUP2 PWR_FLAG_Type = PWR_PWR_WKUPCR_WKUPC2_Msk
	PWR_FLAG_WKUP3 PWR_FLAG_Type = PWR_PWR_WKUPCR_WKUPC3_Msk
	PWR_FLAG_WKUP4 PWR_FLAG_Type = PWR_PWR_WKUPCR_WKUPC4_Msk
	PWR_FLAG_WKUP5 PWR_FLAG_Type = PWR_PWR_WKUPCR_WKUPC5_Msk
	PWR_FLAG_WKUP6 PWR_FLAG_Type = PWR_PWR_WKUPCR_WKUPC6_Msk
)

func (f PWR_FLAG_Type) Get() bool {
	switch f {
	case PWR_FLAG_PVDO:
		return PWR_PWR_CSR1_PVDO_Msk == (PWR.PWR_CSR1.Get() & PWR_PWR_CSR1_PVDO_Msk)
	case PWR_FLAG_AVDO:
		return PWR_PWR_CSR1_AVDO_Msk == (PWR.PWR_CSR1.Get() & PWR_PWR_CSR1_AVDO_Msk)
	case PWR_FLAG_ACTVOSRDY:
		return PWR_PWR_CSR1_ACTVOSRDY_Msk == (PWR.PWR_CSR1.Get() & PWR_PWR_CSR1_ACTVOSRDY_Msk)
	case PWR_FLAG_VOSRDY:
		return PWR_PWR_D3CR_VOSRDY_Msk == (PWR.PWR_D3CR.Get() & PWR_PWR_D3CR_VOSRDY_Msk)
	case PWR_FLAG_SMPSEXTRDY:
		return PWR_PWR_CR3_SDEXTRDY_Msk == (PWR.PWR_CR3.Get() & PWR_PWR_CR3_SDEXTRDY_Msk)
	case PWR_FLAG_BRR:
		return PWR_PWR_CR2_BRRDY_Msk == (PWR.PWR_CR2.Get() & PWR_PWR_CR2_BRRDY_Msk)
	case PWR_FLAG_CPU1_HOLD:
		return PWR_PWR_CPUxCR_HOLDyF_Msk == (PWR_CPU2CR.Get() & PWR_PWR_CPUxCR_HOLDyF_Msk)
	case PWR_FLAG_CPU2_HOLD:
		return PWR_PWR_CPUxCR_HOLDyF_Msk == (PWR_CPU1CR.Get() & PWR_PWR_CPUxCR_HOLDyF_Msk)
	case PWR_FLAG_SB:
		return PWR_PWR_CPUCR_SBF_Msk == (PWR_CPU1CR.Get() & PWR_PWR_CPUCR_SBF_Msk)
	case PWR_FLAG2_SB:
		return PWR_PWR_CPUCR_SBF_Msk == (PWR_CPU2CR.Get() & PWR_PWR_CPUCR_SBF_Msk)
	case PWR_FLAG_STOP:
		return PWR_PWR_CPUCR_STOPF_Msk == (PWR_CPU1CR.Get() & PWR_PWR_CPUCR_STOPF_Msk)
	case PWR_FLAG2_STOP:
		return PWR_PWR_CPUCR_STOPF_Msk == (PWR_CPU2CR.Get() & PWR_PWR_CPUCR_STOPF_Msk)
	case PWR_FLAG_SB_D1:
		return PWR_PWR_CPUCR_SBF_D1_Msk == (PWR_CPU1CR.Get() & PWR_PWR_CPUCR_SBF_D1_Msk)
	case PWR_FLAG2_SB_D1:
		return PWR_PWR_CPUCR_SBF_D1_Msk == (PWR_CPU2CR.Get() & PWR_PWR_CPUCR_SBF_D1_Msk)
	case PWR_FLAG_SB_D2:
		return PWR_PWR_CPUCR_SBF_D2_Msk == (PWR_CPU1CR.Get() & PWR_PWR_CPUCR_SBF_D2_Msk)
	case PWR_FLAG2_SB_D2:
		return PWR_PWR_CPUCR_SBF_D2_Msk == (PWR_CPU2CR.Get() & PWR_PWR_CPUCR_SBF_D2_Msk)
	case PWR_FLAG_USB33RDY:
		return PWR_PWR_CR3_USB33RDY_Msk == (PWR.PWR_CR3.Get() & PWR_PWR_CR3_USB33RDY_Msk)
	case PWR_FLAG_TEMPH:
		return PWR_PWR_CR2_TEMPH_Msk == (PWR.PWR_CR2.Get() & PWR_PWR_CR2_TEMPH_Msk)
	case PWR_FLAG_TEMPL:
		return PWR_PWR_CR2_TEMPL_Msk == (PWR.PWR_CR2.Get() & PWR_PWR_CR2_TEMPL_Msk)
	case PWR_FLAG_VBATH:
		return PWR_PWR_CR2_VBATH_Msk == (PWR.PWR_CR2.Get() & PWR_PWR_CR2_VBATH_Msk)
	case PWR_FLAG_VBATL:
		return PWR_PWR_CR2_VBATL_Msk == (PWR.PWR_CR2.Get() & PWR_PWR_CR2_VBATL_Msk)
	}
	return false
}
