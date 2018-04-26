
package machine

// #include "../runtime/runtime_nrf.h"
import "C"

type GPIO struct {
	Pin uint32
}

type GPIOConfig struct {
	Mode uint8
}

const (
	GPIO_INPUT  = iota
	GPIO_OUTPUT
)

func (p GPIO) Configure(config GPIOConfig) {
	C.gpio_cfg(C.uint(p.Pin), C.gpio_mode_t(config.Mode))
}

func (p GPIO) Set(value bool) {
	C.gpio_set(C.uint(p.Pin), C.bool(value))
}
