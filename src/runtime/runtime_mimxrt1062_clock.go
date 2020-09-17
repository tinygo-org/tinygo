// +build mimxrt1062

package runtime

import (
	"device/nxp"
	"runtime/volatile"
	"unsafe"
)

// Core, bus, and peripheral clock frequencies (Hz)
const (
	CORE_FREQ      = 600000000 // 600     MHz
	OSC_FREQ       = 100000    // 100     kHz  (see note below)
	_                          // -----------
	AHB_FREQ       = 600000000 // 600     MHz
	CAN_FREQ       = 40000000  //  40     MHz
	CKIL_SYNC_FREQ = 32768     //  32.768 kHz
	CSI_FREQ       = 12000000  //  12     MHz
	FLEXIO1_FREQ   = 30000000  //  30     MHz
	FLEXIO2_FREQ   = 30000000  //  30     MHz
	FLEXSPI2_FREQ  = 327724137 // 327.7   MHz
	FLEXSPI_FREQ   = 327724137 // 327.7   MHz
	IPG_FREQ       = 150000000 // 150     MHz
	LCDIF_FREQ     = 39272727  //  39.3   MHz
	LPI2C_FREQ     = 60000000  //  60     MHz
	LPSPI_FREQ     = 132000000 // 132     MHz
	PERCLK_FREQ    = 24000000  //  24     MHz
	SAI1_FREQ      = 63529411  //  63.5   MHz
	SAI2_FREQ      = 63529411  //  63.5   MHz
	SAI3_FREQ      = 63529411  //  63.5   MHz
	SEMC_FREQ      = 163862068 // 163.9   MHz
	SPDIF0_FREQ    = 30000000  //  30     MHz
	TRACE_FREQ     = 132000000 // 132     MHz
	UART_FREQ      = 24000000  //  24     MHz
	USDHC1_FREQ    = 198000000 // 198     MHz
	USDHC2_FREQ    = 198000000 // 198     MHz
)

// Note about OSC_FREQ from Teensyduino (cores/teensy4/startup.c):
//
// |  ARM SysTick is used for most Ardiuno timing functions, delay(), millis(),
// |  micros().  SysTick can run from either the ARM core clock, or from an
// |  "external" clock.  NXP documents it as "24 MHz XTALOSC can be the external
// |  clock source of SYSTICK" (RT1052 ref manual, rev 1, page 411).  However,
// |  NXP actually hid an undocumented divide-by-240 circuit in the hardware, so
// |  the external clock is really 100 kHz.  We use this clock rather than the
// |  ARM clock, to allow SysTick to maintain correct timing even when we change
// |  the ARM clock to run at different speeds.

// initClocks configures the core, buses, and all peripherals' clock source mux
// and dividers for runtime. The clock gates for individual peripherals are all
// disabled prior to configuration and must be enabled afterwards using function
// `enableClocks()`.
func initClocks() {
	// disable low-power mode so that __WFI doesn't lock up at runtime.
	// see: Using the MIMXRT1060/4-EVK with MCUXpresso IDE v10.3.x (v1.0.2,
	// 2019MAR01), chapter 14
	modeClkRun.set()

	// enable 1MHz clock output
	nxp.XTALOSC24M.OSC_CONFIG2.SetBits(nxp.XTALOSC24M_OSC_CONFIG2_ENABLE_1M_Msk)
	// use free 1MHz clock output
	nxp.XTALOSC24M.OSC_CONFIG2.ClearBits(nxp.XTALOSC24M_OSC_CONFIG2_MUX_1M_Msk)

	// initialize external 24 MHz clock
	nxp.CCM_ANALOG.MISC0_CLR.Set(nxp.CCM_ANALOG_MISC0_XTAL_24M_PWD_Msk) // power
	for !nxp.XTALOSC24M.LOWPWR_CTRL.HasBits(nxp.XTALOSC24M_LOWPWR_CTRL_XTALOSC_PWRUP_STAT_Msk) {
	}
	nxp.CCM_ANALOG.MISC0_SET.Set(nxp.CCM_ANALOG_MISC0_OSC_XTALOK_EN_Msk) // detect freq
	for !nxp.CCM_ANALOG.MISC0.HasBits(nxp.CCM_ANALOG_MISC0_OSC_XTALOK_Msk) {
	}
	nxp.CCM_ANALOG.MISC0_CLR.Set(nxp.CCM_ANALOG_MISC0_OSC_XTALOK_EN_Msk)

	// initialize internal RC oscillator 24 MHz clock
	nxp.XTALOSC24M.LOWPWR_CTRL.SetBits(nxp.XTALOSC24M_LOWPWR_CTRL_RC_OSC_EN_Msk)

	// switch clock source to external oscillator
	nxp.XTALOSC24M.LOWPWR_CTRL_CLR.Set(nxp.XTALOSC24M_LOWPWR_CTRL_CLR_OSC_SEL_Msk)

	// set oscillator ready counter value
	nxp.CCM.CCR.Set((nxp.CCM.CCR.Get() & ^uint32(nxp.CCM_CCR_OSCNT_Msk)) |
		((127 << nxp.CCM_CCR_OSCNT_Pos) & nxp.CCM_CCR_OSCNT_Msk))

	// set PERIPH2_CLK and PERIPH to provide stable clock before PLLs initialed
	muxIpPeriphClk2.mux(1) // PERIPH_CLK2 select OSC24M
	muxIpPeriph.mux(1)     // PERIPH select PERIPH_CLK2

	// set VDD_SOC to 1.275V, necessary to config AHB to 600 MHz
	nxp.DCDC.REG3.Set((nxp.DCDC.REG3.Get() & ^uint32(nxp.DCDC_REG3_TRG_Msk)) |
		((13 << nxp.DCDC_REG3_TRG_Pos) & nxp.DCDC_REG3_TRG_Msk))

	// wait until DCDC_STS_DC_OK bit is asserted
	for !nxp.DCDC.REG0.HasBits(nxp.DCDC_REG0_STS_DC_OK_Msk) {
	}

	// -------------------------------------------------------------------- AHB --
	divIpAhb.div(0) // divide AHB_PODF (DIV1)

	// -------------------------------------------------------------------- ADC --
	clkIpAdc1.enable(false) // disable ADC clock gates
	clkIpAdc2.enable(false) //  ~

	// ------------------------------------------------------------------- XBAR --
	clkIpXbar1.enable(false) //  disable XBAR clock gates
	clkIpXbar2.enable(false) //  ~
	clkIpXbar3.enable(false) //  ~

	// ---------------------------------------------------------------- ARM/IPG --
	divIpIpg.div(3)        // divide IPG_PODF (DIV4)
	divIpArm.div(1)        // divide ARM_PODF (DIV2)
	divIpPeriphClk2.div(0) // divide PERIPH_CLK2_PODF (DIV1)

	// ---------------------------------------------------------------- GPT/PIT --
	clkIpGpt1.enable(false)  // disable GPT/PIT clock gates
	clkIpGpt1S.enable(false) //  ~
	clkIpGpt2.enable(false)  //  ~
	clkIpGpt2S.enable(false) //  ~
	clkIpPit.enable(false)   //  ~

	// -------------------------------------------------------------------- PER --
	divIpPerclk.div(0) // divide PERCLK_PODF (DIV1)

	// ------------------------------------------------------------------ USDHC --
	clkIpUsdhc1.enable(false) // disable USDHC1 clock gate
	divIpUsdhc1.div(1)        // divide USDHC1_PODF (DIV2)
	muxIpUsdhc1.mux(1)        // USDHC1 select PLL2_PFD0
	clkIpUsdhc2.enable(false) // disable USDHC2 clock gate
	divIpUsdhc2.div(1)        // divide USDHC2_PODF (DIV2)
	muxIpUsdhc2.mux(1)        // USDHC2 select PLL2_PFD0

	// ------------------------------------------------------------------- SEMC --
	clkIpSemc.enable(false) // disable SEMC clock gate
	divIpSemc.div(1)        // divide SEMC_PODF (DIV2)
	muxIpSemcAlt.mux(0)     // SEMC_ALT select PLL2_PFD2
	muxIpSemc.mux(1)        // SEMC select SEMC_ALT

	// ---------------------------------------------------------------- FLEXSPI --
	if false {
		// TODO: external flash is on this bus, configured via DCD block
		clkIpFlexSpi.enable(false) // disable FLEXSPI clock gate
		divIpFlexSpi.div(0)        // divide FLEXSPI_PODF (DIV1)
		muxIpFlexSpi.mux(2)        // FLEXSPI select PLL2_PFD2
	}
	clkIpFlexSpi2.enable(false) // disable FLEXSPI2 clock gate
	divIpFlexSpi2.div(0)        // divide FLEXSPI2_PODF (DIV1)
	muxIpFlexSpi2.mux(0)        // FLEXSPI2 select PLL2_PFD2

	// -------------------------------------------------------------------- CSI --
	clkIpCsi.enable(false) // disable CSI clock gate
	divIpCsi.div(1)        // divide CSI_PODF (DIV2)
	muxIpCsi.mux(0)        // CSI select OSC24M

	// ------------------------------------------------------------------ LPSPI --
	clkIpLpspi1.enable(false) // disable LPSPI clock gate
	clkIpLpspi2.enable(false) //  ~
	clkIpLpspi3.enable(false) //  ~
	clkIpLpspi4.enable(false) //  ~
	divIpLpspi.div(3)         // divide LPSPI_PODF (DIV4)
	muxIpLpspi.mux(2)         // LPSPI select PLL2

	// ------------------------------------------------------------------ TRACE --
	clkIpTrace.enable(false) // disable TRACE clock gate
	divIpTrace.div(3)        // divide TRACE_PODF (DIV4)
	muxIpTrace.mux(0)        // TRACE select PLL2_MAIN

	// -------------------------------------------------------------------- SAI --
	clkIpSai1.enable(false) // disable SAI1 clock gate
	divIpSai1Pre.div(3)     // divide SAI1_CLK_PRED (DIV4)
	divIpSai1.div(1)        // divide SAI1_CLK_PODF (DIV2)
	muxIpSai1.mux(0)        // SAI1 select PLL3_PFD2
	clkIpSai2.enable(false) // disable SAI2 clock gate
	divIpSai2Pre.div(3)     // divide SAI2_CLK_PRED (DIV4)
	divIpSai2.div(1)        // divide SAI2_CLK_PODF (DIV2)
	muxIpSai2.mux(0)        // SAI2 select PLL3_PFD2
	clkIpSai3.enable(false) // disable SAI3 clock gate
	divIpSai3Pre.div(3)     // divide SAI3_CLK_PRED (DIV4)
	divIpSai3.div(1)        // divide SAI3_CLK_PODF (DIV2)
	muxIpSai3.mux(0)        // SAI3 select PLL3_PFD2

	// ------------------------------------------------------------------ LPI2C --
	clkIpLpi2c1.enable(false) // disable LPI2C clock gate
	clkIpLpi2c2.enable(false) //  ~
	clkIpLpi2c3.enable(false) //  ~
	divIpLpi2c.div(0)         // divide LPI2C_CLK_PODF (DIV1)
	muxIpLpi2c.mux(0)         // LPI2C select PLL3_SW_60M

	// -------------------------------------------------------------------- CAN --
	clkIpCan1.enable(false)  // disable CAN clock gate
	clkIpCan2.enable(false)  //  ~
	clkIpCan3.enable(false)  //  ~
	clkIpCan1S.enable(false) //  ~
	clkIpCan2S.enable(false) //  ~
	clkIpCan3S.enable(false) //  ~
	divIpCan.div(1)          // divide CAN_CLK_PODF (DIV2)
	muxIpCan.mux(2)          // CAN select PLL3_SW_80M

	// ------------------------------------------------------------------- UART --
	clkIpLpuart1.enable(false) // disable UART clock gate
	clkIpLpuart2.enable(false) //  ~
	clkIpLpuart3.enable(false) //  ~
	clkIpLpuart4.enable(false) //  ~
	clkIpLpuart5.enable(false) //  ~
	clkIpLpuart6.enable(false) //  ~
	clkIpLpuart7.enable(false) //  ~
	clkIpLpuart8.enable(false) //  ~
	divIpUart.div(0)           // divide UART_CLK_PODF (DIV1)
	muxIpUart.mux(1)           // UART select OSC

	// -------------------------------------------------------------------- LCD --
	clkIpLcdPixel.enable(false) // disable LCDIF clock gate
	divIpLcdifPre.div(1)        // divide LCDIF_PRED (DIV2)
	divIpLcdif.div(3)           // divide LCDIF_CLK_PODF (DIV4)
	muxIpLcdifPre.mux(5)        // LCDIF_PRE select PLL3_PFD1

	// ------------------------------------------------------------------ SPDIF --
	clkIpSpdif.enable(false) // disable SPDIF clock gate
	divIpSpdif0Pre.div(1)    // divide SPDIF0_CLK_PRED (DIV2)
	divIpSpdif0.div(7)       // divide SPDIF0_CLK_PODF (DIV8)
	muxIpSpdif.mux(3)        // SPDIF select PLL3_SW

	// ----------------------------------------------------------------- FLEXIO --
	clkIpFlexio1.enable(false) // disable FLEXIO1 clock gate
	divIpFlexio1Pre.div(1)     // divide FLEXIO1_CLK_PRED (DIV2)
	divIpFlexio1.div(7)        // divide FLEXIO1_CLK_PODF (DIV8)
	muxIpFlexio1.mux(3)        // FLEXIO1 select PLL3_SW
	clkIpFlexio2.enable(false) // enable FLEXIO2 clock gate
	divIpFlexio2Pre.div(1)     // divide FLEXIO2_CLK_PRED (DIV2)
	divIpFlexio2.div(7)        // divide FLEXIO2_CLK_PODF (DIV8)
	muxIpFlexio2.mux(3)        // FLEXIO2 select PLL3_SW

	// ---------------------------------------------------------------- PLL/PFD --
	muxIpPll3Sw.mux(0) // PLL3_SW select PLL3_MAIN

	armPllConfig.configure() // init ARM PLL
	// SYS PLL (PLL2) @ 528 MHz
	//   PFD0 = 396    MHz -> USDHC1/USDHC2(DIV2)=198 MHz
	//   PFD1 = 594    MHz -> (currently unused)
	//   PFD2 = 327.72 MHz -> SEMC(DIV2)=163.86 MHz, FlexSPI/FlexSPI2=327.72 MHz
	//   PFD3 = 454.73 MHz -> (currently unused)
	sysPllConfig.configure(24, 16, 29, 16) // init SYS PLL and PFDs

	// USB1 PLL (PLL3) @ 480 MHz
	//   PFD0 -> (currently unused)
	//   PFD1 -> (currently unused)
	//   PFD2 -> (currently unused)
	//   PFD3 -> (currently unused)
	usb1PllConfig.configure() // init USB1 PLL and PFDs
	usb2PllConfig.configure() // init USB2 PLL

	// --------------------------------------------------------------- ARM CORE --
	muxIpPrePeriph.mux(3)  // PRE_PERIPH select ARM_PLL
	muxIpPeriph.mux(0)     // PERIPH select PRE_PERIPH
	muxIpPeriphClk2.mux(1) // PERIPH_CLK2 select OSC
	muxIpPerclk.mux(1)     // PERCLK select OSC

	// ------------------------------------------------------------------- LVDS --
	// set LVDS1 clock source
	nxp.CCM_ANALOG.MISC1.Set((nxp.CCM_ANALOG.MISC1.Get() & ^uint32(nxp.CCM_ANALOG_MISC1_LVDS1_CLK_SEL_Msk)) |
		((0 << nxp.CCM_ANALOG_MISC1_LVDS1_CLK_SEL_Pos) & nxp.CCM_ANALOG_MISC1_LVDS1_CLK_SEL_Msk))

	// ----------------------------------------------------------------- CLKOUT --
	// set CLOCK_OUT1 divider
	nxp.CCM.CCOSR.Set((nxp.CCM.CCOSR.Get() & ^uint32(nxp.CCM_CCOSR_CLKO1_DIV_Msk)) |
		((0 << nxp.CCM_CCOSR_CLKO1_DIV_Pos) & nxp.CCM_CCOSR_CLKO1_DIV_Msk))
	// set CLOCK_OUT1 source
	nxp.CCM.CCOSR.Set((nxp.CCM.CCOSR.Get() & ^uint32(nxp.CCM_CCOSR_CLKO1_SEL_Msk)) |
		((1 << nxp.CCM_CCOSR_CLKO1_SEL_Pos) & nxp.CCM_CCOSR_CLKO1_SEL_Msk))
	// set CLOCK_OUT2 divider
	nxp.CCM.CCOSR.Set((nxp.CCM.CCOSR.Get() & ^uint32(nxp.CCM_CCOSR_CLKO2_DIV_Msk)) |
		((0 << nxp.CCM_CCOSR_CLKO2_DIV_Pos) & nxp.CCM_CCOSR_CLKO2_DIV_Msk))
	// set CLOCK_OUT2 source
	nxp.CCM.CCOSR.Set((nxp.CCM.CCOSR.Get() & ^uint32(nxp.CCM_CCOSR_CLKO2_SEL_Msk)) |
		((18 << nxp.CCM_CCOSR_CLKO2_SEL_Pos) & nxp.CCM_CCOSR_CLKO2_SEL_Msk))

	nxp.CCM.CCOSR.ClearBits(nxp.CCM_CCOSR_CLK_OUT_SEL_Msk) // set CLK_OUT1 drives CLK_OUT
	nxp.CCM.CCOSR.SetBits(nxp.CCM_CCOSR_CLKO1_EN_Msk)      // enable CLK_OUT1
	nxp.CCM.CCOSR.SetBits(nxp.CCM_CCOSR_CLKO2_EN_Msk)      // enable CLK_OUT2

	// ----------------------------------------------------------------- IOMUXC --
	clkIpIomuxcGpr.enable(false) // disable IOMUXC_GPR clock gate
	clkIpIomuxc.enable(false)    // disable IOMUXC clock gate
	// set GPT1 High frequency reference clock source
	nxp.IOMUXC_GPR.GPR5.ClearBits(nxp.IOMUXC_GPR_GPR5_VREF_1M_CLK_GPT1_Msk)
	// set GPT2 High frequency reference clock source
	nxp.IOMUXC_GPR.GPR5.ClearBits(nxp.IOMUXC_GPR_GPR5_VREF_1M_CLK_GPT2_Msk)

	// ------------------------------------------------------------------- GPIO --
	clkIpGpio1.enable(false) // disable GPIO clock gates
	clkIpGpio2.enable(false) //  ~
	clkIpGpio3.enable(false) //  ~
	clkIpGpio4.enable(false) //  ~
}

