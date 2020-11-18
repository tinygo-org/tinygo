/**
 * @file
 * @brief      Function implementations for the UART serial communications
 *             peripheral module.
 */
/* *****************************************************************************
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
 * $Date: 2017-02-16 08:57:56 -0600 (Thu, 16 Feb 2017) $
 * $Revision: 26458 $
 *
 **************************************************************************** */

/* **** Includes **** */
#include <string.h>
#include "mxc_config.h"
#include "mxc_assert.h"
#include "mxc_lock.h"
#include "mxc_sys.h"
#include "uart.h"
 


/* **** Definitions **** */
///@cond
#define UART_ERRORS             (MXC_F_UART_INTEN_RX_FIFO_OVERFLOW  | \
                                MXC_F_UART_INTEN_RX_FRAMING_ERR | \
                                MXC_F_UART_INTEN_RX_PARITY_ERR)

#define UART_READ_INTS          (MXC_F_UART_INTEN_RX_FIFO_AF |  \
                                MXC_F_UART_INTEN_RX_FIFO_NOT_EMPTY | \
                                MXC_F_UART_INTEN_RX_STALLED | \
                                UART_ERRORS)

#define UART_WRITE_INTS         (MXC_F_UART_INTEN_TX_UNSTALLED | \
                                MXC_F_UART_INTEN_TX_FIFO_AE)

#define UART_RXFIFO_USABLE     (MXC_UART_FIFO_DEPTH-3)
///@endcond

/**
 * @ingroup uart_top
 * @{
 */
/* **** Globals **** */

/**
 * Saves the state of the non-blocking/asynchronous read requests
 */
static uart_req_t *rx_states[MXC_CFG_UART_INSTANCES];

/**
 * Saves the state of the non-blocking/asynchronous write requests
 */
static uart_req_t *tx_states[MXC_CFG_UART_INSTANCES];

/* **** Functions **** */
/**
 * @brief      Asynchronous low-level handler for UART writes. 
 *
 * @param      uart      Pointer to the UART registers structure for the port being written.
 * @param      req       Pointer to the request structure containing the data left to be written.
 * @param[in]  uart_num  The UART number
 */
static void UART_WriteHandler(mxc_uart_regs_t *uart, uart_req_t *req, int uart_num);
/**
 * @brief      Asynchronous low-level hander for UART reads. 
 *
 * @param      uart      Pointer to the UART registers structure for the port being read.
 * @param      req       Pointer to the request structure containing the data storage for the read.
 * @param[in]  uart_num  The UART number
 * @param[in]  flags     The UART interrupt flags for error detection and handling. 
 */
static void UART_ReadHandler(mxc_uart_regs_t *uart, uart_req_t *req, int uart_num, 
    uint32_t flags);

