//go:build esp32s3

package machine

import (
	"runtime/volatile"
	"unsafe"
)

const (
	efuseBase           = uintptr(0x60007000)
	efuseClkReg         = efuseBase + 0x1c8
	efuseConfReg        = efuseBase + 0x1cc
	efuseCmdReg         = efuseBase + 0x1d4
	efuseDacConfReg     = efuseBase + 0x1e8
	efuseWrTimConf1Reg  = efuseBase + 0x1f4
	efuseWrTimConf2Reg  = efuseBase + 0x1f8
	efuseRdData4Reg     = efuseBase + 0x6c
	efuseRdData5Reg     = efuseBase + 0x70
	efuseRdData7Reg     = efuseBase + 0x78
	efuseReadOpCode     = uint32(0x5AA5)
	efuseClkEnBit       = uint32(1 << 16)
	efuseBlkVersionV1   = 1
	systemPeripClkEn0   = uintptr(0x600C0018)
	systemEfuseClkEnBit = uint32(1 << 14)
)

type Fuse struct{}

var DefaultFuse Fuse

func (f *Fuse) TriggerRead() {
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
	for cmd.Get()&1 != 0 {
	}
}

func (f *Fuse) ReadBlk2Data4Data5() (data4, data5 uint32, blkVer uint8) {
	for i := 0; i < 20; i++ {
	}
	data4 = (*volatile.Register32)(unsafe.Pointer(efuseRdData4Reg)).Get()
	data5 = (*volatile.Register32)(unsafe.Pointer(efuseRdData5Reg)).Get()
	blkVer = uint8(data4 & 3)
	return data4, data5, blkVer
}

func (f *Fuse) ReadBlk2Data7() uint32 {
	return (*volatile.Register32)(unsafe.Pointer(efuseRdData7Reg)).Get()
}

func (f *Fuse) Blk2() (data4, data5 uint32, blkVer uint8) {
	f.TriggerRead()
	return f.ReadBlk2Data4Data5()
}

func (f *Fuse) ADC1InitCodeAtten3() (uint32, bool) {
	for try := 0; try < 2; try++ {
		f.TriggerRead()
		data4, data5, blkVer := f.ReadBlk2Data4Data5()
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

func (f *Fuse) ADC1DigiRefAtten3() (uint32, bool) {
	f.TriggerRead()
	_, _, blkVer := f.ReadBlk2Data4Data5()
	if blkVer != efuseBlkVersionV1 {
		return 0, false
	}
	data7 := f.ReadBlk2Data7()
	diff3 := (data7 >> 1) & 0xFF
	digiRef := diff3 + 900
	if digiRef == 0 {
		return 0, false
	}
	return digiRef, true
}

func GetEfuseAdcCalBlk2() (data4, data5 uint32, blkVer uint8) {
	return DefaultFuse.Blk2()
}
