//go:build rp2350

package machine

import (
	"device/rp"
	"runtime/volatile"
	"unsafe"
)

const (
	cpuFreq          = 150 * MHz
	_NUMBANK0_GPIOS  = 48
	_NUMBANK0_IRQS   = 6
	rp2350ExtraReg   = 1
	_NUMIRQ          = 51
	notimpl          = "rp2350: not implemented"
	RESETS_RESET_Msk = 0x1fffffff
	initUnreset      = rp.RESETS_RESET_ADC |
		rp.RESETS_RESET_SPI0 |
		rp.RESETS_RESET_SPI1 |
		rp.RESETS_RESET_UART0 |
		rp.RESETS_RESET_UART1 |
		rp.RESETS_RESET_USBCTRL
	initDontReset = rp.RESETS_RESET_USBCTRL |
		rp.RESETS_RESET_SYSCFG |
		rp.RESETS_RESET_PLL_USB |
		rp.RESETS_RESET_PLL_SYS |
		rp.RESETS_RESET_PADS_QSPI |
		rp.RESETS_RESET_IO_QSPI |
		rp.RESETS_RESET_JTAG
	padEnableMask = rp.PADS_BANK0_GPIO0_IE_Msk |
		rp.PADS_BANK0_GPIO0_OD_Msk |
		rp.PADS_BANK0_GPIO0_ISO_Msk
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

const (
	clkGPOUT0 clockIndex = iota // GPIO Muxing 0
	clkGPOUT1                   // GPIO Muxing 1
	clkGPOUT2                   // GPIO Muxing 2
	clkGPOUT3                   // GPIO Muxing 3
	clkRef                      // Watchdog and timers reference clock
	clkSys                      // Processors, bus fabric, memory, memory mapped registers
	clkPeri                     // Peripheral clock for UART and SPI
	ClkHSTX                     // High speed interface
	clkUSB                      // USB clock
	clkADC                      // ADC clock
	numClocks
)

func calcClockDiv(srcFreq, freq uint32) uint32 {
	// Div register is 4.16 int.frac divider so multiply by 2^16 (left shift by 16)
	return uint32((uint64(srcFreq) << 16) / uint64(freq))
}

type clocksType struct {
	clk               [numClocks]clockType
	dftclk_xosc_ctrl  volatile.Register32
	dftclk_rosc_ctrl  volatile.Register32
	dftclk_lposc_ctrl volatile.Register32
	resus             struct {
		ctrl   volatile.Register32
		status volatile.Register32
	}
	fc0      fc
	wakeEN0  volatile.Register32
	wakeEN1  volatile.Register32
	sleepEN0 volatile.Register32
	sleepEN1 volatile.Register32
	enabled0 volatile.Register32
	enabled1 volatile.Register32
	intR     volatile.Register32
	intE     volatile.Register32
	intF     volatile.Register32
	intS     volatile.Register32
}

// GPIO function selectors
const (
	// Connect the high-speed transmit peripheral (HSTX) to GPIO.
	fnHSTX pinFunc = 0
	fnSPI  pinFunc = 1 // Connect one of the internal PL022 SPI peripherals to GPIO
	fnUART pinFunc = 2
	fnI2C  pinFunc = 3
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
	fnUSB     pinFunc = 10
	fnUARTAlt pinFunc = 11
	fnNULL    pinFunc = 0x1f
)

// Configure configures the gpio pin as per mode.
func (p Pin) Configure(config PinConfig) {
	if p == NoPin {
		return
	}
	p.init()

	switch config.Mode {
	case PinOutput:
		p.setFunc(fnSIO)
		if is48Pin && p >= 32 {
			mask := uint32(1) << (p % 32)
			rp.SIO.GPIO_HI_OE_SET.Set(mask)
		} else {
			mask := uint32(1) << p
			rp.SIO.GPIO_OE_SET.Set(mask)
		}
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
		// Disable digital input.
		p.padCtrl().ClearBits(rp.PADS_BANK0_GPIO0_IE)
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
// num is the interrupt number which must be in [0,_NUMIRQ).
func irqSet(num uint32, enabled bool) {
	if num >= _NUMIRQ {
		return
	}

	register_index := num / 32
	var mask uint32 = 1 << (num % 32)

	if enabled {
		// Clear pending before enable
		//(if IRQ is actually asserted, it will immediately re-pend)
		if register_index == 0 {
			rp.PPB.NVIC_ICPR0.Set(mask)
			rp.PPB.NVIC_ISER0.Set(mask)
		} else {
			rp.PPB.NVIC_ICPR1.Set(mask)
			rp.PPB.NVIC_ISER1.Set(mask)
		}
	} else {
		if register_index == 0 {
			rp.PPB.NVIC_ICER0.Set(mask)
		} else {
			rp.PPB.NVIC_ICER1.Set(mask)
		}
	}
}

func (clks *clocksType) initRTC() {} // No RTC on RP2350.

func (clks *clocksType) initTicks() {
	rp.TICKS.SetTIMER0_CTRL_ENABLE(0)
	rp.TICKS.SetTIMER0_CYCLES(12)
	rp.TICKS.SetTIMER0_CTRL_ENABLE(1)
}

// startTick starts the watchdog tick.
// On RP2040, the watchdog contained a tick generator used to generate a 1μs tick for the watchdog. This was also
// distributed to the system timer. On RP2350, the watchdog instead takes a tick input from the system-level ticks block. See Section 8.5.
func (wd *watchdogImpl) startTick(cycles uint32) {
	rp.TICKS.WATCHDOG_CTRL.SetBits(1)
}

func adjustCoreVoltage() bool {
	return false
}
