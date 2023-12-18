package task

//export tinygo_handoverTask
func handoverTask(fn uintptr, args uintptr)

//go:extern tinygo_callMachineEntry
var callMachineEntryFn [0]uint8
