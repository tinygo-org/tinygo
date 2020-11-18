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
 * $Date: 2016-04-06 08:18:01 -0500 (Wed, 06 Apr 2016) $
 * $Revision: 22236 $
 *
 ******************************************************************************/

#ifndef _MXC_OWM_REGS_H_
#define _MXC_OWM_REGS_H_

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
    __IO uint32_t cfg;                                  /*  0x0000          1-Wire Master Configuration                                                  */
    __IO uint32_t clk_div_1us;                          /*  0x0004          1-Wire Master Clock Divisor                                                  */
    __IO uint32_t ctrl_stat;                            /*  0x0008          1-Wire Master Control/Status                                                 */
    __IO uint32_t data;                                 /*  0x000C          1-Wire Master Data Buffer                                                    */
    __IO uint32_t intfl;                                /*  0x0010          1-Wire Master Interrupt Flags                                                */
    __IO uint32_t inten;                                /*  0x0014          1-Wire Master Interrupt Enables                                              */
} mxc_owm_regs_t;


/*
   Register offsets for module OWM.
*/

#define MXC_R_OWM_OFFS_CFG                                  ((uint32_t)0x00000000UL)
#define MXC_R_OWM_OFFS_CLK_DIV_1US                          ((uint32_t)0x00000004UL)
#define MXC_R_OWM_OFFS_CTRL_STAT                            ((uint32_t)0x00000008UL)
#define MXC_R_OWM_OFFS_DATA                                 ((uint32_t)0x0000000CUL)
#define MXC_R_OWM_OFFS_INTFL                                ((uint32_t)0x00000010UL)
#define MXC_R_OWM_OFFS_INTEN                                ((uint32_t)0x00000014UL)


/*
   Field positions and masks for module OWM.
*/

#define MXC_F_OWM_CFG_LONG_LINE_MODE_POS                    0
#define MXC_F_OWM_CFG_LONG_LINE_MODE                        ((uint32_t)(0x00000001UL << MXC_F_OWM_CFG_LONG_LINE_MODE_POS))
#define MXC_F_OWM_CFG_FORCE_PRES_DET_POS                    1
#define MXC_F_OWM_CFG_FORCE_PRES_DET                        ((uint32_t)(0x00000001UL << MXC_F_OWM_CFG_FORCE_PRES_DET_POS))
#define MXC_F_OWM_CFG_BIT_BANG_EN_POS                       2
#define MXC_F_OWM_CFG_BIT_BANG_EN                           ((uint32_t)(0x00000001UL << MXC_F_OWM_CFG_BIT_BANG_EN_POS))
#define MXC_F_OWM_CFG_EXT_PULLUP_MODE_POS                   3
#define MXC_F_OWM_CFG_EXT_PULLUP_MODE                       ((uint32_t)(0x00000001UL << MXC_F_OWM_CFG_EXT_PULLUP_MODE_POS))
#define MXC_F_OWM_CFG_EXT_PULLUP_ENABLE_POS                 4
#define MXC_F_OWM_CFG_EXT_PULLUP_ENABLE                     ((uint32_t)(0x00000001UL << MXC_F_OWM_CFG_EXT_PULLUP_ENABLE_POS))
#define MXC_F_OWM_CFG_SINGLE_BIT_MODE_POS                   5
#define MXC_F_OWM_CFG_SINGLE_BIT_MODE                       ((uint32_t)(0x00000001UL << MXC_F_OWM_CFG_SINGLE_BIT_MODE_POS))
#define MXC_F_OWM_CFG_OVERDRIVE_POS                         6
#define MXC_F_OWM_CFG_OVERDRIVE                             ((uint32_t)(0x00000001UL << MXC_F_OWM_CFG_OVERDRIVE_POS))
#define MXC_F_OWM_CFG_INT_PULLUP_ENABLE_POS                 7
#define MXC_F_OWM_CFG_INT_PULLUP_ENABLE                     ((uint32_t)(0x00000001UL << MXC_F_OWM_CFG_INT_PULLUP_ENABLE_POS))

#define MXC_F_OWM_CLK_DIV_1US_DIVISOR_POS                   0
#define MXC_F_OWM_CLK_DIV_1US_DIVISOR                       ((uint32_t)(0x000000FFUL << MXC_F_OWM_CLK_DIV_1US_DIVISOR_POS))

