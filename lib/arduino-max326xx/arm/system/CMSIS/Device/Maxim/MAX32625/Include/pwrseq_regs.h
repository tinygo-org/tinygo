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
 * $Date: 2016-03-11 12:50:27 -0600 (Fri, 11 Mar 2016) $
 * $Revision: 21840 $
 *
 ******************************************************************************/

#ifndef _MXC_PWRSEQ_REGS_H_
#define _MXC_PWRSEQ_REGS_H_

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
    __IO uint32_t reg0;                                 /*  0x0000          Power Sequencer Control Register 0                                           */
    __IO uint32_t reg1;                                 /*  0x0004          Power Sequencer Control Register 1                                           */
    __IO uint32_t reg2;                                 /*  0x0008          Power Sequencer Control Register 2                                           */
    __IO uint32_t reg3;                                 /*  0x000C          Power Sequencer Control Register 3                                           */
    __IO uint32_t reg4;                                 /*  0x0010          Power Sequencer Control Register 4 (Internal Test Only)                      */
    __IO uint32_t reg5;                                 /*  0x0014          Power Sequencer Control Register 5 (Trim 0)                                  */
    __IO uint32_t reg6;                                 /*  0x0018          Power Sequencer Control Register 6 (Trim 1)                                  */
    __IO uint32_t reg7;                                 /*  0x001C          Power Sequencer Control Register 7 (Trim 2)                                  */
    __IO uint32_t flags;                                /*  0x0020          Power Sequencer Flags                                                        */
    __IO uint32_t msk_flags;                            /*  0x0024          Power Sequencer Flags Mask Register                                          */
    __R  uint32_t rsv028;                               /*  0x0028                                                                                       */
    __IO uint32_t wr_protect;                           /*  0x002C          Critical Setting Write Protect Register                                      */
    __IO uint32_t retn_ctrl0;                           /*  0x0030          Retention Control Register 0                                                 */
    __IO uint32_t retn_ctrl1;                           /*  0x0034          Retention Control Register 1                                                 */
    __IO uint32_t pwr_misc;                             /*  0x0038          Power Misc Controls                                                          */
    __IO uint32_t rtc_ctrl2;                            /*  0x003C          RTC Misc Controls                                                            */
} mxc_pwrseq_regs_t;


/*
   Register offsets for module PWRSEQ.
*/

#define MXC_R_PWRSEQ_OFFS_REG0                              ((uint32_t)0x00000000UL)
#define MXC_R_PWRSEQ_OFFS_REG1                              ((uint32_t)0x00000004UL)
#define MXC_R_PWRSEQ_OFFS_REG2                              ((uint32_t)0x00000008UL)
#define MXC_R_PWRSEQ_OFFS_REG3                              ((uint32_t)0x0000000CUL)
#define MXC_R_PWRSEQ_OFFS_REG4                              ((uint32_t)0x00000010UL)
#define MXC_R_PWRSEQ_OFFS_REG5                              ((uint32_t)0x00000014UL)
#define MXC_R_PWRSEQ_OFFS_REG6                              ((uint32_t)0x00000018UL)
#define MXC_R_PWRSEQ_OFFS_REG7                              ((uint32_t)0x0000001CUL)
#define MXC_R_PWRSEQ_OFFS_FLAGS                             ((uint32_t)0x00000020UL)
#define MXC_R_PWRSEQ_OFFS_MSK_FLAGS                         ((uint32_t)0x00000024UL)
#define MXC_R_PWRSEQ_OFFS_WR_PROTECT                        ((uint32_t)0x0000002CUL)
#define MXC_R_PWRSEQ_OFFS_RETN_CTRL0                        ((uint32_t)0x00000030UL)
#define MXC_R_PWRSEQ_OFFS_RETN_CTRL1                        ((uint32_t)0x00000034UL)
#define MXC_R_PWRSEQ_OFFS_PWR_MISC                          ((uint32_t)0x00000038UL)
#define MXC_R_PWRSEQ_OFFS_RTC_CTRL2                         ((uint32_t)0x0000003CUL)


/*
   Field positions and masks for module PWRSEQ.
*/

