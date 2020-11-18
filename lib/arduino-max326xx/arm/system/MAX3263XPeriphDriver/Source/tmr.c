/**
 * @file
 * @brief   Timer Peripheral Driver Source.
 */
/* *****************************************************************************
 * Copyright (C) 2016 Maxim Integrated Products, Inc., All Rights Reserved.
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
 * $Date: 2016-09-08 17:30:35 -0500 (Thu, 08 Sep 2016) $
 * $Revision: 24322 $
 *
 **************************************************************************** */

/* **** Includes **** */
#include <stddef.h>
#include "mxc_assert.h"
#include "tmr.h"
 
/**
 * @ingroup tmr
 * @{
 */
static tmr_prescale_t prescaler[MXC_CFG_TMR_INSTANCES];

/* ************************************************************************* */
int TMR_Init(mxc_tmr_regs_t *tmr, tmr_prescale_t prescale, const sys_cfg_tmr_t *sysCfg)
{
    int err;
    int tmrNum;

    //get the timer number
    tmrNum = MXC_TMR_GET_IDX(tmr);

    //check for valid pointer
    MXC_ASSERT(tmrNum >= 0);

    //steup system GPIO config
    if((err = SYS_TMR_Init(tmr, sysCfg)) != E_NO_ERROR)
        return err;

    //save the prescale value for this timer
    prescaler[tmrNum] = prescale;

    //Disable timer and clear settings
    tmr->ctrl = 0;

    //reset all counts to 0
    tmr->count32 = 0;
    tmr->count16_0 = 0;
    tmr->count16_1 = 0;

    // Clear interrupt flag
    tmr->intfl = MXC_F_TMR_INTFL_TIMER0 | MXC_F_TMR_INTFL_TIMER1;

    return E_NO_ERROR;
}

/* ************************************************************************* */
void TMR32_Config(mxc_tmr_regs_t *tmr, const tmr32_cfg_t *config)
{
    //stop timer
    TMR32_Stop(tmr);

    //setup timer configuration register
    //clear tmr2x16 (32bit mode), mode and polarity bits
    tmr->ctrl &= ~(MXC_F_TMR_CTRL_TMR2X16 | MXC_F_TMR_CTRL_MODE |
                   MXC_F_TMR_CTRL_POLARITY);

    //set mode and polarity
    tmr->ctrl |= ((config->mode << MXC_F_TMR_CTRL_MODE_POS) |
                  (config->polarity << MXC_F_TMR_CTRL_POLARITY_POS));

    //setup timer Tick registers
    tmr->term_cnt32 = config->compareCount;

    return;
}

/* ************************************************************************* */
void TMR32_PWMConfig(mxc_tmr_regs_t *tmr, const tmr32_cfg_pwm_t *config)
{
    //stop timer
    TMR32_Stop(tmr);

    //setup timer configuration register
    //clear tmr2x16 (32bit mode), mode and polarity bits
    tmr->ctrl &= ~(MXC_F_TMR_CTRL_TMR2X16 | MXC_F_TMR_CTRL_MODE |
                   MXC_F_TMR_CTRL_POLARITY);

    //set mode and polarity
    tmr->ctrl |= ((TMR32_MODE_PWM << MXC_F_TMR_CTRL_MODE_POS) |
                  (config->polarity << MXC_F_TMR_CTRL_POLARITY_POS));

    tmr->pwm_cap32 = config->dutyCount;

    //setup timer Tick registers
    tmr->count32 = 0;
    tmr->term_cnt32 = config->periodCount;

    return;
}

/* ************************************************************************* */
void TMR16_Config(mxc_tmr_regs_t *tmr, uint8_t index, const tmr16_cfg_t *config)
{
    //stop timer
    TMR16_Stop(tmr, index);

    if(index > 0) { //configure timer 16_1

        //setup timer configuration register
        tmr->ctrl |= MXC_F_TMR_CTRL_TMR2X16;   //1 = 16bit mode

        //set mode
        if(config->mode)
            tmr->ctrl |= MXC_F_TMR_CTRL_MODE_16_1;
        else
            tmr->ctrl &= ~MXC_F_TMR_CTRL_MODE_16_1;

        //setup timer Ticks registers
        tmr->term_cnt16_1 = config->compareCount;
    } else { //configure timer 16_0

        //setup timer configuration register
        tmr->ctrl |= MXC_F_TMR_CTRL_TMR2X16;    //1 = 16bit mode

        //set mode
        if(config->mode)
            tmr->ctrl |= MXC_F_TMR_CTRL_MODE_16_0;
        else
            tmr->ctrl &= ~MXC_F_TMR_CTRL_MODE_16_0;

        //setup timer Ticks registers
        tmr->term_cnt16_0 = config->compareCount;
    }

    return;
}

