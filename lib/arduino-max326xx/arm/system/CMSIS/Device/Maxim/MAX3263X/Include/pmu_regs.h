/**
 * @file
 * @brief   Registers, Bit Masks and Bit Positions for the PMU Module.
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
 * $Date: 2017-02-14 18:18:10 -0600 (Tue, 14 Feb 2017) $
 * $Revision: 26428 $
 *
 *************************************************************************** */

/* Define to prevent redundant inclusion */
#ifndef _MXC_PMU_REGS_H_
#define _MXC_PMU_REGS_H_

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

/*
   Typedefed structure(s) for module registers (per instance or section) with direct 32-bit
   access to each register in module.
*/
/**
 * @ingroup     pmuGroup
 * @defgroup    pmu_registers Registers
 * @brief       Registers, Bit Masks and Bit Positions for the PMU Module.
 * @{
 */
/**
 * Structure type for the PMU Registers 
 */
typedef struct {
    __IO uint32_t dscadr;                               /**<  <b><tt>0x0000:</tt></b> PMU Channel Next Descriptor Address     */
    __IO uint32_t cfg;                                  /**<  <b><tt>0x0004:</tt></b> PMU Channel Configuration               */
    __IO uint32_t loop;                                 /**<  <b><tt>0x0008:</tt></b> PMU Channel Loop Counters               */
    __R  uint32_t rsv00C[5];                            /**<  <b><tt>0x000C-0x001C:</tt></b>   RESERVED                       */
} mxc_pmu_regs_t;
/**@} end of group pmu_registers */

/*
   Register offsets for module PMU.
*/
/**
 * @ingroup    pmu_registers
 * @defgroup   PMU_Register_Offsets Register Offsets
 * @brief      PMU Register Offsets from the PMU Base Peripheral Address. 
 * @{
 */
#define MXC_R_PMU_OFFS_DSCADR                               ((uint32_t)0x00000000UL)  /**< Offset from the PMU Base Address: <b><tt>0x0000</tt></b>*/
#define MXC_R_PMU_OFFS_CFG                                  ((uint32_t)0x00000004UL)  /**< Offset from the PMU Base Address: <b><tt>0x0004</tt></b>*/
#define MXC_R_PMU_OFFS_LOOP                                 ((uint32_t)0x00000008UL)  /**< Offset from the PMU Base Address: <b><tt>0x0008</tt></b>*/
/**@} end of group PMU_Register_Offsets */

/*
   Field positions and masks for module PMU.
*/
///@cond
#define MXC_F_PMU_CFG_ENABLE_POS                            0
#define MXC_F_PMU_CFG_ENABLE                                ((uint32_t)(0x00000001UL << MXC_F_PMU_CFG_ENABLE_POS))
#define MXC_F_PMU_CFG_LL_STOPPED_POS                        2
#define MXC_F_PMU_CFG_LL_STOPPED                            ((uint32_t)(0x00000001UL << MXC_F_PMU_CFG_LL_STOPPED_POS))
#define MXC_F_PMU_CFG_MANUAL_POS                            3
#define MXC_F_PMU_CFG_MANUAL                                ((uint32_t)(0x00000001UL << MXC_F_PMU_CFG_MANUAL_POS))
#define MXC_F_PMU_CFG_BUS_ERROR_POS                         4
#define MXC_F_PMU_CFG_BUS_ERROR                             ((uint32_t)(0x00000001UL << MXC_F_PMU_CFG_BUS_ERROR_POS))
#define MXC_F_PMU_CFG_TO_STAT_POS                           6
#define MXC_F_PMU_CFG_TO_STAT                               ((uint32_t)(0x00000001UL << MXC_F_PMU_CFG_TO_STAT_POS))
#define MXC_F_PMU_CFG_TO_SEL_POS                            11
#define MXC_F_PMU_CFG_TO_SEL                                ((uint32_t)(0x00000007UL << MXC_F_PMU_CFG_TO_SEL_POS))
#define MXC_F_PMU_CFG_PS_SEL_POS                            14
#define MXC_F_PMU_CFG_PS_SEL                                ((uint32_t)(0x00000003UL << MXC_F_PMU_CFG_PS_SEL_POS))
#define MXC_F_PMU_CFG_INTERRUPT_POS                         16
#define MXC_F_PMU_CFG_INTERRUPT                             ((uint32_t)(0x00000001UL << MXC_F_PMU_CFG_INTERRUPT_POS))
#define MXC_F_PMU_CFG_INT_EN_POS                            17
#define MXC_F_PMU_CFG_INT_EN                                ((uint32_t)(0x00000001UL << MXC_F_PMU_CFG_INT_EN_POS))
#define MXC_F_PMU_CFG_BURST_SIZE_POS                        24
#define MXC_F_PMU_CFG_BURST_SIZE                            ((uint32_t)(0x0000001FUL << MXC_F_PMU_CFG_BURST_SIZE_POS))

