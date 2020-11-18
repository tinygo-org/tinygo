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
 * $Date: 2016-03-11 12:50:27 -0600 (Fri, 11 Mar 2016) $
 * $Revision: 21840 $
 *
 ******************************************************************************/

#ifndef _MXC_CLKMAN_REGS_H_
#define _MXC_CLKMAN_REGS_H_

#ifdef __cplusplus
extern "C" {
#endif

#include <stdint.h>

/*
    If types are not defined elsewhere (CMSIS) define them here
*/
#ifndef __IO
#define __IO volatile
#endif
#ifndef __I
#define __I  volatile const
#endif
#ifndef __O
#define __O  volatile
#endif
#ifndef __R
#define __R  volatile const
#endif


/*
   Typedefed structure(s) for module registers (per instance or section) with direct 32-bit
   access to each register in module.
*/

/*                                                          Offset          Register Description
                                                            =============   ============================================================================ */
typedef struct {
    __IO uint32_t clk_config;                           /*  0x0000          System Clock Configuration                                                   */
    __IO uint32_t clk_ctrl;                             /*  0x0004          System Clock Controls                                                        */
    __IO uint32_t intfl;                                /*  0x0008          Interrupt Flags                                                              */
    __IO uint32_t inten;                                /*  0x000C          Interrupt Enable/Disable Controls                                            */
    __IO uint32_t trim_calc;                            /*  0x0010          Trim Calculation Controls                                                    */
    __IO uint32_t i2c_timer_ctrl;                       /*  0x0014          I2C Timer Control                                                            */
    __IO uint32_t cm4_start_clk_en0;                    /*  0x0018          CM4 Start Clock on Interrupt Enable 0                                        */
    __IO uint32_t cm4_start_clk_en1;                    /*  0x001C          CM4 Start Clock on Interrupt Enable 1                                        */
    __IO uint32_t cm4_start_clk_en2;                    /*  0x0020          CM4 Start Clock on Interrupt Enable 2                                        */
    __R  uint32_t rsv024[7];                            /*  0x0024-0x003C                                                                                */
    __IO uint32_t sys_clk_ctrl_0_cm4;                   /*  0x0040          Control Settings for CLK0 - Cortex M4 Clock                                  */
    __IO uint32_t sys_clk_ctrl_1_sync;                  /*  0x0044          Control Settings for CLK1 - Synchronizer Clock                               */
    __IO uint32_t sys_clk_ctrl_2_spix;                  /*  0x0048          Control Settings for CLK2 - SPI XIP Clock                                    */
    __IO uint32_t sys_clk_ctrl_3_prng;                  /*  0x004C          Control Settings for CLK3 - PRNG Clock                                       */
    __IO uint32_t sys_clk_ctrl_4_wdt0;                  /*  0x0050          Control Settings for CLK4 - Watchdog Timer 0                                 */
    __IO uint32_t sys_clk_ctrl_5_wdt1;                  /*  0x0054          Control Settings for CLK5 - Watchdog Timer 1                                 */
    __IO uint32_t sys_clk_ctrl_6_gpio;                  /*  0x0058          Control Settings for CLK6 - Clock for GPIO Ports                             */
    __IO uint32_t sys_clk_ctrl_7_pt;                    /*  0x005C          Control Settings for CLK7 - Source Clock for All Pulse Trains                */
    __IO uint32_t sys_clk_ctrl_8_uart;                  /*  0x0060          Control Settings for CLK8 - Source Clock for All UARTs                       */
    __IO uint32_t sys_clk_ctrl_9_i2cm;                  /*  0x0064          Control Settings for CLK9 - Source Clock for All I2C Masters                 */
    __IO uint32_t sys_clk_ctrl_10_i2cs;                 /*  0x0068          Control Settings for CLK10 - Source Clock for I2C Slave                      */
    __IO uint32_t sys_clk_ctrl_11_spi0;                 /*  0x006C          Control Settings for CLK11 - SPI Master 0                                    */
    __IO uint32_t sys_clk_ctrl_12_spi1;                 /*  0x0070          Control Settings for CLK12 - SPI Master 1                                    */
    __IO uint32_t sys_clk_ctrl_13_spi2;                 /*  0x0074          Control Settings for CLK13 - SPI Master 2                                    */
    __R  uint32_t rsv078;                               /*  0x0078                                                                                       */
    __IO uint32_t sys_clk_ctrl_15_owm;                  /*  0x007C          Control Settings for CLK15 - 1-Wire Master Clock                             */
    __IO uint32_t sys_clk_ctrl_16_spis;                 /*  0x0080          Control Settings for CLK16 - SPI Slave Clock                                 */
    __R  uint32_t rsv084[31];                           /*  0x0084-0x00FC                                                                                */
    __IO uint32_t crypt_clk_ctrl_0_aes;                 /*  0x0100          Control Settings for Crypto Clock 0 - AES                                    */
    __IO uint32_t crypt_clk_ctrl_1_maa;                 /*  0x0104          Control Settings for Crypto Clock 1 - MAA                                    */
    __IO uint32_t crypt_clk_ctrl_2_prng;                /*  0x0108          Control Settings for Crypto Clock 2 - PRNG                                   */
    __R  uint32_t rsv10C[13];                           /*  0x010C-0x013C                                                                                */
    __IO uint32_t clk_gate_ctrl0;                       /*  0x0140          Dynamic Clock Gating Control Register 0                                      */
    __IO uint32_t clk_gate_ctrl1;                       /*  0x0144          Dynamic Clock Gating Control Register 1                                      */
    __IO uint32_t clk_gate_ctrl2;                       /*  0x0148          Dynamic Clock Gating Control Register 2                                      */
} mxc_clkman_regs_t;


/*
   Register offsets for module CLKMAN.
*/

#define MXC_R_CLKMAN_OFFS_CLK_CONFIG                        ((uint32_t)0x00000000UL)
#define MXC_R_CLKMAN_OFFS_CLK_CTRL                          ((uint32_t)0x00000004UL)
#define MXC_R_CLKMAN_OFFS_INTFL                             ((uint32_t)0x00000008UL)
#define MXC_R_CLKMAN_OFFS_INTEN                             ((uint32_t)0x0000000CUL)
#define MXC_R_CLKMAN_OFFS_TRIM_CALC                         ((uint32_t)0x00000010UL)
#define MXC_R_CLKMAN_OFFS_I2C_TIMER_CTRL                    ((uint32_t)0x00000014UL)
#define MXC_R_CLKMAN_OFFS_CM4_START_CLK_EN0                 ((uint32_t)0x00000018UL)
#define MXC_R_CLKMAN_OFFS_CM4_START_CLK_EN1                 ((uint32_t)0x0000001CUL)
#define MXC_R_CLKMAN_OFFS_CM4_START_CLK_EN2                 ((uint32_t)0x00000020UL)
#define MXC_R_CLKMAN_OFFS_SYS_CLK_CTRL_0_CM4                ((uint32_t)0x00000040UL)
#define MXC_R_CLKMAN_OFFS_SYS_CLK_CTRL_1_SYNC               ((uint32_t)0x00000044UL)
#define MXC_R_CLKMAN_OFFS_SYS_CLK_CTRL_2_SPIX               ((uint32_t)0x00000048UL)
#define MXC_R_CLKMAN_OFFS_SYS_CLK_CTRL_3_PRNG               ((uint32_t)0x0000004CUL)
#define MXC_R_CLKMAN_OFFS_SYS_CLK_CTRL_4_WDT0               ((uint32_t)0x00000050UL)
#define MXC_R_CLKMAN_OFFS_SYS_CLK_CTRL_5_WDT1               ((uint32_t)0x00000054UL)
#define MXC_R_CLKMAN_OFFS_SYS_CLK_CTRL_6_GPIO               ((uint32_t)0x00000058UL)
#define MXC_R_CLKMAN_OFFS_SYS_CLK_CTRL_7_PT                 ((uint32_t)0x0000005CUL)
#define MXC_R_CLKMAN_OFFS_SYS_CLK_CTRL_8_UART               ((uint32_t)0x00000060UL)
#define MXC_R_CLKMAN_OFFS_SYS_CLK_CTRL_9_I2CM               ((uint32_t)0x00000064UL)
#define MXC_R_CLKMAN_OFFS_SYS_CLK_CTRL_10_I2CS              ((uint32_t)0x00000068UL)
#define MXC_R_CLKMAN_OFFS_SYS_CLK_CTRL_11_SPI0              ((uint32_t)0x0000006CUL)
#define MXC_R_CLKMAN_OFFS_SYS_CLK_CTRL_12_SPI1              ((uint32_t)0x00000070UL)
#define MXC_R_CLKMAN_OFFS_SYS_CLK_CTRL_13_SPI2              ((uint32_t)0x00000074UL)
#define MXC_R_CLKMAN_OFFS_SYS_CLK_CTRL_15_OWM               ((uint32_t)0x0000007CUL)
#define MXC_R_CLKMAN_OFFS_SYS_CLK_CTRL_16_SPIS              ((uint32_t)0x00000080UL)
#define MXC_R_CLKMAN_OFFS_CRYPT_CLK_CTRL_0_AES              ((uint32_t)0x00000100UL)
#define MXC_R_CLKMAN_OFFS_CRYPT_CLK_CTRL_1_MAA              ((uint32_t)0x00000104UL)
#define MXC_R_CLKMAN_OFFS_CRYPT_CLK_CTRL_2_PRNG             ((uint32_t)0x00000108UL)
#define MXC_R_CLKMAN_OFFS_CLK_GATE_CTRL0                    ((uint32_t)0x00000140UL)
#define MXC_R_CLKMAN_OFFS_CLK_GATE_CTRL1                    ((uint32_t)0x00000144UL)
#define MXC_R_CLKMAN_OFFS_CLK_GATE_CTRL2                    ((uint32_t)0x00000148UL)


/*
   Field positions and masks for module CLKMAN.
*/

#define MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_ENABLE_POS           0
#define MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_ENABLE               ((uint32_t)(0x00000001UL << MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_ENABLE_POS))
#define MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_POS  4
#define MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT      ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_POS))

#define MXC_F_CLKMAN_CLK_CTRL_SYSTEM_SOURCE_SELECT_POS      0
#define MXC_F_CLKMAN_CLK_CTRL_SYSTEM_SOURCE_SELECT          ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_CTRL_SYSTEM_SOURCE_SELECT_POS))
#define MXC_F_CLKMAN_CLK_CTRL_USB_CLOCK_ENABLE_POS          4
#define MXC_F_CLKMAN_CLK_CTRL_USB_CLOCK_ENABLE              ((uint32_t)(0x00000001UL << MXC_F_CLKMAN_CLK_CTRL_USB_CLOCK_ENABLE_POS))
#define MXC_F_CLKMAN_CLK_CTRL_USB_CLOCK_SELECT_POS          5
#define MXC_F_CLKMAN_CLK_CTRL_USB_CLOCK_SELECT              ((uint32_t)(0x00000001UL << MXC_F_CLKMAN_CLK_CTRL_USB_CLOCK_SELECT_POS))
#define MXC_F_CLKMAN_CLK_CTRL_CRYPTO_CLOCK_ENABLE_POS       8
#define MXC_F_CLKMAN_CLK_CTRL_CRYPTO_CLOCK_ENABLE           ((uint32_t)(0x00000001UL << MXC_F_CLKMAN_CLK_CTRL_CRYPTO_CLOCK_ENABLE_POS))
#define MXC_F_CLKMAN_CLK_CTRL_RTOS_MODE_POS                 12
#define MXC_F_CLKMAN_CLK_CTRL_RTOS_MODE                     ((uint32_t)(0x00000001UL << MXC_F_CLKMAN_CLK_CTRL_RTOS_MODE_POS))
#define MXC_F_CLKMAN_CLK_CTRL_CPU_DYNAMIC_CLOCK_POS         13
#define MXC_F_CLKMAN_CLK_CTRL_CPU_DYNAMIC_CLOCK             ((uint32_t)(0x00000001UL << MXC_F_CLKMAN_CLK_CTRL_CPU_DYNAMIC_CLOCK_POS))
#define MXC_F_CLKMAN_CLK_CTRL_WDT0_CLOCK_ENABLE_POS         16
#define MXC_F_CLKMAN_CLK_CTRL_WDT0_CLOCK_ENABLE             ((uint32_t)(0x00000001UL << MXC_F_CLKMAN_CLK_CTRL_WDT0_CLOCK_ENABLE_POS))
#define MXC_F_CLKMAN_CLK_CTRL_WDT0_CLOCK_SELECT_POS         17
#define MXC_F_CLKMAN_CLK_CTRL_WDT0_CLOCK_SELECT             ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_CTRL_WDT0_CLOCK_SELECT_POS))
#define MXC_F_CLKMAN_CLK_CTRL_WDT1_CLOCK_ENABLE_POS         20
#define MXC_F_CLKMAN_CLK_CTRL_WDT1_CLOCK_ENABLE             ((uint32_t)(0x00000001UL << MXC_F_CLKMAN_CLK_CTRL_WDT1_CLOCK_ENABLE_POS))
#define MXC_F_CLKMAN_CLK_CTRL_WDT1_CLOCK_SELECT_POS         21
#define MXC_F_CLKMAN_CLK_CTRL_WDT1_CLOCK_SELECT             ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_CTRL_WDT1_CLOCK_SELECT_POS))
#define MXC_F_CLKMAN_CLK_CTRL_ADC_CLOCK_ENABLE_POS          24
#define MXC_F_CLKMAN_CLK_CTRL_ADC_CLOCK_ENABLE              ((uint32_t)(0x00000001UL << MXC_F_CLKMAN_CLK_CTRL_ADC_CLOCK_ENABLE_POS))

#define MXC_F_CLKMAN_INTFL_CRYPTO_STABLE_POS                0
#define MXC_F_CLKMAN_INTFL_CRYPTO_STABLE                    ((uint32_t)(0x00000001UL << MXC_F_CLKMAN_INTFL_CRYPTO_STABLE_POS))
#define MXC_F_CLKMAN_INTFL_SYS_RO_STABLE_POS                1
#define MXC_F_CLKMAN_INTFL_SYS_RO_STABLE                    ((uint32_t)(0x00000001UL << MXC_F_CLKMAN_INTFL_SYS_RO_STABLE_POS))

#define MXC_F_CLKMAN_INTEN_CRYPTO_STABLE_POS                0
#define MXC_F_CLKMAN_INTEN_CRYPTO_STABLE                    ((uint32_t)(0x00000001UL << MXC_F_CLKMAN_INTEN_CRYPTO_STABLE_POS))
#define MXC_F_CLKMAN_INTEN_SYS_RO_STABLE_POS                1
#define MXC_F_CLKMAN_INTEN_SYS_RO_STABLE                    ((uint32_t)(0x00000001UL << MXC_F_CLKMAN_INTEN_SYS_RO_STABLE_POS))

#define MXC_F_CLKMAN_TRIM_CALC_TRIM_CLK_SEL_POS             0
#define MXC_F_CLKMAN_TRIM_CALC_TRIM_CLK_SEL                 ((uint32_t)(0x00000001UL << MXC_F_CLKMAN_TRIM_CALC_TRIM_CLK_SEL_POS))
#define MXC_F_CLKMAN_TRIM_CALC_TRIM_CALC_START_POS          1
#define MXC_F_CLKMAN_TRIM_CALC_TRIM_CALC_START              ((uint32_t)(0x00000001UL << MXC_F_CLKMAN_TRIM_CALC_TRIM_CALC_START_POS))
#define MXC_F_CLKMAN_TRIM_CALC_TRIM_CALC_COMPLETED_POS      2
#define MXC_F_CLKMAN_TRIM_CALC_TRIM_CALC_COMPLETED          ((uint32_t)(0x00000001UL << MXC_F_CLKMAN_TRIM_CALC_TRIM_CALC_COMPLETED_POS))
#define MXC_F_CLKMAN_TRIM_CALC_TRIM_ENABLE_POS              3
#define MXC_F_CLKMAN_TRIM_CALC_TRIM_ENABLE                  ((uint32_t)(0x00000001UL << MXC_F_CLKMAN_TRIM_CALC_TRIM_ENABLE_POS))
#define MXC_F_CLKMAN_TRIM_CALC_TRIM_CALC_LENGTH_POS         4
#define MXC_F_CLKMAN_TRIM_CALC_TRIM_CALC_LENGTH             ((uint32_t)(0x00000FFFUL << MXC_F_CLKMAN_TRIM_CALC_TRIM_CALC_LENGTH_POS))
#define MXC_F_CLKMAN_TRIM_CALC_TRIM_CALC_RESULTS_POS        16
#define MXC_F_CLKMAN_TRIM_CALC_TRIM_CALC_RESULTS            ((uint32_t)(0x00003FFFUL << MXC_F_CLKMAN_TRIM_CALC_TRIM_CALC_RESULTS_POS))

#define MXC_F_CLKMAN_I2C_TIMER_CTRL_I2C_1MS_TIMER_EN_POS    0
#define MXC_F_CLKMAN_I2C_TIMER_CTRL_I2C_1MS_TIMER_EN        ((uint32_t)(0x00000001UL << MXC_F_CLKMAN_I2C_TIMER_CTRL_I2C_1MS_TIMER_EN_POS))

#define MXC_F_CLKMAN_CM4_START_CLK_EN0_INTS_POS             0
#define MXC_F_CLKMAN_CM4_START_CLK_EN0_INTS                 ((uint32_t)(0xFFFFFFFFUL << MXC_F_CLKMAN_CM4_START_CLK_EN0_INTS_POS))

#define MXC_F_CLKMAN_CM4_START_CLK_EN1_INTS_POS             0
#define MXC_F_CLKMAN_CM4_START_CLK_EN1_INTS                 ((uint32_t)(0xFFFFFFFFUL << MXC_F_CLKMAN_CM4_START_CLK_EN1_INTS_POS))

#define MXC_F_CLKMAN_CM4_START_CLK_EN2_INTS_POS             0
#define MXC_F_CLKMAN_CM4_START_CLK_EN2_INTS                 ((uint32_t)(0xFFFFFFFFUL << MXC_F_CLKMAN_CM4_START_CLK_EN2_INTS_POS))

#define MXC_F_CLKMAN_SYS_CLK_CTRL_0_CM4_CM4_CLK_SCALE_POS   0
#define MXC_F_CLKMAN_SYS_CLK_CTRL_0_CM4_CM4_CLK_SCALE       ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_SYS_CLK_CTRL_0_CM4_CM4_CLK_SCALE_POS))

#define MXC_F_CLKMAN_SYS_CLK_CTRL_1_SYNC_SYNC_CLK_SCALE_POS  0
#define MXC_F_CLKMAN_SYS_CLK_CTRL_1_SYNC_SYNC_CLK_SCALE     ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_SYS_CLK_CTRL_1_SYNC_SYNC_CLK_SCALE_POS))

#define MXC_F_CLKMAN_SYS_CLK_CTRL_2_SPIX_SPIX_CLK_SCALE_POS  0
#define MXC_F_CLKMAN_SYS_CLK_CTRL_2_SPIX_SPIX_CLK_SCALE     ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_SYS_CLK_CTRL_2_SPIX_SPIX_CLK_SCALE_POS))

#define MXC_F_CLKMAN_SYS_CLK_CTRL_3_PRNG_PRNG_CLK_SCALE_POS  0
#define MXC_F_CLKMAN_SYS_CLK_CTRL_3_PRNG_PRNG_CLK_SCALE     ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_SYS_CLK_CTRL_3_PRNG_PRNG_CLK_SCALE_POS))

#define MXC_F_CLKMAN_SYS_CLK_CTRL_4_WDT0_WATCHDOG0_CLK_SCALE_POS  0
#define MXC_F_CLKMAN_SYS_CLK_CTRL_4_WDT0_WATCHDOG0_CLK_SCALE  ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_SYS_CLK_CTRL_4_WDT0_WATCHDOG0_CLK_SCALE_POS))

#define MXC_F_CLKMAN_SYS_CLK_CTRL_5_WDT1_WATCHDOG1_CLK_SCALE_POS  0
#define MXC_F_CLKMAN_SYS_CLK_CTRL_5_WDT1_WATCHDOG1_CLK_SCALE  ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_SYS_CLK_CTRL_5_WDT1_WATCHDOG1_CLK_SCALE_POS))

#define MXC_F_CLKMAN_SYS_CLK_CTRL_6_GPIO_GPIO_CLK_SCALE_POS  0
#define MXC_F_CLKMAN_SYS_CLK_CTRL_6_GPIO_GPIO_CLK_SCALE     ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_SYS_CLK_CTRL_6_GPIO_GPIO_CLK_SCALE_POS))

#define MXC_F_CLKMAN_SYS_CLK_CTRL_7_PT_PULSE_TRAIN_CLK_SCALE_POS  0
#define MXC_F_CLKMAN_SYS_CLK_CTRL_7_PT_PULSE_TRAIN_CLK_SCALE  ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_SYS_CLK_CTRL_7_PT_PULSE_TRAIN_CLK_SCALE_POS))

#define MXC_F_CLKMAN_SYS_CLK_CTRL_8_UART_UART_CLK_SCALE_POS  0
#define MXC_F_CLKMAN_SYS_CLK_CTRL_8_UART_UART_CLK_SCALE     ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_SYS_CLK_CTRL_8_UART_UART_CLK_SCALE_POS))

#define MXC_F_CLKMAN_SYS_CLK_CTRL_9_I2CM_I2CM_CLK_SCALE_POS  0
#define MXC_F_CLKMAN_SYS_CLK_CTRL_9_I2CM_I2CM_CLK_SCALE     ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_SYS_CLK_CTRL_9_I2CM_I2CM_CLK_SCALE_POS))

#define MXC_F_CLKMAN_SYS_CLK_CTRL_10_I2CS_I2CS_CLK_SCALE_POS  0
#define MXC_F_CLKMAN_SYS_CLK_CTRL_10_I2CS_I2CS_CLK_SCALE    ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_SYS_CLK_CTRL_10_I2CS_I2CS_CLK_SCALE_POS))

#define MXC_F_CLKMAN_SYS_CLK_CTRL_11_SPI0_SPI0_CLK_SCALE_POS  0
#define MXC_F_CLKMAN_SYS_CLK_CTRL_11_SPI0_SPI0_CLK_SCALE    ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_SYS_CLK_CTRL_11_SPI0_SPI0_CLK_SCALE_POS))

#define MXC_F_CLKMAN_SYS_CLK_CTRL_12_SPI1_SPI1_CLK_SCALE_POS  0
#define MXC_F_CLKMAN_SYS_CLK_CTRL_12_SPI1_SPI1_CLK_SCALE    ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_SYS_CLK_CTRL_12_SPI1_SPI1_CLK_SCALE_POS))

#define MXC_F_CLKMAN_SYS_CLK_CTRL_13_SPI2_SPI2_CLK_SCALE_POS  0
#define MXC_F_CLKMAN_SYS_CLK_CTRL_13_SPI2_SPI2_CLK_SCALE    ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_SYS_CLK_CTRL_13_SPI2_SPI2_CLK_SCALE_POS))

#define MXC_F_CLKMAN_SYS_CLK_CTRL_15_OWM_OWM_CLK_SCALE_POS  0
#define MXC_F_CLKMAN_SYS_CLK_CTRL_15_OWM_OWM_CLK_SCALE      ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_SYS_CLK_CTRL_15_OWM_OWM_CLK_SCALE_POS))

#define MXC_F_CLKMAN_SYS_CLK_CTRL_16_SPIS_SPIS_CLK_SCALE_POS  0
#define MXC_F_CLKMAN_SYS_CLK_CTRL_16_SPIS_SPIS_CLK_SCALE    ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_SYS_CLK_CTRL_16_SPIS_SPIS_CLK_SCALE_POS))

#define MXC_F_CLKMAN_CRYPT_CLK_CTRL_0_AES_AES_CLK_SCALE_POS  0
#define MXC_F_CLKMAN_CRYPT_CLK_CTRL_0_AES_AES_CLK_SCALE     ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_CRYPT_CLK_CTRL_0_AES_AES_CLK_SCALE_POS))

#define MXC_F_CLKMAN_CRYPT_CLK_CTRL_1_MAA_MAA_CLK_SCALE_POS  0
#define MXC_F_CLKMAN_CRYPT_CLK_CTRL_1_MAA_MAA_CLK_SCALE     ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_CRYPT_CLK_CTRL_1_MAA_MAA_CLK_SCALE_POS))

#define MXC_F_CLKMAN_CRYPT_CLK_CTRL_2_PRNG_PRNG_CLK_SCALE_POS  0
#define MXC_F_CLKMAN_CRYPT_CLK_CTRL_2_PRNG_PRNG_CLK_SCALE   ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_CRYPT_CLK_CTRL_2_PRNG_PRNG_CLK_SCALE_POS))

#define MXC_F_CLKMAN_CLK_GATE_CTRL0_CM4_CLK_GATER_POS       0
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_CM4_CLK_GATER           ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL0_CM4_CLK_GATER_POS))
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_AHB32_CLK_GATER_POS     2
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_AHB32_CLK_GATER         ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL0_AHB32_CLK_GATER_POS))
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_ICACHE_CLK_GATER_POS    4
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_ICACHE_CLK_GATER        ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL0_ICACHE_CLK_GATER_POS))
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_FLASH_CLK_GATER_POS     6
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_FLASH_CLK_GATER         ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL0_FLASH_CLK_GATER_POS))
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_SRAM_CLK_GATER_POS      8
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_SRAM_CLK_GATER          ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL0_SRAM_CLK_GATER_POS))
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_APB_BRIDGE_CLK_GATER_POS  10
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_APB_BRIDGE_CLK_GATER    ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL0_APB_BRIDGE_CLK_GATER_POS))
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_SYSMAN_CLK_GATER_POS    12
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_SYSMAN_CLK_GATER        ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL0_SYSMAN_CLK_GATER_POS))
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_PTP_CLK_GATER_POS       14
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_PTP_CLK_GATER           ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL0_PTP_CLK_GATER_POS))
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_SSB_MUX_CLK_GATER_POS   16
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_SSB_MUX_CLK_GATER       ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL0_SSB_MUX_CLK_GATER_POS))
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_PAD_CLK_GATER_POS       18
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_PAD_CLK_GATER           ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL0_PAD_CLK_GATER_POS))
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_SPIX_CLK_GATER_POS      20
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_SPIX_CLK_GATER          ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL0_SPIX_CLK_GATER_POS))
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_PMU_CLK_GATER_POS       22
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_PMU_CLK_GATER           ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL0_PMU_CLK_GATER_POS))
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_USB_CLK_GATER_POS       24
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_USB_CLK_GATER           ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL0_USB_CLK_GATER_POS))
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_CRC_CLK_GATER_POS       26
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_CRC_CLK_GATER           ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL0_CRC_CLK_GATER_POS))
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_TPU_CLK_GATER_POS       28
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_TPU_CLK_GATER           ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL0_TPU_CLK_GATER_POS))
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_WATCHDOG0_CLK_GATER_POS  30
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_WATCHDOG0_CLK_GATER     ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL0_WATCHDOG0_CLK_GATER_POS))

