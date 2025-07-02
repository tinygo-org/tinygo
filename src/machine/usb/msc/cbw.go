package msc

import (
	"encoding/binary"
	"machine/usb/msc/csw"
	"machine/usb/msc/scsi"
)

const (
	cbwMsgLen = 31         // Command Block Wrapper (CBW) message length
	Signature = 0x43425355 // "USBC" in little endian
)

type CBW struct {
	HasCmd bool
	Data   []byte
}

func (c *CBW) Tag() uint32 {
	return binary.LittleEndian.Uint32(c.Data[4:8])
}

func (c *CBW) length() int {
	return len(c.Data)
}

func (c *CBW) validLength() bool {
	return len(c.Data) == cbwMsgLen
}

func (c *CBW) validSignature() bool {
	return binary.LittleEndian.Uint32(c.Data[:4]) == Signature
}

func (c *CBW) SCSICmd() scsi.Cmd {
	return scsi.Cmd{Data: c.Data[15:]}
}

func (c *CBW) transferLength() uint32 {
	return binary.LittleEndian.Uint32(c.Data[8:12])
}

// isIn returns true if the command direction is from the device to the host.
func (c *CBW) isIn() bool {
	return c.Data[12]>>7 != 0
}

// isOut returns true if the command direction is from the host to the device.
func (c *CBW) isOut() bool {
	return !c.isIn()
}

func (c *CBW) CSW(status csw.Status, residue uint32, b []byte) {
	// Signature: "USBS" 53425355h (little endian)
	binary.LittleEndian.PutUint32(b[:4], csw.Signature)
	// Tag: (same as CBW)
	copy(b[4:8], c.Data[4:8])
	// Data Residue: (untransferred bytes)
	binary.LittleEndian.PutUint32(b[8:12], residue)
	// Status:
	b[12] = byte(status)
}
