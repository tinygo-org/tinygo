// +build mimxrt1062

package runtime

import (
	"device/nxp"
)

// Core, bus, and peripheral clock frequencies (Hz)
const (
	CORE_FREQ      = 600000000 // 600     MHz
	SYSTICK_FREQ   = 100000    // 100     kHz  (see note below)
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

var (
	ArmPllConfig = nxp.ClockConfigArmPll{
		LoopDivider: 100, // PLL loop divider, Fout = Fin * 50
		Src:         0,   // Bypass clock source, 0 - OSC 24M, 1 - CLK1_P and CLK1_N
	}
	SysPllConfig = nxp.ClockConfigSysPll{
		LoopDivider: 1, // PLL loop divider, Fout = Fin * ( 20 + LoopDivider*2 + Numerator / Denominator )
		Numerator:   0, // 30 bit Numerator of fractional loop divider
		Denominator: 1, // 30 bit Denominator of fractional loop divider
		Src:         0, // Bypass clock source, 0 - OSC 24M, 1 - CLK1_P and CLK1_N
	}
	Usb1PllConfig = nxp.ClockConfigUsbPll{
		Instance:    1, // USB PLL Instance
		LoopDivider: 0, // PLL loop divider, Fout = Fin * 20
		Src:         0, // Bypass clock source, 0 - OSC 24M, 1 - CLK1_P and CLK1_N
	}
	Usb2PllConfig = nxp.ClockConfigUsbPll{
		Instance:    2, // USB PLL Instance
		LoopDivider: 0, // PLL loop divider, Fout = Fin * 20
		Src:         0, // Bypass clock source, 0 - OSC 24M, 1 - CLK1_P and CLK1_N
	}
	VideoPllConfig = nxp.ClockConfigVideoPll{
		LoopDivider: 31, // PLL loop divider, Fout = Fin * ( LoopDivider + Numerator / Denominator )
		PostDivider: 8,  // Divider after PLL
		Numerator:   0,  // 30 bit Numerator of fractional loop divider, Fout = Fin * ( LoopDivider + Numerator / Denominator )
		Denominator: 1,  // 30 bit Denominator of fractional loop divider, Fout = Fin * ( LoopDivider + Numerator / Denominator )
		Src:         0,  // Bypass clock source, 0 - OSC 24M, 1 - CLK1_P and CLK1_N
	}
)

// Note about SYSTICK_FREQ from Teensyduino (cores/teensy4/startup.c):
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
// disabled prior to configuration and must be enabled afterwards using one of
// these `enable*Clocks()` functions or the respective peripheral clocks'
// `Enable()` method from the "device/nxp" package.
func initClocks() {
	// disable low-power mode so that __WFI doesn't lock up at runtime.
	// see: Using the MIMXRT1060/4-EVK with MCUXpresso IDE v10.3.x (v1.0.2,
	// 2019MAR01), chapter 14
	nxp.ModeClkRun.Set()

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
	nxp.MuxIpPeriphClk2.Mux(1) // PERIPH_CLK2 select OSC24M
	nxp.MuxIpPeriph.Mux(1)     // PERIPH select PERIPH_CLK2

	// set VDD_SOC to 1.275V, necessary to config AHB to 600 MHz
	nxp.DCDC.REG3.Set((nxp.DCDC.REG3.Get() & ^uint32(nxp.DCDC_REG3_TRG_Msk)) |
		((13 << nxp.DCDC_REG3_TRG_Pos) & nxp.DCDC_REG3_TRG_Msk))

	// wait until DCDC_STS_DC_OK bit is asserted
	for !nxp.DCDC.REG0.HasBits(nxp.DCDC_REG0_STS_DC_OK_Msk) {
	}

	// -------------------------------------------------------------------- AHB --
	nxp.DivIpAhb.Div(0) // divide AHB_PODF (DIV1)

	// -------------------------------------------------------------------- ADC --
	nxp.ClockIpAdc1.Enable(false) // disable ADC clock gates
	nxp.ClockIpAdc2.Enable(false) //  ~

	// ------------------------------------------------------------------- XBAR --
	nxp.ClockIpXbar1.Enable(false) //  disable XBAR clock gates
	nxp.ClockIpXbar2.Enable(false) //  ~
	nxp.ClockIpXbar3.Enable(false) //  ~

	// ---------------------------------------------------------------- ARM/IPG --
	nxp.DivIpIpg.Div(3)        // divide IPG_PODF (DIV4)
	nxp.DivIpArm.Div(1)        // divide ARM_PODF (DIV2)
	nxp.DivIpPeriphClk2.Div(0) // divide PERIPH_CLK2_PODF (DIV1)

	// ---------------------------------------------------------------- GPT/PIT --
	nxp.ClockIpGpt1.Enable(false)  // disable GPT/PIT clock gates
	nxp.ClockIpGpt1S.Enable(false) //  ~
	nxp.ClockIpGpt2.Enable(false)  //  ~
	nxp.ClockIpGpt2S.Enable(false) //  ~
	nxp.ClockIpPit.Enable(false)   //  ~

	// -------------------------------------------------------------------- PER --
	nxp.DivIpPerclk.Div(0) // divide PERCLK_PODF (DIV1)

	// ------------------------------------------------------------------ USDHC --
	nxp.ClockIpUsdhc1.Enable(false) // disable USDHC1 clock gate
	nxp.DivIpUsdhc1.Div(1)          // divide USDHC1_PODF (DIV2)
	nxp.MuxIpUsdhc1.Mux(1)          // USDHC1 select PLL2_PFD0
	nxp.ClockIpUsdhc2.Enable(false) // disable USDHC2 clock gate
	nxp.DivIpUsdhc2.Div(1)          // divide USDHC2_PODF (DIV2)
	nxp.MuxIpUsdhc2.Mux(1)          // USDHC2 select PLL2_PFD0

	// ------------------------------------------------------------------- SEMC --
	nxp.ClockIpSemc.Enable(false) // disable SEMC clock gate
	nxp.DivIpSemc.Div(1)          // divide SEMC_PODF (DIV2)
	nxp.MuxIpSemcAlt.Mux(0)       // SEMC_ALT select PLL2_PFD2
	nxp.MuxIpSemc.Mux(1)          // SEMC select SEMC_ALT

	// ---------------------------------------------------------------- FLEXSPI --
	if false {
		// TODO: external flash is on this bus, configured via DCD block
		nxp.ClockIpFlexSpi.Enable(false) // disable FLEXSPI clock gate
		nxp.DivIpFlexSpi.Div(0)          // divide FLEXSPI_PODF (DIV1)
		nxp.MuxIpFlexSpi.Mux(2)          // FLEXSPI select PLL2_PFD2
	}
	nxp.ClockIpFlexSpi2.Enable(false) // disable FLEXSPI2 clock gate
	nxp.DivIpFlexSpi2.Div(0)          // divide FLEXSPI2_PODF (DIV1)
	nxp.MuxIpFlexSpi2.Mux(0)          // FLEXSPI2 select PLL2_PFD2

	// -------------------------------------------------------------------- CSI --
	nxp.ClockIpCsi.Enable(false) // disable CSI clock gate
	nxp.DivIpCsi.Div(1)          // divide CSI_PODF (DIV2)
	nxp.MuxIpCsi.Mux(0)          // CSI select OSC24M

	// ------------------------------------------------------------------ LPSPI --
	nxp.ClockIpLpspi1.Enable(false) // disable LPSPI clock gate
	nxp.ClockIpLpspi2.Enable(false) //  ~
	nxp.ClockIpLpspi3.Enable(false) //  ~
	nxp.ClockIpLpspi4.Enable(false) //  ~
	nxp.DivIpLpspi.Div(3)           // divide LPSPI_PODF (DIV4)
	nxp.MuxIpLpspi.Mux(2)           // LPSPI select PLL2

	// ------------------------------------------------------------------ TRACE --
	nxp.ClockIpTrace.Enable(false) // disable TRACE clock gate
	nxp.DivIpTrace.Div(3)          // divide TRACE_PODF (DIV4)
	nxp.MuxIpTrace.Mux(0)          // TRACE select PLL2_MAIN

	// -------------------------------------------------------------------- SAI --
	nxp.ClockIpSai1.Enable(false) // disable SAI1 clock gate
	nxp.DivIpSai1Pre.Div(3)       // divide SAI1_CLK_PRED (DIV4)
	nxp.DivIpSai1.Div(1)          // divide SAI1_CLK_PODF (DIV2)
	nxp.MuxIpSai1.Mux(0)          // SAI1 select PLL3_PFD2
	nxp.ClockIpSai2.Enable(false) // disable SAI2 clock gate
	nxp.DivIpSai2Pre.Div(3)       // divide SAI2_CLK_PRED (DIV4)
	nxp.DivIpSai2.Div(1)          // divide SAI2_CLK_PODF (DIV2)
	nxp.MuxIpSai2.Mux(0)          // SAI2 select PLL3_PFD2
	nxp.ClockIpSai3.Enable(false) // disable SAI3 clock gate
	nxp.DivIpSai3Pre.Div(3)       // divide SAI3_CLK_PRED (DIV4)
	nxp.DivIpSai3.Div(1)          // divide SAI3_CLK_PODF (DIV2)
	nxp.MuxIpSai3.Mux(0)          // SAI3 select PLL3_PFD2

	// ------------------------------------------------------------------ LPI2C --
	nxp.ClockIpLpi2c1.Enable(false) // disable LPI2C clock gate
	nxp.ClockIpLpi2c2.Enable(false) //  ~
	nxp.ClockIpLpi2c3.Enable(false) //  ~
	nxp.DivIpLpi2c.Div(0)           // divide LPI2C_CLK_PODF (DIV1)
	nxp.MuxIpLpi2c.Mux(0)           // LPI2C select PLL3_SW_60M

	// -------------------------------------------------------------------- CAN --
	nxp.ClockIpCan1.Enable(false)  // disable CAN clock gate
	nxp.ClockIpCan2.Enable(false)  //  ~
	nxp.ClockIpCan3.Enable(false)  //  ~
	nxp.ClockIpCan1S.Enable(false) //  ~
	nxp.ClockIpCan2S.Enable(false) //  ~
	nxp.ClockIpCan3S.Enable(false) //  ~
	nxp.DivIpCan.Div(1)            // divide CAN_CLK_PODF (DIV2)
	nxp.MuxIpCan.Mux(2)            // CAN select PLL3_SW_80M

	// ------------------------------------------------------------------- UART --
	nxp.ClockIpLpuart1.Enable(false) // disable UART clock gate
	nxp.ClockIpLpuart2.Enable(false) //  ~
	nxp.ClockIpLpuart3.Enable(false) //  ~
	nxp.ClockIpLpuart4.Enable(false) //  ~
	nxp.ClockIpLpuart5.Enable(false) //  ~
	nxp.ClockIpLpuart6.Enable(false) //  ~
	nxp.ClockIpLpuart7.Enable(false) //  ~
	nxp.ClockIpLpuart8.Enable(false) //  ~
	nxp.DivIpUart.Div(0)             // divide UART_CLK_PODF (DIV1)
	nxp.MuxIpUart.Mux(1)             // UART select OSC

	// -------------------------------------------------------------------- LCD --
	nxp.ClockIpLcdPixel.Enable(false) // disable LCDIF clock gate
	nxp.DivIpLcdifPre.Div(1)          // divide LCDIF_PRED (DIV2)
	nxp.DivIpLcdif.Div(3)             // divide LCDIF_CLK_PODF (DIV4)
	nxp.MuxIpLcdifPre.Mux(5)          // LCDIF_PRE select PLL3_PFD1

	// ------------------------------------------------------------------ SPDIF --
	nxp.ClockIpSpdif.Enable(false) // disable SPDIF clock gate
	nxp.DivIpSpdif0Pre.Div(1)      // divide SPDIF0_CLK_PRED (DIV2)
	nxp.DivIpSpdif0.Div(7)         // divide SPDIF0_CLK_PODF (DIV8)
	nxp.MuxIpSpdif.Mux(3)          // SPDIF select PLL3_SW

	// ----------------------------------------------------------------- FLEXIO --
	nxp.ClockIpFlexio1.Enable(false) // disable FLEXIO1 clock gate
	nxp.DivIpFlexio1Pre.Div(1)       // divide FLEXIO1_CLK_PRED (DIV2)
	nxp.DivIpFlexio1.Div(7)          // divide FLEXIO1_CLK_PODF (DIV8)
	nxp.MuxIpFlexio1.Mux(3)          // FLEXIO1 select PLL3_SW
	nxp.ClockIpFlexio2.Enable(false) // enable FLEXIO2 clock gate
	nxp.DivIpFlexio2Pre.Div(1)       // divide FLEXIO2_CLK_PRED (DIV2)
	nxp.DivIpFlexio2.Div(7)          // divide FLEXIO2_CLK_PODF (DIV8)
	nxp.MuxIpFlexio2.Mux(3)          // FLEXIO2 select PLL3_SW

	// ---------------------------------------------------------------- PLL/PFD --
	nxp.MuxIpPll3Sw.Mux(0) // PLL3_SW select PLL3_MAIN

	ArmPllConfig.Configure() // init ARM PLL
	// SYS PLL (PLL2) @ 528 MHz
	//   PFD0 = 396    MHz -> USDHC1/USDHC2(DIV2)=198 MHz
	//   PFD1 = 594    MHz -> (currently unused)
	//   PFD2 = 327.72 MHz -> SEMC(DIV2)=163.86 MHz, FlexSPI/FlexSPI2=327.72 MHz
	//   PFD3 = 454.73 MHz -> (currently unused)
	SysPllConfig.Configure(24, 16, 29, 16) // init SYS PLL and PFDs

	// USB1 PLL (PLL3) @ 480 MHz
	//   PFD0 -> (currently unused)
	//   PFD1 -> (currently unused)
	//   PFD2 -> (currently unused)
	//   PFD3 -> (currently unused)
	Usb1PllConfig.Configure() // init USB1 PLL and PFDs
	Usb2PllConfig.Configure() // init USB2 PLL

	// --------------------------------------------------------------- ARM CORE --
	nxp.MuxIpPrePeriph.Mux(3)  // PRE_PERIPH select ARM_PLL
	nxp.MuxIpPeriph.Mux(0)     // PERIPH select PRE_PERIPH
	nxp.MuxIpPeriphClk2.Mux(1) // PERIPH_CLK2 select OSC
	nxp.MuxIpPerclk.Mux(1)     // PERCLK select OSC

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
	nxp.ClockIpIomuxcGpr.Enable(false) // disable IOMUXC_GPR clock gate
	nxp.ClockIpIomuxc.Enable(false)    // disable IOMUXC clock gate
	// set GPT1 High frequency reference clock source
	nxp.IOMUXC_GPR.GPR5.ClearBits(nxp.IOMUXC_GPR_GPR5_VREF_1M_CLK_GPT1_Msk)
	// set GPT2 High frequency reference clock source
	nxp.IOMUXC_GPR.GPR5.ClearBits(nxp.IOMUXC_GPR_GPR5_VREF_1M_CLK_GPT2_Msk)

	// ------------------------------------------------------------------- GPIO --
	nxp.ClockIpGpio1.Enable(false) // disable GPIO clock gates
	nxp.ClockIpGpio2.Enable(false) //  ~
	nxp.ClockIpGpio3.Enable(false) //  ~
	nxp.ClockIpGpio4.Enable(false) //  ~
}

func enableTimerClocks() {

	// ---------------------------------------------------------------- GPT/PIT --
	nxp.ClockIpGpt1.Enable(true)  // enable GPT/PIT clock gates
	nxp.ClockIpGpt1S.Enable(true) //  ~
	nxp.ClockIpGpt2.Enable(true)  //  ~
	nxp.ClockIpGpt2S.Enable(true) //  ~
	nxp.ClockIpPit.Enable(true)   //  ~
}

func enablePinClocks() {

	// ----------------------------------------------------------------- IOMUXC --
	nxp.ClockIpIomuxcGpr.Enable(true) // enable IOMUXC clock gates
	nxp.ClockIpIomuxc.Enable(true)    //  ~

	// ------------------------------------------------------------------- GPIO --
	nxp.ClockIpGpio1.Enable(true) // enable GPIO clock gates
	nxp.ClockIpGpio2.Enable(true) //  ~
	nxp.ClockIpGpio3.Enable(true) //  ~
	nxp.ClockIpGpio4.Enable(true) //  ~
}

func enablePeripheralClocks() {

	// -------------------------------------------------------------------- ADC --
	nxp.ClockIpAdc1.Enable(true) // enable ADC clock gates
	nxp.ClockIpAdc2.Enable(true) //  ~

	// ------------------------------------------------------------------- XBAR --
	nxp.ClockIpXbar1.Enable(true) // enable XBAR clock gates
	nxp.ClockIpXbar2.Enable(true) //  ~
	nxp.ClockIpXbar3.Enable(true) //  ~

	// ------------------------------------------------------------------ USDHC --
	nxp.ClockIpUsdhc1.Enable(true) // enable USDHC clock gates
	nxp.ClockIpUsdhc2.Enable(true) //  ~

	// ------------------------------------------------------------------- SEMC --
	nxp.ClockIpSemc.Enable(true) // enable SEMC clock gate

	// ---------------------------------------------------------------- FLEXSPI --
	nxp.ClockIpFlexSpi2.Enable(true) // enable FLEXSPI2 clock gate

	// ------------------------------------------------------------------ LPSPI --
	nxp.ClockIpLpspi1.Enable(true) // enable LPSPI clock gate
	nxp.ClockIpLpspi2.Enable(true) //  ~
	nxp.ClockIpLpspi3.Enable(true) //  ~
	nxp.ClockIpLpspi4.Enable(true) //  ~

	// ------------------------------------------------------------------ LPI2C --
	nxp.ClockIpLpi2c1.Enable(true) // enable LPI2C clock gate
	nxp.ClockIpLpi2c2.Enable(true) //  ~
	nxp.ClockIpLpi2c3.Enable(true) //  ~

	// -------------------------------------------------------------------- CAN --
	nxp.ClockIpCan1.Enable(true)  // enable CAN clock gate
	nxp.ClockIpCan2.Enable(true)  //  ~
	nxp.ClockIpCan3.Enable(true)  //  ~
	nxp.ClockIpCan1S.Enable(true) //  ~
	nxp.ClockIpCan2S.Enable(true) //  ~
	nxp.ClockIpCan3S.Enable(true) //  ~

	// ------------------------------------------------------------------- UART --
	nxp.ClockIpLpuart1.Enable(true) // enable UART clock gate
	nxp.ClockIpLpuart2.Enable(true) //  ~
	nxp.ClockIpLpuart3.Enable(true) //  ~
	nxp.ClockIpLpuart4.Enable(true) //  ~
	nxp.ClockIpLpuart5.Enable(true) //  ~
	nxp.ClockIpLpuart6.Enable(true) //  ~
	nxp.ClockIpLpuart7.Enable(true) //  ~
	nxp.ClockIpLpuart8.Enable(true) //  ~

	// ----------------------------------------------------------------- FLEXIO --
	nxp.ClockIpFlexio1.Enable(true) // enable FLEXIO clock gates
	nxp.ClockIpFlexio2.Enable(true) //  ~
}
