package loader

// This file constructs a new temporary GOROOT directory by merging both the
// standard Go GOROOT and the GOROOT from TinyGo using symlinks.
//
// The goal is to replace specific packages from Go with a TinyGo version. It's
// never a partial replacement, either a package is fully replaced or it is not.
// This is important because if we did allow to merge packages (e.g. by adding
// files to a package), it would lead to a dependency on implementation details
// with all the maintenance burden that results in. Only allowing to replace
// packages as a whole avoids this as packages are already designed to have a
// public (backwards-compatible) API.

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/goenv"
)

var gorootCreateMutex sync.Mutex

// GetCachedGoroot creates a new GOROOT by merging both the standard GOROOT and
// the GOROOT from TinyGo using lots of symbolic links.
func GetCachedGoroot(config *compileopts.Config) (string, error) {
	goroot := goenv.Get("GOROOT")
	if goroot == "" {
		return "", errors.New("could not determine GOROOT")
	}
	tinygoroot := goenv.Get("TINYGOROOT")
	if tinygoroot == "" {
		return "", errors.New("could not determine TINYGOROOT")
	}

	// Find the overrides needed for the goroot.
	overrides := pathsToOverride(config)

	// Resolve the merge links within the goroot.
	merge, err := listGorootMergeLinks(goroot, tinygoroot, overrides)
	if err != nil {
		return "", err
	}

	// Hash the merge links to create a cache key.
	data, err := json.Marshal(merge)
	if err != nil {
		return "", err
	}
	hash := sha512.Sum512_256(data)

	// Do not try to create the cached GOROOT in parallel, that's only a waste
	// of I/O bandwidth and thus speed. Instead, use a mutex to make sure only
	// one goroutine does it at a time.
	// This is not a way to ensure atomicity (a different TinyGo invocation
	// could be creating the same directory), but instead a way to avoid
	// creating it many times in parallel when running tests in parallel.
	gorootCreateMutex.Lock()
	defer gorootCreateMutex.Unlock()

	// Check if the goroot already exists.
	cachedGorootName := "goroot-" + hex.EncodeToString(hash[:])
	cachedgoroot := filepath.Join(goenv.Get("GOCACHE"), cachedGorootName)
	if _, err := os.Stat(cachedgoroot); err == nil {
		return cachedgoroot, nil
	}

	// Create the cache directory if it does not already exist.
	err = os.MkdirAll(goenv.Get("GOCACHE"), 0777)
	if err != nil {
		return "", err
	}

	// Create a temporary directory to construct the goroot within.
	tmpgoroot, err := os.MkdirTemp(goenv.Get("GOCACHE"), cachedGorootName+".tmp")
	if err != nil {
		return "", err
	}

	// Remove the temporary directory if it wasn't moved to the right place
	// (for example, when there was an error).
	defer os.RemoveAll(tmpgoroot)

	// Create the directory structure.
	// The directories are created in sorted order so that nested directories are created without extra work.
	{
		var dirs []string
		for dir, override := range overrides {
			if override.mode != gorootReplace {
				dirs = append(dirs, filepath.Join(tmpgoroot, "src", dir))
			}
		}
		sort.Strings(dirs)

		for _, dir := range dirs {
			err := os.Mkdir(dir, 0777)
			if err != nil {
				return "", err
			}
		}
	}

	// Create all symlinks.
	for dst, src := range merge {
		err := symlink(src, filepath.Join(tmpgoroot, dst))
		if err != nil {
			return "", err
		}
	}

	// Rename the new merged gorooot into place.
	err = os.Rename(tmpgoroot, cachedgoroot)
	if err != nil {
		if errors.Is(err, fs.ErrExist) {
			// Another invocation of TinyGo also seems to have created a GOROOT.
			// Use that one instead. Our new GOROOT will be automatically
			// deleted by the defer above.
			return cachedgoroot, nil
		}
		if runtime.GOOS == "windows" && errors.Is(err, fs.ErrPermission) {
			// On Windows, a rename with a destination directory that already
			// exists does not result in an IsExist error, but rather in an
			// access denied error. To be sure, check for this case by checking
			// whether the target directory exists.
			if _, err := os.Stat(cachedgoroot); err == nil {
				return cachedgoroot, nil
			}
		}
		return "", err
	}
	return cachedgoroot, nil
}

