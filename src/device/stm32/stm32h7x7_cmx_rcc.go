// Hand created file. DO NOT DELETE.
// Type definitions, fields, and constants associated with the RCC peripheral of
// the STM32H7x7 family of dual-core MCUs.
// These definitions are applicable to both the Cortex-M7 and Cortex-M4 cores.

// +build stm32h7x7

package stm32

import (
	"runtime/volatile"
	"unsafe"
)

// Default oscillator frequencies
const (
	// The HSE default is derived from STM32H7 Nucleo(s) and Arduino Portenta H7,
	// but it may not be the same on every device with an STM32H7x7.
	RCC_HSE_FREQ_HZ uint32 = 25000000 // 25 MHz (high-speed, external)
	RCC_HSI_FREQ_HZ uint32 = 64000000 // 64 MHz (high-speed, internal	)
	RCC_CSI_FREQ_HZ uint32 = 4000000  //  4 MHz (low-power, internal)
)

var (
	RCC_CORE1 = (*RCC_CORE_Type)(unsafe.Pointer(uintptr(0x58024530)))
	RCC_CORE2 = (*RCC_CORE_Type)(unsafe.Pointer(uintptr(0x58024590)))
)

//go:linkname ticks runtime.ticks
func ticks() int64

// Force Cortex-M4 core to boot (if held by option byte BCM4 = 0)
func BootM4() {
	RCC.GCR.SetBits(RCC_GCR_BOOT_C2)
}

// Check if Cortex-M4 core was booted by force
func IsBootM4() bool {
	return RCC.GCR.HasBits(RCC_GCR_BOOT_C2)
}

// Force Cortex-M7 core to boot (if held by option byte BCM7 = 0)
func BootM7() {
	RCC.GCR.SetBits(RCC_GCR_BOOT_C1)
}

// Check if Cortex-M7 core was booted by force
func IsBootM7() bool {
	return RCC.GCR.HasBits(RCC_GCR_BOOT_C1)
}

// ClockFreq returns the actual frequency of all internal CPU clocks.
func ClockFreq() (clk RCC_CLK_Type) {
	clk.SYSCLK = sysFreq()
	clk.HCLK = clk.SYSCLK >> (rccCorePrescaler[(((RCC.D1CFGR.Get()&
		RCC_D1CFGR_HPRE_Msk)>>RCC_D1CFGR_HPRE_Pos)&RCC_D1CFGR_HPRE_Msk)>>
		RCC_D1CFGR_HPRE_Pos] & 0x1F)
	clk.PCLK1 = clk.HCLK >> (rccCorePrescaler[(((RCC.D2CFGR.Get()&
		RCC_D2CFGR_D2PPRE1_Msk)>>RCC_D2CFGR_D2PPRE1_Pos)&RCC_D2CFGR_D2PPRE1_Msk)>>
		RCC_D2CFGR_D2PPRE1_Pos] & 0x1F)
	clk.PCLK2 = clk.HCLK >> (rccCorePrescaler[(((RCC.D2CFGR.Get()&
		RCC_D2CFGR_D2PPRE2_Msk)>>RCC_D2CFGR_D2PPRE2_Pos)&RCC_D2CFGR_D2PPRE2_Msk)>>
		RCC_D2CFGR_D2PPRE2_Pos] & 0x1F)
	clk.PCLK3 = clk.HCLK >> (rccCorePrescaler[(((RCC.D1CFGR.Get()&
		RCC_D1CFGR_D1PPRE_Msk)>>RCC_D1CFGR_D1PPRE_Pos)&RCC_D1CFGR_D1PPRE_Msk)>>
		RCC_D1CFGR_D1PPRE_Pos] & 0x1F)
	clk.PCLK4 = clk.HCLK >> (rccCorePrescaler[(((RCC.D3CFGR.Get()&
		RCC_D3CFGR_D3PPRE_Msk)>>RCC_D3CFGR_D3PPRE_Pos)&RCC_D3CFGR_D3PPRE_Msk)>>
		RCC_D3CFGR_D3PPRE_Pos] & 0x1F)
	return
}

// Allocate enables the receiver CPU core's access to the given peripheral.
// By default, some peripherals are only accessible from a single core, and the
// other core must explicitly request read/write access. Official ST reference
// manuals refer to this as "allocation".
func Allocate(bus unsafe.Pointer, core *RCC_CORE_Type) bool {
	var reg *volatile.Register32
	var msk uint32
	switch bus {
	case unsafe.Pointer(FLASH):
		switch core {
		case RCC_CORE1:
			// FLASH is implicitly allocated to core 1 (M7)
			return true
		case RCC_CORE2:
			reg, msk = &RCC.AHB3ENR, RCC_AHB3ENR_FLASHEN
		}
	}
	if nil != reg {
		reg.SetBits(msk)
		// msk must be 1-bit for this to be correct (always true, I think?)
		return reg.HasBits(msk)
	}
	return false // peripheral not supported for the given core
}

// RCC_CLK_Type holds the clock frequencies for all internal CPU clocks. Use
// method (*RCC_Type).ClockFreq() to obtain an RCC_CLK_Type instance populated
// with the current, actual core frequencies.
type RCC_CLK_Type struct {
	SYSCLK uint32 // System core
	CPUCLK uint32 // Cortex-M
	HCLK   uint32 // AHB Bus
	PCLK1  uint32 // APB1
	PCLK2  uint32 // APB2
	PCLK3  uint32 // APB3
	PCLK4  uint32 // APB4
}

type RCC_PLL_PQR_Type struct {
	P, Q, R uint32 // (1-128) division factor for peripheral clocks (P = even)
}

// RCC_PLL_DIV_Type represents PLL frequency parameters.
type RCC_PLL_DIV_Type struct {
	M   uint32 // (1-63) division factor for PLL VCO input clock
	N   uint32 // (4-512) Multiplication factor for PLL VCO output clock
	Nf  uint32 // (0-8191) fractional part of multiplication factor for PLL VCO
	PQR RCC_PLL_PQR_Type
	VCI uint32 // PLL clock input range
	VCO uint32 // PLL clock output range
}

// RCC_PLL_CFG_Type defines an RCC PLL configuration.
type RCC_PLL_CFG_Type struct {
	PLL uint32 // PLL state
	Src uint32 // clock source
	Div RCC_PLL_DIV_Type
}

// RCC_OSC_CFG_Type defines an RCC oscillator (HSE, HSI, CSI, LSE, or LSI)
// configuration.
type RCC_OSC_CFG_Type struct {
	OSC     uint32           // oscillator to be configured
	HSE     uint32           // HSE state
	LSE     uint32           // LSE state
	HSI     uint32           // HSI state
	HSITrim uint32           // (RevY:0-63, RevB+:0-127) calibration trim
	LSI     uint32           // LSI state
	HSI48   uint32           // HSI48 state
	CSI     uint32           // CSI state
	CSITrim uint32           // (RevY:0-31, RevB+:0-63) calibration trim
	PLL1    RCC_PLL_CFG_Type // PLL structure parameters
}

// RCC_CLK_CFG_Type defines an RCC SYS/AHB/APB bus clock configuration.
type RCC_CLK_CFG_Type struct {
	CLK     uint32 // clock to be configured
	SYSSrc  uint32 // system clock (SYSCLK) source
	SYSDiv  uint32 // system clock (SYSCLK) divider
	AHBDiv  uint32 // AHB clock (HCLK) divider (derived from SYSCLK)
	APB3Div uint32 // APB3 clock (D1PCLK1) divider (derived from HCLK)
	APB1Div uint32 // APB1 clock (PCLK1) divider (derived from HCLK)
	APB2Div uint32 // APB2 clock (PCLK2) divider (derived from HCLK)
	APB4Div uint32 // APB4 clock (D3PCLK1) divider (derived from HCLK)
}

