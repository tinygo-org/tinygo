/**
 * @file
 * @brief   Peripheral Management Unit (PMU) Function Implementations. 
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
 * $Date: 2016-09-08 17:44:03 -0500 (Thu, 08 Sep 2016) $
 * $Revision: 24328 $
 *
 **************************************************************************** */

/* **** Includes **** */
#include <stdio.h>
#include <stddef.h>
#include "mxc_config.h"
#include "mxc_assert.h"
#include "pmu.h"
/**
 * @ingroup pmuGroup 
 * @{
 */ 

#if (MXC_PMU_REV == 0)
/* MAX32630 A1 & A2 Erratum #6: PMU only supports channels 0-4 -- workaround */
#include "clkman_regs.h"
/* Channel 5 infinite loop program */
static const uint32_t pmu_0[] = {
  PMU_JUMP(0, 0, (uint32_t)pmu_0)
};
#endif

/* **** Local Function Prototypes **** */
static void (*callbacks[MXC_CFG_PMU_CHANNELS])(int);


/* ************************************************************************* */
void PMU_Handler(void)
{
    int channel;
    uint32_t cfg1, cfg2;
    mxc_pmu_regs_t *MXC_PMUn;

    for (channel = 0; channel < MXC_CFG_PMU_CHANNELS; channel++) {
        MXC_PMUn = &MXC_PMU0[channel];

        if (MXC_PMUn->cfg & MXC_F_PMU_CFG_INTERRUPT) {
            cfg1 = MXC_PMUn->cfg;
            /* Since any set flags will be cleared by the write-back below, mask them off */
            cfg2 = cfg1 & ~(MXC_F_PMU_CFG_LL_STOPPED | MXC_F_PMU_CFG_BUS_ERROR | MXC_F_PMU_CFG_TO_STAT);

            /* Clear the interrupt flag */
            MXC_PMUn->cfg = cfg2 | MXC_F_PMU_CFG_INTERRUPT;

            if (callbacks[channel]) {
                callbacks[channel](cfg1);
            }
        }
    }
}

/* ************************************************************************* */
int PMU_Start(unsigned int channel, const void *program_address, pmu_callback callback)
{
    if(channel >= MXC_CFG_PMU_CHANNELS)
        return E_BAD_PARAM;

    mxc_pmu_regs_t *MXC_PMUn = &MXC_PMU0[channel];
    uint32_t cfg = MXC_PMUn->cfg;

    /* is this channel already running? */
    if (cfg & MXC_F_PMU_CFG_ENABLE) {
        return E_BUSY;
    }

#if (MXC_PMU_REV == 0)
    /* MAX32630 A1 & A2 Erratum #6: PMU only supports channels 0-4 */
    if (channel == 5) {
      /* Channel 5 is used for the work-around */
      return E_BUSY;
    }
    /* Select always-ON clock for PMU */
    MXC_CLKMAN->clk_gate_ctrl0 |= MXC_F_CLKMAN_CLK_GATE_CTRL0_PMU_CLK_GATER;
    /* Start channel 5 with infinite-loop program */
    MXC_PMU5->cfg &= ~MXC_F_PMU_CFG_ENABLE; /* Clear enable and wipe W1C flags */
    MXC_PMU5->dscadr = (uint32_t)pmu_0;
    MXC_PMU5->cfg = MXC_F_PMU_CFG_ENABLE | (0x1c << MXC_F_PMU_CFG_BURST_SIZE_POS);
#endif
    /* Set callback */
    callbacks[channel] = callback;

    /* Set start op-code */
    MXC_PMUn->dscadr = (uint32_t)program_address;

    /* Configure the channel */
    cfg = (cfg & ~(MXC_F_PMU_CFG_MANUAL | MXC_F_PMU_CFG_BURST_SIZE)) | (0x1c << MXC_F_PMU_CFG_BURST_SIZE_POS);

    /* Enable if necessary */
    if (callback) {
        cfg |= MXC_F_PMU_CFG_INT_EN;
    } else {
        cfg &= ~MXC_F_PMU_CFG_INT_EN;
    }

    /* Start the channel */
    cfg |= MXC_F_PMU_CFG_ENABLE;

    /*If any W1C flags are set, this write will clear them */
    MXC_PMUn->cfg = cfg;

    return E_NO_ERROR;
}