/* ************************************************************************* */
int UART_Init(mxc_uart_regs_t *uart, const uart_cfg_t *cfg, const sys_cfg_uart_t *sys_cfg)
{
    int err;
    int uart_num;
    uint32_t uart_clk;
    uint8_t baud_shift;
    uint16_t baud_div;
    uint32_t baud, diff_baud;
    uint32_t baud_1, diff_baud_1;

    // Check the input parameters
    uart_num = MXC_UART_GET_IDX(uart);
    MXC_ASSERT(uart_num >= 0);

    // Set system level configurations
    if(sys_cfg != NULL) {
        if ((err = SYS_UART_Init(uart, cfg, sys_cfg)) != E_NO_ERROR) {
            return err;
        }
    }

    // Initialize state pointers
    rx_states[uart_num] = NULL;
    tx_states[uart_num] = NULL;

    // Drain FIFOs and enable UART
    uart->ctrl = 0;
    uart->ctrl = (MXC_F_UART_CTRL_UART_EN | MXC_F_UART_CTRL_TX_FIFO_EN |
                  MXC_F_UART_CTRL_RX_FIFO_EN | 
                  (UART_RXFIFO_USABLE <<  MXC_F_UART_CTRL_RTS_LEVEL_POS));

    // Configure data size, stop bit, parity, cts, and rts
    uart->ctrl |= ((cfg->size << MXC_F_UART_CTRL_DATA_SIZE_POS) |
                   (cfg->extra_stop << MXC_F_UART_CTRL_EXTRA_STOP_POS) |
                   (cfg->parity << MXC_F_UART_CTRL_PARITY_POS) |
                   (cfg->cts << MXC_F_UART_CTRL_CTS_EN_POS) |
                   (cfg->rts << MXC_F_UART_CTRL_RTS_EN_POS));

    // Configure the baud rate and divisor
    uart_clk = SYS_UART_GetFreq(uart);
    MXC_ASSERT(uart_clk > 0);

    baud_shift = 2;
    baud_div = (uart_clk / (cfg->baud * 4));

    // Can not support higher frequencies
    if(!baud_div) {
        return E_NOT_SUPPORTED;
    }

    // Decrease the divisor if baud_div is overflowing
    while(baud_div > 0xFF) {
        if(baud_shift == 0) {
            return E_NOT_SUPPORTED;
        }
        baud_shift--;
        baud_div = (uart_clk / (cfg->baud * (16 >> baud_shift)));
    }

    // Adjust baud_div so we don't overflow with the calculations below
    if(baud_div == 0xFF) {
        baud_div = 0xFE;
    }
    if(baud_div == 0) {
        baud_div = 1;
    }

    // Figure out if the truncation increased the error
    baud = (uart_clk / (baud_div * (16 >> baud_shift)));
    baud_1 = (uart_clk / ((baud_div+1) * (16 >> baud_shift)));

    if(cfg->baud > baud) {
        diff_baud = cfg->baud - baud;
    } else {
        diff_baud = baud - cfg->baud;
    }

    if(cfg->baud > baud_1) {
        diff_baud_1 = cfg->baud - baud_1;
    } else {
        diff_baud_1 = baud_1 - cfg->baud;
    }

    if(diff_baud < diff_baud_1) {
        uart->baud = ((baud_div & MXC_F_UART_BAUD_BAUD_DIVISOR) |
                      (baud_shift << MXC_F_UART_BAUD_BAUD_MODE_POS));
    } else {
        uart->baud = (((baud_div+1) & MXC_F_UART_BAUD_BAUD_DIVISOR) |
                      (baud_shift << MXC_F_UART_BAUD_BAUD_MODE_POS));
    }

    return E_NO_ERROR;
}

/* ************************************************************************* */
int UART_Shutdown(mxc_uart_regs_t *uart)
{
    int uart_num, err;
    uart_req_t *temp_req;

    uart_num = MXC_UART_GET_IDX(uart);
    MXC_ASSERT(uart_num >= 0);

    // Disable and clear interrupts
    uart->inten = 0;
    uart->intfl = uart->intfl;

    // Disable UART and FIFOS
    uart->ctrl &= ~(MXC_F_UART_CTRL_UART_EN | MXC_F_UART_CTRL_TX_FIFO_EN |
                    MXC_F_UART_CTRL_RX_FIFO_EN);

    // Call all of the pending callbacks for this UART
    if(rx_states[uart_num] != NULL) {

        // Save the request
        temp_req = rx_states[uart_num];

        // Unlock this UART to read
        mxc_free_lock((uint32_t*)&rx_states[uart_num]);

        // Callback if not NULL
        if(temp_req->callback != NULL) {
            temp_req->callback(temp_req, E_SHUTDOWN);
        }
    }

    if(tx_states[uart_num] != NULL) {

        // Save the request
        temp_req = tx_states[uart_num];

        // Unlock this UART to write
        mxc_free_lock((uint32_t*)&tx_states[uart_num]);

        // Callback if not NULL
        if(temp_req->callback != NULL) {
            temp_req->callback(temp_req, E_SHUTDOWN);
        }
    }

    // Clears system level configurations
    if ((err = SYS_UART_Shutdown(uart)) != E_NO_ERROR) {
        return err;
    }

    return E_NO_ERROR;
}

