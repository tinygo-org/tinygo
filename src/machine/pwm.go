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

type pwmer interface {
	Configure(config PWMConfig) error
	Channel(pin Pin) (uint8, error)
	SetPeriod(period uint64) error
	Top() uint32
	Counter() uint32
	Period() uint64
	SetInverting(channel uint8, inverting bool)
	Set(channel uint8, value uint32)
	Get(channel uint8) uint32
	SetTop(top uint32)
	SetCounter(ctr uint32)
	Enable(enable bool)
	IsEnabled() bool
}
