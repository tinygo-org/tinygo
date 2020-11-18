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
 * $Date: 2016-03-11 11:46:02 -0600 (Fri, 11 Mar 2016) $
 * $Revision: 21838 $
 *
 ******************************************************************************/

#ifndef _MXC_SPIB_REGS_H_
#define _MXC_SPIB_REGS_H_

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
    __IO uint32_t master_cfg;                           /*  0x0000          SPIB Master Configuration                                                    */
    __IO uint32_t oob_ctrl;                             /*  0x0004          SPIB OOB Control                                                             */
    __IO uint32_t intfl;                                /*  0x0008          SPIB Interrupt Flags                                                         */
    __IO uint32_t inten;                                /*  0x000C          SPIB Interrupt Enables                                                       */
    __IO uint32_t slave_reg;                            /*  0x0010          SPIB Slave Register Access                                                   */
} mxc_spib_regs_t;


/*
   Register offsets for module SPIB.
*/

#define MXC_R_SPIB_OFFS_MASTER_CFG                          ((uint32_t)0x00000000UL)
#define MXC_R_SPIB_OFFS_OOB_CTRL                            ((uint32_t)0x00000004UL)
#define MXC_R_SPIB_OFFS_INTFL                               ((uint32_t)0x00000008UL)
#define MXC_R_SPIB_OFFS_INTEN                               ((uint32_t)0x0000000CUL)
#define MXC_R_SPIB_OFFS_SLAVE_REG                           ((uint32_t)0x00000010UL)


/*
   Field positions and masks for module SPIB.
*/

#define MXC_F_SPIB_MASTER_CFG_SPI_MODE_POS                  0
#define MXC_F_SPIB_MASTER_CFG_SPI_MODE                      ((uint32_t)(0x00000003UL << MXC_F_SPIB_MASTER_CFG_SPI_MODE_POS))
#define MXC_F_SPIB_MASTER_CFG_SPI_WIDTH_POS                 2
#define MXC_F_SPIB_MASTER_CFG_SPI_WIDTH                     ((uint32_t)(0x00000003UL << MXC_F_SPIB_MASTER_CFG_SPI_WIDTH_POS))
#define MXC_F_SPIB_MASTER_CFG_SCK_HI_CLK_POS                8
#define MXC_F_SPIB_MASTER_CFG_SCK_HI_CLK                    ((uint32_t)(0x0000000FUL << MXC_F_SPIB_MASTER_CFG_SCK_HI_CLK_POS))
#define MXC_F_SPIB_MASTER_CFG_SCK_LO_CLK_POS                12
#define MXC_F_SPIB_MASTER_CFG_SCK_LO_CLK                    ((uint32_t)(0x0000000FUL << MXC_F_SPIB_MASTER_CFG_SCK_LO_CLK_POS))
#define MXC_F_SPIB_MASTER_CFG_ACT_DELAY_POS                 16
#define MXC_F_SPIB_MASTER_CFG_ACT_DELAY                     ((uint32_t)(0x00000003UL << MXC_F_SPIB_MASTER_CFG_ACT_DELAY_POS))
#define MXC_F_SPIB_MASTER_CFG_INACT_DELAY_POS               18
#define MXC_F_SPIB_MASTER_CFG_INACT_DELAY                   ((uint32_t)(0x00000003UL << MXC_F_SPIB_MASTER_CFG_INACT_DELAY_POS))

#define MXC_F_SPIB_OOB_CTRL_MONITOR_SLAVE_INT0_POS          0
#define MXC_F_SPIB_OOB_CTRL_MONITOR_SLAVE_INT0              ((uint32_t)(0x00000001UL << MXC_F_SPIB_OOB_CTRL_MONITOR_SLAVE_INT0_POS))
#define MXC_F_SPIB_OOB_CTRL_MONITOR_SLAVE_INT1_POS          1
#define MXC_F_SPIB_OOB_CTRL_MONITOR_SLAVE_INT1              ((uint32_t)(0x00000001UL << MXC_F_SPIB_OOB_CTRL_MONITOR_SLAVE_INT1_POS))
#define MXC_F_SPIB_OOB_CTRL_MONITOR_SLAVE_INT2_POS          2
#define MXC_F_SPIB_OOB_CTRL_MONITOR_SLAVE_INT2              ((uint32_t)(0x00000001UL << MXC_F_SPIB_OOB_CTRL_MONITOR_SLAVE_INT2_POS))

