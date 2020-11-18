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
 * $Date: 2016-03-23 10:22:39 -0500 (Wed, 23 Mar 2016) $
 * $Revision: 22053 $
 *
 ******************************************************************************/

/**
 * @file  lp.h
 * @brief This is the high level API for the Lower Power
 */

#ifndef _LP_H_
#define _LP_H_

#include "gpio.h"
#include "pwrman_regs.h"
#include "pwrseq_regs.h"

#ifdef __cplusplus
extern "C" {
#endif

/***** Definitions *****/
/**
 * @brief   Enumerations for pull-up and pull-downs
 *
 */
typedef enum {
    LP_WEAK_PULL_DOWN = -1,
    LP_NO_PULL = 0,
    LP_WEAK_PULL_UP = 1
}
lp_pu_pd_select_t;

/***** Function Prototypes *****/

/**
 * @brief   Gets the first boot flag
 *
 * @returns 0 if FIRST_BOOT was not set, or 1 if FIRST_BOOT was set
 */
__STATIC_INLINE unsigned int LP_IsFirstBoot()
{
    return ((MXC_PWRSEQ->reg0 & MXC_F_PWRSEQ_REG0_PWR_FIRST_BOOT) >> MXC_F_PWRSEQ_REG0_PWR_FIRST_BOOT_POS);
}

/**
 * @brief   Clears the first boot flag
 *
 */
__STATIC_INLINE void LP_ClearFirstBoot()
{
    MXC_PWRSEQ->reg0 &= ~MXC_F_PWRSEQ_REG0_PWR_FIRST_BOOT;
}

/**
 * @brief   Determines of program woke up from LP0
 *
 * @returns 0 if not woken up from LP0, or 1 if woken from LP0
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

__STATIC_INLINE unsigned int LP_GetWakeUpFlags(void)
{
    return MXC_PWRSEQ->flags;
}
/**
 * @brief   Clear ALL wake-up configuration on all pins. Disables wake-up entirely.
 */
void LP_ClearWakeUpConfig(void);

/**
 * @brief   Read wake-up flags, clear flags, and return to caller.
 * @returns Wake-up flags from Power Sequencer
 */
unsigned int LP_ClearWakeUpFlags(void);

/**
 * @brief   This function configures one GPIO pin to wake the processor from LP0 or LP1.
 *          It is not used for LP2 wake-up, as normal GPIO interrupt processing is active in that mode.
 * @param   gpio      GPIO pointer describing the port and pin for selected wake-up source
 * @param   act_high If non-zero, the signal is configured for active high wake-up. Otherwise, active low.
 * @param   wk_pu_pd Selection for the 1 Meg ohm pull-up or pull-down on this pin, see #lp_pu_pd_select_t
 * @returns #E_NO_ERROR on success, error if unsuccessful.
 */
int LP_ConfigGPIOWakeUpDetect(const gpio_cfg_t *gpio, unsigned int act_high, lp_pu_pd_select_t wk_pu_pd);

/**
 * @brief   Clear the wake-up configuration on one specific GPIO pin
 * @param   gpio      GPIO pointer describing the port and pin for selected wake-up source
 * @returns #E_NO_ERROR on success, error if unsuccessful.
 */
int LP_ClearGPIOWakeUpDetect(const gpio_cfg_t *gpio);

/**
 * @brief   Check if a specific gpio triggered the wake up
 * @param   gpio      GPIO pointer describing the port and pin(s)
 * @returns 0 = gpio passed in did not trigger a wake up
 *          nonzero = at least one of the gpio passed in triggered a wake up
 *                    the bit set represents which pin is the wake up source
 */
uint8_t LP_IsGPIOWakeUpSource(const gpio_cfg_t *gpio);

/**
 * @brief   Wake on USB plug or unplug
 * @param   plug_en   set to 1 to enable wake-up when USB VBUS is detected
 * @param   unplug_en set to 1 to enable wake-up when USB VBUS disappears
 * @returns #E_NO_ERROR on success, error if unsuccessful.
 */
int LP_ConfigUSBWakeUp(unsigned int plug_en, unsigned int unplug_en);

/**
 * @brief   Wake on any enabled event signal from RTC
 * @param   comp0_en  set to 1 to enable wake-up when RTC Comparison 0 is set
 * @param   comp1_en  set to 1 to enable wake-up when RTC Comparison 1 is set
 * @param   prescale_cmp_en  set to 1 to enable wake-up when RTC Prescaler Compare is set
 * @param   rollover_en  set to 1 to enable wake-up when RTC Roll-over is set
 * @returns #E_NO_ERROR on success, error if unsuccessful.
 */
int LP_ConfigRTCWakeUp(unsigned int comp0_en, unsigned int comp1_en, unsigned int prescale_cmp_en, unsigned int rollover_en);

/**
 * @brief   Enter LP2 power-saving mode
 * @returns #E_NO_ERROR on success, error if unsuccessful.
 */
int LP_EnterLP2(void);

/**
 * @brief   Enter LP1 mode, which saves CPU state and SRAM. Execution resumes after this call.
 * @note    Interrupts should be globally disabled before calling this function.
 * @returns #E_NO_ERROR on success, error if unsuccessful.
 */
int LP_EnterLP1(void);

/**
 * @brief   Enter the lowest-possible power mode, known as LP0. SRAM contents are lost.
 *          Waking up from LP0 is like a system reset. This function does not return.
 * @note    Interrupts are globally disabled upon entering this function.
 */
void LP_EnterLP0(void);

#ifdef __cplusplus
}
#endif

#endif /* _LP_H_ */
