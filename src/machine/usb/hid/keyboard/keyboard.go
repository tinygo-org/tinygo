package keyboard

import (
	"errors"
	"machine/usb/hid"
)

// from usb-hid-keyboard.go
var (
	ErrInvalidCodepoint = errors.New("invalid Unicode codepoint")
	ErrInvalidKeycode   = errors.New("invalid keyboard keycode")
	ErrInvalidUTF8      = errors.New("invalid UTF-8 encoding")
	ErrKeypressMaximum  = errors.New("maximum keypresses exceeded")
)

var Keyboard *keyboard

// Keyboard represents a USB HID keyboard device with support for international
// layouts and various control, system, multimedia, and consumer keycodes.
//
// Keyboard implements the io.Writer interface that translates UTF-8 encoded
// byte strings into sequences of keypress events.
type keyboard struct {
	// led holds the current state of all keyboard LEDs:
	//   1=NumLock  2=CapsLock  4=ScrollLock  8=Compose  16=Kana
	led uint8

	// mod holds the current state of all keyboard modifier keys:
	//    1=LeftCtrl    2=LeftShift    4=LeftAlt     8=LeftGUI
	//   16=RightCtrl  32=RightShift  64=RightAlt  128=RightGUI
	mod uint8

	// key holds a list of all keyboard keys currently pressed.
	key [hidKeyboardKeyCount]uint8
	con [hidKeyboardConCount]uint16
	sys [hidKeyboardSysCount]uint8

	// decode holds the current state of the UTF-8 decoder.
	decode decodeState

	// wideChar holds high bits for the UTF-8 decoder.
	wideChar uint16

	buf     *hid.RingBuffer
	waitTxc bool
}

// decodeState represents a state in the UTF-8 decode state machine.
type decodeState uint8

// Constant enumerated values of type decodeState.
const (
	decodeReset decodeState = iota
	decodeByte1
	decodeByte2
	decodeByte3
)

func init() {
	if Keyboard == nil {
		Keyboard = newKeyboard()
		hid.SetHandler(Keyboard)
	}
}

// New returns hid-keybord.
func New() *keyboard {
	return Keyboard
}

func newKeyboard() *keyboard {
	return &keyboard{
		buf: hid.NewRingBuffer(),
	}
}

func (kb *keyboard) Handler() bool {
	kb.waitTxc = false
	if b, ok := kb.buf.Get(); ok {
		kb.waitTxc = true
		hid.SendUSBPacket(b)
		return true
	}
	return false
}

func (kb *keyboard) tx(b []byte) {
	if kb.waitTxc {
		kb.buf.Put(b)
	} else {
		kb.waitTxc = true
		hid.SendUSBPacket(b)
	}
}

func (kb *keyboard) ready() bool {
	return true
}

// Write transmits press-and-release key sequences for each Keycode translated
// from the given UTF-8 byte string. Write implements the io.Writer interface
// and conforms to all documented conventions for arguments and return values.
func (kb *keyboard) Write(b []byte) (n int, err error) {
	for _, c := range b {
		if err = kb.WriteByte(c); nil != err {
			break
		}
		n += 1
	}
	return
}

// WriteByte processes a single byte from a UTF-8 byte string. This method is a
// stateful method with respect to the receiver Keyboard, meaning that its exact
// behavior will depend on the current state of its UTF-8 decode state machine:
//
//  (a) If the given byte is a valid ASCII encoding (0-127), then a keypress
//      sequence is immediately transmitted for the respective Keycode.
//
//  (b) If the given byte represents the final byte in a multi-byte codepoint,
//      then a keypress sequence is immediately transmitted by translating the
//      multi-byte codepoint to its respective Keycode.
//
//  (c) If the given byte appears to represent high bits for a multi-byte
//      codepoint, then the bits are copied to the receiver's internal state
//      machine buffer for use by a subsequent call to WriteByte() (or Write())
//      that completes the codepoint.
//
//  (d) If the given byte is out of range, or contains illegal bits for the
//      current state of the UTF-8 decoder, then the UTF-8 decode state machine
//      is reset to its initial state.
//
// In cases (c) and (d), a keypress sequence is not generated and no data is
// transmitted. In case (c), additional bytes must be received via WriteByte()
// (or Write()) to complete or discard the current codepoint.
func (kb *keyboard) WriteByte(b byte) error {
	switch {
	case b < 0x80:
		// 1-byte encoding (0x00-0x7F)
		kb.decode = decodeByte1
		return kb.write(uint16(b))

	case b < 0xC0:
		// 2nd, 3rd, or 4th byte (0x80-0xBF)
		b = Keycode(b).key()
		switch kb.decode {
		case decodeByte2:
			kb.decode = decodeByte1
			return kb.write(kb.wideChar | uint16(b))
		case decodeByte3:
			kb.decode = decodeByte2
			kb.wideChar |= uint16(b) << 6
		}

	case b < 0xE0:
		// 2-byte encoding (0xC2-0xDF), or illegal byte 2 (0xC0-0xC1)
		kb.decode = decodeByte2
		kb.wideChar = uint16(b&0x1F) << 6

	case b < 0xF0:
		// 3-byte encoding (0xE0-0xEF)
		kb.decode = decodeByte3
		kb.wideChar = uint16(b&0x0F) << 12

	default:
		// 4-byte encoding unsupported (0xF0-0xF4), or illegal byte 4 (0xF5-0xFF)
		kb.decode = decodeReset
		return ErrInvalidUTF8
	}
	return nil
}

