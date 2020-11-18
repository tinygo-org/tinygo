/**
 * @file
 * @brief   Timer utility functions.
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
#include "tmr_utils.h"


/**
 * @ingroup tmr_utilities
 * @{
 */

/* **** Definitions **** */

/* **** Globals **** */

/* **** Functions **** */

/* ************************************************************************** */
void TMR_Delay(mxc_tmr_regs_t* tmr, unsigned long us)
{
    TMR_TO_Start(tmr, us);

    while(TMR_TO_Check(tmr) != E_TIME_OUT) {}
}

/* ************************************************************************** */
void TMR_TO_Start(mxc_tmr_regs_t* tmr, unsigned long us)
{
    unsigned clk_shift = 0;
    uint64_t max_us;
    uint32_t ticks;

    // Adjust the clk shift amout by how long the timeout is
    // Start with the fastest clock to give the greatest accuracy
    do {
        max_us = (uint64_t)((0xFFFFFFFFUL / ((uint64_t)SystemCoreClock >> clk_shift++)) * 1000000UL);
    } while(us > max_us);

    // Calculate the number of timer ticks we need to wait
    TMR_Init(tmr, (tmr_prescale_t)clk_shift, NULL);
    TMR32_TimeToTicks(tmr, us, TMR_UNIT_MICROSEC, &ticks);

    // Initialize the timer in one-shot mode
    tmr32_cfg_t cfg;
    cfg.mode = TMR32_MODE_ONE_SHOT;
    cfg.compareCount = ticks;
    TMR32_Stop(tmr);
    TMR32_Config(tmr, &cfg);

    TMR32_ClearFlag(tmr);
    TMR32_Start(tmr);
}

/* ************************************************************************** */
int TMR_TO_Check(mxc_tmr_regs_t* tmr)
{
    if(TMR32_GetFlag(tmr)) {
        return E_TIME_OUT;
    }
    return E_NO_ERROR;
}

/* ************************************************************************** */
void TMR_TO_Stop(mxc_tmr_regs_t* tmr)
{
    TMR32_Stop(tmr);
    TMR32_SetCount(tmr, 0x0);
}

/* ************************************************************************** */
void TMR_TO_Clear(mxc_tmr_regs_t* tmr)
{
    TMR32_ClearFlag(tmr);
    TMR32_SetCount(tmr, 0x0);
}

/* ************************************************************************** */
unsigned TMR_TO_Elapsed(mxc_tmr_regs_t* tmr)
{
    uint32_t elapsed;
    tmr_unit_t units;

    TMR_TicksToTime(tmr, TMR32_GetCount(tmr), &elapsed, &units);

    switch(units) {
        case TMR_UNIT_NANOSEC:
        default:
            return (elapsed / 1000);
        case TMR_UNIT_MICROSEC:
            return (elapsed);
        case TMR_UNIT_MILLISEC:
            return (elapsed * 1000);
        case TMR_UNIT_SEC:
            return (elapsed * 1000000);
    }
}

/* ************************************************************************** */
unsigned TMR_TO_Remaining(mxc_tmr_regs_t* tmr)
{
    uint32_t remaining_ticks, remaining_time;
    tmr_unit_t units;

    remaining_ticks = TMR32_GetCompare(tmr) - TMR32_GetCount(tmr);
    TMR_TicksToTime(tmr, remaining_ticks, &remaining_time, &units);

    switch(units) {
        case TMR_UNIT_NANOSEC:
        default:
            return (remaining_time / 1000);
        case TMR_UNIT_MICROSEC:
            return (remaining_time);
        case TMR_UNIT_MILLISEC:
            return (remaining_time * 1000);
        case TMR_UNIT_SEC:
            return (remaining_time * 1000000);
    }
}

/* ************************************************************************** */
void TMR_SW_Start(mxc_tmr_regs_t* tmr)
{
    TMR_TO_Start(tmr, 0xFFFFFFFF);
}

/* ************************************************************************** */
unsigned TMR_SW_Stop(mxc_tmr_regs_t* tmr)
{
    unsigned elapsed = TMR_TO_Elapsed(tmr);
    TMR_TO_Stop(tmr);
    return elapsed;
}
/** @} end of ingroup tmr_utilities */
