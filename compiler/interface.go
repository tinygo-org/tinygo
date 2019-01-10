package compiler

// This file transforms interface-related instructions (*ssa.MakeInterface,
// *ssa.TypeAssert, calls on interface types) to an intermediate IR form, to be
// lowered to the final form by the interface lowering pass. See
// interface-lowering.go for more details.

import (
	"go/token"
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
func (c *Compiler) parseMakeInterface(val llvm.Value, typ types.Type, global string, pos token.Pos) (llvm.Value, error) {
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
			alloc := c.createRuntimeCall("alloc", []llvm.Value{sizeValue}, "makeinterface.alloc")
			itfValueCast := c.builder.CreateBitCast(alloc, llvm.PointerType(val.Type(), 0), "makeinterface.cast.value")
			c.builder.CreateStore(val, itfValueCast)
			itfValue = c.builder.CreateBitCast(itfValueCast, c.i8ptrType, "makeinterface.cast.i8ptr")
		}
	} else if size == 0 {
		itfValue = llvm.ConstPointerNull(c.i8ptrType)
	} else {
		// Directly place the value in the interface.
		switch val.Type().TypeKind() {
		case llvm.IntegerTypeKind:
			itfValue = c.builder.CreateIntToPtr(val, c.i8ptrType, "makeinterface.cast.int")
		case llvm.PointerTypeKind:
			itfValue = c.builder.CreateBitCast(val, c.i8ptrType, "makeinterface.cast.ptr")
		case llvm.StructTypeKind:
			// A bitcast would be useful here, but bitcast doesn't allow
			// aggregate types. So we'll bitcast it using an alloca.
			// Hopefully this will get optimized away.
			mem := c.builder.CreateAlloca(c.i8ptrType, "makeinterface.cast.struct")
			memStructPtr := c.builder.CreateBitCast(mem, llvm.PointerType(val.Type(), 0), "makeinterface.cast.struct.cast")
			c.builder.CreateStore(val, memStructPtr)
			itfValue = c.builder.CreateLoad(mem, "makeinterface.cast.load")
		default:
			return llvm.Value{}, c.makeError(pos, "todo: makeinterface: cast small type to i8*")
		}
	}
	itfTypeCodeGlobal := c.getTypeCode(typ)
	itfMethodSetGlobal, err := c.getTypeMethodSet(typ)
	if err != nil {
		return llvm.Value{}, nil
	}
	itfTypeCode := c.createRuntimeCall("makeInterface", []llvm.Value{itfTypeCodeGlobal, itfMethodSetGlobal}, "makeinterface.typecode")
	itf := llvm.Undef(c.mod.GetTypeByName("runtime._interface"))
	itf = c.builder.CreateInsertValue(itf, itfTypeCode, 0, "")
	itf = c.builder.CreateInsertValue(itf, itfValue, 1, "")
	return itf, nil
}

// getTypeCode returns a reference to a type code.
// It returns a pointer to an external global which should be replaced with the
// real type in the interface lowering pass.
func (c *Compiler) getTypeCode(typ types.Type) llvm.Value {
	global := c.mod.NamedGlobal(typ.String() + "$type")
	if global.IsNil() {
		global = llvm.AddGlobal(c.mod, c.ctx.Int8Type(), typ.String()+"$type")
		global.SetGlobalConstant(true)
	}
	return global
}

// getTypeMethodSet returns a reference (GEP) to a global method set. This
// method set should be unreferenced after the interface lowering pass.
func (c *Compiler) getTypeMethodSet(typ types.Type) (llvm.Value, error) {
	global := c.mod.NamedGlobal(typ.String() + "$methodset")
	zero := llvm.ConstInt(c.ctx.Int32Type(), 0, false)
	if !global.IsNil() {
		// the method set already exists
		return llvm.ConstGEP(global, []llvm.Value{zero, zero}), nil
	}

	ms := c.ir.Program.MethodSets.MethodSet(typ)
	if ms.Len() == 0 {
		// no methods, so can leave that one out
		return llvm.ConstPointerNull(llvm.PointerType(c.mod.GetTypeByName("runtime.interfaceMethodInfo"), 0)), nil
	}

	methods := make([]llvm.Value, ms.Len())
	interfaceMethodInfoType := c.mod.GetTypeByName("runtime.interfaceMethodInfo")
	for i := 0; i < ms.Len(); i++ {
		method := ms.At(i)
		signatureGlobal := c.getMethodSignature(method.Obj().(*types.Func))
		f := c.ir.GetFunction(c.ir.Program.MethodValue(method))
		if f.LLVMFn.IsNil() {
			// compiler error, so panic
			panic("cannot find function: " + f.LinkName())
		}
		fn, err := c.getInterfaceInvokeWrapper(f)
		if err != nil {
			return llvm.Value{}, err
		}
		methodInfo := llvm.ConstNamedStruct(interfaceMethodInfoType, []llvm.Value{
			signatureGlobal,
			llvm.ConstBitCast(fn, c.i8ptrType),
		})
		methods[i] = methodInfo
	}
	arrayType := llvm.ArrayType(interfaceMethodInfoType, len(methods))
	value := llvm.ConstArray(interfaceMethodInfoType, methods)
	global = llvm.AddGlobal(c.mod, arrayType, typ.String()+"$methodset")
	global.SetInitializer(value)
	global.SetGlobalConstant(true)
	global.SetLinkage(llvm.PrivateLinkage)
	return llvm.ConstGEP(global, []llvm.Value{zero, zero}), nil
}

