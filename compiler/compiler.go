package compiler

import (
	"errors"
	"fmt"
	"go/build"
	"go/constant"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/aykevl/go-llvm"
	"github.com/aykevl/tinygo/ir"
	"github.com/aykevl/tinygo/loader"
	"golang.org/x/tools/go/ssa"
)

func init() {
	llvm.InitializeAllTargets()
	llvm.InitializeAllTargetMCs()
	llvm.InitializeAllTargetInfos()
	llvm.InitializeAllAsmParsers()
	llvm.InitializeAllAsmPrinters()
}

// Configure the compiler.
type Config struct {
	Triple     string   // LLVM target triple, e.g. x86_64-unknown-linux-gnu (empty string means default)
	CPU        string   // LLVM CPU name, e.g. atmega328p (empty string means default)
	GOOS       string   //
	GOARCH     string   //
	GC         string   // garbage collection strategy
	CFlags     []string // cflags to pass to cgo
	LDFlags    []string // ldflags to pass to cgo
	DumpSSA    bool     // dump Go SSA, for compiler debugging
	Debug      bool     // add debug symbols for gdb
	RootDir    string   // GOROOT for TinyGo
	GOPATH     string   // GOPATH, like `go env GOPATH`
	BuildTags  []string // build tags for TinyGo (empty means {Config.GOOS/Config.GOARCH})
	InitInterp bool     // use new init interpretation, meaning the old one is disabled
}

type Compiler struct {
	Config
	mod                     llvm.Module
	ctx                     llvm.Context
	builder                 llvm.Builder
	dibuilder               *llvm.DIBuilder
	cu                      llvm.Metadata
	difiles                 map[string]llvm.Metadata
	ditypes                 map[string]llvm.Metadata
	machine                 llvm.TargetMachine
	targetData              llvm.TargetData
	intType                 llvm.Type
	i8ptrType               llvm.Type // for convenience
	uintptrType             llvm.Type
	initFuncs               []llvm.Value
	interfaceInvokeWrappers []interfaceInvokeWrapper
	ir                      *ir.Program
}

type Frame struct {
	fn                *ir.Function
	locals            map[ssa.Value]llvm.Value            // local variables
	blockEntries      map[*ssa.BasicBlock]llvm.BasicBlock // a *ssa.BasicBlock may be split up
	blockExits        map[*ssa.BasicBlock]llvm.BasicBlock // these are the exit blocks
	currentBlock      *ssa.BasicBlock
	phis              []Phi
	taskHandle        llvm.Value
	deferPtr          llvm.Value
	difunc            llvm.Metadata
	allDeferFuncs     []interface{}
	deferFuncs        map[*ir.Function]int
	deferInvokeFuncs  map[string]int
	deferClosureFuncs map[*ir.Function]int
}

type Phi struct {
	ssa  *ssa.Phi
	llvm llvm.Value
}

func NewCompiler(pkgName string, config Config) (*Compiler, error) {
	if config.Triple == "" {
		config.Triple = llvm.DefaultTargetTriple()
	}
	if len(config.BuildTags) == 0 {
		config.BuildTags = []string{config.GOOS, config.GOARCH}
	}
	c := &Compiler{
		Config:  config,
		difiles: make(map[string]llvm.Metadata),
		ditypes: make(map[string]llvm.Metadata),
	}

	target, err := llvm.GetTargetFromTriple(config.Triple)
	if err != nil {
		return nil, err
	}
	c.machine = target.CreateTargetMachine(config.Triple, config.CPU, "", llvm.CodeGenLevelDefault, llvm.RelocStatic, llvm.CodeModelDefault)
	c.targetData = c.machine.CreateTargetData()

	c.ctx = llvm.NewContext()
	c.mod = c.ctx.NewModule(pkgName)
	c.mod.SetTarget(config.Triple)
	c.mod.SetDataLayout(c.targetData.String())
	c.builder = c.ctx.NewBuilder()
	c.dibuilder = llvm.NewDIBuilder(c.mod)

	c.uintptrType = c.ctx.IntType(c.targetData.PointerSize() * 8)
	if c.targetData.PointerSize() <= 4 {
		// 8, 16, 32 bits targets
		c.intType = c.ctx.Int32Type()
	} else if c.targetData.PointerSize() == 8 {
		// 64 bits target
		c.intType = c.ctx.Int64Type()
	} else {
		panic("unknown pointer size")
	}
	c.i8ptrType = llvm.PointerType(c.ctx.Int8Type(), 0)

	return c, nil
}

func (c *Compiler) Packages() []*loader.Package {
	return c.ir.LoaderProgram.Sorted()
}

// Return the LLVM module. Only valid after a successful compile.
func (c *Compiler) Module() llvm.Module {
	return c.mod
}

// Return the LLVM target data object. Only valid after a successful compile.
func (c *Compiler) TargetData() llvm.TargetData {
	return c.targetData
}

// selectGC picks an appropriate GC strategy if none was provided.
func (c *Compiler) selectGC() string {
	gc := c.GC
	if gc == "" {
		gc = "dumb"
	}
	return gc
}

// Compile the given package path or .go file path. Return an error when this
// fails (in any stage).
func (c *Compiler) Compile(mainPath string) error {
	// Prefix the GOPATH with the system GOROOT, as GOROOT is already set to
	// the TinyGo root.
	gopath := c.GOPATH
	if gopath == "" {
		gopath = runtime.GOROOT()
	} else {
		gopath = runtime.GOROOT() + string(filepath.ListSeparator) + gopath
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	lprogram := &loader.Program{
		Build: &build.Context{
			GOARCH:      c.GOARCH,
			GOOS:        c.GOOS,
			GOROOT:      c.RootDir,
			GOPATH:      gopath,
			CgoEnabled:  true,
			UseAllFiles: false,
			Compiler:    "gc", // must be one of the recognized compilers
			BuildTags:   append([]string{"tinygo", "gc." + c.selectGC()}, c.BuildTags...),
		},
		TypeChecker: types.Config{
			Sizes: &StdSizes{
				IntSize:  int64(c.targetData.TypeAllocSize(c.intType)),
				PtrSize:  int64(c.targetData.PointerSize()),
				MaxAlign: int64(c.targetData.PrefTypeAlignment(c.i8ptrType)),
			},
		},
		Dir:    wd,
		CFlags: c.CFlags,
	}
	if strings.HasSuffix(mainPath, ".go") {
		_, err = lprogram.ImportFile(mainPath)
		if err != nil {
			return err
		}
	} else {
		_, err = lprogram.Import(mainPath, wd)
		if err != nil {
			return err
		}
	}
	_, err = lprogram.Import("runtime", "")
	if err != nil {
		return err
	}

	err = lprogram.Parse()
	if err != nil {
		return err
	}

	c.ir = ir.NewProgram(lprogram, mainPath)

	// Run a simple dead code elimination pass.
	c.ir.SimpleDCE()

	// Initialize debug information.
	c.cu = c.dibuilder.CreateCompileUnit(llvm.DICompileUnit{
		Language:  llvm.DW_LANG_Go,
		File:      mainPath,
		Dir:       "",
		Producer:  "TinyGo",
		Optimized: true,
	})

	var frames []*Frame

	// Declare all named struct types.
	for _, t := range c.ir.NamedTypes {
		if named, ok := t.Type.Type().(*types.Named); ok {
			if _, ok := named.Underlying().(*types.Struct); ok {
				t.LLVMType = c.ctx.StructCreateNamed(named.Obj().Pkg().Path() + "." + named.Obj().Name())
			}
		}
	}

	// Define all named struct types.
	for _, t := range c.ir.NamedTypes {
		if named, ok := t.Type.Type().(*types.Named); ok {
			if st, ok := named.Underlying().(*types.Struct); ok {
				llvmType, err := c.getLLVMType(st)
				if err != nil {
					return err
				}
				t.LLVMType.StructSetBody(llvmType.StructElementTypes(), false)
			}
		}
	}

	// Declare all globals. These will get an initializer when parsing "package
	// initializer" functions.
	for _, g := range c.ir.Globals {
		typ := g.Type().(*types.Pointer).Elem()
		llvmType, err := c.getLLVMType(typ)
		if err != nil {
			return err
		}
		global := c.mod.NamedGlobal(g.LinkName())
		if global.IsNil() {
			global = llvm.AddGlobal(c.mod, llvmType, g.LinkName())
		}
		g.LLVMGlobal = global
		if !g.IsExtern() {
			global.SetLinkage(llvm.InternalLinkage)
			initializer, err := c.getZeroValue(llvmType)
			if err != nil {
				return err
			}
			global.SetInitializer(initializer)
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
		if frame.fn.Synthetic == "package initializer" {
			c.initFuncs = append(c.initFuncs, frame.fn.LLVMFn)
			// Try to interpret as much as possible of the init() function.
			// Whenever it hits an instruction that it doesn't understand, it
			// bails out and leaves the rest to the compiler (so initialization
			// continues at runtime).
			// This should only happen when it hits a function call or the end
			// of the block, ideally.
			if !c.InitInterp {
				err := c.ir.Interpret(frame.fn.Blocks[0], c.DumpSSA)
				if err != nil {
					return err
				}
			}
			err = c.parseFunc(frame)
			if err != nil {
				return err
			}
		}
	}

	// Set values for globals (after package initializer has been interpreted).
	for _, g := range c.ir.Globals {
		if g.Initializer() == nil {
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
		if frame.fn.Blocks == nil {
			continue // external function
		}
		var err error
		if frame.fn.Synthetic == "package initializer" {
			continue // already done
		} else {
			err = c.parseFunc(frame)
		}
		if err != nil {
			return err
		}
	}

	// Define the already declared functions that wrap methods for use in
	// interfaces.
	for _, state := range c.interfaceInvokeWrappers {
		err = c.createInterfaceInvokeWrapper(state)
		if err != nil {
			return err
		}
	}

	// After all packages are imported, add a synthetic initializer function
	// that calls the initializer of each package.
	initFn := c.ir.GetFunction(c.ir.Program.ImportedPackage("runtime").Members["initAll"].(*ssa.Function))
	initFn.LLVMFn.SetLinkage(llvm.InternalLinkage)
	initFn.LLVMFn.SetUnnamedAddr(true)
	if c.Debug {
		difunc, err := c.attachDebugInfo(initFn)
		if err != nil {
			return err
		}
		pos := c.ir.Program.Fset.Position(initFn.Pos())
		c.builder.SetCurrentDebugLocation(uint(pos.Line), uint(pos.Column), difunc, llvm.Metadata{})
	}
	block := c.ctx.AddBasicBlock(initFn.LLVMFn, "entry")
	c.builder.SetInsertPointAtEnd(block)
	for _, fn := range c.initFuncs {
		c.builder.CreateCall(fn, []llvm.Value{llvm.Undef(c.i8ptrType), llvm.Undef(c.i8ptrType)}, "")
	}
	c.builder.CreateRetVoid()

	// Conserve for goroutine lowering. Without marking these as external, they
	// would be optimized away.
	realMain := c.mod.NamedFunction(c.ir.MainPkg().Pkg.Path() + ".main")
	realMain.SetLinkage(llvm.ExternalLinkage) // keep alive until goroutine lowering
	c.mod.NamedFunction("runtime.alloc").SetLinkage(llvm.ExternalLinkage)
	c.mod.NamedFunction("runtime.free").SetLinkage(llvm.ExternalLinkage)
	c.mod.NamedFunction("runtime.chanSend").SetLinkage(llvm.ExternalLinkage)
	c.mod.NamedFunction("runtime.chanRecv").SetLinkage(llvm.ExternalLinkage)
	c.mod.NamedFunction("runtime.sleepTask").SetLinkage(llvm.ExternalLinkage)
	c.mod.NamedFunction("runtime.activateTask").SetLinkage(llvm.ExternalLinkage)
	c.mod.NamedFunction("runtime.scheduler").SetLinkage(llvm.ExternalLinkage)

	// see: https://reviews.llvm.org/D18355
	c.mod.AddNamedMetadataOperand("llvm.module.flags",
		c.ctx.MDNode([]llvm.Metadata{
			llvm.ConstInt(c.ctx.Int32Type(), 1, false).ConstantAsMetadata(), // Error on mismatch
			llvm.GlobalContext().MDString("Debug Info Version"),
			llvm.ConstInt(c.ctx.Int32Type(), 3, false).ConstantAsMetadata(), // DWARF version
		}),
	)
	c.dibuilder.Finalize()

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
			return c.ctx.Int1Type(), nil
		case types.Int8, types.Uint8:
			return c.ctx.Int8Type(), nil
		case types.Int16, types.Uint16:
			return c.ctx.Int16Type(), nil
		case types.Int32, types.Uint32:
			return c.ctx.Int32Type(), nil
		case types.Int, types.Uint:
			return c.intType, nil
		case types.Int64, types.Uint64:
			return c.ctx.Int64Type(), nil
		case types.Float32:
			return c.ctx.FloatType(), nil
		case types.Float64:
			return c.ctx.DoubleType(), nil
		case types.Complex64:
			return llvm.VectorType(c.ctx.FloatType(), 2), nil
		case types.Complex128:
			return llvm.VectorType(c.ctx.DoubleType(), 2), nil
		case types.String, types.UntypedString:
			return c.mod.GetTypeByName("runtime._string"), nil
		case types.Uintptr:
			return c.uintptrType, nil
		case types.UnsafePointer:
			return c.i8ptrType, nil
		default:
			return llvm.Type{}, errors.New("todo: unknown basic type: " + typ.String())
		}
	case *types.Chan:
		return llvm.PointerType(c.mod.GetTypeByName("runtime.channel"), 0), nil
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
			returnType = c.ctx.VoidType()
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
			returnType = c.ctx.StructType(members, false)
		}
		// param values
		var paramTypes []llvm.Type
		if typ.Recv() != nil {
			recv, err := c.getLLVMType(typ.Recv().Type())
			if err != nil {
				return llvm.Type{}, err
			}
			if recv.StructName() == "runtime._interface" {
				// This is a call on an interface, not a concrete type.
				// The receiver is not an interface, but a i8* type.
				recv = c.i8ptrType
			}
			paramTypes = append(paramTypes, c.expandFormalParamType(recv)...)
		}
		params := typ.Params()
		for i := 0; i < params.Len(); i++ {
			subType, err := c.getLLVMType(params.At(i).Type())
			if err != nil {
				return llvm.Type{}, err
			}
			paramTypes = append(paramTypes, c.expandFormalParamType(subType)...)
		}
		// make a closure type (with a function pointer type inside):
		// {context, funcptr}
		paramTypes = append(paramTypes, c.i8ptrType) // context
		paramTypes = append(paramTypes, c.i8ptrType) // parent coroutine
		ptr := llvm.PointerType(llvm.FunctionType(returnType, paramTypes, false), 0)
		ptr = c.ctx.StructType([]llvm.Type{c.i8ptrType, ptr}, false)
		return ptr, nil
	case *types.Slice:
		elemType, err := c.getLLVMType(typ.Elem())
		if err != nil {
			return llvm.Type{}, err
		}
		members := []llvm.Type{
			llvm.PointerType(elemType, 0),
			c.uintptrType, // len
			c.uintptrType, // cap
		}
		return c.ctx.StructType(members, false), nil
	case *types.Struct:
		members := make([]llvm.Type, typ.NumFields())
		for i := 0; i < typ.NumFields(); i++ {
			member, err := c.getLLVMType(typ.Field(i).Type())
			if err != nil {
				return llvm.Type{}, err
			}
			members[i] = member
		}
		return c.ctx.StructType(members, false), nil
	default:
		return llvm.Type{}, errors.New("todo: unknown type: " + goType.String())
	}
}

