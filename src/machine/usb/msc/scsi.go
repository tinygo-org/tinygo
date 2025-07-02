package msc

import (
	"encoding/binary"
	"machine/usb"
	"machine/usb/msc/csw"
	"machine/usb/msc/scsi"
)

func (m *msc) scsiCmdBegin() {
	cmd := m.cbw.SCSICmd()
	cmdType := cmd.CmdType()

	// Handle multi-packet commands
	switch cmdType {
	case scsi.CmdRead, scsi.CmdWrite:
		m.scsiCmdReadWrite(cmd)
		return
	case scsi.CmdUnmap:
		m.scsiCmdUnmap(cmd)
		return
	}

	if m.totalBytes > 0 && m.cbw.isOut() {
		// Reject any other multi-packet commands
		if m.totalBytes > m.maxPacketSize {
			m.sendScsiError(csw.StatusFailed, scsi.SenseIllegalRequest, scsi.SenseCodeInvalidCmdOpCode)
			return
		} else {
			// Original comment from TinyUSB:
			// Didn't check for case 9 (Ho > Dn), which requires examining scsi command first
			// but it is OK to just receive data then responded with failed status
		}
	}
	switch cmdType {
	case scsi.CmdTestUnitReady:
		m.scsiTestUnitReady()
	case scsi.CmdReadCapacity:
		m.scsiCmdReadCapacity(cmd)
	case scsi.CmdReadFormatCapacity:
		m.scsiCmdReadFormatCapacity(cmd)
	case scsi.CmdInquiry:
		m.scsiCmdInquiry(cmd)
	case scsi.CmdModeSense6, scsi.CmdModeSense10:
		m.scsiCmdModeSense(cmd)
	case scsi.CmdRequestSense:
		m.scsiCmdRequestSense()
	case scsi.CmdPreventAllowMediumRemoval:
		m.scsiCmdPreventAllowMediumRemoval(cmd)
	default:
		// We don't support this command, error out
		m.sendScsiError(csw.StatusFailed, scsi.SenseIllegalRequest, scsi.SenseCodeInvalidCmdOpCode)
	}

	if len(m.buf) == 0 {
		if m.totalBytes > 0 {
			// 6.7.2 The Thirteen Cases - Case 4 (Hi > Dn)
			// https://usb.org/sites/default/files/usbmassbulk_10.pdf
			m.sendScsiError(csw.StatusFailed, scsi.SenseIllegalRequest, 0)
		} else {
			// 6.7.1 The Thirteen Cases - Case 1 Hn = Dn: all good
			// https://usb.org/sites/default/files/usbmassbulk_10.pdf
			m.state = mscStateStatus
		}
	} else {
		if m.totalBytes == 0 {
			// 6.7.1 The Thirteen Cases - Case 2 (Hn < Di)
			// https://usb.org/sites/default/files/usbmassbulk_10.pdf
			m.sendScsiError(csw.StatusFailed, scsi.SenseIllegalRequest, 0)
		} else {
			// Make sure we don't return more data than the host is expecting
			if m.cbw.transferLength() < uint32(len(m.buf)) {
				m.buf = m.buf[:m.cbw.transferLength()]
			}
			m.sendUSBPacket(m.buf)
		}
	}
}

func (m *msc) scsiDataTransfer(b []byte) bool {
	cmd := m.cbw.SCSICmd()
	cmdType := cmd.CmdType()

	switch cmdType {
	case scsi.CmdWrite, scsi.CmdUnmap:
		if m.readOnly {
			m.sendScsiError(csw.StatusFailed, scsi.SenseDataProtect, scsi.SenseCodeWriteProtected)
			return true
		}
		return m.scsiQueueTask(cmdType, b)
	}

	// Update our sent bytes count to include the just-confirmed bytes
	m.sentBytes += m.queuedBytes

	if m.sentBytes >= m.totalBytes {
		// Transfer complete, send CSW after transfer confirmed
		m.state = mscStateStatus
	} else if cmdType == scsi.CmdRead {
		m.scsiRead(cmd)
	} else {
		// Other multi-packet commands are rejected in m.scsiCmdBegin()
	}

	return true
}

