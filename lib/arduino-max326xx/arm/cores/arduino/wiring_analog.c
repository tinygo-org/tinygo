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
#include "trim_regs.h"
#include "adc_regs.h"

#ifdef __cplusplus
extern "C" {
#endif

static int adcInitialized = 0;
static uint32_t readResolution = READ_DEFAULT_RESOLUTION;
static uint32_t writeResolution = WRITE_DEFAULT_RESOLUTION;
static uint32_t writeSHL = 0;
static uint32_t writeSHR = 0;
static uint32_t readSHL = 0;
static uint32_t readSHR = 0;

// Modified version of standard Maxim API ADC_Init()
// To allow flexibility of selecting ADC reference voltage
static int adcInit(int ref)
{
    int err;

    if ((err = SYS_ADC_Init()) != E_NO_ERROR) {
        return err;
    }

    // Wipe previous configuration
    MXC_ADC->intr = 0;

    // Clear all ADC interrupt flags
    MXC_ADC->intr = MXC_ADC->intr;

    // Enable done interrupt
    MXC_ADC->intr = MXC_F_ADC_INTR_ADC_DONE_IE;

    // Enable ADC interface clock
    MXC_ADC->ctrl = MXC_F_ADC_CTRL_ADC_CLK_EN;

#ifdef MAX32620
    if (ref == EXTERNAL) {
        MXC_ADC->ctrl |= MXC_F_ADC_CTRL_ADC_XREF;
        // Trim values when using external reference
        MXC_TRIM->reg11_adc_trim0 = 0x02000200;
        MXC_TRIM->reg12_adc_trim1 = 0x02000200;
    } else {
        MXC_ADC->ctrl &= ~MXC_F_ADC_CTRL_ADC_XREF;
    }
#endif

    // Power up the ADC
    MXC_ADC->ctrl |= ( MXC_F_ADC_CTRL_ADC_PU |
                       MXC_F_ADC_CTRL_BUF_PU |
                       MXC_F_ADC_CTRL_ADC_REFBUF_PU |
                       MXC_F_ADC_CTRL_ADC_CHGPUMP_PU );

    return E_NO_ERROR;
}

void analogReference(int mode)
{
#ifdef MAX32620
    adcInit(mode);
    adcInitialized = 1;
#endif
}

void analogReadResolution(int newRes)
{
    if ((newRes != readResolution) && (newRes > 0) && (newRes <= READ_MAX_RESOLUTION)) {
        // Calculate the right/left shift values for mapping
        if (newRes > READ_DEFAULT_RESOLUTION) {
            readSHL = newRes - READ_DEFAULT_RESOLUTION;
            readSHR = READ_DEFAULT_RESOLUTION - readSHL;
        } else {
            readSHR = READ_DEFAULT_RESOLUTION - newRes;
        }
        // Update the current resolution
        readResolution = newRes;
    }
}

uint32_t analogRead(uint32_t pin)
{
    uint16_t result = 0;

    if (IS_ANALOG(pin)) {
        // Initialize ADC for first time only
        if (!adcInitialized) {
            adcInit(DEFAULT);
            adcInitialized = 1;
        }

        // 1- adc_scale, 0- bypass RTC oscillator
        ADC_StartConvert(GET_PIN_MASK(pin), 1, 0);

        if (ADC_GetData(&result) == E_OVERFLOW) {
            // Full scale result
            result = (2 << (writeResolution - 1)) - 1;
        }

        // Mapping result to the desired resolution
        if (readResolution != READ_DEFAULT_RESOLUTION) {
            if (readResolution > READ_DEFAULT_RESOLUTION) {
                result = (result << readSHL) | (result >> readSHR);
            } else {
                result >>= readSHR;
            }
        }
    }
    return result;
}

void analogWriteResolution(int newRes)
{
    if ((newRes != writeResolution) && (newRes > 0) && (newRes <= WRITE_MAX_RESOLUTION)) {
        // Calculate the right/left shift values for mapping
        if (newRes < WRITE_DEFAULT_RESOLUTION) {
            writeSHL = WRITE_DEFAULT_RESOLUTION - newRes;
            writeSHR = newRes - writeSHL;
        } else {
            writeSHR = newRes - WRITE_DEFAULT_RESOLUTION;
        }
        // Update the current resolution
        writeResolution = newRes;
    }
}

void analogWrite(uint32_t pin, uint32_t value)
{
    // Analog pins are input only, analogWrite is PWM on Digital pins
    if (IS_DIGITAL(pin)) {
        // Limiting the value, if exceeds
        if (value >= (2 << (writeResolution - 1)) - 1) {
            // Resolution is 1 based, left shift is 0 based
            value = (2 << (writeResolution - 1)) - 1;
        }

        // Mapping value to the desired resolution
        if (writeResolution != WRITE_DEFAULT_RESOLUTION) {
            if (writeResolution < WRITE_DEFAULT_RESOLUTION) {
                value = (value << writeSHL) | (value >> writeSHR);
            } else {
                value >>= writeSHR;
            }
        }

        uint32_t tmp;
        gpio_cfg_t *cfg;
        mxc_tmr_regs_t *tmr;

        cfg = GET_PIN_CFG(pin);

        // Find timer for pin
        tmp = ((cfg->port * MXC_GPIO_MAX_PINS_PER_PORT) + PIN_MASK_TO_PIN(cfg->mask)) % MXC_CFG_TMR_INSTANCES;
        tmr = MXC_TMR_GET_TMR(tmp);

        // Steady states; 0 = low OR MAX_VALUE = high
        if ((value == 0) || (value == (2 << (writeResolution - 1)) - 1)) {
            cfg->func = GPIO_FUNC_GPIO;
            cfg->pad = GPIO_PAD_NORMAL;
            if (value == 0) {
                GPIO_OutClr(cfg);
            } else {
                GPIO_OutSet(cfg);
            }

            // With no sync to timer this switch may/will glitch the output
            GPIO_Config(cfg);

            // Stop timer and clear mode, releasing it for subsequent use
            tmr->ctrl &= ~(MXC_F_TMR_CTRL_ENABLE0 | MXC_F_TMR_CTRL_MODE);

        } else if (cfg->func == GPIO_FUNC_TMR) {
            // Precalculate match value for immediate use after interrupt
            value *= ((SystemCoreClock / PWM_FREQUENCY_HZ) >> PWM_RESOLUTION);

            // Do not get stuck waiting on a disabled timer
            if (tmr->ctrl | MXC_F_TMR_CTRL_ENABLE0) {
                // Wait for terminal count interrupt before updating duty cycle
                while (!(tmr->intfl & MXC_F_TMR_INTFL_TIMER0));
                tmr->pwm_cap32 = value;
                tmr->intfl = MXC_F_TMR_INTFL_TIMER0;
            }
        } else {
            // Using timer configuration to determine availability
            if ((tmr->ctrl & MXC_F_TMR_CTRL_MODE) == MXC_S_TMR_CTRL_MODE_PWM) {
                // Give up if timer currently in use as PWM, this does not cover all use cases
                return;
            }

            // PWM configuration
            tmp = tmr->ctrl & ~MXC_F_TMR_CTRL_ENABLE0;
            tmr->ctrl = tmp;
            tmp &= ~MXC_F_TMR_CTRL_TMR2X16;
            tmp &= ~MXC_F_TMR_CTRL_MODE;
            tmp &= ~MXC_S_TMR_CTRL_PRESCALE_DIVIDE_BY_1;
            tmp |= MXC_S_TMR_CTRL_MODE_PWM;
            tmp |= MXC_F_TMR_CTRL_POLARITY;
            tmr->ctrl = tmp;

            tmr->count32 = 0;
            tmr->term_cnt32 = SystemCoreClock / PWM_FREQUENCY_HZ;
            tmr->pwm_cap32 = value * ((SystemCoreClock / PWM_FREQUENCY_HZ) >> PWM_RESOLUTION);

            // Timer output configuration
            cfg->func = GPIO_FUNC_TMR;
            cfg->pad = GPIO_PAD_NORMAL;
            GPIO_Config(cfg);

            tmr->ctrl |= MXC_F_TMR_CTRL_ENABLE0;
        }
    }
}

#ifdef __cplusplus
}
#endif
