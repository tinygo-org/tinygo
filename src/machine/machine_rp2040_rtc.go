//go:build rp2040

// Implementation based on code located here:
// https://github.com/raspberrypi/pico-sdk/blob/master/src/rp2_common/hardware_rtc/rtc.c

package machine

import (
	"device/rp"
	"errors"
	"runtime/interrupt"
	"unsafe"
)

type rtcType rp.RTC_Type

type rtcTime struct {
	Year  int16
	Month int8
	Day   int8
	Dotw  int8
	Hour  int8
	Min   int8
	Sec   int8
}

var RTC = (*rtcType)(unsafe.Pointer(rp.RTC))

const (
	second = 1
	minute = 60 * second
	hour   = 60 * minute
	day    = 24 * hour
)

var (
	rtcAlarmRepeats bool
	rtcCallback     func()
	rtcEpoch        = rtcTime{
		Year: 1970, Month: 1, Day: 1, Dotw: 4, Hour: 0, Min: 0, Sec: 0,
	}
)

var (
	ErrRtcDelayTooSmall = errors.New("RTC interrupt deplay is too small, shall be at least 1 second")
	ErrRtcDelayTooLarge = errors.New("RTC interrupt deplay is too large, shall be no more than 1 day")
)

// SetInterrupt configures delayed and optionally recurring interrupt by real time clock.
//
// Delay is specified in whole seconds, allowed range depends on platform.
// Zero delay disables previously configured interrupt, if any.
//
// RP2040 implementation allows delay to be up to 1 day, otherwise a respective error is emitted.
func (rtc *rtcType) SetInterrupt(delay uint32, repeat bool, callback func()) error {

	// Verify delay range
	if delay > day {
		return ErrRtcDelayTooLarge
	}

	// De-configure delayed interrupt if delay is zero
	if delay == 0 {
		rtc.disableInterruptMatch()
		return nil
	}

	// Configure delayed interrupt
	rtc.setDivider()

	rtcAlarmRepeats = repeat
	rtcCallback = callback

	err := rtc.setTime(rtcEpoch)
	if err != nil {
		return err
	}
	rtc.setAlarm(toAlarmTime(delay), callback)

	return nil
}

func toAlarmTime(delay uint32) rtcTime {
	result := rtcEpoch
	remainder := delay + 1 // needed "+1", otherwise alarm fires one second too early
	if remainder >= hour {
		result.Hour = int8(remainder / hour)
		remainder %= hour
	}
	if remainder >= minute {
		result.Min = int8(remainder / minute)
		remainder %= minute
	}
	result.Sec = int8(remainder)
	return result
}

func (rtc *rtcType) setDivider() {
	// Get clk_rtc freq and make sure it is running
	rtcFreq := configuredFreq[clkRTC]
	if rtcFreq == 0 {
		panic("can not set RTC divider, clock is not running")
	}

	// Take rtc out of reset now that we know clk_rtc is running
	resetBlock(rp.RESETS_RESET_RTC)
	unresetBlockWait(rp.RESETS_RESET_RTC)

	// Set up the 1 second divider.
	// If rtc_freq is 400 then clkdiv_m1 should be 399
	rtcFreq -= 1

	// Check the freq is not too big to divide
	if rtcFreq > rp.RTC_CLKDIV_M1_CLKDIV_M1_Msk {
		panic("can not set RTC divider, clock frequency is too big to divide")
	}

	// Write divide value
	rtc.CLKDIV_M1.Set(rtcFreq)
}

// setTime configures RTC with supplied time, initialises and activates it.
func (rtc *rtcType) setTime(t rtcTime) error {

	// Disable RTC and wait while it is still running
	rtc.CTRL.Set(0)
	for rtc.isActive() {
	}

	rtc.SETUP_0.Set((uint32(t.Year) << rp.RTC_SETUP_0_YEAR_Pos) |
		(uint32(t.Month) << rp.RTC_SETUP_0_MONTH_Pos) |
		(uint32(t.Day) << rp.RTC_SETUP_0_DAY_Pos))

	rtc.SETUP_1.Set((uint32(t.Dotw) << rp.RTC_SETUP_1_DOTW_Pos) |
		(uint32(t.Hour) << rp.RTC_SETUP_1_HOUR_Pos) |
		(uint32(t.Min) << rp.RTC_SETUP_1_MIN_Pos) |
		(uint32(t.Sec) << rp.RTC_SETUP_1_SEC_Pos))

	// Load setup values into RTC clock domain
	rtc.CTRL.SetBits(rp.RTC_CTRL_LOAD)

	// Enable RTC and wait for it to be running
	rtc.CTRL.SetBits(rp.RTC_CTRL_RTC_ENABLE)
	for !rtc.isActive() {
	}

	return nil
}

func (rtc *rtcType) isActive() bool {
	return rtc.CTRL.HasBits(rp.RTC_CTRL_RTC_ACTIVE)
}