func (m *msc) scsiTestUnitReady() {
	m.resetBuffer(0)
	m.queuedBytes = 0

	// Check if the device is ready
	if !m.ready() {
		// If not ready set sense data
		m.senseKey = scsi.SenseNotReady
		m.addlSenseCode = scsi.SenseCodeMediumNotPresent
		m.addlSenseQualifier = 0x00
	} else {
		m.senseKey = 0
		m.addlSenseCode = 0
		m.addlSenseQualifier = 0
	}
}

func (m *msc) scsiCmdReadCapacity(cmd scsi.Cmd) {
	m.resetBuffer(scsi.ReadCapacityRespLen)
	m.queuedBytes = scsi.ReadCapacityRespLen

	// Last LBA address (big endian)
	binary.BigEndian.PutUint32(m.buf[:4], m.blockCount-1)
	// Block size (big endian)
	binary.BigEndian.PutUint32(m.buf[4:8], m.blockSizeUSB)
}

func (m *msc) scsiCmdReadFormatCapacity(cmd scsi.Cmd) {
	m.resetBuffer(scsi.ReadFormatCapacityRespLen)
	m.queuedBytes = scsi.ReadFormatCapacityRespLen

	// bytes 0-2 - reserved
	m.buf[3] = 8 // Capacity list length

	// Number of blocks (big endian)
	binary.BigEndian.PutUint32(m.buf[4:8], m.blockCount)
	// Block size (24-bit, big endian)
	binary.BigEndian.PutUint32(m.buf[8:12], m.blockSizeUSB)
	// Descriptor Type - formatted media
	m.buf[8] = 2
}

// MODE SENSE(6) / MODE SENSE(10) - Only used here to indicate that the device is write protected
func (m *msc) scsiCmdModeSense(cmd scsi.Cmd) {
	respLen := uint32(scsi.ModeSense6RespLen)
	if cmd.CmdType() == scsi.CmdModeSense10 {
		respLen = scsi.ModeSense10RespLen
	}
	m.resetBuffer(int(respLen))
	m.queuedBytes = respLen

	// The host allows a good amount of leeway in response size
	// Reset total bytes to what we'll actually send
	if m.totalBytes > respLen {
		m.totalBytes = respLen
		m.sendZLP = true
	}

	readOnly := byte(0)
	if m.readOnly {
		readOnly = 0x80
	}

	switch cmd.CmdType() {
	case scsi.CmdModeSense6:
		// byte 0 - Number of bytes after this one
		m.buf[0] = byte(respLen) - 1
		// byte 1 - Medium type (0x00 for direct access block device)
		// Bit 7 indicates write protected
		m.buf[2] = readOnly
		// byte 3 - Block descriptor length: 0 (not supported)
	case scsi.CmdModeSense10:
		// bytes 0-1 - Number of bytes after this one
		m.buf[1] = byte(respLen) - 2
		// byte 2 - Medium type (0x00 for direct access block device)
		// Bit 7 indicates write protected
		m.buf[3] = readOnly
	}
}

// PREVENT/ALLOW MEDIUM REMOVAL - A flash drive doesn't have a removable medium, so this is a no-op
func (m *msc) scsiCmdPreventAllowMediumRemoval(cmd scsi.Cmd) {
	m.resetBuffer(0)
	m.queuedBytes = 0

	// Check if the device is ready
	if !m.ready() {
		// If not ready set sense data
		m.senseKey = scsi.SenseNotReady
		m.addlSenseCode = scsi.SenseCodeMediumNotPresent
		m.addlSenseQualifier = 0x00
	} else {
		m.senseKey = 0
		m.addlSenseCode = 0
		m.addlSenseQualifier = 0
	}

	m.state = mscStateStatus
}