#define MXC_F_OWM_CTRL_STAT_START_OW_RESET_POS              0
#define MXC_F_OWM_CTRL_STAT_START_OW_RESET                  ((uint32_t)(0x00000001UL << MXC_F_OWM_CTRL_STAT_START_OW_RESET_POS))
#define MXC_F_OWM_CTRL_STAT_SRA_MODE_POS                    1
#define MXC_F_OWM_CTRL_STAT_SRA_MODE                        ((uint32_t)(0x00000001UL << MXC_F_OWM_CTRL_STAT_SRA_MODE_POS))
#define MXC_F_OWM_CTRL_STAT_BIT_BANG_OE_POS                 2
#define MXC_F_OWM_CTRL_STAT_BIT_BANG_OE                     ((uint32_t)(0x00000001UL << MXC_F_OWM_CTRL_STAT_BIT_BANG_OE_POS))
#define MXC_F_OWM_CTRL_STAT_OW_INPUT_POS                    3
#define MXC_F_OWM_CTRL_STAT_OW_INPUT                        ((uint32_t)(0x00000001UL << MXC_F_OWM_CTRL_STAT_OW_INPUT_POS))
#define MXC_F_OWM_CTRL_STAT_OD_SPEC_MODE_POS                4
#define MXC_F_OWM_CTRL_STAT_OD_SPEC_MODE                    ((uint32_t)(0x00000001UL << MXC_F_OWM_CTRL_STAT_OD_SPEC_MODE_POS))
#define MXC_F_OWM_CTRL_STAT_EXT_PULLUP_POL_POS              5
#define MXC_F_OWM_CTRL_STAT_EXT_PULLUP_POL                  ((uint32_t)(0x00000001UL << MXC_F_OWM_CTRL_STAT_EXT_PULLUP_POL_POS))
#define MXC_F_OWM_CTRL_STAT_PRESENCE_DETECT_POS             7
#define MXC_F_OWM_CTRL_STAT_PRESENCE_DETECT                 ((uint32_t)(0x00000001UL << MXC_F_OWM_CTRL_STAT_PRESENCE_DETECT_POS))

#define MXC_F_OWM_DATA_TX_RX_POS                            0
#define MXC_F_OWM_DATA_TX_RX                                ((uint32_t)(0x000000FFUL << MXC_F_OWM_DATA_TX_RX_POS))

#define MXC_F_OWM_INTFL_OW_RESET_DONE_POS                   0
#define MXC_F_OWM_INTFL_OW_RESET_DONE                       ((uint32_t)(0x00000001UL << MXC_F_OWM_INTFL_OW_RESET_DONE_POS))
#define MXC_F_OWM_INTFL_TX_DATA_EMPTY_POS                   1
#define MXC_F_OWM_INTFL_TX_DATA_EMPTY                       ((uint32_t)(0x00000001UL << MXC_F_OWM_INTFL_TX_DATA_EMPTY_POS))
#define MXC_F_OWM_INTFL_RX_DATA_READY_POS                   2
#define MXC_F_OWM_INTFL_RX_DATA_READY                       ((uint32_t)(0x00000001UL << MXC_F_OWM_INTFL_RX_DATA_READY_POS))
#define MXC_F_OWM_INTFL_LINE_SHORT_POS                      3
#define MXC_F_OWM_INTFL_LINE_SHORT                          ((uint32_t)(0x00000001UL << MXC_F_OWM_INTFL_LINE_SHORT_POS))
#define MXC_F_OWM_INTFL_LINE_LOW_POS                        4
#define MXC_F_OWM_INTFL_LINE_LOW                            ((uint32_t)(0x00000001UL << MXC_F_OWM_INTFL_LINE_LOW_POS))

#define MXC_F_OWM_INTEN_OW_RESET_DONE_POS                   0
#define MXC_F_OWM_INTEN_OW_RESET_DONE                       ((uint32_t)(0x00000001UL << MXC_F_OWM_INTEN_OW_RESET_DONE_POS))
#define MXC_F_OWM_INTEN_TX_DATA_EMPTY_POS                   1
#define MXC_F_OWM_INTEN_TX_DATA_EMPTY                       ((uint32_t)(0x00000001UL << MXC_F_OWM_INTEN_TX_DATA_EMPTY_POS))
#define MXC_F_OWM_INTEN_RX_DATA_READY_POS                   2
#define MXC_F_OWM_INTEN_RX_DATA_READY                       ((uint32_t)(0x00000001UL << MXC_F_OWM_INTEN_RX_DATA_READY_POS))
#define MXC_F_OWM_INTEN_LINE_SHORT_POS                      3
#define MXC_F_OWM_INTEN_LINE_SHORT                          ((uint32_t)(0x00000001UL << MXC_F_OWM_INTEN_LINE_SHORT_POS))
#define MXC_F_OWM_INTEN_LINE_LOW_POS                        4
#define MXC_F_OWM_INTEN_LINE_LOW                            ((uint32_t)(0x00000001UL << MXC_F_OWM_INTEN_LINE_LOW_POS))

#define MXC_V_OWM_CFG_EXT_PULLUP_MODE_UNUSED         		((uint32_t)(0x00000000UL))
#define MXC_V_OWM_CFG_EXT_PULLUP_MODE_USED          		((uint32_t)(0x00000001UL))

#define MXC_V_OWM_CTRL_STAT_OD_SPEC_MODE_12US               ((uint32_t)(0x00000000UL))
#define MXC_V_OWM_CTRL_STAT_OD_SPEC_MODE_10US               ((uint32_t)(0x00000001UL))

#define MXC_V_OWM_CTRL_STAT_EXT_PULLUP_POL_ACT_HIGH         ((uint32_t)(0x00000000UL))
#define MXC_V_OWM_CTRL_STAT_EXT_PULLUP_POL_ACT_LOW          ((uint32_t)(0x00000001UL))


#ifdef __cplusplus
}
#endif

#endif   /* _MXC_OWM_REGS_H_ */

