/**
 * @file
 * @brief   Type definitions for the Clock Management Interface
 *
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
#ifndef _MXC_CLKMAN_REGS_H_
#define _MXC_CLKMAN_REGS_H_

/* **** Includes **** */
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

/// @cond
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
/// @endcond

/**
 * @ingroup     clkman
 * @defgroup    clkman_registers Registers
 * @brief       Registers, Bit Masks and Bit Positions
 * @{
 */

/**
 * Structure type for the Clock Management module registers allowing direct 32-bit access to each register.
 */
typedef struct {
    __IO uint32_t clk_config;                           /**< <b><tt>0x0000:       </tt></b> CLKMAN_CLK_CONFIG Register - System Clock Configuration                   */
    __IO uint32_t clk_ctrl;                             /**< <b><tt>0x0004:       </tt></b> CLKMAN_CLK_CTRL Register - System Clock Controls                          */
    __IO uint32_t intfl;                                /**< <b><tt>0x0008:       </tt></b> CLKMAN_INTFL Register - Interrupt Flags                                   */
    __IO uint32_t inten;                                /**< <b><tt>0x000C:       </tt></b> CLKMAN_INTEN Register - Interrupt Enable/Disable Controls                 */
    __IO uint32_t trim_calc;                            /**< <b><tt>0x0010:       </tt></b> CLKMAN_TRIM_CALC Register - Trim Calculation Controls                     */
    __IO uint32_t i2c_timer_ctrl;                       /**< <b><tt>0x0014:       </tt></b> CLKMAN_I2C_TIMER_CTRL Register - I2C Timer Control                        */
    __IO uint32_t cm4_start_clk_en0;                    /**< <b><tt>0x0018:       </tt></b> CLKMAN_CM4_START_CLK_EN0 Register - CM4 Start Clock on Interrupt Enable 0 */
    __IO uint32_t cm4_start_clk_en1;                    /**< <b><tt>0x001C:       </tt></b> CLKMAN_CM4_START_CLK_EN1 Register - CM4 Start Clock on Interrupt Enable 1 */
    __IO uint32_t cm4_start_clk_en2;                    /**< <b><tt>0x0020:       </tt></b> CLKMAN_CM4_START_CLK_EN2 Register - CM4 Start Clock on Interrupt Enable 2 */
    __R  uint32_t rsv024[7];                            /**< <b><tt>0x0024-0x003C:</tt></b> RESERVED                                                                  */
    __IO uint32_t sys_clk_ctrl_0_cm4;                   /**< <b><tt>0x0040:       </tt></b> CLKMAN_SYS_CLK_CTRL_0_CM4 Register - Cortex M4 Clock                      */
    __IO uint32_t sys_clk_ctrl_1_sync;                  /**< <b><tt>0x0044:       </tt></b> CLKMAN_SYS_CLK_CTRL_1_SYNC Register - Synchronizer Clock                  */
    __IO uint32_t sys_clk_ctrl_2_spix;                  /**< <b><tt>0x0048:       </tt></b> CLKMAN_SYS_CLK_CTRL_2_SPIX Register - SPI XIP Clock                       */
    __IO uint32_t sys_clk_ctrl_3_prng;                  /**< <b><tt>0x004C:       </tt></b> CLKMAN_SYS_CLK_CTRL_3_PRNG Register - PRNG Clock                          */
    __IO uint32_t sys_clk_ctrl_4_wdt0;                  /**< <b><tt>0x0050:       </tt></b> CLKMAN_SYS_CLK_CTRL_4_WDT0 Register -  Watchdog Timer 0                   */
    __IO uint32_t sys_clk_ctrl_5_wdt1;                  /**< <b><tt>0x0054:       </tt></b> CLKMAN_SYS_CLK_CTRL_5_WDT1 Register - Watchdog Timer 1                    */
    __IO uint32_t sys_clk_ctrl_6_gpio;                  /**< <b><tt>0x0058:       </tt></b> CLKMAN_SYS_CLK_CTRL_6_GPIO Register - Clock for GPIO Ports                */
    __IO uint32_t sys_clk_ctrl_7_pt;                    /**< <b><tt>0x005C:       </tt></b> CLKMAN_SYS_CLK_CTRL_7_PT Register - Source Clock for All Pulse Trains     */
    __IO uint32_t sys_clk_ctrl_8_uart;                  /**< <b><tt>0x0060:       </tt></b> CLKMAN_SYS_CLK_CTRL_8_UART Register - Source Clock for All UARTs          */
    __IO uint32_t sys_clk_ctrl_9_i2cm;                  /**< <b><tt>0x0064:       </tt></b> CLKMAN_SYS_CLK_CTRL_9_I2CM Register  - Source Clock for All I2C Masters   */
    __IO uint32_t sys_clk_ctrl_10_i2cs;                 /**< <b><tt>0x0068:       </tt></b> CLKMAN_SYS_CLK_CTRL_10_I2CS Register - Source Clock for I2C Slave         */
    __IO uint32_t sys_clk_ctrl_11_spi0;                 /**< <b><tt>0x006C:       </tt></b> CLKMAN_SYS_CLK_CTRL_11_SPI0 Register - SPI Master 0                       */
    __IO uint32_t sys_clk_ctrl_12_spi1;                 /**< <b><tt>0x0070:       </tt></b> CLKMAN_SYS_CLK_CTRL_12_SPI1 Register - SPI Master 1                       */
    __IO uint32_t sys_clk_ctrl_13_spi2;                 /**< <b><tt>0x0074:       </tt></b> CLKMAN_SYS_CLK_CTRL_13_SPI2 Register - SPI Master 2                       */
    __IO uint32_t sys_clk_ctrl_14_spib;                 /**< <b><tt>0x0078:       </tt></b> CLKMAN_SYS_CLK_CTRL_14_SPIB Register - SPI Bridge Clock                   */
    __IO uint32_t sys_clk_ctrl_15_owm;                  /**< <b><tt>0x007C:       </tt></b> CLKMAN_SYS_CLK_CTRL_15_OWM Register  - 1-Wire Master Clock                */
    __IO uint32_t sys_clk_ctrl_16_spis;                 /**< <b><tt>0x0080:       </tt></b> CLKMAN_SYS_CLK_CTRL_16_SPIS Register - SPI Slave Clock                    */
    __R  uint32_t rsv084[31];                           /**< <b><tt>0x0084-0x00FC:</tt></b> RESERVED:                                                                 */
    __IO uint32_t crypt_clk_ctrl_0_aes;                 /**< <b><tt>0x0100:       </tt></b> CLKMAN_CRYPT_CLK_CTRL_0_AES Register - AES                                */
    __IO uint32_t crypt_clk_ctrl_1_maa;                 /**< <b><tt>0x0104:       </tt></b> CLKMAN_CRYPT_CLK_CTRL_1_MAA Register - MAA                                */
    __IO uint32_t crypt_clk_ctrl_2_prng;                /**< <b><tt>0x0108:       </tt></b> CLKMAN_CRYPT_CLK_CTRL_2_PRNG Register - PRNG                              */
    __R  uint32_t rsv10C[13];                           /**< <b><tt>0x010C-0x013C:</tt></b> RESERVED                                                                  */
    __IO uint32_t clk_gate_ctrl0;                       /**< <b><tt>0x0140:       </tt></b> CLKMAN_CLK_GATE_CTRL0 Register - Dynamic Clock Gating Control Register 0  */
    __IO uint32_t clk_gate_ctrl1;                       /**< <b><tt>0x0144:       </tt></b> CLKMAN_CLK_GATE_CTRL1 Register - Dynamic Clock Gating Control Register 1  */
    __IO uint32_t clk_gate_ctrl2;                       /**< <b><tt>0x0148:       </tt></b> CLKMAN_CLK_GATE_CTRL2 Register - Dynamic Clock Gating Control Register 2  */
} mxc_clkman_regs_t;
/**@} end of clkman_registers */   

