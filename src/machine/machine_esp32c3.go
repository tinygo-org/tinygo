//go:build esp32c3

package machine

import (
	"device/esp"
	"device/riscv"
	"errors"
	"runtime/interrupt"
	"runtime/volatile"
	"sync"
	"unsafe"
)

const deviceName = esp.Device
const maxPin = 22
const cpuInterruptFromPin = 6

// CPUFrequency returns the current CPU frequency of the chip.
// Currently it is a fixed frequency but it may allow changing in the future.
func CPUFrequency() uint32 {
	return 160e6 // 160MHz
}

const (
	PinOutput PinMode = iota
	PinInput
	PinInputPullup
	PinInputPulldown
	PinAnalog
)

const (
	GPIO0 Pin = 0
	GPIO1 Pin = 1
	GPIO2 Pin = 2
	GPIO3 Pin = 3
	GPIO4 Pin = 4
	GPIO5 Pin = 5
	GPIO6 Pin = 6
)

const (
	ADC0 Pin = GPIO0
	ADC1 Pin = GPIO1
	ADC2 Pin = GPIO2
	ADC3 Pin = GPIO3
	ADC4 Pin = GPIO4
	ADC5 Pin = GPIO5 // avoid when WiFi is used.
)

const (
	GPIO7  Pin = 7
	GPIO8  Pin = 8
	GPIO9  Pin = 9
	GPIO10 Pin = 10
	GPIO11 Pin = 11
	GPIO12 Pin = 12
	GPIO13 Pin = 13
	GPIO14 Pin = 14
	GPIO15 Pin = 15
	GPIO16 Pin = 16
	GPIO17 Pin = 17
	GPIO18 Pin = 18
	GPIO19 Pin = 19
	GPIO20 Pin = 20
	GPIO21 Pin = 21
)

type PinChange uint8

// Pin change interrupt constants for SetInterrupt.
const (
	PinRising PinChange = iota + 1
	PinFalling
	PinToggle
)

// Configure this pin with the given configuration.
func (p Pin) Configure(config PinConfig) {
	if p == NoPin {
		// This simplifies pin configuration in peripherals such as SPI.
		return
	}

	var muxConfig uint32

	// Configure this pin as a GPIO pin.
	const function = 1 // function 1 is GPIO for every pin
	muxConfig |= function << esp.IO_MUX_GPIO_MCU_SEL_Pos

	// FUN_IE: disable for PinAnalog (high-Z for ADC)
	if config.Mode != PinAnalog {
		muxConfig |= esp.IO_MUX_GPIO_FUN_IE
	}

	// Set drive strength: 0 is lowest, 3 is highest.
	muxConfig |= 2 << esp.IO_MUX_GPIO_FUN_DRV_Pos

	// Select pull mode (no pulls for PinAnalog).
	if config.Mode == PinInputPullup {
		muxConfig |= esp.IO_MUX_GPIO_FUN_WPU
	} else if config.Mode == PinInputPulldown {
		muxConfig |= esp.IO_MUX_GPIO_FUN_WPD
	}

	// Configure the pad with the given IO mux configuration.
	p.mux().Set(muxConfig)

	// Set the output signal to the simple GPIO output.
	p.outFunc().Set(0x80)

	switch config.Mode {
	case PinOutput:
		// Set the 'output enable' bit.
		esp.GPIO.ENABLE_W1TS.Set(1 << p)
	case PinInput, PinInputPullup, PinInputPulldown, PinAnalog:
		// Clear the 'output enable' bit.
		esp.GPIO.ENABLE_W1TC.Set(1 << p)
	}
}

// configure is the same as Configure, but allows setting a specific GPIO matrix signal.
func (p Pin) configure(config PinConfig, signal uint32) {
	p.Configure(config)
	if signal == 256 {
		return
	}
	if config.Mode == PinOutput {
		p.outFunc().Set(signal)
	} else {
		inFunc(signal).Set(esp.GPIO_FUNC_IN_SEL_CFG_SEL | uint32(p))
	}
}

