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
 * $Date: 2016-03-11 11:46:37 -0600 (Fri, 11 Mar 2016) $
 * $Revision: 21839 $
 *
 ******************************************************************************/

#ifndef _MXC_I2CS_REGS_H_
#define _MXC_I2CS_REGS_H_

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
    __IO uint32_t clk_div;                              /*  0x0000          I2C Slave Clock Divisor Control                                              */
    __IO uint32_t dev_id;                               /*  0x0004          I2C Slave Device ID Register                                                 */
    __IO uint32_t intfl;                                /*  0x0008          I2CS Interrupt Flags                                                         */
    __IO uint32_t inten;                                /*  0x000C          I2CS Interrupt Enable/Disable Controls                                       */
    __IO uint32_t data_byte[32];                        /*  0x0010-0x008C   I2CS Data Byte                                                               */
} mxc_i2cs_regs_t;


/*
   Register offsets for module I2CS.
*/

#define MXC_R_I2CS_OFFS_CLK_DIV                             ((uint32_t)0x00000000UL)
#define MXC_R_I2CS_OFFS_DEV_ID                              ((uint32_t)0x00000004UL)
#define MXC_R_I2CS_OFFS_INTFL                               ((uint32_t)0x00000008UL)
#define MXC_R_I2CS_OFFS_INTEN                               ((uint32_t)0x0000000CUL)
#define MXC_R_I2CS_OFFS_DATA_BYTE                           ((uint32_t)0x00000010UL)


/*
   Field positions and masks for module I2CS.
*/

#define MXC_F_I2CS_CLK_DIV_FS_FILTER_CLOCK_DIV_POS          0
#define MXC_F_I2CS_CLK_DIV_FS_FILTER_CLOCK_DIV              ((uint32_t)(0x000000FFUL << MXC_F_I2CS_CLK_DIV_FS_FILTER_CLOCK_DIV_POS))

#define MXC_F_I2CS_DEV_ID_SLAVE_DEV_ID_POS                  0
#define MXC_F_I2CS_DEV_ID_SLAVE_DEV_ID                      ((uint32_t)(0x000003FFUL << MXC_F_I2CS_DEV_ID_SLAVE_DEV_ID_POS))
#define MXC_F_I2CS_DEV_ID_TEN_BIT_ID_MODE_POS               12
#define MXC_F_I2CS_DEV_ID_TEN_BIT_ID_MODE                   ((uint32_t)(0x00000001UL << MXC_F_I2CS_DEV_ID_TEN_BIT_ID_MODE_POS))
#define MXC_F_I2CS_DEV_ID_SLAVE_RESET_POS                   14
#define MXC_F_I2CS_DEV_ID_SLAVE_RESET                       ((uint32_t)(0x00000001UL << MXC_F_I2CS_DEV_ID_SLAVE_RESET_POS))