// RCC_CORE_Type
type RCC_CORE_Type struct {
	RSR        volatile.Register32 // RCC Reset status register                            Address offset: 0x00
	AHB3ENR    volatile.Register32 // RCC AHB3 peripheral clock  register                  Address offset: 0x04
	AHB1ENR    volatile.Register32 // RCC AHB1 peripheral clock  register                  Address offset: 0x08
	AHB2ENR    volatile.Register32 // RCC AHB2 peripheral clock  register                  Address offset: 0x0C
	AHB4ENR    volatile.Register32 // RCC AHB4 peripheral clock  register                  Address offset: 0x10
	APB3ENR    volatile.Register32 // RCC APB3 peripheral clock  register                  Address offset: 0x14
	APB1LENR   volatile.Register32 // RCC APB1 peripheral clock  Low Word register         Address offset: 0x18
	APB1HENR   volatile.Register32 // RCC APB1 peripheral clock  High Word register        Address offset: 0x1C
	APB2ENR    volatile.Register32 // RCC APB2 peripheral clock  register                  Address offset: 0x20
	APB4ENR    volatile.Register32 // RCC APB4 peripheral clock  register                  Address offset: 0x24
	_          [4]byte             // Reserved                                             Address offset: 0x28
	AHB3LPENR  volatile.Register32 // RCC AHB3 peripheral sleep clock  register            Address offset: 0x3C
	AHB1LPENR  volatile.Register32 // RCC AHB1 peripheral sleep clock  register            Address offset: 0x40
	AHB2LPENR  volatile.Register32 // RCC AHB2 peripheral sleep clock  register            Address offset: 0x44
	AHB4LPENR  volatile.Register32 // RCC AHB4 peripheral sleep clock  register            Address offset: 0x48
	APB3LPENR  volatile.Register32 // RCC APB3 peripheral sleep clock  register            Address offset: 0x4C
	APB1LLPENR volatile.Register32 // RCC APB1 peripheral sleep clock  Low Word register   Address offset: 0x50
	APB1HLPENR volatile.Register32 // RCC APB1 peripheral sleep clock  High Word register  Address offset: 0x54
	APB2LPENR  volatile.Register32 // RCC APB2 peripheral sleep clock  register            Address offset: 0x58
	APB4LPENR  volatile.Register32 // RCC APB4 peripheral sleep clock  register            Address offset: 0x5C
	_          [16]byte            // Reserved, 0x60-0x6C                                  Address offset: 0x60
}

// RCC_FLAG_Type holds a bitmap referring to a specific bitfield in one of
// several different registers. It and its only method Get() are provided for
// convenience.
type RCC_FLAG_Type uint8

const (
	RCC_FLAG_MASK     RCC_FLAG_Type = 0x1F
	RCC_FLAG_HSIRDY   RCC_FLAG_Type = 0x22
	RCC_FLAG_HSIDIV   RCC_FLAG_Type = 0x25
	RCC_FLAG_CSIRDY   RCC_FLAG_Type = 0x28
	RCC_FLAG_HSI48RDY RCC_FLAG_Type = 0x2D
	RCC_FLAG_D1CKRDY  RCC_FLAG_Type = 0x2E
	RCC_FLAG_CPUCKRDY RCC_FLAG_Type = 0x2E
	RCC_FLAG_D2CKRDY  RCC_FLAG_Type = 0x2F
	RCC_FLAG_CDCKRDY  RCC_FLAG_Type = 0x2F
	RCC_FLAG_HSERDY   RCC_FLAG_Type = 0x31
	RCC_FLAG_PLLRDY   RCC_FLAG_Type = 0x39
	RCC_FLAG_PLL2RDY  RCC_FLAG_Type = 0x3B
	RCC_FLAG_PLL3RDY  RCC_FLAG_Type = 0x3D
	RCC_FLAG_LSERDY   RCC_FLAG_Type = 0x41
	RCC_FLAG_LSIRDY   RCC_FLAG_Type = 0x61
	RCC_FLAG_CPURST   RCC_FLAG_Type = 0x91
	RCC_FLAG_D1RST    RCC_FLAG_Type = 0x93
	RCC_FLAG_CDRST    RCC_FLAG_Type = 0x93
	RCC_FLAG_D2RST    RCC_FLAG_Type = 0x94
	RCC_FLAG_BORRST   RCC_FLAG_Type = 0x95
	RCC_FLAG_PINRST   RCC_FLAG_Type = 0x96
	RCC_FLAG_PORRST   RCC_FLAG_Type = 0x97
	RCC_FLAG_SFTRST   RCC_FLAG_Type = 0x98
	RCC_FLAG_IWDG1RST RCC_FLAG_Type = 0x9A
	RCC_FLAG_WWDG1RST RCC_FLAG_Type = 0x9C
	RCC_FLAG_LPWR1RST RCC_FLAG_Type = 0x9E
	RCC_FLAG_LPWR2RST RCC_FLAG_Type = 0x9F
	RCC_FLAG_C1RST                  = RCC_FLAG_CPURST
	RCC_FLAG_C2RST    RCC_FLAG_Type = 0x92
	RCC_FLAG_SFTR1ST                = RCC_FLAG_SFTRST
	RCC_FLAG_SFTR2ST  RCC_FLAG_Type = 0x99
	RCC_FLAG_WWDG2RST RCC_FLAG_Type = 0x9D
	RCC_FLAG_IWDG2RST RCC_FLAG_Type = 0x9B
)

func (f RCC_FLAG_Type) Get() bool {
	// Derived from the following obnoxious C macro (fully-credited to ST):
	//	 (((((((__FLAG__) >> 5U) == 1U) ?
	//	 	RCC->CR : ((((__FLAG__) >> 5U) == 2U) ?
	//	 		RCC->BDCR : ((((__FLAG__) >> 5U) == 3U) ?
	//	 			RCC->CSR : ((((__FLAG__) >> 5U) == 4U) ?
	//	 				RCC->RSR : RCC->CIFR)))) &
	//	 	(1U << ((__FLAG__) & RCC_FLAG_MASK))) != 0U) ? 1U : 0U)
	switch f >> 5 {
	case 1:
		return RCC.CR.HasBits(1 << (f & RCC_FLAG_MASK))
	case 2:
		return RCC.BDCR.HasBits(1 << (f & RCC_FLAG_MASK))
	case 3:
		return RCC.CSR.HasBits(1 << (f & RCC_FLAG_MASK))
	case 4:
		return RCC.RSR.HasBits(1 << (f & RCC_FLAG_MASK))
	default:
		return RCC.CIFR.HasBits(1 << (f & RCC_FLAG_MASK))
	}
}

// Valused used as CLK identifiers in RCC_CLK_CONFIG_Type
const (
	RCC_CLK_SYSCLK = 1 << iota
	RCC_CLK_HCLK
	RCC_CLK_D1PCLK1
	RCC_CLK_PCLK1
	RCC_CLK_PCLK2
	RCC_CLK_D3PCLK1
)

// Values used as OSC identifiers in RCC_OSC_CONFIG_Type
const (
	RCC_OSC_NONE = 1 << iota
	RCC_OSC_HSE
	RCC_OSC_HSI
	RCC_OSC_LSE
	RCC_OSC_LSI
	RCC_OSC_CSI
	RCC_OSC_HSI48
)

