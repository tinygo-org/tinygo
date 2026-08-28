//go:build scheduler.none && uefi

package task

//go:export tinygo_task_exit
func taskExit() {
	runtimePanic("scheduler is disabled")
}
