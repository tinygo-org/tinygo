package transform

// This file provides function to lower interface intrinsics to their final LLVM
// form, optimizing them in the process.
//
// During SSA construction, the following pseudo-calls are created:
//     runtime.typeAssert(typecode, assertedType)
//     runtime.interfaceImplements(typecode, interfaceMethodSet)
//     runtime.interfaceMethod(typecode, interfaceMethodSet, signature)
// See src/runtime/interface.go for details.
// These calls are to declared but not defined functions, so the optimizer will
// leave them alone.
//
// This pass lowers the above functions to their final form:
//
// typeAssert:
//     Replaced with an icmp instruction so it can be directly used in a type
//     switch. This is very easy to optimize for LLVM: it will often translate a
//     type switch into a regular switch statement.
//
// interfaceImplements:
//     This call is translated into a call that checks whether the underlying
//     type is one of the types implementing this interface.
//
// interfaceMethod:
//     This call is replaced with a call to a function that calls the
//     appropriate method depending on the underlying type.
//     When there is only one type implementing this interface, this call is
//     translated into a direct call of that method.
//     When there is no type implementing this interface, this code is marked
//     unreachable as there is no way such an interface could be constructed.
//
// Note that this way of implementing interfaces is very different from how the
// main Go compiler implements them. For more details on how the main Go
// compiler does it: https://research.swtch.com/interfaces

import (
	"sort"
	"strings"

	"tinygo.org/x/go-llvm"
)

// signatureInfo is a Go signature of an interface method. It does not represent
// any method in particular.
type signatureInfo struct {
	name       string
	global     llvm.Value
	methods    []*methodInfo
	interfaces []*interfaceInfo
}

// methodName takes a method name like "func String()" and returns only the
// name, which is "String" in this case.
func (s *signatureInfo) methodName() string {
	var methodName string
	if strings.HasPrefix(s.name, "reflect/methods.") {
		methodName = s.name[len("reflect/methods."):]
	} else if idx := strings.LastIndex(s.name, ".$methods."); idx >= 0 {
		methodName = s.name[idx+len(".$methods."):]
	} else {
		panic("could not find method name")
	}
	if openingParen := strings.IndexByte(methodName, '('); openingParen < 0 {
		panic("no opening paren in signature name")
	} else {
		return methodName[:openingParen]
	}
}

// methodInfo describes a single method on a concrete type.
type methodInfo struct {
	*signatureInfo
	function llvm.Value
}

// typeInfo describes a single concrete Go type, which can be a basic or a named
// type. If it is a named type, it may have methods.
type typeInfo struct {
	name      string
	typecode  llvm.Value
	methodSet llvm.Value
	methods   []*methodInfo
}

// getMethod looks up the method on this type with the given signature and
// returns it. The method must exist on this type, otherwise getMethod will
// panic.
func (t *typeInfo) getMethod(signature *signatureInfo) *methodInfo {
	for _, method := range t.methods {
		if method.signatureInfo == signature {
			return method
		}
	}
	panic("could not find method")
}

// interfaceInfo keeps information about a Go interface type, including all
// methods it has.
type interfaceInfo struct {
	name        string                        // name with $interface suffix
	methodSet   llvm.Value                    // global which this interfaceInfo describes
	signatures  []*signatureInfo              // method set
	types       []*typeInfo                   // types this interface implements
	assertFunc  llvm.Value                    // runtime.interfaceImplements replacement
	methodFuncs map[*signatureInfo]llvm.Value // runtime.interfaceMethod replacements for each signature
}

// id removes the $interface suffix from the name and returns the clean
// interface name including import path.
func (itf *interfaceInfo) id() string {
	if !strings.HasSuffix(itf.name, "$interface") {
		panic("interface type does not have $interface suffix: " + itf.name)
	}
	return itf.name[:len(itf.name)-len("$interface")]
}

// lowerInterfacesPass keeps state related to the interface lowering pass. The
// pass has been implemented as an object type because of its complexity, but
// should be seen as a regular function call (see LowerInterfaces).
type lowerInterfacesPass struct {
	mod         llvm.Module
	sizeLevel   int // LLVM optimization size level, 1 means -opt=s and 2 means -opt=z
	builder     llvm.Builder
	ctx         llvm.Context
	uintptrType llvm.Type
	types       map[string]*typeInfo
	signatures  map[string]*signatureInfo
	interfaces  map[string]*interfaceInfo
}