// Values used in RCC CFGR fields SW & SWS
const (
	RCC_SYS_SRC_HSI = iota
	RCC_SYS_SRC_CSI
	RCC_SYS_SRC_HSE
	RCC_SYS_SRC_PLL1
)

// Values used in RCC PLLCKSELR field PLLSRC
const (
	RCC_PLL_SRC_HSI = iota
	RCC_PLL_SRC_CSI
	RCC_PLL_SRC_HSE
	RCC_PLL_SRC_NONE
)

const RCC_CR_OSCOFF = 0

const (
	RCC_PLL_NONE = 0
	RCC_PLL_OFF  = 1
	RCC_PLL_ON   = 2
)

// System core frequency prescaler table
var rccCorePrescaler = [...]uint8{
	0, 0, 0, 0, 1, 2, 3, 4, 1, 2, 3, 4, 6, 7, 8, 9,
}

const (
	rccSYSTimeoutMs = 5000 //   5  s
	rccHSETimeoutMs = 5000 //   5  s
	rccHSITimeoutMs = 2    //   2 ms
	rccI48TimeoutMs = 2    //   2 ms
	rccCSITimeoutMs = 2    //   2 ms
	rccLSITimeoutMs = 2    //   2 ms
	rccLSETimeoutMs = 5000 //   5  s
	rccPLLTimeoutMs = 2    //   2 ms
	rccDBPTimeoutMs = 100  // 100 ms
)

