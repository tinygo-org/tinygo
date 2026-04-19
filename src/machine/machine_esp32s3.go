//go:build esp32s3

package machine

import (
	"device/esp"
	"errors"
	"runtime/interrupt"
	"runtime/volatile"
	"sync"
	"unsafe"
)

const deviceName = esp.Device

const xtalClock = 40_000000       // 40MHz
const apbClock = 80_000000        // 80MHz
const cryptoPWMClock = 160_000000 // 160MHz

// GetCPUFrequency returns the current CPU frequency of the chip.
func GetCPUFrequency() (uint32, error) {
	switch esp.SYSTEM.GetSYSCLK_CONF_SOC_CLK_SEL() {
	case 0:
		return xtalClock / (esp.SYSTEM.GetSYSCLK_CONF_PRE_DIV_CNT() + 1), nil
	case 1:
		switch esp.SYSTEM.GetCPU_PER_CONF_CPUPERIOD_SEL() {
		case 0:
			return 80e6, nil
		case 1:
			return 160e6, nil
		case 2:
			// If esp.SYSTEM.GetCPU_PER_CONF_PLL_FREQ_SEL() == 1, this is undefined
			return 240e6, nil
		}
	case 2:
		//RC Fast Clock
		return (175e5) / (esp.SYSTEM.GetSYSCLK_CONF_PRE_DIV_CNT() + 1), nil
	}
	return 0, errors.New("machine: Unable to determine current cpu frequency")
}

// SetCPUFrequency sets the frequency of the CPU to one of several targets
func SetCPUFrequency(frequency uint32) error {
	// Always assume we are on PLL. Lower frequencies can be set with a different
	// clock source, but this will change the behavior of APB clock and Crypto PWM
	// clock
	//esp.SYSTEM.SetSYSCLK_CONF_SOC_CLK_SEL(1)

	switch frequency {
	case 80_000000:
		esp.SYSTEM.SetCPU_PER_CONF_CPUPERIOD_SEL(0)
		esp.SYSTEM.SetCPU_PER_CONF_PLL_FREQ_SEL(0) // Reduce PLL freq when possible
		return nil
	case 160_000000:
		esp.SYSTEM.SetCPU_PER_CONF_CPUPERIOD_SEL(1)
		esp.SYSTEM.SetCPU_PER_CONF_PLL_FREQ_SEL(0)
		return nil
	case 240_000000:
		esp.SYSTEM.SetCPU_PER_CONF_PLL_FREQ_SEL(1) // Increase PLL freq when needed
		esp.SYSTEM.SetCPU_PER_CONF_CPUPERIOD_SEL(2)
		return nil
	}
	return errors.New("machine: Unsupported CPU frequency selected. Supported: 80, 160, 240 MHz")
}

var (
	ErrInvalidSPIBus = errors.New("machine: invalid SPI bus")
)

const (
	PinOutput PinMode = iota
	PinInput
	PinInputPullup
	PinInputPulldown
	PinAnalog
)

// Hardware pin numbers
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
)

const (
	ADC0  Pin = GPIO1
	ADC2  Pin = GPIO2
	ADC3  Pin = GPIO3
	ADC4  Pin = GPIO4
	ADC5  Pin = GPIO5
	ADC6  Pin = GPIO6
	ADC7  Pin = GPIO7
	ADC8  Pin = GPIO8
	ADC9  Pin = GPIO9
	ADC10 Pin = GPIO10
	ADC11 Pin = GPIO11
	ADC12 Pin = GPIO12
	ADC13 Pin = GPIO13
	ADC14 Pin = GPIO14
	ADC15 Pin = GPIO15
	ADC16 Pin = GPIO16
	ADC17 Pin = GPIO17
	ADC18 Pin = GPIO18
	ADC19 Pin = GPIO19
	ADC20 Pin = GPIO20
)

// Configure this pin with the given configuration.
func (p Pin) Configure(config PinConfig) {
	// Output function 256 is a special value reserved for use as a regular GPIO
	// pin. Peripherals (SPI etc) can set a custom output function by calling
	// lowercase configure() instead with a signal name.
	p.configure(config, 256)
}

