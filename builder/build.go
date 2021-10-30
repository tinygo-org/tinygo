// Package builder is the compiler driver of TinyGo. It takes in a package name
// and an output path, and outputs an executable. It manages the entire
// compilation pipeline in between.
package builder

import (
	"crypto/sha512"
	"debug/elf"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"go/types"
	"hash/crc32"
	"io/ioutil"
	"math/bits"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/compiler"
	"github.com/tinygo-org/tinygo/goenv"
	"github.com/tinygo-org/tinygo/interp"
	"github.com/tinygo-org/tinygo/loader"
	"github.com/tinygo-org/tinygo/stacksize"
	"github.com/tinygo-org/tinygo/transform"
	"tinygo.org/x/go-llvm"
)

// BuildResult is the output of a build. This includes the binary itself and
// some other metadata that is obtained while building the binary.
type BuildResult struct {
	// A path to the output binary. It will be removed after Build returns, so
	// if it should be kept it must be copied or moved away.
	Binary string

	// The directory of the main package. This is useful for testing as the test
	// binary must be run in the directory of the tested package.
	MainDir string

	// ImportPath is the import path of the main package. This is useful for
	// correctly printing test results: the import path isn't always the same as
	// the path listed on the command line.
	ImportPath string
}

// packageAction is the struct that is serialized to JSON and hashed, to work as
// a cache key of compiled packages. It should contain all the information that
// goes into a compiled package to avoid using stale data.
//
// Right now it's still important to include a hash of every import, because a
// dependency might have a public constant that this package uses and thus this
// package will need to be recompiled if that constant changes. In the future,
// the type data should be serialized to disk which can then be used as cache
// key, avoiding the need for recompiling all dependencies when only the
// implementation of an imported package changes.
type packageAction struct {
	ImportPath       string
	CompilerVersion  int // compiler.Version
	InterpVersion    int // interp.Version
	LLVMVersion      string
	Config           *compiler.Config
	CFlags           []string
	FileHashes       map[string]string // hash of every file that's part of the package
	Imports          map[string]string // map from imported package to action ID hash
	OptLevel         int               // LLVM optimization level (0-3)
	SizeLevel        int               // LLVM optimization for size level (0-2)
	UndefinedGlobals []string          // globals that are left as external globals (no initializer)
}

