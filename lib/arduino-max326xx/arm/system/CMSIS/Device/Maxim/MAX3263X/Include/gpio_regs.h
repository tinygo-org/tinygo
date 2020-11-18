/**
 * @file
 * @brief   Registers, Bit Masks and Bit Positions for the GPIO Peripheral Module.
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
 * $Date: 2017-02-14 18:13:08 -0600 (Tue, 14 Feb 2017) $
 * $Revision: 26423 $
 *
*************************************************************************** */

/* Define to prevent redundant inclusion */
#ifndef _MXC_GPIO_REGS_H_
#define _MXC_GPIO_REGS_H_

/* **** Includes **** */
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

///@cond
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
///@endcond

/* **** Definitions **** */

/**
 * @ingroup     gpio
 * @defgroup    gpio_registers Registers
 * @brief       Registers, Bit Masks and Bit Positions for the GPIO Peripheral Module.
 * @{
 */

/*
   Typedefed structure(s) for module registers (per instance or section) with direct 32-bit
   access to each register in module.
*/

/**
 * Structure type to access the GPIO Registers
 */
typedef struct {
    __IO uint32_t rst_mode[16];                          /**< <b><tt>0x0000-0x003C</tt></b> GPIO_RST_MODE_P[0..15] Registers - Power-On Reset Output Drive Mode            */
    __IO uint32_t free[16];                              /**< <b><tt>0x0040-0x007C</tt></b> GPIO_FREE_P[0..15] Registers - Free for GPIO Operation Flags                                   */
    __IO uint32_t out_mode[16];                          /**< <b><tt>0x0080-0x00BC</tt></b> GPIO_OUT_MODE_P[0..15] Registers - Output Drive Mode                                               */
    __IO uint32_t out_val[16];                           /**< <b><tt>0x00C0-0x00FC</tt></b> GPIO_OUT_VAL_P[0..15] Registers - GPIO Output Value                                               */
    __IO uint32_t func_sel[16];                          /**< <b><tt>0x0100-0x013C</tt></b> GPIO_FUNC_SEL_P[0..15] Registers - GPIO Function Select                                            */
    __IO uint32_t in_mode[16];                           /**< <b><tt>0x0140-0x017C</tt></b> GPIO_IN_MODE_P[0..15] Registers - GPIO Input Monitoring Mode                                      */
    __IO uint32_t in_val[16];                            /**< <b><tt>0x0180-0x01BC</tt></b> GPIO_IN_VAL_P[0..15] Registers - GPIO Input Value                                                */
    __IO uint32_t int_mode[16];                          /**< <b><tt>0x01C0-0x01FC</tt></b> GPIO_INT_MODE_P[0..15] Registers - Interrupt Detection Mode                                        */
    __IO uint32_t intfl[16];                             /**< <b><tt>0x0200-0x023C</tt></b> GPIO_INTFL_P[0..15] Registers - Interrupt Flags                                                 */
    __IO uint32_t inten[16];                             /**< <b><tt>0x0240-0x027C</tt></b> GPIO_INTEN_P[0..15] Registers - Interrupt Enables                                               */
} mxc_gpio_regs_t;
/**@} end of gpio_registers group */

/*
   Register offsets for module GPIO.
*/
/**
 * @ingroup gpio_registers
 * @defgroup GPIO_Register_Offsets Register Offsets
 * @brief GPIO Register Offsets from the GPIO Base Address.
 * @{
 */
/**
 * @ingroup  GPIO_Register_Offsets
 * @defgroup gpio_rst_mode_offsets Registers GPIO_RST_MODE_P[0..15] Offsets
 * @brief   GPIO_RST_MODE_P[0..15] Register Offsets from the GPIO Base Peripheral Address. 
 * @{
 */
#define MXC_R_GPIO_OFFS_RST_MODE_P0                         ((uint32_t)0x00000000UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0000</tt></b>    */
#define MXC_R_GPIO_OFFS_RST_MODE_P1                         ((uint32_t)0x00000004UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0004</tt></b>    */
#define MXC_R_GPIO_OFFS_RST_MODE_P2                         ((uint32_t)0x00000008UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0008</tt></b>    */
#define MXC_R_GPIO_OFFS_RST_MODE_P3                         ((uint32_t)0x0000000CUL)                        /**< Offset from GPIO Base Address: <b><tt>0x000C</tt></b>    */
#define MXC_R_GPIO_OFFS_RST_MODE_P4                         ((uint32_t)0x00000010UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0010</tt></b>    */
#define MXC_R_GPIO_OFFS_RST_MODE_P5                         ((uint32_t)0x00000014UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0014</tt></b>    */
#define MXC_R_GPIO_OFFS_RST_MODE_P6                         ((uint32_t)0x00000018UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0018</tt></b>    */
#define MXC_R_GPIO_OFFS_RST_MODE_P7                         ((uint32_t)0x0000001CUL)                        /**< Offset from GPIO Base Address: <b><tt>0x001C</tt></b>    */
#define MXC_R_GPIO_OFFS_RST_MODE_P8                         ((uint32_t)0x00000020UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0020</tt></b>    */
#define MXC_R_GPIO_OFFS_RST_MODE_P9                         ((uint32_t)0x00000024UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0024</tt></b>    */
#define MXC_R_GPIO_OFFS_RST_MODE_P10                        ((uint32_t)0x00000028UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0028</tt></b>    */
#define MXC_R_GPIO_OFFS_RST_MODE_P11                        ((uint32_t)0x0000002CUL)                        /**< Offset from GPIO Base Address: <b><tt>0x002C</tt></b>    */
#define MXC_R_GPIO_OFFS_RST_MODE_P12                        ((uint32_t)0x00000030UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0030</tt></b>    */
#define MXC_R_GPIO_OFFS_RST_MODE_P13                        ((uint32_t)0x00000034UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0034</tt></b>    */
#define MXC_R_GPIO_OFFS_RST_MODE_P14                        ((uint32_t)0x00000038UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0038</tt></b>    */
#define MXC_R_GPIO_OFFS_RST_MODE_P15                        ((uint32_t)0x0000003CUL)                        /**< Offset from GPIO Base Address: <b><tt>0x003C</tt></b>    */
/**@} end of gpio_rst_mode group */
/**
 * @ingroup  GPIO_Register_Offsets
 * @defgroup gpio_free_offsets Registers GPIO_FREE_P[0..15] Offsets 
 * @brief    GPIO_FREE_P[0..15] Register Offsets from the GPIO Base Peripheral Address. 
 * @{
 */    
#define MXC_R_GPIO_OFFS_FREE_P0                             ((uint32_t)0x00000040UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0040</tt></b> */
#define MXC_R_GPIO_OFFS_FREE_P1                             ((uint32_t)0x00000044UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0044</tt></b> */
#define MXC_R_GPIO_OFFS_FREE_P2                             ((uint32_t)0x00000048UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0048</tt></b> */
#define MXC_R_GPIO_OFFS_FREE_P3                             ((uint32_t)0x0000004CUL)                        /**< Offset from GPIO Base Address: <b><tt>0x004C</tt></b> */
#define MXC_R_GPIO_OFFS_FREE_P4                             ((uint32_t)0x00000050UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0050</tt></b> */
#define MXC_R_GPIO_OFFS_FREE_P5                             ((uint32_t)0x00000054UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0054</tt></b> */
#define MXC_R_GPIO_OFFS_FREE_P6                             ((uint32_t)0x00000058UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0058</tt></b> */
#define MXC_R_GPIO_OFFS_FREE_P7                             ((uint32_t)0x0000005CUL)                        /**< Offset from GPIO Base Address: <b><tt>0x005C</tt></b> */
#define MXC_R_GPIO_OFFS_FREE_P8                             ((uint32_t)0x00000060UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0060</tt></b> */
#define MXC_R_GPIO_OFFS_FREE_P9                             ((uint32_t)0x00000064UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0064</tt></b> */
#define MXC_R_GPIO_OFFS_FREE_P10                            ((uint32_t)0x00000068UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0068</tt></b> */
#define MXC_R_GPIO_OFFS_FREE_P11                            ((uint32_t)0x0000006CUL)                        /**< Offset from GPIO Base Address: <b><tt>0x006C</tt></b> */
#define MXC_R_GPIO_OFFS_FREE_P12                            ((uint32_t)0x00000070UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0070</tt></b> */
#define MXC_R_GPIO_OFFS_FREE_P13                            ((uint32_t)0x00000074UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0074</tt></b> */
#define MXC_R_GPIO_OFFS_FREE_P14                            ((uint32_t)0x00000078UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0078</tt></b> */
#define MXC_R_GPIO_OFFS_FREE_P15                            ((uint32_t)0x0000007CUL)                        /**< Offset from GPIO Base Address: <b><tt>0x007C</tt></b> */
/**@} end of gpio_free group */
/**
 * @ingroup  GPIO_Register_Offsets
 * @defgroup gpio_out_mode_offsets GPIO_OUT_MODE_P[0..15] Registers 
 * @brief    GPIO_OUT_MODE_P[0..15] Register Offsets from the GPIO Base Peripheral Address. 
 * @{
 */   
