/**
 * @file
 * @brief   Registers, Bit Masks and Bit Positions for the MAA Peripheral Module.
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
 *
 * $Date: 2017-02-14 18:17:18 -0600 (Tue, 14 Feb 2017) $
 * $Revision: 26427 $
 *
 *************************************************************************** */

/* Define to prevent redundant inclusion */
#ifndef _MXC_MAA_REGS_H_
#define _MXC_MAA_REGS_H_

/* **** Includes **** */
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

///@cond
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
///@endcond 



/**
 * @ingroup  icc_registers
 * @defgroup    maa_registers Registers
 * @brief       Registers, Bit Masks and Bit Positions for the MAA Peripheral Module.
x * @{
 */

/**
 * Structure type to access the MAA Peripheral Module Registers.
 */
 typedef struct {
    __IO uint32_t ctrl;                                 /**< <b><tt>0x0000</tt></b> MAA_CTRL - MAA Control, Configuration and Status                          */
    __IO uint32_t maws;                                 /**< <b><tt>0x0004</tt></b> MAA_MAWS - MAA Word (Operand) Size, Big/Little Endian Mode Select         */
} mxc_maa_regs_t;
/**@} end of maa_registers group */


/**
 * @ingroup     maa
 * @defgroup    maa_mem_segments Memory Segment Registers
 * @brief       Registers, Bit Masks and Bit Positions for the MAA Memory Mapped Segments
 * @{
 */
/**
 * Structure type to access the MAA Peripheral Module Memory Mapped Registers. 
 */
typedef struct {
    __IO uint32_t seg0[32];                             /*  0x0000-0x007C   [128 bytes] MAA Memory Segment 0                                             */
    __IO uint32_t seg1[32];                             /*  0x0080-0x00FC   [128 bytes] MAA Memory Segment 1                                             */
    __IO uint32_t seg2[32];                             /*  0x0100-0x017C   [128 bytes] MAA Memory Segment 2                                             */
    __IO uint32_t seg3[32];                             /*  0x0180-0x01FC   [128 bytes] MAA Memory Segment 3                                             */
    __IO uint32_t seg4[32];                             /*  0x0200-0x027C   [128 bytes] MAA Memory Segment 4                                             */
    __IO uint32_t seg5[32];                             /*  0x0280-0x02FC   [128 bytes] MAA Memory Segment 5                                             */
} mxc_maa_mem_regs_t;
/**@} end of maa_mem_segments group */

/**
 * @ingroup maa_registers
 * @defgroup MAA_Register_Offsets Register Offsets
 * @brief MAA Register Offsets from the MAA Peripheral Module Base Address.
 * @{
 */
#define MXC_R_MAA_OFFS_CTRL                                 ((uint32_t)0x00000000UL)      /**< Offset from MAA Base Peripheral Address: <b><tt>0x0000</tt></b>  */
#define MXC_R_MAA_OFFS_MAWS                                 ((uint32_t)0x00000004UL)      /**< Offset from MAA Base Peripheral Address: <b><tt>0x0004</tt></b>  */
/**@} end of group MAA_Register_Offsets */
/**
 * @ingroup maa_mem_segments
 * @defgroup MAA_Register_Mem_Offsets Register Offsets
 * @brief MAA Memory Mapped Register Offsets from the MAA Peripheral Module Base Memory Mapped Address.
 * @{
 */
#define MXC_R_MAA_MEM_OFFS_SEG0                             ((uint32_t)0x00000000UL)      /**< Offset from MAA Base Peripheral Memory Address: <b><tt>0x0000</tt></b>  */
#define MXC_R_MAA_MEM_OFFS_SEG1                             ((uint32_t)0x00000080UL)      /**< Offset from MAA Base Peripheral Memory Address: <b><tt>0x0080</tt></b>  */
#define MXC_R_MAA_MEM_OFFS_SEG2                             ((uint32_t)0x00000100UL)      /**< Offset from MAA Base Peripheral Memory Address: <b><tt>0x0100</tt></b>  */
#define MXC_R_MAA_MEM_OFFS_SEG3                             ((uint32_t)0x00000180UL)      /**< Offset from MAA Base Peripheral Memory Address: <b><tt>0x0180</tt></b>  */
#define MXC_R_MAA_MEM_OFFS_SEG4                             ((uint32_t)0x00000200UL)      /**< Offset from MAA Base Peripheral Memory Address: <b><tt>0x0200</tt></b>  */
#define MXC_R_MAA_MEM_OFFS_SEG5                             ((uint32_t)0x00000280UL)      /**< Offset from MAA Base Peripheral Memory Address: <b><tt>0x0280</tt></b>  */
/**@} end of group MAA_Register_Mem_Offsets */

