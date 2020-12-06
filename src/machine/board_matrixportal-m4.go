// +build matrixportal_m4

package machine

import (
	"device/arm"
	"device/sam"
	"runtime/interrupt"
	"runtime/volatile"
)

// used to reset into bootloader
const RESET_MAGIC_VALUE = 0xF01669EF

// Digital pins
const (
	//    Pin   // Function          SERCOM  PWM  Interrupt
	//    ----  // ----------------  ------  ---  ---------
	D0  = PA01  // UART RX            1[1]   PWM  EXTI1
	D1  = PA00  // UART TX            1[0]   PWM  EXTI0
	D2  = PB22  // Button "Up"                    EXTI6
	D3  = PB23  // Button "Down"                  EXTI7
	D4  = PA23  // NeoPixel                       EXTI7
	D5  = PB31  // I2C SDA            5[1]        EXTI15
	D6  = PB30  // I2C SCL            5[0]        EXTI14
	D7  = PB00  // HUB75 R1                       EXTI0
	D8  = PB01  // HUB75 G1                       EXTI1
	D9  = PB02  // HUB75 B1                       EXTI2
	D10 = PB03  // HUB75 R2                       EXTI3
	D11 = PB04  // HUB75 G2                       EXTI4
	D12 = PB05  // HUB75 B2                       EXTI5
	D13 = PA14  // LED                       PWM  EXTI14
	D14 = PB06  // HUB75 CLK                      EXTI6
	D15 = PB14  // HUB75 LAT                      EXTI14
	D16 = PB12  // HUB75 OE                       EXTI12
	D17 = PB07  // HUB75 ADDR A                   EXTI7
	D18 = PB08  // HUB75 ADDR B                   EXTI8
	D19 = PB09  // HUB75 ADDR C                   EXTI9
	D20 = PB15  // HUB75 ADDR D                   EXTI15
	D21 = PB13  // HUB75 ADDR E                   EXTI13
	D22 = PA02  // ADC (A0)                       EXTI2
	D23 = PA05  // ADC (A1)                       EXTI5
	D24 = PA04  // ADC (A2)                  PWM  EXTI4
	D25 = PA06  // ADC (A3)                  PWM  EXTI6
	D26 = PA07  // ADC (A4)                       EXTI7
	D27 = PA12  // ESP32 UART RX      4[1]   PWM  EXTI12
	D28 = PA13  // ESP32 UART TX      4[0]   PWM  EXTI13
	D29 = PA20  // ESP32 GPIO0               PWM  EXTI4
	D30 = PA21  // ESP32 Reset               PWM  EXTI5
	D31 = PA22  // ESP32 Busy                PWM  EXTI6
	D32 = PA18  // ESP32 RTS                 PWM  EXTI2
	D33 = PB17  // ESP32 SPI CS              PWM  EXTI1
	D34 = PA16  // ESP32 SPI SCK      3[1]   PWM  EXTI0
	D35 = PA17  // ESP32 SPI SDI      3[0]   PWM  EXTI1
	D36 = PA19  // ESP32 SPI SDO      1[3]   PWM  EXTI3
	D37 = NoPin // USB Host enable
	D38 = PA24  // USB DM
	D39 = PA27  // USB DP
	D40 = PA03  // DAC/VREFP
	D41 = PB10  // Flash QSPI SCK
	D42 = PB11  // Flash QSPI CS
	D43 = PA08  // Flash QSPI I00
	D44 = PA09  // Flash QSPI IO1
	D45 = PA10  // Flash QSPI IO2
	D46 = PA11  // Flash QSPI IO3
	D47 = PA27  // LIS3DH IRQ                     EXTI11
	D48 = PA05  // SPI SCK            0[1]        EXTI5
	D49 = PA04  // SPI SDO            0[0]   PWM  EXTI4
	D50 = PA07  // SPI SDI            0[3]        EXTI7
)

// Analog pins
const (
	A0 = PA02 // ADC Channel 0
	A1 = PA05 // ADC Channel 5
	A2 = PA04 // ADC Channel 4
	A3 = PA06 // ADC Channel 6
	A4 = PA07 // ADC Channel 7
)

