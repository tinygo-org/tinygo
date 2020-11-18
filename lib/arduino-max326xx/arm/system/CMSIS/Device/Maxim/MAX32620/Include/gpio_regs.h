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

#ifndef _MXC_GPIO_REGS_H_
#define _MXC_GPIO_REGS_H_

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
    __IO uint32_t rst_mode[16];                          /*  0x0000-0x003C   Port P[0..15] Default (Power-On Reset) Output Drive Mode                      */
    __IO uint32_t free[16];                              /*  0x0040-0x007C   Port P[0..15] Free for GPIO Operation Flags                                   */
    __IO uint32_t out_mode[16];                          /*  0x0080-0x00BC   Port P[0..15] Output Drive Mode                                               */
    __IO uint32_t out_val[16];                           /*  0x00C0-0x00FC   Port P[0..15] GPIO Output Value                                               */
    __IO uint32_t func_sel[16];                          /*  0x0100-0x013C   Port P[0..15] GPIO Function Select                                            */
    __IO uint32_t in_mode[16];                           /*  0x0140-0x017C   Port P[0..15] GPIO Input Monitoring Mode                                      */
    __IO uint32_t in_val[16];                            /*  0x0180-0x01BC   Port P[0..15] GPIO Input Value                                                */
    __IO uint32_t int_mode[16];                          /*  0x01C0-0x01FC   Port P[0..15] Interrupt Detection Mode                                        */
    __IO uint32_t intfl[16];                             /*  0x0200-0x023C   Port P[0..15] Interrupt Flags                                                 */
    __IO uint32_t inten[16];                             /*  0x0240-0x027C   Port P[0..15] Interrupt Enables                                               */
} mxc_gpio_regs_t;


/*
   Register offsets for module GPIO.
*/

