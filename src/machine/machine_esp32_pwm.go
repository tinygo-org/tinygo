//go:build esp32
// +build esp32

package machine

import (
	"device/esp"
	"errors"
	"runtime/volatile"
	"unsafe"
)

// PWM peripheral errors
var (
	errPWMPeriodTooShort = errors.New("pwm: period too short")
)

// PWM is one PWM peripheral, which consists of a timer and the associated
// channels. There are 4 high-speed timers available (PWM0-PWM3), each can
// use any of the 8 high-speed channels.
//
// The ESP32 LEDC peripheral is used for PWM generation. It provides flexible
// frequency and duty cycle control with configurable resolution (1-20 bits).
type PWM struct {
	num uint8 // Timer number (0-3 for high-speed)
}

// pwmChannelInfo tracks the state of each LEDC channel
type pwmChannelInfo struct {
	pin   Pin   // The pin assigned to this channel (NoPin if unused)
	timer uint8 // The timer this channel is bound to
	inUse bool  // Whether this channel is currently in use
}

// Number of PWM timers and channels
const (
	pwmTimerCount   = 4 // Number of high-speed timers (0-3)
	pwmChannelCount = 8 // Number of high-speed channels (0-7)
)

// pwmChannels tracks which pin and timer is assigned to each channel (0-7)
var pwmChannels [pwmChannelCount]pwmChannelInfo

// Hardware PWM peripherals available on ESP32.
// These use the high-speed LEDC timers for glitch-free PWM updates.
var (
	PWM0 = &PWM{num: 0}
	PWM1 = &PWM{num: 1}
	PWM2 = &PWM{num: 2}
	PWM3 = &PWM{num: 3}
)

// LEDC peripheral constants
const (
	// Clock enable bit in DPORT.PERIP_CLK_EN for LEDC
	ledcClockEnable = 1 << 11

	// GPIO matrix output signal numbers for LEDC high-speed channels
	ledcHSSignalBase = 71 // LEDC_HS_SIG_OUT0, channels are consecutive (71-78)

	// GPIO matrix output inversion bit
	gpioMatrixInvertBit = 1 << 9

	// Clock divider fractional bits (register value = actual_divider * 256)
	dividerFracBits = 8

	// Maximum values
	maxDivider    = 0x3FFFF              // 18-bit divider register value
	minDivider    = 1 << dividerFracBits // Minimum divider = 256 (represents 1.0)
	maxResolution = 20                   // Maximum duty resolution in bits
	minResolution = 1                    // Minimum duty resolution in bits

	// Default values
	defaultPeriodNs   = 1_000_000 // 1ms = 1kHz, good for LEDs
	defaultResolution = 13        // 13-bit resolution for default period
)

// LEDC register bit positions and masks for timer configuration
const (
	// HSTIMER_CONF register
	timerDivNumPos   = 5 // DIV_NUM position (bits 5-22)
	timerDivNumMask  = 0x3FFFF << timerDivNumPos
	timerLimMask     = 0x1F // duty_resolution (bits 0-4), stores bit width directly
	timerPauseMask   = 1 << 23
	timerRstMask     = 1 << 24
	timerTickSelMask = 1 << 25 // TICK_SEL bit (0=REF_TICK, 1=APB_CLK)
)

// LEDC register bit positions and masks for channel configuration
const (
	// HSCH_CONF0 register
	chanTimerSelMask = 0x3     // TIMER_SEL (bits 0-1)
	chanSigOutEnMask = 1 << 2  // SIG_OUT_EN bit
	chanClkEnMask    = 1 << 31 // CLK_EN bit

	// HSCH_CONF1 register
	chanDutyCyclePos  = 10 // DUTY_CYCLE position (bits 10-19)
	chanDutyNumPos    = 20 // DUTY_NUM position (bits 20-29)
	chanDutyIncMask   = 1 << 30
	chanDutyStartMask = 1 << 31

	// HSCH_DUTY register - duty value uses bits 0-24 (25 bits total)
	// The lower 4 bits are fractional, upper 20 bits are integer
	chanDutyFracBits = 4
)

// pwmState holds the current configuration for each PWM timer
type pwmState struct {
	resolution uint8 // Current duty resolution in bits
	configured bool  // Whether the timer has been configured
}

var pwmStates [pwmTimerCount]pwmState

