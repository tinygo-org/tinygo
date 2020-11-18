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
 * $Date: 2016-06-17 13:07:24 -0500 (Fri, 17 Jun 2016) $
 * $Revision: 23369 $
 *
 ******************************************************************************/

#include <stddef.h>
#include "mxc_config.h"
#include "mxc_assert.h"
#include "mxc_sys.h"
#include "ioman.h"
#include "clkman.h"
#include "pwrseq_regs.h"
#include "pwrman_regs.h"
#include "spix_regs.h"
#include "trim_regs.h"

/***** Definitions *****/
#define SYS_RTC_CLK 32768UL

/******************************************************************************/
uint32_t SYS_GetFreq(uint32_t clk_scale)
{
    uint32_t freq;
    unsigned int clkdiv;

    if (clk_scale == MXC_V_CLKMAN_CLK_SCALE_DISABLED) {
        freq = 0;
    } else {
        clkdiv = 1 << (clk_scale - 1);
        freq = SystemCoreClock / clkdiv;
    }

    return freq;
}

/******************************************************************************/
uint32_t SYS_CPU_GetFreq(void)
{
    return SYS_GetFreq(CLKMAN_GetClkScale(CLKMAN_CLK_CPU));
}

/******************************************************************************/
int SYS_ADC_Init(void)
{
  /* Power up the ADC AFE, enable clocks */
  MXC_PWRMAN->pwr_rst_ctrl |= MXC_F_PWRMAN_PWR_RST_CTRL_AFE_POWERED;
  MXC_CLKMAN->clk_ctrl |= MXC_F_CLKMAN_CLK_CTRL_ADC_CLOCK_ENABLE;

  return E_NO_ERROR;
} 

/******************************************************************************/
int SYS_AES_Init(void)
{
    /* Set up clocks for AES block */
    /* Enable crypto ring oscillator, which is used by all TPU components (AES, uMAA, etc.) */
    CLKMAN_CryptoClockEnable(1);

    /* Change prescaler to /1 */
    CLKMAN_SetClkScale(CLKMAN_CRYPTO_CLK_AES, CLKMAN_SCALE_DIV_1);

    return E_NO_ERROR;
}

/******************************************************************************/
int SYS_GPIO_Init(void)
{
    if (CLKMAN_GetClkScale(CLKMAN_CLK_GPIO) == CLKMAN_SCALE_DISABLED) {
        CLKMAN_SetClkScale(CLKMAN_CLK_GPIO, CLKMAN_SCALE_DIV_1);
    }

    return E_NO_ERROR;
}

/******************************************************************************/
int SYS_UART_Init(mxc_uart_regs_t *uart, const uart_cfg_t *uart_cfg, const sys_cfg_uart_t *sys_cfg)
{
    static int subsequent_call = 0;
    int err, idx;
    clkman_scale_t clk_scale;
    uint32_t min_baud;

    if(sys_cfg == NULL)
        return E_NULL_PTR;

    if (sys_cfg->clk_scale != CLKMAN_SCALE_AUTO) {
        CLKMAN_SetClkScale(CLKMAN_CLK_UART, sys_cfg->clk_scale);
    } else if (!subsequent_call) {
        /* This clock divider is shared amongst all UARTs. Only change it if it
         * hasn't already been configured. UART_Init() will check for validity
         * for this baudrate.
         */
        subsequent_call = 1;

        /* Setup the clock divider for the given baud rate */
        clk_scale = CLKMAN_SCALE_DISABLED;
        do {
            min_baud = ((SystemCoreClock >> clk_scale++) / (16 * (MXC_F_UART_BAUD_BAUD_DIVISOR >> MXC_F_UART_BAUD_BAUD_DIVISOR_POS)));
        } while (uart_cfg->baud < min_baud && clk_scale < CLKMAN_SCALE_AUTO);

        /* check if baud rate cannot be reached */
        if(uart_cfg->baud < min_baud)
            return E_BAD_STATE;

        CLKMAN_SetClkScale(CLKMAN_CLK_UART, clk_scale);
    }

    if ((err = IOMAN_Config(&sys_cfg->io_cfg)) != E_NO_ERROR) {
        return err;
    }

    /* Reset the peripheral */
    idx = MXC_UART_GET_IDX(uart);
    MXC_PWRMAN->peripheral_reset |= (MXC_F_PWRMAN_PERIPHERAL_RESET_UART0 << idx);
    MXC_PWRMAN->peripheral_reset &= ~((MXC_F_PWRMAN_PERIPHERAL_RESET_UART0 << idx));

    return E_NO_ERROR;
}

