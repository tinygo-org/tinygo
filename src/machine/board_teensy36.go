// +build nxp,mk66f18,teensy36

package machine

import (
	"device/nxp"
)

// //go:keep
// //go:section .flash_config
// var FlashControl = [16]byte{
// 	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
// 	0xFF, 0xFF, 0xFF, 0xFF, 0xDE, 0xF9, 0xFF, 0xFF,
// }

func CPUFrequency() uint32 {
	return 180000000
}

// LED on the Teensy
const LED Pin = 13

var pins = []pin{
	// {bit, control register, gpio register bank}
	0:  {16, &nxp.PORTB.PCR16, nxp.GPIOB},
	1:  {17, &nxp.PORTB.PCR17, nxp.GPIOB},
	2:  {0, &nxp.PORTD.PCR0, nxp.GPIOD},
	3:  {12, &nxp.PORTA.PCR12, nxp.GPIOA},
	4:  {13, &nxp.PORTA.PCR13, nxp.GPIOA},
	5:  {7, &nxp.PORTD.PCR7, nxp.GPIOD},
	6:  {4, &nxp.PORTD.PCR4, nxp.GPIOD},
	7:  {2, &nxp.PORTD.PCR2, nxp.GPIOD},
	8:  {3, &nxp.PORTD.PCR3, nxp.GPIOD},
	9:  {3, &nxp.PORTC.PCR3, nxp.GPIOC},
	10: {4, &nxp.PORTC.PCR4, nxp.GPIOC},
	11: {6, &nxp.PORTC.PCR6, nxp.GPIOC},
	12: {7, &nxp.PORTC.PCR7, nxp.GPIOC},
	13: {5, &nxp.PORTC.PCR5, nxp.GPIOC},
	14: {1, &nxp.PORTD.PCR1, nxp.GPIOD},
	15: {0, &nxp.PORTC.PCR0, nxp.GPIOC},
	16: {0, &nxp.PORTB.PCR0, nxp.GPIOB},
	17: {1, &nxp.PORTB.PCR1, nxp.GPIOB},
	18: {3, &nxp.PORTB.PCR3, nxp.GPIOB},
	19: {2, &nxp.PORTB.PCR2, nxp.GPIOB},
	20: {5, &nxp.PORTD.PCR5, nxp.GPIOD},
	21: {6, &nxp.PORTD.PCR6, nxp.GPIOD},
	22: {1, &nxp.PORTC.PCR1, nxp.GPIOC},
	23: {2, &nxp.PORTC.PCR2, nxp.GPIOC},
	24: {26, &nxp.PORTE.PCR26, nxp.GPIOE},
	25: {5, &nxp.PORTA.PCR5, nxp.GPIOA},
	26: {14, &nxp.PORTA.PCR14, nxp.GPIOA},
	27: {15, &nxp.PORTA.PCR15, nxp.GPIOA},
	28: {16, &nxp.PORTA.PCR16, nxp.GPIOA},
	29: {18, &nxp.PORTB.PCR18, nxp.GPIOB},
	30: {19, &nxp.PORTB.PCR19, nxp.GPIOB},
	31: {10, &nxp.PORTB.PCR10, nxp.GPIOB},
	32: {11, &nxp.PORTB.PCR11, nxp.GPIOB},
	33: {24, &nxp.PORTE.PCR24, nxp.GPIOE},
	34: {25, &nxp.PORTE.PCR25, nxp.GPIOE},
	35: {8, &nxp.PORTC.PCR8, nxp.GPIOC},
	36: {9, &nxp.PORTC.PCR9, nxp.GPIOC},
	37: {10, &nxp.PORTC.PCR10, nxp.GPIOC},
	38: {11, &nxp.PORTC.PCR11, nxp.GPIOC},
	39: {17, &nxp.PORTA.PCR17, nxp.GPIOA},
	40: {28, &nxp.PORTA.PCR28, nxp.GPIOA},
	41: {29, &nxp.PORTA.PCR29, nxp.GPIOA},
	42: {26, &nxp.PORTA.PCR26, nxp.GPIOA},
	43: {20, &nxp.PORTB.PCR20, nxp.GPIOB},
	44: {22, &nxp.PORTB.PCR22, nxp.GPIOB},
	45: {23, &nxp.PORTB.PCR23, nxp.GPIOB},
	46: {21, &nxp.PORTB.PCR21, nxp.GPIOB},
	47: {8, &nxp.PORTD.PCR8, nxp.GPIOD},
	48: {9, &nxp.PORTD.PCR9, nxp.GPIOD},
	49: {4, &nxp.PORTB.PCR4, nxp.GPIOB},
	50: {5, &nxp.PORTB.PCR5, nxp.GPIOB},
	51: {14, &nxp.PORTD.PCR14, nxp.GPIOD},
	52: {13, &nxp.PORTD.PCR13, nxp.GPIOD},
	53: {12, &nxp.PORTD.PCR12, nxp.GPIOD},
	54: {15, &nxp.PORTD.PCR15, nxp.GPIOD},
	55: {11, &nxp.PORTD.PCR11, nxp.GPIOD},
	56: {10, &nxp.PORTE.PCR10, nxp.GPIOE},
	57: {11, &nxp.PORTE.PCR11, nxp.GPIOE},
	58: {0, &nxp.PORTE.PCR0, nxp.GPIOE},
	59: {1, &nxp.PORTE.PCR1, nxp.GPIOE},
	60: {2, &nxp.PORTE.PCR2, nxp.GPIOE},
	61: {3, &nxp.PORTE.PCR3, nxp.GPIOE},
	62: {4, &nxp.PORTE.PCR4, nxp.GPIOE},
	63: {5, &nxp.PORTE.PCR5, nxp.GPIOE},
}

//go:inline
func (p Pin) reg() pin { return pins[p] }