/* ************************************************************************* */
void PMU_Stop(unsigned int channel)
{
    mxc_pmu_regs_t *MXC_PMUn = &MXC_PMU0[channel];
    uint32_t cfg = MXC_PMUn->cfg;

    /* Since any set flags will be cleared by the write-back below, mask them off */
    cfg &= ~(MXC_F_PMU_CFG_LL_STOPPED | MXC_F_PMU_CFG_BUS_ERROR | MXC_F_PMU_CFG_TO_STAT | MXC_F_PMU_CFG_INTERRUPT);

    /* Clear the enable bit to stop the channel */
    cfg &= ~MXC_F_PMU_CFG_ENABLE;

    MXC_PMUn->cfg = cfg;

    /* Remove callback */
    callbacks[channel] = NULL;

#if (MXC_PMU_REV == 0)
    /* MAX32630 A1 & A2 Erratum #6: PMU only supports channels 0-4 */
    /* Check channels 0-4 for any running channels. If none found, stop channel 5 */
    if ((MXC_PMU0->cfg & MXC_F_PMU_CFG_ENABLE) == 0 &&
    (MXC_PMU1->cfg & MXC_F_PMU_CFG_ENABLE) == 0 &&
    (MXC_PMU2->cfg & MXC_F_PMU_CFG_ENABLE) == 0 &&
    (MXC_PMU3->cfg & MXC_F_PMU_CFG_ENABLE) == 0 &&
    (MXC_PMU4->cfg & MXC_F_PMU_CFG_ENABLE) == 0) {
      MXC_PMU5->cfg &= ~MXC_F_PMU_CFG_ENABLE;
    }
#endif

}

/* ************************************************************************* */
int PMU_SetCounter(unsigned int channel, unsigned int counter, uint16_t value)
{
    if((channel >= MXC_CFG_PMU_CHANNELS) || counter > 1)
        return E_BAD_PARAM;

    mxc_pmu_regs_t *MXC_PMUn = &MXC_PMU0[channel];

    if (counter == 0) {
        MXC_PMUn->loop = (MXC_PMUn->loop & ~MXC_F_PMU_LOOP_COUNTER_0) | (value << MXC_F_PMU_LOOP_COUNTER_0_POS);
    } else {
        MXC_PMUn->loop = (MXC_PMUn->loop & ~MXC_F_PMU_LOOP_COUNTER_1) | (value << MXC_F_PMU_LOOP_COUNTER_1_POS);
    }

    return E_NO_ERROR;
}

/* ************************************************************************* */
int PMU_SetTimeout(unsigned int channel, pmu_ps_sel_t timeoutClkScale, pmu_to_sel_t timeoutTicks)
{
    if(channel >= MXC_CFG_PMU_CHANNELS)
        return E_BAD_PARAM;

    mxc_pmu_regs_t *MXC_PMUn = &MXC_PMU0[channel];
    uint32_t cfg = MXC_PMUn->cfg;

    /* Since any set flags will be cleared by the write-back below, mask them off */
    cfg &= ~(MXC_F_PMU_CFG_LL_STOPPED | MXC_F_PMU_CFG_BUS_ERROR | MXC_F_PMU_CFG_TO_STAT | MXC_F_PMU_CFG_INTERRUPT);

    /* Adjust timeout settings */
    cfg &= ~(MXC_F_PMU_CFG_TO_SEL | MXC_F_PMU_CFG_PS_SEL);
    cfg |= ((timeoutClkScale << MXC_F_PMU_CFG_PS_SEL_POS) & MXC_F_PMU_CFG_PS_SEL) |
           ((timeoutTicks << MXC_F_PMU_CFG_TO_SEL_POS) & MXC_F_PMU_CFG_TO_SEL);

    MXC_PMUn->cfg = cfg;

    return E_NO_ERROR;
}

/* ************************************************************************* */
uint32_t PMU_GetFlags(unsigned int channel)
{
    mxc_pmu_regs_t *MXC_PMUn = &MXC_PMU0[channel];
    uint32_t cfg = MXC_PMUn->cfg;

    /* Mask off configuration bits leaving only flag bits */
    cfg &= ~(MXC_F_PMU_CFG_ENABLE | MXC_F_PMU_CFG_MANUAL | MXC_F_PMU_CFG_TO_SEL | MXC_F_PMU_CFG_PS_SEL |
            MXC_F_PMU_CFG_INT_EN | MXC_F_PMU_CFG_BURST_SIZE);

    return cfg;
}

/* ************************************************************************* */
void PMU_ClearFlags(unsigned int channel, unsigned int mask)
{
    mxc_pmu_regs_t *MXC_PMUn = &MXC_PMU0[channel];
    uint32_t cfg = MXC_PMUn->cfg;

    /* Since any set flags will be cleared by the write-back below, mask them off */
    cfg &= ~(MXC_F_PMU_CFG_LL_STOPPED | MXC_F_PMU_CFG_BUS_ERROR | MXC_F_PMU_CFG_TO_STAT | MXC_F_PMU_CFG_INTERRUPT);

    /* Now, apply the caller-supplied bits to clear */
    cfg |= mask;

    MXC_PMUn->cfg = cfg;
}

/* ************************************************************************* */
uint32_t PMU_IsActive(unsigned int channel)
{
    mxc_pmu_regs_t *MXC_PMUn = &MXC_PMU0[channel];
    return (MXC_PMUn->cfg & MXC_F_PMU_CFG_ENABLE);
}
/**@} end of ingroup pmuGroup */
