package msc

import (
	"errors"
	"machine/usb/msc/csw"
	"machine/usb/msc/scsi"
)

var (
	invalidWriteError = errors.New("invalid write offset or length")
	errBlockNotCached = errors.New("block not cached")
)

func (m *msc) scsiCmdReadWrite(cmd scsi.Cmd) {
	m.cachedBlock = -1
	status := m.validateScsiReadWrite(cmd)
	if status != csw.StatusPassed {
		m.sendScsiError(status, scsi.SenseIllegalRequest, scsi.SenseCodeInvalidCmdOpCode)
	} else if m.transferBytes > 0 {
		if cmd.CmdType() == scsi.CmdRead {
			m.scsiReadNext(cmd)
		} else {
			// WRITE(10) and UNMAP commands don't take any action until the data stage begins
		}
	} else {
		// Zero byte transfer. No practical use case
		m.state = mscStateStatus
	}
}

// Validate SCSI READ(10) and WRITE(10) commands
func (m *msc) validateScsiReadWrite(cmd scsi.Cmd) csw.Status {
	blockCount := cmd.BlockCount()
	// CBW wrapper transfer length
	if m.transferBytes == 0 {
		// If the SCSI command's block count doesn't loosely match the wrapper's transfer length something's wrong
		if blockCount > 0 {
			return csw.StatusPhaseError
		}
		// Zero length transfer. No practical use case, but explicitly not an error according to the spec
		return csw.StatusPassed
	}
	if (cmd.CmdType() == scsi.CmdRead && m.cbw.isOut()) || (cmd.CmdType() == scsi.CmdWrite && m.cbw.isIn()) {
		// If the command is READ(10) and the data direction is from host to device that's a problem
		// 6.7.3 The Thirteen Cases - Case 10 (Ho <> Di)
		// If the command is WRITE(10) and the data direction is from device to host that's also a problem
		// 6.7.2 The Thirteen Cases - Case 8 (Hi <> Do)
		// https://usb.org/sites/default/files/usbmassbulk_10.pdf
		return csw.StatusPhaseError
	}
	if blockCount == 0 {
		// We already checked for zero length transfer above, so this is a problem
		// 6.7.2 The Thirteen Cases - Case 4 (Hi > Dn)
		// https://usb.org/sites/default/files/usbmassbulk_10.pdf
		return csw.StatusFailed
	}
	if m.transferBytes/blockCount == 0 {
		// Block size shouldn't be small enough to round to zero
		// 6.7.2 The Thirteen Cases - Case 7 (Hi < Di) READ(10) or
		// 6.7.3 The Thirteen Cases - Case 13 (Ho < Do) WRITE(10)
		// https://usb.org/sites/default/files/usbmassbulk_10.pdf
		return csw.StatusPhaseError
	}
	return csw.StatusPassed
}

func (m *msc) usbToRawOffset(lba, offset uint32) (int64, int64) {
	// Convert the emulated block address to the underlying hardware block's start and offset
	rawLBA := (lba*m.blockSizeUSB + offset) / m.blockSizeRaw
	rawBlockOffset := int64((lba*m.blockSizeUSB + offset) % m.blockSizeRaw)
	return int64(m.blockOffset + rawLBA*m.blockSizeRaw), rawBlockOffset
}

func (m *msc) readBlock(b []byte, lba, offset uint32) (n int, err error) {
	// Convert the emulated block address to the underlying hardware block's start and offset
	blockStart, blockOffset := m.usbToRawOffset(lba, offset)

	if blockStart != m.cachedBlock {
		return 0, errBlockNotCached
	}
	n = len(m.blockCache) - int(blockOffset)
	if n > len(b) {
		n = len(b)
	}

	copy(b, m.blockCache[blockOffset:])

	return n, nil
}

