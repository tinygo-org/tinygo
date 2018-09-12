package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/aykevl/llvm/bindings/go/llvm"
)

// Helper function for Compiler object.
func Compile(pkgName, runtimePath, outpath, target string, printIR, dumpSSA bool) error {
	var buildTags []string
	// TODO: put this somewhere else
	if target == "pca10040" {
		// Pretend to be a WASM target, not ARM (for standard library support).
		buildTags = append(buildTags, "nrf", "nrf52", "nrf52832", "js", "wasm")
		target = "armv7m-none-eabi"
	} else if target == "arduino" {
		// Pretend to be a WASM target, not AVR (for standard library support).
		buildTags = append(buildTags, "avr", "avr8", "atmega", "atmega328p", "js", "wasm")
		target = "avr--"
	} else {
		buildTags = append(buildTags, "linux", "amd64")
	}

	c, err := NewCompiler(pkgName, target, dumpSSA)
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
		return c.Parse(pkgName, buildTags)
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

	err = c.EmitObject(outpath)
	if err != nil {
		return err
	}

	return nil
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s command [-printir] -runtime=<runtime.bc> [-target=<target>] -o <output> <input>\n", os.Args[0])
	fmt.Fprintln(os.Stderr, "\ncommands:")
	fmt.Fprintln(os.Stderr, "  build: compile packages and dependencies")
	fmt.Fprintln(os.Stderr, "  help:  print this help text")
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
	default:
		fmt.Fprintln(os.Stderr, "Unknown command:", command)
		usage()
		os.Exit(1)
	}
}
