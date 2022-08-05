package builder

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/goenv"
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

	// The source directory.
	sourceDir func() string

	// The source files, relative to sourceDir.
	librarySources func(target string) []string

	// The source code for the crt1.o file, relative to sourceDir.
	crt1Source string
}

// Load the library archive, possibly generating and caching it if needed.
// The resulting directory may be stored in the provided tmpdir, which is
// expected to be removed after the Load call.
func (l *Library) Load(config *compileopts.Config, tmpdir string) (dir string, err error) {
	job, unlock, err := l.load(config, tmpdir)
	if err != nil {
		return "", err
	}
	defer unlock()
	err = runJobs(job, config.Options.Semaphore)
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
func (l *Library) load(config *compileopts.Config, tmpdir string) (job *compileJob, abortLock func(), err error) {
	outdir, precompiled := config.LibcPath(l.name)
	archiveFilePath := filepath.Join(outdir, "lib.a")
	if precompiled {
		// Found a precompiled library for this OS/architecture. Return the path
		// directly.
		return dummyCompileJob(archiveFilePath), func() {}, nil
	}

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
				case os.IsExist(err):
					// Another invocation of TinyGo also seems to have already created the headers.

				case runtime.GOOS == "windows" && os.IsPermission(err):
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

	remapDir := filepath.Join(os.TempDir(), "tinygo-"+l.name)
	dir := filepath.Join(tmpdir, "build-lib-"+l.name)
	err = os.Mkdir(dir, 0777)
	if err != nil {
		return nil, nil, err
	}

	// Precalculate the flags to the compiler invocation.
	// Note: -fdebug-prefix-map is necessary to make the output archive
	// reproducible. Otherwise the temporary directory is stored in the archive
	// itself, which varies each run.
	args := append(l.cflags(target, headerPath), "-c", "-Oz", "-g", "-ffunction-sections", "-fdata-sections", "-Wno-macro-redefined", "--target="+target, "-fdebug-prefix-map="+dir+"="+remapDir)
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
	if strings.HasPrefix(target, "arm") || strings.HasPrefix(target, "thumb") {
		args = append(args, "-fshort-enums", "-fomit-frame-pointer", "-mfloat-abi=soft", "-fno-unwind-tables", "-fno-asynchronous-unwind-tables")
	}
	if strings.HasPrefix(target, "avr") {
		// AVR defaults to C float and double both being 32-bit. This deviates
		// from what most code (and certainly compiler-rt) expects. So we need
		// to force the compiler to use 64-bit floating point numbers for
		// double.
		args = append(args, "-mdouble=64")
	}
	if strings.HasPrefix(target, "riscv32-") {
		args = append(args, "-march=rv32imac", "-mabi=ilp32", "-fforce-enable-int128")
	}
	if strings.HasPrefix(target, "riscv64-") {
		args = append(args, "-march=rv64gc", "-mabi=lp64")
	}
	if strings.HasPrefix(target, "xtensa") {
		// Hack to work around an issue in the Xtensa port:
		// https://github.com/espressif/llvm-project/issues/52
		// Hopefully this will be fixed soon (LLVM 14).
		args = append(args, "-D__ELF__")
	}

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
			return os.Rename(f.Name(), archiveFilePath)
		},
	}

	sourceDir := l.sourceDir()

	// Create jobs to compile all sources. These jobs are depended upon by the
	// archive job above, so must be run first.
	for _, path := range l.librarySources(target) {
		// Strip leading "../" parts off the path.
		cleanpath := path
		for strings.HasPrefix(cleanpath, "../") {
			cleanpath = cleanpath[3:]
		}
		srcpath := filepath.Join(sourceDir, path)
		objpath := filepath.Join(dir, cleanpath+".o")
		os.MkdirAll(filepath.Dir(objpath), 0o777)
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

	// Create crt1.o job, if needed.
	// Add this as a (fake) dependency to the ar file so it gets compiled.
	// (It could be done in parallel with creating the ar file, but it probably
	// won't make much of a difference in speed).
	if l.crt1Source != "" {
		srcpath := filepath.Join(sourceDir, l.crt1Source)
		job.dependencies = append(job.dependencies, &compileJob{
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
				err = runCCompiler(compileArgs...)
				if err != nil {
					return &commandError{"failed to build", srcpath, err}
				}
				return os.Rename(tmpfile.Name(), filepath.Join(outdir, "crt1.o"))
			},
		})
	}

	ok = true
	return job, func() {
		once.Do(unlock)
	}, nil
}
