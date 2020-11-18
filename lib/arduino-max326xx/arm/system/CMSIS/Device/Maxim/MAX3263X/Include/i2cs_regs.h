/**
 * @file
 * @brief   Registers, Bit Masks and Bit Positions for the I2CS Peripheral Module.
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
 * $Date: 2016-10-10 18:59:48 -0500 (Mon, 10 Oct 2016) $
 * $Revision: 24661 $
 *
 *************************************************************************** */

/* Define to prevent redundant inclusion */
#ifndef _MXC_I2CS_REGS_H_
#define _MXC_I2CS_REGS_H_

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

/**
 * @ingroup     i2cs
 * @defgroup    i2cs_registers Registers
 * @brief       Registers, Bit Masks and Bit Positions for the I2CS Peripheral Module.
 * @{
 */

/**
 * Structure type to access the I2CS Peripheral Module Registers
 */
 typedef struct {
    __IO uint32_t clk_div;                              /**< <tt>\b 0x0000:</tt> I2CS_CLK_DIV Register - Clock Divisor Control               */
    __IO uint32_t dev_id;                               /**< <tt>\b 0x0004:</tt> I2CS_DEV_ID Register - Device ID Register                   */
    __IO uint32_t intfl;                                /**< <tt>\b 0x0008:</tt> I2CS_INTFL Register - Interrupt Flags                       */
    __IO uint32_t inten;                                /**< <tt>\b 0x000C:</tt> I2CS_INTEN Register - Interrupt Enable                      */
    __IO uint32_t data_byte[32];                        /**< <tt>\b 0x0010-0x008C:</tt> I2CS_DATA_BYTE - Data Byte                                  */
} mxc_i2cs_regs_t;
/**@} end of i2cs_registers */


/*
   Register offsets for module I2CS.
*/
/**
 * @ingroup    i2cs_registers
 * @defgroup   I2CS_Register_Offsets Register Offsets
 * @brief      I2C Slave Register Offsets from the I2CS Base Peripheral Address. 
 * @{
 */
#define MXC_R_I2CS_OFFS_CLK_DIV                             ((uint32_t)0x00000000UL)     /**< Offset from I2CS Base Peripheral Address: <tt>\b 0x0000</tt>  */
#define MXC_R_I2CS_OFFS_DEV_ID                              ((uint32_t)0x00000004UL)     /**< Offset from I2CS Base Peripheral Address: <tt>\b 0x0004</tt>  */
#define MXC_R_I2CS_OFFS_INTFL                               ((uint32_t)0x00000008UL)     /**< Offset from I2CS Base Peripheral Address: <tt>\b 0x0008</tt>  */
#define MXC_R_I2CS_OFFS_INTEN                               ((uint32_t)0x0000000CUL)     /**< Offset from I2CS Base Peripheral Address: <tt>\b 0x000C</tt>  */
#define MXC_R_I2CS_OFFS_DATA_BYTE                           ((uint32_t)0x00000010UL)     /**< Offset from I2CS Base Peripheral Address: <tt>\b 0x0010-0x008C</tt>   */
/**@} I2CS_Register_Offsets */
/*
   Field positions and masks for module I2CS.
*/
/**
 * @ingroup  i2cs_registers
 * @defgroup I2CS_CLK_DIV_Register I2CS_CLK_DIV
 * @brief    Field Positions and Bit Masks for the I2CS_CLK_DIV register
 * @{
 */
#define MXC_F_I2CS_CLK_DIV_FS_FILTER_CLOCK_DIV_POS          0                                                                        /**< FS_FILTER_CLOCK_DIV Position */
#define MXC_F_I2CS_CLK_DIV_FS_FILTER_CLOCK_DIV              ((uint32_t)(0x000000FFUL << MXC_F_I2CS_CLK_DIV_FS_FILTER_CLOCK_DIV_POS)) /**< FS_FILTER_CLOCK_DIV Mask     */
/**@} end group I2CS_CLK_DIV */
/**
 * @ingroup  i2cs_registers
 * @defgroup I2CS_DEV_ID_Register I2CS_DEV_ID
 * @brief    Field Positions and Bit Masks for the I2CS_DEV_ID register
 * @{
 */