const (
	RCC_D1CFGR_HPRE_DIV1       = 0                                 // AHB3 Clock not divided
	RCC_D1CFGR_HPRE_DIV2_Pos   = 3                                 // //
	RCC_D1CFGR_HPRE_DIV2_Msk   = 0x1 << RCC_D1CFGR_HPRE_DIV2_Pos   // 0x00000008
	RCC_D1CFGR_HPRE_DIV2       = RCC_D1CFGR_HPRE_DIV2_Msk          // AHB3 Clock divided by 2
	RCC_D1CFGR_HPRE_DIV4_Pos   = 0                                 //
	RCC_D1CFGR_HPRE_DIV4_Msk   = 0x9 << RCC_D1CFGR_HPRE_DIV4_Pos   // 0x00000009
	RCC_D1CFGR_HPRE_DIV4       = RCC_D1CFGR_HPRE_DIV4_Msk          // AHB3 Clock divided by 4
	RCC_D1CFGR_HPRE_DIV8_Pos   = 1                                 //
	RCC_D1CFGR_HPRE_DIV8_Msk   = 0x5 << RCC_D1CFGR_HPRE_DIV8_Pos   // 0x0000000A
	RCC_D1CFGR_HPRE_DIV8       = RCC_D1CFGR_HPRE_DIV8_Msk          // AHB3 Clock divided by 8
	RCC_D1CFGR_HPRE_DIV16_Pos  = 0                                 //
	RCC_D1CFGR_HPRE_DIV16_Msk  = 0xB << RCC_D1CFGR_HPRE_DIV16_Pos  // 0x0000000B
	RCC_D1CFGR_HPRE_DIV16      = RCC_D1CFGR_HPRE_DIV16_Msk         // AHB3 Clock divided by 16
	RCC_D1CFGR_HPRE_DIV64_Pos  = 2                                 //
	RCC_D1CFGR_HPRE_DIV64_Msk  = 0x3 << RCC_D1CFGR_HPRE_DIV64_Pos  // 0x0000000C
	RCC_D1CFGR_HPRE_DIV64      = RCC_D1CFGR_HPRE_DIV64_Msk         // AHB3 Clock divided by 64
	RCC_D1CFGR_HPRE_DIV128_Pos = 0                                 //
	RCC_D1CFGR_HPRE_DIV128_Msk = 0xD << RCC_D1CFGR_HPRE_DIV128_Pos // 0x0000000D
	RCC_D1CFGR_HPRE_DIV128     = RCC_D1CFGR_HPRE_DIV128_Msk        // AHB3 Clock divided by 128
	RCC_D1CFGR_HPRE_DIV256_Pos = 1                                 //
	RCC_D1CFGR_HPRE_DIV256_Msk = 0x7 << RCC_D1CFGR_HPRE_DIV256_Pos // 0x0000000E
	RCC_D1CFGR_HPRE_DIV256     = RCC_D1CFGR_HPRE_DIV256_Msk        // AHB3 Clock divided by 256
	RCC_D1CFGR_HPRE_DIV512_Pos = 0                                 //
	RCC_D1CFGR_HPRE_DIV512_Msk = 0xF << RCC_D1CFGR_HPRE_DIV512_Pos // 0x0000000F
	RCC_D1CFGR_HPRE_DIV512     = RCC_D1CFGR_HPRE_DIV512_Msk        // AHB3 Clock divided by 512

	RCC_D1CFGR_D1PPRE_DIV1      = 0                                  // APB3 clock not divided
	RCC_D1CFGR_D1PPRE_DIV2_Pos  = 6                                  //
	RCC_D1CFGR_D1PPRE_DIV2_Msk  = 0x1 << RCC_D1CFGR_D1PPRE_DIV2_Pos  // 0x00000040
	RCC_D1CFGR_D1PPRE_DIV2      = RCC_D1CFGR_D1PPRE_DIV2_Msk         // APB3 clock divided by 2
	RCC_D1CFGR_D1PPRE_DIV4_Pos  = 4                                  //
	RCC_D1CFGR_D1PPRE_DIV4_Msk  = 0x5 << RCC_D1CFGR_D1PPRE_DIV4_Pos  // 0x00000050
	RCC_D1CFGR_D1PPRE_DIV4      = RCC_D1CFGR_D1PPRE_DIV4_Msk         // APB3 clock divided by 4
	RCC_D1CFGR_D1PPRE_DIV8_Pos  = 5                                  //
	RCC_D1CFGR_D1PPRE_DIV8_Msk  = 0x3 << RCC_D1CFGR_D1PPRE_DIV8_Pos  // 0x00000060
	RCC_D1CFGR_D1PPRE_DIV8      = RCC_D1CFGR_D1PPRE_DIV8_Msk         // APB3 clock divided by 8
	RCC_D1CFGR_D1PPRE_DIV16_Pos = 4                                  //
	RCC_D1CFGR_D1PPRE_DIV16_Msk = 0x7 << RCC_D1CFGR_D1PPRE_DIV16_Pos // 0x00000070
	RCC_D1CFGR_D1PPRE_DIV16     = RCC_D1CFGR_D1PPRE_DIV16_Msk        // APB3 clock divided by 16

	RCC_D1CFGR_D1CPRE_DIV1       = 0                                   // Domain 1 Core clock not divided
	RCC_D1CFGR_D1CPRE_DIV2_Pos   = 11                                  //
	RCC_D1CFGR_D1CPRE_DIV2_Msk   = 0x1 << RCC_D1CFGR_D1CPRE_DIV2_Pos   // 0x00000800
	RCC_D1CFGR_D1CPRE_DIV2       = RCC_D1CFGR_D1CPRE_DIV2_Msk          // Domain 1 Core clock divided by 2
	RCC_D1CFGR_D1CPRE_DIV4_Pos   = 8                                   //
	RCC_D1CFGR_D1CPRE_DIV4_Msk   = 0x9 << RCC_D1CFGR_D1CPRE_DIV4_Pos   // 0x00000900
	RCC_D1CFGR_D1CPRE_DIV4       = RCC_D1CFGR_D1CPRE_DIV4_Msk          // Domain 1 Core clock divided by 4
	RCC_D1CFGR_D1CPRE_DIV8_Pos   = 9                                   //
	RCC_D1CFGR_D1CPRE_DIV8_Msk   = 0x5 << RCC_D1CFGR_D1CPRE_DIV8_Pos   // 0x00000A00
	RCC_D1CFGR_D1CPRE_DIV8       = RCC_D1CFGR_D1CPRE_DIV8_Msk          // Domain 1 Core clock divided by 8
	RCC_D1CFGR_D1CPRE_DIV16_Pos  = 8                                   //
	RCC_D1CFGR_D1CPRE_DIV16_Msk  = 0xB << RCC_D1CFGR_D1CPRE_DIV16_Pos  // 0x00000B00
	RCC_D1CFGR_D1CPRE_DIV16      = RCC_D1CFGR_D1CPRE_DIV16_Msk         // Domain 1 Core clock divided by 16
	RCC_D1CFGR_D1CPRE_DIV64_Pos  = 10                                  //
	RCC_D1CFGR_D1CPRE_DIV64_Msk  = 0x3 << RCC_D1CFGR_D1CPRE_DIV64_Pos  // 0x00000C00
	RCC_D1CFGR_D1CPRE_DIV64      = RCC_D1CFGR_D1CPRE_DIV64_Msk         // Domain 1 Core clock divided by 64
	RCC_D1CFGR_D1CPRE_DIV128_Pos = 8                                   //
	RCC_D1CFGR_D1CPRE_DIV128_Msk = 0xD << RCC_D1CFGR_D1CPRE_DIV128_Pos // 0x00000D00
	RCC_D1CFGR_D1CPRE_DIV128     = RCC_D1CFGR_D1CPRE_DIV128_Msk        // Domain 1 Core clock divided by 128
	RCC_D1CFGR_D1CPRE_DIV256_Pos = 9                                   //
	RCC_D1CFGR_D1CPRE_DIV256_Msk = 0x7 << RCC_D1CFGR_D1CPRE_DIV256_Pos // 0x00000E00
	RCC_D1CFGR_D1CPRE_DIV256     = RCC_D1CFGR_D1CPRE_DIV256_Msk        // Domain 1 Core clock divided by 256
	RCC_D1CFGR_D1CPRE_DIV512_Pos = 8                                   //
	RCC_D1CFGR_D1CPRE_DIV512_Msk = 0xF << RCC_D1CFGR_D1CPRE_DIV512_Pos // 0x00000F00
	RCC_D1CFGR_D1CPRE_DIV512     = RCC_D1CFGR_D1CPRE_DIV512_Msk        // Domain 1 Core clock divided by 512

	RCC_D2CFGR_D2PPRE1_DIV1      = 0                                   // APB1 clock not divided
	RCC_D2CFGR_D2PPRE1_DIV2_Pos  = 6                                   //
	RCC_D2CFGR_D2PPRE1_DIV2_Msk  = 0x1 << RCC_D2CFGR_D2PPRE1_DIV2_Pos  // 0x00000040
	RCC_D2CFGR_D2PPRE1_DIV2      = RCC_D2CFGR_D2PPRE1_DIV2_Msk         // APB1 clock divided by 2
	RCC_D2CFGR_D2PPRE1_DIV4_Pos  = 4                                   //
	RCC_D2CFGR_D2PPRE1_DIV4_Msk  = 0x5 << RCC_D2CFGR_D2PPRE1_DIV4_Pos  // 0x00000050
	RCC_D2CFGR_D2PPRE1_DIV4      = RCC_D2CFGR_D2PPRE1_DIV4_Msk         // APB1 clock divided by 4
	RCC_D2CFGR_D2PPRE1_DIV8_Pos  = 5                                   //
	RCC_D2CFGR_D2PPRE1_DIV8_Msk  = 0x3 << RCC_D2CFGR_D2PPRE1_DIV8_Pos  // 0x00000060
	RCC_D2CFGR_D2PPRE1_DIV8      = RCC_D2CFGR_D2PPRE1_DIV8_Msk         // APB1 clock divided by 8
	RCC_D2CFGR_D2PPRE1_DIV16_Pos = 4                                   //
	RCC_D2CFGR_D2PPRE1_DIV16_Msk = 0x7 << RCC_D2CFGR_D2PPRE1_DIV16_Pos // 0x00000070
	RCC_D2CFGR_D2PPRE1_DIV16     = RCC_D2CFGR_D2PPRE1_DIV16_Msk        // APB1 clock divided by 16

	RCC_D2CFGR_D2PPRE2_DIV1      = 0                                   // APB2 clock not divided
	RCC_D2CFGR_D2PPRE2_DIV2_Pos  = 10                                  //
	RCC_D2CFGR_D2PPRE2_DIV2_Msk  = 0x1 << RCC_D2CFGR_D2PPRE2_DIV2_Pos  // 0x00000400
	RCC_D2CFGR_D2PPRE2_DIV2      = RCC_D2CFGR_D2PPRE2_DIV2_Msk         // APB2 clock divided by 2
	RCC_D2CFGR_D2PPRE2_DIV4_Pos  = 8                                   //
	RCC_D2CFGR_D2PPRE2_DIV4_Msk  = 0x5 << RCC_D2CFGR_D2PPRE2_DIV4_Pos  // 0x00000500
	RCC_D2CFGR_D2PPRE2_DIV4      = RCC_D2CFGR_D2PPRE2_DIV4_Msk         // APB2 clock divided by 4
	RCC_D2CFGR_D2PPRE2_DIV8_Pos  = 9                                   //
	RCC_D2CFGR_D2PPRE2_DIV8_Msk  = 0x3 << RCC_D2CFGR_D2PPRE2_DIV8_Pos  // 0x00000600
	RCC_D2CFGR_D2PPRE2_DIV8      = RCC_D2CFGR_D2PPRE2_DIV8_Msk         // APB2 clock divided by 8
	RCC_D2CFGR_D2PPRE2_DIV16_Pos = 8                                   //
	RCC_D2CFGR_D2PPRE2_DIV16_Msk = 0x7 << RCC_D2CFGR_D2PPRE2_DIV16_Pos // 0x00000700
	RCC_D2CFGR_D2PPRE2_DIV16     = RCC_D2CFGR_D2PPRE2_DIV16_Msk        // APB2 clock divided by 16

	// The auto-generated TinyGo source contains bitmasks for the wrong table
	// (RCC_AHB3ENR), and doesn't have those required for RCC_D3CFGR (which has
	// been recreated from STM32 HAL SDK below). Need to determine if this is a
	// bug with the SVD itself or the parser/generator.

	RCC_D3CFGR_D3PPRE_Pos       = 4                                  //
	RCC_D3CFGR_D3PPRE_Msk       = 0x7 << RCC_D3CFGR_D3PPRE_Pos       // 0x00000070
	RCC_D3CFGR_D3PPRE           = RCC_D3CFGR_D3PPRE_Msk              // D3PPRE1[2:0] bits (APB4 prescaler)
	RCC_D3CFGR_D3PPRE_DIV1      = 0x00000000                         // APB4 clock not divided
	RCC_D3CFGR_D3PPRE_DIV2_Pos  = 6                                  //
	RCC_D3CFGR_D3PPRE_DIV2_Msk  = 0x1 << RCC_D3CFGR_D3PPRE_DIV2_Pos  // 0x00000040
	RCC_D3CFGR_D3PPRE_DIV2      = RCC_D3CFGR_D3PPRE_DIV2_Msk         // APB4 clock divided by 2
	RCC_D3CFGR_D3PPRE_DIV4_Pos  = 4                                  //
	RCC_D3CFGR_D3PPRE_DIV4_Msk  = 0x5 << RCC_D3CFGR_D3PPRE_DIV4_Pos  // 0x00000050
	RCC_D3CFGR_D3PPRE_DIV4      = RCC_D3CFGR_D3PPRE_DIV4_Msk         // APB4 clock divided by 4
	RCC_D3CFGR_D3PPRE_DIV8_Pos  = 5                                  //
	RCC_D3CFGR_D3PPRE_DIV8_Msk  = 0x3 << RCC_D3CFGR_D3PPRE_DIV8_Pos  // 0x00000060
	RCC_D3CFGR_D3PPRE_DIV8      = RCC_D3CFGR_D3PPRE_DIV8_Msk         // APB4 clock divided by 8
	RCC_D3CFGR_D3PPRE_DIV16_Pos = 4                                  //
	RCC_D3CFGR_D3PPRE_DIV16_Msk = 0x7 << RCC_D3CFGR_D3PPRE_DIV16_Pos // 0x00000070
	RCC_D3CFGR_D3PPRE_DIV16     = RCC_D3CFGR_D3PPRE_DIV16_Msk        // APB4 clock divided by 16
)

