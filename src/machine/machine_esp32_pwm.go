//go:build esp32

// PWM on the classic ESP32 uses the LEDC peripheral.
//
// LEDC here has two halves: a high-speed one and a low-speed one. This file uses
// the high-speed half. It gives 4 timers (PWM0-PWM3) and 8 channels.
//
// The code that is the same on every ESP32 chip is in machine_esp32xx_pwm.go.
// This file adds only the parts that are different here.
//
// Three things are different from the ESP32-C3 and the ESP32-S3:
//
//   - The clock is switched on in DPORT. The newer chips renamed that block to
//     SYSTEM. This chip also has no LEDC.CONF_CLK_EN bit.
//   - Register names start with HS, for example HSTIMER0_CONF and HSCH0_CONF0.
//   - There is no PARA_UP bit. On the C3 and S3 you set PARA_UP to say "apply my
//     changes now". Here the channel applies them by itself, at the end of the
//     period it is in. That is what makes it "high speed".
//     See the ESP32 TRM, chapter "LED PWM Controller".

package machine

import "device/esp"

// The GPIO matrix can send an internal signal to almost any pin. Each signal has
// a number. High-speed LEDC channel 0 is number 71, and the rest follow on from
// there, so channels 0 to 7 are 71 to 78.
// (From Espressif's soc/gpio_sig_map.h: LEDC_HS_SIG_OUT0_IDX.)
const LEDC_HS_SIG_OUT0_IDX = 71

const ledcChannelsESP32 = 8

var (
	PWM0 = &LEDCPWM{SigOutBase: LEDC_HS_SIG_OUT0_IDX, NumChannels: ledcChannelsESP32, timerNum: 0}
	PWM1 = &LEDCPWM{SigOutBase: LEDC_HS_SIG_OUT0_IDX, NumChannels: ledcChannelsESP32, timerNum: 1}
	PWM2 = &LEDCPWM{SigOutBase: LEDC_HS_SIG_OUT0_IDX, NumChannels: ledcChannelsESP32, timerNum: 2}
	PWM3 = &LEDCPWM{SigOutBase: LEDC_HS_SIG_OUT0_IDX, NumChannels: ledcChannelsESP32, timerNum: 3}
)

// enableClock turns the LEDC hardware on and picks its clock.
func (pwm *LEDCPWM) enableClock() {
	// Every peripheral starts switched off, to save power. Turn the LEDC clock
	// on, then take it out of reset. These registers are in DPORT on this chip.
	// The C3 and S3 call the same block SYSTEM.
	esp.DPORT.SetPERIP_RST_EN_LEDC_RST(1)
	esp.DPORT.SetPERIP_CLK_EN_LEDC_CLK_EN(1)
	esp.DPORT.SetPERIP_RST_EN_LEDC_RST(0)

	// Use APB_CLK, which runs at 80MHz. This chip has no CONF_CLK_EN bit, so
	// there is nothing more to switch on.
	esp.LEDC.SetCONF_APB_CLK_SEL(1)
}