// Build performs a single package to executable Go build. It takes in a package
// name, an output path, and set of compile options and from that it manages the
// whole compilation process.
//
// The error value may be of type *MultiError. Callers will likely want to check
// for this case and print such errors individually.
func Build(pkgName, outpath string, config *compileopts.Config, action func(BuildResult) error) error {
	// Create a temporary directory for intermediary files.
	dir, err := ioutil.TempDir("", "tinygo")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	compilerConfig := &compiler.Config{
		Triple:          config.Triple(),
		CPU:             config.CPU(),
		Features:        config.Features(),
		GOOS:            config.GOOS(),
		GOARCH:          config.GOARCH(),
		CodeModel:       config.CodeModel(),
		RelocationModel: config.RelocationModel(),

		Scheduler:          config.Scheduler(),
		FuncImplementation: config.FuncImplementation(),
		AutomaticStackSize: config.AutomaticStackSize(),
		DefaultStackSize:   config.Target.DefaultStackSize,
		NeedsStackObjects:  config.NeedsStackObjects(),
		Debug:              true,
		LLVMFeatures:       config.LLVMFeatures(),
	}

	// Load the target machine, which is the LLVM object that contains all
	// details of a target (alignment restrictions, pointer size, default
	// address spaces, etc).
	machine, err := compiler.NewTargetMachine(compilerConfig)
	if err != nil {
		return err
	}

	// Load entire program AST into memory.
	lprogram, err := loader.Load(config, []string{pkgName}, config.ClangHeaders, types.Config{
		Sizes: compiler.Sizes(machine),
	})
	if err != nil {
		return err
	}
	err = lprogram.Parse()
	if err != nil {
		return err
	}

	// Create the *ssa.Program. This does not yet build the entire SSA of the
	// program so it's pretty fast and doesn't need to be parallelized.
	program := lprogram.LoadSSA()

	// Add jobs to compile each package.
	// Packages that have a cache hit will not be compiled again.
	var packageJobs []*compileJob
	packageBitcodePaths := make(map[string]string)
	packageActionIDs := make(map[string]string)
	optLevel, sizeLevel, _ := config.OptLevels()
	for _, pkg := range lprogram.Sorted() {
		pkg := pkg // necessary to avoid a race condition

		var undefinedGlobals []string
		for name := range config.Options.GlobalValues[pkg.Pkg.Path()] {
			undefinedGlobals = append(undefinedGlobals, name)
		}
		sort.Strings(undefinedGlobals)

		// Create a cache key: a hash from the action ID below that contains all
		// the parameters for the build.
		actionID := packageAction{
			ImportPath:       pkg.ImportPath,
			CompilerVersion:  compiler.Version,
			InterpVersion:    interp.Version,
			LLVMVersion:      llvm.Version,
			Config:           compilerConfig,
			CFlags:           pkg.CFlags,
			FileHashes:       make(map[string]string, len(pkg.FileHashes)),
			Imports:          make(map[string]string, len(pkg.Pkg.Imports())),
			OptLevel:         optLevel,
			SizeLevel:        sizeLevel,
			UndefinedGlobals: undefinedGlobals,
		}
		for filePath, hash := range pkg.FileHashes {
			actionID.FileHashes[filePath] = hex.EncodeToString(hash)
		}
		for _, imported := range pkg.Pkg.Imports() {
			hash, ok := packageActionIDs[imported.Path()]
			if !ok {
				return fmt.Errorf("package %s imports %s but couldn't find dependency", pkg.ImportPath, imported.Path())
			}
			actionID.Imports[imported.Path()] = hash
		}
		buf, err := json.Marshal(actionID)
		if err != nil {
			panic(err) // shouldn't happen
		}
		hash := sha512.Sum512_224(buf)
		packageActionIDs[pkg.ImportPath] = hex.EncodeToString(hash[:])

		// Determine the path of the bitcode file (which is a serialized version
		// of a LLVM module).
		cacheDir := goenv.Get("GOCACHE")
		if cacheDir == "off" {
			// Use temporary build directory instead, effectively disabling the
			// build cache.
			cacheDir = dir
		}
		bitcodePath := filepath.Join(cacheDir, "pkg-"+hex.EncodeToString(hash[:])+".bc")
		packageBitcodePaths[pkg.ImportPath] = bitcodePath

		// Check whether this package has been compiled before, and if so don't
		// compile it again.
		if _, err := os.Stat(bitcodePath); err == nil {
			// Already cached, don't recreate this package.
			continue
		}

		// The package has not yet been compiled, so create a job to do so.
		job := &compileJob{
			description: "compile package " + pkg.ImportPath,
			run: func(*compileJob) error {
				// Compile AST to IR. The compiler.CompilePackage function will
				// build the SSA as needed.
				mod, errs := compiler.CompilePackage(pkg.ImportPath, pkg, program.Package(pkg.Pkg), machine, compilerConfig, config.DumpSSA())
				if errs != nil {
					return newMultiError(errs)
				}
				if err := llvm.VerifyModule(mod, llvm.PrintMessageAction); err != nil {
					return errors.New("verification error after compiling package " + pkg.ImportPath)
				}

				// Load bitcode of CGo headers and join the modules together.
				// This may seem vulnerable to cache problems, but this is not
				// the case: the Go code that was just compiled already tracks
				// all C files that are read and hashes them.
				// These headers could be compiled in parallel but the benefit
				// is so small that it's probably not worth parallelizing.
				// Packages are compiled independently anyway.
				for _, cgoHeader := range pkg.CGoHeaders {
					// Store the header text in a temporary file.
					f, err := ioutil.TempFile(dir, "cgosnippet-*.c")
					if err != nil {
						return err
					}
					_, err = f.Write([]byte(cgoHeader))
					if err != nil {
						return err
					}
					f.Close()

					// Compile the code (if there is any) to bitcode.
					flags := append([]string{"-c", "-emit-llvm", "-o", f.Name() + ".bc", f.Name()}, pkg.CFlags...)
					if config.Options.PrintCommands != nil {
						config.Options.PrintCommands("clang", flags...)
					}
					err = runCCompiler(flags...)
					if err != nil {
						return &commandError{"failed to build CGo header", "", err}
					}

					// Load and link the bitcode.
					// This makes it possible to optimize the functions defined
					// in the header together with the Go code. In particular,
					// this allows inlining. It also ensures there is only one
					// file per package to cache.
					headerMod, err := mod.Context().ParseBitcodeFile(f.Name() + ".bc")
					if err != nil {
						return fmt.Errorf("failed to load bitcode file: %w", err)
					}
					err = llvm.LinkModules(mod, headerMod)
					if err != nil {
						return fmt.Errorf("failed to link module: %w", err)
					}
				}

				// Erase all globals that are part of the undefinedGlobals list.
				// This list comes from the -ldflags="-X pkg.foo=val" option.
				// Instead of setting the value directly in the AST (which would
				// mean the value, which may be a secret, is stored in the build
				// cache), the global itself is left external (undefined) and is
				// only set at the end of the compilation.
				for _, name := range undefinedGlobals {
					globalName := pkg.Pkg.Path() + "." + name
					global := mod.NamedGlobal(globalName)
					if global.IsNil() {
						return errors.New("global not found: " + globalName)
					}
					name := global.Name()
					newGlobal := llvm.AddGlobal(mod, global.Type().ElementType(), name+".tmp")
					global.ReplaceAllUsesWith(newGlobal)
					global.EraseFromParentAsGlobal()
					newGlobal.SetName(name)
				}

				// Try to interpret package initializers at compile time.
				// It may only be possible to do this partially, in which case
				// it is completed after all IR files are linked.
				pkgInit := mod.NamedFunction(pkg.Pkg.Path() + ".init")
				if pkgInit.IsNil() {
					panic("init not found for " + pkg.Pkg.Path())
				}
				err := interp.RunFunc(pkgInit, config.DumpSSA())
				if err != nil {
					return err
				}
				if err := llvm.VerifyModule(mod, llvm.PrintMessageAction); err != nil {
					return errors.New("verification error after interpreting " + pkgInit.Name())
				}

				if sizeLevel >= 2 {
					// Set the "optsize" attribute to make slightly smaller
					// binaries at the cost of some performance.
					kind := llvm.AttributeKindID("optsize")
					attr := mod.Context().CreateEnumAttribute(kind, 0)
					for fn := mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
						fn.AddFunctionAttr(attr)
					}
				}

				// Run function passes for each function in the module.
				// These passes are intended to be run on each function right
				// after they're created to reduce IR size (and maybe also for
				// cache locality to improve performance), but for now they're
				// run here for each function in turn. Maybe this can be
				// improved in the future.
				builder := llvm.NewPassManagerBuilder()
				defer builder.Dispose()
				builder.SetOptLevel(optLevel)
				builder.SetSizeLevel(sizeLevel)
				funcPasses := llvm.NewFunctionPassManagerForModule(mod)
				defer funcPasses.Dispose()
				builder.PopulateFunc(funcPasses)
				funcPasses.InitializeFunc()
				for fn := mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
					if fn.IsDeclaration() {
						continue
					}
					funcPasses.RunFunc(fn)
				}
				funcPasses.FinalizeFunc()

				// Serialize the LLVM module as a bitcode file.
				// Write to a temporary path that is renamed to the destination
				// file to avoid race conditions with other TinyGo invocatiosn
				// that might also be compiling this package at the same time.
				f, err := ioutil.TempFile(filepath.Dir(bitcodePath), filepath.Base(bitcodePath))
				if err != nil {
					return err
				}
				if runtime.GOOS == "windows" {
					// Work around a problem on Windows.
					// For some reason, WriteBitcodeToFile causes TinyGo to
					// exit with the following message:
					//   LLVM ERROR: IO failure on output stream: Bad file descriptor
					buf := llvm.WriteBitcodeToMemoryBuffer(mod)
					defer buf.Dispose()
					_, err = f.Write(buf.Bytes())
				} else {
					// Otherwise, write bitcode directly to the file (probably
					// faster).
					err = llvm.WriteBitcodeToFile(mod, f)
				}
				if err != nil {
					// WriteBitcodeToFile doesn't produce a useful error on its
					// own, so create a somewhat useful error message here.
					return fmt.Errorf("failed to write bitcode for package %s to file %s", pkg.ImportPath, bitcodePath)
				}
				err = f.Close()
				if err != nil {
					return err
				}
				return os.Rename(f.Name(), bitcodePath)
			},
		}
		packageJobs = append(packageJobs, job)
	}

	// Add job that links and optimizes all packages together.
	var mod llvm.Module
	var stackSizeLoads []string
	programJob := &compileJob{
		description:  "link+optimize packages (LTO)",
		dependencies: packageJobs,
		run: func(*compileJob) error {
			// Load and link all the bitcode files. This does not yet optimize
			// anything, it only links the bitcode files together.
			ctx := llvm.NewContext()
			mod = ctx.NewModule("")
			for _, pkg := range lprogram.Sorted() {
				pkgMod, err := ctx.ParseBitcodeFile(packageBitcodePaths[pkg.ImportPath])
				if err != nil {
					return fmt.Errorf("failed to load bitcode file: %w", err)
				}
				err = llvm.LinkModules(mod, pkgMod)
				if err != nil {
					return fmt.Errorf("failed to link module: %w", err)
				}
			}

			// Create runtime.initAll function that calls the runtime
			// initializer of each package.
			llvmInitFn := mod.NamedFunction("runtime.initAll")
			llvmInitFn.SetLinkage(llvm.InternalLinkage)
			llvmInitFn.SetUnnamedAddr(true)
			llvmInitFn.Param(0).SetName("context")
			llvmInitFn.Param(1).SetName("parentHandle")
			block := mod.Context().AddBasicBlock(llvmInitFn, "entry")
			irbuilder := mod.Context().NewBuilder()
			defer irbuilder.Dispose()
			irbuilder.SetInsertPointAtEnd(block)
			i8ptrType := llvm.PointerType(mod.Context().Int8Type(), 0)
			for _, pkg := range lprogram.Sorted() {
				pkgInit := mod.NamedFunction(pkg.Pkg.Path() + ".init")
				if pkgInit.IsNil() {
					panic("init not found for " + pkg.Pkg.Path())
				}
				irbuilder.CreateCall(pkgInit, []llvm.Value{llvm.Undef(i8ptrType), llvm.Undef(i8ptrType)}, "")
			}
			irbuilder.CreateRetVoid()

			// After linking, functions should (as far as possible) be set to
			// private linkage or internal linkage. The compiler package marks
			// non-exported functions by setting the visibility to hidden or
			// (for thunks) to linkonce_odr linkage. Change the linkage here to
			// internal to benefit much more from interprocedural optimizations.
			for fn := mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
				if fn.Visibility() == llvm.HiddenVisibility {
					fn.SetVisibility(llvm.DefaultVisibility)
					fn.SetLinkage(llvm.InternalLinkage)
				} else if fn.Linkage() == llvm.LinkOnceODRLinkage {
					fn.SetLinkage(llvm.InternalLinkage)
				}
			}

			// Do the same for globals.
			for global := mod.FirstGlobal(); !global.IsNil(); global = llvm.NextGlobal(global) {
				if global.Visibility() == llvm.HiddenVisibility {
					global.SetVisibility(llvm.DefaultVisibility)
					global.SetLinkage(llvm.InternalLinkage)
				} else if global.Linkage() == llvm.LinkOnceODRLinkage {
					global.SetLinkage(llvm.InternalLinkage)
				}
			}

			if config.Options.PrintIR {
				fmt.Println("; Generated LLVM IR:")
				fmt.Println(mod.String())
			}

			// Run all optimization passes, which are much more effective now
			// that the optimizer can see the whole program at once.
			err := optimizeProgram(mod, config)
			if err != nil {
				return err
			}

			// Make sure stack sizes are loaded from a separate section so they can be
			// modified after linking.
			if config.AutomaticStackSize() {
				stackSizeLoads = transform.CreateStackSizeLoads(mod, config)
			}
			return nil
		},
	}

	// Check whether we only need to create an object file.
	// If so, we don't need to link anything and will be finished quickly.
	outext := filepath.Ext(outpath)
	if outext == ".o" || outext == ".bc" || outext == ".ll" {
		// Run jobs to produce the LLVM module.
		err := runJobs(programJob)
		if err != nil {
			return err
		}
		// Generate output.
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
			panic("unreachable")
		}
	}

	// Act as a compiler driver, as we need to produce a complete executable.
	// First add all jobs necessary to build this object file, then afterwards
	// run all jobs in parallel as far as possible.

	// Add job to write the output object file.
	objfile := filepath.Join(dir, "main.o")
	outputObjectFileJob := &compileJob{
		description:  "generate output file",
		dependencies: []*compileJob{programJob},
		result:       objfile,
		run: func(*compileJob) error {
			llvmBuf, err := machine.EmitToMemoryBuffer(mod, llvm.ObjectFile)
			if err != nil {
				return err
			}
			return ioutil.WriteFile(objfile, llvmBuf.Bytes(), 0666)
		},
	}

	// Prepare link command.
	linkerDependencies := []*compileJob{outputObjectFileJob}
	executable := filepath.Join(dir, "main")
	tmppath := executable // final file
	ldflags := append(config.LDFlags(), "-o", executable)

	// Add compiler-rt dependency if needed. Usually this is a simple load from
	// a cache.
	if config.Target.RTLib == "compiler-rt" {
		job, err := CompilerRT.load(config.Triple(), config.CPU(), dir)
		if err != nil {
			return err
		}
		linkerDependencies = append(linkerDependencies, job)
	}

	// Add jobs to compile extra files. These files are in C or assembly and
	// contain things like the interrupt vector table and low level operations
	// such as stack switching.
	root := goenv.Get("TINYGOROOT")
	for _, path := range config.ExtraFiles() {
		abspath := filepath.Join(root, path)
		job := &compileJob{
			description: "compile extra file " + path,
			run: func(job *compileJob) error {
				result, err := compileAndCacheCFile(abspath, dir, config.CFlags(), config.Options.PrintCommands)
				job.result = result
				return err
			},
		}
		linkerDependencies = append(linkerDependencies, job)
	}

	// Add jobs to compile C files in all packages. This is part of CGo.
	// TODO: do this as part of building the package to be able to link the
	// bitcode files together.
	for _, pkg := range lprogram.Sorted() {
		pkg := pkg
		for _, filename := range pkg.CFiles {
			abspath := filepath.Join(pkg.Dir, filename)
			job := &compileJob{
				description: "compile CGo file " + abspath,
				run: func(job *compileJob) error {
					result, err := compileAndCacheCFile(abspath, dir, pkg.CFlags, config.Options.PrintCommands)
					job.result = result
					return err
				},
			}
			linkerDependencies = append(linkerDependencies, job)
		}
	}

	// Linker flags from CGo lines:
	//     #cgo LDFLAGS: foo
	if len(lprogram.LDFlags) > 0 {
		ldflags = append(ldflags, lprogram.LDFlags...)
	}

	// Add libc dependency if needed.
	switch config.Target.Libc {
	case "picolibc":
		job, err := Picolibc.load(config.Triple(), config.CPU(), dir)
		if err != nil {
			return err
		}
		linkerDependencies = append(linkerDependencies, job)
	case "wasi-libc":
		path := filepath.Join(root, "lib/wasi-libc/sysroot/lib/wasm32-wasi/libc.a")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return errors.New("could not find wasi-libc, perhaps you need to run `make wasi-libc`?")
		}
		job := dummyCompileJob(path)
		linkerDependencies = append(linkerDependencies, job)
	case "":
		// no library specified, so nothing to do
	default:
		return fmt.Errorf("unknown libc: %s", config.Target.Libc)
	}

	// Strip debug information with -no-debug.
	if !config.Debug() {
		for _, tag := range config.BuildTags() {
			if tag == "baremetal" {
				// Don't use -no-debug on baremetal targets. It makes no sense:
				// the debug information isn't flashed to the device anyway.
				return fmt.Errorf("stripping debug information is unnecessary for baremetal targets")
			}
		}
		if config.Target.Linker == "wasm-ld" {
			// Don't just strip debug information, also compress relocations
			// while we're at it. Relocations can only be compressed when debug
			// information is stripped.
			ldflags = append(ldflags, "--strip-debug", "--compress-relocations")
		} else {
			switch config.GOOS() {
			case "linux":
				// Either real linux or an embedded system (like AVR) that
				// pretends to be Linux. It's a ELF linker wrapped by GCC in any
				// case.
				ldflags = append(ldflags, "-Wl,--strip-debug")
			case "darwin":
				// MacOS (darwin) doesn't have a linker flag to strip debug
				// information. Apple expects you to use the strip command
				// instead.
				return errors.New("cannot remove debug information: MacOS doesn't suppor this linker flag")
			default:
				// Other OSes may have different flags.
				return errors.New("cannot remove debug information: unknown OS: " + config.GOOS())
			}
		}
	}

	// Create a linker job, which links all object files together and does some
	// extra stuff that can only be done after linking.
	linkJob := &compileJob{
		description:  "link",
		dependencies: linkerDependencies,
		run: func(job *compileJob) error {
			for _, dependency := range job.dependencies {
				if dependency.result == "" {
					return errors.New("dependency without result: " + dependency.description)
				}
				ldflags = append(ldflags, dependency.result)
			}
			if config.Options.PrintCommands != nil {
				config.Options.PrintCommands(config.Target.Linker, ldflags...)
			}
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

			// Apply ELF patches
			if config.AutomaticStackSize() {
				// Modify the .tinygo_stacksizes section that contains a stack size
				// for each goroutine.
				err = modifyStackSizes(executable, stackSizeLoads, stackSizes)
				if err != nil {
					return fmt.Errorf("could not modify stack sizes: %w", err)
				}
			}
			if config.RP2040BootPatch() {
				// Patch the second stage bootloader CRC into the .boot2 section
				err = patchRP2040BootCRC(executable)
				if err != nil {
					return fmt.Errorf("could not patch RP2040 second stage boot loader: %w", err)
				}
			}

			// Print code size if requested.
			if config.Options.PrintSizes == "short" || config.Options.PrintSizes == "full" {
				packagePathMap := make(map[string]string, len(lprogram.Packages))
				for _, pkg := range lprogram.Sorted() {
					packagePathMap[pkg.OriginalDir()] = pkg.Pkg.Path()
				}
				sizes, err := loadProgramSize(executable, packagePathMap)
				if err != nil {
					return err
				}
				if config.Options.PrintSizes == "short" {
					fmt.Printf("   code    data     bss |   flash     ram\n")
					fmt.Printf("%7d %7d %7d | %7d %7d\n", sizes.Code+sizes.ROData, sizes.Data, sizes.BSS, sizes.Flash(), sizes.RAM())
				} else {
					if !config.Debug() {
						fmt.Println("warning: data incomplete, remove the -no-debug flag for more detail")
					}
					fmt.Printf("   code  rodata    data     bss |   flash     ram | package\n")
					fmt.Printf("------------------------------- | --------------- | -------\n")
					for _, name := range sizes.sortedPackageNames() {
						pkgSize := sizes.Packages[name]
						fmt.Printf("%7d %7d %7d %7d | %7d %7d | %s\n", pkgSize.Code, pkgSize.ROData, pkgSize.Data, pkgSize.BSS, pkgSize.Flash(), pkgSize.RAM(), name)
					}
					fmt.Printf("------------------------------- | --------------- | -------\n")
					fmt.Printf("%7d %7d %7d %7d | %7d %7d | total\n", sizes.Code, sizes.ROData, sizes.Data, sizes.BSS, sizes.Code+sizes.ROData+sizes.Data, sizes.Data+sizes.BSS)
				}
			}

			// Print goroutine stack sizes, as far as possible.
			if config.Options.PrintStacks {
				printStacks(calculatedStacks, stackSizes)
			}

			return nil
		},
	}

	// Run all jobs to compile and link the program.
	// Do this now (instead of after elf-to-hex and similar conversions) as it
	// is simpler and cannot be parallelized.
	err = runJobs(linkJob)
	if err != nil {
		return err
	}

	// Get an Intel .hex file or .bin file from the .elf file.
	outputBinaryFormat := config.BinaryFormat(outext)
	switch outputBinaryFormat {
	case "elf":
		// do nothing, file is already in ELF format
	case "hex", "bin":
		// Extract raw binary, either encoding it as a hex file or as a raw
		// firmware file.
		tmppath = filepath.Join(dir, "main"+outext)
		err := objcopy(executable, tmppath, outputBinaryFormat)
		if err != nil {
			return err
		}
	case "uf2":
		// Get UF2 from the .elf file.
		tmppath = filepath.Join(dir, "main"+outext)
		err := convertELFFileToUF2File(executable, tmppath, config.Target.UF2FamilyID)
		if err != nil {
			return err
		}
	case "esp32", "esp32c3", "esp8266":
		// Special format for the ESP family of chips (parsed by the ROM
		// bootloader).
		tmppath = filepath.Join(dir, "main"+outext)
		err := makeESPFirmareImage(executable, tmppath, outputBinaryFormat)
		if err != nil {
			return err
		}
	case "nrf-dfu":
		// special format for nrfutil for Nordic chips
		tmphexpath := filepath.Join(dir, "main.hex")
		err := objcopy(executable, tmphexpath, "hex")
		if err != nil {
			return err
		}
		tmppath = filepath.Join(dir, "main"+outext)
		err = makeDFUFirmwareImage(config.Options, tmphexpath, tmppath)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown output binary format: %s", outputBinaryFormat)
	}
	return action(BuildResult{
		Binary:     tmppath,
		MainDir:    lprogram.MainPkg().Dir,
		ImportPath: lprogram.MainPkg().ImportPath,
	})
}

