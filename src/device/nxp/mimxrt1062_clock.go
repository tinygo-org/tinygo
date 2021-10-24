// Hand created file. DO NOT DELETE.
// Type definitions, fields, and constants associated with various clocks and
// peripherals of the NXP MIMXRT1062.

// +build nxp,mimxrt1062

package nxp

import (
	"runtime/volatile"
	"unsafe"
)

// Clock represents an individual peripheral clock that may be enabled/disabled
// at runtime. Clocks also have a method `Mux` for selecting the clock source
// and a method `Div` for selecting the hardware divisor. Note that many
// peripherals have an independent prescalar configuration applied to the output
// of this divisor.
type (
	Clock     uint32
	ClockMode uint8
)

// Enable activates or deactivates the clock gate of receiver Clock c.
func (c Clock) Enable(enable bool) {
	if enable {
		c.setGate(clockNeededRunWait)
	} else {
		c.setGate(clockNotNeeded)
	}
}

// Mux selects a clock source for the mux of the receiver Clock c.
func (c Clock) Mux(mux uint32) { c.setCcm(mux) }

// Div configures the prescalar divisor of the receiver Clock c.
func (c Clock) Div(div uint32) { c.setCcm(div) }

const (
	ClockModeRun  ClockMode = 0 // Remain in run mode
	ClockModeWait ClockMode = 1 // Transfer to wait mode
	ClockModeStop ClockMode = 2 // Transfer to stop mode
)

// Set configures the run mode of the MCU.
func (m ClockMode) Set() {
	CCM.CLPCR.Set((CCM.CLPCR.Get() & ^uint32(CCM_CLPCR_LPM_Msk)) |
		((uint32(m) << CCM_CLPCR_LPM_Pos) & CCM_CLPCR_LPM_Msk))
}

// Named oscillators
const (
	ClockCpu         Clock = 0x0  // CPU clock
	ClockAhb         Clock = 0x1  // AHB clock
	ClockSemc        Clock = 0x2  // SEMC clock
	ClockIpg         Clock = 0x3  // IPG clock
	ClockPer         Clock = 0x4  // PER clock
	ClockOsc         Clock = 0x5  // OSC clock selected by PMU_LOWPWR_CTRL[OSC_SEL]
	ClockRtc         Clock = 0x6  // RTC clock (RTCCLK)
	ClockArmPll      Clock = 0x7  // ARMPLLCLK
	ClockUsb1Pll     Clock = 0x8  // USB1PLLCLK
	ClockUsb1PllPfd0 Clock = 0x9  // USB1PLLPDF0CLK
	ClockUsb1PllPfd1 Clock = 0xA  // USB1PLLPFD1CLK
	ClockUsb1PllPfd2 Clock = 0xB  // USB1PLLPFD2CLK
	ClockUsb1PllPfd3 Clock = 0xC  // USB1PLLPFD3CLK
	ClockUsb2Pll     Clock = 0xD  // USB2PLLCLK
	ClockSysPll      Clock = 0xE  // SYSPLLCLK
	ClockSysPllPfd0  Clock = 0xF  // SYSPLLPDF0CLK
	ClockSysPllPfd1  Clock = 0x10 // SYSPLLPFD1CLK
	ClockSysPllPfd2  Clock = 0x11 // SYSPLLPFD2CLK
	ClockSysPllPfd3  Clock = 0x12 // SYSPLLPFD3CLK
	ClockEnetPll0    Clock = 0x13 // Enet PLLCLK ref_enetpll0
	ClockEnetPll1    Clock = 0x14 // Enet PLLCLK ref_enetpll1
	ClockEnetPll2    Clock = 0x15 // Enet PLLCLK ref_enetpll2
	ClockAudioPll    Clock = 0x16 // Audio PLLCLK
	ClockVideoPll    Clock = 0x17 // Video PLLCLK
)