// Configure enables and configures this PWM peripheral.
// The period is specified in nanoseconds. A period of 0 will select a default
// period suitable for LED dimming (~1kHz).
func (pwm *PWM) Configure(config PWMConfig) error {
	// Enable LEDC peripheral clock
	esp.DPORT.PERIP_CLK_EN.SetBits(ledcClockEnable)
	// Clear reset bit
	esp.DPORT.PERIP_RST_EN.ClearBits(ledcClockEnable)

	// Select APB clock (80MHz) for LEDC
	esp.LEDC.CONF.Set(1) // APB_CLK_SEL = 1

	// Calculate timer configuration
	divider, resolution, err := pwm.calculateConfig(config.Period)
	if err != nil {
		return err
	}

	// Store the resolution for duty cycle calculations
	pwmStates[pwm.num].resolution = resolution
	pwmStates[pwm.num].configured = true

	// Get timer configuration register
	timerConf := pwm.timerConf()

	// Configure timer:
	// - Use APB_CLK (80MHz) as clock source
	// - Set divider
	// - Set resolution (duty_resolution field stores the bit width directly)
	var conf uint32
	conf |= timerTickSelMask                              // APB_CLK
	conf |= (divider << timerDivNumPos) & timerDivNumMask // Divider
	conf |= uint32(resolution) & timerLimMask             // Resolution (bit width)

	// Reset the timer (set rst=1, then rst=0)
	timerConf.Set(conf | timerRstMask)
	timerConf.Set(conf)

	return nil
}

// calculateConfig determines the optimal divider and resolution for a given period.
// Returns divider (with 8 fractional bits), resolution, and any error.
func (pwm *PWM) calculateConfig(period uint64) (uint32, uint8, error) {
	if period == 0 {
		period = defaultPeriodNs
	}

	// Formula: period_ns = (2^resolution * divider) / 80MHz * 1e9
	// Where divider is the actual divider (not the register value)
	// Register value = actual_divider * 256 (8 fractional bits)
	//
	// period_ns = (2^resolution * divider_reg / 256) / 80MHz * 1e9
	// divider_reg = period_ns * 80 * 256 / (2^resolution * 1000)

	var lastDividerReg uint64

	// Try to find the highest resolution that gives a valid divider
	for resolution := uint8(maxResolution); resolution >= minResolution; resolution-- {
		resolutionValue := uint64(1) << resolution

		// Calculate divider register value (includes 8 fractional bits)
		dividerReg := (period * 80 * (1 << dividerFracBits)) / (resolutionValue * 1000)
		lastDividerReg = dividerReg

		if dividerReg < minDivider {
			// Period too short for this resolution, try lower resolution
			continue
		}

		if dividerReg <= maxDivider {
			return uint32(dividerReg), resolution, nil
		}
	}

	// Determine which error to return
	if lastDividerReg < minDivider {
		return 0, 0, errPWMPeriodTooShort
	}
	return 0, 0, ErrPWMPeriodTooLong
}

// Channel returns a PWM channel for the given pin. If the pin is already
// configured for this PWM peripheral, the same channel is returned.
// The pin is configured for PWM output.
func (pwm *PWM) Channel(pin Pin) (uint8, error) {
	if !pwmStates[pwm.num].configured {
		// Timer not configured, configure with default period
		if err := pwm.Configure(PWMConfig{}); err != nil {
			return 0, err
		}
	}

	// Check if this pin is already assigned to a channel on THIS timer
	for ch := uint8(0); ch < pwmChannelCount; ch++ {
		if pwmChannels[ch].inUse &&
			pwmChannels[ch].pin == pin &&
			pwmChannels[ch].timer == pwm.num {
			return ch, nil
		}
	}

	// Check if pin is used by a different timer (error case)
	for ch := uint8(0); ch < pwmChannelCount; ch++ {
		if pwmChannels[ch].inUse && pwmChannels[ch].pin == pin {
			// Pin is already used by a different timer
			return 0, ErrInvalidOutputPin
		}
	}

	// Find an available channel
	for ch := uint8(0); ch < pwmChannelCount; ch++ {
		if !pwmChannels[ch].inUse {
			// Found an available channel
			pwmChannels[ch] = pwmChannelInfo{
				pin:   pin,
				timer: pwm.num,
				inUse: true,
			}

			// Configure the GPIO for PWM output through GPIO matrix
			signal := uint32(ledcHSSignalBase + ch)
			pin.configure(PinConfig{Mode: PinOutput}, signal)

			// Configure the channel
			chanConf0 := pwm.channelConf0(ch)

			// Set channel configuration:
			// - Enable clock
			// - Select this timer
			// - Enable signal output
			var conf uint32
			conf |= chanClkEnMask                      // Enable clock
			conf |= uint32(pwm.num) & chanTimerSelMask // Select timer
			conf |= chanSigOutEnMask                   // Enable output
			chanConf0.Set(conf)

			// Initialize duty to 0
			pwm.channelDuty(ch).Set(0)

			// Set HPOINT to 0 (start of cycle)
			pwm.channelHpoint(ch).Set(0)

			// Configure CONF1 for non-fading operation and trigger duty update
			// duty_scale=0, duty_cycle=1, duty_num=1, duty_inc=1, duty_start=1
			var conf1 uint32
			conf1 |= 1 << chanDutyCyclePos // duty_cycle = 1
			conf1 |= 1 << chanDutyNumPos   // duty_num = 1
			conf1 |= chanDutyIncMask       // duty_inc = 1
			conf1 |= chanDutyStartMask     // duty_start = 1
			pwm.channelConf1(ch).Set(conf1)

			return ch, nil
		}
	}

	return 0, ErrInvalidOutputPin
}

