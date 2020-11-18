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

#ifndef _MXC_SPIM_REGS_H_
#define _MXC_SPIM_REGS_H_

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
    __IO uint32_t mstr_cfg;                             /*  0x0000          SPI Master Configuration Register                                            */
    __IO uint32_t ss_sr_polarity;                       /*  0x0004          SPI Master Polarity Control for SS and SR Signals                            */
    __IO uint32_t gen_ctrl;                             /*  0x0008          SPI Master General Control Register                                          */
    __IO uint32_t fifo_ctrl;                            /*  0x000C          SPI Master FIFO Control Register                                             */
    __IO uint32_t spcl_ctrl;                            /*  0x0010          SPI Master Special Mode Controls                                             */
    __IO uint32_t intfl;                                /*  0x0014          SPI Master Interrupt Flags                                                   */
    __IO uint32_t inten;                                /*  0x0018          SPI Master Interrupt Enable/Disable Settings                                 */
    __IO uint32_t simple_headers;                       /*  0x001C          SPI Master Simple Mode Transaction Headers                                   */
} mxc_spim_regs_t;


/*                                                          Offset          Register Description
                                                            =============   ============================================================================ */
typedef struct {
    union {                                             /*  0x0000-0x07FC   SPI Master FIFO Write Space for Transaction Setup                            */
        __IO uint8_t  trans_8[2048];
        __IO uint16_t trans_16[1024];
        __IO uint32_t trans_32[512];
    };
    union {                                             /*  0x0800-0x0FFC   SPI Master FIFO Read Space for Results Data                                  */
        __IO uint8_t  rslts_8[2048];
        __IO uint16_t rslts_16[1024];
        __IO uint32_t rslts_32[512];
    };
} mxc_spim_fifo_regs_t;


/*
   Register offsets for module SPIM.
*/

#define MXC_R_SPIM_OFFS_MSTR_CFG                            ((uint32_t)0x00000000UL)
#define MXC_R_SPIM_OFFS_SS_SR_POLARITY                      ((uint32_t)0x00000004UL)
#define MXC_R_SPIM_OFFS_GEN_CTRL                            ((uint32_t)0x00000008UL)
#define MXC_R_SPIM_OFFS_FIFO_CTRL                           ((uint32_t)0x0000000CUL)
#define MXC_R_SPIM_OFFS_SPCL_CTRL                           ((uint32_t)0x00000010UL)
#define MXC_R_SPIM_OFFS_INTFL                               ((uint32_t)0x00000014UL)
#define MXC_R_SPIM_OFFS_INTEN                               ((uint32_t)0x00000018UL)
#define MXC_R_SPIM_OFFS_SIMPLE_HEADERS                      ((uint32_t)0x0000001CUL)
#define MXC_R_SPIM_FIFO_OFFS_TRANS                          ((uint32_t)0x00000000UL)
#define MXC_R_SPIM_FIFO_OFFS_RSLTS                          ((uint32_t)0x00000800UL)


/*
   Field positions and masks for module SPIM.
*/