// outFunc returns the FUNCx_OUT_SEL_CFG register used for configuring the
// output function selection.
func (p Pin) outFunc() *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Add(unsafe.Pointer(&esp.GPIO.FUNC0_OUT_SEL_CFG), uintptr(p)*4))
}

func (p Pin) pinReg() *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Add(unsafe.Pointer(&esp.GPIO.PIN0), uintptr(p)*4))
}

// inFunc returns the FUNCy_IN_SEL_CFG register used for configuring the input
// function selection.
func inFunc(signal uint32) *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Add(unsafe.Pointer(&esp.GPIO.FUNC0_IN_SEL_CFG), uintptr(signal)*4))
}

// mux returns the I/O mux configuration register corresponding to the given
// GPIO pin.
func (p Pin) mux() *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Add(unsafe.Pointer(&esp.IO_MUX.GPIO0), uintptr(p)*4))
}

// pin returns the PIN register corresponding to the given GPIO pin.
func (p Pin) pin() *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Add(unsafe.Pointer(&esp.GPIO.PIN0), uintptr(p)*4))
}

// Set the pin to high or low.
// Warning: only use this on an output pin!
func (p Pin) Set(value bool) {
	if value {
		reg, mask := p.portMaskSet()
		reg.Set(mask)
	} else {
		reg, mask := p.portMaskClear()
		reg.Set(mask)
	}
}

// Get returns the current value of a GPIO pin when configured as an input or as
// an output.
func (p Pin) Get() bool {
	reg := &esp.GPIO.IN
	return (reg.Get()>>p)&1 > 0
}

// Return the register and mask to enable a given GPIO pin. This can be used to
// implement bit-banged drivers.
//
// Warning: only use this on an output pin!
func (p Pin) PortMaskSet() (*uint32, uint32) {
	reg, mask := p.portMaskSet()
	return &reg.Reg, mask
}

// Return the register and mask to disable a given GPIO pin. This can be used to
// implement bit-banged drivers.
//
// Warning: only use this on an output pin!
func (p Pin) PortMaskClear() (*uint32, uint32) {
	reg, mask := p.portMaskClear()
	return &reg.Reg, mask
}

func (p Pin) portMaskSet() (*volatile.Register32, uint32) {
	return &esp.GPIO.OUT_W1TS, 1 << p
}

func (p Pin) portMaskClear() (*volatile.Register32, uint32) {
	return &esp.GPIO.OUT_W1TC, 1 << p
}

// SetInterrupt sets an interrupt to be executed when a particular pin changes
// state. The pin should already be configured as an input, including a pull up
// or down if no external pull is provided.
//
// You can pass a nil func to unset the pin change interrupt. If you do so,
// the change parameter is ignored and can be set to any value (such as 0).
// If the pin is already configured with a callback, you must first unset
// this pins interrupt before you can set a new callback.
func (p Pin) SetInterrupt(change PinChange, callback func(Pin)) (err error) {
	if p >= maxPin {
		return ErrInvalidInputPin
	}

	if callback == nil {
		// Disable this pin interrupt
		p.pin().ClearBits(esp.GPIO_PIN_INT_TYPE_Msk | esp.GPIO_PIN_INT_ENA_Msk)

		if pinCallbacks[p] != nil {
			pinCallbacks[p] = nil
		}
		return nil
	}

	if pinCallbacks[p] != nil {
		// The pin was already configured.
		// To properly re-configure a pin, unset it first and set a new
		// configuration.
		return ErrNoPinChangeChannel
	}
	pinCallbacks[p] = callback

	onceSetupPinInterrupt.Do(func() {
		err = setupPinInterrupt()
	})
	if err != nil {
		return err
	}

	p.pin().Set(
		(p.pin().Get() & ^uint32(esp.GPIO_PIN_INT_TYPE_Msk|esp.GPIO_PIN_INT_ENA_Msk)) |
			uint32(change)<<esp.GPIO_PIN_INT_TYPE_Pos | uint32(1)<<esp.GPIO_PIN_INT_ENA_Pos)

	return nil
}

var (
	pinCallbacks          [maxPin]func(Pin)
	onceSetupPinInterrupt sync.Once
)

