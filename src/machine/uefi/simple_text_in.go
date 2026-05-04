package uefi

import "unsafe"

func booleanArg(v BOOLEAN) uintptr {
	if v {
		return 1
	}
	return 0
}

type EFI_KEY_TOGGLE_STATE uint8

const (
	EFI_SCROLL_LOCK_ACTIVE EFI_KEY_TOGGLE_STATE = 0x01
	EFI_NUM_LOCK_ACTIVE    EFI_KEY_TOGGLE_STATE = 0x02
	EFI_CAPS_LOCK_ACTIVE   EFI_KEY_TOGGLE_STATE = 0x04
	EFI_KEY_STATE_EXPOSED  EFI_KEY_TOGGLE_STATE = 0x40
	EFI_TOGGLE_STATE_VALID EFI_KEY_TOGGLE_STATE = 0x80
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

type EFI_INPUT_KEY struct {
	ScanCode    uint16
	UnicodeChar CHAR16
}

type EFI_SIMPLE_TEXT_INPUT_PROTOCOL struct {
	reset         uintptr
	readKeyStroke uintptr
	WaitForKey    EFI_EVENT
}

func (p *EFI_SIMPLE_TEXT_INPUT_PROTOCOL) Reset(extendedVerification BOOLEAN) EFI_STATUS {
	return UefiCall2(p.reset, uintptr(unsafe.Pointer(p)), booleanArg(extendedVerification))
}

func (p *EFI_SIMPLE_TEXT_INPUT_PROTOCOL) ReadKeyStroke(key *EFI_INPUT_KEY) EFI_STATUS {
	return UefiCall2(p.readKeyStroke, uintptr(unsafe.Pointer(p)), uintptr(unsafe.Pointer(key)))
}

func (p *EFI_SIMPLE_TEXT_INPUT_PROTOCOL) GetKey() (EFI_INPUT_KEY, EFI_STATUS) {
	var key EFI_INPUT_KEY
	var status EFI_STATUS
	for {
		status = WaitForEvent(p.WaitForKey)
		if status != EFI_SUCCESS {
			return key, status
		}
		status = p.ReadKeyStroke(&key)
		if status == EFI_SUCCESS {
			return key, EFI_SUCCESS
		}
		if status != EFI_NOT_READY {
			return key, status
		}
	}
}

var SimpleTextInputExProtocolGUID = EFI_GUID{
	0xdd9e7534, 0x7762, 0x4698,
	[8]byte{0x8c, 0x14, 0xf5, 0x85, 0x17, 0xa6, 0x25, 0xaa},
}

type EFI_KEY_STATE struct {
	KeyShiftState  uint32
	KeyToggleState EFI_KEY_TOGGLE_STATE
}

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

func (p *EFI_SIMPLE_TEXT_INPUT_EX_PROTOCOL) Reset(extendedVerification BOOLEAN) EFI_STATUS {
	return UefiCall2(p.resetEx, uintptr(unsafe.Pointer(p)), booleanArg(extendedVerification))
}

func (p *EFI_SIMPLE_TEXT_INPUT_EX_PROTOCOL) ReadKeyStroke(key *EFI_KEY_DATA) EFI_STATUS {
	return UefiCall2(p.readKeyStrokeEx, uintptr(unsafe.Pointer(p)), uintptr(unsafe.Pointer(key)))
}

func (p *EFI_SIMPLE_TEXT_INPUT_EX_PROTOCOL) GetKey() (EFI_KEY_DATA, EFI_STATUS) {
	var key EFI_KEY_DATA
	var status EFI_STATUS
	for {
		status = WaitForEvent(p.WaitForKeyEx)
		if status != EFI_SUCCESS {
			return key, status
		}
		status = p.ReadKeyStroke(&key)
		if status == EFI_SUCCESS {
			return key, EFI_SUCCESS
		}
		if status != EFI_NOT_READY {
			return key, status
		}
	}
}

func SimpleTextInProtocol() (*EFI_SIMPLE_TEXT_INPUT_PROTOCOL, EFI_STATUS) {
	st := ST()
	if st == nil || st.ConIn == nil {
		return nil, EFI_NOT_FOUND
	}
	return st.ConIn, EFI_SUCCESS
}

func SimpleTextInExProtocol() (*EFI_SIMPLE_TEXT_INPUT_EX_PROTOCOL, EFI_STATUS) {
	var iface *EFI_SIMPLE_TEXT_INPUT_EX_PROTOCOL
	status := BS().LocateProtocol(&SimpleTextInputExProtocolGUID, nil, unsafe.Pointer(&iface))
	if status != EFI_SUCCESS {
		return nil, status
	}
	return iface, EFI_SUCCESS
}