#define MXC_F_CLKMAN_CLK_GATE_CTRL1_WATCHDOG1_CLK_GATER_POS  0
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_WATCHDOG1_CLK_GATER     ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL1_WATCHDOG1_CLK_GATER_POS))
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_GPIO_CLK_GATER_POS      2
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_GPIO_CLK_GATER          ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL1_GPIO_CLK_GATER_POS))
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_TIMER0_CLK_GATER_POS    4
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_TIMER0_CLK_GATER        ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL1_TIMER0_CLK_GATER_POS))
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_TIMER1_CLK_GATER_POS    6
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_TIMER1_CLK_GATER        ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL1_TIMER1_CLK_GATER_POS))
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_TIMER2_CLK_GATER_POS    8
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_TIMER2_CLK_GATER        ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL1_TIMER2_CLK_GATER_POS))
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_TIMER3_CLK_GATER_POS    10
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_TIMER3_CLK_GATER        ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL1_TIMER3_CLK_GATER_POS))
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_TIMER4_CLK_GATER_POS    12
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_TIMER4_CLK_GATER        ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL1_TIMER4_CLK_GATER_POS))
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_TIMER5_CLK_GATER_POS    14
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_TIMER5_CLK_GATER        ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL1_TIMER5_CLK_GATER_POS))
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_PULSETRAIN_CLK_GATER_POS  16
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_PULSETRAIN_CLK_GATER    ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL1_PULSETRAIN_CLK_GATER_POS))
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_UART0_CLK_GATER_POS     18
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_UART0_CLK_GATER         ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL1_UART0_CLK_GATER_POS))
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_UART1_CLK_GATER_POS     20
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_UART1_CLK_GATER         ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL1_UART1_CLK_GATER_POS))
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_UART2_CLK_GATER_POS     22
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_UART2_CLK_GATER         ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL1_UART2_CLK_GATER_POS))
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_I2CM0_CLK_GATER_POS     26
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_I2CM0_CLK_GATER         ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL1_I2CM0_CLK_GATER_POS))
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_I2CM1_CLK_GATER_POS     28
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_I2CM1_CLK_GATER         ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL1_I2CM1_CLK_GATER_POS))

