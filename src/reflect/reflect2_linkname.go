package reflect

// github.com/modern-go/reflect2 (used by github.com/json-iterator/go) reaches
// into reflect's unexported runtime helpers via //go:linkname. TinyGo's reflect
// does not implement these, so a program that links reflect2 fails at link time
// with "undefined symbol: reflect.mapassign" etc. Provide matching symbols so
// such programs LINK; they panic if actually exercised. reflect2/jsoniter is
// therefore non-functional under TinyGo — callers should use encoding/json,
// which TinyGo's reflect does support. See the skywire TinyGo build notes.

import "unsafe"

// sliceHeader mirrors reflect2's slice header ABI for typedslicecopy.
type reflect2SliceHeader struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}

//go:linkname reflect2_makemap reflect.makemap
func reflect2_makemap(rtype unsafe.Pointer, cap int) unsafe.Pointer {
	panic("reflect.makemap: reflect2/jsoniter is unsupported under TinyGo (use encoding/json)")
}

//go:linkname reflect2_unsafe_New reflect.unsafe_New
func reflect2_unsafe_New(rtype unsafe.Pointer) unsafe.Pointer {
	panic("reflect.unsafe_New: reflect2/jsoniter is unsupported under TinyGo (use encoding/json)")
}

//go:linkname reflect2_unsafe_NewArray reflect.unsafe_NewArray
func reflect2_unsafe_NewArray(rtype unsafe.Pointer, length int) unsafe.Pointer {
	panic("reflect.unsafe_NewArray: reflect2/jsoniter is unsupported under TinyGo (use encoding/json)")
}

//go:linkname reflect2_typedmemmove reflect.typedmemmove
func reflect2_typedmemmove(rtype, dst, src unsafe.Pointer) {
	panic("reflect.typedmemmove: reflect2/jsoniter is unsupported under TinyGo (use encoding/json)")
}

//go:linkname reflect2_typedslicecopy reflect.typedslicecopy
func reflect2_typedslicecopy(elemType unsafe.Pointer, dst, src reflect2SliceHeader) int {
	panic("reflect.typedslicecopy: reflect2/jsoniter is unsupported under TinyGo (use encoding/json)")
}

//go:linkname reflect2_mapassign reflect.mapassign
func reflect2_mapassign(rtype, m, key, val unsafe.Pointer) {
	panic("reflect.mapassign: reflect2/jsoniter is unsupported under TinyGo (use encoding/json)")
}

//go:linkname reflect2_mapiterinit reflect.mapiterinit
func reflect2_mapiterinit(rtype, m, it unsafe.Pointer) {
	panic("reflect.mapiterinit: reflect2/jsoniter is unsupported under TinyGo (use encoding/json)")
}

//go:linkname reflect2_mapiternext reflect.mapiternext
func reflect2_mapiternext(it unsafe.Pointer) {
	panic("reflect.mapiternext: reflect2/jsoniter is unsupported under TinyGo (use encoding/json)")
}
