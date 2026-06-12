package builder

import (
	"errors"
	"math/rand"
	"os"
	"syscall"
	"time"
)

const robustRenameTimeout = 2 * time.Second

func robustRename(oldpath, newpath string) error {
	var bestErr error
	start := time.Now()
	nextSleep := time.Millisecond
	for {
		err := os.Rename(oldpath, newpath)
		if err == nil || !isEphemeralRenameError(err) {
			return err
		}
		if bestErr == nil {
			bestErr = err
		}
		if d := time.Since(start) + nextSleep; d >= robustRenameTimeout {
			return bestErr
		}
		time.Sleep(nextSleep)
		nextSleep += time.Duration(rand.Int63n(int64(nextSleep)))
	}
}

func isEphemeralRenameError(err error) bool {
	var errno syscall.Errno
	if errors.As(err, &errno) {
		switch errno {
		case syscall.Errno(2), // ERROR_FILE_NOT_FOUND
			syscall.Errno(5),  // ERROR_ACCESS_DENIED
			syscall.Errno(32): // ERROR_SHARING_VIOLATION
			return true
		}
	}
	return false
}
