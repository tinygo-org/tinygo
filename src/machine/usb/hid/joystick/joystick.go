package joystick

import (
	"machine"
	"machine/usb"
	"machine/usb/descriptor"
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
		Joystick = newDefaultJoystick()
	}
}

// UseSettings overrides the Joystick settings. This function must be
// called from init().
func UseSettings(def Definitions, rxHandlerFunc func(b []byte), setupFunc func(setup usb.Setup) bool, hidDesc []byte) *joystick {
	js := &joystick{
		buf:   hid.NewRingBuffer(),
		State: def.NewState(),
	}
	if setupFunc == nil {
		setupFunc = hid.DefaultSetupHandler
	}
	if rxHandlerFunc == nil {
		rxHandlerFunc = js.rxHandlerFunc
	}
	if len(hidDesc) == 0 {
		hidDesc = descriptor.JoystickDefaultHIDReport
	}
	class, err := descriptor.FindClassHIDType(descriptor.CDCJoystick.Configuration, descriptor.ClassHIDJoystick.Bytes())
	if err != nil {
		// TODO: some way to notify about error
		return nil
	}

	class.ClassLength(uint16(len(hidDesc)))
	descriptor.CDCJoystick.HID[2] = hidDesc

	machine.ConfigureUSBEndpoint(descriptor.CDCJoystick,
		[]usb.EndpointConfig{
			{
				No:        usb.HID_ENDPOINT_OUT,
				IsIn:      false,
				Type:      usb.ENDPOINT_TYPE_INTERRUPT,
				RxHandler: rxHandlerFunc,
			},
			{
				No:        usb.HID_ENDPOINT_IN,
				IsIn:      true,
				Type:      usb.ENDPOINT_TYPE_INTERRUPT,
				TxHandler: js.handler,
			},
		},
		[]usb.SetupConfig{
			{
				No:      usb.HID_INTERFACE,
				Handler: setupFunc,
			},
		},
	)
	Joystick = js
	return js
}

func newDefaultJoystick() *joystick {
	def := DefaultDefinitions()
	js := UseSettings(def, nil, nil, nil)
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
