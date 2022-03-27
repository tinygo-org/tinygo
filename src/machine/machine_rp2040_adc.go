//go:build rp2040
// +build rp2040

package machine

import (
	"device/rp"
	"sync"
)

// ADC peripheral mux channel. 0-4.
// Use channel constants to access valid channels.
type ADCChannel uint8

const (
	ADC0_CH ADCChannel = iota
	ADC1_CH
	ADC2_CH
	ADC3_CH     // Note: GPIO29 not broken out on pico board
	ADC_TEMP_CH // Internal temperature sensor
)

// Will be needed for round robin config when implemented
type ADCChConfig struct{}

// Used to serialise ADC sampling
var adcLock sync.Mutex

// Reset the ADC peripheral.
func InitADC() {
	rp.RESETS.RESET.SetBits(rp.RESETS_RESET_ADC)
	rp.RESETS.RESET.ClearBits(rp.RESETS_RESET_ADC)
	for !rp.RESETS.RESET_DONE.HasBits(rp.RESETS_RESET_ADC) {
	}

	// enable ADC
	rp.ADC.CS.Set(rp.ADC_CS_EN)

	waitForReady()
}

// Below methods are for standard tinygo, ADC{} pin based access to the ADC peripheral.

// The Configure method sets the ADC pin to analog input mode.
// Note the ADCConfig parameter is ignored.
// Does nothing if the GPIO is not an ADC muxed pin.
func (a ADC) Configure(config ADCConfig) {
	if c, ok := a.GetADCChannel(); ok {
		c.Configure(ADCChConfig{})
	}
}

// The Get method returns a one-shot ADC sample reading.
func (a ADC) Get() uint16 {
	if c, ok := a.GetADCChannel(); ok {
		return c.GetOnce()
	}
	// Not an ADC pin!
	return 0
}

// The GetADCChannel method returns the channel associated with the ADC pin.
// Returns !ok if the pin is not an ADC pin
func (a ADC) GetADCChannel() (c ADCChannel, ok bool) {
	ok = true
	switch a.Pin {
	case ADC0:
		c = ADC0_CH
	case ADC1:
		c = ADC1_CH
	case ADC2:
		c = ADC2_CH
	case ADC3:
		c = ADC3_CH
	default:
		// Invalid ADC pin
		ok = false
	}
	return c, ok
}

// Below methods are rp2040 specific, channel based access to the ADC peripheral.
// The adcChannel methods provide access to the temperature sensor. As well as the pin based channels.

// The Configure method sets the channel's associated pin to analog input mode.
// Configuring the ADC_TEMP_CH channel will power on the temperature sensor.
// The powered on temperature sensor increases ADC_AVDD current by approximately 40 μA.
func (c ADCChannel) Configure(config ADCChConfig) {
	if p, ok := c.Pin(); ok {
		p.Configure(PinConfig{Mode: PinAnalog})
	}
	if c == ADC_TEMP_CH {
		// Enable temperature sensor bias source
		rp.ADC.CS.SetBits(rp.ADC_CS_TS_EN)
	}
}

// The GetOnce method returns a one-shot ADC sample reading
func (c ADCChannel) GetOnce() uint16 {
	// Make it safe to sample multiple ADC channels in separate go routines.
	adcLock.Lock()
	rp.ADC.CS.ReplaceBits(uint32(c), 0b111, rp.ADC_CS_AINSEL_Pos)
	rp.ADC.CS.SetBits(rp.ADC_CS_START_ONCE)

	waitForReady()
	adcLock.Unlock()

	// rp2040 is a 12-bit ADC, scale raw reading to 16-bits.
	return uint16(rp.ADC.RESULT.Get()) << 4
}

// The GetVoltage method does a one-shot sample and returns the reading as a voltage between 0.0 and 3.3V
func (c ADCChannel) GetVoltage() (volts float32) {
	return float32(c.GetOnce()>>4) * float32(3.3) / float32(1<<12)
}

// The GetTemp method does a one-shot sample of the internal temperature sensor.
// Only works on the ADC_TEMP_CH channel. aka AINSEL=4. Other channels will return 0.0.
func (c ADCChannel) GetTemp() (celsius float32) {
	if c == ADC_TEMP_CH {
		// T = 27 - (ADC_voltage - 0.706)/0.001721
		return float32(27) - (c.GetVoltage()-float32(0.706))/float32(0.001721)
	}
	return 0.0
}

// The waitForReady function spins waiting for the ADC peripheral to become ready
func waitForReady() {
	for !rp.ADC.CS.HasBits(rp.ADC_CS_READY) {
	}
}

// The Pin method returns the GPIO Pin associated with the ADC mux channel, if it has one.
// Returns !ok if the channel does not have an associated GPIO, or if the channel number is invalid.
func (c ADCChannel) Pin() (p Pin, ok bool) {
	ok = true
	switch c {
	case ADC0_CH:
		p = ADC0
	case ADC1_CH:
		p = ADC1
	case ADC2_CH:
		p = ADC2
	case ADC3_CH:
		p = ADC3
	default:
		ok = false
	}
	return p, ok
}