/******************************************************************************/
int SYS_UART_Shutdown(mxc_uart_regs_t *uart)
{
    int err;
    int idx = MXC_UART_GET_IDX(uart);
    ioman_cfg_t io_cfg = (ioman_cfg_t)IOMAN_UART(idx, 0, 0, 0, 0, 0, 0);

    if ((err = IOMAN_Config(&io_cfg)) != E_NO_ERROR) {
        return err;
    }

    return E_NO_ERROR;
}

/******************************************************************************/
uint32_t SYS_UART_GetFreq(mxc_uart_regs_t *uart)
{
    return SYS_GetFreq(CLKMAN_GetClkScale(CLKMAN_CLK_UART));
}

/******************************************************************************/
int SYS_I2CM_Init(mxc_i2cm_regs_t *i2cm, const sys_cfg_i2cm_t *cfg)
{
    int err;

    if(cfg == NULL)
        return E_NULL_PTR;

    CLKMAN_SetClkScale(CLKMAN_CLK_I2CM, cfg->clk_scale);
    MXC_CLKMAN->i2c_timer_ctrl = 1;

    if ((err = IOMAN_Config(&cfg->io_cfg)) != E_NO_ERROR) {
        return err;
    }

    return E_NO_ERROR;
}

/******************************************************************************/
int SYS_I2CM_Shutdown(mxc_i2cm_regs_t *i2cm)
{
    int err;
    int idx = MXC_I2CM_GET_IDX(i2cm);
    ioman_cfg_t io_cfg;

    switch(idx)
    {
        case 0:
            io_cfg = (ioman_cfg_t)IOMAN_I2CM0(0,0);
            break;
        case 1:
            io_cfg = (ioman_cfg_t)IOMAN_I2CM1(0,0);
            break;
        case 2:
            io_cfg = (ioman_cfg_t)IOMAN_I2CM2(0,0);
            break;
        default:
            return E_BAD_PARAM;
    }

    if ((err = IOMAN_Config(&io_cfg)) != E_NO_ERROR) {
        return err;
    }

    return E_NO_ERROR;
}

/******************************************************************************/
uint32_t SYS_I2CM_GetFreq(mxc_i2cm_regs_t *i2cm)
{
    return SYS_GetFreq(CLKMAN_GetClkScale(CLKMAN_CLK_I2CM));
}

/******************************************************************************/
int SYS_I2CS_Init(mxc_i2cs_regs_t *i2cs, const sys_cfg_i2cs_t *cfg)
{
    int err;

    if(cfg == NULL)
        return E_NULL_PTR;

    CLKMAN_SetClkScale(CLKMAN_CLK_I2CS, cfg->clk_scale);
    MXC_CLKMAN->i2c_timer_ctrl = 1;

    if ((err = IOMAN_Config(&cfg->io_cfg)) != E_NO_ERROR) {
        return err;
    }

    return E_NO_ERROR;
}

/******************************************************************************/
int SYS_I2CS_Shutdown(mxc_i2cs_regs_t *i2cs)
{
    int err;
    ioman_cfg_t io_cfg = (ioman_cfg_t)IOMAN_I2CS(0, 0);

    if ((err = IOMAN_Config(&io_cfg)) != E_NO_ERROR) {
        return err;
    }

    return E_NO_ERROR;
}

