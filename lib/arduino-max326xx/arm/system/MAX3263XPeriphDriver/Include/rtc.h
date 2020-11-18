/**
 * @file  
 * @brief Real-Time Clock data types, definitions and function prototypes.
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
 * $Date: 2017-02-16 12:04:19 -0600 (Thu, 16 Feb 2017) $
 * $Revision: 26465 $
 *
 **************************************************************************** */

/* Define to prevent redundant inclusion */
#ifndef _RTC_H_
#define _RTC_H_

/* **** Includes **** */
#include "mxc_config.h"
#include "rtc_regs.h"

#ifdef __cplusplus
extern "C" {
#endif

/**
 * @ingroup periphlibs
 * @defgroup rtc Real-Time Clock (RTC)
 * @brief Functions, types, and registers for the Real-Time Clock Peripheral.
 * @{
 */

/* **** Definitions **** */
/**
 * Enumeration type for scaling down the 4096Hz input clock to the RTC. 
 */
typedef enum {
    RTC_PRESCALE_DIV_2_0 = MXC_V_RTC_PRESCALE_DIV_2_0,      /**< \f$ f_{RTC} = \frac {4096} {2^{0}} = 4096Hz \f$ */
    RTC_PRESCALE_DIV_2_1 = MXC_V_RTC_PRESCALE_DIV_2_1,      /**< \f$ f_{RTC} = \frac {4096} {2^{1}} = 2048Hz \f$ */
    RTC_PRESCALE_DIV_2_2 = MXC_V_RTC_PRESCALE_DIV_2_2,      /**< \f$ f_{RTC} = \frac {4096} {2^{2}} = 1024Hz \f$ */
    RTC_PRESCALE_DIV_2_3 = MXC_V_RTC_PRESCALE_DIV_2_3,      /**< \f$ f_{RTC} = \frac {4096} {2^{3}} = 512Hz \f$ */
    RTC_PRESCALE_DIV_2_4 = MXC_V_RTC_PRESCALE_DIV_2_4,      /**< \f$ f_{RTC} = \frac {4096} {2^{4}} = 256Hz \f$ */
    RTC_PRESCALE_DIV_2_5 = MXC_V_RTC_PRESCALE_DIV_2_5,      /**< \f$ f_{RTC} = \frac {4096} {2^{5}} = 128Hz \f$ */
    RTC_PRESCALE_DIV_2_6 = MXC_V_RTC_PRESCALE_DIV_2_6,      /**< \f$ f_{RTC} = \frac {4096} {2^{6}} = 64Hz \f$ */
    RTC_PRESCALE_DIV_2_7 = MXC_V_RTC_PRESCALE_DIV_2_7,      /**< \f$ f_{RTC} = \frac {4096} {2^{7}} = 32Hz \f$ */
    RTC_PRESCALE_DIV_2_8 = MXC_V_RTC_PRESCALE_DIV_2_8,      /**< \f$ f_{RTC} = \frac {4096} {2^{8}} = 16Hz \f$ */
    RTC_PRESCALE_DIV_2_9 = MXC_V_RTC_PRESCALE_DIV_2_9,      /**< \f$ f_{RTC} = \frac {4096} {2^{9}} = 8Hz \f$ */
    RTC_PRESCALE_DIV_2_10 = MXC_V_RTC_PRESCALE_DIV_2_10,    /**< \f$ f_{RTC} = \frac {4096} {2^{10}} = 4Hz \f$ */
    RTC_PRESCALE_DIV_2_11 = MXC_V_RTC_PRESCALE_DIV_2_11,    /**< \f$ f_{RTC} = \frac {4096} {2^{11}} = 2Hz \f$ */
    RTC_PRESCALE_DIV_2_12 = MXC_V_RTC_PRESCALE_DIV_2_12,    /**< \f$ f_{RTC} = \frac {4096} {2^{12}} = 1Hz \f$ */
} rtc_prescale_t;

/**
 * Mask of the RTC Flags for the Active Transaction.
 */
#define RTC_CTRL_ACTIVE_TRANS           (MXC_F_RTC_CTRL_RTC_ENABLE_ACTIVE       | \
                                        MXC_F_RTC_CTRL_OSC_GOTO_LOW_ACTIVE      | \
                                        MXC_F_RTC_CTRL_OSC_FRCE_SM_EN_ACTIVE    | \
                                        MXC_F_RTC_CTRL_OSC_FRCE_ST_ACTIVE       | \
                                        MXC_F_RTC_CTRL_RTC_SET_ACTIVE           | \
                                        MXC_F_RTC_CTRL_RTC_CLR_ACTIVE           | \
                                        MXC_F_RTC_CTRL_ROLLOVER_CLR_ACTIVE      | \
                                        MXC_F_RTC_CTRL_PRESCALE_CMPR0_ACTIVE    | \
                                        MXC_F_RTC_CTRL_PRESCALE_UPDATE_ACTIVE   | \
                                        MXC_F_RTC_CTRL_CMPR1_CLR_ACTIVE         | \
                                        MXC_F_RTC_CTRL_CMPR0_CLR_ACTIVE         | \
                                        MXC_F_RTC_CTRL_TRIM_ENABLE_ACTIVE       | \
                                        MXC_F_RTC_CTRL_TRIM_SLOWER_ACTIVE       | \
                                        MXC_F_RTC_CTRL_TRIM_CLR_ACTIVE          | \
                                        MXC_F_RTC_CTRL_ACTIVE_TRANS_0)

