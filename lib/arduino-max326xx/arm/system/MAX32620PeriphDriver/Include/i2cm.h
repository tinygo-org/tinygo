/**
 * @file
 * @brief      I2CM (Inter-Integrated Circuit Master) function prototypes and
 *             data types.
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
 * $Date: 2017-02-16 12:08:14 -0600 (Thu, 16 Feb 2017) $
 * $Revision: 26469 $
 *
 *************************************************************************** */

/* Define to prevent redundant inclusion */
#ifndef _I2CM_H_
#define _I2CM_H_

/***** Includes *****/
#include "mxc_config.h"
#include "mxc_sys.h"
/**
 * @ingroup    periphlibs
 * @defgroup   i2cm I2C Master
 * @brief      I2C Master Peripheral API
 * @{
 */
#include "i2cm_regs.h"

#ifdef __cplusplus
extern "C" {
#endif

/* **** Definitions **** */
#ifndef MXC_I2CM_TX_TIMEOUT
#define MXC_I2CM_TX_TIMEOUT     0x5000      /**< Master Transmit Timeout in number of repetitive attempts to receive an ACK/NACK or for a transmission to occur */
#endif

#ifndef MXC_I2CM_RX_TIMEOUT
#define MXC_I2CM_RX_TIMEOUT     0x5000      /**< Master Receive Timeout in number of attempts to check FIFO for received data from a slave */
#endif

#define I2CM_READ_BIT           0x0001      /**< Bit location to specify a read for the I2C protocol */
#define I2CM_FIFO_DEPTH_3Q      ((3 * MXC_I2CM_FIFO_DEPTH) / 4)
#define I2CM_FIFO_DEPTH_2Q      (MXC_I2CM_FIFO_DEPTH / 2)

// Arduino's requirement for errors
// Extending mxc_error codes
#define E_NACK_ON_ADDR_ARDUINO -18
#define E_NACK_ON_DATA_ARDUINO -19

extern const uint32_t clk_div_table[3][8];
/**
 * Enumeration type to select supported I2CM frequencies.
 */
typedef enum {
    I2CM_SPEED_100KHZ = 0,      /**< Use to select a bus communication speed of 100 kHz. */
    I2CM_SPEED_400KHZ = 1       /**< Use to select a bus communication speed of 400 kHz. */
} i2cm_speed_t;

/**
 * Current I2CM speed.
 */
extern i2cm_speed_t current_speed;

/**
 * Structure type for an I2CM Transaction request.
 */
typedef struct i2cm_req i2cm_req_t;

/**
 * @addtogroup i2cm_async
 * @{
 * Function type for the I2C Master callback. The function declaration for the
 * I2CM callback is:
 * @code
 * void callback(i2cm_req_t * req, int error_code);
 * @endcode
 * |        |                                            |
 * | -----: | :----------------------------------------- |
 * | @p req |  Pointer to an #i2cm_req object representing the I2CM active transaction. |
 * | @p error_code | An error code if the active transaction had a failure or #E_NO_ERROR if successful. |
 *
 */
typedef void (*i2cm_callback_fn)(i2cm_req_t * req, int  error_code);
/**@}*/


/**
 * I2CM Transaction request structure.
 * @note       Only supports 7-bit addressing. Driver will shift the address and
 *             add the read bit when necessary.
 */
struct i2cm_req {
    uint8_t addr;               /**< 7-Bit unshifted address of the slave for communication.                                                */
    const uint8_t *cmd_data;    /**< Pointer to a command data buffer to send to the slave before either a read or write transaction.       */
    uint32_t cmd_len;           /**< Number of bytes in command.                                                                            */
    uint8_t *data;              /**< Data to write or read.                                                                                 */
    uint32_t data_len;          /**< Length of data.                                                                                        */
    uint32_t cmd_num;           /**< Number of command bytes sent.                                                                          */
    uint32_t data_num;          /**< Number of data bytes sent.                                                                             */
    i2cm_callback_fn callback;  /**< Function pointer to a callback function.                                                               */
};

// Saves the state of the non-blocking requests
typedef enum {
    I2CM_STATE_READING = 0,
    I2CM_STATE_WRITING = 1
} i2cm_state_t;

typedef struct {
    i2cm_req_t *req;
    i2cm_state_t state;
} i2cm_req_state_t;

extern i2cm_req_state_t i2cm_states[MXC_CFG_I2CM_INSTANCES];

/* **** Globals **** */

/* **** Function Prototypes **** */

void I2CM_Recover(mxc_i2cm_regs_t *i2cm);
int I2CM_WriteTxFifo(mxc_i2cm_regs_t *regs, mxc_i2cm_fifo_regs_t *fifo, const uint16_t data);
int I2CM_TxInProgress(mxc_i2cm_regs_t *i2cm);
void I2CM_FreeCallback(int i2cm_num, int error);
int I2CM_Tx(mxc_i2cm_regs_t *i2cm, mxc_i2cm_fifo_regs_t *fifo, uint8_t addr,
    const uint8_t *data, uint32_t len, uint8_t stop);

int I2CM_Rx(mxc_i2cm_regs_t *i2cm, mxc_i2cm_fifo_regs_t *fifo, uint8_t addr,
    uint8_t *data, uint32_t len);

int I2CM_CmdHandler(mxc_i2cm_regs_t *i2cm, mxc_i2cm_fifo_regs_t *fifo, i2cm_req_t *req);
int I2CM_ReadHandler(mxc_i2cm_regs_t *i2cm, i2cm_req_t *req, int i2cm_num);
int I2CM_WriteHandler(mxc_i2cm_regs_t *i2cm, i2cm_req_t *req, int i2cm_num);

/**
 * @brief      Initialize the I2CM peripheral module.
 *
 * @param      i2cm     Pointer to the I2CM registers, see #mxc_i2cm_regs_t.
 * @param      sys_cfg  Pointer to an I2CM configuration structure of type
 *                      #sys_cfg_i2cm_t.
 * @param      speed    I2CM bus speed, see #i2cm_speed_t.
 *
 * @return     #E_NO_ERROR if initialized successfully, error if unsuccessful.
 */
int I2CM_Init(mxc_i2cm_regs_t *i2cm, const sys_cfg_i2cm_t *sys_cfg, i2cm_speed_t speed);

/**
 * @brief   Shutdown I2CM module.
 *
 * @param   i2cm    Pointer to the I2CM registers, see #mxc_i2cm_regs_t.
 *
 * @returns #E_NO_ERROR if everything is successful, error if unsuccessful.
 *
 */
int I2CM_Shutdown(mxc_i2cm_regs_t *i2cm);

/**
 * @defgroup i2cm_blocking I2CM Blocking Functions
 * @{
 */

/**
 * @brief      Read I2CM data. Will block until transaction is complete.
 *
 * @note       Command is an optional feature where the master will write the @c
 *             cmd_data before reading from the slave. If command is undesired,
 *             set the @c *cmd_data parameter to NULL and pass 0 for the @c
 *             cmd_len parameter.
 * @note       If there is a command, the master will send a repeated start
 *             sequence before attempting to read from the slave.
 * @note       This function blocks until the transaction has completed.
 *
 * @param      i2cm      Pointer to the I2CM registers, see #mxc_i2cm_regs_t.
 * @param      addr      I2C address of the slave.
 * @param      cmd_data  Data to write before reading.
 * @param      cmd_len   Number of bytes to write before reading.
 * @param      data      Where to store the data read.
 * @param      len       Number of bytes to read.
 *
 * @return     Number of bytes read if successful, error code if unsuccessful.
 */
int I2CM_Read(mxc_i2cm_regs_t *i2cm, uint8_t addr, const uint8_t *cmd_data,
    uint32_t cmd_len, uint8_t* data, uint32_t len);

/**
 * @brief      Write data to a slave device.
 *
 * @note       Command is an optional feature where the master will write the @c
 *             cmd_data before writing the @c data to the slave. If command is
 *             not needed, set the @c cmd_data  to @c NULL and set @c cmd_len to
 *             0. If there is a command, the master will send a repeated start
 *             sequence before attempting to read from the slave.
 * @note       This function blocks until the transaction has completed.
 *
 * @param      i2cm      Pointer to the I2CM registers, see #mxc_i2cm_regs_t.
 * @param      addr      I2C address of the slave.
 * @param      cmd_data  Data to write before writing data.
 * @param      cmd_len   Number of bytes to write before writing data.
 * @param      data      Data to be written.
 * @param      len       Number of bytes to Write.
 *
 * @return     Number of bytes writen if successful or an @ref MXC_Error_Codes
 *             "Error Code" if unsuccessful.
 */
int I2CM_Write(mxc_i2cm_regs_t *i2cm, uint8_t addr, const uint8_t *cmd_data,
    uint32_t cmd_len, uint8_t* data, uint32_t len);
/**@} end of i2cm_blocking functions */

/**
 * @defgroup i2cm_async I2CM Asynchrous Functions
 * @{
 */

/**
 * @brief      Asynchronously read I2CM data.
 *
 * @param      i2cm  Pointer to the I2CM registers, see #mxc_i2cm_regs_t.
 * @param      req   Pointer to an I2CM transaction request structure, see
 *                   #i2cm_req.
 *
 * @return     #E_NO_ERROR if everything is successful or an @ref
 *             MXC_Error_Codes "Error Code" if unsuccessful.
 */
int I2CM_ReadAsync(mxc_i2cm_regs_t *i2cm, i2cm_req_t *req);

/**
 * @brief      Asynchronously write I2CM data.
 *
 * @param      i2cm  Pointer to the I2CM registers, see #mxc_i2cm_regs_t.
 * @param      req   Pointer to an I2CM transaction request structure, see
 *                   #i2cm_req.
 *
 * @return     #E_NO_ERROR if everything is successful, error if unsuccessful.
 */
int I2CM_WriteAsync(mxc_i2cm_regs_t *i2cm, i2cm_req_t *req);

/**
 * @brief      Abort asynchronous request.
 * @param      req   Pointer to request for an I2CM transaction.
 * @note       Will call the callback for the request.
 *
 * @return     #E_NO_ERROR if request aborted, error if unsuccessful.
 */
int I2CM_AbortAsync(i2cm_req_t *req);

/**
 * @brief      I2CM interrupt handler.
 *
 * @details    This function is an IRQ handler and will be called by the core if
 *             I2CM interrupts are enabled. Alternately, if the application is
 *             using asynchronous methods, this function can be periodically
 *             called by the application if the I2CM interrupts are disabled.
 *
 * @param      i2cm  Base address of the I2CM module.
 */
void I2CM_Handler(mxc_i2cm_regs_t *i2cm);
/**@} end of i2cm_async */

/**
 * @brief      Returns the status of the I2CM peripheral module.
 *
 * @param      i2cm  Pointer to the I2CM register structure, see
 *                   #mxc_i2cm_regs_t.
 *
 * @return     #E_NO_ERROR if idle.
 * @return     #E_BUSY if in use.
 */
int I2CM_Busy(mxc_i2cm_regs_t *i2cm);

/**
 * @brief      Attempt to prepare the I2CM for sleep.
 * @details    Checks for any ongoing transactions. Disables interrupts if the
 *             I2CM is idle.
 *
 * @param      i2cm  Pointer to the I2CM register structure, see
 *                   #mxc_i2cm_regs_t.
 *
 * @return     #E_NO_ERROR if ready to sleep.
 * @return     #E_BUSY if the bus is not ready for sleep.
 */
int I2CM_PrepForSleep(mxc_i2cm_regs_t *i2cm);

/**
 * @brief      Check the I2C bus to determine if any other masters are using the
 *             bus.
 *
 * @param      i2cm  Pointer to the I2CM register structure, see
 *                   #mxc_i2cm_regs_t.
 *
 * @return     #E_NO_ERROR if SCL and SDA are high,
 * @return     #E_BUSY otherwise.
 */
int I2CM_BusCheck(mxc_i2cm_regs_t *i2cm);

/**
 * @brief      Drain/Empty all of the data in the I2CM Receive FIFO.
 *
 * @param      i2cm  Pointer to the I2CM register structure, see
 *                   #mxc_i2cm_regs_t.
 */
__STATIC_INLINE void I2CM_DrainRX(mxc_i2cm_regs_t *i2cm)
{
    i2cm->ctrl &= ~(MXC_F_I2CM_CTRL_RX_FIFO_EN);
    i2cm->ctrl |= MXC_F_I2CM_CTRL_RX_FIFO_EN;
}

/**
 * @brief      Drain/Empty any data in the I2CM Transmit FIFO.
 *
 * @param      i2cm  Pointer to the I2CM register structure, see
 *                   #mxc_i2cm_regs_t.
 */
__STATIC_INLINE void I2CM_DrainTX(mxc_i2cm_regs_t *i2cm)
{
    i2cm->ctrl &= ~(MXC_F_I2CM_CTRL_TX_FIFO_EN);
    i2cm->ctrl |= MXC_F_I2CM_CTRL_TX_FIFO_EN;
}

/**
 * @brief      Clear interrupt flags.
 *
 * @param      i2cm  Pointer to the I2CM register structure, see
 *                   #mxc_i2cm_regs_t.
 * @param      mask  Mask of I2CM interrupts to clear (1 to clear),
 *             @see I2CM_INTFL_Register for the interrupt flag masks.
 */
__STATIC_INLINE void I2CM_ClearFlags(mxc_i2cm_regs_t *i2cm, uint32_t mask)
{
    i2cm->intfl = mask;
}

/**
 * @brief      Gets the current I2CM interrupt flags.
 * @param      i2cm  Pointer to the I2CM register structure, see
 *                   #mxc_i2cm_regs_t.
 *
 * @return     The currently set interrupt flags, @see I2CM_INTFL_Register
 *             for the interrupt flag masks.
 */
__STATIC_INLINE unsigned I2CM_GetFlags(mxc_i2cm_regs_t *i2cm)
{
    return(i2cm->intfl);
}

/**@} end of group i2cm */

#ifdef __cplusplus
}
#endif

#endif /* _I2CM_H_ */
