//go:build rp2040
// +build rp2040

package machine

import (
	"device/rp"
	"errors"
	"math"
	"runtime/volatile"
	"unsafe"
)

var (
	ErrBadPeriod = errors.New("period outside valid range 8ns..268ms")
)

const (
	maxPWMPins = 29
)

// pwmGroup is one PWM peripheral, which consists of a counter and two output
// channels. You can set the frequency using SetPeriod,
// but only for all the channels in this PWM peripheral at once.
//
// div: integer value to reduce counting rate by. Must be greater than or equal to 1.
//
// cc: counter compare level. Contains 2 channel levels. The 16 LSBs are Channel A's level (Duty Cycle)
// and the 16 MSBs are Channel B's level.
//
// top: Wrap. Highest number counter will reach before wrapping over. usually 0xffff.
//
// csr: Clock mode. PWM_CH0_CSR_DIVMODE_xxx registers have 4 possible modes, of which Free-running is used.
// csr contains output polarity bit at PWM_CH0_CSR_x_INV where x is the channel.
// csr contains phase correction bit at PWM_CH0_CSR_PH_CORRECT_Msk.
// csr contains PWM enable bit at PWM_CH0_CSR_EN. If not enabled PWM will not be active.
//
// ctr: PWM counter value.
type pwmGroup struct {
	CSR volatile.Register32
	DIV volatile.Register32
	CTR volatile.Register32
	CC  volatile.Register32
	TOP volatile.Register32
}

// Equivalent of
//  var pwmSlice []pwmGroup = (*[8]pwmGroup)(unsafe.Pointer(rp.PWM))[:]
//  return &pwmSlice[index]
// 0x14 is the size of a pwmGroup.
func getPWMGroup(index uintptr) *pwmGroup {
	return (*pwmGroup)(unsafe.Pointer(uintptr(unsafe.Pointer(rp.PWM)) + 0x14*index))
}

// Hardware Pulse Width Modulation (PWM) API
// PWM peripherals available on RP2040. Each peripheral has 2 pins available for
// a total of 16 available PWM outputs. Some pins may not be available on some boards.
//
// The RP2040 PWM block has 8 identical slices. Each slice can drive two PWM output signals, or
// measure the frequency or duty cycle of an input signal. This gives a total of up to 16 controllable
// PWM outputs. All 30 GPIOs can be driven by the PWM block
//
// The PWM hardware functions by continuously comparing the input value to a free-running counter. This produces a
// toggling output where the amount of time spent at the high output level is proportional to the input value. The fraction of
// time spent at the high signal level is known as the duty cycle of the signal.
//
// The default behaviour of a PWM slice is to count upward until the wrap value (\ref pwm_config_set_wrap) is reached, and then
// immediately wrap to 0. PWM slices also offer a phase-correct mode, where the counter starts to count downward after
// reaching TOP, until it reaches 0 again.
var (
	PWM0 = getPWMGroup(0)
	PWM1 = getPWMGroup(1)
	PWM2 = getPWMGroup(2)
	PWM3 = getPWMGroup(3)
	PWM4 = getPWMGroup(4)
	PWM5 = getPWMGroup(5)
	PWM6 = getPWMGroup(6)
	PWM7 = getPWMGroup(7)
)

// Configure enables and configures this PWM.
func (pwm *pwmGroup) Configure(config PWMConfig) error {
	return pwm.init(config, true)
}

// Channel returns a PWM channel for the given pin. If pin does
// not belong to PWM peripheral ErrInvalidOutputPin error is returned.
// It also configures pin as PWM output.
func (pwm *pwmGroup) Channel(pin Pin) (channel uint8, err error) {
	if pin > maxPWMPins || pwmGPIOToSlice(pin) != pwm.peripheral() {
		return 3, ErrInvalidOutputPin
	}
	pin.Configure(PinConfig{PinPWM})
	return pwmGPIOToChannel(pin), nil
}

// PWMGroup returns the RP2040 PWM Group for the given pin. If pin does
// not belong to PWM peripheral ErrInvalidOutputPin error is returned.
func PWMGroup(pin Pin) (pwm *pwmGroup, err error) {
	sliceNum, err := PWMPeripheral(pin)
	if err != nil {
		return nil, err
	}
	return getPWMGroup(uintptr(sliceNum)), nil
}

// Peripheral returns the RP2040 PWM peripheral which ranges from 0 to 7. Each
// PWM peripheral has 2 channels, A and B which correspond to 0 and 1 in the program.
// This number corresponds to the package's PWM0 throughout PWM7 handles
func PWMPeripheral(pin Pin) (sliceNum uint8, err error) {
	if pin > maxPWMPins {
		return 0, ErrInvalidOutputPin
	}
	return pwmGPIOToSlice(pin), nil
}

