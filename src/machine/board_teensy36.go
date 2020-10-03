// +build nxp,mk66f18,teensy36

package machine

// CPUFrequency returns the frequency of the ARM core clock (180MHz)
func CPUFrequency() uint32 { return 180000000 }

// ClockFrequency returns the frequency of the external oscillator (16MHz)
func ClockFrequency() uint32 { return 16000000 }

const (
	usb_STRING_MANUFACTURER = "Teensy"
	usb_STRING_PRODUCT      = "Teensy 3.6 USB Serial"

	usb_DEVICE_CLASS    = 0x02
	usb_DEVICE_SUBCLASS = 0
	usb_DEVICE_PROTOCOL = 0

	usb_EP0_SIZE    = 64
	usb_BUFFER_SIZE = usb_EP0_SIZE

	usb_VID        = 0x16C0
	usb_PID        = 0x0483 // teensy usb serial
	usb_BCD_DEVICE = 0x0277 // teensy 3.6
)

var usb_STRING_SERIAL string

// LED on the Teensy
const LED = PC05

// digital IO
const (
	D00 = PB16
	D01 = PB17
	D02 = PD00
	D03 = PA12
	D04 = PA13
	D05 = PD07
	D06 = PD04
	D07 = PD02
	D08 = PD03
	D09 = PC03
	D10 = PC04
	D11 = PC06
	D12 = PC07
	D13 = PC05
	D14 = PD01
	D15 = PC00
	D16 = PB00
	D17 = PB01
	D18 = PB03
	D19 = PB02
	D20 = PD05
	D21 = PD06
	D22 = PC01
	D23 = PC02
	D24 = PE26
	D25 = PA05
	D26 = PA14
	D27 = PA15
	D28 = PA16
	D29 = PB18
	D30 = PB19
	D31 = PB10
	D32 = PB11
	D33 = PE24
	D34 = PE25
	D35 = PC08
	D36 = PC09
	D37 = PC10
	D38 = PC11
	D39 = PA17
	D40 = PA28
	D41 = PA29
	D42 = PA26
	D43 = PB20
	D44 = PB22
	D45 = PB23
	D46 = PB21
	D47 = PD08
	D48 = PD09
	D49 = PB04
	D50 = PB05
	D51 = PD14
	D52 = PD13
	D53 = PD12
	D54 = PD15
	D55 = PD11
	D56 = PE10
	D57 = PE11
	D58 = PE00
	D59 = PE01
	D60 = PE02
	D61 = PE03
	D62 = PE04
	D63 = PE05
)

var (
	TeensyUART1 = &UART0
	TeensyUART2 = &UART1
	TeensyUART3 = &UART2
	TeensyUART4 = &UART3
	TeensyUART5 = &UART4
)

const (
	defaultUART0RX = D00
	defaultUART0TX = D01
	defaultUART1RX = D09
	defaultUART1TX = D10
	defaultUART2RX = D07
	defaultUART2TX = D08
	defaultUART3RX = D31
	defaultUART3TX = D32
	defaultUART4RX = D34
	defaultUART4TX = D33
)

//go:linkname sleepTicks runtime.sleepTicks
func sleepTicks(int64)

func InitPlatform() {
	// for background about this startup delay, please see these conversations
	// https://forum.pjrc.com/threads/36606-startup-time-(400ms)?p=113980&viewfull=1#post113980
	// https://forum.pjrc.com/threads/31290-Teensey-3-2-Teensey-Loader-1-24-Issues?p=87273&viewfull=1#post87273

	sleepTicks(25 * 1000) // 50 for TeensyDuino < 142
	// usb_STRING_SERIAL = readSerialNumber()
	USB0.Configure()
	sleepTicks(275 * 1000) // 350 for TeensyDuino < 142
}