#define MXC_R_GPIO_OFFS_RST_MODE_P0                         ((uint32_t)0x00000000UL)
#define MXC_R_GPIO_OFFS_RST_MODE_P1                         ((uint32_t)0x00000004UL)
#define MXC_R_GPIO_OFFS_RST_MODE_P2                         ((uint32_t)0x00000008UL)
#define MXC_R_GPIO_OFFS_RST_MODE_P3                         ((uint32_t)0x0000000CUL)
#define MXC_R_GPIO_OFFS_RST_MODE_P4                         ((uint32_t)0x00000010UL)
#define MXC_R_GPIO_OFFS_RST_MODE_P5                         ((uint32_t)0x00000014UL)
#define MXC_R_GPIO_OFFS_RST_MODE_P6                         ((uint32_t)0x00000018UL)
#define MXC_R_GPIO_OFFS_RST_MODE_P7                         ((uint32_t)0x0000001CUL)
#define MXC_R_GPIO_OFFS_RST_MODE_P8                         ((uint32_t)0x00000020UL)
#define MXC_R_GPIO_OFFS_RST_MODE_P9                         ((uint32_t)0x00000024UL)
#define MXC_R_GPIO_OFFS_RST_MODE_P10                        ((uint32_t)0x00000028UL)
#define MXC_R_GPIO_OFFS_RST_MODE_P11                        ((uint32_t)0x0000002CUL)
#define MXC_R_GPIO_OFFS_RST_MODE_P12                        ((uint32_t)0x00000030UL)
#define MXC_R_GPIO_OFFS_RST_MODE_P13                        ((uint32_t)0x00000034UL)
#define MXC_R_GPIO_OFFS_RST_MODE_P14                        ((uint32_t)0x00000038UL)
#define MXC_R_GPIO_OFFS_RST_MODE_P15                        ((uint32_t)0x0000003CUL)
#define MXC_R_GPIO_OFFS_FREE_P0                             ((uint32_t)0x00000040UL)
#define MXC_R_GPIO_OFFS_FREE_P1                             ((uint32_t)0x00000044UL)
#define MXC_R_GPIO_OFFS_FREE_P2                             ((uint32_t)0x00000048UL)
#define MXC_R_GPIO_OFFS_FREE_P3                             ((uint32_t)0x0000004CUL)
#define MXC_R_GPIO_OFFS_FREE_P4                             ((uint32_t)0x00000050UL)
#define MXC_R_GPIO_OFFS_FREE_P5                             ((uint32_t)0x00000054UL)
#define MXC_R_GPIO_OFFS_FREE_P6                             ((uint32_t)0x00000058UL)
#define MXC_R_GPIO_OFFS_FREE_P7                             ((uint32_t)0x0000005CUL)
#define MXC_R_GPIO_OFFS_FREE_P8                             ((uint32_t)0x00000060UL)
#define MXC_R_GPIO_OFFS_FREE_P9                             ((uint32_t)0x00000064UL)
#define MXC_R_GPIO_OFFS_FREE_P10                            ((uint32_t)0x00000068UL)
#define MXC_R_GPIO_OFFS_FREE_P11                            ((uint32_t)0x0000006CUL)
#define MXC_R_GPIO_OFFS_FREE_P12                            ((uint32_t)0x00000070UL)
#define MXC_R_GPIO_OFFS_FREE_P13                            ((uint32_t)0x00000074UL)
#define MXC_R_GPIO_OFFS_FREE_P14                            ((uint32_t)0x00000078UL)
#define MXC_R_GPIO_OFFS_FREE_P15                            ((uint32_t)0x0000007CUL)
#define MXC_R_GPIO_OFFS_OUT_MODE_P0                         ((uint32_t)0x00000080UL)
#define MXC_R_GPIO_OFFS_OUT_MODE_P1                         ((uint32_t)0x00000084UL)
#define MXC_R_GPIO_OFFS_OUT_MODE_P2                         ((uint32_t)0x00000088UL)
#define MXC_R_GPIO_OFFS_OUT_MODE_P3                         ((uint32_t)0x0000008CUL)
#define MXC_R_GPIO_OFFS_OUT_MODE_P4                         ((uint32_t)0x00000090UL)
#define MXC_R_GPIO_OFFS_OUT_MODE_P5                         ((uint32_t)0x00000094UL)
#define MXC_R_GPIO_OFFS_OUT_MODE_P6                         ((uint32_t)0x00000098UL)
#define MXC_R_GPIO_OFFS_OUT_MODE_P7                         ((uint32_t)0x0000009CUL)
#define MXC_R_GPIO_OFFS_OUT_MODE_P8                         ((uint32_t)0x000000A0UL)
#define MXC_R_GPIO_OFFS_OUT_MODE_P9                         ((uint32_t)0x000000A4UL)
#define MXC_R_GPIO_OFFS_OUT_MODE_P10                        ((uint32_t)0x000000A8UL)
#define MXC_R_GPIO_OFFS_OUT_MODE_P11                        ((uint32_t)0x000000ACUL)
#define MXC_R_GPIO_OFFS_OUT_MODE_P12                        ((uint32_t)0x000000B0UL)
#define MXC_R_GPIO_OFFS_OUT_MODE_P13                        ((uint32_t)0x000000B4UL)
#define MXC_R_GPIO_OFFS_OUT_MODE_P14                        ((uint32_t)0x000000B8UL)
#define MXC_R_GPIO_OFFS_OUT_MODE_P15                        ((uint32_t)0x000000BCUL)
#define MXC_R_GPIO_OFFS_OUT_VAL_P0                          ((uint32_t)0x000000C0UL)
#define MXC_R_GPIO_OFFS_OUT_VAL_P1                          ((uint32_t)0x000000C4UL)
#define MXC_R_GPIO_OFFS_OUT_VAL_P2                          ((uint32_t)0x000000C8UL)
#define MXC_R_GPIO_OFFS_OUT_VAL_P3                          ((uint32_t)0x000000CCUL)
#define MXC_R_GPIO_OFFS_OUT_VAL_P4                          ((uint32_t)0x000000D0UL)
#define MXC_R_GPIO_OFFS_OUT_VAL_P5                          ((uint32_t)0x000000D4UL)
#define MXC_R_GPIO_OFFS_OUT_VAL_P6                          ((uint32_t)0x000000D8UL)
#define MXC_R_GPIO_OFFS_OUT_VAL_P7                          ((uint32_t)0x000000DCUL)
#define MXC_R_GPIO_OFFS_OUT_VAL_P8                          ((uint32_t)0x000000E0UL)
#define MXC_R_GPIO_OFFS_OUT_VAL_P9                          ((uint32_t)0x000000E4UL)
#define MXC_R_GPIO_OFFS_OUT_VAL_P10                         ((uint32_t)0x000000E8UL)
#define MXC_R_GPIO_OFFS_OUT_VAL_P11                         ((uint32_t)0x000000ECUL)
#define MXC_R_GPIO_OFFS_OUT_VAL_P12                         ((uint32_t)0x000000F0UL)
#define MXC_R_GPIO_OFFS_OUT_VAL_P13                         ((uint32_t)0x000000F4UL)
#define MXC_R_GPIO_OFFS_OUT_VAL_P14                         ((uint32_t)0x000000F8UL)
#define MXC_R_GPIO_OFFS_OUT_VAL_P15                         ((uint32_t)0x000000FCUL)
#define MXC_R_GPIO_OFFS_FUNC_SEL_P0                         ((uint32_t)0x00000100UL)
#define MXC_R_GPIO_OFFS_FUNC_SEL_P1                         ((uint32_t)0x00000104UL)
#define MXC_R_GPIO_OFFS_FUNC_SEL_P2                         ((uint32_t)0x00000108UL)
#define MXC_R_GPIO_OFFS_FUNC_SEL_P3                         ((uint32_t)0x0000010CUL)
#define MXC_R_GPIO_OFFS_FUNC_SEL_P4                         ((uint32_t)0x00000110UL)
#define MXC_R_GPIO_OFFS_FUNC_SEL_P5                         ((uint32_t)0x00000114UL)
#define MXC_R_GPIO_OFFS_FUNC_SEL_P6                         ((uint32_t)0x00000118UL)
#define MXC_R_GPIO_OFFS_FUNC_SEL_P7                         ((uint32_t)0x0000011CUL)
#define MXC_R_GPIO_OFFS_FUNC_SEL_P8                         ((uint32_t)0x00000120UL)
#define MXC_R_GPIO_OFFS_FUNC_SEL_P9                         ((uint32_t)0x00000124UL)
#define MXC_R_GPIO_OFFS_FUNC_SEL_P10                        ((uint32_t)0x00000128UL)
#define MXC_R_GPIO_OFFS_FUNC_SEL_P11                        ((uint32_t)0x0000012CUL)
#define MXC_R_GPIO_OFFS_FUNC_SEL_P12                        ((uint32_t)0x00000130UL)
#define MXC_R_GPIO_OFFS_FUNC_SEL_P13                        ((uint32_t)0x00000134UL)
#define MXC_R_GPIO_OFFS_FUNC_SEL_P14                        ((uint32_t)0x00000138UL)
#define MXC_R_GPIO_OFFS_FUNC_SEL_P15                        ((uint32_t)0x0000013CUL)
#define MXC_R_GPIO_OFFS_IN_MODE_P0                          ((uint32_t)0x00000140UL)
#define MXC_R_GPIO_OFFS_IN_MODE_P1                          ((uint32_t)0x00000144UL)
#define MXC_R_GPIO_OFFS_IN_MODE_P2                          ((uint32_t)0x00000148UL)
#define MXC_R_GPIO_OFFS_IN_MODE_P3                          ((uint32_t)0x0000014CUL)
#define MXC_R_GPIO_OFFS_IN_MODE_P4                          ((uint32_t)0x00000150UL)
#define MXC_R_GPIO_OFFS_IN_MODE_P5                          ((uint32_t)0x00000154UL)
#define MXC_R_GPIO_OFFS_IN_MODE_P6                          ((uint32_t)0x00000158UL)
#define MXC_R_GPIO_OFFS_IN_MODE_P7                          ((uint32_t)0x0000015CUL)
#define MXC_R_GPIO_OFFS_IN_MODE_P8                          ((uint32_t)0x00000160UL)
#define MXC_R_GPIO_OFFS_IN_MODE_P9                          ((uint32_t)0x00000164UL)
#define MXC_R_GPIO_OFFS_IN_MODE_P10                         ((uint32_t)0x00000168UL)
#define MXC_R_GPIO_OFFS_IN_MODE_P11                         ((uint32_t)0x0000016CUL)
#define MXC_R_GPIO_OFFS_IN_MODE_P12                         ((uint32_t)0x00000170UL)
#define MXC_R_GPIO_OFFS_IN_MODE_P13                         ((uint32_t)0x00000174UL)
#define MXC_R_GPIO_OFFS_IN_MODE_P14                         ((uint32_t)0x00000178UL)
#define MXC_R_GPIO_OFFS_IN_MODE_P15                         ((uint32_t)0x0000017CUL)
#define MXC_R_GPIO_OFFS_IN_VAL_P0                           ((uint32_t)0x00000180UL)
#define MXC_R_GPIO_OFFS_IN_VAL_P1                           ((uint32_t)0x00000184UL)
#define MXC_R_GPIO_OFFS_IN_VAL_P2                           ((uint32_t)0x00000188UL)
#define MXC_R_GPIO_OFFS_IN_VAL_P3                           ((uint32_t)0x0000018CUL)
#define MXC_R_GPIO_OFFS_IN_VAL_P4                           ((uint32_t)0x00000190UL)
#define MXC_R_GPIO_OFFS_IN_VAL_P5                           ((uint32_t)0x00000194UL)
#define MXC_R_GPIO_OFFS_IN_VAL_P6                           ((uint32_t)0x00000198UL)
#define MXC_R_GPIO_OFFS_IN_VAL_P7                           ((uint32_t)0x0000019CUL)
#define MXC_R_GPIO_OFFS_IN_VAL_P8                           ((uint32_t)0x000001A0UL)
#define MXC_R_GPIO_OFFS_IN_VAL_P9                           ((uint32_t)0x000001A4UL)
#define MXC_R_GPIO_OFFS_IN_VAL_P10                          ((uint32_t)0x000001A8UL)
#define MXC_R_GPIO_OFFS_IN_VAL_P11                          ((uint32_t)0x000001ACUL)
#define MXC_R_GPIO_OFFS_IN_VAL_P12                          ((uint32_t)0x000001B0UL)
#define MXC_R_GPIO_OFFS_IN_VAL_P13                          ((uint32_t)0x000001B4UL)
#define MXC_R_GPIO_OFFS_IN_VAL_P14                          ((uint32_t)0x000001B8UL)
#define MXC_R_GPIO_OFFS_IN_VAL_P15                          ((uint32_t)0x000001BCUL)
#define MXC_R_GPIO_OFFS_INT_MODE_P0                         ((uint32_t)0x000001C0UL)
#define MXC_R_GPIO_OFFS_INT_MODE_P1                         ((uint32_t)0x000001C4UL)
#define MXC_R_GPIO_OFFS_INT_MODE_P2                         ((uint32_t)0x000001C8UL)
#define MXC_R_GPIO_OFFS_INT_MODE_P3                         ((uint32_t)0x000001CCUL)
#define MXC_R_GPIO_OFFS_INT_MODE_P4                         ((uint32_t)0x000001D0UL)
#define MXC_R_GPIO_OFFS_INT_MODE_P5                         ((uint32_t)0x000001D4UL)
#define MXC_R_GPIO_OFFS_INT_MODE_P6                         ((uint32_t)0x000001D8UL)
#define MXC_R_GPIO_OFFS_INT_MODE_P7                         ((uint32_t)0x000001DCUL)
#define MXC_R_GPIO_OFFS_INT_MODE_P8                         ((uint32_t)0x000001E0UL)
#define MXC_R_GPIO_OFFS_INT_MODE_P9                         ((uint32_t)0x000001E4UL)
#define MXC_R_GPIO_OFFS_INT_MODE_P10                        ((uint32_t)0x000001E8UL)
#define MXC_R_GPIO_OFFS_INT_MODE_P11                        ((uint32_t)0x000001ECUL)
#define MXC_R_GPIO_OFFS_INT_MODE_P12                        ((uint32_t)0x000001F0UL)
#define MXC_R_GPIO_OFFS_INT_MODE_P13                        ((uint32_t)0x000001F4UL)
#define MXC_R_GPIO_OFFS_INT_MODE_P14                        ((uint32_t)0x000001F8UL)
#define MXC_R_GPIO_OFFS_INT_MODE_P15                        ((uint32_t)0x000001FCUL)
#define MXC_R_GPIO_OFFS_INTFL_P0                            ((uint32_t)0x00000200UL)
#define MXC_R_GPIO_OFFS_INTFL_P1                            ((uint32_t)0x00000204UL)
#define MXC_R_GPIO_OFFS_INTFL_P2                            ((uint32_t)0x00000208UL)
#define MXC_R_GPIO_OFFS_INTFL_P3                            ((uint32_t)0x0000020CUL)
#define MXC_R_GPIO_OFFS_INTFL_P4                            ((uint32_t)0x00000210UL)
#define MXC_R_GPIO_OFFS_INTFL_P5                            ((uint32_t)0x00000214UL)
#define MXC_R_GPIO_OFFS_INTFL_P6                            ((uint32_t)0x00000218UL)
#define MXC_R_GPIO_OFFS_INTFL_P7                            ((uint32_t)0x0000021CUL)
#define MXC_R_GPIO_OFFS_INTFL_P8                            ((uint32_t)0x00000220UL)
#define MXC_R_GPIO_OFFS_INTFL_P9                            ((uint32_t)0x00000224UL)
#define MXC_R_GPIO_OFFS_INTFL_P10                           ((uint32_t)0x00000228UL)
#define MXC_R_GPIO_OFFS_INTFL_P11                           ((uint32_t)0x0000022CUL)
#define MXC_R_GPIO_OFFS_INTFL_P12                           ((uint32_t)0x00000230UL)
#define MXC_R_GPIO_OFFS_INTFL_P13                           ((uint32_t)0x00000234UL)
#define MXC_R_GPIO_OFFS_INTFL_P14                           ((uint32_t)0x00000238UL)
#define MXC_R_GPIO_OFFS_INTFL_P15                           ((uint32_t)0x0000023CUL)
#define MXC_R_GPIO_OFFS_INTEN_P0                            ((uint32_t)0x00000240UL)
#define MXC_R_GPIO_OFFS_INTEN_P1                            ((uint32_t)0x00000244UL)
#define MXC_R_GPIO_OFFS_INTEN_P2                            ((uint32_t)0x00000248UL)
#define MXC_R_GPIO_OFFS_INTEN_P3                            ((uint32_t)0x0000024CUL)
#define MXC_R_GPIO_OFFS_INTEN_P4                            ((uint32_t)0x00000250UL)
#define MXC_R_GPIO_OFFS_INTEN_P5                            ((uint32_t)0x00000254UL)
#define MXC_R_GPIO_OFFS_INTEN_P6                            ((uint32_t)0x00000258UL)
#define MXC_R_GPIO_OFFS_INTEN_P7                            ((uint32_t)0x0000025CUL)
#define MXC_R_GPIO_OFFS_INTEN_P8                            ((uint32_t)0x00000260UL)
#define MXC_R_GPIO_OFFS_INTEN_P9                            ((uint32_t)0x00000264UL)
#define MXC_R_GPIO_OFFS_INTEN_P10                           ((uint32_t)0x00000268UL)
#define MXC_R_GPIO_OFFS_INTEN_P11                           ((uint32_t)0x0000026CUL)
#define MXC_R_GPIO_OFFS_INTEN_P12                           ((uint32_t)0x00000270UL)
#define MXC_R_GPIO_OFFS_INTEN_P13                           ((uint32_t)0x00000274UL)
#define MXC_R_GPIO_OFFS_INTEN_P14                           ((uint32_t)0x00000278UL)
#define MXC_R_GPIO_OFFS_INTEN_P15                           ((uint32_t)0x0000027CUL)