// LED pins
const (
	LED      = D13
	NEOPIXEL = D4
)

// Button pins
const (
	BUTTON_UP   = D2
	BUTTON_DOWN = D3
)

// UART pins
const (
	UART1_RX_PIN = D0 // SERCOM1[1]
	UART1_TX_PIN = D1 // SERCOM1[0]

	UART2_RX_PIN = D27 // SERCOM4[1] (ESP32 RX)
	UART2_TX_PIN = D28 // SERCOM4[0] (ESP32 TX)

	UART_RX_PIN = UART1_RX_PIN
	UART_TX_PIN = UART1_TX_PIN
)

// SPI pins
const (
	SPI0_SCK_PIN = D34 // SERCOM3[1] (ESP32 SCK)
	SPI0_SDO_PIN = D36 // SERCOM1[3] (ESP32 SDO)
	SPI0_SDI_PIN = D35 // SERCOM3[0] (ESP32 SDI)

	SPI1_SCK_PIN = D48 // SERCOM0[1]
	SPI1_SDO_PIN = D49 // SERCOM0[0]
	SPI1_SDI_PIN = D50 // SERCOM0[3]

	SPI_SCK_PIN = SPI0_SCK_PIN
	SPI_SDO_PIN = SPI0_SDO_PIN
	SPI_SDI_PIN = SPI0_SDI_PIN
)

// I2C pins
const (
	I2C0_SDA_PIN = D5 // SERCOM5[1]
	I2C0_SCL_PIN = D6 // SERCOM5[0]

	I2C_SDA_PIN = I2C0_SDA_PIN
	I2C_SCL_PIN = I2C0_SCL_PIN

	SDA_PIN = I2C_SDA_PIN // awkward naming required by machine_atsamd51.go
	SCL_PIN = I2C_SCL_PIN //
)

// ESP32 pins
const (
	NINA_ACK    = D31
	NINA_GPIO0  = D29
	NINA_RESETN = D30

	NINA_RX  = UART2_RX_PIN
	NINA_TX  = UART2_TX_PIN
	NINA_RTS = D32

	NINA_CS  = D33
	NINA_SDO = SPI0_SDO_PIN
	NINA_SDI = SPI0_SDI_PIN
	NINA_SCK = SPI0_SCK_PIN
)

// USB CDC pins (UART0)
const (
	USBCDC_DM_PIN = D38
	USBCDC_DP_PIN = D39

	UART0_RX_PIN = USBCDC_DM_PIN
	UART0_TX_PIN = USBCDC_DP_PIN
)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "Matrix Portal M4"
	usb_STRING_MANUFACTURER = "Adafruit Industries"
)

var (
	usb_VID uint16 = 0x239A
	usb_PID uint16 = 0x80C9
)

// HUB75 pins
const (
	HUB75_R1 = D7
	HUB75_G1 = D8
	HUB75_B1 = D9
	HUB75_R2 = D10
	HUB75_G2 = D11
	HUB75_B2 = D12

	HUB75_CLK    = D14
	HUB75_LAT    = D15
	HUB75_OE     = D16
	HUB75_ADDR_A = D17
	HUB75_ADDR_B = D18
	HUB75_ADDR_C = D19
	HUB75_ADDR_D = D20
	HUB75_ADDR_E = D21
)

// Peripheral configuration for HUB75
const (
	HUB75_TC      = 4           // timer peripheral (TC4)
	HUB75_TC_IRQ  = sam.IRQ_TC4 // interrupt ID (TC4)
	HUB75_TC_FREQ = 48000000    // timer peripheral clock frequency (48 MHz)
)

// hub75 is a singleton, implementing interface Hub75 from package "machine".
type hub75 struct{ isr func() }

// HUB75 represents the on-board HUB75 connector for RGB LED matrices.
var HUB75 = &hub75{}

