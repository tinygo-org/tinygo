//go:build rp2350

package machine

import (
	"device/rp"
	"unsafe"
)

const (
	LED             = GPIO25
	_NUMBANK0_GPIOS = 48
	_NUMBANK0_IRQS  = 6
	xoscFreq        = 12 // Pico 2 Crystal oscillator Abracon ABM8-272-T3 frequency in MHz
	rp2350ExtraReg  = 1
	notimpl         = "rp2350: not implemented"
	initUnreset     = rp.RESETS_RESET_ADC |
		rp.RESETS_RESET_SPI0 |
		rp.RESETS_RESET_SPI1 |
		rp.RESETS_RESET_UART0 |
		rp.RESETS_RESET_UART1 |
		rp.RESETS_RESET_USBCTRL
)

const (
	PinOutput PinMode = iota
	PinInput
	PinInputPulldown
	PinInputPullup
	PinAnalog
	PinUART
	PinPWM
	PinI2C
	PinSPI
	PinPIO0
	PinPIO1
	PinPIO2
)

// GPIO function selectors
const (
	// Connect the high-speed transmit peripheral (HSTX) to GPIO.
	fnHSTX       pinFunc = 0
	fnSPI        pinFunc = 1 // Connect one of the internal PL022 SPI peripherals to GPIO
	fnUARTctsrts pinFunc = 2
	fnI2C        pinFunc = 3
	// Connect a PWM slice to GPIO. There are eight PWM slices,
	// each with two outputchannels (A/B). The B pin can also be used as an input,
	// for frequency and duty cyclemeasurement
	fnPWM pinFunc = 4
	// Software control of GPIO, from the single-cycle IO (SIO) block.
	// The SIO function (F5)must be selected for the processors to drive a GPIO,
	// but the input is always connected,so software can check the state of GPIOs at any time.
	fnSIO pinFunc = 5
	// Connect one of the programmable IO blocks (PIO) to GPIO. PIO can implement a widevariety of interfaces,
	// and has its own internal pin mapping hardware, allowing flexibleplacement of digital interfaces on bank 0 GPIOs.
	// The PIO function (F6, F7, F8) must beselected for PIO to drive a GPIO, but the input is always connected,
	// so the PIOs canalways see the state of all pins.
	fnPIO0, fnPIO1, fnPIO2 pinFunc = 6, 7, 8
	// General purpose clock outputs. Can drive a number of internal clocks (including PLL
	// 	outputs) onto GPIOs, with optional integer divide.
	fnGPCK pinFunc = 9
	// QSPI memory interface peripheral, used for execute-in-place from external QSPI flash or PSRAM memory devices.
	fnQMI pinFunc = 9
	// USB power control signals to/from the internal USB controller.
	fnUSB  pinFunc = 10
	fnUART pinFunc = 11
	fnNULL pinFunc = 0x1f
)

// Configure configures the gpio pin as per mode.
func (p Pin) Configure(config PinConfig) {
	if p == NoPin {
		return
	}
	p.init()
	mask := uint32(1) << p
	switch config.Mode {
	case PinOutput:
		p.setFunc(fnSIO)
		rp.SIO.GPIO_OE_SET.Set(mask)
	case PinInput:
		p.setFunc(fnSIO)
		p.pulloff()
	case PinInputPulldown:
		p.setFunc(fnSIO)
		p.pulldown()
	case PinInputPullup:
		p.setFunc(fnSIO)
		p.pullup()
	case PinAnalog:
		p.setFunc(fnNULL)
		p.pulloff()
	case PinUART:
		p.setFunc(fnUART)
	case PinPWM:
		p.setFunc(fnPWM)
	case PinI2C:
		// IO config according to 4.3.1.3 of rp2040 datasheet.
		p.setFunc(fnI2C)
		p.pullup()
		p.setSchmitt(true)
		p.setSlew(false)
	case PinSPI:
		p.setFunc(fnSPI)
	case PinPIO0:
		p.setFunc(fnPIO0)
	case PinPIO1:
		p.setFunc(fnPIO1)
	case PinPIO2:
		p.setFunc(fnPIO2)
	}
}

var (
	timer = (*timerType)(unsafe.Pointer(rp.TIMER0))
)

// Enable or disable a specific interrupt on the executing core.
// num is the interrupt number which must be in [0,31].
func irqSet(num uint32, enabled bool) {
	if num >= _NUMIRQ {
		return
	}
	irqSetMask(1<<num, enabled)
}

func irqSetMask(mask uint32, enabled bool) {
	panic(notimpl)
}

func (clks *clocksType) initRTC() {} // No RTC on RP2350.

// startTick starts the watchdog tick.
// On RP2040, the watchdog contained a tick generator used to generate a 1μs tick for the watchdog. This was also
// distributed to the system timer. On RP2350, the watchdog instead takes a tick input from the system-level ticks block. See Section 8.5.
func (wd *watchdogImpl) startTick(cycles uint32) {
	rp.TICKS.WATCHDOG_CTRL.SetBits(1)
}
