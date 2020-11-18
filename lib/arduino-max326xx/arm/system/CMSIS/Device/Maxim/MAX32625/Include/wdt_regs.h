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

#ifndef _MXC_WDT_REGS_H_
#define _MXC_WDT_REGS_H_

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
    __IO uint32_t ctrl;                                 /*  0x0000          Watchdog Timer Control Register                                              */
    __IO uint32_t clear;                                /*  0x0004          Watchdog Timer Clear Register (Feed Dog)                                     */
    __IO uint32_t flags;                                /*  0x0008          Watchdog Timer Interrupt and Reset Flags                                     */
    __IO uint32_t enable;                               /*  0x000C          Watchdog Timer Interrupt/Reset Enable/Disable Controls                       */
    __R  uint32_t rsv010;                               /*  0x0010                                                                                       */
    __IO uint32_t lock_ctrl;                            /*  0x0014          Watchdog Timer Register Setting Lock for Control Register                    */
} mxc_wdt_regs_t;


/*
   Register offsets for module WDT.
*/

#define MXC_R_WDT_OFFS_CTRL                                 ((uint32_t)0x00000000UL)
#define MXC_R_WDT_OFFS_CLEAR                                ((uint32_t)0x00000004UL)
#define MXC_R_WDT_OFFS_FLAGS                                ((uint32_t)0x00000008UL)
#define MXC_R_WDT_OFFS_ENABLE                               ((uint32_t)0x0000000CUL)
#define MXC_R_WDT_OFFS_LOCK_CTRL                            ((uint32_t)0x00000014UL)


/*
   Field positions and masks for module WDT.
*/

#define MXC_F_WDT_CTRL_INT_PERIOD_POS                       0
#define MXC_F_WDT_CTRL_INT_PERIOD                           ((uint32_t)(0x0000000FUL << MXC_F_WDT_CTRL_INT_PERIOD_POS))
#define MXC_F_WDT_CTRL_RST_PERIOD_POS                       4
#define MXC_F_WDT_CTRL_RST_PERIOD                           ((uint32_t)(0x0000000FUL << MXC_F_WDT_CTRL_RST_PERIOD_POS))
#define MXC_F_WDT_CTRL_EN_TIMER_POS                         8
#define MXC_F_WDT_CTRL_EN_TIMER                             ((uint32_t)(0x00000001UL << MXC_F_WDT_CTRL_EN_TIMER_POS))
#define MXC_F_WDT_CTRL_EN_CLOCK_POS                         9
#define MXC_F_WDT_CTRL_EN_CLOCK                             ((uint32_t)(0x00000001UL << MXC_F_WDT_CTRL_EN_CLOCK_POS))
#define MXC_F_WDT_CTRL_WAIT_PERIOD_POS                      12
#define MXC_F_WDT_CTRL_WAIT_PERIOD                          ((uint32_t)(0x0000000FUL << MXC_F_WDT_CTRL_WAIT_PERIOD_POS))

#define MXC_F_WDT_FLAGS_TIMEOUT_POS                         0
#define MXC_F_WDT_FLAGS_TIMEOUT                             ((uint32_t)(0x00000001UL << MXC_F_WDT_FLAGS_TIMEOUT_POS))
#define MXC_F_WDT_FLAGS_PRE_WIN_POS                         1
#define MXC_F_WDT_FLAGS_PRE_WIN                             ((uint32_t)(0x00000001UL << MXC_F_WDT_FLAGS_PRE_WIN_POS))
#define MXC_F_WDT_FLAGS_RESET_OUT_POS                       2
#define MXC_F_WDT_FLAGS_RESET_OUT                           ((uint32_t)(0x00000001UL << MXC_F_WDT_FLAGS_RESET_OUT_POS))

#define MXC_F_WDT_ENABLE_TIMEOUT_POS                        0
#define MXC_F_WDT_ENABLE_TIMEOUT                            ((uint32_t)(0x00000001UL << MXC_F_WDT_ENABLE_TIMEOUT_POS))
#define MXC_F_WDT_ENABLE_PRE_WIN_POS                        1
#define MXC_F_WDT_ENABLE_PRE_WIN                            ((uint32_t)(0x00000001UL << MXC_F_WDT_ENABLE_PRE_WIN_POS))
#define MXC_F_WDT_ENABLE_RESET_OUT_POS                      2
#define MXC_F_WDT_ENABLE_RESET_OUT                          ((uint32_t)(0x00000001UL << MXC_F_WDT_ENABLE_RESET_OUT_POS))

