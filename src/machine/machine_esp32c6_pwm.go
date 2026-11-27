//go:build esp32c6

package machine

import "device/esp"

// LEDC on the ESP32-C6 has 4 timers and 6 low-speed channels. The clock comes
// from PCR and fade settings live in gamma RAM. See ESP-IDF hal/esp32c6/include/hal/ledc_ll.h.

// LEDC_LS_SIG_OUT0_IDX is from ESP-IDF soc/esp32c6/include/soc/gpio_sig_map.h.
const LEDC_LS_SIG_OUT0_IDX = 0

const ledcChannelsC6 = 6

var (
	PWM0 = &LEDCPWM{SigOutBase: LEDC_LS_SIG_OUT0_IDX, NumChannels: ledcChannelsC6, timerNum: 0}
	PWM1 = &LEDCPWM{SigOutBase: LEDC_LS_SIG_OUT0_IDX, NumChannels: ledcChannelsC6, timerNum: 1}
	PWM2 = &LEDCPWM{SigOutBase: LEDC_LS_SIG_OUT0_IDX, NumChannels: ledcChannelsC6, timerNum: 2}
	PWM3 = &LEDCPWM{SigOutBase: LEDC_LS_SIG_OUT0_IDX, NumChannels: ledcChannelsC6, timerNum: 3}
)

// enableClock turns the LEDC hardware on and picks the 80 MHz PLL clock.
func (pwm *LEDCPWM) enableClock() {
	// The reset clears every timer and channel, so it runs only once.
	if !ledcStarted {
		esp.PCR.SetLEDC_CONF_LEDC_CLK_EN(1)
		esp.PCR.SetLEDC_CONF_LEDC_RST_EN(1)
		esp.PCR.SetLEDC_CONF_LEDC_RST_EN(0)
		ledcStarted = true
	}
	esp.PCR.SetLEDC_SCLK_CONF_LEDC_SCLK_SEL(1)
	esp.PCR.SetLEDC_SCLK_CONF_LEDC_SCLK_EN(1)
	esp.LEDC.SetCONF_CLK_EN(1)
}

func (pwm *LEDCPWM) setTimerConf(dutyRes uint8, divReg uint32) {
	switch pwm.timerNum {
	case 0:
		esp.LEDC.SetTIMER0_CONF_DUTY_RES(uint32(dutyRes))
		esp.LEDC.SetTIMER0_CONF_CLK_DIV(divReg)
		esp.LEDC.SetTIMER0_CONF_TICK_SEL(0)
		esp.LEDC.SetTIMER0_CONF_PAUSE(0)
		esp.LEDC.SetTIMER0_CONF_RST(1)
		esp.LEDC.SetTIMER0_CONF_RST(0)
		esp.LEDC.SetTIMER0_CONF_PARA_UP(1)
	case 1:
		esp.LEDC.SetTIMER1_CONF_DUTY_RES(uint32(dutyRes))
		esp.LEDC.SetTIMER1_CONF_CLK_DIV(divReg)
		esp.LEDC.SetTIMER1_CONF_TICK_SEL(0)
		esp.LEDC.SetTIMER1_CONF_PAUSE(0)
		esp.LEDC.SetTIMER1_CONF_RST(1)
		esp.LEDC.SetTIMER1_CONF_RST(0)
		esp.LEDC.SetTIMER1_CONF_PARA_UP(1)
	case 2:
		esp.LEDC.SetTIMER2_CONF_DUTY_RES(uint32(dutyRes))
		esp.LEDC.SetTIMER2_CONF_CLK_DIV(divReg)
		esp.LEDC.SetTIMER2_CONF_TICK_SEL(0)
		esp.LEDC.SetTIMER2_CONF_PAUSE(0)
		esp.LEDC.SetTIMER2_CONF_RST(1)
		esp.LEDC.SetTIMER2_CONF_RST(0)
		esp.LEDC.SetTIMER2_CONF_PARA_UP(1)
	case 3:
		esp.LEDC.SetTIMER3_CONF_DUTY_RES(uint32(dutyRes))
		esp.LEDC.SetTIMER3_CONF_CLK_DIV(divReg)
		esp.LEDC.SetTIMER3_CONF_TICK_SEL(0)
		esp.LEDC.SetTIMER3_CONF_PAUSE(0)
		esp.LEDC.SetTIMER3_CONF_RST(1)
		esp.LEDC.SetTIMER3_CONF_RST(0)
		esp.LEDC.SetTIMER3_CONF_PARA_UP(1)
	}
}