/*
   Register offsets for module CLKMAN.
*/
/**
 * @ingroup    clkman_registers
 * @defgroup   CLKMAN_Register_Offsets Register Offsets
 * @brief      Clock Management Controller Register Offsets from the CLKMAN Base Peripheral Address. 
 * @{
 */
#define MXC_R_CLKMAN_OFFS_CLK_CONFIG                        ((uint32_t)0x00000000UL)   /**< Offset from the CLKMAN Base Peripheral Address:<b><tt>0x0000</tt></b> */
#define MXC_R_CLKMAN_OFFS_CLK_CTRL                          ((uint32_t)0x00000004UL)   /**< Offset from the CLKMAN Base Peripheral Address:<b><tt>0x0004</tt></b> */
#define MXC_R_CLKMAN_OFFS_INTFL                             ((uint32_t)0x00000008UL)   /**< Offset from the CLKMAN Base Peripheral Address:<b><tt>0x0008</tt></b> */
#define MXC_R_CLKMAN_OFFS_INTEN                             ((uint32_t)0x0000000CUL)   /**< Offset from the CLKMAN Base Peripheral Address:<b><tt>0x000C</tt></b> */
#define MXC_R_CLKMAN_OFFS_TRIM_CALC                         ((uint32_t)0x00000010UL)   /**< Offset from the CLKMAN Base Peripheral Address:<b><tt>0x0010</tt></b> */
#define MXC_R_CLKMAN_OFFS_I2C_TIMER_CTRL                    ((uint32_t)0x00000014UL)   /**< Offset from the CLKMAN Base Peripheral Address:<b><tt>0x0014</tt></b> */
#define MXC_R_CLKMAN_OFFS_CM4_START_CLK_EN0                 ((uint32_t)0x00000018UL)   /**< Offset from the CLKMAN Base Peripheral Address:<b><tt>0x0018</tt></b> */
#define MXC_R_CLKMAN_OFFS_CM4_START_CLK_EN1                 ((uint32_t)0x0000001CUL)   /**< Offset from the CLKMAN Base Peripheral Address:<b><tt>0x001C</tt></b> */
#define MXC_R_CLKMAN_OFFS_CM4_START_CLK_EN2                 ((uint32_t)0x00000020UL)   /**< Offset from the CLKMAN Base Peripheral Address:<b><tt>0x0020</tt></b> */
#define MXC_R_CLKMAN_OFFS_SYS_CLK_CTRL_0_CM4                ((uint32_t)0x00000040UL)   /**< Offset from the CLKMAN Base Peripheral Address:<b><tt>0x0040</tt></b> */
#define MXC_R_CLKMAN_OFFS_SYS_CLK_CTRL_1_SYNC               ((uint32_t)0x00000044UL)   /**< Offset from the CLKMAN Base Peripheral Address:<b><tt>0x0044</tt></b> */
#define MXC_R_CLKMAN_OFFS_SYS_CLK_CTRL_2_SPIX               ((uint32_t)0x00000048UL)   /**< Offset from the CLKMAN Base Peripheral Address:<b><tt>0x0048</tt></b> */
#define MXC_R_CLKMAN_OFFS_SYS_CLK_CTRL_3_PRNG               ((uint32_t)0x0000004CUL)   /**< Offset from the CLKMAN Base Peripheral Address:<b><tt>0x004C</tt></b> */
#define MXC_R_CLKMAN_OFFS_SYS_CLK_CTRL_4_WDT0               ((uint32_t)0x00000050UL)   /**< Offset from the CLKMAN Base Peripheral Address:<b><tt>0x0050</tt></b> */
#define MXC_R_CLKMAN_OFFS_SYS_CLK_CTRL_5_WDT1               ((uint32_t)0x00000054UL)   /**< Offset from the CLKMAN Base Peripheral Address:<b><tt>0x0054</tt></b> */
#define MXC_R_CLKMAN_OFFS_SYS_CLK_CTRL_6_GPIO               ((uint32_t)0x00000058UL)   /**< Offset from the CLKMAN Base Peripheral Address:<b><tt>0x0058</tt></b> */
#define MXC_R_CLKMAN_OFFS_SYS_CLK_CTRL_7_PT                 ((uint32_t)0x0000005CUL)   /**< Offset from the CLKMAN Base Peripheral Address:<b><tt>0x005C</tt></b> */
#define MXC_R_CLKMAN_OFFS_SYS_CLK_CTRL_8_UART               ((uint32_t)0x00000060UL)   /**< Offset from the CLKMAN Base Peripheral Address:<b><tt>0x0060</tt></b> */
#define MXC_R_CLKMAN_OFFS_SYS_CLK_CTRL_9_I2CM               ((uint32_t)0x00000064UL)   /**< Offset from the CLKMAN Base Peripheral Address:<b><tt>0x0064</tt></b> */
#define MXC_R_CLKMAN_OFFS_SYS_CLK_CTRL_10_I2CS              ((uint32_t)0x00000068UL)   /**< Offset from the CLKMAN Base Peripheral Address:<b><tt>0x0068</tt></b> */
#define MXC_R_CLKMAN_OFFS_SYS_CLK_CTRL_11_SPI0              ((uint32_t)0x0000006CUL)   /**< Offset from the CLKMAN Base Peripheral Address:<b><tt>0x006C</tt></b> */
#define MXC_R_CLKMAN_OFFS_SYS_CLK_CTRL_12_SPI1              ((uint32_t)0x00000070UL)   /**< Offset from the CLKMAN Base Peripheral Address:<b><tt>0x0070</tt></b> */
#define MXC_R_CLKMAN_OFFS_SYS_CLK_CTRL_13_SPI2              ((uint32_t)0x00000074UL)   /**< Offset from the CLKMAN Base Peripheral Address:<b><tt>0x0074</tt></b> */
#define MXC_R_CLKMAN_OFFS_SYS_CLK_CTRL_14_SPIB              ((uint32_t)0x00000078UL)   /**< Offset from the CLKMAN Base Peripheral Address:<b><tt>0x0078</tt></b> */
#define MXC_R_CLKMAN_OFFS_SYS_CLK_CTRL_15_OWM               ((uint32_t)0x0000007CUL)   /**< Offset from the CLKMAN Base Peripheral Address:<b><tt>0x007C</tt></b> */
#define MXC_R_CLKMAN_OFFS_SYS_CLK_CTRL_16_SPIS              ((uint32_t)0x00000080UL)   /**< Offset from the CLKMAN Base Peripheral Address:<b><tt>0x0080</tt></b> */
#define MXC_R_CLKMAN_OFFS_CRYPT_CLK_CTRL_0_AES              ((uint32_t)0x00000100UL)   /**< Offset from the CLKMAN Base Peripheral Address:<b><tt>0x0100</tt></b> */
#define MXC_R_CLKMAN_OFFS_CRYPT_CLK_CTRL_1_MAA              ((uint32_t)0x00000104UL)   /**< Offset from the CLKMAN Base Peripheral Address:<b><tt>0x0104</tt></b> */
#define MXC_R_CLKMAN_OFFS_CRYPT_CLK_CTRL_2_PRNG             ((uint32_t)0x00000108UL)   /**< Offset from the CLKMAN Base Peripheral Address:<b><tt>0x0108</tt></b> */
#define MXC_R_CLKMAN_OFFS_CLK_GATE_CTRL0                    ((uint32_t)0x00000140UL)   /**< Offset from the CLKMAN Base Peripheral Address:<b><tt>0x0140</tt></b> */
#define MXC_R_CLKMAN_OFFS_CLK_GATE_CTRL1                    ((uint32_t)0x00000144UL)   /**< Offset from the CLKMAN Base Peripheral Address:<b><tt>0x0144</tt></b> */
#define MXC_R_CLKMAN_OFFS_CLK_GATE_CTRL2                    ((uint32_t)0x00000148UL)   /**< Offset from the CLKMAN Base Peripheral Address:<b><tt>0x0148</tt></b> */
/**@} end of CLKMAN_Register_Offsets */
/**
 * @ingroup     clkman_registers
 * @defgroup    clkman_clk_config CLKMAN_CLK_CONFIG Register
 * @brief       Field Positions and Masks 
 */
