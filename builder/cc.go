package builder

// This file implements a wrapper around the C compiler (Clang) which uses a
// build cache.

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/tinygo-org/tinygo/goenv"
	"tinygo.org/x/go-llvm"
)

// compileAndCacheCFile compiles a C or assembly file using a build cache.
// Compiling the same file again (if nothing changed, including included header
// files) the output is loaded from the build cache instead.
//
// Its operation is a bit complex (more complex than Go package build caching),
// because the list of file dependencies depends on C include path resolution.
// TinyGo asks Clang for the current dependency list before looking for an object
// cache hit, then uses the hashes of those dependencies in the object key.
//
//	dependencies = clang -M source
//	outfile = hash(path, compiler, cflags, dependencies, ...)
//	if outfile exists:
//	  # cache hit
//	  return outfile
//	tmpfile = compile file
//	rename tmpfile to outfile
//
// The Makefile syntax that compilers output has issues, see readDepFile for
// details.
func compileAndCacheCFile(abspath, tmpdir string, cflags []string, printCommands func(string, ...string)) (string, error) {
	// Hash input file.
	fileHash, err := hashFile(abspath)
	if err != nil {
		return "", err
	}

	// Acquire a lock (if supported).
	unlock := lock(filepath.Join(goenv.Get("GOCACHE"), fileHash+".c.lock"))
	defer unlock()

	compilerID, err := clangCompilerIdentity()
	if err != nil {
		return "", err
	}

	dependencies, err := scanCFileDependencies(abspath, tmpdir, cflags, printCommands)
	if err != nil {
		return "", err
	}
	outpath, err := makeCFileCachePath(abspath, cFileCompileArgs(abspath, "$OBJ", cflags), compilerID, dependencies)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(outpath); err == nil {
		return outpath, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	objTmpFile, err := compileCFile(goenv.Get("GOCACHE"), abspath, cflags, printCommands)
	if err != nil {
		return "", err
	}
	if err := os.Rename(objTmpFile, outpath); err != nil {
		os.Remove(objTmpFile)
		return "", err
	}
	return outpath, nil
}

func scanCFileDependencies(abspath, tmpdir string, cflags []string, printCommands func(string, ...string)) ([]string, error) {
	depTmpFile, err := os.CreateTemp(tmpdir, "dep-*.d")
	if err != nil {
		return nil, err
	}
	depTmpFile.Close()
	defer os.Remove(depTmpFile.Name())

	flags := appendCacheStableCFlags(append([]string{}, cflags...))
	flags = append(flags, "-M", "-MV", "-MTdeps", "-MF", depTmpFile.Name(), abspath)
	if isAssemblyFile(abspath) {
		// If this is an assembly file (.s or .S, lowercase or uppercase), then
		// we'll need to add -Qunused-arguments because many parameters are
		// relevant to C, not assembly. And with -Werror, having meaningless
		// flags (for the assembler) is a compiler error.
		flags = append(flags, "-Qunused-arguments")
	}
	if printCommands != nil {
		printCommands("clang", flags...)
	}
	err = runCCompiler(flags...)
	if err != nil {
		return nil, &commandError{"failed to scan dependencies", abspath, err}
	}

	dependencyPaths, err := readDepFile(depTmpFile.Name())
	if err != nil {
		return nil, err
	}
	dependencyPaths = append(dependencyPaths, abspath) // necessary for .s files
	dependencySet := make(map[string]struct{}, len(dependencyPaths))
	var dependencySlice []string
	for _, path := range dependencyPaths {
		if _, ok := dependencySet[path]; ok {
			continue
		}
		dependencySet[path] = struct{}{}
		dependencySlice = append(dependencySlice, path)
	}
	sort.Strings(dependencySlice)
	return dependencySlice, nil
}

func cFileCompileArgs(abspath, objpath string, cflags []string) []string {
	flags := append([]string{}, cflags...)
	flags = append(flags, "-flto=thin")
	flags = append(flags, "-c", "-o", objpath, abspath)
	if isAssemblyFile(abspath) {
		flags = append(flags, "-Qunused-arguments")
	}
	return appendCacheStableCFlags(flags)
}

func compileCFile(cacheDir, abspath string, cflags []string, printCommands func(string, ...string)) (string, error) {
	objTmpFile, err := os.CreateTemp(cacheDir, "tmp-*.bc")
	if err != nil {
		return "", err
	}
	objTmpFile.Close()

	flags := cFileCompileArgs(abspath, objTmpFile.Name(), cflags)
	if printCommands != nil {
		printCommands("clang", flags...)
	}
	err = runCCompiler(flags...)
	if err != nil {
		os.Remove(objTmpFile.Name())
		return "", &commandError{"failed to build", abspath, err}
	}
	return objTmpFile.Name(), nil
}

// Create a cache path (a path in GOCACHE) to store the output of a compiler
// job. This path is based on the compiler identity, compiler flags, and the
// hash of all dependency files.
func makeCFileCachePath(path string, flags []string, compilerID string, dependencies []string) (string, error) {
	// Hash all input files.
	fileHashes := make(map[string]string, len(dependencies))
	for _, path := range dependencies {
		hash, err := hashFile(path)
		if err != nil {
			return "", err
		}
		fileHashes[path] = hash
	}

	// Calculate a cache key based on the above hashes.
	buf, err := json.Marshal(struct {
		Path             string
		Flags            []string
		LLVMVersion      string
		CompilerIdentity string
		FileHashes       map[string]string
	}{
		Path:             path,
		Flags:            flags,
		LLVMVersion:      llvm.Version,
		CompilerIdentity: compilerID,
		FileHashes:       fileHashes,
	})
	if err != nil {
		panic(err) // shouldn't happen
	}
	outFileNameBuf := sha512.Sum512_224(buf)
	cacheKey := hex.EncodeToString(outFileNameBuf[:])

	outpath := filepath.Join(goenv.Get("GOCACHE"), "obj-"+cacheKey+".bc")
	return outpath, nil
}

func isAssemblyFile(path string) bool {
	return strings.ToLower(filepath.Ext(path)) == ".s"
}

// hashFile hashes the given file path and returns the hash as a hex string.
func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to hash file: %w", err)
	}
	defer f.Close()
	fileHasher := sha512.New512_224()
	_, err = io.Copy(fileHasher, f)
	if err != nil {
		return "", fmt.Errorf("failed to hash file: %w", err)
	}
	return hex.EncodeToString(fileHasher.Sum(nil)), nil
}

