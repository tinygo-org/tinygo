//go:build mimxrt1062
// +build mimxrt1062

package machine

import (
	"device/nxp"
	"math/bits"
	"runtime/interrupt"
	"runtime/volatile"
)

// Peripheral abstraction layer for the MIMXRT1062

const deviceName = nxp.Device

func CPUFrequency() uint32 {
	return 600000000
}

const (
	// GPIO
	PinInput PinMode = iota
	PinInputPullUp
	PinInputPullDown
	PinOutput
	PinOutputOpenDrain
	PinDisable

	// ADC
	PinInputAnalog

	// UART
	PinModeUARTTX
	PinModeUARTRX

	// SPI
	PinModeSPISDI
	PinModeSPISDO
	PinModeSPICLK
	PinModeSPICS

	// I2C
	PinModeI2CSDA
	PinModeI2CSCL
)

type PinChange uint8

const (
	PinRising PinChange = iota + 2
	PinFalling
	PinToggle
)

// pinJumpTable represents a function lookup table for all 128 GPIO pins.
//
// There are 4 GPIO ports (A-D) and 32 pins (0-31) on each port. The uint8 value
// of a Pin is used as table index. The number of pins with a defined (non-nil)
// function is recorded in the uint8 field numDefined.
type pinJumpTable struct {
	lut        [4 * 32]func(Pin)
	numDefined uint8
}

// pinISR stores the interrupt callbacks for GPIO pins, and pinInterrupt holds
// an interrupt service routine that dispatches the interrupt callbacks.
var (
	pinISR       pinJumpTable
	pinInterrupt *interrupt.Interrupt
)

// From the i.MXRT1062 Processor Reference Manual (Chapter 12 - GPIO):
//
// |  High-speed GPIOs exist in this device:
// |    - GPIO1-5 are standard-speed GPIOs that run off the IPG_CLK_ROOT, while
// |      GPIO6-9 are high-speed GPIOs that run at the AHB_CLK_ROOT frequency.
// |      See the table "System Clocks, Gating, and Override" in CCM chapter.
// |    - Regular GPIO and high speed GPIO are paired (GPIO1 and GPIO6 share the
// |      same pins, GPIO2 and GPIO7 share, etc). The IOMUXC_GPR_GPR26-29
// |      registers are used to determine if the regular or high-speed GPIO
// |      module is used for the GPIO pins on a given port.
//
// Therefore, we do not even use GPIO1-5 and instead use their high-speed
// partner for all pins. This is configured at startup in the runtime package
// (func initPins() in `runtime_mimxrt1062.go`).
// We cannot declare 32 pins for all available ports (GPIO1-9) anyway, since Pin
// is only uint8, and 9*32=288 > 256, so something has to be sacrificed.

const (
	portA Pin = iota * 32 // GPIO1(6)
	portB                 // GPIO2(7)
	portC                 // GPIO3(8)
	portD                 // GPIO4(9)
)

