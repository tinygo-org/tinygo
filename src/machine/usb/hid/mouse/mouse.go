package mouse

import (
	"machine/usb/hid"
)

var Mouse *mouse

type Button byte

const (
	Left Button = 1 << iota
	Right
	Middle
)

type mouse struct {
	buf     *hid.RingBuffer
	button  Button
	waitTxc bool
}

func init() {
	if Mouse == nil {
		Mouse = newMouse()
		hid.SetCallbackHandler(Mouse)
	}
}

// New returns hid-mouse.
func New() *mouse {
	return Mouse
}

func newMouse() *mouse {
	return &mouse{
		buf: hid.NewRingBuffer(),
	}
}

func (m *mouse) Callback() bool {
	m.waitTxc = false
	if b, ok := m.buf.Get(); ok {
		m.waitTxc = true
		hid.SendUSBPacket(b[:5])
		return true
	}
	return false
}

func (m *mouse) tx(b []byte) {
	if m.waitTxc {
		m.buf.Put(b)
	} else {
		m.waitTxc = true
		hid.SendUSBPacket(b)
	}
}

// Move is a function that moves the mouse cursor.
func (m *mouse) Move(vx, vy int) {
	if vx == 0 && vy == 0 {
		return
	}

	if vx < -128 {
		vx = -128
	}
	if vx > 127 {
		vx = 127
	}

	if vy < -128 {
		vy = -128
	}
	if vy > 127 {
		vy = 127
	}

	m.tx([]byte{
		0x01, byte(m.button), byte(vx), byte(vy), 0x00,
	})
}

// Cilck clicks the mouse button.
func (m *mouse) Click(btn Button) {
	m.Press(btn)
	m.Release(btn)
}

// Press presses the given mouse buttons.
func (m *mouse) Press(btn Button) {
	m.button |= btn
	m.tx([]byte{
		0x01, byte(m.button), 0x00, 0x00, 0x00,
	})
}

// Release releases the given mouse buttons.
func (m *mouse) Release(btn Button) {
	m.button &= ^btn
	m.tx([]byte{
		0x01, byte(m.button), 0x00, 0x00, 0x00,
	})
}

// Wheel controls the mouse wheel.
func (m *mouse) Wheel(v int) {
	if v == 0 {
		return
	}

	if v < -128 {
		v = -128
	}
	if v > 127 {
		v = 127
	}

	m.tx([]byte{
		0x01, byte(m.button), 0x00, 0x00, byte(v),
	})
}

// WheelDown turns the mouse wheel down.
func (m *mouse) WheelDown() {
	m.Wheel(-1)
}

// WheelUp turns the mouse wheel up.
func (m *mouse) WheelUp() {
	m.Wheel(1)
}