// ReleaseChannel releases a PWM channel, making it available for other uses.
// The pin is not reconfigured; call pin.Configure() to change its function.
func (pwm *PWM) ReleaseChannel(channel uint8) error {
	if !pwm.isValidChannel(channel) {
		return ErrInvalidOutputPin
	}

	if !pwmChannels[channel].inUse || pwmChannels[channel].timer != pwm.num {
		// Channel not in use or belongs to different timer
		return ErrInvalidOutputPin
	}

	// Disable signal output
	pwm.channelConf0(channel).ClearBits(chanSigOutEnMask)

	// Clear channel tracking
	pwmChannels[channel] = pwmChannelInfo{}

	return nil
}

// IsConnected returns true if the given channel is in use by this PWM peripheral.
func (pwm *PWM) IsConnected(channel uint8) bool {
	if !pwm.isValidChannel(channel) {
		return false
	}
	return pwmChannels[channel].inUse && pwmChannels[channel].timer == pwm.num
}

// isValidChannel returns true if the channel number is valid.
func (pwm *PWM) isValidChannel(channel uint8) bool {
	return channel < pwmChannelCount
}

// Set updates the channel value. This is used to control the channel duty
// cycle. For example, to set it to a 25% duty cycle, use:
//
//	pwm.Set(channel, pwm.Top() / 4)
//
// pwm.Set(channel, 0) will set the output to low and pwm.Set(channel,
// pwm.Top()) will set the output to high, assuming the output isn't inverted.
func (pwm *PWM) Set(channel uint8, value uint32) {
	if !pwm.isValidChannel(channel) {
		return
	}

	resolution := pwmStates[pwm.num].resolution
	if resolution == 0 {
		resolution = defaultResolution
	}
	maxValue := uint32((1 << resolution) - 1)

	// Clamp value to valid range
	// Note: Setting duty to exactly 2^resolution causes hardware overflow
	if value > maxValue {
		value = maxValue
	}

	// The duty register uses 4 fractional bits, so shift left by 4
	dutyValue := value << chanDutyFracBits

	// Set the duty value
	pwm.channelDuty(channel).Set(dutyValue)

	// Configure CONF1 and trigger duty update
	// For non-fading operation: duty_scale=0, duty_cycle=1, duty_num=1, duty_inc=1
	// Then set duty_start=1 to apply the new duty value
	var conf1 uint32
	conf1 |= 1 << chanDutyCyclePos // duty_cycle = 1
	conf1 |= 1 << chanDutyNumPos   // duty_num = 1
	conf1 |= chanDutyIncMask       // duty_inc = 1
	conf1 |= chanDutyStartMask     // duty_start = 1
	pwm.channelConf1(channel).Set(conf1)

	// Ensure signal output is enabled
	pwm.channelConf0(channel).SetBits(chanSigOutEnMask)
}

// Get returns the current duty cycle value for the given channel.
func (pwm *PWM) Get(channel uint8) uint32 {
	if !pwm.isValidChannel(channel) {
		return 0
	}

	// Read from the duty read register and remove fractional bits
	return pwm.channelDutyR(channel).Get() >> chanDutyFracBits
}

