package uefi

import "unsafe"

func booleanArg(v BOOLEAN) uintptr {
	if v {
		return 1
	}
	return 0
}

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

func (p *EFI_SIMPLE_TEXT_OUTPUT_PROTOCOL) Reset(extendedVerification BOOLEAN) EFI_STATUS {
	return UefiCall2(p.reset, uintptr(unsafe.Pointer(p)), booleanArg(extendedVerification))
}

func (p *EFI_SIMPLE_TEXT_OUTPUT_PROTOCOL) OutputString(s *CHAR16) EFI_STATUS {
	return UefiCall2(p.outputString, uintptr(unsafe.Pointer(p)), uintptr(unsafe.Pointer(s)))
}

func (p *EFI_SIMPLE_TEXT_OUTPUT_PROTOCOL) TestString(s *CHAR16) EFI_STATUS {
	return UefiCall2(p.testString, uintptr(unsafe.Pointer(p)), uintptr(unsafe.Pointer(s)))
}

func (p *EFI_SIMPLE_TEXT_OUTPUT_PROTOCOL) QueryMode(modeNumber UINTN, columns *UINTN, rows *UINTN) EFI_STATUS {
	return UefiCall4(
		p.queryMode,
		uintptr(unsafe.Pointer(p)),
		uintptr(modeNumber),
		uintptr(unsafe.Pointer(columns)),
		uintptr(unsafe.Pointer(rows)),
	)
}

func (p *EFI_SIMPLE_TEXT_OUTPUT_PROTOCOL) SetMode(modeNumber UINTN) EFI_STATUS {
	return UefiCall2(p.setMode, uintptr(unsafe.Pointer(p)), uintptr(modeNumber))
}

func (p *EFI_SIMPLE_TEXT_OUTPUT_PROTOCOL) SetAttribute(attribute UINTN) EFI_STATUS {
	return UefiCall2(p.setAttribute, uintptr(unsafe.Pointer(p)), uintptr(attribute))
}

func (p *EFI_SIMPLE_TEXT_OUTPUT_PROTOCOL) ClearScreen() EFI_STATUS {
	return UefiCall1(p.clearScreen, uintptr(unsafe.Pointer(p)))
}

func (p *EFI_SIMPLE_TEXT_OUTPUT_PROTOCOL) SetCursorPosition(column UINTN, row UINTN) EFI_STATUS {
	return UefiCall3(p.setCursorPosition, uintptr(unsafe.Pointer(p)), uintptr(column), uintptr(row))
}

func (p *EFI_SIMPLE_TEXT_OUTPUT_PROTOCOL) EnableCursor(visible BOOLEAN) EFI_STATUS {
	return UefiCall2(p.enableCursor, uintptr(unsafe.Pointer(p)), booleanArg(visible))
}