#define MXC_F_SPIM_MSTR_CFG_SLAVE_SEL_POS                   0
#define MXC_F_SPIM_MSTR_CFG_SLAVE_SEL                       ((uint32_t)(0x00000007UL << MXC_F_SPIM_MSTR_CFG_SLAVE_SEL_POS))
#define MXC_F_SPIM_MSTR_CFG_THREE_WIRE_MODE_POS             3
#define MXC_F_SPIM_MSTR_CFG_THREE_WIRE_MODE                 ((uint32_t)(0x00000001UL << MXC_F_SPIM_MSTR_CFG_THREE_WIRE_MODE_POS))
#define MXC_F_SPIM_MSTR_CFG_SPI_MODE_POS                    4
#define MXC_F_SPIM_MSTR_CFG_SPI_MODE                        ((uint32_t)(0x00000003UL << MXC_F_SPIM_MSTR_CFG_SPI_MODE_POS))
#define MXC_F_SPIM_MSTR_CFG_PAGE_SIZE_POS                   6
#define MXC_F_SPIM_MSTR_CFG_PAGE_SIZE                       ((uint32_t)(0x00000003UL << MXC_F_SPIM_MSTR_CFG_PAGE_SIZE_POS))
#define MXC_F_SPIM_MSTR_CFG_SCK_HI_CLK_POS                  8
#define MXC_F_SPIM_MSTR_CFG_SCK_HI_CLK                      ((uint32_t)(0x0000000FUL << MXC_F_SPIM_MSTR_CFG_SCK_HI_CLK_POS))
#define MXC_F_SPIM_MSTR_CFG_SCK_LO_CLK_POS                  12
#define MXC_F_SPIM_MSTR_CFG_SCK_LO_CLK                      ((uint32_t)(0x0000000FUL << MXC_F_SPIM_MSTR_CFG_SCK_LO_CLK_POS))
#define MXC_F_SPIM_MSTR_CFG_ACT_DELAY_POS                   16
#define MXC_F_SPIM_MSTR_CFG_ACT_DELAY                       ((uint32_t)(0x00000003UL << MXC_F_SPIM_MSTR_CFG_ACT_DELAY_POS))
#define MXC_F_SPIM_MSTR_CFG_INACT_DELAY_POS                 18
#define MXC_F_SPIM_MSTR_CFG_INACT_DELAY                     ((uint32_t)(0x00000003UL << MXC_F_SPIM_MSTR_CFG_INACT_DELAY_POS))
#define MXC_F_SPIM_MSTR_CFG_SDIO_SAMPLE_POINT_POS           20
#define MXC_F_SPIM_MSTR_CFG_SDIO_SAMPLE_POINT               ((uint32_t)(0x0000000FUL << MXC_F_SPIM_MSTR_CFG_SDIO_SAMPLE_POINT_POS))

#define MXC_V_SPIM_MSTR_CFG_PAGE_SIZE_4B                    ((uint32_t)0x00000000UL)
#define MXC_V_SPIM_MSTR_CFG_PAGE_SIZE_8B                    ((uint32_t)0x00000001UL)
#define MXC_V_SPIM_MSTR_CFG_PAGE_SIZE_16B                   ((uint32_t)0x00000002UL)
#define MXC_V_SPIM_MSTR_CFG_PAGE_SIZE_32B                   ((uint32_t)0x00000003UL)

#define MXC_S_SPIM_MSTR_CFG_PAGE_4B                         (MXC_V_SPIM_MSTR_CFG_PAGE_SIZE_4B  << MXC_F_SPIM_MSTR_CFG_PAGE_SIZE_POS)
#define MXC_S_SPIM_MSTR_CFG_PAGE_8B                         (MXC_V_SPIM_MSTR_CFG_PAGE_SIZE_8B  << MXC_F_SPIM_MSTR_CFG_PAGE_SIZE_POS)
#define MXC_S_SPIM_MSTR_CFG_PAGE_16B                        (MXC_V_SPIM_MSTR_CFG_PAGE_SIZE_16B << MXC_F_SPIM_MSTR_CFG_PAGE_SIZE_POS)
#define MXC_S_SPIM_MSTR_CFG_PAGE_32B                        (MXC_V_SPIM_MSTR_CFG_PAGE_SIZE_32B << MXC_F_SPIM_MSTR_CFG_PAGE_SIZE_POS)

#define MXC_F_SPIM_SS_SR_POLARITY_SS_POLARITY_POS           0
#define MXC_F_SPIM_SS_SR_POLARITY_SS_POLARITY               ((uint32_t)(0x000000FFUL << MXC_F_SPIM_SS_SR_POLARITY_SS_POLARITY_POS))
#define MXC_F_SPIM_SS_SR_POLARITY_FC_POLARITY_POS           8
#define MXC_F_SPIM_SS_SR_POLARITY_FC_POLARITY               ((uint32_t)(0x000000FFUL << MXC_F_SPIM_SS_SR_POLARITY_FC_POLARITY_POS))