func setupPinInterrupt() error {
	esp.INTERRUPT_CORE0.GPIO_INTERRUPT_PRO_MAP.Set(cpuInterruptFromPin)
	return interrupt.New(cpuInterruptFromPin, func(interrupt.Interrupt) {
		status := esp.GPIO.STATUS.Get()
		// Clear before processing so new edges during callbacks are not lost.
		esp.GPIO.STATUS_W1TC.Set(status)
		for i, mask := 0, uint32(1); i < maxPin; i, mask = i+1, mask<<1 {
			if (status&mask) != 0 && pinCallbacks[i] != nil {
				pinCallbacks[i](Pin(i))
			}
		}
	}).Enable()
}

func init() {
	UART0.interrupt = interrupt.New(cpuInterruptFromUART0, _UART0.serveInterrupt)
	UART1.interrupt = interrupt.New(cpuInterruptFromUART1, _UART1.serveInterrupt)
}

var (
	DefaultUART = UART0

	UART0  = &_UART0
	_UART0 = UART{Bus: esp.UART0, Buffer: NewRingBuffer()}
	UART1  = &_UART1
	_UART1 = UART{Bus: esp.UART1, Buffer: NewRingBuffer()}

	errSamePins         = errors.New("UART: invalid pin combination")
	errWrongUART        = errors.New("UART: unsupported UARTn")
	errWrongBitSize     = errors.New("UART: invalid data size")
	errWrongStopBitSize = errors.New("UART: invalid bit size")
)

const uartClockFreq = 80_000_000 // APB_CLK, SCLK_SEL 1. See ESP-IDF hal/esp32c3/include/hal/uart_ll.h uart_ll_set_sclk.

type registerSet struct {
	cpuInterrupt     uint32
	interruptMapReg  *volatile.Register32
	uartClockBitMask uint32
	gpioMatrixSignal uint32
}

const (
	cpuInterruptFromUART0 = 7
	cpuInterruptFromUART1 = 8
)

func busFor(bus *esp.UART_Type) registerSet {
	switch bus {
	case esp.UART0:
		return registerSet{
			cpuInterrupt:     cpuInterruptFromUART0,
			interruptMapReg:  &esp.INTERRUPT_CORE0.UART_INTR_MAP,
			uartClockBitMask: esp.SYSTEM_PERIP_CLK_EN0_UART_CLK_EN,
			gpioMatrixSignal: 6,
		}
	case esp.UART1:
		return registerSet{
			cpuInterrupt:     cpuInterruptFromUART1,
			interruptMapReg:  &esp.INTERRUPT_CORE0.UART1_INTR_MAP,
			uartClockBitMask: esp.SYSTEM_PERIP_CLK_EN0_UART1_CLK_EN,
			gpioMatrixSignal: 9,
		}
	default:
		return registerSet{}
	}
}

func (uart *UART) configure(config UARTConfig, regs registerSet) error {
	// - disable TX/RX clock to make sure the UART transmitter or receiver is not at work during configuration
	uart.Bus.SetCLK_CONF_TX_SCLK_EN(0)
	uart.Bus.SetCLK_CONF_RX_SCLK_EN(0)

	// Configure static registers (Ref: Configuring URATn Communication)

	// - default clock source: 1=APB_CLK, 2=FOSC_CLK, 3=XTAL_CLK
	uart.Bus.SetCLK_CONF_SCLK_SEL(1)
	// reset divisor of the divider via UART_SCLK_DIV_NUM, UART_SCLK_DIV_A, and UART_SCLK_DIV_B
	uart.Bus.SetCLK_CONF_SCLK_DIV_NUM(0)
	uart.Bus.SetCLK_CONF_SCLK_DIV_A(0)
	uart.Bus.SetCLK_CONF_SCLK_DIV_B(0)

	// - the baud rate
	uart.SetBaudRate(config.BaudRate)
	// - the data format
	uart.setFormat(defaultDataBits, defaultStopBit, defaultParity)
	// - set UART mode
	uart.Bus.SetRS485_CONF_RS485_EN(0)
	uart.Bus.SetRS485_CONF_RS485TX_RX_EN(0)
	uart.Bus.SetRS485_CONF_RS485RXBY_TX_EN(0)
	uart.Bus.SetCONF0_IRDA_EN(0)
	// - disable hw-flow control
	uart.Bus.SetCONF0_TX_FLOW_EN(0)
	uart.Bus.SetCONF1_RX_FLOW_EN(0)

	// synchronize values into Core Clock
	uart.Bus.SetID_REG_UPDATE(1)

	uart.setupPins(config, regs)
	uart.configureInterrupt(regs)
	uart.enableTransmitter()
	uart.enableReceiver()

	// Start TX/RX
	uart.Bus.SetCLK_CONF_TX_SCLK_EN(1)
	uart.Bus.SetCLK_CONF_RX_SCLK_EN(1)
	return nil
}

