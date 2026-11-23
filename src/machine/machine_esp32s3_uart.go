//go:build esp32s3

package machine

import (
	"device/esp"
	"errors"
	"runtime/interrupt"
	"runtime/volatile"
)

func init() {
	UART0.interrupt = interrupt.New(cpuInterruptFromUART0, _UART0.serveInterrupt)
	UART1.interrupt = interrupt.New(cpuInterruptFromUART1, _UART1.serveInterrupt)
	UART2.interrupt = interrupt.New(cpuInterruptFromUART2, _UART2.serveInterrupt)
}

var (
	DefaultUART = UART0

	UART0  = &_UART0
	_UART0 = UART{Bus: esp.UART0, Buffer: NewRingBuffer()}
	UART1  = &_UART1
	_UART1 = UART{Bus: esp.UART1, Buffer: NewRingBuffer()}
	UART2  = &_UART2
	_UART2 = UART{Bus: esp.UART2, Buffer: NewRingBuffer()}

	errSamePins         = errors.New("UART: invalid pin combination")
	errWrongUART        = errors.New("UART: unsupported UARTn")
	errWrongBitSize     = errors.New("UART: invalid data size")
	errWrongStopBitSize = errors.New("UART: invalid bit size")
)

const uartClockFreq = 80_000_000 // APB_CLK, SCLK_SEL 1. See ESP-IDF hal/esp32s3/include/hal/uart_ll.h uart_ll_set_sclk.

// Free level-triggered CPU interrupts. See ESP-IDF
// esp_hw_support/port/esp32s3/esp_cpu_intr.c.
const (
	cpuInterruptFromUART0 = 1
	cpuInterruptFromUART1 = 2
	cpuInterruptFromUART2 = 5
)

type registerSet struct {
	cpuInterrupt     uint32
	interruptMapReg  *volatile.Register32
	clockReg         *volatile.Register32
	resetReg         *volatile.Register32
	uartClockBitMask uint32
	gpioMatrixSignal uint32
}

// busFor gives the registers of a UART. The GPIO matrix signals are from ESP-IDF
// soc/esp32s3/include/soc/gpio_sig_map.h.
func busFor(bus *esp.UART_Type) registerSet {
	switch bus {
	case esp.UART0:
		return registerSet{
			cpuInterrupt:     cpuInterruptFromUART0,
			interruptMapReg:  &esp.INTERRUPT_CORE0.UART_INTR_MAP,
			clockReg:         &esp.SYSTEM.PERIP_CLK_EN0,
			resetReg:         &esp.SYSTEM.PERIP_RST_EN0,
			uartClockBitMask: esp.SYSTEM_PERIP_CLK_EN0_UART_CLK_EN,
			gpioMatrixSignal: 12,
		}
	case esp.UART1:
		return registerSet{
			cpuInterrupt:     cpuInterruptFromUART1,
			interruptMapReg:  &esp.INTERRUPT_CORE0.UART1_INTR_MAP,
			clockReg:         &esp.SYSTEM.PERIP_CLK_EN0,
			resetReg:         &esp.SYSTEM.PERIP_RST_EN0,
			uartClockBitMask: esp.SYSTEM_PERIP_CLK_EN0_UART1_CLK_EN,
			gpioMatrixSignal: 15,
		}
	case esp.UART2:
		return registerSet{
			cpuInterrupt:     cpuInterruptFromUART2,
			interruptMapReg:  &esp.INTERRUPT_CORE0.UART2_INTR_MAP,
			clockReg:         &esp.SYSTEM.PERIP_CLK_EN1,
			resetReg:         &esp.SYSTEM.PERIP_RST_EN1,
			uartClockBitMask: esp.SYSTEM_PERIP_CLK_EN1_UART2_CLK_EN,
			gpioMatrixSignal: 18,
		}
	default:
		return registerSet{}
	}
}

func (uart *UART) configure(config UARTConfig, regs registerSet) error {
	uart.Bus.SetCLK_CONF_TX_SCLK_EN(0)
	uart.Bus.SetCLK_CONF_RX_SCLK_EN(0)

	uart.Bus.SetCLK_CONF_SCLK_SEL(1)
	uart.Bus.SetCLK_CONF_SCLK_DIV_NUM(0)
	uart.Bus.SetCLK_CONF_SCLK_DIV_A(0)
	uart.Bus.SetCLK_CONF_SCLK_DIV_B(0)

	uart.SetBaudRate(config.BaudRate)
	uart.setFormat(defaultDataBits, defaultStopBit, defaultParity)
	uart.Bus.SetRS485_CONF_RS485_EN(0)
	uart.Bus.SetRS485_CONF_RS485TX_RX_EN(0)
	uart.Bus.SetRS485_CONF_RS485RXBY_TX_EN(0)
	uart.Bus.SetCONF0_IRDA_EN(0)
	uart.Bus.SetCONF0_TX_FLOW_EN(0)
	uart.Bus.SetCONF1_RX_FLOW_EN(0)

	uart.setupPins(config, regs)
	uart.configureInterrupt(regs)
	uart.enableTransmitter()
	uart.enableReceiver()

	uart.Bus.SetCLK_CONF_TX_SCLK_EN(1)
	uart.Bus.SetCLK_CONF_RX_SCLK_EN(1)
	return nil
}

// uartSync does nothing. The S3 configuration registers do not need a sync.
// See ESP-IDF hal/esp32s3/include/hal/uart_ll.h uart_ll_set_data_bit_num.
func uartSync(bus *esp.UART_Type) {}

func (uart *UART) setFormat(dataBits, stopBits int, parity UARTParity) {
	uart.Bus.SetCONF0_BIT_NUM(uint32(dataBits - 5))
	uart.Bus.SetCONF0_STOP_BIT_NUM(uint32(stopBits))
	switch parity {
	case ParityNone:
		uart.Bus.SetCONF0_PARITY_EN(0)
	case ParityEven:
		uart.Bus.SetCONF0_PARITY_EN(1)
		uart.Bus.SetCONF0_PARITY(0)
	case ParityOdd:
		uart.Bus.SetCONF0_PARITY_EN(1)
		uart.Bus.SetCONF0_PARITY(1)
	}
}

func initUARTClock(bus *esp.UART_Type, regs registerSet) {
	esp.SYSTEM.PERIP_CLK_EN0.SetBits(esp.SYSTEM_PERIP_CLK_EN0_UART_MEM_CLK_EN)
	regs.clockReg.SetBits(regs.uartClockBitMask)
	regs.resetReg.ClearBits(regs.uartClockBitMask)
	bus.SetCLK_CONF_RST_CORE(1)
	regs.resetReg.SetBits(regs.uartClockBitMask)
	regs.resetReg.ClearBits(regs.uartClockBitMask)
	bus.SetCLK_CONF_RST_CORE(0)
}

func (uart *UART) SetBaudRate(baudRate uint32) {
	clk_div, sclk_div := uartDividers(baudRate)
	uart.Bus.SetCLKDIV(clk_div >> 4)
	uart.Bus.SetCLKDIV_FRAG(clk_div & 0xf)
	uart.Bus.SetCLK_CONF_SCLK_DIV_NUM(sclk_div - 1)
}

func uartEnableTX(pin Pin) {}

func uartSetTXFIFORST(bus *esp.UART_Type, value uint32) {
	bus.SetCONF0_TXFIFO_RST(value)
}

func uartSetRXFIFORST(bus *esp.UART_Type, value uint32) {
	bus.SetCONF0_RXFIFO_RST(value)
}
