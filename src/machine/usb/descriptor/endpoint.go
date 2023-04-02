package descriptor

import (
	"encoding/binary"
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

var endpointEP2OUT = [endpointTypeLen]byte{
	endpointTypeLen,
	TypeEndpoint,
	0x02, // EndpointAddress
	0x02, // Attributes
	0x40, // MaxPacketSizeL
	0x00, // MaxPacketSizeH
	0x00, // Interval
}

var EndpointEP2OUT = EndpointType{
	data: endpointEP2OUT[:],
}

var endpointEP3IN = [endpointTypeLen]byte{
	endpointTypeLen,
	TypeEndpoint,
	0x83, // EndpointAddress
	0x02, // Attributes
	0x40, // MaxPacketSizeL
	0x00, // MaxPacketSizeH
	0x00, // Interval
}

var EndpointEP3IN = EndpointType{
	data: endpointEP3IN[:],
}

var endpointEP4IN = [endpointTypeLen]byte{
	endpointTypeLen,
	TypeEndpoint,
	0x84, // EndpointAddress
	0x03, // Attributes
	0x40, // MaxPacketSizeL
	0x00, // MaxPacketSizeH
	0x01, // Interval
}

var EndpointEP4IN = EndpointType{
	data: endpointEP4IN[:],
}

var endpointEP5OUT = [endpointTypeLen]byte{
	endpointTypeLen,
	TypeEndpoint,
	0x05, // EndpointAddress
	0x03, // Attributes
	0x40, // MaxPacketSizeL
	0x00, // MaxPacketSizeH
	0x01, // Interval
}

var EndpointEP5OUT = EndpointType{
	data: endpointEP5OUT[:],
}

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
