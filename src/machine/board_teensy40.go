//go:build teensy40
// +build teensy40

package machine

import (
	"runtime/interrupt"

	"tinygo.org/x/device/nxp"
)

// Digital pins
const (
	//  = Pin  // [Pad]:       Alt Func 0       Alt Func 1       Alt Func 2       Alt Func 3      Alt Func 4            Alt Func 5  Alt Func 6            Alt Func 7             Alt Func 8             Alt Func 9
	//  = ----    -----------  ---------------  ---------------  ---------------  --------------  --------------------  ----------  --------------------  ---------------------  ---------------------  ----------------
	D0  = PA3  // [AD_B0_03]:  FLEXCAN2_RX      XBAR1_INOUT17    LPUART6_RX       USB_OTG1_OC     FLEXPWM1_PWMX01       GPIO1_IO03  REF_CLK_24M           LPSPI3_PCS0             ~                      ~
	D1  = PA2  // [AD_B0_02]:  FLEXCAN2_TX      XBAR1_INOUT16    LPUART6_TX       USB_OTG1_PWR    FLEXPWM1_PWMX00       GPIO1_IO02  LPI2C1_HREQ           LPSPI3_SDI              ~                      ~
	D2  = PD4  // [EMC_04]:    SEMC_DATA04      FLEXPWM4_PWMA02  SAI2_TX_DATA     XBAR1_INOUT06   FLEXIO1_FLEXIO04      GPIO4_IO04   ~                     ~                      ~                      ~
	D3  = PD5  // [EMC_05]:    SEMC_DATA05      FLEXPWM4_PWMB02  SAI2_TX_SYNC     XBAR1_INOUT07   FLEXIO1_FLEXIO05      GPIO4_IO05   ~                     ~                      ~                      ~
	D4  = PD6  // [EMC_06]:    SEMC_DATA06      FLEXPWM2_PWMA00  SAI2_TX_BCLK     XBAR1_INOUT08   FLEXIO1_FLEXIO06      GPIO4_IO06   ~                     ~                      ~                      ~
	D5  = PD8  // [EMC_08]:    SEMC_DM00        FLEXPWM2_PWMA01  SAI2_RX_DATA     XBAR1_INOUT17   FLEXIO1_FLEXIO08      GPIO4_IO08   ~                     ~                      ~                      ~
	D6  = PB10 // [B0_10]:     LCD_DATA06       QTIMER4_TIMER1   FLEXPWM2_PWMA02  SAI1_TX_DATA03  FLEXIO2_FLEXIO10      GPIO2_IO10  SRC_BOOT_CFG06        ENET2_CRS               ~                      ~
	D7  = PB17 // [B1_01]:     LCD_DATA13       XBAR1_INOUT15    LPUART4_RX       SAI1_TX_DATA00  FLEXIO2_FLEXIO17      GPIO2_IO17  FLEXPWM1_PWMB03       ENET2_RDATA00          FLEXIO3_FLEXIO17        ~
	D8  = PB16 // [B1_00]:     LCD_DATA12       XBAR1_INOUT14    LPUART4_TX       SAI1_RX_DATA00  FLEXIO2_FLEXIO16      GPIO2_IO16  FLEXPWM1_PWMA03       ENET2_RX_ER            FLEXIO3_FLEXIO16        ~
	D9  = PB11 // [B0_11]:     LCD_DATA07       QTIMER4_TIMER2   FLEXPWM2_PWMB02  SAI1_TX_DATA02  FLEXIO2_FLEXIO11      GPIO2_IO11  SRC_BOOT_CFG07        ENET2_COL               ~                      ~
	D10 = PB0  // [B0_00]:     LCD_CLK          QTIMER1_TIMER0   MQS_RIGHT        LPSPI4_PCS0     FLEXIO2_FLEXIO00      GPIO2_IO00  SEMC_CSX01            ENET2_MDC               ~                      ~
	D11 = PB2  // [B0_02]:     LCD_HSYNC        QTIMER1_TIMER2   FLEXCAN1_TX      LPSPI4_SDO      FLEXIO2_FLEXIO02      GPIO2_IO02  SEMC_CSX03            ENET2_1588_EVENT0_OUT   ~                      ~
	D12 = PB1  // [B0_01]:     LCD_ENABLE       QTIMER1_TIMER1   MQS_LEFT         LPSPI4_SDI      FLEXIO2_FLEXIO01      GPIO2_IO01  SEMC_CSX02            ENET2_MDIO              ~                      ~
	D13 = PB3  // [B0_03]:     LCD_VSYNC        QTIMER2_TIMER0   FLEXCAN1_RX      LPSPI4_SCK      FLEXIO2_FLEXIO03      GPIO2_IO03  WDOG2_RESET_B_DEB     ENET2_1588_EVENT0_IN    ~                      ~
	D14 = PA18 // [AD_B1_02]:  USB_OTG1_ID      QTIMER3_TIMER2   LPUART2_TX       SPDIF_OUT       ENET_1588_EVENT2_OUT  GPIO1_IO18  USDHC1_CD_B           KPP_ROW06              GPT2_CLK               FLEXIO3_FLEXIO02
	D15 = PA19 // [AD_B1_03]:  USB_OTG1_OC      QTIMER3_TIMER3   LPUART2_RX       SPDIF_IN        ENET_1588_EVENT2_IN   GPIO1_IO19  USDHC2_CD_B           KPP_COL06              GPT2_CAPTURE1          FLEXIO3_FLEXIO03
	D16 = PA23 // [AD_B1_07]:  FLEXSPIB_DATA00  LPI2C3_SCL       LPUART3_RX       SPDIF_EXT_CLK   CSI_HSYNC             GPIO1_IO23  USDHC2_DATA3          KPP_COL04              GPT2_COMPARE3          FLEXIO3_FLEXIO07
	D17 = PA22 // [AD_B1_06]:  FLEXSPIB_DATA01  LPI2C3_SDA       LPUART3_TX       SPDIF_LOCK      CSI_VSYNC             GPIO1_IO22  USDHC2_DATA2          KPP_ROW04              GPT2_COMPARE2          FLEXIO3_FLEXIO06
	D18 = PA17 // [AD_B1_01]:  USB_OTG1_PWR     QTIMER3_TIMER1   LPUART2_RTS_B    LPI2C1_SDA      CCM_PMIC_READY        GPIO1_IO17  USDHC1_VSELECT        KPP_COL07              ENET2_1588_EVENT0_IN   FLEXIO3_FLEXIO01
	D19 = PA16 // [AD_B1_00]:  USB_OTG2_ID      QTIMER3_TIMER0   LPUART2_CTS_B    LPI2C1_SCL      WDOG1_B               GPIO1_IO16  USDHC1_WP             KPP_ROW07              ENET2_1588_EVENT0_OUT  FLEXIO3_FLEXIO00
	D20 = PA26 // [AD_B1_10]:  FLEXSPIA_DATA03  WDOG1_B          LPUART8_TX       SAI1_RX_SYNC    CSI_DATA07            GPIO1_IO26  USDHC2_WP             KPP_ROW02              ENET2_1588_EVENT1_OUT  FLEXIO3_FLEXIO10
	D21 = PA27 // [AD_B1_11]:  FLEXSPIA_DATA02  EWM_OUT_B        LPUART8_RX       SAI1_RX_BCLK    CSI_DATA06            GPIO1_IO27  USDHC2_RESET_B        KPP_COL02              ENET2_1588_EVENT1_IN   FLEXIO3_FLEXIO11
	D22 = PA24 // [AD_B1_08]:  FLEXSPIA_SS1_B   FLEXPWM4_PWMA00  FLEXCAN1_TX      CCM_PMIC_READY  CSI_DATA09            GPIO1_IO24  USDHC2_CMD            KPP_ROW03              FLEXIO3_FLEXIO08        ~
	D23 = PA25 // [AD_B1_09]:  FLEXSPIA_DQS     FLEXPWM4_PWMA01  FLEXCAN1_RX      SAI1_MCLK       CSI_DATA08            GPIO1_IO25  USDHC2_CLK            KPP_COL03              FLEXIO3_FLEXIO09        ~
	D24 = PA12 // [AD_B0_12]:  LPI2C4_SCL       CCM_PMIC_READY   LPUART1_TX       WDOG2_WDOG_B    FLEXPWM1_PWMX02       GPIO1_IO12  ENET_1588_EVENT1_OUT  NMI_GLUE_NMI            ~                      ~
	D25 = PA13 // [AD_B0_13]:  LPI2C4_SDA       GPT1_CLK         LPUART1_RX       EWM_OUT_B       FLEXPWM1_PWMX03       GPIO1_IO13  ENET_1588_EVENT1_IN   REF_CLK_24M             ~                      ~
	D26 = PA30 // [AD_B1_14]:  FLEXSPIA_SCLK    ACMP_OUT02       LPSPI3_SDO       SAI1_TX_BCLK    CSI_DATA03            GPIO1_IO30  USDHC2_DATA6          KPP_ROW00              ENET2_1588_EVENT3_OUT  FLEXIO3_FLEXIO14
	D27 = PA31 // [AD_B1_15]:  FLEXSPIA_SS0_B   ACMP_OUT03       LPSPI3_SCK       SAI1_TX_SYNC    CSI_DATA02            GPIO1_IO31  USDHC2_DATA7          KPP_COL00              ENET2_1588_EVENT3_IN   FLEXIO3_FLEXIO15
	D28 = PC18 // [EMC_32]:    SEMC_DATA10      FLEXPWM3_PWMB01  LPUART7_RX       CCM_PMIC_RDY    CSI_DATA21            GPIO3_IO18  ENET2_TX_EN            ~                      ~                      ~
	D29 = PD31 // [EMC_31]:    SEMC_DATA09      FLEXPWM3_PWMA01  LPUART7_TX       LPSPI1_PCS1     CSI_DATA22            GPIO4_IO31  ENET2_TDATA01          ~                      ~                      ~
	D30 = PC23 // [EMC_37]:    SEMC_DATA15      XBAR1_IN23       GPT1_COMPARE3    SAI3_MCLK       CSI_DATA16            GPIO3_IO23  USDHC2_WP             ENET2_RX_EN            FLEXCAN3_RX             ~
	D31 = PC22 // [EMC_36]:    SEMC_DATA14      XBAR1_IN22       GPT1_COMPARE2    SAI3_TX_DATA    CSI_DATA17            GPIO3_IO22  USDHC1_WP             ENET2_RDATA01          FLEXCAN3_TX             ~
	D32 = PB12 // [B0_12]:     LCD_DATA08       XBAR1_INOUT10    ARM_TRACE_CLK    SAI1_TX_DATA01  FLEXIO2_FLEXIO12      GPIO2_IO12  SRC_BOOT_CFG08        ENET2_TDATA00           ~                      ~
	D33 = PD7  // [EMC_07]:    SEMC_DATA07      FLEXPWM2_PWMB00  SAI2_MCLK        XBAR1_INOUT09   FLEXIO1_FLEXIO07      GPIO4_IO07   ~                     ~                      ~                      ~
	D34 = PC15 // [SD_B0_03]:  USDHC1_DATA1     FLEXPWM1_PWMB01  LPUART8_RTS_B    XBAR1_INOUT07   LPSPI1_SDI            GPIO3_IO15  ENET2_RDATA00         SEMC_CLK6               ~                      ~
	D35 = PC14 // [SD_B0_02]:  USDHC1_DATA0     FLEXPWM1_PWMA01  LPUART8_CTS_B    XBAR1_INOUT06   LPSPI1_SDO            GPIO3_IO14  ENET2_RX_ER           SEMC_CLK5               ~                      ~
	D36 = PC13 // [SD_B0_01]:  USDHC1_CLK       FLEXPWM1_PWMB00  LPI2C3_SDA       XBAR1_INOUT05   LPSPI1_PCS0           GPIO3_IO13  FLEXSPIB_SS1_B        ENET2_TX_CLK           ENET2_REF_CLK2          ~
	D37 = PC12 // [SD_B0_00]:  USDHC1_CMD       FLEXPWM1_PWMA00  LPI2C3_SCL       XBAR1_INOUT04   LPSPI1_SCK            GPIO3_IO12  FLEXSPIA_SS1_B        ENET2_TX_EN            SEMC_DQS4               ~
	D38 = PC17 // [SD_B0_05]:  USDHC1_DATA3     FLEXPWM1_PWMB02  LPUART8_RX       XBAR1_INOUT09   FLEXSPIB_DQS          GPIO3_IO17  CCM_CLKO2             ENET2_RX_EN             ~                      ~
	D39 = PC16 // [SD_B0_04]:  USDHC1_DATA2     FLEXPWM1_PWMA02  LPUART8_TX       XBAR1_INOUT08   FLEXSPIB_SS0_B        GPIO3_IO16  CCM_CLKO1             ENET2_RDATA01           ~                      ~
)

