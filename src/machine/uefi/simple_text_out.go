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

// EFI_SIMPLE_TEXT_OUTPUT_PROTOCOL
// The SIMPLE_TEXT_OUTPUT protocol is used to control text-based output devices.
// It is the minimum required protocol for any handle supplied as the ConsoleOut
// or StandardError device. In addition, the minimum supported text mode of such
// devices is at least 80 x 25 characters.
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

// Reset
// Reset the text output device hardware and optionally run diagnostics
// @param  This                 The protocol instance pointer.
// @param  ExtendedVerification Driver may perform more exhaustive verification
// .............................operation of the device during reset.
// @retval EFI_SUCCESS          The text output device was reset.
// @retval EFI_DEVICE_ERROR     The text output device is not functioning correctly and
// .............................could not be reset.
func (p *EFI_SIMPLE_TEXT_OUTPUT_PROTOCOL) Reset(ExtendedVerification BOOLEAN) EFI_STATUS {
	return UefiCall2(p.reset, uintptr(unsafe.Pointer(p)), convertBoolean(ExtendedVerification))
}

// OutputString
// Write a string to the output device.
// @param  This   The protocol instance pointer.
// @param  String The NULL-terminated string to be displayed on the output
// ...............device(s). All output devices must also support the Unicode
// ...............drawing character codes defined in this file.
// @retval EFI_SUCCESS             The string was output to the device.
// @retval EFI_DEVICE_ERROR        The device reported an error while attempting to output
// ................................the text.
// @retval EFI_UNSUPPORTED         The output device's mode is not currently in a
// ................................defined text mode.
// @retval EFI_WARN_UNKNOWN_GLYPH  This warning code indicates that some of the
// ................................characters in the string could not be
// ................................rendered and were skipped.
func (p *EFI_SIMPLE_TEXT_OUTPUT_PROTOCOL) OutputString(String *CHAR16) EFI_STATUS {
	return UefiCall2(p.outputString, uintptr(unsafe.Pointer(p)), uintptr(unsafe.Pointer(String)))
}

// TestString
// Verifies that all characters in a string can be output to the
// target device.
// @param  This   The protocol instance pointer.
// @param  String The NULL-terminated string to be examined for the output
// ...............device(s).
// @retval EFI_SUCCESS      The device(s) are capable of rendering the output string.
// @retval EFI_UNSUPPORTED  Some of the characters in the string cannot be
// .........................rendered by one or more of the output devices mapped
// .........................by the EFI handle.
func (p *EFI_SIMPLE_TEXT_OUTPUT_PROTOCOL) TestString(String *CHAR16) EFI_STATUS {
	return UefiCall2(p.testString, uintptr(unsafe.Pointer(p)), uintptr(unsafe.Pointer(String)))
}

// QueryMode
// Returns information for an available text mode that the output device(s)
// supports.
// @param  This       The protocol instance pointer.
// @param  ModeNumber The mode number to return information on.
// @param  Columns    Returns the geometry of the text output device for the
// ...................requested ModeNumber.
// @param  Rows       Returns the geometry of the text output device for the
// ...................requested ModeNumber.
// @retval EFI_SUCCESS      The requested mode information was returned.
// @retval EFI_DEVICE_ERROR The device had an error and could not complete the request.
// @retval EFI_UNSUPPORTED  The mode number was not valid.
func (p *EFI_SIMPLE_TEXT_OUTPUT_PROTOCOL) QueryMode(ModeNumber UINTN, Columns *UINTN, Rows *UINTN) EFI_STATUS {
	return UefiCall4(p.queryMode, uintptr(unsafe.Pointer(p)), uintptr(ModeNumber), uintptr(unsafe.Pointer(Columns)), uintptr(unsafe.Pointer(Rows)))
}

// SetMode
// Sets the output device(s) to a specified mode.
// @param  This       The protocol instance pointer.
// @param  ModeNumber The mode number to set.
// @retval EFI_SUCCESS      The requested text mode was set.
// @retval EFI_DEVICE_ERROR The device had an error and could not complete the request.
// @retval EFI_UNSUPPORTED  The mode number was not valid.
func (p *EFI_SIMPLE_TEXT_OUTPUT_PROTOCOL) SetMode(ModeNumber UINTN) EFI_STATUS {
	return UefiCall2(p.setMode, uintptr(unsafe.Pointer(p)), uintptr(ModeNumber))
}

// SetAttribute
// Sets the background and foreground colors for the OutputString () and
// ClearScreen () functions.
// @param  This      The protocol instance pointer.
// @param  Attribute The attribute to set. Bits 0..3 are the foreground color, and
// ..................bits 4..6 are the background color. All other bits are undefined
// ..................and must be zero. The valid Attributes are defined in this file.
// @retval EFI_SUCCESS       The attribute was set.
// @retval EFI_DEVICE_ERROR  The device had an error and could not complete the request.
// @retval EFI_UNSUPPORTED   The attribute requested is not defined.
func (p *EFI_SIMPLE_TEXT_OUTPUT_PROTOCOL) SetAttribute(Attribute UINTN) EFI_STATUS {
	return UefiCall2(p.setAttribute, uintptr(unsafe.Pointer(p)), uintptr(Attribute))
}

// ClearScreen
// Clears the output device(s) display to the currently selected background
// color.
// @param  This              The protocol instance pointer.
// @retval  EFI_SUCCESS      The operation completed successfully.
// @retval  EFI_DEVICE_ERROR The device had an error and could not complete the request.
// @retval  EFI_UNSUPPORTED  The output device is not in a valid text mode.
func (p *EFI_SIMPLE_TEXT_OUTPUT_PROTOCOL) ClearScreen() EFI_STATUS {
	return UefiCall1(p.clearScreen, uintptr(unsafe.Pointer(p)))
}

// SetCursorPosition
// Sets the current coordinates of the cursor position
// @param  This        The protocol instance pointer.
// @param  Column      The position to set the cursor to. Must be greater than or
// ....................equal to zero and less than the number of columns and rows
// ....................by QueryMode ().
// @param  Row         The position to set the cursor to. Must be greater than or
// ....................equal to zero and less than the number of columns and rows
// ....................by QueryMode ().
// @retval EFI_SUCCESS      The operation completed successfully.
// @retval EFI_DEVICE_ERROR The device had an error and could not complete the request.
// @retval EFI_UNSUPPORTED  The output device is not in a valid text mode, or the
// .........................cursor position is invalid for the current mode.
func (p *EFI_SIMPLE_TEXT_OUTPUT_PROTOCOL) SetCursorPosition(Column UINTN, Row UINTN) EFI_STATUS {
	return UefiCall3(p.setCursorPosition, uintptr(unsafe.Pointer(p)), uintptr(Column), uintptr(Row))
}

// EnableCursor
// Makes the cursor visible or invisible
// @param  This    The protocol instance pointer.
// @param  Visible If TRUE, the cursor is set to be visible. If FALSE, the cursor is
// ................set to be invisible.
// @retval EFI_SUCCESS      The operation completed successfully.
// @retval EFI_DEVICE_ERROR The device had an error and could not complete the
// .........................request, or the device does not support changing
// .........................the cursor mode.
// @retval EFI_UNSUPPORTED  The output device is not in a valid text mode.
func (p *EFI_SIMPLE_TEXT_OUTPUT_PROTOCOL) EnableCursor(Visible BOOLEAN) EFI_STATUS {
	return UefiCall2(p.enableCursor, uintptr(unsafe.Pointer(p)), convertBoolean(Visible))
}