/* ************************************************************************* */
int UART_Write(mxc_uart_regs_t *uart, uint8_t* data, int len)
{
    int num, uart_num;
    mxc_uart_fifo_regs_t *fifo;

    uart_num = MXC_UART_GET_IDX(uart);
    MXC_ASSERT(uart_num >= 0);

    if(data == NULL) {
        return E_NULL_PTR;
    }

    // Make sure the UART has been initialized
    if(!(uart->ctrl & MXC_F_UART_CTRL_UART_EN)) {
        return E_UNINITIALIZED;
    }

    if(!(len > 0)) {
        return E_NO_ERROR;
    }

    // Lock this UART from writing
    while(mxc_get_lock((uint32_t*)&tx_states[uart_num], 1) != E_NO_ERROR) {}

    // Get the FIFO for this UART
    fifo = MXC_UART_GET_FIFO(uart_num);

    num = 0;

    while(num < len) {

        // Wait for TXFIFO to not be full
        while((uart->tx_fifo_ctrl & MXC_F_UART_TX_FIFO_CTRL_FIFO_ENTRY) ==
                MXC_F_UART_TX_FIFO_CTRL_FIFO_ENTRY) {}

        // Write the data to the FIFO
#if(MXC_UART_REV == 0)
        uart->intfl = MXC_F_UART_INTFL_TX_DONE;
#endif
        fifo->tx = data[num++];
    }

    // Unlock this UART to write
    mxc_free_lock((uint32_t*)&tx_states[uart_num]);

    return num;
}

/* ************************************************************************* */
int UART_Read(mxc_uart_regs_t *uart, uint8_t* data, int len, int *num)
{
    int num_local, remain, uart_num;
    mxc_uart_fifo_regs_t *fifo;

    uart_num = MXC_UART_GET_IDX(uart);
    MXC_ASSERT(uart_num >= 0);
    
    if(data == NULL) {
        return E_NULL_PTR;
    }

    // Make sure the UART has been initialized
    if(!(uart->ctrl & MXC_F_UART_CTRL_UART_EN)) {
        return E_UNINITIALIZED;
    }

    if(!(len > 0)) {
        return E_NO_ERROR;
    }

    // Lock this UART from reading
    while(mxc_get_lock((uint32_t*)&rx_states[uart_num], 1) != E_NO_ERROR) {}

    // Get the FIFO for this UART
    fifo = MXC_UART_GET_FIFO(uart_num);

    num_local = 0;
    remain = len;
    while(remain) {

        // Save the data in the FIFO
        while((uart->rx_fifo_ctrl & MXC_F_UART_RX_FIFO_CTRL_FIFO_ENTRY) && remain) {
            data[num_local] = fifo->rx;
            num_local++;
            remain--;
        }

        // Break if there is an error
        if(uart->intfl & UART_ERRORS) {
            break;
        }
    }

    // Save the number of bytes read if pointer is valid
    if(num != NULL) {
        *num = num_local;
    }

    // Check for errors
    if(uart->intfl & MXC_F_UART_INTFL_RX_FIFO_OVERFLOW) {

        // Clear errors and return error code
        uart->intfl = UART_ERRORS;


        // Unlock this UART to read
        mxc_free_lock((uint32_t*)&rx_states[uart_num]);

        return E_OVERFLOW;

    } else if(uart->intfl & (MXC_F_UART_INTFL_RX_FRAMING_ERR |
                             MXC_F_UART_INTFL_RX_PARITY_ERR)) {

        // Clear errors and return error code
        uart->intfl = UART_ERRORS;


        // Unlock this UART to read
        mxc_free_lock((uint32_t*)&rx_states[uart_num]);

        return E_COMM_ERR;
    }

    // Unlock this UART to read
    mxc_free_lock((uint32_t*)&rx_states[uart_num]);

    return num_local;
}

/* ************************************************************************* */
int UART_WriteAsync(mxc_uart_regs_t *uart, uart_req_t *req)
{
    int uart_num = MXC_UART_GET_IDX(uart);
    MXC_ASSERT(uart_num >= 0);

    // Check the input parameters
    if(req->data == NULL) {
        return E_NULL_PTR;
    }

    // Make sure the UART has been initialized
    if(!(uart->ctrl & MXC_F_UART_CTRL_UART_EN)) {
        return E_UNINITIALIZED;
    }

    if(!(req->len > 0)) {
        return E_NO_ERROR;
    }

    // Attempt to register this write request
    if(mxc_get_lock((uint32_t*)&tx_states[uart_num], (uint32_t)req) != E_NO_ERROR) {
        return E_BUSY;
    }

    // Clear the number of bytes counter
    req->num = 0;

    // Start the write
    UART_WriteHandler(uart, req, uart_num);

    return E_NO_ERROR;
}