// Analog pins
const (
	//  = Pin  // Dig | [Pad]      {ADC1/ADC2}
	A0  = PA18 // D14 | [AD_B1_02] {  7 / 7  }
	A1  = PA19 // D15 | [AD_B1_03] {  8 / 8  }
	A2  = PA23 // D16 | [AD_B1_07] { 12 / 12 }
	A3  = PA22 // D17 | [AD_B1_06] { 11 / 11 }
	A4  = PA17 // D18 | [AD_B1_01] {  6 / 6  }
	A5  = PA16 // D19 | [AD_B1_00] {  5 / 5  }
	A6  = PA26 // D20 | [AD_B1_10] { 15 / 15 }
	A7  = PA27 // D21 | [AD_B1_11] {  0 / 0  }
	A8  = PA24 // D22 | [AD_B1_08] { 13 / 13 }
	A9  = PA25 // D23 | [AD_B1_09] { 14 / 14 }
	A10 = PA12 // D24 | [AD_B0_12] {  1 / -  }
	A11 = PA13 // D25 | [AD_B0_13] {  2 / -  }
	A12 = PA30 // D26 | [AD_B1_14] {  - / 3  }
	A13 = PA31 // D27 | [AD_B1_15] {  - / 4  }
)

// Default peripheral pins
const (
	LED = D13

	UART_RX_PIN = UART1_RX_PIN // D0
	UART_TX_PIN = UART1_TX_PIN // D1

	SPI_SDI_PIN = SPI1_SDI_PIN // D12
	SPI_SDO_PIN = SPI1_SDO_PIN // D11
	SPI_SCK_PIN = SPI1_SCK_PIN // D13
	SPI_CS_PIN  = SPI1_CS_PIN  // D10

	I2C_SDA_PIN = I2C1_SDA_PIN // D18/A4
	I2C_SCL_PIN = I2C1_SCL_PIN // D19/A5
)

