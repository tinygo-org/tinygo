/**
 * @file
 * @brief   Low Power management API implementation. 
 */
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
 * $Date: 2017-02-16 14:50:14 -0600 (Thu, 16 Feb 2017) $
 * $Revision: 26488 $
 * **************************************************************************** */

/* **** Includes **** */
#include "mxc_config.h"
#include "mxc_assert.h"
#include "lp.h"
#include "ioman_regs.h"

/**
 * @ingroup lp_lib
 * @{
 */

/* **** Definitions **** */

#ifndef LP0_PRE_HOOK
#define LP0_PRE_HOOK        /**< Performance-measurement hook, may be defined as nothing */
#endif
#ifndef LP1_PRE_HOOK
#define LP1_PRE_HOOK        /**< Performance-measurement hook, may be defined as nothing */
#endif
#ifndef LP1_POST_HOOK
#define LP1_POST_HOOK       /**< Performance-measurement hook, may be defined as nothing */
#endif

/* **** Globals **** */

/* **** Functions **** */

/* ************************************************************************** */
void LP_ClearWakeUpConfig(void)
{
    /* Function clears all wake-up configuration */
    /* Clear GPIO WUD event and configuration registers, globally */
    MXC_PWRSEQ->reg1 |= (MXC_F_PWRSEQ_REG1_PWR_CLR_IO_EVENT_LATCH |
                         MXC_F_PWRSEQ_REG1_PWR_CLR_IO_CFG_LATCH);
    MXC_PWRSEQ->reg1 &= ~(MXC_F_PWRSEQ_REG1_PWR_CLR_IO_EVENT_LATCH |
                          MXC_F_PWRSEQ_REG1_PWR_CLR_IO_CFG_LATCH);

    /* Mask off all wake-up sources */
    MXC_PWRSEQ->msk_flags &= ~(MXC_F_PWRSEQ_MSK_FLAGS_PWR_IOWAKEUP |
                               MXC_F_PWRSEQ_MSK_FLAGS_PWR_USB_PLUG_WAKEUP |
                               MXC_F_PWRSEQ_MSK_FLAGS_PWR_USB_REMOVE_WAKEUP |
                               MXC_F_PWRSEQ_MSK_FLAGS_RTC_CMPR0 |
                               MXC_F_PWRSEQ_MSK_FLAGS_RTC_CMPR1 |
                               MXC_F_PWRSEQ_MSK_FLAGS_RTC_PRESCALE_CMP |
                               MXC_F_PWRSEQ_MSK_FLAGS_RTC_ROLLOVER);
}


/* ************************************************************************** */
unsigned int LP_ClearWakeUpFlags(void)
{
    /* Function Clears wake-up flags */
    unsigned int flags_tmp;

    /* Get flags */
    flags_tmp = MXC_PWRSEQ->flags;

    /* Clear GPIO WUD event registers, globally */
    MXC_PWRSEQ->reg1 |= (MXC_F_PWRSEQ_REG1_PWR_CLR_IO_EVENT_LATCH);
    MXC_PWRSEQ->reg1 &= ~(MXC_F_PWRSEQ_REG1_PWR_CLR_IO_EVENT_LATCH);

    /* Clear power sequencer event flags (write-1-to-clear) */
    MXC_PWRSEQ->flags = flags_tmp;

    return flags_tmp;
}