#define MXC_R_GPIO_OFFS_OUT_MODE_P0                         ((uint32_t)0x00000080UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0080</tt></b> */
#define MXC_R_GPIO_OFFS_OUT_MODE_P1                         ((uint32_t)0x00000084UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0084</tt></b> */
#define MXC_R_GPIO_OFFS_OUT_MODE_P2                         ((uint32_t)0x00000088UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0088</tt></b> */
#define MXC_R_GPIO_OFFS_OUT_MODE_P3                         ((uint32_t)0x0000008CUL)                        /**< Offset from GPIO Base Address: <b><tt>0x008C</tt></b> */
#define MXC_R_GPIO_OFFS_OUT_MODE_P4                         ((uint32_t)0x00000090UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0090</tt></b> */
#define MXC_R_GPIO_OFFS_OUT_MODE_P5                         ((uint32_t)0x00000094UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0094</tt></b> */
#define MXC_R_GPIO_OFFS_OUT_MODE_P6                         ((uint32_t)0x00000098UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0098</tt></b> */
#define MXC_R_GPIO_OFFS_OUT_MODE_P7                         ((uint32_t)0x0000009CUL)                        /**< Offset from GPIO Base Address: <b><tt>0x009C</tt></b> */
#define MXC_R_GPIO_OFFS_OUT_MODE_P8                         ((uint32_t)0x000000A0UL)                        /**< Offset from GPIO Base Address: <b><tt>0x00A0</tt></b> */
#define MXC_R_GPIO_OFFS_OUT_MODE_P9                         ((uint32_t)0x000000A4UL)                        /**< Offset from GPIO Base Address: <b><tt>0x00A4</tt></b> */
#define MXC_R_GPIO_OFFS_OUT_MODE_P10                        ((uint32_t)0x000000A8UL)                        /**< Offset from GPIO Base Address: <b><tt>0x00A8</tt></b>  */
#define MXC_R_GPIO_OFFS_OUT_MODE_P11                        ((uint32_t)0x000000ACUL)                        /**< Offset from GPIO Base Address: <b><tt>0x00AC</tt></b>  */
#define MXC_R_GPIO_OFFS_OUT_MODE_P12                        ((uint32_t)0x000000B0UL)                        /**< Offset from GPIO Base Address: <b><tt>0x00B0</tt></b>  */
#define MXC_R_GPIO_OFFS_OUT_MODE_P13                        ((uint32_t)0x000000B4UL)                        /**< Offset from GPIO Base Address: <b><tt>0x00B4</tt></b>  */
#define MXC_R_GPIO_OFFS_OUT_MODE_P14                        ((uint32_t)0x000000B8UL)                        /**< Offset from GPIO Base Address: <b><tt>0x00B8</tt></b>  */
#define MXC_R_GPIO_OFFS_OUT_MODE_P15                        ((uint32_t)0x000000BCUL)                        /**< Offset from GPIO Base Address: <b><tt>0x00BC</tt></b>  */
/**@} end of gpio_out_mode group */
/**
 * @ingroup  GPIO_Register_Offsets
 * @defgroup gpio_out_val_offsets GPIO_OUT_VAL_P[0..15] Registers 
 * @brief    GPIO_OUT_VAL_P[0..15] Register Offsets from the GPIO Base Peripheral Address. 
 * @{
 */   
#define MXC_R_GPIO_OFFS_OUT_VAL_P0                          ((uint32_t)0x000000C0UL)                        /**< Offset from GPIO Base Address: <b><tt>0x00C0</tt></b> */
#define MXC_R_GPIO_OFFS_OUT_VAL_P1                          ((uint32_t)0x000000C4UL)                        /**< Offset from GPIO Base Address: <b><tt>0x00C4</tt></b> */
#define MXC_R_GPIO_OFFS_OUT_VAL_P2                          ((uint32_t)0x000000C8UL)                        /**< Offset from GPIO Base Address: <b><tt>0x00C8</tt></b> */
#define MXC_R_GPIO_OFFS_OUT_VAL_P3                          ((uint32_t)0x000000CCUL)                        /**< Offset from GPIO Base Address: <b><tt>0x00CC</tt></b> */
#define MXC_R_GPIO_OFFS_OUT_VAL_P4                          ((uint32_t)0x000000D0UL)                        /**< Offset from GPIO Base Address: <b><tt>0x00D0</tt></b> */
#define MXC_R_GPIO_OFFS_OUT_VAL_P5                          ((uint32_t)0x000000D4UL)                        /**< Offset from GPIO Base Address: <b><tt>0x00D4</tt></b> */
#define MXC_R_GPIO_OFFS_OUT_VAL_P6                          ((uint32_t)0x000000D8UL)                        /**< Offset from GPIO Base Address: <b><tt>0x00D8</tt></b> */
#define MXC_R_GPIO_OFFS_OUT_VAL_P7                          ((uint32_t)0x000000DCUL)                        /**< Offset from GPIO Base Address: <b><tt>0x00DC</tt></b> */
#define MXC_R_GPIO_OFFS_OUT_VAL_P8                          ((uint32_t)0x000000E0UL)                        /**< Offset from GPIO Base Address: <b><tt>0x00E0</tt></b> */
#define MXC_R_GPIO_OFFS_OUT_VAL_P9                          ((uint32_t)0x000000E4UL)                        /**< Offset from GPIO Base Address: <b><tt>0x00E4</tt></b> */
#define MXC_R_GPIO_OFFS_OUT_VAL_P10                         ((uint32_t)0x000000E8UL)                        /**< Offset from GPIO Base Address: <b><tt>0x00E8</tt></b>  */
#define MXC_R_GPIO_OFFS_OUT_VAL_P11                         ((uint32_t)0x000000ECUL)                        /**< Offset from GPIO Base Address: <b><tt>0x00EC</tt></b>  */
#define MXC_R_GPIO_OFFS_OUT_VAL_P12                         ((uint32_t)0x000000F0UL)                        /**< Offset from GPIO Base Address: <b><tt>0x00F0</tt></b>  */
#define MXC_R_GPIO_OFFS_OUT_VAL_P13                         ((uint32_t)0x000000F4UL)                        /**< Offset from GPIO Base Address: <b><tt>0x00F4</tt></b>  */
#define MXC_R_GPIO_OFFS_OUT_VAL_P14                         ((uint32_t)0x000000F8UL)                        /**< Offset from GPIO Base Address: <b><tt>0x00F8</tt></b>  */
#define MXC_R_GPIO_OFFS_OUT_VAL_P15                         ((uint32_t)0x000000FCUL)                        /**< Offset from GPIO Base Address: <b><tt>0x00FC</tt></b>  */
/**@} end of gpio_out_val group */
/**
 * @ingroup  GPIO_Register_Offsets
 * @defgroup gpio_func_sel_offsets GPIO_FUNC_SEL_P[0..15] Registers 
 * @brief    GPIO_FUNC_SEL_P[0..15] Register Offsets from the GPIO Base Peripheral Address. 
 * @{
 */   
#define MXC_R_GPIO_OFFS_FUNC_SEL_P0                         ((uint32_t)0x00000100UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0100</tt></b> */
#define MXC_R_GPIO_OFFS_FUNC_SEL_P1                         ((uint32_t)0x00000104UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0104</tt></b> */
#define MXC_R_GPIO_OFFS_FUNC_SEL_P2                         ((uint32_t)0x00000108UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0108</tt></b> */
#define MXC_R_GPIO_OFFS_FUNC_SEL_P3                         ((uint32_t)0x0000010CUL)                        /**< Offset from GPIO Base Address: <b><tt>0x010C</tt></b> */
#define MXC_R_GPIO_OFFS_FUNC_SEL_P4                         ((uint32_t)0x00000110UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0110</tt></b> */
#define MXC_R_GPIO_OFFS_FUNC_SEL_P5                         ((uint32_t)0x00000114UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0114</tt></b> */
#define MXC_R_GPIO_OFFS_FUNC_SEL_P6                         ((uint32_t)0x00000118UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0118</tt></b> */
#define MXC_R_GPIO_OFFS_FUNC_SEL_P7                         ((uint32_t)0x0000011CUL)                        /**< Offset from GPIO Base Address: <b><tt>0x011C</tt></b> */
#define MXC_R_GPIO_OFFS_FUNC_SEL_P8                         ((uint32_t)0x00000120UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0120</tt></b> */
#define MXC_R_GPIO_OFFS_FUNC_SEL_P9                         ((uint32_t)0x00000124UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0124</tt></b> */
#define MXC_R_GPIO_OFFS_FUNC_SEL_P10                        ((uint32_t)0x00000128UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0128</tt></b>  */
#define MXC_R_GPIO_OFFS_FUNC_SEL_P11                        ((uint32_t)0x0000012CUL)                        /**< Offset from GPIO Base Address: <b><tt>0x012C</tt></b>  */
#define MXC_R_GPIO_OFFS_FUNC_SEL_P12                        ((uint32_t)0x00000130UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0130</tt></b>  */
#define MXC_R_GPIO_OFFS_FUNC_SEL_P13                        ((uint32_t)0x00000134UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0134</tt></b>  */
#define MXC_R_GPIO_OFFS_FUNC_SEL_P14                        ((uint32_t)0x00000138UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0138</tt></b>  */
#define MXC_R_GPIO_OFFS_FUNC_SEL_P15                        ((uint32_t)0x0000013CUL)                        /**< Offset from GPIO Base Address: <b><tt>0x013C</tt></b>  */
/**@} end of gpio_func_sel */
/**
 * @ingroup  GPIO_Register_Offsets
 * @defgroup gpio_in_mode_offsets GPIO_IN_MODE_P[0..15] Registers 
 * @brief    GPIO_IN_MODE_P[0..15] Register Offsets from the GPIO Base Peripheral Address. 
 * @{
 */ 
