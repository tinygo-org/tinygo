package uefi

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