#define MXC_F_PMU_LOOP_COUNTER_0_POS                        0
#define MXC_F_PMU_LOOP_COUNTER_0                            ((uint32_t)(0x0000FFFFUL << MXC_F_PMU_LOOP_COUNTER_0_POS))
#define MXC_F_PMU_LOOP_COUNTER_1_POS                        16
#define MXC_F_PMU_LOOP_COUNTER_1                            ((uint32_t)(0x0000FFFFUL << MXC_F_PMU_LOOP_COUNTER_1_POS))

/*
   Field values
*/

#define MXC_V_PMU_CFG_TO_SEL_TICKS_4                        ((uint32_t)(0x00000000UL))
#define MXC_V_PMU_CFG_TO_SEL_TICKS_8                        ((uint32_t)(0x00000001UL))
#define MXC_V_PMU_CFG_TO_SEL_TICKS_16                       ((uint32_t)(0x00000002UL))
#define MXC_V_PMU_CFG_TO_SEL_TICKS_32                       ((uint32_t)(0x00000003UL))
#define MXC_V_PMU_CFG_TO_SEL_TICKS_64                       ((uint32_t)(0x00000004UL))
#define MXC_V_PMU_CFG_TO_SEL_TICKS_128                      ((uint32_t)(0x00000005UL))
#define MXC_V_PMU_CFG_TO_SEL_TICKS_256                      ((uint32_t)(0x00000006UL))
#define MXC_V_PMU_CFG_TO_SEL_TICKS_512                      ((uint32_t)(0x00000007UL))

#define MXC_V_PMU_CFG_PS_SEL_DISABLE                        ((uint32_t)(0x00000000UL))
#define MXC_V_PMU_CFG_PS_SEL_DIV_2_8                        ((uint32_t)(0x00000001UL))
#define MXC_V_PMU_CFG_PS_SEL_DIV_2_16                       ((uint32_t)(0x00000002UL))
#define MXC_V_PMU_CFG_PS_SEL_DIV_2_24                       ((uint32_t)(0x00000003UL))

/* Op codes */
#define PMU_MOVE_OP                                         0
#define PMU_WRITE_OP                                        1
#define PMU_WAIT_OP                                         2
#define PMU_JUMP_OP                                         3
#define PMU_LOOP_OP                                         4
#define PMU_POLL_OP                                         5
#define PMU_BRANCH_OP                                       6
#define PMU_TRANSFER_OP                                     7

/* Bit values used in all decroptiors */
#define PMU_NO_INTERRUPT                                    0 /**< Interrupt flag is NOT set at end of channel execution */
#define PMU_INTERRUPT                                       1 /**< Interrupt flag is set at end of channel execution */

#define PMU_NO_STOP                                         0 /**< Do not stop channel after this descriptor ends */
#define PMU_STOP                                            1 /**< Halt PMU channel after this descriptor ends */

/* Interrupt and Stop bit positions */
#define PMU_INT_POS                                         3
#define PMU_STOP_POS                                        4

/* MOVE descriptor bit values */
#define PMU_MOVE_READ_8_BIT                                 0 /**< Read size = 8 */
#define PMU_MOVE_READ_16_BIT                                1 /**< Read size = 16  */
#define PMU_MOVE_READ_32_BIT                                2 /**< Read size = 32 */

