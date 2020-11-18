/*******************************************************************************
 * Copyright (C) 2017 Maxim Integrated Products, Inc., All Rights Reserved.
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
 * $Date: 2017-03-01 13:26:51 -0600 (Wed, 01 Mar 2017) $
 * $Revision: 26782 $
 *
 ******************************************************************************/

#include <string.h>
#include <stdio.h>
#include <stdlib.h>
#include "max3263x.h"
#include "clkman_regs.h"
#include "adc_regs.h"
#include "pwrseq_regs.h"
#include "pwrman_regs.h"
#include "icc_regs.h"
#include "flc_regs.h"
#include "rtc_regs.h"
#include "trim_regs.h"

#ifndef RO_FREQ
#define RO_FREQ     96000000
#endif

#ifndef LP0_POST_HOOK
#define LP0_POST_HOOK
#endif

extern void (* const __isr_vector[])(void);
uint32_t SystemCoreClock = RO_FREQ/2;

void SystemCoreClockUpdate(void)
{
#ifdef EMULATOR
    SystemCoreClock = RO_FREQ;
#else /* real hardware */
    if(MXC_PWRSEQ->reg0 & MXC_F_PWRSEQ_REG0_PWR_RCEN_RUN) {
        /* 4 MHz source */
        if(MXC_PWRSEQ->reg3 & MXC_F_PWRSEQ_REG3_PWR_RC_DIV) {
            SystemCoreClock = (4000000 / (0x1 << ((MXC_PWRSEQ->reg3 & MXC_F_PWRSEQ_REG3_PWR_RC_DIV) >>
                MXC_F_PWRSEQ_REG3_PWR_RC_DIV_POS)));
        } else {
            SystemCoreClock = 4000000;
        }
    } else {
        /* 96 MHz source */
        if(MXC_PWRSEQ->reg3 & MXC_F_PWRSEQ_REG3_PWR_RO_DIV) {
            SystemCoreClock = (RO_FREQ / (0x1 << ((MXC_PWRSEQ->reg3 & MXC_F_PWRSEQ_REG3_PWR_RO_DIV) >>
                MXC_F_PWRSEQ_REG3_PWR_RO_DIV_POS)));
        } else {
            SystemCoreClock = RO_FREQ;
        }
    }
#endif
}

void CLKMAN_TrimRO(void)
{
    uint32_t running;
    uint32_t trim;

    /* Step 1: enable 32KHz RTC */
    running = MXC_PWRSEQ->reg0 & MXC_F_PWRSEQ_REG0_PWR_RTCEN_RUN;
    MXC_PWRSEQ->reg0 |= MXC_F_PWRSEQ_REG0_PWR_RTCEN_RUN;

    /* Wait for RTC warm-up */
    while(MXC_RTCCFG->osc_ctrl & MXC_F_RTC_OSC_CTRL_OSC_WARMUP_ENABLE) {}

    /* Step 2: enable RO calibration complete interrupt */
    MXC_ADC->intr |= MXC_F_ADC_INTR_RO_CAL_DONE_IE;

    /* Step 3: clear RO calibration complete interrupt */
    MXC_ADC->intr |= MXC_F_ADC_INTR_RO_CAL_DONE_IF;

    /* Step 4: -- NO LONGER NEEDED / HANDLED BY STARTUP CODE -- */

    /* Step 5: write initial trim to frequency calibration initial condition register */
    trim = (MXC_PWRSEQ->reg6 & MXC_F_PWRSEQ_REG6_PWR_TRIM_OSC_VREF) >> MXC_F_PWRSEQ_REG6_PWR_TRIM_OSC_VREF_POS;
    MXC_ADC->ro_cal1 = (MXC_ADC->ro_cal1 & ~MXC_F_ADC_RO_CAL1_TRM_INIT) |
                       ((trim << MXC_F_ADC_RO_CAL1_TRM_INIT_POS) & MXC_F_ADC_RO_CAL1_TRM_INIT);

    /* Step 6: load initial trim to active frequency trim register */
    MXC_ADC->ro_cal0 |= MXC_F_ADC_RO_CAL0_RO_CAL_LOAD;

    /* Step 7: enable frequency loop to control RO trim */
    MXC_ADC->ro_cal0 |= MXC_F_ADC_RO_CAL0_RO_CAL_EN;

    /* Step 8: run frequency calibration in atomic mode */
    MXC_ADC->ro_cal0 |= MXC_F_ADC_RO_CAL0_RO_CAL_ATOMIC;

    /* Step 9: waiting for ro_cal_done flag */
    while(!(MXC_ADC->intr & MXC_F_ADC_INTR_RO_CAL_DONE_IF));

    /* Step 10: stop frequency calibration */
    MXC_ADC->ro_cal0 &= ~MXC_F_ADC_RO_CAL0_RO_CAL_RUN;

    /* Step 11: disable RO calibration complete interrupt */
    MXC_ADC->intr &= ~MXC_F_ADC_INTR_RO_CAL_DONE_IE;

    /* Step 12: read final frequency trim value */
    trim = (MXC_ADC->ro_cal0 & MXC_F_ADC_RO_CAL0_RO_TRM) >> MXC_F_ADC_RO_CAL0_RO_TRM_POS;

    /* Step 13: write final trim to RO flash trim shadow register */
    MXC_PWRSEQ->reg6 = (MXC_PWRSEQ->reg6 & ~MXC_F_PWRSEQ_REG6_PWR_TRIM_OSC_VREF) |
                       ((trim << MXC_F_PWRSEQ_REG6_PWR_TRIM_OSC_VREF_POS) & MXC_F_PWRSEQ_REG6_PWR_TRIM_OSC_VREF);

    /* Step 14: restore RTC status */
    if (!running) {
        MXC_PWRSEQ->reg0 &= ~MXC_F_PWRSEQ_REG0_PWR_RTCEN_RUN;
    }

    /* Step 15: disable frequency loop to control RO trim */
    MXC_ADC->ro_cal0 &= ~MXC_F_ADC_RO_CAL0_RO_CAL_EN;
}