#define MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_ENABLE_POS           0                                                                                   /**< CRYPTO_ENABLE Position              */
#define MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_ENABLE               ((uint32_t)(0x00000001UL << MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_ENABLE_POS))             /**< CRYPTO_ENABLE Mask                  */
#define MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_POS  4                                                                                   /**< CRYPTO_STABILITY_COUNT Position     */
#define MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT      ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_POS))    /**< CRYPTO_STABILITY_COUNT Mask         */
/**@}*/
/**
 * @ingroup     clkman_registers
 * @defgroup   clkman_clk_ctrl CLKMAN_CLK_CTRL Register
 * @brief      Field Positions and Masks 
 */
#define MXC_F_CLKMAN_CLK_CTRL_SYSTEM_SOURCE_SELECT_POS      0                                                                                   /**< SYSTEM_SOURCE_SELECT Position         */
#define MXC_F_CLKMAN_CLK_CTRL_SYSTEM_SOURCE_SELECT          ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_CTRL_SYSTEM_SOURCE_SELECT_POS))        /**< SYSTEM_SOURCE_SELECT Mask             */
#define MXC_F_CLKMAN_CLK_CTRL_USB_CLOCK_ENABLE_POS          4                                                                                   /**< USB_CLOCK_ENABLE Position             */
#define MXC_F_CLKMAN_CLK_CTRL_USB_CLOCK_ENABLE              ((uint32_t)(0x00000001UL << MXC_F_CLKMAN_CLK_CTRL_USB_CLOCK_ENABLE_POS))            /**< USB_CLOCK_ENABLE Mask                 */
#define MXC_F_CLKMAN_CLK_CTRL_USB_CLOCK_SELECT_POS          5                                                                                   /**< USB_CLOCK_SELECT Position             */
#define MXC_F_CLKMAN_CLK_CTRL_USB_CLOCK_SELECT              ((uint32_t)(0x00000001UL << MXC_F_CLKMAN_CLK_CTRL_USB_CLOCK_SELECT_POS))            /**< USB_CLOCK_SELECT Mask                 */
#define MXC_F_CLKMAN_CLK_CTRL_CRYPTO_CLOCK_ENABLE_POS       8                                                                                   /**< CRYPTO_CLOCK_ENABLE Position          */
#define MXC_F_CLKMAN_CLK_CTRL_CRYPTO_CLOCK_ENABLE           ((uint32_t)(0x00000001UL << MXC_F_CLKMAN_CLK_CTRL_CRYPTO_CLOCK_ENABLE_POS))         /**< CRYPTO_CLOCK_ENABLE Mask              */
#define MXC_F_CLKMAN_CLK_CTRL_RTOS_MODE_POS                 12                                                                                  /**< RTOS_MODE Field Position              */
#define MXC_F_CLKMAN_CLK_CTRL_RTOS_MODE                     ((uint32_t)(0x00000001UL << MXC_F_CLKMAN_CLK_CTRL_RTOS_MODE_POS))                   /**< RTOS_MODE Field Mask                  */
#define MXC_F_CLKMAN_CLK_CTRL_CPU_DYNAMIC_CLOCK_POS         13                                                                                  /**< CPU_DYNAMIC_CLOCK Field Position      */
#define MXC_F_CLKMAN_CLK_CTRL_CPU_DYNAMIC_CLOCK             ((uint32_t)(0x00000001UL << MXC_F_CLKMAN_CLK_CTRL_CPU_DYNAMIC_CLOCK_POS))           /**< CPU_DYNAMIC_CLOCK Field Mask          */
#define MXC_F_CLKMAN_CLK_CTRL_WDT0_CLOCK_ENABLE_POS         16                                                                                  /**< WDT0_CLOCK_ENABLE Field Position      */
#define MXC_F_CLKMAN_CLK_CTRL_WDT0_CLOCK_ENABLE             ((uint32_t)(0x00000001UL << MXC_F_CLKMAN_CLK_CTRL_WDT0_CLOCK_ENABLE_POS))           /**< WDT0_CLOCK_ENABLE Field Mask          */
#define MXC_F_CLKMAN_CLK_CTRL_WDT0_CLOCK_SELECT_POS         17                                                                                  /**< WDT0_CLOCK_SELECT Field Position      */
#define MXC_F_CLKMAN_CLK_CTRL_WDT0_CLOCK_SELECT             ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_CTRL_WDT0_CLOCK_SELECT_POS))           /**< WDT0_CLOCK_SELECT Field Mask          */
#define MXC_F_CLKMAN_CLK_CTRL_WDT1_CLOCK_ENABLE_POS         20                                                                                  /**< WDT1_CLOCK_ENABLE Field Position      */
#define MXC_F_CLKMAN_CLK_CTRL_WDT1_CLOCK_ENABLE             ((uint32_t)(0x00000001UL << MXC_F_CLKMAN_CLK_CTRL_WDT1_CLOCK_ENABLE_POS))           /**< WDT1_CLOCK_ENABLE Field Mask          */
#define MXC_F_CLKMAN_CLK_CTRL_WDT1_CLOCK_SELECT_POS         21                                                                                  /**< WDT1_CLOCK_SELECT Field Position      */
#define MXC_F_CLKMAN_CLK_CTRL_WDT1_CLOCK_SELECT             ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_CTRL_WDT1_CLOCK_SELECT_POS))           /**< WDT1_CLOCK_SELECT Field Mask          */
#define MXC_F_CLKMAN_CLK_CTRL_ADC_CLOCK_ENABLE_POS          24                                                                                  /**< ADC_CLOCK_ENABLE Field Position       */
#define MXC_F_CLKMAN_CLK_CTRL_ADC_CLOCK_ENABLE              ((uint32_t)(0x00000001UL << MXC_F_CLKMAN_CLK_CTRL_ADC_CLOCK_ENABLE_POS))            /**< ADC_CLOCK_ENABLE Field Mask           */
/**@}*/
/**
 * @ingroup     clkman_registers
 * @defgroup   clkman_int_flags CLKMAN_INTFL Register
 * @brief      Interrupt Flag Positions and Masks 
 */
#define MXC_F_CLKMAN_INTFL_CRYPTO_STABLE_POS                0                                                                                   /**< CRYPTO_STABLE Interrupt Flag Position    */
#define MXC_F_CLKMAN_INTFL_CRYPTO_STABLE                    ((uint32_t)(0x00000001UL << MXC_F_CLKMAN_INTFL_CRYPTO_STABLE_POS))                  /**< CRYPTO_STABLE Interrupt Flag Mask        */
#define MXC_F_CLKMAN_INTFL_SYS_RO_STABLE_POS                1                                                                                   /**< SYS_RO_STABLE Interrupt Flag Position    */
#define MXC_F_CLKMAN_INTFL_SYS_RO_STABLE                    ((uint32_t)(0x00000001UL << MXC_F_CLKMAN_INTFL_SYS_RO_STABLE_POS))                  /**< SYS_RO_STABLE Interrupt Flag Mask        */
/**@}*/
/**
 * @ingroup     clkman_registers
 * @defgroup   clkman_int_enable CLKMAN_INTEN Register
 * @brief      Interrupt Enable Positions and Masks 
 */
