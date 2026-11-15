//go:build esp32p4

// ESP32-P4 CPLL (CPU PLL) configuration.
//
// At ROM boot the CPU runs from the 40 MHz XTAL.  This file provides
// EnableCPLL360 which powers up the CPLL, programmes it to 360 MHz via the
// REGI2C analog I2C master, and switches the CPU clock source to CPLL.
//
// After EnableCPLL360:
//   - CPU_CLK  = 360 MHz (CPLL / 1)
//   - MEM_CLK  = 180 MHz (CPU / 2)
//   - SYS_CLK  = 180 MHz (MEM / 1)
//   - APB_CLK  =  90 MHz (SYS / 2)
//
// UART (PLL_F80M from SPLL), TIMG0 (XTAL), and PSRAM (MPLL) are unaffected.
//
// Reference: extra/esp-idf/components/esp_hal_clock/esp32p4/include/hal/clk_tree_ll.h
// and extra/esp-idf/components/esp_hw_support/port/esp32p4/rtc_clk.c
package machine

import (
	"device/esp"
	"device/riscv"
	"runtime/volatile"
	"unsafe"
)

// CPLL REGI2C register addresses (from soc/regi2c_cpll.h).
const (
	cpllOcRefDiv = 2 // OC_REF_DIV: bits [3:0]=div_ref, [6:4]=dchgp, [7]=oc_enb_fcal
	cpllOcDiv7_0 = 3 // OC_DIV_7_0: PLL feedback divider
	cpllOcDcur   = 6 // OC_DCUR:    bits [2:0]=dcur, [5:4]=dhref_sel, [6]=dlref_sel
)

// PMU_IMM_HP_CK_POWER bit positions for CPLL (from pmu_reg.h).
// These are WT (write-trigger) bits in the PMU.  ESP-IDF uses SET_PERI_REG_MASK
// (read-modify-write) which works because the hardware treats 0-writes as no-op.
const (
	pmuImm        = 0x501150CC // PMU_IMM_HP_CK_POWER_REG
	pmuXpdCPLL    = 1 << 27    // TIE_HIGH_XPD_CPLL (power on CPLL analog)
	pmuXpdCPLLI2C = 1 << 23    // TIE_HIGH_XPD_CPLL_I2C (power on CPLL I2C)
	pmuCPLLICG    = 1 << 17    // TIE_HIGH_GLOBAL_CPLL_ICG (clock gate enable)
)

// pmuImmReg returns the PMU_IMM_HP_CK_POWER register.
func pmuImmReg() *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Pointer(uintptr(pmuImm)))
}

// cpllPowerOn powers up the CPLL analog circuit and enables its clock gate.
// Matches ESP-IDF clk_ll_cpll_enable().
func cpllPowerOn() {
	r := pmuImmReg()
	r.Set(r.Get() | pmuXpdCPLL | pmuXpdCPLLI2C)
	r.Set(r.Get() | pmuCPLLICG)
}

// cpllConfigure programmes the CPLL analog dividers for the target frequency
// via REGI2C and runs the self-calibration cycle.
//
// For 360 MHz on ECO1+ (CrowPanel is ESP32-P4-Rev1 / ECO2):
//   - div_ref = 0, dchgp = 5, oc_enb_fcal = 0  → reg 2 = 0x50
//   - div7_0  = 9                                → reg 3 = 9   (XTAL*9/1 = 360)
//   - dcur = 3, dhref_sel = 3, dlref_sel = 1     → reg 6 = 0x73
//
// Returns true if calibration completed successfully.
func cpllConfigure() bool {
	regi2cEnableGate()

	// Configure the 160 MHz clock source for the analog I2C master.
	esp.LP_I2C_ANA_MST.CLK160M.SetBits(1) // clk_i2c_mst_sel_160m = 1

	// Start calibration (clear the stop bit).
	esp.HP_SYS_CLKRST.SetANA_PLL_CTRL0_REG_CPU_PLL_CAL_STOP(0)

	// Write CPLL analog configuration via REGI2C.
	regi2cWrite(r2cBlockCPLL, cpllOcRefDiv, 0x50)
	regi2cWrite(r2cBlockCPLL, cpllOcDiv7_0, 9) // 360 MHz on ECO1+
	regi2cWrite(r2cBlockCPLL, cpllOcDcur, 0x73)

	// Wait for calibration to complete (bounded).
	done := false
	for i := 0; i < regi2cMaxBusy; i++ {
		if esp.HP_SYS_CLKRST.GetANA_PLL_CTRL0_REG_CPU_PLL_CAL_END() != 0 {
			done = true
			break
		}
	}

	// Wait 10 µs for true stop (per ESP-IDF esp_rom_delay_us(10)).
	for i := 0; i < 400; i++ { // ~400 nops ≈ 10 µs at 40 MHz
		riscv.Asm("nop")
	}

	// Stop calibration.
	esp.HP_SYS_CLKRST.SetANA_PLL_CTRL0_REG_CPU_PLL_CAL_STOP(1)

	return done
}