// REQUEST SENSE - Returns error status codes when an error status is sent
func (m *msc) scsiCmdRequestSense() {
	// Set the buffer size to the SCSI sense message size and clear
	m.resetBuffer(scsi.RequestSenseRespLen)
	m.queuedBytes = scsi.RequestSenseRespLen
	m.totalBytes = scsi.RequestSenseRespLen

	// 0x70 - current error, 0x71 - deferred error (not used)
	m.buf[0] = 0xF0 // 0x70 for current error plus 0x80 for valid flag bit
	// byte 1 - reserved
	m.buf[2] = uint8(m.senseKey) & 0x0F // Incorrect Length Indicator bit not supported
	// bytes 3-6 - Information (not used)
	// byte 7 - Additional Sense Length (bytes remaining in the message)
	m.buf[7] = scsi.RequestSenseRespLen - 8
	// bytes 8-11 - Command Specific Information (not used)
	m.buf[12] = byte(m.addlSenseCode) // Additional Sense Code (optional)
	m.buf[13] = m.addlSenseQualifier  // Additional Sense Code Qualifier (optional)
	// bytes 14-17 - reserved

	// Clear sense data after copied to buffer
	m.senseKey = 0
	m.addlSenseCode = 0
	m.addlSenseQualifier = 0
}

func (m *msc) scsiCmdUnmap(cmd scsi.Cmd) {
	// Unmap sends a header in the CBW and a parameter list in the data stage
	// The parameter list has an 8 byte header and 16 bytes per item. If it's less than 24 bytes it's
	// not the format we're expecting and we won't be able to decode it. Same for if there isn't a
	// 8 byte header plus multiples of 16 bytes after that
	paramLen := binary.BigEndian.Uint16(m.cbw.Data[7:9])
	if paramLen < 24 || (paramLen-8)%16 != 0 {
		m.sendScsiError(csw.StatusFailed, scsi.SenseIllegalRequest, scsi.SenseCodeInvalidFieldInCDB)
		return
	}
}

func (m *msc) scsiQueueTask(cmdType scsi.CmdType, b []byte) bool {
	// Check if the incoming data is larger than our buffer
	if int(m.queuedBytes)+len(b) > cap(m.buf) {
		m.sendScsiError(csw.StatusFailed, scsi.SenseIllegalRequest, scsi.SenseCodeInvalidFieldInCDB)
		return true
	}

	// Save the incoming data in our buffer for processing outside of interrupt context.
	if m.taskQueued {
		// If we already have a full task queue we can't accept this data
		m.sendScsiError(csw.StatusFailed, scsi.SenseAbortedCommand, scsi.SenseCodeMsgReject)
		return true
	}

	// Copy the queued task data into our buffer
	start := m.queuedBytes
	end := start + uint32(len(b))
	m.buf = m.buf[:end]
	copy(m.buf[start:end], b)
	m.queuedBytes += uint32(len(b))

	switch cmdType {
	case scsi.CmdWrite:
		// If we're writing data wait until we have a full write block of data that can be processed.
		if m.queuedBytes == uint32(cap(m.blockCache)) {
			m.taskQueued = true
		}
	case scsi.CmdUnmap:
		m.taskQueued = true
	}

	// Don't acknowledge the incoming data until we can process it.
	return !m.taskQueued
}

func (m *msc) sendScsiError(status csw.Status, key scsi.Sense, code scsi.SenseCode) {
	// Generate CSW into m.cswBuf
	residue := m.totalBytes - m.sentBytes

	// Prepare to send CSW
	m.sendZLP = true // Ensure the transaction is signaled as ended before a CSW is sent
	m.respStatus = status
	m.state = mscStateStatus

	// Set the sense data
	m.senseKey = key
	m.addlSenseCode = code
	m.addlSenseQualifier = 0x00 // Not used

	if m.totalBytes > 0 && residue > 0 {
		if m.cbw.isIn() {
			m.stallEndpoint(usb.MSC_ENDPOINT_IN)
		} else {
			m.stallEndpoint(usb.MSC_ENDPOINT_OUT)
		}
	}
}
