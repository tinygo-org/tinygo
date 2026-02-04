//go:build uefi

package uefi

import (
	"unsafe"
)

// SimpleTextInputProtocol provides text input services.
// Kept for SystemTable layout compatibility.
type SimpleTextInputProtocol struct {
	Reset         uintptr
	ReadKeyStroke uintptr
	WaitForKey    uintptr
}

// SimpleTextInputExProtocol provides extended text input services.
// Located via BootServices->LocateProtocol using its GUID.
type SimpleTextInputExProtocol struct {
	Reset               uintptr
	ReadKeyStrokeEx     uintptr
	WaitForKeyEx        uintptr
	SetState            uintptr
	RegisterKeyNotify   uintptr
	UnregisterKeyNotify uintptr
}

// InputKey represents a keystroke from UEFI console input.
type InputKey struct {
	ScanCode    uint16
	UnicodeChar uint16
}

// KeyState contains the shift and toggle state for a key press.
type KeyState struct {
	KeyShiftState  uint32
	KeyToggleState uint8
	_              [3]byte
}

// KeyData is the extended key data returned by ReadKeyStrokeEx.
type KeyData struct {
	Key      InputKey
	KeyState KeyState
}

// GUID for EFI_SIMPLE_TEXT_INPUT_EX_PROTOCOL
var simpleTextInputExGUID = GUID{
	Data1: 0xdd9e7534,
	Data2: 0x7762,
	Data3: 0x4698,
	Data4: [8]uint8{0x8c, 0x14, 0xf5, 0x85, 0x17, 0xa6, 0x25, 0xaa},
}

// Cached pointer to the Ex protocol
var conInEx *SimpleTextInputExProtocol

// getConInEx locates and caches the SimpleTextInputExProtocol.
func getConInEx() *SimpleTextInputExProtocol {
	if conInEx != nil {
		return conInEx
	}

	st := GetSystemTable()
	if st == nil || st.BootServices == nil || st.BootServices.LocateProtocol == 0 {
		return nil
	}

	var iface uintptr
	status := Call(
		st.BootServices.LocateProtocol,
		uintptr(unsafe.Pointer(&simpleTextInputExGUID)),
		0, // Registration = NULL
		uintptr(unsafe.Pointer(&iface)),
	)
	if status != Success || iface == 0 {
		return nil
	}

	conInEx = (*SimpleTextInputExProtocol)(unsafe.Pointer(iface))
	return conInEx
}

// Scan codes for special keys
const (
	ScanNull  = 0x00
	ScanUp    = 0x01
	ScanDown  = 0x02
	ScanRight = 0x03
	ScanLeft  = 0x04
	ScanHome  = 0x05
	ScanEnd   = 0x06
	ScanIns   = 0x07
	ScanDel   = 0x08
	ScanPgUp  = 0x09
	ScanPgDn  = 0x0A
	ScanF1    = 0x0B
	ScanF2    = 0x0C
	ScanF3    = 0x0D
	ScanF4    = 0x0E
	ScanF5    = 0x0F
	ScanF6    = 0x10
	ScanF7    = 0x11
	ScanF8    = 0x12
	ScanF9    = 0x13
	ScanF10   = 0x14
	ScanEsc   = 0x17
)

// Key input buffer (stores bytes not InputKey)
var (
	keyBuffer     [256]byte
	keyBufferHead int
	keyBufferTail int
)

// KeyBufferAvailable returns the number of bytes in the buffer.
func KeyBufferAvailable() int {
	if keyBufferHead >= keyBufferTail {
		return keyBufferHead - keyBufferTail
	}
	return len(keyBuffer) - keyBufferTail + keyBufferHead
}

// keyBufferPushByte adds a byte to the buffer (internal use).
func keyBufferPushByte(b byte) bool {
	nextHead := (keyBufferHead + 1) % len(keyBuffer)
	if nextHead == keyBufferTail {
		return false // Buffer full
	}
	keyBuffer[keyBufferHead] = b
	keyBufferHead = nextHead
	return true
}