/*
   Field positions and masks for module GPIO.
*/

#define MXC_F_GPIO_RST_MODE_PIN0_POS                        0
#define MXC_F_GPIO_RST_MODE_PIN0                            ((uint32_t)(0x00000007UL << MXC_F_GPIO_RST_MODE_PIN0_POS))
#define MXC_F_GPIO_RST_MODE_PIN1_POS                        4
#define MXC_F_GPIO_RST_MODE_PIN1                            ((uint32_t)(0x00000007UL << MXC_F_GPIO_RST_MODE_PIN1_POS))
#define MXC_F_GPIO_RST_MODE_PIN2_POS                        8
#define MXC_F_GPIO_RST_MODE_PIN2                            ((uint32_t)(0x00000007UL << MXC_F_GPIO_RST_MODE_PIN2_POS))
#define MXC_F_GPIO_RST_MODE_PIN3_POS                        12
#define MXC_F_GPIO_RST_MODE_PIN3                            ((uint32_t)(0x00000007UL << MXC_F_GPIO_RST_MODE_PIN3_POS))
#define MXC_F_GPIO_RST_MODE_PIN4_POS                        16
#define MXC_F_GPIO_RST_MODE_PIN4                            ((uint32_t)(0x00000007UL << MXC_F_GPIO_RST_MODE_PIN4_POS))
#define MXC_F_GPIO_RST_MODE_PIN5_POS                        20
#define MXC_F_GPIO_RST_MODE_PIN5                            ((uint32_t)(0x00000007UL << MXC_F_GPIO_RST_MODE_PIN5_POS))
#define MXC_F_GPIO_RST_MODE_PIN6_POS                        24
#define MXC_F_GPIO_RST_MODE_PIN6                            ((uint32_t)(0x00000007UL << MXC_F_GPIO_RST_MODE_PIN6_POS))
#define MXC_F_GPIO_RST_MODE_PIN7_POS                        28
#define MXC_F_GPIO_RST_MODE_PIN7                            ((uint32_t)(0x00000007UL << MXC_F_GPIO_RST_MODE_PIN7_POS))

