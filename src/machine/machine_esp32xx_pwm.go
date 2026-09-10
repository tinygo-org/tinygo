//go:build esp32 || esp32c3 || esp32s3

// PWM on ESP32 chips uses the LEDC peripheral. LEDC means "LED Control", but it
// works for any PWM job, not only LEDs.
//
// A timer sets the frequency. Channels hang off a timer. Every channel on the
// same timer shares that frequency, but each channel has its own duty. The
// signal reaches a pin through the GPIO matrix, using signal number
// SigOutBase + channel.
//
// This file holds the part that is the same on every chip. Three functions are
// different per chip and live in other files:
//
//	enableClock   turn the LEDC hardware on and pick its clock
//	setTimerConf  program one timer
//	chanOp        set up a channel, change its duty, or invert it
//
// The classic ESP32 has them in machine_esp32_pwm.go. The C3 and S3 have them in
// machine_esp32xx_ls_pwm.go and machine_esp32{c3,s3}_pwm.go.
//
// The order of setup is the one ESP-IDF uses: configure the timer, then the
// channel, then write the duty and tell the hardware to use it. See
// https://docs.espressif.com/projects/esp-idf/en/latest/esp32/api-reference/peripherals/ledc.html

package machine

import "errors"

const ledcApbClock = 80_000000

const ledcDutyFracBits = 4 // DUTY register has 4 fractional bits; write value<<4

const ledcDividerFracBits = 8 // Clock divider register = actual_divider * 256

var errPWMNoChannel = errors.New("pwm: no free channel")

type LEDCPWM struct {
	SigOutBase  uint32 // GPIO matrix signal index for channel 0 (e.g. 73 on S3, 45 on C3)
	NumChannels uint8
	timerNum    uint8 // 0–3: which LEDC timer (frequency) this PWM uses
	dutyRes     uint8
	configured  bool
	channelPin  [8]Pin
}

type ledcChanOp uint8

const (
	ledcChanOpInit      ledcChanOp = iota // initial per-channel setup (timer, enable, HPOINT/DUTY/CONF1, PARA_UP)
	ledcChanOpSetDuty                     // update duty and latch it (DUTY + CONF1 + PARA_UP)
	ledcChanOpSetInvert                   // change idle level (IDLE_LV)
)

func (pwm *LEDCPWM) Configure(config PWMConfig) error {
	// Turn the LEDC hardware on and pick its clock source. The registers have
	// different names on each chip, so this lives in the per-chip file.
	pwm.enableClock()

	period := config.Period
	if period == 0 {
		period = 1_000_000
	}
	freq := uint64(1e9) / period
	dutyRes := uint8(10)
	switch {
	case freq < 100:
		dutyRes = 14
	case freq < 1000:
		dutyRes = 12
	case freq > 100_000:
		dutyRes = 8
	}

	// Timer divider: period_ns = (2^dutyRes * divActual/256) / 80MHz * 1e9 => divReg = divActual<<8.
	divActual := ledcApbClock / (uint32(freq) * (1 << dutyRes))
	if divActual == 0 {
		divActual = 1
	}
	divReg := divActual << ledcDividerFracBits
	if divReg > 0x3ffff {
		return ErrPWMPeriodTooLong
	}

	// Selected timer: resolution, divider, no pause, reset then latch config with PARA_UP.
	pwm.setTimerConf(dutyRes, divReg)

	pwm.dutyRes = dutyRes
	pwm.configured = true
	for i := range pwm.channelPin {
		pwm.channelPin[i] = NoPin
	}
	return nil
}

func (pwm *LEDCPWM) Channel(pin Pin) (uint8, error) {
	if !pwm.configured {
		return 0, errors.New("pwm: not configured")
	}
	if pin == NoPin {
		return 0, ErrInvalidOutputPin
	}
	var ch uint8
	for ch = 0; ch < pwm.NumChannels; ch++ {
		if pwm.channelPin[ch] == NoPin {
			break
		}
	}
	if ch >= pwm.NumChannels {
		return 0, errPWMNoChannel
	}

	pwm.channelPin[ch] = pin
	signal := pwm.SigOutBase + uint32(ch)
	pin.configure(PinConfig{Mode: PinOutput}, signal) // GPIO matrix: pin <- LEDC_LS_SIG_OUTn
	pwm.chanOp(ch, ledcChanOpInit, 0, false)
	return ch, nil
}

func (pwm *LEDCPWM) Set(channel uint8, value uint32) {
	if channel >= pwm.NumChannels {
		return
	}
	top := uint32(1<<pwm.dutyRes) - 1
	if value > top {
		value = top
	}
	dutyVal := value << ledcDutyFracBits
	pwm.chanOp(channel, ledcChanOpSetDuty, dutyVal, false)
}

func (pwm *LEDCPWM) Top() uint32 {
	if !pwm.configured {
		return 0
	}
	return uint32(1<<pwm.dutyRes) - 1
}

func (pwm *LEDCPWM) SetInverting(channel uint8, inverting bool) {
	if channel >= pwm.NumChannels {
		return
	}
	pwm.chanOp(channel, ledcChanOpSetInvert, 0, inverting)
}