#define MXC_F_SPIB_INTFL_SLAVE_INT0_SYS_INT_POS             0
#define MXC_F_SPIB_INTFL_SLAVE_INT0_SYS_INT                 ((uint32_t)(0x00000001UL << MXC_F_SPIB_INTFL_SLAVE_INT0_SYS_INT_POS))
#define MXC_F_SPIB_INTFL_SLAVE_INT1_SYS_RST_POS             1
#define MXC_F_SPIB_INTFL_SLAVE_INT1_SYS_RST                 ((uint32_t)(0x00000001UL << MXC_F_SPIB_INTFL_SLAVE_INT1_SYS_RST_POS))
#define MXC_F_SPIB_INTFL_SLAVE_INT2_BAD_AHB_RESP_POS        2
#define MXC_F_SPIB_INTFL_SLAVE_INT2_BAD_AHB_RESP            ((uint32_t)(0x00000001UL << MXC_F_SPIB_INTFL_SLAVE_INT2_BAD_AHB_RESP_POS))

#define MXC_F_SPIB_INTEN_SLAVE_INT0_SYS_INT_POS             0
#define MXC_F_SPIB_INTEN_SLAVE_INT0_SYS_INT                 ((uint32_t)(0x00000001UL << MXC_F_SPIB_INTEN_SLAVE_INT0_SYS_INT_POS))
#define MXC_F_SPIB_INTEN_SLAVE_INT1_SYS_RST_POS             1
#define MXC_F_SPIB_INTEN_SLAVE_INT1_SYS_RST                 ((uint32_t)(0x00000001UL << MXC_F_SPIB_INTEN_SLAVE_INT1_SYS_RST_POS))
#define MXC_F_SPIB_INTEN_SLAVE_INT2_BAD_AHB_RESP_POS        2
#define MXC_F_SPIB_INTEN_SLAVE_INT2_BAD_AHB_RESP            ((uint32_t)(0x00000001UL << MXC_F_SPIB_INTEN_SLAVE_INT2_BAD_AHB_RESP_POS))

#define MXC_F_SPIB_SLAVE_REG_ENABLE_SLAVE_REG_ACCESS_POS    0
#define MXC_F_SPIB_SLAVE_REG_ENABLE_SLAVE_REG_ACCESS        ((uint32_t)(0x00000001UL << MXC_F_SPIB_SLAVE_REG_ENABLE_SLAVE_REG_ACCESS_POS))
#define MXC_F_SPIB_SLAVE_REG_START_ACCESS_CYCLE_POS         1
#define MXC_F_SPIB_SLAVE_REG_START_ACCESS_CYCLE             ((uint32_t)(0x00000001UL << MXC_F_SPIB_SLAVE_REG_START_ACCESS_CYCLE_POS))
#define MXC_F_SPIB_SLAVE_REG_ACCESS_TYPE_POS                2
#define MXC_F_SPIB_SLAVE_REG_ACCESS_TYPE                    ((uint32_t)(0x00000001UL << MXC_F_SPIB_SLAVE_REG_ACCESS_TYPE_POS))
#define MXC_F_SPIB_SLAVE_REG_SLAVE_REG_WRITE_DATA_POS       8
#define MXC_F_SPIB_SLAVE_REG_SLAVE_REG_WRITE_DATA           ((uint32_t)(0x000000FFUL << MXC_F_SPIB_SLAVE_REG_SLAVE_REG_WRITE_DATA_POS))
#define MXC_F_SPIB_SLAVE_REG_SLAVE_REG_READ_DATA_POS        16
#define MXC_F_SPIB_SLAVE_REG_SLAVE_REG_READ_DATA            ((uint32_t)(0x000000FFUL << MXC_F_SPIB_SLAVE_REG_SLAVE_REG_READ_DATA_POS))