const (
	//                   [Pad]:      Alt Func 0       Alt Func 1       Alt Func 2           Alt Func 3            Alt Func 4            Alt Func 5   Alt Func 6            Alt Func 7            Alt Func 8             Alt Func 9
	//                   ----------  ---------------  ---------------  -------------------  --------------------  --------------------  -----------  --------------------  --------------------  ---------------------  ----------------
	PA0  = portA + 0  // [AD_B0_00]: FLEXPWM2_PWMA03  XBAR1_INOUT14    REF_CLK_32K          USB_OTG2_ID           LPI2C1_SCLS           GPIO1_IO00   USDHC1_RESET_B        LPSPI3_SCK             ~                      ~
	PA1  = portA + 1  // [AD_B0_01]: FLEXPWM2_PWMB03  XBAR1_INOUT15    REF_CLK_24M          USB_OTG1_ID           LPI2C1_SDAS           GPIO1_IO01   EWM_OUT_B             LPSPI3_SDO             ~                      ~
	PA2  = portA + 2  // [AD_B0_02]: FLEXCAN2_TX      XBAR1_INOUT16    LPUART6_TX           USB_OTG1_PWR          FLEXPWM1_PWMX00       GPIO1_IO02   LPI2C1_HREQ           LPSPI3_SDI             ~                      ~
	PA3  = portA + 3  // [AD_B0_03]: FLEXCAN2_RX      XBAR1_INOUT17    LPUART6_RX           USB_OTG1_OC           FLEXPWM1_PWMX01       GPIO1_IO03   REF_CLK_24M           LPSPI3_PCS0            ~                      ~
	PA4  = portA + 4  // [AD_B0_04]: SRC_BOOT_MODE00  MQS_RIGHT        ENET_TX_DATA03       SAI2_TX_SYNC          CSI_DATA09            GPIO1_IO04   PIT_TRIGGER00         LPSPI3_PCS1            ~                      ~
	PA5  = portA + 5  // [AD_B0_05]: SRC_BOOT_MODE01  MQS_LEFT         ENET_TX_DATA02       SAI2_TX_BCLK          CSI_DATA08            GPIO1_IO05   XBAR1_INOUT17         LPSPI3_PCS2            ~                      ~
	PA6  = portA + 6  // [AD_B0_06]: JTAG_TMS         GPT2_COMPARE1    ENET_RX_CLK          SAI2_RX_BCLK          CSI_DATA07            GPIO1_IO06   XBAR1_INOUT18         LPSPI3_PCS3            ~                      ~
	PA7  = portA + 7  // [AD_B0_07]: JTAG_TCK         GPT2_COMPARE2    ENET_TX_ER           SAI2_RX_SYNC          CSI_DATA06            GPIO1_IO07   XBAR1_INOUT19         ENET_1588_EVENT3_OUT   ~                      ~
	PA8  = portA + 8  // [AD_B0_08]: JTAG_MOD         GPT2_COMPARE3    ENET_RX_DATA03       SAI2_RX_DATA          CSI_DATA05            GPIO1_IO08   XBAR1_IN20            ENET_1588_EVENT3_IN    ~                      ~
	PA9  = portA + 9  // [AD_B0_09]: JTAG_TDI         FLEXPWM2_PWMA03  ENET_RX_DATA02       SAI2_TX_DATA          CSI_DATA04            GPIO1_IO09   XBAR1_IN21            GPT2_CLK              SEMC_DQS4               ~
	PA10 = portA + 10 // [AD_B0_10]: JTAG_TDO         FLEXPWM1_PWMA03  ENET_CRS             SAI2_MCLK             CSI_DATA03            GPIO1_IO10   XBAR1_IN22            ENET_1588_EVENT0_OUT  FLEXCAN3_TX            ARM_TRACE_SWO
	PA11 = portA + 11 // [AD_B0_11]: JTAG_TRSTB       FLEXPWM1_PWMB03  ENET_COL             WDOG1_WDOG_B          CSI_DATA02            GPIO1_IO11   XBAR1_IN23            ENET_1588_EVENT0_IN   FLEXCAN3_RX            SEMC_CLK6
	PA12 = portA + 12 // [AD_B0_12]: LPI2C4_SCL       CCM_PMIC_READY   LPUART1_TX           WDOG2_WDOG_B          FLEXPWM1_PWMX02       GPIO1_IO12   ENET_1588_EVENT1_OUT  NMI_GLUE_NMI           ~                      ~
	PA13 = portA + 13 // [AD_B0_13]: LPI2C4_SDA       GPT1_CLK         LPUART1_RX           EWM_OUT_B             FLEXPWM1_PWMX03       GPIO1_IO13   ENET_1588_EVENT1_IN   REF_CLK_24M            ~                      ~
	PA14 = portA + 14 // [AD_B0_14]: USB_OTG2_OC      XBAR1_IN24       LPUART1_CTS_B        ENET_1588_EVENT0_OUT  CSI_VSYNC             GPIO1_IO14   FLEXCAN2_TX           FLEXCAN3_TX            ~                      ~
	PA15 = portA + 15 // [AD_B0_15]: USB_OTG2_PWR     XBAR1_IN25       LPUART1_RTS_B        ENET_1588_EVENT0_IN   CSI_HSYNC             GPIO1_IO15   FLEXCAN2_RX           WDOG1_WDOG_RST_B_DEB  FLEXCAN3_RX             ~
	PA16 = portA + 16 // [AD_B1_00]: USB_OTG2_ID      QTIMER3_TIMER0   LPUART2_CTS_B        LPI2C1_SCL            WDOG1_B               GPIO1_IO16   USDHC1_WP             KPP_ROW07             ENET2_1588_EVENT0_OUT  FLEXIO3_FLEXIO00
	PA17 = portA + 17 // [AD_B1_01]: USB_OTG1_PWR     QTIMER3_TIMER1   LPUART2_RTS_B        LPI2C1_SDA            CCM_PMIC_READY        GPIO1_IO17   USDHC1_VSELECT        KPP_COL07             ENET2_1588_EVENT0_IN   FLEXIO3_FLEXIO01
	PA18 = portA + 18 // [AD_B1_02]: USB_OTG1_ID      QTIMER3_TIMER2   LPUART2_TX           SPDIF_OUT             ENET_1588_EVENT2_OUT  GPIO1_IO18   USDHC1_CD_B           KPP_ROW06             GPT2_CLK               FLEXIO3_FLEXIO02
	PA19 = portA + 19 // [AD_B1_03]: USB_OTG1_OC      QTIMER3_TIMER3   LPUART2_RX           SPDIF_IN              ENET_1588_EVENT2_IN   GPIO1_IO19   USDHC2_CD_B           KPP_COL06             GPT2_CAPTURE1          FLEXIO3_FLEXIO03
	PA20 = portA + 20 // [AD_B1_04]: FLEXSPIB_DATA03  ENET_MDC         LPUART3_CTS_B        SPDIF_SR_CLK          CSI_PIXCLK            GPIO1_IO20   USDHC2_DATA0          KPP_ROW05             GPT2_CAPTURE2          FLEXIO3_FLEXIO04
	PA21 = portA + 21 // [AD_B1_05]: FLEXSPIB_DATA02  ENET_MDIO        LPUART3_RTS_B        SPDIF_OUT             CSI_MCLK              GPIO1_IO21   USDHC2_DATA1          KPP_COL05             GPT2_COMPARE1          FLEXIO3_FLEXIO05
	PA22 = portA + 22 // [AD_B1_06]: FLEXSPIB_DATA01  LPI2C3_SDA       LPUART3_TX           SPDIF_LOCK            CSI_VSYNC             GPIO1_IO22   USDHC2_DATA2          KPP_ROW04             GPT2_COMPARE2          FLEXIO3_FLEXIO06
	PA23 = portA + 23 // [AD_B1_07]: FLEXSPIB_DATA00  LPI2C3_SCL       LPUART3_RX           SPDIF_EXT_CLK         CSI_HSYNC             GPIO1_IO23   USDHC2_DATA3          KPP_COL04             GPT2_COMPARE3          FLEXIO3_FLEXIO07
	PA24 = portA + 24 // [AD_B1_08]: FLEXSPIA_SS1_B   FLEXPWM4_PWMA00  FLEXCAN1_TX          CCM_PMIC_READY        CSI_DATA09            GPIO1_IO24   USDHC2_CMD            KPP_ROW03             FLEXIO3_FLEXIO08        ~
	PA25 = portA + 25 // [AD_B1_09]: FLEXSPIA_DQS     FLEXPWM4_PWMA01  FLEXCAN1_RX          SAI1_MCLK             CSI_DATA08            GPIO1_IO25   USDHC2_CLK            KPP_COL03             FLEXIO3_FLEXIO09        ~
	PA26 = portA + 26 // [AD_B1_10]: FLEXSPIA_DATA03  WDOG1_B          LPUART8_TX           SAI1_RX_SYNC          CSI_DATA07            GPIO1_IO26   USDHC2_WP             KPP_ROW02             ENET2_1588_EVENT1_OUT  FLEXIO3_FLEXIO10
	PA27 = portA + 27 // [AD_B1_11]: FLEXSPIA_DATA02  EWM_OUT_B        LPUART8_RX           SAI1_RX_BCLK          CSI_DATA06            GPIO1_IO27   USDHC2_RESET_B        KPP_COL02             ENET2_1588_EVENT1_IN   FLEXIO3_FLEXIO11
	PA28 = portA + 28 // [AD_B1_12]: FLEXSPIA_DATA01  ACMP_OUT00       LPSPI3_PCS0          SAI1_RX_DATA00        CSI_DATA05            GPIO1_IO28   USDHC2_DATA4          KPP_ROW01             ENET2_1588_EVENT2_OUT  FLEXIO3_FLEXIO12
	PA29 = portA + 29 // [AD_B1_13]: FLEXSPIA_DATA00  ACMP_OUT01       LPSPI3_SDI           SAI1_TX_DATA00        CSI_DATA04            GPIO1_IO29   USDHC2_DATA5          KPP_COL01             ENET2_1588_EVENT2_IN   FLEXIO3_FLEXIO13
	PA30 = portA + 30 // [AD_B1_14]: FLEXSPIA_SCLK    ACMP_OUT02       LPSPI3_SDO           SAI1_TX_BCLK          CSI_DATA03            GPIO1_IO30   USDHC2_DATA6          KPP_ROW00             ENET2_1588_EVENT3_OUT  FLEXIO3_FLEXIO14
	PA31 = portA + 31 // [AD_B1_15]: FLEXSPIA_SS0_B   ACMP_OUT03       LPSPI3_SCK           SAI1_TX_SYNC          CSI_DATA02            GPIO1_IO31   USDHC2_DATA7          KPP_COL00             ENET2_1588_EVENT3_IN   FLEXIO3_FLEXIO15

	PB0  = portB + 0  // [B0_00]:    LCD_CLK          QTIMER1_TIMER0   MQS_RIGHT            LPSPI4_PCS0           FLEXIO2_FLEXIO00      GPIO2_IO00   SEMC_CSX01            ENET2_MDC              ~                      ~
	PB1  = portB + 1  // [B0_01]:    LCD_ENABLE       QTIMER1_TIMER1   MQS_LEFT             LPSPI4_SDI            FLEXIO2_FLEXIO01      GPIO2_IO01   SEMC_CSX02            ENET2_MDIO             ~                      ~
	PB2  = portB + 2  // [B0_02]:    LCD_HSYNC        QTIMER1_TIMER2   FLEXCAN1_TX          LPSPI4_SDO            FLEXIO2_FLEXIO02      GPIO2_IO02   SEMC_CSX03            ENET2_1588_EVENT0_OUT  ~                      ~
	PB3  = portB + 3  // [B0_03]:    LCD_VSYNC        QTIMER2_TIMER0   FLEXCAN1_RX          LPSPI4_SCK            FLEXIO2_FLEXIO03      GPIO2_IO03   WDOG2_RESET_B_DEB     ENET2_1588_EVENT0_IN   ~                      ~
	PB4  = portB + 4  // [B0_04]:    LCD_DATA00       QTIMER2_TIMER1   LPI2C2_SCL           ARM_TRACE0            FLEXIO2_FLEXIO04      GPIO2_IO04   SRC_BOOT_CFG00        ENET2_TDATA03          ~                      ~
	PB5  = portB + 5  // [B0_05]:    LCD_DATA01       QTIMER2_TIMER2   LPI2C2_SDA           ARM_TRACE1            FLEXIO2_FLEXIO05      GPIO2_IO05   SRC_BOOT_CFG01        ENET2_TDATA02          ~                      ~
	PB6  = portB + 6  // [B0_06]:    LCD_DATA02       QTIMER3_TIMER0   FLEXPWM2_PWMA00      ARM_TRACE2            FLEXIO2_FLEXIO06      GPIO2_IO06   SRC_BOOT_CFG02        ENET2_RX_CLK           ~                      ~
	PB7  = portB + 7  // [B0_07]:    LCD_DATA03       QTIMER3_TIMER1   FLEXPWM2_PWMB00      ARM_TRACE3            FLEXIO2_FLEXIO07      GPIO2_IO07   SRC_BOOT_CFG03        ENET2_TX_ER            ~                      ~
	PB8  = portB + 8  // [B0_08]:    LCD_DATA04       QTIMER3_TIMER2   FLEXPWM2_PWMA01      LPUART3_TX            FLEXIO2_FLEXIO08      GPIO2_IO08   SRC_BOOT_CFG04        ENET2_RDATA03          ~                      ~
	PB9  = portB + 9  // [B0_09]:    LCD_DATA05       QTIMER4_TIMER0   FLEXPWM2_PWMB01      LPUART3_RX            FLEXIO2_FLEXIO09      GPIO2_IO09   SRC_BOOT_CFG05        ENET2_RDATA02          ~                      ~
	PB10 = portB + 10 // [B0_10]:    LCD_DATA06       QTIMER4_TIMER1   FLEXPWM2_PWMA02      SAI1_TX_DATA03        FLEXIO2_FLEXIO10      GPIO2_IO10   SRC_BOOT_CFG06        ENET2_CRS              ~                      ~
	PB11 = portB + 11 // [B0_11]:    LCD_DATA07       QTIMER4_TIMER2   FLEXPWM2_PWMB02      SAI1_TX_DATA02        FLEXIO2_FLEXIO11      GPIO2_IO11   SRC_BOOT_CFG07        ENET2_COL              ~                      ~
	PB12 = portB + 12 // [B0_12]:    LCD_DATA08       XBAR1_INOUT10    ARM_TRACE_CLK        SAI1_TX_DATA01        FLEXIO2_FLEXIO12      GPIO2_IO12   SRC_BOOT_CFG08        ENET2_TDATA00          ~                      ~
	PB13 = portB + 13 // [B0_13]:    LCD_DATA09       XBAR1_INOUT11    ARM_TRACE_SWO        SAI1_MCLK             FLEXIO2_FLEXIO13      GPIO2_IO13   SRC_BOOT_CFG09        ENET2_TDATA01          ~                      ~
	PB14 = portB + 14 // [B0_14]:    LCD_DATA10       XBAR1_INOUT12    ARM_TXEV             SAI1_RX_SYNC          FLEXIO2_FLEXIO14      GPIO2_IO14   SRC_BOOT_CFG10        ENET2_TX_EN            ~                      ~
	PB15 = portB + 15 // [B0_15]:    LCD_DATA11       XBAR1_INOUT13    ARM_RXEV             SAI1_RX_BCLK          FLEXIO2_FLEXIO15      GPIO2_IO15   SRC_BOOT_CFG11        ENET2_TX_CLK          ENET2_REF_CLK2          ~
	PB16 = portB + 16 // [B1_00]:    LCD_DATA12       XBAR1_INOUT14    LPUART4_TX           SAI1_RX_DATA00        FLEXIO2_FLEXIO16      GPIO2_IO16   FLEXPWM1_PWMA03       ENET2_RX_ER           FLEXIO3_FLEXIO16        ~
	PB17 = portB + 17 // [B1_01]:    LCD_DATA13       XBAR1_INOUT15    LPUART4_RX           SAI1_TX_DATA00        FLEXIO2_FLEXIO17      GPIO2_IO17   FLEXPWM1_PWMB03       ENET2_RDATA00         FLEXIO3_FLEXIO17        ~
	PB18 = portB + 18 // [B1_02]:    LCD_DATA14       XBAR1_INOUT16    LPSPI4_PCS2          SAI1_TX_BCLK          FLEXIO2_FLEXIO18      GPIO2_IO18   FLEXPWM2_PWMA03       ENET2_RDATA01         FLEXIO3_FLEXIO18        ~
	PB19 = portB + 19 // [B1_03]:    LCD_DATA15       XBAR1_INOUT17    LPSPI4_PCS1          SAI1_TX_SYNC          FLEXIO2_FLEXIO19      GPIO2_IO19   FLEXPWM2_PWMB03       ENET2_RX_EN           FLEXIO3_FLEXIO19        ~
	PB20 = portB + 20 // [B1_04]:    LCD_DATA16       LPSPI4_PCS0      CSI_DATA15           ENET_RX_DATA00        FLEXIO2_FLEXIO20      GPIO2_IO20   GPT1_CLK              FLEXIO3_FLEXIO20       ~                      ~
	PB21 = portB + 21 // [B1_05]:    LCD_DATA17       LPSPI4_SDI       CSI_DATA14           ENET_RX_DATA01        FLEXIO2_FLEXIO21      GPIO2_IO21   GPT1_CAPTURE1         FLEXIO3_FLEXIO21       ~                      ~
	PB22 = portB + 22 // [B1_06]:    LCD_DATA18       LPSPI4_SDO       CSI_DATA13           ENET_RX_EN            FLEXIO2_FLEXIO22      GPIO2_IO22   GPT1_CAPTURE2         FLEXIO3_FLEXIO22       ~                      ~
	PB23 = portB + 23 // [B1_07]:    LCD_DATA19       LPSPI4_SCK       CSI_DATA12           ENET_TX_DATA00        FLEXIO2_FLEXIO23      GPIO2_IO23   GPT1_COMPARE1         FLEXIO3_FLEXIO23       ~                      ~
	PB24 = portB + 24 // [B1_08]:    LCD_DATA20       QTIMER1_TIMER3   CSI_DATA11           ENET_TX_DATA01        FLEXIO2_FLEXIO24      GPIO2_IO24   FLEXCAN2_TX           GPT1_COMPARE2         FLEXIO3_FLEXIO24        ~
	PB25 = portB + 25 // [B1_09]:    LCD_DATA21       QTIMER2_TIMER3   CSI_DATA10           ENET_TX_EN            FLEXIO2_FLEXIO25      GPIO2_IO25   FLEXCAN2_RX           GPT1_COMPARE3         FLEXIO3_FLEXIO25        ~
	PB26 = portB + 26 // [B1_10]:    LCD_DATA22       QTIMER3_TIMER3   CSI_DATA00           ENET_TX_CLK           FLEXIO2_FLEXIO26      GPIO2_IO26   ENET_REF_CLK          FLEXIO3_FLEXIO26       ~                      ~
	PB27 = portB + 27 // [B1_11]:    LCD_DATA23       QTIMER4_TIMER3   CSI_DATA01           ENET_RX_ER            FLEXIO2_FLEXIO27      GPIO2_IO27   LPSPI4_PCS3           FLEXIO3_FLEXIO27       ~                      ~
	PB28 = portB + 28 // [B1_12]:    LPUART5_TX       CSI_PIXCLK       ENET_1588_EVENT0_IN  FLEXIO2_FLEXIO28      GPIO2_IO28            USDHC1_CD_B  FLEXIO3_FLEXIO28       ~                     ~                      ~
	PB29 = portB + 29 // [B1_13]:    WDOG1_B          LPUART5_RX       CSI_VSYNC            ENET_1588_EVENT0_OUT  FLEXIO2_FLEXIO29      GPIO2_IO29   USDHC1_WP             SEMC_DQS4             FLEXIO3_FLEXIO29        ~
	PB30 = portB + 30 // [B1_14]:    ENET_MDC         FLEXPWM4_PWMA02  CSI_HSYNC            XBAR1_IN02            FLEXIO2_FLEXIO30      GPIO2_IO30   USDHC1_VSELECT        ENET2_TDATA00         FLEXIO3_FLEXIO30        ~
	PB31 = portB + 31 // [B1_15]:    ENET_MDIO        FLEXPWM4_PWMA03  CSI_MCLK             XBAR1_IN03            FLEXIO2_FLEXIO31      GPIO2_IO31   USDHC1_RESET_B        ENET2_TDATA01         FLEXIO3_FLEXIO31        ~

	PC0  = portC + 0  // [SD_B1_00]: USDHC2_DATA3     FLEXSPIB_DATA03  FLEXPWM1_PWMA03      SAI1_TX_DATA03        LPUART4_TX            GPIO3_IO00   SAI3_RX_DATA           ~                     ~                      ~
	PC1  = portC + 1  // [SD_B1_01]: USDHC2_DATA2     FLEXSPIB_DATA02  FLEXPWM1_PWMB03      SAI1_TX_DATA02        LPUART4_RX            GPIO3_IO01   SAI3_TX_DATA           ~                     ~                      ~
	PC2  = portC + 2  // [SD_B1_02]: USDHC2_DATA1     FLEXSPIB_DATA01  FLEXPWM2_PWMA03      SAI1_TX_DATA01        FLEXCAN1_TX           GPIO3_IO02   CCM_WAIT              SAI3_TX_SYNC           ~                      ~
	PC3  = portC + 3  // [SD_B1_03]: USDHC2_DATA0     FLEXSPIB_DATA00  FLEXPWM2_PWMB03      SAI1_MCLK             FLEXCAN1_RX           GPIO3_IO03   CCM_PMIC_READY        SAI3_TX_BCLK           ~                      ~
	PC4  = portC + 4  // [SD_B1_04]: USDHC2_CLK       FLEXSPIB_SCLK    LPI2C1_SCL           SAI1_RX_SYNC          FLEXSPIA_SS1_B        GPIO3_IO04   CCM_STOP              SAI3_MCLK              ~                      ~
	PC5  = portC + 5  // [SD_B1_05]: USDHC2_CMD       FLEXSPIA_DQS     LPI2C1_SDA           SAI1_RX_BCLK          FLEXSPIB_SS0_B        GPIO3_IO05   SAI3_RX_SYNC           ~                     ~                      ~
	PC6  = portC + 6  // [SD_B1_06]: USDHC2_RESET_B   FLEXSPIA_SS0_B   LPUART7_CTS_B        SAI1_RX_DATA00        LPSPI2_PCS0           GPIO3_IO06   SAI3_RX_BCLK           ~                     ~                      ~
	PC7  = portC + 7  // [SD_B1_07]: SEMC_CSX01       FLEXSPIA_SCLK    LPUART7_RTS_B        SAI1_TX_DATA00        LPSPI2_SCK            GPIO3_IO07    ~                     ~                     ~                      ~
	PC8  = portC + 8  // [SD_B1_08]: USDHC2_DATA4     FLEXSPIA_DATA00  LPUART7_TX           SAI1_TX_BCLK          LPSPI2_SD0            GPIO3_IO08   SEMC_CSX02             ~                     ~                      ~
	PC9  = portC + 9  // [SD_B1_09]: USDHC2_DATA5     FLEXSPIA_DATA01  LPUART7_RX           SAI1_TX_SYNC          LPSPI2_SDI            GPIO3_IO09    ~                     ~                     ~                      ~
	PC10 = portC + 10 // [SD_B1_10]: USDHC2_DATA6     FLEXSPIA_DATA02  LPUART2_RX           LPI2C2_SDA            LPSPI2_PCS2           GPIO3_IO10    ~                     ~                     ~                      ~
	PC11 = portC + 11 // [SD_B1_11]: USDHC2_DATA7     FLEXSPIA_DATA03  LPUART2_TX           LPI2C2_SCL            LPSPI2_PCS3           GPIO3_IO11    ~                     ~                     ~                      ~
	PC12 = portC + 12 // [SD_B0_00]: USDHC1_CMD       FLEXPWM1_PWMA00  LPI2C3_SCL           XBAR1_INOUT04         LPSPI1_SCK            GPIO3_IO12   FLEXSPIA_SS1_B        ENET2_TX_EN           SEMC_DQS4               ~
	PC13 = portC + 13 // [SD_B0_01]: USDHC1_CLK       FLEXPWM1_PWMB00  LPI2C3_SDA           XBAR1_INOUT05         LPSPI1_PCS0           GPIO3_IO13   FLEXSPIB_SS1_B        ENET2_TX_CLK          ENET2_REF_CLK2          ~
	PC14 = portC + 14 // [SD_B0_02]: USDHC1_DATA0     FLEXPWM1_PWMA01  LPUART8_CTS_B        XBAR1_INOUT06         LPSPI1_SDO            GPIO3_IO14   ENET2_RX_ER           SEMC_CLK5              ~                      ~
	PC15 = portC + 15 // [SD_B0_03]: USDHC1_DATA1     FLEXPWM1_PWMB01  LPUART8_RTS_B        XBAR1_INOUT07         LPSPI1_SDI            GPIO3_IO15   ENET2_RDATA00         SEMC_CLK6              ~                      ~
	PC16 = portC + 16 // [SD_B0_04]: USDHC1_DATA2     FLEXPWM1_PWMA02  LPUART8_TX           XBAR1_INOUT08         FLEXSPIB_SS0_B        GPIO3_IO16   CCM_CLKO1             ENET2_RDATA01          ~                      ~
	PC17 = portC + 17 // [SD_B0_05]: USDHC1_DATA3     FLEXPWM1_PWMB02  LPUART8_RX           XBAR1_INOUT09         FLEXSPIB_DQS          GPIO3_IO17   CCM_CLKO2             ENET2_RX_EN            ~                      ~
	PC18 = portC + 18 // [EMC_32]:   SEMC_DATA10      FLEXPWM3_PWMB01  LPUART7_RX           CCM_PMIC_RDY          CSI_DATA21            GPIO3_IO18   ENET2_TX_EN            ~                     ~                      ~
	PC19 = portC + 19 // [EMC_33]:   SEMC_DATA11      FLEXPWM3_PWMA02  USDHC1_RESET_B       SAI3_RX_DATA          CSI_DATA20            GPIO3_IO19   ENET2_TX_CLK          ENET2_REF_CLK2         ~                      ~
	PC20 = portC + 20 // [EMC_34]:   SEMC_DATA12      FLEXPWM3_PWMB02  USDHC1_VSELECT       SAI3_RX_SYNC          CSI_DATA19            GPIO3_IO20   ENET2_RX_ER            ~                     ~                      ~
	PC21 = portC + 21 // [EMC_35]:   SEMC_DATA13      XBAR1_INOUT18    GPT1_COMPARE1        SAI3_RX_BCLK          CSI_DATA18            GPIO3_IO21   USDHC1_CD_B           ENET2_RDATA00          ~                      ~
	PC22 = portC + 22 // [EMC_36]:   SEMC_DATA14      XBAR1_IN22       GPT1_COMPARE2        SAI3_TX_DATA          CSI_DATA17            GPIO3_IO22   USDHC1_WP             ENET2_RDATA01         FLEXCAN3_TX             ~
	PC23 = portC + 23 // [EMC_37]:   SEMC_DATA15      XBAR1_IN23       GPT1_COMPARE3        SAI3_MCLK             CSI_DATA16            GPIO3_IO23   USDHC2_WP             ENET2_RX_EN           FLEXCAN3_RX             ~
	PC24 = portC + 24 // [EMC_38]:   SEMC_DM01        FLEXPWM1_PWMA03  LPUART8_TX           SAI3_TX_BCLK          CSI_FIELD             GPIO3_IO24   USDHC2_VSELECT        ENET2_MDC              ~                      ~
	PC25 = portC + 25 // [EMC_39]:   SEMC_DQS         FLEXPWM1_PWMB03  LPUART8_RX           SAI3_TX_SYNC          WDOG1_WDOG_B          GPIO3_IO25   USDHC2_CD_B           ENET2_MDIO            SEMC_DQS4               ~
	PC26 = portC + 26 // [EMC_40]:   SEMC_RDY         GPT2_CAPTURE2    LPSPI1_PCS2          USB_OTG2_OC           ENET_MDC              GPIO3_IO26   USDHC2_RESET_B        SEMC_CLK5              ~                      ~
	PC27 = portC + 27 // [EMC_41]:   SEMC_CSX00       GPT2_CAPTURE1    LPSPI1_PCS3          USB_OTG2_PWR          ENET_MDIO             GPIO3_IO27   USDHC1_VSELECT         ~                     ~                      ~
	_    = portC + 28 //
	_    = portC + 29 //
	_    = portC + 30 //
	_    = portC + 31 //

	PD0  = portD + 0  // [EMC_00]:   SEMC_DATA00      FLEXPWM4_PWMA00  LPSPI2_SCK           XBAR1_XBAR_IN02       FLEXIO1_FLEXIO00      GPIO4_IO00    ~                     ~                     ~                      ~
	PD1  = portD + 1  // [EMC_01]:   SEMC_DATA01      FLEXPWM4_PWMB00  LPSPI2_PCS0          XBAR1_IN03            FLEXIO1_FLEXIO01      GPIO4_IO01    ~                     ~                     ~                      ~
	PD2  = portD + 2  // [EMC_02]:   SEMC_DATA02      FLEXPWM4_PWMA01  LPSPI2_SDO           XBAR1_INOUT04         FLEXIO1_FLEXIO02      GPIO4_IO02    ~                     ~                     ~                      ~
	PD3  = portD + 3  // [EMC_03]:   SEMC_DATA03      FLEXPWM4_PWMB01  LPSPI2_SDI           XBAR1_INOUT05         FLEXIO1_FLEXIO03      GPIO4_IO03    ~                     ~                     ~                      ~
	PD4  = portD + 4  // [EMC_04]:   SEMC_DATA04      FLEXPWM4_PWMA02  SAI2_TX_DATA         XBAR1_INOUT06         FLEXIO1_FLEXIO04      GPIO4_IO04    ~                     ~                     ~                      ~
	PD5  = portD + 5  // [EMC_05]:   SEMC_DATA05      FLEXPWM4_PWMB02  SAI2_TX_SYNC         XBAR1_INOUT07         FLEXIO1_FLEXIO05      GPIO4_IO05    ~                     ~                     ~                      ~
	PD6  = portD + 6  // [EMC_06]:   SEMC_DATA06      FLEXPWM2_PWMA00  SAI2_TX_BCLK         XBAR1_INOUT08         FLEXIO1_FLEXIO06      GPIO4_IO06    ~                     ~                     ~                      ~
	PD7  = portD + 7  // [EMC_07]:   SEMC_DATA07      FLEXPWM2_PWMB00  SAI2_MCLK            XBAR1_INOUT09         FLEXIO1_FLEXIO07      GPIO4_IO07    ~                     ~                     ~                      ~
	PD8  = portD + 8  // [EMC_08]:   SEMC_DM00        FLEXPWM2_PWMA01  SAI2_RX_DATA         XBAR1_INOUT17         FLEXIO1_FLEXIO08      GPIO4_IO08    ~                     ~                     ~                      ~
	PD9  = portD + 9  // [EMC_09]:   SEMC_ADDR00      FLEXPWM2_PWMB01  SAI2_RX_SYNC         FLEXCAN2_TX           FLEXIO1_FLEXIO09      GPIO4_IO09   FLEXSPI2_B_SS1_B       ~                     ~                      ~
	PD10 = portD + 10 // [EMC_10]:   SEMC_ADDR01      FLEXPWM2_PWMA02  SAI2_RX_BCLK         FLEXCAN2_RX           FLEXIO1_FLEXIO10      GPIO4_IO10   FLEXSPI2_B_SS0_B       ~                     ~                      ~
	PD11 = portD + 11 // [EMC_11]:   SEMC_ADDR02      FLEXPWM2_PWMB02  LPI2C4_SDA           USDHC2_RESET_B        FLEXIO1_FLEXIO11      GPIO4_IO11   FLEXSPI2_B_DQS         ~                     ~                      ~
	PD12 = portD + 12 // [EMC_12]:   SEMC_ADDR03      XBAR1_IN24       LPI2C4_SCL           USDHC1_WP             FLEXPWM1_PWMA03       GPIO4_IO12   FLEXSPI2_B_SCLK        ~                     ~                      ~
	PD13 = portD + 13 // [EMC_13]:   SEMC_ADDR04      XBAR1_IN25       LPUART3_TX           MQS_RIGHT             FLEXPWM1_PWMB03       GPIO4_IO13   FLEXSPI2_B_DATA00      ~                     ~                      ~
	PD14 = portD + 14 // [EMC_14]:   SEMC_ADDR05      XBAR1_INOUT19    LPUART3_RX           MQS_LEFT              LPSPI2_PCS1           GPIO4_IO14   FLEXSPI2_B_DATA01      ~                     ~                      ~
	PD15 = portD + 15 // [EMC_15]:   SEMC_ADDR06      XBAR1_IN20       LPUART3_CTS_B        SPDIF_OUT             QTIMER3_TIMER0        GPIO4_IO15   FLEXSPI2_B_DATA02      ~                     ~                      ~
	PD16 = portD + 16 // [EMC_16]:   SEMC_ADDR07      XBAR1_IN21       LPUART3_RTS_B        SPDIF_IN              QTIMER3_TIMER1        GPIO4_IO16   FLEXSPI2_B_DATA03      ~                     ~                      ~
	PD17 = portD + 17 // [EMC_17]:   SEMC_ADDR08      FLEXPWM4_PWMA03  LPUART4_CTS_B        FLEXCAN1_TX           QTIMER3_TIMER2        GPIO4_IO17    ~                     ~                     ~                      ~
	PD18 = portD + 18 // [EMC_18]:   SEMC_ADDR09      FLEXPWM4_PWMB03  LPUART4_RTS_B        FLEXCAN1_RX           QTIMER3_TIMER3        GPIO4_IO18   SNVS_VIO_5_CTL         ~                     ~                      ~
	PD19 = portD + 19 // [EMC_19]:   SEMC_ADDR11      FLEXPWM2_PWMA03  LPUART4_TX           ENET_RDATA01          QTIMER2_TIMER0        GPIO4_IO19   SNVS_VIO_5             ~                     ~                      ~
	PD20 = portD + 20 // [EMC_20]:   SEMC_ADDR12      FLEXPWM2_PWMB03  LPUART4_RX           ENET_RDATA00          QTIMER2_TIMER1        GPIO4_IO20    ~                     ~                     ~                      ~
	PD21 = portD + 21 // [EMC_21]:   SEMC_BA0         FLEXPWM3_PWMA03  LPI2C3_SDA           ENET_TDATA01          QTIMER2_TIMER2        GPIO4_IO21    ~                     ~                     ~                      ~
	PD22 = portD + 22 // [EMC_22]:   SEMC_BA1         FLEXPWM3_PWMB03  LPI2C3_SCL           ENET_TDATA00          QTIMER2_TIMER3        GPIO4_IO22   FLEXSPI2_A_SS1_B       ~                     ~                      ~
	PD23 = portD + 23 // [EMC_23]:   SEMC_ADDR10      FLEXPWM1_PWMA00  LPUART5_TX           ENET_RX_EN            GPT1_CAPTURE2         GPIO4_IO23   FLEXSPI2_A_DQS         ~                     ~                      ~
	PD24 = portD + 24 // [EMC_24]:   SEMC_CAS         FLEXPWM1_PWMB00  LPUART5_RX           ENET_TX_EN            GPT1_CAPTURE1         GPIO4_IO24   FLEXSPI2_A_SS0_B       ~                     ~                      ~
	PD25 = portD + 25 // [EMC_25]:   SEMC_RAS         FLEXPWM1_PWMA01  LPUART6_TX           ENET_TX_CLK           ENET_REF_CLK          GPIO4_IO25   FLEXSPI2_A_SCLK        ~                     ~                      ~
	PD26 = portD + 26 // [EMC_26]:   SEMC_CLK         FLEXPWM1_PWMB01  LPUART6_RX           ENET_RX_ER            FLEXIO1_FLEXIO12      GPIO4_IO26   FLEXSPI2_A_DATA00      ~                     ~                      ~
	PD27 = portD + 27 // [EMC_27]:   SEMC_CKE         FLEXPWM1_PWMA02  LPUART5_RTS_B        LPSPI1_SCK            FLEXIO1_FLEXIO13      GPIO4_IO27   FLEXSPI2_A_DATA01      ~                     ~                      ~
	PD28 = portD + 28 // [EMC_28]:   SEMC_WE          FLEXPWM1_PWMB02  LPUART5_CTS_B        LPSPI1_SDO            FLEXIO1_FLEXIO14      GPIO4_IO28   FLEXSPI2_A_DATA02      ~                     ~                      ~
	PD29 = portD + 29 // [EMC_29]:   SEMC_CS0         FLEXPWM3_PWMA00  LPUART6_RTS_B        LPSPI1_SDI            FLEXIO1_FLEXIO15      GPIO4_IO29   FLEXSPI2_A_DATA03      ~                     ~                      ~
	PD30 = portD + 30 // [EMC_30]:   SEMC_DATA08      FLEXPWM3_PWMB00  LPUART6_CTS_B        LPSPI1_PCS0           CSI_DATA23            GPIO4_IO30   ENET2_TDATA00          ~                     ~                      ~
	PD31 = portD + 31 // [EMC_31]:   SEMC_DATA09      FLEXPWM3_PWMA01  LPUART7_TX           LPSPI1_PCS1           CSI_DATA22            GPIO4_IO31   ENET2_TDATA01          ~                     ~                      ~
)