// LowerInterfaces lowers all intermediate interface calls and globals that are
// emitted by the compiler as higher-level intrinsics. They need some lowering
// before LLVM can work on them. This is done so that a few cleanup passes can
// run before assigning the final type codes.
func LowerInterfaces(mod llvm.Module, sizeLevel int) error {
	p := &lowerInterfacesPass{
		mod:         mod,
		sizeLevel:   sizeLevel,
		builder:     mod.Context().NewBuilder(),
		ctx:         mod.Context(),
		uintptrType: mod.Context().IntType(llvm.NewTargetData(mod.DataLayout()).PointerSize() * 8),
		types:       make(map[string]*typeInfo),
		signatures:  make(map[string]*signatureInfo),
		interfaces:  make(map[string]*interfaceInfo),
	}
	return p.run()
}

// run runs the pass itself.
func (p *lowerInterfacesPass) run() error {
	// Collect all type codes.
	for global := p.mod.FirstGlobal(); !global.IsNil(); global = llvm.NextGlobal(global) {
		if strings.HasPrefix(global.Name(), "reflect/types.type:") {
			// Retrieve Go type information based on an opaque global variable.
			// Only the name of the global is relevant, the object itself is
			// discarded afterwards.
			name := strings.TrimPrefix(global.Name(), "reflect/types.type:")
			if _, ok := p.types[name]; !ok {
				t := &typeInfo{
					name:     name,
					typecode: global,
				}
				p.types[name] = t
				initializer := global.Initializer()
				if initializer.IsNil() {
					continue
				}
				methodSet := llvm.ConstExtractValue(initializer, []uint32{2})
				p.addTypeMethods(t, methodSet)
			}
		}
	}

	// Find all interface method calls.
	interfaceMethod := p.mod.NamedFunction("runtime.interfaceMethod")
	interfaceMethodUses := getUses(interfaceMethod)
	for _, use := range interfaceMethodUses {
		methodSet := use.Operand(1).Operand(0)
		name := methodSet.Name()
		if _, ok := p.interfaces[name]; !ok {
			p.addInterface(methodSet)
		}
	}

	// Find all interface type asserts.
	interfaceImplements := p.mod.NamedFunction("runtime.interfaceImplements")
	interfaceImplementsUses := getUses(interfaceImplements)
	for _, use := range interfaceImplementsUses {
		methodSet := use.Operand(1).Operand(0)
		name := methodSet.Name()
		if _, ok := p.interfaces[name]; !ok {
			p.addInterface(methodSet)
		}
	}

	// Find all the interfaces that are implemented per type.
	for _, t := range p.types {
		// This type has no methods, so don't spend time calculating them.
		if len(t.methods) == 0 {
			continue
		}

		// Pre-calculate a set of signatures that this type has, for easy
		// lookup/check.
		typeSignatureSet := make(map[*signatureInfo]struct{})
		for _, method := range t.methods {
			typeSignatureSet[method.signatureInfo] = struct{}{}
		}

		// A set of interfaces, mapped from the name to the info.
		// When the name maps to a nil pointer, one of the methods of this type
		// exists in the given interface but not all of them so this type
		// doesn't implement the interface.
		satisfiesInterfaces := make(map[string]*interfaceInfo)

		for _, method := range t.methods {
			for _, itf := range method.interfaces {
				if _, ok := satisfiesInterfaces[itf.name]; ok {
					// interface already checked with a different method
					continue
				}
				// check whether this interface satisfies this type
				satisfies := true
				for _, itfSignature := range itf.signatures {
					if _, ok := typeSignatureSet[itfSignature]; !ok {
						satisfiesInterfaces[itf.name] = nil // does not satisfy
						satisfies = false
						break
					}
				}
				if !satisfies {
					continue
				}
				satisfiesInterfaces[itf.name] = itf
			}
		}

		// Add this type to all interfaces that satisfy this type.
		for _, itf := range satisfiesInterfaces {
			if itf == nil {
				// Interface does not implement this type, but one of the
				// methods on this type also exists on the interface.
				continue
			}
			itf.types = append(itf.types, t)
		}
	}

	// Sort all types added to the interfaces.
	for _, itf := range p.interfaces {
		sort.Slice(itf.types, func(i, j int) bool {
			return itf.types[i].name > itf.types[j].name
		})
	}

	// Replace all interface methods with their uses, if possible.
	for _, use := range interfaceMethodUses {
		typecode := use.Operand(0)
		signature := p.signatures[use.Operand(2).Name()]

		methodSet := use.Operand(1).Operand(0) // global variable
		itf := p.interfaces[methodSet.Name()]

		// Delegate calling the right function to a special wrapper function.
		inttoptrs := getUses(use)
		if len(inttoptrs) != 1 || inttoptrs[0].IsAIntToPtrInst().IsNil() {
			return errorAt(use, "internal error: expected exactly one inttoptr use of runtime.interfaceMethod")
		}
		inttoptr := inttoptrs[0]
		calls := getUses(inttoptr)
		for _, call := range calls {
			// Set up parameters for the call. First copy the regular params...
			params := make([]llvm.Value, call.OperandsCount())
			paramTypes := make([]llvm.Type, len(params))
			for i := 0; i < len(params)-1; i++ {
				params[i] = call.Operand(i)
				paramTypes[i] = params[i].Type()
			}
			// then add the typecode to the end of the list.
			params[len(params)-1] = typecode
			paramTypes[len(params)-1] = p.uintptrType

			// Create a function that redirects the call to the destination
			// call, after selecting the right concrete type.
			redirector := p.getInterfaceMethodFunc(itf, signature, call.Type(), paramTypes)

			// Replace the old lookup/inttoptr/call with the new call.
			p.builder.SetInsertPointBefore(call)
			retval := p.builder.CreateCall(redirector, append(params, llvm.ConstNull(llvm.PointerType(p.ctx.Int8Type(), 0))), "")
			if retval.Type().TypeKind() != llvm.VoidTypeKind {
				call.ReplaceAllUsesWith(retval)
			}
			call.EraseFromParentAsInstruction()
		}
		inttoptr.EraseFromParentAsInstruction()
		use.EraseFromParentAsInstruction()
	}

	// Replace all typeasserts on interface types with matches on their concrete
	// types, if possible.
	for _, use := range interfaceImplementsUses {
		actualType := use.Operand(0)

		methodSet := use.Operand(1).Operand(0) // global variable
		itf := p.interfaces[methodSet.Name()]
		// Create a function that does a type switch on all available types
		// that implement this interface.
		fn := p.getInterfaceImplementsFunc(itf)
		p.builder.SetInsertPointBefore(use)
		commaOk := p.builder.CreateCall(fn, []llvm.Value{actualType}, "typeassert.ok")
		use.ReplaceAllUsesWith(commaOk)
		use.EraseFromParentAsInstruction()
	}

	// Replace each type assert with an actual type comparison or (if the type
	// assert is impossible) the constant false.
	llvmFalse := llvm.ConstInt(p.ctx.Int1Type(), 0, false)
	for _, use := range getUses(p.mod.NamedFunction("runtime.typeAssert")) {
		actualType := use.Operand(0)
		name := strings.TrimPrefix(use.Operand(1).Name(), "reflect/types.typeid:")
		if t, ok := p.types[name]; ok {
			// The type exists in the program, so lower to a regular integer
			// comparison.
			p.builder.SetInsertPointBefore(use)
			commaOk := p.builder.CreateICmp(llvm.IntEQ, llvm.ConstPtrToInt(t.typecode, p.uintptrType), actualType, "typeassert.ok")
			use.ReplaceAllUsesWith(commaOk)
		} else {
			// The type does not exist in the program, so lower to a constant
			// false. This is trivially further optimized.
			// TODO: eventually it'll be necessary to handle reflect.PtrTo and
			// reflect.New calls which create new types not present in the
			// original program.
			use.ReplaceAllUsesWith(llvmFalse)
		}
		use.EraseFromParentAsInstruction()
	}

	// Remove all method sets, which are now unnecessary and inhibit later
	// optimizations if they are left in place.
	for _, t := range p.types {
		initializer := t.typecode.Initializer()
		methodSet := llvm.ConstExtractValue(initializer, []uint32{2})
		initializer = llvm.ConstInsertValue(initializer, llvm.ConstNull(methodSet.Type()), []uint32{2})
		t.typecode.SetInitializer(initializer)
	}

	return nil
}

