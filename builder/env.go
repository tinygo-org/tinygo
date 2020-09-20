package builder

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
)

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
			dirs, err := ioutil.ReadDir(clangVersionRoot)
			if err != nil {
				// Example include path:
				//     /usr/lib/llvm-9/lib/clang/9.0.1/include/
				clangVersionRoot = filepath.Join(llvmRoot, "lib", "clang")
				dirs, err = ioutil.ReadDir(clangVersionRoot)
				if err != nil {
					// Unexpected.
					continue
				}
			}
			dirnames := make([]string, len(dirs))
			for i, d := range dirs {
				dirnames[i] = d.Name()
			}
			sort.Strings(dirnames)
			// Check for the highest version first.
			for i := len(dirnames) - 1; i >= 0; i-- {
				path := filepath.Join(clangVersionRoot, dirnames[i], "include")
				_, err := os.Stat(filepath.Join(path, "stdint.h"))
				if err == nil {
					return path
				}
			}
		}
	}

	// Could not find it.
	return ""
}