func (p Pin) getPos() uint8   { return uint8(p % 32) }
func (p Pin) getMask() uint32 { return uint32(1) << p.getPos() }
func (p Pin) getPort() Pin    { return Pin(p/32) * 32 }

// Configure sets the GPIO pad and pin properties, and selects the appropriate
// alternate function, for a given Pin and PinConfig.
func (p Pin) Configure(config PinConfig) {
	var (
		sre = uint32(0x01 << 0)
		dse = func(n uint32) uint32 { return (n & 0x07) << 3 }
		spd = func(n uint32) uint32 { return (n & 0x03) << 6 }
		ode = uint32(0x01 << 11)
		pke = uint32(0x01 << 12)
		pue = uint32(0x01 << 13)
		pup = func(n uint32) uint32 { return (n & 0x03) << 14 }
		hys = uint32(0x01 << 16)
	)

	_, gpio := p.getGPIO() // use fast GPIO for all pins
	pad, mux := p.getPad()

	// first configure the pad characteristics
	switch config.Mode {
	case PinInput:
		gpio.GDIR.ClearBits(p.getMask())
		pad.Set(dse(7))

	case PinInputPullUp:
		gpio.GDIR.ClearBits(p.getMask())
		pad.Set(dse(7) | pke | pue | pup(3) | hys)

	case PinInputPullDown:
		gpio.GDIR.ClearBits(p.getMask())
		pad.Set(dse(7) | pke | pue | hys)

	case PinOutput:
		gpio.GDIR.SetBits(p.getMask())
		pad.Set(dse(7))

	case PinOutputOpenDrain:
		gpio.GDIR.SetBits(p.getMask())
		pad.Set(dse(7) | ode)

	case PinDisable:
		gpio.GDIR.ClearBits(p.getMask())
		pad.Set(dse(7) | hys)

	case PinInputAnalog:
		gpio.GDIR.ClearBits(p.getMask())
		pad.Set(dse(7))

	case PinModeUARTTX:
		pad.Set(sre | dse(3) | spd(3))

	case PinModeUARTRX:
		pad.Set(dse(7) | pke | pue | pup(3) | hys)

	case PinModeSPISDI:
		pad.Set(dse(7) | spd(2))

	case PinModeSPISDO:
		pad.Set(dse(7) | spd(2))

	case PinModeSPICLK:
		pad.Set(dse(7) | spd(2))

	case PinModeSPICS:
		pad.Set(dse(7))

	case PinModeI2CSDA, PinModeI2CSCL:
		pad.Set(ode | sre | dse(4) | spd(1) | pke | pue | pup(3))
	}

	// then configure the alternate function mux
	mux.Set(p.getMuxMode(config))
}

