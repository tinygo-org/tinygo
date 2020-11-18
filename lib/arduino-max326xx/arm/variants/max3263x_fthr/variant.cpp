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

#include "mxc_device.h"
#include "mxc_sys.h"
#include "variant.h"
#include "adc.h"
#include "MAX14690.h"

/*
pinLut is a lookup table which provides reference to the actual port,
pin, pin function, and pin direction for the given external pin number.

PORT_x               Port number x(0 to 4)
PIN_x                Pin number x(0 to 7) of a particular port
PIN_ANALOG           Operates in Analog mode. Not available for Digital GPIO
PIN_NC               Not Connected
GPIO_FUNC_GPIO       Operates in GPIO mode. Other modes: Pulse Train(PT), Timer(TMR)
GPIO_PAD_INPUT       Pin is Input unless specified to be INPUT_PULLUP or OUTPUT
 */

gpio_cfg_t pinLut[] = {                                 // Pin#     Octal
                        /* PORT 0 */
    { PORT_0, PIN_0, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 00       000
    { PORT_0, PIN_1, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 01       001
    { PORT_0, PIN_2, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 02       002
    { PORT_0, PIN_3, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 03       003
    { PORT_0, PIN_4, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 04       004
    { PORT_0, PIN_5, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 05       005
    { PORT_0, PIN_6, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 06       006
    { PORT_0, PIN_7, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 07       007

                        /* PORT 1 */
    { PORT_1, PIN_0, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 08       010
    { PORT_1, PIN_1, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 09       011
    { PORT_1, PIN_2, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 10       012
    { PORT_1, PIN_3, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 11       013
    { PORT_1, PIN_4, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 12       014
    { PORT_1, PIN_5, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 13       015
    { PORT_1, PIN_6, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 14       016
    { PORT_1, PIN_7, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 15       017

                        /* PORT 2 */
    { PORT_2, PIN_0, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 16       020
    { PORT_2, PIN_1, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 17       021
    { PORT_2, PIN_2, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 18       022
    { PORT_2, PIN_3, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 19       023
    { PORT_2, PIN_4, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 20       024
    { PORT_2, PIN_5, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 21       025
    { PORT_2, PIN_6, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 22       026
    { PORT_2, PIN_7, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 23       027

                        /* PORT 3 */
    { PORT_3, PIN_0, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 24       030
    { PORT_3, PIN_1, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 25       031
    { PORT_3, PIN_2, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 26       032
    { PORT_3, PIN_3, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 27       033
    { PORT_3, PIN_4, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 28       034
    { PORT_3, PIN_5, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 29       035
    { PORT_3, PIN_6, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 30       036
    { PORT_3, PIN_7, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 31       037

                        /* PORT 4 */
    { PORT_4, PIN_0, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 32       040
    { PORT_4, PIN_1, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 33       041
    { PORT_4, PIN_2, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 34       042
    { PORT_4, PIN_3, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 35       043
    { PORT_4, PIN_4, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 36       044
    { PORT_4, PIN_5, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 37       045
    { PORT_4, PIN_6, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 38       046
    { PORT_4, PIN_7, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 39       047

                        /* PORT 5 */
    { PORT_5, PIN_0, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 40       050
    { PORT_5, PIN_1, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 41       051
    { PORT_5, PIN_2, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 42       052
    { PORT_5, PIN_3, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 43       053
    { PORT_5, PIN_4, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 44       054
    { PORT_5, PIN_5, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 45       055
    { PORT_5, PIN_6, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 46       056
    { PORT_5, PIN_7, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 47       057

                        /* PORT 6 */
    { PORT_6, PIN_0, GPIO_FUNC_GPIO, GPIO_PAD_INPUT },  // 48       060

                      /* Analog Pins */
    { PIN_ANALOG, ADC_CH_0_DIV_5 },                     // 49
    { PIN_ANALOG, ADC_CH_1_DIV_5 },                     // 50
    { PIN_ANALOG, ADC_CH_2 },                           // 51
    { PIN_ANALOG, ADC_CH_3 },                           // 52
};

// Use VDDIOH 3.3v on all the pins below
const gpio_cfg_t VDDIOH_pins[] = {
    // micro SD card pins
    { PORT_0, PIN_4 | PIN_5 | PIN_6 | PIN_7, (gpio_func_t)0, (gpio_pad_t)0 },

    // LED pins
    { PORT_2, PIN_4 | PIN_5 | PIN_6, (gpio_func_t)0, (gpio_pad_t)0 },

    // Header pins
    { PORT_3, PIN_0 | PIN_1 | PIN_2
            | PIN_3 | PIN_4 | PIN_5, (gpio_func_t)0, (gpio_pad_t)0 },

    { PORT_4, PIN_0 | PIN_1 | PIN_2 | PIN_3
            | PIN_4 | PIN_5 | PIN_6 | PIN_7, (gpio_func_t)0, (gpio_pad_t)0 },

    { PORT_5, PIN_0 | PIN_1 | PIN_2 | PIN_3
            | PIN_4 | PIN_5 | PIN_6, (gpio_func_t)0, (gpio_pad_t)0 },
};

// Initilization code specific to MAX32630FTHR
void initVariant(void)
{
    MAX14690 pmic;

    // Override the default values
    pmic.ldo2Millivolts = 3300;
    pmic.ldo3Millivolts = 3300;
    pmic.ldo2Mode = MAX14690::LDO_ENABLED;
    pmic.ldo3Mode = MAX14690::LDO_ENABLED;
    pmic.monCfg = MAX14690::MON_HI_Z;
    // Note that writing the local value does directly affect the part
    // The buck-boost regulator will remain off until init is called

    // Call init to apply all settings to the PMIC
    pmic.init();

    // Set GPIO pins to 3.3v
    uint8_t port;
    for ( port = 0; port < (sizeof(VDDIOH_pins)/sizeof(VDDIOH_pins[0])); port++ ) {
        SYS_IOMAN_UseVDDIOH(&VDDIOH_pins[port]);
    }
}

// Change to VDDIOH voltage (3.3v)
// pin to change voltage
// returns -1 if VDDIOH not allowed, 0 otherwise
int useVDDIOH(int pin)
{
    // Pins which can be used at 3.3v(VDDIOH)
    if ((pin > 3 && pin < 8) ||     // port 0
        (pin > 19 && pin < 23) ||   // port 2
        (pin > 23 && pin < 30) ||   // port 3
        (pin > 31 && pin < 40) ||   // port 4
        (pin > 39 && pin < 47)) {   // port 5
            SYS_IOMAN_UseVDDIOH(GET_PIN_CFG(pin));
            return 0;
    } else {
        return -1;  // VDDIOH not allowed on this pin
    }
}

// Change to VDDIO voltage (1.8v)
// pin to change voltage
int useVDDIO(int pin)
{
    // All pins allowed to run at 1.8v(VDDIO)
    SYS_IOMAN_UseVDDIO(GET_PIN_CFG(pin));
    return 0;
}
