// +build stm32h7

package machine

// Peripheral abstraction layer for the STM32H7 which contains definitions that
// are common or shared between both cores (Cortex-M7 and -M4) of the MCU.
//
// Constructs that are unique to one of the two cores are defined in the
// respective source file:
//   machine_stm32h7x7_cm7.go - Cortex-M7 core (primary)
//   machine_stm32h7x7_cm4.go - Cortex-M4 core

import (
	"device/stm32"
	"runtime/volatile"
	"unsafe"
)

const PinsPerPort = 16

const (
	portA Pin = iota * PinsPerPort
	portB
	portC
	portD
	portE
	portF
	portG
	portH
	portI
	portJ
	portK
)

const (
	PA00 = portA + 0  // 0x00
	PA01 = portA + 1  // 0x01
	PA02 = portA + 2  // 0x02
	PA03 = portA + 3  // 0x03
	PA04 = portA + 4  // 0x04
	PA05 = portA + 5  // 0x05
	PA06 = portA + 6  // 0x06
	PA07 = portA + 7  // 0x07
	PA08 = portA + 8  // 0x08
	PA09 = portA + 9  // 0x09
	PA10 = portA + 10 // 0x0A
	PA11 = portA + 11 // 0x0B
	PA12 = portA + 12 // 0x0C
	PA13 = portA + 13 // 0x0D
	PA14 = portA + 14 // 0x0E
	PA15 = portA + 15 // 0x0F

	PB00 = portB + 0  // 0x10
	PB01 = portB + 1  // 0x11
	PB02 = portB + 2  // 0x12
	PB03 = portB + 3  // 0x13
	PB04 = portB + 4  // 0x14
	PB05 = portB + 5  // 0x15
	PB06 = portB + 6  // 0x16
	PB07 = portB + 7  // 0x17
	PB08 = portB + 8  // 0x18
	PB09 = portB + 9  // 0x19
	PB10 = portB + 10 // 0x1A
	PB11 = portB + 11 // 0x1B
	PB12 = portB + 12 // 0x1C
	PB13 = portB + 13 // 0x1D
	PB14 = portB + 14 // 0x1E
	PB15 = portB + 15 // 0x1F

	PC00 = portC + 0  // 0x20
	PC01 = portC + 1  // 0x21
	PC02 = portC + 2  // 0x22
	PC03 = portC + 3  // 0x23
	PC04 = portC + 4  // 0x24
	PC05 = portC + 5  // 0x25
	PC06 = portC + 6  // 0x26
	PC07 = portC + 7  // 0x27
	PC08 = portC + 8  // 0x28
	PC09 = portC + 9  // 0x29
	PC10 = portC + 10 // 0x2A
	PC11 = portC + 11 // 0x2B
	PC12 = portC + 12 // 0x2C
	PC13 = portC + 13 // 0x2D
	PC14 = portC + 14 // 0x2E
	PC15 = portC + 15 // 0x2F

	PD00 = portD + 0  // 0x30
	PD01 = portD + 1  // 0x31
	PD02 = portD + 2  // 0x32
	PD03 = portD + 3  // 0x33
	PD04 = portD + 4  // 0x34
	PD05 = portD + 5  // 0x35
	PD06 = portD + 6  // 0x36
	PD07 = portD + 7  // 0x37
	PD08 = portD + 8  // 0x38
	PD09 = portD + 9  // 0x39
	PD10 = portD + 10 // 0x3A
	PD11 = portD + 11 // 0x3B
	PD12 = portD + 12 // 0x3C
	PD13 = portD + 13 // 0x3D
	PD14 = portD + 14 // 0x3E
	PD15 = portD + 15 // 0x3F

	PE00 = portE + 0  // 0x40
	PE01 = portE + 1  // 0x41
	PE02 = portE + 2  // 0x42
	PE03 = portE + 3  // 0x43
	PE04 = portE + 4  // 0x44
	PE05 = portE + 5  // 0x45
	PE06 = portE + 6  // 0x46
	PE07 = portE + 7  // 0x47
	PE08 = portE + 8  // 0x48
	PE09 = portE + 9  // 0x49
	PE10 = portE + 10 // 0x4A
	PE11 = portE + 11 // 0x4B
	PE12 = portE + 12 // 0x4C
	PE13 = portE + 13 // 0x4D
	PE14 = portE + 14 // 0x4E
	PE15 = portE + 15 // 0x4F

	PF00 = portF + 0  // 0x50
	PF01 = portF + 1  // 0x51
	PF02 = portF + 2  // 0x52
	PF03 = portF + 3  // 0x53
	PF04 = portF + 4  // 0x54
	PF05 = portF + 5  // 0x55
	PF06 = portF + 6  // 0x56
	PF07 = portF + 7  // 0x57
	PF08 = portF + 8  // 0x58
	PF09 = portF + 9  // 0x59
	PF10 = portF + 10 // 0x5A
	PF11 = portF + 11 // 0x5B
	PF12 = portF + 12 // 0x5C
	PF13 = portF + 13 // 0x5D
	PF14 = portF + 14 // 0x5E
	PF15 = portF + 15 // 0x5F

	PG00 = portG + 0  // 0x60
	PG01 = portG + 1  // 0x61
	PG02 = portG + 2  // 0x62
	PG03 = portG + 3  // 0x63
	PG04 = portG + 4  // 0x64
	PG05 = portG + 5  // 0x65
	PG06 = portG + 6  // 0x66
	PG07 = portG + 7  // 0x67
	PG08 = portG + 8  // 0x68
	PG09 = portG + 9  // 0x69
	PG10 = portG + 10 // 0x6A
	PG11 = portG + 11 // 0x6B
	PG12 = portG + 12 // 0x6C
	PG13 = portG + 13 // 0x6D
	PG14 = portG + 14 // 0x6E
	PG15 = portG + 15 // 0x6F

	PH00 = portH + 0  // 0x70
	PH01 = portH + 1  // 0x71
	PH02 = portH + 2  // 0x72
	PH03 = portH + 3  // 0x73
	PH04 = portH + 4  // 0x74
	PH05 = portH + 5  // 0x75
	PH06 = portH + 6  // 0x76
	PH07 = portH + 7  // 0x77
	PH08 = portH + 8  // 0x78
	PH09 = portH + 9  // 0x79
	PH10 = portH + 10 // 0x7A
	PH11 = portH + 11 // 0x7B
	PH12 = portH + 12 // 0x7C
	PH13 = portH + 13 // 0x7D
	PH14 = portH + 14 // 0x7E
	PH15 = portH + 15 // 0x7F

	PI00 = portI + 0  // 0x80
	PI01 = portI + 1  // 0x81
	PI02 = portI + 2  // 0x82
	PI03 = portI + 3  // 0x83
	PI04 = portI + 4  // 0x84
	PI05 = portI + 5  // 0x85
	PI06 = portI + 6  // 0x86
	PI07 = portI + 7  // 0x87
	PI08 = portI + 8  // 0x88
	PI09 = portI + 9  // 0x89
	PI10 = portI + 10 // 0x8A
	PI11 = portI + 11 // 0x8B
	PI12 = portI + 12 // 0x8C
	PI13 = portI + 13 // 0x8D
	PI14 = portI + 14 // 0x8E
	PI15 = portI + 15 // 0x8F

	PJ00 = portJ + 0  // 0x90
	PJ01 = portJ + 1  // 0x91
	PJ02 = portJ + 2  // 0x92
	PJ03 = portJ + 3  // 0x93
	PJ04 = portJ + 4  // 0x94
	PJ05 = portJ + 5  // 0x95
	PJ06 = portJ + 6  // 0x96
	PJ07 = portJ + 7  // 0x97
	PJ08 = portJ + 8  // 0x98
	PJ09 = portJ + 9  // 0x99
	PJ10 = portJ + 10 // 0x9A
	PJ11 = portJ + 11 // 0x9B
	PJ12 = portJ + 12 // 0x9C
	PJ13 = portJ + 13 // 0x9D
	PJ14 = portJ + 14 // 0x9E
	PJ15 = portJ + 15 // 0x9F

	PK00 = portK + 0  // 0xA0
	PK01 = portK + 1  // 0xA1
	PK02 = portK + 2  // 0xA2
	PK03 = portK + 3  // 0xA3
	PK04 = portK + 4  // 0xA4
	PK05 = portK + 5  // 0xA5
	PK06 = portK + 6  // 0xA6
	PK07 = portK + 7  // 0xA7
	PK08 = portK + 8  // 0xA8
	PK09 = portK + 9  // 0xA9
	PK10 = portK + 10 // 0xAA
	PK11 = portK + 11 // 0xAB
	PK12 = portK + 12 // 0xAC
	PK13 = portK + 13 // 0xAD
	PK14 = portK + 14 // 0xAE
	PK15 = portK + 15 // 0xAF
)

