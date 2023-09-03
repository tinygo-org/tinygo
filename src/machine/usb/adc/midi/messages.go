package midi

import (
	"errors"
)

// From USB-MIDI section 4.1 "Code Index Number (CIN) Classifications"
const (
	CINSystemCommon2   = 0x2
	CINSystemCommon3   = 0x3
	CINSysExStart      = 0x4
	CINSysExEnd1       = 0x5
	CINSysExEnd2       = 0x6
	CINSysExEnd3       = 0x7
	CINNoteOff         = 0x8
	CINNoteOn          = 0x9
	CINPoly            = 0xA
	CINControlChange   = 0xB
	CINProgramChange   = 0xC
	CINChannelPressure = 0xD
	CINPitchBendChange = 0xE
	CINSingleByte      = 0xF
)

// Standard MIDI channel messages
const (
	MsgNoteOff           = 0x80
	MsgNoteOn            = 0x90
	MsgPolyAftertouch    = 0xA0
	MsgControlChange     = 0xB0
	MsgProgramChange     = 0xC0
	MsgChannelAftertouch = 0xD0
	MsgPitchBend         = 0xE0
	MsgSysExStart        = 0xF0
	MsgSysExEnd          = 0xF7
)

// Standard MIDI control change messages
const (
	CCModulationWheel       = 0x01
	CCBreathController      = 0x02
	CCFootPedal             = 0x04
	CCPortamentoTime        = 0x05
	CCDataEntry             = 0x06
	CCVolume                = 0x07
	CCBalance               = 0x08
	CCPan                   = 0x0A
	CCExpression            = 0x0B
	CCEffectControl1        = 0x0C
	CCEffectControl2        = 0x0D
	CCGeneralPurpose1       = 0x10
	CCGeneralPurpose2       = 0x11
	CCGeneralPurpose3       = 0x12
	CCGeneralPurpose4       = 0x13
	CCBankSelect            = 0x20
	CCModulationDepthRange  = 0x21
	CCBreathControllerDepth = 0x22
	CCFootPedalDepth        = 0x24
	CCEffectsLevel          = 0x5B
	CCTremeloLevel          = 0x5C
	CCChorusLevel           = 0x5D
	CCCelesteLevel          = 0x5E
	CCPhaserLevel           = 0x5F
	CCDataIncrement         = 0x60
	CCDataDecrement         = 0x61
	CCNRPNLSB               = 0x62
	CCNRPNMSB               = 0x63
	CCRPNLSB                = 0x64
	CCRPNMSB                = 0x65
	CCAllSoundOff           = 0x78
	CCResetAllControllers   = 0x79
	CCAllNotesOff           = 0x7B
	CCChannelVolume         = 0x7F
)

var (
	errInvalidMIDICable        = errors.New("invalid MIDI cable")
	errInvalidMIDIChannel      = errors.New("invalid MIDI channel")
	errInvalidMIDIVelocity     = errors.New("invalid MIDI velocity")
	errInvalidMIDIControl      = errors.New("invalid MIDI control number")
	errInvalidMIDIControlValue = errors.New("invalid MIDI control value")
	errInvalidMIDIPatch        = errors.New("invalid MIDI patch number")
	errInvalidMIDIPitchBend    = errors.New("invalid MIDI pitch bend value")
	errInvalidMIDISysExData    = errors.New("invalid MIDI SysEx data")
)

// NoteOn sends a channel note on message.
// The cable parameter is the cable number 0-15.
// The channel parameter is the MIDI channel number 1-16.
func (m *midi) NoteOn(cable, channel uint8, note Note, velocity uint8) error {
	switch {
	case cable > 15:
		return errInvalidMIDICable
	case channel == 0 || channel > 16:
		return errInvalidMIDIChannel
	case velocity > 127:
		return errInvalidMIDIVelocity
	}

	m.msg[0], m.msg[1], m.msg[2], m.msg[3] = (cable&0xf<<4)|CINNoteOn, MsgNoteOn|(channel-1&0xf), byte(note)&0x7f, velocity&0x7f
	_, err := m.Write(m.msg[:])
	return err
}

// NoteOff sends a channel note off message.
// The cable parameter is the cable number 0-15.
// The channel parameter is the MIDI channel number 1-16.
func (m *midi) NoteOff(cable, channel uint8, note Note, velocity uint8) error {
	switch {
	case cable > 15:
		return errInvalidMIDICable
	case channel == 0 || channel > 16:
		return errInvalidMIDIChannel
	case velocity > 127:
		return errInvalidMIDIVelocity
	}

	m.msg[0], m.msg[1], m.msg[2], m.msg[3] = (cable&0xf<<4)|CINNoteOff, MsgNoteOff|(channel-1&0xf), byte(note)&0x7f, velocity&0x7f
	_, err := m.Write(m.msg[:])
	return err
}

