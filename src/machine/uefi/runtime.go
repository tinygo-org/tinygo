package uefi

import "unsafe"

var imageHandle EFI_HANDLE
var systemTable *EFI_SYSTEM_TABLE

func Init(handle uintptr, table uintptr) {
	imageHandle = EFI_HANDLE(handle)
	systemTable = (*EFI_SYSTEM_TABLE)(unsafe.Pointer(table))
}

func ST() *EFI_SYSTEM_TABLE {
	return systemTable
}

func BS() *EFI_BOOT_SERVICES {
	return systemTable.BootServices
}

func GetImageHandle() EFI_HANDLE {
	return imageHandle
}
