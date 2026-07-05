package builder

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/goenv"
	"tinygo.org/x/go-llvm"
)

// Library is a container for information about a single C library, such as a
// compiler runtime or libc.
type Library struct {
	// The library name, such as compiler-rt or picolibc.
	name string

	// makeHeaders creates a header include dir for the library
	makeHeaders func(target, includeDir string) error

	// cflags returns the C flags specific to this library
	cflags func(target, headerPath string) []string

	// cflagsForFile returns additional C flags for a particular source file.
	cflagsForFile func(path string) []string

	// needsLibc is set to true if this library needs libc headers.
	needsLibc bool

	// The source directory.
	sourceDir func() string

	// The input directory that contains headers and other non-source inputs.
	inputDir func() string

	// The source files, relative to sourceDir.
	librarySources func(target string, libcNeedsMalloc bool) ([]string, error)

	// The source code for the crt1.o file, relative to sourceDir.
	crt1Source string
}

const (
	libraryHeaderPathPlaceholder = "$HEADER"
	libraryBuildDirPlaceholder   = "$BUILDDIR"
	// Increment when the archive construction changes in a way that can affect linking.
	libraryArchiveFormatVersion = 1
)

type librarySourceInput struct {
	Path   string
	Hash   string
	CFlags []string
}

type libraryCacheInput struct {
	Name             string
	Target           string
	LibcNeedsMalloc  bool
	ArchiveFormat    int
	LLVMVersion      string
	CompilerIdentity string
	ResourceDir      string
	SourceDir        string
	InputDir         string
	InputFiles       map[string]string
	CompileInputs    map[string]map[string]string
	GeneratedHeaders map[string]string
	CompileArgs      []string
	Sources          []librarySourceInput
	Crt1Source       string
	Crt1Hash         string
}

type configuredLibrarySet struct {
	libc   *Library
	linker []*Library
}

func configuredLibraries(config *compileopts.Config) (configuredLibrarySet, error) {
	var libraries configuredLibrarySet
	switch config.Target.Libc {
	case "musl":
		libraries.libc = &libMusl
	case "picolibc":
		libraries.libc = &libPicolibc
	case "wasi-libc":
		libraries.libc = &libWasiLibc
	case "wasmbuiltins":
		libraries.libc = &libWasmBuiltins
	case "mingw-w64":
		libraries.libc = &libMinGW
	case "darwin-libSystem", "":
		// These libc configurations don't use a Library-backed cache.
	default:
		return configuredLibrarySet{}, fmt.Errorf("unknown libc: %s", config.Target.Libc)
	}
	if config.Target.RTLib == "compiler-rt" {
		libraries.linker = append(libraries.linker, &libCompilerRT)
	}
	if config.GC() == "boehm" {
		libraries.linker = append(libraries.linker, &BoehmGC)
	}
	return libraries, nil
}

func (libraries configuredLibrarySet) all() []*Library {
	result := make([]*Library, 0, 1+len(libraries.linker))
	if libraries.libc != nil {
		result = append(result, libraries.libc)
	}
	result = append(result, libraries.linker...)
	return result
}

func makeLibraryCacheInputs(config *compileopts.Config, libraries configuredLibrarySet) (map[string]*libraryCacheInput, map[string]string, error) {
	inputs := make(map[string]*libraryCacheInput)
	keys := make(map[string]string)
	keyConfig := *config
	keyConfig.LibraryKeys = keys

	setKey := func(l *Library) error {
		if _, ok := keys[l.name]; ok {
			return nil
		}
		input, err := l.cacheInput(&keyConfig)
		if err != nil {
			return err
		}
		inputs[l.name] = input
		keys[l.name] = input.key()
		return nil
	}

	for _, library := range libraries.all() {
		if err := setKey(library); err != nil {
			return nil, nil, err
		}
	}
	return inputs, keys, nil
}