#define MXC_F_I2CS_INTFL_BYTE0_POS                          0
#define MXC_F_I2CS_INTFL_BYTE0                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE0_POS))
#define MXC_F_I2CS_INTFL_BYTE1_POS                          1
#define MXC_F_I2CS_INTFL_BYTE1                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE1_POS))
#define MXC_F_I2CS_INTFL_BYTE2_POS                          2
#define MXC_F_I2CS_INTFL_BYTE2                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE2_POS))
#define MXC_F_I2CS_INTFL_BYTE3_POS                          3
#define MXC_F_I2CS_INTFL_BYTE3                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE3_POS))
#define MXC_F_I2CS_INTFL_BYTE4_POS                          4
#define MXC_F_I2CS_INTFL_BYTE4                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE4_POS))
#define MXC_F_I2CS_INTFL_BYTE5_POS                          5
#define MXC_F_I2CS_INTFL_BYTE5                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE5_POS))
#define MXC_F_I2CS_INTFL_BYTE6_POS                          6
#define MXC_F_I2CS_INTFL_BYTE6                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE6_POS))
#define MXC_F_I2CS_INTFL_BYTE7_POS                          7
#define MXC_F_I2CS_INTFL_BYTE7                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE7_POS))
#define MXC_F_I2CS_INTFL_BYTE8_POS                          8
#define MXC_F_I2CS_INTFL_BYTE8                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE8_POS))
#define MXC_F_I2CS_INTFL_BYTE9_POS                          9
#define MXC_F_I2CS_INTFL_BYTE9                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE9_POS))
#define MXC_F_I2CS_INTFL_BYTE10_POS                         10
#define MXC_F_I2CS_INTFL_BYTE10                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE10_POS))
#define MXC_F_I2CS_INTFL_BYTE11_POS                         11
#define MXC_F_I2CS_INTFL_BYTE11                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE11_POS))
#define MXC_F_I2CS_INTFL_BYTE12_POS                         12
#define MXC_F_I2CS_INTFL_BYTE12                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE12_POS))
#define MXC_F_I2CS_INTFL_BYTE13_POS                         13
#define MXC_F_I2CS_INTFL_BYTE13                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE13_POS))
#define MXC_F_I2CS_INTFL_BYTE14_POS                         14
#define MXC_F_I2CS_INTFL_BYTE14                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE14_POS))
#define MXC_F_I2CS_INTFL_BYTE15_POS                         15
#define MXC_F_I2CS_INTFL_BYTE15                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE15_POS))
#define MXC_F_I2CS_INTFL_BYTE16_POS                         16
#define MXC_F_I2CS_INTFL_BYTE16                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE16_POS))
#define MXC_F_I2CS_INTFL_BYTE17_POS                         17
#define MXC_F_I2CS_INTFL_BYTE17                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE17_POS))
#define MXC_F_I2CS_INTFL_BYTE18_POS                         18
#define MXC_F_I2CS_INTFL_BYTE18                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE18_POS))
#define MXC_F_I2CS_INTFL_BYTE19_POS                         19
#define MXC_F_I2CS_INTFL_BYTE19                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE19_POS))
#define MXC_F_I2CS_INTFL_BYTE20_POS                         20
#define MXC_F_I2CS_INTFL_BYTE20                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE20_POS))
#define MXC_F_I2CS_INTFL_BYTE21_POS                         21
#define MXC_F_I2CS_INTFL_BYTE21                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE21_POS))
#define MXC_F_I2CS_INTFL_BYTE22_POS                         22
#define MXC_F_I2CS_INTFL_BYTE22                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE22_POS))
#define MXC_F_I2CS_INTFL_BYTE23_POS                         23
#define MXC_F_I2CS_INTFL_BYTE23                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE23_POS))
#define MXC_F_I2CS_INTFL_BYTE24_POS                         24
#define MXC_F_I2CS_INTFL_BYTE24                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE24_POS))
#define MXC_F_I2CS_INTFL_BYTE25_POS                         25
#define MXC_F_I2CS_INTFL_BYTE25                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE25_POS))
#define MXC_F_I2CS_INTFL_BYTE26_POS                         26
#define MXC_F_I2CS_INTFL_BYTE26                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE26_POS))
#define MXC_F_I2CS_INTFL_BYTE27_POS                         27
#define MXC_F_I2CS_INTFL_BYTE27                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE27_POS))
#define MXC_F_I2CS_INTFL_BYTE28_POS                         28
#define MXC_F_I2CS_INTFL_BYTE28                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE28_POS))
#define MXC_F_I2CS_INTFL_BYTE29_POS                         29
#define MXC_F_I2CS_INTFL_BYTE29                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE29_POS))
#define MXC_F_I2CS_INTFL_BYTE30_POS                         30
#define MXC_F_I2CS_INTFL_BYTE30                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE30_POS))
#define MXC_F_I2CS_INTFL_BYTE31_POS                         31
#define MXC_F_I2CS_INTFL_BYTE31                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTFL_BYTE31_POS))

