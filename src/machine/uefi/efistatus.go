package uefi

const (
	uintnSize = 32 << (^uintptr(0) >> 63)
	errorMask = 1 << uintptr(uintnSize-1)
)

const (
	EFI_SUCCESS           EFI_STATUS = 0
	EFI_INVALID_PARAMETER EFI_STATUS = errorMask | 2
	EFI_UNSUPPORTED       EFI_STATUS = errorMask | 3
	EFI_NOT_READY         EFI_STATUS = errorMask | 6
	EFI_OUT_OF_RESOURCES  EFI_STATUS = errorMask | 9
	EFI_ABORTED           EFI_STATUS = errorMask | 21
)
