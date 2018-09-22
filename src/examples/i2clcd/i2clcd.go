// Connects to an jhd1313m1 I2C LCD display.
package main

import (
	"machine"
	"time"
)

const (
	REG_RED   = 0x04
	REG_GREEN = 0x03
	REG_BLUE  = 0x02

	LCD_CLEARDISPLAY        = 0x01
	LCD_RETURNHOME          = 0x02
	LCD_ENTRYMODESET        = 0x04
	LCD_DISPLAYCONTROL      = 0x08
	LCD_CURSORSHIFT         = 0x10
	LCD_FUNCTIONSET         = 0x20
	LCD_SETCGRAMADDR        = 0x40
	LCD_SETDDRAMADDR        = 0x80
	LCD_ENTRYRIGHT          = 0x00
	LCD_ENTRYLEFT           = 0x02
	LCD_ENTRYSHIFTINCREMENT = 0x01
	LCD_ENTRYSHIFTDECREMENT = 0x00
	LCD_DISPLAYON           = 0x04
	LCD_DISPLAYOFF          = 0x00
	LCD_CURSORON            = 0x02
	LCD_CURSOROFF           = 0x00
	LCD_BLINKON             = 0x01
	LCD_BLINKOFF            = 0x00
	LCD_DISPLAYMOVE         = 0x08
	LCD_CURSORMOVE          = 0x00
	LCD_MOVERIGHT           = 0x04
	LCD_MOVELEFT            = 0x00
	LCD_2LINE               = 0x08
	LCD_CMD                 = 0x80
	LCD_DATA                = 0x40

	LCD_2NDLINEOFFSET = 0x40
)

func main() {
	machine.I2CInit()
	time.Sleep(100 * time.Microsecond)

	// Init display
	initBytes := []byte{LCD_CMD, LCD_FUNCTIONSET | LCD_2LINE}
	machine.I2CWriteTo(0x3e, initBytes)

	// uncommmenting this next line will cause compile to fail
	// time.Sleep(450 * time.Microsecond)
	// machine.I2CWriteTo(0x3e, initBytes)
	// time.Sleep(150 * time.Microsecond)
	// machine.I2CWriteTo(0x3e, initBytes)

	// // Turn on display
	// machine.I2CWriteTo(0x3e, []byte{LCD_CMD, LCD_DISPLAYCONTROL | LCD_DISPLAYON})
	// time.Sleep(100 * time.Microsecond)

	// // Clear display
	// machine.I2CWriteTo(0x3e, []byte{LCD_CLEARDISPLAY})

	for {

	}
}
