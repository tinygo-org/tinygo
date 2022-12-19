//go:build sam

// This is the definition for I2S bus functions.
// Actual implementations if available for any given hardware
// are to be found in its the board definition.
//
// For more info about I2S, see: https://en.wikipedia.org/wiki/I%C2%B2S
//

package machine

type I2SMode uint8
type I2SStandard uint8
type I2SClockSource uint8
type I2SDataFormat uint8

const (
	I2SModeSource I2SMode = iota
	I2SModeReceiver
	I2SModePDM
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

// All fields are optional and may not be required or used on a particular platform.
type I2SConfig struct {
	SCK             Pin
	WS              Pin
	SD              Pin
	Mode            I2SMode
	Standard        I2SStandard
	ClockSource     I2SClockSource
	DataFormat      I2SDataFormat
	AudioFrequency  uint32
	MainClockOutput bool
	Stereo          bool
}