// addTypeMethods reads the method set of the given type info struct. It
// retrieves the signatures and the references to the method functions
// themselves for later type<->interface matching.
func (p *lowerInterfacesPass) addTypeMethods(t *typeInfo, methodSet llvm.Value) {
	if !t.methodSet.IsNil() || methodSet.IsNull() {
		// no methods or methods already read
		return
	}
	methodSet = methodSet.Operand(0) // get global from GEP

	// This type has methods, collect all methods of this type.
	t.methodSet = methodSet
	set := methodSet.Initializer() // get value from global
	for i := 0; i < set.Type().ArrayLength(); i++ {
		methodData := llvm.ConstExtractValue(set, []uint32{uint32(i)})
		signatureGlobal := llvm.ConstExtractValue(methodData, []uint32{0})
		signatureName := signatureGlobal.Name()
		function := llvm.ConstExtractValue(methodData, []uint32{1}).Operand(0)
		signature := p.getSignature(signatureName, signatureGlobal)
		method := &methodInfo{
			function:      function,
			signatureInfo: signature,
		}
		signature.methods = append(signature.methods, method)
		t.methods = append(t.methods, method)
	}
}

// addInterface reads information about an interface, which is the
// fully-qualified name and the signatures of all methods it has.
func (p *lowerInterfacesPass) addInterface(methodSet llvm.Value) {
	name := methodSet.Name()
	t := &interfaceInfo{
		name:      name,
		methodSet: methodSet,
	}
	p.interfaces[name] = t
	methodSet = methodSet.Initializer() // get global value from getelementptr
	for i := 0; i < methodSet.Type().ArrayLength(); i++ {
		signatureGlobal := llvm.ConstExtractValue(methodSet, []uint32{uint32(i)})
		signatureName := signatureGlobal.Name()
		signature := p.getSignature(signatureName, signatureGlobal)
		signature.interfaces = append(signature.interfaces, t)
		t.signatures = append(t.signatures, signature)
	}
}