/******************************************************************************/
uint32_t SYS_I2CS_GetFreq(mxc_i2cs_regs_t *i2cs)
{
    uint32_t freq, clkdiv;

    if (CLKMAN_GetClkScale(CLKMAN_CLK_I2CS) == MXC_V_CLKMAN_CLK_SCALE_DISABLED) {
        freq = 0;
    } else {
        clkdiv = 1 << (CLKMAN_GetClkScale(CLKMAN_CLK_I2CS) - 1);
        freq = (SystemCoreClock / clkdiv);
    }

    return freq;
}

/******************************************************************************/
int SYS_SPIM_Init(mxc_spim_regs_t *spim, const spim_cfg_t *spim_cfg, const sys_cfg_spim_t *sys_cfg)
{
    int err, idx;
    clkman_scale_t clk_scale;
    uint32_t max_baud;

    if(sys_cfg == NULL)
        return E_NULL_PTR;

    idx = MXC_SPIM_GET_IDX(spim);

    if (sys_cfg->clk_scale != CLKMAN_SCALE_AUTO) {
        if(spim_cfg->baud > ((SystemCoreClock >> (sys_cfg->clk_scale - 1))/2)) {
            return E_BAD_PARAM;
        }
        CLKMAN_SetClkScale((clkman_clk_t)(CLKMAN_CLK_SPIM0 + idx), sys_cfg->clk_scale);
    } else {

        if(spim_cfg->baud > (SystemCoreClock/2)) {
            return E_BAD_PARAM;
        }

        /* Setup the clock divider for the given baud rate */
        clk_scale = CLKMAN_SCALE_DISABLED;
        do {
            max_baud = ((SystemCoreClock >> clk_scale++) / 2);
        } while (spim_cfg->baud < max_baud && clk_scale < CLKMAN_SCALE_AUTO);

        if(clk_scale == CLKMAN_SCALE_AUTO) {
            clk_scale--;
        }
        
        CLKMAN_SetClkScale((clkman_clk_t)(CLKMAN_CLK_SPIM0 + idx), clk_scale);
    }

    if ((err = IOMAN_Config(&sys_cfg->io_cfg)) != E_NO_ERROR) {
        return err;
    }

    return E_NO_ERROR;
}

/******************************************************************************/
int SYS_SPIM_Shutdown(mxc_spim_regs_t *spim)
{
    int err;
    int idx = MXC_SPIM_GET_IDX(spim);
    ioman_cfg_t io_cfg;

    switch(idx)
    {
        case 0:
            io_cfg = (ioman_cfg_t)IOMAN_SPIM0(0, 0, 0, 0, 0, 0, 0, 0);
            break;
        case 1:
            io_cfg = (ioman_cfg_t)IOMAN_SPIM1(0, 0, 0, 0, 0, 0);
            break;
        case 2:
            io_cfg = (ioman_cfg_t)IOMAN_SPIM2(0, 0, 0, 0, 0, 0, 0, 0, 0);
            break;
        default:
            return E_BAD_PARAM;
    }

    if ((err = IOMAN_Config(&io_cfg)) != E_NO_ERROR) {
        return err;
    }

    return E_NO_ERROR;
}

/******************************************************************************/
uint32_t SYS_SPIM_GetFreq(mxc_spim_regs_t *spim)
{
    int idx = MXC_SPIM_GET_IDX(spim);
    return SYS_GetFreq(CLKMAN_GetClkScale((clkman_clk_t)(CLKMAN_CLK_SPIM0 + idx)));
}

