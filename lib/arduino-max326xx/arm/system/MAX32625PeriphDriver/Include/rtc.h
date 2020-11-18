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
 * $Date: 2016-03-21 10:05:23 -0500 (Mon, 21 Mar 2016) $
 * $Revision: 22008 $
 *
 ******************************************************************************/

/**
 * @file  rtc.h
 * @addtogroup rtc Real-time Clock
 * @{
 * @brief This is the high level API for the real-time clock module
 */

#ifndef _RTC_H
#define _RTC_H

#include "mxc_config.h"
#include "rtc_regs.h"

#ifdef __cplusplus
extern "C" {
#endif

/// @enum rtc_prescale_t Defines clock divider for 4096Hz input clock. 
typedef enum {
    RTC_PRESCALE_DIV_2_0 = MXC_V_RTC_PRESCALE_DIV_2_0,      ///< (4.096kHz) divide input clock by \f$ 2^{0} = 1 \f$  (0x001)
    RTC_PRESCALE_DIV_2_1 = MXC_V_RTC_PRESCALE_DIV_2_1,      ///< (2.048kHz) divide input clock by \f$ 2^{1} = 2 \f$ (0x002)
    RTC_PRESCALE_DIV_2_2 = MXC_V_RTC_PRESCALE_DIV_2_2,      ///< (1.024kHz) divide input clock by \f$ 2^{2} = 4 \f$
    RTC_PRESCALE_DIV_2_3 = MXC_V_RTC_PRESCALE_DIV_2_3,      ///< (512Hz) divide input clock by \f$ 2^{3} = 8 \f$
    RTC_PRESCALE_DIV_2_4 = MXC_V_RTC_PRESCALE_DIV_2_4,      ///< (256Hz) divide input clock by \f$ 2^{4} = 16 \f$
    RTC_PRESCALE_DIV_2_5 = MXC_V_RTC_PRESCALE_DIV_2_5,      ///< (128Hz) divide input clock by \f$ 2^{5} = 32 \f$
    RTC_PRESCALE_DIV_2_6 = MXC_V_RTC_PRESCALE_DIV_2_6,      ///< (64Hz) divide input clock by \f$ 2^{6} = 64 \f$
    RTC_PRESCALE_DIV_2_7 = MXC_V_RTC_PRESCALE_DIV_2_7,      ///< (32Hz) divide input clock by \f$ 2^{7} = 128 \f$
    RTC_PRESCALE_DIV_2_8 = MXC_V_RTC_PRESCALE_DIV_2_8,      ///< (16Hz) divide input clock by \f$ 2^{8} = 256 \f$
    RTC_PRESCALE_DIV_2_9 = MXC_V_RTC_PRESCALE_DIV_2_9,      ///< (8Hz) divide input clock by \f$ 2^{9} = 512 \f$
    RTC_PRESCALE_DIV_2_10 = MXC_V_RTC_PRESCALE_DIV_2_10,    ///< (4Hz) divide input clock by \f$ 2^{10} = 1024 \f$
    RTC_PRESCALE_DIV_2_11 = MXC_V_RTC_PRESCALE_DIV_2_11,    ///< (2Hz) divide input clock by \f$ 2^{11} = 2048 \f$ (0x07FF)
    RTC_PRESCALE_DIV_2_12 = MXC_V_RTC_PRESCALE_DIV_2_12,    ///< (1Hz) divide input clock by \f$ 2^{12} = 4096 \f$ (0x0FFF)
} rtc_prescale_t;

/// @def RTC_CTRL_ACTIVE_TRANS Active Transaction Flags for the RTC
#define RTC_CTRL_ACTIVE_TRANS           (MXC_F_RTC_CTRL_RTC_ENABLE_ACTIVE 		| \
                                        MXC_F_RTC_CTRL_OSC_GOTO_LOW_ACTIVE		| \
                                        MXC_F_RTC_CTRL_OSC_FRCE_SM_EN_ACTIVE	| \
                                        MXC_F_RTC_CTRL_OSC_FRCE_ST_ACTIVE 		| \
                                        MXC_F_RTC_CTRL_RTC_SET_ACTIVE 			| \
                                        MXC_F_RTC_CTRL_RTC_CLR_ACTIVE 			| \
                                        MXC_F_RTC_CTRL_ROLLOVER_CLR_ACTIVE 		| \
                                        MXC_F_RTC_CTRL_PRESCALE_CMPR0_ACTIVE	| \
                                        MXC_F_RTC_CTRL_PRESCALE_UPDATE_ACTIVE 	| \
                                        MXC_F_RTC_CTRL_CMPR1_CLR_ACTIVE 		| \
                                        MXC_F_RTC_CTRL_CMPR0_CLR_ACTIVE 		| \
                                        MXC_F_RTC_CTRL_TRIM_ENABLE_ACTIVE 		| \
                                        MXC_F_RTC_CTRL_TRIM_SLOWER_ACTIVE 		| \
                                        MXC_F_RTC_CTRL_TRIM_CLR_ACTIVE 			| \
                                        MXC_F_RTC_CTRL_ACTIVE_TRANS_0)

