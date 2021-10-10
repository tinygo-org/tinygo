// +build esp32c3

package machine

import (
	"device/esp"
	"device/riscv"
	"errors"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

const (
	clk_freq                  = 80 * 1000000
	uart_empty_thresh_default = 10
)

var DefaultUART = UART0
var ErrSamePins = errors.New("machine: same pin used for UART's RX and TX")
var ErrDataBits = errors.New("machine: invalid data size")

var (
	UART0  = &_UART0
	_UART0 = UART{Bus: esp.UART0, Buffer: NewRingBuffer(), module: PERIPH_UART0_MODULE}
	UART1  = &_UART1
	_UART1 = UART{Bus: esp.UART1, Buffer: NewRingBuffer(), module: PERIPH_UART1_MODULE}
)

type UART struct {
	module               Module
	Bus                  *esp.UART_Type
	Buffer               *RingBuffer
	ParityErrorDetected  bool
	DataErrorDetected    bool
	DataOverflowDetected bool
}

func (uart *UART) Configure(config *UARTConfig) error {
	// check configuration parameters
	if err := uart.verifyConfig(config); err != nil {
		return err
	}

	uart.initUART()
	uart.configure(config)

	// Finish Config process
	uart.Bus.ID.SetBits(esp.UART_ID_REG_UPDATE)

	// Setup GPIO Pins
	uart.setupPins(config)
	uart.setupInterrupt()
	uart.enableTransmitter()
	uart.enableReceiver()

	// Start TX/RX
	uart.Bus.CLK_CONF.SetBits(esp.UART_CLK_CONF_TX_SCLK_EN_Msk | esp.UART_CLK_CONF_RX_SCLK_EN_Msk)
	return nil
}

func (uart *UART) Good() bool {
	return !(uart.ParityErrorDetected || uart.DataErrorDetected || uart.DataOverflowDetected)
}

func (uart *UART) WriteByte(b byte) error {
	for (uart.Bus.STATUS.Get()&esp.UART_STATUS_TXFIFO_CNT_Msk)>>esp.UART_STATUS_TXFIFO_CNT_Pos >= 128 {
		// Read UART_TXFIFO_CNT from the status register, which indicates how
		// many bytes there are in the transmit buffer. Wait until there are
		// less than 128 bytes in this buffer (the default buffer size).
	}

	// Write data to be sent to UART_RXFIFO_RD_BYTE
	uart.Bus.FIFO.Set(uint32(b))
	return nil
}

func (uart *UART) verifyConfig(config *UARTConfig) error {

	if err := isUARTPinValid(config.RX); err != nil {
		return err
	}
	if err := isUARTPinValid(config.TX); err != nil {
		return err
	}
	if config.TX == config.RX {
		return ErrSamePins
	}

	if config.DataBits == 0 {
		config.DataBits = 8
	} else if config.DataBits < 5 || config.DataBits > 8 {
		return ErrDataBits
	}
	if config.StopBits == UARTStopBits_Default {
		config.StopBits = UARTStopBits_1
	}

	return nil
}

func isUARTPinValid(pin Pin) error {
	if pin < 0 || pin > 21 {
		return ErrInvalidPinNumber
	}
	return nil
}

func (uart *UART) initUART() {
	uartClock := uart.module.getClockEnableRegister()
	uartClockReset := uart.module.getClockResetRegister()

	// To initialize URATn based on technical reference document:
	// - enable the clock for UART RAM by setting SYSTEM_UART_MEM_CLK_EN to 1 (SYSTEM_CPU_PERI_CLK_EN0_REG)
	uartClock.SetBits(esp.SYSTEM_PERIP_CLK_EN0_UART_MEM_CLK_EN)
	// - enable APB_CLK for UARTn by setting SYSTEM_UARTn_CLK_EN to 1 (SYSTEM_PERIP_CLK_EN0_UART1_CLK_EN)
	uartClock.SetBits(uart.module.getClockBit())
	// - clear SYSTEM_UARTn_RST (SYSTEM_PERIP_RST_EN0_UART1_RST)
	uartClockReset.ClearBits(uart.module.getClockBit())
	// - write 1 to UART_RST_CORE
	uart.Bus.CLK_CONF.SetBits(esp.UART_CLK_CONF_RST_CORE)
	// - write 1 to SYSTEM_UARTn_RST (SYSTEM_PERIP_RST_EN0_UART1_RST)
	// - clear SYSTEM_UARTn_RST (SYSTEM_PERIP_RST_EN0_UART1_RST)
	uartClockReset.SetBits(uart.module.getClockBit())
	uartClockReset.ClearBits(uart.module.getClockBit())
	// - clear UART_RST_CORE
	uart.Bus.CLK_CONF.ClearBits(esp.UART_CLK_CONF_RST_CORE)
	// enable register synchronization by clearing UART_UPDATE_CTRL
	uart.Bus.ID.ClearBits(esp.UART_ID_REG_UPDATE_Msk)

	// enable RTC clock ???
	esp.RTC_CNTL.RTC_CLK_CONF.SetBits(esp.RTC_CNTL_RTC_CLK_CONF_DIG_CLK8M_EN_Msk)

	// wait for Core Clock to ready for configuration
	for uart.Bus.ID.HasBits(esp.UART_ID_REG_UPDATE_Msk) {
		riscv.Asm("nop")
	}
}

func (uart *UART) configure(config *UARTConfig) {

	// Write static registers
	// - disbale TX/RX clock to make sure the UART transmitter or receiver is not at work
	uart.Bus.CLK_CONF.ClearBits(esp.UART_CLK_CONF_TX_SCLK_EN_Msk | esp.UART_CLK_CONF_RX_SCLK_EN_Msk)

	// To configure URATn communication:
	// Configure static registers (if any) following Section 20.5.1.2;

	// - Set default clock source (UART_SCLK_APB)
	// UART_SCLK_SEL UART clock source select. 1: APB_CLK; 2: FOSC_CLK; 3: XTAL_CLK.
	// uart.Bus.CLK_CONF.ReplaceBits(1, esp.UART_CLK_CONF_SCLK_SEL_Msk, esp.UART_CLK_CONF_SCLK_SEL_Pos)
	uart.Bus.CLK_CONF.ReplaceBits(1<<esp.UART_CLK_CONF_SCLK_SEL_Pos, esp.UART_CLK_CONF_SCLK_SEL_Msk, 0)
	// configure divisor of the divider via UART_SCLK_DIV_NUM, UART_SCLK_DIV_A, and UART_SCLK_DIV_B
	// uart.Bus.CLK_CONF.ReplaceBits(0, esp.UART_CLK_CONF_SCLK_DIV_NUM_Msk, esp.UART_CLK_CONF_SCLK_DIV_NUM_Pos)
	uart.Bus.CLK_CONF.ReplaceBits(0, esp.UART_CLK_CONF_SCLK_DIV_NUM_Msk, 0)
	// uart.Bus.CLK_CONF.ReplaceBits(0, esp.UART_CLK_CONF_SCLK_DIV_A_Msk, esp.UART_CLK_CONF_SCLK_DIV_A_Pos)
	uart.Bus.CLK_CONF.ReplaceBits(0, esp.UART_CLK_CONF_SCLK_DIV_A_Msk, 0)
	// uart.Bus.CLK_CONF.ReplaceBits(0, esp.UART_CLK_CONF_SCLK_DIV_B_Msk, esp.UART_CLK_CONF_SCLK_DIV_B_Pos)
	uart.Bus.CLK_CONF.ReplaceBits(0, esp.UART_CLK_CONF_SCLK_DIV_B_Msk, 0)
	// configure the baud rate for transmission via UART_CLKDIV and UART_CLKDIV_FRAG
	// - Set baud rate
	max_div := uint32((1 << 12) - 1)
	clk_div := (clk_freq + (max_div * config.BaudRate) - 1) / (max_div * config.BaudRate)
	div := (clk_freq << 4) / (config.BaudRate * clk_div)
	// The baud rate configuration register is divided into an integer part and a fractional part.
	// uart.Bus.CLKDIV.ReplaceBits((div >> 4), esp.UART_CLKDIV_CLKDIV_Msk, esp.UART_CLKDIV_CLKDIV_Pos)
	uart.Bus.CLKDIV.ReplaceBits((div>>4)<<esp.UART_CLKDIV_CLKDIV_Pos, esp.UART_CLKDIV_CLKDIV_Msk, 0)
	// uart.Bus.CLKDIV.ReplaceBits((div & 0xf), esp.UART_CLKDIV_FRAG_Msk, esp.UART_CLKDIV_FRAG_Pos)
	uart.Bus.CLKDIV.ReplaceBits((div&0xf)<<esp.UART_CLKDIV_FRAG_Pos, esp.UART_CLKDIV_FRAG_Msk, 0)
	// configure data length via UART_BIT_NUM;
	// uart.Bus.CONF0.ReplaceBits(uint32(config.DataBits-5), esp.UART_CONF0_BIT_NUM_Msk, esp.UART_CONF0_BIT_NUM_Pos)
	uart.Bus.CONF0.ReplaceBits(uint32(config.DataBits-5)<<esp.UART_CONF0_BIT_NUM_Pos, esp.UART_CONF0_BIT_NUM_Msk, 0)
	// - stop bit
	// uart.Bus.CONF0.ReplaceBits(uint32(config.StopBits), esp.UART_CONF0_STOP_BIT_NUM_Msk, esp.UART_CONF0_STOP_BIT_NUM_Pos)
	uart.Bus.CONF0.ReplaceBits(uint32(config.StopBits)<<esp.UART_CONF0_STOP_BIT_NUM_Pos, esp.UART_CONF0_STOP_BIT_NUM_Msk, 0)
	// configure odd or even parity check via UART_PARITY_EN and UART_PARITY;
	switch config.Parity {
	case ParityNone:
		uart.Bus.CONF0.ClearBits(esp.UART_CONF0_PARITY_EN)
	case ParityEven:
		uart.Bus.CONF0.SetBits(esp.UART_CONF0_PARITY_EN)
		uart.Bus.CONF0.ClearBits(esp.UART_CONF0_PARITY)
	case ParityOdd:
		uart.Bus.CONF0.SetBits(esp.UART_CONF0_PARITY_EN)
		uart.Bus.CONF0.SetBits(esp.UART_CONF0_PARITY)
	}
	// - Set UART mode.
	uart.Bus.RS485_CONF.ClearBits(esp.UART_RS485_CONF_RS485_EN_Msk | esp.UART_RS485_CONF_RS485TX_RX_EN_Msk | esp.UART_RS485_CONF_RS485RXBY_TX_EN_Msk)
	uart.Bus.CONF0.ClearBits(esp.UART_CONF0_IRDA_EN)
	// - Disable hw-flow control
	uart.Bus.CONF0.ClearBits(esp.UART_CONF0_TX_FLOW_EN)
	uart.Bus.CONF1.ClearBits(esp.UART_CONF1_RX_FLOW_EN)
}

func (uart *UART) setupPins(config *UARTConfig) error {
	config.RX.Configure(PinConfig{Mode: PinInputPullup})
	config.TX.Configure(PinConfig{Mode: PinInputPullup})

	// link TX with signal 9 (technical reference manual 5.10) (this is not interrupt signal!)
	reg := (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(&esp.GPIO.FUNC0_OUT_SEL_CFG)) + uintptr(config.TX)*4)))
	reg.Set(9)

	// link RX with signal 9 and route signals via GPIO matrix (GPIO_SIGn_IN_SEL 0x40)
	reg = (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(&esp.GPIO.FUNC0_IN_SEL_CFG)) + uintptr(9)*4)))
	reg.Set(uint32(config.RX) | 0x40)

	return nil
}

