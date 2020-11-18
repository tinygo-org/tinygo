/**
 * @file
 * @brief   Low-Power Mode Management and Configuration API.
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
 * $Date: 2017-02-16 14:41:58 -0600 (Thu, 16 Feb 2017) $
 * $Revision: 26486 $
 *
 *************************************************************************** */

/* **** Includes **** */
#include "gpio.h"
#include "pwrman_regs.h"
#include "pwrseq_regs.h"

/* Define to prevent redundant inclusion */
#ifndef _LP_H_
#define _LP_H_

#ifdef __cplusplus
extern "C" {
#endif

// Doxy group definition for this peripheral module
/**
 * @ingroup syscfg
 * @defgroup lp_lib Low-Power Management
 * @brief Power Mode Configuration and Management API
 * @{
 */

/* **** Definitions **** */
/**
 * @brief   Enumerations for pull-up and pull-down types for GPIO pins.
 */
typedef enum {
    LP_WEAK_PULL_DOWN = -1,     /**< Weak pull-down enabled during low-power mode.*/
    LP_NO_PULL = 0,             /**< No pull-up/pull-down during low-power mode.*/
    LP_WEAK_PULL_UP = 1         /**< Weak pull-up during low-power mode. */
}
lp_pu_pd_select_t;

/* **** Function Prototypes **** */

/**
 * @brief      Gets the first boot flag
 *
 * @retval     0     FIRST_BOOT was not set
 * @retval     1     FIRST_BOOT <em><b>set</b></em>
 */
__STATIC_INLINE unsigned int LP_IsFirstBoot()
{
    return ((MXC_PWRSEQ->reg0 & MXC_F_PWRSEQ_REG0_PWR_FIRST_BOOT) >> MXC_F_PWRSEQ_REG0_PWR_FIRST_BOOT_POS);
}

/**
 * @brief      Clears the first boot flag
 */
__STATIC_INLINE void LP_ClearFirstBoot()
{
    MXC_PWRSEQ->reg0 &= ~MXC_F_PWRSEQ_REG0_PWR_FIRST_BOOT;
}

/**
 * @brief      Determines if the application woke up from LP0
 *
 * @return     0  Application <em><b>was not</b></em> woken up from LP0.
 * @return     1  Application <em><b>was</b></em> woken up from
 */
__STATIC_INLINE unsigned int LP_IsLP0WakeUp()
{
    //POR should be set and first boot clear
    if((MXC_PWRMAN->pwr_rst_ctrl & MXC_F_PWRMAN_PWR_RST_CTRL_POR) &&
            ((MXC_PWRSEQ->reg0 & MXC_F_PWRSEQ_REG0_PWR_FIRST_BOOT) == 0))
        return 1;
    else
        return 0;

}

/**
 * @brief   Returns the Low-Power Wake-up Flags.
 * @returns    The wake-up flags register value. 
 */
__STATIC_INLINE unsigned int LP_GetWakeUpFlags(void)
{
    return MXC_PWRSEQ->flags;
}
/**
 * @brief      Clear \e ALL wake-up configuration on all pins. Disables wake-up
 *             entirely.
 */
void LP_ClearWakeUpConfig(void);

/**
 * @brief      Reads wake-up flags, clears all wake-up flags, and returns read
 *             flags to the caller.
 * @return     Wake-up flags from Power Sequencer
 */
unsigned int LP_ClearWakeUpFlags(void);

/**
 * @brief      This function configures one GPIO pin to wake the processor from
 *             LP0 or LP1. It is not used for LP2 wake-up, as normal GPIO
 *             interrupt processing is active in that mode.
 * @param      gpio      GPIO pointer to a #gpio_cfg_t object describing the
 *                       port and pin for selected wake-up source.
 * @param      act_high  @arg Non-zero for Active High Wake-Up @arg 0 for Active Low Wake-Up
 * @param      wk_pu_pd  Selection for the 1 Meg ohm pull-up or pull-down on
 *                       this pin, see #lp_pu_pd_select_t
 * @return     #E_NO_ERROR if the GPIO pin is configured successfully for
 *             Wake-Up detection, @ref MXC_Error_Codes "Error Code" if
 *             unsuccessful.
 */
int LP_ConfigGPIOWakeUpDetect(const gpio_cfg_t *gpio, unsigned int act_high, lp_pu_pd_select_t wk_pu_pd);

/**
 * @brief      Clear the wake-up configuration for a specific GPIO pin.
 * @param      gpio  GPIO pointer to a #gpio_cfg_t object describing the port and pin for selected
 *                   wake-up source
 * @return     #E_NO_ERROR if the wake-up configuration is cleared for the @p
 *             gpio pin, @ref MXC_Error_Codes "Error Code" if unsuccessful.
 */
int LP_ClearGPIOWakeUpDetect(const gpio_cfg_t *gpio);

/**
 * @brief      Check if a specific gpio triggered the wake up
 * @param      gpio      GPIO pointer to a #gpio_cfg_t object describing the
 *                       port and pin(s)
 * @retval     0         @p gpio did not trigger the wake up.
 * @retval     Non-zero  At least one of the @p gpio triggered wake up event,
 *                       the bit position set in the return value indicates
 *                       which pin caused the wake-up event.
 */
uint8_t LP_IsGPIOWakeUpSource(const gpio_cfg_t *gpio);

/**
 * @brief      Wake on USB plug or unplug
 * @param      plug_en    set to 1 to enable wake-up when USB VBUS is detected
 * @param      unplug_en  set to 1 to enable wake-up when USB VBUS disappears
 * @return     #E_NO_ERROR  if configured successfully, @ref MXC_Error_Codes
 *             "Error Code" if unsuccessful.
 */
int LP_ConfigUSBWakeUp(unsigned int plug_en, unsigned int unplug_en);

/**
 * @brief      Wake on any enabled event signal from the @ref rtc "RTC".
 * @param      comp0_en         set to 1 to enable wake-up when RTC Comparison 0
 *                              is set
 * @param      comp1_en         set to 1 to enable wake-up when RTC Comparison 1
 *                              is set
 * @param      prescale_cmp_en  set to 1 to enable wake-up when RTC Prescaler
 *                              Compare is set
 * @param      rollover_en      set to 1 to enable wake-up when RTC Roll-over is
 *                              set
 * @return     #E_NO_ERROR if successfully configured, @ref MXC_Error_Codes
 *             "Error Code" if unsuccessful.
 */
int LP_ConfigRTCWakeUp(unsigned int comp0_en, unsigned int comp1_en, unsigned int prescale_cmp_en, unsigned int rollover_en);

/**
 * @brief      Enter power-saving Low-Power Mode 2 (LP2).
 * @return     #E_NO_ERROR on success, @ref MXC_Error_Codes "Error Code" if
 *             unsuccessful.
 */
int LP_EnterLP2(void);

/**
 * @brief      Enter Low-Power 1 (LP1) mode, which saves the CPU's state and
 *             SRAM.
 * @note       Execution resumes with the return from this function call when
 *             the CPU wakes-up from LP1 mode.
 * @note       Interrupts should be globally disabled before calling this
 *             function.
 * @return     #E_NO_ERROR on success, @ref MXC_Error_Codes "Error Code" if
 *             unsuccessful.
 */
int LP_EnterLP1(void);

/**
 * @brief      Enter Low-Power 0 (LP0) mode, the lowest-possible power mode. 
 * @note 
 * @par Important notes for LP0 mode entry and exit 
 * @parblock
 *             @arg SRAM contents are lost
 *             @arg Waking up from LP0 is similar to a system reset, code execution begins from address 0
 *             @arg <em><b>This function does not return.</b></em>
 * @endparblock
 * @note       Interrupts are globally disabled upon entering this function.
 */
void LP_EnterLP0(void);
/**@}*/
#ifdef __cplusplus
}
#endif

#endif /* _LP_H_ */