// Get returns the current value of a GPIO pin.
func (p Pin) Get() bool {
	_, gpio := p.getGPIO() // use fast GPIO for all pins
	return gpio.PSR.HasBits(p.getMask())
}

// Set changes the value of the GPIO pin. The pin must be configured as output.
func (p Pin) Set(value bool) {
	_, gpio := p.getGPIO() // use fast GPIO for all pins
	if value {
		gpio.DR_SET.Set(p.getMask())
	} else {
		gpio.DR_CLEAR.Set(p.getMask())
	}
}

// Toggle switches an output pin from low to high or from high to low.
func (p Pin) Toggle() {
	_, gpio := p.getGPIO() // use fast GPIO for all pins
	gpio.DR_TOGGLE.Set(p.getMask())
}

// dispatchInterrupt invokes the user-provided callback functions for external
// interrupts generated on the high-speed GPIO pins.
//
// Unfortunately, all four high-speed GPIO ports (A-D) are connected to just a
// single interrupt control line. Therefore, the interrupt status register (ISR)
// must be checked in all four GPIO ports on every interrupt.
func (jt *pinJumpTable) dispatchInterrupt(interrupt.Interrupt) {
	handle := func(gpio *nxp.GPIO_Type, port Pin) {
		if status := gpio.ISR.Get() & gpio.IMR.Get(); status != 0 {
			gpio.ISR.Set(status) // clear interrupt
			for status != 0 {
				off := Pin(bits.TrailingZeros32(status)) // ctz
				pin := Pin(port + off)
				jt.lut[pin](pin)
				status &^= 1 << off
			}
		}
	}
	if jt.numDefined > 0 {
		handle(nxp.GPIO6, portA)
		handle(nxp.GPIO7, portB)
		handle(nxp.GPIO8, portC)
		handle(nxp.GPIO9, portD)
	}
}

