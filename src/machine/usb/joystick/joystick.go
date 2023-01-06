package joystick

import (
	"machine"
	"machine/usb"
)

const (
	jsEndpointOut = usb.HID_ENDPOINT_OUT // from PC
	jsEndpointIn  = usb.HID_ENDPOINT_IN  // to PC
)

var js *Joystick

type Joystick struct {
	State
	buf           *RingBuffer
	waitTxc       bool
	rxHandlerFunc func(b []byte)
	setupFunc     func(setup usb.Setup) bool
}

func Enable(def Definitions, rxHandlerFunc func(b []byte),
	setupFunc func(setup usb.Setup) bool, hidDesc []byte) *Joystick {
	m := &Joystick{
		buf:   NewRingBuffer(),
		State: def.NewState(),
	}
	m.State = def.NewState()
	if setupFunc == nil {
		setupFunc = m.setupFunc
	}
	machine.EnableJoystick(m.handler, rxHandlerFunc, setupFunc, hidDesc)
	js = m
	return m
}

// Port returns the USB Joystick port.
func Port() *Joystick {
	if js == nil {
		def := DefaultDefinitions()
		js = &Joystick{
			State: def.NewState(),
			buf:   NewRingBuffer(),
		}
		machine.EnableJoystick(js.handler, nil, js.setupFunc, def.Descriptor())
	}
	return js
}

func (m *Joystick) sendUSBPacket(b []byte) {
	machine.SendUSBInPacket(jsEndpointIn, b)
}

// from InterruptIn
func (m *Joystick) handler() {
	m.waitTxc = false
	if b, ok := m.buf.Get(); ok {
		m.waitTxc = true
		m.sendUSBPacket(b)
	}
}

func (m *Joystick) setup(setup usb.Setup) bool {
	if setup.BmRequestType == usb.SET_REPORT_TYPE && setup.BRequest == usb.SET_IDLE {
		machine.SendZlp()
		return true
	}
	return false
}

func (m *Joystick) rxHandler(b []byte) {
	if m.rxHandlerFunc != nil {
		m.rxHandlerFunc(b)
	}
}

func (m *Joystick) tx(b []byte) {
	if m.waitTxc {
		m.buf.Put(b)
	} else {
		m.waitTxc = true
		m.sendUSBPacket(b)
	}
}

// to InterruptOut
func (m *Joystick) SendReport(reportID byte, b []byte) {
	m.tx(append([]byte{reportID}, b...))
}

func (m *Joystick) SendState() {
	b, _ := m.State.MarshalBinary()
	m.tx(b)
}
