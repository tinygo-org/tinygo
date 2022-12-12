//go:build nrf

package machine

import (
	"device/nrf"
	"runtime/interrupt"
)

const deviceName = nrf.Device

const (
	PinInput         PinMode = (nrf.GPIO_PIN_CNF_DIR_Input << nrf.GPIO_PIN_CNF_DIR_Pos) | (nrf.GPIO_PIN_CNF_INPUT_Connect << nrf.GPIO_PIN_CNF_INPUT_Pos)
	PinInputPullup   PinMode = PinInput | (nrf.GPIO_PIN_CNF_PULL_Pullup << nrf.GPIO_PIN_CNF_PULL_Pos)
	PinInputPulldown PinMode = PinInput | (nrf.GPIO_PIN_CNF_PULL_Pulldown << nrf.GPIO_PIN_CNF_PULL_Pos)
	PinOutput        PinMode = (nrf.GPIO_PIN_CNF_DIR_Output << nrf.GPIO_PIN_CNF_DIR_Pos) | (nrf.GPIO_PIN_CNF_INPUT_Connect << nrf.GPIO_PIN_CNF_INPUT_Pos)
)

type PinChange uint8

// Pin change interrupt constants for SetInterrupt.
const (
	PinRising  PinChange = nrf.GPIOTE_CONFIG_POLARITY_LoToHi
	PinFalling PinChange = nrf.GPIOTE_CONFIG_POLARITY_HiToLo
	PinToggle  PinChange = nrf.GPIOTE_CONFIG_POLARITY_Toggle
)

// Callbacks to be called for pins configured with SetInterrupt.
var pinCallbacks [len(nrf.GPIOTE.CONFIG)]func(Pin)

// Configure this pin with the given configuration.
func (p Pin) Configure(config PinConfig) {
	cfg := config.Mode | nrf.GPIO_PIN_CNF_DRIVE_S0S1 | nrf.GPIO_PIN_CNF_SENSE_Disabled
	port, pin := p.getPortPin()
	port.PIN_CNF[pin].Set(uint32(cfg))
}

// Set the pin to high or low.
// Warning: only use this on an output pin!
func (p Pin) Set(high bool) {
	port, pin := p.getPortPin()
	if high {
		port.OUTSET.Set(1 << pin)
	} else {
		port.OUTCLR.Set(1 << pin)
	}
}

// Return the register and mask to enable a given GPIO pin. This can be used to
// implement bit-banged drivers.
func (p Pin) PortMaskSet() (*uint32, uint32) {
	port, pin := p.getPortPin()
	return &port.OUTSET.Reg, 1 << pin
}

// Return the register and mask to disable a given port. This can be used to
// implement bit-banged drivers.
func (p Pin) PortMaskClear() (*uint32, uint32) {
	port, pin := p.getPortPin()
	return &port.OUTCLR.Reg, 1 << pin
}

// Get returns the current value of a GPIO pin when the pin is configured as an
// input or as an output.
func (p Pin) Get() bool {
	port, pin := p.getPortPin()
	return (port.IN.Get()>>pin)&1 != 0
}

// SetInterrupt sets an interrupt to be executed when a particular pin changes
// state. The pin should already be configured as an input, including a pull up
// or down if no external pull is provided.
//
// This call will replace a previously set callback on this pin. You can pass a
// nil func to unset the pin change interrupt. If you do so, the change
// parameter is ignored and can be set to any value (such as 0).
func (p Pin) SetInterrupt(change PinChange, callback func(Pin)) error {
	// Some variables to easily check whether a channel was already configured
	// as an event channel for the given pin.
	// This is not just an optimization, this is requred: the datasheet says
	// that configuring more than one channel for a given pin results in
	// unpredictable behavior.
	expectedConfigMask := uint32(nrf.GPIOTE_CONFIG_MODE_Msk | nrf.GPIOTE_CONFIG_PSEL_Msk)
	expectedConfig := nrf.GPIOTE_CONFIG_MODE_Event<<nrf.GPIOTE_CONFIG_MODE_Pos | uint32(p)<<nrf.GPIOTE_CONFIG_PSEL_Pos

	foundChannel := false
	for i := range nrf.GPIOTE.CONFIG {
		config := nrf.GPIOTE.CONFIG[i].Get()
		if config == 0 || config&expectedConfigMask == expectedConfig {
			// Found an empty GPIOTE channel or one that was already configured
			// for this pin.
			if callback == nil {
				// Disable this channel.
				nrf.GPIOTE.INTENCLR.Set(uint32(1 << uint(i)))
				pinCallbacks[i] = nil
				return nil
			}
			// Enable this channel with the given callback.
			nrf.GPIOTE.INTENCLR.Set(uint32(1 << uint(i)))
			nrf.GPIOTE.CONFIG[i].Set(nrf.GPIOTE_CONFIG_MODE_Event<<nrf.GPIOTE_CONFIG_MODE_Pos |
				uint32(p)<<nrf.GPIOTE_CONFIG_PSEL_Pos |
				uint32(change)<<nrf.GPIOTE_CONFIG_POLARITY_Pos)
			pinCallbacks[i] = callback
			nrf.GPIOTE.INTENSET.Set(uint32(1 << uint(i)))
			foundChannel = true
			break
		}
	}

	if !foundChannel {
		return ErrNoPinChangeChannel
	}

	// Set and enable the GPIOTE interrupt. It's not a problem if this happens
	// more than once.
	interrupt.New(nrf.IRQ_GPIOTE, func(interrupt.Interrupt) {
		for i := range nrf.GPIOTE.EVENTS_IN {
			if nrf.GPIOTE.EVENTS_IN[i].Get() != 0 {
				nrf.GPIOTE.EVENTS_IN[i].Set(0)
				pin := Pin((nrf.GPIOTE.CONFIG[i].Get() & nrf.GPIOTE_CONFIG_PSEL_Msk) >> nrf.GPIOTE_CONFIG_PSEL_Pos)
				pinCallbacks[i](pin)
			}
		}
	}).Enable()

	// Everything was configured correctly.
	return nil
}