#define MXC_F_CLKMAN_INTEN_CRYPTO_STABLE_POS                0                                                                                   /**< CRYPTO_STABLE Field Position         */
#define MXC_F_CLKMAN_INTEN_CRYPTO_STABLE                    ((uint32_t)(0x00000001UL << MXC_F_CLKMAN_INTEN_CRYPTO_STABLE_POS))                  /**< CRYPTO_STABLE Field Mask             */
#define MXC_F_CLKMAN_INTEN_SYS_RO_STABLE_POS                1                                                                                   /**< SYS_RO_STABLE Field Position         */
#define MXC_F_CLKMAN_INTEN_SYS_RO_STABLE                    ((uint32_t)(0x00000001UL << MXC_F_CLKMAN_INTEN_SYS_RO_STABLE_POS))                  /**< SYS_RO_STABLE Field Mask             */
/**@}*/
/**
 * @ingroup     clkman_registers
 * @defgroup   clkman_trim_calc CLKMAN_TRIM_CALC Register
 * @brief      Field Positions and Masks 
 */
#define MXC_F_CLKMAN_TRIM_CALC_TRIM_CLK_SEL_POS             0                                                                                   /**< TRIM_CLK_SEL Field Position          */
#define MXC_F_CLKMAN_TRIM_CALC_TRIM_CLK_SEL                 ((uint32_t)(0x00000001UL << MXC_F_CLKMAN_TRIM_CALC_TRIM_CLK_SEL_POS))               /**< TRIM_CLK_SEL Field Mask              */    
#define MXC_F_CLKMAN_TRIM_CALC_TRIM_CALC_START_POS          1                                                                                   /**< TRIM_CALC_START Field Position       */
#define MXC_F_CLKMAN_TRIM_CALC_TRIM_CALC_START              ((uint32_t)(0x00000001UL << MXC_F_CLKMAN_TRIM_CALC_TRIM_CALC_START_POS))            /**< TRIM_CALC_START Field Mask           */        
#define MXC_F_CLKMAN_TRIM_CALC_TRIM_CALC_COMPLETED_POS      2                                                                                   /**< TRIM_CALC_COMPLETED Field Position   */
#define MXC_F_CLKMAN_TRIM_CALC_TRIM_CALC_COMPLETED          ((uint32_t)(0x00000001UL << MXC_F_CLKMAN_TRIM_CALC_TRIM_CALC_COMPLETED_POS))        /**< TRIM_CALC_COMPLETED Field Mask       */            
#define MXC_F_CLKMAN_TRIM_CALC_TRIM_ENABLE_POS              3                                                                                   /**< TRIM_ENABLE Field Position           */
#define MXC_F_CLKMAN_TRIM_CALC_TRIM_ENABLE                  ((uint32_t)(0x00000001UL << MXC_F_CLKMAN_TRIM_CALC_TRIM_ENABLE_POS))                /**< TRIM_ENABLE Field Mask               */    
#define MXC_F_CLKMAN_TRIM_CALC_TRIM_CALC_RESULTS_POS        16                                                                                  /**< TRIM_CALC_RESULTS Field Position     */
#define MXC_F_CLKMAN_TRIM_CALC_TRIM_CALC_RESULTS            ((uint32_t)(0x000003FFUL << MXC_F_CLKMAN_TRIM_CALC_TRIM_CALC_RESULTS_POS))          /**< TRIM_CALC_RESULTS Field Mask         */        
/**@}*/
/**
 * @ingroup     clkman_registers
 * @defgroup    clkman_i2c_1ms CLKMAN_I2C_TIMER_CTRL Register
 * @brief       Field Positions and Masks 
 */
#define MXC_F_CLKMAN_I2C_TIMER_CTRL_I2C_1MS_TIMER_EN_POS    0                                                                                   /**< I2C_1MS_TIMER_EN Position       */
#define MXC_F_CLKMAN_I2C_TIMER_CTRL_I2C_1MS_TIMER_EN        ((uint32_t)(0x00000001UL << MXC_F_CLKMAN_I2C_TIMER_CTRL_I2C_1MS_TIMER_EN_POS))      /**< I2C_1MS_TIMER_EN Mask           */
/**@}*/
/**
 * @ingroup     clkman_registers
 * @defgroup    clkman_cm4 CLKMAN_CM4 Register
 * @brief       Field Positions and Masks 
 */
#define MXC_F_CLKMAN_CM4_START_CLK_EN0_INTS_POS             0                                                                                   /**< CLK_EN0_INTS Position                */
#define MXC_F_CLKMAN_CM4_START_CLK_EN0_INTS                 ((uint32_t)(0xFFFFFFFFUL << MXC_F_CLKMAN_CM4_START_CLK_EN0_INTS_POS))               /**< CLK_EN0_INTS Mask                    */

#define MXC_F_CLKMAN_CM4_START_CLK_EN1_INTS_POS             0                                                                                   /**< CLK_EN1_INTS Position                */
#define MXC_F_CLKMAN_CM4_START_CLK_EN1_INTS                 ((uint32_t)(0xFFFFFFFFUL << MXC_F_CLKMAN_CM4_START_CLK_EN1_INTS_POS))               /**< CLK_EN1_INTS Mask                    */

#define MXC_F_CLKMAN_CM4_START_CLK_EN2_INTS_POS             0                                                                                   /**< CLK_EN2_INTS Position                */
#define MXC_F_CLKMAN_CM4_START_CLK_EN2_INTS                 ((uint32_t)(0xFFFFFFFFUL << MXC_F_CLKMAN_CM4_START_CLK_EN2_INTS_POS))               /**< CLK_EN2_INTS Mask                    */
/**@}*/
/**
 * @ingroup     clkman_registers
 * @defgroup    clkman_sysclk_ctrl CLKMAN_SYS_CLK_CTRL Register
 * @brief       Field Positions and Masks 
 */
