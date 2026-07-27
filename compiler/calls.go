package compiler

import (
	"go/token"
	"go/types"
	"strconv"

	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

// For a description of the calling convention in prose, see:
// https://tinygo.org/compiler-internals/calling-convention/

// The maximum number of arguments that can be expanded from a single struct. If
// a struct contains more fields, it is passed as a struct without expanding.
const maxFieldsPerParam = 3

// paramInfo contains some information collected about a function parameter,
// useful while declaring or defining a function.
type paramInfo struct {
	llvmType llvm.Type
	name     string     // name, possibly with suffixes for e.g. struct fields
	elemSize uint64     // size of pointer element type, or 0 if this isn't a pointer
	flags    paramFlags // extra flags for this parameter
}

// paramFlags identifies parameter attributes for flags. Most importantly, it
// determines which parameters are dereferenceable_or_null and which aren't.
type paramFlags uint8

const (
	// Whether this is a full or partial Go parameter (int, slice, etc).
	// The extra context parameter is not a Go parameter.
	paramIsGoParam = 1 << iota

	// Whether this is a readonly parameter (for example, a string pointer).
	paramIsReadonly

	// Whether this parameter is passed through backing storage.
	paramIsIndirect
)

// createRuntimeCallCommon creates a runtime call. Use createRuntimeCall or
// createRuntimeInvoke instead.
func (b *builder) createRuntimeCallCommon(fnName string, args []llvm.Value, name string, isInvoke bool) llvm.Value {
	fnType, llvmFn := b.getRuntimeFunction(fnName)
	args = append(args, llvm.Undef(b.dataPtrType)) // unused context parameter
	if isInvoke {
		// chanSend is the only panic-capable runtime operation that can also
		// suspend the task.
		return b.createInvokeWithAnalysis(fnType, llvmFn, args, name, true, fnName == "chanSend")
	}
	return b.createCall(fnType, llvmFn, args, name)
}

func (b *builder) getRuntimeFunction(name string) (llvm.Type, llvm.Value) {
	member := b.program.ImportedPackage("runtime").Members[name]
	if member == nil {
		panic("unknown runtime call: " + name)
	}
	fn := member.(*ssa.Function)
	fnType, llvmFn := b.getFunction(fn)
	if llvmFn.IsNil() {
		panic("trying to call non-existent function: " + fn.RelString(nil))
	}
	return fnType, llvmFn
}

// createRuntimeCall creates a new call to runtime.<fnName> with the given
// arguments.
func (b *builder) createRuntimeCall(fnName string, args []llvm.Value, name string) llvm.Value {
	return b.createRuntimeCallCommon(fnName, args, name, false)
}

// createRuntimeInvoke creates a new call to runtime.<fnName> with the given
// arguments. If the runtime call panics, control flow is diverted to the
// landing pad block.
// Note that "invoke" here is meant in the LLVM sense (a call that can
// panic/throw), not in the Go sense (an interface method call).
func (b *builder) createRuntimeInvoke(fnName string, args []llvm.Value, name string) llvm.Value {
	return b.createRuntimeCallCommon(fnName, args, name, true)
}

// createCall creates a call to the given function with the arguments possibly
// expanded.
func (b *builder) createCall(fnType llvm.Type, fn llvm.Value, args []llvm.Value, name string) llvm.Value {
	expanded := b.expandFormalParams(args)
	call := b.CreateCall(fnType, fn, expanded, name)
	if !fn.IsAFunction().IsNil() {
		if cc := fn.FunctionCallConv(); cc != llvm.CCallConv {
			// Set a different calling convention if needed.
			// This is needed for GetModuleHandleExA on Windows, for example.
			call.SetInstructionCallConv(cc)
		}
	}
	return call
}

func (b *builder) expandFormalParams(args []llvm.Value) []llvm.Value {
	expanded := make([]llvm.Value, 0, len(args))
	for _, arg := range args {
		fragments := b.expandFormalParam(arg)
		expanded = append(expanded, fragments...)
	}
	return expanded
}

// createInvoke emits a Go call, adding unwind handling only when the static
// call graph indicates the call can unwind.
func (b *builder) createInvoke(fnType llvm.Type, fn llvm.Value, args []llvm.Value, name string, call *ssa.CallCommon) llvm.Value {
	properties := b.callPropertiesFor(call)
	return b.createInvokeWithAnalysis(fnType, fn, args, name, properties&callMayUnwind != 0, properties&callMaySuspend != 0)
}

func (b *builder) createInvokeWithAnalysis(fnType llvm.Type, fn llvm.Value, args []llvm.Value, name string, mayUnwind, maySuspend bool) llvm.Value {
	if !mayUnwind {
		return b.createCall(fnType, fn, args, name)
	}
	if b.usesReturnUnwind() {
		var result llvm.Value
		if b.usesAsyncifyUnwind() && b.hasDeferFrame() {
			if maySuspend {
				result = b.createAsyncifySuspendInvoke(fnType, fn, args, name)
			} else {
				result = b.createAsyncifyNoSuspendInvoke(fnType, fn, args, name)
			}
		} else {
			result = b.createCall(fnType, fn, args, name)
		}
		if b.needsPostCallUnwindCheck() {
			b.createUnwindCheck(true)
		}

		return result
	}
	if b.hasDeferFrame() {
		b.createInvokeCheckpoint()
	}
	return b.createCall(fnType, fn, args, name)
}

func (b *builder) createAsyncifyNoSuspendInvoke(fnType llvm.Type, fn llvm.Value, args []llvm.Value, name string) llvm.Value {
	expanded := b.expandFormalParams(args)
	paramTypes := append([]llvm.Type(nil), fnType.ParamTypes()...)
	direct := !fn.IsAFunction().IsNil()
	if !direct {
		paramTypes = append([]llvm.Type{fn.Type()}, paramTypes...)
	}
	wrapperType := llvm.FunctionType(fnType.ReturnType(), paramTypes, false)
	wrapperName := b.llvmFn.Name() + ".asyncifycatch." + strconv.Itoa(b.asyncifyCatchIndex)
	b.asyncifyCatchIndex++
	wrapper := llvm.AddFunction(b.mod, wrapperName, wrapperType)
	wrapper.SetLinkage(llvm.InternalLinkage)
	wrapper.AddFunctionAttr(b.ctx.CreateEnumAttribute(llvm.AttributeKindID("noinline"), 0))

	builder := b.ctx.NewBuilder()
	entry := b.ctx.AddBasicBlock(wrapper, "entry")
	builder.SetInsertPointAtEnd(entry)
	target := fn
	paramOffset := 0
	if !direct {
		target = wrapper.Param(0)
		paramOffset = 1
	}
	callArgs := make([]llvm.Value, len(fnType.ParamTypes()))
	for i := range callArgs {
		callArgs[i] = wrapper.Param(i + paramOffset)
	}
	result := builder.CreateCall(fnType, target, callArgs, "")
	unwindType, unwindLLVMFn := b.getRuntimeFunction("unwindPending")
	unwinding := builder.CreateCall(unwindType, unwindLLVMFn, []llvm.Value{llvm.Undef(b.dataPtrType)}, "")
	b.emitAsyncifyCatchReturn(builder, fnType, result, entry, unwinding)
	builder.Dispose()

	wrapperArgs := expanded
	if !direct {
		wrapperArgs = append([]llvm.Value{fn}, expanded...)
	}
	return b.CreateCall(wrapperType, wrapper, wrapperArgs, name)
}

func (b *builder) createAsyncifySuspendInvoke(fnType llvm.Type, fn llvm.Value, args []llvm.Value, name string) llvm.Value {
	// Binaryen does not instrument a function containing stop_unwind, so a
	// suspend-capable call needs two wrappers:
	//
	//   defer caller (instrumented)
	//       |
	//       | volatile indirect call
	//       v
	//   panic catcher (not instrumented; contains stop_unwind)
	//       |
	//       v
	//   target trampoline (instrumented)
	//       |
	//       | indirect call
	//       v
	//   real target
	//
	// A scheduler unwind crosses both wrappers. A panic unwind stops in the
	// catcher, then the caller branches to its defer landing pad.
	expanded := b.expandFormalParams(args)
	paramTypes := append([]llvm.Type{fn.Type()}, fnType.ParamTypes()...)
	wrapperType := llvm.FunctionType(fnType.ReturnType(), paramTypes, false)
	wrapperArgs := append([]llvm.Value{fn}, expanded...)
	if catchGlobal, ok := b.asyncifyCatchers[wrapperType]; ok {
		catchPointer := b.CreateLoad(catchGlobal.GlobalValueType(), catchGlobal, "")
		catchPointer.SetVolatile(true)
		return b.CreateCall(wrapperType, catchPointer, wrapperArgs, name)
	}
	wrapperName := b.llvmFn.Name() + ".asyncifysuspendcatch." + strconv.Itoa(b.asyncifyCatchIndex)
	b.asyncifyCatchIndex++

	targetType := wrapperType
	var resultGlobal llvm.Value
	if fnType.ReturnType().TypeKind() != llvm.VoidTypeKind {
		targetType = llvm.FunctionType(b.ctx.VoidType(), paramTypes, false)
		resultGlobal = llvm.AddGlobal(b.mod, b.dataPtrType, wrapperName+".result.ptr")
		resultGlobal.SetLinkage(llvm.InternalLinkage)
		resultGlobal.SetInitializer(llvm.ConstNull(b.dataPtrType))
	}
	targetWrapper := llvm.AddFunction(b.mod, wrapperName+".target", targetType)
	targetWrapper.SetLinkage(llvm.InternalLinkage)
	targetWrapper.AddFunctionAttr(b.ctx.CreateEnumAttribute(llvm.AttributeKindID("noinline"), 0))
	builder := b.ctx.NewBuilder()
	entry := b.ctx.AddBasicBlock(targetWrapper, "entry")
	builder.SetInsertPointAtEnd(entry)
	target := targetWrapper.Param(0)
	callArgs := make([]llvm.Value, len(fnType.ParamTypes()))
	for i := range callArgs {
		callArgs[i] = targetWrapper.Param(i + 1)
	}
	result := builder.CreateCall(fnType, target, callArgs, "")
	if fnType.ReturnType().TypeKind() == llvm.VoidTypeKind {
		builder.CreateRetVoid()
	} else {
		resultPointer := builder.CreateLoad(b.dataPtrType, resultGlobal, "")
		resultPointer.SetVolatile(true)
		builder.CreateStore(result, resultPointer)
		builder.CreateRetVoid()
	}
	builder.Dispose()

	catchWrapper := llvm.AddFunction(b.mod, wrapperName+".paniccatch", wrapperType)
	catchWrapper.SetLinkage(llvm.InternalLinkage)
	catchWrapper.AddFunctionAttr(b.ctx.CreateEnumAttribute(llvm.AttributeKindID("noinline"), 0))
	builder = b.ctx.NewBuilder()
	entry = b.ctx.AddBasicBlock(catchWrapper, "entry")
	builder.SetInsertPointAtEnd(entry)
	catchArgs := make([]llvm.Value, len(paramTypes))
	for i := range catchArgs {
		catchArgs[i] = catchWrapper.Param(i)
	}
	if fnType.ReturnType().TypeKind() == llvm.VoidTypeKind {
		result = builder.CreateCall(targetType, targetWrapper, catchArgs, "")
	} else {
		// Asyncify can restore a stale hidden result pointer during rewind. Use
		// a volatile slot to select the current catcher's storage, restoring its
		// previous value before the unwind can reach the scheduler. This also
		// makes recursive use safe.
		resultStorage := builder.CreateAlloca(fnType.ReturnType(), "")
		previousResultPointer := builder.CreateLoad(b.dataPtrType, resultGlobal, "")
		previousResultPointer.SetVolatile(true)
		storeResultPointer := builder.CreateStore(resultStorage, resultGlobal)
		storeResultPointer.SetVolatile(true)
		builder.CreateCall(targetType, targetWrapper, catchArgs, "")
		restoreResultPointer := builder.CreateStore(previousResultPointer, resultGlobal)
		restoreResultPointer.SetVolatile(true)
		result = builder.CreateLoad(fnType.ReturnType(), resultStorage, "")
	}
	unwindType, unwindLLVMFn := b.getRuntimeFunction("unwindPending")
	// This call must stay opaque to AddUnwindAssumptions. This catcher observes
	// the signal set by its target, so assuming a clear signal on entry would
	// let LLVM remove the check.
	unwindGlobal := llvm.AddGlobal(b.mod, unwindLLVMFn.Type(), wrapperName+".unwind.ptr")
	unwindGlobal.SetLinkage(llvm.InternalLinkage)
	unwindGlobal.SetInitializer(unwindLLVMFn)
	unwindPointer := builder.CreateLoad(unwindLLVMFn.Type(), unwindGlobal, "")
	unwindPointer.SetVolatile(true)
	unwinding := builder.CreateCall(unwindType, unwindPointer, []llvm.Value{llvm.Undef(b.dataPtrType)}, "")
	b.emitAsyncifyCatchReturn(builder, fnType, result, entry, unwinding)
	builder.Dispose()

	// Keep the caller-to-catcher edge opaque so Binaryen instruments the
	// indirect call instead of treating it as a call into bottommost runtime.
	catchGlobal := llvm.AddGlobal(b.mod, catchWrapper.Type(), wrapperName+".paniccatch.ptr")
	catchGlobal.SetLinkage(llvm.InternalLinkage)
	catchGlobal.SetInitializer(catchWrapper)
	b.asyncifyCatchers[wrapperType] = catchGlobal
	catchPointer := b.CreateLoad(catchWrapper.Type(), catchGlobal, "")
	catchPointer.SetVolatile(true)
	return b.CreateCall(wrapperType, catchPointer, wrapperArgs, name)
}

func (b *builder) emitAsyncifyCatchReturn(builder llvm.Builder, fnType llvm.Type, result llvm.Value, entry llvm.BasicBlock, unwinding llvm.Value) {
	wrapper := entry.Parent()
	stopBlock := b.ctx.AddBasicBlock(wrapper, "unwind.stop")
	returnBlock := b.ctx.AddBasicBlock(wrapper, "return")
	builder.CreateCondBr(unwinding, stopBlock, returnBlock)

	builder.SetInsertPointAtEnd(stopBlock)
	stopType, stopFn := b.getRuntimeFunction("asyncifyStopUnwindImport")
	builder.CreateCall(stopType, stopFn, nil, "")
	builder.CreateBr(returnBlock)

	builder.SetInsertPointAtEnd(returnBlock)
	if fnType.ReturnType().TypeKind() == llvm.VoidTypeKind {
		builder.CreateRetVoid()
		return
	}
	phi := builder.CreatePHI(fnType.ReturnType(), "")
	phi.AddIncoming([]llvm.Value{result, result}, []llvm.BasicBlock{entry, stopBlock})
	builder.CreateRet(phi)
}

type callProperties uint8

const (
	callMaySuspend callProperties = 1 << iota
	callMayUnwind
)

type functionCallProperties struct {
	properties callProperties
	visiting   bool
	complete   bool
}

func (b *builder) functionCallProperties(fn *ssa.Function) callProperties {
	analysis := b.callProperties[fn]
	if analysis.complete {
		return analysis.properties
	}
	if analysis.visiting || len(fn.Blocks) == 0 {
		return callMaySuspend | callMayUnwind
	}

	analysis.visiting = true
	b.callProperties[fn] = analysis

	var properties callProperties
	if fn.Pkg != nil && fn.Pkg.Pkg.Path() == "internal/task" && fn.Name() == "Pause" {
		properties |= callMaySuspend
	}
	for _, block := range fn.Blocks {
		if block == nil {
			continue
		}
		for _, instruction := range block.Instrs {
			properties |= instructionCallProperties(instruction)
			call, ok := instruction.(ssa.CallInstruction)
			if !ok {
				continue
			}
			common := call.Common()
			if builtin, ok := common.Value.(*ssa.Builtin); ok {
				switch builtin.Name() {
				case "close", "delete", "panic":
					properties |= callMayUnwind
				}
				continue
			}
			callee := common.StaticCallee()
			if callee == nil {
				properties |= callMaySuspend | callMayUnwind
			} else {
				properties |= b.functionCallProperties(callee)
			}
			if properties == callMaySuspend|callMayUnwind {
				break
			}
		}
	}

	b.callProperties[fn] = functionCallProperties{
		properties: properties,
		complete:   true,
	}
	return properties
}

func (b *builder) callPropertiesFor(call *ssa.CallCommon) callProperties {
	callee := call.StaticCallee()
	if callee == nil {
		return callMaySuspend | callMayUnwind
	}
	return b.functionCallProperties(callee)
}

func (b *builder) functionMaySuspend(fn *ssa.Function) bool {
	return b.functionCallProperties(fn)&callMaySuspend != 0
}

func (b *builder) functionMayUnwind(fn *ssa.Function) bool {
	return b.functionCallProperties(fn)&callMayUnwind != 0
}

func instructionCallProperties(instruction ssa.Instruction) callProperties {
	switch instruction := instruction.(type) {
	case *ssa.Alloc, *ssa.Call, *ssa.ChangeInterface, *ssa.ChangeType,
		*ssa.Convert, *ssa.DebugRef, *ssa.Defer, *ssa.Extract, *ssa.Field,
		*ssa.Go, *ssa.If, *ssa.Jump, *ssa.MakeClosure, *ssa.MakeInterface,
		*ssa.MakeMap, *ssa.Phi, *ssa.Range, *ssa.Return, *ssa.RunDefers:
		return 0
	case *ssa.Send:
		return callMaySuspend | callMayUnwind
	case *ssa.Next:
		return callMaySuspend
	case *ssa.Select:
		return callMaySuspend | callMayUnwind
	case *ssa.FieldAddr, *ssa.Index, *ssa.IndexAddr, *ssa.Lookup,
		*ssa.MakeChan, *ssa.MakeSlice, *ssa.MapUpdate, *ssa.Panic,
		*ssa.Slice, *ssa.SliceToArrayPointer, *ssa.Store, *ssa.TypeAssert:
		return callMayUnwind
	case *ssa.BinOp:
		if binOpMayUnwind(instruction) {
			return callMayUnwind
		}
	case *ssa.UnOp:
		var properties callProperties
		if instruction.Op == token.ARROW {
			properties |= callMaySuspend
		}
		if instruction.Op == token.MUL {
			properties |= callMayUnwind
		}
		return properties
	}
	// New SSA instructions must be treated conservatively until their lowering
	// is audited for suspension and unwind paths.
	return callMaySuspend | callMayUnwind
}

func binOpMayUnwind(instruction *ssa.BinOp) bool {
	switch instruction.Op {
	case token.QUO, token.REM:
		basic, ok := instruction.X.Type().Underlying().(*types.Basic)
		return ok && basic.Info()&types.IsInteger != 0
	case token.SHL, token.SHR:
		basic, ok := instruction.Y.Type().Underlying().(*types.Basic)
		return ok && basic.Info()&types.IsUnsigned == 0
	case token.EQL, token.NEQ:
		return typeMayPanicOnCompare(instruction.X.Type())
	default:
		return false
	}
}

func typeMayPanicOnCompare(typ types.Type) bool {
	switch typ := typ.Underlying().(type) {
	case *types.Interface:
		return true
	case *types.Array:
		return typeMayPanicOnCompare(typ.Elem())
	case *types.Struct:
		for i := 0; i < typ.NumFields(); i++ {
			if typeMayPanicOnCompare(typ.Field(i).Type()) {
				return true
			}
		}
	}
	return false
}

func (b *builder) inFunctionBody() bool {
	block := b.GetInsertBlock()
	return b.loweringBody && !b.llvmFn.IsNil() && !block.IsNil() && block.Parent() == b.llvmFn
}

func (b *builder) needsPostCallUnwindCheck() bool {
	if !b.inFunctionBody() || b.runningDefers || b.isUnwindRuntime() {
		return false
	}
	// Explicit unwinding propagates through every caller. Asyncify already
	// unwinds ordinary callers itself; only a defer frame stops it and needs a
	// check that branches to the landing pad.
	return !b.usesAsyncifyUnwind() || b.hasDeferFrame()
}

func (b *builder) isUnwindRuntime() bool {
	if !b.usesReturnUnwind() {
		return false
	}
	if b.fn.Pkg == nil {
		return false
	}
	if b.fn.Pkg.Pkg.Path() != "runtime" {
		return false
	}
	switch b.fn.Name() {
	case "startUnwind", "currentDeferFrame", "unwindPending", "clearUnwind",
		"getUnwindSignal", "setUnwindSignal":
		return true
	default:
		return false
	}
}

// When catch is true, route the unwind to this function's defer landing pad.
func (b *builder) createUnwindCheck(catch bool) {
	unwind := b.createRuntimeCall("unwindPending", nil, "unwind")
	continueBB := b.insertBasicBlock("unwind.continue")
	if catch && b.hasDeferFrame() {
		b.CreateCondBr(unwind, b.landingpad, continueBB)
	} else {
		b.CreateCondBr(unwind, b.unwindReturnBlock(), continueBB)
	}

	b.SetInsertPointAtEnd(continueBB)
	if !b.inFaultBlock {
		b.currentBlockInfo.exit = continueBB
	}
}

func (b *builder) unwindReturnBlock() llvm.BasicBlock {
	if !b.unwindReturn.IsNil() {
		return b.unwindReturn
	}

	savedBlock := b.GetInsertBlock()
	b.unwindReturn = b.ctx.AddBasicBlock(b.llvmFn, "unwind.return")
	b.SetInsertPointAtEnd(b.unwindReturn)
	returnType := b.llvmFn.GlobalValueType().ReturnType()
	if returnType.TypeKind() == llvm.VoidTypeKind {
		b.CreateRetVoid()
	} else {
		b.CreateRet(llvm.Undef(returnType))
	}
	b.SetInsertPointAtEnd(savedBlock)
	return b.unwindReturn
}

func (b *builder) createUnwindReturnOrUnreachable() {
	if b.usesAsyncifyUnwind() {
		b.CreateBr(b.unwindReturnBlock())
	} else {
		b.CreateUnreachable()
	}
}

// Expand an argument type to a list that can be used in a function call
// parameter list.
func (c *compilerContext) expandFormalParamType(t llvm.Type, name string, goType types.Type) []paramInfo {
	if c.isIndirectAggregate(t) {
		return []paramInfo{{
			llvmType: c.dataPtrType,
			name:     name,
			elemSize: c.targetData.TypeAllocSize(t),
			flags:    paramIsGoParam | paramIsReadonly | paramIsIndirect,
		}}
	}
	return c.expandDirectFormalParamType(t, name, goType)
}

func (c *compilerContext) expandDirectFormalParamType(t llvm.Type, name string, goType types.Type) []paramInfo {
	switch t.TypeKind() {
	case llvm.StructTypeKind:
		fieldInfos := c.flattenAggregateType(t, name, goType)
		if len(fieldInfos) <= maxFieldsPerParam {
			// managed to expand this parameter
			return fieldInfos
		}
		// failed to expand this parameter: too many fields
	}
	// TODO: split small arrays
	return []paramInfo{c.getParamInfo(t, name, goType)}
}

func (c *compilerContext) storedParamType(t llvm.Type, exported bool) llvm.Type {
	if c.isIndirectParam(t, exported) {
		return c.dataPtrType
	}
	return t
}

func (c *compilerContext) isIndirectParam(t llvm.Type, exported bool) bool {
	return !exported && c.isIndirectAggregate(t)
}

func (b *builder) appendStoredValueTypes(valueTypes []llvm.Type, values []ssa.Value, exported bool) []llvm.Type {
	for _, value := range values {
		valueTypes = append(valueTypes, b.storedParamType(b.getLLVMType(value.Type()), exported))
	}
	return valueTypes
}

func (b *builder) appendStoredParamTypes(valueTypes []llvm.Type, params []*types.Var, exported bool) []llvm.Type {
	for _, param := range params {
		valueTypes = append(valueTypes, b.storedParamType(b.getLLVMType(param.Type()), exported))
	}
	return valueTypes
}

func (b *builder) prependIndirectResult(sig *types.Signature, exported bool, params []llvm.Value, name string) []llvm.Value {
	if resultType, indirect := b.hasIndirectResult(sig); !exported && indirect {
		return append([]llvm.Value{b.createIndirectStorage(resultType, name)}, params...)
	}
	return params
}

// expandFormalParamOffsets returns a list of offsets from the start of an
// object of type t after it would have been split up by expandFormalParam. This
// is useful for debug information, where it is necessary to know the offset
// from the start of the combined object.
func (b *builder) expandFormalParamOffsets(t llvm.Type) []uint64 {
	switch t.TypeKind() {
	case llvm.StructTypeKind:
		fields := b.flattenAggregateTypeOffsets(t)
		if len(fields) <= maxFieldsPerParam {
			return fields
		} else {
			// failed to lower
			return []uint64{0}
		}
	default:
		// TODO: split small arrays
		return []uint64{0}
	}
}

// expandFormalParam splits a formal param value into pieces, so it can be
// passed directly as part of a function call. For example, it splits up small
// structs into individual fields. It is the equivalent of expandFormalParamType
// for parameter values.
func (b *builder) expandFormalParam(v llvm.Value) []llvm.Value {
	switch v.Type().TypeKind() {
	case llvm.StructTypeKind:
		fieldInfos := b.flattenAggregateType(v.Type(), "", nil)
		if len(fieldInfos) <= maxFieldsPerParam {
			fields := b.flattenAggregate(v)
			if len(fields) != len(fieldInfos) {
				panic("type and value param lowering don't match")
			}
			return fields
		} else {
			// failed to lower
			return []llvm.Value{v}
		}
	default:
		// TODO: split small arrays
		return []llvm.Value{v}
	}
}

// Try to flatten a struct type to a list of types. Returns a 1-element slice
// with the passed in type if this is not possible.
func (c *compilerContext) flattenAggregateType(t llvm.Type, name string, goType types.Type) []paramInfo {
	switch t.TypeKind() {
	case llvm.StructTypeKind:
		var paramInfos []paramInfo
		for i, subfield := range t.StructElementTypes() {
			if c.targetData.TypeAllocSize(subfield) == 0 {
				continue
			}
			suffix := strconv.Itoa(i)
			isString := false
			if goType != nil {
				// Try to come up with a good suffix for this struct field,
				// depending on which Go type it's based on.
				switch goType := goType.Underlying().(type) {
				case *types.Interface:
					suffix = []string{"typecode", "value"}[i]
				case *types.Slice:
					suffix = []string{"data", "len", "cap"}[i]
				case *types.Struct:
					suffix = goType.Field(i).Name()
				case *types.Basic:
					switch goType.Kind() {
					case types.Complex64, types.Complex128:
						suffix = []string{"r", "i"}[i]
					case types.String:
						suffix = []string{"data", "len"}[i]
						isString = true
					}
				case *types.Signature:
					suffix = []string{"context", "funcptr"}[i]
				}
			}
			subInfos := c.flattenAggregateType(subfield, name+"."+suffix, extractSubfield(goType, i))
			if isString {
				subInfos[0].flags |= paramIsReadonly
			}
			paramInfos = append(paramInfos, subInfos...)
		}
		return paramInfos
	default:
		return []paramInfo{c.getParamInfo(t, name, goType)}
	}
}

// getParamInfo collects information about a parameter. For example, if this
// parameter is pointer-like, it will also store the element type for the
// dereferenceable_or_null attribute.
func (c *compilerContext) getParamInfo(t llvm.Type, name string, goType types.Type) paramInfo {
	info := paramInfo{
		llvmType: t,
		name:     name,
		flags:    paramIsGoParam,
	}
	if goType != nil {
		switch underlying := goType.Underlying().(type) {
		case *types.Pointer:
			// Pointers in Go must either point to an object or be nil.
			info.elemSize = c.targetData.TypeAllocSize(c.getLLVMType(underlying.Elem()))
		case *types.Chan:
			// Channels are implemented simply as a *runtime.channel.
			info.elemSize = c.targetData.TypeAllocSize(c.getLLVMRuntimeType("channel"))
		case *types.Map:
			// Maps are similar to channels: they are implemented as a
			// *runtime.hashmap.
			info.elemSize = c.targetData.TypeAllocSize(c.getLLVMRuntimeType("hashmap"))
		}
	}
	return info
}

// extractSubfield extracts a field from a struct, or returns null if this is
// not a struct and thus no subfield can be obtained.
func extractSubfield(t types.Type, field int) types.Type {
	if t == nil {
		return nil
	}
	switch t := t.Underlying().(type) {
	case *types.Struct:
		return t.Field(field).Type()
	case *types.Interface, *types.Slice, *types.Basic, *types.Signature:
		// These Go types are (sometimes) implemented as LLVM structs but can't
		// really be split further up in Go (with the possible exception of
		// complex numbers).
		return nil
	default:
		// This should be unreachable.
		panic("cannot split subfield: " + t.String())
	}
}

// flattenAggregateTypeOffsets returns the offsets from the start of an object of
// type t if this object were flattened like in flattenAggregate. Used together
// with flattenAggregate to know the start indices of each value in the
// non-flattened object.
//
// Note: this is an implementation detail, use expandFormalParamOffsets instead.
func (c *compilerContext) flattenAggregateTypeOffsets(t llvm.Type) []uint64 {
	switch t.TypeKind() {
	case llvm.StructTypeKind:
		var fields []uint64
		for fieldIndex, field := range t.StructElementTypes() {
			if c.targetData.TypeAllocSize(field) == 0 {
				continue
			}
			suboffsets := c.flattenAggregateTypeOffsets(field)
			offset := c.targetData.ElementOffset(t, fieldIndex)
			for i := range suboffsets {
				suboffsets[i] += offset
			}
			fields = append(fields, suboffsets...)
		}
		return fields
	default:
		return []uint64{0}
	}
}

// flattenAggregate breaks down a struct into its elementary values for argument
// passing. It is the value equivalent of flattenAggregateType
func (b *builder) flattenAggregate(v llvm.Value) []llvm.Value {
	switch v.Type().TypeKind() {
	case llvm.StructTypeKind:
		var fields []llvm.Value
		for i, field := range v.Type().StructElementTypes() {
			if b.targetData.TypeAllocSize(field) == 0 {
				continue
			}
			subfield := b.CreateExtractValue(v, i, "")
			subfields := b.flattenAggregate(subfield)
			fields = append(fields, subfields...)
		}
		return fields
	default:
		return []llvm.Value{v}
	}
}

// collapseFormalParam combines an aggregate object back into the original
// value. This is used to join multiple LLVM parameters into a single Go value
// in the function entry block.
func (b *builder) collapseFormalParam(t llvm.Type, fields []llvm.Value) llvm.Value {
	param, remaining := b.collapseFormalParamInternal(t, fields)
	if len(remaining) != 0 {
		panic("failed to expand back all fields")
	}
	return param
}

// collapseFormalParamInternal is an implementation detail of
// collapseFormalParam: it works by recursing until there are no fields left.
func (b *builder) collapseFormalParamInternal(t llvm.Type, fields []llvm.Value) (llvm.Value, []llvm.Value) {
	switch t.TypeKind() {
	case llvm.StructTypeKind:
		flattened := b.flattenAggregateType(t, "", nil)
		if len(flattened) <= maxFieldsPerParam {
			value := llvm.ConstNull(t)
			for i, subtyp := range t.StructElementTypes() {
				if b.targetData.TypeAllocSize(subtyp) == 0 {
					continue
				}
				structField, remaining := b.collapseFormalParamInternal(subtyp, fields)
				fields = remaining
				value = b.CreateInsertValue(value, structField, i, "")
			}
			return value, fields
		} else {
			// this struct was not flattened
			return fields[0], fields[1:]
		}
	default:
		return fields[0], fields[1:]
	}
}
