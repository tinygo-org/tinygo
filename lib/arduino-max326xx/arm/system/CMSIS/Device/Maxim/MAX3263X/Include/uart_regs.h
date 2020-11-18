/**
 * @file
 * @brief Registers, Bit Masks and Bit Positions for the UART Peripheral Module.
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
 * $Date: 2017-02-16 08:55:49 -0600 (Thu, 16 Feb 2017) $
 * $Revision: 26457 $
 *
 *************************************************************************** */

/* **** Includes **** */
#include <stdint.h>
/* Define to prevent redundant inclusion */
#ifndef _MXC_UART_REGS_H_
#define _MXC_UART_REGS_H_

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
 * @ingroup    uart_top
 * @defgroup   uart_registers UART Registers
 * @brief      Hardware interface definitions for the UART Peripheral.
 * @details    Definitions for the Hardware Access Layer of the UART
 *             Peripherals.
 */
/**
 * @ingroup uart_registers
 * @{
 */
/**
 * Structure type for the UART peripheral registers allowing direct 32-bit access to each register.
 */
typedef struct {
    __IO uint32_t ctrl;                                 /**< <b><tt> 0x0000: </tt></b> UART_CTRL Register - UART Control Register.                    */
    __IO uint32_t baud;                                 /**< <b><tt> 0x0004: </tt></b> UART_BAUD Register - UART Baud Control Register.               */
    __IO uint32_t tx_fifo_ctrl;                         /**< <b><tt> 0x0008: </tt></b> UART_TX_FIFO_CTRL Register - UART TX FIFO Control Register.    */
    __IO uint32_t rx_fifo_ctrl;                         /**< <b><tt> 0x000C: </tt></b> UART_RX_FIFO_CTRL Register - UART RX FIFO Control Register.    */
    __IO uint32_t md_ctrl;                              /**< <b><tt> 0x0010: </tt></b> UART_MD_CTRL Register - UART Multidrop Control Register.       */
    __IO uint32_t intfl;                                /**< <b><tt> 0x0014: </tt></b> UART_INTFL Register - UART Interrupt Flags.                    */
    __IO uint32_t inten;                                /**< <b><tt> 0x0018: </tt></b> UART_INTEN Register - UART Interrupt Enable/Disable Control.   */
#if (MXC_UART_REV > 0)
    __R  uint32_t idle;                                 /**< <b><tt> 0x001C: </tt></b> UART_IDLE Register - UART Idle Status                          */
#endif
} mxc_uart_regs_t;
/**@} uart_registers */
/*
 * @ingroup uart_registers
 * @defgroup uart_fifos UART TX and RX FIFOs
 * @brief TX and RX FIFO access for reads and writes using 8-bit, 16-bit and 32-bit data types. 
 * @{
 */

/**
 * Structure type for accessing the UART Transmit and Receive FIFOs.
 */
typedef struct {
    union {                                             // tx fifo
        __IO uint8_t  tx;                               /**< TX FIFO write point for data to transmit.                                          */
        __IO uint8_t  tx_8[2048];                       /**< 8-bit access to TX FIFO.                                                           */
        __IO uint16_t tx_16[1024];                      /**< 16-bit access to TX FIFO.                                                          */
        __IO uint32_t tx_32[512];                       /**< 32-bit access to TX FIFO.                                                          */
    };
    union {                                             // rx fifo
        __IO uint8_t  rx;                               /**< RX FIFO read point for received data.                                              */
        __IO uint8_t  rx_8[2048];                       /**< 8-bit access to RX FIFO.                                                           */
        __IO uint16_t rx_16[1024];                      /**< 16-bit access to RX FIFO.                                                          */
        __IO uint32_t rx_32[512];                       /**< 32-bit access to RX FIFO.                                                          */
    };
} mxc_uart_fifo_regs_t;
/** @} uart_fifos */
/**
 * @ingroup    uart_registers
 * @defgroup   uart_register_offsets Register Offsets
 * @brief      UART Register offsets from the \c UARTn Base Peripheral Address, where  \c n \c = UART Instance Number. 
 * @{
 */
