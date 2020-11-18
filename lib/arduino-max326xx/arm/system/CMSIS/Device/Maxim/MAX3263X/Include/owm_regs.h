/**
 * @file
 * @brief Type definitions for the 1-Wire Master Interface
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
 * $Date: 2016-10-10 19:22:03 -0500 (Mon, 10 Oct 2016) $
 * $Revision: 24666 $
 *
 **************************************************************************** */

/* Define to prevent redundant inclusion */
#ifndef _MXC_OWM_REGS_H_
#define _MXC_OWM_REGS_H_

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

/*
    If types are not defined elsewhere (CMSIS) define them here
*/
///@cond
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
 * @ingroup     owm
 * @defgroup    owm_registers Registers
 * @brief       Registers, Bit Masks and Bit Positions
 * @{
 */
/**
 * Structure type for the 1-Wire Master module registers allowing direct 32-bit access to each register. 
 */
typedef struct {
    __IO uint32_t cfg;                                  /**< <tt>\b 0x0000:</tt> OWM_CFG Register - 1-Wire Master Configuration           */
    __IO uint32_t clk_div_1us;                          /**< <tt>\b 0x0004:</tt> OWM_CLK_DIV_1US Register - 1-Wire Master Clock Divisor   */
    __IO uint32_t ctrl_stat;                            /**< <tt>\b 0x0008:</tt> OWM_CTRL_STAT Register - 1-Wire Master Control/Status    */
    __IO uint32_t data;                                 /**< <tt>\b 0x000C:</tt> OWM_DATA Register - 1-Wire Master Data Buffer            */
    __IO uint32_t intfl;                                /**< <tt>\b 0x0010:</tt> OWM_INTFL Register - 1-Wire Master Interrupt Flags       */
    __IO uint32_t inten;                                /**< <tt>\b 0x0014:</tt> OWM_INTEN Register - 1-Wire Master Interrupt Enables     */
} mxc_owm_regs_t;
/**@} end of group owm_registers */

/**
 * @ingroup    owm_registers
 * @defgroup   OWM_Register_Offsets Register Offsets
 * @brief      1-Wire Master register offsets from the 1-Wire Master Base Peripheral Address. 
 * @{
 */
#define MXC_R_OWM_OFFS_CFG                                  ((uint32_t)0x00000000UL)        /**< Offset from the OWM Base Peripheral Address:<tt>\b 0x0000:</tt>*/
#define MXC_R_OWM_OFFS_CLK_DIV_1US                          ((uint32_t)0x00000004UL)        /**< Offset from the OWM Base Peripheral Address:<tt>\b 0x0004:</tt>*/
#define MXC_R_OWM_OFFS_CTRL_STAT                            ((uint32_t)0x00000008UL)        /**< Offset from the OWM Base Peripheral Address:<tt>\b 0x0008:</tt>*/
#define MXC_R_OWM_OFFS_DATA                                 ((uint32_t)0x0000000CUL)        /**< Offset from the OWM Base Peripheral Address:<tt>\b 0x000C:</tt>*/
#define MXC_R_OWM_OFFS_INTFL                                ((uint32_t)0x00000010UL)        /**< Offset from the OWM Base Peripheral Address:<tt>\b 0x0010:</tt>*/
#define MXC_R_OWM_OFFS_INTEN                                ((uint32_t)0x00000014UL)        /**< Offset from the OWM Base Peripheral Address:<tt>\b 0x0014:</tt>*/
/**@} end of group OWM_Register_Offsets */

/*
   Field positions and masks for module OWM.
*/
/**
 * @ingroup     owm_registers
 * @defgroup    owm_cfg OWM_CFG 
 * @brief       Field Positions and Masks 
 */
#define MXC_F_OWM_CFG_LONG_LINE_MODE_POS                    0                                                                     /**< LONG_LINE_MODE Position  */
#define MXC_F_OWM_CFG_LONG_LINE_MODE                        ((uint32_t)(0x00000001UL << MXC_F_OWM_CFG_LONG_LINE_MODE_POS))        /**< LONG_LINE_MODE Mask  */
#define MXC_F_OWM_CFG_FORCE_PRES_DET_POS                    1                                                                     /**< FORCE_PRES_DET Position  */
#define MXC_F_OWM_CFG_FORCE_PRES_DET                        ((uint32_t)(0x00000001UL << MXC_F_OWM_CFG_FORCE_PRES_DET_POS))        /**< FORCE_PRES_DET Mask  */
#define MXC_F_OWM_CFG_BIT_BANG_EN_POS                       2                                                                     /**< BIT_BANG_EN Position  */
#define MXC_F_OWM_CFG_BIT_BANG_EN                           ((uint32_t)(0x00000001UL << MXC_F_OWM_CFG_BIT_BANG_EN_POS))           /**< BIT_BANG_EN Mask  */
#define MXC_F_OWM_CFG_EXT_PULLUP_MODE_POS                   3                                                                     /**< EXT_PULLUP_MODE Position  */
#define MXC_F_OWM_CFG_EXT_PULLUP_MODE                       ((uint32_t)(0x00000001UL << MXC_F_OWM_CFG_EXT_PULLUP_MODE_POS))       /**< EXT_PULLUP_MODE Mask  */
#define MXC_F_OWM_CFG_EXT_PULLUP_ENABLE_POS                 4                                                                     /**< EXT_PULLUP_ENABLE Position  */
#define MXC_F_OWM_CFG_EXT_PULLUP_ENABLE                     ((uint32_t)(0x00000001UL << MXC_F_OWM_CFG_EXT_PULLUP_ENABLE_POS))     /**< EXT_PULLUP_ENABLE Mask  */
#define MXC_F_OWM_CFG_SINGLE_BIT_MODE_POS                   5                                                                     /**< SINGLE_BIT_MODE Position  */
#define MXC_F_OWM_CFG_SINGLE_BIT_MODE                       ((uint32_t)(0x00000001UL << MXC_F_OWM_CFG_SINGLE_BIT_MODE_POS))       /**< SINGLE_BIT_MODE Mask  */
#define MXC_F_OWM_CFG_OVERDRIVE_POS                         6                                                                     /**< OVERDRIVE Position  */
#define MXC_F_OWM_CFG_OVERDRIVE                             ((uint32_t)(0x00000001UL << MXC_F_OWM_CFG_OVERDRIVE_POS))             /**< OVERDRIVE Mask  */
#define MXC_F_OWM_CFG_INT_PULLUP_ENABLE_POS                 7                                                                     /**< INT_PULLUP_ENABLE Position  */
#define MXC_F_OWM_CFG_INT_PULLUP_ENABLE                     ((uint32_t)(0x00000001UL << MXC_F_OWM_CFG_INT_PULLUP_ENABLE_POS))     /**< INT_PULLUP_ENABLE Mask  */
/**@} end of group owm_cfg*/
/**
 * @ingroup     owm_registers
 * @defgroup    owm_clk_div OWM_CLK_DIV
 * @brief       Field Positions and Masks 
 */
#define MXC_F_OWM_CLK_DIV_1US_DIVISOR_POS                   0                                                                     /**< 1US_DIVISOR Position  */
#define MXC_F_OWM_CLK_DIV_1US_DIVISOR                       ((uint32_t)(0x000000FFUL << MXC_F_OWM_CLK_DIV_1US_DIVISOR_POS))       /**< 1US_DIVISOR Mask  */
/**@} end of group owm_clk_cfg*/
/**
 * @ingroup     owm_registers
 * @defgroup    owm_ctrl_stat OWM_CTRL_STAT
 * @brief       Field Positions and Masks 
 */
#define MXC_F_OWM_CTRL_STAT_START_OW_RESET_POS              0                                                                     /**< START_OW_RESET Position  */
#define MXC_F_OWM_CTRL_STAT_START_OW_RESET                  ((uint32_t)(0x00000001UL << MXC_F_OWM_CTRL_STAT_START_OW_RESET_POS))  /**< START_OW_RESET Mask  */
#define MXC_F_OWM_CTRL_STAT_SRA_MODE_POS                    1                                                                     /**< SRA_MODE Position  */
#define MXC_F_OWM_CTRL_STAT_SRA_MODE                        ((uint32_t)(0x00000001UL << MXC_F_OWM_CTRL_STAT_SRA_MODE_POS))        /**< SRA_MODE Mask  */
#define MXC_F_OWM_CTRL_STAT_BIT_BANG_OE_POS                 2                                                                     /**< BIT_BANG_OE Position  */
#define MXC_F_OWM_CTRL_STAT_BIT_BANG_OE                     ((uint32_t)(0x00000001UL << MXC_F_OWM_CTRL_STAT_BIT_BANG_OE_POS))     /**< BIT_BANG_OE Mask  */
#define MXC_F_OWM_CTRL_STAT_OW_INPUT_POS                    3                                                                     /**< OW_INPUT Position  */
#define MXC_F_OWM_CTRL_STAT_OW_INPUT                        ((uint32_t)(0x00000001UL << MXC_F_OWM_CTRL_STAT_OW_INPUT_POS))        /**< OW_INPUT Mask  */
#define MXC_F_OWM_CTRL_STAT_OD_SPEC_MODE_POS                4                                                                     /**< OD_SPEC_MODE Position  */
#define MXC_F_OWM_CTRL_STAT_OD_SPEC_MODE                    ((uint32_t)(0x00000001UL << MXC_F_OWM_CTRL_STAT_OD_SPEC_MODE_POS))    /**< OD_SPEC_MODE Mask  */
#define MXC_F_OWM_CTRL_STAT_EXT_PULLUP_POL_POS              5                                                                     /**< EXT_PULLUP_POL Position  */
#define MXC_F_OWM_CTRL_STAT_EXT_PULLUP_POL                  ((uint32_t)(0x00000001UL << MXC_F_OWM_CTRL_STAT_EXT_PULLUP_POL_POS))  /**< EXT_PULLUP_POL Mask  */
#define MXC_F_OWM_CTRL_STAT_PRESENCE_DETECT_POS             7                                                                     /**< PRESENCE_DETECT Position  */
#define MXC_F_OWM_CTRL_STAT_PRESENCE_DETECT                 ((uint32_t)(0x00000001UL << MXC_F_OWM_CTRL_STAT_PRESENCE_DETECT_POS)) /**< PRESENCE_DETECT Mask  */
/**@} end of group owm_ctrl*/
/**
 * @ingroup     owm_registers
 * @defgroup    owm_data OWM_DATA
 * @brief       Field Positions and Masks 
 */
#define MXC_F_OWM_DATA_TX_RX_POS                            0                                                                     /**< TX_RX Position  */
#define MXC_F_OWM_DATA_TX_RX                                ((uint32_t)(0x000000FFUL << MXC_F_OWM_DATA_TX_RX_POS))                /**< TX_RX Mask  */
/**@} end of group owm_data*/
/**
 * @ingroup     owm_registers
 * @defgroup    owm_intfl OWM_INTFL
 * @brief       Field Positions and Masks 
 */
#define MXC_F_OWM_INTFL_OW_RESET_DONE_POS                   0                                                                     /**< OW_RESET_DONE Position  */
#define MXC_F_OWM_INTFL_OW_RESET_DONE                       ((uint32_t)(0x00000001UL << MXC_F_OWM_INTFL_OW_RESET_DONE_POS))       /**< OW_RESET_DONE Mask  */
#define MXC_F_OWM_INTFL_TX_DATA_EMPTY_POS                   1                                                                     /**< TX_DATA_EMPTY Position  */
#define MXC_F_OWM_INTFL_TX_DATA_EMPTY                       ((uint32_t)(0x00000001UL << MXC_F_OWM_INTFL_TX_DATA_EMPTY_POS))       /**< TX_DATA_EMPTY Mask  */
#define MXC_F_OWM_INTFL_RX_DATA_READY_POS                   2                                                                     /**< RX_DATA_READY Position  */
#define MXC_F_OWM_INTFL_RX_DATA_READY                       ((uint32_t)(0x00000001UL << MXC_F_OWM_INTFL_RX_DATA_READY_POS))       /**< RX_DATA_READY Mask  */
#define MXC_F_OWM_INTFL_LINE_SHORT_POS                      3                                                                     /**< LINE_SHORT Position  */
#define MXC_F_OWM_INTFL_LINE_SHORT                          ((uint32_t)(0x00000001UL << MXC_F_OWM_INTFL_LINE_SHORT_POS))          /**< LINE_SHORT Mask  */
#define MXC_F_OWM_INTFL_LINE_LOW_POS                        4                                                                     /**< LINE_LOW Position  */
#define MXC_F_OWM_INTFL_LINE_LOW                            ((uint32_t)(0x00000001UL << MXC_F_OWM_INTFL_LINE_LOW_POS))            /**< LINE_LOW Mask  */
/**@} end of group owm_intfl*/
/**
 * @ingroup     owm_registers
 * @defgroup    owm_inten OWM_INTEN
 * @brief       Field Positions and Masks 
 */
#define MXC_F_OWM_INTEN_OW_RESET_DONE_POS                   0                                                                     /**< OW_RESET_DONE Position  */
#define MXC_F_OWM_INTEN_OW_RESET_DONE                       ((uint32_t)(0x00000001UL << MXC_F_OWM_INTEN_OW_RESET_DONE_POS))       /**< OW_RESET_DONE Mask  */
#define MXC_F_OWM_INTEN_TX_DATA_EMPTY_POS                   1                                                                     /**< TX_DATA_EMPTY Position  */
#define MXC_F_OWM_INTEN_TX_DATA_EMPTY                       ((uint32_t)(0x00000001UL << MXC_F_OWM_INTEN_TX_DATA_EMPTY_POS))       /**< TX_DATA_EMPTY Mask  */
#define MXC_F_OWM_INTEN_RX_DATA_READY_POS                   2                                                                     /**< RX_DATA_READY Position  */
#define MXC_F_OWM_INTEN_RX_DATA_READY                       ((uint32_t)(0x00000001UL << MXC_F_OWM_INTEN_RX_DATA_READY_POS))       /**< RX_DATA_READY Mask  */
#define MXC_F_OWM_INTEN_LINE_SHORT_POS                      3                                                                     /**< LINE_SHORT Position  */
#define MXC_F_OWM_INTEN_LINE_SHORT                          ((uint32_t)(0x00000001UL << MXC_F_OWM_INTEN_LINE_SHORT_POS))          /**< LINE_SHORT Mask  */
#define MXC_F_OWM_INTEN_LINE_LOW_POS                        4                                                                     /**< LINE_LOW Position  */
#define MXC_F_OWM_INTEN_LINE_LOW                            ((uint32_t)(0x00000001UL << MXC_F_OWM_INTEN_LINE_LOW_POS))            /**< LINE_LOW Mask  */
/**@} end of group owm_inten*/
/**
 * @ingroup     owm_cfg
 * @{
 */
#define MXC_V_OWM_CFG_EXT_PULLUP_MODE_UNUSED                ((uint32_t)(0x00000000UL))        /**< External Pullup Mode Value: Unused */
#define MXC_V_OWM_CFG_EXT_PULLUP_MODE_USED                  ((uint32_t)(0x00000001UL))        /**< External Pullup Mode Value: Used */
/**@}*/
/**
 * @ingroup    owm_ctrl_stat
 * @{
 */
#define MXC_V_OWM_CTRL_STAT_OD_SPEC_MODE_12US               ((uint32_t)(0x00000000UL))        /**< Overdrive speed setting 12us. */
#define MXC_V_OWM_CTRL_STAT_OD_SPEC_MODE_10US               ((uint32_t)(0x00000001UL))        /**< Overdrive speed setting 10us. */

#define MXC_V_OWM_CTRL_STAT_EXT_PULLUP_POL_ACT_HIGH         ((uint32_t)(0x00000000UL))        /**< External Pullup Pin Polarity Active High */
#define MXC_V_OWM_CTRL_STAT_EXT_PULLUP_POL_ACT_LOW          ((uint32_t)(0x00000001UL))        /**< External Pullup Pin Polarity Active Low */
/**@}*/

#ifdef __cplusplus
}
#endif

#endif   /* _MXC_OWM_REGS_H_ */