#define MXC_F_CLKMAN_CLK_GATE_CTRL2_I2CS_CLK_GATER_POS      0
#define MXC_F_CLKMAN_CLK_GATE_CTRL2_I2CS_CLK_GATER          ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL2_I2CS_CLK_GATER_POS))
#define MXC_F_CLKMAN_CLK_GATE_CTRL2_SPI0_CLK_GATER_POS      2
#define MXC_F_CLKMAN_CLK_GATE_CTRL2_SPI0_CLK_GATER          ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL2_SPI0_CLK_GATER_POS))
#define MXC_F_CLKMAN_CLK_GATE_CTRL2_SPI1_CLK_GATER_POS      4
#define MXC_F_CLKMAN_CLK_GATE_CTRL2_SPI1_CLK_GATER          ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL2_SPI1_CLK_GATER_POS))
#define MXC_F_CLKMAN_CLK_GATE_CTRL2_SPI2_CLK_GATER_POS      6
#define MXC_F_CLKMAN_CLK_GATE_CTRL2_SPI2_CLK_GATER          ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL2_SPI2_CLK_GATER_POS))
#define MXC_F_CLKMAN_CLK_GATE_CTRL2_OWM_CLK_GATER_POS       10
#define MXC_F_CLKMAN_CLK_GATE_CTRL2_OWM_CLK_GATER           ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL2_OWM_CLK_GATER_POS))
#define MXC_F_CLKMAN_CLK_GATE_CTRL2_ADC_CLK_GATER_POS       12
#define MXC_F_CLKMAN_CLK_GATE_CTRL2_ADC_CLK_GATER           ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL2_ADC_CLK_GATER_POS))
#define MXC_F_CLKMAN_CLK_GATE_CTRL2_SPIS_CLK_GATER_POS      14
#define MXC_F_CLKMAN_CLK_GATE_CTRL2_SPIS_CLK_GATER          ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL2_SPIS_CLK_GATER_POS))