// getInterfaceMethodSet returns a global variable with the method set of the
// given named interface type. This method set is used by the interface lowering
// pass.
func (c *Compiler) getInterfaceMethodSet(typ *types.Named) llvm.Value {
	global := c.mod.NamedGlobal(typ.String() + "$interface")
	zero := llvm.ConstInt(c.ctx.Int32Type(), 0, false)
	if !global.IsNil() {
		// method set already exist, return it
		return llvm.ConstGEP(global, []llvm.Value{zero, zero})
	}

	// Every method is a *i16 reference indicating the signature of this method.
	methods := make([]llvm.Value, typ.Underlying().(*types.Interface).NumMethods())
	for i := range methods {
		method := typ.Underlying().(*types.Interface).Method(i)
		methods[i] = c.getMethodSignature(method)
	}

	value := llvm.ConstArray(methods[0].Type(), methods)
	global = llvm.AddGlobal(c.mod, value.Type(), typ.String()+"$interface")
	global.SetInitializer(value)
	global.SetGlobalConstant(true)
	global.SetLinkage(llvm.PrivateLinkage)
	return llvm.ConstGEP(global, []llvm.Value{zero, zero})
}

// getMethodSignature returns a global variable which is a reference to an
// external *i16 indicating the indicating the signature of this method. It is
// used during the interface lowering pass.
func (c *Compiler) getMethodSignature(method *types.Func) llvm.Value {
	signature := ir.MethodSignature(method)
	signatureGlobal := c.mod.NamedGlobal("func " + signature)
	if signatureGlobal.IsNil() {
		signatureGlobal = llvm.AddGlobal(c.mod, c.ctx.Int8Type(), "func "+signature)
		signatureGlobal.SetGlobalConstant(true)
	}
	return signatureGlobal
}

