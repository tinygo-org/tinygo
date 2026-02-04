//go:build windows || uefi

package runtime

import "unsafe"

// MS-DOS stub with PE header offset:
// https://docs.microsoft.com/en-us/windows/win32/debug/pe-format#ms-dos-stub-image-only
type dosHeader struct {
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

// Mark global variables.
// Unfortunately, the linker doesn't provide symbols for the start and end of
// the data/bss sections. Therefore these addresses need to be determined at
// runtime. This might seem complex and it kind of is, but it only compiles to
// around 160 bytes of amd64 instructions.
// Most of this function is based on the documentation in
// https://docs.microsoft.com/en-us/windows/win32/debug/pe-format.
func findGlobalsFromPE(dosHeader *dosHeader, found func(start, end uintptr)) {
	// Constants used in this function.
	const (
		// https://docs.microsoft.com/en-us/windows/win32/debug/pe-format
		IMAGE_SCN_MEM_WRITE = 0x80000000
	)

	// Find the PE header at offset 0x3C.
	pe := (*peHeader)(unsafe.Add(unsafe.Pointer(dosHeader), uintptr(dosHeader.peHeader)))
	if gcAsserts && pe.magic != 0x00004550 { // 0x4550 is "PE"
		runtimePanic("cannot find PE header")
	}

	// Iterate through sections.
	section := (*peSection)(unsafe.Pointer(uintptr(unsafe.Pointer(pe)) + uintptr(pe.sizeOfOptionalHeader) + unsafe.Sizeof(peHeader{})))
	for i := 0; i < int(pe.numberOfSections); i++ {
		if section.characteristics&IMAGE_SCN_MEM_WRITE != 0 {
			// Found a writable section. Scan the entire section for roots.
			start := uintptr(unsafe.Pointer(dosHeader)) + uintptr(section.virtualAddress)
			end := uintptr(unsafe.Pointer(dosHeader)) + uintptr(section.virtualAddress) + uintptr(section.virtualSize)
			found(start, end)
		}
		section = (*peSection)(unsafe.Add(unsafe.Pointer(section), unsafe.Sizeof(peSection{})))
	}
}