#define MXC_R_UART_OFFS_CTRL                                ((uint32_t)0x00000000UL) /**< UART_CTRL Register offset from UARTn Base Address: <b><tt> 0x0000 </tt></b>           */
#define MXC_R_UART_OFFS_BAUD                                ((uint32_t)0x00000004UL) /**< UART_BAUD register offset from UARTn Base Address: <b><tt> 0x0000 </tt></b>           */
#define MXC_R_UART_OFFS_TX_FIFO_CTRL                        ((uint32_t)0x00000008UL) /**< UART_TX_FIFO_CTRL register offset from UARTn Base Address: <b><tt> 0x0000 </tt></b>   */
#define MXC_R_UART_OFFS_RX_FIFO_CTRL                        ((uint32_t)0x0000000CUL) /**< UART_RX_FIFO_CTRL register offset from UARTn Base Address: <b><tt> 0x0000 </tt></b>   */
#define MXC_R_UART_OFFS_MD_CTRL                             ((uint32_t)0x00000010UL) /**< UART_MD_CTRL register offset from UARTn Base Address: <b><tt> 0x0000 </tt></b>        */
#define MXC_R_UART_OFFS_INTFL                               ((uint32_t)0x00000014UL) /**< UART_INTFL register offset from UARTn Base Address: <b><tt> 0x0000 </tt></b>          */
#define MXC_R_UART_OFFS_INTEN                               ((uint32_t)0x00000018UL) /**< UART_INTEN register offset from UARTn Base Address: <b><tt> 0x0000 </tt></b>          */
#if (MXC_UART_REV > 0)
#define MXC_R_UART_OFFS_IDLE                                ((uint32_t)0x0000001CUL) /**< UART_IDLE register offset from UARTn Base Address: <b><tt> 0x0000 </tt></b>          */
#endif
/** @} uart_register_offsets*/

/**
 * @ingroup    uart_registers
 * @defgroup   uart_fifo_offs FIFO Register Offsets
 * @brief      UART FIFO offsets from the UART_FIFOn Base FIFO Peripheral Address, where  \c n \c = UART Instance Number. 
 * @{ 
 */     
#define MXC_R_UART_FIFO_OFFS_TX                             ((uint32_t)0x00000000UL) /**< UART_FIFO_TX register offset from UART_FIFOn Base Address: <b><tt> 0x0000 </tt></b>   */
#define MXC_R_UART_FIFO_OFFS_RX                             ((uint32_t)0x00000800UL) /**< UART_FIFO_RX register offset from UART_FIFOn Base Address: <b><tt> 0x0000 </tt></b>   */
/** @} uart_fifo_offs */

/**
 * @ingroup  uart_registers
 * @defgroup UART_CTRL_register UART_CTRL
 * @brief    UART_CTRL register fields and masks
 * @{
 */
