/**
 * @file
 * @brief   Registers, Bit Masks and Bit Positions for the ADC Peripheral Module.
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
 * $Date: 2017-02-14 18:08:19 -0600 (Tue, 14 Feb 2017) $
 * $Revision: 26419 $
 *
 *************************************************************************** */

/* Define to prevent redundant inclusion */
#ifndef _MXC_ADC_REGS_H_
#define _MXC_ADC_REGS_H_
/// @cond
/* **** Includes **** */
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif


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
/// @endcond

/* **** Definitions **** */
/**
 * @ingroup     group_adc
 * @defgroup    adc_registers Registers
 * @brief       Registers, Bit Masks and Bit Positions
 * @{
 */

/**
 * Structure type to access the ADC Registers.
 */
typedef struct {
    __IO uint32_t ctrl;                                     /**< <b><tt>0x000</tt>:</b> ADC CTRL Register                            */
    __IO uint32_t status;                                   /**< <b><tt>0x004</tt>:</b> ADC STATUS Register                          */
    __IO uint32_t data;                                     /**< <b><tt>0x008</tt>:</b> ADC DATA Register                            */
    __IO uint32_t intr;                                     /**< <b><tt>0x00C</tt>:</b> ADC INTR Register                            */
    __IO uint32_t limit[4];                                 /**< <b><tt>0x010</tt>:</b> ADC LIMIT0, LIMIT1, LIMIT2, LIMIT3 Register  */
    __IO uint32_t afe_ctrl;                                 /**< <b><tt>0x020</tt>:</b> ADC AFE_CTRL Register                        */
    __IO uint32_t ro_cal0;                                  /**< <b><tt>0x024</tt>:</b> ADC RO_CAL0 Register                         */
    __IO uint32_t ro_cal1;                                  /**< <b><tt>0x028</tt>:</b> ADC RO_CAL1 Register                         */
    __IO uint32_t ro_cal2;                                  /**< <b><tt>0x02C</tt>:</b> ADC RO_CAL2 Register                         */
} mxc_adc_regs_t;
/** @} */

/**
 * @ingroup    adc_registers
 * @defgroup   ADC_Register_Offsets Register Offsets
 * @brief      ADC Peripheral Register Offsets from the ADC Base Peripheral Address. 
 * @{
 */
#define MXC_R_ADC_OFFS_CTRL                                 ((uint32_t)0x00000000UL) /**< Offset from ADC Base Address: <b><tt>0x000</tt></b>  */
#define MXC_R_ADC_OFFS_STATUS                               ((uint32_t)0x00000004UL) /**< Offset from ADC Base Address: <b><tt>0x004</tt></b>  */
#define MXC_R_ADC_OFFS_DATA                                 ((uint32_t)0x00000008UL) /**< Offset from ADC Base Address: <b><tt>0x008</tt></b>  */
#define MXC_R_ADC_OFFS_INTR                                 ((uint32_t)0x0000000CUL) /**< Offset from ADC Base Address: <b><tt>0x00C</tt></b>  */
#define MXC_R_ADC_OFFS_LIMIT0                               ((uint32_t)0x00000010UL) /**< Offset from ADC Base Address: <b><tt>0x010</tt></b>  */
#define MXC_R_ADC_OFFS_LIMIT1                               ((uint32_t)0x00000014UL) /**< Offset from ADC Base Address: <b><tt>0x014</tt></b>  */
#define MXC_R_ADC_OFFS_LIMIT2                               ((uint32_t)0x00000018UL) /**< Offset from ADC Base Address: <b><tt>0x018</tt></b>  */
#define MXC_R_ADC_OFFS_LIMIT3                               ((uint32_t)0x0000001CUL) /**< Offset from ADC Base Address: <b><tt>0x01C</tt></b>  */
#define MXC_R_ADC_OFFS_AFE_CTRL                             ((uint32_t)0x00000020UL) /**< Offset from ADC Base Address: <b><tt>0x020</tt></b>  */
#define MXC_R_ADC_OFFS_RO_CAL0                              ((uint32_t)0x00000024UL) /**< Offset from ADC Base Address: <b><tt>0x024</tt></b>  */
#define MXC_R_ADC_OFFS_RO_CAL1                              ((uint32_t)0x00000028UL) /**< Offset from ADC Base Address: <b><tt>0x028</tt></b>  */
#define MXC_R_ADC_OFFS_RO_CAL2                              ((uint32_t)0x0000002CUL) /**< Offset from ADC Base Address: <b><tt>0x02C</tt></b>  */
/** @} end of group  */

/**
 * @ingroup  adc_registers
 * @defgroup ADC_CTRL_Register ADC_CTRL
 * @brief    Field Positions and Bit Masks for the ADC_CTRL register
 * @{
 */
