package builder

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/tinygo-org/tinygo/goenv"
)

// Library is a container for information about a single C library, such as a
// compiler runtime or libc.
type Library struct {
	// The library name, such as compiler-rt or picolibc.
	name string

	cflags func() []string

	// The source directory, relative to TINYGOROOT.
	sourceDir string

	// The source files, relative to sourceDir.
	sources func(target string) []string
}

// fullPath returns the full path to the source directory.
func (l *Library) fullPath() string {
	return filepath.Join(goenv.Get("TINYGOROOT"), l.sourceDir)
}

// sourcePaths returns a slice with the full paths to the source files.
func (l *Library) sourcePaths(target string) []string {
	sources := l.sources(target)
	paths := make([]string, len(sources))
	for i, name := range sources {
		paths[i] = filepath.Join(l.fullPath(), name)
	}
	return paths
}

// Load the library archive, possibly generating and caching it if needed.
// The resulting file is stored in the provided tmpdir, which is expected to be
// removed after the Load call.
func (l *Library) Load(target, tmpdir string) (path string, err error) {
	path, job, err := l.load(target, tmpdir)
	if err != nil {
		return "", err
	}
	if job != nil {
		jobs := append([]*compileJob{job}, job.dependencies...)
		err = runJobs(jobs)
	}
	return path, err
}

// load returns a path to the library file for the given target, loading it from
// cache if possible. It will return a non-zero compiler job if the library
// wasn't cached, this job (and its dependencies) must be run before the library
// path is valid.
// The provided tmpdir will be used to store intermediary files and possibly the
// output archive file, it is expected to be removed after use.
func (l *Library) load(target, tmpdir string) (path string, job *compileJob, err error) {
	// Try to load a precompiled library.
	precompiledPath := filepath.Join(goenv.Get("TINYGOROOT"), "pkg", target, l.name+".a")
	if _, err := os.Stat(precompiledPath); err == nil {
		// Found a precompiled library for this OS/architecture. Return the path
		// directly.
		return precompiledPath, nil, nil
	}

	outfile := l.name + "-" + target + ".a"

	// Try to fetch this library from the cache.
	if path, err := cacheLoad(outfile, commands["clang"][0], l.sourcePaths(target)); path != "" || err != nil {
		// Cache hit.
		return path, nil, err
	}
	// Cache miss, build it now.

	remapDir := filepath.Join(os.TempDir(), "tinygo-"+l.name)
	dir := filepath.Join(tmpdir, "build-lib-"+l.name)
	err = os.Mkdir(dir, 0777)
	if err != nil {
		return "", nil, err
	}

	// Precalculate the flags to the compiler invocation.
	// Note: -fdebug-prefix-map is necessary to make the output archive
	// reproducible. Otherwise the temporary directory is stored in the archive
	// itself, which varies each run.
	args := append(l.cflags(), "-c", "-Oz", "-g", "-ffunction-sections", "-fdata-sections", "-Wno-macro-redefined", "--target="+target, "-fdebug-prefix-map="+dir+"="+remapDir)
	if strings.HasPrefix(target, "arm") || strings.HasPrefix(target, "thumb") {
		args = append(args, "-fshort-enums", "-fomit-frame-pointer", "-mfloat-abi=soft")
	}
	if strings.HasPrefix(target, "riscv32-") {
		args = append(args, "-march=rv32imac", "-mabi=ilp32", "-fforce-enable-int128")
	}
	if strings.HasPrefix(target, "riscv64-") {
		args = append(args, "-march=rv64gc", "-mabi=lp64")
	}

	// Create job to put all the object files in a single archive. This archive
	// file is the (static) library file.
	var objs []string
	arpath := filepath.Join(dir, l.name+".a")
	job = &compileJob{
		description: "ar " + l.name + ".a",
		run: func() error {
			// Create an archive of all object files.
			err := makeArchive(arpath, objs)
			if err != nil {
				return err
			}
			// Store this archive in the cache.
			_, err = cacheStore(arpath, outfile, commands["clang"][0], l.sourcePaths(target))
			return err
		},
	}

	// Create jobs to compile all sources. These jobs are depended upon by the
	// archive job above, so must be run first.
	for _, srcpath := range l.sourcePaths(target) {
		srcpath := srcpath // avoid concurrency issues by redefining inside the loop
		objpath := filepath.Join(dir, filepath.Base(srcpath)+".o")
		objs = append(objs, objpath)
		job.dependencies = append(job.dependencies, &compileJob{
			description: "compile " + srcpath,
			run: func() error {
				var compileArgs []string
				compileArgs = append(compileArgs, args...)
				compileArgs = append(compileArgs, "-o", objpath, srcpath)
				err := runCCompiler("clang", compileArgs...)
				if err != nil {
					return &commandError{"failed to build", srcpath, err}
				}
				return nil
			},
		})
	}

	return arpath, job, nil
}