// uartSync does nothing. The C3 configuration registers do not need a sync.
// See ESP-IDF hal/esp32c3/include/hal/uart_ll.h uart_ll_set_data_bit_num.
func uartSync(bus *esp.UART_Type) {}

func (uart *UART) setFormat(dataBits, stopBits int, parity UARTParity) {
	// - data length
	uart.Bus.SetCONF0_BIT_NUM(uint32(dataBits - 5))
	// - stop bit
	uart.Bus.SetCONF0_STOP_BIT_NUM(uint32(stopBits))
	// - parity check
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
	uartClock := &esp.SYSTEM.PERIP_CLK_EN0
	uartClockReset := &esp.SYSTEM.PERIP_RST_EN0

	// Initialize/reset URATn (Ref: Initializing URATn)
	// - enable the clock for UART RAM
	uartClock.SetBits(esp.SYSTEM_PERIP_CLK_EN0_UART_MEM_CLK_EN)
	// - enable APB_CLK for UARTn
	uartClock.SetBits(regs.uartClockBitMask)
	// - reset sequence
	uartClockReset.ClearBits(regs.uartClockBitMask)
	bus.SetCLK_CONF_RST_CORE(1)
	uartClockReset.SetBits(regs.uartClockBitMask)
	uartClockReset.ClearBits(regs.uartClockBitMask)
	bus.SetCLK_CONF_RST_CORE(0)
	// synchronize core register
	bus.SetID_REG_UPDATE(0)
	// enable RTC clock
	esp.RTC_CNTL.SetCLK_CONF_DIG_CLK8M_EN(1)
	// wait for Core Clock to ready for configuration
	for bus.GetID_REG_UPDATE() > 0 {
		riscv.Asm("nop")
	}
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

// GetRNG returns 32-bit random numbers using the ESP32-C3 true random number generator,
// Random numbers are generated based on the thermal noise in the system and the
// asynchronous clock mismatch.
// For maximum entropy also make sure that the SAR_ADC is enabled.
// See esp32-c3_technical_reference_manual_en.pdf p.524
func GetRNG() (ret uint32, err error) {
	// ensure ADC clock is initialized
	initADCClock()

	// ensure fast RTC clock is enabled
	if esp.RTC_CNTL.GetCLK_CONF_DIG_CLK8M_EN() == 0 {
		esp.RTC_CNTL.SetCLK_CONF_DIG_CLK8M_EN(1)
	}

	return esp.APB_CTRL.GetRND_DATA(), nil
}

func initADCClock() {
	if esp.APB_SARADC.GetCLKM_CONF_CLK_EN() == 1 {
		return
	}

	// only support ADC_CTRL_CLK set to 1
	esp.APB_SARADC.SetCLKM_CONF_CLK_SEL(1)

	esp.APB_SARADC.SetCTRL_SARADC_SAR_CLK_GATED(1)

	esp.APB_SARADC.SetCLKM_CONF_CLKM_DIV_NUM(15)
	esp.APB_SARADC.SetCLKM_CONF_CLKM_DIV_B(1)
	esp.APB_SARADC.SetCLKM_CONF_CLKM_DIV_A(0)

	esp.APB_SARADC.SetCTRL_SARADC_SAR_CLK_DIV(1)
	esp.APB_SARADC.SetCLKM_CONF_CLK_EN(1)
}
