/**
 * @file
 * @brief WDT2 peripheral module API.
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
 * $Date: 2017-02-16 15:06:56 -0600 (Thu, 16 Feb 2017) $
 * $Revision: 26491 $
 *
 **************************************************************************** */

/* Define to prevent redundant inclusion */
#ifndef _WDT2_H
#define _WDT2_H

/* **** Includes **** */
#include "mxc_config.h"
#include "wdt2_regs.h"
#include "mxc_assert.h"

#ifdef __cplusplus
extern "C" {
#endif

/**
 * @ingroup wdttimers
 * @defgroup wdt2 Watch Dog Timer 2
 * @brief WDT2 configuration and control API.
 * @{
 */

/** 
 * Definition to clear all WDT2 flags 
 */
#define WDT2_FLAGS_CLEAR_ALL    (MXC_F_WDT2_FLAGS_TIMEOUT | MXC_F_WDT2_FLAGS_RESET_OUT)
/**
 * Enumeration type to select the Watchdog Timer's Period 
 */
typedef enum {
    WDT2_PERIOD_2_25_CLKS  = MXC_V_WDT2_CTRL_INT_PERIOD_2_25_NANO_CLKS, /**< \f$ 2^{25}\f$ Nano Ring clocks. */
    WDT2_PERIOD_2_24_CLKS  = MXC_V_WDT2_CTRL_INT_PERIOD_2_24_NANO_CLKS, /**< \f$ 2^{24}\f$ Nano Ring clocks. */
    WDT2_PERIOD_2_23_CLKS  = MXC_V_WDT2_CTRL_INT_PERIOD_2_23_NANO_CLKS, /**< \f$ 2^{23}\f$ Nano Ring clocks. */
    WDT2_PERIOD_2_22_CLKS  = MXC_V_WDT2_CTRL_INT_PERIOD_2_22_NANO_CLKS, /**< \f$ 2^{22}\f$ Nano Ring clocks. */
    WDT2_PERIOD_2_21_CLKS  = MXC_V_WDT2_CTRL_INT_PERIOD_2_21_NANO_CLKS, /**< \f$ 2^{21}\f$ Nano Ring clocks. */
    WDT2_PERIOD_2_20_CLKS  = MXC_V_WDT2_CTRL_INT_PERIOD_2_20_NANO_CLKS, /**< \f$ 2^{20}\f$ Nano Ring clocks. */
    WDT2_PERIOD_2_19_CLKS  = MXC_V_WDT2_CTRL_INT_PERIOD_2_19_NANO_CLKS, /**< \f$ 2^{19}\f$ Nano Ring clocks. */
    WDT2_PERIOD_2_18_CLKS  = MXC_V_WDT2_CTRL_INT_PERIOD_2_18_NANO_CLKS, /**< \f$ 2^{18}\f$ Nano Ring clocks. */
    WDT2_PERIOD_2_17_CLKS  = MXC_V_WDT2_CTRL_INT_PERIOD_2_17_NANO_CLKS, /**< \f$ 2^{17}\f$ Nano Ring clocks. */
    WDT2_PERIOD_2_16_CLKS  = MXC_V_WDT2_CTRL_INT_PERIOD_2_16_NANO_CLKS, /**< \f$ 2^{16}\f$ Nano Ring clocks. */
    WDT2_PERIOD_2_15_CLKS  = MXC_V_WDT2_CTRL_INT_PERIOD_2_15_NANO_CLKS, /**< \f$ 2^{15}\f$ Nano Ring clocks. */
    WDT2_PERIOD_2_14_CLKS  = MXC_V_WDT2_CTRL_INT_PERIOD_2_14_NANO_CLKS, /**< \f$ 2^{14}\f$ Nano Ring clocks. */
    WDT2_PERIOD_2_13_CLKS  = MXC_V_WDT2_CTRL_INT_PERIOD_2_13_NANO_CLKS, /**< \f$ 2^{13}\f$ Nano Ring clocks. */
    WDT2_PERIOD_2_12_CLKS  = MXC_V_WDT2_CTRL_INT_PERIOD_2_12_NANO_CLKS, /**< \f$ 2^{12}\f$ Nano Ring clocks. */
    WDT2_PERIOD_2_11_CLKS  = MXC_V_WDT2_CTRL_INT_PERIOD_2_11_NANO_CLKS, /**< \f$ 2^{11}\f$ Nano Ring clocks. */
    WDT2_PERIOD_2_10_CLKS  = MXC_V_WDT2_CTRL_INT_PERIOD_2_10_NANO_CLKS, /**< \f$ 2^{10}\f$ Nano Ring clocks. */
    WDT2_PERIOD_MAX                                                     /**< Maximum Period is Max - 1 */
} wdt2_period_t;


/**
 * @brief      Initializes the NanoRing for the watchdog clock and sets watchdog
 *             in a known disabled state
 * @param      runInSleep    If non-zero, the WDT2 operates in Sleep Modes for
 *                           the device, 0 disables the WDT2 during Sleep Modes.
 * @param      unlock_key    The WDT2 unlock key value, use
 *                           #MXC_V_WDT2_UNLOCK_KEY
 *
 * @retval     #E_NO_ERROR   Watchdog Timer initialized as requested
 * @retval     #E_BAD_STATE  Invalid unlock_key, WDT2 failed to unlock.
 */
int WDT2_Init(uint8_t runInSleep, uint8_t unlock_key);

/**
 * @brief      Configures and enables the wake-up timeout for the watchdog
 *             specified.
 *
 * @param      int_period    Interrupt period.
 * @param      unlock_key    Key to unlock watchdog.
 *
 * @retval     #E_NO_ERROR   WDT2 Interrupt period enabled with the int_period
 *                           time.
 * @retval     #E_INVALID    Requested Period is greater than the maximum
 *                           supported
 * @retval     #E_BAD_STATE  Invalid unlock_key, WDT2 failed to unlock.
 */
int WDT2_EnableWakeUp(wdt2_period_t int_period, uint8_t unlock_key);

/**
 * @brief      Disables the interrupt timeout for the watchdog specified.
 *
 * @param      unlock_key    Key to unlock watchdog.
 *
 * @retval     #E_NO_ERROR   Wakeup disabled.
 * @retval     #E_BAD_STATE  Invalid unlock_key, WDT2 failed to unlock.
 */
int WDT2_DisableWakeUp(uint8_t unlock_key);

/**
 * @brief      Configures and enables the reset timeout for the watchdog
 *             specified.
 *
 * @param      rst_period    Reset period.
 * @param      unlock_key    Key to unlock watchdog.
 *
 * @retval     #E_NO_ERROR   Reset timeout enabled with the rst_period time.
 * @retval     #E_INVALID    Requested Period is greater than the maximum
 *                           supported
 * @retval     #E_BAD_STATE  Invalid unlock_key, WDT2 failed to unlock.
 */
int WDT2_EnableReset(wdt2_period_t rst_period, uint8_t unlock_key);

/**
 * @brief      Disables the reset timeout for the watchdog specified.
 *
 * @param      unlock_key    Key to unlock watchdog.
 *
 * @retval     #E_NO_ERROR   Reset disabled.
 * @retval     #E_BAD_STATE  Invalid unlock_key, WDT2 failed to unlock.
 */
int WDT2_DisableReset(uint8_t unlock_key);

/**
 * @brief      Gets the watchdog flags
 *
 * @retval     0         No flags set.
 * @retval     non-zero  The WDT2 interrupt flags that are
 *                       set, see @ref WDT2_FLAGS_Register "WDT2_FLAGS
 *                       register".
 */
__STATIC_INLINE uint32_t WDT2_GetFlags(void)
{
    return (MXC_WDT2->flags);
}

/**
 * @brief      Clears the watchdog flags based on the @p mask.
 *
 * @param      mask  bits to clear
 */
__STATIC_INLINE void WDT2_ClearFlags(uint32_t mask)
{
    MXC_WDT2->flags = mask;
}

/**
 * @brief      Starts the watchdog specified.
 *
 * @param      unlock_key    Key to unlock watchdog.
 *
 * @retval     #E_NO_ERROR   WDT2 started.
 * @retval     #E_BAD_STATE  Invalid unlock_key, WDT2 failed to unlock.
 */
int WDT2_Start(uint8_t unlock_key);

/**
 * @brief      Feeds the watchdog specified.
 *
 * @retval     #E_NO_ERROR  WDT2 reset successfully.
 */
void WDT2_Reset(void);

/**
 * @brief      Stops the WatchDog Timer 2.
 *
 * @param      unlock_key    Key to unlock watchdog.
 *
 * @retval     #E_NO_ERROR   WDT2 stopped.
 * @retval     #E_BAD_STATE  Invalid unlock_key, WDT2 failed to unlock.
 */
int WDT2_Stop(uint8_t unlock_key);

/**
 * @brief      Determines if the watchdog is running
 *
 * @retval     0         Inactive
 * @retval     non-zero  Active
 */
__STATIC_INLINE int WDT2_IsActive(void)
{
    return (!!(MXC_WDT2->ctrl & MXC_F_WDT2_CTRL_EN_TIMER));
}

/**@} end of group wdt2*/

#ifdef __cplusplus
}
#endif

#endif /* _WDT_H */
