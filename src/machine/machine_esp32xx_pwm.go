//go:build esp32c3 || esp32s3

// PWM on ESP32-C3/S3 uses the LEDC (LED Control) peripheral, low-speed mode only.
// One timer drives multiple channels; each channel has its own duty, shared frequency.
// Pin routing is via GPIO matrix (SigOutBase + channel index).
//
// Channel config (chanOp) follows the hardware contract from:
//   - ESP-IDF: https://docs.espressif.com/projects/esp-idf/en/latest/esp32/api-reference/peripherals/ledc.html
//     (timer config → channel config → duty + update_duty).
//   - SVD (e.g. lib/cmsis-svd/data/Espressif/esp32s3.svd): CONF0.PARA_UP "updates
//     HPOINT, DUTY_START, SIG_OUT_EN, TIMER_SEL, DUTY_NUM, DUTY_CYCLE, DUTY_SCALE,
//     DUTY_INC for channel and is auto-cleared by hardware"; CONF1.DUTY_START "other
//     CONF1 fields take effect when this bit is set to 1".

package machine

import (
	"device/esp"
	"errors"
)

const ledcApbClock = 80_000000

const ledcDutyFracBits = 4 // DUTY register has 4 fractional bits; write value<<4

const ledcDividerFracBits = 8 // Clock divider register = actual_divider * 256

var errPWMNoChannel = errors.New("pwm: no free channel")

type LEDCPWM struct {
	SigOutBase  uint32 // GPIO matrix signal index for channel 0 (e.g. 73 on S3, 45 on C3)
	NumChannels uint8
	timerNum    uint8 // 0–3: which LEDC timer (frequency) this PWM uses
	dutyRes     uint8
	configured  bool
	channelPin  [8]Pin
}

type ledcChanOp uint8

const (
	ledcChanOpInit      ledcChanOp = iota // initial per-channel setup (timer, enable, HPOINT/DUTY/CONF1, PARA_UP)
	ledcChanOpSetDuty                     // update duty and latch it (DUTY + CONF1 + PARA_UP)
	ledcChanOpSetInvert                   // change idle level (IDLE_LV)
)

func (pwm *LEDCPWM) Configure(config PWMConfig) error {
	// Enable LEDC clock and release reset (SYSTEM perip_clk_en0 / perip_rst_en0).
	esp.SYSTEM.SetPERIP_RST_EN0_LEDC_RST(1)
	esp.SYSTEM.SetPERIP_CLK_EN0_LEDC_CLK_EN(1)
	esp.SYSTEM.SetPERIP_RST_EN0_LEDC_RST(0)

	// LEDC global: APB clock source, enable internal clock.
	esp.LEDC.SetCONF_APB_CLK_SEL(1)
	esp.LEDC.SetCONF_CLK_EN(1)

	period := config.Period
	if period == 0 {
		period = 1_000_000
	}
	freq := uint64(1e9) / period
	dutyRes := uint8(10)
	switch {
	case freq < 100:
		dutyRes = 14
	case freq < 1000:
		dutyRes = 12
	case freq > 100_000:
		dutyRes = 8
	}

	// Timer divider: period_ns = (2^dutyRes * divActual/256) / 80MHz * 1e9 => divReg = divActual<<8.
	divActual := ledcApbClock / (uint32(freq) * (1 << dutyRes))
	if divActual == 0 {
		divActual = 1
	}
	divReg := divActual << ledcDividerFracBits
	if divReg > 0x3ffff {
		return ErrPWMPeriodTooLong
	}

	// Selected timer: resolution, divider, no pause, reset then latch config with PARA_UP.
	pwm.setTimerConf(dutyRes, divReg)

	pwm.dutyRes = dutyRes
	pwm.configured = true
	for i := range pwm.channelPin {
		pwm.channelPin[i] = NoPin
	}
	return nil
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

func (pwm *LEDCPWM) Channel(pin Pin) (uint8, error) {
	if !pwm.configured {
		return 0, errors.New("pwm: not configured")
	}
	if pin == NoPin {
		return 0, ErrInvalidOutputPin
	}
	var ch uint8
	for ch = 0; ch < pwm.NumChannels; ch++ {
		if pwm.channelPin[ch] == NoPin {
			break
		}
	}
	if ch >= pwm.NumChannels {
		return 0, errPWMNoChannel
	}

	pwm.channelPin[ch] = pin
	signal := pwm.SigOutBase + uint32(ch)
	pin.configure(PinConfig{Mode: PinOutput}, signal) // GPIO matrix: pin <- LEDC_LS_SIG_OUTn
	pwm.chanOp(ch, ledcChanOpInit, 0, false)
	return ch, nil
}

func (pwm *LEDCPWM) Set(channel uint8, value uint32) {
	if channel >= pwm.NumChannels {
		return
	}
	top := uint32(1<<pwm.dutyRes) - 1
	if value > top {
		value = top
	}
	dutyVal := value << ledcDutyFracBits
	pwm.chanOp(channel, ledcChanOpSetDuty, dutyVal, false)
}

func (pwm *LEDCPWM) Top() uint32 {
	if !pwm.configured {
		return 0
	}
	return uint32(1<<pwm.dutyRes) - 1
}

func (pwm *LEDCPWM) SetInverting(channel uint8, inverting bool) {
	if channel >= pwm.NumChannels {
		return
	}
	pwm.chanOp(channel, ledcChanOpSetInvert, 0, inverting)
}