#define MXC_F_CLKMAN_SYS_CLK_CTRL_0_CM4_CM4_CLK_SCALE_POS   0                                                                                           /**< CM4_CM4_CLK_SCALE Position      */
#define MXC_F_CLKMAN_SYS_CLK_CTRL_0_CM4_CM4_CLK_SCALE       ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_SYS_CLK_CTRL_0_CM4_CM4_CLK_SCALE_POS))             /**< CM4_CM4_CLK_SCALE Mask          */
#define MXC_F_CLKMAN_SYS_CLK_CTRL_1_SYNC_SYNC_CLK_SCALE_POS  0                                                                                          /**< SYNC_SYNC_CLK_SCALE Position    */
#define MXC_F_CLKMAN_SYS_CLK_CTRL_1_SYNC_SYNC_CLK_SCALE     ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_SYS_CLK_CTRL_1_SYNC_SYNC_CLK_SCALE_POS))           /**< SYNC_SYNC_CLK_SCALE Mask        */
#define MXC_F_CLKMAN_SYS_CLK_CTRL_2_SPIX_SPIX_CLK_SCALE_POS  0                                                                                          /**< SPIX_SPIX_CLK_SCALE Position    */
#define MXC_F_CLKMAN_SYS_CLK_CTRL_2_SPIX_SPIX_CLK_SCALE     ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_SYS_CLK_CTRL_2_SPIX_SPIX_CLK_SCALE_POS))           /**< SPIX_SPIX_CLK_SCALE Mask        */
#define MXC_F_CLKMAN_SYS_CLK_CTRL_3_PRNG_PRNG_CLK_SCALE_POS  0                                                                                          /**< PRNG_PRNG_CLK_SCALE Position    */
#define MXC_F_CLKMAN_SYS_CLK_CTRL_3_PRNG_PRNG_CLK_SCALE     ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_SYS_CLK_CTRL_3_PRNG_PRNG_CLK_SCALE_POS))           /**< PRNG_PRNG_CLK_SCALE Mask        */
#define MXC_F_CLKMAN_SYS_CLK_CTRL_4_WDT0_WATCHDOG0_CLK_SCALE_POS  0                                                                                     /**< WDT0_WATCHDOG0_CLK_ Position    */
#define MXC_F_CLKMAN_SYS_CLK_CTRL_4_WDT0_WATCHDOG0_CLK_SCALE  ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_SYS_CLK_CTRL_4_WDT0_WATCHDOG0_CLK_SCALE_POS))    /**< WDT0_WATCHDOG0_CLK_ Mask        */
#define MXC_F_CLKMAN_SYS_CLK_CTRL_5_WDT1_WATCHDOG1_CLK_SCALE_POS  0                                                                                     /**< WDT1_WATCHDOG1_CLK_ Position    */
#define MXC_F_CLKMAN_SYS_CLK_CTRL_5_WDT1_WATCHDOG1_CLK_SCALE  ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_SYS_CLK_CTRL_5_WDT1_WATCHDOG1_CLK_SCALE_POS))    /**< WDT1_WATCHDOG1_CLK_ Mask        */
#define MXC_F_CLKMAN_SYS_CLK_CTRL_6_GPIO_GPIO_CLK_SCALE_POS  0                                                                                          /**< GPIO_GPIO_CLK_SCALE Position    */
#define MXC_F_CLKMAN_SYS_CLK_CTRL_6_GPIO_GPIO_CLK_SCALE     ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_SYS_CLK_CTRL_6_GPIO_GPIO_CLK_SCALE_POS))           /**< GPIO_GPIO_CLK_SCALE Mask        */
#define MXC_F_CLKMAN_SYS_CLK_CTRL_7_PT_PULSE_TRAIN_CLK_SCALE_POS  0                                                                                     /**< PT_PULSE_TRAIN_CLK_ Position    */
#define MXC_F_CLKMAN_SYS_CLK_CTRL_7_PT_PULSE_TRAIN_CLK_SCALE  ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_SYS_CLK_CTRL_7_PT_PULSE_TRAIN_CLK_SCALE_POS))    /**< PT_PULSE_TRAIN_CLK_ Mask        */
#define MXC_F_CLKMAN_SYS_CLK_CTRL_8_UART_UART_CLK_SCALE_POS  0                                                                                          /**< UART_UART_CLK_SCALE Position    */
#define MXC_F_CLKMAN_SYS_CLK_CTRL_8_UART_UART_CLK_SCALE     ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_SYS_CLK_CTRL_8_UART_UART_CLK_SCALE_POS))           /**< UART_UART_CLK_SCALE Mask        */
#define MXC_F_CLKMAN_SYS_CLK_CTRL_9_I2CM_I2CM_CLK_SCALE_POS  0                                                                                          /**< I2CM_I2CM_CLK_SCALE Position    */
#define MXC_F_CLKMAN_SYS_CLK_CTRL_9_I2CM_I2CM_CLK_SCALE     ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_SYS_CLK_CTRL_9_I2CM_I2CM_CLK_SCALE_POS))           /**< I2CM_I2CM_CLK_SCALE Mask        */
#define MXC_F_CLKMAN_SYS_CLK_CTRL_10_I2CS_I2CS_CLK_SCALE_POS  0                                                                                         /**< I2CS_I2CS_CLK_SCALE Position  */                                                                                 
#define MXC_F_CLKMAN_SYS_CLK_CTRL_10_I2CS_I2CS_CLK_SCALE    ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_SYS_CLK_CTRL_10_I2CS_I2CS_CLK_SCALE_POS))          /**< I2CS_I2CS_CLK_SCALE Mask      */
#define MXC_F_CLKMAN_SYS_CLK_CTRL_11_SPI0_SPI0_CLK_SCALE_POS  0                                                                                         /**< PI0_SPI0_CLK_SCALE Position   */                                                                                 
#define MXC_F_CLKMAN_SYS_CLK_CTRL_11_SPI0_SPI0_CLK_SCALE    ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_SYS_CLK_CTRL_11_SPI0_SPI0_CLK_SCALE_POS))          /**< SPI0_SPI0_CLK_SCALE Mask      */
#define MXC_F_CLKMAN_SYS_CLK_CTRL_12_SPI1_SPI1_CLK_SCALE_POS  0                                                                                         /**< SPI1_SPI1_CLK_SCALE Position  */                                                                                 
#define MXC_F_CLKMAN_SYS_CLK_CTRL_12_SPI1_SPI1_CLK_SCALE    ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_SYS_CLK_CTRL_12_SPI1_SPI1_CLK_SCALE_POS))          /**< SPI1_SPI1_CLK_SCALE Mask       */
#define MXC_F_CLKMAN_SYS_CLK_CTRL_13_SPI2_SPI2_CLK_SCALE_POS  0                                                                                         /**< SPI2_SPI2_CLK_SCALE Position  */                                                                                 
#define MXC_F_CLKMAN_SYS_CLK_CTRL_13_SPI2_SPI2_CLK_SCALE    ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_SYS_CLK_CTRL_13_SPI2_SPI2_CLK_SCALE_POS))          /**< SPI2_SPI2_CLK_SCALE Mask     */
#define MXC_F_CLKMAN_SYS_CLK_CTRL_14_SPIB_SPIB_CLK_SCALE_POS  0                                                                                         /**< SPIB_SPIB_CLK_SCALE Position  */                                                                                 
#define MXC_F_CLKMAN_SYS_CLK_CTRL_14_SPIB_SPIB_CLK_SCALE    ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_SYS_CLK_CTRL_14_SPIB_SPIB_CLK_SCALE_POS))          /**< SPIB_SPIB_CLK_SCALE Mask      */
#define MXC_F_CLKMAN_SYS_CLK_CTRL_15_OWM_OWM_CLK_SCALE_POS  0                                                                                           /**< OWM_OWM_CLK_SCALE Position    */                                                                                   
#define MXC_F_CLKMAN_SYS_CLK_CTRL_15_OWM_OWM_CLK_SCALE      ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_SYS_CLK_CTRL_15_OWM_OWM_CLK_SCALE_POS))            /**< OWM_OWM_CLK_SCALE Mask        */
#define MXC_F_CLKMAN_SYS_CLK_CTRL_16_SPIS_SPIS_CLK_SCALE_POS  0                                                                                         /**< PIS_SPIS_CLK_SCALE Position   */                                                                                 
#define MXC_F_CLKMAN_SYS_CLK_CTRL_16_SPIS_SPIS_CLK_SCALE    ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_SYS_CLK_CTRL_16_SPIS_SPIS_CLK_SCALE_POS))          /**< SPIS_SPIS_CLK_SCALE Mask      */
/**@}*/
/**
 * @ingroup     clkman_registers
 * @defgroup    clkman_crypt_clk_ctrl CLKMAN_CRYPT_CLK_CTRL Register
 * @brief       Field Positions and Masks 
 */
