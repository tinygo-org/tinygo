/*
  Copyright (c) 2011 Arduino.  All right reserved.

  This library is free software; you can redistribute it and/or
  modify it under the terms of the GNU Lesser General Public
  License as published by the Free Software Foundation; either
  version 2.1 of the License, or (at your option) any later version.

  This library is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
  See the GNU Lesser General Public License for more details.

  You should have received a copy of the GNU Lesser General Public
  License along with this library; if not, write to the Free Software
  Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA

  Modified 2017 by Maxim Integrated for MAX326xx
*/

#include "Arduino.h"
#include "tmr_regs.h"
#include "gpio.h"

#ifdef __cplusplus
 extern "C" {
#endif

void pinMode(uint32_t pin, uint32_t mode)
{
    if (IS_DIGITAL(pin)) {
        uint32_t m =   (mode == INPUT)        ? GPIO_PAD_INPUT :         \
                       (mode == INPUT_PULLUP) ? GPIO_PAD_INPUT_PULLUP :  \
                                                GPIO_PAD_NORMAL;
        // Set the mode in pinLut
        SET_PIN_MODE(pin, m);

        // Set initial state
        if (GPIO_InGet(GET_PIN_CFG(pin)) || IS_LED(pin)) {
            GPIO_OutSet(GET_PIN_CFG(pin));
        } else {
            GPIO_OutClr(GET_PIN_CFG(pin));
        }

        GPIO_Config(GET_PIN_CFG(pin));
    }
}

void digitalWrite(uint32_t pin, uint32_t val)
{
    if (IS_DIGITAL(pin)) {
        if (val) {
            GPIO_OutSet(GET_PIN_CFG(pin));
        } else {
            GPIO_OutClr(GET_PIN_CFG(pin));
        }
    }
}

int digitalRead(uint32_t pin)
{
    uint32_t tmp;
    gpio_cfg_t *cfg;
    mxc_tmr_regs_t *tmr;

    if (!IS_DIGITAL(pin)) {
        return LOW;
    }

    cfg = GET_PIN_CFG(pin);

    // Find timer for pin
    tmp = ((cfg->port * MXC_GPIO_MAX_PINS_PER_PORT) + PIN_MASK_TO_PIN(cfg->mask)) %
            MXC_CFG_TMR_INSTANCES;
    tmr = MXC_TMR_GET_TMR(tmp);

    // Disable timer and turn off PWM mode
    if ((tmr->ctrl & MXC_F_TMR_CTRL_MODE) == MXC_S_TMR_CTRL_MODE_PWM) {
        tmr->ctrl &= ~(MXC_F_TMR_CTRL_ENABLE0 | MXC_S_TMR_CTRL_MODE_PWM);
        // Change the pin function to GPIO and input mode
        cfg->func = GPIO_FUNC_GPIO;
        cfg->pad = GPIO_PAD_INPUT;
    }
    return (GPIO_InGet(cfg) == 0) ? LOW : HIGH;
}

#ifdef __cplusplus
}
#endif
