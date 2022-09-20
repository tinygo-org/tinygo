package transform

// This file provides function to lower interface intrinsics to their final LLVM
// form, optimizing them in the process.
//
// During SSA construction, the following pseudo-call is created (see
// src/runtime/interface.go):
//     runtime.typeAssert(typecode, assertedType)
// Additionally, interface type asserts and interface invoke functions are
// declared but not defined, so the optimizer will leave them alone.
//
// This pass lowers these functions to their final form:
//
// typeAssert:
//     Replaced with an icmp instruction so it can be directly used in a type
//     switch. This is very easy to optimize for LLVM: it will often translate a
//     type switch into a regular switch statement.
//
// interface type assert:
//     These functions are defined by creating a big type switch over all the
//     concrete types implementing this interface.
//
// interface invoke:
//     These functions are defined with a similar type switch, but instead of
//     checking for the appropriate type, these functions will call the
//     underlying method instead.
//
// Note that this way of implementing interfaces is very different from how the
// main Go compiler implements them. For more details on how the main Go
// compiler does it: https://research.swtch.com/interfaces

import (
	"sort"
	"strings"

	"github.com/tinygo-org/tinygo/compileopts"
	"tinygo.org/x/go-llvm"
)

// signatureInfo is a Go signature of an interface method. It does not represent
// any method in particular.
type signatureInfo struct {
	name       string
	methods    []*methodInfo
	interfaces []*interfaceInfo
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
	name       string                    // "tinygo-methods" attribute
	signatures map[string]*signatureInfo // method set
	types      []*typeInfo               // types this interface implements
}