/* ************************************************************************** */
int LP_ConfigGPIOWakeUpDetect(const gpio_cfg_t *gpio, unsigned int act_high, lp_pu_pd_select_t wk_pu_pd)
{
    /* Function configures the selected pin for wake-up detect */
    int result = E_NO_ERROR;
    unsigned int pin;

    /* Check that port and pin are within range */
    MXC_ASSERT(gpio->port < MXC_GPIO_NUM_PORTS);
    MXC_ASSERT(gpio->mask > 0);

    /* Ports 0-3 are controlled by wud_req0, while 4-7 are controlled by wud_req1, 8 is controlled by wud_req2 */
    if (gpio->port < 4) {
        MXC_IOMAN->wud_req0 |= (gpio->mask << (gpio->port << 3));
        if (MXC_IOMAN->wud_ack0 != MXC_IOMAN->wud_req0) { /* Order of volatile access does not matter here */
            result = E_BUSY;
        }
    } else if (gpio->port < 8) {
        MXC_IOMAN->wud_req1 |= (gpio->mask << ((gpio->port - 4) << 3));
        if (MXC_IOMAN->wud_ack1 != MXC_IOMAN->wud_req1) { /* Order of volatile access does not matter here */
            result = E_BUSY;
        }
    } else {
        MXC_IOMAN->wud_req2 |= (gpio->mask << ((gpio->port - 8) << 3));
        if (MXC_IOMAN->wud_ack2 != MXC_IOMAN->wud_req2) { /* Order of volatile access does not matter here */
            result = E_BUSY;
        }
    }

    if (result == E_NO_ERROR) {

        for (pin = 0; pin < MXC_GPIO_MAX_PINS_PER_PORT; pin++) {

            if (gpio->mask & (1 << pin)) {

                /* Enable modifications to WUD configuration */
                MXC_PWRMAN->wud_ctrl = MXC_F_PWRMAN_WUD_CTRL_CTRL_ENABLE;

                /* Select pad in WUD control */
                /* Note: Pads are numbered from 0-48; {0-7} => {P0.0-P0.7}, {8-15} => {P1.0-P1.7}, etc. */
                MXC_PWRMAN->wud_ctrl |= (gpio->port * 8) + pin;

                /* Configure sense level on this pad */
                MXC_PWRMAN->wud_ctrl |= (MXC_E_PWRMAN_PAD_MODE_ACT_HI_LO << MXC_F_PWRMAN_WUD_CTRL_PAD_MODE_POS);

                if (act_high) {
                    /* Select active high with PULSE0 (backwards from what you'd expect) */
                    MXC_PWRMAN->wud_pulse0 = 1;
                } else {
                    /* Select active low with PULSE1 (backwards from what you'd expect) */
                    MXC_PWRMAN->wud_pulse1 = 1;
                }

                /* Clear out the pad mode */
                MXC_PWRMAN->wud_ctrl &= ~(MXC_F_PWRMAN_WUD_CTRL_PAD_MODE);

                /* Select this pad to have the wake-up function enabled */
                MXC_PWRMAN->wud_ctrl |= (MXC_E_PWRMAN_PAD_MODE_CLEAR_SET << MXC_F_PWRMAN_WUD_CTRL_PAD_MODE_POS);

                /* Activate with PULSE1 */
                MXC_PWRMAN->wud_pulse1 = 1;

                if (wk_pu_pd != LP_NO_PULL) {
                    /* Select weak pull-up/pull-down on this pad while in LP1 */
                    MXC_PWRMAN->wud_ctrl |= (MXC_E_PWRMAN_PAD_MODE_WEAK_HI_LO << MXC_F_PWRMAN_WUD_CTRL_PAD_MODE_POS);

                    /* Again, logic is opposite of what you'd expect */
                    if (wk_pu_pd == LP_WEAK_PULL_UP) {
                        MXC_PWRMAN->wud_pulse0 = 1;
                    } else {
                        MXC_PWRMAN->wud_pulse1 = 1;
                    }
                }

                /* Disable configuration each time, required by hardware */
                MXC_PWRMAN->wud_ctrl = 0;
            }
        }
    }

    /* Disable configuration */
    MXC_IOMAN->wud_req0 = 0;
    MXC_IOMAN->wud_req1 = 0;
    MXC_IOMAN->wud_req2 = 0;

    /* Enable IOWakeup, as there is at least 1 GPIO pin configured as a wake source */
    MXC_PWRSEQ->msk_flags |= MXC_F_PWRSEQ_MSK_FLAGS_PWR_IOWAKEUP;

    return result;
}