/*
   Field values and shifted values for module CLKMAN.
*/

#define MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_8_CLOCKS           ((uint32_t)(0x00000000UL))
#define MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_9_CLOCKS           ((uint32_t)(0x00000001UL))
#define MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_10_CLOCKS          ((uint32_t)(0x00000002UL))
#define MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_11_CLOCKS          ((uint32_t)(0x00000003UL))
#define MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_12_CLOCKS          ((uint32_t)(0x00000004UL))
#define MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_13_CLOCKS          ((uint32_t)(0x00000005UL))
#define MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_14_CLOCKS          ((uint32_t)(0x00000006UL))
#define MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_15_CLOCKS          ((uint32_t)(0x00000007UL))
#define MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_16_CLOCKS          ((uint32_t)(0x00000008UL))
#define MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_17_CLOCKS          ((uint32_t)(0x00000009UL))
#define MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_18_CLOCKS          ((uint32_t)(0x0000000AUL))
#define MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_19_CLOCKS          ((uint32_t)(0x0000000BUL))
#define MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_20_CLOCKS          ((uint32_t)(0x0000000CUL))
#define MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_21_CLOCKS          ((uint32_t)(0x0000000DUL))
#define MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_22_CLOCKS          ((uint32_t)(0x0000000EUL))
#define MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_23_CLOCKS          ((uint32_t)(0x0000000FUL))