#define MXC_F_UART_CTRL_UART_EN_POS                         0                                                               /**< UART_EN Field Position         */
#define MXC_F_UART_CTRL_UART_EN                             ((uint32_t)(0x00000001UL << MXC_F_UART_CTRL_UART_EN_POS))       /**< UART_EN Field Mask             */
#define MXC_F_UART_CTRL_RX_FIFO_EN_POS                      1                                                               /**< RX_FIFO_EN Field Position      */
#define MXC_F_UART_CTRL_RX_FIFO_EN                          ((uint32_t)(0x00000001UL << MXC_F_UART_CTRL_RX_FIFO_EN_POS))    /**< RX_FIFO_EN Field Mask          */
#define MXC_F_UART_CTRL_TX_FIFO_EN_POS                      2                                                               /**< TX_FIFO_EN Field Position      */
#define MXC_F_UART_CTRL_TX_FIFO_EN                          ((uint32_t)(0x00000001UL << MXC_F_UART_CTRL_TX_FIFO_EN_POS))    /**< TX_FIFO_EN Field Mask          */
#define MXC_F_UART_CTRL_DATA_SIZE_POS                       4                                                               /**< DATA_SIZE Field Position       */
#define MXC_F_UART_CTRL_DATA_SIZE                           ((uint32_t)(0x00000003UL << MXC_F_UART_CTRL_DATA_SIZE_POS))     /**< DATA_SIZE Field Mask           */
#define MXC_F_UART_CTRL_EXTRA_STOP_POS                      8                                                               /**< EXTRA_STOP ield Position       */
#define MXC_F_UART_CTRL_EXTRA_STOP                          ((uint32_t)(0x00000001UL << MXC_F_UART_CTRL_EXTRA_STOP_POS))    /**< EXTRA_STOP Field Mask          */
#define MXC_F_UART_CTRL_PARITY_POS                          12                                                              /**< PARITY Field Position          */
#define MXC_F_UART_CTRL_PARITY                              ((uint32_t)(0x00000003UL << MXC_F_UART_CTRL_PARITY_POS))        /**< PARITY Field Mask              */
#define MXC_F_UART_CTRL_CTS_EN_POS                          16                                                              /**< CTS_EN Field Position          */
#define MXC_F_UART_CTRL_CTS_EN                              ((uint32_t)(0x00000001UL << MXC_F_UART_CTRL_CTS_EN_POS))        /**< CTS_EN Field Mask              */
#define MXC_F_UART_CTRL_CTS_POLARITY_POS                    17                                                              /**< CTS_POLARITY Field Position    */
#define MXC_F_UART_CTRL_CTS_POLARITY                        ((uint32_t)(0x00000001UL << MXC_F_UART_CTRL_CTS_POLARITY_POS))  /**< CTS_POLARITY Field Mask        */
#define MXC_F_UART_CTRL_RTS_EN_POS                          18                                                              /**< RTS_EN Field Position          */
#define MXC_F_UART_CTRL_RTS_EN                              ((uint32_t)(0x00000001UL << MXC_F_UART_CTRL_RTS_EN_POS))        /**< RTS_EN Field Mask              */
#define MXC_F_UART_CTRL_RTS_POLARITY_POS                    19                                                              /**< RTS_POLARITY Field Position    */
#define MXC_F_UART_CTRL_RTS_POLARITY                        ((uint32_t)(0x00000001UL << MXC_F_UART_CTRL_RTS_POLARITY_POS))  /**< RTS_POLARITY Field Mask        */
#define MXC_F_UART_CTRL_RTS_LEVEL_POS                       20                                                              /**< RTS_LEVEL Field Position       */
#define MXC_F_UART_CTRL_RTS_LEVEL                           ((uint32_t)(0x0000003FUL << MXC_F_UART_CTRL_RTS_LEVEL_POS))     /**< RTS_LEVEL Field Mask           */
/** @} */

/**
 * @ingroup  uart_registers
 * @defgroup UART_BAUD_register UART_BAUD
 * @brief    UART_BAUD register fields and masks
 * @{
 */
#define MXC_F_UART_BAUD_BAUD_DIVISOR_POS                    0                                                                       /**< BAUD_DIVISOR Field Position.   */
#define MXC_F_UART_BAUD_BAUD_DIVISOR                        ((uint32_t)(0x000000FFUL << MXC_F_UART_BAUD_BAUD_DIVISOR_POS))          /**< BAUD_DIVISOR Field Mask.       */
#define MXC_F_UART_BAUD_BAUD_MODE_POS                       8                                                                       /**< BAUD_MODE Field Position.      */
#define MXC_F_UART_BAUD_BAUD_MODE                           ((uint32_t)(0x00000003UL << MXC_F_UART_BAUD_BAUD_MODE_POS))             /**< BAUD_MODE Field Mask.          */
/** @} */

/**
 * @ingroup  uart_registers
 * @defgroup UART_TX_FIFO_CTRL_register UART_TX_FIFO_CTRL
 * @brief    UART_TX_FIFO_CTRL register fields and masks
 * @{
 */