#define MXC_F_I2CS_DEV_ID_SLAVE_DEV_ID_POS                  0                                                                       /**< SLAVE_DEV_ID Position          */
#define MXC_F_I2CS_DEV_ID_SLAVE_DEV_ID                      ((uint32_t)(0x000003FFUL << MXC_F_I2CS_DEV_ID_SLAVE_DEV_ID_POS))        /**< SLAVE_DEV_ID Mask              */  
#define MXC_F_I2CS_DEV_ID_TEN_BIT_ID_MODE_POS               12                                                                      /**< TEN_BIT_ID_MODE Position       */
#define MXC_F_I2CS_DEV_ID_TEN_BIT_ID_MODE                   ((uint32_t)(0x00000001UL << MXC_F_I2CS_DEV_ID_TEN_BIT_ID_MODE_POS))     /**< TEN_BIT_ID_MODE Mask           */
#define MXC_F_I2CS_DEV_ID_SLAVE_RESET_POS                   14                                                                      /**< SLAVE_RESET Position           */
#define MXC_F_I2CS_DEV_ID_SLAVE_RESET                       ((uint32_t)(0x00000001UL << MXC_F_I2CS_DEV_ID_SLAVE_RESET_POS))         /**< SLAVE_RESET Mask               */
/**@} end group I2CS_DEV_ID */
/**
 * @ingroup  i2cs_registers
 * @defgroup I2CS_INTFL_Register I2CS_INTFL
 * @brief    Field Positions and Bit Masks for the I2CS_INTFL register
 * @{
 */
