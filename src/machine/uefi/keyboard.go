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

var specialKeys = map[uint16][]byte{
	ScanUp:    []byte{'\x1b', '[', 'A'},           // ESC [ A
	ScanDown:  []byte{'\x1b', '[', 'B'},           // ESC [ B
	ScanRight: []byte{'\x1b', '[', 'C'},           // ESC [ C
	ScanLeft:  []byte{'\x1b', '[', 'D'},           // ESC [ D
	ScanHome:  []byte{'\x1b', '[', 'H'},           // ESC [ H
	ScanEnd:   []byte{'\x1b', '[', 'F'},           // ESC [ F
	ScanIns:   []byte{'\x1b', '[', '2', '~'},      // ESC [ 2 ~
	ScanDel:   []byte{'\x1b', '[', '3', '~'},      // ESC [ 3 ~
	ScanPgUp:  []byte{'\x1b', '[', '5', '~'},      // ESC [ 5 ~
	ScanPgDn:  []byte{'\x1b', '[', '6', '~'},      // ESC [ 6 ~
	ScanF1:    []byte{'\x1b', 'O', 'P'},           // ESC O P
	ScanF2:    []byte{'\x1b', 'O', 'Q'},           // ESC O Q
	ScanF3:    []byte{'\x1b', 'O', 'R'},           // ESC O R
	ScanF4:    []byte{'\x1b', 'O', 'S'},           // ESC O S
	ScanF5:    []byte{'\x1b', '[', '1', '5', '~'}, // ESC [ 1 5 ~
	ScanF6:    []byte{'\x1b', '[', '1', '7', '~'}, // ESC [ 1 7 ~
	ScanF7:    []byte{'\x1b', '[', '1', '8', '~'}, // ESC [ 1 8 ~
	ScanF8:    []byte{'\x1b', '[', '1', '9', '~'}, // ESC [ 1 9 ~
	ScanF9:    []byte{'\x1b', '[', '2', '0', '~'}, // ESC [ 2 0 ~
	ScanF10:   []byte{'\x1b', '[', '2', '1', '~'}, // ESC [ 2 1 ~
	ScanEsc:   []byte{'\x1b'},                     // ESC
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

		// Converts an InputKey to bytes (handling special keys).
		if keyData.Key.ScanCode != ScanNull {
			if seq, ok := specialKeys[keyData.Key.ScanCode]; ok {
				keyBufferPushBytes(seq)
			}
			continue
		}

		// Handle regular characters
		if c := byte(keyData.Key.UnicodeChar); c < 128 {
			if c != 0 {
				keyBufferPushByte(c)
			}
			continue
		}

		// Non-ASCII character, output '?'
		keyBufferPushByte('?')
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
