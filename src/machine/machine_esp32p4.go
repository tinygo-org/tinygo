//go:build esp32p4

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

// maxPin is the number of GPIO pins (0-54 = 55 pins).
// Pins > 31 require the secondary register set (OUT1_W1TS, ENABLE1_W1TS, IN1).
const maxPin = 55

const cpuInterruptFromPin = 6

// CPUFrequency returns the current CPU frequency of the chip.
func CPUFrequency() uint32 {
	return 360e6 // 360MHz (default HP CPU clock)
}

const (
	PinOutput PinMode = iota
	PinInput
	PinInputPullup
	PinInputPulldown
	PinAnalog
)

// Hardware pin numbers for ESP32-P4.
const (
	GPIO0  Pin = 0
	GPIO1  Pin = 1
	GPIO2  Pin = 2
	GPIO3  Pin = 3
	GPIO4  Pin = 4
	GPIO5  Pin = 5
	GPIO6  Pin = 6
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
	GPIO22 Pin = 22
	GPIO23 Pin = 23
	GPIO24 Pin = 24
	GPIO25 Pin = 25
	GPIO26 Pin = 26
	GPIO27 Pin = 27
	GPIO28 Pin = 28
	GPIO29 Pin = 29
	GPIO30 Pin = 30
	GPIO31 Pin = 31
	GPIO32 Pin = 32
	GPIO33 Pin = 33
	GPIO34 Pin = 34
	GPIO35 Pin = 35
	GPIO36 Pin = 36
	GPIO37 Pin = 37
	GPIO38 Pin = 38
	GPIO39 Pin = 39
	GPIO40 Pin = 40
	GPIO41 Pin = 41
	GPIO42 Pin = 42
	GPIO43 Pin = 43
	GPIO44 Pin = 44
	GPIO45 Pin = 45
	GPIO46 Pin = 46
	GPIO47 Pin = 47
	GPIO48 Pin = 48
	GPIO49 Pin = 49
	GPIO50 Pin = 50
	GPIO51 Pin = 51
	GPIO52 Pin = 52
	GPIO53 Pin = 53
	GPIO54 Pin = 54
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
	// Output function 256 is SIG_GPIO_OUT_IDX, the simple GPIO output.
	p.configure(config, 256)
}

