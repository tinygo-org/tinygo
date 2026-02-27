//go:build esp32c3 && !m5stamp_c3

package machine

import (
	"runtime/volatile"
	"unsafe"
)

const (
	// Base address for eFuse controller on ESP32-C3.
	// Matches DR_REG_EFUSE_BASE in IDF (see soc/efuse_reg.h).
	efuseBaseC3 = uintptr(0x60008800)

	// Read register range for EFUSE_BLK2 (SYS_DATA), derived from
	// esp_efuse_utility.c: range_read_addr_blocks[EFUSE_BLK2].
	// BLK2 spans EFUSE_RD_SYS_PART1_DATA0_REG .. DATA7.
	efuseRdBlk2Word0 = efuseBaseC3 + 0xa0 // EFUSE_RD_SYS_PART1_DATA0_REG

	// Calibration layout (see esp_efuse_table.csv for esp32c3):
	//   ADC1_INIT_CODE_ATTEN3  : EFUSE_BLK2, bit 178, len 10
	//   ADC1_CAL_VOL_ATTEN3    : EFUSE_BLK2, bit 218, len 10
	adc1InitCodeAtten3Bit = uint32(178)
	adc1InitCodeAtten3Len = uint32(10)
	adc1CalVolAtten3Bit   = uint32(218)
	adc1CalVolAtten3Len   = uint32(10)
)

type Fuse struct{}

func (f *Fuse) readBlk2Word(index uint32) uint32 {
	reg := (*volatile.Register32)(unsafe.Pointer(efuseRdBlk2Word0 + uintptr(index*4)))
	return reg.Get()
}

func extractBits(val, bit, length uint32) uint32 {
	return (val >> bit) & ((1 << length) - 1)
}

// ADC1InitCodeAtten3 returns the factory-trimmed INIT_CODE (offset) for
// ADC1 at attenuation index 3 (11 dB) from EFUSE_BLK2, following the
// logic from esp_efuse_rtc_calib_get_init_code() for ESP32-C3:
//   - 10-bit field ADC1_INIT_CODE_ATTEN3 (bit 178, len 10),
//   - value is interpreted as unsigned and then +1000 is applied.
func (f *Fuse) ADC1InitCodeAtten3() (uint32, bool) {
	offset := adc1InitCodeAtten3Bit
	length := adc1InitCodeAtten3Len

	wordIdx := offset / 32
	bitInWord := offset % 32

	word := f.readBlk2Word(wordIdx)
	var raw uint32
	if bitInWord+length <= 32 {
		raw = extractBits(word, bitInWord, length)
	} else {
		low := extractBits(word, bitInWord, 32-bitInWord)
		high := extractBits(f.readBlk2Word(wordIdx+1), 0, length-(32-bitInWord))
		raw = low | (high << (32 - bitInWord))
	}

	if raw == 0 {
		return 0, false
	}
	return raw + 1000, true
}

// ADC1DigiRefAtten3 returns the calibration point (DIGI_REF and voltage)
// for ADC1 at attenuation index 3 using the same encoding as
// esp_efuse_rtc_calib_get_cal_voltage() V1 for ESP32-C3:
//   - 10-bit field ADC1_CAL_VOL_ATTEN3 (bit 218, len 10),
//   - signed value around 0..511 with bit9 as sign,
//   - digi_ref = 2000 + signed(cal_vol),
//   - expected calibration voltage is 1370 mV.
//
// We only return digi_ref here; the fixed 1370 mV is used by higher
// layers when needed.
func (f *Fuse) ADC1DigiRefAtten3() (uint32, bool) {
	offset := adc1CalVolAtten3Bit
	length := adc1CalVolAtten3Len

	wordIdx := offset / 32
	bitInWord := offset % 32

	word := f.readBlk2Word(wordIdx)
	var cal uint32
	if bitInWord+length <= 32 {
		cal = extractBits(word, bitInWord, length)
	} else {
		low := extractBits(word, bitInWord, 32-bitInWord)
		high := extractBits(f.readBlk2Word(wordIdx+1), 0, length-(32-bitInWord))
		cal = low | (high << (32 - bitInWord))
	}

	if cal == 0 {
		return 0, false
	}

	// Interpret 10-bit value as signed with bit9 as sign, per IDF logic.
	const signBit = uint32(1 << 9)
	var signed int32
	if cal&signBit != 0 {
		signed = -int32(cal & ^signBit)
	} else {
		signed = int32(cal)
	}

	digi := uint32(int32(2000) + signed)
	if digi == 0 {
		return 0, false
	}
	return digi, true
}