#define MXC_F_I2CS_INTEN_BYTE0_POS                          0
#define MXC_F_I2CS_INTEN_BYTE0                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE0_POS))
#define MXC_F_I2CS_INTEN_BYTE1_POS                          1
#define MXC_F_I2CS_INTEN_BYTE1                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE1_POS))
#define MXC_F_I2CS_INTEN_BYTE2_POS                          2
#define MXC_F_I2CS_INTEN_BYTE2                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE2_POS))
#define MXC_F_I2CS_INTEN_BYTE3_POS                          3
#define MXC_F_I2CS_INTEN_BYTE3                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE3_POS))
#define MXC_F_I2CS_INTEN_BYTE4_POS                          4
#define MXC_F_I2CS_INTEN_BYTE4                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE4_POS))
#define MXC_F_I2CS_INTEN_BYTE5_POS                          5
#define MXC_F_I2CS_INTEN_BYTE5                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE5_POS))
#define MXC_F_I2CS_INTEN_BYTE6_POS                          6
#define MXC_F_I2CS_INTEN_BYTE6                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE6_POS))
#define MXC_F_I2CS_INTEN_BYTE7_POS                          7
#define MXC_F_I2CS_INTEN_BYTE7                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE7_POS))
#define MXC_F_I2CS_INTEN_BYTE8_POS                          8
#define MXC_F_I2CS_INTEN_BYTE8                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE8_POS))
#define MXC_F_I2CS_INTEN_BYTE9_POS                          9
#define MXC_F_I2CS_INTEN_BYTE9                              ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE9_POS))
#define MXC_F_I2CS_INTEN_BYTE10_POS                         10
#define MXC_F_I2CS_INTEN_BYTE10                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE10_POS))
#define MXC_F_I2CS_INTEN_BYTE11_POS                         11
#define MXC_F_I2CS_INTEN_BYTE11                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE11_POS))
#define MXC_F_I2CS_INTEN_BYTE12_POS                         12
#define MXC_F_I2CS_INTEN_BYTE12                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE12_POS))
#define MXC_F_I2CS_INTEN_BYTE13_POS                         13
#define MXC_F_I2CS_INTEN_BYTE13                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE13_POS))
#define MXC_F_I2CS_INTEN_BYTE14_POS                         14
#define MXC_F_I2CS_INTEN_BYTE14                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE14_POS))
#define MXC_F_I2CS_INTEN_BYTE15_POS                         15
#define MXC_F_I2CS_INTEN_BYTE15                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE15_POS))
#define MXC_F_I2CS_INTEN_BYTE16_POS                         16
#define MXC_F_I2CS_INTEN_BYTE16                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE16_POS))
#define MXC_F_I2CS_INTEN_BYTE17_POS                         17
#define MXC_F_I2CS_INTEN_BYTE17                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE17_POS))
#define MXC_F_I2CS_INTEN_BYTE18_POS                         18
#define MXC_F_I2CS_INTEN_BYTE18                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE18_POS))
#define MXC_F_I2CS_INTEN_BYTE19_POS                         19
#define MXC_F_I2CS_INTEN_BYTE19                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE19_POS))
#define MXC_F_I2CS_INTEN_BYTE20_POS                         20
#define MXC_F_I2CS_INTEN_BYTE20                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE20_POS))
#define MXC_F_I2CS_INTEN_BYTE21_POS                         21
#define MXC_F_I2CS_INTEN_BYTE21                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE21_POS))
#define MXC_F_I2CS_INTEN_BYTE22_POS                         22
#define MXC_F_I2CS_INTEN_BYTE22                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE22_POS))
#define MXC_F_I2CS_INTEN_BYTE23_POS                         23
#define MXC_F_I2CS_INTEN_BYTE23                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE23_POS))
#define MXC_F_I2CS_INTEN_BYTE24_POS                         24
#define MXC_F_I2CS_INTEN_BYTE24                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE24_POS))
#define MXC_F_I2CS_INTEN_BYTE25_POS                         25
#define MXC_F_I2CS_INTEN_BYTE25                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE25_POS))
#define MXC_F_I2CS_INTEN_BYTE26_POS                         26
#define MXC_F_I2CS_INTEN_BYTE26                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE26_POS))
#define MXC_F_I2CS_INTEN_BYTE27_POS                         27
#define MXC_F_I2CS_INTEN_BYTE27                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE27_POS))
#define MXC_F_I2CS_INTEN_BYTE28_POS                         28
#define MXC_F_I2CS_INTEN_BYTE28                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE28_POS))
#define MXC_F_I2CS_INTEN_BYTE29_POS                         29
#define MXC_F_I2CS_INTEN_BYTE29                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE29_POS))
#define MXC_F_I2CS_INTEN_BYTE30_POS                         30
#define MXC_F_I2CS_INTEN_BYTE30                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE30_POS))
#define MXC_F_I2CS_INTEN_BYTE31_POS                         31
#define MXC_F_I2CS_INTEN_BYTE31                             ((uint32_t)(0x00000001UL << MXC_F_I2CS_INTEN_BYTE31_POS))

#define MXC_F_I2CS_DATA_BYTE_DATA_FIELD_POS                 0
#define MXC_F_I2CS_DATA_BYTE_DATA_FIELD                     ((uint32_t)(0x000000FFUL << MXC_F_I2CS_DATA_BYTE_DATA_FIELD_POS))
#define MXC_F_I2CS_DATA_BYTE_READ_ONLY_FL_POS               8
#define MXC_F_I2CS_DATA_BYTE_READ_ONLY_FL                   ((uint32_t)(0x00000001UL << MXC_F_I2CS_DATA_BYTE_READ_ONLY_FL_POS))
#define MXC_F_I2CS_DATA_BYTE_DATA_UPDATED_FL_POS            9
#define MXC_F_I2CS_DATA_BYTE_DATA_UPDATED_FL                ((uint32_t)(0x00000001UL << MXC_F_I2CS_DATA_BYTE_DATA_UPDATED_FL_POS))



#ifdef __cplusplus
}
#endif

#endif   /* _MXC_I2CS_REGS_H_ */