/* ************************************************************************** */
uint8_t LP_IsGPIOWakeUpSource(const gpio_cfg_t *gpio)
{
    uint8_t gpioWokeUp = 0;

    /* Check that port and pin are within range */
    MXC_ASSERT(gpio->port < MXC_GPIO_NUM_PORTS);
    MXC_ASSERT(gpio->mask > 0);

    /* Ports 0-3 are wud_seen0, while 4-7 are wud_seen1, 8 is wud_seen2 */
    if (gpio->port < 4) {
        gpioWokeUp = (MXC_PWRMAN->wud_seen0 >> (gpio->port << 3)) & gpio->mask;
    } else if (gpio->port < 8) {
        gpioWokeUp = (MXC_PWRMAN->wud_seen1 >> ((gpio->port - 4) << 3)) & gpio->mask;
    } else {
        gpioWokeUp = (MXC_PWRMAN->wud_seen2 >> ((gpio->port - 8) << 3)) & gpio->mask;
    }

    return gpioWokeUp;
}

/* ************************************************************************** */
int LP_ClearGPIOWakeUpDetect(const gpio_cfg_t *gpio)
{
    int result = E_NO_ERROR;
    unsigned int pin;

    /* Check that port and pin are within range */
    MXC_ASSERT(gpio->port < MXC_GPIO_NUM_PORTS);
    MXC_ASSERT(gpio->mask > 0);

    /* Ports 0-3 are controlled by wud_req0, while 4-7 are controlled by wud_req1, 8 is controlled by wud_req2 */
    if (gpio->port < 4) {
        MXC_IOMAN->wud_req0 |= (gpio->mask << (gpio->port << 3));
        if (MXC_IOMAN->wud_ack0 != MXC_IOMAN->wud_req0) { /* Order of volatile access does not matter here */
            result = E_BUSY;
        }
    } else if (gpio->port < 8) {
        MXC_IOMAN->wud_req1 |= (gpio->mask << ((gpio->port - 4) << 3));
        if (MXC_IOMAN->wud_ack1 != MXC_IOMAN->wud_req1) { /* Order of volatile access does not matter here */
            result = E_BUSY;
        }
    } else {
        MXC_IOMAN->wud_req2 |= (gpio->mask << ((gpio->port - 8) << 3));
        if (MXC_IOMAN->wud_ack2 != MXC_IOMAN->wud_req2) { /* Order of volatile access does not matter here */
            result = E_BUSY;
        }
    }

    if (result == E_NO_ERROR) {
        for (pin = 0; pin < MXC_GPIO_MAX_PINS_PER_PORT; pin++) {
            if (gpio->mask & (1 << pin)) {

                /* Enable modifications to WUD configuration */
                MXC_PWRMAN->wud_ctrl = MXC_F_PWRMAN_WUD_CTRL_CTRL_ENABLE;

                /* Select pad in WUD control */
                /* Note: Pads are numbered from 0-48; {0-7} => {P0.0-P0.7}, {8-15} => {P1.0-P1.7}, etc. */
                MXC_PWRMAN->wud_ctrl |= (gpio->port * 8) + pin;

                /* Clear out the pad mode */
                MXC_PWRMAN->wud_ctrl &= ~(MXC_F_PWRMAN_WUD_CTRL_PAD_MODE);

                /* Select the wake up function on this pad */
                MXC_PWRMAN->wud_ctrl |= (MXC_E_PWRMAN_PAD_MODE_CLEAR_SET << MXC_F_PWRMAN_WUD_CTRL_PAD_MODE_POS);

                /* disable wake up with PULSE0 */
                MXC_PWRMAN->wud_pulse0 = 1;

                /* Disable configuration each time, required by hardware */
                MXC_PWRMAN->wud_ctrl = 0;
            }
        }
    }

    /* Disable configuration */
    MXC_IOMAN->wud_req0 = 0;
    MXC_IOMAN->wud_req1 = 0;
    MXC_IOMAN->wud_req2 = 0;

    return result;
}

