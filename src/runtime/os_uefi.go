//go:build uefi

package runtime

import (
	"machine/uefi"
	"unsafe"
)

// loadedImageBase is a cache for the ImageBase so we don't risk an allocation while
// running GC. Prevents crashes with gc.precise and scheduler.tasks.
var loadedImageBase uintptr

func init() {
	var img *uefi.EFI_LOADED_IMAGE_PROTOCOL
	if uefi.BS().HandleProtocol(
		uefi.GetImageHandle(),
		&uefi.EFI_LOADED_IMAGE_GUID,
		unsafe.Pointer(&img),
	) == uefi.EFI_SUCCESS {
		loadedImageBase = uintptr(unsafe.Pointer(img.ImageBase))
	}
	module = (*exeHeader)(unsafe.Pointer(loadedImageBase))
}

// Mark global variables.
// Unfortunately, the linker doesn't provide symbols for the start and end of
// the data/bss sections. Therefore these addresses need to be determined at
// runtime. This might seem complex and it kind of is, but it only compiles to
// around 160 bytes of amd64 instructions.
// Most of this function is based on the documentation in
// https://docs.microsoft.com/en-us/windows/win32/debug/pe-format.
func findGlobals(found func(start, end uintptr)) {
	if loadedImageBase == 0 {
		return // header not available; skip globals
	}

	findGlobalsForPE(found)
}