static void ICC_Enable(void)
{
    /* Invalidate cache and wait until ready */
    MXC_ICC->invdt_all = 1;
    while (!(MXC_ICC->ctrl_stat & MXC_F_ICC_CTRL_STAT_READY));

    /* Enable cache */
    MXC_ICC->ctrl_stat |= MXC_F_ICC_CTRL_STAT_ENABLE;

    /* Must invalidate a second time for proper use */
    MXC_ICC->invdt_all = 1;
}

/* This function is called before C runtime initialization and can be
 * implemented by the application for early initializations. If a value other
 * than '0' is returned, the C runtime initialization will be skipped.
 *
 * You may over-ride this function in your program by defining a custom
 *  PreInit(), but care should be taken to reproduce the initilization steps
 *  or a non-functional system may result.
 */
__weak int PreInit(void)
{
    /* Increase system clock to 96 MHz */
    MXC_CLKMAN->clk_ctrl = MXC_V_CLKMAN_CLK_CTRL_SYSTEM_SOURCE_SELECT_96MHZ_RO;

    /* Performance-measurement hook, may be defined as nothing */
    LP0_POST_HOOK;

    /* Enable cache here to reduce boot time */
    ICC_Enable();

    return 0;
}

/* This function can be implemented by the application to initialize the board */
__weak int Board_Init(void)
{
    /* Do nothing */
    return 0;
}

/* This function is called just before control is transferred to main().
 *
 * You may over-ride this function in your program by defining a custom
 *  SystemInit(), but care should be taken to reproduce the initialization
 *  steps or a non-functional system may result.
 */
