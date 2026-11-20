//go:build esp32p4

// ESP32-P4 internal LDO regulator driver.
//
// The ESP32-P4 has four general-purpose internal LDO regulators (VO1..VO4,
// channels 1..4).  On the Elecrow CrowPanel ESP32-P4 10.1" they power:
//
//	LDO3 = 2.5 V — MIPI DSI PHY (VFB3/VO3 in the V1.2 schematic)
//	LDO4 = 3.3 V — peripheral power (VFB4/VO4)
//
// These must be configured before any peripheral that relies on them (e.g.
// the MIPI DSI PHY in Phase 5) is initialised.
//
// This is a port of ESP-IDF:
//   - components/esp_hw_support/ldo/esp_ldo_regulator.c (acquire sequence)
//   - components/esp_hal_pmu/esp32p4/include/hal/ldo_ll.h  (register ops)
//
// Register notes (ESP32-P4 PMU, base 0x50115000):
//   - The six EXT_LDO register pairs start at offset 0x1B8, 8 bytes each
//     ({main, ana}).  IMPORTANT: the LDO units are NOT in order — ESP-IDF
//     maps unit (channel−1) to register index via {0, 3, 1, 4}:
//     VO1→ext_ldo[0], VO2→ext_ldo[3], VO3→ext_ldo[1], VO4→ext_ldo[4].
//   - All six register pairs share one field layout (verified against the
//     SVD constants), so the helpers below use shared bit positions.
//   - Voltage is set via DREF/MUL in the ANA register.  The output model is
//     Vref = (dref < 9) ? 0.5 + dref*0.05 : 1 + (dref-9)*0.1
//     Vout = (Vref*K + Vos) * (1 + 0.25*mul*C)
//     where K, Vos, C are per-chip calibration constants in eFuse BLOCK2
//     (RD_MAC_SYS_2/3/4); defaults K=1.0, Vos=0, C=1.0 apply when the eFuse
//     block version is 0.
//   - Requesting exactly 3300 mV bypasses the LDO (TIEH=1) and connects the
//     3.3 V rail directly; DREF/MUL are ignored in that case.
package machine

import (
	"device/esp"
	"errors"
	"runtime/volatile"
	"unsafe"
)

// LDOChannel identifies one of the four internal LDO regulators.
// Channel IDs are 1-indexed to match the hardware naming (VO1..VO4).
type LDOChannel uint8

const (
	LDO1 LDOChannel = 1 + iota // VO1
	LDO2                       // VO2
	LDO3                       // VO3 — 2.5 V MIPI DSI PHY on CrowPanel
	LDO4                       // VO4 — 3.3 V peripheral rail on CrowPanel
)

// Errors returned by LDOConfigure / LDODisable.
var (
	ErrLDOBadChannel = errors.New("machine: invalid LDO channel (want 1..4)")
	ErrLDOBadVoltage = errors.New("machine: LDO voltage out of range (500..2700 mV, or exactly 3300 for rail bypass)")
)

const (
	ldoRailVoltageMV uint32 = 3300 // bypass: connect the 3.3 V rail directly
	ldoMinVoltageMV  uint32 = 500
	ldoMaxVoltageMV  uint32 = 2700
)

// ldoExtIndex maps an LDO unit (channel−1) to its PMU EXT_LDO register-pair
// index.  From ESP-IDF ldo_ll.h: index_array[4] = {0, 3, 1, 4}.
var ldoExtIndex = [4]uint8{0, 3, 1, 4}

// Shared EXT_LDO field layout (identical for all six register pairs).
const (
	ldoForceTiehSel = 1 << 7  // main: 1 = software controls the LDO
	ldoXPD          = 1 << 8  // main: power-down (1 = enabled)
	ldoTieh         = 1 << 14 // main: 1 = bypass to 3.3 V rail

	ldoMulPos   = 23                // ana: multiply factor [25:23]
	ldoMulMsk   = 0x7 << ldoMulPos  //
	ldoEnVdet   = 1 << 26           // ana: ripple suppression
	ldoEnCurLim = 1 << 27           // ana: current limit (used while ramping)
	ldoDrefPos  = 28                // ana: internal reference selector [31:28]
	ldoDrefMsk  = 0xf << ldoDrefPos //
)

// ldoMainReg / ldoAnaReg return the EXT_LDO register pair for a unit.
// The six pairs are contiguous 8-byte entries starting at PMU+0x1B8.
func ldoMainReg(unit uint8) *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Add(unsafe.Pointer(&esp.PMU.EXT_LDO_P0_0P1A),
		uintptr(ldoExtIndex[unit])*8))
}

func ldoAnaReg(unit uint8) *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Add(unsafe.Pointer(&esp.PMU.EXT_LDO_P0_0P1A_ANA),
		uintptr(ldoExtIndex[unit])*8))
}

