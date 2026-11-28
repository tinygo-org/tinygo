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
	"slices"
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
	libraryHeaderPathPlaceholder = "\x00HEADER\x00"
	libraryBuildDirPlaceholder   = "\x00BUILDDIR\x00"
	libraryCacheDirPlaceholder   = "\x00GOCACHE\x00"
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
	ImplicitIncludes []string
	CompileInputs    map[string]map[string]string
	SystemHeaders    map[string]string `json:",omitempty"`
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

func configuredLibc(config *compileopts.Config) (*Library, error) {
	switch config.Target.Libc {
	case "musl":
		return &libMusl, nil
	case "picolibc":
		return &libPicolibc, nil
	case "wasi-libc":
		if slices.Contains(config.BuildTags(), "wasip2") {
			return &libWasiLibcWasip2, nil
		}
		return &libWasiLibc, nil
	case "wasmbuiltins":
		return &libWasmBuiltins, nil
	case "mingw-w64":
		return &libMinGW, nil
	case "darwin-libSystem", "glibc", "":
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown libc: %s", config.Target.Libc)
	}
}

func configuredLibraries(config *compileopts.Config) (configuredLibrarySet, error) {
	libc, err := configuredLibc(config)
	if err != nil {
		return configuredLibrarySet{}, err
	}
	if libc == &libWasiLibcWasip2 && llvm.VersionMajor < 22 {
		return configuredLibrarySet{}, fmt.Errorf("target wasip2 requires clang 22 or newer, but TinyGo uses LLVM %s", llvm.Version)
	}
	libraries := configuredLibrarySet{libc: libc}
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
	resourceDir := goenv.ClangResourceDir(config.TrimPath())
	if resourceDir == "" && !hasBuiltinTools {
		resourceDir = clangCompilerResourceDir
	}
	compileArgs := l.compileArgs(config, target, libraryHeaderPathPlaceholder, libraryBuildDirPlaceholder, resourceDir)

	input := libraryCacheInput{
		Name:             l.name,
		Target:           target,
		LibcNeedsMalloc:  config.LibcNeedsMalloc(),
		ArchiveFormat:    libraryArchiveFormatVersion,
		LLVMVersion:      llvm.Version,
		CompilerIdentity: compilerID,
		ResourceDir:      resourceDir,
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
	inputArgs := append([]string(nil), compileArgs...)
	coveredRoots := []string{sourceDir, inputDir}
	for name, key := range config.LibraryKeys {
		if key != "" {
			coveredRoots = append(coveredRoots, config.LibraryPath(name))
		}
	}
	compileInputs, err := hashLibraryCompileInputs(inputArgs, coveredRoots...)
	if err != nil {
		return nil, err
	}
	input.CompileInputs = compileInputs
	if l.needsLibc && config.Target.Libc == "glibc" {
		sema := config.Options.Semaphore
		sema <- struct{}{}
		defer func() { <-sema }()
		input.ImplicitIncludes, err = clangIncludePaths(expandLibraryCompileArgs(compileArgs, os.DevNull, os.DevNull))
		if err != nil {
			return nil, err
		}
		input.SystemHeaders, err = l.hashSystemHeaders(&input, coveredRoots, config.Options.PrintCommands)
		if err != nil {
			return nil, err
		}
	}
	generatedHeaders, err := l.hashGeneratedHeaders(target)
	if err != nil {
		return nil, err
	}
	input.GeneratedHeaders = generatedHeaders

	return &input, nil
}

func (l *Library) hashSystemHeaders(input *libraryCacheInput, coveredRoots []string, printCommands func(string, ...string)) (map[string]string, error) {
	dir, err := os.MkdirTemp("", "tinygo-lib-deps-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)
	headerDir := filepath.Join(dir, "include")
	if l.makeHeaders != nil {
		if err := os.MkdirAll(headerDir, 0o777); err != nil {
			return nil, err
		}
		if err := l.makeHeaders(input.Target, headerDir); err != nil {
			return nil, err
		}
	}
	roots := append(append([]string(nil), coveredRoots...), dir)
	if input.ResourceDir != "" {
		roots = append(roots, appendPath(input.ResourceDir, "include"))
	}
	args := expandLibraryCompileArgs(input.CompileArgs, headerDir, dir)
	sources := append([]librarySourceInput(nil), input.Sources...)
	if input.Crt1Source != "" {
		sources = append(sources, librarySourceInput{Path: input.Crt1Source})
	}
	hashes := make(map[string]string)
	for _, source := range sources {
		if strings.ToLower(filepath.Ext(source.Path)) == ".s" {
			continue
		}
		path := filepath.Join(input.SourceDir, filepath.FromSlash(source.Path))
		flags := append(append([]string(nil), args...), source.CFlags...)
		flags = append(flags, "-o", os.DevNull, path)
		dependencies, err := scanCDependencies(path, dir, flags, printCommands)
		if err != nil {
			return nil, err
		}
		for _, dependency := range dependencies {
			if libraryInputCovered(dependency, roots) {
				continue
			}
			if _, ok := hashes[dependency]; ok {
				continue
			}
			hash, err := hashFile(dependency)
			if err != nil {
				return nil, err
			}
			hashes[dependency] = hash
		}
	}
	return hashes, nil
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
	resolvedDir, err = filepath.Abs(resolvedDir)
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
		if entry.Type()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			if filepath.IsAbs(target) {
				parent, suffix := target, ""
				for {
					resolved, err := filepath.EvalSymlinks(parent)
					if err == nil {
						target = filepath.Join(resolved, suffix)
						break
					}
					if !errors.Is(err, fs.ErrNotExist) {
						return err
					}
					if filepath.Dir(parent) == parent {
						break
					}
					suffix = filepath.Join(filepath.Base(parent), suffix)
					parent = filepath.Dir(parent)
				}
				if strings.EqualFold(filepath.VolumeName(resolvedDir), filepath.VolumeName(target)) {
					target, err = filepath.Rel(resolvedDir, target)
					if err != nil {
						return err
					}
				}
			}
			hashes["\x00symlink:"+filepath.ToSlash(rel)] = target
		}
		info, err := os.Stat(path)
		if errors.Is(err, fs.ErrNotExist) && entry.Type()&os.ModeSymlink != 0 {
			continue
		}
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

func libraryInputCovered(path string, roots []string) bool {
	resolvedPath, err := filepath.EvalSymlinks(path)
	missing := err != nil
	if missing {
		// A keyed library may not have published its headers yet.
		if !errors.Is(err, fs.ErrNotExist) || filepath.Clean(path) != path {
			return false
		}
		resolvedPath = path
	}
	for _, root := range roots {
		resolvedRoot := root
		if !missing {
			var resolveErr error
			resolvedRoot, resolveErr = filepath.EvalSymlinks(root)
			if resolveErr != nil {
				continue
			}
		}
		rel, err := filepath.Rel(resolvedRoot, resolvedPath)
		if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

func hashLibraryCompileInputs(args []string, coveredRoots ...string) (map[string]map[string]string, error) {
	paths := compilerInputPaths(args)
	inputs := make(map[string]map[string]string, len(paths))
	for _, path := range paths {
		if strings.Contains(path, libraryHeaderPathPlaceholder) ||
			strings.Contains(path, libraryBuildDirPlaceholder) {
			continue
		}
		if libraryInputCovered(path, coveredRoots) {
			continue
		}
		_, err := os.Stat(path)
		if errors.Is(err, fs.ErrNotExist) {
			inputs[path] = nil
			continue
		}
		if err != nil {
			return nil, err
		}
		hashes, err := hashLibraryInputFiles(path)
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
				path := args[i]
				if arg == "-resource-dir" {
					path = appendPath(path, "include")
				}
				paths = append(paths, path)
			}
		default:
			for _, prefix := range []string{
				"-I", "-isystem", "-iquote", "-idirafter", "-include", "-imacros",
				"-resource-dir=", "--sysroot=", "-isysroot",
			} {
				if strings.HasPrefix(arg, prefix) && len(arg) != len(prefix) {
					path := strings.TrimPrefix(arg, prefix)
					if prefix == "-resource-dir=" {
						path = appendPath(path, "include")
					}
					paths = append(paths, path)
					break
				}
			}
		}
	}
	return paths
}

func (l *Library) compileArgs(config *compileopts.Config, target, headerPath, dir, resourceDir string) []string {
	remapDir := filepath.Join(".tinygo-build", l.name)
	if config.TrimPath() {
		remapDir = config.CSourcePath(filepath.Join("github.com/tinygo-org/tinygo", ".tinygo-build", l.name))
	}
	debugPrefixMap := "-fdebug-prefix-map=" + dir + "=" + remapDir
	if config.TrimPath() {
		debugPrefixMap = compileopts.DebugPrefixMap(dir, remapDir)
	}
	args := append(l.cflags(target, headerPath), "-c", "-Oz", "-gdwarf-4", "-ffunction-sections", "-fdata-sections", "-Wno-macro-redefined", "--target="+compileopts.LLVMTriple(target), debugPrefixMap)
	if config.TrimPath() {
		args = append(args,
			compileopts.FilePrefixMap(goenv.Get("TINYGOROOT"), config.CSourcePath("github.com/tinygo-org/tinygo")),
			compileopts.FilePrefixMap(libraryCacheDirPlaceholder, config.CSourcePath("tinygo-cache")),
			"-gno-record-command-line",
		)
		if resourceDir != "" {
			args = append(args, compileopts.FilePrefixMap(resourceDir, config.CSourcePath("clang")))
		}
	}
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
	if config.Target.Arch != "" {
		args = append(args, "-march="+config.Target.Arch)
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
		args = append(args, "-fforce-enable-int128")
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
		args = append(args, config.LibcHeaderPrefixMaps()...)
	}
	if config.TrimPath() && l.makeHeaders != nil {
		args = append(args, compileopts.FilePrefixMap(headerPath, config.CSourcePath("tinygo-cache/"+l.name+"/include")))
	}
	return appendCacheStableCFlags(args)
}

func expandLibraryCompileArgs(args []string, headerPath, dir string) []string {
	expanded := append([]string(nil), args...)
	for i, arg := range expanded {
		arg = strings.ReplaceAll(arg, libraryHeaderPathPlaceholder, headerPath)
		arg = strings.ReplaceAll(arg, libraryBuildDirPlaceholder, dir)
		arg = strings.ReplaceAll(arg, libraryCacheDirPlaceholder, goenv.Get("GOCACHE"))
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

	headerPath := filepath.Join(outdir, "include")
	if err := l.prepareHeaders(config.Triple(), outdir); err != nil {
		return nil, nil, err
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
		cleanpath := source.Path
		for strings.HasPrefix(cleanpath, "../") {
			cleanpath = cleanpath[3:]
		}
		srcpath := filepath.Join(sourceDir, filepath.FromSlash(source.Path))
		objpath := filepath.Join(dir, filepath.FromSlash(cleanpath)+".o")
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

func (l *Library) prepareHeaders(target, outdir string) error {
	if err := os.MkdirAll(outdir, 0o777); err != nil {
		return err
	}
	if l.makeHeaders == nil {
		return nil
	}
	headerPath := filepath.Join(outdir, "include")
	if _, err := os.Stat(headerPath); err == nil {
		return nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	temporaryHeaderPath, err := os.MkdirTemp(outdir, "include.tmp*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temporaryHeaderPath)
	if err := l.makeHeaders(target, temporaryHeaderPath); err != nil {
		return err
	}
	if err := os.Chmod(temporaryHeaderPath, 0o755); err != nil {
		return err
	}
	if err := os.Rename(temporaryHeaderPath, headerPath); err != nil {
		switch {
		case errors.Is(err, fs.ErrExist):
		case runtime.GOOS == "windows" && errors.Is(err, fs.ErrPermission):
			if _, statErr := os.Stat(headerPath); statErr != nil {
				return err
			}
		default:
			return err
		}
	}
	return nil
}
