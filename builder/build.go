// Package builder is the compiler driver of TinyGo. It takes in a package name
// and an output path, and outputs an executable. It manages the entire
// compilation pipeline in between.
package builder

import (
	"debug/elf"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/compiler"
	"github.com/tinygo-org/tinygo/goenv"
	"github.com/tinygo-org/tinygo/interp"
	"github.com/tinygo-org/tinygo/stacksize"
	"github.com/tinygo-org/tinygo/transform"
	"tinygo.org/x/go-llvm"
)

// Build performs a single package to executable Go build. It takes in a package
// name, an output path, and set of compile options and from that it manages the
// whole compilation process.
//
// The error value may be of type *MultiError. Callers will likely want to check
// for this case and print such errors individually.
func Build(pkgName, outpath string, config *compileopts.Config, action func(string) error) error {
	// Compile Go code to IR.
	machine, err := compiler.NewTargetMachine(config)
	if err != nil {
		return err
	}
	mod, extraFiles, extraLDFlags, errs := compiler.Compile(pkgName, machine, config)
	if errs != nil {
		return newMultiError(errs)
	}

	if config.Options.PrintIR {
		fmt.Println("; Generated LLVM IR:")
		fmt.Println(mod.String())
	}
	if err := llvm.VerifyModule(mod, llvm.PrintMessageAction); err != nil {
		return errors.New("verification error after IR construction")
	}

	err = interp.Run(mod, config.DumpSSA())
	if err != nil {
		return err
	}
	if err := llvm.VerifyModule(mod, llvm.PrintMessageAction); err != nil {
		return errors.New("verification error after interpreting runtime.initAll")
	}

	if config.GOOS() != "darwin" {
		transform.ApplyFunctionSections(mod) // -ffunction-sections
	}

	// Browsers cannot handle external functions that have type i64 because it
	// cannot be represented exactly in JavaScript (JS only has doubles). To
	// keep functions interoperable, pass int64 types as pointers to
	// stack-allocated values.
	// Use -wasm-abi=generic to disable this behaviour.
	if config.Options.WasmAbi == "js" && strings.HasPrefix(config.Triple(), "wasm") {
		err := transform.ExternalInt64AsPtr(mod)
		if err != nil {
			return err
		}
	}

	// Optimization levels here are roughly the same as Clang, but probably not
	// exactly.
	errs = nil
	switch config.Options.Opt {
	/*
		Currently, turning optimizations off causes compile failures.
		We rely on the optimizer removing some dead symbols.
		Avoid providing an option that does not work right now.
		In the future once everything has been fixed we can re-enable this.

		case "none", "0":
			errs = transform.Optimize(mod, config, 0, 0, 0) // -O0
	*/
	case "1":
		errs = transform.Optimize(mod, config, 1, 0, 0) // -O1
	case "2":
		errs = transform.Optimize(mod, config, 2, 0, 225) // -O2
	case "s":
		errs = transform.Optimize(mod, config, 2, 1, 225) // -Os
	case "z":
		errs = transform.Optimize(mod, config, 2, 2, 5) // -Oz, default
	default:
		errs = []error{errors.New("unknown optimization level: -opt=" + config.Options.Opt)}
	}
	if len(errs) > 0 {
		return newMultiError(errs)
	}
	if err := llvm.VerifyModule(mod, llvm.PrintMessageAction); err != nil {
		return errors.New("verification failure after LLVM optimization passes")
	}

	// On the AVR, pointers can point either to flash or to RAM, but we don't
	// know. As a temporary fix, load all global variables in RAM.
	// In the future, there should be a compiler pass that determines which
	// pointers are flash and which are in RAM so that pointers can have a
	// correct address space parameter (address space 1 is for flash).
	if strings.HasPrefix(config.Triple(), "avr") {
		transform.NonConstGlobals(mod)
		if err := llvm.VerifyModule(mod, llvm.PrintMessageAction); err != nil {
			return errors.New("verification error after making all globals non-constant on AVR")
		}
	}

	// Generate output.
	outext := filepath.Ext(outpath)
	switch outext {
	case ".o":
		llvmBuf, err := machine.EmitToMemoryBuffer(mod, llvm.ObjectFile)
		if err != nil {
			return err
		}
		return ioutil.WriteFile(outpath, llvmBuf.Bytes(), 0666)
	case ".bc":
		data := llvm.WriteBitcodeToMemoryBuffer(mod).Bytes()
		return ioutil.WriteFile(outpath, data, 0666)
	case ".ll":
		data := []byte(mod.String())
		return ioutil.WriteFile(outpath, data, 0666)
	default:
		// Act as a compiler driver.

		// Create a temporary directory for intermediary files.
		dir, err := ioutil.TempDir("", "tinygo")
		if err != nil {
			return err
		}
		defer os.RemoveAll(dir)

		// Write the object file.
		objfile := filepath.Join(dir, "main.o")
		llvmBuf, err := machine.EmitToMemoryBuffer(mod, llvm.ObjectFile)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(objfile, llvmBuf.Bytes(), 0666)
		if err != nil {
			return err
		}

		// Prepare link command.
		executable := filepath.Join(dir, "main")
		tmppath := executable // final file
		ldflags := append(config.LDFlags(), "-o", executable, objfile)

		// Load builtins library from the cache, possibly compiling it on the
		// fly.
		if config.Target.RTLib == "compiler-rt" {
			librt, err := CompilerRT.Load(config.Triple())
			if err != nil {
				return err
			}
			ldflags = append(ldflags, librt)
		}

		// Add libc.
		if config.Target.Libc == "picolibc" {
			libc, err := Picolibc.Load(config.Triple())
			if err != nil {
				return err
			}
			ldflags = append(ldflags, libc)
		}

		// Compile extra files.
		root := goenv.Get("TINYGOROOT")
		for i, path := range config.ExtraFiles() {
			abspath := filepath.Join(root, path)
			outpath := filepath.Join(dir, "extra-"+strconv.Itoa(i)+"-"+filepath.Base(path)+".o")
			err := runCCompiler(config.Target.Compiler, append(config.CFlags(), "-c", "-o", outpath, abspath)...)
			if err != nil {
				return &commandError{"failed to build", path, err}
			}
			ldflags = append(ldflags, outpath)
		}

		// Compile C files in packages.
		for i, file := range extraFiles {
			outpath := filepath.Join(dir, "pkg"+strconv.Itoa(i)+"-"+filepath.Base(file)+".o")
			err := runCCompiler(config.Target.Compiler, append(config.CFlags(), "-c", "-o", outpath, file)...)
			if err != nil {
				return &commandError{"failed to build", file, err}
			}
			ldflags = append(ldflags, outpath)
		}

		if len(extraLDFlags) > 0 {
			ldflags = append(ldflags, extraLDFlags...)
		}

		// Link the object files together.
		err = link(config.Target.Linker, ldflags...)
		if err != nil {
			return &commandError{"failed to link", executable, err}
		}

		if config.Options.PrintSizes == "short" || config.Options.PrintSizes == "full" {
			sizes, err := loadProgramSize(executable)
			if err != nil {
				return err
			}
			if config.Options.PrintSizes == "short" {
				fmt.Printf("   code    data     bss |   flash     ram\n")
				fmt.Printf("%7d %7d %7d | %7d %7d\n", sizes.Code, sizes.Data, sizes.BSS, sizes.Code+sizes.Data, sizes.Data+sizes.BSS)
			} else {
				fmt.Printf("   code  rodata    data     bss |   flash     ram | package\n")
				for _, name := range sizes.sortedPackageNames() {
					pkgSize := sizes.Packages[name]
					fmt.Printf("%7d %7d %7d %7d | %7d %7d | %s\n", pkgSize.Code, pkgSize.ROData, pkgSize.Data, pkgSize.BSS, pkgSize.Flash(), pkgSize.RAM(), name)
				}
				fmt.Printf("%7d %7d %7d %7d | %7d %7d | (sum)\n", sizes.Sum.Code, sizes.Sum.ROData, sizes.Sum.Data, sizes.Sum.BSS, sizes.Sum.Flash(), sizes.Sum.RAM())
				fmt.Printf("%7d       - %7d %7d | %7d %7d | (all)\n", sizes.Code, sizes.Data, sizes.BSS, sizes.Code+sizes.Data, sizes.Data+sizes.BSS)
			}
		}

		// Print goroutine stack sizes, as far as possible.
		if config.Options.PrintStacks {
			printStacks(mod, executable)
		}

		// Get an Intel .hex file or .bin file from the .elf file.
		if outext == ".hex" || outext == ".bin" || outext == ".gba" {
			tmppath = filepath.Join(dir, "main"+outext)
			err := objcopy(executable, tmppath)
			if err != nil {
				return err
			}
		} else if outext == ".uf2" {
			// Get UF2 from the .elf file.
			tmppath = filepath.Join(dir, "main"+outext)
			err := convertELFFileToUF2File(executable, tmppath, config.Target.UF2FamilyID)
			if err != nil {
				return err
			}
		}
		return action(tmppath)
	}
}

