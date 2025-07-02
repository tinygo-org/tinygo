package scsi

import (
	"encoding/binary"
	"fmt"
)

type Cmd struct {
	Data []byte
}

func (c *Cmd) CmdType() CmdType {
	return CmdType(c.Data[0])
}

func (c *Cmd) BlockCount() uint32 {
	return uint32(binary.BigEndian.Uint16(c.Data[7:9]))
}

func (c *Cmd) LBA() uint32 {
	return binary.BigEndian.Uint32(c.Data[2:6])
}

func (c Cmd) String() string {
	cmdType := c.CmdType()
	switch cmdType {
	case CmdRead:
		return fmt.Sprintf("%-28s LBA: % 3d, Block Count: %d", cmdType, c.LBA(), c.BlockCount())
	case CmdWrite:
		return fmt.Sprintf("%-28s LBA: % 3d, Block Count: %d", cmdType, c.LBA(), c.BlockCount())
	default:
		return fmt.Sprintf("%-28s % x", cmdType, c.Data)
	}
}

type CmdType uint8

const (
	CmdTestUnitReady             CmdType = 0x00 // TEST UNIT READY is used to determine if a device is ready to transfer data (read/write). The device does not perform a self-test operation
	CmdRequestSense              CmdType = 0x03 // REQUEST SENSE returns the current sense data (status or error information)
	CmdInquiry                   CmdType = 0x12 // INQUIRY is used to obtain basic information from a target device
	CmdModeSelect6               CmdType = 0x15 // MODE SELECT (6) provides a means for the application client to specify medium, logical unit, or peripheral device parameters to the device server
	CmdModeSelect10              CmdType = 0x55 // MODE SELECT (10) provides a means for the application client to specify medium, logical unit, or peripheral device parameters to the device server
	CmdModeSense6                CmdType = 0x1A // MODE SENSE (6) provides a means for a device server to report parameters to an application client
	CmdModeSense10               CmdType = 0x5A // MODE SENSE (10) provides a means for a device server to report parameters to an application client with 64-bit logical block addressing
	CmdStartStopUnit             CmdType = 0x1B // START STOP UNIT is used to start or stop the medium in a device server
	CmdPreventAllowMediumRemoval CmdType = 0x1E // PREVENT ALLOW MEDIUM REMOVAL is used to prevent or allow the removal of storage medium from a device server
	CmdReadFormatCapacity        CmdType = 0x23 // READ FORMAT CAPACITY allows the Host to request a list of the possible format capacities for an installed writable media
	CmdReadCapacity              CmdType = 0x25 // READ CAPACITY command is used to obtain data capacity information from a target device
	CmdRead                      CmdType = 0x28 // READ (10) requests that the device server read the specified logical block(s) and transfer them to the data-in buffer
	CmdWrite                     CmdType = 0x2A // WRITE (10) requests that the device server transfer the specified logical block(s) from the data-out buffer and write them
	CmdUnmap                     CmdType = 0x42 // UNMAP command is used to inform the device server that the specified logical block(s) are no longer in use
)

func (c CmdType) String() string {
	switch c {
	case CmdTestUnitReady:
		return "TEST UNIT READY"
	case CmdRequestSense:
		return "REQUEST SENSE"
	case CmdInquiry:
		return "INQUIRY"
	case CmdModeSelect6:
		return "MODE SELECT (6)"
	case CmdModeSelect10:
		return "MODE SELECT (10)"
	case CmdModeSense6:
		return "MODE SENSE (6)"
	case CmdModeSense10:
		return "MODE SENSE (10)"
	case CmdStartStopUnit:
		return "START STOP UNIT"
	case CmdPreventAllowMediumRemoval:
		return "PREVENT ALLOW MEDIUM REMOVAL"
	case CmdReadFormatCapacity:
		return "READ FORMAT CAPACITY"
	case CmdReadCapacity:
		return "READ CAPACITY"
	case CmdRead:
		return "READ (10)"
	case CmdWrite:
		return "WRITE (10)"
	case CmdUnmap:
		return "UNMAP"
	default:
		return fmt.Sprintf("Unknown Command (0x%0x)", byte(c))
	}
}

type Sense uint8

const (
	// 4.5.6 Sense key and sense code definitions
	// https://www.t10.org/ftp/t10/document.08/08-309r0.pdf
	SenseNone           Sense = 0x00 // No specific Sense Key. This indicates no error condition
	SenseRecoveredError Sense = 0x01 // The last command completed successfully, but with some recovery action performed
	SenseNotReady       Sense = 0x02 // The LUN addressed is not ready to be accessed
	SenseMediumError    Sense = 0x03 // The command terminated with an unrecoverable error condition
	SenseHardwareError  Sense = 0x04 // The drive detected an unrecoverable hardware failure while performing the command or during a self test
	SenseIllegalRequest Sense = 0x05 // An illegal parameter was provided in the command descriptor block or the additional parameters
	SenseUnitAttention  Sense = 0x06 // The disk drive may have been reset
	SenseDataProtect    Sense = 0x07 // A read or write command was attempted on a block that is protected from this operation and was not performed
	SenseBlankCheck     Sense = 0x08 // A write-once device or a sequential-access device encountered blank medium or format-defined end-of-data indication while reading or that a write-once device encountered a non-blank medium while writing
	SenseFirmwareError  Sense = 0x09 // Vendor specific sense key
	SenseAbortedCommand Sense = 0x0B // The disk drive aborted the command
	SenseVolumeOverflow Sense = 0x0D // A buffered peripheral device has reached the end of medium partition and data remains in the buffer that has not been written to the medium
	SenseMiscompare     Sense = 0x0E // The source data did not match the data read from the medium
)

type SenseCode uint8

const (
	// SenseNotReady
	SenseCodeMediumNotPresent SenseCode = 0x3A // The storage medium is not present in the device (e.g. empty CD-ROM drive or flash card reader)

	// SenseIllegalRequest
	SenseCodeInvalidCmdOpCode  SenseCode = 0x20 // The command operation code is not supported by the device
	SenseCodeInvalidFieldInCDB SenseCode = 0x24 // The command descriptor block (CDB) contains an invalid field

	// SenseDataProtect
	SenseCodeWriteProtected SenseCode = 0x27 // The media is write protected

	// SenseAbortedCommand
	SenseCodeLUNCommFailure      SenseCode = 0x08 // LUN communication failure
	SenseCodeAbortedCmd          SenseCode = 0x0B // The command was aborted by the device
	SenseCodeMsgReject           SenseCode = 0x43 // The command was rejected by the device
	SenseCodeOverlapCmdAttempted SenseCode = 0x4E // The command was rejected by the device because it was overlapped by another command

	// SenseVolumeOverflow
	SenseCodeLBAOutOfRange SenseCode = 0x21 // The logical block address (LBA) is beyond the end of the volume
)

const (
	InquiryRespLen            = 36
	ModeSense6RespLen         = 4
	ModeSense10RespLen        = 8
	ReadCapacityRespLen       = 8
	ReadFormatCapacityRespLen = 12
	RequestSenseRespLen       = 18
)