// Return a zero LLVM value for any LLVM type. Setting this value as an
// initializer has the same effect as setting 'zeroinitializer' on a value.
// Sadly, I haven't found a way to do it directly with the Go API but this works
// just fine.
func (c *Compiler) getZeroValue(typ llvm.Type) (llvm.Value, error) {
	switch typ.TypeKind() {
	case llvm.ArrayTypeKind:
		subTyp := typ.ElementType()
		subVal, err := c.getZeroValue(subTyp)
		if err != nil {
			return llvm.Value{}, err
		}
		vals := make([]llvm.Value, typ.ArrayLength())
		for i := range vals {
			vals[i] = subVal
		}
		return llvm.ConstArray(subTyp, vals), nil
	case llvm.FloatTypeKind, llvm.DoubleTypeKind:
		return llvm.ConstFloat(typ, 0.0), nil
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
			return c.ctx.ConstStruct(vals, false), nil
		}
	case llvm.VectorTypeKind:
		zero, err := c.getZeroValue(typ.ElementType())
		if err != nil {
			return llvm.Value{}, err
		}
		vals := make([]llvm.Value, typ.VectorSize())
		for i := range vals {
			vals[i] = zero
		}
		return llvm.ConstVector(vals, false), nil
	default:
		return llvm.Value{}, errors.New("todo: LLVM zero initializer: " + typ.String())
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

// Get the DWARF type for this Go type.
func (c *Compiler) getDIType(typ types.Type) (llvm.Metadata, error) {
	name := typ.String()
	if dityp, ok := c.ditypes[name]; ok {
		return dityp, nil
	} else {
		llvmType, err := c.getLLVMType(typ)
		if err != nil {
			return llvm.Metadata{}, err
		}
		sizeInBytes := c.targetData.TypeAllocSize(llvmType)
		var encoding llvm.DwarfTypeEncoding
		switch typ := typ.(type) {
		case *types.Basic:
			if typ.Info()&types.IsBoolean != 0 {
				encoding = llvm.DW_ATE_boolean
			} else if typ.Info()&types.IsFloat != 0 {
				encoding = llvm.DW_ATE_float
			} else if typ.Info()&types.IsComplex != 0 {
				encoding = llvm.DW_ATE_complex_float
			} else if typ.Info()&types.IsUnsigned != 0 {
				encoding = llvm.DW_ATE_unsigned
			} else if typ.Info()&types.IsInteger != 0 {
				encoding = llvm.DW_ATE_signed
			} else if typ.Kind() == types.UnsafePointer {
				encoding = llvm.DW_ATE_address
			}
		case *types.Pointer:
			encoding = llvm.DW_ATE_address
		}
		// TODO: other types
		dityp = c.dibuilder.CreateBasicType(llvm.DIBasicType{
			Name:       name,
			SizeInBits: sizeInBytes * 8,
			Encoding:   encoding,
		})
		c.ditypes[name] = dityp
		return dityp, nil
	}
}

func (c *Compiler) parseFuncDecl(f *ir.Function) (*Frame, error) {
	frame := &Frame{
		fn:           f,
		locals:       make(map[ssa.Value]llvm.Value),
		blockEntries: make(map[*ssa.BasicBlock]llvm.BasicBlock),
		blockExits:   make(map[*ssa.BasicBlock]llvm.BasicBlock),
	}

	var retType llvm.Type
	if f.Signature.Results() == nil {
		retType = c.ctx.VoidType()
	} else if f.Signature.Results().Len() == 1 {
		var err error
		retType, err = c.getLLVMType(f.Signature.Results().At(0).Type())
		if err != nil {
			return nil, err
		}
	} else {
		results := make([]llvm.Type, 0, f.Signature.Results().Len())
		for i := 0; i < f.Signature.Results().Len(); i++ {
			typ, err := c.getLLVMType(f.Signature.Results().At(i).Type())
			if err != nil {
				return nil, err
			}
			results = append(results, typ)
		}
		retType = c.ctx.StructType(results, false)
	}

	var paramTypes []llvm.Type
	for _, param := range f.Params {
		paramType, err := c.getLLVMType(param.Type())
		if err != nil {
			return nil, err
		}
		paramTypeFragments := c.expandFormalParamType(paramType)
		paramTypes = append(paramTypes, paramTypeFragments...)
	}

	// Add an extra parameter as the function context. This context is used in
	// closures and bound methods, but should be optimized away when not used.
	if !f.IsExported() {
		paramTypes = append(paramTypes, c.i8ptrType) // context
		paramTypes = append(paramTypes, c.i8ptrType) // parent coroutine
	}

	fnType := llvm.FunctionType(retType, paramTypes, false)

	name := f.LinkName()
	frame.fn.LLVMFn = c.mod.NamedFunction(name)
	if frame.fn.LLVMFn.IsNil() {
		frame.fn.LLVMFn = llvm.AddFunction(c.mod, name, fnType)
	}

	return frame, nil
}

func (c *Compiler) attachDebugInfo(f *ir.Function) (llvm.Metadata, error) {
	pos := c.ir.Program.Fset.Position(f.Syntax().Pos())
	return c.attachDebugInfoRaw(f, f.LLVMFn, "", pos.Filename, pos.Line)
}

func (c *Compiler) attachDebugInfoRaw(f *ir.Function, llvmFn llvm.Value, suffix, filename string, line int) (llvm.Metadata, error) {
	if _, ok := c.difiles[filename]; !ok {
		dir, file := filepath.Split(filename)
		if dir != "" {
			dir = dir[:len(dir)-1]
		}
		c.difiles[filename] = c.dibuilder.CreateFile(file, dir)
	}

	// Debug info for this function.
	diparams := make([]llvm.Metadata, 0, len(f.Params))
	for _, param := range f.Params {
		ditype, err := c.getDIType(param.Type())
		if err != nil {
			return llvm.Metadata{}, err
		}
		diparams = append(diparams, ditype)
	}
	diFuncType := c.dibuilder.CreateSubroutineType(llvm.DISubroutineType{
		File:       c.difiles[filename],
		Parameters: diparams,
		Flags:      0, // ?
	})
	difunc := c.dibuilder.CreateFunction(c.difiles[filename], llvm.DIFunction{
		Name:         f.RelString(nil) + suffix,
		LinkageName:  f.LinkName() + suffix,
		File:         c.difiles[filename],
		Line:         line,
		Type:         diFuncType,
		LocalToUnit:  true,
		IsDefinition: true,
		ScopeLine:    0,
		Flags:        llvm.FlagPrototyped,
		Optimized:    true,
	})
	llvmFn.SetSubprogram(difunc)
	return difunc, nil
}

// Create a new global hashmap bucket, for map initialization.
func (c *Compiler) initMapNewBucket(prefix string, mapType *types.Map) (llvm.Value, uint64, uint64, error) {
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
	bucketType := c.ctx.StructType([]llvm.Type{
		llvm.ArrayType(c.ctx.Int8Type(), 8), // tophash
		c.i8ptrType,                         // next bucket
		llvm.ArrayType(llvmKeyType, 8),      // key type
		llvm.ArrayType(llvmValueType, 8),    // value type
	}, false)
	bucketValue, err := c.getZeroValue(bucketType)
	if err != nil {
		return llvm.Value{}, 0, 0, err
	}
	bucket := llvm.AddGlobal(c.mod, bucketType, prefix+"$hashmap$bucket")
	bucket.SetInitializer(bucketValue)
	bucket.SetLinkage(llvm.InternalLinkage)
	return bucket, keySize, valueSize, nil
}

func (c *Compiler) parseGlobalInitializer(g *ir.Global) error {
	if g.IsExtern() {
		return nil
	}
	llvmValue, err := c.getInterpretedValue(g.LinkName(), g.Initializer())
	if err != nil {
		return err
	}
	g.LLVMGlobal.SetInitializer(llvmValue)
	return nil
}

// Turn a computed Value type (ConstValue, ArrayValue, etc.) into a LLVM value.
// This is used to set the initializer of globals after they have been
// calculated by the package initializer interpreter.
func (c *Compiler) getInterpretedValue(prefix string, value ir.Value) (llvm.Value, error) {
	switch value := value.(type) {
	case *ir.ArrayValue:
		vals := make([]llvm.Value, len(value.Elems))
		for i, elem := range value.Elems {
			val, err := c.getInterpretedValue(prefix+"$arrayval", elem)
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

	case *ir.ConstValue:
		return c.parseConst(prefix, value.Expr)

	case *ir.FunctionValue:
		if value.Elem == nil {
			llvmType, err := c.getLLVMType(value.Type)
			if err != nil {
				return llvm.Value{}, err
			}
			return c.getZeroValue(llvmType)
		}
		fn := c.ir.GetFunction(value.Elem)
		ptr := fn.LLVMFn
		// Create closure value: {context, function pointer}
		ptr = c.ctx.ConstStruct([]llvm.Value{llvm.ConstPointerNull(c.i8ptrType), ptr}, false)
		return ptr, nil

	case *ir.GlobalValue:
		zero := llvm.ConstInt(c.ctx.Int32Type(), 0, false)
		ptr := llvm.ConstInBoundsGEP(value.Global.LLVMGlobal, []llvm.Value{zero})
		return ptr, nil

	case *ir.MapValue:
		// Create initial bucket.
		firstBucketGlobal, keySize, valueSize, err := c.initMapNewBucket(prefix, value.Type)
		if err != nil {
			return llvm.Value{}, err
		}

		// Insert each key/value pair in the hashmap.
		bucketGlobal := firstBucketGlobal
		for i, key := range value.Keys {
			llvmKey, err := c.getInterpretedValue(prefix, key)
			if err != nil {
				return llvm.Value{}, nil
			}
			llvmValue, err := c.getInterpretedValue(prefix, value.Values[i])
			if err != nil {
				return llvm.Value{}, nil
			}

			constVal := key.(*ir.ConstValue).Expr
			var keyBuf []byte
			switch constVal.Type().Underlying().(*types.Basic).Kind() {
			case types.String, types.UntypedString:
				keyBuf = []byte(constant.StringVal(constVal.Value))
			case types.Int:
				keyBuf = make([]byte, c.targetData.TypeAllocSize(c.intType))
				n, _ := constant.Uint64Val(constVal.Value)
				for i := range keyBuf {
					keyBuf[i] = byte(n)
					n >>= 8
				}
			default:
				return llvm.Value{}, errors.New("todo: init: map key not implemented: " + constVal.Type().Underlying().String())
			}
			hash := hashmapHash(keyBuf)

			if i%8 == 0 && i != 0 {
				// Bucket is full, create a new one.
				newBucketGlobal, _, _, err := c.initMapNewBucket(prefix, value.Type)
				if err != nil {
					return llvm.Value{}, err
				}
				zero := llvm.ConstInt(c.ctx.Int32Type(), 0, false)
				newBucketPtr := llvm.ConstInBoundsGEP(newBucketGlobal, []llvm.Value{zero})
				newBucketPtrCast := llvm.ConstBitCast(newBucketPtr, c.i8ptrType)
				// insert pointer into old bucket
				bucket := bucketGlobal.Initializer()
				bucket = llvm.ConstInsertValue(bucket, newBucketPtrCast, []uint32{1})
				bucketGlobal.SetInitializer(bucket)
				// switch to next bucket
				bucketGlobal = newBucketGlobal
			}

			tophashValue := llvm.ConstInt(c.ctx.Int8Type(), uint64(hashmapTopHash(hash)), false)
			bucket := bucketGlobal.Initializer()
			bucket = llvm.ConstInsertValue(bucket, tophashValue, []uint32{0, uint32(i % 8)})
			bucket = llvm.ConstInsertValue(bucket, llvmKey, []uint32{2, uint32(i % 8)})
			bucket = llvm.ConstInsertValue(bucket, llvmValue, []uint32{3, uint32(i % 8)})
			bucketGlobal.SetInitializer(bucket)
		}

		// Create the hashmap itself.
		zero := llvm.ConstInt(c.ctx.Int32Type(), 0, false)
		bucketPtr := llvm.ConstInBoundsGEP(firstBucketGlobal, []llvm.Value{zero})
		hashmapType := c.mod.GetTypeByName("runtime.hashmap")
		hashmap := llvm.ConstNamedStruct(hashmapType, []llvm.Value{
			llvm.ConstPointerNull(llvm.PointerType(hashmapType, 0)),      // next
			llvm.ConstBitCast(bucketPtr, c.i8ptrType),                    // buckets
			llvm.ConstInt(c.uintptrType, uint64(len(value.Keys)), false), // count
			llvm.ConstInt(c.ctx.Int8Type(), keySize, false),              // keySize
			llvm.ConstInt(c.ctx.Int8Type(), valueSize, false),            // valueSize
			llvm.ConstInt(c.ctx.Int8Type(), 0, false),                    // bucketBits
		})

		// Create a pointer to this hashmap.
		hashmapPtr := llvm.AddGlobal(c.mod, hashmap.Type(), prefix+"$hashmap")
		hashmapPtr.SetInitializer(hashmap)
		hashmapPtr.SetLinkage(llvm.InternalLinkage)
		return llvm.ConstInBoundsGEP(hashmapPtr, []llvm.Value{zero}), nil

	case *ir.PointerBitCastValue:
		elem, err := c.getInterpretedValue(prefix, value.Elem)
		if err != nil {
			return llvm.Value{}, err
		}
		llvmType, err := c.getLLVMType(value.Type)
		if err != nil {
			return llvm.Value{}, err
		}
		return llvm.ConstBitCast(elem, llvmType), nil

	case *ir.PointerToUintptrValue:
		elem, err := c.getInterpretedValue(prefix, value.Elem)
		if err != nil {
			return llvm.Value{}, err
		}
		return llvm.ConstPtrToInt(elem, c.uintptrType), nil

	case *ir.PointerValue:
		if value.Elem == nil {
			typ, err := c.getLLVMType(value.Type)
			if err != nil {
				return llvm.Value{}, err
			}
			return llvm.ConstPointerNull(typ), nil
		}
		elem, err := c.getInterpretedValue(prefix, *value.Elem)
		if err != nil {
			return llvm.Value{}, err
		}

		obj := llvm.AddGlobal(c.mod, elem.Type(), prefix+"$ptrvalue")
		obj.SetInitializer(elem)
		obj.SetLinkage(llvm.InternalLinkage)
		elem = obj

		zero := llvm.ConstInt(c.ctx.Int32Type(), 0, false)
		ptr := llvm.ConstInBoundsGEP(elem, []llvm.Value{zero})
		return ptr, nil

	case *ir.SliceValue:
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
			array, err := c.getInterpretedValue(prefix, value.Array)
			if err != nil {
				return llvm.Value{}, err
			}
			// make global from array
			global := llvm.AddGlobal(c.mod, array.Type(), prefix+"$array")
			global.SetInitializer(array)
			global.SetLinkage(llvm.InternalLinkage)

			// get pointer to global
			zero := llvm.ConstInt(c.ctx.Int32Type(), 0, false)
			globalPtr = c.builder.CreateInBoundsGEP(global, []llvm.Value{zero, zero}, "")

			arrayLength = uint64(len(value.Array.Elems))
		}

		// make slice
		sliceTyp, err := c.getLLVMType(value.Type)
		if err != nil {
			return llvm.Value{}, err
		}
		llvmLen := llvm.ConstInt(c.uintptrType, arrayLength, false)
		slice := llvm.ConstNamedStruct(sliceTyp, []llvm.Value{
			globalPtr, // ptr
			llvmLen,   // len
			llvmLen,   // cap
		})
		return slice, nil

	case *ir.StructValue:
		fields := make([]llvm.Value, len(value.Fields))
		for i, elem := range value.Fields {
			field, err := c.getInterpretedValue(prefix, elem)
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
			return c.ctx.ConstStruct(fields, false), nil
		default:
			return llvm.Value{}, errors.New("init: unknown struct type: " + value.Type.String())
		}

	case *ir.ZeroBasicValue:
		llvmType, err := c.getLLVMType(value.Type)
		if err != nil {
			return llvm.Value{}, err
		}
		return c.getZeroValue(llvmType)

	default:
		return llvm.Value{}, errors.New("init: unknown initializer type: " + fmt.Sprintf("%#v", value))
	}
}

func (c *Compiler) parseFunc(frame *Frame) error {
	if c.DumpSSA {
		fmt.Printf("\nfunc %s:\n", frame.fn.Function)
	}
	if !frame.fn.IsExported() {
		frame.fn.LLVMFn.SetLinkage(llvm.InternalLinkage)
		frame.fn.LLVMFn.SetUnnamedAddr(true)
	}
	if frame.fn.IsInterrupt() && strings.HasPrefix(c.Triple, "avr") {
		frame.fn.LLVMFn.SetFunctionCallConv(85) // CallingConv::AVR_SIGNAL
	}

	// Add debug info, if needed.
	if c.Debug {
		if frame.fn.Synthetic == "package initializer" {
			// Package initializers have no debug info. Create some fake debug
			// info to at least have *something*.
			difunc, err := c.attachDebugInfoRaw(frame.fn, frame.fn.LLVMFn, "", "", 0)
			if err != nil {
				return err
			}
			frame.difunc = difunc
		} else if frame.fn.Syntax() != nil {
			// Create debug info file if needed.
			difunc, err := c.attachDebugInfo(frame.fn)
			if err != nil {
				return err
			}
			frame.difunc = difunc
		}
		pos := c.ir.Program.Fset.Position(frame.fn.Pos())
		c.builder.SetCurrentDebugLocation(uint(pos.Line), uint(pos.Column), frame.difunc, llvm.Metadata{})
	}

	// Pre-create all basic blocks in the function.
	for _, block := range frame.fn.DomPreorder() {
		llvmBlock := c.ctx.AddBasicBlock(frame.fn.LLVMFn, block.Comment)
		frame.blockEntries[block] = llvmBlock
		frame.blockExits[block] = llvmBlock
	}
	entryBlock := frame.blockEntries[frame.fn.Blocks[0]]
	c.builder.SetInsertPointAtEnd(entryBlock)

	// Load function parameters
	llvmParamIndex := 0
	for i, param := range frame.fn.Params {
		llvmType, err := c.getLLVMType(param.Type())
		if err != nil {
			return err
		}
		fields := make([]llvm.Value, 0, 1)
		for range c.expandFormalParamType(llvmType) {
			fields = append(fields, frame.fn.LLVMFn.Param(llvmParamIndex))
			llvmParamIndex++
		}
		frame.locals[param] = c.collapseFormalParam(llvmType, fields)

		// Add debug information to this parameter (if available)
		if c.Debug && frame.fn.Syntax() != nil {
			pos := c.ir.Program.Fset.Position(frame.fn.Syntax().Pos())
			dityp, err := c.getDIType(param.Type())
			if err != nil {
				return err
			}
			c.dibuilder.CreateParameterVariable(frame.difunc, llvm.DIParameterVariable{
				Name:           param.Name(),
				File:           c.difiles[pos.Filename],
				Line:           pos.Line,
				Type:           dityp,
				AlwaysPreserve: true,
				ArgNo:          i + 1,
			})
			// TODO: set the value of this parameter.
		}
	}

	// Load free variables from the context. This is a closure (or bound
	// method).
	var context llvm.Value
	if !frame.fn.IsExported() {
		parentHandle := frame.fn.LLVMFn.LastParam()
		parentHandle.SetName("parentHandle")
		context = llvm.PrevParam(parentHandle)
		context.SetName("context")
	}
	if len(frame.fn.FreeVars) != 0 {
		// Determine the context type. It's a struct containing all variables.
		freeVarTypes := make([]llvm.Type, 0, len(frame.fn.FreeVars))
		for _, freeVar := range frame.fn.FreeVars {
			typ, err := c.getLLVMType(freeVar.Type())
			if err != nil {
				return err
			}
			freeVarTypes = append(freeVarTypes, typ)
		}
		contextType := c.ctx.StructType(freeVarTypes, false)

		// Get a correctly-typed pointer to the context.
		contextAlloc := llvm.Value{}
		if c.targetData.TypeAllocSize(contextType) <= c.targetData.TypeAllocSize(c.i8ptrType) {
			// Context stored directly in pointer. Load it using an alloca.
			contextRawAlloc := c.builder.CreateAlloca(llvm.PointerType(c.i8ptrType, 0), "context.raw.alloc")
			contextRawValue := c.builder.CreateBitCast(context, llvm.PointerType(c.i8ptrType, 0), "context.raw.value")
			c.builder.CreateStore(contextRawValue, contextRawAlloc)
			contextAlloc = c.builder.CreateBitCast(contextRawAlloc, llvm.PointerType(contextType, 0), "context.alloc")
		} else {
			// Context stored in the heap. Bitcast the passed-in pointer to the
			// correct pointer type.
			contextAlloc = c.builder.CreateBitCast(context, llvm.PointerType(contextType, 0), "context.raw.ptr")
		}

		// Load each free variable from the context.
		// A free variable is always a pointer when this is a closure, but it
		// can be another type when it is a wrapper for a bound method (these
		// wrappers are generated by the ssa package).
		for i, freeVar := range frame.fn.FreeVars {
			indices := []llvm.Value{
				llvm.ConstInt(c.ctx.Int32Type(), 0, false),
				llvm.ConstInt(c.ctx.Int32Type(), uint64(i), false),
			}
			gep := c.builder.CreateInBoundsGEP(contextAlloc, indices, "")
			frame.locals[freeVar] = c.builder.CreateLoad(gep, "")
		}
	}

	if frame.fn.Recover != nil {
		// This function has deferred function calls. Set some things up for
		// them.
		c.deferInitFunc(frame)
	}

	// Fill blocks with instructions.
	for _, block := range frame.fn.DomPreorder() {
		if c.DumpSSA {
			fmt.Printf("%d: %s:\n", block.Index, block.Comment)
		}
		c.builder.SetInsertPointAtEnd(frame.blockEntries[block])
		frame.currentBlock = block
		for _, instr := range block.Instrs {
			if _, ok := instr.(*ssa.DebugRef); ok {
				continue
			}
			if c.DumpSSA {
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
		if frame.fn.Name() == "init" && len(block.Instrs) == 0 {
			c.builder.CreateRetVoid()
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
			llvmBlock := frame.blockExits[block.Preds[i]]
			phi.llvm.AddIncoming([]llvm.Value{llvmVal}, []llvm.BasicBlock{llvmBlock})
		}
	}

	return nil
}

func (c *Compiler) parseInstr(frame *Frame, instr ssa.Instruction) error {
	if c.Debug {
		pos := c.ir.Program.Fset.Position(instr.Pos())
		c.builder.SetCurrentDebugLocation(uint(pos.Line), uint(pos.Column), frame.difunc, llvm.Metadata{})
	}

	switch instr := instr.(type) {
	case ssa.Value:
		value, err := c.parseExpr(frame, instr)
		if err == ir.ErrCGoWrapper {
			// Ignore CGo global variables which we don't use.
			return nil
		}
		frame.locals[instr] = value
		return err
	case *ssa.DebugRef:
		return nil // ignore
	case *ssa.Defer:
		return c.emitDefer(frame, instr)
	case *ssa.Go:
		if instr.Call.IsInvoke() {
			return c.makeError(instr.Pos(), "todo: go on method receiver")
		}
		callee := instr.Call.StaticCallee()
		if callee == nil {
			return c.makeError(instr.Pos(), "todo: go on non-direct function (function pointer, etc.)")
		}
		calleeFn := c.ir.GetFunction(callee)

		// Mark this function as a 'go' invocation and break invalid
		// interprocedural optimizations. For example, heap-to-stack
		// transformations are not sound as goroutines can outlive their parent.
		calleeType := calleeFn.LLVMFn.Type()
		calleeValue := c.builder.CreateBitCast(calleeFn.LLVMFn, c.i8ptrType, "")
		calleeValue = c.createRuntimeCall("makeGoroutine", []llvm.Value{calleeValue}, "")
		calleeValue = c.builder.CreateBitCast(calleeValue, calleeType, "")

		// Get all function parameters to pass to the goroutine.
		var params []llvm.Value
		for _, param := range instr.Call.Args {
			val, err := c.parseExpr(frame, param)
			if err != nil {
				return err
			}
			params = append(params, val)
		}
		if !calleeFn.IsExported() {
			params = append(params, llvm.Undef(c.i8ptrType)) // context parameter
			params = append(params, llvm.Undef(c.i8ptrType)) // parent coroutine handle
		}

		c.createCall(calleeValue, params, "")
		return nil
	case *ssa.If:
		cond, err := c.parseExpr(frame, instr.Cond)
		if err != nil {
			return err
		}
		block := instr.Block()
		blockThen := frame.blockEntries[block.Succs[0]]
		blockElse := frame.blockEntries[block.Succs[1]]
		c.builder.CreateCondBr(cond, blockThen, blockElse)
		return nil
	case *ssa.Jump:
		blockJump := frame.blockEntries[instr.Block().Succs[0]]
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
		return c.emitMapUpdate(mapType.Key(), m, key, value, instr.Pos())
	case *ssa.Panic:
		value, err := c.parseExpr(frame, instr.X)
		if err != nil {
			return err
		}
		c.createRuntimeCall("_panic", []llvm.Value{value}, "")
		c.builder.CreateUnreachable()
		return nil
	case *ssa.Return:
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
			retVal, err := c.getZeroValue(frame.fn.LLVMFn.Type().ElementType().ReturnType())
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
	case *ssa.RunDefers:
		return c.emitRunDefers(frame)
	case *ssa.Send:
		return c.emitChanSend(frame, instr)
	case *ssa.Store:
		llvmAddr, err := c.parseExpr(frame, instr.Addr)
		if err == ir.ErrCGoWrapper {
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
		if c.targetData.TypeAllocSize(llvmVal.Type()) == 0 {
			// nothing to store
			return nil
		}
		store := c.builder.CreateStore(llvmVal, llvmAddr)
		valType := instr.Addr.Type().(*types.Pointer).Elem()
		if c.ir.IsVolatile(valType) {
			// Volatile store, for memory-mapped registers.
			store.SetVolatile(true)
		}
		return nil
	default:
		return c.makeError(instr.Pos(), "unknown instruction: "+instr.String())
	}
}

func (c *Compiler) parseBuiltin(frame *Frame, args []ssa.Value, callName string, pos token.Pos) (llvm.Value, error) {
	switch callName {
	case "append":
		src, err := c.parseExpr(frame, args[0])
		if err != nil {
			return llvm.Value{}, err
		}
		elems, err := c.parseExpr(frame, args[1])
		if err != nil {
			return llvm.Value{}, err
		}
		srcBuf := c.builder.CreateExtractValue(src, 0, "append.srcBuf")
		srcPtr := c.builder.CreateBitCast(srcBuf, c.i8ptrType, "append.srcPtr")
		srcLen := c.builder.CreateExtractValue(src, 1, "append.srcLen")
		srcCap := c.builder.CreateExtractValue(src, 2, "append.srcCap")
		elemsBuf := c.builder.CreateExtractValue(elems, 0, "append.elemsBuf")
		elemsPtr := c.builder.CreateBitCast(elemsBuf, c.i8ptrType, "append.srcPtr")
		elemsLen := c.builder.CreateExtractValue(elems, 1, "append.elemsLen")
		elemType := srcBuf.Type().ElementType()
		elemSize := llvm.ConstInt(c.uintptrType, c.targetData.TypeAllocSize(elemType), false)
		result := c.createRuntimeCall("sliceAppend", []llvm.Value{srcPtr, elemsPtr, srcLen, srcCap, elemsLen, elemSize}, "append.new")
		newPtr := c.builder.CreateExtractValue(result, 0, "append.newPtr")
		newBuf := c.builder.CreateBitCast(newPtr, srcBuf.Type(), "append.newBuf")
		newLen := c.builder.CreateExtractValue(result, 1, "append.newLen")
		newCap := c.builder.CreateExtractValue(result, 2, "append.newCap")
		newSlice := llvm.Undef(src.Type())
		newSlice = c.builder.CreateInsertValue(newSlice, newBuf, 0, "")
		newSlice = c.builder.CreateInsertValue(newSlice, newLen, 1, "")
		newSlice = c.builder.CreateInsertValue(newSlice, newCap, 2, "")
		return newSlice, nil
	case "cap":
		value, err := c.parseExpr(frame, args[0])
		if err != nil {
			return llvm.Value{}, err
		}
		switch args[0].Type().(type) {
		case *types.Chan:
			// Channel. Buffered channels haven't been implemented yet so always
			// return 0.
			return llvm.ConstInt(c.intType, 0, false), nil
		case *types.Slice:
			return c.builder.CreateExtractValue(value, 2, "cap"), nil
		default:
			return llvm.Value{}, c.makeError(pos, "todo: cap: unknown type")
		}
	case "close":
		return llvm.Value{}, c.emitChanClose(frame, args[0])
	case "complex":
		r, err := c.parseExpr(frame, args[0])
		if err != nil {
			return llvm.Value{}, err
		}
		i, err := c.parseExpr(frame, args[1])
		if err != nil {
			return llvm.Value{}, err
		}
		t := args[0].Type().Underlying().(*types.Basic)
		var cplx llvm.Value
		switch t.Kind() {
		case types.Float32:
			cplx = llvm.Undef(llvm.VectorType(c.ctx.FloatType(), 2))
		case types.Float64:
			cplx = llvm.Undef(llvm.VectorType(c.ctx.DoubleType(), 2))
		default:
			return llvm.Value{}, c.makeError(pos, "unsupported type in complex builtin: "+t.String())
		}
		cplx = c.builder.CreateInsertElement(cplx, r, llvm.ConstInt(c.ctx.Int8Type(), 0, false), "")
		cplx = c.builder.CreateInsertElement(cplx, i, llvm.ConstInt(c.ctx.Int8Type(), 1, false), "")
		return cplx, nil
	case "copy":
		dst, err := c.parseExpr(frame, args[0])
		if err != nil {
			return llvm.Value{}, err
		}
		src, err := c.parseExpr(frame, args[1])
		if err != nil {
			return llvm.Value{}, err
		}
		dstLen := c.builder.CreateExtractValue(dst, 1, "copy.dstLen")
		srcLen := c.builder.CreateExtractValue(src, 1, "copy.srcLen")
		dstBuf := c.builder.CreateExtractValue(dst, 0, "copy.dstArray")
		srcBuf := c.builder.CreateExtractValue(src, 0, "copy.srcArray")
		elemType := dstBuf.Type().ElementType()
		dstBuf = c.builder.CreateBitCast(dstBuf, c.i8ptrType, "copy.dstPtr")
		srcBuf = c.builder.CreateBitCast(srcBuf, c.i8ptrType, "copy.srcPtr")
		elemSize := llvm.ConstInt(c.uintptrType, c.targetData.TypeAllocSize(elemType), false)
		return c.createRuntimeCall("sliceCopy", []llvm.Value{dstBuf, srcBuf, dstLen, srcLen, elemSize}, "copy.n"), nil
	case "delete":
		m, err := c.parseExpr(frame, args[0])
		if err != nil {
			return llvm.Value{}, err
		}
		key, err := c.parseExpr(frame, args[1])
		if err != nil {
			return llvm.Value{}, err
		}
		return llvm.Value{}, c.emitMapDelete(args[1].Type(), m, key, pos)
	case "imag":
		cplx, err := c.parseExpr(frame, args[0])
		if err != nil {
			return llvm.Value{}, err
		}
		index := llvm.ConstInt(c.ctx.Int32Type(), 1, false)
		return c.builder.CreateExtractElement(cplx, index, "imag"), nil
	case "len":
		value, err := c.parseExpr(frame, args[0])
		if err != nil {
			return llvm.Value{}, err
		}
		var llvmLen llvm.Value
		switch args[0].Type().Underlying().(type) {
		case *types.Basic, *types.Slice:
			// string or slice
			llvmLen = c.builder.CreateExtractValue(value, 1, "len")
		case *types.Chan:
			// Channel. Buffered channels haven't been implemented yet so always
			// return 0.
			llvmLen = llvm.ConstInt(c.intType, 0, false)
		case *types.Map:
			llvmLen = c.createRuntimeCall("hashmapLen", []llvm.Value{value}, "len")
		default:
			return llvm.Value{}, c.makeError(pos, "todo: len: unknown type")
		}
		if c.targetData.TypeAllocSize(llvmLen.Type()) < c.targetData.TypeAllocSize(c.intType) {
			llvmLen = c.builder.CreateZExt(llvmLen, c.intType, "len.int")
		}
		return llvmLen, nil
	case "print", "println":
		for i, arg := range args {
			if i >= 1 && callName == "println" {
				c.createRuntimeCall("printspace", nil, "")
			}
			value, err := c.parseExpr(frame, arg)
			if err != nil {
				return llvm.Value{}, err
			}
			typ := arg.Type().Underlying()
			switch typ := typ.(type) {
			case *types.Basic:
				switch typ.Kind() {
				case types.String, types.UntypedString:
					c.createRuntimeCall("printstring", []llvm.Value{value}, "")
				case types.Uintptr:
					c.createRuntimeCall("printptr", []llvm.Value{value}, "")
				case types.UnsafePointer:
					ptrValue := c.builder.CreatePtrToInt(value, c.uintptrType, "")
					c.createRuntimeCall("printptr", []llvm.Value{ptrValue}, "")
				default:
					// runtime.print{int,uint}{8,16,32,64}
					if typ.Info()&types.IsInteger != 0 {
						name := "print"
						if typ.Info()&types.IsUnsigned != 0 {
							name += "uint"
						} else {
							name += "int"
						}
						name += strconv.FormatUint(c.targetData.TypeAllocSize(value.Type())*8, 10)
						c.createRuntimeCall(name, []llvm.Value{value}, "")
					} else if typ.Kind() == types.Bool {
						c.createRuntimeCall("printbool", []llvm.Value{value}, "")
					} else if typ.Kind() == types.Float32 {
						c.createRuntimeCall("printfloat32", []llvm.Value{value}, "")
					} else if typ.Kind() == types.Float64 {
						c.createRuntimeCall("printfloat64", []llvm.Value{value}, "")
					} else if typ.Kind() == types.Complex64 {
						c.createRuntimeCall("printcomplex64", []llvm.Value{value}, "")
					} else if typ.Kind() == types.Complex128 {
						c.createRuntimeCall("printcomplex128", []llvm.Value{value}, "")
					} else {
						return llvm.Value{}, c.makeError(pos, "unknown basic arg type: "+typ.String())
					}
				}
			case *types.Interface:
				c.createRuntimeCall("printitf", []llvm.Value{value}, "")
			case *types.Map:
				c.createRuntimeCall("printmap", []llvm.Value{value}, "")
			case *types.Pointer:
				ptrValue := c.builder.CreatePtrToInt(value, c.uintptrType, "")
				c.createRuntimeCall("printptr", []llvm.Value{ptrValue}, "")
			default:
				return llvm.Value{}, c.makeError(pos, "unknown arg type: "+typ.String())
			}
		}
		if callName == "println" {
			c.createRuntimeCall("printnl", nil, "")
		}
		return llvm.Value{}, nil // print() or println() returns void
	case "real":
		cplx, err := c.parseExpr(frame, args[0])
		if err != nil {
			return llvm.Value{}, err
		}
		index := llvm.ConstInt(c.ctx.Int32Type(), 0, false)
		return c.builder.CreateExtractElement(cplx, index, "real"), nil
	case "recover":
		return c.createRuntimeCall("_recover", nil, ""), nil
	case "ssa:wrapnilchk":
		// TODO: do an actual nil check?
		return c.parseExpr(frame, args[0])
	default:
		return llvm.Value{}, c.makeError(pos, "todo: builtin: "+callName)
	}
}

func (c *Compiler) parseFunctionCall(frame *Frame, args []ssa.Value, llvmFn, context llvm.Value, exported bool) (llvm.Value, error) {
	var params []llvm.Value
	for _, param := range args {
		val, err := c.parseExpr(frame, param)
		if err != nil {
			return llvm.Value{}, err
		}
		params = append(params, val)
	}

	if !exported {
		// This function takes a context parameter.
		// Add it to the end of the parameter list.
		params = append(params, context)

		// Parent coroutine handle.
		params = append(params, llvm.Undef(c.i8ptrType))
	}

	return c.createCall(llvmFn, params, ""), nil
}

func (c *Compiler) parseCall(frame *Frame, instr *ssa.CallCommon) (llvm.Value, error) {
	if instr.IsInvoke() {
		fnCast, args, err := c.getInvokeCall(frame, instr)
		if err != nil {
			return llvm.Value{}, err
		}
		return c.createCall(fnCast, args, ""), nil
	}

	// Try to call the function directly for trivially static calls.
	if fn := instr.StaticCallee(); fn != nil {
		if fn.RelString(nil) == "device/arm.Asm" || fn.RelString(nil) == "device/avr.Asm" {
			// Magic function: insert inline assembly instead of calling it.
			fnType := llvm.FunctionType(c.ctx.VoidType(), []llvm.Type{}, false)
			asm := constant.StringVal(instr.Args[0].(*ssa.Const).Value)
			target := llvm.InlineAsm(fnType, asm, "", true, false, 0)
			return c.builder.CreateCall(target, nil, ""), nil
		}

		if fn.RelString(nil) == "device/arm.ReadRegister" {
			// Magic function: return the given register.
			fnType := llvm.FunctionType(c.uintptrType, []llvm.Type{}, false)
			regname := constant.StringVal(instr.Args[0].(*ssa.Const).Value)
			target := llvm.InlineAsm(fnType, "mov $0, "+regname, "=r", false, false, 0)
			return c.builder.CreateCall(target, nil, ""), nil
		}

		if strings.HasPrefix(fn.RelString(nil), "device/arm.SVCall") {
			// Magic function: inline this call as a SVC instruction.
			num, _ := constant.Uint64Val(instr.Args[0].(*ssa.Const).Value)
			args := []llvm.Value{}
			argTypes := []llvm.Type{}
			asm := "svc #" + strconv.FormatUint(num, 10)
			constraints := "={r0}"
			for i, arg := range instr.Args[1:] {
				arg = arg.(*ssa.MakeInterface).X
				if i == 0 {
					constraints += ",0"
				} else {
					constraints += ",{r" + strconv.Itoa(i) + "}"
				}
				llvmValue, err := c.parseExpr(frame, arg)
				if err != nil {
					return llvm.Value{}, err
				}
				args = append(args, llvmValue)
				argTypes = append(argTypes, llvmValue.Type())
			}
			// Implement the ARM calling convention by marking r1-r3 as
			// clobbered. r0 is used as an output register so doesn't have to be
			// marked as clobbered.
			constraints += ",~{r1},~{r2},~{r3}"
			fnType := llvm.FunctionType(c.uintptrType, argTypes, false)
			target := llvm.InlineAsm(fnType, asm, constraints, true, false, 0)
			return c.builder.CreateCall(target, args, ""), nil
		}

		if fn.RelString(nil) == "device/arm.AsmFull" || fn.RelString(nil) == "device/avr.AsmFull" {
			asmString := constant.StringVal(instr.Args[0].(*ssa.Const).Value)
			registers := map[string]llvm.Value{}
			registerMap := instr.Args[1].(*ssa.MakeMap)
			for _, r := range *registerMap.Referrers() {
				switch r := r.(type) {
				case *ssa.DebugRef:
					// ignore
				case *ssa.MapUpdate:
					if r.Block() != registerMap.Block() {
						return llvm.Value{}, c.makeError(instr.Pos(), "register value map must be created in the same basic block")
					}
					key := constant.StringVal(r.Key.(*ssa.Const).Value)
					//println("value:", r.Value.(*ssa.MakeInterface).X.String())
					value, err := c.parseExpr(frame, r.Value.(*ssa.MakeInterface).X)
					if err != nil {
						return llvm.Value{}, err
					}
					registers[key] = value
				case *ssa.Call:
					if r.Common() == instr {
						break
					}
				default:
					return llvm.Value{}, c.makeError(instr.Pos(), "don't know how to handle argument to inline assembly: "+r.String())
				}
			}
			// TODO: handle dollar signs in asm string
			registerNumbers := map[string]int{}
			var err error
			argTypes := []llvm.Type{}
			args := []llvm.Value{}
			constraints := []string{}
			asmString = regexp.MustCompile("\\{[a-zA-Z]+\\}").ReplaceAllStringFunc(asmString, func(s string) string {
				// TODO: skip strings like {r4} etc. that look like ARM push/pop
				// instructions.
				name := s[1 : len(s)-1]
				if _, ok := registers[name]; !ok {
					if err == nil {
						err = c.makeError(instr.Pos(), "unknown register name: "+name)
					}
					return s
				}
				if _, ok := registerNumbers[name]; !ok {
					registerNumbers[name] = len(registerNumbers)
					argTypes = append(argTypes, registers[name].Type())
					args = append(args, registers[name])
					switch registers[name].Type().TypeKind() {
					case llvm.IntegerTypeKind:
						constraints = append(constraints, "r")
					case llvm.PointerTypeKind:
						constraints = append(constraints, "*m")
					default:
						err = c.makeError(instr.Pos(), "unknown type in inline assembly for value: "+name)
						return s
					}
				}
				return fmt.Sprintf("${%v}", registerNumbers[name])
			})
			if err != nil {
				return llvm.Value{}, err
			}
			fnType := llvm.FunctionType(c.ctx.VoidType(), argTypes, false)
			target := llvm.InlineAsm(fnType, asmString, strings.Join(constraints, ","), true, false, 0)
			return c.builder.CreateCall(target, args, ""), nil
		}

		targetFunc := c.ir.GetFunction(fn)
		if targetFunc.LLVMFn.IsNil() {
			return llvm.Value{}, c.makeError(instr.Pos(), "undefined function: "+targetFunc.LinkName())
		}
		var context llvm.Value
		// This function call is to a (potential) closure, not a regular
		// function. See whether it is a closure and if so, call it as such.
		// Else, supply a dummy nil pointer as the last parameter.
		if targetFunc.IsExported() {
			// don't pass a context parameter
		} else if mkClosure, ok := instr.Value.(*ssa.MakeClosure); ok {
			// closure is {context, function pointer}
			closure, err := c.parseExpr(frame, mkClosure)
			if err != nil {
				return llvm.Value{}, err
			}
			context = c.builder.CreateExtractValue(closure, 0, "")
		} else {
			context = llvm.Undef(c.i8ptrType)
		}
		return c.parseFunctionCall(frame, instr.Args, targetFunc.LLVMFn, context, targetFunc.IsExported())
	}

	// Builtin or function pointer.
	switch call := instr.Value.(type) {
	case *ssa.Builtin:
		return c.parseBuiltin(frame, instr.Args, call.Name(), instr.Pos())
	default: // function pointer
		value, err := c.parseExpr(frame, instr.Value)
		if err != nil {
			return llvm.Value{}, err
		}
		// 'value' is a closure, not a raw function pointer.
		// Extract the function pointer and the context pointer.
		// closure: {context, function pointer}
		context := c.builder.CreateExtractValue(value, 0, "")
		value = c.builder.CreateExtractValue(value, 1, "")
		return c.parseFunctionCall(frame, instr.Args, value, context, false)
	}
}

func (c *Compiler) emitBoundsCheck(frame *Frame, arrayLen, index llvm.Value, indexType types.Type) {
	if frame.fn.IsNoBounds() {
		// The //go:nobounds pragma was added to the function to avoid bounds
		// checking.
		return
	}

	// Sometimes, the index can be e.g. an uint8 or int8, and we have to
	// correctly extend that type.
	if index.Type().IntTypeWidth() < arrayLen.Type().IntTypeWidth() {
		if indexType.(*types.Basic).Info()&types.IsUnsigned == 0 {
			index = c.builder.CreateZExt(index, arrayLen.Type(), "")
		} else {
			index = c.builder.CreateSExt(index, arrayLen.Type(), "")
		}
	}

	// Optimize away trivial cases.
	// LLVM would do this anyway with interprocedural optimizations, but it
	// helps to see cases where bounds check elimination would really help.
	if index.IsConstant() && arrayLen.IsConstant() && !arrayLen.IsUndef() {
		index := index.SExtValue()
		arrayLen := arrayLen.SExtValue()
		if index >= 0 && index < arrayLen {
			return
		}
	}

	if index.Type().IntTypeWidth() > c.intType.IntTypeWidth() {
		// Index is too big for the regular bounds check. Use the one for int64.
		c.createRuntimeCall("lookupBoundsCheckLong", []llvm.Value{arrayLen, index}, "")
	} else {
		c.createRuntimeCall("lookupBoundsCheck", []llvm.Value{arrayLen, index}, "")
	}
}

func (c *Compiler) emitSliceBoundsCheck(frame *Frame, capacity, low, high llvm.Value, lowType, highType *types.Basic) {
	if frame.fn.IsNoBounds() {
		// The //go:nobounds pragma was added to the function to avoid bounds
		// checking.
		return
	}

	uintptrWidth := c.uintptrType.IntTypeWidth()
	if low.Type().IntTypeWidth() > uintptrWidth || high.Type().IntTypeWidth() > uintptrWidth {
		if low.Type().IntTypeWidth() < 64 {
			if lowType.Info()&types.IsUnsigned != 0 {
				low = c.builder.CreateZExt(low, c.ctx.Int64Type(), "")
			} else {
				low = c.builder.CreateSExt(low, c.ctx.Int64Type(), "")
			}
		}
		if high.Type().IntTypeWidth() < 64 {
			if highType.Info()&types.IsUnsigned != 0 {
				high = c.builder.CreateZExt(high, c.ctx.Int64Type(), "")
			} else {
				high = c.builder.CreateSExt(high, c.ctx.Int64Type(), "")
			}
		}
		// TODO: 32-bit or even 16-bit slice bounds checks for 8-bit platforms
		c.createRuntimeCall("sliceBoundsCheck64", []llvm.Value{capacity, low, high}, "")
	} else {
		c.createRuntimeCall("sliceBoundsCheck", []llvm.Value{capacity, low, high}, "")
	}
}

func (c *Compiler) parseExpr(frame *Frame, expr ssa.Value) (llvm.Value, error) {
	if value, ok := frame.locals[expr]; ok {
		// Value is a local variable that has already been computed.
		if value.IsNil() {
			return llvm.Value{}, c.makeError(expr.Pos(), "undefined local var (from cgo?)")
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
			buf = c.createRuntimeCall("alloc", []llvm.Value{size}, expr.Comment)
			buf = c.builder.CreateBitCast(buf, llvm.PointerType(typ, 0), "")
		} else {
			buf = c.builder.CreateAlloca(typ, expr.Comment)
			if c.targetData.TypeAllocSize(typ) != 0 {
				zero, err := c.getZeroValue(typ)
				if err != nil {
					return llvm.Value{}, err
				}
				c.builder.CreateStore(zero, buf) // zero-initialize var
			}
		}
		return buf, nil
	case *ssa.BinOp:
		x, err := c.parseExpr(frame, expr.X)
		if err != nil {
			return llvm.Value{}, err
		}
		y, err := c.parseExpr(frame, expr.Y)
		if err != nil {
			return llvm.Value{}, err
		}
		return c.parseBinOp(expr.Op, expr.X.Type(), x, y, expr.Pos())
	case *ssa.Call:
		// Passing the current task here to the subroutine. It is only used when
		// the subroutine is blocking.
		return c.parseCall(frame, expr.Common())
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
		// The only case when we need to bitcast is when casting between named
		// struct types, as those are actually different in LLVM. Let's just
		// bitcast all struct types for ease of use.
		if _, ok := expr.Type().Underlying().(*types.Struct); ok {
			llvmType, err := c.getLLVMType(expr.X.Type())
			if err != nil {
				return llvm.Value{}, err
			}
			return c.builder.CreateBitCast(x, llvmType, "changetype"), nil
		}
		return x, nil
	case *ssa.Const:
		return c.parseConst(frame.fn.LinkName(), expr)
	case *ssa.Convert:
		x, err := c.parseExpr(frame, expr.X)
		if err != nil {
			return llvm.Value{}, err
		}
		return c.parseConvert(expr.X.Type(), expr.Type(), x, expr.Pos())
	case *ssa.Extract:
		value, err := c.parseExpr(frame, expr.Tuple)
		if err != nil {
			return llvm.Value{}, err
		}
		result := c.builder.CreateExtractValue(value, expr.Index, "")
		return result, nil
	case *ssa.Field:
		value, err := c.parseExpr(frame, expr.X)
		if err != nil {
			return llvm.Value{}, err
		}
		result := c.builder.CreateExtractValue(value, expr.Field, "")
		return result, nil
	case *ssa.FieldAddr:
		val, err := c.parseExpr(frame, expr.X)
		if err != nil {
			return llvm.Value{}, err
		}
		indices := []llvm.Value{
			llvm.ConstInt(c.ctx.Int32Type(), 0, false),
			llvm.ConstInt(c.ctx.Int32Type(), uint64(expr.Field), false),
		}
		return c.builder.CreateGEP(val, indices, ""), nil
	case *ssa.Function:
		fn := c.ir.GetFunction(expr)
		if fn.IsExported() {
			return llvm.Value{}, c.makeError(expr.Pos(), "cannot use an exported function as value")
		}
		// Create closure for function pointer.
		// Closure is: {context, function pointer}
		return c.ctx.ConstStruct([]llvm.Value{
			llvm.Undef(c.i8ptrType),
			fn.LLVMFn,
		}, false), nil
	case *ssa.Global:
		if strings.HasPrefix(expr.Name(), "__cgofn__cgo_") || strings.HasPrefix(expr.Name(), "_cgo_") {
			// Ignore CGo global variables which we don't use.
			return llvm.Value{}, ir.ErrCGoWrapper
		}
		value := c.ir.GetGlobal(expr).LLVMGlobal
		if value.IsNil() {
			return llvm.Value{}, c.makeError(expr.Pos(), "global not found: "+c.ir.GetGlobal(expr).LinkName())
		}
		return value, nil
	case *ssa.Index:
		array, err := c.parseExpr(frame, expr.X)
		if err != nil {
			return llvm.Value{}, err
		}
		index, err := c.parseExpr(frame, expr.Index)
		if err != nil {
			return llvm.Value{}, err
		}

		// Check bounds.
		arrayLen := expr.X.Type().(*types.Array).Len()
		arrayLenLLVM := llvm.ConstInt(c.uintptrType, uint64(arrayLen), false)
		c.emitBoundsCheck(frame, arrayLenLLVM, index, expr.Index.Type())

		// Can't load directly from array (as index is non-constant), so have to
		// do it using an alloca+gep+load.
		alloca := c.builder.CreateAlloca(array.Type(), "index.alloca")
		c.builder.CreateStore(array, alloca)
		zero := llvm.ConstInt(c.ctx.Int32Type(), 0, false)
		ptr := c.builder.CreateGEP(alloca, []llvm.Value{zero, index}, "index.gep")
		return c.builder.CreateLoad(ptr, "index.load"), nil
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
				buflen = llvm.ConstInt(c.uintptrType, uint64(typ.Len()), false)
			default:
				return llvm.Value{}, c.makeError(expr.Pos(), "todo: indexaddr: "+typ.String())
			}
		case *types.Slice:
			bufptr = c.builder.CreateExtractValue(val, 0, "indexaddr.ptr")
			buflen = c.builder.CreateExtractValue(val, 1, "indexaddr.len")
		default:
			return llvm.Value{}, c.makeError(expr.Pos(), "todo: indexaddr: "+ptrTyp.String())
		}

		// Bounds check.
		// LLVM optimizes this away in most cases.
		c.emitBoundsCheck(frame, buflen, index, expr.Index.Type())

		switch expr.X.Type().Underlying().(type) {
		case *types.Pointer:
			indices := []llvm.Value{
				llvm.ConstInt(c.ctx.Int32Type(), 0, false),
				index,
			}
			return c.builder.CreateGEP(bufptr, indices, ""), nil
		case *types.Slice:
			return c.builder.CreateGEP(bufptr, []llvm.Value{index}, ""), nil
		default:
			panic("unreachable")
		}
	case *ssa.Lookup:
		value, err := c.parseExpr(frame, expr.X)
		if err != nil {
			return llvm.Value{}, nil
		}
		index, err := c.parseExpr(frame, expr.Index)
		if err != nil {
			return llvm.Value{}, nil
		}
		switch xType := expr.X.Type().Underlying().(type) {
		case *types.Basic:
			// Value type must be a string, which is a basic type.
			if xType.Info()&types.IsString == 0 {
				panic("lookup on non-string?")
			}

			// Bounds check.
			// LLVM optimizes this away in most cases.
			length, err := c.parseBuiltin(frame, []ssa.Value{expr.X}, "len", expr.Pos())
			if err != nil {
				return llvm.Value{}, err // shouldn't happen
			}
			c.emitBoundsCheck(frame, length, index, expr.Index.Type())

			// Lookup byte
			buf := c.builder.CreateExtractValue(value, 0, "")
			bufPtr := c.builder.CreateGEP(buf, []llvm.Value{index}, "")
			return c.builder.CreateLoad(bufPtr, ""), nil
		case *types.Map:
			valueType := expr.Type()
			if expr.CommaOk {
				valueType = valueType.(*types.Tuple).At(0).Type()
			}
			return c.emitMapLookup(xType.Key(), valueType, value, index, expr.CommaOk, expr.Pos())
		default:
			panic("unknown lookup type: " + expr.String())
		}
	case *ssa.MakeChan:
		return c.emitMakeChan(expr)
	case *ssa.MakeClosure:
		// A closure returns a function pointer with context:
		// {context, fp}
		return c.parseMakeClosure(frame, expr)
	case *ssa.MakeInterface:
		val, err := c.parseExpr(frame, expr.X)
		if err != nil {
			return llvm.Value{}, err
		}
		return c.parseMakeInterface(val, expr.X.Type(), "", expr.Pos())
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
		keySize := c.targetData.TypeAllocSize(llvmKeyType)
		valueSize := c.targetData.TypeAllocSize(llvmValueType)
		llvmKeySize := llvm.ConstInt(c.ctx.Int8Type(), keySize, false)
		llvmValueSize := llvm.ConstInt(c.ctx.Int8Type(), valueSize, false)
		hashmap := c.createRuntimeCall("hashmapMake", []llvm.Value{llvmKeySize, llvmValueSize}, "")
		return hashmap, nil
	case *ssa.MakeSlice:
		sliceLen, err := c.parseExpr(frame, expr.Len)
		if err != nil {
			return llvm.Value{}, nil
		}
		sliceCap, err := c.parseExpr(frame, expr.Cap)
		if err != nil {
			return llvm.Value{}, nil
		}
		sliceType := expr.Type().Underlying().(*types.Slice)
		llvmElemType, err := c.getLLVMType(sliceType.Elem())
		if err != nil {
			return llvm.Value{}, nil
		}
		elemSize := c.targetData.TypeAllocSize(llvmElemType)

		// Bounds checking.
		if !frame.fn.IsNoBounds() {
			if sliceLen.Type().IntTypeWidth() < c.uintptrType.IntTypeWidth() {
				if expr.Len.Type().(*types.Basic).Info()&types.IsUnsigned != 0 {
					sliceLen = c.builder.CreateZExt(sliceLen, c.uintptrType, "")
				} else {
					sliceLen = c.builder.CreateSExt(sliceLen, c.uintptrType, "")
				}
			}
			if sliceCap.Type().IntTypeWidth() < c.uintptrType.IntTypeWidth() {
				if expr.Cap.Type().(*types.Basic).Info()&types.IsUnsigned != 0 {
					sliceCap = c.builder.CreateZExt(sliceCap, c.uintptrType, "")
				} else {
					sliceCap = c.builder.CreateSExt(sliceCap, c.uintptrType, "")
				}
			}
			c.createRuntimeCall("sliceBoundsCheckMake", []llvm.Value{sliceLen, sliceCap}, "")
		}

		// Allocate the backing array.
		// TODO: escape analysis
		elemSizeValue := llvm.ConstInt(c.uintptrType, elemSize, false)
		sliceCapCast, err := c.parseConvert(expr.Cap.Type(), types.Typ[types.Uintptr], sliceCap, expr.Pos())
		if err != nil {
			return llvm.Value{}, err
		}
		sliceSize := c.builder.CreateBinOp(llvm.Mul, elemSizeValue, sliceCapCast, "makeslice.cap")
		slicePtr := c.createRuntimeCall("alloc", []llvm.Value{sliceSize}, "makeslice.buf")
		slicePtr = c.builder.CreateBitCast(slicePtr, llvm.PointerType(llvmElemType, 0), "makeslice.array")

		if c.targetData.TypeAllocSize(sliceLen.Type()) > c.targetData.TypeAllocSize(c.uintptrType) {
			sliceLen = c.builder.CreateTrunc(sliceLen, c.uintptrType, "")
			sliceCap = c.builder.CreateTrunc(sliceCap, c.uintptrType, "")
		}

		// Create the slice.
		slice := c.ctx.ConstStruct([]llvm.Value{
			llvm.Undef(slicePtr.Type()),
			llvm.Undef(c.uintptrType),
			llvm.Undef(c.uintptrType),
		}, false)
		slice = c.builder.CreateInsertValue(slice, slicePtr, 0, "")
		slice = c.builder.CreateInsertValue(slice, sliceLen, 1, "")
		slice = c.builder.CreateInsertValue(slice, sliceCap, 2, "")
		return slice, nil
	case *ssa.Next:
		rangeVal := expr.Iter.(*ssa.Range).X
		llvmRangeVal, err := c.parseExpr(frame, rangeVal)
		if err != nil {
			return llvm.Value{}, err
		}
		it, err := c.parseExpr(frame, expr.Iter)
		if err != nil {
			return llvm.Value{}, err
		}
		if expr.IsString {
			return c.createRuntimeCall("stringNext", []llvm.Value{llvmRangeVal, it}, "range.next"), nil
		} else { // map
			llvmKeyType, err := c.getLLVMType(rangeVal.Type().Underlying().(*types.Map).Key())
			if err != nil {
				return llvm.Value{}, err
			}
			llvmValueType, err := c.getLLVMType(rangeVal.Type().Underlying().(*types.Map).Elem())
			if err != nil {
				return llvm.Value{}, err
			}

			mapKeyAlloca := c.builder.CreateAlloca(llvmKeyType, "range.key")
			mapKeyPtr := c.builder.CreateBitCast(mapKeyAlloca, c.i8ptrType, "range.keyptr")
			mapValueAlloca := c.builder.CreateAlloca(llvmValueType, "range.value")
			mapValuePtr := c.builder.CreateBitCast(mapValueAlloca, c.i8ptrType, "range.valueptr")
			ok := c.createRuntimeCall("hashmapNext", []llvm.Value{llvmRangeVal, it, mapKeyPtr, mapValuePtr}, "range.next")

			tuple := llvm.Undef(c.ctx.StructType([]llvm.Type{c.ctx.Int1Type(), llvmKeyType, llvmValueType}, false))
			tuple = c.builder.CreateInsertValue(tuple, ok, 0, "")
			tuple = c.builder.CreateInsertValue(tuple, c.builder.CreateLoad(mapKeyAlloca, ""), 1, "")
			tuple = c.builder.CreateInsertValue(tuple, c.builder.CreateLoad(mapValueAlloca, ""), 2, "")
			return tuple, nil
		}
	case *ssa.Phi:
		t, err := c.getLLVMType(expr.Type())
		if err != nil {
			return llvm.Value{}, err
		}
		phi := c.builder.CreatePHI(t, "")
		frame.phis = append(frame.phis, Phi{expr, phi})
		return phi, nil
	case *ssa.Range:
		var iteratorType llvm.Type
		switch typ := expr.X.Type().Underlying().(type) {
		case *types.Basic: // string
			iteratorType = c.mod.GetTypeByName("runtime.stringIterator")
		case *types.Map:
			iteratorType = c.mod.GetTypeByName("runtime.hashmapIterator")
		default:
			panic("unknown type in range: " + typ.String())
		}
		it := c.builder.CreateAlloca(iteratorType, "range.it")
		zero, err := c.getZeroValue(iteratorType)
		if err != nil {
			return llvm.Value{}, nil
		}
		c.builder.CreateStore(zero, it)
		return it, nil
	case *ssa.Slice:
		if expr.Max != nil {
			return llvm.Value{}, c.makeError(expr.Pos(), "todo: full slice expressions (with max): "+expr.Type().String())
		}
		value, err := c.parseExpr(frame, expr.X)
		if err != nil {
			return llvm.Value{}, err
		}

		var lowType, highType *types.Basic
		var low, high llvm.Value

		if expr.Low != nil {
			lowType = expr.Low.Type().Underlying().(*types.Basic)
			low, err = c.parseExpr(frame, expr.Low)
			if err != nil {
				return llvm.Value{}, nil
			}
			if low.Type().IntTypeWidth() < c.uintptrType.IntTypeWidth() {
				if lowType.Info()&types.IsUnsigned != 0 {
					low = c.builder.CreateZExt(low, c.uintptrType, "")
				} else {
					low = c.builder.CreateSExt(low, c.uintptrType, "")
				}
			}
		} else {
			lowType = types.Typ[types.Int]
			low = llvm.ConstInt(c.intType, 0, false)
		}

		if expr.High != nil {
			highType = expr.High.Type().Underlying().(*types.Basic)
			high, err = c.parseExpr(frame, expr.High)
			if err != nil {
				return llvm.Value{}, nil
			}
			if high.Type().IntTypeWidth() < c.uintptrType.IntTypeWidth() {
				if highType.Info()&types.IsUnsigned != 0 {
					high = c.builder.CreateZExt(high, c.uintptrType, "")
				} else {
					high = c.builder.CreateSExt(high, c.uintptrType, "")
				}
			}
		} else {
			highType = types.Typ[types.Uintptr]
		}

		switch typ := expr.X.Type().Underlying().(type) {
		case *types.Pointer: // pointer to array
			// slice an array
			length := typ.Elem().Underlying().(*types.Array).Len()
			llvmLen := llvm.ConstInt(c.uintptrType, uint64(length), false)
			if high.IsNil() {
				high = llvmLen
			}
			indices := []llvm.Value{
				llvm.ConstInt(c.ctx.Int32Type(), 0, false),
				low,
			}

			// This check is optimized away in most cases.
			c.emitSliceBoundsCheck(frame, llvmLen, low, high, lowType, highType)

			if c.targetData.TypeAllocSize(high.Type()) > c.targetData.TypeAllocSize(c.uintptrType) {
				high = c.builder.CreateTrunc(high, c.uintptrType, "")
			}
			if c.targetData.TypeAllocSize(low.Type()) > c.targetData.TypeAllocSize(c.uintptrType) {
				low = c.builder.CreateTrunc(low, c.uintptrType, "")
			}

			sliceLen := c.builder.CreateSub(high, low, "slice.len")
			slicePtr := c.builder.CreateGEP(value, indices, "slice.ptr")
			sliceCap := c.builder.CreateSub(llvmLen, low, "slice.cap")

			slice := c.ctx.ConstStruct([]llvm.Value{
				llvm.Undef(slicePtr.Type()),
				llvm.Undef(c.uintptrType),
				llvm.Undef(c.uintptrType),
			}, false)
			slice = c.builder.CreateInsertValue(slice, slicePtr, 0, "")
			slice = c.builder.CreateInsertValue(slice, sliceLen, 1, "")
			slice = c.builder.CreateInsertValue(slice, sliceCap, 2, "")
			return slice, nil

		case *types.Slice:
			// slice a slice
			oldPtr := c.builder.CreateExtractValue(value, 0, "")
			oldLen := c.builder.CreateExtractValue(value, 1, "")
			oldCap := c.builder.CreateExtractValue(value, 2, "")
			if high.IsNil() {
				high = oldLen
			}

			c.emitSliceBoundsCheck(frame, oldCap, low, high, lowType, highType)

			if c.targetData.TypeAllocSize(low.Type()) > c.targetData.TypeAllocSize(c.uintptrType) {
				low = c.builder.CreateTrunc(low, c.uintptrType, "")
			}
			if c.targetData.TypeAllocSize(high.Type()) > c.targetData.TypeAllocSize(c.uintptrType) {
				high = c.builder.CreateTrunc(high, c.uintptrType, "")
			}

			newPtr := c.builder.CreateGEP(oldPtr, []llvm.Value{low}, "")
			newLen := c.builder.CreateSub(high, low, "")
			newCap := c.builder.CreateSub(oldCap, low, "")
			slice := c.ctx.ConstStruct([]llvm.Value{
				llvm.Undef(newPtr.Type()),
				llvm.Undef(c.uintptrType),
				llvm.Undef(c.uintptrType),
			}, false)
			slice = c.builder.CreateInsertValue(slice, newPtr, 0, "")
			slice = c.builder.CreateInsertValue(slice, newLen, 1, "")
			slice = c.builder.CreateInsertValue(slice, newCap, 2, "")
			return slice, nil

		case *types.Basic:
			if typ.Info()&types.IsString == 0 {
				return llvm.Value{}, c.makeError(expr.Pos(), "unknown slice type: "+typ.String())
			}
			// slice a string
			oldPtr := c.builder.CreateExtractValue(value, 0, "")
			oldLen := c.builder.CreateExtractValue(value, 1, "")
			if high.IsNil() {
				high = oldLen
			}

			c.emitSliceBoundsCheck(frame, oldLen, low, high, lowType, highType)

			newPtr := c.builder.CreateGEP(oldPtr, []llvm.Value{low}, "")
			newLen := c.builder.CreateSub(high, low, "")
			str, err := c.getZeroValue(c.mod.GetTypeByName("runtime._string"))
			if err != nil {
				return llvm.Value{}, err
			}
			str = c.builder.CreateInsertValue(str, newPtr, 0, "")
			str = c.builder.CreateInsertValue(str, newLen, 1, "")
			return str, nil

		default:
			return llvm.Value{}, c.makeError(expr.Pos(), "unknown slice type: "+typ.String())
		}
	case *ssa.TypeAssert:
		return c.parseTypeAssert(frame, expr)
	case *ssa.UnOp:
		return c.parseUnOp(frame, expr)
	default:
		return llvm.Value{}, c.makeError(expr.Pos(), "todo: unknown expression: "+expr.String())
	}
}

func (c *Compiler) parseBinOp(op token.Token, typ types.Type, x, y llvm.Value, pos token.Pos) (llvm.Value, error) {
	switch typ := typ.Underlying().(type) {
	case *types.Basic:
		if typ.Info()&types.IsInteger != 0 {
			// Operations on integers
			signed := typ.Info()&types.IsUnsigned == 0
			switch op {
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
				} else if sizeX < sizeY {
					// What about shifting more than the integer width?
					// I'm not entirely sure what the Go spec is on that, but as
					// Intel CPUs have undefined behavior when shifting more
					// than the integer width I'm assuming it is also undefined
					// in Go.
					y = c.builder.CreateTrunc(y, x.Type(), "")
				}
				switch op {
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
			case token.EQL: // ==
				return c.builder.CreateICmp(llvm.IntEQ, x, y, ""), nil
			case token.NEQ: // !=
				return c.builder.CreateICmp(llvm.IntNE, x, y, ""), nil
			case token.AND_NOT: // &^
				// Go specific. Calculate "and not" with x & (~y)
				inv := c.builder.CreateNot(y, "") // ~y
				return c.builder.CreateAnd(x, inv, ""), nil
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
				return llvm.Value{}, c.makeError(pos, "todo: binop on integer: "+op.String())
			}
		} else if typ.Info()&types.IsFloat != 0 {
			// Operations on floats
			switch op {
			case token.ADD: // +
				return c.builder.CreateFAdd(x, y, ""), nil
			case token.SUB: // -
				return c.builder.CreateFSub(x, y, ""), nil
			case token.MUL: // *
				return c.builder.CreateFMul(x, y, ""), nil
			case token.QUO: // /
				return c.builder.CreateFDiv(x, y, ""), nil
			case token.REM: // %
				return c.builder.CreateFRem(x, y, ""), nil
			case token.EQL: // ==
				return c.builder.CreateFCmp(llvm.FloatOEQ, x, y, ""), nil
			case token.NEQ: // !=
				return c.builder.CreateFCmp(llvm.FloatONE, x, y, ""), nil
			case token.LSS: // <
				return c.builder.CreateFCmp(llvm.FloatOLT, x, y, ""), nil
			case token.LEQ: // <=
				return c.builder.CreateFCmp(llvm.FloatOLE, x, y, ""), nil
			case token.GTR: // >
				return c.builder.CreateFCmp(llvm.FloatOGT, x, y, ""), nil
			case token.GEQ: // >=
				return c.builder.CreateFCmp(llvm.FloatOGE, x, y, ""), nil
			default:
				return llvm.Value{}, c.makeError(pos, "todo: binop on float: "+op.String())
			}
		} else if typ.Info()&types.IsBoolean != 0 {
			// Operations on booleans
			switch op {
			case token.EQL: // ==
				return c.builder.CreateICmp(llvm.IntEQ, x, y, ""), nil
			case token.NEQ: // !=
				return c.builder.CreateICmp(llvm.IntNE, x, y, ""), nil
			default:
				return llvm.Value{}, c.makeError(pos, "todo: binop on boolean: "+op.String())
			}
		} else if typ.Kind() == types.UnsafePointer {
			// Operations on pointers
			switch op {
			case token.EQL: // ==
				return c.builder.CreateICmp(llvm.IntEQ, x, y, ""), nil
			case token.NEQ: // !=
				return c.builder.CreateICmp(llvm.IntNE, x, y, ""), nil
			default:
				return llvm.Value{}, c.makeError(pos, "todo: binop on pointer: "+op.String())
			}
		} else if typ.Info()&types.IsString != 0 {
			// Operations on strings
			switch op {
			case token.ADD: // +
				return c.createRuntimeCall("stringConcat", []llvm.Value{x, y}, ""), nil
			case token.EQL: // ==
				return c.createRuntimeCall("stringEqual", []llvm.Value{x, y}, ""), nil
			case token.NEQ: // !=
				result := c.createRuntimeCall("stringEqual", []llvm.Value{x, y}, "")
				return c.builder.CreateNot(result, ""), nil
			case token.LSS: // <
				return c.createRuntimeCall("stringLess", []llvm.Value{x, y}, ""), nil
			case token.LEQ: // <=
				result := c.createRuntimeCall("stringLess", []llvm.Value{y, x}, "")
				return c.builder.CreateNot(result, ""), nil
			case token.GTR: // >
				result := c.createRuntimeCall("stringLess", []llvm.Value{x, y}, "")
				return c.builder.CreateNot(result, ""), nil
			case token.GEQ: // >=
				return c.createRuntimeCall("stringLess", []llvm.Value{y, x}, ""), nil
			default:
				return llvm.Value{}, c.makeError(pos, "todo: binop on string: "+op.String())
			}
		} else {
			return llvm.Value{}, c.makeError(pos, "todo: unknown basic type in binop: "+typ.String())
		}
	case *types.Signature:
		// Extract function pointers from the function values (closures).
		// This is safe: function pointers are generally not comparable
		// against each other, only against nil. So one or both has to be
		// nil, so we can ignore the closure context.
		x = c.builder.CreateExtractValue(x, 1, "")
		y = c.builder.CreateExtractValue(y, 1, "")
		switch op {
		case token.EQL: // ==
			return c.builder.CreateICmp(llvm.IntEQ, x, y, ""), nil
		case token.NEQ: // !=
			return c.builder.CreateICmp(llvm.IntNE, x, y, ""), nil
		default:
			return llvm.Value{}, c.makeError(pos, "binop on signature: "+op.String())
		}
	case *types.Interface:
		switch op {
		case token.EQL, token.NEQ: // ==, !=
			result := c.createRuntimeCall("interfaceEqual", []llvm.Value{x, y}, "")
			if op == token.NEQ {
				result = c.builder.CreateNot(result, "")
			}
			return result, nil
		default:
			return llvm.Value{}, c.makeError(pos, "binop on interface: "+op.String())
		}
	case *types.Map, *types.Pointer:
		// Maps are in general not comparable, but can be compared against nil
		// (which is a nil pointer). This means they can be trivially compared
		// by treating them as a pointer.
		switch op {
		case token.EQL: // ==
			return c.builder.CreateICmp(llvm.IntEQ, x, y, ""), nil
		case token.NEQ: // !=
			return c.builder.CreateICmp(llvm.IntNE, x, y, ""), nil
		default:
			return llvm.Value{}, c.makeError(pos, "todo: binop on pointer: "+op.String())
		}
	case *types.Slice:
		// Slices are in general not comparable, but can be compared against
		// nil. Assume at least one of them is nil to make the code easier.
		xPtr := c.builder.CreateExtractValue(x, 0, "")
		yPtr := c.builder.CreateExtractValue(y, 0, "")
		switch op {
		case token.EQL: // ==
			return c.builder.CreateICmp(llvm.IntEQ, xPtr, yPtr, ""), nil
		case token.NEQ: // !=
			return c.builder.CreateICmp(llvm.IntNE, xPtr, yPtr, ""), nil
		default:
			return llvm.Value{}, c.makeError(pos, "todo: binop on slice: "+op.String())
		}
	case *types.Array:
		// Compare each array element and combine the result. From the spec:
		//     Array values are comparable if values of the array element type
		//     are comparable. Two array values are equal if their corresponding
		//     elements are equal.
		result := llvm.ConstInt(c.ctx.Int1Type(), 1, true)
		for i := 0; i < int(typ.Len()); i++ {
			xField := c.builder.CreateExtractValue(x, i, "")
			yField := c.builder.CreateExtractValue(y, i, "")
			fieldEqual, err := c.parseBinOp(token.EQL, typ.Elem(), xField, yField, pos)
			if err != nil {
				return llvm.Value{}, err
			}
			result = c.builder.CreateAnd(result, fieldEqual, "")
		}
		switch op {
		case token.EQL: // ==
			return result, nil
		case token.NEQ: // !=
			return c.builder.CreateNot(result, ""), nil
		default:
			return llvm.Value{}, c.makeError(pos, "unknown: binop on struct: "+op.String())
		}
		return result, nil
	case *types.Struct:
		// Compare each struct field and combine the result. From the spec:
		//     Struct values are comparable if all their fields are comparable.
		//     Two struct values are equal if their corresponding non-blank
		//     fields are equal.
		result := llvm.ConstInt(c.ctx.Int1Type(), 1, true)
		for i := 0; i < typ.NumFields(); i++ {
			if typ.Field(i).Name() == "_" {
				// skip blank fields
				continue
			}
			fieldType := typ.Field(i).Type()
			xField := c.builder.CreateExtractValue(x, i, "")
			yField := c.builder.CreateExtractValue(y, i, "")
			fieldEqual, err := c.parseBinOp(token.EQL, fieldType, xField, yField, pos)
			if err != nil {
				return llvm.Value{}, err
			}
			result = c.builder.CreateAnd(result, fieldEqual, "")
		}
		switch op {
		case token.EQL: // ==
			return result, nil
		case token.NEQ: // !=
			return c.builder.CreateNot(result, ""), nil
		default:
			return llvm.Value{}, c.makeError(pos, "unknown: binop on struct: "+op.String())
		}
		return result, nil
	default:
		return llvm.Value{}, c.makeError(pos, "todo: binop type: "+typ.String())
	}
}

func (c *Compiler) parseConst(prefix string, expr *ssa.Const) (llvm.Value, error) {
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
		} else if typ.Info()&types.IsString != 0 {
			str := constant.StringVal(expr.Value)
			strLen := llvm.ConstInt(c.uintptrType, uint64(len(str)), false)
			objname := prefix + "$string"
			global := llvm.AddGlobal(c.mod, llvm.ArrayType(c.ctx.Int8Type(), len(str)), objname)
			global.SetInitializer(c.ctx.ConstString(str, false))
			global.SetLinkage(llvm.InternalLinkage)
			global.SetGlobalConstant(true)
			global.SetUnnamedAddr(true)
			zero := llvm.ConstInt(c.ctx.Int32Type(), 0, false)
			strPtr := c.builder.CreateInBoundsGEP(global, []llvm.Value{zero, zero}, "")
			strObj := llvm.ConstNamedStruct(c.mod.GetTypeByName("runtime._string"), []llvm.Value{strPtr, strLen})
			return strObj, nil
		} else if typ.Kind() == types.UnsafePointer {
			if !expr.IsNil() {
				value, _ := constant.Uint64Val(expr.Value)
				return llvm.ConstIntToPtr(llvm.ConstInt(c.uintptrType, value, false), c.i8ptrType), nil
			}
			return llvm.ConstNull(c.i8ptrType), nil
		} else if typ.Info()&types.IsUnsigned != 0 {
			n, _ := constant.Uint64Val(expr.Value)
			return llvm.ConstInt(llvmType, n, false), nil
		} else if typ.Info()&types.IsInteger != 0 { // signed
			n, _ := constant.Int64Val(expr.Value)
			return llvm.ConstInt(llvmType, uint64(n), true), nil
		} else if typ.Info()&types.IsFloat != 0 {
			n, _ := constant.Float64Val(expr.Value)
			return llvm.ConstFloat(llvmType, n), nil
		} else if typ.Kind() == types.Complex128 {
			r, err := c.parseConst(prefix, ssa.NewConst(constant.Real(expr.Value), types.Typ[types.Float64]))
			if err != nil {
				return llvm.Value{}, err
			}
			i, err := c.parseConst(prefix, ssa.NewConst(constant.Imag(expr.Value), types.Typ[types.Float64]))
			if err != nil {
				return llvm.Value{}, err
			}
			cplx := llvm.Undef(llvm.VectorType(c.ctx.DoubleType(), 2))
			cplx = c.builder.CreateInsertElement(cplx, r, llvm.ConstInt(c.ctx.Int8Type(), 0, false), "")
			cplx = c.builder.CreateInsertElement(cplx, i, llvm.ConstInt(c.ctx.Int8Type(), 1, false), "")
			return cplx, nil
		} else {
			return llvm.Value{}, errors.New("todo: unknown constant: " + expr.String())
		}
	case *types.Signature:
		if expr.Value != nil {
			return llvm.Value{}, errors.New("non-nil signature constant")
		}
		sig, err := c.getLLVMType(expr.Type())
		if err != nil {
			return llvm.Value{}, err
		}
		return c.getZeroValue(sig)
	case *types.Interface:
		if expr.Value != nil {
			return llvm.Value{}, errors.New("non-nil interface constant")
		}
		// Create a generic nil interface with no dynamic type (typecode=0).
		fields := []llvm.Value{
			llvm.ConstInt(c.uintptrType, 0, false),
			llvm.ConstPointerNull(c.i8ptrType),
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
		llvmLen := llvm.ConstInt(c.uintptrType, 0, false)
		slice := c.ctx.ConstStruct([]llvm.Value{
			llvmPtr, // backing array
			llvmLen, // len
			llvmLen, // cap
		}, false)
		return slice, nil
	case *types.Map:
		if !expr.IsNil() {
			// I believe this is not allowed by the Go spec.
			panic("non-nil map constant")
		}
		llvmType, err := c.getLLVMType(typ)
		if err != nil {
			return llvm.Value{}, err
		}
		return c.getZeroValue(llvmType)
	default:
		return llvm.Value{}, errors.New("todo: unknown constant: " + expr.String())
	}
}

func (c *Compiler) parseConvert(typeFrom, typeTo types.Type, value llvm.Value, pos token.Pos) (llvm.Value, error) {
	llvmTypeFrom := value.Type()
	llvmTypeTo, err := c.getLLVMType(typeTo)
	if err != nil {
		return llvm.Value{}, err
	}

	// Conversion between unsafe.Pointer and uintptr.
	isPtrFrom := isPointer(typeFrom.Underlying())
	isPtrTo := isPointer(typeTo.Underlying())
	if isPtrFrom && !isPtrTo {
		return c.builder.CreatePtrToInt(value, llvmTypeTo, ""), nil
	} else if !isPtrFrom && isPtrTo {
		return c.builder.CreateIntToPtr(value, llvmTypeTo, ""), nil
	}

	// Conversion between pointers and unsafe.Pointer.
	if isPtrFrom && isPtrTo {
		return c.builder.CreateBitCast(value, llvmTypeTo, ""), nil
	}

	switch typeTo := typeTo.Underlying().(type) {
	case *types.Basic:
		sizeFrom := c.targetData.TypeAllocSize(llvmTypeFrom)

		if typeTo.Info()&types.IsString != 0 {
			switch typeFrom := typeFrom.Underlying().(type) {
			case *types.Basic:
				// Assume a Unicode code point, as that is the only possible
				// value here.
				// Cast to an i32 value as expected by
				// runtime.stringFromUnicode.
				if sizeFrom > 4 {
					value = c.builder.CreateTrunc(value, c.ctx.Int32Type(), "")
				} else if sizeFrom < 4 && typeTo.Info()&types.IsUnsigned != 0 {
					value = c.builder.CreateZExt(value, c.ctx.Int32Type(), "")
				} else if sizeFrom < 4 {
					value = c.builder.CreateSExt(value, c.ctx.Int32Type(), "")
				}
				return c.createRuntimeCall("stringFromUnicode", []llvm.Value{value}, ""), nil
			case *types.Slice:
				switch typeFrom.Elem().(*types.Basic).Kind() {
				case types.Byte:
					return c.createRuntimeCall("stringFromBytes", []llvm.Value{value}, ""), nil
				default:
					return llvm.Value{}, c.makeError(pos, "todo: convert to string: "+typeFrom.String())
				}
			default:
				return llvm.Value{}, c.makeError(pos, "todo: convert to string: "+typeFrom.String())
			}
		}

		typeFrom := typeFrom.Underlying().(*types.Basic)
		sizeTo := c.targetData.TypeAllocSize(llvmTypeTo)

		if typeFrom.Info()&types.IsInteger != 0 && typeTo.Info()&types.IsInteger != 0 {
			// Conversion between two integers.
			if sizeFrom > sizeTo {
				return c.builder.CreateTrunc(value, llvmTypeTo, ""), nil
			} else if typeTo.Info()&types.IsUnsigned != 0 { // if unsigned
				return c.builder.CreateZExt(value, llvmTypeTo, ""), nil
			} else { // if signed
				return c.builder.CreateSExt(value, llvmTypeTo, ""), nil
			}
		}

		if typeFrom.Info()&types.IsFloat != 0 && typeTo.Info()&types.IsFloat != 0 {
			// Conversion between two floats.
			if sizeFrom > sizeTo {
				return c.builder.CreateFPTrunc(value, llvmTypeTo, ""), nil
			} else if sizeFrom < sizeTo {
				return c.builder.CreateFPExt(value, llvmTypeTo, ""), nil
			} else {
				return value, nil
			}
		}

		if typeFrom.Info()&types.IsFloat != 0 && typeTo.Info()&types.IsInteger != 0 {
			// Conversion from float to int.
			if typeTo.Info()&types.IsUnsigned != 0 { // if unsigned
				return c.builder.CreateFPToUI(value, llvmTypeTo, ""), nil
			} else { // if signed
				return c.builder.CreateFPToSI(value, llvmTypeTo, ""), nil
			}
		}

		if typeFrom.Info()&types.IsInteger != 0 && typeTo.Info()&types.IsFloat != 0 {
			// Conversion from int to float.
			if typeFrom.Info()&types.IsUnsigned != 0 { // if unsigned
				return c.builder.CreateUIToFP(value, llvmTypeTo, ""), nil
			} else { // if signed
				return c.builder.CreateSIToFP(value, llvmTypeTo, ""), nil
			}
		}

		if typeFrom.Kind() == types.Complex128 && typeTo.Kind() == types.Complex64 {
			// Conversion from complex128 to complex64.
			r := c.builder.CreateExtractElement(value, llvm.ConstInt(c.ctx.Int32Type(), 0, false), "real.f64")
			i := c.builder.CreateExtractElement(value, llvm.ConstInt(c.ctx.Int32Type(), 1, false), "imag.f64")
			r = c.builder.CreateFPTrunc(r, c.ctx.FloatType(), "real.f32")
			i = c.builder.CreateFPTrunc(i, c.ctx.FloatType(), "imag.f32")
			cplx := llvm.Undef(llvm.VectorType(c.ctx.FloatType(), 2))
			cplx = c.builder.CreateInsertElement(cplx, r, llvm.ConstInt(c.ctx.Int8Type(), 0, false), "")
			cplx = c.builder.CreateInsertElement(cplx, i, llvm.ConstInt(c.ctx.Int8Type(), 1, false), "")
			return cplx, nil
		}

		if typeFrom.Kind() == types.Complex64 && typeTo.Kind() == types.Complex128 {
			// Conversion from complex64 to complex128.
			r := c.builder.CreateExtractElement(value, llvm.ConstInt(c.ctx.Int32Type(), 0, false), "real.f32")
			i := c.builder.CreateExtractElement(value, llvm.ConstInt(c.ctx.Int32Type(), 1, false), "imag.f32")
			r = c.builder.CreateFPExt(r, c.ctx.DoubleType(), "real.f64")
			i = c.builder.CreateFPExt(i, c.ctx.DoubleType(), "imag.f64")
			cplx := llvm.Undef(llvm.VectorType(c.ctx.DoubleType(), 2))
			cplx = c.builder.CreateInsertElement(cplx, r, llvm.ConstInt(c.ctx.Int8Type(), 0, false), "")
			cplx = c.builder.CreateInsertElement(cplx, i, llvm.ConstInt(c.ctx.Int8Type(), 1, false), "")
			return cplx, nil
		}

		return llvm.Value{}, c.makeError(pos, "todo: convert: basic non-integer type: "+typeFrom.String()+" -> "+typeTo.String())

	case *types.Slice:
		if basic, ok := typeFrom.(*types.Basic); !ok || basic.Info()&types.IsString == 0 {
			panic("can only convert from a string to a slice")
		}

		elemType := typeTo.Elem().Underlying().(*types.Basic) // must be byte or rune
		switch elemType.Kind() {
		case types.Byte:
			return c.createRuntimeCall("stringToBytes", []llvm.Value{value}, ""), nil
		default:
			return llvm.Value{}, c.makeError(pos, "todo: convert from string: "+elemType.String())
		}

	default:
		return llvm.Value{}, c.makeError(pos, "todo: convert "+typeTo.String()+" <- "+typeFrom.String())
	}
}

func (c *Compiler) parseMakeClosure(frame *Frame, expr *ssa.MakeClosure) (llvm.Value, error) {
	if len(expr.Bindings) == 0 {
		panic("unexpected: MakeClosure without bound variables")
	}
	f := c.ir.GetFunction(expr.Fn.(*ssa.Function))

	// Collect all bound variables.
	boundVars := make([]llvm.Value, 0, len(expr.Bindings))
	boundVarTypes := make([]llvm.Type, 0, len(expr.Bindings))
	for _, binding := range expr.Bindings {
		// The context stores the bound variables.
		llvmBoundVar, err := c.parseExpr(frame, binding)
		if err != nil {
			return llvm.Value{}, err
		}
		boundVars = append(boundVars, llvmBoundVar)
		boundVarTypes = append(boundVarTypes, llvmBoundVar.Type())
	}
	contextType := c.ctx.StructType(boundVarTypes, false)

	// Allocate memory for the context.
	contextAlloc := llvm.Value{}
	contextHeapAlloc := llvm.Value{}
	if c.targetData.TypeAllocSize(contextType) <= c.targetData.TypeAllocSize(c.i8ptrType) {
		// Context fits in a pointer - e.g. when it is a pointer. Store it
		// directly in the stack after a convert.
		// Because contextType is a struct and we have to cast it to a *i8,
		// store it in an alloca first for bitcasting (store+bitcast+load).
		contextAlloc = c.builder.CreateAlloca(contextType, "")
	} else {
		// Context is bigger than a pointer, so allocate it on the heap.
		size := c.targetData.TypeAllocSize(contextType)
		sizeValue := llvm.ConstInt(c.uintptrType, size, false)
		contextHeapAlloc = c.createRuntimeCall("alloc", []llvm.Value{sizeValue}, "")
		contextAlloc = c.builder.CreateBitCast(contextHeapAlloc, llvm.PointerType(contextType, 0), "")
	}

	// Store all bound variables in the alloca or heap pointer.
	for i, boundVar := range boundVars {
		indices := []llvm.Value{
			llvm.ConstInt(c.ctx.Int32Type(), 0, false),
			llvm.ConstInt(c.ctx.Int32Type(), uint64(i), false),
		}
		gep := c.builder.CreateInBoundsGEP(contextAlloc, indices, "")
		c.builder.CreateStore(boundVar, gep)
	}

	context := llvm.Value{}
	if c.targetData.TypeAllocSize(contextType) <= c.targetData.TypeAllocSize(c.i8ptrType) {
		// Load value (as *i8) from the alloca.
		contextAlloc = c.builder.CreateBitCast(contextAlloc, llvm.PointerType(c.i8ptrType, 0), "")
		context = c.builder.CreateLoad(contextAlloc, "")
	} else {
		// Get the original heap allocation pointer, which already is an
		// *i8.
		context = contextHeapAlloc
	}

	// Get the function signature type, which is a closure type.
	// A closure is a tuple of {context, function pointer}.
	typ, err := c.getLLVMType(f.Signature)
	if err != nil {
		return llvm.Value{}, err
	}

	// Create the closure, which is a struct: {context, function pointer}.
	closure, err := c.getZeroValue(typ)
	if err != nil {
		return llvm.Value{}, err
	}
	closure = c.builder.CreateInsertValue(closure, f.LLVMFn, 1, "")
	closure = c.builder.CreateInsertValue(closure, context, 0, "")
	return closure, nil
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
		if typ, ok := unop.X.Type().Underlying().(*types.Basic); ok {
			if typ.Info()&types.IsInteger != 0 {
				return c.builder.CreateSub(llvm.ConstInt(x.Type(), 0, false), x, ""), nil
			} else if typ.Info()&types.IsFloat != 0 {
				return c.builder.CreateFSub(llvm.ConstFloat(x.Type(), 0.0), x, ""), nil
			} else {
				return llvm.Value{}, c.makeError(unop.Pos(), "todo: unknown basic type for negate: "+typ.String())
			}
		} else {
			return llvm.Value{}, c.makeError(unop.Pos(), "todo: unknown type for negate: "+unop.X.Type().Underlying().String())
		}
	case token.MUL: // *x, dereference pointer
		valType := unop.X.Type().(*types.Pointer).Elem()
		if c.targetData.TypeAllocSize(x.Type().ElementType()) == 0 {
			// zero-length data
			return c.getZeroValue(x.Type().ElementType())
		} else {
			load := c.builder.CreateLoad(x, "")
			if c.ir.IsVolatile(valType) {
				// Volatile load, for memory-mapped registers.
				load.SetVolatile(true)
			}
			return load, nil
		}
	case token.XOR: // ^x, toggle all bits in integer
		return c.builder.CreateXor(x, llvm.ConstInt(x.Type(), ^uint64(0), false), ""), nil
	case token.ARROW: // <-x, receive from channel
		return c.emitChanRecv(frame, unop)
	default:
		return llvm.Value{}, c.makeError(unop.Pos(), "todo: unknown unop")
	}
}