// ledcGammaSync waits for gamma RAM writes to cross clock domains.
// ESP-IDF ledc_ll_set_fade_param_range waits 5 us.
func ledcGammaSync() {
	end := nanotime() + 5000
	for nanotime() < end {
	}
}

// chanDisable stops a channel from driving its pin.
func chanDisable(ch uint8) {
	switch ch {
	case 0:
		esp.LEDC.SetCH0_CONF0_SIG_OUT_EN(0)
		esp.LEDC.SetCH0_CONF0_PARA_UP(1)
	case 1:
		esp.LEDC.SetCH1_CONF0_SIG_OUT_EN(0)
		esp.LEDC.SetCH1_CONF0_PARA_UP(1)
	case 2:
		esp.LEDC.SetCH2_CONF0_SIG_OUT_EN(0)
		esp.LEDC.SetCH2_CONF0_PARA_UP(1)
	case 3:
		esp.LEDC.SetCH3_CONF0_SIG_OUT_EN(0)
		esp.LEDC.SetCH3_CONF0_PARA_UP(1)
	case 4:
		esp.LEDC.SetCH4_CONF0_SIG_OUT_EN(0)
		esp.LEDC.SetCH4_CONF0_PARA_UP(1)
	case 5:
		esp.LEDC.SetCH5_CONF0_SIG_OUT_EN(0)
		esp.LEDC.SetCH5_CONF0_PARA_UP(1)
	}
}