const (
	RCC_PLL1DIVR_N1_Pos = 0
	RCC_PLL1DIVR_N1_Msk = 0x1FF << RCC_PLL1DIVR_N1_Pos // 0x000001FF
	RCC_PLL1DIVR_N1     = RCC_PLL1DIVR_N1_Msk
	RCC_PLL1DIVR_P1_Pos = 9
	RCC_PLL1DIVR_P1_Msk = 0x7F << RCC_PLL1DIVR_P1_Pos // 0x0000FE00
	RCC_PLL1DIVR_P1     = RCC_PLL1DIVR_P1_Msk
	RCC_PLL1DIVR_Q1_Pos = 16
	RCC_PLL1DIVR_Q1_Msk = 0x7F << RCC_PLL1DIVR_Q1_Pos // 0x007F0000
	RCC_PLL1DIVR_Q1     = RCC_PLL1DIVR_Q1_Msk
	RCC_PLL1DIVR_R1_Pos = 24
	RCC_PLL1DIVR_R1_Msk = 0x7F << RCC_PLL1DIVR_R1_Pos // 0x7F000000
	RCC_PLL1DIVR_R1     = RCC_PLL1DIVR_R1_Msk
)

const (
	RCC_PLL1_VCI_RANGE_0 = 0x0 << RCC_PLLCFGR_PLL1RGE_Pos // 0x00000000: Clock range frequency between 1 and 2 MHz
	RCC_PLL1_VCI_RANGE_1 = 0x1 << RCC_PLLCFGR_PLL1RGE_Pos // 0x00000004: Clock range frequency between 2 and 4 MHz
	RCC_PLL1_VCI_RANGE_2 = 0x2 << RCC_PLLCFGR_PLL1RGE_Pos // 0x00000008: Clock range frequency between 4 and 8 MHz
	RCC_PLL1_VCI_RANGE_3 = 0x3 << RCC_PLLCFGR_PLL1RGE_Pos // 0x0000000C: Clock range frequency between 8 and 16 MHz

	RCC_PLL1_VCO_WIDE   = 0
	RCC_PLL1_VCO_MEDIUM = RCC_PLLCFGR_PLL1VCOSEL
)

const (
	RCC_AHB3ENR_FLASHEN_Pos = 8
	RCC_AHB3ENR_FLASHEN_Msk = 0x1 << RCC_AHB3ENR_FLASHEN_Pos // 0x00000100
	RCC_AHB3ENR_FLASHEN     = RCC_AHB3ENR_FLASHEN_Msk
)

const (
	RCC_CSITRIM = 0x20
	RCC_HSITRIM = 0x40

	RCC_HSITRIM_REV_Y_Pos = 12
	RCC_HSITRIM_REV_Y_Msk = 0x3F000
	RCC_CSITRIM_REV_Y_Pos = 26
	RCC_CSITRIM_REV_Y_Msk = 0x7C000000
)

