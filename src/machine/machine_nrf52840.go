//go:build nrf52840
// +build nrf52840

package machine

import (
	"device/nrf"
	"errors"
	"unsafe"
)

// Get peripheral and pin number for this GPIO pin.
func (p Pin) getPortPin() (*nrf.GPIO_Type, uint32) {
	if p >= 32 {
		return nrf.P1, uint32(p - 32)
	} else {
		return nrf.P0, uint32(p)
	}
}

func (uart *UART) setPins(tx, rx Pin) {
	nrf.UART0.PSEL.TXD.Set(uint32(tx))
	nrf.UART0.PSEL.RXD.Set(uint32(rx))
}

func (i2c *I2C) setPins(scl, sda Pin) {
	i2c.Bus.PSEL.SCL.Set(uint32(scl))
	i2c.Bus.PSEL.SDA.Set(uint32(sda))
}

// PWM
var (
	PWM0 = &PWM{PWM: nrf.PWM0}
	PWM1 = &PWM{PWM: nrf.PWM1}
	PWM2 = &PWM{PWM: nrf.PWM2}
	PWM3 = &PWM{PWM: nrf.PWM3}
)

// PDM represents a PDM device
type PDM struct {
	device        *nrf.PDM_Type
	defaultBuffer int16
}

// Configure is intended to set up the PDM interface prior to use.
func (pdm *PDM) Configure(config PDMConfig) error {
	if config.DIN == 0 {
		return errors.New("No DIN pin provided in configuration")
	}

	if config.CLK == 0 {
		return errors.New("No CLK pin provided in configuration")
	}

	config.DIN.Configure(PinConfig{Mode: PinInput})
	config.CLK.Configure(PinConfig{Mode: PinOutput})
	pdm.device = nrf.PDM
	pdm.device.PSEL.DIN.Set(uint32(config.DIN))
	pdm.device.PSEL.CLK.Set(uint32(config.CLK))
	pdm.device.PDMCLKCTRL.Set(nrf.PDM_PDMCLKCTRL_FREQ_Default)
	pdm.device.RATIO.Set(nrf.PDM_RATIO_RATIO_Ratio64)
	pdm.device.GAINL.Set(nrf.PDM_GAINL_GAINL_DefaultGain)
	pdm.device.GAINR.Set(nrf.PDM_GAINR_GAINR_DefaultGain)
	pdm.device.ENABLE.Set(nrf.PDM_ENABLE_ENABLE_Enabled)

	if config.Stereo {
		pdm.device.MODE.Set(nrf.PDM_MODE_OPERATION_Stereo | nrf.PDM_MODE_EDGE_LeftRising)
	} else {
		pdm.device.MODE.Set(nrf.PDM_MODE_OPERATION_Mono | nrf.PDM_MODE_EDGE_LeftRising)
	}

	pdm.device.SAMPLE.SetPTR(uint32(uintptr(unsafe.Pointer(&pdm.defaultBuffer))))
	pdm.device.SAMPLE.SetMAXCNT_BUFFSIZE(1)
	pdm.device.SetTASKS_START(1)
	return nil
}

// Read stores a set of samples in the given target buffer. Pointer should
// represent first element of buffer slice, not the pointer to the slice itself.
func (pdm *PDM) Read(buf []int16) (uint32, error) {
	pdm.device.SAMPLE.SetPTR(uint32(uintptr(unsafe.Pointer(&buf[0]))))
	pdm.device.SAMPLE.MAXCNT.Set(uint32(len(buf)))
	pdm.device.EVENTS_STARTED.Set(0)

	// Step 1: wait for new sampling to start for target buffer
	for !pdm.device.EVENTS_STARTED.HasBits(nrf.PDM_EVENTS_STARTED_EVENTS_STARTED) {
	}
	pdm.device.EVENTS_END.Set(0)

	// Step 2: swap out buffers for next recording so we don't continue to
	// write to the target buffer
	pdm.device.EVENTS_STARTED.Set(0)
	pdm.device.SAMPLE.SetPTR(uint32(uintptr(unsafe.Pointer(&pdm.defaultBuffer))))
	pdm.device.SAMPLE.MAXCNT.Set(1)

	// Step 3: wait for original event to end
	for pdm.device.EVENTS_END.HasBits(nrf.PDM_EVENTS_STOPPED_EVENTS_STOPPED) {
	}

	// Step 4: wait for default buffer to start recording before proceeding
	// otherwise we see the contents of target buffer change later
	for !pdm.device.EVENTS_STARTED.HasBits(nrf.PDM_EVENTS_STARTED_EVENTS_STARTED) {
	}

	return uint32(len(buf)), nil
}