// Named clocks of integrated peripherals
const (
	ClockIpAipsTz1     Clock = (0 << 8) | CCM_CCGR0_CG0_Pos  // CCGR0, CG0
	ClockIpAipsTz2     Clock = (0 << 8) | CCM_CCGR0_CG1_Pos  // CCGR0, CG1
	ClockIpMqs         Clock = (0 << 8) | CCM_CCGR0_CG2_Pos  // CCGR0, CG2
	ClockIpFlexSpiExsc Clock = (0 << 8) | CCM_CCGR0_CG3_Pos  // CCGR0, CG3
	ClockIpSimMMain    Clock = (0 << 8) | CCM_CCGR0_CG4_Pos  // CCGR0, CG4
	ClockIpDcp         Clock = (0 << 8) | CCM_CCGR0_CG5_Pos  // CCGR0, CG5
	ClockIpLpuart3     Clock = (0 << 8) | CCM_CCGR0_CG6_Pos  // CCGR0, CG6
	ClockIpCan1        Clock = (0 << 8) | CCM_CCGR0_CG7_Pos  // CCGR0, CG7
	ClockIpCan1S       Clock = (0 << 8) | CCM_CCGR0_CG8_Pos  // CCGR0, CG8
	ClockIpCan2        Clock = (0 << 8) | CCM_CCGR0_CG9_Pos  // CCGR0, CG9
	ClockIpCan2S       Clock = (0 << 8) | CCM_CCGR0_CG10_Pos // CCGR0, CG10
	ClockIpTrace       Clock = (0 << 8) | CCM_CCGR0_CG11_Pos // CCGR0, CG11
	ClockIpGpt2        Clock = (0 << 8) | CCM_CCGR0_CG12_Pos // CCGR0, CG12
	ClockIpGpt2S       Clock = (0 << 8) | CCM_CCGR0_CG13_Pos // CCGR0, CG13
	ClockIpLpuart2     Clock = (0 << 8) | CCM_CCGR0_CG14_Pos // CCGR0, CG14
	ClockIpGpio2       Clock = (0 << 8) | CCM_CCGR0_CG15_Pos // CCGR0, CG15

	ClockIpLpspi1   Clock = (1 << 8) | CCM_CCGR1_CG0_Pos  // CCGR1, CG0
	ClockIpLpspi2   Clock = (1 << 8) | CCM_CCGR1_CG1_Pos  // CCGR1, CG1
	ClockIpLpspi3   Clock = (1 << 8) | CCM_CCGR1_CG2_Pos  // CCGR1, CG2
	ClockIpLpspi4   Clock = (1 << 8) | CCM_CCGR1_CG3_Pos  // CCGR1, CG3
	ClockIpAdc2     Clock = (1 << 8) | CCM_CCGR1_CG4_Pos  // CCGR1, CG4
	ClockIpEnet     Clock = (1 << 8) | CCM_CCGR1_CG5_Pos  // CCGR1, CG5
	ClockIpPit      Clock = (1 << 8) | CCM_CCGR1_CG6_Pos  // CCGR1, CG6
	ClockIpAoi2     Clock = (1 << 8) | CCM_CCGR1_CG7_Pos  // CCGR1, CG7
	ClockIpAdc1     Clock = (1 << 8) | CCM_CCGR1_CG8_Pos  // CCGR1, CG8
	ClockIpSemcExsc Clock = (1 << 8) | CCM_CCGR1_CG9_Pos  // CCGR1, CG9
	ClockIpGpt1     Clock = (1 << 8) | CCM_CCGR1_CG10_Pos // CCGR1, CG10
	ClockIpGpt1S    Clock = (1 << 8) | CCM_CCGR1_CG11_Pos // CCGR1, CG11
	ClockIpLpuart4  Clock = (1 << 8) | CCM_CCGR1_CG12_Pos // CCGR1, CG12
	ClockIpGpio1    Clock = (1 << 8) | CCM_CCGR1_CG13_Pos // CCGR1, CG13
	ClockIpCsu      Clock = (1 << 8) | CCM_CCGR1_CG14_Pos // CCGR1, CG14
	ClockIpGpio5    Clock = (1 << 8) | CCM_CCGR1_CG15_Pos // CCGR1, CG15

	ClockIpOcramExsc  Clock = (2 << 8) | CCM_CCGR2_CG0_Pos  // CCGR2, CG0
	ClockIpCsi        Clock = (2 << 8) | CCM_CCGR2_CG1_Pos  // CCGR2, CG1
	ClockIpIomuxcSnvs Clock = (2 << 8) | CCM_CCGR2_CG2_Pos  // CCGR2, CG2
	ClockIpLpi2c1     Clock = (2 << 8) | CCM_CCGR2_CG3_Pos  // CCGR2, CG3
	ClockIpLpi2c2     Clock = (2 << 8) | CCM_CCGR2_CG4_Pos  // CCGR2, CG4
	ClockIpLpi2c3     Clock = (2 << 8) | CCM_CCGR2_CG5_Pos  // CCGR2, CG5
	ClockIpOcotp      Clock = (2 << 8) | CCM_CCGR2_CG6_Pos  // CCGR2, CG6
	ClockIpXbar3      Clock = (2 << 8) | CCM_CCGR2_CG7_Pos  // CCGR2, CG7
	ClockIpIpmux1     Clock = (2 << 8) | CCM_CCGR2_CG8_Pos  // CCGR2, CG8
	ClockIpIpmux2     Clock = (2 << 8) | CCM_CCGR2_CG9_Pos  // CCGR2, CG9
	ClockIpIpmux3     Clock = (2 << 8) | CCM_CCGR2_CG10_Pos // CCGR2, CG10
	ClockIpXbar1      Clock = (2 << 8) | CCM_CCGR2_CG11_Pos // CCGR2, CG11
	ClockIpXbar2      Clock = (2 << 8) | CCM_CCGR2_CG12_Pos // CCGR2, CG12
	ClockIpGpio3      Clock = (2 << 8) | CCM_CCGR2_CG13_Pos // CCGR2, CG13
	ClockIpLcd        Clock = (2 << 8) | CCM_CCGR2_CG14_Pos // CCGR2, CG14
	ClockIpPxp        Clock = (2 << 8) | CCM_CCGR2_CG15_Pos // CCGR2, CG15

	ClockIpFlexio2       Clock = (3 << 8) | CCM_CCGR3_CG0_Pos  // CCGR3, CG0
	ClockIpLpuart5       Clock = (3 << 8) | CCM_CCGR3_CG1_Pos  // CCGR3, CG1
	ClockIpSemc          Clock = (3 << 8) | CCM_CCGR3_CG2_Pos  // CCGR3, CG2
	ClockIpLpuart6       Clock = (3 << 8) | CCM_CCGR3_CG3_Pos  // CCGR3, CG3
	ClockIpAoi1          Clock = (3 << 8) | CCM_CCGR3_CG4_Pos  // CCGR3, CG4
	ClockIpLcdPixel      Clock = (3 << 8) | CCM_CCGR3_CG5_Pos  // CCGR3, CG5
	ClockIpGpio4         Clock = (3 << 8) | CCM_CCGR3_CG6_Pos  // CCGR3, CG6
	ClockIpEwm0          Clock = (3 << 8) | CCM_CCGR3_CG7_Pos  // CCGR3, CG7
	ClockIpWdog1         Clock = (3 << 8) | CCM_CCGR3_CG8_Pos  // CCGR3, CG8
	ClockIpFlexRam       Clock = (3 << 8) | CCM_CCGR3_CG9_Pos  // CCGR3, CG9
	ClockIpAcmp1         Clock = (3 << 8) | CCM_CCGR3_CG10_Pos // CCGR3, CG10
	ClockIpAcmp2         Clock = (3 << 8) | CCM_CCGR3_CG11_Pos // CCGR3, CG11
	ClockIpAcmp3         Clock = (3 << 8) | CCM_CCGR3_CG12_Pos // CCGR3, CG12
	ClockIpAcmp4         Clock = (3 << 8) | CCM_CCGR3_CG13_Pos // CCGR3, CG13
	ClockIpOcram         Clock = (3 << 8) | CCM_CCGR3_CG14_Pos // CCGR3, CG14
	ClockIpIomuxcSnvsGpr Clock = (3 << 8) | CCM_CCGR3_CG15_Pos // CCGR3, CG15

	ClockIpIomuxc    Clock = (4 << 8) | CCM_CCGR4_CG1_Pos  // CCGR4, CG1
	ClockIpIomuxcGpr Clock = (4 << 8) | CCM_CCGR4_CG2_Pos  // CCGR4, CG2
	ClockIpBee       Clock = (4 << 8) | CCM_CCGR4_CG3_Pos  // CCGR4, CG3
	ClockIpSimM7     Clock = (4 << 8) | CCM_CCGR4_CG4_Pos  // CCGR4, CG4
	ClockIpTsc       Clock = (4 << 8) | CCM_CCGR4_CG5_Pos  // CCGR4, CG5
	ClockIpSimM      Clock = (4 << 8) | CCM_CCGR4_CG6_Pos  // CCGR4, CG6
	ClockIpSimEms    Clock = (4 << 8) | CCM_CCGR4_CG7_Pos  // CCGR4, CG7
	ClockIpPwm1      Clock = (4 << 8) | CCM_CCGR4_CG8_Pos  // CCGR4, CG8
	ClockIpPwm2      Clock = (4 << 8) | CCM_CCGR4_CG9_Pos  // CCGR4, CG9
	ClockIpPwm3      Clock = (4 << 8) | CCM_CCGR4_CG10_Pos // CCGR4, CG10
	ClockIpPwm4      Clock = (4 << 8) | CCM_CCGR4_CG11_Pos // CCGR4, CG11
	ClockIpEnc1      Clock = (4 << 8) | CCM_CCGR4_CG12_Pos // CCGR4, CG12
	ClockIpEnc2      Clock = (4 << 8) | CCM_CCGR4_CG13_Pos // CCGR4, CG13
	ClockIpEnc3      Clock = (4 << 8) | CCM_CCGR4_CG14_Pos // CCGR4, CG14
	ClockIpEnc4      Clock = (4 << 8) | CCM_CCGR4_CG15_Pos // CCGR4, CG15

	ClockIpRom     Clock = (5 << 8) | CCM_CCGR5_CG0_Pos  // CCGR5, CG0
	ClockIpFlexio1 Clock = (5 << 8) | CCM_CCGR5_CG1_Pos  // CCGR5, CG1
	ClockIpWdog3   Clock = (5 << 8) | CCM_CCGR5_CG2_Pos  // CCGR5, CG2
	ClockIpDma     Clock = (5 << 8) | CCM_CCGR5_CG3_Pos  // CCGR5, CG3
	ClockIpKpp     Clock = (5 << 8) | CCM_CCGR5_CG4_Pos  // CCGR5, CG4
	ClockIpWdog2   Clock = (5 << 8) | CCM_CCGR5_CG5_Pos  // CCGR5, CG5
	ClockIpAipsTz4 Clock = (5 << 8) | CCM_CCGR5_CG6_Pos  // CCGR5, CG6
	ClockIpSpdif   Clock = (5 << 8) | CCM_CCGR5_CG7_Pos  // CCGR5, CG7
	ClockIpSimMain Clock = (5 << 8) | CCM_CCGR5_CG8_Pos  // CCGR5, CG8
	ClockIpSai1    Clock = (5 << 8) | CCM_CCGR5_CG9_Pos  // CCGR5, CG9
	ClockIpSai2    Clock = (5 << 8) | CCM_CCGR5_CG10_Pos // CCGR5, CG10
	ClockIpSai3    Clock = (5 << 8) | CCM_CCGR5_CG11_Pos // CCGR5, CG11
	ClockIpLpuart1 Clock = (5 << 8) | CCM_CCGR5_CG12_Pos // CCGR5, CG12
	ClockIpLpuart7 Clock = (5 << 8) | CCM_CCGR5_CG13_Pos // CCGR5, CG13
	ClockIpSnvsHp  Clock = (5 << 8) | CCM_CCGR5_CG14_Pos // CCGR5, CG14
	ClockIpSnvsLp  Clock = (5 << 8) | CCM_CCGR5_CG15_Pos // CCGR5, CG15

	ClockIpUsbOh3  Clock = (6 << 8) | CCM_CCGR6_CG0_Pos  // CCGR6, CG0
	ClockIpUsdhc1  Clock = (6 << 8) | CCM_CCGR6_CG1_Pos  // CCGR6, CG1
	ClockIpUsdhc2  Clock = (6 << 8) | CCM_CCGR6_CG2_Pos  // CCGR6, CG2
	ClockIpDcdc    Clock = (6 << 8) | CCM_CCGR6_CG3_Pos  // CCGR6, CG3
	ClockIpIpmux4  Clock = (6 << 8) | CCM_CCGR6_CG4_Pos  // CCGR6, CG4
	ClockIpFlexSpi Clock = (6 << 8) | CCM_CCGR6_CG5_Pos  // CCGR6, CG5
	ClockIpTrng    Clock = (6 << 8) | CCM_CCGR6_CG6_Pos  // CCGR6, CG6
	ClockIpLpuart8 Clock = (6 << 8) | CCM_CCGR6_CG7_Pos  // CCGR6, CG7
	ClockIpTimer4  Clock = (6 << 8) | CCM_CCGR6_CG8_Pos  // CCGR6, CG8
	ClockIpAipsTz3 Clock = (6 << 8) | CCM_CCGR6_CG9_Pos  // CCGR6, CG9
	ClockIpSimPer  Clock = (6 << 8) | CCM_CCGR6_CG10_Pos // CCGR6, CG10
	ClockIpAnadig  Clock = (6 << 8) | CCM_CCGR6_CG11_Pos // CCGR6, CG11
	ClockIpLpi2c4  Clock = (6 << 8) | CCM_CCGR6_CG12_Pos // CCGR6, CG12
	ClockIpTimer1  Clock = (6 << 8) | CCM_CCGR6_CG13_Pos // CCGR6, CG13
	ClockIpTimer2  Clock = (6 << 8) | CCM_CCGR6_CG14_Pos // CCGR6, CG14
	ClockIpTimer3  Clock = (6 << 8) | CCM_CCGR6_CG15_Pos // CCGR6, CG15

	ClockIpEnet2    Clock = (7 << 8) | CCM_CCGR7_CG0_Pos // CCGR7, CG0
	ClockIpFlexSpi2 Clock = (7 << 8) | CCM_CCGR7_CG1_Pos // CCGR7, CG1
	ClockIpAxbsL    Clock = (7 << 8) | CCM_CCGR7_CG2_Pos // CCGR7, CG2
	ClockIpCan3     Clock = (7 << 8) | CCM_CCGR7_CG3_Pos // CCGR7, CG3
	ClockIpCan3S    Clock = (7 << 8) | CCM_CCGR7_CG4_Pos // CCGR7, CG4
	ClockIpAipsLite Clock = (7 << 8) | CCM_CCGR7_CG5_Pos // CCGR7, CG5
	ClockIpFlexio3  Clock = (7 << 8) | CCM_CCGR7_CG6_Pos // CCGR7, CG6
)

