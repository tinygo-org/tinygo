/**
 * @file
 * @brief   SPI execute in place driver.
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
 * $Date: 2016-09-08 17:31:41 -0500 (Thu, 08 Sep 2016) $
 * $Revision: 24323 $
 *
 **************************************************************************** */

/* **** Includes **** */
#include <stddef.h>
#include "mxc_config.h"
#include "mxc_assert.h"
#include "spix.h"
#include "spix_regs.h"

/**
 * @ingroup spix 
 * @{
 */ 

/* **** Definitions **** */
#define CMD_CLOCKS          8
#define ADDR_3BYTE_CLOCKS   24
#define ADDR_4BYTE_CLOCKS   32

/***** Globals *****/

/***** Functions *****/

/******************************************************************************/
#if defined ( __GNUC__ )
#undef IAR_SPIX_PRAGMA //Make sure this is not defined for GCC
#endif 

#if IAR_SPIX_PRAGMA
// IAR memory section declaration for the SPIX functions to be loaded in RAM.
#pragma section=".spix_config"
#endif

#if(MXC_SPIX_REV == 0)

#if defined ( __GNUC__ )
__attribute__ ((section(".spix_config"), noinline))
#endif /* __GNUC */

#if IAR_SPIX_PRAGMA
#pragma location=".spix_config" // IAR locate function in RAM section .spix_config
#pragma optimize=no_inline      // IAR no inline optimization on this function
#endif /* IAR_PRAGMA */

static void SPIX_UpdateFBIgnore()
{
    // Update the feedback ignore clocks
    uint8_t clocks = 0;
    uint8_t no_cmd_clocks = 0;

    // Adjust the clocks for the command
    if((MXC_SPIX->fetch_ctrl & MXC_F_SPIX_FETCH_CTRL_CMD_WIDTH) ==
        MXC_S_SPIX_FETCH_CTRL_CMD_WIDTH_QUAD_IO) {

        clocks += CMD_CLOCKS/4;
    } else if((MXC_SPIX->fetch_ctrl & MXC_F_SPIX_FETCH_CTRL_CMD_WIDTH) ==
        MXC_S_SPIX_FETCH_CTRL_CMD_WIDTH_DUAL_IO) {

        clocks += CMD_CLOCKS/2;
    } else {

        clocks += CMD_CLOCKS;
    }

    // Adjust the clocks for the address
    if((MXC_SPIX->fetch_ctrl & MXC_F_SPIX_FETCH_CTRL_ADDR_WIDTH) ==
        MXC_S_SPIX_FETCH_CTRL_ADDR_WIDTH_QUAD_IO) {

        if(MXC_SPIX->fetch_ctrl & MXC_F_SPIX_FETCH_CTRL_FOUR_BYTE_ADDR) {
            clocks += ADDR_4BYTE_CLOCKS/4;
            no_cmd_clocks += ADDR_4BYTE_CLOCKS/4;
        } else {
            clocks += ADDR_3BYTE_CLOCKS/4;
            no_cmd_clocks += ADDR_3BYTE_CLOCKS/4;
        }

    } else if((MXC_SPIX->fetch_ctrl & MXC_F_SPIX_FETCH_CTRL_ADDR_WIDTH) ==
        MXC_S_SPIX_FETCH_CTRL_ADDR_WIDTH_DUAL_IO) {

        if(MXC_SPIX->fetch_ctrl & MXC_F_SPIX_FETCH_CTRL_FOUR_BYTE_ADDR) {
            clocks += ADDR_4BYTE_CLOCKS/2;
            no_cmd_clocks += ADDR_4BYTE_CLOCKS/2;
        } else {
            clocks += ADDR_3BYTE_CLOCKS/2;
            no_cmd_clocks += ADDR_3BYTE_CLOCKS/2;
        }
    } else {

        if(MXC_SPIX->fetch_ctrl & MXC_F_SPIX_FETCH_CTRL_FOUR_BYTE_ADDR) {
            clocks += ADDR_4BYTE_CLOCKS;
            no_cmd_clocks += ADDR_4BYTE_CLOCKS;
        } else {
            clocks += ADDR_3BYTE_CLOCKS;
            no_cmd_clocks += ADDR_3BYTE_CLOCKS;
        }
    }

    // Adjust for the mode clocks
    clocks += ((MXC_SPIX->mode_ctrl & MXC_F_SPIX_MODE_CTRL_MODE_CLOCKS) >>
        MXC_F_SPIX_MODE_CTRL_MODE_CLOCKS_POS);

    // Set the FB Ignore clocks
    MXC_SPIX->sck_fb_ctrl = ((MXC_SPIX->sck_fb_ctrl & ~MXC_F_SPIX_SCK_FB_CTRL_IGNORE_CLKS) |
        (clocks << MXC_F_SPIX_SCK_FB_CTRL_IGNORE_CLKS_POS));

    MXC_SPIX->sck_fb_ctrl = ((MXC_SPIX->sck_fb_ctrl & ~MXC_F_SPIX_SCK_FB_CTRL_IGNORE_CLKS_NO_CMD) |
        (no_cmd_clocks << MXC_F_SPIX_SCK_FB_CTRL_IGNORE_CLKS_NO_CMD_POS));
}
#endif /* MXC_SPIX_REV==0 */

