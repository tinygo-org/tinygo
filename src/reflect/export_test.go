package reflect

import "internal/reflectlite"

type OtherPkgFields struct {
	OtherExported   int
	otherUnexported int
}

type Buffer struct {
	buf []byte
}

type EmbedWithUnexpMeth struct{}

func (EmbedWithUnexpMeth) f() {}

func MakeRO(v Value) Value {
	return Value{reflectlite.MakeRO(v.Value, true)}
}

func IsRO(v Value) bool {
	return reflectlite.IsRO(v.Value)
}

// Stubs for upstream gc-internal test helpers. Tests that rely on these are
// skipped in tools/stdlibtest/rules.go.
var (
	callGC bool
	CallGC = &callGC
)

func FuncLayout(t Type, rcvr Type) (frametype Type, argSize, retOffset uintptr, stack, gc, inReg, outReg []byte, ptrs bool) {
	panic("reflect.FuncLayout is gc-specific")
}

func TypeLinks() []string {
	panic("reflect.TypeLinks is gc-specific")
}

func GCBits(any) []byte {
	panic("reflect.GCBits is gc-specific")
}

func FirstMethodNameBytes(t Type) *byte {
	panic("reflect.FirstMethodNameBytes is gc-specific")
}

func IsExported(t Type) bool {
	panic("reflect.IsExported is gc-specific")
}

func ResolveReflectName(s string) {
	panic("reflect.ResolveReflectName is gc-specific")
}

func SetArgRegs(ints, floats int, floatSize uintptr) (oldInts, oldFloats int, oldFloatSize uintptr) {
	panic("reflect.SetArgRegs is gc-specific")
}

func MethodValueCallCodePtr() uintptr {
	panic("reflect.MethodValueCallCodePtr is gc-specific")
}

func InternalIsZero(b []byte) bool {
	panic("reflect.InternalIsZero is gc-specific")
}

func IsRegularMemory(t Type) bool {
	panic("reflect.IsRegularMemory is gc-specific")
}

func MapGroupOf(x, y Type) Type {
	panic("reflect.MapGroupOf is gc-specific")
}
