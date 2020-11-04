package builder

import (
	"os"
	"path/filepath"
	"runtime"
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
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return path
	}

	// Check whether we're running from the installation directory.
	path = filepath.Join(TINYGOROOT, "lib", "clang", "include")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return path
	}

	var patterns []string
	if runtime.GOOS == "linux" {
		// It looks like we are built with a system-installed LLVM. Do a last
		// attempt: try to use Clang headers for the currently used Clang version.
		// The path used on Linux distributions is normally in this form:
		//     /usr/lib/clang/10/include
		llvmMajor := strings.Split(llvm.Version, ".")[0]
		patterns = append(patterns, "/usr/lib*/clang/"+llvmMajor+"/include")
	}

	for _, pattern := range patterns {
		includeDirs, _ := filepath.Glob(pattern)
		if len(includeDirs) == 0 {
			// Unexpected. Maybe the headers are not installed or the above path is
			// incorrect?
			continue
		}
		// Glob does not guarantee the output is sorted (although it currently
		// is).
		sort.Strings(includeDirs)
		// Pick the last in the list.
		// There should be only one (even with multiple LLVM versions
		// installed), but if there are multiple the higher one might be a later
		// version.
		// Skip the entry if it doesn't exist (for example, it's a broken
		// symlink).
		candidate := includeDirs[len(includeDirs)-1]
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	// Could not find it.
	return ""
}
