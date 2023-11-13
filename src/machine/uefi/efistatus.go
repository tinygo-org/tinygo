package uefi

const (
	uintnSize = 32 << (^uintptr(0) >> 63) // 32 or 64nt
	errorMask = 1 << uintptr(uintnSize-1)

	EFI_SUCCESS            EFI_STATUS = 0
	EFI_LOAD_ERROR         EFI_STATUS = errorMask | 1
	EFI_INVALID_PARAMETER  EFI_STATUS = errorMask | 2
	EFI_UNSUPPORTED        EFI_STATUS = errorMask | 3
	EFI_BAD_BUFFER_SIZE    EFI_STATUS = errorMask | 4
	EFI_BUFFER_TOO_SMALL   EFI_STATUS = errorMask | 5
	EFI_NOT_READY          EFI_STATUS = errorMask | 6
	EFI_DEVICE_ERROR       EFI_STATUS = errorMask | 7
	EFI_WRITE_PROTECTED    EFI_STATUS = errorMask | 8
	EFI_OUT_OF_RESOURCES   EFI_STATUS = errorMask | 9
	EFI_NOT_FOUND          EFI_STATUS = errorMask | 14
	EFI_ABORTED            EFI_STATUS = errorMask | 21
	EFI_SECURITY_VIOLATION EFI_STATUS = errorMask | 26
)
