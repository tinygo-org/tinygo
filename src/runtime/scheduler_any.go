// +build !scheduler.none

package runtime

import "internal/task"

// Pause the current task for a given time.
//go:linkname sleep time.Sleep
func sleep(duration int64) {
	addSleepTask(task.Current(), duration)
	task.Pause()
}
