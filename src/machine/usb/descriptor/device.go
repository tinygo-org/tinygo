package descriptor

import (
	"internal/binary"
)

const (
	deviceTypeLen = 18
)

type DeviceType struct {
	data []byte
}

func (d DeviceType) Bytes() []byte {
	return d.data
}

func (d DeviceType) Length(v uint8) {
	d.data[0] = byte(v)
}

func (d DeviceType) Type(v uint8) {
	d.data[1] = byte(v)
}

func (d DeviceType) USB(v uint16) {
	binary.LittleEndian.PutUint16(d.data[2:4], v)
}

func (d DeviceType) DeviceClass(v uint8) {
	d.data[4] = byte(v)
}

func (d DeviceType) DeviceSubClass(v uint8) {
	d.data[5] = byte(v)
}

func (d DeviceType) DeviceProtocol(v uint8) {
	d.data[6] = byte(v)
}

func (d DeviceType) MaxPacketSize0(v uint8) {
	d.data[7] = byte(v)
}

func (d DeviceType) VendorID(v uint16) {
	binary.LittleEndian.PutUint16(d.data[8:10], v)
}

func (d DeviceType) ProductID(v uint16) {
	binary.LittleEndian.PutUint16(d.data[10:12], v)
}

func (d DeviceType) Device(v uint16) {
	binary.LittleEndian.PutUint16(d.data[12:14], v)
}

func (d DeviceType) Manufacturer(v uint8) {
	d.data[14] = byte(v)
}

func (d DeviceType) Product(v uint8) {
	d.data[15] = byte(v)
}

func (d DeviceType) SerialNumber(v uint8) {
	d.data[16] = byte(v)
}

func (d DeviceType) NumConfigurations(v uint8) {
	d.data[17] = byte(v)
}