/******************************************************************************/
int SYS_SPIX_Init(const sys_cfg_spix_t *sys_cfg, uint32_t baud)
{
    int err;
    clkman_scale_t clk_scale;
    uint32_t min_baud;

    if (sys_cfg->clk_scale != CLKMAN_SCALE_AUTO) {
         CLKMAN_SetClkScale((clkman_clk_t)(CLKMAN_CLK_SPIX), sys_cfg->clk_scale);
    } else {
        /* Setup the clock divider for the given baud rate */
        clk_scale = CLKMAN_SCALE_DISABLED;
        do {
            min_baud = ((SystemCoreClock >> clk_scale++) / (2 * 
                (MXC_F_SPIX_MASTER_CFG_SCK_HI_CLK >> MXC_F_SPIX_MASTER_CFG_SCK_HI_CLK_POS)));
        } while (baud < min_baud && clk_scale < CLKMAN_SCALE_AUTO);

        /* check if baud rate cannot be reached */
        if(baud < min_baud)
            return E_BAD_STATE;

        CLKMAN_SetClkScale((clkman_clk_t)(CLKMAN_CLK_SPIX), clk_scale);
    }

    if ((err = IOMAN_Config(&sys_cfg->io_cfg)) != E_NO_ERROR) {
        return err;
    }

    return E_NO_ERROR;
}

/******************************************************************************/
int SYS_SPIX_Shutdown()
{
    int err;
    ioman_cfg_t io_cfg = IOMAN_SPIX(0, 0, 0, 0, 0, 0);

    if ((err = IOMAN_Config(&io_cfg)) != E_NO_ERROR) {
        return err;
    }

    return E_NO_ERROR;
}

/******************************************************************************/
uint32_t SYS_SPIX_GetFreq()
{
    return SYS_GetFreq(CLKMAN_GetClkScale((clkman_clk_t)(CLKMAN_CLK_SPIX)));
}

/******************************************************************************/
int SYS_SPIS_Init(const sys_cfg_spis_t *sys_cfg)
{
    int err;

    if (sys_cfg->clk_scale != CLKMAN_SCALE_AUTO) {
         CLKMAN_SetClkScale((clkman_clk_t)(CLKMAN_CLK_SPIS), sys_cfg->clk_scale);
    } else {
        CLKMAN_SetClkScale((clkman_clk_t)(CLKMAN_CLK_SPIS), CLKMAN_SCALE_DIV_1);
    }

    if ((err = IOMAN_Config(&sys_cfg->io_cfg)) != E_NO_ERROR) {
        return err;
    }

    return E_NO_ERROR;
}

/******************************************************************************/
int SYS_SPIS_Shutdown()
{
    int err;
    ioman_cfg_t io_cfg = IOMAN_SPIS(0, 0, 0);

    if ((err = IOMAN_Config(&io_cfg)) != E_NO_ERROR) {
        return err;
    }

    return E_NO_ERROR;
}

/******************************************************************************/
uint32_t SYS_SPIS_GetFreq()
{
    return SYS_GetFreq(CLKMAN_GetClkScale((clkman_clk_t)(CLKMAN_CLK_SPIS)));
}

/******************************************************************************/
int SYS_OWM_Init(mxc_owm_regs_t *owm, const sys_cfg_owm_t *sys_cfg)
{
    int err;

    if(sys_cfg == NULL)
        return E_NULL_PTR;

    if (sys_cfg->clk_scale != CLKMAN_SCALE_AUTO)
    {
        CLKMAN_SetClkScale(CLKMAN_CLK_OWM, sys_cfg->clk_scale);
    }
    else
    {
        CLKMAN_SetClkScale(CLKMAN_CLK_OWM, CLKMAN_SCALE_DIV_1);
    }

    if ((err = IOMAN_Config(&sys_cfg->io_cfg)) != E_NO_ERROR) {
        return err;
    }

    return E_NO_ERROR;
}

/******************************************************************************/
int SYS_OWM_Shutdown(mxc_owm_regs_t *owm)
{
    int err;

    ioman_cfg_t io_cfg = IOMAN_OWM(0, 0);

    if ((err = IOMAN_Config(&io_cfg)) != E_NO_ERROR) {
        return err;
    }

    return E_NO_ERROR;
}