// listGorootMergeLinks searches goroot and tinygoroot for all symlinks that must be created within the merged goroot.
func listGorootMergeLinks(goroot, tinygoroot string, overrides map[string]gorootOverride) (map[string]string, error) {
	goSrc := filepath.Join(goroot, "src")
	tinygoSrc := filepath.Join(tinygoroot, "src")
	merges := make(map[string]string)
	for dir, override := range overrides {
		mode := override.mode
		if mode == gorootReplace {
			// Use the TinyGo version.
			merges[filepath.Join("src", dir)] = filepath.Join(tinygoSrc, dir)
			continue
		}

		// Add files from TinyGo.
		tinygoDir := filepath.Join(tinygoSrc, dir)
		tinygoEntries, err := os.ReadDir(tinygoDir)
		if err != nil {
			return nil, err
		}
		var hasTinyGoFiles bool
		for _, e := range tinygoEntries {
			name := e.Name()
			if e.IsDir() {
				if mode == gorootMergeTests {
					if _, ok := overrides[path.Join(dir, name)+"/"]; !ok {
						merges[filepath.Join("src", dir, name)] = filepath.Join(tinygoDir, name)
					}
				}
				continue
			}

			// Link this file.
			merges[filepath.Join("src", dir, name)] = filepath.Join(tinygoDir, name)

			hasTinyGoFiles = true
		}

		// Add entries from $GOROOT that are not overridden by TinyGo.
		goDir := filepath.Join(goSrc, dir)
		goEntries, err := os.ReadDir(goDir)
		if err != nil {
			return nil, err
		}
		for _, e := range goEntries {
			name := e.Name()
			if _, ok := overrides[path.Join(dir, name)+"/"]; ok {
				// This entry is overridden by TinyGo.
				// It has/will be merged elsewhere.
				continue
			}

			dst := filepath.Join("src", dir, name)
			isDir := e.IsDir()
			if mode == gorootMergeTests {
				if !isDir {
					if !strings.HasSuffix(name, "_test.go") || slices.Contains(override.skipTestFiles, name) {
						continue
					}
					if _, ok := merges[dst]; ok {
						return nil, fmt.Errorf("goroot merge conflict: %s exists in both TinyGo and upstream Go; rename the TinyGo file or list %q in skipTestFiles", dst, name)
					}
				} else if _, ok := merges[dst]; ok {
					continue
				}
			} else {
				if _, ok := merges[dst]; ok {
					continue
				}
				if hasTinyGoFiles && !isDir {
					// Only merge files from Go if TinyGo does not have any files.
					// Otherwise we'd end up with a weird mix from both Go
					// implementations.
					continue
				}
			}

			// Add a link to this entry
			merges[dst] = filepath.Join(goDir, name)
		}
	}

	// Merge the special directories from goroot.
	for _, dir := range []string{"bin", "lib", "pkg"} {
		merges[dir] = filepath.Join(goroot, dir)
	}

	// Required starting in Go 1.21 due to https://github.com/golang/go/issues/61928
	if _, err := os.Stat(filepath.Join(goroot, "go.env")); err == nil {
		merges["go.env"] = filepath.Join(goroot, "go.env")
	}

	return merges, nil
}

// needsSyscallPackage returns whether the syscall package should be overridden
// with the TinyGo version. This is the case on some targets.
func needsSyscallPackage(buildTags []string) bool {
	for _, tag := range buildTags {
		if tag == "baremetal" || tag == "nintendoswitch" || tag == "tinygo.wasm" {
			return true
		}
	}
	return false
}

type gorootOverrideMode uint8

const (
	gorootReplace gorootOverrideMode = iota
	gorootMergeSubdirs
	gorootMergeTests
)

type gorootOverride struct {
	mode          gorootOverrideMode
	skipTestFiles []string
}

func gorootMergeTestsWith(skipTestFiles ...string) gorootOverride {
	return gorootOverride{
		mode:          gorootMergeTests,
		skipTestFiles: skipTestFiles,
	}
}