func (p Pin) Port() Pin { return PinsPerPort * ((p >> 4) & 0xF) } // high nybble
func (p Pin) Bit() Pin  { return p & 0xF }                        // low nybble

func (p Pin) Bus() *stm32.GPIO_Type {
	o := p.Port()
	switch o {
	case portA:
		return stm32.GPIOA
	case portB:
		return stm32.GPIOB
	case portC:
		return stm32.GPIOC
	case portD:
		return stm32.GPIOD
	case portE:
		return stm32.GPIOE
	case portF:
		return stm32.GPIOF
	case portG:
		return stm32.GPIOG
	case portH:
		return stm32.GPIOH
	case portI:
		return stm32.GPIOI
	case portJ:
		return stm32.GPIOJ
	case portK:
		return stm32.GPIOK
	}
	return nil
}

type PinMode uint32

// Constant definitions for the most common pin configurations and alternate
// functions. Each of these PinMode values are valid and supported
// configurations (if permitted by the intended GPIO pin).
const (
	PinInput           = pinModeModeInput | pinModePupdNoPull
	PinInputAnalog     = pinModeModeInput | pinModeModeAnalog
	PinOutput          = PinOutputPushPull
	PinOutputPushPull  = pinModeModeOutput | pinModeOTypePushPull | pinModePupdNoPull
	PinOutputOpenDrain = pinModeModeOutput | pinModeOTypeOpenDrain | pinModePupdPullUp
	PinOutputAnalog    = pinModeModeOutput | pinModeModeAnalog
)

