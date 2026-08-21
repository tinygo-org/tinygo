package main

var libc_test_trampoline_addr uintptr
var libc_ioctl_trampoline_addr uintptr
var libc_open_trampoline_addr uintptr
var libc_openat_trampoline_addr uintptr
var libc_fcntl_trampoline_addr uintptr
var libc_nolib_trampoline_addr uintptr
var libc_self_trampoline_addr uintptr
var libc_badtype_trampoline_addr uint32

//go:cgo_import_dynamic libc_test remote$INODE64 "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_ioctl ioctl "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_open open "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_openat openat "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_fcntl fcntl "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_nolib remote_nolib
//go:cgo_import_dynamic libc_self
//go:cgo_import_dynamic libc_badtype bad_remote "/usr/lib/libSystem.B.dylib"

func loadImportedFunctionAddress() uintptr {
	return libc_test_trampoline_addr
}

func loadImportedIoctlAddress() uintptr {
	return libc_ioctl_trampoline_addr
}

func loadImportedOpenAddress() uintptr {
	return libc_open_trampoline_addr
}

func loadImportedOpenatAddress() uintptr {
	return libc_openat_trampoline_addr
}

func loadImportedFcntlAddress() uintptr {
	return libc_fcntl_trampoline_addr
}

func loadImportedNoLibraryAddress() uintptr {
	return libc_nolib_trampoline_addr
}

func loadImportedSelfAddress() uintptr {
	return libc_self_trampoline_addr
}

func loadImportedBadTypeAddress() uint32 {
	return libc_badtype_trampoline_addr
}