#define PMU_MOVE_READ_NO_INC                                0 /**< read address not incremented */
#define PMU_MOVE_READ_INC                                   1 /**< Auto-Increment read address */

#define PMU_MOVE_WRITE_8_BIT                                0 /**< Write Size = 8 */
#define PMU_MOVE_WRITE_16_BIT                               1 /**< Write Size = 16 */
#define PMU_MOVE_WRITE_32_BIT                               2 /**< Write Size = 32 */

#define PMU_MOVE_WRITE_NO_INC                               0 /**< Write address not incremented */
#define PMU_MOVE_WRITE_INC                                  1 /**< Auto_Increment write address */

#define PMU_MOVE_NO_CONT                                    0 /**< MOVE does not rely on previous MOVE */
#define PMU_MOVE_CONT                                       1 /**< MOVE continues from read/write address and INC values defined in previous MOVE */

/* MOVE bit positions */
#define PMU_MOVE_READS_POS                                  5
#define PMU_MOVE_READI_POS                                  7
#define PMU_MOVE_WRITES_POS                                 8
#define PMU_MOVE_WRITEI_POS                                 10
#define PMU_MOVE_CONT_POS                                   11
#define PMU_MOVE_LEN_POS                                    12

/* WRITE descriptor bit values */
#define PMU_WRITE_MASKED_WRITE_VALUE                        0  /**< Value = READ_VALUE & (~WRITE_MASK) | WRITE_VALUE */
#define PMU_WRITE_PLUS_1                                    1  /**< Value = READ_VALUE + 1 */
#define PMU_WRITE_MINUS_1                                   2  /**< Value = READ_VALUE - 1 */
#define PMU_WRITE_SHIFT_RT_1                                3  /**< Value = READ_VALUE >> 1 */
#define PMU_WRITE_SHIFT_LT_1                                4  /**< Value = READ_VALUE << 1 */
#define PMU_WRITE_ROTATE_RT_1                               5  /**< Value = READ_VALUE rotated right by 1 (bit 0 becomes bit 31) */
#define PMU_WRITE_ROTATE_LT_1                               6  /**< Value = READ_VALUE rotated left by 1 (bit 31 becomes bit 0) */
#define PMU_WRITE_NOT_READ_VAL                              7  /**< Value = ~READ_VALUE */
#define PMU_WRITE_XOR_MASK                                  8  /**< Value = READ_VALUE XOR WRITE_MASK */
#define PMU_WRITE_OR_MASK                                   9  /**< Value = READ_VALUE | WRITE_MASK */
#define PMU_WRITE_AND_MASK                                  10 /**< Value = READ_VALUE & WRITE_MASK */

/* WRITE bit positions */
#define PMU_WRITE_METHOD_POS                                8

/* WAIT descriptor bit values */
#define PMU_WAIT_SEL_0                                      0 /**< Select the interrupt source */
#define PMU_WAIT_SEL_1                                      1

/* WAIT bit positions */
#define PMU_WAIT_WAIT_POS                                   5
#define PMU_WAIT_SEL_POS                                    6

/* LOOP descriptor bit values */
#define PMU_LOOP_SEL_COUNTER0                               0 /**< select Counter0 to count down from */
#define PMU_LOOP_SEL_COUNTER1                               1 /**< select Counter1 to count down from */

/* LOOP bit positions */
#define PMU_LOOP_SEL_COUNTER_POS                            5

/* POLL descriptor bit values */
#define PMU_POLL_OR                                         0 /**< polling ends when at least one mask bit matches expected data */
#define PMU_POLL_AND                                        1 /**< polling ends when all mask bits matches expected data */

/* POLL bit positions */
#define PMU_POLL_AND_POS                                    7

/* BRANCH descriptor bit values */
#define PMU_BRANCH_OR                                       0 /**< branch when any mask bit = or != expected data (based on = or != branch type) */
#define PMU_BRANCH_AND                                      1 /**< branch when all mask bit = or != expected data (based on = or != branch type) */

