
package main

import (
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/build"
	"go/constant"
	"go/token"
	"go/types"
	"os"
	"sort"
	"strings"

	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
	"llvm.org/llvm/bindings/go/llvm"
)

// Silently ignore this instruction.
var ErrCGoIgnore = errors.New("cgo: ignore")

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
	stringLenType   llvm.Type
	stringType      llvm.Type
	interfaceType   llvm.Type
	typeassertType  llvm.Type
	panicFunc       llvm.Value
	boundsCheckFunc llvm.Value
	printstringFunc llvm.Value
	printintFunc    llvm.Value
	printuintFunc   llvm.Value
	printbyteFunc   llvm.Value
	printspaceFunc  llvm.Value
	printnlFunc     llvm.Value
	memsetIntrinsic llvm.Value
	itfTypeNumbers  map[types.Type]uint64
	itfTypes        []types.Type
}

type Frame struct {
	pkgPrefix string
	llvmFn    llvm.Value
	params    map[*ssa.Parameter]int   // arguments to the function
	locals    map[ssa.Value]llvm.Value // local variables
	blocks    map[*ssa.BasicBlock]llvm.BasicBlock
	phis      []Phi
}

type Phi struct {
	ssa  *ssa.Phi
	llvm llvm.Value
}

func NewCompiler(pkgName, triple string) (*Compiler, error) {
	c := &Compiler{
		triple:         triple,
		itfTypeNumbers: make(map[types.Type]uint64),
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

	// Go string: tuple of (len, ptr)
	c.stringType = llvm.StructType([]llvm.Type{c.stringLenType, llvm.PointerType(llvm.Int8Type(), 0)}, false)

	// Go interface: tuple of (type, ptr)
	c.interfaceType = llvm.StructType([]llvm.Type{llvm.Int32Type(), llvm.PointerType(llvm.Int8Type(), 0)}, false)

	// Go typeassert result: tuple of (ptr, bool)
	c.typeassertType = llvm.StructType([]llvm.Type{llvm.PointerType(llvm.Int8Type(), 0), llvm.Int1Type()}, false)

	panicType := llvm.FunctionType(llvm.VoidType(), []llvm.Type{c.interfaceType}, false)
	c.panicFunc = llvm.AddFunction(c.mod, "runtime._panic", panicType)

	boundsCheckType := llvm.FunctionType(llvm.VoidType(), []llvm.Type{llvm.Int1Type()}, false)
	c.boundsCheckFunc = llvm.AddFunction(c.mod, "runtime.boundsCheck", boundsCheckType)

	printstringType := llvm.FunctionType(llvm.VoidType(), []llvm.Type{c.stringType}, false)
	c.printstringFunc = llvm.AddFunction(c.mod, "runtime.printstring", printstringType)
	printintType := llvm.FunctionType(llvm.VoidType(), []llvm.Type{c.intType}, false)
	c.printintFunc = llvm.AddFunction(c.mod, "runtime.printint", printintType)
	printuintType := llvm.FunctionType(llvm.VoidType(), []llvm.Type{c.intType}, false)
	c.printuintFunc = llvm.AddFunction(c.mod, "runtime.printuint", printuintType)
	printbyteType := llvm.FunctionType(llvm.VoidType(), []llvm.Type{llvm.Int8Type()}, false)
	c.printbyteFunc = llvm.AddFunction(c.mod, "runtime.printbyte", printbyteType)
	printspaceType := llvm.FunctionType(llvm.VoidType(), nil, false)
	c.printspaceFunc = llvm.AddFunction(c.mod, "runtime.printspace", printspaceType)
	printnlType := llvm.FunctionType(llvm.VoidType(), nil, false)
	c.printnlFunc = llvm.AddFunction(c.mod, "runtime.printnl", printnlType)

	// Intrinsic functions
	memsetType := llvm.FunctionType(
		llvm.VoidType(), []llvm.Type{
			llvm.PointerType(llvm.Int8Type(), 0),
			llvm.Int8Type(),
			llvm.Int32Type(),
			llvm.Int1Type(),
		}, false)
	c.memsetIntrinsic = llvm.AddFunction(c.mod, "llvm.memset.p0i8.i32", memsetType)

	return c, nil
}

func (c *Compiler) Parse(pkgName string) error {
	tripleSplit := strings.Split(c.triple, "-")

	config := loader.Config {
		// TODO: TypeChecker.Sizes
		Build: &build.Context {
			GOARCH:      tripleSplit[0],
			GOOS:        tripleSplit[2],
			GOROOT:      ".",
			CgoEnabled:  true,
			UseAllFiles: false,
			Compiler:    "gc", // must be one of the recognized compilers
			BuildTags:   []string{"tgo"},
		},
		AllowErrors: true,
	}
	config.Import("runtime")
	config.Import(pkgName)
	lprogram, err := config.Load()
	if err != nil {
		return err
	}

	// TODO: pick the error of the first package, not a random package
	for _, pkgInfo := range lprogram.AllPackages {
		fmt.Println("package:", pkgInfo.Pkg.Name())
		if len(pkgInfo.Errors) != 0 {
			return pkgInfo.Errors[0]
		}
	}

	program := ssautil.CreateProgram(lprogram, ssa.SanityCheckFunctions | ssa.BareInits)
	program.Build()
	// TODO: order of packages is random
	for _, pkg := range program.AllPackages() {
		fmt.Println("package:", pkg.Pkg.Path())

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

			pkgPrefix := pkg.Pkg.Path()
			if pkg.Pkg.Name() == "main" {
				pkgPrefix = "main"
			}

			switch member := member.(type) {
			case *ssa.Function:
				frame, err := c.parseFuncDecl(pkgPrefix, member)
				if err != nil {
					return err
				}
				frames[member] = frame
			case *ssa.NamedConst:
				// Ignore package-level untyped constants. The SSA form doesn't
				// need them.
			case *ssa.Global:
				typ, err := c.getLLVMType(member.Type())
				if err != nil {
					return err
				}
				global := llvm.AddGlobal(c.mod, typ, pkgPrefix + "." +  member.Name())
				if ast.IsExported(member.Name()) {
					global.SetLinkage(llvm.PrivateLinkage)
				}
			case *ssa.Type:
				ms := program.MethodSets.MethodSet(member.Type())
				for i := 0; i < ms.Len(); i++ {
					fn := program.MethodValue(ms.At(i))
					frame, err := c.parseFuncDecl(pkgPrefix, fn)
					if err != nil {
						return err
					}
					frames[fn] = frame
				}
			default:
				return errors.New("todo: member: " + fmt.Sprintf("%#v", member))
			}
		}

		// Now, add definitions to those declarations.
		for _, name := range memberNames {
			member := pkg.Members[name]
			fmt.Println("member:", member.Token(), member)

			switch member := member.(type) {
			case *ssa.Function:
				if strings.HasPrefix(name, "_Cfunc_") {
					// CGo function. Don't implement it's body.
					continue
				}
				if member.Blocks == nil {
					continue // external function
				}
				err := c.parseFunc(frames[member], member)
				if err != nil {
					return err
				}
			case *ssa.Type:
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

func (c *Compiler) getLLVMType(goType types.Type) (llvm.Type, error) {
	fmt.Println("        type:", goType)
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
			return c.stringType, nil
		case types.UnsafePointer:
			return llvm.PointerType(llvm.Int8Type(), 0), nil
		default:
			return llvm.Type{}, errors.New("todo: unknown basic type: " + fmt.Sprintf("%#v", typ))
		}
	case *types.Interface:
		return c.interfaceType, nil
	case *types.Named:
		return c.getLLVMType(typ.Underlying())
	case *types.Pointer:
		ptrTo, err := c.getLLVMType(typ.Elem())
		if err != nil {
			return llvm.Type{}, err
		}
		return llvm.PointerType(ptrTo, 0), nil
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

func (c *Compiler) getInterfaceType(typ types.Type) llvm.Value {
	if _, ok := c.itfTypeNumbers[typ]; !ok {
		num := uint64(len(c.itfTypes))
		c.itfTypes = append(c.itfTypes, typ)
		c.itfTypeNumbers[typ] = num
	}
	return llvm.ConstInt(llvm.Int32Type(), c.itfTypeNumbers[typ], false)
}

func (c *Compiler) getFunctionName(pkgPrefix string, fn *ssa.Function) string {
	if fn.Signature.Recv() != nil {
		// Method on a defined type.
		typeName := fn.Params[0].Type().(*types.Named).Obj().Name()
		return pkgPrefix + "." + typeName + "." + fn.Name()
	} else {
		// Bare function.
		return pkgPrefix + "." + fn.Name()
	}
}

func (c *Compiler) parseFuncDecl(pkgPrefix string, f *ssa.Function) (*Frame, error) {
	name := c.getFunctionName(pkgPrefix, f)
	if strings.HasPrefix(name, pkgPrefix + "._Cfunc_") {
		// CGo wrapper declaration.
		// Don't wrap the function, instead declare it.
		name = name[len(pkgPrefix + "._Cfunc_"):]
	}

	frame := &Frame{
		pkgPrefix: pkgPrefix,
		params:    make(map[*ssa.Parameter]int),
		locals:    make(map[ssa.Value]llvm.Value),
		blocks:    make(map[*ssa.BasicBlock]llvm.BasicBlock),
	}

	var retType llvm.Type
	if f.Signature.Results() == nil {
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
	for i, param := range f.Params {
		paramType, err := c.getLLVMType(param.Type())
		if err != nil {
			return nil, err
		}
		paramTypes = append(paramTypes, paramType)
		frame.params[param] = i
	}

	fnType := llvm.FunctionType(retType, paramTypes, false)

	frame.llvmFn = c.mod.NamedFunction(name)
	if frame.llvmFn.IsNil() {
		frame.llvmFn = llvm.AddFunction(c.mod, name, fnType)
	}
	return frame, nil
}

func (c *Compiler) parseFunc(frame *Frame, f *ssa.Function) error {
	// Pre-create all basic blocks in the function.
	for _, block := range f.DomPreorder() {
		llvmBlock := c.ctx.AddBasicBlock(frame.llvmFn, block.Comment)
		frame.blocks[block] = llvmBlock
	}

	// Load function parameters
	for _, param := range f.Params {
		llvmParam := frame.llvmFn.Param(frame.params[param])
		frame.locals[param] = llvmParam
	}

	// Fill those blocks with instructions.
	for _, block := range f.DomPreorder() {
		c.builder.SetInsertPointAtEnd(frame.blocks[block])
		for _, instr := range block.Instrs {
			fmt.Printf("  instr: %v\n", instr)
			err := c.parseInstr(frame, instr)
			if err == ErrCGoIgnore {
				continue // ignore irrelevant CGo instruction
			}
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
		c.builder.CreateCall(c.panicFunc, []llvm.Value{value}, "")
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
			return errors.New("todo: return value")
		}
	case *ssa.Store:
		addr, err := c.parseExpr(frame, instr.Addr)
		if err != nil {
			return err
		}
		val, err := c.parseExpr(frame, instr.Val)
		if err != nil {
			return err
		}
		c.builder.CreateStore(val, addr)
		return nil
	default:
		return errors.New("unknown instruction: " + fmt.Sprintf("%#v", instr))
	}
}

func (c *Compiler) parseBuiltin(frame *Frame, args []ssa.Value, callName string) (llvm.Value, error) {
	fmt.Printf("    builtin: %v\n", callName)

	switch callName {
	case "print", "println":
		for i, arg := range args {
			if i >= 1 {
				c.builder.CreateCall(c.printspaceFunc, nil, "")
			}
			fmt.Printf("    arg: %s\n", arg);
			value, err := c.parseExpr(frame, arg)
			if err != nil {
				return llvm.Value{}, err
			}
			switch typ := arg.Type().(type) {
			case *types.Basic:
				switch typ.Kind() {
				case types.Uint8:
					c.builder.CreateCall(c.printbyteFunc, []llvm.Value{value}, "")
				case types.Int, types.Int32: // TODO: assumes a 32-bit int type
					c.builder.CreateCall(c.printintFunc, []llvm.Value{value}, "")
				case types.Uint, types.Uint32:
					c.builder.CreateCall(c.printuintFunc, []llvm.Value{value}, "")
				case types.String:
					c.builder.CreateCall(c.printstringFunc, []llvm.Value{value}, "")
				default:
					return llvm.Value{}, errors.New("unknown basic arg type: " + fmt.Sprintf("%#v", typ))
				}
			default:
				return llvm.Value{}, errors.New("unknown arg type: " + fmt.Sprintf("%#v", typ))
			}
		}
		if callName == "println" {
			c.builder.CreateCall(c.printnlFunc, nil, "")
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

func (c *Compiler) parseFunctionCall(frame *Frame, call *ssa.CallCommon, fn *ssa.Function) (llvm.Value, error) {
	fmt.Printf("    function: %s\n", fn)

	name := c.getFunctionName(frame.pkgPrefix, fn)
	if strings.HasPrefix(name, frame.pkgPrefix + "._Cfunc_") {
		// Call C function directly.
		name = name[len(frame.pkgPrefix + "._Cfunc_"):]
	}

	target := c.mod.NamedFunction(name)
	if target.IsNil() {
		return llvm.Value{}, errors.New("undefined function: " + name)
	}

	var params []llvm.Value
	for _, param := range call.Args {
		val, err := c.parseExpr(frame, param)
		if err != nil {
			return llvm.Value{}, err
		}
		params = append(params, val)
	}

	return c.builder.CreateCall(target, params, ""), nil
}

func (c *Compiler) parseCall(frame *Frame, instr *ssa.Call) (llvm.Value, error) {
	fmt.Printf("    call: %s\n", instr)

	switch call := instr.Common().Value.(type) {
	case *ssa.Builtin:
		return c.parseBuiltin(frame, instr.Common().Args, call.Name())
	case *ssa.Function:
		return c.parseFunctionCall(frame, instr.Common(), call)
	default:
		return llvm.Value{}, errors.New("todo: unknown call type: " + fmt.Sprintf("%#v", call))
	}
}

func (c *Compiler) parseExpr(frame *Frame, expr ssa.Value) (llvm.Value, error) {
	fmt.Printf("      expr: %v\n", expr)

	if frame != nil {
		if value, ok := frame.locals[expr]; ok {
			// Value is a local variable that has already been computed.
			fmt.Println("        from local var")
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
			buf = c.builder.CreateMalloc(typ, expr.Comment)
		} else {
			buf = c.builder.CreateAlloca(typ, expr.Comment)
		}
		width := c.targetData.TypeAllocSize(typ)
		if err != nil {
			return llvm.Value{}, err
		}
		llvmWidth := llvm.ConstInt(llvm.Int32Type(), width, false)
		bufBytes := c.builder.CreateBitCast(buf, llvm.PointerType(llvm.Int8Type(), 0), "")
		c.builder.CreateCall(
			c.memsetIntrinsic,
			[]llvm.Value{
				bufBytes,
				llvm.ConstInt(llvm.Int8Type(), 0, false), // value to set (zero)
				llvmWidth,                                // size to zero
				llvm.ConstInt(llvm.Int1Type(), 0, false), // volatile
			}, "")
		return buf, nil
	case *ssa.BinOp:
		return c.parseBinOp(frame, expr)
	case *ssa.Call:
		return c.parseCall(frame, expr)
	case *ssa.Const:
		return c.parseConst(expr)
	case *ssa.Convert:
		return c.parseConvert(frame, expr)
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
	case *ssa.Global:
		if strings.HasPrefix(expr.Name(), "__cgofn__cgo_") || strings.HasPrefix(expr.Name(), "_cgo_") {
			return llvm.Value{}, ErrCGoIgnore
		}
		value := c.mod.NamedGlobal(expr.Name())
		if value.IsNil() {
			return llvm.Value{}, errors.New("global not found: " + expr.Name())
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

		// Bounds check
		// TODO: inline, and avoid if possible
		constZero := llvm.ConstInt(c.intType, 0, false)
		isNegative := c.builder.CreateICmp(llvm.IntSLT, index, constZero, "") // index < 0
		isTooBig := c.builder.CreateICmp(llvm.IntSGE, index, buflen, "") // index >= len(value)
		isOverflow := c.builder.CreateOr(isNegative, isTooBig, "")
		c.builder.CreateCall(c.boundsCheckFunc, []llvm.Value{isOverflow}, "")

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

		// Bounds check
		// TODO: inline, and avoid if possible
		if frame.llvmFn.Name() != "runtime.boundsCheck" {
			constZero := llvm.ConstInt(c.intType, 0, false)
			isNegative := c.builder.CreateICmp(llvm.IntSLT, index, constZero, "") // index < 0
			strlen, err := c.parseBuiltin(frame, []ssa.Value{expr.X}, "len")
			if err != nil {
				return llvm.Value{}, err // shouldn't happen
			}
			isTooBig := c.builder.CreateICmp(llvm.IntSGE, index, strlen, "") // index >= len(value)
			isOverflow := c.builder.CreateOr(isNegative, isTooBig, "")
			c.builder.CreateCall(c.boundsCheckFunc, []llvm.Value{isOverflow}, "")
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
		switch typ := expr.X.Type().(type) {
		case *types.Basic:
			itfValueType := llvm.PointerType(llvm.Int8Type(), 0)
			if typ.Info() & types.IsInteger != 0 { // TODO: 64-bit int on 32-bit platform
				itfValue = c.builder.CreateIntToPtr(val, itfValueType, "")
			} else if typ.Kind() == types.String {
				// TODO: escape analysis
				itfValue = c.builder.CreateMalloc(c.stringType, "")
				c.builder.CreateStore(val, itfValue)
				itfValue = c.builder.CreateBitCast(itfValue, itfValueType, "")
			} else {
				return llvm.Value{}, errors.New("todo: make interface: unknown basic type")
			}
		default:
			return llvm.Value{}, errors.New("todo: make interface: unknown type")
		}
		itfType := c.getInterfaceType(expr.X.Type())
		itf := c.ctx.ConstStruct([]llvm.Value{itfType, llvm.Undef(llvm.PointerType(llvm.Int8Type(), 0))}, false)
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
		switch typ := expr.AssertedType.(type) {
		case *types.Basic:
			if typ.Info() & types.IsInteger != 0 {
				value = c.builder.CreatePtrToInt(valuePtr, assertedType, "")
			} else if typ.Kind() == types.String {
				valueStringPtr := c.builder.CreateBitCast(valuePtr, llvm.PointerType(c.stringType, 0), "")
				value = c.builder.CreateLoad(valueStringPtr, "")
			} else {
				return llvm.Value{}, errors.New("todo: typeassert: unknown basic type")
			}
		default:
			return llvm.Value{}, errors.New("todo: typeassert: unknown type")
		}
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
	signed := binop.X.Type().(*types.Basic).Info() & types.IsUnsigned == 0
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
	switch expr.Value.Kind() {
	case constant.String:
		str := constant.StringVal(expr.Value)
		strLen := llvm.ConstInt(c.stringLenType, uint64(len(str)), false)
		strPtr := c.builder.CreateGlobalStringPtr(str, ".str") // TODO: remove \0 at end
		strObj := llvm.ConstStruct([]llvm.Value{strLen, strPtr}, false)
		return strObj, nil
	case constant.Int:
		return c.parseConstInt(expr, expr.Type())
	default:
		return llvm.Value{}, errors.New("todo: unknown constant")
	}
}

func (c *Compiler) parseConstInt(expr *ssa.Const, typ types.Type) (llvm.Value, error) {
	switch typ := typ.(type) {
	case *types.Basic:
		llvmType, err := c.getLLVMType(typ)
		if err != nil {
			return llvm.Value{}, err
		}
		if typ.Info() & types.IsUnsigned != 0 || typ.Info() & types.IsBoolean != 0 {
			n, _ := constant.Uint64Val(expr.Value)
			return llvm.ConstInt(llvmType, n, false), nil
		} else if typ.Info() & types.IsInteger != 0 { // signed
			n, _ := constant.Int64Val(expr.Value)
			return llvm.ConstInt(llvmType, uint64(n), true), nil
		} else {
			return llvm.Value{}, errors.New("unknown integer constant")
		}
	case *types.Named:
		return c.parseConstInt(expr, typ.Underlying())
	default:
		return llvm.Value{}, errors.New("todo: unknown constant: " + fmt.Sprintf("%#v", typ))
	}
}

func (c *Compiler) parseConvert(frame *Frame, expr *ssa.Convert) (llvm.Value, error) {
	value, err := c.parseExpr(frame, expr.X)
	if err != nil {
		return value, nil
	}

	typeFrom, err := c.getLLVMType(expr.X.Type())
	if err != nil {
		return llvm.Value{}, err
	}
	sizeFrom := c.targetData.TypeAllocSize(typeFrom)
	typeTo, err := c.getLLVMType(expr.Type())
	if err != nil {
		return llvm.Value{}, err
	}
	sizeTo := c.targetData.TypeAllocSize(typeTo)

	if sizeFrom > sizeTo {
		return c.builder.CreateTrunc(value, typeTo, ""), nil
	} else if sizeFrom == sizeTo {
		return c.builder.CreateBitCast(value, typeTo, ""), nil
	} else { // sizeFrom < sizeTo: extend
		switch typ := expr.X.Type().(type) { // typeFrom
		case *types.Basic:
			if typ.Info() & types.IsInteger == 0 { // if not integer
				return llvm.Value{}, errors.New("todo: convert: extend non-integer type")
			}
			if typ.Info() & types.IsUnsigned != 0 { // if unsigned
				return c.builder.CreateZExt(value, typeTo, ""), nil
			} else {
				return c.builder.CreateSExt(value, typeTo, ""), nil
			}
		default:
			return llvm.Value{}, errors.New("todo: convert: extend non-basic type")
		}
	}
}

func (c *Compiler) parseUnOp(frame *Frame, unop *ssa.UnOp) (llvm.Value, error) {
	x, err := c.parseExpr(frame, unop.X)
	if err != nil {
		return llvm.Value{}, err
	}
	switch unop.Op {
	case token.NOT: // !
		return c.builder.CreateNot(x, ""), nil
	case token.SUB: // -num
		return c.builder.CreateSub(llvm.ConstInt(x.Type(), 0, false), x, ""), nil
	case token.MUL: // *ptr, dereference pointer
		return c.builder.CreateLoad(x, ""), nil
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

func (c *Compiler) Optimize(optLevel int) {
	builder := llvm.NewPassManagerBuilder()
	defer builder.Dispose()
	builder.SetOptLevel(optLevel)
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
	buf, err := c.machine.EmitToMemoryBuffer(c.mod, llvm.ObjectFile)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	f.Write(buf.Bytes())
	f.Close()
	return nil
}

// Helper function for Compiler object.
func Compile(pkgName, outpath, target string, printIR bool) error {
	c, err := NewCompiler(pkgName, target)
	if err != nil {
		return err
	}

	parseErr := c.Parse(pkgName)
	if printIR {
		fmt.Println(c.IR())
	}
	if parseErr != nil {
		return parseErr
	}

	if err := c.Verify(); err != nil {
		return err
	}
	c.Optimize(2)
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
	target := flag.String("target", llvm.DefaultTargetTriple(), "LLVM target")
	printIR := flag.Bool("printir", false, "print LLVM IR after optimizing")

	flag.Parse()

	if *outpath == "" || flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "usage: %s [-printir] [-target=<target>] -o <output> <input>", os.Args[0])
		flag.PrintDefaults()
		return
	}

	err := Compile(flag.Args()[0], *outpath, *target, *printIR)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
