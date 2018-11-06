package compiler

import (
	"errors"
	"go/types"

	"github.com/aykevl/go-llvm"
	"github.com/aykevl/tinygo/ir"
	"golang.org/x/tools/go/ssa"
)

// parseMakeInterface emits the LLVM IR for the *ssa.MakeInterface instruction.
// It tries to put the type in the interface value, but if that's not possible,
// it will do an allocation of the right size and put that in the interface
// value field.
//
// An interface value is a {typecode, value} tuple, or {i16, i8*} to be exact.
func (c *Compiler) parseMakeInterface(val llvm.Value, typ types.Type, global string) (llvm.Value, error) {
	var itfValue llvm.Value
	size := c.targetData.TypeAllocSize(val.Type())
	if size > c.targetData.TypeAllocSize(c.i8ptrType) {
		if global != "" {
			// Allocate in a global variable.
			global := llvm.AddGlobal(c.mod, val.Type(), global+"$itfvalue")
			global.SetInitializer(val)
			global.SetLinkage(llvm.InternalLinkage)
			global.SetGlobalConstant(true)
			zero := llvm.ConstInt(c.ctx.Int32Type(), 0, false)
			itfValueRaw := llvm.ConstInBoundsGEP(global, []llvm.Value{zero, zero})
			itfValue = llvm.ConstBitCast(itfValueRaw, c.i8ptrType)
		} else {
			// Allocate on the heap and put a pointer in the interface.
			// TODO: escape analysis.
			sizeValue := llvm.ConstInt(c.uintptrType, size, false)
			alloc := c.createRuntimeCall("alloc", []llvm.Value{sizeValue}, "")
			itfValueCast := c.builder.CreateBitCast(alloc, llvm.PointerType(val.Type(), 0), "")
			c.builder.CreateStore(val, itfValueCast)
			itfValue = c.builder.CreateBitCast(itfValueCast, c.i8ptrType, "")
		}
	} else {
		// Directly place the value in the interface.
		switch val.Type().TypeKind() {
		case llvm.IntegerTypeKind:
			itfValue = c.builder.CreateIntToPtr(val, c.i8ptrType, "")
		case llvm.PointerTypeKind:
			itfValue = c.builder.CreateBitCast(val, c.i8ptrType, "")
		case llvm.StructTypeKind:
			// A bitcast would be useful here, but bitcast doesn't allow
			// aggregate types. So we'll bitcast it using an alloca.
			// Hopefully this will get optimized away.
			mem := c.builder.CreateAlloca(c.i8ptrType, "")
			memStructPtr := c.builder.CreateBitCast(mem, llvm.PointerType(val.Type(), 0), "")
			c.builder.CreateStore(val, memStructPtr)
			itfValue = c.builder.CreateLoad(mem, "")
		default:
			return llvm.Value{}, errors.New("todo: makeinterface: cast small type to i8*")
		}
	}
	itfTypeNum, _ := c.ir.TypeNum(typ)
	if itfTypeNum >= 1<<16 {
		return llvm.Value{}, errors.New("interface typecodes do not fit in a 16-bit integer")
	}
	itf := llvm.ConstNamedStruct(c.mod.GetTypeByName("runtime._interface"), []llvm.Value{llvm.ConstInt(c.ctx.Int16Type(), uint64(itfTypeNum), false), llvm.Undef(c.i8ptrType)})
	itf = c.builder.CreateInsertValue(itf, itfValue, 1, "")
	return itf, nil
}