#define MXC_F_PWRSEQ_REG0_PWR_LP1_POS                       0
#define MXC_F_PWRSEQ_REG0_PWR_LP1                           ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG0_PWR_LP1_POS))
#define MXC_F_PWRSEQ_REG0_PWR_FIRST_BOOT_POS                1
#define MXC_F_PWRSEQ_REG0_PWR_FIRST_BOOT                    ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG0_PWR_FIRST_BOOT_POS))
#define MXC_F_PWRSEQ_REG0_PWR_SYS_REBOOT_POS                2
#define MXC_F_PWRSEQ_REG0_PWR_SYS_REBOOT                    ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG0_PWR_SYS_REBOOT_POS))
#define MXC_F_PWRSEQ_REG0_PWR_FLASHEN_RUN_POS               3
#define MXC_F_PWRSEQ_REG0_PWR_FLASHEN_RUN                   ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG0_PWR_FLASHEN_RUN_POS))
#define MXC_F_PWRSEQ_REG0_PWR_FLASHEN_SLP_POS               4
#define MXC_F_PWRSEQ_REG0_PWR_FLASHEN_SLP                   ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG0_PWR_FLASHEN_SLP_POS))
#define MXC_F_PWRSEQ_REG0_PWR_RETREGEN_RUN_POS              5
#define MXC_F_PWRSEQ_REG0_PWR_RETREGEN_RUN                  ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG0_PWR_RETREGEN_RUN_POS))
#define MXC_F_PWRSEQ_REG0_PWR_RETREGEN_SLP_POS              6
#define MXC_F_PWRSEQ_REG0_PWR_RETREGEN_SLP                  ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG0_PWR_RETREGEN_SLP_POS))
#define MXC_F_PWRSEQ_REG0_PWR_ROEN_RUN_POS                  7
#define MXC_F_PWRSEQ_REG0_PWR_ROEN_RUN                      ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG0_PWR_ROEN_RUN_POS))
#define MXC_F_PWRSEQ_REG0_PWR_ROEN_SLP_POS                  8
#define MXC_F_PWRSEQ_REG0_PWR_ROEN_SLP                      ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG0_PWR_ROEN_SLP_POS))
#define MXC_F_PWRSEQ_REG0_PWR_NREN_RUN_POS                  9
#define MXC_F_PWRSEQ_REG0_PWR_NREN_RUN                      ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG0_PWR_NREN_RUN_POS))
#define MXC_F_PWRSEQ_REG0_PWR_NREN_SLP_POS                  10
#define MXC_F_PWRSEQ_REG0_PWR_NREN_SLP                      ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG0_PWR_NREN_SLP_POS))
#define MXC_F_PWRSEQ_REG0_PWR_RTCEN_RUN_POS                 11
#define MXC_F_PWRSEQ_REG0_PWR_RTCEN_RUN                     ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG0_PWR_RTCEN_RUN_POS))
#define MXC_F_PWRSEQ_REG0_PWR_RTCEN_SLP_POS                 12
#define MXC_F_PWRSEQ_REG0_PWR_RTCEN_SLP                     ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG0_PWR_RTCEN_SLP_POS))
#define MXC_F_PWRSEQ_REG0_PWR_SVM12EN_RUN_POS               13
#define MXC_F_PWRSEQ_REG0_PWR_SVM12EN_RUN                   ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG0_PWR_SVM12EN_RUN_POS))
#define MXC_F_PWRSEQ_REG0_PWR_SVM18EN_RUN_POS               15
#define MXC_F_PWRSEQ_REG0_PWR_SVM18EN_RUN                   ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG0_PWR_SVM18EN_RUN_POS))
#define MXC_F_PWRSEQ_REG0_PWR_SVMRTCEN_RUN_POS              17
#define MXC_F_PWRSEQ_REG0_PWR_SVMRTCEN_RUN                  ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG0_PWR_SVMRTCEN_RUN_POS))
#define MXC_F_PWRSEQ_REG0_PWR_SVM_VDDB_RUN_POS              19
#define MXC_F_PWRSEQ_REG0_PWR_SVM_VDDB_RUN                  ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG0_PWR_SVM_VDDB_RUN_POS))
#define MXC_F_PWRSEQ_REG0_PWR_SVMTVDD12EN_RUN_POS           21
#define MXC_F_PWRSEQ_REG0_PWR_SVMTVDD12EN_RUN               ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG0_PWR_SVMTVDD12EN_RUN_POS))
#define MXC_F_PWRSEQ_REG0_PWR_VDD12_SWEN_RUN_POS            23
#define MXC_F_PWRSEQ_REG0_PWR_VDD12_SWEN_RUN                ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG0_PWR_VDD12_SWEN_RUN_POS))
#define MXC_F_PWRSEQ_REG0_PWR_VDD12_SWEN_SLP_POS            24
#define MXC_F_PWRSEQ_REG0_PWR_VDD12_SWEN_SLP                ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG0_PWR_VDD12_SWEN_SLP_POS))
#define MXC_F_PWRSEQ_REG0_PWR_VDD18_SWEN_RUN_POS            25
#define MXC_F_PWRSEQ_REG0_PWR_VDD18_SWEN_RUN                ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG0_PWR_VDD18_SWEN_RUN_POS))
#define MXC_F_PWRSEQ_REG0_PWR_VDD18_SWEN_SLP_POS            26
#define MXC_F_PWRSEQ_REG0_PWR_VDD18_SWEN_SLP                ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG0_PWR_VDD18_SWEN_SLP_POS))
#define MXC_F_PWRSEQ_REG0_PWR_TVDD12_SWEN_RUN_POS           27
#define MXC_F_PWRSEQ_REG0_PWR_TVDD12_SWEN_RUN               ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG0_PWR_TVDD12_SWEN_RUN_POS))
#define MXC_F_PWRSEQ_REG0_PWR_TVDD12_SWEN_SLP_POS           28
#define MXC_F_PWRSEQ_REG0_PWR_TVDD12_SWEN_SLP               ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG0_PWR_TVDD12_SWEN_SLP_POS))
#define MXC_F_PWRSEQ_REG0_PWR_RCEN_RUN_POS                  29
#define MXC_F_PWRSEQ_REG0_PWR_RCEN_RUN                      ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG0_PWR_RCEN_RUN_POS))
#define MXC_F_PWRSEQ_REG0_PWR_RCEN_SLP_POS                  30
#define MXC_F_PWRSEQ_REG0_PWR_RCEN_SLP                      ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG0_PWR_RCEN_SLP_POS))
#define MXC_F_PWRSEQ_REG0_PWR_OSC_SELECT_POS                31
#define MXC_F_PWRSEQ_REG0_PWR_OSC_SELECT                    ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG0_PWR_OSC_SELECT_POS))

