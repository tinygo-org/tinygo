
package main

import (
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"os"

	"golang.org/x/tools/go/loader"
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
	printstringFunc llvm.Value
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

	printstringType := llvm.FunctionType(llvm.VoidType(), []llvm.Type{llvm.PointerType(llvm.Int8Type(), 0)}, false)
	c.printstringFunc = llvm.AddFunction(c.mod, "__go_printstring", printstringType)
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
	program, err := config.Load()
	if err != nil {
		return err
	}

	pkgInfo := program.Created[0] // main?
	if len(pkgInfo.Errors) != 0 {
		return pkgInfo.Errors[0] // TODO: better error checking
	}
	fmt.Println("package:", pkgInfo.Pkg.Name())
	for _, file := range pkgInfo.Files {
		err := c.parseFile(pkgInfo.Pkg.Name(), file)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Compiler) parseFile(pkgName string, file *ast.File) error {
	for _, decl := range file.Decls {
		switch v := decl.(type) {
		case *ast.FuncDecl:
			err := c.parseFunc(pkgName, v)
			if err != nil {
				return err
			}
		default:
			return errors.New("unknown declaration")
		}
	}

	return nil
}

func (c *Compiler) parseFunc(pkgName string, f *ast.FuncDecl) error {
	fmt.Println("func:", f.Name)

	var fnType llvm.Type
	if f.Type.Results == nil {
		fnType = llvm.FunctionType(llvm.VoidType(), nil, false)
	} else {
		return errors.New("todo: return values")
	}

	fn := llvm.AddFunction(c.mod, pkgName + "." + f.Name.Name, fnType)
	start := c.ctx.AddBasicBlock(fn, "start")
	c.builder.SetInsertPointAtEnd(start)

	// TODO: external functions
	for _, stmt := range f.Body.List {
		err := c.parseStmt(stmt)
		if err != nil {
			return err
		}
	}

	if f.Type.Results == nil {
		c.builder.CreateRetVoid()
	//} else if len(f.Type.Results.List) == 1 {
	//	c.builder.CreateRet(llvm.ConstInt(llvm.Int32Type(), 0, false))
	}
	return nil
}

func (c *Compiler) parseStmt(stmt ast.Node) error {
	switch v := stmt.(type) {
	case *ast.ExprStmt:
		err := c.parseExpr(v.X)
		if err != nil {
			return err
		}
	default:
		return errors.New("unknown stmt")
	}
	return nil
}

func (c *Compiler) parseExpr(expr ast.Expr) error {
	switch v := expr.(type) {
	case *ast.CallExpr:
		zero := llvm.ConstInt(llvm.Int32Type(), 0, false)
		name := v.Fun.(*ast.Ident).Name
		fmt.Printf("  call: %s\n", name)

		printnl := false
		if name == "println" {
			printnl = true
		} else if name == "print" {
		} else {
			return errors.New("todo: call anything other than println()")
		}

		for i, arg := range v.Args {
			if i >= 1 {
				c.builder.CreateCall(c.printspaceFunc, nil, "")
			}
			switch arg := arg.(type) {
			case *ast.BasicLit:
				fmt.Printf("    arg: %s\n", arg.Value)
				val := constant.MakeFromLiteral(arg.Value, arg.Kind, 0)
				switch arg.Kind {
				case token.STRING:
					str := constant.StringVal(val)
					strObj := c.builder.CreateGlobalString(str + "\x00", "")
					msgCast := c.builder.CreateGEP(strObj, []llvm.Value{zero, zero}, "")
					c.builder.CreateCall(c.printstringFunc, []llvm.Value{msgCast}, "")
				default:
					return errors.New("todo: print anything other than strings")
				}
			default:
				return errors.New("unknown arg type")
			}
		}
		if printnl {
			c.builder.CreateCall(c.printnlFunc, nil, "")
		}
	default:
		return errors.New("unknown expr")
	}
	return nil
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