// Pre-computed GPIO PORT bitmasks for the HUB75-related pins
const (
	rgbGroup  = uint32(HUB75_R1) >> 5
	rgbR1Mask = uint32(1 << (HUB75_R1 & 0x1F))
	rgbG1Mask = uint32(1 << (HUB75_G1 & 0x1F))
	rgbB1Mask = uint32(1 << (HUB75_B1 & 0x1F))
	rgbR2Mask = uint32(1 << (HUB75_R2 & 0x1F))
	rgbG2Mask = uint32(1 << (HUB75_G2 & 0x1F))
	rgbB2Mask = uint32(1 << (HUB75_B2 & 0x1F))
	rgbMask   = rgbR1Mask | rgbG1Mask | rgbB1Mask |
		rgbR2Mask | rgbG2Mask | rgbB2Mask

	addrGroup = uint32(HUB75_ADDR_A) >> 5
	addrAMask = uint32(1 << (HUB75_ADDR_A & 0x1F))
	addrBMask = uint32(1 << (HUB75_ADDR_B & 0x1F))
	addrCMask = uint32(1 << (HUB75_ADDR_C & 0x1F))
	addrDMask = uint32(1 << (HUB75_ADDR_D & 0x1F))
	addrEMask = uint32(1 << (HUB75_ADDR_E & 0x1F))
	addrMask  = addrAMask | addrBMask | addrCMask | addrDMask | addrEMask
)

// timer holds a reference, IRQ, bus clock (with mask), and GCLK ID for each
// SAMD51 general-purpose timer/counter (TC) peripheral in 16-bit counter mode.
//
// For HUB75, we use the 16-bit counter mode (TC_COUNT16_Type) instead of the
// 8-bit (TC_COUNT8_Type) or 32-bit (TC_COUNT32_Type) modes; the latter mode is
// actually just implemented as two 16-bit counters, where TC[n] 32-bit would
// use both TC[n] and TC[n+1] (e.g., TC2 32-bit requires both TC2 and TC3).
var timer = []struct {
	id int
	ok bool
	bc *volatile.Register32
	cm uint32
	tc *sam.TC_COUNT16_Type
}{
	{
		id: sam.PCHCTRL_GCLK_TC0,
		ok: false,
		bc: &sam.MCLK.APBAMASK,
		cm: sam.MCLK_APBAMASK_TC0_,
		tc: sam.TC0_COUNT16,
	},
	{
		id: sam.PCHCTRL_GCLK_TC1,
		ok: false,
		bc: &sam.MCLK.APBAMASK,
		cm: sam.MCLK_APBAMASK_TC1_,
		tc: sam.TC1_COUNT16,
	},
	{
		id: sam.PCHCTRL_GCLK_TC2,
		ok: false,
		bc: &sam.MCLK.APBBMASK,
		cm: sam.MCLK_APBBMASK_TC2_,
		tc: sam.TC2_COUNT16,
	},
	{
		id: sam.PCHCTRL_GCLK_TC3,
		ok: false,
		bc: &sam.MCLK.APBBMASK,
		cm: sam.MCLK_APBBMASK_TC3_,
		tc: sam.TC3_COUNT16,
	},
	{
		id: sam.PCHCTRL_GCLK_TC4,
		ok: false,
		bc: &sam.MCLK.APBCMASK,
		cm: sam.MCLK_APBCMASK_TC4_,
		tc: sam.TC4_COUNT16,
	},
	{
		id: sam.PCHCTRL_GCLK_TC5,
		ok: false,
		bc: &sam.MCLK.APBCMASK,
		cm: sam.MCLK_APBCMASK_TC5_,
		tc: sam.TC5_COUNT16,
	},
}

// SetRgb sets/clears each of the 6 RGB data pins.
func (hub *hub75) SetRgb(r1, g1, b1, r2, g2, b2 bool) {
	var data uint32
	if r1 {
		data |= rgbR1Mask
	}
	if g1 {
		data |= rgbG1Mask
	}
	if b1 {
		data |= rgbB1Mask
	}
	if r2 {
		data |= rgbR2Mask
	}
	if g2 {
		data |= rgbG2Mask
	}
	if b2 {
		data |= rgbB2Mask
	}
	hub.SetRgbMask(data)
}

