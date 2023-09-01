package runtime

import "unsafe"

const GOOS = "windows"

//export GetModuleHandleExA
func _GetModuleHandleExA(dwFlags uint32, lpModuleName unsafe.Pointer, phModule **exeHeader) bool

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

// Mark global variables.
// Unfortunately, the linker doesn't provide symbols for the start and end of
// the data/bss sections. Therefore these addresses need to be determined at
// runtime. This might seem complex and it kind of is, but it only compiles to
// around 160 bytes of amd64 instructions.
// Most of this function is based on the documentation in
// https://docs.microsoft.com/en-us/windows/win32/debug/pe-format.
func findGlobals(found func(start, end uintptr)) {
	// Constants used in this function.
	const (
		// https://docs.microsoft.com/en-us/windows/win32/api/libloaderapi/nf-libloaderapi-getmodulehandleexa
		GET_MODULE_HANDLE_EX_FLAG_UNCHANGED_REFCOUNT = 0x00000002

		// https://docs.microsoft.com/en-us/windows/win32/debug/pe-format
		IMAGE_SCN_MEM_WRITE = 0x80000000
	)

	if module == nil {
		// Obtain a handle to the currently executing image. What we're getting
		// here is really just __ImageBase, but it's probably better to obtain
		// it using GetModuleHandle to account for ASLR etc.
		result := _GetModuleHandleExA(GET_MODULE_HANDLE_EX_FLAG_UNCHANGED_REFCOUNT, nil, &module)
		if gcAsserts && (!result || module.signature != 0x5A4D) { // 0x4D5A is "MZ"
			runtimePanic("cannot get module handle")
		}
	}

	// Find the PE header at offset 0x3C.
	pe := (*peHeader)(unsafe.Add(unsafe.Pointer(module), module.peHeader))
	if gcAsserts && pe.magic != 0x00004550 { // 0x4550 is "PE"
		runtimePanic("cannot find PE header")
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

type systeminfo struct {
	anon0                       [4]byte
	dwpagesize                  uint32
	lpminimumapplicationaddress *byte
	lpmaximumapplicationaddress *byte
	dwactiveprocessormask       uintptr
	dwnumberofprocessors        uint32
	dwprocessortype             uint32
	dwallocationgranularity     uint32
	wprocessorlevel             uint16
	wprocessorrevision          uint16
}

//export GetSystemInfo
func _GetSystemInfo(lpSystemInfo unsafe.Pointer)

//go:linkname syscall_Getpagesize syscall.Getpagesize
func syscall_Getpagesize() int {
	var info systeminfo
	_GetSystemInfo(unsafe.Pointer(&info))
	return int(info.dwpagesize)
}