// Constant definitions for common attributes that can be applied (bitwise-OR)
// to any PinMode. Of course, not every combination is valid or supported.
const (
	PinFloating      = pinModePupdNoPull
	PinPullUp        = pinModePupdPullUp
	PinPullDown      = pinModePupdPullDown
	PinPushPull      = pinModeOTypePushPull
	PinOpenDrain     = pinModeOTypeOpenDrain
	PinLowSpeed      = pinModeSpeedLow
	PinMediumSpeed   = pinModeSpeedMedium
	PinHighSpeed     = pinModeSpeedHigh
	PinVeryHighSpeed = pinModeSpeedVeryHigh
	PinAltFunc       = pinModeModeAltFunc
	PinAnalog        = pinModeModeAnalog
)

// Internally we encode pin configuration in type PinMode (32-bit) as follows:
//
//    Bits:  1 0 9 8 7 6 5 4 3 2 1 0 9 8 7 6 5 4 3 2 1 0 9 8 7 6 5 4 3 2 1 0
//   Field:  - - - - - - - - - - - - C I H_H_H_H A_A_A_A - - S_S P_P O - F_F
//
//    Bits   Field             Description (GPIO Register)
//   ------- ----- ------------------------------------------------------
//   [ 0:1 ]   F   Function (MODER): Input / Output / Alt / Analog
//   [   2 ]   -   (Reserved)
//   [   3 ]   O   Output (OTYPER): Push-Pull / Open-Drain
//   [ 4:5 ]   P   Pull (PUPDR): No-Pull / Pull-Up / Pull-Down
//   [ 6:7 ]   S   Speed (OSPEEDR)
//   [ 8:9 ]   -   (Reserved)
//   [10:13]   A   Alternate Function (AFRL/AFRH)
//   [14:17]   H   Channel (Analog/Timer specific)
//   [   18]   I   Inverted (Analog/Timer specific)
//   [   19]   C   Analog ADC control
//   [20:31]   -   (Reserved)
//
// The following bitmasks may be combined together (bitwise-OR) to define a pin
// configuration with each of the specified attributes. Of course, not every
// combination is valid or supported.
const (
	pinModeModePos     PinMode = 0                   //
	pinModeModeMsk     PinMode = 0x03                //
	pinModeModeInput   PinMode = 0 << pinModeModePos //
	pinModeModeOutput  PinMode = 1 << pinModeModePos //
	pinModeModeAltFunc PinMode = 2 << pinModeModePos // PA13-PA15, PB04, PB03
	pinModeModeAnalog  PinMode = 3 << pinModeModePos // Default for all others

	pinModeOTypePos       PinMode = 3
	pinModeOTypeMsk       PinMode = 0x01
	pinModeOTypePushPull  PinMode = 0 << pinModeOTypePos
	pinModeOTypeOpenDrain PinMode = 1 << pinModeOTypePos

	pinModePupdPos      PinMode = 4
	pinModePupdMsk      PinMode = 0x03
	pinModePupdNoPull   PinMode = 0 << pinModePupdPos
	pinModePupdPullUp   PinMode = 1 << pinModePupdPos
	pinModePupdPullDown PinMode = 2 << pinModePupdPos

	pinModeSpeedPos      PinMode = 6                    //
	pinModeSpeedMsk      PinMode = 0x03                 //
	pinModeSpeedLow      PinMode = 0 << pinModeSpeedPos // Default for all others
	pinModeSpeedMedium   PinMode = 1 << pinModeSpeedPos //
	pinModeSpeedHigh     PinMode = 2 << pinModeSpeedPos //
	pinModeSpeedVeryHigh PinMode = 3 << pinModeSpeedPos // PA13, PB03

	pinModeAltFuncPos PinMode = 10
	pinModeAltFuncMsk PinMode = 0x0F

	pinModeAChanPos PinMode = 14
	pinModeAChanMsk PinMode = 0x0F

	pinModeInvPos      PinMode = 18
	pinModeInvMsk      PinMode = 0x01
	pinModeInvInverted PinMode = 1 << pinModeInvPos

	pinModeACtrlPos     PinMode = 19
	pinModeACtrlMsk     PinMode = 0x01
	pinModeACtrlControl PinMode = 1 << pinModeACtrlPos
)