#define MXC_R_GPIO_OFFS_IN_MODE_P0                          ((uint32_t)0x00000140UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0140</tt></b> */
#define MXC_R_GPIO_OFFS_IN_MODE_P1                          ((uint32_t)0x00000144UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0144</tt></b> */
#define MXC_R_GPIO_OFFS_IN_MODE_P2                          ((uint32_t)0x00000148UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0148</tt></b> */
#define MXC_R_GPIO_OFFS_IN_MODE_P3                          ((uint32_t)0x0000014CUL)                        /**< Offset from GPIO Base Address: <b><tt>0x014C</tt></b> */
#define MXC_R_GPIO_OFFS_IN_MODE_P4                          ((uint32_t)0x00000150UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0150</tt></b> */
#define MXC_R_GPIO_OFFS_IN_MODE_P5                          ((uint32_t)0x00000154UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0154</tt></b> */
#define MXC_R_GPIO_OFFS_IN_MODE_P6                          ((uint32_t)0x00000158UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0158</tt></b> */
#define MXC_R_GPIO_OFFS_IN_MODE_P7                          ((uint32_t)0x0000015CUL)                        /**< Offset from GPIO Base Address: <b><tt>0x015C</tt></b> */
#define MXC_R_GPIO_OFFS_IN_MODE_P8                          ((uint32_t)0x00000160UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0160</tt></b> */
#define MXC_R_GPIO_OFFS_IN_MODE_P9                          ((uint32_t)0x00000164UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0164</tt></b> */
#define MXC_R_GPIO_OFFS_IN_MODE_P10                         ((uint32_t)0x00000168UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0168</tt></b>  */
#define MXC_R_GPIO_OFFS_IN_MODE_P11                         ((uint32_t)0x0000016CUL)                        /**< Offset from GPIO Base Address: <b><tt>0x016C</tt></b>  */
#define MXC_R_GPIO_OFFS_IN_MODE_P12                         ((uint32_t)0x00000170UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0170</tt></b>  */
#define MXC_R_GPIO_OFFS_IN_MODE_P13                         ((uint32_t)0x00000174UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0174</tt></b>  */
#define MXC_R_GPIO_OFFS_IN_MODE_P14                         ((uint32_t)0x00000178UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0178</tt></b>  */
#define MXC_R_GPIO_OFFS_IN_MODE_P15                         ((uint32_t)0x0000017CUL)                        /**< Offset from GPIO Base Address: <b><tt>0x017C</tt></b>  */
/**@} end of gpio_in_mode group */
/**
 * @ingroup  GPIO_Register_Offsets
 * @defgroup gpio_in_val_offsets GPIO_IN_VAL_P[0..15] Registers 
 * @brief    GPIO_IN_VAL_P[0..15] Register Offsets from the GPIO Base Peripheral Address. 
 * @{
 */ 
#define MXC_R_GPIO_OFFS_IN_VAL_P0                           ((uint32_t)0x00000180UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0180</tt></b> */
#define MXC_R_GPIO_OFFS_IN_VAL_P1                           ((uint32_t)0x00000184UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0184</tt></b> */
#define MXC_R_GPIO_OFFS_IN_VAL_P2                           ((uint32_t)0x00000188UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0188</tt></b> */
#define MXC_R_GPIO_OFFS_IN_VAL_P3                           ((uint32_t)0x0000018CUL)                        /**< Offset from GPIO Base Address: <b><tt>0x018C</tt></b> */
#define MXC_R_GPIO_OFFS_IN_VAL_P4                           ((uint32_t)0x00000190UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0190</tt></b> */
#define MXC_R_GPIO_OFFS_IN_VAL_P5                           ((uint32_t)0x00000194UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0194</tt></b> */
#define MXC_R_GPIO_OFFS_IN_VAL_P6                           ((uint32_t)0x00000198UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0198</tt></b> */
#define MXC_R_GPIO_OFFS_IN_VAL_P7                           ((uint32_t)0x0000019CUL)                        /**< Offset from GPIO Base Address: <b><tt>0x019C</tt></b> */
#define MXC_R_GPIO_OFFS_IN_VAL_P8                           ((uint32_t)0x000001A0UL)                        /**< Offset from GPIO Base Address: <b><tt>0x01A0</tt></b> */
#define MXC_R_GPIO_OFFS_IN_VAL_P9                           ((uint32_t)0x000001A4UL)                        /**< Offset from GPIO Base Address: <b><tt>0x01A4</tt></b> */
#define MXC_R_GPIO_OFFS_IN_VAL_P10                          ((uint32_t)0x000001A8UL)                        /**< Offset from GPIO Base Address: <b><tt>0x01A8</tt></b>  */
#define MXC_R_GPIO_OFFS_IN_VAL_P11                          ((uint32_t)0x000001ACUL)                        /**< Offset from GPIO Base Address: <b><tt>0x01AC</tt></b>  */
#define MXC_R_GPIO_OFFS_IN_VAL_P12                          ((uint32_t)0x000001B0UL)                        /**< Offset from GPIO Base Address: <b><tt>0x01B0</tt></b>  */
#define MXC_R_GPIO_OFFS_IN_VAL_P13                          ((uint32_t)0x000001B4UL)                        /**< Offset from GPIO Base Address: <b><tt>0x01B4</tt></b>  */
#define MXC_R_GPIO_OFFS_IN_VAL_P14                          ((uint32_t)0x000001B8UL)                        /**< Offset from GPIO Base Address: <b><tt>0x01B8</tt></b>  */
#define MXC_R_GPIO_OFFS_IN_VAL_P15                          ((uint32_t)0x000001BCUL)                        /**< Offset from GPIO Base Address: <b><tt>0x01BC</tt></b>  */
/**@} end of gpio_in_val group */
/**
 * @ingroup  GPIO_Register_Offsets
 * @defgroup gpio_int_mode_offsets GPIO_INT_MODE_P[0..15] Registers 
 * @brief    GPIO_INT_MODE_P[0..15] Register Offsets from the GPIO Base Peripheral Address. 
 * @{
 */ 