#define MXC_F_CLKMAN_CRYPT_CLK_CTRL_0_AES_AES_CLK_SCALE_POS  0                                                                                          /**< AES_AES_CLK_SCALE Position    */                                                                                  
#define MXC_F_CLKMAN_CRYPT_CLK_CTRL_0_AES_AES_CLK_SCALE      ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_CRYPT_CLK_CTRL_0_AES_AES_CLK_SCALE_POS))          /**< AES_AES_CLK_SCALE Mask       */
#define MXC_F_CLKMAN_CRYPT_CLK_CTRL_1_MAA_MAA_CLK_SCALE_POS  0                                                                                          /**< MAA_MAA_CLK_SCALE Position    */                                                                                  
#define MXC_F_CLKMAN_CRYPT_CLK_CTRL_1_MAA_MAA_CLK_SCALE      ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_CRYPT_CLK_CTRL_1_MAA_MAA_CLK_SCALE_POS))          /**< MAA_MAA_CLK_SCALE Mask        */
#define MXC_F_CLKMAN_CRYPT_CLK_CTRL_2_PRNG_PRNG_CLK_SCALE_POS  0                                                                                        /**< PRNG_PRNG_CLK_SCALE Position  */                                                                                    
#define MXC_F_CLKMAN_CRYPT_CLK_CTRL_2_PRNG_PRNG_CLK_SCALE    ((uint32_t)(0x0000000FUL << MXC_F_CLKMAN_CRYPT_CLK_CTRL_2_PRNG_PRNG_CLK_SCALE_POS))        /**< PRNG_PRNG_CLK_SCALE Mask      */
/**@}*/
/**
 * @ingroup     clkman_registers
 * @defgroup    clkman_clk_gate_ctrl CLKMAN_CLK_GATE_CTRL Register
 * @brief       Peripheral Clock Gating Field Positions and Masks 
 */
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_CM4_CLK_GATER_POS       0                                                                                           /**< CM4_CLK_GATER Position          */
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_CM4_CLK_GATER           ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL0_CM4_CLK_GATER_POS))                 /**< CM4_CLK_GATER Mask              */
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_AHB32_CLK_GATER_POS     2                                                                                           /**< AHB32_CLK_GATER Position        */
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_AHB32_CLK_GATER         ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL0_AHB32_CLK_GATER_POS))               /**< AHB32_CLK_GATER Mask            */
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_ICACHE_CLK_GATER_POS    4                                                                                           /**< ICACHE_CLK_GATER Position       */
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_ICACHE_CLK_GATER        ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL0_ICACHE_CLK_GATER_POS))              /**< ICACHE_CLK_GATER Mask           */
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_FLASH_CLK_GATER_POS     6                                                                                           /**< FLASH_CLK_GATER Position        */
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_FLASH_CLK_GATER         ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL0_FLASH_CLK_GATER_POS))               /**< FLASH_CLK_GATER Mask            */
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_SRAM_CLK_GATER_POS      8                                                                                           /**< SRAM_CLK_GATER Position         */
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_SRAM_CLK_GATER          ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL0_SRAM_CLK_GATER_POS))                /**< SRAM_CLK_GATER Mask             */
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_APB_BRIDGE_CLK_GATER_POS  10                                                                                        /**< APB_BRIDGE_CLK_GATER Position   */
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_APB_BRIDGE_CLK_GATER    ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL0_APB_BRIDGE_CLK_GATER_POS))          /**< APB_BRIDGE_CLK_GATER Mask       */
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_SYSMAN_CLK_GATER_POS    12                                                                                          /**< SYSMAN_CLK_GATER Position       */
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_SYSMAN_CLK_GATER        ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL0_SYSMAN_CLK_GATER_POS))              /**< SYSMAN_CLK_GATER Mask           */
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_PTP_CLK_GATER_POS       14                                                                                          /**< PTP_CLK_GATER Position          */
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_PTP_CLK_GATER           ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL0_PTP_CLK_GATER_POS))                 /**< PTP_CLK_GATER Mask              */
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_SSB_MUX_CLK_GATER_POS   16                                                                                          /**< SSB_MUX_CLK_GATER Position      */
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_SSB_MUX_CLK_GATER       ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL0_SSB_MUX_CLK_GATER_POS))             /**< SSB_MUX_CLK_GATER Mask          */
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_PAD_CLK_GATER_POS       18                                                                                          /**< PAD_CLK_GATER Position          */
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_PAD_CLK_GATER           ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL0_PAD_CLK_GATER_POS))                 /**< PAD_CLK_GATER Mask              */
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_SPIX_CLK_GATER_POS      20                                                                                          /**< SPIX_CLK_GATER Position         */
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_SPIX_CLK_GATER          ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL0_SPIX_CLK_GATER_POS))                /**< SPIX_CLK_GATER Mask             */
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_PMU_CLK_GATER_POS       22                                                                                          /**< PMU_CLK_GATER Position          */
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_PMU_CLK_GATER           ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL0_PMU_CLK_GATER_POS))                 /**< PMU_CLK_GATER Mask              */
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_USB_CLK_GATER_POS       24                                                                                          /**< USB_CLK_GATER Position          */
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_USB_CLK_GATER           ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL0_USB_CLK_GATER_POS))                 /**< USB_CLK_GATER Mask              */
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_CRC_CLK_GATER_POS       26                                                                                          /**< CRC_CLK_GATER Position          */
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_CRC_CLK_GATER           ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL0_CRC_CLK_GATER_POS))                 /**< CRC_CLK_GATER Mask              */
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_TPU_CLK_GATER_POS       28                                                                                          /**< TPU_CLK_GATER Position          */
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_TPU_CLK_GATER           ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL0_TPU_CLK_GATER_POS))                 /**< TPU_CLK_GATER Mask              */
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_WATCHDOG0_CLK_GATER_POS  30                                                                                         /**< WATCHDOG0_CLK_GATER Position    */
#define MXC_F_CLKMAN_CLK_GATE_CTRL0_WATCHDOG0_CLK_GATER     ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL0_WATCHDOG0_CLK_GATER_POS))           /**< WATCHDOG0_CLK_GATER  Mask       */
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_WATCHDOG1_CLK_GATER_POS  0                                                                                          /**< WATCHDOG1_CLK_GATER Position    */ 
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_WATCHDOG1_CLK_GATER     ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL1_WATCHDOG1_CLK_GATER_POS))           /**< WATCHDOG1_CLK_GATER Mask        */           
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_GPIO_CLK_GATER_POS      2                                                                                           /**< GPIO_CLK_GATER Position         */ 
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_GPIO_CLK_GATER          ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL1_GPIO_CLK_GATER_POS))                /**< GPIO_CLK_GATER Mask             */               
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_TIMER0_CLK_GATER_POS    4                                                                                           /**< TIMER0_CLK_GATER Position       */     
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_TIMER0_CLK_GATER        ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL1_TIMER0_CLK_GATER_POS))              /**< TIMER0_CLK_GATER Mask           */              
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_TIMER1_CLK_GATER_POS    6                                                                                           /**< TIMER1_CLK_GATER Position       */     
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_TIMER1_CLK_GATER        ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL1_TIMER1_CLK_GATER_POS))              /**< TIMER1_CLK_GATER Mask           */          
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_TIMER2_CLK_GATER_POS    8                                                                                           /**< TIMER2_CLK_GATER Position       */ 
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_TIMER2_CLK_GATER        ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL1_TIMER2_CLK_GATER_POS))              /**< TIMER2_CLK_GATER Mask           */          
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_TIMER3_CLK_GATER_POS    10                                                                                          /**< TIMER3_CLK_GATER Position       */ 
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_TIMER3_CLK_GATER        ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL1_TIMER3_CLK_GATER_POS))              /**< TIMER3_CLK_GATER Mask           */         
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_TIMER4_CLK_GATER_POS    12                                                                                          /**< TIMER4_CLK_GATER Position       */ 
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_TIMER4_CLK_GATER        ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL1_TIMER4_CLK_GATER_POS))              /**< TIMER4_CLK_GATER Mask           */          
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_TIMER5_CLK_GATER_POS    14                                                                                          /**< TIMER5_CLK_GATER Position       */ 
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_TIMER5_CLK_GATER        ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL1_TIMER5_CLK_GATER_POS))              /**< TIMER5_CLK_GATER Mask           */          
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_PULSETRAIN_CLK_GATER_POS  16                                                                                        /**< PULSETRAIN_CLK_GATER Position   */  
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_PULSETRAIN_CLK_GATER    ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL1_PULSETRAIN_CLK_GATER_POS))          /**< PULSETRAIN_CLK_GATER  Mask      */         
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_UART0_CLK_GATER_POS     18                                                                                          /**< UART0_CLK_GATER Position        */ 
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_UART0_CLK_GATER         ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL1_UART0_CLK_GATER_POS))               /**< UART0_CLK_GATER Mask            */           
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_UART1_CLK_GATER_POS     20                                                                                          /**< UART1_CLK_GATER Position        */ 
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_UART1_CLK_GATER         ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL1_UART1_CLK_GATER_POS))               /**< UART1_CLK_GATER Mask            */           
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_UART2_CLK_GATER_POS     22                                                                                          /**< UART2_CLK_GATER Position        */ 
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_UART2_CLK_GATER         ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL1_UART2_CLK_GATER_POS))               /**< UART2_CLK_GATER Mask            */           
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_UART3_CLK_GATER_POS     24                                                                                          /**< UART3_CLK_GATER Position        */ 
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_UART3_CLK_GATER         ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL1_UART3_CLK_GATER_POS))               /**< UART3_CLK_GATER Mask            */           
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_I2CM0_CLK_GATER_POS     26                                                                                          /**< I2CM0_CLK_GATER Position        */ 
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_I2CM0_CLK_GATER         ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL1_I2CM0_CLK_GATER_POS))               /**< I2CM0_CLK_GATER Mask            */           
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_I2CM1_CLK_GATER_POS     28                                                                                          /**< I2CM1_CLK_GATER Position        */ 
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_I2CM1_CLK_GATER         ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL1_I2CM1_CLK_GATER_POS))               /**< I2CM1_CLK_GATER Mask            */           
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_I2CM2_CLK_GATER_POS     30                                                                                          /**< I2CM2_CLK_GATER Position        */ 
#define MXC_F_CLKMAN_CLK_GATE_CTRL1_I2CM2_CLK_GATER         ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL1_I2CM2_CLK_GATER_POS))               /**< I2CM2_CLK_GATER Mask            */           
#define MXC_F_CLKMAN_CLK_GATE_CTRL2_I2CS_CLK_GATER_POS      0                                                                                           /**< I2CS_CLK_GATER Position         */ 
#define MXC_F_CLKMAN_CLK_GATE_CTRL2_I2CS_CLK_GATER          ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL2_I2CS_CLK_GATER_POS))                /**< I2CS_CLK_GATER  Mask            */           
#define MXC_F_CLKMAN_CLK_GATE_CTRL2_SPI0_CLK_GATER_POS      2                                                                                           /**< SPI0_CLK_GATER Position         */ 
#define MXC_F_CLKMAN_CLK_GATE_CTRL2_SPI0_CLK_GATER          ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL2_SPI0_CLK_GATER_POS))                /**< SPI0_CLK_GATER Mask             */            
#define MXC_F_CLKMAN_CLK_GATE_CTRL2_SPI1_CLK_GATER_POS      4                                                                                           /**< SPI1_CLK_GATER Position         */ 
#define MXC_F_CLKMAN_CLK_GATE_CTRL2_SPI1_CLK_GATER          ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL2_SPI1_CLK_GATER_POS))                /**< SPI1_CLK_GATER Mask             */            
#define MXC_F_CLKMAN_CLK_GATE_CTRL2_SPI2_CLK_GATER_POS      6                                                                                           /**< SPI2_CLK_GATER Position         */ 
#define MXC_F_CLKMAN_CLK_GATE_CTRL2_SPI2_CLK_GATER          ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL2_SPI2_CLK_GATER_POS))                /**< SPI2_CLK_GATER Mask             */           
#define MXC_F_CLKMAN_CLK_GATE_CTRL2_SPI_BRIDGE_CLK_GATER_POS  8                                                                                         /**< SPI_BRIDGE_CLK_GATER Position   */ 
#define MXC_F_CLKMAN_CLK_GATE_CTRL2_SPI_BRIDGE_CLK_GATER    ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL2_SPI_BRIDGE_CLK_GATER_POS))          /**< SPI_BRIDGE_CLK_GATER Mask       */          
#define MXC_F_CLKMAN_CLK_GATE_CTRL2_OWM_CLK_GATER_POS       10                                                                                          /**< OWM_CLK_GATER Position          */ 
#define MXC_F_CLKMAN_CLK_GATE_CTRL2_OWM_CLK_GATER           ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL2_OWM_CLK_GATER_POS))                 /**< OWM_CLK_GATER Mask              */         
#define MXC_F_CLKMAN_CLK_GATE_CTRL2_ADC_CLK_GATER_POS       12                                                                                          /**< ADC_CLK_GATER Position          */ 
#define MXC_F_CLKMAN_CLK_GATE_CTRL2_ADC_CLK_GATER           ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL2_ADC_CLK_GATER_POS))                 /**< ADC_CLK_GATER  Mask             */            
#define MXC_F_CLKMAN_CLK_GATE_CTRL2_SPIS_CLK_GATER_POS      14                                                                                          /**< SPIS_CLK_GATER Position         */ 
#define MXC_F_CLKMAN_CLK_GATE_CTRL2_SPIS_CLK_GATER          ((uint32_t)(0x00000003UL << MXC_F_CLKMAN_CLK_GATE_CTRL2_SPIS_CLK_GATER_POS))                /**< SPIS_CLK_GATER Mask             */            
/**@}*/
/**
 * @ingroup     clkman_clk_config
 * @defgroup    clkman_crypto_stability_count CRYPTO_STABILITY_COUNT Value Settings and Shifted Value Settings
 * @brief       Crypto Clock Stability Count Setting Values and Shifted Values  
 */