#define MXC_F_WDT_LOCK_CTRL_WDLOCK_POS                      0
#define MXC_F_WDT_LOCK_CTRL_WDLOCK                          ((uint32_t)(0x000000FFUL << MXC_F_WDT_LOCK_CTRL_WDLOCK_POS))



/*
   Field values and shifted values for module WDT.
*/

#define MXC_V_WDT_CTRL_INT_PERIOD_2_31_CLKS                                     ((uint32_t)(0x00000000UL))
#define MXC_V_WDT_CTRL_INT_PERIOD_2_30_CLKS                                     ((uint32_t)(0x00000001UL))
#define MXC_V_WDT_CTRL_INT_PERIOD_2_29_CLKS                                     ((uint32_t)(0x00000002UL))
#define MXC_V_WDT_CTRL_INT_PERIOD_2_28_CLKS                                     ((uint32_t)(0x00000003UL))
#define MXC_V_WDT_CTRL_INT_PERIOD_2_27_CLKS                                     ((uint32_t)(0x00000004UL))
#define MXC_V_WDT_CTRL_INT_PERIOD_2_26_CLKS                                     ((uint32_t)(0x00000005UL))
#define MXC_V_WDT_CTRL_INT_PERIOD_2_25_CLKS                                     ((uint32_t)(0x00000006UL))
#define MXC_V_WDT_CTRL_INT_PERIOD_2_24_CLKS                                     ((uint32_t)(0x00000007UL))
#define MXC_V_WDT_CTRL_INT_PERIOD_2_23_CLKS                                     ((uint32_t)(0x00000008UL))
#define MXC_V_WDT_CTRL_INT_PERIOD_2_22_CLKS                                     ((uint32_t)(0x00000009UL))
#define MXC_V_WDT_CTRL_INT_PERIOD_2_21_CLKS                                     ((uint32_t)(0x0000000AUL))
#define MXC_V_WDT_CTRL_INT_PERIOD_2_20_CLKS                                     ((uint32_t)(0x0000000BUL))
#define MXC_V_WDT_CTRL_INT_PERIOD_2_19_CLKS                                     ((uint32_t)(0x0000000CUL))
#define MXC_V_WDT_CTRL_INT_PERIOD_2_18_CLKS                                     ((uint32_t)(0x0000000DUL))
#define MXC_V_WDT_CTRL_INT_PERIOD_2_17_CLKS                                     ((uint32_t)(0x0000000EUL))
#define MXC_V_WDT_CTRL_INT_PERIOD_2_16_CLKS                                     ((uint32_t)(0x0000000FUL))

#define MXC_S_WDT_CTRL_INT_PERIOD_2_31_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_INT_PERIOD_2_31_CLKS   << MXC_F_WDT_CTRL_INT_PERIOD_POS))
#define MXC_S_WDT_CTRL_INT_PERIOD_2_30_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_INT_PERIOD_2_30_CLKS   << MXC_F_WDT_CTRL_INT_PERIOD_POS))
#define MXC_S_WDT_CTRL_INT_PERIOD_2_29_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_INT_PERIOD_2_29_CLKS   << MXC_F_WDT_CTRL_INT_PERIOD_POS))
#define MXC_S_WDT_CTRL_INT_PERIOD_2_28_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_INT_PERIOD_2_28_CLKS   << MXC_F_WDT_CTRL_INT_PERIOD_POS))
#define MXC_S_WDT_CTRL_INT_PERIOD_2_27_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_INT_PERIOD_2_27_CLKS   << MXC_F_WDT_CTRL_INT_PERIOD_POS))
#define MXC_S_WDT_CTRL_INT_PERIOD_2_26_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_INT_PERIOD_2_26_CLKS   << MXC_F_WDT_CTRL_INT_PERIOD_POS))
#define MXC_S_WDT_CTRL_INT_PERIOD_2_25_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_INT_PERIOD_2_25_CLKS   << MXC_F_WDT_CTRL_INT_PERIOD_POS))
#define MXC_S_WDT_CTRL_INT_PERIOD_2_24_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_INT_PERIOD_2_24_CLKS   << MXC_F_WDT_CTRL_INT_PERIOD_POS))
#define MXC_S_WDT_CTRL_INT_PERIOD_2_23_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_INT_PERIOD_2_23_CLKS   << MXC_F_WDT_CTRL_INT_PERIOD_POS))
#define MXC_S_WDT_CTRL_INT_PERIOD_2_22_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_INT_PERIOD_2_22_CLKS   << MXC_F_WDT_CTRL_INT_PERIOD_POS))
#define MXC_S_WDT_CTRL_INT_PERIOD_2_21_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_INT_PERIOD_2_21_CLKS   << MXC_F_WDT_CTRL_INT_PERIOD_POS))
#define MXC_S_WDT_CTRL_INT_PERIOD_2_20_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_INT_PERIOD_2_20_CLKS   << MXC_F_WDT_CTRL_INT_PERIOD_POS))
#define MXC_S_WDT_CTRL_INT_PERIOD_2_19_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_INT_PERIOD_2_19_CLKS   << MXC_F_WDT_CTRL_INT_PERIOD_POS))
#define MXC_S_WDT_CTRL_INT_PERIOD_2_18_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_INT_PERIOD_2_18_CLKS   << MXC_F_WDT_CTRL_INT_PERIOD_POS))
#define MXC_S_WDT_CTRL_INT_PERIOD_2_17_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_INT_PERIOD_2_17_CLKS   << MXC_F_WDT_CTRL_INT_PERIOD_POS))
#define MXC_S_WDT_CTRL_INT_PERIOD_2_16_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_INT_PERIOD_2_16_CLKS   << MXC_F_WDT_CTRL_INT_PERIOD_POS))