#define MXC_R_GPIO_OFFS_INT_MODE_P0                         ((uint32_t)0x000001C0UL)                        /**< Offset from GPIO Base Address: <b><tt>0x01C0</tt></b> */
#define MXC_R_GPIO_OFFS_INT_MODE_P1                         ((uint32_t)0x000001C4UL)                        /**< Offset from GPIO Base Address: <b><tt>0x01C4</tt></b> */
#define MXC_R_GPIO_OFFS_INT_MODE_P2                         ((uint32_t)0x000001C8UL)                        /**< Offset from GPIO Base Address: <b><tt>0x01C8</tt></b> */
#define MXC_R_GPIO_OFFS_INT_MODE_P3                         ((uint32_t)0x000001CCUL)                        /**< Offset from GPIO Base Address: <b><tt>0x01CC</tt></b> */
#define MXC_R_GPIO_OFFS_INT_MODE_P4                         ((uint32_t)0x000001D0UL)                        /**< Offset from GPIO Base Address: <b><tt>0x01D0</tt></b> */
#define MXC_R_GPIO_OFFS_INT_MODE_P5                         ((uint32_t)0x000001D4UL)                        /**< Offset from GPIO Base Address: <b><tt>0x01D4</tt></b> */
#define MXC_R_GPIO_OFFS_INT_MODE_P6                         ((uint32_t)0x000001D8UL)                        /**< Offset from GPIO Base Address: <b><tt>0x01D8</tt></b> */
#define MXC_R_GPIO_OFFS_INT_MODE_P7                         ((uint32_t)0x000001DCUL)                        /**< Offset from GPIO Base Address: <b><tt>0x01DC</tt></b> */
#define MXC_R_GPIO_OFFS_INT_MODE_P8                         ((uint32_t)0x000001E0UL)                        /**< Offset from GPIO Base Address: <b><tt>0x01E0</tt></b> */
#define MXC_R_GPIO_OFFS_INT_MODE_P9                         ((uint32_t)0x000001E4UL)                        /**< Offset from GPIO Base Address: <b><tt>0x01E4</tt></b> */
#define MXC_R_GPIO_OFFS_INT_MODE_P10                        ((uint32_t)0x000001E8UL)                        /**< Offset from GPIO Base Address: <b><tt>0x01E8</tt></b>  */
#define MXC_R_GPIO_OFFS_INT_MODE_P11                        ((uint32_t)0x000001ECUL)                        /**< Offset from GPIO Base Address: <b><tt>0x01EC</tt></b>  */
#define MXC_R_GPIO_OFFS_INT_MODE_P12                        ((uint32_t)0x000001F0UL)                        /**< Offset from GPIO Base Address: <b><tt>0x01F0</tt></b>  */
#define MXC_R_GPIO_OFFS_INT_MODE_P13                        ((uint32_t)0x000001F4UL)                        /**< Offset from GPIO Base Address: <b><tt>0x01F4</tt></b>  */
#define MXC_R_GPIO_OFFS_INT_MODE_P14                        ((uint32_t)0x000001F8UL)                        /**< Offset from GPIO Base Address: <b><tt>0x01F8</tt></b>  */
#define MXC_R_GPIO_OFFS_INT_MODE_P15                        ((uint32_t)0x000001FCUL)                        /**< Offset from GPIO Base Address: <b><tt>0x01FC</tt></b>  */
/**@} end of gpio_int_mode group */
/**
 * @ingroup  GPIO_Register_Offsets
 * @defgroup gpio_int_flag_offsets GPIO_INTFL_P[0..15] Registers 
 * @brief    GPIO_INTFL_P[0..15] Register Offsets from the GPIO Base Peripheral Address. 
 * @{
 */ 
#define MXC_R_GPIO_OFFS_INTFL_P0                            ((uint32_t)0x00000200UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0200</tt></b> */
#define MXC_R_GPIO_OFFS_INTFL_P1                            ((uint32_t)0x00000204UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0204</tt></b> */
#define MXC_R_GPIO_OFFS_INTFL_P2                            ((uint32_t)0x00000208UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0208</tt></b> */
#define MXC_R_GPIO_OFFS_INTFL_P3                            ((uint32_t)0x0000020CUL)                        /**< Offset from GPIO Base Address: <b><tt>0x020C</tt></b> */
#define MXC_R_GPIO_OFFS_INTFL_P4                            ((uint32_t)0x00000210UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0210</tt></b> */
#define MXC_R_GPIO_OFFS_INTFL_P5                            ((uint32_t)0x00000214UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0214</tt></b> */
#define MXC_R_GPIO_OFFS_INTFL_P6                            ((uint32_t)0x00000218UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0218</tt></b> */
#define MXC_R_GPIO_OFFS_INTFL_P7                            ((uint32_t)0x0000021CUL)                        /**< Offset from GPIO Base Address: <b><tt>0x021C</tt></b> */
#define MXC_R_GPIO_OFFS_INTFL_P8                            ((uint32_t)0x00000220UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0220</tt></b> */
#define MXC_R_GPIO_OFFS_INTFL_P9                            ((uint32_t)0x00000224UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0224</tt></b> */
#define MXC_R_GPIO_OFFS_INTFL_P10                           ((uint32_t)0x00000228UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0228</tt></b>  */
#define MXC_R_GPIO_OFFS_INTFL_P11                           ((uint32_t)0x0000022CUL)                        /**< Offset from GPIO Base Address: <b><tt>0x022C</tt></b>  */
#define MXC_R_GPIO_OFFS_INTFL_P12                           ((uint32_t)0x00000230UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0230</tt></b>  */
#define MXC_R_GPIO_OFFS_INTFL_P13                           ((uint32_t)0x00000234UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0234</tt></b>  */
#define MXC_R_GPIO_OFFS_INTFL_P14                           ((uint32_t)0x00000238UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0238</tt></b>  */
#define MXC_R_GPIO_OFFS_INTFL_P15                           ((uint32_t)0x0000023CUL)                        /**< Offset from GPIO Base Address: <b><tt>0x023C</tt></b>  */
/**@} end of gpio_int_flag group */
/**
 * @ingroup  GPIO_Register_Offsets
 * @defgroup gpio_int_enable_offsets GPIO_INTEN_P[0..15] Registers 
 * @brief    GPIO_INTEN_P[0..15] Register Offsets from the GPIO Base Peripheral Address. 
 * @{
 */ 
#define MXC_R_GPIO_OFFS_INTEN_P0                            ((uint32_t)0x00000240UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0240</tt></b> */
#define MXC_R_GPIO_OFFS_INTEN_P1                            ((uint32_t)0x00000244UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0244</tt></b> */
#define MXC_R_GPIO_OFFS_INTEN_P2                            ((uint32_t)0x00000248UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0248</tt></b> */
#define MXC_R_GPIO_OFFS_INTEN_P3                            ((uint32_t)0x0000024CUL)                        /**< Offset from GPIO Base Address: <b><tt>0x024C</tt></b> */
#define MXC_R_GPIO_OFFS_INTEN_P4                            ((uint32_t)0x00000250UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0250</tt></b> */
#define MXC_R_GPIO_OFFS_INTEN_P5                            ((uint32_t)0x00000254UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0254</tt></b> */
#define MXC_R_GPIO_OFFS_INTEN_P6                            ((uint32_t)0x00000258UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0258</tt></b> */
#define MXC_R_GPIO_OFFS_INTEN_P7                            ((uint32_t)0x0000025CUL)                        /**< Offset from GPIO Base Address: <b><tt>0x025C</tt></b> */
#define MXC_R_GPIO_OFFS_INTEN_P8                            ((uint32_t)0x00000260UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0260</tt></b> */
#define MXC_R_GPIO_OFFS_INTEN_P9                            ((uint32_t)0x00000264UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0264</tt></b> */
#define MXC_R_GPIO_OFFS_INTEN_P10                           ((uint32_t)0x00000268UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0268</tt></b>  */
#define MXC_R_GPIO_OFFS_INTEN_P11                           ((uint32_t)0x0000026CUL)                        /**< Offset from GPIO Base Address: <b><tt>0x026C</tt></b>  */
#define MXC_R_GPIO_OFFS_INTEN_P12                           ((uint32_t)0x00000270UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0270</tt></b>  */
#define MXC_R_GPIO_OFFS_INTEN_P13                           ((uint32_t)0x00000274UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0274</tt></b>  */
#define MXC_R_GPIO_OFFS_INTEN_P14                           ((uint32_t)0x00000278UL)                        /**< Offset from GPIO Base Address: <b><tt>0x0278</tt></b>  */
#define MXC_R_GPIO_OFFS_INTEN_P15                           ((uint32_t)0x0000027CUL)                        /**< Offset from GPIO Base Address: <b><tt>0x027C</tt></b>  */
/**@}*/
/**@} end of GPIO_Register_Offsets */

/*
   Field positions and masks for module GPIO.
*/
/**
 * @ingroup    gpio_registers
 * @defgroup   GPIO_RST_MODE_Register GPIO_RST_MODE
 * @brief      Field Positions and Bit Masks for the GPIO_RST_MODE register. 
 * @{
 */
