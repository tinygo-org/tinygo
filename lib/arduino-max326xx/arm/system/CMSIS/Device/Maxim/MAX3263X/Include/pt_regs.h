/**
 * @file
 * @brief   Type definitions for the Pulse Train Engine.
 *
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
 * $Date: 2017-02-16 08:46:02 -0600 (Thu, 16 Feb 2017) $
 * $Revision: 26453 $
 *
 **************************************************************************** */

/* Define to prevent redundant inclusion */
#ifndef _MXC_PT_REGS_H_
#define _MXC_PT_REGS_H_

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
 * @ingroup     pulsetrain
 * @defgroup    pulsetrain_registers Registers
 * @brief       Registers, Bit Masks and Bit Positions
 * @{
 */

/**
 * Structure type for the Pulse Train Global module registers allowing direct 32-bit access to each register.
 */
 typedef struct {
    __IO uint32_t enable;                               /**< <b><tt>0x0000:</tt></b> \c PTG_ENABLE Register - Global Enable/Disable Controls for All Pulse Trains. */
    __IO uint32_t resync;                               /**< <b><tt>0x0004:</tt></b> \c PTG_RESYNC Register - Global Resync (All Pulse Trains) Control.            */
    __IO uint32_t intfl;                                /**< <b><tt>0x0008:</tt></b> \c PTG_INTFL Register - Pulse Train Interrupt Flags.                          */
    __IO uint32_t inten;                                /**< <b><tt>0x000C:</tt></b> \c PTG_INTEN Register - Pulse Train Interrupt Enable/Disable.                 */
} mxc_ptg_regs_t;

/**
 * Structure type for the Pulse Train configuration registers allowing direct 32-bit access to each register.
 */
typedef struct {
    __IO uint32_t rate_length;                          /**< <b><tt>0x0000:</tt></b>\c PT_RATE_LENGTH Register - Pulse Train Configuration.           */
    __IO uint32_t train;                                /**< <b><tt>0x0004:</tt></b>\c PT_TRAIN Register - Pulse Train Output Pattern.                */
    __IO uint32_t loop;                                 /**< <b><tt>0x0008:</tt></b>\c PT_LOOP Register - Pulse Train Loop Configuration.             */
    __IO uint32_t restart;                              /**< <b><tt>0x000C:</tt></b>\c PT_RESTART Register - Pulse Train Auto-Restart Configuration.  */
} mxc_pt_regs_t;
/**@} end of pulsetrain_registers group*/

/*
   Register offsets for module PT.
*/
/**
 * @ingroup    pulsetrain_registers
 * @defgroup   PTG_Register_Offsets Global Register Offsets
 * @brief      Pluse Train Global Control Register Offsets from the Pulse Train Global Base Peripheral Address. 
 * @{
 */
#define MXC_R_PTG_OFFS_ENABLE                               ((uint32_t)0x00000000UL)    /**< Offset from the PTG Base Peripheral Address:<b><tt>0x0000</tt></b> */
#define MXC_R_PTG_OFFS_RESYNC                               ((uint32_t)0x00000004UL)    /**< Offset from the PTG Base Peripheral Address:<b><tt>0x0004</tt></b> */
#define MXC_R_PTG_OFFS_INTFL                                ((uint32_t)0x00000008UL)    /**< Offset from the PTG Base Peripheral Address:<b><tt>0x0008</tt></b> */
#define MXC_R_PTG_OFFS_INTEN                                ((uint32_t)0x0000000CUL)    /**< Offset from the PTG Base Peripheral Address:<b><tt>0x000C</tt></b> */
/**@} end of group PTG_Register_Offsets*/
/**
 * @ingroup    pulsetrain_registers
 * @defgroup   PT_Register_Offsets Register Offsets: Configuration
 * @brief      Pluse Train Configuration Register Offsets from the Pulse Train Base Peripheral Address. 
 * @{
 */  
#define MXC_R_PT_OFFS_RATE_LENGTH                           ((uint32_t)0x00000000UL)    /**< Offset from the PT Base Peripheral Address:<b><tt>0x0000</tt></b> */
#define MXC_R_PT_OFFS_TRAIN                                 ((uint32_t)0x00000004UL)    /**< Offset from the PT Base Peripheral Address:<b><tt>0x0004</tt></b> */
#define MXC_R_PT_OFFS_LOOP                                  ((uint32_t)0x00000008UL)    /**< Offset from the PT Base Peripheral Address:<b><tt>0x0008</tt></b> */
#define MXC_R_PT_OFFS_RESTART                               ((uint32_t)0x0000000CUL)    /**< Offset from the PT Base Peripheral Address:<b><tt>0x000C</tt></b> */
/**@} end of group PT_Register_Offsets*/

/*
   Field positions and masks for module PT.
*/
/**
 * @ingroup pulsetrain_registers
 * @defgroup PT_ENABLE_Register PT_ENABLE
 * @brief Field Positions and Masks
 * @{
 */
#define MXC_F_PT_ENABLE_PT0_POS                             0                                                       /**< PT0 Position */
#define MXC_F_PT_ENABLE_PT0                                 ((uint32_t)(0x00000001UL << MXC_F_PT_ENABLE_PT0_POS))   /**< PT0 Mask     */
#define MXC_F_PT_ENABLE_PT1_POS                             1                                                       /**< PT1 Position */
#define MXC_F_PT_ENABLE_PT1                                 ((uint32_t)(0x00000001UL << MXC_F_PT_ENABLE_PT1_POS))   /**< PT1 Mask     */
#define MXC_F_PT_ENABLE_PT2_POS                             2                                                       /**< PT2 Position */
#define MXC_F_PT_ENABLE_PT2                                 ((uint32_t)(0x00000001UL << MXC_F_PT_ENABLE_PT2_POS))   /**< PT2 Mask     */
#define MXC_F_PT_ENABLE_PT3_POS                             3                                                       /**< PT3 Position */
#define MXC_F_PT_ENABLE_PT3                                 ((uint32_t)(0x00000001UL << MXC_F_PT_ENABLE_PT3_POS))   /**< PT3 Mask     */
#define MXC_F_PT_ENABLE_PT4_POS                             4                                                       /**< PT4 Position */
#define MXC_F_PT_ENABLE_PT4                                 ((uint32_t)(0x00000001UL << MXC_F_PT_ENABLE_PT4_POS))   /**< PT4 Mask     */
#define MXC_F_PT_ENABLE_PT5_POS                             5                                                       /**< PT5 Position */
#define MXC_F_PT_ENABLE_PT5                                 ((uint32_t)(0x00000001UL << MXC_F_PT_ENABLE_PT5_POS))   /**< PT5 Mask     */
#define MXC_F_PT_ENABLE_PT6_POS                             6                                                       /**< PT6 Position */
#define MXC_F_PT_ENABLE_PT6                                 ((uint32_t)(0x00000001UL << MXC_F_PT_ENABLE_PT6_POS))   /**< PT6 Mask     */
#define MXC_F_PT_ENABLE_PT7_POS                             7                                                       /**< PT7 Position */
#define MXC_F_PT_ENABLE_PT7                                 ((uint32_t)(0x00000001UL << MXC_F_PT_ENABLE_PT7_POS))   /**< PT7 Mask     */
#define MXC_F_PT_ENABLE_PT8_POS                             8                                                       /**< PT8 Position */
#define MXC_F_PT_ENABLE_PT8                                 ((uint32_t)(0x00000001UL << MXC_F_PT_ENABLE_PT8_POS))   /**< PT8 Mask     */
#define MXC_F_PT_ENABLE_PT9_POS                             9                                                       /**< PT9 Position */
#define MXC_F_PT_ENABLE_PT9                                 ((uint32_t)(0x00000001UL << MXC_F_PT_ENABLE_PT9_POS))   /**< PT9 Mask     */
#define MXC_F_PT_ENABLE_PT10_POS                            10                                                      /**< PT10 Position */
#define MXC_F_PT_ENABLE_PT10                                ((uint32_t)(0x00000001UL << MXC_F_PT_ENABLE_PT10_POS))  /**< PT10 Mask     */
#define MXC_F_PT_ENABLE_PT11_POS                            11                                                      /**< PT11 Position */
#define MXC_F_PT_ENABLE_PT11                                ((uint32_t)(0x00000001UL << MXC_F_PT_ENABLE_PT11_POS))  /**< PT11 Mask     */
#define MXC_F_PT_ENABLE_PT12_POS                            12                                                      /**< PT12 Position */
#define MXC_F_PT_ENABLE_PT12                                ((uint32_t)(0x00000001UL << MXC_F_PT_ENABLE_PT12_POS))  /**< PT12 Mask     */
#define MXC_F_PT_ENABLE_PT13_POS                            13                                                      /**< PT13 Position */
#define MXC_F_PT_ENABLE_PT13                                ((uint32_t)(0x00000001UL << MXC_F_PT_ENABLE_PT13_POS))  /**< PT13 Mask     */
#define MXC_F_PT_ENABLE_PT14_POS                            14                                                      /**< PT14 Position */
#define MXC_F_PT_ENABLE_PT14                                ((uint32_t)(0x00000001UL << MXC_F_PT_ENABLE_PT14_POS))  /**< PT14 Mask     */
#define MXC_F_PT_ENABLE_PT15_POS                            15                                                      /**< PT15 Position */
#define MXC_F_PT_ENABLE_PT15                                ((uint32_t)(0x00000001UL << MXC_F_PT_ENABLE_PT15_POS))  /**< PT15 Mask     */
/**@} PT_ENABLE_Register*/
/**
 * @ingroup pulsetrain_registers
 * @defgroup PT_RESYNC_Register PT_RESYNC
 * @brief Field Positions and Masks
 * @{
 */
#define MXC_F_PT_RESYNC_PT0_POS                             0                                                       /**< PT0 Position */
#define MXC_F_PT_RESYNC_PT0                                 ((uint32_t)(0x00000001UL << MXC_F_PT_RESYNC_PT0_POS))   /**< PT0 Mask     */
#define MXC_F_PT_RESYNC_PT1_POS                             1                                                       /**< PT1 Position */
#define MXC_F_PT_RESYNC_PT1                                 ((uint32_t)(0x00000001UL << MXC_F_PT_RESYNC_PT1_POS))   /**< PT1 Mask     */
#define MXC_F_PT_RESYNC_PT2_POS                             2                                                       /**< PT2 Position */
#define MXC_F_PT_RESYNC_PT2                                 ((uint32_t)(0x00000001UL << MXC_F_PT_RESYNC_PT2_POS))   /**< PT2 Mask     */
#define MXC_F_PT_RESYNC_PT3_POS                             3                                                       /**< PT3 Position */
#define MXC_F_PT_RESYNC_PT3                                 ((uint32_t)(0x00000001UL << MXC_F_PT_RESYNC_PT3_POS))   /**< PT3 Mask     */
#define MXC_F_PT_RESYNC_PT4_POS                             4                                                       /**< PT4 Position */
#define MXC_F_PT_RESYNC_PT4                                 ((uint32_t)(0x00000001UL << MXC_F_PT_RESYNC_PT4_POS))   /**< PT4 Mask     */
#define MXC_F_PT_RESYNC_PT5_POS                             5                                                       /**< PT5 Position */
#define MXC_F_PT_RESYNC_PT5                                 ((uint32_t)(0x00000001UL << MXC_F_PT_RESYNC_PT5_POS))   /**< PT5 Mask     */
#define MXC_F_PT_RESYNC_PT6_POS                             6                                                       /**< PT6 Position */
#define MXC_F_PT_RESYNC_PT6                                 ((uint32_t)(0x00000001UL << MXC_F_PT_RESYNC_PT6_POS))   /**< PT6 Mask     */
#define MXC_F_PT_RESYNC_PT7_POS                             7                                                       /**< PT7 Position */
#define MXC_F_PT_RESYNC_PT7                                 ((uint32_t)(0x00000001UL << MXC_F_PT_RESYNC_PT7_POS))   /**< PT7 Mask     */
#define MXC_F_PT_RESYNC_PT8_POS                             8                                                       /**< PT8 Position */
#define MXC_F_PT_RESYNC_PT8                                 ((uint32_t)(0x00000001UL << MXC_F_PT_RESYNC_PT8_POS))   /**< PT8 Mask     */
#define MXC_F_PT_RESYNC_PT9_POS                             9                                                       /**< PT9 Position */
#define MXC_F_PT_RESYNC_PT9                                 ((uint32_t)(0x00000001UL << MXC_F_PT_RESYNC_PT9_POS))   /**< PT9 Mask     */
#define MXC_F_PT_RESYNC_PT10_POS                            10                                                      /**< PT10 Position */
#define MXC_F_PT_RESYNC_PT10                                ((uint32_t)(0x00000001UL << MXC_F_PT_RESYNC_PT10_POS))  /**< PT10 Mask     */
#define MXC_F_PT_RESYNC_PT11_POS                            11                                                      /**< PT11 Position */
#define MXC_F_PT_RESYNC_PT11                                ((uint32_t)(0x00000001UL << MXC_F_PT_RESYNC_PT11_POS))  /**< PT11 Mask     */
#define MXC_F_PT_RESYNC_PT12_POS                            12                                                      /**< PT12 Position */
#define MXC_F_PT_RESYNC_PT12                                ((uint32_t)(0x00000001UL << MXC_F_PT_RESYNC_PT12_POS))  /**< PT12 Mask     */
#define MXC_F_PT_RESYNC_PT13_POS                            13                                                      /**< PT13 Position */
#define MXC_F_PT_RESYNC_PT13                                ((uint32_t)(0x00000001UL << MXC_F_PT_RESYNC_PT13_POS))  /**< PT13 Mask     */
#define MXC_F_PT_RESYNC_PT14_POS                            14                                                      /**< PT14 Position */
#define MXC_F_PT_RESYNC_PT14                                ((uint32_t)(0x00000001UL << MXC_F_PT_RESYNC_PT14_POS))  /**< PT14 Mask     */
#define MXC_F_PT_RESYNC_PT15_POS                            15                                                      /**< PT15 Position */
#define MXC_F_PT_RESYNC_PT15                                ((uint32_t)(0x00000001UL << MXC_F_PT_RESYNC_PT15_POS))  /**< PT15 Mask     */
/**@} PT_RESYNC_Register*/
/**
 * @ingroup pulsetrain_registers
 * @defgroup PT_INTFL_Register PT_INTFL
 * @brief Field Positions and Masks
 * @{
 */
#define MXC_F_PT_INTFL_PT0_POS                              0                                                       /**< PT0 Position */
#define MXC_F_PT_INTFL_PT0                                  ((uint32_t)(0x00000001UL << MXC_F_PT_INTFL_PT0_POS))    /**< PT0 Mask     */
#define MXC_F_PT_INTFL_PT1_POS                              1                                                       /**< PT1 Position */
#define MXC_F_PT_INTFL_PT1                                  ((uint32_t)(0x00000001UL << MXC_F_PT_INTFL_PT1_POS))    /**< PT1 Mask     */
#define MXC_F_PT_INTFL_PT2_POS                              2                                                       /**< PT2 Position */
#define MXC_F_PT_INTFL_PT2                                  ((uint32_t)(0x00000001UL << MXC_F_PT_INTFL_PT2_POS))    /**< PT2 Mask     */
#define MXC_F_PT_INTFL_PT3_POS                              3                                                       /**< PT3 Position */
#define MXC_F_PT_INTFL_PT3                                  ((uint32_t)(0x00000001UL << MXC_F_PT_INTFL_PT3_POS))    /**< PT3 Mask     */
#define MXC_F_PT_INTFL_PT4_POS                              4                                                       /**< PT4 Position */
#define MXC_F_PT_INTFL_PT4                                  ((uint32_t)(0x00000001UL << MXC_F_PT_INTFL_PT4_POS))    /**< PT4 Mask     */
#define MXC_F_PT_INTFL_PT5_POS                              5                                                       /**< PT5 Position */
#define MXC_F_PT_INTFL_PT5                                  ((uint32_t)(0x00000001UL << MXC_F_PT_INTFL_PT5_POS))    /**< PT5 Mask     */
#define MXC_F_PT_INTFL_PT6_POS                              6                                                       /**< PT6 Position */
#define MXC_F_PT_INTFL_PT6                                  ((uint32_t)(0x00000001UL << MXC_F_PT_INTFL_PT6_POS))    /**< PT6 Mask     */
#define MXC_F_PT_INTFL_PT7_POS                              7                                                       /**< PT7 Position */
#define MXC_F_PT_INTFL_PT7                                  ((uint32_t)(0x00000001UL << MXC_F_PT_INTFL_PT7_POS))    /**< PT7 Mask     */
#define MXC_F_PT_INTFL_PT8_POS                              8                                                       /**< PT8 Position */
#define MXC_F_PT_INTFL_PT8                                  ((uint32_t)(0x00000001UL << MXC_F_PT_INTFL_PT8_POS))    /**< PT8 Mask     */
#define MXC_F_PT_INTFL_PT9_POS                              9                                                       /**< PT9 Position */
#define MXC_F_PT_INTFL_PT9                                  ((uint32_t)(0x00000001UL << MXC_F_PT_INTFL_PT9_POS))    /**< PT9 Mask     */
#define MXC_F_PT_INTFL_PT10_POS                             10                                                      /**< PT10 Position */
#define MXC_F_PT_INTFL_PT10                                 ((uint32_t)(0x00000001UL << MXC_F_PT_INTFL_PT10_POS))   /**< PT10 Mask     */
#define MXC_F_PT_INTFL_PT11_POS                             11                                                      /**< PT11 Position */
#define MXC_F_PT_INTFL_PT11                                 ((uint32_t)(0x00000001UL << MXC_F_PT_INTFL_PT11_POS))   /**< PT11 Mask     */
#define MXC_F_PT_INTFL_PT12_POS                             12                                                      /**< PT12 Position */
#define MXC_F_PT_INTFL_PT12                                 ((uint32_t)(0x00000001UL << MXC_F_PT_INTFL_PT12_POS))   /**< PT12 Mask     */
#define MXC_F_PT_INTFL_PT13_POS                             13                                                      /**< PT13 Position */
#define MXC_F_PT_INTFL_PT13                                 ((uint32_t)(0x00000001UL << MXC_F_PT_INTFL_PT13_POS))   /**< PT13 Mask     */
#define MXC_F_PT_INTFL_PT14_POS                             14                                                      /**< PT14 Position */
#define MXC_F_PT_INTFL_PT14                                 ((uint32_t)(0x00000001UL << MXC_F_PT_INTFL_PT14_POS))   /**< PT14 Mask     */
#define MXC_F_PT_INTFL_PT15_POS                             15                                                      /**< PT15 Position */
#define MXC_F_PT_INTFL_PT15                                 ((uint32_t)(0x00000001UL << MXC_F_PT_INTFL_PT15_POS))   /**< PT15 Mask     */
/**@} PT_INTFL_Register*/
/**
 * @ingroup pulsetrain_registers
 * @defgroup PT_INTEN_Register PT_INTEN
 * @brief Field Positions and Masks
 * @{
 */
#define MXC_F_PT_INTEN_PT0_POS                              0                                                       /**< PT0 Position */
#define MXC_F_PT_INTEN_PT0                                  ((uint32_t)(0x00000001UL << MXC_F_PT_INTEN_PT0_POS))    /**< PT0 Mask     */
#define MXC_F_PT_INTEN_PT1_POS                              1                                                       /**< PT1 Position */
#define MXC_F_PT_INTEN_PT1                                  ((uint32_t)(0x00000001UL << MXC_F_PT_INTEN_PT1_POS))    /**< PT1 Mask     */
#define MXC_F_PT_INTEN_PT2_POS                              2                                                       /**< PT2 Position */
#define MXC_F_PT_INTEN_PT2                                  ((uint32_t)(0x00000001UL << MXC_F_PT_INTEN_PT2_POS))    /**< PT2 Mask     */
#define MXC_F_PT_INTEN_PT3_POS                              3                                                       /**< PT3 Position */
#define MXC_F_PT_INTEN_PT3                                  ((uint32_t)(0x00000001UL << MXC_F_PT_INTEN_PT3_POS))    /**< PT3 Mask     */
#define MXC_F_PT_INTEN_PT4_POS                              4                                                       /**< PT4 Position */
#define MXC_F_PT_INTEN_PT4                                  ((uint32_t)(0x00000001UL << MXC_F_PT_INTEN_PT4_POS))    /**< PT4 Mask     */
#define MXC_F_PT_INTEN_PT5_POS                              5                                                       /**< PT5 Position */
#define MXC_F_PT_INTEN_PT5                                  ((uint32_t)(0x00000001UL << MXC_F_PT_INTEN_PT5_POS))    /**< PT5 Mask     */
#define MXC_F_PT_INTEN_PT6_POS                              6                                                       /**< PT6 Position */
#define MXC_F_PT_INTEN_PT6                                  ((uint32_t)(0x00000001UL << MXC_F_PT_INTEN_PT6_POS))    /**< PT6 Mask     */
#define MXC_F_PT_INTEN_PT7_POS                              7                                                       /**< PT7 Position */
#define MXC_F_PT_INTEN_PT7                                  ((uint32_t)(0x00000001UL << MXC_F_PT_INTEN_PT7_POS))    /**< PT7 Mask     */
#define MXC_F_PT_INTEN_PT8_POS                              8                                                       /**< PT8 Position */
#define MXC_F_PT_INTEN_PT8                                  ((uint32_t)(0x00000001UL << MXC_F_PT_INTEN_PT8_POS))    /**< PT8 Mask     */
#define MXC_F_PT_INTEN_PT9_POS                              9                                                       /**< PT9 Position */
#define MXC_F_PT_INTEN_PT9                                  ((uint32_t)(0x00000001UL << MXC_F_PT_INTEN_PT9_POS))    /**< PT9 Mask     */
#define MXC_F_PT_INTEN_PT10_POS                             10                                                      /**< PT10 Position*/
#define MXC_F_PT_INTEN_PT10                                 ((uint32_t)(0x00000001UL << MXC_F_PT_INTEN_PT10_POS))   /**< PT10 Mask    */
#define MXC_F_PT_INTEN_PT11_POS                             11                                                      /**< PT11 Position*/
#define MXC_F_PT_INTEN_PT11                                 ((uint32_t)(0x00000001UL << MXC_F_PT_INTEN_PT11_POS))   /**< PT11 Mask    */
#define MXC_F_PT_INTEN_PT12_POS                             12                                                      /**< PT12 Position*/
#define MXC_F_PT_INTEN_PT12                                 ((uint32_t)(0x00000001UL << MXC_F_PT_INTEN_PT12_POS))   /**< PT12 Mask    */
#define MXC_F_PT_INTEN_PT13_POS                             13                                                      /**< PT13 Position*/
#define MXC_F_PT_INTEN_PT13                                 ((uint32_t)(0x00000001UL << MXC_F_PT_INTEN_PT13_POS))   /**< PT13 Mask    */
#define MXC_F_PT_INTEN_PT14_POS                             14                                                      /**< PT14 Position*/
#define MXC_F_PT_INTEN_PT14                                 ((uint32_t)(0x00000001UL << MXC_F_PT_INTEN_PT14_POS))   /**< PT14 Mask    */
#define MXC_F_PT_INTEN_PT15_POS                             15                                                      /**< PT15 Position*/
#define MXC_F_PT_INTEN_PT15                                 ((uint32_t)(0x00000001UL << MXC_F_PT_INTEN_PT15_POS))   /**< PT15 Mask    */
/**@} PT_INTEN_Register*/
/**
 * @ingroup pulsetrain_registers
 * @defgroup PT_RATE_LENGTH_Register PT_RATE_LENGTH
 * @brief Field Positions and Masks
 * @{
 */
#define MXC_F_PT_RATE_LENGTH_RATE_CONTROL_POS               0                                                                   /**< RATE_CONTROL Position */
#define MXC_F_PT_RATE_LENGTH_RATE_CONTROL                   ((uint32_t)(0x07FFFFFFUL << MXC_F_PT_RATE_LENGTH_RATE_CONTROL_POS)) /**< RATE_CONTROL Mask */
#define MXC_F_PT_RATE_LENGTH_MODE_POS                       27                                                                  /**< MODE Position */
#define MXC_F_PT_RATE_LENGTH_MODE                           ((uint32_t)(0x0000001FUL << MXC_F_PT_RATE_LENGTH_MODE_POS))         /**< MODE Mask */
/**@} PT_RATE_Register*/
/**
 * @ingroup pulsetrain_registers
 * @defgroup PT_LOOP_Register PT_LOOP
 * @brief Field Positions and Masks
 * @{
 */
#define MXC_F_PT_LOOP_COUNT_POS                             0                                                       /**< COUNT Position */
#define MXC_F_PT_LOOP_COUNT                                 ((uint32_t)(0x0000FFFFUL << MXC_F_PT_LOOP_COUNT_POS))   /**< COUNT Mask */
#define MXC_F_PT_LOOP_DELAY_POS                             16                                                      /**< DELAY Position */
#define MXC_F_PT_LOOP_DELAY                                 ((uint32_t)(0x00000FFFUL << MXC_F_PT_LOOP_DELAY_POS))   /**< DELAY Mask */
/**@}PT_LOOP_Register*/
/**
 * @ingroup pulsetrain_registers
 * @defgroup PT_RESTART_Register PT_RESTART
 * @brief Field Positions and Masks
 * @{
 */
#define MXC_F_PT_RESTART_PT_X_SELECT_POS                    0                                                                     /**< PT_X_SELECT Position */
#define MXC_F_PT_RESTART_PT_X_SELECT                        ((uint32_t)(0x0000001FUL << MXC_F_PT_RESTART_PT_X_SELECT_POS))        /**< PT_X_SELECT Mask */
#define MXC_F_PT_RESTART_ON_PT_X_LOOP_EXIT_POS              7                                                                     /**< ON_PT_X_LOOP_EXIT Position */
#define MXC_F_PT_RESTART_ON_PT_X_LOOP_EXIT                  ((uint32_t)(0x00000001UL << MXC_F_PT_RESTART_ON_PT_X_LOOP_EXIT_POS))  /**< ON_PT_X_LOOP_EXIT Mask */
#define MXC_F_PT_RESTART_PT_Y_SELECT_POS                    8                                                                     /**< PT_Y_SELECT Position */
#define MXC_F_PT_RESTART_PT_Y_SELECT                        ((uint32_t)(0x0000001FUL << MXC_F_PT_RESTART_PT_Y_SELECT_POS))        /**< PT_Y_SELECT Mask */
#define MXC_F_PT_RESTART_ON_PT_Y_LOOP_EXIT_POS              15                                                                    /**< ON_PT_Y_LOOP_EXIT Position */
#define MXC_F_PT_RESTART_ON_PT_Y_LOOP_EXIT                  ((uint32_t)(0x00000001UL << MXC_F_PT_RESTART_ON_PT_Y_LOOP_EXIT_POS))  /**< ON_PT_Y_LOOP_EXIT Mask */
/**@} PT_RESTART_Register */


/*
   Field values and shifted values for module PT.
*/
/**
 * @ingroup PT_RATE_LENGTH_Register
 * @defgroup pt_mode_v_sv PT Rates
 * @brief   Pulse train rate/length values for the PT_RATE_LENGTH_MODE field.
 */
#define MXC_V_PT_RATE_LENGTH_MODE_32_BIT                                        ((uint32_t)(0x00000000UL))    /**< Value for 32-BIT.      */
#define MXC_V_PT_RATE_LENGTH_MODE_SQUARE_WAVE                                   ((uint32_t)(0x00000001UL))    /**< Value for SQUARE_WAVE. */
#define MXC_V_PT_RATE_LENGTH_MODE_2_BIT                                         ((uint32_t)(0x00000002UL))    /**< Value for 2-BIT.       */
#define MXC_V_PT_RATE_LENGTH_MODE_3_BIT                                         ((uint32_t)(0x00000003UL))    /**< Value for 3-BIT.       */
#define MXC_V_PT_RATE_LENGTH_MODE_4_BIT                                         ((uint32_t)(0x00000004UL))    /**< Value for 4-BIT.       */
#define MXC_V_PT_RATE_LENGTH_MODE_5_BIT                                         ((uint32_t)(0x00000005UL))    /**< Value for 5-BIT.       */
#define MXC_V_PT_RATE_LENGTH_MODE_6_BIT                                         ((uint32_t)(0x00000006UL))    /**< Value for 6-BIT.       */
#define MXC_V_PT_RATE_LENGTH_MODE_7_BIT                                         ((uint32_t)(0x00000007UL))    /**< Value for 7-BIT.       */
#define MXC_V_PT_RATE_LENGTH_MODE_8_BIT                                         ((uint32_t)(0x00000008UL))    /**< Value for 8-BIT.       */
#define MXC_V_PT_RATE_LENGTH_MODE_9_BIT                                         ((uint32_t)(0x00000009UL))    /**< Value for 9-BIT.       */
#define MXC_V_PT_RATE_LENGTH_MODE_10_BIT                                        ((uint32_t)(0x0000000AUL))    /**< Value for 10-BIT.      */
#define MXC_V_PT_RATE_LENGTH_MODE_11_BIT                                        ((uint32_t)(0x0000000BUL))    /**< Value for 11-BIT.      */
#define MXC_V_PT_RATE_LENGTH_MODE_12_BIT                                        ((uint32_t)(0x0000000CUL))    /**< Value for 12-BIT.      */
#define MXC_V_PT_RATE_LENGTH_MODE_13_BIT                                        ((uint32_t)(0x0000000DUL))    /**< Value for 13-BIT.      */
#define MXC_V_PT_RATE_LENGTH_MODE_14_BIT                                        ((uint32_t)(0x0000000EUL))    /**< Value for 14-BIT.      */
#define MXC_V_PT_RATE_LENGTH_MODE_15_BIT                                        ((uint32_t)(0x0000000FUL))    /**< Value for 15-BIT.      */
#define MXC_V_PT_RATE_LENGTH_MODE_16_BIT                                        ((uint32_t)(0x00000010UL))    /**< Value for 16-BIT.      */
#define MXC_V_PT_RATE_LENGTH_MODE_17_BIT                                        ((uint32_t)(0x00000011UL))    /**< Value for 17-BIT.      */
#define MXC_V_PT_RATE_LENGTH_MODE_18_BIT                                        ((uint32_t)(0x00000012UL))    /**< Value for 18-BIT.      */
#define MXC_V_PT_RATE_LENGTH_MODE_19_BIT                                        ((uint32_t)(0x00000013UL))    /**< Value for 19-BIT.      */
#define MXC_V_PT_RATE_LENGTH_MODE_20_BIT                                        ((uint32_t)(0x00000014UL))    /**< Value for 20-BIT.      */
#define MXC_V_PT_RATE_LENGTH_MODE_21_BIT                                        ((uint32_t)(0x00000015UL))    /**< Value for 21-BIT.      */
#define MXC_V_PT_RATE_LENGTH_MODE_22_BIT                                        ((uint32_t)(0x00000016UL))    /**< Value for 22-BIT.      */
#define MXC_V_PT_RATE_LENGTH_MODE_23_BIT                                        ((uint32_t)(0x00000017UL))    /**< Value for 23-BIT.      */
#define MXC_V_PT_RATE_LENGTH_MODE_24_BIT                                        ((uint32_t)(0x00000018UL))    /**< Value for 24-BIT.      */
#define MXC_V_PT_RATE_LENGTH_MODE_25_BIT                                        ((uint32_t)(0x00000019UL))    /**< Value for 25-BIT.      */
#define MXC_V_PT_RATE_LENGTH_MODE_26_BIT                                        ((uint32_t)(0x0000001AUL))    /**< Value for 26-BIT.      */
#define MXC_V_PT_RATE_LENGTH_MODE_27_BIT                                        ((uint32_t)(0x0000001BUL))    /**< Value for 27-BIT.      */
#define MXC_V_PT_RATE_LENGTH_MODE_28_BIT                                        ((uint32_t)(0x0000001CUL))    /**< Value for 28-BIT.      */
#define MXC_V_PT_RATE_LENGTH_MODE_29_BIT                                        ((uint32_t)(0x0000001DUL))    /**< Value for 29-BIT.      */
#define MXC_V_PT_RATE_LENGTH_MODE_30_BIT                                        ((uint32_t)(0x0000001EUL))    /**< Value for 30-BIT.      */
#define MXC_V_PT_RATE_LENGTH_MODE_31_BIT                                        ((uint32_t)(0x0000001FUL))    /**< Value for 31-BIT.      */

#define MXC_S_PT_RATE_LENGTH_MODE_32_BIT                                        ((uint32_t)(MXC_V_PT_RATE_LENGTH_MODE_32_BIT       << MXC_F_PT_RATE_LENGTH_MODE_POS))   /**< Shifted Value for 32-BIT.      */
#define MXC_S_PT_RATE_LENGTH_MODE_SQUARE_WAVE                                   ((uint32_t)(MXC_V_PT_RATE_LENGTH_MODE_SQUARE_WAVE  << MXC_F_PT_RATE_LENGTH_MODE_POS))   /**< Shifted Value for SQUARE_WAVE. */
#define MXC_S_PT_RATE_LENGTH_MODE_2_BIT                                         ((uint32_t)(MXC_V_PT_RATE_LENGTH_MODE_2_BIT        << MXC_F_PT_RATE_LENGTH_MODE_POS))   /**< Shifted Value for 2-BIT.       */
#define MXC_S_PT_RATE_LENGTH_MODE_3_BIT                                         ((uint32_t)(MXC_V_PT_RATE_LENGTH_MODE_3_BIT        << MXC_F_PT_RATE_LENGTH_MODE_POS))   /**< Shifted Value for 3-BIT.       */
#define MXC_S_PT_RATE_LENGTH_MODE_4_BIT                                         ((uint32_t)(MXC_V_PT_RATE_LENGTH_MODE_4_BIT        << MXC_F_PT_RATE_LENGTH_MODE_POS))   /**< Shifted Value for 4-BIT.       */
#define MXC_S_PT_RATE_LENGTH_MODE_5_BIT                                         ((uint32_t)(MXC_V_PT_RATE_LENGTH_MODE_5_BIT        << MXC_F_PT_RATE_LENGTH_MODE_POS))   /**< Shifted Value for 5-BIT.       */
#define MXC_S_PT_RATE_LENGTH_MODE_6_BIT                                         ((uint32_t)(MXC_V_PT_RATE_LENGTH_MODE_6_BIT        << MXC_F_PT_RATE_LENGTH_MODE_POS))   /**< Shifted Value for 6-BIT.       */
#define MXC_S_PT_RATE_LENGTH_MODE_7_BIT                                         ((uint32_t)(MXC_V_PT_RATE_LENGTH_MODE_7_BIT        << MXC_F_PT_RATE_LENGTH_MODE_POS))   /**< Shifted Value for 7-BIT.       */
#define MXC_S_PT_RATE_LENGTH_MODE_8_BIT                                         ((uint32_t)(MXC_V_PT_RATE_LENGTH_MODE_8_BIT        << MXC_F_PT_RATE_LENGTH_MODE_POS))   /**< Shifted Value for 8-BIT.       */
#define MXC_S_PT_RATE_LENGTH_MODE_9_BIT                                         ((uint32_t)(MXC_V_PT_RATE_LENGTH_MODE_9_BIT        << MXC_F_PT_RATE_LENGTH_MODE_POS))   /**< Shifted Value for 9-BIT.       */
#define MXC_S_PT_RATE_LENGTH_MODE_10_BIT                                        ((uint32_t)(MXC_V_PT_RATE_LENGTH_MODE_10_BIT       << MXC_F_PT_RATE_LENGTH_MODE_POS))   /**< Shifted Value for 10-BIT.      */
#define MXC_S_PT_RATE_LENGTH_MODE_11_BIT                                        ((uint32_t)(MXC_V_PT_RATE_LENGTH_MODE_11_BIT       << MXC_F_PT_RATE_LENGTH_MODE_POS))   /**< Shifted Value for 11-BIT.      */
#define MXC_S_PT_RATE_LENGTH_MODE_12_BIT                                        ((uint32_t)(MXC_V_PT_RATE_LENGTH_MODE_12_BIT       << MXC_F_PT_RATE_LENGTH_MODE_POS))   /**< Shifted Value for 12-BIT.      */
#define MXC_S_PT_RATE_LENGTH_MODE_13_BIT                                        ((uint32_t)(MXC_V_PT_RATE_LENGTH_MODE_13_BIT       << MXC_F_PT_RATE_LENGTH_MODE_POS))   /**< Shifted Value for 13-BIT.      */
#define MXC_S_PT_RATE_LENGTH_MODE_14_BIT                                        ((uint32_t)(MXC_V_PT_RATE_LENGTH_MODE_14_BIT       << MXC_F_PT_RATE_LENGTH_MODE_POS))   /**< Shifted Value for 14-BIT.      */
#define MXC_S_PT_RATE_LENGTH_MODE_15_BIT                                        ((uint32_t)(MXC_V_PT_RATE_LENGTH_MODE_15_BIT       << MXC_F_PT_RATE_LENGTH_MODE_POS))   /**< Shifted Value for 15-BIT.      */
#define MXC_S_PT_RATE_LENGTH_MODE_16_BIT                                        ((uint32_t)(MXC_V_PT_RATE_LENGTH_MODE_16_BIT       << MXC_F_PT_RATE_LENGTH_MODE_POS))   /**< Shifted Value for 16-BIT.      */
#define MXC_S_PT_RATE_LENGTH_MODE_17_BIT                                        ((uint32_t)(MXC_V_PT_RATE_LENGTH_MODE_17_BIT       << MXC_F_PT_RATE_LENGTH_MODE_POS))   /**< Shifted Value for 17-BIT.      */
#define MXC_S_PT_RATE_LENGTH_MODE_18_BIT                                        ((uint32_t)(MXC_V_PT_RATE_LENGTH_MODE_18_BIT       << MXC_F_PT_RATE_LENGTH_MODE_POS))   /**< Shifted Value for 18-BIT.      */
#define MXC_S_PT_RATE_LENGTH_MODE_19_BIT                                        ((uint32_t)(MXC_V_PT_RATE_LENGTH_MODE_19_BIT       << MXC_F_PT_RATE_LENGTH_MODE_POS))   /**< Shifted Value for 19-BIT.      */
#define MXC_S_PT_RATE_LENGTH_MODE_20_BIT                                        ((uint32_t)(MXC_V_PT_RATE_LENGTH_MODE_20_BIT       << MXC_F_PT_RATE_LENGTH_MODE_POS))   /**< Shifted Value for 20-BIT.      */
#define MXC_S_PT_RATE_LENGTH_MODE_21_BIT                                        ((uint32_t)(MXC_V_PT_RATE_LENGTH_MODE_21_BIT       << MXC_F_PT_RATE_LENGTH_MODE_POS))   /**< Shifted Value for 21-BIT.      */
#define MXC_S_PT_RATE_LENGTH_MODE_22_BIT                                        ((uint32_t)(MXC_V_PT_RATE_LENGTH_MODE_22_BIT       << MXC_F_PT_RATE_LENGTH_MODE_POS))   /**< Shifted Value for 22-BIT.      */
#define MXC_S_PT_RATE_LENGTH_MODE_23_BIT                                        ((uint32_t)(MXC_V_PT_RATE_LENGTH_MODE_23_BIT       << MXC_F_PT_RATE_LENGTH_MODE_POS))   /**< Shifted Value for 23-BIT.      */
#define MXC_S_PT_RATE_LENGTH_MODE_24_BIT                                        ((uint32_t)(MXC_V_PT_RATE_LENGTH_MODE_24_BIT       << MXC_F_PT_RATE_LENGTH_MODE_POS))   /**< Shifted Value for 24-BIT.      */
#define MXC_S_PT_RATE_LENGTH_MODE_25_BIT                                        ((uint32_t)(MXC_V_PT_RATE_LENGTH_MODE_25_BIT       << MXC_F_PT_RATE_LENGTH_MODE_POS))   /**< Shifted Value for 25-BIT.      */
#define MXC_S_PT_RATE_LENGTH_MODE_26_BIT                                        ((uint32_t)(MXC_V_PT_RATE_LENGTH_MODE_26_BIT       << MXC_F_PT_RATE_LENGTH_MODE_POS))   /**< Shifted Value for 26-BIT.      */
#define MXC_S_PT_RATE_LENGTH_MODE_27_BIT                                        ((uint32_t)(MXC_V_PT_RATE_LENGTH_MODE_27_BIT       << MXC_F_PT_RATE_LENGTH_MODE_POS))   /**< Shifted Value for 27-BIT.      */
#define MXC_S_PT_RATE_LENGTH_MODE_28_BIT                                        ((uint32_t)(MXC_V_PT_RATE_LENGTH_MODE_28_BIT       << MXC_F_PT_RATE_LENGTH_MODE_POS))   /**< Shifted Value for 28-BIT.      */
#define MXC_S_PT_RATE_LENGTH_MODE_29_BIT                                        ((uint32_t)(MXC_V_PT_RATE_LENGTH_MODE_29_BIT       << MXC_F_PT_RATE_LENGTH_MODE_POS))   /**< Shifted Value for 29-BIT.      */
#define MXC_S_PT_RATE_LENGTH_MODE_30_BIT                                        ((uint32_t)(MXC_V_PT_RATE_LENGTH_MODE_30_BIT       << MXC_F_PT_RATE_LENGTH_MODE_POS))   /**< Shifted Value for 30-BIT.      */
#define MXC_S_PT_RATE_LENGTH_MODE_31_BIT                                        ((uint32_t)(MXC_V_PT_RATE_LENGTH_MODE_31_BIT       << MXC_F_PT_RATE_LENGTH_MODE_POS))   /**< Shifted Value for 31-BIT.      */
/**@} pt_mode_v_sv*/

#ifdef __cplusplus
}
#endif

#endif   /* _MXC_PT_REGS_H_ */

