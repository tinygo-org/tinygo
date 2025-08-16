package reflect

import (
	"internal/reflectlite"
	"unsafe"
)

type Value struct {
	reflectlite.Value
}

func Indirect(v Value) Value {
	return Value{reflectlite.Indirect(v.Value)}
}

func ValueOf(i interface{}) Value {
	return Value{reflectlite.ValueOf(i)}
}

func (v Value) Type() Type {
	return toType(v.Value.Type())
}

func (v Value) Addr() Value {
	return Value{v.Value.Addr()}
}

func (v Value) Slice(i, j int) Value {
	return Value{v.Value.Slice(i, j)}
}

func (v Value) Slice3(i, j, k int) Value {
	return Value{v.Value.Slice3(i, j, k)}
}

func (v Value) Elem() Value {
	return Value{v.Value.Elem()}
}

// Field returns the value of the i'th field of this struct.
func (v Value) Field(i int) Value {
	return Value{v.Value.Field(i)}
}

func (v Value) Index(i int) Value {
	return Value{v.Value.Index(i)}
}

func (v Value) MapKeys() []Value {
	keys := v.Value.MapKeys()
	return *(*[]Value)(unsafe.Pointer(&keys))
}

func (v Value) MapIndex(key Value) Value {
	return Value{v.Value.MapIndex(key.Value)}
}

func (v Value) MapRange() *MapIter {
	return (*MapIter)(v.Value.MapRange())
}

type MapIter reflectlite.MapIter

func (it *MapIter) Key() Value {
	return Value{((*reflectlite.MapIter)(it)).Key()}
}

func (v Value) SetIterKey(iter *MapIter) {
	v.Value.SetIterKey((*reflectlite.MapIter)(iter))
}

func (it *MapIter) Value() Value {
	return Value{((*reflectlite.MapIter)(it)).Value()}
}

func (v Value) SetIterValue(iter *MapIter) {
	v.Value.SetIterValue((*reflectlite.MapIter)(iter))
}

func (it *MapIter) Next() bool {
	return ((*reflectlite.MapIter)(it)).Next()
}

func (it *MapIter) Reset(v Value) {
	(*reflectlite.MapIter)(it).Reset(v.Value)
}

func (v Value) Set(x Value) {
	v.Value.Set(x.Value)
}

func (v Value) CanConvert(t Type) bool {
	return v.Value.CanConvert(toRawType(t))
}

func (v Value) Convert(t Type) Value {
	return Value{v.Value.Convert(toRawType(t))}
}

func MakeSlice(typ Type, len, cap int) Value {
	return Value{reflectlite.MakeSlice(toRawType(typ), len, cap)}
}

func Zero(typ Type) Value {
	return Value{reflectlite.Zero(toRawType(typ))}
}

// New is the reflect equivalent of the new(T) keyword, returning a pointer to a
// new value of the given type.
func New(typ Type) Value {
	return Value{reflectlite.New(toRawType(typ))}
}

type ValueError = reflectlite.ValueError

// Copy copies the contents of src into dst until either
// dst has been filled or src has been exhausted.
func Copy(dst, src Value) int {
	return reflectlite.Copy(dst.Value, src.Value)
}

// Append appends the values x to a slice s and returns the resulting slice.
// As in Go, each x's value must be assignable to the slice's element type.
func Append(v Value, x ...Value) Value {
	y := *(*[]reflectlite.Value)(unsafe.Pointer(&x))
	return Value{reflectlite.Append(v.Value, y...)}
}

// AppendSlice appends a slice t to a slice s and returns the resulting slice.
// The slices s and t must have the same element type.
func AppendSlice(s, t Value) Value {
	return Value{reflectlite.AppendSlice(s.Value, t.Value)}
}

func (v Value) SetMapIndex(key, elem Value) {
	v.Value.SetMapIndex(key.Value, elem.Value)
}

// FieldByIndex returns the nested field corresponding to index.
func (v Value) FieldByIndex(index []int) Value {
	return Value{v.Value.FieldByIndex(index)}
}

// FieldByIndexErr returns the nested field corresponding to index.
func (v Value) FieldByIndexErr(index []int) (Value, error) {
	out, err := v.Value.FieldByIndexErr(index)
	return Value{out}, err
}

func (v Value) FieldByName(name string) Value {
	return Value{v.Value.FieldByName(name)}
}

func (v Value) FieldByNameFunc(match func(string) bool) Value {
	return Value{v.Value.FieldByNameFunc(match)}
}

type SelectDir int

const (
	_             SelectDir = iota
	SelectSend              // case Chan <- Send
	SelectRecv              // case <-Chan:
	SelectDefault           // default
)

type SelectCase struct {
	Dir  SelectDir // direction of case
	Chan Value     // channel to use (for send or receive)
	Send Value     // value to send (for send)
}

func Select(cases []SelectCase) (chosen int, recv Value, recvOK bool) {
	panic("unimplemented: reflect.Select")
}

func (v Value) Send(x Value) {
	panic("unimplemented: reflect.Value.Send()")
}

func (v Value) TrySend(x Value) bool {
	panic("unimplemented: reflect.Value.TrySend()")
}

func (v Value) Close() {
	panic("unimplemented: reflect.Value.Close()")
}

// MakeMap creates a new map with the specified type.
func MakeMap(typ Type) Value {
	return Value{reflectlite.MakeMap(toRawType(typ))}
}

// MakeMapWithSize creates a new map with the specified type and initial space
// for approximately n elements.
func MakeMapWithSize(typ Type, n int) Value {
	return Value{reflectlite.MakeMapWithSize(toRawType(typ), n)}
}

func (v Value) Call(in []Value) []Value {
	panic("unimplemented: (reflect.Value).Call()")
}

func (v Value) CallSlice(in []Value) []Value {
	panic("unimplemented: (reflect.Value).CallSlice()")
}

func (v Value) Equal(u Value) bool {
	return v.Value.Equal(u.Value)
}

func (v Value) Method(i int) Value {
	panic("unimplemented: (reflect.Value).Method()")
}

func (v Value) MethodByName(name string) Value {
	panic("unimplemented: (reflect.Value).MethodByName()")
}

func (v Value) Recv() (x Value, ok bool) {
	panic("unimplemented: (reflect.Value).Recv()")
}

func (v Value) TryRecv() (x Value, ok bool) {
	panic("unimplemented: (reflect.Value).TryRecv()")
}

func NewAt(typ Type, p unsafe.Pointer) Value {
	panic("unimplemented: reflect.New()")
}

// Deprecated: Use unsafe.Slice or unsafe.SliceData instead.
type SliceHeader struct {
	Data uintptr
	Len  intw
	Cap  intw
}

// Deprecated: Use unsafe.String or unsafe.StringData instead.
type StringHeader struct {
	Data uintptr
	Len  intw
}

// Verify SliceHeader and StringHeader sizes.
// See https://github.com/tinygo-org/tinygo/pull/4156
// and https://github.com/tinygo-org/tinygo/issues/1284.
var (
	_ [unsafe.Sizeof([]byte{})]byte = [unsafe.Sizeof(SliceHeader{})]byte{}
	_ [unsafe.Sizeof("")]byte       = [unsafe.Sizeof(StringHeader{})]byte{}
)
