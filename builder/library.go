package builder

import (
	"errors"
	"io/fs"
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
	librarySources func(target string) ([]string, error)

	// The source code for the crt1.o file, relative to sourceDir.
	crt1Source string
}

// load returns a compile job to build this library file for the given target
// and CPU. It may return a dummy compileJob if the library build is already
// cached. The path is stored as job.result but is only valid after the job has
// been run.
// The provided tmpdir will be used to store intermediary files and possibly the
// output archive file, it is expected to be removed after use.
// As a side effect, this call creates the library header files if they didn't
// exist yet.
// The provided libc job (if not null) will cause this libc to be added as a
// dependency for all C compiler jobs, and adds libc headers for the given
// target config. In other words, pass this libc if the library needs a libc to
// compile.
func (l *Library) load(config *compileopts.Config, tmpdir string, libc *compileJob) (job *compileJob, abortLock func(), err error) {
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
	args := append(l.cflags(target, headerPath), "-c", "-Oz", "-gdwarf-4", "-ffunction-sections", "-fdata-sections", "-Wno-macro-redefined", "--target="+target, "-fdebug-prefix-map="+dir+"="+remapDir)
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
		args = append(args, "-march=rv32imac", "-fforce-enable-int128")
	case "riscv64":
		args = append(args, "-march=rv64gc")
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
	if libc != nil {
		args = append(args, config.LibcCFlags()...)
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
	paths, err := l.librarySources(target)
	if err != nil {
		return nil, nil, err
	}
	for _, path := range paths {
		// Strip leading "../" parts off the path.
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
		if libc != nil {
			objfile.dependencies = append(objfile.dependencies, libc)
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
		if libc != nil {
			crt1Job.dependencies = append(crt1Job.dependencies, libc)
		}
		job.dependencies = append(job.dependencies, crt1Job)
	}

	ok = true
	return job, func() {
		once.Do(unlock)
	}, nil
}
