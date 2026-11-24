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
//     switch.
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
	"fmt"
	"go/scanner"
	"go/token"
	"slices"
	"sort"
	"strings"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/internal/llvmutil"
	"tinygo.org/x/go-llvm"
)

// numMethodHasMethodSet is a flag in bit 15 of the numMethod field (uint16) in
// Named, Pointer, and Struct type descriptors. When set, an inline method set
// is present in the type descriptor. Must match the constant in
// src/internal/reflectlite/type.go.
const numMethodHasMethodSet = 0x8000

// Meta byte values from src/internal/reflectlite/type.go.
const (
	kindMask  = 31
	flagNamed = 32
	kindFunc  = 24
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
	name            string
	typecode        llvm.Value
	typecodeGEP     llvm.Value
	methodSet       llvm.Value
	methods         []*methodInfo
	callAdapter     llvm.Value
	makeFuncAdapter llvm.Value
	typeID          uint8 // dense ID stored in descriptor byte 1, 0 if none
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
	mod             llvm.Module
	config          *compileopts.Config
	builder         llvm.Builder
	dibuilder       *llvm.DIBuilder
	difiles         map[string]llvm.Metadata
	ctx             llvm.Context
	uintptrType     llvm.Type
	targetData      llvm.TargetData
	ptrType         llvm.Type
	types           map[string]*typeInfo
	signatures      map[string]*signatureInfo
	interfaces      map[string]*interfaceInfo
	keepMethodNames bool
	methodFuncNames map[string]struct{} // nil keeps all exported method functions
}