/* ************************************************************************* */
int UART_ReadAsync(mxc_uart_regs_t *uart, uart_req_t *req)
{
    int uart_num;
    uint32_t flags;
    
    uart_num = MXC_UART_GET_IDX(uart);
    MXC_ASSERT(uart_num >= 0);

    if(req->data == NULL) {
        return E_NULL_PTR;
    }

    // Make sure the UART has been initialized
    if(!(uart->ctrl & MXC_F_UART_CTRL_UART_EN)) {
        return E_UNINITIALIZED;
    }

    if(!(req->len > 0)) {
        return E_NO_ERROR;
    }

    // Attempt to register this write request
    if(mxc_get_lock((uint32_t*)&rx_states[uart_num], (uint32_t)req) != E_NO_ERROR) {
        return E_BUSY;
    }

    // Clear the number of bytes counter
    req->num = 0;

    // Start the read
    flags = uart->intfl;
    uart->intfl = flags;
    UART_ReadHandler(uart, req, uart_num, flags);

    return E_NO_ERROR;
}

/* ************************************************************************* */
int UART_AbortAsync(uart_req_t *req)
{
    int uart_num;
    
    // Figure out if this was a read or write request, find the request, set to NULL
    for(uart_num = 0; uart_num < MXC_CFG_UART_INSTANCES; uart_num++) {
        if(req == rx_states[uart_num]) {

            // Disable read interrupts, clear flags.
            MXC_UART_GET_UART(uart_num)->inten &= ~UART_READ_INTS;
            MXC_UART_GET_UART(uart_num)->intfl = UART_READ_INTS;

            // Unlock this UART to read
            mxc_free_lock((uint32_t*)&rx_states[uart_num]);

            // Callback if not NULL
            if(req->callback != NULL) {
                req->callback(req, E_ABORT);
            }

            return E_NO_ERROR;
        }

        if(req == tx_states[uart_num]) {

            // Disable write interrupts, clear flags.
            MXC_UART_GET_UART(uart_num)->inten &= ~(UART_WRITE_INTS);
            MXC_UART_GET_UART(uart_num)->intfl = UART_WRITE_INTS;

            // Unlock this UART to write
            mxc_free_lock((uint32_t*)&tx_states[uart_num]);

            // Callback if not NULL
            if(req->callback != NULL) {
                req->callback(req, E_ABORT);
            }

            return E_NO_ERROR;
        }
    }

    return E_BAD_PARAM;
}

/* ************************************************************************* */
void UART_Handler(mxc_uart_regs_t *uart)
{
    int uart_num;
    uint32_t flags;

    uart_num = MXC_UART_GET_IDX(uart);
    MXC_ASSERT(uart_num >= 0);

    flags = uart->intfl;
    uart->intfl = flags;

    // Figure out if this UART has an active Read request
    if((rx_states[uart_num] != NULL) && (flags & UART_READ_INTS)) {
        UART_ReadHandler(uart, rx_states[uart_num], uart_num, flags);
    }

    // Figure out if this UART has an active Write request
    if((tx_states[uart_num] != NULL) && (flags & (UART_WRITE_INTS))) {

        UART_WriteHandler(uart, tx_states[uart_num], uart_num);
    }
}
/* ************************************************************************* */
int UART_Busy(mxc_uart_regs_t *uart)
{
    int uart_num = MXC_UART_GET_IDX(uart);
    MXC_ASSERT(uart_num >= 0);

    // Check to see if there are any ongoing transactions or if the UART is disabled
    if(((tx_states[uart_num] == NULL) &&
            !(uart->tx_fifo_ctrl & MXC_F_UART_TX_FIFO_CTRL_FIFO_ENTRY) &&
#if(MXC_UART_REV == 0)
            (uart->intfl & MXC_F_UART_INTFL_TX_DONE)) ||
#else
            (uart->idle & MXC_F_UART_IDLE_TX_RX_IDLE)) ||
#endif
            !(uart->ctrl & MXC_F_UART_CTRL_UART_EN)) {

        return E_NO_ERROR;
    }

    return E_BUSY;
}

