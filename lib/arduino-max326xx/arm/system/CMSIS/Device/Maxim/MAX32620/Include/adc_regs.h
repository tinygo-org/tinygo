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
 * $Date: 2016-03-18 09:21:00 -0500 (Fri, 18 Mar 2016) $
 * $Revision: 21976 $
 *
 ******************************************************************************/

#ifndef _MXC_ADC_REGS_H_
#define _MXC_ADC_REGS_H_

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
    __IO uint32_t ctrl;                                 /*  0x0000          ADC Control                                                                  */
    __IO uint32_t status;                               /*  0x0004          ADC Status                                                                   */
    __IO uint32_t data;                                 /*  0x0008          ADC Output Data                                                              */
    __IO uint32_t intr;                                 /*  0x000C          ADC Interrupt Control Register                                               */
    __IO uint32_t limit[4];                             /*  0x0010-0x001C   ADC Limit 0..3                                                               */
    __IO uint32_t afe_ctrl;                             /*  0x0020          AFE Control Register                                                         */
    __IO uint32_t ro_cal0;                              /*  0x0024          RO Trim Calibration Register 0                                               */
    __IO uint32_t ro_cal1;                              /*  0x0028          RO Trim Calibration Register 1                                               */
    __IO uint32_t ro_cal2;                              /*  0x002C          RO Trim Calibration Register 2                                               */
} mxc_adc_regs_t;


/*
   Register offsets for module ADC.
*/

#define MXC_R_ADC_OFFS_CTRL                                 ((uint32_t)0x00000000UL)
#define MXC_R_ADC_OFFS_STATUS                               ((uint32_t)0x00000004UL)
#define MXC_R_ADC_OFFS_DATA                                 ((uint32_t)0x00000008UL)
#define MXC_R_ADC_OFFS_INTR                                 ((uint32_t)0x0000000CUL)
#define MXC_R_ADC_OFFS_LIMIT0                               ((uint32_t)0x00000010UL)
#define MXC_R_ADC_OFFS_LIMIT1                               ((uint32_t)0x00000014UL)
#define MXC_R_ADC_OFFS_LIMIT2                               ((uint32_t)0x00000018UL)
#define MXC_R_ADC_OFFS_LIMIT3                               ((uint32_t)0x0000001CUL)
#define MXC_R_ADC_OFFS_AFE_CTRL                             ((uint32_t)0x00000020UL)
#define MXC_R_ADC_OFFS_RO_CAL0                              ((uint32_t)0x00000024UL)
#define MXC_R_ADC_OFFS_RO_CAL1                              ((uint32_t)0x00000028UL)
#define MXC_R_ADC_OFFS_RO_CAL2                              ((uint32_t)0x0000002CUL)


/*
   Field positions and masks for module ADC.
*/