// SetRgbMask sets/clears each of the 6 RGB data pins from the given bitmask.
func (hub *hub75) SetRgbMask(data uint32) {
	// replace the first 6 bits in PORT register OUT with the corresponding bits
	// from the given data, updating R1,G1,B1,R2,G2,B2 all simultaneously
	sam.PORT.GROUP[rgbGroup].OUT.ReplaceBits(data, rgbMask, 0)
}

// SetRow sets the active pair of data rows with the given index.
func (hub *hub75) SetRow(row int) {
	var addr uint32
	if 0 != row&(1<<0) { // 0x01
		addr |= addrAMask
	}
	if 0 != row&(1<<1) { // 0x02
		addr |= addrBMask
	}
	if 0 != row&(1<<2) { // 0x04
		addr |= addrCMask
	}
	if 0 != row&(1<<3) { // 0x08
		addr |= addrDMask
	}
	if 0 != row&(1<<4) { // 0x10
		addr |= addrEMask
	}
	sam.PORT.GROUP[addrGroup].OUT.ReplaceBits(addr, addrMask, 0)
}

// NOTE:
//   Clock hold timings for the SAMD51 (standard and overclocked [O/C] CPU) are
//   based on the Adafruit Protomatter (Arduino/C++) library here:
//   	 https://github.com/adafruit/Adafruit_Protomatter/blob/master/src/arch/samd51.h

// HoldClkLow sets and holds the clock line low for sufficient time when
// transferring one bit of data.
func (hub *hub75) HoldClkLow() {
	HUB75_CLK.Low()
	//arm.Asm("nop; nop") // CPU = 200 MHz [O/C]
	//arm.Asm("nop")      // CPU = 180 MHz [O/C]
	//arm.Asm("nop")      // CPU = 150 MHz [O/C]
	arm.Asm("nop") // CPU = 120 MHz
}

// HoldClkHigh sets and holds the clock line high for sufficient time when
// transferring one bit of data.
func (hub *hub75) HoldClkHigh() {
	HUB75_CLK.High()
	//arm.Asm("nop; nop; nop; nop; nop") // CPU = 200 MHz [O/C]
	//arm.Asm("nop; nop; nop; nop")      // CPU = 180 MHz [O/C]
	//arm.Asm("nop; nop; nop")           // CPU = 150 MHz [O/C]
	arm.Asm("nop; nop; nop") // CPU = 120 MHz
}

// HoldLatLow sets and holds the latch line low for sufficient time to stop
// signalling all shift registers from outputting their content.
func (hub *hub75) HoldLatLow() {
	HUB75_LAT.Low()
	arm.Asm("nop")
	HUB75_LAT.Low()
}

// HoldLatHigh sets and holds the latch line high for sufficient time to start
// signalling all shift registers to begin outputting their content.
func (hub *hub75) HoldLatHigh() {
	HUB75_LAT.High()
	arm.Asm("nop")
	HUB75_LAT.High()
}

// GetPinGroupAlignment returns true if and only if all given Pins are on the
// same GPIO port, and returns the minimum size of the group to which all pins
// belong (8, 16, or 32 if true, otherwise 0). Returns (true, 0) if no Pins
// are provided.
func (hub *hub75) GetPinGroupAlignment(pin ...Pin) (samePort bool, alignment uint8) {

	// return a unique condition if no pins provided
	if len(pin) == 0 {
		return true, 0
	}

	var (
		bitMask  uint32 // position of each pin
		byteMask uint8  // byte-wise grouping of pins
		group    uint8  // common GPIO PORT group
	)

	for i, p := range pin {
		grp, pos := p.getPinGrouping()
		if i == 0 {
			group = grp
		} else {
			if group != grp {
				return false, 0 // error: all pins not on same GPIO PORT
			}
		}
		bitMask |= 1 << pos
	}

	// if we've reached here, all pins are on the same GPIO PORT. now we need to
	// decide if they are all within an 8-, 16-, or 32-bit group.

	if bitMask&0x000000FF != 0 {
		byteMask |= 1 << 0 // 0x1
	}
	if bitMask&0x0000FF00 != 0 {
		byteMask |= 1 << 1 // 0x2
	}
	if bitMask&0x00FF0000 != 0 {
		byteMask |= 1 << 2 // 0x4
	}
	if bitMask&0xFF000000 != 0 {
		byteMask |= 1 << 3 // 0x8
	}

	// the above can be written as the following loop, but I don't think it's as
	// clear to the reader what we're doing.
	//for i, m := 0, uint32(0xFF); i < 4; i, m = i+1, m<<8 {
	//	if 0 != bitMask&m {
	//		byteMask |= 1 << i
	//	}
	//}

	// we have determined which group(s) of bytes to which all of the pins belong,
	// now we just have to count the number of adjacent groups there are.

	switch byteMask {
	case 0x1, 0x2, 0x4, 0x8:
		// all pins are in the same byte (0b0001, 0b0010, 0b0100, 0b1000)
		return true, 8 // 8-bit alignment

	case 0x3, 0x6, 0xC:
		// all pins are in the same word (0b0011, 0b0110, 0b1100)
		return true, 16 // 16-bit alignment

	default:
		// otherwise, the pins are spread out all across the register
		return true, 32 // 32-bit alignment
	}
}

