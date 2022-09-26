//go:build stm32
// +build stm32

package machine

// The type alias `arrtype` should be defined to either uint32 or uint16
// depending on the size of that register in the MCU's TIM_Type structure.

import (
	"device/stm32"
	"runtime/interrupt"
	"runtime/volatile"
)

const PWM_MODE1 = 0x6

type TimerCallback func()
type ChannelCallback func(channel uint8)

type PinFunction struct {
	Pin     Pin
	AltFunc uint8
}

type TimerChannel struct {
	Pins []PinFunction
}

type TIM struct {
	EnableRegister *volatile.Register32
	EnableFlag     uint32
	Device         *stm32.TIM_Type
	Channels       [4]TimerChannel
	UpInterrupt    interrupt.Interrupt
	OCInterrupt    interrupt.Interrupt

	wraparoundCallback TimerCallback
	channelCallbacks   [4]ChannelCallback

	busFreq uint64
}

// Configure enables and configures this PWM.
func (t *TIM) Configure(config PWMConfig) error {
	// Enable device
	t.EnableRegister.SetBits(t.EnableFlag)

	err := t.setPeriod(config.Period, true)
	if err != nil {
		return err
	}

	// Auto-repeat
	t.Device.EGR.SetBits(stm32.TIM_EGR_UG)

	// Enable the timer
	t.Device.CR1.SetBits(stm32.TIM_CR1_CEN | stm32.TIM_CR1_ARPE)

	return nil
}

func (t *TIM) Count() uint32 {
	return uint32(t.Device.CNT.Get())
}

// SetWraparoundInterrupt configures a callback to be called each
// time the timer 'wraps-around'.
//
// For example, if `Configure(PWMConfig{Period:1000000})` is used,
// to set the timer period to 1ms, this callback will be called every
// 1ms.
func (t *TIM) SetWraparoundInterrupt(callback TimerCallback) error {
	// Disable this interrupt to prevent race conditions
	//t.UpInterrupt.Disable()

	// Ensure the interrupt handler for Update events is registered
	t.UpInterrupt = t.registerUPInterrupt()

	// Clear update flag
	t.Device.SR.ClearBits(stm32.TIM_SR_UIF)

	t.wraparoundCallback = callback
	t.UpInterrupt.SetPriority(0xc1)
	t.UpInterrupt.Enable()

	// Enable the hardware interrupt
	t.Device.DIER.SetBits(stm32.TIM_DIER_UIE)

	return nil
}

// Sets a callback to be called when a channel reaches it's set-point.
//
// For example, if `t.Set(ch, t.Top() / 4)` is used then the callback will
// be called every quarter-period of the timer's base Period.
func (t *TIM) SetMatchInterrupt(channel uint8, callback ChannelCallback) error {
	t.channelCallbacks[channel] = callback

	// Ensure the interrupt handler for Output Compare events is registered
	t.OCInterrupt = t.registerOCInterrupt()

	// Clear the interrupt flag
	t.Device.SR.ClearBits(stm32.TIM_SR_CC1IF << channel)

	// Enable the interrupt
	t.OCInterrupt.SetPriority(0xc1)
	t.OCInterrupt.Enable()

	// Enable the hardware interrupt
	t.Device.DIER.SetBits(stm32.TIM_DIER_CC1IE << channel)

	return nil
}

// SetPeriod updates the period of this PWM peripheral.
// To set a particular frequency, use the following formula:
//
//	period = 1e9 / frequency
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
func (t *TIM) SetPeriod(period uint64) error {
	return t.setPeriod(period, false)
}

func (t *TIM) setPeriod(period uint64, updatePrescaler bool) error {
	var top uint64
	if period == 0 {
		top = ARR_MAX
	} else {
		top = (period / 1000) * (t.busFreq / 1000) / 1000
	}

	var psc uint64
	if updatePrescaler {
		if top > ARR_MAX*PSC_MAX {
			return ErrPWMPeriodTooLong
		}

		// Select the minimum PSC that scales the ARR value into
		// range to maintain precision in ARR for changing frequencies
		// later
		psc = ceil(top, ARR_MAX)
		top = top / psc

		t.Device.PSC.Set(uint32(psc - 1))
	} else {
		psc = uint64(t.Device.PSC.Get()) + 1
		top = top / psc

		if top > ARR_MAX {
			return ErrPWMPeriodTooLong
		}
	}

	t.Device.ARR.Set(arrtype(top - 1))
	return nil
}

// Top returns the current counter top, for use in duty cycle calculation. It
// will only change with a call to Configure or SetPeriod, otherwise it is
// constant.
//
// The value returned here is hardware dependent. In general, it's best to treat
// it as an opaque value that can be divided by some number and passed to
// pwm.Set (see pwm.Set for more information).
func (t *TIM) Top() uint32 {
	return uint32(t.Device.ARR.Get()) + 1
}

// Channel returns a PWM channel for the given pin.
func (t *TIM) Channel(pin Pin) (uint8, error) {

	for chi, ch := range t.Channels {
		for _, p := range ch.Pins {
			if p.Pin == pin {
				t.configurePin(uint8(chi), p)
				//p.Pin.ConfigureAltFunc(PinConfig{Mode: pinModePWMOutput}, p.AltFunc)
				return uint8(chi), nil
			}
		}
	}

	return 0, ErrInvalidOutputPin
}

