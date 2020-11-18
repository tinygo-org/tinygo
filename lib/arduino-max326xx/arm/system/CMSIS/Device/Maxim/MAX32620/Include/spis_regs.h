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
 * $Date: 2016-06-03 14:21:38 -0500 (Fri, 03 Jun 2016) $
 * $Revision: 23194 $
 *
 ******************************************************************************/

#ifndef _MXC_SPIS_REGS_H_
#define _MXC_SPIS_REGS_H_

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
    __IO uint32_t gen_ctrl;                             /*  0x0000          SPI Slave General Control Register                                           */
    __IO uint32_t fifo_ctrl;                            /*  0x0004          SPI Slave FIFO Control Register                                              */
    __IO uint32_t fifo_stat;                            /*  0x0008          SPI Slave FIFO Status Register                                               */
    __IO uint32_t intfl;                                /*  0x000C          SPI Slave Interrupt Flags                                                    */
    __IO uint32_t inten;                                /*  0x0010          SPI Slave Interrupt Enable/Disable Settings                                  */
} mxc_spis_regs_t;


/*                                                          Offset          Register Description
                                                            =============   ============================================================================ */
typedef struct {
    union {                                             /*  0x0000-0x07FC   SPI Slave FIFO TX Write Space                                                */
        __IO uint8_t  tx_8[2048];
        __IO uint16_t tx_16[1024];
        __IO uint32_t tx_32[512];
    };
    union {                                             /*  0x0800-0x0FFC   SPI Slave FIFO RX Read Space                                                 */
        __IO uint8_t  rx_8[2048];
        __IO uint16_t rx_16[1024];
        __IO uint32_t rx_32[512];
    };
} mxc_spis_fifo_regs_t;


/*
   Register offsets for module SPIS.
*/

#define MXC_R_SPIS_OFFS_GEN_CTRL                            ((uint32_t)0x00000000UL)
#define MXC_R_SPIS_OFFS_FIFO_CTRL                           ((uint32_t)0x00000004UL)
#define MXC_R_SPIS_OFFS_FIFO_STAT                           ((uint32_t)0x00000008UL)
#define MXC_R_SPIS_OFFS_INTFL                               ((uint32_t)0x0000000CUL)
#define MXC_R_SPIS_OFFS_INTEN                               ((uint32_t)0x00000010UL)
#define MXC_R_SPIS_FIFO_OFFS_TX                             ((uint32_t)0x00000000UL)
#define MXC_R_SPIS_FIFO_OFFS_RX                             ((uint32_t)0x00000800UL)


/*
   Field positions and masks for module SPIS.
*/

#define MXC_F_SPIS_GEN_CTRL_SPI_SLAVE_EN_POS                0
#define MXC_F_SPIS_GEN_CTRL_SPI_SLAVE_EN                    ((uint32_t)(0x00000001UL << MXC_F_SPIS_GEN_CTRL_SPI_SLAVE_EN_POS))
#define MXC_F_SPIS_GEN_CTRL_TX_FIFO_EN_POS                  1
#define MXC_F_SPIS_GEN_CTRL_TX_FIFO_EN                      ((uint32_t)(0x00000001UL << MXC_F_SPIS_GEN_CTRL_TX_FIFO_EN_POS))
#define MXC_F_SPIS_GEN_CTRL_RX_FIFO_EN_POS                  2
#define MXC_F_SPIS_GEN_CTRL_RX_FIFO_EN                      ((uint32_t)(0x00000001UL << MXC_F_SPIS_GEN_CTRL_RX_FIFO_EN_POS))
#define MXC_F_SPIS_GEN_CTRL_DATA_WIDTH_POS                  4
#define MXC_F_SPIS_GEN_CTRL_DATA_WIDTH                      ((uint32_t)(0x00000003UL << MXC_F_SPIS_GEN_CTRL_DATA_WIDTH_POS))
#define MXC_F_SPIS_GEN_CTRL_SPI_MODE_POS                    16
#define MXC_F_SPIS_GEN_CTRL_SPI_MODE                        ((uint32_t)(0x00000003UL << MXC_F_SPIS_GEN_CTRL_SPI_MODE_POS))
#define MXC_F_SPIS_GEN_CTRL_TX_CLK_INVERT_POS               20
#define MXC_F_SPIS_GEN_CTRL_TX_CLK_INVERT                   ((uint32_t)(0x00000001UL << MXC_F_SPIS_GEN_CTRL_TX_CLK_INVERT_POS))

#define MXC_F_SPIS_FIFO_CTRL_TX_FIFO_AE_LVL_POS             0
#define MXC_F_SPIS_FIFO_CTRL_TX_FIFO_AE_LVL                 ((uint32_t)(0x0000001FUL << MXC_F_SPIS_FIFO_CTRL_TX_FIFO_AE_LVL_POS))
#define MXC_F_SPIS_FIFO_CTRL_RX_FIFO_AF_LVL_POS             8
#define MXC_F_SPIS_FIFO_CTRL_RX_FIFO_AF_LVL                 ((uint32_t)(0x0000001FUL << MXC_F_SPIS_FIFO_CTRL_RX_FIFO_AF_LVL_POS))