/* ************************************************************************* */
void TMR32_Start(mxc_tmr_regs_t *tmr)
{
    int tmrNum;
    uint32_t ctrl;

    //get the timer number
    tmrNum = MXC_TMR_GET_IDX(tmr);

    //prescaler gets reset to 0 when timer is disabled
    //set the prescale to the saved value for this timer
    ctrl = tmr->ctrl;
    ctrl &= ~(MXC_F_TMR_CTRL_PRESCALE); //clear prescaler bits
    ctrl |= prescaler[tmrNum] << MXC_F_TMR_CTRL_PRESCALE_POS;        //set prescaler
    ctrl |= MXC_F_TMR_CTRL_ENABLE0;     //set enable to start the timer

    tmr->ctrl = ctrl;

    return;
}

/* ************************************************************************* */
void TMR16_Start(mxc_tmr_regs_t *tmr, uint8_t index)
{
    int tmrNum;
    uint32_t ctrl;

    //get the timer number
    tmrNum = MXC_TMR_GET_IDX(tmr);

    ctrl = tmr->ctrl;

    //prescaler gets reset to 0 when both 16 bit timers are disabled
    //set the prescale to the saved value for this timer if is is not already set
    if((ctrl & MXC_F_TMR_CTRL_PRESCALE) != (prescaler[tmrNum] << MXC_F_TMR_CTRL_PRESCALE_POS)) {
        ctrl &= ~(MXC_F_TMR_CTRL_PRESCALE); //clear prescaler bits
        ctrl |= prescaler[tmrNum] << MXC_F_TMR_CTRL_PRESCALE_POS;   //set prescaler
    }

    if(index > 0)
        ctrl |= MXC_F_TMR_CTRL_ENABLE1; //start timer 16_1
    else
        ctrl |= MXC_F_TMR_CTRL_ENABLE0; //start timer 16_0

    tmr->ctrl = ctrl;

    return;
}

/* ************************************************************************* */
uint32_t TMR_GetPrescaler(mxc_tmr_regs_t *tmr)
{
    int tmrNum;

    //get the timer number
    tmrNum = MXC_TMR_GET_IDX(tmr);

    return ((uint32_t)prescaler[tmrNum]);
}


/* ************************************************************************* */
int TMR32_GetPWMTicks(mxc_tmr_regs_t *tmr, uint8_t dutyPercent, uint32_t freq, uint32_t *dutyTicks, uint32_t *periodTicks)
{
    uint32_t timerClock;
    uint32_t prescale;
    uint64_t ticks;

    if(dutyPercent > 100)
        return E_BAD_PARAM;

    if(freq == 0)
        return E_BAD_PARAM;

    timerClock = SYS_TMR_GetFreq(tmr);
    prescale = TMR_GetPrescaler(tmr);

    if(timerClock == 0 || prescale > TMR_PRESCALE_DIV_2_12)
        return E_UNINITIALIZED;

    ticks = timerClock / (1 << (prescale & 0xF)) / freq;

    //make sure ticks is within a 32 bit value
    if (!(ticks & 0xffffffff00000000)  && (ticks & 0xffffffff)) {
        *periodTicks = ticks;

        *dutyTicks = ((uint64_t)*periodTicks * dutyPercent) / 100;

        return E_NO_ERROR;
    }

    return E_INVALID;
}

