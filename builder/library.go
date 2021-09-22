package builder

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/goenv"
)

// Library is a container for information about a single C library, such as a
// compiler runtime or libc.
type Library struct {
	// The library name, such as compiler-rt or picolibc.
	name string

	// makeHeaders creates a header include dir for the library
	makeHeaders func(includeDir string) error

	// cflags returns the C flags specific to this library
	cflags func(headerPath string) []string

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
// The resulting directory may be stored in the provided tmpdir, which is
// expected to be removed after the Load call.
func (l *Library) Load(config *compileopts.Config, tmpdir string) (dir string, err error) {
	job, err := l.load(config, tmpdir)
	if err != nil {
		return "", err
	}
	err = runJobs(job, config.Options.Parallelism)
	return filepath.Dir(job.result), err
}

// load returns a compile job to build this library file for the given target
// and CPU. It may return a dummy compileJob if the library build is already
// cached. The path is stored as job.result but is only valid after the job has
// been run.
// The provided tmpdir will be used to store intermediary files and possibly the
// output archive file, it is expected to be removed after use.
// As a side effect, this call creates the library header files if they didn't
// exist yet.
func (l *Library) load(config *compileopts.Config, tmpdir string) (job *compileJob, err error) {
	outdir, precompiled := config.LibcPath(l.name)
	archiveFilePath := filepath.Join(outdir, "lib.a")
	if precompiled {
		// Found a precompiled library for this OS/architecture. Return the path
		// directly.
		return dummyCompileJob(archiveFilePath), nil
	}

	// Try to fetch this library from the cache.
	if _, err := os.Stat(archiveFilePath); err == nil {
		return dummyCompileJob(archiveFilePath), nil
	}
	// Cache miss, build it now.

	// Create the destination directory where the components of this library
	// (lib.a file, include directory) are placed.
	outname := filepath.Base(outdir)
	err = os.MkdirAll(filepath.Join(goenv.Get("GOCACHE"), outname), 0o777)
	if err != nil {
		// Could not create directory (and not because it already exists).
		return nil, err
	}

	// Make headers if needed.
	headerPath := filepath.Join(outdir, "include")
	if l.makeHeaders != nil {
		if _, err = os.Stat(headerPath); err != nil {
			temporaryHeaderPath, err := ioutil.TempDir(outdir, "include.tmp*")
			if err != nil {
				return nil, err
			}
			defer os.RemoveAll(temporaryHeaderPath)
			err = l.makeHeaders(temporaryHeaderPath)
			if err != nil {
				return nil, err
			}
			err = os.Rename(temporaryHeaderPath, headerPath)
			if err != nil {
				return nil, err
			}
		}
	}

	remapDir := filepath.Join(os.TempDir(), "tinygo-"+l.name)
	dir := filepath.Join(tmpdir, "build-lib-"+l.name)
	err = os.Mkdir(dir, 0777)
	if err != nil {
		return nil, err
	}

	// Precalculate the flags to the compiler invocation.
	// Note: -fdebug-prefix-map is necessary to make the output archive
	// reproducible. Otherwise the temporary directory is stored in the archive
	// itself, which varies each run.
	target := config.Triple()
	args := append(l.cflags(headerPath), "-c", "-Oz", "-g", "-ffunction-sections", "-fdata-sections", "-Wno-macro-redefined", "--target="+target, "-fdebug-prefix-map="+dir+"="+remapDir)
	cpu := config.CPU()
	if cpu != "" {
		args = append(args, "-mcpu="+cpu)
	}
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
	job = &compileJob{
		description: "ar " + l.name + "/lib.a",
		result:      filepath.Join(goenv.Get("GOCACHE"), outname, "lib.a"),
		run: func(*compileJob) error {
			// Create an archive of all object files.
			f, err := ioutil.TempFile(outdir, "libc.a.tmp*")
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
			// Store this archive in the cache.
			return os.Rename(f.Name(), archiveFilePath)
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
			run: func(*compileJob) error {
				var compileArgs []string
				compileArgs = append(compileArgs, args...)
				compileArgs = append(compileArgs, "-o", objpath, srcpath)
				err := runCCompiler(compileArgs...)
				if err != nil {
					return &commandError{"failed to build", srcpath, err}
				}
				return nil
			},
		})
	}

	return job, nil
}