#define MXC_F_PWRSEQ_REG1_PWR_CLR_IO_EVENT_LATCH_POS        0
#define MXC_F_PWRSEQ_REG1_PWR_CLR_IO_EVENT_LATCH            ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG1_PWR_CLR_IO_EVENT_LATCH_POS))
#define MXC_F_PWRSEQ_REG1_PWR_CLR_IO_CFG_LATCH_POS          1
#define MXC_F_PWRSEQ_REG1_PWR_CLR_IO_CFG_LATCH              ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG1_PWR_CLR_IO_CFG_LATCH_POS))
#define MXC_F_PWRSEQ_REG1_PWR_MBUS_GATE_POS                 2
#define MXC_F_PWRSEQ_REG1_PWR_MBUS_GATE                     ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG1_PWR_MBUS_GATE_POS))
#define MXC_F_PWRSEQ_REG1_PWR_DISCHARGE_EN_POS              3
#define MXC_F_PWRSEQ_REG1_PWR_DISCHARGE_EN                  ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG1_PWR_DISCHARGE_EN_POS))
#define MXC_F_PWRSEQ_REG1_PWR_TVDD12_WELL_POS               4
#define MXC_F_PWRSEQ_REG1_PWR_TVDD12_WELL                   ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG1_PWR_TVDD12_WELL_POS))
#define MXC_F_PWRSEQ_REG1_PWR_SRAM_NWELL_SW_POS             5
#define MXC_F_PWRSEQ_REG1_PWR_SRAM_NWELL_SW                 ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG1_PWR_SRAM_NWELL_SW_POS))
#define MXC_F_PWRSEQ_REG1_PWR_AUTO_MBUS_GATE_POS            6
#define MXC_F_PWRSEQ_REG1_PWR_AUTO_MBUS_GATE                ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG1_PWR_AUTO_MBUS_GATE_POS))
#define MXC_F_PWRSEQ_REG1_PWR_SVM_VDDIO_EN_RUN_POS          8
#define MXC_F_PWRSEQ_REG1_PWR_SVM_VDDIO_EN_RUN              ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG1_PWR_SVM_VDDIO_EN_RUN_POS))
#define MXC_F_PWRSEQ_REG1_PWR_SVM_VDDIOH_EN_RUN_POS         10
#define MXC_F_PWRSEQ_REG1_PWR_SVM_VDDIOH_EN_RUN             ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG1_PWR_SVM_VDDIOH_EN_RUN_POS))
#define MXC_F_PWRSEQ_REG1_PWR_RETREG_SRC_V12_POS            12
#define MXC_F_PWRSEQ_REG1_PWR_RETREG_SRC_V12                ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG1_PWR_RETREG_SRC_V12_POS))
#define MXC_F_PWRSEQ_REG1_PWR_RETREG_SRC_VRTC_POS           13
#define MXC_F_PWRSEQ_REG1_PWR_RETREG_SRC_VRTC               ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG1_PWR_RETREG_SRC_VRTC_POS))
#define MXC_F_PWRSEQ_REG1_PWR_RETREG_SRC_V18_POS            14
#define MXC_F_PWRSEQ_REG1_PWR_RETREG_SRC_V18                ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG1_PWR_RETREG_SRC_V18_POS))
#define MXC_F_PWRSEQ_REG1_PWR_VDDIO_EN_ISO_POR_POS          16
#define MXC_F_PWRSEQ_REG1_PWR_VDDIO_EN_ISO_POR              ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG1_PWR_VDDIO_EN_ISO_POR_POS))
#define MXC_F_PWRSEQ_REG1_PWR_VDDIOH_EN_ISO_POR_POS         17
#define MXC_F_PWRSEQ_REG1_PWR_VDDIOH_EN_ISO_POR             ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG1_PWR_VDDIOH_EN_ISO_POR_POS))
#define MXC_F_PWRSEQ_REG1_PWR_LP0_CORE_RESUME_EN_POS        18
#define MXC_F_PWRSEQ_REG1_PWR_LP0_CORE_RESUME_EN            ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG1_PWR_LP0_CORE_RESUME_EN_POS))
#define MXC_F_PWRSEQ_REG1_PWR_LP1_CORE_RSTN_EN_POS          19
#define MXC_F_PWRSEQ_REG1_PWR_LP1_CORE_RSTN_EN              ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG1_PWR_LP1_CORE_RSTN_EN_POS))

