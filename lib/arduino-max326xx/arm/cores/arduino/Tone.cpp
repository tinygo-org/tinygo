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
#include "Tone.h"
#include "pt.h"

void tone(uint8_t pin, uint32_t freq, uint32_t duration)
{
    uint8_t clk_scale;
    uint32_t tmp, min_freq;
    sys_cfg_pt_t *pt_cfg;
    mxc_pt_regs_t *pt;

    pinMode(pin, OUTPUT);
    pt_cfg = GET_PIN_CFG(pin);

    // Change function to Pulse Train
    pt_cfg->func = GPIO_FUNC_PT;

    // Find pulseTrain for the pin
    tmp = ((pt_cfg->port * MXC_GPIO_MAX_PINS_PER_PORT) +
           (PIN_MASK_TO_PIN(pt_cfg->mask))) % MXC_CFG_PT_INSTANCES;
    pt = MXC_PT_GET_PT(tmp);

    // Find the clock scale for given frequency
    clk_scale = CLKMAN_SCALE_DISABLED;
    do {
        min_freq = SystemCoreClock >> clk_scale++;
    } while (freq < min_freq && clk_scale < CLKMAN_SCALE_DIV_256);

    // Initialize and configure for square wave i.e. 50% duty cycle
#ifdef MAX32630
    sys_cfg_ptg_t ptg_cfg;
    ptg_cfg.clk_scale = (clkman_scale_t)clk_scale;
    PT_Init(&ptg_cfg);
#else
    PT_Init((clkman_scale_t)clk_scale);
#endif
    PT_SqrWaveConfig(pt, freq, pt_cfg);

    PT_Start(pt);

    // Stop after certain duration, if specified
    if (duration) {
        delay(duration);
        PT_Stop(pt);
    }
}

void noTone(uint8_t pin)
{
    uint32_t tmp;
    sys_cfg_pt_t *pt_cfg;
    mxc_pt_regs_t *pt;

    pt_cfg = GET_PIN_CFG(pin);

    // Find pulseTrain for the pin
    tmp = ((pt_cfg->port * MXC_GPIO_MAX_PINS_PER_PORT) +
           (PIN_MASK_TO_PIN(pt_cfg->mask))) % MXC_CFG_PT_INSTANCES;
    pt = MXC_PT_GET_PT(tmp);

    if (PT_IsActive(pt)) {
        PT_Stop(pt);
    }
}
