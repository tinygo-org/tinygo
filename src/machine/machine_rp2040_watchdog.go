//go:build rp2040
// +build rp2040

package machine

// Hardware Watchdog Timer API
//
// The RP2040 has a built in HW watchdog Timer. This is a countdown timer that can restart parts of the chip if it reaches zero.
// For example, this can be used to restart the processor if the software running on it gets stuck in an infinite loop
// or similar. The programmer has to periodically write a value to the watchdog to stop it reaching zero.

import (
	"device/rp"
	"errors"
	"runtime/volatile"
	"unsafe"
)

const watchdogRegularBootMagic = 0x6ab73121
const watchdogJumpBootMagic = 0xb007c0d3
const watchdogJumpBootMagicInverse = 0x4ff83f2d

const WatchdogMaxPeriod = uint32(0x7fffff)

const (
	WatchdogReasonNone uint32 = iota
	WatchdogReasonTimer
	WatchdogReasonForce
)

type WatchdogConfig struct {
	// Watchdog ticks before it elapses. Each tick is 1 microsecond.
	// Minimum value is 1.
	// Maximum value is 8388607 (0x7fffff) or approximately 8.4 seconds.
	Ticks uint32

	// Watchdog should be paused when the debugger is stepping through code.
	PauseOnDebug bool

	// If Zero, a standard boot will be performed, if non-zero this is the program counter to jump to on reset.
	ProgramCounter uint32

	// If ProgramCounter is non-zero, this will be the stack pointer used.
	StackPointer uint32
}

type watchdogType struct {
	ctrl    volatile.Register32
	load    volatile.Register32
	reason  volatile.Register32
	scratch [8]volatile.Register32
	tick    volatile.Register32
}

type watchdog struct {
	reg   *watchdogType
	load  uint32
	pause bool
	pc    uint32
	sp    uint32
}

var Watchdog = watchdog{
	reg:   (*watchdogType)(unsafe.Pointer(rp.WATCHDOG)),
	load:  0,
	pause: false,
}

// Configure watchdog but not enable/start it, shall call Enable() method separately.
func (wd *watchdog) Configure(config WatchdogConfig) error {
	if config.Ticks < 1 {
		return errors.New("watchdog period too short")
	}
	if config.Ticks > WatchdogMaxPeriod {
		return errors.New("watchdog period too long")
	}
	// Note, we have x2 here as the watchdog HW currently decrements twice per tick
	wd.load = config.Ticks * 2
	wd.pause = config.PauseOnDebug
	wd.pc = config.ProgramCounter
	wd.sp = config.StackPointer
	return nil
}

// Enable watchdog, set registers and start tick counter
func (wd *watchdog) Enable() error {

	wd.reg.ctrl.ClearBits(rp.WATCHDOG_CTRL_ENABLE)

	// Tick cycles needs to be a divider that when applied to the xosc input,
	// produces a 1MHz clock. So if the xosc frequency is 12MHz, this will need to be 12.
	wd.reg.tick.Set(xoscFreq | rp.WATCHDOG_TICK_ENABLE)

	// Reset everything apart from ROSC and XOSC
	psm.wdsel.SetBits(0x0001ffff & ^(rp.PSM_WDSEL_ROSC | rp.PSM_WDSEL_XOSC))

	// Jump on reboot
	if wd.pc != 0 {
		pc := wd.pc | 1 // thumb mode (sic!), as seen in Pico SDK
		// See section "2.8.1.1. Watchdog Boot" of datasheet
		wd.reg.scratch[4].Set(watchdogJumpBootMagic)
		wd.reg.scratch[5].Set(pc ^ watchdogJumpBootMagicInverse)
		wd.reg.scratch[6].Set(wd.sp)
		wd.reg.scratch[7].Set(pc)
	} else {
		wd.reg.scratch[4].Set(watchdogRegularBootMagic)
	}

	dbgBits := uint32(rp.WATCHDOG_CTRL_PAUSE_DBG0 | rp.WATCHDOG_CTRL_PAUSE_DBG1 | rp.WATCHDOG_CTRL_PAUSE_JTAG)
	if wd.pause {
		wd.reg.ctrl.SetBits(dbgBits)
	} else {
		wd.reg.ctrl.ClearBits(dbgBits)
	}

	err := wd.Kick()
	if err != nil {
		return err
	}

	wd.reg.ctrl.SetBits(rp.WATCHDOG_CTRL_ENABLE)

	return nil
}

// IsEnabled can be used to check if watchdog is active
func (wd *watchdog) IsEnabled() (bool, error) {
	return wd.reg.ctrl.HasBits(rp.WATCHDOG_CTRL_ENABLE), nil
}

// Disable watchdog, no watchdog reboot shall happen
func (wd *watchdog) Disable() error {
	wd.reg.ctrl.ClearBits(rp.WATCHDOG_CTRL_ENABLE)
	return nil
}

// Kick watchdog and reset the tick counter to the configured value
func (wd *watchdog) Kick() error {
	if wd.load == 0 {
		return errors.New("watchdog not configured")
	}
	wd.reg.load.Set(wd.load)
	return nil
}

// Force watchdog trigger (reboots system)
func (wd *watchdog) Force() error {
	wd.reg.ctrl.SetBits(rp.WATCHDOG_CTRL_TRIGGER)
	return nil
}

// Reason tells the kind of the last reboot.
// 0) hardware reset (power cycle or reset button toggled);
// 1) watchdog timer reset (tick counter elapses);
// 2) watchdog forced reset (Force() method call)
func (wd *watchdog) Reason() (uint32, error) {
	return wd.reg.reason.Get(), nil
}

// IsRegularBoot can be used to distinguish regular from jump boots
func (wd *watchdog) IsRegularBoot() (bool, error) {
	return wd.reg.scratch[4].Get() == watchdogRegularBootMagic, nil
}