#define PMU_BRANCH_TYPE_NOT_EQUAL                           0 /**< Branch when polled data != expected data */
#define PMU_BRANCH_TYPE_EQUAL                               1 /**< Branch when polled data = expected data */
#define PMU_BRANCH_TYPE_LESS_OR_EQUAL                       2 /**< Branch when polled data <= expected data */
#define PMU_BRANCH_TYPE_GREAT_OR_EQUAL                      3 /**< Branch when polled data >= expected data */
#define PMU_BRANCH_TYPE_LESSER                              4 /**< Branch when polled data < expected data */
#define PMU_BRANCH_TYPE_GREATER                             5 /**< Branch when polled data > expected data */

/* BRANCH bit positions */
#define PMU_BRANCH_AND_POS                                  7
#define PMU_BRANCH_TYPE_POS                                 8

/* TRANSFER descriptor bit values */
#define PMU_TX_READ_8_BIT                                   0 /**< Read size = 8 */
#define PMU_TX_READ_16_BIT                                  1 /**< Read size = 16 */
#define PMU_TX_READ_32_BIT                                  2 /**< Read size = 32 */

#define PMU_TX_READ_NO_INC                                  0 /**< read address not incremented */
#define PMU_TX_READ_INC                                     1 /**< Auto-Increment read address */

#define PMU_TX_WRITE_8_BIT                                  0 /**< Write Size = 8 */
#define PMU_TX_WRITE_16_BIT                                 1 /**< Write Size = 16 */
#define PMU_TX_WRITE_32_BIT                                 2 /**< Write Size = 32 */

#define PMU_TX_WRITE_NO_INC                                 0 /**< Write address not incremented */
#define PMU_TX_WRITE_INC                                    1 /**< Auto_Increment write address */

/* TRANSFER bit positions */
#define PMU_TX_READS_POS                                    5
#define PMU_TX_READI_POS                                    7
#define PMU_TX_WRITES_POS                                   8
#define PMU_TX_WRITEI_POS                                   10
#define PMU_TX_LEN_POS                                      12
#define PMU_TX_BS_POS                                       26