// Set configures the system oscillators using the receiver oscillator
// configuration struct.
func (osc RCC_OSC_CFG_Type) Set() bool {

	sysSrc, pllSrc :=
		(RCC.CFGR.Get()&RCC_CFGR_SWS_Msk)>>RCC_CFGR_SWS_Pos,
		(RCC.PLLCKSELR.Get()&RCC_PLLCKSELR_PLLSRC_Msk)>>RCC_PLLCKSELR_PLLSRC_Pos

	// Configure HSE
	if osc.OSC&RCC_OSC_HSE == RCC_OSC_HSE {
		if (RCC_SYS_SRC_HSE == sysSrc) ||
			(RCC_SYS_SRC_PLL1 == sysSrc && RCC_PLL_SRC_HSE == pllSrc) {
			// HSE is currently source of SYSCLK.
			if RCC_FLAG_HSERDY.Get() && osc.HSE == RCC_CR_OSCOFF {
				return false // cannot disable HSE while driving SYSCLK
			}
		} else {
			// HSE is currently NOT source of SYSCLK.
			// Configure HSE according to its state given in configuration struct.
			switch osc.HSE {
			case RCC_CR_HSEON:
				RCC.CR.SetBits(RCC_CR_HSEON)
			case RCC_CR_OSCOFF:
				RCC.CR.ClearBits(RCC_CR_HSEON)
				RCC.CR.ClearBits(RCC_CR_HSEBYP)
			case RCC_CR_HSEBYP | RCC_CR_HSEON:
				RCC.CR.SetBits(RCC_CR_HSEBYP)
				RCC.CR.SetBits(RCC_CR_HSEON)
			default:
				RCC.CR.ClearBits(RCC_CR_HSEON)
				RCC.CR.ClearBits(RCC_CR_HSEBYP)
			}
			// Wait for new oscillator configuration to take effect within a fixed
			// window of time; if time expires, return early with error condition.
			if osc.HSE != RCC_CR_OSCOFF {
				start := ticks()
				for !RCC_FLAG_HSERDY.Get() {
					if ticks()-start > rccHSETimeoutMs {
						return false
					}
				}
			} else {
				start := ticks()
				for RCC_FLAG_HSERDY.Get() {
					if ticks()-start > rccHSETimeoutMs {
						return false
					}
				}
			}
		}
	}
	// Configure HSI
	if osc.OSC&RCC_OSC_HSI == RCC_OSC_HSI {
		if (RCC_SYS_SRC_HSI == sysSrc) ||
			(RCC_SYS_SRC_PLL1 == sysSrc && RCC_PLL_SRC_HSI == pllSrc) {
			// HSI is currently source of SYSCLK
			if RCC_FLAG_HSIRDY.Get() && osc.HSI == RCC_CR_OSCOFF {
				return false // cannot disable HSI while driving SYSCLK
			} else {
				// apply HSI calibration factor
				if MCURevision() <= DBGMCU_MCU_REVISION_Y {
					if osc.HSITrim == RCC_HSITRIM {
						RCC.HSICFGR.ReplaceBits(
							0x20<<RCC_HSITRIM_REV_Y_Pos,
							RCC_HSITRIM_REV_Y_Msk, 0)
					} else {
						RCC.HSICFGR.ReplaceBits(
							osc.HSITrim<<RCC_HSITRIM_REV_Y_Pos,
							RCC_HSITRIM_REV_Y_Msk, 0)
					}
				} else {
					RCC.HSICFGR.ReplaceBits(
						osc.HSITrim<<RCC_HSICFGR_HSITRIM_Pos,
						RCC_HSICFGR_HSITRIM_Msk, 0)
				}
			}
		} else {
			// HSI is currently NOT source of SYSCLK.
			// Configure HSI according to its state given in configuration struct.
			if osc.HSI != RCC_CR_OSCOFF {
				// Enable HSI
				RCC.CR.ReplaceBits(osc.HSI, RCC_CR_HSION_Msk|RCC_CR_HSIDIV_Msk, 0)
				// Verify HSI was enabled
				start := ticks()
				for !RCC_FLAG_HSIRDY.Get() {
					if ticks()-start > rccHSITimeoutMs {
						return false
					}
				}
				// apply HSI calibration factor
				if MCURevision() <= DBGMCU_MCU_REVISION_Y {
					if osc.HSITrim == RCC_HSITRIM {
						RCC.HSICFGR.ReplaceBits(
							0x20<<RCC_HSITRIM_REV_Y_Pos,
							RCC_HSITRIM_REV_Y_Msk, 0)
					} else {
						RCC.HSICFGR.ReplaceBits(
							osc.HSITrim<<RCC_HSITRIM_REV_Y_Pos,
							RCC_HSITRIM_REV_Y_Msk, 0)
					}
				} else {
					RCC.HSICFGR.ReplaceBits(
						osc.HSITrim<<RCC_HSICFGR_HSITRIM_Pos,
						RCC_HSICFGR_HSITRIM_Msk, 0)
				}
			} else {
				// Disable HSI
				RCC.CR.ClearBits(RCC_CR_HSION)
				// Verify HSI was disabled
				start := ticks()
				for RCC_FLAG_HSIRDY.Get() {
					if ticks()-start > rccHSITimeoutMs {
						return false
					}
				}
			}
		}
	}
	// Configure CSI
	if osc.OSC&RCC_OSC_CSI == RCC_OSC_CSI {
		if (RCC_SYS_SRC_CSI == sysSrc) ||
			(RCC_SYS_SRC_PLL1 == sysSrc && RCC_PLL_SRC_CSI == pllSrc) {
			// CSI is currently source of SYSCLK
			if RCC_FLAG_CSIRDY.Get() && osc.CSI == RCC_CR_OSCOFF {
				return false // cannot disable HSI while driving SYSCLK
			} else {
				// apply CSI calibration factor
				if MCURevision() <= DBGMCU_MCU_REVISION_Y {
					if osc.CSITrim == RCC_CSITRIM {
						RCC.CSICFGR.ReplaceBits(
							0x10<<RCC_CSITRIM_REV_Y_Pos,
							RCC_CSITRIM_REV_Y_Msk, 0)
					} else {
						RCC.CSICFGR.ReplaceBits(
							osc.CSITrim<<RCC_CSITRIM_REV_Y_Pos,
							RCC_CSITRIM_REV_Y_Msk, 0)
					}
				} else {
					RCC.CSICFGR.ReplaceBits(
						osc.CSITrim<<RCC_CSICFGR_CSITRIM_Pos,
						RCC_CSICFGR_CSITRIM_Msk, 0)
				}
			}
		} else {
			// CSI is currently NOT source of SYSCLK.
			// Configure CSI according to its state given in configuration struct.
			if osc.CSI != RCC_CR_OSCOFF {
				// Enable CSI
				RCC.CR.SetBits(RCC_CR_CSION)
				// Verify CSI was enabled
				start := ticks()
				for !RCC_FLAG_CSIRDY.Get() {
					if ticks()-start > rccCSITimeoutMs {
						return false
					}
				}
				// apply CSI calibration factor
				if MCURevision() <= DBGMCU_MCU_REVISION_Y {
					if osc.CSITrim == RCC_CSITRIM {
						RCC.CSICFGR.ReplaceBits(
							0x10<<RCC_CSITRIM_REV_Y_Pos,
							RCC_CSITRIM_REV_Y_Msk, 0)
					} else {
						RCC.CSICFGR.ReplaceBits(
							osc.CSITrim<<RCC_CSITRIM_REV_Y_Pos,
							RCC_CSITRIM_REV_Y_Msk, 0)
					}
				} else {
					RCC.CSICFGR.ReplaceBits(
						osc.CSITrim<<RCC_CSICFGR_CSITRIM_Pos,
						RCC_CSICFGR_CSITRIM_Msk, 0)
				}
			} else {
				// Disable CSI
				RCC.CR.ClearBits(RCC_CR_CSION)
				// Verify CSI was disabled
				start := ticks()
				for RCC_FLAG_CSIRDY.Get() {
					if ticks()-start > rccCSITimeoutMs {
						return false
					}
				}
			}
		}
	}
	// Configure LSI
	if osc.OSC&RCC_OSC_LSI == RCC_OSC_LSI {
		if osc.LSI != RCC_CR_OSCOFF {
			// Enable LSI
			RCC.CSR.SetBits(RCC_CSR_LSION)
			// Verify LSI was enabled
			start := ticks()
			for !RCC_FLAG_LSIRDY.Get() {
				if ticks()-start > rccLSITimeoutMs {
					return false
				}
			}
		} else {
			// Disable LSI
			RCC.CSR.ClearBits(RCC_CSR_LSION)
			// Verify LSI was disabled
			start := ticks()
			for RCC_FLAG_LSIRDY.Get() {
				if ticks()-start > rccLSITimeoutMs {
					return false
				}
			}
		}
	}
	// Configure HSI48
	if osc.OSC&RCC_OSC_HSI48 == RCC_OSC_HSI48 {
		if osc.HSI48 != RCC_CR_OSCOFF {
			// Enable HSI48
			RCC.CR.SetBits(RCC_CR_RC48ON)
			// Verify HSI48 was enabled
			start := ticks()
			for !RCC_FLAG_HSI48RDY.Get() {
				if ticks()-start > rccI48TimeoutMs {
					return false
				}
			}
		} else {
			// Disable HSI48
			RCC.CR.ClearBits(RCC_CR_RC48ON)
			// Verify HSI48 was disabled
			start := ticks()
			for RCC_FLAG_HSI48RDY.Get() {
				if ticks()-start > rccI48TimeoutMs {
					return false
				}
			}
		}
	}
	// Configure LSE
	if osc.OSC&RCC_OSC_LSE == RCC_OSC_LSE {
		// ensure write access to backup domain
		PWR.PWR_CR1.SetBits(PWR_PWR_CR1_DBP)
		// bail out with error condition if write protection can't be removed.
		start := ticks()
		for !PWR.PWR_CR1.HasBits(PWR_PWR_CR1_DBP) {
			if ticks()-start > rccDBPTimeoutMs {
				return false
			}
		}
		// Configure LSE according to its state given in configuration struct.
		switch osc.LSE {
		case RCC_BDCR_LSEON:
			RCC.BDCR.SetBits(RCC_BDCR_LSEON)
		case RCC_CR_OSCOFF:
			RCC.BDCR.ClearBits(RCC_BDCR_LSEON)
			RCC.BDCR.ClearBits(RCC_BDCR_LSEBYP)
		case RCC_BDCR_LSEBYP | RCC_BDCR_LSEON:
			RCC.BDCR.SetBits(RCC_BDCR_LSEBYP)
			RCC.BDCR.SetBits(RCC_BDCR_LSEON)
		default:
			RCC.BDCR.ClearBits(RCC_BDCR_LSEON)
			RCC.BDCR.ClearBits(RCC_BDCR_LSEBYP)
		}
		// Wait for new oscillator configuration to take effect within a fixed
		// window of time; if time expires, return early with error condition.
		if osc.LSE != RCC_CR_OSCOFF {
			start := ticks()
			for !RCC_FLAG_LSERDY.Get() {
				if ticks()-start > rccLSETimeoutMs {
					return false
				}
			}
		} else {
			// Verify LSE was disabled
			start := ticks()
			for RCC_FLAG_LSERDY.Get() {
				if ticks()-start > rccLSETimeoutMs {
					return false
				}
			}
		}
	}
	// Configure PLL
	if osc.PLL1.PLL != RCC_PLL_NONE {
		// Check if PLL is currently being used as source of SYSCLK.
		if sysSrc != RCC_SYS_SRC_PLL1 {
			// PLL is currently NOT source of SYSCLK.
			// Check if we are requesting to enabled or disable PLL.
			if osc.PLL1.PLL == RCC_PLL_ON {
				// Enabling PLL ...
				// First, disable PLL prior to reconfiguring
				RCC.CR.ClearBits(RCC_CR_PLL1ON)
				// Wait for PLL to disable
				start := ticks()
				for RCC_FLAG_PLLRDY.Get() {
					if ticks()-start > rccPLLTimeoutMs {
						return false
					}
				}

				// Configure PLL clock source
				RCC.PLLCKSELR.ReplaceBits(
					(osc.PLL1.Src<<RCC_PLLCKSELR_PLLSRC_Pos)|
						(osc.PLL1.Div.M<<RCC_PLLCKSELR_DIVM1_Pos),
					(RCC_PLLCKSELR_PLLSRC_Msk | RCC_PLLCKSELR_DIVM1_Msk), 0)

				// Configure PLL multiplication and division factors
				n := ((osc.PLL1.Div.N - 1) << RCC_PLL1DIVR_DIVN1_Pos) & RCC_PLL1DIVR_DIVN1_Msk
				p := ((osc.PLL1.Div.PQR.P - 1) << RCC_PLL1DIVR_DIVP1_Pos) & RCC_PLL1DIVR_DIVP1_Msk
				q := ((osc.PLL1.Div.PQR.Q - 1) << RCC_PLL1DIVR_DIVQ1_Pos) & RCC_PLL1DIVR_DIVQ1_Msk
				r := ((osc.PLL1.Div.PQR.R - 1) << RCC_PLL1DIVR_DIVR1_Pos) & RCC_PLL1DIVR_DIVR1_Msk

				RCC.PLL1DIVR.Set(n | p | q | r)

				RCC.PLLCFGR.ClearBits(RCC_PLLCFGR_PLL1FRACEN)

				RCC.PLL1FRACR.ReplaceBits(
					osc.PLL1.Div.Nf<<RCC_PLL1FRACR_FRACN1_Pos,
					RCC_PLL1FRACR_FRACN1_Msk, 0)

				RCC.PLLCFGR.ReplaceBits(
					osc.PLL1.Div.VCI,
					RCC_PLLCFGR_PLL1RGE_Msk, 0)

				RCC.PLLCFGR.ReplaceBits(
					osc.PLL1.Div.VCO,
					RCC_PLLCFGR_PLL1VCOSEL_Msk, 0)

				RCC.PLLCFGR.SetBits((RCC_PLLCFGR_DIVP1EN))
				RCC.PLLCFGR.SetBits((RCC_PLLCFGR_DIVQ1EN))
				RCC.PLLCFGR.SetBits((RCC_PLLCFGR_DIVR1EN))
				RCC.PLLCFGR.SetBits(RCC_PLLCFGR_PLL1FRACEN)
				RCC.CR.SetBits(RCC_CR_PLL1ON)
				start = ticks()
				for !RCC_FLAG_PLLRDY.Get() {
					if ticks()-start > rccPLLTimeoutMs {
						return false
					}
				}

			} else {
				// Disabling PLL ...
				// Clear PLL enable bits
				RCC.CR.ClearBits(RCC_CR_PLL1ON)
				// Verify PLL was disabled
				start := ticks()
				for RCC_FLAG_PLLRDY.Get() {
					if ticks()-start > rccPLLTimeoutMs {
						return false
					}
				}
			}
		} else {
			// PLL is currently source of SYSCLK
		}
	}
	return true
}