// Default peripherals
var (
	DefaultUART = UART1
)

func init() {
	// register any interrupt handlers for this board's peripherals
	_UART1.Interrupt = interrupt.New(nxp.IRQ_LPUART6, _UART1.handleInterrupt)
	_UART2.Interrupt = interrupt.New(nxp.IRQ_LPUART4, _UART2.handleInterrupt)
	_UART3.Interrupt = interrupt.New(nxp.IRQ_LPUART2, _UART3.handleInterrupt)
	_UART4.Interrupt = interrupt.New(nxp.IRQ_LPUART3, _UART4.handleInterrupt)
	_UART5.Interrupt = interrupt.New(nxp.IRQ_LPUART8, _UART5.handleInterrupt)
	_UART6.Interrupt = interrupt.New(nxp.IRQ_LPUART1, _UART6.handleInterrupt)
	_UART7.Interrupt = interrupt.New(nxp.IRQ_LPUART7, _UART7.handleInterrupt)
}

// #=====================================================#
// |                        UART                         |
// #===========#===========#=============#===============#
// | Interface | Hardware  | Clock(Freq) |  RX/TX  : Alt |
// #===========#===========#=============#=========-=====#
// |   UART1   |  LPUART6  | OSC(24 MHz) |  D0/D1  : 2/2 |
// |   UART2   |  LPUART4  | OSC(24 MHz) |  D7/D8  : 2/2 |
// |   UART3   |  LPUART2  | OSC(24 MHz) | D15/D14 : 2/2 |
// |   UART4   |  LPUART3  | OSC(24 MHz) | D16/D17 : 2/2 |
// |   UART5   |  LPUART8  | OSC(24 MHz) | D21/D20 : 2/2 |
// |   UART6   |  LPUART1  | OSC(24 MHz) | D25/D24 : 2/2 |
// |   UART7   |  LPUART7  | OSC(24 MHz) | D28/D29 : 2/2 |
// #===========#===========#=============#=========-=====#
const (
	UART1_RX_PIN = D0
	UART1_TX_PIN = D1

	UART2_RX_PIN = D7
	UART2_TX_PIN = D8

	UART3_RX_PIN = D15
	UART3_TX_PIN = D14

	UART4_RX_PIN = D16
	UART4_TX_PIN = D17

	UART5_RX_PIN = D21
	UART5_TX_PIN = D20

	UART6_RX_PIN = D25
	UART6_TX_PIN = D24

	UART7_RX_PIN = D28
	UART7_TX_PIN = D29
)

