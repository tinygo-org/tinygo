//go:build rp2040 || rp2350

package machine

import (
	"device/rp"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

type ioType struct {
	status volatile.Register32
	ctrl   volatile.Register32
}

type irqCtrl struct {
	intE [_NUMBANK0_IRQS]volatile.Register32
	intF [_NUMBANK0_IRQS]volatile.Register32
	intS [_NUMBANK0_IRQS]volatile.Register32
}

type irqSummary struct {
	proc [2]struct {
		secure    [2]volatile.Register32
		nonsecure [2]volatile.Register32
	}
	comaWake struct {
		secure    [2]volatile.Register32
		nonsecure [2]volatile.Register32
	}
}

type ioBank0Type struct {
	io                 [_NUMBANK0_GPIOS]ioType
	irqsum             [rp2350ExtraReg]irqSummary
	intR               [_NUMBANK0_IRQS]volatile.Register32
	proc0IRQctrl       irqCtrl
	proc1IRQctrl       irqCtrl
	dormantWakeIRQctrl irqCtrl
}

var ioBank0 = (*ioBank0Type)(unsafe.Pointer(rp.IO_BANK0))

type padsBank0Type struct {
	voltageSelect volatile.Register32
	io            [_NUMBANK0_GPIOS]volatile.Register32
}

var padsBank0 = (*padsBank0Type)(unsafe.Pointer(rp.PADS_BANK0))

// pinFunc represents a GPIO function.
//
// Each GPIO can have one function selected at a time.
// Likewise, each peripheral input (e.g. UART0 RX) should only be  selected
// on one GPIO at a time. If the same peripheral input is connected to multiple GPIOs,
// the peripheral sees the logical OR of these GPIO inputs.
type pinFunc uint8

func (p Pin) PortMaskSet() (*uint32, uint32) {
	return (*uint32)(unsafe.Pointer(&rp.SIO.GPIO_OUT_SET)), 1 << p
}

// set drives the pin high
func (p Pin) set() {
	mask := uint32(1) << p
	rp.SIO.GPIO_OUT_SET.Set(mask)
}

func (p Pin) PortMaskClear() (*uint32, uint32) {
	return (*uint32)(unsafe.Pointer(&rp.SIO.GPIO_OUT_CLR)), 1 << p
}

// clr drives the pin low
func (p Pin) clr() {
	mask := uint32(1) << p
	rp.SIO.GPIO_OUT_CLR.Set(mask)
}

// xor toggles the pin
func (p Pin) xor() {
	mask := uint32(1) << p
	rp.SIO.GPIO_OUT_XOR.Set(mask)
}

// get returns the pin value
func (p Pin) get() bool {
	return rp.SIO.GPIO_IN.HasBits(1 << p)
}

func (p Pin) ioCtrl() *volatile.Register32 {
	return &ioBank0.io[p].ctrl
}

func (p Pin) padCtrl() *volatile.Register32 {
	return &padsBank0.io[p]
}

func (p Pin) pullup() {
	p.padCtrl().SetBits(rp.PADS_BANK0_GPIO0_PUE)
	p.padCtrl().ClearBits(rp.PADS_BANK0_GPIO0_PDE)
}

func (p Pin) pulldown() {
	p.padCtrl().SetBits(rp.PADS_BANK0_GPIO0_PDE)
	p.padCtrl().ClearBits(rp.PADS_BANK0_GPIO0_PUE)
}

func (p Pin) pulloff() {
	p.padCtrl().ClearBits(rp.PADS_BANK0_GPIO0_PDE)
	p.padCtrl().ClearBits(rp.PADS_BANK0_GPIO0_PUE)
}

// setSlew sets pad slew rate control.
// true sets to fast. false sets to slow.
func (p Pin) setSlew(sr bool) {
	p.padCtrl().ReplaceBits(boolToBit(sr)<<rp.PADS_BANK0_GPIO0_SLEWFAST_Pos, rp.PADS_BANK0_GPIO0_SLEWFAST_Msk, 0)
}

// setSchmitt enables or disables Schmitt trigger.
func (p Pin) setSchmitt(trigger bool) {
	p.padCtrl().ReplaceBits(boolToBit(trigger)<<rp.PADS_BANK0_GPIO0_SCHMITT_Pos, rp.PADS_BANK0_GPIO0_SCHMITT_Msk, 0)
}

// setFunc will set pin function to fn.
func (p Pin) setFunc(fn pinFunc) {
	// Set input enable, Clear output disable
	p.padCtrl().ReplaceBits(rp.PADS_BANK0_GPIO0_IE,
		rp.PADS_BANK0_GPIO0_IE_Msk|rp.PADS_BANK0_GPIO0_OD_Msk, 0)

	// Zero all fields apart from fsel; we want this IO to do what the peripheral tells it.
	// This doesn't affect e.g. pullup/pulldown, as these are in pad controls.
	p.ioCtrl().Set(uint32(fn) << rp.IO_BANK0_GPIO0_CTRL_FUNCSEL_Pos)
}

// init initializes the gpio pin
func (p Pin) init() {
	mask := uint32(1) << p
	rp.SIO.GPIO_OE_CLR.Set(mask)
	p.clr()
}

// Set drives the pin high if value is true else drives it low.
func (p Pin) Set(value bool) {
	if p == NoPin {
		return
	}
	if value {
		p.set()
	} else {
		p.clr()
	}
}

// Get reads the pin value.
func (p Pin) Get() bool {
	return p.get()
}

// PinChange represents one or more trigger events that can happen on a given GPIO pin
// on the RP2040. ORed PinChanges are valid input to most IRQ functions.
type PinChange uint8

// Pin change interrupt constants for SetInterrupt.
const (
	// Edge falling
	PinFalling PinChange = 4 << iota
	// Edge rising
	PinRising

	PinToggle = PinFalling | PinRising
)

// Callbacks to be called for pins configured with SetInterrupt.
var (
	pinCallbacks [2][_NUMBANK0_GPIOS]func(Pin)
	setInt       [2][_NUMBANK0_GPIOS]bool
)

// SetInterrupt sets an interrupt to be executed when a particular pin changes
// state. The pin should already be configured as an input, including a pull up
// or down if no external pull is provided.
//
// This call will replace a previously set callback on this pin. You can pass a
// nil func to unset the pin change interrupt. If you do so, the change
// parameter is ignored and can be set to any value (such as 0).
func (p Pin) SetInterrupt(change PinChange, callback func(Pin)) error {
	if p == NoPin {
		return nil
	}
	if p > 31 || p < 0 {
		return ErrInvalidInputPin
	}
	core := CurrentCore()
	if callback == nil {
		// disable current interrupt
		p.setInterrupt(change, false)
		pinCallbacks[core][p] = nil
		return nil
	}

	if pinCallbacks[core][p] != nil {
		// Callback already configured. Should disable callback by passing a nil callback first.
		return ErrNoPinChangeChannel
	}
	p.setInterrupt(change, true)
	pinCallbacks[core][p] = callback

	if setInt[core][p] {
		// interrupt has already been set. Exit.
		return nil
	}
	interrupt.New(rp.IRQ_IO_IRQ_BANK0, gpioHandleInterrupt).Enable()
	irqSet(rp.IRQ_IO_IRQ_BANK0, true)
	return nil
}

// gpioHandleInterrupt finds the corresponding pin for the interrupt.
// C SDK equivalent of gpio_irq_handler
func gpioHandleInterrupt(intr interrupt.Interrupt) {

	core := CurrentCore()
	var gpio Pin
	for gpio = 0; gpio < _NUMBANK0_GPIOS; gpio++ {
		var base *irqCtrl
		switch core {
		case 0:
			base = &ioBank0.proc0IRQctrl
		case 1:
			base = &ioBank0.proc1IRQctrl
		}

		statreg := base.intS[gpio>>3].Get()
		change := getIntChange(gpio, statreg)
		if change != 0 {
			gpio.acknowledgeInterrupt(change)
			callback := pinCallbacks[core][gpio]
			if callback != nil {
				callback(gpio)
			}
		}
	}
}

// events returns the bit representation of the pin change for the rp2040.
func (change PinChange) events() uint32 {
	return uint32(change)
}

// intBit is the bit storage form of a PinChange for a given Pin
// in the IO_BANK0 interrupt registers (page 269 RP2040 Datasheet).
func (p Pin) ioIntBit(change PinChange) uint32 {
	return change.events() << (4 * (p % 8))
}

// Acquire interrupt data from a INT status register.
func getIntChange(p Pin, status uint32) PinChange {
	return PinChange(status>>(4*(p%8))) & 0xf
}
