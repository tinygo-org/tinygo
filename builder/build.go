// Package builder is the compiler driver of TinyGo. It takes in a package name
// and an output path, and outputs an executable. It manages the entire
// compilation pipeline in between.
package builder

import (
	"crypto/sha256"
	"crypto/sha512"
	"debug/elf"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"go/types"
	"hash/crc32"
	"io/fs"
	"math/bits"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/gofrs/flock"
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
	// The executable directly from the linker, usually including debug
	// information. Used for GDB for example.
	Executable string

	// A path to the output binary. It is stored in the tmpdir directory of the
	// Build function, so if it should be kept it must be copied or moved away.
	// It is often the same as Executable, but differs if the output format is
	// .hex for example (instead of the usual ELF).
	Binary string

	// The directory of the main package. This is useful for testing as the test
	// binary must be run in the directory of the tested package.
	MainDir string

	// The root of the Go module tree.  This is used for running tests in emulator
	// that restrict file system access to allow them to grant access to the entire
	// source tree they're likely to need to read testdata from.
	ModuleRoot string

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
	CompilerBuildID  string
	TinyGoVersion    string
	LLVMVersion      string
	Config           *compiler.Config
	CFlags           []string
	FileHashes       map[string]string // hash of every file that's part of the package
	EmbeddedFiles    map[string]string // hash of all the //go:embed files in the package
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
func Build(pkgName, outpath, tmpdir string, config *compileopts.Config) (BuildResult, error) {
	// Read the build ID of the tinygo binary.
	// Used as a cache key for package builds.
	compilerBuildID, err := ReadBuildID()
	if err != nil {
		return BuildResult{}, err
	}

	if config.Options.Work {
		fmt.Printf("WORK=%s\n", tmpdir)
	}

	// Look up the build cache directory, which is used to speed up incremental
	// builds.
	cacheDir := goenv.Get("GOCACHE")
	if cacheDir == "off" {
		// Use temporary build directory instead, effectively disabling the
		// build cache.
		cacheDir = tmpdir
	}

	// Check for a libc dependency.
	// As a side effect, this also creates the headers for the given libc, if
	// the libc needs them.
	root, err := loader.GetCachedGoroot(config)
	if err != nil {
		return BuildResult{}, err
	}
	var libcDependencies []*compileJob
	switch config.Target.Libc {
	case "darwin-libSystem":
		job := makeDarwinLibSystemJob(config, tmpdir)
		libcDependencies = append(libcDependencies, job)
	case "musl":
		job, unlock, err := Musl.load(config, tmpdir)
		if err != nil {
			return BuildResult{}, err
		}
		defer unlock()
		libcDependencies = append(libcDependencies, dummyCompileJob(filepath.Join(filepath.Dir(job.result), "crt1.o")))
		libcDependencies = append(libcDependencies, job)
	case "picolibc":
		libcJob, unlock, err := Picolibc.load(config, tmpdir)
		if err != nil {
			return BuildResult{}, err
		}
		defer unlock()
		libcDependencies = append(libcDependencies, libcJob)
	case "wasi-libc":
		path := filepath.Join(root, "lib/wasi-libc/sysroot/lib/wasm32-wasi/libc.a")
		if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
			return BuildResult{}, errors.New("could not find wasi-libc, perhaps you need to run `make wasi-libc`?")
		}
		libcDependencies = append(libcDependencies, dummyCompileJob(path))
	case "mingw-w64":
		_, unlock, err := MinGW.load(config, tmpdir)
		if err != nil {
			return BuildResult{}, err
		}
		unlock()
		libcDependencies = append(libcDependencies, makeMinGWExtraLibs(tmpdir)...)
	case "":
		// no library specified, so nothing to do
	default:
		return BuildResult{}, fmt.Errorf("unknown libc: %s", config.Target.Libc)
	}

	optLevel, sizeLevel, _ := config.OptLevels()
	compilerConfig := &compiler.Config{
		Triple:          config.Triple(),
		CPU:             config.CPU(),
		Features:        config.Features(),
		ABI:             config.ABI(),
		GOOS:            config.GOOS(),
		GOARCH:          config.GOARCH(),
		CodeModel:       config.CodeModel(),
		RelocationModel: config.RelocationModel(),
		SizeLevel:       sizeLevel,

		Scheduler:          config.Scheduler(),
		AutomaticStackSize: config.AutomaticStackSize(),
		DefaultStackSize:   config.StackSize(),
		NeedsStackObjects:  config.NeedsStackObjects(),
		Debug:              true,
	}

	// Load the target machine, which is the LLVM object that contains all
	// details of a target (alignment restrictions, pointer size, default
	// address spaces, etc).
	machine, err := compiler.NewTargetMachine(compilerConfig)
	if err != nil {
		return BuildResult{}, err
	}
	defer machine.Dispose()

	// Load entire program AST into memory.
	lprogram, err := loader.Load(config, pkgName, config.ClangHeaders, types.Config{
		Sizes: compiler.Sizes(machine),
	})
	if err != nil {
		return BuildResult{}, err
	}
	err = lprogram.Parse()
	if err != nil {
		return BuildResult{}, err
	}

	// Create the *ssa.Program. This does not yet build the entire SSA of the
	// program so it's pretty fast and doesn't need to be parallelized.
	program := lprogram.LoadSSA()

	// Add jobs to compile each package.
	// Packages that have a cache hit will not be compiled again.
	var packageJobs []*compileJob
	packageActionIDJobs := make(map[string]*compileJob)

	if config.Options.GlobalValues["runtime"]["buildVersion"] == "" {
		version := goenv.Version
		if strings.HasSuffix(goenv.Version, "-dev") && goenv.GitSha1 != "" {
			version += "-" + goenv.GitSha1
		}
		if config.Options.GlobalValues == nil {
			config.Options.GlobalValues = make(map[string]map[string]string)
		}
		if config.Options.GlobalValues["runtime"] == nil {
			config.Options.GlobalValues["runtime"] = make(map[string]string)
		}
		config.Options.GlobalValues["runtime"]["buildVersion"] = version
	}

	var embedFileObjects []*compileJob
	for _, pkg := range lprogram.Sorted() {
		pkg := pkg // necessary to avoid a race condition

		var undefinedGlobals []string
		for name := range config.Options.GlobalValues[pkg.Pkg.Path()] {
			undefinedGlobals = append(undefinedGlobals, name)
		}
		sort.Strings(undefinedGlobals)

		// Make compile jobs to load files to be embedded in the output binary.
		var actionIDDependencies []*compileJob
		allFiles := map[string][]*loader.EmbedFile{}
		for _, files := range pkg.EmbedGlobals {
			for _, file := range files {
				allFiles[file.Name] = append(allFiles[file.Name], file)
			}
		}
		for name, files := range allFiles {
			name := name
			files := files
			job := &compileJob{
				description: "make object file for " + name,
				run: func(job *compileJob) error {
					// Read the file contents in memory.
					path := filepath.Join(pkg.Dir, name)
					data, err := os.ReadFile(path)
					if err != nil {
						return err
					}

					// Hash the file.
					sum := sha256.Sum256(data)
					hexSum := hex.EncodeToString(sum[:16])

					for _, file := range files {
						file.Size = uint64(len(data))
						file.Hash = hexSum
						if file.NeedsData {
							file.Data = data
						}
					}

					job.result, err = createEmbedObjectFile(string(data), hexSum, name, pkg.OriginalDir(), tmpdir, compilerConfig)
					return err
				},
			}
			actionIDDependencies = append(actionIDDependencies, job)
			embedFileObjects = append(embedFileObjects, job)
		}

		// Action ID jobs need to know the action ID of all the jobs the package
		// imports.
		var importedPackages []*compileJob
		for _, imported := range pkg.Pkg.Imports() {
			job, ok := packageActionIDJobs[imported.Path()]
			if !ok {
				return BuildResult{}, fmt.Errorf("package %s imports %s but couldn't find dependency", pkg.ImportPath, imported.Path())
			}
			importedPackages = append(importedPackages, job)
			actionIDDependencies = append(actionIDDependencies, job)
		}

		// Create a job that will calculate the action ID for a package compile
		// job. The action ID is the cache key that is used for caching this
		// package.
		packageActionIDJob := &compileJob{
			description:  "calculate cache key for package " + pkg.ImportPath,
			dependencies: actionIDDependencies,
			run: func(job *compileJob) error {
				// Create a cache key: a hash from the action ID below that contains all
				// the parameters for the build.
				actionID := packageAction{
					ImportPath:       pkg.ImportPath,
					CompilerBuildID:  string(compilerBuildID),
					TinyGoVersion:    goenv.Version,
					LLVMVersion:      llvm.Version,
					Config:           compilerConfig,
					CFlags:           pkg.CFlags,
					FileHashes:       make(map[string]string, len(pkg.FileHashes)),
					EmbeddedFiles:    make(map[string]string, len(allFiles)),
					Imports:          make(map[string]string, len(pkg.Pkg.Imports())),
					OptLevel:         optLevel,
					SizeLevel:        sizeLevel,
					UndefinedGlobals: undefinedGlobals,
				}
				for filePath, hash := range pkg.FileHashes {
					actionID.FileHashes[filePath] = hex.EncodeToString(hash)
				}
				for name, files := range allFiles {
					actionID.EmbeddedFiles[name] = files[0].Hash
				}
				for i, imported := range pkg.Pkg.Imports() {
					actionID.Imports[imported.Path()] = importedPackages[i].result
				}
				buf, err := json.Marshal(actionID)
				if err != nil {
					return err // shouldn't happen
				}
				hash := sha512.Sum512_224(buf)
				job.result = hex.EncodeToString(hash[:])
				return nil
			},
		}
		packageActionIDJobs[pkg.ImportPath] = packageActionIDJob

		// Now create the job to actually build the package. It will exit early
		// if the package is already compiled.
		job := &compileJob{
			description:  "compile package " + pkg.ImportPath,
			dependencies: []*compileJob{packageActionIDJob},
			run: func(job *compileJob) error {
				job.result = filepath.Join(cacheDir, "pkg-"+packageActionIDJob.result+".bc")
				// Acquire a lock (if supported).
				unlock := lock(job.result + ".lock")
				defer unlock()

				if _, err := os.Stat(job.result); err == nil {
					// Already cached, don't recreate this package.
					return nil
				}

				// Compile AST to IR. The compiler.CompilePackage function will
				// build the SSA as needed.
				mod, errs := compiler.CompilePackage(pkg.ImportPath, pkg, program.Package(pkg.Pkg), machine, compilerConfig, config.DumpSSA())
				defer mod.Context().Dispose()
				defer mod.Dispose()
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
					f, err := os.CreateTemp(tmpdir, "cgosnippet-*.c")
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
					newGlobal := llvm.AddGlobal(mod, global.GlobalValueType(), name+".tmp")
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
				err := interp.RunFunc(pkgInit, config.Options.InterpTimeout, config.DumpSSA())
				if err != nil {
					return err
				}
				if err := llvm.VerifyModule(mod, llvm.PrintMessageAction); err != nil {
					return errors.New("verification error after interpreting " + pkgInit.Name())
				}

				transform.OptimizePackage(mod, config)

				// Serialize the LLVM module as a bitcode file.
				// Write to a temporary path that is renamed to the destination
				// file to avoid race conditions with other TinyGo invocatiosn
				// that might also be compiling this package at the same time.
				f, err := os.CreateTemp(filepath.Dir(job.result), filepath.Base(job.result))
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
					return fmt.Errorf("failed to write bitcode for package %s to file %s", pkg.ImportPath, job.result)
				}
				err = f.Close()
				if err != nil {
					return err
				}
				return os.Rename(f.Name(), job.result)
			},
		}
		packageJobs = append(packageJobs, job)
	}

	// Add job that links and optimizes all packages together.
	var mod llvm.Module
	defer func() {
		if !mod.IsNil() {
			ctx := mod.Context()
			mod.Dispose()
			ctx.Dispose()
		}
	}()
	var stackSizeLoads []string
	programJob := &compileJob{
		description:  "link+optimize packages (LTO)",
		dependencies: packageJobs,
		run: func(*compileJob) error {
			// Load and link all the bitcode files. This does not yet optimize
			// anything, it only links the bitcode files together.
			ctx := llvm.NewContext()
			mod = ctx.NewModule("main")
			for _, pkgJob := range packageJobs {
				pkgMod, err := ctx.ParseBitcodeFile(pkgJob.result)
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
			transform.AddStandardAttributes(llvmInitFn, config)
			llvmInitFn.Param(0).SetName("context")
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
				irbuilder.CreateCall(pkgInit.GlobalValueType(), pkgInit, []llvm.Value{llvm.Undef(i8ptrType)}, "")
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
		err := runJobs(programJob, config.Options.Semaphore)
		if err != nil {
			return BuildResult{}, err
		}
		// Generate output.
		switch outext {
		case ".o":
			llvmBuf, err := machine.EmitToMemoryBuffer(mod, llvm.ObjectFile)
			if err != nil {
				return BuildResult{}, err
			}
			defer llvmBuf.Dispose()
			return BuildResult{}, os.WriteFile(outpath, llvmBuf.Bytes(), 0666)
		case ".bc":
			var buf llvm.MemoryBuffer
			if config.UseThinLTO() {
				buf = llvm.WriteThinLTOBitcodeToMemoryBuffer(mod)
			} else {
				buf = llvm.WriteBitcodeToMemoryBuffer(mod)
			}
			defer buf.Dispose()
			return BuildResult{}, os.WriteFile(outpath, buf.Bytes(), 0666)
		case ".ll":
			data := []byte(mod.String())
			return BuildResult{}, os.WriteFile(outpath, data, 0666)
		default:
			panic("unreachable")
		}
	}

	// Act as a compiler driver, as we need to produce a complete executable.
	// First add all jobs necessary to build this object file, then afterwards
	// run all jobs in parallel as far as possible.

	// Add job to write the output object file.
	objfile := filepath.Join(tmpdir, "main.o")
	outputObjectFileJob := &compileJob{
		description:  "generate output file",
		dependencies: []*compileJob{programJob},
		result:       objfile,
		run: func(*compileJob) error {
			var llvmBuf llvm.MemoryBuffer
			if config.UseThinLTO() {
				llvmBuf = llvm.WriteThinLTOBitcodeToMemoryBuffer(mod)
			} else {
				var err error
				llvmBuf, err = machine.EmitToMemoryBuffer(mod, llvm.ObjectFile)
				if err != nil {
					return err
				}
			}
			defer llvmBuf.Dispose()
			return os.WriteFile(objfile, llvmBuf.Bytes(), 0666)
		},
	}

	// Prepare link command.
	linkerDependencies := []*compileJob{outputObjectFileJob}
	executable := filepath.Join(tmpdir, "main")
	if config.GOOS() == "windows" {
		executable += ".exe"
	}
	tmppath := executable // final file
	ldflags := append(config.LDFlags(), "-o", executable)

	// Add compiler-rt dependency if needed. Usually this is a simple load from
	// a cache.
	if config.Target.RTLib == "compiler-rt" {
		job, unlock, err := CompilerRT.load(config, tmpdir)
		if err != nil {
			return BuildResult{}, err
		}
		defer unlock()
		linkerDependencies = append(linkerDependencies, job)
	}

	// Add jobs to compile extra files. These files are in C or assembly and
	// contain things like the interrupt vector table and low level operations
	// such as stack switching.
	for _, path := range config.ExtraFiles() {
		abspath := filepath.Join(root, path)
		job := &compileJob{
			description: "compile extra file " + path,
			run: func(job *compileJob) error {
				result, err := compileAndCacheCFile(abspath, tmpdir, config.CFlags(), config.UseThinLTO(), config.Options.PrintCommands)
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
					result, err := compileAndCacheCFile(abspath, tmpdir, pkg.CFlags, config.UseThinLTO(), config.Options.PrintCommands)
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

	// Add libc dependencies, if they exist.
	linkerDependencies = append(linkerDependencies, libcDependencies...)

	// Add embedded files.
	linkerDependencies = append(linkerDependencies, embedFileObjects...)

	// Determine whether the compilation configuration would result in debug
	// (DWARF) information in the object files.
	var hasDebug = true
	if config.GOOS() == "darwin" {
		// Debug information isn't stored in the binary itself on MacOS but
		// is left in the object files by default. The binary does store the
		// path to these object files though.
		hasDebug = false
	}

	// Strip debug information with -no-debug.
	if hasDebug && !config.Debug() {
		if config.Target.Linker == "wasm-ld" {
			// Don't just strip debug information, also compress relocations
			// while we're at it. Relocations can only be compressed when debug
			// information is stripped.
			ldflags = append(ldflags, "--strip-debug", "--compress-relocations")
		} else if config.Target.Linker == "ld.lld" {
			// ld.lld is also used on Linux.
			ldflags = append(ldflags, "--strip-debug")
		} else {
			// Other linkers may have different flags.
			return BuildResult{}, errors.New("cannot remove debug information: unknown linker: " + config.Target.Linker)
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
			if config.UseThinLTO() {
				ldflags = append(ldflags, "-mllvm", "-mcpu="+config.CPU())
				if config.GOOS() == "windows" {
					// Options for the MinGW wrapper for the lld COFF linker.
					ldflags = append(ldflags,
						"-Xlink=/opt:lldlto="+strconv.Itoa(optLevel),
						"--thinlto-cache-dir="+filepath.Join(cacheDir, "thinlto"))
				} else if config.GOOS() == "darwin" {
					// Options for the ld64-compatible lld linker.
					ldflags = append(ldflags,
						"--lto-O"+strconv.Itoa(optLevel),
						"-cache_path_lto", filepath.Join(cacheDir, "thinlto"))
				} else {
					// Options for the ELF linker.
					ldflags = append(ldflags,
						"--lto-O"+strconv.Itoa(optLevel),
						"--thinlto-cache-dir="+filepath.Join(cacheDir, "thinlto"),
					)
				}
				if config.CodeModel() != "default" {
					ldflags = append(ldflags,
						"-mllvm", "-code-model="+config.CodeModel())
				}
				if sizeLevel >= 2 {
					// Workaround with roughly the same effect as
					// https://reviews.llvm.org/D119342.
					// Can hopefully be removed in LLVM 15.
					ldflags = append(ldflags,
						"-mllvm", "--rotation-max-header-size=0")
				}
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

			// Run wasm-opt for wasm binaries
			if arch := strings.Split(config.Triple(), "-")[0]; arch == "wasm32" {
				var opt string
				switch config.Options.Opt {
				case "none", "0":
					opt = "-O0"
				case "1":
					opt = "-O1"
				case "2":
					opt = "-O2"
				case "s":
					opt = "-Os"
				case "z":
					opt = "-Oz"
				default:
					return fmt.Errorf("unknown opt level: %q", config.Options.Opt)
				}

				var args []string

				if config.Scheduler() == "asyncify" {
					args = append(args, "--asyncify")
				}

				args = append(args,
					opt,
					"-g",
					executable,
					"--output", executable,
				)

				cmd := exec.Command(goenv.Get("WASMOPT"), args...)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr

				err := cmd.Run()
				if err != nil {
					return fmt.Errorf("wasm-opt failed: %w", err)
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
	err = runJobs(linkJob, config.Options.Semaphore)
	if err != nil {
		return BuildResult{}, err
	}

	// Get an Intel .hex file or .bin file from the .elf file.
	outputBinaryFormat := config.BinaryFormat(outext)
	switch outputBinaryFormat {
	case "elf":
		// do nothing, file is already in ELF format
	case "hex", "bin":
		// Extract raw binary, either encoding it as a hex file or as a raw
		// firmware file.
		tmppath = filepath.Join(tmpdir, "main"+outext)
		err := objcopy(executable, tmppath, outputBinaryFormat)
		if err != nil {
			return BuildResult{}, err
		}
	case "uf2":
		// Get UF2 from the .elf file.
		tmppath = filepath.Join(tmpdir, "main"+outext)
		err := convertELFFileToUF2File(executable, tmppath, config.Target.UF2FamilyID)
		if err != nil {
			return BuildResult{}, err
		}
	case "esp32", "esp32-img", "esp32c3", "esp8266":
		// Special format for the ESP family of chips (parsed by the ROM
		// bootloader).
		tmppath = filepath.Join(tmpdir, "main"+outext)
		err := makeESPFirmareImage(executable, tmppath, outputBinaryFormat)
		if err != nil {
			return BuildResult{}, err
		}
	case "nrf-dfu":
		// special format for nrfutil for Nordic chips
		tmphexpath := filepath.Join(tmpdir, "main.hex")
		err := objcopy(executable, tmphexpath, "hex")
		if err != nil {
			return BuildResult{}, err
		}
		tmppath = filepath.Join(tmpdir, "main"+outext)
		err = makeDFUFirmwareImage(config.Options, tmphexpath, tmppath)
		if err != nil {
			return BuildResult{}, err
		}
	default:
		return BuildResult{}, fmt.Errorf("unknown output binary format: %s", outputBinaryFormat)
	}

	// If there's a module root, use that.
	moduleroot := lprogram.MainPkg().Module.Dir
	if moduleroot == "" {
		// if not, just the regular root
		moduleroot = lprogram.MainPkg().Root
	}

	return BuildResult{
		Executable: executable,
		Binary:     tmppath,
		MainDir:    lprogram.MainPkg().Dir,
		ModuleRoot: moduleroot,
		ImportPath: lprogram.MainPkg().ImportPath,
	}, nil
}

// createEmbedObjectFile creates a new object file with the given contents, for
// the embed package.
func createEmbedObjectFile(data, hexSum, sourceFile, sourceDir, tmpdir string, compilerConfig *compiler.Config) (string, error) {
	// TODO: this works for small files, but can be a problem for larger files.
	// For larger files, it seems more appropriate to generate the object file
	// manually without going through LLVM.
	// On the other hand, generating DWARF like we do here can be difficult
	// without assistance from LLVM.

	// Create new LLVM module just for this file.
	ctx := llvm.NewContext()
	defer ctx.Dispose()
	mod := ctx.NewModule("data")
	defer mod.Dispose()

	// Create data global.
	value := ctx.ConstString(data, false)
	globalName := "embed/file_" + hexSum
	global := llvm.AddGlobal(mod, value.Type(), globalName)
	global.SetInitializer(value)
	global.SetLinkage(llvm.LinkOnceODRLinkage)
	global.SetGlobalConstant(true)
	global.SetUnnamedAddr(true)
	global.SetAlignment(1)
	if compilerConfig.GOOS != "darwin" {
		// MachO doesn't support COMDATs, while COFF requires it (to avoid
		// "duplicate symbol" errors). ELF works either way.
		// Therefore, only use a COMDAT on non-MachO systems (aka non-MacOS).
		global.SetComdat(mod.Comdat(globalName))
	}

	// Add DWARF debug information to this global, so that it is
	// correctly counted when compiling with the -size= flag.
	dibuilder := llvm.NewDIBuilder(mod)
	dibuilder.CreateCompileUnit(llvm.DICompileUnit{
		Language:  0xb, // DW_LANG_C99 (0xc, off-by-one?)
		File:      sourceFile,
		Dir:       sourceDir,
		Producer:  "TinyGo",
		Optimized: false,
	})
	ditype := dibuilder.CreateArrayType(llvm.DIArrayType{
		SizeInBits:  uint64(len(data)) * 8,
		AlignInBits: 8,
		ElementType: dibuilder.CreateBasicType(llvm.DIBasicType{
			Name:       "byte",
			SizeInBits: 8,
			Encoding:   llvm.DW_ATE_unsigned_char,
		}),
		Subscripts: []llvm.DISubrange{
			{
				Lo:    0,
				Count: int64(len(data)),
			},
		},
	})
	difile := dibuilder.CreateFile(sourceFile, sourceDir)
	diglobalexpr := dibuilder.CreateGlobalVariableExpression(difile, llvm.DIGlobalVariableExpression{
		Name:        globalName,
		File:        difile,
		Line:        1,
		Type:        ditype,
		Expr:        dibuilder.CreateExpression(nil),
		AlignInBits: 8,
	})
	global.AddMetadata(0, diglobalexpr)
	mod.AddNamedMetadataOperand("llvm.module.flags",
		ctx.MDNode([]llvm.Metadata{
			llvm.ConstInt(ctx.Int32Type(), 2, false).ConstantAsMetadata(), // Warning on mismatch
			ctx.MDString("Debug Info Version"),
			llvm.ConstInt(ctx.Int32Type(), 3, false).ConstantAsMetadata(),
		}),
	)
	mod.AddNamedMetadataOperand("llvm.module.flags",
		ctx.MDNode([]llvm.Metadata{
			llvm.ConstInt(ctx.Int32Type(), 7, false).ConstantAsMetadata(), // Max on mismatch
			ctx.MDString("Dwarf Version"),
			llvm.ConstInt(ctx.Int32Type(), 4, false).ConstantAsMetadata(),
		}),
	)
	dibuilder.Finalize()
	dibuilder.Destroy()

	// Write this LLVM module out as an object file.
	machine, err := compiler.NewTargetMachine(compilerConfig)
	if err != nil {
		return "", err
	}
	defer machine.Dispose()
	outfile, err := os.CreateTemp(tmpdir, "embed-"+hexSum+"-*.o")
	if err != nil {
		return "", err
	}
	defer outfile.Close()
	buf, err := machine.EmitToMemoryBuffer(mod, llvm.ObjectFile)
	if err != nil {
		return "", err
	}
	defer buf.Dispose()
	_, err = outfile.Write(buf.Bytes())
	if err != nil {
		return "", err
	}
	return outfile.Name(), outfile.Close()
}

// optimizeProgram runs a series of optimizations and transformations that are
// needed to convert a program to its final form. Some transformations are not
// optional and must be run as the compiler expects them to run.
func optimizeProgram(mod llvm.Module, config *compileopts.Config) error {
	err := interp.Run(mod, config.Options.InterpTimeout, config.DumpSSA())
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

	if config.GOOS() != "darwin" && !config.UseThinLTO() {
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
	if config.WasmAbi() == "js" {
		err := transform.ExternalInt64AsPtr(mod, config)
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
			initializerType := global.GlobalValueType()
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
			ptr := llvm.ConstGEP(bufInitializer.Type(), buf, []llvm.Value{zero, zero})
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
//	function                         stack usage (in bytes)
//	Reset_Handler                    316
//	examples/blinky2.led1            92
//	runtime.run$1                    300
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

// lock may acquire a lock at the specified path.
// It returns a function to release the lock.
// If flock is not supported, it does nothing.
func lock(path string) func() {
	flock := flock.New(path)
	err := flock.Lock()
	if err != nil {
		return func() {}
	}

	return func() { flock.Close() }
}