// set associates a function with a given Pin in the receiver lookup table. If
// the function is nil, the given Pin's associated function is removed.
func (jt *pinJumpTable) set(pin Pin, fn func(Pin)) {
	if int(pin) < len(jt.lut) {
		if nil != fn {
			if nil == jt.lut[pin] {
				jt.numDefined++
			}
			jt.lut[pin] = fn
		} else {
			if nil != jt.lut[pin] {
				jt.numDefined--
			}
			jt.lut[pin] = nil
		}
	}
}

// SetInterrupt sets an interrupt to be executed when a particular pin changes
// state. The pin should already be configured as an input, including a pull up
// or down if no external pull is provided.
//
// This call will replace a previously set callback on this pin. You can pass a
// nil func to unset the pin change interrupt. If you do so, the change
// parameter is ignored and can be set to any value (such as 0).
func (p Pin) SetInterrupt(change PinChange, callback func(Pin)) error {
	_, gpio := p.getGPIO() // use fast GPIO for all pins
	mask := p.getMask()
	if nil != callback {
		switch change {
		case PinRising, PinFalling:
			gpio.EDGE_SEL.ClearBits(mask)
			var reg *volatile.Register32
			var pos uint8
			if pos = p.getPos(); pos < 16 {
				reg = &gpio.ICR1 // ICR1 = pins 0-15
			} else {
				reg = &gpio.ICR2 // ICR2 = pins 16-31
				pos -= 16
			}
			reg.ReplaceBits(uint32(change), 0x3, pos*2)
		case PinToggle:
			gpio.EDGE_SEL.SetBits(mask)
		}
		pinISR.set(p, callback) // associate the callback with the pin
		gpio.ISR.Set(mask)      // clear any pending interrupt (W1C)
		gpio.IMR.SetBits(mask)  // enable external interrupt
	} else {
		pinISR.set(p, nil)       // remove any associated callback from the pin
		gpio.ISR.Set(mask)       // clear any pending interrupt (W1C)
		gpio.IMR.ClearBits(mask) // disable external interrupt
	}
	// enable or disable the interrupt based on number of defined callbacks
	if pinISR.numDefined > 0 {
		if nil == pinInterrupt {
			// create the Interrupt if it is not yet defined
			irq := interrupt.New(nxp.IRQ_GPIO6_7_8_9, pinISR.dispatchInterrupt)
			pinInterrupt = &irq
			pinInterrupt.Enable()
		}
	} else {
		if nil != pinInterrupt {
			// disable the interrupt if it is defined
			pinInterrupt.Disable()
		}
	}
	return nil
}

