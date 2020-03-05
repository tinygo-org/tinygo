package loader

// This file constructs a new temporary GOROOT directory by merging both the
// standard Go GOROOT and the GOROOT from TinyGo using symlinks.

import (
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/goenv"
)

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

	needsSyscallPackage := false
	for _, tag := range config.BuildTags() {
		if tag == "baremetal" || tag == "darwin" {
			needsSyscallPackage = true
		}
	}

	// Determine the location of the cached GOROOT.
	version, err := goenv.GorootVersionString(goroot)
	if err != nil {
		return "", err
	}
	gorootsHash := sha512.Sum512_256([]byte(goroot + "\x00" + tinygoroot))
	gorootsHashHex := hex.EncodeToString(gorootsHash[:])
	cachedgoroot := filepath.Join(goenv.Get("GOCACHE"), "goroot-"+version+"-"+gorootsHashHex)
	if needsSyscallPackage {
		cachedgoroot += "-syscall"
	}

	if _, err := os.Stat(cachedgoroot); err == nil {
		return cachedgoroot, nil
	}
	tmpgoroot := cachedgoroot + ".tmp" + strconv.Itoa(rand.Int())
	err = os.MkdirAll(tmpgoroot, 0777)
	if err != nil {
		return "", err
	}

	for _, name := range []string{"bin", "lib", "pkg"} {
		err = os.Symlink(filepath.Join(goroot, name), filepath.Join(tmpgoroot, name))
		if err != nil {
			return "", err
		}
	}
	err = mergeDirectory(goroot, tinygoroot, tmpgoroot, "", pathsToOverride(needsSyscallPackage))
	if err != nil {
		return "", err
	}
	err = os.Rename(tmpgoroot, cachedgoroot)
	if err != nil {
		if os.IsExist(err) {
			// Another invocation of TinyGo also seems to have created a GOROOT.
			// Use that one instead and delete ours.
			os.RemoveAll(tmpgoroot)
			return cachedgoroot, nil
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
			return os.Symlink(oldname, newname)
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
				err := os.Symlink(oldname, newname)
				if err != nil {
					return err
				}
			}
		}

		// Symlink all directories from $GOROOT that are not part of the TinyGo
		// overrides.
		gorootEntries, err := ioutil.ReadDir(filepath.Join(goroot, "src", importPath))
		if err != nil {
			return err
		}
		for _, e := range gorootEntries {
			if !e.IsDir() {
				// Don't merge in files from Go. Otherwise we'd end up with a
				// weird syscall package with files from both roots.
				continue
			}
			if _, ok := overrides[path.Join(importPath, e.Name())+"/"]; ok {
				// Already included above, so don't bother trying to create this
				// symlink.
				continue
			}
			newname := filepath.Join(tmpgoroot, "src", importPath, e.Name())
			oldname := filepath.Join(goroot, "src", importPath, e.Name())
			err := os.Symlink(oldname, newname)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// The boolean indicates whether to merge the subdirs. True means merge, false
// means use the TinyGo version.
func pathsToOverride(needsSyscallPackage bool) map[string]bool {
	paths := map[string]bool{
		"/":                     true,
		"device/":               false,
		"examples/":             false,
		"internal/":             true,
		"internal/reflectlite/": false,
		"internal/task/":        false,
		"machine/":              false,
		"os/":                   true,
		"reflect/":              false,
		"runtime/":              false,
		"sync/":                 true,
		"testing/":              false,
	}
	if needsSyscallPackage {
		paths["syscall/"] = true // include syscall/js
	}
	return paths
}