var (
	UART1  = &_UART1
	_UART1 = UART{
		Bus:      nxp.LPUART6,
		Buffer:   NewRingBuffer(),
		txBuffer: NewRingBuffer(),
		muxRX: muxSelect{ // D0 (PA3 [AD_B0_03])
			mux: nxp.IOMUXC_LPUART6_RX_SELECT_INPUT_DAISY_GPIO_AD_B0_03_ALT2,
			sel: &nxp.IOMUXC.LPUART6_RX_SELECT_INPUT,
		},
		muxTX: muxSelect{ // D1 (PA2 [AD_B0_02])
			mux: nxp.IOMUXC_LPUART6_TX_SELECT_INPUT_DAISY_GPIO_AD_B0_02_ALT2,
			sel: &nxp.IOMUXC.LPUART6_TX_SELECT_INPUT,
		},
	}
	UART2  = &_UART2
	_UART2 = UART{
		Bus:      nxp.LPUART4,
		Buffer:   NewRingBuffer(),
		txBuffer: NewRingBuffer(),
		muxRX: muxSelect{ // D7 (PB17 [B1_01])
			mux: nxp.IOMUXC_LPUART4_RX_SELECT_INPUT_DAISY_GPIO_B1_01_ALT2,
			sel: &nxp.IOMUXC.LPUART4_RX_SELECT_INPUT,
		},
		muxTX: muxSelect{ // D8 (PB16 [B1_00])
			mux: nxp.IOMUXC_LPUART4_TX_SELECT_INPUT_DAISY_GPIO_B1_00_ALT2,
			sel: &nxp.IOMUXC.LPUART4_TX_SELECT_INPUT,
		},
	}
	UART3  = &_UART3
	_UART3 = UART{
		Bus:      nxp.LPUART2,
		Buffer:   NewRingBuffer(),
		txBuffer: NewRingBuffer(),
		muxRX: muxSelect{ // D15 (PA19 [AD_B1_03])
			mux: nxp.IOMUXC_LPUART2_RX_SELECT_INPUT_DAISY_GPIO_AD_B1_03_ALT2,
			sel: &nxp.IOMUXC.LPUART2_RX_SELECT_INPUT,
		},
		muxTX: muxSelect{ // D14 (PA18 [AD_B1_02])
			mux: nxp.IOMUXC_LPUART2_TX_SELECT_INPUT_DAISY_GPIO_AD_B1_02_ALT2,
			sel: &nxp.IOMUXC.LPUART2_TX_SELECT_INPUT,
		},
	}
	UART4  = &_UART4
	_UART4 = UART{
		Bus:      nxp.LPUART3,
		Buffer:   NewRingBuffer(),
		txBuffer: NewRingBuffer(),
		muxRX: muxSelect{ // D16 (PA23 [AD_B1_07])
			mux: nxp.IOMUXC_LPUART3_RX_SELECT_INPUT_DAISY_GPIO_AD_B1_07_ALT2,
			sel: &nxp.IOMUXC.LPUART3_RX_SELECT_INPUT,
		},
		muxTX: muxSelect{ // D17 (PA22 [AD_B1_06])
			mux: nxp.IOMUXC_LPUART3_TX_SELECT_INPUT_DAISY_GPIO_AD_B1_06_ALT2,
			sel: &nxp.IOMUXC.LPUART3_TX_SELECT_INPUT,
		},
	}
	UART5  = &_UART5
	_UART5 = UART{
		Bus:      nxp.LPUART8,
		Buffer:   NewRingBuffer(),
		txBuffer: NewRingBuffer(),
		muxRX: muxSelect{ // D21 (PA27 [AD_B1_11])
			mux: nxp.IOMUXC_LPUART8_RX_SELECT_INPUT_DAISY_GPIO_AD_B1_11_ALT2,
			sel: &nxp.IOMUXC.LPUART8_RX_SELECT_INPUT,
		},
		muxTX: muxSelect{ // D20 (PA26 [AD_B1_10])
			mux: nxp.IOMUXC_LPUART8_TX_SELECT_INPUT_DAISY_GPIO_AD_B1_10_ALT2,
			sel: &nxp.IOMUXC.LPUART8_TX_SELECT_INPUT,
		},
	}
	UART6  = &_UART6
	_UART6 = UART{
		Bus:      nxp.LPUART1,
		Buffer:   NewRingBuffer(),
		txBuffer: NewRingBuffer(),
		// LPUART1 not connected via IOMUXC
		//   RX: D24 (PA12 [AD_B0_12])
		//   TX: D25 (PA13 [AD_B0_13])
	}
	UART7  = &_UART7
	_UART7 = UART{
		Bus:      nxp.LPUART7,
		Buffer:   NewRingBuffer(),
		txBuffer: NewRingBuffer(),
		muxRX: muxSelect{ // D28 (PC18 [EMC_32])
			mux: nxp.IOMUXC_LPUART7_RX_SELECT_INPUT_DAISY_GPIO_EMC_32_ALT2,
			sel: &nxp.IOMUXC.LPUART7_RX_SELECT_INPUT,
		},
		muxTX: muxSelect{ // D29 (PD31 [EMC_31])
			mux: nxp.IOMUXC_LPUART7_TX_SELECT_INPUT_DAISY_GPIO_EMC_31_ALT2,
			sel: &nxp.IOMUXC.LPUART7_TX_SELECT_INPUT,
		},
	}
)

