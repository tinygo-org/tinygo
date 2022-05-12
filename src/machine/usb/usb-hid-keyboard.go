//go:build usb.hid
// +build usb.hid

package usb

import "errors"

var (
	ErrInvalidCodepoint = errors.New("invalid Unicode codepoint")
	ErrInvalidKeycode   = errors.New("invalid keyboard keycode")
	ErrInvalidUTF8      = errors.New("invalid UTF-8 encoding")
	ErrKeypressMaximum  = errors.New("maximum keypresses exceeded")
)

// Keyboard represents a USB HID keyboard device with support for international
// layouts and various control, system, multimedia, and consumer keycodes.
//
// Keyboard implements the io.Writer interface that translates UTF-8 encoded
// byte strings into sequences of keypress events.
type Keyboard struct {
	dc *dcd
	hc *descHIDClass

	// led holds the current state of all keyboard LEDs:
	//   1=NumLock  2=CapsLock  4=ScrollLock  8=Compose  16=Kana
	led uint8

	// mod holds the current state of all keyboard modifier keys:
	//    1=LeftCtrl    2=LeftShift    4=LeftAlt     8=LeftGUI
	//   16=RightCtrl  32=RightShift  64=RightAlt  128=RightGUI
	mod uint8

	// key holds a list of all keyboard keys currently pressed.
	key *[hidKeyboardKeyCount]uint8
	con *[hidKeyboardConCount]uint16
	sys *[hidKeyboardSysCount]uint8

	// decode holds the current state of the UTF-8 decoder.
	decode decodeState

	// wideChar holds high bits for the UTF-8 decoder.
	wideChar uint16
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

// configure initializes the receiver Keyboard by associating it with the given
// USB device controller driver and HID class configuration.
func (kb *Keyboard) configure(dc *dcd, hc *descHIDClass) {
	kb.dc = dc
	kb.hc = hc
}

func (kb *Keyboard) ready() bool {
	return kb.dc != nil && kb.hc != nil
}

// Write transmits press-and-release key sequences for each Keycode translated
// from the given UTF-8 byte string. Write implements the io.Writer interface
// and conforms to all documented conventions for arguments and return values.
func (kb *Keyboard) Write(b []byte) (n int, err error) {
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
func (kb *Keyboard) WriteByte(b byte) error {
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

func (kb *Keyboard) write(p uint16) error {
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

func (kb *Keyboard) writeKeycode(c Keycode) error {
	kb.mod = c.mod()
	kb.key[0] = c.key()
	kb.key[1] = 0
	kb.key[2] = 0
	kb.key[3] = 0
	kb.key[4] = 0
	kb.key[5] = 0
	if !kb.dc.keyboardSendKeys(false) {
		return ErrHIDReportTransfer
	}
	kb.mod = 0
	kb.key[0] = 0
	if !kb.dc.keyboardSendKeys(false) {
		return ErrHIDReportTransfer
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
func (kb *Keyboard) Press(c Keycode) error {
	if err := kb.Down(c); nil != err {
		return err
	}
	return kb.Up(c)
}

// Down transmits a key-down event for the given Keycode.
//
// The host will interpret the key as being held down continuously until a
// corresponding key-up event is transmitted, e.g., via method Up().
//
// See godoc comment on method Press() for details on what input is accepted and
// how it is interpreted.
func (kb *Keyboard) Down(c Keycode) error {
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
			kb.dc.keyboardSendKeys(false)
		}
		kb.down(d.key(), d.mod())
		kb.up(d.key(), d.mod())
	}
	return kb.down(c.key(), c.mod()|res)
}

func (kb *Keyboard) down(key uint8, mod uint8) error {
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
		if !kb.dc.keyboardSendKeys(false) {
			return ErrHIDReportTransfer
		}
	}
	return nil
}

func (kb *Keyboard) downCon(key uint16) error {
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
			if !kb.dc.keyboardSendKeys(true) {
				return ErrHIDReportTransfer
			}
			return nil
		}
	}
	return ErrKeypressMaximum
}

func (kb *Keyboard) downSys(key uint8) error {
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
			if !kb.dc.keyboardSendKeys(true) {
				return ErrHIDReportTransfer
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
func (kb *Keyboard) Up(c Keycode) error {
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
func (kb *Keyboard) Release() error {

	bits := uint16(kb.mod)
	kb.mod = 0
	for i, k := range kb.key {
		bits |= uint16(k)
		kb.key[i] = 0
	}
	if 0 != bits {
		if !kb.dc.keyboardSendKeys(false) {
			return ErrHIDReportTransfer
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
		if !kb.dc.keyboardSendKeys(true) {
			return ErrHIDReportTransfer
		}
	}
	return nil
}

func (kb *Keyboard) up(key uint8, mod uint8) error {
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
		if !kb.dc.keyboardSendKeys(false) {
			return ErrHIDReportTransfer
		}
	}
	return nil
}

func (kb *Keyboard) upCon(key uint16) error {
	if 0 == key {
		return ErrInvalidKeycode
	}
	for i, k := range kb.con {
		if key == k {
			kb.con[i] = 0
			if !kb.dc.keyboardSendKeys(true) {
				return ErrHIDReportTransfer
			}
			return nil
		}
	}
	return nil
}

func (kb *Keyboard) upSys(key uint8) error {
	if 0 == key {
		return ErrInvalidKeycode
	}
	for i, k := range kb.sys {
		if key == k {
			kb.sys[i] = 0
			if !kb.dc.keyboardSendKeys(true) {
				return ErrHIDReportTransfer
			}
			return nil
		}
	}
	return nil
}

// Keycode is a package-defined bitmap used to encode the value of a given key.
type Keycode uint16

// keycode returns the given Unicode codepoint translated to a Keycode sequence.
// Unicode codepoints greater than U+FFFF are unsupported.
//go:inline
func keycode(p uint16) Keycode {
	if p < 0x80 {
		return ascii[p]
	} else if p >= 0xA0 && p < 0x0100 {
		return iso88591[p-0xA0]
	} else if uint16(UNICODE20AC) == p {
		return UNICODE20AC.mask()
	}
	return 0
}

//go:inline
func deadkey(c Keycode) Keycode {
	switch c & deadkeysMask {
	case acuteAccentBits:
		return deadkeyAcuteAccent
	case circumflexBits:
		return deadkeyCircumflex
	case diaeresisBits:
		return deadkeyDiaeresis
	case graveAccentBits:
		return deadkeyGraveAccent
	case tildeBits:
		return deadkeyTilde
	}
	return 0
}

//go:inline
func (c Keycode) mask() Keycode { return c & keycodeMask }

//go:inline
func (c Keycode) key() uint8 { return uint8(c & keyMask) }

//go:inline
func (c Keycode) mod() uint8 {
	var m Keycode
	if 0 != c&shiftMask {
		m |= KeyModifierShift
	}
	if 0 != c&altgrMask {
		m |= KeyModifierRightAlt
	}
	return uint8(m)
}

//go:inline
func (c Keycode) Shift() Keycode { return c | KeyModifierShift }

const (
	hidKeyboardKeyCount = 6 // Max number of simultaneous keypresses
	hidKeyboardSysCount = 3
	hidKeyboardConCount = 4
)

// Keycodes common to all Keyboard layouts
const (
	KeyModifierCtrl       Keycode = 0x01 | 0xE000
	KeyModifierShift      Keycode = 0x02 | 0xE000
	KeyModifierAlt        Keycode = 0x04 | 0xE000
	KeyModifierGUI        Keycode = 0x08 | 0xE000
	KeyModifierLeftCtrl   Keycode = 0x01 | 0xE000
	KeyModifierLeftShift  Keycode = 0x02 | 0xE000
	KeyModifierLeftAlt    Keycode = 0x04 | 0xE000
	KeyModifierLeftGUI    Keycode = 0x08 | 0xE000
	KeyModifierRightCtrl  Keycode = 0x10 | 0xE000
	KeyModifierRightShift Keycode = 0x20 | 0xE000
	KeyModifierRightAlt   Keycode = 0x40 | 0xE000
	KeyModifierRightGUI   Keycode = 0x80 | 0xE000

	KeySystemPowerDown Keycode = 0x81 | 0xE200
	KeySystemSleep     Keycode = 0x82 | 0xE200
	KeySystemWakeUp    Keycode = 0x83 | 0xE200

	KeyMediaPlay        Keycode = 0xB0 | 0xE400
	KeyMediaPause       Keycode = 0xB1 | 0xE400
	KeyMediaRecord      Keycode = 0xB2 | 0xE400
	KeyMediaFastForward Keycode = 0xB3 | 0xE400
	KeyMediaRewind      Keycode = 0xB4 | 0xE400
	KeyMediaNextTrack   Keycode = 0xB5 | 0xE400
	KeyMediaPrevTrack   Keycode = 0xB6 | 0xE400
	KeyMediaStop        Keycode = 0xB7 | 0xE400
	KeyMediaEject       Keycode = 0xB8 | 0xE400
	KeyMediaRandomPlay  Keycode = 0xB9 | 0xE400
	KeyMediaPlayPause   Keycode = 0xCD | 0xE400
	KeyMediaPlaySkip    Keycode = 0xCE | 0xE400
	KeyMediaMute        Keycode = 0xE2 | 0xE400
	KeyMediaVolumeInc   Keycode = 0xE9 | 0xE400
	KeyMediaVolumeDec   Keycode = 0xEA | 0xE400

	KeyA           Keycode = 4 | 0xF000
	KeyB           Keycode = 5 | 0xF000
	KeyC           Keycode = 6 | 0xF000
	KeyD           Keycode = 7 | 0xF000
	KeyE           Keycode = 8 | 0xF000
	KeyF           Keycode = 9 | 0xF000
	KeyG           Keycode = 10 | 0xF000
	KeyH           Keycode = 11 | 0xF000
	KeyI           Keycode = 12 | 0xF000
	KeyJ           Keycode = 13 | 0xF000
	KeyK           Keycode = 14 | 0xF000
	KeyL           Keycode = 15 | 0xF000
	KeyM           Keycode = 16 | 0xF000
	KeyN           Keycode = 17 | 0xF000
	KeyO           Keycode = 18 | 0xF000
	KeyP           Keycode = 19 | 0xF000
	KeyQ           Keycode = 20 | 0xF000
	KeyR           Keycode = 21 | 0xF000
	KeyS           Keycode = 22 | 0xF000
	KeyT           Keycode = 23 | 0xF000
	KeyU           Keycode = 24 | 0xF000
	KeyV           Keycode = 25 | 0xF000
	KeyW           Keycode = 26 | 0xF000
	KeyX           Keycode = 27 | 0xF000
	KeyY           Keycode = 28 | 0xF000
	KeyZ           Keycode = 29 | 0xF000
	Key1           Keycode = 30 | 0xF000
	Key2           Keycode = 31 | 0xF000
	Key3           Keycode = 32 | 0xF000
	Key4           Keycode = 33 | 0xF000
	Key5           Keycode = 34 | 0xF000
	Key6           Keycode = 35 | 0xF000
	Key7           Keycode = 36 | 0xF000
	Key8           Keycode = 37 | 0xF000
	Key9           Keycode = 38 | 0xF000
	Key0           Keycode = 39 | 0xF000
	KeyEnter       Keycode = 40 | 0xF000
	KeyEsc         Keycode = 41 | 0xF000
	KeyBackspace   Keycode = 42 | 0xF000
	KeyTab         Keycode = 43 | 0xF000
	KeySpace       Keycode = 44 | 0xF000
	KeyMinus       Keycode = 45 | 0xF000
	KeyEqual       Keycode = 46 | 0xF000
	KeyLeftBrace   Keycode = 47 | 0xF000
	KeyRightBrace  Keycode = 48 | 0xF000
	KeyBackslash   Keycode = 49 | 0xF000
	KeyNonUsNum    Keycode = 50 | 0xF000
	KeySemicolon   Keycode = 51 | 0xF000
	KeyQuote       Keycode = 52 | 0xF000
	KeyTilde       Keycode = 53 | 0xF000
	KeyComma       Keycode = 54 | 0xF000
	KeyPeriod      Keycode = 55 | 0xF000
	KeySlash       Keycode = 56 | 0xF000
	KeyCapsLock    Keycode = 57 | 0xF000
	KeyF1          Keycode = 58 | 0xF000
	KeyF2          Keycode = 59 | 0xF000
	KeyF3          Keycode = 60 | 0xF000
	KeyF4          Keycode = 61 | 0xF000
	KeyF5          Keycode = 62 | 0xF000
	KeyF6          Keycode = 63 | 0xF000
	KeyF7          Keycode = 64 | 0xF000
	KeyF8          Keycode = 65 | 0xF000
	KeyF9          Keycode = 66 | 0xF000
	KeyF10         Keycode = 67 | 0xF000
	KeyF11         Keycode = 68 | 0xF000
	KeyF12         Keycode = 69 | 0xF000
	KeyPrintscreen Keycode = 70 | 0xF000
	KeyScrollLock  Keycode = 71 | 0xF000
	KeyPause       Keycode = 72 | 0xF000
	KeyInsert      Keycode = 73 | 0xF000
	KeyHome        Keycode = 74 | 0xF000
	KeyPageUp      Keycode = 75 | 0xF000
	KeyDelete      Keycode = 76 | 0xF000
	KeyEnd         Keycode = 77 | 0xF000
	KeyPageDown    Keycode = 78 | 0xF000
	KeyRight       Keycode = 79 | 0xF000
	KeyLeft        Keycode = 80 | 0xF000
	KeyDown        Keycode = 81 | 0xF000
	KeyUp          Keycode = 82 | 0xF000
	KeyNumLock     Keycode = 83 | 0xF000
	KeypadSlash    Keycode = 84 | 0xF000
	KeypadAsterisk Keycode = 85 | 0xF000
	KeypadMinus    Keycode = 86 | 0xF000
	KeypadPlus     Keycode = 87 | 0xF000
	KeypadEnter    Keycode = 88 | 0xF000
	Keypad1        Keycode = 89 | 0xF000
	Keypad2        Keycode = 90 | 0xF000
	Keypad3        Keycode = 91 | 0xF000
	Keypad4        Keycode = 92 | 0xF000
	Keypad5        Keycode = 93 | 0xF000
	Keypad6        Keycode = 94 | 0xF000
	Keypad7        Keycode = 95 | 0xF000
	Keypad8        Keycode = 96 | 0xF000
	Keypad9        Keycode = 97 | 0xF000
	Keypad0        Keycode = 98 | 0xF000
	KeypadPeriod   Keycode = 99 | 0xF000
	KeyNonUSBS     Keycode = 100 | 0xF000
	KeyMenu        Keycode = 101 | 0xF000
	KeyF13         Keycode = 104 | 0xF000
	KeyF14         Keycode = 105 | 0xF000
	KeyF15         Keycode = 106 | 0xF000
	KeyF16         Keycode = 107 | 0xF000
	KeyF17         Keycode = 108 | 0xF000
	KeyF18         Keycode = 109 | 0xF000
	KeyF19         Keycode = 110 | 0xF000
	KeyF20         Keycode = 111 | 0xF000
	KeyF21         Keycode = 112 | 0xF000
	KeyF22         Keycode = 113 | 0xF000
	KeyF23         Keycode = 114 | 0xF000
	KeyF24         Keycode = 115 | 0xF000

	KeyUpArrow    Keycode = KeyUp
	KeyDownArrow  Keycode = KeyDown
	KeyLeftArrow  Keycode = KeyLeft
	KeyRightArrow Keycode = KeyRight
	KeyReturn     Keycode = KeyEnter
	KeyLeftCtrl   Keycode = KeyModifierLeftCtrl
	KeyLeftShift  Keycode = KeyModifierLeftShift
	KeyLeftAlt    Keycode = KeyModifierLeftAlt
	KeyLeftGUI    Keycode = KeyModifierLeftGUI
	KeyRightCtrl  Keycode = KeyModifierRightCtrl
	KeyRightShift Keycode = KeyModifierRightShift
	KeyRightAlt   Keycode = KeyModifierRightAlt
	KeyRightGUI   Keycode = KeyModifierRightGUI
)

// Keycodes for layout US English (0x0904)
const (
	keycodeMask Keycode = 0x07FF
	keyMask     Keycode = 0x003F

	shiftMask          Keycode = 0x0040
	altgrMask          Keycode = 0x0080
	deadkeysMask       Keycode = 0x0700
	circumflexBits     Keycode = 0x0100
	acuteAccentBits    Keycode = 0x0200
	graveAccentBits    Keycode = 0x0300
	tildeBits          Keycode = 0x0400
	diaeresisBits      Keycode = 0x0500
	deadkeyCircumflex  Keycode = Key6 | shiftMask
	deadkeyAcuteAccent Keycode = KeyQuote
	deadkeyGraveAccent Keycode = KeyTilde
	deadkeyTilde       Keycode = KeyTilde | shiftMask
	deadkeyDiaeresis   Keycode = KeyQuote | shiftMask

	ASCII00 Keycode = 0            //   0  NUL
	ASCII01 Keycode = 0            //   1  SOH
	ASCII02 Keycode = 0            //   2  STX
	ASCII03 Keycode = 0            //   3  ETX
	ASCII04 Keycode = 0            //   4  EOT
	ASCII05 Keycode = 0            //   5  ENQ
	ASCII06 Keycode = 0            //   6  ACK
	ASCII07 Keycode = 0            //   7  BEL
	ASCII08 Keycode = KeyBackspace //   8  BS
	ASCII09 Keycode = KeyTab       //   9  TAB
	ASCII0A Keycode = KeyEnter     //  10  LF
	ASCII0B Keycode = 0            //  11  VT
	ASCII0C Keycode = 0            //  12  FF
	ASCII0D Keycode = 0            //  13  CR
	ASCII0E Keycode = 0            //  14  SO
	ASCII0F Keycode = 0            //  15  SI
	ASCII10 Keycode = 0            //  16  DEL
	ASCII11 Keycode = 0            //  17  DC1
	ASCII12 Keycode = 0            //  18  DC2
	ASCII13 Keycode = 0            //  19  DC3
	ASCII14 Keycode = 0            //  20  DC4
	ASCII15 Keycode = 0            //  21  NAK
	ASCII16 Keycode = 0            //  22  SYN
	ASCII17 Keycode = 0            //  23  ETB
	ASCII18 Keycode = 0            //  24  CAN
	ASCII19 Keycode = 0            //  25  EM
	ASCII1A Keycode = 0            //  26  SUB
	ASCII1B Keycode = 0            //  27  ESC
	ASCII1C Keycode = 0            //  28  FS
	ASCII1D Keycode = 0            //  29  GS
	ASCII1E Keycode = 0            //  30  RS
	ASCII1F Keycode = 0            //  31  US

	ASCII20     Keycode = KeySpace                             //  32   SPACE
	ASCII21     Keycode = Key1 | shiftMask                     //  33   !
	ASCII22     Keycode = diaeresisBits | KeySpace             //  34   "
	ASCII23     Keycode = Key3 | shiftMask                     //  35   #
	ASCII24     Keycode = Key4 | shiftMask                     //  36   $
	ASCII25     Keycode = Key5 | shiftMask                     //  37   %
	ASCII26     Keycode = Key7 | shiftMask                     //  38   &
	ASCII27     Keycode = acuteAccentBits | KeySpace           //  39   '
	ASCII28     Keycode = Key9 | shiftMask                     //  40   (
	ASCII29     Keycode = Key0 | shiftMask                     //  41   )
	ASCII2A     Keycode = Key8 | shiftMask                     //  42   *
	ASCII2B     Keycode = KeyEqual | shiftMask                 //  43   +
	ASCII2C     Keycode = KeyComma                             //  44   ,
	ASCII2D     Keycode = KeyMinus                             //  45   -
	ASCII2E     Keycode = KeyPeriod                            //  46   .
	ASCII2F     Keycode = KeySlash                             //  47   /
	ASCII30     Keycode = Key0                                 //  48   0
	ASCII31     Keycode = Key1                                 //  49   1
	ASCII32     Keycode = Key2                                 //  50   2
	ASCII33     Keycode = Key3                                 //  51   3
	ASCII34     Keycode = Key4                                 //  52   4
	ASCII35     Keycode = Key5                                 //  53   5
	ASCII36     Keycode = Key6                                 //  54   6
	ASCII37     Keycode = Key7                                 //  55   7
	ASCII38     Keycode = Key8                                 //  55   8
	ASCII39     Keycode = Key9                                 //  57   9
	ASCII3A     Keycode = KeySemicolon | shiftMask             //  58   :
	ASCII3B     Keycode = KeySemicolon                         //  59   ;
	ASCII3C     Keycode = KeyComma | shiftMask                 //  60   <
	ASCII3D     Keycode = KeyEqual                             //  61   =
	ASCII3E     Keycode = KeyPeriod | shiftMask                //  62   >
	ASCII3F     Keycode = KeySlash | shiftMask                 //  63   ?
	ASCII40     Keycode = Key2 | shiftMask                     //  64   @
	ASCII41     Keycode = KeyA | shiftMask                     //  65   A
	ASCII42     Keycode = KeyB | shiftMask                     //  66   B
	ASCII43     Keycode = KeyC | shiftMask                     //  67   C
	ASCII44     Keycode = KeyD | shiftMask                     //  68   D
	ASCII45     Keycode = KeyE | shiftMask                     //  69   E
	ASCII46     Keycode = KeyF | shiftMask                     //  70   F
	ASCII47     Keycode = KeyG | shiftMask                     //  71   G
	ASCII48     Keycode = KeyH | shiftMask                     //  72   H
	ASCII49     Keycode = KeyI | shiftMask                     //  73   I
	ASCII4A     Keycode = KeyJ | shiftMask                     //  74   J
	ASCII4B     Keycode = KeyK | shiftMask                     //  75   K
	ASCII4C     Keycode = KeyL | shiftMask                     //  76   L
	ASCII4D     Keycode = KeyM | shiftMask                     //  77   M
	ASCII4E     Keycode = KeyN | shiftMask                     //  78   N
	ASCII4F     Keycode = KeyO | shiftMask                     //  79   O
	ASCII50     Keycode = KeyP | shiftMask                     //  80   P
	ASCII51     Keycode = KeyQ | shiftMask                     //  81   Q
	ASCII52     Keycode = KeyR | shiftMask                     //  82   R
	ASCII53     Keycode = KeyS | shiftMask                     //  83   S
	ASCII54     Keycode = KeyT | shiftMask                     //  84   T
	ASCII55     Keycode = KeyU | shiftMask                     //  85   U
	ASCII56     Keycode = KeyV | shiftMask                     //  86   V
	ASCII57     Keycode = KeyW | shiftMask                     //  87   W
	ASCII58     Keycode = KeyX | shiftMask                     //  88   X
	ASCII59     Keycode = KeyY | shiftMask                     //  89   Y
	ASCII5A     Keycode = KeyZ | shiftMask                     //  90   Z
	ASCII5B     Keycode = KeyLeftBrace                         //  91   [
	ASCII5C     Keycode = KeyBackslash                         //  92   \
	ASCII5D     Keycode = KeyRightBrace                        //  93   ]
	ASCII5E     Keycode = circumflexBits | KeySpace            //  94   ^
	ASCII5F     Keycode = KeyMinus | shiftMask                 //  95
	ASCII60     Keycode = graveAccentBits | KeySpace           //  96   `
	ASCII61     Keycode = KeyA                                 //  97   a
	ASCII62     Keycode = KeyB                                 //  98   b
	ASCII63     Keycode = KeyC                                 //  99   c
	ASCII64     Keycode = KeyD                                 // 100   d
	ASCII65     Keycode = KeyE                                 // 101   e
	ASCII66     Keycode = KeyF                                 // 102   f
	ASCII67     Keycode = KeyG                                 // 103   g
	ASCII68     Keycode = KeyH                                 // 104   h
	ASCII69     Keycode = KeyI                                 // 105   i
	ASCII6A     Keycode = KeyJ                                 // 106   j
	ASCII6B     Keycode = KeyK                                 // 107   k
	ASCII6C     Keycode = KeyL                                 // 108   l
	ASCII6D     Keycode = KeyM                                 // 109   m
	ASCII6E     Keycode = KeyN                                 // 110   n
	ASCII6F     Keycode = KeyO                                 // 111   o
	ASCII70     Keycode = KeyP                                 // 112   p
	ASCII71     Keycode = KeyQ                                 // 113   q
	ASCII72     Keycode = KeyR                                 // 114   r
	ASCII73     Keycode = KeyS                                 // 115   s
	ASCII74     Keycode = KeyT                                 // 116   t
	ASCII75     Keycode = KeyU                                 // 117   u
	ASCII76     Keycode = KeyV                                 // 118   v
	ASCII77     Keycode = KeyW                                 // 119   w
	ASCII78     Keycode = KeyX                                 // 120   x
	ASCII79     Keycode = KeyY                                 // 121   y
	ASCII7A     Keycode = KeyZ                                 // 122   z
	ASCII7B     Keycode = KeyLeftBrace | shiftMask             // 123   {
	ASCII7C     Keycode = KeyBackslash | shiftMask             // 124   |
	ASCII7D     Keycode = KeyRightBrace | shiftMask            // 125   }
	ASCII7E     Keycode = tildeBits | KeySpace                 // 126   ~
	ASCII7F     Keycode = KeyBackspace                         // 127   DEL
	ISO88591A0  Keycode = KeySpace                             // 160         Nonbreakng Space
	ISO88591A1  Keycode = Key1 | altgrMask                     // 161   ¡     Inverted Exclamation
	ISO88591A2  Keycode = KeyC | altgrMask | shiftMask         // 162   ¢     Cent SIGN
	ISO88591A3  Keycode = Key4 | altgrMask | shiftMask         // 163   £     Pound Sign
	ISO88591A4  Keycode = Key4 | altgrMask                     // 164   ¤     Currency or Euro Sign
	ISO88591A5  Keycode = KeyMinus | altgrMask                 // 165   ¥     YEN SIGN
	ISO88591A6  Keycode = KeyBackslash | altgrMask | shiftMask // 166   ¦     BROKEN BAR			??
	ISO88591A7  Keycode = KeyS | altgrMask | shiftMask         // 167   §     SECTION SIGN
	ISO88591A8  Keycode = KeyQuote | altgrMask | shiftMask     // 168   ¨     DIAERESIS
	ISO88591A9  Keycode = KeyC | altgrMask                     // 169   ©     COPYRIGHT SIGN
	ISO88591AA  Keycode = 0                                    // 170   ª     FEMININE ORDINAL
	ISO88591AB  Keycode = KeyLeftBrace | altgrMask             // 171   «     LEFT DOUBLE ANGLE QUOTE
	ISO88591AC  Keycode = KeyBackslash | altgrMask             // 172   ¬     NOT SIGN			??
	ISO88591AD  Keycode = 0                                    // 173         SOFT HYPHEN
	ISO88591AE  Keycode = KeyR | altgrMask                     // 174   ®     REGISTERED SIGN
	ISO88591AF  Keycode = 0                                    // 175   ¯     MACRON
	ISO88591B0  Keycode = KeySemicolon | altgrMask | shiftMask // 176   °     DEGREE SIGN
	ISO88591B1  Keycode = 0                                    // 177   ±     PLUS-MINUS SIGN
	ISO88591B2  Keycode = Key2 | altgrMask                     // 178   ²     SUPERSCRIPT TWO
	ISO88591B3  Keycode = Key3 | altgrMask                     // 179   ³     SUPERSCRIPT THREE
	ISO88591B4  Keycode = KeyQuote | altgrMask                 // 180   ´     ACUTE ACCENT
	ISO88591B5  Keycode = KeyM | altgrMask                     // 181   µ     MICRO SIGN
	ISO88591B6  Keycode = KeySemicolon | altgrMask             // 182   ¶     PILCROW SIGN
	ISO88591B7  Keycode = 0                                    // 183   ·     MIDDLE DOT
	ISO88591B8  Keycode = 0                                    // 184   ¸     CEDILLA
	ISO88591B9  Keycode = Key1 | altgrMask | shiftMask         // 185   ¹     SUPERSCRIPT ONE
	ISO88591BA  Keycode = 0                                    // 186   º     MASCULINE ORDINAL
	ISO88591BB  Keycode = KeyRightBrace | altgrMask            // 187   »     RIGHT DOUBLE ANGLE QUOTE
	ISO88591BC  Keycode = Key6 | altgrMask                     // 188   ¼     FRACTION ONE QUARTER
	ISO88591BD  Keycode = Key7 | altgrMask                     // 189   ½     FRACTION ONE HALF
	ISO88591BE  Keycode = Key8 | altgrMask                     // 190   ¾     FRACTION THREE QUARTERS
	ISO88591BF  Keycode = KeySlash | altgrMask                 // 191   ¿     INVERTED QUESTION MARK
	ISO88591C0  Keycode = graveAccentBits | KeyA | shiftMask   // 192   À     A GRAVE
	ISO88591C1  Keycode = KeyA | altgrMask | shiftMask         // 193   Á     A ACUTE
	ISO88591C2  Keycode = circumflexBits | KeyA | shiftMask    // 194   Â     A CIRCUMFLEX
	ISO88591C3  Keycode = tildeBits | KeyA | shiftMask         // 195   Ã     A TILDE
	ISO88591C4  Keycode = KeyQ | altgrMask | shiftMask         // 196   Ä     A DIAERESIS
	ISO88591C5  Keycode = KeyW | altgrMask | shiftMask         // 197   Å     A RING ABOVE
	ISO88591C6  Keycode = KeyZ | altgrMask | shiftMask         // 198   Æ     AE
	ISO88591C7  Keycode = KeyComma | altgrMask | shiftMask     // 199   Ç     C CEDILLA
	ISO88591C8  Keycode = graveAccentBits | KeyE | shiftMask   // 200   È     E GRAVE
	ISO88591C9  Keycode = KeyE | altgrMask | shiftMask         // 201   É     E ACUTE
	ISO88591CA  Keycode = circumflexBits | KeyE | shiftMask    // 202   Ê     E CIRCUMFLEX
	ISO88591CB  Keycode = diaeresisBits | KeyE | shiftMask     // 203   Ë     E DIAERESIS
	ISO88591CC  Keycode = graveAccentBits | KeyI | shiftMask   // 204   Ì     I GRAVE
	ISO88591CD  Keycode = KeyI | altgrMask | shiftMask         // 205   Í     I ACUTE
	ISO88591CE  Keycode = circumflexBits | KeyI | shiftMask    // 206   Î     I CIRCUMFLEX
	ISO88591CF  Keycode = diaeresisBits | KeyI | shiftMask     // 207   Ï     I DIAERESIS
	ISO88591D0  Keycode = KeyD | altgrMask | shiftMask         // 208   Ð     ETH
	ISO88591D1  Keycode = KeyN | altgrMask | shiftMask         // 209   Ñ     N TILDE
	ISO88591D2  Keycode = graveAccentBits | KeyO | shiftMask   // 210   Ò     O GRAVE
	ISO88591D3  Keycode = KeyO | altgrMask | shiftMask         // 211   Ó     O ACUTE
	ISO88591D4  Keycode = circumflexBits | KeyO | shiftMask    // 212   Ô     O CIRCUMFLEX
	ISO88591D5  Keycode = tildeBits | KeyO | shiftMask         // 213   Õ     O TILDE
	ISO88591D6  Keycode = KeyP | altgrMask | shiftMask         // 214   Ö     O DIAERESIS
	ISO88591D7  Keycode = KeyEqual | altgrMask                 // 215   ×     MULTIPLICATION
	ISO88591D8  Keycode = KeyL | altgrMask | shiftMask         // 216   Ø     O STROKE
	ISO88591D9  Keycode = graveAccentBits | KeyU | shiftMask   // 217   Ù     U GRAVE
	ISO88591DA  Keycode = KeyU | altgrMask | shiftMask         // 218   Ú     U ACUTE
	ISO88591DB  Keycode = circumflexBits | KeyU | shiftMask    // 219   Û     U CIRCUMFLEX
	ISO88591DC  Keycode = KeyY | altgrMask | shiftMask         // 220   Ü     U DIAERESIS
	ISO88591DD  Keycode = acuteAccentBits | KeyY | shiftMask   // 221   Ý     Y ACUTE
	ISO88591DE  Keycode = KeyT | altgrMask | shiftMask         // 222   Þ     THORN
	ISO88591DF  Keycode = KeyS | altgrMask                     // 223   ß     SHARP S
	ISO88591E0  Keycode = graveAccentBits | KeyA               // 224   à     a GRAVE
	ISO88591E1  Keycode = KeyA | altgrMask                     // 225   á     a ACUTE
	ISO88591E2  Keycode = circumflexBits | KeyA                // 226   â     a CIRCUMFLEX
	ISO88591E3  Keycode = tildeBits | KeyA                     // 227   ã     a TILDE
	ISO88591E4  Keycode = diaeresisBits | KeyA                 // 228   ä     a DIAERESIS
	ISO88591E5  Keycode = KeyW | altgrMask                     // 229   å     a RING ABOVE
	ISO88591E6  Keycode = KeyZ | altgrMask                     // 230   æ     ae
	ISO88591E7  Keycode = KeyComma | altgrMask                 // 231   ç     c CEDILLA
	ISO88591E8  Keycode = graveAccentBits | KeyE               // 232   è     e GRAVE
	ISO88591E9  Keycode = acuteAccentBits | KeyE               // 233   é     e ACUTE
	ISO88591EA  Keycode = circumflexBits | KeyE                // 234   ê     e CIRCUMFLEX
	ISO88591EB  Keycode = diaeresisBits | KeyE                 // 235   ë     e DIAERESIS
	ISO88591EC  Keycode = graveAccentBits | KeyI               // 236   ì     i GRAVE
	ISO88591ED  Keycode = KeyI | altgrMask                     // 237   í     i ACUTE
	ISO88591EE  Keycode = circumflexBits | KeyI                // 238   î     i CIRCUMFLEX
	ISO88591EF  Keycode = diaeresisBits | KeyI                 // 239   ï     i DIAERESIS
	ISO88591F0  Keycode = KeyD | altgrMask                     // 240   ð     ETH
	ISO88591F1  Keycode = KeyN | altgrMask                     // 241   ñ     n TILDE
	ISO88591F2  Keycode = graveAccentBits | KeyO               // 242   ò     o GRAVE
	ISO88591F3  Keycode = KeyO | altgrMask                     // 243   ó     o ACUTE
	ISO88591F4  Keycode = circumflexBits | KeyO                // 244   ô     o CIRCUMFLEX
	ISO88591F5  Keycode = tildeBits | KeyO                     // 245   õ     o TILDE
	ISO88591F6  Keycode = KeyP | altgrMask                     // 246   ö     o DIAERESIS
	ISO88591F7  Keycode = KeyEqual | altgrMask | shiftMask     // 247   ÷     DIVISION
	ISO88591F8  Keycode = KeyL | altgrMask                     // 248   ø     o STROKE
	ISO88591F9  Keycode = graveAccentBits | KeyU               // 249   ù     u GRAVE
	ISO88591FA  Keycode = KeyU | altgrMask                     // 250   ú     u ACUTE
	ISO88591FB  Keycode = circumflexBits | KeyU                // 251   û     u CIRCUMFLEX
	ISO88591FC  Keycode = KeyY | altgrMask                     // 252   ü     u DIAERESIS
	ISO88591FD  Keycode = acuteAccentBits | KeyY               // 253   ý     y ACUTE
	ISO88591FE  Keycode = KeyT | altgrMask                     // 254   þ     THORN
	ISO88591FF  Keycode = diaeresisBits | KeyY                 // 255   ÿ     y DIAERESIS
	UNICODE20AC Keycode = Key5 | altgrMask                     // 20AC  €     Euro Sign
)

var ascii = [...]Keycode{
	ASCII00.mask(), ASCII01.mask(), ASCII02.mask(), ASCII03.mask(),
	ASCII04.mask(), ASCII05.mask(), ASCII06.mask(), ASCII07.mask(),
	ASCII08.mask(), ASCII09.mask(), ASCII0A.mask(), ASCII0B.mask(),
	ASCII0C.mask(), ASCII0D.mask(), ASCII0E.mask(), ASCII0F.mask(),
	ASCII10.mask(), ASCII11.mask(), ASCII12.mask(), ASCII13.mask(),
	ASCII14.mask(), ASCII15.mask(), ASCII16.mask(), ASCII17.mask(),
	ASCII18.mask(), ASCII19.mask(), ASCII1A.mask(), ASCII1B.mask(),
	ASCII1C.mask(), ASCII1D.mask(), ASCII1E.mask(), ASCII1F.mask(),
	ASCII20.mask(), ASCII21.mask(), ASCII22.mask(), ASCII23.mask(),
	ASCII24.mask(), ASCII25.mask(), ASCII26.mask(), ASCII27.mask(),
	ASCII28.mask(), ASCII29.mask(), ASCII2A.mask(), ASCII2B.mask(),
	ASCII2C.mask(), ASCII2D.mask(), ASCII2E.mask(), ASCII2F.mask(),
	ASCII30.mask(), ASCII31.mask(), ASCII32.mask(), ASCII33.mask(),
	ASCII34.mask(), ASCII35.mask(), ASCII36.mask(), ASCII37.mask(),
	ASCII38.mask(), ASCII39.mask(), ASCII3A.mask(), ASCII3B.mask(),
	ASCII3C.mask(), ASCII3D.mask(), ASCII3E.mask(), ASCII3F.mask(),
	ASCII40.mask(), ASCII41.mask(), ASCII42.mask(), ASCII43.mask(),
	ASCII44.mask(), ASCII45.mask(), ASCII46.mask(), ASCII47.mask(),
	ASCII48.mask(), ASCII49.mask(), ASCII4A.mask(), ASCII4B.mask(),
	ASCII4C.mask(), ASCII4D.mask(), ASCII4E.mask(), ASCII4F.mask(),
	ASCII50.mask(), ASCII51.mask(), ASCII52.mask(), ASCII53.mask(),
	ASCII54.mask(), ASCII55.mask(), ASCII56.mask(), ASCII57.mask(),
	ASCII58.mask(), ASCII59.mask(), ASCII5A.mask(), ASCII5B.mask(),
	ASCII5C.mask(), ASCII5D.mask(), ASCII5E.mask(), ASCII5F.mask(),
	ASCII60.mask(), ASCII61.mask(), ASCII62.mask(), ASCII63.mask(),
	ASCII64.mask(), ASCII65.mask(), ASCII66.mask(), ASCII67.mask(),
	ASCII68.mask(), ASCII69.mask(), ASCII6A.mask(), ASCII6B.mask(),
	ASCII6C.mask(), ASCII6D.mask(), ASCII6E.mask(), ASCII6F.mask(),
	ASCII70.mask(), ASCII71.mask(), ASCII72.mask(), ASCII73.mask(),
	ASCII74.mask(), ASCII75.mask(), ASCII76.mask(), ASCII77.mask(),
	ASCII78.mask(), ASCII79.mask(), ASCII7A.mask(), ASCII7B.mask(),
	ASCII7C.mask(), ASCII7D.mask(), ASCII7E.mask(), ASCII7F.mask(),
}

var iso88591 = [...]Keycode{
	ISO88591A0.mask(), ISO88591A1.mask(), ISO88591A2.mask(), ISO88591A3.mask(),
	ISO88591A4.mask(), ISO88591A5.mask(), ISO88591A6.mask(), ISO88591A7.mask(),
	ISO88591A8.mask(), ISO88591A9.mask(), ISO88591AA.mask(), ISO88591AB.mask(),
	ISO88591AC.mask(), ISO88591AD.mask(), ISO88591AE.mask(), ISO88591AF.mask(),
	ISO88591B0.mask(), ISO88591B1.mask(), ISO88591B2.mask(), ISO88591B3.mask(),
	ISO88591B4.mask(), ISO88591B5.mask(), ISO88591B6.mask(), ISO88591B7.mask(),
	ISO88591B8.mask(), ISO88591B9.mask(), ISO88591BA.mask(), ISO88591BB.mask(),
	ISO88591BC.mask(), ISO88591BD.mask(), ISO88591BE.mask(), ISO88591BF.mask(),
	ISO88591C0.mask(), ISO88591C1.mask(), ISO88591C2.mask(), ISO88591C3.mask(),
	ISO88591C4.mask(), ISO88591C5.mask(), ISO88591C6.mask(), ISO88591C7.mask(),
	ISO88591C8.mask(), ISO88591C9.mask(), ISO88591CA.mask(), ISO88591CB.mask(),
	ISO88591CC.mask(), ISO88591CD.mask(), ISO88591CE.mask(), ISO88591CF.mask(),
	ISO88591D0.mask(), ISO88591D1.mask(), ISO88591D2.mask(), ISO88591D3.mask(),
	ISO88591D4.mask(), ISO88591D5.mask(), ISO88591D6.mask(), ISO88591D7.mask(),
	ISO88591D8.mask(), ISO88591D9.mask(), ISO88591DA.mask(), ISO88591DB.mask(),
	ISO88591DC.mask(), ISO88591DD.mask(), ISO88591DE.mask(), ISO88591DF.mask(),
	ISO88591E0.mask(), ISO88591E1.mask(), ISO88591E2.mask(), ISO88591E3.mask(),
	ISO88591E4.mask(), ISO88591E5.mask(), ISO88591E6.mask(), ISO88591E7.mask(),
	ISO88591E8.mask(), ISO88591E9.mask(), ISO88591EA.mask(), ISO88591EB.mask(),
	ISO88591EC.mask(), ISO88591ED.mask(), ISO88591EE.mask(), ISO88591EF.mask(),
	ISO88591F0.mask(), ISO88591F1.mask(), ISO88591F2.mask(), ISO88591F3.mask(),
	ISO88591F4.mask(), ISO88591F5.mask(), ISO88591F6.mask(), ISO88591F7.mask(),
	ISO88591F8.mask(), ISO88591F9.mask(), ISO88591FA.mask(), ISO88591FB.mask(),
	ISO88591FC.mask(), ISO88591FD.mask(), ISO88591FE.mask(), ISO88591FF.mask(),
}