// printStacks prints the maximum stack depth for functions that are started as
// goroutines. Stack sizes cannot always be determined statically, in particular
// recursive functions and functions that call interface methods or function
// pointers may have an unknown stack depth (depending on what the optimizer
// manages to optimize away).
//
// It might print something like the following:
//
//     function                         stack usage (in bytes)
//     Reset_Handler                    316
//     examples/blinky2.led1            92
//     runtime.run$1                    300
func printStacks(mod llvm.Module, executable string) {
	var callsIndirectFunction []string
	gowrappers := []string{}
	gowrapperNames := make(map[string]string)
	for fn := mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
		// Determine which functions call a function pointer.
		for bb := fn.FirstBasicBlock(); !bb.IsNil(); bb = llvm.NextBasicBlock(bb) {
			for inst := bb.FirstInstruction(); !inst.IsNil(); inst = llvm.NextInstruction(inst) {
				if inst.IsACallInst().IsNil() {
					continue
				}
				if callee := inst.CalledValue(); callee.IsAFunction().IsNil() && callee.IsAInlineAsm().IsNil() {
					callsIndirectFunction = append(callsIndirectFunction, fn.Name())
				}
			}
		}

		// Get a list of "go wrappers", small wrapper functions that decode
		// parameters when starting a new goroutine.
		attr := fn.GetStringAttributeAtIndex(-1, "tinygo-gowrapper")
		if !attr.IsNil() {
			gowrappers = append(gowrappers, fn.Name())
			gowrapperNames[fn.Name()] = attr.GetStringValue()
		}
	}
	sort.Strings(gowrappers)

	// Load the ELF binary.
	f, err := elf.Open(executable)
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not load executable for stack size analysis:", err)
		return
	}
	defer f.Close()

	// Determine the frame size of each function (if available) and the callgraph.
	functions, err := stacksize.CallGraph(f, callsIndirectFunction)
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not parse executable for stack size analysis:", err)
		return
	}

	switch f.Machine {
	case elf.EM_ARM:
		// Add the reset handler, which runs startup code and is the
		// interrupt/scheduler stack with -scheduler=tasks.
		// Note that because interrupts happen on this stack, the stack needed
		// by just the Reset_Handler is not enough. Stacks needed by interrupt
		// handlers should also be taken into account.
		gowrappers = append([]string{"Reset_Handler"}, gowrappers...)
		gowrapperNames["Reset_Handler"] = "Reset_Handler"
	}

	// Print the sizes of all stacks.
	fmt.Printf("%-32s %s\n", "function", "stack usage (in bytes)")
	for _, name := range gowrappers {
		for _, fn := range functions[name] {
			stackSize, stackSizeType, missingStackSize := fn.StackSize()
			funcName := gowrapperNames[name]
			if funcName == "" {
				funcName = "<unknown>"
			}
			switch stackSizeType {
			case stacksize.Bounded:
				fmt.Printf("%-32s %d\n", funcName, stackSize)
			case stacksize.Unknown:
				fmt.Printf("%-32s unknown, %s does not have stack frame information\n", funcName, missingStackSize)
			case stacksize.Recursive:
				fmt.Printf("%-32s recursive, %s may call itself\n", funcName, missingStackSize)
			case stacksize.IndirectCall:
				fmt.Printf("%-32s unknown, %s calls a function pointer\n", funcName, missingStackSize)
			}
		}
	}
}
