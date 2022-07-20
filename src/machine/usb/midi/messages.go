package midi

// NoteOn sends a note on message.
func (m *midi) NoteOn(cable, channel, note, velocity uint8) {
	m.msg[0], m.msg[1], m.msg[2], m.msg[3] = (cable&0xf<<4)|0x9, 0x90|(channel&0xf), note&0x7f, velocity&0x7f
	m.Write(m.msg[:])
}

// NoteOff sends a note off message.
func (m *midi) NoteOff(cable, channel, note, velocity uint8) {
	m.msg[0], m.msg[1], m.msg[2], m.msg[3] = (cable&0xf<<4)|0x8, 0x80|(channel&0xf), note&0x7f, velocity&0x7f
	m.Write(m.msg[:])
}

// SendCC sends a continuous controller message.
func (m *midi) SendCC(cable, channel, control, value uint8) {
	m.msg[0], m.msg[1], m.msg[2], m.msg[3] = (cable&0xf<<4)|0xB, 0xB0|(channel&0xf), control&0x7f, value&0x7f
	m.Write(m.msg[:])
}
