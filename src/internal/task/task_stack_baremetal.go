//go:build scheduler.tasks && baremetal
// +build scheduler.tasks,baremetal

package task

//go:extern _stack_bottom
var _stack_bottom uintptr

func init() {
	mainTask.state.canaryPtr = &_stack_bottom
	*mainTask.state.canaryPtr = stackCanary
}