// returns the number of the pwm peripheral (0-7)
func (pwm *pwmGroup) peripheral() uint8 {
	return uint8((uintptr(unsafe.Pointer(pwm)) - uintptr(unsafe.Pointer(rp.PWM))) / 0x14)
}

// SetPeriod updates the period of this PWM peripheral in nanoseconds.
// To set a particular frequency, use the following formula:
//
//     period = 1e9 / frequency
//
// Where frequency is in hertz. If you use a period of 0, a period
// that works well for LEDs will be picked.
//
// SetPeriod will try not to modify TOP if possible to reach the target period.
// If the period is unattainable with current TOP SetPeriod will modify TOP
// by the bare minimum to reach the target period. It will also enable phase
// correct to reach periods above 130ms.
func (p *pwmGroup) SetPeriod(period uint64) error {
	if period == 0 {
		period = 1e5
	}
	return p.setPeriod(period)
}

// Top returns the current counter top, for use in duty cycle calculation.
//
// The value returned here is hardware dependent. In general, it's best to treat
// it as an opaque value that can be divided by some number and passed to Set
// (see Set documentation for more information).
func (p *pwmGroup) Top() uint32 {
	return p.getWrap()
}

// Counter returns the current counter value of the timer in this PWM
// peripheral. It may be useful for debugging.
func (p *pwmGroup) Counter() uint32 {
	return (p.CTR.Get() & rp.PWM_CH0_CTR_CH0_CTR_Msk) >> rp.PWM_CH0_CTR_CH0_CTR_Pos
}

// Period returns the used PWM period in nanoseconds.
func (p *pwmGroup) Period() uint64 {
	periodPerCycle := cpuPeriod()
	top := p.getWrap()
	phc := p.getPhaseCorrect()
	Int, frac := p.getClockDiv()
	// Line below can overflow if operations done without care.
	return (16*uint64(Int) + uint64(frac)) * uint64((top+1)*(phc+1)*periodPerCycle) / 16 // cycles = (TOP+1) * (CSRPHCorrect + 1) * (DIV_INT + DIV_FRAC/16)
}

// SetInverting sets whether to invert the output of this channel.
// Without inverting, a 25% duty cycle would mean the output is high for 25% of
// the time and low for the rest. Inverting flips the output as if a NOT gate
// was placed at the output, meaning that the output would be 25% low and 75%
// high with a duty cycle of 25%.
func (p *pwmGroup) SetInverting(channel uint8, inverting bool) {
	channel &= 1
	p.setInverting(channel, inverting)
}

// Set updates the channel value. This is used to control the channel duty
// cycle, in other words the fraction of time the channel output is high (or low
// when inverted). For example, to set it to a 25% duty cycle, use:
//
//     pwm.Set(channel, pwm.Top() / 4)
//
// pwm.Set(channel, 0) will set the output to low and pwm.Set(channel,
// pwm.Top()) will set the output to high, assuming the output isn't inverted.
func (p *pwmGroup) Set(channel uint8, value uint32) {
	val := uint16(value)
	channel &= 1
	p.setChanLevel(channel, val)
}

// Get current level (last set by Set). Default value on initialization is 0.
func (p *pwmGroup) Get(channel uint8) (value uint32) {
	channel &= 1
	return uint32(p.getChanLevel(channel))
}

// SetTop sets TOP control register. Max value is 16bit (0xffff).
func (p *pwmGroup) SetTop(top uint32) {
	p.setWrap(uint16(top))
}

// SetCounter sets counter control register. Max value is 16bit (0xffff).
// Useful for synchronising two different PWM peripherals.
func (p *pwmGroup) SetCounter(ctr uint32) {
	p.CTR.Set(ctr)
}

// Enable enables or disables PWM peripheral channels.
func (p *pwmGroup) Enable(enable bool) {
	p.enable(enable)
}

// IsEnabled returns true if peripheral is enabled.
func (p *pwmGroup) IsEnabled() (enabled bool) {
	return (p.CSR.Get()&rp.PWM_CH0_CSR_EN_Msk)>>rp.PWM_CH0_CSR_EN_Pos != 0
}

