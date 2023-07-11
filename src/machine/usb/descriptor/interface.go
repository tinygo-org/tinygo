package descriptor

const (
	interfaceTypeLen = 9
)

type InterfaceType struct {
	data []byte
}

func (d InterfaceType) Bytes() []byte {
	return d.data
}

func (d InterfaceType) Length(v uint8) {
	d.data[0] = byte(v)
}

func (d InterfaceType) Type(v uint8) {
	d.data[1] = byte(v)
}

func (d InterfaceType) InterfaceNumber(v uint8) {
	d.data[2] = byte(v)
}

func (d InterfaceType) AlternateSetting(v uint8) {
	d.data[3] = byte(v)
}

func (d InterfaceType) NumEndpoints(v uint8) {
	d.data[4] = byte(v)
}

func (d InterfaceType) InterfaceClass(v uint8) {
	d.data[5] = byte(v)
}

func (d InterfaceType) InterfaceSubClass(v uint8) {
	d.data[6] = byte(v)
}

func (d InterfaceType) InterfaceProtocol(v uint8) {
	d.data[7] = byte(v)
}

func (d InterfaceType) Interface(v uint8) {
	d.data[8] = byte(v)
}
