/**
 * @file  
 * @brief Real-Time Clock Function Implementations.
 */
/* ****************************************************************************
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
 * $Date: 2016-09-08 17:43:13 -0500 (Thu, 08 Sep 2016) $
 * $Revision: 24326 $
 *
 **************************************************************************** */

/* **** Includes **** */
#include <string.h>

#include "rtc.h"
#include "mxc_assert.h"
#include "mxc_sys.h"

 /**
 * @ingroup rtc 
 * @{
 */ 
/* ************************************************************************* */
int RTC_Init(const rtc_cfg_t *cfg)
{
    int err;
    int i = 0;

    //init function -> validate configuration pointer is not NULL
    if (cfg == NULL)
        return E_NULL_PTR;
    //check to make sure that the passed in parameters, prescaler mask and snooze, are valid
    if ((cfg->prescalerMask > ((rtc_prescale_t)cfg->prescaler)) || (cfg->snoozeCount > MXC_F_RTC_SNZ_VAL_VALUE))
        return E_INVALID;

    // Set system level configurations
    if ((err = SYS_RTC_Init()) != E_NO_ERROR) {
        return err;
    }

    //disable rtc
    MXC_RTCTMR->ctrl &= ~(MXC_F_RTC_CTRL_ENABLE);

    //disable all interrupts
    MXC_RTCTMR->inten = 0;

    //clear all interrupts
    MXC_RTCTMR->flags = RTC_FLAGS_CLEAR_ALL;

    //reset starting count
    MXC_RTCTMR->timer = 0;

    //set the compare registers to the values passed in
    for(i = 0; i < RTC_NUM_COMPARE; i++)
        MXC_RTCTMR->comp[i] = cfg->compareCount[i];

    // set the prescaler
    MXC_RTCTMR->prescale = cfg->prescaler;
    // set the prescale mask, checked it for validity on entry
    MXC_RTCTMR->prescale_mask = cfg->prescalerMask;

    //set snooze mode (rtc_snooze_t)
    MXC_RTCTMR->ctrl &= (~MXC_F_RTC_CTRL_SNOOZE_ENABLE);
    MXC_RTCTMR->ctrl |= (cfg->snoozeMode << MXC_F_RTC_CTRL_SNOOZE_ENABLE_POS);

    //set the snooze count. Checked for validity on entry.
    MXC_RTCTMR->snz_val = (cfg->snoozeCount << MXC_F_RTC_SNZ_VAL_VALUE_POS) & MXC_F_RTC_SNZ_VAL_VALUE;

    //wait for pending actions to complete
    while(MXC_RTCTMR->ctrl & MXC_F_RTC_CTRL_PENDING);

    //reset trim to defaults, trim disabled, trim faster override disabled
    MXC_RTCTMR->trim_ctrl &= ~(MXC_F_RTC_TRIM_CTRL_TRIM_ENABLE_R | MXC_F_RTC_TRIM_CTRL_TRIM_FASTER_OVR_R);

    //set trim slower control bit to 0, which is trim faster by default
    MXC_RTCTMR->trim_value &= ~(MXC_F_RTC_TRIM_VALUE_TRIM_SLOWER_CONTROL);

    return E_NO_ERROR;
}

/* ************************************************************************* */
int RTC_SetCompare(uint8_t compareIndex, uint32_t counts)
{
    //check for invalid index
    if (compareIndex >= RTC_NUM_COMPARE)
        return E_INVALID;

    MXC_RTCTMR->comp[compareIndex] = counts;

    //wait for pending actions to complete
    while(MXC_RTCTMR->ctrl & MXC_F_RTC_CTRL_PENDING);

    return E_NO_ERROR;
}

/* ************************************************************************* */
uint32_t RTC_GetCompare(uint8_t compareIndex)
{
    //Debug Assert for Invalid Index
    MXC_ASSERT(compareIndex < RTC_NUM_COMPARE);
    //check for invalid index
    if (compareIndex >= RTC_NUM_COMPARE)
        return (uint32_t)(E_BAD_PARAM); /* Unsigned int, so if out of bounds we return 0xFFFFFFFD (-3) */

    return MXC_RTCTMR->comp[compareIndex];
}

/* ************************************************************************* */
int RTC_SetTrim(uint32_t trim, uint8_t trimSlow)
{
    // make sure rtc is disabled
    if(MXC_RTCTMR->ctrl & MXC_F_RTC_CTRL_ENABLE)
        return E_BAD_STATE; // RTC is active, bad state

    // Can check against this because it starts at bit 0 in the register
    // Need to check because too large of a value messes with the upper bits in
    // the trim register.
    if (trim > MXC_F_RTC_TRIM_VALUE_TRIM_VALUE)
        return E_INVALID;

    // write the trim to the hardware trim_value register
    MXC_RTCTMR->trim_value = (trim << MXC_F_RTC_TRIM_VALUE_TRIM_VALUE_POS) & MXC_F_RTC_TRIM_VALUE_TRIM_VALUE;

    if(trimSlow)
        MXC_RTCTMR->trim_value |= MXC_F_RTC_TRIM_VALUE_TRIM_SLOWER_CONTROL;
    else
        MXC_RTCTMR->trim_value &= ~MXC_F_RTC_TRIM_VALUE_TRIM_SLOWER_CONTROL;

    //wait for pending actions to complete
    while(MXC_RTCTMR->ctrl & MXC_F_RTC_CTRL_PENDING);

    return E_NO_ERROR;
}

/* ************************************************************************* */
uint32_t RTC_GetTrim()
{
    return MXC_RTCTMR->trim_value; // return the register value for trim
}

/* ************************************************************************* */
int RTC_TrimEnable(void)
{
    // make sure rtc is disabled
    if(MXC_RTCTMR->ctrl & MXC_F_RTC_CTRL_ENABLE)
        return E_BAD_STATE; // RTC is active, bad state

    MXC_RTCTMR->trim_ctrl = MXC_F_RTC_TRIM_CTRL_TRIM_ENABLE_R;

    //wait for pending actions to complete
    while(MXC_RTCTMR->ctrl & MXC_F_RTC_CTRL_PENDING);

    return E_NO_ERROR;
}

/* ************************************************************************* */
void RTC_TrimDisable(void)
{
    // clear the trim enable bit
    MXC_RTCTMR->trim_ctrl &= ~MXC_F_RTC_TRIM_CTRL_TRIM_ENABLE_R;

    //wait for pending actions to complete
    while(MXC_RTCTMR->ctrl & MXC_F_RTC_CTRL_PENDING);

    return;
}

/**@} end of ingroup rtc*/