// IR returns the whole IR as a human-readable string.
func (c *Compiler) IR() string {
	return c.mod.String()
}

func (c *Compiler) Verify() error {
	return llvm.VerifyModule(c.mod, llvm.PrintMessageAction)
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

// Turn all global constants into global variables. This works around a
// limitation on Harvard architectures (e.g. AVR), where constant and
// non-constant pointers point to a different address space.
func (c *Compiler) NonConstGlobals() {
	global := c.mod.FirstGlobal()
	for !global.IsNil() {
		global.SetGlobalConstant(false)
		global = llvm.NextGlobal(global)
	}
}

// When -wasm-abi flag set to "js" (default),
// replace i64 in an external function with a stack-allocated i64*, to work
// around the lack of 64-bit integers in JavaScript (commonly used together with
// WebAssembly). Once that's resolved, this pass may be avoided.
// See also the -wasm-abi= flag
// https://github.com/WebAssembly/design/issues/1172
func (c *Compiler) ExternalInt64AsPtr() error {
	int64Type := c.ctx.Int64Type()
	int64PtrType := llvm.PointerType(int64Type, 0)
	for fn := c.mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
		if fn.Linkage() != llvm.ExternalLinkage {
			// Only change externally visible functions (exports and imports).
			continue
		}
		if strings.HasPrefix(fn.Name(), "llvm.") || strings.HasPrefix(fn.Name(), "runtime.") {
			// Do not try to modify the signature of internal LLVM functions and
			// assume that runtime functions are only temporarily exported for
			// coroutine lowering.
			continue
		}

		hasInt64 := false
		paramTypes := []llvm.Type{}

		// Check return type for 64-bit integer.
		fnType := fn.Type().ElementType()
		returnType := fnType.ReturnType()
		if returnType == int64Type {
			hasInt64 = true
			paramTypes = append(paramTypes, int64PtrType)
			returnType = c.ctx.VoidType()
		}

		// Check param types for 64-bit integers.
		for param := fn.FirstParam(); !param.IsNil(); param = llvm.NextParam(param) {
			if param.Type() == int64Type {
				hasInt64 = true
				paramTypes = append(paramTypes, int64PtrType)
			} else {
				paramTypes = append(paramTypes, param.Type())
			}
		}

		if !hasInt64 {
			// No i64 in the paramter list.
			continue
		}

		// Add $i64wrapper to the real function name as it is only used
		// internally.
		// Add a new function with the correct signature that is exported.
		name := fn.Name()
		fn.SetName(name + "$i64wrap")
		externalFnType := llvm.FunctionType(returnType, paramTypes, fnType.IsFunctionVarArg())
		externalFn := llvm.AddFunction(c.mod, name, externalFnType)

		if fn.IsDeclaration() {
			// Just a declaration: the definition doesn't exist on the Go side
			// so it cannot be called from external code.
			// Update all users to call the external function.
			// The old $i64wrapper function could be removed, but it may as well
			// be left in place.
			for use := fn.FirstUse(); !use.IsNil(); use = use.NextUse() {
				call := use.User()
				c.builder.SetInsertPointBefore(call)
				callParams := []llvm.Value{}
				var retvalAlloca llvm.Value
				if fnType.ReturnType() == int64Type {
					retvalAlloca = c.builder.CreateAlloca(int64Type, "i64asptr")
					callParams = append(callParams, retvalAlloca)
				}
				for i := 0; i < call.OperandsCount()-1; i++ {
					operand := call.Operand(i)
					if operand.Type() == int64Type {
						// Pass a stack-allocated pointer instead of the value
						// itself.
						alloca := c.builder.CreateAlloca(int64Type, "i64asptr")
						c.builder.CreateStore(operand, alloca)
						callParams = append(callParams, alloca)
					} else {
						// Unchanged parameter.
						callParams = append(callParams, operand)
					}
				}
				if fnType.ReturnType() == int64Type {
					// Pass a stack-allocated pointer as the first parameter
					// where the return value should be stored, instead of using
					// the regular return value.
					c.builder.CreateCall(externalFn, callParams, call.Name())
					returnValue := c.builder.CreateLoad(retvalAlloca, "retval")
					call.ReplaceAllUsesWith(returnValue)
					call.EraseFromParentAsInstruction()
				} else {
					newCall := c.builder.CreateCall(externalFn, callParams, call.Name())
					call.ReplaceAllUsesWith(newCall)
					call.EraseFromParentAsInstruction()
				}
			}
		} else {
			// The function has a definition in Go. This means that it may still
			// be called both Go and from external code.
			// Keep existing calls with the existing convention in place (for
			// better performance), but export a new wrapper function with the
			// correct calling convention.
			fn.SetLinkage(llvm.InternalLinkage)
			fn.SetUnnamedAddr(true)
			entryBlock := llvm.AddBasicBlock(externalFn, "entry")
			c.builder.SetInsertPointAtEnd(entryBlock)
			var callParams []llvm.Value
			if fnType.ReturnType() == int64Type {
				return errors.New("not yet implemented: exported function returns i64 with -wasm-abi=js; " +
					"see https://tinygo.org/compiler-internals/calling-convention/")
			}
			for i, origParam := range fn.Params() {
				paramValue := externalFn.Param(i)
				if origParam.Type() == int64Type {
					paramValue = c.builder.CreateLoad(paramValue, "i64")
				}
				callParams = append(callParams, paramValue)
			}
			retval := c.builder.CreateCall(fn, callParams, "")
			if retval.Type().TypeKind() == llvm.VoidTypeKind {
				c.builder.CreateRetVoid()
			} else {
				c.builder.CreateRet(retval)
			}
		}
	}

	return nil
}

// Emit object file (.o).
func (c *Compiler) EmitObject(path string) error {
	llvmBuf, err := c.machine.EmitToMemoryBuffer(c.mod, llvm.ObjectFile)
	if err != nil {
		return err
	}
	return c.writeFile(llvmBuf.Bytes(), path)
}

// Emit LLVM bitcode file (.bc).
func (c *Compiler) EmitBitcode(path string) error {
	data := llvm.WriteBitcodeToMemoryBuffer(c.mod).Bytes()
	return c.writeFile(data, path)
}

// Emit LLVM IR source file (.ll).
func (c *Compiler) EmitText(path string) error {
	data := []byte(c.mod.String())
	return c.writeFile(data, path)
}

// Write the data to the file specified by path.
func (c *Compiler) writeFile(data []byte, path string) error {
	// Write output to file
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err != nil {
		return err
	}
	return f.Close()
}
