/**
 * @file
 * @brief   Watchdog Timer 2 Function Implementations.
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
 * $Date: 2017-02-16 12:02:20 -0600 (Thu, 16 Feb 2017) $
 * $Revision: 26462 $
 *
 **************************************************************************** */

/* **** Includes **** */
#include <stddef.h>
#include "wdt2.h"
#include "pwrseq_regs.h"

/**
 * @addtogroup wdt2
 * @{
 */
static uint32_t interruptEnable = 0;   //keeps track to interrupts to enable in start function

/* ************************************************************************* */
int WDT2_Init(uint8_t runInSleep, uint8_t unlock_key)
{
    //enable nanoring in run and sleep mode
    MXC_PWRSEQ->reg0 |= (MXC_F_PWRSEQ_REG0_PWR_NREN_RUN);

    //unlock ctrl to be writable
    MXC_WDT2->lock_ctrl = unlock_key;

    //check to make sure it unlocked
    if (MXC_WDT2->lock_ctrl & 0x01)
        return E_BAD_STATE;

    //disable all interrupts
    interruptEnable = 0;
    MXC_WDT2->enable = interruptEnable;

    //enable the watchdog clock and clear all other settings
    MXC_WDT2->ctrl = (MXC_F_WDT2_CTRL_EN_CLOCK);

    //clear all interrupt flags
    MXC_WDT2->flags = WDT2_FLAGS_CLEAR_ALL;

    if(runInSleep) {
        // turn on nanoring during sleep
        MXC_PWRSEQ->reg0 |= (MXC_F_PWRSEQ_REG0_PWR_NREN_SLP);
        //turn on timer during sleep
        MXC_WDT2->ctrl |= MXC_F_WDT2_CTRL_EN_TIMER_SLP;
    } else {
        // turn off nanoring during sleep
        MXC_PWRSEQ->reg0 &= ~(MXC_F_PWRSEQ_REG0_PWR_NREN_SLP);
        //turn off timer during sleep
        MXC_WDT2->ctrl &= ~(MXC_F_WDT2_CTRL_EN_TIMER_SLP);
    }

    //lock ctrl to read-only
    MXC_WDT2->lock_ctrl = MXC_V_WDT2_LOCK_KEY;

    return E_NO_ERROR;
}

/* ************************************************************************* */
int WDT2_EnableWakeUp(wdt2_period_t int_period, uint8_t unlock_key)
{
    // Make sure interrupt period is valid
    if (int_period >= WDT2_PERIOD_MAX)
        return E_INVALID; 

    //unlock ctrl to be writable
    MXC_WDT2->lock_ctrl = unlock_key;

    //check to make sure it unlocked
    if (MXC_WDT2->lock_ctrl & 0x01)
        return E_BAD_STATE;

    //stop timer and clear interval period
    MXC_WDT2->ctrl &= ~(MXC_F_WDT2_CTRL_INT_PERIOD | MXC_F_WDT2_CTRL_EN_TIMER);

    //set interval period
    MXC_WDT2->ctrl |= (int_period << MXC_F_WDT2_CTRL_INT_PERIOD_POS);

    //enable timeout wake-up
    interruptEnable |= MXC_F_WDT2_ENABLE_TIMEOUT;

    //lock ctrl to read-only
    MXC_WDT2->lock_ctrl = MXC_V_WDT2_LOCK_KEY;

    // Enable wake-up
    MXC_PWRSEQ->msk_flags |= MXC_F_PWRSEQ_MSK_FLAGS_PWR_NANORING_WAKEUP_FLAG;

    return E_NO_ERROR;
}

/* ************************************************************************* */
int WDT2_DisableWakeUp(uint8_t unlock_key)
{
    //unlock register to be writable
    MXC_WDT2->lock_ctrl = unlock_key;

    //check to make sure it unlocked
    if (MXC_WDT2->lock_ctrl & 0x01)
        return E_BAD_STATE;

    //disable timeout wake-up
    interruptEnable &= ~MXC_F_WDT2_ENABLE_TIMEOUT;
    MXC_WDT2->enable = interruptEnable;

    //lock register to read-only
    MXC_WDT2->lock_ctrl = MXC_V_WDT2_LOCK_KEY;

    // disable wake-up
    MXC_PWRSEQ->msk_flags &= ~MXC_F_PWRSEQ_MSK_FLAGS_PWR_NANORING_WAKEUP_FLAG;

    return E_NO_ERROR;
}

