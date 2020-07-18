// Package builder is the compiler driver of TinyGo. It takes in a package name
// and an output path, and outputs an executable. It manages the entire
// compilation pipeline in between.
package builder

import (
	"debug/elf"
	"encoding/binary"
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

	// Make sure stack sizes are loaded from a separate section so they can be
	// modified after linking.
	var stackSizeLoads []string
	if config.AutomaticStackSize() {
		stackSizeLoads = transform.CreateStackSizeLoads(mod, config)
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

		var calculatedStacks []string
		var stackSizes map[string]functionStackSize
		if config.Options.PrintStacks || config.AutomaticStackSize() {
			// Try to determine stack sizes at compile time.
			// Don't do this by default as it usually doesn't work on
			// unsupported architectures.
			calculatedStacks, stackSizes, err = determineStackSizes(mod, executable)
			if err != nil {
				return err
			}
		}
		if config.AutomaticStackSize() {
			// Modify the .tinygo_stacksizes section that contains a stack size
			// for each goroutine.
			err = modifyStackSizes(executable, stackSizeLoads, stackSizes)
			if err != nil {
				return fmt.Errorf("could not modify stack sizes: %w", err)
			}
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
			printStacks(calculatedStacks, stackSizes)
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

// functionStackSizes keeps stack size information about a single function
// (usually a goroutine).
type functionStackSize struct {
	humanName        string
	stackSize        uint64
	stackSizeType    stacksize.SizeType
	missingStackSize *stacksize.CallNode
}

// determineStackSizes tries to determine the stack sizes of all started
// goroutines and of the reset vector. The LLVM module is necessary to find
// functions that call a function pointer.
func determineStackSizes(mod llvm.Module, executable string) ([]string, map[string]functionStackSize, error) {
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
		return nil, nil, fmt.Errorf("could not load executable for stack size analysis: %w", err)
	}
	defer f.Close()

	// Determine the frame size of each function (if available) and the callgraph.
	functions, err := stacksize.CallGraph(f, callsIndirectFunction)
	if err != nil {
		return nil, nil, fmt.Errorf("could not parse executable for stack size analysis: %w", err)
	}

	// Goroutines need to be started and finished and take up some stack space
	// that way. This can be measured by measuing the stack size of
	// tinygo_startTask.
	if numFuncs := len(functions["tinygo_startTask"]); numFuncs != 1 {
		return nil, nil, fmt.Errorf("expected exactly one definition of tinygo_startTask, got %d", numFuncs)
	}
	baseStackSize, baseStackSizeType, baseStackSizeFailedAt := functions["tinygo_startTask"][0].StackSize()

	sizes := make(map[string]functionStackSize)

	// Add the reset handler function, for convenience. The reset handler runs
	// startup code and the scheduler. The listed stack size is not the full
	// stack size: interrupts are not counted.
	var resetFunction string
	switch f.Machine {
	case elf.EM_ARM:
		// Note: all interrupts happen on this stack so the real size is bigger.
		resetFunction = "Reset_Handler"
	}
	if resetFunction != "" {
		funcs := functions[resetFunction]
		if len(funcs) != 1 {
			return nil, nil, fmt.Errorf("expected exactly one definition of %s in the callgraph, found %d", resetFunction, len(funcs))
		}
		stackSize, stackSizeType, missingStackSize := funcs[0].StackSize()
		sizes[resetFunction] = functionStackSize{
			stackSize:        stackSize,
			stackSizeType:    stackSizeType,
			missingStackSize: missingStackSize,
			humanName:        resetFunction,
		}
	}

	// Add all goroutine wrapper functions.
	for _, name := range gowrappers {
		funcs := functions[name]
		if len(funcs) != 1 {
			return nil, nil, fmt.Errorf("expected exactly one definition of %s in the callgraph, found %d", name, len(funcs))
		}
		humanName := gowrapperNames[name]
		if humanName == "" {
			humanName = name // fallback
		}
		stackSize, stackSizeType, missingStackSize := funcs[0].StackSize()
		if baseStackSizeType != stacksize.Bounded {
			// It was not possible to determine the stack size at compile time
			// because tinygo_startTask does not have a fixed stack size. This
			// can happen when using -opt=1.
			stackSizeType = baseStackSizeType
			missingStackSize = baseStackSizeFailedAt
		} else if stackSize < baseStackSize {
			// This goroutine has a very small stack, but still needs to fit all
			// registers to start and suspend the goroutine. Otherwise a stack
			// overflow will occur even before the goroutine is started.
			stackSize = baseStackSize
		}
		sizes[name] = functionStackSize{
			stackSize:        stackSize,
			stackSizeType:    stackSizeType,
			missingStackSize: missingStackSize,
			humanName:        humanName,
		}
	}

	if resetFunction != "" {
		return append([]string{resetFunction}, gowrappers...), sizes, nil
	}
	return gowrappers, sizes, nil
}

// modifyStackSizes modifies the .tinygo_stacksizes section with the updated
// stack size information. Before this modification, all stack sizes in the
// section assume the default stack size (which is relatively big).
func modifyStackSizes(executable string, stackSizeLoads []string, stackSizes map[string]functionStackSize) error {
	fp, err := os.OpenFile(executable, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer fp.Close()

	elfFile, err := elf.NewFile(fp)
	if err != nil {
		return err
	}

	section := elfFile.Section(".tinygo_stacksizes")
	if section == nil {
		return errors.New("could not find .tinygo_stacksizes section")
	}

	if section.Size != section.FileSize {
		// Sanity check.
		return fmt.Errorf("expected .tinygo_stacksizes to have identical size and file size, got %d and %d", section.Size, section.FileSize)
	}

	// Read all goroutine stack sizes.
	data := make([]byte, section.Size)
	_, err = fp.ReadAt(data, int64(section.Offset))
	if err != nil {
		return err
	}

	if len(stackSizeLoads)*4 != len(data) {
		// Note: while AVR should use 2 byte stack sizes, even 64-bit platforms
		// should probably stick to 4 byte stack sizes as a larger than 4GB
		// stack doesn't make much sense.
		return errors.New("expected 4 byte stack sizes")
	}

	// Modify goroutine stack sizes with a compile-time known worst case stack
	// size.
	for i, name := range stackSizeLoads {
		fn, ok := stackSizes[name]
		if !ok {
			return fmt.Errorf("could not find symbol %s in ELF file", name)
		}
		if fn.stackSizeType == stacksize.Bounded {
			// Note: adding 4 for the stack canary. Even though the size may be
			// automatically determined, stack overflow checking is still
			// important as the stack size cannot be determined for all
			// goroutines.
			binary.LittleEndian.PutUint32(data[i*4:], uint32(fn.stackSize)+4)
		}
	}

	// Write back the modified stack sizes.
	_, err = fp.WriteAt(data, int64(section.Offset))
	if err != nil {
		return err
	}

	return nil
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
func printStacks(calculatedStacks []string, stackSizes map[string]functionStackSize) {
	// Print the sizes of all stacks.
	fmt.Printf("%-32s %s\n", "function", "stack usage (in bytes)")
	for _, name := range calculatedStacks {
		fn := stackSizes[name]
		switch fn.stackSizeType {
		case stacksize.Bounded:
			fmt.Printf("%-32s %d\n", fn.humanName, fn.stackSize)
		case stacksize.Unknown:
			fmt.Printf("%-32s unknown, %s does not have stack frame information\n", fn.humanName, fn.missingStackSize)
		case stacksize.Recursive:
			fmt.Printf("%-32s recursive, %s may call itself\n", fn.humanName, fn.missingStackSize)
		case stacksize.IndirectCall:
			fmt.Printf("%-32s unknown, %s calls a function pointer\n", fn.humanName, fn.missingStackSize)
		}
	}
}
