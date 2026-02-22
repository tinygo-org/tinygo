//go:build esp32s3

package machine

// LEDC PWM for ESP32-S3: 4 timers (PWM0–PWM3), 8 channels per timer; each timer has its own frequency.
// Range: frequency from a few Hz up to ~40 MHz (at 1-bit resolution); duty resolution 1–15 bits
// (higher frequency gives lower resolution). Clock source: APB 80 MHz. Low-speed mode only.
// Duty must not equal 2^resolution (overflow risk). See ESP-IDF LEDC driver, TRM LED PWM Controller.

// GPIO matrix output signal indices for LEDC (soc/gpio_sig_map.h)
const (
	LEDC_LS_SIG_OUT0_IDX = 73
)

const ledcChannelsS3 = 8

var (
	PWM0 = &LEDCPWM{SigOutBase: LEDC_LS_SIG_OUT0_IDX, NumChannels: ledcChannelsS3, timerNum: 0}
	PWM1 = &LEDCPWM{SigOutBase: LEDC_LS_SIG_OUT0_IDX, NumChannels: ledcChannelsS3, timerNum: 1}
	PWM2 = &LEDCPWM{SigOutBase: LEDC_LS_SIG_OUT0_IDX, NumChannels: ledcChannelsS3, timerNum: 2}
	PWM3 = &LEDCPWM{SigOutBase: LEDC_LS_SIG_OUT0_IDX, NumChannels: ledcChannelsS3, timerNum: 3}
)