// chanOp sets up a channel or changes its duty.
// Init writes one gamma entry with no fade steps, as ESP-IDF ledc_duty_config does.
func (pwm *LEDCPWM) chanOp(ch uint8, op ledcChanOp, duty uint32) {
	switch ch {
	case 0:
		switch op {
		case ledcChanOpInit:
			esp.LEDC.SetCH0_CONF0_TIMER_SEL(uint32(pwm.timerNum))
			esp.LEDC.SetCH0_CONF0_IDLE_LV(0)
			esp.LEDC.SetCH0_HPOINT_HPOINT(0)
			esp.LEDC.SetCH0_DUTY_DUTY(0)
			esp.LEDC.SetCH0_GAMMA_WR_ADDR_CH_GAMMA_WR_ADDR(0)
			ledcGammaSync()
			esp.LEDC.SetCH0_GAMMA_WR_CH_GAMMA_DUTY_INC(1)
			esp.LEDC.SetCH0_GAMMA_WR_CH_GAMMA_DUTY_CYCLE(1)
			esp.LEDC.SetCH0_GAMMA_WR_CH_GAMMA_SCALE(0)
			esp.LEDC.SetCH0_GAMMA_WR_CH_GAMMA_DUTY_NUM(0)
			esp.LEDC.SetCH0_GAMMA_WR_ADDR_CH_GAMMA_WR_ADDR(0)
			ledcGammaSync()
			esp.LEDC.SetCH0_GAMMA_CONF_CH_GAMMA_ENTRY_NUM(1)
			esp.LEDC.SetCH0_CONF1_DUTY_START(1)
			esp.LEDC.SetCH0_CONF0_SIG_OUT_EN(1)
			esp.LEDC.SetCH0_CONF0_PARA_UP(1)
		case ledcChanOpSetDuty:
			esp.LEDC.SetCH0_DUTY_DUTY(duty)
			esp.LEDC.SetCH0_CONF1_DUTY_START(1)
			esp.LEDC.SetCH0_CONF0_SIG_OUT_EN(1)
			esp.LEDC.SetCH0_CONF0_PARA_UP(1)
		}
	case 1:
		switch op {
		case ledcChanOpInit:
			esp.LEDC.SetCH1_CONF0_TIMER_SEL(uint32(pwm.timerNum))
			esp.LEDC.SetCH1_CONF0_IDLE_LV(0)
			esp.LEDC.SetCH1_HPOINT_HPOINT(0)
			esp.LEDC.SetCH1_DUTY_DUTY(0)
			esp.LEDC.SetCH1_GAMMA_WR_ADDR_CH_GAMMA_WR_ADDR(0)
			ledcGammaSync()
			esp.LEDC.SetCH1_GAMMA_WR_CH_GAMMA_DUTY_INC(1)
			esp.LEDC.SetCH1_GAMMA_WR_CH_GAMMA_DUTY_CYCLE(1)
			esp.LEDC.SetCH1_GAMMA_WR_CH_GAMMA_SCALE(0)
			esp.LEDC.SetCH1_GAMMA_WR_CH_GAMMA_DUTY_NUM(0)
			esp.LEDC.SetCH1_GAMMA_WR_ADDR_CH_GAMMA_WR_ADDR(0)
			ledcGammaSync()
			esp.LEDC.SetCH1_GAMMA_CONF_CH_GAMMA_ENTRY_NUM(1)
			esp.LEDC.SetCH1_CONF1_DUTY_START(1)
			esp.LEDC.SetCH1_CONF0_SIG_OUT_EN(1)
			esp.LEDC.SetCH1_CONF0_PARA_UP(1)
		case ledcChanOpSetDuty:
			esp.LEDC.SetCH1_DUTY_DUTY(duty)
			esp.LEDC.SetCH1_CONF1_DUTY_START(1)
			esp.LEDC.SetCH1_CONF0_SIG_OUT_EN(1)
			esp.LEDC.SetCH1_CONF0_PARA_UP(1)
		}
	case 2:
		switch op {
		case ledcChanOpInit:
			esp.LEDC.SetCH2_CONF0_TIMER_SEL(uint32(pwm.timerNum))
			esp.LEDC.SetCH2_CONF0_IDLE_LV(0)
			esp.LEDC.SetCH2_HPOINT_HPOINT(0)
			esp.LEDC.SetCH2_DUTY_DUTY(0)
			esp.LEDC.SetCH2_GAMMA_WR_ADDR_CH_GAMMA_WR_ADDR(0)
			ledcGammaSync()
			esp.LEDC.SetCH2_GAMMA_WR_CH_GAMMA_DUTY_INC(1)
			esp.LEDC.SetCH2_GAMMA_WR_CH_GAMMA_DUTY_CYCLE(1)
			esp.LEDC.SetCH2_GAMMA_WR_CH_GAMMA_SCALE(0)
			esp.LEDC.SetCH2_GAMMA_WR_CH_GAMMA_DUTY_NUM(0)
			esp.LEDC.SetCH2_GAMMA_WR_ADDR_CH_GAMMA_WR_ADDR(0)
			ledcGammaSync()
			esp.LEDC.SetCH2_GAMMA_CONF_CH_GAMMA_ENTRY_NUM(1)
			esp.LEDC.SetCH2_CONF1_DUTY_START(1)
			esp.LEDC.SetCH2_CONF0_SIG_OUT_EN(1)
			esp.LEDC.SetCH2_CONF0_PARA_UP(1)
		case ledcChanOpSetDuty:
			esp.LEDC.SetCH2_DUTY_DUTY(duty)
			esp.LEDC.SetCH2_CONF1_DUTY_START(1)
			esp.LEDC.SetCH2_CONF0_SIG_OUT_EN(1)
			esp.LEDC.SetCH2_CONF0_PARA_UP(1)
		}
	case 3:
		switch op {
		case ledcChanOpInit:
			esp.LEDC.SetCH3_CONF0_TIMER_SEL(uint32(pwm.timerNum))
			esp.LEDC.SetCH3_CONF0_IDLE_LV(0)
			esp.LEDC.SetCH3_HPOINT_HPOINT(0)
			esp.LEDC.SetCH3_DUTY_DUTY(0)
			esp.LEDC.SetCH3_GAMMA_WR_ADDR_CH_GAMMA_WR_ADDR(0)
			ledcGammaSync()
			esp.LEDC.SetCH3_GAMMA_WR_CH_GAMMA_DUTY_INC(1)
			esp.LEDC.SetCH3_GAMMA_WR_CH_GAMMA_DUTY_CYCLE(1)
			esp.LEDC.SetCH3_GAMMA_WR_CH_GAMMA_SCALE(0)
			esp.LEDC.SetCH3_GAMMA_WR_CH_GAMMA_DUTY_NUM(0)
			esp.LEDC.SetCH3_GAMMA_WR_ADDR_CH_GAMMA_WR_ADDR(0)
			ledcGammaSync()
			esp.LEDC.SetCH3_GAMMA_CONF_CH_GAMMA_ENTRY_NUM(1)
			esp.LEDC.SetCH3_CONF1_DUTY_START(1)
			esp.LEDC.SetCH3_CONF0_SIG_OUT_EN(1)
			esp.LEDC.SetCH3_CONF0_PARA_UP(1)
		case ledcChanOpSetDuty:
			esp.LEDC.SetCH3_DUTY_DUTY(duty)
			esp.LEDC.SetCH3_CONF1_DUTY_START(1)
			esp.LEDC.SetCH3_CONF0_SIG_OUT_EN(1)
			esp.LEDC.SetCH3_CONF0_PARA_UP(1)
		}
	case 4:
		switch op {
		case ledcChanOpInit:
			esp.LEDC.SetCH4_CONF0_TIMER_SEL(uint32(pwm.timerNum))
			esp.LEDC.SetCH4_CONF0_IDLE_LV(0)
			esp.LEDC.SetCH4_HPOINT_HPOINT(0)
			esp.LEDC.SetCH4_DUTY_DUTY(0)
			esp.LEDC.SetCH4_GAMMA_WR_ADDR_CH_GAMMA_WR_ADDR(0)
			ledcGammaSync()
			esp.LEDC.SetCH4_GAMMA_WR_CH_GAMMA_DUTY_INC(1)
			esp.LEDC.SetCH4_GAMMA_WR_CH_GAMMA_DUTY_CYCLE(1)
			esp.LEDC.SetCH4_GAMMA_WR_CH_GAMMA_SCALE(0)
			esp.LEDC.SetCH4_GAMMA_WR_CH_GAMMA_DUTY_NUM(0)
			esp.LEDC.SetCH4_GAMMA_WR_ADDR_CH_GAMMA_WR_ADDR(0)
			ledcGammaSync()
			esp.LEDC.SetCH4_GAMMA_CONF_CH_GAMMA_ENTRY_NUM(1)
			esp.LEDC.SetCH4_CONF1_DUTY_START(1)
			esp.LEDC.SetCH4_CONF0_SIG_OUT_EN(1)
			esp.LEDC.SetCH4_CONF0_PARA_UP(1)
		case ledcChanOpSetDuty:
			esp.LEDC.SetCH4_DUTY_DUTY(duty)
			esp.LEDC.SetCH4_CONF1_DUTY_START(1)
			esp.LEDC.SetCH4_CONF0_SIG_OUT_EN(1)
			esp.LEDC.SetCH4_CONF0_PARA_UP(1)
		}
	case 5:
		switch op {
		case ledcChanOpInit:
			esp.LEDC.SetCH5_CONF0_TIMER_SEL(uint32(pwm.timerNum))
			esp.LEDC.SetCH5_CONF0_IDLE_LV(0)
			esp.LEDC.SetCH5_HPOINT_HPOINT(0)
			esp.LEDC.SetCH5_DUTY_DUTY(0)
			esp.LEDC.SetCH5_GAMMA_WR_ADDR_CH_GAMMA_WR_ADDR(0)
			ledcGammaSync()
			esp.LEDC.SetCH5_GAMMA_WR_CH_GAMMA_DUTY_INC(1)
			esp.LEDC.SetCH5_GAMMA_WR_CH_GAMMA_DUTY_CYCLE(1)
			esp.LEDC.SetCH5_GAMMA_WR_CH_GAMMA_SCALE(0)
			esp.LEDC.SetCH5_GAMMA_WR_CH_GAMMA_DUTY_NUM(0)
			esp.LEDC.SetCH5_GAMMA_WR_ADDR_CH_GAMMA_WR_ADDR(0)
			ledcGammaSync()
			esp.LEDC.SetCH5_GAMMA_CONF_CH_GAMMA_ENTRY_NUM(1)
			esp.LEDC.SetCH5_CONF1_DUTY_START(1)
			esp.LEDC.SetCH5_CONF0_SIG_OUT_EN(1)
			esp.LEDC.SetCH5_CONF0_PARA_UP(1)
		case ledcChanOpSetDuty:
			esp.LEDC.SetCH5_DUTY_DUTY(duty)
			esp.LEDC.SetCH5_CONF1_DUTY_START(1)
			esp.LEDC.SetCH5_CONF0_SIG_OUT_EN(1)
			esp.LEDC.SetCH5_CONF0_PARA_UP(1)
		}
	}
}
