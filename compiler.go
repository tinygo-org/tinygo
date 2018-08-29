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

var cgoWrapperError = errors.New("tinygo internal: cgo wrapper")

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
	c.ir.SimpleDCE()                   // remove most dead code
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
	// initializer" functions.
	for _, g := range c.ir.Globals {
		typ := g.g.Type().(*types.Pointer).Elem()
		llvmType, err := c.getLLVMType(typ)
		if err != nil {
			return err
		}
		global := llvm.AddGlobal(c.mod, llvmType, g.LinkName())
		g.llvmGlobal = global
		if !strings.HasPrefix(g.LinkName(), "_extern_") {
			global.SetLinkage(llvm.PrivateLinkage)
			if g.LinkName() == "runtime.TargetBits" {
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

	// Find and interpret package initializers.
	for _, frame := range frames {
		if frame.fn.fn.Synthetic == "package initializer" {
			c.initFuncs = append(c.initFuncs, frame.fn.llvmFn)
			if len(frame.fn.fn.Blocks) != 1 {
				panic("unexpected number of basic blocks in package initializer")
			}
			// Try to interpret as much as possible of the init() function.
			// Whenever it hits an instruction that it doesn't understand, it
			// bails out and leaves the rest to the compiler (so initialization
			// continues at runtime).
			// This should only happen when it hits a function call or the end
			// of the block, ideally.
			err := c.ir.Interpret(frame.fn.fn.Blocks[0])
			if err != nil {
				return err
			}
			err = c.parseFunc(frame)
			if err != nil {
				return err
			}
		}
	}

	// Set values for globals (after package initializer has been interpreted).
	for _, g := range c.ir.Globals {
		if g.initializer == nil {
			continue
		}
		err := c.parseGlobalInitializer(g)
		if err != nil {
			return err
		}
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
			continue // already done
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
	realMain := c.mod.NamedFunction(c.ir.mainPkg.Pkg.Path() + ".main")
	if c.ir.NeedsScheduler() {
		c.mod.NamedFunction("main.main$async").ReplaceAllUsesWith(realMain)
	} else {
		c.mod.NamedFunction("main.main").ReplaceAllUsesWith(realMain)
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
	// See src/runtime/interface.go for more details.
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
			llvm.ConstInt(llvm.Int16Type(), uint64(startIndex), false),
			llvm.ConstInt(llvm.Int16Type(), uint64(len(meta.Methods)), false),
		}
		rangeValue := llvm.ConstNamedStruct(rangeType, rangeValues)
		ranges = append(ranges, rangeValue)
		for _, method := range meta.Methods {
			f := c.ir.GetFunction(program.MethodValue(method))
			if f.llvmFn.IsNil() {
				return errors.New("cannot find function: " + f.LinkName())
			}
			fn := llvm.ConstBitCast(f.llvmFn, c.i8ptrType)
			funcPointers = append(funcPointers, fn)
			signatureNum := c.ir.MethodNum(method.Obj().(*types.Func))
			signature := llvm.ConstInt(llvm.Int16Type(), uint64(signatureNum), false)
			signatures = append(signatures, signature)
		}
		startIndex += len(meta.Methods)
	}

	if len(ranges) >= 1<<16 {
		return errors.New("method call numbers do not fit in a 16-bit integer")
	}

	// Replace the pre-created arrays with the generated arrays.
	rangeArray := llvm.ConstArray(rangeType, ranges)
	rangeArrayNewGlobal := llvm.AddGlobal(c.mod, rangeArray.Type(), "runtime.methodSetRanges.tmp")
	rangeArrayNewGlobal.SetInitializer(rangeArray)
	rangeArrayNewGlobal.SetLinkage(llvm.PrivateLinkage)
	rangeArrayOldGlobal := c.mod.NamedGlobal("runtime.methodSetRanges")
	rangeArrayOldGlobal.ReplaceAllUsesWith(llvm.ConstBitCast(rangeArrayNewGlobal, rangeArrayOldGlobal.Type()))
	rangeArrayOldGlobal.EraseFromParentAsGlobal()
	rangeArrayNewGlobal.SetName("runtime.methodSetRanges")
	funcArray := llvm.ConstArray(c.i8ptrType, funcPointers)
	funcArrayNewGlobal := llvm.AddGlobal(c.mod, funcArray.Type(), "runtime.methodSetFunctions.tmp")
	funcArrayNewGlobal.SetInitializer(funcArray)
	funcArrayNewGlobal.SetLinkage(llvm.PrivateLinkage)
	funcArrayOldGlobal := c.mod.NamedGlobal("runtime.methodSetFunctions")
	funcArrayOldGlobal.ReplaceAllUsesWith(llvm.ConstBitCast(funcArrayNewGlobal, funcArrayOldGlobal.Type()))
	funcArrayOldGlobal.EraseFromParentAsGlobal()
	funcArrayNewGlobal.SetName("runtime.methodSetFunctions")
	signatureArray := llvm.ConstArray(llvm.Int16Type(), signatures)
	signatureArrayNewGlobal := llvm.AddGlobal(c.mod, signatureArray.Type(), "runtime.methodSetSignatures.tmp")
	signatureArrayNewGlobal.SetInitializer(signatureArray)
	signatureArrayNewGlobal.SetLinkage(llvm.PrivateLinkage)
	signatureArrayOldGlobal := c.mod.NamedGlobal("runtime.methodSetSignatures")
	signatureArrayOldGlobal.ReplaceAllUsesWith(llvm.ConstBitCast(signatureArrayNewGlobal, signatureArrayOldGlobal.Type()))
	signatureArrayOldGlobal.EraseFromParentAsGlobal()
	signatureArrayNewGlobal.SetName("runtime.methodSetSignatures")

	c.mod.NamedGlobal("runtime.firstInterfaceNum").SetInitializer(llvm.ConstInt(llvm.Int16Type(), uint64(c.ir.FirstDynamicType()), false))

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
		case types.Bool, types.UntypedBool:
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
			return llvm.Type{}, errors.New("todo: unknown basic type: " + typ.String())
		}
	case *types.Interface:
		return c.mod.GetTypeByName("runtime._interface"), nil
	case *types.Map:
		return llvm.PointerType(c.mod.GetTypeByName("runtime.hashmap"), 0), nil
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
			// Multiple return values. Put them together in a struct.
			members := make([]llvm.Type, typ.Results().Len())
			for i := 0; i < typ.Results().Len(); i++ {
				returnType, err := c.getLLVMType(typ.Results().At(i).Type())
				if err != nil {
					return llvm.Type{}, err
				}
				members[i] = returnType
			}
			returnType = llvm.StructType(members, false)
		}
		// param values
		var paramTypes []llvm.Type
		if typ.Recv() != nil {
			recv, err := c.getLLVMType(typ.Recv().Type())
			if err != nil {
				return llvm.Type{}, err
			}
			if recv.StructName() == "runtime._interface" {
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
		results := make([]llvm.Type, 0, f.fn.Signature.Results().Len())
		for i := 0; i < f.fn.Signature.Results().Len(); i++ {
			typ, err := c.getLLVMType(f.fn.Signature.Results().At(i).Type())
			if err != nil {
				return nil, err
			}
			results = append(results, typ)
		}
		retType = llvm.StructType(results, false)
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

	name := f.LinkName()
	frame.fn.llvmFn = c.mod.NamedFunction(name)
	if frame.fn.llvmFn.IsNil() {
		frame.fn.llvmFn = llvm.AddFunction(c.mod, name, fnType)
	}
	return frame, nil
}

// Create a new global hashmap bucket, for map initialization.
func (c *Compiler) initMapNewBucket(mapType *types.Map) (llvm.Value, uint64, uint64, error) {
	llvmKeyType, err := c.getLLVMType(mapType.Key().Underlying())
	if err != nil {
		return llvm.Value{}, 0, 0, err
	}
	llvmValueType, err := c.getLLVMType(mapType.Elem().Underlying())
	if err != nil {
		return llvm.Value{}, 0, 0, err
	}
	keySize := c.targetData.TypeAllocSize(llvmKeyType)
	valueSize := c.targetData.TypeAllocSize(llvmValueType)
	bucketType := llvm.StructType([]llvm.Type{
		llvm.ArrayType(llvm.Int8Type(), 8), // tophash
		c.i8ptrType,                        // next bucket
		llvm.ArrayType(llvmKeyType, 8),     // key type
		llvm.ArrayType(llvmValueType, 8),   // value type
	}, false)
	bucketValue, err := getZeroValue(bucketType)
	if err != nil {
		return llvm.Value{}, 0, 0, err
	}
	bucket := llvm.AddGlobal(c.mod, bucketType, ".hashmap.bucket")
	bucket.SetInitializer(bucketValue)
	bucket.SetLinkage(llvm.PrivateLinkage)
	return bucket, keySize, valueSize, nil
}

func (c *Compiler) parseGlobalInitializer(g *Global) error {
	llvmValue, err := c.getInterpretedValue(g.initializer)
	if err != nil {
		return err
	}
	g.llvmGlobal.SetInitializer(llvmValue)
	return nil
}

// Turn a computed Value type (ConstValue, ArrayValue, etc.) into a LLVM value.
// This is used to set the initializer of globals after they have been
// calculated by the package initializer interpreter.
func (c *Compiler) getInterpretedValue(value Value) (llvm.Value, error) {
	switch value := value.(type) {
	case *ArrayValue:
		vals := make([]llvm.Value, len(value.Elems))
		for i, elem := range value.Elems {
			val, err := c.getInterpretedValue(elem)
			if err != nil {
				return llvm.Value{}, err
			}
			vals[i] = val
		}
		subTyp, err := c.getLLVMType(value.ElemType)
		if err != nil {
			return llvm.Value{}, err
		}
		return llvm.ConstArray(subTyp, vals), nil

	case *ConstValue:
		return c.parseConst(value.Expr)

	case *GlobalValue:
		zero := llvm.ConstInt(llvm.Int32Type(), 0, false)
		ptr := llvm.ConstInBoundsGEP(value.Global.llvmGlobal, []llvm.Value{zero})
		return ptr, nil

	case *InterfaceValue:
		itfTypeNum, ok := c.ir.TypeNum(value.Type)
		if !ok {
			panic("interface number is unknown")
		}
		if itfTypeNum >= 1<<16 {
			return llvm.Value{}, errors.New("interface typecodes do not fit in a 16-bit integer")
		}
		fields := []llvm.Value{
			llvm.ConstInt(llvm.Int16Type(), uint64(itfTypeNum), false),
			llvm.Undef(c.i8ptrType),
		}
		itf := llvm.ConstNamedStruct(c.mod.GetTypeByName("runtime._interface"), fields)
		return itf, nil

	case *MapValue:
		// Create initial bucket.
		firstBucketGlobal, keySize, valueSize, err := c.initMapNewBucket(value.Type)
		if err != nil {
			return llvm.Value{}, err
		}

		// Insert each key/value pair in the hashmap.
		bucketGlobal := firstBucketGlobal
		for i, key := range value.Keys {
			llvmKey, err := c.getInterpretedValue(key)
			if err != nil {
				return llvm.Value{}, nil
			}
			llvmValue, err := c.getInterpretedValue(value.Values[i])
			if err != nil {
				return llvm.Value{}, nil
			}

			keyString := constant.StringVal(key.(*ConstValue).Expr.Value)
			hash := stringhash(&keyString)

			if i%8 == 0 && i != 0 {
				// Bucket is full, create a new one.
				newBucketGlobal, _, _, err := c.initMapNewBucket(value.Type)
				if err != nil {
					return llvm.Value{}, err
				}
				zero := llvm.ConstInt(llvm.Int32Type(), 0, false)
				newBucketPtr := llvm.ConstInBoundsGEP(newBucketGlobal, []llvm.Value{zero})
				newBucketPtrCast := llvm.ConstBitCast(newBucketPtr, c.i8ptrType)
				// insert pointer into old bucket
				bucket := bucketGlobal.Initializer()
				bucket = llvm.ConstInsertValue(bucket, newBucketPtrCast, []uint32{1})
				bucketGlobal.SetInitializer(bucket)
				// switch to next bucket
				bucketGlobal = newBucketGlobal
			}

			tophashValue := llvm.ConstInt(llvm.Int8Type(), uint64(hashmapTopHash(hash)), false)
			bucket := bucketGlobal.Initializer()
			bucket = llvm.ConstInsertValue(bucket, tophashValue, []uint32{0, uint32(i % 8)})
			bucket = llvm.ConstInsertValue(bucket, llvmKey, []uint32{2, uint32(i % 8)})
			bucket = llvm.ConstInsertValue(bucket, llvmValue, []uint32{3, uint32(i % 8)})
			bucketGlobal.SetInitializer(bucket)
		}

		// Create the hashmap itself.
		zero := llvm.ConstInt(llvm.Int32Type(), 0, false)
		bucketPtr := llvm.ConstInBoundsGEP(bucketGlobal, []llvm.Value{zero})
		hashmapType := c.mod.GetTypeByName("runtime.hashmap")
		hashmap := llvm.ConstNamedStruct(hashmapType, []llvm.Value{
			llvm.ConstPointerNull(llvm.PointerType(hashmapType, 0)),  // next
			llvm.ConstBitCast(bucketPtr, c.i8ptrType),                // buckets
			llvm.ConstInt(c.lenType, uint64(len(value.Keys)), false), // count
			llvm.ConstInt(llvm.Int8Type(), keySize, false),           // keySize
			llvm.ConstInt(llvm.Int8Type(), valueSize, false),         // valueSize
			llvm.ConstInt(llvm.Int8Type(), 0, false),                 // bucketBits
		})

		// Create a pointer to this hashmap.
		hashmapPtr := llvm.AddGlobal(c.mod, hashmap.Type(), ".hashmap")
		hashmapPtr.SetInitializer(hashmap)
		hashmapPtr.SetLinkage(llvm.PrivateLinkage)
		return llvm.ConstInBoundsGEP(hashmapPtr, []llvm.Value{zero}), nil

	case *PointerBitCastValue:
		elem, err := c.getInterpretedValue(value.Elem)
		if err != nil {
			return llvm.Value{}, err
		}
		llvmType, err := c.getLLVMType(value.Type)
		if err != nil {
			return llvm.Value{}, err
		}
		return llvm.ConstBitCast(elem, llvmType), nil

	case *PointerToUintptrValue:
		elem, err := c.getInterpretedValue(value.Elem)
		if err != nil {
			return llvm.Value{}, err
		}
		return llvm.ConstPtrToInt(elem, c.uintptrType), nil

	case *PointerValue:
		elem, err := c.getInterpretedValue(*value.Elem)
		if err != nil {
			return llvm.Value{}, err
		}

		obj := llvm.AddGlobal(c.mod, elem.Type(), ".obj")
		obj.SetInitializer(elem)
		obj.SetLinkage(llvm.PrivateLinkage)
		elem = obj

		zero := llvm.ConstInt(llvm.Int32Type(), 0, false)
		ptr := llvm.ConstInBoundsGEP(elem, []llvm.Value{zero})
		return ptr, nil

	case *SliceValue:
		var globalPtr llvm.Value
		var arrayLength uint64
		if value.Array == nil {
			arrayType, err := c.getLLVMType(value.Type.Elem())
			if err != nil {
				return llvm.Value{}, err
			}
			globalPtr = llvm.ConstPointerNull(llvm.PointerType(arrayType, 0))
		} else {
			// make array
			array, err := c.getInterpretedValue(value.Array)
			if err != nil {
				return llvm.Value{}, err
			}
			// make global from array
			global := llvm.AddGlobal(c.mod, array.Type(), ".array")
			global.SetInitializer(array)
			global.SetLinkage(llvm.PrivateLinkage)

			// get pointer to global
			zero := llvm.ConstInt(llvm.Int32Type(), 0, false)
			globalPtr = c.builder.CreateInBoundsGEP(global, []llvm.Value{zero, zero}, "")

			arrayLength = uint64(len(value.Array.Elems))
		}

		// make slice
		sliceTyp, err := c.getLLVMType(value.Type)
		if err != nil {
			return llvm.Value{}, err
		}
		llvmLen := llvm.ConstInt(c.lenType, arrayLength, false)
		slice := llvm.ConstNamedStruct(sliceTyp, []llvm.Value{
			globalPtr, // ptr
			llvmLen,   // len
			llvmLen,   // cap
		})
		return slice, nil

	case *StructValue:
		fields := make([]llvm.Value, len(value.Fields))
		for i, elem := range value.Fields {
			field, err := c.getInterpretedValue(elem)
			if err != nil {
				return llvm.Value{}, err
			}
			fields[i] = field
		}
		switch value.Type.(type) {
		case *types.Named:
			llvmType, err := c.getLLVMType(value.Type)
			if err != nil {
				return llvm.Value{}, err
			}
			return llvm.ConstNamedStruct(llvmType, fields), nil
		case *types.Struct:
			return llvm.ConstStruct(fields, false), nil
		default:
			return llvm.Value{}, errors.New("init: unknown struct type: " + value.Type.String())
		}

	case *ZeroBasicValue:
		llvmType, err := c.getLLVMType(value.Type)
		if err != nil {
			return llvm.Value{}, err
		}
		return getZeroValue(llvmType)

	default:
		return llvm.Value{}, errors.New("init: unknown initializer type: " + fmt.Sprintf("%#v", value))
	}
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
		if err == cgoWrapperError {
			// Ignore CGo global variables which we don't use.
			return nil
		}
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
	case *ssa.MapUpdate:
		m, err := c.parseExpr(frame, instr.Map)
		if err != nil {
			return err
		}
		key, err := c.parseExpr(frame, instr.Key)
		if err != nil {
			return err
		}
		value, err := c.parseExpr(frame, instr.Value)
		if err != nil {
			return err
		}
		mapType := instr.Map.Type().Underlying().(*types.Map)
		switch keyType := mapType.Key().Underlying().(type) {
		case *types.Basic:
			if keyType.Kind() == types.String {
				valueAlloca := c.builder.CreateAlloca(value.Type(), "hashmap.value")
				c.builder.CreateStore(value, valueAlloca)
				valuePtr := c.builder.CreateBitCast(valueAlloca, c.i8ptrType, "hashmap.valueptr")
				params := []llvm.Value{m, key, valuePtr}
				fn := c.mod.NamedFunction("runtime.hashmapSet")
				c.builder.CreateCall(fn, params, "")
				return nil
			} else {
				return errors.New("todo: map update key type: " + keyType.String())
			}
		default:
			return errors.New("todo: map update key type: " + keyType.String())
		}
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
				// Multiple return values. Put them all in a struct.
				retVal, err := getZeroValue(frame.fn.llvmFn.Type().ElementType().ReturnType())
				if err != nil {
					return err
				}
				for i, result := range instr.Results {
					val, err := c.parseExpr(frame, result)
					if err != nil {
						return err
					}
					retVal = c.builder.CreateInsertValue(retVal, val, i, "")
				}
				c.builder.CreateRet(retVal)
				return nil
			}
		}
	case *ssa.Store:
		llvmAddr, err := c.parseExpr(frame, instr.Addr)
		if err == cgoWrapperError {
			// Ignore CGo global variables which we don't use.
			return nil
		}
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
		return errors.New("unknown instruction: " + instr.String())
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
		case *types.Map:
			indices := []llvm.Value{
				llvm.ConstInt(llvm.Int32Type(), 0, false),
				llvm.ConstInt(llvm.Int32Type(), 2, false), // hashmap.count
			}
			ptr := c.builder.CreateGEP(value, indices, "lenptr")
			return c.builder.CreateLoad(ptr, "len"), nil
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
			typ := arg.Type().Underlying()
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
					} else if typ.Kind() == types.Bool {
						c.builder.CreateCall(c.mod.NamedFunction("runtime.printbool"), []llvm.Value{value}, "")
					} else {
						return llvm.Value{}, errors.New("unknown basic arg type: " + typ.String())
					}
				}
			case *types.Pointer:
				ptrValue := c.builder.CreatePtrToInt(value, c.uintptrType, "")
				c.builder.CreateCall(c.mod.NamedFunction("runtime.printptr"), []llvm.Value{ptrValue}, "")
			default:
				return llvm.Value{}, errors.New("unknown arg type: " + typ.String())
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
			llvm.ConstInt(llvm.Int16Type(), uint64(c.ir.MethodNum(instr.Method)), false),
		}
		fn := c.builder.CreateCall(c.mod.NamedFunction("runtime.interfaceMethod"), values, "invoke.func")
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
		targetFunc := c.ir.GetFunction(call)
		name := targetFunc.LinkName()
		llvmFn := c.mod.NamedFunction(name)
		if llvmFn.IsNil() {
			return llvm.Value{}, errors.New("undefined function: " + name)
		}
		return c.parseFunctionCall(frame, instr.Args, llvmFn, targetFunc.blocking, parentHandle)
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
		x, err := c.parseExpr(frame, expr.X)
		if err != nil {
			return llvm.Value{}, err
		}
		return c.parseConvert(expr.X.Type(), expr.Type(), x)
	case *ssa.Const:
		return c.parseConst(expr)
	case *ssa.Convert:
		x, err := c.parseExpr(frame, expr.X)
		if err != nil {
			return llvm.Value{}, err
		}
		return c.parseConvert(expr.X.Type(), expr.Type(), x)
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
		return c.mod.NamedFunction(c.ir.GetFunction(expr).LinkName()), nil
	case *ssa.Global:
		if strings.HasPrefix(expr.Name(), "__cgofn__cgo_") || strings.HasPrefix(expr.Name(), "_cgo_") {
			// Ignore CGo global variables which we don't use.
			return llvm.Value{}, cgoWrapperError
		}
		value := c.ir.GetGlobal(expr).llvmGlobal
		if value.IsNil() {
			return llvm.Value{}, errors.New("global not found: " + c.ir.GetGlobal(expr).LinkName())
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
		switch ptrTyp := expr.X.Type().Underlying().(type) {
		case *types.Pointer:
			typ := expr.X.Type().(*types.Pointer).Elem().Underlying()
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
		if !lookupBoundsCheck.IsNil() && frame.fn.llvmFn.Name() != "runtime.interfaceMethod" {
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
		value, err := c.parseExpr(frame, expr.X)
		if err != nil {
			return llvm.Value{}, nil
		}
		index, err := c.parseExpr(frame, expr.Index)
		if err != nil {
			return llvm.Value{}, nil
		}
		switch xType := expr.X.Type().(type) {
		case *types.Basic:
			// Value type must be a string, which is a basic type.
			if xType.Kind() != types.String {
				panic("lookup on non-string?")
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
		case *types.Map:
			switch keyType := xType.Key().Underlying().(type) {
			case *types.Basic:
				if keyType.Kind() == types.String {
					llvmValueType, err := c.getLLVMType(expr.Type())
					if err != nil {
						return llvm.Value{}, err
					}
					mapValueAlloca := c.builder.CreateAlloca(llvmValueType, "hashmap.value")
					mapValuePtr := c.builder.CreateBitCast(mapValueAlloca, c.i8ptrType, "hashmap.valueptr")
					params := []llvm.Value{value, index, mapValuePtr}
					fn := c.mod.NamedFunction("runtime.hashmapGet")
					c.builder.CreateCall(fn, params, "")
					return c.builder.CreateLoad(mapValueAlloca, ""), nil
				} else {
					return llvm.Value{}, errors.New("todo: map lookup key type: " + keyType.String())
				}
			default:
				return llvm.Value{}, errors.New("todo: map lookup key type: " + keyType.String())
			}
		default:
			panic("unknown lookup type: " + expr.String())
		}
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
		if itfTypeNum >= 1<<16 {
			return llvm.Value{}, errors.New("interface typecodes do not fit in a 16-bit integer")
		}
		itf := llvm.ConstNamedStruct(c.mod.GetTypeByName("runtime._interface"), []llvm.Value{llvm.ConstInt(llvm.Int16Type(), uint64(itfTypeNum), false), llvm.Undef(c.i8ptrType)})
		itf = c.builder.CreateInsertValue(itf, itfValue, 1, "")
		return itf, nil
	case *ssa.MakeMap:
		mapType := expr.Type().Underlying().(*types.Map)
		llvmKeyType, err := c.getLLVMType(mapType.Key().Underlying())
		if err != nil {
			return llvm.Value{}, err
		}
		llvmValueType, err := c.getLLVMType(mapType.Elem().Underlying())
		if err != nil {
			return llvm.Value{}, err
		}
		switch keyType := mapType.Key().Underlying().(type) {
		case *types.Basic:
			if keyType.Kind() == types.String {
				keySize := c.targetData.TypeAllocSize(llvmKeyType)
				valueSize := c.targetData.TypeAllocSize(llvmValueType)
				hashmapMake := c.mod.NamedFunction("runtime.hashmapMake")
				llvmKeySize := llvm.ConstInt(llvm.Int8Type(), keySize, false)
				llvmValueSize := llvm.ConstInt(llvm.Int8Type(), valueSize, false)
				hashmap := c.builder.CreateCall(hashmapMake, []llvm.Value{llvmKeySize, llvmValueSize}, "")
				return hashmap, nil
			} else {
				return llvm.Value{}, errors.New("todo: map key type: " + keyType.String())
			}
		default:
			return llvm.Value{}, errors.New("todo: map key type: " + keyType.String())
		}
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
		if _, ok := expr.AssertedType.Underlying().(*types.Interface); ok {
			return llvm.Value{}, errors.New("todo: assert on interface")
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
		if assertedTypeNum >= 1<<16 {
			return llvm.Value{}, errors.New("interface typecodes do not fit in a 16-bit integer")
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
		commaOk := c.builder.CreateICmp(llvm.IntEQ, llvm.ConstInt(llvm.Int16Type(), uint64(assertedTypeNum), false), actualTypeNum, "")
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
	typ := binop.X.Type().Underlying()
	signed := false
	if typ, ok := typ.(*types.Basic); ok {
		signed = typ.Info()&types.IsUnsigned == 0
	}
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
	case token.SHL, token.SHR:
		sizeX := c.targetData.TypeAllocSize(x.Type())
		sizeY := c.targetData.TypeAllocSize(y.Type())
		if sizeX > sizeY {
			// x and y must have equal sizes, make Y bigger in this case.
			// y is unsigned, this has been checked by the Go type checker.
			y = c.builder.CreateZExt(y, x.Type(), "")
		}
		switch binop.Op {
		case token.SHL: // <<
			return c.builder.CreateShl(x, y, ""), nil
		case token.SHR: // >>
			if signed {
				return c.builder.CreateAShr(x, y, ""), nil
			} else {
				return c.builder.CreateLShr(x, y, ""), nil
			}
		default:
			panic("unreachable")
		}
	case token.AND_NOT: // &^
		// Go specific. Calculate "and not" with x & (~y)
		inv := c.builder.CreateNot(y, "") // ~y
		return c.builder.CreateAnd(x, inv, ""), nil
	case token.EQL, token.NEQ: // ==, !=
		switch typ := binop.X.Type().Underlying().(type) {
		case *types.Basic:
			if typ.Info()&types.IsInteger != 0 || typ.Kind() == types.UnsafePointer {
				if binop.Op == token.EQL {
					return c.builder.CreateICmp(llvm.IntEQ, x, y, ""), nil
				} else {
					return c.builder.CreateICmp(llvm.IntNE, x, y, ""), nil
				}
			} else if typ.Kind() == types.String {
				result := c.builder.CreateCall(c.mod.NamedFunction("runtime.stringequal"), []llvm.Value{x, y}, "")
				if binop.Op == token.NEQ {
					result = c.builder.CreateNot(result, "")
				}
				return result, nil
			} else {
				return llvm.Value{}, errors.New("todo: equality operator on unknown basic type: " + typ.String())
			}
		case *types.Pointer:
			if binop.Op == token.EQL {
				return c.builder.CreateICmp(llvm.IntEQ, x, y, ""), nil
			} else {
				return c.builder.CreateICmp(llvm.IntNE, x, y, ""), nil
			}
		default:
			return llvm.Value{}, errors.New("todo: equality operator on unknown type: " + typ.String())
		}
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
	switch typ := expr.Type().Underlying().(type) {
	case *types.Basic:
		llvmType, err := c.getLLVMType(typ)
		if err != nil {
			return llvm.Value{}, err
		}
		if typ.Info()&types.IsBoolean != 0 {
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
			global.SetGlobalConstant(true)
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
			return llvm.Value{}, errors.New("todo: unknown constant: " + expr.String())
		}
	case *types.Interface:
		if expr.Value != nil {
			return llvm.Value{}, errors.New("non-nil interface constant")
		}
		itfTypeNum, ok := c.ir.TypeNum(expr.Type())
		if itfTypeNum >= 1<<16 {
			return llvm.Value{}, errors.New("interface typecodes do not fit in a 16-bit integer")
		}
		if !ok {
			panic("interface number is unknown")
		}
		fields := []llvm.Value{
			llvm.ConstInt(llvm.Int16Type(), uint64(itfTypeNum), false),
			llvm.Undef(c.i8ptrType),
		}
		itf := llvm.ConstNamedStruct(c.mod.GetTypeByName("runtime._interface"), fields)
		return itf, nil
	case *types.Pointer:
		if expr.Value != nil {
			return llvm.Value{}, errors.New("non-nil pointer constant")
		}
		llvmType, err := c.getLLVMType(typ)
		if err != nil {
			return llvm.Value{}, err
		}
		return llvm.ConstPointerNull(llvmType), nil
	case *types.Slice:
		if expr.Value != nil {
			return llvm.Value{}, errors.New("non-nil slice constant")
		}
		elemType, err := c.getLLVMType(typ.Elem())
		if err != nil {
			return llvm.Value{}, err
		}
		llvmPtr := llvm.ConstPointerNull(llvm.PointerType(elemType, 0))
		llvmLen := llvm.ConstInt(c.lenType, 0, false)
		slice := llvm.ConstStruct([]llvm.Value{
			llvmPtr, // backing array
			llvmLen, // len
			llvmLen, // cap
		}, false)
		return slice, nil
	default:
		return llvm.Value{}, errors.New("todo: unknown constant: " + expr.String())
	}
}

func (c *Compiler) parseConvert(typeFrom, typeTo types.Type, value llvm.Value) (llvm.Value, error) {
	llvmTypeFrom := value.Type()
	llvmTypeTo, err := c.getLLVMType(typeTo)
	if err != nil {
		return llvm.Value{}, err
	}

	switch typeTo := typeTo.Underlying().(type) {
	case *types.Basic:
		isPtrFrom := isPointer(typeFrom)
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
	case *types.Pointer:
		return c.builder.CreateBitCast(value, llvmTypeTo, ""), nil
	default:
		return llvm.Value{}, errors.New("todo: convert: extend non-basic type: " + typeTo.String())
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
