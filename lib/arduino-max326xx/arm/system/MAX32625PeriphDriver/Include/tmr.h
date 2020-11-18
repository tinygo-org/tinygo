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
 * $Date: 2016-03-21 14:44:55 -0500 (Mon, 21 Mar 2016) $
 * $Revision: 22017 $
 *
 ******************************************************************************/

/**
 * @file  tmr.h
 * @addtogroup timer Timer
 * @{
 * @brief This is the high level API for the general purpose timer module
 *
 */

#ifndef _TIMER_H
#define _TIMER_H

#include "mxc_config.h"
#include "tmr_regs.h"
#include "mxc_sys.h"

#ifdef __cplusplus
extern "C" {
#endif

///@brief Define values for units of time
typedef enum {
    /** nanosecond */
    TMR_UNIT_NANOSEC = 0,
    /** microsecond */
    TMR_UNIT_MICROSEC,
    /** millisecond */
    TMR_UNIT_MILLISEC,
    /** second */
    TMR_UNIT_SEC,
} tmr_unit_t;

///@brief Defines timer modes for 32-bit timers
typedef enum {
    TMR32_MODE_ONE_SHOT   = MXC_V_TMR_CTRL_MODE_ONE_SHOT,
    TMR32_MODE_CONTINUOUS = MXC_V_TMR_CTRL_MODE_CONTINUOUS,
    TMR32_MODE_COUNTER    = MXC_V_TMR_CTRL_MODE_COUNTER,
    TMR32_MODE_PWM        = MXC_V_TMR_CTRL_MODE_PWM,
    TMR32_MODE_CAPTURE    = MXC_V_TMR_CTRL_MODE_CAPTURE,
    TMR32_MODE_COMPARE    = MXC_V_TMR_CTRL_MODE_COMPARE,
    TMR32_MODE_GATED      = MXC_V_TMR_CTRL_MODE_GATED,
    TMR32_MODE_MEASURE    = MXC_V_TMR_CTRL_MODE_MEASURE
} tmr32_mode_t;

///@brief Defines timer modes for 16-bit timers
///@note 16-bit times only support One Shot and Continuous timers
typedef enum {
    TMR16_MODE_ONE_SHOT   = MXC_V_TMR_CTRL_MODE_ONE_SHOT,
    TMR16_MODE_CONTINUOUS = MXC_V_TMR_CTRL_MODE_CONTINUOUS
} tmr16_mode_t;

/// @brief Defines prescaler clock divider
typedef enum {
    // divide input clock by 2^0 = 1
    TMR_PRESCALE_DIV_2_0 = MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_1,
    // divide input clock by 2^1 = 2
    TMR_PRESCALE_DIV_2_1 = MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_2,
    // divide input clock by 2^2 = 4
    TMR_PRESCALE_DIV_2_2 = MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_4,
    // divide input clock by 2^3 = 8
    TMR_PRESCALE_DIV_2_3 = MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_8,
    // divide input clock by 2^4 = 16
    TMR_PRESCALE_DIV_2_4 = MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_16,
    // divide input clock by 2^5 = 32
    TMR_PRESCALE_DIV_2_5 = MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_32,
    // divide input clock by 2^6 = 64
    TMR_PRESCALE_DIV_2_6 = MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_64,
    // divide input clock by 2^7 = 128
    TMR_PRESCALE_DIV_2_7 = MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_128,
    // divide input clock by 2^8 = 256
    TMR_PRESCALE_DIV_2_8 = MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_256,
    // divide input clock by 2^9 = 512
    TMR_PRESCALE_DIV_2_9 = MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_512,
    // divide input clock by 2^10 = 1024
    TMR_PRESCALE_DIV_2_10 = MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_1024,
    // divide input clock by 2^11 = 2048
    TMR_PRESCALE_DIV_2_11 = MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_2048,
    // divide input clock by 2^12 = 4096
    TMR_PRESCALE_DIV_2_12 = MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_4096,
} tmr_prescale_t;

/// @brief Defines the polarity bit for pwm timer
typedef enum {
    TMR_POLARITY_UNUSED = 0,

    TMR_POLARITY_INIT_LOW = 0,      ///GPIO initial output level = low
    TMR_POLARITY_INIT_HIGH = 1,     ///GPIO initial output level = high

    TMR_POLARITY_RISING_EDGE = 0,   ///timer trigger on GPIO rising edge
    TMR_POLARITY_FALLING_EDGE = 1,  ///timer trigger on GPIO falling edge
} tmr_polarity_t;

/// @brief Defines the polarity bit for pwm timer
typedef enum {
    TMR_PWM_INVERTED = 0,   ///duty cycle = low pulse
    TMR_PWM_NONINVERTED,    ///duty cycle = high pulse
} tmr_pwm_polarity_t;

/// @brief 32bit Timer Configurations (non-PWM)
typedef struct {
    tmr32_mode_t mode;            /// Desired timer mode
    tmr_polarity_t  polarity;   /// Polarity for GPIO
    uint32_t compareCount;      /// Ticks to stop, reset back to 1, or compare timer
} tmr32_cfg_t;

/// @brief PWM Mode Configurations
typedef struct {
    tmr_pwm_polarity_t polarity;    /// PWM polarity
    uint32_t periodCount;           /// PWM period in timer counts
    uint32_t dutyCount;             /// PWM duty in timer counts
} tmr32_cfg_pwm_t;

/// @brief 16bit Timer Configurations
typedef struct {
    tmr16_mode_t mode;      /// Desired timer mode (only supports one shot or continuous timers)
    uint16_t compareCount;  /// Ticks to stop or reset timer
} tmr16_cfg_t;



/**
 * @brief   Initializes the timer to a known state.
 * @details This function initializes the timer to a known state and saves the
 *          prescaler. The timer will be left disabled. TMR_Init should be called
 *          before TMR_Config.
 *
 * @param   tmr         timer module to operate on
 * @param   prescale    clock divider for the timer clock
 * @param   cfg         pointer to timer system GPIO configuration
 *                      (can be NULL if not using GPIO timer input/output functions)
 *
 * @returns #E_NO_ERROR if everything is successful, error if unsuccessful.
 */
int TMR_Init(mxc_tmr_regs_t *tmr, tmr_prescale_t prescale, const sys_cfg_tmr_t *sysCfg);

/**
 * @brief   Configures the timer in the specified mode.
 * @details The parameters in config structure must be set before calling this function.
 *          This function should be used for configuring the timer in all 32 bit modes other than PWM.
 * @note    The timer cannot be running when this function is called
 *
 * @param   tmr         timer module to operate on
 * @param   config      pointer to timer configuration
 *
 */
void TMR32_Config(mxc_tmr_regs_t *tmr, const tmr32_cfg_t *config);

/**
 * @brief   Configures the timer in PWM mode.
 * @details The parameters in config structure must be set before calling this function.
 *          This function should be used for configuring the timer in PWM mode only.
 * @note    The timer cannot be running when this function is called
 *
 * @param   tmr         timer module to operate on
 * @param   config      pointer to timer configuration
 *
 */
void TMR32_PWMConfig(mxc_tmr_regs_t *tmr, const tmr32_cfg_pwm_t *config);

/**
 * @brief   Configures the timer in the specified mode.
 * @details The parameters in config structure must be set before calling this function.
 *          This function should be used for configuring the timer in all 16 bit modes.
 * @note    The timer cannot be running when this function is called
 *
 * @param   tmr         timer module to operate on
 * @param   index       selects which 16 bit timer (0 = 16_0 or 1 = 16_1)
 * @param   config      pointer to timer configuration
 *
 */
void TMR16_Config(mxc_tmr_regs_t *tmr, uint8_t index, const tmr16_cfg_t *config);

/**
 * @brief   Starts the specified timer.
 *
 * @param   tmr         timer module to operate on
 *
 */
void TMR32_Start(mxc_tmr_regs_t *tmr);

/**
 * @brief   Starts the specified timer.
 *
 * @param   tmr     timer module to operate on
 * @param   index   selects which 16 bit timer (0 = 16_0 or 1 = 16_1)
 *
 */
void TMR16_Start(mxc_tmr_regs_t *tmr, uint8_t index);

/**
 * @brief   Stops the specified 32 bit timer.
 * @details All other timer states are left unchanged.
 *
 * @param   tmr     timer module to operate on
 *
 */
__STATIC_INLINE void TMR32_Stop(mxc_tmr_regs_t *tmr)
{
    tmr->ctrl &= ~MXC_F_TMR_CTRL_ENABLE0;
}

/**
 * @brief   Stop the specified 16 bit timer.
 * @details All other timer states are left unchanged.
 *
 * @param   tmr     timer module to operate on
 * @param   index   selects which 16 bit timer (0 = 16_0 or 1 = 16_1)
 *
 */
__STATIC_INLINE void TMR16_Stop(mxc_tmr_regs_t *tmr, uint8_t index)
{
    if (index)
        tmr->ctrl &= ~MXC_F_TMR_CTRL_ENABLE1;
    else
        tmr->ctrl &= ~MXC_F_TMR_CTRL_ENABLE0;
}

/**
 * @brief   Determines if the timer is running
 *
 * @param   tmr     timer module to operate on
 *
 * @return  0 = timer is off, non-zero = timer is on
 */
__STATIC_INLINE uint32_t TMR32_IsActive(mxc_tmr_regs_t *tmr)
{
    return (tmr->ctrl & MXC_F_TMR_CTRL_ENABLE0);
}

/**
 * @brief   Determines if the timer is running
 *
 * @param   tmr     timer module to operate on
 * @param   index   selects which 16 bit timer (0 = 16_0 or 1 = 16_1)
 *
 * @return  0 = timer is off, non-zero = timer is on
 */
__STATIC_INLINE uint32_t TMR16_IsActive(mxc_tmr_regs_t *tmr, uint8_t index)
{
    if (index)
        return (tmr->ctrl & MXC_F_TMR_CTRL_ENABLE1);
    else
        return (tmr->ctrl & MXC_F_TMR_CTRL_ENABLE0);
}

/**
 * @brief   Enables the timer's interrupt
 *
 * @param   tmr     timer module to operate on
 *
 */
__STATIC_INLINE void TMR32_EnableINT(mxc_tmr_regs_t *tmr)
{
    tmr->inten |= MXC_F_TMR_INTEN_TIMER0;
}

/**
 * @brief   Enables the timer's interrupt
 *
 * @param   tmr     timer module to operate on
 * @param   index   selects which 16 bit timer (0 = 16_0 or 1 = 16_1)
 *
 */
__STATIC_INLINE void TMR16_EnableINT(mxc_tmr_regs_t *tmr, uint8_t index)
{
    if (index)
        tmr->inten |= MXC_F_TMR_INTEN_TIMER1;
    else
        tmr->inten |= MXC_F_TMR_INTEN_TIMER0;
}

/**
 * @brief   Disables the timer's interrupt
 *
 * @param   tmr     timer module to operate on
 *
 */
__STATIC_INLINE void TMR32_DisableINT(mxc_tmr_regs_t *tmr)
{
    tmr->inten &= ~MXC_F_TMR_INTEN_TIMER0;
}

/**
 * @brief   Disables the timer's interrupt
 *
 * @param   tmr     timer module to operate on
 * @param   index   selects which 16 bit timer (0 = 16_0 or 1 = 16_1)
 *
 */
__STATIC_INLINE void TMR16_DisableINT(mxc_tmr_regs_t *tmr, uint8_t index)
{
    if (index)
        tmr->inten &= ~MXC_F_TMR_INTEN_TIMER1;
    else
        tmr->inten &= ~MXC_F_TMR_INTEN_TIMER0;
}

/**
 * @brief   Gets the timer's interrupt flag
 *
 * @param   tmr     timer module to operate on
 *
 * @return  0 = flag not set, non-zero = flag is set
 */
__STATIC_INLINE uint32_t TMR32_GetFlag(mxc_tmr_regs_t *tmr)
{
    return (tmr->intfl & MXC_F_TMR_INTFL_TIMER0);
}

/**
 * @brief   Gets the timer's interrupt flag
 *
 * @param   tmr     timer module to operate on
 * @param   index   selects which 16 bit timer (0 = 16_0 or 1 = 16_1)
 *
 * @return  0 = flag not set, non-zero = flag is set
 */
__STATIC_INLINE uint32_t TMR16_GetFlag(mxc_tmr_regs_t *tmr, uint8_t index)
{
    if (index)
        return (tmr->intfl & MXC_F_TMR_INTFL_TIMER1);
    else
        return (tmr->intfl & MXC_F_TMR_INTFL_TIMER0);
}

/**
 * @brief   Clears the timer interrupt flag
 *
 * @param   tmr     timer module to operate on
 *
 */
__STATIC_INLINE void TMR32_ClearFlag(mxc_tmr_regs_t *tmr)
{
    tmr->intfl = MXC_F_TMR_INTFL_TIMER0;
}

/**
 * @brief   Clears the timer interrupt flag for the specified index
 *
 * @param   tmr     timer module to operate on
 * @param   index   selects which 16 bit timer (0 = 16_0 or 1 = 16_1)
 *
 */
__STATIC_INLINE void TMR16_ClearFlag(mxc_tmr_regs_t *tmr, uint8_t index)
{
    if (index)
        tmr->intfl = MXC_F_TMR_INTFL_TIMER1;
    else
        tmr->intfl = MXC_F_TMR_INTFL_TIMER0;
}

/**
 * @brief   Set the current tick value to start counting from.
 *
 * @param   tmr     timer module to operate on
 * @param   count   value to set the current ticks
 *
 */
__STATIC_INLINE void TMR32_SetCount(mxc_tmr_regs_t *tmr, uint32_t count)
{
    tmr->count32 = count;
}

/**
 * @brief   Set the current tick value to start counting from.
 *
 * @param   tmr     timer module to operate on
 * @param   index   selects which 16 bit timer (0 = 16_0 or 1 = 16_1)
 * @param   count   value to set the current ticks
 *
 */
__STATIC_INLINE void TMR16_SetCount(mxc_tmr_regs_t *tmr, uint8_t index, uint16_t count)
{
    if (index)
        tmr->count16_1 = count;
    else
        tmr->count16_0 = count;
}

/**
 * @brief   Gets the most current count value.
 *
 * @param   tmr     timer module to operate on
 *
 * @return  current count value in ticks
 */
__STATIC_INLINE uint32_t TMR32_GetCount(mxc_tmr_regs_t *tmr)
{
    return (tmr->count32);
}

/**
 * @brief  Gets the most current count value.
 *
 * @param  tmr     timer module to operate on
 * @param  index   selects which 16 bit timer(0 = 16_0 or 1 = 16_1)
 *
 * @return current count value in ticks
 */
__STATIC_INLINE uint32_t TMR16_GetCount(mxc_tmr_regs_t *tmr, uint8_t index)
{
    if (index)
        return tmr->count16_1;
    else
        return tmr->count16_0;
}

/**
 * @brief   Gets the most recent capture value.
 * @details Used in Capture or Measure timer modes
 *
 * @param   tmr     timer module to operate on
 *
 * @return  capture value in ticks
 */
__STATIC_INLINE uint32_t TMR32_GetCapture(mxc_tmr_regs_t *tmr)
{
    return (tmr->pwm_cap32);
}

/**
 * @brief   Set a new compare tick value for timer
 * @details Depending on the timer mode this is the tick value to
 *          stop the timer, reset ticks to 1, or compare the timer
 *
 * @param   tmr         timer module to operate on
 * @param   count       new terminal/compare value in timer counts
 *
 */
__STATIC_INLINE void TMR32_SetCompare(mxc_tmr_regs_t *tmr, uint32_t count)
{
    tmr->term_cnt32 = count;
}

/**
 * @brief   Get compare tick value for timer
 * @param   tmr     timer module to operate on
 * @return  compare value in ticks
 *
 */
__STATIC_INLINE uint32_t TMR32_GetCompare(mxc_tmr_regs_t *tmr)
{
    return tmr->term_cnt32;
}

/**
 * @brief   Set a new compare tick value for timer
 * @details Depending on the timer mode this is the tick value to
 *          stop the timer, reset ticks to 1, or compare the timer
 *
 * @param   tmr         timer module to operate on
 * @param   index       selects which 16 bit timer (0 = 16_0 or 1 = 16_1)
 * @param   count       new terminal/compare value in timer counts
 *
 */
__STATIC_INLINE void TMR16_SetCompare(mxc_tmr_regs_t *tmr, uint8_t index, uint16_t count)
{
    if (index)
        tmr->term_cnt16_1 = count;
    else
        tmr->term_cnt16_0 = count;
}

/**
 * @brief   Get compare tick value for timer
 * @param   tmr         timer module to operate on
 * @param   index       selects which 16 bit timer (0 = 16_0 or 1 = 16_1)
 * @return  compare value in ticks
 *
 */
__STATIC_INLINE uint32_t TMR16_GetCompare(mxc_tmr_regs_t *tmr, uint8_t index)
{
    if (index)
        return tmr->term_cnt16_1;
    return tmr->term_cnt16_0;
}

/**
 * @brief   Returns the prescale value used by the timer
 *
 * @param   tmr         timer module to operate on
 *
 * @returns prescaler
 */
uint32_t TMR_GetPrescaler(mxc_tmr_regs_t *tmr);

/**
 * @brief   Set a new duty cycle when the timer is used in PWM mode.
 *
 * @param   tmr         timer module to operate on
 * @param   dutyCount   duty cycle value in timer counts
 *
 */
__STATIC_INLINE void TMR32_SetDuty(mxc_tmr_regs_t *tmr, uint32_t dutyCount)
{
    tmr->pwm_cap32 = dutyCount;
}

/**
 * @brief  Set a new duty cycle when the timer is used in PWM mode.
 *
 * @param  tmr         timer module to operate on
 * @param  dutyPercent duty cycle value in percent (0 to 100%)
 *
 */
__STATIC_INLINE void TMR32_SetDutyPer(mxc_tmr_regs_t *tmr, uint32_t dutyPercent)
{
    uint32_t periodCount = tmr->term_cnt32;
    tmr->pwm_cap32 = ((uint64_t)periodCount * dutyPercent) / 100;
}

/**
 * @brief   Set a new period value for PWM timer
 *
 * @param   tmr         timer module to operate on
 * @param   count       new period value in timer counts
 *
 */
__STATIC_INLINE void TMR32_SetPeriod(mxc_tmr_regs_t *tmr, uint32_t period)
{
    tmr->term_cnt32 = period;
}

/**
 * @brief Converts frequency and duty cycle % to period and duty ticks
 * @note TMR_Init should be called before this function to set the prescaler
 *
 * @param   tmr         timer module to operate on
 * @param   dutyPercent duty cycle in percent (0 to 100%)
 * @param   freq        frequency of pwm signal in Hz
 * @param   dutyTicks   calculated duty cycle in ticks
 * @param   periodTicks calculated period in ticks
 *
 * @returns #E_NO_ERROR if everything is successful, error if unsuccessful.
 *
 */
int TMR32_GetPWMTicks(mxc_tmr_regs_t *tmr, uint8_t dutyPercent, uint32_t freq, uint32_t *dutyTicks, uint32_t *periodTicks);

/**
 * @brief Converts a time and units to a number of ticks for the 32-bit timer.
 * @note TMR_Init should be called before this function to set the prescaler
 *
 * @param  tmr         timer module to operate on
 * @param  time        time value.
 * @param  unit        time units.
 * @param  ticks       calculated number of ticks.
 *
 * @returns #E_NO_ERROR if everything is successful, error if unsuccessful.
 */
int TMR32_TimeToTicks(mxc_tmr_regs_t *tmr, uint32_t time, tmr_unit_t units, uint32_t *ticks);

/**
 * @brief Converts a time and units to a number of ticks for the 16-bit timer.
 * @note TMR_Init should be called before this function to set the prescaler
 *
 * @param  tmr         timer module to operate on
 * @param  time        time value.
 * @param  unit        time units.
 * @param  ticks       calculated number of ticks.
 *
 * @returns #E_NO_ERROR if everything is successful, error if unsuccessful.
 */
int TMR16_TimeToTicks(mxc_tmr_regs_t *tmr, uint32_t time, tmr_unit_t units, uint16_t *ticks);

/**
 * @brief Converts a number of ticks to a time and units for the timer.
 * @note TMR_Init should be called before this function to set the prescaler
 *
 * @param  tmr         timer module to operate on
 * @param  ticks       number of ticks.
 * @param  time        calculated time value.
 * @param  units       calculated time units.
 *
 * @returns #E_NO_ERROR if everything is successful, error if unsuccessful.
 */
int TMR_TicksToTime(mxc_tmr_regs_t *tmr, uint32_t ticks, uint32_t *time, tmr_unit_t *units);

/** @} */

#ifdef __cplusplus
}
#endif

#endif /* _TIMER_H */
