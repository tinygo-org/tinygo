package midi

import (
	"machine"
)

const (
	midiEndpointOut = 5 // from PC
	midiEndpointIn  = 6 // to PC
)

var Midi *midi

type midi struct {
	msg       [4]byte
	buf       *RingBuffer
	rxHandler func([]byte)
	waitTxc   bool
}

func init() {
	if Midi == nil {
		Midi = newMidi()
	}
}

// Port returns the USB midi port.
func Port() *midi {
	return Midi
}

func newMidi() *midi {
	m := &midi{
		buf: NewRingBuffer(),
	}
	machine.EnableMIDI(m.Handler, m.RxHandler, nil)
	return m
}

func (m *midi) SetHandler(rxHandler func([]byte)) {
	m.rxHandler = rxHandler
}

func (m *midi) Write(b []byte) (n int, err error) {
	i := 0
	for i = 0; i < len(b); i += 4 {
		m.tx(b[i : i+4])
	}
	return i, nil
}

// sendUSBPacket sends a MIDIPacket.
func (m *midi) sendUSBPacket(b []byte) {
	machine.SendUSBInPacket(midiEndpointIn, b)
}

// from BulkIn
func (m *midi) Handler() {
	m.waitTxc = false
	if b, ok := m.buf.Get(); ok {
		m.waitTxc = true
		m.sendUSBPacket(b)
	}
}

func (m *midi) tx(b []byte) {
	if m.waitTxc {
		m.buf.Put(b)
	} else {
		m.waitTxc = true
		m.sendUSBPacket(b)
	}
}

// from BulkOut
func (m *midi) RxHandler(b []byte) {
	if m.rxHandler != nil {
		m.rxHandler(b)
	}
}