#define MXC_F_UART_TX_FIFO_CTRL_FIFO_ENTRY_POS              0                                                                       /**< FIFO_ENTRY Field Position.     */
#define MXC_F_UART_TX_FIFO_CTRL_FIFO_ENTRY                  ((uint32_t)(0x0000001FUL << MXC_F_UART_TX_FIFO_CTRL_FIFO_ENTRY_POS))    /**< FIFO_ENTRY Field Mask.         */
#define MXC_F_UART_TX_FIFO_CTRL_FIFO_AE_LVL_POS             16                                                                      /**< FIFO_AE_LVL Field Position.    */
#define MXC_F_UART_TX_FIFO_CTRL_FIFO_AE_LVL                 ((uint32_t)(0x0000001FUL << MXC_F_UART_TX_FIFO_CTRL_FIFO_AE_LVL_POS))   /**< FIFO_AE_LVL Field Mask.        */
/** @} */

/**
 * @ingroup  uart_registers
 * @defgroup UART_RX_FIFO_CTRL_register UART_RX_FIFO_CTRL
 * @brief    UART_RX_FIFO_CTRL register fields and masks
 * @{
 */
#define MXC_F_UART_RX_FIFO_CTRL_FIFO_ENTRY_POS              0                                                                       /**< FIFO_ENTRY Field Position.  */
#define MXC_F_UART_RX_FIFO_CTRL_FIFO_ENTRY                  ((uint32_t)(0x0000001FUL << MXC_F_UART_RX_FIFO_CTRL_FIFO_ENTRY_POS))    /**< FIFO_ENTRY Field Mask.      */
#define MXC_F_UART_RX_FIFO_CTRL_FIFO_AF_LVL_POS             16                                                                      /**< FIFO_AF_LVL Field Position. */
#define MXC_F_UART_RX_FIFO_CTRL_FIFO_AF_LVL                 ((uint32_t)(0x0000001FUL << MXC_F_UART_RX_FIFO_CTRL_FIFO_AF_LVL_POS))   /**< FIFO_AF_LVL Field Mask.     */
/** @} */

/**
 * @ingroup  uart_registers
 * @defgroup UART_MD_CTRL_register UART_MD_CTRL
 * @brief    UART_MD_CTRL register fields and masks
 * @{
 */
#define MXC_F_UART_MD_CTRL_SLAVE_ADDR_POS                   0                                                                       /**< SLAVE_ADDR Field Position.      */
#define MXC_F_UART_MD_CTRL_SLAVE_ADDR                       ((uint32_t)(0x000000FFUL << MXC_F_UART_MD_CTRL_SLAVE_ADDR_POS))         /**< SLAVE_ADDR Field Mask.          */
#define MXC_F_UART_MD_CTRL_SLAVE_ADDR_MSK_POS               8                                                                       /**< SLAVE_ADDR_MSK Field Position.  */
#define MXC_F_UART_MD_CTRL_SLAVE_ADDR_MSK                   ((uint32_t)(0x000000FFUL << MXC_F_UART_MD_CTRL_SLAVE_ADDR_MSK_POS))     /**< SLAVE_ADDR_MSK Field Mask.      */
#define MXC_F_UART_MD_CTRL_MD_MSTR_POS                      16                                                                      /**< MD_MSTR Field Position.         */
#define MXC_F_UART_MD_CTRL_MD_MSTR                          ((uint32_t)(0x00000001UL << MXC_F_UART_MD_CTRL_MD_MSTR_POS))            /**< MD_MSTR Field Mask.             */
#define MXC_F_UART_MD_CTRL_TX_ADDR_MARK_POS                 17                                                                      /**< TX_ADDR_MARK Field Position.    */
#define MXC_F_UART_MD_CTRL_TX_ADDR_MARK                     ((uint32_t)(0x00000001UL << MXC_F_UART_MD_CTRL_TX_ADDR_MARK_POS))       /**< TX_ADDR_MARK Field Mask.        */
/** @} */

/**
 * @ingroup uart_registers
 * @defgroup UART_INTFL_Register UART_INTFL
 * @brief    UART_INTFL register fields and masks
 * @{
 */