/**
 * Mask used to clear all RTC interrupt flags, see \ref RTC_FLAGS_Register Register. 
 */
#define RTC_FLAGS_CLEAR_ALL            (MXC_F_RTC_FLAGS_COMP0  | \
                                        MXC_F_RTC_FLAGS_COMP1| \
                                        MXC_F_RTC_FLAGS_PRESCALE_COMP | \
                                        MXC_F_RTC_FLAGS_OVERFLOW | \
                                        MXC_F_RTC_FLAGS_TRIM)
/**
 * Enumeration type to select the type of RTC Snooze Mode for an alarm condition. 
 */
typedef enum {
    RTC_SNOOZE_DISABLE = MXC_V_RTC_CTRL_SNOOZE_DISABLE,       /**< Snooze Mode Disabled */
    RTC_SNOOZE_MODE_A  = MXC_V_RTC_CTRL_SNOOZE_MODE_A,        /**< \f$ COMP1 = COMP1 + RTC\_SNZ\_VALUE \f$ when snooze flag is set */
    RTC_SNOOZE_MODE_B  = MXC_V_RTC_CTRL_SNOOZE_MODE_B,        /**< \f$ COMP1 = RTC\_TIMER + RTC\_SNZ\_VALUE \f$ when snooze flag is set */
} rtc_snooze_t; 

/**
 * Number of RTC Compare registers for this peripheral instance. 
 */
#define RTC_NUM_COMPARE 2

/**
 * Structure type that represents the current configuration of the RTC.
 */

typedef struct {
    rtc_prescale_t prescaler;               /**< prescale value for the input 4096Hz clock. */
    rtc_prescale_t prescalerMask;           /**< Mask value used to compare to the rtc prescale value, when the \f$ (Count_{prescaler}\,\&\,Prescale\,Mask) == 0 \f$, the prescale compare flag will be set. */
    uint32_t compareCount[RTC_NUM_COMPARE]; /**< Values used for setting the RTC alarms. See RTC_SetCompare(uint8_t compareIndex, uint32_t counts) and RTC_GetCompare(). */
    uint32_t snoozeCount;                   /**< The number of RTC ticks to snooze if enabled. */
    rtc_snooze_t snoozeMode;                /**< The desired snooze mode, see #rtc_snooze_t. */
} rtc_cfg_t;

/**
 * @brief      Initializes the RTC
 * @note       Must setup clocking and power prior to this function.
 *
 * @param      cfg          RTC configuration object.
 *
 * @retval     #E_NO_ERROR  RTC initialized successfully.
 * @retval     #E_NULL_PTR  \p cfg pointer is NULL.
 * @retval     #E_INVALID   if comparison index, prescaler mask or snooze mask
 *                          are out of bounds, see #rtc_cfg_t.
 */
int RTC_Init(const rtc_cfg_t *cfg);

/**
 * @brief    Enable and start the real-time clock continuing from its current value.
 */
__STATIC_INLINE void RTC_Start(void)
{
    MXC_RTCTMR->ctrl |= MXC_F_RTC_CTRL_ENABLE;

    //wait for pending actions to complete
    while(MXC_RTCTMR->ctrl & MXC_F_RTC_CTRL_PENDING);
}