#define MXC_F_ADC_CTRL_CPU_ADC_START_POS                    0                                                                   /**< CPU_ADC_START Position         */
#define MXC_F_ADC_CTRL_CPU_ADC_START                        ((uint32_t)(0x00000001UL << MXC_F_ADC_CTRL_CPU_ADC_START_POS))      /**< CPU_ADC_START Mask             */
#define MXC_F_ADC_CTRL_ADC_PU_POS                           1                                                                   /**< ADC_PU Position                */ 
#define MXC_F_ADC_CTRL_ADC_PU                               ((uint32_t)(0x00000001UL << MXC_F_ADC_CTRL_ADC_PU_POS))             /**< ADC_PU Mask                    */
#define MXC_F_ADC_CTRL_BUF_PU_POS                           2                                                                   /**< BUF_PU Position                */
#define MXC_F_ADC_CTRL_BUF_PU                               ((uint32_t)(0x00000001UL << MXC_F_ADC_CTRL_BUF_PU_POS))             /**< BUF_PU Mask                    */
#define MXC_F_ADC_CTRL_ADC_REFBUF_PU_POS                    3                                                                   /**< REFBUF_PU Position             */
#define MXC_F_ADC_CTRL_ADC_REFBUF_PU                        ((uint32_t)(0x00000001UL << MXC_F_ADC_CTRL_ADC_REFBUF_PU_POS))      /**< REFBUF_PU Mask                 */
#define MXC_F_ADC_CTRL_ADC_CHGPUMP_PU_POS                   4                                                                   /**< CHGPUMP_PU Position            */
#define MXC_F_ADC_CTRL_ADC_CHGPUMP_PU                       ((uint32_t)(0x00000001UL << MXC_F_ADC_CTRL_ADC_CHGPUMP_PU_POS))     /**< CHGPUMP_PU Mask                */
#define MXC_F_ADC_CTRL_BUF_CHOP_DIS_POS                     5                                                                   /**< BUF_CHOP_DIS Position          */
#define MXC_F_ADC_CTRL_BUF_CHOP_DIS                         ((uint32_t)(0x00000001UL << MXC_F_ADC_CTRL_BUF_CHOP_DIS_POS))       /**< BUF_CHOP_DIS Mask              */
#define MXC_F_ADC_CTRL_BUF_PUMP_DIS_POS                     6                                                                   /**< BUF_PUMP_DIS Position          */
#define MXC_F_ADC_CTRL_BUF_PUMP_DIS                         ((uint32_t)(0x00000001UL << MXC_F_ADC_CTRL_BUF_PUMP_DIS_POS))       /**< BUF_PUMP_DIS Mask              */
#define MXC_F_ADC_CTRL_BUF_BYPASS_POS                       7                                                                   /**< BUF_BYPASS Position            */
#define MXC_F_ADC_CTRL_BUF_BYPASS                           ((uint32_t)(0x00000001UL << MXC_F_ADC_CTRL_BUF_BYPASS_POS))         /**< BUF_BYPASS Mask                */
#define MXC_F_ADC_CTRL_ADC_REFSCL_POS                       8                                                                   /**< ADC_REFSCL Position            */
#define MXC_F_ADC_CTRL_ADC_REFSCL                           ((uint32_t)(0x00000001UL << MXC_F_ADC_CTRL_ADC_REFSCL_POS))         /**< ADC_REFSCL Mask                */
#define MXC_F_ADC_CTRL_ADC_SCALE_POS                        9                                                                   /**< ADC_SCALE Position             */
#define MXC_F_ADC_CTRL_ADC_SCALE                            ((uint32_t)(0x00000001UL << MXC_F_ADC_CTRL_ADC_SCALE_POS))          /**< ADC_SCALE Mask                 */
#define MXC_F_ADC_CTRL_ADC_REFSEL_POS                       10                                                                  /**< ADC_REFSEL Position            */ 
#define MXC_F_ADC_CTRL_ADC_REFSEL                           ((uint32_t)(0x00000001UL << MXC_F_ADC_CTRL_ADC_REFSEL_POS))         /**< ADC_REFSEL Mask                */
#define MXC_F_ADC_CTRL_ADC_CLK_EN_POS                       11                                                                  /**< ADC_CLK_EN Position            */ 
#define MXC_F_ADC_CTRL_ADC_CLK_EN                           ((uint32_t)(0x00000001UL << MXC_F_ADC_CTRL_ADC_CLK_EN_POS))         /**< ADC_CLK_EN Mask                */
#define MXC_F_ADC_CTRL_ADC_CHSEL_POS                        12                                                                  /**< ADC_CHSEL Position             */ 
#define MXC_F_ADC_CTRL_ADC_CHSEL                            ((uint32_t)(0x0000000FUL << MXC_F_ADC_CTRL_ADC_CHSEL_POS))          /**< ADC_CHSEL Mask                 */