// PLL name
const (
	ClockPllArm     Clock = ((offPllArm & 0xFFF) << 16) | CCM_ANALOG_PLL_ARM_ENABLE_Pos            // PLL ARM
	ClockPllSys     Clock = ((offPllSys & 0xFFF) << 16) | CCM_ANALOG_PLL_SYS_ENABLE_Pos            // PLL SYS
	ClockPllUsb1    Clock = ((offPllUsb1 & 0xFFF) << 16) | CCM_ANALOG_PLL_USB1_ENABLE_Pos          // PLL USB1
	ClockPllAudio   Clock = ((offPllAudio & 0xFFF) << 16) | CCM_ANALOG_PLL_AUDIO_ENABLE_Pos        // PLL Audio
	ClockPllVideo   Clock = ((offPllVideo & 0xFFF) << 16) | CCM_ANALOG_PLL_VIDEO_ENABLE_Pos        // PLL Video
	ClockPllEnet    Clock = ((offPllEnet & 0xFFF) << 16) | CCM_ANALOG_PLL_ENET_ENABLE_Pos          // PLL Enet0
	ClockPllEnet2   Clock = ((offPllEnet & 0xFFF) << 16) | CCM_ANALOG_PLL_ENET_ENET2_REF_EN_Pos    // PLL Enet1
	ClockPllEnet25M Clock = ((offPllEnet & 0xFFF) << 16) | CCM_ANALOG_PLL_ENET_ENET_25M_REF_EN_Pos // PLL Enet2
	ClockPllUsb2    Clock = ((offPllUsb2 & 0xFFF) << 16) | CCM_ANALOG_PLL_USB2_ENABLE_Pos          // PLL USB2
)

