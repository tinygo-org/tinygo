/**
 * @file
 * @brief      Registers, Bit Masks and Bit Positions for the Timer Peripheral
 *             Module.
 */
/* *****************************************************************************
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
 * $Date: 2016-10-10 19:49:16 -0500 (Mon, 10 Oct 2016) $
 * $Revision: 24675 $
 *
 *************************************************************************** */

/* Define to prevent redundant inclusion */
#ifndef _MXC_TMR_REGS_H_
#define _MXC_TMR_REGS_H_

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
 * @ingroup    tmr
 * @defgroup   tmr_registers Timer Registers
 * @brief      Hardware interface definitions for the Timer Peripheral.
 * @details    Definitions for the Hardware Access Layer of the Timer
 *             Peripherals. Includes:
 * - Registers
 * - Fields
 *   - Positions
 *   - Values
 *   - Masks 
 * @{
 */

/* **** Definitions **** */

/**
 * Structure type to access the Timer Registers, see #MXC_TMR_GET_TMR(i) to get a pointer to the Timer[i] register structure. 
 */
typedef struct {
    __IO uint32_t ctrl;                                 /**< <tt>\b 0x0000</tt> - TMR_CTRL Register - Timer Control Register                                            */
    __IO uint32_t count32;                              /**< <tt>\b 0x0004</tt> - TMR_COUNT32 Register - Timer [32 bit] Current Count Value                             */
    __IO uint32_t term_cnt32;                           /**< <tt>\b 0x0008</tt> - TMR_TERM_CNT32 Register - Timer [32 bit] Terminal Count Setting                       */
    __IO uint32_t pwm_cap32;                            /**< <tt>\b 0x000C</tt> - TMR_PWM_CAP32 Register - Timer [32 bit] PWM Compare Setting or Capture/Measure Value  */
    __IO uint32_t count16_0;                            /**< <tt>\b 0x0010</tt> - TMR_COUNT16_0 Register - Timer [16 bit] Current Count Value, 16-bit Timer 0           */
    __IO uint32_t term_cnt16_0;                         /**< <tt>\b 0x0014</tt> - TMR_TERM_CNT16_0 Register - Timer [16 bit] Terminal Count Setting, 16-bit Timer 0     */
    __IO uint32_t count16_1;                            /**< <tt>\b 0x0018</tt> - TMR_COUNT16_1 Register - Timer [16 bit] Current Count Value, 16-bit Timer 1           */
    __IO uint32_t term_cnt16_1;                         /**< <tt>\b 0x001C</tt> - TMR_TERM_CNT16_1 Register - Timer [16 bit] Terminal Count Setting, 16-bit Timer 1     */
    __IO uint32_t intfl;                                /**< <tt>\b 0x0020</tt> - TMR_INTFL Register - Timer Interrupt Flags                                            */
    __IO uint32_t inten;                                /**< <tt>\b 0x0024</tt> - TMR_INTEN Register - Timer Interrupt Enable/Disable Settings                          */
} mxc_tmr_regs_t;
/**@} end of group tmr_registers. */


/*
   Register offsets for module TMR.
*/
/**
 * @ingroup    tmr_registers
 * @defgroup   TMR_Register_Offsets Register Offsets
 * @brief      Timer Register Offsets from the Timer[n] Base Peripheral Address, where n is between 0 and #MXC_CFG_TMR_INSTANCES for the \MXIM_Device. Use #MXC_TMR_GET_BASE(i) to get the base address for a specific timer number. 
 * @{
 */
#define MXC_R_TMR_OFFS_CTRL                                 ((uint32_t)0x00000000UL)        /**< Offset from TMR[n] Base Address: TMR_CTRL         : <tt>\b 0x0x0000 </tt>  */
#define MXC_R_TMR_OFFS_COUNT32                              ((uint32_t)0x00000004UL)        /**< Offset from TMR[n] Base Address: TMR_COUNT32      : <tt>\b 0x0x0004 </tt>  */
#define MXC_R_TMR_OFFS_TERM_CNT32                           ((uint32_t)0x00000008UL)        /**< Offset from TMR[n] Base Address: TMR_TERM_CNT32   : <tt>\b 0x0x0008 </tt>  */
#define MXC_R_TMR_OFFS_PWM_CAP32                            ((uint32_t)0x0000000CUL)        /**< Offset from TMR[n] Base Address: TMR_PWM_CAP32    : <tt>\b 0x0x000C </tt>  */
#define MXC_R_TMR_OFFS_COUNT16_0                            ((uint32_t)0x00000010UL)        /**< Offset from TMR[n] Base Address: TMR_COUNT16_0    : <tt>\b 0x0x0010 </tt>  */
#define MXC_R_TMR_OFFS_TERM_CNT16_0                         ((uint32_t)0x00000014UL)        /**< Offset from TMR[n] Base Address: TMR_TERM_CNT16_0 : <tt>\b 0x0x0014 </tt>  */
#define MXC_R_TMR_OFFS_COUNT16_1                            ((uint32_t)0x00000018UL)        /**< Offset from TMR[n] Base Address: TMR_COUNT16_1    : <tt>\b 0x0x0018 </tt>  */
#define MXC_R_TMR_OFFS_TERM_CNT16_1                         ((uint32_t)0x0000001CUL)        /**< Offset from TMR[n] Base Address: TMR_TERM_CNT16_1 : <tt>\b 0x0x001C </tt>  */
#define MXC_R_TMR_OFFS_INTFL                                ((uint32_t)0x00000020UL)        /**< Offset from TMR[n] Base Address: TMR_INTFL        : <tt>\b 0x0x0020 </tt>  */
#define MXC_R_TMR_OFFS_INTEN                                ((uint32_t)0x00000024UL)        /**< Offset from TMR[n] Base Address: TMR_INTEN        : <tt>\b 0x0x0024 </tt>  */
/**@} end of group TMR_Register_Offsets */

/**
 * @ingroup  tmr_registers
 * @defgroup TMR_CTRL_Register TMR_CTRL Register
 * @brief    Field Positions and Bit Masks for the TMR_CTRL register
 * @{
 */
#define MXC_F_TMR_CTRL_MODE_POS                             0                                                           /**< MODE Field Position for 32-bit timer if TMR2X16 Field is 0 (Default) */
#define MXC_F_TMR_CTRL_MODE                                 ((uint32_t)(0x00000007UL << MXC_F_TMR_CTRL_MODE_POS))       /**< MODE Field Shifted Position for 32-bit timer if TMR2X16 Field is 0 (Default) */
#define MXC_F_TMR_CTRL_TMR2X16_POS                          3                                                           /**< TMR2X16 Field Position */
#define MXC_F_TMR_CTRL_TMR2X16                              ((uint32_t)(0x00000001UL << MXC_F_TMR_CTRL_TMR2X16_POS))    /**< TMR2X16 Field Shifted Position */
#define MXC_F_TMR_CTRL_PRESCALE_POS                         4                                                           /**< PRESCALE Field Position */
#define MXC_F_TMR_CTRL_PRESCALE                             ((uint32_t)(0x0000000FUL << MXC_F_TMR_CTRL_PRESCALE_POS))   /**< PRESCALE Field Shifted Position */
#define MXC_F_TMR_CTRL_POLARITY_POS                         8                                                           /**< POLARITY Field Position */
#define MXC_F_TMR_CTRL_POLARITY                             ((uint32_t)(0x00000001UL << MXC_F_TMR_CTRL_POLARITY_POS))   /**< POLARITY Field Shifted Position */
#define MXC_F_TMR_CTRL_ENABLE0_POS                          12                                                          /**< ENABLE0 Field Position */
#define MXC_F_TMR_CTRL_ENABLE0                              ((uint32_t)(0x00000001UL << MXC_F_TMR_CTRL_ENABLE0_POS))    /**< ENABLE0 Field Shifted Position */
#define MXC_F_TMR_CTRL_ENABLE1_POS                          13                                                          /**< ENABLE1 Field Position */
#define MXC_F_TMR_CTRL_ENABLE1                              ((uint32_t)(0x00000001UL << MXC_F_TMR_CTRL_ENABLE1_POS))    /**< ENABLE1 Field Shifted Position */
/**@} end of group TMR_CTRL */

/**
 * @ingroup  tmr_registers
 * @defgroup TMR_COUNT16_0_Register TMR_COUNT16_0 Register
 * @brief    Field Positions and Bit Masks for the TMR_COUNT16_0 register. This field indicates the current count value of the <b> 16-bit Timer 0 </b> instance.
 * @{
 */  
#define MXC_F_TMR_COUNT16_0_VALUE_POS                       0                                                           /**< VALUE Field Position for the 16-bit timer 0 when the Timer is set to 2 16-bit timers, TMR2X16 is 1. */
#define MXC_F_TMR_COUNT16_0_VALUE                           ((uint32_t)(0x0000FFFFUL << MXC_F_TMR_COUNT16_0_VALUE_POS)) /**< VALUE Field Mask for the 16-bit timer 0 when the Timer is set to 2 16-bit timers, TMR2X16 is 1. */
/**@} end of group TMR_COUNT16_0 */

/**
 * @ingroup  tmr_registers
 * @defgroup TMR_TERM_CNT16_0_Register TMR_TERM_CNT16_0 Register
 * @brief    Field Positions and Bit Masks for the TMR_TERM_CNT16_0 register. This field indicates the termination count value for the <b> 16-bit Timer 0 </b> instance if the Timer is set to 2 16-bit Timers.
 * @{
 */ 
#define MXC_F_TMR_TERM_CNT16_0_TERM_COUNT_POS               0                                                                   /**< TERM_COUNT Field Position for the 16-bit timer 0 when the Timer is set to 2 16-bit timers, TMR2X16 is 1. */
#define MXC_F_TMR_TERM_CNT16_0_TERM_COUNT                   ((uint32_t)(0x0000FFFFUL << MXC_F_TMR_TERM_CNT16_0_TERM_COUNT_POS)) /**< TERM_COUNT Field Mask for the 16-bit timer 0 when the Timer is set to 2 16-bit timers, TMR2X16 is 1. */
/**@} end of group TMR_TERM_CNT16_0 */

/**
 * @ingroup  tmr_registers
 * @defgroup TMR_COUNT16_1__Register _TMR_COUNT16_1_ Register
 * @brief    Field Positions and Bit Masks for the _TMR_COUNT16_1_ register. This field indicates the current count value of the <b> 16-bit Timer 0 </b> instance.
 * @{
 */ 
#define MXC_F_TMR_COUNT16_1_VALUE_POS                       0                                                           /**< VALUE Field Position for the 16-bit timer 1 when the Timer is set to 2 16-bit timers, TMR2X16 is 1. */
#define MXC_F_TMR_COUNT16_1_VALUE                           ((uint32_t)(0x0000FFFFUL << MXC_F_TMR_COUNT16_1_VALUE_POS)) /**< VALUE Field Mask for the 16-bit timer 1 when the Timer is set to 2 16-bit timers, TMR2X16 is 1. */
/**@} end of group TMR_COUNT16_1 */

/**
 * @ingroup  tmr_registers
 * @defgroup TMR_TERM_CNT16_1_Register TMR_TERM_CNT16_1 Register
 * @brief    Field Positions and Bit Masks for the TMR_TERM_CNT16_1 register. This field indicates the termination count value for the <b> 16-bit Timer 1 </b> instance if the Timer is set to 2 16-bit Timers.
 * @{
 */ 
#define MXC_F_TMR_TERM_CNT16_1_TERM_COUNT_POS               0                                                                   /**< TERM_COUNT Field Position for the 16-bit timer 1 when the Timer is set to 2 16-bit timers, TMR2X16 is 1. */
#define MXC_F_TMR_TERM_CNT16_1_TERM_COUNT                   ((uint32_t)(0x0000FFFFUL << MXC_F_TMR_TERM_CNT16_1_TERM_COUNT_POS)) /**< TERM_COUNT Field Mask for the 16-bit timer 1 when the Timer is set to 2 16-bit timers, TMR2X16 is 1. */
/**@} end of group TMR_TERM_CNT16_1 */

/**
 * @ingroup  tmr_registers
 * @defgroup TMR_INTFL_Register TMR_INTFL Register
 * @brief    Field Positions and Bit Masks for the TMR_INTFL register. This register includes the interrupt flags for both <b> 16-bit Timer 0 and 16-bit Timer 1</b>.
 * @{
 */ 
#define MXC_F_TMR_INTFL_TIMER0_POS                          0                                                         /**< TIMER0 Interrupt Flag Field Position */
#define MXC_F_TMR_INTFL_TIMER0                              ((uint32_t)(0x00000001UL << MXC_F_TMR_INTFL_TIMER0_POS))  /**< TIMER0 Interrupt Flag Shifted Field */
#define MXC_F_TMR_INTFL_TIMER1_POS                          1                                                         /**< TIMER1 Interrupt Flag Field Position */
#define MXC_F_TMR_INTFL_TIMER1                              ((uint32_t)(0x00000001UL << MXC_F_TMR_INTFL_TIMER1_POS))  /**< TIMER1 Interrupt Flag Shifted Field */
/**@} end of group TMR_INTFL */

/**
 * @ingroup  tmr_registers
 * @defgroup TMR_INTEN_Register TMR_INTEN Register
 * @brief    Field Positions and Bit Masks for the TMR_INTEN register. This register includes the interrupt enable bits for both <b> 16-bit Timer 0 and 16-bit Timer 1</b>.
 * @{
 */
#define MXC_F_TMR_INTEN_TIMER0_POS                          0                                                         /**< TIMER0 Interrupt Enable Field Position */
#define MXC_F_TMR_INTEN_TIMER0                              ((uint32_t)(0x00000001UL << MXC_F_TMR_INTEN_TIMER0_POS))  /**< TIMER0 Interrupt Enable Shifted Field */
#define MXC_F_TMR_INTEN_TIMER1_POS                          1                                                         /**< TIMER1 Interrupt Enable Field Position */
#define MXC_F_TMR_INTEN_TIMER1                              ((uint32_t)(0x00000001UL << MXC_F_TMR_INTEN_TIMER1_POS))  /**< TIMER1 Interrupt Enable Shifted Field */
/**@} end of group TMR_INTEN */



/*
   Field values and shifted values for module TMR.
*/
/**
 * @ingroup  TMR_CTRL_Register
 * @defgroup TMR_CTRL_field_values TMR_CTRL Field and Shifted Field Values
 * @brief    Field values and Shifted Field values for the TMR_CTRL register. Shifted field values are field values shifted to the loacation of the field in the register.  
 */
/**
 * @ingroup TMR_CTRL_field_values
 * @defgroup TMR_CTRL_MODE_Field Mode Field for 32-bit Timer Operation.
 * @brief This field is used to select the timer mode for a 32-bit timer. 
 * @details The mode field is used to set the 32-bit timer instance to one of the supported modes, e.g. 1-Shot, Continuous, etc. 
 * @note If the 32-bit timer is set to operate as 2 16-bit timers, see @ref TMR_CTRL_MODE_16_Field.
 * @{
 */
#define MXC_V_TMR_CTRL_MODE_ONE_SHOT                                            ((uint32_t)(0x00000000UL))    /**< Field value to set a 32-bit Timer to 1-Shot Timer mode. */
#define MXC_V_TMR_CTRL_MODE_CONTINUOUS                                          ((uint32_t)(0x00000001UL))    /**< Field value to set a 32-bit Timer to continuous mode. */
#define MXC_V_TMR_CTRL_MODE_COUNTER                                             ((uint32_t)(0x00000002UL))    /**< Field value to set a 32-bit Timer to counter mode. */
#define MXC_V_TMR_CTRL_MODE_PWM                                                 ((uint32_t)(0x00000003UL))    /**< Field value to set a 32-bit Timer to pulse-width mode. */
#define MXC_V_TMR_CTRL_MODE_CAPTURE                                             ((uint32_t)(0x00000004UL))    /**< Field value to set a 32-bit Timer to capture mode. */
#define MXC_V_TMR_CTRL_MODE_COMPARE                                             ((uint32_t)(0x00000005UL))    /**< Field value to set a 32-bit Timer to compare mode. */
#define MXC_V_TMR_CTRL_MODE_GATED                                               ((uint32_t)(0x00000006UL))    /**< Field value to set a 32-bit Timer to gated mode. */
#define MXC_V_TMR_CTRL_MODE_MEASURE                                             ((uint32_t)(0x00000007UL))    /**< Field value to set a 32-bit Timer to measurement mode. */

#define MXC_S_TMR_CTRL_MODE_ONE_SHOT                                            ((uint32_t)(MXC_V_TMR_CTRL_MODE_ONE_SHOT    << MXC_F_TMR_CTRL_MODE_POS))    /**< Shifted Field value to set a 32-bit Timer to 1-Shot Timer mode. */
#define MXC_S_TMR_CTRL_MODE_CONTINUOUS                                          ((uint32_t)(MXC_V_TMR_CTRL_MODE_CONTINUOUS  << MXC_F_TMR_CTRL_MODE_POS))    /**< Shifted Field value to set a 32-bit Timer to continuous mode. */
#define MXC_S_TMR_CTRL_MODE_COUNTER                                             ((uint32_t)(MXC_V_TMR_CTRL_MODE_COUNTER     << MXC_F_TMR_CTRL_MODE_POS))    /**< Shifted Field value to set a 32-bit Timer to counter mode. */
#define MXC_S_TMR_CTRL_MODE_PWM                                                 ((uint32_t)(MXC_V_TMR_CTRL_MODE_PWM         << MXC_F_TMR_CTRL_MODE_POS))    /**< Shifted Field value to set a 32-bit Timer to pulse-width mode. */
#define MXC_S_TMR_CTRL_MODE_CAPTURE                                             ((uint32_t)(MXC_V_TMR_CTRL_MODE_CAPTURE     << MXC_F_TMR_CTRL_MODE_POS))    /**< Shifted Field value to set a 32-bit Timer to capture mode. */
#define MXC_S_TMR_CTRL_MODE_COMPARE                                             ((uint32_t)(MXC_V_TMR_CTRL_MODE_COMPARE     << MXC_F_TMR_CTRL_MODE_POS))    /**< Shifted Field value to set a 32-bit Timer to compare mode. */
#define MXC_S_TMR_CTRL_MODE_GATED                                               ((uint32_t)(MXC_V_TMR_CTRL_MODE_GATED       << MXC_F_TMR_CTRL_MODE_POS))    /**< Shifted Field value to set a 32-bit Timer to gated mode. */
#define MXC_S_TMR_CTRL_MODE_MEASURE                                             ((uint32_t)(MXC_V_TMR_CTRL_MODE_MEASURE     << MXC_F_TMR_CTRL_MODE_POS))    /**< Shifted Field value to set a 32-bit Timer to measurement mode. */
/**@} end of group TMR_CTRL_MODE_Field */
/**
 * @ingroup TMR_CTRL_field_values
 * @defgroup TMR_CTRL_MODE_16_Field 16-bit Timer Mode Field and Shifted Field Values.
 * @brief This field is used to select the timer mode when the timer is set to a dual 16-bit timer. The mode field is used to set the 16-bit timer instance to one of the supported modes, e.g. 1-Shot, Continuous, etc. 
 * @{
 */
#define MXC_F_TMR_CTRL_MODE_16_0_POS     0
#define MXC_F_TMR_CTRL_MODE_16_0         ((uint32_t)(0x00000001UL << MXC_F_TMR_CTRL_MODE_16_0_POS))

#define MXC_F_TMR_CTRL_MODE_16_1_POS     1
#define MXC_F_TMR_CTRL_MODE_16_1         ((uint32_t)(0x00000001UL << MXC_F_TMR_CTRL_MODE_16_1_POS))
/**@} end of group TMR_CTRL_MODE_16_Field */

/**
 * @ingroup TMR_CTRL_field_values
 * @defgroup TMR_CTRL_PRESCALE_Field Prescale Divide Selection Field and Shifted Field Values.
 * @brief Timer Clock Prescaler divide values and shifted values. The Prescale Divide field is used to scale the timer instance peripheral clock by the specified value. 
 * @{
 */
#define MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_1                                     ((uint32_t)(0x00000000UL))    /**< Field value to divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{0}= 1 \f$       */
#define MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_2                                     ((uint32_t)(0x00000001UL))    /**< Field value to divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{1}= 2 \f$       */
#define MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_4                                     ((uint32_t)(0x00000002UL))    /**< Field value to divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{2}= 4 \f$       */
#define MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_8                                     ((uint32_t)(0x00000003UL))    /**< Field value to divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{3}= 8 \f$       */
#define MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_16                                    ((uint32_t)(0x00000004UL))    /**< Field value to divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{4}= 16\f$       */
#define MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_32                                    ((uint32_t)(0x00000005UL))    /**< Field value to divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{5}= 32 \f$      */
#define MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_64                                    ((uint32_t)(0x00000006UL))    /**< Field value to divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{6}= 64 \f$      */
#define MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_128                                   ((uint32_t)(0x00000007UL))    /**< Field value to divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{7}= 128  \f$    */
#define MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_256                                   ((uint32_t)(0x00000008UL))    /**< Field value to divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{8}= 256  \f$    */
#define MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_512                                   ((uint32_t)(0x00000009UL))    /**< Field value to divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{9}= 512  \f$    */
#define MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_1024                                  ((uint32_t)(0x0000000AUL))    /**< Field value to divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{10} = 1024 \f$  */
#define MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_2048                                  ((uint32_t)(0x0000000BUL))    /**< Field value to divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{11} = 2048 \f$  */
#define MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_4096                                  ((uint32_t)(0x0000000CUL))    /**< Field value to divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{12} = 4096 \f$  */

#define MXC_S_TMR_CTRL_PRESCALE_DIVIDE_BY_1                                     ((uint32_t)(MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_1      << MXC_F_TMR_CTRL_PRESCALE_POS))    /**< Shifted field value to divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{0}= 1  \f$     */
#define MXC_S_TMR_CTRL_PRESCALE_DIVIDE_BY_2                                     ((uint32_t)(MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_2      << MXC_F_TMR_CTRL_PRESCALE_POS))    /**< Shifted field value to divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{1}= 2  \f$     */
#define MXC_S_TMR_CTRL_PRESCALE_DIVIDE_BY_4                                     ((uint32_t)(MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_4      << MXC_F_TMR_CTRL_PRESCALE_POS))    /**< Shifted field value to divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{2}= 4  \f$     */
#define MXC_S_TMR_CTRL_PRESCALE_DIVIDE_BY_8                                     ((uint32_t)(MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_8      << MXC_F_TMR_CTRL_PRESCALE_POS))    /**< Shifted field value to divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{3}= 8  \f$     */
#define MXC_S_TMR_CTRL_PRESCALE_DIVIDE_BY_16                                    ((uint32_t)(MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_16     << MXC_F_TMR_CTRL_PRESCALE_POS))    /**< Shifted field value to divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{4}= 16 \f$     */
#define MXC_S_TMR_CTRL_PRESCALE_DIVIDE_BY_32                                    ((uint32_t)(MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_32     << MXC_F_TMR_CTRL_PRESCALE_POS))    /**< Shifted field value to divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{5}= 32 \f$     */
#define MXC_S_TMR_CTRL_PRESCALE_DIVIDE_BY_64                                    ((uint32_t)(MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_64     << MXC_F_TMR_CTRL_PRESCALE_POS))    /**< Shifted field value to divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{6}= 64 \f$     */
#define MXC_S_TMR_CTRL_PRESCALE_DIVIDE_BY_128                                   ((uint32_t)(MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_128    << MXC_F_TMR_CTRL_PRESCALE_POS))    /**< Shifted field value to divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{7}= 128 \f$    */
#define MXC_S_TMR_CTRL_PRESCALE_DIVIDE_BY_256                                   ((uint32_t)(MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_256    << MXC_F_TMR_CTRL_PRESCALE_POS))    /**< Shifted field value to divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{8}= 256 \f$    */
#define MXC_S_TMR_CTRL_PRESCALE_DIVIDE_BY_512                                   ((uint32_t)(MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_512    << MXC_F_TMR_CTRL_PRESCALE_POS))    /**< Shifted field value to divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{9}= 512 \f$    */
#define MXC_S_TMR_CTRL_PRESCALE_DIVIDE_BY_1024                                  ((uint32_t)(MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_1024   << MXC_F_TMR_CTRL_PRESCALE_POS))    /**< Shifted field value to divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{10} = 1024 \f$ */
#define MXC_S_TMR_CTRL_PRESCALE_DIVIDE_BY_2048                                  ((uint32_t)(MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_2048   << MXC_F_TMR_CTRL_PRESCALE_POS))    /**< Shifted field value to divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{11} = 2048 \f$ */
#define MXC_S_TMR_CTRL_PRESCALE_DIVIDE_BY_4096                                  ((uint32_t)(MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_4096   << MXC_F_TMR_CTRL_PRESCALE_POS))    /**< Shifted field value to divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{12} = 4096 \f$ */
/**@} end of group TMR_CTRL_PRESCALE_Field */


/*
 *  These two 1-bit fields replace the standard 3-bit mode field when the associated TMR module
 *  is in dual 16-bit timer mode.
 */



#ifdef __cplusplus
}
#endif

#endif   /* _MXC_TMR_REGS_H_ */