#define MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_8_CLOCKS           ((uint32_t)(0x00000000UL))      /**< CRYPTO_STABILITY_COUNT Value = 2<SUP>8</SUP> */ 
#define MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_9_CLOCKS           ((uint32_t)(0x00000001UL))      /**< CRYPTO_STABILITY_COUNT Value = 2<SUP>9</SUP> */ 
#define MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_10_CLOCKS          ((uint32_t)(0x00000002UL))      /**< CRYPTO_STABILITY_COUNT Value = 2<SUP>10</SUP> */
#define MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_11_CLOCKS          ((uint32_t)(0x00000003UL))      /**< CRYPTO_STABILITY_COUNT Value = 2<SUP>11</SUP> */
#define MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_12_CLOCKS          ((uint32_t)(0x00000004UL))      /**< CRYPTO_STABILITY_COUNT Value = 2<SUP>12</SUP> */
#define MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_13_CLOCKS          ((uint32_t)(0x00000005UL))      /**< CRYPTO_STABILITY_COUNT Value = 2<SUP>13</SUP> */
#define MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_14_CLOCKS          ((uint32_t)(0x00000006UL))      /**< CRYPTO_STABILITY_COUNT Value = 2<SUP>14</SUP> */
#define MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_15_CLOCKS          ((uint32_t)(0x00000007UL))      /**< CRYPTO_STABILITY_COUNT Value = 2<SUP>15</SUP> */
#define MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_16_CLOCKS          ((uint32_t)(0x00000008UL))      /**< CRYPTO_STABILITY_COUNT Value = 2<SUP>16</SUP> */
#define MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_17_CLOCKS          ((uint32_t)(0x00000009UL))      /**< CRYPTO_STABILITY_COUNT Value = 2<SUP>17</SUP> */
#define MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_18_CLOCKS          ((uint32_t)(0x0000000AUL))      /**< CRYPTO_STABILITY_COUNT Value = 2<SUP>18</SUP> */
#define MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_19_CLOCKS          ((uint32_t)(0x0000000BUL))      /**< CRYPTO_STABILITY_COUNT Value = 2<SUP>19</SUP> */
#define MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_20_CLOCKS          ((uint32_t)(0x0000000CUL))      /**< CRYPTO_STABILITY_COUNT Value = 2<SUP>20</SUP> */
#define MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_21_CLOCKS          ((uint32_t)(0x0000000DUL))      /**< CRYPTO_STABILITY_COUNT Value = 2<SUP>21</SUP> */
#define MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_22_CLOCKS          ((uint32_t)(0x0000000EUL))      /**< CRYPTO_STABILITY_COUNT Value = 2<SUP>22</SUP> */
#define MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_23_CLOCKS          ((uint32_t)(0x0000000FUL))      /**< CRYPTO_STABILITY_COUNT Value = 2<SUP>23</SUP> */