#define MXC_F_SPIM_GEN_CTRL_SPI_MSTR_EN_POS                 0
#define MXC_F_SPIM_GEN_CTRL_SPI_MSTR_EN                     ((uint32_t)(0x00000001UL << MXC_F_SPIM_GEN_CTRL_SPI_MSTR_EN_POS))
#define MXC_F_SPIM_GEN_CTRL_TX_FIFO_EN_POS                  1
#define MXC_F_SPIM_GEN_CTRL_TX_FIFO_EN                      ((uint32_t)(0x00000001UL << MXC_F_SPIM_GEN_CTRL_TX_FIFO_EN_POS))
#define MXC_F_SPIM_GEN_CTRL_RX_FIFO_EN_POS                  2
#define MXC_F_SPIM_GEN_CTRL_RX_FIFO_EN                      ((uint32_t)(0x00000001UL << MXC_F_SPIM_GEN_CTRL_RX_FIFO_EN_POS))
#define MXC_F_SPIM_GEN_CTRL_BIT_BANG_MODE_POS               3
#define MXC_F_SPIM_GEN_CTRL_BIT_BANG_MODE                   ((uint32_t)(0x00000001UL << MXC_F_SPIM_GEN_CTRL_BIT_BANG_MODE_POS))
#define MXC_F_SPIM_GEN_CTRL_BB_SS_IN_OUT_POS                4
#define MXC_F_SPIM_GEN_CTRL_BB_SS_IN_OUT                    ((uint32_t)(0x00000001UL << MXC_F_SPIM_GEN_CTRL_BB_SS_IN_OUT_POS))
#define MXC_F_SPIM_GEN_CTRL_BB_SR_IN_POS                    5
#define MXC_F_SPIM_GEN_CTRL_BB_SR_IN                        ((uint32_t)(0x00000001UL << MXC_F_SPIM_GEN_CTRL_BB_SR_IN_POS))
#define MXC_F_SPIM_GEN_CTRL_BB_SCK_IN_OUT_POS               6
#define MXC_F_SPIM_GEN_CTRL_BB_SCK_IN_OUT                   ((uint32_t)(0x00000001UL << MXC_F_SPIM_GEN_CTRL_BB_SCK_IN_OUT_POS))
#define MXC_F_SPIM_GEN_CTRL_BB_SDIO_IN_POS                  8
#define MXC_F_SPIM_GEN_CTRL_BB_SDIO_IN                      ((uint32_t)(0x0000000FUL << MXC_F_SPIM_GEN_CTRL_BB_SDIO_IN_POS))
#define MXC_F_SPIM_GEN_CTRL_BB_SDIO_OUT_POS                 12
#define MXC_F_SPIM_GEN_CTRL_BB_SDIO_OUT                     ((uint32_t)(0x0000000FUL << MXC_F_SPIM_GEN_CTRL_BB_SDIO_OUT_POS))
#define MXC_F_SPIM_GEN_CTRL_BB_SDIO_DR_EN_POS               16
#define MXC_F_SPIM_GEN_CTRL_BB_SDIO_DR_EN                   ((uint32_t)(0x0000000FUL << MXC_F_SPIM_GEN_CTRL_BB_SDIO_DR_EN_POS))
#define MXC_F_SPIM_GEN_CTRL_SIMPLE_MODE_POS                 20
#define MXC_F_SPIM_GEN_CTRL_SIMPLE_MODE                     ((uint32_t)(0x00000001UL << MXC_F_SPIM_GEN_CTRL_SIMPLE_MODE_POS))
#define MXC_F_SPIM_GEN_CTRL_START_RX_ONLY_POS               21
#define MXC_F_SPIM_GEN_CTRL_START_RX_ONLY                   ((uint32_t)(0x00000001UL << MXC_F_SPIM_GEN_CTRL_START_RX_ONLY_POS))
#define MXC_F_SPIM_GEN_CTRL_DEASSERT_ACT_SS_POS             22
#define MXC_F_SPIM_GEN_CTRL_DEASSERT_ACT_SS                 ((uint32_t)(0x00000001UL << MXC_F_SPIM_GEN_CTRL_DEASSERT_ACT_SS_POS))
#define MXC_F_SPIM_GEN_CTRL_ENABLE_SCK_FB_MODE_POS          24
#define MXC_F_SPIM_GEN_CTRL_ENABLE_SCK_FB_MODE              ((uint32_t)(0x00000001UL << MXC_F_SPIM_GEN_CTRL_ENABLE_SCK_FB_MODE_POS))
#define MXC_F_SPIM_GEN_CTRL_INVERT_SCK_FB_CLK_POS           25
#define MXC_F_SPIM_GEN_CTRL_INVERT_SCK_FB_CLK               ((uint32_t)(0x00000001UL << MXC_F_SPIM_GEN_CTRL_INVERT_SCK_FB_CLK_POS))

