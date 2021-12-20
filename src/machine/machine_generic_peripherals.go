// +build !baremetal,!arduino_mkr1000,!arduino_mkrwifi1010,!arduino_nano33,!arduino_zero,!circuitplay_express,!feather_m0,!feather_m4,!grandcentral_m4,!itsybitsy_m0,!itsybitsy_m4,!matrixportal_m4,!metro_m4_airlift,!p1am_100,!pybadge,!pygamer,!pyportal,!qtpy,!trinket_m0,!wioterminal,!xiao

package machine

// These peripherals are defined separately so that they can be excluded on
// boards that define their peripherals in the board file (e.g. board_qtpy.go).

var (
	SPI0 = SPI{0}
	I2C0 = &I2C{0}
)