func enableTimerClocks() {

	// ---------------------------------------------------------------- GPT/PIT --
	clkIpGpt1.enable(true)  // enable GPT/PIT clock gates
	clkIpGpt1S.enable(true) //  ~
	clkIpGpt2.enable(true)  //  ~
	clkIpGpt2S.enable(true) //  ~
	clkIpPit.enable(true)   //  ~
}

func enablePinClocks() {

	// ----------------------------------------------------------------- IOMUXC --
	clkIpIomuxcGpr.enable(true) // enable IOMUXC clock gates
	clkIpIomuxc.enable(true)    //  ~

	// ------------------------------------------------------------------- GPIO --
	clkIpGpio1.enable(true) // enable GPIO clock gates
	clkIpGpio2.enable(true) //  ~
	clkIpGpio3.enable(true) //  ~
	clkIpGpio4.enable(true) //  ~
}

// enableClocks enables the clock gates for peripherals used by this runtime.
func enablePeripheralClocks() {

	// -------------------------------------------------------------------- ADC --
	clkIpAdc1.enable(true) // enable ADC clock gates
	clkIpAdc2.enable(true) //  ~

	// ------------------------------------------------------------------- XBAR --
	clkIpXbar1.enable(true) // enable XBAR clock gates
	clkIpXbar2.enable(true) //  ~
	clkIpXbar3.enable(true) //  ~

	// ------------------------------------------------------------------ USDHC --
	clkIpUsdhc1.enable(true) // enable USDHC clock gates
	clkIpUsdhc2.enable(true) //  ~

	// ------------------------------------------------------------------- SEMC --
	clkIpSemc.enable(true) // enable SEMC clock gate

	// ---------------------------------------------------------------- FLEXSPI --
	clkIpFlexSpi2.enable(true) // enable FLEXSPI2 clock gate

	// ------------------------------------------------------------------ LPSPI --
	clkIpLpspi1.enable(true) // enable LPSPI clock gate
	clkIpLpspi2.enable(true) //  ~
	clkIpLpspi3.enable(true) //  ~
	clkIpLpspi4.enable(true) //  ~

	// ------------------------------------------------------------------ LPI2C --
	clkIpLpi2c1.enable(true) // enable LPI2C clock gate
	clkIpLpi2c2.enable(true) //  ~
	clkIpLpi2c3.enable(true) //  ~

	// -------------------------------------------------------------------- CAN --
	clkIpCan1.enable(true)  // enable CAN clock gate
	clkIpCan2.enable(true)  //  ~
	clkIpCan3.enable(true)  //  ~
	clkIpCan1S.enable(true) //  ~
	clkIpCan2S.enable(true) //  ~
	clkIpCan3S.enable(true) //  ~

	// ------------------------------------------------------------------- UART --
	clkIpLpuart1.enable(true) // enable UART clock gate
	clkIpLpuart2.enable(true) //  ~
	clkIpLpuart3.enable(true) //  ~
	clkIpLpuart4.enable(true) //  ~
	clkIpLpuart5.enable(true) //  ~
	clkIpLpuart6.enable(true) //  ~
	clkIpLpuart7.enable(true) //  ~
	clkIpLpuart8.enable(true) //  ~

	// ----------------------------------------------------------------- FLEXIO --
	clkIpFlexio1.enable(true) // enable FLEXIO clock gates
	clkIpFlexio2.enable(true) //  ~
}

type clock uint32

// Named oscillators
const (
	clkCpu         clock = 0x0  // CPU clock
	clkAhb         clock = 0x1  // AHB clock
	clkSemc        clock = 0x2  // SEMC clock
	clkIpg         clock = 0x3  // IPG clock
	clkPer         clock = 0x4  // PER clock
	clkOsc         clock = 0x5  // OSC clock selected by PMU_LOWPWR_CTRL[OSC_SEL]
	clkRtc         clock = 0x6  // RTC clock (RTCCLK)
	clkArmPll      clock = 0x7  // ARMPLLCLK
	clkUsb1Pll     clock = 0x8  // USB1PLLCLK
	clkUsb1PllPfd0 clock = 0x9  // USB1PLLPDF0CLK
	clkUsb1PllPfd1 clock = 0xA  // USB1PLLPFD1CLK
	clkUsb1PllPfd2 clock = 0xB  // USB1PLLPFD2CLK
	clkUsb1PllPfd3 clock = 0xC  // USB1PLLPFD3CLK
	clkUsb2Pll     clock = 0xD  // USB2PLLCLK
	clkSysPll      clock = 0xE  // SYSPLLCLK
	clkSysPllPfd0  clock = 0xF  // SYSPLLPDF0CLK
	clkSysPllPfd1  clock = 0x10 // SYSPLLPFD1CLK
	clkSysPllPfd2  clock = 0x11 // SYSPLLPFD2CLK
	clkSysPllPfd3  clock = 0x12 // SYSPLLPFD3CLK
	clkEnetPll0    clock = 0x13 // Enet PLLCLK ref_enetpll0
	clkEnetPll1    clock = 0x14 // Enet PLLCLK ref_enetpll1
	clkEnetPll2    clock = 0x15 // Enet PLLCLK ref_enetpll2
	clkAudioPll    clock = 0x16 // Audio PLLCLK
	clkVideoPll    clock = 0x17 // Video PLLCLK
)

