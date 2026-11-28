package uefi

import "unsafe"

// EFI_INPUT_KEY is a key press read from the console. Keys without a
// Unicode character, such as the arrow keys, have UnicodeChar set to 0 and
// are identified by ScanCode.
type EFI_INPUT_KEY struct {
	ScanCode    uint16
	UnicodeChar CHAR16
}

// EFI_SIMPLE_TEXT_INPUT_PROTOCOL reads keys from the console.
// See UEFI Specification 2.10, section 12.3.
type EFI_SIMPLE_TEXT_INPUT_PROTOCOL struct {
	reset         uintptr
	readKeyStroke uintptr
	WaitForKey    EFI_EVENT
}

func (p *EFI_SIMPLE_TEXT_INPUT_PROTOCOL) Reset(extendedVerification BOOLEAN) EFI_STATUS {
	return UefiCall2(p.reset, uintptr(unsafe.Pointer(p)), booleanArg(extendedVerification))
}

// ReadKeyStroke reads the next key without waiting. It returns EFI_NOT_READY
// if no key is waiting.
func (p *EFI_SIMPLE_TEXT_INPUT_PROTOCOL) ReadKeyStroke(key *EFI_INPUT_KEY) EFI_STATUS {
	return UefiCall2(p.readKeyStroke, uintptr(unsafe.Pointer(p)), uintptr(unsafe.Pointer(key)))
}