// Set updates the channel value. This is used to control the channel duty
// cycle. For example, to set it to a 25% duty cycle, use:
//
//	t.Set(ch, t.Top() / 4)
//
// ch.Set(0) will set the output to low and ch.Set(ch.Top()) will set the output
// to high, assuming the output isn't inverted.
func (t *TIM) Set(channel uint8, value uint32) {
	t.enableMainOutput()

	ccr := t.channelCCR(channel)
	ccmr, offset := t.channelCCMR(channel)

	// Disable interrupts whilst programming to prevent spurious OC interrupts
	mask := interrupt.Disable()

	// Set the PWM to Mode 1 (active below set value, inactive above)
	// Preload is disabled so we can change OC value within one update period.
	var ccmrVal uint32
	ccmrVal |= PWM_MODE1 << stm32.TIM_CCMR1_Output_OC1M_Pos
	ccmr.ReplaceBits(ccmrVal, 0xFF, offset)

	// Set the compare value
	ccr.Set(arrtype(value))

	// Enable the channel (if not already)
	t.Device.CCER.ReplaceBits(stm32.TIM_CCER_CC1E, 0xD, channel*4)

	// Force update
	t.Device.EGR.SetBits(stm32.TIM_EGR_CC1G << channel)

	// Reset Interrupt Flag
	t.Device.SR.ClearBits(stm32.TIM_SR_CC1IF << channel)

	// Restore interrupts
	interrupt.Restore(mask)
}

// Unset disables a channel, including any configured interrupts.
func (t *TIM) Unset(channel uint8) {
	// Disable interrupts whilst programming to prevent spurious OC interrupts
	mask := interrupt.Disable()

	// Disable the channel
	t.Device.CCER.ReplaceBits(0, 0xD, channel*4)

	// Reset to zero value
	ccr := t.channelCCR(channel)
	ccr.Set(0)

	// Disable the hardware interrupt
	t.Device.DIER.ClearBits(stm32.TIM_DIER_CC1IE << channel)

	// Clear the interrupt flag
	t.Device.SR.ClearBits(stm32.TIM_SR_CC1IF << channel)

	// Restore interrupts
	interrupt.Restore(mask)
}

// SetInverting sets whether to invert the output of this channel.
// Without inverting, a 25% duty cycle would mean the output is high for 25% of
// the time and low for the rest. Inverting flips the output as if a NOT gate
// was placed at the output, meaning that the output would be 25% low and 75%
// high with a duty cycle of 25%.
func (t *TIM) SetInverting(channel uint8, inverting bool) {
	// Enable the channel (if not already)

	var val = uint32(0)
	if inverting {
		val |= stm32.TIM_CCER_CC1P
	}

	t.Device.CCER.ReplaceBits(val, stm32.TIM_CCER_CC1P_Msk, channel*4)
}

func (t *TIM) handleUPInterrupt(interrupt.Interrupt) {
	if t.Device.SR.HasBits(stm32.TIM_SR_UIF) {
		// clear the update flag
		t.Device.SR.ClearBits(stm32.TIM_SR_UIF)

		if t.wraparoundCallback != nil {
			t.wraparoundCallback()
		}
	}
}

func (t *TIM) handleOCInterrupt(interrupt.Interrupt) {
	if t.Device.SR.HasBits(stm32.TIM_SR_CC1IF) {
		if t.channelCallbacks[0] != nil {
			t.channelCallbacks[0](0)
		}
	}
	if t.Device.SR.HasBits(stm32.TIM_SR_CC2IF) {
		if t.channelCallbacks[1] != nil {
			t.channelCallbacks[1](1)
		}
	}
	if t.Device.SR.HasBits(stm32.TIM_SR_CC3IF) {
		if t.channelCallbacks[2] != nil {
			t.channelCallbacks[2](2)
		}
	}
	if t.Device.SR.HasBits(stm32.TIM_SR_CC4IF) {
		if t.channelCallbacks[3] != nil {
			t.channelCallbacks[3](3)
		}
	}

	// Reset interrupt flags
	t.Device.SR.ClearBits(stm32.TIM_SR_CC1IF | stm32.TIM_SR_CC2IF | stm32.TIM_SR_CC3IF | stm32.TIM_SR_CC4IF)
}

func (t *TIM) channelCCR(channel uint8) *arrRegType {
	switch channel {
	case 0:
		return &t.Device.CCR1
	case 1:
		return &t.Device.CCR2
	case 2:
		return &t.Device.CCR3
	case 3:
		return &t.Device.CCR4
	}

	return nil
}

func (t *TIM) channelCCMR(channel uint8) (reg *volatile.Register32, offset uint8) {
	switch channel {
	case 0:
		return &t.Device.CCMR1_Output, 0
	case 1:
		return &t.Device.CCMR1_Output, 8
	case 2:
		return &t.Device.CCMR2_Output, 0
	case 3:
		return &t.Device.CCMR2_Output, 8
	}

	return nil, 0
}

//go:inline
func ceil(num uint64, denom uint64) uint64 {
	return (num + denom - 1) / denom
}