func (p Pin) Configure(config PinConfig) {

	bus := p.Bus()
	bit := uint8(p.Bit())

	_ = EnableClock(unsafe.Pointer(bus), true)

	switch config.Mode & (pinModeModeMsk << pinModeModePos) {
	// output or alternate function mode bits set
	case pinModeModeOutput, pinModeModeAltFunc:
		// configure output speed
		speed := uint32((config.Mode >> pinModeSpeedPos) & pinModeSpeedMsk)
		bus.GPIO_OSPEEDR.ReplaceBits(speed, uint32(pinModeSpeedMsk), 2*bit)
		// configure output type
		otype := uint32((config.Mode >> pinModeOTypePos) & pinModeOTypeMsk)
		bus.GPIO_OTYPER.ReplaceBits(otype, uint32(pinModeOTypeMsk), bit)
	}

	// configure pull-up/down resistor
	pupd := uint32((config.Mode >> pinModePupdPos) & pinModePupdMsk)
	bus.GPIO_PUPDR.ReplaceBits(pupd, uint32(pinModePupdMsk), 2*bit)

	switch config.Mode & (pinModeModeMsk << pinModeModePos) {
	// alternate function mode bit set
	case pinModeModeAltFunc:
		var reg *volatile.Register32
		var pos uint8
		if bit < 8 {
			reg = &bus.GPIO_AFRL // AFRL register used for pins 0-7
		} else {
			reg = &bus.GPIO_AFRH // AFRH register used for pins 8-15
		}
		altf := uint32((config.Mode >> pinModeAltFuncPos) & pinModeAltFuncMsk)
		reg.ReplaceBits(altf, uint32(pinModeAltFuncMsk), 4*(pos%8))
	}

	// configure IO direction mode
	mode := uint32((config.Mode >> pinModeModePos) & pinModeModeMsk)
	bus.GPIO_MODER.ReplaceBits(mode, uint32(pinModeModeMsk), 2*bit)
}

func (p Pin) Toggle() {
	p.Set(!p.Bus().GPIO_ODR.HasBits(1 << p.Bit()))
}

func (p Pin) Set(high bool) {
	if high {
		p.Bus().GPIO_BSRR.Set(1 << p.Bit())
	} else {
		p.Bus().GPIO_BSRR.Set(1 << (p.Bit() + 16))
	}
}

func (p Pin) Get() bool {
	return p.Bus().GPIO_IDR.HasBits(1 << p.Bit())
}