// #==================================================================#
// |                               SPI                                |
// #===========#==========#===============#===========================#
// | Interface | Hardware |  Clock(Freq)  | SDI/SDO/SCK/CS  :   Alt   |
// #===========#==========#===============#=================-=========#
// |   SPI1    |  LPSPI4  | PLL2(132 MHz) | D12/D11/D13/D10 : 3/3/3/3 |
// |   SPI2    |  LPSPI3  | PLL2(132 MHz) |  D1/D26/D27/D0  : 7/2/2/7 |
// |   SPI3    |  LPSPI1  | PLL2(132 MHz) | D34/D35/D37/D36 : 4/4/4/4 |
// #===========#==========#===============#=================-=========#
const (
	SPI1_SDI_PIN = D12
	SPI1_SDO_PIN = D11
	SPI1_SCK_PIN = D13
	SPI1_CS_PIN  = D10

	SPI2_SDI_PIN = D1
	SPI2_SDO_PIN = D26
	SPI2_SCK_PIN = D27
	SPI2_CS_PIN  = D0

	SPI3_SDI_PIN = D34
	SPI3_SDO_PIN = D35
	SPI3_SCK_PIN = D37
	SPI3_CS_PIN  = D36
)

var (
	SPI0  = SPI1 // SPI0 is an alias of SPI1 (LPSPI4)
	SPI1  = &_SPI1
	_SPI1 = SPI{
		Bus: nxp.LPSPI4,
		muxSDI: muxSelect{ // D12 (PB1 [B0_01])
			mux: nxp.IOMUXC_LPSPI4_SDI_SELECT_INPUT_DAISY_GPIO_B0_01_ALT3,
			sel: &nxp.IOMUXC.LPSPI4_SDI_SELECT_INPUT,
		},
		muxSDO: muxSelect{ // D11 (PB2 [B0_02])
			mux: nxp.IOMUXC_LPSPI4_SDO_SELECT_INPUT_DAISY_GPIO_B0_02_ALT3,
			sel: &nxp.IOMUXC.LPSPI4_SDO_SELECT_INPUT,
		},
		muxSCK: muxSelect{ // D13 (PB3 [B0_03])
			mux: nxp.IOMUXC_LPSPI4_SCK_SELECT_INPUT_DAISY_GPIO_B0_03_ALT3,
			sel: &nxp.IOMUXC.LPSPI4_SCK_SELECT_INPUT,
		},
		muxCS: muxSelect{ // D10 (PB0 [B0_00])
			mux: nxp.IOMUXC_LPSPI4_PCS0_SELECT_INPUT_DAISY_GPIO_B0_00_ALT3,
			sel: &nxp.IOMUXC.LPSPI4_PCS0_SELECT_INPUT,
		},
	}
	SPI2  = &_SPI2
	_SPI2 = SPI{
		Bus: nxp.LPSPI3,
		muxSDI: muxSelect{ // D1 (PA2 [AD_B0_02])
			mux: nxp.IOMUXC_LPSPI3_SDI_SELECT_INPUT_DAISY_GPIO_AD_B0_02_ALT7,
			sel: &nxp.IOMUXC.LPSPI3_SDI_SELECT_INPUT,
		},
		muxSDO: muxSelect{ // D26 (PA30 [AD_B1_14])
			mux: nxp.IOMUXC_LPSPI3_SDO_SELECT_INPUT_DAISY_GPIO_AD_B1_14_ALT2,
			sel: &nxp.IOMUXC.LPSPI3_SDO_SELECT_INPUT,
		},
		muxSCK: muxSelect{ // D27 (PA31 [AD_B1_15])
			mux: nxp.IOMUXC_LPSPI3_SCK_SELECT_INPUT_DAISY_GPIO_AD_B1_15,
			sel: &nxp.IOMUXC.LPSPI3_SCK_SELECT_INPUT,
		},
		muxCS: muxSelect{ // D0 (PA3 [AD_B0_03])
			mux: nxp.IOMUXC_LPSPI3_PCS0_SELECT_INPUT_DAISY_GPIO_AD_B0_03_ALT7,
			sel: &nxp.IOMUXC.LPSPI3_PCS0_SELECT_INPUT,
		},
	}
	SPI3  = &_SPI3
	_SPI3 = SPI{
		Bus: nxp.LPSPI1,
		muxSDI: muxSelect{ // D34 (PC15 [SD_B0_03])
			mux: nxp.IOMUXC_LPSPI1_SDI_SELECT_INPUT_DAISY_GPIO_SD_B0_03_ALT4,
			sel: &nxp.IOMUXC.LPSPI1_SDI_SELECT_INPUT,
		},
		muxSDO: muxSelect{ // D35 (PC14 [SD_B0_02])
			mux: nxp.IOMUXC_LPSPI1_SDO_SELECT_INPUT_DAISY_GPIO_SD_B0_02_ALT4,
			sel: &nxp.IOMUXC.LPSPI1_SDO_SELECT_INPUT,
		},
		muxSCK: muxSelect{ // D37 (PC12 [SD_B0_00])
			mux: nxp.IOMUXC_LPSPI1_SCK_SELECT_INPUT_DAISY_GPIO_SD_B0_00_ALT4,
			sel: &nxp.IOMUXC.LPSPI1_SCK_SELECT_INPUT,
		},
		muxCS: muxSelect{ // D36 (PC13 [SD_B0_01])
			mux: nxp.IOMUXC_LPSPI1_PCS0_SELECT_INPUT_DAISY_GPIO_SD_B0_01_ALT4,
			sel: &nxp.IOMUXC.LPSPI1_PCS0_SELECT_INPUT,
		},
	}
)

