package uefi

type UINTN uintptr
type EFI_STATUS UINTN
type EFI_LBA uint64
type EFI_TPL UINTN
type EFI_HANDLE uintptr
type EFI_EVENT uintptr

type CHAR16 uint16
type BOOLEAN bool
type VOID any

type EFI_GUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}