// getSignature returns a new *signatureInfo, creating it if it doesn't already
// exist.
func (p *lowerInterfacesPass) getSignature(name string, global llvm.Value) *signatureInfo {
	if _, ok := p.signatures[name]; !ok {
		p.signatures[name] = &signatureInfo{
			name:   name,
			global: global,
		}
	}
	return p.signatures[name]
}

// getInterfaceImplementsFunc returns a function that checks whether a given
// interface type implements a given interface, by checking all possible types
// that implement this interface.
//
// The type match is implemented using an if/else chain over all possible types.
// This if/else chain is easily converted to a big switch over all possible
// types by the LLVM simplifycfg pass.
func (p *lowerInterfacesPass) getInterfaceImplementsFunc(itf *interfaceInfo) llvm.Value {
	if !itf.assertFunc.IsNil() {
		return itf.assertFunc
	}

	// Create the function and function signature.
	// TODO: debug info
	fnName := itf.id() + "$typeassert"
	fnType := llvm.FunctionType(p.ctx.Int1Type(), []llvm.Type{p.uintptrType}, false)
	fn := llvm.AddFunction(p.mod, fnName, fnType)
	itf.assertFunc = fn
	fn.Param(0).SetName("actualType")
	fn.SetLinkage(llvm.InternalLinkage)
	fn.SetUnnamedAddr(true)
	if p.sizeLevel >= 2 {
		fn.AddFunctionAttr(p.ctx.CreateEnumAttribute(llvm.AttributeKindID("optsize"), 0))
	}

	// Start the if/else chain at the entry block.
	entry := p.ctx.AddBasicBlock(fn, "entry")
	thenBlock := p.ctx.AddBasicBlock(fn, "then")
	p.builder.SetInsertPointAtEnd(entry)

	// Iterate over all possible types.  Each iteration creates a new branch
	// either to the 'then' block (success) or the .next block, for the next
	// check.
	actualType := fn.Param(0)
	for _, typ := range itf.types {
		nextBlock := p.ctx.AddBasicBlock(fn, typ.name+".next")
		cmp := p.builder.CreateICmp(llvm.IntEQ, actualType, llvm.ConstPtrToInt(typ.typecode, p.uintptrType), typ.name+".icmp")
		p.builder.CreateCondBr(cmp, thenBlock, nextBlock)
		p.builder.SetInsertPointAtEnd(nextBlock)
	}

	// The builder is now inserting at the last *.next block.  Once we reach
	// this point, all types have been checked so the type assert will have
	// failed.
	p.builder.CreateRet(llvm.ConstInt(p.ctx.Int1Type(), 0, false))

	// Fill 'then' block (type assert was successful).
	p.builder.SetInsertPointAtEnd(thenBlock)
	p.builder.CreateRet(llvm.ConstInt(p.ctx.Int1Type(), 1, false))

	return itf.assertFunc
}

