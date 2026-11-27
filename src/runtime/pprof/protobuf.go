package pprof

// protobuf is a minimal protocol buffer encoder for profile.proto.
type protobuf struct {
	data []byte
}

func (b *protobuf) varint(x uint64) {
	for x >= 0x80 {
		b.data = append(b.data, byte(x)|0x80)
		x >>= 7
	}
	b.data = append(b.data, byte(x))
}

func (b *protobuf) tag(field, wireType int) {
	b.varint(uint64(field)<<3 | uint64(wireType))
}

func (b *protobuf) uint64(field int, x uint64) {
	if x == 0 {
		return
	}
	b.tag(field, 0)
	b.varint(x)
}

func (b *protobuf) int64(field int, x int64) {
	b.uint64(field, uint64(x))
}

func (b *protobuf) string(field int, s string) {
	b.tag(field, 2)
	b.varint(uint64(len(s)))
	b.data = append(b.data, s...)
}

func (b *protobuf) packedUint64(field int, xs []uint64) {
	var inner protobuf
	for _, x := range xs {
		inner.varint(x)
	}
	b.tag(field, 2)
	b.varint(uint64(len(inner.data)))
	b.data = append(b.data, inner.data...)
}

func (b *protobuf) packedInt64(field int, xs []int64) {
	var inner protobuf
	for _, x := range xs {
		inner.varint(uint64(x))
	}
	b.tag(field, 2)
	b.varint(uint64(len(inner.data)))
	b.data = append(b.data, inner.data...)
}

func (b *protobuf) message(field int, f func(*protobuf)) {
	var inner protobuf
	f(&inner)
	b.tag(field, 2)
	b.varint(uint64(len(inner.data)))
	b.data = append(b.data, inner.data...)
}