func (uart *UART) setupInterrupt() {
	// Disable all UART interrupts
	uart.Bus.INT_ENA.ClearBits(0x0fffff)
	// Clear the UART interrupt status
	uart.Bus.INT_CLR.SetBits(0x0fffff)
	uart.Bus.INT_CLR.ClearBits(0x0fffff)
	// define interrupt function and map UART to CPU interrupt
	if uart.module == PERIPH_UART0_MODULE {
		intr := interrupt.New(int(PERIPH_UART0_MODULE), func(i interrupt.Interrupt) {
			UART0.serveInterrupt()
		})
		intr.Enable(PERIPH_UART0_MODULE.interruptForModule(), PERIPH_UART0_MODULE.mapRegisterForModule())
	} else {
		intr := interrupt.New(int(PERIPH_UART1_MODULE), func(i interrupt.Interrupt) {
			UART1.serveInterrupt()
		})
		intr.Enable(PERIPH_UART1_MODULE.interruptForModule(), PERIPH_UART1_MODULE.mapRegisterForModule())
	}
}

func (uart *UART) enableTransmitter() {
	uart.Bus.CONF0.SetBits(esp.UART_CONF0_TXFIFO_RST)
	uart.Bus.CONF0.ClearBits(esp.UART_CONF0_TXFIFO_RST)
	// TXINFO empty threshold is when txfifo_empty_int interrupt produced after the amount of data in Tx-FIFO is less than this register value.
	// uart.Bus.CONF1.ReplaceBits(uart_empty_thresh_default, esp.UART_CONF1_TXFIFO_EMPTY_THRHD_Msk, esp.UART_CONF1_TXFIFO_EMPTY_THRHD_Pos)
	uart.Bus.CONF1.ReplaceBits(uart_empty_thresh_default<<esp.UART_CONF1_TXFIFO_EMPTY_THRHD_Pos, esp.UART_CONF1_TXFIFO_EMPTY_THRHD_Msk, 0)
	// we are not using interrut on TX since write we are waiting for FIFO to have space.
	// uart.Bus.INT_ENA.SetBits(esp.UART_INT_ENA_TXFIFO_EMPTY_INT_ENA)
}

