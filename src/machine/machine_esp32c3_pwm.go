//go:build esp32c3

package machine

import "device/esp"

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

// chanOp implements LEDC low-speed channel ops for ESP32-C3 (channels 0–5 only).
func (pwm *LEDCPWM) chanOp(ch uint8, op ledcChanOp, duty uint32, inverting bool) {
	invVal := uint32(0)
	if inverting {
		invVal = 1
	}
	switch ch {
	case 0:
		switch op {
		case ledcChanOpInit:
			esp.LEDC.SetCH0_CONF0_TIMER_SEL(uint32(pwm.timerNum))
			esp.LEDC.SetCH0_CONF0_SIG_OUT_EN(1)
			esp.LEDC.SetCH0_CONF0_IDLE_LV(0)
			esp.LEDC.SetCH0_HPOINT_HPOINT(0)
			esp.LEDC.SetCH0_DUTY_DUTY(0)
			esp.LEDC.SetCH0_CONF1_DUTY_CYCLE(1)
			esp.LEDC.SetCH0_CONF1_DUTY_NUM(1)
			esp.LEDC.SetCH0_CONF1_DUTY_INC(1)
			esp.LEDC.SetCH0_CONF1_DUTY_START(1)
			esp.LEDC.SetCH0_CONF0_PARA_UP(1)
		case ledcChanOpSetDuty:
			esp.LEDC.SetCH0_DUTY_DUTY(duty)
			esp.LEDC.SetCH0_CONF1_DUTY_CYCLE(1)
			esp.LEDC.SetCH0_CONF1_DUTY_NUM(1)
			esp.LEDC.SetCH0_CONF1_DUTY_INC(1)
			esp.LEDC.SetCH0_CONF1_DUTY_START(1)
			esp.LEDC.SetCH0_CONF0_SIG_OUT_EN(1)
			esp.LEDC.SetCH0_CONF0_PARA_UP(1)
		case ledcChanOpSetInvert:
			esp.LEDC.SetCH0_CONF0_IDLE_LV(invVal)
		}
	case 1:
		switch op {
		case ledcChanOpInit:
			esp.LEDC.SetCH1_CONF0_TIMER_SEL(uint32(pwm.timerNum))
			esp.LEDC.SetCH1_CONF0_SIG_OUT_EN(1)
			esp.LEDC.SetCH1_CONF0_IDLE_LV(0)
			esp.LEDC.SetCH1_HPOINT_HPOINT(0)
			esp.LEDC.SetCH1_DUTY_DUTY(0)
			esp.LEDC.SetCH1_CONF1_DUTY_CYCLE(1)
			esp.LEDC.SetCH1_CONF1_DUTY_NUM(1)
			esp.LEDC.SetCH1_CONF1_DUTY_INC(1)
			esp.LEDC.SetCH1_CONF1_DUTY_START(1)
			esp.LEDC.SetCH1_CONF0_PARA_UP(1)
		case ledcChanOpSetDuty:
			esp.LEDC.SetCH1_DUTY_DUTY(duty)
			esp.LEDC.SetCH1_CONF1_DUTY_CYCLE(1)
			esp.LEDC.SetCH1_CONF1_DUTY_NUM(1)
			esp.LEDC.SetCH1_CONF1_DUTY_INC(1)
			esp.LEDC.SetCH1_CONF1_DUTY_START(1)
			esp.LEDC.SetCH1_CONF0_SIG_OUT_EN(1)
			esp.LEDC.SetCH1_CONF0_PARA_UP(1)
		case ledcChanOpSetInvert:
			esp.LEDC.SetCH1_CONF0_IDLE_LV(invVal)
		}
	case 2:
		switch op {
		case ledcChanOpInit:
			esp.LEDC.SetCH2_CONF0_TIMER_SEL(uint32(pwm.timerNum))
			esp.LEDC.SetCH2_CONF0_SIG_OUT_EN(1)
			esp.LEDC.SetCH2_CONF0_IDLE_LV(0)
			esp.LEDC.SetCH2_HPOINT_HPOINT(0)
			esp.LEDC.SetCH2_DUTY_DUTY(0)
			esp.LEDC.SetCH2_CONF1_DUTY_CYCLE(1)
			esp.LEDC.SetCH2_CONF1_DUTY_NUM(1)
			esp.LEDC.SetCH2_CONF1_DUTY_INC(1)
			esp.LEDC.SetCH2_CONF1_DUTY_START(1)
			esp.LEDC.SetCH2_CONF0_PARA_UP(1)
		case ledcChanOpSetDuty:
			esp.LEDC.SetCH2_DUTY_DUTY(duty)
			esp.LEDC.SetCH2_CONF1_DUTY_CYCLE(1)
			esp.LEDC.SetCH2_CONF1_DUTY_NUM(1)
			esp.LEDC.SetCH2_CONF1_DUTY_INC(1)
			esp.LEDC.SetCH2_CONF1_DUTY_START(1)
			esp.LEDC.SetCH2_CONF0_SIG_OUT_EN(1)
			esp.LEDC.SetCH2_CONF0_PARA_UP(1)
		case ledcChanOpSetInvert:
			esp.LEDC.SetCH2_CONF0_IDLE_LV(invVal)
		}
	case 3:
		switch op {
		case ledcChanOpInit:
			esp.LEDC.SetCH3_CONF0_TIMER_SEL(uint32(pwm.timerNum))
			esp.LEDC.SetCH3_CONF0_SIG_OUT_EN(1)
			esp.LEDC.SetCH3_CONF0_IDLE_LV(0)
			esp.LEDC.SetCH3_HPOINT_HPOINT(0)
			esp.LEDC.SetCH3_DUTY_DUTY(0)
			esp.LEDC.SetCH3_CONF1_DUTY_CYCLE(1)
			esp.LEDC.SetCH3_CONF1_DUTY_NUM(1)
			esp.LEDC.SetCH3_CONF1_DUTY_INC(1)
			esp.LEDC.SetCH3_CONF1_DUTY_START(1)
			esp.LEDC.SetCH3_CONF0_PARA_UP(1)
		case ledcChanOpSetDuty:
			esp.LEDC.SetCH3_DUTY_DUTY(duty)
			esp.LEDC.SetCH3_CONF1_DUTY_CYCLE(1)
			esp.LEDC.SetCH3_CONF1_DUTY_NUM(1)
			esp.LEDC.SetCH3_CONF1_DUTY_INC(1)
			esp.LEDC.SetCH3_CONF1_DUTY_START(1)
			esp.LEDC.SetCH3_CONF0_SIG_OUT_EN(1)
			esp.LEDC.SetCH3_CONF0_PARA_UP(1)
		case ledcChanOpSetInvert:
			esp.LEDC.SetCH3_CONF0_IDLE_LV(invVal)
		}
	case 4:
		switch op {
		case ledcChanOpInit:
			esp.LEDC.SetCH4_CONF0_TIMER_SEL(uint32(pwm.timerNum))
			esp.LEDC.SetCH4_CONF0_SIG_OUT_EN(1)
			esp.LEDC.SetCH4_CONF0_IDLE_LV(0)
			esp.LEDC.SetCH4_HPOINT_HPOINT(0)
			esp.LEDC.SetCH4_DUTY_DUTY(0)
			esp.LEDC.SetCH4_CONF1_DUTY_CYCLE(1)
			esp.LEDC.SetCH4_CONF1_DUTY_NUM(1)
			esp.LEDC.SetCH4_CONF1_DUTY_INC(1)
			esp.LEDC.SetCH4_CONF1_DUTY_START(1)
			esp.LEDC.SetCH4_CONF0_PARA_UP(1)
		case ledcChanOpSetDuty:
			esp.LEDC.SetCH4_DUTY_DUTY(duty)
			esp.LEDC.SetCH4_CONF1_DUTY_CYCLE(1)
			esp.LEDC.SetCH4_CONF1_DUTY_NUM(1)
			esp.LEDC.SetCH4_CONF1_DUTY_INC(1)
			esp.LEDC.SetCH4_CONF1_DUTY_START(1)
			esp.LEDC.SetCH4_CONF0_SIG_OUT_EN(1)
			esp.LEDC.SetCH4_CONF0_PARA_UP(1)
		case ledcChanOpSetInvert:
			esp.LEDC.SetCH4_CONF0_IDLE_LV(invVal)
		}
	case 5:
		switch op {
		case ledcChanOpInit:
			esp.LEDC.SetCH5_CONF0_TIMER_SEL(uint32(pwm.timerNum))
			esp.LEDC.SetCH5_CONF0_SIG_OUT_EN(1)
			esp.LEDC.SetCH5_CONF0_IDLE_LV(0)
			esp.LEDC.SetCH5_HPOINT_HPOINT(0)
			esp.LEDC.SetCH5_DUTY_DUTY(0)
			esp.LEDC.SetCH5_CONF1_DUTY_CYCLE(1)
			esp.LEDC.SetCH5_CONF1_DUTY_NUM(1)
			esp.LEDC.SetCH5_CONF1_DUTY_INC(1)
			esp.LEDC.SetCH5_CONF1_DUTY_START(1)
			esp.LEDC.SetCH5_CONF0_PARA_UP(1)
		case ledcChanOpSetDuty:
			esp.LEDC.SetCH5_DUTY_DUTY(duty)
			esp.LEDC.SetCH5_CONF1_DUTY_CYCLE(1)
			esp.LEDC.SetCH5_CONF1_DUTY_NUM(1)
			esp.LEDC.SetCH5_CONF1_DUTY_INC(1)
			esp.LEDC.SetCH5_CONF1_DUTY_START(1)
			esp.LEDC.SetCH5_CONF0_SIG_OUT_EN(1)
			esp.LEDC.SetCH5_CONF0_PARA_UP(1)
		case ledcChanOpSetInvert:
			esp.LEDC.SetCH5_CONF0_IDLE_LV(invVal)
		}
	}
}