#define MXC_F_I2CS_INTFL_BYTE0_POS                          0                                                                        /**< BYTE0 Position                  */
#define MXC_F_I2CS_INTFL_BYTE0                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE0_POS))                 /**< BYTE0 Mask                      */
#define MXC_F_I2CS_INTFL_BYTE1_POS                          1                                                                        /**< BYTE1 Position                  */
#define MXC_F_I2CS_INTFL_BYTE1                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE1_POS))                 /**< BYTE1 Mask                      */
#define MXC_F_I2CS_INTFL_BYTE2_POS                          2                                                                        /**< BYTE2 Position                  */
#define MXC_F_I2CS_INTFL_BYTE2                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE2_POS))                 /**< BYTE2 Mask                      */
#define MXC_F_I2CS_INTFL_BYTE3_POS                          3                                                                        /**< BYTE3 Position                  */
#define MXC_F_I2CS_INTFL_BYTE3                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE3_POS))                 /**< BYTE3 Mask                      */
#define MXC_F_I2CS_INTFL_BYTE4_POS                          4                                                                        /**< BYTE4 Position                  */
#define MXC_F_I2CS_INTFL_BYTE4                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE4_POS))                 /**< BYTE4 Mask                      */
#define MXC_F_I2CS_INTFL_BYTE5_POS                          5                                                                        /**< BYTE5 Position                  */
#define MXC_F_I2CS_INTFL_BYTE5                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE5_POS))                 /**< BYTE5 Mask                      */
#define MXC_F_I2CS_INTFL_BYTE6_POS                          6                                                                        /**< BYTE6 Position                  */
#define MXC_F_I2CS_INTFL_BYTE6                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE6_POS))                 /**< BYTE6 Mask                      */
#define MXC_F_I2CS_INTFL_BYTE7_POS                          7                                                                        /**< BYTE7 Position                  */
#define MXC_F_I2CS_INTFL_BYTE7                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE7_POS))                 /**< BYTE7 Mask                      */
#define MXC_F_I2CS_INTFL_BYTE8_POS                          8                                                                        /**< BYTE8 Position                  */
#define MXC_F_I2CS_INTFL_BYTE8                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE8_POS))                 /**< BYTE8 Mask                      */
#define MXC_F_I2CS_INTFL_BYTE9_POS                          9                                                                        /**< BYTE9 Position                  */
#define MXC_F_I2CS_INTFL_BYTE9                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE9_POS))                 /**< BYTE9 Mask                      */
#define MXC_F_I2CS_INTFL_BYTE10_POS                         10                                                                       /**< BYTE10 Position                 */
#define MXC_F_I2CS_INTFL_BYTE10                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE10_POS))                /**< BYTE10 Mask                     */ 
#define MXC_F_I2CS_INTFL_BYTE11_POS                         11                                                                       /**< BYTE11 Position                 */
#define MXC_F_I2CS_INTFL_BYTE11                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE11_POS))                /**< BYTE11 Mask                     */ 
#define MXC_F_I2CS_INTFL_BYTE12_POS                         12                                                                       /**< BYTE12 Position                 */
#define MXC_F_I2CS_INTFL_BYTE12                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE12_POS))                /**< BYTE12 Mask                     */ 
#define MXC_F_I2CS_INTFL_BYTE13_POS                         13                                                                       /**< BYTE13 Position                 */
#define MXC_F_I2CS_INTFL_BYTE13                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE13_POS))                /**< BYTE13 Mask                     */ 
#define MXC_F_I2CS_INTFL_BYTE14_POS                         14                                                                       /**< BYTE14 Position                 */
#define MXC_F_I2CS_INTFL_BYTE14                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE14_POS))                /**< BYTE14 Mask                     */ 
#define MXC_F_I2CS_INTFL_BYTE15_POS                         15                                                                       /**< BYTE15 Position                 */
#define MXC_F_I2CS_INTFL_BYTE15                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE15_POS))                /**< BYTE15 Mask                     */ 
#define MXC_F_I2CS_INTFL_BYTE16_POS                         16                                                                       /**< BYTE16 Position                 */
#define MXC_F_I2CS_INTFL_BYTE16                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE16_POS))                /**< BYTE16 Mask                     */ 
#define MXC_F_I2CS_INTFL_BYTE17_POS                         17                                                                       /**< BYTE17 Position                 */
#define MXC_F_I2CS_INTFL_BYTE17                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE17_POS))                /**< BYTE17 Mask                     */ 
#define MXC_F_I2CS_INTFL_BYTE18_POS                         18                                                                       /**< BYTE18 Position                 */ 
#define MXC_F_I2CS_INTFL_BYTE18                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE18_POS))                /**< BYTE18 Mask                     */ 
#define MXC_F_I2CS_INTFL_BYTE19_POS                         19                                                                       /**< BYTE19 Position                 */ 
#define MXC_F_I2CS_INTFL_BYTE19                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE19_POS))                /**< BYTE19 Mask                     */ 
#define MXC_F_I2CS_INTFL_BYTE20_POS                         20                                                                       /**< BYTE20 Position                 */ 
#define MXC_F_I2CS_INTFL_BYTE20                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE20_POS))                /**< BYTE20 Mask                     */ 
#define MXC_F_I2CS_INTFL_BYTE21_POS                         21                                                                       /**< BYTE21 Position                 */ 
#define MXC_F_I2CS_INTFL_BYTE21                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE21_POS))                /**< BYTE21 Mask                     */ 
#define MXC_F_I2CS_INTFL_BYTE22_POS                         22                                                                       /**< BYTE22 Position                 */ 
#define MXC_F_I2CS_INTFL_BYTE22                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE22_POS))                /**< BYTE22 Mask                     */ 
#define MXC_F_I2CS_INTFL_BYTE23_POS                         23                                                                       /**< BYTE23 Position                 */ 
#define MXC_F_I2CS_INTFL_BYTE23                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE23_POS))                /**< BYTE23 Mask                     */ 
#define MXC_F_I2CS_INTFL_BYTE24_POS                         24                                                                       /**< BYTE24 Position                 */ 
#define MXC_F_I2CS_INTFL_BYTE24                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE24_POS))                /**< BYTE24 Mask                     */ 
#define MXC_F_I2CS_INTFL_BYTE25_POS                         25                                                                       /**< BYTE25 Position                 */ 
#define MXC_F_I2CS_INTFL_BYTE25                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE25_POS))                /**< BYTE25 Mask                     */ 
#define MXC_F_I2CS_INTFL_BYTE26_POS                         26                                                                       /**< BYTE26 Position                 */ 
#define MXC_F_I2CS_INTFL_BYTE26                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE26_POS))                /**< BYTE26 Mask                     */ 
#define MXC_F_I2CS_INTFL_BYTE27_POS                         27                                                                       /**< BYTE27 Position                 */ 
#define MXC_F_I2CS_INTFL_BYTE27                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE27_POS))                /**< BYTE27 Mask                     */ 
#define MXC_F_I2CS_INTFL_BYTE28_POS                         28                                                                       /**< BYTE28 Position                 */ 
#define MXC_F_I2CS_INTFL_BYTE28                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE28_POS))                /**< BYTE28 Mask                     */ 
#define MXC_F_I2CS_INTFL_BYTE29_POS                         29                                                                       /**< BYTE29 Position                 */ 
#define MXC_F_I2CS_INTFL_BYTE29                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE29_POS))                /**< BYTE29 Mask                     */ 
#define MXC_F_I2CS_INTFL_BYTE30_POS                         30                                                                       /**< BYTE30 Position                 */ 
#define MXC_F_I2CS_INTFL_BYTE30                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE30_POS))                /**< BYTE30 Mask                     */ 
#define MXC_F_I2CS_INTFL_BYTE31_POS                         31                                                                       /**< BYTE31 Position                 */ 
#define MXC_F_I2CS_INTFL_BYTE31                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE31_POS))                /**< BYTE31 Mask                     */ 
/**@} end group I2CS_INTFL */
/**
 * @ingroup  i2cs_registers
 * @defgroup I2CS_INTEN_Register I2CS_INTEN
 * @brief    Field Positions and Bit Masks for the I2CS_INTEN register
 * @{
 */