#define MXC_F_SPIM_FIFO_CTRL_TX_FIFO_AE_LVL_POS             0
#define MXC_F_SPIM_FIFO_CTRL_TX_FIFO_AE_LVL                 ((uint32_t)(0x0000000FUL << MXC_F_SPIM_FIFO_CTRL_TX_FIFO_AE_LVL_POS))
#define MXC_F_SPIM_FIFO_CTRL_TX_FIFO_USED_POS               8
#define MXC_F_SPIM_FIFO_CTRL_TX_FIFO_USED                   ((uint32_t)(0x0000001FUL << MXC_F_SPIM_FIFO_CTRL_TX_FIFO_USED_POS))
#define MXC_F_SPIM_FIFO_CTRL_RX_FIFO_AF_LVL_POS             16
#define MXC_F_SPIM_FIFO_CTRL_RX_FIFO_AF_LVL                 ((uint32_t)(0x0000001FUL << MXC_F_SPIM_FIFO_CTRL_RX_FIFO_AF_LVL_POS))
#define MXC_F_SPIM_FIFO_CTRL_RX_FIFO_USED_POS               24
#define MXC_F_SPIM_FIFO_CTRL_RX_FIFO_USED                   ((uint32_t)(0x0000003FUL << MXC_F_SPIM_FIFO_CTRL_RX_FIFO_USED_POS))

#define MXC_F_SPIM_SPCL_CTRL_SS_SAMPLE_MODE_POS             0
#define MXC_F_SPIM_SPCL_CTRL_SS_SAMPLE_MODE                 ((uint32_t)(0x00000001UL << MXC_F_SPIM_SPCL_CTRL_SS_SAMPLE_MODE_POS))
#define MXC_F_SPIM_SPCL_CTRL_MISO_FC_EN_POS                 1
#define MXC_F_SPIM_SPCL_CTRL_MISO_FC_EN                     ((uint32_t)(0x00000001UL << MXC_F_SPIM_SPCL_CTRL_MISO_FC_EN_POS))
#define MXC_F_SPIM_SPCL_CTRL_SS_SA_SDIO_OUT_POS             4
#define MXC_F_SPIM_SPCL_CTRL_SS_SA_SDIO_OUT                 ((uint32_t)(0x0000000FUL << MXC_F_SPIM_SPCL_CTRL_SS_SA_SDIO_OUT_POS))
#define MXC_F_SPIM_SPCL_CTRL_SS_SA_SDIO_DR_EN_POS           8
#define MXC_F_SPIM_SPCL_CTRL_SS_SA_SDIO_DR_EN               ((uint32_t)(0x0000000FUL << MXC_F_SPIM_SPCL_CTRL_SS_SA_SDIO_DR_EN_POS))

#if (MXC_SPIM_REV == 0)
#define MXC_F_SPIM_SPCL_CTRL_SPECIAL_MODE_3_EN_POS          16
#define MXC_F_SPIM_SPCL_CTRL_SPECIAL_MODE_3_EN              ((uint32_t)(0x00000001UL << MXC_F_SPIM_SPCL_CTRL_SPECIAL_MODE_3_EN_POS))
#else
#define MXC_F_SPIM_SPCL_CTRL_RX_FIFO_MARGIN_POS             12
#define MXC_F_SPIM_SPCL_CTRL_RX_FIFO_MARGIN                 ((uint32_t)(0x00000007UL << MXC_F_SPIM_SPCL_CTRL_RX_FIFO_MARGIN_POS))
#define MXC_F_SPIM_SPCL_CTRL_SCK_FB_DELAY_POS               16
#define MXC_F_SPIM_SPCL_CTRL_SCK_FB_DELAY                   ((uint32_t)(0x0000000FUL << MXC_F_SPIM_SPCL_CTRL_SCK_FB_DELAY_POS))
#define MXC_F_SPIM_SPCL_CTRL_SPARE_RESERVED_POS             20
#define MXC_F_SPIM_SPCL_CTRL_SPARE_RESERVED                 ((uint32_t)(0x00000FFFUL << MXC_F_SPIM_SPCL_CTRL_SPARE_RESERVED_POS))
#endif