// parseTypeAssert will emit the code for a typeassert, used in if statements
// and in switch statements (Go SSA does not have type switches, only if/else
// chains). Note that even though the Go SSA does not contain type switches,
// LLVM will recognize the pattern and make it a real switch in many cases.
//
// Type asserts on concrete types are trivial: just compare type numbers. Type
// asserts on interfaces are more difficult to implement and so are delegated to
// a runtime library function.
func (c *Compiler) parseTypeAssert(frame *Frame, expr *ssa.TypeAssert) (llvm.Value, error) {
	itf, err := c.parseExpr(frame, expr.X)
	if err != nil {
		return llvm.Value{}, err
	}
	assertedType, err := c.getLLVMType(expr.AssertedType)
	if err != nil {
		return llvm.Value{}, err
	}
	valueNil, err := c.getZeroValue(assertedType)
	if err != nil {
		return llvm.Value{}, err
	}

	actualTypeNum := c.builder.CreateExtractValue(itf, 0, "interface.type")
	commaOk := llvm.Value{}
	if itf, ok := expr.AssertedType.Underlying().(*types.Interface); ok {
		// Type assert on interface type.
		// This is slightly non-trivial: at runtime the list of methods
		// needs to be checked to see whether it implements the interface.
		// At the same time, the interface value itself is unchanged.
		itfTypeNum := c.ir.InterfaceNum(itf)
		itfTypeNumValue := llvm.ConstInt(c.ctx.Int16Type(), uint64(itfTypeNum), false)
		commaOk = c.createRuntimeCall("interfaceImplements", []llvm.Value{actualTypeNum, itfTypeNumValue}, "")

	} else {
		// Type assert on concrete type.
		// This is easy: just compare the type number.
		assertedTypeNum, typeExists := c.ir.TypeNum(expr.AssertedType)
		if !typeExists {
			// Static analysis has determined this type assert will never apply.
			// Using undef here so that LLVM knows we'll never get here and
			// can optimize accordingly.
			undef := llvm.Undef(assertedType)
			commaOk := llvm.ConstInt(c.ctx.Int1Type(), 0, false)
			if expr.CommaOk {
				return c.ctx.ConstStruct([]llvm.Value{undef, commaOk}, false), nil
			} else {
				c.createRuntimeCall("interfaceTypeAssert", []llvm.Value{commaOk}, "")
				return undef, nil
			}
		}
		if assertedTypeNum >= 1<<16 {
			return llvm.Value{}, errors.New("interface typecodes do not fit in a 16-bit integer")
		}

		assertedTypeNumValue := llvm.ConstInt(c.ctx.Int16Type(), uint64(assertedTypeNum), false)
		commaOk = c.builder.CreateICmp(llvm.IntEQ, assertedTypeNumValue, actualTypeNum, "")
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

	prevBlock := c.builder.GetInsertBlock()
	okBlock := c.ctx.AddBasicBlock(frame.fn.LLVMFn, "typeassert.ok")
	nextBlock := c.ctx.AddBasicBlock(frame.fn.LLVMFn, "typeassert.next")
	frame.blockExits[frame.currentBlock] = nextBlock // adjust outgoing block for phi nodes
	c.builder.CreateCondBr(commaOk, okBlock, nextBlock)

	// Retrieve the value from the interface if the type assert was
	// successful.
	c.builder.SetInsertPointAtEnd(okBlock)
	var valueOk llvm.Value
	if _, ok := expr.AssertedType.Underlying().(*types.Interface); ok {
		// Type assert on interface type. Easy: just return the same
		// interface value.
		valueOk = itf
	} else {
		// Type assert on concrete type. Extract the underlying type from
		// the interface (but only after checking it matches).
		valuePtr := c.builder.CreateExtractValue(itf, 1, "typeassert.value.ptr")
		if c.targetData.TypeAllocSize(assertedType) > c.targetData.TypeAllocSize(c.i8ptrType) {
			// Value was stored in an allocated buffer, load it from there.
			valuePtrCast := c.builder.CreateBitCast(valuePtr, llvm.PointerType(assertedType, 0), "")
			valueOk = c.builder.CreateLoad(valuePtrCast, "typeassert.value.ok")
		} else {
			// Value was stored directly in the interface.
			switch assertedType.TypeKind() {
			case llvm.IntegerTypeKind:
				valueOk = c.builder.CreatePtrToInt(valuePtr, assertedType, "typeassert.value.ok")
			case llvm.PointerTypeKind:
				valueOk = c.builder.CreateBitCast(valuePtr, assertedType, "typeassert.value.ok")
			case llvm.StructTypeKind:
				// A bitcast would be useful here, but bitcast doesn't allow
				// aggregate types. So we'll bitcast it using an alloca.
				// Hopefully this will get optimized away.
				mem := c.builder.CreateAlloca(c.i8ptrType, "")
				c.builder.CreateStore(valuePtr, mem)
				memStructPtr := c.builder.CreateBitCast(mem, llvm.PointerType(assertedType, 0), "")
				valueOk = c.builder.CreateLoad(memStructPtr, "typeassert.value.ok")
			default:
				return llvm.Value{}, errors.New("todo: typeassert: bitcast small types")
			}
		}
	}
	c.builder.CreateBr(nextBlock)

	// Continue after the if statement.
	c.builder.SetInsertPointAtEnd(nextBlock)
	phi := c.builder.CreatePHI(assertedType, "typeassert.value")
	phi.AddIncoming([]llvm.Value{valueNil, valueOk}, []llvm.BasicBlock{prevBlock, okBlock})

	if expr.CommaOk {
		tuple := c.ctx.ConstStruct([]llvm.Value{llvm.Undef(assertedType), llvm.Undef(c.ctx.Int1Type())}, false) // create empty tuple
		tuple = c.builder.CreateInsertValue(tuple, phi, 0, "")                                                  // insert value
		tuple = c.builder.CreateInsertValue(tuple, commaOk, 1, "")                                              // insert 'comma ok' boolean
		return tuple, nil
	} else {
		// This is kind of dirty as the branch above becomes mostly useless,
		// but hopefully this gets optimized away.
		c.createRuntimeCall("interfaceTypeAssert", []llvm.Value{commaOk}, "")
		return phi, nil
	}
}

// getInvokeCall creates and returns the function pointer and parameters of an
// interface call. It can be used in a call or defer instruction.
func (c *Compiler) getInvokeCall(frame *Frame, instr *ssa.CallCommon) (llvm.Value, []llvm.Value, error) {
	// Call an interface method with dynamic dispatch.
	itf, err := c.parseExpr(frame, instr.Value) // interface
	if err != nil {
		return llvm.Value{}, nil, err
	}

	llvmFnType, err := c.getLLVMType(instr.Method.Type())
	if err != nil {
		return llvm.Value{}, nil, err
	}
	if c.ir.SignatureNeedsContext(instr.Method.Type().(*types.Signature)) {
		// This is somewhat of a hack.
		// getLLVMType() has created a closure type for us, but we don't
		// actually want a closure type as an interface call can never be a
		// closure call. So extract the function pointer type from the
		// closure.
		// This happens because somewhere the same function signature is
		// used in a closure or bound method.
		llvmFnType = llvmFnType.Subtypes()[1]
	}

	typecode := c.builder.CreateExtractValue(itf, 0, "invoke.typecode")
	values := []llvm.Value{
		typecode,
		llvm.ConstInt(c.ctx.Int16Type(), uint64(c.ir.MethodNum(instr.Method)), false),
	}
	fn := c.createRuntimeCall("interfaceMethod", values, "invoke.func")
	fnCast := c.builder.CreateBitCast(fn, llvmFnType, "invoke.func.cast")
	receiverValue := c.builder.CreateExtractValue(itf, 1, "invoke.func.receiver")

	args := []llvm.Value{receiverValue}
	for _, arg := range instr.Args {
		val, err := c.parseExpr(frame, arg)
		if err != nil {
			return llvm.Value{}, nil, err
		}
		args = append(args, val)
	}
	if c.ir.SignatureNeedsContext(instr.Method.Type().(*types.Signature)) {
		// This function takes an extra context parameter. An interface call
		// cannot also be a closure but we have to supply the nil pointer
		// anyway.
		args = append(args, llvm.ConstPointerNull(c.i8ptrType))
	}

	return fnCast, args, nil
}

// Initialize runtime type information, for interfaces.
// See src/runtime/interface.go for more details.
func (c *Compiler) createInterfaceRTTI() error {
	dynamicTypes := c.ir.AllDynamicTypes()
	numDynamicTypes := 0
	for _, meta := range dynamicTypes {
		numDynamicTypes += len(meta.Methods)
	}
	ranges := make([]llvm.Value, 0, len(dynamicTypes))
	funcPointers := make([]llvm.Value, 0, numDynamicTypes)
	signatures := make([]llvm.Value, 0, numDynamicTypes)
	startIndex := 0
	rangeType := c.mod.GetTypeByName("runtime.methodSetRange")
	for _, meta := range dynamicTypes {
		rangeValues := []llvm.Value{
			llvm.ConstInt(c.ctx.Int16Type(), uint64(startIndex), false),
			llvm.ConstInt(c.ctx.Int16Type(), uint64(len(meta.Methods)), false),
		}
		rangeValue := llvm.ConstNamedStruct(rangeType, rangeValues)
		ranges = append(ranges, rangeValue)
		methods := make([]*types.Selection, 0, len(meta.Methods))
		for _, method := range meta.Methods {
			methods = append(methods, method)
		}
		c.ir.SortMethods(methods)
		for _, method := range methods {
			f := c.ir.GetFunction(c.ir.Program.MethodValue(method))
			if f.LLVMFn.IsNil() {
				return errors.New("cannot find function: " + f.LinkName())
			}
			fn, err := c.wrapInterfaceInvoke(f)
			if err != nil {
				return err
			}
			fnPtr := llvm.ConstBitCast(fn, c.i8ptrType)
			funcPointers = append(funcPointers, fnPtr)
			signatureNum := c.ir.MethodNum(method.Obj().(*types.Func))
			signature := llvm.ConstInt(c.ctx.Int16Type(), uint64(signatureNum), false)
			signatures = append(signatures, signature)
		}
		startIndex += len(meta.Methods)
	}

	interfaceTypes := c.ir.AllInterfaces()
	interfaceIndex := make([]llvm.Value, len(interfaceTypes))
	interfaceLengths := make([]llvm.Value, len(interfaceTypes))
	interfaceMethods := make([]llvm.Value, 0)
	for i, itfType := range interfaceTypes {
		if itfType.Type.NumMethods() > 0xff {
			return errors.New("too many methods for interface " + itfType.Type.String())
		}
		interfaceIndex[i] = llvm.ConstInt(c.ctx.Int16Type(), uint64(i), false)
		interfaceLengths[i] = llvm.ConstInt(c.ctx.Int8Type(), uint64(itfType.Type.NumMethods()), false)
		funcs := make([]*types.Func, itfType.Type.NumMethods())
		for i := range funcs {
			funcs[i] = itfType.Type.Method(i)
		}
		c.ir.SortFuncs(funcs)
		for _, f := range funcs {
			id := llvm.ConstInt(c.ctx.Int16Type(), uint64(c.ir.MethodNum(f)), false)
			interfaceMethods = append(interfaceMethods, id)
		}
	}

	if len(ranges) >= 1<<16 {
		return errors.New("method call numbers do not fit in a 16-bit integer")
	}

	// Replace the pre-created arrays with the generated arrays.
	rangeArray := llvm.ConstArray(rangeType, ranges)
	rangeArrayNewGlobal := llvm.AddGlobal(c.mod, rangeArray.Type(), "runtime.methodSetRanges.tmp")
	rangeArrayNewGlobal.SetInitializer(rangeArray)
	rangeArrayNewGlobal.SetLinkage(llvm.InternalLinkage)
	rangeArrayOldGlobal := c.mod.NamedGlobal("runtime.methodSetRanges")
	rangeArrayOldGlobal.ReplaceAllUsesWith(llvm.ConstBitCast(rangeArrayNewGlobal, rangeArrayOldGlobal.Type()))
	rangeArrayOldGlobal.EraseFromParentAsGlobal()
	rangeArrayNewGlobal.SetName("runtime.methodSetRanges")
	funcArray := llvm.ConstArray(c.i8ptrType, funcPointers)
	funcArrayNewGlobal := llvm.AddGlobal(c.mod, funcArray.Type(), "runtime.methodSetFunctions.tmp")
	funcArrayNewGlobal.SetInitializer(funcArray)
	funcArrayNewGlobal.SetLinkage(llvm.InternalLinkage)
	funcArrayOldGlobal := c.mod.NamedGlobal("runtime.methodSetFunctions")
	funcArrayOldGlobal.ReplaceAllUsesWith(llvm.ConstBitCast(funcArrayNewGlobal, funcArrayOldGlobal.Type()))
	funcArrayOldGlobal.EraseFromParentAsGlobal()
	funcArrayNewGlobal.SetName("runtime.methodSetFunctions")
	signatureArray := llvm.ConstArray(c.ctx.Int16Type(), signatures)
	signatureArrayNewGlobal := llvm.AddGlobal(c.mod, signatureArray.Type(), "runtime.methodSetSignatures.tmp")
	signatureArrayNewGlobal.SetInitializer(signatureArray)
	signatureArrayNewGlobal.SetLinkage(llvm.InternalLinkage)
	signatureArrayOldGlobal := c.mod.NamedGlobal("runtime.methodSetSignatures")
	signatureArrayOldGlobal.ReplaceAllUsesWith(llvm.ConstBitCast(signatureArrayNewGlobal, signatureArrayOldGlobal.Type()))
	signatureArrayOldGlobal.EraseFromParentAsGlobal()
	signatureArrayNewGlobal.SetName("runtime.methodSetSignatures")
	interfaceIndexArray := llvm.ConstArray(c.ctx.Int16Type(), interfaceIndex)
	interfaceIndexArrayNewGlobal := llvm.AddGlobal(c.mod, interfaceIndexArray.Type(), "runtime.interfaceIndex.tmp")
	interfaceIndexArrayNewGlobal.SetInitializer(interfaceIndexArray)
	interfaceIndexArrayNewGlobal.SetLinkage(llvm.InternalLinkage)
	interfaceIndexArrayOldGlobal := c.mod.NamedGlobal("runtime.interfaceIndex")
	interfaceIndexArrayOldGlobal.ReplaceAllUsesWith(llvm.ConstBitCast(interfaceIndexArrayNewGlobal, interfaceIndexArrayOldGlobal.Type()))
	interfaceIndexArrayOldGlobal.EraseFromParentAsGlobal()
	interfaceIndexArrayNewGlobal.SetName("runtime.interfaceIndex")
	interfaceLengthsArray := llvm.ConstArray(c.ctx.Int8Type(), interfaceLengths)
	interfaceLengthsArrayNewGlobal := llvm.AddGlobal(c.mod, interfaceLengthsArray.Type(), "runtime.interfaceLengths.tmp")
	interfaceLengthsArrayNewGlobal.SetInitializer(interfaceLengthsArray)
	interfaceLengthsArrayNewGlobal.SetLinkage(llvm.InternalLinkage)
	interfaceLengthsArrayOldGlobal := c.mod.NamedGlobal("runtime.interfaceLengths")
	interfaceLengthsArrayOldGlobal.ReplaceAllUsesWith(llvm.ConstBitCast(interfaceLengthsArrayNewGlobal, interfaceLengthsArrayOldGlobal.Type()))
	interfaceLengthsArrayOldGlobal.EraseFromParentAsGlobal()
	interfaceLengthsArrayNewGlobal.SetName("runtime.interfaceLengths")
	interfaceMethodsArray := llvm.ConstArray(c.ctx.Int16Type(), interfaceMethods)
	interfaceMethodsArrayNewGlobal := llvm.AddGlobal(c.mod, interfaceMethodsArray.Type(), "runtime.interfaceMethods.tmp")
	interfaceMethodsArrayNewGlobal.SetInitializer(interfaceMethodsArray)
	interfaceMethodsArrayNewGlobal.SetLinkage(llvm.InternalLinkage)
	interfaceMethodsArrayOldGlobal := c.mod.NamedGlobal("runtime.interfaceMethods")
	interfaceMethodsArrayOldGlobal.ReplaceAllUsesWith(llvm.ConstBitCast(interfaceMethodsArrayNewGlobal, interfaceMethodsArrayOldGlobal.Type()))
	interfaceMethodsArrayOldGlobal.EraseFromParentAsGlobal()
	interfaceMethodsArrayNewGlobal.SetName("runtime.interfaceMethods")

	c.mod.NamedGlobal("runtime.firstTypeWithMethods").SetInitializer(llvm.ConstInt(c.ctx.Int16Type(), uint64(c.ir.FirstDynamicType()), false))

	return nil
}

// Wrap an interface method function pointer. The wrapper takes in a pointer to
// the underlying value, dereferences it, and calls the real method. This
// wrapper is only needed when the interface value actually doesn't fit in a
// pointer and a pointer to the value must be created.
func (c *Compiler) wrapInterfaceInvoke(f *ir.Function) (llvm.Value, error) {
	receiverType, err := c.getLLVMType(f.Params[0].Type())
	if err != nil {
		return llvm.Value{}, err
	}
	expandedReceiverType := c.expandFormalParamType(receiverType)

	if c.targetData.TypeAllocSize(receiverType) <= c.targetData.TypeAllocSize(c.i8ptrType) && len(expandedReceiverType) == 1 {
		// nothing to wrap
		return f.LLVMFn, nil
	}

	// create wrapper function
	fnType := f.LLVMFn.Type().ElementType()
	paramTypes := append([]llvm.Type{c.i8ptrType}, fnType.ParamTypes()[len(expandedReceiverType):]...)
	wrapFnType := llvm.FunctionType(fnType.ReturnType(), paramTypes, false)
	wrapper := llvm.AddFunction(c.mod, f.LinkName()+"$invoke", wrapFnType)
	wrapper.SetLinkage(llvm.InternalLinkage)
	wrapper.SetUnnamedAddr(true)

	// add debug info
	if c.Debug {
		pos := c.ir.Program.Fset.Position(f.Pos())
		difunc, err := c.attachDebugInfoRaw(f, wrapper, "$invoke", pos.Filename, pos.Line)
		if err != nil {
			return llvm.Value{}, err
		}
		c.builder.SetCurrentDebugLocation(uint(pos.Line), uint(pos.Column), difunc, llvm.Metadata{})
	}

	// set up IR builder
	block := c.ctx.AddBasicBlock(wrapper, "entry")
	c.builder.SetInsertPointAtEnd(block)

	var receiverPtr llvm.Value
	if c.targetData.TypeAllocSize(receiverType) > c.targetData.TypeAllocSize(c.i8ptrType) {
		// The receiver is passed in using a pointer. We have to load it here
		// and pass it by value to the real function.

		// Load the underlying value.
		receiverPtrType := llvm.PointerType(receiverType, 0)
		receiverPtr = c.builder.CreateBitCast(wrapper.Param(0), receiverPtrType, "receiver.ptr")
	} else if len(expandedReceiverType) != 1 {
		// The value is stored in the interface, but it is of type struct which
		// is expanded to multiple parameters (e.g. {i8, i8}). So we have to
		// receive the struct as parameter, expand it, and pass it on to the
		// real function.

		// Cast the passed-in i8* to the struct value (using an alloca) and
		// extract its values.
		alloca := c.builder.CreateAlloca(c.i8ptrType, "receiver.alloca")
		c.builder.CreateStore(wrapper.Param(0), alloca)
		receiverPtr = c.builder.CreateBitCast(alloca, llvm.PointerType(receiverType, 0), "receiver.ptr")
	} else {
		panic("unreachable")
	}

	receiverValue := c.builder.CreateLoad(receiverPtr, "receiver")
	params := append(c.expandFormalParam(receiverValue), wrapper.Params()[1:]...)
	if fnType.ReturnType().TypeKind() == llvm.VoidTypeKind {
		c.builder.CreateCall(f.LLVMFn, params, "")
		c.builder.CreateRetVoid()
	} else {
		ret := c.builder.CreateCall(f.LLVMFn, params, "ret")
		c.builder.CreateRet(ret)
	}

	return wrapper, nil
}
