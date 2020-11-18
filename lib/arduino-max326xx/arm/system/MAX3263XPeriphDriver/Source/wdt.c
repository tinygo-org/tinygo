/**
 * @file
 * @brief   Watchdog driver source.
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
 * $Date: 2016-09-08 17:27:05 -0500 (Thu, 08 Sep 2016) $
 * $Revision: 24321 $
 *
 **************************************************************************** */

/* **** Includes **** */
#include <stddef.h>
#include "wdt.h"
/**
 * @ingroup wdt
 * @{
 */
static uint32_t interruptEnable = 0;   //keeps track to interrupts to enable in start function

/* ************************************************************************* */
int WDT_Init(mxc_wdt_regs_t *wdt, const sys_cfg_wdt_t *cfg, uint8_t unlock_key)
{
    if ((wdt == NULL) || (cfg == NULL))
        return E_NULL_PTR;

    //setup watchdog clock
    SYS_WDT_Init(wdt, cfg);

    //unlock ctrl to be writable
    wdt->lock_ctrl = unlock_key;

    //check to make sure it unlocked
    if (wdt->lock_ctrl & 0x01)
        return E_BAD_STATE;

    //disable all interrupts
    interruptEnable = 0;
    wdt->enable = interruptEnable;

    //enable the watchdog clock and clear all other settings
    wdt->ctrl = MXC_F_WDT_CTRL_EN_CLOCK;

    //clear all interrupt flags
    wdt->flags = WDT_FLAGS_CLEAR_ALL;

    //lock ctrl to read-only
    wdt->lock_ctrl = MXC_V_WDT_LOCK_KEY;

    return E_NO_ERROR;
}

/* ************************************************************************* */
int WDT_EnableInt(mxc_wdt_regs_t *wdt, wdt_period_t int_period, uint8_t unlock_key)
{
    //unlock ctrl to be writable
    wdt->lock_ctrl = unlock_key;

    //check to make sure it unlocked
    if (wdt->lock_ctrl & 0x01)
        return E_BAD_STATE;

    //stop timer and clear interval period
    wdt->ctrl &= ~(MXC_F_WDT_CTRL_INT_PERIOD | MXC_F_WDT_CTRL_EN_TIMER);

    //set interval period
    wdt->ctrl |= (int_period << MXC_F_WDT_CTRL_INT_PERIOD_POS);

    //enable timeout interrupt
    interruptEnable |= MXC_F_WDT_ENABLE_TIMEOUT;

    //lock ctrl to read-only
    wdt->lock_ctrl = MXC_V_WDT_LOCK_KEY;

    return E_NO_ERROR;
}

/* ************************************************************************* */
int WDT_DisableInt(mxc_wdt_regs_t *wdt, uint8_t unlock_key)
{
    //unlock register to be writable
    wdt->lock_ctrl = unlock_key;

    //check to make sure it unlocked
    if (wdt->lock_ctrl & 0x01)
        return E_BAD_STATE;

    //disable timeout interrupt
    interruptEnable &= ~MXC_F_WDT_ENABLE_TIMEOUT;
    wdt->enable = interruptEnable;

    //lock register to read-only
    wdt->lock_ctrl = MXC_V_WDT_LOCK_KEY;

    return E_NO_ERROR;
}

/* ************************************************************************* */
int WDT_EnableWait(mxc_wdt_regs_t *wdt, wdt_period_t wait_period, uint8_t unlock_key)
{
    // Make sure wait_period is valid
    if (wait_period >= WDT_PERIOD_MAX)
        return E_INVALID;

    //unlock ctrl to be writable
    wdt->lock_ctrl = unlock_key;

    //check to make sure it unlocked
    if (wdt->lock_ctrl & 0x01)
        return E_BAD_STATE;

    //stop timer and clear wait period
    wdt->ctrl &= ~(MXC_F_WDT_CTRL_WAIT_PERIOD | MXC_F_WDT_CTRL_EN_TIMER);

    //set wait period
    wdt->ctrl |= (wait_period << MXC_F_WDT_CTRL_WAIT_PERIOD_POS);

    //enable wait interrupt
    interruptEnable |= MXC_F_WDT_ENABLE_PRE_WIN;

    //lock ctrl to read-only
    wdt->lock_ctrl = MXC_V_WDT_LOCK_KEY;

    return E_NO_ERROR;
}

