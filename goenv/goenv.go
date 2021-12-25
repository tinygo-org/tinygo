// Package goenv returns environment variables that are used in various parts of
// the compiler. You can query it manually with the `tinygo env` subcommand.
package goenv

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// Keys is a slice of all available environment variable keys.
var Keys = []string{
	"GO",
	"GOOS",
	"GOARCH",
	"GOROOT",
	"GOPATH",
	"GOCACHE",
	"CGO_ENABLED",
	"TINYGOROOT",
}

func init() {
	if Get("GOARCH") == "arm" {
		Keys = append(Keys, "GOARM")
	}
}

// TINYGOROOT is the path to the final location for checking tinygo files. If
// unset (by a -X ldflag), then sourceDir() will fallback to the original build
// directory.
var TINYGOROOT string

// Get returns a single environment variable, possibly calculating it on-demand.
// The empty string is returned for unknown environment variables.
func Get(name string) string {
	entry := cacheLookup(name)

	entry.once.Do(func() {
		entry.val = get(name)
	})

	return entry.val
}

func cacheLookup(name string) *envCacheEntry {
	envCache.mu.Lock()
	defer envCache.mu.Unlock()

	entry, ok := envCache.entries[name]
	if ok {
		return entry
	}

	entry = &envCacheEntry{}
	envCache.entries[name] = entry
	return entry
}

var envCache = struct {
	mu      sync.Mutex
	entries map[string]*envCacheEntry
}{entries: make(map[string]*envCacheEntry)}

type envCacheEntry struct {
	once sync.Once
	val  string
}

func get(name string) string {
	switch name {
	case "GO":
		return findGo()
	case "GOOS":
		if goos := os.Getenv("GOOS"); goos != "" {
			return goos
		}
		return runtime.GOOS
	case "GOARCH":
		if goarch := os.Getenv("GOARCH"); goarch != "" {
			return goarch
		}
		return runtime.GOARCH
	case "GOARM":
		if goarm := os.Getenv("GOARM"); goarm != "" {
			return goarm
		}
		if goos := Get("GOOS"); goos == "windows" || goos == "android" {
			// Assume Windows and Android are running on modern CPU cores.
			// This matches upstream Go.
			return "7"
		}
		// Default to ARMv6 on other devices.
		// The difference between ARMv5 and ARMv6 is big, much bigger than the
		// difference between ARMv6 and ARMv7. ARMv6 binaries are much smaller,
		// especially when floating point instructions are involved.
		return "6"
	case "GOROOT", "GOPATH":
		return envFromGo(name)
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
	case "WASMOPT":
		if path := os.Getenv("WASMOPT"); path != "" {
			err := wasmOptCheckVersion(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "cannot use %q as wasm-opt (from WASMOPT environment variable): %s", path, err.Error())
				os.Exit(1)
			}

			return path
		}

		return findWasmOpt()
	default:
		return ""
	}
}

// Find wasm-opt, or exit with an error.
func findWasmOpt() string {
	tinygoroot := sourceDir()
	searchPaths := []string{
		tinygoroot + "/bin/wasm-opt",
		tinygoroot + "/build/wasm-opt",
	}

	var paths []string
	for _, path := range searchPaths {
		if runtime.GOOS == "windows" {
			path += ".exe"
		}

		_, err := os.Stat(path)
		if err != nil && os.IsNotExist(err) {
			continue
		}

		paths = append(paths, path)
	}

	if path, err := exec.LookPath("wasm-opt"); err == nil {
		paths = append(paths, path)
	}

	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, "error: could not find wasm-opt, set the WASMOPT environment variable to override")
		os.Exit(1)
	}

	errs := make([]error, len(paths))
	for i, path := range paths {
		err := wasmOptCheckVersion(path)
		if err == nil {
			return path
		}

		errs[i] = err
	}
	fmt.Fprintln(os.Stderr, "no usable wasm-opt found, update or run \"make binaryen\"")
	for i, path := range paths {
		fmt.Fprintf(os.Stderr, "\t%s: %s\n", path, errs[i].Error())
	}
	os.Exit(1)
	panic("unreachable")
}

// wasmOptCheckVersion checks if a copy of wasm-opt is usable.
func wasmOptCheckVersion(path string) error {
	cmd := exec.Command(path, "--version")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}

	str := buf.String()
	if strings.Contains(str, "(") {
		// The git tag may be placed in parentheses after the main version string.
		str = strings.Split(str, "(")[0]
	}

	str = strings.TrimSpace(str)
	var ver uint
	_, err = fmt.Sscanf(str, "wasm-opt version %d", &ver)
	if err != nil || ver < 102 {
		return errors.New("incompatible wasm-opt (need 102 or newer)")
	}

	return nil
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

// envFromGo gets the value of a variable by querying Go.
func envFromGo(name string) string {
	goCmd := os.Getenv("GO")
	if goCmd == "" {
		goCmd = "go"
	}
	cmd := exec.Command(goCmd, "env", name)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(out.String())
}

// findGo searches for the Go command to run.
func findGo() string {
	if goCmd := os.Getenv("GO"); goCmd != "" {
		// If this was `go install`ed, then it will try to override GOROOT internally.
		// Find the underlying binary.
		if goroot := Get("GOROOT"); goroot != "" {
			path := filepath.Join(goroot, "bin", "go")
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}

		return goCmd
	}

	return "go"
}