// Set configures the CPU, AHB, and APB bus clocks using the receiver clock
// configuration struct.
func (clk RCC_CLK_CFG_Type) Set(latency uint32) bool {

	if latency > FLASH.ACR.Get()&FLASH_ACR_LATENCY_Msk {
		FLASH.ACR.ReplaceBits(latency, FLASH_ACR_LATENCY_Msk, 0)
		if latency != FLASH.ACR.Get()&FLASH_ACR_LATENCY_Msk {
			return false
		}
	}
	if clk.CLK&RCC_CLK_D1PCLK1 == RCC_CLK_D1PCLK1 {
		if clk.APB3Div > RCC.D1CFGR.Get()&RCC_D1CFGR_D1PPRE_Msk {
			RCC.D1CFGR.ReplaceBits(clk.APB3Div, RCC_D1CFGR_D1PPRE_Msk, 0)
		}
	}
	if clk.CLK&RCC_CLK_PCLK1 == RCC_CLK_PCLK1 {
		if clk.APB1Div > RCC.D2CFGR.Get()&RCC_D2CFGR_D2PPRE1_Msk {
			RCC.D2CFGR.ReplaceBits(clk.APB1Div, RCC_D2CFGR_D2PPRE1_Msk, 0)
		}
	}
	if clk.CLK&RCC_CLK_PCLK2 == RCC_CLK_PCLK2 {
		if clk.APB2Div > RCC.D2CFGR.Get()&RCC_D2CFGR_D2PPRE2_Msk {
			RCC.D2CFGR.ReplaceBits(clk.APB2Div, RCC_D2CFGR_D2PPRE2_Msk, 0)
		}
	}
	if clk.CLK&RCC_CLK_D3PCLK1 == RCC_CLK_D3PCLK1 {
		// See notes below regarding TinyGo source and MCU table "RCC_D3CFGR".
		if clk.APB4Div > RCC.D3CFGR.Get()&RCC_D3CFGR_D3PPRE {
			RCC.D3CFGR.ReplaceBits(clk.APB4Div, RCC_D3CFGR_D3PPRE, 0)
		}
	}
	if clk.CLK&RCC_CLK_HCLK == RCC_CLK_HCLK {
		if clk.AHBDiv > RCC.D1CFGR.Get()&RCC_D1CFGR_HPRE_Msk {
			RCC.D1CFGR.ReplaceBits(clk.AHBDiv, RCC_D1CFGR_HPRE_Msk, 0)
		}
	}
	if clk.CLK&RCC_CLK_SYSCLK == RCC_CLK_SYSCLK {
		RCC.D1CFGR.ReplaceBits(clk.SYSDiv, RCC_D1CFGR_D1CPRE_Msk, 0)
		var flag RCC_FLAG_Type
		switch clk.SYSSrc {
		case RCC_SYS_SRC_HSI:
			flag = RCC_FLAG_HSIRDY
		case RCC_SYS_SRC_CSI:
			flag = RCC_FLAG_CSIRDY
		case RCC_SYS_SRC_HSE:
			flag = RCC_FLAG_HSERDY
		case RCC_SYS_SRC_PLL1:
			flag = RCC_FLAG_HSERDY
		default:
			return false
		}
		if !flag.Get() {
			return false
		}
		RCC.CFGR.ReplaceBits(clk.SYSSrc, RCC_CFGR_SW_Msk, 0)
		start := ticks()
		for (RCC.CFGR.Get()&RCC_CFGR_SWS_Msk)>>RCC_CFGR_SWS_Pos != clk.SYSSrc {
			if ticks()-start > rccSYSTimeoutMs {
				return false
			}
		}
	}
	if clk.CLK&RCC_CLK_HCLK == RCC_CLK_HCLK {
		if clk.AHBDiv < RCC.D1CFGR.Get()&RCC_D1CFGR_HPRE_Msk {
			RCC.D1CFGR.ReplaceBits(clk.AHBDiv, RCC_D1CFGR_HPRE_Msk, 0)
		}
	}
	if latency < FLASH.ACR.Get()&FLASH_ACR_LATENCY_Msk {
		FLASH.ACR.ReplaceBits(latency, FLASH_ACR_LATENCY_Msk, 0)
		if latency != FLASH.ACR.Get()&FLASH_ACR_LATENCY_Msk {
			return false
		}
	}
	if clk.CLK&RCC_CLK_D1PCLK1 == RCC_CLK_D1PCLK1 {
		if clk.APB3Div < RCC.D1CFGR.Get()&RCC_D1CFGR_D1PPRE_Msk {
			RCC.D1CFGR.ReplaceBits(clk.APB3Div, RCC_D1CFGR_D1PPRE_Msk, 0)
		}
	}
	if clk.CLK&RCC_CLK_PCLK1 == RCC_CLK_PCLK1 {
		if clk.APB1Div < RCC.D2CFGR.Get()&RCC_D2CFGR_D2PPRE1_Msk {
			RCC.D2CFGR.ReplaceBits(clk.APB1Div, RCC_D2CFGR_D2PPRE1_Msk, 0)
		}
	}
	if clk.CLK&RCC_CLK_PCLK2 == RCC_CLK_PCLK2 {
		if clk.APB2Div < RCC.D2CFGR.Get()&RCC_D2CFGR_D2PPRE2_Msk {
			RCC.D2CFGR.ReplaceBits(clk.APB2Div, RCC_D2CFGR_D2PPRE2_Msk, 0)
		}
	}
	if clk.CLK&RCC_CLK_D3PCLK1 == RCC_CLK_D3PCLK1 {
		// See notes below regarding TinyGo source and MCU table "RCC_D3CFGR".
		if clk.APB4Div < RCC.D3CFGR.Get()&RCC_D3CFGR_D3PPRE {
			RCC.D3CFGR.ReplaceBits(clk.APB4Div, RCC_D3CFGR_D3PPRE, 0)
		}
	}

	return true
}

