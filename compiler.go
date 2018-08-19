package main

import (
	"errors"
	"fmt"
	"go/build"
	"go/constant"
	"go/token"
	"go/types"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/aykevl/llvm/bindings/go/llvm"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

func init() {
	llvm.InitializeAllTargets()
	llvm.InitializeAllTargetMCs()
	llvm.InitializeAllTargetInfos()
	llvm.InitializeAllAsmParsers()
	llvm.InitializeAllAsmPrinters()
}

type Compiler struct {
	dumpSSA         bool
	triple          string
	mod             llvm.Module
	ctx             llvm.Context
	builder         llvm.Builder
	machine         llvm.TargetMachine
	targetData      llvm.TargetData
	intType         llvm.Type
	i8ptrType       llvm.Type // for convenience
	uintptrType     llvm.Type
	lenType         llvm.Type
	allocFunc       llvm.Value
	freeFunc        llvm.Value
	coroIdFunc      llvm.Value
	coroSizeFunc    llvm.Value
	coroBeginFunc   llvm.Value
	coroSuspendFunc llvm.Value
	coroEndFunc     llvm.Value
	coroFreeFunc    llvm.Value
	initFuncs       []llvm.Value
	ir              *Program
}

type Frame struct {
	fn           *Function
	params       map[*ssa.Parameter]int   // arguments to the function
	locals       map[ssa.Value]llvm.Value // local variables
	blocks       map[*ssa.BasicBlock]llvm.BasicBlock
	phis         []Phi
	blocking     bool
	taskHandle   llvm.Value
	cleanupBlock llvm.BasicBlock
	suspendBlock llvm.BasicBlock
}

type Phi struct {
	ssa  *ssa.Phi
	llvm llvm.Value
}

func NewCompiler(pkgName, triple string, dumpSSA bool) (*Compiler, error) {
	c := &Compiler{
		dumpSSA: dumpSSA,
		triple:  triple,
	}

	target, err := llvm.GetTargetFromTriple(triple)
	if err != nil {
		return nil, err
	}
	c.machine = target.CreateTargetMachine(triple, "", "", llvm.CodeGenLevelDefault, llvm.RelocDefault, llvm.CodeModelDefault)
	c.targetData = c.machine.CreateTargetData()

	c.mod = llvm.NewModule(pkgName)
	c.ctx = c.mod.Context()
	c.builder = c.ctx.NewBuilder()

	// Depends on platform (32bit or 64bit), but fix it here for now.
	c.intType = llvm.Int32Type()
	c.lenType = llvm.Int32Type()
	c.uintptrType = c.targetData.IntPtrType()
	c.i8ptrType = llvm.PointerType(llvm.Int8Type(), 0)

	// Go string: tuple of (len, ptr)
	t := c.ctx.StructCreateNamed("string")
	t.StructSetBody([]llvm.Type{c.lenType, c.i8ptrType}, false)

	allocType := llvm.FunctionType(c.i8ptrType, []llvm.Type{c.uintptrType}, false)
	c.allocFunc = llvm.AddFunction(c.mod, "runtime.alloc", allocType)

	freeType := llvm.FunctionType(llvm.VoidType(), []llvm.Type{c.i8ptrType}, false)
	c.freeFunc = llvm.AddFunction(c.mod, "runtime.free", freeType)

	coroIdType := llvm.FunctionType(c.ctx.TokenType(), []llvm.Type{llvm.Int32Type(), c.i8ptrType, c.i8ptrType, c.i8ptrType}, false)
	c.coroIdFunc = llvm.AddFunction(c.mod, "llvm.coro.id", coroIdType)

	coroSizeType := llvm.FunctionType(llvm.Int32Type(), nil, false)
	c.coroSizeFunc = llvm.AddFunction(c.mod, "llvm.coro.size.i32", coroSizeType)

	coroBeginType := llvm.FunctionType(c.i8ptrType, []llvm.Type{c.ctx.TokenType(), c.i8ptrType}, false)
	c.coroBeginFunc = llvm.AddFunction(c.mod, "llvm.coro.begin", coroBeginType)

	coroSuspendType := llvm.FunctionType(llvm.Int8Type(), []llvm.Type{c.ctx.TokenType(), llvm.Int1Type()}, false)
	c.coroSuspendFunc = llvm.AddFunction(c.mod, "llvm.coro.suspend", coroSuspendType)

	coroEndType := llvm.FunctionType(llvm.Int1Type(), []llvm.Type{c.i8ptrType, llvm.Int1Type()}, false)
	c.coroEndFunc = llvm.AddFunction(c.mod, "llvm.coro.end", coroEndType)

	coroFreeType := llvm.FunctionType(c.i8ptrType, []llvm.Type{c.ctx.TokenType(), c.i8ptrType}, false)
	c.coroFreeFunc = llvm.AddFunction(c.mod, "llvm.coro.free", coroFreeType)

	return c, nil
}

func (c *Compiler) Parse(mainPath string, buildTags []string) error {
	tripleSplit := strings.Split(c.triple, "-")

	config := loader.Config{
		TypeChecker: types.Config{
			Sizes: &types.StdSizes{
				int64(c.targetData.PointerSize()),
				int64(c.targetData.PrefTypeAlignment(c.i8ptrType)),
			},
		},
		Build: &build.Context{
			GOARCH:      tripleSplit[0],
			GOOS:        tripleSplit[2],
			GOROOT:      ".",
			GOPATH:      runtime.GOROOT(),
			CgoEnabled:  true,
			UseAllFiles: false,
			Compiler:    "gc", // must be one of the recognized compilers
			BuildTags:   append([]string{"tgo"}, buildTags...),
		},
		AllowErrors: true,
	}
	config.Import("runtime")
	config.Import(mainPath)
	lprogram, err := config.Load()
	if err != nil {
		return err
	}

	// TODO: pick the error of the first package, not a random package
	for _, pkgInfo := range lprogram.AllPackages {
		if len(pkgInfo.Errors) != 0 {
			return pkgInfo.Errors[0]
		}
	}

	program := ssautil.CreateProgram(lprogram, ssa.SanityCheckFunctions|ssa.BareInits)
	program.Build()
	c.ir = NewProgram(program, mainPath)

	// Make a list of packages in import order.
	packageList := []*ssa.Package{}
	packageSet := map[string]struct{}{}
	worklist := []string{"runtime", mainPath}
	for len(worklist) != 0 {
		pkgPath := worklist[0]
		pkg := program.ImportedPackage(pkgPath)
		if pkg == nil {
			// Non-SSA package (e.g. cgo).
			packageSet[pkgPath] = struct{}{}
			worklist = worklist[1:]
			continue
		}
		if _, ok := packageSet[pkgPath]; ok {
			// Package already in the final package list.
			worklist = worklist[1:]
			continue
		}

		unsatisfiedImports := make([]string, 0)
		imports := pkg.Pkg.Imports()
		for _, pkg := range imports {
			if _, ok := packageSet[pkg.Path()]; ok {
				continue
			}
			unsatisfiedImports = append(unsatisfiedImports, pkg.Path())
		}
		if len(unsatisfiedImports) == 0 {
			// All dependencies of this package are satisfied, so add this
			// package to the list.
			packageList = append(packageList, pkg)
			packageSet[pkgPath] = struct{}{}
			worklist = worklist[1:]
		} else {
			// Prepend all dependencies to the worklist and reconsider this
			// package (by not removing it from the worklist). At that point, it
			// must be possible to add it to packageList.
			worklist = append(unsatisfiedImports, worklist...)
		}
	}

	for _, pkg := range packageList {
		c.ir.AddPackage(pkg)
	}
	c.ir.AnalyseCallgraph()            // set up callgraph
	c.ir.AnalyseInterfaceConversions() // determine which types are converted to an interface
	c.ir.AnalyseBlockingRecursive()    // make all parents of blocking calls blocking (transitively)
	c.ir.AnalyseGoCalls()              // check whether we need a scheduler

	var frames []*Frame

	// Declare all named struct types.
	for _, t := range c.ir.NamedTypes {
		if named, ok := t.t.Type().(*types.Named); ok {
			if _, ok := named.Underlying().(*types.Struct); ok {
				t.llvmType = c.ctx.StructCreateNamed(named.Obj().Pkg().Path() + "." + named.Obj().Name())
			}
		}
	}

	// Define all named struct types.
	for _, t := range c.ir.NamedTypes {
		if named, ok := t.t.Type().(*types.Named); ok {
			if st, ok := named.Underlying().(*types.Struct); ok {
				llvmType, err := c.getLLVMType(st)
				if err != nil {
					return err
				}
				t.llvmType.StructSetBody(llvmType.StructElementTypes(), false)
			}
		}
	}

	// Declare all globals. These will get an initializer when parsing "package
	// initializer" packages.
	for _, g := range c.ir.Globals {
		typ := g.g.Type()
		if typPtr, ok := typ.(*types.Pointer); ok {
			typ = typPtr.Elem()
		} else {
			return errors.New("global is not a pointer")
		}
		llvmType, err := c.getLLVMType(typ)
		if err != nil {
			return err
		}
		global := llvm.AddGlobal(c.mod, llvmType, g.Name())
		g.llvmGlobal = global
		if !strings.HasPrefix(g.Name(), "_extern_") {
			global.SetLinkage(llvm.PrivateLinkage)
			if g.Name() == "runtime.TargetBits" {
				bitness := c.targetData.PointerSize() * 8
				if bitness < 32 {
					// Only 8 and 32+ architectures supported at the moment.
					// On 8 bit architectures, pointers are normally bigger
					// than 8 bits to do anything meaningful.
					// TODO: clean up this hack to support 16-bit
					// architectures.
					bitness = 8
				}
				global.SetInitializer(llvm.ConstInt(llvm.Int8Type(), uint64(bitness), false))
				global.SetGlobalConstant(true)
			} else {
				initializer, err := getZeroValue(llvmType)
				if err != nil {
					return err
				}
				global.SetInitializer(initializer)
			}
		}
	}

	// Declare all functions.
	for _, f := range c.ir.Functions {
		frame, err := c.parseFuncDecl(f)
		if err != nil {
			return err
		}
		frames = append(frames, frame)
	}

	// Add definitions to declarations.
	for _, frame := range frames {
		if frame.fn.CName() != "" {
			continue
		}
		if frame.fn.fn.Blocks == nil {
			continue // external function
		}
		var err error
		if frame.fn.fn.Synthetic == "package initializer" {
			err = c.parseInitFunc(frame)
		} else {
			err = c.parseFunc(frame)
		}
		if err != nil {
			return err
		}
	}

	// After all packages are imported, add a synthetic initializer function
	// that calls the initializer of each package.
	initFn := c.mod.NamedFunction("runtime.initAll")
	if initFn.IsNil() {
		initType := llvm.FunctionType(llvm.VoidType(), nil, false)
		initFn = llvm.AddFunction(c.mod, "runtime.initAll", initType)
	}
	initFn.SetLinkage(llvm.PrivateLinkage)
	block := c.ctx.AddBasicBlock(initFn, "entry")
	c.builder.SetInsertPointAtEnd(block)
	for _, fn := range c.initFuncs {
		c.builder.CreateCall(fn, nil, "")
	}
	c.builder.CreateRetVoid()

	// Adjust main function.
	main := c.mod.NamedFunction("main.main")
	realMain := c.mod.NamedFunction(c.ir.mainPkg.Pkg.Path() + ".main")
	if !realMain.IsNil() {
		main.ReplaceAllUsesWith(realMain)
	}
	mainAsync := c.mod.NamedFunction("main.main$async")
	realMainAsync := c.mod.NamedFunction(c.ir.mainPkg.Pkg.Path() + ".main$async")
	if !realMainAsync.IsNil() {
		mainAsync.ReplaceAllUsesWith(realMainAsync)
	}

	// Set functions referenced in runtime.ll to internal linkage, to improve
	// optimization (hopefully).
	c.mod.NamedFunction("runtime.scheduler").SetLinkage(llvm.PrivateLinkage)

	// Only use a scheduler when necessary.
	if c.ir.NeedsScheduler() {
		// Enable the scheduler.
		c.mod.NamedGlobal("has_scheduler").SetInitializer(llvm.ConstInt(llvm.Int1Type(), 1, false))
	}

	// Initialize runtime type information, for interfaces.
	dynamicTypes := c.ir.AllDynamicTypes()
	numDynamicTypes := 0
	for _, meta := range dynamicTypes {
		numDynamicTypes += len(meta.Methods)
	}
	tuples := make([]llvm.Value, 0, len(dynamicTypes))
	funcPointers := make([]llvm.Value, 0, numDynamicTypes)
	signatures := make([]llvm.Value, 0, numDynamicTypes)
	startIndex := 0
	tupleType := c.mod.GetTypeByName("interface_tuple")
	for _, meta := range dynamicTypes {
		tupleValues := []llvm.Value{
			llvm.ConstInt(llvm.Int32Type(), uint64(startIndex), false),
			llvm.ConstInt(llvm.Int32Type(), uint64(len(meta.Methods)), false),
		}
		tuple := llvm.ConstNamedStruct(tupleType, tupleValues)
		tuples = append(tuples, tuple)
		for _, method := range meta.Methods {
			f := c.ir.GetFunction(program.MethodValue(method))
			if f.llvmFn.IsNil() {
				return errors.New("cannot find function: " + f.Name(false))
			}
			fn := llvm.ConstBitCast(f.llvmFn, c.i8ptrType)
			funcPointers = append(funcPointers, fn)
			signatureNum := c.ir.MethodNum(method.Obj().(*types.Func))
			signature := llvm.ConstInt(llvm.Int32Type(), uint64(signatureNum), false)
			signatures = append(signatures, signature)
		}
		startIndex += len(meta.Methods)
	}
	// Replace the pre-created arrays with the generated arrays.
	tupleArray := llvm.ConstArray(tupleType, tuples)
	tupleArrayNewGlobal := llvm.AddGlobal(c.mod, tupleArray.Type(), "interface_tuples.tmp")
	tupleArrayNewGlobal.SetInitializer(tupleArray)
	tupleArrayOldGlobal := c.mod.NamedGlobal("interface_tuples")
	tupleArrayOldGlobal.ReplaceAllUsesWith(llvm.ConstBitCast(tupleArrayNewGlobal, tupleArrayOldGlobal.Type()))
	tupleArrayOldGlobal.EraseFromParentAsGlobal()
	tupleArrayNewGlobal.SetName("interface_tuples")
	funcArray := llvm.ConstArray(c.i8ptrType, funcPointers)
	funcArrayNewGlobal := llvm.AddGlobal(c.mod, funcArray.Type(), "interface_functions.tmp")
	funcArrayNewGlobal.SetInitializer(funcArray)
	funcArrayOldGlobal := c.mod.NamedGlobal("interface_functions")
	funcArrayOldGlobal.ReplaceAllUsesWith(llvm.ConstBitCast(funcArrayNewGlobal, funcArrayOldGlobal.Type()))
	funcArrayOldGlobal.EraseFromParentAsGlobal()
	funcArrayNewGlobal.SetName("interface_functions")
	signatureArray := llvm.ConstArray(llvm.Int32Type(), signatures)
	signatureArrayNewGlobal := llvm.AddGlobal(c.mod, signatureArray.Type(), "interface_signatures.tmp")
	signatureArrayNewGlobal.SetInitializer(signatureArray)
	signatureArrayOldGlobal := c.mod.NamedGlobal("interface_signatures")
	signatureArrayOldGlobal.ReplaceAllUsesWith(llvm.ConstBitCast(signatureArrayNewGlobal, signatureArrayOldGlobal.Type()))
	signatureArrayOldGlobal.EraseFromParentAsGlobal()
	signatureArrayNewGlobal.SetName("interface_signatures")

	c.mod.NamedGlobal("first_interface_num").SetInitializer(llvm.ConstInt(llvm.Int32Type(), uint64(c.ir.FirstDynamicType()), false))

	return nil
}

func (c *Compiler) getLLVMType(goType types.Type) (llvm.Type, error) {
	switch typ := goType.(type) {
	case *types.Array:
		elemType, err := c.getLLVMType(typ.Elem())
		if err != nil {
			return llvm.Type{}, err
		}
		return llvm.ArrayType(elemType, int(typ.Len())), nil
	case *types.Basic:
		switch typ.Kind() {
		case types.Bool:
			return llvm.Int1Type(), nil
		case types.Int8, types.Uint8:
			return llvm.Int8Type(), nil
		case types.Int16, types.Uint16:
			return llvm.Int16Type(), nil
		case types.Int32, types.Uint32:
			return llvm.Int32Type(), nil
		case types.Int, types.Uint:
			return c.intType, nil
		case types.Int64, types.Uint64:
			return llvm.Int64Type(), nil
		case types.String:
			return c.mod.GetTypeByName("string"), nil
		case types.Uintptr:
			return c.uintptrType, nil
		case types.UnsafePointer:
			return c.i8ptrType, nil
		default:
			return llvm.Type{}, errors.New("todo: unknown basic type: " + fmt.Sprintf("%#v", typ))
		}
	case *types.Interface:
		return c.mod.GetTypeByName("interface"), nil
	case *types.Named:
		if _, ok := typ.Underlying().(*types.Struct); ok {
			llvmType := c.mod.GetTypeByName(typ.Obj().Pkg().Path() + "." + typ.Obj().Name())
			if llvmType.IsNil() {
				return llvm.Type{}, errors.New("type not found: " + typ.Obj().Pkg().Path() + "." + typ.Obj().Name())
			}
			return llvmType, nil
		}
		return c.getLLVMType(typ.Underlying())
	case *types.Pointer:
		ptrTo, err := c.getLLVMType(typ.Elem())
		if err != nil {
			return llvm.Type{}, err
		}
		return llvm.PointerType(ptrTo, 0), nil
	case *types.Signature: // function pointer
		// return value
		var err error
		var returnType llvm.Type
		if typ.Results().Len() == 0 {
			returnType = llvm.VoidType()
		} else if typ.Results().Len() == 1 {
			returnType, err = c.getLLVMType(typ.Results().At(0).Type())
			if err != nil {
				return llvm.Type{}, err
			}
		} else {
			return llvm.Type{}, errors.New("todo: multiple return values in function pointer")
		}
		// param values
		var paramTypes []llvm.Type
		if typ.Recv() != nil {
			recv, err := c.getLLVMType(typ.Recv().Type())
			if err != nil {
				return llvm.Type{}, err
			}
			if recv.StructName() == "interface" {
				recv = c.i8ptrType
			}
			paramTypes = append(paramTypes, recv)
		}
		params := typ.Params()
		for i := 0; i < params.Len(); i++ {
			subType, err := c.getLLVMType(params.At(i).Type())
			if err != nil {
				return llvm.Type{}, err
			}
			paramTypes = append(paramTypes, subType)
		}
		// make a function pointer of it
		return llvm.PointerType(llvm.FunctionType(returnType, paramTypes, false), 0), nil
	case *types.Slice:
		elemType, err := c.getLLVMType(typ.Elem())
		if err != nil {
			return llvm.Type{}, err
		}
		members := []llvm.Type{
			llvm.PointerType(elemType, 0),
			c.lenType, // len
			c.lenType, // cap
		}
		return llvm.StructType(members, false), nil
	case *types.Struct:
		members := make([]llvm.Type, typ.NumFields())
		for i := 0; i < typ.NumFields(); i++ {
			member, err := c.getLLVMType(typ.Field(i).Type())
			if err != nil {
				return llvm.Type{}, err
			}
			members[i] = member
		}
		return llvm.StructType(members, false), nil
	default:
		return llvm.Type{}, errors.New("todo: unknown type: " + goType.String())
	}
}

// Return a zero LLVM value for any LLVM type. Setting this value as an
// initializer has the same effect as setting 'zeroinitializer' on a value.
// Sadly, I haven't found a way to do it directly with the Go API but this works
// just fine.
func getZeroValue(typ llvm.Type) (llvm.Value, error) {
	switch typ.TypeKind() {
	case llvm.ArrayTypeKind:
		subTyp := typ.ElementType()
		vals := make([]llvm.Value, typ.ArrayLength())
		for i := range vals {
			val, err := getZeroValue(subTyp)
			if err != nil {
				return llvm.Value{}, err
			}
			vals[i] = val
		}
		return llvm.ConstArray(subTyp, vals), nil
	case llvm.IntegerTypeKind:
		return llvm.ConstInt(typ, 0, false), nil
	case llvm.PointerTypeKind:
		return llvm.ConstPointerNull(typ), nil
	case llvm.StructTypeKind:
		types := typ.StructElementTypes()
		vals := make([]llvm.Value, len(types))
		for i, subTyp := range types {
			val, err := getZeroValue(subTyp)
			if err != nil {
				return llvm.Value{}, err
			}
			vals[i] = val
		}
		if typ.StructName() != "" {
			return llvm.ConstNamedStruct(typ, vals), nil
		} else {
			return llvm.ConstStruct(vals, false), nil
		}
	default:
		return llvm.Value{}, errors.New("todo: LLVM zero initializer")
	}
}

// Is this a pointer type of some sort? Can be unsafe.Pointer or any *T pointer.
func isPointer(typ types.Type) bool {
	if _, ok := typ.(*types.Pointer); ok {
		return true
	} else if typ, ok := typ.(*types.Basic); ok && typ.Kind() == types.UnsafePointer {
		return true
	} else {
		return false
	}
}

// Get all methods of a type: both value receivers and pointer receivers.
func getAllMethods(prog *ssa.Program, typ types.Type) []*types.Selection {
	var methods []*types.Selection

	// value receivers
	ms := prog.MethodSets.MethodSet(typ)
	for i := 0; i < ms.Len(); i++ {
		methods = append(methods, ms.At(i))
	}

	// pointer receivers
	ms = prog.MethodSets.MethodSet(types.NewPointer(typ))
	for i := 0; i < ms.Len(); i++ {
		methods = append(methods, ms.At(i))
	}

	return methods
}

// Return true if this is a CGo-internal function that can be ignored.
func isCGoInternal(name string) bool {
	if strings.HasPrefix(name, "_Cgo_") || strings.HasPrefix(name, "_cgo") {
		// _Cgo_ptr, _Cgo_use, _cgoCheckResult, _cgo_runtime_cgocall
		return true // CGo-internal functions
	}
	if strings.HasPrefix(name, "__cgofn__cgo_") {
		return true // CGo function pointer in global scope
	}
	return false
}

func (c *Compiler) parseFuncDecl(f *Function) (*Frame, error) {
	frame := &Frame{
		fn:       f,
		params:   make(map[*ssa.Parameter]int),
		locals:   make(map[ssa.Value]llvm.Value),
		blocks:   make(map[*ssa.BasicBlock]llvm.BasicBlock),
		blocking: c.ir.IsBlocking(f),
	}

	var retType llvm.Type
	if frame.blocking {
		if f.fn.Signature.Results() != nil {
			return nil, errors.New("todo: return values in blocking function")
		}
		retType = c.i8ptrType
	} else if f.fn.Signature.Results() == nil {
		retType = llvm.VoidType()
	} else if f.fn.Signature.Results().Len() == 1 {
		var err error
		retType, err = c.getLLVMType(f.fn.Signature.Results().At(0).Type())
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("todo: return values")
	}

	var paramTypes []llvm.Type
	if frame.blocking {
		paramTypes = append(paramTypes, c.i8ptrType) // parent coroutine
	}
	for i, param := range f.fn.Params {
		paramType, err := c.getLLVMType(param.Type())
		if err != nil {
			return nil, err
		}
		paramTypes = append(paramTypes, paramType)
		frame.params[param] = i
	}

	fnType := llvm.FunctionType(retType, paramTypes, false)

	name := f.Name(frame.blocking)
	frame.fn.llvmFn = c.mod.NamedFunction(name)
	if frame.fn.llvmFn.IsNil() {
		frame.fn.llvmFn = llvm.AddFunction(c.mod, name, fnType)
	}
	return frame, nil
}

// Special function parser for generated package initializers (which also
// initializes global variables).
func (c *Compiler) parseInitFunc(frame *Frame) error {
	frame.fn.llvmFn.SetLinkage(llvm.PrivateLinkage)
	llvmBlock := c.ctx.AddBasicBlock(frame.fn.llvmFn, "entry")
	c.builder.SetInsertPointAtEnd(llvmBlock)

	for _, block := range frame.fn.fn.DomPreorder() {
		for _, instr := range block.Instrs {
			var err error
			switch instr := instr.(type) {
			case *ssa.Call, *ssa.Return:
				err = c.parseInstr(frame, instr)
			case *ssa.Convert:
				// Ignore: CGo pointer conversion.
			case *ssa.FieldAddr, *ssa.IndexAddr:
				// Ignore: handled below with *ssa.Store.
			case *ssa.Store:
				switch addr := instr.Addr.(type) {
				case *ssa.Global:
					// Regular store, like a global int variable.
					if strings.HasPrefix(addr.Name(), "__cgofn__cgo_") || strings.HasPrefix(addr.Name(), "_cgo_") {
						// Ignore CGo global variables which we don't use.
						continue
					}
					val, err := c.parseExpr(frame, instr.Val)
					if err != nil {
						return err
					}
					llvmAddr := c.ir.GetGlobal(addr).llvmGlobal
					llvmAddr.SetInitializer(val)
				case *ssa.FieldAddr:
					// Initialize field of a global struct.
					// LLVM does not allow setting an initializer on part of a
					// global variable. So we take the current initializer, add
					// the field, and replace the initializer with the new
					// initializer.
					val, err := c.parseExpr(frame, instr.Val)
					if err != nil {
						return err
					}
					global := addr.X.(*ssa.Global)
					llvmAddr := c.ir.GetGlobal(global).llvmGlobal
					llvmValue := llvmAddr.Initializer()
					if llvmValue.IsNil() {
						llvmValue, err = getZeroValue(llvmAddr.Type().ElementType())
						if err != nil {
							return err
						}
					}
					llvmValue = c.builder.CreateInsertValue(llvmValue, val, addr.Field, "")
					llvmAddr.SetInitializer(llvmValue)
				case *ssa.IndexAddr:
					val, err := c.parseExpr(frame, instr.Val)
					if err != nil {
						return err
					}
					constIndex := addr.Index.(*ssa.Const)
					index, exact := constant.Int64Val(constIndex.Value)
					if !exact {
						return errors.New("could not get store index: " + constIndex.Value.ExactString())
					}
					fieldAddr := addr.X.(*ssa.FieldAddr)
					global := fieldAddr.X.(*ssa.Global)
					llvmAddr := c.ir.GetGlobal(global).llvmGlobal
					llvmValue := llvmAddr.Initializer()
					if llvmValue.IsNil() {
						llvmValue, err = getZeroValue(llvmAddr.Type().ElementType())
						if err != nil {
							return err
						}
					}
					llvmFieldValue := c.builder.CreateExtractValue(llvmValue, fieldAddr.Field, "")
					llvmFieldValue = c.builder.CreateInsertValue(llvmFieldValue, val, int(index), "")
					llvmValue = c.builder.CreateInsertValue(llvmValue, llvmFieldValue, fieldAddr.Field, "")
					llvmAddr.SetInitializer(llvmValue)
				default:
					return errors.New("unknown init store: " + fmt.Sprintf("%#v", addr))
				}
			default:
				return errors.New("unknown init instruction: " + fmt.Sprintf("%#v", instr))
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Compiler) parseFunc(frame *Frame) error {
	if c.dumpSSA {
		fmt.Printf("\nfunc %s:\n", frame.fn.fn)
	}
	frame.fn.llvmFn.SetLinkage(llvm.PrivateLinkage)

	// Pre-create all basic blocks in the function.
	for _, block := range frame.fn.fn.DomPreorder() {
		llvmBlock := c.ctx.AddBasicBlock(frame.fn.llvmFn, block.Comment)
		frame.blocks[block] = llvmBlock
	}
	if frame.blocking {
		frame.cleanupBlock = c.ctx.AddBasicBlock(frame.fn.llvmFn, "task.cleanup")
		frame.suspendBlock = c.ctx.AddBasicBlock(frame.fn.llvmFn, "task.suspend")
	}

	// Load function parameters
	for _, param := range frame.fn.fn.Params {
		llvmParam := frame.fn.llvmFn.Param(frame.params[param])
		frame.locals[param] = llvmParam
	}

	if frame.blocking {
		// Coroutine initialization.
		c.builder.SetInsertPointAtEnd(frame.blocks[frame.fn.fn.Blocks[0]])
		taskState := c.builder.CreateAlloca(c.mod.GetTypeByName("runtime.taskState"), "task.state")
		stateI8 := c.builder.CreateBitCast(taskState, c.i8ptrType, "task.state.i8")
		id := c.builder.CreateCall(c.coroIdFunc, []llvm.Value{
			llvm.ConstInt(llvm.Int32Type(), 0, false),
			stateI8,
			llvm.ConstNull(c.i8ptrType),
			llvm.ConstNull(c.i8ptrType),
		}, "task.token")
		size := c.builder.CreateCall(c.coroSizeFunc, nil, "task.size")
		if c.targetData.TypeAllocSize(size.Type()) > c.targetData.TypeAllocSize(c.uintptrType) {
			size = c.builder.CreateTrunc(size, c.uintptrType, "task.size.uintptr")
		} else if c.targetData.TypeAllocSize(size.Type()) < c.targetData.TypeAllocSize(c.uintptrType) {
			size = c.builder.CreateZExt(size, c.uintptrType, "task.size.uintptr")
		}
		data := c.builder.CreateCall(c.allocFunc, []llvm.Value{size}, "task.data")
		frame.taskHandle = c.builder.CreateCall(c.coroBeginFunc, []llvm.Value{id, data}, "task.handle")

		// Coroutine cleanup. Free resources associated with this coroutine.
		c.builder.SetInsertPointAtEnd(frame.cleanupBlock)
		mem := c.builder.CreateCall(c.coroFreeFunc, []llvm.Value{id, frame.taskHandle}, "task.data.free")
		c.builder.CreateCall(c.freeFunc, []llvm.Value{mem}, "")
		// re-insert parent coroutine
		c.builder.CreateCall(c.mod.NamedFunction("runtime.scheduleTask"), []llvm.Value{frame.fn.llvmFn.FirstParam()}, "")
		c.builder.CreateBr(frame.suspendBlock)

		// Coroutine suspend. A call to llvm.coro.suspend() will branch here.
		c.builder.SetInsertPointAtEnd(frame.suspendBlock)
		c.builder.CreateCall(c.coroEndFunc, []llvm.Value{frame.taskHandle, llvm.ConstInt(llvm.Int1Type(), 0, false)}, "unused")
		c.builder.CreateRet(frame.taskHandle)
	}

	// Fill blocks with instructions.
	for _, block := range frame.fn.fn.DomPreorder() {
		if c.dumpSSA {
			fmt.Printf("%s:\n", block.Comment)
		}
		c.builder.SetInsertPointAtEnd(frame.blocks[block])
		for _, instr := range block.Instrs {
			if c.dumpSSA {
				if val, ok := instr.(ssa.Value); ok && val.Name() != "" {
					fmt.Printf("\t%s = %s\n", val.Name(), val.String())
				} else {
					fmt.Printf("\t%s\n", instr.String())
				}
			}
			err := c.parseInstr(frame, instr)
			if err != nil {
				return err
			}
		}
	}

	// Resolve phi nodes
	for _, phi := range frame.phis {
		block := phi.ssa.Block()
		for i, edge := range phi.ssa.Edges {
			llvmVal, err := c.parseExpr(frame, edge)
			if err != nil {
				return err
			}
			llvmBlock := frame.blocks[block.Preds[i]]
			phi.llvm.AddIncoming([]llvm.Value{llvmVal}, []llvm.BasicBlock{llvmBlock})
		}
	}

	return nil
}

func (c *Compiler) parseInstr(frame *Frame, instr ssa.Instruction) error {
	switch instr := instr.(type) {
	case ssa.Value:
		value, err := c.parseExpr(frame, instr)
		frame.locals[instr] = value
		return err
	case *ssa.Go:
		if instr.Common().Method != nil {
			return errors.New("todo: go on method receiver")
		}

		// Execute non-blocking calls (including builtins) directly.
		// parentHandle param is ignored.
		if !c.ir.IsBlocking(c.ir.GetFunction(instr.Common().Value.(*ssa.Function))) {
			_, err := c.parseCall(frame, instr.Common(), llvm.Value{})
			return err // probably nil
		}

		// Start this goroutine.
		// parentHandle is nil, as the goroutine has no parent frame (it's a new
		// stack).
		handle, err := c.parseCall(frame, instr.Common(), llvm.Value{})
		if err != nil {
			return err
		}
		c.builder.CreateCall(c.mod.NamedFunction("runtime.scheduleTask"), []llvm.Value{handle}, "")
		return nil
	case *ssa.If:
		cond, err := c.parseExpr(frame, instr.Cond)
		if err != nil {
			return err
		}
		block := instr.Block()
		blockThen := frame.blocks[block.Succs[0]]
		blockElse := frame.blocks[block.Succs[1]]
		c.builder.CreateCondBr(cond, blockThen, blockElse)
		return nil
	case *ssa.Jump:
		blockJump := frame.blocks[instr.Block().Succs[0]]
		c.builder.CreateBr(blockJump)
		return nil
	case *ssa.Panic:
		value, err := c.parseExpr(frame, instr.X)
		if err != nil {
			return err
		}
		c.builder.CreateCall(c.mod.NamedFunction("runtime._panic"), []llvm.Value{value}, "")
		c.builder.CreateUnreachable()
		return nil
	case *ssa.Return:
		if frame.blocking {
			if len(instr.Results) != 0 {
				return errors.New("todo: return values from blocking function")
			}
			// Final suspend.
			continuePoint := c.builder.CreateCall(c.coroSuspendFunc, []llvm.Value{
				llvm.ConstNull(c.ctx.TokenType()),
				llvm.ConstInt(llvm.Int1Type(), 1, false), // final=true
			}, "")
			sw := c.builder.CreateSwitch(continuePoint, frame.suspendBlock, 2)
			sw.AddCase(llvm.ConstInt(llvm.Int8Type(), 1, false), frame.cleanupBlock)
			return nil
		} else {
			if len(instr.Results) == 0 {
				c.builder.CreateRetVoid()
				return nil
			} else if len(instr.Results) == 1 {
				val, err := c.parseExpr(frame, instr.Results[0])
				if err != nil {
					return err
				}
				c.builder.CreateRet(val)
				return nil
			} else {
				return errors.New("todo: return value")
			}
		}
	case *ssa.Store:
		llvmAddr, err := c.parseExpr(frame, instr.Addr)
		if err != nil {
			return err
		}
		llvmVal, err := c.parseExpr(frame, instr.Val)
		if err != nil {
			return err
		}
		valType := instr.Addr.Type().(*types.Pointer).Elem()
		if valType, ok := valType.(*types.Named); ok && valType.Obj().Name() == "__reg" {
			// Magic type name to transform this store to a register store.
			registerAddr := c.builder.CreateLoad(llvmAddr, "")
			ptr := c.builder.CreateIntToPtr(registerAddr, llvmAddr.Type(), "")
			store := c.builder.CreateStore(llvmVal, ptr)
			store.SetVolatile(true)
		} else {
			c.builder.CreateStore(llvmVal, llvmAddr)
		}
		return nil
	default:
		return errors.New("unknown instruction: " + fmt.Sprintf("%#v", instr))
	}
}

func (c *Compiler) parseBuiltin(frame *Frame, args []ssa.Value, callName string) (llvm.Value, error) {
	switch callName {
	case "cap":
		value, err := c.parseExpr(frame, args[0])
		if err != nil {
			return llvm.Value{}, err
		}
		switch args[0].Type().(type) {
		case *types.Slice:
			return c.builder.CreateExtractValue(value, 2, "cap"), nil
		default:
			return llvm.Value{}, errors.New("todo: cap: unknown type")
		}
	case "len":
		value, err := c.parseExpr(frame, args[0])
		if err != nil {
			return llvm.Value{}, err
		}
		switch typ := args[0].Type().(type) {
		case *types.Basic:
			switch typ.Kind() {
			case types.String:
				return c.builder.CreateExtractValue(value, 0, "len"), nil
			default:
				return llvm.Value{}, errors.New("todo: len: unknown basic type")
			}
		case *types.Slice:
			return c.builder.CreateExtractValue(value, 1, "len"), nil
		default:
			return llvm.Value{}, errors.New("todo: len: unknown type")
		}
	case "print", "println":
		for i, arg := range args {
			if i >= 1 {
				c.builder.CreateCall(c.mod.NamedFunction("runtime.printspace"), nil, "")
			}
			value, err := c.parseExpr(frame, arg)
			if err != nil {
				return llvm.Value{}, err
			}
			typ := arg.Type()
			if _, ok := typ.(*types.Named); ok {
				typ = typ.Underlying()
			}
			switch typ := typ.(type) {
			case *types.Basic:
				switch typ.Kind() {
				case types.String:
					c.builder.CreateCall(c.mod.NamedFunction("runtime.printstring"), []llvm.Value{value}, "")
				case types.Uintptr:
					c.builder.CreateCall(c.mod.NamedFunction("runtime.printptr"), []llvm.Value{value}, "")
				case types.UnsafePointer:
					ptrValue := c.builder.CreatePtrToInt(value, c.uintptrType, "")
					c.builder.CreateCall(c.mod.NamedFunction("runtime.printptr"), []llvm.Value{ptrValue}, "")
				default:
					// runtime.print{int,uint}{8,16,32,64}
					if typ.Info()&types.IsInteger != 0 {
						name := "runtime.print"
						if typ.Info()&types.IsUnsigned != 0 {
							name += "uint"
						} else {
							name += "int"
						}
						name += strconv.FormatUint(c.targetData.TypeAllocSize(value.Type())*8, 10)
						fn := c.mod.NamedFunction(name)
						if fn.IsNil() {
							panic("undefined: " + name)
						}
						c.builder.CreateCall(fn, []llvm.Value{value}, "")
						continue
					} else {
						return llvm.Value{}, errors.New("unknown basic arg type: " + fmt.Sprintf("%#v", typ))
					}
				}
			case *types.Pointer:
				ptrValue := c.builder.CreatePtrToInt(value, c.uintptrType, "")
				c.builder.CreateCall(c.mod.NamedFunction("runtime.printptr"), []llvm.Value{ptrValue}, "")
			default:
				return llvm.Value{}, errors.New("unknown arg type: " + fmt.Sprintf("%#v", typ))
			}
		}
		if callName == "println" {
			c.builder.CreateCall(c.mod.NamedFunction("runtime.printnl"), nil, "")
		}
		return llvm.Value{}, nil // print() or println() returns void
	case "ssa:wrapnilchk":
		// TODO: do an actual nil check?
		return c.parseExpr(frame, args[0])
	default:
		return llvm.Value{}, errors.New("todo: builtin: " + callName)
	}
}

func (c *Compiler) parseFunctionCall(frame *Frame, args []ssa.Value, llvmFn llvm.Value, blocking bool, parentHandle llvm.Value) (llvm.Value, error) {
	var params []llvm.Value
	if blocking {
		if parentHandle.IsNil() {
			// Started from 'go' statement.
			params = append(params, llvm.ConstNull(c.i8ptrType))
		} else {
			// Blocking function calls another blocking function.
			params = append(params, parentHandle)
		}
	}
	for _, param := range args {
		val, err := c.parseExpr(frame, param)
		if err != nil {
			return llvm.Value{}, err
		}
		params = append(params, val)
	}

	if frame.blocking && llvmFn.Name() == "runtime.Sleep" {
		// Set task state to TASK_STATE_SLEEP and set the duration.
		c.builder.CreateCall(c.mod.NamedFunction("runtime.sleepTask"), []llvm.Value{frame.taskHandle, params[0]}, "")

		// Yield to scheduler.
		continuePoint := c.builder.CreateCall(c.coroSuspendFunc, []llvm.Value{
			llvm.ConstNull(c.ctx.TokenType()),
			llvm.ConstInt(llvm.Int1Type(), 0, false),
		}, "")
		wakeup := c.ctx.InsertBasicBlock(llvm.NextBasicBlock(c.builder.GetInsertBlock()), "task.wakeup")
		sw := c.builder.CreateSwitch(continuePoint, frame.suspendBlock, 2)
		sw.AddCase(llvm.ConstInt(llvm.Int8Type(), 0, false), wakeup)
		sw.AddCase(llvm.ConstInt(llvm.Int8Type(), 1, false), frame.cleanupBlock)
		c.builder.SetInsertPointAtEnd(wakeup)

		return llvm.Value{}, nil
	}

	result := c.builder.CreateCall(llvmFn, params, "")
	if blocking && !parentHandle.IsNil() {
		// Calling a blocking function as a regular function call.
		// This is done by passing the current coroutine as a parameter to the
		// new coroutine and dropping the current coroutine from the scheduler
		// (with the TASK_STATE_CALL state). When the subroutine is finished, it
		// will reactivate the parent (this frame) in it's destroy function.

		c.builder.CreateCall(c.mod.NamedFunction("runtime.scheduleTask"), []llvm.Value{result}, "")

		// Set task state to TASK_STATE_CALL.
		c.builder.CreateCall(c.mod.NamedFunction("runtime.waitForAsyncCall"), []llvm.Value{frame.taskHandle}, "")

		// Yield to the scheduler.
		continuePoint := c.builder.CreateCall(c.coroSuspendFunc, []llvm.Value{
			llvm.ConstNull(c.ctx.TokenType()),
			llvm.ConstInt(llvm.Int1Type(), 0, false),
		}, "")
		resume := c.ctx.InsertBasicBlock(llvm.NextBasicBlock(c.builder.GetInsertBlock()), "task.callComplete")
		sw := c.builder.CreateSwitch(continuePoint, frame.suspendBlock, 2)
		sw.AddCase(llvm.ConstInt(llvm.Int8Type(), 0, false), resume)
		sw.AddCase(llvm.ConstInt(llvm.Int8Type(), 1, false), frame.cleanupBlock)
		c.builder.SetInsertPointAtEnd(resume)
	}
	return result, nil
}

func (c *Compiler) parseCall(frame *Frame, instr *ssa.CallCommon, parentHandle llvm.Value) (llvm.Value, error) {
	if instr.IsInvoke() {
		// Call an interface method with dynamic dispatch.
		itf, err := c.parseExpr(frame, instr.Value) // interface
		if err != nil {
			return llvm.Value{}, err
		}
		llvmFnType, err := c.getLLVMType(instr.Method.Type())
		if err != nil {
			return llvm.Value{}, err
		}
		values := []llvm.Value{
			itf,
			llvm.ConstInt(llvm.Int32Type(), uint64(c.ir.MethodNum(instr.Method)), false),
		}
		fn := c.builder.CreateCall(c.mod.NamedFunction("itfmethod"), values, "invoke.func")
		fnCast := c.builder.CreateBitCast(fn, llvmFnType, "invoke.func.cast")
		receiverValue := c.builder.CreateExtractValue(itf, 1, "invoke.func.receiver")
		args := []llvm.Value{receiverValue}
		for _, arg := range instr.Args {
			val, err := c.parseExpr(frame, arg)
			if err != nil {
				return llvm.Value{}, err
			}
			args = append(args, val)
		}
		// TODO: blocking methods (needs analysis)
		return c.builder.CreateCall(fnCast, args, ""), nil
	}

	// Regular function, builtin, or function pointer.
	switch call := instr.Value.(type) {
	case *ssa.Builtin:
		return c.parseBuiltin(frame, instr.Args, call.Name())
	case *ssa.Function:
		if call.Name() == "Asm" && len(instr.Args) == 1 {
			// Magic function: insert inline assembly instead of calling it.
			if named, ok := instr.Args[0].Type().(*types.Named); ok && named.Obj().Name() == "__asm" {
				fnType := llvm.FunctionType(llvm.VoidType(), []llvm.Type{}, false)
				asm := constant.StringVal(instr.Args[0].(*ssa.Const).Value)
				target := llvm.InlineAsm(fnType, asm, "", true, false, 0)
				return c.builder.CreateCall(target, nil, ""), nil
			}
		}
		targetBlocks := false
		name := c.ir.GetFunction(call).Name(targetBlocks)
		llvmFn := c.mod.NamedFunction(name)
		if llvmFn.IsNil() {
			targetBlocks = true
			nameAsync := c.ir.GetFunction(call).Name(targetBlocks)
			llvmFn = c.mod.NamedFunction(nameAsync)
			if llvmFn.IsNil() {
				return llvm.Value{}, errors.New("undefined function: " + name)
			}
		}
		return c.parseFunctionCall(frame, instr.Args, llvmFn, targetBlocks, parentHandle)
	default: // function pointer
		value, err := c.parseExpr(frame, instr.Value)
		if err != nil {
			return llvm.Value{}, err
		}
		// TODO: blocking function pointers (needs analysis)
		return c.parseFunctionCall(frame, instr.Args, value, false, parentHandle)
	}
}

func (c *Compiler) parseExpr(frame *Frame, expr ssa.Value) (llvm.Value, error) {
	if value, ok := frame.locals[expr]; ok {
		// Value is a local variable that has already been computed.
		if value.IsNil() {
			return llvm.Value{}, errors.New("undefined local var (from cgo?)")
		}
		return value, nil
	}

	switch expr := expr.(type) {
	case *ssa.Alloc:
		typ, err := c.getLLVMType(expr.Type().Underlying().(*types.Pointer).Elem())
		if err != nil {
			return llvm.Value{}, err
		}
		var buf llvm.Value
		if expr.Heap {
			// TODO: escape analysis
			size := llvm.ConstInt(c.uintptrType, c.targetData.TypeAllocSize(typ), false)
			buf = c.builder.CreateCall(c.allocFunc, []llvm.Value{size}, expr.Comment)
			buf = c.builder.CreateBitCast(buf, llvm.PointerType(typ, 0), "")
		} else {
			buf = c.builder.CreateAlloca(typ, expr.Comment)
			zero, err := getZeroValue(typ)
			if err != nil {
				return llvm.Value{}, err
			}
			c.builder.CreateStore(zero, buf) // zero-initialize var
		}
		return buf, nil
	case *ssa.BinOp:
		return c.parseBinOp(frame, expr)
	case *ssa.Call:
		// Passing the current task here to the subroutine. It is only used when
		// the subroutine is blocking.
		return c.parseCall(frame, expr.Common(), frame.taskHandle)
	case *ssa.ChangeInterface:
		// Do not change between interface types: always use the underlying
		// (concrete) type in the type number of the interface. Every method
		// call on an interface will do a lookup which method to call.
		// This is different from how the official Go compiler works, because of
		// heap allocation and because it's easier to implement, see:
		// https://research.swtch.com/interfaces
		return c.parseExpr(frame, expr.X)
	case *ssa.ChangeType:
		return c.parseConvert(frame, expr.Type(), expr.X)
	case *ssa.Const:
		return c.parseConst(expr)
	case *ssa.Convert:
		return c.parseConvert(frame, expr.Type(), expr.X)
	case *ssa.Extract:
		value, err := c.parseExpr(frame, expr.Tuple)
		if err != nil {
			return llvm.Value{}, err
		}
		result := c.builder.CreateExtractValue(value, expr.Index, "")
		return result, nil
	case *ssa.FieldAddr:
		val, err := c.parseExpr(frame, expr.X)
		if err != nil {
			return llvm.Value{}, err
		}
		indices := []llvm.Value{
			llvm.ConstInt(llvm.Int32Type(), 0, false),
			llvm.ConstInt(llvm.Int32Type(), uint64(expr.Field), false),
		}
		return c.builder.CreateGEP(val, indices, ""), nil
	case *ssa.Function:
		return c.mod.NamedFunction(c.ir.GetFunction(expr).Name(false)), nil
	case *ssa.Global:
		value := c.ir.GetGlobal(expr).llvmGlobal
		if value.IsNil() {
			return llvm.Value{}, errors.New("global not found: " + c.ir.GetGlobal(expr).Name())
		}
		return value, nil
	case *ssa.IndexAddr:
		val, err := c.parseExpr(frame, expr.X)
		if err != nil {
			return llvm.Value{}, err
		}
		index, err := c.parseExpr(frame, expr.Index)
		if err != nil {
			return llvm.Value{}, err
		}

		// Get buffer pointer and length
		var bufptr, buflen llvm.Value
		switch ptrTyp := expr.X.Type().(type) {
		case *types.Pointer:
			typ := expr.X.Type().(*types.Pointer).Elem()
			switch typ := typ.(type) {
			case *types.Array:
				bufptr = val
				buflen = llvm.ConstInt(llvm.Int32Type(), uint64(typ.Len()), false)
			default:
				return llvm.Value{}, errors.New("todo: indexaddr: " + typ.String())
			}
		case *types.Slice:
			bufptr = c.builder.CreateExtractValue(val, 0, "indexaddr.ptr")
			buflen = c.builder.CreateExtractValue(val, 1, "indexaddr.len")
		default:
			return llvm.Value{}, errors.New("todo: indexaddr: " + ptrTyp.String())
		}

		// Bounds check.
		// LLVM optimizes this away in most cases.
		// TODO: runtime.lookupBoundsCheck is undefined in packages imported by
		// package runtime, so we have to remove it. This should be fixed.
		lookupBoundsCheck := c.mod.NamedFunction("runtime.lookupBoundsCheck")
		if !lookupBoundsCheck.IsNil() {
			c.builder.CreateCall(lookupBoundsCheck, []llvm.Value{buflen, index}, "")
		}

		switch expr.X.Type().(type) {
		case *types.Pointer:
			indices := []llvm.Value{
				llvm.ConstInt(llvm.Int32Type(), 0, false),
				index,
			}
			return c.builder.CreateGEP(bufptr, indices, ""), nil
		case *types.Slice:
			return c.builder.CreateGEP(bufptr, []llvm.Value{index}, ""), nil
		default:
			panic("unreachable")
		}
	case *ssa.Lookup:
		if expr.CommaOk {
			return llvm.Value{}, errors.New("todo: lookup with comma-ok")
		}
		if _, ok := expr.X.Type().(*types.Map); ok {
			return llvm.Value{}, errors.New("todo: lookup in map")
		}
		// Value type must be a string, which is a basic type.
		if expr.X.Type().(*types.Basic).Kind() != types.String {
			panic("lookup on non-string?")
		}
		value, err := c.parseExpr(frame, expr.X)
		if err != nil {
			return llvm.Value{}, nil
		}
		index, err := c.parseExpr(frame, expr.Index)
		if err != nil {
			return llvm.Value{}, nil
		}

		// Bounds check.
		// LLVM optimizes this away in most cases.
		if frame.fn.llvmFn.Name() != "runtime.lookupBoundsCheck" {
			length, err := c.parseBuiltin(frame, []ssa.Value{expr.X}, "len")
			if err != nil {
				return llvm.Value{}, err // shouldn't happen
			}
			c.builder.CreateCall(c.mod.NamedFunction("runtime.lookupBoundsCheck"), []llvm.Value{length, index}, "")
		}

		// Lookup byte
		buf := c.builder.CreateExtractValue(value, 1, "")
		bufPtr := c.builder.CreateGEP(buf, []llvm.Value{index}, "")
		return c.builder.CreateLoad(bufPtr, ""), nil
	case *ssa.MakeInterface:
		val, err := c.parseExpr(frame, expr.X)
		if err != nil {
			return llvm.Value{}, err
		}
		var itfValue llvm.Value
		size := c.targetData.TypeAllocSize(val.Type())
		if size > c.targetData.TypeAllocSize(c.i8ptrType) {
			// Allocate on the heap and put a pointer in the interface.
			// TODO: escape analysis.
			sizeValue := llvm.ConstInt(c.uintptrType, size, false)
			itfValue = c.builder.CreateCall(c.allocFunc, []llvm.Value{sizeValue}, "")
			itfValueCast := c.builder.CreateBitCast(itfValue, llvm.PointerType(val.Type(), 0), "")
			c.builder.CreateStore(val, itfValueCast)
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
		itfTypeNum, _ := c.ir.TypeNum(expr.X.Type())
		itf := llvm.ConstNamedStruct(c.mod.GetTypeByName("interface"), []llvm.Value{llvm.ConstInt(llvm.Int32Type(), uint64(itfTypeNum), false), llvm.Undef(c.i8ptrType)})
		itf = c.builder.CreateInsertValue(itf, itfValue, 1, "")
		return itf, nil
	case *ssa.Phi:
		t, err := c.getLLVMType(expr.Type())
		if err != nil {
			return llvm.Value{}, err
		}
		phi := c.builder.CreatePHI(t, "")
		frame.phis = append(frame.phis, Phi{expr, phi})
		return phi, nil
	case *ssa.Slice:
		if expr.Max != nil {
			return llvm.Value{}, errors.New("todo: full slice expressions (with max): " + expr.Type().String())
		}
		value, err := c.parseExpr(frame, expr.X)
		if err != nil {
			return llvm.Value{}, err
		}
		switch typ := expr.X.Type().Underlying().(type) {
		case *types.Pointer: // pointer to array
			// slice an array
			length := typ.Elem().(*types.Array).Len()
			llvmLen := llvm.ConstInt(c.lenType, uint64(length), false)
			var low, high llvm.Value
			if expr.Low == nil {
				low = llvm.ConstInt(c.lenType, 0, false)
			} else {
				low, err = c.parseExpr(frame, expr.Low)
				if err != nil {
					return llvm.Value{}, nil
				}
			}
			if expr.High == nil {
				high = llvmLen
			} else {
				high, err = c.parseExpr(frame, expr.High)
				if err != nil {
					return llvm.Value{}, nil
				}
			}
			indices := []llvm.Value{
				llvm.ConstInt(llvm.Int32Type(), 0, false),
				low,
			}
			slicePtr := c.builder.CreateGEP(value, indices, "slice.ptr")
			sliceLen := c.builder.CreateSub(high, low, "slice.len")
			sliceCap := c.builder.CreateSub(llvmLen, low, "slice.cap")
			sliceTyp, err := c.getLLVMType(expr.Type())
			if err != nil {
				return llvm.Value{}, err
			}

			// This check is optimized away in most cases.
			sliceBoundsCheck := c.mod.NamedFunction("runtime.sliceBoundsCheck")
			c.builder.CreateCall(sliceBoundsCheck, []llvm.Value{llvmLen, low, high}, "")

			slice := llvm.ConstNamedStruct(sliceTyp, []llvm.Value{
				llvm.Undef(slicePtr.Type()),
				llvm.Undef(c.lenType),
				llvm.Undef(c.lenType),
			})
			slice = c.builder.CreateInsertValue(slice, slicePtr, 0, "")
			slice = c.builder.CreateInsertValue(slice, sliceLen, 1, "")
			slice = c.builder.CreateInsertValue(slice, sliceCap, 2, "")
			return slice, nil
		case *types.Slice:
			// slice a slice
			return llvm.Value{}, errors.New("todo: slice a slice: " + typ.String())
		case *types.Basic:
			// slice a string
			if typ.Kind() != types.String {
				return llvm.Value{}, errors.New("unknown slice type: " + typ.String())
			}
			return llvm.Value{}, errors.New("todo: slice a string")
		default:
			return llvm.Value{}, errors.New("unknown slice type: " + typ.String())
		}
	case *ssa.TypeAssert:
		if !expr.CommaOk {
			return llvm.Value{}, errors.New("todo: type assert without comma-ok")
		}
		itf, err := c.parseExpr(frame, expr.X)
		if err != nil {
			return llvm.Value{}, err
		}
		assertedType, err := c.getLLVMType(expr.AssertedType)
		if err != nil {
			return llvm.Value{}, err
		}
		assertedTypeNum, typeExists := c.ir.TypeNum(expr.AssertedType)
		if !typeExists {
			// Static analysis has determined this type assert will never apply.
			return llvm.ConstStruct([]llvm.Value{llvm.Undef(assertedType), llvm.ConstInt(llvm.Int1Type(), 0, false)}, false), nil
		}
		actualTypeNum := c.builder.CreateExtractValue(itf, 0, "interface.type")
		valuePtr := c.builder.CreateExtractValue(itf, 1, "interface.value")
		var value llvm.Value
		if c.targetData.TypeAllocSize(assertedType) > c.targetData.TypeAllocSize(c.i8ptrType) {
			// Value was stored in an allocated buffer, load it from there.
			valuePtrCast := c.builder.CreateBitCast(valuePtr, llvm.PointerType(assertedType, 0), "")
			value = c.builder.CreateLoad(valuePtrCast, "")
		} else {
			// Value was stored directly in the interface.
			switch assertedType.TypeKind() {
			case llvm.IntegerTypeKind:
				value = c.builder.CreatePtrToInt(valuePtr, assertedType, "")
			case llvm.PointerTypeKind:
				value = c.builder.CreateBitCast(valuePtr, assertedType, "")
			case llvm.StructTypeKind:
				// A bitcast would be useful here, but bitcast doesn't allow
				// aggregate types. So we'll bitcast it using an alloca.
				// Hopefully this will get optimized away.
				mem := c.builder.CreateAlloca(c.i8ptrType, "")
				c.builder.CreateStore(valuePtr, mem)
				memStructPtr := c.builder.CreateBitCast(mem, llvm.PointerType(assertedType, 0), "")
				value = c.builder.CreateLoad(memStructPtr, "")
			default:
				return llvm.Value{}, errors.New("todo: typeassert: bitcast small types")
			}
		}
		// TODO: for interfaces, check whether the type implements the
		// interface.
		commaOk := c.builder.CreateICmp(llvm.IntEQ, llvm.ConstInt(llvm.Int32Type(), uint64(assertedTypeNum), false), actualTypeNum, "")
		tuple := llvm.ConstStruct([]llvm.Value{llvm.Undef(assertedType), llvm.Undef(llvm.Int1Type())}, false) // create empty tuple
		tuple = c.builder.CreateInsertValue(tuple, value, 0, "")                                              // insert value
		tuple = c.builder.CreateInsertValue(tuple, commaOk, 1, "")                                            // insert 'comma ok' boolean
		return tuple, nil
	case *ssa.UnOp:
		return c.parseUnOp(frame, expr)
	default:
		return llvm.Value{}, errors.New("todo: unknown expression: " + expr.String())
	}
}

func (c *Compiler) parseBinOp(frame *Frame, binop *ssa.BinOp) (llvm.Value, error) {
	x, err := c.parseExpr(frame, binop.X)
	if err != nil {
		return llvm.Value{}, err
	}
	y, err := c.parseExpr(frame, binop.Y)
	if err != nil {
		return llvm.Value{}, err
	}
	typ := binop.X.Type()
	if typNamed, ok := typ.(*types.Named); ok {
		typ = typNamed.Underlying()
	}
	signed := typ.(*types.Basic).Info()&types.IsUnsigned == 0
	switch binop.Op {
	case token.ADD: // +
		return c.builder.CreateAdd(x, y, ""), nil
	case token.SUB: // -
		return c.builder.CreateSub(x, y, ""), nil
	case token.MUL: // *
		return c.builder.CreateMul(x, y, ""), nil
	case token.QUO: // /
		if signed {
			return c.builder.CreateSDiv(x, y, ""), nil
		} else {
			return c.builder.CreateUDiv(x, y, ""), nil
		}
	case token.REM: // %
		if signed {
			return c.builder.CreateSRem(x, y, ""), nil
		} else {
			return c.builder.CreateURem(x, y, ""), nil
		}
	case token.AND: // &
		return c.builder.CreateAnd(x, y, ""), nil
	case token.OR: // |
		return c.builder.CreateOr(x, y, ""), nil
	case token.XOR: // ^
		return c.builder.CreateXor(x, y, ""), nil
	case token.SHL: // <<
		return c.builder.CreateShl(x, y, ""), nil
	case token.SHR: // >>
		if signed {
			return c.builder.CreateAShr(x, y, ""), nil
		} else {
			return c.builder.CreateLShr(x, y, ""), nil
		}
	case token.AND_NOT: // &^
		// Go specific. Calculate "and not" with x & (~y)
		inv := c.builder.CreateNot(y, "") // ~y
		return c.builder.CreateAnd(x, inv, ""), nil
	case token.EQL: // ==
		return c.builder.CreateICmp(llvm.IntEQ, x, y, ""), nil
	case token.NEQ: // !=
		return c.builder.CreateICmp(llvm.IntNE, x, y, ""), nil
	case token.LSS: // <
		if signed {
			return c.builder.CreateICmp(llvm.IntSLT, x, y, ""), nil
		} else {
			return c.builder.CreateICmp(llvm.IntULT, x, y, ""), nil
		}
	case token.LEQ: // <=
		if signed {
			return c.builder.CreateICmp(llvm.IntSLE, x, y, ""), nil
		} else {
			return c.builder.CreateICmp(llvm.IntULE, x, y, ""), nil
		}
	case token.GTR: // >
		if signed {
			return c.builder.CreateICmp(llvm.IntSGT, x, y, ""), nil
		} else {
			return c.builder.CreateICmp(llvm.IntUGT, x, y, ""), nil
		}
	case token.GEQ: // >=
		if signed {
			return c.builder.CreateICmp(llvm.IntSGE, x, y, ""), nil
		} else {
			return c.builder.CreateICmp(llvm.IntUGE, x, y, ""), nil
		}
	default:
		return llvm.Value{}, errors.New("unknown binop")
	}
}

func (c *Compiler) parseConst(expr *ssa.Const) (llvm.Value, error) {
	typ := expr.Type()
	if named, ok := typ.(*types.Named); ok {
		typ = named.Underlying()
	}
	switch typ := typ.(type) {
	case *types.Basic:
		llvmType, err := c.getLLVMType(typ)
		if err != nil {
			return llvm.Value{}, err
		}
		if typ.Kind() == types.Bool {
			b := constant.BoolVal(expr.Value)
			n := uint64(0)
			if b {
				n = 1
			}
			return llvm.ConstInt(llvmType, n, false), nil
		} else if typ.Kind() == types.String {
			str := constant.StringVal(expr.Value)
			strLen := llvm.ConstInt(c.lenType, uint64(len(str)), false)
			global := llvm.AddGlobal(c.mod, llvm.ArrayType(llvm.Int8Type(), len(str)), ".str")
			global.SetInitializer(c.ctx.ConstString(str, false))
			global.SetLinkage(llvm.PrivateLinkage)
			global.SetGlobalConstant(false)
			zero := llvm.ConstInt(llvm.Int32Type(), 0, false)
			strPtr := c.builder.CreateInBoundsGEP(global, []llvm.Value{zero, zero}, "")
			strObj := llvm.ConstNamedStruct(c.mod.GetTypeByName("string"), []llvm.Value{strLen, strPtr})
			return strObj, nil
		} else if typ.Kind() == types.UnsafePointer {
			if !expr.IsNil() {
				return llvm.Value{}, errors.New("todo: non-null constant pointer")
			}
			return llvm.ConstNull(c.i8ptrType), nil
		} else if typ.Info()&types.IsUnsigned != 0 {
			n, _ := constant.Uint64Val(expr.Value)
			return llvm.ConstInt(llvmType, n, false), nil
		} else if typ.Info()&types.IsInteger != 0 { // signed
			n, _ := constant.Int64Val(expr.Value)
			return llvm.ConstInt(llvmType, uint64(n), true), nil
		} else {
			return llvm.Value{}, errors.New("todo: unknown constant: " + fmt.Sprintf("%v", typ))
		}
	default:
		return llvm.Value{}, errors.New("todo: unknown constant: " + fmt.Sprintf("%#v", typ))
	}
}

func (c *Compiler) parseConvert(frame *Frame, typeTo types.Type, x ssa.Value) (llvm.Value, error) {
	value, err := c.parseExpr(frame, x)
	if err != nil {
		return value, nil
	}

	llvmTypeFrom, err := c.getLLVMType(x.Type())
	if err != nil {
		return llvm.Value{}, err
	}
	llvmTypeTo, err := c.getLLVMType(typeTo)
	if err != nil {
		return llvm.Value{}, err
	}

	switch typeTo := typeTo.(type) {
	case *types.Basic:
		isPtrFrom := isPointer(x.Type())
		isPtrTo := isPointer(typeTo)
		if isPtrFrom && !isPtrTo {
			return c.builder.CreatePtrToInt(value, llvmTypeTo, ""), nil
		} else if !isPtrFrom && isPtrTo {
			return c.builder.CreateIntToPtr(value, llvmTypeTo, ""), nil
		}

		sizeFrom := c.targetData.TypeAllocSize(llvmTypeFrom)
		sizeTo := c.targetData.TypeAllocSize(llvmTypeTo)
		if sizeFrom == sizeTo {
			return c.builder.CreateBitCast(value, llvmTypeTo, ""), nil
		}

		if typeTo.Info()&types.IsInteger == 0 { // if not integer
			return llvm.Value{}, errors.New("todo: convert: extend non-integer type")
		}

		if sizeFrom > sizeTo {
			return c.builder.CreateTrunc(value, llvmTypeTo, ""), nil
		} else if typeTo.Info()&types.IsUnsigned != 0 { // if unsigned
			return c.builder.CreateZExt(value, llvmTypeTo, ""), nil
		} else { // if signed
			return c.builder.CreateSExt(value, llvmTypeTo, ""), nil
		}
	case *types.Named:
		return c.parseConvert(frame, typeTo.Underlying(), x)
	case *types.Pointer:
		return c.builder.CreateBitCast(value, llvmTypeTo, ""), nil
	default:
		return llvm.Value{}, errors.New("todo: convert: extend non-basic type: " + fmt.Sprintf("%#v", typeTo))
	}
}

func (c *Compiler) parseUnOp(frame *Frame, unop *ssa.UnOp) (llvm.Value, error) {
	x, err := c.parseExpr(frame, unop.X)
	if err != nil {
		return llvm.Value{}, err
	}
	switch unop.Op {
	case token.NOT: // !x
		return c.builder.CreateNot(x, ""), nil
	case token.SUB: // -x
		return c.builder.CreateSub(llvm.ConstInt(x.Type(), 0, false), x, ""), nil
	case token.MUL: // *x, dereference pointer
		valType := unop.X.Type().(*types.Pointer).Elem()
		if valType, ok := valType.(*types.Named); ok && valType.Obj().Name() == "__reg" {
			// Magic type name: treat the value as a register pointer.
			register := unop.X.(*ssa.FieldAddr)
			global := register.X.(*ssa.Global)
			llvmGlobal := c.ir.GetGlobal(global).llvmGlobal
			llvmAddr := c.builder.CreateExtractValue(llvmGlobal.Initializer(), register.Field, "")
			ptr := llvm.ConstIntToPtr(llvmAddr, x.Type())
			load := c.builder.CreateLoad(ptr, "")
			load.SetVolatile(true)
			return load, nil
		} else {
			return c.builder.CreateLoad(x, ""), nil
		}
	case token.XOR: // ^x, toggle all bits in integer
		return c.builder.CreateXor(x, llvm.ConstInt(x.Type(), ^uint64(0), false), ""), nil
	default:
		return llvm.Value{}, errors.New("todo: unknown unop")
	}
}

// IR returns the whole IR as a human-readable string.
func (c *Compiler) IR() string {
	return c.mod.String()
}

func (c *Compiler) Verify() error {
	return llvm.VerifyModule(c.mod, 0)
}

func (c *Compiler) LinkModule(mod llvm.Module) error {
	return llvm.LinkModules(c.mod, mod)
}

func (c *Compiler) ApplyFunctionSections() {
	// Put every function in a separate section. This makes it possible for the
	// linker to remove dead code (-ffunction-sections).
	llvmFn := c.mod.FirstFunction()
	for !llvmFn.IsNil() {
		if !llvmFn.IsDeclaration() {
			name := llvmFn.Name()
			if strings.HasSuffix(name, "$async") {
				name = name[:len(name)-len("$async")]
			}
			llvmFn.SetSection(".text." + name)
		}
		llvmFn = llvm.NextFunction(llvmFn)
	}
}

func (c *Compiler) Optimize(optLevel, sizeLevel int) {
	builder := llvm.NewPassManagerBuilder()
	defer builder.Dispose()
	builder.SetOptLevel(optLevel)
	builder.SetSizeLevel(sizeLevel)
	builder.UseInlinerWithThreshold(200) // TODO depend on opt level, and -Os

	funcPasses := llvm.NewFunctionPassManagerForModule(c.mod)
	defer funcPasses.Dispose()
	builder.PopulateFunc(funcPasses)

	modPasses := llvm.NewPassManager()
	defer modPasses.Dispose()
	builder.Populate(modPasses)

	modPasses.Run(c.mod)
}

func (c *Compiler) EmitObject(path string) error {
	// Generate output
	var buf []byte
	if strings.HasSuffix(path, ".o") {
		llvmBuf, err := c.machine.EmitToMemoryBuffer(c.mod, llvm.ObjectFile)
		if err != nil {
			return err
		}
		buf = llvmBuf.Bytes()
	} else if strings.HasSuffix(path, ".bc") {
		buf = llvm.WriteBitcodeToMemoryBuffer(c.mod).Bytes()
	} else if strings.HasSuffix(path, ".ll") {
		buf = []byte(c.mod.String())
	} else {
		return errors.New("unknown output file extension")
	}

	// Write output to file
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	f.Write(buf)
	f.Close()
	return nil
}