// configure is the same as Configure, but allows for setting a specific input
// or output signal.
// Signals are always routed through the GPIO matrix for simplicity. Output
// signals are configured in FUNCx_OUT_SEL_CFG which selects a particular signal
// to output on a given pin. Input signals are configured in FUNCy_IN_SEL_CFG,
// which sets the pin to use for a particular input signal.
func (p Pin) configure(config PinConfig, signal uint32) {
	if p == NoPin {
		// This simplifies pin configuration in peripherals such as SPI.
		return
	}

	ioConfig := uint32(0)

	// MCU_SEL: Function 1 is always GPIO
	ioConfig |= (1 << esp.IO_MUX_GPIO_MCU_SEL_Pos)

	// FUN_IE: disable for PinAnalog (high-Z for ADC), enable for digital
	if config.Mode != PinAnalog {
		ioConfig |= esp.IO_MUX_GPIO_FUN_IE
	}

	// DRV: Set drive strength to 20 mA as a default. Pins 17 and 18 are special
	var drive uint32
	if p == GPIO17 || p == GPIO18 {
		drive = 1 // 20 mA
	} else {
		drive = 2 // 20 mA
	}
	ioConfig |= (drive << esp.IO_MUX_GPIO_FUN_DRV_Pos)

	// WPU/WPD: no pulls for PinAnalog
	if config.Mode == PinInputPullup {
		ioConfig |= esp.IO_MUX_GPIO_FUN_WPU
	} else if config.Mode == PinInputPulldown {
		ioConfig |= esp.IO_MUX_GPIO_FUN_WPD
	}

	// Set configuration
	ioRegister := p.ioMuxReg()
	ioRegister.Set(ioConfig)

	switch config.Mode {
	case PinOutput:
		// Set the 'output enable' bit.
		if p < 32 {
			esp.GPIO.ENABLE_W1TS.Set(1 << p)
		} else {
			esp.GPIO.ENABLE1_W1TS.Set(1 << (p - 32))
		}
		// Set the signal to read the output value from. It can be a peripheral
		// output signal, or the special value 256 which indicates regular GPIO
		// usage.
		p.outFunc().Set(signal)
	case PinInput, PinInputPullup, PinInputPulldown, PinAnalog:
		// Clear the 'output enable' bit.
		if p < 32 {
			esp.GPIO.ENABLE_W1TC.Set(1 << p)
		} else {
			esp.GPIO.ENABLE1_W1TC.Set(1 << (p - 32))
		}
		if signal != 256 && config.Mode != PinAnalog {
			// Signal is a peripheral function (not a simple GPIO). Connect this
			// signal to the pin.
			// Note that outFunc and inFunc work in the opposite direction.
			// outFunc configures a pin to use a given output signal, while
			// inFunc specifies a pin to use to read the signal from.
			inFunc(signal).Set(esp.GPIO_FUNC_IN_SEL_CFG_SEL | uint32(p)<<esp.GPIO_FUNC_IN_SEL_CFG_IN_SEL_Pos)
		}
	}
}

// ioMuxReg returns the IO_MUX_n_REG register used for configuring the io mux for
// this pin
func (p Pin) ioMuxReg() *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Add(unsafe.Pointer(&esp.IO_MUX.GPIO0), uintptr(p)*4))
}

// outFunc returns the FUNCx_OUT_SEL_CFG register used for configuring the
// output function selection.
func (p Pin) outFunc() *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Add(unsafe.Pointer(&esp.GPIO.FUNC0_OUT_SEL_CFG), uintptr(p)*4))
}

// inFunc returns the FUNCy_IN_SEL_CFG register used for configuring the input
// function selection.
func inFunc(signal uint32) *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Add(unsafe.Pointer(&esp.GPIO.FUNC0_IN_SEL_CFG), uintptr(signal)*4))
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

// Get returns the current value of a GPIO pin when the pin is configured as an
// input or as an output.
func (p Pin) Get() bool {
	if p < 32 {
		return esp.GPIO.IN.Get()&(1<<p) != 0
	} else {
		return esp.GPIO.IN1.Get()&(1<<(p-32)) != 0
	}
}

func (p Pin) pinReg() *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Add(unsafe.Pointer(&esp.GPIO.PIN0), uintptr(p)*4))
}

const maxPin = 49

