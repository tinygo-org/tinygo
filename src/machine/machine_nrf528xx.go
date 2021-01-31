// +build nrf52 nrf52840 nrf52833

package machine

import (
	"device/nrf"
	"unsafe"
)

func CPUFrequency() uint32 {
	return 64000000
}

// InitADC initializes the registers needed for ADC.
func InitADC() {
	return // no specific setup on nrf52 machine.
}

// Configure configures an ADC pin to be able to read analog data.
func (a ADC) Configure(ADCConfig) {
	return // no pin specific setup on nrf52 machine.
}

// Get returns the current value of a ADC pin in the range 0..0xffff.
func (a ADC) Get() uint16 {
	var pwmPin uint32
	var value int16

	switch a.Pin {
	case 2:
		pwmPin = nrf.SAADC_CH_PSELP_PSELP_AnalogInput0

	case 3:
		pwmPin = nrf.SAADC_CH_PSELP_PSELP_AnalogInput1

	case 4:
		pwmPin = nrf.SAADC_CH_PSELP_PSELP_AnalogInput2

	case 5:
		pwmPin = nrf.SAADC_CH_PSELP_PSELP_AnalogInput3

	case 28:
		pwmPin = nrf.SAADC_CH_PSELP_PSELP_AnalogInput4

	case 29:
		pwmPin = nrf.SAADC_CH_PSELP_PSELP_AnalogInput5

	case 30:
		pwmPin = nrf.SAADC_CH_PSELP_PSELP_AnalogInput6

	case 31:
		pwmPin = nrf.SAADC_CH_PSELP_PSELP_AnalogInput7

	default:
		return 0
	}

	nrf.SAADC.RESOLUTION.Set(nrf.SAADC_RESOLUTION_VAL_12bit)

	// Enable ADC.
	nrf.SAADC.ENABLE.Set(nrf.SAADC_ENABLE_ENABLE_Enabled << nrf.SAADC_ENABLE_ENABLE_Pos)
	for i := 0; i < 8; i++ {
		nrf.SAADC.CH[i].PSELN.Set(nrf.SAADC_CH_PSELP_PSELP_NC)
		nrf.SAADC.CH[i].PSELP.Set(nrf.SAADC_CH_PSELP_PSELP_NC)
	}

	// Configure ADC.
	nrf.SAADC.CH[0].CONFIG.Set(((nrf.SAADC_CH_CONFIG_RESP_Bypass << nrf.SAADC_CH_CONFIG_RESP_Pos) & nrf.SAADC_CH_CONFIG_RESP_Msk) |
		((nrf.SAADC_CH_CONFIG_RESP_Bypass << nrf.SAADC_CH_CONFIG_RESN_Pos) & nrf.SAADC_CH_CONFIG_RESN_Msk) |
		((nrf.SAADC_CH_CONFIG_GAIN_Gain1_5 << nrf.SAADC_CH_CONFIG_GAIN_Pos) & nrf.SAADC_CH_CONFIG_GAIN_Msk) |
		((nrf.SAADC_CH_CONFIG_REFSEL_Internal << nrf.SAADC_CH_CONFIG_REFSEL_Pos) & nrf.SAADC_CH_CONFIG_REFSEL_Msk) |
		((nrf.SAADC_CH_CONFIG_TACQ_3us << nrf.SAADC_CH_CONFIG_TACQ_Pos) & nrf.SAADC_CH_CONFIG_TACQ_Msk) |
		((nrf.SAADC_CH_CONFIG_MODE_SE << nrf.SAADC_CH_CONFIG_MODE_Pos) & nrf.SAADC_CH_CONFIG_MODE_Msk))

	// Set pin to read.
	nrf.SAADC.CH[0].PSELN.Set(pwmPin)
	nrf.SAADC.CH[0].PSELP.Set(pwmPin)

	// Destination for sample result.
	nrf.SAADC.RESULT.PTR.Set(uint32(uintptr(unsafe.Pointer(&value))))
	nrf.SAADC.RESULT.MAXCNT.Set(1) // One sample

	// Start tasks.
	nrf.SAADC.TASKS_START.Set(1)
	for nrf.SAADC.EVENTS_STARTED.Get() == 0 {
	}
	nrf.SAADC.EVENTS_STARTED.Set(0x00)

	// Start the sample task.
	nrf.SAADC.TASKS_SAMPLE.Set(1)

	// Wait until the sample task is done.
	for nrf.SAADC.EVENTS_END.Get() == 0 {
	}
	nrf.SAADC.EVENTS_END.Set(0x00)

	// Stop the ADC
	nrf.SAADC.TASKS_STOP.Set(1)
	for nrf.SAADC.EVENTS_STOPPED.Get() == 0 {
	}
	nrf.SAADC.EVENTS_STOPPED.Set(0)

	// Disable the ADC.
	nrf.SAADC.ENABLE.Set(nrf.SAADC_ENABLE_ENABLE_Disabled << nrf.SAADC_ENABLE_ENABLE_Pos)

	if value < 0 {
		value = 0
	}

	// Return 16-bit result from 12-bit value.
	return uint16(value << 4)
}