// SetPeriod updates the period of this PWM peripheral in nanoseconds.
// To set a particular frequency, use the following formula:
//
//	period = 1e9 / frequency
//
// SetPeriod will try to maintain the current duty cycle ratio when changing
// the period for channels bound to this timer.
func (pwm *PWM) SetPeriod(period uint64) error {
	// Calculate new configuration
	divider, resolution, err := pwm.calculateConfig(period)
	if err != nil {
		return err
	}

	oldResolution := pwmStates[pwm.num].resolution
	pwmStates[pwm.num].resolution = resolution

	// Update timer configuration
	timerConf := pwm.timerConf()

	// Read current config, update divider and resolution
	conf := timerConf.Get()
	conf &^= timerDivNumMask | timerLimMask
	conf |= (divider << timerDivNumPos) & timerDivNumMask
	conf |= uint32(resolution) & timerLimMask
	timerConf.Set(conf)

	// If resolution changed, scale duty values for channels bound to THIS timer
	if resolution != oldResolution {
		for ch := uint8(0); ch < pwmChannelCount; ch++ {
			if pwmChannels[ch].inUse && pwmChannels[ch].timer == pwm.num {
				// Read current duty (includes fractional bits)
				currentDuty := pwm.channelDuty(ch).Get() >> chanDutyFracBits

				// Scale to new resolution
				if oldResolution > resolution {
					currentDuty >>= (oldResolution - resolution)
				} else {
					currentDuty <<= (resolution - oldResolution)
				}

				// Apply new duty
				pwm.Set(ch, currentDuty)
			}
		}
	}

	return nil
}

// Top returns the current counter top, for use in duty cycle calculation.
// The value returned is (2^resolution - 1), which is the maximum value
// that can be passed to Set().
func (pwm *PWM) Top() uint32 {
	resolution := pwmStates[pwm.num].resolution
	if resolution == 0 {
		resolution = defaultResolution
	}
	return (1 << resolution) - 1
}

// Counter returns the current counter value of the timer.
// This may be useful for debugging.
func (pwm *PWM) Counter() uint32 {
	return pwm.timerValue().Get()
}

// Period returns the current period in nanoseconds.
func (pwm *PWM) Period() uint64 {
	conf := pwm.timerConf().Get()
	dividerReg := (conf & timerDivNumMask) >> timerDivNumPos
	resolution := conf & timerLimMask

	if resolution == 0 || dividerReg == 0 {
		return 0
	}

	// period_ns = (2^resolution * divider_reg / 256) / 80MHz * 1e9
	// period_ns = (2^resolution * divider_reg * 1000) / (80 * (1 << dividerFracBits))
	resolutionValue := uint64(1) << resolution
	return resolutionValue * uint64(dividerReg) * 1000 / (80 << dividerFracBits)
}

// Frequency returns the current PWM frequency in Hz.
func (pwm *PWM) Frequency() uint32 {
	period := pwm.Period()
	if period == 0 {
		return 0
	}
	return uint32(1_000_000_000 / period)
}

// Resolution returns the current duty cycle resolution in bits.
func (pwm *PWM) Resolution() uint8 {
	resolution := pwmStates[pwm.num].resolution
	if resolution == 0 {
		return defaultResolution
	}
	return resolution
}

// SetInverting sets whether to invert the output of this channel.
// Without inverting, a 25% duty cycle would mean the output is high for 25% of
// the time and low for the rest. Inverting flips the output as if a NOT gate
// was placed at the output, meaning that the output would be 25% low and 75%
// high with a duty cycle of 25%.
func (pwm *PWM) SetInverting(channel uint8, inverting bool) {
	if !pwm.isValidChannel(channel) {
		return
	}

	// Get the pin for this channel to configure GPIO matrix inversion
	if !pwmChannels[channel].inUse || pwmChannels[channel].timer != pwm.num {
		return
	}

	pin := pwmChannels[channel].pin

	// Reconfigure the GPIO with inversion setting through GPIO matrix
	// The GPIO matrix FUNC_OUT_SEL_CFG register has an inversion bit (bit 9)
	signal := uint32(ledcHSSignalBase + channel)

	// Get the GPIO function output select register
	outFunc := pin.outFunc()

	if inverting {
		outFunc.Set(signal | gpioMatrixInvertBit)
	} else {
		outFunc.Set(signal)
	}
}