/* ************************************************************************* */
int WDT_DisableWait(mxc_wdt_regs_t *wdt, uint8_t unlock_key)
{
    //unlock register to be writable
    wdt->lock_ctrl = unlock_key;

    //check to make sure it unlocked
    if (wdt->lock_ctrl & 0x01)
        return E_BAD_STATE;

    //disable wait interrupt
    interruptEnable &= ~MXC_F_WDT_ENABLE_PRE_WIN;
    wdt->enable = interruptEnable;

    //lock register to read-only
    wdt->lock_ctrl = MXC_V_WDT_LOCK_KEY;

    return E_NO_ERROR;
}

/* ************************************************************************* */
int WDT_EnableReset(mxc_wdt_regs_t *wdt, wdt_period_t rst_period, uint8_t unlock_key)
{
    // Make sure wait_period is valid
    if (rst_period >= WDT_PERIOD_MAX)
        return E_INVALID;

    //unlock ctrl to be writable
    wdt->lock_ctrl = unlock_key;

    //check to make sure it unlocked
    if (wdt->lock_ctrl & 0x01)
        return E_BAD_STATE;

    //stop timer and clear reset period
    wdt->ctrl &= ~(MXC_F_WDT_CTRL_RST_PERIOD | MXC_F_WDT_CTRL_EN_TIMER);

    //set reset period
    wdt->ctrl |= (rst_period << MXC_F_WDT_CTRL_RST_PERIOD_POS);

    //enable reset0
    interruptEnable |= MXC_F_WDT_ENABLE_RESET_OUT;

    //lock ctrl to read-only
    wdt->lock_ctrl = MXC_V_WDT_LOCK_KEY;

    return E_NO_ERROR;
}

/* ************************************************************************* */
int WDT_DisableReset(mxc_wdt_regs_t *wdt, uint8_t unlock_key)
{
    //unlock register to be writable
    wdt->lock_ctrl = unlock_key;

    //check to make sure it unlocked
    if (wdt->lock_ctrl & 0x01)
        return E_BAD_STATE;

    //disable reset0
    interruptEnable &= ~MXC_F_WDT_ENABLE_RESET_OUT;
    wdt->enable = interruptEnable;

    //lock register to read-only
    wdt->lock_ctrl = MXC_V_WDT_LOCK_KEY;

    return E_NO_ERROR;
}

/* ************************************************************************* */
int WDT_Start(mxc_wdt_regs_t *wdt, uint8_t unlock_key)
{
    //check if watchdog is already running
    if(WDT_IsActive(wdt))
        return E_BAD_STATE;

    //unlock ctrl to be writable
    wdt->lock_ctrl = unlock_key;

    //check to make sure it unlocked
    if (wdt->lock_ctrl & 0x01)
        return E_BAD_STATE;

    WDT_Reset(wdt);

    //enable interrupts
    wdt->enable = interruptEnable;

    //start timer
    wdt->ctrl |= MXC_F_WDT_CTRL_EN_TIMER;

    //lock ctrl to read-only
    wdt->lock_ctrl = MXC_V_WDT_LOCK_KEY;

    return E_NO_ERROR;
}

/* ************************************************************************* */
void WDT_Reset(mxc_wdt_regs_t *wdt)
{
    //reset the watchdog counter
    wdt->clear = MXC_V_WDT_RESET_KEY_0;
    wdt->clear = MXC_V_WDT_RESET_KEY_1;

    //clear all interrupt flags
    wdt->flags = WDT_FLAGS_CLEAR_ALL;

    //wait for all interrupts to clear
    while(wdt->flags != 0) {
        wdt->flags = WDT_FLAGS_CLEAR_ALL;
    }

    return;
}

/* ************************************************************************* */
int WDT_Stop(mxc_wdt_regs_t *wdt, uint8_t unlock_key)
{
    //unlock ctrl to be writable
    wdt->lock_ctrl = unlock_key;

    //check to make sure it unlocked
    if (wdt->lock_ctrl & 0x01)
        return E_BAD_STATE;

    //disabled the timer and interrupts
    wdt->enable = 0;
    wdt->ctrl &= ~(MXC_F_WDT_CTRL_EN_TIMER);

    //lock ctrl to read-only
    wdt->lock_ctrl = MXC_V_WDT_LOCK_KEY;

    return E_NO_ERROR;
}
/**@} end of ingroup wdt */
