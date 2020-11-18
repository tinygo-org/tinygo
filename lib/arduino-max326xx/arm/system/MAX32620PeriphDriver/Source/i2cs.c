/**
 * @file
 * @brief      This file contains the function implementations for the I2CS
 *             (Inter-Integrated Circuit Slave) peripheral module.
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
 * $Date: 2016-09-08 18:05:59 -0500 (Thu, 08 Sep 2016) $ 
 * $Revision: 24332 $
 *
 *************************************************************************** */

/* **** Includes **** */
#include <string.h>
#include "mxc_assert.h"
#include "mxc_errors.h"
#include "mxc_sys.h"
#include "i2cs.h"

/**
 * @ingroup i2cs
 * @{
 */
/* **** Definitions **** */

/* **** Globals ***** */


// No Doxygen documentation for the items between here and endcond. 
/* Clock divider lookup table */
static const uint32_t clk_div_table[2][8] = {
 /* I2CS_SPEED_100KHZ */
    {
        // 12000000
        (6 << MXC_F_I2CS_CLK_DIV_FS_FILTER_CLOCK_DIV_POS),
        // 24000000
        (12 << MXC_F_I2CS_CLK_DIV_FS_FILTER_CLOCK_DIV_POS),
        // 36000000 NOT SUPPORTED
        0,
        // 48000000
        (24 << MXC_F_I2CS_CLK_DIV_FS_FILTER_CLOCK_DIV_POS),
        // 60000000 NOT SUPPORTED
        0,
        // 72000000 NOT SUPPORTED
        0,
        // 84000000 NOT SUPPORTED
        0,
        // 96000000
        (48 << MXC_F_I2CS_CLK_DIV_FS_FILTER_CLOCK_DIV_POS)
    },
    /* I2CS_SPEED_400KHZ */
    {
        // 12000000
        (2 << MXC_F_I2CS_CLK_DIV_FS_FILTER_CLOCK_DIV_POS),
        // 24000000
        (3 << MXC_F_I2CS_CLK_DIV_FS_FILTER_CLOCK_DIV_POS),
        // 36000000 NOT SUPPORTED
        0,
        // 48000000
        (6 << MXC_F_I2CS_CLK_DIV_FS_FILTER_CLOCK_DIV_POS),
        // 60000000 NOT SUPPORTED
        0,
        // 72000000 NOT SUPPORTED
        0,
        // 84000000 NOT SUPPORTED
        0,
        // 96000000
        (12 << MXC_F_I2CS_CLK_DIV_FS_FILTER_CLOCK_DIV_POS)
    },
};


static void (*callbacks[MXC_CFG_I2CS_INSTANCES][MXC_CFG_I2CS_BUFFER_SIZE])(uint8_t);

/* **** Functions **** */

/* ************************************************************************* */
int I2CS_Init(mxc_i2cs_regs_t *i2cs, const sys_cfg_i2cs_t *sys_cfg, i2cs_speed_t speed, 
    uint16_t address, i2cs_addr_t addr_len)
{
    int err, i, i2cs_index;

    i2cs_index = MXC_I2CS_GET_IDX(i2cs);
    MXC_ASSERT(i2cs_index >= 0);

     // Set system level configurations
    if ((err = SYS_I2CS_Init(i2cs, sys_cfg)) != E_NO_ERROR) {
        return err;
    }

    // Compute clock array index
    int clki = ((SYS_I2CS_GetFreq(i2cs) / 12000000) - 1);

    // Get clock divider settings from lookup table
    if ((speed == I2CS_SPEED_100KHZ) && (clk_div_table[I2CS_SPEED_100KHZ][clki] > 0)) {
        i2cs->clk_div = clk_div_table[I2CS_SPEED_100KHZ][clki];
    } else if ((speed == I2CS_SPEED_400KHZ) && (clk_div_table[I2CS_SPEED_400KHZ][clki] > 0)) {
        i2cs->clk_div = clk_div_table[I2CS_SPEED_400KHZ][clki];
    } else {
        MXC_ASSERT_FAIL();
    }

    // Clear the interrupt callbacks
    for(i = 0; i < MXC_CFG_I2CS_BUFFER_SIZE; i++) {
        callbacks[i2cs_index][i] = NULL;
    }

    // Reset module
    i2cs->dev_id = MXC_F_I2CS_DEV_ID_SLAVE_RESET;
    i2cs->dev_id = ((((address >> 0) << MXC_F_I2CS_DEV_ID_SLAVE_DEV_ID_POS) 
        & MXC_F_I2CS_DEV_ID_SLAVE_DEV_ID) | addr_len);

    return E_NO_ERROR;
}

/* ************************************************************************* */
int I2CS_Shutdown(mxc_i2cs_regs_t *i2cs)
{
   int err;

   // Disable and clear interrupts
   i2cs->inten = 0;
   i2cs->intfl = i2cs->intfl;

   // clears system level configurations
   if ((err = SYS_I2CS_Shutdown(i2cs)) != E_NO_ERROR) {
       return err;
   }

   return E_NO_ERROR;
}

/* ************************************************************************* */
void I2CS_Handler(mxc_i2cs_regs_t *i2cs)
{
    uint32_t intfl;
    uint8_t i;
    int i2cs_index = MXC_I2CS_GET_IDX(i2cs);

    // Save and clear the interrupt flags
    intfl = i2cs->intfl;
    i2cs->intfl = intfl;

    // Process each interrupt
    for(i = 0; i < 32; i++) {
        if(intfl & (0x1 << i)) {
            if(callbacks[i2cs_index][i] != NULL) {
                callbacks[i2cs_index][i](i);
            }
        }
    }

}

/* ************************************************************************* */
void I2CS_RegisterCallback(mxc_i2cs_regs_t *i2cs, uint8_t addr, i2cs_callback_fn callback)
{
    int i2cs_index = MXC_I2CS_GET_IDX(i2cs);

    // Make sure we don't overflow
    MXC_ASSERT(addr < MXC_CFG_I2CS_BUFFER_SIZE);

    if(callback != NULL) {
        // Save the callback address
        callbacks[i2cs_index][addr] = callback;

        // Clear and Enable the interrupt for the given byte
        i2cs->intfl = (0x1 << addr);
        i2cs->inten |= (0x1 << addr);
    } else {
        // Disable and clear the interrupt
        i2cs->inten &= ~(0x1 << addr);
        i2cs->intfl = (0x1 << addr);

        // Clear the callback address
        callbacks[i2cs_index][addr] = NULL;
    }
}

/**@} end of group i2cs*/    
