//go:build esp32c3 || esp32s3

// Low-speed LEDC code, shared by the ESP32-C3 and the ESP32-S3.
//
// Both chips have only the low-speed LEDC block. Their clock, timer and channel
// registers use the same names, so one copy of this code serves both.
//
// The classic ESP32 has a high-speed block instead, with different register
// names. Its versions of these functions are in machine_esp32_pwm.go.
//
// The SVD (lib/esp32-svd/svd/esp32s3.svd) describes the two bits that make a
// change take effect on these chips. CONF0.PARA_UP updates HPOINT, DUTY_START,
// SIG_OUT_EN, TIMER_SEL, DUTY_NUM, DUTY_CYCLE, DUTY_SCALE and DUTY_INC for the
// channel and is cleared by hardware. CONF1.DUTY_START makes the other CONF1
// fields take effect.

package machine

import "device/esp"

// enableClock turns the LEDC hardware on and picks APB_CLK as its clock.
func (pwm *LEDCPWM) enableClock() {
	// Enable LEDC clock and release reset (SYSTEM perip_clk_en0 / perip_rst_en0).
	// The reset clears every timer and channel, so it runs only once. Without
	// the guard a Configure of one timer would erase all the others.
	if !ledcStarted {
		esp.SYSTEM.SetPERIP_CLK_EN0_LEDC_CLK_EN(1)
		esp.SYSTEM.SetPERIP_RST_EN0_LEDC_RST(1)
		esp.SYSTEM.SetPERIP_RST_EN0_LEDC_RST(0)
		ledcStarted = true
	}

	// LEDC global: APB clock source, enable internal clock.
	esp.LEDC.SetCONF_APB_CLK_SEL(1)
	esp.LEDC.SetCONF_CLK_EN(1)
}

func (pwm *LEDCPWM) setTimerConf(dutyRes uint8, divReg uint32) {
	t := &esp.LEDC.TIMER[pwm.timerNum]
	t.SetCONF_DUTY_RES(uint32(dutyRes))
	t.SetCONF_CLK_DIV(divReg)
	t.SetCONF_TICK_SEL(0)
	t.SetCONF_PAUSE(0)
	t.SetCONF_RST(1)
	t.SetCONF_RST(0)
	t.SetCONF_PARA_UP(1)
}

// chanDisable stops a channel from driving its pin.
func chanDisable(ch uint8) {
	c := &esp.LEDC.CH[ch]
	c.SetCONF0_SIG_OUT_EN(0)
	c.SetCONF0_PARA_UP(1)
}

// chanOp implements LEDC low-speed channel ops.
// DUTY_NUM is 0 so a new DUTY_START is not dropped (see machine_esp32_pwm.go).
// Two Set calls within about 100 ns can still both be dropped.
func (pwm *LEDCPWM) chanOp(ch uint8, op ledcChanOp, duty uint32) {
	c := &esp.LEDC.CH[ch]
	switch op {
	case ledcChanOpInit:
		c.SetCONF0_TIMER_SEL(uint32(pwm.timerNum))
		c.SetCONF0_SIG_OUT_EN(1)
		c.SetCONF0_IDLE_LV(0)
		c.SetHPOINT(0)
		c.SetDUTY(0)
		c.SetCONF1_DUTY_CYCLE(0)
		c.SetCONF1_DUTY_NUM(0)
		c.SetCONF1_DUTY_INC(1)
		c.SetCONF1_DUTY_START(1)
		c.SetCONF0_PARA_UP(1)
	case ledcChanOpSetDuty:
		c.SetDUTY(duty)
		c.SetCONF1_DUTY_CYCLE(0)
		c.SetCONF1_DUTY_NUM(0)
		c.SetCONF1_DUTY_INC(1)
		c.SetCONF1_DUTY_START(1)
		c.SetCONF0_SIG_OUT_EN(1)
		c.SetCONF0_PARA_UP(1)
	}
}
