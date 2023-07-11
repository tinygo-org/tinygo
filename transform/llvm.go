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

// hasUses returns whether the given value has any uses. It is equivalent to
// getUses(value) != nil but faster.
func hasUses(value llvm.Value) bool {
	if value.IsNil() {
		return false
	}
	return !value.FirstUse().IsNil()
}

// makeGlobalArray creates a new LLVM global with the given name and integers as
// contents, and returns the global and initializer type.
// Note that it is left with the default linkage etc., you should set
// linkage/constant/etc properties yourself.
func makeGlobalArray(mod llvm.Module, bufItf interface{}, name string, elementType llvm.Type) (llvm.Type, llvm.Value) {
	buf := reflect.ValueOf(bufItf)
	var values []llvm.Value
	for i := 0; i < buf.Len(); i++ {
		ch := buf.Index(i).Uint()
		values = append(values, llvm.ConstInt(elementType, ch, false))
	}
	value := llvm.ConstArray(elementType, values)
	global := llvm.AddGlobal(mod, value.Type(), name)
	global.SetInitializer(value)
	return value.Type(), global
}

// getGlobalBytes returns the slice contained in the array of the provided
// global. It can recover the bytes originally created using makeGlobalArray, if
// makeGlobalArray was given a byte slice.
//
// The builder parameter is only used for constant operations.
func getGlobalBytes(global llvm.Value, builder llvm.Builder) []byte {
	value := global.Initializer()
	buf := make([]byte, value.Type().ArrayLength())
	for i := range buf {
		buf[i] = byte(builder.CreateExtractValue(value, i, "").ZExtValue())
	}
	return buf
}

// replaceGlobalByteWithArray replaces a global integer type in the module with
// an integer array, using a GEP to make the types match. It is a convenience
// function used for creating reflection sidetables, for example.
func replaceGlobalIntWithArray(mod llvm.Module, name string, buf interface{}) llvm.Value {
	oldGlobal := mod.NamedGlobal(name)
	globalType, global := makeGlobalArray(mod, buf, name+".tmp", oldGlobal.GlobalValueType())
	gep := llvm.ConstGEP(globalType, global, []llvm.Value{
		llvm.ConstInt(mod.Context().Int32Type(), 0, false),
		llvm.ConstInt(mod.Context().Int32Type(), 0, false),
	})
	oldGlobal.ReplaceAllUsesWith(gep)
	oldGlobal.EraseFromParentAsGlobal()
	global.SetName(name)
	return global
}

// stripPointerCasts strips instruction pointer casts (getelementptr and
// bitcast) and returns the original value without the casts.
func stripPointerCasts(value llvm.Value) llvm.Value {
	if !value.IsAConstantExpr().IsNil() {
		switch value.Opcode() {
		case llvm.GetElementPtr, llvm.BitCast:
			return stripPointerCasts(value.Operand(0))
		}
	}
	if !value.IsAInstruction().IsNil() {
		switch value.InstructionOpcode() {
		case llvm.GetElementPtr, llvm.BitCast:
			return stripPointerCasts(value.Operand(0))
		}
	}
	return value
}
