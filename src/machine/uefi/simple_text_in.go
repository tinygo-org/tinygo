package uefi

import (
	"unsafe"
)

type EFI_KEY_TOGGLE_STATE uint8

const (
	EFI_SCROLL_LOCK_ACTIVE EFI_KEY_TOGGLE_STATE = 0x01
	EFI_NUM_LOCK_ACTIVE                         = 0x02
	EFI_CAPS_LOCK_ACTIVE                        = 0x04
	EFI_KEY_STATE_EXPOSED                       = 0x40
	EFI_TOGGLE_STATE_VALID                      = 0x80
)

const (
	EFI_SHIFT_STATE_VALID     = 0x80000000
	EFI_RIGHT_SHIFT_PRESSED   = 0x00000001
	EFI_LEFT_SHIFT_PRESSED    = 0x00000002
	EFI_RIGHT_CONTROL_PRESSED = 0x00000004
	EFI_LEFT_CONTROL_PRESSED  = 0x00000008
	EFI_RIGHT_ALT_PRESSED     = 0x00000010
	EFI_LEFT_ALT_PRESSED      = 0x00000020
	EFI_RIGHT_LOGO_PRESSED    = 0x00000040
	EFI_LEFT_LOGO_PRESSED     = 0x00000080
	EFI_MENU_KEY_PRESSED      = 0x00000100
	EFI_SYS_REQ_PRESSED       = 0x00000200
)

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

// var EFI_SIMPLE_TEXT_INPUT_EX_PROTOCOL_GUID = EFI_GUID{
var SimpleTextInputExProtocolGUID = EFI_GUID{
	0xdd9e7534, 0x7762, 0x4698,
	[...]byte{0x8c, 0x14, 0xf5, 0x85, 0x17, 0xa6, 0x25, 0xaa}}

type EFI_KEY_DATA struct {
	Key      EFI_INPUT_KEY
	KeyState EFI_KEY_STATE
}

type EFI_SIMPLE_TEXT_INPUT_EX_PROTOCOL struct {
	resetEx                   uintptr
	readKeyStrokeEx           uintptr
	WaitForKeyEx              EFI_EVENT
	setState                  uintptr
	registerKeystrokeNotify   uintptr
	unregisterKeystrokeNotify uintptr
}

type EFI_KEY_STATE struct {
	KeyShiftState  uint32
	KeyToggleState EFI_KEY_TOGGLE_STATE
}

func (p *EFI_SIMPLE_TEXT_INPUT_EX_PROTOCOL) Reset(ExtendedVerification BOOLEAN) EFI_STATUS {
	return UefiCall2(p.resetEx, uintptr(unsafe.Pointer(p)), convertBoolean(ExtendedVerification))
}

func (p *EFI_SIMPLE_TEXT_INPUT_EX_PROTOCOL) ReadKeyStroke(Key *EFI_KEY_DATA) EFI_STATUS {
	return UefiCall2(p.readKeyStrokeEx, uintptr(unsafe.Pointer(p)), uintptr(unsafe.Pointer(Key)))
}

// SimpleTextInExProtocol finds and returns the first handle implementing this protocol
// and returns it. Usually there will only be one or multiple get multiplexed together.
func SimpleTextInExProtocol() (*EFI_SIMPLE_TEXT_INPUT_EX_PROTOCOL, error) {
	st := ST()
	var iFace unsafe.Pointer
	status := (*st).BootServices.LocateProtocol(
		&SimpleTextInputExProtocolGUID,
		nil,
		unsafe.Pointer(&iFace))
	if status == EFI_SUCCESS {
		stiep := (*EFI_SIMPLE_TEXT_INPUT_EX_PROTOCOL)(iFace)
		return stiep, nil
	}

	return nil, StatusError(status)
}

// GetKey blocks while waiting to receive a key. It yields to the scheduler
// so other goroutines may continue to work while waiting for a key press.
func (p *EFI_SIMPLE_TEXT_INPUT_EX_PROTOCOL) GetKey() EFI_KEY_DATA {
	var key EFI_KEY_DATA

	// Wait for key event, while yielding to other routines
	WaitForEvent(p.WaitForKeyEx)

	// Read key stroke
	status := p.ReadKeyStroke(&key)
	if status != EFI_SUCCESS {
		return key
	}

	return key
}
