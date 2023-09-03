package midi

import (
	"machine"
	"machine/usb"
	"machine/usb/descriptor"
)

const (
	midiEndpointOut = usb.MIDI_ENDPOINT_OUT // from PC
	midiEndpointIn  = usb.MIDI_ENDPOINT_IN  // to PC
)

var Midi *midi

type midi struct {
	msg       [4]byte
	buf       *RingBuffer
	rxHandler func([]byte)
	txHandler func()
	waitTxc   bool
}

func init() {
	if Midi == nil {
		Midi = newMidi()
	}
}

// New returns the USB MIDI port.
// Deprecated, better to just use Port()
func New() *midi {
	return Port()
}

// Port returns the USB midi port.
func Port() *midi {
	return Midi
}

func newMidi() *midi {
	m := &midi{
		buf: NewRingBuffer(),
	}
	machine.ConfigureUSBEndpoint(descriptor.CDCMIDI,
		[]usb.EndpointConfig{
			{
				Index:     usb.MIDI_ENDPOINT_OUT,
				IsIn:      false,
				Type:      usb.ENDPOINT_TYPE_BULK,
				RxHandler: m.RxHandler,
			},
			{
				Index:     usb.MIDI_ENDPOINT_IN,
				IsIn:      true,
				Type:      usb.ENDPOINT_TYPE_BULK,
				TxHandler: m.TxHandler,
			},
		},
		[]usb.SetupConfig{},
	)
	return m
}

// SetHandler is now deprecated, please use SetRxHandler().
func (m *midi) SetHandler(rxHandler func([]byte)) {
	m.SetRxHandler(rxHandler)
}

// SetRxHandler sets the handler function for incoming MIDI messages.
func (m *midi) SetRxHandler(rxHandler func([]byte)) {
	m.rxHandler = rxHandler
}

// SetTxHandler sets the handler function for outgoing MIDI messages.
func (m *midi) SetTxHandler(txHandler func()) {
	m.txHandler = txHandler
}

func (m *midi) Write(b []byte) (n int, err error) {
	s, e := 0, 0
	for s = 0; s < len(b); s += 4 {
		e = s + 4
		if e > len(b) {
			e = len(b)
		}

		m.tx(b[s:e])
	}
	return e, nil
}

// sendUSBPacket sends a MIDIPacket.
func (m *midi) sendUSBPacket(b []byte) {
	machine.SendUSBInPacket(midiEndpointIn, b)
}

// from BulkIn
func (m *midi) TxHandler() {
	if m.txHandler != nil {
		m.txHandler()
	}

	m.waitTxc = false
	if b, ok := m.buf.Get(); ok {
		m.waitTxc = true
		m.sendUSBPacket(b)
	}
}

func (m *midi) tx(b []byte) {
	if machine.USBDev.InitEndpointComplete {
		if m.waitTxc {
			m.buf.Put(b)
		} else {
			m.waitTxc = true
			m.sendUSBPacket(b)
		}
	}
}

// from BulkOut
func (m *midi) RxHandler(b []byte) {
	if m.rxHandler != nil {
		m.rxHandler(b)
	}
}