#define MXC_F_PWRSEQ_REG2_PWR_VDD12_HYST_POS                0
#define MXC_F_PWRSEQ_REG2_PWR_VDD12_HYST                    ((uint32_t)(0x00000003UL << MXC_F_PWRSEQ_REG2_PWR_VDD12_HYST_POS))
#define MXC_F_PWRSEQ_REG2_PWR_VDD18_HYST_POS                2
#define MXC_F_PWRSEQ_REG2_PWR_VDD18_HYST                    ((uint32_t)(0x00000003UL << MXC_F_PWRSEQ_REG2_PWR_VDD18_HYST_POS))
#define MXC_F_PWRSEQ_REG2_PWR_VRTC_HYST_POS                 4
#define MXC_F_PWRSEQ_REG2_PWR_VRTC_HYST                     ((uint32_t)(0x00000003UL << MXC_F_PWRSEQ_REG2_PWR_VRTC_HYST_POS))
#define MXC_F_PWRSEQ_REG2_PWR_VDDB_HYST_POS                 6
#define MXC_F_PWRSEQ_REG2_PWR_VDDB_HYST                     ((uint32_t)(0x00000003UL << MXC_F_PWRSEQ_REG2_PWR_VDDB_HYST_POS))
#define MXC_F_PWRSEQ_REG2_PWR_TVDD12_HYST_POS               8
#define MXC_F_PWRSEQ_REG2_PWR_TVDD12_HYST                   ((uint32_t)(0x00000003UL << MXC_F_PWRSEQ_REG2_PWR_TVDD12_HYST_POS))
#define MXC_F_PWRSEQ_REG2_PWR_VDDIO_HYST_POS                10
#define MXC_F_PWRSEQ_REG2_PWR_VDDIO_HYST                    ((uint32_t)(0x00000003UL << MXC_F_PWRSEQ_REG2_PWR_VDDIO_HYST_POS))
#define MXC_F_PWRSEQ_REG2_PWR_VDDIOH_HYST_POS               12
#define MXC_F_PWRSEQ_REG2_PWR_VDDIOH_HYST                   ((uint32_t)(0x00000003UL << MXC_F_PWRSEQ_REG2_PWR_VDDIOH_HYST_POS))

#define MXC_F_PWRSEQ_REG3_PWR_ROSEL_POS                     0
#define MXC_F_PWRSEQ_REG3_PWR_ROSEL                         ((uint32_t)(0x00000007UL << MXC_F_PWRSEQ_REG3_PWR_ROSEL_POS))
#define MXC_F_PWRSEQ_REG3_PWR_FLTRROSEL_POS                 3
#define MXC_F_PWRSEQ_REG3_PWR_FLTRROSEL                     ((uint32_t)(0x00000007UL << MXC_F_PWRSEQ_REG3_PWR_FLTRROSEL_POS))
#define MXC_F_PWRSEQ_REG3_PWR_SVM_CLK_MUX_POS               6
#define MXC_F_PWRSEQ_REG3_PWR_SVM_CLK_MUX                   ((uint32_t)(0x00000003UL << MXC_F_PWRSEQ_REG3_PWR_SVM_CLK_MUX_POS))
#define MXC_F_PWRSEQ_REG3_PWR_RO_CLK_MUX_POS                8
#define MXC_F_PWRSEQ_REG3_PWR_RO_CLK_MUX                    ((uint32_t)(0x00000003UL << MXC_F_PWRSEQ_REG3_PWR_RO_CLK_MUX_POS))
#define MXC_F_PWRSEQ_REG3_PWR_FAILSEL_POS                   10
#define MXC_F_PWRSEQ_REG3_PWR_FAILSEL                       ((uint32_t)(0x00000007UL << MXC_F_PWRSEQ_REG3_PWR_FAILSEL_POS))
#define MXC_F_PWRSEQ_REG3_PWR_RO_DIV_POS                    16
#define MXC_F_PWRSEQ_REG3_PWR_RO_DIV                        ((uint32_t)(0x00000007UL << MXC_F_PWRSEQ_REG3_PWR_RO_DIV_POS))
#define MXC_F_PWRSEQ_REG3_PWR_RC_DIV_POS                    20
#define MXC_F_PWRSEQ_REG3_PWR_RC_DIV                        ((uint32_t)(0x00000003UL << MXC_F_PWRSEQ_REG3_PWR_RC_DIV_POS))

#define MXC_F_PWRSEQ_REG4_PWR_TM_PS_2_GPIO_POS              0
#define MXC_F_PWRSEQ_REG4_PWR_TM_PS_2_GPIO                  ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG4_PWR_TM_PS_2_GPIO_POS))
#define MXC_F_PWRSEQ_REG4_PWR_TM_FAST_TIMERS_POS            1
#define MXC_F_PWRSEQ_REG4_PWR_TM_FAST_TIMERS                ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG4_PWR_TM_FAST_TIMERS_POS))
#define MXC_F_PWRSEQ_REG4_PWR_USB_DIS_COMP_POS              3
#define MXC_F_PWRSEQ_REG4_PWR_USB_DIS_COMP                  ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG4_PWR_USB_DIS_COMP_POS))
#define MXC_F_PWRSEQ_REG4_PWR_RO_TSTCLK_EN_POS              4
#define MXC_F_PWRSEQ_REG4_PWR_RO_TSTCLK_EN                  ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG4_PWR_RO_TSTCLK_EN_POS))
#define MXC_F_PWRSEQ_REG4_PWR_NR_CLK_GATE_EN_POS            5
#define MXC_F_PWRSEQ_REG4_PWR_NR_CLK_GATE_EN                ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG4_PWR_NR_CLK_GATE_EN_POS))
#define MXC_F_PWRSEQ_REG4_PWR_EXT_CLK_IN_EN_POS             6
#define MXC_F_PWRSEQ_REG4_PWR_EXT_CLK_IN_EN                 ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG4_PWR_EXT_CLK_IN_EN_POS))
#define MXC_F_PWRSEQ_REG4_PWR_PSEQ_32K_EN_POS               7
#define MXC_F_PWRSEQ_REG4_PWR_PSEQ_32K_EN                   ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG4_PWR_PSEQ_32K_EN_POS))
#define MXC_F_PWRSEQ_REG4_PWR_RTC_MUX_POS                   8
#define MXC_F_PWRSEQ_REG4_PWR_RTC_MUX                       ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG4_PWR_RTC_MUX_POS))
#define MXC_F_PWRSEQ_REG4_PWR_RETREG_TRIM_LP1_EN_POS        9
#define MXC_F_PWRSEQ_REG4_PWR_RETREG_TRIM_LP1_EN            ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG4_PWR_RETREG_TRIM_LP1_EN_POS))
#define MXC_F_PWRSEQ_REG4_PWR_USB_XVR_TST_EN_POS            10
#define MXC_F_PWRSEQ_REG4_PWR_USB_XVR_TST_EN                ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG4_PWR_USB_XVR_TST_EN_POS))