#define MXC_S_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_8_CLOCKS           ((uint32_t)(MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_8_CLOCKS    << MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_POS))
#define MXC_S_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_9_CLOCKS           ((uint32_t)(MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_9_CLOCKS    << MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_POS))
#define MXC_S_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_10_CLOCKS          ((uint32_t)(MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_10_CLOCKS   << MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_POS))
#define MXC_S_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_11_CLOCKS          ((uint32_t)(MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_11_CLOCKS   << MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_POS))
#define MXC_S_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_12_CLOCKS          ((uint32_t)(MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_12_CLOCKS   << MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_POS))
#define MXC_S_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_13_CLOCKS          ((uint32_t)(MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_13_CLOCKS   << MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_POS))
#define MXC_S_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_14_CLOCKS          ((uint32_t)(MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_14_CLOCKS   << MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_POS))
#define MXC_S_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_15_CLOCKS          ((uint32_t)(MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_15_CLOCKS   << MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_POS))
#define MXC_S_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_16_CLOCKS          ((uint32_t)(MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_16_CLOCKS   << MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_POS))
#define MXC_S_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_17_CLOCKS          ((uint32_t)(MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_17_CLOCKS   << MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_POS))
#define MXC_S_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_18_CLOCKS          ((uint32_t)(MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_18_CLOCKS   << MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_POS))
#define MXC_S_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_19_CLOCKS          ((uint32_t)(MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_19_CLOCKS   << MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_POS))
#define MXC_S_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_20_CLOCKS          ((uint32_t)(MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_20_CLOCKS   << MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_POS))
#define MXC_S_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_21_CLOCKS          ((uint32_t)(MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_21_CLOCKS   << MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_POS))
#define MXC_S_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_22_CLOCKS          ((uint32_t)(MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_22_CLOCKS   << MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_POS))
#define MXC_S_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_23_CLOCKS          ((uint32_t)(MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_23_CLOCKS   << MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_POS))