// Named clocks of integrated peripherals
const (
	clkIpAipsTz1     clock = (0 << 8) | nxp.CCM_CCGR0_CG0_Pos  // CCGR0, CG0
	clkIpAipsTz2     clock = (0 << 8) | nxp.CCM_CCGR0_CG1_Pos  // CCGR0, CG1
	clkIpMqs         clock = (0 << 8) | nxp.CCM_CCGR0_CG2_Pos  // CCGR0, CG2
	clkIpFlexSpiExsc clock = (0 << 8) | nxp.CCM_CCGR0_CG3_Pos  // CCGR0, CG3
	clkIpSimMMain    clock = (0 << 8) | nxp.CCM_CCGR0_CG4_Pos  // CCGR0, CG4
	clkIpDcp         clock = (0 << 8) | nxp.CCM_CCGR0_CG5_Pos  // CCGR0, CG5
	clkIpLpuart3     clock = (0 << 8) | nxp.CCM_CCGR0_CG6_Pos  // CCGR0, CG6
	clkIpCan1        clock = (0 << 8) | nxp.CCM_CCGR0_CG7_Pos  // CCGR0, CG7
	clkIpCan1S       clock = (0 << 8) | nxp.CCM_CCGR0_CG8_Pos  // CCGR0, CG8
	clkIpCan2        clock = (0 << 8) | nxp.CCM_CCGR0_CG9_Pos  // CCGR0, CG9
	clkIpCan2S       clock = (0 << 8) | nxp.CCM_CCGR0_CG10_Pos // CCGR0, CG10
	clkIpTrace       clock = (0 << 8) | nxp.CCM_CCGR0_CG11_Pos // CCGR0, CG11
	clkIpGpt2        clock = (0 << 8) | nxp.CCM_CCGR0_CG12_Pos // CCGR0, CG12
	clkIpGpt2S       clock = (0 << 8) | nxp.CCM_CCGR0_CG13_Pos // CCGR0, CG13
	clkIpLpuart2     clock = (0 << 8) | nxp.CCM_CCGR0_CG14_Pos // CCGR0, CG14
	clkIpGpio2       clock = (0 << 8) | nxp.CCM_CCGR0_CG15_Pos // CCGR0, CG15

	clkIpLpspi1   clock = (1 << 8) | nxp.CCM_CCGR1_CG0_Pos  // CCGR1, CG0
	clkIpLpspi2   clock = (1 << 8) | nxp.CCM_CCGR1_CG1_Pos  // CCGR1, CG1
	clkIpLpspi3   clock = (1 << 8) | nxp.CCM_CCGR1_CG2_Pos  // CCGR1, CG2
	clkIpLpspi4   clock = (1 << 8) | nxp.CCM_CCGR1_CG3_Pos  // CCGR1, CG3
	clkIpAdc2     clock = (1 << 8) | nxp.CCM_CCGR1_CG4_Pos  // CCGR1, CG4
	clkIpEnet     clock = (1 << 8) | nxp.CCM_CCGR1_CG5_Pos  // CCGR1, CG5
	clkIpPit      clock = (1 << 8) | nxp.CCM_CCGR1_CG6_Pos  // CCGR1, CG6
	clkIpAoi2     clock = (1 << 8) | nxp.CCM_CCGR1_CG7_Pos  // CCGR1, CG7
	clkIpAdc1     clock = (1 << 8) | nxp.CCM_CCGR1_CG8_Pos  // CCGR1, CG8
	clkIpSemcExsc clock = (1 << 8) | nxp.CCM_CCGR1_CG9_Pos  // CCGR1, CG9
	clkIpGpt1     clock = (1 << 8) | nxp.CCM_CCGR1_CG10_Pos // CCGR1, CG10
	clkIpGpt1S    clock = (1 << 8) | nxp.CCM_CCGR1_CG11_Pos // CCGR1, CG11
	clkIpLpuart4  clock = (1 << 8) | nxp.CCM_CCGR1_CG12_Pos // CCGR1, CG12
	clkIpGpio1    clock = (1 << 8) | nxp.CCM_CCGR1_CG13_Pos // CCGR1, CG13
	clkIpCsu      clock = (1 << 8) | nxp.CCM_CCGR1_CG14_Pos // CCGR1, CG14
	clkIpGpio5    clock = (1 << 8) | nxp.CCM_CCGR1_CG15_Pos // CCGR1, CG15

	clkIpOcramExsc  clock = (2 << 8) | nxp.CCM_CCGR2_CG0_Pos  // CCGR2, CG0
	clkIpCsi        clock = (2 << 8) | nxp.CCM_CCGR2_CG1_Pos  // CCGR2, CG1
	clkIpIomuxcSnvs clock = (2 << 8) | nxp.CCM_CCGR2_CG2_Pos  // CCGR2, CG2
	clkIpLpi2c1     clock = (2 << 8) | nxp.CCM_CCGR2_CG3_Pos  // CCGR2, CG3
	clkIpLpi2c2     clock = (2 << 8) | nxp.CCM_CCGR2_CG4_Pos  // CCGR2, CG4
	clkIpLpi2c3     clock = (2 << 8) | nxp.CCM_CCGR2_CG5_Pos  // CCGR2, CG5
	clkIpOcotp      clock = (2 << 8) | nxp.CCM_CCGR2_CG6_Pos  // CCGR2, CG6
	clkIpXbar3      clock = (2 << 8) | nxp.CCM_CCGR2_CG7_Pos  // CCGR2, CG7
	clkIpIpmux1     clock = (2 << 8) | nxp.CCM_CCGR2_CG8_Pos  // CCGR2, CG8
	clkIpIpmux2     clock = (2 << 8) | nxp.CCM_CCGR2_CG9_Pos  // CCGR2, CG9
	clkIpIpmux3     clock = (2 << 8) | nxp.CCM_CCGR2_CG10_Pos // CCGR2, CG10
	clkIpXbar1      clock = (2 << 8) | nxp.CCM_CCGR2_CG11_Pos // CCGR2, CG11
	clkIpXbar2      clock = (2 << 8) | nxp.CCM_CCGR2_CG12_Pos // CCGR2, CG12
	clkIpGpio3      clock = (2 << 8) | nxp.CCM_CCGR2_CG13_Pos // CCGR2, CG13
	clkIpLcd        clock = (2 << 8) | nxp.CCM_CCGR2_CG14_Pos // CCGR2, CG14
	clkIpPxp        clock = (2 << 8) | nxp.CCM_CCGR2_CG15_Pos // CCGR2, CG15

	clkIpFlexio2       clock = (3 << 8) | nxp.CCM_CCGR3_CG0_Pos  // CCGR3, CG0
	clkIpLpuart5       clock = (3 << 8) | nxp.CCM_CCGR3_CG1_Pos  // CCGR3, CG1
	clkIpSemc          clock = (3 << 8) | nxp.CCM_CCGR3_CG2_Pos  // CCGR3, CG2
	clkIpLpuart6       clock = (3 << 8) | nxp.CCM_CCGR3_CG3_Pos  // CCGR3, CG3
	clkIpAoi1          clock = (3 << 8) | nxp.CCM_CCGR3_CG4_Pos  // CCGR3, CG4
	clkIpLcdPixel      clock = (3 << 8) | nxp.CCM_CCGR3_CG5_Pos  // CCGR3, CG5
	clkIpGpio4         clock = (3 << 8) | nxp.CCM_CCGR3_CG6_Pos  // CCGR3, CG6
	clkIpEwm0          clock = (3 << 8) | nxp.CCM_CCGR3_CG7_Pos  // CCGR3, CG7
	clkIpWdog1         clock = (3 << 8) | nxp.CCM_CCGR3_CG8_Pos  // CCGR3, CG8
	clkIpFlexRam       clock = (3 << 8) | nxp.CCM_CCGR3_CG9_Pos  // CCGR3, CG9
	clkIpAcmp1         clock = (3 << 8) | nxp.CCM_CCGR3_CG10_Pos // CCGR3, CG10
	clkIpAcmp2         clock = (3 << 8) | nxp.CCM_CCGR3_CG11_Pos // CCGR3, CG11
	clkIpAcmp3         clock = (3 << 8) | nxp.CCM_CCGR3_CG12_Pos // CCGR3, CG12
	clkIpAcmp4         clock = (3 << 8) | nxp.CCM_CCGR3_CG13_Pos // CCGR3, CG13
	clkIpOcram         clock = (3 << 8) | nxp.CCM_CCGR3_CG14_Pos // CCGR3, CG14
	clkIpIomuxcSnvsGpr clock = (3 << 8) | nxp.CCM_CCGR3_CG15_Pos // CCGR3, CG15

	clkIpIomuxc    clock = (4 << 8) | nxp.CCM_CCGR4_CG1_Pos  // CCGR4, CG1
	clkIpIomuxcGpr clock = (4 << 8) | nxp.CCM_CCGR4_CG2_Pos  // CCGR4, CG2
	clkIpBee       clock = (4 << 8) | nxp.CCM_CCGR4_CG3_Pos  // CCGR4, CG3
	clkIpSimM7     clock = (4 << 8) | nxp.CCM_CCGR4_CG4_Pos  // CCGR4, CG4
	clkIpTsc       clock = (4 << 8) | nxp.CCM_CCGR4_CG5_Pos  // CCGR4, CG5
	clkIpSimM      clock = (4 << 8) | nxp.CCM_CCGR4_CG6_Pos  // CCGR4, CG6
	clkIpSimEms    clock = (4 << 8) | nxp.CCM_CCGR4_CG7_Pos  // CCGR4, CG7
	clkIpPwm1      clock = (4 << 8) | nxp.CCM_CCGR4_CG8_Pos  // CCGR4, CG8
	clkIpPwm2      clock = (4 << 8) | nxp.CCM_CCGR4_CG9_Pos  // CCGR4, CG9
	clkIpPwm3      clock = (4 << 8) | nxp.CCM_CCGR4_CG10_Pos // CCGR4, CG10
	clkIpPwm4      clock = (4 << 8) | nxp.CCM_CCGR4_CG11_Pos // CCGR4, CG11
	clkIpEnc1      clock = (4 << 8) | nxp.CCM_CCGR4_CG12_Pos // CCGR4, CG12
	clkIpEnc2      clock = (4 << 8) | nxp.CCM_CCGR4_CG13_Pos // CCGR4, CG13
	clkIpEnc3      clock = (4 << 8) | nxp.CCM_CCGR4_CG14_Pos // CCGR4, CG14
	clkIpEnc4      clock = (4 << 8) | nxp.CCM_CCGR4_CG15_Pos // CCGR4, CG15

	clkIpRom     clock = (5 << 8) | nxp.CCM_CCGR5_CG0_Pos  // CCGR5, CG0
	clkIpFlexio1 clock = (5 << 8) | nxp.CCM_CCGR5_CG1_Pos  // CCGR5, CG1
	clkIpWdog3   clock = (5 << 8) | nxp.CCM_CCGR5_CG2_Pos  // CCGR5, CG2
	clkIpDma     clock = (5 << 8) | nxp.CCM_CCGR5_CG3_Pos  // CCGR5, CG3
	clkIpKpp     clock = (5 << 8) | nxp.CCM_CCGR5_CG4_Pos  // CCGR5, CG4
	clkIpWdog2   clock = (5 << 8) | nxp.CCM_CCGR5_CG5_Pos  // CCGR5, CG5
	clkIpAipsTz4 clock = (5 << 8) | nxp.CCM_CCGR5_CG6_Pos  // CCGR5, CG6
	clkIpSpdif   clock = (5 << 8) | nxp.CCM_CCGR5_CG7_Pos  // CCGR5, CG7
	clkIpSimMain clock = (5 << 8) | nxp.CCM_CCGR5_CG8_Pos  // CCGR5, CG8
	clkIpSai1    clock = (5 << 8) | nxp.CCM_CCGR5_CG9_Pos  // CCGR5, CG9
	clkIpSai2    clock = (5 << 8) | nxp.CCM_CCGR5_CG10_Pos // CCGR5, CG10
	clkIpSai3    clock = (5 << 8) | nxp.CCM_CCGR5_CG11_Pos // CCGR5, CG11
	clkIpLpuart1 clock = (5 << 8) | nxp.CCM_CCGR5_CG12_Pos // CCGR5, CG12
	clkIpLpuart7 clock = (5 << 8) | nxp.CCM_CCGR5_CG13_Pos // CCGR5, CG13
	clkIpSnvsHp  clock = (5 << 8) | nxp.CCM_CCGR5_CG14_Pos // CCGR5, CG14
	clkIpSnvsLp  clock = (5 << 8) | nxp.CCM_CCGR5_CG15_Pos // CCGR5, CG15

	clkIpUsbOh3  clock = (6 << 8) | nxp.CCM_CCGR6_CG0_Pos  // CCGR6, CG0
	clkIpUsdhc1  clock = (6 << 8) | nxp.CCM_CCGR6_CG1_Pos  // CCGR6, CG1
	clkIpUsdhc2  clock = (6 << 8) | nxp.CCM_CCGR6_CG2_Pos  // CCGR6, CG2
	clkIpDcdc    clock = (6 << 8) | nxp.CCM_CCGR6_CG3_Pos  // CCGR6, CG3
	clkIpIpmux4  clock = (6 << 8) | nxp.CCM_CCGR6_CG4_Pos  // CCGR6, CG4
	clkIpFlexSpi clock = (6 << 8) | nxp.CCM_CCGR6_CG5_Pos  // CCGR6, CG5
	clkIpTrng    clock = (6 << 8) | nxp.CCM_CCGR6_CG6_Pos  // CCGR6, CG6
	clkIpLpuart8 clock = (6 << 8) | nxp.CCM_CCGR6_CG7_Pos  // CCGR6, CG7
	clkIpTimer4  clock = (6 << 8) | nxp.CCM_CCGR6_CG8_Pos  // CCGR6, CG8
	clkIpAipsTz3 clock = (6 << 8) | nxp.CCM_CCGR6_CG9_Pos  // CCGR6, CG9
	clkIpSimPer  clock = (6 << 8) | nxp.CCM_CCGR6_CG10_Pos // CCGR6, CG10
	clkIpAnadig  clock = (6 << 8) | nxp.CCM_CCGR6_CG11_Pos // CCGR6, CG11
	clkIpLpi2c4  clock = (6 << 8) | nxp.CCM_CCGR6_CG12_Pos // CCGR6, CG12
	clkIpTimer1  clock = (6 << 8) | nxp.CCM_CCGR6_CG13_Pos // CCGR6, CG13
	clkIpTimer2  clock = (6 << 8) | nxp.CCM_CCGR6_CG14_Pos // CCGR6, CG14
	clkIpTimer3  clock = (6 << 8) | nxp.CCM_CCGR6_CG15_Pos // CCGR6, CG15

	clkIpEnet2    clock = (7 << 8) | nxp.CCM_CCGR7_CG0_Pos // CCGR7, CG0
	clkIpFlexSpi2 clock = (7 << 8) | nxp.CCM_CCGR7_CG1_Pos // CCGR7, CG1
	clkIpAxbsL    clock = (7 << 8) | nxp.CCM_CCGR7_CG2_Pos // CCGR7, CG2
	clkIpCan3     clock = (7 << 8) | nxp.CCM_CCGR7_CG3_Pos // CCGR7, CG3
	clkIpCan3S    clock = (7 << 8) | nxp.CCM_CCGR7_CG4_Pos // CCGR7, CG4
	clkIpAipsLite clock = (7 << 8) | nxp.CCM_CCGR7_CG5_Pos // CCGR7, CG5
	clkIpFlexio3  clock = (7 << 8) | nxp.CCM_CCGR7_CG6_Pos // CCGR7, CG6
)