// LowerInterfaces lowers all intermediate interface calls and globals that are
// emitted by the compiler as higher-level intrinsics. They need some lowering
// before LLVM can work on them. This is done so that a few cleanup passes can
// run before assigning the final type codes.
func LowerInterfaces(mod llvm.Module, config *compileopts.Config) error {
	ctx := mod.Context()
	targetData := llvm.NewTargetData(mod.DataLayout())
	defer targetData.Dispose()
	p := &lowerInterfacesPass{
		mod:         mod,
		config:      config,
		builder:     ctx.NewBuilder(),
		ctx:         ctx,
		targetData:  targetData,
		uintptrType: mod.Context().IntType(targetData.PointerSize() * 8),
		ptrType:     llvm.PointerType(ctx.Int8Type(), 0),
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

	if err := p.pruneUnusedMethodMetadata(); err != nil {
		return err
	}

	// Collect all type codes.
	for global := range llvmutil.Globals(p.mod) {
		if after, ok := strings.CutPrefix(global.Name(), "reflect/types.type:"); ok {
			// Retrieve Go type information based on an opaque global variable.
			// Only the name of the global is relevant, the object itself is
			// discarded afterwards.
			name := after
			if _, ok := p.types[name]; !ok {
				t := &typeInfo{
					name:     name,
					typecode: global,
				}
				p.types[name] = t
				// The names of struct fields also use this prefix, and they
				// are i8 arrays, not structs. Thus, examine the type of the
				// first element, and read its value only from a struct.
				initializer := global.Initializer()
				if elementType(initializer.Type(), 0) != p.ctx.Int8Type() {
					firstField := initializer.AggregateElement(0)
					// Temporary metadata precedes the kind byte at index 1.
					t.typecodeGEP = llvm.ConstGEP(global.GlobalValueType(), global, []llvm.Value{
						llvm.ConstInt(p.ctx.Int32Type(), 0, false),
						llvm.ConstInt(p.ctx.Int32Type(), 1, false),
					})
					if strings.HasPrefix(name, "func:") {
						t.callAdapter = stripPointerCasts(firstField.Operand(0))
						t.makeFuncAdapter = stripPointerCasts(firstField.Operand(1))
					} else {
						methodSet := stripPointerCasts(firstField)
						if !strings.HasSuffix(methodSet.Name(), "$methodset") {
							panic("expected method set")
						}
						p.addTypeMethods(t, methodSet)
					}
				} else {
					// This type has no method set.
					t.typecodeGEP = global
				}
			}
		}
	}

	// Find all interface type asserts and interface method thunks.
	var interfaceAssertFunctions []llvm.Value
	var interfaceInvokeFunctions []llvm.Value
	for fn := range llvmutil.Functions(p.mod) {
		stub := llvmutil.ReadInterfaceStub(fn)
		if stub == nil {
			continue
		}
		if !hasUses(fn) {
			// Don't bother defining this function.
			continue
		}
		p.addInterface(stub.Methods)
		if stub.Invoke == "" {
			// Type assert.
			interfaceAssertFunctions = append(interfaceAssertFunctions, fn)
		} else {
			// Interface invoke.
			interfaceInvokeFunctions = append(interfaceInvokeFunctions, fn)
		}
	}

	// runtime.composeInterface can make an interface value from any typecode,
	// for example the element typecode that reflect reads from a chan type.
	// When it is used, all types are boxable.
	allBoxable := hasUses(p.mod.NamedFunction("runtime.composeInterface"))

	// Find all the interfaces that are implemented per type.
	for _, t := range p.types {
		// This type has no methods, so don't spend time calculating them.
		if len(t.methods) == 0 {
			continue
		}

		// A value of this type can not be in an interface, so no interface
		// method call or type assert can match it.
		if !allBoxable && !isBoxable(t.typecode) {
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

	p.assignTypeIDs(interfaceInvokeFunctions, interfaceAssertFunctions)

	// Report incompatible exported methods before defining any invoke thunks.
	var abiErrors scanner.ErrorList
	seenABIErrors := make(map[llvm.Value]struct{})
	for _, fn := range interfaceInvokeFunctions {
		stub := llvmutil.ReadInterfaceStub(fn)
		itf := p.interfaces[stub.Methods]
		signature := itf.signatures[stub.Invoke]
		for _, typ := range itf.types {
			function := typ.getMethod(signature).function
			if attr := function.GetStringAttributeAtIndex(-1, "tinygo-interface-abi-error"); !attr.IsNil() {
				if _, ok := seenABIErrors[function]; ok {
					continue
				}
				err := errorAt(function, attr.GetStringValue())
				abiErrors = append(abiErrors, &err)
				seenABIErrors[function] = struct{}{}
			}
		}
	}
	if len(abiErrors) != 0 {
		return abiErrors
	}

	// Define all interface invoke thunks.
	for _, fn := range interfaceInvokeFunctions {
		stub := llvmutil.ReadInterfaceStub(fn)
		itf := p.interfaces[stub.Methods]
		signature := itf.signatures[stub.Invoke]
		p.defineInterfaceMethodFunc(fn, itf, signature)
	}

	// Define all interface type assert functions.
	for _, fn := range interfaceAssertFunctions {
		itf := p.interfaces[llvmutil.ReadInterfaceStub(fn).Methods]
		p.defineInterfaceAssertFunc(fn, itf)
	}

	// Replace each type assert with an actual type comparison or (if the type
	// assert is impossible) the constant false.
	llvmFalse := llvm.ConstInt(p.ctx.Int1Type(), 0, false)
	for _, use := range getUses(p.mod.NamedFunction("runtime.typeAssert")) {
		actualType := use.Operand(0)
		name := strings.TrimPrefix(use.Operand(1).Name(), "reflect/types.typeid:")
		gepOffset := uint64(0)
		for strings.HasPrefix(name, "pointer:pointer:") {
			// This is a type like **int, which has the name pointer:pointer:int
			// but is encoded using pointer tagging.
			// Calculate the pointer tag, which is emitted as a GEP instruction.
			name = name[len("pointer:"):]
			gepOffset++
		}

		if t, ok := p.types[name]; ok && (allBoxable || isBoxable(t.typecode)) {
			// Values of this type can occur in an interface.
			p.builder.SetInsertPointBefore(use)
			typecodeGEP := t.typecodeGEP
			if gepOffset != 0 {
				// This is a tagged pointer.
				typecodeGEP = llvm.ConstInBoundsGEP(p.ctx.Int8Type(), typecodeGEP, []llvm.Value{
					llvm.ConstInt(p.ctx.Int64Type(), gepOffset, false),
				})
			}
			commaOk := p.builder.CreateICmp(llvm.IntEQ, typecodeGEP, actualType, "typeassert.ok")
			use.ReplaceAllUsesWith(commaOk)
		} else {
			// This type cannot occur in an interface, so the assertion is false.
			use.ReplaceAllUsesWith(llvmFalse)
		}
		use.EraseFromParentAsInstruction()
	}

	// Create a sorted list of type names, for predictable iteration.
	var typeNames []string
	for name := range p.types {
		typeNames = append(typeNames, name)
	}
	sort.Strings(typeNames)

	p.createReflectTypeLinks(typeNames)
	p.createReflectCallLinks(typeNames)

	// Check whether runtime.typeImplementsMethodSet still has uses. Now that
	// interface type assertions have been lowered to type-ID comparison
	// chains, the only remaining callers would be from reflect
	// (AssignableTo/Implements). If none remain, we can strip the inline
	// method-set data from type descriptors to save binary size.
	stripMethodSets := false
	typeImplementsFn := p.mod.NamedFunction("runtime.typeImplementsMethodSet")
	if !typeImplementsFn.IsNil() && !hasUses(typeImplementsFn) {
		stripMethodSets = true
	}

	usage := analyzeReflection(p.mod, afterInterfaceLowering)
	keepAllMethods := usage.methods
	if keepAllMethods {
		stripMethodSets = false
	}
	p.keepMethodNames = keepAllMethods
	p.methodFuncNames = usage.methodNames

	// Collect all method signatures that appear in any interface type
	// descriptor. When reflect is imported and method sets are kept,
	// concrete type method sets are pruned: individual methods not in any
	// interface are removed, and types that can't fully satisfy at least
	// one interface have their method sets emptied entirely.
	//
	// When method sets are stripped entirely (reflect not imported),
	// methodFilter is nil and filterMethodSet replaces with empty.
	var methodFilter map[string]struct{}
	var ifaceMethodSets []map[string]struct{}
	if !stripMethodSets && !keepAllMethods {
		methodFilter = make(map[string]struct{})
		for _, name := range typeNames {
			if !strings.HasPrefix(name, "interface:") {
				continue
			}
			t := p.types[name]
			initializer := t.typecode.Initializer()
			ifaceSet := make(map[string]struct{})
			for i := 0; i < initializer.Type().StructElementTypesCount(); i++ {
				field := initializer.AggregateElement(i)
				for _, sig := range p.extractMethodSigs(field) {
					methodFilter[sig] = struct{}{}
					ifaceSet[sig] = struct{}{}
				}
			}
			if len(ifaceSet) > 0 {
				ifaceMethodSets = append(ifaceMethodSets, ifaceSet)
			}
		}
	}

	// Remove temporary type metadata, which is now unnecessary and inhibits later
	// optimizations if they are left in place.
	for _, name := range typeNames {
		t := p.types[name]
		if !t.methodSet.IsNil() || !t.callAdapter.IsNil() {
			initializer := t.typecode.Initializer()
			numFields := initializer.Type().StructElementTypesCount()

			// Read numMethods from the original type descriptor (index 2:
			// after prefix pointer at 0 and kind byte at 1). For Named,
			// Pointer, and Struct types, the numMethodHasMethodSet flag
			// indicates that an inline method set is present.
			var numMethodsConst uint64
			var numMethodsIsI16 bool
			if numFields > 2 {
				nmField := initializer.AggregateElement(2)
				if nmField.Type() == p.ctx.Int16Type() {
					numMethodsConst = nmField.ZExtValue()
					numMethodsIsI16 = true
				}
			}

			var newInitializerFields []llvm.Value
			for i := 1; i < numFields; i++ {
				field := initializer.AggregateElement(i)
				if keepAllMethods {
					field = p.filterMethodFunctions(field)
				} else {
					field = p.filterMethodSet(field, methodFilter, ifaceMethodSets)
				}
				// Strip empty inline method sets for Named, Pointer, and
				// Struct types. When the method set is pruned to empty, we
				// remove it and clear the numMethodHasMethodSet flag (bit 15
				// of numMethod) so the runtime skips reading it.
				if numMethodsIsI16 && numMethodsConst&numMethodHasMethodSet != 0 && p.isMethodSetType(field.Type()) {
					elems := field.Type().StructElementTypes()
					if elems[1].ArrayLength() == 0 {
						clearedNumMethods := numMethodsConst & ^uint64(numMethodHasMethodSet)
						newInitializerFields[1] = llvm.ConstInt(p.ctx.Int16Type(), clearedNumMethods, false)
						continue
					}
				}
				newInitializerFields = append(newInitializerFields, field)
			}
			if t.typeID != 0 {
				// The ID goes in the padding byte before numMethods.
				newInitializerFields = append([]llvm.Value{
					newInitializerFields[0],
					llvm.ConstInt(p.ctx.Int8Type(), uint64(t.typeID), false),
				}, newInitializerFields[1:]...)
			}
			newInitializer := p.ctx.ConstStruct(newInitializerFields, false)
			typecodeName := t.typecode.Name()
			newGlobal := llvm.AddGlobal(p.mod, newInitializer.Type(), typecodeName+".tmp")
			newGlobal.SetInitializer(newInitializer)
			newGlobal.SetLinkage(t.typecode.Linkage())
			newGlobal.SetGlobalConstant(true)
			newGlobal.SetAlignment(t.typecode.Alignment())
			for _, use := range getUses(t.typecode) {
				if !use.IsAConstantExpr().IsNil() {
					opcode := use.Opcode()
					if opcode == llvm.GetElementPtr && use.OperandsCount() == 3 {
						if use.Operand(1).ZExtValue() == 0 && use.Operand(2).ZExtValue() == 1 {
							use.ReplaceAllUsesWith(newGlobal)
						}
					}
				}
			}
			// Fallback.
			if hasUses(t.typecode) {
				negativeOffset := -int64(p.targetData.ElementOffset(t.typecode.GlobalValueType(), 1))
				gep := llvm.ConstInBoundsGEP(p.ctx.Int8Type(), newGlobal, []llvm.Value{llvm.ConstInt(p.ctx.Int32Type(), uint64(negativeOffset), true)})
				t.typecode.ReplaceAllUsesWith(gep)
			}
			t.typecode.EraseFromParentAsGlobal()
			newGlobal.SetName(typecodeName)
			t.typecode = newGlobal
		} else {
			initializer := t.typecode.Initializer()
			if initializer.Type().TypeKind() != llvm.StructTypeKind {
				continue
			}
			keepInterfaceMethods := strings.HasPrefix(name, "interface:")
			numFields := initializer.Type().StructElementTypesCount()
			changed := false
			fields := make([]llvm.Value, 0, numFields)
			for i := 0; i < numFields; i++ {
				field := initializer.AggregateElement(i)
				filtered := field
				if !keepInterfaceMethods {
					if keepAllMethods {
						filtered = p.filterMethodFunctions(field)
					} else {
						filtered = p.filterMethodSet(field, methodFilter, ifaceMethodSets)
					}
				}
				if filtered.C != field.C {
					changed = true
				}
				fields = append(fields, filtered)
			}
			if changed {
				t.typecode.SetInitializer(p.ctx.ConstStruct(fields, false))
			}
		}
	}

	p.createReflectMakeFuncLinks(typeNames)

	return nil
}

func (p *lowerInterfacesPass) pruneUnusedMethodMetadata() error {
	if analyzeReflection(p.mod, beforeInterfaceLowering).methods {
		return nil
	}

	changed := false
	for global := range llvmutil.Globals(p.mod) {
		name, ok := strings.CutPrefix(global.Name(), "reflect/types.type:")
		if !ok || strings.HasPrefix(name, "interface:") {
			continue
		}
		initializer := global.Initializer()
		if initializer.Type().TypeKind() != llvm.StructTypeKind {
			continue
		}
		fields := make([]llvm.Value, initializer.Type().StructElementTypesCount())
		globalChanged := false
		for i := range fields {
			field := initializer.AggregateElement(i)
			if p.isMethodSetType(field.Type()) {
				names := field.AggregateElement(2)
				types := field.AggregateElement(3)
				functions := field.AggregateElement(4)
				if !names.IsNull() || !types.IsNull() || !functions.IsNull() {
					field = p.ctx.ConstStruct([]llvm.Value{
						field.AggregateElement(0),
						field.AggregateElement(1),
						llvm.ConstNull(names.Type()),
						llvm.ConstNull(types.Type()),
						llvm.ConstNull(functions.Type()),
					}, false)
					globalChanged = true
				}
			}
			fields[i] = field
		}
		if globalChanged {
			global.SetInitializer(p.ctx.ConstStruct(fields, false))
			changed = true
		}
	}
	if !changed {
		return nil
	}

	// Discard metadata-only types before deciding which interfaces they implement.
	if nilPanic := p.mod.NamedFunction("runtime.nilPanic"); !nilPanic.IsNil() {
		llvmutil.AppendToGlobal(p.mod, "llvm.compiler.used", nilPanic)
		llvmutil.AppendToGlobal(p.mod, "tinygo.interface.metadata.used", nilPanic)
		defer llvmutil.RemoveGlobalReferences(p.mod, "llvm.compiler.used", "tinygo.interface.metadata.used")
	}
	options := llvm.NewPassBuilderOptions()
	defer options.Dispose()
	if err := p.mod.RunPasses("globaldce", llvm.TargetMachine{}, options); err != nil {
		return fmt.Errorf("could not eliminate unused method metadata: %w", err)
	}
	return nil
}

// isBoxable reports whether the program can put a value with this typecode
// into an interface. This is true when the typecode has a use outside other
// typecodes. A typecode that is only the element of a different typecode (for
// example chan T or *T) is not boxable.
func isBoxable(value llvm.Value) bool {
	for use := range llvmutil.Uses(value) {
		if !use.IsAGlobalVariable().IsNil() {
			if !strings.HasPrefix(use.Name(), "reflect/types.type:") {
				return true
			}
			continue
		}
		if !use.IsAConstant().IsNil() {
			if isBoxable(use) {
				return true
			}
			continue
		}
		return true
	}
	return false
}

type reflectionUsage struct {
	methods     bool
	methodNames map[string]struct{} // nil selects all exported methods
	makeFunc    bool
}

type interfaceLoweringStage bool

const (
	beforeInterfaceLowering interfaceLoweringStage = false
	afterInterfaceLowering  interfaceLoweringStage = true
)

func analyzeReflection(mod llvm.Module, stage interfaceLoweringStage) reflectionUsage {
	ctx := mod.Context()
	targetData := llvm.NewTargetData(mod.DataLayout())
	defer targetData.Dispose()
	uintptrType := ctx.IntType(targetData.PointerSize() * 8)
	ptrType := llvm.PointerType(ctx.Int8Type(), 0)
	usage := reflectionUsage{methodNames: make(map[string]struct{})}

	var work []llvm.Value
	seen := make(map[llvm.Value]struct{})
	enqueue := func(v llvm.Value) {
		if v.IsNil() || v.IsNull() || v.IsAConstant().IsNil() {
			return
		}
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			work = append(work, v)
		}
	}
	addRoot := func(v llvm.Value) {
		switch v.Linkage() {
		case llvm.InternalLinkage, llvm.PrivateLinkage, llvm.LinkOnceAnyLinkage, llvm.LinkOnceODRLinkage, llvm.AvailableExternallyLinkage:
		default:
			enqueue(v)
		}
	}
	// Aliases and ifuncs use a global or a constant expression of one.
	indirectSymbols := make(map[llvm.Value]struct{})
	var addIndirectRoots func(llvm.Value)
	addIndirectRoots = func(v llvm.Value) {
		for use := v.FirstUse(); !use.IsNil(); use = use.NextUse() {
			user := use.User()
			switch {
			case !user.IsAConstantExpr().IsNil():
			case !user.IsAGlobalValue().IsNil() && user.IsAFunction().IsNil() && user.IsAGlobalVariable().IsNil():
				addRoot(user)
			default:
				continue
			}
			if _, ok := indirectSymbols[user]; !ok {
				indirectSymbols[user] = struct{}{}
				addIndirectRoots(user)
			}
		}
	}
	for fn := range llvmutil.Functions(mod) {
		if !fn.IsDeclaration() {
			addRoot(fn)
		}
		addIndirectRoots(fn)
	}
	for global := range llvmutil.Globals(mod) {
		addRoot(global)
		addIndirectRoots(global)
	}

	// Activate callable metadata edges as reachable lookups select their names.
	pending := make(map[string][]llvm.Value)
	for len(work) != 0 {
		v := work[len(work)-1]
		work = work[:len(work)-1]
		switch {
		case !v.IsAFunction().IsNil():
			attrs := llvmutil.ReadFunctionUsageAttributes(v)
			if _, ok := attrs["tinygo-reflect-makefunc"]; ok || v.Name() == "reflect.MakeFunc" {
				usage.makeFunc = true
			}
			if _, ok := attrs["tinygo-reflect-method"]; ok {
				usage.methods = true
				usage.methodNames = nil
				for name, functions := range pending {
					for _, fn := range functions {
						enqueue(fn)
					}
					delete(pending, name)
				}
			}
			if value, ok := attrs["tinygo-reflect-method-names"]; ok {
				usage.methods = true
				if usage.methodNames != nil {
					for _, name := range strings.Fields(value) {
						usage.methodNames[name] = struct{}{}
						for _, fn := range pending[name] {
							enqueue(fn)
						}
						delete(pending, name)
					}
				}
			}
			for bb := range llvmutil.BasicBlocks(v) {
				for inst := range llvmutil.Instructions(bb) {
					for i := range inst.OperandsCount() {
						enqueue(inst.Operand(i))
					}
				}
			}
			continue
		case !v.IsAGlobalVariable().IsNil():
			name := v.Name()
			if stage == afterInterfaceLowering && strings.HasSuffix(name, "$methodset") {
				// Interface dispatch edges are already present in the lowered thunks.
				continue
			}
			init := v.Initializer()
			if !strings.HasPrefix(name, "reflect/types.type:") || init.IsNil() || init.Type().TypeKind() != llvm.StructTypeKind {
				enqueue(init)
				continue
			}
			for i := range init.Type().StructElementTypesCount() {
				field := init.AggregateElement(i)
				if !isMethodSetType(field.Type(), uintptrType, ptrType) {
					enqueue(field)
					continue
				}
				for j := range 4 {
					enqueue(field.AggregateElement(j))
				}
				signatures := field.AggregateElement(1)
				functions := field.AggregateElement(4)
				for j := range functions.Type().ArrayLength() {
					fn := functions.AggregateElement(j)
					if fn.IsNull() {
						continue
					}
					name := reflectedMethodName(signatures.AggregateElement(j))
					if !token.IsExported(name) || strings.Contains(name, ".") {
						continue
					}
					if _, ok := usage.methodNames[name]; usage.methodNames == nil || ok {
						enqueue(fn)
					} else {
						pending[name] = append(pending[name], fn)
					}
				}
			}
			continue
		}
		for i := range v.OperandsCount() {
			enqueue(v.Operand(i))
		}
	}
	return usage
}

func (p *lowerInterfacesPass) createReflectTypeLinks(typeNames []string) {
	for _, spec := range []struct {
		prefix string
		data   string
		length string
	}{
		{"slice:", "internal/reflectlite.sliceTypeLinks", "internal/reflectlite.sliceTypeLinksLen"},
		{"map:", "internal/reflectlite.mapTypeLinks", "internal/reflectlite.mapTypeLinksLen"},
		{"array:", "internal/reflectlite.arrayTypeLinks", "internal/reflectlite.arrayTypeLinksLen"},
		{"chan:", "internal/reflectlite.chanTypeLinks", "internal/reflectlite.chanTypeLinksLen"},
		{"struct:", "internal/reflectlite.structTypeLinks", "internal/reflectlite.structTypeLinksLen"},
		{strings.Repeat("pointer:", compileopts.MaxTypePointerDepth), "internal/reflectlite.pointerTypeLinks", "internal/reflectlite.pointerTypeLinksLen"},
		{"func:", "internal/reflectlite.funcTypeLinks", "internal/reflectlite.funcTypeLinksLen"},
	} {
		dataGlobal := p.mod.NamedGlobal(spec.data)
		lengthGlobal := p.mod.NamedGlobal(spec.length)
		if dataGlobal.IsNil() && lengthGlobal.IsNil() {
			continue
		}
		if dataGlobal.IsNil() || lengthGlobal.IsNil() {
			panic("reflect type link globals must be defined together")
		}

		var typeCodes []llvm.Value
		for _, name := range typeNames {
			if !strings.HasPrefix(name, spec.prefix) {
				continue
			}
			typeCodes = append(typeCodes, p.types[name].typecodeGEP)
		}

		lengthGlobal.SetInitializer(llvm.ConstInt(p.uintptrType, uint64(len(typeCodes)), false))
		lengthGlobal.SetGlobalConstant(true)
		if len(typeCodes) == 0 {
			dataGlobal.SetInitializer(llvm.ConstPointerNull(p.ptrType))
			dataGlobal.SetGlobalConstant(true)
			continue
		}

		arrayType := llvm.ArrayType(p.ptrType, len(typeCodes))
		array := llvm.AddGlobal(p.mod, arrayType, spec.data+".data")
		array.SetInitializer(llvm.ConstArray(p.ptrType, typeCodes))
		array.SetGlobalConstant(true)
		array.SetLinkage(llvm.InternalLinkage)
		dataGlobal.SetInitializer(array)
		dataGlobal.SetGlobalConstant(true)
	}
}

func (p *lowerInterfacesPass) createReflectCallLinks(typeNames []string) {
	typesGlobal := p.mod.NamedGlobal("internal/reflectlite.funcCallTypes")
	adaptersGlobal := p.mod.NamedGlobal("internal/reflectlite.funcCallAdapters")
	lengthGlobal := p.mod.NamedGlobal("internal/reflectlite.funcCallLinksLen")
	if typesGlobal.IsNil() && adaptersGlobal.IsNil() && lengthGlobal.IsNil() {
		p.internalizeFunctionsWithPrefix("reflect/call:")
		return
	}
	if typesGlobal.IsNil() || adaptersGlobal.IsNil() || lengthGlobal.IsNil() {
		panic("reflect call link globals must be defined together")
	}

	var typeCodes, adapters []llvm.Value
	for _, name := range typeNames {
		if !strings.HasPrefix(name, "func:") {
			continue
		}
		typ := p.types[name]
		if typ.callAdapter.IsNil() {
			panic("missing reflect call adapter for " + name)
		}
		typeCodes = append(typeCodes, typ.typecodeGEP)
		adapters = append(adapters, llvm.ConstIntToPtr(typ.callAdapter, p.ptrType))
	}
	p.internalizeFunctionsWithPrefix("reflect/call:")

	lengthGlobal.SetInitializer(llvm.ConstInt(p.uintptrType, uint64(len(typeCodes)), false))
	lengthGlobal.SetGlobalConstant(true)
	for _, table := range []struct {
		global llvm.Value
		values []llvm.Value
		name   string
	}{
		{typesGlobal, typeCodes, "internal/reflectlite.funcCallTypes.data"},
		{adaptersGlobal, adapters, "internal/reflectlite.funcCallAdapters.data"},
	} {
		if len(table.values) == 0 {
			table.global.SetInitializer(llvm.ConstPointerNull(p.ptrType))
			table.global.SetGlobalConstant(true)
			continue
		}
		arrayType := llvm.ArrayType(p.ptrType, len(table.values))
		array := llvm.AddGlobal(p.mod, arrayType, table.name)
		array.SetInitializer(llvm.ConstArray(p.ptrType, table.values))
		array.SetGlobalConstant(true)
		array.SetLinkage(llvm.InternalLinkage)
		table.global.SetInitializer(array)
		table.global.SetGlobalConstant(true)
	}
}

func pruneUnusedReflectMakeFunc(mod llvm.Module) {
	p := lowerInterfacesPass{mod: mod}
	if analyzeReflection(mod, beforeInterfaceLowering).makeFunc {
		return
	}

	for global := mod.FirstGlobal(); !global.IsNil(); global = llvm.NextGlobal(global) {
		if !strings.HasPrefix(global.Name(), "reflect/types.type:func:") {
			continue
		}
		initializer := global.Initializer()
		adapters := initializer.Operand(0)
		fields := []llvm.Value{llvm.ConstArray(adapters.Type().ElementType(), []llvm.Value{
			adapters.Operand(0), llvm.ConstNull(adapters.Operand(1).Type()),
		})}
		for i := 1; i < initializer.OperandsCount(); i++ {
			fields = append(fields, initializer.Operand(i))
		}
		global.SetInitializer(mod.Context().ConstStruct(fields, false))
	}
	p.internalizeFunctionsWithPrefix("reflect/makefunc:")
}

func (p *lowerInterfacesPass) createReflectMakeFuncLinks(typeNames []string) {
	used := analyzeReflection(p.mod, afterInterfaceLowering).makeFunc

	typesGlobal := p.mod.NamedGlobal("internal/reflectlite.makeFuncTypes")
	adaptersGlobal := p.mod.NamedGlobal("internal/reflectlite.makeFuncAdapters")
	lengthGlobal := p.mod.NamedGlobal("internal/reflectlite.makeFuncLinksLen")
	if typesGlobal.IsNil() && adaptersGlobal.IsNil() && lengthGlobal.IsNil() {
		p.internalizeFunctionsWithPrefix("reflect/makefunc:")
		return
	}
	if typesGlobal.IsNil() || adaptersGlobal.IsNil() || lengthGlobal.IsNil() {
		panic("reflect MakeFunc link globals must be defined together")
	}
	if !used {
		p.internalizeFunctionsWithPrefix("reflect/makefunc:")
		lengthGlobal.SetInitializer(llvm.ConstInt(p.uintptrType, 0, false))
		lengthGlobal.SetGlobalConstant(true)
		typesGlobal.SetInitializer(llvm.ConstPointerNull(p.ptrType))
		typesGlobal.SetGlobalConstant(true)
		adaptersGlobal.SetInitializer(llvm.ConstPointerNull(p.ptrType))
		adaptersGlobal.SetGlobalConstant(true)
		return
	}

	var typeCodes, adapters []llvm.Value
	for _, name := range typeNames {
		if !strings.HasPrefix(name, "func:") {
			continue
		}
		typ := p.types[name]
		if typ.makeFuncAdapter.IsNil() || typ.makeFuncAdapter.IsNull() {
			panic("missing reflect MakeFunc adapter for " + name)
		}
		typeCodes = append(typeCodes, typ.typecode)
		adapters = append(adapters, llvm.ConstIntToPtr(typ.makeFuncAdapter, p.ptrType))
	}
	p.internalizeFunctionsWithPrefix("reflect/makefunc:")

	lengthGlobal.SetInitializer(llvm.ConstInt(p.uintptrType, uint64(len(typeCodes)), false))
	lengthGlobal.SetGlobalConstant(true)
	for _, table := range []struct {
		global llvm.Value
		values []llvm.Value
		name   string
	}{
		{typesGlobal, typeCodes, "internal/reflectlite.makeFuncTypes.data"},
		{adaptersGlobal, adapters, "internal/reflectlite.makeFuncAdapters.data"},
	} {
		if len(table.values) == 0 {
			table.global.SetInitializer(llvm.ConstPointerNull(p.ptrType))
			table.global.SetGlobalConstant(true)
			continue
		}
		arrayType := llvm.ArrayType(p.ptrType, len(table.values))
		array := llvm.AddGlobal(p.mod, arrayType, table.name)
		array.SetInitializer(llvm.ConstArray(p.ptrType, table.values))
		array.SetGlobalConstant(true)
		array.SetLinkage(llvm.InternalLinkage)
		table.global.SetInitializer(array)
		table.global.SetGlobalConstant(true)
	}
}

func (p *lowerInterfacesPass) internalizeFunctionsWithPrefix(prefix string) {
	for fn := p.mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
		if strings.HasPrefix(fn.Name(), prefix) {
			fn.SetLinkage(llvm.InternalLinkage)
		}
	}
}

// addTypeMethods reads the method set of the given type info struct. It
// retrieves the signatures and the references to the method functions
// themselves for later type<->interface matching.
func (p *lowerInterfacesPass) addTypeMethods(t *typeInfo, methodSet llvm.Value) {
	if !t.methodSet.IsNil() {
		// no methods or methods already read
		return
	}

	// This type has methods, collect all methods of this type.
	t.methodSet = methodSet
	set := methodSet.Initializer() // get value from global
	signatures := set.AggregateElement(1)
	wrappers := set.AggregateElement(2)
	numMethods := signatures.Type().ArrayLength()
	for i := range numMethods {
		signatureGlobal := signatures.AggregateElement(i)
		function := wrappers.AggregateElement(i)
		function = stripPointerCasts(function) // strip bitcasts
		signatureName := signatureGlobal.Name()
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
	for method := range strings.SplitSeq(methodsString, "; ") {
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

// defineInterfaceMethodFunc defines this thunk by calling the concrete method
// of the type that implements this interface.
//
// Above typeIDSwitchThreshold implementers it switches on the type ID, and
// otherwise compares the typecode against each type.
func (p *lowerInterfacesPass) defineInterfaceMethodFunc(fn llvm.Value, itf *interfaceInfo, signature *signatureInfo) {
	context := fn.LastParam()
	actualType := llvm.PrevParam(context)
	returnType := fn.GlobalValueType().ReturnType()
	resultOffset := 0
	if llvmutil.ReadInterfaceStub(fn).IndirectResult {
		resultOffset = 1
	}
	context.SetName("context")
	actualType.SetName("actualType")
	fn.SetLinkage(llvm.InternalLinkage)
	fn.SetUnnamedAddr(true)
	AddStandardAttributes(fn, p.config)

	// Collect the params that will be passed to the functions to call.
	// These params exclude the receiver (which may actually consist of multiple
	// parts).
	params := make([]llvm.Value, fn.ParamsCount()-3-resultOffset)
	for i := range params {
		params[i] = fn.Param(i + 1 + resultOffset)
	}
	params = append(params,
		llvm.Undef(p.ptrType),
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

	// Reached for a nil interface or a type with no match.
	panicBlock := p.ctx.AddBasicBlock(fn, "nilpanic")

	receiver := fn.Param(resultOffset)
	callBlock := func(typ *typeInfo) llvm.BasicBlock {
		bb := p.ctx.AddBasicBlock(fn, typ.name)
		saved := p.builder.GetInsertBlock()
		p.builder.SetInsertPointAtEnd(bb)
		function := typ.getMethod(signature).function
		callParams := make([]llvm.Value, 0, len(params)+2+resultOffset)
		if resultOffset != 0 {
			callParams = append(callParams, fn.FirstParam())
		}
		callParams = append(callParams, receiver)
		callParams = append(callParams, params...)
		paramTypes := make([]llvm.Type, len(callParams))
		for i, param := range callParams {
			paramTypes[i] = param.Type()
		}
		functionType := llvm.FunctionType(returnType, paramTypes, false)
		retval := p.builder.CreateCall(functionType, function, callParams, "")
		if retval.Type().TypeKind() == llvm.VoidTypeKind {
			p.builder.CreateRetVoid()
		} else {
			p.builder.CreateRet(retval)
		}
		p.builder.SetInsertPointAtEnd(saved)
		return bb
	}

	chain := p.createTypeIDSwitch(fn, actualType, itf.types, panicBlock, false, callBlock)
	for _, typ := range chain {
		next := p.ctx.AddBasicBlock(fn, typ.name+".next")
		cmp := p.builder.CreateICmp(llvm.IntEQ, actualType, typ.typecodeGEP, typ.name+".icmp")
		p.builder.CreateCondBr(cmp, callBlock(typ), next)
		p.builder.SetInsertPointAtEnd(next)
	}
	p.builder.CreateBr(panicBlock)

	// The only other possible value remaining is nil for nil interfaces. We
	// could panic with a different message here such as "nil interface" but
	// that would increase code size and "nil panic" is close enough. Most
	// importantly, it avoids undefined behavior when accidentally calling a
	// method on a nil interface.
	p.builder.SetInsertPointAtEnd(panicBlock)
	nilPanic := p.mod.NamedFunction("runtime.nilPanic")
	p.builder.CreateCall(nilPanic.GlobalValueType(), nilPanic, []llvm.Value{
		llvm.Undef(p.ptrType),
	}, "")
	p.builder.CreateUnreachable()
}

// typeIDSwitchThreshold is the number of implementers above which a thunk
// switches on the type ID instead of comparing typecodes.
const typeIDSwitchThreshold = 4

// assignTypeIDs gives IDs 1 to 255 to implementers of invoked or asserted
// interfaces above typeIDSwitchThreshold. Larger interfaces go first.
func (p *lowerInterfacesPass) assignTypeIDs(invokeFunctions, assertFunctions []llvm.Value) {
	seen := make(map[*interfaceInfo]struct{})
	var itfs []*interfaceInfo
	for _, fn := range slices.Concat(invokeFunctions, assertFunctions) {
		itf := p.interfaces[llvmutil.ReadInterfaceStub(fn).Methods]
		if _, ok := seen[itf]; ok || len(itf.types) <= typeIDSwitchThreshold {
			continue
		}
		seen[itf] = struct{}{}
		itfs = append(itfs, itf)
	}
	sort.Slice(itfs, func(i, j int) bool {
		if len(itfs[i].types) != len(itfs[j].types) {
			return len(itfs[i].types) > len(itfs[j].types)
		}
		return itfs[i].name < itfs[j].name
	})
	next := 1
	for _, itf := range itfs {
		for _, t := range itf.types {
			if t.typeID != 0 || next > 255 || !p.hasTypeIDByte(t) {
				continue
			}
			t.typeID = uint8(next)
			next++
		}
	}
}

// createTypeIDSwitch switches on the type ID when types is above the threshold,
// and returns the types that still need a typecode compare.
func (p *lowerInterfacesPass) createTypeIDSwitch(fn, actualType llvm.Value, types []*typeInfo, failBlock llvm.BasicBlock, anyType bool, caseBlock func(*typeInfo) llvm.BasicBlock) []*typeInfo {
	if len(types) <= typeIDSwitchThreshold {
		return types
	}
	var switched, chain []*typeInfo
	for _, typ := range types {
		if typ.typeID != 0 {
			switched = append(switched, typ)
		} else {
			chain = append(chain, typ)
		}
	}
	if len(switched) == 0 {
		return chain
	}
	bad := p.builder.CreateICmp(llvm.IntEQ, actualType, llvm.ConstNull(actualType.Type()), "isnil")
	if anyType {
		// Pointer to pointer typecodes are tagged in the low 2 bits, see ptrtag
		// in src/internal/reflectlite/type.go.
		tag := p.builder.CreateAnd(p.builder.CreatePtrToInt(actualType, p.uintptrType, ""), llvm.ConstInt(p.uintptrType, 3, false), "tag")
		isTagged := p.builder.CreateICmp(llvm.IntNE, tag, llvm.ConstInt(p.uintptrType, 0, false), "istagged")
		bad = p.builder.CreateOr(bad, isTagged, "")
	}
	idBlock := p.ctx.AddBasicBlock(fn, "typeid")
	p.builder.CreateCondBr(bad, failBlock, idBlock)
	p.builder.SetInsertPointAtEnd(idBlock)
	idPtr := p.builder.CreateInBoundsGEP(p.ctx.Int8Type(), actualType, []llvm.Value{
		llvm.ConstInt(p.ctx.Int32Type(), 1, false),
	}, "typeid.ptr")
	id := p.builder.CreateLoad(p.ctx.Int8Type(), idPtr, "typeid")
	id.SetAlignment(1)
	defaultBlock := p.ctx.AddBasicBlock(fn, "typeid.default")
	sw := p.builder.CreateSwitch(id, defaultBlock, len(switched))
	for _, typ := range switched {
		sw.AddCase(llvm.ConstInt(p.ctx.Int8Type(), uint64(typ.typeID), false), caseBlock(typ))
	}
	p.builder.SetInsertPointAtEnd(defaultBlock)
	return chain
}

// hasTypeIDByte reports whether byte 1 of the type descriptor is padding before
// an aligned numMethods field. Index 0 is still the temporary method set.
func (p *lowerInterfacesPass) hasTypeIDByte(t *typeInfo) bool {
	if t.methodSet.IsNil() {
		return false
	}
	init := t.typecode.Initializer()
	if init.Type().StructElementTypesCount() < 3 {
		return false
	}
	elems := init.Type().StructElementTypes()
	if elems[1] != p.ctx.Int8Type() || elems[2] != p.ctx.Int16Type() {
		return false
	}
	// On AVR numMethods is not aligned, so there is no padding byte.
	ty := init.Type()
	return p.targetData.ElementOffset(ty, 2)-p.targetData.ElementOffset(ty, 1) >= 2
}

func (p *lowerInterfacesPass) getDIFile(file string) llvm.Metadata {
	difile, ok := p.difiles[file]
	if !ok {
		difile = p.dibuilder.CreateFile(file, "")
		p.difiles[file] = difile
	}
	return difile
}

// defineInterfaceAssertFunc defines a $typeassert function for the given
// interface. The function returns true if the concrete type (passed as a
// type-ID pointer) implements the interface, using a switch on the type ID or
// a chain of type-ID comparisons. This avoids pulling in
// runtime.typeImplementsMethodSet for programs that don't use reflect.
func (p *lowerInterfacesPass) defineInterfaceAssertFunc(fn llvm.Value, itf *interfaceInfo) {
	actualType := fn.FirstParam()
	actualType.SetName("actualType")
	fn.SetLinkage(llvm.InternalLinkage)
	fn.SetUnnamedAddr(true)
	AddStandardAttributes(fn, p.config)

	entry := p.ctx.AddBasicBlock(fn, "entry")
	p.builder.SetInsertPointAtEnd(entry)

	if p.dibuilder != nil {
		difile := p.getDIFile("<Go interface type assert>")
		diFuncType := p.dibuilder.CreateSubroutineType(llvm.DISubroutineType{
			File: difile,
		})
		difunc := p.dibuilder.CreateFunction(difile, llvm.DIFunction{
			Name:         "(Go interface type assert)",
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

	llvmFalse := llvm.ConstInt(p.ctx.Int1Type(), 0, false)
	falseBlock := p.ctx.AddBasicBlock(fn, "false")
	trueBlock := p.ctx.AddBasicBlock(fn, "true")
	chain := p.createTypeIDSwitch(fn, actualType, itf.types, falseBlock, true, func(*typeInfo) llvm.BasicBlock {
		return trueBlock
	})

	// Build an OR chain: return (type == T1) || (type == T2) || ...
	result := llvmFalse
	for _, typ := range chain {
		cmp := p.builder.CreateICmp(llvm.IntEQ, actualType, typ.typecodeGEP, typ.name+".icmp")
		result = p.builder.CreateOr(result, cmp, "")
	}
	p.builder.CreateRet(result)

	if !hasUses(trueBlock.AsValue()) {
		falseBlock.EraseFromParent()
		trueBlock.EraseFromParent()
		return
	}
	p.builder.SetInsertPointAtEnd(falseBlock)
	p.builder.CreateRet(llvmFalse)

	// Byte 1 of an unnamed func type is numIn, not a type ID.
	p.builder.SetInsertPointAtEnd(trueBlock)
	meta := p.builder.CreateLoad(p.ctx.Int8Type(), actualType, "meta")
	kind := p.builder.CreateAnd(meta, llvm.ConstInt(p.ctx.Int8Type(), kindMask|flagNamed, false), "")
	p.builder.CreateRet(p.builder.CreateICmp(llvm.IntNE, kind, llvm.ConstInt(p.ctx.Int8Type(), kindFunc, false), ""))
}

func (p *lowerInterfacesPass) isMethodSetType(ty llvm.Type) bool {
	return isMethodSetType(ty, p.uintptrType, p.ptrType)
}

// isMethodSetType reports whether ty has the shape of a method-set struct:
// { uintptr, [N x ptr], [N x ptr], [N x ptr], [N x uintptr] }.
func isMethodSetType(ty, uintptrType, ptrType llvm.Type) bool {
	if ty.TypeKind() != llvm.StructTypeKind {
		return false
	}
	elems := ty.StructElementTypes()
	if len(elems) != 5 {
		return false
	}
	if elems[0] != uintptrType || elems[1].TypeKind() != llvm.ArrayTypeKind {
		return false
	}
	length := elems[1].ArrayLength()
	for _, elem := range elems[1:4] {
		if elem.TypeKind() != llvm.ArrayTypeKind ||
			elem.ElementType() != ptrType ||
			elem.ArrayLength() != length {
			return false
		}
	}
	if elems[4].TypeKind() != llvm.ArrayTypeKind ||
		elems[4].ElementType() != uintptrType ||
		elems[4].ArrayLength() != length {
		return false
	}
	return true
}

// extractMethodSigs returns the names of method signature globals inside a
// method-set field. Returns nil if field is not a method set.
func (p *lowerInterfacesPass) extractMethodSigs(field llvm.Value) []string {
	if !p.isMethodSetType(field.Type()) {
		return nil
	}
	methodArray := field.AggregateElement(1)
	n := methodArray.Type().ArrayLength()
	sigs := make([]string, 0, n)
	for j := range n {
		sig := methodArray.AggregateElement(j)
		sig = stripPointerCasts(sig)
		sigs = append(sigs, sig.Name())
	}
	return sigs
}

func (p *lowerInterfacesPass) filterMethodFunctions(field llvm.Value) llvm.Value {
	if !p.isMethodSetType(field.Type()) {
		return field
	}
	functionArray := p.builder.CreateExtractValue(field, 4, "")
	signatureArray := p.builder.CreateExtractValue(field, 1, "")
	functions := make([]llvm.Value, functionArray.Type().ArrayLength())
	for i := range functions {
		name := reflectedMethodName(signatureArray.AggregateElement(i))
		function := llvm.ConstInt(p.uintptrType, 0, false)
		if token.IsExported(name) && !strings.Contains(name, ".") {
			_, named := p.methodFuncNames[name]
			if p.methodFuncNames == nil || named {
				function = p.builder.CreateExtractValue(functionArray, i, "")
			}
		}
		functions[i] = function
	}
	return p.ctx.ConstStruct([]llvm.Value{
		p.builder.CreateExtractValue(field, 0, ""),
		p.builder.CreateExtractValue(field, 1, ""),
		p.builder.CreateExtractValue(field, 2, ""),
		p.builder.CreateExtractValue(field, 3, ""),
		llvm.ConstArray(p.uintptrType, functions),
	}, false)
}

func reflectedMethodName(signature llvm.Value) string {
	name, _, _ := strings.Cut(strings.TrimPrefix(stripPointerCasts(signature).Name(), "reflect/types.signature:"), ":")
	return name
}

// filterMethodSet processes a type-descriptor field that may be a method set.
// Non-method-set fields are returned unchanged.
//
// If keepSigs is nil, the method set is replaced with an empty one (strip mode,
// used when reflect is not imported). If keepSigs is non-nil, the method set is
// pruned in two stages: first, methods not in keepSigs (the union of all
// interface signatures) are removed; then, if the remaining methods cannot
// fully satisfy at least one interface in ifaceSets, the entire method set is
// emptied.
func (p *lowerInterfacesPass) filterMethodSet(field llvm.Value, keepSigs map[string]struct{}, ifaceSets []map[string]struct{}) llvm.Value {
	if !p.isMethodSetType(field.Type()) {
		return field
	}

	methodArray := field.AggregateElement(1)
	nameArray := field.AggregateElement(2)
	typeArray := field.AggregateElement(3)
	functionArray := field.AggregateElement(4)
	numMethods := methodArray.Type().ArrayLength()

	// Strip mode: replace with empty method set.
	if keepSigs == nil {
		return p.ctx.ConstStruct([]llvm.Value{
			llvm.ConstInt(p.uintptrType, 0, false),
			llvm.ConstArray(p.ptrType, nil),
			llvm.ConstArray(p.ptrType, nil),
			llvm.ConstArray(p.ptrType, nil),
			llvm.ConstArray(p.uintptrType, nil),
		}, false)
	}

	if numMethods == 0 {
		return field
	}

	// Extract all methods and their signature names.
	type methodEntry struct {
		signature llvm.Value
		namePtr   llvm.Value
		typePtr   llvm.Value
		function  llvm.Value
		name      string
	}
	entries := make([]methodEntry, numMethods)
	nameSet := make(map[string]struct{}, numMethods)
	for j := range numMethods {
		sig := methodArray.AggregateElement(j)
		stripped := stripPointerCasts(sig)
		name := stripped.Name()
		entries[j] = methodEntry{
			signature: sig,
			namePtr:   nameArray.AggregateElement(j),
			typePtr:   typeArray.AggregateElement(j),
			function:  functionArray.AggregateElement(j),
			name:      name,
		}
		nameSet[name] = struct{}{}
	}

	// Check whether this type can fully implement at least one interface.
	// If not, its method set can never produce a true result from
	// typeImplementsMethodSet, so we can empty it entirely.
	implementsAny := false
	for _, ifaceSet := range ifaceSets {
		if isSubsetOf(ifaceSet, nameSet) {
			implementsAny = true
			break
		}
	}
	if !implementsAny {
		return p.ctx.ConstStruct([]llvm.Value{
			llvm.ConstInt(p.uintptrType, 0, false),
			llvm.ConstArray(p.ptrType, nil),
			llvm.ConstArray(p.ptrType, nil),
			llvm.ConstArray(p.ptrType, nil),
			llvm.ConstArray(p.uintptrType, nil),
		}, false)
	}

	// Prune: keep only methods whose signature appears in keepSigs.
	var keptSignatures []llvm.Value
	var keptNames []llvm.Value
	var keptTypes []llvm.Value
	var keptFunctions []llvm.Value
	for _, e := range entries {
		if _, ok := keepSigs[e.name]; ok {
			keptSignatures = append(keptSignatures, e.signature)
			if p.keepMethodNames {
				keptNames = append(keptNames, e.namePtr)
				keptTypes = append(keptTypes, e.typePtr)
				keptFunctions = append(keptFunctions, e.function)
			} else {
				keptNames = append(keptNames, llvm.ConstNull(p.ptrType))
				keptTypes = append(keptTypes, llvm.ConstNull(p.ptrType))
				keptFunctions = append(keptFunctions, llvm.ConstInt(p.uintptrType, 0, false))
			}
		}
	}

	if len(keptSignatures) == numMethods && p.keepMethodNames {
		return field
	}

	return p.ctx.ConstStruct([]llvm.Value{
		llvm.ConstInt(p.uintptrType, uint64(len(keptSignatures)), false),
		llvm.ConstArray(p.ptrType, keptSignatures),
		llvm.ConstArray(p.ptrType, keptNames),
		llvm.ConstArray(p.ptrType, keptTypes),
		llvm.ConstArray(p.uintptrType, keptFunctions),
	}, false)
}

// isSubsetOf reports whether every key in sub is also in super.
func isSubsetOf(sub, super map[string]struct{}) bool {
	for k := range sub {
		if _, ok := super[k]; !ok {
			return false
		}
	}
	return true
}