#define MXC_F_UART_INTFL_TX_DONE_POS                        0                                                                       /**< TX_DONE Field Position.            */
#define MXC_F_UART_INTFL_TX_DONE                            ((uint32_t)(0x00000001UL << MXC_F_UART_INTFL_TX_DONE_POS))              /**< TX_DONE Field Mask.                */
#define MXC_F_UART_INTFL_TX_UNSTALLED_POS                   1                                                                       /**< TX_UNSTALLED Field Position.       */
#define MXC_F_UART_INTFL_TX_UNSTALLED                       ((uint32_t)(0x00000001UL << MXC_F_UART_INTFL_TX_UNSTALLED_POS))         /**< TX_UNSTALLED Field Mask.           */
#define MXC_F_UART_INTFL_TX_FIFO_AE_POS                     2                                                                       /**< TX_FIFO_AE Field Position.         */
#define MXC_F_UART_INTFL_TX_FIFO_AE                         ((uint32_t)(0x00000001UL << MXC_F_UART_INTFL_TX_FIFO_AE_POS))           /**< TX_FIFO_AE Field Mask.             */
#define MXC_F_UART_INTFL_RX_FIFO_NOT_EMPTY_POS              3                                                                       /**< RX_FIFO_NOT_EMPTY Field Position.  */
#define MXC_F_UART_INTFL_RX_FIFO_NOT_EMPTY                  ((uint32_t)(0x00000001UL << MXC_F_UART_INTFL_RX_FIFO_NOT_EMPTY_POS))    /**< RX_FIFO_NOT_EMPTY Field Mask.      */
#define MXC_F_UART_INTFL_RX_STALLED_POS                     4                                                                       /**< RX_STALLED Field Position.         */
#define MXC_F_UART_INTFL_RX_STALLED                         ((uint32_t)(0x00000001UL << MXC_F_UART_INTFL_RX_STALLED_POS))           /**< RX_STALLED Field Mask.             */
#define MXC_F_UART_INTFL_RX_FIFO_AF_POS                     5                                                                       /**< RX_FIFO_AF Field Position.         */
#define MXC_F_UART_INTFL_RX_FIFO_AF                         ((uint32_t)(0x00000001UL << MXC_F_UART_INTFL_RX_FIFO_AF_POS))           /**< RX_FIFO_AF Field Mask.             */
#define MXC_F_UART_INTFL_RX_FIFO_OVERFLOW_POS               6                                                                       /**< RX_FIFO_OVERFLOW Field Position.   */
#define MXC_F_UART_INTFL_RX_FIFO_OVERFLOW                   ((uint32_t)(0x00000001UL << MXC_F_UART_INTFL_RX_FIFO_OVERFLOW_POS))     /**< RX_FIFO_OVERFLOW Field Mask.       */
#define MXC_F_UART_INTFL_RX_FRAMING_ERR_POS                 7                                                                       /**< RX_FRAMING_ERR Field Position.     */
#define MXC_F_UART_INTFL_RX_FRAMING_ERR                     ((uint32_t)(0x00000001UL << MXC_F_UART_INTFL_RX_FRAMING_ERR_POS))       /**< RX_FRAMING_ERR Field Mask.         */
#define MXC_F_UART_INTFL_RX_PARITY_ERR_POS                  8                                                                       /**< RX_PARITY_ERR Field Position.      */
#define MXC_F_UART_INTFL_RX_PARITY_ERR                      ((uint32_t)(0x00000001UL << MXC_F_UART_INTFL_RX_PARITY_ERR_POS))        /**< RX_PARITY_ERR Field Mask.          */
/** @} */

/**
 * @ingroup uart_registers
 * @defgroup UART_INTEN_Register UART_INTEN
 * @brief    UART_INTEN register fields and masks
 * @{
 */
