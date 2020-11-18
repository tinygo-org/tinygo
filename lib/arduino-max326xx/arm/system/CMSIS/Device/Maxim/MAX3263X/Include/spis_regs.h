/**
 * @file
 * @brief   Registers, Bit Masks and Bit Positions for the SPIS Peripheral Module.
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
 * $Date: 2017-06-02 08:57:10 -0500 (Fri, 02 Jun 2017) $
 * $Revision: 28318 $
 *
 *************************************************************************** */

/* Define to prevent redundant inclusion */
#ifndef _MXC_SPIS_REGS_H_
#define _MXC_SPIS_REGS_H_

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
 * @ingroup     spis
 * @defgroup    spis_registers Registers
 * @brief       Registers, Bit Masks and Bit Positions for the SPIS Peripheral Module.
 * @{
 */

/**
 * Structure type to access the SPI Slave Peripheral Module Registers
 */
typedef struct {
    __IO uint32_t gen_ctrl;                             /**< SPIS_GEN_CTRL Register - SPI Slave General Control Register                                           */
    __IO uint32_t fifo_ctrl;                            /**< SPIS_FIFO_CTRL Register - SPI Slave FIFO Control Register                                              */
    __IO uint32_t fifo_stat;                            /**< SPIS_FIFO_STAT Register - SPI Slave FIFO Status Register                                               */
    __IO uint32_t intfl;                                /**< SPIS_INTFL Register - SPI Slave Interrupt Flags                                                    */
    __IO uint32_t inten;                                /**< SPIS_INTEN Register - SPI Slave Interrupt Enable/Disable Settings                                  */
} mxc_spis_regs_t;


/**
 * Structure type for the SPI Slave Transmit and Receive FIFOs. 
 */
typedef struct {
    union {                                             /*  0x0000-0x07FC   SPI Slave FIFO TX Write Space                                                */
        __IO uint8_t  tx_8[2048];                       /**< 8-bit access to Transmit FIFO */
        __IO uint16_t tx_16[1024];                      /**< 16-bit access to Transmit FIFO */
        __IO uint32_t tx_32[512];                       /**< 32-bit access to Transmit FIFO */
    };
    union {                                             /*  0x0800-0x0FFC   SPI Slave FIFO RX Read Space                                                 */
        __IO uint8_t  rx_8[2048];                       /**< 8-bit access to Receive FIFO */
        __IO uint16_t rx_16[1024];                      /**< 16-bit access to Receive FIFO */
        __IO uint32_t rx_32[512];                       /**< 32-bit access to Receive FIFO */
    };
} mxc_spis_fifo_regs_t;
/**@} end of group spis_registers */

/*
   Register offsets for module SPIS.
*/
/**
 * @ingroup    spis_registers
 * @defgroup   SPIS_Register_Offsets Register Offsets
 * @brief      SPI Slave Register Offsets from the SPIS[n] Base Peripheral Address, where  \c n \c = SPIS Instance Number. 
 * @{
 */
#define MXC_R_SPIS_OFFS_GEN_CTRL                            ((uint32_t)0x00000000UL)        /**< Offset from SPIS[n] Base Peripheral Address: <tt>\b 0x0000</tt>*/
#define MXC_R_SPIS_OFFS_FIFO_CTRL                           ((uint32_t)0x00000004UL)        /**< Offset from SPIS[n] Base Peripheral Address: <tt>\b 0x0004</tt>*/
#define MXC_R_SPIS_OFFS_FIFO_STAT                           ((uint32_t)0x00000008UL)        /**< Offset from SPIS[n] Base Peripheral Address: <tt>\b 0x0008</tt>*/
#define MXC_R_SPIS_OFFS_INTFL                               ((uint32_t)0x0000000CUL)        /**< Offset from SPIS[n] Base Peripheral Address: <tt>\b 0x000C</tt>*/
#define MXC_R_SPIS_OFFS_INTEN                               ((uint32_t)0x00000010UL)        /**< Offset from SPIS[n] Base Peripheral Address: <tt>\b 0x0010</tt>*/
/**@} end of group SPIS_Register_Offsets*/
/**
 * @ingroup    spis_registers
 * @defgroup   SPIS_FIFO_Offsets FIFO Offsets
 * @brief      SPI Slave FIFO Offsets from the SPIS[n] Base FIFO Address, where  \c n \c = SPIS Instance Number. 
 * @{
 */
#define MXC_R_SPIS_FIFO_OFFS_TX                             ((uint32_t)0x00000000UL)    /**< Offset from SPIS[n] Base FIFO Address: <tt>\b 0x0000</tt> */
#define MXC_R_SPIS_FIFO_OFFS_RX                             ((uint32_t)0x00000800UL)    /**< Offset from SPIS[n] Base FIFO Address: <tt>\b 0x0800</tt> */
/**@} end of group SPIS_FIFO_Offsets*/


/*
   Field positions and masks for module SPIS.
*/
/**
 * @ingroup  spis_registers
 * @defgroup SPIS_GEN_CTRL_Register SPIS_GEN_CTRL
 * @brief    Field Positions and Bit Masks for the SPIS_GEN_CTRL register
 * @{
 */
#define MXC_F_SPIS_GEN_CTRL_SPI_SLAVE_EN_POS                0                                                                   /**< SPI_SLAVE_EN Position */
#define MXC_F_SPIS_GEN_CTRL_SPI_SLAVE_EN                    ((uint32_t)(0x00000001UL << MXC_F_SPIS_GEN_CTRL_SPI_SLAVE_EN_POS))  /**< SPI_SLAVE_EN Mask */
#define MXC_F_SPIS_GEN_CTRL_TX_FIFO_EN_POS                  1                                                                   /**< TX_FIFO_EN Position */
#define MXC_F_SPIS_GEN_CTRL_TX_FIFO_EN                      ((uint32_t)(0x00000001UL << MXC_F_SPIS_GEN_CTRL_TX_FIFO_EN_POS))    /**< TX_FIFO_EN Mask */
#define MXC_F_SPIS_GEN_CTRL_RX_FIFO_EN_POS                  2                                                                   /**< RX_FIFO_EN Position */
#define MXC_F_SPIS_GEN_CTRL_RX_FIFO_EN                      ((uint32_t)(0x00000001UL << MXC_F_SPIS_GEN_CTRL_RX_FIFO_EN_POS))    /**< RX_FIFO_EN Mask */
#define MXC_F_SPIS_GEN_CTRL_DATA_WIDTH_POS                  4                                                                   /**< DATA_WIDTH Position */
#define MXC_F_SPIS_GEN_CTRL_DATA_WIDTH                      ((uint32_t)(0x00000003UL << MXC_F_SPIS_GEN_CTRL_DATA_WIDTH_POS))    /**< DATA_WIDTH Mask */
#define MXC_F_SPIS_GEN_CTRL_SPI_MODE_POS                    16                                                                  /**< SPI_MODE Position */
#define MXC_F_SPIS_GEN_CTRL_SPI_MODE                        ((uint32_t)(0x00000003UL << MXC_F_SPIS_GEN_CTRL_SPI_MODE_POS))      /**< SPI_MODE Mask */
#define MXC_F_SPIS_GEN_CTRL_TX_CLK_INVERT_POS               20                                                                  /**< TX_CLK_INVERT Position */
#define MXC_F_SPIS_GEN_CTRL_TX_CLK_INVERT                   ((uint32_t)(0x00000001UL << MXC_F_SPIS_GEN_CTRL_TX_CLK_INVERT_POS)) /**< TX_CLK_INVERT Mask */
/**@} end of group SPIS_GEN_CTRL*/
/**
 * @ingroup  spis_registers
 * @defgroup SPIS_FIFO_CTRL_Register SPIS_FIFO_CTRL
 * @brief    Field Positions and Bit Masks for the SPIS_FIFO_CTRL register
 * @{
 */ 
#define MXC_F_SPIS_FIFO_CTRL_TX_FIFO_AE_LVL_POS             0                                                                       /**< TX_FIFO_AE_LVL Position */
#define MXC_F_SPIS_FIFO_CTRL_TX_FIFO_AE_LVL                 ((uint32_t)(0x0000001FUL << MXC_F_SPIS_FIFO_CTRL_TX_FIFO_AE_LVL_POS))   /**< TX_FIFO_AE_LVL Mask */
#define MXC_F_SPIS_FIFO_CTRL_RX_FIFO_AF_LVL_POS             8                                                                       /**< RX_FIFO_AF_LVL Position */
#define MXC_F_SPIS_FIFO_CTRL_RX_FIFO_AF_LVL                 ((uint32_t)(0x0000001FUL << MXC_F_SPIS_FIFO_CTRL_RX_FIFO_AF_LVL_POS))   /**< RX_FIFO_AF_LVL Mask */
/**@} end of group SPIS_FIFO_CTRL_Register*/
/**
 * @ingroup  spis_registers
 * @defgroup SPIS_FIFO_STAT_Register SPIS_FIFO_STAT
 * @brief    Field Positions and Bit Masks for the SPIS_FIFO_STAT register
 * @{
 */ 
#define MXC_F_SPIS_FIFO_STAT_TX_FIFO_USED_POS               0                                                                       /**< TX_FIFO_USED Position */ 
#define MXC_F_SPIS_FIFO_STAT_TX_FIFO_USED                   ((uint32_t)(0x0000003FUL << MXC_F_SPIS_FIFO_STAT_TX_FIFO_USED_POS))     /**< TX_FIFO_USED Mask */
#define MXC_F_SPIS_FIFO_STAT_RX_FIFO_USED_POS               8                                                                       /**< RX_FIFO_USED Position */ 
#define MXC_F_SPIS_FIFO_STAT_RX_FIFO_USED                   ((uint32_t)(0x0000003FUL << MXC_F_SPIS_FIFO_STAT_RX_FIFO_USED_POS))     /**< RX_FIFO_USED Mask */
/**@} end of group SPIS_FIFO_STAT_Register*/
/**
 * @ingroup  spis_registers
 * @defgroup SPIS_INTFL_Register SPIS_INTFL
 * @brief    Field Positions and Bit Masks for the SPIS_INTFL register
 * @{
 */ 
#define MXC_F_SPIS_INTFL_TX_FIFO_AE_POS                     0                                                                   /**< TX_FIFO_AE Position */
#define MXC_F_SPIS_INTFL_TX_FIFO_AE                         ((uint32_t)(0x00000001UL << MXC_F_SPIS_INTFL_TX_FIFO_AE_POS))       /**< TX_FIFO_AE Mask */
#define MXC_F_SPIS_INTFL_RX_FIFO_AF_POS                     1                                                                   /**< RX_FIFO_AF  Position */
#define MXC_F_SPIS_INTFL_RX_FIFO_AF                         ((uint32_t)(0x00000001UL << MXC_F_SPIS_INTFL_RX_FIFO_AF_POS))       /**< RX_FIFO_AF Mask */
#define MXC_F_SPIS_INTFL_TX_NO_DATA_POS                     2                                                                   /**< TX_NO_DATA Position */
#define MXC_F_SPIS_INTFL_TX_NO_DATA                         ((uint32_t)(0x00000001UL << MXC_F_SPIS_INTFL_TX_NO_DATA_POS))       /**< TX_NO_DATA Mask */
#define MXC_F_SPIS_INTFL_RX_LOST_DATA_POS                   3                                                                   /**< RX_LOST_DATA Position */
#define MXC_F_SPIS_INTFL_RX_LOST_DATA                       ((uint32_t)(0x00000001UL << MXC_F_SPIS_INTFL_RX_LOST_DATA_POS))     /**< RX_LOST_DATA Mask */
#define MXC_F_SPIS_INTFL_TX_UNDERFLOW_POS                   4                                                                   /**< TX_UNDERFLOW Position */
#define MXC_F_SPIS_INTFL_TX_UNDERFLOW                       ((uint32_t)(0x00000001UL << MXC_F_SPIS_INTFL_TX_UNDERFLOW_POS))     /**< TX_UNDERFLOW Mask */
#define MXC_F_SPIS_INTFL_SS_ASSERTED_POS                    5                                                                   /**< SS_ASSERTED Position */
#define MXC_F_SPIS_INTFL_SS_ASSERTED                        ((uint32_t)(0x00000001UL << MXC_F_SPIS_INTFL_SS_ASSERTED_POS))      /**< SS_ASSERTED Mask */
#define MXC_F_SPIS_INTFL_SS_DEASSERTED_POS                  6                                                                   /**< SS_DEASSERTED Position */
#define MXC_F_SPIS_INTFL_SS_DEASSERTED                      ((uint32_t)(0x00000001UL << MXC_F_SPIS_INTFL_SS_DEASSERTED_POS))    /**< SS_DEASSERTED Mask */
/**@} end of group SPIS_INTFL_Register*/
/**
 * @ingroup  spis_registers
 * @defgroup SPIS_INTEN_Register SPIS_INTEN
 * @brief    Field Positions and Bit Masks for the SPIS_INTEN register
 * @{
 */ 
#define MXC_F_SPIS_INTEN_TX_FIFO_AE_POS                     0                                                                   /**< TX_FIFO_AE Position */
#define MXC_F_SPIS_INTEN_TX_FIFO_AE                         ((uint32_t)(0x00000001UL << MXC_F_SPIS_INTEN_TX_FIFO_AE_POS))       /**< TX_FIFO_AE Mask */
#define MXC_F_SPIS_INTEN_RX_FIFO_AF_POS                     1                                                                   /**< RX_FIFO_AF  Position */
#define MXC_F_SPIS_INTEN_RX_FIFO_AF                         ((uint32_t)(0x00000001UL << MXC_F_SPIS_INTEN_RX_FIFO_AF_POS))       /**< RX_FIFO_AF Mask */
#define MXC_F_SPIS_INTEN_TX_NO_DATA_POS                     2                                                                   /**< TX_NO_DATA Position */
#define MXC_F_SPIS_INTEN_TX_NO_DATA                         ((uint32_t)(0x00000001UL << MXC_F_SPIS_INTEN_TX_NO_DATA_POS))       /**< TX_NO_DATA Mask */
#define MXC_F_SPIS_INTEN_RX_LOST_DATA_POS                   3                                                                   /**< RX_LOST_DATA Position */
#define MXC_F_SPIS_INTEN_RX_LOST_DATA                       ((uint32_t)(0x00000001UL << MXC_F_SPIS_INTEN_RX_LOST_DATA_POS))     /**< RX_LOST_DATA Mask */
#define MXC_F_SPIS_INTEN_TX_UNDERFLOW_POS                   4                                                                   /**< TX_UNDERFLOW Position */
#define MXC_F_SPIS_INTEN_TX_UNDERFLOW                       ((uint32_t)(0x00000001UL << MXC_F_SPIS_INTEN_TX_UNDERFLOW_POS))     /**< TX_UNDERFLOW Mask */
#define MXC_F_SPIS_INTEN_SS_ASSERTED_POS                    5                                                                   /**< SS_ASSERTED Position */
#define MXC_F_SPIS_INTEN_SS_ASSERTED                        ((uint32_t)(0x00000001UL << MXC_F_SPIS_INTEN_SS_ASSERTED_POS))      /**< SS_ASSERTED Mask */
#define MXC_F_SPIS_INTEN_SS_DEASSERTED_POS                  6                                                                   /**< SS_DEASSERTED Position */
#define MXC_F_SPIS_INTEN_SS_DEASSERTED                      ((uint32_t)(0x00000001UL << MXC_F_SPIS_INTEN_SS_DEASSERTED_POS))    /**< SS_DEASSERTED Mask */
/**@} end of group SPIS_INTEN_Register*/
#ifdef __cplusplus
}
#endif

#endif   /* _MXC_SPIS_REGS_H_ */