/*
   Field values and shifted values for module SPIB.
*/

#define MXC_V_SPIB_MASTER_CFG_SPI_MODE_SCK_HI_SAMPLE_RISING                     ((uint32_t)(0x00000000UL))
#define MXC_V_SPIB_MASTER_CFG_SPI_MODE_SCK_LO_SAMPLE_FALLING                    ((uint32_t)(0x00000003UL))

#define MXC_S_SPIB_MASTER_CFG_SPI_MODE_SCK_HI_SAMPLE_RISING                     ((uint32_t)(MXC_V_SPIB_MASTER_CFG_SPI_MODE_SCK_HI_SAMPLE_RISING    << MXC_F_SPIB_MASTER_CFG_SPI_MODE_POS))
#define MXC_S_SPIB_MASTER_CFG_SPI_MODE_SCK_LO_SAMPLE_FALLING                    ((uint32_t)(MXC_V_SPIB_MASTER_CFG_SPI_MODE_SCK_LO_SAMPLE_FALLING   << MXC_F_SPIB_MASTER_CFG_SPI_MODE_POS))

#define MXC_V_SPIB_MASTER_CFG_SPI_WIDTH_ACTIVE_HIGH                             ((uint32_t)(0x00000000UL))
#define MXC_V_SPIB_MASTER_CFG_SPI_WIDTH_ACTIVE_LOW                              ((uint32_t)(0x00000001UL))

#define MXC_S_SPIB_MASTER_CFG_SPI_WIDTH_ACTIVE_HIGH                             ((uint32_t)(MXC_V_SPIB_MASTER_CFG_SPI_WIDTH_ACTIVE_HIGH  << MXC_F_SPIB_MASTER_CFG_SPI_WIDTH_POS))
#define MXC_S_SPIB_MASTER_CFG_SPI_WIDTH_ACTIVE_LOW                              ((uint32_t)(MXC_V_SPIB_MASTER_CFG_SPI_WIDTH_ACTIVE_LOW   << MXC_F_SPIB_MASTER_CFG_SPI_WIDTH_POS))

#define MXC_V_SPIB_MASTER_CFG_ACT_DELAY_OFF                                     ((uint32_t)(0x00000000UL))
#define MXC_V_SPIB_MASTER_CFG_ACT_DELAY_FOR_2_MOD_CLK                           ((uint32_t)(0x00000001UL))
#define MXC_V_SPIB_MASTER_CFG_ACT_DELAY_FOR_4_MOD_CLK                           ((uint32_t)(0x00000002UL))
#define MXC_V_SPIB_MASTER_CFG_ACT_DELAY_FOR_8_MOD_CLK                           ((uint32_t)(0x00000003UL))

#define MXC_S_SPIB_MASTER_CFG_ACT_DELAY_OFF                                     ((uint32_t)(MXC_V_SPIB_MASTER_CFG_ACT_DELAY_OFF             << MXC_F_SPIB_MASTER_CFG_ACT_DELAY_POS))
#define MXC_S_SPIB_MASTER_CFG_ACT_DELAY_FOR_2_MOD_CLK                           ((uint32_t)(MXC_V_SPIB_MASTER_CFG_ACT_DELAY_FOR_2_MOD_CLK   << MXC_F_SPIB_MASTER_CFG_ACT_DELAY_POS))
#define MXC_S_SPIB_MASTER_CFG_ACT_DELAY_FOR_4_MOD_CLK                           ((uint32_t)(MXC_V_SPIB_MASTER_CFG_ACT_DELAY_FOR_4_MOD_CLK   << MXC_F_SPIB_MASTER_CFG_ACT_DELAY_POS))
#define MXC_S_SPIB_MASTER_CFG_ACT_DELAY_FOR_8_MOD_CLK                           ((uint32_t)(MXC_V_SPIB_MASTER_CFG_ACT_DELAY_FOR_8_MOD_CLK   << MXC_F_SPIB_MASTER_CFG_ACT_DELAY_POS))