func (l *Library) cacheInput(config *compileopts.Config) (*libraryCacheInput, error) {
	target := config.Triple()
	sourceDir := l.sourceDir()
	sources, err := l.librarySources(target, config.LibcNeedsMalloc())
	if err != nil {
		return nil, err
	}

	inputDir := sourceDir
	if l.inputDir != nil {
		inputDir = l.inputDir()
	}
	compilerID, err := clangCompilerIdentity()
	if err != nil {
		return nil, err
	}
	compileArgs := l.compileArgs(config, target, libraryHeaderPathPlaceholder, libraryBuildDirPlaceholder)

	input := libraryCacheInput{
		Name:             l.name,
		Target:           target,
		LibcNeedsMalloc:  config.LibcNeedsMalloc(),
		ArchiveFormat:    libraryArchiveFormatVersion,
		LLVMVersion:      llvm.Version,
		CompilerIdentity: compilerID,
		ResourceDir:      goenv.ClangResourceDir(false),
		SourceDir:        sourceDir,
		InputDir:         inputDir,
		CompileArgs:      compileArgs,
		Sources:          make([]librarySourceInput, 0, len(sources)),
		Crt1Source:       l.crt1Source,
	}
	for _, source := range sources {
		hash, err := hashFile(filepath.Join(sourceDir, source))
		if err != nil {
			return nil, err
		}
		sourceInput := librarySourceInput{
			Path: filepath.ToSlash(source),
			Hash: hash,
		}
		if l.cflagsForFile != nil {
			sourceInput.CFlags = l.cflagsForFile(source)
		}
		input.Sources = append(input.Sources, sourceInput)
	}
	if l.crt1Source != "" {
		hash, err := hashFile(filepath.Join(sourceDir, l.crt1Source))
		if err != nil {
			return nil, err
		}
		input.Crt1Source = filepath.ToSlash(l.crt1Source)
		input.Crt1Hash = hash
	}
	inputFiles, err := hashLibraryInputFiles(inputDir)
	if err != nil {
		return nil, err
	}
	input.InputFiles = inputFiles
	compileInputs, err := hashLibraryCompileInputs(compileArgs, sourceDir, inputDir)
	if err != nil {
		return nil, err
	}
	input.CompileInputs = compileInputs
	generatedHeaders, err := l.hashGeneratedHeaders(target)
	if err != nil {
		return nil, err
	}
	input.GeneratedHeaders = generatedHeaders

	return &input, nil
}

func (input *libraryCacheInput) key() string {
	data, err := json.Marshal(input)
	if err != nil {
		panic(err)
	}
	sum := sha512.Sum512_224(data)
	return hex.EncodeToString(sum[:])
}

func (l *Library) hashGeneratedHeaders(target string) (map[string]string, error) {
	if l.makeHeaders == nil {
		return nil, nil
	}
	dir, err := os.MkdirTemp("", "tinygo-lib-headers-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)
	if err := l.makeHeaders(target, dir); err != nil {
		return nil, err
	}
	return hashLibraryInputFiles(dir)
}

func hashLibraryInputFiles(root string) (map[string]string, error) {
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return nil, err
	}
	hashes := map[string]string{}
	info, err := os.Stat(resolvedRoot)
	if err != nil {
		return nil, err
	}
	if info.Mode().IsRegular() {
		hash, err := hashFile(resolvedRoot)
		if err != nil {
			return nil, err
		}
		hashes["."] = hash
		return hashes, nil
	}
	if !info.IsDir() {
		return hashes, nil
	}
	err = hashLibraryInputDir(resolvedRoot, "", hashes, make(map[string]bool))
	return hashes, err
}

func hashLibraryInputDir(dir, prefix string, hashes map[string]string, ancestors map[string]bool) error {
	resolvedDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return err
	}
	if ancestors[resolvedDir] {
		return nil
	}
	ancestors[resolvedDir] = true
	defer delete(ancestors, resolvedDir)

	entries, err := os.ReadDir(resolvedDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		path := filepath.Join(resolvedDir, entry.Name())
		rel := filepath.Join(prefix, entry.Name())
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			if err := hashLibraryInputDir(path, rel, hashes, ancestors); err != nil {
				return err
			}
			continue
		}
		if !info.Mode().IsRegular() {
			continue
		}
		hash, err := hashFile(path)
		if err != nil {
			return err
		}
		hashes[filepath.ToSlash(rel)] = hash
	}
	return nil
}

