package msc

import (
	"errors"
	"time"
)

// RamDisk implements machine.BlockDevice in memory.
type RamDisk struct {
	Data       []byte
	BlockSize  int64
	WriteDelay time.Duration // FIXME: Cleanup
}

// NewRamDisk creates a new RamDisk with the given size.
func NewRamDisk(size int64) *RamDisk {
	return &RamDisk{
		Data:      make([]byte, size),
		BlockSize: 512,
	}
}

// ReadAt reads the given number of bytes from the block device.
func (r *RamDisk) ReadAt(p []byte, off int64) (n int, err error) {
	if off >= int64(len(r.Data)) {
		return 0, errors.New("read beyond end of ramdisk")
	}
	n = copy(p, r.Data[off:])
	return n, nil
}

// WriteAt writes the given number of bytes to the block device.
func (r *RamDisk) WriteAt(p []byte, off int64) (n int, err error) {
	if off >= int64(len(r.Data)) {
		return 0, errors.New("write beyond end of ramdisk")
	}
	n = copy(r.Data[off:], p)
	time.Sleep(r.WriteDelay) // FIXME: Cleanup
	if n < len(p) {
		return n, errors.New("write beyond end of ramdisk")
	}
	return n, nil
}

// Size returns the number of bytes in this block device.
func (r *RamDisk) Size() int64 {
	return int64(len(r.Data))
}

// WriteBlockSize returns the block size in which data can be written to
// memory.
func (r *RamDisk) WriteBlockSize() int64 {
	return r.BlockSize
}

// EraseBlockSize returns the smallest erasable area on this particular chip
// in bytes.
func (r *RamDisk) EraseBlockSize() int64 {
	return r.BlockSize
}

// EraseBlocks erases the given number of blocks.
func (r *RamDisk) EraseBlocks(start, len int64) error {
	// Convert block numbers to byte offsets
	startOffset := start * r.EraseBlockSize()
	lengthBytes := len * r.EraseBlockSize()

	if startOffset+lengthBytes > int64(cap(r.Data)) {
		return errors.New("erase beyond end of ramdisk")
	}

	for i := int64(0); i < lengthBytes; i++ {
		r.Data[startOffset+i] = 0xFF
	}
	return nil
}
