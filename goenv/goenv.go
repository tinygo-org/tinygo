// Package goenv returns environment variables that are used in various parts of
// the compiler. You can query it manually with the `tinygo env` subcommand.
package goenv

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
)

// Keys is a slice of all available environment variable keys.
var Keys = []string{
	"GOOS",
	"GOARCH",
	"GOROOT",
	"GOPATH",
	"GOCACHE",
	"CGO_ENABLED",
	"TINYGOROOT",
}

// TINYGOROOT is the path to the final location for checking tinygo files. If
// unset (by a -X ldflag), then sourceDir() will fallback to the original build
// directory.
var TINYGOROOT string

// Get returns a single environment variable, possibly calculating it on-demand.
// The empty string is returned for unknown environment variables.
func Get(name string) string {
	switch name {
	case "GOOS":
		if dir := os.Getenv("GOOS"); dir != "" {
			return dir
		}
		return runtime.GOOS
	case "GOARCH":
		if dir := os.Getenv("GOARCH"); dir != "" {
			return dir
		}
		return runtime.GOARCH
	case "GOROOT":
		return getGoroot()
	case "GOPATH":
		if dir := os.Getenv("GOPATH"); dir != "" {
			return dir
		}

		// fallback
		home := getHomeDir()
		return filepath.Join(home, "go")
	case "GOCACHE":
		// Get the cache directory, usually ~/.cache/tinygo
		dir, err := os.UserCacheDir()
		if err != nil {
			panic("could not find cache dir: " + err.Error())
		}
		return filepath.Join(dir, "tinygo")
	case "CGO_ENABLED":
		val := os.Getenv("CGO_ENABLED")
		if val == "1" || val == "0" {
			return val
		}
		// Default to enabling CGo.
		return "1"
	case "TINYGOROOT":
		return sourceDir()
	default:
		return ""
	}
}

// Return the TINYGOROOT, or exit with an error.
func sourceDir() string {
	// Use $TINYGOROOT as root, if available.
	root := os.Getenv("TINYGOROOT")
	if root != "" {
		if !isSourceDir(root) {
			fmt.Fprintln(os.Stderr, "error: $TINYGOROOT was not set to the correct root")
			os.Exit(1)
		}
		return root
	}

	if TINYGOROOT != "" {
		if !isSourceDir(TINYGOROOT) {
			fmt.Fprintln(os.Stderr, "error: TINYGOROOT was not set to the correct root")
			os.Exit(1)
		}
		return TINYGOROOT
	}

	// Find root from executable path.
	path, err := os.Executable()
	if err != nil {
		// Very unlikely. Bail out if it happens.
		panic("could not get executable path: " + err.Error())
	}
	root = filepath.Dir(filepath.Dir(path))
	if isSourceDir(root) {
		return root
	}

	// Fallback: use the original directory from where it was built
	// https://stackoverflow.com/a/32163888/559350
	_, path, _, _ = runtime.Caller(0)
	root = filepath.Dir(filepath.Dir(path))
	if isSourceDir(root) {
		return root
	}

	fmt.Fprintln(os.Stderr, "error: could not autodetect root directory, set the TINYGOROOT environment variable to override")
	os.Exit(1)
	panic("unreachable")
}

// isSourceDir returns true if the directory looks like a TinyGo source directory.
func isSourceDir(root string) bool {
	_, err := os.Stat(filepath.Join(root, "src/runtime/internal/sys/zversion.go"))
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(root, "src/device/arm/arm.go"))
	return err == nil
}

func getHomeDir() string {
	u, err := user.Current()
	if err != nil {
		panic("cannot get current user: " + err.Error())
	}
	if u.HomeDir == "" {
		// This is very unlikely, so panic here.
		// Not the nicest solution, however.
		panic("could not find home directory")
	}
	return u.HomeDir
}

// getGoroot returns an appropriate GOROOT from various sources. If it can't be
// found, it returns an empty string.
func getGoroot() string {
	// An explicitly set GOROOT always has preference.
	goroot := os.Getenv("GOROOT")
	if goroot != "" {
		// If environmental GOROOT points to a TinyGo cache, use the physical path
		// the cache points into.
		if isCache, phyGoroot := isCachedGoroot(goroot); isCache {
			return phyGoroot
		}
		return goroot
	}

	// Check for the location of the 'go' binary and base GOROOT on that.
	binpath, err := exec.LookPath("go")
	if err == nil {
		binpath, err = filepath.EvalSymlinks(binpath)
		if err == nil {
			goroot := filepath.Dir(filepath.Dir(binpath))
			if isGoroot(goroot) {
				return goroot
			}
		}
	}

	// Check what GOROOT was at compile time.
	if isGoroot(runtime.GOROOT()) {
		return runtime.GOROOT()
	}

	// Check for some standard locations, as a last resort.
	var candidates []string
	switch runtime.GOOS {
	case "linux":
		candidates = []string{
			"/usr/local/go", // manually installed
			"/usr/lib/go",   // from the distribution
		}
	case "darwin":
		candidates = []string{
			"/usr/local/go",             // manually installed
			"/usr/local/opt/go/libexec", // from Homebrew
		}
	}

	for _, candidate := range candidates {
		if isGoroot(candidate) {
			return candidate
		}
	}

	// Can't find GOROOT...
	return ""
}

// isGoroot checks whether the given path looks like a GOROOT.
func isGoroot(goroot string) bool {
	_, err := os.Stat(filepath.Join(goroot, "src", "runtime", "internal", "sys", "zversion.go"))
	return err == nil
}

// isCachedGoroot checks whether the given path looks like a cached GOROOT.
// If it does appear to be a cached GOROOT, it returns true and the physical
// path to which the cache is linked. Otherwise, false and the empty string
// are returned.
func isCachedGoroot(goroot string) (bool, string) {

	// The physical GOROOT path to which the cache is linked.
	var phyGoroot string

	// Name and type of directory entries indicative of a cached GOROOT.
	// As these are encountered, they are deleted from the map. If, after
	// reading all entries in the given path, this map has zero remaining
	// elements, then we have a cached GOROOT.
	cacheEntry := map[string]os.FileMode{
		"src": os.ModeDir,
		"bin": os.ModeSymlink,
		"lib": os.ModeSymlink,
		"pkg": os.ModeSymlink,
	}

	info, err := ioutil.ReadDir(goroot)
	if err != nil {
		return false, ""
	}
	for _, f := range info {
		if mode, ok := cacheEntry[f.Name()]; ok {
			if mode != f.Mode().Type() {
				return false, ""
			}
			// Remove the verified entry
			delete(cacheEntry, f.Name())
			if mode&os.ModeSymlink == 0 {
				continue
			}
			// Entry is a symlink, get the physical GOROOT to which it points
			root, err := os.Readlink(filepath.Join(goroot, f.Name()))
			if err != nil {
				return false, ""
			}
			// Assume parent directory of the symlink destination is GOROOT
			root = filepath.Dir(root)
			if phyGoroot == "" {
				// Keep first symlink encountered
				phyGoroot = root
			} else if phyGoroot != root {
				// Destination GOROOT doesn't match all previously discovered
				return false, ""
			}
		}
	}

	if len(cacheEntry) == 0 {
		// All expected directory entries were found, and all symlinks were
		// pointing to subdirectories of the same exact parent directory.
		return true, phyGoroot
	}

	return false, ""
}