/*
   Field positions and masks for module MAA.
*/
/**
 * @ingroup    maa_registers
 * @defgroup   maa_ctrl MAA_CTRL
 * @brief      Field Positions and Masks 
 */
#define MXC_F_MAA_CTRL_START_POS                            0                                                             /**< START Position */
#define MXC_F_MAA_CTRL_START                                ((uint32_t)(0x00000001UL << MXC_F_MAA_CTRL_START_POS))        /**< START Mask */
#define MXC_F_MAA_CTRL_OPSEL_POS                            1                                                             /**< OPSEL Position */
#define MXC_F_MAA_CTRL_OPSEL                                ((uint32_t)(0x00000007UL << MXC_F_MAA_CTRL_OPSEL_POS))        /**< OPSEL Mask */
#define MXC_F_MAA_CTRL_OCALC_POS                            4                                                             /**< OCALC Position */
#define MXC_F_MAA_CTRL_OCALC                                ((uint32_t)(0x00000001UL << MXC_F_MAA_CTRL_OCALC_POS))        /**< OCALC Mask */
#define MXC_F_MAA_CTRL_IF_DONE_POS                          5                                                             /**< IF_DONE Position */
#define MXC_F_MAA_CTRL_IF_DONE                              ((uint32_t)(0x00000001UL << MXC_F_MAA_CTRL_IF_DONE_POS))      /**< IF_DONE Mask */
#define MXC_F_MAA_CTRL_INTEN_POS                            6                                                             /**< INTEN Position */
#define MXC_F_MAA_CTRL_INTEN                                ((uint32_t)(0x00000001UL << MXC_F_MAA_CTRL_INTEN_POS))        /**< INTEN Mask */
#define MXC_F_MAA_CTRL_IF_ERROR_POS                         7                                                             /**< IF_ERROR Position */
#define MXC_F_MAA_CTRL_IF_ERROR                             ((uint32_t)(0x00000001UL << MXC_F_MAA_CTRL_IF_ERROR_POS))     /**< IF_ERROR Mask */
#define MXC_F_MAA_CTRL_OFS_A_POS                            8                                                             /**< OFS_A Position */
#define MXC_F_MAA_CTRL_OFS_A                                ((uint32_t)(0x00000003UL << MXC_F_MAA_CTRL_OFS_A_POS))        /**< OFS_A Mask */
#define MXC_F_MAA_CTRL_OFS_B_POS                            10                                                            /**< OFS_B Position */
#define MXC_F_MAA_CTRL_OFS_B                                ((uint32_t)(0x00000003UL << MXC_F_MAA_CTRL_OFS_B_POS))        /**< OFS_B Mask */
#define MXC_F_MAA_CTRL_OFS_EXP_POS                          12                                                            /**< OFS_EXP Position */
#define MXC_F_MAA_CTRL_OFS_EXP                              ((uint32_t)(0x00000003UL << MXC_F_MAA_CTRL_OFS_EXP_POS))      /**< OFS_EXP Mask */
#define MXC_F_MAA_CTRL_OFS_MOD_POS                          14                                                            /**< OFS_MOD Position */
#define MXC_F_MAA_CTRL_OFS_MOD                              ((uint32_t)(0x00000003UL << MXC_F_MAA_CTRL_OFS_MOD_POS))      /**< OFS_MOD Mask */
#define MXC_F_MAA_CTRL_SEG_A_POS                            16                                                            /**< SEG_A Position */
#define MXC_F_MAA_CTRL_SEG_A                                ((uint32_t)(0x0000000FUL << MXC_F_MAA_CTRL_SEG_A_POS))        /**< SEG_A Mask */
#define MXC_F_MAA_CTRL_SEG_B_POS                            20                                                            /**< SEG_B Position */
#define MXC_F_MAA_CTRL_SEG_B                                ((uint32_t)(0x0000000FUL << MXC_F_MAA_CTRL_SEG_B_POS))        /**< SEG_B Mask */
#define MXC_F_MAA_CTRL_SEG_RES_POS                          24                                                            /**< SEG_RES Position */
#define MXC_F_MAA_CTRL_SEG_RES                              ((uint32_t)(0x0000000FUL << MXC_F_MAA_CTRL_SEG_RES_POS))      /**< SEG_RES Mask */
#define MXC_F_MAA_CTRL_SEG_TMP_POS                          28                                                            /**< SEG_TMP Position */
#define MXC_F_MAA_CTRL_SEG_TMP                              ((uint32_t)(0x0000000FUL << MXC_F_MAA_CTRL_SEG_TMP_POS))      /**< SEG_TMP Mask */
/**@} end of maa_ctrl group */