// optimizeProgram runs a series of optimizations and transformations that are
// needed to convert a program to its final form. Some transformations are not
// optional and must be run as the compiler expects them to run.
func optimizeProgram(mod llvm.Module, config *compileopts.Config) error {
	err := interp.Run(mod, config.DumpSSA())
	if err != nil {
		return err
	}
	if config.VerifyIR() {
		// Only verify if we really need it.
		// The IR has already been verified before writing the bitcode to disk
		// and the interp function above doesn't need to do a lot as most of the
		// package initializers have already run. Additionally, verifying this
		// linked IR is _expensive_ because dead code hasn't been removed yet,
		// easily costing a few hundred milliseconds. Therefore, only do it when
		// specifically requested.
		if err := llvm.VerifyModule(mod, llvm.PrintMessageAction); err != nil {
			return errors.New("verification error after interpreting runtime.initAll")
		}
	}

	if config.GOOS() != "darwin" {
		transform.ApplyFunctionSections(mod) // -ffunction-sections
	}

	// Insert values from -ldflags="-X ..." into the IR.
	err = setGlobalValues(mod, config.Options.GlobalValues)
	if err != nil {
		return err
	}

	// Browsers cannot handle external functions that have type i64 because it
	// cannot be represented exactly in JavaScript (JS only has doubles). To
	// keep functions interoperable, pass int64 types as pointers to
	// stack-allocated values.
	// Use -wasm-abi=generic to disable this behaviour.
	if config.WasmAbi() == "js" {
		err := transform.ExternalInt64AsPtr(mod)
		if err != nil {
			return err
		}
	}

	// Optimization levels here are roughly the same as Clang, but probably not
	// exactly.
	optLevel, sizeLevel, inlinerThreshold := config.OptLevels()
	errs := transform.Optimize(mod, config, optLevel, sizeLevel, inlinerThreshold)
	if len(errs) > 0 {
		return newMultiError(errs)
	}
	if err := llvm.VerifyModule(mod, llvm.PrintMessageAction); err != nil {
		return errors.New("verification failure after LLVM optimization passes")
	}

	// LLVM 11 by default tries to emit tail calls (even with the target feature
	// disabled) unless it is explicitly disabled with a function attribute.
	// This is a problem, as it tries to emit them and prints an error when it
	// can't with this feature disabled.
	// Because as of september 2020 tail calls are not yet widely supported,
	// they need to be disabled until they are widely supported (at which point
	// the +tail-call target feautre can be set).
	if strings.HasPrefix(config.Triple(), "wasm") {
		transform.DisableTailCalls(mod)
	}

	return nil
}

