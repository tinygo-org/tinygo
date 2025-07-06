//go:build scheduler.none && uefi

package task

//go:export tinygo_task_exit
func task_exit() {
}
