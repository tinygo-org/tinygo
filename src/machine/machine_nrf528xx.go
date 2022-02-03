//go:build nrf52 || nrf52840 || nrf52833
// +build nrf52 nrf52840 nrf52833

package machine

import (
	"device/nrf"
	"runtime/volatile"
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
	buf *[1]byte // 1-byte buffer for the Transfer method
}

// There are 3 SPI interfaces on the NRF528xx.
var (
	SPI0 = SPI{Bus: nrf.SPIM0, buf: new([1]byte)}
	SPI1 = SPI{Bus: nrf.SPIM1, buf: new([1]byte)}
	SPI2 = SPI{Bus: nrf.SPIM2, buf: new([1]byte)}
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
	buf := spi.buf[:]
	buf[0] = w
	err := spi.Tx(buf[:], buf[:])
	return buf[0], err
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

// PWM is one PWM peripheral, which consists of a counter and multiple output
// channels (that can be connected to actual pins). You can set the frequency
// using SetPeriod, but only for all the channels in this PWM peripheral at
// once.
type PWM struct {
	PWM *nrf.PWM_Type

	channelValues [4]volatile.Register16
}

// Configure enables and configures this PWM.
// On the nRF52 series, the maximum period is around 0.26s.
func (pwm *PWM) Configure(config PWMConfig) error {
	// Enable the peripheral.
	pwm.PWM.ENABLE.Set(nrf.PWM_ENABLE_ENABLE_Enabled << nrf.PWM_ENABLE_ENABLE_Pos)

	// Use up counting only. TODO: allow configuring as up-and-down.
	pwm.PWM.MODE.Set(nrf.PWM_MODE_UPDOWN_Up << nrf.PWM_MODE_UPDOWN_Pos)

	// Indicate there are four channels that each have a different value.
	pwm.PWM.DECODER.Set(nrf.PWM_DECODER_LOAD_Individual<<nrf.PWM_DECODER_LOAD_Pos | nrf.PWM_DECODER_MODE_RefreshCount<<nrf.PWM_DECODER_MODE_Pos)

	err := pwm.setPeriod(config.Period, true)
	if err != nil {
		return err
	}

	// Set the EasyDMA buffer, which has 4 values (one for each channel).
	pwm.PWM.SEQ[0].PTR.Set(uint32(uintptr(unsafe.Pointer(&pwm.channelValues[0]))))
	pwm.PWM.SEQ[0].CNT.Set(4)

	// SEQ[0] is not yet started, it will be started on the first
	// PWMChannel.Set() call.

	return nil
}

// SetPeriod updates the period of this PWM peripheral.
// To set a particular frequency, use the following formula:
//
//     period = 1e9 / frequency
//
// If you use a period of 0, a period that works well for LEDs will be picked.
//
// SetPeriod will not change the prescaler, but also won't change the current
// value in any of the channels. This means that you may need to update the
// value for the particular channel.
//
// Note that you cannot pick any arbitrary period after the PWM peripheral has
// been configured. If you want to switch between frequencies, pick the lowest
// frequency (longest period) once when calling Configure and adjust the
// frequency here as needed.
func (pwm *PWM) SetPeriod(period uint64) error {
	return pwm.setPeriod(period, false)
}

func (pwm *PWM) setPeriod(period uint64, updatePrescaler bool) error {
	const maxTop = 0x7fff // 15 bits counter

	// The top value is the number of PWM ticks a PWM period takes. It is
	// initially picked assuming an unlimited COUNTERTOP and no PWM prescaler.
	var top uint64
	if period == 0 {
		// The period is 0, which means "pick something reasonable for LEDs".
		top = maxTop
	} else {
		// The formula below calculates the following formula, optimized:
		//     period * (16e6 / 1e9)
		// The max frequency (16e6 or 16MHz) is set by the hardware.
		top = period * 2 / 125
	}

	// The ideal PWM period may be larger than would fit in the PWM counter,
	// which is only 15 bits (see maxTop). Therefore, try to make the PWM clock
	// speed lower with a prescaler to make the top value fit the COUNTERTOP.
	if updatePrescaler {
		// This function was called during Configure().
		switch {
		case top <= maxTop:
			pwm.PWM.PRESCALER.Set(nrf.PWM_PRESCALER_PRESCALER_DIV_1)
		case top/2 <= maxTop:
			pwm.PWM.PRESCALER.Set(nrf.PWM_PRESCALER_PRESCALER_DIV_2)
			top /= 2
		case top/4 <= maxTop:
			pwm.PWM.PRESCALER.Set(nrf.PWM_PRESCALER_PRESCALER_DIV_4)
			top /= 4
		case top/8 <= maxTop:
			pwm.PWM.PRESCALER.Set(nrf.PWM_PRESCALER_PRESCALER_DIV_8)
			top /= 8
		case top/16 <= maxTop:
			pwm.PWM.PRESCALER.Set(nrf.PWM_PRESCALER_PRESCALER_DIV_16)
			top /= 16
		case top/32 <= maxTop:
			pwm.PWM.PRESCALER.Set(nrf.PWM_PRESCALER_PRESCALER_DIV_32)
			top /= 32
		case top/64 <= maxTop:
			pwm.PWM.PRESCALER.Set(nrf.PWM_PRESCALER_PRESCALER_DIV_64)
			top /= 64
		case top/128 <= maxTop:
			pwm.PWM.PRESCALER.Set(nrf.PWM_PRESCALER_PRESCALER_DIV_128)
			top /= 128
		default:
			return ErrPWMPeriodTooLong
		}
	} else {
		// Do not update the prescaler, but use the already-configured
		// prescaler. This is the normal SetPeriod case, where the prescaler
		// must not be changed.
		prescaler := pwm.PWM.PRESCALER.Get()
		switch prescaler {
		case nrf.PWM_PRESCALER_PRESCALER_DIV_1:
			top /= 1
		case nrf.PWM_PRESCALER_PRESCALER_DIV_2:
			top /= 2
		case nrf.PWM_PRESCALER_PRESCALER_DIV_4:
			top /= 4
		case nrf.PWM_PRESCALER_PRESCALER_DIV_8:
			top /= 8
		case nrf.PWM_PRESCALER_PRESCALER_DIV_16:
			top /= 16
		case nrf.PWM_PRESCALER_PRESCALER_DIV_32:
			top /= 32
		case nrf.PWM_PRESCALER_PRESCALER_DIV_64:
			top /= 64
		case nrf.PWM_PRESCALER_PRESCALER_DIV_128:
			top /= 128
		}
		if top > maxTop {
			return ErrPWMPeriodTooLong
		}
	}
	pwm.PWM.COUNTERTOP.Set(uint32(top))

	// Apparently this is needed to apply the new COUNTERTOP.
	pwm.PWM.TASKS_SEQSTART[0].Set(1)

	return nil
}

// Top returns the current counter top, for use in duty cycle calculation. It
// will only change with a call to Configure or SetPeriod, otherwise it is
// constant.
//
// The value returned here is hardware dependent. In general, it's best to treat
// it as an opaque value that can be divided by some number and passed to
// pwm.Set (see pwm.Set for more information).
func (pwm *PWM) Top() uint32 {
	return pwm.PWM.COUNTERTOP.Get()
}

// Channel returns a PWM channel for the given pin.
func (pwm *PWM) Channel(pin Pin) (uint8, error) {
	config := uint32(pin)
	for ch := uint8(0); ch < 4; ch++ {
		channelConfig := pwm.PWM.PSEL.OUT[ch].Get()
		if channelConfig == 0xffffffff {
			// Unused channel. Configure it.
			pwm.PWM.PSEL.OUT[ch].Set(config)
			// Configure the pin (required by the reference manual).
			pin.Configure(PinConfig{Mode: PinOutput})
			// Set channel to zero and non-inverting.
			pwm.channelValues[ch].Set(0x8000)
			return ch, nil
		} else if channelConfig == config {
			// This channel is already configured for this pin.
			return ch, nil
		}
	}

	// All four pins are already in use with other pins.
	return 0, ErrInvalidOutputPin
}

// SetInverting sets whether to invert the output of this channel.
// Without inverting, a 25% duty cycle would mean the output is high for 25% of
// the time and low for the rest. Inverting flips the output as if a NOT gate
// was placed at the output, meaning that the output would be 25% low and 75%
// high with a duty cycle of 25%.
func (pwm *PWM) SetInverting(channel uint8, inverting bool) {
	ptr := &pwm.channelValues[channel]
	if inverting {
		ptr.Set(ptr.Get() &^ 0x8000)
	} else {
		ptr.Set(ptr.Get() | 0x8000)
	}
}

// Set updates the channel value. This is used to control the channel duty
// cycle. For example, to set it to a 25% duty cycle, use:
//
//     ch.Set(ch.Top() / 4)
//
// ch.Set(0) will set the output to low and ch.Set(ch.Top()) will set the output
// to high, assuming the output isn't inverted.
func (pwm *PWM) Set(channel uint8, value uint32) {
	// Update the channel value while retaining the polarity bit.
	ptr := &pwm.channelValues[channel]
	ptr.Set(ptr.Get()&0x8000 | uint16(value)&0x7fff)

	// Start the PWM, if it isn't already running.
	pwm.PWM.TASKS_SEQSTART[0].Set(1)
}