#define MXC_F_PWRSEQ_REG5_PWR_TRIM_SVM_BG_POS               0
#define MXC_F_PWRSEQ_REG5_PWR_TRIM_SVM_BG                   ((uint32_t)(0x000001FFUL << MXC_F_PWRSEQ_REG5_PWR_TRIM_SVM_BG_POS))
#define MXC_F_PWRSEQ_REG5_PWR_TRIM_BIAS_POS                 9
#define MXC_F_PWRSEQ_REG5_PWR_TRIM_BIAS                     ((uint32_t)(0x0000003FUL << MXC_F_PWRSEQ_REG5_PWR_TRIM_BIAS_POS))
#define MXC_F_PWRSEQ_REG5_PWR_TRIM_RETREG_5_0_POS           15
#define MXC_F_PWRSEQ_REG5_PWR_TRIM_RETREG_5_0               ((uint32_t)(0x0000003FUL << MXC_F_PWRSEQ_REG5_PWR_TRIM_RETREG_5_0_POS))
#define MXC_F_PWRSEQ_REG5_PWR_RTC_TRIM_POS                  21
#define MXC_F_PWRSEQ_REG5_PWR_RTC_TRIM                      ((uint32_t)(0x0000000FUL << MXC_F_PWRSEQ_REG5_PWR_RTC_TRIM_POS))
#define MXC_F_PWRSEQ_REG5_PWR_TRIM_RETREG_7_6_POS           25
#define MXC_F_PWRSEQ_REG5_PWR_TRIM_RETREG_7_6               ((uint32_t)(0x00000003UL << MXC_F_PWRSEQ_REG5_PWR_TRIM_RETREG_7_6_POS))

#define MXC_F_PWRSEQ_REG6_PWR_TRIM_USB_BIAS_POS             0
#define MXC_F_PWRSEQ_REG6_PWR_TRIM_USB_BIAS                 ((uint32_t)(0x00000007UL << MXC_F_PWRSEQ_REG6_PWR_TRIM_USB_BIAS_POS))
#define MXC_F_PWRSEQ_REG6_PWR_TRIM_USB_PM_RES_POS           3
#define MXC_F_PWRSEQ_REG6_PWR_TRIM_USB_PM_RES               ((uint32_t)(0x0000000FUL << MXC_F_PWRSEQ_REG6_PWR_TRIM_USB_PM_RES_POS))
#define MXC_F_PWRSEQ_REG6_PWR_TRIM_USB_DM_RES_POS           7
#define MXC_F_PWRSEQ_REG6_PWR_TRIM_USB_DM_RES               ((uint32_t)(0x0000000FUL << MXC_F_PWRSEQ_REG6_PWR_TRIM_USB_DM_RES_POS))
#define MXC_F_PWRSEQ_REG6_PWR_TRIM_OSC_VREF_POS             11
#define MXC_F_PWRSEQ_REG6_PWR_TRIM_OSC_VREF                 ((uint32_t)(0x000001FFUL << MXC_F_PWRSEQ_REG6_PWR_TRIM_OSC_VREF_POS))
#define MXC_F_PWRSEQ_REG6_PWR_TRIM_CRYPTO_OSC_POS           20
#define MXC_F_PWRSEQ_REG6_PWR_TRIM_CRYPTO_OSC               ((uint32_t)(0x000001FFUL << MXC_F_PWRSEQ_REG6_PWR_TRIM_CRYPTO_OSC_POS))

#define MXC_F_PWRSEQ_REG7_PWR_FLASH_PD_LOOKAHEAD_POS        0
#define MXC_F_PWRSEQ_REG7_PWR_FLASH_PD_LOOKAHEAD            ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_REG7_PWR_FLASH_PD_LOOKAHEAD_POS))
#define MXC_F_PWRSEQ_REG7_PWR_TRIM_RC_POS                   16
#define MXC_F_PWRSEQ_REG7_PWR_TRIM_RC                       ((uint32_t)(0x0000FFFFUL << MXC_F_PWRSEQ_REG7_PWR_TRIM_RC_POS))

