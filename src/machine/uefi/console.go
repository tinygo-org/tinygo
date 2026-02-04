//go:build uefi

package uefi

import "unsafe"

// SimpleTextOutputProtocol provides text output services.
type SimpleTextOutputProtocol struct {
	Reset             uintptr
	OutputString      uintptr
	TestString        uintptr
	QueryMode         uintptr
	SetMode           uintptr
	SetAttribute      uintptr
	ClearScreen       uintptr
	SetCursorPosition uintptr
	EnableCursor      uintptr
	Mode              uintptr
}

// OutputString prints a UCS-2 string to the console.
func OutputString(str *uint16) {
	st := GetSystemTable()
	if st == nil || st.ConOut == nil || st.ConOut.OutputString == 0 {
		return
	}
	Call(
		st.ConOut.OutputString,
		uintptr(unsafe.Pointer(st.ConOut)),
		uintptr(unsafe.Pointer(str)),
	)
}
