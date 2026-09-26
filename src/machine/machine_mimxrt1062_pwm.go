//go:build mimxrt1062

package machine

// PWM driver for the four FlexPWM modules of the i.MX RT1062. Each module
// has four submodules with an own 16-bit counter and up to three outputs
// (A, B and X). One PWM instance drives one submodule, so the outputs of a
// submodule share the period but have an own duty cycle. The register
// programming follows Teensyduino cores/teensy4/pwm.c.

import (
	"device/nxp"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

// pwmClockHz is the FlexPWM input clock: the IPG clock, the 600 MHz AHB
// clock divided by 4 (see runtime_mimxrt1062_clock.go).
const pwmClockHz = 150000000

// pwmDefaultPeriod is the period used for PWMConfig.Period == 0. It is the
// Teensyduino default of about 4.482 kHz, which is suitable for LEDs.
const pwmDefaultPeriod = 223100 // ns

// PWM channel positions within a submodule
const (
	pwmChannelA = 0
	pwmChannelB = 1
	pwmChannelX = 2
)

// PWM is one FlexPWM submodule.
type PWM struct {
	regs  *nxp.PWM_Type
	sm    uint8 // submodule 0-3
	clock nxp.Clock
}

var (
	PWM1_0 = &PWM{regs: nxp.PWM1, sm: 0, clock: nxp.ClockIpPwm1}
	PWM1_1 = &PWM{regs: nxp.PWM1, sm: 1, clock: nxp.ClockIpPwm1}
	PWM1_2 = &PWM{regs: nxp.PWM1, sm: 2, clock: nxp.ClockIpPwm1}
	PWM1_3 = &PWM{regs: nxp.PWM1, sm: 3, clock: nxp.ClockIpPwm1}
	PWM2_0 = &PWM{regs: nxp.PWM2, sm: 0, clock: nxp.ClockIpPwm2}
	PWM2_1 = &PWM{regs: nxp.PWM2, sm: 1, clock: nxp.ClockIpPwm2}
	PWM2_2 = &PWM{regs: nxp.PWM2, sm: 2, clock: nxp.ClockIpPwm2}
	PWM2_3 = &PWM{regs: nxp.PWM2, sm: 3, clock: nxp.ClockIpPwm2}
	PWM3_0 = &PWM{regs: nxp.PWM3, sm: 0, clock: nxp.ClockIpPwm3}
	PWM3_1 = &PWM{regs: nxp.PWM3, sm: 1, clock: nxp.ClockIpPwm3}
	PWM3_2 = &PWM{regs: nxp.PWM3, sm: 2, clock: nxp.ClockIpPwm3}
	PWM3_3 = &PWM{regs: nxp.PWM3, sm: 3, clock: nxp.ClockIpPwm3}
	PWM4_0 = &PWM{regs: nxp.PWM4, sm: 0, clock: nxp.ClockIpPwm4}
	PWM4_1 = &PWM{regs: nxp.PWM4, sm: 1, clock: nxp.ClockIpPwm4}
	PWM4_2 = &PWM{regs: nxp.PWM4, sm: 2, clock: nxp.ClockIpPwm4}
	PWM4_3 = &PWM{regs: nxp.PWM4, sm: 3, clock: nxp.ClockIpPwm4}
)

// pwmPinInfo describes one PWM output pin of the board: the submodule that
// drives it, the channel within the submodule, and the IOMUXC alternate
// function of the pad. The pwmPins table is defined per board.
type pwmPinInfo struct {
	pin     Pin
	pwm     *PWM
	channel uint8
	mux     uint8
}

// pwmSubmodule is the register block of one FlexPWM submodule. The generated
// device file declares the registers of all four submodules as flat fields,
// this mirrors one 0x60 byte block so the submodule can be indexed.
type pwmSubmodule struct {
	CNT      volatile.Register16 // 0x00
	INIT     volatile.Register16 // 0x02
	CTRL2    volatile.Register16 // 0x04
	CTRL     volatile.Register16 // 0x06
	_        uint16
	VAL0     volatile.Register16 // 0x0A
	FRACVAL1 volatile.Register16 // 0x0C
	VAL1     volatile.Register16 // 0x0E
	FRACVAL2 volatile.Register16 // 0x10
	VAL2     volatile.Register16 // 0x12
	FRACVAL3 volatile.Register16 // 0x14
	VAL3     volatile.Register16 // 0x16
	FRACVAL4 volatile.Register16 // 0x18
	VAL4     volatile.Register16 // 0x1A
	FRACVAL5 volatile.Register16 // 0x1C
	VAL5     volatile.Register16 // 0x1E
	FRCTRL   volatile.Register16 // 0x20
	OCTRL    volatile.Register16 // 0x22
	STS      volatile.Register16 // 0x24
	INTEN    volatile.Register16 // 0x26
	DMAEN    volatile.Register16 // 0x28
	TCTRL    volatile.Register16 // 0x2A
	DISMAP0  volatile.Register16 // 0x2C
	DISMAP1  volatile.Register16 // 0x2E
	DTCNT0   volatile.Register16 // 0x30
	DTCNT1   volatile.Register16 // 0x32
}

func (pwm *PWM) regsSM() *pwmSubmodule {
	return (*pwmSubmodule)(unsafe.Add(unsafe.Pointer(pwm.regs), 0x60*uintptr(pwm.sm)))
}

// mask returns the submodule bit for the LDOK, CLDOK, RUN and OUTEN fields.
func (pwm *PWM) mask() uint16 {
	return 1 << pwm.sm
}

// Configure enables and configures this PWM. All outputs start with a duty
// cycle of zero, use Channel and Set to drive a pin.
func (pwm *PWM) Configure(config PWMConfig) error {
	pwm.clock.Enable(true)

	p := pwm.regs
	sm := pwm.regsSM()
	m := pwm.mask()

	mask := interrupt.Disable()

	// the fault inputs are not used: fault level logic high, clear the
	// fault flags (shared by the four submodules of the module)
	p.FCTRL0.Set(0xF << nxp.PWM_FCTRL0_FLVL_Pos)
	p.FSTS0.Set(0xF)
	p.FFILT0.Set(0)

	p.MCTRL.SetBits(m << nxp.PWM_MCTRL_CLDOK_Pos)
	sm.CTRL2.Set(nxp.PWM_SM0CTRL2_INDEP_Msk | // A and B are independent outputs
		nxp.PWM_SM0CTRL2_WAITEN_Msk | nxp.PWM_SM0CTRL2_DBGEN_Msk)
	sm.OCTRL.Set(0)
	sm.DTCNT0.Set(0)
	sm.INIT.Set(0)
	sm.VAL0.Set(0)
	sm.VAL1.Set(0)
	sm.VAL2.Set(0)
	sm.VAL3.Set(0)
	sm.VAL4.Set(0)
	sm.VAL5.Set(0)
	p.MCTRL.SetBits(m << nxp.PWM_MCTRL_LDOK_Pos)

	err := pwm.setPeriod(config.Period)
	if err == nil {
		p.MCTRL.SetBits(m << nxp.PWM_MCTRL_RUN_Pos)
	}

	interrupt.Restore(mask)
	return err
}

// Channel returns a PWM channel for the given pin. If the pin is not an
// output of this PWM submodule, ErrInvalidOutputPin is returned. The pad of
// the pin is connected to the PWM output.
func (pwm *PWM) Channel(pin Pin) (channel uint8, err error) {
	for _, info := range pwmPins {
		if info.pin != pin || info.pwm != pwm {
			continue
		}
		// configure the pad like a UART TX pad: fast slew, medium drive
		pad, mux := pin.getPad()
		pad.Set((1 << 0) | (3 << 3) | (3 << 6)) // SRE | DSE(3) | SPEED(3)
		mux.Set(uint32(info.mux))
		return info.channel, nil
	}
	return 0, ErrInvalidOutputPin
}

// PWMPeripheral returns the PWM submodule that can drive the given pin, for
// use with (*PWM).Channel. ErrInvalidOutputPin is returned for a pin without
// PWM output.
func PWMPeripheral(pin Pin) (*PWM, error) {
	for _, info := range pwmPins {
		if info.pin == pin {
			return info.pwm, nil
		}
	}
	return nil, ErrInvalidOutputPin
}

// SetPeriod updates the period of this PWM submodule in nanoseconds. The
// duty cycles of the channels are scaled to the new period. To set a
// particular frequency, use the following formula:
//
//	period = 1e9 / frequency
//
// A period of 0 picks a period that works well for LEDs (about 4.5 kHz).
// The longest attainable period is about 55.9 ms.
func (pwm *PWM) SetPeriod(period uint64) error {
	mask := interrupt.Disable()
	err := pwm.setPeriod(period)
	interrupt.Restore(mask)
	return err
}

// setPeriod must run with interrupts disabled.
func (pwm *PWM) setPeriod(period uint64) error {
	if period == 0 {
		period = pwmDefaultPeriod
	}

	// counts of the PWM clock per period, reduced by the 1/2/../128
	// prescaler until the modulo fits the 16-bit counter
	div := period * pwmClockHz / 1000000000
	prescale := uint16(0)
	for div > 0xFFFF && prescale < 7 {
		div >>= 1
		prescale++
	}
	if div > 0xFFFF {
		return ErrPWMPeriodTooLong
	}
	if div < 2 {
		div = 2
	}

	p := pwm.regs
	sm := pwm.regsSM()
	m := pwm.mask()
	olddiv := uint32(sm.VAL1.Get()) + 1

	p.MCTRL.SetBits(m << nxp.PWM_MCTRL_CLDOK_Pos)
	sm.CTRL.Set(nxp.PWM_SM0CTRL_FULL_Msk | prescale<<nxp.PWM_SM0CTRL_PRSC_Pos)
	sm.VAL1.Set(uint16(div - 1))
	// scale the duty cycles to the new period
	sm.VAL0.Set(uint16(uint64(sm.VAL0.Get()) * div / uint64(olddiv)))
	sm.VAL3.Set(uint16(uint64(sm.VAL3.Get()) * div / uint64(olddiv)))
	sm.VAL5.Set(uint16(uint64(sm.VAL5.Get()) * div / uint64(olddiv)))
	p.MCTRL.SetBits(m << nxp.PWM_MCTRL_LDOK_Pos)

	return nil
}

// Period returns the used PWM period in nanoseconds.
func (pwm *PWM) Period() uint64 {
	sm := pwm.regsSM()
	div := uint64(sm.VAL1.Get()) + 1
	prescale := (sm.CTRL.Get() & nxp.PWM_SM0CTRL_PRSC_Msk) >> nxp.PWM_SM0CTRL_PRSC_Pos
	return div * (1 << prescale) * 1000000000 / pwmClockHz
}

// Top returns the current counter top, for use in duty cycle calculation.
func (pwm *PWM) Top() uint32 {
	return uint32(pwm.regsSM().VAL1.Get())
}

// Counter returns the current counter value of this submodule, for
// debugging.
func (pwm *PWM) Counter() uint32 {
	return uint32(pwm.regsSM().CNT.Get())
}

// Set updates the channel value. This is used to control the channel duty
// cycle, in other words the fraction of time the channel output is high (or
// low when inverted). For example, to set it to a 25% duty cycle, use:
//
//	pwm.Set(channel, pwm.Top() / 4)
//
// pwm.Set(channel, 0) sets the output to low and pwm.Set(channel,
// pwm.Top()) to high, assuming the output isn't inverted.
func (pwm *PWM) Set(channel uint8, value uint32) {
	p := pwm.regs
	sm := pwm.regsSM()
	m := pwm.mask()

	mask := interrupt.Disable()
	modulo := uint32(sm.VAL1.Get())
	if value > modulo {
		value = modulo
	}
	p.MCTRL.SetBits(m << nxp.PWM_MCTRL_CLDOK_Pos)
	switch channel {
	case pwmChannelA:
		sm.VAL3.Set(uint16(value))
		p.OUTEN.SetBits(m << nxp.PWM_OUTEN_PWMA_EN_Pos)
	case pwmChannelB:
		sm.VAL5.Set(uint16(value))
		p.OUTEN.SetBits(m << nxp.PWM_OUTEN_PWMB_EN_Pos)
	case pwmChannelX:
		// the X output is low from the counter wrap to VAL0 and high
		// from VAL0 to the wrap, so the high time is modulo-VAL0
		sm.VAL0.Set(uint16(modulo - value))
		p.OUTEN.SetBits(m << nxp.PWM_OUTEN_PWMX_EN_Pos)
	}
	p.MCTRL.SetBits(m << nxp.PWM_MCTRL_LDOK_Pos)
	interrupt.Restore(mask)
}

// Get returns the current level of the channel (last set by Set).
func (pwm *PWM) Get(channel uint8) (value uint32) {
	sm := pwm.regsSM()
	switch channel {
	case pwmChannelA:
		return uint32(sm.VAL3.Get())
	case pwmChannelB:
		return uint32(sm.VAL5.Get())
	case pwmChannelX:
		return uint32(sm.VAL1.Get()) - uint32(sm.VAL0.Get())
	}
	return 0
}

// SetInverting sets whether to invert the output of this channel. Without
// inverting, a 25% duty cycle means the output is high for 25% of the time.
func (pwm *PWM) SetInverting(channel uint8, inverting bool) {
	var bit uint16
	switch channel {
	case pwmChannelA:
		bit = nxp.PWM_SM0OCTRL_POLA_Msk
	case pwmChannelB:
		bit = nxp.PWM_SM0OCTRL_POLB_Msk
	case pwmChannelX:
		bit = nxp.PWM_SM0OCTRL_POLX_Msk
	default:
		return
	}
	mask := interrupt.Disable()
	if inverting {
		pwm.regsSM().OCTRL.SetBits(bit)
	} else {
		pwm.regsSM().OCTRL.ClearBits(bit)
	}
	interrupt.Restore(mask)
}

// Enable starts or stops the counter of this PWM submodule.
func (pwm *PWM) Enable(enable bool) {
	mask := interrupt.Disable()
	if enable {
		pwm.regs.MCTRL.SetBits(pwm.mask() << nxp.PWM_MCTRL_RUN_Pos)
	} else {
		pwm.regs.MCTRL.ClearBits(pwm.mask() << nxp.PWM_MCTRL_RUN_Pos)
	}
	interrupt.Restore(mask)
}

// IsEnabled returns true if the counter of this PWM submodule runs.
func (pwm *PWM) IsEnabled() bool {
	return pwm.regs.MCTRL.HasBits(pwm.mask() << nxp.PWM_MCTRL_RUN_Pos)
}