// getGPIO returns both the normal (IPG_CLK_ROOT) and high-speed (AHB_CLK_ROOT)
// GPIO peripherals to which a given Pin is connected.
//
// Note that, currently, the device is configured to use high-speed GPIO for all
// pins (GPIO6-9), so the first return value should not be used (GPIO1-4).
// See the remarks and documentation reference in the comments preceding the
// const Pin definitions above.
func (p Pin) getGPIO() (norm *nxp.GPIO_Type, fast *nxp.GPIO_Type) {
	switch p.getPort() {
	case portA:
		return nxp.GPIO1, nxp.GPIO6
	case portB:
		return nxp.GPIO2, nxp.GPIO7
	case portC:
		return nxp.GPIO3, nxp.GPIO8
	case portD:
		return nxp.GPIO4, nxp.GPIO9
	default:
		panic("machine: unknown port")
	}
}

// getPad returns both the pad and mux configration registers for a given Pin.
func (p Pin) getPad() (pad *volatile.Register32, mux *volatile.Register32) {
	switch p.getPort() {
	case portA:
		switch p.getPos() {
		case 0:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_AD_B0_00, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_AD_B0_00
		case 1:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_AD_B0_01, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_AD_B0_01
		case 2:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_AD_B0_02, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_AD_B0_02
		case 3:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_AD_B0_03, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_AD_B0_03
		case 4:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_AD_B0_04, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_AD_B0_04
		case 5:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_AD_B0_05, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_AD_B0_05
		case 6:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_AD_B0_06, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_AD_B0_06
		case 7:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_AD_B0_07, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_AD_B0_07
		case 8:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_AD_B0_08, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_AD_B0_08
		case 9:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_AD_B0_09, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_AD_B0_09
		case 10:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_AD_B0_10, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_AD_B0_10
		case 11:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_AD_B0_11, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_AD_B0_11
		case 12:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_AD_B0_12, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_AD_B0_12
		case 13:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_AD_B0_13, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_AD_B0_13
		case 14:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_AD_B0_14, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_AD_B0_14
		case 15:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_AD_B0_15, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_AD_B0_15
		case 16:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_AD_B1_00, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_AD_B1_00
		case 17:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_AD_B1_01, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_AD_B1_01
		case 18:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_AD_B1_02, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_AD_B1_02
		case 19:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_AD_B1_03, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_AD_B1_03
		case 20:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_AD_B1_04, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_AD_B1_04
		case 21:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_AD_B1_05, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_AD_B1_05
		case 22:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_AD_B1_06, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_AD_B1_06
		case 23:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_AD_B1_07, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_AD_B1_07
		case 24:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_AD_B1_08, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_AD_B1_08
		case 25:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_AD_B1_09, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_AD_B1_09
		case 26:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_AD_B1_10, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_AD_B1_10
		case 27:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_AD_B1_11, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_AD_B1_11
		case 28:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_AD_B1_12, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_AD_B1_12
		case 29:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_AD_B1_13, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_AD_B1_13
		case 30:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_AD_B1_14, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_AD_B1_14
		case 31:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_AD_B1_15, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_AD_B1_15
		}
	case portB:
		switch p.getPos() {
		case 0:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_B0_00, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_B0_00
		case 1:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_B0_01, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_B0_01
		case 2:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_B0_02, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_B0_02
		case 3:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_B0_03, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_B0_03
		case 4:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_B0_04, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_B0_04
		case 5:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_B0_05, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_B0_05
		case 6:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_B0_06, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_B0_06
		case 7:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_B0_07, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_B0_07
		case 8:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_B0_08, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_B0_08
		case 9:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_B0_09, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_B0_09
		case 10:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_B0_10, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_B0_10
		case 11:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_B0_11, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_B0_11
		case 12:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_B0_12, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_B0_12
		case 13:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_B0_13, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_B0_13
		case 14:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_B0_14, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_B0_14
		case 15:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_B0_15, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_B0_15
		case 16:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_B1_00, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_B1_00
		case 17:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_B1_01, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_B1_01
		case 18:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_B1_02, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_B1_02
		case 19:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_B1_03, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_B1_03
		case 20:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_B1_04, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_B1_04
		case 21:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_B1_05, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_B1_05
		case 22:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_B1_06, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_B1_06
		case 23:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_B1_07, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_B1_07
		case 24:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_B1_08, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_B1_08
		case 25:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_B1_09, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_B1_09
		case 26:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_B1_10, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_B1_10
		case 27:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_B1_11, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_B1_11
		case 28:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_B1_12, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_B1_12
		case 29:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_B1_13, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_B1_13
		case 30:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_B1_14, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_B1_14
		case 31:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_B1_15, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_B1_15
		}
	case portC:
		switch p.getPos() {
		case 0:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_SD_B1_00, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_SD_B1_00
		case 1:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_SD_B1_01, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_SD_B1_01
		case 2:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_SD_B1_02, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_SD_B1_02
		case 3:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_SD_B1_03, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_SD_B1_03
		case 4:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_SD_B1_04, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_SD_B1_04
		case 5:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_SD_B1_05, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_SD_B1_05
		case 6:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_SD_B1_06, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_SD_B1_06
		case 7:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_SD_B1_07, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_SD_B1_07
		case 8:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_SD_B1_08, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_SD_B1_08
		case 9:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_SD_B1_09, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_SD_B1_09
		case 10:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_SD_B1_10, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_SD_B1_10
		case 11:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_SD_B1_11, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_SD_B1_11
		case 12:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_SD_B0_00, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_SD_B0_00
		case 13:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_SD_B0_01, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_SD_B0_01
		case 14:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_SD_B0_02, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_SD_B0_02
		case 15:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_SD_B0_03, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_SD_B0_03
		case 16:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_SD_B0_04, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_SD_B0_04
		case 17:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_SD_B0_05, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_SD_B0_05
		case 18:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_32, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_32
		case 19:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_33, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_33
		case 20:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_34, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_34
		case 21:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_35, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_35
		case 22:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_36, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_36
		case 23:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_37, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_37
		case 24:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_38, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_38
		case 25:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_39, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_39
		case 26:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_40, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_40
		case 27:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_41, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_41
		case 28, 29, 30, 31:
		}
	case portD:
		switch p.getPos() {
		case 0:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_00, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_00
		case 1:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_01, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_01
		case 2:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_02, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_02
		case 3:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_03, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_03
		case 4:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_04, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_04
		case 5:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_05, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_05
		case 6:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_06, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_06
		case 7:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_07, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_07
		case 8:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_08, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_08
		case 9:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_09, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_09
		case 10:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_10, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_10
		case 11:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_11, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_11
		case 12:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_12, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_12
		case 13:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_13, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_13
		case 14:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_14, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_14
		case 15:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_15, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_15
		case 16:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_16, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_16
		case 17:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_17, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_17
		case 18:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_18, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_18
		case 19:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_19, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_19
		case 20:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_20, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_20
		case 21:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_21, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_21
		case 22:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_22, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_22
		case 23:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_23, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_23
		case 24:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_24, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_24
		case 25:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_25, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_25
		case 26:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_26, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_26
		case 27:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_27, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_27
		case 28:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_28, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_28
		case 29:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_29, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_29
		case 30:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_30, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_30
		case 31:
			return &nxp.IOMUXC.SW_PAD_CTL_PAD_GPIO_EMC_31, &nxp.IOMUXC.SW_MUX_CTL_PAD_GPIO_EMC_31
		}
	}
	panic("machine: invalid pin")
}

