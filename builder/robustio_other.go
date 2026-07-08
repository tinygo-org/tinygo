//go:build !windows

package builder

import "os"

func robustRename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}