#if (MXC_ADC_REV == 0)
#define MXC_F_ADC_CTRL_ADC_XREF_POS                         16                                                                  /**< ADC_XREF Position              */ 
#define MXC_F_ADC_CTRL_ADC_XREF                             ((uint32_t)(0x00000001UL << MXC_F_ADC_CTRL_ADC_XREF_POS))           /**< ADC_XREF Mask                  */
#endif
#define MXC_F_ADC_CTRL_ADC_DATAALIGN_POS                    17                                                                  /**< ADC_DATAALIGN Position         */
#define MXC_F_ADC_CTRL_ADC_DATAALIGN                        ((uint32_t)(0x00000001UL << MXC_F_ADC_CTRL_ADC_DATAALIGN_POS))      /**< ADC_DATAALIGN Mask             */
#define MXC_F_ADC_CTRL_AFE_PWR_UP_DLY_POS                   24                                                                  /**< AFE_PWR_UP_DLY Position        */ 
#define MXC_F_ADC_CTRL_AFE_PWR_UP_DLY                       ((uint32_t)(0x000000FFUL << MXC_F_ADC_CTRL_AFE_PWR_UP_DLY_POS))     /**< AFE_PWR_UP_DLY Mask            */

/** @} end of group adc_ctrl_register */

/**
 * @ingroup  adc_registers
 * @defgroup ADC_STATUS_Register ADC_STATUS
 * @brief    Field Positions and Bit Masks for the ADC_STATUS register
 * @{
 */
#define MXC_F_ADC_STATUS_ADC_ACTIVE_POS                     0                                                                       /**< ADC_ACTIVE Position            */
#define MXC_F_ADC_STATUS_ADC_ACTIVE                         ((uint32_t)(0x00000001UL << MXC_F_ADC_STATUS_ADC_ACTIVE_POS))           /**< ADC_ACTIVE Mask                */
#define MXC_F_ADC_STATUS_RO_CAL_ATOMIC_ACTIVE_POS           1                                                                       /**< RO_CAL_ATOMIC_ACTIVE Position  */
#define MXC_F_ADC_STATUS_RO_CAL_ATOMIC_ACTIVE               ((uint32_t)(0x00000001UL << MXC_F_ADC_STATUS_RO_CAL_ATOMIC_ACTIVE_POS)) /**< RO_CAL_ATOMIC_ACTIVE Mask      */
#define MXC_F_ADC_STATUS_AFE_PWR_UP_ACTIVE_POS              2                                                                       /**< AFE_PWR_UP_ACTIVE Position     */
#define MXC_F_ADC_STATUS_AFE_PWR_UP_ACTIVE                  ((uint32_t)(0x00000001UL << MXC_F_ADC_STATUS_AFE_PWR_UP_ACTIVE_POS))    /**< AFE_PWR_UP_ACTIVE Mask         */
#define MXC_F_ADC_STATUS_ADC_OVERFLOW_POS                   3                                                                       /**< ADC_OVERFLOW Position          */
#define MXC_F_ADC_STATUS_ADC_OVERFLOW                       ((uint32_t)(0x00000001UL << MXC_F_ADC_STATUS_ADC_OVERFLOW_POS))         /**< ADC_OVERFLOW Mask              */
/** @} end of group ADC_STATUS_register */

/**
 * @ingroup  adc_registers
 * @defgroup ADC_DATA_Register ADC_DATA
 * @brief    Field Positions and Bit Masks for the ADC_DATA register
 * @{
 */
#define MXC_F_ADC_DATA_ADC_DATA_POS                         0                                                                       /**< ADC_DATA Position          */
#define MXC_F_ADC_DATA_ADC_DATA                             ((uint32_t)(0x0000FFFFUL << MXC_F_ADC_DATA_ADC_DATA_POS))               /**< ADC_DATA Mask              */
/** @} end of group ADC_DATA_register */

/**
 * @ingroup     adc_registers
 * @defgroup    ADC_INTR_Register ADC_INTR Register
 * @brief       Interrupt Enable and Interrupt Flag Field Positions and Bit Masks
 */
/**
 * @ingroup     ADC_INTR_Register
 * @defgroup    ADC_INTR_IE_Register Interrupt Enable Bits
 * @brief       Interrupt Enable Bit Positions and Masks
 * @{
 */  
