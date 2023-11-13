//go:build uefi

package runtime

import (
	"machine/uefi"
	"unsafe"
)

// Mark global variables.
// Unfortunately, the linker doesn't provide symbols for the start and end of
// the data/bss sections. Therefore these addresses need to be determined at
// runtime. This might seem complex and it kind of is, but it only compiles to
// around 160 bytes of amd64 instructions.
// Most of this function is based on the documentation in
// https://docs.microsoft.com/en-us/windows/win32/debug/pe-format.
func findGlobals(found func(start, end uintptr)) {
	if module == nil {
		var loadedImage *uefi.EFI_LOADED_IMAGE_PROTOCOL

		status := uefi.BS().HandleProtocol(uefi.GetImageHandle(), &uefi.EFI_LOADED_IMAGE_GUID, unsafe.Pointer(&loadedImage))
		if status != uefi.EFI_SUCCESS {
			uefi.DebugPrint("EFI_LOADED_IMAGE_GUID failed", uint64(status))
			return
		}

		module = (*exeHeader)(unsafe.Pointer(loadedImage.ImageBase))
	}

	findGlobalsForPE(found)
}
