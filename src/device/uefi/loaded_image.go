package uefi

import "unsafe"

var EFI_LOADED_IMAGE_PROTOCOL_GUID = EFI_GUID{
	0x5B1B31A1, 0x9562, 0x11D2,
	[8]byte{0x8E, 0x3F, 0x00, 0xA0, 0xC9, 0x69, 0x72, 0x3B},
}

type EFI_LOADED_IMAGE_PROTOCOL struct {
	Revision        uint32
	ParentHandle    EFI_HANDLE
	SystemTable     *EFI_SYSTEM_TABLE
	DeviceHandle    EFI_HANDLE
	FilePath        *EFI_DEVICE_PATH_PROTOCOL
	Reserved        *VOID
	LoadOptionsSize uint32
	LoadOptions     *VOID
	ImageBase       *VOID
	ImageSize       uint64
	ImageCodeType   EFI_MEMORY_TYPE
	ImageDataType   EFI_MEMORY_TYPE
	unload          uintptr
}

func GetLoadedImageProtocol() (*EFI_LOADED_IMAGE_PROTOCOL, EFI_STATUS) {
	var lip *EFI_LOADED_IMAGE_PROTOCOL
	status := BS().HandleProtocol(
		GetImageHandle(),
		&EFI_LOADED_IMAGE_PROTOCOL_GUID,
		unsafe.Pointer(&lip),
	)
	if status != EFI_SUCCESS {
		return nil, status
	}
	return lip, EFI_SUCCESS
}
