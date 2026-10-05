//go:build wasip2

package syscall

import "internal/wasi/filesystem/v0.2.12/preopens"

func firstPreopenDirectory() string {
	dirs := preopens.GetDirectories().Slice()
	var name string
	if len(dirs) != 0 {
		name = dirs[0].F1
	}
	for _, dir := range dirs {
		dir.F0.ResourceDrop()
	}
	return name
}
