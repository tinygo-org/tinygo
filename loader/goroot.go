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
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
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

	// Determine the location of the cached GOROOT.
	version, err := goenv.GorootVersionString(goroot)
	if err != nil {
		return "", err
	}
	// This hash is really a cache key, that contains (hopefully) enough
	// information to make collisions unlikely during development.
	// By including the Go version and TinyGo version, cache collisions should
	// not happen outside of development.
	hash := sha512.New512_256()
	fmt.Fprintln(hash, goroot)
	fmt.Fprintln(hash, version)
	fmt.Fprintln(hash, goenv.Version)
	fmt.Fprintln(hash, tinygoroot)
	gorootsHash := hash.Sum(nil)
	gorootsHashHex := hex.EncodeToString(gorootsHash[:])
	cachedgorootName := "goroot-" + version + "-" + gorootsHashHex
	cachedgoroot := filepath.Join(goenv.Get("GOCACHE"), cachedgorootName)
	if needsSyscallPackage(config.BuildTags()) {
		cachedgoroot += "-syscall"
	}

	// Do not try to create the cached GOROOT in parallel, that's only a waste
	// of I/O bandwidth and thus speed. Instead, use a mutex to make sure only
	// one goroutine does it at a time.
	// This is not a way to ensure atomicity (a different TinyGo invocation
	// could be creating the same directory), but instead a way to avoid
	// creating it many times in parallel when running tests in parallel.
	gorootCreateMutex.Lock()
	defer gorootCreateMutex.Unlock()

	if _, err := os.Stat(cachedgoroot); err == nil {
		return cachedgoroot, nil
	}
	err = os.MkdirAll(goenv.Get("GOCACHE"), 0777)
	if err != nil {
		return "", err
	}
	tmpgoroot, err := ioutil.TempDir(goenv.Get("GOCACHE"), cachedgorootName+".tmp")
	if err != nil {
		return "", err
	}

	// Remove the temporary directory if it wasn't moved to the right place
	// (for example, when there was an error).
	defer os.RemoveAll(tmpgoroot)

	for _, name := range []string{"bin", "lib", "pkg"} {
		err = symlink(filepath.Join(goroot, name), filepath.Join(tmpgoroot, name))
		if err != nil {
			return "", err
		}
	}
	err = mergeDirectory(goroot, tinygoroot, tmpgoroot, "", pathsToOverride(needsSyscallPackage(config.BuildTags())))
	if err != nil {
		return "", err
	}
	err = os.Rename(tmpgoroot, cachedgoroot)
	if err != nil {
		if os.IsExist(err) {
			// Another invocation of TinyGo also seems to have created a GOROOT.
			// Use that one instead. Our new GOROOT will be automatically
			// deleted by the defer above.
			return cachedgoroot, nil
		}
		if runtime.GOOS == "windows" && os.IsPermission(err) {
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

// mergeDirectory merges two roots recursively. The tmpgoroot is the directory
// that will be created by this call by either symlinking the directory from
// goroot or tinygoroot, or by creating the directory and merging the contents.
func mergeDirectory(goroot, tinygoroot, tmpgoroot, importPath string, overrides map[string]bool) error {
	if mergeSubdirs, ok := overrides[importPath+"/"]; ok {
		if !mergeSubdirs {
			// This directory and all subdirectories should come from the TinyGo
			// root, so simply make a symlink.
			newname := filepath.Join(tmpgoroot, "src", importPath)
			oldname := filepath.Join(tinygoroot, "src", importPath)
			return symlink(oldname, newname)
		}

		// Merge subdirectories. Start by making the directory to merge.
		err := os.Mkdir(filepath.Join(tmpgoroot, "src", importPath), 0777)
		if err != nil {
			return err
		}

		// Symlink all files from TinyGo, and symlink directories from TinyGo
		// that need to be overridden.
		tinygoEntries, err := ioutil.ReadDir(filepath.Join(tinygoroot, "src", importPath))
		if err != nil {
			return err
		}
		hasTinyGoFiles := false
		for _, e := range tinygoEntries {
			if e.IsDir() {
				// A directory, so merge this thing.
				err := mergeDirectory(goroot, tinygoroot, tmpgoroot, path.Join(importPath, e.Name()), overrides)
				if err != nil {
					return err
				}
			} else {
				// A file, so symlink this.
				newname := filepath.Join(tmpgoroot, "src", importPath, e.Name())
				oldname := filepath.Join(tinygoroot, "src", importPath, e.Name())
				err := symlink(oldname, newname)
				if err != nil {
					return err
				}
				hasTinyGoFiles = true
			}
		}

		// Symlink all directories from $GOROOT that are not part of the TinyGo
		// overrides.
		gorootEntries, err := ioutil.ReadDir(filepath.Join(goroot, "src", importPath))
		if err != nil {
			return err
		}
		for _, e := range gorootEntries {
			if e.IsDir() {
				if _, ok := overrides[path.Join(importPath, e.Name())+"/"]; ok {
					// Already included above, so don't bother trying to create this
					// symlink.
					continue
				}
				newname := filepath.Join(tmpgoroot, "src", importPath, e.Name())
				oldname := filepath.Join(goroot, "src", importPath, e.Name())
				err := symlink(oldname, newname)
				if err != nil {
					return err
				}
			} else {
				// Only merge files from Go if TinyGo does not have any files.
				// Otherwise we'd end up with a weird mix from both Go
				// implementations.
				if !hasTinyGoFiles {
					newname := filepath.Join(tmpgoroot, "src", importPath, e.Name())
					oldname := filepath.Join(goroot, "src", importPath, e.Name())
					err := symlink(oldname, newname)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
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
func pathsToOverride(needsSyscallPackage bool) map[string]bool {
	paths := map[string]bool{
		"/":                     true,
		"crypto/":               true,
		"crypto/rand/":          false,
		"device/":               false,
		"examples/":             false,
		"internal/":             true,
		"internal/bytealg/":     false,
		"internal/reflectlite/": false,
		"internal/task/":        false,
		"internal/itoa/":        false, // TODO: Remove when we drop support for go 1.16
		"machine/":              false,
		"net/":                  true,
		"os/":                   true,
		"reflect/":              false,
		"runtime/":              false,
		"sync/":                 true,
		"testing/":              true,
	}
	if needsSyscallPackage {
		paths["syscall/"] = true // include syscall/js
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
