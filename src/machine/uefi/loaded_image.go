package uefi

import "unsafe"

var EFI_LOADED_IMAGE_PROTOCOL_GUID = EFI_GUID{
	0x5B1B31A1, 0x9562, 0x11d2,
	[...]byte{0x8E, 0x3F, 0x00, 0xA0, 0xC9, 0x69, 0x72, 0x3B}}

// EFI_LOADED_IMAGE_PROTOCOL
// Can be used on any image handle to obtain information about the loaded image.
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

// Unload
// Unloads an image.
// @param[in]  ImageHandle       Handle that identifies the image to be unloaded.
// @retval EFI_SUCCESS           The image has been unloaded.
// @retval EFI_INVALID_PARAMETER ImageHandle is not a valid image handle.
func (p *EFI_LOADED_IMAGE_PROTOCOL) Unload(ImageHandle EFI_HANDLE) EFI_STATUS {
	return UefiCall1(p.unload, uintptr(imageHandle))
}

func GetLoadedImageProtocol() (*EFI_LOADED_IMAGE_PROTOCOL, error) {
	var lip *EFI_LOADED_IMAGE_PROTOCOL
	status := BS().HandleProtocol(
		GetImageHandle(),
		&EFI_LOADED_IMAGE_PROTOCOL_GUID,
		unsafe.Pointer(&lip),
	)

	if status == EFI_SUCCESS {
		return lip, nil
	}

	return nil, StatusError(status)
}
