/**
 * @file
 * @brief UART data types, definitions and function prototypes.
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
 * $Date: 2017-02-27 18:18:01 -0600 (Mon, 27 Feb 2017) $
 * $Revision: 26714 $
 *
 **************************************************************************** */

/* **** Includes **** */ 
#include "mxc_config.h"
#include "mxc_sys.h"
#include "uart_regs.h"

/* Define to prevent redundant inclusion */
#ifndef _UART_H_
#define _UART_H_
/**
 * @ingroup periphlibs
 * @defgroup uart_top UART API and Registers
 * @{
 */
#ifdef __cplusplus
extern "C" {
#endif



/* **** Definitions **** */

/**
 * Enumeration type for defining the number of bits per character. 
 */
typedef enum {
    UART_DATA_SIZE_5_BITS = MXC_V_UART_CTRL_DATA_SIZE_5_BITS, /**< 5 Data bits per UART character. */
    UART_DATA_SIZE_6_BITS = MXC_V_UART_CTRL_DATA_SIZE_6_BITS, /**< 6 Data bits per UART character. */
    UART_DATA_SIZE_7_BITS = MXC_V_UART_CTRL_DATA_SIZE_7_BITS, /**< 7 Data bits per UART character. */
    UART_DATA_SIZE_8_BITS = MXC_V_UART_CTRL_DATA_SIZE_8_BITS  /**< 8 Data bits per UART character. */
}
uart_data_size_t;

/**
 * Enumeration type for selecting Parity and type. 
 */
typedef enum {
    UART_PARITY_DISABLE = MXC_V_UART_CTRL_PARITY_DISABLE,    /**< Parity disabled.  */
    UART_PARITY_ODD     = MXC_V_UART_CTRL_PARITY_ODD,        /**< Odd parity.       */
    UART_PARITY_EVEN    = MXC_V_UART_CTRL_PARITY_EVEN,       /**< Even parity.      */
    UART_PARITY_MARK    = MXC_V_UART_CTRL_PARITY_MARK        /**< Mark parity.      */
} uart_parity_t;

/**
 * Configuration structure type for a UART port. 
 */
typedef struct {
    uint8_t extra_stop;             /**< @c 0 for one stop bit, @c 1 for two stop bits*/
    uint8_t cts;                    /**< CTS Enable/Disable, @c 1 to enable CTS, @c 0 to disable CTS.*/
    uint8_t rts;                    /**< @c 1 to enable RTS, @c 0 to disable RTS. */
    uint32_t baud;                  /**< Baud rate in Hz.                                               */
    uart_data_size_t size;          /**< Set the number of bits per character, see uart_data_size_t.   */
    uart_parity_t parity;           /**< Set the parity, see uart_parity_t for supported parity types. */
} uart_cfg_t;
/** @} */
/**
 * @ingroup uart_top
 * @defgroup uart_async UART Asynchronous Functions
 * @{
 */
  
/**
 * Structure type for a UART asynchronous transaction request.
 */
typedef struct uart_req uart_req_t;

/**
 * @brief   Type alias for a UART callback function used for asynchronous operations. 
 * @details  
 * Type alias \c uart_async_callback for a callback function with signature of: 
 * @code{.c}
 *    void callback_fn(uart_req_t *request , int error_code);
 * @endcode
 * @p uart_req_t *    Pointer to the transaction request.
 * @p error_code      Return code for the UART request. See @ref MXC_Error_Codes for possible codes. 
 */
typedef void (*uart_async_callback)(uart_req_t*, int); 

/**
 * Structure for a UART asynchronous transaction request.
 * @note       When using this structure for an asynchronous operation, the
 *             structure must remain allocated until the callback is completed.
 */
struct uart_req {
    uint8_t *data;                  /**< Data buffer for characters.                                    */
    unsigned len;                   /**< Length of characters in data to send or receive.               */
    unsigned num;                   /**< Number of characters actually sent or received.                */
    uart_async_callback callback;   /**< Pointer to a callback function of type uart_async_callback()   */
};

/** @} */ 
/* **** Globals **** */

/* **** Function Prototypes **** */
/**
 * @ingroup uart_top
 * @{
 */
/**
 * @brief   Initialize and enable UART module.
 * @param   uart        Pointer to the UART registers.
 * @param   cfg         Pointer to UART configuration.
 * @param   sys_cfg     Pointer to system configuration object
 * @returns #E_NO_ERROR UART initialized successfully, @ref MXC_Error_Codes "error" if
 *             unsuccessful.
 */
int UART_Init(mxc_uart_regs_t *uart, const uart_cfg_t *cfg, const sys_cfg_uart_t *sys_cfg);

/**
 * @brief   Shutdown UART module.
 * @param   uart    Pointer to the UART registers.
 * @returns #E_NO_ERROR UART shutdown successfully, @ref MXC_Error_Codes "error" if
 *             unsuccessful.
 */
int UART_Shutdown(mxc_uart_regs_t *uart);
/**@}*/
/**
 * @ingroup uart_top
 * @defgroup uart_sync UART Synchronous Functions
 * @brief Synchronous/blocking Functions for the UART peripheral.
 * @{
 */

/**
 * @brief      Write UART data. This function blocks until the write transaction
 *             is complete.
 * @param      uart  Pointer to the UART registers.
 * @param      data  Pointer to buffer for write data.
 * @param      len   Number of bytes to write.
 * @note       Will return once data has been put into FIFO, not necessarily
 *             transmitted.
 * @return     Number of bytes written if successful, error if unsuccessful.
 */
int UART_Write(mxc_uart_regs_t *uart, uint8_t* data, int len);

/**
 * @brief      Read UART data, <em>blocking</em> until transaction is complete.
 * @param      uart  Pointer to the UART registers.
 * @param      data  Pointer to buffer to save the data read.
 * @param      len   Number of bytes to read.
 * @param      num   Pointer to store the number of bytes actually read, pass NULL if not needed.
 *
 * @return     Number of bytes read, @ref MXC_Error_Codes "error" if
 *             unsuccessful.
 */
int UART_Read(mxc_uart_regs_t *uart, uint8_t* data, int len, int *num);
/** @} */
/**
 * @ingroup uart_async
 * @{
 */
/**
 * @brief      Asynchronously write/transmit data to the UART.
 *
 * @param      uart  Pointer to the UART registers.
 * @param      req   Request for a UART transaction.
 * @note       Request struct must remain allocated until callback.
 *
 * @return     #E_NO_ERROR Asynchronous write successfully started, @ref
 *             MXC_Error_Codes "error" if unsuccessful.
 */
int UART_WriteAsync(mxc_uart_regs_t *uart, uart_req_t *req);

/**
 * Asynchronously Read UART data.
 * @param      uart  Pointer to the UART registers.
 * @param      req   Pointer to request for a UART transaction.
 * 
 * @return     #E_NO_ERROR Asynchronous read successfully started, @ref
 *             MXC_Error_Codes "error" if unsuccessful.
 * @note       Request struct must remain allocated until callback function is called.
 */
int UART_ReadAsync(mxc_uart_regs_t *uart, uart_req_t *req);

/**
 * Abort asynchronous request.
 * @param      req   Pointer to a request for a UART transaction, see #uart_req.
 * @return     #E_NO_ERROR Asynchronous request aborted successfully, error if unsuccessful.
 */
int UART_AbortAsync(uart_req_t *req);

/**
 * @brief      UART interrupt handler.
 * @details    This function should be called by the application from the
 *             interrupt handler if UART interrupts are enabled. Alternately,
 *             this function can be periodically called by the application if
 *             UART interrupts are disabled. Only necessary to call this when
 *             using asynchronous functions.
 *
 * @param      uart  Pointer to the UART registers.
 */
void UART_Handler(mxc_uart_regs_t *uart);

/**
 * @brief      Check to see if the UART is busy.
 *
 * @param      uart  Pointer to the UART registers.
 *
 * @return     #E_NO_ERROR UART is idle.
 * @return     #E_BUSY UART is in use.
 */
int UART_Busy(mxc_uart_regs_t *uart);
/** @}*/
/**
 * @ingroup uart_top
 * @{
 */
/**
 * @brief      Prepare the UART for entry into a Low-Power mode (LP0/LP1).
 * @details    Checks for any ongoing transactions. Disables interrupts if the
 *             UART is idle.
 *
 * @param      uart         Pointer to the UART registers.
 * @return     #E_NO_ERROR  UART is ready to enter Low-Power modes (LP0/LP1).
 * @return     #E_BUSY      UART is active and busy and not ready to enter a
 *                          Low-Power mode (LP0/LP1).
 */
int UART_PrepForSleep(mxc_uart_regs_t *uart);

/**
 * @brief      Enables the UART.
 * @note       This function does not change the existing UART configuration.
 * @param      uart  Pointer to the UART registers.
 */
__STATIC_INLINE void UART_Enable(mxc_uart_regs_t *uart)
{
    uart->ctrl |= (MXC_F_UART_CTRL_UART_EN | MXC_F_UART_CTRL_TX_FIFO_EN |
                   MXC_F_UART_CTRL_RX_FIFO_EN);
}

/**
 * @brief      Drains/empties and data in the RX FIFO.
 * @param      uart  Pointer to the UART registers.
 */
__STATIC_INLINE void UART_DrainRX(mxc_uart_regs_t *uart)
{
    uint32_t ctrl_save = uart->ctrl;
    uart->ctrl = (ctrl_save & ~MXC_F_UART_CTRL_RX_FIFO_EN);
    uart->ctrl = ctrl_save;
}

/**
 * @brief      Drains/empties any data in the TX FIFO.
 * @param      uart  Pointer to the UART registers.
 */
__STATIC_INLINE void UART_DrainTX(mxc_uart_regs_t *uart)
{
    uint32_t ctrl_save = uart->ctrl;
    uart->ctrl = (ctrl_save & ~MXC_F_UART_CTRL_TX_FIFO_EN);
    uart->ctrl = ctrl_save;
}

/**
 * @brief      Returns the number of unused bytes available in the UART TX FIFO.
 * @param      uart  Pointer to the UART registers.
 * @return     Number of unused bytes in the TX FIFO.
 */
__STATIC_INLINE unsigned UART_NumWriteAvail(mxc_uart_regs_t *uart)
{
    return (MXC_UART_FIFO_DEPTH - (uart->tx_fifo_ctrl & MXC_F_UART_TX_FIFO_CTRL_FIFO_ENTRY));
}

/**
 * @brief      Returns the number of bytes available to be read from the RX
 *             FIFO.
 * @param      uart  Pointer to the UART registers.
 * @return     The number of bytes available to read in the RX FIFO.
 */
__STATIC_INLINE unsigned UART_NumReadAvail(mxc_uart_regs_t *uart)
{
    return (uart->rx_fifo_ctrl & MXC_F_UART_RX_FIFO_CTRL_FIFO_ENTRY);
}

/**
 * @brief      Clear interrupt flags.
 * @param      uart  Pointer to the UART registers.
 * @param      mask  Mask of the UART interrupts to clear, see @ref UART_INTFL_Register Register.
 */
__STATIC_INLINE void UART_ClearFlags(mxc_uart_regs_t *uart, uint32_t mask)
{
    uart->intfl = mask;
}

/**
 * @brief      Get interrupt flags.
 * @param      uart  Pointer to the UART registers.
 * @return     Mask of active flags.
 */
__STATIC_INLINE unsigned UART_GetFlags(mxc_uart_regs_t *uart)
{
    return (uart->intfl);
}
/** @} */
#ifdef __cplusplus
}
#endif

#endif /* _UART_H_ */
