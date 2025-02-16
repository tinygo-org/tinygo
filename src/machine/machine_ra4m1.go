//go:build ra4m1

package machine

import (
	"device/renesas"
	"runtime/volatile"
	"unsafe"
)

const deviceName = renesas.Device

const (
	P0_00 Pin = 0
	P0_01 Pin = 1
	P0_02 Pin = 2
	P0_03 Pin = 3
	P0_04 Pin = 4
	P0_05 Pin = 5
	P0_06 Pin = 6
	P0_07 Pin = 7
	P0_08 Pin = 8
	P0_09 Pin = 9
	P0_10 Pin = 10
	P0_11 Pin = 11
	P0_12 Pin = 12
	P0_13 Pin = 13
	P0_14 Pin = 14
	P0_15 Pin = 15
	P1_00 Pin = 16
	P1_01 Pin = 17
	P1_02 Pin = 18
	P1_03 Pin = 19
	P1_04 Pin = 20
	P1_05 Pin = 21
	P1_06 Pin = 22
	P1_07 Pin = 23
	P1_08 Pin = 24
	P1_09 Pin = 25
	P1_10 Pin = 26
	P1_11 Pin = 27
	P1_12 Pin = 28
	P1_13 Pin = 29
	P1_14 Pin = 30
	P1_15 Pin = 31
	P2_00 Pin = 32
	P2_01 Pin = 33
	P2_02 Pin = 34
	P2_03 Pin = 35
	P2_04 Pin = 36
	P2_05 Pin = 37
	P2_06 Pin = 38
	P2_07 Pin = 39
	P2_08 Pin = 40
	P2_09 Pin = 41
	P2_10 Pin = 42
	P2_11 Pin = 43
	P2_12 Pin = 44
	P2_13 Pin = 45
	P2_14 Pin = 46
	P2_15 Pin = 47
	P3_00 Pin = 48
	P3_01 Pin = 49
	P3_02 Pin = 50
	P3_03 Pin = 51
	P3_04 Pin = 52
	P3_05 Pin = 53
	P3_06 Pin = 54
	P3_07 Pin = 55
	P3_08 Pin = 56
	P3_09 Pin = 57
	P3_10 Pin = 58
	P3_11 Pin = 59
	P3_12 Pin = 60
	P3_13 Pin = 61
	P3_14 Pin = 62
	P3_15 Pin = 63
	P4_00 Pin = 64
	P4_01 Pin = 65
	P4_02 Pin = 66
	P4_03 Pin = 67
	P4_04 Pin = 68
	P4_05 Pin = 69
	P4_06 Pin = 70
	P4_07 Pin = 71
	P4_08 Pin = 72
	P4_09 Pin = 73
	P4_10 Pin = 74
	P4_11 Pin = 75
	P4_12 Pin = 76
	P4_13 Pin = 77
	P4_14 Pin = 78
	P4_15 Pin = 79
	P5_00 Pin = 80
	P5_01 Pin = 81
	P5_02 Pin = 82
	P5_03 Pin = 83
	P5_04 Pin = 84
	P5_05 Pin = 85
	P5_06 Pin = 86
	P5_07 Pin = 87
	P5_08 Pin = 88
	P5_09 Pin = 89
	P5_10 Pin = 90
	P5_11 Pin = 91
	P5_12 Pin = 92
	P5_13 Pin = 93
	P5_14 Pin = 94
	P5_15 Pin = 95
)

const (
	PinOutput PinMode = iota
	PinInput
	PinInputPullUp
	PinInputPullDown
)

// Configure configures the gpio pin as per mode.
func (p Pin) Configure(config PinConfig) {
	if p == NoPin {
		return
	}

	enableWritingPmnPFS()

	switch config.Mode {
	case PinOutput:
		setPDR(p, renesas.PFS_P000PFS_PDR_1)
	default:
		setPDR(p, renesas.PFS_P000PFS_PDR_0)
	}

	disableWritingPmnPFS()
}

// Set drives the pin high if value is true else drives it low.
func (p Pin) Set(value bool) {
	if p == NoPin {
		return
	}

	port := getPort(p)
	if port == nil {
		return
	}

	if value == true {
		port.set(p)
	} else {
		port.clr(p)
	}
}

// Get reads the pin value.
func (p Pin) Get() bool {
	return false
}

var (
	gpioPort0 = gpioPortType0{renesas.PORT0}
	gpioPort1 = gpioPortType1{renesas.PORT1}
	gpioPort2 = gpioPortType1{renesas.PORT2}
	gpioPort3 = gpioPortType1{renesas.PORT3}
	gpioPort4 = gpioPortType1{renesas.PORT4}
	gpioPort5 = gpioPortType0{renesas.PORT5}
)

func getPort(p Pin) gpioPort {
	switch getPortNumber(p) {
	case 0:
		return gpioPort0
	case 1:
		return gpioPort1
	case 2:
		return gpioPort2
	case 3:
		return gpioPort3
	case 4:
		return gpioPort4
	case 5:
		return gpioPort5
	default:
		return nil
	}
}

func getPortNumber(p Pin) int {
	return int(p) >> 8
}

type gpioPort interface {
	set(p Pin)
	clr(p Pin)
}

type gpioPortType0 struct {
	port *renesas.PORT0_Type
}

func (p gpioPortType0) set(pin Pin) {
	p.port.SetPCNTR3_PORR(1 << pin)
}

func (p gpioPortType0) clr(pin Pin) {
	p.port.SetPCNTR3_POSR(1 << pin)
}

type gpioPortType1 struct {
	port *renesas.PORT1_Type
}

func (p gpioPortType1) set(pin Pin) {
	p.port.SetPCNTR3_PORR(1 << pin)
}

func (p gpioPortType1) clr(pin Pin) {
	p.port.SetPCNTR3_POSR(1 << pin)
}

func enableWritingPmnPFS() {
	renesas.PMISC.PWPR.Set(0x0)
	renesas.PMISC.SetPWPR_PFSWE(0x1)
}

func disableWritingPmnPFS() {
	renesas.PMISC.PWPR.Set(0x0)
	renesas.PMISC.SetPWPR_B0WI(0x1)
}

func setPDR(p Pin, value uint32) {
	port := getPortNumber(p)
	bit := int(p) & 0xff
	reg := (*volatile.Register32)(unsafe.Add(unsafe.Pointer(uintptr(0x40040800)), 0x40*port+4*bit))
	reg.SetBits(value << renesas.PFS_P000PFS_PDR_Pos)
}