// lowerInterfacesPass keeps state related to the interface lowering pass. The
// pass has been implemented as an object type because of its complexity, but
// should be seen as a regular function call (see LowerInterfaces).
type lowerInterfacesPass struct {
	mod         llvm.Module
	config      *compileopts.Config
	builder     llvm.Builder
	dibuilder   *llvm.DIBuilder
	difiles     map[string]llvm.Metadata
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
func LowerInterfaces(mod llvm.Module, config *compileopts.Config) error {
	targetData := llvm.NewTargetData(mod.DataLayout())
	defer targetData.Dispose()
	p := &lowerInterfacesPass{
		mod:         mod,
		config:      config,
		builder:     mod.Context().NewBuilder(),
		ctx:         mod.Context(),
		uintptrType: mod.Context().IntType(targetData.PointerSize() * 8),
		types:       make(map[string]*typeInfo),
		signatures:  make(map[string]*signatureInfo),
		interfaces:  make(map[string]*interfaceInfo),
	}
	defer p.builder.Dispose()

	if config.Debug() {
		p.dibuilder = llvm.NewDIBuilder(mod)
		defer p.dibuilder.Destroy()
		defer p.dibuilder.Finalize()
		p.difiles = make(map[string]llvm.Metadata)
	}

	return p.run()
}

// run runs the pass itself.
func (p *lowerInterfacesPass) run() error {
	if p.dibuilder != nil {
		p.dibuilder.CreateCompileUnit(llvm.DICompileUnit{
			Language:  0xb, // DW_LANG_C99 (0xc, off-by-one?)
			File:      "<unknown>",
			Dir:       "",
			Producer:  "TinyGo",
			Optimized: true,
		})
	}

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
				methodSet := p.builder.CreateExtractValue(initializer, 2, "")
				p.addTypeMethods(t, methodSet)
			}
		}
	}

	// Find all interface type asserts and interface method thunks.
	var interfaceAssertFunctions []llvm.Value
	var interfaceInvokeFunctions []llvm.Value
	for fn := p.mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
		methodsAttr := fn.GetStringAttributeAtIndex(-1, "tinygo-methods")
		if methodsAttr.IsNil() {
			continue
		}
		if !hasUses(fn) {
			// Don't bother defining this function.
			continue
		}
		p.addInterface(methodsAttr.GetStringValue())
		invokeAttr := fn.GetStringAttributeAtIndex(-1, "tinygo-invoke")
		if invokeAttr.IsNil() {
			// Type assert.
			interfaceAssertFunctions = append(interfaceAssertFunctions, fn)
		} else {
			// Interface invoke.
			interfaceInvokeFunctions = append(interfaceInvokeFunctions, fn)
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

	// Define all interface invoke thunks.
	for _, fn := range interfaceInvokeFunctions {
		methodsAttr := fn.GetStringAttributeAtIndex(-1, "tinygo-methods")
		invokeAttr := fn.GetStringAttributeAtIndex(-1, "tinygo-invoke")
		itf := p.interfaces[methodsAttr.GetStringValue()]
		signature := itf.signatures[invokeAttr.GetStringValue()]
		p.defineInterfaceMethodFunc(fn, itf, signature)
	}

	// Define all interface type assert functions.
	for _, fn := range interfaceAssertFunctions {
		methodsAttr := fn.GetStringAttributeAtIndex(-1, "tinygo-methods")
		itf := p.interfaces[methodsAttr.GetStringValue()]
		p.defineInterfaceImplementsFunc(fn, itf)
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
	// optimizations if they are left in place. Also remove references to the
	// interface type assert functions just to be sure.
	zeroUintptr := llvm.ConstNull(p.uintptrType)
	for _, t := range p.types {
		initializer := t.typecode.Initializer()
		methodSet := p.builder.CreateExtractValue(initializer, 2, "")
		initializer = p.builder.CreateInsertValue(initializer, llvm.ConstNull(methodSet.Type()), 2, "")
		initializer = p.builder.CreateInsertValue(initializer, zeroUintptr, 4, "")
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
		methodData := p.builder.CreateExtractValue(set, i, "")
		signatureGlobal := p.builder.CreateExtractValue(methodData, 0, "")
		signatureName := signatureGlobal.Name()
		function := p.builder.CreateExtractValue(methodData, 1, "").Operand(0)
		signature := p.getSignature(signatureName)
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
func (p *lowerInterfacesPass) addInterface(methodsString string) {
	if _, ok := p.interfaces[methodsString]; ok {
		return
	}
	t := &interfaceInfo{
		name:       methodsString,
		signatures: make(map[string]*signatureInfo),
	}
	p.interfaces[methodsString] = t
	for _, method := range strings.Split(methodsString, "; ") {
		signature := p.getSignature(method)
		signature.interfaces = append(signature.interfaces, t)
		t.signatures[method] = signature
	}
}

// getSignature returns a new *signatureInfo, creating it if it doesn't already
// exist.
func (p *lowerInterfacesPass) getSignature(name string) *signatureInfo {
	if _, ok := p.signatures[name]; !ok {
		p.signatures[name] = &signatureInfo{
			name: name,
		}
	}
	return p.signatures[name]
}

// defineInterfaceImplementsFunc defines the interface type assert function. It
// checks whether the given interface type (passed as an argument) is one of the
// types it implements.
//
// The type match is implemented using an if/else chain over all possible types.
// This if/else chain is easily converted to a big switch over all possible
// types by the LLVM simplifycfg pass.
func (p *lowerInterfacesPass) defineInterfaceImplementsFunc(fn llvm.Value, itf *interfaceInfo) {
	// Create the function and function signature.
	fn.Param(0).SetName("actualType")
	fn.SetLinkage(llvm.InternalLinkage)
	fn.SetUnnamedAddr(true)
	AddStandardAttributes(fn, p.config)

	// Start the if/else chain at the entry block.
	entry := p.ctx.AddBasicBlock(fn, "entry")
	thenBlock := p.ctx.AddBasicBlock(fn, "then")
	p.builder.SetInsertPointAtEnd(entry)

	if p.dibuilder != nil {
		difile := p.getDIFile("<Go interface assert>")
		diFuncType := p.dibuilder.CreateSubroutineType(llvm.DISubroutineType{
			File: difile,
		})
		difunc := p.dibuilder.CreateFunction(difile, llvm.DIFunction{
			Name:         "(Go interface assert)",
			File:         difile,
			Line:         0,
			Type:         diFuncType,
			LocalToUnit:  true,
			IsDefinition: true,
			ScopeLine:    0,
			Flags:        llvm.FlagPrototyped,
			Optimized:    true,
		})
		fn.SetSubprogram(difunc)
		p.builder.SetCurrentDebugLocation(0, 0, difunc, llvm.Metadata{})
	}

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
}

// defineInterfaceMethodFunc defines this thunk by calling the concrete method
// of the type that implements this interface.
//
// Matching the actual type is implemented using an if/else chain over all
// possible types.  This is later converted to a switch statement by the LLVM
// simplifycfg pass.
func (p *lowerInterfacesPass) defineInterfaceMethodFunc(fn llvm.Value, itf *interfaceInfo, signature *signatureInfo) {
	context := fn.LastParam()
	actualType := llvm.PrevParam(context)
	returnType := fn.Type().ElementType().ReturnType()
	context.SetName("context")
	actualType.SetName("actualType")
	fn.SetLinkage(llvm.InternalLinkage)
	fn.SetUnnamedAddr(true)
	AddStandardAttributes(fn, p.config)

	// Collect the params that will be passed to the functions to call.
	// These params exclude the receiver (which may actually consist of multiple
	// parts).
	params := make([]llvm.Value, fn.ParamsCount()-3)
	for i := range params {
		params[i] = fn.Param(i + 1)
	}
	params = append(params,
		llvm.Undef(llvm.PointerType(p.ctx.Int8Type(), 0)),
	)

	// Start chain in the entry block.
	entry := p.ctx.AddBasicBlock(fn, "entry")
	p.builder.SetInsertPointAtEnd(entry)

	if p.dibuilder != nil {
		difile := p.getDIFile("<Go interface method>")
		diFuncType := p.dibuilder.CreateSubroutineType(llvm.DISubroutineType{
			File: difile,
		})
		difunc := p.dibuilder.CreateFunction(difile, llvm.DIFunction{
			Name:         "(Go interface method)",
			File:         difile,
			Line:         0,
			Type:         diFuncType,
			LocalToUnit:  true,
			IsDefinition: true,
			ScopeLine:    0,
			Flags:        llvm.FlagPrototyped,
			Optimized:    true,
		})
		fn.SetSubprogram(difunc)
		p.builder.SetCurrentDebugLocation(0, 0, difunc, llvm.Metadata{})
	}

	// Define all possible functions that can be called.
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
		functionType := llvm.FunctionType(returnType, paramTypes, false)
		sig := llvm.PointerType(functionType, calledFunctionType.PointerAddressSpace())
		if sig != function.Type() {
			function = p.builder.CreateBitCast(function, sig, "")
		}

		retval := p.builder.CreateCall(functionType, function, append([]llvm.Value{receiver}, params...), "")
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
	p.builder.CreateCall(nilPanic.GlobalValueType(), nilPanic, []llvm.Value{
		llvm.Undef(llvm.PointerType(p.ctx.Int8Type(), 0)),
	}, "")
	p.builder.CreateUnreachable()
}

func (p *lowerInterfacesPass) getDIFile(file string) llvm.Metadata {
	difile, ok := p.difiles[file]
	if !ok {
		difile = p.dibuilder.CreateFile(file, "")
		p.difiles[file] = difile
	}
	return difile
}