// muxSelect is yet another level of indirection required to connect pins in an
// alternate function state to a desired peripheral (since more than one pin can
// provide a given alternate function).
//
// Once a pin is configured with a given alternate function mode, the IOMUXC
// device must then be configured to select which alternate function pin to
// route to the desired peripheral.
//
// The reference manual refers to this functionality as a "Daisy Chain". The
// associated docs are found in the i.MX RT1060 Processor Reference Manual:
// "Chapter 11.3.3 Daisy chain - multi pads driving same module input pin"
type muxSelect struct {
	mux uint8                // AF mux selection (NOT a Pin type)
	sel *volatile.Register32 // AF selection register
}

// connect configures the IOMUXC controller to route a given pin with alternate
// function to a desired peripheral (see godoc comments on type muxSelect).
func (s muxSelect) connect() {
	s.sel.Set(uint32(s.mux))
}

// getMuxMode acts as a callback from the `(Pin).Configure(PinMode)` routine to
// determine the alternate function setting for a given Pin and PinConfig.
// This value is used in the IOMUXC device's SW_MUX_CTL_PAD_GPIO_* registers.
func (p Pin) getMuxMode(config PinConfig) uint32 {
	const forcePath = true // TODO: should be input parameter?
	switch config.Mode {

	// GPIO
	case PinInput, PinInputPullUp, PinInputPullDown,
		PinOutput, PinOutputOpenDrain, PinDisable:
		mode := uint32(0x5) // GPIO is always alternate function 5
		if forcePath {
			mode |= 0x10 // SION bit
		}
		return mode

	// ADC
	case PinInputAnalog:
		mode := uint32(0x5) // use alternate function 5 (GPIO)
		if forcePath {
			mode |= 0x10 // SION bit
		}
		return mode

	// UART RX/TX
	case PinModeUARTRX, PinModeUARTTX:
		mode := uint32(0x2) // UART is usually alternate function 2 on Teensy 4.x
		// Teensy 4.1 has a UART (LPUART5) with alternate function 1
		if p == PB28 || p == PB29 {
			mode = 0x1
		}
		return mode

	// SPI SDI
	case PinModeSPISDI:
		var mode uint32
		switch p {
		case PC15: // LPSPI1 SDI on PC15 alternate function 4
			mode = uint32(0x4)
		case PA2: // LPSPI3 SDI on PA2 alternate function 7
			mode = uint32(0x7)
		case PB1: // LPSPI4 SDI on PB1 alternate function 3
			mode = uint32(0x3)
		default:
			panic("machine: invalid SPI SDI pin")
		}
		if forcePath {
			mode |= 0x10 // SION bit
		}
		return mode

	// SPI SDO
	case PinModeSPISDO:
		var mode uint32
		switch p {
		case PC14: // LPSPI1 SDO on PC14 alternate function 4
			mode = uint32(0x4)
		case PA30: // LPSPI3 SDO on PA30 alternate function 2
			mode = uint32(0x2)
		case PB2: // LPSPI4 SDO on PB2 alternate function 3
			mode = uint32(0x3)
		default:
			panic("machine: invalid SPI SDO pin")
		}
		if forcePath {
			mode |= 0x10 // SION bit
		}
		return mode

	// SPI SCK
	case PinModeSPICLK:
		var mode uint32
		switch p {
		case PC12: // LPSPI1 SCK on PC12 alternate function 4
			mode = uint32(0x4)
		case PA31: // LPSPI3 SCK on PA31 alternate function 2
			mode = uint32(0x2)
		case PB3: // LPSPI4 SCK on PB3 alternate function 3
			mode = uint32(0x3)
		default:
			panic("machine: invalid SPI CLK pin")
		}
		if forcePath {
			mode |= 0x10 // SION bit
		}
		return mode

	// SPI CS
	case PinModeSPICS:
		var mode uint32
		switch p {
		case PC13: // LPSPI1 CS on PC13 alternate function 4
			mode = uint32(0x4)
		case PA3: // LPSPI3 CS on PA3 alternate function 7
			mode = uint32(0x7)
		case PB0: // LPSPI4 CS on PB0 alternate function 3
			mode = uint32(0x3)
		default: // use alternate function 5 (GPIO) if non-CS pin selected
			mode = uint32(0x5)
		}
		if forcePath {
			mode |= 0x10 // SION bit
		}
		return mode

	// I2C SDA
	case PinModeI2CSDA:
		var mode uint32
		switch p {
		case PA13: // LPI2C4 SDA on PA13 alternate function 0
			mode = uint32(0)
		case PA17: // LPI2C1 SDA on PA17 alternate function 3
			mode = uint32(3)
		case PA22: // LPI2C3 SDA on PA22 alternate function 1
			mode = uint32(1)
		default:
			panic("machine: invalid I2C SDA pin")
		}
		if forcePath {
			mode |= 0x10 // SION bit
		}
		return mode

	// I2C SCL
	case PinModeI2CSCL:
		var mode uint32
		switch p {
		case PA12: // LPI2C4 SCL on PA12 alternate function 0
			mode = uint32(0)
		case PA16: // LPI2C1 SCL on PA16 alternate function 3
			mode = uint32(3)
		case PA23: // LPI2C3 SCL on PA23 alternate function 1
			mode = uint32(1)
		default:
			panic("machine: invalid I2C SCL pin")
		}
		if forcePath {
			mode |= 0x10 // SION bit
		}
		return mode

	default:
		panic("machine: invalid pin mode")
	}
}

// maximum ADC value for the currently configured resolution (used for scaling)
var adcMaximum uint32

// InitADC is not used by this machine. Use `(ADC).Configure()`.
func InitADC() {}