func sysFreq() uint32 {
	switch (RCC.CFGR.Get() & RCC_CFGR_SWS_Msk) >> RCC_CFGR_SWS_Pos {
	case RCC_SYS_SRC_HSI:
		div := (RCC.CR.Get() & RCC_CR_HSIDIV_Msk) >> RCC_CR_HSIDIV_Pos
		return RCC_HSI_FREQ_HZ >> div
	case RCC_SYS_SRC_CSI:
		return RCC_CSI_FREQ_HZ
	case RCC_SYS_SRC_HSE:
		return RCC_HSE_FREQ_HZ
	case RCC_SYS_SRC_PLL1:
		clk := pll1Freq()
		return clk.P
	}
	return 0
}

func pll1Freq() (clk RCC_PLL_PQR_Type) {
	var in uint32
	switch (RCC.PLLCKSELR.Get() & RCC_PLLCKSELR_PLLSRC_Msk) >>
		RCC_PLLCKSELR_PLLSRC_Pos {
	case RCC_PLL_SRC_HSI:
		if RCC.CR.HasBits(RCC_CR_HSIRDY) {
			div := (RCC.CR.Get() & RCC_CR_HSIDIV_Msk) >> RCC_CR_HSIDIV_Pos
			in = RCC_HSI_FREQ_HZ >> div
		}
	case RCC_PLL_SRC_CSI:
		if RCC.CR.HasBits(RCC_CR_CSIRDY) {
			in = RCC_CSI_FREQ_HZ
		}
	case RCC_PLL_SRC_HSE:
		if RCC.CR.HasBits(RCC_CR_HSERDY) {
			in = RCC_HSE_FREQ_HZ
		}
	case RCC_PLL_SRC_NONE:
		// PLL disabled
	}
	clk.P = 0
	clk.Q = 0
	clk.R = 0
	m := (RCC.PLLCKSELR.Get() & RCC_PLLCKSELR_DIVM1_Msk) >>
		RCC_PLLCKSELR_DIVM1_Pos
	n := ((RCC.PLL1DIVR.Get() & RCC_PLL1DIVR_DIVN1_Msk) >>
		RCC_PLL1DIVR_DIVN1_Pos) + 1
	var fracn uint32
	if RCC.PLLCFGR.HasBits(RCC_PLLCFGR_PLL1FRACEN) {
		fracn = (RCC.PLL1FRACR.Get() & RCC_PLL1FRACR_FRACN1_Msk) >>
			RCC_PLL1FRACR_FRACN1_Pos
	}
	if 0 != m {
		if RCC.PLLCFGR.HasBits(RCC_PLLCFGR_DIVP1EN) {
			p := ((RCC.PLL1DIVR.Get() & RCC_PLL1DIVR_DIVP1_Msk) >>
				RCC_PLL1DIVR_DIVP1_Pos) + 1
			clk.P = pllFreq(in, m, n, fracn, p)
		}
		if RCC.PLLCFGR.HasBits(RCC_PLLCFGR_DIVQ1EN) {
			q := ((RCC.PLL1DIVR.Get() & RCC_PLL1DIVR_DIVQ1_Msk) >>
				RCC_PLL1DIVR_DIVQ1_Pos) + 1
			clk.Q = pllFreq(in, m, n, fracn, q)
		}
		if RCC.PLLCFGR.HasBits(RCC_PLLCFGR_DIVR1EN) {
			r := ((RCC.PLL1DIVR.Get() & RCC_PLL1DIVR_DIVR1_Msk) >>
				RCC_PLL1DIVR_DIVR1_Pos) + 1
			clk.R = pllFreq(in, m, n, fracn, r)
		}
	}
	return clk
}

func pllFreq(in, m, n, fracn, pqr uint32) (freq uint32) {

	// We are trying to find H, where:
	//   S := HSE or HSI or CSI (based on PLL source)
	//   H := (S/M * (N + F/8192)) / P
	//
	// return uint32(((float32(in) / float32(m) *
	// 	(float32(n) + (float32(fracn) / 0x2000))) / float32(pqr)))
	//
	// BUG(?): The compiler builds and links this code just fine, with obvious
	//         FP instructions in the disassembly, but we always get a hardfault
	//         at runtime when this function is called from ANY context.
	//         I know we don't currently support hardware FPU, but shouldn't it be
	//         able to use softfp?

	return uint32(uint64(in) * uint64(8192*n+fracn) / uint64(8192*m*pqr))
}
