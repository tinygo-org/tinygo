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
 * $Date: 2016-03-21 15:44:11 -0500 (Mon, 21 Mar 2016) $
 * $Revision: 22024 $
 *
 ******************************************************************************/

/**
 * @file wdt.h
 * @addtogroup wdt WDT
 * @{
 * @brief This is the high level API for the watchdog timer interface module
 *
 */

#ifndef _WDT_H
#define _WDT_H

#include "mxc_config.h"
#include "wdt_regs.h"
#include "mxc_assert.h"
#include "mxc_sys.h"

#ifdef __cplusplus
extern "C" {
#endif

#define WDT_FLAGS_CLEAR_ALL                 (MXC_F_WDT_FLAGS_TIMEOUT| \
                                            MXC_F_WDT_FLAGS_PRE_WIN | \
                                            MXC_F_WDT_FLAGS_RESET_OUT)
/** @enum wdt_period_t */
/** @brief Enumeration for the Watchdog Timer Period */
typedef enum {
    WDT_PERIOD_2_31_CLKS  = MXC_V_WDT_CTRL_INT_PERIOD_2_31_CLKS, /**< 2^31 WDT clocks. */
    WDT_PERIOD_2_30_CLKS  = MXC_V_WDT_CTRL_INT_PERIOD_2_30_CLKS, /**< 2^30 WDT clocks. */
    WDT_PERIOD_2_29_CLKS  = MXC_V_WDT_CTRL_INT_PERIOD_2_29_CLKS, /**< 2^29 WDT clocks. */
    WDT_PERIOD_2_28_CLKS  = MXC_V_WDT_CTRL_INT_PERIOD_2_28_CLKS, /**< 2^28 WDT clocks. */
    WDT_PERIOD_2_27_CLKS  = MXC_V_WDT_CTRL_INT_PERIOD_2_27_CLKS, /**< 2^27 WDT clocks. */
    WDT_PERIOD_2_26_CLKS  = MXC_V_WDT_CTRL_INT_PERIOD_2_26_CLKS, /**< 2^26 WDT clocks. */
    WDT_PERIOD_2_25_CLKS  = MXC_V_WDT_CTRL_INT_PERIOD_2_25_CLKS, /**< 2^25 WDT clocks. */
    WDT_PERIOD_2_24_CLKS  = MXC_V_WDT_CTRL_INT_PERIOD_2_24_CLKS, /**< 2^24 WDT clocks. */
    WDT_PERIOD_2_23_CLKS  = MXC_V_WDT_CTRL_INT_PERIOD_2_23_CLKS, /**< 2^23 WDT clocks. */
    WDT_PERIOD_2_22_CLKS  = MXC_V_WDT_CTRL_INT_PERIOD_2_22_CLKS, /**< 2^22 WDT clocks. */
    WDT_PERIOD_2_21_CLKS  = MXC_V_WDT_CTRL_INT_PERIOD_2_21_CLKS, /**< 2^21 WDT clocks. */
    WDT_PERIOD_2_20_CLKS  = MXC_V_WDT_CTRL_INT_PERIOD_2_20_CLKS, /**< 2^20 WDT clocks. */
    WDT_PERIOD_2_19_CLKS  = MXC_V_WDT_CTRL_INT_PERIOD_2_19_CLKS, /**< 2^19 WDT clocks. */
    WDT_PERIOD_2_18_CLKS  = MXC_V_WDT_CTRL_INT_PERIOD_2_18_CLKS, /**< 2^18 WDT clocks. */
    WDT_PERIOD_2_17_CLKS  = MXC_V_WDT_CTRL_INT_PERIOD_2_17_CLKS, /**< 2^17 WDT clocks. */
    WDT_PERIOD_2_16_CLKS  = MXC_V_WDT_CTRL_INT_PERIOD_2_16_CLKS,  /**< 2^16 WDT clocks. */
    WDT_PERIOD_MAX                                                /**< Maximum Period is Max - 1 */
} wdt_period_t;

/**
 * @brief   Initializes system level clocks and sets watchdog in a known disabled state
 * @note    The clk_scale in cfg is only used if the system clock is selected for clk.
 *
 * @param   wdt         Pointer to the Watchdog Timer Instance
 * @param   cfg         Watchdog system configuration, see sys_cfg_wdt_t.
 * @param   unlock_key  Watchdog unlock key
 *
 * @retval  E_NO_ERROR  Watchdog Timer initialized as requested
 * @retval  E_NULL_PTR  NULL pointer for Watchdog Timer Instance or Configuration parameters.
 * @retval  E_BAD_STATE Invalid unlock_key, WDT failed to unlock.
 */
int WDT_Init(mxc_wdt_regs_t *wdt, const sys_cfg_wdt_t *cfg, uint8_t unlock_key);

/**
 * @brief   Configures and enables the interrupt timeout for the watchdog specified.
 *
 * @param   wdt         Watchdog module to operate on
 * @param   int_period  Interrupt period as defined by wdt_period_t.
 * @param   unlock_key  Key to unlock watchdog. See MXC_WDT_UNLOCK_KEY.
 *
 * @retval  E_NO_ERROR  Interrupt enabled
 * @retval  E_BAD_STATE Invalid unlock_key, WDT failed to unlock.
 */
int WDT_EnableInt(mxc_wdt_regs_t *wdt, wdt_period_t int_period, uint8_t unlock_key);