#define MXC_F_SPIS_FIFO_STAT_TX_FIFO_USED_POS               0
#define MXC_F_SPIS_FIFO_STAT_TX_FIFO_USED                   ((uint32_t)(0x0000003FUL << MXC_F_SPIS_FIFO_STAT_TX_FIFO_USED_POS))
#define MXC_F_SPIS_FIFO_STAT_RX_FIFO_USED_POS               8
#define MXC_F_SPIS_FIFO_STAT_RX_FIFO_USED                   ((uint32_t)(0x0000003FUL << MXC_F_SPIS_FIFO_STAT_RX_FIFO_USED_POS))

#define MXC_F_SPIS_INTFL_TX_FIFO_AE_POS                     0
#define MXC_F_SPIS_INTFL_TX_FIFO_AE                         ((uint32_t)(0x00000001UL << MXC_F_SPIS_INTFL_TX_FIFO_AE_POS))
#define MXC_F_SPIS_INTFL_RX_FIFO_AF_POS                     1
#define MXC_F_SPIS_INTFL_RX_FIFO_AF                         ((uint32_t)(0x00000001UL << MXC_F_SPIS_INTFL_RX_FIFO_AF_POS))
#define MXC_F_SPIS_INTFL_TX_NO_DATA_POS                     2
#define MXC_F_SPIS_INTFL_TX_NO_DATA                         ((uint32_t)(0x00000001UL << MXC_F_SPIS_INTFL_TX_NO_DATA_POS))
#define MXC_F_SPIS_INTFL_RX_LOST_DATA_POS                   3
#define MXC_F_SPIS_INTFL_RX_LOST_DATA                       ((uint32_t)(0x00000001UL << MXC_F_SPIS_INTFL_RX_LOST_DATA_POS))
#define MXC_F_SPIS_INTFL_TX_UNDERFLOW_POS                   4
#define MXC_F_SPIS_INTFL_TX_UNDERFLOW                       ((uint32_t)(0x00000001UL << MXC_F_SPIS_INTFL_TX_UNDERFLOW_POS))
#define MXC_F_SPIS_INTFL_SS_ASSERTED_POS                    5
#define MXC_F_SPIS_INTFL_SS_ASSERTED                        ((uint32_t)(0x00000001UL << MXC_F_SPIS_INTFL_SS_ASSERTED_POS))
#define MXC_F_SPIS_INTFL_SS_DEASSERTED_POS                  6
#define MXC_F_SPIS_INTFL_SS_DEASSERTED                      ((uint32_t)(0x00000001UL << MXC_F_SPIS_INTFL_SS_DEASSERTED_POS))

#define MXC_F_SPIS_INTEN_TX_FIFO_AE_POS                     0
#define MXC_F_SPIS_INTEN_TX_FIFO_AE                         ((uint32_t)(0x00000001UL << MXC_F_SPIS_INTEN_TX_FIFO_AE_POS))
#define MXC_F_SPIS_INTEN_RX_FIFO_AF_POS                     1
#define MXC_F_SPIS_INTEN_RX_FIFO_AF                         ((uint32_t)(0x00000001UL << MXC_F_SPIS_INTEN_RX_FIFO_AF_POS))
#define MXC_F_SPIS_INTEN_TX_NO_DATA_POS                     2
#define MXC_F_SPIS_INTEN_TX_NO_DATA                         ((uint32_t)(0x00000001UL << MXC_F_SPIS_INTEN_TX_NO_DATA_POS))
#define MXC_F_SPIS_INTEN_RX_LOST_DATA_POS                   3
#define MXC_F_SPIS_INTEN_RX_LOST_DATA                       ((uint32_t)(0x00000001UL << MXC_F_SPIS_INTEN_RX_LOST_DATA_POS))
#define MXC_F_SPIS_INTEN_TX_UNDERFLOW_POS                   4
#define MXC_F_SPIS_INTEN_TX_UNDERFLOW                       ((uint32_t)(0x00000001UL << MXC_F_SPIS_INTEN_TX_UNDERFLOW_POS))
#define MXC_F_SPIS_INTEN_SS_ASSERTED_POS                    5
#define MXC_F_SPIS_INTEN_SS_ASSERTED                        ((uint32_t)(0x00000001UL << MXC_F_SPIS_INTEN_SS_ASSERTED_POS))
#define MXC_F_SPIS_INTEN_SS_DEASSERTED_POS                  6
#define MXC_F_SPIS_INTEN_SS_DEASSERTED                      ((uint32_t)(0x00000001UL << MXC_F_SPIS_INTEN_SS_DEASSERTED_POS))

#ifdef __cplusplus
}
#endif

#endif   /* _MXC_SPIS_REGS_H_ */

