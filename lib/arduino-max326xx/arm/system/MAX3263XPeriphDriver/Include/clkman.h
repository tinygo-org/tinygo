/**
 * @file
 * @brief      Registers, Bit Masks and Bit Positions for the System Clock
 *             Management (CLKMAN) module.
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
 * $Date: 2016-08-15 11:08:12 -0500 (Mon, 15 Aug 2016) $
 * $Revision: 24058 $
 *
 **************************************************************************** */

/* Define to prevent redundant inclusion */
#ifndef _CLKMAN_H_
#define _CLKMAN_H_

/* **** Includes **** */
#include "mxc_config.h"
#include "clkman_regs.h"

#ifdef __cplusplus
extern "C" {
#endif

/**
 * @ingroup syscfg
 * @defgroup clkman Clock Management
 * @brief System and Peripheral Clock Management API
 * @{
 */

/* **** Definitions **** */

/**
 * Enumeration type specifying the System Clock Rate. @see CLKMAN_SYSTEM_SOURCE_values
 */
typedef enum {
    CLKMAN_SYSTEM_SOURCE_96MHZ  = 0,                    /**< Clock select for 96MHz oscillator.*/
    CLKMAN_SYSTEM_SOURCE_4MHZ   = 1                     /**< Clock select for 4MHz oscillator. */
}
clkman_system_source_select_t;

/**
 * Enumeration type for setting the system clock divider.
 * @note 4MHz System source can only be divided down by a maximum factor of 8.
 */
typedef enum {
    CLKMAN_SYSTEM_SCALE_DIV_1   = 0,                    /**< Clock scale for dividing system by 1.  */
    CLKMAN_SYSTEM_SCALE_DIV_2   = 1,                    /**< Clock scale for dividing system by 2.  */
    CLKMAN_SYSTEM_SCALE_DIV_4   = 2,                    /**< Clock scale for dividing system by 4.  */
    CLKMAN_SYSTEM_SCALE_DIV_8   = 3,                    /**< Clock scale for dividing system by 8.  */
    CLKMAN_SYSTEM_SCALE_DIV_16  = 4                     /**< Clock scale for dividing system by 16. */
} clkman_system_scale_t;

/**
 * Enumeration type for selecting a peripheral module for setting and getting it's clock scale.
 */
typedef enum {
    CLKMAN_CLK_CPU              = 0,                    /**< CPU clock.                                     */    
    CLKMAN_CLK_SYNC             = 1,                    /**< Synchronizer clock.                            */
    CLKMAN_CLK_SPIX             = 2,                    /**< SPI XIP module clock.                          */
    CLKMAN_CLK_PRNG             = 3,                    /**< PRNG module clock.                             */
    CLKMAN_CLK_WDT0             = 4,                    /**< Watchdog Timer 0 clock.                        */
    CLKMAN_CLK_WDT1             = 5,                    /**< Watchdog Timer 1 clock.                        */
    CLKMAN_CLK_GPIO             = 6,                    /**< GPIO module clock.                             */
    CLKMAN_CLK_PT               = 7,                    /**< Pulse Train engine clock.                      */
    CLKMAN_CLK_UART             = 8,                    /**< UART clock.                                    */
    CLKMAN_CLK_I2CM             = 9,                    /**< I2C Master module clock (for all instances).   */
    CLKMAN_CLK_I2CS             = 10,                   /**< I2C Slave module clock.                        */
    CLKMAN_CLK_SPIM0            = 11,                   /**< SPI Master instance 0 module clock.            */
    CLKMAN_CLK_SPIM1            = 12,                   /**< SPI Master instance 1 module clock.            */
    CLKMAN_CLK_SPIM2            = 13,                   /**< SPI Master instance 2 module clock.            */
    CLKMAN_CLK_SPIB             = 14,                   /**< SPI Bridge module clock.                       */
    CLKMAN_CLK_OWM              = 15,                   /**< OWM module clock.                              */
    CLKMAN_CLK_SPIS             = 16,                   /**< SPI Slave module clock.                        */
    CLKMAN_CRYPTO_CLK_AES       = 17,                   /**< AES engine clock.                              */
    CLKMAN_CRYPTO_CLK_MAA       = 18,                   /**< Modular Arithmetic Accelerator (MAA) clock.    */
    CLKMAN_CRYPTO_CLK_PRNG      = 19,                   /**< Pseudo-random number Generator (PRNG) clock.   */
    CLKMAN_CLK_MAX                                      /**< Maximum value of enum for limit checking.      */
} clkman_clk_t;

/**
 * Enumeration type for selecting a peripheral module (USB, Cryto, ADC, WDT0, WDT1 and RTC/RTOS)
 * to enable/disable clock gating. 
 */
typedef enum {
    CLKMAN_USB_CLOCK    = MXC_F_CLKMAN_CLK_CTRL_USB_CLOCK_ENABLE,       /**< Enable/Disable mask for USB.               */
    CLKMAN_CRYPTO_CLOCK = MXC_F_CLKMAN_CLK_CTRL_CRYPTO_CLOCK_ENABLE,    /**< Enable/Disable mask for Crypto Clock.      */
    CLKMAN_ADC_CLOCK    = MXC_F_CLKMAN_CLK_CTRL_ADC_CLOCK_ENABLE,       /**< Enable/Disable mask for ADC.               */
    CLKMAN_WDT0_CLOCK   = MXC_F_CLKMAN_CLK_CTRL_WDT0_CLOCK_ENABLE,      /**< Enable/Disable mask for Watch Dog Timer 0. */
    CLKMAN_WDT1_CLOCK   = MXC_F_CLKMAN_CLK_CTRL_WDT1_CLOCK_ENABLE,      /**< Enable/Disable mask for Watch Dog Timer 1. */
    CLKMAN_RTOS_MODE    = MXC_F_CLKMAN_CLK_CTRL_RTOS_MODE               /**< Enable/Disable mask for 32kHz clock in LP1 required to use JTAG for debug. */
} clkman_enable_clk_t;

/**
 * Enumeration type for selecting the clock scale for the system or peripheral module.
 */
typedef enum {
    CLKMAN_SCALE_DISABLED = MXC_V_CLKMAN_CLK_SCALE_DISABLED,            /**< Clock disabled.                        */
    CLKMAN_SCALE_DIV_1    = MXC_V_CLKMAN_CLK_SCALE_DIV_1,               /**< Clock scale for dividing by 1.         */
    CLKMAN_SCALE_DIV_2    = MXC_V_CLKMAN_CLK_SCALE_DIV_2,               /**< Clock scale for dividing by 2.         */
    CLKMAN_SCALE_DIV_4    = MXC_V_CLKMAN_CLK_SCALE_DIV_4,               /**< Clock scale for dividing by 4.         */
    CLKMAN_SCALE_DIV_8    = MXC_V_CLKMAN_CLK_SCALE_DIV_8,               /**< Clock scale for dividing by 8.         */
    CLKMAN_SCALE_DIV_16   = MXC_V_CLKMAN_CLK_SCALE_DIV_16,              /**< Clock scale for dividing by 16.        */
    CLKMAN_SCALE_DIV_32   = MXC_V_CLKMAN_CLK_SCALE_DIV_32,              /**< Clock scale for dividing by 32.        */
    CLKMAN_SCALE_DIV_64   = MXC_V_CLKMAN_CLK_SCALE_DIV_64,              /**< Clock scale for dividing by 64.        */
    CLKMAN_SCALE_DIV_128  = MXC_V_CLKMAN_CLK_SCALE_DIV_128,             /**< Clock scale for dividing by 128.       */
    CLKMAN_SCALE_DIV_256  = MXC_V_CLKMAN_CLK_SCALE_DIV_256,             /**< Clock scale for dividing by 256.       */
    CLKMAN_SCALE_AUTO                                                   /**< Clock scale to auto select divider.    */
} clkman_scale_t;

/*
 * Enumeration type for selecting the source clock for the Watch Dog Timers.
 * | Enumeration Selection                    | Value | WDT Clock Source            |
 * | :--------------------------------------: | :---: | :-------------------------- |
 * | CLKMAN_WDT_SELECT_SCALED_SYS_CLK_CTRL    | 0     | Scaled System Clock         |
 * | CLKMAN_WDT_SELECT_32KHZ_RTC_OSCILLATOR   | 1     | 32 kHz Real-Time Clock      |
 * | CLKMAN_WDT_SELECT_96MHZ_OSCILLATOR       | 2     | 96 MHz Oscillator unscaled  |
 * | CLKMAN_WDT_SELECT_NANO_RING_OSCILLATOR   | 3     | Nano-ring clock             |
 * | CLKMAN_WDT_SELECT_DISABLED               | 4     | WDT0 Clock is disabled      |
 */
typedef enum {
    CLKMAN_WDT_SELECT_SCALED_SYS_CLK_CTRL  = MXC_V_CLKMAN_WDT0_CLOCK_SELECT_SCALED_SYS_CLK_CTRL_4_WDT0,     /**< Use scaled system clock for Watchdog Timer 0.              */
    CLKMAN_WDT_SELECT_32KHZ_RTC_OSCILLATOR = MXC_V_CLKMAN_WDT0_CLOCK_SELECT_32KHZ_RTC_OSCILLATOR,           /**< Use 32kHz oscillator for Watchdog Timer 0.                 */
    CLKMAN_WDT_SELECT_96MHZ_OSCILLATOR     = MXC_V_CLKMAN_WDT0_CLOCK_SELECT_96MHZ_OSCILLATOR,               /**< Use 96MHz clock for Watchdog Timer 0.                      */
    CLKMAN_WDT_SELECT_NANO_RING_OSCILLATOR = MXC_V_CLKMAN_WDT0_CLOCK_SELECT_NANO_RING_OSCILLATOR,           /**< Use Nano-Ring Oscillator (8kHz) for Watchdog Timer 0 clock.*/
    CLKMAN_WDT_SELECT_DISABLED                                                                              /**< Watchdog Timer 0 clock disabled.                           */
} clkman_wdt_clk_select_t;


/* **** Function Prototypes **** */

/**
 * @brief      Selects the system clock source,
 * @note       4MHz System source can only be divided down by a maximum factor
 *             of 8.
 *
 * @param      select  System clock source.
 * @param      scale   System clock scaler.
 */
void CLKMAN_SetSystemClock(clkman_system_source_select_t select, clkman_system_scale_t scale);

/**
 * @brief      Enables/disables the Crypto/TPU relaxation oscillator
 *
 * @param      enable  |:------- | :---: |
 *                     | Enable  |   1   |
 *                     | Disable |   0   |
 */
void CLKMAN_CryptoClockEnable(int enable);

/**
 * @brief      Enables/Disables clock gating for the specified peripheral
 *             module.
 *
 * @param      clk     Peripheral module to enable/disable clock gating.
 * @param      enable  Enable (1) or Disable (0).
 */
void CLKMAN_ClockGate(clkman_enable_clk_t clk, int enable);

/**
 * @brief      Sets the specified clock scaler value.
 *
 * @param      clk    Peripheral module to set the desired clock scale.
 * @param      scale  Clock scale/divisor for the specified peripheral module.
 */
void CLKMAN_SetClkScale(clkman_clk_t clk, clkman_scale_t scale);

/**
 * @brief      Get the clock scaler/divisor value for the specified peripheral
 *             module.
 *
 * @param      clk   The peripheral module to get the current clock scale setting, see #clkman_clk_t.
 * @return     A value indicating the clock divisor/scale of the requested
 *             peripheral module.
 */
clkman_scale_t CLKMAN_GetClkScale(clkman_clk_t clk);

/**
 * @brief      Selects the clock source for the specified watchdog timer.
 *
 * @param      idx     Value indicating the WDT to set the clock source on.
 * @param      select  Value of the desired clock source for the WDT.
 */
int CLKMAN_WdtClkSelect(unsigned int idx, clkman_wdt_clk_select_t select);

/**
 * @brief      Get the interrupt flags for the CLKMAN module.
 *
 * @return     The current interrupt flags.
 */
__STATIC_INLINE uint32_t CLKMAN_GetFlags(void)
{
    return MXC_CLKMAN->intfl;
}

/**
 * @brief      Clear the specified interrupt flags
 *
 * @param      mask  mask of clock management interrupt flags to clear
 */
__STATIC_INLINE void CLKMAN_ClrFlags(uint32_t mask)
{
    MXC_CLKMAN->intfl = mask;
}

/**
 * @brief      Enable the interrupts specified in the mask parameter.
 *
 * @param      mask  Mask of clock management interrupts to enable, 1 to enable
 *                   a specific interrupt.
 */
__STATIC_INLINE void CLKMAN_EnableInt(uint32_t mask)
{
    MXC_CLKMAN->inten |= mask;
}

/**
 * @brief      Disable the specified interrupts
 *
 * @param      mask  Mask of CLKMAN interrupts to disable, 1 to disable a
 *                   specific interrupt.
 */
__STATIC_INLINE void CLKMAN_DisableInt(uint32_t mask)
{
    MXC_CLKMAN->inten &= ~mask;
}

/**
 * @brief      Trim the ring oscillator.
 * @note       CLKMAN_TrimRO() is implemented in system_max32XXX.c
 */
void CLKMAN_TrimRO(void);

/**@} end of group clkman */

#ifdef __cplusplus
}
#endif

#endif /* _CLKMAN_H_ */