#define MXC_F_SPIM_INTFL_TX_STALLED_POS                     0
#define MXC_F_SPIM_INTFL_TX_STALLED                         ((uint32_t)(0x00000001UL << MXC_F_SPIM_INTFL_TX_STALLED_POS))
#define MXC_F_SPIM_INTFL_RX_STALLED_POS                     1
#define MXC_F_SPIM_INTFL_RX_STALLED                         ((uint32_t)(0x00000001UL << MXC_F_SPIM_INTFL_RX_STALLED_POS))
#define MXC_F_SPIM_INTFL_TX_READY_POS                       2
#define MXC_F_SPIM_INTFL_TX_READY                           ((uint32_t)(0x00000001UL << MXC_F_SPIM_INTFL_TX_READY_POS))
#define MXC_F_SPIM_INTFL_RX_DONE_POS                        3
#define MXC_F_SPIM_INTFL_RX_DONE                            ((uint32_t)(0x00000001UL << MXC_F_SPIM_INTFL_RX_DONE_POS))
#define MXC_F_SPIM_INTFL_TX_FIFO_AE_POS                     4
#define MXC_F_SPIM_INTFL_TX_FIFO_AE                         ((uint32_t)(0x00000001UL << MXC_F_SPIM_INTFL_TX_FIFO_AE_POS))
#define MXC_F_SPIM_INTFL_RX_FIFO_AF_POS                     5
#define MXC_F_SPIM_INTFL_RX_FIFO_AF                         ((uint32_t)(0x00000001UL << MXC_F_SPIM_INTFL_RX_FIFO_AF_POS))

#define MXC_F_SPIM_INTEN_TX_STALLED_POS                     0
#define MXC_F_SPIM_INTEN_TX_STALLED                         ((uint32_t)(0x00000001UL << MXC_F_SPIM_INTEN_TX_STALLED_POS))
#define MXC_F_SPIM_INTEN_RX_STALLED_POS                     1
#define MXC_F_SPIM_INTEN_RX_STALLED                         ((uint32_t)(0x00000001UL << MXC_F_SPIM_INTEN_RX_STALLED_POS))
#define MXC_F_SPIM_INTEN_TX_READY_POS                       2
#define MXC_F_SPIM_INTEN_TX_READY                           ((uint32_t)(0x00000001UL << MXC_F_SPIM_INTEN_TX_READY_POS))
#define MXC_F_SPIM_INTEN_RX_DONE_POS                        3
#define MXC_F_SPIM_INTEN_RX_DONE                            ((uint32_t)(0x00000001UL << MXC_F_SPIM_INTEN_RX_DONE_POS))
#define MXC_F_SPIM_INTEN_TX_FIFO_AE_POS                     4
#define MXC_F_SPIM_INTEN_TX_FIFO_AE                         ((uint32_t)(0x00000001UL << MXC_F_SPIM_INTEN_TX_FIFO_AE_POS))
#define MXC_F_SPIM_INTEN_RX_FIFO_AF_POS                     5
#define MXC_F_SPIM_INTEN_RX_FIFO_AF                         ((uint32_t)(0x00000001UL << MXC_F_SPIM_INTEN_RX_FIFO_AF_POS))

#define MXC_F_SPIM_SIMPLE_HEADERS_TX_BIDIR_HEADER_POS       0
#define MXC_F_SPIM_SIMPLE_HEADERS_TX_BIDIR_HEADER           ((uint32_t)(0x00003FFFUL << MXC_F_SPIM_SIMPLE_HEADERS_TX_BIDIR_HEADER_POS))
#define MXC_F_SPIM_SIMPLE_HEADERS_RX_ONLY_HEADER_POS        16
#define MXC_F_SPIM_SIMPLE_HEADERS_RX_ONLY_HEADER            ((uint32_t)(0x00003FFFUL << MXC_F_SPIM_SIMPLE_HEADERS_RX_ONLY_HEADER_POS))



#ifdef __cplusplus
}
#endif

#endif   /* _MXC_SPIM_REGS_H_ */

