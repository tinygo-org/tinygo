package compiler

import (
	"errors"
	"fmt"
	"go/ast"
	"go/build"
	"go/constant"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/tinygo-org/tinygo/ir"
	"github.com/tinygo-org/tinygo/loader"
	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

func init() {
	llvm.InitializeAllTargets()
	llvm.InitializeAllTargetMCs()
	llvm.InitializeAllTargetInfos()
	llvm.InitializeAllAsmParsers()
	llvm.InitializeAllAsmPrinters()
}

// The TinyGo import path.
const tinygoPath = "github.com/tinygo-org/tinygo"

// Configure the compiler.
type Config struct {
	Triple        string   // LLVM target triple, e.g. x86_64-unknown-linux-gnu (empty string means default)
	CPU           string   // LLVM CPU name, e.g. atmega328p (empty string means default)
	Features      []string // LLVM CPU features
	GOOS          string   //
	GOARCH        string   //
	GC            string   // garbage collection strategy
	PanicStrategy string   // panic strategy ("abort" or "trap")
	CFlags        []string // cflags to pass to cgo
	LDFlags       []string // ldflags to pass to cgo
	DumpSSA       bool     // dump Go SSA, for compiler debugging
	Debug         bool     // add debug symbols for gdb
	GOROOT        string   // GOROOT
	TINYGOROOT    string   // GOROOT for TinyGo
	GOPATH        string   // GOPATH, like `go env GOPATH`
	BuildTags     []string // build tags for TinyGo (empty means {Config.GOOS/Config.GOARCH})
}

type Compiler struct {
	Config
	mod                     llvm.Module
	ctx                     llvm.Context
	builder                 llvm.Builder
	dibuilder               *llvm.DIBuilder
	cu                      llvm.Metadata
	difiles                 map[string]llvm.Metadata
	machine                 llvm.TargetMachine
	targetData              llvm.TargetData
	intType                 llvm.Type
	i8ptrType               llvm.Type // for convenience
	funcPtrAddrSpace        int
	uintptrType             llvm.Type
	initFuncs               []llvm.Value
	interfaceInvokeWrappers []interfaceInvokeWrapper
	ir                      *ir.Program
	diagnostics             []error
	astComments             map[string]*ast.CommentGroup
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
	selectRecvBuf     map[*ssa.Select]llvm.Value
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
	}

	target, err := llvm.GetTargetFromTriple(config.Triple)
	if err != nil {
		return nil, err
	}
	features := ""
	if len(config.Features) > 0 {
		features = strings.Join(config.Features, `,`)
	}
	c.machine = target.CreateTargetMachine(config.Triple, config.CPU, features, llvm.CodeGenLevelDefault, llvm.RelocStatic, llvm.CodeModelDefault)
	c.targetData = c.machine.CreateTargetData()

	c.ctx = llvm.NewContext()
	c.mod = c.ctx.NewModule(pkgName)
	c.mod.SetTarget(config.Triple)
	c.mod.SetDataLayout(c.targetData.String())
	c.builder = c.ctx.NewBuilder()
	if c.Debug {
		c.dibuilder = llvm.NewDIBuilder(c.mod)
	}

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

	dummyFuncType := llvm.FunctionType(c.ctx.VoidType(), nil, false)
	dummyFunc := llvm.AddFunction(c.mod, "tinygo.dummy", dummyFuncType)
	c.funcPtrAddrSpace = dummyFunc.Type().PointerAddressSpace()
	dummyFunc.EraseFromParentAsFunction()

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
func (c *Compiler) Compile(mainPath string) []error {
	// Prefix the GOPATH with the system GOROOT, as GOROOT is already set to
	// the TinyGo root.
	overlayGopath := c.GOPATH
	if overlayGopath == "" {
		overlayGopath = c.GOROOT
	} else {
		overlayGopath = c.GOROOT + string(filepath.ListSeparator) + overlayGopath
	}

	wd, err := os.Getwd()
	if err != nil {
		return []error{err}
	}
	lprogram := &loader.Program{
		Build: &build.Context{
			GOARCH:      c.GOARCH,
			GOOS:        c.GOOS,
			GOROOT:      c.GOROOT,
			GOPATH:      c.GOPATH,
			CgoEnabled:  true,
			UseAllFiles: false,
			Compiler:    "gc", // must be one of the recognized compilers
			BuildTags:   append([]string{"tinygo", "gc." + c.selectGC()}, c.BuildTags...),
		},
		OverlayBuild: &build.Context{
			GOARCH:      c.GOARCH,
			GOOS:        c.GOOS,
			GOROOT:      c.TINYGOROOT,
			GOPATH:      overlayGopath,
			CgoEnabled:  true,
			UseAllFiles: false,
			Compiler:    "gc", // must be one of the recognized compilers
			BuildTags:   append([]string{"tinygo", "gc." + c.selectGC()}, c.BuildTags...),
		},
		OverlayPath: func(path string) string {
			// Return the (overlay) import path when it should be overlaid, and
			// "" if it should not.
			if strings.HasPrefix(path, tinygoPath+"/src/") {
				// Avoid issues with packages that are imported twice, one from
				// GOPATH and one from TINYGOPATH.
				path = path[len(tinygoPath+"/src/"):]
			}
			switch path {
			case "machine", "os", "reflect", "runtime", "runtime/volatile", "sync":
				return path
			default:
				if strings.HasPrefix(path, "device/") || strings.HasPrefix(path, "examples/") {
					return path
				} else if path == "syscall" {
					for _, tag := range c.BuildTags {
						if tag == "avr" || tag == "cortexm" || tag == "darwin" {
							return path
						}
					}
				}
			}
			return ""
		},
		TypeChecker: types.Config{
			Sizes: &StdSizes{
				IntSize:  int64(c.targetData.TypeAllocSize(c.intType)),
				PtrSize:  int64(c.targetData.PointerSize()),
				MaxAlign: int64(c.targetData.PrefTypeAlignment(c.i8ptrType)),
			},
		},
		Dir:        wd,
		TINYGOROOT: c.TINYGOROOT,
		CFlags:     c.CFlags,
	}
	if strings.HasSuffix(mainPath, ".go") {
		_, err = lprogram.ImportFile(mainPath)
		if err != nil {
			return []error{err}
		}
	} else {
		_, err = lprogram.Import(mainPath, wd)
		if err != nil {
			return []error{err}
		}
	}
	_, err = lprogram.Import("runtime", "")
	if err != nil {
		return []error{err}
	}

	err = lprogram.Parse()
	if err != nil {
		return []error{err}
	}

	c.ir = ir.NewProgram(lprogram, mainPath)

	// Run a simple dead code elimination pass.
	c.ir.SimpleDCE()

	// Initialize debug information.
	if c.Debug {
		c.cu = c.dibuilder.CreateCompileUnit(llvm.DICompileUnit{
			Language:  0xb, // DW_LANG_C99 (0xc, off-by-one?)
			File:      mainPath,
			Dir:       "",
			Producer:  "TinyGo",
			Optimized: true,
		})
	}

	var frames []*Frame

	c.loadASTComments(lprogram)

	// Declare all functions.
	for _, f := range c.ir.Functions {
		frames = append(frames, c.parseFuncDecl(f))
	}

	// Add definitions to declarations.
	for _, frame := range frames {
		if frame.fn.Synthetic == "package initializer" {
			c.initFuncs = append(c.initFuncs, frame.fn.LLVMFn)
		}
		if frame.fn.CName() != "" {
			continue
		}
		if frame.fn.Blocks == nil {
			continue // external function
		}
		c.parseFunc(frame)
	}

	// Define the already declared functions that wrap methods for use in
	// interfaces.
	for _, state := range c.interfaceInvokeWrappers {
		c.createInterfaceInvokeWrapper(state)
	}

	// After all packages are imported, add a synthetic initializer function
	// that calls the initializer of each package.
	initFn := c.ir.GetFunction(c.ir.Program.ImportedPackage("runtime").Members["initAll"].(*ssa.Function))
	initFn.LLVMFn.SetLinkage(llvm.InternalLinkage)
	initFn.LLVMFn.SetUnnamedAddr(true)
	if c.Debug {
		difunc := c.attachDebugInfo(initFn)
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
	c.mod.NamedFunction("runtime.sleepTask").SetLinkage(llvm.ExternalLinkage)
	c.mod.NamedFunction("runtime.setTaskPromisePtr").SetLinkage(llvm.ExternalLinkage)
	c.mod.NamedFunction("runtime.getTaskPromisePtr").SetLinkage(llvm.ExternalLinkage)
	c.mod.NamedFunction("runtime.activateTask").SetLinkage(llvm.ExternalLinkage)
	c.mod.NamedFunction("runtime.scheduler").SetLinkage(llvm.ExternalLinkage)

	// Load some attributes
	getAttr := func(attrName string) llvm.Attribute {
		attrKind := llvm.AttributeKindID(attrName)
		return c.ctx.CreateEnumAttribute(attrKind, 0)
	}
	nocapture := getAttr("nocapture")
	writeonly := getAttr("writeonly")
	readonly := getAttr("readonly")

	// Tell the optimizer that runtime.alloc is an allocator, meaning that it
	// returns values that are never null and never alias to an existing value.
	for _, attrName := range []string{"noalias", "nonnull"} {
		c.mod.NamedFunction("runtime.alloc").AddAttributeAtIndex(0, getAttr(attrName))
	}

	// See emitNilCheck in asserts.go.
	c.mod.NamedFunction("runtime.isnil").AddAttributeAtIndex(1, nocapture)

	// Memory copy operations do not capture pointers, even though some weird
	// pointer arithmetic is happening in the Go implementation.
	for _, fnName := range []string{"runtime.memcpy", "runtime.memmove"} {
		fn := c.mod.NamedFunction(fnName)
		fn.AddAttributeAtIndex(1, nocapture)
		fn.AddAttributeAtIndex(1, writeonly)
		fn.AddAttributeAtIndex(2, nocapture)
		fn.AddAttributeAtIndex(2, readonly)
	}

	// see: https://reviews.llvm.org/D18355
	if c.Debug {
		c.mod.AddNamedMetadataOperand("llvm.module.flags",
			c.ctx.MDNode([]llvm.Metadata{
				llvm.ConstInt(c.ctx.Int32Type(), 1, false).ConstantAsMetadata(), // Error on mismatch
				llvm.GlobalContext().MDString("Debug Info Version"),
				llvm.ConstInt(c.ctx.Int32Type(), 3, false).ConstantAsMetadata(), // DWARF version
			}),
		)
		c.mod.AddNamedMetadataOperand("llvm.module.flags",
			c.ctx.MDNode([]llvm.Metadata{
				llvm.ConstInt(c.ctx.Int32Type(), 1, false).ConstantAsMetadata(),
				llvm.GlobalContext().MDString("Dwarf Version"),
				llvm.ConstInt(c.ctx.Int32Type(), 4, false).ConstantAsMetadata(),
			}),
		)
		c.dibuilder.Finalize()
	}

	return c.diagnostics
}

// getRuntimeType obtains a named type from the runtime package and returns it
// as a Go type.
func (c *Compiler) getRuntimeType(name string) types.Type {
	return c.ir.Program.ImportedPackage("runtime").Type(name).Type()
}

// getLLVMRuntimeType obtains a named type from the runtime package and returns
// it as a LLVM type, creating it if necessary. It is a shorthand for
// getLLVMType(getRuntimeType(name)).
func (c *Compiler) getLLVMRuntimeType(name string) llvm.Type {
	return c.getLLVMType(c.getRuntimeType(name))
}

// getLLVMType creates and returns a LLVM type for a Go type. In the case of
// named struct types (or Go types implemented as named LLVM structs such as
// strings) it also creates it first if necessary.
func (c *Compiler) getLLVMType(goType types.Type) llvm.Type {
	switch typ := goType.(type) {
	case *types.Array:
		elemType := c.getLLVMType(typ.Elem())
		return llvm.ArrayType(elemType, int(typ.Len()))
	case *types.Basic:
		switch typ.Kind() {
		case types.Bool, types.UntypedBool:
			return c.ctx.Int1Type()
		case types.Int8, types.Uint8:
			return c.ctx.Int8Type()
		case types.Int16, types.Uint16:
			return c.ctx.Int16Type()
		case types.Int32, types.Uint32:
			return c.ctx.Int32Type()
		case types.Int, types.Uint:
			return c.intType
		case types.Int64, types.Uint64:
			return c.ctx.Int64Type()
		case types.Float32:
			return c.ctx.FloatType()
		case types.Float64:
			return c.ctx.DoubleType()
		case types.Complex64:
			return c.ctx.StructType([]llvm.Type{c.ctx.FloatType(), c.ctx.FloatType()}, false)
		case types.Complex128:
			return c.ctx.StructType([]llvm.Type{c.ctx.DoubleType(), c.ctx.DoubleType()}, false)
		case types.String, types.UntypedString:
			return c.getLLVMRuntimeType("_string")
		case types.Uintptr:
			return c.uintptrType
		case types.UnsafePointer:
			return c.i8ptrType
		default:
			panic("unknown basic type: " + typ.String())
		}
	case *types.Chan:
		return llvm.PointerType(c.getLLVMRuntimeType("channel"), 0)
	case *types.Interface:
		return c.getLLVMRuntimeType("_interface")
	case *types.Map:
		return llvm.PointerType(c.getLLVMRuntimeType("hashmap"), 0)
	case *types.Named:
		if st, ok := typ.Underlying().(*types.Struct); ok {
			// Structs are a special case. While other named types are ignored
			// in LLVM IR, named structs are implemented as named structs in
			// LLVM. This is because it is otherwise impossible to create
			// self-referencing types such as linked lists.
			llvmName := typ.Obj().Pkg().Path() + "." + typ.Obj().Name()
			llvmType := c.mod.GetTypeByName(llvmName)
			if llvmType.IsNil() {
				llvmType = c.ctx.StructCreateNamed(llvmName)
				underlying := c.getLLVMType(st)
				llvmType.StructSetBody(underlying.StructElementTypes(), false)
			}
			return llvmType
		}
		return c.getLLVMType(typ.Underlying())
	case *types.Pointer:
		ptrTo := c.getLLVMType(typ.Elem())
		return llvm.PointerType(ptrTo, 0)
	case *types.Signature: // function value
		return c.getFuncType(typ)
	case *types.Slice:
		elemType := c.getLLVMType(typ.Elem())
		members := []llvm.Type{
			llvm.PointerType(elemType, 0),
			c.uintptrType, // len
			c.uintptrType, // cap
		}
		return c.ctx.StructType(members, false)
	case *types.Struct:
		members := make([]llvm.Type, typ.NumFields())
		for i := 0; i < typ.NumFields(); i++ {
			members[i] = c.getLLVMType(typ.Field(i).Type())
		}
		if len(members) > 2 && typ.Field(0).Name() == "C union" {
			// Not a normal struct but a C union emitted by cgo.
			// Such a field name cannot be entered in regular Go code, this must
			// be manually inserted in the AST so this is safe.
			maxAlign := 0
			maxSize := uint64(0)
			mainType := members[0]
			for _, member := range members {
				align := c.targetData.ABITypeAlignment(member)
				size := c.targetData.TypeAllocSize(member)
				if align > maxAlign {
					maxAlign = align
					mainType = member
				} else if align == maxAlign && size > maxSize {
					maxAlign = align
					maxSize = size
					mainType = member
				} else if size > maxSize {
					maxSize = size
				}
			}
			members = []llvm.Type{mainType}
			mainTypeSize := c.targetData.TypeAllocSize(mainType)
			if mainTypeSize < maxSize {
				members = append(members, llvm.ArrayType(c.ctx.Int8Type(), int(maxSize-mainTypeSize)))
			}
		}
		return c.ctx.StructType(members, false)
	case *types.Tuple:
		members := make([]llvm.Type, typ.Len())
		for i := 0; i < typ.Len(); i++ {
			members[i] = c.getLLVMType(typ.At(i).Type())
		}
		return c.ctx.StructType(members, false)
	default:
		panic("unknown type: " + goType.String())
	}
}

// Return a zero LLVM value for any LLVM type. Setting this value as an
// initializer has the same effect as setting 'zeroinitializer' on a value.
// Sadly, I haven't found a way to do it directly with the Go API but this works
// just fine.
func (c *Compiler) getZeroValue(typ llvm.Type) llvm.Value {
	switch typ.TypeKind() {
	case llvm.ArrayTypeKind:
		subTyp := typ.ElementType()
		subVal := c.getZeroValue(subTyp)
		vals := make([]llvm.Value, typ.ArrayLength())
		for i := range vals {
			vals[i] = subVal
		}
		return llvm.ConstArray(subTyp, vals)
	case llvm.FloatTypeKind, llvm.DoubleTypeKind:
		return llvm.ConstFloat(typ, 0.0)
	case llvm.IntegerTypeKind:
		return llvm.ConstInt(typ, 0, false)
	case llvm.PointerTypeKind:
		return llvm.ConstPointerNull(typ)
	case llvm.StructTypeKind:
		types := typ.StructElementTypes()
		vals := make([]llvm.Value, len(types))
		for i, subTyp := range types {
			vals[i] = c.getZeroValue(subTyp)
		}
		if typ.StructName() != "" {
			return llvm.ConstNamedStruct(typ, vals)
		} else {
			return c.ctx.ConstStruct(vals, false)
		}
	default:
		panic("unknown LLVM zero inititializer: " + typ.String())
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
func (c *Compiler) getDIType(typ types.Type) llvm.Metadata {
	llvmType := c.getLLVMType(typ)
	sizeInBytes := c.targetData.TypeAllocSize(llvmType)
	switch typ := typ.(type) {
	case *types.Array:
		return c.dibuilder.CreateArrayType(llvm.DIArrayType{
			SizeInBits:  sizeInBytes * 8,
			AlignInBits: uint32(c.targetData.ABITypeAlignment(llvmType)) * 8,
			ElementType: c.getDIType(typ.Elem()),
			Subscripts: []llvm.DISubrange{
				llvm.DISubrange{
					Lo:    0,
					Count: typ.Len(),
				},
			},
		})
	case *types.Basic:
		var encoding llvm.DwarfTypeEncoding
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
			return c.dibuilder.CreatePointerType(llvm.DIPointerType{
				Name:         "unsafe.Pointer",
				SizeInBits:   c.targetData.TypeAllocSize(llvmType) * 8,
				AlignInBits:  uint32(c.targetData.ABITypeAlignment(llvmType)) * 8,
				AddressSpace: 0,
			})
		} else if typ.Info()&types.IsString != 0 {
			return c.dibuilder.CreateStructType(llvm.Metadata{}, llvm.DIStructType{
				Name:        "string",
				SizeInBits:  sizeInBytes * 8,
				AlignInBits: uint32(c.targetData.ABITypeAlignment(llvmType)) * 8,
				Elements: []llvm.Metadata{
					c.dibuilder.CreateMemberType(llvm.Metadata{}, llvm.DIMemberType{
						Name:         "ptr",
						SizeInBits:   c.targetData.TypeAllocSize(c.i8ptrType) * 8,
						AlignInBits:  uint32(c.targetData.ABITypeAlignment(c.i8ptrType)) * 8,
						OffsetInBits: 0,
						Type:         c.getDIType(types.NewPointer(types.Typ[types.Byte])),
					}),
					c.dibuilder.CreateMemberType(llvm.Metadata{}, llvm.DIMemberType{
						Name:         "len",
						SizeInBits:   c.targetData.TypeAllocSize(c.uintptrType) * 8,
						AlignInBits:  uint32(c.targetData.ABITypeAlignment(c.uintptrType)) * 8,
						OffsetInBits: c.targetData.ElementOffset(llvmType, 1) * 8,
						Type:         c.getDIType(types.Typ[types.Uintptr]),
					}),
				},
			})
		} else {
			panic("unknown basic type")
		}
		return c.dibuilder.CreateBasicType(llvm.DIBasicType{
			Name:       typ.String(),
			SizeInBits: sizeInBytes * 8,
			Encoding:   encoding,
		})
	case *types.Chan:
		return c.getDIType(types.NewPointer(c.ir.Program.ImportedPackage("runtime").Members["channel"].(*ssa.Type).Type()))
	case *types.Interface:
		return c.getDIType(c.ir.Program.ImportedPackage("runtime").Members["_interface"].(*ssa.Type).Type())
	case *types.Map:
		return c.getDIType(types.NewPointer(c.ir.Program.ImportedPackage("runtime").Members["hashmap"].(*ssa.Type).Type()))
	case *types.Named:
		return c.dibuilder.CreateTypedef(llvm.DITypedef{
			Type: c.getDIType(typ.Underlying()),
			Name: typ.String(),
		})
	case *types.Pointer:
		return c.dibuilder.CreatePointerType(llvm.DIPointerType{
			Pointee:      c.getDIType(typ.Elem()),
			SizeInBits:   c.targetData.TypeAllocSize(llvmType) * 8,
			AlignInBits:  uint32(c.targetData.ABITypeAlignment(llvmType)) * 8,
			AddressSpace: 0,
		})
	case *types.Signature:
		// actually a closure
		fields := llvmType.StructElementTypes()
		return c.dibuilder.CreateStructType(llvm.Metadata{}, llvm.DIStructType{
			SizeInBits:  sizeInBytes * 8,
			AlignInBits: uint32(c.targetData.ABITypeAlignment(llvmType)) * 8,
			Elements: []llvm.Metadata{
				c.dibuilder.CreateMemberType(llvm.Metadata{}, llvm.DIMemberType{
					Name:         "context",
					SizeInBits:   c.targetData.TypeAllocSize(fields[1]) * 8,
					AlignInBits:  uint32(c.targetData.ABITypeAlignment(fields[1])) * 8,
					OffsetInBits: 0,
					Type:         c.getDIType(types.Typ[types.UnsafePointer]),
				}),
				c.dibuilder.CreateMemberType(llvm.Metadata{}, llvm.DIMemberType{
					Name:         "fn",
					SizeInBits:   c.targetData.TypeAllocSize(fields[0]) * 8,
					AlignInBits:  uint32(c.targetData.ABITypeAlignment(fields[0])) * 8,
					OffsetInBits: c.targetData.ElementOffset(llvmType, 1) * 8,
					Type:         c.getDIType(types.Typ[types.UnsafePointer]),
				}),
			},
		})
	case *types.Slice:
		fields := llvmType.StructElementTypes()
		return c.dibuilder.CreateStructType(llvm.Metadata{}, llvm.DIStructType{
			Name:        typ.String(),
			SizeInBits:  sizeInBytes * 8,
			AlignInBits: uint32(c.targetData.ABITypeAlignment(llvmType)) * 8,
			Elements: []llvm.Metadata{
				c.dibuilder.CreateMemberType(llvm.Metadata{}, llvm.DIMemberType{
					Name:         "ptr",
					SizeInBits:   c.targetData.TypeAllocSize(fields[0]) * 8,
					AlignInBits:  uint32(c.targetData.ABITypeAlignment(fields[0])) * 8,
					OffsetInBits: 0,
					Type:         c.getDIType(types.NewPointer(typ.Elem())),
				}),
				c.dibuilder.CreateMemberType(llvm.Metadata{}, llvm.DIMemberType{
					Name:         "len",
					SizeInBits:   c.targetData.TypeAllocSize(c.uintptrType) * 8,
					AlignInBits:  uint32(c.targetData.ABITypeAlignment(c.uintptrType)) * 8,
					OffsetInBits: c.targetData.ElementOffset(llvmType, 1) * 8,
					Type:         c.getDIType(types.Typ[types.Uintptr]),
				}),
				c.dibuilder.CreateMemberType(llvm.Metadata{}, llvm.DIMemberType{
					Name:         "cap",
					SizeInBits:   c.targetData.TypeAllocSize(c.uintptrType) * 8,
					AlignInBits:  uint32(c.targetData.ABITypeAlignment(c.uintptrType)) * 8,
					OffsetInBits: c.targetData.ElementOffset(llvmType, 2) * 8,
					Type:         c.getDIType(types.Typ[types.Uintptr]),
				}),
			},
		})
	case *types.Struct:
		elements := make([]llvm.Metadata, typ.NumFields())
		for i := range elements {
			field := typ.Field(i)
			fieldType := field.Type()
			if _, ok := fieldType.Underlying().(*types.Pointer); ok {
				// XXX hack to avoid recursive types
				fieldType = types.Typ[types.UnsafePointer]
			}
			llvmField := c.getLLVMType(fieldType)
			elements[i] = c.dibuilder.CreateMemberType(llvm.Metadata{}, llvm.DIMemberType{
				Name:         field.Name(),
				SizeInBits:   c.targetData.TypeAllocSize(llvmField) * 8,
				AlignInBits:  uint32(c.targetData.ABITypeAlignment(llvmField)) * 8,
				OffsetInBits: c.targetData.ElementOffset(llvmType, i) * 8,
				Type:         c.getDIType(fieldType),
			})
		}
		return c.dibuilder.CreateStructType(llvm.Metadata{}, llvm.DIStructType{
			SizeInBits:  sizeInBytes * 8,
			AlignInBits: uint32(c.targetData.ABITypeAlignment(llvmType)) * 8,
			Elements:    elements,
		})
	default:
		panic("unknown type while generating DWARF debug type: " + typ.String())
	}
}

func (c *Compiler) parseFuncDecl(f *ir.Function) *Frame {
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
		retType = c.getLLVMType(f.Signature.Results().At(0).Type())
	} else {
		results := make([]llvm.Type, 0, f.Signature.Results().Len())
		for i := 0; i < f.Signature.Results().Len(); i++ {
			results = append(results, c.getLLVMType(f.Signature.Results().At(i).Type()))
		}
		retType = c.ctx.StructType(results, false)
	}

	var paramTypes []llvm.Type
	for _, param := range f.Params {
		paramType := c.getLLVMType(param.Type())
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

	// External/exported functions may not retain pointer values.
	// https://golang.org/cmd/cgo/#hdr-Passing_pointers
	if f.IsExported() {
		nocaptureKind := llvm.AttributeKindID("nocapture")
		nocapture := c.ctx.CreateEnumAttribute(nocaptureKind, 0)
		for i, typ := range paramTypes {
			if typ.TypeKind() == llvm.PointerTypeKind {
				frame.fn.LLVMFn.AddAttributeAtIndex(i+1, nocapture)
			}
		}
	}

	return frame
}

func (c *Compiler) attachDebugInfo(f *ir.Function) llvm.Metadata {
	pos := c.ir.Program.Fset.Position(f.Syntax().Pos())
	return c.attachDebugInfoRaw(f, f.LLVMFn, "", pos.Filename, pos.Line)
}

func (c *Compiler) attachDebugInfoRaw(f *ir.Function, llvmFn llvm.Value, suffix, filename string, line int) llvm.Metadata {
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
		diparams = append(diparams, c.getDIType(param.Type()))
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
	return difunc
}

func (c *Compiler) parseFunc(frame *Frame) {
	if c.DumpSSA {
		fmt.Printf("\nfunc %s:\n", frame.fn.Function)
	}
	if !frame.fn.LLVMFn.IsDeclaration() {
		c.addError(frame.fn.Pos(), "function is already defined:"+frame.fn.LLVMFn.Name())
		return
	}
	if !frame.fn.IsExported() {
		frame.fn.LLVMFn.SetLinkage(llvm.InternalLinkage)
		frame.fn.LLVMFn.SetUnnamedAddr(true)
	}
	if frame.fn.IsInterrupt() && strings.HasPrefix(c.Triple, "avr") {
		frame.fn.LLVMFn.SetFunctionCallConv(85) // CallingConv::AVR_SIGNAL
	}

	// Some functions have a pragma controlling the inlining level.
	switch frame.fn.Inline() {
	case ir.InlineHint:
		// Add LLVM inline hint to functions with //go:inline pragma.
		inline := c.ctx.CreateEnumAttribute(llvm.AttributeKindID("inlinehint"), 0)
		frame.fn.LLVMFn.AddFunctionAttr(inline)
	}

	// Add debug info, if needed.
	if c.Debug {
		if frame.fn.Synthetic == "package initializer" {
			// Package initializers have no debug info. Create some fake debug
			// info to at least have *something*.
			frame.difunc = c.attachDebugInfoRaw(frame.fn, frame.fn.LLVMFn, "", "", 0)
		} else if frame.fn.Syntax() != nil {
			// Create debug info file if needed.
			frame.difunc = c.attachDebugInfo(frame.fn)
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
		llvmType := c.getLLVMType(param.Type())
		fields := make([]llvm.Value, 0, 1)
		for range c.expandFormalParamType(llvmType) {
			fields = append(fields, frame.fn.LLVMFn.Param(llvmParamIndex))
			llvmParamIndex++
		}
		frame.locals[param] = c.collapseFormalParam(llvmType, fields)

		// Add debug information to this parameter (if available)
		if c.Debug && frame.fn.Syntax() != nil {
			pos := c.ir.Program.Fset.Position(frame.fn.Syntax().Pos())
			diType := c.getDIType(param.Type())
			dbgParam := c.dibuilder.CreateParameterVariable(frame.difunc, llvm.DIParameterVariable{
				Name:           param.Name(),
				File:           c.difiles[pos.Filename],
				Line:           pos.Line,
				Type:           diType,
				AlwaysPreserve: true,
				ArgNo:          i + 1,
			})
			loc := c.builder.GetCurrentDebugLocation()
			if len(fields) == 1 {
				expr := c.dibuilder.CreateExpression(nil)
				c.dibuilder.InsertValueAtEnd(fields[0], dbgParam, expr, loc, entryBlock)
			} else {
				fieldOffsets := c.expandFormalParamOffsets(llvmType)
				for i, field := range fields {
					expr := c.dibuilder.CreateExpression([]int64{
						0x1000,                     // DW_OP_LLVM_fragment
						int64(fieldOffsets[i]) * 8, // offset in bits
						int64(c.targetData.TypeAllocSize(field.Type())) * 8, // size in bits
					})
					c.dibuilder.InsertValueAtEnd(field, dbgParam, expr, loc, entryBlock)
				}
			}
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
		// Get a list of all variable types in the context.
		freeVarTypes := make([]llvm.Type, len(frame.fn.FreeVars))
		for i, freeVar := range frame.fn.FreeVars {
			freeVarTypes[i] = c.getLLVMType(freeVar.Type())
		}

		// Load each free variable from the context pointer.
		// A free variable is always a pointer when this is a closure, but it
		// can be another type when it is a wrapper for a bound method (these
		// wrappers are generated by the ssa package).
		for i, val := range c.emitPointerUnpack(context, freeVarTypes) {
			frame.locals[frame.fn.FreeVars[i]] = val
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
			c.parseInstr(frame, instr)
		}
		if frame.fn.Name() == "init" && len(block.Instrs) == 0 {
			c.builder.CreateRetVoid()
		}
	}

	// Resolve phi nodes
	for _, phi := range frame.phis {
		block := phi.ssa.Block()
		for i, edge := range phi.ssa.Edges {
			llvmVal := c.getValue(frame, edge)
			llvmBlock := frame.blockExits[block.Preds[i]]
			phi.llvm.AddIncoming([]llvm.Value{llvmVal}, []llvm.BasicBlock{llvmBlock})
		}
	}
}

func (c *Compiler) parseInstr(frame *Frame, instr ssa.Instruction) {
	if c.Debug {
		pos := c.ir.Program.Fset.Position(instr.Pos())
		c.builder.SetCurrentDebugLocation(uint(pos.Line), uint(pos.Column), frame.difunc, llvm.Metadata{})
	}

	switch instr := instr.(type) {
	case ssa.Value:
		if value, err := c.parseExpr(frame, instr); err != nil {
			// This expression could not be parsed. Add the error to the list
			// of diagnostics and continue with an undef value.
			// The resulting IR will be incorrect (but valid). However,
			// compilation can proceed which is useful because there may be
			// more compilation errors which can then all be shown together to
			// the user.
			c.diagnostics = append(c.diagnostics, err)
			frame.locals[instr] = llvm.Undef(c.getLLVMType(instr.Type()))
		} else {
			frame.locals[instr] = value
		}
	case *ssa.DebugRef:
		// ignore
	case *ssa.Defer:
		c.emitDefer(frame, instr)
	case *ssa.Go:
		if instr.Call.IsInvoke() {
			c.addError(instr.Pos(), "todo: go on method receiver")
			return
		}
		callee := instr.Call.StaticCallee()
		if callee == nil {
			c.addError(instr.Pos(), "todo: go on non-direct function (function pointer, etc.)")
			return
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
			params = append(params, c.getValue(frame, param))
		}
		if !calleeFn.IsExported() {
			params = append(params, llvm.Undef(c.i8ptrType)) // context parameter
			params = append(params, llvm.Undef(c.i8ptrType)) // parent coroutine handle
		}

		c.createCall(calleeValue, params, "")
	case *ssa.If:
		cond := c.getValue(frame, instr.Cond)
		block := instr.Block()
		blockThen := frame.blockEntries[block.Succs[0]]
		blockElse := frame.blockEntries[block.Succs[1]]
		c.builder.CreateCondBr(cond, blockThen, blockElse)
	case *ssa.Jump:
		blockJump := frame.blockEntries[instr.Block().Succs[0]]
		c.builder.CreateBr(blockJump)
	case *ssa.MapUpdate:
		m := c.getValue(frame, instr.Map)
		key := c.getValue(frame, instr.Key)
		value := c.getValue(frame, instr.Value)
		mapType := instr.Map.Type().Underlying().(*types.Map)
		c.emitMapUpdate(mapType.Key(), m, key, value, instr.Pos())
	case *ssa.Panic:
		value := c.getValue(frame, instr.X)
		c.createRuntimeCall("_panic", []llvm.Value{value}, "")
		c.builder.CreateUnreachable()
	case *ssa.Return:
		if len(instr.Results) == 0 {
			c.builder.CreateRetVoid()
		} else if len(instr.Results) == 1 {
			c.builder.CreateRet(c.getValue(frame, instr.Results[0]))
		} else {
			// Multiple return values. Put them all in a struct.
			retVal := c.getZeroValue(frame.fn.LLVMFn.Type().ElementType().ReturnType())
			for i, result := range instr.Results {
				val := c.getValue(frame, result)
				retVal = c.builder.CreateInsertValue(retVal, val, i, "")
			}
			c.builder.CreateRet(retVal)
		}
	case *ssa.RunDefers:
		c.emitRunDefers(frame)
	case *ssa.Send:
		c.emitChanSend(frame, instr)
	case *ssa.Store:
		llvmAddr := c.getValue(frame, instr.Addr)
		llvmVal := c.getValue(frame, instr.Val)
		c.emitNilCheck(frame, llvmAddr, "store")
		if c.targetData.TypeAllocSize(llvmVal.Type()) == 0 {
			// nothing to store
			return
		}
		c.builder.CreateStore(llvmVal, llvmAddr)
	default:
		c.addError(instr.Pos(), "unknown instruction: "+instr.String())
	}
}

func (c *Compiler) parseBuiltin(frame *Frame, args []ssa.Value, callName string, pos token.Pos) (llvm.Value, error) {
	switch callName {
	case "append":
		src := c.getValue(frame, args[0])
		elems := c.getValue(frame, args[1])
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
		value := c.getValue(frame, args[0])
		var llvmCap llvm.Value
		switch args[0].Type().(type) {
		case *types.Chan:
			// Channel. Buffered channels haven't been implemented yet so always
			// return 0.
			llvmCap = llvm.ConstInt(c.intType, 0, false)
		case *types.Slice:
			llvmCap = c.builder.CreateExtractValue(value, 2, "cap")
		default:
			return llvm.Value{}, c.makeError(pos, "todo: cap: unknown type")
		}
		if c.targetData.TypeAllocSize(llvmCap.Type()) < c.targetData.TypeAllocSize(c.intType) {
			llvmCap = c.builder.CreateZExt(llvmCap, c.intType, "len.int")
		}
		return llvmCap, nil
	case "close":
		c.emitChanClose(frame, args[0])
		return llvm.Value{}, nil
	case "complex":
		r := c.getValue(frame, args[0])
		i := c.getValue(frame, args[1])
		t := args[0].Type().Underlying().(*types.Basic)
		var cplx llvm.Value
		switch t.Kind() {
		case types.Float32:
			cplx = llvm.Undef(c.ctx.StructType([]llvm.Type{c.ctx.FloatType(), c.ctx.FloatType()}, false))
		case types.Float64:
			cplx = llvm.Undef(c.ctx.StructType([]llvm.Type{c.ctx.DoubleType(), c.ctx.DoubleType()}, false))
		default:
			return llvm.Value{}, c.makeError(pos, "unsupported type in complex builtin: "+t.String())
		}
		cplx = c.builder.CreateInsertValue(cplx, r, 0, "")
		cplx = c.builder.CreateInsertValue(cplx, i, 1, "")
		return cplx, nil
	case "copy":
		dst := c.getValue(frame, args[0])
		src := c.getValue(frame, args[1])
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
		m := c.getValue(frame, args[0])
		key := c.getValue(frame, args[1])
		return llvm.Value{}, c.emitMapDelete(args[1].Type(), m, key, pos)
	case "imag":
		cplx := c.getValue(frame, args[0])
		return c.builder.CreateExtractValue(cplx, 1, "imag"), nil
	case "len":
		value := c.getValue(frame, args[0])
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
			value := c.getValue(frame, arg)
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
		cplx := c.getValue(frame, args[0])
		return c.builder.CreateExtractValue(cplx, 0, "real"), nil
	case "recover":
		return c.createRuntimeCall("_recover", nil, ""), nil
	case "ssa:wrapnilchk":
		// TODO: do an actual nil check?
		return c.getValue(frame, args[0]), nil
	default:
		return llvm.Value{}, c.makeError(pos, "todo: builtin: "+callName)
	}
}

func (c *Compiler) parseFunctionCall(frame *Frame, args []ssa.Value, llvmFn, context llvm.Value, exported bool) llvm.Value {
	var params []llvm.Value
	for _, param := range args {
		params = append(params, c.getValue(frame, param))
	}

	if !exported {
		// This function takes a context parameter.
		// Add it to the end of the parameter list.
		params = append(params, context)

		// Parent coroutine handle.
		params = append(params, llvm.Undef(c.i8ptrType))
	}

	return c.createCall(llvmFn, params, "")
}

func (c *Compiler) parseCall(frame *Frame, instr *ssa.CallCommon) (llvm.Value, error) {
	if instr.IsInvoke() {
		fnCast, args := c.getInvokeCall(frame, instr)
		return c.createCall(fnCast, args, ""), nil
	}

	// Try to call the function directly for trivially static calls.
	if fn := instr.StaticCallee(); fn != nil {
		name := fn.RelString(nil)
		switch {
		case name == "device/arm.ReadRegister":
			return c.emitReadRegister(instr.Args)
		case name == "device/arm.Asm" || name == "device/avr.Asm":
			return c.emitAsm(instr.Args)
		case name == "device/arm.AsmFull" || name == "device/avr.AsmFull":
			return c.emitAsmFull(frame, instr)
		case strings.HasPrefix(name, "device/arm.SVCall"):
			return c.emitSVCall(frame, instr.Args)
		case strings.HasPrefix(name, "syscall.Syscall"):
			return c.emitSyscall(frame, instr)
		case strings.HasPrefix(name, "runtime/volatile.Load"):
			return c.emitVolatileLoad(frame, instr)
		case strings.HasPrefix(name, "runtime/volatile.Store"):
			return c.emitVolatileStore(frame, instr)
		}

		targetFunc := c.ir.GetFunction(fn)
		if targetFunc.LLVMFn.IsNil() {
			return llvm.Value{}, c.makeError(instr.Pos(), "undefined function: "+targetFunc.LinkName())
		}
		var context llvm.Value
		switch value := instr.Value.(type) {
		case *ssa.Function:
			// Regular function call. No context is necessary.
			context = llvm.Undef(c.i8ptrType)
		case *ssa.MakeClosure:
			// A call on a func value, but the callee is trivial to find. For
			// example: immediately applied functions.
			funcValue := c.getValue(frame, value)
			context = c.extractFuncContext(funcValue)
		default:
			panic("StaticCallee returned an unexpected value")
		}
		return c.parseFunctionCall(frame, instr.Args, targetFunc.LLVMFn, context, targetFunc.IsExported()), nil
	}

	// Builtin or function pointer.
	switch call := instr.Value.(type) {
	case *ssa.Builtin:
		return c.parseBuiltin(frame, instr.Args, call.Name(), instr.Pos())
	default: // function pointer
		value := c.getValue(frame, instr.Value)
		// This is a func value, which cannot be called directly. We have to
		// extract the function pointer and context first from the func value.
		funcPtr, context := c.decodeFuncValue(value, instr.Value.Type().Underlying().(*types.Signature))
		c.emitNilCheck(frame, funcPtr, "fpcall")
		return c.parseFunctionCall(frame, instr.Args, funcPtr, context, false), nil
	}
}

// getValue returns the LLVM value of a constant, function value, global, or
// already processed SSA expression.
func (c *Compiler) getValue(frame *Frame, expr ssa.Value) llvm.Value {
	switch expr := expr.(type) {
	case *ssa.Const:
		return c.parseConst(frame.fn.LinkName(), expr)
	case *ssa.Function:
		fn := c.ir.GetFunction(expr)
		if fn.IsExported() {
			c.addError(expr.Pos(), "cannot use an exported function as value: "+expr.String())
			return llvm.Undef(c.getLLVMType(expr.Type()))
		}
		return c.createFuncValue(fn.LLVMFn, llvm.Undef(c.i8ptrType), fn.Signature)
	case *ssa.Global:
		value := c.getGlobal(expr)
		if value.IsNil() {
			c.addError(expr.Pos(), "global not found: "+expr.RelString(nil))
			return llvm.Undef(c.getLLVMType(expr.Type()))
		}
		return value
	default:
		// other (local) SSA value
		if value, ok := frame.locals[expr]; ok {
			return value
		} else {
			// indicates a compiler bug
			panic("local has not been parsed: " + expr.String())
		}
	}
}

// parseExpr translates a Go SSA expression to a LLVM instruction.
func (c *Compiler) parseExpr(frame *Frame, expr ssa.Value) (llvm.Value, error) {
	if _, ok := frame.locals[expr]; ok {
		// sanity check
		panic("local has already been parsed: " + expr.String())
	}

	switch expr := expr.(type) {
	case *ssa.Alloc:
		typ := c.getLLVMType(expr.Type().Underlying().(*types.Pointer).Elem())
		if expr.Heap {
			size := c.targetData.TypeAllocSize(typ)
			// Calculate ^uintptr(0)
			maxSize := llvm.ConstNot(llvm.ConstInt(c.uintptrType, 0, false)).ZExtValue()
			if size > maxSize {
				// Size would be truncated if truncated to uintptr.
				return llvm.Value{}, c.makeError(expr.Pos(), fmt.Sprintf("value is too big (%v bytes)", size))
			}
			// TODO: escape analysis
			sizeValue := llvm.ConstInt(c.uintptrType, size, false)
			buf := c.createRuntimeCall("alloc", []llvm.Value{sizeValue}, expr.Comment)
			buf = c.builder.CreateBitCast(buf, llvm.PointerType(typ, 0), "")
			return buf, nil
		} else {
			buf := c.createEntryBlockAlloca(typ, expr.Comment)
			if c.targetData.TypeAllocSize(typ) != 0 {
				c.builder.CreateStore(c.getZeroValue(typ), buf) // zero-initialize var
			}
			return buf, nil
		}
	case *ssa.BinOp:
		x := c.getValue(frame, expr.X)
		y := c.getValue(frame, expr.Y)
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
		return c.getValue(frame, expr.X), nil
	case *ssa.ChangeType:
		// This instruction changes the type, but the underlying value remains
		// the same. This is often a no-op, but sometimes we have to change the
		// LLVM type as well.
		x := c.getValue(frame, expr.X)
		llvmType := c.getLLVMType(expr.Type())
		if x.Type() == llvmType {
			// Different Go type but same LLVM type (for example, named int).
			// This is the common case.
			return x, nil
		}
		// Figure out what kind of type we need to cast.
		switch llvmType.TypeKind() {
		case llvm.StructTypeKind:
			// Unfortunately, we can't just bitcast structs. We have to
			// actually create a new struct of the correct type and insert the
			// values from the previous struct in there.
			value := llvm.Undef(llvmType)
			for i := 0; i < llvmType.StructElementTypesCount(); i++ {
				field := c.builder.CreateExtractValue(x, i, "changetype.field")
				value = c.builder.CreateInsertValue(value, field, i, "changetype.struct")
			}
			return value, nil
		case llvm.PointerTypeKind:
			// This can happen with pointers to structs. This case is easy:
			// simply bitcast the pointer to the destination type.
			return c.builder.CreateBitCast(x, llvmType, "changetype.pointer"), nil
		default:
			return llvm.Value{}, errors.New("todo: unknown ChangeType type: " + expr.X.Type().String())
		}
	case *ssa.Const:
		panic("const is not an expression")
	case *ssa.Convert:
		x := c.getValue(frame, expr.X)
		return c.parseConvert(expr.X.Type(), expr.Type(), x, expr.Pos())
	case *ssa.Extract:
		if _, ok := expr.Tuple.(*ssa.Select); ok {
			return c.getChanSelectResult(frame, expr), nil
		}
		value := c.getValue(frame, expr.Tuple)
		return c.builder.CreateExtractValue(value, expr.Index, ""), nil
	case *ssa.Field:
		value := c.getValue(frame, expr.X)
		if s := expr.X.Type().Underlying().(*types.Struct); s.NumFields() > 2 && s.Field(0).Name() == "C union" {
			// Extract a field from a CGo union.
			// This could be done directly, but as this is a very infrequent
			// operation it's much easier to bitcast it through an alloca.
			resultType := c.getLLVMType(expr.Type())
			alloca, allocaPtr, allocaSize := c.createTemporaryAlloca(value.Type(), "union.alloca")
			c.builder.CreateStore(value, alloca)
			bitcast := c.builder.CreateBitCast(alloca, llvm.PointerType(resultType, 0), "union.bitcast")
			result := c.builder.CreateLoad(bitcast, "union.result")
			c.emitLifetimeEnd(allocaPtr, allocaSize)
			return result, nil
		}
		result := c.builder.CreateExtractValue(value, expr.Field, "")
		return result, nil
	case *ssa.FieldAddr:
		val := c.getValue(frame, expr.X)
		// Check for nil pointer before calculating the address, from the spec:
		// > For an operand x of type T, the address operation &x generates a
		// > pointer of type *T to x. [...] If the evaluation of x would cause a
		// > run-time panic, then the evaluation of &x does too.
		c.emitNilCheck(frame, val, "gep")
		if s := expr.X.Type().(*types.Pointer).Elem().Underlying().(*types.Struct); s.NumFields() > 2 && s.Field(0).Name() == "C union" {
			// This is not a regular struct but actually an union.
			// That simplifies things, as we can just bitcast the pointer to the
			// right type.
			ptrType := c.getLLVMType(expr.Type())
			return c.builder.CreateBitCast(val, ptrType, ""), nil
		} else {
			// Do a GEP on the pointer to get the field address.
			indices := []llvm.Value{
				llvm.ConstInt(c.ctx.Int32Type(), 0, false),
				llvm.ConstInt(c.ctx.Int32Type(), uint64(expr.Field), false),
			}
			return c.builder.CreateInBoundsGEP(val, indices, ""), nil
		}
	case *ssa.Function:
		panic("function is not an expression")
	case *ssa.Global:
		panic("global is not an expression")
	case *ssa.Index:
		array := c.getValue(frame, expr.X)
		index := c.getValue(frame, expr.Index)

		// Check bounds.
		arrayLen := expr.X.Type().(*types.Array).Len()
		arrayLenLLVM := llvm.ConstInt(c.uintptrType, uint64(arrayLen), false)
		c.emitLookupBoundsCheck(frame, arrayLenLLVM, index, expr.Index.Type())

		// Can't load directly from array (as index is non-constant), so have to
		// do it using an alloca+gep+load.
		alloca, allocaPtr, allocaSize := c.createTemporaryAlloca(array.Type(), "index.alloca")
		c.builder.CreateStore(array, alloca)
		zero := llvm.ConstInt(c.ctx.Int32Type(), 0, false)
		ptr := c.builder.CreateInBoundsGEP(alloca, []llvm.Value{zero, index}, "index.gep")
		result := c.builder.CreateLoad(ptr, "index.load")
		c.emitLifetimeEnd(allocaPtr, allocaSize)
		return result, nil
	case *ssa.IndexAddr:
		val := c.getValue(frame, expr.X)
		index := c.getValue(frame, expr.Index)

		// Get buffer pointer and length
		var bufptr, buflen llvm.Value
		switch ptrTyp := expr.X.Type().Underlying().(type) {
		case *types.Pointer:
			typ := expr.X.Type().Underlying().(*types.Pointer).Elem().Underlying()
			switch typ := typ.(type) {
			case *types.Array:
				bufptr = val
				buflen = llvm.ConstInt(c.uintptrType, uint64(typ.Len()), false)
				// Check for nil pointer before calculating the address, from
				// the spec:
				// > For an operand x of type T, the address operation &x
				// > generates a pointer of type *T to x. [...] If the
				// > evaluation of x would cause a run-time panic, then the
				// > evaluation of &x does too.
				c.emitNilCheck(frame, bufptr, "gep")
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
		c.emitLookupBoundsCheck(frame, buflen, index, expr.Index.Type())

		switch expr.X.Type().Underlying().(type) {
		case *types.Pointer:
			indices := []llvm.Value{
				llvm.ConstInt(c.ctx.Int32Type(), 0, false),
				index,
			}
			return c.builder.CreateInBoundsGEP(bufptr, indices, ""), nil
		case *types.Slice:
			return c.builder.CreateInBoundsGEP(bufptr, []llvm.Value{index}, ""), nil
		default:
			panic("unreachable")
		}
	case *ssa.Lookup:
		value := c.getValue(frame, expr.X)
		index := c.getValue(frame, expr.Index)
		switch xType := expr.X.Type().Underlying().(type) {
		case *types.Basic:
			// Value type must be a string, which is a basic type.
			if xType.Info()&types.IsString == 0 {
				panic("lookup on non-string?")
			}

			// Bounds check.
			length := c.builder.CreateExtractValue(value, 1, "len")
			c.emitLookupBoundsCheck(frame, length, index, expr.Index.Type())

			// Lookup byte
			buf := c.builder.CreateExtractValue(value, 0, "")
			bufPtr := c.builder.CreateInBoundsGEP(buf, []llvm.Value{index}, "")
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
		return c.parseMakeClosure(frame, expr)
	case *ssa.MakeInterface:
		val := c.getValue(frame, expr.X)
		return c.parseMakeInterface(val, expr.X.Type(), expr.Pos()), nil
	case *ssa.MakeMap:
		mapType := expr.Type().Underlying().(*types.Map)
		llvmKeyType := c.getLLVMType(mapType.Key().Underlying())
		llvmValueType := c.getLLVMType(mapType.Elem().Underlying())
		keySize := c.targetData.TypeAllocSize(llvmKeyType)
		valueSize := c.targetData.TypeAllocSize(llvmValueType)
		llvmKeySize := llvm.ConstInt(c.ctx.Int8Type(), keySize, false)
		llvmValueSize := llvm.ConstInt(c.ctx.Int8Type(), valueSize, false)
		sizeHint := llvm.ConstInt(c.uintptrType, 8, false)
		if expr.Reserve != nil {
			sizeHint = c.getValue(frame, expr.Reserve)
			var err error
			sizeHint, err = c.parseConvert(expr.Reserve.Type(), types.Typ[types.Uintptr], sizeHint, expr.Pos())
			if err != nil {
				return llvm.Value{}, err
			}
		}
		hashmap := c.createRuntimeCall("hashmapMake", []llvm.Value{llvmKeySize, llvmValueSize, sizeHint}, "")
		return hashmap, nil
	case *ssa.MakeSlice:
		sliceLen := c.getValue(frame, expr.Len)
		sliceCap := c.getValue(frame, expr.Cap)
		sliceType := expr.Type().Underlying().(*types.Slice)
		llvmElemType := c.getLLVMType(sliceType.Elem())
		elemSize := c.targetData.TypeAllocSize(llvmElemType)
		elemSizeValue := llvm.ConstInt(c.uintptrType, elemSize, false)

		// Calculate (^uintptr(0)) >> 1, which is the max value that fits in
		// uintptr if uintptr were signed.
		maxSize := llvm.ConstLShr(llvm.ConstNot(llvm.ConstInt(c.uintptrType, 0, false)), llvm.ConstInt(c.uintptrType, 1, false))
		if elemSize > maxSize.ZExtValue() {
			// This seems to be checked by the typechecker already, but let's
			// check it again just to be sure.
			return llvm.Value{}, c.makeError(expr.Pos(), fmt.Sprintf("slice element type is too big (%v bytes)", elemSize))
		}

		// Bounds checking.
		c.emitSliceBoundsCheck(frame, maxSize, sliceLen, sliceCap, expr.Len.Type().(*types.Basic), expr.Cap.Type().(*types.Basic))

		// Allocate the backing array.
		// TODO: escape analysis
		sliceCapCast, err := c.parseConvert(expr.Cap.Type(), types.Typ[types.Uintptr], sliceCap, expr.Pos())
		if err != nil {
			return llvm.Value{}, err
		}
		sliceSize := c.builder.CreateBinOp(llvm.Mul, elemSizeValue, sliceCapCast, "makeslice.cap")
		slicePtr := c.createRuntimeCall("alloc", []llvm.Value{sliceSize}, "makeslice.buf")
		slicePtr = c.builder.CreateBitCast(slicePtr, llvm.PointerType(llvmElemType, 0), "makeslice.array")

		// Extend or truncate if necessary. This is safe as we've already done
		// the bounds check.
		sliceLen, err = c.parseConvert(expr.Len.Type(), types.Typ[types.Uintptr], sliceLen, expr.Pos())
		if err != nil {
			return llvm.Value{}, err
		}
		sliceCap, err = c.parseConvert(expr.Cap.Type(), types.Typ[types.Uintptr], sliceCap, expr.Pos())
		if err != nil {
			return llvm.Value{}, err
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
		llvmRangeVal := c.getValue(frame, rangeVal)
		it := c.getValue(frame, expr.Iter)
		if expr.IsString {
			return c.createRuntimeCall("stringNext", []llvm.Value{llvmRangeVal, it}, "range.next"), nil
		} else { // map
			llvmKeyType := c.getLLVMType(rangeVal.Type().Underlying().(*types.Map).Key())
			llvmValueType := c.getLLVMType(rangeVal.Type().Underlying().(*types.Map).Elem())

			mapKeyAlloca, mapKeyPtr, mapKeySize := c.createTemporaryAlloca(llvmKeyType, "range.key")
			mapValueAlloca, mapValuePtr, mapValueSize := c.createTemporaryAlloca(llvmValueType, "range.value")
			ok := c.createRuntimeCall("hashmapNext", []llvm.Value{llvmRangeVal, it, mapKeyPtr, mapValuePtr}, "range.next")

			tuple := llvm.Undef(c.ctx.StructType([]llvm.Type{c.ctx.Int1Type(), llvmKeyType, llvmValueType}, false))
			tuple = c.builder.CreateInsertValue(tuple, ok, 0, "")
			tuple = c.builder.CreateInsertValue(tuple, c.builder.CreateLoad(mapKeyAlloca, ""), 1, "")
			tuple = c.builder.CreateInsertValue(tuple, c.builder.CreateLoad(mapValueAlloca, ""), 2, "")
			c.emitLifetimeEnd(mapKeyPtr, mapKeySize)
			c.emitLifetimeEnd(mapValuePtr, mapValueSize)
			return tuple, nil
		}
	case *ssa.Phi:
		phi := c.builder.CreatePHI(c.getLLVMType(expr.Type()), "")
		frame.phis = append(frame.phis, Phi{expr, phi})
		return phi, nil
	case *ssa.Range:
		var iteratorType llvm.Type
		switch typ := expr.X.Type().Underlying().(type) {
		case *types.Basic: // string
			iteratorType = c.getLLVMRuntimeType("stringIterator")
		case *types.Map:
			iteratorType = c.getLLVMRuntimeType("hashmapIterator")
		default:
			panic("unknown type in range: " + typ.String())
		}
		it, _, _ := c.createTemporaryAlloca(iteratorType, "range.it")
		c.builder.CreateStore(c.getZeroValue(iteratorType), it)
		return it, nil
	case *ssa.Select:
		return c.emitSelect(frame, expr), nil
	case *ssa.Slice:
		if expr.Max != nil {
			return llvm.Value{}, c.makeError(expr.Pos(), "todo: full slice expressions (with max): "+expr.Type().String())
		}
		value := c.getValue(frame, expr.X)

		var lowType, highType *types.Basic
		var low, high llvm.Value

		if expr.Low != nil {
			lowType = expr.Low.Type().Underlying().(*types.Basic)
			low = c.getValue(frame, expr.Low)
			if low.Type().IntTypeWidth() < c.uintptrType.IntTypeWidth() {
				if lowType.Info()&types.IsUnsigned != 0 {
					low = c.builder.CreateZExt(low, c.uintptrType, "")
				} else {
					low = c.builder.CreateSExt(low, c.uintptrType, "")
				}
			}
		} else {
			lowType = types.Typ[types.Uintptr]
			low = llvm.ConstInt(c.uintptrType, 0, false)
		}

		if expr.High != nil {
			highType = expr.High.Type().Underlying().(*types.Basic)
			high = c.getValue(frame, expr.High)
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

			c.emitSliceBoundsCheck(frame, llvmLen, low, high, lowType, highType)

			// Truncate ints bigger than uintptr. This is after the bounds
			// check so it's safe.
			if c.targetData.TypeAllocSize(high.Type()) > c.targetData.TypeAllocSize(c.uintptrType) {
				high = c.builder.CreateTrunc(high, c.uintptrType, "")
			}
			if c.targetData.TypeAllocSize(low.Type()) > c.targetData.TypeAllocSize(c.uintptrType) {
				low = c.builder.CreateTrunc(low, c.uintptrType, "")
			}

			sliceLen := c.builder.CreateSub(high, low, "slice.len")
			slicePtr := c.builder.CreateInBoundsGEP(value, indices, "slice.ptr")
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

			// Truncate ints bigger than uintptr. This is after the bounds
			// check so it's safe.
			if c.targetData.TypeAllocSize(low.Type()) > c.targetData.TypeAllocSize(c.uintptrType) {
				low = c.builder.CreateTrunc(low, c.uintptrType, "")
			}
			if c.targetData.TypeAllocSize(high.Type()) > c.targetData.TypeAllocSize(c.uintptrType) {
				high = c.builder.CreateTrunc(high, c.uintptrType, "")
			}

			newPtr := c.builder.CreateInBoundsGEP(oldPtr, []llvm.Value{low}, "")
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

			// Truncate ints bigger than uintptr. This is after the bounds
			// check so it's safe.
			if c.targetData.TypeAllocSize(low.Type()) > c.targetData.TypeAllocSize(c.uintptrType) {
				low = c.builder.CreateTrunc(low, c.uintptrType, "")
			}
			if c.targetData.TypeAllocSize(high.Type()) > c.targetData.TypeAllocSize(c.uintptrType) {
				high = c.builder.CreateTrunc(high, c.uintptrType, "")
			}

			newPtr := c.builder.CreateInBoundsGEP(oldPtr, []llvm.Value{low}, "")
			newLen := c.builder.CreateSub(high, low, "")
			str := llvm.Undef(c.getLLVMRuntimeType("_string"))
			str = c.builder.CreateInsertValue(str, newPtr, 0, "")
			str = c.builder.CreateInsertValue(str, newLen, 1, "")
			return str, nil

		default:
			return llvm.Value{}, c.makeError(expr.Pos(), "unknown slice type: "+typ.String())
		}
	case *ssa.TypeAssert:
		return c.parseTypeAssert(frame, expr), nil
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
				panic("binop on integer: " + op.String())
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
			case token.EQL: // ==
				return c.builder.CreateFCmp(llvm.FloatUEQ, x, y, ""), nil
			case token.NEQ: // !=
				return c.builder.CreateFCmp(llvm.FloatUNE, x, y, ""), nil
			case token.LSS: // <
				return c.builder.CreateFCmp(llvm.FloatULT, x, y, ""), nil
			case token.LEQ: // <=
				return c.builder.CreateFCmp(llvm.FloatULE, x, y, ""), nil
			case token.GTR: // >
				return c.builder.CreateFCmp(llvm.FloatUGT, x, y, ""), nil
			case token.GEQ: // >=
				return c.builder.CreateFCmp(llvm.FloatUGE, x, y, ""), nil
			default:
				panic("binop on float: " + op.String())
			}
		} else if typ.Info()&types.IsComplex != 0 {
			r1 := c.builder.CreateExtractValue(x, 0, "r1")
			r2 := c.builder.CreateExtractValue(y, 0, "r2")
			i1 := c.builder.CreateExtractValue(x, 1, "i1")
			i2 := c.builder.CreateExtractValue(y, 1, "i2")
			switch op {
			case token.EQL: // ==
				req := c.builder.CreateFCmp(llvm.FloatOEQ, r1, r2, "")
				ieq := c.builder.CreateFCmp(llvm.FloatOEQ, i1, i2, "")
				return c.builder.CreateAnd(req, ieq, ""), nil
			case token.NEQ: // !=
				req := c.builder.CreateFCmp(llvm.FloatOEQ, r1, r2, "")
				ieq := c.builder.CreateFCmp(llvm.FloatOEQ, i1, i2, "")
				neq := c.builder.CreateAnd(req, ieq, "")
				return c.builder.CreateNot(neq, ""), nil
			case token.ADD, token.SUB:
				var r, i llvm.Value
				switch op {
				case token.ADD:
					r = c.builder.CreateFAdd(r1, r2, "")
					i = c.builder.CreateFAdd(i1, i2, "")
				case token.SUB:
					r = c.builder.CreateFSub(r1, r2, "")
					i = c.builder.CreateFSub(i1, i2, "")
				default:
					panic("unreachable")
				}
				cplx := llvm.Undef(c.ctx.StructType([]llvm.Type{r.Type(), i.Type()}, false))
				cplx = c.builder.CreateInsertValue(cplx, r, 0, "")
				cplx = c.builder.CreateInsertValue(cplx, i, 1, "")
				return cplx, nil
			case token.MUL:
				// Complex multiplication follows the current implementation in
				// the Go compiler, with the difference that complex64
				// components are not first scaled up to float64 for increased
				// precision.
				// https://github.com/golang/go/blob/170b8b4b12be50eeccbcdadb8523fb4fc670ca72/src/cmd/compile/internal/gc/ssa.go#L2089-L2127
				// The implementation is as follows:
				//   r := real(a) * real(b) - imag(a) * imag(b)
				//   i := real(a) * imag(b) + imag(a) * real(b)
				// Note: this does NOT follow the C11 specification (annex G):
				// http://www.open-std.org/jtc1/sc22/wg14/www/docs/n1548.pdf#page=549
				// See https://github.com/golang/go/issues/29846 for a related
				// discussion.
				r := c.builder.CreateFSub(c.builder.CreateFMul(r1, r2, ""), c.builder.CreateFMul(i1, i2, ""), "")
				i := c.builder.CreateFAdd(c.builder.CreateFMul(r1, i2, ""), c.builder.CreateFMul(i1, r2, ""), "")
				cplx := llvm.Undef(c.ctx.StructType([]llvm.Type{r.Type(), i.Type()}, false))
				cplx = c.builder.CreateInsertValue(cplx, r, 0, "")
				cplx = c.builder.CreateInsertValue(cplx, i, 1, "")
				return cplx, nil
			case token.QUO:
				// Complex division.
				// Do this in a library call because it's too difficult to do
				// inline.
				switch r1.Type().TypeKind() {
				case llvm.FloatTypeKind:
					return c.createRuntimeCall("complex64div", []llvm.Value{x, y}, ""), nil
				case llvm.DoubleTypeKind:
					return c.createRuntimeCall("complex128div", []llvm.Value{x, y}, ""), nil
				default:
					panic("unexpected complex type")
				}
			default:
				panic("binop on complex: " + op.String())
			}
		} else if typ.Info()&types.IsBoolean != 0 {
			// Operations on booleans
			switch op {
			case token.EQL: // ==
				return c.builder.CreateICmp(llvm.IntEQ, x, y, ""), nil
			case token.NEQ: // !=
				return c.builder.CreateICmp(llvm.IntNE, x, y, ""), nil
			default:
				panic("binop on bool: " + op.String())
			}
		} else if typ.Kind() == types.UnsafePointer {
			// Operations on pointers
			switch op {
			case token.EQL: // ==
				return c.builder.CreateICmp(llvm.IntEQ, x, y, ""), nil
			case token.NEQ: // !=
				return c.builder.CreateICmp(llvm.IntNE, x, y, ""), nil
			default:
				panic("binop on pointer: " + op.String())
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
				panic("binop on string: " + op.String())
			}
		} else {
			return llvm.Value{}, c.makeError(pos, "todo: unknown basic type in binop: "+typ.String())
		}
	case *types.Signature:
		// Get raw scalars from the function value and compare those.
		// Function values may be implemented in multiple ways, but they all
		// have some way of getting a scalar value identifying the function.
		// This is safe: function pointers are generally not comparable
		// against each other, only against nil. So one of these has to be nil.
		x = c.extractFuncScalar(x)
		y = c.extractFuncScalar(y)
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
	case *types.Chan, *types.Map, *types.Pointer:
		// Maps are in general not comparable, but can be compared against nil
		// (which is a nil pointer). This means they can be trivially compared
		// by treating them as a pointer.
		// Channels behave as pointers in that they are equal as long as they
		// are created with the same call to make or if both are nil.
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
	default:
		return llvm.Value{}, c.makeError(pos, "todo: binop type: "+typ.String())
	}
}

func (c *Compiler) parseConst(prefix string, expr *ssa.Const) llvm.Value {
	switch typ := expr.Type().Underlying().(type) {
	case *types.Basic:
		llvmType := c.getLLVMType(typ)
		if typ.Info()&types.IsBoolean != 0 {
			b := constant.BoolVal(expr.Value)
			n := uint64(0)
			if b {
				n = 1
			}
			return llvm.ConstInt(llvmType, n, false)
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
			strObj := llvm.ConstNamedStruct(c.getLLVMRuntimeType("_string"), []llvm.Value{strPtr, strLen})
			return strObj
		} else if typ.Kind() == types.UnsafePointer {
			if !expr.IsNil() {
				value, _ := constant.Uint64Val(expr.Value)
				return llvm.ConstIntToPtr(llvm.ConstInt(c.uintptrType, value, false), c.i8ptrType)
			}
			return llvm.ConstNull(c.i8ptrType)
		} else if typ.Info()&types.IsUnsigned != 0 {
			n, _ := constant.Uint64Val(expr.Value)
			return llvm.ConstInt(llvmType, n, false)
		} else if typ.Info()&types.IsInteger != 0 { // signed
			n, _ := constant.Int64Val(expr.Value)
			return llvm.ConstInt(llvmType, uint64(n), true)
		} else if typ.Info()&types.IsFloat != 0 {
			n, _ := constant.Float64Val(expr.Value)
			return llvm.ConstFloat(llvmType, n)
		} else if typ.Kind() == types.Complex64 {
			r := c.parseConst(prefix, ssa.NewConst(constant.Real(expr.Value), types.Typ[types.Float32]))
			i := c.parseConst(prefix, ssa.NewConst(constant.Imag(expr.Value), types.Typ[types.Float32]))
			cplx := llvm.Undef(c.ctx.StructType([]llvm.Type{c.ctx.FloatType(), c.ctx.FloatType()}, false))
			cplx = c.builder.CreateInsertValue(cplx, r, 0, "")
			cplx = c.builder.CreateInsertValue(cplx, i, 1, "")
			return cplx
		} else if typ.Kind() == types.Complex128 {
			r := c.parseConst(prefix, ssa.NewConst(constant.Real(expr.Value), types.Typ[types.Float64]))
			i := c.parseConst(prefix, ssa.NewConst(constant.Imag(expr.Value), types.Typ[types.Float64]))
			cplx := llvm.Undef(c.ctx.StructType([]llvm.Type{c.ctx.DoubleType(), c.ctx.DoubleType()}, false))
			cplx = c.builder.CreateInsertValue(cplx, r, 0, "")
			cplx = c.builder.CreateInsertValue(cplx, i, 1, "")
			return cplx
		} else {
			panic("unknown constant of basic type: " + expr.String())
		}
	case *types.Chan:
		if expr.Value != nil {
			panic("expected nil chan constant")
		}
		return c.getZeroValue(c.getLLVMType(expr.Type()))
	case *types.Signature:
		if expr.Value != nil {
			panic("expected nil signature constant")
		}
		return c.getZeroValue(c.getLLVMType(expr.Type()))
	case *types.Interface:
		if expr.Value != nil {
			panic("expected nil interface constant")
		}
		// Create a generic nil interface with no dynamic type (typecode=0).
		fields := []llvm.Value{
			llvm.ConstInt(c.uintptrType, 0, false),
			llvm.ConstPointerNull(c.i8ptrType),
		}
		return llvm.ConstNamedStruct(c.getLLVMRuntimeType("_interface"), fields)
	case *types.Pointer:
		if expr.Value != nil {
			panic("expected nil pointer constant")
		}
		return llvm.ConstPointerNull(c.getLLVMType(typ))
	case *types.Slice:
		if expr.Value != nil {
			panic("expected nil slice constant")
		}
		elemType := c.getLLVMType(typ.Elem())
		llvmPtr := llvm.ConstPointerNull(llvm.PointerType(elemType, 0))
		llvmLen := llvm.ConstInt(c.uintptrType, 0, false)
		slice := c.ctx.ConstStruct([]llvm.Value{
			llvmPtr, // backing array
			llvmLen, // len
			llvmLen, // cap
		}, false)
		return slice
	case *types.Map:
		if !expr.IsNil() {
			// I believe this is not allowed by the Go spec.
			panic("non-nil map constant")
		}
		llvmType := c.getLLVMType(typ)
		return c.getZeroValue(llvmType)
	default:
		panic("unknown constant: " + expr.String())
	}
}

func (c *Compiler) parseConvert(typeFrom, typeTo types.Type, value llvm.Value, pos token.Pos) (llvm.Value, error) {
	llvmTypeFrom := value.Type()
	llvmTypeTo := c.getLLVMType(typeTo)

	// Conversion between unsafe.Pointer and uintptr.
	isPtrFrom := isPointer(typeFrom.Underlying())
	isPtrTo := isPointer(typeTo.Underlying())
	if isPtrFrom && !isPtrTo {
		return c.builder.CreatePtrToInt(value, llvmTypeTo, ""), nil
	} else if !isPtrFrom && isPtrTo {
		if !value.IsABinaryOperator().IsNil() && value.InstructionOpcode() == llvm.Add {
			// This is probably a pattern like the following:
			// unsafe.Pointer(uintptr(ptr) + index)
			// Used in functions like memmove etc. for lack of pointer
			// arithmetic. Convert it to real pointer arithmatic here.
			ptr := value.Operand(0)
			index := value.Operand(1)
			if !index.IsAPtrToIntInst().IsNil() {
				// Swap if necessary, if ptr and index are reversed.
				ptr, index = index, ptr
			}
			if !ptr.IsAPtrToIntInst().IsNil() {
				origptr := ptr.Operand(0)
				if origptr.Type() == c.i8ptrType {
					// This pointer can be calculated from the original
					// ptrtoint instruction with a GEP. The leftover inttoptr
					// instruction is trivial to optimize away.
					// Making it an in bounds GEP even though it's easy to
					// create a GEP that is not in bounds. However, we're
					// talking about unsafe code here so the programmer has to
					// be careful anyway.
					return c.builder.CreateInBoundsGEP(origptr, []llvm.Value{index}, ""), nil
				}
			}
		}
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
			} else if typeFrom.Info()&types.IsUnsigned != 0 { // if unsigned
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
			r := c.builder.CreateExtractValue(value, 0, "real.f64")
			i := c.builder.CreateExtractValue(value, 1, "imag.f64")
			r = c.builder.CreateFPTrunc(r, c.ctx.FloatType(), "real.f32")
			i = c.builder.CreateFPTrunc(i, c.ctx.FloatType(), "imag.f32")
			cplx := llvm.Undef(c.ctx.StructType([]llvm.Type{c.ctx.FloatType(), c.ctx.FloatType()}, false))
			cplx = c.builder.CreateInsertValue(cplx, r, 0, "")
			cplx = c.builder.CreateInsertValue(cplx, i, 1, "")
			return cplx, nil
		}

		if typeFrom.Kind() == types.Complex64 && typeTo.Kind() == types.Complex128 {
			// Conversion from complex64 to complex128.
			r := c.builder.CreateExtractValue(value, 0, "real.f32")
			i := c.builder.CreateExtractValue(value, 1, "imag.f32")
			r = c.builder.CreateFPExt(r, c.ctx.DoubleType(), "real.f64")
			i = c.builder.CreateFPExt(i, c.ctx.DoubleType(), "imag.f64")
			cplx := llvm.Undef(c.ctx.StructType([]llvm.Type{c.ctx.DoubleType(), c.ctx.DoubleType()}, false))
			cplx = c.builder.CreateInsertValue(cplx, r, 0, "")
			cplx = c.builder.CreateInsertValue(cplx, i, 1, "")
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

func (c *Compiler) parseUnOp(frame *Frame, unop *ssa.UnOp) (llvm.Value, error) {
	x := c.getValue(frame, unop.X)
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
		unop.X.Type().Underlying().(*types.Pointer).Elem()
		if c.targetData.TypeAllocSize(x.Type().ElementType()) == 0 {
			// zero-length data
			return c.getZeroValue(x.Type().ElementType()), nil
		} else if strings.HasSuffix(unop.X.String(), "$funcaddr") {
			// CGo function pointer. The cgo part has rewritten CGo function
			// pointers as stub global variables of the form:
			//     var C.add unsafe.Pointer
			// Instead of a load from the global, create a bitcast of the
			// function pointer itself.
			globalName := c.getGlobalInfo(unop.X.(*ssa.Global)).linkName
			name := globalName[:len(globalName)-len("$funcaddr")]
			fn := c.mod.NamedFunction(name)
			if fn.IsNil() {
				return llvm.Value{}, c.makeError(unop.Pos(), "cgo function not found: "+name)
			}
			return c.builder.CreateBitCast(fn, c.i8ptrType, ""), nil
		} else {
			c.emitNilCheck(frame, x, "deref")
			load := c.builder.CreateLoad(x, "")
			return load, nil
		}
	case token.XOR: // ^x, toggle all bits in integer
		return c.builder.CreateXor(x, llvm.ConstInt(x.Type(), ^uint64(0), false), ""), nil
	case token.ARROW: // <-x, receive from channel
		return c.emitChanRecv(frame, unop), nil
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