#define MXC_F_ADC_INTR_ADC_DONE_IE_POS                      0                                                                   /**< ADC_DONE_IE Position       */
#define MXC_F_ADC_INTR_ADC_DONE_IE                          ((uint32_t)(0x00000001UL << MXC_F_ADC_INTR_ADC_DONE_IE_POS))        /**< ADC_DONE_IE Mask           */
#define MXC_F_ADC_INTR_ADC_REF_READY_IE_POS                 1                                                                   /**< ADC_REF_READY_IE Position  */
#define MXC_F_ADC_INTR_ADC_REF_READY_IE                     ((uint32_t)(0x00000001UL << MXC_F_ADC_INTR_ADC_REF_READY_IE_POS))   /**< ADC_REF_READY_IE Mask      */
#define MXC_F_ADC_INTR_ADC_HI_LIMIT_IE_POS                  2                                                                   /**< ADC_HI_LIMIT_IE Position   */
#define MXC_F_ADC_INTR_ADC_HI_LIMIT_IE                      ((uint32_t)(0x00000001UL << MXC_F_ADC_INTR_ADC_HI_LIMIT_IE_POS))    /**< ADC_HI_LIMIT_IE Mask       */
#define MXC_F_ADC_INTR_ADC_LO_LIMIT_IE_POS                  3                                                                   /**< ADC_LO_LIMIT_IE Position   */
#define MXC_F_ADC_INTR_ADC_LO_LIMIT_IE                      ((uint32_t)(0x00000001UL << MXC_F_ADC_INTR_ADC_LO_LIMIT_IE_POS))    /**< ADC_LO_LIMIT_IE Mask       */
#define MXC_F_ADC_INTR_ADC_OVERFLOW_IE_POS                  4                                                                   /**< ADC_OVERFLOW_IE Position   */
#define MXC_F_ADC_INTR_ADC_OVERFLOW_IE                      ((uint32_t)(0x00000001UL << MXC_F_ADC_INTR_ADC_OVERFLOW_IE_POS))    /**< ADC_OVERFLOW_IE Mask       */
#define MXC_F_ADC_INTR_RO_CAL_DONE_IE_POS                   5                                                                   /**< RO_CAL_DONE_IE Position    */
#define MXC_F_ADC_INTR_RO_CAL_DONE_IE                       ((uint32_t)(0x00000001UL << MXC_F_ADC_INTR_RO_CAL_DONE_IE_POS))     /**< RO_CAL_DONE_IE Mask        */
/** @} end of group ADC_INTR_IE_Register */


/**
 * @ingroup     ADC_INTR_Register
 * @defgroup    ADC_INTR_IF_Register Interrupt Flag Bits
 * @brief       Interrupt Flag Bit Positions and Masks
 * @{
 */
#define MXC_F_ADC_INTR_ADC_DONE_IF_POS                      16                                                                  /**< ADC_DONE_IF Position           */
#define MXC_F_ADC_INTR_ADC_DONE_IF                          ((uint32_t)(0x00000001UL << MXC_F_ADC_INTR_ADC_DONE_IF_POS))        /**< ADC_DONE_IF Mask               */
#define MXC_F_ADC_INTR_ADC_REF_READY_IF_POS                 17                                                                  /**< ADC_REF_READY_IF Position      */
#define MXC_F_ADC_INTR_ADC_REF_READY_IF                     ((uint32_t)(0x00000001UL << MXC_F_ADC_INTR_ADC_REF_READY_IF_POS))   /**< ADC_REF_READY_IF Mask          */
#define MXC_F_ADC_INTR_ADC_HI_LIMIT_IF_POS                  18                                                                  /**< ADC_HI_LIMIT_IF Position       */
#define MXC_F_ADC_INTR_ADC_HI_LIMIT_IF                      ((uint32_t)(0x00000001UL << MXC_F_ADC_INTR_ADC_HI_LIMIT_IF_POS))    /**< ADC_HI_LIMIT_IF Mask           */
#define MXC_F_ADC_INTR_ADC_LO_LIMIT_IF_POS                  19                                                                  /**< ADC_LO_LIMIT_IF Position       */
#define MXC_F_ADC_INTR_ADC_LO_LIMIT_IF                      ((uint32_t)(0x00000001UL << MXC_F_ADC_INTR_ADC_LO_LIMIT_IF_POS))    /**< ADC_LO_LIMIT_IF Mask           */
#define MXC_F_ADC_INTR_ADC_OVERFLOW_IF_POS                  20                                                                  /**< ADC_OVERFLOW_IF Position       */
#define MXC_F_ADC_INTR_ADC_OVERFLOW_IF                      ((uint32_t)(0x00000001UL << MXC_F_ADC_INTR_ADC_OVERFLOW_IF_POS))    /**< ADC_OVERFLOW_IF Mask           */
#define MXC_F_ADC_INTR_RO_CAL_DONE_IF_POS                   21                                                                  /**< RO_CAL_DONE_IF Position        */
#define MXC_F_ADC_INTR_RO_CAL_DONE_IF                       ((uint32_t)(0x00000001UL << MXC_F_ADC_INTR_RO_CAL_DONE_IF_POS))     /**< RO_CAL_DONE_IF Mask            */
#define MXC_F_ADC_INTR_ADC_INT_PENDING_POS                  22                                                                  /**< ADC_INT_PENDING Position       */
#define MXC_F_ADC_INTR_ADC_INT_PENDING                      ((uint32_t)(0x00000001UL << MXC_F_ADC_INTR_ADC_INT_PENDING_POS))    /**< ADC_INT_PENDING Mask           */
/** @} end of group ADC_INTR_IF_Register */

