//go:build esp32s3

package machine

import (
	"runtime/volatile"
	"unsafe"
)

const (
	// Base address for eFuse controller (EFUSE_BLKx region in TRM).
	efuseBase = uintptr(0x60007000)

	// EFUSE_*_REG offsets mirror ESP-IDF's efuse_reg.h layout.
	efuseClkReg        = efuseBase + 0x1c8
	efuseConfReg       = efuseBase + 0x1cc
	efuseCmdReg        = efuseBase + 0x1d4
	efuseDacConfReg    = efuseBase + 0x1e8
	efuseWrTimConf1Reg = efuseBase + 0x1f4
	efuseWrTimConf2Reg = efuseBase + 0x1f8
	efuseRdData4Reg    = efuseBase + 0x6c // EFUSE_RD_WR_DIS_REG / RD_DATA4
	efuseRdData5Reg    = efuseBase + 0x70 // EFUSE_RD_REPEAT_DATA1_REG / RD_DATA5
	efuseRdData7Reg    = efuseBase + 0x78 // EFUSE_RD_REPEAT_DATA3_REG / RD_DATA7

	// Read opcode and clock enable bit used by EFUSE HAL (see efuse_ll).
	efuseReadOpCode   = uint32(0x5AA5)
	efuseClkEnBit     = uint32(1 << 16)
	efuseBlkVersionV1 = 1 // EFUSE_BLK_VERSION major version = 1

	// SYSTEM_PERIP_CLK_EN0 register and EFUSE clock gate bit.
	systemPeripClkEn0   = uintptr(0x600C0018)
	systemEfuseClkEnBit = uint32(1 << 14)
)

type Fuse struct{}

// triggerReadSequence performs one eFuse read operation using the
// controller's timing/opcode sequence. This roughly corresponds to
// the low-level logic in the ESP-IDF eFuse HAL (see efuse_ll_* in
// the IDF sources and the "eFuse Manager" docs:
// https://docs.espressif.com/projects/esp-idf/en/latest/esp32s3/api-reference/system/efuse.html).
func (f *Fuse) triggerReadSequence() {
	clk := (*volatile.Register32)(unsafe.Pointer(systemPeripClkEn0))
	clk.Set(clk.Get() | systemEfuseClkEnBit)
	efuseClk := (*volatile.Register32)(unsafe.Pointer(efuseClkReg))
	efuseClk.Set(efuseClk.Get() | efuseClkEnBit)
	for i := 0; i < 50; i++ {
	}
	dac := (*volatile.Register32)(unsafe.Pointer(efuseDacConfReg))
	dac.Set(0x28 | (0xFF << 9))
	(*volatile.Register32)(unsafe.Pointer(efuseWrTimConf1Reg)).Set(0x3000 << 8)
	(*volatile.Register32)(unsafe.Pointer(efuseWrTimConf2Reg)).Set(0x190)
	(*volatile.Register32)(unsafe.Pointer(efuseConfReg)).Set(efuseReadOpCode)
	cmd := (*volatile.Register32)(unsafe.Pointer(efuseCmdReg))
	cmd.Set(1)
	for cmd.Get()&1 != 0 {
	}
}

// readBlock2Data4Data5 reads the EFUSE_BLK2 data words that contain
// ADC calibration and version information. It returns RD_DATA4,
// RD_DATA5 and the decoded block version (BLK_VERSION).
//
// Layout is derived from the ESP32-S3 TRM and IDF eFuse tables.
func (f *Fuse) readBlock2Data4Data5() (data4, data5 uint32, blkVer uint8) {
	for i := 0; i < 20; i++ {
	}
	data4 = (*volatile.Register32)(unsafe.Pointer(efuseRdData4Reg)).Get()
	data5 = (*volatile.Register32)(unsafe.Pointer(efuseRdData5Reg)).Get()
	blkVer = uint8(data4 & 3)
	return data4, data5, blkVer
}

// readBlock2Data7 reads RD_DATA7 from EFUSE_BLK2, which for ADC
// calibration contains additional reference (DIGI_REF) data fields.
func (f *Fuse) readBlock2Data7() uint32 {
	return (*volatile.Register32)(unsafe.Pointer(efuseRdData7Reg)).Get()
}

// ReadAdcCalibBlock2 triggers an eFuse read and returns the raw
// EFUSE_BLK2 words used for ADC calibration (RD_DATA4/5) along
// with the decoded block version. This is a small helper similar
// in spirit to the internal IDF helpers around EFUSE_BLK2.
func (f *Fuse) ReadAdcCalibBlock2() (data4, data5 uint32, blkVer uint8) {
	f.triggerReadSequence()
	return f.readBlock2Data4Data5()
}

// ADC1InitCodeAtten3 extracts the ADC1 INIT_CODE (offset trim) for
// attenuation index 3 (typically 11 dB) from EFUSE_BLK2. This mirrors
// the logic used by ESP-IDF's ADC calibration HAL for ESP32-S3.
//
// The code is built from four differential eFuse fields (diff0..diff3)
// and constant offsets (1850, 90, 70) as described in Espressif's
// internal calibration formulas.
func (f *Fuse) ADC1InitCodeAtten3() (uint32, bool) {
	for try := 0; try < 2; try++ {
		f.triggerReadSequence()
		data4, data5, blkVer := f.readBlock2Data4Data5()
		if blkVer != efuseBlkVersionV1 {
			continue
		}
		diff0 := (data4 >> 21) & 0xFF
		diff1 := (data4 >> 29) | ((data5 & 7) << 3)
		diff2 := (data5 >> 3) & 0x3F
		diff3 := (data5 >> 9) & 0x3F
		icode0 := diff0 + 1850
		icode1 := diff1 + icode0 + 90
		icode2 := diff2 + icode1
		icode3 := diff3 + icode2 + 70
		return icode3, true
	}
	return 0, false
}

// ADC1DigiRefAtten3 reads the digital reference (DIGI_REF) for
// ADC1 at attenuation index 3 from EFUSE_BLK2 / RD_DATA7. This is
// similar to what the ESP-IDF ADC calibration HAL uses when present.
func (f *Fuse) ADC1DigiRefAtten3() (uint32, bool) {
	f.triggerReadSequence()
	_, _, blkVer := f.readBlock2Data4Data5()
	if blkVer != efuseBlkVersionV1 {
		return 0, false
	}
	data7 := f.readBlock2Data7()
	diff3 := (data7 >> 1) & 0xFF
	digiRef := diff3 + 900
	if digiRef == 0 {
		return 0, false
	}
	return digiRef, true
}

// GetEfuseAdcCalBlk2 is a tiny wrapper that exposes the raw EFUSE_BLK2
// ADC calibration words for debugging / inspection from other packages.
func GetEfuseAdcCalBlk2() (data4, data5 uint32, blkVer uint8) {
	return (&Fuse{}).ReadAdcCalibBlock2()
}