// Selected clock offsets
const (
	offCCSR   = 0x0C
	offCBCDR  = 0x14
	offCBCMR  = 0x18
	offCSCMR1 = 0x1C
	offCSCMR2 = 0x20
	offCSCDR1 = 0x24
	offCDCDR  = 0x30
	offCSCDR2 = 0x38
	offCSCDR3 = 0x3C
	offCACRR  = 0x10
	offCS1CDR = 0x28
	offCS2CDR = 0x2C

	offPllArm   = 0x00
	offPllSys   = 0x30
	offPllUsb1  = 0x10
	offPllAudio = 0x70
	offPllVideo = 0xA0
	offPllEnet  = 0xE0
	offPllUsb2  = 0x20

	noBusyWait = 0x20
)

// analog pll definition
const (
	pllBypassPos       = 16
	pllBypassClkSrcMsk = 0xC000
	pllBypassClkSrcPos = 14
)

// getFreq returns the calculated frequency of the receiver clock clk.
func (clk clock) getFreq() uint32 {
	switch clk {
	case clkCpu, clkAhb:
		return getAhbFreq()
	case clkSemc:
		return getSemcFreq()
	case clkIpg:
		return getIpgFreq()
	case clkPer:
		return getPerClkFreq()
	case clkOsc:
		return getOscFreq()
	case clkRtc:
		return getRtcFreq()
	case clkArmPll:
		return clkPllArm.getPllFreq()
	case clkUsb1Pll:
		return clkPllUsb1.getPllFreq()
	case clkUsb1PllPfd0:
		return clkPfd0.getUsb1PfdFreq()
	case clkUsb1PllPfd1:
		return clkPfd1.getUsb1PfdFreq()
	case clkUsb1PllPfd2:
		return clkPfd2.getUsb1PfdFreq()
	case clkUsb1PllPfd3:
		return clkPfd3.getUsb1PfdFreq()
	case clkUsb2Pll:
		return clkPllUsb2.getPllFreq()
	case clkSysPll:
		return clkPllSys.getPllFreq()
	case clkSysPllPfd0:
		return clkPfd0.getSysPfdFreq()
	case clkSysPllPfd1:
		return clkPfd1.getSysPfdFreq()
	case clkSysPllPfd2:
		return clkPfd2.getSysPfdFreq()
	case clkSysPllPfd3:
		return clkPfd3.getSysPfdFreq()
	case clkEnetPll0:
		return clkPllEnet.getPllFreq()
	case clkEnetPll1:
		return clkPllEnet2.getPllFreq()
	case clkEnetPll2:
		return clkPllEnet25M.getPllFreq()
	case clkAudioPll:
		return clkPllAudio.getPllFreq()
	case clkVideoPll:
		return clkPllVideo.getPllFreq()
	default:
		panic("runtime: invalid clock")
	}
}

// getOscFreq returns the XTAL OSC clock frequency
func getOscFreq() uint32 {
	return 24000000 // 24 MHz
}

// getRtcFreq returns the RTC clock frequency
func getRtcFreq() uint32 {
	return 32768 // 32.768 kHz
}

func ccmCbcmrPeriphClk2Sel(n uint32) uint32 {
	return (n << nxp.CCM_CBCMR_PERIPH_CLK2_SEL_Pos) & nxp.CCM_CBCMR_PERIPH_CLK2_SEL_Msk
}

func ccmCbcmrPrePeriphClkSel(n uint32) uint32 {
	return (n << nxp.CCM_CBCMR_PRE_PERIPH_CLK_SEL_Pos) & nxp.CCM_CBCMR_PRE_PERIPH_CLK_SEL_Msk
}

// getPeriphClkFreq returns the PERIPH clock frequency
func getPeriphClkFreq() uint32 {
	freq := uint32(0)
	if nxp.CCM.CBCDR.HasBits(nxp.CCM_CBCDR_PERIPH_CLK_SEL_Msk) {
		// Periph_clk2_clk -> Periph_clk
		switch nxp.CCM.CBCMR.Get() & nxp.CCM_CBCMR_PERIPH_CLK2_SEL_Msk {
		case ccmCbcmrPeriphClk2Sel(0):
			// Pll3_sw_clk -> Periph_clk2_clk -> Periph_clk
			freq = clkPllUsb1.getPllFreq()
		case ccmCbcmrPeriphClk2Sel(1):
			// Osc_clk -> Periph_clk2_clk -> Periph_clk
			freq = getOscFreq()
		case ccmCbcmrPeriphClk2Sel(2):
			freq = clkPllSys.getPllFreq()
		case ccmCbcmrPeriphClk2Sel(3):
			freq = 0
		}
		freq /= ((nxp.CCM.CBCDR.Get() & nxp.CCM_CBCDR_PERIPH_CLK2_PODF_Msk) >> nxp.CCM_CBCDR_PERIPH_CLK2_PODF_Pos) + 1
	} else {
		// Pre_Periph_clk -> Periph_clk
		switch nxp.CCM.CBCMR.Get() & nxp.CCM_CBCMR_PRE_PERIPH_CLK_SEL_Msk {
		case ccmCbcmrPrePeriphClkSel(0):
			// PLL2 -> Pre_Periph_clk -> Periph_clk
			freq = clkPllSys.getPllFreq()
		case ccmCbcmrPrePeriphClkSel(1):
			// PLL2 PFD2 -> Pre_Periph_clk -> Periph_clk
			freq = clkPfd2.getSysPfdFreq()
		case ccmCbcmrPrePeriphClkSel(2):
			// PLL2 PFD0 -> Pre_Periph_clk -> Periph_clk
			freq = clkPfd0.getSysPfdFreq()
		case ccmCbcmrPrePeriphClkSel(3):
			// PLL1 divided(/2) -> Pre_Periph_clk -> Periph_clk
			freq = clkPllArm.getPllFreq() / (((nxp.CCM.CACRR.Get() & nxp.CCM_CACRR_ARM_PODF_Msk) >> nxp.CCM_CACRR_ARM_PODF_Pos) + 1)
		}
	}
	return freq
}

// getAhbFreq returns the AHB clock frequency
func getAhbFreq() uint32 {
	return getPeriphClkFreq() / (((nxp.CCM.CBCDR.Get() & nxp.CCM_CBCDR_AHB_PODF_Msk) >> nxp.CCM_CBCDR_AHB_PODF_Pos) + 1)
}

// getSemcFreq returns the SEMC clock frequency
func getSemcFreq() uint32 {

	freq := uint32(0)

	if nxp.CCM.CBCDR.HasBits(nxp.CCM_CBCDR_SEMC_CLK_SEL_Msk) {
		// SEMC alternative clock -> SEMC Clock

		if nxp.CCM.CBCDR.HasBits(nxp.CCM_CBCDR_SEMC_ALT_CLK_SEL_Msk) {
			// PLL3 PFD1 -> SEMC alternative clock -> SEMC Clock
			freq = clkPfd1.getUsb1PfdFreq()
		} else {
			// PLL2 PFD2 -> SEMC alternative clock -> SEMC Clock
			freq = clkPfd2.getSysPfdFreq()
		}
	} else {
		// Periph_clk -> SEMC Clock
		freq = getPeriphClkFreq()
	}

	freq /= ((nxp.CCM.CBCDR.Get() & nxp.CCM_CBCDR_SEMC_PODF_Msk) >> nxp.CCM_CBCDR_SEMC_PODF_Pos) + 1

	return freq
}

