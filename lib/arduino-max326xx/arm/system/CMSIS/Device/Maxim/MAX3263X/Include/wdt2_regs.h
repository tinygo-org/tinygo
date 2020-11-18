/**
 * @file    
 * @brief   Registers, Bit Masks and Bit Positions for the WDT2 Peripheral Module.
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
 * $Date: 2017-02-16 12:02:20 -0600 (Thu, 16 Feb 2017) $
 * $Revision: 26462 $
 *
 *************************************************************************** */

/* **** Includes **** */
#include <stdint.h>

/* Define to prevent redundant inclusion */
#ifndef _MXC_WDT2_REGS_H_
#define _MXC_WDT2_REGS_H_

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
 * @addtogroup     wdt2
 * @{
 * @defgroup    wdt2_registers WDT2 Registers
 * @brief       Registers, Bit Masks and Bit Positions
 * @{
 */
/*
   Typedefed structure(s) for module registers (per instance or section) with direct 32-bit
   access to each register in module.
*/
/**
 * Structure type to access the WDT2 Registers, see #MXC_WDT2 to get a pointer to the WDT2 register structure.
 * @note This is an always-on watchdog timer, it operates in all modes of operation. 
 */
typedef struct {
    __IO uint32_t ctrl;                                 /**< WDT2_CTRL Register - WDT Control Register                                   */
    __IO uint32_t clear;                                /**< WDT2_CLEAR Register - WDT Clear Register to prevent a WDT Reset (Feed Dog)  */
    __IO uint32_t flags;                                /**< WDT2_FLAGS Register - WDT Interrupt and Reset Flags                         */
    __IO uint32_t enable;                               /**< WDT2_ENABLE Register - WDT Reset and Interrupt Enable/Disable Controls      */
    __R  uint32_t rsv010;                               /**< <em><b>RESERVED, DO NOT MODIFY</b></em>.                                                   */
    __IO uint32_t lock_ctrl;                            /**< WDT2_LOCK_CTRL Register - Lock for Control Register                         */
} mxc_wdt2_regs_t;
/**@} end of group wdt2_registers.*/


/*
   Register offsets for module WDT2.
*/
/**
 * @ingroup    wdt2_registers
 * @defgroup   WDT2_Register_Offsets Register Offsets
 * @brief      Watchdog Timer 2 Register Offsets from the WDT2 Base Peripheral Address. 
 * @details    Use #MXC_WDT2 for the WDT2 Base Peripheral Address. 
 * @{
 */
#define MXC_R_WDT2_OFFS_CTRL                                ((uint32_t)0x00000000UL)    /**< WDT2_CTRL Offset: <tt>0x0000</tt> */
#define MXC_R_WDT2_OFFS_CLEAR                               ((uint32_t)0x00000004UL)    /**< WDT2_CLEAR Offset: <tt>0x0004</tt> */
#define MXC_R_WDT2_OFFS_FLAGS                               ((uint32_t)0x00000008UL)    /**< WDT2_FLAGS Offset: <tt>0x0008</tt> */
#define MXC_R_WDT2_OFFS_ENABLE                              ((uint32_t)0x0000000CUL)    /**< WDT2_ENABLE Offset: <tt>0x000C</tt> */
#define MXC_R_WDT2_OFFS_LOCK_CTRL                           ((uint32_t)0x00000014UL)    /**< WDT2_LOCK_CTRL Offset: <tt>0x0014</tt> */
/**@} end of group WDT2_Register_Offsets */


/*
   Field positions and masks for module WDT2.
*/
/**
 * @ingroup    wdt2_registers
 * @defgroup   WDT2_CTRL_Register WDT2_CTRL Register
 * @brief      Field Positions and Bit Masks for the WDT2_CTRL register 
 * @{
 */
#define MXC_F_WDT2_CTRL_INT_PERIOD_POS                      0                                                               /**< INT_PERIOD Field Position */
#define MXC_F_WDT2_CTRL_INT_PERIOD                          ((uint32_t)(0x0000000FUL << MXC_F_WDT2_CTRL_INT_PERIOD_POS))    /**< INT_PERIOD Field Mask - This field is used to set the interrupt period on the WDT. */
#define MXC_F_WDT2_CTRL_RST_PERIOD_POS                      4                                                               /**< RST_PERIOD Field Position */
#define MXC_F_WDT2_CTRL_RST_PERIOD                          ((uint32_t)(0x0000000FUL << MXC_F_WDT2_CTRL_RST_PERIOD_POS))    /**< RST_PERIOD Field Mask - This field sets the time after an
                                                                                                                              * interrupt period has expired before the device resets. If the
                                                                                                                              * INT_PERIOD Flag is cleared prior to the RST_PERIOD expiration,
                                                                                                                              * the device will not reset. */
#define MXC_F_WDT2_CTRL_EN_TIMER_POS                        8                                                               /**< EN_TIMER Field Position    */
#define MXC_F_WDT2_CTRL_EN_TIMER                            ((uint32_t)(0x00000001UL << MXC_F_WDT2_CTRL_EN_TIMER_POS))      /**< EN_TIMER Field Mask        */
#define MXC_F_WDT2_CTRL_EN_CLOCK_POS                        9                                                               /**< EN_CLOCK Field Position    */
#define MXC_F_WDT2_CTRL_EN_CLOCK                            ((uint32_t)(0x00000001UL << MXC_F_WDT2_CTRL_EN_CLOCK_POS))      /**< EN_CLOCK Field Mask        */
#define MXC_F_WDT2_CTRL_EN_TIMER_SLP_POS                    10                                                              /**< WAIT_PERIOD Field Position */
#define MXC_F_WDT2_CTRL_EN_TIMER_SLP                        ((uint32_t)(0x00000001UL << MXC_F_WDT2_CTRL_EN_TIMER_SLP_POS))  /**< WAIT_PERIOD Field Mask     */
/**@} end of group WDT2_CTRL */
/**
 * @ingroup  wdt2_registers
 * @defgroup WDT2_FLAGS_Register WDT2_FLAGS Register
 * @brief    Field Positions and Bit Masks for the WDT2_FLAGS register. Watchdog Timer 2 Flags for Interrupts and Reset.
 * @{
 */  
#define MXC_F_WDT2_FLAGS_TIMEOUT_POS                        0                                                               /**< TIMEOUT Flag Position */
#define MXC_F_WDT2_FLAGS_TIMEOUT                            ((uint32_t)(0x00000001UL << MXC_F_WDT2_FLAGS_TIMEOUT_POS))      /**< TIMEOUT Flag Mask - if this flag is set it indicates the Watchdog Timer 2 timed out. */
#define MXC_F_WDT2_FLAGS_RESET_OUT_POS                      2                                                               /**< RESET_OUT Flag Position */
#define MXC_F_WDT2_FLAGS_RESET_OUT                          ((uint32_t)(0x00000001UL << MXC_F_WDT2_FLAGS_RESET_OUT_POS))    /**< RESET_FLAG Flag Mask - This flag indicates that the watchdog timer timed out and the reset period elapsed without the timer being cleared. This will result in a system restart. */
/**@} end of group WDT2_FLAGS */

/**
 * @ingroup  wdt2_registers
 * @defgroup WDT2_ENABLE_Register WDT2_ENABLE Register
 * @brief    Field Positions and Bit Masks for the WDT2_ENABLE register. 
 * @{
 */
#define MXC_F_WDT2_ENABLE_TIMEOUT_POS                       0                                                               /**< ENABLE_TIMEOUT Field Position */
#define MXC_F_WDT2_ENABLE_TIMEOUT                           ((uint32_t)(0x00000001UL << MXC_F_WDT2_ENABLE_TIMEOUT_POS))     /**< ENABLE_TIMEOUT Field Mask */
#define MXC_F_WDT2_ENABLE_RESET_OUT_POS                     2                                                               /**< ENABLE_RESET_OUT Field Position */
#define MXC_F_WDT2_ENABLE_RESET_OUT                         ((uint32_t)(0x00000001UL << MXC_F_WDT2_ENABLE_RESET_OUT_POS))   /**< ENABLE_RESET_OUT Field Mask */
/**@} end of group WDT2_ENABLE */

/**
 * @ingroup  wdt2_registers
 * @defgroup WDT2_LOCK_CTRL_Register WDT2_LOCK_CTRL Register
 * @brief    The WDT2_LOCK_CTRL register controls read/write access to the \ref WDT2_CTRL_Register.
 * @{
 */
#define MXC_F_WDT2_LOCK_CTRL_WDLOCK_POS                     0                                                               /**< WDLOCK Field's position in the WDT2_LOCK_CTRL register. */ 
#define MXC_F_WDT2_LOCK_CTRL_WDLOCK                         ((uint32_t)(0x000000FFUL << MXC_F_WDT2_LOCK_CTRL_WDLOCK_POS))   /**< WDLOCK Field mask for the WDT2_LOCK_CTRL register.  Reading a value of */
/**@} end of group WDT2_ENABLE */



/*
   Field values and shifted values for module WDT2.
*/
/**
 * @ingroup WDT2_CTRL_Register 
 * @defgroup WDT2_CTRL_field_values WDT2_CTRL Register Field and Shifted Field Values
 * @brief    Field values and Shifted Field values for the WDT2_CTRL register. 
 * @details  Shifted field values are field values shifted to the loacation of the field in the register.  
 */
/**
 * @ingroup    WDT2_CTRL_field_values
 * @defgroup   WDT2_CTRL_INT_PERIOD_Value Watchdog Timer Interrupt Period
 * @brief      Sets the duration of the watchdog interrupt period.
 * @details    The INT_PERIOD field sets the duration of the watchdog interrupt
 *             period, which is the time period from the WDT2 being
 *             enabled/cleared until the WDT2 flag, #MXC_F_WDT2_FLAGS_TIMEOUT, is
 *             set.
 *             The values defined are in the number of watchdog clock cycles.
 * @{
 */
#define MXC_V_WDT2_CTRL_INT_PERIOD_2_25_NANO_CLKS                               ((uint32_t)(0x00000000UL))  /**< Interupt Period of \f$ 2^{25} \f$ WDT2 CLK Cycles */
#define MXC_V_WDT2_CTRL_INT_PERIOD_2_24_NANO_CLKS                               ((uint32_t)(0x00000001UL))  /**< Interupt Period of \f$ 2^{24} \f$ WDT2 CLK Cycles */
#define MXC_V_WDT2_CTRL_INT_PERIOD_2_23_NANO_CLKS                               ((uint32_t)(0x00000002UL))  /**< Interupt Period of \f$ 2^{23} \f$ WDT2 CLK Cycles */
#define MXC_V_WDT2_CTRL_INT_PERIOD_2_22_NANO_CLKS                               ((uint32_t)(0x00000003UL))  /**< Interupt Period of \f$ 2^{22} \f$ WDT2 CLK Cycles */
#define MXC_V_WDT2_CTRL_INT_PERIOD_2_21_NANO_CLKS                               ((uint32_t)(0x00000004UL))  /**< Interupt Period of \f$ 2^{21} \f$ WDT2 CLK Cycles */
#define MXC_V_WDT2_CTRL_INT_PERIOD_2_20_NANO_CLKS                               ((uint32_t)(0x00000005UL))  /**< Interupt Period of \f$ 2^{20} \f$ WDT2 CLK Cycles */
#define MXC_V_WDT2_CTRL_INT_PERIOD_2_19_NANO_CLKS                               ((uint32_t)(0x00000006UL))  /**< Interupt Period of \f$ 2^{19} \f$ WDT2 CLK Cycles */
#define MXC_V_WDT2_CTRL_INT_PERIOD_2_18_NANO_CLKS                               ((uint32_t)(0x00000007UL))  /**< Interupt Period of \f$ 2^{18} \f$ WDT2 CLK Cycles */
#define MXC_V_WDT2_CTRL_INT_PERIOD_2_17_NANO_CLKS                               ((uint32_t)(0x00000008UL))  /**< Interupt Period of \f$ 2^{17} \f$ WDT2 CLK Cycles */
#define MXC_V_WDT2_CTRL_INT_PERIOD_2_16_NANO_CLKS                               ((uint32_t)(0x00000009UL))  /**< Interupt Period of \f$ 2^{16} \f$ WDT2 CLK Cycles */
#define MXC_V_WDT2_CTRL_INT_PERIOD_2_15_NANO_CLKS                               ((uint32_t)(0x0000000AUL))  /**< Interupt Period of \f$ 2^{15} \f$ WDT2 CLK Cycles */
#define MXC_V_WDT2_CTRL_INT_PERIOD_2_14_NANO_CLKS                               ((uint32_t)(0x0000000BUL))  /**< Interupt Period of \f$ 2^{14} \f$ WDT2 CLK Cycles */
#define MXC_V_WDT2_CTRL_INT_PERIOD_2_13_NANO_CLKS                               ((uint32_t)(0x0000000CUL))  /**< Interupt Period of \f$ 2^{13} \f$ WDT2 CLK Cycles */
#define MXC_V_WDT2_CTRL_INT_PERIOD_2_12_NANO_CLKS                               ((uint32_t)(0x0000000DUL))  /**< Interupt Period of \f$ 2^{12} \f$ WDT2 CLK Cycles */
#define MXC_V_WDT2_CTRL_INT_PERIOD_2_11_NANO_CLKS                               ((uint32_t)(0x0000000EUL))  /**< Interupt Period of \f$ 2^{11} \f$ WDT2 CLK Cycles */
#define MXC_V_WDT2_CTRL_INT_PERIOD_2_10_NANO_CLKS                               ((uint32_t)(0x0000000FUL))  /**< Interupt Period of \f$ 2^{10} \f$ WDT2 CLK Cycles */
/**@} end of group WDT2_CTRL_INT_PERIOD_Value */

/**
 * @ingroup    WDT2_CTRL_field_values
 * @defgroup   WDT2_CTRL_INT_PERIOD_Shifted Watchdog Timer Interrupt Period Shifted Values
 * @brief      Shifted values for the \ref WDT2_CTRL_INT_PERIOD_Value
 * @details    The shifted value is
 *             shifted to align with the fields location in the WDT2_CTRL register. 
 * @{
 */
#define MXC_S_WDT2_CTRL_INT_PERIOD_2_25_NANO_CLKS                               ((uint32_t)(MXC_V_WDT2_CTRL_INT_PERIOD_2_25_NANO_CLKS   << MXC_F_WDT2_CTRL_INT_PERIOD_POS))  /**< Shifted Field Value for #MXC_V_WDT2_CTRL_INT_PERIOD_2_25_NANO_CLKS */
#define MXC_S_WDT2_CTRL_INT_PERIOD_2_24_NANO_CLKS                               ((uint32_t)(MXC_V_WDT2_CTRL_INT_PERIOD_2_24_NANO_CLKS   << MXC_F_WDT2_CTRL_INT_PERIOD_POS))  /**< Shifted Field Value for #MXC_V_WDT2_CTRL_INT_PERIOD_2_24_NANO_CLKS */
#define MXC_S_WDT2_CTRL_INT_PERIOD_2_23_NANO_CLKS                               ((uint32_t)(MXC_V_WDT2_CTRL_INT_PERIOD_2_23_NANO_CLKS   << MXC_F_WDT2_CTRL_INT_PERIOD_POS))  /**< Shifted Field Value for #MXC_V_WDT2_CTRL_INT_PERIOD_2_23_NANO_CLKS */
#define MXC_S_WDT2_CTRL_INT_PERIOD_2_22_NANO_CLKS                               ((uint32_t)(MXC_V_WDT2_CTRL_INT_PERIOD_2_22_NANO_CLKS   << MXC_F_WDT2_CTRL_INT_PERIOD_POS))  /**< Shifted Field Value for #MXC_V_WDT2_CTRL_INT_PERIOD_2_22_NANO_CLKS */
#define MXC_S_WDT2_CTRL_INT_PERIOD_2_21_NANO_CLKS                               ((uint32_t)(MXC_V_WDT2_CTRL_INT_PERIOD_2_21_NANO_CLKS   << MXC_F_WDT2_CTRL_INT_PERIOD_POS))  /**< Shifted Field Value for #MXC_V_WDT2_CTRL_INT_PERIOD_2_21_NANO_CLKS */
#define MXC_S_WDT2_CTRL_INT_PERIOD_2_20_NANO_CLKS                               ((uint32_t)(MXC_V_WDT2_CTRL_INT_PERIOD_2_20_NANO_CLKS   << MXC_F_WDT2_CTRL_INT_PERIOD_POS))  /**< Shifted Field Value for #MXC_V_WDT2_CTRL_INT_PERIOD_2_20_NANO_CLKS */
#define MXC_S_WDT2_CTRL_INT_PERIOD_2_19_NANO_CLKS                               ((uint32_t)(MXC_V_WDT2_CTRL_INT_PERIOD_2_19_NANO_CLKS   << MXC_F_WDT2_CTRL_INT_PERIOD_POS))  /**< Shifted Field Value for #MXC_V_WDT2_CTRL_INT_PERIOD_2_19_NANO_CLKS */
#define MXC_S_WDT2_CTRL_INT_PERIOD_2_18_NANO_CLKS                               ((uint32_t)(MXC_V_WDT2_CTRL_INT_PERIOD_2_18_NANO_CLKS   << MXC_F_WDT2_CTRL_INT_PERIOD_POS))  /**< Shifted Field Value for #MXC_V_WDT2_CTRL_INT_PERIOD_2_18_NANO_CLKS */
#define MXC_S_WDT2_CTRL_INT_PERIOD_2_17_NANO_CLKS                               ((uint32_t)(MXC_V_WDT2_CTRL_INT_PERIOD_2_17_NANO_CLKS   << MXC_F_WDT2_CTRL_INT_PERIOD_POS))  /**< Shifted Field Value for #MXC_V_WDT2_CTRL_INT_PERIOD_2_17_NANO_CLKS */
#define MXC_S_WDT2_CTRL_INT_PERIOD_2_16_NANO_CLKS                               ((uint32_t)(MXC_V_WDT2_CTRL_INT_PERIOD_2_16_NANO_CLKS   << MXC_F_WDT2_CTRL_INT_PERIOD_POS))  /**< Shifted Field Value for #MXC_V_WDT2_CTRL_INT_PERIOD_2_16_NANO_CLKS */
#define MXC_S_WDT2_CTRL_INT_PERIOD_2_15_NANO_CLKS                               ((uint32_t)(MXC_V_WDT2_CTRL_INT_PERIOD_2_15_NANO_CLKS   << MXC_F_WDT2_CTRL_INT_PERIOD_POS))  /**< Shifted Field Value for #MXC_V_WDT2_CTRL_INT_PERIOD_2_15_NANO_CLKS */
#define MXC_S_WDT2_CTRL_INT_PERIOD_2_14_NANO_CLKS                               ((uint32_t)(MXC_V_WDT2_CTRL_INT_PERIOD_2_14_NANO_CLKS   << MXC_F_WDT2_CTRL_INT_PERIOD_POS))  /**< Shifted Field Value for #MXC_V_WDT2_CTRL_INT_PERIOD_2_14_NANO_CLKS */
#define MXC_S_WDT2_CTRL_INT_PERIOD_2_13_NANO_CLKS                               ((uint32_t)(MXC_V_WDT2_CTRL_INT_PERIOD_2_13_NANO_CLKS   << MXC_F_WDT2_CTRL_INT_PERIOD_POS))  /**< Shifted Field Value for #MXC_V_WDT2_CTRL_INT_PERIOD_2_13_NANO_CLKS */
#define MXC_S_WDT2_CTRL_INT_PERIOD_2_12_NANO_CLKS                               ((uint32_t)(MXC_V_WDT2_CTRL_INT_PERIOD_2_12_NANO_CLKS   << MXC_F_WDT2_CTRL_INT_PERIOD_POS))  /**< Shifted Field Value for #MXC_V_WDT2_CTRL_INT_PERIOD_2_12_NANO_CLKS */
#define MXC_S_WDT2_CTRL_INT_PERIOD_2_11_NANO_CLKS                               ((uint32_t)(MXC_V_WDT2_CTRL_INT_PERIOD_2_11_NANO_CLKS   << MXC_F_WDT2_CTRL_INT_PERIOD_POS))  /**< Shifted Field Value for #MXC_V_WDT2_CTRL_INT_PERIOD_2_11_NANO_CLKS */
#define MXC_S_WDT2_CTRL_INT_PERIOD_2_10_NANO_CLKS                               ((uint32_t)(MXC_V_WDT2_CTRL_INT_PERIOD_2_10_NANO_CLKS   << MXC_F_WDT2_CTRL_INT_PERIOD_POS))  /**< Shifted Field Value for #MXC_V_WDT2_CTRL_INT_PERIOD_2_10_NANO_CLKS */
/**@} end of group WDT2_CTRL_INT_PERIOD_Shifted */
/**
 * @ingroup    WDT2_CTRL_field_values
 * @defgroup   WDT2_CTRL_RST_PERIOD_Value Watchdog Timer Reset Period
 * @brief      Sets the duration of the watchdog reset period.
 * @details    The RST_PERIOD field sets the duration of the watchdog reset
 *             period, which is the time period from the WDT being
 *             enabled/cleared until the WDT2 flag, #MXC_F_WDT2_CTRL_RST_PERIOD is
 *             set.
 *             The values defined are in the number of watchdog clock cycles.
 * @{
 */
#define MXC_V_WDT2_CTRL_RST_PERIOD_2_25_NANO_CLKS                               ((uint32_t)(0x00000000UL))        /**< Reset Period of \f$ 2^{25} \f$ WDT2 CLK CYCLES  */
#define MXC_V_WDT2_CTRL_RST_PERIOD_2_24_NANO_CLKS                               ((uint32_t)(0x00000001UL))        /**< Reset Period of \f$ 2^{24} \f$ WDT2 CLK CYCLES  */
#define MXC_V_WDT2_CTRL_RST_PERIOD_2_23_NANO_CLKS                               ((uint32_t)(0x00000002UL))        /**< Reset Period of \f$ 2^{23} \f$ WDT2 CLK CYCLES  */
#define MXC_V_WDT2_CTRL_RST_PERIOD_2_22_NANO_CLKS                               ((uint32_t)(0x00000003UL))        /**< Reset Period of \f$ 2^{22} \f$ WDT2 CLK CYCLES  */
#define MXC_V_WDT2_CTRL_RST_PERIOD_2_21_NANO_CLKS                               ((uint32_t)(0x00000004UL))        /**< Reset Period of \f$ 2^{21} \f$ WDT2 CLK CYCLES  */
#define MXC_V_WDT2_CTRL_RST_PERIOD_2_20_NANO_CLKS                               ((uint32_t)(0x00000005UL))        /**< Reset Period of \f$ 2^{20} \f$ WDT2 CLK CYCLES  */
#define MXC_V_WDT2_CTRL_RST_PERIOD_2_19_NANO_CLKS                               ((uint32_t)(0x00000006UL))        /**< Reset Period of \f$ 2^{19} \f$ WDT2 CLK CYCLES  */
#define MXC_V_WDT2_CTRL_RST_PERIOD_2_18_NANO_CLKS                               ((uint32_t)(0x00000007UL))        /**< Reset Period of \f$ 2^{18} \f$ WDT2 CLK CYCLES  */
#define MXC_V_WDT2_CTRL_RST_PERIOD_2_17_NANO_CLKS                               ((uint32_t)(0x00000008UL))        /**< Reset Period of \f$ 2^{17} \f$ WDT2 CLK CYCLES  */
#define MXC_V_WDT2_CTRL_RST_PERIOD_2_16_NANO_CLKS                               ((uint32_t)(0x00000009UL))        /**< Reset Period of \f$ 2^{16} \f$ WDT2 CLK CYCLES  */
#define MXC_V_WDT2_CTRL_RST_PERIOD_2_15_NANO_CLKS                               ((uint32_t)(0x0000000AUL))        /**< Reset Period of \f$ 2^{15} \f$ WDT2 CLK CYCLES  */
#define MXC_V_WDT2_CTRL_RST_PERIOD_2_14_NANO_CLKS                               ((uint32_t)(0x0000000BUL))        /**< Reset Period of \f$ 2^{14} \f$ WDT2 CLK CYCLES  */
#define MXC_V_WDT2_CTRL_RST_PERIOD_2_13_NANO_CLKS                               ((uint32_t)(0x0000000CUL))        /**< Reset Period of \f$ 2^{13} \f$ WDT2 CLK CYCLES  */
#define MXC_V_WDT2_CTRL_RST_PERIOD_2_12_NANO_CLKS                               ((uint32_t)(0x0000000DUL))        /**< Reset Period of \f$ 2^{12} \f$ WDT2 CLK CYCLES  */
#define MXC_V_WDT2_CTRL_RST_PERIOD_2_11_NANO_CLKS                               ((uint32_t)(0x0000000EUL))        /**< Reset Period of \f$ 2^{11} \f$ WDT2 CLK CYCLES  */
#define MXC_V_WDT2_CTRL_RST_PERIOD_2_10_NANO_CLKS                               ((uint32_t)(0x0000000FUL))        /**< Reset Period of \f$ 2^{10} \f$ WDT2 CLK CYCLES  */
/**@} end of group WDT2_CTRL_RST_PERIOD_Value */

/**
 * @ingroup    WDT2_CTRL_field_values
 * @defgroup   WDT2_CTRL_RST_PERIOD_Shifted Watchdog Timer Reset Period Shifted Values
 * @brief      Shifted values for the \ref WDT2_CTRL_RST_PERIOD_Value
 * @details    These values are shifted to align with the field's location in the WDT2_CTRL register. 
 * @{
 */
#define MXC_S_WDT2_CTRL_RST_PERIOD_2_25_NANO_CLKS                               ((uint32_t)(MXC_V_WDT2_CTRL_RST_PERIOD_2_25_NANO_CLKS   << MXC_F_WDT2_CTRL_RST_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT2_CTRL_RST_PERIOD_2_25_NANO_CLKS */
#define MXC_S_WDT2_CTRL_RST_PERIOD_2_24_NANO_CLKS                               ((uint32_t)(MXC_V_WDT2_CTRL_RST_PERIOD_2_24_NANO_CLKS   << MXC_F_WDT2_CTRL_RST_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT2_CTRL_RST_PERIOD_2_24_NANO_CLKS */
#define MXC_S_WDT2_CTRL_RST_PERIOD_2_23_NANO_CLKS                               ((uint32_t)(MXC_V_WDT2_CTRL_RST_PERIOD_2_23_NANO_CLKS   << MXC_F_WDT2_CTRL_RST_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT2_CTRL_RST_PERIOD_2_23_NANO_CLKS */
#define MXC_S_WDT2_CTRL_RST_PERIOD_2_22_NANO_CLKS                               ((uint32_t)(MXC_V_WDT2_CTRL_RST_PERIOD_2_22_NANO_CLKS   << MXC_F_WDT2_CTRL_RST_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT2_CTRL_RST_PERIOD_2_22_NANO_CLKS */
#define MXC_S_WDT2_CTRL_RST_PERIOD_2_21_NANO_CLKS                               ((uint32_t)(MXC_V_WDT2_CTRL_RST_PERIOD_2_21_NANO_CLKS   << MXC_F_WDT2_CTRL_RST_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT2_CTRL_RST_PERIOD_2_21_NANO_CLKS */
#define MXC_S_WDT2_CTRL_RST_PERIOD_2_20_NANO_CLKS                               ((uint32_t)(MXC_V_WDT2_CTRL_RST_PERIOD_2_20_NANO_CLKS   << MXC_F_WDT2_CTRL_RST_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT2_CTRL_RST_PERIOD_2_20_NANO_CLKS */
#define MXC_S_WDT2_CTRL_RST_PERIOD_2_19_NANO_CLKS                               ((uint32_t)(MXC_V_WDT2_CTRL_RST_PERIOD_2_19_NANO_CLKS   << MXC_F_WDT2_CTRL_RST_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT2_CTRL_RST_PERIOD_2_19_NANO_CLKS */
#define MXC_S_WDT2_CTRL_RST_PERIOD_2_18_NANO_CLKS                               ((uint32_t)(MXC_V_WDT2_CTRL_RST_PERIOD_2_18_NANO_CLKS   << MXC_F_WDT2_CTRL_RST_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT2_CTRL_RST_PERIOD_2_18_NANO_CLKS */
#define MXC_S_WDT2_CTRL_RST_PERIOD_2_17_NANO_CLKS                               ((uint32_t)(MXC_V_WDT2_CTRL_RST_PERIOD_2_17_NANO_CLKS   << MXC_F_WDT2_CTRL_RST_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT2_CTRL_RST_PERIOD_2_17_NANO_CLKS */
#define MXC_S_WDT2_CTRL_RST_PERIOD_2_16_NANO_CLKS                               ((uint32_t)(MXC_V_WDT2_CTRL_RST_PERIOD_2_16_NANO_CLKS   << MXC_F_WDT2_CTRL_RST_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT2_CTRL_RST_PERIOD_2_16_NANO_CLKS */
#define MXC_S_WDT2_CTRL_RST_PERIOD_2_15_NANO_CLKS                               ((uint32_t)(MXC_V_WDT2_CTRL_RST_PERIOD_2_15_NANO_CLKS   << MXC_F_WDT2_CTRL_RST_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT2_CTRL_RST_PERIOD_2_15_NANO_CLKS */
#define MXC_S_WDT2_CTRL_RST_PERIOD_2_14_NANO_CLKS                               ((uint32_t)(MXC_V_WDT2_CTRL_RST_PERIOD_2_14_NANO_CLKS   << MXC_F_WDT2_CTRL_RST_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT2_CTRL_RST_PERIOD_2_14_NANO_CLKS */
#define MXC_S_WDT2_CTRL_RST_PERIOD_2_13_NANO_CLKS                               ((uint32_t)(MXC_V_WDT2_CTRL_RST_PERIOD_2_13_NANO_CLKS   << MXC_F_WDT2_CTRL_RST_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT2_CTRL_RST_PERIOD_2_13_NANO_CLKS */
#define MXC_S_WDT2_CTRL_RST_PERIOD_2_12_NANO_CLKS                               ((uint32_t)(MXC_V_WDT2_CTRL_RST_PERIOD_2_12_NANO_CLKS   << MXC_F_WDT2_CTRL_RST_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT2_CTRL_RST_PERIOD_2_12_NANO_CLKS */
#define MXC_S_WDT2_CTRL_RST_PERIOD_2_11_NANO_CLKS                               ((uint32_t)(MXC_V_WDT2_CTRL_RST_PERIOD_2_11_NANO_CLKS   << MXC_F_WDT2_CTRL_RST_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT2_CTRL_RST_PERIOD_2_11_NANO_CLKS */
#define MXC_S_WDT2_CTRL_RST_PERIOD_2_10_NANO_CLKS                               ((uint32_t)(MXC_V_WDT2_CTRL_RST_PERIOD_2_10_NANO_CLKS   << MXC_F_WDT2_CTRL_RST_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT2_CTRL_RST_PERIOD_2_10_NANO_CLKS */
/**@} end of group WDT2_CTRL_RST_PERIOD_Shifted */
/**
 * @ingroup    WDT2_LOCK_CTRL_Register
 * @defgroup   WDT2_LOCK_field_values Watchdog Timer WDT2_LOCK field values
 * @brief      Lock/Unlock values for the watchdog timer \ref WDT2_CTRL_Register. 
 * @{
 */
#define MXC_V_WDT2_LOCK_KEY    0x24     /**< Writing this value to the WDT2_LOCK field of the \ref WDT2_LOCK_CTRL_Register \b locks the \ref WDT2_CTRL_Register making it read only. */
#define MXC_V_WDT2_UNLOCK_KEY  0x42     /**< Writing this value to the WDT2_LOCK field of the \ref WDT2_LOCK_CTRL_Register \b unlocks the \ref WDT2_CTRL_Register making it read/write. */
/**@} end of group WDT2_LOCK_field_values */
///@cond
/**
 * @internal 
 * @ingroup    WDT2_CLEAR_Register
 * @defgroup   WDT2_CLEAR_field_values Watchdog Timer Clear Sequence Values
 * @brief      Writing the sequence of #MXC_V_WDT2_RESET_KEY_0, #MXC_V_WDT2_RESET_KEY_1 to the \ref WDT2_CLEAR_Register will clear/reset the watchdog timer count.
 * @note       The values #MXC_V_WDT2_RESET_KEY_0, #MXC_V_WDT2_RESET_KEY_1 must be written sequentially to the \ref WDT2_CLEAR_Register to clear the watchdog counter.
 * @{
 */
#define MXC_V_WDT2_RESET_KEY_0 0xA5    /**< First value to write to the \ref WDT2_CLEAR_Register to perform a WDT2 clear. */
#define MXC_V_WDT2_RESET_KEY_1 0x5A    /**< Second value to write to the \ref WDT2_CLEAR_Register to perform a WDT2 clear. */
/**
 * @} end of group WDT2_CLEAR_field_values 
 * @endinternal 
 */
///@endcond
/**@} wdt2_registers*/
/**@} add to group wdt2*/
#ifdef __cplusplus
}
#endif

#endif   /* _MXC_WDT2_REGS_H_ */

