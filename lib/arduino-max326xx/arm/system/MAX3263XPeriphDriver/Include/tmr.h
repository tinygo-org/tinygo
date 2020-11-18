/**
 * @file
 * @brief Timer0 & Timer1 32/16-Bit Peripheral Driver.
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
#ifndef _TIMER_H
#define _TIMER_H

/* **** Includes **** */
#include "mxc_config.h"
#include "tmr_regs.h"
#include "mxc_sys.h"

/* **** Extern CPP **** */
#ifdef __cplusplus
extern "C" {
#endif

/**
 * @ingroup periphlibs
 * @defgroup tmr Timers
 * @brief 32/16-bit Timers
 * @{
 */

/**
 * Enumeration type for units of time.
 */
typedef enum {
    TMR_UNIT_NANOSEC = 0,       /**< Nanosecond Unit Indicator. */
    TMR_UNIT_MICROSEC,          /**< Microsecond Unit Indicator. */
    TMR_UNIT_MILLISEC,          /**< Millisecond Unit Indicator. */
    TMR_UNIT_SEC,               /**< Second Unit Indicator. */
} tmr_unit_t;

/**
 * Enumeration type to select the 32-bit Timer Mode.
 */
typedef enum {
    TMR32_MODE_ONE_SHOT   = MXC_V_TMR_CTRL_MODE_ONE_SHOT,           /**< One-shot Mode            */
    TMR32_MODE_CONTINUOUS = MXC_V_TMR_CTRL_MODE_CONTINUOUS,         /**< Continuous Mode          */
    TMR32_MODE_COUNTER    = MXC_V_TMR_CTRL_MODE_COUNTER,            /**< Counter Mode                   */
    TMR32_MODE_PWM        = MXC_V_TMR_CTRL_MODE_PWM,                /**< Pulse Width Modulation Mode    */
    TMR32_MODE_CAPTURE    = MXC_V_TMR_CTRL_MODE_CAPTURE,            /**< Capture Mode                   */
    TMR32_MODE_COMPARE    = MXC_V_TMR_CTRL_MODE_COMPARE,            /**< Compare Mode                   */
    TMR32_MODE_GATED      = MXC_V_TMR_CTRL_MODE_GATED,              /**< Gated Mode                     */
    TMR32_MODE_MEASURE    = MXC_V_TMR_CTRL_MODE_MEASURE             /**< Measure Mode                   */    
} tmr32_mode_t;

/**
 * Enumeration type to select a 16-bit Timer Mode.
 * @note 16-bit times only support One Shot and Continuous timers.
 */
typedef enum {
    TMR16_MODE_ONE_SHOT   = MXC_V_TMR_CTRL_MODE_ONE_SHOT,   /**< One-Shot Mode. */
    TMR16_MODE_CONTINUOUS = MXC_V_TMR_CTRL_MODE_CONTINUOUS  /**< Continuous Mode. */
} tmr16_mode_t;

/**
 * Enumeration type to select the Prescale Divider for the timer module. The prescaler 
 * divides the peripheral input clock to the timer by a selectable divisor. 
 */
typedef enum {
    TMR_PRESCALE_DIV_2_0 = MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_1,         /**< Divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{0} = 1 \f$       */
    TMR_PRESCALE_DIV_2_1 = MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_2,         /**< Divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{1} = 2 \f$       */
    TMR_PRESCALE_DIV_2_2 = MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_4,         /**< Divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{2} = 4 \f$       */
    TMR_PRESCALE_DIV_2_3 = MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_8,         /**< Divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{3} = 8 \f$       */
    TMR_PRESCALE_DIV_2_4 = MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_16,        /**< Divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{4} = 16 \f$      */
    TMR_PRESCALE_DIV_2_5 = MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_32,        /**< Divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{5} = 32 \f$      */
    TMR_PRESCALE_DIV_2_6 = MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_64,        /**< Divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{6} = 64 \f$      */
    TMR_PRESCALE_DIV_2_7 = MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_128,       /**< Divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{7} = 128 \f$     */
    TMR_PRESCALE_DIV_2_8 = MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_256,       /**< Divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{8} = 256 \f$     */
    TMR_PRESCALE_DIV_2_9 = MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_512,       /**< Divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{9} = 512 \f$     */
    TMR_PRESCALE_DIV_2_10 = MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_1024,     /**< Divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{10} = 1024 \f$   */
    TMR_PRESCALE_DIV_2_11 = MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_2048,     /**< Divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{11} = 2048 \f$   */
    TMR_PRESCALE_DIV_2_12 = MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_4096,     /**< Divide the peripheral input clock by \f$ TMR_{Prescaler} = 2^{12} = 4096 \f$   */
} tmr_prescale_t;

/**
 * Enumeration type to set the polarity bit for pwm timer.
 */
typedef enum {
    TMR_POLARITY_UNUSED = 0,        /**< @internal Unused polarity @endinternal */

    TMR_POLARITY_INIT_LOW = 0,      /**< GPIO initial output level = low    */
    TMR_POLARITY_INIT_HIGH = 1,     /**< GPIO initial output level = high   */
    
    TMR_POLARITY_RISING_EDGE = 0,   /**< timer trigger on GPIO rising edge  */
    TMR_POLARITY_FALLING_EDGE = 1,  /**< timer trigger on GPIO falling edge */
} tmr_polarity_t;

/**
 * Enumeration type to set the polarity bit for pwm timer.
 */
typedef enum {
    TMR_PWM_INVERTED = 0,   /**< duty cycle = low pulse     */
    TMR_PWM_NONINVERTED,    /**< duty cycle = high pulse    */
} tmr_pwm_polarity_t;

/**
 * Structure type for Configuring a 32-bit timer in all modes except PWM.
 */
typedef struct {
    tmr32_mode_t mode;              /**< Desired timer mode, see #tmr32_mode_t for valid modes. @note If PWM mode is desired, setting the mode to TMR32_MODE_PWM is valid. To configure PWM Mode, see #tmr32_cfg_pwm_t. */
    tmr_polarity_t  polarity;       /**< Polarity for GPIO   */
    uint32_t compareCount;          /**< Ticks to stop, reset back to 1, or compare timer    */
} tmr32_cfg_t;

/**
 * Structure type for Configuring PWM Mode for a 32-bit timer.
 */
typedef struct {
    tmr_pwm_polarity_t polarity;    /**< PWM polarity selection, see #tmr_pwm_polarity_t.   */
    uint32_t periodCount;           /**< PWM period in number of timer ticks.               */
    uint32_t dutyCount;             /**< PWM duty cycle in number of timer ticks.           */
} tmr32_cfg_pwm_t;

/**
 * Structure type for Configuring a 16-bit Timer. 
 */
typedef struct {
    tmr16_mode_t mode;      /**< 16-bit timer mode, see #tmr16_mode_t for supported modes.    */
    uint16_t compareCount;  /**< Number of timer ticks to either stop or reset the timer.     */
} tmr16_cfg_t;



/**
 * @brief      Initializes the timer to a known state.
 * @details    This function initializes the timer to a known state and saves
 *             the prescaler. The timer will be left disabled. TMR_Init should
 *             be called before TMR_Config.
 *
 * @param      tmr       Pointer to timer registers for the timer instance to modify.
 * @param      prescale  clock divider for the timer clock
 * @param      sysCfg    Pointer to the timer system GPIO configuration. If not
 *                       using GPIO for the timer instance, set this parameter
 *                       to NULL.
 *
 * @retval     #E_NO_ERROR Timer initialized successfully. 
 * @retval     Error Code  Timer was not initialized, see @ref MXC_Error_Codes.
 */
int TMR_Init(mxc_tmr_regs_t *tmr, tmr_prescale_t prescale, const sys_cfg_tmr_t *sysCfg);

/**
 * @brief      Configures the timer in the specified mode.
 * @details    The parameters in config structure must be set before calling
 *             this function. This function should be used for configuring the
 *             timer in all 32 bit modes other than PWM.
 * @note       The timer cannot be running when this function is called
 *
 * @param      tmr     Pointer to timer registers for the timer instance to modify.
 * @param      config  pointer to timer configuration
 */
void TMR32_Config(mxc_tmr_regs_t *tmr, const tmr32_cfg_t *config);

/**
 * @brief      Configures the timer in PWM mode.
 * @details    The parameters in config structure must be set before calling
 *             this function. This function should be used for configuring the
 *             timer in PWM mode only.
 * @note       The timer cannot be running when this function is called
 *
 * @param      tmr     Pointer to timer registers for the timer instance to modify.
 * @param      config  pointer to timer configuration
 */
void TMR32_PWMConfig(mxc_tmr_regs_t *tmr, const tmr32_cfg_pwm_t *config);

/**
 * @brief      Configures the timer in the specified mode.
 * @details    The parameters in config structure must be set before calling
 *             this function. This function should be used for configuring the
 *             timer in all 16 bit modes.
 * @note       The timer cannot be running when this function is called
 *
 * @param      tmr     Pointer to timer registers for the timer instance to modify.
 * @param      index   selects which 16 bit timer (0 = 16_0 or 1 = 16_1)
 * @param      config  pointer to timer configuration
 */
void TMR16_Config(mxc_tmr_regs_t *tmr, uint8_t index, const tmr16_cfg_t *config);

/**
 * @brief      Starts the specified timer.
 *
 * @param      tmr   Pointer to timer registers for the timer instance to modify.
 */
void TMR32_Start(mxc_tmr_regs_t *tmr);

/**
 * @brief      Starts the specified timer.
 *
 * @param      tmr    Pointer to timer registers for the timer instance to modify.
 * @param      index  selects which 16 bit timer (0 = 16_0 or 1 = 16_1)
 */
void TMR16_Start(mxc_tmr_regs_t *tmr, uint8_t index);

/**
 * @brief      Stops the specified 32 bit timer.
 * @details    All other timer states are left unchanged.
 *
 * @param      tmr   Pointer to timer registers for the timer instance to modify.
 */
__STATIC_INLINE void TMR32_Stop(mxc_tmr_regs_t *tmr)
{
    tmr->ctrl &= ~MXC_F_TMR_CTRL_ENABLE0;
}

/**
 * @brief      Stop the specified 16 bit timer.
 * @details    All other timer states are left unchanged.
 *
 * @param      tmr    Pointer to timer registers for the timer instance to modify.
 * @param      index  selects which 16 bit timer (0 = 16_0 or 1 = 16_1)
 */
__STATIC_INLINE void TMR16_Stop(mxc_tmr_regs_t *tmr, uint8_t index)
{
    if(index)
        tmr->ctrl &= ~MXC_F_TMR_CTRL_ENABLE1;
    else
        tmr->ctrl &= ~MXC_F_TMR_CTRL_ENABLE0;
}

/**
 * @brief      Determines if the timer is running
 *
 * @param      tmr   Pointer to timer registers for the timer instance to modify.
 *
 * @return     0 = timer is off, non-zero = timer is on
 */
__STATIC_INLINE uint32_t TMR32_IsActive(mxc_tmr_regs_t *tmr)
{
    return (tmr->ctrl & MXC_F_TMR_CTRL_ENABLE0);
}

/**
 * @brief      Determines if the timer is running
 *
 * @param      tmr    Pointer to timer registers for the timer instance to modify.
 * @param      index  selects which 16 bit timer (0 = 16_0 or 1 = 16_1)
 *
 * @return     0    Timer is off.
 * @return  Non-zero    Timer is on.
 */
__STATIC_INLINE uint32_t TMR16_IsActive(mxc_tmr_regs_t *tmr, uint8_t index)
{
    if(index)
        return (tmr->ctrl & MXC_F_TMR_CTRL_ENABLE1);
    else
        return (tmr->ctrl & MXC_F_TMR_CTRL_ENABLE0);
}

/**
 * @brief      Enables the timer's interrupt
 *
 * @param      tmr   Pointer to timer registers for the timer instance to modify.
 */
__STATIC_INLINE void TMR32_EnableINT(mxc_tmr_regs_t *tmr)
{
    tmr->inten |= MXC_F_TMR_INTEN_TIMER0;
}

/**
 * @brief      Enables the timer's interrupt
 *
 * @param      tmr    Pointer to timer registers for the timer instance to modify.
 * @param      index  selects which 16 bit timer (0 = 16_0 or 1 = 16_1)
 */
__STATIC_INLINE void TMR16_EnableINT(mxc_tmr_regs_t *tmr, uint8_t index)
{
    if(index)
        tmr->inten |= MXC_F_TMR_INTEN_TIMER1;
    else
        tmr->inten |= MXC_F_TMR_INTEN_TIMER0;
}

/**
 * @brief      Disables the timer's interrupt
 *
 * @param      tmr   Pointer to timer registers for the timer instance to modify.
 */
__STATIC_INLINE void TMR32_DisableINT(mxc_tmr_regs_t *tmr)
{
    tmr->inten &= ~MXC_F_TMR_INTEN_TIMER0;
}

/**
 * @brief      Disables the timer's interrupt
 *
 * @param      tmr    Pointer to timer registers for the timer instance to modify.
 * @param      index  selects which 16 bit timer (0 = 16_0 or 1 = 16_1)
 */
__STATIC_INLINE void TMR16_DisableINT(mxc_tmr_regs_t *tmr, uint8_t index)
{
    if(index)
        tmr->inten &= ~MXC_F_TMR_INTEN_TIMER1;
    else
        tmr->inten &= ~MXC_F_TMR_INTEN_TIMER0;
}

/**
 * @brief      Gets the timer's interrupt flag
 *
 * @param      tmr   Pointer to timer registers for the timer instance to modify.
 *
 * @return     0 = flag not set, non-zero = flag is set
 */
__STATIC_INLINE uint32_t TMR32_GetFlag(mxc_tmr_regs_t *tmr)
{
    return (tmr->intfl & MXC_F_TMR_INTFL_TIMER0);
}

/**
 * @brief      Gets the timer's interrupt flag
 *
 * @param      tmr    Pointer to timer registers for the timer instance to modify.
 * @param      index  selects which 16 bit timer (0 = 16_0 or 1 = 16_1)
 *
 * @return     0 = flag not set, non-zero = flag is set
 */
__STATIC_INLINE uint32_t TMR16_GetFlag(mxc_tmr_regs_t *tmr, uint8_t index)
{
    if(index)
        return (tmr->intfl & MXC_F_TMR_INTFL_TIMER1);
    else
        return (tmr->intfl & MXC_F_TMR_INTFL_TIMER0);
}

/**
 * @brief      Clears the timer interrupt flag
 *
 * @param      tmr   Pointer to timer registers for the timer instance to modify.
 */
__STATIC_INLINE void TMR32_ClearFlag(mxc_tmr_regs_t *tmr)
{
    tmr->intfl = MXC_F_TMR_INTFL_TIMER0;
}

/**
 * @brief      Clears the timer interrupt flag for the specified index
 *
 * @param      tmr    Pointer to timer registers for the timer instance to modify.
 * @param      index  selects which 16 bit timer (0 = 16_0 or 1 = 16_1)
 */
__STATIC_INLINE void TMR16_ClearFlag(mxc_tmr_regs_t *tmr, uint8_t index)
{
    if(index)
        tmr->intfl = MXC_F_TMR_INTFL_TIMER1;
    else
        tmr->intfl = MXC_F_TMR_INTFL_TIMER0;
}

/**
 * @brief      Set the current tick value to start counting from.
 *
 * @param      tmr    Pointer to timer registers for the timer instance to modify.
 * @param      count  value to set the current ticks
 */
__STATIC_INLINE void TMR32_SetCount(mxc_tmr_regs_t *tmr, uint32_t count)
{
    tmr->count32 = count;
}

/**
 * @brief      Set the current tick value to start counting from.
 *
 * @param      tmr    Pointer to timer registers for the timer instance to modify.
 * @param      index  selects which 16 bit timer (0 = 16_0 or 1 = 16_1)
 * @param      count  value to set the current ticks
 */
__STATIC_INLINE void TMR16_SetCount(mxc_tmr_regs_t *tmr, uint8_t index, uint16_t count)
{
    if (index)
        tmr->count16_1 = count;
    else
        tmr->count16_0 = count;
}

/**
 * @brief      Gets the most current count value.
 *
 * @param      tmr   Pointer to timer registers for the timer instance to modify.
 *
 * @return     current count value in ticks
 */
__STATIC_INLINE uint32_t TMR32_GetCount(mxc_tmr_regs_t *tmr)
{
    return (tmr->count32);
}

/**
 * @brief      Gets the most current count value.
 *
 * @param      tmr    Pointer to timer registers for the timer instance to modify.
 * @param      index  selects which 16 bit timer(0 = 16_0 or 1 = 16_1)
 *
 * @return     current count value in ticks
 */
__STATIC_INLINE uint32_t TMR16_GetCount(mxc_tmr_regs_t *tmr, uint8_t index)
{
    if(index)
        return tmr->count16_1;
    else
        return tmr->count16_0;
}

/**
 * @brief      Gets the most recent capture value.
 * @details    Used in Capture or Measure timer modes
 *
 * @param      tmr   Pointer to timer registers for the timer instance to modify.
 *
 * @return     capture value in ticks
 */
__STATIC_INLINE uint32_t TMR32_GetCapture(mxc_tmr_regs_t *tmr)
{
    return (tmr->pwm_cap32);
}

/**
 * @brief      Set a new compare tick value for timer
 * @details    Depending on the timer mode this is the tick value to stop the
 *             timer, reset ticks to 1, or compare the timer
 *
 * @param      tmr    Pointer to timer registers for the timer instance to modify.
 * @param      count  new terminal/compare value in timer counts
 */
__STATIC_INLINE void TMR32_SetCompare(mxc_tmr_regs_t *tmr, uint32_t count)
{
    tmr->term_cnt32 = count;
}

/**
 * @brief      Get compare tick value for timer
 * @param      tmr   Pointer to timer registers for the timer instance to modify.
 * @return     compare value in ticks
 */
__STATIC_INLINE uint32_t TMR32_GetCompare(mxc_tmr_regs_t *tmr)
{
    return tmr->term_cnt32;
}

/**
 * @brief      Set a new compare tick value for timer
 * @details    Depending on the timer mode this is the tick value to stop the
 *             timer, reset ticks to 1, or compare the timer
 *
 * @param      tmr    Pointer to timer registers for the timer instance to modify.
 * @param      index  selects which 16 bit timer (0 = 16_0 or 1 = 16_1)
 * @param      count  new terminal/compare value in timer counts
 */
__STATIC_INLINE void TMR16_SetCompare(mxc_tmr_regs_t *tmr, uint8_t index, uint16_t count)
{
    if (index)
        tmr->term_cnt16_1 = count;
    else
        tmr->term_cnt16_0 = count;
}

/**
 * @brief      Get compare tick value for timer
 * @param      tmr    Pointer to timer registers for the timer instance to modify.
 * @param      index  selects which 16 bit timer (0 = 16_0 or 1 = 16_1)
 * @return     compare value in ticks
 */
__STATIC_INLINE uint32_t TMR16_GetCompare(mxc_tmr_regs_t *tmr, uint8_t index)
{
    if (index)
        return tmr->term_cnt16_1;
    return tmr->term_cnt16_0;
}

/**
 * @brief      Returns the prescale value used by the timer
 *
 * @param      tmr   Pointer to timer registers for the timer instance to modify.
 *
 * @return     prescaler
 */
uint32_t TMR_GetPrescaler(mxc_tmr_regs_t *tmr);

/**
 * @brief      Set a new duty cycle when the timer is used in PWM mode.
 *
 * @param      tmr        Pointer to timer registers for the timer instance to modify.
 * @param      dutyCount  duty cycle value in timer counts
 */
__STATIC_INLINE void TMR32_SetDuty(mxc_tmr_regs_t *tmr, uint32_t dutyCount)
{
    tmr->pwm_cap32 = dutyCount;
}

/**
 * @brief      Set a new duty cycle when the timer is used in PWM mode.
 *
 * @param      tmr          Pointer to timer registers for the timer instance to modify.
 * @param      dutyPercent  duty cycle value in percent (0 to 100%)
 */
__STATIC_INLINE void TMR32_SetDutyPer(mxc_tmr_regs_t *tmr, uint32_t dutyPercent)
{
    uint32_t periodCount = tmr->term_cnt32;
    tmr->pwm_cap32 = ((uint64_t)periodCount * dutyPercent) / 100;
}

/**
 * @brief      Set a new period value for PWM timer
 *
 * @param      tmr    Pointer to timer registers for the timer instance to modify.
 * @param      period new period value in timer counts
 */
__STATIC_INLINE void TMR32_SetPeriod(mxc_tmr_regs_t *tmr, uint32_t period)
{
    tmr->term_cnt32 = period;
}

/**
 * @brief Converts frequency and duty cycle % to period and duty ticks
 * @note TMR_Init should be called before this function to set the prescaler
 *
 * @param   tmr         Pointer to timer registers for the timer instance to modify.
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
 * @param  tmr         Pointer to timer registers for the timer instance to modify.
 * @param  time        time value.
 * @param  units       time units.
 * @param  ticks       calculated number of ticks.
 *
 * @returns #E_NO_ERROR if everything is successful, error if unsuccessful.
 */
int TMR32_TimeToTicks(mxc_tmr_regs_t *tmr, uint32_t time, tmr_unit_t units, uint32_t *ticks);

/**
 * @brief Converts a time and units to a number of ticks for the 16-bit timer.
 * @note TMR_Init should be called before this function to set the prescaler
 *
 * @param  tmr         Pointer to timer registers for the timer instance to modify.
 * @param  time        time value.
 * @param  units        time units.
 * @param  ticks       calculated number of ticks.
 *
 * @returns #E_NO_ERROR if everything is successful, error if unsuccessful.
 */
int TMR16_TimeToTicks(mxc_tmr_regs_t *tmr, uint32_t time, tmr_unit_t units, uint16_t *ticks);

/**
 * @brief Converts a number of ticks to a time and units for the timer.
 * @note TMR_Init should be called before this function to set the prescaler
 *
 * @param  tmr         Pointer to timer registers for the timer instance to modify.
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
