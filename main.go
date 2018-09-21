package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/aykevl/llvm/bindings/go/llvm"
)

// Helper function for Compiler object.
func Compile(pkgName, outpath string, spec *TargetSpec, printIR, dumpSSA bool, printSizes string, action func(string) error) error {
	c, err := NewCompiler(pkgName, spec.Triple, dumpSSA)
	if err != nil {
		return err
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

	// On the AVR, pointers can point either to flash or to RAM, but we don't
	// know. As a temporary fix, load all global variables in RAM.
	// In the future, there should be a compiler pass that determines which
	// pointers are flash and which are in RAM so that pointers can have a
	// correct address space parameter (address space 1 is for flash).
	if strings.HasPrefix(spec.Triple, "avr") {
		c.NonConstGlobals()
		if err := c.Verify(); err != nil {
			return err
		}
	}

	// Generate output.
	if strings.HasSuffix(outpath, ".o") {
		return c.EmitObject(outpath)
	} else if strings.HasSuffix(outpath, ".bc") {
		return c.EmitBitcode(outpath)
	} else if strings.HasSuffix(outpath, ".ll") {
		return c.EmitText(outpath)
	} else {
		// Act as a compiler driver.

		// Create a temporary directory for intermediary files.
		dir, err := ioutil.TempDir("", "tinygo")
		if err != nil {
			return err
		}
		defer os.RemoveAll(dir)

		// Write the object file.
		objfile := filepath.Join(dir, "main.o")
		err = c.EmitObject(objfile)
		if err != nil {
			return err
		}

		// Link the object file with the system compiler.
		executable := filepath.Join(dir, "main")
		tmppath := executable // final file
		args := append(spec.PreLinkArgs, "-o", executable, objfile)
		cmd := exec.Command(spec.Linker, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			return err
		}

		if printSizes == "short" || printSizes == "full" {
			sizes, err := Sizes(executable)
			if err != nil {
				return err
			}
			if printSizes == "short" {
				fmt.Printf("   code    data     bss |   flash     ram\n")
				fmt.Printf("%7d %7d %7d | %7d %7d\n", sizes.Code, sizes.Data, sizes.BSS, sizes.Code+sizes.Data, sizes.Data+sizes.BSS)
			} else {
				fmt.Printf("   code  rodata    data     bss |   flash     ram | package\n")
				for _, name := range sizes.SortedPackageNames() {
					pkgSize := sizes.Packages[name]
					fmt.Printf("%7d %7d %7d %7d | %7d %7d | %s\n", pkgSize.Code, pkgSize.ROData, pkgSize.Data, pkgSize.BSS, pkgSize.Flash(), pkgSize.RAM(), name)
				}
				fmt.Printf("%7d %7d %7d %7d | %7d %7d | (sum)\n", sizes.Sum.Code, sizes.Sum.ROData, sizes.Sum.Data, sizes.Sum.BSS, sizes.Sum.Flash(), sizes.Sum.RAM())
				fmt.Printf("%7d       - %7d %7d | %7d %7d | (all)\n", sizes.Code, sizes.Data, sizes.BSS, sizes.Code+sizes.Data, sizes.Data+sizes.BSS)
			}
		}

		if strings.HasSuffix(outpath, ".hex") {
			// Get an Intel .hex file from the .elf file.
			tmppath = filepath.Join(dir, "main.hex")
			cmd := exec.Command(spec.Objcopy, "-O", "ihex", executable, tmppath)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err = cmd.Run()
			if err != nil {
				return err
			}
		}
		return action(tmppath)
	}
}

func Build(pkgName, outpath, target string, printIR, dumpSSA bool, printSizes string) error {
	spec, err := LoadTarget(target)
	if err != nil {
		return err
	}

	return Compile(pkgName, outpath, spec, printIR, dumpSSA, printSizes, func(tmppath string) error {
		if err := os.Rename(tmppath, outpath); err != nil {
			// Moving failed. Do a file copy.
			inf, err := os.Open(tmppath)
			if err != nil {
				return err
			}
			defer inf.Close()
			outf, err := os.OpenFile(outpath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
			if err != nil {
				return err
			}

			// Copy data to output file.
			_, err = io.Copy(outf, inf)
			if err != nil {
				return err
			}

			// Check whether file writing was successful.
			return outf.Close()
		} else {
			// Move was successful.
			return nil
		}
	})
}

func Flash(pkgName, target, port string, printIR, dumpSSA bool, printSizes string) error {
	spec, err := LoadTarget(target)
	if err != nil {
		return err
	}

	return Compile(pkgName, ".hex", spec, printIR, dumpSSA, printSizes, func(tmppath string) error {
		// Create the command.
		flashCmd := spec.Flasher
		flashCmd = strings.Replace(flashCmd, "{hex}", tmppath, -1)
		flashCmd = strings.Replace(flashCmd, "{port}", port, -1)

		// Execute the command.
		cmd := exec.Command("/bin/sh", "-c", flashCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	})
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
	fmt.Fprintf(os.Stderr, "usage: %s command [-printir] [-target=<target>] -o <output> <input>\n", os.Args[0])
	fmt.Fprintln(os.Stderr, "\ncommands:")
	fmt.Fprintln(os.Stderr, "  build: compile packages and dependencies")
	fmt.Fprintln(os.Stderr, "  flash: compile and flash to the device")
	fmt.Fprintln(os.Stderr, "  help:  print this help text")
	fmt.Fprintln(os.Stderr, "  run:   run package in an interpreter")
	fmt.Fprintln(os.Stderr, "\nflags:")
	flag.PrintDefaults()
}

func main() {
	outpath := flag.String("o", "", "output filename")
	printIR := flag.Bool("printir", false, "print LLVM IR")
	dumpSSA := flag.Bool("dumpssa", false, "dump internal Go SSA")
	target := flag.String("target", llvm.DefaultTargetTriple(), "LLVM target")
	printSize := flag.String("size", "", "print sizes (none, short, full)")
	port := flag.String("port", "/dev/ttyACM0", "flash port")

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
		err := Build(flag.Arg(0), *outpath, *target, *printIR, *dumpSSA, *printSize)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	case "flash":
		if *outpath != "" {
			fmt.Fprintln(os.Stderr, "Output cannot be specified with the flash command.")
			usage()
			os.Exit(1)
		}
		err := Flash(flag.Arg(0), *target, *port, *printIR, *dumpSSA, *printSize)
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
