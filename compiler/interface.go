package compiler

// This file transforms interface-related instructions (*ssa.MakeInterface,
// *ssa.TypeAssert, calls on interface types) to an intermediate IR form, to be
// lowered to the final form by the interface lowering pass. See
// interface-lowering.go for more details.

import (
	"go/token"
	"go/types"
	"strconv"
	"strings"

	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

// createMakeInterface emits the LLVM IR for the *ssa.MakeInterface instruction.
// It tries to put the type in the interface value, but if that's not possible,
// it will do an allocation of the right size and put that in the interface
// value field.
//
// An interface value is a {typecode, value} tuple named runtime._interface.
func (b *builder) createMakeInterface(val llvm.Value, typ types.Type, pos token.Pos) llvm.Value {
	itfValue := b.emitPointerPack([]llvm.Value{val})
	itfTypeCodeGlobal := b.getTypeCode(typ)
	itfTypeCode := b.CreatePtrToInt(itfTypeCodeGlobal, b.uintptrType, "")
	itf := llvm.Undef(b.getLLVMRuntimeType("_interface"))
	itf = b.CreateInsertValue(itf, itfTypeCode, 0, "")
	itf = b.CreateInsertValue(itf, itfValue, 1, "")
	return itf
}

// extractValueFromInterface extract the value from an interface value
// (runtime._interface) under the assumption that it is of the type given in
// llvmType. The behavior is undefied if the interface is nil or llvmType
// doesn't match the underlying type of the interface.
func (b *builder) extractValueFromInterface(itf llvm.Value, llvmType llvm.Type) llvm.Value {
	valuePtr := b.CreateExtractValue(itf, 1, "typeassert.value.ptr")
	return b.emitPointerUnpack(valuePtr, []llvm.Type{llvmType})[0]
}

// getTypeCode returns a reference to a type code.
// It returns a pointer to an external global which should be replaced with the
// real type in the interface lowering pass.
func (c *compilerContext) getTypeCode(typ types.Type) llvm.Value {
	globalName := "reflect/types.type:" + getTypeCodeName(typ)
	global := c.mod.NamedGlobal(globalName)
	if global.IsNil() {
		// Create a new typecode global.
		global = llvm.AddGlobal(c.mod, c.getLLVMRuntimeType("typecodeID"), globalName)
		// Some type classes contain more information for underlying types or
		// element types. Store it directly in the typecode global to make
		// reflect lowering simpler.
		var references llvm.Value
		var length int64
		var methodSet llvm.Value
		var ptrTo llvm.Value
		var typeAssert llvm.Value
		switch typ := typ.(type) {
		case *types.Named:
			references = c.getTypeCode(typ.Underlying())
		case *types.Chan:
			references = c.getTypeCode(typ.Elem())
		case *types.Pointer:
			references = c.getTypeCode(typ.Elem())
		case *types.Slice:
			references = c.getTypeCode(typ.Elem())
		case *types.Array:
			references = c.getTypeCode(typ.Elem())
			length = typ.Len()
		case *types.Struct:
			// Take a pointer to the typecodeID of the first field (if it exists).
			structGlobal := c.makeStructTypeFields(typ)
			references = llvm.ConstBitCast(structGlobal, global.Type())
		case *types.Interface:
			methodSetGlobal := c.getInterfaceMethodSet(typ)
			references = llvm.ConstBitCast(methodSetGlobal, global.Type())
		}
		if _, ok := typ.Underlying().(*types.Interface); !ok {
			methodSet = c.getTypeMethodSet(typ)
		} else {
			typeAssert = c.getInterfaceImplementsFunc(typ)
			typeAssert = llvm.ConstPtrToInt(typeAssert, c.uintptrType)
		}
		if _, ok := typ.Underlying().(*types.Pointer); !ok {
			ptrTo = c.getTypeCode(types.NewPointer(typ))
		}
		globalValue := llvm.ConstNull(global.Type().ElementType())
		if !references.IsNil() {
			globalValue = llvm.ConstInsertValue(globalValue, references, []uint32{0})
		}
		if length != 0 {
			lengthValue := llvm.ConstInt(c.uintptrType, uint64(length), false)
			globalValue = llvm.ConstInsertValue(globalValue, lengthValue, []uint32{1})
		}
		if !methodSet.IsNil() {
			globalValue = llvm.ConstInsertValue(globalValue, methodSet, []uint32{2})
		}
		if !ptrTo.IsNil() {
			globalValue = llvm.ConstInsertValue(globalValue, ptrTo, []uint32{3})
		}
		if !typeAssert.IsNil() {
			globalValue = llvm.ConstInsertValue(globalValue, typeAssert, []uint32{4})
		}
		global.SetInitializer(globalValue)
		global.SetLinkage(llvm.LinkOnceODRLinkage)
		global.SetGlobalConstant(true)
	}
	return global
}

// makeStructTypeFields creates a new global that stores all type information
// related to this struct type, and returns the resulting global. This global is
// actually an array of all the fields in the structs.
func (c *compilerContext) makeStructTypeFields(typ *types.Struct) llvm.Value {
	// The global is an array of runtime.structField structs.
	runtimeStructField := c.getLLVMRuntimeType("structField")
	structGlobalType := llvm.ArrayType(runtimeStructField, typ.NumFields())
	structGlobal := llvm.AddGlobal(c.mod, structGlobalType, "reflect/types.structFields")
	structGlobalValue := llvm.ConstNull(structGlobalType)
	for i := 0; i < typ.NumFields(); i++ {
		fieldGlobalValue := llvm.ConstNull(runtimeStructField)
		fieldGlobalValue = llvm.ConstInsertValue(fieldGlobalValue, c.getTypeCode(typ.Field(i).Type()), []uint32{0})
		fieldName := c.makeGlobalArray([]byte(typ.Field(i).Name()), "reflect/types.structFieldName", c.ctx.Int8Type())
		fieldName.SetLinkage(llvm.PrivateLinkage)
		fieldName.SetUnnamedAddr(true)
		fieldName = llvm.ConstGEP(fieldName, []llvm.Value{
			llvm.ConstInt(c.ctx.Int32Type(), 0, false),
			llvm.ConstInt(c.ctx.Int32Type(), 0, false),
		})
		fieldGlobalValue = llvm.ConstInsertValue(fieldGlobalValue, fieldName, []uint32{1})
		if typ.Tag(i) != "" {
			fieldTag := c.makeGlobalArray([]byte(typ.Tag(i)), "reflect/types.structFieldTag", c.ctx.Int8Type())
			fieldTag.SetLinkage(llvm.PrivateLinkage)
			fieldTag.SetUnnamedAddr(true)
			fieldTag = llvm.ConstGEP(fieldTag, []llvm.Value{
				llvm.ConstInt(c.ctx.Int32Type(), 0, false),
				llvm.ConstInt(c.ctx.Int32Type(), 0, false),
			})
			fieldGlobalValue = llvm.ConstInsertValue(fieldGlobalValue, fieldTag, []uint32{2})
		}
		if typ.Field(i).Embedded() {
			fieldEmbedded := llvm.ConstInt(c.ctx.Int1Type(), 1, false)
			fieldGlobalValue = llvm.ConstInsertValue(fieldGlobalValue, fieldEmbedded, []uint32{3})
		}
		structGlobalValue = llvm.ConstInsertValue(structGlobalValue, fieldGlobalValue, []uint32{uint32(i)})
	}
	structGlobal.SetInitializer(structGlobalValue)
	structGlobal.SetUnnamedAddr(true)
	structGlobal.SetLinkage(llvm.PrivateLinkage)
	return structGlobal
}

// getTypeCodeName returns a name for this type that can be used in the
// interface lowering pass to assign type codes as expected by the reflect
// package. See getTypeCodeNum.
func getTypeCodeName(t types.Type) string {
	switch t := t.(type) {
	case *types.Named:
		return "named:" + t.String()
	case *types.Array:
		return "array:" + strconv.FormatInt(t.Len(), 10) + ":" + getTypeCodeName(t.Elem())
	case *types.Basic:
		var kind string
		switch t.Kind() {
		case types.Bool:
			kind = "bool"
		case types.Int:
			kind = "int"
		case types.Int8:
			kind = "int8"
		case types.Int16:
			kind = "int16"
		case types.Int32:
			kind = "int32"
		case types.Int64:
			kind = "int64"
		case types.Uint:
			kind = "uint"
		case types.Uint8:
			kind = "uint8"
		case types.Uint16:
			kind = "uint16"
		case types.Uint32:
			kind = "uint32"
		case types.Uint64:
			kind = "uint64"
		case types.Uintptr:
			kind = "uintptr"
		case types.Float32:
			kind = "float32"
		case types.Float64:
			kind = "float64"
		case types.Complex64:
			kind = "complex64"
		case types.Complex128:
			kind = "complex128"
		case types.String:
			kind = "string"
		case types.UnsafePointer:
			kind = "unsafeptr"
		default:
			panic("unknown basic type: " + t.Name())
		}
		return "basic:" + kind
	case *types.Chan:
		return "chan:" + getTypeCodeName(t.Elem())
	case *types.Interface:
		methods := make([]string, t.NumMethods())
		for i := 0; i < t.NumMethods(); i++ {
			name := t.Method(i).Name()
			if !token.IsExported(name) {
				name = t.Method(i).Pkg().Path() + "." + name
			}
			methods[i] = name + ":" + getTypeCodeName(t.Method(i).Type())
		}
		return "interface:" + "{" + strings.Join(methods, ",") + "}"
	case *types.Map:
		keyType := getTypeCodeName(t.Key())
		elemType := getTypeCodeName(t.Elem())
		return "map:" + "{" + keyType + "," + elemType + "}"
	case *types.Pointer:
		return "pointer:" + getTypeCodeName(t.Elem())
	case *types.Signature:
		params := make([]string, t.Params().Len())
		for i := 0; i < t.Params().Len(); i++ {
			params[i] = getTypeCodeName(t.Params().At(i).Type())
		}
		results := make([]string, t.Results().Len())
		for i := 0; i < t.Results().Len(); i++ {
			results[i] = getTypeCodeName(t.Results().At(i).Type())
		}
		return "func:" + "{" + strings.Join(params, ",") + "}{" + strings.Join(results, ",") + "}"
	case *types.Slice:
		return "slice:" + getTypeCodeName(t.Elem())
	case *types.Struct:
		elems := make([]string, t.NumFields())
		for i := 0; i < t.NumFields(); i++ {
			embedded := ""
			if t.Field(i).Embedded() {
				embedded = "#"
			}
			elems[i] = embedded + t.Field(i).Name() + ":" + getTypeCodeName(t.Field(i).Type())
			if t.Tag(i) != "" {
				elems[i] += "`" + t.Tag(i) + "`"
			}
		}
		return "struct:" + "{" + strings.Join(elems, ",") + "}"
	default:
		panic("unknown type: " + t.String())
	}
}

// getTypeMethodSet returns a reference (GEP) to a global method set. This
// method set should be unreferenced after the interface lowering pass.
func (c *compilerContext) getTypeMethodSet(typ types.Type) llvm.Value {
	global := c.mod.NamedGlobal(typ.String() + "$methodset")
	zero := llvm.ConstInt(c.ctx.Int32Type(), 0, false)
	if !global.IsNil() {
		// the method set already exists
		return llvm.ConstGEP(global, []llvm.Value{zero, zero})
	}

	ms := c.program.MethodSets.MethodSet(typ)
	if ms.Len() == 0 {
		// no methods, so can leave that one out
		return llvm.ConstPointerNull(llvm.PointerType(c.getLLVMRuntimeType("interfaceMethodInfo"), 0))
	}

	methods := make([]llvm.Value, ms.Len())
	interfaceMethodInfoType := c.getLLVMRuntimeType("interfaceMethodInfo")
	for i := 0; i < ms.Len(); i++ {
		method := ms.At(i)
		signatureGlobal := c.getMethodSignature(method.Obj().(*types.Func))
		fn := c.program.MethodValue(method)
		llvmFn := c.getFunction(fn)
		if llvmFn.IsNil() {
			// compiler error, so panic
			panic("cannot find function: " + c.getFunctionInfo(fn).linkName)
		}
		wrapper := c.getInterfaceInvokeWrapper(fn, llvmFn)
		methodInfo := llvm.ConstNamedStruct(interfaceMethodInfoType, []llvm.Value{
			signatureGlobal,
			llvm.ConstPtrToInt(wrapper, c.uintptrType),
		})
		methods[i] = methodInfo
	}
	arrayType := llvm.ArrayType(interfaceMethodInfoType, len(methods))
	value := llvm.ConstArray(interfaceMethodInfoType, methods)
	global = llvm.AddGlobal(c.mod, arrayType, typ.String()+"$methodset")
	global.SetInitializer(value)
	global.SetGlobalConstant(true)
	global.SetLinkage(llvm.LinkOnceODRLinkage)
	return llvm.ConstGEP(global, []llvm.Value{zero, zero})
}

// getInterfaceMethodSet returns a global variable with the method set of the
// given named interface type. This method set is used by the interface lowering
// pass.
func (c *compilerContext) getInterfaceMethodSet(typ types.Type) llvm.Value {
	name := typ.String()
	if _, ok := typ.(*types.Named); !ok {
		// Anonymous interface.
		name = "reflect/types.interface:" + name
	}
	global := c.mod.NamedGlobal(name + "$interface")
	zero := llvm.ConstInt(c.ctx.Int32Type(), 0, false)
	if !global.IsNil() {
		// method set already exist, return it
		return llvm.ConstGEP(global, []llvm.Value{zero, zero})
	}

	// Every method is a *i8 reference indicating the signature of this method.
	methods := make([]llvm.Value, typ.Underlying().(*types.Interface).NumMethods())
	for i := range methods {
		method := typ.Underlying().(*types.Interface).Method(i)
		methods[i] = c.getMethodSignature(method)
	}

	value := llvm.ConstArray(c.i8ptrType, methods)
	global = llvm.AddGlobal(c.mod, value.Type(), name+"$interface")
	global.SetInitializer(value)
	global.SetGlobalConstant(true)
	global.SetLinkage(llvm.LinkOnceODRLinkage)
	return llvm.ConstGEP(global, []llvm.Value{zero, zero})
}

// getMethodSignatureName returns a unique name (that can be used as the name of
// a global) for the given method.
func (c *compilerContext) getMethodSignatureName(method *types.Func) string {
	signature := methodSignature(method)
	var globalName string
	if token.IsExported(method.Name()) {
		globalName = "reflect/methods." + signature
	} else {
		globalName = method.Type().(*types.Signature).Recv().Pkg().Path() + ".$methods." + signature
	}
	return globalName
}

// getMethodSignature returns a global variable which is a reference to an
// external *i8 indicating the indicating the signature of this method. It is
// used during the interface lowering pass.
func (c *compilerContext) getMethodSignature(method *types.Func) llvm.Value {
	globalName := c.getMethodSignatureName(method)
	signatureGlobal := c.mod.NamedGlobal(globalName)
	if signatureGlobal.IsNil() {
		// TODO: put something useful in these globals, such as the method
		// signature. Useful to one day implement reflect.Value.Method(n).
		signatureGlobal = llvm.AddGlobal(c.mod, c.ctx.Int8Type(), globalName)
		signatureGlobal.SetInitializer(llvm.ConstInt(c.ctx.Int8Type(), 0, false))
		signatureGlobal.SetLinkage(llvm.LinkOnceODRLinkage)
		signatureGlobal.SetGlobalConstant(true)
		signatureGlobal.SetAlignment(1)
	}
	return signatureGlobal
}

// createTypeAssert will emit the code for a typeassert, used in if statements
// and in type switches (Go SSA does not have type switches, only if/else
// chains). Note that even though the Go SSA does not contain type switches,
// LLVM will recognize the pattern and make it a real switch in many cases.
//
// Type asserts on concrete types are trivial: just compare type numbers. Type
// asserts on interfaces are more difficult, see the comments in the function.
func (b *builder) createTypeAssert(expr *ssa.TypeAssert) llvm.Value {
	itf := b.getValue(expr.X)
	assertedType := b.getLLVMType(expr.AssertedType)

	actualTypeNum := b.CreateExtractValue(itf, 0, "interface.type")
	commaOk := llvm.Value{}
	if _, ok := expr.AssertedType.Underlying().(*types.Interface); ok {
		// Type assert on interface type.
		// This is a call to an interface type assert function.
		// The interface lowering pass will define this function by filling it
		// with a type switch over all concrete types that implement this
		// interface, and returning whether it's one of the matched types.
		// This is very different from how interface asserts are implemented in
		// the main Go compiler, where the runtime checks whether the type
		// implements each method of the interface. See:
		// https://research.swtch.com/interfaces
		fn := b.getInterfaceImplementsFunc(expr.AssertedType)
		commaOk = b.CreateCall(fn, []llvm.Value{actualTypeNum}, "")

	} else {
		globalName := "reflect/types.typeid:" + getTypeCodeName(expr.AssertedType)
		assertedTypeCodeGlobal := b.mod.NamedGlobal(globalName)
		if assertedTypeCodeGlobal.IsNil() {
			// Create a new typecode global.
			assertedTypeCodeGlobal = llvm.AddGlobal(b.mod, b.ctx.Int8Type(), globalName)
			assertedTypeCodeGlobal.SetGlobalConstant(true)
		}
		// Type assert on concrete type.
		// Call runtime.typeAssert, which will be lowered to a simple icmp or
		// const false in the interface lowering pass.
		commaOk = b.createRuntimeCall("typeAssert", []llvm.Value{actualTypeNum, assertedTypeCodeGlobal}, "typecode")
	}

	// Add 2 new basic blocks (that should get optimized away): one for the
	// 'ok' case and one for all instructions following this type assert.
	// This is necessary because we need to insert the casted value or the
	// nil value based on whether the assert was successful. Casting before
	// this check tells LLVM that it can use this value and may
	// speculatively dereference pointers before the check. This can lead to
	// a miscompilation resulting in a segfault at runtime.
	// Additionally, this is even required by the Go spec: a failed
	// typeassert should return a zero value, not an incorrectly casted
	// value.

	prevBlock := b.GetInsertBlock()
	okBlock := b.ctx.AddBasicBlock(b.llvmFn, "typeassert.ok")
	nextBlock := b.ctx.AddBasicBlock(b.llvmFn, "typeassert.next")
	b.blockExits[b.currentBlock] = nextBlock // adjust outgoing block for phi nodes
	b.CreateCondBr(commaOk, okBlock, nextBlock)

	// Retrieve the value from the interface if the type assert was
	// successful.
	b.SetInsertPointAtEnd(okBlock)
	var valueOk llvm.Value
	if _, ok := expr.AssertedType.Underlying().(*types.Interface); ok {
		// Type assert on interface type. Easy: just return the same
		// interface value.
		valueOk = itf
	} else {
		// Type assert on concrete type. Extract the underlying type from
		// the interface (but only after checking it matches).
		valueOk = b.extractValueFromInterface(itf, assertedType)
	}
	b.CreateBr(nextBlock)

	// Continue after the if statement.
	b.SetInsertPointAtEnd(nextBlock)
	phi := b.CreatePHI(assertedType, "typeassert.value")
	phi.AddIncoming([]llvm.Value{llvm.ConstNull(assertedType), valueOk}, []llvm.BasicBlock{prevBlock, okBlock})

	if expr.CommaOk {
		tuple := b.ctx.ConstStruct([]llvm.Value{llvm.Undef(assertedType), llvm.Undef(b.ctx.Int1Type())}, false) // create empty tuple
		tuple = b.CreateInsertValue(tuple, phi, 0, "")                                                          // insert value
		tuple = b.CreateInsertValue(tuple, commaOk, 1, "")                                                      // insert 'comma ok' boolean
		return tuple
	} else {
		// This is kind of dirty as the branch above becomes mostly useless,
		// but hopefully this gets optimized away.
		b.createRuntimeCall("interfaceTypeAssert", []llvm.Value{commaOk}, "")
		return phi
	}
}

// getMethodsString returns a string to be used in the "tinygo-methods" string
// attribute for interface functions.
func (c *compilerContext) getMethodsString(itf *types.Interface) string {
	methods := make([]string, itf.NumMethods())
	for i := range methods {
		methods[i] = c.getMethodSignatureName(itf.Method(i))
	}
	return strings.Join(methods, "; ")
}

// getInterfaceImplementsfunc returns a declared function that works as a type
// switch. The interface lowering pass will define this function.
func (c *compilerContext) getInterfaceImplementsFunc(assertedType types.Type) llvm.Value {
	fnName := getTypeCodeName(assertedType.Underlying()) + ".$typeassert"
	llvmFn := c.mod.NamedFunction(fnName)
	if llvmFn.IsNil() {
		llvmFnType := llvm.FunctionType(c.ctx.Int1Type(), []llvm.Type{c.uintptrType}, false)
		llvmFn = llvm.AddFunction(c.mod, fnName, llvmFnType)
		c.addStandardDeclaredAttributes(llvmFn)
		methods := c.getMethodsString(assertedType.Underlying().(*types.Interface))
		llvmFn.AddFunctionAttr(c.ctx.CreateStringAttribute("tinygo-methods", methods))
	}
	return llvmFn
}

// getInvokeFunction returns the thunk to call the given interface method. The
// thunk is declared, not defined: it will be defined by the interface lowering
// pass.
func (c *compilerContext) getInvokeFunction(instr *ssa.CallCommon) llvm.Value {
	fnName := getTypeCodeName(instr.Value.Type().Underlying()) + "." + instr.Method.Name() + "$invoke"
	llvmFn := c.mod.NamedFunction(fnName)
	if llvmFn.IsNil() {
		sig := instr.Method.Type().(*types.Signature)
		var paramTuple []*types.Var
		for i := 0; i < sig.Params().Len(); i++ {
			paramTuple = append(paramTuple, sig.Params().At(i))
		}
		paramTuple = append(paramTuple, types.NewVar(token.NoPos, nil, "$typecode", types.Typ[types.Uintptr]))
		llvmFnType := c.getRawFuncType(types.NewSignature(sig.Recv(), types.NewTuple(paramTuple...), sig.Results(), false)).ElementType()
		llvmFn = llvm.AddFunction(c.mod, fnName, llvmFnType)
		c.addStandardDeclaredAttributes(llvmFn)
		llvmFn.AddFunctionAttr(c.ctx.CreateStringAttribute("tinygo-invoke", c.getMethodSignatureName(instr.Method)))
		methods := c.getMethodsString(instr.Value.Type().Underlying().(*types.Interface))
		llvmFn.AddFunctionAttr(c.ctx.CreateStringAttribute("tinygo-methods", methods))
	}
	return llvmFn
}

// getInterfaceInvokeWrapper returns a wrapper for the given method so it can be
// invoked from an interface. The wrapper takes in a pointer to the underlying
// value, dereferences or unpacks it if necessary, and calls the real method.
// If the method to wrap has a pointer receiver, no wrapping is necessary and
// the function is returned directly.
func (c *compilerContext) getInterfaceInvokeWrapper(fn *ssa.Function, llvmFn llvm.Value) llvm.Value {
	wrapperName := llvmFn.Name() + "$invoke"
	wrapper := c.mod.NamedFunction(wrapperName)
	if !wrapper.IsNil() {
		// Wrapper already created. Return it directly.
		return wrapper
	}

	// Get the expanded receiver type.
	receiverType := c.getLLVMType(fn.Signature.Recv().Type())
	var expandedReceiverType []llvm.Type
	for _, info := range c.expandFormalParamType(receiverType, "", nil) {
		expandedReceiverType = append(expandedReceiverType, info.llvmType)
	}

	// Does this method even need any wrapping?
	if len(expandedReceiverType) == 1 && receiverType.TypeKind() == llvm.PointerTypeKind {
		// Nothing to wrap.
		// Casting a function signature to a different signature and calling it
		// with a receiver pointer bitcasted to *i8 (as done in calls on an
		// interface) is hopefully a safe (defined) operation.
		return llvmFn
	}

	// create wrapper function
	fnType := llvmFn.Type().ElementType()
	paramTypes := append([]llvm.Type{c.i8ptrType}, fnType.ParamTypes()[len(expandedReceiverType):]...)
	wrapFnType := llvm.FunctionType(fnType.ReturnType(), paramTypes, false)
	wrapper = llvm.AddFunction(c.mod, wrapperName, wrapFnType)
	c.addStandardAttributes(wrapper)

	wrapper.SetLinkage(llvm.LinkOnceODRLinkage)
	wrapper.SetUnnamedAddr(true)

	// Create a new builder just to create this wrapper.
	b := builder{
		compilerContext: c,
		Builder:         c.ctx.NewBuilder(),
	}
	defer b.Builder.Dispose()

	// add debug info if needed
	if c.Debug {
		pos := c.program.Fset.Position(fn.Pos())
		difunc := c.attachDebugInfoRaw(fn, wrapper, "$invoke", pos.Filename, pos.Line)
		b.SetCurrentDebugLocation(uint(pos.Line), uint(pos.Column), difunc, llvm.Metadata{})
	}

	// set up IR builder
	block := b.ctx.AddBasicBlock(wrapper, "entry")
	b.SetInsertPointAtEnd(block)

	receiverValue := b.emitPointerUnpack(wrapper.Param(0), []llvm.Type{receiverType})[0]
	params := append(b.expandFormalParam(receiverValue), wrapper.Params()[1:]...)
	if llvmFn.Type().ElementType().ReturnType().TypeKind() == llvm.VoidTypeKind {
		b.CreateCall(llvmFn, params, "")
		b.CreateRetVoid()
	} else {
		ret := b.CreateCall(llvmFn, params, "ret")
		b.CreateRet(ret)
	}

	return wrapper
}

// methodSignature creates a readable version of a method signature (including
// the function name, excluding the receiver name). This string is used
// internally to match interfaces and to call the correct method on an
// interface. Examples:
//
//     String() string
//     Read([]byte) (int, error)
func methodSignature(method *types.Func) string {
	return method.Name() + signature(method.Type().(*types.Signature))
}

// Make a readable version of a function (pointer) signature.
// Examples:
//
//     () string
//     (string, int) (int, error)
func signature(sig *types.Signature) string {
	s := ""
	if sig.Params().Len() == 0 {
		s += "()"
	} else {
		s += "("
		for i := 0; i < sig.Params().Len(); i++ {
			if i > 0 {
				s += ", "
			}
			s += sig.Params().At(i).Type().String()
		}
		s += ")"
	}
	if sig.Results().Len() == 0 {
		// keep as-is
	} else if sig.Results().Len() == 1 {
		s += " " + sig.Results().At(0).Type().String()
	} else {
		s += " ("
		for i := 0; i < sig.Results().Len(); i++ {
			if i > 0 {
				s += ", "
			}
			s += sig.Results().At(i).Type().String()
		}
		s += ")"
	}
	return s
}