/******************************************************************************/
#if defined ( __GNUC__ )
__attribute__ ((section(".spix_config"), noinline))
#endif /* __GNUC */

#if IAR_SPIX_PRAGMA
#pragma location=".spix_config" // IAR locate function in RAM section .spix_config
#pragma optimize=no_inline      // IAR no inline optimization on this function
#endif /* IAR_SPIX_PRAGMA */
int SPIX_ConfigClock(const sys_cfg_spix_t *sys_cfg, uint32_t baud, uint8_t sample)
{
    int err;
    uint32_t spix_clk, clocks;

    // Check the input parameters
    if(sys_cfg == NULL) {
        return E_NULL_PTR;
    }

    // Set system level configurations
    if ((err = SYS_SPIX_Init(sys_cfg, baud)) != E_NO_ERROR) {
        return err;
    }

    // Configure the mode and baud
    spix_clk = SYS_SPIX_GetFreq();
    if(spix_clk <= 0) {
        return E_UNINITIALIZED;
    }

    // Make sure that we can generate this frequency
    clocks = (spix_clk / (2*baud));
    if((clocks <= 0) || (clocks >= 0x10)) {
        return E_BAD_PARAM;
    }

    // Set the baud
    MXC_SPIX->master_cfg = ((MXC_SPIX->master_cfg &
        ~(MXC_F_SPIX_MASTER_CFG_SCK_HI_CLK | MXC_F_SPIX_MASTER_CFG_SCK_LO_CLK)) |
        (clocks << MXC_F_SPIX_MASTER_CFG_SCK_HI_CLK_POS) |
        (clocks << MXC_F_SPIX_MASTER_CFG_SCK_LO_CLK_POS));

    if(sample != 0) {
        // Use sample mode
        MXC_SPIX->master_cfg = ((MXC_SPIX->master_cfg & ~MXC_F_SPIX_MASTER_CFG_SDIO_SAMPLE_POINT) |
            (sample << MXC_F_SPIX_MASTER_CFG_SDIO_SAMPLE_POINT_POS));

        MXC_SPIX->sck_fb_ctrl &= ~(MXC_F_SPIX_SCK_FB_CTRL_ENABLE_SCK_FB_MODE |
            MXC_F_SPIX_SCK_FB_CTRL_INVERT_SCK_FB_CLK);
    } else {
        // Use Feedback mode
        MXC_SPIX->master_cfg &= ~(MXC_F_SPIX_MASTER_CFG_SDIO_SAMPLE_POINT);

        MXC_SPIX->sck_fb_ctrl |= (MXC_F_SPIX_SCK_FB_CTRL_ENABLE_SCK_FB_MODE |
            MXC_F_SPIX_SCK_FB_CTRL_INVERT_SCK_FB_CLK);


#if(MXC_SPIX_REV == 0)
        SPIX_UpdateFBIgnore();
#endif
    }

    return E_NO_ERROR;
}

/******************************************************************************/
#if defined ( __GNUC__ )
__attribute__ ((section(".spix_config"), noinline))
#endif /* __GNUC */

#if IAR_SPIX_PRAGMA
#pragma location=".spix_config" // IAR locate function in RAM section .spix_config
#pragma optimize=no_inline      // IAR no inline optimization on this function
#endif /* IAR_SPIX_PRAGMA */