#define MXC_F_ADC_CTRL_CPU_ADC_START_POS                    0
#define MXC_F_ADC_CTRL_CPU_ADC_START                        ((uint32_t)(0x00000001UL << MXC_F_ADC_CTRL_CPU_ADC_START_POS))
#define MXC_F_ADC_CTRL_ADC_PU_POS                           1
#define MXC_F_ADC_CTRL_ADC_PU                               ((uint32_t)(0x00000001UL << MXC_F_ADC_CTRL_ADC_PU_POS))
#define MXC_F_ADC_CTRL_BUF_PU_POS                           2
#define MXC_F_ADC_CTRL_BUF_PU                               ((uint32_t)(0x00000001UL << MXC_F_ADC_CTRL_BUF_PU_POS))
#define MXC_F_ADC_CTRL_ADC_REFBUF_PU_POS                    3
#define MXC_F_ADC_CTRL_ADC_REFBUF_PU                        ((uint32_t)(0x00000001UL << MXC_F_ADC_CTRL_ADC_REFBUF_PU_POS))
#define MXC_F_ADC_CTRL_ADC_CHGPUMP_PU_POS                   4
#define MXC_F_ADC_CTRL_ADC_CHGPUMP_PU                       ((uint32_t)(0x00000001UL << MXC_F_ADC_CTRL_ADC_CHGPUMP_PU_POS))
#define MXC_F_ADC_CTRL_BUF_CHOP_DIS_POS                     5
#define MXC_F_ADC_CTRL_BUF_CHOP_DIS                         ((uint32_t)(0x00000001UL << MXC_F_ADC_CTRL_BUF_CHOP_DIS_POS))
#define MXC_F_ADC_CTRL_BUF_PUMP_DIS_POS                     6
#define MXC_F_ADC_CTRL_BUF_PUMP_DIS                         ((uint32_t)(0x00000001UL << MXC_F_ADC_CTRL_BUF_PUMP_DIS_POS))
#define MXC_F_ADC_CTRL_BUF_BYPASS_POS                       7
#define MXC_F_ADC_CTRL_BUF_BYPASS                           ((uint32_t)(0x00000001UL << MXC_F_ADC_CTRL_BUF_BYPASS_POS))
#define MXC_F_ADC_CTRL_ADC_REFSCL_POS                       8
#define MXC_F_ADC_CTRL_ADC_REFSCL                           ((uint32_t)(0x00000001UL << MXC_F_ADC_CTRL_ADC_REFSCL_POS))
#define MXC_F_ADC_CTRL_ADC_SCALE_POS                        9
#define MXC_F_ADC_CTRL_ADC_SCALE                            ((uint32_t)(0x00000001UL << MXC_F_ADC_CTRL_ADC_SCALE_POS))
#define MXC_F_ADC_CTRL_ADC_REFSEL_POS                       10
#define MXC_F_ADC_CTRL_ADC_REFSEL                           ((uint32_t)(0x00000001UL << MXC_F_ADC_CTRL_ADC_REFSEL_POS))
#define MXC_F_ADC_CTRL_ADC_CLK_EN_POS                       11
#define MXC_F_ADC_CTRL_ADC_CLK_EN                           ((uint32_t)(0x00000001UL << MXC_F_ADC_CTRL_ADC_CLK_EN_POS))
#define MXC_F_ADC_CTRL_ADC_CHSEL_POS                        12
#define MXC_F_ADC_CTRL_ADC_CHSEL                            ((uint32_t)(0x0000000FUL << MXC_F_ADC_CTRL_ADC_CHSEL_POS))

#if (MXC_ADC_REV > 0)
#define MXC_F_ADC_CTRL_ADC_XREF_POS                         16
#define MXC_F_ADC_CTRL_ADC_XREF                             ((uint32_t)(0x00000001UL << MXC_F_ADC_CTRL_ADC_XREF_POS))
#endif

#define MXC_F_ADC_CTRL_ADC_DATAALIGN_POS                    17
#define MXC_F_ADC_CTRL_ADC_DATAALIGN                        ((uint32_t)(0x00000001UL << MXC_F_ADC_CTRL_ADC_DATAALIGN_POS))
#define MXC_F_ADC_CTRL_AFE_PWR_UP_DLY_POS                   24
#define MXC_F_ADC_CTRL_AFE_PWR_UP_DLY                       ((uint32_t)(0x000000FFUL << MXC_F_ADC_CTRL_AFE_PWR_UP_DLY_POS))

#define MXC_F_ADC_STATUS_ADC_ACTIVE_POS                     0
#define MXC_F_ADC_STATUS_ADC_ACTIVE                         ((uint32_t)(0x00000001UL << MXC_F_ADC_STATUS_ADC_ACTIVE_POS))
#define MXC_F_ADC_STATUS_RO_CAL_ATOMIC_ACTIVE_POS           1
#define MXC_F_ADC_STATUS_RO_CAL_ATOMIC_ACTIVE               ((uint32_t)(0x00000001UL << MXC_F_ADC_STATUS_RO_CAL_ATOMIC_ACTIVE_POS))
#define MXC_F_ADC_STATUS_AFE_PWR_UP_ACTIVE_POS              2
#define MXC_F_ADC_STATUS_AFE_PWR_UP_ACTIVE                  ((uint32_t)(0x00000001UL << MXC_F_ADC_STATUS_AFE_PWR_UP_ACTIVE_POS))
#define MXC_F_ADC_STATUS_ADC_OVERFLOW_POS                   3
#define MXC_F_ADC_STATUS_ADC_OVERFLOW                       ((uint32_t)(0x00000001UL << MXC_F_ADC_STATUS_ADC_OVERFLOW_POS))