/* ************************************************************************* */
int TMR32_TimeToTicks(mxc_tmr_regs_t *tmr, uint32_t time, tmr_unit_t units, uint32_t *ticks)
{
    uint32_t unit_div0, unit_div1;
    uint32_t timerClock;
    uint32_t prescale;
    uint64_t temp_ticks;

    timerClock = SYS_TMR_GetFreq(tmr);
    prescale = TMR_GetPrescaler(tmr);

    if(timerClock == 0 || prescale > TMR_PRESCALE_DIV_2_12)
        return E_UNINITIALIZED;

    switch (units) {
        case TMR_UNIT_NANOSEC:
            unit_div0 = 1000000;
            unit_div1 = 1000;
            break;
        case TMR_UNIT_MICROSEC:
            unit_div0 = 1000;
            unit_div1 = 1000;
            break;
        case TMR_UNIT_MILLISEC:
            unit_div0 = 1;
            unit_div1 = 1000;
            break;
        case TMR_UNIT_SEC:
            unit_div0 = 1;
            unit_div1 = 1;
            break;
        default:
            return E_BAD_PARAM;
    }

    temp_ticks = (uint64_t)time * (timerClock / unit_div0) / (unit_div1 * (1 << (prescale & 0xF)));

    //make sure ticks is within a 32 bit value
    if (!(temp_ticks & 0xffffffff00000000)  && (temp_ticks & 0xffffffff)) {
        *ticks = temp_ticks;
        return E_NO_ERROR;
    }

    return E_INVALID;
}

/* ************************************************************************* */
int TMR16_TimeToTicks(mxc_tmr_regs_t *tmr, uint32_t time, tmr_unit_t units, uint16_t *ticks)
{
    uint32_t unit_div0, unit_div1;
    uint32_t timerClock;
    uint32_t prescale;
    uint64_t temp_ticks;

    timerClock = SYS_TMR_GetFreq(tmr);
    prescale = TMR_GetPrescaler(tmr);

    if(timerClock == 0 || prescale > TMR_PRESCALE_DIV_2_12)
        return E_UNINITIALIZED;

    switch (units) {
        case TMR_UNIT_NANOSEC:
            unit_div0 = 1000000;
            unit_div1 = 1000;
            break;
        case TMR_UNIT_MICROSEC:
            unit_div0 = 1000;
            unit_div1 = 1000;
            break;
        case TMR_UNIT_MILLISEC:
            unit_div0 = 1;
            unit_div1 = 1000;
            break;
        case TMR_UNIT_SEC:
            unit_div0 = 1;
            unit_div1 = 1;
            break;
        default:
            return E_BAD_PARAM;
    }

    temp_ticks = (uint64_t)time * (timerClock / unit_div0) / (unit_div1 * (1 << (prescale & 0xF)));

    //make sure ticks is within a 32 bit value
    if (!(temp_ticks & 0xffffffffffff0000) && (temp_ticks & 0xffff)) {
        *ticks = temp_ticks;
        return E_NO_ERROR;
    }

    return E_INVALID;
}


/* ************************************************************************* */
int TMR_TicksToTime(mxc_tmr_regs_t *tmr, uint32_t ticks, uint32_t *time, tmr_unit_t *units)
{
    uint64_t temp_time = 0;

    uint32_t timerClock = SYS_TMR_GetFreq(tmr);
    uint32_t prescale = TMR_GetPrescaler(tmr);

    if(timerClock == 0 || prescale > TMR_PRESCALE_DIV_2_12)
        return E_UNINITIALIZED;

    tmr_unit_t temp_unit = TMR_UNIT_NANOSEC;
    temp_time = (uint64_t)ticks * 1000 * (1 << (prescale & 0xF)) / (timerClock / 1000000);
    if (!(temp_time & 0xffffffff00000000)) {
        *time = temp_time;
        *units = temp_unit;
        return E_NO_ERROR;
    }

    temp_unit = TMR_UNIT_MICROSEC;
    temp_time = (uint64_t)ticks * 1000 * (1 << (prescale & 0xF)) / (timerClock / 1000);
    if (!(temp_time & 0xffffffff00000000)) {
        *time = temp_time;
        *units = temp_unit;
        return E_NO_ERROR;
    }

    temp_unit = TMR_UNIT_MILLISEC;
    temp_time = (uint64_t)ticks * 1000 * (1 << (prescale & 0xF)) / timerClock;
    if (!(temp_time & 0xffffffff00000000)) {
        *time = temp_time;
        *units = temp_unit;
        return E_NO_ERROR;
    }

    temp_unit = TMR_UNIT_SEC;
    temp_time = (uint64_t)ticks * (1 << (prescale & 0xF)) / timerClock;
    if (!(temp_time & 0xffffffff00000000)) {
        *time = temp_time;
        *units = temp_unit;
        return E_NO_ERROR;
    }

    return E_INVALID;
}
/**@} end of ingroup tmr */
