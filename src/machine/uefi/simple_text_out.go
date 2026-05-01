package uefi

import "unsafe"

type EFI_SIMPLE_TEXT_OUTPUT_MODE struct {
	MaxMode       int32
	Mode          int32
	Attribute     int32
	CursorColumn  int32
	CursorRow     int32
	CursorVisible BOOLEAN
}

type EFI_SIMPLE_TEXT_OUTPUT_PROTOCOL struct {
	reset             uintptr
	outputString      uintptr
	testString        uintptr
	queryMode         uintptr
	setMode           uintptr
	setAttribute      uintptr
	clearScreen       uintptr
	setCursorPosition uintptr
	enableCursor      uintptr
	Mode              *EFI_SIMPLE_TEXT_OUTPUT_MODE
}

func (p *EFI_SIMPLE_TEXT_OUTPUT_PROTOCOL) OutputString(s *CHAR16) EFI_STATUS {
	return UefiCall2(p.outputString, uintptr(unsafe.Pointer(p)), uintptr(unsafe.Pointer(s)))
}
