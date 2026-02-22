//go:build esp32c3 || esp32s3

// PWM on ESP32-C3/S3 uses the LEDC (LED Control) peripheral, low-speed mode only.
// One timer drives multiple channels; each channel has its own duty, shared frequency.
// Pin routing is via GPIO matrix (SigOutBase + channel index).

package machine

import (
	"device/esp"
	"errors"
	"runtime/volatile"
	"unsafe"
)

const ledcApbClock = 80_000000

const ledcChannelBlockSize = 20

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
	if freq < 100 {
		dutyRes = 14
	} else if freq < 1000 {
		dutyRes = 12
	} else if freq > 100_000 {
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

	baseOff := uintptr(ch) * ledcChannelBlockSize
	conf0 := (*volatile.Register32)(unsafe.Add(unsafe.Pointer(&esp.LEDC.CH0_CONF0), baseOff))
	conf0.Set((uint32(pwm.timerNum) << 0) | (1 << 2) | (0 << 3)) // timer_sel, sig_out_en=1, idle_lv=0
	hpointReg := (*volatile.Register32)(unsafe.Add(unsafe.Pointer(&esp.LEDC.CH0_HPOINT), baseOff))
	dutyReg := (*volatile.Register32)(unsafe.Add(unsafe.Pointer(&esp.LEDC.CH0_DUTY), baseOff))
	hpointReg.Set(0)
	dutyReg.Set(0)
	conf1 := (*volatile.Register32)(unsafe.Add(unsafe.Pointer(&esp.LEDC.CH0_CONF1), baseOff))
	conf1.Set((1 << 10) | (1 << 20) | (1 << 30) | (1 << 31)) // duty_cycle=1, duty_num=1, duty_inc=1, duty_start=1
	conf0.SetBits(1 << 4)                                    // low_speed_update: apply channel config
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
	baseOff := uintptr(channel) * ledcChannelBlockSize
	dutyReg := (*volatile.Register32)(unsafe.Add(unsafe.Pointer(&esp.LEDC.CH0_DUTY), baseOff))
	dutyReg.Set(dutyVal)
	conf1 := (*volatile.Register32)(unsafe.Add(unsafe.Pointer(&esp.LEDC.CH0_CONF1), baseOff))
	conf1.Set((1 << 10) | (1 << 20) | (1 << 30) | (1 << 31)) // duty_start=1 to latch new duty
	conf0 := (*volatile.Register32)(unsafe.Add(unsafe.Pointer(&esp.LEDC.CH0_CONF0), baseOff))
	conf0.SetBits((1 << 2) | (1 << 4)) // sig_out_en + low_speed_update
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
	base := unsafe.Pointer(&esp.LEDC.CH0_CONF0)
	conf0 := (*volatile.Register32)(unsafe.Add(base, uintptr(channel)*ledcChannelBlockSize))
	v := conf0.Get() & ^uint32(1<<3)
	if inverting {
		v |= 1 << 3
	}
	conf0.Set(v)
}
