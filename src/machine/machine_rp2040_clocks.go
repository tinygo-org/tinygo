//go:build rp2040

package machine

import (
	"device/arm"
	"device/rp"
	"runtime/volatile"
	"unsafe"
)

func CPUFrequency() uint32 {
	return 125 * MHz
}

// Returns the period of a clock cycle for the raspberry pi pico in nanoseconds.
// Used in PWM API.
func cpuPeriod() uint32 {
	return 1e9 / CPUFrequency()
}

// clockIndex identifies a hardware clock
type clockIndex uint8

const (
	clkGPOUT0 clockIndex = iota // GPIO Muxing 0
	clkGPOUT1                   // GPIO Muxing 1
	clkGPOUT2                   // GPIO Muxing 2
	clkGPOUT3                   // GPIO Muxing 3
	clkRef                      // Watchdog and timers reference clock
	clkSys                      // Processors, bus fabric, memory, memory mapped registers
	clkPeri                     // Peripheral clock for UART and SPI
	clkUSB                      // USB clock
	clkADC                      // ADC clock
	clkRTC                      // Real time clock
	numClocks
)

type clockType struct {
	ctrl     volatile.Register32
	div      volatile.Register32
	selected volatile.Register32
}

type fc struct {
	refKHz   volatile.Register32
	minKHz   volatile.Register32
	maxKHz   volatile.Register32
	delay    volatile.Register32
	interval volatile.Register32
	src      volatile.Register32
	status   volatile.Register32
	result   volatile.Register32
}

type clocksType struct {
	clk   [numClocks]clockType
	resus struct {
		ctrl   volatile.Register32
		status volatile.Register32
	}
	fc0      fc
	wakeEN0  volatile.Register32
	wakeEN1  volatile.Register32
	sleepEN0 volatile.Register32
	sleepEN1 volatile.Register32
	enabled0 volatile.Register32
	enabled1 volatile.Register32
	intR     volatile.Register32
	intE     volatile.Register32
	intF     volatile.Register32
	intS     volatile.Register32
}

var clocks = (*clocksType)(unsafe.Pointer(rp.CLOCKS))

var configuredFreq [numClocks]uint32

type clock struct {
	*clockType
	cix clockIndex
}

// clock returns the clock identified by cix.
func (clks *clocksType) clock(cix clockIndex) clock {
	return clock{
		&clks.clk[cix],
		cix,
	}
}

// hasGlitchlessMux returns true if clock contains a glitchless multiplexer.
//
// Clock muxing consists of two components:
//
// A glitchless mux, which can be switched freely, but whose inputs must be
// free-running.
//
// An auxiliary (glitchy) mux, whose output glitches when switched, but has
// no constraints on its inputs.
//
// Not all clocks have both types of mux.
func (clk *clock) hasGlitchlessMux() bool {
	return clk.cix == clkSys || clk.cix == clkRef
}

// configure configures the clock by selecting the main clock source src
// and the auxiliary clock source auxsrc
// and finally setting the clock frequency to freq
// given the input clock source frequency srcFreq.
func (clk *clock) configure(src, auxsrc, srcFreq, freq uint32) {
	if freq > srcFreq {
		panic("clock frequency cannot be greater than source frequency")
	}

	// Div register is 24.8 int.frac divider so multiply by 2^8 (left shift by 8)
	div := uint32((uint64(srcFreq) << 8) / uint64(freq))

	// If increasing divisor, set divisor before source. Otherwise set source
	// before divisor. This avoids a momentary overspeed when e.g. switching
	// to a faster source and increasing divisor to compensate.
	if div > clk.div.Get() {
		clk.div.Set(div)
	}

	// If switching a glitchless slice (ref or sys) to an aux source, switch
	// away from aux *first* to avoid passing glitches when changing aux mux.
	// Assume (!!!) glitchless source 0 is no faster than the aux source.
	if clk.hasGlitchlessMux() && src == rp.CLOCKS_CLK_SYS_CTRL_SRC_CLKSRC_CLK_SYS_AUX {
		clk.ctrl.ClearBits(rp.CLOCKS_CLK_REF_CTRL_SRC_Msk)
		for !clk.selected.HasBits(1) {
		}
	} else
	// If no glitchless mux, cleanly stop the clock to avoid glitches
	// propagating when changing aux mux. Note it would be a really bad idea
	// to do this on one of the glitchless clocks (clkSys, clkRef).
	{
		// Disable clock. On clkRef and clkSys this does nothing,
		// all other clocks have the ENABLE bit in the same position.
		clk.ctrl.ClearBits(rp.CLOCKS_CLK_GPOUT0_CTRL_ENABLE_Msk)
		if configuredFreq[clk.cix] > 0 {
			// Delay for 3 cycles of the target clock, for ENABLE propagation.
			// Note XOSC_COUNT is not helpful here because XOSC is not
			// necessarily running, nor is timer... so, 3 cycles per loop:
			delayCyc := configuredFreq[clkSys]/configuredFreq[clk.cix] + 1
			for delayCyc != 0 {
				// This could be done more efficiently but TinyGo inline
				// assembly is not yet capable enough to express that. In the
				// meantime, this forces at least 3 cycles per loop.
				delayCyc--
				arm.Asm("nop\nnop\nnop")
			}
		}
	}

	// Set aux mux first, and then glitchless mux if this clock has one.
	clk.ctrl.ReplaceBits(auxsrc<<rp.CLOCKS_CLK_SYS_CTRL_AUXSRC_Pos,
		rp.CLOCKS_CLK_SYS_CTRL_AUXSRC_Msk, 0)

	if clk.hasGlitchlessMux() {
		clk.ctrl.ReplaceBits(src<<rp.CLOCKS_CLK_REF_CTRL_SRC_Pos,
			rp.CLOCKS_CLK_REF_CTRL_SRC_Msk, 0)
		for !clk.selected.HasBits(1 << src) {
		}
	}

	// Enable clock. On clkRef and clkSys this does nothing,
	// all other clocks have the ENABLE bit in the same position.
	clk.ctrl.SetBits(rp.CLOCKS_CLK_GPOUT0_CTRL_ENABLE)

	// Now that the source is configured, we can trust that the user-supplied
	// divisor is a safe value.
	clk.div.Set(div)

	// Store the configured frequency
	configuredFreq[clk.cix] = freq

}