#define MXC_F_PWRSEQ_FLAGS_PWR_FIRST_BOOT_POS               0
#define MXC_F_PWRSEQ_FLAGS_PWR_FIRST_BOOT                   ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_FLAGS_PWR_FIRST_BOOT_POS))
#define MXC_F_PWRSEQ_FLAGS_PWR_SYS_REBOOT_POS               1
#define MXC_F_PWRSEQ_FLAGS_PWR_SYS_REBOOT                   ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_FLAGS_PWR_SYS_REBOOT_POS))
#define MXC_F_PWRSEQ_FLAGS_PWR_POWER_FAIL_POS               2
#define MXC_F_PWRSEQ_FLAGS_PWR_POWER_FAIL                   ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_FLAGS_PWR_POWER_FAIL_POS))
#define MXC_F_PWRSEQ_FLAGS_PWR_BOOT_FAIL_POS                3
#define MXC_F_PWRSEQ_FLAGS_PWR_BOOT_FAIL                    ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_FLAGS_PWR_BOOT_FAIL_POS))
#define MXC_F_PWRSEQ_FLAGS_PWR_FLASH_DISCHARGE_POS          4
#define MXC_F_PWRSEQ_FLAGS_PWR_FLASH_DISCHARGE              ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_FLAGS_PWR_FLASH_DISCHARGE_POS))
#define MXC_F_PWRSEQ_FLAGS_PWR_IOWAKEUP_POS                 5
#define MXC_F_PWRSEQ_FLAGS_PWR_IOWAKEUP                     ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_FLAGS_PWR_IOWAKEUP_POS))
#define MXC_F_PWRSEQ_FLAGS_PWR_VDD12_RST_BAD_POS            6
#define MXC_F_PWRSEQ_FLAGS_PWR_VDD12_RST_BAD                ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_FLAGS_PWR_VDD12_RST_BAD_POS))
#define MXC_F_PWRSEQ_FLAGS_PWR_VDD18_RST_BAD_POS            7
#define MXC_F_PWRSEQ_FLAGS_PWR_VDD18_RST_BAD                ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_FLAGS_PWR_VDD18_RST_BAD_POS))
#define MXC_F_PWRSEQ_FLAGS_PWR_VRTC_RST_BAD_POS             8
#define MXC_F_PWRSEQ_FLAGS_PWR_VRTC_RST_BAD                 ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_FLAGS_PWR_VRTC_RST_BAD_POS))
#define MXC_F_PWRSEQ_FLAGS_PWR_VDDB_RST_BAD_POS             9
#define MXC_F_PWRSEQ_FLAGS_PWR_VDDB_RST_BAD                 ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_FLAGS_PWR_VDDB_RST_BAD_POS))
#define MXC_F_PWRSEQ_FLAGS_PWR_TVDD12_RST_BAD_POS           10
#define MXC_F_PWRSEQ_FLAGS_PWR_TVDD12_RST_BAD               ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_FLAGS_PWR_TVDD12_RST_BAD_POS))
#define MXC_F_PWRSEQ_FLAGS_PWR_POR18Z_FAIL_LATCH_POS        11
#define MXC_F_PWRSEQ_FLAGS_PWR_POR18Z_FAIL_LATCH            ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_FLAGS_PWR_POR18Z_FAIL_LATCH_POS))
#define MXC_F_PWRSEQ_FLAGS_RTC_CMPR0_POS                    12
#define MXC_F_PWRSEQ_FLAGS_RTC_CMPR0                        ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_FLAGS_RTC_CMPR0_POS))
#define MXC_F_PWRSEQ_FLAGS_RTC_CMPR1_POS                    13
#define MXC_F_PWRSEQ_FLAGS_RTC_CMPR1                        ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_FLAGS_RTC_CMPR1_POS))
#define MXC_F_PWRSEQ_FLAGS_RTC_PRESCALE_CMP_POS             14
#define MXC_F_PWRSEQ_FLAGS_RTC_PRESCALE_CMP                 ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_FLAGS_RTC_PRESCALE_CMP_POS))
#define MXC_F_PWRSEQ_FLAGS_RTC_ROLLOVER_POS                 15
#define MXC_F_PWRSEQ_FLAGS_RTC_ROLLOVER                     ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_FLAGS_RTC_ROLLOVER_POS))
#define MXC_F_PWRSEQ_FLAGS_PWR_USB_PLUG_WAKEUP_POS          16
#define MXC_F_PWRSEQ_FLAGS_PWR_USB_PLUG_WAKEUP              ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_FLAGS_PWR_USB_PLUG_WAKEUP_POS))
#define MXC_F_PWRSEQ_FLAGS_PWR_USB_REMOVE_WAKEUP_POS        17
#define MXC_F_PWRSEQ_FLAGS_PWR_USB_REMOVE_WAKEUP            ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_FLAGS_PWR_USB_REMOVE_WAKEUP_POS))
#define MXC_F_PWRSEQ_FLAGS_PWR_TVDD12_BAD_POS               18
#define MXC_F_PWRSEQ_FLAGS_PWR_TVDD12_BAD                   ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_FLAGS_PWR_TVDD12_BAD_POS))
#define MXC_F_PWRSEQ_FLAGS_PWR_VDDIO_RST_BAD_POS            19
#define MXC_F_PWRSEQ_FLAGS_PWR_VDDIO_RST_BAD                ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_FLAGS_PWR_VDDIO_RST_BAD_POS))
#define MXC_F_PWRSEQ_FLAGS_PWR_VDDIOH_RST_BAD_POS           20
#define MXC_F_PWRSEQ_FLAGS_PWR_VDDIOH_RST_BAD               ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_FLAGS_PWR_VDDIOH_RST_BAD_POS))
#define MXC_F_PWRSEQ_FLAGS_PWR_ISOZ_VDDIO_FAIL_POS          21
#define MXC_F_PWRSEQ_FLAGS_PWR_ISOZ_VDDIO_FAIL              ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_FLAGS_PWR_ISOZ_VDDIO_FAIL_POS))
#define MXC_F_PWRSEQ_FLAGS_PWR_ISOZ_VDDIOH_FAIL_POS         22
#define MXC_F_PWRSEQ_FLAGS_PWR_ISOZ_VDDIOH_FAIL             ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_FLAGS_PWR_ISOZ_VDDIOH_FAIL_POS))
#define MXC_F_PWRSEQ_FLAGS_PWR_NANORING_WAKEUP_FLAG_POS     23
#define MXC_F_PWRSEQ_FLAGS_PWR_NANORING_WAKEUP_FLAG         ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_FLAGS_PWR_NANORING_WAKEUP_FLAG_POS))
#define MXC_F_PWRSEQ_FLAGS_PWR_WATCHDOG_RSTN_FLAG_POS       24
#define MXC_F_PWRSEQ_FLAGS_PWR_WATCHDOG_RSTN_FLAG           ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_FLAGS_PWR_WATCHDOG_RSTN_FLAG_POS))

