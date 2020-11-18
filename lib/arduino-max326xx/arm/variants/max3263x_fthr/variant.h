/*******************************************************************************
 * Copyright (C) 2017 Maxim Integrated Products, Inc., All Rights Reserved.
 *
 * Permission is hereby granted, free of charge, to any person obtaining a
 * copy of this software and associated documentation files (the "Software"),
 * to deal in the Software without restriction, including without limitation
 * the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the
 * Software is furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included
 * in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS
 * OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
 * MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 * IN NO EVENT SHALL MAXIM INTEGRATED BE LIABLE FOR ANY CLAIM, DAMAGES
 * OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 * ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 * OTHER DEALINGS IN THE SOFTWARE.
 *
 * Except as contained in this notice, the name of Maxim Integrated
 * Products, Inc. shall not be used except as stated in the Maxim Integrated
 * Products, Inc. Branding Policy.
 *
 * The mere transfer of this software does not imply any licenses
 * of trade secrets, proprietary technology, copyrights, patents,
 * trademarks, maskwork rights, or any other form of intellectual
 * property whatsoever. Maxim Integrated Products, Inc. retains all
 * ownership rights.
 *
 ******************************************************************************/

#ifndef _VARIANT_MAX32630FTHR_H_
#define _VARIANT_MAX32630FTHR_H_

#include "gpio.h"

#define VARIANT_MCK RO_FREQ


// These serial port names are intended to allow libraries and architecture-neutral
// sketches to automatically default to the correct port name for a particular type
// of use.  For example, a GPS module would normally connect to SERIAL_PORT_HARDWARE_OPEN,
// the first hardware serial port whose RX/TX pins are not dedicated to another use.
//
// SERIAL_PORT_MONITOR        Port which normally prints to the Arduino Serial Monitor
//
// SERIAL_PORT_USBVIRTUAL     Port which is USB virtual serial
//
// SERIAL_PORT_LINUXBRIDGE    Port which connects to a Linux system via Bridge library
//
// SERIAL_PORT_HARDWARE       Hardware serial port, physical RX & TX pins.
//
// SERIAL_PORT_HARDWARE_OPEN  Hardware serial ports which are open for use.  Their RX & TX
//                            pins are NOT connected to anything by default.
#define SERIAL_PORT_MONITOR         Serial
#define SERIAL_PORT_HARDWARE        Serial
#define SERIAL_PORT_HARDWARE1       Serial1
#define SERIAL_PORT_HARDWARE2       Serial2
#define SERIAL_PORT_HARDWARE3       Serial3
#define SERIAL_PORT_HARDWARE_OPEN   Serial1
#define SERIAL_PORT_HARDWARE_OPEN1  Serial2
#define SERIAL_PORT_HARDWARE_OPEN2  Serial3

#define PIN_NC      -1
#define NOT_CONNECTED PIN_NC
#define PIN_ANALOG  -2
#define NUM_OF_PINS 53

#define DEFAULT_I2CM_PORT 1
#define DEFAULT_SPIM_PORT 2
#define DEFAULT_SERIAL_PORT 1

// LEDs
// ----
#define RED_LED     (20u)
#define GREEN_LED   (21u)
#define BLUE_LED    (22u)
#define PIN_LED     RED_LED
#define LED_BUILTIN RED_LED

// Push button
// ----
#define SW1 P2_3

// Standardized button names
#define BUTTON1 SW1


// Analog pins
// -----------
#define AIN_0  49
#define AIN_1  50
#define AIN_2  51
#define AIN_3  52
#define PIN_A0 AIN_0
#define PIN_A1 AIN_1
#define PIN_A2 AIN_2
#define PIN_A3 AIN_3
static const uint8_t A0 = AIN_0;
static const uint8_t A1 = AIN_1;
static const uint8_t A2 = AIN_2;
static const uint8_t A3 = AIN_3;
#define ADC_RESOLUTION 10

// SPI Interfaces
// --------------
#define SPI_INTERFACES_COUNT 3

// SPI
#define PIN_SPI_MISO    42
#define PIN_SPI_MOSI    41
#define PIN_SPI_SCK     40
#define PIN_SPI_SS      43
static const uint8_t SS   = PIN_SPI_SS;   // SPI Slave SS not used. Set here only for reference.
static const uint8_t MOSI = PIN_SPI_MOSI;
static const uint8_t MISO = PIN_SPI_MISO;
static const uint8_t SCK  = PIN_SPI_SCK;

// SPI0: On-board microSD card slot
#define PIN_SPI0_MISO   6
#define PIN_SPI0_MOSI   5
#define PIN_SPI0_SCK    4
#define PIN_SPI0_SS     7
static const uint8_t SS0   = PIN_SPI0_SS;
static const uint8_t MOSI0 = PIN_SPI0_MOSI;
static const uint8_t MISO0 = PIN_SPI0_MISO;
static const uint8_t SCK0  = PIN_SPI0_SCK;

