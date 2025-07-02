//go:build sam && atsamd21

// This is the definition for I2S bus functions.
// Actual implementations if available for any given hardware
// are to be found in its the board definition.
//
// For more info about I2S, see: https://en.wikipedia.org/wiki/I%C2%B2S
//

package machine

import "errors"

// If you are getting a compile error on this line please check to see you've
// correctly implemented the methods on the I2S type. They must match
// the interface method signatures type to type perfectly.
// If not implementing the I2S type please remove your target from the build tags
// at the top of this file.
var _ interface {
	SetSampleFrequency(freq uint32) error
	ReadMono(b []uint16) (int, error)
	ReadStereo(b []uint32) (int, error)
	WriteMono(b []uint16) (int, error)
	WriteStereo(b []uint32) (int, error)
	Enable(enabled bool)
} = (*I2S)(nil)

type I2SMode uint8
type I2SStandard uint8
type I2SClockSource uint8
type I2SDataFormat uint8

const (
	I2SModeSource I2SMode = iota
	I2SModeReceiver
	I2SModePDM
	I2SModeSourceReceiver
)

const (
	I2StandardPhilips I2SStandard = iota
	I2SStandardMSB
	I2SStandardLSB
)

const (
	I2SClockSourceInternal I2SClockSource = iota
	I2SClockSourceExternal
)

const (
	I2SDataFormatDefault I2SDataFormat = 0
	I2SDataFormat8bit                  = 8
	I2SDataFormat16bit                 = 16
	I2SDataFormat24bit                 = 24
	I2SDataFormat32bit                 = 32
)

var (
	ErrInvalidSampleFrequency = errors.New("i2s: invalid sample frequency")
)

// All fields are optional and may not be required or used on a particular platform.
type I2SConfig struct {
	// clock
	SCK Pin
	// word select
	WS Pin
	// data out
	SDO Pin
	// data in
	SDI             Pin
	Mode            I2SMode
	Standard        I2SStandard
	ClockSource     I2SClockSource
	DataFormat      I2SDataFormat
	AudioFrequency  uint32
	MainClockOutput bool
	Stereo          bool
}
