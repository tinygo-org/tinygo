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

	"github.com/tinygo-org/tinygo/compiler/llvmutil"
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
	if !strings.HasPrefix(s.name, "func ") {
		panic("signature must start with \"func \"")
	}
	methodName := s.name[len("func "):]
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
	name             string
	typecode         llvm.Value
	methodSet        llvm.Value
	num              uint64 // the type number after lowering
	countTypeAsserts int    // how often a type assert happens on this method
	methods          []*methodInfo
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

// typeInfoSlice implements sort.Slice, sorting the most commonly used types
// first.
type typeInfoSlice []*typeInfo

func (t typeInfoSlice) Len() int { return len(t) }
func (t typeInfoSlice) Less(i, j int) bool {
	// Try to sort the most commonly used types first.
	if t[i].countTypeAsserts != t[j].countTypeAsserts {
		return t[i].countTypeAsserts < t[j].countTypeAsserts
	}
	return t[i].name < t[j].name
}
func (t typeInfoSlice) Swap(i, j int) { t[i], t[j] = t[j], t[i] }

// interfaceInfo keeps information about a Go interface type, including all
// methods it has.
type interfaceInfo struct {
	name        string                        // name with $interface suffix
	methodSet   llvm.Value                    // global which this interfaceInfo describes
	signatures  []*signatureInfo              // method set
	types       typeInfoSlice                 // types this interface implements
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
	var typecodeIDs []llvm.Value
	for global := p.mod.FirstGlobal(); !global.IsNil(); global = llvm.NextGlobal(global) {
		if strings.HasPrefix(global.Name(), "reflect/types.type:") {
			// Retrieve Go type information based on an opaque global variable.
			// Only the name of the global is relevant, the object itself is
			// discarded afterwards.
			name := strings.TrimPrefix(global.Name(), "reflect/types.type:")
			if _, ok := p.types[name]; !ok {
				typecodeIDs = append(typecodeIDs, global)
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

	// Count per type how often it is type asserted on (e.g. in a switch
	// statement).
	typeAssert := p.mod.NamedFunction("runtime.typeAssert")
	typeAssertUses := getUses(typeAssert)
	for _, use := range typeAssertUses {
		typecode := use.Operand(1)
		name := strings.TrimPrefix(typecode.Name(), "reflect/types.typeid:")
		if t, ok := p.types[name]; ok {
			t.countTypeAsserts++
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

	// Sort all types added to the interfaces, to check for more common types
	// first.
	for _, itf := range p.interfaces {
		sort.Sort(itf.types)
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

	// Make a slice of types sorted by frequency of use.
	typeSlice := make(typeInfoSlice, 0, len(p.types))
	for _, t := range p.types {
		typeSlice = append(typeSlice, t)
	}
	sort.Sort(sort.Reverse(typeSlice))

	// Assign a type code for each type.
	assignTypeCodes(p.mod, typeSlice)

	// Replace each use of a ptrtoint runtime.typecodeID with the constant type
	// code.
	for _, global := range typecodeIDs {
		for _, use := range getUses(global) {
			if use.IsAConstantExpr().IsNil() {
				continue
			}
			t := p.types[strings.TrimPrefix(global.Name(), "reflect/types.type:")]
			typecode := llvm.ConstInt(p.uintptrType, t.num, false)
			switch use.Opcode() {
			case llvm.PtrToInt:
				// Already of the correct type.
			case llvm.BitCast:
				// Could happen when stored in an interface (which is of type
				// i8*).
				typecode = llvm.ConstIntToPtr(typecode, use.Type())
			default:
				panic("unexpected constant expression")
			}
			use.ReplaceAllUsesWith(typecode)
		}
	}

	// Replace each type assert with an actual type comparison or (if the type
	// assert is impossible) the constant false.
	llvmFalse := llvm.ConstInt(p.ctx.Int1Type(), 0, false)
	for _, use := range typeAssertUses {
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

	// Fill in each helper function for type asserts on interfaces
	// (interface-to-interface matches).
	for _, itf := range p.interfaces {
		if !itf.assertFunc.IsNil() {
			p.createInterfaceImplementsFunc(itf)
		}
		for signature := range itf.methodFuncs {
			p.createInterfaceMethodFunc(itf, signature)
		}
	}

	// Replace all ptrtoint typecode placeholders with their final type code
	// numbers.
	for _, typ := range p.types {
		for _, use := range getUses(typ.typecode) {
			if !use.IsAConstantExpr().IsNil() && use.Opcode() == llvm.PtrToInt {
				use.ReplaceAllUsesWith(llvm.ConstInt(p.uintptrType, typ.num, false))
			}
		}
	}

	// Remove most objects created for interface and reflect lowering.
	// Unnecessary, but cleans up the IR for inspection and testing.
	for _, typ := range p.types {
		// Only some typecodes have an initializer.
		initializer := typ.typecode.Initializer()
		if !initializer.IsNil() {
			references := llvm.ConstExtractValue(initializer, []uint32{0})
			typ.typecode.SetInitializer(llvm.ConstNull(initializer.Type()))
			if strings.HasPrefix(typ.name, "reflect/types.type:struct:") {
				// Structs have a 'references' field that is not a typecode but
				// a pointer to a runtime.structField array and therefore a
				// bitcast. This global should be erased separately, otherwise
				// typecode objects cannot be erased.
				structFields := references.Operand(0)
				structFields.EraseFromParentAsGlobal()
			}
		}

		if !typ.methodSet.IsNil() {
			typ.methodSet.EraseFromParentAsGlobal()
			typ.methodSet = llvm.Value{}
		}
	}
	for _, itf := range p.interfaces {
		// Remove method sets of interfaces.
		itf.methodSet.EraseFromParentAsGlobal()
		itf.methodSet = llvm.Value{}
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

// replaceInvokeWithCall replaces a runtime.interfaceMethod + inttoptr with a
// concrete method. This can be done when only one type implements the
// interface.
func (p *lowerInterfacesPass) replaceInvokeWithCall(use llvm.Value, typ *typeInfo, signature *signatureInfo) error {
	inttoptrs := getUses(use)
	if len(inttoptrs) != 1 || inttoptrs[0].IsAIntToPtrInst().IsNil() {
		return errorAt(use, "internal error: expected exactly one inttoptr use of runtime.interfaceMethod")
	}
	inttoptr := inttoptrs[0]
	function := typ.getMethod(signature).function
	if inttoptr.Type() == function.Type() {
		// Easy case: the types are the same. Simply replace the inttoptr
		// result (which is directly called) with the actual function.
		inttoptr.ReplaceAllUsesWith(function)
	} else {
		// Harder case: the type is not actually the same. Go through each call
		// (of which there should be only one), extract the receiver params for
		// this call and replace the call with a direct call to the target
		// function.
		for _, call := range getUses(inttoptr) {
			if call.IsACallInst().IsNil() || call.CalledValue() != inttoptr {
				return errorAt(call, "internal error: expected the inttoptr to be called as a method, this is not a method call")
			}
			operands := make([]llvm.Value, call.OperandsCount()-1)
			for i := range operands {
				operands[i] = call.Operand(i)
			}
			paramTypes := function.Type().ElementType().ParamTypes()
			receiverParamTypes := paramTypes[:len(paramTypes)-(len(operands)-1)]
			methodParamTypes := paramTypes[len(paramTypes)-(len(operands)-1):]
			for i, methodParamType := range methodParamTypes {
				if methodParamType != operands[i+1].Type() {
					return errorAt(call, "internal error: expected method call param type and function param type to be the same")
				}
			}
			p.builder.SetInsertPointBefore(call)
			receiverParams := llvmutil.EmitPointerUnpack(p.builder, p.mod, operands[0], receiverParamTypes)
			result := p.builder.CreateCall(function, append(receiverParams, operands[1:]...), "")
			if result.Type().TypeKind() != llvm.VoidTypeKind {
				call.ReplaceAllUsesWith(result)
			}
			call.EraseFromParentAsInstruction()
		}
	}
	inttoptr.EraseFromParentAsInstruction()
	use.EraseFromParentAsInstruction()
	return nil
}

// getInterfaceImplementsFunc returns a function that checks whether a given
// interface type implements a given interface, by checking all possible types
// that implement this interface.
func (p *lowerInterfacesPass) getInterfaceImplementsFunc(itf *interfaceInfo) llvm.Value {
	if !itf.assertFunc.IsNil() {
		return itf.assertFunc
	}

	// Create the function and function signature.
	// TODO: debug info
	fnName := itf.id() + "$typeassert"
	fnType := llvm.FunctionType(p.ctx.Int1Type(), []llvm.Type{p.uintptrType}, false)
	itf.assertFunc = llvm.AddFunction(p.mod, fnName, fnType)
	itf.assertFunc.Param(0).SetName("actualType")

	// Type asserts will be made for each type, so increment the counter for
	// those.
	for _, typ := range itf.types {
		typ.countTypeAsserts++
	}

	return itf.assertFunc
}

// createInterfaceImplementsFunc finishes the work of
// getInterfaceImplementsFunc, because it needs to run after types have a type
// code assigned.
//
// The type match is implemented using a big type switch over all possible
// types.
func (p *lowerInterfacesPass) createInterfaceImplementsFunc(itf *interfaceInfo) {
	fn := itf.assertFunc
	fn.SetLinkage(llvm.InternalLinkage)
	fn.SetUnnamedAddr(true)
	if p.sizeLevel >= 2 {
		fn.AddFunctionAttr(p.ctx.CreateEnumAttribute(llvm.AttributeKindID("optsize"), 0))
	}

	// TODO: debug info

	// Create all used basic blocks.
	entry := p.ctx.AddBasicBlock(fn, "entry")
	thenBlock := p.ctx.AddBasicBlock(fn, "then")
	elseBlock := p.ctx.AddBasicBlock(fn, "else")

	// Add all possible types as cases.
	p.builder.SetInsertPointAtEnd(entry)
	actualType := fn.Param(0)
	sw := p.builder.CreateSwitch(actualType, elseBlock, len(itf.types))
	for _, typ := range itf.types {
		sw.AddCase(llvm.ConstInt(p.uintptrType, typ.num, false), thenBlock)
	}

	// Fill 'then' block (type assert was successful).
	p.builder.SetInsertPointAtEnd(thenBlock)
	p.builder.CreateRet(llvm.ConstInt(p.ctx.Int1Type(), 1, false))

	// Fill 'else' block (type asserted failed).
	p.builder.SetInsertPointAtEnd(elseBlock)
	p.builder.CreateRet(llvm.ConstInt(p.ctx.Int1Type(), 0, false))
}

// getInterfaceMethodFunc returns a thunk for calling a method on an interface.
// It only declares the function, createInterfaceMethodFunc actually defines the
// function.
func (p *lowerInterfacesPass) getInterfaceMethodFunc(itf *interfaceInfo, signature *signatureInfo, returnType llvm.Type, params []llvm.Type) llvm.Value {
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
	fnType := llvm.FunctionType(returnType, append(params, llvm.PointerType(p.ctx.Int8Type(), 0)), false)
	fn := llvm.AddFunction(p.mod, fnName, fnType)
	llvm.PrevParam(fn.LastParam()).SetName("actualType")
	fn.LastParam().SetName("parentHandle")
	itf.methodFuncs[signature] = fn
	return fn
}

// createInterfaceMethodFunc finishes the work of getInterfaceMethodFunc,
// because it needs to run after type codes have been assigned to concrete
// types.
//
// Matching the actual type is implemented using a big type switch over all
// possible types.
func (p *lowerInterfacesPass) createInterfaceMethodFunc(itf *interfaceInfo, signature *signatureInfo) {
	fn := itf.methodFuncs[signature]
	fn.SetLinkage(llvm.InternalLinkage)
	fn.SetUnnamedAddr(true)
	if p.sizeLevel >= 2 {
		fn.AddFunctionAttr(p.ctx.CreateEnumAttribute(llvm.AttributeKindID("optsize"), 0))
	}

	// TODO: debug info

	// Create entry block.
	entry := p.ctx.AddBasicBlock(fn, "entry")

	// Create default block and call runtime.nilPanic.
	// The only other possible value remaining is nil for nil interfaces. We
	// could panic with a different message here such as "nil interface" but
	// that would increase code size and "nil panic" is close enough. Most
	// importantly, it avoids undefined behavior when accidentally calling a
	// method on a nil interface.
	defaultBlock := p.ctx.AddBasicBlock(fn, "default")
	p.builder.SetInsertPointAtEnd(defaultBlock)
	nilPanic := p.mod.NamedFunction("runtime.nilPanic")
	p.builder.CreateCall(nilPanic, []llvm.Value{
		llvm.Undef(llvm.PointerType(p.ctx.Int8Type(), 0)),
		llvm.Undef(llvm.PointerType(p.ctx.Int8Type(), 0)),
	}, "")
	p.builder.CreateUnreachable()

	// Create type switch in entry block.
	p.builder.SetInsertPointAtEnd(entry)
	actualType := llvm.PrevParam(fn.LastParam())
	sw := p.builder.CreateSwitch(actualType, defaultBlock, len(itf.types))

	// Collect the params that will be passed to the functions to call.
	// These params exclude the receiver (which may actually consist of multiple
	// parts).
	params := make([]llvm.Value, fn.ParamsCount()-3)
	for i := range params {
		params[i] = fn.Param(i + 1)
	}

	// Define all possible functions that can be called.
	for _, typ := range itf.types {
		bb := p.ctx.AddBasicBlock(fn, typ.name)
		sw.AddCase(llvm.ConstInt(p.uintptrType, typ.num, false), bb)

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
	}
}
