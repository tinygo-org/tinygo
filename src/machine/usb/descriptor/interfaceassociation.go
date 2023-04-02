package descriptor

const (
	interfaceAssociationTypeLen = 8
)

type InterfaceAssociationType struct {
	data []byte
}

func (d InterfaceAssociationType) Bytes() []byte {
	return d.data
}

func (d InterfaceAssociationType) Length(v uint8) {
	d.data[0] = byte(v)
}

func (d InterfaceAssociationType) Type(v uint8) {
	d.data[1] = byte(v)
}

func (d InterfaceAssociationType) FirstInterface(v uint8) {
	d.data[2] = byte(v)
}

func (d InterfaceAssociationType) InterfaceCount(v uint8) {
	d.data[3] = byte(v)
}

func (d InterfaceAssociationType) FunctionClass(v uint8) {
	d.data[4] = byte(v)
}

func (d InterfaceAssociationType) FunctionSubClass(v uint8) {
	d.data[5] = byte(v)
}

func (d InterfaceAssociationType) FunctionProtocol(v uint8) {
	d.data[6] = byte(v)
}

func (d InterfaceAssociationType) Function(v uint8) {
	d.data[7] = byte(v)
}