/* ************************************************************************* */
int WDT2_EnableReset(wdt2_period_t rst_period, uint8_t unlock_key)
{
    // Make sure reset period is valid
    if (rst_period >= WDT2_PERIOD_MAX)
        return E_INVALID; 

    //unlock ctrl to be writable
    MXC_WDT2->lock_ctrl = unlock_key;

    //check to make sure it unlocked
    if (MXC_WDT2->lock_ctrl & 0x01)
        return E_BAD_STATE;

    //stop timer and clear reset period
    MXC_WDT2->ctrl &= ~(MXC_F_WDT2_CTRL_RST_PERIOD | MXC_F_WDT2_CTRL_EN_TIMER);

    //set reset period
    MXC_WDT2->ctrl |= (rst_period << MXC_F_WDT2_CTRL_RST_PERIOD_POS);

    //int flag has to be clear before interrupt enable can be written
    MXC_WDT2->flags  = MXC_F_WDT2_FLAGS_RESET_OUT;

    //enable reset0
    interruptEnable |= MXC_F_WDT2_ENABLE_RESET_OUT;

    //lock ctrl to read-only
    MXC_WDT2->lock_ctrl = MXC_V_WDT2_LOCK_KEY;

    //enable RSTN on WDT2 reset
    MXC_PWRSEQ->msk_flags |= MXC_F_PWRSEQ_MSK_FLAGS_PWR_WATCHDOG_RSTN_FLAG;

    return E_NO_ERROR;
}

/* ************************************************************************* */
int WDT2_DisableReset(uint8_t unlock_key)
{
    //unlock register to be writable
    MXC_WDT2->lock_ctrl = unlock_key;

    //check to make sure it unlocked
    if (MXC_WDT2->lock_ctrl & 0x01)
        return E_BAD_STATE;

    //disable reset
    interruptEnable &= ~MXC_F_WDT2_ENABLE_RESET_OUT;
    MXC_WDT2->enable = interruptEnable;

    //lock register to read-only
    MXC_WDT2->lock_ctrl = MXC_V_WDT2_LOCK_KEY;

    //disable RSTN on WDT2 reset
    MXC_PWRSEQ->msk_flags &= ~MXC_F_PWRSEQ_MSK_FLAGS_PWR_WATCHDOG_RSTN_FLAG;

    return E_NO_ERROR;
}

/* ************************************************************************* */
int WDT2_Start(uint8_t unlock_key)
{
    //check if watchdog is already running
    if(WDT2_IsActive())
        return E_BAD_STATE;

    //unlock ctrl to be writable
    MXC_WDT2->lock_ctrl = unlock_key;

    //check to make sure it unlocked
    if (MXC_WDT2->lock_ctrl & 0x01)
        return E_BAD_STATE;

    WDT2_Reset();

    //enable interrupts
    MXC_WDT2->enable = interruptEnable;

    //start timer
    MXC_WDT2->ctrl |= (MXC_F_WDT2_CTRL_EN_TIMER);

    //lock ctrl to read-only
    MXC_WDT2->lock_ctrl = MXC_V_WDT2_LOCK_KEY;

    return E_NO_ERROR;
}

/* ************************************************************************* */
void WDT2_Reset(void)
{
    //reset the watchdog counter
    MXC_WDT2->clear = MXC_V_WDT2_RESET_KEY_0;
    MXC_WDT2->clear = MXC_V_WDT2_RESET_KEY_1;

    //clear all interrupt flags
    MXC_WDT2->flags = WDT2_FLAGS_CLEAR_ALL;

    //wait for all interrupts to clear
    while(MXC_WDT2->flags != 0) {
        MXC_WDT2->flags = WDT2_FLAGS_CLEAR_ALL;
    }

    return;
}

/* ************************************************************************* */
int WDT2_Stop(uint8_t unlock_key)
{
    //check if watchdog is not running
    if(!WDT2_IsActive())
        return E_BAD_STATE;

    //unlock ctrl to be writable
    MXC_WDT2->lock_ctrl = unlock_key;

    //check to make sure it unlocked
    if (MXC_WDT2->lock_ctrl & 0x01)
        return E_BAD_STATE;

    //disabled the timer and interrupts
    MXC_WDT2->enable = 0;
    MXC_WDT2->ctrl &= ~(MXC_F_WDT2_CTRL_EN_TIMER);

    //lock ctrl to read-only
    MXC_WDT2->lock_ctrl = MXC_V_WDT2_LOCK_KEY;

    return E_NO_ERROR;
}
/**@} end of group wdt2*/