#define MXC_F_ADC_DATA_ADC_DATA_POS                         0
#define MXC_F_ADC_DATA_ADC_DATA                             ((uint32_t)(0x0000FFFFUL << MXC_F_ADC_DATA_ADC_DATA_POS))

#define MXC_F_ADC_INTR_ADC_DONE_IE_POS                      0
#define MXC_F_ADC_INTR_ADC_DONE_IE                          ((uint32_t)(0x00000001UL << MXC_F_ADC_INTR_ADC_DONE_IE_POS))
#define MXC_F_ADC_INTR_ADC_REF_READY_IE_POS                 1
#define MXC_F_ADC_INTR_ADC_REF_READY_IE                     ((uint32_t)(0x00000001UL << MXC_F_ADC_INTR_ADC_REF_READY_IE_POS))
#define MXC_F_ADC_INTR_ADC_HI_LIMIT_IE_POS                  2
#define MXC_F_ADC_INTR_ADC_HI_LIMIT_IE                      ((uint32_t)(0x00000001UL << MXC_F_ADC_INTR_ADC_HI_LIMIT_IE_POS))
#define MXC_F_ADC_INTR_ADC_LO_LIMIT_IE_POS                  3
#define MXC_F_ADC_INTR_ADC_LO_LIMIT_IE                      ((uint32_t)(0x00000001UL << MXC_F_ADC_INTR_ADC_LO_LIMIT_IE_POS))
#define MXC_F_ADC_INTR_ADC_OVERFLOW_IE_POS                  4
#define MXC_F_ADC_INTR_ADC_OVERFLOW_IE                      ((uint32_t)(0x00000001UL << MXC_F_ADC_INTR_ADC_OVERFLOW_IE_POS))
#define MXC_F_ADC_INTR_RO_CAL_DONE_IE_POS                   5
#define MXC_F_ADC_INTR_RO_CAL_DONE_IE                       ((uint32_t)(0x00000001UL << MXC_F_ADC_INTR_RO_CAL_DONE_IE_POS))
#define MXC_F_ADC_INTR_ADC_DONE_IF_POS                      16
#define MXC_F_ADC_INTR_ADC_DONE_IF                          ((uint32_t)(0x00000001UL << MXC_F_ADC_INTR_ADC_DONE_IF_POS))
#define MXC_F_ADC_INTR_ADC_REF_READY_IF_POS                 17
#define MXC_F_ADC_INTR_ADC_REF_READY_IF                     ((uint32_t)(0x00000001UL << MXC_F_ADC_INTR_ADC_REF_READY_IF_POS))
#define MXC_F_ADC_INTR_ADC_HI_LIMIT_IF_POS                  18
#define MXC_F_ADC_INTR_ADC_HI_LIMIT_IF                      ((uint32_t)(0x00000001UL << MXC_F_ADC_INTR_ADC_HI_LIMIT_IF_POS))
#define MXC_F_ADC_INTR_ADC_LO_LIMIT_IF_POS                  19
#define MXC_F_ADC_INTR_ADC_LO_LIMIT_IF                      ((uint32_t)(0x00000001UL << MXC_F_ADC_INTR_ADC_LO_LIMIT_IF_POS))
#define MXC_F_ADC_INTR_ADC_OVERFLOW_IF_POS                  20
#define MXC_F_ADC_INTR_ADC_OVERFLOW_IF                      ((uint32_t)(0x00000001UL << MXC_F_ADC_INTR_ADC_OVERFLOW_IF_POS))
#define MXC_F_ADC_INTR_RO_CAL_DONE_IF_POS                   21
#define MXC_F_ADC_INTR_RO_CAL_DONE_IF                       ((uint32_t)(0x00000001UL << MXC_F_ADC_INTR_RO_CAL_DONE_IF_POS))
#define MXC_F_ADC_INTR_ADC_INT_PENDING_POS                  22
#define MXC_F_ADC_INTR_ADC_INT_PENDING                      ((uint32_t)(0x00000001UL << MXC_F_ADC_INTR_ADC_INT_PENDING_POS))