/**
 * @brief    Disable and stop the real-time clock counting. 
 */
__STATIC_INLINE void RTC_Stop(void)
{
    MXC_RTCTMR->ctrl &= ~(MXC_F_RTC_CTRL_ENABLE);

    //wait for pending actions to complete
    while(MXC_RTCTMR->ctrl & MXC_F_RTC_CTRL_PENDING);
}

/**
 * @brief      Returns the state (running or disabled) for the RTC.
 *
 * @retval     0         Disabled.
 * @retval     Non-zero  Active.
 */
__STATIC_INLINE uint32_t RTC_IsActive(void)
{
    return (MXC_RTCTMR->ctrl & MXC_F_RTC_CTRL_ENABLE);
}

/**
 * @brief    Set the current count of the RTC
 *
 * @param    count   The desired count value to set for the RTC count.
 */
__STATIC_INLINE void RTC_SetCount(uint32_t count)
{
    MXC_RTCTMR->timer = count;

    //wait for pending actions to complete
    while(MXC_RTCTMR->ctrl & MXC_F_RTC_CTRL_PENDING);
}

/**
 * @brief    Get the current count value of the RTC.
 *
 * @retval   The value of the RTC counter.
 */
__STATIC_INLINE uint32_t RTC_GetCount(void)
{
    return (MXC_RTCTMR->timer);
}

/**
 * @brief      Sets the compare value for the RTC.
 *
 * @param      compareIndex  Index of comparator to set, see #RTC_NUM_COMPARE
 *                           for the total number of compare registers
 *                           available.
 * @param      counts        The value to set for the compare.
 * @retval     #E_NO_ERROR   Compare count register set successfully for
 *                           requested comparator.
 * @retval     #E_INVALID    compareIndex is @>= RTC_NUM_COMPARE.
 */
int RTC_SetCompare(uint8_t compareIndex, uint32_t counts);

/**
 * @brief    Gets the compare value for the RTC.
 *
 * @param    compareIndex   Index of the compare value to return. See #RTC_NUM_COMPARE
 *                          for the total number of compare registers available.
 *
 * @returns  The current value of the specified compare register for the RTC.
 */
uint32_t RTC_GetCompare(uint8_t compareIndex);