// InitTimer is used to initialize a timer service that fires an interrupt at
// regular frequency, which, for HUB75, is used to signal row data transmission.
// The timer does not begin raising interrupts until ResumeTimer is called.
func (hub *hub75) InitTimer(handle func()) {

	id, bc, cm, tc :=
		timer[HUB75_TC].id, // GCLK peripheral channel ID
		timer[HUB75_TC].bc, // peripheral bus clock
		timer[HUB75_TC].cm, // peripheral bus clock mask
		timer[HUB75_TC].tc // SAMD51 TC peripheral (16-bit)

	// set the actual interrupt handler for row data transmission
	hub.isr = handle

	// disable clock source
	sam.GCLK.PCHCTRL[id].ClearBits(sam.GCLK_PCHCTRL_CHEN)
	for sam.GCLK.PCHCTRL[id].HasBits(sam.GCLK_PCHCTRL_CHEN) {
	} // wait for it to disable

	// run timer off of GCLK1
	sam.GCLK.PCHCTRL[id].ReplaceBits(
		sam.GCLK_PCHCTRL_GEN_GCLK1,
		sam.GCLK_PCHCTRL_GEN_Msk,
		sam.GCLK_PCHCTRL_GEN_Pos)

	// enable clock source
	sam.GCLK.PCHCTRL[id].SetBits(sam.GCLK_PCHCTRL_CHEN)
	for !sam.GCLK.PCHCTRL[id].HasBits(sam.GCLK_PCHCTRL_CHEN) {
	} // wait for it to enable

	// disable timer before configuring
	tc.CTRLA.ClearBits(sam.TC_COUNT16_CTRLA_ENABLE)
	for tc.SYNCBUSY.HasBits(sam.TC_COUNT16_SYNCBUSY_ENABLE) {
	} // wait for it to disable

	// // reset timer to default settings
	// tc.CTRLA.Set(sam.TC_COUNT16_CTRLA_SWRST)
	// for tc.SYNCBUSY.HasBits(sam.TC_COUNT16_SYNCBUSY_SWRST) {
	// } // wait for it to reset

	// enable the TC bus clock
	bc.SetBits(cm)

	// use 16-bit counter mode, DIV1 prescalar (1:1)
	mode, pdiv :=
		uint32(sam.TC_COUNT16_CTRLA_MODE_COUNT16<<
			sam.TC_COUNT16_CTRLA_MODE_Pos)&
			sam.TC_COUNT16_CTRLA_MODE_Msk,
		uint32(sam.TC_COUNT16_CTRLA_PRESCALER_DIV1<<
			sam.TC_COUNT16_CTRLA_PRESCALER_Pos)&
			sam.TC_COUNT16_CTRLA_PRESCALER_Msk
	tc.CTRLA.SetBits(mode | pdiv)

	// use match frequency (MFRQ) mode
	tc.WAVE.Set((sam.TC_COUNT16_WAVE_WAVEGEN_MFRQ <<
		sam.TC_COUNT16_WAVE_WAVEGEN_Pos) &
		sam.TC_COUNT16_WAVE_WAVEGEN_Msk)

	// use up-counter
	tc.CTRLBCLR.Set(sam.TC_COUNT16_CTRLBCLR_DIR)
	for tc.SYNCBUSY.HasBits(sam.TC_COUNT16_SYNCBUSY_CTRLB) {
	}

	// interrupt on overflow
	tc.INTENSET.Set(sam.TC_COUNT16_INTENSET_OVF)
	tc.INTFLAG.Set(sam.TC_COUNT16_INTFLAG_OVF) // clear any interrupt flag

	// install interrupt handler
	in := interrupt.New(HUB75_TC_IRQ, handleTimer)
	//in.SetPriority(0) // use highest priority
	in.Enable()
}

