package builder

import (
	"errors"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"tinygo.org/x/go-llvm"
)

// getClangHeaderPath returns the path to the built-in Clang headers. It tries
// multiple locations, which should make it find the directory when installed in
// various ways.
func getClangHeaderPath(TINYGOROOT string) string {
	// Check whether we're running from the source directory.
	path := filepath.Join(TINYGOROOT, "llvm-project", "clang", "lib", "Headers")
	if _, err := os.Stat(path); !errors.Is(err, fs.ErrNotExist) {
		return path
	}

	// Check whether we're running from the installation directory.
	path = filepath.Join(TINYGOROOT, "lib", "clang", "include")
	if _, err := os.Stat(path); !errors.Is(err, fs.ErrNotExist) {
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
