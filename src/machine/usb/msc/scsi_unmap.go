package msc

import (
	"encoding/binary"
	"machine/usb/msc/csw"
	"machine/usb/msc/scsi"
)

type Error int

const (
	errorLBAOutOfRange Error = iota
)

func (e Error) Error() string {
	switch e {
	case errorLBAOutOfRange:
		return "LBA out of range"
	default:
		return "unknown error"
	}
}

func (m *msc) scsiUnmap(b []byte) {
	// Execute Order 66 (0x42) to wipe out the blocks
	// 3.54 Unmap Command (SBC-4)
	// https://www.seagate.com/files/staticfiles/support/docs/manual/Interface%20manuals/100293068j.pdf
	if m.readOnly {
		m.sendScsiError(csw.StatusFailed, scsi.SenseDataProtect, scsi.SenseCodeWriteProtected)
		return
	}

	// blockDescLen is the remaining length of block descriptors in the message, offset 8 bytes from
	// the start of this packet
	var blockDescLen uint16

	// Decode the parameter list
	msgLen := binary.BigEndian.Uint16(b[:2])
	// Length of the block descriptor portion of the message
	blockDescLen = binary.BigEndian.Uint16(b[2:4])
	// Do some sanity checks on the message lengths (max 3 block descriptors to fit in one 64 byte packet)
	if msgLen < 8 || blockDescLen < 16 || msgLen-blockDescLen != 6 || blockDescLen > (3*16) {
		m.sendScsiError(csw.StatusFailed, scsi.SenseIllegalRequest, scsi.SenseCodeInvalidFieldInCDB)
		return
	}

	// descEnd marks the end of the last full block descriptor in this packet
	descEnd := int(blockDescLen + 8)

	// Unmap the blocks we can from this packet
	for i := 8; i < descEnd; i += 16 {
		err := m.unmapBlocksFromDescriptor(b[i:], uint64(m.blockCount))
		if err != nil {
			// TODO: Might need a better error code here for device errors?
			m.sendScsiError(csw.StatusFailed, scsi.SenseVolumeOverflow, scsi.SenseCodeLBAOutOfRange)
			return
		}
	}

	// FIXME: We need to handle erase block alignment

	m.sentBytes += uint32(len(b))
	if m.sentBytes >= m.totalBytes {
		// Order 66 complete, send CSW to establish galactic empire
		m.state = mscStateStatus
		m.run([]byte{}, true)
	}
}

func (m *msc) unmapBlocksFromDescriptor(b []byte, numBlocks uint64) error {
	blockCount := binary.BigEndian.Uint32(b[8:12])
	if blockCount == 0 {
		// No blocks to unmap. Explicitly not an error per the spec
		return nil
	}
	// This is technically a 64-bit LBA, but we can't address that many bytes
	// let alone blocks, so we just use the lower 32 bits
	lba := binary.BigEndian.Uint32(b[4:8])

	// Make sure the unmap command doesn't extend past the end of the volume
	if lba+blockCount > m.blockCount {
		return errorLBAOutOfRange
	}

	// Convert the emulated block size to the underlying hardware erase block size
	blockStart := int64(lba*m.blockSizeUSB) / m.dev.EraseBlockSize()
	rawBlockCount := int64(blockCount*m.blockSizeUSB) / m.dev.EraseBlockSize()

	// Unmap the blocks
	return m.dev.EraseBlocks(blockStart, rawBlockCount)
}