/**
 * @ingroup  adc_registers
 * @defgroup ADC_LIMIT0_Register ADC_LIMIT0
 * @brief     Field Positions and Bit Masks for the ADC_LIMIT0 register
 * @{
 */
#define MXC_F_ADC_LIMIT0_CH_LO_LIMIT_POS                    0                                                                   /**< CH_LO_LIMIT Position       */
#define MXC_F_ADC_LIMIT0_CH_LO_LIMIT                        ((uint32_t)(0x000003FFUL << MXC_F_ADC_LIMIT0_CH_LO_LIMIT_POS))      /**< CH_LO_LIMIT Mask           */
#define MXC_F_ADC_LIMIT0_CH_HI_LIMIT_POS                    12                                                                  /**< CH_HI_LIMIT Position       */
#define MXC_F_ADC_LIMIT0_CH_HI_LIMIT                        ((uint32_t)(0x000003FFUL << MXC_F_ADC_LIMIT0_CH_HI_LIMIT_POS))      /**< CH_HI_LIMIT Mask           */
#define MXC_F_ADC_LIMIT0_CH_SEL_POS                         24                                                                  /**< CH_SEL Position            */
#define MXC_F_ADC_LIMIT0_CH_SEL                             ((uint32_t)(0x0000000FUL << MXC_F_ADC_LIMIT0_CH_SEL_POS))           /**< CH_SEL Mask                */
#define MXC_F_ADC_LIMIT0_CH_LO_LIMIT_EN_POS                 28                                                                  /**< CH_LO_LIMIT_EN Position    */
#define MXC_F_ADC_LIMIT0_CH_LO_LIMIT_EN                     ((uint32_t)(0x00000001UL << MXC_F_ADC_LIMIT0_CH_LO_LIMIT_EN_POS))   /**< CH_LO_LIMIT_EN Mask        */
#define MXC_F_ADC_LIMIT0_CH_HI_LIMIT_EN_POS                 29                                                                  /**< CH_HI_LIMIT_EN Position    */
#define MXC_F_ADC_LIMIT0_CH_HI_LIMIT_EN                     ((uint32_t)(0x00000001UL << MXC_F_ADC_LIMIT0_CH_HI_LIMIT_EN_POS))   /**< CH_HI_LIMIT_EN Mask        */
/** @} end of group ADC_LIMIT0_register */

/**
 * @ingroup  adc_registers
 * @defgroup ADC_LIMIT1_Register ADC_LIMIT1
 * @brief     Field Positions and Bit Masks for the ADC_LIMIT1 register
 * @{
 */
#define MXC_F_ADC_LIMIT1_CH_LO_LIMIT_POS                    0                                                                   /**< CH_LO_LIMIT Position       */    
#define MXC_F_ADC_LIMIT1_CH_LO_LIMIT                        ((uint32_t)(0x000003FFUL << MXC_F_ADC_LIMIT1_CH_LO_LIMIT_POS))      /**< CH_LO_LIMIT Mask           */
#define MXC_F_ADC_LIMIT1_CH_HI_LIMIT_POS                    12                                                                  /**< CH_HI_LIMIT Position       */    
#define MXC_F_ADC_LIMIT1_CH_HI_LIMIT                        ((uint32_t)(0x000003FFUL << MXC_F_ADC_LIMIT1_CH_HI_LIMIT_POS))      /**< CH_HI_LIMIT Mask           */
#define MXC_F_ADC_LIMIT1_CH_SEL_POS                         24                                                                  /**< CH_SEL Position            */    
#define MXC_F_ADC_LIMIT1_CH_SEL                             ((uint32_t)(0x0000000FUL << MXC_F_ADC_LIMIT1_CH_SEL_POS))           /**< CH_SEL Mask                */
#define MXC_F_ADC_LIMIT1_CH_LO_LIMIT_EN_POS                 28                                                                  /**< CH_LO_LIMIT_EN Position    */    
#define MXC_F_ADC_LIMIT1_CH_LO_LIMIT_EN                     ((uint32_t)(0x00000001UL << MXC_F_ADC_LIMIT1_CH_LO_LIMIT_EN_POS))   /**< CH_LO_LIMIT_EN Mask        */
#define MXC_F_ADC_LIMIT1_CH_HI_LIMIT_EN_POS                 29                                                                  /**< CH_HI_LIMIT_EN Position    */    
#define MXC_F_ADC_LIMIT1_CH_HI_LIMIT_EN                     ((uint32_t)(0x00000001UL << MXC_F_ADC_LIMIT1_CH_HI_LIMIT_EN_POS))   /**< CH_HI_LIMIT_EN Mask        */
/** @} end of group ADC_LIMIT1_register */