// readDepFile reads a dependency file in NMake (Visual Studio make) format. The
// file is assumed to have a single target named deps.
//
// There are roughly three make syntax variants:
//   - BSD make, which doesn't support any escaping. This means that many special
//     characters are not supported in file names.
//   - GNU make, which supports escaping using a backslash but when it fails to
//     find a file it tries to fall back with the literal path name (to match BSD
//     make).
//   - NMake (Visual Studio) and Jom, which simply quote the string if there are
//     any weird characters.
//
// Clang supports two variants: a format that's a compromise between BSD and GNU
// make (and is buggy to match GCC which is equally buggy), and NMake/Jom, which
// is at least somewhat sane. This last format isn't perfect either: it does not
// correctly handle filenames with quote marks in them. Those are generally not
// allowed on Windows, but of course can be used on POSIX like systems. Still,
// it's the most sane of any of the formats so readDepFile will use that format.
func readDepFile(filename string) ([]string, error) {
	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	if len(buf) == 0 {
		return nil, nil
	}
	return parseDepFile(string(buf))
}

func parseDepFile(s string) ([]string, error) {
	// This function makes no attempt at parsing anything other than Clang -MD
	// -MV output.

	// For Windows: replace CRLF with LF to make the logic below simpler.
	s = strings.ReplaceAll(s, "\r\n", "\n")

	// Collapse all lines ending in a backslash. These backslashes are really
	// just a way to continue a line without making very long lines.
	s = strings.ReplaceAll(s, "\\\n", " ")

	// Only use the first line, which is expected to begin with "deps:".
	line, _, _ := strings.Cut(s, "\n")
	if !strings.HasPrefix(line, "deps:") {
		return nil, errors.New("readDepFile: expected 'deps:' prefix")
	}
	line = strings.TrimSpace(line[len("deps:"):])

	var deps []string
	for line != "" {
		if line[0] == '"' {
			// File path is quoted. Path ends with double quote.
			// This does not handle double quotes in path names, which is a
			// problem on non-Windows systems.
			line = line[1:]
			end := strings.IndexByte(line, '"')
			if end < 0 {
				return nil, errors.New("readDepFile: path is incorrectly quoted")
			}
			dep := line[:end]
			line = strings.TrimSpace(line[end+1:])
			deps = append(deps, dep)
		} else {
			// File path is not quoted. Path ends in space or EOL.
			end := strings.IndexFunc(line, unicode.IsSpace)
			if end < 0 {
				// last dependency
				deps = append(deps, line)
				break
			}
			dep := line[:end]
			line = strings.TrimSpace(line[end:])
			deps = append(deps, dep)
		}
	}
	return deps, nil
}