// EnableCPLL360 powers on the CPLL, configures it to 360 MHz, sets the clock
// dividers, and switches the CPU clock source from XTAL to CPLL.
//
// After this call CPU=360 MHz, MEM=180 MHz, SYS=180 MHz, APB=90 MHz.
//
// Returns true if the PLL calibration succeeded and the clock switch completed.
// On failure, the CPU remains at 40 MHz XTAL — the system is still functional.
func EnableCPLL360() bool {
	// Step 1: Power on CPLL.
	cpllPowerOn()

	// Step 2: Configure PLL to 360 MHz and run calibration.
	if !cpllConfigure() {
		return false
	}

	// Step 3: Set clock dividers (upscaling order: APB → SYS → MEM → CPU).
	cpllSetDividers()

	// Step 4: Switch CPU clock source to CPLL.
	cpllSwitchToCPLL()

	// Step 5: Ensure instruction cache sees the new clock domain.
	riscv.Asm("fence.i")

	return true
}

// cpllSetDividers configures the clock dividers for 360 MHz CPLL operation.
// Upscaling order: set the slowest downstream clocks first to avoid transient
// over-clocking.  Matches ESP-IDF rtc_clk_cpu_freq_to_cpll_mhz() upscaling path.
//
// Target: CPU=360/1, MEM=360/2=180, SYS=180/1, APB=180/2=90.
// All divider register values are (divider - 1).
func cpllSetDividers() {
	clk := esp.HP_SYS_CLKRST

	// 1. APB divider = 2 (reg value = 1)
	clk.SetROOT_CLK_CTRL2_REG_APB_CLK_DIV_NUM(1)
	cpllBusUpdate()

	// 2. SYS divider = 1 (reg value = 0)
	clk.SetROOT_CLK_CTRL1_REG_SYS_CLK_DIV_NUM(0)
	cpllBusUpdate()

	// 3. MEM divider = 2 (reg value = 1)
	clk.SetROOT_CLK_CTRL1_REG_MEM_CLK_DIV_NUM(1)
	cpllBusUpdate()

	// 4. CPU divider = 1 (reg value = 0), clear fractional parts
	clk.SetROOT_CLK_CTRL0_REG_CPU_CLK_DIV_NUM(0)
	clk.SetROOT_CLK_CTRL0_REG_CPU_CLK_DIV_NUMERATOR(0)
	clk.SetROOT_CLK_CTRL0_REG_CPU_CLK_DIV_DENOMINATOR(0)
	cpllBusUpdate()
}

// cpllBusUpdate latches the current divider values atomically.
// Sets SOC_CLK_DIV_UPDATE and polls until the hardware clears it.
func cpllBusUpdate() {
	clk := esp.HP_SYS_CLKRST
	clk.SetROOT_CLK_CTRL0_REG_SOC_CLK_DIV_UPDATE(1)
	for clk.GetROOT_CLK_CTRL0_REG_SOC_CLK_DIV_UPDATE() != 0 {
		riscv.Asm("nop")
	}
}

// cpllSwitchToCPLL switches the HP root clock source from XTAL (0) to CPLL (1).
// After this write the CPU immediately runs at the CPLL-derived frequency.
func cpllSwitchToCPLL() {
	esp.LP_AON_CLKRST.SetLP_AONCLKRST_HP_CLK_CTRL_LP_AONCLKRST_HP_ROOT_CLK_SRC_SEL(1) // 1 = CPLL
}
