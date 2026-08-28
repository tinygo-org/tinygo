//go:build scheduler.none

package task

//go:export tinygo_task_exit
func taskExit() {
	runtimePanic("scheduler is disabled")
}
