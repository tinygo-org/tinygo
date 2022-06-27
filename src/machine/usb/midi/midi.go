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
	buf            *RingBuffer
	callbackFuncRx func([]byte)
	waitTxc        bool
}

func init() {
	if Midi == nil {
		Midi = newMidi()
	}
}

// New returns hid-mouse.
func New() *midi {
	return Midi
}

func newMidi() *midi {
	m := &midi{
		buf: NewRingBuffer(),
	}
	machine.EnableMIDI(m.Callback, m.CallbackRx, nil)
	return m
}

func (m *midi) SetCallback(callbackRx func([]byte)) {
	m.callbackFuncRx = callbackRx
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
func (m *midi) Callback() {
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
func (m *midi) CallbackRx(b []byte) {
	if m.callbackFuncRx != nil {
		m.callbackFuncRx(b)
	}
}