#define MXC_V_SPIB_MASTER_CFG_INACT_DELAY_OFF                                   ((uint32_t)(0x00000000UL))
#define MXC_V_SPIB_MASTER_CFG_INACT_DELAY_FOR_2_MOD_CLK                         ((uint32_t)(0x00000001UL))
#define MXC_V_SPIB_MASTER_CFG_INACT_DELAY_FOR_4_MOD_CLK                         ((uint32_t)(0x00000002UL))
#define MXC_V_SPIB_MASTER_CFG_INACT_DELAY_FOR_8_MOD_CLK                         ((uint32_t)(0x00000003UL))

#define MXC_S_SPIB_MASTER_CFG_INACT_DELAY_OFF                                   ((uint32_t)(MXC_V_SPIB_MASTER_CFG_INACT_DELAY_OFF             << MXC_F_SPIB_MASTER_CFG_INACT_DELAY_POS))
#define MXC_S_SPIB_MASTER_CFG_INACT_DELAY_FOR_2_MOD_CLK                         ((uint32_t)(MXC_V_SPIB_MASTER_CFG_INACT_DELAY_FOR_2_MOD_CLK   << MXC_F_SPIB_MASTER_CFG_INACT_DELAY_POS))
#define MXC_S_SPIB_MASTER_CFG_INACT_DELAY_FOR_4_MOD_CLK                         ((uint32_t)(MXC_V_SPIB_MASTER_CFG_INACT_DELAY_FOR_4_MOD_CLK   << MXC_F_SPIB_MASTER_CFG_INACT_DELAY_POS))
#define MXC_S_SPIB_MASTER_CFG_INACT_DELAY_FOR_8_MOD_CLK                         ((uint32_t)(MXC_V_SPIB_MASTER_CFG_INACT_DELAY_FOR_8_MOD_CLK   << MXC_F_SPIB_MASTER_CFG_INACT_DELAY_POS))

#define MXC_V_SPIB_OOB_CTRL_MONITOR_SLAVE_INT0_DISABLED                         ((uint32_t)(0x00000000UL))
#define MXC_V_SPIB_OOB_CTRL_MONITOR_SLAVE_INT0_ENABLED                          ((uint32_t)(0x00000001UL))

#define MXC_S_SPIB_OOB_CTRL_MONITOR_SLAVE_INT0_DISABLED                         ((uint32_t)(MXC_V_SPIB_OOB_CTRL_MONITOR_SLAVE_INT0_DISABLED  << MXC_F_SPIB_OOB_CTRL_MONITOR_SLAVE_INT0_POS))
#define MXC_S_SPIB_OOB_CTRL_MONITOR_SLAVE_INT0_ENABLED                          ((uint32_t)(MXC_V_SPIB_OOB_CTRL_MONITOR_SLAVE_INT0_ENABLED   << MXC_F_SPIB_OOB_CTRL_MONITOR_SLAVE_INT0_POS))

#define MXC_V_SPIB_OOB_CTRL_MONITOR_SLAVE_INT1_DISABLED                         ((uint32_t)(0x00000000UL))
#define MXC_V_SPIB_OOB_CTRL_MONITOR_SLAVE_INT1_ENABLED                          ((uint32_t)(0x00000001UL))

#define MXC_S_SPIB_OOB_CTRL_MONITOR_SLAVE_INT1_DISABLED                         ((uint32_t)(MXC_V_SPIB_OOB_CTRL_MONITOR_SLAVE_INT1_DISABLED  << MXC_F_SPIB_OOB_CTRL_MONITOR_SLAVE_INT1_POS))
#define MXC_S_SPIB_OOB_CTRL_MONITOR_SLAVE_INT1_ENABLED                          ((uint32_t)(MXC_V_SPIB_OOB_CTRL_MONITOR_SLAVE_INT1_ENABLED   << MXC_F_SPIB_OOB_CTRL_MONITOR_SLAVE_INT1_POS))