/* ************************************************************************* */
int UART_PrepForSleep(mxc_uart_regs_t *uart)
{
    if(UART_Busy(uart) != E_NO_ERROR) {
        return E_BUSY;
    }

    // Leave read interrupts enabled, if already enabled
    uart->inten &= UART_READ_INTS;
    
    return E_NO_ERROR;
}

/* ************************************************************************* */
static void UART_WriteHandler(mxc_uart_regs_t *uart, uart_req_t *req, int uart_num)
{
    int avail, remain;
    mxc_uart_fifo_regs_t *fifo;

    // Disable write interrupts
    uart->inten &= ~(UART_WRITE_INTS);

    // Get the FIFO for this UART
    fifo = MXC_UART_GET_FIFO(uart_num);

    // Refill the TX FIFO
    avail = UART_NumWriteAvail(uart);
    remain = req->len - req->num;

    while(avail && remain) {

        // Write the data to the FIFO
#if(MXC_UART_REV == 0)
        uart->intfl = MXC_F_UART_INTFL_TX_DONE;
#endif
        fifo->tx = req->data[req->num++];
        remain--;
        avail--;
    }

    // All of the bytes have been written to the FIFO
    if(!remain) {

        // Unlock this UART to write
        mxc_free_lock((uint32_t*)&tx_states[uart_num]);

        if(req->callback != NULL) {
            req->callback(req, E_NO_ERROR);
        }

    } else {

        // Interrupt when there is one byte left in the TXFIFO
        uart->tx_fifo_ctrl = ((MXC_UART_FIFO_DEPTH - 1) << MXC_F_UART_TX_FIFO_CTRL_FIFO_AE_LVL_POS);

        // Enable almost empty interrupt
        uart->inten |= (MXC_F_UART_INTEN_TX_FIFO_AE);
    }
}

/* ************************************************************************* */
static void UART_ReadHandler(mxc_uart_regs_t *uart, uart_req_t *req, int uart_num,
    uint32_t flags)
{
    int avail, remain;
    mxc_uart_fifo_regs_t *fifo;

    // Disable interrupts
    uart->inten &= ~UART_READ_INTS;

    // Get the FIFO for this UART, uart_num
    fifo = MXC_UART_GET_FIFO(uart_num);

    // Save the data in the FIFO while we still need data
    avail = UART_NumReadAvail(uart);
    remain = req->len - req->num;
    while(avail && remain) {
        req->data[req->num++] = fifo->rx;
        remain--;
        avail--;
    }

    // Check for errors
    if(flags & MXC_F_UART_INTFL_RX_FIFO_OVERFLOW) {

        // Unlock this UART to read
        mxc_free_lock((uint32_t*)&rx_states[uart_num]);

        if(req->callback != NULL) {
            req->callback(req, E_OVERFLOW);
        }

        return;
    }

    if(flags & (MXC_F_UART_INTFL_RX_FRAMING_ERR |
                MXC_F_UART_INTFL_RX_PARITY_ERR)) {

        // Unlock this UART to read
        mxc_free_lock((uint32_t*)&rx_states[uart_num]);

        if(req->callback != NULL)         {
            req->callback(req, E_COMM_ERR);
        }

        return;
    }

    // Check to see if we're done receiving
    if(remain == 0) {

        // Unlock this UART to read
        mxc_free_lock((uint32_t*)&rx_states[uart_num]);

        if(req->callback != NULL) {
            req->callback(req, E_NO_ERROR);
        }

        return;
    }

    if(remain == 1) {
        uart->inten |= (MXC_F_UART_INTEN_RX_FIFO_NOT_EMPTY | UART_ERRORS);

    } else {
        // Set the RX FIFO AF threshold
        if(remain < UART_RXFIFO_USABLE) {
            uart->rx_fifo_ctrl = ((remain - 1) << 
                MXC_F_UART_RX_FIFO_CTRL_FIFO_AF_LVL_POS);
        } else {
            uart->rx_fifo_ctrl = (UART_RXFIFO_USABLE <<
                MXC_F_UART_RX_FIFO_CTRL_FIFO_AF_LVL_POS);
        }
        uart->inten |= (MXC_F_UART_INTEN_RX_FIFO_AF | UART_ERRORS);
    }
}
/** @} */