#define MXC_F_I2CS_INTEN_BYTE0_POS                          0                                                                        /**< BYTE0 Position                  */
#define MXC_F_I2CS_INTEN_BYTE0                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE0_POS))                 /**< BYTE0 Mask                      */
#define MXC_F_I2CS_INTEN_BYTE1_POS                          1                                                                        /**< BYTE1 Position                  */
#define MXC_F_I2CS_INTEN_BYTE1                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE1_POS))                 /**< BYTE1 Mask                      */
#define MXC_F_I2CS_INTEN_BYTE2_POS                          2                                                                        /**< BYTE2 Position                  */
#define MXC_F_I2CS_INTEN_BYTE2                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE2_POS))                 /**< BYTE2 Mask                      */
#define MXC_F_I2CS_INTEN_BYTE3_POS                          3                                                                        /**< BYTE3 Position                  */
#define MXC_F_I2CS_INTEN_BYTE3                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE3_POS))                 /**< BYTE3 Mask                      */
#define MXC_F_I2CS_INTEN_BYTE4_POS                          4                                                                        /**< BYTE4 Position                  */
#define MXC_F_I2CS_INTEN_BYTE4                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE4_POS))                 /**< BYTE4 Mask                      */
#define MXC_F_I2CS_INTEN_BYTE5_POS                          5                                                                        /**< BYTE5 Position                  */
#define MXC_F_I2CS_INTEN_BYTE5                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE5_POS))                 /**< BYTE5 Mask                      */
#define MXC_F_I2CS_INTEN_BYTE6_POS                          6                                                                        /**< BYTE6 Position                  */
#define MXC_F_I2CS_INTEN_BYTE6                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE6_POS))                 /**< BYTE6 Mask                      */
#define MXC_F_I2CS_INTEN_BYTE7_POS                          7                                                                        /**< BYTE7 Position                  */
#define MXC_F_I2CS_INTEN_BYTE7                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE7_POS))                 /**< BYTE7 Mask                      */
#define MXC_F_I2CS_INTEN_BYTE8_POS                          8                                                                        /**< BYTE8 Position                  */
#define MXC_F_I2CS_INTEN_BYTE8                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE8_POS))                 /**< BYTE8 Mask                      */
#define MXC_F_I2CS_INTEN_BYTE9_POS                          9                                                                        /**< BYTE9 Position                  */
#define MXC_F_I2CS_INTEN_BYTE9                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE9_POS))                 /**< BYTE9 Mask                      */
#define MXC_F_I2CS_INTEN_BYTE10_POS                         10                                                                       /**< BYTE10 Position                 */ 
#define MXC_F_I2CS_INTEN_BYTE10                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE10_POS))                /**< BYTE10 Mask                     */ 
#define MXC_F_I2CS_INTEN_BYTE11_POS                         11                                                                       /**< BYTE11 Position                 */ 
#define MXC_F_I2CS_INTEN_BYTE11                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE11_POS))                /**< BYTE11 Mask                     */ 
#define MXC_F_I2CS_INTEN_BYTE12_POS                         12                                                                       /**< BYTE12 Position                 */ 
#define MXC_F_I2CS_INTEN_BYTE12                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE12_POS))                /**< BYTE12 Mask                     */ 
#define MXC_F_I2CS_INTEN_BYTE13_POS                         13                                                                       /**< BYTE13 Position                 */ 
#define MXC_F_I2CS_INTEN_BYTE13                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE13_POS))                /**< BYTE13 Mask                     */ 
#define MXC_F_I2CS_INTEN_BYTE14_POS                         14                                                                       /**< BYTE14 Position                 */ 
#define MXC_F_I2CS_INTEN_BYTE14                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE14_POS))                /**< BYTE14 Mask                     */ 
#define MXC_F_I2CS_INTEN_BYTE15_POS                         15                                                                       /**< BYTE15 Position                 */ 
#define MXC_F_I2CS_INTEN_BYTE15                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE15_POS))                /**< BYTE15 Mask                     */ 
#define MXC_F_I2CS_INTEN_BYTE16_POS                         16                                                                       /**< BYTE16 Position                 */ 
#define MXC_F_I2CS_INTEN_BYTE16                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE16_POS))                /**< BYTE16 Mask                     */ 
#define MXC_F_I2CS_INTEN_BYTE17_POS                         17                                                                       /**< BYTE17 Position                 */ 
#define MXC_F_I2CS_INTEN_BYTE17                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE17_POS))                /**< BYTE17 Mask                     */ 
#define MXC_F_I2CS_INTEN_BYTE18_POS                         18                                                                       /**< BYTE18 Position                 */ 
#define MXC_F_I2CS_INTEN_BYTE18                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE18_POS))                /**< BYTE18 Mask                     */ 
#define MXC_F_I2CS_INTEN_BYTE19_POS                         19                                                                       /**< BYTE19 Position                 */ 
#define MXC_F_I2CS_INTEN_BYTE19                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE19_POS))                /**< BYTE19 Mask                     */ 
#define MXC_F_I2CS_INTEN_BYTE20_POS                         20                                                                       /**< BYTE20 Position                 */ 
#define MXC_F_I2CS_INTEN_BYTE20                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE20_POS))                /**< BYTE20 Mask                     */ 
#define MXC_F_I2CS_INTEN_BYTE21_POS                         21                                                                       /**< BYTE21 Position                 */ 
#define MXC_F_I2CS_INTEN_BYTE21                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE21_POS))                /**< BYTE21 Mask                     */ 
#define MXC_F_I2CS_INTEN_BYTE22_POS                         22                                                                       /**< BYTE22 Position                 */ 
#define MXC_F_I2CS_INTEN_BYTE22                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE22_POS))                /**< BYTE22 Mask                     */ 
#define MXC_F_I2CS_INTEN_BYTE23_POS                         23                                                                       /**< BYTE23 Position                 */ 
#define MXC_F_I2CS_INTEN_BYTE23                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE23_POS))                /**< BYTE23 Mask                     */ 
#define MXC_F_I2CS_INTEN_BYTE24_POS                         24                                                                       /**< BYTE24 Position                 */ 
#define MXC_F_I2CS_INTEN_BYTE24                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE24_POS))                /**< BYTE24 Mask                     */ 
#define MXC_F_I2CS_INTEN_BYTE25_POS                         25                                                                       /**< BYTE25 Position                 */ 
#define MXC_F_I2CS_INTEN_BYTE25                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE25_POS))                /**< BYTE25 Mask                     */ 
#define MXC_F_I2CS_INTEN_BYTE26_POS                         26                                                                       /**< BYTE26 Position                 */ 
#define MXC_F_I2CS_INTEN_BYTE26                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE26_POS))                /**< BYTE26 Mask                     */ 
#define MXC_F_I2CS_INTEN_BYTE27_POS                         27                                                                       /**< BYTE27 Position                 */ 
#define MXC_F_I2CS_INTEN_BYTE27                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE27_POS))                /**< BYTE27 Mask                     */ 
#define MXC_F_I2CS_INTEN_BYTE28_POS                         28                                                                       /**< BYTE28 Position                 */ 
#define MXC_F_I2CS_INTEN_BYTE28                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE28_POS))                /**< BYTE28 Mask                     */ 
#define MXC_F_I2CS_INTEN_BYTE29_POS                         29                                                                       /**< BYTE29 Position                 */ 
#define MXC_F_I2CS_INTEN_BYTE29                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE29_POS))                /**< BYTE29 Mask                     */ 
#define MXC_F_I2CS_INTEN_BYTE30_POS                         30                                                                       /**< BYTE30 Position                 */ 
#define MXC_F_I2CS_INTEN_BYTE30                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE30_POS))                /**< BYTE30 Mask                     */ 
#define MXC_F_I2CS_INTEN_BYTE31_POS                         31                                                                       /**< BYTE31 Position                 */ 
#define MXC_F_I2CS_INTEN_BYTE31                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE31_POS))                /**< BYTE31 Mask                     */ 
/**@} end group I2CS_INTEN */
/**
 * @ingroup  i2cs_registers
 * @defgroup I2CS_DATA_BYTE_Register I2CS_DATA_BYTE
 * @brief    Field Positions and Bit Masks for the I2CS_DATA_BYTE register
 * @{
 */
