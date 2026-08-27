//go:build scheduler.none && uefi

package task

//export tinygo_task_exit
func taskExit() {
	runtimePanic("scheduler is disabled")
}
