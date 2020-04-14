package builder

import (
	"io/ioutil"
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
func (l *Library) Load(target string) (path string, err error) {
	// Try to load a precompiled library.
	precompiledPath := filepath.Join(goenv.Get("TINYGOROOT"), "pkg", target, l.name+".a")
	if _, err := os.Stat(precompiledPath); err == nil {
		// Found a precompiled library for this OS/architecture. Return the path
		// directly.
		return precompiledPath, nil
	}

	outfile := l.name + "-" + target + ".a"

	// Try to fetch this library from the cache.
	if path, err := cacheLoad(outfile, commands["clang"][0], l.sourcePaths(target)); path != "" || err != nil {
		// Cache hit.
		return path, err
	}
	// Cache miss, build it now.

	dirPrefix := "tinygo-" + l.name
	remapDir := filepath.Join(os.TempDir(), dirPrefix)
	dir, err := ioutil.TempDir(os.TempDir(), dirPrefix)
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)

	// Precalculate the flags to the compiler invocation.
	args := append(l.cflags(), "-c", "-Oz", "-g", "-ffunction-sections", "-fdata-sections", "-Wno-macro-redefined", "--target="+target, "-fdebug-prefix-map="+dir+"="+remapDir)
	if strings.HasPrefix(target, "arm") || strings.HasPrefix(target, "thumb") {
		args = append(args, "-fshort-enums", "-fomit-frame-pointer", "-mfloat-abi=soft")
	}
	if strings.HasPrefix(target, "riscv32-") {
		args = append(args, "-march=rv32imac", "-mabi=ilp32", "-fforce-enable-int128")
	}
	if strings.HasPrefix(target, "avr-") {
		args = append(args, "-mdouble=64", "-mmcu=atmega1284p")
	}

	// Compile all sources.
	var objs []string
	for _, srcpath := range l.sourcePaths(target) {
		objpath := filepath.Join(dir, filepath.Base(srcpath)+".o")
		objs = append(objs, objpath)
		// Note: -fdebug-prefix-map is necessary to make the output archive
		// reproducible. Otherwise the temporary directory is stored in the
		// archive itself, which varies each run.
		err := runCCompiler("clang", append(args, "-o", objpath, srcpath)...)
		if err != nil {
			return "", &commandError{"failed to build", srcpath, err}
		}
	}

	// Put all the object files in a single archive. This archive file will be
	// used to statically link this library.
	arpath := filepath.Join(dir, l.name+".a")
	err = makeArchive(arpath, objs)
	if err != nil {
		return "", err
	}

	// Store this archive in the cache.
	return cacheStore(arpath, outfile, commands["clang"][0], l.sourcePaths(target))
}