// ControlChange sends a channel continuous controller message.
// The cable parameter is the cable number 0-15.
// The channel parameter is the MIDI channel number 1-16.
// The control parameter is the controller number 0-127.
// The value parameter is the controller value 0-127.
func (m *midi) ControlChange(cable, channel, control, value uint8) error {
	switch {
	case cable > 15:
		return errInvalidMIDICable
	case channel == 0 || channel > 16:
		return errInvalidMIDIChannel
	case control > 127:
		return errInvalidMIDIControl
	case value > 127:
		return errInvalidMIDIControlValue
	}

	m.msg[0], m.msg[1], m.msg[2], m.msg[3] = (cable&0xf<<4)|CINControlChange, MsgControlChange|(channel-1&0xf), control&0x7f, value&0x7f
	_, err := m.Write(m.msg[:])
	return err
}

// ProgramChange sends a channel program change message.
// The cable parameter is the cable number 0-15.
// The channel parameter is the MIDI channel number 1-16.
// The patch parameter is the program number 0-127.
func (m *midi) ProgramChange(cable, channel uint8, patch uint8) error {
	switch {
	case cable > 15:
		return errInvalidMIDICable
	case channel == 0 || channel > 16:
		return errInvalidMIDIChannel
	case patch > 127:
		return errInvalidMIDIPatch
	}

	m.msg[0], m.msg[1], m.msg[2] = (cable&0xf<<4)|CINProgramChange, MsgProgramChange|(channel-1&0xf), patch&0x7f
	_, err := m.Write(m.msg[:3])
	return err
}

// PitchBend sends a channel pitch bend message.
// The cable parameter is the cable number 0-15.
// The channel parameter is the MIDI channel number 1-16.
// The bend parameter is the 14-bit pitch bend value (maximum 0x3FFF).
// Setting bend above 0x2000 (up to 0x3FFF) will increase the pitch.
// Setting bend below 0x2000 (down to 0x0000) will decrease the pitch.
func (m *midi) PitchBend(cable, channel uint8, bend uint16) error {
	switch {
	case cable > 15:
		return errInvalidMIDICable
	case channel == 0 || channel > 16:
		return errInvalidMIDIChannel
	case bend > 0x3FFF:
		return errInvalidMIDIPitchBend
	}

	m.msg[0], m.msg[1], m.msg[2], m.msg[3] = (cable&0xf<<4)|CINPitchBendChange, MsgPitchBend|(channel-1&0xf), byte(bend&0x7f), byte(bend>>8)&0x7f
	_, err := m.Write(m.msg[:])
	return err
}

// SysEx sends a System Exclusive message.
// The cable parameter is the cable number 0-15.
// The data parameter is a slice with the data to send.
// It needs to start with the manufacturer ID, which is either
// 1 or 3 bytes in length.
// The data slice should not include the SysEx start (0xF0) or
// end (0xF7) bytes, only the data in between.
func (m *midi) SysEx(cable uint8, data []byte) error {
	switch {
	case cable > 15:
		return errInvalidMIDICable
	case len(data) < 3:
		return errInvalidMIDISysExData
	}

	// write start
	m.msg[0], m.msg[1] = (cable&0xf<<4)|CINSysExStart, MsgSysExStart
	m.msg[2], m.msg[3] = data[0], data[1]
	if _, err := m.Write(m.msg[:]); err != nil {
		return err
	}

	// write middle
	i := 2
	for ; i < len(data)-2; i += 3 {
		m.msg[0], m.msg[1] = (cable&0xf<<4)|CINSysExStart, data[i]
		m.msg[2], m.msg[3] = data[i+1], data[i+2]
		if _, err := m.Write(m.msg[:]); err != nil {
			return err
		}
	}
	// write end
	switch len(data) - i {
	case 2:
		m.msg[0], m.msg[1] = (cable&0xf<<4)|CINSysExEnd3, data[i]
		m.msg[2], m.msg[3] = data[i+1], MsgSysExEnd
	case 1:
		m.msg[0], m.msg[1] = (cable&0xf<<4)|CINSysExEnd2, data[i]
		m.msg[2], m.msg[3] = MsgSysExEnd, 0
	case 0:
		m.msg[0], m.msg[1] = (cable&0xf<<4)|CINSysExEnd1, MsgSysExEnd
		m.msg[2], m.msg[3] = 0, 0
	}
	if _, err := m.Write(m.msg[:]); err != nil {
		return err
	}

	return nil
}