#define MXC_F_GPIO_RST_MODE_PIN0_POS                        0                                                                 /**< PIN0 Position                  */
#define MXC_F_GPIO_RST_MODE_PIN0                            ((uint32_t)(0x00000007UL << MXC_F_GPIO_RST_MODE_PIN0_POS))        /**< PIN0 Mask                      */
#define MXC_F_GPIO_RST_MODE_PIN1_POS                        4                                                                 /**< PIN1 Position                  */
#define MXC_F_GPIO_RST_MODE_PIN1                            ((uint32_t)(0x00000007UL << MXC_F_GPIO_RST_MODE_PIN1_POS))        /**< PIN1 Mask                      */
#define MXC_F_GPIO_RST_MODE_PIN2_POS                        8                                                                 /**< PIN2 Position                  */
#define MXC_F_GPIO_RST_MODE_PIN2                            ((uint32_t)(0x00000007UL << MXC_F_GPIO_RST_MODE_PIN2_POS))        /**< PIN2 Mask                      */
#define MXC_F_GPIO_RST_MODE_PIN3_POS                        12                                                                /**< PIN3 Position                  */
#define MXC_F_GPIO_RST_MODE_PIN3                            ((uint32_t)(0x00000007UL << MXC_F_GPIO_RST_MODE_PIN3_POS))        /**< PIN3 Mask                      */
#define MXC_F_GPIO_RST_MODE_PIN4_POS                        16                                                                /**< PIN4 Position                  */
#define MXC_F_GPIO_RST_MODE_PIN4                            ((uint32_t)(0x00000007UL << MXC_F_GPIO_RST_MODE_PIN4_POS))        /**< PIN4 Mask                      */
#define MXC_F_GPIO_RST_MODE_PIN5_POS                        20                                                                /**< PIN5 Position                  */
#define MXC_F_GPIO_RST_MODE_PIN5                            ((uint32_t)(0x00000007UL << MXC_F_GPIO_RST_MODE_PIN5_POS))        /**< PIN5 Mask                      */
#define MXC_F_GPIO_RST_MODE_PIN6_POS                        24                                                                /**< PIN6 Position                  */
#define MXC_F_GPIO_RST_MODE_PIN6                            ((uint32_t)(0x00000007UL << MXC_F_GPIO_RST_MODE_PIN6_POS))        /**< PIN6 Mask                      */
#define MXC_F_GPIO_RST_MODE_PIN7_POS                        28                                                                /**< PIN7 Position                  */
#define MXC_F_GPIO_RST_MODE_PIN7                            ((uint32_t)(0x00000007UL << MXC_F_GPIO_RST_MODE_PIN7_POS))        /**< PIN7 Mask                      */
/**@} end of group GPIO_FREE */
/**
 * @ingroup    gpio_registers
 * @defgroup   GPIO_FREE_Register GPIO_FREE
 * @brief      Field Positions and Bit Masks for the GPIO_FREE register. 
 * @{
 */
#define MXC_F_GPIO_FREE_PIN0_POS                            0                                                                 /**< PIN0 Position                  */
#define MXC_F_GPIO_FREE_PIN0                                ((uint32_t)(0x00000001UL << MXC_F_GPIO_FREE_PIN0_POS))            /**< PIN0 Mask                      */
#define MXC_F_GPIO_FREE_PIN1_POS                            1                                                                 /**< PIN1 Position                  */
#define MXC_F_GPIO_FREE_PIN1                                ((uint32_t)(0x00000001UL << MXC_F_GPIO_FREE_PIN1_POS))            /**< PIN1 Mask                      */
#define MXC_F_GPIO_FREE_PIN2_POS                            2                                                                 /**< PIN2 Position                  */
#define MXC_F_GPIO_FREE_PIN2                                ((uint32_t)(0x00000001UL << MXC_F_GPIO_FREE_PIN2_POS))            /**< PIN2 Mask                      */
#define MXC_F_GPIO_FREE_PIN3_POS                            3                                                                 /**< PIN3 Position                  */
#define MXC_F_GPIO_FREE_PIN3                                ((uint32_t)(0x00000001UL << MXC_F_GPIO_FREE_PIN3_POS))            /**< PIN3 Mask                      */
#define MXC_F_GPIO_FREE_PIN4_POS                            4                                                                 /**< PIN4 Position                  */
#define MXC_F_GPIO_FREE_PIN4                                ((uint32_t)(0x00000001UL << MXC_F_GPIO_FREE_PIN4_POS))            /**< PIN4 Mask                      */
#define MXC_F_GPIO_FREE_PIN5_POS                            5                                                                 /**< PIN5 Position                  */
#define MXC_F_GPIO_FREE_PIN5                                ((uint32_t)(0x00000001UL << MXC_F_GPIO_FREE_PIN5_POS))            /**< PIN5 Mask                      */
#define MXC_F_GPIO_FREE_PIN6_POS                            6                                                                 /**< PIN6 Position                  */
#define MXC_F_GPIO_FREE_PIN6                                ((uint32_t)(0x00000001UL << MXC_F_GPIO_FREE_PIN6_POS))            /**< PIN6 Mask                      */
#define MXC_F_GPIO_FREE_PIN7_POS                            7                                                                 /**< PIN7 Position                  */
#define MXC_F_GPIO_FREE_PIN7                                ((uint32_t)(0x00000001UL << MXC_F_GPIO_FREE_PIN7_POS))            /**< PIN7 Mask                      */
/**@} end of group GPIO_FREE */
/**
 * @ingroup    gpio_registers
 * @defgroup   GPIO_OUT_MODE_Register GPIO_OUT_MODE
 * @brief      Field Positions and Bit Masks for the GPIO_OUT_MODE register. 
 * @{
 */
#define MXC_F_GPIO_OUT_MODE_PIN0_POS                        0                                                                 /**< PIN0 Position                  */
#define MXC_F_GPIO_OUT_MODE_PIN0                            ((uint32_t)(0x0000000FUL << MXC_F_GPIO_OUT_MODE_PIN0_POS))        /**< PIN0 Mask                      */    
#define MXC_F_GPIO_OUT_MODE_PIN1_POS                        4                                                                 /**< PIN1 Position                  */
#define MXC_F_GPIO_OUT_MODE_PIN1                            ((uint32_t)(0x0000000FUL << MXC_F_GPIO_OUT_MODE_PIN1_POS))        /**< PIN1 Mask                      */    
#define MXC_F_GPIO_OUT_MODE_PIN2_POS                        8                                                                 /**< PIN2 Position                  */
#define MXC_F_GPIO_OUT_MODE_PIN2                            ((uint32_t)(0x0000000FUL << MXC_F_GPIO_OUT_MODE_PIN2_POS))        /**< PIN2 Mask                      */    
#define MXC_F_GPIO_OUT_MODE_PIN3_POS                        12                                                                /**< PIN3 Position                  */  
#define MXC_F_GPIO_OUT_MODE_PIN3                            ((uint32_t)(0x0000000FUL << MXC_F_GPIO_OUT_MODE_PIN3_POS))        /**< PIN3 Mask                      */    
#define MXC_F_GPIO_OUT_MODE_PIN4_POS                        16                                                                /**< PIN4 Position                  */  
#define MXC_F_GPIO_OUT_MODE_PIN4                            ((uint32_t)(0x0000000FUL << MXC_F_GPIO_OUT_MODE_PIN4_POS))        /**< PIN4 Mask                      */    
#define MXC_F_GPIO_OUT_MODE_PIN5_POS                        20                                                                /**< PIN5 Position                  */  
#define MXC_F_GPIO_OUT_MODE_PIN5                            ((uint32_t)(0x0000000FUL << MXC_F_GPIO_OUT_MODE_PIN5_POS))        /**< PIN5 Mask                      */    
#define MXC_F_GPIO_OUT_MODE_PIN6_POS                        24                                                                /**< PIN6 Position                  */  
#define MXC_F_GPIO_OUT_MODE_PIN6                            ((uint32_t)(0x0000000FUL << MXC_F_GPIO_OUT_MODE_PIN6_POS))        /**< PIN6 Mask                      */    
#define MXC_F_GPIO_OUT_MODE_PIN7_POS                        28                                                                /**< PIN7 Position                  */  
#define MXC_F_GPIO_OUT_MODE_PIN7                            ((uint32_t)(0x0000000FUL << MXC_F_GPIO_OUT_MODE_PIN7_POS))        /**< PIN7 Mask                      */    
/**@} end of group GPIO_OUT_MODE */
/**
 * @ingroup    gpio_registers
 * @defgroup   GPIO_OUT_VAL_Register GPIO_OUT_VAL
 * @brief      Field Positions and Bit Masks for the GPIO_OUT_VAL register. 
 * @{
 */
#define MXC_F_GPIO_OUT_VAL_PIN0_POS                         0                                                                 /**< PIN0 Position                  */
#define MXC_F_GPIO_OUT_VAL_PIN0                             ((uint32_t)(0x00000001UL << MXC_F_GPIO_OUT_VAL_PIN0_POS))         /**< PIN0 Mask                      */  
#define MXC_F_GPIO_OUT_VAL_PIN1_POS                         1                                                                 /**< PIN1 Position                  */
#define MXC_F_GPIO_OUT_VAL_PIN1                             ((uint32_t)(0x00000001UL << MXC_F_GPIO_OUT_VAL_PIN1_POS))         /**< PIN1 Mask                      */  
#define MXC_F_GPIO_OUT_VAL_PIN2_POS                         2                                                                 /**< PIN2 Position                  */
#define MXC_F_GPIO_OUT_VAL_PIN2                             ((uint32_t)(0x00000001UL << MXC_F_GPIO_OUT_VAL_PIN2_POS))         /**< PIN2 Mask                      */  
#define MXC_F_GPIO_OUT_VAL_PIN3_POS                         3                                                                 /**< PIN3 Position                  */
#define MXC_F_GPIO_OUT_VAL_PIN3                             ((uint32_t)(0x00000001UL << MXC_F_GPIO_OUT_VAL_PIN3_POS))         /**< PIN3 Mask                      */  
#define MXC_F_GPIO_OUT_VAL_PIN4_POS                         4                                                                 /**< PIN4 Position                  */
#define MXC_F_GPIO_OUT_VAL_PIN4                             ((uint32_t)(0x00000001UL << MXC_F_GPIO_OUT_VAL_PIN4_POS))         /**< PIN4 Mask                      */  
#define MXC_F_GPIO_OUT_VAL_PIN5_POS                         5                                                                 /**< PIN5 Position                  */
#define MXC_F_GPIO_OUT_VAL_PIN5                             ((uint32_t)(0x00000001UL << MXC_F_GPIO_OUT_VAL_PIN5_POS))         /**< PIN5 Mask                      */  
#define MXC_F_GPIO_OUT_VAL_PIN6_POS                         6                                                                 /**< PIN6 Position                  */
#define MXC_F_GPIO_OUT_VAL_PIN6                             ((uint32_t)(0x00000001UL << MXC_F_GPIO_OUT_VAL_PIN6_POS))         /**< PIN6 Mask                      */  
#define MXC_F_GPIO_OUT_VAL_PIN7_POS                         7                                                                 /**< PIN7 Position                  */
#define MXC_F_GPIO_OUT_VAL_PIN7                             ((uint32_t)(0x00000001UL << MXC_F_GPIO_OUT_VAL_PIN7_POS))         /**< PIN7 Mask                      */  
/**@} end of group GPIO_OUT_VAL */
/**
 * @ingroup    gpio_registers
 * @defgroup   GPIO_FUNC_SEL_Register GPIO_FUNC_SEL
 * @brief      Field Positions and Bit Masks for the GPIO_FUNC_SEL register. 
 * @{
 */