/******************************************************************************/
uint32_t SYS_OWM_GetFreq(mxc_owm_regs_t *owm)
{
    return SYS_GetFreq(CLKMAN_GetClkScale(CLKMAN_CLK_OWM));
}

/******************************************************************************/
uint32_t SYS_TMR_GetFreq(mxc_tmr_regs_t *tmr)
{
    return SystemCoreClock;
}

/******************************************************************************/
int SYS_TMR_Init(mxc_tmr_regs_t *tmr, const sys_cfg_tmr_t *cfg)
{
    int pin, gpio_index, tmr_index;

    if (cfg != NULL)
    {
        /* Make sure the given GPIO mapps to the given TMR */
        for (pin = 0; pin < MXC_GPIO_MAX_PINS_PER_PORT; pin++)
        {
            if(cfg->mask & (1 << pin))
            {
                gpio_index = (MXC_GPIO_MAX_PINS_PER_PORT * cfg->port) + pin;
                tmr_index = gpio_index % MXC_CFG_TMR_INSTANCES;

                if(tmr_index == MXC_TMR_GET_IDX(tmr))
                    return GPIO_Config(cfg);
                else
                    return E_BAD_PARAM;
            }
        }

        return E_BAD_PARAM;

    } else {
        return E_NO_ERROR;
    }
}

/******************************************************************************/
uint32_t SYS_SysTick_GetFreq(void)
{
    /* Determine is using internal (SystemCoreClock) or external (32768) clock */
    if ( (SysTick->CTRL & SysTick_CTRL_CLKSOURCE_Msk) || !(SysTick->CTRL & SysTick_CTRL_ENABLE_Msk)) {
        return SystemCoreClock;
    } else {
        return SYS_RTC_CLK;
    }
}

/******************************************************************************/
uint32_t SYS_PT_GetFreq(void)
{
    return SYS_GetFreq(CLKMAN_GetClkScale(CLKMAN_CLK_PT));
}

/******************************************************************************/
void SYS_PT_Init(sys_pt_clk_scale clk_scale)
{
    /* setup clock divider for pulse train clock */
    CLKMAN_SetClkScale(CLKMAN_CLK_PT, clk_scale);
}

/******************************************************************************/
int SYS_PT_Config(mxc_pt_regs_t *pt, const sys_cfg_pt_t *cfg)
{
    int pt_index;

    /* Make sure the given GPIO mapps to the given PT */
    pt_index = MXC_PT_GET_IDX(pt);
    if(pt_index < 0) {
        return E_NOT_SUPPORTED;
    }

    /* Even number port */
    if(cfg->port%2 == 0) {
        /* Pin number should match PT number */
        if(!(cfg->mask & (0x1 << pt_index))) {
            return E_NOT_SUPPORTED;
        }
    } else {
        /* Pin number+8 should match PT */
        if(!((cfg->mask << 8) & (0x1 << pt_index))) {
            return E_NOT_SUPPORTED;
        }
    }

    return GPIO_Config(cfg);
}

/******************************************************************************/
void SYS_USB_Enable(uint8_t enable)
{
    /* Enable USB clock */
    CLKMAN_ClockGate(CLKMAN_USB_CLOCK, enable);

    if(enable) {
        /* Enable USB Power */
        MXC_PWRMAN->pwr_rst_ctrl |= MXC_F_PWRMAN_PWR_RST_CTRL_USB_POWERED;
    } else {
        /* Disable USB Power */
        MXC_PWRMAN->pwr_rst_ctrl &= ~MXC_F_PWRMAN_PWR_RST_CTRL_USB_POWERED;
    }
}