// getIpgFreq returns the IPG clock frequency
func getIpgFreq() uint32 {
	return getAhbFreq() / (((nxp.CCM.CBCDR.Get() & nxp.CCM_CBCDR_IPG_PODF_Msk) >> nxp.CCM_CBCDR_IPG_PODF_Pos) + 1)
}

// getPerClkFreq returns the PER clock frequency
func getPerClkFreq() uint32 {
	freq := uint32(0)
	if nxp.CCM.CSCMR1.HasBits(nxp.CCM_CSCMR1_PERCLK_CLK_SEL_Msk) {
		// Osc_clk -> PER Cloc
		freq = getOscFreq()
	} else {
		// Periph_clk -> AHB Clock -> IPG Clock -> PER Clock
		freq = getIpgFreq()
	}
	return freq/((nxp.CCM.CSCMR1.Get()&nxp.CCM_CSCMR1_PERCLK_PODF_Msk)>>
		nxp.CCM_CSCMR1_PERCLK_PODF_Pos) + 1
}

// getPllFreq returns the clock frequency of the receiver (PLL) clock clk
func (clk clock) getPllFreq() uint32 {
	enetRefClkFreq := []uint32{
		25000000,  // 25 MHz
		50000000,  // 50 MHz
		100000000, // 100 MHz
		125000000, // 125 MHz
	}
	// check if PLL is enabled
	if !clk.isPllEnabled() {
		return 0
	}
	// get pll reference clock
	freq := clk.getBypassFreq()
	// check if pll is bypassed
	if clk.isPllBypassed() {
		return freq
	}
	switch clk {
	case clkPllArm:
		freq *= (nxp.CCM_ANALOG.PLL_ARM.Get() & nxp.CCM_ANALOG_PLL_ARM_DIV_SELECT_Msk) >> nxp.CCM_ANALOG_PLL_ARM_DIV_SELECT_Pos
		freq >>= 1
	case clkPllSys:
		// PLL output frequency = Fref * (DIV_SELECT + NUM/DENOM).
		fFreq := float64(freq) * float64(nxp.CCM_ANALOG.PLL_SYS_NUM.Get())
		fFreq /= float64(nxp.CCM_ANALOG.PLL_SYS_DENOM.Get())
		if nxp.CCM_ANALOG.PLL_SYS.HasBits(nxp.CCM_ANALOG_PLL_SYS_DIV_SELECT_Msk) {
			freq *= 22
		} else {
			freq *= 20
		}
		freq += uint32(fFreq)
	case clkPllUsb1:
		if nxp.CCM_ANALOG.PLL_USB1.HasBits(nxp.CCM_ANALOG_PLL_USB1_DIV_SELECT_Msk) {
			freq *= 22
		} else {
			freq *= 20
		}
	case clkPllEnet:
		divSelect := (nxp.CCM_ANALOG.PLL_ENET.Get() & nxp.CCM_ANALOG_PLL_ENET_DIV_SELECT_Msk) >> nxp.CCM_ANALOG_PLL_ENET_DIV_SELECT_Pos
		freq = enetRefClkFreq[divSelect]
	case clkPllEnet2:
		divSelect := (nxp.CCM_ANALOG.PLL_ENET.Get() & nxp.CCM_ANALOG_PLL_ENET_ENET2_DIV_SELECT_Msk) >> nxp.CCM_ANALOG_PLL_ENET_ENET2_DIV_SELECT_Pos
		freq = enetRefClkFreq[divSelect]
	case clkPllEnet25M:
		// ref_enetpll1 if fixed at 25MHz.
		freq = 25000000
	case clkPllUsb2:
		if nxp.CCM_ANALOG.PLL_USB2.HasBits(nxp.CCM_ANALOG_PLL_USB2_DIV_SELECT_Msk) {
			freq *= 22
		} else {
			freq *= 20
		}
	default:
		freq = 0
	}
	return freq
}

// getSysPfdFreq returns current system PLL PFD output frequency
func (clk clock) getSysPfdFreq() uint32 {
	freq := clkPllSys.getPllFreq()
	switch clk {
	case clkPfd0:
		freq /= (nxp.CCM_ANALOG.PFD_528.Get() & nxp.CCM_ANALOG_PFD_528_PFD0_FRAC_Msk) >> nxp.CCM_ANALOG_PFD_528_PFD0_FRAC_Pos
	case clkPfd1:
		freq /= (nxp.CCM_ANALOG.PFD_528.Get() & nxp.CCM_ANALOG_PFD_528_PFD1_FRAC_Msk) >> nxp.CCM_ANALOG_PFD_528_PFD1_FRAC_Pos
	case clkPfd2:
		freq /= (nxp.CCM_ANALOG.PFD_528.Get() & nxp.CCM_ANALOG_PFD_528_PFD2_FRAC_Msk) >> nxp.CCM_ANALOG_PFD_528_PFD2_FRAC_Pos
	case clkPfd3:
		freq /= (nxp.CCM_ANALOG.PFD_528.Get() & nxp.CCM_ANALOG_PFD_528_PFD3_FRAC_Msk) >> nxp.CCM_ANALOG_PFD_528_PFD3_FRAC_Pos
	default:
		freq = 0
	}
	return freq * 18
}

// getUsb1PfdFreq returns current USB1 PLL PFD output frequency
func (clk clock) getUsb1PfdFreq() uint32 {
	freq := clkPllUsb1.getPllFreq()
	switch clk {
	case clkPfd0:
		freq /= (nxp.CCM_ANALOG.PFD_480.Get() & nxp.CCM_ANALOG_PFD_480_PFD0_FRAC_Msk) >> nxp.CCM_ANALOG_PFD_480_PFD0_FRAC_Pos
	case clkPfd1:
		freq /= (nxp.CCM_ANALOG.PFD_480.Get() & nxp.CCM_ANALOG_PFD_480_PFD1_FRAC_Msk) >> nxp.CCM_ANALOG_PFD_480_PFD1_FRAC_Pos
	case clkPfd2:
		freq /= (nxp.CCM_ANALOG.PFD_480.Get() & nxp.CCM_ANALOG_PFD_480_PFD2_FRAC_Msk) >> nxp.CCM_ANALOG_PFD_480_PFD2_FRAC_Pos
	case clkPfd3:
		freq /= (nxp.CCM_ANALOG.PFD_480.Get() & nxp.CCM_ANALOG_PFD_480_PFD3_FRAC_Msk) >> nxp.CCM_ANALOG_PFD_480_PFD3_FRAC_Pos
	default:
		freq = 0
	}
	return freq * 18
}

// getCCGR returns the CCM clock gating register for the receiver clk.
func (clk clock) getCCGR() *volatile.Register32 {
	switch clk >> 8 {
	case 0:
		return &nxp.CCM.CCGR0
	case 1:
		return &nxp.CCM.CCGR1
	case 2:
		return &nxp.CCM.CCGR2
	case 3:
		return &nxp.CCM.CCGR3
	case 4:
		return &nxp.CCM.CCGR4
	case 5:
		return &nxp.CCM.CCGR5
	case 6:
		return &nxp.CCM.CCGR6
	case 7:
		return &nxp.CCM.CCGR7
	default:
		panic("runtime: invalid clock")
	}
}

// control enables or disables the receiver clk using its gating register.
func (clk clock) control(value gate) {
	reg := clk.getCCGR()
	shift := clk & 0x1F
	reg.Set((reg.Get() & ^(3 << shift)) | (uint32(value) << shift))
}

func (clk clock) enable(enable bool) {
	if enable {
		clk.control(gateClkNeededRunWait)
	} else {
		clk.control(gateClkNotNeeded)
	}
}

func (clk clock) mux(mux uint32) { clk.ccm(mux) }
func (clk clock) div(div uint32) { clk.ccm(div) }

func (clk clock) ccm(value uint32) {
	const ccmBase = 0x400fc000
	reg := (*volatile.Register32)(unsafe.Pointer(uintptr(ccmBase + (uint32(clk) & 0xFF))))
	msk := ((uint32(clk) >> 13) & 0x1FFF) << ((uint32(clk) >> 8) & 0x1F)
	pos := (uint32(clk) >> 8) & 0x1F
	bsy := (uint32(clk) >> 26) & 0x3F
	reg.Set((reg.Get() & ^uint32(msk)) | ((value << pos) & msk))
	if bsy < noBusyWait {
		for nxp.CCM.CDHIPR.HasBits(1 << bsy) {
		}
	}
}

func setSysPfd(value ...uint32) {
	for i, val := range value {
		pfd528 := nxp.CCM_ANALOG.PFD_528.Get() &
			^((nxp.CCM_ANALOG_PFD_528_PFD0_CLKGATE_Msk | nxp.CCM_ANALOG_PFD_528_PFD0_FRAC_Msk) << (8 * uint32(i)))
		frac := (val << nxp.CCM_ANALOG_PFD_528_PFD0_FRAC_Pos) & nxp.CCM_ANALOG_PFD_528_PFD0_FRAC_Msk
		// disable the clock output first
		nxp.CCM_ANALOG.PFD_528.Set(pfd528 | (nxp.CCM_ANALOG_PFD_528_PFD0_CLKGATE_Msk << (8 * uint32(i))))
		// set the new value and enable output
		nxp.CCM_ANALOG.PFD_528.Set(pfd528 | (frac << (8 * uint32(i))))
	}
}

func setUsb1Pfd(value ...uint32) {
	for i, val := range value {
		pfd480 := nxp.CCM_ANALOG.PFD_480.Get() &
			^((nxp.CCM_ANALOG_PFD_480_PFD0_CLKGATE_Msk | nxp.CCM_ANALOG_PFD_480_PFD0_FRAC_Msk) << (8 * uint32(i)))
		frac := (val << nxp.CCM_ANALOG_PFD_480_PFD0_FRAC_Pos) & nxp.CCM_ANALOG_PFD_480_PFD0_FRAC_Msk
		// disable the clock output first
		nxp.CCM_ANALOG.PFD_480.Set(pfd480 | (nxp.CCM_ANALOG_PFD_480_PFD0_CLKGATE_Msk << (8 * uint32(i))))
		// set the new value and enable output
		nxp.CCM_ANALOG.PFD_480.Set(pfd480 | (frac << (8 * uint32(i))))
	}
}

func (clk clock) isPllEnabled() bool {
	const ccmAnalogBase = 0x400d8000
	addr := ccmAnalogBase + ((uint32(clk) >> 16) & 0xFFF)
	pos := uint32(1 << (uint32(clk) & 0x1F))
	return ((*volatile.Register32)(unsafe.Pointer(uintptr(addr)))).HasBits(pos)
}

func (clk clock) isPllBypassed() bool {
	const ccmAnalogBase = 0x400d8000
	addr := ccmAnalogBase + ((uint32(clk) >> 16) & 0xFFF)
	pos := uint32(1 << pllBypassPos)
	return ((*volatile.Register32)(unsafe.Pointer(uintptr(addr)))).HasBits(pos)
}

func (clk clock) getBypassFreq() uint32 {
	const ccmAnalogBase = 0x400d8000
	addr := ccmAnalogBase + ((uint32(clk) >> 16) & 0xFFF)
	src := (((*volatile.Register32)(unsafe.Pointer(uintptr(addr)))).Get() &
		pllBypassClkSrcMsk) >> pllBypassClkSrcPos
	if src == uint32(pllSrc24M) {
		return getOscFreq()
	}
	return 0
}