// #====================================================#
// |                         I2C                        |
// #===========#==========#=============#===============#
// | Interface | Hardware | Clock(Freq) | SDA/SCL : Alt |
// #===========#==========#=============#=========-=====#
// |   I2C1    |  LPI2C1  | OSC(24 MHz) | D18/D19 : 3/3 |
// |   I2C2    |  LPI2C3  | OSC(24 MHz) | D17/D16 : 1/1 |
// |   I2C3    |  LPI2C4  | OSC(24 MHz) | D25/D24 : 0/0 |
// #===========#==========#=============#=========-=====#
const (
	I2C1_SDA_PIN = D18
	I2C1_SCL_PIN = D19

	I2C2_SDA_PIN = D17
	I2C2_SCL_PIN = D16

	I2C3_SDA_PIN = D25
	I2C3_SCL_PIN = D24
)

var (
	I2C0  = I2C1 // I2C0 is an alias for I2C1 (LPI2C1)
	I2C1  = &_I2C1
	_I2C1 = I2C{
		Bus: nxp.LPI2C1,
		sda: I2C1_SDA_PIN, // D18 (PA17 [AD_B1_01])
		scl: I2C1_SCL_PIN, // D19 (PA16 [AD_B1_00])
		muxSDA: muxSelect{
			mux: nxp.IOMUXC_LPI2C1_SDA_SELECT_INPUT_DAISY_GPIO_AD_B1_01_ALT3,
			sel: &nxp.IOMUXC.LPI2C1_SDA_SELECT_INPUT,
		},
		muxSCL: muxSelect{
			mux: nxp.IOMUXC_LPI2C1_SCL_SELECT_INPUT_DAISY_GPIO_AD_B1_00_ALT3,
			sel: &nxp.IOMUXC.LPI2C1_SCL_SELECT_INPUT,
		},
	}
	I2C2  = &_I2C2
	_I2C2 = I2C{
		Bus: nxp.LPI2C3,
		sda: I2C2_SDA_PIN, // D17 (PA22 [AD_B1_06])
		scl: I2C2_SCL_PIN, // D16 (PA23 [AD_B1_07])
		muxSDA: muxSelect{
			mux: nxp.IOMUXC_LPI2C3_SDA_SELECT_INPUT_DAISY_GPIO_AD_B1_06_ALT1,
			sel: &nxp.IOMUXC.LPI2C3_SDA_SELECT_INPUT,
		},
		muxSCL: muxSelect{
			mux: nxp.IOMUXC_LPI2C3_SCL_SELECT_INPUT_DAISY_GPIO_AD_B1_07_ALT1,
			sel: &nxp.IOMUXC.LPI2C3_SCL_SELECT_INPUT,
		},
	}
	I2C3  = &_I2C3
	_I2C3 = I2C{
		Bus: nxp.LPI2C4,
		sda: I2C3_SDA_PIN, // D25 (PA13 [AD_B0_13])
		scl: I2C3_SCL_PIN, // D24 (PA12 [AD_B0_12])
		muxSDA: muxSelect{
			mux: nxp.IOMUXC_LPI2C4_SDA_SELECT_INPUT_DAISY_GPIO_AD_B0_13_ALT0,
			sel: &nxp.IOMUXC.LPI2C4_SDA_SELECT_INPUT,
		},
		muxSCL: muxSelect{
			mux: nxp.IOMUXC_LPI2C4_SCL_SELECT_INPUT_DAISY_GPIO_AD_B0_12_ALT0,
			sel: &nxp.IOMUXC.LPI2C4_SCL_SELECT_INPUT,
		},
	}
)
