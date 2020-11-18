/**
 * @file
 * @brief   Registers, Bit Masks and Bit Positions for the Instruction Cache Controller.
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
 * $Date: 2017-02-14 18:15:12 -0600 (Tue, 14 Feb 2017) $
 * $Revision: 26425 $
 *
 *************************************************************************** */

/* Define to prevent redundant inclusion */
#ifndef _MXC_ICC_REGS_H_
#define _MXC_ICC_REGS_H_

/* **** Includes **** */
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

/// @cond
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

/* **** Definitions **** */

/**
 * @ingroup     icc
 * @defgroup    icc_registers Registers
 * @brief       Registers, Bit Masks and Bit Positions for the ICC.
 * @{
 */

/**
 * Structure type to access the ICC Registers.
 */
typedef struct {
    __IO uint32_t id;                                   /**< <b><tt>0x0000: </tt></b> ICC_ID Register \warning INTERNAL USE ONLY, DO NOT MODIFY  */
    __IO uint32_t mem_cfg;                              /**< <b><tt>0x0004: </tt></b> ICC_MEM_CFG Register                                       */
    __R  uint32_t rsv008[62];                           /**< <b><tt>0x0008-0x00FC: </tt></b> RESERVED                                                   */
    __IO uint32_t ctrl_stat;                            /**< <b><tt>0x0100: </tt></b> ICC_CTRL_STAT Register                                     */
    __R  uint32_t rsv104[383];                          /**< <b><tt>0x0104-0x06FC: </tt></b> RESERVED                                                   */
    __IO uint32_t invdt_all;                            /**< <b><tt>0x0700: </tt></b> ICC_INVDT_ALL Register                                     */
} mxc_icc_regs_t;
/**@} end of group icc_registers*/





/*
   Register offsets for module ICC.
*/
/**
 * @ingroup    icc_registers
 * @defgroup   ICC_Register_Offsets Register Offsets
 * @brief      Instruction Cache Controller Register Offsets from the ICC Base Address. 
 * @{
 */
#define MXC_R_ICC_OFFS_ID                                   ((uint32_t)0x00000000UL)  /**< Offset from ICC Base Address: <b><tt>0x0000</tt></b>         */
#define MXC_R_ICC_OFFS_MEM_CFG                              ((uint32_t)0x00000004UL)  /**< Offset from ICC Base Address: <b><tt>0x0004</tt></b>         */
#define MXC_R_ICC_OFFS_CTRL_STAT                            ((uint32_t)0x00000100UL)  /**< Offset from ICC Base Address: <b><tt>0x0100</tt></b>         */
#define MXC_R_ICC_OFFS_INVDT_ALL                            ((uint32_t)0x00000700UL)  /**< Offset from ICC Base Address: <b><tt>0x0700</tt></b>         */
/**@} end of group icc_registers */

/*
   Field positions and masks for module ICC.
*/
/**
 * @ingroup  icc_registers
 * @defgroup ICC_ID_Register ICC_ID
 * @brief    Field Positions and Bit Masks for the ICC_ID register
 * @{
 */
#define MXC_F_ICC_ID_RTL_VERSION_POS                        0                                                                     /**< RTL_VERSION Position           */
#define MXC_F_ICC_ID_RTL_VERSION                            ((uint32_t)(0x0000003FUL << MXC_F_ICC_ID_RTL_VERSION_POS))            /**< RTL_VERSION Mask               */
#define MXC_F_ICC_ID_PART_NUM_POS                           6                                                                     /**< PART_NUM Position              */
#define MXC_F_ICC_ID_PART_NUM                               ((uint32_t)(0x0000000FUL << MXC_F_ICC_ID_PART_NUM_POS))               /**< PART_NUM Mask                  */
#define MXC_F_ICC_ID_CACHE_ID_POS                           10                                                                    /**< CACHE_ID Position              */
#define MXC_F_ICC_ID_CACHE_ID                               ((uint32_t)(0x0000003FUL << MXC_F_ICC_ID_CACHE_ID_POS))               /**< CACHE_ID Mask                  */
/**@} end of group ICC_ID_register */
/**
 * @ingroup  icc_registers
 * @defgroup ICC_MEM_CFG_Register ICC_MEM_CFG
 * @brief    Field Positions and Bit Masks for the ICC_MEM_CFG register
 * @{
 */
#define MXC_F_ICC_MEM_CFG_CACHE_SIZE_POS                    0                                                                     /**< CACHE_SIZE Position       */
#define MXC_F_ICC_MEM_CFG_CACHE_SIZE                        ((uint32_t)(0x0000FFFFUL << MXC_F_ICC_MEM_CFG_CACHE_SIZE_POS))        /**< CACHE_SIZE Mask           */
#define MXC_F_ICC_MEM_CFG_MAIN_MEMORY_SIZE_POS              16                                                                    /**< MAIN_MEMORY_SIZE Position */
#define MXC_F_ICC_MEM_CFG_MAIN_MEMORY_SIZE                  ((uint32_t)(0x0000FFFFUL << MXC_F_ICC_MEM_CFG_MAIN_MEMORY_SIZE_POS))  /**< MAIN_MEMORY_SIZE Mask     */
/**@} end of group ICC_MEM_CFG_register */
/**
 * @ingroup  icc_registers
 * @defgroup ICC_CTRL_STAT_Register ICC_CTRL_STAT
 * @brief    Field Positions and Bit Masks for the ICC_CTRL_STAT register
 * @{
 */
#define MXC_F_ICC_CTRL_STAT_ENABLE_POS                      0                                                                     /**< ENABLE Position         */
#define MXC_F_ICC_CTRL_STAT_ENABLE                          ((uint32_t)(0x00000001UL << MXC_F_ICC_CTRL_STAT_ENABLE_POS))          /**< ENABLE Mask             */
#define MXC_F_ICC_CTRL_STAT_READY_POS                       16                                                                    /**< READY Position          */
#define MXC_F_ICC_CTRL_STAT_READY                           ((uint32_t)(0x00000001UL << MXC_F_ICC_CTRL_STAT_READY_POS))           /**< READY Mask              */
/**@} end of group ICC_CTRL_STAT_register */


#ifdef __cplusplus
}
#endif

#endif   /* _MXC_ICC_REGS_H_ */

