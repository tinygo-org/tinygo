package machine

import "errors"

var (
	ErrPWMPeriodTooLong = errors.New("pwm: period too long")
)

// PWMConfig allows setting some configuration while configuring a PWM
// peripheral. A zero PWMConfig is ready to use for simple applications such as
// dimming LEDs.
type PWMConfig struct {
	// PWM period in nanosecond. Leaving this zero will pick a reasonable period
	// value for use with LEDs.
	// If you want to configure a frequency instead of a period, you can use the
	// following formula to calculate a period from a frequency:
	//
	//     period = 1e9 / frequency
	//
	Period uint64
}
