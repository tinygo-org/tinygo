//go:build esp32c3 || esp32c6 || esp32h2 || esp32p4

package machine

import (
	"device/esp"
	"runtime/interrupt"
	"runtime/volatile"
)

type UART struct {
	Bus                  *esp.UART_Type
	Buffer               *RingBuffer
	ParityErrorDetected  bool // set when parity error detected
	DataErrorDetected    bool // set when data corruption detected
	DataOverflowDetected bool // set when data overflow detected in UART FIFO buffer or RingBuffer
}

const (
	defaultDataBits = 8
	defaultStopBit  = 1
	defaultParity   = ParityNone

	uartInterrupts = esp.UART_INT_ENA_RXFIFO_FULL_INT_ENA |
		esp.UART_INT_ENA_PARITY_ERR_INT_ENA |
		esp.UART_INT_ENA_FRM_ERR_INT_ENA |
		esp.UART_INT_ENA_RXFIFO_OVF_INT_ENA |
		esp.UART_INT_ENA_GLITCH_DET_INT_ENA

	uart_empty_thresh_default = 10
)

func (uart *UART) Configure(config UARTConfig) error {
	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}
	if uart.Bus == esp.UART0 && config.TX == 0 && config.RX == 0 {
		config.TX = UART0_TX_PIN
		config.RX = UART0_RX_PIN
	}
	if config.TX == config.RX {
		return errSamePins
	}
	bus := busFor(uart.Bus)
	if bus.interruptMapReg == nil {
		return errWrongUART
	}
	if err := validateFormat(defaultDataBits, defaultStopBit); err != nil {
		return err
	}
	initUARTClock(uart.Bus, bus)
	return uart.configure(config, bus)
}

func (uart *UART) SetFormat(dataBits, stopBits int, parity UARTParity) error {
	if err := validateFormat(dataBits, stopBits); err != nil {
		return err
	}
	uart.setFormat(dataBits, stopBits, parity)
	uartSync(uart.Bus)
	return nil
}

func validateFormat(dataBits, stopBits int) error {
	if dataBits < 5 {
		return errWrongBitSize
	}
	if stopBits > 1 {
		return errWrongStopBitSize
	}
	return nil
}

func (uart *UART) setupPins(config UARTConfig, regs registerSet) {
	config.RX.Configure(PinConfig{Mode: PinInputPullup})
	config.TX.Configure(PinConfig{Mode: PinInputPullup})
	uartEnableTX(config.TX)

	// link TX with GPIO signal X (technical reference manual, GPIO matrix)
	if config.InvertTX {
		config.TX.outFunc().Set(regs.gpioMatrixSignal | esp.GPIO_FUNC_OUT_SEL_CFG_INV_SEL)
	} else {
		config.TX.outFunc().Set(regs.gpioMatrixSignal)
	}
	// link RX with GPIO signal X and route signals via GPIO matrix
	if config.InvertRX {
		inFunc(regs.gpioMatrixSignal).Set(esp.GPIO_FUNC_IN_SEL_CFG_SEL | uint32(config.RX) | esp.GPIO_FUNC_IN_SEL_CFG_IN_INV_SEL)
	} else {
		inFunc(regs.gpioMatrixSignal).Set(esp.GPIO_FUNC_IN_SEL_CFG_SEL | uint32(config.RX))
	}
}

func (uart *UART) configureInterrupt(intrMapReg *volatile.Register32) {
	// Disable all UART interrupts
	uart.Bus.INT_ENA.ClearBits(0x0ffff)

	intrMapReg.Set(cpuInterruptFromUART)
	onceUart.Do(func() {
		_ = interrupt.New(cpuInterruptFromUART, func(i interrupt.Interrupt) {
			UART0.serveInterrupt(0)
			UART1.serveInterrupt(1)
		}).Enable()
	})
}

func (uart *UART) serveInterrupt(num int) {
	// get interrupt status
	interruptFlag := uart.Bus.INT_ST.Get()
	if (interruptFlag & uartInterrupts) == 0 {
		return
	}

	// block UART interrupts while processing
	uart.Bus.INT_ENA.ClearBits(uartInterrupts)

	if interruptFlag&esp.UART_INT_ENA_RXFIFO_FULL_INT_ENA > 0 {
		for uart.Bus.GetSTATUS_RXFIFO_CNT() > 0 {
			b := uart.Bus.GetFIFO_RXFIFO_RD_BYTE()
			if !uart.Buffer.Put(byte(b & 0xff)) {
				uart.DataOverflowDetected = true
			}
		}
	}
	if interruptFlag&esp.UART_INT_ENA_PARITY_ERR_INT_ENA > 0 {
		uart.ParityErrorDetected = true
	}
	if 0 != interruptFlag&esp.UART_INT_ENA_FRM_ERR_INT_ENA {
		uart.DataErrorDetected = true
	}
	if 0 != interruptFlag&esp.UART_INT_ENA_RXFIFO_OVF_INT_ENA {
		uart.DataOverflowDetected = true
	}
	if 0 != interruptFlag&esp.UART_INT_ENA_GLITCH_DET_INT_ENA {
		uart.DataErrorDetected = true
	}

	// Clear the UART interrupt status
	uart.Bus.INT_CLR.SetBits(interruptFlag)
	uart.Bus.INT_CLR.ClearBits(interruptFlag)
	// Enable interrupts
	uart.Bus.INT_ENA.Set(uartInterrupts)
}

func (uart *UART) flush() {}

// uartDividers gives the baud rate divider, in 1/16 steps, and the source clock
// divider for baudRate. See ESP-IDF hal/esp32c3/include/hal/uart_ll.h uart_ll_set_baudrate.
func uartDividers(baudRate uint32) (clkDiv, sclkDiv uint32) {
	const maxDiv = 1<<12 - 1
	if baudRate == 0 {
		baudRate = 115200
	}
	baud := uint64(baudRate)
	sclk := (uartClockFreq + maxDiv*baud - 1) / (maxDiv * baud)
	clk := (uartClockFreq << 4) / (baud * sclk)
	return uint32(clk), uint32(sclk)
}