#define MXC_F_ADC_LIMIT0_CH_LO_LIMIT_POS                    0
#define MXC_F_ADC_LIMIT0_CH_LO_LIMIT                        ((uint32_t)(0x000003FFUL << MXC_F_ADC_LIMIT0_CH_LO_LIMIT_POS))
#define MXC_F_ADC_LIMIT0_CH_HI_LIMIT_POS                    12
#define MXC_F_ADC_LIMIT0_CH_HI_LIMIT                        ((uint32_t)(0x000003FFUL << MXC_F_ADC_LIMIT0_CH_HI_LIMIT_POS))
#define MXC_F_ADC_LIMIT0_CH_SEL_POS                         24
#define MXC_F_ADC_LIMIT0_CH_SEL                             ((uint32_t)(0x0000000FUL << MXC_F_ADC_LIMIT0_CH_SEL_POS))
#define MXC_F_ADC_LIMIT0_CH_LO_LIMIT_EN_POS                 28
#define MXC_F_ADC_LIMIT0_CH_LO_LIMIT_EN                     ((uint32_t)(0x00000001UL << MXC_F_ADC_LIMIT0_CH_LO_LIMIT_EN_POS))
#define MXC_F_ADC_LIMIT0_CH_HI_LIMIT_EN_POS                 29
#define MXC_F_ADC_LIMIT0_CH_HI_LIMIT_EN                     ((uint32_t)(0x00000001UL << MXC_F_ADC_LIMIT0_CH_HI_LIMIT_EN_POS))

#define MXC_F_ADC_LIMIT1_CH_LO_LIMIT_POS                    0
#define MXC_F_ADC_LIMIT1_CH_LO_LIMIT                        ((uint32_t)(0x000003FFUL << MXC_F_ADC_LIMIT1_CH_LO_LIMIT_POS))
#define MXC_F_ADC_LIMIT1_CH_HI_LIMIT_POS                    12
#define MXC_F_ADC_LIMIT1_CH_HI_LIMIT                        ((uint32_t)(0x000003FFUL << MXC_F_ADC_LIMIT1_CH_HI_LIMIT_POS))
#define MXC_F_ADC_LIMIT1_CH_SEL_POS                         24
#define MXC_F_ADC_LIMIT1_CH_SEL                             ((uint32_t)(0x0000000FUL << MXC_F_ADC_LIMIT1_CH_SEL_POS))
#define MXC_F_ADC_LIMIT1_CH_LO_LIMIT_EN_POS                 28
#define MXC_F_ADC_LIMIT1_CH_LO_LIMIT_EN                     ((uint32_t)(0x00000001UL << MXC_F_ADC_LIMIT1_CH_LO_LIMIT_EN_POS))
#define MXC_F_ADC_LIMIT1_CH_HI_LIMIT_EN_POS                 29
#define MXC_F_ADC_LIMIT1_CH_HI_LIMIT_EN                     ((uint32_t)(0x00000001UL << MXC_F_ADC_LIMIT1_CH_HI_LIMIT_EN_POS))