/* ************************************************************************** */
int LP_ConfigUSBWakeUp(unsigned int plug_en, unsigned int unplug_en)
{
    /* Enable or disable wake on USB plug-in */
    if (plug_en) {
        MXC_PWRSEQ->msk_flags |= MXC_F_PWRSEQ_MSK_FLAGS_PWR_USB_PLUG_WAKEUP;
    } else {
        MXC_PWRSEQ->msk_flags &= ~(MXC_F_PWRSEQ_MSK_FLAGS_PWR_USB_PLUG_WAKEUP);
    }

    /* Enable or disable wake on USB unplug */
    if (unplug_en) {
        MXC_PWRSEQ->msk_flags |= MXC_F_PWRSEQ_MSK_FLAGS_PWR_USB_REMOVE_WAKEUP;
    } else {
        MXC_PWRSEQ->msk_flags &= ~(MXC_F_PWRSEQ_MSK_FLAGS_PWR_USB_REMOVE_WAKEUP);
    }

    return E_NO_ERROR;
}

/* ************************************************************************** */
int LP_ConfigRTCWakeUp(unsigned int comp0_en, unsigned int comp1_en,
                       unsigned int prescale_cmp_en, unsigned int rollover_en)
{
    /* Note: MXC_PWRSEQ.pwr_misc[0] should be set to have the mask be active low */

    /* Enable or disable wake on RTC Compare 0 */
    if (comp0_en) {
        MXC_PWRSEQ->msk_flags |= MXC_F_PWRSEQ_FLAGS_RTC_CMPR0;

    } else {
        MXC_PWRSEQ->msk_flags &= ~(MXC_F_PWRSEQ_FLAGS_RTC_CMPR0);
    }

    /* Enable or disable wake on RTC Compare 1 */
    if (comp1_en) {
        MXC_PWRSEQ->msk_flags |= MXC_F_PWRSEQ_FLAGS_RTC_CMPR1;

    } else {
        MXC_PWRSEQ->msk_flags &= ~(MXC_F_PWRSEQ_FLAGS_RTC_CMPR1);
    }

    /* Enable or disable wake on RTC Prescaler */
    if (prescale_cmp_en) {
        MXC_PWRSEQ->msk_flags |= MXC_F_PWRSEQ_FLAGS_RTC_PRESCALE_CMP;

    } else {
        MXC_PWRSEQ->msk_flags &= ~(MXC_F_PWRSEQ_FLAGS_RTC_PRESCALE_CMP);
    }

    /* Enable or disable wake on RTC Rollover */
    if (rollover_en) {
        MXC_PWRSEQ->msk_flags |= MXC_F_PWRSEQ_FLAGS_RTC_ROLLOVER;

    } else {
        MXC_PWRSEQ->msk_flags &= ~(MXC_F_PWRSEQ_FLAGS_RTC_ROLLOVER);
    }

    return E_NO_ERROR;
}

/* ************************************************************************** */
int LP_EnterLP2(void)
{
    /* Clear SLEEPDEEP bit to avoid LP1/LP0 entry*/
    SCB->SCR &= ~SCB_SCR_SLEEPDEEP_Msk;

    /* Go into LP2 mode and wait for an interrupt to wake the processor */
    __WFI();

    return E_NO_ERROR;
}

