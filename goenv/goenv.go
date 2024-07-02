// Package goenv returns environment variables that are used in various parts of
// the compiler. You can query it manually with the `tinygo env` subcommand.
package goenv

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"tinygo.org/x/go-llvm"
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

func init() {
	if Get("GOARCH") == "arm" {
		Keys = append(Keys, "GOARM")
	}
}

// Set to true if we're linking statically against LLVM.
var hasBuiltinTools = false

// TINYGOROOT is the path to the final location for checking tinygo files. If
// unset (by a -X ldflag), then sourceDir() will fallback to the original build
// directory.
var TINYGOROOT string

// If a particular Clang resource dir must always be used and TinyGo can't
// figure out the directory using heuristics, this global can be set using a
// linker flag.
// This is needed for Nix.
var clangResourceDir string

// Variables read from a `go env` command invocation.
var goEnvVars struct {
	GOPATH    string
	GOROOT    string
	GOVERSION string
}

var goEnvVarsOnce sync.Once
var goEnvVarsErr error // error returned from cmd.Run

// Make sure goEnvVars is fresh. This can be called multiple times, the first
// time will update all environment variables in goEnvVars.
func readGoEnvVars() error {
	goEnvVarsOnce.Do(func() {
		cmd := exec.Command("go", "env", "-json", "GOPATH", "GOROOT", "GOVERSION")
		output, err := cmd.Output()
		if err != nil {
			// Check for "command not found" error.
			if execErr, ok := err.(*exec.Error); ok {
				goEnvVarsErr = fmt.Errorf("could not find '%s' command: %w", execErr.Name, execErr.Err)
				return
			}
			// It's perhaps a bit ugly to handle this error here, but I couldn't
			// think of a better place further up in the call chain.
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() != 0 {
				if len(exitErr.Stderr) != 0 {
					// The 'go' command exited with an error message. Print that
					// message and exit, so we behave in a similar way.
					os.Stderr.Write(exitErr.Stderr)
					os.Exit(exitErr.ExitCode())
				}
			}
			// Other errors. Not sure whether there are any, but just in case.
			goEnvVarsErr = err
			return
		}
		err = json.Unmarshal(output, &goEnvVars)
		if err != nil {
			// This should never happen if we have a sane Go toolchain
			// installed.
			goEnvVarsErr = fmt.Errorf("unexpected error while unmarshalling `go env` output: %w", err)
		}
	})

	return goEnvVarsErr
}

// Get returns a single environment variable, possibly calculating it on-demand.
// The empty string is returned for unknown environment variables.
func Get(name string) string {
	switch name {
	case "GOOS":
		goos := os.Getenv("GOOS")
		if goos == "" {
			goos = runtime.GOOS
		}
		if goos == "android" {
			goos = "linux"
		}
		return goos
	case "GOARCH":
		if dir := os.Getenv("GOARCH"); dir != "" {
			return dir
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
	case "GOROOT":
		readGoEnvVars()
		return goEnvVars.GOROOT
	case "GOPATH":
		readGoEnvVars()
		return goEnvVars.GOPATH
	case "GOCACHE":
		// Get the cache directory, usually ~/.cache/tinygo
		dir, err := os.UserCacheDir()
		if err != nil {
			panic("could not find cache dir: " + err.Error())
		}
		return filepath.Join(dir, "tinygo")
	case "CGO_ENABLED":
		// Always enable CGo. It is required by a number of targets, including
		// macOS and the rp2040.
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
	case "WASMTOOLS":
		if path := os.Getenv("WASMTOOLS"); path != "" {
			return path
		}
		return "wasm-tools"
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
		if err != nil && errors.Is(err, fs.ErrNotExist) {
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

// ClangResourceDir returns the clang resource dir if available. This is the
// -resource-dir flag. If it isn't available, an empty string is returned and
// -resource-dir should be left unset.
// The libclang flag must be set if the resource dir is read for use by
// libclang.
// In that case, the resource dir is always returned (even when linking
// dynamically against LLVM) because libclang always needs this directory.
func ClangResourceDir(libclang bool) string {
	if clangResourceDir != "" {
		// The resource dir is forced to a particular value at build time.
		// This is needed on Nix for example, where Clang and libclang don't
		// know their own resource dir.
		// Also see:
		// https://discourse.nixos.org/t/why-is-the-clang-resource-dir-split-in-a-separate-package/34114
		return clangResourceDir
	}

	if !hasBuiltinTools && !libclang {
		// Using external tools, so the resource dir doesn't need to be
		// specified. Clang knows where to find it.
		return ""
	}

	// Check whether we're running from a TinyGo release directory.
	// This is the case for release binaries on GitHub.
	root := Get("TINYGOROOT")
	releaseHeaderDir := filepath.Join(root, "lib", "clang")
	if _, err := os.Stat(releaseHeaderDir); !errors.Is(err, fs.ErrNotExist) {
		return releaseHeaderDir
	}

	if hasBuiltinTools {
		// We are statically linked to LLVM.
		// Check whether we're running from the source directory.
		// This typically happens when TinyGo was built using `make` as part of
		// development.
		llvmMajor := strings.Split(llvm.Version, ".")[0]
		buildResourceDir := filepath.Join(root, "llvm-build", "lib", "clang", llvmMajor)
		if _, err := os.Stat(buildResourceDir); !errors.Is(err, fs.ErrNotExist) {
			return buildResourceDir
		}
	} else {
		// We use external tools, either when installed using `go install` or
		// when packaged in a Linux distribution (Linux distros typically prefer
		// dynamic linking).
		// Try to detect the system clang resources directory.
		resourceDir := findSystemClangResources(root)
		if resourceDir != "" {
			return resourceDir
		}
	}

	// Resource directory not found.
	return ""
}

// Find the Clang resource dir on this particular system.
// Return the empty string when they aren't found.
func findSystemClangResources(TINYGOROOT string) string {
	llvmMajor := strings.Split(llvm.Version, ".")[0]

	switch runtime.GOOS {
	case "linux", "android":
		// Header files are typically stored in /usr/lib/clang/<version>/include.
		// Tested on Fedora 39, Debian 12, and Arch Linux.
		path := filepath.Join("/usr/lib/clang", llvmMajor)
		_, err := os.Stat(filepath.Join(path, "include", "stdint.h"))
		if err == nil {
			return path
		}
	case "darwin":
		// This assumes a Homebrew installation, like in builder/commands.go.
		var prefix string
		switch runtime.GOARCH {
		case "amd64":
			prefix = "/usr/local/opt/llvm@" + llvmMajor
		case "arm64":
			prefix = "/opt/homebrew/opt/llvm@" + llvmMajor
		default:
			return "" // very unlikely for now
		}
		path := fmt.Sprintf("%s/lib/clang/%s", prefix, llvmMajor)
		_, err := os.Stat(path + "/include/stdint.h")
		if err == nil {
			return path
		}
	}

	// Could not find it.
	return ""
}
