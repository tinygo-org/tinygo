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
//
// Register addresses and bit positions come from the ESP32 TRM, chapter 14
// "LED PWM Controller", section 14.4 "Register Summary".

package machine

import (
	"device/esp"
	"runtime/volatile"
	"unsafe"
)

// The GPIO matrix can send an internal signal to almost any pin. Each signal has
// a number. High-speed LEDC channel 0 is number 71, and the rest follow on from
// there, so channels 0 to 7 are 71 to 78.
// (From Espressif's soc/gpio_sig_map.h: LEDC_HS_SIG_OUT0_IDX.)
const LEDC_HS_SIG_OUT0_IDX = 71

const ledcChannelsESP32 = 8

// Each channel has 5 registers and each timer has 2, so the blocks repeat at a
// fixed distance. HSCH0_CONF0 is at 0x0 and HSCH1_CONF0 is at 0x14.
// HSTIMER0_CONF is at 0x140 and HSTIMER1_CONF is at 0x148.
const (
	ledcChannelStride = 0x14
	ledcTimerStride   = 0x8
)

// Bit positions in HSCHn_CONF0, HSCHn_CONF1 and HSTIMERn_CONF.
const (
	ledcTimerSelPos  = 0       // CONF0, which timer the channel follows
	ledcSigOutEn     = 1 << 2  // CONF0, let the channel drive the pin
	ledcDutyCyclePos = 10      // CONF1
	ledcDutyNumPos   = 20      // CONF1
	ledcDutyInc      = 1 << 30 // CONF1
	ledcDutyStart    = 1 << 31 // CONF1, apply the other CONF1 fields
	ledcDivNumPos    = 5       // TIMER CONF, clock divider
	ledcTimerRst     = 1 << 24 // TIMER CONF
	ledcTickSelAPB   = 1 << 25 // TIMER CONF, 1 is APB_CLK and 0 is REF_TICK
)

var (
	PWM0 = &LEDCPWM{SigOutBase: LEDC_HS_SIG_OUT0_IDX, NumChannels: ledcChannelsESP32, timerNum: 0}
	PWM1 = &LEDCPWM{SigOutBase: LEDC_HS_SIG_OUT0_IDX, NumChannels: ledcChannelsESP32, timerNum: 1}
	PWM2 = &LEDCPWM{SigOutBase: LEDC_HS_SIG_OUT0_IDX, NumChannels: ledcChannelsESP32, timerNum: 2}
	PWM3 = &LEDCPWM{SigOutBase: LEDC_HS_SIG_OUT0_IDX, NumChannels: ledcChannelsESP32, timerNum: 3}
)

// chanReg returns a register of channel ch, given the register of channel 0.
func chanReg(channel0 *volatile.Register32, ch uint8) *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Add(unsafe.Pointer(channel0), uintptr(ch)*ledcChannelStride))
}

// timerReg returns a register of timer t, given the register of timer 0.
func timerReg(timer0 *volatile.Register32, t uint8) *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Add(unsafe.Pointer(timer0), uintptr(t)*ledcTimerStride))
}

// enableClock turns the LEDC hardware on and picks its clock.
func (pwm *LEDCPWM) enableClock() {
	// Every peripheral starts switched off, to save power. These registers are
	// in DPORT on this chip. The C3 and S3 call the same block SYSTEM.
	if !ledcStarted {
		esp.DPORT.SetPERIP_CLK_EN_LEDC_CLK_EN(1)
		esp.DPORT.SetPERIP_RST_EN_LEDC_RST(1)
		esp.DPORT.SetPERIP_RST_EN_LEDC_RST(0)
		ledcStarted = true
	}

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
// The low speed timers on the C3 and S3 use the opposite meaning, so their code
// writes 0. Writing 0 here would make every frequency 80 times too slow.
//
// These timers have no PARA_UP bit. The short reset pulse at the end is what
// makes the new values take effect.
func (pwm *LEDCPWM) setTimerConf(dutyRes uint8, divReg uint32) {
	conf := timerReg(&esp.LEDC.HSTIMER0_CONF, pwm.timerNum)
	value := uint32(dutyRes) | divReg<<ledcDivNumPos | ledcTickSelAPB
	conf.Set(value | ledcTimerRst)
	conf.Set(value)
}

// chanDisable stops a channel from driving its pin.
func chanDisable(ch uint8) {
	chanReg(&esp.LEDC.HSCH0_CONF0, ch).ClearBits(ledcSigOutEn)
}

// chanOp does the work for one channel, numbered 0 to 7. It either sets the
// channel up or changes its duty.
//
// DUTY_SCALE stays 0. LEDC can fade slowly from one duty to the next, and 0
// turns that off, so the duty changes in a single step.
func (pwm *LEDCPWM) chanOp(ch uint8, op ledcChanOp, duty uint32) {
	conf0 := chanReg(&esp.LEDC.HSCH0_CONF0, ch)

	// DUTY_NUM and DUTY_CYCLE are 1 step of 1 period, which is the smallest
	// fade. With DUTY_SCALE at 0 the step size is 0, so no fade happens.
	const conf1 = 1<<ledcDutyCyclePos | 1<<ledcDutyNumPos | ledcDutyInc | ledcDutyStart

	switch op {
	case ledcChanOpInit:
		chanReg(&esp.LEDC.HSCH0_HPOINT, ch).Set(0)
		chanReg(&esp.LEDC.HSCH0_DUTY, ch).Set(0)
		chanReg(&esp.LEDC.HSCH0_CONF1, ch).Set(conf1)
		conf0.Set(uint32(pwm.timerNum)<<ledcTimerSelPos | ledcSigOutEn)
	case ledcChanOpSetDuty:
		chanReg(&esp.LEDC.HSCH0_DUTY, ch).Set(duty)
		chanReg(&esp.LEDC.HSCH0_CONF1, ch).Set(conf1)
		conf0.SetBits(ledcSigOutEn)
	}
}
