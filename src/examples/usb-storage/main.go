package main

import (
	"machine"
	"machine/usb/msc"
	"time"
)

// Disk geometry.
const (
	sectorSize  = 512
	diskSectors = 128 // 64 KB total
)

var diskData [diskSectors * sectorSize]byte

type ramDisk struct{}

func (r *ramDisk) ReadAt(p []byte, off int64) (int, error) {
	return copy(p, diskData[off:]), nil
}

func (r *ramDisk) WriteAt(p []byte, off int64) (int, error) {
	return copy(diskData[off:], p), nil
}

func (r *ramDisk) Size() int64                        { return int64(diskSectors * sectorSize) }
func (r *ramDisk) WriteBlockSize() int64              { return sectorSize }
func (r *ramDisk) EraseBlockSize() int64              { return sectorSize }
func (r *ramDisk) EraseBlocks(start, len int64) error { return nil }

func init() {
	formatFAT12(diskData[:])
}

// formatFAT12 writes a minimal FAT12 volume boot record and FAT tables so the
// host OS can mount the disk without reformatting.
func formatFAT12(d []byte) {
	// --- Sector 0: Volume Boot Record ---
	s := d[0:]
	s[0] = 0xEB
	s[1] = 0x3C
	s[2] = 0x90 // short JMP + NOP
	copy(s[3:11], "MSDOS5.0")
	// BPB fields (little-endian)
	s[11] = 0x00
	s[12] = 0x02 // bytesPerSector = 512
	s[13] = 0x01 // sectorsPerCluster = 1
	s[14] = 0x01
	s[15] = 0x00 // reservedSectors = 1
	s[16] = 0x02 // numFATs = 2
	s[17] = 0x20
	s[18] = 0x00 // rootEntryCount = 32
	s[19] = 0x80
	s[20] = 0x00 // totalSectors16 = 128
	s[21] = 0xF8 // mediaType = fixed disk
	s[22] = 0x01
	s[23] = 0x00 // sectorsPerFAT = 1
	s[24] = 0x80
	s[25] = 0x00 // sectorsPerTrack = 128
	s[26] = 0x01
	s[27] = 0x00 // numHeads = 1
	// hiddenSectors[28:32] = 0
	// totalSectors32[32:36] = 0
	s[38] = 0x29 // extBootSig
	s[39] = 0x47
	s[40] = 0x4F
	s[41] = 0x30
	s[42] = 0x31                  // volumeID "GO01"
	copy(s[43:54], "TINYGO     ") // volumeLabel (11 bytes)
	copy(s[54:62], "FAT12   ")    // fsType
	s[510] = 0x55
	s[511] = 0xAA // boot sector signature

	// --- Sector 1: FAT1 ---
	// Entry 0 = 0xFF8 (media byte), entry 1 = 0xFFF (EOC); all others = free.
	d[512] = 0xF8
	d[513] = 0xFF
	d[514] = 0xFF

	// --- Sector 2: FAT2 (identical copy) ---
	copy(d[1024:1027], d[512:515])
}

func main() {
	msc.Port(&ramDisk{})
	machine.USBDev.Configure(machine.UARTConfig{})

	for {
		time.Sleep(2 * time.Second)
	}
}
