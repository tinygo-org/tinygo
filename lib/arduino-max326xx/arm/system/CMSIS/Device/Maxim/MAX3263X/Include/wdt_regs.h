/**
 * @file
 * @brief   Type definitions for the Watchdog Timer Peripheral
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
 * $Date: 2016-10-10 19:53:06 -0500 (Mon, 10 Oct 2016) $
 * $Revision: 24677 $
 *
 *
 **************************************************************************** */

/* Define to prevent redundant inclusion */
#ifndef _MXC_WDT_REGS_H_
#define _MXC_WDT_REGS_H_

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
 * @ingroup     wdt0
 * @defgroup    wdt_registers Registers
 * @brief       Registers, Bit Masks and Bit Positions
 * @{
 */
/*
   Typedefed structure(s) for module registers (per instance or section) with direct 32-bit
   access to each register in module.
*/
/**
 * Structure type to access the WDT Registers, see #MXC_WDT_GET_WDT(i) to get a pointer to the WDT[i] register structure.
 * @note For the Always-On Watch Dog Timer, see \ref wdt2. 
 */
typedef struct {
    __IO uint32_t ctrl;                                 /**< <tt>\b 0x0000:</tt> WDT_CTRL Register - WDT Control Register                                   */
    __IO uint32_t clear;                                /**< <tt>\b 0x0004:</tt> WDT_CLEAR Register - WDT Clear Register to prevent a WDT Reset (Feed Dog)  */
    __IO uint32_t flags;                                /**< <tt>\b 0x0008:</tt> WDT_FLAGS Register - WDT Interrupt and Reset Flags                         */
    __IO uint32_t enable;                               /**< <tt>\b 0x000C:</tt> WDT_ENABLE Register - WDT Reset and Interrupt Enable/Disable Controls      */
    __R  uint32_t rsv010;                               /**< <tt>\b 0x0010:</tt> RESERVED, DO NOT MODIFY.                                                   */
    __IO uint32_t lock_ctrl;                            /**< <tt>\b 0x0014:</tt> WDT_LOCK_CTRL Register - Lock for Control Register                         */
} mxc_wdt_regs_t;
/**@} end of group wdt_registers.*/

/*
   Register offsets for module WDT.
*/
/**
 * @ingroup    wdt_registers
 * @defgroup   WDT_Register_Offsets Register Offsets
 * @brief      Watchdog Timer Register Offsets from the WDT[n] Base Peripheral Address, where n is between 0 and #MXC_CFG_WDT_INSTANCES for the \MXIM_Device. 
 * @details    Use #MXC_WDT_GET_BASE(i) to get the base address for a specific watchdog timer instance. 
 * @note       See \ref wdt2 for the Always-On Watchdog Timer Peripheral driver. 
 * @{
 */
#define MXC_R_WDT_OFFS_CTRL                                 ((uint32_t)0x00000000UL)    /**< Offset from the WDT[i] Base Peripheral Address : WDT_CTRL : <tt>\b 0x0000 </tt>      */
#define MXC_R_WDT_OFFS_CLEAR                                ((uint32_t)0x00000004UL)    /**< Offset from the WDT[i] Base Peripheral Address : WDT_CLEAR : <tt>\b 0x0004 </tt>     */
#define MXC_R_WDT_OFFS_FLAGS                                ((uint32_t)0x00000008UL)    /**< Offset from the WDT[i] Base Peripheral Address : WDT_FLAGS : <tt>\b 0x0008 </tt>     */
#define MXC_R_WDT_OFFS_ENABLE                               ((uint32_t)0x0000000CUL)    /**< Offset from the WDT[i] Base Peripheral Address : WDT_ENABLE : <tt>\b 0x000C </tt>    */
#define MXC_R_WDT_OFFS_LOCK_CTRL                            ((uint32_t)0x00000014UL)    /**< Offset from the WDT[i] Base Peripheral Address : WDT_LOCK_CTRL : <tt>\b 0x0014 </tt> */
/**@} end of group WDT_Register_Offsets */

/*
   Field positions and masks for module WDT.
*/
/**
 * @ingroup  wdt_registers
 * @defgroup WDT_CTRL_Register WDT_CTRL Register
 * @brief    Field Positions and Bit Masks for the WDT_CTRL register
 * @{
 */
#define MXC_F_WDT_CTRL_INT_PERIOD_POS                       0                                                                   /**< INT_PERIOD Field Position */
#define MXC_F_WDT_CTRL_INT_PERIOD                           ((uint32_t)(0x0000000FUL << MXC_F_WDT_CTRL_INT_PERIOD_POS))         /**< INT_PERIOD Field Mask - This field is used to set the interrupt period on the WDT. */
#define MXC_F_WDT_CTRL_RST_PERIOD_POS                       4                                                                   /**< RST_PERIOD Field Position */
#define MXC_F_WDT_CTRL_RST_PERIOD                           ((uint32_t)(0x0000000FUL << MXC_F_WDT_CTRL_RST_PERIOD_POS))         /**< RST_PERIOD Field Mask - This field sets the time after an
                                                                                                                                  * interrupt period has expired before the device resets. If the
                                                                                                                                  * INT_PERIOD Flag is cleared prior to the RST_PERIOD expiration,
                                                                                                                                  * the device will not reset. */
#define MXC_F_WDT_CTRL_EN_TIMER_POS                         8                                                                   /**< EN_TIMER Field Position    */
#define MXC_F_WDT_CTRL_EN_TIMER                             ((uint32_t)(0x00000001UL << MXC_F_WDT_CTRL_EN_TIMER_POS))           /**< EN_TIMER Field Mask        */
#define MXC_F_WDT_CTRL_EN_CLOCK_POS                         9                                                                   /**< EN_CLOCK Field Position    */
#define MXC_F_WDT_CTRL_EN_CLOCK                             ((uint32_t)(0x00000001UL << MXC_F_WDT_CTRL_EN_CLOCK_POS))           /**< EN_CLOCK Field Mask        */
#define MXC_F_WDT_CTRL_WAIT_PERIOD_POS                      12                                                                  /**< WAIT_PERIOD Field Position */
#define MXC_F_WDT_CTRL_WAIT_PERIOD                          ((uint32_t)(0x0000000FUL << MXC_F_WDT_CTRL_WAIT_PERIOD_POS))        /**< WAIT_PERIOD Field Mask     */
/**@} end of group WDT_CTRL */
/**
 * @ingroup  wdt_registers
 * @defgroup WDT_FLAGS_Register WDT_FLAGS Register
 * @brief    Field Positions and Bit Masks for the WDT_FLAGS register. Watchdog Timer Flags for Interrupts and Reset.
 * @{
 */  
#define MXC_F_WDT_FLAGS_TIMEOUT_POS                         0                                                                   /**< TIMEOUT Flag Position */
#define MXC_F_WDT_FLAGS_TIMEOUT                             ((uint32_t)(0x00000001UL << MXC_F_WDT_FLAGS_TIMEOUT_POS))           /**< TIMEOUT Flag Mask - if this flag is set it indicates the Watchdog Timer timed out. */
#define MXC_F_WDT_FLAGS_PRE_WIN_POS                         1                                                                   /**< PRE_WIN Flag Position */
#define MXC_F_WDT_FLAGS_PRE_WIN                             ((uint32_t)(0x00000001UL << MXC_F_WDT_FLAGS_PRE_WIN_POS))           /**< PRE_WIN Flag Mask - If the PRE_WIN flag is set it indicates the Watchdog Timer was cleared by firmware writing to the WDT_CLEAR register <b><em> during the pre-window period</em></b>. */
#define MXC_F_WDT_FLAGS_RESET_OUT_POS                       2                                                                   /**< RESET_OUT Flag Position */
#define MXC_F_WDT_FLAGS_RESET_OUT                           ((uint32_t)(0x00000001UL << MXC_F_WDT_FLAGS_RESET_OUT_POS))         /**< RESET_FLAG Flag Mask - This flag indicates that the watchdog timer timed out and the reset period elapsed without the timer being cleared. This will result in a system restart. */
/**@} end of group WDT_FLAGS */

/**
 * @ingroup  wdt_registers
 * @defgroup WDT_ENABLE_Register WDT_ENABLE Register
 * @brief    Field Positions and Bit Masks for the WDT_ENABLE register. 
 * @{
 */
#define MXC_F_WDT_ENABLE_TIMEOUT_POS                        0                                                                   /**< ENABLE_TIMEOUT Field Position */
#define MXC_F_WDT_ENABLE_TIMEOUT                            ((uint32_t)(0x00000001UL << MXC_F_WDT_ENABLE_TIMEOUT_POS))          /**< ENABLE_TIMEOUT Field Mask */
#define MXC_F_WDT_ENABLE_PRE_WIN_POS                        1                                                                   /**< ENABLE_PRE_WIN Field Position */
#define MXC_F_WDT_ENABLE_PRE_WIN                            ((uint32_t)(0x00000001UL << MXC_F_WDT_ENABLE_PRE_WIN_POS))          /**< ENABLE_PRE_WIN Field Mask */
#define MXC_F_WDT_ENABLE_RESET_OUT_POS                      2                                                                   /**< ENABLE_RESET_OUT Field Position */
#define MXC_F_WDT_ENABLE_RESET_OUT                          ((uint32_t)(0x00000001UL << MXC_F_WDT_ENABLE_RESET_OUT_POS))        /**< ENABLE_RESET_OUT Field Mask */
/**@} end of group WDT_ENABLE */

/**
 * @ingroup  wdt_registers
 * @defgroup WDT_LOCK_CTRL_Register WDT_LOCK_CTRL Register
 * @brief    The WDT_LOCK_CTRL register controls read/write access to the \ref WDT_CTRL_Register.
 * @{
 */
#define MXC_F_WDT_LOCK_CTRL_WDLOCK_POS                      0                                                             /**< WDLOCK Field's position in the WDT_LOCK_CTRL register. */ 
#define MXC_F_WDT_LOCK_CTRL_WDLOCK                          ((uint32_t)(0x000000FFUL << MXC_F_WDT_LOCK_CTRL_WDLOCK_POS))  /**< WDLOCK Field mask for the WDT_LOCK_CTRL register.  Reading a value of */
/**@} end of group WDT_ENABLE */



/*
   Field values and shifted values for module WDT.
*/
/**
 * @ingroup WDT_CTRL_Register 
 * @defgroup WDT_CTRL_field_values WDT_CTRL Register Field and Shifted Field Values
 * @brief    Field values and Shifted Field values for the WDT_CTRL register. 
 * @details  Shifted field values are field values shifted to the loacation of the field in the register.  
 */
/**
 * @ingroup    WDT_CTRL_field_values
 * @defgroup   WDT_CTRL_INT_PERIOD_Value Watchdog Timer Interrupt Period
 * @brief      Sets the duration of the watchdog interrupt period.
 * @details    The INT_PERIOD field sets the duration of the watchdog interrupt
 *             period, which is the time period from the WDT being
 *             enabled/cleared until the WDT flag, #MXC_F_WDT_FLAGS_TIMEOUT, is
 *             set.
 *             The values defined are in the number of watchdog clock cycles.
 * @{
 */
#define MXC_V_WDT_CTRL_INT_PERIOD_2_31_CLKS                                     ((uint32_t)(0x00000000UL))  /**< Interupt Period of \f$ 2^{31} \f$ WDT CLK Cycles */
#define MXC_V_WDT_CTRL_INT_PERIOD_2_30_CLKS                                     ((uint32_t)(0x00000001UL))  /**< Interupt Period of \f$ 2^{30} \f$ WDT CLK Cycles */
#define MXC_V_WDT_CTRL_INT_PERIOD_2_29_CLKS                                     ((uint32_t)(0x00000002UL))  /**< Interupt Period of \f$ 2^{29} \f$ WDT CLK Cycles */
#define MXC_V_WDT_CTRL_INT_PERIOD_2_28_CLKS                                     ((uint32_t)(0x00000003UL))  /**< Interupt Period of \f$ 2^{28} \f$ WDT CLK Cycles */
#define MXC_V_WDT_CTRL_INT_PERIOD_2_27_CLKS                                     ((uint32_t)(0x00000004UL))  /**< Interupt Period of \f$ 2^{27} \f$ WDT CLK Cycles */
#define MXC_V_WDT_CTRL_INT_PERIOD_2_26_CLKS                                     ((uint32_t)(0x00000005UL))  /**< Interupt Period of \f$ 2^{26} \f$ WDT CLK Cycles */
#define MXC_V_WDT_CTRL_INT_PERIOD_2_25_CLKS                                     ((uint32_t)(0x00000006UL))  /**< Interupt Period of \f$ 2^{25} \f$ WDT CLK Cycles */
#define MXC_V_WDT_CTRL_INT_PERIOD_2_24_CLKS                                     ((uint32_t)(0x00000007UL))  /**< Interupt Period of \f$ 2^{24} \f$ WDT CLK Cycles */
#define MXC_V_WDT_CTRL_INT_PERIOD_2_23_CLKS                                     ((uint32_t)(0x00000008UL))  /**< Interupt Period of \f$ 2^{23} \f$ WDT CLK Cycles */
#define MXC_V_WDT_CTRL_INT_PERIOD_2_22_CLKS                                     ((uint32_t)(0x00000009UL))  /**< Interupt Period of \f$ 2^{22} \f$ WDT CLK Cycles */
#define MXC_V_WDT_CTRL_INT_PERIOD_2_21_CLKS                                     ((uint32_t)(0x0000000AUL))  /**< Interupt Period of \f$ 2^{21} \f$ WDT CLK Cycles */
#define MXC_V_WDT_CTRL_INT_PERIOD_2_20_CLKS                                     ((uint32_t)(0x0000000BUL))  /**< Interupt Period of \f$ 2^{20} \f$ WDT CLK Cycles */
#define MXC_V_WDT_CTRL_INT_PERIOD_2_19_CLKS                                     ((uint32_t)(0x0000000CUL))  /**< Interupt Period of \f$ 2^{19} \f$ WDT CLK Cycles */
#define MXC_V_WDT_CTRL_INT_PERIOD_2_18_CLKS                                     ((uint32_t)(0x0000000DUL))  /**< Interupt Period of \f$ 2^{18} \f$ WDT CLK Cycles */
#define MXC_V_WDT_CTRL_INT_PERIOD_2_17_CLKS                                     ((uint32_t)(0x0000000EUL))  /**< Interupt Period of \f$ 2^{17} \f$ WDT CLK Cycles */
#define MXC_V_WDT_CTRL_INT_PERIOD_2_16_CLKS                                     ((uint32_t)(0x0000000FUL))  /**< Interupt Period of \f$ 2^{16} \f$ WDT CLK Cycles */
/**@} end of group WDT_CTRL_INT_PERIOD_Value */

/**
 * @ingroup    WDT_CTRL_field_values
 * @defgroup   WDT_CTRL_INT_PERIOD_Shifted Watchdog Timer Interrupt Period Shifted Values
 * @brief      Shifted values for the \ref WDT_CTRL_INT_PERIOD_Value
 * @details    The shifted value is
 *             shifted to align with the fields location in the WDT_CTRL register. 
 * @{
 */
#define MXC_S_WDT_CTRL_INT_PERIOD_2_31_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_INT_PERIOD_2_31_CLKS   << MXC_F_WDT_CTRL_INT_PERIOD_POS))  /**< Shifted Field Value for #MXC_V_WDT_CTRL_INT_PERIOD_2_31_CLKS */
#define MXC_S_WDT_CTRL_INT_PERIOD_2_30_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_INT_PERIOD_2_30_CLKS   << MXC_F_WDT_CTRL_INT_PERIOD_POS))  /**< Shifted Field Value for #MXC_V_WDT_CTRL_INT_PERIOD_2_30_CLKS */
#define MXC_S_WDT_CTRL_INT_PERIOD_2_29_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_INT_PERIOD_2_29_CLKS   << MXC_F_WDT_CTRL_INT_PERIOD_POS))  /**< Shifted Field Value for #MXC_V_WDT_CTRL_INT_PERIOD_2_29_CLKS */
#define MXC_S_WDT_CTRL_INT_PERIOD_2_28_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_INT_PERIOD_2_28_CLKS   << MXC_F_WDT_CTRL_INT_PERIOD_POS))  /**< Shifted Field Value for #MXC_V_WDT_CTRL_INT_PERIOD_2_28_CLKS */
#define MXC_S_WDT_CTRL_INT_PERIOD_2_27_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_INT_PERIOD_2_27_CLKS   << MXC_F_WDT_CTRL_INT_PERIOD_POS))  /**< Shifted Field Value for #MXC_V_WDT_CTRL_INT_PERIOD_2_27_CLKS */
#define MXC_S_WDT_CTRL_INT_PERIOD_2_26_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_INT_PERIOD_2_26_CLKS   << MXC_F_WDT_CTRL_INT_PERIOD_POS))  /**< Shifted Field Value for #MXC_V_WDT_CTRL_INT_PERIOD_2_26_CLKS */
#define MXC_S_WDT_CTRL_INT_PERIOD_2_25_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_INT_PERIOD_2_25_CLKS   << MXC_F_WDT_CTRL_INT_PERIOD_POS))  /**< Shifted Field Value for #MXC_V_WDT_CTRL_INT_PERIOD_2_25_CLKS */
#define MXC_S_WDT_CTRL_INT_PERIOD_2_24_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_INT_PERIOD_2_24_CLKS   << MXC_F_WDT_CTRL_INT_PERIOD_POS))  /**< Shifted Field Value for #MXC_V_WDT_CTRL_INT_PERIOD_2_24_CLKS */
#define MXC_S_WDT_CTRL_INT_PERIOD_2_23_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_INT_PERIOD_2_23_CLKS   << MXC_F_WDT_CTRL_INT_PERIOD_POS))  /**< Shifted Field Value for #MXC_V_WDT_CTRL_INT_PERIOD_2_23_CLKS */
#define MXC_S_WDT_CTRL_INT_PERIOD_2_22_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_INT_PERIOD_2_22_CLKS   << MXC_F_WDT_CTRL_INT_PERIOD_POS))  /**< Shifted Field Value for #MXC_V_WDT_CTRL_INT_PERIOD_2_22_CLKS */
#define MXC_S_WDT_CTRL_INT_PERIOD_2_21_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_INT_PERIOD_2_21_CLKS   << MXC_F_WDT_CTRL_INT_PERIOD_POS))  /**< Shifted Field Value for #MXC_V_WDT_CTRL_INT_PERIOD_2_21_CLKS */
#define MXC_S_WDT_CTRL_INT_PERIOD_2_20_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_INT_PERIOD_2_20_CLKS   << MXC_F_WDT_CTRL_INT_PERIOD_POS))  /**< Shifted Field Value for #MXC_V_WDT_CTRL_INT_PERIOD_2_20_CLKS */
#define MXC_S_WDT_CTRL_INT_PERIOD_2_19_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_INT_PERIOD_2_19_CLKS   << MXC_F_WDT_CTRL_INT_PERIOD_POS))  /**< Shifted Field Value for #MXC_V_WDT_CTRL_INT_PERIOD_2_19_CLKS */
#define MXC_S_WDT_CTRL_INT_PERIOD_2_18_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_INT_PERIOD_2_18_CLKS   << MXC_F_WDT_CTRL_INT_PERIOD_POS))  /**< Shifted Field Value for #MXC_V_WDT_CTRL_INT_PERIOD_2_18_CLKS */
#define MXC_S_WDT_CTRL_INT_PERIOD_2_17_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_INT_PERIOD_2_17_CLKS   << MXC_F_WDT_CTRL_INT_PERIOD_POS))  /**< Shifted Field Value for #MXC_V_WDT_CTRL_INT_PERIOD_2_17_CLKS */
#define MXC_S_WDT_CTRL_INT_PERIOD_2_16_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_INT_PERIOD_2_16_CLKS   << MXC_F_WDT_CTRL_INT_PERIOD_POS))  /**< Shifted Field Value for #MXC_V_WDT_CTRL_INT_PERIOD_2_16_CLKS */
/**@} end of group WDT_CTRL_INT_PERIOD_Shifted */
/**
 * @ingroup    WDT_CTRL_field_values
 * @defgroup   WDT_CTRL_RST_PERIOD_Value Watchdog Timer Reset Period
 * @brief      Sets the duration of the watchdog reset period.
 * @details    The RST_PERIOD field sets the duration of the watchdog reset
 *             period, which is the time period from the WDT being
 *             enabled/cleared until the WDT flag, #MXC_F_WDT_CTRL_RST_PERIOD is
 *             set.
 *             The values defined are in the number of watchdog clock cycles.
 * @{
 */
#define MXC_V_WDT_CTRL_RST_PERIOD_2_31_CLKS                                     ((uint32_t)(0x00000000UL))        /**< Reset Period of \f$ 2^{31} \f$ WDT CLK CYCLES  */
#define MXC_V_WDT_CTRL_RST_PERIOD_2_30_CLKS                                     ((uint32_t)(0x00000001UL))        /**< Reset Period of \f$ 2^{30} \f$ WDT CLK CYCLES  */
#define MXC_V_WDT_CTRL_RST_PERIOD_2_29_CLKS                                     ((uint32_t)(0x00000002UL))        /**< Reset Period of \f$ 2^{29} \f$ WDT CLK CYCLES  */
#define MXC_V_WDT_CTRL_RST_PERIOD_2_28_CLKS                                     ((uint32_t)(0x00000003UL))        /**< Reset Period of \f$ 2^{28} \f$ WDT CLK CYCLES  */
#define MXC_V_WDT_CTRL_RST_PERIOD_2_27_CLKS                                     ((uint32_t)(0x00000004UL))        /**< Reset Period of \f$ 2^{27} \f$ WDT CLK CYCLES  */
#define MXC_V_WDT_CTRL_RST_PERIOD_2_26_CLKS                                     ((uint32_t)(0x00000005UL))        /**< Reset Period of \f$ 2^{26} \f$ WDT CLK CYCLES  */
#define MXC_V_WDT_CTRL_RST_PERIOD_2_25_CLKS                                     ((uint32_t)(0x00000006UL))        /**< Reset Period of \f$ 2^{25} \f$ WDT CLK CYCLES  */
#define MXC_V_WDT_CTRL_RST_PERIOD_2_24_CLKS                                     ((uint32_t)(0x00000007UL))        /**< Reset Period of \f$ 2^{24} \f$ WDT CLK CYCLES  */
#define MXC_V_WDT_CTRL_RST_PERIOD_2_23_CLKS                                     ((uint32_t)(0x00000008UL))        /**< Reset Period of \f$ 2^{23} \f$ WDT CLK CYCLES  */
#define MXC_V_WDT_CTRL_RST_PERIOD_2_22_CLKS                                     ((uint32_t)(0x00000009UL))        /**< Reset Period of \f$ 2^{22} \f$ WDT CLK CYCLES  */
#define MXC_V_WDT_CTRL_RST_PERIOD_2_21_CLKS                                     ((uint32_t)(0x0000000AUL))        /**< Reset Period of \f$ 2^{21} \f$ WDT CLK CYCLES  */
#define MXC_V_WDT_CTRL_RST_PERIOD_2_20_CLKS                                     ((uint32_t)(0x0000000BUL))        /**< Reset Period of \f$ 2^{20} \f$ WDT CLK CYCLES  */
#define MXC_V_WDT_CTRL_RST_PERIOD_2_19_CLKS                                     ((uint32_t)(0x0000000CUL))        /**< Reset Period of \f$ 2^{19} \f$ WDT CLK CYCLES  */
#define MXC_V_WDT_CTRL_RST_PERIOD_2_18_CLKS                                     ((uint32_t)(0x0000000DUL))        /**< Reset Period of \f$ 2^{18} \f$ WDT CLK CYCLES  */
#define MXC_V_WDT_CTRL_RST_PERIOD_2_17_CLKS                                     ((uint32_t)(0x0000000EUL))        /**< Reset Period of \f$ 2^{17} \f$ WDT CLK CYCLES  */
#define MXC_V_WDT_CTRL_RST_PERIOD_2_16_CLKS                                     ((uint32_t)(0x0000000FUL))        /**< Reset Period of \f$ 2^{16} \f$ WDT CLK CYCLES  */
/**@} end of group WDT_CTRL_RST_PERIOD_Value */

/**
 * @ingroup    WDT_CTRL_field_values
 * @defgroup   WDT_CTRL_RST_PERIOD_Shifted Watchdog Timer Reset Period Shifted Values
 * @brief      Shifted values for the \ref WDT_CTRL_RST_PERIOD_Value
 * @details    These values are shifted to align with the field's location in the WDT_CTRL register. 
 * @{
 */
#define MXC_S_WDT_CTRL_RST_PERIOD_2_31_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_RST_PERIOD_2_31_CLKS   << MXC_F_WDT_CTRL_RST_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT_CTRL_RST_PERIOD_2_31_CLKS */
#define MXC_S_WDT_CTRL_RST_PERIOD_2_30_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_RST_PERIOD_2_30_CLKS   << MXC_F_WDT_CTRL_RST_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT_CTRL_RST_PERIOD_2_30_CLKS */
#define MXC_S_WDT_CTRL_RST_PERIOD_2_29_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_RST_PERIOD_2_29_CLKS   << MXC_F_WDT_CTRL_RST_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT_CTRL_RST_PERIOD_2_29_CLKS */
#define MXC_S_WDT_CTRL_RST_PERIOD_2_28_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_RST_PERIOD_2_28_CLKS   << MXC_F_WDT_CTRL_RST_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT_CTRL_RST_PERIOD_2_28_CLKS */
#define MXC_S_WDT_CTRL_RST_PERIOD_2_27_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_RST_PERIOD_2_27_CLKS   << MXC_F_WDT_CTRL_RST_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT_CTRL_RST_PERIOD_2_27_CLKS */
#define MXC_S_WDT_CTRL_RST_PERIOD_2_26_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_RST_PERIOD_2_26_CLKS   << MXC_F_WDT_CTRL_RST_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT_CTRL_RST_PERIOD_2_26_CLKS */
#define MXC_S_WDT_CTRL_RST_PERIOD_2_25_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_RST_PERIOD_2_25_CLKS   << MXC_F_WDT_CTRL_RST_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT_CTRL_RST_PERIOD_2_25_CLKS */
#define MXC_S_WDT_CTRL_RST_PERIOD_2_24_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_RST_PERIOD_2_24_CLKS   << MXC_F_WDT_CTRL_RST_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT_CTRL_RST_PERIOD_2_24_CLKS */
#define MXC_S_WDT_CTRL_RST_PERIOD_2_23_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_RST_PERIOD_2_23_CLKS   << MXC_F_WDT_CTRL_RST_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT_CTRL_RST_PERIOD_2_23_CLKS */
#define MXC_S_WDT_CTRL_RST_PERIOD_2_22_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_RST_PERIOD_2_22_CLKS   << MXC_F_WDT_CTRL_RST_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT_CTRL_RST_PERIOD_2_22_CLKS */
#define MXC_S_WDT_CTRL_RST_PERIOD_2_21_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_RST_PERIOD_2_21_CLKS   << MXC_F_WDT_CTRL_RST_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT_CTRL_RST_PERIOD_2_21_CLKS */
#define MXC_S_WDT_CTRL_RST_PERIOD_2_20_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_RST_PERIOD_2_20_CLKS   << MXC_F_WDT_CTRL_RST_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT_CTRL_RST_PERIOD_2_20_CLKS */
#define MXC_S_WDT_CTRL_RST_PERIOD_2_19_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_RST_PERIOD_2_19_CLKS   << MXC_F_WDT_CTRL_RST_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT_CTRL_RST_PERIOD_2_19_CLKS */
#define MXC_S_WDT_CTRL_RST_PERIOD_2_18_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_RST_PERIOD_2_18_CLKS   << MXC_F_WDT_CTRL_RST_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT_CTRL_RST_PERIOD_2_18_CLKS */
#define MXC_S_WDT_CTRL_RST_PERIOD_2_17_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_RST_PERIOD_2_17_CLKS   << MXC_F_WDT_CTRL_RST_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT_CTRL_RST_PERIOD_2_17_CLKS */
#define MXC_S_WDT_CTRL_RST_PERIOD_2_16_CLKS                                     ((uint32_t)(MXC_V_WDT_CTRL_RST_PERIOD_2_16_CLKS   << MXC_F_WDT_CTRL_RST_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT_CTRL_RST_PERIOD_2_16_CLKS */
/**@} end of group WDT_CTRL_RST_PERIOD_Shifted */
/**
 * @ingroup    WDT_CTRL_field_values
 * @defgroup   WDT_CTRL_WAIT_PERIOD_Value Watchdog Timer Wait Period
 * @brief      Sets the duration of the watchdog wait window period.
 * @details    The WAIT_PERIOD field sets the duration of the watchdog pre-window
 *             period. If the watchdog is reset before the wait period has finished, an out-of-window interrupt will occur. 
 *             This sets the minimum amount of time between watchdog enable/clear to resetting the WDT count and assists in detecting 
 *             overclocking or an invalid clock.
 *             The values defined are in the number of watchdog clock cycles.
 * @{
 */
#define MXC_V_WDT_CTRL_WAIT_PERIOD_2_31_CLKS                                    ((uint32_t)(0x00000000UL))        /**< Wait Period of \f$ 2^{31} \f$ WDT CLK CYCLES  */
#define MXC_V_WDT_CTRL_WAIT_PERIOD_2_30_CLKS                                    ((uint32_t)(0x00000001UL))        /**< Wait Period of \f$ 2^{30} \f$ WDT CLK CYCLES  */
#define MXC_V_WDT_CTRL_WAIT_PERIOD_2_29_CLKS                                    ((uint32_t)(0x00000002UL))        /**< Wait Period of \f$ 2^{29} \f$ WDT CLK CYCLES  */
#define MXC_V_WDT_CTRL_WAIT_PERIOD_2_28_CLKS                                    ((uint32_t)(0x00000003UL))        /**< Wait Period of \f$ 2^{28} \f$ WDT CLK CYCLES  */
#define MXC_V_WDT_CTRL_WAIT_PERIOD_2_27_CLKS                                    ((uint32_t)(0x00000004UL))        /**< Wait Period of \f$ 2^{27} \f$ WDT CLK CYCLES  */
#define MXC_V_WDT_CTRL_WAIT_PERIOD_2_26_CLKS                                    ((uint32_t)(0x00000005UL))        /**< Wait Period of \f$ 2^{26} \f$ WDT CLK CYCLES  */
#define MXC_V_WDT_CTRL_WAIT_PERIOD_2_25_CLKS                                    ((uint32_t)(0x00000006UL))        /**< Wait Period of \f$ 2^{25} \f$ WDT CLK CYCLES  */
#define MXC_V_WDT_CTRL_WAIT_PERIOD_2_24_CLKS                                    ((uint32_t)(0x00000007UL))        /**< Wait Period of \f$ 2^{24} \f$ WDT CLK CYCLES  */
#define MXC_V_WDT_CTRL_WAIT_PERIOD_2_23_CLKS                                    ((uint32_t)(0x00000008UL))        /**< Wait Period of \f$ 2^{23} \f$ WDT CLK CYCLES  */
#define MXC_V_WDT_CTRL_WAIT_PERIOD_2_22_CLKS                                    ((uint32_t)(0x00000009UL))        /**< Wait Period of \f$ 2^{22} \f$ WDT CLK CYCLES  */
#define MXC_V_WDT_CTRL_WAIT_PERIOD_2_21_CLKS                                    ((uint32_t)(0x0000000AUL))        /**< Wait Period of \f$ 2^{21} \f$ WDT CLK CYCLES  */
#define MXC_V_WDT_CTRL_WAIT_PERIOD_2_20_CLKS                                    ((uint32_t)(0x0000000BUL))        /**< Wait Period of \f$ 2^{20} \f$ WDT CLK CYCLES  */
#define MXC_V_WDT_CTRL_WAIT_PERIOD_2_19_CLKS                                    ((uint32_t)(0x0000000CUL))        /**< Wait Period of \f$ 2^{19} \f$ WDT CLK CYCLES  */
#define MXC_V_WDT_CTRL_WAIT_PERIOD_2_18_CLKS                                    ((uint32_t)(0x0000000DUL))        /**< Wait Period of \f$ 2^{18} \f$ WDT CLK CYCLES  */
#define MXC_V_WDT_CTRL_WAIT_PERIOD_2_17_CLKS                                    ((uint32_t)(0x0000000EUL))        /**< Wait Period of \f$ 2^{17} \f$ WDT CLK CYCLES  */
#define MXC_V_WDT_CTRL_WAIT_PERIOD_2_16_CLKS                                    ((uint32_t)(0x0000000FUL))        /**< Wait Period of \f$ 2^{16} \f$ WDT CLK CYCLES  */
/**@} end of group WDT_CTRL_WAIT_PERIOD_Value */

/**
 * @ingroup    WDT_CTRL_field_values
 * @defgroup   WDT_CTRL_WAIT_PERIOD_Shifted Watchdog Timer Wait Period Shifted Values
 * @brief      Shifted values for the \ref WDT_CTRL_WAIT_PERIOD_Value
 * @details    These values are shifted to align with the WAIT_PERIOD field's location in the WDT_CTRL register. 
 * @{
 */
#define MXC_S_WDT_CTRL_WAIT_PERIOD_2_31_CLKS                                    ((uint32_t)(MXC_V_WDT_CTRL_WAIT_PERIOD_2_31_CLKS   << MXC_F_WDT_CTRL_WAIT_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT_CTRL_WAIT_PERIOD_2_31_CLKS */
#define MXC_S_WDT_CTRL_WAIT_PERIOD_2_30_CLKS                                    ((uint32_t)(MXC_V_WDT_CTRL_WAIT_PERIOD_2_30_CLKS   << MXC_F_WDT_CTRL_WAIT_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT_CTRL_WAIT_PERIOD_2_30_CLKS */
#define MXC_S_WDT_CTRL_WAIT_PERIOD_2_29_CLKS                                    ((uint32_t)(MXC_V_WDT_CTRL_WAIT_PERIOD_2_29_CLKS   << MXC_F_WDT_CTRL_WAIT_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT_CTRL_WAIT_PERIOD_2_29_CLKS */
#define MXC_S_WDT_CTRL_WAIT_PERIOD_2_28_CLKS                                    ((uint32_t)(MXC_V_WDT_CTRL_WAIT_PERIOD_2_28_CLKS   << MXC_F_WDT_CTRL_WAIT_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT_CTRL_WAIT_PERIOD_2_28_CLKS */
#define MXC_S_WDT_CTRL_WAIT_PERIOD_2_27_CLKS                                    ((uint32_t)(MXC_V_WDT_CTRL_WAIT_PERIOD_2_27_CLKS   << MXC_F_WDT_CTRL_WAIT_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT_CTRL_WAIT_PERIOD_2_27_CLKS */
#define MXC_S_WDT_CTRL_WAIT_PERIOD_2_26_CLKS                                    ((uint32_t)(MXC_V_WDT_CTRL_WAIT_PERIOD_2_26_CLKS   << MXC_F_WDT_CTRL_WAIT_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT_CTRL_WAIT_PERIOD_2_26_CLKS */
#define MXC_S_WDT_CTRL_WAIT_PERIOD_2_25_CLKS                                    ((uint32_t)(MXC_V_WDT_CTRL_WAIT_PERIOD_2_25_CLKS   << MXC_F_WDT_CTRL_WAIT_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT_CTRL_WAIT_PERIOD_2_25_CLKS */
#define MXC_S_WDT_CTRL_WAIT_PERIOD_2_24_CLKS                                    ((uint32_t)(MXC_V_WDT_CTRL_WAIT_PERIOD_2_24_CLKS   << MXC_F_WDT_CTRL_WAIT_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT_CTRL_WAIT_PERIOD_2_24_CLKS */
#define MXC_S_WDT_CTRL_WAIT_PERIOD_2_23_CLKS                                    ((uint32_t)(MXC_V_WDT_CTRL_WAIT_PERIOD_2_23_CLKS   << MXC_F_WDT_CTRL_WAIT_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT_CTRL_WAIT_PERIOD_2_23_CLKS */
#define MXC_S_WDT_CTRL_WAIT_PERIOD_2_22_CLKS                                    ((uint32_t)(MXC_V_WDT_CTRL_WAIT_PERIOD_2_22_CLKS   << MXC_F_WDT_CTRL_WAIT_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT_CTRL_WAIT_PERIOD_2_22_CLKS */
#define MXC_S_WDT_CTRL_WAIT_PERIOD_2_21_CLKS                                    ((uint32_t)(MXC_V_WDT_CTRL_WAIT_PERIOD_2_21_CLKS   << MXC_F_WDT_CTRL_WAIT_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT_CTRL_WAIT_PERIOD_2_21_CLKS */
#define MXC_S_WDT_CTRL_WAIT_PERIOD_2_20_CLKS                                    ((uint32_t)(MXC_V_WDT_CTRL_WAIT_PERIOD_2_20_CLKS   << MXC_F_WDT_CTRL_WAIT_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT_CTRL_WAIT_PERIOD_2_20_CLKS */
#define MXC_S_WDT_CTRL_WAIT_PERIOD_2_19_CLKS                                    ((uint32_t)(MXC_V_WDT_CTRL_WAIT_PERIOD_2_19_CLKS   << MXC_F_WDT_CTRL_WAIT_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT_CTRL_WAIT_PERIOD_2_19_CLKS */
#define MXC_S_WDT_CTRL_WAIT_PERIOD_2_18_CLKS                                    ((uint32_t)(MXC_V_WDT_CTRL_WAIT_PERIOD_2_18_CLKS   << MXC_F_WDT_CTRL_WAIT_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT_CTRL_WAIT_PERIOD_2_18_CLKS */
#define MXC_S_WDT_CTRL_WAIT_PERIOD_2_17_CLKS                                    ((uint32_t)(MXC_V_WDT_CTRL_WAIT_PERIOD_2_17_CLKS   << MXC_F_WDT_CTRL_WAIT_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT_CTRL_WAIT_PERIOD_2_17_CLKS */
#define MXC_S_WDT_CTRL_WAIT_PERIOD_2_16_CLKS                                    ((uint32_t)(MXC_V_WDT_CTRL_WAIT_PERIOD_2_16_CLKS   << MXC_F_WDT_CTRL_WAIT_PERIOD_POS))      /**< Shifted Field Value for #MXC_V_WDT_CTRL_WAIT_PERIOD_2_16_CLKS */
/**@} end of group WDT_CTRL_WAIT_PERIOD_Shifted */

/**
 * @ingroup    WDT_LOCK_CTRL_Register
 * @defgroup   WDT_LOCK_field_values Watchdog Timer WDT_LOCK field values
 * @brief      Lock/Unlock values for the watchdog timer \ref WDT_CTRL_Register. 
 * @{
 */
#define MXC_V_WDT_LOCK_KEY    0x24    /**< Writing this value to the WDT_LOCK field of the \ref WDT_LOCK_CTRL_Register \b locks the \ref WDT_CTRL_Register making it read only. */
#define MXC_V_WDT_UNLOCK_KEY  0x42    /**< Writing this value to the WDT_LOCK field of the \ref WDT_LOCK_CTRL_Register \b unlocks the \ref WDT_CTRL_Register making it read/write. */
/**@} end of group WDT_LOCK_field_values */
///@cond
/*
 * @internal
 * @ingroup    WDT_CLEAR_Register
 * @defgroup   WDT_CLEAR_field_values Watchdog Timer Clear Sequence Values
 * @brief      Writing the sequence of #MXC_V_WDT_RESET_KEY_0, #MXC_V_WDT_RESET_KEY_1 to the \ref WDT_CLEAR_Register will clear/reset the watchdog timer count.
 * @note       The values #MXC_V_WDT_RESET_KEY_0, #MXC_V_WDT_RESET_KEY_1 must be written sequentially to the \ref WDT_CLEAR_Register to clear the watchdog counter.
 * @{
 */
#define MXC_V_WDT_RESET_KEY_0 0xA5    /**< First value to write to the \ref WDT_CLEAR_Register to perform a WDT clear. */
#define MXC_V_WDT_RESET_KEY_1 0x5A    /**< Second value to write to the \ref WDT_CLEAR_Register to perform a WDT clear. */
/**
 * @} end of group WDT_CLEAR_field_values 
 * @endinternal 
 */
///@endcond
/**@} wdt_registers*/

#ifdef __cplusplus
}
#endif

#endif   /* _MXC_WDT_REGS_H_ */

