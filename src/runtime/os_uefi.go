//go:build uefi

package runtime

import (
	"device/uefi"
	"unsafe"
)

// loadedImageBase caches the loaded image base so PE globals discovery can
// locate writable sections without relying on linker-provided symbols.
var loadedImageBase uintptr

func init() {
	lip, status := uefi.GetLoadedImageProtocol()
	if status != uefi.EFI_SUCCESS || lip == nil {
		return
	}

	loadedImageBase = uintptr(unsafe.Pointer(lip.ImageBase))
	module = (*exeHeader)(unsafe.Pointer(loadedImageBase))
}

func findGlobals(found func(start, end uintptr)) {
	if loadedImageBase == 0 {
		return
	}

	findGlobalsForPE(found)
}