// fillBlockCache reads the block for the next READ(10) packet. It runs in
// processTasks, and the block is only kept if the command is still current.
func (m *msc) fillBlockCache(gen uint32) {
	state := m.mu.lock()
	if gen != m.cmdGen {
		m.mu.unlock(state)
		return
	}
	cmd := m.cbw.SCSICmd()
	blockStart, _ := m.usbToRawOffset(cmd.LBA(), m.sentBytes)
	if blockStart == m.cachedBlock {
		m.mu.unlock(state)
		return
	}
	m.cachedBlock = -1
	m.mu.unlock(state)

	n, err := m.dev.ReadAt(m.blockCache, blockStart)

	state = m.mu.lock()
	if err == nil && n == len(m.blockCache) && gen == m.cmdGen {
		m.cachedBlock = blockStart
	}
	m.mu.unlock(state)
}

func (m *msc) writeBlock(b []byte, lba, offset uint32) (n int, err error) {
	// Convert the emulated block address to the underlying hardware block's start and offset
	blockStart, blockOffset := m.usbToRawOffset(lba, offset)

	if blockOffset != 0 || len(b) != int(m.blockSizeRaw) {
		return 0, invalidWriteError
	}

	// Write the full block to the underlying device
	n, err = m.dev.WriteAt(b, blockStart)
	n -= int(blockOffset)
	if n > len(b) {
		n = len(b)
	}

	return n, err
}

// scsiReadNext sends the next packet from the block cache, or leaves the
// device read to processTasks so BlockDevice.ReadAt never runs in an interrupt.
func (m *msc) scsiReadNext(cmd scsi.Cmd) {
	blockStart, _ := m.usbToRawOffset(cmd.LBA(), m.sentBytes)
	if blockStart == m.cachedBlock {
		m.scsiRead(cmd)
	} else {
		m.taskCmd = scsi.CmdRead
		m.taskGen = m.cmdGen
		m.taskQueued = true
	}
}

func (m *msc) scsiRead(cmd scsi.Cmd) {
	// Make sure we don't exceed the buffer size
	readEnd := m.transferBytes - m.sentBytes
	if readEnd > m.maxPacketSize {
		readEnd = m.maxPacketSize
	}
	// Resize the buffer to fit the read size
	m.resetBuffer(int(readEnd))

	// Read data from the emulated block device
	n, err := m.readBlock(m.buf[:readEnd], cmd.LBA(), m.sentBytes)
	if err != nil || n == 0 {
		m.sendScsiError(csw.StatusFailed, scsi.SenseNotReady, scsi.SenseCodeMediumNotPresent)
		return
	}

	m.queuedBytes = readEnd
	m.sendUSBPacket(m.buf)
}

// scsiWrite runs in processTasks. It drops the write if a reset or a new
// CBW replaced the command it was queued for.
func (m *msc) scsiWrite(b []byte, gen uint32) {
	state := m.mu.lock()
	if gen != m.cmdGen {
		m.mu.unlock(state)
		return
	}
	if m.readOnly {
		m.sendScsiError(csw.StatusFailed, scsi.SenseDataProtect, scsi.SenseCodeWriteProtected)
		m.mu.unlock(state)
		return
	}
	cmd := m.cbw.SCSICmd()
	lba, offset := cmd.LBA(), m.sentBytes
	m.mu.unlock(state)

	// Write data to the block device
	n, err := m.writeBlock(b, lba, offset)
	state = m.mu.lock()
	defer m.mu.unlock(state)
	if gen != m.cmdGen {
		return
	}
	if err != nil || n < len(b) {
		m.sentBytes += uint32(n)
		m.sendScsiError(csw.StatusFailed, scsi.SenseNotReady, scsi.SenseCodeMediumNotPresent)
	} else {
		m.sentBytes += uint32(len(b))
	}

	if m.sentBytes >= m.transferBytes {
		// Data transfer is complete, send CSW
		m.state = mscStateStatus
		m.run([]byte{}, true)
	}
}