// SPI on the NRF.
type SPI struct {
	Bus *nrf.SPIM_Type
}

// There are 3 SPI interfaces on the NRF528xx.
var (
	SPI0 = SPI{Bus: nrf.SPIM0}
	SPI1 = SPI{Bus: nrf.SPIM1}
	SPI2 = SPI{Bus: nrf.SPIM2}
)

// SPIConfig is used to store config info for SPI.
type SPIConfig struct {
	Frequency uint32
	SCK       Pin
	SDO       Pin
	SDI       Pin
	LSBFirst  bool
	Mode      uint8
}

// Configure is intended to setup the SPI interface.
func (spi SPI) Configure(config SPIConfig) {
	// Disable bus to configure it
	spi.Bus.ENABLE.Set(nrf.SPIM_ENABLE_ENABLE_Disabled)

	// Pick a default frequency.
	if config.Frequency == 0 {
		config.Frequency = 4000000 // 4MHz
	}

	// set frequency
	var freq uint32
	switch {
	case config.Frequency >= 8000000:
		freq = nrf.SPIM_FREQUENCY_FREQUENCY_M8
	case config.Frequency >= 4000000:
		freq = nrf.SPIM_FREQUENCY_FREQUENCY_M4
	case config.Frequency >= 2000000:
		freq = nrf.SPIM_FREQUENCY_FREQUENCY_M2
	case config.Frequency >= 1000000:
		freq = nrf.SPIM_FREQUENCY_FREQUENCY_M1
	case config.Frequency >= 500000:
		freq = nrf.SPIM_FREQUENCY_FREQUENCY_K500
	case config.Frequency >= 250000:
		freq = nrf.SPIM_FREQUENCY_FREQUENCY_K250
	default: // below 250kHz, default to the lowest speed available
		freq = nrf.SPIM_FREQUENCY_FREQUENCY_K125
	}
	spi.Bus.FREQUENCY.Set(freq)

	var conf uint32

	// set bit transfer order
	if config.LSBFirst {
		conf = (nrf.SPIM_CONFIG_ORDER_LsbFirst << nrf.SPIM_CONFIG_ORDER_Pos)
	}

	// set mode
	switch config.Mode {
	case 0:
		conf &^= (nrf.SPIM_CONFIG_CPOL_ActiveHigh << nrf.SPIM_CONFIG_CPOL_Pos)
		conf &^= (nrf.SPIM_CONFIG_CPHA_Leading << nrf.SPIM_CONFIG_CPHA_Pos)
	case 1:
		conf &^= (nrf.SPIM_CONFIG_CPOL_ActiveHigh << nrf.SPIM_CONFIG_CPOL_Pos)
		conf |= (nrf.SPIM_CONFIG_CPHA_Trailing << nrf.SPIM_CONFIG_CPHA_Pos)
	case 2:
		conf |= (nrf.SPIM_CONFIG_CPOL_ActiveLow << nrf.SPIM_CONFIG_CPOL_Pos)
		conf &^= (nrf.SPIM_CONFIG_CPHA_Leading << nrf.SPIM_CONFIG_CPHA_Pos)
	case 3:
		conf |= (nrf.SPIM_CONFIG_CPOL_ActiveLow << nrf.SPIM_CONFIG_CPOL_Pos)
		conf |= (nrf.SPIM_CONFIG_CPHA_Trailing << nrf.SPIM_CONFIG_CPHA_Pos)
	default: // to mode
		conf &^= (nrf.SPIM_CONFIG_CPOL_ActiveHigh << nrf.SPIM_CONFIG_CPOL_Pos)
		conf &^= (nrf.SPIM_CONFIG_CPHA_Leading << nrf.SPIM_CONFIG_CPHA_Pos)
	}
	spi.Bus.CONFIG.Set(conf)

	// set pins
	if config.SCK == 0 && config.SDO == 0 && config.SDI == 0 {
		config.SCK = SPI0_SCK_PIN
		config.SDO = SPI0_SDO_PIN
		config.SDI = SPI0_SDI_PIN
	}
	spi.Bus.PSEL.SCK.Set(uint32(config.SCK))
	spi.Bus.PSEL.MOSI.Set(uint32(config.SDO))
	spi.Bus.PSEL.MISO.Set(uint32(config.SDI))

	// Re-enable bus now that it is configured.
	spi.Bus.ENABLE.Set(nrf.SPIM_ENABLE_ENABLE_Enabled)
}