// setTimerConf writes the resolution and the divider that Configure worked out
// into one timer.
//
// Watch TICK_SEL. It is 1 here, not 0. On this chip:
//
//	1 = APB_CLK, 80MHz   <- what we want
//	0 = REF_TICK, 1MHz
//
// The low-speed timers on the C3 and S3 use the opposite meaning, so their code
// writes 0. Writing 0 here would make every frequency 80 times too slow.
//
// These timers have no PARA_UP bit. The short reset pulse at the end is what
// makes the new values take effect.
func (pwm *LEDCPWM) setTimerConf(dutyRes uint8, divReg uint32) {
	switch pwm.timerNum {
	case 0:
		esp.LEDC.SetHSTIMER0_CONF_DUTY_RES(uint32(dutyRes))
		esp.LEDC.SetHSTIMER0_CONF_DIV_NUM(divReg)
		esp.LEDC.SetHSTIMER0_CONF_TICK_SEL(1)
		esp.LEDC.SetHSTIMER0_CONF_PAUSE(0)
		esp.LEDC.SetHSTIMER0_CONF_RST(1)
		esp.LEDC.SetHSTIMER0_CONF_RST(0)
	case 1:
		esp.LEDC.SetHSTIMER1_CONF_DUTY_RES(uint32(dutyRes))
		esp.LEDC.SetHSTIMER1_CONF_DIV_NUM(divReg)
		esp.LEDC.SetHSTIMER1_CONF_TICK_SEL(1)
		esp.LEDC.SetHSTIMER1_CONF_PAUSE(0)
		esp.LEDC.SetHSTIMER1_CONF_RST(1)
		esp.LEDC.SetHSTIMER1_CONF_RST(0)
	case 2:
		esp.LEDC.SetHSTIMER2_CONF_DUTY_RES(uint32(dutyRes))
		esp.LEDC.SetHSTIMER2_CONF_DIV_NUM(divReg)
		esp.LEDC.SetHSTIMER2_CONF_TICK_SEL(1)
		esp.LEDC.SetHSTIMER2_CONF_PAUSE(0)
		esp.LEDC.SetHSTIMER2_CONF_RST(1)
		esp.LEDC.SetHSTIMER2_CONF_RST(0)
	case 3:
		esp.LEDC.SetHSTIMER3_CONF_DUTY_RES(uint32(dutyRes))
		esp.LEDC.SetHSTIMER3_CONF_DIV_NUM(divReg)
		esp.LEDC.SetHSTIMER3_CONF_TICK_SEL(1)
		esp.LEDC.SetHSTIMER3_CONF_PAUSE(0)
		esp.LEDC.SetHSTIMER3_CONF_RST(1)
		esp.LEDC.SetHSTIMER3_CONF_RST(0)
	}
}