func (uart *UART) enableReceiver() {
	uart.Bus.CONF0.SetBits(esp.UART_CONF0_RXFIFO_RST)
	uart.Bus.CONF0.ClearBits(esp.UART_CONF0_RXFIFO_RST)
	// using value 1 so that we can start populate ring buffer with data as we get it
	// uart.Bus.CONF1.ReplaceBits(1, esp.UART_CONF1_RXFIFO_FULL_THRHD_Msk, esp.UART_CONF1_RXFIFO_FULL_THRHD_Pos)
	uart.Bus.CONF1.ReplaceBits(1<<esp.UART_CONF1_RXFIFO_FULL_THRHD_Pos, esp.UART_CONF1_RXFIFO_FULL_THRHD_Msk, 0)
	// enable UART_RXFIFO_FULL_INT interrupt by setting UART_RXFIFO_FULL_INT_ENA
	uart.Bus.INT_ENA.SetBits(esp.UART_INT_ENA_RXFIFO_FULL_INT_ENA |
		esp.UART_INT_ENA_FRM_ERR_INT_ENA |
		esp.UART_INT_ENA_PARITY_ERR_INT_ENA |
		esp.UART_INT_ENA_GLITCH_DET_INT_ENA |
		esp.UART_INT_ENA_RXFIFO_OVF_INT_ENA)
}

