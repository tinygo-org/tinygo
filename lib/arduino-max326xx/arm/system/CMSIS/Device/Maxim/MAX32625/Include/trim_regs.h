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
 * $Date: 2016-04-20 16:18:27 -0500 (Wed, 20 Apr 2016) $
 * $Revision: 22447 $
 *
 ******************************************************************************/

#ifndef _MXC_TRIM_REGS_H_
#define _MXC_TRIM_REGS_H_

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
    __R  uint32_t rsv000[10];                           /*  0x0000-0x0024                                                                                */
    __IO uint32_t reg10_mem_size;                       /*  0x0028          Shadow Trim for Flash and SRAM Memory Size                                   */
    __IO uint32_t reg11_adc_trim0;                      /*  0x002C          Shadow Trim for ADC R0                                                       */
    __IO uint32_t reg12_adc_trim1;                      /*  0x0030          Shadow Trim for ADC R1                                                       */
    __IO uint32_t for_pwr_reg5;                         /*  0x0034          Shadow Trim for PWRSEQ Register REG5                                         */
    __IO uint32_t for_pwr_reg6;                         /*  0x0038          Shadow Trim for PWRSEQ Register REG6                                         */
    __IO uint32_t for_pwr_reg7;                         /*  0x003C          Shadow Trim for PWRSEQ Register REG7                                         */
} mxc_trim_regs_t;


/*
   Register offsets for module TRIM.
*/

#define MXC_R_TRIM_OFFS_REG10_MEM_SIZE                      ((uint32_t)0x00000028UL)
#define MXC_R_TRIM_OFFS_REG11_ADC_TRIM0                     ((uint32_t)0x0000002CUL)
#define MXC_R_TRIM_OFFS_REG12_ADC_TRIM1                     ((uint32_t)0x00000030UL)
#define MXC_R_TRIM_OFFS_FOR_PWR_REG5                        ((uint32_t)0x00000034UL)
#define MXC_R_TRIM_OFFS_FOR_PWR_REG6                        ((uint32_t)0x00000038UL)
#define MXC_R_TRIM_OFFS_FOR_PWR_REG7                        ((uint32_t)0x0000003CUL)


/*
   Field positions and masks for module TRIM.
*/

#define MXC_F_TRIM_REG10_MEM_SIZE_SRAM_POS                  0
#define MXC_F_TRIM_REG10_MEM_SIZE_SRAM                      ((uint32_t)(0x00000003UL << MXC_F_TRIM_REG10_MEM_SIZE_SRAM_POS))
#define MXC_F_TRIM_REG10_MEM_SIZE_FLASH_POS                 2
#define MXC_F_TRIM_REG10_MEM_SIZE_FLASH                     ((uint32_t)(0x00000007UL << MXC_F_TRIM_REG10_MEM_SIZE_FLASH_POS))

#define MXC_V_TRIM_REG10_MEM_SRAM_FULL_SIZE                 ((uint32_t)(0x00000000UL))
#define MXC_V_TRIM_REG10_MEM_SRAM_THREE_FOURTHS_SIZE        ((uint32_t)(0x00000001UL))
#define MXC_V_TRIM_REG10_MEM_SRAM_HALF_SIZE                 ((uint32_t)(0x00000002UL))

#define MXC_V_TRIM_REG10_MEM_FLASH_FULL_SIZE                ((uint32_t)(0x00000000UL))
#define MXC_V_TRIM_REG10_MEM_FLASH_THREE_FOURTHS_SIZE       ((uint32_t)(0x00000001UL))
#define MXC_V_TRIM_REG10_MEM_FLASH_HALF_SIZE                ((uint32_t)(0x00000002UL))
#define MXC_V_TRIM_REG10_MEM_FLASH_THREE_EIGHTHS_SIZE       ((uint32_t)(0x00000003UL))
#define MXC_V_TRIM_REG10_MEM_FLASH_FOURTH_SIZE              ((uint32_t)(0x00000004UL))

#define MXC_F_TRIM_REG11_ADC_TRIM0_ADCTRIM_X0R0_POS         0
#define MXC_F_TRIM_REG11_ADC_TRIM0_ADCTRIM_X0R0             ((uint32_t)(0x000003FFUL << MXC_F_TRIM_REG11_ADC_TRIM0_ADCTRIM_X0R0_POS))
#define MXC_F_TRIM_REG11_ADC_TRIM0_ADCTRIM_X1R0_POS         16
#define MXC_F_TRIM_REG11_ADC_TRIM0_ADCTRIM_X1R0             ((uint32_t)(0x000003FFUL << MXC_F_TRIM_REG11_ADC_TRIM0_ADCTRIM_X1R0_POS))

#define MXC_F_TRIM_REG12_ADC_TRIM1_ADCTRIM_X0R1_POS         0
#define MXC_F_TRIM_REG12_ADC_TRIM1_ADCTRIM_X0R1             ((uint32_t)(0x000003FFUL << MXC_F_TRIM_REG12_ADC_TRIM1_ADCTRIM_X0R1_POS))
#define MXC_F_TRIM_REG12_ADC_TRIM1_ADCTRIM_X1R1_POS         16
#define MXC_F_TRIM_REG12_ADC_TRIM1_ADCTRIM_X1R1             ((uint32_t)(0x000003FFUL << MXC_F_TRIM_REG12_ADC_TRIM1_ADCTRIM_X1R1_POS))
#define MXC_F_TRIM_REG12_ADC_TRIM1_ADCTRIM_DC_POS           28
#define MXC_F_TRIM_REG12_ADC_TRIM1_ADCTRIM_DC               ((uint32_t)(0x0000000FUL << MXC_F_TRIM_REG12_ADC_TRIM1_ADCTRIM_DC_POS))



#ifdef __cplusplus
}
#endif

#endif   /* _MXC_TRIM_REGS_H_ */