/**
 * @ingroup  adc_registers
 * @defgroup ADC_LIMIT2_Register ADC_LIMIT2
 * @brief     Field Positions and Bit Masks for the ADC_LIMIT2 register
 * @{
 */
#define MXC_F_ADC_LIMIT2_CH_LO_LIMIT_POS                    0                                                                   /**< CH_LO_LIMIT Position       */   
#define MXC_F_ADC_LIMIT2_CH_LO_LIMIT                        ((uint32_t)(0x000003FFUL << MXC_F_ADC_LIMIT2_CH_LO_LIMIT_POS))      /**< CH_LO_LIMIT Mask           */
#define MXC_F_ADC_LIMIT2_CH_HI_LIMIT_POS                    12                                                                  /**< CH_HI_LIMIT Position       */    
#define MXC_F_ADC_LIMIT2_CH_HI_LIMIT                        ((uint32_t)(0x000003FFUL << MXC_F_ADC_LIMIT2_CH_HI_LIMIT_POS))      /**< CH_HI_LIMIT Mask           */
#define MXC_F_ADC_LIMIT2_CH_SEL_POS                         24                                                                  /**< CH_SEL Position            */    
#define MXC_F_ADC_LIMIT2_CH_SEL                             ((uint32_t)(0x0000000FUL << MXC_F_ADC_LIMIT2_CH_SEL_POS))           /**< CH_SEL Mask                */
#define MXC_F_ADC_LIMIT2_CH_LO_LIMIT_EN_POS                 28                                                                  /**< CH_LO_LIMIT_EN Position    */    
#define MXC_F_ADC_LIMIT2_CH_LO_LIMIT_EN                     ((uint32_t)(0x00000001UL << MXC_F_ADC_LIMIT2_CH_LO_LIMIT_EN_POS))   /**< CH_LO_LIMIT_EN Mask        */
#define MXC_F_ADC_LIMIT2_CH_HI_LIMIT_EN_POS                 29                                                                  /**< CH_HI_LIMIT_EN Position    */    
#define MXC_F_ADC_LIMIT2_CH_HI_LIMIT_EN                     ((uint32_t)(0x00000001UL << MXC_F_ADC_LIMIT2_CH_HI_LIMIT_EN_POS))   /**< CH_HI_LIMIT_EN Mask        */
/** @} end of group ADC_LIMIT2_register */

/**
 * @ingroup  adc_registers
 * @defgroup ADC_LIMIT3_Register ADC_LIMIT3
 * @brief     Field Positions and Bit Masks for the ADC_LIMIT3 register
 * @{
 */
#define MXC_F_ADC_LIMIT3_CH_LO_LIMIT_POS                    0                                                                   /**< CH_LO_LIMIT Position       */    
#define MXC_F_ADC_LIMIT3_CH_LO_LIMIT                        ((uint32_t)(0x000003FFUL << MXC_F_ADC_LIMIT3_CH_LO_LIMIT_POS))      /**< CH_LO_LIMIT Mask           */
#define MXC_F_ADC_LIMIT3_CH_HI_LIMIT_POS                    12                                                                  /**< CH_HI_LIMIT Position       */    
#define MXC_F_ADC_LIMIT3_CH_HI_LIMIT                        ((uint32_t)(0x000003FFUL << MXC_F_ADC_LIMIT3_CH_HI_LIMIT_POS))      /**< CH_HI_LIMIT Mask           */
#define MXC_F_ADC_LIMIT3_CH_SEL_POS                         24                                                                  /**< CH_SEL Position            */     
#define MXC_F_ADC_LIMIT3_CH_SEL                             ((uint32_t)(0x0000000FUL << MXC_F_ADC_LIMIT3_CH_SEL_POS))           /**< CH_SEL Mask                */
#define MXC_F_ADC_LIMIT3_CH_LO_LIMIT_EN_POS                 28                                                                  /**< CH_LO_LIMIT_EN Position    */     
#define MXC_F_ADC_LIMIT3_CH_LO_LIMIT_EN                     ((uint32_t)(0x00000001UL << MXC_F_ADC_LIMIT3_CH_LO_LIMIT_EN_POS))   /**< CH_LO_LIMIT_EN Mask        */
#define MXC_F_ADC_LIMIT3_CH_HI_LIMIT_EN_POS                 29                                                                  /**< CH_HI_LIMIT_EN Position    */     
#define MXC_F_ADC_LIMIT3_CH_HI_LIMIT_EN                     ((uint32_t)(0x00000001UL << MXC_F_ADC_LIMIT3_CH_HI_LIMIT_EN_POS))   /**< CH_HI_LIMIT_EN Mask        */
/** @} end of group ADC_LIMIT3_register */