/* ************************************************************************** */
int LP_EnterLP1(void)
{
    /* Turn on retention controller */
    MXC_PWRSEQ->retn_ctrl0 |= MXC_F_PWRSEQ_RETN_CTRL0_RETN_CTRL_EN;

    /* Clear the firstboot bit, which is generated by a POR event and locks out LPx modes */
    MXC_PWRSEQ->reg0 &= ~(MXC_F_PWRSEQ_REG0_PWR_FIRST_BOOT);

    /* Set the LP1 select bit so CPU goes to LP1 during SLEEPDEEP */
    MXC_PWRSEQ->reg0 |= MXC_F_PWRSEQ_REG0_PWR_LP1;

    /* The SLEEPDEEP bit will cause a WFE() to trigger LP0/LP1 (depending on ..._REG0_PWR_LP1 state) */
    SCB->SCR |= SCB_SCR_SLEEPDEEP_Msk;

    /* Performance-measurement hook, may be defined as nothing */
    LP1_PRE_HOOK;
    
    /* Freeze GPIO using MBUS so that it doesn't change while digital core is alseep */
    MXC_PWRSEQ->reg1 |= MXC_F_PWRSEQ_REG1_PWR_MBUS_GATE;

    /* Dummy read to make sure SSB writes are complete */
    MXC_PWRSEQ->reg0;

    /* Enter LP1 -- sequence is per instructions from ARM, Ltd. */
    __SEV();
    __WFE();
    __WFE();

    /* Performance-measurement hook, may be defined as nothing */
    LP1_POST_HOOK;
    
    /* Unfreeze the GPIO by clearing MBUS_GATE (always safe to do) */
    MXC_PWRSEQ->reg1 &= ~(MXC_F_PWRSEQ_REG1_PWR_MBUS_GATE);

    /* Clear SLEEPDEEP bit */
    SCB->SCR &= ~SCB_SCR_SLEEPDEEP_Msk;

    /* No error */
    return E_NO_ERROR;
}
/* ************************************************************************** */
void LP_EnterLP0(void)
{
    /* Disable interrupts, ok not to save state as exit LP0 is a reset */
    __disable_irq();

    /* Clear the firstboot bit, which is generated by a POR event and locks out LPx modes */
    MXC_PWRSEQ->reg0 &= ~(MXC_F_PWRSEQ_REG0_PWR_FIRST_BOOT);

    /* Turn off retention controller */
    MXC_PWRSEQ->retn_ctrl0 &= ~(MXC_F_PWRSEQ_RETN_CTRL0_RETN_CTRL_EN);

    /* Turn off retention regulator */
    MXC_PWRSEQ->reg0 &= ~(MXC_F_PWRSEQ_REG0_PWR_RETREGEN_RUN | MXC_F_PWRSEQ_REG0_PWR_RETREGEN_SLP);

    /* LP0 ONLY to eliminate ~50nA of leakage on VDD12 */
    MXC_PWRSEQ->reg1 |= MXC_F_PWRSEQ_REG1_PWR_SRAM_NWELL_SW;

    /* Clear the LP1 select bit so CPU goes to LP0 during SLEEPDEEP */
    MXC_PWRSEQ->reg0 &= ~(MXC_F_PWRSEQ_REG0_PWR_LP1);

    /* The SLEEPDEEP bit will cause a WFE() to trigger LP0/LP1 (depending on ..._REG0_PWR_LP1 state) */
    SCB->SCR |= SCB_SCR_SLEEPDEEP_Msk;

    /* Performance-measurement hook, may be defined as nothing */
    LP0_PRE_HOOK;
    
    /* Freeze GPIO using MBUS so that it doesn't change while digital core is alseep */
    MXC_PWRSEQ->reg1 |= MXC_F_PWRSEQ_REG1_PWR_MBUS_GATE;

    /* Dummy read to make sure SSB writes are complete */
    MXC_PWRSEQ->reg0;

    /* Go into LP0 -- sequence is per instructions from ARM, Ltd. */
    __SEV();
    __WFE();
    __WFE();

    /* Catch the case where this code does not properly sleep */
    /* Unfreeze the GPIO by clearing MBUS_GATE (always safe to do) */
    MXC_PWRSEQ->reg1 &= ~(MXC_F_PWRSEQ_REG1_PWR_MBUS_GATE);
    MXC_ASSERT_FAIL();
    while (1) {
        __NOP();
    }

    /* Does not actually return */
}
/**@} end of ingroup lp_lib */