func (clk clock) bypass(bypass bool) {
	const ccmAnalogBase = 0x400d8000
	if bypass {
		addr := ccmAnalogBase + ((uint32(clk) >> 16) & 0xFFF) + 4
		((*volatile.Register32)(unsafe.Pointer(uintptr(addr)))).Set(1 << pllBypassPos)
	} else {
		addr := ccmAnalogBase + ((uint32(clk) >> 16) & 0xFFF) + 8
		((*volatile.Register32)(unsafe.Pointer(uintptr(addr)))).Set(1 << pllBypassPos)
	}
}

const (
	oscRc   = 0 // On chip OSC
	oscXtal = 1 // 24M Xtal OSC
)

type gate uint8

const (
	gateClkNotNeeded     gate = 0 // Clock is off during all modes
	gateClkNeededRun     gate = 1 // Clock is on in run mode, but off in WAIT and STOP modes
	gateClkNeededRunWait gate = 3 // Clock is on during all modes, except STOP mode
)

type clockMode uint8

const (
	modeClkRun  clockMode = 0 // Remain in run mode
	modeClkWait clockMode = 1 // Transfer to wait mode
	modeClkStop clockMode = 2 // Transfer to stop mode
)

func (m clockMode) set() {
	nxp.CCM.CLPCR.Set((nxp.CCM.CLPCR.Get() & ^uint32(nxp.CCM_CLPCR_LPM_Msk)) |
		((uint32(m) << nxp.CCM_CLPCR_LPM_Pos) & nxp.CCM_CLPCR_LPM_Msk))
}