/******************************************************************************/
int SYS_SysTick_Config(uint32_t ticks, int clk_src)
{

    if(ticks == 0)
        return E_BAD_PARAM;

    /* If SystemClock, call default CMSIS config and return */
    if (clk_src) {
        return SysTick_Config(ticks);
    } else { /* External clock source requested
                enable RTC clock in run mode*/
        MXC_PWRSEQ->reg0 |= (MXC_F_PWRSEQ_REG0_PWR_RTCEN_RUN);

        /* Disable SysTick Timer */
        SysTick->CTRL = 0;
        /* Check reload value for valid */
        if ((ticks - 1) > SysTick_LOAD_RELOAD_Msk) {
            /* Reload value impossible */
            return E_BAD_PARAM;
        }
        /* set reload register */
        SysTick->LOAD  = ticks - 1;

        /* set Priority for Systick Interrupt */
        NVIC_SetPriority(SysTick_IRQn, (1<<__NVIC_PRIO_BITS) - 1);

        /* Load the SysTick Counter Value */
        SysTick->VAL   = 0;

        /* Enable SysTick IRQ and SysTick Timer leaving clock source as external */
        SysTick->CTRL  = SysTick_CTRL_TICKINT_Msk | SysTick_CTRL_ENABLE_Msk;

        /* Function successful */
        return E_NO_ERROR;
    }
}

/******************************************************************************/
int SYS_SysTick_Delay(uint32_t ticks)
{
    uint32_t cur_ticks, num_full, num_remain, previous_ticks, num_subtract, i;
    uint32_t reload, value, ctrl; /* save/restore variables */

    if(ticks == 0)
        return E_BAD_PARAM;

    /* If SysTick is not enabled we can take it for our delay */
    if (!(SysTick->CTRL & SysTick_CTRL_ENABLE_Msk)) {

        /* Save current state in case it's disabled but already configured, restore at return.*/
        reload = SysTick->LOAD;
        value = SysTick->VAL;
        ctrl = SysTick->CTRL;

        /* get the number of ticks less than max RELOAD. */
        num_remain = ticks % SysTick_LOAD_RELOAD_Msk;

        /* if ticks is < Max SysTick Reload num_full will be 0, otherwise it will
           give us the number of max SysTicks cycles required */
        num_full = (ticks - 1) / SysTick_LOAD_RELOAD_Msk;

        /* Do the required full systick countdowns */
        if (num_full) {
            /* load the max count value into systick */
            SysTick->LOAD = SysTick_LOAD_RELOAD_Msk;
            /* load the starting value */
            SysTick->VAL = 0;
            /*enable SysTick counter with SystemClock source internal, immediately forces LOAD register into VAL register */
            SysTick->CTRL = SysTick_CTRL_CLKSOURCE_Msk | SysTick_CTRL_ENABLE_Msk;
            /* CountFlag will get set when VAL reaches zero */
            for (i = num_full; i > 0; i--) {
                do {
                    cur_ticks = SysTick->CTRL;
                } while (!(cur_ticks & SysTick_CTRL_COUNTFLAG_Msk));
            }
            /* Disable systick */
            SysTick->CTRL = 0;
        }
        /* Now handle the remainder of ticks */
        if (num_remain) {
            SysTick->LOAD = num_remain;
            SysTick->VAL = 0;
            SysTick->CTRL = SysTick_CTRL_CLKSOURCE_Msk | SysTick_CTRL_ENABLE_Msk;
            /* wait for countflag to get set */
            do {
                cur_ticks = SysTick->CTRL;
            } while (!(cur_ticks & SysTick_CTRL_COUNTFLAG_Msk));
            /* Disable systick */
            SysTick->CTRL = 0;
        }

        /* restore original state of SysTick and return */
        SysTick->LOAD = reload;
        SysTick->VAL =  value;
        SysTick->CTRL = ctrl;

        return E_NO_ERROR;

    } else { /* SysTick is enabled
           When SysTick is enabled count flag can not be used
           and the reload can not be changed.
           Do not read the CTRL register -> clears count flag */

        /* Get the reload value for wrap/reload case */
        reload = SysTick->LOAD;

        /* Read the starting systick value */
        previous_ticks = SysTick->VAL;

        do {
            /* get current SysTick value */
            cur_ticks = SysTick->VAL;
            /* Check for wrap/reload of timer countval */
            if (cur_ticks > previous_ticks) {
                /* subtract count to 0 (previous_ticks) and wrap (reload value - cur_ticks) */
                num_subtract = (previous_ticks + (reload - cur_ticks));
            } else { /* standard case (no wrap)
                        subtract off the number of ticks since last pass */
                num_subtract = (previous_ticks - cur_ticks);
            }
            /* check to see if we are done. */
            if (num_subtract >= ticks)
                return E_NO_ERROR;
            else
                ticks -= num_subtract;
            /* cur_ticks becomes previous_ticks for next timer read. */
            previous_ticks = cur_ticks;
        } while (ticks > 0);
        /* Should not ever be reached */
        return E_NO_ERROR;
    }
}