#define MXC_F_ADC_LIMIT2_CH_LO_LIMIT_POS                    0
#define MXC_F_ADC_LIMIT2_CH_LO_LIMIT                        ((uint32_t)(0x000003FFUL << MXC_F_ADC_LIMIT2_CH_LO_LIMIT_POS))
#define MXC_F_ADC_LIMIT2_CH_HI_LIMIT_POS                    12
#define MXC_F_ADC_LIMIT2_CH_HI_LIMIT                        ((uint32_t)(0x000003FFUL << MXC_F_ADC_LIMIT2_CH_HI_LIMIT_POS))
#define MXC_F_ADC_LIMIT2_CH_SEL_POS                         24
#define MXC_F_ADC_LIMIT2_CH_SEL                             ((uint32_t)(0x0000000FUL << MXC_F_ADC_LIMIT2_CH_SEL_POS))
#define MXC_F_ADC_LIMIT2_CH_LO_LIMIT_EN_POS                 28
#define MXC_F_ADC_LIMIT2_CH_LO_LIMIT_EN                     ((uint32_t)(0x00000001UL << MXC_F_ADC_LIMIT2_CH_LO_LIMIT_EN_POS))
#define MXC_F_ADC_LIMIT2_CH_HI_LIMIT_EN_POS                 29
#define MXC_F_ADC_LIMIT2_CH_HI_LIMIT_EN                     ((uint32_t)(0x00000001UL << MXC_F_ADC_LIMIT2_CH_HI_LIMIT_EN_POS))

#define MXC_F_ADC_LIMIT3_CH_LO_LIMIT_POS                    0
#define MXC_F_ADC_LIMIT3_CH_LO_LIMIT                        ((uint32_t)(0x000003FFUL << MXC_F_ADC_LIMIT3_CH_LO_LIMIT_POS))
#define MXC_F_ADC_LIMIT3_CH_HI_LIMIT_POS                    12
#define MXC_F_ADC_LIMIT3_CH_HI_LIMIT                        ((uint32_t)(0x000003FFUL << MXC_F_ADC_LIMIT3_CH_HI_LIMIT_POS))
#define MXC_F_ADC_LIMIT3_CH_SEL_POS                         24
#define MXC_F_ADC_LIMIT3_CH_SEL                             ((uint32_t)(0x0000000FUL << MXC_F_ADC_LIMIT3_CH_SEL_POS))
#define MXC_F_ADC_LIMIT3_CH_LO_LIMIT_EN_POS                 28
#define MXC_F_ADC_LIMIT3_CH_LO_LIMIT_EN                     ((uint32_t)(0x00000001UL << MXC_F_ADC_LIMIT3_CH_LO_LIMIT_EN_POS))
#define MXC_F_ADC_LIMIT3_CH_HI_LIMIT_EN_POS                 29
#define MXC_F_ADC_LIMIT3_CH_HI_LIMIT_EN                     ((uint32_t)(0x00000001UL << MXC_F_ADC_LIMIT3_CH_HI_LIMIT_EN_POS))

#define MXC_F_ADC_AFE_CTRL_TMON_INTBIAS_EN_POS              8
#define MXC_F_ADC_AFE_CTRL_TMON_INTBIAS_EN                  ((uint32_t)(0x00000001UL << MXC_F_ADC_AFE_CTRL_TMON_INTBIAS_EN_POS))
#define MXC_F_ADC_AFE_CTRL_TMON_EXTBIAS_EN_POS              9
#define MXC_F_ADC_AFE_CTRL_TMON_EXTBIAS_EN                  ((uint32_t)(0x00000001UL << MXC_F_ADC_AFE_CTRL_TMON_EXTBIAS_EN_POS))

#define MXC_F_ADC_RO_CAL0_RO_CAL_EN_POS                     0
#define MXC_F_ADC_RO_CAL0_RO_CAL_EN                         ((uint32_t)(0x00000001UL << MXC_F_ADC_RO_CAL0_RO_CAL_EN_POS))
#define MXC_F_ADC_RO_CAL0_RO_CAL_RUN_POS                    1
#define MXC_F_ADC_RO_CAL0_RO_CAL_RUN                        ((uint32_t)(0x00000001UL << MXC_F_ADC_RO_CAL0_RO_CAL_RUN_POS))
#define MXC_F_ADC_RO_CAL0_RO_CAL_LOAD_POS                   2
#define MXC_F_ADC_RO_CAL0_RO_CAL_LOAD                       ((uint32_t)(0x00000001UL << MXC_F_ADC_RO_CAL0_RO_CAL_LOAD_POS))
#define MXC_F_ADC_RO_CAL0_RO_CAL_ATOMIC_POS                 4
#define MXC_F_ADC_RO_CAL0_RO_CAL_ATOMIC                     ((uint32_t)(0x00000001UL << MXC_F_ADC_RO_CAL0_RO_CAL_ATOMIC_POS))
#define MXC_F_ADC_RO_CAL0_DUMMY_POS                         5
#define MXC_F_ADC_RO_CAL0_DUMMY                             ((uint32_t)(0x00000007UL << MXC_F_ADC_RO_CAL0_DUMMY_POS))
#define MXC_F_ADC_RO_CAL0_TRM_MU_POS                        8
#define MXC_F_ADC_RO_CAL0_TRM_MU                            ((uint32_t)(0x00000FFFUL << MXC_F_ADC_RO_CAL0_TRM_MU_POS))
#define MXC_F_ADC_RO_CAL0_RO_TRM_POS                        23
#define MXC_F_ADC_RO_CAL0_RO_TRM                            ((uint32_t)(0x000001FFUL << MXC_F_ADC_RO_CAL0_RO_TRM_POS))