// cpuInterruptFromPin selects an edge-triggered CPU interrupt line for GPIO.
// CPU interrupt 10 is edge-triggered level-1 on the Xtensa LX7, which prevents
// the ISR from re-entering continuously when other peripherals (e.g. SPI via
// the GPIO Matrix) keep GPIO.STATUS bits asserted.
const cpuInterruptFromPin = 10

type PinChange uint8

// Pin change interrupt constants for SetInterrupt.
const (
	PinRising PinChange = iota + 1
	PinFalling
	PinToggle
)

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
		p.pinReg().ClearBits(esp.GPIO_PIN_INT_TYPE_Msk | esp.GPIO_PIN_INT_ENA_Msk)

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

	p.pinReg().Set(
		(p.pinReg().Get() & ^uint32(esp.GPIO_PIN_INT_TYPE_Msk|esp.GPIO_PIN_INT_ENA_Msk)) |
			uint32(change)<<esp.GPIO_PIN_INT_TYPE_Pos | uint32(1)<<esp.GPIO_PIN_INT_ENA_Pos)

	return nil
}

var (
	pinCallbacks          [maxPin]func(Pin)
	onceSetupPinInterrupt sync.Once
)

func setupPinInterrupt() error {
	esp.INTERRUPT_CORE0.SetGPIO_INTERRUPT_PRO_MAP(cpuInterruptFromPin)
	return interrupt.New(cpuInterruptFromPin, func(interrupt.Interrupt) {
		// Read and immediately clear interrupt status bits.
		// Clearing before processing is critical for edge-triggered CPU
		// interrupts: any new GPIO events that arrive during callback
		// execution will set fresh STATUS bits, generating a new edge
		// on the CPU interrupt line so they are not lost.
		status := esp.GPIO.STATUS.Get()
		status1 := esp.GPIO.STATUS1.Get()
		esp.GPIO.STATUS_W1TC.Set(status)
		esp.GPIO.STATUS1_W1TC.Set(status1)

		// Check status for GPIO0-31
		for i, mask := 0, uint32(1); i < 32; i, mask = i+1, mask<<1 {
			if (status&mask) != 0 && pinCallbacks[i] != nil {
				pinCallbacks[i](Pin(i))
			}
		}
		// Check status for GPIO32-48
		for i, mask := 32, uint32(1); i < maxPin; i, mask = i+1, mask<<1 {
			if (status1&mask) != 0 && pinCallbacks[i] != nil {
				pinCallbacks[i](Pin(i))
			}
		}
	}).Enable()
}

var DefaultUART = UART0

var (
	UART0  = &_UART0
	_UART0 = UART{Bus: esp.UART0, Buffer: NewRingBuffer()}
	UART1  = &_UART1
	_UART1 = UART{Bus: esp.UART1, Buffer: NewRingBuffer()}
	UART2  = &_UART2
	_UART2 = UART{Bus: esp.UART2, Buffer: NewRingBuffer()}
)

type UART struct {
	Bus    *esp.UART_Type
	Buffer *RingBuffer
}

func (uart *UART) Configure(config UARTConfig) {
	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}
	// Crystal clock source is selected by default
	uart.Bus.CLKDIV.Set(xtalClock / config.BaudRate)
}

func (uart *UART) writeByte(b byte) error {
	for (uart.Bus.STATUS.Get()>>16)&0xff >= 128 {
		// Read UART_TXFIFO_CNT from the status register, which indicates how
		// many bytes there are in the transmit buffer. Wait until there are
		// less than 128 bytes in this buffer (the default buffer size).
	}
	uart.Bus.FIFO.Set(uint32(b))
	return nil
}

func (uart *UART) flush() {}

// GetRNG returns 32-bit random numbers using the ESP32-S3 true random number generator,
// Random numbers are generated based on the thermal noise in the system and the
// asynchronous clock mismatch.
// For maximum entropy also make sure that the SAR_ADC is enabled.
// See esp32-s3_technical_reference_manual_en.pdf p.920
func GetRNG() (ret uint32, err error) {
	// ensure ADC clock is initialized
	initADCClock()

	// ensure fast RTC clock is enabled
	if esp.RTC_CNTL.GetCLK_CONF_DIG_CLK8M_EN() == 0 {
		esp.RTC_CNTL.SetCLK_CONF_DIG_CLK8M_EN(1)
	}

	return esp.RNG.DATA.Get(), nil
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