#define MXC_V_WDT_CTRL_RST_PERIOD_2_31_CLKS                                     ((uint32_t)(0x00000000UL))
#define MXC_V_WDT_CTRL_RST_PERIOD_2_30_CLKS                                     ((uint32_t)(0x00000001UL))
#define MXC_V_WDT_CTRL_RST_PERIOD_2_29_CLKS                                     ((uint32_t)(0x00000002UL))
#define MXC_V_WDT_CTRL_RST_PERIOD_2_28_CLKS                                     ((uint32_t)(0x00000003UL))
#define MXC_V_WDT_CTRL_RST_PERIOD_2_27_CLKS                                     ((uint32_t)(0x00000004UL))
#define MXC_V_WDT_CTRL_RST_PERIOD_2_26_CLKS                                     ((uint32_t)(0x00000005UL))
#define MXC_V_WDT_CTRL_RST_PERIOD_2_25_CLKS                                     ((uint32_t)(0x00000006UL))
#define MXC_V_WDT_CTRL_RST_PERIOD_2_24_CLKS                                     ((uint32_t)(0x00000007UL))
#define MXC_V_WDT_CTRL_RST_PERIOD_2_23_CLKS                                     ((uint32_t)(0x00000008UL))
#define MXC_V_WDT_CTRL_RST_PERIOD_2_22_CLKS                                     ((uint32_t)(0x00000009UL))
#define MXC_V_WDT_CTRL_RST_PERIOD_2_21_CLKS                                     ((uint32_t)(0x0000000AUL))
#define MXC_V_WDT_CTRL_RST_PERIOD_2_20_CLKS                                     ((uint32_t)(0x0000000BUL))
#define MXC_V_WDT_CTRL_RST_PERIOD_2_19_CLKS                                     ((uint32_t)(0x0000000CUL))
#define MXC_V_WDT_CTRL_RST_PERIOD_2_18_CLKS                                     ((uint32_t)(0x0000000DUL))
#define MXC_V_WDT_CTRL_RST_PERIOD_2_17_CLKS                                     ((uint32_t)(0x0000000EUL))
#define MXC_V_WDT_CTRL_RST_PERIOD_2_16_CLKS                                     ((uint32_t)(0x0000000FUL))

