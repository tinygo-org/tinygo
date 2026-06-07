package uefi

import (
	"unicode/utf16"
	"unsafe"
)

// StringToCHAR16 converts a Go string to a UTF-16 code unit slice.
func StringToCHAR16(s string) []CHAR16 {
	if s == "" {
		return nil
	}

	encoded := utf16.Encode([]rune(s))
	out := make([]CHAR16, len(encoded))
	for i, r := range encoded {
		out[i] = CHAR16(r)
	}
	return out
}

// StringToCHAR16Z converts a Go string to a NUL-terminated UTF-16 code unit slice.
func StringToCHAR16Z(s string) []CHAR16 {
	out := StringToCHAR16(s)
	return append(out, 0)
}

// BytesToCHAR16 converts UTF-8 text bytes to a UTF-16 code unit slice.
func BytesToCHAR16(b []byte) []CHAR16 {
	return StringToCHAR16(string(b))
}

// BytesToCHAR16Z converts UTF-8 text bytes to a NUL-terminated UTF-16 code unit slice.
func BytesToCHAR16Z(b []byte) []CHAR16 {
	return StringToCHAR16Z(string(b))
}

// CHAR16ToString converts a UTF-16 code unit slice to a Go string.
func CHAR16ToString(input []CHAR16) string {
	if len(input) == 0 {
		return ""
	}

	units := make([]uint16, len(input))
	for i, c := range input {
		units[i] = uint16(c)
	}
	return string(utf16.Decode(units))
}

// CHAR16ToBytes converts a UTF-16 code unit slice to UTF-8 text bytes.
func CHAR16ToBytes(input []CHAR16) []byte {
	return []byte(CHAR16ToString(input))
}

// CHAR16PtrToString converts a NUL-terminated UTF-16 string pointer to a Go string.
func CHAR16PtrToString(input *CHAR16) string {
	if input == nil {
		return ""
	}

	ptr := uintptr(unsafe.Pointer(input))
	length := 0
	for *(*CHAR16)(unsafe.Pointer(ptr)) != 0 {
		length++
		ptr += 2
	}
	return CHAR16PtrLenToString(input, length)
}

// CHAR16PtrLenToString converts a UTF-16 string pointer with a known code unit count to a Go string.
func CHAR16PtrLenToString(input *CHAR16, length int) string {
	if input == nil || length <= 0 {
		return ""
	}
	return CHAR16ToString(unsafe.Slice(input, length))
}