#define MXC_F_I2CS_DATA_BYTE_DATA_FIELD_POS                 0                                                                       /**< DATA_FIELD Position         */
#define MXC_F_I2CS_DATA_BYTE_DATA_FIELD                     ((uint32_t)(0x000000FFUL << MXC_F_I2CS_DATA_BYTE_DATA_FIELD_POS))       /**< DATA_FIELD                  */
#define MXC_F_I2CS_DATA_BYTE_READ_ONLY_FL_POS               8                                                                       /**< READ_ONLY_FL Position       */
#define MXC_F_I2CS_DATA_BYTE_READ_ONLY_FL                   ((uint32_t)(0x00000001UL << MXC_F_I2CS_DATA_BYTE_READ_ONLY_FL_POS))     /**< READ_ONLY_FL                */
#define MXC_F_I2CS_DATA_BYTE_DATA_UPDATED_FL_POS            9                                                                       /**< DATA_UPDATED_FL Position    */
#define MXC_F_I2CS_DATA_BYTE_DATA_UPDATED_FL                ((uint32_t)(0x00000001UL << MXC_F_I2CS_DATA_BYTE_DATA_UPDATED_FL_POS))  /**< DATA_UPDATED_FL             */
/**@} end group I2CS_DATA_BYTE */

#ifdef __cplusplus
}
#endif

#endif   /* _MXC_I2CS_REGS_H_ */
