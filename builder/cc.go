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
	"io/ioutil"
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
// Its operation is a bit complex (more complex than Go package build caching)
// because the list of file dependencies is only known after the file is
// compiled. However, luckily compilers have a flag to write a list of file
// dependencies in Makefile syntax which can be used for caching.
//
// Because of this complexity, every file has in fact two cached build outputs:
// the file itself, and the list of dependencies. Its operation is as follows:
//
//	depfile = hash(path, compiler, cflags, ...)
//	if depfile exists:
//	  outfile = hash of all files and depfile name
//	  if outfile exists:
//	    # cache hit
//	    return outfile
//	# cache miss
//	tmpfile = compile file
//	read dependencies (side effect of compile)
//	write depfile
//	outfile = hash of all files and depfile name
//	rename tmpfile to outfile
//
// There are a few edge cases that are not handled:
//   - If a file is added to an include path, that file may be included instead of
//     some other file. This would be fixed by also including lookup failures in the
//     dependencies file, but I'm not aware of a compiler which does that.
//   - The Makefile syntax that compilers output has issues, see readDepFile for
//     details.
//   - A header file may be changed to add/remove an include. This invalidates the
//     depfile but without invalidating its name. For this reason, the depfile is
//     written on each new compilation (even when it seems unnecessary). However, it
//     could in rare cases lead to a stale file fetched from the cache.
func compileAndCacheCFile(abspath, tmpdir string, cflags []string, thinlto bool, printCommands func(string, ...string)) (string, error) {
	// Hash input file.
	fileHash, err := hashFile(abspath)
	if err != nil {
		return "", err
	}

	// Acquire a lock (if supported).
	unlock := lock(filepath.Join(goenv.Get("GOCACHE"), fileHash+".c.lock"))
	defer unlock()

	ext := ".o"
	if thinlto {
		ext = ".bc"
	}

	// Create cache key for the dependencies file.
	buf, err := json.Marshal(struct {
		Path        string
		Hash        string
		Flags       []string
		LLVMVersion string
	}{
		Path:        abspath,
		Hash:        fileHash,
		Flags:       cflags,
		LLVMVersion: llvm.Version,
	})
	if err != nil {
		panic(err) // shouldn't happen
	}
	depfileNameHashBuf := sha512.Sum512_224(buf)
	depfileNameHash := hex.EncodeToString(depfileNameHashBuf[:])

	// Load dependencies file, if possible.
	depfileName := "dep-" + depfileNameHash + ".json"
	depfileCachePath := filepath.Join(goenv.Get("GOCACHE"), depfileName)
	depfileBuf, err := ioutil.ReadFile(depfileCachePath)
	var dependencies []string // sorted list of dependency paths
	if err == nil {
		// There is a dependency file, that's great!
		// Parse it first.
		err := json.Unmarshal(depfileBuf, &dependencies)
		if err != nil {
			return "", fmt.Errorf("could not parse dependencies JSON: %w", err)
		}

		// Obtain hashes of all the files listed as a dependency.
		outpath, err := makeCFileCachePath(dependencies, depfileNameHash, ext)
		if err == nil {
			if _, err := os.Stat(outpath); err == nil {
				return outpath, nil
			} else if !os.IsNotExist(err) {
				return "", err
			}
		}
	} else if !os.IsNotExist(err) {
		// expected either nil or IsNotExist
		return "", err
	}

	objTmpFile, err := ioutil.TempFile(goenv.Get("GOCACHE"), "tmp-*"+ext)
	if err != nil {
		return "", err
	}
	objTmpFile.Close()
	depTmpFile, err := ioutil.TempFile(tmpdir, "dep-*.d")
	if err != nil {
		return "", err
	}
	depTmpFile.Close()
	flags := append([]string{}, cflags...)                                   // copy cflags
	flags = append(flags, "-MD", "-MV", "-MTdeps", "-MF", depTmpFile.Name()) // autogenerate dependencies
	if thinlto {
		flags = append(flags, "-flto=thin")
	}
	flags = append(flags, "-c", "-o", objTmpFile.Name(), abspath)
	if strings.ToLower(filepath.Ext(abspath)) == ".s" {
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
		return "", &commandError{"failed to build", abspath, err}
	}

	// Create sorted and uniqued slice of dependencies.
	dependencyPaths, err := readDepFile(depTmpFile.Name())
	if err != nil {
		return "", err
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

	// Write dependencies file.
	f, err := ioutil.TempFile(filepath.Dir(depfileCachePath), depfileName)
	if err != nil {
		return "", err
	}

	buf, err = json.MarshalIndent(dependencySlice, "", "\t")
	if err != nil {
		panic(err) // shouldn't happen
	}
	_, err = f.Write(buf)
	if err != nil {
		return "", err
	}
	err = f.Close()
	if err != nil {
		return "", err
	}
	err = os.Rename(f.Name(), depfileCachePath)
	if err != nil {
		return "", err
	}

	// Move temporary object file to final location.
	outpath, err := makeCFileCachePath(dependencySlice, depfileNameHash, ext)
	if err != nil {
		return "", err
	}
	err = os.Rename(objTmpFile.Name(), outpath)
	if err != nil {
		return "", err
	}

	return outpath, nil
}

// Create a cache path (a path in GOCACHE) to store the output of a compiler
// job. This path is based on the dep file name (which is a hash of metadata
// including compiler flags) and the hash of all input files in the paths slice.
func makeCFileCachePath(paths []string, depfileNameHash, ext string) (string, error) {
	// Hash all input files.
	fileHashes := make(map[string]string, len(paths))
	for _, path := range paths {
		hash, err := hashFile(path)
		if err != nil {
			return "", err
		}
		fileHashes[path] = hash
	}

	// Calculate a cache key based on the above hashes.
	buf, err := json.Marshal(struct {
		DepfileHash string
		FileHashes  map[string]string
	}{
		DepfileHash: depfileNameHash,
		FileHashes:  fileHashes,
	})
	if err != nil {
		panic(err) // shouldn't happen
	}
	outFileNameBuf := sha512.Sum512_224(buf)
	cacheKey := hex.EncodeToString(outFileNameBuf[:])

	outpath := filepath.Join(goenv.Get("GOCACHE"), "obj-"+cacheKey+ext)
	return outpath, nil
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
	buf, err := ioutil.ReadFile(filename)
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
	line := strings.SplitN(s, "\n", 2)[0]
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
