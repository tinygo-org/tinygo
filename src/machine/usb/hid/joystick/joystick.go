package joystick

import (
	"machine"
	"machine/usb"
	"machine/usb/hid"
)

var Joystick *joystick

type joystick struct {
	State
	buf           *hid.RingBuffer
	waitTxc       bool
	rxHandlerFunc func(b []byte)
	setupFunc     func(setup usb.Setup) bool
}

func init() {
	if Joystick == nil {
		Joystick = newJoystick()
	}
}

func newJoystick() *joystick {
	def := DefaultDefinitions()
	js := &joystick{
		State: def.NewState(),
		buf:   hid.NewRingBuffer(),
	}
	machine.EnableJoystick(js.handler, js.rxHandler, hid.DefaultSetupHandler, def.Descriptor())
	return js
}

// Port returns the USB Joystick port.
func Port() *joystick {
	return Joystick
}

func (m *joystick) handler() {
	m.waitTxc = false
	if b, ok := m.buf.Get(); ok {
		m.waitTxc = true
		hid.SendUSBPacket(b)
	}
}

func (m *joystick) tx(b []byte) {
	if machine.USBDev.InitEndpointComplete {
		if m.waitTxc {
			m.buf.Put(b)
		} else {
			m.waitTxc = true
			hid.SendUSBPacket(b)
		}
	}
}

func (m *joystick) ready() bool {
	return true
}

func (m *joystick) rxHandler(b []byte) {
	if m.rxHandlerFunc != nil {
		m.rxHandlerFunc(b)
	}
}

// to InterruptOut
func (m *joystick) SendReport(reportID byte, b []byte) {
	m.tx(append([]byte{reportID}, b...))
}

func (m *joystick) SendState() {
	b, _ := m.State.MarshalBinary()
	m.tx(b)
}