#define MXC_F_GPIO_FUNC_SEL_PIN0_POS                        0                                                                 /**< PIN0 Position                  */
#define MXC_F_GPIO_FUNC_SEL_PIN0                            ((uint32_t)(0x0000000FUL << MXC_F_GPIO_FUNC_SEL_PIN0_POS))        /**< PIN0 Mask                      */    
#define MXC_F_GPIO_FUNC_SEL_PIN1_POS                        4                                                                 /**< PIN1 Position                  */
#define MXC_F_GPIO_FUNC_SEL_PIN1                            ((uint32_t)(0x0000000FUL << MXC_F_GPIO_FUNC_SEL_PIN1_POS))        /**< PIN1 Mask                      */    
#define MXC_F_GPIO_FUNC_SEL_PIN2_POS                        8                                                                 /**< PIN2 Position                  */
#define MXC_F_GPIO_FUNC_SEL_PIN2                            ((uint32_t)(0x0000000FUL << MXC_F_GPIO_FUNC_SEL_PIN2_POS))        /**< PIN2 Mask                      */    
#define MXC_F_GPIO_FUNC_SEL_PIN3_POS                        12                                                                /**< PIN3 Position                  */  
#define MXC_F_GPIO_FUNC_SEL_PIN3                            ((uint32_t)(0x0000000FUL << MXC_F_GPIO_FUNC_SEL_PIN3_POS))        /**< PIN3 Mask                      */    
#define MXC_F_GPIO_FUNC_SEL_PIN4_POS                        16                                                                /**< PIN4 Position                  */  
#define MXC_F_GPIO_FUNC_SEL_PIN4                            ((uint32_t)(0x0000000FUL << MXC_F_GPIO_FUNC_SEL_PIN4_POS))        /**< PIN4 Mask                      */    
#define MXC_F_GPIO_FUNC_SEL_PIN5_POS                        20                                                                /**< PIN5 Position                  */  
#define MXC_F_GPIO_FUNC_SEL_PIN5                            ((uint32_t)(0x0000000FUL << MXC_F_GPIO_FUNC_SEL_PIN5_POS))        /**< PIN5 Mask                      */    
#define MXC_F_GPIO_FUNC_SEL_PIN6_POS                        24                                                                /**< PIN6 Position                  */  
#define MXC_F_GPIO_FUNC_SEL_PIN6                            ((uint32_t)(0x0000000FUL << MXC_F_GPIO_FUNC_SEL_PIN6_POS))        /**< PIN6 Mask                      */    
#define MXC_F_GPIO_FUNC_SEL_PIN7_POS                        28                                                                /**< PIN7 Position                  */  
#define MXC_F_GPIO_FUNC_SEL_PIN7                            ((uint32_t)(0x0000000FUL << MXC_F_GPIO_FUNC_SEL_PIN7_POS))        /**< PIN7 Mask                      */    
/**@} end of group GPIO_FUNC_SEL */
/**
 * @ingroup    gpio_registers
 * @defgroup   GPIO_IN_MODE_Register GPIO_IN_MODE
 * @brief      Field Positions and Bit Masks for the GPIO_IN_MODE register. 
 * @{
 */
#define MXC_F_GPIO_IN_MODE_PIN0_POS                         0                                                                 /**< PIN0 Position                  */
#define MXC_F_GPIO_IN_MODE_PIN0                             ((uint32_t)(0x00000003UL << MXC_F_GPIO_IN_MODE_PIN0_POS))         /**< PIN0 Mask                      */  
#define MXC_F_GPIO_IN_MODE_PIN1_POS                         4                                                                 /**< PIN1 Position                  */
#define MXC_F_GPIO_IN_MODE_PIN1                             ((uint32_t)(0x00000003UL << MXC_F_GPIO_IN_MODE_PIN1_POS))         /**< PIN1 Mask                      */  
#define MXC_F_GPIO_IN_MODE_PIN2_POS                         8                                                                 /**< PIN2 Position                  */
#define MXC_F_GPIO_IN_MODE_PIN2                             ((uint32_t)(0x00000003UL << MXC_F_GPIO_IN_MODE_PIN2_POS))         /**< PIN2 Mask                      */  
#define MXC_F_GPIO_IN_MODE_PIN3_POS                         12                                                                /**< PIN3 Position                  */  
#define MXC_F_GPIO_IN_MODE_PIN3                             ((uint32_t)(0x00000003UL << MXC_F_GPIO_IN_MODE_PIN3_POS))         /**< PIN3 Mask                      */  
#define MXC_F_GPIO_IN_MODE_PIN4_POS                         16                                                                /**< PIN4 Position                  */  
#define MXC_F_GPIO_IN_MODE_PIN4                             ((uint32_t)(0x00000003UL << MXC_F_GPIO_IN_MODE_PIN4_POS))         /**< PIN4 Mask                      */  
#define MXC_F_GPIO_IN_MODE_PIN5_POS                         20                                                                /**< PIN5 Position                  */  
#define MXC_F_GPIO_IN_MODE_PIN5                             ((uint32_t)(0x00000003UL << MXC_F_GPIO_IN_MODE_PIN5_POS))         /**< PIN5 Mask                      */  
#define MXC_F_GPIO_IN_MODE_PIN6_POS                         24                                                                /**< PIN6 Position                  */  
#define MXC_F_GPIO_IN_MODE_PIN6                             ((uint32_t)(0x00000003UL << MXC_F_GPIO_IN_MODE_PIN6_POS))         /**< PIN6 Mask                      */  
#define MXC_F_GPIO_IN_MODE_PIN7_POS                         28                                                                /**< PIN7 Position                  */  
#define MXC_F_GPIO_IN_MODE_PIN7                             ((uint32_t)(0x00000003UL << MXC_F_GPIO_IN_MODE_PIN7_POS))         /**< PIN7 Mask                      */  
/**@} end of group GPIO_IN_MODE */
/**
 * @ingroup    gpio_registers
 * @defgroup   GPIO_IN_VAL_Register GPIO_IN_VAL
 * @brief      Field Positions and Bit Masks for the GPIO_IN_VAL register. 
 * @{
 */