// Initialise a PWM with settings from a configuration object.
// If start is true then PWM starts on initialization.
func (pwm *pwmGroup) init(config PWMConfig, start bool) error {
	// Not enable Phase correction
	pwm.setPhaseCorrect(false)

	// Clock mode set by default to Free running
	pwm.setDivMode(rp.PWM_CH0_CSR_DIVMODE_DIV)

	// Set Output polarity (false/false)
	pwm.setInverting(0, false)
	pwm.setInverting(1, false)

	// Set wrap. The highest value the counter will reach before returning to zero, also known as TOP.
	pwm.setWrap(0xffff)
	// period is set after TOP (Wrap).
	err := pwm.SetPeriod(config.Period)
	if err != nil {
		return err
	}
	// period already set beforea
	// Reset counter and compare (pwm level set to zero)
	pwm.CTR.ReplaceBits(0, rp.PWM_CH0_CTR_CH0_CTR_Msk, 0) // PWM_CH0_CTR_RESET
	pwm.CC.Set(0)                                         // PWM_CH0_CC_RESET

	pwm.enable(start)
	return nil
}

func (pwm *pwmGroup) setPhaseCorrect(correct bool) {
	pwm.CSR.ReplaceBits(boolToBit(correct)<<rp.PWM_CH0_CSR_PH_CORRECT_Pos, rp.PWM_CH0_CSR_PH_CORRECT_Msk, 0)
}

// Takes any of the following:
//  rp.PWM_CH0_CSR_DIVMODE_DIV, rp.PWM_CH0_CSR_DIVMODE_FALL,
//  rp.PWM_CH0_CSR_DIVMODE_LEVEL, rp.PWM_CH0_CSR_DIVMODE_RISE
func (pwm *pwmGroup) setDivMode(mode uint32) {
	pwm.CSR.ReplaceBits(mode<<rp.PWM_CH0_CSR_DIVMODE_Pos, rp.PWM_CH0_CSR_DIVMODE_Msk, 0)
}

// setPeriod sets the pwm peripheral period (frequency). Calculates DIV_INT,DIV_FRAC and sets it from following equation:
//  cycles = (TOP+1) * (CSRPHCorrect + 1) * (DIV_INT + DIV_FRAC/16)
// where cycles is amount of clock cycles per PWM period.
func (pwm *pwmGroup) setPeriod(period uint64) error {
	// This period calculation algorithm consists of
	// 1. Calculating best-fit prescale at a slightly lower-than-max TOP value
	// 2. Calculate TOP value to reach target period given the calculated prescale
	// 3. Apply calculated Prescale from step 1 and calculated Top from step 2
	const (
		maxTop = math.MaxUint16
		// start algorithm at 95% Top. This allows us to undershoot period with prescale.
		topStart     = 95 * maxTop / 100
		milliseconds = 1_000_000_000
		// Maximum Period is 268369920ns on rp2040, given by (16*255+15)*8*(1+0xffff)*(1+1)/16
		// With no phase shift max period is half of this value.
		maxPeriod = 268 * milliseconds
	)

	if period > maxPeriod || period < 8 {
		return ErrBadPeriod
	}
	if period > maxPeriod/2 {
		pwm.setPhaseCorrect(true) // Must enable Phase correct to reach large periods.
	}

	// clearing above expression:
	//  DIV_INT + DIV_FRAC/16 = cycles / ( (TOP+1) * (CSRPHCorrect+1) )  // DIV_FRAC/16 is always 0 in this equation
	// where cycles must be converted to time:
	//  target_period = cycles * period_per_cycle ==> cycles = target_period/period_per_cycle
	periodPerCycle := uint64(cpuPeriod())
	phc := uint64(pwm.getPhaseCorrect())
	rhs := 16 * period / ((1 + phc) * periodPerCycle * (1 + topStart)) // right-hand-side of equation, scaled so frac is not divided
	whole := rhs / 16
	frac := rhs % 16
	if whole > 0xff {
		whole = 0xff
	}

	// Step 2 is acquiring a better top value. Clearing the equation:
	// TOP =  cycles / ( (DIVINT+DIVFRAC/16) * (CSRPHCorrect+1) ) - 1
	top := 16*period/((16*whole+frac)*periodPerCycle*(1+phc)) - 1
	if top > maxTop {
		top = maxTop
	}
	pwm.SetTop(uint32(top))
	pwm.setClockDiv(uint8(whole), uint8(frac))
	return nil
}

// Int is integer value to reduce counting rate by. Must be greater than or equal to 1. DIV_INT is bits 4:11 (8 bits).
// frac's (DIV_FRAC) default value on reset is 0. Max value for frac is 15 (4 bits). This is known as a fixed-point
// fractional number.
//
//  cycles = (TOP+1) * (CSRPHCorrect + 1) * (DIV_INT + DIV_FRAC/16)
func (pwm *pwmGroup) setClockDiv(Int, frac uint8) {
	pwm.DIV.ReplaceBits((uint32(frac)<<rp.PWM_CH0_DIV_FRAC_Pos)|
		u32max(uint32(Int), 1)<<rp.PWM_CH0_DIV_INT_Pos, rp.PWM_CH0_DIV_FRAC_Msk|rp.PWM_CH0_DIV_INT_Msk, 0)
}