#define MXC_F_ADC_RO_CAL1_TRM_INIT_POS                      0
#define MXC_F_ADC_RO_CAL1_TRM_INIT                          ((uint32_t)(0x000001FFUL << MXC_F_ADC_RO_CAL1_TRM_INIT_POS))
#define MXC_F_ADC_RO_CAL1_TRM_MIN_POS                       10
#define MXC_F_ADC_RO_CAL1_TRM_MIN                           ((uint32_t)(0x000001FFUL << MXC_F_ADC_RO_CAL1_TRM_MIN_POS))
#define MXC_F_ADC_RO_CAL1_TRM_MAX_POS                       20
#define MXC_F_ADC_RO_CAL1_TRM_MAX                           ((uint32_t)(0x000001FFUL << MXC_F_ADC_RO_CAL1_TRM_MAX_POS))

#define MXC_F_ADC_RO_CAL2_AUTO_CAL_DONE_CNT_POS             0
#define MXC_F_ADC_RO_CAL2_AUTO_CAL_DONE_CNT                 ((uint32_t)(0x000000FFUL << MXC_F_ADC_RO_CAL2_AUTO_CAL_DONE_CNT_POS))

#define MXC_V_ADC_CTRL_ADC_CHSEL_AIN0                       ((uint32_t)(0x00000000UL))
#define MXC_V_ADC_CTRL_ADC_CHSEL_AIN1                       ((uint32_t)(0x00000001UL))
#define MXC_V_ADC_CTRL_ADC_CHSEL_AIN2                       ((uint32_t)(0x00000002UL))
#define MXC_V_ADC_CTRL_ADC_CHSEL_AIN3                       ((uint32_t)(0x00000003UL))
#define MXC_V_ADC_CTRL_ADC_CHSEL_AIN0_DIV_5                 ((uint32_t)(0x00000004UL))
#define MXC_V_ADC_CTRL_ADC_CHSEL_AIN1_DIV_5                 ((uint32_t)(0x00000005UL))
#define MXC_V_ADC_CTRL_ADC_CHSEL_VDDB_DIV_4                 ((uint32_t)(0x00000006UL))
#define MXC_V_ADC_CTRL_ADC_CHSEL_VDD18                      ((uint32_t)(0x00000007UL))
#define MXC_V_ADC_CTRL_ADC_CHSEL_VDD12                      ((uint32_t)(0x00000008UL))
#define MXC_V_ADC_CTRL_ADC_CHSEL_VRTC_DIV_2                 ((uint32_t)(0x00000009UL))
#define MXC_V_ADC_CTRL_ADC_CHSEL_TMON                       ((uint32_t)(0x0000000AUL))

#if(MXC_ADC_REV > 0)
#define MXC_V_ADC_CTRL_ADC_CHSEL_VDDIO_DIV_4                ((uint32_t)(0x0000000BUL))
#define MXC_V_ADC_CTRL_ADC_CHSEL_VDDIOH_DIV_4               ((uint32_t)(0x0000000CUL))
#endif

#ifdef __cplusplus
}
#endif

#endif   /* _MXC_ADC_REGS_H_ */