// SPI1: Option to add XIP Flash on the board
#define PIN_SPI1_MISO   10
#define PIN_SPI1_MOSI   9
#define PIN_SPI1_SCK    8
#define PIN_SPI1_SS     11
static const uint8_t SS1   = PIN_SPI1_SS;
static const uint8_t MOSI1 = PIN_SPI1_MOSI;
static const uint8_t MISO1 = PIN_SPI1_MISO;
static const uint8_t SCK1  = PIN_SPI1_SCK;

// SPI2: Feather pins
#define PIN_SPI2_MISO   42
#define PIN_SPI2_MOSI   41
#define PIN_SPI2_SCK    40
#define PIN_SPI2_SS     43
static const uint8_t SS2   = PIN_SPI2_SS;
static const uint8_t MOSI2 = PIN_SPI2_MOSI;
static const uint8_t MISO2 = PIN_SPI2_MISO;
static const uint8_t SCK2  = PIN_SPI2_SCK;

// Wire Interfaces
// ---------------
#define WIRE_INTERFACES_COUNT 3

// Wire
#define PIN_WIRE_SDA    28
#define PIN_WIRE_SCL    29
static const uint8_t SDA = PIN_WIRE_SDA;
static const uint8_t SCL = PIN_WIRE_SCL;

// Wire0: Not connected (I2CM0)
#define PIN_WIRE0_SDA   14
#define PIN_WIRE0_SCL   15
static const uint8_t SDA0 = PIN_WIRE0_SDA;
static const uint8_t SCL0 = PIN_WIRE0_SCL;

// Wire1: Feather pins (I2CM1)
#define PIN_WIRE1_SDA   28
#define PIN_WIRE1_SCL   29
static const uint8_t SDA1 = PIN_WIRE1_SDA;
static const uint8_t SCL1 = PIN_WIRE1_SCL;

// Wire2: Feather pins, BMI160, MAX14690 (I2CM2)
#define PIN_WIRE2_SDA   47
#define PIN_WIRE2_SCL   48
static const uint8_t SDA2 = PIN_WIRE2_SDA;
static const uint8_t SCL2 = PIN_WIRE2_SCL;

// 1-Wire Master
// ------
#define OWM P4_0

// MAX14690N hardwired
// ------
#define PMIC_INT P3_7
#define MPC P2_7
#define MON AIN_0

// microSD hardwired
// ------
#define DETECT P2_2

// Macros
// ------
#define PIN_MASK_TO_PIN(msk) (31 - __CLZ(msk))

#define IS_VALID(p)       ( (p) < NUM_OF_PINS )
#define IS_NC(p)          ( pinLut[(p)].port == PIN_NC )
#define IS_ANALOG(p)      ( IS_VALID(p) && (pinLut[(p)].port == PIN_ANALOG) )
#define IS_DIGITAL(p)     ( IS_VALID(p) && !IS_NC(p) && !IS_ANALOG(p) )
#define GET_PIN_IRQ(p)    ( (uint32_t)MXC_GPIO_GET_IRQ(pinLut[p].port) )
#define GET_PIN_MASK(p)   ( pinLut[p].mask )
#define GET_PIN_PORT(p)   ( pinLut[p].port )
#define GET_PIN_CFG(p)    ( &pinLut[p] )
#define SET_PIN_MODE(p,m) ( pinLut[p].pad = m )
#define IS_LED(p)         ( (p == RED_LED) || (p == GREEN_LED) || (p == BLUE_LED) )

// Do not const pinLut, pin func and pad can be changed
extern gpio_cfg_t pinLut[];
extern const gpio_cfg_t VDDIOH_pins[];

// Switch I/O voltage
int useVDDIOH(int pin);
int useVDDIO(int pin);

// Enumeration to allow portability between mbed and arduino
enum mbedPins {
    P0_0 = 0, P0_1, P0_2, P0_3, P0_4, P0_5, P0_6, P0_7,
    P1_0 = 8, P1_1, P1_2, P1_3, P1_4, P1_5, P1_6, P1_7,
    P2_0 = 16, P2_1, P2_2, P2_3, P2_4, P2_5, P2_6, P2_7,
    P3_0 = 24, P3_1, P3_2, P3_3, P3_4, P3_5, P3_6, P3_7,
    P4_0 = 32, P4_1, P4_2, P4_3, P4_4, P4_5, P4_6, P4_7,
    P5_0 = 40, P5_1, P5_2, P5_3, P5_4, P5_5, P5_6, P5_7,
    P6_0 = 48,
};

#endif // _VARIANT_MAX32630FTHR_H_