// init initializes the clock hardware.
//
// Must be called before any other clock function.
// Must be called on start and wake up from deep sleep
func (clks *clocksType) init() {
	// Start the watchdog tick
	watchdog.startTick(xoscFreq)

	// Disable resus that may be enabled from previous software
	clks.resus.ctrl.Set(0)

	// Enable the xosc
	xosc.init()

	// Before we touch PLLs, switch sys and ref cleanly away from their aux sources.
	clks.clk[clkSys].ctrl.ClearBits(rp.CLOCKS_CLK_SYS_CTRL_SRC_Msk)
	for !clks.clk[clkSys].selected.HasBits(0x1) {
	}

	clks.clk[clkRef].ctrl.ClearBits(rp.CLOCKS_CLK_REF_CTRL_SRC_Msk)
	for !clks.clk[clkRef].selected.HasBits(0x1) {
	}

	// Configure PLLs
	//                   REF     FBDIV VCO            POSTDIV
	// pllSys: 12 / 1 = 12MHz * 125 = 1500MHZ / 6 / 2 = 125MHz
	// pllUSB: 12 / 1 = 12MHz * 40  = 480 MHz / 5 / 2 =  48MHz
	pllSys.init(1, 1500*MHz, 6, 2)
	pllUSB.init(1, 480*MHz, 5, 2)

	// Configure clocks
	// clkRef = xosc (12MHz) / 1 = 12MHz
	clkref := clks.clock(clkRef)
	clkref.configure(rp.CLOCKS_CLK_REF_CTRL_SRC_XOSC_CLKSRC,
		0, // No aux mux
		12*MHz,
		12*MHz)

	// clkSys = pllSys (125MHz) / 1 = 125MHz
	clksys := clks.clock(clkSys)
	clksys.configure(rp.CLOCKS_CLK_SYS_CTRL_SRC_CLKSRC_CLK_SYS_AUX,
		rp.CLOCKS_CLK_SYS_CTRL_AUXSRC_CLKSRC_PLL_SYS,
		125*MHz,
		125*MHz)

	// clkUSB = pllUSB (48MHz) / 1 = 48MHz
	clkusb := clks.clock(clkUSB)
	clkusb.configure(0, // No GLMUX
		rp.CLOCKS_CLK_USB_CTRL_AUXSRC_CLKSRC_PLL_USB,
		48*MHz,
		48*MHz)

	// clkADC = pllUSB (48MHZ) / 1 = 48MHz
	clkadc := clks.clock(clkADC)
	clkadc.configure(0, // No GLMUX
		rp.CLOCKS_CLK_ADC_CTRL_AUXSRC_CLKSRC_PLL_USB,
		48*MHz,
		48*MHz)

	// clkRTC = xosc (12MHz) / 256 = 46875Hz
	clkrtc := clks.clock(clkRTC)
	clkrtc.configure(0, // No GLMUX
		rp.CLOCKS_CLK_RTC_CTRL_AUXSRC_XOSC_CLKSRC,
		12*MHz,
		46875)

	// clkPeri = clkSys. Used as reference clock for Peripherals.
	// No dividers so just select and enable.
	// Normally choose clkSys or clkUSB.
	clkperi := clks.clock(clkPeri)
	clkperi.configure(0,
		rp.CLOCKS_CLK_PERI_CTRL_AUXSRC_CLK_SYS,
		125*MHz,
		125*MHz)
}

// stop the clock
func (clk *clock) stop() {
	// Disable clock. On clkRef and clkSys this does nothing,
	// all other clocks have the ENABLE bit in the same position.
	clk.ctrl.ClearBits(rp.CLOCKS_CLK_GPOUT0_CTRL_ENABLE)
	configuredFreq[clk.cix] = 0
}

// Stop most clocks & PLLs in preparation for sleep.
// Ref, Sys and Rtc clocks run from XOSC.
// After awakening, need to call configure() to re-initialize.
func (clks *clocksType) sleep() {

	// Configure clocks
	// clkRef = xosc (12MHz) / 1 = 12MHz

	// clkSys = xosc (12MHz) / 1 = 12MHz
	clksys := clks.clock(clkSys)
	clksys.configure(rp.CLOCKS_CLK_SYS_CTRL_SRC_CLKSRC_CLK_SYS_AUX,
		rp.CLOCKS_CLK_SYS_CTRL_AUXSRC_XOSC_CLKSRC,
		12*MHz,
		12*MHz)

	// clkUSB = 0MHz
	clkusb := clks.clock(clkUSB)
	clkusb.stop()

	// clkADC = 0MHz
	clkadc := clks.clock(clkADC)
	clkadc.stop()

	// clkRTC = xosc (12MHz) / 256 = 46875Hz

	// clkPeri = 0MHz
	clkperi := clks.clock(clkPeri)
	clkperi.stop()

	// Stop the PLLs
	pllSys.stop()
	pllUSB.stop()
}
