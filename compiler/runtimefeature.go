package compiler

import (
	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

func (b *builder) markRuntimeFeatureUse(call *ssa.CallCommon) {
	callee := call.StaticCallee()
	if callee == nil {
		return
	}

	var markerName string
	switch b.getFunctionInfo(callee).linkName {
	case "runtime.GOROOT":
		markerName = "tinygo.runtime.feature.goroot"
	case "(*internal/godebug.Setting).Value":
		markerName = "tinygo.runtime.feature.godebug"
	case "internal/godebug.setUpdate":
		if b.fn.Pkg != nil && b.fn.Pkg.Pkg.Path() == "internal/godebug" {
			return
		}
		markerName = "tinygo.runtime.feature.godebug"
	default:
		return
	}

	marker := b.mod.NamedFunction(markerName)
	if marker.IsNil() {
		marker = llvm.AddFunction(b.mod, markerName, llvm.FunctionType(b.ctx.VoidType(), nil, false))
	}
	b.CreateCall(marker.GlobalValueType(), marker, nil, "")
}