#define MXC_F_GPIO_FREE_PIN0_POS                            0
#define MXC_F_GPIO_FREE_PIN0                                ((uint32_t)(0x00000001UL << MXC_F_GPIO_FREE_PIN0_POS))
#define MXC_F_GPIO_FREE_PIN1_POS                            1
#define MXC_F_GPIO_FREE_PIN1                                ((uint32_t)(0x00000001UL << MXC_F_GPIO_FREE_PIN1_POS))
#define MXC_F_GPIO_FREE_PIN2_POS                            2
#define MXC_F_GPIO_FREE_PIN2                                ((uint32_t)(0x00000001UL << MXC_F_GPIO_FREE_PIN2_POS))
#define MXC_F_GPIO_FREE_PIN3_POS                            3
#define MXC_F_GPIO_FREE_PIN3                                ((uint32_t)(0x00000001UL << MXC_F_GPIO_FREE_PIN3_POS))
#define MXC_F_GPIO_FREE_PIN4_POS                            4
#define MXC_F_GPIO_FREE_PIN4                                ((uint32_t)(0x00000001UL << MXC_F_GPIO_FREE_PIN4_POS))
#define MXC_F_GPIO_FREE_PIN5_POS                            5
#define MXC_F_GPIO_FREE_PIN5                                ((uint32_t)(0x00000001UL << MXC_F_GPIO_FREE_PIN5_POS))
#define MXC_F_GPIO_FREE_PIN6_POS                            6
#define MXC_F_GPIO_FREE_PIN6                                ((uint32_t)(0x00000001UL << MXC_F_GPIO_FREE_PIN6_POS))
#define MXC_F_GPIO_FREE_PIN7_POS                            7
#define MXC_F_GPIO_FREE_PIN7                                ((uint32_t)(0x00000001UL << MXC_F_GPIO_FREE_PIN7_POS))