// setAlarm configures alarm in RTC and arms it.
// The callback is executed in the context of an interrupt handler,
// so regular restructions for this sort of code apply: no blocking, no memory allocation, etc.
func (rtc *rtcType) setAlarm(t rtcTime, callback func()) {

	rtc.disableInterruptMatch()

	// Clear all match enable bits
	rtc.IRQ_SETUP_0.ClearBits(rp.RTC_IRQ_SETUP_0_YEAR_ENA | rp.RTC_IRQ_SETUP_0_MONTH_ENA | rp.RTC_IRQ_SETUP_0_DAY_ENA)
	rtc.IRQ_SETUP_1.ClearBits(rp.RTC_IRQ_SETUP_1_DOTW_ENA | rp.RTC_IRQ_SETUP_1_HOUR_ENA | rp.RTC_IRQ_SETUP_1_MIN_ENA | rp.RTC_IRQ_SETUP_1_SEC_ENA)

	// Only add to setup if it isn't -1 and set the match enable bits for things we care about
	if t.Year >= 0 {
		rtc.IRQ_SETUP_0.SetBits(uint32(t.Year) << rp.RTC_SETUP_0_YEAR_Pos)
		rtc.IRQ_SETUP_0.SetBits(rp.RTC_IRQ_SETUP_0_YEAR_ENA)
	}

	if t.Month >= 0 {
		rtc.IRQ_SETUP_0.SetBits(uint32(t.Month) << rp.RTC_SETUP_0_MONTH_Pos)
		rtc.IRQ_SETUP_0.SetBits(rp.RTC_IRQ_SETUP_0_MONTH_ENA)
	}

	if t.Day >= 0 {
		rtc.IRQ_SETUP_0.SetBits(uint32(t.Day) << rp.RTC_SETUP_0_DAY_Pos)
		rtc.IRQ_SETUP_0.SetBits(rp.RTC_IRQ_SETUP_0_DAY_ENA)
	}

	if t.Dotw >= 0 {
		rtc.IRQ_SETUP_1.SetBits(uint32(t.Dotw) << rp.RTC_SETUP_1_DOTW_Pos)
		rtc.IRQ_SETUP_1.SetBits(rp.RTC_IRQ_SETUP_1_DOTW_ENA)
	}

	if t.Hour >= 0 {
		rtc.IRQ_SETUP_1.SetBits(uint32(t.Hour) << rp.RTC_SETUP_1_HOUR_Pos)
		rtc.IRQ_SETUP_1.SetBits(rp.RTC_IRQ_SETUP_1_HOUR_ENA)
	}

	if t.Min >= 0 {
		rtc.IRQ_SETUP_1.SetBits(uint32(t.Min) << rp.RTC_SETUP_1_MIN_Pos)
		rtc.IRQ_SETUP_1.SetBits(rp.RTC_IRQ_SETUP_1_MIN_ENA)
	}

	if t.Sec >= 0 {
		rtc.IRQ_SETUP_1.SetBits(uint32(t.Sec) << rp.RTC_SETUP_1_SEC_Pos)
		rtc.IRQ_SETUP_1.SetBits(rp.RTC_IRQ_SETUP_1_SEC_ENA)
	}

	// Enable the IRQ at the proc
	interrupt.New(rp.IRQ_RTC_IRQ, rtcHandleInterrupt).Enable()

	// Enable the IRQ at the peri
	rtc.INTE.Set(rp.RTC_INTE_RTC)

	rtc.enableInterruptMatch()
}

func (rtc *rtcType) enableInterruptMatch() {
	// Set matching and wait for it to be enabled
	rtc.IRQ_SETUP_0.SetBits(rp.RTC_IRQ_SETUP_0_MATCH_ENA)
	for !rtc.IRQ_SETUP_0.HasBits(rp.RTC_IRQ_SETUP_0_MATCH_ACTIVE) {
	}
}

func (rtc *rtcType) disableInterruptMatch() {
	// Disable matching and wait for it to stop being active
	rtc.IRQ_SETUP_0.ClearBits(rp.RTC_IRQ_SETUP_0_MATCH_ENA)
	for rtc.IRQ_SETUP_0.HasBits(rp.RTC_IRQ_SETUP_0_MATCH_ACTIVE) {
	}
}

func rtcHandleInterrupt(itr interrupt.Interrupt) {
	// Always disable the alarm to clear the current IRQ.
	// Even if it is a repeatable alarm, we don't want it to keep firing.
	// If it matches on a second it can keep firing for that second.
	RTC.disableInterruptMatch()

	// Call user callback function
	if rtcCallback != nil {
		rtcCallback()
	}

	if rtcAlarmRepeats {
		// If it is a repeatable alarm, reset time and re-enable the alarm.
		RTC.setTime(rtcEpoch)
		RTC.enableInterruptMatch()
	}
}