const (
	muxIpPll3Sw     clock = (offCCSR & 0xFF) | (nxp.CCM_CCSR_PLL3_SW_CLK_SEL_Pos << 8) | (((nxp.CCM_CCSR_PLL3_SW_CLK_SEL_Msk >> nxp.CCM_CCSR_PLL3_SW_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                              // pll3_sw_clk mux name
	muxIpPeriph     clock = (offCBCDR & 0xFF) | (nxp.CCM_CBCDR_PERIPH_CLK_SEL_Pos << 8) | (((nxp.CCM_CBCDR_PERIPH_CLK_SEL_Msk >> nxp.CCM_CBCDR_PERIPH_CLK_SEL_Pos) & 0x1FFF) << 13) | (nxp.CCM_CDHIPR_PERIPH_CLK_SEL_BUSY_Pos << 26) // periph mux name
	muxIpSemcAlt    clock = (offCBCDR & 0xFF) | (nxp.CCM_CBCDR_SEMC_ALT_CLK_SEL_Pos << 8) | (((nxp.CCM_CBCDR_SEMC_ALT_CLK_SEL_Msk >> nxp.CCM_CBCDR_SEMC_ALT_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                       // semc mux name
	muxIpSemc       clock = (offCBCDR & 0xFF) | (nxp.CCM_CBCDR_SEMC_CLK_SEL_Pos << 8) | (((nxp.CCM_CBCDR_SEMC_CLK_SEL_Msk >> nxp.CCM_CBCDR_SEMC_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                                   // semc mux name
	muxIpPrePeriph  clock = (offCBCMR & 0xFF) | (nxp.CCM_CBCMR_PRE_PERIPH_CLK_SEL_Pos << 8) | (((nxp.CCM_CBCMR_PRE_PERIPH_CLK_SEL_Msk >> nxp.CCM_CBCMR_PRE_PERIPH_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                 // pre-periph mux name
	muxIpTrace      clock = (offCBCMR & 0xFF) | (nxp.CCM_CBCMR_TRACE_CLK_SEL_Pos << 8) | (((nxp.CCM_CBCMR_TRACE_CLK_SEL_Msk >> nxp.CCM_CBCMR_TRACE_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                                // trace mux name
	muxIpPeriphClk2 clock = (offCBCMR & 0xFF) | (nxp.CCM_CBCMR_PERIPH_CLK2_SEL_Pos << 8) | (((nxp.CCM_CBCMR_PERIPH_CLK2_SEL_Msk >> nxp.CCM_CBCMR_PERIPH_CLK2_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                          // periph clock2 mux name
	muxIpFlexSpi2   clock = (offCBCMR & 0xFF) | (nxp.CCM_CBCMR_FLEXSPI2_CLK_SEL_Pos << 8) | (((nxp.CCM_CBCMR_FLEXSPI2_CLK_SEL_Msk >> nxp.CCM_CBCMR_FLEXSPI2_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                       // flexspi2 mux name
	muxIpLpspi      clock = (offCBCMR & 0xFF) | (nxp.CCM_CBCMR_LPSPI_CLK_SEL_Pos << 8) | (((nxp.CCM_CBCMR_LPSPI_CLK_SEL_Msk >> nxp.CCM_CBCMR_LPSPI_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                                // lpspi mux name
	muxIpFlexSpi    clock = (offCSCMR1 & 0xFF) | (nxp.CCM_CSCMR1_FLEXSPI_CLK_SEL_Pos << 8) | (((nxp.CCM_CSCMR1_FLEXSPI_CLK_SEL_Msk >> nxp.CCM_CSCMR1_FLEXSPI_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                      // flexspi mux name
	muxIpUsdhc2     clock = (offCSCMR1 & 0xFF) | (nxp.CCM_CSCMR1_USDHC2_CLK_SEL_Pos << 8) | (((nxp.CCM_CSCMR1_USDHC2_CLK_SEL_Msk >> nxp.CCM_CSCMR1_USDHC2_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                         // usdhc2 mux name
	muxIpUsdhc1     clock = (offCSCMR1 & 0xFF) | (nxp.CCM_CSCMR1_USDHC1_CLK_SEL_Pos << 8) | (((nxp.CCM_CSCMR1_USDHC1_CLK_SEL_Msk >> nxp.CCM_CSCMR1_USDHC1_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                         // usdhc1 mux name
	muxIpSai3       clock = (offCSCMR1 & 0xFF) | (nxp.CCM_CSCMR1_SAI3_CLK_SEL_Pos << 8) | (((nxp.CCM_CSCMR1_SAI3_CLK_SEL_Msk >> nxp.CCM_CSCMR1_SAI3_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                               // sai3 mux name
	muxIpSai2       clock = (offCSCMR1 & 0xFF) | (nxp.CCM_CSCMR1_SAI2_CLK_SEL_Pos << 8) | (((nxp.CCM_CSCMR1_SAI2_CLK_SEL_Msk >> nxp.CCM_CSCMR1_SAI2_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                               // sai2 mux name
	muxIpSai1       clock = (offCSCMR1 & 0xFF) | (nxp.CCM_CSCMR1_SAI1_CLK_SEL_Pos << 8) | (((nxp.CCM_CSCMR1_SAI1_CLK_SEL_Msk >> nxp.CCM_CSCMR1_SAI1_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                               // sai1 mux name
	muxIpPerclk     clock = (offCSCMR1 & 0xFF) | (nxp.CCM_CSCMR1_PERCLK_CLK_SEL_Pos << 8) | (((nxp.CCM_CSCMR1_PERCLK_CLK_SEL_Msk >> nxp.CCM_CSCMR1_PERCLK_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                         // perclk mux name
	muxIpFlexio2    clock = (offCSCMR2 & 0xFF) | (nxp.CCM_CSCMR2_FLEXIO2_CLK_SEL_Pos << 8) | (((nxp.CCM_CSCMR2_FLEXIO2_CLK_SEL_Msk >> nxp.CCM_CSCMR2_FLEXIO2_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                      // flexio2 mux name
	muxIpCan        clock = (offCSCMR2 & 0xFF) | (nxp.CCM_CSCMR2_CAN_CLK_SEL_Pos << 8) | (((nxp.CCM_CSCMR2_CAN_CLK_SEL_Msk >> nxp.CCM_CSCMR2_CAN_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                                  // can mux name
	muxIpUart       clock = (offCSCDR1 & 0xFF) | (nxp.CCM_CSCDR1_UART_CLK_SEL_Pos << 8) | (((nxp.CCM_CSCDR1_UART_CLK_SEL_Msk >> nxp.CCM_CSCDR1_UART_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                               // uart mux name
	muxIpSpdif      clock = (offCDCDR & 0xFF) | (nxp.CCM_CDCDR_SPDIF0_CLK_SEL_Pos << 8) | (((nxp.CCM_CDCDR_SPDIF0_CLK_SEL_Msk >> nxp.CCM_CDCDR_SPDIF0_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                             // spdif mux name
	muxIpFlexio1    clock = (offCDCDR & 0xFF) | (nxp.CCM_CDCDR_FLEXIO1_CLK_SEL_Pos << 8) | (((nxp.CCM_CDCDR_FLEXIO1_CLK_SEL_Msk >> nxp.CCM_CDCDR_FLEXIO1_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                          // flexio1 mux name
	muxIpLpi2c      clock = (offCSCDR2 & 0xFF) | (nxp.CCM_CSCDR2_LPI2C_CLK_SEL_Pos << 8) | (((nxp.CCM_CSCDR2_LPI2C_CLK_SEL_Msk >> nxp.CCM_CSCDR2_LPI2C_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                            // lpi2c mux name
	muxIpLcdifPre   clock = (offCSCDR2 & 0xFF) | (nxp.CCM_CSCDR2_LCDIF_PRE_CLK_SEL_Pos << 8) | (((nxp.CCM_CSCDR2_LCDIF_PRE_CLK_SEL_Msk >> nxp.CCM_CSCDR2_LCDIF_PRE_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                // lcdif pre mux name
	muxIpCsi        clock = (offCSCDR3 & 0xFF) | (nxp.CCM_CSCDR3_CSI_CLK_SEL_Pos << 8) | (((nxp.CCM_CSCDR3_CSI_CLK_SEL_Msk >> nxp.CCM_CSCDR3_CSI_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                                  // csi mux name
)

const (
	divIpArm        clock = (offCACRR & 0xFF) | (nxp.CCM_CACRR_ARM_PODF_Pos << 8) | (((nxp.CCM_CACRR_ARM_PODF_Msk >> nxp.CCM_CACRR_ARM_PODF_Pos) & 0x1FFF) << 13) | (nxp.CCM_CDHIPR_ARM_PODF_BUSY_Pos << 26)       // core div name
	divIpPeriphClk2 clock = (offCBCDR & 0xFF) | (nxp.CCM_CBCDR_PERIPH_CLK2_PODF_Pos << 8) | (((nxp.CCM_CBCDR_PERIPH_CLK2_PODF_Msk >> nxp.CCM_CBCDR_PERIPH_CLK2_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)     // periph clock2 div name
	divIpSemc       clock = (offCBCDR & 0xFF) | (nxp.CCM_CBCDR_SEMC_PODF_Pos << 8) | (((nxp.CCM_CBCDR_SEMC_PODF_Msk >> nxp.CCM_CBCDR_SEMC_PODF_Pos) & 0x1FFF) << 13) | (nxp.CCM_CDHIPR_SEMC_PODF_BUSY_Pos << 26)   // semc div name
	divIpAhb        clock = (offCBCDR & 0xFF) | (nxp.CCM_CBCDR_AHB_PODF_Pos << 8) | (((nxp.CCM_CBCDR_AHB_PODF_Msk >> nxp.CCM_CBCDR_AHB_PODF_Pos) & 0x1FFF) << 13) | (nxp.CCM_CDHIPR_AHB_PODF_BUSY_Pos << 26)       // ahb div name
	divIpIpg        clock = (offCBCDR & 0xFF) | (nxp.CCM_CBCDR_IPG_PODF_Pos << 8) | (((nxp.CCM_CBCDR_IPG_PODF_Msk >> nxp.CCM_CBCDR_IPG_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                             // ipg div name
	divIpFlexSpi2   clock = (offCBCMR & 0xFF) | (nxp.CCM_CBCMR_FLEXSPI2_PODF_Pos << 8) | (((nxp.CCM_CBCMR_FLEXSPI2_PODF_Msk >> nxp.CCM_CBCMR_FLEXSPI2_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)              // flexspi2 div name
	divIpLpspi      clock = (offCBCMR & 0xFF) | (nxp.CCM_CBCMR_LPSPI_PODF_Pos << 8) | (((nxp.CCM_CBCMR_LPSPI_PODF_Msk >> nxp.CCM_CBCMR_LPSPI_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                       // lpspi div name
	divIpLcdif      clock = (offCBCMR & 0xFF) | (nxp.CCM_CBCMR_LCDIF_PODF_Pos << 8) | (((nxp.CCM_CBCMR_LCDIF_PODF_Msk >> nxp.CCM_CBCMR_LCDIF_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                       // lcdif div name
	divIpFlexSpi    clock = (offCSCMR1 & 0xFF) | (nxp.CCM_CSCMR1_FLEXSPI_PODF_Pos << 8) | (((nxp.CCM_CSCMR1_FLEXSPI_PODF_Msk >> nxp.CCM_CSCMR1_FLEXSPI_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)             // flexspi div name
	divIpPerclk     clock = (offCSCMR1 & 0xFF) | (nxp.CCM_CSCMR1_PERCLK_PODF_Pos << 8) | (((nxp.CCM_CSCMR1_PERCLK_PODF_Msk >> nxp.CCM_CSCMR1_PERCLK_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                // perclk div name
	divIpCan        clock = (offCSCMR2 & 0xFF) | (nxp.CCM_CSCMR2_CAN_CLK_PODF_Pos << 8) | (((nxp.CCM_CSCMR2_CAN_CLK_PODF_Msk >> nxp.CCM_CSCMR2_CAN_CLK_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)             // can div name
	divIpTrace      clock = (offCSCDR1 & 0xFF) | (nxp.CCM_CSCDR1_TRACE_PODF_Pos << 8) | (((nxp.CCM_CSCDR1_TRACE_PODF_Msk >> nxp.CCM_CSCDR1_TRACE_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                   // trace div name
	divIpUsdhc2     clock = (offCSCDR1 & 0xFF) | (nxp.CCM_CSCDR1_USDHC2_PODF_Pos << 8) | (((nxp.CCM_CSCDR1_USDHC2_PODF_Msk >> nxp.CCM_CSCDR1_USDHC2_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                // usdhc2 div name
	divIpUsdhc1     clock = (offCSCDR1 & 0xFF) | (nxp.CCM_CSCDR1_USDHC1_PODF_Pos << 8) | (((nxp.CCM_CSCDR1_USDHC1_PODF_Msk >> nxp.CCM_CSCDR1_USDHC1_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                // usdhc1 div name
	divIpUart       clock = (offCSCDR1 & 0xFF) | (nxp.CCM_CSCDR1_UART_CLK_PODF_Pos << 8) | (((nxp.CCM_CSCDR1_UART_CLK_PODF_Msk >> nxp.CCM_CSCDR1_UART_CLK_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)          // uart div name
	divIpFlexio2    clock = (offCS1CDR & 0xFF) | (nxp.CCM_CS1CDR_FLEXIO2_CLK_PODF_Pos << 8) | (((nxp.CCM_CS1CDR_FLEXIO2_CLK_PODF_Msk >> nxp.CCM_CS1CDR_FLEXIO2_CLK_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26) // flexio2 pre div name
	divIpSai3Pre    clock = (offCS1CDR & 0xFF) | (nxp.CCM_CS1CDR_SAI3_CLK_PRED_Pos << 8) | (((nxp.CCM_CS1CDR_SAI3_CLK_PRED_Msk >> nxp.CCM_CS1CDR_SAI3_CLK_PRED_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)          // sai3 pre div name
	divIpSai3       clock = (offCS1CDR & 0xFF) | (nxp.CCM_CS1CDR_SAI3_CLK_PODF_Pos << 8) | (((nxp.CCM_CS1CDR_SAI3_CLK_PODF_Msk >> nxp.CCM_CS1CDR_SAI3_CLK_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)          // sai3 div name
	divIpFlexio2Pre clock = (offCS1CDR & 0xFF) | (nxp.CCM_CS1CDR_FLEXIO2_CLK_PRED_Pos << 8) | (((nxp.CCM_CS1CDR_FLEXIO2_CLK_PRED_Msk >> nxp.CCM_CS1CDR_FLEXIO2_CLK_PRED_Pos) & 0x1FFF) << 13) | (noBusyWait << 26) // sai3 pre div name
	divIpSai1Pre    clock = (offCS1CDR & 0xFF) | (nxp.CCM_CS1CDR_SAI1_CLK_PRED_Pos << 8) | (((nxp.CCM_CS1CDR_SAI1_CLK_PRED_Msk >> nxp.CCM_CS1CDR_SAI1_CLK_PRED_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)          // sai1 pre div name
	divIpSai1       clock = (offCS1CDR & 0xFF) | (nxp.CCM_CS1CDR_SAI1_CLK_PODF_Pos << 8) | (((nxp.CCM_CS1CDR_SAI1_CLK_PODF_Msk >> nxp.CCM_CS1CDR_SAI1_CLK_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)          // sai1 div name
	divIpSai2Pre    clock = (offCS2CDR & 0xFF) | (nxp.CCM_CS2CDR_SAI2_CLK_PRED_Pos << 8) | (((nxp.CCM_CS2CDR_SAI2_CLK_PRED_Msk >> nxp.CCM_CS2CDR_SAI2_CLK_PRED_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)          // sai2 pre div name
	divIpSai2       clock = (offCS2CDR & 0xFF) | (nxp.CCM_CS2CDR_SAI2_CLK_PODF_Pos << 8) | (((nxp.CCM_CS2CDR_SAI2_CLK_PODF_Msk >> nxp.CCM_CS2CDR_SAI2_CLK_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)          // sai2 div name
	divIpSpdif0Pre  clock = (offCDCDR & 0xFF) | (nxp.CCM_CDCDR_SPDIF0_CLK_PRED_Pos << 8) | (((nxp.CCM_CDCDR_SPDIF0_CLK_PRED_Msk >> nxp.CCM_CDCDR_SPDIF0_CLK_PRED_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)        // spdif pre div name
	divIpSpdif0     clock = (offCDCDR & 0xFF) | (nxp.CCM_CDCDR_SPDIF0_CLK_PODF_Pos << 8) | (((nxp.CCM_CDCDR_SPDIF0_CLK_PODF_Msk >> nxp.CCM_CDCDR_SPDIF0_CLK_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)        // spdif div name
	divIpFlexio1Pre clock = (offCDCDR & 0xFF) | (nxp.CCM_CDCDR_FLEXIO1_CLK_PRED_Pos << 8) | (((nxp.CCM_CDCDR_FLEXIO1_CLK_PRED_Msk >> nxp.CCM_CDCDR_FLEXIO1_CLK_PRED_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)     // flexio1 pre div name
	divIpFlexio1    clock = (offCDCDR & 0xFF) | (nxp.CCM_CDCDR_FLEXIO1_CLK_PODF_Pos << 8) | (((nxp.CCM_CDCDR_FLEXIO1_CLK_PODF_Msk >> nxp.CCM_CDCDR_FLEXIO1_CLK_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)     // flexio1 div name
	divIpLpi2c      clock = (offCSCDR2 & 0xFF) | (nxp.CCM_CSCDR2_LPI2C_CLK_PODF_Pos << 8) | (((nxp.CCM_CSCDR2_LPI2C_CLK_PODF_Msk >> nxp.CCM_CSCDR2_LPI2C_CLK_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)       // lpi2c div name
	divIpLcdifPre   clock = (offCSCDR2 & 0xFF) | (nxp.CCM_CSCDR2_LCDIF_PRED_Pos << 8) | (((nxp.CCM_CSCDR2_LCDIF_PRED_Msk >> nxp.CCM_CSCDR2_LCDIF_PRED_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                   // lcdif pre div name
	divIpCsi        clock = (offCSCDR3 & 0xFF) | (nxp.CCM_CSCDR3_CSI_PODF_Pos << 8) | (((nxp.CCM_CSCDR3_CSI_PODF_Msk >> nxp.CCM_CSCDR3_CSI_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                         // csi div name
)

// USB clock source
const (
	usbSrc480M   = 0          // Use 480M
	usbSrcUnused = 0xFFFFFFFF // Used when the function does not care the clock source
)

// Source of the USB HS PHY
const (
	usbhsPhy480M = 0 // Use 480M
)

// PLL clock source, bypass cloco source also
const (
	pllSrc24M   = 0 // Pll clock source 24M
	pllSrcClkPN = 1 // Pll clock source CLK1_P and CLK1_N
)

// PLL configuration for ARM
type clockConfigArmPll struct {
	loopDivider uint32 // PLL loop divider. Valid range for divider value: 54-108. Fout=Fin*loopDivider/2.
	src         uint8  // Pll clock source, reference _clock_pll_clk_src
}

func (cfg clockConfigArmPll) configure() {

	// bypass PLL first
	src := (uint32(cfg.src) << nxp.CCM_ANALOG_PLL_ARM_BYPASS_CLK_SRC_Pos) & nxp.CCM_ANALOG_PLL_ARM_BYPASS_CLK_SRC_Msk
	nxp.CCM_ANALOG.PLL_ARM.Set(
		(nxp.CCM_ANALOG.PLL_ARM.Get() & ^uint32(nxp.CCM_ANALOG_PLL_ARM_BYPASS_CLK_SRC_Msk)) |
			nxp.CCM_ANALOG_PLL_ARM_BYPASS_Msk | src)

	sel := (cfg.loopDivider << nxp.CCM_ANALOG_PLL_ARM_DIV_SELECT_Pos) & nxp.CCM_ANALOG_PLL_ARM_DIV_SELECT_Msk
	nxp.CCM_ANALOG.PLL_ARM.Set(
		(nxp.CCM_ANALOG.PLL_ARM.Get() & ^uint32(nxp.CCM_ANALOG_PLL_ARM_DIV_SELECT_Msk|nxp.CCM_ANALOG_PLL_ARM_POWERDOWN_Msk)) |
			nxp.CCM_ANALOG_PLL_ARM_ENABLE_Msk | sel)

	for !nxp.CCM_ANALOG.PLL_ARM.HasBits(nxp.CCM_ANALOG_PLL_ARM_LOCK_Msk) {
	}

	// disable bypass
	nxp.CCM_ANALOG.PLL_ARM.ClearBits(nxp.CCM_ANALOG_PLL_ARM_BYPASS_Msk)
}

// PLL configuration for System
type clockConfigSysPll struct {
	loopDivider uint8  // PLL loop divider. Intended to be 1 (528M): 0 - Fout=Fref*20, 1 - Fout=Fref*22
	numerator   uint32 // 30 bit numerator of fractional loop divider.
	denominator uint32 // 30 bit denominator of fractional loop divider
	src         uint8  // Pll clock source, reference _clock_pll_clk_src
	ssStop      uint16 // Stop value to get frequency change.
	ssEnable    uint8  // Enable spread spectrum modulation
	ssStep      uint16 // Step value to get frequency change step.
}

func (cfg clockConfigSysPll) configure(pfd ...uint32) {

	// bypass PLL first
	src := (uint32(cfg.src) << nxp.CCM_ANALOG_PLL_SYS_BYPASS_CLK_SRC_Pos) & nxp.CCM_ANALOG_PLL_SYS_BYPASS_CLK_SRC_Msk
	nxp.CCM_ANALOG.PLL_SYS.Set(
		(nxp.CCM_ANALOG.PLL_SYS.Get() & ^uint32(nxp.CCM_ANALOG_PLL_SYS_BYPASS_CLK_SRC_Msk)) |
			nxp.CCM_ANALOG_PLL_SYS_BYPASS_Msk | src)

	sel := (uint32(cfg.loopDivider) << nxp.CCM_ANALOG_PLL_SYS_DIV_SELECT_Pos) & nxp.CCM_ANALOG_PLL_SYS_DIV_SELECT_Msk
	nxp.CCM_ANALOG.PLL_SYS.Set(
		(nxp.CCM_ANALOG.PLL_SYS.Get() & ^uint32(nxp.CCM_ANALOG_PLL_SYS_DIV_SELECT_Msk|nxp.CCM_ANALOG_PLL_SYS_POWERDOWN_Msk)) |
			nxp.CCM_ANALOG_PLL_SYS_ENABLE_Msk | sel)

	// initialize the fractional mode
	nxp.CCM_ANALOG.PLL_SYS_NUM.Set((cfg.numerator << nxp.CCM_ANALOG_PLL_SYS_NUM_A_Pos) & nxp.CCM_ANALOG_PLL_SYS_NUM_A_Msk)
	nxp.CCM_ANALOG.PLL_SYS_DENOM.Set((cfg.denominator << nxp.CCM_ANALOG_PLL_SYS_DENOM_B_Pos) & nxp.CCM_ANALOG_PLL_SYS_DENOM_B_Msk)

	// initialize the spread spectrum mode
	inc := (uint32(cfg.ssStep) << nxp.CCM_ANALOG_PLL_SYS_SS_STEP_Pos) & nxp.CCM_ANALOG_PLL_SYS_SS_STEP_Msk
	enb := (uint32(cfg.ssEnable) << nxp.CCM_ANALOG_PLL_SYS_SS_ENABLE_Pos) & nxp.CCM_ANALOG_PLL_SYS_SS_ENABLE_Msk
	stp := (uint32(cfg.ssStop) << nxp.CCM_ANALOG_PLL_SYS_SS_STOP_Pos) & nxp.CCM_ANALOG_PLL_SYS_SS_STOP_Msk
	nxp.CCM_ANALOG.PLL_SYS_SS.Set(inc | enb | stp)

	for !nxp.CCM_ANALOG.PLL_SYS.HasBits(nxp.CCM_ANALOG_PLL_SYS_LOCK_Msk) {
	}

	// disable bypass
	nxp.CCM_ANALOG.PLL_SYS.ClearBits(nxp.CCM_ANALOG_PLL_SYS_BYPASS_Msk)

	// update PFDs after update
	setSysPfd(pfd...)
}

// PLL configuration for USB
type clockConfigUsbPll struct {
	instance    uint8 // USB PLL number (1 or 2)
	loopDivider uint8 // PLL loop divider: 0 - Fout=Fref*20, 1 - Fout=Fref*22
	src         uint8 // Pll clock source, reference _clock_pll_clk_src
}

func (cfg clockConfigUsbPll) configure(pfd ...uint32) {

	switch cfg.instance {
	case 1:

		// bypass PLL first
		src := (uint32(cfg.src) << nxp.CCM_ANALOG_PLL_USB1_BYPASS_CLK_SRC_Pos) & nxp.CCM_ANALOG_PLL_USB1_BYPASS_CLK_SRC_Msk
		nxp.CCM_ANALOG.PLL_USB1.Set(
			(nxp.CCM_ANALOG.PLL_USB1.Get() & ^uint32(nxp.CCM_ANALOG_PLL_USB1_BYPASS_CLK_SRC_Msk)) |
				nxp.CCM_ANALOG_PLL_USB1_BYPASS_Msk | src)

		sel := uint32((cfg.loopDivider << nxp.CCM_ANALOG_PLL_USB1_DIV_SELECT_Pos) & nxp.CCM_ANALOG_PLL_USB1_DIV_SELECT_Msk)
		nxp.CCM_ANALOG.PLL_USB1_SET.Set(
			(nxp.CCM_ANALOG.PLL_USB1.Get() & ^uint32(nxp.CCM_ANALOG_PLL_USB1_DIV_SELECT_Msk)) |
				nxp.CCM_ANALOG_PLL_USB1_ENABLE_Msk | nxp.CCM_ANALOG_PLL_USB1_POWER_Msk |
				nxp.CCM_ANALOG_PLL_USB1_EN_USB_CLKS_Msk | sel)

		for !nxp.CCM_ANALOG.PLL_USB1.HasBits(nxp.CCM_ANALOG_PLL_USB1_LOCK_Msk) {
		}

		// disable bypass
		nxp.CCM_ANALOG.PLL_USB1_CLR.Set(nxp.CCM_ANALOG_PLL_USB1_BYPASS_Msk)

		// update PFDs after update
		setUsb1Pfd(pfd...)

	case 2:
		// bypass PLL first
		src := (uint32(cfg.src) << nxp.CCM_ANALOG_PLL_USB2_BYPASS_CLK_SRC_Pos) & nxp.CCM_ANALOG_PLL_USB2_BYPASS_CLK_SRC_Msk
		nxp.CCM_ANALOG.PLL_USB2.Set(
			(nxp.CCM_ANALOG.PLL_USB2.Get() & ^uint32(nxp.CCM_ANALOG_PLL_USB2_BYPASS_CLK_SRC_Msk)) |
				nxp.CCM_ANALOG_PLL_USB2_BYPASS_Msk | src)

		sel := uint32((cfg.loopDivider << nxp.CCM_ANALOG_PLL_USB2_DIV_SELECT_Pos) & nxp.CCM_ANALOG_PLL_USB2_DIV_SELECT_Msk)
		nxp.CCM_ANALOG.PLL_USB2.Set(
			(nxp.CCM_ANALOG.PLL_USB2.Get() & ^uint32(nxp.CCM_ANALOG_PLL_USB2_DIV_SELECT_Msk)) |
				nxp.CCM_ANALOG_PLL_USB2_ENABLE_Msk | nxp.CCM_ANALOG_PLL_USB2_POWER_Msk |
				nxp.CCM_ANALOG_PLL_USB2_EN_USB_CLKS_Msk | sel)

		for !nxp.CCM_ANALOG.PLL_USB2.HasBits(nxp.CCM_ANALOG_PLL_USB2_LOCK_Msk) {
		}

		// disable bypass
		nxp.CCM_ANALOG.PLL_USB2.ClearBits(nxp.CCM_ANALOG_PLL_USB2_BYPASS_Msk)

	default:
		panic("runtime: invalid USB PLL")
	}
}

// PLL configuration for AUDIO
type clockConfigAudioPll struct {
	loopDivider uint8  // PLL loop divider. Valid range for DIV_SELECT divider value: 27~54.
	postDivider uint8  // Divider after the PLL, should only be 1, 2, 4, 8, 16.
	numerator   uint32 // 30 bit numerator of fractional loop divider.
	denominator uint32 // 30 bit denominator of fractional loop divider
	src         uint8  // Pll clock source, reference _clock_pll_clk_src
}

// PLL configuration for VIDEO
type clockConfigVideoPll struct {
	loopDivider uint8  // PLL loop divider. Valid range for DIV_SELECT divider value: 27~54.
	postDivider uint8  // Divider after the PLL, should only be 1, 2, 4, 8, 16.
	numerator   uint32 // 30 bit numerator of fractional loop divider.
	denominator uint32 // 30 bit denominator of fractional loop divider
	src         uint8  // Pll clock source, reference _clock_pll_clk_src
}

// PLL configuration for ENET
type clockConfigEnetPll struct {
	enableClkOutput    bool  // Power on and enable PLL clock output for ENET0.
	enableClkOutput25M bool  // Power on and enable PLL clock output for ENET2.
	loopDivider        uint8 // Controls the frequency of the ENET0 reference clock: b00=25MHz, b01=50MHz, b10=100MHz (not 50% duty cycle), b11=125MHz
	src                uint8 // Pll clock source, reference _clock_pll_clk_src
	enableClkOutput1   bool  // Power on and enable PLL clock output for ENET1.
	loopDivider1       uint8 // Controls the frequency of the ENET1 reference clock: b00 25MHz, b01 50MHz, b10 100MHz (not 50% duty cycle), b11 125MHz
}

// PLL name
const (
	clkPllArm     clock = ((offPllArm & 0xFFF) << 16) | nxp.CCM_ANALOG_PLL_ARM_ENABLE_Pos            // PLL ARM
	clkPllSys     clock = ((offPllSys & 0xFFF) << 16) | nxp.CCM_ANALOG_PLL_SYS_ENABLE_Pos            // PLL SYS
	clkPllUsb1    clock = ((offPllUsb1 & 0xFFF) << 16) | nxp.CCM_ANALOG_PLL_USB1_ENABLE_Pos          // PLL USB1
	clkPllAudio   clock = ((offPllAudio & 0xFFF) << 16) | nxp.CCM_ANALOG_PLL_AUDIO_ENABLE_Pos        // PLL Audio
	clkPllVideo   clock = ((offPllVideo & 0xFFF) << 16) | nxp.CCM_ANALOG_PLL_VIDEO_ENABLE_Pos        // PLL Video
	clkPllEnet    clock = ((offPllEnet & 0xFFF) << 16) | nxp.CCM_ANALOG_PLL_ENET_ENABLE_Pos          // PLL Enet0
	clkPllEnet2   clock = ((offPllEnet & 0xFFF) << 16) | nxp.CCM_ANALOG_PLL_ENET_ENET2_REF_EN_Pos    // PLL Enet1
	clkPllEnet25M clock = ((offPllEnet & 0xFFF) << 16) | nxp.CCM_ANALOG_PLL_ENET_ENET_25M_REF_EN_Pos // PLL Enet2
	clkPllUsb2    clock = ((offPllUsb2 & 0xFFF) << 16) | nxp.CCM_ANALOG_PLL_USB2_ENABLE_Pos          // PLL USB2
)

// PLL PFD name
const (
	clkPfd0 clock = 0 // PLL PFD0
	clkPfd1 clock = 1 // PLL PFD1
	clkPfd2 clock = 2 // PLL PFD2
	clkPfd3 clock = 3 // PLL PFD3
)

var (
	armPllConfig = clockConfigArmPll{
		loopDivider: 100, // PLL loop divider, Fout = Fin * 50
		src:         0,   // Bypass clock source, 0 - OSC 24M, 1 - CLK1_P and CLK1_N
	}
	sysPllConfig = clockConfigSysPll{
		loopDivider: 1, // PLL loop divider, Fout = Fin * ( 20 + loopDivider*2 + numerator / denominator )
		numerator:   0, // 30 bit numerator of fractional loop divider
		denominator: 1, // 30 bit denominator of fractional loop divider
		src:         0, // Bypass clock source, 0 - OSC 24M, 1 - CLK1_P and CLK1_N
	}
	usb1PllConfig = clockConfigUsbPll{
		instance:    1, // USB PLL instance
		loopDivider: 0, // PLL loop divider, Fout = Fin * 20
		src:         0, // Bypass clock source, 0 - OSC 24M, 1 - CLK1_P and CLK1_N
	}
	usb2PllConfig = clockConfigUsbPll{
		instance:    2, // USB PLL instance
		loopDivider: 0, // PLL loop divider, Fout = Fin * 20
		src:         0, // Bypass clock source, 0 - OSC 24M, 1 - CLK1_P and CLK1_N
	}
	videoPllConfig = clockConfigVideoPll{
		loopDivider: 31, // PLL loop divider, Fout = Fin * ( loopDivider + numerator / denominator )
		postDivider: 8,  // Divider after PLL
		numerator:   0,  // 30 bit numerator of fractional loop divider, Fout = Fin * ( loopDivider + numerator / denominator )
		denominator: 1,  // 30 bit denominator of fractional loop divider, Fout = Fin * ( loopDivider + numerator / denominator )
		src:         0,  // Bypass clock source, 0 - OSC 24M, 1 - CLK1_P and CLK1_N
	}
)
