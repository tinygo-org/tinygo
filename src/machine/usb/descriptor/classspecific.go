package descriptor

const (
	classSpecificTypeLen = 5
)

type ClassSpecificType struct {
	data []byte
}

func (d ClassSpecificType) Bytes() []byte {
	return d.data[:d.data[0]]
}

func (d ClassSpecificType) Length(v uint8) {
	d.data[0] = byte(v)
}