/**
 * @ingroup  adc_registers
 * @defgroup ADC_AFE_CTRL_Register ADC_AFE_CTRL
 * @brief      Field Positions and Bit Masks for the ADC_AFE_CTRL register
 * @{
 */
#define MXC_F_ADC_AFE_CTRL_TMON_INTBIAS_EN_POS              8                                                                     /**< TMON_INTBIAS_EN Position */  
#define MXC_F_ADC_AFE_CTRL_TMON_INTBIAS_EN                  ((uint32_t)(0x00000001UL << MXC_F_ADC_AFE_CTRL_TMON_INTBIAS_EN_POS))  /**< TMON_INTBIAS_EN Mask     */  
#define MXC_F_ADC_AFE_CTRL_TMON_EXTBIAS_EN_POS              9                                                                     /**< TMON_EXTBIAS_EN Position */  
#define MXC_F_ADC_AFE_CTRL_TMON_EXTBIAS_EN                  ((uint32_t)(0x00000001UL << MXC_F_ADC_AFE_CTRL_TMON_EXTBIAS_EN_POS))  /**< TMON_EXTBIAS_EN Mask     */  
/** @} end of group ADC_AFE_CTRL_register */

/**
 * @ingroup  adc_registers
 * @defgroup ADC_RO_CAL0_Register ADC_RO_CAL0
 * @brief     Field Positions and Bit Masks for the ADC_RO_CAL0 register
 * @{
 */
#define MXC_F_ADC_RO_CAL0_RO_CAL_EN_POS                     0                                                                       /**< RO_CAL_EN Position     */  
#define MXC_F_ADC_RO_CAL0_RO_CAL_EN                         ((uint32_t)(0x00000001UL << MXC_F_ADC_RO_CAL0_RO_CAL_EN_POS))           /**< RO_CAL_EN Mask         */   
#define MXC_F_ADC_RO_CAL0_RO_CAL_RUN_POS                    1                                                                       /**< RO_CAL_RUN Position    */  
#define MXC_F_ADC_RO_CAL0_RO_CAL_RUN                        ((uint32_t)(0x00000001UL << MXC_F_ADC_RO_CAL0_RO_CAL_RUN_POS))          /**< RO_CAL_RUN Mask        */   
#define MXC_F_ADC_RO_CAL0_RO_CAL_LOAD_POS                   2                                                                       /**< RO_CAL_LOAD Position   */  
#define MXC_F_ADC_RO_CAL0_RO_CAL_LOAD                       ((uint32_t)(0x00000001UL << MXC_F_ADC_RO_CAL0_RO_CAL_LOAD_POS))         /**< RO_CAL_LOAD Mask       */  
#define MXC_F_ADC_RO_CAL0_RO_CAL_ATOMIC_POS                 4                                                                       /**< RO_CAL_ATOMIC Position */  
#define MXC_F_ADC_RO_CAL0_RO_CAL_ATOMIC                     ((uint32_t)(0x00000001UL << MXC_F_ADC_RO_CAL0_RO_CAL_ATOMIC_POS))       /**< RO_CAL_ATOMIC Mask     */  
#define MXC_F_ADC_RO_CAL0_DUMMY_POS                         5                                                                       /**< DUMMY Position         */   
#define MXC_F_ADC_RO_CAL0_DUMMY                             ((uint32_t)(0x00000007UL << MXC_F_ADC_RO_CAL0_DUMMY_POS))               /**< DUMMY Mask             */   
#define MXC_F_ADC_RO_CAL0_TRM_MU_POS                        8                                                                       /**< TRM_MU Position        */   
#define MXC_F_ADC_RO_CAL0_TRM_MU                            ((uint32_t)(0x00000FFFUL << MXC_F_ADC_RO_CAL0_TRM_MU_POS))              /**< TRM_MU Mask            */   
#define MXC_F_ADC_RO_CAL0_RO_TRM_POS                        23                                                                      /**< RO_TRM Position        */   
#define MXC_F_ADC_RO_CAL0_RO_TRM                            ((uint32_t)(0x000001FFUL << MXC_F_ADC_RO_CAL0_RO_TRM_POS))              /**< RO_TRM Mask            */   
/** @} end of group ADC_RO_CAL0_register */

