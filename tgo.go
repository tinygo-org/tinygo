
package main

import (
	"errors"
	"flag"
	"fmt"
	"go/build"
	"go/constant"
	"go/token"
	"go/types"
	"os"
	"sort"
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
	triple          string
	mod             llvm.Module
	ctx             llvm.Context
	builder         llvm.Builder
	machine         llvm.TargetMachine
	targetData      llvm.TargetData
	intType         llvm.Type
	i8ptrType       llvm.Type // for convenience
	uintptrType     llvm.Type
	stringLenType   llvm.Type
	taskDataType    llvm.Type
	allocFunc       llvm.Value
	freeFunc        llvm.Value
	coroIdFunc      llvm.Value
	coroSizeFunc    llvm.Value
	coroBeginFunc   llvm.Value
	coroSuspendFunc llvm.Value
	coroEndFunc     llvm.Value
	coroFreeFunc    llvm.Value
	itfTypeNumbers  map[types.Type]uint64
	itfTypes        []types.Type
	initFuncs       []llvm.Value
	analysis        *Analysis
}

type Frame struct {
	fn           *ssa.Function
	llvmFn       llvm.Value
	params       map[*ssa.Parameter]int   // arguments to the function
	locals       map[ssa.Value]llvm.Value // local variables
	blocks       map[*ssa.BasicBlock]llvm.BasicBlock
	phis         []Phi
	blocking     bool
	taskState    llvm.Value
	taskHandle   llvm.Value
	cleanupBlock llvm.BasicBlock
	suspendBlock llvm.BasicBlock
}

func pkgPrefix(pkg *ssa.Package) string {
	if pkg.Pkg.Name() == "main" {
		return "main"
	}
	return pkg.Pkg.Path()
}

type Phi struct {
	ssa  *ssa.Phi
	llvm llvm.Value
}

