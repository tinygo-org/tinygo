package descriptor

import (
	"internal/binary"
)

/* Endpoint Descriptor
USB 2.0 Specification: 9.6.6 Endpoint
*/

const (
	TransferTypeControl uint8 = iota
	TransferTypeIsochronous
	TransferTypeBulk
	TransferTypeInterrupt
)

var endpointEP1IN = [endpointTypeLen]byte{
	endpointTypeLen,
	TypeEndpoint,
	0x81, // EndpointAddress
	0x03, // Attributes
	0x10, // MaxPacketSizeL
	0x00, // MaxPacketSizeH
	0x10, // Interval
}

var EndpointEP1IN = EndpointType{
	data: endpointEP1IN[:],
}

var endpointEP1OUT = [endpointTypeLen]byte{
	endpointTypeLen,
	TypeEndpoint,
	0x01, // EndpointAddress
	0x02, // Attributes
	0x40, // MaxPacketSizeL
	0x00, // MaxPacketSizeH
	0x00, // Interval
}

var EndpointEP1OUT = EndpointType{
	data: endpointEP1OUT[:],
}

var endpointEP2IN = [endpointTypeLen]byte{
	endpointTypeLen,
	TypeEndpoint,
	0x82, // EndpointAddress
	0x02, // Attributes
	0x40, // MaxPacketSizeL
	0x00, // MaxPacketSizeH
	0x00, // Interval
}

var EndpointEP2IN = EndpointType{
	data: endpointEP2IN[:],
}

var endpointEP2OUT = [endpointTypeLen]byte{
	endpointTypeLen,
	TypeEndpoint,
	0x02, // EndpointAddress
	0x03, // Attributes
	0x40, // MaxPacketSizeL
	0x00, // MaxPacketSizeH
	0x01, // Interval
}

var EndpointEP2OUT = EndpointType{
	data: endpointEP2OUT[:],
}

var endpointEP3IN = [endpointTypeLen]byte{
	endpointTypeLen,
	TypeEndpoint,
	0x83, // EndpointAddress
	0x03, // Attributes
	0x40, // MaxPacketSizeL
	0x00, // MaxPacketSizeH
	0x01, // Interval
}

var EndpointEP3IN = EndpointType{
	data: endpointEP3IN[:],
}

// Mass Storage Class bulk in endpoint
var endpointEP5IN = [endpointTypeLen]byte{
	endpointTypeLen,
	TypeEndpoint,
	0x85,             // EndpointAddress
	TransferTypeBulk, // Attributes
	0x40,             // MaxPacketSizeL (64 bytes)
	0x00,             // MaxPacketSizeH
	0x00,             // Interval
}

var EndpointEP5IN = EndpointType{
	data: endpointEP5IN[:],
}

// Mass Storage Class bulk out endpoint
var endpointEP4OUT = [endpointTypeLen]byte{
	endpointTypeLen,
	TypeEndpoint,
	0x04,             // EndpointAddress
	TransferTypeBulk, // Attributes
	0x40,             // MaxPacketSizeL (64 bytes)
	0x00,             // MaxPacketSizeH
	0x00,             // Interval
}

var EndpointEP4OUT = EndpointType{
	data: endpointEP4OUT[:],
}

// Aliases for easier reuse
var (
	EndpointCDCACMIN = &EndpointEP1IN
	EndpointCDCOUT   = &EndpointEP1OUT
	EndpointCDCIN    = &EndpointEP2IN
	EndpointHIDOUT   = &EndpointEP2OUT
	EndpointHIDIN    = &EndpointEP3IN
	EndpointMSCOUT   = &EndpointEP4OUT
	EndpointMSCIN    = &EndpointEP5IN
)

const (
	endpointTypeLen = 7
)

type EndpointType struct {
	data []byte
}

func (d EndpointType) Bytes() []byte {
	return d.data
}

func (d EndpointType) Length(v uint8) {
	d.data[0] = byte(v)
}

func (d EndpointType) Type(v uint8) {
	d.data[1] = byte(v)
}

func (d EndpointType) EndpointAddress(v uint8) {
	d.data[2] = byte(v)
}

func (d EndpointType) Attributes(v uint8) {
	d.data[3] = byte(v)
}

func (d EndpointType) MaxPacketSize(v uint16) {
	binary.LittleEndian.PutUint16(d.data[4:6], v)
}

func (d EndpointType) Interval(v uint8) {
	d.data[6] = byte(v)
}

func (d EndpointType) GetMaxPacketSize() uint16 {
	return binary.LittleEndian.Uint16(d.data[4:6])
}