/// @def RTC_FLAGS_CLEAR_ALL Number of RTC compare registers
#define RTC_FLAGS_CLEAR_ALL            (MXC_F_RTC_FLAGS_COMP0  | \
                                        MXC_F_RTC_FLAGS_COMP1| \
                                        MXC_F_RTC_FLAGS_PRESCALE_COMP | \
                                        MXC_F_RTC_FLAGS_OVERFLOW | \
                                        MXC_F_RTC_FLAGS_TRIM)
/// @enum rtc_snooze_t Defines the snooze modes
typedef enum {
    RTC_SNOOZE_DISABLE = MXC_V_RTC_CTRL_SNOOZE_DISABLE,	      ///< Snooze Mode Disabled
    RTC_SNOOZE_MODE_A  = MXC_V_RTC_CTRL_SNOOZE_MODE_A,        ///< COMP1 = COMP1 + RTC_SNZ_VALUE when snooze flag is set
    RTC_SNOOZE_MODE_B  = MXC_V_RTC_CTRL_SNOOZE_MODE_B,        ///< COMP1 = RTC_TIMER + RTC_SNZ_VALUE when snooze flag is set
} rtc_snooze_t; /// Defines the snooze modes

/// @def RTC_NUM_COMPARE Number of RTC compare registers
#define RTC_NUM_COMPARE 2

/// @brief A structure that represents the configuration of the RTC peripheral
typedef struct {
    rtc_prescale_t prescaler; /// prescale value rtc_prescale_t
    rtc_prescale_t prescalerMask; /// Mask value used to compare to the rtc prescale value, when the \f$ \big((Count_{prescaler}\,\&\,Prescale\,Mask) == 0\big) \f$, the prescale compare flag will be set.
    uint32_t compareCount[RTC_NUM_COMPARE]; /// Values used for the RTC alarms. See RTC_SetCompare() and RTC_GetCompare()
    uint32_t snoozeCount; /// The number of RTC ticks to snooze if enabled.
    rtc_snooze_t snoozeMode; /// The desired snooze mode
} rtc_cfg_t;

/**
 * @brief    Initializes the RTC
 * @note     Must setup clocking and power prior to this function.
 *
 * @param    cfg        configuration
 *
 * @retval   E_NO_ERROR if everything is successful
 * @retval   E_NULL_PTR if cfg pointer is NULL
 * @retval   E_INVALID if comparison index, prescaler mask or snooze mask are
 *           out of bounds
 */
int RTC_Init(const rtc_cfg_t *cfg);

/**
 * @brief    Enable and start the real-time clock continuing from its current value
 */
__STATIC_INLINE void RTC_Start(void)
{
    MXC_RTCTMR->ctrl |= MXC_F_RTC_CTRL_ENABLE;

    //wait for pending actions to complete
    while(MXC_RTCTMR->ctrl & MXC_F_RTC_CTRL_PENDING);
}

/**
 * @brief    Disable and stop the real-time clock
 */
__STATIC_INLINE void RTC_Stop(void)
{
    MXC_RTCTMR->ctrl &= ~(MXC_F_RTC_CTRL_ENABLE);

    //wait for pending actions to complete
    while(MXC_RTCTMR->ctrl & MXC_F_RTC_CTRL_PENDING);
}

/**
 * @brief    Determines if the RTC is running or not.
 *
 * @retval  0 if Disabled, Non-zero if Active
 */
__STATIC_INLINE uint32_t RTC_IsActive(void)
{
    return (MXC_RTCTMR->ctrl & MXC_F_RTC_CTRL_ENABLE);
}

/**
 * @brief    Set the current count of the RTC
 *
 * @param    count   count value to set current real-time count.
 */