#define MXC_V_CLKMAN_CLK_CTRL_SYSTEM_SOURCE_SELECT_96MHZ_RO_DIV_2               ((uint32_t)(0x00000000UL))
#define MXC_V_CLKMAN_CLK_CTRL_SYSTEM_SOURCE_SELECT_96MHZ_RO                     ((uint32_t)(0x00000001UL))

#define MXC_S_CLKMAN_CLK_CTRL_SYSTEM_SOURCE_SELECT_96MHZ_RO_DIV_2               ((uint32_t)(MXC_V_CLKMAN_CLK_CTRL_SYSTEM_SOURCE_SELECT_96MHZ_RO_DIV_2  << MXC_F_CLKMAN_CLK_CTRL_SYSTEM_SOURCE_SELECT_POS))
#define MXC_S_CLKMAN_CLK_CTRL_SYSTEM_SOURCE_SELECT_96MHZ_RO                     ((uint32_t)(MXC_V_CLKMAN_CLK_CTRL_SYSTEM_SOURCE_SELECT_96MHZ_RO        << MXC_F_CLKMAN_CLK_CTRL_SYSTEM_SOURCE_SELECT_POS))

#define MXC_V_CLKMAN_WDT0_CLOCK_SELECT_SCALED_SYS_CLK_CTRL_4_WDT0               ((uint32_t)(0x00000000UL))
#define MXC_V_CLKMAN_WDT0_CLOCK_SELECT_32KHZ_RTC_OSCILLATOR                     ((uint32_t)(0x00000001UL))
#define MXC_V_CLKMAN_WDT0_CLOCK_SELECT_96MHZ_OSCILLATOR                         ((uint32_t)(0x00000002UL))
#define MXC_V_CLKMAN_WDT0_CLOCK_SELECT_NANO_RING_OSCILLATOR                     ((uint32_t)(0x00000003UL))

