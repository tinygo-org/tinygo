package transform

import (
	"tinygo.org/x/go-llvm"
)

// OptimizeMaps eliminates created but unused maps.
//
// In the future, this should statically allocate created but never modified
// maps. This has not yet been implemented, however.
func OptimizeMaps(mod llvm.Module) {
	hashmapMake := mod.NamedFunction("runtime.hashmapMake")
	hashmapMakeGeneric := mod.NamedFunction("runtime.hashmapMakeGeneric")

	hashmapBinarySet := mod.NamedFunction("runtime.hashmapBinarySet")
	hashmapStringSet := mod.NamedFunction("runtime.hashmapStringSet")
	hashmapGenericSet := mod.NamedFunction("runtime.hashmapGenericSet")

	optimizeMapMake := func(makeFunc llvm.Value) {
		if makeFunc.IsNil() {
			return
		}
		for _, makeInst := range getUses(makeFunc) {
			updateInsts := []llvm.Value{}
			unknownUses := false

			for _, use := range getUses(makeInst) {
				if use := use.IsACallInst(); !use.IsNil() {
					switch use.CalledValue() {
					case hashmapBinarySet, hashmapStringSet, hashmapGenericSet:
						updateInsts = append(updateInsts, use)
					default:
						unknownUses = true
					}
				} else {
					unknownUses = true
				}
			}

			if !unknownUses {
				// This map can be entirely removed, as it is only created
				// but never used.
				for _, inst := range updateInsts {
					inst.EraseFromParentAsInstruction()
				}
				makeInst.EraseFromParentAsInstruction()
			}
		}
	}

	optimizeMapMake(hashmapMake)
	optimizeMapMake(hashmapMakeGeneric)
}