#define MXC_V_SPIB_OOB_CTRL_MONITOR_SLAVE_INT2_DISABLED                         ((uint32_t)(0x00000000UL))
#define MXC_V_SPIB_OOB_CTRL_MONITOR_SLAVE_INT2_ENABLED                          ((uint32_t)(0x00000001UL))

#define MXC_S_SPIB_OOB_CTRL_MONITOR_SLAVE_INT2_DISABLED                         ((uint32_t)(MXC_V_SPIB_OOB_CTRL_MONITOR_SLAVE_INT2_DISABLED  << MXC_F_SPIB_OOB_CTRL_MONITOR_SLAVE_INT2_POS))
#define MXC_S_SPIB_OOB_CTRL_MONITOR_SLAVE_INT2_ENABLED                          ((uint32_t)(MXC_V_SPIB_OOB_CTRL_MONITOR_SLAVE_INT2_ENABLED   << MXC_F_SPIB_OOB_CTRL_MONITOR_SLAVE_INT2_POS))

#define MXC_V_SPIB_INTFL_SLAVE_INT0_SYS_INT_NONE                                ((uint32_t)(0x00000000UL))
#define MXC_V_SPIB_INTFL_SLAVE_INT0_SYS_INT_DETECTED                            ((uint32_t)(0x00000001UL))

#define MXC_S_SPIB_INTFL_SLAVE_INT0_SYS_INT_NONE                                ((uint32_t)(MXC_V_SPIB_INTFL_SLAVE_INT0_SYS_INT_NONE       << MXC_F_SPIB_INTFL_SLAVE_INT0_SYS_INT_POS))
#define MXC_S_SPIB_INTFL_SLAVE_INT0_SYS_INT_DETECTED                            ((uint32_t)(MXC_V_SPIB_INTFL_SLAVE_INT0_SYS_INT_DETECTED   << MXC_F_SPIB_INTFL_SLAVE_INT0_SYS_INT_POS))

#define MXC_V_SPIB_INTFL_SLAVE_INT1_SYS_RST_NONE                                ((uint32_t)(0x00000000UL))
#define MXC_V_SPIB_INTFL_SLAVE_INT1_SYS_RST_DETECTED                            ((uint32_t)(0x00000001UL))

#define MXC_S_SPIB_INTFL_SLAVE_INT1_SYS_RST_NONE                                ((uint32_t)(MXC_V_SPIB_INTFL_SLAVE_INT1_SYS_RST_NONE       << MXC_F_SPIB_INTFL_SLAVE_INT1_SYS_RST_POS))
#define MXC_S_SPIB_INTFL_SLAVE_INT1_SYS_RST_DETECTED                            ((uint32_t)(MXC_V_SPIB_INTFL_SLAVE_INT1_SYS_RST_DETECTED   << MXC_F_SPIB_INTFL_SLAVE_INT1_SYS_RST_POS))

#define MXC_V_SPIB_INTFL_SLAVE_INT2_BAD_AHB_RESP_NONE                           ((uint32_t)(0x00000000UL))
#define MXC_V_SPIB_INTFL_SLAVE_INT2_BAD_AHB_RESP_DETECTED                       ((uint32_t)(0x00000001UL))

#define MXC_S_SPIB_INTFL_SLAVE_INT2_BAD_AHB_RESP_NONE                           ((uint32_t)(MXC_V_SPIB_INTFL_SLAVE_INT2_BAD_AHB_RESP_NONE       << MXC_F_SPIB_INTFL_SLAVE_INT2_BAD_AHB_RESP_POS))
#define MXC_S_SPIB_INTFL_SLAVE_INT2_BAD_AHB_RESP_DETECTED                       ((uint32_t)(MXC_V_SPIB_INTFL_SLAVE_INT2_BAD_AHB_RESP_DETECTED   << MXC_F_SPIB_INTFL_SLAVE_INT2_BAD_AHB_RESP_POS))

