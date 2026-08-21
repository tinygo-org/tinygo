package main

var libc_test_trampoline_addr uintptr
var libc_ioctl_trampoline_addr uintptr

//go:cgo_import_dynamic libc_test remote$INODE64 "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_ioctl ioctl "/usr/lib/libSystem.B.dylib"

func loadImportedFunctionAddress() uintptr {
	return libc_test_trampoline_addr
}

func loadImportedIoctlAddress() uintptr {
	return libc_ioctl_trampoline_addr
}