/******************************************************************************/
int SYS_RTC_Init(void)
{
    /* Enable power for RTC for all LPx states */
    MXC_PWRSEQ->reg0 |= (MXC_F_PWRSEQ_REG0_PWR_RTCEN_RUN |
                         MXC_F_PWRSEQ_REG0_PWR_RTCEN_SLP);

    /* Enable clock to synchronizers */
    CLKMAN_SetClkScale(CLKMAN_CLK_SYNC, CLKMAN_SCALE_DIV_1);

    return E_NO_ERROR;
}

/******************************************************************************/
void SYS_IOMAN_UseVDDIO(const gpio_cfg_t *cfg)
{
    unsigned int startbit = (cfg->port * 8);
    volatile uint32_t *use_vddioh_reg = &MXC_IOMAN->use_vddioh_0 + (startbit / 32);
    *use_vddioh_reg &= ~cfg->mask << (startbit % 32);
}

/******************************************************************************/
void SYS_IOMAN_UseVDDIOH(const gpio_cfg_t *cfg)
{
    unsigned int startbit = (cfg->port * 8);
    volatile uint32_t *use_vddioh_reg = &MXC_IOMAN->use_vddioh_0 + (startbit / 32);
    *use_vddioh_reg |= cfg->mask << (startbit % 32);
}

/******************************************************************************/
void SYS_WDT_Init(mxc_wdt_regs_t *wdt, const sys_cfg_wdt_t *cfg)
{

    if(cfg->clk == CLKMAN_WDT_SELECT_NANO_RING_OSCILLATOR)
    {
        /*enable nanoring in run mode */
        MXC_PWRSEQ->reg0 |= (MXC_F_PWRSEQ_REG0_PWR_NREN_RUN);
    }
    else if(cfg->clk == CLKMAN_WDT_SELECT_32KHZ_RTC_OSCILLATOR)
    {
        /*enabled RTC in run mode */
        MXC_PWRSEQ->reg0 |= (MXC_F_PWRSEQ_REG0_PWR_RTCEN_RUN);
    }

    if(wdt == MXC_WDT0) {
        /*select clock source */
        CLKMAN_WdtClkSelect(0, cfg->clk);

        /*Set scale of clock (only used for system clock as source) */
        CLKMAN_SetClkScale(CLKMAN_CLK_WDT0, cfg->clk_scale);

        /*Enable clock */
        CLKMAN_ClockGate(CLKMAN_WDT0_CLOCK, 1);
    } else if (wdt == MXC_WDT1) {
        /*select clock source */
        CLKMAN_WdtClkSelect(1, cfg->clk);

        /*Set scale of clock (only used for system clock as source) */
        CLKMAN_SetClkScale(CLKMAN_CLK_WDT1, cfg->clk_scale);

        /*Enable clock */
        CLKMAN_ClockGate(CLKMAN_WDT1_CLOCK, 1);
    }
}