#define MXC_V_SPIB_INTEN_SLAVE_INT0_SYS_INT_DISABLED                            ((uint32_t)(0x00000000UL))
#define MXC_V_SPIB_INTEN_SLAVE_INT0_SYS_INT_ENABLED                             ((uint32_t)(0x00000001UL))

#define MXC_S_SPIB_INTEN_SLAVE_INT0_SYS_INT_DISABLED                            ((uint32_t)(MXC_V_SPIB_INTEN_SLAVE_INT0_SYS_INT_DISABLED  << MXC_F_SPIB_INTEN_SLAVE_INT0_SYS_INT_POS))
#define MXC_S_SPIB_INTEN_SLAVE_INT0_SYS_INT_ENABLED                             ((uint32_t)(MXC_V_SPIB_INTEN_SLAVE_INT0_SYS_INT_ENABLED   << MXC_F_SPIB_INTEN_SLAVE_INT0_SYS_INT_POS))

#define MXC_V_SPIB_INTEN_SLAVE_INT1_SYS_RST_DISABLED                            ((uint32_t)(0x00000000UL))
#define MXC_V_SPIB_INTEN_SLAVE_INT1_SYS_RST_ENABLED                             ((uint32_t)(0x00000001UL))

#define MXC_S_SPIB_INTEN_SLAVE_INT1_SYS_RST_DISABLED                            ((uint32_t)(MXC_V_SPIB_INTEN_SLAVE_INT1_SYS_RST_DISABLED  << MXC_F_SPIB_INTEN_SLAVE_INT1_SYS_RST_POS))
#define MXC_S_SPIB_INTEN_SLAVE_INT1_SYS_RST_ENABLED                             ((uint32_t)(MXC_V_SPIB_INTEN_SLAVE_INT1_SYS_RST_ENABLED   << MXC_F_SPIB_INTEN_SLAVE_INT1_SYS_RST_POS))

#define MXC_V_SPIB_INTEN_SLAVE_INT2_BAD_AHB_RESP_DISABLED                       ((uint32_t)(0x00000000UL))
#define MXC_V_SPIB_INTEN_SLAVE_INT2_BAD_AHB_RESP_ENABLED                        ((uint32_t)(0x00000001UL))

#define MXC_S_SPIB_INTEN_SLAVE_INT2_BAD_AHB_RESP_DISABLED                       ((uint32_t)(MXC_V_SPIB_INTEN_SLAVE_INT2_BAD_AHB_RESP_DISABLED  << MXC_F_SPIB_INTEN_SLAVE_INT2_BAD_AHB_RESP_POS))
#define MXC_S_SPIB_INTEN_SLAVE_INT2_BAD_AHB_RESP_ENABLED                        ((uint32_t)(MXC_V_SPIB_INTEN_SLAVE_INT2_BAD_AHB_RESP_ENABLED   << MXC_F_SPIB_INTEN_SLAVE_INT2_BAD_AHB_RESP_POS))

#define MXC_V_SPIB_SLAVE_REG_ACCESS_TYPE_READ                                   ((uint32_t)(0x00000000UL))
#define MXC_V_SPIB_SLAVE_REG_ACCESS_TYPE_WRITE                                  ((uint32_t)(0x00000001UL))

#define MXC_S_SPIB_SLAVE_REG_ACCESS_TYPE_READ                                   ((uint32_t)(MXC_V_SPIB_SLAVE_REG_ACCESS_TYPE_READ    << MXC_F_SPIB_SLAVE_REG_ACCESS_TYPE_POS))
#define MXC_S_SPIB_SLAVE_REG_ACCESS_TYPE_WRITE                                  ((uint32_t)(MXC_V_SPIB_SLAVE_REG_ACCESS_TYPE_WRITE   << MXC_F_SPIB_SLAVE_REG_ACCESS_TYPE_POS))



#ifdef __cplusplus
}
#endif

#endif   /* _MXC_SPIB_REGS_H_ */