// configure is the same as Configure, but allows setting a specific GPIO matrix signal.
func (p Pin) configure(config PinConfig, signal uint32) {
	if p == NoPin {
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

	switch config.Mode {
	case PinOutput:
		// Set the 'output enable' bit.
		if p < 32 {
			esp.GPIO.ENABLE_W1TS.Set(1 << p)
		} else {
			esp.GPIO.ENABLE1_W1TS.Set(1 << (p - 32))
		}
		// Set the output signal.
		p.outFunc().Set(signal)
	case PinInput, PinInputPullup, PinInputPulldown, PinAnalog:
		// Clear the 'output enable' bit.
		if p < 32 {
			esp.GPIO.ENABLE_W1TC.Set(1 << p)
		} else {
			esp.GPIO.ENABLE1_W1TC.Set(1 << (p - 32))
		}
		if signal != 256 && config.Mode != PinAnalog {
			inFunc(signal).Set(esp.GPIO_FUNC_IN_SEL_CFG_SEL | uint32(p))
		}
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
// function selection. On the ESP32-P4, GPIO_Type starts at FUNC1_IN_SEL_CFG
// (signal 0 is reserved), so we offset from FUNC1.
func inFunc(signal uint32) *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Add(unsafe.Pointer(&esp.GPIO.FUNC1_IN_SEL_CFG), uintptr(signal-1)*4))
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
	if p < 32 {
		return esp.GPIO.IN.Get()&(1<<p) != 0
	} else {
		return esp.GPIO.IN1.Get()&(1<<(p-32)) != 0
	}
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
	if p < 32 {
		return &esp.GPIO.OUT_W1TS, 1 << p
	} else {
		return &esp.GPIO.OUT1_W1TS, 1 << (p - 32)
	}
}

func (p Pin) portMaskClear() (*volatile.Register32, uint32) {
	if p < 32 {
		return &esp.GPIO.OUT_W1TC, 1 << p
	} else {
		return &esp.GPIO.OUT1_W1TC, 1 << (p - 32)
	}
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
	esp.INTERRUPT_CORE0.GPIO_INT0_MAP.Set(cpuInterruptFromPin)
	return interrupt.New(cpuInterruptFromPin, func(interrupt.Interrupt) {
		status := esp.GPIO.STATUS.Get()
		status1 := esp.GPIO.STATUS1.Get()
		esp.GPIO.STATUS_W1TC.Set(status)
		esp.GPIO.STATUS1_W1TC.Set(status1)

		for i, mask := 0, uint32(1); i < 32; i, mask = i+1, mask<<1 {
			if (status&mask) != 0 && pinCallbacks[i] != nil {
				pinCallbacks[i](Pin(i))
			}
		}
		for i, mask := 32, uint32(1); i < maxPin; i, mask = i+1, mask<<1 {
			if (status1&mask) != 0 && pinCallbacks[i] != nil {
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

// GPIO matrix signal indices for UART peripherals on the ESP32-P4.
// Source: ESP-IDF components/soc/esp32p4/include/soc/gpio_sig_map.h
const (
	uartSignalU0 = uint32(10) // UART0_TXD_PAD_OUT_IDX / UART0_RXD_PAD_IN_IDX
	uartSignalU1 = uint32(13) // UART1_TXD_PAD_OUT_IDX / UART1_RXD_PAD_IN_IDX
)

const uartClockFreq = 40_000_000 // XTAL_CLK, CLK_SRC_SEL 0. See ESP-IDF hal/esp32p4/include/hal/uart_ll.h uart_ll_set_sclk.

type registerSet struct {
	cpuInterrupt     uint32
	interruptMapReg  *volatile.Register32
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
			interruptMapReg:  &esp.INTERRUPT_CORE0.UART0_INT_MAP,
			gpioMatrixSignal: uartSignalU0,
		}
	case esp.UART1:
		return registerSet{
			cpuInterrupt:     cpuInterruptFromUART1,
			interruptMapReg:  &esp.INTERRUPT_CORE0.UART1_INT_MAP,
			gpioMatrixSignal: uartSignalU1,
		}
	default:
		return registerSet{}
	}
}

func (uart *UART) configure(config UARTConfig, regs registerSet) error {
	// Disable TX/RX clock during configuration.
	uart.Bus.SetCLK_CONF_TX_SCLK_EN(0)
	uart.Bus.SetCLK_CONF_RX_SCLK_EN(0)

	uart.SetBaudRate(config.BaudRate)
	uart.setFormat(defaultDataBits, defaultStopBit, defaultParity)

	// Set UART mode (disable RS485 and IrDA).
	uart.Bus.SetRS485_CONF_SYNC_RS485_EN(0)
	uart.Bus.SetRS485_CONF_SYNC_RS485TX_RX_EN(0)
	uart.Bus.SetRS485_CONF_SYNC_RS485RXBY_TX_EN(0)
	uart.Bus.SetCONF0_SYNC_IRDA_EN(0)
	// Disable hw-flow control.
	uart.Bus.SetCONF0_SYNC_TX_FLOW_EN(0)

	// Synchronize values into Core Clock.
	uart.Bus.SetREG_UPDATE(1)

	uart.setupPins(config, regs)
	uart.configureInterrupt(regs)
	uart.enableTransmitter()
	uart.enableReceiver()

	// Synchronize CONF0_SYNC/CONF1 writes from enableTransmitter/enableReceiver.
	uart.Bus.SetREG_UPDATE(1)

	// Start TX/RX.
	uart.Bus.SetCLK_CONF_TX_SCLK_EN(1)
	uart.Bus.SetCLK_CONF_RX_SCLK_EN(1)
	return nil
}

// uartSync copies the configuration registers into the UART core clock domain.
// See ESP-IDF hal/esp32p4/include/hal/uart_ll.h uart_ll_update.
func uartSync(bus *esp.UART_Type) {
	bus.SetREG_UPDATE(1)
	for bus.GetREG_UPDATE() != 0 {
	}
}

func (uart *UART) setFormat(dataBits, stopBits int, parity UARTParity) {
	uart.Bus.SetCONF0_SYNC_BIT_NUM(uint32(dataBits - 5))
	uart.Bus.SetCONF0_SYNC_STOP_BIT_NUM(uint32(stopBits))
	switch parity {
	case ParityNone:
		uart.Bus.SetCONF0_SYNC_PARITY_EN(0)
	case ParityEven:
		uart.Bus.SetCONF0_SYNC_PARITY_EN(1)
		uart.Bus.SetCONF0_SYNC_PARITY(0)
	case ParityOdd:
		uart.Bus.SetCONF0_SYNC_PARITY_EN(1)
		uart.Bus.SetCONF0_SYNC_PARITY(1)
	}
}

// initUARTClock enables and configures the UART clock via HP_SYS_CLKRST.
// On ESP32-P4, UART clock control is in HP_SYS_CLKRST (not PCR like ESP32-C6).
func initUARTClock(bus *esp.UART_Type, regs registerSet) {
	switch bus {
	case esp.UART0:
		// Enable bus clocks for UART0 (matches uart_ll_enable_bus_clock):
		//   soc_clk_ctrl2.uart0_apb_clk_en = APB bus access clock
		//   soc_clk_ctrl1.uart0_sys_clk_en = system clock (required for register access)
		esp.HP_SYS_CLKRST.SetSOC_CLK_CTRL2_REG_UART0_APB_CLK_EN(1)
		esp.HP_SYS_CLKRST.SetSOC_CLK_CTRL1_REG_UART0_SYS_CLK_EN(1)
		// Select XTAL (40MHz) as clock source — guaranteed available after ROM boot.
		// PLL_F80M requires the main SPLL to be running, which the ROM does not
		// guarantee on POWERON reset.
		// SEL encoding: 0=XTAL(40MHz), 1=RC_FAST(17.5MHz), 2=PLL_F80M(80MHz).
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL110_REG_UART0_CLK_SRC_SEL(0)
		// Enable UART0 peripheral clock.
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL110_REG_UART0_CLK_EN(1)
		// Reset SCLK divisors.
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL111_REG_UART0_SCLK_DIV_NUM(0)
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL111_REG_UART0_SCLK_DIV_NUMERATOR(0)
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL111_REG_UART0_SCLK_DIV_DENOMINATOR(0)
		// Reset UART core.
		bus.SetCLK_CONF_TX_RST_CORE(1)
		bus.SetCLK_CONF_RX_RST_CORE(1)
		riscv.Asm("nop")
		bus.SetCLK_CONF_TX_RST_CORE(0)
		bus.SetCLK_CONF_RX_RST_CORE(0)
	case esp.UART1:
		esp.HP_SYS_CLKRST.SetSOC_CLK_CTRL2_REG_UART1_APB_CLK_EN(1)
		esp.HP_SYS_CLKRST.SetSOC_CLK_CTRL1_REG_UART1_SYS_CLK_EN(1)
		// SEL encoding: 0=XTAL(40MHz), 1=RC_FAST(17.5MHz), 2=PLL_F80M(80MHz).
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL111_REG_UART1_CLK_SRC_SEL(0)
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL111_REG_UART1_CLK_EN(1)
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL112_REG_UART1_SCLK_DIV_NUM(0)
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL112_REG_UART1_SCLK_DIV_NUMERATOR(0)
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL112_REG_UART1_SCLK_DIV_DENOMINATOR(0)
		bus.SetCLK_CONF_TX_RST_CORE(1)
		bus.SetCLK_CONF_RX_RST_CORE(1)
		riscv.Asm("nop")
		bus.SetCLK_CONF_TX_RST_CORE(0)
		bus.SetCLK_CONF_RX_RST_CORE(0)
	}
	// Wait for Core Clock to be ready for configuration.
	for bus.GetREG_UPDATE() > 0 {
		riscv.Asm("nop")
	}
}

func (uart *UART) SetBaudRate(baudRate uint32) {
	clk_div, sclk_div := uartDividers(baudRate)
	uart.Bus.SetCLKDIV_SYNC_CLKDIV(clk_div >> 4)
	uart.Bus.SetCLKDIV_SYNC_CLKDIV_FRAG(clk_div & 0xf)
	// Set the SCLK divisor in HP_SYS_CLKRST.
	switch uart.Bus {
	case esp.UART0:
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL111_REG_UART0_SCLK_DIV_NUM(sclk_div - 1)
	case esp.UART1:
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL112_REG_UART1_SCLK_DIV_NUM(sclk_div - 1)
	}
}

func uartEnableTX(pin Pin) {
	if pin < 32 {
		esp.GPIO.ENABLE_W1TS.Set(1 << pin)
	} else {
		esp.GPIO.ENABLE1_W1TS.Set(1 << (pin - 32))
	}
}

func uartSetTXFIFORST(bus *esp.UART_Type, value uint32) {
	bus.SetCONF0_SYNC_TXFIFO_RST(value)
}

func uartSetRXFIFORST(bus *esp.UART_Type, value uint32) {
	bus.SetCONF0_SYNC_RXFIFO_RST(value)
}
