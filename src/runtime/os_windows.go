package runtime

import "unsafe"

const GOOS = "windows"

//export GetModuleHandleExA
func _GetModuleHandleExA(dwFlags uint32, lpModuleName unsafe.Pointer, phModule **exeHeader) bool

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

	findGlobalsForPE(found)
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