/**
 * @brief   Set the prescale reload value for the real-time clock.
 * @details The prescale reload value determines the number of 4kHz ticks
 *          that will occur before the timer is incremented. 
 *          <table>
 *          <caption id="prescaler_val">Prescaler Settings and Corresponding RTC Resolutions</caption>
 *          <tr><th>PRESCALE <th>Prescale Reload <th>4kHz ticks in LSB <th>Min Timer Value (sec) <th> Max Timer Value (sec) <th>Max Timer Value (Days) <th> Max Timer Value (Years)
 *          <tr><td align="right">0h<td align="center">RTC_PRESCALE_DIV_2_0<td  align="right">1<td align="right">0.00024<td align="right">1048576<td align="right">12<td align="right">0.0
 *          <tr><td align="right">1h<td align="center">RTC_PRESCALE_DIV_2_1<td align="right">2<td align="right">0.00049<td align="right">2097152<td align="right">24<td align="right">0.1
 *          <tr><td align="right">2h<td align="center">RTC_PRESCALE_DIV_2_2<td align="right">4<td align="right">0.00098<td align="right">4194304<td align="right">49<td align="right">0.1
 *          <tr><td align="right">3h<td align="center">RTC_PRESCALE_DIV_2_3<td align="right">8<td align="right">0.00195<td align="right">8388608<td align="right">97<td align="right">0.3
 *          <tr><td align="right">4h<td align="center">RTC_PRESCALE_DIV_2_4<td align="right">16<td align="right">0.00391<td align="right">16777216<td align="right">194 <td align="right">0.5
 *          <tr><td align="right">5h<td align="center">RTC_PRESCALE_DIV_2_5<td align="right">32<td align="right">0.00781<td align="right">33554432<td align="right">388 <td align="right">1.1
 *          <tr><td align="right">6h<td align="center">RTC_PRESCALE_DIV_2_6<td align="right">64<td align="right">0.01563<td align="right">67108864<td align="right">777 <td align="right">2.2
 *          <tr><td align="right">7h<td align="center">RTC_PRESCALE_DIV_2_7<td align="right">128<td align="right">0.03125<td align="right">134217728<td align="right">1553 <td align="right">4.4
 *          <tr><td align="right">8h<td align="center">RTC_PRESCALE_DIV_2_8<td align="right">256<td align="right">0.06250<td align="right">268435456<td align="right">3107 <td align="right">8.7
 *          <tr><td align="right">9h<td align="center">RTC_PRESCALE_DIV_2_9<td align="right">512<td align="right">0.12500<td align="right">536870912<td align="right">6214 <td align="right">17.5
 *          <tr><td align="right">Ah<td align="center">RTC_PRESCALE_DIV_2_10<td align="right">1024 <td align="right">0.25000<td align="right">1073741824<td align="right">12428<td align="right">34.9
 *          <tr><td align="right">Bh<td align="center">RTC_PRESCALE_DIV_2_11<td align="right">2048 <td align="right">0.50000<td align="right">2147483648<td align="right">24855<td align="right">69.8
 *          <tr><td align="right">Ch<td align="center">RTC_PRESCALE_DIV_2_12<td align="right">4096 <td align="right">1.00000<td align="right">4294967296<td align="right">49710<td align="right">139.6
 *          </table>
 *
 * @param   prescaler   Prescale value to set, see #rtc_prescale_t.
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
 *                          See #rtc_prescale_t for values of the prescaler.
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
 * @details         When \f$ Count_{prescaler}\,\&\,Prescale\,Mask = 0 \f$,  the prescale compare flag is set
 * @retval  int     Returns #E_NO_ERROR if prescale value is valid and is set.
 * @retval  int     Returns #E_INVALID if mask is \> than prescaler value
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
 * @brief      Set the count for snooze mode. See RTC_Snooze().
 * @details    This value is used to set the snooze count. The meaning of this value is dependant on the snooze mode. p
 * See RTC_SetSnoozeMode() for details of calculating the snooze time period based on the mode and snooze count.
 * @param      count        Sets the count used for snooze when snooze mode is
 *                          enabled and the snooze flag is set.
 * @retval     #E_NO_ERROR  Snooze value is set correctly and value is valid.
 * @retval     #E_INVALID   SnoozeCount exceeds maximum supported, see
 *                          MXC_F_RTC_SNZ_VAL_VALUE
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
 * @brief      Gets the current Snooze Count value.
 * @details    Returns the current value for the Snooze Count. This value is
 *             used as part of the snooze calculation depending on the snooze
 *             mode. See RTC_SetSnoozeMode() for details of calculating the
 *             snooze time period based on the mode and count.
 * @return     Value of the snooze register.
 */
__STATIC_INLINE uint32_t RTC_GetSnoozeCount(void)
{
    return MXC_RTCTMR->snz_val;
}

/**
 * @brief       Activates snooze mode.
 * @details     Begins a snooze of the RTC. When this function is called
 *              the snooze time period is determined based on the snooze mode and the count. 
 *              See RTC_GetCount() and RTC_SetSnoozeMode()
 */
__STATIC_INLINE void RTC_Snooze(void)
{
    MXC_RTCTMR->flags = MXC_F_RTC_FLAGS_SNOOZE_A | MXC_F_RTC_FLAGS_SNOOZE_B;

    //wait for pending actions to complete
    while(MXC_RTCTMR->ctrl & MXC_F_RTC_CTRL_PENDING);
}

