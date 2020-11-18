/**
 * @file
 * @brief   System Clock Management (CLKMAN) Function Implementations. 
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
 * $Date: 2016-08-15 11:08:12 -0500 (Mon, 15 Aug 2016) $
 * $Revision: 24058 $
 *
 **************************************************************************** */

/* **** Includes **** */
#include "mxc_config.h"
#include "mxc_assert.h"
#include "clkman.h"
#include "pwrseq_regs.h"

 /**
 * @ingroup clkman 
 * @{
 */ 

/* ************************************************************************* */
void CLKMAN_SetSystemClock(clkman_system_source_select_t select, clkman_system_scale_t scale)
{
    MXC_CLKMAN->clk_ctrl = ((MXC_CLKMAN->clk_ctrl & ~MXC_F_CLKMAN_CLK_CTRL_SYSTEM_SOURCE_SELECT) |
                            (MXC_V_CLKMAN_CLK_CTRL_SYSTEM_SOURCE_SELECT_96MHZ_RO));

    switch(select) {
        case CLKMAN_SYSTEM_SOURCE_96MHZ:
        default:
            // Enable and select the 96MHz oscillator
            MXC_PWRSEQ->reg0 |= (MXC_F_PWRSEQ_REG0_PWR_ROEN_RUN);
            MXC_PWRSEQ->reg0 &= ~(MXC_F_PWRSEQ_REG0_PWR_OSC_SELECT);

            // Disable the 4MHz oscillator
            MXC_PWRSEQ->reg0 &= ~MXC_F_PWRSEQ_REG0_PWR_RCEN_RUN;

            // Divide the system clock by the scale
            MXC_PWRSEQ->reg3 = ((MXC_PWRSEQ->reg3 & ~MXC_F_PWRSEQ_REG3_PWR_RO_DIV) |
                                (scale << MXC_F_PWRSEQ_REG3_PWR_RO_DIV_POS));

            break;
        case CLKMAN_SYSTEM_SOURCE_4MHZ:
            // Enable and select the 4MHz oscillator
            MXC_PWRSEQ->reg0 |= (MXC_F_PWRSEQ_REG0_PWR_RCEN_RUN);
            MXC_PWRSEQ->reg0 |= (MXC_F_PWRSEQ_REG0_PWR_OSC_SELECT);

            // Disable the 96MHz oscillator
            MXC_PWRSEQ->reg0 &= ~MXC_F_PWRSEQ_REG0_PWR_ROEN_RUN;

            // 4MHz System source can only be divided down by a maximum factor of 8
            MXC_ASSERT(scale <= CLKMAN_SYSTEM_SCALE_DIV_8);

            // Divide the system clock by the scale
            MXC_PWRSEQ->reg3 = ((MXC_PWRSEQ->reg3 & ~MXC_F_PWRSEQ_REG3_PWR_RC_DIV) |
                                (scale << MXC_F_PWRSEQ_REG3_PWR_RC_DIV_POS));
            break;
    }

    SystemCoreClockUpdate();
}

/* ************************************************************************* */
void CLKMAN_CryptoClockEnable(int enable)
{
    if (enable) {
        /* Enable oscillator */
        MXC_CLKMAN->clk_config |= MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_ENABLE;
        /* Un-gate clock to TPU modules */
        MXC_CLKMAN->clk_ctrl |= MXC_F_CLKMAN_CLK_CTRL_CRYPTO_CLOCK_ENABLE;
    } else {
        /* Gate clock off */
        MXC_CLKMAN->clk_ctrl &= ~MXC_F_CLKMAN_CLK_CTRL_CRYPTO_CLOCK_ENABLE;
        /* Disable oscillator */
        MXC_CLKMAN->clk_config &= ~MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_ENABLE;
    }
}

/* ************************************************************************* */
void CLKMAN_SetClkScale(clkman_clk_t clk, clkman_scale_t scale)
{
    volatile uint32_t *clk_ctrl_reg;

    MXC_ASSERT(clk <= CLKMAN_CLK_MAX);
    MXC_ASSERT(scale != CLKMAN_SCALE_AUTO);

    if (clk < CLKMAN_CRYPTO_CLK_AES) {
        clk_ctrl_reg = &MXC_CLKMAN->sys_clk_ctrl_0_cm4 + clk;
    } else {
        clk_ctrl_reg = &MXC_CLKMAN->crypt_clk_ctrl_0_aes + (clk - CLKMAN_CRYPTO_CLK_AES);
    }

    *clk_ctrl_reg = scale;
}

/* ************************************************************************* */
clkman_scale_t CLKMAN_GetClkScale(clkman_clk_t clk)
{
    volatile uint32_t *clk_ctrl_reg;
    MXC_ASSERT(clk <= CLKMAN_CLK_MAX);

    if (clk < CLKMAN_CRYPTO_CLK_AES) {
        clk_ctrl_reg = &MXC_CLKMAN->sys_clk_ctrl_0_cm4 + clk;
    } else {
        clk_ctrl_reg = &MXC_CLKMAN->crypt_clk_ctrl_0_aes + (clk - CLKMAN_CRYPTO_CLK_AES);
    }

    return (clkman_scale_t)*clk_ctrl_reg;
}

/* ************************************************************************* */
void CLKMAN_ClockGate(clkman_enable_clk_t clk, int enable)
{
    if (enable) {
        MXC_CLKMAN->clk_ctrl |= clk;
    } else {
        MXC_CLKMAN->clk_ctrl &= ~clk;
    }
}

/* ************************************************************************ */
int CLKMAN_WdtClkSelect(unsigned int idx, clkman_wdt_clk_select_t select)
{
    MXC_ASSERT(idx < MXC_CFG_WDT_INSTANCES);

    if (select == CLKMAN_WDT_SELECT_DISABLED) {
        if (idx == 0) {
            MXC_CLKMAN->clk_ctrl &= ~MXC_F_CLKMAN_CLK_CTRL_WDT0_CLOCK_ENABLE;
        } else if (idx == 1) {
            MXC_CLKMAN->clk_ctrl &= ~MXC_F_CLKMAN_CLK_CTRL_WDT1_CLOCK_ENABLE;
        } else {
            return E_BAD_PARAM;
        }
    } else {
        if (idx == 0) {
            MXC_CLKMAN->clk_ctrl = (MXC_CLKMAN->clk_ctrl & ~MXC_F_CLKMAN_CLK_CTRL_WDT0_CLOCK_SELECT) |
                                   MXC_F_CLKMAN_CLK_CTRL_WDT0_CLOCK_ENABLE |
                                   ((select << MXC_F_CLKMAN_CLK_CTRL_WDT0_CLOCK_SELECT_POS) & MXC_F_CLKMAN_CLK_CTRL_WDT0_CLOCK_SELECT);
        } else if (idx == 1) {
            MXC_CLKMAN->clk_ctrl = (MXC_CLKMAN->clk_ctrl & ~MXC_F_CLKMAN_CLK_CTRL_WDT1_CLOCK_SELECT) |
                                   MXC_F_CLKMAN_CLK_CTRL_WDT1_CLOCK_ENABLE |
                                   ((select << MXC_F_CLKMAN_CLK_CTRL_WDT1_CLOCK_SELECT_POS) & MXC_F_CLKMAN_CLK_CTRL_WDT1_CLOCK_SELECT);
        } else {
            return E_BAD_PARAM;
        }
    }

    return E_NO_ERROR;
}

/* ************************************************************************* */
/* NOTE: CLKMAN_TrimRO() is implemented in system_max32XXX.c                 */
/* ************************************************************************* */

/**@} end of group clkman */
