package descriptor

import (
	"internal/binary"
)

const (
	configurationTypeLen = 9
)

type ConfigurationType struct {
	data []byte
}

func (d ConfigurationType) Bytes() []byte {
	return d.data
}

func (d ConfigurationType) Length(v uint8) {
	d.data[0] = byte(v)
}

func (d ConfigurationType) Type(v uint8) {
	d.data[1] = byte(v)
}

func (d ConfigurationType) TotalLength(v uint16) {
	binary.LittleEndian.PutUint16(d.data[2:4], v)
}

func (d ConfigurationType) NumInterfaces(v uint8) {
	d.data[4] = byte(v)
}

func (d ConfigurationType) ConfigurationValue(v uint8) {
	d.data[5] = byte(v)
}

func (d ConfigurationType) Configuration(v uint8) {
	d.data[6] = byte(v)
}

func (d ConfigurationType) Attributes(v uint8) {
	d.data[7] = byte(v)
}

func (d ConfigurationType) MaxPower(v uint8) {
	d.data[8] = byte(v)
}
