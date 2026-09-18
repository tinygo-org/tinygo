package main

var libc_misplaced_trampoline_addr uintptr //go:cgo_import_dynamic libc_misplaced misplaced_remote "/usr/lib/libSystem.B.dylib"
var libc_indented_trampoline_addr uintptr
var libc_brace_trampoline_addr uintptr
var libc_collide_trampoline_addr uintptr

//go:cgo_import_dynamic libc_collide collide_remote "/usr/lib/libSystem.B.dylib"

//go:extern collide_remote
var collideTarget uintptr

// ERROR: libc_misplaced_trampoline_addr has no //go:cgo_import_dynamic directive
func loadMisplacedAddress() uintptr {
	return libc_misplaced_trampoline_addr
}

// ERROR: libc_indented_trampoline_addr has no //go:cgo_import_dynamic directive
func loadIndentedAddress() uintptr {
	//go:cgo_import_dynamic libc_indented indented_remote "/usr/lib/libSystem.B.dylib"
	return libc_indented_trampoline_addr
}

// ERROR: libc_brace_trampoline_addr has no //go:cgo_import_dynamic directive
func loadBraceAddress() uintptr { //go:cgo_import_dynamic libc_brace brace_remote "/usr/lib/libSystem.B.dylib"
	return libc_brace_trampoline_addr
}

// ERROR: cgo_import_dynamic remote symbol collide_remote is already a global variable
func loadCollideAddress() uintptr {
	collideTarget = 1
	return libc_collide_trampoline_addr
}