func (kb *keyboard) write(p uint16) error {
	c := keycode(p)
	if 0 == c {
		return ErrInvalidCodepoint
	}
	if d := deadkey(c); 0 != d {
		if err := kb.writeKeycode(d); nil != err {
			return err
		}
	}
	return kb.writeKeycode(c)
}

func (kb *keyboard) writeKeycode(c Keycode) error {
	var b [9]byte
	b[0] = 0x02
	b[1] = c.mod()
	b[2] = 0
	b[3] = c.key()
	b[4] = 0
	b[5] = 0
	b[6] = 0
	b[7] = 0
	b[8] = 0
	if !kb.sendKey(false, b[:]) {
		return hid.ErrHIDReportTransfer
	}

	b[1] = 0
	b[3] = 0
	if !kb.sendKey(false, b[:]) {
		return hid.ErrHIDReportTransfer
	}
	return nil
}

// Press transmits a press-and-release sequence for the given Keycode, which
// simulates a discrete keypress event.
//
// The following values of Keycode are supported:
//
//   0x0020 - 0x007F  ASCII               (U+0020 to U+007F)  [USES LAYOUT]
//   0x0080 - 0xC1FF  Unicode             (U+0080 to U+C1FF)  [USES LAYOUT]
//   0xC200 - 0xDFFF  UTF-8 packed        (U+0080 to U+07FF)  [USES LAYOUT]
//   0xE000 - 0xE0FF  Modifier key        (bitmap, 8 keys, Shift/Ctrl/Alt/GUI)
//   0xE200 - 0xE2FF  System key          (HID usage code, page 1)
//   0xE400 - 0xE7FF  Media/Consumer key  (HID usage code, page 12)
//   0xF000 - 0xFFFF  Normal key          (HID usage code, page 7)
func (kb *keyboard) Press(c Keycode) error {
	if err := kb.Down(c); nil != err {
		return err
	}
	return kb.Up(c)
}

func (kb *keyboard) sendKey(consumer bool, b []byte) bool {
	kb.tx(b)
	return true
}

func (kb *keyboard) keyboardSendKeys(consumer bool) bool {
	var b [9]byte
	b[0] = 0x02
	b[1] = kb.mod
	b[2] = 0x02
	b[3] = kb.key[0]
	b[4] = kb.key[1]
	b[5] = kb.key[2]
	b[6] = kb.key[3]
	b[7] = kb.key[4]
	b[8] = kb.key[5]
	return kb.sendKey(consumer, b[:])
}

// Down transmits a key-down event for the given Keycode.
//
// The host will interpret the key as being held down continuously until a
// corresponding key-up event is transmitted, e.g., via method Up().
//
// See godoc comment on method Press() for details on what input is accepted and
// how it is interpreted.
func (kb *keyboard) Down(c Keycode) error {
	var res uint8
	msb := c >> 8
	if msb >= 0xC2 {
		if msb < 0xE0 {
			c = ((msb & 0x1F) << 6) | Keycode(c.key())
		} else {
			switch msb {
			case 0xF0:
				return kb.down(uint8(c), 0)

			case 0xE0:
				return kb.down(0, uint8(c))

			case 0xE2:
				return kb.downSys(uint8(c))

			default:
				if 0xE4 <= msb && msb <= 0xE7 {
					return kb.downCon(uint16(c & 0x03FF))
				}
				return ErrInvalidKeycode
			}
		}
	}
	c = keycode(uint16(c))
	if 0 == c {
		return ErrInvalidCodepoint
	}
	if d := deadkey(c); 0 != d {
		res = kb.mod
		if 0 != res {
			kb.mod = 0
			kb.keyboardSendKeys(false)
		}
		kb.down(d.key(), d.mod())
		kb.up(d.key(), d.mod())
	}
	return kb.down(c.key(), c.mod()|res)
}