#define MXC_F_UART_INTEN_TX_DONE_POS                        0                                                                       /**< TX_DONE Field Position.                   */
#define MXC_F_UART_INTEN_TX_DONE                            ((uint32_t)(0x00000001UL << MXC_F_UART_INTEN_TX_DONE_POS))              /**< TX_DONE Field Mask.           */
#define MXC_F_UART_INTEN_TX_UNSTALLED_POS                   1                                                                       /**< TX_UNSTALLED Field Position.              */
#define MXC_F_UART_INTEN_TX_UNSTALLED                       ((uint32_t)(0x00000001UL << MXC_F_UART_INTEN_TX_UNSTALLED_POS))         /**< TX_UNSTALLED Field Mask.      */
#define MXC_F_UART_INTEN_TX_FIFO_AE_POS                     2                                                                       /**< TX_FIFO_AE Field Position.                */
#define MXC_F_UART_INTEN_TX_FIFO_AE                         ((uint32_t)(0x00000001UL << MXC_F_UART_INTEN_TX_FIFO_AE_POS))           /**< TX_FIFO_AE Field Mask.        */
#define MXC_F_UART_INTEN_RX_FIFO_NOT_EMPTY_POS              3                                                                       /**< RX_FIFO_NOT_EMPTY Field Position.         */
#define MXC_F_UART_INTEN_RX_FIFO_NOT_EMPTY                  ((uint32_t)(0x00000001UL << MXC_F_UART_INTEN_RX_FIFO_NOT_EMPTY_POS))    /**< RX_FIFO_NOT_EMPTY Field Mask. */
#define MXC_F_UART_INTEN_RX_STALLED_POS                     4                                                                       /**< RX_STALLED_ Field Position.               */
#define MXC_F_UART_INTEN_RX_STALLED                         ((uint32_t)(0x00000001UL << MXC_F_UART_INTEN_RX_STALLED_POS))           /**< RX_STALLED_ Field Mask.       */
#define MXC_F_UART_INTEN_RX_FIFO_AF_POS                     5                                                                       /**< RX_FIFO_AF Field Position.                */
#define MXC_F_UART_INTEN_RX_FIFO_AF                         ((uint32_t)(0x00000001UL << MXC_F_UART_INTEN_RX_FIFO_AF_POS))           /**< RX_FIFO_AF Field Mask.        */
#define MXC_F_UART_INTEN_RX_FIFO_OVERFLOW_POS               6                                                                       /**< RX_FIFO_OVERFLOW Field Position.          */
#define MXC_F_UART_INTEN_RX_FIFO_OVERFLOW                   ((uint32_t)(0x00000001UL << MXC_F_UART_INTEN_RX_FIFO_OVERFLOW_POS))     /**< RX_FIFO_OVERFLOW Field Mask.  */
#define MXC_F_UART_INTEN_RX_FRAMING_ERR_POS                 7                                                                       /**< RX_FRAMING_ERR Field Position.            */
#define MXC_F_UART_INTEN_RX_FRAMING_ERR                     ((uint32_t)(0x00000001UL << MXC_F_UART_INTEN_RX_FRAMING_ERR_POS))       /**< RX_FRAMING_ERR Field Mask.    */
#define MXC_F_UART_INTEN_RX_PARITY_ERR_POS                  8                                                                       /**< RX_PARITY_ERR Field Position.             */
#define MXC_F_UART_INTEN_RX_PARITY_ERR                      ((uint32_t)(0x00000001UL << MXC_F_UART_INTEN_RX_PARITY_ERR_POS))        /**< RX_PARITY_ERR Field Mask.     */
/** @} */

#if (MXC_UART_REV > 0)
/**
 * @ingroup uart_registers
 * @defgroup UART_IDLE_Register UART_IDLE
 * @brief    UART_IDLE register fields and masks
 * @{
 */
#define MXC_F_UART_IDLE_TX_RX_IDLE_POS                      0                                                                       /**< TX_RX_IDLE Field Position.             */
#define MXC_F_UART_IDLE_TX_RX_IDLE                          ((uint32_t)(0x00000001UL << MXC_F_UART_IDLE_TX_RX_IDLE_POS))            /**< TX_RX_IDLE Field Mask.                 */
#define MXC_F_UART_IDLE_TX_IDLE_POS                         1                                                                       /**< TX_IDLE Field Position.                */
#define MXC_F_UART_IDLE_TX_IDLE                             ((uint32_t)(0x00000001UL << MXC_F_UART_IDLE_TX_IDLE_POS))               /**< TX_IDLE Field Mask.                    */
#define MXC_F_UART_IDLE_RX_IDLE_POS                         2                                                                       /**< RX_IDLE Field Position.                */
#define MXC_F_UART_IDLE_RX_IDLE                             ((uint32_t)(0x00000001UL << MXC_F_UART_IDLE_RX_IDLE_POS))               /**< RX_IDLE Field Mask.                    */
/** @} */
#endif
/**
 * @ingroup UART_CTRL_register
 * @defgroup UART_DATA_SIZE_Values Data Size Settings
 * @brief    UART Data Size Field Settings and Masks
 * @{
 */