// parseTypeAssert will emit the code for a typeassert, used in if statements
// and in type switches (Go SSA does not have type switches, only if/else
// chains). Note that even though the Go SSA does not contain type switches,
// LLVM will recognize the pattern and make it a real switch in many cases.
//
// Type asserts on concrete types are trivial: just compare type numbers. Type
// asserts on interfaces are more difficult, see the comments in the function.
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
	if _, ok := expr.AssertedType.Underlying().(*types.Interface); ok {
		// Type assert on interface type.
		// This pseudo call will be lowered in the interface lowering pass to a
		// real call which checks whether the provided typecode is any of the
		// concrete types that implements this interface.
		// This is very different from how interface asserts are implemented in
		// the main Go compiler, where the runtime checks whether the type
		// implements each method of the interface. See:
		// https://research.swtch.com/interfaces
		methodSet := c.getInterfaceMethodSet(expr.AssertedType.(*types.Named))
		commaOk = c.createRuntimeCall("interfaceImplements", []llvm.Value{actualTypeNum, methodSet}, "")

	} else {
		// Type assert on concrete type.
		// Call runtime.typeAssert, which will be lowered to a simple icmp or
		// const false in the interface lowering pass.
		assertedTypeCodeGlobal := c.getTypeCode(expr.AssertedType)
		commaOk = c.createRuntimeCall("typeAssert", []llvm.Value{actualTypeNum, assertedTypeCodeGlobal}, "typecode")
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
		size := c.targetData.TypeAllocSize(assertedType)
		if size > c.targetData.TypeAllocSize(c.i8ptrType) {
			// Value was stored in an allocated buffer, load it from there.
			valuePtrCast := c.builder.CreateBitCast(valuePtr, llvm.PointerType(assertedType, 0), "")
			valueOk = c.builder.CreateLoad(valuePtrCast, "typeassert.value.ok")
		} else if size == 0 {
			valueOk, err = c.getZeroValue(assertedType)
			if err != nil {
				return llvm.Value{}, err
			}
		} else {
			// Value was stored directly in the interface.
			switch assertedType.TypeKind() {
			case llvm.IntegerTypeKind:
				valueOk = c.builder.CreatePtrToInt(valuePtr, assertedType, "typeassert.value.ok")
			case llvm.PointerTypeKind:
				valueOk = c.builder.CreateBitCast(valuePtr, assertedType, "typeassert.value.ok")
			default: // struct, float, etc.
				// A bitcast would be useful here, but bitcast doesn't allow
				// aggregate types. So we'll bitcast it using an alloca.
				// Hopefully this will get optimized away.
				mem := c.builder.CreateAlloca(c.i8ptrType, "")
				c.builder.CreateStore(valuePtr, mem)
				memCast := c.builder.CreateBitCast(mem, llvm.PointerType(assertedType, 0), "")
				valueOk = c.builder.CreateLoad(memCast, "typeassert.value.ok")
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
	// getLLVMType() has created a closure type for us, but we don't actually
	// want a closure type as an interface call can never be a closure call. So
	// extract the function pointer type from the closure.
	llvmFnType = llvmFnType.Subtypes()[1]

	typecode := c.builder.CreateExtractValue(itf, 0, "invoke.typecode")
	values := []llvm.Value{
		typecode,
		c.getInterfaceMethodSet(instr.Value.Type().(*types.Named)),
		c.getMethodSignature(instr.Method),
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
	// Add the context parameter. An interface call never takes a context but we
	// have to supply the parameter anyway.
	args = append(args, llvm.Undef(c.i8ptrType))
	// Add the parent goroutine handle.
	args = append(args, llvm.Undef(c.i8ptrType))

	return fnCast, args, nil
}

// interfaceInvokeWrapper keeps some state between getInterfaceInvokeWrapper and
// createInterfaceInvokeWrapper. The former is called during IR construction
// itself and the latter is called when finishing up the IR.
type interfaceInvokeWrapper struct {
	fn           *ir.Function
	wrapper      llvm.Value
	receiverType llvm.Type
}

// Wrap an interface method function pointer. The wrapper takes in a pointer to
// the underlying value, dereferences it, and calls the real method. This
// wrapper is only needed when the interface value actually doesn't fit in a
// pointer and a pointer to the value must be created.
func (c *Compiler) getInterfaceInvokeWrapper(f *ir.Function) (llvm.Value, error) {
	wrapperName := f.LinkName() + "$invoke"
	wrapper := c.mod.NamedFunction(wrapperName)
	if !wrapper.IsNil() {
		// Wrapper already created. Return it directly.
		return wrapper, nil
	}

	// Get the expanded receiver type.
	receiverType, err := c.getLLVMType(f.Params[0].Type())
	if err != nil {
		return llvm.Value{}, err
	}
	expandedReceiverType := c.expandFormalParamType(receiverType)

	// Does this method even need any wrapping?
	if len(expandedReceiverType) == 1 && receiverType.TypeKind() == llvm.PointerTypeKind {
		// Nothing to wrap.
		// Casting a function signature to a different signature and calling it
		// with a receiver pointer bitcasted to *i8 (as done in calls on an
		// interface) is hopefully a safe (defined) operation.
		return f.LLVMFn, nil
	}

	// create wrapper function
	fnType := f.LLVMFn.Type().ElementType()
	paramTypes := append([]llvm.Type{c.i8ptrType}, fnType.ParamTypes()[len(expandedReceiverType):]...)
	wrapFnType := llvm.FunctionType(fnType.ReturnType(), paramTypes, false)
	wrapper = llvm.AddFunction(c.mod, wrapperName, wrapFnType)
	c.interfaceInvokeWrappers = append(c.interfaceInvokeWrappers, interfaceInvokeWrapper{
		fn:           f,
		wrapper:      wrapper,
		receiverType: receiverType,
	})
	return wrapper, nil
}

// createInterfaceInvokeWrapper finishes the work of getInterfaceInvokeWrapper,
// see that function for details.
func (c *Compiler) createInterfaceInvokeWrapper(state interfaceInvokeWrapper) error {
	wrapper := state.wrapper
	fn := state.fn
	receiverType := state.receiverType
	wrapper.SetLinkage(llvm.InternalLinkage)
	wrapper.SetUnnamedAddr(true)

	// add debug info if needed
	if c.Debug {
		pos := c.ir.Program.Fset.Position(fn.Pos())
		difunc, err := c.attachDebugInfoRaw(fn, wrapper, "$invoke", pos.Filename, pos.Line)
		if err != nil {
			return err
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
	} else {
		// The value is stored in the interface, but it is of type struct which
		// is expanded to multiple parameters (e.g. {i8, i8}). So we have to
		// receive the struct as parameter, expand it, and pass it on to the
		// real function.

		// Cast the passed-in i8* to the struct value (using an alloca) and
		// extract its values.
		alloca := c.builder.CreateAlloca(c.i8ptrType, "receiver.alloca")
		c.builder.CreateStore(wrapper.Param(0), alloca)
		receiverPtr = c.builder.CreateBitCast(alloca, llvm.PointerType(receiverType, 0), "receiver.ptr")
	}

	receiverValue := c.builder.CreateLoad(receiverPtr, "receiver")
	params := append(c.expandFormalParam(receiverValue), wrapper.Params()[1:]...)
	if fn.LLVMFn.Type().ElementType().ReturnType().TypeKind() == llvm.VoidTypeKind {
		c.builder.CreateCall(fn.LLVMFn, params, "")
		c.builder.CreateRetVoid()
	} else {
		ret := c.builder.CreateCall(fn.LLVMFn, params, "ret")
		c.builder.CreateRet(ret)
	}

	return nil
}
