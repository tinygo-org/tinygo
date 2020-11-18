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

#ifndef _MXC_TMR_REGS_H_
#define _MXC_TMR_REGS_H_

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
    __IO uint32_t ctrl;                                 /*  0x0000          Timer Control Register                                                       */
    __IO uint32_t count32;                              /*  0x0004          Timer [32 bit] Current Count Value                                           */
    __IO uint32_t term_cnt32;                           /*  0x0008          Timer [32 bit] Terminal Count Setting                                        */
    __IO uint32_t pwm_cap32;                            /*  0x000C          Timer [32 bit] PWM Compare Setting or Capture/Measure Value                  */
    __IO uint32_t count16_0;                            /*  0x0010          Timer [16 bit] Current Count Value, 16-bit Timer 0                           */
    __IO uint32_t term_cnt16_0;                         /*  0x0014          Timer [16 bit] Terminal Count Setting, 16-bit Timer 0                        */
    __IO uint32_t count16_1;                            /*  0x0018          Timer [16 bit] Current Count Value, 16-bit Timer 1                           */
    __IO uint32_t term_cnt16_1;                         /*  0x001C          Timer [16 bit] Terminal Count Setting, 16-bit Timer 1                        */
    __IO uint32_t intfl;                                /*  0x0020          Timer Interrupt Flags                                                        */
    __IO uint32_t inten;                                /*  0x0024          Timer Interrupt Enable/Disable Settings                                      */
} mxc_tmr_regs_t;


/*
   Register offsets for module TMR.
*/

#define MXC_R_TMR_OFFS_CTRL                                 ((uint32_t)0x00000000UL)
#define MXC_R_TMR_OFFS_COUNT32                              ((uint32_t)0x00000004UL)
#define MXC_R_TMR_OFFS_TERM_CNT32                           ((uint32_t)0x00000008UL)
#define MXC_R_TMR_OFFS_PWM_CAP32                            ((uint32_t)0x0000000CUL)
#define MXC_R_TMR_OFFS_COUNT16_0                            ((uint32_t)0x00000010UL)
#define MXC_R_TMR_OFFS_TERM_CNT16_0                         ((uint32_t)0x00000014UL)
#define MXC_R_TMR_OFFS_COUNT16_1                            ((uint32_t)0x00000018UL)
#define MXC_R_TMR_OFFS_TERM_CNT16_1                         ((uint32_t)0x0000001CUL)
#define MXC_R_TMR_OFFS_INTFL                                ((uint32_t)0x00000020UL)
#define MXC_R_TMR_OFFS_INTEN                                ((uint32_t)0x00000024UL)


/*
   Field positions and masks for module TMR.
*/

#define MXC_F_TMR_CTRL_MODE_POS                             0
#define MXC_F_TMR_CTRL_MODE                                 ((uint32_t)(0x00000007UL << MXC_F_TMR_CTRL_MODE_POS))
#define MXC_F_TMR_CTRL_TMR2X16_POS                          3
#define MXC_F_TMR_CTRL_TMR2X16                              ((uint32_t)(0x00000001UL << MXC_F_TMR_CTRL_TMR2X16_POS))
#define MXC_F_TMR_CTRL_PRESCALE_POS                         4
#define MXC_F_TMR_CTRL_PRESCALE                             ((uint32_t)(0x0000000FUL << MXC_F_TMR_CTRL_PRESCALE_POS))
#define MXC_F_TMR_CTRL_POLARITY_POS                         8
#define MXC_F_TMR_CTRL_POLARITY                             ((uint32_t)(0x00000001UL << MXC_F_TMR_CTRL_POLARITY_POS))
#define MXC_F_TMR_CTRL_ENABLE0_POS                          12
#define MXC_F_TMR_CTRL_ENABLE0                              ((uint32_t)(0x00000001UL << MXC_F_TMR_CTRL_ENABLE0_POS))
#define MXC_F_TMR_CTRL_ENABLE1_POS                          13
#define MXC_F_TMR_CTRL_ENABLE1                              ((uint32_t)(0x00000001UL << MXC_F_TMR_CTRL_ENABLE1_POS))

#define MXC_F_TMR_COUNT16_0_VALUE_POS                       0
#define MXC_F_TMR_COUNT16_0_VALUE                           ((uint32_t)(0x0000FFFFUL << MXC_F_TMR_COUNT16_0_VALUE_POS))

