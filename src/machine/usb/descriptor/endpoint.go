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

type EndpointNumber uint8

const (
	EndpointEP1 EndpointNumber = iota
	EndpointEP2
	EndpointEP3
	EndpointEP4
)

const (
	maxEndpoints = 4
)

var (
	endpointEPIn  = [maxEndpoints][endpointTypeLen]byte{}
	endpointEPOut = [maxEndpoints][endpointTypeLen]byte{}
)

func EndpointIN(ep EndpointNumber, transferType uint8, maxPacketSize uint16, interval uint8) EndpointType {
	e := EndpointType{data: endpointEPIn[ep][:]}
	e.Length(endpointTypeLen)
	e.Type(TypeEndpoint)
	e.EndpointAddress(uint8(ep+1) | 0x80) // EndpointNumber is 0-based, addresses are 1-based
	e.Attributes(transferType)
	e.MaxPacketSize(maxPacketSize)
	e.Interval(interval)
	return e
}

func EndpointOUT(ep EndpointNumber, transferType uint8, maxPacketSize uint16, interval uint8) EndpointType {
	e := EndpointType{data: endpointEPOut[ep][:]}
	e.Length(endpointTypeLen)
	e.Type(TypeEndpoint)
	e.EndpointAddress(uint8(ep + 1)) // EndpointNumber is 0-based, addresses are 1-based
	e.Attributes(transferType)
	e.MaxPacketSize(maxPacketSize)
	e.Interval(interval)
	return e
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

func (d EndpointType) GetMaxPacketSize() uint16 {
	return binary.LittleEndian.Uint16(d.data[4:6])
}