#define MXC_F_GPIO_OUT_MODE_PIN0_POS                        0
#define MXC_F_GPIO_OUT_MODE_PIN0                            ((uint32_t)(0x0000000FUL << MXC_F_GPIO_OUT_MODE_PIN0_POS))
#define MXC_F_GPIO_OUT_MODE_PIN1_POS                        4
#define MXC_F_GPIO_OUT_MODE_PIN1                            ((uint32_t)(0x0000000FUL << MXC_F_GPIO_OUT_MODE_PIN1_POS))
#define MXC_F_GPIO_OUT_MODE_PIN2_POS                        8
#define MXC_F_GPIO_OUT_MODE_PIN2                            ((uint32_t)(0x0000000FUL << MXC_F_GPIO_OUT_MODE_PIN2_POS))
#define MXC_F_GPIO_OUT_MODE_PIN3_POS                        12
#define MXC_F_GPIO_OUT_MODE_PIN3                            ((uint32_t)(0x0000000FUL << MXC_F_GPIO_OUT_MODE_PIN3_POS))
#define MXC_F_GPIO_OUT_MODE_PIN4_POS                        16
#define MXC_F_GPIO_OUT_MODE_PIN4                            ((uint32_t)(0x0000000FUL << MXC_F_GPIO_OUT_MODE_PIN4_POS))
#define MXC_F_GPIO_OUT_MODE_PIN5_POS                        20
#define MXC_F_GPIO_OUT_MODE_PIN5                            ((uint32_t)(0x0000000FUL << MXC_F_GPIO_OUT_MODE_PIN5_POS))
#define MXC_F_GPIO_OUT_MODE_PIN6_POS                        24
#define MXC_F_GPIO_OUT_MODE_PIN6                            ((uint32_t)(0x0000000FUL << MXC_F_GPIO_OUT_MODE_PIN6_POS))
#define MXC_F_GPIO_OUT_MODE_PIN7_POS                        28
#define MXC_F_GPIO_OUT_MODE_PIN7                            ((uint32_t)(0x0000000FUL << MXC_F_GPIO_OUT_MODE_PIN7_POS))

#define MXC_F_GPIO_OUT_VAL_PIN0_POS                         0
#define MXC_F_GPIO_OUT_VAL_PIN0                             ((uint32_t)(0x00000001UL << MXC_F_GPIO_OUT_VAL_PIN0_POS))
#define MXC_F_GPIO_OUT_VAL_PIN1_POS                         1
#define MXC_F_GPIO_OUT_VAL_PIN1                             ((uint32_t)(0x00000001UL << MXC_F_GPIO_OUT_VAL_PIN1_POS))
#define MXC_F_GPIO_OUT_VAL_PIN2_POS                         2
#define MXC_F_GPIO_OUT_VAL_PIN2                             ((uint32_t)(0x00000001UL << MXC_F_GPIO_OUT_VAL_PIN2_POS))
#define MXC_F_GPIO_OUT_VAL_PIN3_POS                         3
#define MXC_F_GPIO_OUT_VAL_PIN3                             ((uint32_t)(0x00000001UL << MXC_F_GPIO_OUT_VAL_PIN3_POS))
#define MXC_F_GPIO_OUT_VAL_PIN4_POS                         4
#define MXC_F_GPIO_OUT_VAL_PIN4                             ((uint32_t)(0x00000001UL << MXC_F_GPIO_OUT_VAL_PIN4_POS))
#define MXC_F_GPIO_OUT_VAL_PIN5_POS                         5
#define MXC_F_GPIO_OUT_VAL_PIN5                             ((uint32_t)(0x00000001UL << MXC_F_GPIO_OUT_VAL_PIN5_POS))
#define MXC_F_GPIO_OUT_VAL_PIN6_POS                         6
#define MXC_F_GPIO_OUT_VAL_PIN6                             ((uint32_t)(0x00000001UL << MXC_F_GPIO_OUT_VAL_PIN6_POS))
#define MXC_F_GPIO_OUT_VAL_PIN7_POS                         7
#define MXC_F_GPIO_OUT_VAL_PIN7                             ((uint32_t)(0x00000001UL << MXC_F_GPIO_OUT_VAL_PIN7_POS))