func (uart *UART) serveInterrupt() {
	// get interrupt status
	interrutFlag := uart.Bus.INT_ST.Get()
	// mask will be used to enable interrupts back
	mask := uart.Bus.INT_ENA.Get() & interrutFlag

	// Disable UART interrupts
	uart.Bus.INT_ENA.ClearBits(mask)

	if 0 != interrutFlag&esp.UART_INT_ENA_RXFIFO_FULL_INT_ENA {
		b := uart.Bus.FIFO.Get()
		if !uart.Buffer.Put(byte(b & 0xff)) {
			uart.DataOverflowDetected = true
		}
	}
	if 0 != interrutFlag&esp.UART_INT_ENA_PARITY_ERR_INT_ENA {
		uart.ParityErrorDetected = true
	}
	if 0 != interrutFlag&esp.UART_INT_ENA_FRM_ERR_INT_ENA {
		uart.DataErrorDetected = true
	}
	if 0 != interrutFlag&esp.UART_INT_ENA_RXFIFO_OVF_INT_ENA {
		uart.DataOverflowDetected = true
	}
	if 0 != interrutFlag&esp.UART_INT_ENA_GLITCH_DET_INT_ENA {
		uart.DataErrorDetected = true
	}
	// * Keep all interrupt in case we need them but comment to reduce code size
	// if 0 != interrutFlag&esp.UART_INT_ENA_TXFIFO_EMPTY_INT_ENA {
	// 	println(" UART", uart.module, ":This is the enable bit for UART_TXFIFO_EMPTY_INT_ST register. (R/W)")
	// }
	// if 0 != interrutFlag&esp.UART_INT_ENA_DSR_CHG_INT_ENA {
	// 	println(" UART", uart.module, ":This is the enable bit for UART_DSR_CHG_INT_ST register. (R/W)")
	// }
	// if 0 != interrutFlag&esp.UART_INT_ENA_CTS_CHG_INT_ENA {
	// 	println(" UART", uart.module, ":This is the enable bit for UART_CTS_CHG_INT_ST register. (R/W)")
	// }
	// if 0 != interrutFlag&esp.UART_INT_ENA_BRK_DET_INT_ENA {
	// 	println(" UART", uart.module, ":This is the enable bit for UART_BRK_DET_INT_ST register. (R/W)")
	// }
	// if 0 != interrutFlag&esp.UART_INT_ENA_RXFIFO_TOUT_INT_ENA {
	// 	println(" UART", uart.module, ":This is the enable bit for UART_RXFIFO_TOUT_INT_ST register. (R/W)")
	// }
	// if 0 != interrutFlag&esp.UART_INT_ENA_SW_XON_INT_ENA {
	// 	println(" UART", uart.module, ":This is the enable bit for UART_SW_XON_INT_ST register. (R/W)")
	// }
	// if 0 != interrutFlag&esp.UART_INT_ENA_SW_XOFF_INT_ENA {
	// 	println(" UART", uart.module, ":This is the enable bit for UART_SW_XOFF_INT_ST register. (R/W)")
	// }
	// if 0 != interrutFlag&esp.UART_INT_ENA_TX_BRK_DONE_INT_ENA {
	// 	println(" UART", uart.module, ":This is the enable bit for UART_TX_BRK_DONE_INT_ST register. (R/W)")
	// }
	// if 0 != interrutFlag&esp.UART_INT_ENA_TX_BRK_IDLE_DONE_INT_ENA {
	// 	println(" UART", uart.module, ":This is the enable bit for UART_TX_BRK_IDLE_DONE_INT_ST register. (R/W)")
	// }
	// if 0 != interrutFlag&esp.UART_INT_ENA_TX_DONE_INT_ENA {
	// 	println(" UART", uart.module, ":This is the enable bit for UART_TX_DONE_INT_ST register. (R/W)")
	// }
	// if 0 != interrutFlag&esp.UART_INT_ENA_RS485_PARITY_ERR_INT_ENA {
	// 	println(" UART", uart.module, ":This is the enable bit for UART_RS485_PARITY_ERR_INT_ST register. (R/W)")
	// }
	// if 0 != interrutFlag&esp.UART_INT_ENA_RS485_FRM_ERR_INT_ENA {
	// 	println(" UART", uart.module, ":This is the enable bit for UART_RS485_PARITY_ERR_INT_ST register. (R/W)")
	// }
	// if 0 != interrutFlag&esp.UART_INT_ENA_RS485_CLASH_INT_ENA {
	// 	println(" UART", uart.module, ":This is the enable bit for UART_RS485_CLASH_INT_ST register. (R/W)")
	// }
	// if 0 != interrutFlag&esp.UART_INT_ENA_AT_CMD_CHAR_DET_INT_ENA {
	// 	println(" UART", uart.module, ":This is the enable bit for UART_AT_CMD_CHAR_DET_INT_ST register. (R/W)")
	// }
	// if 0 != interrutFlag&esp.UART_INT_ENA_WAKEUP_INT_ENA {
	// 	println(" UART", uart.module, ":This is the enable bit for UART_WAKEUP_INT_ST register. (R/W)")
	// }

	// Clear the UART interrupt status
	uart.Bus.INT_CLR.SetBits(interrutFlag)
	// Enable interrupts
	if mask != 0 {
		uart.Bus.INT_ENA.SetBits(mask)
	}
}