// pathsToOverride returns the GOROOT directories overridden by TinyGo.
//
// gorootReplace links the TinyGo directory directly, gorootMergeSubdirs merges
// subdirectories, and gorootMergeTests keeps TinyGo's implementation while
// linking *_test.go files and subdirectories from GOROOT. Upstream test files
// that TinyGo replaces or omits must be listed in skipTestFiles.
func pathsToOverride(config *compileopts.Config) map[string]gorootOverride {
	replace := gorootOverride{mode: gorootReplace}
	mergeSubdirs := gorootOverride{mode: gorootMergeSubdirs}
	paths := map[string]gorootOverride{
		"":                            mergeSubdirs,
		"crypto/":                     mergeSubdirs,
		"crypto/internal/":            mergeSubdirs,
		"crypto/internal/boring/":     mergeSubdirs,
		"crypto/internal/boring/sig/": replace,
		// Go 1.26 added a CPU jitter entropy source for FIPS 140-3 that
		// allocates a 32 MiB global buffer (ScratchBuffer [1<<25]byte).
		// This is fine on systems with virtual memory, but causes RAM
		// overflow on microcontrollers. Replace with a zero-size stub
		// since TinyGo targets never use FIPS jitter entropy.
		"crypto/internal/entropy/":        mergeSubdirs,
		"crypto/internal/entropy/v1.0.0/": replace,
		"crypto/rand/":                    replace,
		"crypto/tls/":                     replace,
		"crypto/x509/":                    mergeSubdirs,
		"crypto/x509/internal/":           mergeSubdirs,
		"crypto/x509/internal/macos/":     replace,
		"device/":                         replace,
		"examples/":                       replace,
		"internal/":                       mergeSubdirs,
		"internal/abi/":                   replace,
		"internal/bytealg/":               replace,
		"internal/cm/":                    replace,
		"internal/futex/":                 replace,
		"internal/fuzz/":                  replace,
		"internal/itoa/":                  replace,
		"internal/reflectlite/":           replace,
		"internal/gclayout":               replace,
		"internal/task/":                  replace,
		"internal/wasi/":                  replace,
		"machine/":                        replace,
		"net/":                            mergeSubdirs,
		"net/http/":                       replace,
		"os/":                             mergeSubdirs,
		"reflect/":                        replace,
		"runtime/":                        replace,
		"sync/":                           mergeSubdirs,
		"testing/":                        mergeSubdirs,
		"tinygo/":                         replace,
		// Go 1.26 added SWAR optimizations to unicode/utf8 that use
		// constants assuming at least 32-bit uintptr. TinyGo supports
		// 16-bit targets (AVR) where these constants overflow, so we
		// provide a patched version.
		"unicode/":      mergeSubdirs,
		"unicode/utf8/": replace,
		"unique/":       replace,
	}

	if needsSyscallPackage(config.BuildTags()) {
		paths["internal/poll/"] = replace
		paths["syscall/"] = mergeSubdirs // include syscall/js
		paths["internal/syscall/"] = mergeSubdirs
		paths["internal/syscall/unix/"] = replace
	}

	return paths
}

// symlink creates a symlink or something similar. On Unix-like systems, it
// always creates a symlink. On Windows, it tries to create a symlink and if
// that fails, creates a hardlink or directory junction instead.
//
// Note that while Windows 10 does support symlinks and allows them to be
// created using os.Symlink, it requires developer mode to be enabled.
// Therefore provide a fallback for when symlinking is not possible.
// Unfortunately this fallback only works when TinyGo is installed on the same
// filesystem as the TinyGo cache and the Go installation (which is usually the
// C drive).
func symlink(oldname, newname string) error {
	symlinkErr := os.Symlink(oldname, newname)
	if runtime.GOOS == "windows" && symlinkErr != nil {
		// Fallback for when developer mode is disabled.
		// Note that we return the symlink error even if something else fails
		// later on. This is because symlinks are the easiest to support
		// (they're also used on Linux and MacOS) and enabling them is easy:
		// just enable developer mode.
		st, err := os.Stat(oldname)
		if err != nil {
			return symlinkErr
		}
		if st.IsDir() {
			// Make a directory junction. There may be a way to do this
			// programmatically, but it involves a lot of magic. Use the mklink
			// command built into cmd instead (mklink is a builtin, not an
			// external command).
			err := exec.Command("cmd", "/k", "mklink", "/J", newname, oldname).Run()
			if err != nil {
				return symlinkErr
			}
		} else {
			// Try making a hard link.
			err := os.Link(oldname, newname)
			if err != nil {
				// Making a hardlink failed. Try copying the file as a last
				// fallback.
				inf, err := os.Open(oldname)
				if err != nil {
					return err
				}
				defer inf.Close()
				outf, err := os.Create(newname)
				if err != nil {
					return err
				}
				defer outf.Close()
				_, err = io.Copy(outf, inf)
				if err != nil {
					os.Remove(newname)
					return err
				}
				// File was copied.
			}
		}
		return nil // success
	}
	return symlinkErr
}