#define MXC_F_GPIO_FUNC_SEL_PIN0_POS                        0
#define MXC_F_GPIO_FUNC_SEL_PIN0                            ((uint32_t)(0x0000000FUL << MXC_F_GPIO_FUNC_SEL_PIN0_POS))
#define MXC_F_GPIO_FUNC_SEL_PIN1_POS                        4
#define MXC_F_GPIO_FUNC_SEL_PIN1                            ((uint32_t)(0x0000000FUL << MXC_F_GPIO_FUNC_SEL_PIN1_POS))
#define MXC_F_GPIO_FUNC_SEL_PIN2_POS                        8
#define MXC_F_GPIO_FUNC_SEL_PIN2                            ((uint32_t)(0x0000000FUL << MXC_F_GPIO_FUNC_SEL_PIN2_POS))
#define MXC_F_GPIO_FUNC_SEL_PIN3_POS                        12
#define MXC_F_GPIO_FUNC_SEL_PIN3                            ((uint32_t)(0x0000000FUL << MXC_F_GPIO_FUNC_SEL_PIN3_POS))
#define MXC_F_GPIO_FUNC_SEL_PIN4_POS                        16
#define MXC_F_GPIO_FUNC_SEL_PIN4                            ((uint32_t)(0x0000000FUL << MXC_F_GPIO_FUNC_SEL_PIN4_POS))
#define MXC_F_GPIO_FUNC_SEL_PIN5_POS                        20
#define MXC_F_GPIO_FUNC_SEL_PIN5                            ((uint32_t)(0x0000000FUL << MXC_F_GPIO_FUNC_SEL_PIN5_POS))
#define MXC_F_GPIO_FUNC_SEL_PIN6_POS                        24
#define MXC_F_GPIO_FUNC_SEL_PIN6                            ((uint32_t)(0x0000000FUL << MXC_F_GPIO_FUNC_SEL_PIN6_POS))
#define MXC_F_GPIO_FUNC_SEL_PIN7_POS                        28
#define MXC_F_GPIO_FUNC_SEL_PIN7                            ((uint32_t)(0x0000000FUL << MXC_F_GPIO_FUNC_SEL_PIN7_POS))

#define MXC_F_GPIO_IN_MODE_PIN0_POS                         0
#define MXC_F_GPIO_IN_MODE_PIN0                             ((uint32_t)(0x00000003UL << MXC_F_GPIO_IN_MODE_PIN0_POS))
#define MXC_F_GPIO_IN_MODE_PIN1_POS                         4
#define MXC_F_GPIO_IN_MODE_PIN1                             ((uint32_t)(0x00000003UL << MXC_F_GPIO_IN_MODE_PIN1_POS))
#define MXC_F_GPIO_IN_MODE_PIN2_POS                         8
#define MXC_F_GPIO_IN_MODE_PIN2                             ((uint32_t)(0x00000003UL << MXC_F_GPIO_IN_MODE_PIN2_POS))
#define MXC_F_GPIO_IN_MODE_PIN3_POS                         12
#define MXC_F_GPIO_IN_MODE_PIN3                             ((uint32_t)(0x00000003UL << MXC_F_GPIO_IN_MODE_PIN3_POS))
#define MXC_F_GPIO_IN_MODE_PIN4_POS                         16
#define MXC_F_GPIO_IN_MODE_PIN4                             ((uint32_t)(0x00000003UL << MXC_F_GPIO_IN_MODE_PIN4_POS))
#define MXC_F_GPIO_IN_MODE_PIN5_POS                         20
#define MXC_F_GPIO_IN_MODE_PIN5                             ((uint32_t)(0x00000003UL << MXC_F_GPIO_IN_MODE_PIN5_POS))
#define MXC_F_GPIO_IN_MODE_PIN6_POS                         24
#define MXC_F_GPIO_IN_MODE_PIN6                             ((uint32_t)(0x00000003UL << MXC_F_GPIO_IN_MODE_PIN6_POS))
#define MXC_F_GPIO_IN_MODE_PIN7_POS                         28
#define MXC_F_GPIO_IN_MODE_PIN7                             ((uint32_t)(0x00000003UL << MXC_F_GPIO_IN_MODE_PIN7_POS))

#define MXC_F_GPIO_IN_VAL_PIN0_POS                          0
#define MXC_F_GPIO_IN_VAL_PIN0                              ((uint32_t)(0x00000001UL << MXC_F_GPIO_IN_VAL_PIN0_POS))
#define MXC_F_GPIO_IN_VAL_PIN1_POS                          1
#define MXC_F_GPIO_IN_VAL_PIN1                              ((uint32_t)(0x00000001UL << MXC_F_GPIO_IN_VAL_PIN1_POS))
#define MXC_F_GPIO_IN_VAL_PIN2_POS                          2
#define MXC_F_GPIO_IN_VAL_PIN2                              ((uint32_t)(0x00000001UL << MXC_F_GPIO_IN_VAL_PIN2_POS))
#define MXC_F_GPIO_IN_VAL_PIN3_POS                          3
#define MXC_F_GPIO_IN_VAL_PIN3                              ((uint32_t)(0x00000001UL << MXC_F_GPIO_IN_VAL_PIN3_POS))
#define MXC_F_GPIO_IN_VAL_PIN4_POS                          4
#define MXC_F_GPIO_IN_VAL_PIN4                              ((uint32_t)(0x00000001UL << MXC_F_GPIO_IN_VAL_PIN4_POS))
#define MXC_F_GPIO_IN_VAL_PIN5_POS                          5
#define MXC_F_GPIO_IN_VAL_PIN5                              ((uint32_t)(0x00000001UL << MXC_F_GPIO_IN_VAL_PIN5_POS))
#define MXC_F_GPIO_IN_VAL_PIN6_POS                          6
#define MXC_F_GPIO_IN_VAL_PIN6                              ((uint32_t)(0x00000001UL << MXC_F_GPIO_IN_VAL_PIN6_POS))
#define MXC_F_GPIO_IN_VAL_PIN7_POS                          7
#define MXC_F_GPIO_IN_VAL_PIN7                              ((uint32_t)(0x00000001UL << MXC_F_GPIO_IN_VAL_PIN7_POS))