// keyBufferPushBytes adds multiple bytes to the buffer (internal use).
func keyBufferPushBytes(bytes []byte) bool {
	for _, b := range bytes {
		if !keyBufferPushByte(b) {
			return false
		}
	}
	return true
}

// KeyBufferPop removes and returns a byte from the buffer.
func KeyBufferPop() (byte, bool) {
	if keyBufferHead == keyBufferTail {
		return 0, false // Buffer empty
	}
	b := keyBuffer[keyBufferTail]
	keyBufferTail = (keyBufferTail + 1) % len(keyBuffer)
	return b, true
}

// convertKeyToBytes converts an InputKey to bytes (handling special keys).
func convertKeyToBytes(key InputKey) {
	// Handle special scan codes (arrow keys, function keys, etc.)
	if key.ScanCode != ScanNull {
		var seq []byte
		switch key.ScanCode {
		case ScanUp:
			seq = []byte{0x1b, '[', 'A'}
		case ScanDown:
			seq = []byte{0x1b, '[', 'B'}
		case ScanRight:
			seq = []byte{0x1b, '[', 'C'}
		case ScanLeft:
			seq = []byte{0x1b, '[', 'D'}
		case ScanHome:
			seq = []byte{0x1b, '[', 'H'}
		case ScanEnd:
			seq = []byte{0x1b, '[', 'F'}
		case ScanIns:
			seq = []byte{0x1b, '[', '2', '~'}
		case ScanDel:
			seq = []byte{0x1b, '[', '3', '~'}
		case ScanPgUp:
			seq = []byte{0x1b, '[', '5', '~'}
		case ScanPgDn:
			seq = []byte{0x1b, '[', '6', '~'}
		case ScanF1:
			seq = []byte{0x1b, 'O', 'P'}
		case ScanF2:
			seq = []byte{0x1b, 'O', 'Q'}
		case ScanF3:
			seq = []byte{0x1b, 'O', 'R'}
		case ScanF4:
			seq = []byte{0x1b, 'O', 'S'}
		case ScanF5:
			seq = []byte{0x1b, '[', '1', '5', '~'}
		case ScanF6:
			seq = []byte{0x1b, '[', '1', '7', '~'}
		case ScanF7:
			seq = []byte{0x1b, '[', '1', '8', '~'}
		case ScanF8:
			seq = []byte{0x1b, '[', '1', '9', '~'}
		case ScanF9:
			seq = []byte{0x1b, '[', '2', '0', '~'}
		case ScanF10:
			seq = []byte{0x1b, '[', '2', '1', '~'}
		case ScanEsc:
			seq = []byte{0x1b}
		default:
			return // Unknown scan code, ignore
		}
		keyBufferPushBytes(seq)
		return
	}

	// Handle regular characters
	if c := byte(key.UnicodeChar); c < 128 {
		if c != 0 {
			keyBufferPushByte(c)
		}
		return
	}

	// Non-ASCII character, output '?'
	keyBufferPushByte('?')
}

// ReadKey reads all available keys using the Ex protocol and buffers them.
func ReadKey() {
	ex := getConInEx()
	if ex == nil {
		return
	}

	for {
		var keyData KeyData
		status := Call(
			ex.ReadKeyStrokeEx,
			uintptr(unsafe.Pointer(ex)),
			uintptr(unsafe.Pointer(&keyData)),
		)
		if status != Success {
			break
		}
		convertKeyToBytes(keyData.Key)
	}
}

// IsKeyPressed triggers UEFI's console driver to check for pending input
// using the Ex protocol's WaitForKeyEx event.
func IsKeyPressed() bool {
	ex := getConInEx()
	if ex == nil {
		return false
	}

	st := GetSystemTable()
	if st == nil || st.BootServices == nil || st.BootServices.CheckEvent == 0 {
		return false
	}

	return Call(st.BootServices.CheckEvent, ex.WaitForKeyEx) == Success
}
