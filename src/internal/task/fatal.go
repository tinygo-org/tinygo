package task

import _ "unsafe"

//go:linkname runtimeFatal runtime.runtimeFatal
func runtimeFatal(msg string)