#define MXC_S_CLKMAN_WDT0_CLOCK_SELECT_SCALED_SYS_CLK_CTRL_4_WDT0               ((uint32_t)(MXC_V_CLKMAN_WDT0_CLOCK_SELECT_SCALED_SYS_CLK_CTRL_4_WDT0  << MXC_F_CLKMAN_CLK_CTRL_WDT0_CLOCK_SELECT_POS))
#define MXC_S_CLKMAN_WDT0_CLOCK_SELECT_32KHZ_RTC_OSCILLATOR                     ((uint32_t)(MXC_V_CLKMAN_WDT0_CLOCK_SELECT_32KHZ_RTC_OSCILLATOR        << MXC_F_CLKMAN_CLK_CTRL_WDT0_CLOCK_SELECT_POS))
#define MXC_S_CLKMAN_WDT0_CLOCK_SELECT_96MHZ_OSCILLATOR                         ((uint32_t)(MXC_V_CLKMAN_WDT0_CLOCK_SELECT_96MHZ_OSCILLATOR            << MXC_F_CLKMAN_CLK_CTRL_WDT0_CLOCK_SELECT_POS))
#define MXC_S_CLKMAN_WDT0_CLOCK_SELECT_NANO_RING_OSCILLATOR                     ((uint32_t)(MXC_V_CLKMAN_WDT0_CLOCK_SELECT_NANO_RING_OSCILLATOR        << MXC_F_CLKMAN_CLK_CTRL_WDT0_CLOCK_SELECT_POS))

#define MXC_V_CLKMAN_WDT1_CLOCK_SELECT_SCALED_SYS_CLK_CTRL_4_WDT1               ((uint32_t)(0x00000000UL))
#define MXC_V_CLKMAN_WDT1_CLOCK_SELECT_32KHZ_RTC_OSCILLATOR                     ((uint32_t)(0x00000001UL))
#define MXC_V_CLKMAN_WDT1_CLOCK_SELECT_96MHZ_OSCILLATOR                         ((uint32_t)(0x00000002UL))
#define MXC_V_CLKMAN_WDT1_CLOCK_SELECT_NANO_RING_OSCILLATOR                     ((uint32_t)(0x00000003UL))

