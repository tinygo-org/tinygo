//go:build mimxrt1062

package machine

// Flash driver for the FlexSPI serial NOR flash on Teensy 4.x boards.
// The method follows Teensyduino cores/teensy4/eeprom.c.

import (
	"device/arm"
	"device/nxp"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

const (
	// serial NOR page program granularity
	flashPageSize  = 256
	flashPageWords = flashPageSize / 4

	// memory-mapped (AHB) base address of the flash
	flashBaseAddr = 0x60000000

	eraseBlockSizeValue = 4096 // 4 KiB sector erase
)

// FlexSPI LUT sequence words. The instruction format is in the i.MX RT1060
// Reference Manual, chapter 27.5.7 "Lookup table". Sequence 15 is free.
const (
	flashLUTWriteEnable  = 0x00000406 // CMD 0x06 (write enable)
	flashLUTEraseSector  = 0x08180420 // CMD 0x20, RADDR 24 bits (4K sector erase)
	flashLUTPageProgram0 = 0x08180432 // CMD 0x32, RADDR 24 bits (quad page program)
	flashLUTPageProgram1 = 0x00002201 // WRITE on 4 pads
	flashLUTReadStatus   = 0x24010405 // CMD 0x05, READ 1 byte (status register 1)

	flexspiLUTKey = 0x5AF05AF0
)

// compile-time check for ensuring we fulfill BlockDevice interface
var _ BlockDevice = flashBlockDevice{}

var Flash flashBlockDevice

type flashBlockDevice struct {
}

// staging buffer for one flash page, kept in RAM
var flashPageBuf [flashPageWords]uint32

// ReadAt reads the given number of bytes from the block device.
func (f flashBlockDevice) ReadAt(p []byte, off int64) (n int, err error) {
	if readAddress(off) > FlashDataEnd() {
		return 0, errFlashCannotReadPastEOF
	}

	data := unsafe.Slice((*byte)(unsafe.Pointer(readAddress(off))), len(p))
	copy(p, data)

	return len(p), nil
}

// WriteAt writes the given number of bytes to the block device. The driver
// programs full pages padded with 0xFF. The destination must be erased.
func (f flashBlockDevice) WriteAt(p []byte, off int64) (n int, err error) {
	start := readAddress(off)
	if start+uintptr(len(p)) > FlashDataEnd() {
		return 0, errFlashCannotWritePastEOF
	}

	for n < len(p) {
		pageAddr := start &^ (flashPageSize - 1)
		offInPage := start - pageAddr
		chunk := flashPageSize - int(offInPage)
		if chunk > len(p)-n {
			chunk = len(p) - n
		}

		for i := range flashPageBuf {
			flashPageBuf[i] = 0xFFFFFFFF
		}
		page := (*[flashPageSize]byte)(unsafe.Pointer(&flashPageBuf))
		copy(page[offInPage:int(offInPage)+chunk], p[n:n+chunk])

		mask := interrupt.Disable()
		flashTransaction(uint32(pageAddr-flashBaseAddr), true)
		interrupt.Restore(mask)
		flashInvalidateDCache(pageAddr, flashPageSize)

		start += uintptr(chunk)
		n += chunk
	}

	return n, nil
}

// Size returns the number of bytes in this block device.
func (f flashBlockDevice) Size() int64 {
	return int64(FlashDataEnd() - FlashDataStart())
}

const writeBlockSize = flashPageSize

// WriteBlockSize returns the block size in which data can be written to
// memory. It can be used by a client to optimize writes, non-aligned writes
// should always work correctly.
func (f flashBlockDevice) WriteBlockSize() int64 {
	return writeBlockSize
}

func eraseBlockSize() int64 {
	return eraseBlockSizeValue
}

// EraseBlockSize returns the smallest erasable area on this particular chip
// in bytes. This is used for the block size in EraseBlocks.
func (f flashBlockDevice) EraseBlockSize() int64 {
	return eraseBlockSize()
}

// EraseBlocks erases the given number of blocks. An implementation may
// transparently coalesce ranges of blocks into larger bundles if the chip
// supports this. The start and len parameters are in block numbers, use
// EraseBlockSize to map addresses to blocks.
func (f flashBlockDevice) EraseBlocks(start, length int64) error {
	addr := readAddress(start * f.EraseBlockSize())
	if addr+uintptr(length)*eraseBlockSizeValue > FlashDataEnd() {
		return errFlashCannotErasePastEOF
	}

	for i := int64(0); i < length; i++ {
		mask := interrupt.Disable()
		flashTransaction(uint32(addr-flashBaseAddr), false)
		interrupt.Restore(mask)
		flashInvalidateDCache(addr, eraseBlockSizeValue)
		addr += eraseBlockSizeValue
	}

	return nil
}

// return the correct address to be used for reads
func readAddress(off int64) uintptr {
	return FlashDataStart() + uintptr(off)
}

// flashInvalidateDCache removes cached copies of the given flash range.
// The operations are no-ops while the data cache is off.
func flashInvalidateDCache(addr, size uintptr) {
	// DCIMVAC register, see Arm DDI 0403 (Armv7-M ARM) section B3.2.2
	const dcimvacAddr = 0xE000EF5C
	dcimvac := (*volatile.Register32)(unsafe.Pointer(uintptr(dcimvacAddr)))
	arm.Asm("dsb")
	for a := addr &^ 31; a < addr+size; a += 32 {
		dcimvac.Set(uint32(a))
	}
	arm.Asm("dsb")
	arm.Asm("isb")
}

// flashTransaction erases the 4 KiB sector at offset, or programs one page
// from flashPageBuf at offset. Call it with interrupts disabled.
//
//go:section .ramfuncs
//go:nobounds
func flashTransaction(offset uint32, program bool) {
	// This code runs from RAM. The flash gives no data during the operation.
	// Use only volatile loads and stores, the compiler inlines them.
	fs := nxp.FLEXSPI

	// unlock the LUT and load sequence 15 with "write enable"
	volatile.StoreUint32(&fs.LUTKEY.Reg, flexspiLUTKey)
	volatile.StoreUint32(&fs.LUTCR.Reg, nxp.FlexSPI_LUTCR_UNLOCK_Msk)
	volatile.StoreUint32(&fs.LUT[60].Reg, flashLUTWriteEnable)
	volatile.StoreUint32(&fs.LUT[61].Reg, 0)
	volatile.StoreUint32(&fs.LUT[62].Reg, 0)
	volatile.StoreUint32(&fs.LUT[63].Reg, 0)

	// issue write enable
	volatile.StoreUint32(&fs.IPCR0.Reg, 0)
	volatile.StoreUint32(&fs.IPCR1.Reg, 15<<nxp.FlexSPI_IPCR1_ISEQID_Pos)
	volatile.StoreUint32(&fs.IPCMD.Reg, nxp.FlexSPI_IPCMD_TRG_Msk)
	for volatile.LoadUint32(&fs.INTR.Reg)&nxp.FlexSPI_INTR_IPCMDDONE_Msk == 0 {
	}
	volatile.StoreUint32(&fs.INTR.Reg, nxp.FlexSPI_INTR_IPCMDDONE_Msk)

	if program {
		// program one page from flashPageBuf
		volatile.StoreUint32(&fs.LUT[60].Reg, flashLUTPageProgram0)
		volatile.StoreUint32(&fs.LUT[61].Reg, flashLUTPageProgram1)
		volatile.StoreUint32(&fs.IPTXFCR.Reg, nxp.FlexSPI_IPTXFCR_CLRIPTXF_Msk)
		volatile.StoreUint32(&fs.IPCR0.Reg, offset)
		volatile.StoreUint32(&fs.IPCR1.Reg, 15<<nxp.FlexSPI_IPCR1_ISEQID_Pos|flashPageSize)
		volatile.StoreUint32(&fs.IPCMD.Reg, nxp.FlexSPI_IPCMD_TRG_Msk)
		i := 0
		for {
			intr := volatile.LoadUint32(&fs.INTR.Reg)
			if intr&nxp.FlexSPI_INTR_IPCMDDONE_Msk != 0 {
				break
			}
			if intr&nxp.FlexSPI_INTR_IPTXWE_Msk != 0 {
				// fill the TX FIFO up to its 8-byte watermark
				if i < flashPageWords {
					volatile.StoreUint32(&fs.TFDR[0].Reg, flashPageBuf[i])
					volatile.StoreUint32(&fs.TFDR[1].Reg, flashPageBuf[i+1])
					i += 2
				}
				volatile.StoreUint32(&fs.INTR.Reg, nxp.FlexSPI_INTR_IPTXWE_Msk)
			}
		}
		volatile.StoreUint32(&fs.INTR.Reg,
			nxp.FlexSPI_INTR_IPCMDDONE_Msk|nxp.FlexSPI_INTR_IPTXWE_Msk)
	} else {
		// erase the 4 KiB sector at offset
		volatile.StoreUint32(&fs.LUT[60].Reg, flashLUTEraseSector)
		volatile.StoreUint32(&fs.IPCR0.Reg, offset)
		volatile.StoreUint32(&fs.IPCR1.Reg, 15<<nxp.FlexSPI_IPCR1_ISEQID_Pos)
		volatile.StoreUint32(&fs.IPCMD.Reg, nxp.FlexSPI_IPCMD_TRG_Msk)
		for volatile.LoadUint32(&fs.INTR.Reg)&nxp.FlexSPI_INTR_IPCMDDONE_Msk == 0 {
		}
		volatile.StoreUint32(&fs.INTR.Reg, nxp.FlexSPI_INTR_IPCMDDONE_Msk)
	}

	// poll the status register until the write-in-progress bit clears
	volatile.StoreUint32(&fs.LUT[60].Reg, flashLUTReadStatus)
	volatile.StoreUint32(&fs.LUT[61].Reg, 0)
	for {
		volatile.StoreUint32(&fs.IPRXFCR.Reg, nxp.FlexSPI_IPRXFCR_CLRIPRXF_Msk)
		volatile.StoreUint32(&fs.IPCR0.Reg, 0)
		volatile.StoreUint32(&fs.IPCR1.Reg, 15<<nxp.FlexSPI_IPCR1_ISEQID_Pos|1)
		volatile.StoreUint32(&fs.IPCMD.Reg, nxp.FlexSPI_IPCMD_TRG_Msk)
		for volatile.LoadUint32(&fs.INTR.Reg)&nxp.FlexSPI_INTR_IPCMDDONE_Msk == 0 {
		}
		volatile.StoreUint32(&fs.INTR.Reg, nxp.FlexSPI_INTR_IPCMDDONE_Msk)
		if volatile.LoadUint32(&fs.RFDR[0].Reg)&0x01 == 0 {
			break
		}
	}

	// the software reset removes stale data from the AHB read buffers
	volatile.StoreUint32(&fs.MCR0.Reg,
		volatile.LoadUint32(&fs.MCR0.Reg)|nxp.FlexSPI_MCR0_SWRESET_Msk)
	for volatile.LoadUint32(&fs.MCR0.Reg)&nxp.FlexSPI_MCR0_SWRESET_Msk != 0 {
	}
}