// ResumeTimer resumes the timer service, with given current value, that signals
// row data transmission for HUB75 by raising interrupts with given periodicity.
func (hub *hub75) ResumeTimer(value, period int) {

	// don't do anything if we already resumed
	if timer[HUB75_TC].ok {
		return
	}

	tc := timer[HUB75_TC].tc

	// clamp given current value to range of uint16
	if value < 0 {
		value = 0
	} else if value > 0xFFFF {
		value = 0xFFFF
	}
	// reset the current counter value
	tc.COUNT.Set(uint16(value))
	for tc.SYNCBUSY.HasBits(sam.TC_COUNT16_SYNCBUSY_COUNT) {
	} // wait for counter sync

	// clamp given period to range of uint16
	if period < 0 {
		period = 0 // TBD: a period of 0 probably doesn't make sense, is invalid?
	} else if period > 0xFFFF {
		period = 0xFFFF
	}
	// set the given period. note that if period is less than current counter
	// value, the counter will continue counting up until it has overflowed the
	// 16-bit storage, and then has to count back up to overflow period before it
	// will raise the next interrupt.
	tc.CC[0].Set(uint16(period))
	for tc.SYNCBUSY.HasBits(sam.TC_COUNT16_SYNCBUSY_CC0) {
	} // wait for period sync

	// enable the counter
	tc.CTRLA.Set(sam.TC_COUNT16_CTRLA_ENABLE)
	for tc.SYNCBUSY.HasBits(sam.TC_COUNT16_SYNCBUSY_ENABLE) {
	} // wait for it to enable

	timer[HUB75_TC].ok = true // timer has resumed
}

// PauseTimer pauses the timer service that signals row data transmission for
// HUB75 and returns the current value of the timer.
func (hub *hub75) PauseTimer() int {

	// don't do anything if we are already paused
	if !timer[HUB75_TC].ok {
		return -1
	}

	tc := timer[HUB75_TC].tc

	// request a synchronized read
	tc.CTRLBSET.Set((sam.TC_COUNT16_CTRLBSET_CMD_READSYNC <<
		sam.TC_COUNT16_CTRLBSET_CMD_Pos) &
		sam.TC_COUNT16_CTRLBSET_CMD_Msk)
	for tc.SYNCBUSY.HasBits(sam.TC_COUNT16_SYNCBUSY_CTRLB) {
	} // wait for command sync
	val := tc.COUNT.Get() // now safe to read

	// disable the counter
	tc.CTRLA.ClearBits(sam.TC_COUNT16_CTRLA_ENABLE)
	for tc.SYNCBUSY.HasBits(sam.TC_COUNT16_SYNCBUSY_STATUS) {
	} // wait for it to disable

	timer[HUB75_TC].ok = false // timer is now paused

	return int(val)
}

// handleTimer is a wrapper interrupt service routine (ISR) that simply calls
// the user-provided handler for row data transmission on each timer interrupt.
//
// The interrupt.New method requires constants for both IRQ and ISR, thus we
// can't use the argument to InitTimer as handler and must use this indirection
// as a constant handler.
func handleTimer(interrupt.Interrupt) {
	timer[HUB75_TC].tc.INTFLAG.Set(sam.TC_COUNT16_INTFLAG_OVF) // clear interrupt
	HUB75.isr()
}