#define MXC_F_GPIO_INT_MODE_PIN0_POS                        0
#define MXC_F_GPIO_INT_MODE_PIN0                            ((uint32_t)(0x00000007UL << MXC_F_GPIO_INT_MODE_PIN0_POS))
#define MXC_F_GPIO_INT_MODE_PIN1_POS                        4
#define MXC_F_GPIO_INT_MODE_PIN1                            ((uint32_t)(0x00000007UL << MXC_F_GPIO_INT_MODE_PIN1_POS))
#define MXC_F_GPIO_INT_MODE_PIN2_POS                        8
#define MXC_F_GPIO_INT_MODE_PIN2                            ((uint32_t)(0x00000007UL << MXC_F_GPIO_INT_MODE_PIN2_POS))
#define MXC_F_GPIO_INT_MODE_PIN3_POS                        12
#define MXC_F_GPIO_INT_MODE_PIN3                            ((uint32_t)(0x00000007UL << MXC_F_GPIO_INT_MODE_PIN3_POS))
#define MXC_F_GPIO_INT_MODE_PIN4_POS                        16
#define MXC_F_GPIO_INT_MODE_PIN4                            ((uint32_t)(0x00000007UL << MXC_F_GPIO_INT_MODE_PIN4_POS))
#define MXC_F_GPIO_INT_MODE_PIN5_POS                        20
#define MXC_F_GPIO_INT_MODE_PIN5                            ((uint32_t)(0x00000007UL << MXC_F_GPIO_INT_MODE_PIN5_POS))
#define MXC_F_GPIO_INT_MODE_PIN6_POS                        24
#define MXC_F_GPIO_INT_MODE_PIN6                            ((uint32_t)(0x00000007UL << MXC_F_GPIO_INT_MODE_PIN6_POS))
#define MXC_F_GPIO_INT_MODE_PIN7_POS                        28
#define MXC_F_GPIO_INT_MODE_PIN7                            ((uint32_t)(0x00000007UL << MXC_F_GPIO_INT_MODE_PIN7_POS))

#define MXC_F_GPIO_INTFL_PIN0_POS                           0
#define MXC_F_GPIO_INTFL_PIN0                               ((uint32_t)(0x00000001UL << MXC_F_GPIO_INTFL_PIN0_POS))
#define MXC_F_GPIO_INTFL_PIN1_POS                           1
#define MXC_F_GPIO_INTFL_PIN1                               ((uint32_t)(0x00000001UL << MXC_F_GPIO_INTFL_PIN1_POS))
#define MXC_F_GPIO_INTFL_PIN2_POS                           2
#define MXC_F_GPIO_INTFL_PIN2                               ((uint32_t)(0x00000001UL << MXC_F_GPIO_INTFL_PIN2_POS))
#define MXC_F_GPIO_INTFL_PIN3_POS                           3
#define MXC_F_GPIO_INTFL_PIN3                               ((uint32_t)(0x00000001UL << MXC_F_GPIO_INTFL_PIN3_POS))
#define MXC_F_GPIO_INTFL_PIN4_POS                           4
#define MXC_F_GPIO_INTFL_PIN4                               ((uint32_t)(0x00000001UL << MXC_F_GPIO_INTFL_PIN4_POS))
#define MXC_F_GPIO_INTFL_PIN5_POS                           5
#define MXC_F_GPIO_INTFL_PIN5                               ((uint32_t)(0x00000001UL << MXC_F_GPIO_INTFL_PIN5_POS))
#define MXC_F_GPIO_INTFL_PIN6_POS                           6
#define MXC_F_GPIO_INTFL_PIN6                               ((uint32_t)(0x00000001UL << MXC_F_GPIO_INTFL_PIN6_POS))
#define MXC_F_GPIO_INTFL_PIN7_POS                           7
#define MXC_F_GPIO_INTFL_PIN7                               ((uint32_t)(0x00000001UL << MXC_F_GPIO_INTFL_PIN7_POS))

