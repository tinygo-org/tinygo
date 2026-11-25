//go:build !windows

package loader

import "path/filepath"

func analysisPath(name string) (string, error) {
	return filepath.EvalSymlinks(name)
}