// I2CConfig is used to store config info for I2C.
type I2CConfig struct {
	Frequency uint32
	SCL       Pin
	SDA       Pin
}

// Configure is intended to setup the I2C interface.
func (i2c *I2C) Configure(config I2CConfig) error {

	i2c.Bus.ENABLE.Set(nrf.TWI_ENABLE_ENABLE_Disabled)

	// Default I2C bus speed is 100 kHz.
	if config.Frequency == 0 {
		config.Frequency = 100 * KHz
	}
	// Default I2C pins if not set.
	if config.SDA == 0 && config.SCL == 0 {
		config.SDA = SDA_PIN
		config.SCL = SCL_PIN
	}

	// do config
	sclPort, sclPin := config.SCL.getPortPin()
	sclPort.PIN_CNF[sclPin].Set((nrf.GPIO_PIN_CNF_DIR_Input << nrf.GPIO_PIN_CNF_DIR_Pos) |
		(nrf.GPIO_PIN_CNF_INPUT_Connect << nrf.GPIO_PIN_CNF_INPUT_Pos) |
		(nrf.GPIO_PIN_CNF_PULL_Pullup << nrf.GPIO_PIN_CNF_PULL_Pos) |
		(nrf.GPIO_PIN_CNF_DRIVE_S0D1 << nrf.GPIO_PIN_CNF_DRIVE_Pos) |
		(nrf.GPIO_PIN_CNF_SENSE_Disabled << nrf.GPIO_PIN_CNF_SENSE_Pos))

	sdaPort, sdaPin := config.SDA.getPortPin()
	sdaPort.PIN_CNF[sdaPin].Set((nrf.GPIO_PIN_CNF_DIR_Input << nrf.GPIO_PIN_CNF_DIR_Pos) |
		(nrf.GPIO_PIN_CNF_INPUT_Connect << nrf.GPIO_PIN_CNF_INPUT_Pos) |
		(nrf.GPIO_PIN_CNF_PULL_Pullup << nrf.GPIO_PIN_CNF_PULL_Pos) |
		(nrf.GPIO_PIN_CNF_DRIVE_S0D1 << nrf.GPIO_PIN_CNF_DRIVE_Pos) |
		(nrf.GPIO_PIN_CNF_SENSE_Disabled << nrf.GPIO_PIN_CNF_SENSE_Pos))

	if config.Frequency >= 400*KHz {
		i2c.Bus.FREQUENCY.Set(nrf.TWI_FREQUENCY_FREQUENCY_K400)
	} else {
		i2c.Bus.FREQUENCY.Set(nrf.TWI_FREQUENCY_FREQUENCY_K100)
	}

	i2c.setPins(config.SCL, config.SDA)

	i2c.Bus.ENABLE.Set(nrf.TWI_ENABLE_ENABLE_Enabled)

	return nil
}

// signalStop sends a stop signal to the I2C peripheral and waits for confirmation.
func (i2c *I2C) signalStop() {
	i2c.Bus.TASKS_STOP.Set(1)
	for i2c.Bus.EVENTS_STOPPED.Get() == 0 {
	}
	i2c.Bus.EVENTS_STOPPED.Set(0)
}