// PLL PFD name
const (
	ClockPfd0 Clock = 0 // PLL PFD0
	ClockPfd1 Clock = 1 // PLL PFD1
	ClockPfd2 Clock = 2 // PLL PFD2
	ClockPfd3 Clock = 3 // PLL PFD3
)

// Named clock muxes of integrated peripherals
const (
	MuxIpPll3Sw     Clock = (offCCSR & 0xFF) | (CCM_CCSR_PLL3_SW_CLK_SEL_Pos << 8) | (((CCM_CCSR_PLL3_SW_CLK_SEL_Msk >> CCM_CCSR_PLL3_SW_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                          // pll3_sw_clk mux name
	MuxIpPeriph     Clock = (offCBCDR & 0xFF) | (CCM_CBCDR_PERIPH_CLK_SEL_Pos << 8) | (((CCM_CBCDR_PERIPH_CLK_SEL_Msk >> CCM_CBCDR_PERIPH_CLK_SEL_Pos) & 0x1FFF) << 13) | (CCM_CDHIPR_PERIPH_CLK_SEL_BUSY_Pos << 26) // periph mux name
	MuxIpSemcAlt    Clock = (offCBCDR & 0xFF) | (CCM_CBCDR_SEMC_ALT_CLK_SEL_Pos << 8) | (((CCM_CBCDR_SEMC_ALT_CLK_SEL_Msk >> CCM_CBCDR_SEMC_ALT_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                   // semc mux name
	MuxIpSemc       Clock = (offCBCDR & 0xFF) | (CCM_CBCDR_SEMC_CLK_SEL_Pos << 8) | (((CCM_CBCDR_SEMC_CLK_SEL_Msk >> CCM_CBCDR_SEMC_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                               // semc mux name
	MuxIpPrePeriph  Clock = (offCBCMR & 0xFF) | (CCM_CBCMR_PRE_PERIPH_CLK_SEL_Pos << 8) | (((CCM_CBCMR_PRE_PERIPH_CLK_SEL_Msk >> CCM_CBCMR_PRE_PERIPH_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)             // pre-periph mux name
	MuxIpTrace      Clock = (offCBCMR & 0xFF) | (CCM_CBCMR_TRACE_CLK_SEL_Pos << 8) | (((CCM_CBCMR_TRACE_CLK_SEL_Msk >> CCM_CBCMR_TRACE_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                            // trace mux name
	MuxIpPeriphClk2 Clock = (offCBCMR & 0xFF) | (CCM_CBCMR_PERIPH_CLK2_SEL_Pos << 8) | (((CCM_CBCMR_PERIPH_CLK2_SEL_Msk >> CCM_CBCMR_PERIPH_CLK2_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                      // periph clock2 mux name
	MuxIpFlexSpi2   Clock = (offCBCMR & 0xFF) | (CCM_CBCMR_FLEXSPI2_CLK_SEL_Pos << 8) | (((CCM_CBCMR_FLEXSPI2_CLK_SEL_Msk >> CCM_CBCMR_FLEXSPI2_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                   // flexspi2 mux name
	MuxIpLpspi      Clock = (offCBCMR & 0xFF) | (CCM_CBCMR_LPSPI_CLK_SEL_Pos << 8) | (((CCM_CBCMR_LPSPI_CLK_SEL_Msk >> CCM_CBCMR_LPSPI_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                            // lpspi mux name
	MuxIpFlexSpi    Clock = (offCSCMR1 & 0xFF) | (CCM_CSCMR1_FLEXSPI_CLK_SEL_Pos << 8) | (((CCM_CSCMR1_FLEXSPI_CLK_SEL_Msk >> CCM_CSCMR1_FLEXSPI_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                  // flexspi mux name
	MuxIpUsdhc2     Clock = (offCSCMR1 & 0xFF) | (CCM_CSCMR1_USDHC2_CLK_SEL_Pos << 8) | (((CCM_CSCMR1_USDHC2_CLK_SEL_Msk >> CCM_CSCMR1_USDHC2_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                     // usdhc2 mux name
	MuxIpUsdhc1     Clock = (offCSCMR1 & 0xFF) | (CCM_CSCMR1_USDHC1_CLK_SEL_Pos << 8) | (((CCM_CSCMR1_USDHC1_CLK_SEL_Msk >> CCM_CSCMR1_USDHC1_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                     // usdhc1 mux name
	MuxIpSai3       Clock = (offCSCMR1 & 0xFF) | (CCM_CSCMR1_SAI3_CLK_SEL_Pos << 8) | (((CCM_CSCMR1_SAI3_CLK_SEL_Msk >> CCM_CSCMR1_SAI3_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                           // sai3 mux name
	MuxIpSai2       Clock = (offCSCMR1 & 0xFF) | (CCM_CSCMR1_SAI2_CLK_SEL_Pos << 8) | (((CCM_CSCMR1_SAI2_CLK_SEL_Msk >> CCM_CSCMR1_SAI2_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                           // sai2 mux name
	MuxIpSai1       Clock = (offCSCMR1 & 0xFF) | (CCM_CSCMR1_SAI1_CLK_SEL_Pos << 8) | (((CCM_CSCMR1_SAI1_CLK_SEL_Msk >> CCM_CSCMR1_SAI1_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                           // sai1 mux name
	MuxIpPerclk     Clock = (offCSCMR1 & 0xFF) | (CCM_CSCMR1_PERCLK_CLK_SEL_Pos << 8) | (((CCM_CSCMR1_PERCLK_CLK_SEL_Msk >> CCM_CSCMR1_PERCLK_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                     // perclk mux name
	MuxIpFlexio2    Clock = (offCSCMR2 & 0xFF) | (CCM_CSCMR2_FLEXIO2_CLK_SEL_Pos << 8) | (((CCM_CSCMR2_FLEXIO2_CLK_SEL_Msk >> CCM_CSCMR2_FLEXIO2_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                  // flexio2 mux name
	MuxIpCan        Clock = (offCSCMR2 & 0xFF) | (CCM_CSCMR2_CAN_CLK_SEL_Pos << 8) | (((CCM_CSCMR2_CAN_CLK_SEL_Msk >> CCM_CSCMR2_CAN_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                              // can mux name
	MuxIpUart       Clock = (offCSCDR1 & 0xFF) | (CCM_CSCDR1_UART_CLK_SEL_Pos << 8) | (((CCM_CSCDR1_UART_CLK_SEL_Msk >> CCM_CSCDR1_UART_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                           // uart mux name
	MuxIpSpdif      Clock = (offCDCDR & 0xFF) | (CCM_CDCDR_SPDIF0_CLK_SEL_Pos << 8) | (((CCM_CDCDR_SPDIF0_CLK_SEL_Msk >> CCM_CDCDR_SPDIF0_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                         // spdif mux name
	MuxIpFlexio1    Clock = (offCDCDR & 0xFF) | (CCM_CDCDR_FLEXIO1_CLK_SEL_Pos << 8) | (((CCM_CDCDR_FLEXIO1_CLK_SEL_Msk >> CCM_CDCDR_FLEXIO1_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                      // flexio1 mux name
	MuxIpLpi2c      Clock = (offCSCDR2 & 0xFF) | (CCM_CSCDR2_LPI2C_CLK_SEL_Pos << 8) | (((CCM_CSCDR2_LPI2C_CLK_SEL_Msk >> CCM_CSCDR2_LPI2C_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                        // lpi2c mux name
	MuxIpLcdifPre   Clock = (offCSCDR2 & 0xFF) | (CCM_CSCDR2_LCDIF_PRE_CLK_SEL_Pos << 8) | (((CCM_CSCDR2_LCDIF_PRE_CLK_SEL_Msk >> CCM_CSCDR2_LCDIF_PRE_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)            // lcdif pre mux name
	MuxIpCsi        Clock = (offCSCDR3 & 0xFF) | (CCM_CSCDR3_CSI_CLK_SEL_Pos << 8) | (((CCM_CSCDR3_CSI_CLK_SEL_Msk >> CCM_CSCDR3_CSI_CLK_SEL_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                              // csi mux name
)

// Named hardware clock divisors of integrated peripherals
const (
	DivIpArm        Clock = (offCACRR & 0xFF) | (CCM_CACRR_ARM_PODF_Pos << 8) | (((CCM_CACRR_ARM_PODF_Msk >> CCM_CACRR_ARM_PODF_Pos) & 0x1FFF) << 13) | (CCM_CDHIPR_ARM_PODF_BUSY_Pos << 26)           // core div name
	DivIpPeriphClk2 Clock = (offCBCDR & 0xFF) | (CCM_CBCDR_PERIPH_CLK2_PODF_Pos << 8) | (((CCM_CBCDR_PERIPH_CLK2_PODF_Msk >> CCM_CBCDR_PERIPH_CLK2_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)     // periph clock2 div name
	DivIpSemc       Clock = (offCBCDR & 0xFF) | (CCM_CBCDR_SEMC_PODF_Pos << 8) | (((CCM_CBCDR_SEMC_PODF_Msk >> CCM_CBCDR_SEMC_PODF_Pos) & 0x1FFF) << 13) | (CCM_CDHIPR_SEMC_PODF_BUSY_Pos << 26)       // semc div name
	DivIpAhb        Clock = (offCBCDR & 0xFF) | (CCM_CBCDR_AHB_PODF_Pos << 8) | (((CCM_CBCDR_AHB_PODF_Msk >> CCM_CBCDR_AHB_PODF_Pos) & 0x1FFF) << 13) | (CCM_CDHIPR_AHB_PODF_BUSY_Pos << 26)           // ahb div name
	DivIpIpg        Clock = (offCBCDR & 0xFF) | (CCM_CBCDR_IPG_PODF_Pos << 8) | (((CCM_CBCDR_IPG_PODF_Msk >> CCM_CBCDR_IPG_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                             // ipg div name
	DivIpFlexSpi2   Clock = (offCBCMR & 0xFF) | (CCM_CBCMR_FLEXSPI2_PODF_Pos << 8) | (((CCM_CBCMR_FLEXSPI2_PODF_Msk >> CCM_CBCMR_FLEXSPI2_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)              // flexspi2 div name
	DivIpLpspi      Clock = (offCBCMR & 0xFF) | (CCM_CBCMR_LPSPI_PODF_Pos << 8) | (((CCM_CBCMR_LPSPI_PODF_Msk >> CCM_CBCMR_LPSPI_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                       // lpspi div name
	DivIpLcdif      Clock = (offCBCMR & 0xFF) | (CCM_CBCMR_LCDIF_PODF_Pos << 8) | (((CCM_CBCMR_LCDIF_PODF_Msk >> CCM_CBCMR_LCDIF_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                       // lcdif div name
	DivIpFlexSpi    Clock = (offCSCMR1 & 0xFF) | (CCM_CSCMR1_FLEXSPI_PODF_Pos << 8) | (((CCM_CSCMR1_FLEXSPI_PODF_Msk >> CCM_CSCMR1_FLEXSPI_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)             // flexspi div name
	DivIpPerclk     Clock = (offCSCMR1 & 0xFF) | (CCM_CSCMR1_PERCLK_PODF_Pos << 8) | (((CCM_CSCMR1_PERCLK_PODF_Msk >> CCM_CSCMR1_PERCLK_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                // perclk div name
	DivIpCan        Clock = (offCSCMR2 & 0xFF) | (CCM_CSCMR2_CAN_CLK_PODF_Pos << 8) | (((CCM_CSCMR2_CAN_CLK_PODF_Msk >> CCM_CSCMR2_CAN_CLK_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)             // can div name
	DivIpTrace      Clock = (offCSCDR1 & 0xFF) | (CCM_CSCDR1_TRACE_PODF_Pos << 8) | (((CCM_CSCDR1_TRACE_PODF_Msk >> CCM_CSCDR1_TRACE_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                   // trace div name
	DivIpUsdhc2     Clock = (offCSCDR1 & 0xFF) | (CCM_CSCDR1_USDHC2_PODF_Pos << 8) | (((CCM_CSCDR1_USDHC2_PODF_Msk >> CCM_CSCDR1_USDHC2_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                // usdhc2 div name
	DivIpUsdhc1     Clock = (offCSCDR1 & 0xFF) | (CCM_CSCDR1_USDHC1_PODF_Pos << 8) | (((CCM_CSCDR1_USDHC1_PODF_Msk >> CCM_CSCDR1_USDHC1_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                // usdhc1 div name
	DivIpUart       Clock = (offCSCDR1 & 0xFF) | (CCM_CSCDR1_UART_CLK_PODF_Pos << 8) | (((CCM_CSCDR1_UART_CLK_PODF_Msk >> CCM_CSCDR1_UART_CLK_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)          // uart div name
	DivIpFlexio2    Clock = (offCS1CDR & 0xFF) | (CCM_CS1CDR_FLEXIO2_CLK_PODF_Pos << 8) | (((CCM_CS1CDR_FLEXIO2_CLK_PODF_Msk >> CCM_CS1CDR_FLEXIO2_CLK_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26) // flexio2 pre div name
	DivIpSai3Pre    Clock = (offCS1CDR & 0xFF) | (CCM_CS1CDR_SAI3_CLK_PRED_Pos << 8) | (((CCM_CS1CDR_SAI3_CLK_PRED_Msk >> CCM_CS1CDR_SAI3_CLK_PRED_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)          // sai3 pre div name
	DivIpSai3       Clock = (offCS1CDR & 0xFF) | (CCM_CS1CDR_SAI3_CLK_PODF_Pos << 8) | (((CCM_CS1CDR_SAI3_CLK_PODF_Msk >> CCM_CS1CDR_SAI3_CLK_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)          // sai3 div name
	DivIpFlexio2Pre Clock = (offCS1CDR & 0xFF) | (CCM_CS1CDR_FLEXIO2_CLK_PRED_Pos << 8) | (((CCM_CS1CDR_FLEXIO2_CLK_PRED_Msk >> CCM_CS1CDR_FLEXIO2_CLK_PRED_Pos) & 0x1FFF) << 13) | (noBusyWait << 26) // sai3 pre div name
	DivIpSai1Pre    Clock = (offCS1CDR & 0xFF) | (CCM_CS1CDR_SAI1_CLK_PRED_Pos << 8) | (((CCM_CS1CDR_SAI1_CLK_PRED_Msk >> CCM_CS1CDR_SAI1_CLK_PRED_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)          // sai1 pre div name
	DivIpSai1       Clock = (offCS1CDR & 0xFF) | (CCM_CS1CDR_SAI1_CLK_PODF_Pos << 8) | (((CCM_CS1CDR_SAI1_CLK_PODF_Msk >> CCM_CS1CDR_SAI1_CLK_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)          // sai1 div name
	DivIpSai2Pre    Clock = (offCS2CDR & 0xFF) | (CCM_CS2CDR_SAI2_CLK_PRED_Pos << 8) | (((CCM_CS2CDR_SAI2_CLK_PRED_Msk >> CCM_CS2CDR_SAI2_CLK_PRED_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)          // sai2 pre div name
	DivIpSai2       Clock = (offCS2CDR & 0xFF) | (CCM_CS2CDR_SAI2_CLK_PODF_Pos << 8) | (((CCM_CS2CDR_SAI2_CLK_PODF_Msk >> CCM_CS2CDR_SAI2_CLK_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)          // sai2 div name
	DivIpSpdif0Pre  Clock = (offCDCDR & 0xFF) | (CCM_CDCDR_SPDIF0_CLK_PRED_Pos << 8) | (((CCM_CDCDR_SPDIF0_CLK_PRED_Msk >> CCM_CDCDR_SPDIF0_CLK_PRED_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)        // spdif pre div name
	DivIpSpdif0     Clock = (offCDCDR & 0xFF) | (CCM_CDCDR_SPDIF0_CLK_PODF_Pos << 8) | (((CCM_CDCDR_SPDIF0_CLK_PODF_Msk >> CCM_CDCDR_SPDIF0_CLK_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)        // spdif div name
	DivIpFlexio1Pre Clock = (offCDCDR & 0xFF) | (CCM_CDCDR_FLEXIO1_CLK_PRED_Pos << 8) | (((CCM_CDCDR_FLEXIO1_CLK_PRED_Msk >> CCM_CDCDR_FLEXIO1_CLK_PRED_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)     // flexio1 pre div name
	DivIpFlexio1    Clock = (offCDCDR & 0xFF) | (CCM_CDCDR_FLEXIO1_CLK_PODF_Pos << 8) | (((CCM_CDCDR_FLEXIO1_CLK_PODF_Msk >> CCM_CDCDR_FLEXIO1_CLK_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)     // flexio1 div name
	DivIpLpi2c      Clock = (offCSCDR2 & 0xFF) | (CCM_CSCDR2_LPI2C_CLK_PODF_Pos << 8) | (((CCM_CSCDR2_LPI2C_CLK_PODF_Msk >> CCM_CSCDR2_LPI2C_CLK_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)       // lpi2c div name
	DivIpLcdifPre   Clock = (offCSCDR2 & 0xFF) | (CCM_CSCDR2_LCDIF_PRED_Pos << 8) | (((CCM_CSCDR2_LCDIF_PRED_Msk >> CCM_CSCDR2_LCDIF_PRED_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                   // lcdif pre div name
	DivIpCsi        Clock = (offCSCDR3 & 0xFF) | (CCM_CSCDR3_CSI_PODF_Pos << 8) | (((CCM_CSCDR3_CSI_PODF_Msk >> CCM_CSCDR3_CSI_PODF_Pos) & 0x1FFF) << 13) | (noBusyWait << 26)                         // csi div name
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

// analog PLL definition
const (
	pllBypassPos       = 16
	pllBypassClkSrcMsk = 0xC000
	pllBypassClkSrcPos = 14
)

// PLL clock source, bypass cloco source also
const (
	pllSrc24M   = 0 // Pll clock source 24M
	pllSrcClkPN = 1 // Pll clock source CLK1_P and CLK1_N
)

const (
	clockNotNeeded     uint32 = 0 // Clock is off during all modes
	clockNeededRun     uint32 = 1 // Clock is on in run mode, but off in WAIT and STOP modes
	clockNeededRunWait uint32 = 3 // Clock is on during all modes, except STOP mode
)

// getGate returns the CCM clock gating register for the receiver clk.
func (clk Clock) getGate() *volatile.Register32 {
	switch clk >> 8 {
	case 0:
		return &CCM.CCGR0.Register32
	case 1:
		return &CCM.CCGR1.Register32
	case 2:
		return &CCM.CCGR2.Register32
	case 3:
		return &CCM.CCGR3.Register32
	case 4:
		return &CCM.CCGR4.Register32
	case 5:
		return &CCM.CCGR5.Register32
	case 6:
		return &CCM.CCGR6.Register32
	case 7:
		return &CCM.CCGR7.Register32
	default:
		panic("nxp: invalid clock")
	}
}

// setGate enables or disables the receiver clk using its gating register.
func (clk Clock) setGate(value uint32) {
	reg := clk.getGate()
	shift := clk & 0x1F
	reg.Set((reg.Get() & ^(3 << shift)) | (value << shift))
}

func (clk Clock) setCcm(value uint32) {
	const ccmBase = 0x400fc000
	reg := (*volatile.Register32)(unsafe.Pointer(uintptr(ccmBase + (uint32(clk) & 0xFF))))
	msk := ((uint32(clk) >> 13) & 0x1FFF) << ((uint32(clk) >> 8) & 0x1F)
	pos := (uint32(clk) >> 8) & 0x1F
	bsy := (uint32(clk) >> 26) & 0x3F
	reg.Set((reg.Get() & ^uint32(msk)) | ((value << pos) & msk))
	if bsy < noBusyWait {
		for CCM.CDHIPR.HasBits(1 << bsy) {
		}
	}
}

func setSysPfd(value ...uint32) {
	for i, val := range value {
		pfd528 := CCM_ANALOG.PFD_528.Get() &
			^((CCM_ANALOG_PFD_528_PFD0_CLKGATE_Msk | CCM_ANALOG_PFD_528_PFD0_FRAC_Msk) << (8 * uint32(i)))
		frac := (val << CCM_ANALOG_PFD_528_PFD0_FRAC_Pos) & CCM_ANALOG_PFD_528_PFD0_FRAC_Msk
		// disable the clock output first
		CCM_ANALOG.PFD_528.Set(pfd528 | (CCM_ANALOG_PFD_528_PFD0_CLKGATE_Msk << (8 * uint32(i))))
		// set the new value and enable output
		CCM_ANALOG.PFD_528.Set(pfd528 | (frac << (8 * uint32(i))))
	}
}

func setUsb1Pfd(value ...uint32) {
	for i, val := range value {
		pfd480 := CCM_ANALOG.PFD_480.Get() &
			^((CCM_ANALOG_PFD_480_PFD0_CLKGATE_Msk | CCM_ANALOG_PFD_480_PFD0_FRAC_Msk) << (8 * uint32(i)))
		frac := (val << CCM_ANALOG_PFD_480_PFD0_FRAC_Pos) & CCM_ANALOG_PFD_480_PFD0_FRAC_Msk
		// disable the clock output first
		CCM_ANALOG.PFD_480.Set(pfd480 | (CCM_ANALOG_PFD_480_PFD0_CLKGATE_Msk << (8 * uint32(i))))
		// set the new value and enable output
		CCM_ANALOG.PFD_480.Set(pfd480 | (frac << (8 * uint32(i))))
	}
}

// PLL configuration for ARM
type ClockConfigArmPll struct {
	LoopDivider uint32 // PLL loop divider. Valid range for divider value: 54-108. Fout=Fin*LoopDivider/2.
	Src         uint8  // Pll clock source, reference _clock_pll_clk_src
}

func (cfg ClockConfigArmPll) Configure() {

	// bypass PLL first
	src := (uint32(cfg.Src) << CCM_ANALOG_PLL_ARM_BYPASS_CLK_SRC_Pos) & CCM_ANALOG_PLL_ARM_BYPASS_CLK_SRC_Msk
	CCM_ANALOG.PLL_ARM.Set(
		(CCM_ANALOG.PLL_ARM.Get() & ^uint32(CCM_ANALOG_PLL_ARM_BYPASS_CLK_SRC_Msk)) |
			CCM_ANALOG_PLL_ARM_BYPASS_Msk | src)

	sel := (cfg.LoopDivider << CCM_ANALOG_PLL_ARM_DIV_SELECT_Pos) & CCM_ANALOG_PLL_ARM_DIV_SELECT_Msk
	CCM_ANALOG.PLL_ARM.Set(
		(CCM_ANALOG.PLL_ARM.Get() & ^uint32(CCM_ANALOG_PLL_ARM_DIV_SELECT_Msk|CCM_ANALOG_PLL_ARM_POWERDOWN_Msk)) |
			CCM_ANALOG_PLL_ARM_ENABLE_Msk | sel)

	for !CCM_ANALOG.PLL_ARM.HasBits(CCM_ANALOG_PLL_ARM_LOCK_Msk) {
	}

	// disable bypass
	CCM_ANALOG.PLL_ARM.ClearBits(CCM_ANALOG_PLL_ARM_BYPASS_Msk)
}

// PLL configuration for System
type ClockConfigSysPll struct {
	LoopDivider uint8  // PLL loop divider. Intended to be 1 (528M): 0 - Fout=Fref*20, 1 - Fout=Fref*22
	Numerator   uint32 // 30 bit Numerator of fractional loop divider.
	Denominator uint32 // 30 bit Denominator of fractional loop divider
	Src         uint8  // Pll clock source, reference _clock_pll_clk_src
	SsStop      uint16 // Stop value to get frequency change.
	SsEnable    uint8  // Enable spread spectrum modulation
	SsStep      uint16 // Step value to get frequency change step.
}

func (cfg ClockConfigSysPll) Configure(pfd ...uint32) {

	// bypass PLL first
	src := (uint32(cfg.Src) << CCM_ANALOG_PLL_SYS_BYPASS_CLK_SRC_Pos) & CCM_ANALOG_PLL_SYS_BYPASS_CLK_SRC_Msk
	CCM_ANALOG.PLL_SYS.Set(
		(CCM_ANALOG.PLL_SYS.Get() & ^uint32(CCM_ANALOG_PLL_SYS_BYPASS_CLK_SRC_Msk)) |
			CCM_ANALOG_PLL_SYS_BYPASS_Msk | src)

	sel := (uint32(cfg.LoopDivider) << CCM_ANALOG_PLL_SYS_DIV_SELECT_Pos) & CCM_ANALOG_PLL_SYS_DIV_SELECT_Msk
	CCM_ANALOG.PLL_SYS.Set(
		(CCM_ANALOG.PLL_SYS.Get() & ^uint32(CCM_ANALOG_PLL_SYS_DIV_SELECT_Msk|CCM_ANALOG_PLL_SYS_POWERDOWN_Msk)) |
			CCM_ANALOG_PLL_SYS_ENABLE_Msk | sel)

	// initialize the fractional mode
	CCM_ANALOG.PLL_SYS_NUM.Set((cfg.Numerator << CCM_ANALOG_PLL_SYS_NUM_A_Pos) & CCM_ANALOG_PLL_SYS_NUM_A_Msk)
	CCM_ANALOG.PLL_SYS_DENOM.Set((cfg.Denominator << CCM_ANALOG_PLL_SYS_DENOM_B_Pos) & CCM_ANALOG_PLL_SYS_DENOM_B_Msk)

	// initialize the spread spectrum mode
	inc := (uint32(cfg.SsStep) << CCM_ANALOG_PLL_SYS_SS_STEP_Pos) & CCM_ANALOG_PLL_SYS_SS_STEP_Msk
	enb := (uint32(cfg.SsEnable) << CCM_ANALOG_PLL_SYS_SS_ENABLE_Pos) & CCM_ANALOG_PLL_SYS_SS_ENABLE_Msk
	stp := (uint32(cfg.SsStop) << CCM_ANALOG_PLL_SYS_SS_STOP_Pos) & CCM_ANALOG_PLL_SYS_SS_STOP_Msk
	CCM_ANALOG.PLL_SYS_SS.Set(inc | enb | stp)

	for !CCM_ANALOG.PLL_SYS.HasBits(CCM_ANALOG_PLL_SYS_LOCK_Msk) {
	}

	// disable bypass
	CCM_ANALOG.PLL_SYS.ClearBits(CCM_ANALOG_PLL_SYS_BYPASS_Msk)

	// update PFDs after update
	setSysPfd(pfd...)
}

// PLL configuration for USB
type ClockConfigUsbPll struct {
	Instance    uint8 // USB PLL number (1 or 2)
	LoopDivider uint8 // PLL loop divider: 0 - Fout=Fref*20, 1 - Fout=Fref*22
	Src         uint8 // Pll clock source, reference _clock_pll_clk_src
}

func (cfg ClockConfigUsbPll) Configure(pfd ...uint32) {

	switch cfg.Instance {
	case 1:

		// bypass PLL first
		src := (uint32(cfg.Src) << CCM_ANALOG_PLL_USB1_BYPASS_CLK_SRC_Pos) & CCM_ANALOG_PLL_USB1_BYPASS_CLK_SRC_Msk
		CCM_ANALOG.PLL_USB1.Set(
			(CCM_ANALOG.PLL_USB1.Get() & ^uint32(CCM_ANALOG_PLL_USB1_BYPASS_CLK_SRC_Msk)) |
				CCM_ANALOG_PLL_USB1_BYPASS_Msk | src)

		sel := uint32((cfg.LoopDivider << CCM_ANALOG_PLL_USB1_DIV_SELECT_Pos) & CCM_ANALOG_PLL_USB1_DIV_SELECT_Msk)
		CCM_ANALOG.PLL_USB1_SET.Set(
			(CCM_ANALOG.PLL_USB1.Get() & ^uint32(CCM_ANALOG_PLL_USB1_DIV_SELECT_Msk)) |
				CCM_ANALOG_PLL_USB1_ENABLE_Msk | CCM_ANALOG_PLL_USB1_POWER_Msk |
				CCM_ANALOG_PLL_USB1_EN_USB_CLKS_Msk | sel)

		for !CCM_ANALOG.PLL_USB1.HasBits(CCM_ANALOG_PLL_USB1_LOCK_Msk) {
		}

		// disable bypass
		CCM_ANALOG.PLL_USB1_CLR.Set(CCM_ANALOG_PLL_USB1_BYPASS_Msk)

		// update PFDs after update
		setUsb1Pfd(pfd...)

	case 2:
		// bypass PLL first
		src := (uint32(cfg.Src) << CCM_ANALOG_PLL_USB2_BYPASS_CLK_SRC_Pos) & CCM_ANALOG_PLL_USB2_BYPASS_CLK_SRC_Msk
		CCM_ANALOG.PLL_USB2.Set(
			(CCM_ANALOG.PLL_USB2.Get() & ^uint32(CCM_ANALOG_PLL_USB2_BYPASS_CLK_SRC_Msk)) |
				CCM_ANALOG_PLL_USB2_BYPASS_Msk | src)

		sel := uint32((cfg.LoopDivider << CCM_ANALOG_PLL_USB2_DIV_SELECT_Pos) & CCM_ANALOG_PLL_USB2_DIV_SELECT_Msk)
		CCM_ANALOG.PLL_USB2.Set(
			(CCM_ANALOG.PLL_USB2.Get() & ^uint32(CCM_ANALOG_PLL_USB2_DIV_SELECT_Msk)) |
				CCM_ANALOG_PLL_USB2_ENABLE_Msk | CCM_ANALOG_PLL_USB2_POWER_Msk |
				CCM_ANALOG_PLL_USB2_EN_USB_CLKS_Msk | sel)

		for !CCM_ANALOG.PLL_USB2.HasBits(CCM_ANALOG_PLL_USB2_LOCK_Msk) {
		}

		// disable bypass
		CCM_ANALOG.PLL_USB2.ClearBits(CCM_ANALOG_PLL_USB2_BYPASS_Msk)

	default:
		panic("nxp: invalid USB PLL")
	}
}
