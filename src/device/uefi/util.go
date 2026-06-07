//go:build uefi

package uefi

import "unsafe"

var systemTable *EFI_SYSTEM_TABLE
var imageHandle uintptr

//go:nobounds
func Init(argImageHandle uintptr, argSystemTable uintptr) {
	systemTable = (*EFI_SYSTEM_TABLE)(unsafe.Pointer(argSystemTable))
	imageHandle = argImageHandle
}

func ST() *EFI_SYSTEM_TABLE {
	return systemTable
}

func BS() *EFI_BOOT_SERVICES {
	if systemTable == nil {
		return nil
	}
	return systemTable.BootServices
}

func GetImageHandle() EFI_HANDLE {
	return EFI_HANDLE(imageHandle)
}