func hashLibraryCompileInputs(args []string, coveredRoots ...string) (map[string]map[string]string, error) {
	paths := compilerInputPaths(args)
	inputs := make(map[string]map[string]string, len(paths))
	cacheDir := filepath.Clean(goenv.Get("GOCACHE"))
	for _, path := range paths {
		if strings.Contains(path, libraryHeaderPathPlaceholder) ||
			strings.Contains(path, libraryBuildDirPlaceholder) {
			continue
		}
		path = filepath.Clean(path)
		covered := false
		for _, root := range coveredRoots {
			rel, err := filepath.Rel(filepath.Clean(root), path)
			if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				covered = true
				break
			}
		}
		if covered {
			continue
		}
		if rel, err := filepath.Rel(cacheDir, path); err == nil &&
			rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			// Cached library include paths contain the dependency's content key.
			continue
		}
		hashes, err := hashLibraryInputFiles(path)
		if errors.Is(err, fs.ErrNotExist) {
			inputs[path] = nil
			continue
		}
		if err != nil {
			return nil, err
		}
		inputs[path] = hashes
	}
	return inputs, nil
}

func compilerInputPaths(args []string) []string {
	var paths []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-I", "-isystem", "-iquote", "-idirafter", "-include", "-imacros",
			"-resource-dir", "--sysroot", "-isysroot":
			if i+1 < len(args) {
				i++
				paths = append(paths, args[i])
			}
		default:
			for _, prefix := range []string{
				"-I", "-isystem", "-iquote", "-idirafter", "-include", "-imacros",
				"-resource-dir=", "--sysroot=", "-isysroot",
			} {
				if strings.HasPrefix(arg, prefix) && len(arg) != len(prefix) {
					paths = append(paths, strings.TrimPrefix(arg, prefix))
					break
				}
			}
		}
	}
	return paths
}

func (l *Library) compileArgs(config *compileopts.Config, target, headerPath, dir string) []string {
	remapDir := filepath.Join(os.TempDir(), "tinygo-"+l.name)
	args := append(l.cflags(target, headerPath), "-c", "-Oz", "-gdwarf-4", "-ffunction-sections", "-fdata-sections", "-Wno-macro-redefined", "--target="+compileopts.ClangTriple(target), "-fdebug-prefix-map="+dir+"="+remapDir)
	resourceDir := goenv.ClangResourceDir(false)
	if resourceDir != "" {
		args = append(args, "-resource-dir="+resourceDir)
	}
	cpu := config.CPU()
	if cpu != "" {
		// X86 has deprecated the -mcpu flag, so we need to use -march instead.
		// However, ARM has not done this.
		if strings.HasPrefix(target, "i386") || strings.HasPrefix(target, "x86_64") {
			args = append(args, "-march="+cpu)
		} else if strings.HasPrefix(target, "avr") {
			args = append(args, "-mmcu="+cpu)
		} else {
			args = append(args, "-mcpu="+cpu)
		}
	}
	if config.ABI() != "" {
		args = append(args, "-mabi="+config.ABI())
	}
	switch compileopts.CanonicalArchName(target) {
	case "arm":
		if strings.Split(target, "-")[2] == "linux" {
			args = append(args, "-fno-unwind-tables", "-fno-asynchronous-unwind-tables")
		} else {
			args = append(args, "-fshort-enums", "-fomit-frame-pointer", "-mfloat-abi=soft", "-fno-unwind-tables", "-fno-asynchronous-unwind-tables")
		}
	case "avr":
		// AVR defaults to C float and double both being 32-bit. This deviates
		// from what most code (and certainly compiler-rt) expects. So we need
		// to force the compiler to use 64-bit floating point numbers for
		// double.
		args = append(args, "-mdouble=64")
	case "riscv32":
		args = append(args, "-march="+riscvMarch(config, "rv32imac"), "-fforce-enable-int128")
	case "riscv64":
		args = append(args, "-march="+riscvMarch(config, "rv64gc"))
	case "mips":
		args = append(args, "-fno-pic")
	}
	if config.Target.SoftFloat {
		// Use softfloat instead of floating point instructions. This is
		// supported on many architectures.
		args = append(args, "-msoft-float")
	} else {
		if strings.HasPrefix(target, "armv5") {
			// On ARMv5 we need to explicitly enable hardware floating point
			// instructions: Clang appears to assume the hardware doesn't have a
			// FPU otherwise.
			args = append(args, "-mfpu=vfpv2")
		}
	}
	if l.needsLibc {
		args = append(args, config.LibcCFlags()...)
	}
	return appendCacheStableCFlags(args)
}

