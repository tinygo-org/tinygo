//go:build esp32c3

package machine

// LEDC PWM for ESP32-C3: 4 timers (PWM0–PWM3), 6 channels per timer; each timer has its own frequency.
// Range: frequency from a few Hz up to ~40 MHz (at 1-bit resolution); duty resolution 1–15 bits
// (higher frequency gives lower resolution). Clock source: APB 80 MHz. Low-speed mode only.
// See ESP-IDF LEDC driver, TRM LED PWM Controller.

// GPIO matrix output signal indices for LEDC (soc/gpio_sig_map.h)
const (
	LEDC_LS_SIG_OUT0_IDX = 45
)

const ledcChannelsC3 = 6

var (
	PWM0 = &LEDCPWM{SigOutBase: LEDC_LS_SIG_OUT0_IDX, NumChannels: ledcChannelsC3, timerNum: 0}
	PWM1 = &LEDCPWM{SigOutBase: LEDC_LS_SIG_OUT0_IDX, NumChannels: ledcChannelsC3, timerNum: 1}
	PWM2 = &LEDCPWM{SigOutBase: LEDC_LS_SIG_OUT0_IDX, NumChannels: ledcChannelsC3, timerNum: 2}
	PWM3 = &LEDCPWM{SigOutBase: LEDC_LS_SIG_OUT0_IDX, NumChannels: ledcChannelsC3, timerNum: 3}
)