#define MXC_F_TMR_TERM_CNT16_0_TERM_COUNT_POS               0
#define MXC_F_TMR_TERM_CNT16_0_TERM_COUNT                   ((uint32_t)(0x0000FFFFUL << MXC_F_TMR_TERM_CNT16_0_TERM_COUNT_POS))

#define MXC_F_TMR_COUNT16_1_VALUE_POS                       0
#define MXC_F_TMR_COUNT16_1_VALUE                           ((uint32_t)(0x0000FFFFUL << MXC_F_TMR_COUNT16_1_VALUE_POS))

#define MXC_F_TMR_TERM_CNT16_1_TERM_COUNT_POS               0
#define MXC_F_TMR_TERM_CNT16_1_TERM_COUNT                   ((uint32_t)(0x0000FFFFUL << MXC_F_TMR_TERM_CNT16_1_TERM_COUNT_POS))

#define MXC_F_TMR_INTFL_TIMER0_POS                          0
#define MXC_F_TMR_INTFL_TIMER0                              ((uint32_t)(0x00000001UL << MXC_F_TMR_INTFL_TIMER0_POS))
#define MXC_F_TMR_INTFL_TIMER1_POS                          1
#define MXC_F_TMR_INTFL_TIMER1                              ((uint32_t)(0x00000001UL << MXC_F_TMR_INTFL_TIMER1_POS))

#define MXC_F_TMR_INTEN_TIMER0_POS                          0
#define MXC_F_TMR_INTEN_TIMER0                              ((uint32_t)(0x00000001UL << MXC_F_TMR_INTEN_TIMER0_POS))
#define MXC_F_TMR_INTEN_TIMER1_POS                          1
#define MXC_F_TMR_INTEN_TIMER1                              ((uint32_t)(0x00000001UL << MXC_F_TMR_INTEN_TIMER1_POS))



/*
   Field values and shifted values for module TMR.
*/

#define MXC_V_TMR_CTRL_MODE_ONE_SHOT                                            ((uint32_t)(0x00000000UL))
#define MXC_V_TMR_CTRL_MODE_CONTINUOUS                                          ((uint32_t)(0x00000001UL))
#define MXC_V_TMR_CTRL_MODE_COUNTER                                             ((uint32_t)(0x00000002UL))
#define MXC_V_TMR_CTRL_MODE_PWM                                                 ((uint32_t)(0x00000003UL))
#define MXC_V_TMR_CTRL_MODE_CAPTURE                                             ((uint32_t)(0x00000004UL))
#define MXC_V_TMR_CTRL_MODE_COMPARE                                             ((uint32_t)(0x00000005UL))
#define MXC_V_TMR_CTRL_MODE_GATED                                               ((uint32_t)(0x00000006UL))
#define MXC_V_TMR_CTRL_MODE_MEASURE                                             ((uint32_t)(0x00000007UL))

#define MXC_S_TMR_CTRL_MODE_ONE_SHOT                                            ((uint32_t)(MXC_V_TMR_CTRL_MODE_ONE_SHOT    << MXC_F_TMR_CTRL_MODE_POS))
#define MXC_S_TMR_CTRL_MODE_CONTINUOUS                                          ((uint32_t)(MXC_V_TMR_CTRL_MODE_CONTINUOUS  << MXC_F_TMR_CTRL_MODE_POS))
#define MXC_S_TMR_CTRL_MODE_COUNTER                                             ((uint32_t)(MXC_V_TMR_CTRL_MODE_COUNTER     << MXC_F_TMR_CTRL_MODE_POS))
#define MXC_S_TMR_CTRL_MODE_PWM                                                 ((uint32_t)(MXC_V_TMR_CTRL_MODE_PWM         << MXC_F_TMR_CTRL_MODE_POS))
#define MXC_S_TMR_CTRL_MODE_CAPTURE                                             ((uint32_t)(MXC_V_TMR_CTRL_MODE_CAPTURE     << MXC_F_TMR_CTRL_MODE_POS))
#define MXC_S_TMR_CTRL_MODE_COMPARE                                             ((uint32_t)(MXC_V_TMR_CTRL_MODE_COMPARE     << MXC_F_TMR_CTRL_MODE_POS))
#define MXC_S_TMR_CTRL_MODE_GATED                                               ((uint32_t)(MXC_V_TMR_CTRL_MODE_GATED       << MXC_F_TMR_CTRL_MODE_POS))
#define MXC_S_TMR_CTRL_MODE_MEASURE                                             ((uint32_t)(MXC_V_TMR_CTRL_MODE_MEASURE     << MXC_F_TMR_CTRL_MODE_POS))