#define MXC_S_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_8_CLOCKS           ((uint32_t)(MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_8_CLOCKS    << MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_POS))  /**< CRYPTO_STABILITY_COUNT Shifted Value for 2<SUP>8</SUP>   */ 
#define MXC_S_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_9_CLOCKS           ((uint32_t)(MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_9_CLOCKS    << MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_POS))  /**< CRYPTO_STABILITY_COUNT Shifted Value for 2<SUP>9</SUP>   */ 
#define MXC_S_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_10_CLOCKS          ((uint32_t)(MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_10_CLOCKS   << MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_POS))  /**< CRYPTO_STABILITY_COUNT Shifted Value for 2<SUP>10</SUP>  */
#define MXC_S_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_11_CLOCKS          ((uint32_t)(MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_11_CLOCKS   << MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_POS))  /**< CRYPTO_STABILITY_COUNT Shifted Value for 2<SUP>11</SUP>  */
#define MXC_S_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_12_CLOCKS          ((uint32_t)(MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_12_CLOCKS   << MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_POS))  /**< CRYPTO_STABILITY_COUNT Shifted Value for 2<SUP>12</SUP>  */
#define MXC_S_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_13_CLOCKS          ((uint32_t)(MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_13_CLOCKS   << MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_POS))  /**< CRYPTO_STABILITY_COUNT Shifted Value for 2<SUP>13</SUP>  */
#define MXC_S_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_14_CLOCKS          ((uint32_t)(MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_14_CLOCKS   << MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_POS))  /**< CRYPTO_STABILITY_COUNT Shifted Value for 2<SUP>14</SUP>  */
#define MXC_S_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_15_CLOCKS          ((uint32_t)(MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_15_CLOCKS   << MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_POS))  /**< CRYPTO_STABILITY_COUNT Shifted Value for 2<SUP>15</SUP>  */
#define MXC_S_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_16_CLOCKS          ((uint32_t)(MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_16_CLOCKS   << MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_POS))  /**< CRYPTO_STABILITY_COUNT Shifted Value for 2<SUP>16</SUP>  */
#define MXC_S_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_17_CLOCKS          ((uint32_t)(MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_17_CLOCKS   << MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_POS))  /**< CRYPTO_STABILITY_COUNT Shifted Value for 2<SUP>17</SUP>  */
#define MXC_S_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_18_CLOCKS          ((uint32_t)(MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_18_CLOCKS   << MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_POS))  /**< CRYPTO_STABILITY_COUNT Shifted Value for 2<SUP>18</SUP>  */
#define MXC_S_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_19_CLOCKS          ((uint32_t)(MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_19_CLOCKS   << MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_POS))  /**< CRYPTO_STABILITY_COUNT Shifted Value for 2<SUP>19</SUP>  */
#define MXC_S_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_20_CLOCKS          ((uint32_t)(MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_20_CLOCKS   << MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_POS))  /**< CRYPTO_STABILITY_COUNT Shifted Value for 2<SUP>20</SUP>  */
#define MXC_S_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_21_CLOCKS          ((uint32_t)(MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_21_CLOCKS   << MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_POS))  /**< CRYPTO_STABILITY_COUNT Shifted Value for 2<SUP>21</SUP>  */
#define MXC_S_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_22_CLOCKS          ((uint32_t)(MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_22_CLOCKS   << MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_POS))  /**< CRYPTO_STABILITY_COUNT Shifted Value for 2<SUP>22</SUP>  */
#define MXC_S_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_23_CLOCKS          ((uint32_t)(MXC_V_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_2_EXP_23_CLOCKS   << MXC_F_CLKMAN_CLK_CONFIG_CRYPTO_STABILITY_COUNT_POS))  /**< CRYPTO_STABILITY_COUNT Shifted Value for 2<SUP>23</SUP>  */

/**@} clkman_crypto_stability_count */

/**
 * @ingroup clkman_clk_ctrl
 * @defgroup clkman_sysclock_select System Clock Select Values
 * @brief System Clock Selection Values and Shifted Values for selecting the system clock source 
 * @{
 */
#define MXC_V_CLKMAN_CLK_CTRL_SYSTEM_SOURCE_SELECT_96MHZ_RO_DIV_2               ((uint32_t)(0x00000000UL))                                                                                                  /**< Value Mask: SYSTEM_SOURCE_SELECT_96MHZ_RO_DIV_2                */
#define MXC_V_CLKMAN_CLK_CTRL_SYSTEM_SOURCE_SELECT_96MHZ_RO                     ((uint32_t)(0x00000001UL))                                                                                                  /**< Value Mask: SYSTEM_SOURCE_SELECT_96MHZ_RO                      */
#define MXC_S_CLKMAN_CLK_CTRL_SYSTEM_SOURCE_SELECT_96MHZ_RO_DIV_2               ((uint32_t)(MXC_V_CLKMAN_CLK_CTRL_SYSTEM_SOURCE_SELECT_96MHZ_RO_DIV_2  << MXC_F_CLKMAN_CLK_CTRL_SYSTEM_SOURCE_SELECT_POS))  /**< Value Shifted: SYSTEM_SOURCE_SELECT_96MHZ_RO_DIV_2             */
#define MXC_S_CLKMAN_CLK_CTRL_SYSTEM_SOURCE_SELECT_96MHZ_RO                     ((uint32_t)(MXC_V_CLKMAN_CLK_CTRL_SYSTEM_SOURCE_SELECT_96MHZ_RO        << MXC_F_CLKMAN_CLK_CTRL_SYSTEM_SOURCE_SELECT_POS))  /**< Value Shifted: SYSTEM_SOURCE_SELECT_96MHZ_RO                   */
/**@} end of clkman_sysclock_select group */
///@cond

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
///@endcond

#ifdef __cplusplus
}
#endif

#endif   /* _MXC_CLKMAN_REGS_H_ */