__weak void SystemInit(void)
{
    /* Configure the interrupt controller to use the application vector table in */
    /* the application space */
#if defined ( __GNUC__)
    /* IAR sets the VTOR pointer prior to SystemInit and causes stack corruption to change it here. */
    __disable_irq(); /* Disable interrupts */
    SCB->VTOR = (uint32_t)__isr_vector; /* set the Vector Table to point at our ISR table */
    __DSB();                        /* bus sync */
    __enable_irq();                 /* enable interrupts */
#endif /* __GNUC__ */
    /* Copy trim information from shadow registers into power manager registers */
    /* NOTE: Checks have been added to prevent bad/missing trim values from being loaded */
    if ((MXC_FLC->ctrl & MXC_F_FLC_CTRL_INFO_BLOCK_VALID) &&
            (MXC_TRIM->for_pwr_reg5 != 0xffffffff) &&
            (MXC_TRIM->for_pwr_reg6 != 0xffffffff)) {
        MXC_PWRSEQ->reg5 = MXC_TRIM->for_pwr_reg5;
        MXC_PWRSEQ->reg6 = MXC_TRIM->for_pwr_reg6;
    } else {
        /* No valid info block, use some reasonable defaults */
        MXC_PWRSEQ->reg6 &= ~MXC_F_PWRSEQ_REG6_PWR_TRIM_OSC_VREF;
        MXC_PWRSEQ->reg6 |= (0x1e0 << MXC_F_PWRSEQ_REG6_PWR_TRIM_OSC_VREF_POS);
    }

    /* Improve flash access timing */
    MXC_FLC->perform |= (MXC_F_FLC_PERFORM_EN_BACK2BACK_RDS |
                         MXC_F_FLC_PERFORM_EN_MERGE_GRAB_GNT |
                         MXC_F_FLC_PERFORM_AUTO_TACC |
                         MXC_F_FLC_PERFORM_AUTO_CLKDIV);

    /* First, eliminate the unnecessary RTC handshake between clock domains. Must be set as a pair. */
    MXC_RTCTMR->ctrl |= (MXC_F_RTC_CTRL_USE_ASYNC_FLAGS |
                         MXC_F_RTC_CTRL_AGGRESSIVE_RST);
    /* Enable fast read of the RTC timer value, and fast write of all other RTC registers */
    MXC_PWRSEQ->rtc_ctrl2 |= (MXC_F_PWRSEQ_RTC_CTRL2_TIMER_AUTO_UPDATE |
                              MXC_F_PWRSEQ_RTC_CTRL2_TIMER_ASYNC_WR);

    MXC_PWRSEQ->rtc_ctrl2 &= ~(MXC_F_PWRSEQ_RTC_CTRL2_TIMER_ASYNC_RD);

    /* Clear the GPIO WUD event if not waking up from LP0 */
    /* this is necessary because WUD flops come up in undetermined state out of POR or SRST*/
    if(MXC_PWRSEQ->reg0 & MXC_F_PWRSEQ_REG0_PWR_FIRST_BOOT ||
       !(MXC_PWRMAN->pwr_rst_ctrl & MXC_F_PWRMAN_PWR_RST_CTRL_POR)) {
        /* Clear GPIO WUD event and configuration registers, globally */
        MXC_PWRSEQ->reg1 |= (MXC_F_PWRSEQ_REG1_PWR_CLR_IO_EVENT_LATCH |
                             MXC_F_PWRSEQ_REG1_PWR_CLR_IO_CFG_LATCH);
        MXC_PWRSEQ->reg1 &= ~(MXC_F_PWRSEQ_REG1_PWR_CLR_IO_EVENT_LATCH |
                              MXC_F_PWRSEQ_REG1_PWR_CLR_IO_CFG_LATCH);
    } else {
        /* Unfreeze the GPIO by clearing MBUS_GATE, when returning from LP0 */
        MXC_PWRSEQ->reg1 &= ~(MXC_F_PWRSEQ_REG1_PWR_MBUS_GATE);
        /* LP0 wake-up: Turn off special switch to eliminate ~50nA of leakage on VDD12 */
        MXC_PWRSEQ->reg1 &= ~MXC_F_PWRSEQ_REG1_PWR_SRAM_NWELL_SW;
    }

    /* Turn on retention regulator */
    MXC_PWRSEQ->reg0 |= (MXC_F_PWRSEQ_REG0_PWR_RETREGEN_RUN |
                         MXC_F_PWRSEQ_REG0_PWR_RETREGEN_SLP);

    /* Adjust settings in the retention controller for fastest wake-up time */
    MXC_PWRSEQ->retn_ctrl0 |= (MXC_F_PWRSEQ_RETN_CTRL0_RC_REL_CCG_EARLY |
                               MXC_F_PWRSEQ_RETN_CTRL0_RC_POLL_FLASH);
    MXC_PWRSEQ->retn_ctrl0 &= ~(MXC_F_PWRSEQ_RETN_CTRL0_RC_USE_FLC_TWK);


    /* Set retention controller TWake cycle count to 1us to minimize the wake-up time */
    /* NOTE: flash polling (...PWRSEQ_RETN_CTRL0_RC_POLL_FLASH) must be enabled before changing POR default! */
    MXC_PWRSEQ->retn_ctrl1 = (MXC_PWRSEQ->retn_ctrl1 & ~MXC_F_PWRSEQ_RETN_CTRL1_RC_TWK) |
                             (1 << MXC_F_PWRSEQ_RETN_CTRL1_RC_TWK_POS);

    /* Improve wake-up time by changing ROSEL to 140ns */
    MXC_PWRSEQ->reg3 = (1 << MXC_F_PWRSEQ_REG3_PWR_ROSEL_POS) |
        (1 << MXC_F_PWRSEQ_REG3_PWR_FAILSEL_POS) |
        (MXC_PWRSEQ->reg3 & ~(MXC_F_PWRSEQ_REG3_PWR_ROSEL |
	       MXC_F_PWRSEQ_REG3_PWR_FLTRROSEL));

    /* Enable RTOS Mode: Enable 32kHz clock synchronizer to SysTick external clock input */
    MXC_CLKMAN->clk_ctrl |= MXC_F_CLKMAN_CLK_CTRL_RTOS_MODE;

    /* Set this so all bits of PWR_MSK_FLAGS are active low to mask the corresponding flags */
    MXC_PWRSEQ->pwr_misc |= MXC_F_PWRSEQ_PWR_MISC_INVERT_4_MASK_BITS;

    /* Enable FPU on Cortex-M4, which occupies coprocessor slots 10 & 11 */
    /* Grant full access, per "Table B3-24 CPACR bit assignments". */
    /* DDI0403D "ARMv7-M Architecture Reference Manual" */
    SCB->CPACR |= SCB_CPACR_CP10_Msk | SCB_CPACR_CP11_Msk;
    __DSB();
    __ISB();

    /* Perform an initial trim of the internal ring oscillator */
    CLKMAN_TrimRO();

#if !defined (__CC_ARM) // Prevent Keil tools from calling these functions until post scatter load
    SystemCoreClockUpdate();
    Board_Init();
#endif /* ! __CC_ARM */

}

#if defined ( __CC_ARM )
/* Function called post memory initialization in the Keil Toolchain, which
 * we are using to call the system core clock upddate and board initialization
 * to prevent data corruption if they are called from SystemInit. */
extern void $Super$$__main_after_scatterload(void);
void $Sub$$__main_after_scatterload(void)
{
    SystemCoreClockUpdate();
    Board_Init();
    $Super$$__main_after_scatterload();
}
#endif /* __CC_ARM */
