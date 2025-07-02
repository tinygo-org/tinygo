package msc

import (
	"encoding/binary"
	"errors"
	"fmt"
	"machine"
	"time"
)

var (
	errWriteOutOfBounds = errors.New("WriteAt offset out of bounds")
)

// RegisterBlockDevice registers a BlockDevice provider with the MSC driver
func (m *msc) RegisterBlockDevice(dev machine.BlockDevice) {
	m.dev = dev

	if cap(m.blockCache) != int(dev.WriteBlockSize()) {
		m.blockCache = make([]byte, dev.WriteBlockSize())
		m.buf = make([]byte, dev.WriteBlockSize())
	}

	m.blockSizeRaw = uint32(m.dev.WriteBlockSize())
	m.blockCount = uint32(m.dev.Size()) / m.blockSizeUSB
	// Read/write/erase operations must be aligned to the underlying hardware blocks. In order to align
	// them we assume the provided block device is aligned to the end of the underlying hardware block
	// device and offset all reads/writes by the remaining bytes that don't make up a full block.
	m.blockOffset = uint32(m.dev.Size()) % m.blockSizeUSB
	// FIXME: Figure out what to do if the emulated write block size is larger than the erase block size

	// Set VPD UNMAP fields
	for i := range vpdPages {
		if vpdPages[i].PageCode == 0xb0 {
			// 0xb0 - 5.4.5 Block Limits VPD page (B0h)
			if len(vpdPages[i].Data) >= 28 {
				// Set the OPTIMAL UNMAP GRANULARITY (write blocks per erase block)
				granularity := uint32(dev.EraseBlockSize()) / m.blockSizeUSB
				binary.BigEndian.PutUint32(vpdPages[i].Data[24:28], granularity)
			}
			if len(vpdPages[i].Data) >= 32 {
				// Set the UNMAP GRANULARITY ALIGNMENT (first sector of first full erase block)
				// The unmap granularity alignment is used to calculate an optimal unmap request starting LBA as follows:
				// optimal unmap request starting LBA = (n * OPTIMAL UNMAP GRANULARITY) + UNMAP GRANULARITY ALIGNMENT
				// where n is zero or any positive integer value
				// https://www.seagate.com/files/staticfiles/support/docs/manual/Interface%20manuals/100293068j.pdf

				// We assume the block device is aligned to the end of the underlying block device
				blockOffset := uint32(dev.EraseBlockSize()) % m.blockSizeUSB
				// Set the UGAVALID bit to indicate that the UNMAP GRANULARITY ALIGNMENT is valid
				blockOffset |= 0x80000000
				binary.BigEndian.PutUint32(vpdPages[i].Data[28:32], blockOffset)
			}
			break
		}
	}
}

var _ machine.BlockDevice = (*RecorderDisk)(nil)

// RecorderDisk is a block device that records actions taken on it
type RecorderDisk struct {
	dev  machine.BlockDevice
	log  []RecorderRecord
	last time.Time
	time time.Time
}

type RecorderRecord struct {
	OpCode RecorderOpCode
	Offset int64
	Length int
	Data   []byte
	Time   int64
	valid  bool
}

type RecorderOpCode uint8

const (
	RecorderOpCodeRead RecorderOpCode = iota
	RecorderOpCodeWrite
	RecorderOpCodeEraseBlocks
)

// NewRecorderDisk creates a new RecorderDisk instance
func NewRecorderDisk(dev machine.BlockDevice, count int) *RecorderDisk {
	d := &RecorderDisk{
		dev:  dev,
		log:  make([]RecorderRecord, 0, count),
		last: time.Now(),
	}
	for i := 0; i < count; i++ {
		d.log = append(d.log, RecorderRecord{
			OpCode: RecorderOpCodeRead,
			Offset: 0,
			Length: 0,
			Data:   make([]byte, dev.WriteBlockSize()),
			Time:   0,
		})
	}
	return d
}

func (d *RecorderDisk) Size() int64 {
	return d.dev.Size()
}

func (d *RecorderDisk) WriteBlockSize() int64 {
	return d.dev.WriteBlockSize()
}

func (d *RecorderDisk) EraseBlockSize() int64 {
	return d.dev.EraseBlockSize()
}

func (d *RecorderDisk) EraseBlocks(startBlock, numBlocks int64) error {
	d.Record(RecorderOpCodeEraseBlocks, startBlock, int(numBlocks), []byte{})
	return d.dev.EraseBlocks(startBlock, numBlocks)
}

func (d *RecorderDisk) ReadAt(buffer []byte, offset int64) (int, error) {
	n, err := d.dev.ReadAt(buffer, offset)
	d.Record(RecorderOpCodeRead, offset, n, buffer)
	return n, err
}

func (d *RecorderDisk) WriteAt(buffer []byte, offset int64) (int, error) {
	n, err := d.dev.WriteAt(buffer, offset)
	d.Record(RecorderOpCodeWrite, offset, n, buffer)
	return n, err
}

func (d *RecorderDisk) Record(opCode RecorderOpCode, offset int64, length int, data []byte) {
	n := len(d.log) - 1
	// Shift the log entries up to make room for a new entry
	for i := 0; i < n; i++ {
		d.log[i].OpCode = d.log[i+1].OpCode
		d.log[i].Offset = d.log[i+1].Offset
		d.log[i].Length = d.log[i+1].Length
		d.log[i].Data = d.log[i].Data[:len(d.log[i+1].Data)]
		copy(d.log[i].Data, d.log[i+1].Data)
		d.log[i].Time = d.log[i+1].Time
		d.log[i].valid = d.log[i+1].valid
	}

	// Append the new record
	d.log[n].OpCode = opCode
	d.log[n].Offset = offset
	d.log[n].Length = length
	d.log[n].Data = d.log[n].Data[:len(data)]
	copy(d.log[n].Data, data)
	d.log[n].Time = time.Since(d.time).Microseconds()
	d.time = d.time.Add(time.Since(d.time))
	d.log[n].valid = true
}

func (d *RecorderDisk) ClearLog() {
	for i := range d.log {
		d.log[i].valid = false
	}
	d.time = time.Now()
}

func (d *RecorderDisk) GetLog() []RecorderRecord {
	return d.log
}

func (r *RecorderRecord) String() (string, bool) {
	opCode := "Unknown"
	switch r.OpCode {
	case RecorderOpCodeRead:
		opCode = "Read"
	case RecorderOpCodeWrite:
		opCode = "Write"
	case RecorderOpCodeEraseBlocks:
		opCode = "EraseBlocks"
	}
	return fmt.Sprintf("%s: %05d+%02d t:%d | % 0x", opCode, r.Offset, r.Length, r.Time, r.Data), r.valid
}