/******************************************************************************/
void SYS_PRNG_Init(void)
{
    /* Start crypto ring, unconditionally */
    CLKMAN_CryptoClockEnable(1);

    /* If we find the dividers in anything other than off, don't touch them */
    if (CLKMAN_GetClkScale(CLKMAN_CRYPTO_CLK_PRNG) == CLKMAN_SCALE_DISABLED) {
        /* Div 1 mode */
        CLKMAN_SetClkScale(CLKMAN_CRYPTO_CLK_PRNG, CLKMAN_SCALE_DIV_1);
    }

    if (CLKMAN_GetClkScale(CLKMAN_CLK_PRNG) == CLKMAN_SCALE_DISABLED) {
        /* Div 1 mode */
        CLKMAN_SetClkScale(CLKMAN_CLK_PRNG, CLKMAN_SCALE_DIV_1);
    }
}

/******************************************************************************/
void SYS_MAA_Init(void)
{
    /* Start crypto ring, unconditionally */
    CLKMAN_CryptoClockEnable(1);

    /* If we find the dividers in anything other than off, don't touch them */
    if (CLKMAN_GetClkScale(CLKMAN_CRYPTO_CLK_MAA) == CLKMAN_SCALE_DISABLED) {
        /* Div 1 mode */
        CLKMAN_SetClkScale(CLKMAN_CRYPTO_CLK_MAA, CLKMAN_SCALE_DIV_1);
    }
}

/******************************************************************************/
uint32_t SYS_SRAM_GetSize(void)
{
    uint32_t memSize;

    /* Read TRIM value*/
    int SRAMtrim = (MXC_TRIM->reg10_mem_size & MXC_F_TRIM_REG10_MEM_SIZE_SRAM) >> MXC_F_TRIM_REG10_MEM_SIZE_SRAM_POS;

    /* Decode trim value into memory size in bytes */
    switch(SRAMtrim)
    {
        case MXC_V_TRIM_REG10_MEM_SRAM_THREE_FOURTHS_SIZE:
            memSize = (MXC_SRAM_FULL_MEM_SIZE >> 2) * 3;
            break;

        case MXC_V_TRIM_REG10_MEM_SRAM_HALF_SIZE:
            memSize = MXC_SRAM_FULL_MEM_SIZE >> 1;
            break;

        default: /* other values are FULL size */
            memSize = MXC_SRAM_FULL_MEM_SIZE;
            break;
    }

    /* Returns size in bytes */
    return memSize;
}

/******************************************************************************/
uint32_t SYS_FLASH_GetSize(void)
{
    uint32_t memSize;

    /* Read TRIM value */
    int FLASHtrim = (MXC_TRIM->reg10_mem_size & MXC_F_TRIM_REG10_MEM_SIZE_FLASH) >> MXC_F_TRIM_REG10_MEM_SIZE_FLASH_POS;

    /* Decode trim value into memory size in bytes*/
    switch(FLASHtrim)
    {
        case MXC_V_TRIM_REG10_MEM_FLASH_THREE_FOURTHS_SIZE:
            memSize = (MXC_FLASH_FULL_MEM_SIZE >> 2) * 3;
            break;
        case MXC_V_TRIM_REG10_MEM_FLASH_HALF_SIZE:
            memSize = (MXC_FLASH_FULL_MEM_SIZE >> 1);
            break;
        case MXC_V_TRIM_REG10_MEM_FLASH_THREE_EIGHTHS_SIZE:
            memSize = (MXC_FLASH_FULL_MEM_SIZE >> 3) * 3;
            break;
        case MXC_V_TRIM_REG10_MEM_FLASH_FOURTH_SIZE:
            memSize = (MXC_FLASH_FULL_MEM_SIZE >> 2);
            break;
        default: /* other values are FULL size */
            memSize = MXC_FLASH_FULL_MEM_SIZE;
            break;
    }

    /* Returns size in bytes */
    return memSize;
}