func NewCompiler(pkgName, triple string) (*Compiler, error) {
	c := &Compiler{
		triple:         triple,
		itfTypeNumbers: make(map[types.Type]uint64),
		analysis:       NewAnalysis(),
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
	c.stringLenType = llvm.Int32Type()
	c.uintptrType = c.targetData.IntPtrType()
	c.i8ptrType = llvm.PointerType(llvm.Int8Type(), 0)

	// Go string: tuple of (len, ptr)
	t := c.ctx.StructCreateNamed("string")
	t.StructSetBody([]llvm.Type{c.stringLenType, c.i8ptrType}, false)

	// Go interface: tuple of (type, ptr)
	t = c.ctx.StructCreateNamed("interface")
	t.StructSetBody([]llvm.Type{llvm.Int32Type(), c.i8ptrType}, false)

	// Goroutine / task data: {i8 state, i32 data, i8* next}
	c.taskDataType = llvm.StructType([]llvm.Type{llvm.Int8Type(), llvm.Int32Type(), c.i8ptrType}, false)

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

	config := loader.Config {
		TypeChecker: types.Config{
			Sizes: &types.StdSizes{
				int64(c.targetData.PointerSize()),
				int64(c.targetData.PrefTypeAlignment(c.i8ptrType)),
			},
		},
		Build: &build.Context {
			GOARCH:      tripleSplit[0],
			GOOS:        tripleSplit[2],
			GOROOT:      ".",
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

	program := ssautil.CreateProgram(lprogram, ssa.SanityCheckFunctions | ssa.BareInits)
	program.Build()

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
		c.analysis.AddPackage(pkg)
	}
	c.analysis.AnalyseCallgraph()         // set up callgraph
	c.analysis.AnalyseBlockingRecursive() // make all parents of blocking calls blocking (transitively)
	c.analysis.AnalyseGoCalls()           // check whether we need a scheduler

	// Transform each package into LLVM IR.
	for _, pkg := range packageList {
		err := c.parsePackage(program, pkg)
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

	// Set functions referenced in runtime.ll to internal linkage, to improve
	// optimization (hopefully).
	main := c.mod.NamedFunction("main.main")
	if !main.IsDeclaration() {
		main.SetLinkage(llvm.PrivateLinkage)
	}
	mainAsync := c.mod.NamedFunction("main.main$async")
	if !mainAsync.IsDeclaration() {
		mainAsync.SetLinkage(llvm.PrivateLinkage)
	}
	c.mod.NamedFunction("runtime.scheduler").SetLinkage(llvm.PrivateLinkage)

	if c.analysis.NeedsScheduler() {
		// Enable the scheduler.
		c.mod.NamedGlobal(".has_scheduler").SetInitializer(llvm.ConstInt(llvm.Int1Type(), 1, false))
	}

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
		return llvm.Type{}, errors.New("todo: unknown type: " + fmt.Sprintf("%#v", goType))
	}
}

func (c *Compiler) getZeroValue(typ llvm.Type) (llvm.Value, error) {
	switch typ.TypeKind() {
	case llvm.ArrayTypeKind:
		subTyp := typ.ElementType()
		vals := make([]llvm.Value, typ.ArrayLength())
		for i := range vals {
			val, err := c.getZeroValue(subTyp)
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
			val, err := c.getZeroValue(subTyp)
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

func (c *Compiler) getInterfaceType(typ types.Type) llvm.Value {
	if _, ok := c.itfTypeNumbers[typ]; !ok {
		num := uint64(len(c.itfTypes))
		c.itfTypes = append(c.itfTypes, typ)
		c.itfTypeNumbers[typ] = num
	}
	return llvm.ConstInt(llvm.Int32Type(), c.itfTypeNumbers[typ], false)
}

func (c *Compiler) isPointer(typ types.Type) bool {
	if _, ok := typ.(*types.Pointer); ok {
		return true
	} else if typ, ok := typ.(*types.Basic); ok && typ.Kind() == types.UnsafePointer {
		return true
	} else {
		return false
	}
}

func getFunctionName(fn *ssa.Function, blocking bool) string {
	suffix := ""
	if blocking {
		suffix = "$async"
	}
	if fn.Signature.Recv() != nil {
		// Method on a defined type.
		typeName := fn.Params[0].Type().(*types.Named).Obj().Name()
		return pkgPrefix(fn.Pkg) + "." + typeName + "." + fn.Name() + suffix
	} else {
		// Bare function.
		if strings.HasPrefix(fn.Name(), "_Cfunc_") {
			// Name CGo functions directly.
			return fn.Name()[len("_Cfunc_"):]
		} else {
			name := pkgPrefix(fn.Pkg) + "." + fn.Name() + suffix
			if fn.Pkg.Pkg.Path() == "runtime" && strings.HasPrefix(fn.Name(), "_llvm_") {
				// Special case for LLVM intrinsics in the runtime.
				name = "llvm." + strings.Replace(fn.Name()[len("_llvm_"):], "_", ".", -1)
			}
			return name
		}
	}
}

func getGlobalName(global *ssa.Global) string {
	if strings.HasPrefix(global.Name(), "_extern_") {
		return global.Name()[len("_extern_"):]
	} else {
		return pkgPrefix(global.Pkg) + "." +  global.Name()
	}
}

func (c *Compiler) parsePackage(program *ssa.Program, pkg *ssa.Package) error {
	// Make sure we're walking through all members in a constant order every
	// run.
	memberNames := make([]string, 0)
	for name := range pkg.Members {
		if strings.HasPrefix(name, "_Cgo_") || strings.HasPrefix(name, "_cgo") {
			// _Cgo_ptr, _Cgo_use, _cgoCheckResult, _cgo_runtime_cgocall
			continue // CGo-internal functions
		}
		if strings.HasPrefix(name, "__cgofn__cgo_") {
			continue // CGo function pointer in global scope
		}
		memberNames = append(memberNames, name)
	}
	sort.Strings(memberNames)

	frames := make(map[*ssa.Function]*Frame)

	// First, build all function declarations.
	for _, name := range memberNames {
		member := pkg.Members[name]

		switch member := member.(type) {
		case *ssa.Function:
			frame, err := c.parseFuncDecl(member)
			if err != nil {
				return err
			}
			frames[member] = frame
			if member.Synthetic == "package initializer" {
				c.initFuncs = append(c.initFuncs, frame.llvmFn)
			}
			// TODO: recursively anonymous functions
			for _, child := range member.AnonFuncs {
				frame, err := c.parseFuncDecl(child)
				if err != nil {
					return err
				}
				frames[child] = frame
			}
		case *ssa.NamedConst:
			// Ignore package-level untyped constants. The SSA form doesn't need
			// them.
		case *ssa.Global:
			typ := member.Type()
			if typPtr, ok := typ.(*types.Pointer); ok {
				typ = typPtr.Elem()
			} else {
				return errors.New("global is not a pointer")
			}
			llvmType, err := c.getLLVMType(typ)
			if err != nil {
				return err
			}
			global := llvm.AddGlobal(c.mod, llvmType, getGlobalName(member))
			if !strings.HasPrefix(member.Name(), "_extern_") {
				global.SetLinkage(llvm.PrivateLinkage)
				if getGlobalName(member) == "runtime.TargetBits" {
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
					initializer, err := c.getZeroValue(llvmType)
					if err != nil {
						return err
					}
					global.SetInitializer(initializer)
				}
			}
		case *ssa.Type:
			if !types.IsInterface(member.Type()) {
				ms := program.MethodSets.MethodSet(member.Type())
				for i := 0; i < ms.Len(); i++ {
					fn := program.MethodValue(ms.At(i))
					frame, err := c.parseFuncDecl(fn)
					if err != nil {
						return err
					}
					frames[fn] = frame
				}
			}
		default:
			return errors.New("todo: member: " + fmt.Sprintf("%#v", member))
		}
	}

	// Now, add definitions to those declarations.
	for _, name := range memberNames {
		member := pkg.Members[name]
		switch member := member.(type) {
		case *ssa.Function:
			if strings.HasPrefix(name, "_Cfunc_") {
				// CGo function. Don't implement it's body.
				continue
			}
			if member.Blocks == nil {
				continue // external function
			}
			var err error
			if member.Synthetic == "package initializer" {
				err = c.parseInitFunc(frames[member], member)
			} else {
				err = c.parseFunc(frames[member], member)
			}
			if err != nil {
				return err
			}
		case *ssa.Type:
			if !types.IsInterface(member.Type()) {
				ms := program.MethodSets.MethodSet(member.Type())
				for i := 0; i < ms.Len(); i++ {
					fn := program.MethodValue(ms.At(i))
					err := c.parseFunc(frames[fn], fn)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (c *Compiler) parseFuncDecl(f *ssa.Function) (*Frame, error) {
	f.WriteTo(os.Stdout)

	frame := &Frame{
		fn:       f,
		params:   make(map[*ssa.Parameter]int),
		locals:   make(map[ssa.Value]llvm.Value),
		blocks:   make(map[*ssa.BasicBlock]llvm.BasicBlock),
		blocking: c.analysis.IsBlocking(f),
	}

	var retType llvm.Type
	if frame.blocking {
		if f.Signature.Results() != nil {
			return nil, errors.New("todo: return values in blocking function")
		}
		retType = c.i8ptrType
	} else if f.Signature.Results() == nil {
		retType = llvm.VoidType()
	} else if f.Signature.Results().Len() == 1 {
		var err error
		retType, err = c.getLLVMType(f.Signature.Results().At(0).Type())
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
	for i, param := range f.Params {
		paramType, err := c.getLLVMType(param.Type())
		if err != nil {
			return nil, err
		}
		paramTypes = append(paramTypes, paramType)
		frame.params[param] = i
	}

	fnType := llvm.FunctionType(retType, paramTypes, false)

	name := getFunctionName(f, frame.blocking)
	frame.llvmFn = c.mod.NamedFunction(name)
	if frame.llvmFn.IsNil() {
		frame.llvmFn = llvm.AddFunction(c.mod, name, fnType)
	}
	return frame, nil
}

// Special function parser for generated package initializers (which also
// initializes global variables).
func (c *Compiler) parseInitFunc(frame *Frame, f *ssa.Function) error {
	frame.llvmFn.SetLinkage(llvm.PrivateLinkage)
	llvmBlock := c.ctx.AddBasicBlock(frame.llvmFn, "entry")
	c.builder.SetInsertPointAtEnd(llvmBlock)

	for _, block := range f.DomPreorder() {
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
					llvmAddr := c.mod.NamedGlobal(getGlobalName(addr))
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
					llvmAddr := c.mod.NamedGlobal(getGlobalName(global))
					llvmValue := llvmAddr.Initializer()
					if llvmValue.IsNil() {
						llvmValue, err = c.getZeroValue(llvmAddr.Type().ElementType())
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
					llvmAddr := c.mod.NamedGlobal(getGlobalName(global))
					llvmValue := llvmAddr.Initializer()
					if llvmValue.IsNil() {
						llvmValue, err = c.getZeroValue(llvmAddr.Type().ElementType())
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

func (c *Compiler) parseFunc(frame *Frame, f *ssa.Function) error {
	frame.llvmFn.SetLinkage(llvm.PrivateLinkage)

	// Pre-create all basic blocks in the function.
	for _, block := range f.DomPreorder() {
		llvmBlock := c.ctx.AddBasicBlock(frame.llvmFn, block.Comment)
		frame.blocks[block] = llvmBlock
	}
	if frame.blocking {
		frame.cleanupBlock = c.ctx.AddBasicBlock(frame.llvmFn, "task.cleanup")
		frame.suspendBlock = c.ctx.AddBasicBlock(frame.llvmFn, "task.suspend")
	}

	// Load function parameters
	for _, param := range f.Params {
		llvmParam := frame.llvmFn.Param(frame.params[param])
		frame.locals[param] = llvmParam
	}

	if frame.blocking {
		// Coroutine initialization.
		c.builder.SetInsertPointAtEnd(frame.blocks[f.Blocks[0]])
		frame.taskState = c.builder.CreateAlloca(c.taskDataType, "task.state")
		stateI8 := c.builder.CreateBitCast(frame.taskState, c.i8ptrType, "task.state.i8")
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
		c.builder.CreateCall(c.mod.NamedFunction("runtime.scheduleTask"), []llvm.Value{frame.llvmFn.FirstParam()}, "")
		c.builder.CreateBr(frame.suspendBlock)

		// Coroutine suspend. A call to llvm.coro.suspend() will branch here.
		c.builder.SetInsertPointAtEnd(frame.suspendBlock)
		c.builder.CreateCall(c.coroEndFunc, []llvm.Value{frame.taskHandle, llvm.ConstInt(llvm.Int1Type(), 0, false)}, "unused")
		c.builder.CreateRet(frame.taskHandle)
	}

	// Fill blocks with instructions.
	for _, block := range f.DomPreorder() {
		c.builder.SetInsertPointAtEnd(frame.blocks[block])
		for _, instr := range block.Instrs {
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
		if !c.analysis.IsBlocking(instr.Common().Value) {
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
				case types.Int8:
					c.builder.CreateCall(c.mod.NamedFunction("runtime.printint8"), []llvm.Value{value}, "")
				case types.Uint8:
					c.builder.CreateCall(c.mod.NamedFunction("runtime.printuint8"), []llvm.Value{value}, "")
				case types.Int, types.Int32: // TODO: assumes a 32-bit int type
					c.builder.CreateCall(c.mod.NamedFunction("runtime.printint32"), []llvm.Value{value}, "")
				case types.Uint, types.Uint32:
					c.builder.CreateCall(c.mod.NamedFunction("runtime.printuint32"), []llvm.Value{value}, "")
				case types.Int64:
					c.builder.CreateCall(c.mod.NamedFunction("runtime.printint64"), []llvm.Value{value}, "")
				case types.Uint64:
					c.builder.CreateCall(c.mod.NamedFunction("runtime.printuint64"), []llvm.Value{value}, "")
				case types.String:
					c.builder.CreateCall(c.mod.NamedFunction("runtime.printstring"), []llvm.Value{value}, "")
				case types.Uintptr:
					c.builder.CreateCall(c.mod.NamedFunction("runtime.printptr"), []llvm.Value{value}, "")
				case types.UnsafePointer:
					ptrValue := c.builder.CreatePtrToInt(value, c.uintptrType, "")
					c.builder.CreateCall(c.mod.NamedFunction("runtime.printptr"), []llvm.Value{ptrValue}, "")
				default:
					return llvm.Value{}, errors.New("unknown basic arg type: " + fmt.Sprintf("%#v", typ))
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
		default:
			return llvm.Value{}, errors.New("todo: len: unknown type")
		}
	default:
		return llvm.Value{}, errors.New("todo: builtin: " + callName)
	}
}

func (c *Compiler) parseFunctionCall(frame *Frame, call *ssa.CallCommon, llvmFn llvm.Value, blocking bool, parentHandle llvm.Value) (llvm.Value, error) {
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
	for _, param := range call.Args {
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
		name := getFunctionName(call, targetBlocks)
		llvmFn := c.mod.NamedFunction(name)
		if llvmFn.IsNil() {
			targetBlocks = true
			nameAsync := getFunctionName(call, targetBlocks)
			llvmFn = c.mod.NamedFunction(nameAsync)
			if llvmFn.IsNil() {
				return llvm.Value{}, errors.New("undefined function: " + name)
			}
		}
		return c.parseFunctionCall(frame, instr, llvmFn, targetBlocks, parentHandle)
	default: // function pointer
		value, err := c.parseExpr(frame, instr.Value)
		if err != nil {
			return llvm.Value{}, err
		}
		// TODO: blocking function pointers (needs analysis)
		return c.parseFunctionCall(frame, instr, value, false, parentHandle)
	}
}

func (c *Compiler) parseExpr(frame *Frame, expr ssa.Value) (llvm.Value, error) {
	if frame != nil {
		if value, ok := frame.locals[expr]; ok {
			// Value is a local variable that has already been computed.
			if value.IsNil() {
				return llvm.Value{}, errors.New("undefined local var (from cgo?)")
			}
			return value, nil
		}
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
			zero, err := c.getZeroValue(typ)
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
		return c.mod.NamedFunction(getFunctionName(expr, false)), nil
	case *ssa.Global:
		fullName := getGlobalName(expr)
		value := c.mod.NamedGlobal(fullName)
		if value.IsNil() {
			return llvm.Value{}, errors.New("global not found: " + fullName)
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

		// Get buffer length
		var buflen llvm.Value
		typ := expr.X.Type().(*types.Pointer).Elem()
		switch typ := typ.(type) {
		case *types.Array:
			buflen = llvm.ConstInt(llvm.Int32Type(), uint64(typ.Len()), false)
		default:
			return llvm.Value{}, errors.New("todo: indexaddr: len")
		}

		// Bounds check.
		// LLVM optimizes this away in most cases.
		// TODO: runtime.boundsCheck is undefined in packages imported by
		// package runtime, so we have to remove it. This should be fixed.
		boundsCheck := c.mod.NamedFunction("runtime.boundsCheck")
		if !boundsCheck.IsNil() {
			constZero := llvm.ConstInt(c.intType, 0, false)
			isNegative := c.builder.CreateICmp(llvm.IntSLT, index, constZero, "") // index < 0
			isTooBig := c.builder.CreateICmp(llvm.IntSGE, index, buflen, "") // index >= len(value)
			isOverflow := c.builder.CreateOr(isNegative, isTooBig, "")
			c.builder.CreateCall(boundsCheck, []llvm.Value{isOverflow}, "")
		}

		indices := []llvm.Value{
			llvm.ConstInt(llvm.Int32Type(), 0, false),
			index,
		}
		return c.builder.CreateGEP(val, indices, ""), nil
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
		if frame.llvmFn.Name() != "runtime.boundsCheck" {
			constZero := llvm.ConstInt(c.intType, 0, false)
			isNegative := c.builder.CreateICmp(llvm.IntSLT, index, constZero, "") // index < 0
			strlen, err := c.parseBuiltin(frame, []ssa.Value{expr.X}, "len")
			if err != nil {
				return llvm.Value{}, err // shouldn't happen
			}
			isTooBig := c.builder.CreateICmp(llvm.IntSGE, index, strlen, "") // index >= len(value)
			isOverflow := c.builder.CreateOr(isNegative, isTooBig, "")
			c.builder.CreateCall(c.mod.NamedFunction("runtime.boundsCheck"), []llvm.Value{isOverflow}, "")
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
			// TODO: non-integers
			itfValue = c.builder.CreateIntToPtr(val, c.i8ptrType, "")
		}
		itfTypeNum := c.getInterfaceType(expr.X.Type())
		itf := llvm.ConstNamedStruct(c.mod.GetTypeByName("interface"), []llvm.Value{itfTypeNum, llvm.Undef(c.i8ptrType)})
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
		assertedTypeNum := c.getInterfaceType(expr.AssertedType)
		actualTypeNum := c.builder.CreateExtractValue(itf, 0, "interface.type")
		valuePtr := c.builder.CreateExtractValue(itf, 1, "interface.value")
		var value llvm.Value
		if c.targetData.TypeAllocSize(assertedType) > c.targetData.TypeAllocSize(c.i8ptrType) {
			// Value was stored in an allocated buffer, load it from there.
			valuePtrCast := c.builder.CreateBitCast(valuePtr, llvm.PointerType(assertedType, 0), "")
			value = c.builder.CreateLoad(valuePtrCast, "")
		} else {
			// Value was stored directly in the interface.
			// TODO: non-integer values.
			value = c.builder.CreatePtrToInt(valuePtr, assertedType, "")
		}
		// TODO: for interfaces, check whether the type implements the
		// interface.
		commaOk := c.builder.CreateICmp(llvm.IntEQ, assertedTypeNum, actualTypeNum, "")
		tuple := llvm.ConstStruct([]llvm.Value{llvm.Undef(assertedType), llvm.Undef(llvm.Int1Type())}, false) // create empty tuple
		tuple = c.builder.CreateInsertValue(tuple, value, 0, "") // insert value
		tuple = c.builder.CreateInsertValue(tuple, commaOk, 1, "") // insert 'comma ok' boolean
		return tuple, nil
	case *ssa.UnOp:
		return c.parseUnOp(frame, expr)
	default:
		return llvm.Value{}, errors.New("todo: unknown expression: " + fmt.Sprintf("%#v", expr))
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
	signed := typ.(*types.Basic).Info() & types.IsUnsigned == 0
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
	case token.OR:  // |
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
			strLen := llvm.ConstInt(c.stringLenType, uint64(len(str)), false)
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
		} else if typ.Info() & types.IsUnsigned != 0 {
			n, _ := constant.Uint64Val(expr.Value)
			return llvm.ConstInt(llvmType, n, false), nil
		} else if typ.Info() & types.IsInteger != 0 { // signed
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
		isPtrFrom := c.isPointer(x.Type())
		isPtrTo := c.isPointer(typeTo)
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

		if typeTo.Info() & types.IsInteger == 0 { // if not integer
			return llvm.Value{}, errors.New("todo: convert: extend non-integer type")
		}

		if sizeFrom > sizeTo {
			return c.builder.CreateTrunc(value, llvmTypeTo, ""), nil
		} else if typeTo.Info() & types.IsUnsigned != 0 { // if unsigned
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
			llvmGlobal := c.mod.NamedGlobal(getGlobalName(global))
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

// Helper function for Compiler object.
func Compile(pkgName, runtimePath, outpath, target string, printIR bool) error {
	var buildTags []string
	// TODO: put this somewhere else
	if target == "pca10040" {
		buildTags = append(buildTags, "nrf", "nrf52", "nrf52832")
		target = "armv7m-none-eabi"
	} else if target == "arduino" {
		buildTags = append(buildTags, "avr", "avr8", "atmega", "atmega328p")
		target = "avr--"
	}

	c, err := NewCompiler(pkgName, target)
	if err != nil {
		return err
	}

	// Add C/LLVM runtime.
	runtime, err := llvm.ParseBitcodeFile(runtimePath)
	if err != nil {
		return err
	}
	err = c.LinkModule(runtime)
	if err != nil {
		return err
	}

	// Compile Go code to IR.
	parseErr := func() error {
		if printIR {
			// Run this even if c.Parse() panics.
			defer func() {
				fmt.Println("IR until the error:")
				fmt.Println(c.IR())
			}()
		}
		return c.Parse(pkgName, buildTags)
	}()
	if parseErr != nil {
		return parseErr
	}

	c.ApplyFunctionSections() // -ffunction-sections

	if err := c.Verify(); err != nil {
		return err
	}
	//c.Optimize(2, 1) // -O2 -Os
	if err := c.Verify(); err != nil {
		return err
	}

	err = c.EmitObject(outpath)
	if err != nil {
		return err
	}

	return nil
}


func main() {
	outpath := flag.String("o", "", "output filename")
	printIR := flag.Bool("printir", false, "print LLVM IR after optimizing")
	runtime := flag.String("runtime", "", "runtime LLVM bitcode files (from C sources)")
	target := flag.String("target", llvm.DefaultTargetTriple(), "LLVM target")

	flag.Parse()

	if *outpath == "" || flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "usage: %s [-printir] -runtime=<runtime.bc> [-target=<target>] -o <output> <input>", os.Args[0])
		flag.PrintDefaults()
		return
	}

	os.Setenv("CC", "clang -target=" + *target)

	err := Compile(flag.Args()[0], *runtime, *outpath, *target, *printIR)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