#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_SYS_REBOOT_POS           1
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_SYS_REBOOT               ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_MSK_FLAGS_PWR_SYS_REBOOT_POS))
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_POWER_FAIL_POS           2
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_POWER_FAIL               ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_MSK_FLAGS_PWR_POWER_FAIL_POS))
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_BOOT_FAIL_POS            3
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_BOOT_FAIL                ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_MSK_FLAGS_PWR_BOOT_FAIL_POS))
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_FLASH_DISCHARGE_POS      4
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_FLASH_DISCHARGE          ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_MSK_FLAGS_PWR_FLASH_DISCHARGE_POS))
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_IOWAKEUP_POS             5
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_IOWAKEUP                 ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_MSK_FLAGS_PWR_IOWAKEUP_POS))
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_VDD12_RST_BAD_POS        6
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_VDD12_RST_BAD            ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_MSK_FLAGS_PWR_VDD12_RST_BAD_POS))
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_VDD18_RST_BAD_POS        7
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_VDD18_RST_BAD            ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_MSK_FLAGS_PWR_VDD18_RST_BAD_POS))
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_VRTC_RST_BAD_POS         8
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_VRTC_RST_BAD             ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_MSK_FLAGS_PWR_VRTC_RST_BAD_POS))
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_VDDB_RST_BAD_POS         9
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_VDDB_RST_BAD             ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_MSK_FLAGS_PWR_VDDB_RST_BAD_POS))
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_TVDD12_RST_BAD_POS       10
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_TVDD12_RST_BAD           ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_MSK_FLAGS_PWR_TVDD12_RST_BAD_POS))
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_POR18Z_FAIL_LATCH_POS    11
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_POR18Z_FAIL_LATCH        ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_MSK_FLAGS_PWR_POR18Z_FAIL_LATCH_POS))
#define MXC_F_PWRSEQ_MSK_FLAGS_RTC_CMPR0_POS                12
#define MXC_F_PWRSEQ_MSK_FLAGS_RTC_CMPR0                    ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_MSK_FLAGS_RTC_CMPR0_POS))
#define MXC_F_PWRSEQ_MSK_FLAGS_RTC_CMPR1_POS                13
#define MXC_F_PWRSEQ_MSK_FLAGS_RTC_CMPR1                    ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_MSK_FLAGS_RTC_CMPR1_POS))
#define MXC_F_PWRSEQ_MSK_FLAGS_RTC_PRESCALE_CMP_POS         14
#define MXC_F_PWRSEQ_MSK_FLAGS_RTC_PRESCALE_CMP             ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_MSK_FLAGS_RTC_PRESCALE_CMP_POS))
#define MXC_F_PWRSEQ_MSK_FLAGS_RTC_ROLLOVER_POS             15
#define MXC_F_PWRSEQ_MSK_FLAGS_RTC_ROLLOVER                 ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_MSK_FLAGS_RTC_ROLLOVER_POS))
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_USB_PLUG_WAKEUP_POS      16
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_USB_PLUG_WAKEUP          ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_MSK_FLAGS_PWR_USB_PLUG_WAKEUP_POS))
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_USB_REMOVE_WAKEUP_POS    17
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_USB_REMOVE_WAKEUP        ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_MSK_FLAGS_PWR_USB_REMOVE_WAKEUP_POS))
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_TVDD12_BAD_POS           18
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_TVDD12_BAD               ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_MSK_FLAGS_PWR_TVDD12_BAD_POS))
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_VDDIO_RST_BAD_POS        19
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_VDDIO_RST_BAD            ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_MSK_FLAGS_PWR_VDDIO_RST_BAD_POS))
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_VDDIOH_RST_BAD_POS       20
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_VDDIOH_RST_BAD           ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_MSK_FLAGS_PWR_VDDIOH_RST_BAD_POS))
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_ISOZ_VDDIO_FAIL_POS      21
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_ISOZ_VDDIO_FAIL          ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_MSK_FLAGS_PWR_ISOZ_VDDIO_FAIL_POS))
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_ISOZ_VDDIOH_FAIL_POS     22
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_ISOZ_VDDIOH_FAIL         ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_MSK_FLAGS_PWR_ISOZ_VDDIOH_FAIL_POS))
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_NANORING_WAKEUP_FLAG_POS  23
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_NANORING_WAKEUP_FLAG     ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_MSK_FLAGS_PWR_NANORING_WAKEUP_FLAG_POS))
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_WATCHDOG_RSTN_FLAG_POS   24
#define MXC_F_PWRSEQ_MSK_FLAGS_PWR_WATCHDOG_RSTN_FLAG       ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_MSK_FLAGS_PWR_WATCHDOG_RSTN_FLAG_POS))