// Enable enables or disables this PWM timer. When disabled (paused), the timer
// stops counting and all channels using this timer will hold their current state.
func (pwm *PWM) Enable(enable bool) {
	timerConf := pwm.timerConf()
	if enable {
		timerConf.ClearBits(timerPauseMask)
	} else {
		timerConf.SetBits(timerPauseMask)
	}
}

// IsEnabled returns true if this PWM timer is running (not paused).
func (pwm *PWM) IsEnabled() bool {
	return (pwm.timerConf().Get() & timerPauseMask) == 0
}

// ResetCounter resets the timer counter to 0. This can be used to
// synchronize multiple PWM timers.
func (pwm *PWM) ResetCounter() {
	timerConf := pwm.timerConf()
	conf := timerConf.Get()
	timerConf.Set(conf | timerRstMask)
	timerConf.Set(conf)
}

// GetPin returns the pin assigned to the given channel, or NoPin if the
// channel is not in use by this PWM peripheral.
func (pwm *PWM) GetPin(channel uint8) Pin {
	if !pwm.isValidChannel(channel) {
		return NoPin
	}
	if !pwmChannels[channel].inUse || pwmChannels[channel].timer != pwm.num {
		return NoPin
	}
	return pwmChannels[channel].pin
}

// Register access helpers

// timerConf returns the configuration register for this timer.
func (pwm *PWM) timerConf() *volatile.Register32 {
	// HSTIMER0_CONF is at offset 0x140, each timer is 0x8 bytes apart
	base := uintptr(unsafe.Pointer(&esp.LEDC.HSTIMER0_CONF))
	return (*volatile.Register32)(unsafe.Pointer(base + uintptr(pwm.num)*0x8))
}

// timerValue returns the value register for this timer.
func (pwm *PWM) timerValue() *volatile.Register32 {
	// HSTIMER0_VALUE is at offset 0x144, each timer is 0x8 bytes apart
	base := uintptr(unsafe.Pointer(&esp.LEDC.HSTIMER0_VALUE))
	return (*volatile.Register32)(unsafe.Pointer(base + uintptr(pwm.num)*0x8))
}

// channelConf0 returns the CONF0 register for the given channel.
func (pwm *PWM) channelConf0(ch uint8) *volatile.Register32 {
	// HSCH0_CONF0 is at offset 0x0, each channel is 0x14 bytes apart
	base := uintptr(unsafe.Pointer(&esp.LEDC.HSCH0_CONF0))
	return (*volatile.Register32)(unsafe.Pointer(base + uintptr(ch)*0x14))
}

// channelHpoint returns the HPOINT register for the given channel.
func (pwm *PWM) channelHpoint(ch uint8) *volatile.Register32 {
	// HSCH0_HPOINT is at offset 0x4, each channel is 0x14 bytes apart
	base := uintptr(unsafe.Pointer(&esp.LEDC.HSCH0_HPOINT))
	return (*volatile.Register32)(unsafe.Pointer(base + uintptr(ch)*0x14))
}

// channelDuty returns the DUTY register for the given channel.
func (pwm *PWM) channelDuty(ch uint8) *volatile.Register32 {
	// HSCH0_DUTY is at offset 0x8, each channel is 0x14 bytes apart
	base := uintptr(unsafe.Pointer(&esp.LEDC.HSCH0_DUTY))
	return (*volatile.Register32)(unsafe.Pointer(base + uintptr(ch)*0x14))
}

// channelConf1 returns the CONF1 register for the given channel.
func (pwm *PWM) channelConf1(ch uint8) *volatile.Register32 {
	// HSCH0_CONF1 is at offset 0xC, each channel is 0x14 bytes apart
	base := uintptr(unsafe.Pointer(&esp.LEDC.HSCH0_CONF1))
	return (*volatile.Register32)(unsafe.Pointer(base + uintptr(ch)*0x14))
}

// channelDutyR returns the DUTY_R (read) register for the given channel.
func (pwm *PWM) channelDutyR(ch uint8) *volatile.Register32 {
	// HSCH0_DUTY_R is at offset 0x10, each channel is 0x14 bytes apart
	base := uintptr(unsafe.Pointer(&esp.LEDC.HSCH0_DUTY_R))
	return (*volatile.Register32)(unsafe.Pointer(base + uintptr(ch)*0x14))
}