#define MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_1                                     ((uint32_t)(0x00000000UL))
#define MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_2                                     ((uint32_t)(0x00000001UL))
#define MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_4                                     ((uint32_t)(0x00000002UL))
#define MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_8                                     ((uint32_t)(0x00000003UL))
#define MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_16                                    ((uint32_t)(0x00000004UL))
#define MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_32                                    ((uint32_t)(0x00000005UL))
#define MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_64                                    ((uint32_t)(0x00000006UL))
#define MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_128                                   ((uint32_t)(0x00000007UL))
#define MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_256                                   ((uint32_t)(0x00000008UL))
#define MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_512                                   ((uint32_t)(0x00000009UL))
#define MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_1024                                  ((uint32_t)(0x0000000AUL))
#define MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_2048                                  ((uint32_t)(0x0000000BUL))
#define MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_4096                                  ((uint32_t)(0x0000000CUL))

#define MXC_S_TMR_CTRL_PRESCALE_DIVIDE_BY_1                                     ((uint32_t)(MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_1      << MXC_F_TMR_CTRL_PRESCALE_POS))
#define MXC_S_TMR_CTRL_PRESCALE_DIVIDE_BY_2                                     ((uint32_t)(MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_2      << MXC_F_TMR_CTRL_PRESCALE_POS))
#define MXC_S_TMR_CTRL_PRESCALE_DIVIDE_BY_4                                     ((uint32_t)(MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_4      << MXC_F_TMR_CTRL_PRESCALE_POS))
#define MXC_S_TMR_CTRL_PRESCALE_DIVIDE_BY_8                                     ((uint32_t)(MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_8      << MXC_F_TMR_CTRL_PRESCALE_POS))
#define MXC_S_TMR_CTRL_PRESCALE_DIVIDE_BY_16                                    ((uint32_t)(MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_16     << MXC_F_TMR_CTRL_PRESCALE_POS))
#define MXC_S_TMR_CTRL_PRESCALE_DIVIDE_BY_32                                    ((uint32_t)(MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_32     << MXC_F_TMR_CTRL_PRESCALE_POS))
#define MXC_S_TMR_CTRL_PRESCALE_DIVIDE_BY_64                                    ((uint32_t)(MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_64     << MXC_F_TMR_CTRL_PRESCALE_POS))
#define MXC_S_TMR_CTRL_PRESCALE_DIVIDE_BY_128                                   ((uint32_t)(MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_128    << MXC_F_TMR_CTRL_PRESCALE_POS))
#define MXC_S_TMR_CTRL_PRESCALE_DIVIDE_BY_256                                   ((uint32_t)(MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_256    << MXC_F_TMR_CTRL_PRESCALE_POS))
#define MXC_S_TMR_CTRL_PRESCALE_DIVIDE_BY_512                                   ((uint32_t)(MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_512    << MXC_F_TMR_CTRL_PRESCALE_POS))
#define MXC_S_TMR_CTRL_PRESCALE_DIVIDE_BY_1024                                  ((uint32_t)(MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_1024   << MXC_F_TMR_CTRL_PRESCALE_POS))
#define MXC_S_TMR_CTRL_PRESCALE_DIVIDE_BY_2048                                  ((uint32_t)(MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_2048   << MXC_F_TMR_CTRL_PRESCALE_POS))
#define MXC_S_TMR_CTRL_PRESCALE_DIVIDE_BY_4096                                  ((uint32_t)(MXC_V_TMR_CTRL_PRESCALE_DIVIDE_BY_4096   << MXC_F_TMR_CTRL_PRESCALE_POS))


/*
 *  These two 1-bit fields replace the standard 3-bit mode field when the associated TMR module
 *  is in dual 16-bit timer mode.
 */

#define MXC_F_TMR_CTRL_MODE_16_0_POS     0
#define MXC_F_TMR_CTRL_MODE_16_0         ((uint32_t)(0x00000001UL << MXC_F_TMR_CTRL_MODE_16_0_POS))

#define MXC_F_TMR_CTRL_MODE_16_1_POS     1
#define MXC_F_TMR_CTRL_MODE_16_1         ((uint32_t)(0x00000001UL << MXC_F_TMR_CTRL_MODE_16_1_POS))


#ifdef __cplusplus
}
#endif

#endif   /* _MXC_TMR_REGS_H_ */

