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
 * $Date: 2016-04-27 11:55:43 -0500 (Wed, 27 Apr 2016) $
 * $Revision: 22541 $
 *
 ******************************************************************************/

/**
 * @file    uart.h
 * @brief   UART driver header file.
 */

#include "mxc_config.h"
#include "mxc_sys.h"
#include "uart_regs.h"

#ifndef _UART_H_
#define _UART_H_

#ifdef __cplusplus
extern "C" {
#endif

/***** Definitions *****/

/// @brief Defines number of data bits per transmission/reception
typedef enum {
    UART_DATA_SIZE_5_BITS = MXC_V_UART_CTRL_DATA_SIZE_5_BITS,
    UART_DATA_SIZE_6_BITS = MXC_V_UART_CTRL_DATA_SIZE_6_BITS,
    UART_DATA_SIZE_7_BITS = MXC_V_UART_CTRL_DATA_SIZE_7_BITS,
    UART_DATA_SIZE_8_BITS = MXC_V_UART_CTRL_DATA_SIZE_8_BITS
}
uart_data_size_t;

/// @brief Defines number of data bits per transmission/reception
typedef enum {
    UART_PARITY_DISABLE = MXC_V_UART_CTRL_PARITY_DISABLE,
    UART_PARITY_ODD     = MXC_V_UART_CTRL_PARITY_ODD,
    UART_PARITY_EVEN    = MXC_V_UART_CTRL_PARITY_EVEN,
    UART_PARITY_MARK    = MXC_V_UART_CTRL_PARITY_MARK
} uart_parity_t;

/// @brief  UART configuration type.
typedef struct {
    uint8_t extra_stop;             ///< 0 for one stop bit, 1 for two stop bits.
    uint8_t cts;                    ///< 1 to enable CTS.
    uint8_t rts;                    ///< 1 to enable RTS.
    uint32_t baud;                  ///< Baud rate in Hz.
    uart_data_size_t size;          ///< Number of bits in each character.
    uart_parity_t parity;           ///< Even or odd parity.
} uart_cfg_t;

/// @brief UART Transaction request, must remain allocated until callback.
typedef struct uart_req uart_req_t;
struct uart_req {
    uint8_t *data;    ///< Data buffer for characters.
    unsigned len;     ///< Length of characters in data to send or receive.
    unsigned num;     ///< Number of characters actually sent or received.

    /**
     * @brief   Callback for asynchronous request.
     * @param   uart_req_t*     Pointer to the transaction request.
     * @param   int             Error code.
     */
    void (*callback)(uart_req_t*, int);
};

/***** Globals *****/

/***** Function Prototypes *****/

/**
 * @brief   Initialize and enable UART module.
 * @param   uart        Pointer to UART regs.
 * @param   cfg         Pointer to UART configuration.
 * @param   sys_cfg     Pointer to system configuration object
 * @returns #E_NO_ERROR if everything is successful
 */
int UART_Init(mxc_uart_regs_t *uart, const uart_cfg_t *cfg, const sys_cfg_uart_t *sys_cfg);

/**
 * @brief   Shutdown UART module.
 * @param   uart    Pointer to UART regs.
 * @returns #E_NO_ERROR if everything is successful
 */
int UART_Shutdown(mxc_uart_regs_t *uart);

/**
 * @brief   Write UART data. Will block until transaction is complete.
 * @param   uart    Pointer to UART regs.
 * @param   data    Pointer to buffer for write data.
 * @param   len     Number of bytes to write.
 * @note    Will return once data has been put into FIFO, not necessarily transmitted.
 * @returns Number of bytes written if successful, error if unsuccessful.
 */
int UART_Write(mxc_uart_regs_t *uart, uint8_t* data, int len);

/**
 * @brief   Read UART data. Will block until transaction is complete.
 * @param   uart    Pointer to UART regs.
 * @param   data    Pointer to buffer for read data.
 * @param   len     Number of bytes to read.
 * @param   num     Optional pointer to number of bytes actually read.
 *                  Pass NULL if undesired.
 * @returns Number of bytes read is successful, error if unsuccessful.
 */
int UART_Read(mxc_uart_regs_t *uart, uint8_t* data, int len, int *num);

/**
 * @brief   Asynchronously Write UART data.
 * @param   uart    Pointer to UART regs.
 * @param   req     Request for a UART transaction.
 * @note    Request struct must remain allocated until callback.
 * @returns #E_NO_ERROR if everything is successful, error if unsuccessful.
 */
int UART_WriteAsync(mxc_uart_regs_t *uart, uart_req_t *req);

/**
 * @brief   Asynchronously Read UART data.
 * @param   uart    Pointer to UART regs.
 * @param   req     Pointer to request for a UART transaction.
 * @note    Request struct must remain allocated until callback.
 * @returns #E_NO_ERROR if everything is successful, error if unsuccessful.
 */
int UART_ReadAsync(mxc_uart_regs_t *uart, uart_req_t *req);

/**
 * @brief   Abort asynchronous request.
 * @param   req     Pointer to request for a UART transaction.
 * @returns #E_NO_ERROR if request aborted, error if unsuccessful.
 */
int UART_AbortAsync(uart_req_t *req);

/**
 * @brief   UART interrupt handler.
 * @details This function should be called by the application from the interrupt
 *          handler if UART interrupts are enabled. Alternately, this function
 *          can be periodically called by the application if UART interrupts are
 *          disabled. Only necessary to call this when using asynchronous functions.
 * @param   uart     Pointer to UART regs.
 */
void UART_Handler(mxc_uart_regs_t *uart);

/**
 * @brief   Check to see if the UART is busy.
 * @param   uart     Pointer to UART regs.
 * @returns #E_NO_ERROR if idle, #E_BUSY if in use.
 */
int UART_Busy(mxc_uart_regs_t *uart);

/**
 * @brief   Attempt to prepare the UART for sleep.
 * @param   uart     Pointer to UART regs.
 * @details Checks for any ongoing transactions. Disables interrupts if the I2CM
            is idle.
 * @returns #E_NO_ERROR if ready to sleep, #E_BUSY if not ready for sleep.
 */
int UART_PrepForSleep(mxc_uart_regs_t *uart);

/**
 * @brief   Enables the UART without overwriting existing configuration.
 * @param   uart    Pointer to UART regs.
 */
__STATIC_INLINE void UART_Enable(mxc_uart_regs_t *uart)
{
    uart->ctrl |= (MXC_F_UART_CTRL_UART_EN | MXC_F_UART_CTRL_TX_FIFO_EN |
                   MXC_F_UART_CTRL_RX_FIFO_EN);
}

/**
 * @brief   Drain all of the data in the RXFIFO.
 * @param   uart    Pointer to UART regs.
 */
__STATIC_INLINE void UART_DrainRX(mxc_uart_regs_t *uart)
{
    uint32_t ctrl_save = uart->ctrl;
    uart->ctrl = (ctrl_save & ~MXC_F_UART_CTRL_RX_FIFO_EN);
    uart->ctrl = ctrl_save;
}

/**
 * @brief   Drain all of the data in the TXFIFO.
 * @param   uart    Pointer to UART regs.
 */
__STATIC_INLINE void UART_DrainTX(mxc_uart_regs_t *uart)
{
    uint32_t ctrl_save = uart->ctrl;
    uart->ctrl = (ctrl_save & ~MXC_F_UART_CTRL_TX_FIFO_EN);
    uart->ctrl = ctrl_save;
}

/**
 * @brief   Write FIFO availability.
 * @param   uart    Pointer to UART regs.
 * @returns Number of empty bytes available in write FIFO.
 */
__STATIC_INLINE unsigned UART_NumWriteAvail(mxc_uart_regs_t *uart)
{
    return (MXC_UART_FIFO_DEPTH - (uart->tx_fifo_ctrl & MXC_F_UART_TX_FIFO_CTRL_FIFO_ENTRY));
}

/**
 * @brief   Read FIFO availability.
 * @param   uart    Pointer to UART regs.
 * @returns Number of bytes in read FIFO.
 */
__STATIC_INLINE unsigned UART_NumReadAvail(mxc_uart_regs_t *uart)
{
    return (uart->rx_fifo_ctrl & MXC_F_UART_RX_FIFO_CTRL_FIFO_ENTRY);
}

/**
 * @brief   Clear interrupt flags.
 * @param   uart    Pointer to UART regs.
 * @param   mask    Mask of interrupts to clear.
 */
__STATIC_INLINE void UART_ClearFlags(mxc_uart_regs_t *uart, uint32_t mask)
{
    uart->intfl = mask;
}

/**
 * @brief   Get interrupt flags.
 * @param   uart    Pointer to UART regs.
 * @returns Mask of active flags.
 */
__STATIC_INLINE unsigned UART_GetFlags(mxc_uart_regs_t *uart)
{
    return (uart->intfl);
}

#ifdef __cplusplus
}
#endif

#endif /* _UART_H_ */