// Configure initializes the receiver's ADC peripheral and pin for analog input.
func (a ADC) Configure(config ADCConfig) {
	// if not specified, use defaults: 10-bit resolution, 4 samples/conversion
	const (
		defaultResolution = uint32(10)
		defaultSamples    = uint32(4)
	)

	a.Pin.Configure(PinConfig{Mode: PinInputAnalog})

	resolution, samples := config.Resolution, config.Samples
	if 0 == resolution {
		resolution = defaultResolution
	}
	if 0 == samples {
		samples = defaultSamples
	}
	if resolution > 12 {
		resolution = 12 // maximum resolution of 12 bits
	}
	adcMaximum = (uint32(1) << resolution) - 1

	mode, average := a.mode(resolution, samples)

	nxp.ADC1.CFG.Set(mode | nxp.ADC_CFG_ADHSC) // configure ADC1
	nxp.ADC2.CFG.Set(mode | nxp.ADC_CFG_ADHSC) // configure ADC2

	// begin calibration
	nxp.ADC1.GC.Set(average | nxp.ADC_GC_CAL)
	nxp.ADC2.GC.Set(average | nxp.ADC_GC_CAL)

	for a.isCalibrating() {
	} // wait for calibration
}

// Get performs a single ADC conversion, returning a 16-bit unsigned integer.
// The value returned will be scaled (uniformly distributed) if necessary so
// that it is always in the range [0..65535], regardless of the ADC's configured
// bit size (resolution).
func (a ADC) Get() uint16 {
	if ch1, ch2, ok := a.Pin.getADCChannel(); ok {
		for a.isCalibrating() {
		} // wait for calibration
		var val uint32
		if noADCChannel != ch1 {
			nxp.ADC1.HC0.Set(uint32(ch1))
			for !nxp.ADC1.HS.HasBits(nxp.ADC_HS_COCO0) {
			}
			val = nxp.ADC1.R0.Get() & 0xFFFF
		} else {
			nxp.ADC2.HC0.Set(uint32(ch2))
			for !nxp.ADC2.HS.HasBits(nxp.ADC_HS_COCO0) {
			}
			val = nxp.ADC2.R0.Get() & 0xFFFF
		}
		// should never be zero, but just in case, use UINT16_MAX so that the scalar
		// gets factored out of the conversion result, leaving the original reading
		// to be returned unaltered/unscaled.
		if adcMaximum == 0 {
			adcMaximum = 0xFFFF
		}
		// scale up to a 16-bit value
		return uint16((val * 0xFFFF) / adcMaximum)
	}
	return 0
}

// mode constructs bit masks for mode and average - used in ADC configuration
// registers - from a given ADC bit size (resolution) and sample count.
func (a ADC) mode(resolution, samples uint32) (mode, average uint32) {

	// use asynchronous clock (ADACK) (0 = IPG, 1 = IPG/2, or 3 = ADACK)
	mode = (nxp.ADC_CFG_ADICLK_ADICLK_3 << nxp.ADC_CFG_ADICLK_Pos) & nxp.ADC_CFG_ADICLK_Msk

	// input clock DIV2 (0 = DIV1, 1 = DIV2, 2 = DIV4, or 3 = DIV8)
	mode |= (nxp.ADC_CFG_ADIV_ADIV_1 << nxp.ADC_CFG_ADIV_Pos) & nxp.ADC_CFG_ADIV_Msk

	switch resolution {
	case 8: // 8-bit conversion, sample period (ADC clocks) = 8
		mode |= (nxp.ADC_CFG_MODE_MODE_0 << nxp.ADC_CFG_MODE_Pos) & nxp.ADC_CFG_MODE_Msk
		mode |= (nxp.ADC_CFG_ADSTS_ADSTS_3 << nxp.ADC_CFG_ADSTS_Pos) & nxp.ADC_CFG_ADSTS_Msk

	case 12: // 12-bit conversion, sample period (ADC clocks) = 24
		mode |= (nxp.ADC_CFG_MODE_MODE_2 << nxp.ADC_CFG_MODE_Pos) & nxp.ADC_CFG_MODE_Msk
		mode |= (nxp.ADC_CFG_ADSTS_ADSTS_3 << nxp.ADC_CFG_ADSTS_Pos) & nxp.ADC_CFG_ADSTS_Msk
		mode |= nxp.ADC_CFG_ADLSMP

	default: // 10-bit conversion, sample period (ADC clocks) = 20
		mode |= (nxp.ADC_CFG_MODE_MODE_1 << nxp.ADC_CFG_MODE_Pos) & nxp.ADC_CFG_MODE_Msk
		mode |= (nxp.ADC_CFG_ADSTS_ADSTS_2 << nxp.ADC_CFG_ADSTS_Pos) & nxp.ADC_CFG_ADSTS_Msk
		mode |= nxp.ADC_CFG_ADLSMP
	}

	if samples >= 4 {
		if samples >= 32 {
			// 32 samples averaged
			mode |= (nxp.ADC_CFG_AVGS_AVGS_3 << nxp.ADC_CFG_AVGS_Pos) & nxp.ADC_CFG_AVGS_Msk
		} else if samples >= 16 {
			// 16 samples averaged
			mode |= (nxp.ADC_CFG_AVGS_AVGS_2 << nxp.ADC_CFG_AVGS_Pos) & nxp.ADC_CFG_AVGS_Msk
		} else if samples >= 8 {
			// 8 samples averaged
			mode |= (nxp.ADC_CFG_AVGS_AVGS_1 << nxp.ADC_CFG_AVGS_Pos) & nxp.ADC_CFG_AVGS_Msk
		} else {
			// 4 samples averaged
			mode |= (nxp.ADC_CFG_AVGS_AVGS_0 << nxp.ADC_CFG_AVGS_Pos) & nxp.ADC_CFG_AVGS_Msk
		}
		average = nxp.ADC_GC_AVGE
	}

	return mode, average
}

// isCalibrating returns true if and only if either one (or both) of ADC1 and
// ADC2 have their calibrating flags set. ADC reads must wait until these flags
// are clear before attempting a conversion.
func (a ADC) isCalibrating() bool {
	return nxp.ADC1.GC.HasBits(nxp.ADC_GC_CAL) || nxp.ADC2.GC.HasBits(nxp.ADC_GC_CAL)
}

const noADCChannel = uint8(0xFF)

// getADCChannel returns the input channel for ADC1/ADC2 of the receiver Pin p.
func (p Pin) getADCChannel() (adc1, adc2 uint8, ok bool) {
	switch p {
	case PA12: // [AD_B0_12]:       ADC1_IN1        ~
		return 1, noADCChannel, true
	case PA13: // [AD_B0_13]:       ADC1_IN2        ~
		return 2, noADCChannel, true
	case PA14: // [AD_B0_14]:       ADC1_IN3        ~
		return 3, noADCChannel, true
	case PA15: // [AD_B0_15]:       ADC1_IN4        ~
		return 4, noADCChannel, true
	case PA16: // [AD_B1_00]:       ADC1_IN5     ADC2_IN5
		return 5, 5, true
	case PA17: // [AD_B1_01]:       ADC1_IN6     ADC2_IN6
		return 6, 6, true
	case PA18: // [AD_B1_02]:       ADC1_IN7     ADC2_IN7
		return 7, 7, true
	case PA19: // [AD_B1_03]:       ADC1_IN8     ADC2_IN8
		return 8, 8, true
	case PA20: // [AD_B1_04]:       ADC1_IN9     ADC2_IN9
		return 9, 9, true
	case PA21: // [AD_B1_05]:       ADC1_IN10    ADC2_IN10
		return 10, 10, true
	case PA22: // [AD_B1_06]:       ADC1_IN11    ADC2_IN11
		return 11, 11, true
	case PA23: // [AD_B1_07]:       ADC1_IN12    ADC2_IN12
		return 12, 12, true
	case PA24: // [AD_B1_08]:       ADC1_IN13    ADC2_IN13
		return 13, 13, true
	case PA25: // [AD_B1_09]:       ADC1_IN14    ADC2_IN14
		return 14, 14, true
	case PA26: // [AD_B1_10]:       ADC1_IN15    ADC2_IN15
		return 15, 15, true
	case PA27: // [AD_B1_11]:       ADC1_IN0     ADC2_IN0
		return 16, 16, true
	case PA28: // [AD_B1_12]:          ~         ADC2_IN1
		return noADCChannel, 1, true
	case PA29: // [AD_B1_13]:          ~         ADC2_IN2
		return noADCChannel, 2, true
	case PA30: // [AD_B1_14]:          ~         ADC2_IN3
		return noADCChannel, 3, true
	case PA31: // [AD_B1_15]:          ~         ADC2_IN4
		return noADCChannel, 4, true
	default:
		return noADCChannel, noADCChannel, false
	}
}