/**
 * @brief   Disables the interrupt timeout for the watchdog specified.
 *
 * @param   wdt         Pointer to the Watchdog Timer Instance
 * @param   unlock_key  Key to unlock watchdog. See MXC_WDT_UNLOCK_KEY.
 *
 * @retval  E_NO_ERROR  Interrupt disabled.
 * @retval  E_BAD_STATE Invalid unlock_key, WDT failed to unlock.
 */
int WDT_DisableInt(mxc_wdt_regs_t *wdt, uint8_t unlock_key);

/**
 * @brief   Configures and enables the pre-window timeout for the watchdog specified.
 *
 * @param   wdt         Pointer to the Watchdog Timer Instance
 * @param   wait_period Pre-window period, see wdt_period_t for accepted values.
 * @param   unlock_key  Key to unlock watchdog. See MXC_WDT_UNLOCK_KEY.
 *
 * @retval  E_NO_ERROR  Pre-window timeout set to wait_period
 * @retval  E_BAD_STATE WDT unable to be unlocked
 * @retval  E_INVALID   Requested Period is greater than the maximum supported
 */
int WDT_EnableWait(mxc_wdt_regs_t *wdt, wdt_period_t wait_period, uint8_t unlock_key);

/**
 * @brief Disables the pre-window timeout for the watchdog specified.
 *
 * @param   wdt         Pointer to the Watchdog Timer Instance
 * @param   unlock_key  Key to unlock watchdog. See MXC_WDT_UNLOCK_KEY.
 *
 * @retval  E_NO_ERROR  Wait disabled.
 * @retval  E_BAD_STATE Invalid unlock_key, WDT failed to unlock.
 */
int WDT_DisableWait(mxc_wdt_regs_t *wdt, uint8_t unlock_key);

/**
 * @brief Configures and enables the reset timeout for the watchdog specified.
 *
 * @param   wdt         Pointer to the Watchdog Timer Instance
 * @param   rst_period  Reset period, see wdt_period_t for accepted values.
 * @param   unlock_key  Key to unlock watchdog. See MXC_WDT_UNLOCK_KEY.
 *
 * @retval  E_NO_ERROR  Watchdog Timer Reset enabled with the rst_period time.
 * @retval  E_BAD_STATE Invalid unlock_key, WDT failed to unlock.
 * @retval  E_INVALID   Requested Period is greater than the maximum supported
 */
int WDT_EnableReset(mxc_wdt_regs_t *wdt, wdt_period_t rst_period, uint8_t unlock_key);

/**
 * @brief Disables the reset timeout for the watchdog specified.
 *
 * @param   wdt         Pointer to the Watchdog Timer Instance
 * @param   unlock_key  Key to unlock watchdog. See MXC_WDT_UNLOCK_KEY.
 *
 * @retval  E_NO_ERROR  Reset disabled.
 * @retval  E_BAD_STATE Invalid unlock_key, WDT failed to unlock.
 */
int WDT_DisableReset(mxc_wdt_regs_t *wdt, uint8_t unlock_key);

/**
 * @brief   Gets the watchdog interrupt flags
 *
 * @param   wdt         Pointer to the Watchdog Timer Instance.
 *
 * @retval uint32_t    Value of the Watchdog Timer Flags.
 *
 */
__STATIC_INLINE uint32_t WDT_GetFlags(mxc_wdt_regs_t *wdt)
{
    return (wdt->flags);
}

/**
 * @brief   Clears the watchdog interrupt flags based on the mask
 *
 * @param   wdt     Pointer to the Watchdog Timer Instance
 * @param   mask    Watchdog Flags to clear
 *
 */
__STATIC_INLINE void WDT_ClearFlags(mxc_wdt_regs_t *wdt, uint32_t mask)
{
    wdt->flags = mask;
}

/**
 * @brief Starts the specified Watchdog Timer instance.
 *
 * @param wdt           Pointer to the Watchdog Timer instance
 * @param unlock_key    Key to unlock watchdog.
 *
 * @retval  E_NO_ERROR  Interrupt enabled.
 * @retval  E_BAD_STATE WDT2 Already Running
 * @retval  E_BAD_STATE Invalid unlock_key, WDT failed to unlock.
 *
 */
int WDT_Start(mxc_wdt_regs_t *wdt, uint8_t unlock_key);

/**
 * @brief Feeds the watchdog specified.
 *
 * @param   wdt         Watchdog module to operate on
 *
 */
void WDT_Reset(mxc_wdt_regs_t *wdt);

/**
 * @brief Stops the watchdog specified.
 *
 * @param   wdt         Pointer to the Watchdog Timer Instance
 * @param   unlock_key  Key to unlock watchdog.
 *
 * @retval  E_NO_ERROR  Interrupt enabled.
 * @retval  E_BAD_STATE Invalid unlock_key, WDT failed to unlock.
 */
int WDT_Stop(mxc_wdt_regs_t *wdt, uint8_t unlock_key);

/**
 * @brief   Determines if the watchdog is running
 *
 * @param   wdt         Pointer to the Watchdog Timer Instance
 *
 * @retval  0           Watchdog timer is Disabled.
 * @retval  non-zero    Watchdog timer is Active
 */
__STATIC_INLINE int WDT_IsActive(mxc_wdt_regs_t *wdt)
{
    return (!!(wdt->ctrl & MXC_F_WDT_CTRL_EN_TIMER));
}

/**
 * @}
 */

#ifdef __cplusplus
}
#endif

#endif /* _WDT_H */
