
package main

import (
	"errors"
	"flag"
	"fmt"
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

func init() {
	llvm.InitializeAllTargets()
	llvm.InitializeAllTargetMCs()
	llvm.InitializeAllTargetInfos()
	llvm.InitializeAllAsmParsers()
	llvm.InitializeAllAsmPrinters()
}

type Compiler struct {
	mod             llvm.Module
	ctx             llvm.Context
	builder         llvm.Builder
	machine         llvm.TargetMachine
	intType         llvm.Type
	stringLenType   llvm.Type
	stringType      llvm.Type
	stringPtrType   llvm.Type
	printstringFunc llvm.Value
	printintFunc    llvm.Value
	printspaceFunc  llvm.Value
	printnlFunc     llvm.Value
	funcs           map[*ssa.Function]*Function
}

type Function struct {
	name   string
	params map[*ssa.Parameter]int
}

func NewCompiler(path, triple string) (*Compiler, error) {
	c := &Compiler{
		funcs: make(map[*ssa.Function]*Function),
	}

	target, err := llvm.GetTargetFromTriple(triple)
	if err != nil {
		return nil, err
	}
	c.machine = target.CreateTargetMachine(triple, "", "", llvm.CodeGenLevelDefault, llvm.RelocDefault, llvm.CodeModelDefault)

	c.mod = llvm.NewModule(path)
	c.ctx = c.mod.Context()
	c.builder = c.ctx.NewBuilder()

	// Depends on platform (32bit or 64bit), but fix it here for now.
	c.intType = llvm.Int32Type()
	c.stringLenType = llvm.Int32Type()

	// Length-prefixed string.
	c.stringType = llvm.StructType([]llvm.Type{c.stringLenType, llvm.ArrayType(llvm.Int8Type(), 0)}, false)
	c.stringPtrType = llvm.PointerType(c.stringType, 0)

	printstringType := llvm.FunctionType(llvm.VoidType(), []llvm.Type{c.stringPtrType}, false)
	c.printstringFunc = llvm.AddFunction(c.mod, "__go_printstring", printstringType)
	printintType := llvm.FunctionType(llvm.VoidType(), []llvm.Type{c.intType}, false)
	c.printintFunc = llvm.AddFunction(c.mod, "__go_printint", printintType)
	printspaceType := llvm.FunctionType(llvm.VoidType(), nil, false)
	c.printspaceFunc = llvm.AddFunction(c.mod, "__go_printspace", printspaceType)
	printnlType := llvm.FunctionType(llvm.VoidType(), nil, false)
	c.printnlFunc = llvm.AddFunction(c.mod, "__go_printnl", printnlType)

	return c, nil
}

func (c *Compiler) Parse(path string) error {
	config := loader.Config {
		// TODO: TypeChecker.Sizes
		// TODO: Build (build.Context) - GOOS, GOARCH, GOPATH, etc
	}
	config.CreateFromFilenames("main", path)
	lprogram, err := config.Load()
	if err != nil {
		return err
	}

	program := ssautil.CreateProgram(lprogram, ssa.SanityCheckFunctions)
	program.Build()
	for _, pkg := range program.AllPackages() {
		fmt.Println("package:", pkg.Pkg.Name())

		// Make sure we're walking through all members in a constant order every
		// run.
		memberNames := make([]string, 0)
		for name := range pkg.Members {
			memberNames = append(memberNames, name)
		}
		sort.Strings(memberNames)

		// First, build all function declarations.
		for _, name := range memberNames {
			member := pkg.Members[name]
			if member, ok := member.(*ssa.Function); ok {
				err := c.parseFuncDecl(pkg.Pkg.Name(), member)
				if err != nil {
					return err
				}
			}
		}

		// Now, add definitions to those declarations.
		for _, name := range memberNames {
			member := pkg.Members[name]
			fmt.Println("member:", member.Token(), member)
			if name == "init" {
				continue
			}
			switch member := member.(type) {
			case *ssa.Function:
				err := c.parseFunc(pkg.Pkg.Name(), member)
				if err != nil {
					return err
				}

			default:
				fmt.Println("  TODO")
			}
		}
	}

	return nil
}

func (c *Compiler) parseFuncDecl(pkgName string, f *ssa.Function) error {
	name := pkgName + "." + f.Name()
	c.funcs[f] = &Function{
		name:   name,
		params: make(map[*ssa.Parameter]int),
	}

	var retType llvm.Type
	if f.Signature.Results() == nil {
		retType = llvm.VoidType()
	} else if f.Signature.Results().Len() == 1 {
		result := f.Signature.Results().At(0)
		switch typ := result.Type().(type) {
		case *types.Basic:
			switch typ.Kind() {
			case types.Int:
				retType = c.intType
			case types.Int32:
				retType = llvm.Int32Type()
			default:
				return errors.New("todo: unknown basic return type")
			}
		default:
			return errors.New("todo: unknown return type")
		}
	} else {
		return errors.New("todo: return values")
	}

	var paramTypes []llvm.Type
	for i, param := range f.Params {
		switch typ := param.Type().(type) {
		case *types.Basic:
			var paramType llvm.Type
			switch typ.Kind() {
			case types.Int:
				paramType = c.intType
			case types.Int32:
				paramType = llvm.Int32Type()
			default:
				return errors.New("todo: unknown basic param type")
			}
			paramTypes = append(paramTypes, paramType)
			c.funcs[f].params[param] = i
		default:
			return errors.New("todo: unknown param type")
		}
	}

	fnType := llvm.FunctionType(retType, paramTypes, false)
	llvm.AddFunction(c.mod, name, fnType)
	return nil
}

func (c *Compiler) parseFunc(pkgName string, f *ssa.Function) error {
	fmt.Println("func:", f.Name())

	fn := c.funcs[f]
	llvmFn := c.mod.NamedFunction(fn.name)
	start := c.ctx.AddBasicBlock(llvmFn, "start")
	c.builder.SetInsertPointAtEnd(start)

	// TODO: external functions
	for _, block := range f.Blocks {
		for _, instr := range block.Instrs {
			fmt.Printf("  instr: %v\n", instr)
			err := c.parseInstr(pkgName, instr)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Compiler) parseInstr(pkgName string, instr ssa.Instruction) error {
	switch instr := instr.(type) {
	case *ssa.Call:
		_, err := c.parseCall(pkgName, instr)
		return err
	case *ssa.Return:
		if len(instr.Results) == 0 {
			c.builder.CreateRetVoid()
			return nil
		} else if len(instr.Results) == 1 {
			val, err := c.parseExpr(pkgName, instr.Results[0])
			if err != nil {
				return err
			}
			c.builder.CreateRet(*val)
			return nil
		} else {
			return errors.New("todo: return value")
		}
	case *ssa.BinOp:
		_, err := c.parseBinOp(pkgName, instr)
		return err
	default:
		return errors.New("unknown instruction: " + fmt.Sprintf("%#v", instr))
	}
}

func (c *Compiler) parseBuiltin(pkgName string, instr *ssa.CallCommon, call *ssa.Builtin) (*llvm.Value, error) {
	fmt.Printf("    builtin: %#v\n", call)
	name := call.Name()

	switch name {
	case "print", "println":
		for i, arg := range instr.Args {
			if i >= 1 {
				c.builder.CreateCall(c.printspaceFunc, nil, "")
			}
			fmt.Printf("    arg: %s\n", arg);
			expr, err := c.parseExpr(pkgName, arg)
			if err != nil {
				return nil, err
			}
			switch expr.Type() {
			case c.stringPtrType:
				c.builder.CreateCall(c.printstringFunc, []llvm.Value{*expr}, "")
			case c.intType:
				c.builder.CreateCall(c.printintFunc, []llvm.Value{*expr}, "")
			default:
				return nil, errors.New("unknown arg type")
			}
		}
		if name == "println" {
			c.builder.CreateCall(c.printnlFunc, nil, "")
		}
		return nil, nil // print() or println() returns void
	default:
		return nil, errors.New("todo: builtin: " + name)
	}
}

func (c *Compiler) parseFunctionCall(pkgName string, call *ssa.CallCommon, fn *ssa.Function) (*llvm.Value, error) {
	fmt.Printf("    function: %s\n", fn)

	name := fn.Name()
	if strings.IndexByte(name, '.') == -1 {
		// TODO: import path instead of pkgName
		name = pkgName + "." + name
	}
	target := c.mod.NamedFunction(name)
	if target.IsNil() {
		return nil, errors.New("undefined function: " + name)
	}

	var params []llvm.Value
	for _, param := range call.Args {
		val, err := c.parseExpr(pkgName, param)
		if err != nil {
			return nil, err
		}
		params = append(params, *val)
	}

	val := c.builder.CreateCall(target, params, "")
	return &val, nil
}

func (c *Compiler) parseCall(pkgName string, instr *ssa.Call) (*llvm.Value, error) {
	fmt.Printf("    call: %s\n", instr)

	switch call := instr.Common().Value.(type) {
	case *ssa.Builtin:
		return c.parseBuiltin(pkgName, instr.Common(), call)
	case *ssa.Function:
		return c.parseFunctionCall(pkgName, instr.Common(), call)
	default:
		return nil, errors.New("todo: unknown call type: " + fmt.Sprintf("%#v", call))
	}
}

func (c *Compiler) parseBinOp(pkgName string, binop *ssa.BinOp) (*llvm.Value, error) {
	x, err := c.parseExpr(pkgName, binop.X)
	if err != nil {
		return nil, err
	}
	y, err := c.parseExpr(pkgName, binop.Y)
	if err != nil {
		return nil, err
	}
	switch binop.Op {
	case token.ADD:
		val := c.builder.CreateBinOp(llvm.Add, *x, *y, "")
		return &val, nil
	case token.MUL:
		val := c.builder.CreateBinOp(llvm.Mul, *x, *y, "")
		return &val, nil
	default:
		return nil, errors.New("todo: unknown binop")
	}
}

func (c *Compiler) parseExpr(pkgName string, expr ssa.Value) (*llvm.Value, error) {
	fmt.Printf("      expr: %v\n", expr)
	switch expr := expr.(type) {
	case *ssa.Const:
		switch expr.Value.Kind() {
		case constant.String:
			str := constant.StringVal(expr.Value)
			strVal := c.ctx.ConstString(str, false)
			strLen := llvm.ConstInt(c.stringLenType, uint64(len(str)), false)
			strObj := llvm.ConstStruct([]llvm.Value{strLen, strVal}, false)
			ptr := llvm.AddGlobal(c.mod, strObj.Type(), ".str")
			ptr.SetInitializer(strObj)
			ptr.SetLinkage(llvm.InternalLinkage)
			ptrCast := llvm.ConstPointerCast(ptr, c.stringPtrType)
			return &ptrCast, nil
		case constant.Int:
			n, _ := constant.Int64Val(expr.Value) // TODO: do something with the 'exact' return value?
			val := llvm.ConstInt(c.intType, uint64(n), true)
			return &val, nil
		default:
			return nil, errors.New("todo: unknown constant")
		}
	case *ssa.BinOp:
		return c.parseBinOp(pkgName, expr)
	case *ssa.Call:
		return c.parseCall(pkgName, expr)
	case *ssa.Parameter:
		fn := c.funcs[expr.Parent()]
		llvmFn := c.mod.NamedFunction(fn.name)
		param := llvmFn.Param(fn.params[expr])
		return &param, nil
	default:
		return nil, errors.New("todo: unknown expression: " + fmt.Sprintf("%#v", expr))
	}
}

// IR returns the whole IR as a human-readable string.
func (c *Compiler) IR() string {
	return c.mod.String()
}

func (c *Compiler) Verify() error {
	return llvm.VerifyModule(c.mod, llvm.PrintMessageAction)
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
func Compile(inpath, outpath, target string, printIR bool) error {
	c, err := NewCompiler(inpath, target)
	if err != nil {
		return err
	}

	err = c.Parse(inpath)
	if err != nil {
		return err
	}

	if err := c.Verify(); err != nil {
		return err
	}
	c.Optimize(2)
	if err := c.Verify(); err != nil {
		return err
	}

	if printIR {
		fmt.Println(c.IR())
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