#define MXC_V_UART_CTRL_DATA_SIZE_5_BITS                                        ((uint32_t)(0x00000000UL))  /**< Value to set to 5 data bits. */
#define MXC_V_UART_CTRL_DATA_SIZE_6_BITS                                        ((uint32_t)(0x00000001UL))  /**< Value to set to 6 data bits. */
#define MXC_V_UART_CTRL_DATA_SIZE_7_BITS                                        ((uint32_t)(0x00000002UL))  /**< Value to set to 7 data bits. */
#define MXC_V_UART_CTRL_DATA_SIZE_8_BITS                                        ((uint32_t)(0x00000003UL))  /**< Value to set to 8 data bits. */

#define MXC_S_UART_CTRL_DATA_SIZE_5_BITS                                        ((uint32_t)(MXC_V_UART_CTRL_DATA_SIZE_5_BITS   << MXC_F_UART_CTRL_DATA_SIZE_POS))   /**< Shifted value to set to 5 data bits. */
#define MXC_S_UART_CTRL_DATA_SIZE_6_BITS                                        ((uint32_t)(MXC_V_UART_CTRL_DATA_SIZE_6_BITS   << MXC_F_UART_CTRL_DATA_SIZE_POS))   /**< Shifted value to set to 6 data bits. */
#define MXC_S_UART_CTRL_DATA_SIZE_7_BITS                                        ((uint32_t)(MXC_V_UART_CTRL_DATA_SIZE_7_BITS   << MXC_F_UART_CTRL_DATA_SIZE_POS))   /**< Shifted value to set to 7 data bits. */
#define MXC_S_UART_CTRL_DATA_SIZE_8_BITS                                        ((uint32_t)(MXC_V_UART_CTRL_DATA_SIZE_8_BITS   << MXC_F_UART_CTRL_DATA_SIZE_POS))   /**< Shifted value to set to 8 data bits. */
/** @} */
/**
 * @ingroup UART_CTRL_register
 * @defgroup UART_PARITY_Values Parity Settings
 * @brief    UART Parity Field Settings and Masks
 * @{
 */
#define MXC_V_UART_CTRL_PARITY_DISABLE                                          ((uint32_t)(0x00000000UL))  /**< Parity disabled value. */
#define MXC_V_UART_CTRL_PARITY_ODD                                              ((uint32_t)(0x00000001UL))  /**< Odd Parity value.      */
#define MXC_V_UART_CTRL_PARITY_EVEN                                             ((uint32_t)(0x00000002UL))  /**< Even Parity value.     */
#define MXC_V_UART_CTRL_PARITY_MARK                                             ((uint32_t)(0x00000003UL))  /**< Mark Parity value.     */

#define MXC_S_UART_CTRL_PARITY_DISABLE                                          ((uint32_t)(MXC_V_UART_CTRL_PARITY_DISABLE  << MXC_F_UART_CTRL_PARITY_POS))  /**< Parity disabled shifted value. */
#define MXC_S_UART_CTRL_PARITY_ODD                                              ((uint32_t)(MXC_V_UART_CTRL_PARITY_ODD      << MXC_F_UART_CTRL_PARITY_POS))  /**< Odd Parity shifted value.      */
#define MXC_S_UART_CTRL_PARITY_EVEN                                             ((uint32_t)(MXC_V_UART_CTRL_PARITY_EVEN     << MXC_F_UART_CTRL_PARITY_POS))  /**< Even Parity shifted value.     */
#define MXC_S_UART_CTRL_PARITY_MARK                                             ((uint32_t)(MXC_V_UART_CTRL_PARITY_MARK     << MXC_F_UART_CTRL_PARITY_POS))  /**< Mark Parity shifted value.     */
/** @} */

/** @} */
#ifdef __cplusplus
}
#endif

#endif   /* _MXC_UART_REGS_H_ */