// setGlobalValues sets the global values from the -ldflags="-X ..." compiler
// option in the given module. An error may be returned if the global is not of
// the expected type.
func setGlobalValues(mod llvm.Module, globals map[string]map[string]string) error {
	var pkgPaths []string
	for pkgPath := range globals {
		pkgPaths = append(pkgPaths, pkgPath)
	}
	sort.Strings(pkgPaths)
	for _, pkgPath := range pkgPaths {
		pkg := globals[pkgPath]
		var names []string
		for name := range pkg {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			value := pkg[name]
			globalName := pkgPath + "." + name
			global := mod.NamedGlobal(globalName)
			if global.IsNil() || !global.Initializer().IsNil() {
				// The global either does not exist (optimized away?) or has
				// some value, in which case it has already been initialized at
				// package init time.
				continue
			}

			// A strin is a {ptr, len} pair. We need these types to build the
			// initializer.
			initializerType := global.Type().ElementType()
			if initializerType.TypeKind() != llvm.StructTypeKind || initializerType.StructName() == "" {
				return fmt.Errorf("%s: not a string", globalName)
			}
			elementTypes := initializerType.StructElementTypes()
			if len(elementTypes) != 2 {
				return fmt.Errorf("%s: not a string", globalName)
			}

			// Create a buffer for the string contents.
			bufInitializer := mod.Context().ConstString(value, false)
			buf := llvm.AddGlobal(mod, bufInitializer.Type(), ".string")
			buf.SetInitializer(bufInitializer)
			buf.SetAlignment(1)
			buf.SetUnnamedAddr(true)
			buf.SetLinkage(llvm.PrivateLinkage)

			// Create the string value, which is a {ptr, len} pair.
			zero := llvm.ConstInt(mod.Context().Int32Type(), 0, false)
			ptr := llvm.ConstGEP(buf, []llvm.Value{zero, zero})
			if ptr.Type() != elementTypes[0] {
				return fmt.Errorf("%s: not a string", globalName)
			}
			length := llvm.ConstInt(elementTypes[1], uint64(len(value)), false)
			initializer := llvm.ConstNamedStruct(initializerType, []llvm.Value{
				ptr,
				length,
			})

			// Set the initializer. No initializer should be set at this point.
			global.SetInitializer(initializer)
		}
	}
	return nil
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
	data, fileHeader, err := getElfSectionData(executable, ".tinygo_stacksizes")
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
			stackSize := uint32(fn.stackSize)

			// Add stack size used by interrupts.
			switch fileHeader.Machine {
			case elf.EM_ARM:
				if stackSize%8 != 0 {
					// If the stack isn't a multiple of 8, it means the leaf
					// function with the biggest stack depth doesn't have an aligned
					// stack. If the STKALIGN flag is set (which it is by default)
					// the interrupt controller will forcibly align the stack before
					// storing in-use registers. This will thus overwrite one word
					// past the end of the stack (off-by-one).
					stackSize += 4
				}

				// On Cortex-M (assumed here), this stack size is 8 words or 32
				// bytes. This is only to store the registers that the interrupt
				// may modify, the interrupt will switch to the interrupt stack
				// (MSP).
				// Some background:
				// https://interrupt.memfault.com/blog/cortex-m-rtos-context-switching
				stackSize += 32

				// Adding 4 for the stack canary, and another 4 to keep the
				// stack aligned. Even though the size may be automatically
				// determined, stack overflow checking is still important as the
				// stack size cannot be determined for all goroutines.
				stackSize += 8
			default:
				return fmt.Errorf("unknown architecture: %s", fileHeader.Machine.String())
			}

			// Finally write the stack size to the binary.
			binary.LittleEndian.PutUint32(data[i*4:], stackSize)
		}
	}

	return replaceElfSection(executable, ".tinygo_stacksizes", data)
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

