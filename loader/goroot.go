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
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
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
	overrides := pathsToOverride(config.GoMinorVersion, needsSyscallPackage(config.BuildTags()))

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
	config.CacheDir = cachedgoroot
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
		for dir, merge := range overrides {
			if merge {
				dirs = append(dirs, filepath.Join(tmpgoroot, "src", dir))
			}
		}
		for dst, src := range merge {
			info, err := os.Stat(src)
			if err != nil {
				return "", err
			}
			if !info.IsDir() {
				dirs = append(dirs, filepath.Join(tmpgoroot, filepath.Dir(dst)))
			}
		}
		sort.Strings(dirs)

		var lastDir string
		for _, dir := range dirs {
			if dir == lastDir {
				continue
			}
			err := os.MkdirAll(dir, 0777)
			if err != nil {
				return "", err
			}
			lastDir = dir
		}
	}

	// TODO: Reduce merge symlinks by grouping symlinks with top-most base directory
	// that doesn't mix different source base directories for files/dirs.

	// Create all symlinks.
	for dst, src := range merge {
		err := compileopts.Symlink(src, filepath.Join(tmpgoroot, dst))
		if err != nil {
			return "", err
		}
	}

	// Rename the new merged goroot into place.
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
func listGorootMergeLinks(goroot, tinygoroot string, overrides map[string]bool) (map[string]string, error) {
	goSrc := filepath.Join(goroot, "src")
	tinygoSrc := filepath.Join(tinygoroot, "src")
	tinygoRootBundle := filepath.Join(tinygoroot, "bundle")
	merges := make(map[string]string)

	tinygoBundleEntries, err := ioutil.ReadDir(tinygoRootBundle)
	if err != nil {
		return nil, err
	}
	for _, be := range tinygoBundleEntries {
		bundleName := be.Name()
		tinygoBundle := filepath.Join(tinygoRootBundle, bundleName)
		for _, dir := range []string{"bin", "lib"} {
			tinygoBundleSubdir := filepath.Join(tinygoBundle, dir)
			if _, err := os.Stat(tinygoBundleSubdir); err != nil {
				// TODO: handle non-existing Vs. EPERM
				continue
			}
			goEntries, err := ioutil.ReadDir(tinygoBundleSubdir)
			if err != nil {
				return nil, err
			}
			for _, e := range goEntries {
				merges[filepath.Join(dir, e.Name())] = filepath.Join(tinygoBundleSubdir, e.Name())
			}
		}
		tinygoBundleSrc := filepath.Join(tinygoBundle, "src")
		if _, err := os.Stat(tinygoBundleSrc); err != nil {
			// TODO: handle non-existing Vs. EPERM
			continue
		}
		err := filepath.Walk(tinygoBundleSrc, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				merges[path[len(tinygoBundle)+1:]] = path
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	for dir, merge := range overrides {
		if !merge {
			// Use the TinyGo version.
			// merges[filepath.Join("src", dir)] = filepath.Join(tinygoSrc, dir)
			walkPath := filepath.Join(tinygoSrc, dir)
			err := filepath.Walk(walkPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() {
					merges[path[len(tinygoroot)+1:]] = path
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
			continue
		}

		// Add files from TinyGo.
		tinygoDir := filepath.Join(tinygoSrc, dir)
		tinygoEntries, err := ioutil.ReadDir(tinygoDir)
		if err != nil {
			return nil, err
		}
		var hasTinyGoFiles bool
		for _, e := range tinygoEntries {
			if e.IsDir() {
				continue
			}

			// Link this file.
			name := e.Name()
			merges[filepath.Join("src", dir, name)] = filepath.Join(tinygoDir, name)

			hasTinyGoFiles = true
		}

		// Add all directories from $GOROOT that are not part of the TinyGo
		// overrides.
		goDir := filepath.Join(goSrc, dir)
		goEntries, err := ioutil.ReadDir(goDir)
		if err != nil {
			return nil, err
		}
		for _, e := range goEntries {
			if hasTinyGoFiles && !e.IsDir() {
				// Only merge files from Go if TinyGo does not have any files.
				// Otherwise we'd end up with a weird mix from both Go
				// implementations.
				continue
			}

			name := e.Name()
			if _, ok := overrides[path.Join(dir, name)+"/"]; ok {
				// This entry is overridden by TinyGo.
				// It has/will be merged elsewhere.
				continue
			}

			// Add a link to this entry
			merges[filepath.Join("src", dir, name)] = filepath.Join(goDir, name)
		}
	}

	// Merge the special directories from tinygo root and goroot.
	for _, topdir := range []string{tinygoroot, goroot} {
		for _, dir := range []string{"bin", "lib"} {
			goDir := filepath.Join(topdir, dir)
			goEntries, err := ioutil.ReadDir(goDir)
			if err != nil {
				return nil, err
			}
			for _, e := range goEntries {
				merges[filepath.Join(dir, e.Name())] = filepath.Join(goDir, e.Name())
			}
		}
	}
	merges["pkg"] = filepath.Join(goroot, "pkg")
	merges["targets"] = filepath.Join(goenv.Get("GOCACHE"), "targets")
	// // Merge the special directories from goroot.
	// for _, dir := range []string{"bin", "lib", "pkg"} {
	// 	merges[dir] = filepath.Join(goroot, dir)
	// }

	return merges, nil
}

// needsSyscallPackage returns whether the syscall package should be overriden
// with the TinyGo version. This is the case on some targets.
func needsSyscallPackage(buildTags []string) bool {
	for _, tag := range buildTags {
		if tag == "baremetal" || tag == "darwin" || tag == "nintendoswitch" || tag == "tinygo.wasm" {
			return true
		}
	}
	return false
}

// The boolean indicates whether to merge the subdirs. True means merge, false
// means use the TinyGo version.
func pathsToOverride(goMinor int, needsSyscallPackage bool) map[string]bool {
	paths := map[string]bool{
		"":                      true,
		"crypto/":               true,
		"crypto/rand/":          false,
		"device/":               false,
		"examples/":             false,
		"internal/":             true,
		"internal/fuzz/":        false,
		"internal/bytealg/":     false,
		"internal/reflectlite/": false,
		"internal/task/":        false,
		"machine/":              false,
		"net/":                  true,
		"os/":                   true,
		"reflect/":              false,
		"runtime/":              false,
		"sync/":                 true,
		"testing/":              true,
	}

	if goMinor >= 19 {
		paths["crypto/internal/"] = true
		paths["crypto/internal/boring/"] = true
		paths["crypto/internal/boring/sig/"] = false
	}

	if needsSyscallPackage {
		paths["syscall/"] = true // include syscall/js
	}
	return paths
}