// chanOp does the work for one channel, numbered 0 to 7. It either sets the
// channel up, changes its duty, or flips its idle level.
//
// Each channel has its own registers, and each register has its own generated
// setter function. There is no array to index, so this has to be a long switch.
// The C3 and S3 files have the same shape.
//
// DUTY_SCALE is set to 0. LEDC can fade slowly from one duty to the next; 0
// turns that off, so the duty changes in a single step.
func (pwm *LEDCPWM) chanOp(ch uint8, op ledcChanOp, duty uint32, inverting bool) {
	invVal := uint32(0)
	if inverting {
		invVal = 1
	}
	switch ch {
	case 0:
		switch op {
		case ledcChanOpInit:
			esp.LEDC.SetHSCH0_CONF0_TIMER_SEL(uint32(pwm.timerNum))
			esp.LEDC.SetHSCH0_CONF0_SIG_OUT_EN(1)
			esp.LEDC.SetHSCH0_CONF0_IDLE_LV(0)
			esp.LEDC.SetHSCH0_HPOINT_HPOINT(0)
			esp.LEDC.SetHSCH0_DUTY_DUTY(0)
			esp.LEDC.SetHSCH0_CONF1_DUTY_SCALE(0)
			esp.LEDC.SetHSCH0_CONF1_DUTY_CYCLE(1)
			esp.LEDC.SetHSCH0_CONF1_DUTY_NUM(1)
			esp.LEDC.SetHSCH0_CONF1_DUTY_INC(1)
			esp.LEDC.SetHSCH0_CONF1_DUTY_START(1)
		case ledcChanOpSetDuty:
			esp.LEDC.SetHSCH0_DUTY_DUTY(duty)
			esp.LEDC.SetHSCH0_CONF1_DUTY_CYCLE(1)
			esp.LEDC.SetHSCH0_CONF1_DUTY_NUM(1)
			esp.LEDC.SetHSCH0_CONF1_DUTY_INC(1)
			esp.LEDC.SetHSCH0_CONF1_DUTY_START(1)
			esp.LEDC.SetHSCH0_CONF0_SIG_OUT_EN(1)
		case ledcChanOpSetInvert:
			esp.LEDC.SetHSCH0_CONF0_IDLE_LV(invVal)
		}
	case 1:
		switch op {
		case ledcChanOpInit:
			esp.LEDC.SetHSCH1_CONF0_TIMER_SEL(uint32(pwm.timerNum))
			esp.LEDC.SetHSCH1_CONF0_SIG_OUT_EN(1)
			esp.LEDC.SetHSCH1_CONF0_IDLE_LV(0)
			esp.LEDC.SetHSCH1_HPOINT_HPOINT(0)
			esp.LEDC.SetHSCH1_DUTY_DUTY(0)
			esp.LEDC.SetHSCH1_CONF1_DUTY_SCALE(0)
			esp.LEDC.SetHSCH1_CONF1_DUTY_CYCLE(1)
			esp.LEDC.SetHSCH1_CONF1_DUTY_NUM(1)
			esp.LEDC.SetHSCH1_CONF1_DUTY_INC(1)
			esp.LEDC.SetHSCH1_CONF1_DUTY_START(1)
		case ledcChanOpSetDuty:
			esp.LEDC.SetHSCH1_DUTY_DUTY(duty)
			esp.LEDC.SetHSCH1_CONF1_DUTY_CYCLE(1)
			esp.LEDC.SetHSCH1_CONF1_DUTY_NUM(1)
			esp.LEDC.SetHSCH1_CONF1_DUTY_INC(1)
			esp.LEDC.SetHSCH1_CONF1_DUTY_START(1)
			esp.LEDC.SetHSCH1_CONF0_SIG_OUT_EN(1)
		case ledcChanOpSetInvert:
			esp.LEDC.SetHSCH1_CONF0_IDLE_LV(invVal)
		}
	case 2:
		switch op {
		case ledcChanOpInit:
			esp.LEDC.SetHSCH2_CONF0_TIMER_SEL(uint32(pwm.timerNum))
			esp.LEDC.SetHSCH2_CONF0_SIG_OUT_EN(1)
			esp.LEDC.SetHSCH2_CONF0_IDLE_LV(0)
			esp.LEDC.SetHSCH2_HPOINT_HPOINT(0)
			esp.LEDC.SetHSCH2_DUTY_DUTY(0)
			esp.LEDC.SetHSCH2_CONF1_DUTY_SCALE(0)
			esp.LEDC.SetHSCH2_CONF1_DUTY_CYCLE(1)
			esp.LEDC.SetHSCH2_CONF1_DUTY_NUM(1)
			esp.LEDC.SetHSCH2_CONF1_DUTY_INC(1)
			esp.LEDC.SetHSCH2_CONF1_DUTY_START(1)
		case ledcChanOpSetDuty:
			esp.LEDC.SetHSCH2_DUTY_DUTY(duty)
			esp.LEDC.SetHSCH2_CONF1_DUTY_CYCLE(1)
			esp.LEDC.SetHSCH2_CONF1_DUTY_NUM(1)
			esp.LEDC.SetHSCH2_CONF1_DUTY_INC(1)
			esp.LEDC.SetHSCH2_CONF1_DUTY_START(1)
			esp.LEDC.SetHSCH2_CONF0_SIG_OUT_EN(1)
		case ledcChanOpSetInvert:
			esp.LEDC.SetHSCH2_CONF0_IDLE_LV(invVal)
		}
	case 3:
		switch op {
		case ledcChanOpInit:
			esp.LEDC.SetHSCH3_CONF0_TIMER_SEL(uint32(pwm.timerNum))
			esp.LEDC.SetHSCH3_CONF0_SIG_OUT_EN(1)
			esp.LEDC.SetHSCH3_CONF0_IDLE_LV(0)
			esp.LEDC.SetHSCH3_HPOINT_HPOINT(0)
			esp.LEDC.SetHSCH3_DUTY_DUTY(0)
			esp.LEDC.SetHSCH3_CONF1_DUTY_SCALE(0)
			esp.LEDC.SetHSCH3_CONF1_DUTY_CYCLE(1)
			esp.LEDC.SetHSCH3_CONF1_DUTY_NUM(1)
			esp.LEDC.SetHSCH3_CONF1_DUTY_INC(1)
			esp.LEDC.SetHSCH3_CONF1_DUTY_START(1)
		case ledcChanOpSetDuty:
			esp.LEDC.SetHSCH3_DUTY_DUTY(duty)
			esp.LEDC.SetHSCH3_CONF1_DUTY_CYCLE(1)
			esp.LEDC.SetHSCH3_CONF1_DUTY_NUM(1)
			esp.LEDC.SetHSCH3_CONF1_DUTY_INC(1)
			esp.LEDC.SetHSCH3_CONF1_DUTY_START(1)
			esp.LEDC.SetHSCH3_CONF0_SIG_OUT_EN(1)
		case ledcChanOpSetInvert:
			esp.LEDC.SetHSCH3_CONF0_IDLE_LV(invVal)
		}
	case 4:
		switch op {
		case ledcChanOpInit:
			esp.LEDC.SetHSCH4_CONF0_TIMER_SEL(uint32(pwm.timerNum))
			esp.LEDC.SetHSCH4_CONF0_SIG_OUT_EN(1)
			esp.LEDC.SetHSCH4_CONF0_IDLE_LV(0)
			esp.LEDC.SetHSCH4_HPOINT_HPOINT(0)
			esp.LEDC.SetHSCH4_DUTY_DUTY(0)
			esp.LEDC.SetHSCH4_CONF1_DUTY_SCALE(0)
			esp.LEDC.SetHSCH4_CONF1_DUTY_CYCLE(1)
			esp.LEDC.SetHSCH4_CONF1_DUTY_NUM(1)
			esp.LEDC.SetHSCH4_CONF1_DUTY_INC(1)
			esp.LEDC.SetHSCH4_CONF1_DUTY_START(1)
		case ledcChanOpSetDuty:
			esp.LEDC.SetHSCH4_DUTY_DUTY(duty)
			esp.LEDC.SetHSCH4_CONF1_DUTY_CYCLE(1)
			esp.LEDC.SetHSCH4_CONF1_DUTY_NUM(1)
			esp.LEDC.SetHSCH4_CONF1_DUTY_INC(1)
			esp.LEDC.SetHSCH4_CONF1_DUTY_START(1)
			esp.LEDC.SetHSCH4_CONF0_SIG_OUT_EN(1)
		case ledcChanOpSetInvert:
			esp.LEDC.SetHSCH4_CONF0_IDLE_LV(invVal)
		}
	case 5:
		switch op {
		case ledcChanOpInit:
			esp.LEDC.SetHSCH5_CONF0_TIMER_SEL(uint32(pwm.timerNum))
			esp.LEDC.SetHSCH5_CONF0_SIG_OUT_EN(1)
			esp.LEDC.SetHSCH5_CONF0_IDLE_LV(0)
			esp.LEDC.SetHSCH5_HPOINT_HPOINT(0)
			esp.LEDC.SetHSCH5_DUTY_DUTY(0)
			esp.LEDC.SetHSCH5_CONF1_DUTY_SCALE(0)
			esp.LEDC.SetHSCH5_CONF1_DUTY_CYCLE(1)
			esp.LEDC.SetHSCH5_CONF1_DUTY_NUM(1)
			esp.LEDC.SetHSCH5_CONF1_DUTY_INC(1)
			esp.LEDC.SetHSCH5_CONF1_DUTY_START(1)
		case ledcChanOpSetDuty:
			esp.LEDC.SetHSCH5_DUTY_DUTY(duty)
			esp.LEDC.SetHSCH5_CONF1_DUTY_CYCLE(1)
			esp.LEDC.SetHSCH5_CONF1_DUTY_NUM(1)
			esp.LEDC.SetHSCH5_CONF1_DUTY_INC(1)
			esp.LEDC.SetHSCH5_CONF1_DUTY_START(1)
			esp.LEDC.SetHSCH5_CONF0_SIG_OUT_EN(1)
		case ledcChanOpSetInvert:
			esp.LEDC.SetHSCH5_CONF0_IDLE_LV(invVal)
		}
	case 6:
		switch op {
		case ledcChanOpInit:
			esp.LEDC.SetHSCH6_CONF0_TIMER_SEL(uint32(pwm.timerNum))
			esp.LEDC.SetHSCH6_CONF0_SIG_OUT_EN(1)
			esp.LEDC.SetHSCH6_CONF0_IDLE_LV(0)
			esp.LEDC.SetHSCH6_HPOINT_HPOINT(0)
			esp.LEDC.SetHSCH6_DUTY_DUTY(0)
			esp.LEDC.SetHSCH6_CONF1_DUTY_SCALE(0)
			esp.LEDC.SetHSCH6_CONF1_DUTY_CYCLE(1)
			esp.LEDC.SetHSCH6_CONF1_DUTY_NUM(1)
			esp.LEDC.SetHSCH6_CONF1_DUTY_INC(1)
			esp.LEDC.SetHSCH6_CONF1_DUTY_START(1)
		case ledcChanOpSetDuty:
			esp.LEDC.SetHSCH6_DUTY_DUTY(duty)
			esp.LEDC.SetHSCH6_CONF1_DUTY_CYCLE(1)
			esp.LEDC.SetHSCH6_CONF1_DUTY_NUM(1)
			esp.LEDC.SetHSCH6_CONF1_DUTY_INC(1)
			esp.LEDC.SetHSCH6_CONF1_DUTY_START(1)
			esp.LEDC.SetHSCH6_CONF0_SIG_OUT_EN(1)
		case ledcChanOpSetInvert:
			esp.LEDC.SetHSCH6_CONF0_IDLE_LV(invVal)
		}
	case 7:
		switch op {
		case ledcChanOpInit:
			esp.LEDC.SetHSCH7_CONF0_TIMER_SEL(uint32(pwm.timerNum))
			esp.LEDC.SetHSCH7_CONF0_SIG_OUT_EN(1)
			esp.LEDC.SetHSCH7_CONF0_IDLE_LV(0)
			esp.LEDC.SetHSCH7_HPOINT_HPOINT(0)
			esp.LEDC.SetHSCH7_DUTY_DUTY(0)
			esp.LEDC.SetHSCH7_CONF1_DUTY_SCALE(0)
			esp.LEDC.SetHSCH7_CONF1_DUTY_CYCLE(1)
			esp.LEDC.SetHSCH7_CONF1_DUTY_NUM(1)
			esp.LEDC.SetHSCH7_CONF1_DUTY_INC(1)
			esp.LEDC.SetHSCH7_CONF1_DUTY_START(1)
		case ledcChanOpSetDuty:
			esp.LEDC.SetHSCH7_DUTY_DUTY(duty)
			esp.LEDC.SetHSCH7_CONF1_DUTY_CYCLE(1)
			esp.LEDC.SetHSCH7_CONF1_DUTY_NUM(1)
			esp.LEDC.SetHSCH7_CONF1_DUTY_INC(1)
			esp.LEDC.SetHSCH7_CONF1_DUTY_START(1)
			esp.LEDC.SetHSCH7_CONF0_SIG_OUT_EN(1)
		case ledcChanOpSetInvert:
			esp.LEDC.SetHSCH7_CONF0_IDLE_LV(invVal)
		}
	}
}
