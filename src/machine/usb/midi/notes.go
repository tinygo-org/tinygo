package midi

// Note represents a MIDI note number. For example, Note(69) is A4 or 440Hz.
type Note uint8

// Define all the notes in a format similar to the Tone library in the Arduino
// IDE.
const (
	A0 Note = iota + 21 // 27.5Hz
	AS0
	B0
	C1
	CS1
	D1
	DS1
	E1
	F1
	FS1
	G1
	GS1
	A1 // 55Hz
	AS1
	B1
	C2
	CS2
	D2
	DS2
	E2
	F2
	FS2
	G2
	GS2
	A2 // 110Hz
	AS2
	B2
	C3
	CS3
	D3
	DS3
	E3
	F3
	FS3
	G3
	GS3
	A3 // 220Hz
	AS3
	B3
	C4
	CS4
	D4
	DS4
	E4
	F4
	FS4
	G4
	GS4
	A4 // 440Hz
	AS4
	B4
	C5
	CS5
	D5
	DS5
	E5
	F5
	FS5
	G5
	GS5
	A5 // 880Hz
	AS5
	B5
	C6
	CS6
	D6
	DS6
	E6
	F6
	FS6
	G6
	GS6
	A6 // 1760Hz
	AS6
	B6
	C7
	CS7
	D7
	DS7
	E7
	F7
	FS7
	G7
	GS7
	A7 // 3520Hz
	AS7
	B7
	C8
	CS8
	D8
	DS8
	E8
	F8
	FS8
	G8
	GS8
	A8 // 7040Hz
	AS8
	B8
)
