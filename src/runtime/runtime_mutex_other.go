//go:build !scheduler.cores

package runtime

import "internal/task"

type timerLock = task.PMutex
type synctestLock = task.PMutex
