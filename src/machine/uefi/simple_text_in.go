package uefi

import "unsafe"

// EFI_INPUT_KEY
// The keystroke information for the key that was pressed.
type EFI_INPUT_KEY struct {
	ScanCode    uint16
	UnicodeChar CHAR16
}

// EFI_SIMPLE_TEXT_INPUT_PROTOCOL
// The EFI_SIMPLE_TEXT_INPUT_PROTOCOL is used on the ConsoleIn device.
// It is the minimum required protocol for ConsoleIn.
type EFI_SIMPLE_TEXT_INPUT_PROTOCOL struct {
	reset         uintptr
	readKeyStroke uintptr
	WaitForKey    EFI_EVENT
}

// Reset
// Reset the input device and optionally run diagnostics
// @param  This                 Protocol instance pointer.
// @param  ExtendedVerification Driver may perform diagnostics on reset.
// @retval EFI_SUCCESS          The device was reset.
// @retval EFI_DEVICE_ERROR     The device is not functioning properly and could not be reset.
func (p *EFI_SIMPLE_TEXT_INPUT_PROTOCOL) Reset(ExtendedVerification BOOLEAN) EFI_STATUS {
	return UefiCall2(p.reset, uintptr(unsafe.Pointer(p)), convertBoolean(ExtendedVerification))
}

// ReadKeyStroke
// Reads the next keystroke from the input device. The WaitForKey Event can
// be used to test for existence of a keystroke via WaitForEvent () call.
// @param  This  Protocol instance pointer.
// @param  Key   A pointer to a buffer that is filled in with the keystroke
// ..............information for the key that was pressed.
// @retval EFI_SUCCESS      The keystroke information was returned.
// @retval EFI_NOT_READY    There was no keystroke data available.
// @retval EFI_DEVICE_ERROR The keystroke information was not returned due to
// .........................hardware errors.
func (p *EFI_SIMPLE_TEXT_INPUT_PROTOCOL) ReadKeyStroke(Key *EFI_INPUT_KEY) EFI_STATUS {
	return UefiCall2(p.readKeyStroke, uintptr(unsafe.Pointer(p)), uintptr(unsafe.Pointer(Key)))
}