/**
 * @ingroup  adc_registers
 * @defgroup ADC_RO_CAL1_Register ADC_RO_CAL1
 * @brief     Field Positions and Bit Masks for the ADC_RO_CAL1 register
 * @{
 */
#define MXC_F_ADC_RO_CAL1_TRM_INIT_POS                      0                                                                    /**< TRM_INIT Position         */   
#define MXC_F_ADC_RO_CAL1_TRM_INIT                          ((uint32_t)(0x000001FFUL << MXC_F_ADC_RO_CAL1_TRM_INIT_POS))         /**< TRM_INIT Mask             */   
#define MXC_F_ADC_RO_CAL1_TRM_MIN_POS                       10                                                                   /**< TRM_MIN Position          */   
#define MXC_F_ADC_RO_CAL1_TRM_MIN                           ((uint32_t)(0x000001FFUL << MXC_F_ADC_RO_CAL1_TRM_MIN_POS))          /**< TRM_MIN Mask              */
#define MXC_F_ADC_RO_CAL1_TRM_MAX_POS                       20                                                                   /**< TRM_MAX Position          */    
#define MXC_F_ADC_RO_CAL1_TRM_MAX                           ((uint32_t)(0x000001FFUL << MXC_F_ADC_RO_CAL1_TRM_MAX_POS))          /**< TRM_MAX Mask              */
/** @} end of group RO_CAL1_register */

/**
 * @ingroup  adc_registers
 * @defgroup ADC_RO_CAL2_Register ADC_RO_CAL2
 * @brief     Field Positions and Bit Masks for the ADC_RO_CAL2 register
 * @{
 */
#define MXC_F_ADC_RO_CAL2_AUTO_CAL_DONE_CNT_POS             0                                                                       /**< AUTO_CAL_DONE_CNT Position */
#define MXC_F_ADC_RO_CAL2_AUTO_CAL_DONE_CNT                 ((uint32_t)(0x000000FFUL << MXC_F_ADC_RO_CAL2_AUTO_CAL_DONE_CNT_POS))   /**< AUTO_CAL_DONE_CNT Mask     */    
/** @} end of group RO_CAL2_register */

/**
 * @ingroup  ADC_CTRL_Register
 * @defgroup ADC_CHSEL_values ADC Channel Select Values
 * @brief    Channel Select Values
 * @{
 */
#define MXC_V_ADC_CTRL_ADC_CHSEL_AIN0                       ((uint32_t)(0x00000000UL))  /**< Channel 0 Select       */
#define MXC_V_ADC_CTRL_ADC_CHSEL_AIN1                       ((uint32_t)(0x00000001UL))  /**< Channel 1 Select       */
#define MXC_V_ADC_CTRL_ADC_CHSEL_AIN2                       ((uint32_t)(0x00000002UL))  /**< Channel 2 Select       */
#define MXC_V_ADC_CTRL_ADC_CHSEL_AIN3                       ((uint32_t)(0x00000003UL))  /**< Channel 3 Select       */
#define MXC_V_ADC_CTRL_ADC_CHSEL_AIN0_DIV_5                 ((uint32_t)(0x00000004UL))  /**< Channel 0 divided by 5 */
#define MXC_V_ADC_CTRL_ADC_CHSEL_AIN1_DIV_5                 ((uint32_t)(0x00000005UL))  /**< Channel 1 divided by 5 */
#define MXC_V_ADC_CTRL_ADC_CHSEL_VDDB_DIV_4                 ((uint32_t)(0x00000006UL))  /**< VDDB divided by 4      */
#define MXC_V_ADC_CTRL_ADC_CHSEL_VDD18                      ((uint32_t)(0x00000007UL))  /**< VDD18 input select     */
#define MXC_V_ADC_CTRL_ADC_CHSEL_VDD12                      ((uint32_t)(0x00000008UL))  /**< VDD12 input select     */
#define MXC_V_ADC_CTRL_ADC_CHSEL_VRTC_DIV_2                 ((uint32_t)(0x00000009UL))  /**< VRTC divided by 2      */
#define MXC_V_ADC_CTRL_ADC_CHSEL_TMON                       ((uint32_t)(0x0000000AUL))  /**< TMON input select      */

#if(MXC_ADC_REV > 0)
#define MXC_V_ADC_CTRL_ADC_CHSEL_VDDIO_DIV_4                ((uint32_t)(0x0000000BUL)) /**< VDDIO divided by 4 select   */
#define MXC_V_ADC_CTRL_ADC_CHSEL_VDDIOH_DIV_4               ((uint32_t)(0x0000000CUL)) /**< VDDIOH divided by 4 select  */
#endif
/** @} end of group ADC_CHSEL_values */

#ifdef __cplusplus
}
#endif

#endif   /* _MXC_ADC_REGS_H_ */