func (kb *keyboard) down(key uint8, mod uint8) error {
	send := false
	if 0 != mod {
		if kb.mod&mod != mod {
			kb.mod |= mod
			send = true
		}
	}
	if 0 != key {
		for _, k := range kb.key {
			if k == key {
				goto end
			}
		}
		for i, k := range kb.key {
			if 0 == k {
				kb.key[i] = key
				send = true
				goto end
			}
		}
		return ErrKeypressMaximum
	}
end:
	if send {
		if !kb.keyboardSendKeys(false) {
			return hid.ErrHIDReportTransfer
		}
	}
	return nil
}

func (kb *keyboard) downCon(key uint16) error {
	if 0 == key {
		return ErrInvalidKeycode
	}
	for _, k := range kb.con {
		if key == k {
			return nil // already pressed
		}
	}
	for i, k := range kb.con {
		if 0 == k {
			kb.con[i] = key
			if !kb.keyboardSendKeys(true) {
				return hid.ErrHIDReportTransfer
			}
			return nil
		}
	}
	return ErrKeypressMaximum
}

func (kb *keyboard) downSys(key uint8) error {
	if 0 == key {
		return ErrInvalidKeycode
	}
	for _, k := range kb.sys {
		if key == k {
			return nil // already pressed
		}
	}
	for i, k := range kb.sys {
		if 0 == k {
			kb.sys[i] = key
			if !kb.keyboardSendKeys(true) {
				return hid.ErrHIDReportTransfer
			}
			return nil
		}
	}
	return ErrKeypressMaximum
}

// Up transmits a key-up event for the given Keycode.
//
// See godoc comment on method Press() for details on what input is accepted and
// how it is interpreted.
func (kb *keyboard) Up(c Keycode) error {
	msb := c >> 8
	if msb >= 0xC2 {
		if msb < 0xE0 {
			c = ((msb & 0x1F) << 6) | Keycode(c.key())
		} else {
			switch msb {
			case 0xF0:
				return kb.up(uint8(c), 0)

			case 0xE0:
				return kb.up(0, uint8(c))

			case 0xE2:
				return kb.upSys(uint8(c))

			default:
				if 0xE4 <= msb && msb <= 0xE7 {
					return kb.upCon(uint16(c & 0x03FF))
				}
				return ErrInvalidKeycode
			}
		}
	}
	c = keycode(uint16(c))
	if 0 == c {
		return ErrInvalidCodepoint
	}
	return kb.up(c.key(), c.mod())
}

// Release transmits a key-up event for all keyboard keys currently pressed as
// if the user removed his/her hands from the keyboard entirely.
func (kb *keyboard) Release() error {

	bits := uint16(kb.mod)
	kb.mod = 0
	for i, k := range kb.key {
		bits |= uint16(k)
		kb.key[i] = 0
	}
	if 0 != bits {
		if !kb.keyboardSendKeys(false) {
			return hid.ErrHIDReportTransfer
		}
	}
	bits = 0
	for i, k := range kb.con {
		bits |= k
		kb.con[i] = 0
	}
	for i, k := range kb.sys {
		bits |= uint16(k)
		kb.sys[i] = 0
	}
	if 0 != bits {
		if !kb.keyboardSendKeys(true) {
			return hid.ErrHIDReportTransfer
		}
	}
	return nil
}

func (kb *keyboard) up(key uint8, mod uint8) error {
	send := false
	if 0 != mod {
		if kb.mod&mod != 0 {
			kb.mod &^= mod
			send = true
		}
	}
	if 0 != key {
		for i, k := range kb.key {
			if key == k {
				kb.key[i] = 0
				send = true
			}
		}
	}
	if send {
		if !kb.keyboardSendKeys(false) {
			return hid.ErrHIDReportTransfer
		}
	}
	return nil
}

func (kb *keyboard) upCon(key uint16) error {
	if 0 == key {
		return ErrInvalidKeycode
	}
	for i, k := range kb.con {
		if key == k {
			kb.con[i] = 0
			if !kb.keyboardSendKeys(true) {
				return hid.ErrHIDReportTransfer
			}
			return nil
		}
	}
	return nil
}

func (kb *keyboard) upSys(key uint8) error {
	if 0 == key {
		return ErrInvalidKeycode
	}
	for i, k := range kb.sys {
		if key == k {
			kb.sys[i] = 0
			if !kb.keyboardSendKeys(true) {
				return hid.ErrHIDReportTransfer
			}
			return nil
		}
	}
	return nil
}
