package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/aykevl/llvm/bindings/go/llvm"
)

// Helper function for Compiler object.
func Compile(pkgName, runtimePath, outpath, target string, printIR, dumpSSA bool) error {
	spec, err := LoadTarget(target)

	c, err := NewCompiler(pkgName, spec.Triple, dumpSSA)
	if err != nil {
		return err
	}

	// Add C/LLVM runtime.
	if runtimePath != "" {
		runtime, err := llvm.ParseBitcodeFile(runtimePath)
		if err != nil {
			return err
		}
		err = c.LinkModule(runtime)
		if err != nil {
			return err
		}
	}

	// Compile Go code to IR.
	parseErr := func() error {
		if printIR {
			// Run this even if c.Parse() panics.
			defer func() {
				fmt.Println("Generated LLVM IR:")
				fmt.Println(c.IR())
			}()
		}
		return c.Parse(pkgName, spec.BuildTags)
	}()
	if parseErr != nil {
		return parseErr
	}

	c.ApplyFunctionSections() // -ffunction-sections

	if err := c.Verify(); err != nil {
		return err
	}

	// TODO: provide a flag to disable (most) optimizations.
	c.Optimize(2, 2, 5) // -Oz params
	if err := c.Verify(); err != nil {
		return err
	}

	// Generate output.
	if strings.HasSuffix(outpath, ".o") {
		return c.EmitObject(outpath)
	} else if strings.HasSuffix(outpath, ".bc") {
		return c.EmitBitcode(outpath)
	} else if strings.HasSuffix(outpath, ".ll") {
		return c.EmitText(outpath)
	} else {
		return errors.New("unknown output file extension")
	}
}

// Run the specified package directly (using JIT or interpretation).
func Run(pkgName string) error {
	c, err := NewCompiler(pkgName, llvm.DefaultTargetTriple(), false)
	if err != nil {
		return errors.New("compiler: " + err.Error())
	}
	err = c.Parse(pkgName, []string{runtime.GOOS, runtime.GOARCH})
	if err != nil {
		return errors.New("compiler: " + err.Error())
	}
	if err := c.Verify(); err != nil {
		return errors.New("compiler error: failed to verify module: " + err.Error())
	}
	c.Optimize(1, 0, 0) // -O1, the fastest optimization level that doesn't crash

	engine, err := llvm.NewExecutionEngine(c.mod)
	if err != nil {
		return errors.New("interpreter setup: " + err.Error())
	}
	defer engine.Dispose()

	main := engine.FindFunction("main")
	if main.IsNil() {
		return errors.New("could not find main function")
	}
	engine.RunFunction(main, nil)

	return nil
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s command [-printir] -runtime=<runtime.bc> [-target=<target>] -o <output> <input>\n", os.Args[0])
	fmt.Fprintln(os.Stderr, "\ncommands:")
	fmt.Fprintln(os.Stderr, "  build: compile packages and dependencies")
	fmt.Fprintln(os.Stderr, "  help:  print this help text")
	fmt.Fprintln(os.Stderr, "  run:   run package in an interpreter")
	fmt.Fprintln(os.Stderr, "\nflags:")
	flag.PrintDefaults()
}

func main() {
	outpath := flag.String("o", "", "output filename")
	printIR := flag.Bool("printir", false, "print LLVM IR")
	dumpSSA := flag.Bool("dumpssa", false, "dump internal Go SSA")
	runtime := flag.String("runtime", "", "runtime LLVM bitcode files (from C sources)")
	target := flag.String("target", llvm.DefaultTargetTriple(), "LLVM target")

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "No command-line arguments supplied.")
		usage()
		os.Exit(1)
	}
	command := os.Args[1]

	flag.CommandLine.Parse(os.Args[2:])

	os.Setenv("CC", "clang -target="+*target)

	switch command {
	case "build":
		if *outpath == "" {
			fmt.Fprintln(os.Stderr, "No output filename supplied (-o).")
			usage()
			os.Exit(1)
		}
		if flag.NArg() != 1 {
			fmt.Fprintln(os.Stderr, "No package specified.")
			usage()
			os.Exit(1)
		}
		err := Compile(flag.Arg(0), *runtime, *outpath, *target, *printIR, *dumpSSA)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	case "help":
		usage()
	case "run":
		if flag.NArg() != 1 {
			fmt.Fprintln(os.Stderr, "No package specified.")
			usage()
			os.Exit(1)
		}
		if *target != llvm.DefaultTargetTriple() {
			fmt.Fprintf(os.Stderr, "Cannot run %s: target triple does not match host triple.")
			os.Exit(1)
		}
		err := Run(flag.Arg(0))
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintln(os.Stderr, "Unknown command:", command)
		usage()
		os.Exit(1)
	}
}
