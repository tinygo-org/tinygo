package builder

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"tinygo.org/x/go-llvm"
)

// parseGorootVersion returns the major and minor version for a given Go version
// string of the form `goX.Y.Z`.
// Returns (0, 0) if the version cannot be determined.
func parseGorootVersion(version string) (int, int, error) {
	var (
		maj, min int
		pch      string
	)
	n, err := fmt.Sscanf(version, "go%d.%d%s", &maj, &min, &pch)
	if n == 2 && io.EOF == err {
		// Means there were no trailing characters (i.e., not an alpha/beta)
		err = nil
	}
	if nil != err {
		return 0, 0, fmt.Errorf("failed to parse version: %s", err)
	}
	return maj, min, nil
}

// getGorootVersion returns the major and minor version for a given GOROOT path.
// If the version cannot be determined, (0, 0) is returned.
func getGorootVersion(goroot string) (int, int, error) {
	const errPrefix = "could not parse Go version"
	s, err := GorootVersionString(goroot)
	if err != nil {
		return 0, 0, err
	}

	if "" == s {
		return 0, 0, fmt.Errorf("%s: version string is empty", errPrefix)
	}

	if strings.HasPrefix(s, "devel") {
		maj, min, err := getGorootApiVersion(goroot)
		if nil != err {
			return 0, 0, fmt.Errorf("%s: invalid GOROOT API version: %s", errPrefix, err)
		}
		return maj, min, nil
	}

	if !strings.HasPrefix(s, "go") {
		return 0, 0, fmt.Errorf("%s: [%s] version does not start with 'go' prefix", s, errPrefix)
	}

	parts := strings.Split(s[2:], ".")
	if len(parts) < 2 {
		return 0, 0, fmt.Errorf("%s: version has less than two parts", errPrefix)
	}

	return parseGorootVersion(s)
}

// getGorootApiVersion returns the major and minor version of the Go API files
// defined for a given GOROOT path.
// If the version cannot be determined, (0, 0) is returned.
func getGorootApiVersion(goroot string) (int, int, error) {
	info, err := ioutil.ReadDir(filepath.Join(goroot, "api"))
	if nil != err {
		return 0, 0, fmt.Errorf("could not read API feature directory: %s", err)
	}
	maj, min := -1, -1
	for _, f := range info {
		if !strings.HasPrefix(f.Name(), "go") || f.IsDir() {
			continue
		}
		vers := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
		part := strings.Split(vers[2:], ".")
		if len(part) < 2 {
			continue
		}
		vmaj, vmin, err := parseGorootVersion(vers)
		if nil != err {
			continue
		}
		if vmaj >= maj && vmin > min {
			maj, min = vmaj, vmin
		}
	}
	if maj < 0 || min < 0 {
		return 0, 0, errors.New("no valid API feature files")
	}
	return maj, min, nil
}

// GorootVersionString returns the version string as reported by the Go
// toolchain for the given GOROOT path. It is usually of the form `go1.x.y` but
// can have some variations (for beta releases, for example).
func GorootVersionString(goroot string) (string, error) {
	if data, err := ioutil.ReadFile(filepath.Join(
		goroot, "src", "runtime", "internal", "sys", "zversion.go")); err == nil {
		r := regexp.MustCompile("const TheVersion = `(.*)`")
		matches := r.FindSubmatch(data)
		if len(matches) != 2 {
			return "", errors.New("Invalid go version output:\n" + string(data))
		}
		return string(matches[1]), nil
	} else if data, err := ioutil.ReadFile(filepath.Join(goroot, "VERSION")); err == nil {
		return string(data), nil
	} else {
		return "", err
	}
}

// getClangHeaderPath returns the path to the built-in Clang headers. It tries
// multiple locations, which should make it find the directory when installed in
// various ways.
func getClangHeaderPath(TINYGOROOT string) string {
	// Check whether we're running from the source directory.
	path := filepath.Join(TINYGOROOT, "llvm-project", "clang", "lib", "Headers")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return path
	}

	// Check whether we're running from the installation directory.
	path = filepath.Join(TINYGOROOT, "lib", "clang", "include")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return path
	}

	// It looks like we are built with a system-installed LLVM. Do a last
	// attempt: try to use Clang headers relative to the clang binary.
	llvmMajor := strings.Split(llvm.Version, ".")[0]
	for _, cmdName := range commands["clang"] {
		binpath, err := exec.LookPath(cmdName)
		if err == nil {
			// This should be the command that will also be used by
			// execCommand. To avoid inconsistencies, make sure we use the
			// headers relative to this command.
			binpath, err = filepath.EvalSymlinks(binpath)
			if err != nil {
				// Unexpected.
				return ""
			}
			// Example executable:
			//     /usr/lib/llvm-9/bin/clang
			// Example include path:
			//     /usr/lib/llvm-9/lib64/clang/9.0.1/include/
			llvmRoot := filepath.Dir(filepath.Dir(binpath))
			clangVersionRoot := filepath.Join(llvmRoot, "lib64", "clang")
			dirs64, err64 := ioutil.ReadDir(clangVersionRoot)
			// Example include path:
			//     /usr/lib/llvm-9/lib/clang/9.0.1/include/
			clangVersionRoot = filepath.Join(llvmRoot, "lib", "clang")
			dirs32, err32 := ioutil.ReadDir(clangVersionRoot)
			if err64 != nil && err32 != nil {
				// Unexpected.
				continue
			}
			dirnames := make([]string, len(dirs64)+len(dirs32))
			dirCount := 0
			for _, d := range dirs32 {
				name := d.Name()
				if name == llvmMajor || strings.HasPrefix(name, llvmMajor+".") {
					dirnames[dirCount] = filepath.Join(llvmRoot, "lib", "clang", name)
					dirCount++
				}
			}
			for _, d := range dirs64 {
				name := d.Name()
				if name == llvmMajor || strings.HasPrefix(name, llvmMajor+".") {
					dirnames[dirCount] = filepath.Join(llvmRoot, "lib64", "clang", name)
					dirCount++
				}
			}
			sort.Strings(dirnames)
			// Check for the highest version first.
			for i := dirCount - 1; i >= 0; i-- {
				path := filepath.Join(dirnames[i], "include")
				_, err := os.Stat(filepath.Join(path, "stdint.h"))
				if err == nil {
					return path
				}
			}
		}
	}

	// On Arch Linux, the clang executable is stored in /usr/bin rather than being symlinked from there.
	// Search directly in /usr/lib for clang.
	if matches, err := filepath.Glob("/usr/lib/clang/" + llvmMajor + ".*.*"); err == nil {
		// Check for the highest version first.
		sort.Strings(matches)
		for i := len(matches) - 1; i >= 0; i-- {
			path := filepath.Join(matches[i], "include")
			_, err := os.Stat(filepath.Join(path, "stdint.h"))
			if err == nil {
				return path
			}
		}
	}

	// Could not find it.
	return ""
}