__STATIC_INLINE void RTC_SetCount(uint32_t count)
{
    MXC_RTCTMR->timer = count;

    //wait for pending actions to complete
    while(MXC_RTCTMR->ctrl & MXC_F_RTC_CTRL_PENDING);
}

/**
 * @brief    Get the current timer value of the RTC.
 *
 * @retval   The value of the RTC counter.
 */
__STATIC_INLINE uint32_t RTC_GetCount(void)
{
    return (MXC_RTCTMR->timer);
}

/**
 * @brief    Set the comparator value
 *
 * @param    compareIndex   Index of comparator to set, see RTC_NUM_COMPARE
 *                          for the total number of compare registers available.
 * @param    counts         Unsigned 32-bit compare value to set.
 * @retval   E_NO_ERROR     Compare count register set successfully for requested
 *                          comparator.
 * @retval   E_INVALID      compareIndex is \>= RTC_NUM_COMPARE.
 */
int RTC_SetCompare(uint8_t compareIndex, uint32_t counts);

/**
 * @brief    Get the comparator value
 *
 * @param    compareIndex   Index of the comparator to get. See RTC_NUM_COMPARE
 *                          for the total number of compare registers available.
 *
 * @retval   uint32_t       The current value of the specified compare register for the RTC
 */
uint32_t RTC_GetCompare(uint8_t compareIndex);

/**
 * @brief   Set the prescale reload value for the real-time clock.
 * @details The prescale reload value determines the number of 4kHz ticks
 *          occur before the timer is incremented. See @ref prescaler_val "Table"
 *          for accepted values and corresponding timer resolution.
 *
 *          <table>
 *          <caption id="prescaler_val">Prescaler Settings and Corresponding RTC Resolutions</caption>
 *          <tr><th>PRESCALE <th>Prescale Reload <th>4kHz ticks in LSB <th>Min Timer Value (sec) <th> Max Timer Value (sec) <th>Max Timer Value (Days) <th> Max Timer Value (Years)
 *          <tr><td>0h <td> RTC_PRESCALE_DIV_2_0 <td> 1 <td> 0.00024 <td> 1048576 <td> 12 <td> 0.0
 *          <tr><td>1h <td> RTC_PRESCALE_DIV_2_1 <td> 2 <td> 0.00049 <td> 2097152 <td> 24 <td> 0.1
 *          <tr><td>2h <td> RTC_PRESCALE_DIV_2_2 <td> 4 <td> 0.00098 <td> 4194304 <td> 49 <td> 0.1
 *          <tr><td>3h <td> RTC_PRESCALE_DIV_2_3 <td> 8 <td> 0.00195 <td> 8388608 <td> 97 <td> 0.3
 *          <tr><td>4h <td> RTC_PRESCALE_DIV_2_4 <td> 16 <td> 0.00391 <td> 16777216 <td> 194 <td> 0.5
 *          <tr><td>5h <td> RTC_PRESCALE_DIV_2_5 <td> 32 <td> 0.00781 <td> 33554432 <td> 388 <td> 1.1
 *          <tr><td>6h <td> RTC_PRESCALE_DIV_2_6 <td> 64 <td> 0.01563 <td> 67108864 <td> 777 <td> 2.2
 *          <tr><td>7h <td> RTC_PRESCALE_DIV_2_7 <td> 128 <td> 0.03125 <td> 134217728 <td> 1553 <td> 4.4
 *          <tr><td>8h <td> RTC_PRESCALE_DIV_2_8 <td> 256 <td> 0.06250 <td> 268435456 <td> 3107 <td> 8.7
 *          <tr><td>9h <td> RTC_PRESCALE_DIV_2_9 <td> 512 <td> 0.12500 <td> 536870912 <td> 6214 <td> 17.5
 *          <tr><td>Ah <td> RTC_PRESCALE_DIV_2_10 <td> 1024 <td> 0.25000 <td> 1073741824 <td> 12428 <td> 34.9
 *          <tr><td>Bh <td> RTC_PRESCALE_DIV_2_11 <td> 2048 <td> 0.50000 <td> 2147483648 <td> 24855 <td> 69.8
 *          <tr><td>Ch <td> RTC_PRESCALE_DIV_2_12 <td> 4096 <td> 1.00000 <td> 4294967296 <td> 49710 <td> 139.6
 *          </table>
 *
 * @param   prescaler   Prescale value to set, see rtc_prescale_t.
 */