#define MXC_F_PWRSEQ_WR_PROTECT_BYPASS_SEQ_POS              0
#define MXC_F_PWRSEQ_WR_PROTECT_BYPASS_SEQ                  ((uint32_t)(0x000000FFUL << MXC_F_PWRSEQ_WR_PROTECT_BYPASS_SEQ_POS))
#define MXC_F_PWRSEQ_WR_PROTECT_RTC_SEQ_POS                 8
#define MXC_F_PWRSEQ_WR_PROTECT_RTC_SEQ                     ((uint32_t)(0x000000FFUL << MXC_F_PWRSEQ_WR_PROTECT_RTC_SEQ_POS))
#define MXC_F_PWRSEQ_WR_PROTECT_RTC_POS                     28
#define MXC_F_PWRSEQ_WR_PROTECT_RTC                         ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_WR_PROTECT_RTC_POS))
#define MXC_F_PWRSEQ_WR_PROTECT_INFO_POS                    29
#define MXC_F_PWRSEQ_WR_PROTECT_INFO                        ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_WR_PROTECT_INFO_POS))
#define MXC_F_PWRSEQ_WR_PROTECT_BYPASS_POS                  30
#define MXC_F_PWRSEQ_WR_PROTECT_BYPASS                      ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_WR_PROTECT_BYPASS_POS))
#define MXC_F_PWRSEQ_WR_PROTECT_WP_POS                      31
#define MXC_F_PWRSEQ_WR_PROTECT_WP                          ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_WR_PROTECT_WP_POS))

#define MXC_F_PWRSEQ_RETN_CTRL0_RETN_CTRL_EN_POS            0
#define MXC_F_PWRSEQ_RETN_CTRL0_RETN_CTRL_EN                ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_RETN_CTRL0_RETN_CTRL_EN_POS))
#define MXC_F_PWRSEQ_RETN_CTRL0_RC_REL_CCG_EARLY_POS        1
#define MXC_F_PWRSEQ_RETN_CTRL0_RC_REL_CCG_EARLY            ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_RETN_CTRL0_RC_REL_CCG_EARLY_POS))
#define MXC_F_PWRSEQ_RETN_CTRL0_RC_USE_FLC_TWK_POS          2
#define MXC_F_PWRSEQ_RETN_CTRL0_RC_USE_FLC_TWK              ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_RETN_CTRL0_RC_USE_FLC_TWK_POS))
#define MXC_F_PWRSEQ_RETN_CTRL0_RC_POLL_FLASH_POS           3
#define MXC_F_PWRSEQ_RETN_CTRL0_RC_POLL_FLASH               ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_RETN_CTRL0_RC_POLL_FLASH_POS))
#define MXC_F_PWRSEQ_RETN_CTRL0_RESTORE_OVERRIDE_POS        4
#define MXC_F_PWRSEQ_RETN_CTRL0_RESTORE_OVERRIDE            ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_RETN_CTRL0_RESTORE_OVERRIDE_POS))

#define MXC_F_PWRSEQ_RETN_CTRL1_RC_TWK_POS                  0
#define MXC_F_PWRSEQ_RETN_CTRL1_RC_TWK                      ((uint32_t)(0x0000000FUL << MXC_F_PWRSEQ_RETN_CTRL1_RC_TWK_POS))
#define MXC_F_PWRSEQ_RETN_CTRL1_SRAM_FMS_POS                4
#define MXC_F_PWRSEQ_RETN_CTRL1_SRAM_FMS                    ((uint32_t)(0x0000000FUL << MXC_F_PWRSEQ_RETN_CTRL1_SRAM_FMS_POS))

#define MXC_F_PWRSEQ_PWR_MISC_INVERT_4_MASK_BITS_POS        0
#define MXC_F_PWRSEQ_PWR_MISC_INVERT_4_MASK_BITS            ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_PWR_MISC_INVERT_4_MASK_BITS_POS))

#define MXC_F_PWRSEQ_RTC_CTRL2_TIMER_ASYNC_RD_POS           0
#define MXC_F_PWRSEQ_RTC_CTRL2_TIMER_ASYNC_RD               ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_RTC_CTRL2_TIMER_ASYNC_RD_POS))
#define MXC_F_PWRSEQ_RTC_CTRL2_TIMER_ASYNC_WR_POS           1
#define MXC_F_PWRSEQ_RTC_CTRL2_TIMER_ASYNC_WR               ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_RTC_CTRL2_TIMER_ASYNC_WR_POS))
#define MXC_F_PWRSEQ_RTC_CTRL2_TIMER_AUTO_UPDATE_POS        2
#define MXC_F_PWRSEQ_RTC_CTRL2_TIMER_AUTO_UPDATE            ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_RTC_CTRL2_TIMER_AUTO_UPDATE_POS))
#define MXC_F_PWRSEQ_RTC_CTRL2_SSB_PERFORMANCE_POS          3
#define MXC_F_PWRSEQ_RTC_CTRL2_SSB_PERFORMANCE              ((uint32_t)(0x00000001UL << MXC_F_PWRSEQ_RTC_CTRL2_SSB_PERFORMANCE_POS))
#define MXC_F_PWRSEQ_RTC_CTRL2_CFG_LOCK_POS                 24
#define MXC_F_PWRSEQ_RTC_CTRL2_CFG_LOCK                     ((uint32_t)(0x000000FFUL << MXC_F_PWRSEQ_RTC_CTRL2_CFG_LOCK_POS))



#ifdef __cplusplus
}
#endif

#endif   /* _MXC_PWRSEQ_REGS_H_ */