#define MXC_S_WDT_CTRL_RST_PERIOD_2_31_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_RST_PERIOD_2_31_CLKS   << MXC_F_WDT_CTRL_RST_PERIOD_POS))
#define MXC_S_WDT_CTRL_RST_PERIOD_2_30_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_RST_PERIOD_2_30_CLKS   << MXC_F_WDT_CTRL_RST_PERIOD_POS))
#define MXC_S_WDT_CTRL_RST_PERIOD_2_29_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_RST_PERIOD_2_29_CLKS   << MXC_F_WDT_CTRL_RST_PERIOD_POS))
#define MXC_S_WDT_CTRL_RST_PERIOD_2_28_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_RST_PERIOD_2_28_CLKS   << MXC_F_WDT_CTRL_RST_PERIOD_POS))
#define MXC_S_WDT_CTRL_RST_PERIOD_2_27_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_RST_PERIOD_2_27_CLKS   << MXC_F_WDT_CTRL_RST_PERIOD_POS))
#define MXC_S_WDT_CTRL_RST_PERIOD_2_26_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_RST_PERIOD_2_26_CLKS   << MXC_F_WDT_CTRL_RST_PERIOD_POS))
#define MXC_S_WDT_CTRL_RST_PERIOD_2_25_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_RST_PERIOD_2_25_CLKS   << MXC_F_WDT_CTRL_RST_PERIOD_POS))
#define MXC_S_WDT_CTRL_RST_PERIOD_2_24_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_RST_PERIOD_2_24_CLKS   << MXC_F_WDT_CTRL_RST_PERIOD_POS))
#define MXC_S_WDT_CTRL_RST_PERIOD_2_23_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_RST_PERIOD_2_23_CLKS   << MXC_F_WDT_CTRL_RST_PERIOD_POS))
#define MXC_S_WDT_CTRL_RST_PERIOD_2_22_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_RST_PERIOD_2_22_CLKS   << MXC_F_WDT_CTRL_RST_PERIOD_POS))
#define MXC_S_WDT_CTRL_RST_PERIOD_2_21_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_RST_PERIOD_2_21_CLKS   << MXC_F_WDT_CTRL_RST_PERIOD_POS))
#define MXC_S_WDT_CTRL_RST_PERIOD_2_20_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_RST_PERIOD_2_20_CLKS   << MXC_F_WDT_CTRL_RST_PERIOD_POS))
#define MXC_S_WDT_CTRL_RST_PERIOD_2_19_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_RST_PERIOD_2_19_CLKS   << MXC_F_WDT_CTRL_RST_PERIOD_POS))
#define MXC_S_WDT_CTRL_RST_PERIOD_2_18_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_RST_PERIOD_2_18_CLKS   << MXC_F_WDT_CTRL_RST_PERIOD_POS))
#define MXC_S_WDT_CTRL_RST_PERIOD_2_17_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_RST_PERIOD_2_17_CLKS   << MXC_F_WDT_CTRL_RST_PERIOD_POS))
#define MXC_S_WDT_CTRL_RST_PERIOD_2_16_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_RST_PERIOD_2_16_CLKS   << MXC_F_WDT_CTRL_RST_PERIOD_POS))

#define MXC_V_WDT_CTRL_WAIT_PERIOD_2_31_CLKS                                    ((uint32_t)(0x00000000UL))
#define MXC_V_WDT_CTRL_WAIT_PERIOD_2_30_CLKS                                    ((uint32_t)(0x00000001UL))
#define MXC_V_WDT_CTRL_WAIT_PERIOD_2_29_CLKS                                    ((uint32_t)(0x00000002UL))
#define MXC_V_WDT_CTRL_WAIT_PERIOD_2_28_CLKS                                    ((uint32_t)(0x00000003UL))
#define MXC_V_WDT_CTRL_WAIT_PERIOD_2_27_CLKS                                    ((uint32_t)(0x00000004UL))
#define MXC_V_WDT_CTRL_WAIT_PERIOD_2_26_CLKS                                    ((uint32_t)(0x00000005UL))
#define MXC_V_WDT_CTRL_WAIT_PERIOD_2_25_CLKS                                    ((uint32_t)(0x00000006UL))
#define MXC_V_WDT_CTRL_WAIT_PERIOD_2_24_CLKS                                    ((uint32_t)(0x00000007UL))
#define MXC_V_WDT_CTRL_WAIT_PERIOD_2_23_CLKS                                    ((uint32_t)(0x00000008UL))
#define MXC_V_WDT_CTRL_WAIT_PERIOD_2_22_CLKS                                    ((uint32_t)(0x00000009UL))
#define MXC_V_WDT_CTRL_WAIT_PERIOD_2_21_CLKS                                    ((uint32_t)(0x0000000AUL))
#define MXC_V_WDT_CTRL_WAIT_PERIOD_2_20_CLKS                                    ((uint32_t)(0x0000000BUL))
#define MXC_V_WDT_CTRL_WAIT_PERIOD_2_19_CLKS                                    ((uint32_t)(0x0000000CUL))
#define MXC_V_WDT_CTRL_WAIT_PERIOD_2_18_CLKS                                    ((uint32_t)(0x0000000DUL))
#define MXC_V_WDT_CTRL_WAIT_PERIOD_2_17_CLKS                                    ((uint32_t)(0x0000000EUL))
#define MXC_V_WDT_CTRL_WAIT_PERIOD_2_16_CLKS                                    ((uint32_t)(0x0000000FUL))