#define MXC_F_GPIO_IN_VAL_PIN0_POS                          0                                                                 /**< PIN0 Position                  */
#define MXC_F_GPIO_IN_VAL_PIN0                              ((uint32_t)(0x00000001UL << MXC_F_GPIO_IN_VAL_PIN0_POS))          /**< PIN0 Mask                      */  
#define MXC_F_GPIO_IN_VAL_PIN1_POS                          1                                                                 /**< PIN1 Position                  */
#define MXC_F_GPIO_IN_VAL_PIN1                              ((uint32_t)(0x00000001UL << MXC_F_GPIO_IN_VAL_PIN1_POS))          /**< PIN1 Mask                      */  
#define MXC_F_GPIO_IN_VAL_PIN2_POS                          2                                                                 /**< PIN2 Position                  */
#define MXC_F_GPIO_IN_VAL_PIN2                              ((uint32_t)(0x00000001UL << MXC_F_GPIO_IN_VAL_PIN2_POS))          /**< PIN2 Mask                      */  
#define MXC_F_GPIO_IN_VAL_PIN3_POS                          3                                                                 /**< PIN3 Position                  */
#define MXC_F_GPIO_IN_VAL_PIN3                              ((uint32_t)(0x00000001UL << MXC_F_GPIO_IN_VAL_PIN3_POS))          /**< PIN3 Mask                      */  
#define MXC_F_GPIO_IN_VAL_PIN4_POS                          4                                                                 /**< PIN4 Position                  */
#define MXC_F_GPIO_IN_VAL_PIN4                              ((uint32_t)(0x00000001UL << MXC_F_GPIO_IN_VAL_PIN4_POS))          /**< PIN4 Mask                      */  
#define MXC_F_GPIO_IN_VAL_PIN5_POS                          5                                                                 /**< PIN5 Position                  */
#define MXC_F_GPIO_IN_VAL_PIN5                              ((uint32_t)(0x00000001UL << MXC_F_GPIO_IN_VAL_PIN5_POS))          /**< PIN5 Mask                      */  
#define MXC_F_GPIO_IN_VAL_PIN6_POS                          6                                                                 /**< PIN6 Position                  */
#define MXC_F_GPIO_IN_VAL_PIN6                              ((uint32_t)(0x00000001UL << MXC_F_GPIO_IN_VAL_PIN6_POS))          /**< PIN6 Mask                      */  
#define MXC_F_GPIO_IN_VAL_PIN7_POS                          7                                                                 /**< PIN7 Position                  */
#define MXC_F_GPIO_IN_VAL_PIN7                              ((uint32_t)(0x00000001UL << MXC_F_GPIO_IN_VAL_PIN7_POS))          /**< PIN7 Mask                      */  
/**@} end of group GPIO_IN_VAL */
/**
 * @ingroup    gpio_registers
 * @defgroup   GPIO_INT_MODE_Register GPIO_INT_MODE
 * @brief      Field Positions and Bit Masks for the GPIO_INT_MODE register. 
 * @{
 */
#define MXC_F_GPIO_INT_MODE_PIN0_POS                        0                                                                 /**< PIN0 Position                  */
#define MXC_F_GPIO_INT_MODE_PIN0                            ((uint32_t)(0x00000007UL << MXC_F_GPIO_INT_MODE_PIN0_POS))        /**< PIN0 Mask                      */    
#define MXC_F_GPIO_INT_MODE_PIN1_POS                        4                                                                 /**< PIN1 Position                  */
#define MXC_F_GPIO_INT_MODE_PIN1                            ((uint32_t)(0x00000007UL << MXC_F_GPIO_INT_MODE_PIN1_POS))        /**< PIN1 Mask                      */    
#define MXC_F_GPIO_INT_MODE_PIN2_POS                        8                                                                 /**< PIN2 Position                  */
#define MXC_F_GPIO_INT_MODE_PIN2                            ((uint32_t)(0x00000007UL << MXC_F_GPIO_INT_MODE_PIN2_POS))        /**< PIN2 Mask                      */    
#define MXC_F_GPIO_INT_MODE_PIN3_POS                        12                                                                /**< PIN3 Position                  */  
#define MXC_F_GPIO_INT_MODE_PIN3                            ((uint32_t)(0x00000007UL << MXC_F_GPIO_INT_MODE_PIN3_POS))        /**< PIN3 Mask                      */    
#define MXC_F_GPIO_INT_MODE_PIN4_POS                        16                                                                /**< PIN4 Position                  */  
#define MXC_F_GPIO_INT_MODE_PIN4                            ((uint32_t)(0x00000007UL << MXC_F_GPIO_INT_MODE_PIN4_POS))        /**< PIN4 Mask                      */    
#define MXC_F_GPIO_INT_MODE_PIN5_POS                        20                                                                /**< PIN5 Position                  */  
#define MXC_F_GPIO_INT_MODE_PIN5                            ((uint32_t)(0x00000007UL << MXC_F_GPIO_INT_MODE_PIN5_POS))        /**< PIN5 Mask                      */    
#define MXC_F_GPIO_INT_MODE_PIN6_POS                        24                                                                /**< PIN6 Position                  */  
#define MXC_F_GPIO_INT_MODE_PIN6                            ((uint32_t)(0x00000007UL << MXC_F_GPIO_INT_MODE_PIN6_POS))        /**< PIN6 Mask                      */    
#define MXC_F_GPIO_INT_MODE_PIN7_POS                        28                                                                /**< PIN7 Position                  */  
#define MXC_F_GPIO_INT_MODE_PIN7                            ((uint32_t)(0x00000007UL << MXC_F_GPIO_INT_MODE_PIN7_POS))        /**< PIN7 Mask                      */    
/**@} end of group GPIO_INT_MODE */
/**
 * @ingroup    gpio_registers
 * @defgroup   GPIO_INTFL_Register GPIO_INTFL
 * @brief      Field Positions and Bit Masks for the GPIO_INTFL register. 
 * @{
 */
#define MXC_F_GPIO_INTFL_PIN0_POS                           0                                                                 /**< PIN0 Position                  */
#define MXC_F_GPIO_INTFL_PIN0                               ((uint32_t)(0x00000001UL << MXC_F_GPIO_INTFL_PIN0_POS))           /**< PIN0 Mask                      */
#define MXC_F_GPIO_INTFL_PIN1_POS                           1                                                                 /**< PIN1 Position                  */
#define MXC_F_GPIO_INTFL_PIN1                               ((uint32_t)(0x00000001UL << MXC_F_GPIO_INTFL_PIN1_POS))           /**< PIN1 Mask                      */
#define MXC_F_GPIO_INTFL_PIN2_POS                           2                                                                 /**< PIN2 Position                  */
#define MXC_F_GPIO_INTFL_PIN2                               ((uint32_t)(0x00000001UL << MXC_F_GPIO_INTFL_PIN2_POS))           /**< PIN2 Mask                      */
#define MXC_F_GPIO_INTFL_PIN3_POS                           3                                                                 /**< PIN3 Position                  */
#define MXC_F_GPIO_INTFL_PIN3                               ((uint32_t)(0x00000001UL << MXC_F_GPIO_INTFL_PIN3_POS))           /**< PIN3 Mask                      */
#define MXC_F_GPIO_INTFL_PIN4_POS                           4                                                                 /**< PIN4 Position                  */
#define MXC_F_GPIO_INTFL_PIN4                               ((uint32_t)(0x00000001UL << MXC_F_GPIO_INTFL_PIN4_POS))           /**< PIN4 Mask                      */
#define MXC_F_GPIO_INTFL_PIN5_POS                           5                                                                 /**< PIN5 Position                  */
#define MXC_F_GPIO_INTFL_PIN5                               ((uint32_t)(0x00000001UL << MXC_F_GPIO_INTFL_PIN5_POS))           /**< PIN5 Mask                      */
#define MXC_F_GPIO_INTFL_PIN6_POS                           6                                                                 /**< PIN6 Position                  */
#define MXC_F_GPIO_INTFL_PIN6                               ((uint32_t)(0x00000001UL << MXC_F_GPIO_INTFL_PIN6_POS))           /**< PIN6 Mask                      */
#define MXC_F_GPIO_INTFL_PIN7_POS                           7                                                                 /**< PIN7 Position                  */
#define MXC_F_GPIO_INTFL_PIN7                               ((uint32_t)(0x00000001UL << MXC_F_GPIO_INTFL_PIN7_POS))           /**< PIN7 Mask                      */
/**@} end of group GPIO_INTFL */
/**
 * @ingroup    gpio_registers
 * @defgroup   GPIO_INTEN_Register GPIO_INTEN
 * @brief      Field Positions and Bit Masks for the GPIO_INTEN register. 
 * @{
 */
#define MXC_F_GPIO_INTEN_PIN0_POS                           0                                                                 /**< PIN0 Position                  */
#define MXC_F_GPIO_INTEN_PIN0                               ((uint32_t)(0x00000001UL << MXC_F_GPIO_INTEN_PIN0_POS))           /**< PIN0 Mask                      */
#define MXC_F_GPIO_INTEN_PIN1_POS                           1                                                                 /**< PIN1 Position                  */
#define MXC_F_GPIO_INTEN_PIN1                               ((uint32_t)(0x00000001UL << MXC_F_GPIO_INTEN_PIN1_POS))           /**< PIN1 Mask                      */
#define MXC_F_GPIO_INTEN_PIN2_POS                           2                                                                 /**< PIN2 Position                  */
#define MXC_F_GPIO_INTEN_PIN2                               ((uint32_t)(0x00000001UL << MXC_F_GPIO_INTEN_PIN2_POS))           /**< PIN2 Mask                      */
#define MXC_F_GPIO_INTEN_PIN3_POS                           3                                                                 /**< PIN3 Position                  */
#define MXC_F_GPIO_INTEN_PIN3                               ((uint32_t)(0x00000001UL << MXC_F_GPIO_INTEN_PIN3_POS))           /**< PIN3 Mask                      */
#define MXC_F_GPIO_INTEN_PIN4_POS                           4                                                                 /**< PIN4 Position                  */
#define MXC_F_GPIO_INTEN_PIN4                               ((uint32_t)(0x00000001UL << MXC_F_GPIO_INTEN_PIN4_POS))           /**< PIN4 Mask                      */
#define MXC_F_GPIO_INTEN_PIN5_POS                           5                                                                 /**< PIN5 Position                  */
#define MXC_F_GPIO_INTEN_PIN5                               ((uint32_t)(0x00000001UL << MXC_F_GPIO_INTEN_PIN5_POS))           /**< PIN5 Mask                      */
#define MXC_F_GPIO_INTEN_PIN6_POS                           6                                                                 /**< PIN6 Position                  */
#define MXC_F_GPIO_INTEN_PIN6                               ((uint32_t)(0x00000001UL << MXC_F_GPIO_INTEN_PIN6_POS))           /**< PIN6 Mask                      */
#define MXC_F_GPIO_INTEN_PIN7_POS                           7                                                                 /**< PIN7 Position                  */
#define MXC_F_GPIO_INTEN_PIN7                               ((uint32_t)(0x00000001UL << MXC_F_GPIO_INTEN_PIN7_POS))           /**< PIN7 Mask                      */
/**@} end group GPIO_INTEN_Register */


