
package main

import (
	"errors"
	"flag"
	"fmt"
	"go/constant"
	"go/token"
	"os"

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
	stringType      llvm.Type
	stringPtrType   llvm.Type
	printstringFunc llvm.Value
	printintFunc    llvm.Value
	printspaceFunc  llvm.Value
	printnlFunc     llvm.Value
}

func NewCompiler(path, triple string) (*Compiler, error) {
	c := &Compiler{}

	target, err := llvm.GetTargetFromTriple(triple)
	if err != nil {
		return nil, err
	}
	c.machine = target.CreateTargetMachine(triple, "", "", llvm.CodeGenLevelDefault, llvm.RelocDefault, llvm.CodeModelDefault)

	c.mod = llvm.NewModule(path)
	c.ctx = c.mod.Context()
	c.builder = c.ctx.NewBuilder()

	// Length-prefixed string.
	c.stringType = llvm.StructType([]llvm.Type{llvm.Int32Type(), llvm.ArrayType(llvm.Int8Type(), 0)}, false)
	c.stringPtrType = llvm.PointerType(c.stringType, 0)

	printstringType := llvm.FunctionType(llvm.VoidType(), []llvm.Type{c.stringPtrType}, false)
	c.printstringFunc = llvm.AddFunction(c.mod, "__go_printstring", printstringType)
	printintType := llvm.FunctionType(llvm.VoidType(), []llvm.Type{llvm.Int32Type()}, false)
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
		for name, member := range pkg.Members {
			fmt.Println("member:", name, member, member.Token())
			if member.Name() == "init" {
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

func (c *Compiler) parseFunc(pkgName string, f *ssa.Function) error {
	fmt.Println("func:", f.Name(), f.Blocks, "len:", len(f.Blocks))

	var fnType llvm.Type
	if f.Signature.Results() == nil {
		fnType = llvm.FunctionType(llvm.VoidType(), nil, false)
	} else {
		return errors.New("todo: return values")
	}

	fn := llvm.AddFunction(c.mod, pkgName + "." + f.Name(), fnType)
	start := c.ctx.AddBasicBlock(fn, "start")
	c.builder.SetInsertPointAtEnd(start)

	// TODO: external functions
	for _, block := range f.Blocks {
		for _, instr := range block.Instrs {
			fmt.Printf("  instr: %v\n", instr)
			err := c.parseInstr(instr)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Compiler) parseInstr(instr ssa.Instruction) error {
	switch instr := instr.(type) {
	case *ssa.Call:
		switch call := instr.Common().Value.(type) {
		case *ssa.Builtin:
			return c.parseBuiltin(instr.Common(), call)
		default:
			return errors.New("todo: unknown call type: " + fmt.Sprintf("%#v", call))
		}
	case *ssa.Return:
		if len(instr.Results) == 0 {
			c.builder.CreateRetVoid()
			return nil
		} else {
			return errors.New("todo: return value")
		}
	case *ssa.BinOp:
		return c.parseBinOp(instr)
	default:
		return errors.New("unknown instruction: " + fmt.Sprintf("%#v", instr))
	}
}

func (c *Compiler) parseBuiltin(instr *ssa.CallCommon, call *ssa.Builtin) error {
	fmt.Printf("    builtin: %#v\n", call)
	name := call.Name()

	switch name {
	case "print", "println":
		for i, arg := range instr.Args {
			if i >= 1 {
				c.builder.CreateCall(c.printspaceFunc, nil, "")
			}
			fmt.Printf("    arg: %s\n", arg);
			expr, err := c.parseExpr(arg)
			if err != nil {
				return err
			}
			switch expr.Type() {
			case c.stringPtrType:
				c.builder.CreateCall(c.printstringFunc, []llvm.Value{*expr}, "")
			case llvm.Int32Type():
				c.builder.CreateCall(c.printintFunc, []llvm.Value{*expr}, "")
			default:
				return errors.New("unknown arg type")
			}
		}
		if name == "println" {
			c.builder.CreateCall(c.printnlFunc, nil, "")
		}
	}

	return nil
}

func (c *Compiler) parseBinOp(binop *ssa.BinOp) error {
	x, err := c.parseExpr(binop.X)
	if err != nil {
		return err
	}
	y, err := c.parseExpr(binop.Y)
	if err != nil {
		return err
	}
	switch binop.Op {
	case token.ADD:
		c.builder.CreateBinOp(llvm.Add, *x, *y, "")
		return nil
	case token.MUL:
		c.builder.CreateBinOp(llvm.Mul, *x, *y, "")
		return nil
	}
	return errors.New("todo: unknown binop")
}

func (c *Compiler) parseExpr(expr ssa.Value) (*llvm.Value, error) {
	fmt.Printf("      expr: %v\n", expr)
	switch expr := expr.(type) {
	case *ssa.Const:
		switch expr.Value.Kind() {
		case constant.String:
			str := constant.StringVal(expr.Value)
			strVal := c.ctx.ConstString(str, false)
			strLen := llvm.ConstInt(llvm.Int32Type(), uint64(len(str)), false)
			strObj := llvm.ConstStruct([]llvm.Value{strLen, strVal}, false)
			ptr := llvm.AddGlobal(c.mod, strObj.Type(), ".str")
			ptr.SetInitializer(strObj)
			ptr.SetLinkage(llvm.InternalLinkage)
			ptrCast := llvm.ConstPointerCast(ptr, c.stringPtrType)
			return &ptrCast, nil
		case constant.Int:
			n, _ := constant.Int64Val(expr.Value) // TODO: do something with the 'exact' return value?
			val := llvm.ConstInt(llvm.Int32Type(), uint64(n), true)
			return &val, nil
		default:
			return nil, errors.New("todo: unknown constant")
		}
	}
	return nil, errors.New("todo: unknown expression: " + fmt.Sprintf("%#v", expr))
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