// LDOConfigure sets the given LDO channel to voltageMV millivolts and enables
// it.  voltageMV must be in 500..2700 mV, or exactly 3300 to bypass the LDO
// and use the 3.3 V rail directly.
//
// The sequence follows esp_ldo_acquire_channel(): current-limit on, program
// DREF/MUL, take software ownership, ripple suppression on, enable, then
// current-limit off.
func LDOConfigure(channel LDOChannel, voltageMV uint32) error {
	if channel < LDO1 || channel > LDO4 {
		return ErrLDOBadChannel
	}
	if voltageMV != ldoRailVoltageMV && (voltageMV < ldoMinVoltageMV || voltageMV > ldoMaxVoltageMV) {
		return ErrLDOBadVoltage
	}
	unit := uint8(channel - 1)
	main, ana := ldoMainReg(unit), ldoAnaReg(unit)

	// 1. Limit inrush current while the output ramps.
	ana.SetBits(ldoEnCurLim)

	// 2. Program the output voltage (or rail bypass).
	dref, mul, useRail := ldoVoltageToDrefMul(unit, voltageMV)
	if useRail {
		main.SetBits(ldoTieh)
	} else {
		main.ClearBits(ldoTieh)
	}
	ana.ReplaceBits(uint32(dref), ldoDrefMsk, ldoDrefPos)
	ana.ReplaceBits(uint32(mul), ldoMulMsk, ldoMulPos)

	// 3. Take software ownership (force_tieh_sel=1, tieh_sel=0=tieh field).
	main.SetBits(ldoForceTiehSel)
	main.ClearBits(0x3 << 9) // TIEH_SEL[10:9] = 0

	// 4. Suppress voltage ripple.
	ana.SetBits(ldoEnVdet)

	// 5. Enable the regulator.
	main.SetBits(ldoXPD)

	// 6. Release the current limit once up.
	ana.ClearBits(ldoEnCurLim)
	return nil
}

// LDODisable powers down the given LDO channel.
func LDODisable(channel LDOChannel) error {
	if channel < LDO1 || channel > LDO4 {
		return ErrLDOBadChannel
	}
	ldoMainReg(uint8(channel - 1)).ClearBits(ldoXPD)
	return nil
}

// LDOState returns the raw EXT_LDO main/ana register contents of a channel.
// Intended as a board bring-up diagnostic (verify XPD/DREF/MUL/TIEH fields).
func LDOState(channel LDOChannel) (main, ana uint32, err error) {
	if channel < LDO1 || channel > LDO4 {
		return 0, 0, ErrLDOBadChannel
	}
	unit := uint8(channel - 1)
	return ldoMainReg(unit).Get(), ldoAnaReg(unit).Get(), nil
}

// ldoVoltageToDrefMul converts a target voltage to (dref, mul, useRail).
// Port of ldo_ll_voltage_to_dref_mul() with fixed-point arithmetic (all
// constants scaled by 1000; every intermediate fits in int32).
func ldoVoltageToDrefMul(unit uint8, voltageMV uint32) (dref, mul uint8, useRail bool) {
	k1000, vos1000, c1000 := ldoEfuseConstants(unit)

	// Iterate all dref/mul combinations and keep the closest match.
	var bestDiff int32 = 0x7FFFFFFF
	for d := uint8(0); d < 16; d++ {
		var vref20 int32 // Vref * 20
		if d < 9 {
			vref20 = 10 + int32(d)
		} else {
			vref20 = 20 + int32(d-9)*2
		}
		for m := uint8(0); m < 8; m++ {
			vout := (vref20*k1000 + 20*vos1000) * (4000 + int32(m)*c1000)
			diff := int32(voltageMV)*80000 - vout
			if diff < 0 {
				diff = -diff
			}
			if diff < bestDiff {
				bestDiff = diff
				dref, mul = d, m
			}
		}
	}
	return dref, mul, voltageMV == ldoRailVoltageMV
}

// ldoEfuseConstants returns the per-chip calibration constants (K, Vos, C),
// scaled by 1000.  Without calibration (eFuse block version 0) the defaults
// K=1.0, Vos=0, C=1.0 are used.
func ldoEfuseConstants(unit uint8) (k1000, vos1000, c1000 int32) {
	k1000, vos1000, c1000 = 1000, 0, 1000
	if ldoBlkVersion() == 0 {
		return
	}

	var k, vos, c uint32
	sys3 := esp.EFUSE.RD_MAC_SYS_3.Get()
	sys4 := esp.EFUSE.RD_MAC_SYS_4.Get()
	switch unit {
	case 2: // VO3: RD_MAC_SYS_3 k[13:6], vos[19:14], c[25:20]
		k = (sys3 >> 6) & 0xFF
		vos = (sys3 >> 14) & 0x3F
		c = (sys3 >> 20) & 0x3F
	case 3: // VO4: k = sys4[1:0]<<6 | sys3[31:26], vos = sys4[7:2], c = sys4[13:8]
		k = (sys4&0x3)<<6 | (sys3>>26)&0x3F
		vos = (sys4 >> 2) & 0x3F
		c = (sys4 >> 8) & 0x3F
	default: // VO1/VO2 have no K/Vos/C fields; defaults apply.
		return
	}
	if k != 0 { // signed 8-bit, bias 975
		if k&0x80 != 0 {
			k1000 = -int32(k&0x7F) + 975
		} else {
			k1000 = int32(k) + 975
		}
	}
	if vos != 0 { // signed 6-bit, bias −3
		if vos&0x20 != 0 {
			vos1000 = -int32(vos&0x1F) - 3
		} else {
			vos1000 = int32(vos) - 3
		}
	}
	if c != 0 { // signed 6-bit, bias 990
		if c&0x20 != 0 {
			c1000 = -int32(c&0x1F) + 990
		} else {
			c1000 = int32(c) + 990
		}
	}
	return
}

// ldoBlkVersion returns the eFuse BLOCK2 version (major*100 + minor), read
// from RD_MAC_SYS_2 bits [12:11] (major) and [10:8] (minor).
func ldoBlkVersion() uint32 {
	sys2 := esp.EFUSE.RD_MAC_SYS_2.Get()
	major := (sys2 >> 11) & 0x3
	minor := (sys2 >> 8) & 0x7
	return major*100 + minor
}