// Transfer writes/reads a single byte using the SPI interface.
func (spi SPI) Transfer(w byte) (byte, error) {
	var wbuf, rbuf [1]byte
	wbuf[0] = w
	err := spi.Tx(wbuf[:], rbuf[:])
	return rbuf[0], err
}

// Tx handles read/write operation for SPI interface. Since SPI is a syncronous
// write/read interface, there must always be the same number of bytes written
// as bytes read. Therefore, if the number of bytes don't match it will be
// padded until they fit: if len(w) > len(r) the extra bytes received will be
// dropped and if len(w) < len(r) extra 0 bytes will be sent.
func (spi SPI) Tx(w, r []byte) error {
	// Unfortunately the hardware (on the nrf52832) only supports up to 255
	// bytes in the buffers, so if either w or r is longer than that the
	// transfer needs to be broken up in pieces.
	// The nrf52840 supports far larger buffers however, which isn't yet
	// supported.
	for len(r) != 0 || len(w) != 0 {
		// Prepare the SPI transfer: set the DMA pointers and lengths.
		if len(r) != 0 {
			spi.Bus.RXD.PTR.Set(uint32(uintptr(unsafe.Pointer(&r[0]))))
			n := uint32(len(r))
			if n > 255 {
				n = 255
			}
			spi.Bus.RXD.MAXCNT.Set(n)
			r = r[n:]
		}
		if len(w) != 0 {
			spi.Bus.TXD.PTR.Set(uint32(uintptr(unsafe.Pointer(&w[0]))))
			n := uint32(len(w))
			if n > 255 {
				n = 255
			}
			spi.Bus.TXD.MAXCNT.Set(n)
			w = w[n:]
		}

		// Do the transfer.
		// Note: this can be improved by not waiting until the transfer is
		// finished if the transfer is send-only (a common case).
		spi.Bus.TASKS_START.Set(1)
		for spi.Bus.EVENTS_END.Get() == 0 {
		}
		spi.Bus.EVENTS_END.Set(0)
	}

	return nil
}

// InitPWM initializes the registers needed for PWM.
func InitPWM() {
	return
}

// Configure configures a PWM pin for output.
func (pwm PWM) Configure() {
}

// Set turns on the duty cycle for a PWM pin using the provided value.
func (pwm PWM) Set(value uint16) {
	for i := 0; i < len(pwmChannelPins); i++ {
		if pwmChannelPins[i] == 0xFFFFFFFF || pwmChannelPins[i] == uint32(pwm.Pin) {
			pwmChannelPins[i] = uint32(pwm.Pin)
			pwmChannelSequence[i] = (value >> 2) | 0x8000 // set bit 15 to invert polarity

			p := pwms[i]

			p.PSEL.OUT[0].Set(uint32(pwm.Pin))
			p.PSEL.OUT[1].Set(uint32(pwm.Pin))
			p.PSEL.OUT[2].Set(uint32(pwm.Pin))
			p.PSEL.OUT[3].Set(uint32(pwm.Pin))
			p.ENABLE.Set(nrf.PWM_ENABLE_ENABLE_Enabled << nrf.PWM_ENABLE_ENABLE_Pos)
			p.PRESCALER.Set(nrf.PWM_PRESCALER_PRESCALER_DIV_2)
			p.MODE.Set(nrf.PWM_MODE_UPDOWN_Up)
			p.COUNTERTOP.Set(16384) // frequency
			p.LOOP.Set(0)
			p.DECODER.Set((nrf.PWM_DECODER_LOAD_Common << nrf.PWM_DECODER_LOAD_Pos) | (nrf.PWM_DECODER_MODE_RefreshCount << nrf.PWM_DECODER_MODE_Pos))
			p.SEQ[0].PTR.Set(uint32(uintptr(unsafe.Pointer(&pwmChannelSequence[i]))))
			p.SEQ[0].CNT.Set(1)
			p.SEQ[0].REFRESH.Set(1)
			p.SEQ[0].ENDDELAY.Set(0)
			p.TASKS_SEQSTART[0].Set(1)

			break
		}
	}
}