// getInterfaceMethodFunc returns a thunk for calling a method on an interface.
//
// Matching the actual type is implemented using an if/else chain over all
// possible types.  This is later converted to a switch statement by the LLVM
// simplifycfg pass.
func (p *lowerInterfacesPass) getInterfaceMethodFunc(itf *interfaceInfo, signature *signatureInfo, returnType llvm.Type, paramTypes []llvm.Type) llvm.Value {
	if fn, ok := itf.methodFuncs[signature]; ok {
		// This function has already been created.
		return fn
	}
	if itf.methodFuncs == nil {
		// initialize the above map
		itf.methodFuncs = make(map[*signatureInfo]llvm.Value)
	}

	// Construct the function name, which is of the form:
	//     (main.Stringer).String
	fnName := "(" + itf.id() + ")." + signature.methodName()
	fnType := llvm.FunctionType(returnType, append(paramTypes, llvm.PointerType(p.ctx.Int8Type(), 0)), false)
	fn := llvm.AddFunction(p.mod, fnName, fnType)
	llvm.PrevParam(fn.LastParam()).SetName("actualType")
	fn.LastParam().SetName("parentHandle")
	itf.methodFuncs[signature] = fn
	fn.SetLinkage(llvm.InternalLinkage)
	fn.SetUnnamedAddr(true)
	if p.sizeLevel >= 2 {
		fn.AddFunctionAttr(p.ctx.CreateEnumAttribute(llvm.AttributeKindID("optsize"), 0))
	}

	// TODO: debug info

	// Collect the params that will be passed to the functions to call.
	// These params exclude the receiver (which may actually consist of multiple
	// parts).
	params := make([]llvm.Value, fn.ParamsCount()-3)
	for i := range params {
		params[i] = fn.Param(i + 1)
	}

	// Start chain in the entry block.
	entry := p.ctx.AddBasicBlock(fn, "entry")
	p.builder.SetInsertPointAtEnd(entry)

	// Define all possible functions that can be called.
	actualType := llvm.PrevParam(fn.LastParam())
	for _, typ := range itf.types {
		// Create type check (if/else).
		bb := p.ctx.AddBasicBlock(fn, typ.name)
		next := p.ctx.AddBasicBlock(fn, typ.name+".next")
		cmp := p.builder.CreateICmp(llvm.IntEQ, actualType, llvm.ConstPtrToInt(typ.typecode, p.uintptrType), typ.name+".icmp")
		p.builder.CreateCondBr(cmp, bb, next)

		// The function we will redirect to when the interface has this type.
		function := typ.getMethod(signature).function

		p.builder.SetInsertPointAtEnd(bb)
		receiver := fn.FirstParam()
		if receiver.Type() != function.FirstParam().Type() {
			// When the receiver is a pointer, it is not wrapped. This means the
			// i8* has to be cast to the correct pointer type of the target
			// function.
			receiver = p.builder.CreateBitCast(receiver, function.FirstParam().Type(), "")
		}

		// Check whether the called function has the same signature as would be
		// expected from the parameters. This can happen in rare cases when
		// named struct types are renamed after merging multiple LLVM modules.
		paramTypes := []llvm.Type{receiver.Type()}
		for _, param := range params {
			paramTypes = append(paramTypes, param.Type())
		}
		calledFunctionType := function.Type()
		sig := llvm.PointerType(llvm.FunctionType(calledFunctionType.ElementType().ReturnType(), paramTypes, false), calledFunctionType.PointerAddressSpace())
		if sig != function.Type() {
			function = p.builder.CreateBitCast(function, sig, "")
		}

		retval := p.builder.CreateCall(function, append([]llvm.Value{receiver}, params...), "")
		if retval.Type().TypeKind() == llvm.VoidTypeKind {
			p.builder.CreateRetVoid()
		} else {
			p.builder.CreateRet(retval)
		}

		// Start next comparison in the 'next' block (which is jumped to when
		// the type doesn't match).
		p.builder.SetInsertPointAtEnd(next)
	}

	// The builder now points to the last *.then block, after all types have
	// been checked. Call runtime.nilPanic here.
	// The only other possible value remaining is nil for nil interfaces. We
	// could panic with a different message here such as "nil interface" but
	// that would increase code size and "nil panic" is close enough. Most
	// importantly, it avoids undefined behavior when accidentally calling a
	// method on a nil interface.
	nilPanic := p.mod.NamedFunction("runtime.nilPanic")
	p.builder.CreateCall(nilPanic, []llvm.Value{
		llvm.Undef(llvm.PointerType(p.ctx.Int8Type(), 0)),
		llvm.Undef(llvm.PointerType(p.ctx.Int8Type(), 0)),
	}, "")
	p.builder.CreateUnreachable()

	return fn
}