void SPIX_ConfigSlave(uint8_t ssel, uint8_t pol, uint8_t act_delay, uint8_t inact_delay)
{

    // Set the slave select
    MXC_SPIX->master_cfg = ((MXC_SPIX->master_cfg & ~MXC_F_SPIX_MASTER_CFG_SLAVE_SEL) |
        (ssel << MXC_F_SPIX_MASTER_CFG_SLAVE_SEL_POS));

    if(pol != 0) {
        // Active high
        MXC_SPIX->master_cfg &= ~(MXC_F_SPIX_MASTER_CFG_SS_ACT_LO);
    } else {
        // Active low
        MXC_SPIX->master_cfg |= MXC_F_SPIX_MASTER_CFG_SS_ACT_LO;
    }

    // Set the delays
    MXC_SPIX->master_cfg = ((MXC_SPIX->master_cfg & ~(MXC_F_SPIX_MASTER_CFG_ACT_DELAY |
        MXC_F_SPIX_MASTER_CFG_INACT_DELAY)) |
        (act_delay << MXC_F_SPIX_MASTER_CFG_ACT_DELAY_POS) |
        (inact_delay << MXC_F_SPIX_MASTER_CFG_INACT_DELAY_POS));
}

/******************************************************************************/
#if defined ( __GNUC__ )
__attribute__ ((section(".spix_config"), noinline))
#endif /* __GNUC */

#if IAR_SPIX_PRAGMA
#pragma location=".spix_config" // IAR locate function in RAM section .spix_config
#pragma optimize=no_inline      // IAR no inline optimization on this function
#endif /* IAR_SPIX_PRAGMA */

void SPIX_ConfigFetch(const spix_fetch_t *fetch)
{
    // Configure how the SPIX fetches data
    MXC_SPIX->fetch_ctrl = (((fetch->cmd << MXC_F_SPIX_FETCH_CTRL_CMD_VALUE_POS) & MXC_F_SPIX_FETCH_CTRL_CMD_VALUE) |
        ((fetch->cmd_width << MXC_F_SPIX_FETCH_CTRL_CMD_WIDTH_POS) & MXC_F_SPIX_FETCH_CTRL_CMD_WIDTH) |
        ((fetch->addr_width << MXC_F_SPIX_FETCH_CTRL_ADDR_WIDTH_POS) & MXC_F_SPIX_FETCH_CTRL_ADDR_WIDTH) |
        ((fetch->data_width << MXC_F_SPIX_FETCH_CTRL_DATA_WIDTH_POS) & MXC_F_SPIX_FETCH_CTRL_DATA_WIDTH) |
        ((fetch->addr_size << MXC_F_SPIX_FETCH_CTRL_FOUR_BYTE_ADDR_POS) & MXC_F_SPIX_FETCH_CTRL_FOUR_BYTE_ADDR));

    // Set the command mode and clocks
    MXC_SPIX->mode_ctrl = (((fetch->mode_clocks << MXC_F_SPIX_MODE_CTRL_MODE_CLOCKS_POS) & MXC_F_SPIX_MODE_CTRL_MODE_CLOCKS) |
        (!!fetch->no_cmd_mode << MXC_F_SPIX_MODE_CTRL_NO_CMD_MODE_POS));

    MXC_SPIX->mode_data = (((fetch->mode_data << MXC_F_SPIX_MODE_DATA_MODE_DATA_BITS_POS) & MXC_F_SPIX_MODE_DATA_MODE_DATA_BITS) |
        MXC_F_SPIX_MODE_DATA_MODE_DATA_OE);

#if(MXC_SPIX_REV == 0)
    SPIX_UpdateFBIgnore();
#endif
}

/******************************************************************************/
#if defined ( __GNUC__ )
__attribute__ ((section(".spix_config"), noinline))
#endif /* __GNUC */

#if IAR_SPIX_PRAGMA
#pragma location=".spix_config" // IAR locate function in RAM section .spix_config
#pragma optimize=no_inline      // IAR no inline optimization on this function
#endif /* IAR_SPIX_PRAGMA */

int SPIX_Shutdown(mxc_spix_regs_t *spix)
{
    int err;

    // Clear system level configurations
    if ((err = SYS_SPIX_Shutdown()) != E_NO_ERROR) {
        return err;
    }

    return E_NO_ERROR;
}