__STATIC_INLINE void RTC_SetPrescaler(rtc_prescale_t prescaler)
{
    MXC_RTCTMR->prescale = prescaler;

    //wait for pending actions to complete
    while(MXC_RTCTMR->ctrl & MXC_F_RTC_CTRL_PENDING);
}

/**
 * @brief   Get the current value of the real-time clock prescaler.
 *
 * @retval  rtc_prescale_t  Returns the current RTC prescaler setting,
 *                          See rtc_prescale_t for values of the prescaler.
 */
__STATIC_INLINE rtc_prescale_t RTC_GetPrescaler(void)
{
    return (rtc_prescale_t)(MXC_RTCTMR->prescale);
}

/**
 * @brief           Set the prescaler mask, which is used to set the RTC prescale counter
 *                  compare flag when the prescaler timer matches the bits indicated
 *                  by the mask.
 * @param   mask    A bit mask that is used to set the prescale compare flag if the
 *                  prescale timer has the corresponding bits set. @note This mask must
 *                  be less than or equal to the prescaler reload value.
 *                  See RTC_SetPrescaler()
 * @details         When \f$ \big((Count_{prescaler}\,\&\,Prescale\,Mask) == 0\big) \f$,  the prescale compare flag is set
 * @retval  int     Returns E_NO_ERROR if prescale value is valid and is set.
 * @retval  int     Returns E_INVALID if mask is \> than prescaler value
 */
__STATIC_INLINE int RTC_SetPrescalerMask(rtc_prescale_t mask)
{
    if (mask > ((rtc_prescale_t)(MXC_RTCTMR->prescale)))
    {
        return E_INVALID;
    }
    MXC_RTCTMR->prescale_mask = mask;

    //wait for pending actions to complete
    while(MXC_RTCTMR->ctrl & MXC_F_RTC_CTRL_PENDING);
    return E_NO_ERROR;
}

/**
 * @brief   Set the number of ticks for snooze mode. See RTC_Snooze().
 * @param   count       Sets the count used for snoozing when snooze mode is enabled and
 *                      the snooze flag is set.
 * @retval  E_NO_ERROR  If snooze value is set correctly and value is valid.
 * @retval  E_INVALID   If SnoozeCount exceeds maximum supported, see MXC_F_RTC_SNZ_VAL_VALUE
 *
 */
__STATIC_INLINE int RTC_SetSnoozeCount(uint32_t count)
{
    // Check to make sure max value is not being exceeded
    if (count > MXC_F_RTC_SNZ_VAL_VALUE)
        return E_INVALID;

    MXC_RTCTMR->snz_val = count;

    //wait for pending actions to complete
    while(MXC_RTCTMR->ctrl & MXC_F_RTC_CTRL_PENDING);
    return E_NO_ERROR;
}

/**
 * @brief       Gets the Snooze Count that is currently loaded in the RTC timer
 * @details     Returns the current value for the Snooze. This value is used as
 *              part of the snooze calculation depending on the snooze mode. @see RTC_SetSnoozeMode
 * @retval      uint32_t value of the snooze register
 *
 */
__STATIC_INLINE uint32_t RTC_GetSnoozeCount(void)
{
    return MXC_RTCTMR->snz_val;
}

/**
 * @brief       Set the flags to activate the snooze
 * @details     Begins a snooze of the RTC. When this function is called
 *              the snooze count is determined based on the snooze mode. 
 *              See RTC_GetCount() and RTC_SetSnoozeMode()
 */
__STATIC_INLINE void RTC_Snooze(void)
{
    MXC_RTCTMR->flags = MXC_F_RTC_FLAGS_SNOOZE_A | MXC_F_RTC_FLAGS_SNOOZE_B;

    //wait for pending actions to complete
    while(MXC_RTCTMR->ctrl & MXC_F_RTC_CTRL_PENDING);
}

/**
 * @brief   Sets the Snooze Mode.
 * @details <table>
 *          <caption id="snoozeModesTable">Snooze Modes</caption>
 *          <tr><th>Mode<th>Snooze Time Calculation
 *          <tr><td>RTC_SNOOZE_DISABLE<td>Snooze Disabled
 *          <tr><td>RTC_SNOOZE_MODE_A<td>\f$ compare1 = compare1 + snoozeCount \f$
 *          <tr><td>RTC_SNOOZE_MODE_B<td>\f$ compare1 = count + snoozeCount \f$
 *          </table>
 *          \a count is the value of the RTC counter when RTC_Snooze() is called to begin snooze
 */
