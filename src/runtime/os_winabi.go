//go:build windows || uefi

package runtime

import (
	"unsafe"
)

// MS-DOS stub with PE header offset:
// https://docs.microsoft.com/en-us/windows/win32/debug/pe-format#ms-dos-stub-image-only
type exeHeader struct {
	signature uint16
	_         [58]byte // skip DOS header
	peHeader  uint32   // at offset 0x3C
}

// COFF file header:
// https://docs.microsoft.com/en-us/windows/win32/debug/pe-format#file-headers
type peHeader struct {
	magic                uint32
	machine              uint16
	numberOfSections     uint16
	timeDateStamp        uint32
	pointerToSymbolTable uint32
	numberOfSymbols      uint32
	sizeOfOptionalHeader uint16
	characteristics      uint16
}

// COFF section header:
// https://docs.microsoft.com/en-us/windows/win32/debug/pe-format#section-table-section-headers
type peSection struct {
	name                 [8]byte
	virtualSize          uint32
	virtualAddress       uint32
	sizeOfRawData        uint32
	pointerToRawData     uint32
	pointerToRelocations uint32
	pointerToLinenumbers uint32
	numberOfRelocations  uint16
	numberOfLinenumbers  uint16
	characteristics      uint32
}

var module *exeHeader

func findGlobalsForPE(found func(start, end uintptr)) {
	// Constants used in this function.
	const (
		// https://docs.microsoft.com/en-us/windows/win32/debug/pe-format
		IMAGE_SCN_MEM_WRITE = 0x80000000
	)

	pe := (*peHeader)(unsafe.Add(unsafe.Pointer(module), module.peHeader))
	if pe.magic != 0x00004550 { // 0x4550 is "PE"
		return
	}

	// Iterate through sections.
	section := (*peSection)(unsafe.Pointer(uintptr(unsafe.Pointer(pe)) + uintptr(pe.sizeOfOptionalHeader) + unsafe.Sizeof(peHeader{})))
	for i := 0; i < int(pe.numberOfSections); i++ {
		if section.characteristics&IMAGE_SCN_MEM_WRITE != 0 {
			// Found a writable section. Scan the entire section for roots.
			start := uintptr(unsafe.Pointer(module)) + uintptr(section.virtualAddress)
			end := uintptr(unsafe.Pointer(module)) + uintptr(section.virtualAddress) + uintptr(section.virtualSize)
			found(start, end)
		}
		section = (*peSection)(unsafe.Add(unsafe.Pointer(section), unsafe.Sizeof(peSection{})))
	}
}