/* PMU interrupt sources for the WAIT opcode */
#define PMU_WAIT_IRQ_MASK1_SEL0_UART0_TX_FIFO_AE            ((uint32_t)(0x00000001UL << 0))
#define PMU_WAIT_IRQ_MASK1_SEL0_UART0_RX_FIFO_AF            ((uint32_t)(0x00000001UL << 1))
#define PMU_WAIT_IRQ_MASK1_SEL0_UART1_TX_FIFO_AE            ((uint32_t)(0x00000001UL << 2))
#define PMU_WAIT_IRQ_MASK1_SEL0_UART1_RX_FIFO_AF            ((uint32_t)(0x00000001UL << 3))
#define PMU_WAIT_IRQ_MASK1_SEL0_UART2_TX_FIFO_AE            ((uint32_t)(0x00000001UL << 4))
#define PMU_WAIT_IRQ_MASK1_SEL0_UART2_RX_FIFO_AF            ((uint32_t)(0x00000001UL << 5))
#define PMU_WAIT_IRQ_MASK1_SEL0_UART3_TX_FIFO_AE            ((uint32_t)(0x00000001UL << 6))
#define PMU_WAIT_IRQ_MASK1_SEL0_UART3_RX_FIFO_AF            ((uint32_t)(0x00000001UL << 7))
#define PMU_WAIT_IRQ_MASK1_SEL0_SPI0_TX_FIFO_AE             ((uint32_t)(0x00000001UL << 8))
#define PMU_WAIT_IRQ_MASK1_SEL0_SPI0_RX_FIFO_AF             ((uint32_t)(0x00000001UL << 9))
#define PMU_WAIT_IRQ_MASK1_SEL0_SPI1_TX_FIFO_AE             ((uint32_t)(0x00000001UL << 10))
#define PMU_WAIT_IRQ_MASK1_SEL0_SPI1_RX_FIFO_AF             ((uint32_t)(0x00000001UL << 11))
#define PMU_WAIT_IRQ_MASK1_SEL0_SPI2_TX_FIFO_AE             ((uint32_t)(0x00000001UL << 12))
#define PMU_WAIT_IRQ_MASK1_SEL0_SPI2_RX_FIFO_AF             ((uint32_t)(0x00000001UL << 13))
#define PMU_WAIT_IRQ_MASK1_SEL0_I2CM0_TX_FIFO_EMPTY         ((uint32_t)(0x00000001UL << 14))
#define PMU_WAIT_IRQ_MASK1_SEL0_I2CM0_RX_FIFO_NOT_EMPTY     ((uint32_t)(0x00000001UL << 15))
#define PMU_WAIT_IRQ_MASK1_SEL0_I2CM1_TX_FIFO_EMPTY         ((uint32_t)(0x00000001UL << 16))
#define PMU_WAIT_IRQ_MASK1_SEL0_I2CM1_RX_FIFO_NOT_EMPTY     ((uint32_t)(0x00000001UL << 17))
#define PMU_WAIT_IRQ_MASK1_SEL0_I2CM2_TX_FIFO_EMPTY         ((uint32_t)(0x00000001UL << 18))
#define PMU_WAIT_IRQ_MASK1_SEL0_I2CM2_RX_FIFO_NOT_EMPTY     ((uint32_t)(0x00000001UL << 19))
#define PMU_WAIT_IRQ_MASK1_SEL0_SPI0_TX_RX_STALLED          ((uint32_t)(0x00000001UL << 20))
#define PMU_WAIT_IRQ_MASK1_SEL0_SPI1_TX_RX_STALLED          ((uint32_t)(0x00000001UL << 21))
#define PMU_WAIT_IRQ_MASK1_SEL0_SPI2_TX_RX_STALLED          ((uint32_t)(0x00000001UL << 22))
#define PMU_WAIT_IRQ_MASK1_SEL0_SPIB                        ((uint32_t)(0x00000001UL << 23))
#define PMU_WAIT_IRQ_MASK1_SEL0_I2CM0_DONE                  ((uint32_t)(0x00000001UL << 24))
#define PMU_WAIT_IRQ_MASK1_SEL0_I2CM1_DONE                  ((uint32_t)(0x00000001UL << 25))
#define PMU_WAIT_IRQ_MASK1_SEL0_I2CM2_DONE                  ((uint32_t)(0x00000001UL << 26))
#define PMU_WAIT_IRQ_MASK1_SEL0_I2CS                        ((uint32_t)(0x00000001UL << 27))
#define PMU_WAIT_IRQ_MASK1_SEL0_ADC_DONE                    ((uint32_t)(0x00000001UL << 28))
#define PMU_WAIT_IRQ_MASK1_SEL0_ADC_READY                   ((uint32_t)(0x00000001UL << 29))
#define PMU_WAIT_IRQ_MASK1_SEL0_ADC_HI                      ((uint32_t)(0x00000001UL << 30))
#define PMU_WAIT_IRQ_MASK1_SEL0_ADC_LOW                     ((uint32_t)(0x00000001UL << 31))
#define PMU_WAIT_IRQ_MASK2_SEL0_RTC_COMP0                   ((uint32_t)(0x00000001UL << 0))
#define PMU_WAIT_IRQ_MASK2_SEL0_RTC_COMP1                   ((uint32_t)(0x00000001UL << 1))
#define PMU_WAIT_IRQ_MASK2_SEL0_RTC_PRESCALE                ((uint32_t)(0x00000001UL << 2))
#define PMU_WAIT_IRQ_MASK2_SEL0_RTC_OVERFLOW                ((uint32_t)(0x00000001UL << 3))
#define PMU_WAIT_IRQ_MASK2_SEL0_PT0_DISABLED                ((uint32_t)(0x00000001UL << 4))
#define PMU_WAIT_IRQ_MASK2_SEL0_PT1_DISABLED                ((uint32_t)(0x00000001UL << 5))
#define PMU_WAIT_IRQ_MASK2_SEL0_PT2_DISABLED                ((uint32_t)(0x00000001UL << 6))
#define PMU_WAIT_IRQ_MASK2_SEL0_PT3_DISABLED                ((uint32_t)(0x00000001UL << 7))
#define PMU_WAIT_IRQ_MASK2_SEL0_PT4_DISABLED                ((uint32_t)(0x00000001UL << 8))
#define PMU_WAIT_IRQ_MASK2_SEL0_PT5_DISABLED                ((uint32_t)(0x00000001UL << 9))
#define PMU_WAIT_IRQ_MASK2_SEL0_PT6_DISABLED                ((uint32_t)(0x00000001UL << 10))
#define PMU_WAIT_IRQ_MASK2_SEL0_PT7_DISABLED                ((uint32_t)(0x00000001UL << 11))
#define PMU_WAIT_IRQ_MASK2_SEL0_PT8_DISABLED                ((uint32_t)(0x00000001UL << 12))
#define PMU_WAIT_IRQ_MASK2_SEL0_PT9_DISABLED                ((uint32_t)(0x00000001UL << 13))
#define PMU_WAIT_IRQ_MASK2_SEL0_PT10_DISABLED               ((uint32_t)(0x00000001UL << 14))
#define PMU_WAIT_IRQ_MASK2_SEL0_PT11_DISABLED               ((uint32_t)(0x00000001UL << 15))
#define PMU_WAIT_IRQ_MASK2_SEL0_TMR0                        ((uint32_t)(0x00000001UL << 16))
#define PMU_WAIT_IRQ_MASK2_SEL0_TMR1                        ((uint32_t)(0x00000001UL << 17))
#define PMU_WAIT_IRQ_MASK2_SEL0_TMR2                        ((uint32_t)(0x00000001UL << 18))
#define PMU_WAIT_IRQ_MASK2_SEL0_TMR3                        ((uint32_t)(0x00000001UL << 19))
#define PMU_WAIT_IRQ_MASK2_SEL0_TMR4                        ((uint32_t)(0x00000001UL << 20))
#define PMU_WAIT_IRQ_MASK2_SEL0_TMR5                        ((uint32_t)(0x00000001UL << 21))
#define PMU_WAIT_IRQ_MASK2_SEL0_GPIO0                       ((uint32_t)(0x00000001UL << 22))
#define PMU_WAIT_IRQ_MASK2_SEL0_GPIO1                       ((uint32_t)(0x00000001UL << 23))
#define PMU_WAIT_IRQ_MASK2_SEL0_GPIO2                       ((uint32_t)(0x00000001UL << 24))
#define PMU_WAIT_IRQ_MASK2_SEL0_GPIO3                       ((uint32_t)(0x00000001UL << 25))
#define PMU_WAIT_IRQ_MASK2_SEL0_GPIO4                       ((uint32_t)(0x00000001UL << 26))
#define PMU_WAIT_IRQ_MASK2_SEL0_GPIO5                       ((uint32_t)(0x00000001UL << 27))
#define PMU_WAIT_IRQ_MASK2_SEL0_GPIO6                       ((uint32_t)(0x00000001UL << 28))
#define PMU_WAIT_IRQ_MASK2_SEL0_AES                         ((uint32_t)(0x00000001UL << 29))
#define PMU_WAIT_IRQ_MASK2_SEL0_MAA_DONE                    ((uint32_t)(0x00000001UL << 30))
#define PMU_WAIT_IRQ_MASK2_SEL0_OWM                         ((uint32_t)(0x00000001UL << 31))
#define PMU_WAIT_IRQ_MASK1_SEL1_GPIO7                       ((uint32_t)(0x00000001UL << 0))
#define PMU_WAIT_IRQ_MASK1_SEL1_GPIO8                       ((uint32_t)(0x00000001UL << 1))
#define PMU_WAIT_IRQ_MASK1_SEL1_PT12_DISABLED               ((uint32_t)(0x00000001UL << 2))
#define PMU_WAIT_IRQ_MASK1_SEL1_PT13_DISABLED               ((uint32_t)(0x00000001UL << 3))
#define PMU_WAIT_IRQ_MASK1_SEL1_PT14_DISABLED               ((uint32_t)(0x00000001UL << 4))
#define PMU_WAIT_IRQ_MASK1_SEL1_PT15_DISABLED               ((uint32_t)(0x00000001UL << 5))
#define PMU_WAIT_IRQ_MASK1_SEL1_PT0_INT                     ((uint32_t)(0x00000001UL << 6))
#define PMU_WAIT_IRQ_MASK1_SEL1_PT1_INT                     ((uint32_t)(0x00000001UL << 7))
#define PMU_WAIT_IRQ_MASK1_SEL1_PT2_INT                     ((uint32_t)(0x00000001UL << 8))
#define PMU_WAIT_IRQ_MASK1_SEL1_PT3_INT                     ((uint32_t)(0x00000001UL << 9))
#define PMU_WAIT_IRQ_MASK1_SEL1_PT4_INT                     ((uint32_t)(0x00000001UL << 10))
#define PMU_WAIT_IRQ_MASK1_SEL1_PT5_INT                     ((uint32_t)(0x00000001UL << 11))
#define PMU_WAIT_IRQ_MASK1_SEL1_PT6_INT                     ((uint32_t)(0x00000001UL << 12))
#define PMU_WAIT_IRQ_MASK1_SEL1_PT7_INT                     ((uint32_t)(0x00000001UL << 13))
#define PMU_WAIT_IRQ_MASK1_SEL1_PT8_INT                     ((uint32_t)(0x00000001UL << 14))
#define PMU_WAIT_IRQ_MASK1_SEL1_PT9_INT                     ((uint32_t)(0x00000001UL << 15))
#define PMU_WAIT_IRQ_MASK1_SEL1_PT10_INT                    ((uint32_t)(0x00000001UL << 16))
#define PMU_WAIT_IRQ_MASK1_SEL1_PT11_INT                    ((uint32_t)(0x00000001UL << 17))
#define PMU_WAIT_IRQ_MASK1_SEL1_PT12_INT                    ((uint32_t)(0x00000001UL << 18))
#define PMU_WAIT_IRQ_MASK1_SEL1_PT13_INT                    ((uint32_t)(0x00000001UL << 19))
#define PMU_WAIT_IRQ_MASK1_SEL1_PT14_INT                    ((uint32_t)(0x00000001UL << 20))
#define PMU_WAIT_IRQ_MASK1_SEL1_PT15_INT                    ((uint32_t)(0x00000001UL << 21))
#define PMU_WAIT_IRQ_MASK1_SEL1_SPIS_TX_FIFO_AE             ((uint32_t)(0x00000001UL << 22))
#define PMU_WAIT_IRQ_MASK1_SEL1_SPIS_RX_FIFO_AF             ((uint32_t)(0x00000001UL << 23))
#define PMU_WAIT_IRQ_MASK1_SEL1_SPIS_TX_NO_DATA             ((uint32_t)(0x00000001UL << 24))
#define PMU_WAIT_IRQ_MASK1_SEL1_SPIS_RX_DATA_LOST           ((uint32_t)(0x00000001UL << 25))
#define PMU_WAIT_IRQ_MASK1_SEL1_SPI0_TX_READY               ((uint32_t)(0x00000001UL << 26))
#define PMU_WAIT_IRQ_MASK1_SEL1_SPI1_TX_READY               ((uint32_t)(0x00000001UL << 27))
#define PMU_WAIT_IRQ_MASK1_SEL1_SPI2_TX_READY               ((uint32_t)(0x00000001UL << 28))
#define PMU_WAIT_IRQ_MASK1_SEL1_UART0_TX_DONE               ((uint32_t)(0x00000001UL << 29))
#define PMU_WAIT_IRQ_MASK1_SEL1_UART1_TX_DONE               ((uint32_t)(0x00000001UL << 30))
#define PMU_WAIT_IRQ_MASK1_SEL1_UART2_TX_DONE               ((uint32_t)(0x00000001UL << 31))
#define PMU_WAIT_IRQ_MASK2_SEL1_UART3_TX_DONE               ((uint32_t)(0x00000001UL << 0))
#define PMU_WAIT_IRQ_MASK2_SEL1_UART0_RX_DATA_READY         ((uint32_t)(0x00000001UL << 1))
#define PMU_WAIT_IRQ_MASK2_SEL1_UART1_RX_DATA_READY         ((uint32_t)(0x00000001UL << 2))
#define PMU_WAIT_IRQ_MASK2_SEL1_UART2_RX_DATA_READY         ((uint32_t)(0x00000001UL << 3))
#define PMU_WAIT_IRQ_MASK2_SEL1_UART3_RX_DATA_READY         ((uint32_t)(0x00000001UL << 4))