__STATIC_INLINE void RTC_SetSnoozeMode(rtc_snooze_t mode)
{
    uint32_t ctrl;
    // Get the control register and mask off the non-snooze bits
    ctrl = (MXC_RTCTMR->ctrl & ~(MXC_F_RTC_CTRL_SNOOZE_ENABLE));
    // set the requested snooze mode bits and save the settings
    MXC_RTCTMR->ctrl = (ctrl | (mode << MXC_F_RTC_CTRL_SNOOZE_ENABLE_POS));

    //wait for pending actions to complete
    while(MXC_RTCTMR->ctrl & MXC_F_RTC_CTRL_PENDING);
}

/**
 * @brief   Enables the interrupts defined by the mask for the RTC.
 * @details <table>
 *          <caption id="RTC_interrupts">RTC Interrupts</caption>
 *          <tr><th>Interrupt<th>Mask
 *          <tr><td>Compare 0<td>MXC_F_RTC_INTEN_COMP0
 *          <tr><td>Compare 1 \\ Snooze<td>MXC_F_RTC_INTEN_COMP1
 *          <tr><td>Prescale Comp<td>MXC_F_RTC_FLAGS_INTEN_COMP
 *          <tr><td>RTC Count Overflow<td>MXC_F_RTC_INTEN_OVERFLOW
 *          <tr><td>Trim<td>MXC_F_RTC_INTEN_TRIM
 *          </table>
 * @param    mask        set the bits of the interrupts to enable
 */
__STATIC_INLINE void RTC_EnableINT(uint32_t mask)
{
    MXC_RTCTMR->inten |= mask;
}

/**
 * @brief    Disable RTC interrupts based on the mask, See @ref RTC_interrupts
 *
 * @param    mask   set the bits of the interrupts to disable
 */
__STATIC_INLINE void RTC_DisableINT(uint32_t mask)
{
    MXC_RTCTMR->inten &= ~mask;
}

/**
 * @brief    Gets the RTC's interrupt flags
 *
 * @retval   uint32_t    interrupt flags
 */
__STATIC_INLINE uint32_t RTC_GetFlags(void)
{
    return (MXC_RTCTMR->flags);
}

/**
 * @brief    Clears the RTC's interrupt flags based on the mask provided
 *
 * @param    mask    masked used to clear individual interrupt flags
 */
__STATIC_INLINE void RTC_ClearFlags(uint32_t mask)
{
    MXC_RTCTMR->flags = mask;

    //wait for pending actions to complete
    while(MXC_RTCTMR->ctrl & MXC_F_RTC_CTRL_PENDING);
}

/**
 * @brief    Gets the active transaction flags
 *
 * @retval  0 = no active transactions , nonzero = active transactions bits
 */
__STATIC_INLINE uint32_t RTC_GetActiveTrans(void)
{
    return (MXC_RTCTMR->ctrl & RTC_CTRL_ACTIVE_TRANS);
}

/**
 * @brief    Sets the trim value and trim slow/fast option
 * @note     Ensure RTC is disabled prior to calling this function
 *
 * @param    trim        trim value - maximum trim value setting of 0x03FFFF
 * @param    trimSlow    1 = trim slow, 0 = trim fast
 *
 * @retval   E_NO_ERROR  Trim value is valid and set.
 * @retval   E_INVALID   Trim value exceeds max trim.
 * @retval   E_BAD_STATE RTC is not disabled.
 *
 */
int RTC_SetTrim(uint32_t trim, uint8_t trimSlow);

/**
 * @brief    Gets the trim value currently set
 * @note     Ensure RTC is disabled prior to calling this function
 *
 * @retval   uint32_t    Current trim value of RTC.
 */
uint32_t RTC_GetTrim(void);

/**
 * @brief    Enabled the trim.
 * @note     Ensure RTC is disabled prior to calling this function
 * @retval   E_NO_ERROR Trim enabled
 * @retval   E_INVALID
 */
int RTC_TrimEnable(void);

/**
 * @brief    Disable the trim.
 */
void RTC_TrimDisable(void);

/**
 * @}
 */

#ifdef __cplusplus
}
#endif

#endif /* _RTC_H */
