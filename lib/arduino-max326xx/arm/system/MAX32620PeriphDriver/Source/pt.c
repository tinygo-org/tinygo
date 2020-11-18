/**
 * @file
 * @brief   Pulse Train Engine Function Implementations. 
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
 * $Date: 2016-09-08 17:43:36 -0500 (Thu, 08 Sep 2016) $
 * $Revision: 24327 $
 *
 **************************************************************************** */

/* **** Includes **** */
#include <stddef.h>
#include "pt.h"

/**
 * @ingroup pulsetrain 
 * @{
 */

/* ************************************************************************* */
void PT_Init(sys_pt_clk_scale clk_scale)
{
    //disable all pulse trains
    MXC_PTG->enable = 0;

    //clear all interrupts
    MXC_PTG->intfl = MXC_PTG->intfl;

    SYS_PT_Init(clk_scale);
}

/* ************************************************************************* */
int PT_PTConfig(mxc_pt_regs_t *pt, pt_pt_cfg_t *cfg, const sys_cfg_pt_t *sysCfg)
{
    int err;
    uint32_t ptClock;
    uint32_t rate;

    //check for valid base pointer
    MXC_ASSERT(MXC_PT_GET_IDX(pt) >= 0);

    if(cfg == NULL)
        return E_NULL_PTR;

    if(cfg->bps == 0)
        return E_BAD_PARAM;

    //disable pulse train
    PT_Stop(pt);

    //setup system GPIO configuration
    if((err = SYS_PT_Config(pt, sysCfg)) != E_NO_ERROR)
        return err;

    //get PT clock frequency from SYS level
    ptClock = SYS_PT_GetFreq();

    if(ptClock == 0)
        return E_UNINITIALIZED;

    if(ptClock < (cfg->bps))
        return E_BAD_STATE;

    rate = (ptClock / (cfg->bps));

    pt->rate_length = ((rate << MXC_F_PT_RATE_LENGTH_RATE_CONTROL_POS)
                       & MXC_F_PT_RATE_LENGTH_RATE_CONTROL) |
                      ((cfg->ptLength << MXC_F_PT_RATE_LENGTH_MODE_POS)
                       & MXC_F_PT_RATE_LENGTH_MODE);

    pt->train = cfg->pattern;
    pt->loop = ((cfg->loop << MXC_F_PT_LOOP_COUNT_POS) & MXC_F_PT_LOOP_COUNT) |
               ((cfg->loopDelay << MXC_F_PT_LOOP_DELAY_POS) & MXC_F_PT_LOOP_DELAY);

    return E_NO_ERROR;
}

/* ************************************************************************* */
int PT_SqrWaveConfig(mxc_pt_regs_t *pt, uint32_t freq, const sys_cfg_pt_t *sysCfg)
{
    int err;
    uint32_t ptClock;
    uint32_t rate;

    //check for valid base pointer
    MXC_ASSERT(MXC_PT_GET_IDX(pt) >= 0);

    if(freq == 0)
        return E_BAD_PARAM;

    //disable pulse train
    PT_Stop(pt);

    //setup system GPIO configuration
    if((err = SYS_PT_Config(pt, sysCfg)) != E_NO_ERROR)
        return err;

    //get PT clock frequency from SYS level
    ptClock = SYS_PT_GetFreq();

    if(ptClock == 0)
        return E_UNINITIALIZED;

    if(ptClock < (2*freq))
        return E_BAD_STATE;

    rate = (ptClock / (2*freq)) + 1;

    pt->rate_length = ((rate << MXC_F_PT_RATE_LENGTH_RATE_CONTROL_POS)
                       & MXC_F_PT_RATE_LENGTH_RATE_CONTROL) |
                      (MXC_V_PT_RATE_LENGTH_MODE_SQUARE_WAVE << MXC_F_PT_RATE_LENGTH_MODE_POS);

    return E_NO_ERROR;
}
/**@} end of ingroup pulsetrain*/