// Set the highest value the counter will reach before returning to 0. Also
// known as TOP.
//
// The counter wrap value is double-buffered in hardware. This means that,
// when the PWM is running, a write to the counter wrap value does not take
// effect until after the next time the PWM slice wraps (or, in phase-correct
// mode, the next time the slice reaches 0). If the PWM is not running, the
// write is latched in immediately.
func (pwm *pwmGroup) setWrap(wrap uint16) {
	pwm.TOP.ReplaceBits(uint32(wrap)<<rp.PWM_CH0_TOP_CH0_TOP_Pos, rp.PWM_CH0_TOP_CH0_TOP_Msk, 0)
}

// enables/disables the PWM peripheral with rp.PWM_CH0_CSR_EN bit.
func (pwm *pwmGroup) enable(enable bool) {
	pwm.CSR.ReplaceBits(boolToBit(enable)<<rp.PWM_CH0_CSR_EN_Pos, rp.PWM_CH0_CSR_EN_Msk, 0)
}

func (pwm *pwmGroup) setInverting(channel uint8, invert bool) {
	var pos uint8
	var msk uint32
	switch channel {
	case 0:
		pos = rp.PWM_CH0_CSR_A_INV_Pos
		msk = rp.PWM_CH0_CSR_A_INV_Msk
	case 1:
		pos = rp.PWM_CH0_CSR_B_INV_Pos
		msk = rp.PWM_CH0_CSR_B_INV_Msk
	}
	pwm.CSR.ReplaceBits(boolToBit(invert)<<pos, msk, 0)
}

// Set the current PWM counter compare value for one channel
//
// The counter compare register is double-buffered in hardware. This means
// that, when the PWM is running, a write to the counter compare values does
// not take effect until the next time the PWM slice wraps (or, in
// phase-correct mode, the next time the slice reaches 0). If the PWM is not
// running, the write is latched in immediately.
// Channel is 0 for A, 1 for B.
func (pwm *pwmGroup) setChanLevel(channel uint8, level uint16) {
	var pos uint8
	var mask uint32
	switch channel {
	case 0:
		pos = rp.PWM_CH0_CC_A_Pos
		mask = rp.PWM_CH0_CC_A_Msk
	case 1:
		pos = rp.PWM_CH0_CC_B_Pos
		mask = rp.PWM_CH0_CC_B_Msk
	}
	pwm.CC.ReplaceBits(uint32(level)<<pos, mask, 0)
}

func (pwm *pwmGroup) getChanLevel(channel uint8) (level uint16) {
	var pos uint8
	var mask uint32
	switch channel {
	case 0:
		pos = rp.PWM_CH0_CC_A_Pos
		mask = rp.PWM_CH0_CC_A_Msk
	case 1:
		pos = rp.PWM_CH0_CC_B_Pos
		mask = rp.PWM_CH0_CC_B_Msk
	}

	level = uint16((pwm.CC.Get() & mask) >> pos)
	return level
}

func (pwm *pwmGroup) getWrap() (top uint32) {
	return (pwm.TOP.Get() & rp.PWM_CH0_TOP_CH0_TOP_Msk) >> rp.PWM_CH0_TOP_CH0_TOP_Pos
}

func (pwm *pwmGroup) getPhaseCorrect() (phCorrect uint32) {
	return (pwm.CSR.Get() & rp.PWM_CH0_CSR_PH_CORRECT_Msk) >> rp.PWM_CH0_CSR_PH_CORRECT_Pos
}

func (pwm *pwmGroup) getClockDiv() (Int, frac uint8) {
	div := pwm.DIV.Get()
	return uint8((div & rp.PWM_CH0_DIV_INT_Msk) >> rp.PWM_CH0_DIV_INT_Pos), uint8((div & rp.PWM_CH0_DIV_FRAC_Msk) >> rp.PWM_CH0_DIV_FRAC_Pos)
}

// pwmGPIOToSlice Determine the PWM channel that is attached to the specified GPIO.
// gpio must be less than 30. Returns the PWM slice number that controls the specified GPIO.
func pwmGPIOToSlice(gpio Pin) (slicenum uint8) {
	return (uint8(gpio) >> 1) & 7
}

// Determine the PWM channel that is attached to the specified GPIO.
// Each slice 0 to 7 has two channels, A and B.
func pwmGPIOToChannel(gpio Pin) (channel uint8) {
	return uint8(gpio) & 1
}