#define MXC_S_WDT_CTRL_WAIT_PERIOD_2_31_CLKS                                    ((uint32_t)(MXC_V_WDT_CTRL_WAIT_PERIOD_2_31_CLKS   << MXC_F_WDT_CTRL_WAIT_PERIOD_POS))
#define MXC_S_WDT_CTRL_WAIT_PERIOD_2_30_CLKS                                    ((uint32_t)(MXC_V_WDT_CTRL_WAIT_PERIOD_2_30_CLKS   << MXC_F_WDT_CTRL_WAIT_PERIOD_POS))
#define MXC_S_WDT_CTRL_WAIT_PERIOD_2_29_CLKS                                    ((uint32_t)(MXC_V_WDT_CTRL_WAIT_PERIOD_2_29_CLKS   << MXC_F_WDT_CTRL_WAIT_PERIOD_POS))
#define MXC_S_WDT_CTRL_WAIT_PERIOD_2_28_CLKS                                    ((uint32_t)(MXC_V_WDT_CTRL_WAIT_PERIOD_2_28_CLKS   << MXC_F_WDT_CTRL_WAIT_PERIOD_POS))
#define MXC_S_WDT_CTRL_WAIT_PERIOD_2_27_CLKS                                    ((uint32_t)(MXC_V_WDT_CTRL_WAIT_PERIOD_2_27_CLKS   << MXC_F_WDT_CTRL_WAIT_PERIOD_POS))
#define MXC_S_WDT_CTRL_WAIT_PERIOD_2_26_CLKS                                    ((uint32_t)(MXC_V_WDT_CTRL_WAIT_PERIOD_2_26_CLKS   << MXC_F_WDT_CTRL_WAIT_PERIOD_POS))
#define MXC_S_WDT_CTRL_WAIT_PERIOD_2_25_CLKS                                    ((uint32_t)(MXC_V_WDT_CTRL_WAIT_PERIOD_2_25_CLKS   << MXC_F_WDT_CTRL_WAIT_PERIOD_POS))
#define MXC_S_WDT_CTRL_WAIT_PERIOD_2_24_CLKS                                    ((uint32_t)(MXC_V_WDT_CTRL_WAIT_PERIOD_2_24_CLKS   << MXC_F_WDT_CTRL_WAIT_PERIOD_POS))
#define MXC_S_WDT_CTRL_WAIT_PERIOD_2_23_CLKS                                    ((uint32_t)(MXC_V_WDT_CTRL_WAIT_PERIOD_2_23_CLKS   << MXC_F_WDT_CTRL_WAIT_PERIOD_POS))
#define MXC_S_WDT_CTRL_WAIT_PERIOD_2_22_CLKS                                    ((uint32_t)(MXC_V_WDT_CTRL_WAIT_PERIOD_2_22_CLKS   << MXC_F_WDT_CTRL_WAIT_PERIOD_POS))
#define MXC_S_WDT_CTRL_WAIT_PERIOD_2_21_CLKS                                    ((uint32_t)(MXC_V_WDT_CTRL_WAIT_PERIOD_2_21_CLKS   << MXC_F_WDT_CTRL_WAIT_PERIOD_POS))
#define MXC_S_WDT_CTRL_WAIT_PERIOD_2_20_CLKS                                    ((uint32_t)(MXC_V_WDT_CTRL_WAIT_PERIOD_2_20_CLKS   << MXC_F_WDT_CTRL_WAIT_PERIOD_POS))
#define MXC_S_WDT_CTRL_WAIT_PERIOD_2_19_CLKS                                    ((uint32_t)(MXC_V_WDT_CTRL_WAIT_PERIOD_2_19_CLKS   << MXC_F_WDT_CTRL_WAIT_PERIOD_POS))
#define MXC_S_WDT_CTRL_WAIT_PERIOD_2_18_CLKS                                    ((uint32_t)(MXC_V_WDT_CTRL_WAIT_PERIOD_2_18_CLKS   << MXC_F_WDT_CTRL_WAIT_PERIOD_POS))
#define MXC_S_WDT_CTRL_WAIT_PERIOD_2_17_CLKS                                    ((uint32_t)(MXC_V_WDT_CTRL_WAIT_PERIOD_2_17_CLKS   << MXC_F_WDT_CTRL_WAIT_PERIOD_POS))
#define MXC_S_WDT_CTRL_WAIT_PERIOD_2_16_CLKS                                    ((uint32_t)(MXC_V_WDT_CTRL_WAIT_PERIOD_2_16_CLKS   << MXC_F_WDT_CTRL_WAIT_PERIOD_POS))


#define MXC_V_WDT_LOCK_KEY    0x24
#define MXC_V_WDT_UNLOCK_KEY  0x42

#define MXC_V_WDT_RESET_KEY_0 0xA5
#define MXC_V_WDT_RESET_KEY_1 0x5A


#ifdef __cplusplus
}
#endif

#endif   /* _MXC_WDT_REGS_H_ */