/* PMU interrupt sources for the TRANSFER opcode  */
#define PMU_TRANSFER_IRQ_UART0_TX_FIFO_AE                   ((uint32_t)(0x00000001UL << 0))
#define PMU_TRANSFER_IRQ_UART0_RX_FIFO_AF                   ((uint32_t)(0x00000001UL << 1))
#define PMU_TRANSFER_IRQ_UART1_TX_FIFO_AE                   ((uint32_t)(0x00000001UL << 2))
#define PMU_TRANSFER_IRQ_UART1_RX_FIFO_AF                   ((uint32_t)(0x00000001UL << 3))
#define PMU_TRANSFER_IRQ_UART2_TX_FIFO_AE                   ((uint32_t)(0x00000001UL << 4))
#define PMU_TRANSFER_IRQ_UART2_RX_FIFO_AF                   ((uint32_t)(0x00000001UL << 5))
#define PMU_TRANSFER_IRQ_UART3_TX_FIFO_AE                   ((uint32_t)(0x00000001UL << 6))
#define PMU_TRANSFER_IRQ_UART3_RX_FIFO_AF                   ((uint32_t)(0x00000001UL << 7))
#define PMU_TRANSFER_IRQ_SPI0_TX_FIFO_AE                    ((uint32_t)(0x00000001UL << 8))
#define PMU_TRANSFER_IRQ_SPI0_RX_FIFO_AF                    ((uint32_t)(0x00000001UL << 9))
#define PMU_TRANSFER_IRQ_SPI1_TX_FIFO_AE                    ((uint32_t)(0x00000001UL << 10))
#define PMU_TRANSFER_IRQ_SPI1_RX_FIFO_AF                    ((uint32_t)(0x00000001UL << 11))
#define PMU_TRANSFER_IRQ_SPI2_TX_FIFO_AE                    ((uint32_t)(0x00000001UL << 12))
#define PMU_TRANSFER_IRQ_SPI2_RX_FIFO_AF                    ((uint32_t)(0x00000001UL << 13))
#define PMU_TRANSFER_IRQ_I2CM0_TX_FIFO_EMPTY                ((uint32_t)(0x00000001UL << 14))
#define PMU_TRANSFER_IRQ_I2CM0_RX_FIFO_NOT_EMPTY            ((uint32_t)(0x00000001UL << 15))
#define PMU_TRANSFER_IRQ_I2CM0_RX_FIFO_FULL                 ((uint32_t)(0x00000001UL << 16))
#define PMU_TRANSFER_IRQ_I2CM1_TX_FIFO_EMPTY                ((uint32_t)(0x00000001UL << 17))
#define PMU_TRANSFER_IRQ_I2CM1_RX_FIFO_NOT_EMPTY            ((uint32_t)(0x00000001UL << 18))
#define PMU_TRANSFER_IRQ_I2CM1_RX_FIFO_FULL                 ((uint32_t)(0x00000001UL << 19))
#define PMU_TRANSFER_IRQ_I2CM2_TX_FIFO_EMPTY                ((uint32_t)(0x00000001UL << 20))
#define PMU_TRANSFER_IRQ_I2CM2_RX_FIFO_NOT_EMPTY            ((uint32_t)(0x00000001UL << 21))
#define PMU_TRANSFER_IRQ_I2CM2_RX_FIFO_FULL                 ((uint32_t)(0x00000001UL << 22))
#define PMU_TRANSFER_IRQ_SPIS_TX_FIFO_AE                    ((uint32_t)(0x00000001UL << 23))
#define PMU_TRANSFER_IRQ_SPIS_RX_FIFO_AF                    ((uint32_t)(0x00000001UL << 24))
///@endcond
#ifdef __cplusplus
}
#endif

#endif   /* _MXC_PMU_REGS_H_ */

