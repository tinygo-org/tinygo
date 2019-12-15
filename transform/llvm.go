package transform

import (
	"reflect"

	"tinygo.org/x/go-llvm"
)

// Return a list of values (actually, instructions) where this value is used as
// an operand.
func getUses(value llvm.Value) []llvm.Value {
	if value.IsNil() {
		return nil
	}
	var uses []llvm.Value
	use := value.FirstUse()
	for !use.IsNil() {
		uses = append(uses, use.User())
		use = use.NextUse()
	}
	return uses
}

// makeGlobalArray creates a new LLVM global with the given name and integers as
// contents, and returns the global.
// Note that it is left with the default linkage etc., you should set
// linkage/constant/etc properties yourself.
func makeGlobalArray(mod llvm.Module, bufItf interface{}, name string, elementType llvm.Type) llvm.Value {
	buf := reflect.ValueOf(bufItf)
	globalType := llvm.ArrayType(elementType, buf.Len())
	global := llvm.AddGlobal(mod, globalType, name)
	value := llvm.Undef(globalType)
	for i := 0; i < buf.Len(); i++ {
		ch := buf.Index(i).Uint()
		value = llvm.ConstInsertValue(value, llvm.ConstInt(elementType, ch, false), []uint32{uint32(i)})
	}
	global.SetInitializer(value)
	return global
}

// getGlobalBytes returns the slice contained in the array of the provided
// global. It can recover the bytes originally created using makeGlobalArray, if
// makeGlobalArray was given a byte slice.
func getGlobalBytes(global llvm.Value) []byte {
	value := global.Initializer()
	buf := make([]byte, value.Type().ArrayLength())
	for i := range buf {
		buf[i] = byte(llvm.ConstExtractValue(value, []uint32{uint32(i)}).ZExtValue())
	}
	return buf
}

// replaceGlobalByteWithArray replaces a global integer type in the module with
// an integer array, using a GEP to make the types match. It is a convenience
// function used for creating reflection sidetables, for example.
func replaceGlobalIntWithArray(mod llvm.Module, name string, buf interface{}) llvm.Value {
	oldGlobal := mod.NamedGlobal(name)
	global := makeGlobalArray(mod, buf, name+".tmp", oldGlobal.Type().ElementType())
	gep := llvm.ConstGEP(global, []llvm.Value{
		llvm.ConstInt(mod.Context().Int32Type(), 0, false),
		llvm.ConstInt(mod.Context().Int32Type(), 0, false),
	})
	oldGlobal.ReplaceAllUsesWith(gep)
	oldGlobal.EraseFromParentAsGlobal()
	global.SetName(name)
	return global
}

// typeHasPointers returns whether this type is a pointer or contains pointers.
// If the type is an aggregate type, it will check whether there is a pointer
// inside.
func typeHasPointers(t llvm.Type) bool {
	switch t.TypeKind() {
	case llvm.PointerTypeKind:
		return true
	case llvm.StructTypeKind:
		for _, subType := range t.StructElementTypes() {
			if typeHasPointers(subType) {
				return true
			}
		}
		return false
	case llvm.ArrayTypeKind:
		if typeHasPointers(t.ElementType()) {
			return true
		}
		return false
	default:
		return false
	}
}

// isFunctionLocal returns true if (and only if) this value is local to a
// function. That is, it returns true for instructions and parameters, and false
// for constants and globals.
func isFunctionLocal(val llvm.Value) bool {
	if !val.IsAConstant().IsNil() {
		return false
	}
	if !val.IsAInstruction().IsNil() {
		return true
	}
	if !val.IsAGlobalValue().IsNil() {
		return false
	}
	panic("unknown value kind")
}