func expandLibraryCompileArgs(args []string, headerPath, dir string) []string {
	expanded := append([]string(nil), args...)
	for i, arg := range expanded {
		arg = strings.ReplaceAll(arg, libraryHeaderPathPlaceholder, headerPath)
		arg = strings.ReplaceAll(arg, libraryBuildDirPlaceholder, dir)
		expanded[i] = arg
	}
	return expanded
}

// load returns a compile job to build this library file for the given target
// and CPU. It may return a dummy compileJob if the library build is already
// cached. The path is stored as job.result but is only valid after the job has
// been run.
// The provided tmpdir will be used to store intermediary files and possibly the
// output archive file, it is expected to be removed after use.
// As a side effect, this call creates the library header files if they didn't
// exist yet.
func (l *Library) load(config *compileopts.Config, tmpdir string, input *libraryCacheInput) (job *compileJob, abortLock func(), err error) {
	key := input.key()
	if existingKey, ok := config.LibraryKeys[l.name]; ok {
		if existingKey != key {
			return nil, nil, fmt.Errorf("library cache key changed for %s", l.name)
		}
	} else {
		return nil, nil, fmt.Errorf("library cache key missing for %s", l.name)
	}
	outdir := config.LibraryPath(l.name)
	archiveFilePath := filepath.Join(outdir, "lib.a")

	// Create a lock on the output (if supported).
	// This is a bit messy, but avoids a deadlock because it is ordered consistently with other library loads within a build.
	outname := filepath.Base(outdir)
	unlock := lock(filepath.Join(goenv.Get("GOCACHE"), outname+".lock"))
	var ok bool
	defer func() {
		if !ok {
			unlock()
		}
	}()

	// Try to fetch this library from the cache.
	if _, err := os.Stat(archiveFilePath); err == nil {
		return dummyCompileJob(archiveFilePath), func() {}, nil
	}
	// Cache miss, build it now.

	// Create the destination directory where the components of this library
	// (lib.a file, include directory) are placed.
	err = os.MkdirAll(filepath.Join(goenv.Get("GOCACHE"), outname), 0o777)
	if err != nil {
		// Could not create directory (and not because it already exists).
		return nil, nil, err
	}

	// Make headers if needed.
	headerPath := filepath.Join(outdir, "include")
	target := config.Triple()
	if l.makeHeaders != nil {
		if _, err = os.Stat(headerPath); err != nil {
			temporaryHeaderPath, err := os.MkdirTemp(outdir, "include.tmp*")
			if err != nil {
				return nil, nil, err
			}
			defer os.RemoveAll(temporaryHeaderPath)
			err = l.makeHeaders(target, temporaryHeaderPath)
			if err != nil {
				return nil, nil, err
			}
			err = os.Chmod(temporaryHeaderPath, 0o755) // TempDir uses 0o700 by default
			if err != nil {
				return nil, nil, err
			}
			err = os.Rename(temporaryHeaderPath, headerPath)
			if err != nil {
				switch {
				case errors.Is(err, fs.ErrExist):
					// Another invocation of TinyGo also seems to have already created the headers.

				case runtime.GOOS == "windows" && errors.Is(err, fs.ErrPermission):
					// On Windows, a rename with a destination directory that already
					// exists does not result in an IsExist error, but rather in an
					// access denied error. To be sure, check for this case by checking
					// whether the target directory exists.
					if _, err := os.Stat(headerPath); err == nil {
						break
					}
					fallthrough

				default:
					return nil, nil, err
				}
			}
		}
	}

	dir := filepath.Join(tmpdir, "build-lib-"+l.name)
	err = os.Mkdir(dir, 0777)
	if err != nil {
		return nil, nil, err
	}

	// Precalculate the flags to the compiler invocation.
	// Note: -fdebug-prefix-map is necessary to make the output archive
	// reproducible. Otherwise the temporary directory is stored in the archive
	// itself, which varies each run.
	args := expandLibraryCompileArgs(input.CompileArgs, headerPath, dir)

	var once sync.Once

	// Create job to put all the object files in a single archive. This archive
	// file is the (static) library file.
	var objs []string
	job = &compileJob{
		description: "ar " + l.name + "/lib.a",
		result:      filepath.Join(goenv.Get("GOCACHE"), outname, "lib.a"),
		run: func(*compileJob) error {
			defer once.Do(unlock)

			// Create an archive of all object files.
			f, err := os.CreateTemp(outdir, "libc.a.tmp*")
			if err != nil {
				return err
			}
			err = makeArchive(f, objs)
			if err != nil {
				return err
			}
			err = f.Close()
			if err != nil {
				return err
			}
			err = os.Chmod(f.Name(), 0o644) // TempFile uses 0o600 by default
			if err != nil {
				return err
			}
			// Store this archive in the cache.
			return robustRename(f.Name(), archiveFilePath)
		},
	}

	sourceDir := input.SourceDir

	// Create jobs to compile all sources. These jobs are depended upon by the
	// archive job above, so must be run first.
	for _, source := range input.Sources {
		// Strip leading "../" parts off the path.
		source := source
		path := filepath.FromSlash(source.Path)
		cleanpath := path
		for strings.HasPrefix(cleanpath, "../") {
			cleanpath = cleanpath[3:]
		}
		srcpath := filepath.Join(sourceDir, path)
		objpath := filepath.Join(dir, cleanpath+".o")
		os.MkdirAll(filepath.Dir(objpath), 0o777)
		objs = append(objs, objpath)
		objfile := &compileJob{
			description: "compile " + srcpath,
			run: func(*compileJob) error {
				var compileArgs []string
				compileArgs = append(compileArgs, args...)
				compileArgs = append(compileArgs, source.CFlags...)
				compileArgs = append(compileArgs, "-o", objpath, srcpath)
				if config.Options.PrintCommands != nil {
					config.Options.PrintCommands("clang", compileArgs...)
				}
				err := runCCompiler(compileArgs...)
				if err != nil {
					return &commandError{"failed to build", srcpath, err}
				}
				return nil
			},
		}
		job.dependencies = append(job.dependencies, objfile)
	}

	// Create crt1.o job, if needed.
	// Add this as a (fake) dependency to the ar file so it gets compiled.
	// (It could be done in parallel with creating the ar file, but it probably
	// won't make much of a difference in speed).
	if l.crt1Source != "" {
		srcpath := filepath.Join(sourceDir, l.crt1Source)
		crt1Job := &compileJob{
			description: "compile " + srcpath,
			run: func(*compileJob) error {
				var compileArgs []string
				compileArgs = append(compileArgs, args...)
				tmpfile, err := os.CreateTemp(outdir, "crt1.o.tmp*")
				if err != nil {
					return err
				}
				tmpfile.Close()
				compileArgs = append(compileArgs, "-o", tmpfile.Name(), srcpath)
				if config.Options.PrintCommands != nil {
					config.Options.PrintCommands("clang", compileArgs...)
				}
				err = runCCompiler(compileArgs...)
				if err != nil {
					return &commandError{"failed to build", srcpath, err}
				}
				return os.Rename(tmpfile.Name(), filepath.Join(outdir, "crt1.o"))
			},
		}
		job.dependencies = append(job.dependencies, crt1Job)
	}

	ok = true
	return job, func() {
		once.Do(unlock)
	}, nil
}

// riscvMarch returns the -march value for RISC-V library compilation.
// It extracts the value from the target's cflags if present, otherwise
// falls back to the provided default. This ensures libraries are compiled
// with the correct ISA extensions for each target (e.g. rv32imc for
// ESP32-C3 which lacks the atomic extension).
func riscvMarch(config *compileopts.Config, defaultMarch string) string {
	for _, flag := range config.Target.CFlags {
		if strings.HasPrefix(flag, "-march=") {
			return flag[len("-march="):]
		}
	}
	return defaultMarch
}