/*
   Field values and shifted values for module GPIO.
*/
/**
 * @ingroup GPIO_RST_MODE_Register
 * @defgroup GPIO_RST_MODE_Values Reset Mode Values
 * @brief   Mode Values for setting the GPIO_RST_MODE Field for different pad modes
 * @{
 */
#define MXC_V_GPIO_RST_MODE_DRIVE_0                                             ((uint32_t)(0x00000000UL))              /**< DRIVE_0                   */
#define MXC_V_GPIO_RST_MODE_WEAK_PULLDOWN                                       ((uint32_t)(0x00000001UL))              /**< WEAK_PULLDOWN             */
#define MXC_V_GPIO_RST_MODE_WEAK_PULLUP                                         ((uint32_t)(0x00000002UL))              /**< WEAK_PULLUP               */
#define MXC_V_GPIO_RST_MODE_DRIVE_1                                             ((uint32_t)(0x00000003UL))              /**< DRIVE_1                   */
#define MXC_V_GPIO_RST_MODE_HIGH_Z                                              ((uint32_t)(0x00000004UL))              /**< HIGH_Z                    */
/**@}*/

/**
 * @ingroup GPIO_FREE_Register
 * @defgroup GPIO_FREE_Values Reset Mode Values
 * @brief   Mode Values for setting the GPIO_FREE to Available or Unavailable
 * @{
 */
#define MXC_V_GPIO_FREE_NOT_AVAILABLE                                           ((uint32_t)(0x00000000UL))              /**< GPIO Pin is Unavailable                */
#define MXC_V_GPIO_FREE_AVAILABLE                                               ((uint32_t)(0x00000001UL))              /**< GPIO Pin is Available                  */
/**@}*/

/**
 * @ingroup GPIO_FREE_Register
 * @defgroup GPIO_OUT_MODE_Values Output Mode Values
 * @brief   GPIO_OUT_MODE values for setting the different port pin output modes 
 * @{
 */
#define MXC_V_GPIO_OUT_MODE_HIGH_Z_WEAK_PULLUP                                  ((uint32_t)(0x00000000UL))              /**< See \MXIM_Device User Guide for details: HIGH_Z_WEAK_PULLUP        */
#define MXC_V_GPIO_OUT_MODE_OPEN_DRAIN                                          ((uint32_t)(0x00000001UL))              /**< See \MXIM_Device User Guide for details: OPEN_DRAIN                */
#define MXC_V_GPIO_OUT_MODE_OPEN_DRAIN_WEAK_PULLUP                              ((uint32_t)(0x00000002UL))              /**< See \MXIM_Device User Guide for details: OPEN_DRAIN_WEAK_PULLUP    */
#define MXC_V_GPIO_OUT_MODE_NORMAL_HIGH_Z                                       ((uint32_t)(0x00000004UL))              /**< See \MXIM_Device User Guide for details: NORMAL_HIGH_Z             */
#define MXC_V_GPIO_OUT_MODE_NORMAL                                              ((uint32_t)(0x00000005UL))              /**< See \MXIM_Device User Guide for details: NORMAL                    */
#define MXC_V_GPIO_OUT_MODE_SLOW_HIGH_Z                                         ((uint32_t)(0x00000006UL))              /**< See \MXIM_Device User Guide for details: SLOW_HIGH_Z               */
#define MXC_V_GPIO_OUT_MODE_SLOW_DRIVE                                          ((uint32_t)(0x00000007UL))              /**< See \MXIM_Device User Guide for details: SLOW_DRIVE                */
#define MXC_V_GPIO_OUT_MODE_FAST_HIGH_Z                                         ((uint32_t)(0x00000008UL))              /**< See \MXIM_Device User Guide for details: FAST_HIGH_Z               */
#define MXC_V_GPIO_OUT_MODE_FAST_DRIVE                                          ((uint32_t)(0x00000009UL))              /**< See \MXIM_Device User Guide for details: FAST_DRIVE                */
#define MXC_V_GPIO_OUT_MODE_HIGH_Z_WEAK_PULLDOWN                                ((uint32_t)(0x0000000AUL))              /**< See \MXIM_Device User Guide for details: HIGH_Z_WEAK_PULLDOWN      */
#define MXC_V_GPIO_OUT_MODE_OPEN_SOURCE                                         ((uint32_t)(0x0000000BUL))              /**< See \MXIM_Device User Guide for details: OPEN_SOURCE               */
#define MXC_V_GPIO_OUT_MODE_OPEN_SOURCE_WEAK_PULLDOWN                           ((uint32_t)(0x0000000CUL))              /**< See \MXIM_Device User Guide for details: OPEN_SOURCE_WEAK_PULLDOWN */
#define MXC_V_GPIO_OUT_MODE_HIGH_Z_INPUT_DISABLED                               ((uint32_t)(0x0000000FUL))              /**< See \MXIM_Device User Guide for details: HIGH_Z_INPUT_DISABLED     */
/**@}*/

/**
 * @ingroup GPIO_FUNC_SEL_Register
 * @defgroup GPIO_FUNC_SEL_Values Function type selection values
 * @brief   Function selection values for the GPIO_FUNC_SEL Register.
 * @{
 */
#define MXC_V_GPIO_FUNC_SEL_MODE_GPIO                                           ((uint32_t)(0x00000000UL))              /**< Standard GPIO Mode                 */
#define MXC_V_GPIO_FUNC_SEL_MODE_PT                                             ((uint32_t)(0x00000001UL))              /**< Pulse Train Mode                   */
#define MXC_V_GPIO_FUNC_SEL_MODE_TMR                                            ((uint32_t)(0x00000002UL))              /**< Timer Mode                         */
/**@}*/

/**
 * @ingroup GPIO_IN_MODE_Register
 * @defgroup GPIO_IN_MODE_Values Input mode selection values
 * @brief   Input mode values for selecting the GPIO input mode. 
 * @{
 */
#define MXC_V_GPIO_IN_MODE_NORMAL                                               ((uint32_t)(0x00000000UL))              /**< Normal Input Mode                  */
#define MXC_V_GPIO_IN_MODE_INVERTED                                             ((uint32_t)(0x00000001UL))              /**< Inverted Input Mode                */
#define MXC_V_GPIO_IN_MODE_ALWAYS_ZERO                                          ((uint32_t)(0x00000002UL))              /**< Always reads 0                     */
#define MXC_V_GPIO_IN_MODE_ALWAYS_ONE                                           ((uint32_t)(0x00000003UL))              /**< Always reads 1                     */
/**@}*/

/**
 * @ingroup GPIO_INT_MODE_Register
 * @defgroup GPIO_INT_MODE_Values Interrupt mode selection values
 * @brief   Values for setting the interrupt mode of a GPIO input pin. 
 * @{
 */
#define MXC_V_GPIO_INT_MODE_DISABLE                                             ((uint32_t)(0x00000000UL))              /**< Disable Interrupt for a given port pin */
#define MXC_V_GPIO_INT_MODE_FALLING_EDGE                                        ((uint32_t)(0x00000001UL))              /**< Interrupt on falling edge              */
#define MXC_V_GPIO_INT_MODE_RISING_EDGE                                         ((uint32_t)(0x00000002UL))              /**< Interrupt on rising edge               */
#define MXC_V_GPIO_INT_MODE_ANY_EDGE                                            ((uint32_t)(0x00000003UL))              /**< Interrupt on rising or falling edge    */
#define MXC_V_GPIO_INT_MODE_LOW_LVL                                             ((uint32_t)(0x00000004UL))              /**< Interrupt on Low Level                 */
#define MXC_V_GPIO_INT_MODE_HIGH_LVL                                            ((uint32_t)(0x00000005UL))              /**< Interrupt on High Level                */
/**@}*/

/**@}*/
#ifdef __cplusplus
}
#endif

#endif   /* _MXC_GPIO_REGS_H_ */