#define MXC_S_CLKMAN_WDT1_CLOCK_SELECT_SCALED_SYS_CLK_CTRL_4_WDT1               ((uint32_t)(MXC_V_CLKMAN_WDT1_CLOCK_SELECT_SCALED_SYS_CLK_CTRL_4_WDT1  << MXC_F_CLKMAN_CLK_CTRL_WDT1_CLOCK_SELECT_POS))
#define MXC_S_CLKMAN_WDT1_CLOCK_SELECT_32KHZ_RTC_OSCILLATOR                     ((uint32_t)(MXC_V_CLKMAN_WDT1_CLOCK_SELECT_32KHZ_RTC_OSCILLATOR        << MXC_F_CLKMAN_CLK_CTRL_WDT1_CLOCK_SELECT_POS))
#define MXC_S_CLKMAN_WDT1_CLOCK_SELECT_96MHZ_OSCILLATOR                         ((uint32_t)(MXC_V_CLKMAN_WDT1_CLOCK_SELECT_96MHZ_OSCILLATOR            << MXC_F_CLKMAN_CLK_CTRL_WDT1_CLOCK_SELECT_POS))
#define MXC_S_CLKMAN_WDT1_CLOCK_SELECT_NANO_RING_OSCILLATOR                     ((uint32_t)(MXC_V_CLKMAN_WDT1_CLOCK_SELECT_NANO_RING_OSCILLATOR        << MXC_F_CLKMAN_CLK_CTRL_WDT1_CLOCK_SELECT_POS))

#define MXC_V_CLKMAN_CLK_SCALE_DISABLED                                         ((uint32_t)(0x00000000UL))
#define MXC_V_CLKMAN_CLK_SCALE_DIV_1                                            ((uint32_t)(0x00000001UL))
#define MXC_V_CLKMAN_CLK_SCALE_DIV_2                                            ((uint32_t)(0x00000002UL))
#define MXC_V_CLKMAN_CLK_SCALE_DIV_4                                            ((uint32_t)(0x00000003UL))
#define MXC_V_CLKMAN_CLK_SCALE_DIV_8                                            ((uint32_t)(0x00000004UL))
#define MXC_V_CLKMAN_CLK_SCALE_DIV_16                                           ((uint32_t)(0x00000005UL))
#define MXC_V_CLKMAN_CLK_SCALE_DIV_32                                           ((uint32_t)(0x00000006UL))
#define MXC_V_CLKMAN_CLK_SCALE_DIV_64                                           ((uint32_t)(0x00000007UL))
#define MXC_V_CLKMAN_CLK_SCALE_DIV_128                                          ((uint32_t)(0x00000008UL))
#define MXC_V_CLKMAN_CLK_SCALE_DIV_256                                          ((uint32_t)(0x00000009UL))

#define MXC_S_CLKMAN_CLK_SCALE_DISABLED                                         ((uint32_t)(MXC_V_CLKMAN_CLK_SCALE_DISABLED  << MXC_F_CLKMAN_SYS_CLK_CTRL_0_CM4_CM4_CLK_SCALE_POS))
#define MXC_S_CLKMAN_CLK_SCALE_DIV_1                                            ((uint32_t)(MXC_V_CLKMAN_CLK_SCALE_DIV_1     << MXC_F_CLKMAN_SYS_CLK_CTRL_0_CM4_CM4_CLK_SCALE_POS))
#define MXC_S_CLKMAN_CLK_SCALE_DIV_2                                            ((uint32_t)(MXC_V_CLKMAN_CLK_SCALE_DIV_2     << MXC_F_CLKMAN_SYS_CLK_CTRL_0_CM4_CM4_CLK_SCALE_POS))
#define MXC_S_CLKMAN_CLK_SCALE_DIV_4                                            ((uint32_t)(MXC_V_CLKMAN_CLK_SCALE_DIV_4     << MXC_F_CLKMAN_SYS_CLK_CTRL_0_CM4_CM4_CLK_SCALE_POS))
#define MXC_S_CLKMAN_CLK_SCALE_DIV_8                                            ((uint32_t)(MXC_V_CLKMAN_CLK_SCALE_DIV_8     << MXC_F_CLKMAN_SYS_CLK_CTRL_0_CM4_CM4_CLK_SCALE_POS))
#define MXC_S_CLKMAN_CLK_SCALE_DIV_16                                           ((uint32_t)(MXC_V_CLKMAN_CLK_SCALE_DIV_16    << MXC_F_CLKMAN_SYS_CLK_CTRL_0_CM4_CM4_CLK_SCALE_POS))
#define MXC_S_CLKMAN_CLK_SCALE_DIV_32                                           ((uint32_t)(MXC_V_CLKMAN_CLK_SCALE_DIV_32    << MXC_F_CLKMAN_SYS_CLK_CTRL_0_CM4_CM4_CLK_SCALE_POS))
#define MXC_S_CLKMAN_CLK_SCALE_DIV_64                                           ((uint32_t)(MXC_V_CLKMAN_CLK_SCALE_DIV_64    << MXC_F_CLKMAN_SYS_CLK_CTRL_0_CM4_CM4_CLK_SCALE_POS))
#define MXC_S_CLKMAN_CLK_SCALE_DIV_128                                          ((uint32_t)(MXC_V_CLKMAN_CLK_SCALE_DIV_128   << MXC_F_CLKMAN_SYS_CLK_CTRL_0_CM4_CM4_CLK_SCALE_POS))
#define MXC_S_CLKMAN_CLK_SCALE_DIV_256                                          ((uint32_t)(MXC_V_CLKMAN_CLK_SCALE_DIV_256   << MXC_F_CLKMAN_SYS_CLK_CTRL_0_CM4_CM4_CLK_SCALE_POS))



#ifdef __cplusplus
}
#endif

#endif   /* _MXC_CLKMAN_REGS_H_ */

