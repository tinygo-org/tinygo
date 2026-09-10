//go:build esp32c3 || esp32s3

// Low-speed LEDC code, shared by the ESP32-C3 and the ESP32-S3.
//
// Both chips have only the low-speed LEDC block. Their clock and timer registers
// use the same names, so one copy of this code serves both.
//
// The classic ESP32 has a high-speed block instead, with different register
// names. Its versions of these two functions are in machine_esp32_pwm.go.
//
// The SVD (lib/cmsis-svd/data/Espressif/esp32s3.svd) describes the two bits that
// make a change take effect on these chips. CONF0.PARA_UP "updates HPOINT,
// DUTY_START, SIG_OUT_EN, TIMER_SEL, DUTY_NUM, DUTY_CYCLE, DUTY_SCALE, DUTY_INC
// for channel and is auto-cleared by hardware". CONF1.DUTY_START "other CONF1
// fields take effect when this bit is set to 1".

package machine

import "device/esp"

// enableClock turns the LEDC hardware on and picks APB_CLK as its clock.
func (pwm *LEDCPWM) enableClock() {
	// Enable LEDC clock and release reset (SYSTEM perip_clk_en0 / perip_rst_en0).
	esp.SYSTEM.SetPERIP_RST_EN0_LEDC_RST(1)
	esp.SYSTEM.SetPERIP_CLK_EN0_LEDC_CLK_EN(1)
	esp.SYSTEM.SetPERIP_RST_EN0_LEDC_RST(0)

	// LEDC global: APB clock source, enable internal clock.
	esp.LEDC.SetCONF_APB_CLK_SEL(1)
	esp.LEDC.SetCONF_CLK_EN(1)
}

func (pwm *LEDCPWM) setTimerConf(dutyRes uint8, divReg uint32) {
	t := pwm.timerNum
	switch t {
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