// RP2040 second stage bootloader CRC32 calculation
//
// Spec: https://datasheets.raspberrypi.org/rp2040/rp2040-datasheet.pdf
// Section: 2.8.1.3.1. Checksum
func patchRP2040BootCRC(executable string) error {
	bytes, _, err := getElfSectionData(executable, ".boot2")
	if err != nil {
		return err
	}

	if len(bytes) != 256 {
		return fmt.Errorf("rp2040 .boot2 section must be exactly 256 bytes")
	}

	// From the 'official' RP2040 checksum script:
	//
	//  Our bootrom CRC32 is slightly bass-ackward but it's
	//  best to work around for now (FIXME)
	//  100% worth it to save two Thumb instructions
	revBytes := make([]byte, len(bytes))
	for i := range bytes {
		revBytes[i] = bits.Reverse8(bytes[i])
	}

	// crc32.Update does an initial negate and negates the
	// result, so to meet RP2040 spec, pass 0x0 as initial
	// hash and negate returned value.
	//
	// Note: checksum is over 252 bytes (256 - 4)
	hash := bits.Reverse32(crc32.Update(0x0, crc32.IEEETable, revBytes[:252]) ^ 0xFFFFFFFF)

	// Write the CRC to the end of the bootloader.
	binary.LittleEndian.PutUint32(bytes[252:], hash)

	// Update the .boot2 section to included the CRC
	return replaceElfSection(executable, ".boot2", bytes)
}