/**
 * @ingroup    maa_registers
 * @defgroup   maa_maws MAA_MAWS
 * @brief      Field Positions and Masks 
 */
#define MXC_F_MAA_MAWS_MODLEN_POS                           0                                                             /**< MODLEN Position */
#define MXC_F_MAA_MAWS_MODLEN                               ((uint32_t)(0x000007FFUL << MXC_F_MAA_MAWS_MODLEN_POS))       /**< MODLEN Mask */
#define MXC_F_MAA_MAWS_BYTESWAP_POS                         15                                                            /**< BYTESWAP Position */
#define MXC_F_MAA_MAWS_BYTESWAP                             ((uint32_t)(0x00000001UL << MXC_F_MAA_MAWS_BYTESWAP_POS))     /**< BYTESWAP Mask */
/**@} end of group MAA_MAWS */


/*
   Field values and shifted values for module MAA.
*/
/**
 * @ingroup    maa_ctrl
 * @defgroup   maa_oppsel MAA_OPSEL
 * @brief      MAA Operation Select - Field Values and Shifted Field Values.
 */
#define MXC_V_MAA_OPSEL_EXP                                                     ((uint32_t)(0x00000000UL))              /**< Field Value: OPSEL_EXP    */
#define MXC_V_MAA_OPSEL_SQR                                                     ((uint32_t)(0x00000001UL))              /**< Field Value: OPSEL_SQR    */
#define MXC_V_MAA_OPSEL_MUL                                                     ((uint32_t)(0x00000002UL))              /**< Field Value: OPSEL_MUL    */
#define MXC_V_MAA_OPSEL_SQRMUL                                                  ((uint32_t)(0x00000003UL))              /**< Field Value: OPSEL_SQRMUL */
#define MXC_V_MAA_OPSEL_ADD                                                     ((uint32_t)(0x00000004UL))              /**< Field Value: OPSEL_ADD    */
#define MXC_V_MAA_OPSEL_SUB                                                     ((uint32_t)(0x00000005UL))              /**< Field Value: OPSEL_SUB    */

#define MXC_S_MAA_OPSEL_EXP                                                     ((uint32_t)(MXC_V_MAA_OPSEL_EXP     << MXC_F_MAA_CTRL_OPSEL_POS))     /**< Shifted Field Value: OPSEL_EXP    */
#define MXC_S_MAA_OPSEL_SQR                                                     ((uint32_t)(MXC_V_MAA_OPSEL_SQR     << MXC_F_MAA_CTRL_OPSEL_POS))     /**< Shifted Field Value: OPSEL_SQR    */
#define MXC_S_MAA_OPSEL_MUL                                                     ((uint32_t)(MXC_V_MAA_OPSEL_MUL     << MXC_F_MAA_CTRL_OPSEL_POS))     /**< Shifted Field Value: OPSEL_MUL    */
#define MXC_S_MAA_OPSEL_SQRMUL                                                  ((uint32_t)(MXC_V_MAA_OPSEL_SQRMUL  << MXC_F_MAA_CTRL_OPSEL_POS))     /**< Shifted Field Value: OPSEL_SQRMUL */
#define MXC_S_MAA_OPSEL_ADD                                                     ((uint32_t)(MXC_V_MAA_OPSEL_ADD     << MXC_F_MAA_CTRL_OPSEL_POS))     /**< Shifted Field Value: OPSEL_ADD    */
#define MXC_S_MAA_OPSEL_SUB                                                     ((uint32_t)(MXC_V_MAA_OPSEL_SUB     << MXC_F_MAA_CTRL_OPSEL_POS))     /**< Shifted Field Value: OPSEL_SUB    */
/**@} end of group maa_opsel_values */


#ifdef __cplusplus
}
#endif

#endif   /* _MXC_MAA_REGS_H_ */