/**
 * @brief      Sets the Snooze Mode.
 * @details    <table> <caption id="snoozeModesTable">Snooze Modes</caption>
 *             <tr><th>Mode<th>Snooze Time Calculation
 *             <tr><td>RTC_SNOOZE_DISABLE<td>Snooze Disabled
 *             <tr><td>RTC_SNOOZE_MODE_A<td>\f$ compare1 = compare1 + snoozeCount \f$ 
 *             <tr><td>RTC_SNOOZE_MODE_B<td>\f$ compare1 = count + snoozeCount \f$
 *             </table> 
 * @note       @a count is the value of the RTC counter when RTC_Snooze() 
 *             is called to start a snooze cycle and @a snoozeCount is the value set by the RTC_SetSnoozeCount(uint32_t count) function.
 *             
 * @param      mode Specifies the desired snooze mode, see #rtc_snooze_t.
 * 
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
 * @brief   Enables the interrupts specified for the RTC.
 * @details <table>
 *          <caption id="RTC_interrupts">RTC Interrupts</caption>
 *          <tr><th>Interrupt<th>Mask
 *          <tr><td>Compare 0<td>#MXC_F_RTC_INTEN_COMP0
 *          <tr><td>Compare 1 and Snooze<td>#MXC_F_RTC_INTEN_COMP1
 *          <tr><td>Prescale Comp<td>#MXC_F_RTC_INTEN_PRESCALE_COMP
 *          <tr><td>RTC Count Overflow<td>#MXC_F_RTC_INTEN_OVERFLOW
 *          <tr><td>Trim<td>#MXC_F_RTC_INTEN_TRIM
 *          </table>
 * @param    mask  A mask of the RTC interrupts to enable, 1 = Enable. See 
 *                 @ref RTC_FLAGS_Register Register for the RTC interrupt enable bit masks and positions.
 */
__STATIC_INLINE void RTC_EnableINT(uint32_t mask)
{
    MXC_RTCTMR->inten |= mask;
}

/**
 * @brief      Disable the interrupts specified for the RTC. See the 
 *             @ref RTC_INTEN_Register Register for the RTC interrupt enable bit masks and positions.
 *
 * @param      mask  A mask of the RTC interrupts to disable, 1 = Disable. 
 */
__STATIC_INLINE void RTC_DisableINT(uint32_t mask)
{
    MXC_RTCTMR->inten &= ~mask;
}

/**
 * @brief      Returns the current interrupt flags that are set.
 *
 * @return     A mask of the current interrupt flags, see the
 *             @ref RTC_FLAGS_Register Register for the details of the RTC interrupt
 *             flags.
 */
__STATIC_INLINE uint32_t RTC_GetFlags(void)
{
    return (MXC_RTCTMR->flags);
}

/**
 * @brief      Clears the interrupt flags specified.
 *
 * @param      mask  A mask of interrupts to clear, see the @ref RTC_FLAGS_Register
 *                   Register for the interrupt flag bit masks and positions.
 */
__STATIC_INLINE void RTC_ClearFlags(uint32_t mask)
{
    MXC_RTCTMR->flags = mask;

    //wait for pending actions to complete
    while(MXC_RTCTMR->ctrl & MXC_F_RTC_CTRL_PENDING);
}

/**
 * @brief    Gets the active transaction flags, see @ref RTC_CTRL_Register Register for the list of ACTIVE flags. 
 *
 * @retval   0      No active transactions.
 * @retval  nonzero A mask of active transaction bits.
 */
__STATIC_INLINE uint32_t RTC_GetActiveTrans(void)
{
    return (MXC_RTCTMR->ctrl & RTC_CTRL_ACTIVE_TRANS);
}

/**
 * @brief      Sets the trim value and trim slow/fast option.
 * @warning    The RTC must be disabled prior to calling this function, see RTC_Stop(void) to disable the RTC.
 *
 * @param      trim      The desired trim value. @note The maximum trim value setting is 0x03FFFF.
 * @param      trimSlow  1 = trim slow, 0 = trim fast
 *
 * @return     #E_NO_ERROR  Trim value is valid and set.
 * @return     #E_INVALID   Trim value exceeds max trim.
 * @return     #E_BAD_STATE RTC is active.
 */
int RTC_SetTrim(uint32_t trim, uint8_t trimSlow);

/**
 * @brief    Gets the current trim value.
 * @note     Ensure RTC is disabled prior to calling this function, see RTC_Stop(void).
 *
 * @retval   uint32_t    Current trim value of RTC.
 */
uint32_t RTC_GetTrim(void);

/**
 * @brief    Enable the trim.
 * @warning    The RTC must be disabled prior to calling this function, see RTC_Stop(void) to disable the RTC.
 * @retval   #E_NO_ERROR Trim is enabled.
 * @retval   #E_INVALID RTC is active, see RTC_Stop(void).
 */
int RTC_TrimEnable(void);

/**
 * @brief    Disable the trim.
 */
void RTC_TrimDisable(void);

/**@}*/

#ifdef __cplusplus
}
#endif

#endif /* _RTC_H */
