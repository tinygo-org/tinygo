/*******************************************************************************
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
 * $Date: 2016-03-11 11:46:37 -0600 (Fri, 11 Mar 2016) $
 * $Revision: 21839 $
 *
 ******************************************************************************/

#ifndef _MXC_SPIX_REGS_H_
#define _MXC_SPIX_REGS_H_

#ifdef __cplusplus
extern "C" {
#endif

#include <stdint.h>

/*
    If types are not defined elsewhere (CMSIS) define them here
*/
#ifndef __IO
#define __IO volatile
#endif
#ifndef __I
#define __I  volatile const
#endif
#ifndef __O
#define __O  volatile
#endif
#ifndef __R
#define __R  volatile const
#endif


/*
   Typedefed structure(s) for module registers (per instance or section) with direct 32-bit
   access to each register in module.
*/

/*                                                          Offset          Register Description
                                                            =============   ============================================================================ */
typedef struct {
    __IO uint32_t master_cfg;                           /*  0x0000          SPIX Master Configuration                                                    */
    __IO uint32_t fetch_ctrl;                           /*  0x0004          SPIX Fetch Control                                                           */
    __IO uint32_t mode_ctrl;                            /*  0x0008          SPIX Mode Control                                                            */
    __IO uint32_t mode_data;                            /*  0x000C          SPIX Mode Data                                                               */
    __IO uint32_t sck_fb_ctrl;                          /*  0x0010          SPIX SCK_FB Control Register                                                 */
} mxc_spix_regs_t;


/*
   Register offsets for module SPIX.
*/

#define MXC_R_SPIX_OFFS_MASTER_CFG                          ((uint32_t)0x00000000UL)
#define MXC_R_SPIX_OFFS_FETCH_CTRL                          ((uint32_t)0x00000004UL)
#define MXC_R_SPIX_OFFS_MODE_CTRL                           ((uint32_t)0x00000008UL)
#define MXC_R_SPIX_OFFS_MODE_DATA                           ((uint32_t)0x0000000CUL)
#define MXC_R_SPIX_OFFS_SCK_FB_CTRL                         ((uint32_t)0x00000010UL)


/*
   Field positions and masks for module SPIX.
*/

#define MXC_F_SPIX_MASTER_CFG_SPI_MODE_POS                  0
#define MXC_F_SPIX_MASTER_CFG_SPI_MODE                      ((uint32_t)(0x00000003UL << MXC_F_SPIX_MASTER_CFG_SPI_MODE_POS))
#define MXC_F_SPIX_MASTER_CFG_SS_ACT_LO_POS                 2
#define MXC_F_SPIX_MASTER_CFG_SS_ACT_LO                     ((uint32_t)(0x00000001UL << MXC_F_SPIX_MASTER_CFG_SS_ACT_LO_POS))
#define MXC_F_SPIX_MASTER_CFG_ALT_TIMING_EN_POS             3
#define MXC_F_SPIX_MASTER_CFG_ALT_TIMING_EN                 ((uint32_t)(0x00000001UL << MXC_F_SPIX_MASTER_CFG_ALT_TIMING_EN_POS))
#define MXC_F_SPIX_MASTER_CFG_SLAVE_SEL_POS                 4
#define MXC_F_SPIX_MASTER_CFG_SLAVE_SEL                     ((uint32_t)(0x00000007UL << MXC_F_SPIX_MASTER_CFG_SLAVE_SEL_POS))
#define MXC_F_SPIX_MASTER_CFG_SCK_LO_CLK_POS                8
#define MXC_F_SPIX_MASTER_CFG_SCK_LO_CLK                    ((uint32_t)(0x0000000FUL << MXC_F_SPIX_MASTER_CFG_SCK_LO_CLK_POS))
#define MXC_F_SPIX_MASTER_CFG_SCK_HI_CLK_POS                12
#define MXC_F_SPIX_MASTER_CFG_SCK_HI_CLK                    ((uint32_t)(0x0000000FUL << MXC_F_SPIX_MASTER_CFG_SCK_HI_CLK_POS))
#define MXC_F_SPIX_MASTER_CFG_ACT_DELAY_POS                 16
#define MXC_F_SPIX_MASTER_CFG_ACT_DELAY                     ((uint32_t)(0x00000003UL << MXC_F_SPIX_MASTER_CFG_ACT_DELAY_POS))
#define MXC_F_SPIX_MASTER_CFG_INACT_DELAY_POS               18
#define MXC_F_SPIX_MASTER_CFG_INACT_DELAY                   ((uint32_t)(0x00000003UL << MXC_F_SPIX_MASTER_CFG_INACT_DELAY_POS))
#define MXC_F_SPIX_MASTER_CFG_ALT_SCK_LO_CLK_POS            20
#define MXC_F_SPIX_MASTER_CFG_ALT_SCK_LO_CLK                ((uint32_t)(0x0000000FUL << MXC_F_SPIX_MASTER_CFG_ALT_SCK_LO_CLK_POS))
#define MXC_F_SPIX_MASTER_CFG_ALT_SCK_HI_CLK_POS            24
#define MXC_F_SPIX_MASTER_CFG_ALT_SCK_HI_CLK                ((uint32_t)(0x0000000FUL << MXC_F_SPIX_MASTER_CFG_ALT_SCK_HI_CLK_POS))
#define MXC_F_SPIX_MASTER_CFG_SDIO_SAMPLE_POINT_POS         28
#define MXC_F_SPIX_MASTER_CFG_SDIO_SAMPLE_POINT             ((uint32_t)(0x0000000FUL << MXC_F_SPIX_MASTER_CFG_SDIO_SAMPLE_POINT_POS))

#define MXC_F_SPIX_FETCH_CTRL_CMD_VALUE_POS                 0
#define MXC_F_SPIX_FETCH_CTRL_CMD_VALUE                     ((uint32_t)(0x000000FFUL << MXC_F_SPIX_FETCH_CTRL_CMD_VALUE_POS))
#define MXC_F_SPIX_FETCH_CTRL_CMD_WIDTH_POS                 8
#define MXC_F_SPIX_FETCH_CTRL_CMD_WIDTH                     ((uint32_t)(0x00000003UL << MXC_F_SPIX_FETCH_CTRL_CMD_WIDTH_POS))
#define MXC_F_SPIX_FETCH_CTRL_ADDR_WIDTH_POS                10
#define MXC_F_SPIX_FETCH_CTRL_ADDR_WIDTH                    ((uint32_t)(0x00000003UL << MXC_F_SPIX_FETCH_CTRL_ADDR_WIDTH_POS))
#define MXC_F_SPIX_FETCH_CTRL_DATA_WIDTH_POS                12
#define MXC_F_SPIX_FETCH_CTRL_DATA_WIDTH                    ((uint32_t)(0x00000003UL << MXC_F_SPIX_FETCH_CTRL_DATA_WIDTH_POS))
#define MXC_F_SPIX_FETCH_CTRL_FOUR_BYTE_ADDR_POS            16
#define MXC_F_SPIX_FETCH_CTRL_FOUR_BYTE_ADDR                ((uint32_t)(0x00000001UL << MXC_F_SPIX_FETCH_CTRL_FOUR_BYTE_ADDR_POS))

#define MXC_F_SPIX_MODE_CTRL_MODE_CLOCKS_POS                0
#define MXC_F_SPIX_MODE_CTRL_MODE_CLOCKS                    ((uint32_t)(0x0000000FUL << MXC_F_SPIX_MODE_CTRL_MODE_CLOCKS_POS))
#define MXC_F_SPIX_MODE_CTRL_NO_CMD_MODE_POS                8
#define MXC_F_SPIX_MODE_CTRL_NO_CMD_MODE                    ((uint32_t)(0x00000001UL << MXC_F_SPIX_MODE_CTRL_NO_CMD_MODE_POS))

#define MXC_F_SPIX_MODE_DATA_MODE_DATA_BITS_POS             0
#define MXC_F_SPIX_MODE_DATA_MODE_DATA_BITS                 ((uint32_t)(0x0000FFFFUL << MXC_F_SPIX_MODE_DATA_MODE_DATA_BITS_POS))
#define MXC_F_SPIX_MODE_DATA_MODE_DATA_OE_POS               16
#define MXC_F_SPIX_MODE_DATA_MODE_DATA_OE                   ((uint32_t)(0x0000FFFFUL << MXC_F_SPIX_MODE_DATA_MODE_DATA_OE_POS))

#define MXC_F_SPIX_SCK_FB_CTRL_ENABLE_SCK_FB_MODE_POS       0
#define MXC_F_SPIX_SCK_FB_CTRL_ENABLE_SCK_FB_MODE           ((uint32_t)(0x00000001UL << MXC_F_SPIX_SCK_FB_CTRL_ENABLE_SCK_FB_MODE_POS))
#define MXC_F_SPIX_SCK_FB_CTRL_INVERT_SCK_FB_CLK_POS        1
#define MXC_F_SPIX_SCK_FB_CTRL_INVERT_SCK_FB_CLK            ((uint32_t)(0x00000001UL << MXC_F_SPIX_SCK_FB_CTRL_INVERT_SCK_FB_CLK_POS))

#if(MXC_SPIX_REV == 0)
#define MXC_F_SPIX_SCK_FB_CTRL_IGNORE_CLKS_POS              4
#define MXC_F_SPIX_SCK_FB_CTRL_IGNORE_CLKS                  ((uint32_t)(0x0000003FUL << MXC_F_SPIX_SCK_FB_CTRL_IGNORE_CLKS_POS))
#define MXC_F_SPIX_SCK_FB_CTRL_IGNORE_CLKS_NO_CMD_POS       12
#define MXC_F_SPIX_SCK_FB_CTRL_IGNORE_CLKS_NO_CMD           ((uint32_t)(0x0000003FUL << MXC_F_SPIX_SCK_FB_CTRL_IGNORE_CLKS_NO_CMD_POS))
#endif


/*
   Field values and shifted values for module SPIX.
*/

#define MXC_V_SPIX_MASTER_CFG_SPI_MODE_SCK_HI_SAMPLE_RISING                     ((uint32_t)(0x00000000UL))
#define MXC_V_SPIX_MASTER_CFG_SPI_MODE_SCK_LO_SAMPLE_FALLING                    ((uint32_t)(0x00000003UL))

#define MXC_S_SPIX_MASTER_CFG_SPI_MODE_SCK_HI_SAMPLE_RISING                     ((uint32_t)(MXC_V_SPIX_MASTER_CFG_SPI_MODE_SCK_HI_SAMPLE_RISING    << MXC_F_SPIX_MASTER_CFG_SPI_MODE_POS))
#define MXC_S_SPIX_MASTER_CFG_SPI_MODE_SCK_LO_SAMPLE_FALLING                    ((uint32_t)(MXC_V_SPIX_MASTER_CFG_SPI_MODE_SCK_LO_SAMPLE_FALLING   << MXC_F_SPIX_MASTER_CFG_SPI_MODE_POS))

#define MXC_V_SPIX_MASTER_CFG_SS_ACT_LO_ACTIVE_HIGH                             ((uint32_t)(0x00000000UL))
#define MXC_V_SPIX_MASTER_CFG_SS_ACT_LO_ACTIVE_LOW                              ((uint32_t)(0x00000001UL))

#define MXC_S_SPIX_MASTER_CFG_SS_ACT_LO_ACTIVE_HIGH                             ((uint32_t)(MXC_V_SPIX_MASTER_CFG_SS_ACT_LO_ACTIVE_HIGH  << MXC_F_SPIX_MASTER_CFG_SS_ACT_LO_POS))
#define MXC_S_SPIX_MASTER_CFG_SS_ACT_LO_ACTIVE_LOW                              ((uint32_t)(MXC_V_SPIX_MASTER_CFG_SS_ACT_LO_ACTIVE_LOW   << MXC_F_SPIX_MASTER_CFG_SS_ACT_LO_POS))

#define MXC_V_SPIX_MASTER_CFG_ALT_TIMING_EN_DISABLED                            ((uint32_t)(0x00000000UL))
#define MXC_V_SPIX_MASTER_CFG_ALT_TIMING_EN_ENABLED_AS_NEEDED                   ((uint32_t)(0x00000001UL))

#define MXC_S_SPIX_MASTER_CFG_ALT_TIMING_EN_DISABLED                            ((uint32_t)(MXC_V_SPIX_MASTER_CFG_ALT_TIMING_EN_DISABLED            << MXC_F_SPIX_MASTER_CFG_ALT_TIMING_EN_POS))
#define MXC_S_SPIX_MASTER_CFG_ALT_TIMING_EN_ENABLED_AS_NEEDED                   ((uint32_t)(MXC_V_SPIX_MASTER_CFG_ALT_TIMING_EN_ENABLED_AS_NEEDED   << MXC_F_SPIX_MASTER_CFG_ALT_TIMING_EN_POS))

#define MXC_V_SPIX_MASTER_CFG_ACT_DELAY_OFF                                     ((uint32_t)(0x00000000UL))
#define MXC_V_SPIX_MASTER_CFG_ACT_DELAY_FOR_2_MOD_CLK                           ((uint32_t)(0x00000001UL))
#define MXC_V_SPIX_MASTER_CFG_ACT_DELAY_FOR_4_MOD_CLK                           ((uint32_t)(0x00000002UL))
#define MXC_V_SPIX_MASTER_CFG_ACT_DELAY_FOR_8_MOD_CLK                           ((uint32_t)(0x00000003UL))

#define MXC_S_SPIX_MASTER_CFG_ACT_DELAY_OFF                                     ((uint32_t)(MXC_V_SPIX_MASTER_CFG_ACT_DELAY_OFF             << MXC_F_SPIX_MASTER_CFG_ACT_DELAY_POS))
#define MXC_S_SPIX_MASTER_CFG_ACT_DELAY_FOR_2_MOD_CLK                           ((uint32_t)(MXC_V_SPIX_MASTER_CFG_ACT_DELAY_FOR_2_MOD_CLK   << MXC_F_SPIX_MASTER_CFG_ACT_DELAY_POS))
#define MXC_S_SPIX_MASTER_CFG_ACT_DELAY_FOR_4_MOD_CLK                           ((uint32_t)(MXC_V_SPIX_MASTER_CFG_ACT_DELAY_FOR_4_MOD_CLK   << MXC_F_SPIX_MASTER_CFG_ACT_DELAY_POS))
#define MXC_S_SPIX_MASTER_CFG_ACT_DELAY_FOR_8_MOD_CLK                           ((uint32_t)(MXC_V_SPIX_MASTER_CFG_ACT_DELAY_FOR_8_MOD_CLK   << MXC_F_SPIX_MASTER_CFG_ACT_DELAY_POS))

#define MXC_V_SPIX_MASTER_CFG_INACT_DELAY_OFF                                   ((uint32_t)(0x00000000UL))
#define MXC_V_SPIX_MASTER_CFG_INACT_DELAY_FOR_2_MOD_CLK                         ((uint32_t)(0x00000001UL))
#define MXC_V_SPIX_MASTER_CFG_INACT_DELAY_FOR_4_MOD_CLK                         ((uint32_t)(0x00000002UL))
#define MXC_V_SPIX_MASTER_CFG_INACT_DELAY_FOR_8_MOD_CLK                         ((uint32_t)(0x00000003UL))

#define MXC_S_SPIX_MASTER_CFG_INACT_DELAY_OFF                                   ((uint32_t)(MXC_V_SPIX_MASTER_CFG_INACT_DELAY_OFF             << MXC_F_SPIX_MASTER_CFG_INACT_DELAY_POS))
#define MXC_S_SPIX_MASTER_CFG_INACT_DELAY_FOR_2_MOD_CLK                         ((uint32_t)(MXC_V_SPIX_MASTER_CFG_INACT_DELAY_FOR_2_MOD_CLK   << MXC_F_SPIX_MASTER_CFG_INACT_DELAY_POS))
#define MXC_S_SPIX_MASTER_CFG_INACT_DELAY_FOR_4_MOD_CLK                         ((uint32_t)(MXC_V_SPIX_MASTER_CFG_INACT_DELAY_FOR_4_MOD_CLK   << MXC_F_SPIX_MASTER_CFG_INACT_DELAY_POS))
#define MXC_S_SPIX_MASTER_CFG_INACT_DELAY_FOR_8_MOD_CLK                         ((uint32_t)(MXC_V_SPIX_MASTER_CFG_INACT_DELAY_FOR_8_MOD_CLK   << MXC_F_SPIX_MASTER_CFG_INACT_DELAY_POS))

#define MXC_V_SPIX_FETCH_CTRL_CMD_WIDTH_SINGLE                                  ((uint32_t)(0x00000000UL))
#define MXC_V_SPIX_FETCH_CTRL_CMD_WIDTH_DUAL_IO                                 ((uint32_t)(0x00000001UL))
#define MXC_V_SPIX_FETCH_CTRL_CMD_WIDTH_QUAD_IO                                 ((uint32_t)(0x00000002UL))

#define MXC_S_SPIX_FETCH_CTRL_CMD_WIDTH_SINGLE                                  ((uint32_t)(MXC_V_SPIX_FETCH_CTRL_CMD_WIDTH_SINGLE    << MXC_F_SPIX_FETCH_CTRL_CMD_WIDTH_POS))
#define MXC_S_SPIX_FETCH_CTRL_CMD_WIDTH_DUAL_IO                                 ((uint32_t)(MXC_V_SPIX_FETCH_CTRL_CMD_WIDTH_DUAL_IO   << MXC_F_SPIX_FETCH_CTRL_CMD_WIDTH_POS))
#define MXC_S_SPIX_FETCH_CTRL_CMD_WIDTH_QUAD_IO                                 ((uint32_t)(MXC_V_SPIX_FETCH_CTRL_CMD_WIDTH_QUAD_IO   << MXC_F_SPIX_FETCH_CTRL_CMD_WIDTH_POS))

#define MXC_V_SPIX_FETCH_CTRL_ADDR_WIDTH_SINGLE                                 ((uint32_t)(0x00000000UL))
#define MXC_V_SPIX_FETCH_CTRL_ADDR_WIDTH_DUAL_IO                                ((uint32_t)(0x00000001UL))
#define MXC_V_SPIX_FETCH_CTRL_ADDR_WIDTH_QUAD_IO                                ((uint32_t)(0x00000002UL))

#define MXC_S_SPIX_FETCH_CTRL_ADDR_WIDTH_SINGLE                                 ((uint32_t)(MXC_V_SPIX_FETCH_CTRL_ADDR_WIDTH_SINGLE    << MXC_F_SPIX_FETCH_CTRL_ADDR_WIDTH_POS))
#define MXC_S_SPIX_FETCH_CTRL_ADDR_WIDTH_DUAL_IO                                ((uint32_t)(MXC_V_SPIX_FETCH_CTRL_ADDR_WIDTH_DUAL_IO   << MXC_F_SPIX_FETCH_CTRL_ADDR_WIDTH_POS))
#define MXC_S_SPIX_FETCH_CTRL_ADDR_WIDTH_QUAD_IO                                ((uint32_t)(MXC_V_SPIX_FETCH_CTRL_ADDR_WIDTH_QUAD_IO   << MXC_F_SPIX_FETCH_CTRL_ADDR_WIDTH_POS))

#define MXC_V_SPIX_FETCH_CTRL_DATA_WIDTH_SINGLE                                 ((uint32_t)(0x00000000UL))
#define MXC_V_SPIX_FETCH_CTRL_DATA_WIDTH_DUAL_IO                                ((uint32_t)(0x00000001UL))
#define MXC_V_SPIX_FETCH_CTRL_DATA_WIDTH_QUAD_IO                                ((uint32_t)(0x00000002UL))

#define MXC_S_SPIX_FETCH_CTRL_DATA_WIDTH_SINGLE                                 ((uint32_t)(MXC_V_SPIX_FETCH_CTRL_DATA_WIDTH_SINGLE    << MXC_F_SPIX_FETCH_CTRL_DATA_WIDTH_POS))
#define MXC_S_SPIX_FETCH_CTRL_DATA_WIDTH_DUAL_IO                                ((uint32_t)(MXC_V_SPIX_FETCH_CTRL_DATA_WIDTH_DUAL_IO   << MXC_F_SPIX_FETCH_CTRL_DATA_WIDTH_POS))
#define MXC_S_SPIX_FETCH_CTRL_DATA_WIDTH_QUAD_IO                                ((uint32_t)(MXC_V_SPIX_FETCH_CTRL_DATA_WIDTH_QUAD_IO   << MXC_F_SPIX_FETCH_CTRL_DATA_WIDTH_POS))



#ifdef __cplusplus
}
#endif

#endif   /* _MXC_SPIX_REGS_H_ */