#define MXC_F_GPIO_INTEN_PIN0_POS                           0
#define MXC_F_GPIO_INTEN_PIN0                               ((uint32_t)(0x00000001UL << MXC_F_GPIO_INTEN_PIN0_POS))
#define MXC_F_GPIO_INTEN_PIN1_POS                           1
#define MXC_F_GPIO_INTEN_PIN1                               ((uint32_t)(0x00000001UL << MXC_F_GPIO_INTEN_PIN1_POS))
#define MXC_F_GPIO_INTEN_PIN2_POS                           2
#define MXC_F_GPIO_INTEN_PIN2                               ((uint32_t)(0x00000001UL << MXC_F_GPIO_INTEN_PIN2_POS))
#define MXC_F_GPIO_INTEN_PIN3_POS                           3
#define MXC_F_GPIO_INTEN_PIN3                               ((uint32_t)(0x00000001UL << MXC_F_GPIO_INTEN_PIN3_POS))
#define MXC_F_GPIO_INTEN_PIN4_POS                           4
#define MXC_F_GPIO_INTEN_PIN4                               ((uint32_t)(0x00000001UL << MXC_F_GPIO_INTEN_PIN4_POS))
#define MXC_F_GPIO_INTEN_PIN5_POS                           5
#define MXC_F_GPIO_INTEN_PIN5                               ((uint32_t)(0x00000001UL << MXC_F_GPIO_INTEN_PIN5_POS))
#define MXC_F_GPIO_INTEN_PIN6_POS                           6
#define MXC_F_GPIO_INTEN_PIN6                               ((uint32_t)(0x00000001UL << MXC_F_GPIO_INTEN_PIN6_POS))
#define MXC_F_GPIO_INTEN_PIN7_POS                           7
#define MXC_F_GPIO_INTEN_PIN7                               ((uint32_t)(0x00000001UL << MXC_F_GPIO_INTEN_PIN7_POS))



/*
   Field values and shifted values for module GPIO.
*/

#define MXC_V_GPIO_RST_MODE_DRIVE_0                                             ((uint32_t)(0x00000000UL))
#define MXC_V_GPIO_RST_MODE_WEAK_PULLDOWN                                       ((uint32_t)(0x00000001UL))
#define MXC_V_GPIO_RST_MODE_WEAK_PULLUP                                         ((uint32_t)(0x00000002UL))
#define MXC_V_GPIO_RST_MODE_DRIVE_1                                             ((uint32_t)(0x00000003UL))
#define MXC_V_GPIO_RST_MODE_HIGH_Z                                              ((uint32_t)(0x00000004UL))

#define MXC_V_GPIO_FREE_NOT_AVAILABLE                                           ((uint32_t)(0x00000000UL))
#define MXC_V_GPIO_FREE_AVAILABLE                                               ((uint32_t)(0x00000001UL))

#define MXC_V_GPIO_OUT_MODE_HIGH_Z_WEAK_PULLUP                                  ((uint32_t)(0x00000000UL))
#define MXC_V_GPIO_OUT_MODE_OPEN_DRAIN                                          ((uint32_t)(0x00000001UL))
#define MXC_V_GPIO_OUT_MODE_OPEN_DRAIN_WEAK_PULLUP                              ((uint32_t)(0x00000002UL))
#define MXC_V_GPIO_OUT_MODE_NORMAL_HIGH_Z                                       ((uint32_t)(0x00000004UL))
#define MXC_V_GPIO_OUT_MODE_NORMAL                                              ((uint32_t)(0x00000005UL))
#define MXC_V_GPIO_OUT_MODE_SLOW_HIGH_Z                                         ((uint32_t)(0x00000006UL))
#define MXC_V_GPIO_OUT_MODE_SLOW_DRIVE                                          ((uint32_t)(0x00000007UL))
#define MXC_V_GPIO_OUT_MODE_FAST_HIGH_Z                                         ((uint32_t)(0x00000008UL))
#define MXC_V_GPIO_OUT_MODE_FAST_DRIVE                                          ((uint32_t)(0x00000009UL))
#define MXC_V_GPIO_OUT_MODE_HIGH_Z_WEAK_PULLDOWN                                ((uint32_t)(0x0000000AUL))
#define MXC_V_GPIO_OUT_MODE_OPEN_SOURCE                                         ((uint32_t)(0x0000000BUL))
#define MXC_V_GPIO_OUT_MODE_OPEN_SOURCE_WEAK_PULLDOWN                           ((uint32_t)(0x0000000CUL))
#define MXC_V_GPIO_OUT_MODE_HIGH_Z_INPUT_DISABLED                               ((uint32_t)(0x0000000FUL))

#define MXC_V_GPIO_FUNC_SEL_MODE_GPIO                                           ((uint32_t)(0x00000000UL))
#define MXC_V_GPIO_FUNC_SEL_MODE_PT                                             ((uint32_t)(0x00000001UL))
#define MXC_V_GPIO_FUNC_SEL_MODE_TMR                                            ((uint32_t)(0x00000002UL))

#define MXC_V_GPIO_IN_MODE_NORMAL                                               ((uint32_t)(0x00000000UL))
#define MXC_V_GPIO_IN_MODE_INVERTED                                             ((uint32_t)(0x00000001UL))
#define MXC_V_GPIO_IN_MODE_ALWAYS_ZERO                                          ((uint32_t)(0x00000002UL))
#define MXC_V_GPIO_IN_MODE_ALWAYS_ONE                                           ((uint32_t)(0x00000003UL))

#define MXC_V_GPIO_INT_MODE_DISABLE                                             ((uint32_t)(0x00000000UL))
#define MXC_V_GPIO_INT_MODE_FALLING_EDGE                                        ((uint32_t)(0x00000001UL))
#define MXC_V_GPIO_INT_MODE_RISING_EDGE                                         ((uint32_t)(0x00000002UL))
#define MXC_V_GPIO_INT_MODE_ANY_EDGE                                            ((uint32_t)(0x00000003UL))
#define MXC_V_GPIO_INT_MODE_LOW_LVL                                             ((uint32_t)(0x00000004UL))
#define MXC_V_GPIO_INT_MODE_HIGH_LVL                                            ((uint32_t)(0x00000005UL))



#ifdef __cplusplus
}
#endif

#endif   /* _MXC_GPIO_REGS_H_ */

