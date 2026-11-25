//go:build windows

package syscall

import "io/fs"

func (e Errno) Is(target error) bool {
	if target != fs.ErrNotExist {
		return false
	}
	return e == 2 || // ERROR_FILE_NOT_FOUND
		e == 3 || // ERROR_PATH_NOT_FOUND
		e == 53 // ERROR_BAD_NETPATH
}
