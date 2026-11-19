//go:build esp32c3 || esp32c6 || esp32h2 || esp32p4

package machine

import "device/esp"

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
