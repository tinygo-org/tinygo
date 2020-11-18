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
 * @file    i2cm.h
 * @brief   I2C Master driver header file.
 */

#ifndef _I2CM_H_
#define _I2CM_H_

/***** Includes *****/
#include "mxc_config.h"
#include "mxc_sys.h"
#include "i2cm_regs.h"

#ifdef __cplusplus
extern "C" {
#endif

/***** Definitions *****/
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
/// @brief I2CM frequencies.
typedef enum {
    I2CM_SPEED_100KHZ = 0,
    I2CM_SPEED_400KHZ
} i2cm_speed_t;

/**
 * Current I2CM speed.
 */
extern i2cm_speed_t current_speed;


/// @brief I2CM Transaction request.
typedef struct i2cm_req i2cm_req_t;
struct i2cm_req {

    /**
     * @details Only supports 7-bit addressing. Driver will shift the address and
     *          add the read bit when necessary.
     */
    uint8_t addr;
    const uint8_t *cmd_data;    ///< Optional command data to write before reading.
    uint32_t cmd_len;           ///< Number of bytes in command.
    uint8_t *data;              ///< Data to write or read.
    uint32_t data_len;          ///< Length of data.
    uint32_t cmd_num;           ///< Number of command bytes sent
    uint32_t data_num;          ///< Number of data bytes sent

    /**
     * @brief   Callback for asynchronous request.
     * @param   i2cm_req_t*  Pointer to the transaction request.
     * @param   int         Error code.
     */
    void (*callback)(i2cm_req_t*, int);
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
 * @brief   Initialize I2CM module.
 * @param   i2cm        Pointer to I2CM regs.
 * @param   cfg         Pointer to I2CM configuration.
 * @param   speed       I2CM frequency.
 * @returns #E_NO_ERROR if everything is successful, error if unsuccessful.
 */
int I2CM_Init(mxc_i2cm_regs_t *i2cm, const sys_cfg_i2cm_t *sys_cfg, i2cm_speed_t speed);

/**
 * @brief   Shutdown I2CM module.
 * @param   i2cm    Pointer to I2CM regs.
 * @returns #E_NO_ERROR if everything is successful, error if unsuccessful.
 */
int I2CM_Shutdown(mxc_i2cm_regs_t *i2cm);

/**
 * @brief   Read I2CM data. Will block until transaction is complete.
 * @param   i2cm        Pointer to I2CM regs.
 * @param   addr        I2C address of the slave.
 * @param   cmd_data    Data to write before reading.
 * @param   cmd_len     Number of bytes to write before reading.
 * @param   data        Where to store read data.
 * @param   len         Number of bytes to read.
 * @details Command is an optional feature where the master will write the cmd_data
 *          before reading from the slave. If command is undesired, leave the pointer
 *          NULL and cmd_len 0. If there is a command, the master will send a
            repeated start before reading. Will block until transaction has completed.
 * @returns Bytes transacted if everything is successful, error if unsuccessful.
 */
int I2CM_Read(mxc_i2cm_regs_t *i2cm, uint8_t addr, const uint8_t *cmd_data,
    uint32_t cmd_len, uint8_t* data, uint32_t len);

/**
 * @brief   Write I2CM data. Will block until transaction is complete.
 * @param   i2cm        Pointer to I2CM regs.
 * @param   addr        I2C address of the slave.
 * @param   cmd_data    Data to write before writing data.
 * @param   cmd_len     Number of bytes to write before writing data.
 * @param   data        Data to be written.
 * @param   len         Number of bytes to Write.
 * @details Command is an optional feature where the master will write the cmd_data
 *          before writing to the slave. If command is undesired, leave the pointer
 *          NULL and cmd_len 0. If there is a command, the master will send a
            repeated start before writing again. Will block until transaction has completed.
 * @returns Bytes transacted if everything is successful, error if unsuccessful.
 */
int I2CM_Write(mxc_i2cm_regs_t *i2cm, uint8_t addr, const uint8_t *cmd_data,
    uint32_t cmd_len, uint8_t* data, uint32_t len);

/**
 * @brief   Asynchronously read I2CM data.
 * @param   i2cm    Pointer to I2CM regs.
 * @param   req     Request for an I2CM transaction.
 * @returns #E_NO_ERROR if everything is successful, error if unsuccessful.
 */
int I2CM_ReadAsync(mxc_i2cm_regs_t *i2cm, i2cm_req_t *req);

/**
 * @brief   Asynchronously write I2CM data.
 * @param   i2cm    Pointer to I2CM regs.
 * @param   req     Request for an I2CM transaction.
 * @returns #E_NO_ERROR if everything is successful, error if unsuccessful.
 */
int I2CM_WriteAsync(mxc_i2cm_regs_t *i2cm, i2cm_req_t *req);

/**
 * @brief   Abort asynchronous request.
 * @param   req     Pointer to request for a I2CM transaction.
 * @note    Will call the callback for the request.
 * @returns #E_NO_ERROR if request aborted, error if unsuccessful.
 */
int I2CM_AbortAsync(i2cm_req_t *req);

/**
 * @brief   I2CM interrupt handler.
 * @details This function should be called by the application from the interrupt
 *          handler if I2CM interrupts are enabled. Alternately, this function
 *          can be periodically called by the application if I2CM interrupts are
 *          disabled.
 * @param   i2cm    Base address of the I2CM module.
 */
void I2CM_Handler(mxc_i2cm_regs_t *i2cm);

/**
 * @brief   Checks to see if the I2CM is busy.
 * @param   i2cm    Pointer to I2CM regs.
 * @returns #E_NO_ERROR if idle, #E_BUSY if in use.
 */
int I2CM_Busy(mxc_i2cm_regs_t *i2cm);

/**
 * @brief   Attempt to prepare the I2CM for sleep.
 * @param   i2cm    Pointer to I2CM regs.
 * @details Checks for any ongoing transactions. Disables interrupts if the I2CM
            is idle.
 * @returns #E_NO_ERROR if ready to sleep, #E_BUSY if not ready for sleep.
 */
int I2CM_PrepForSleep(mxc_i2cm_regs_t *i2cm);

/**
 * @brief   Check the I2C bus.
 * @param   i2cm    Pointer to I2CM regs.
 * @details Checks the I2CM bus to determine if there is any other master using
 *          the bus.
 * @returns #E_NO_ERROR if SCL and SDA are high, #E_BUSY otherwise.
 */
int I2CM_BusCheck(mxc_i2cm_regs_t *i2cm);

/**
 * @brief   Drain all of the data in the RXFIFO.
 * @param   i2cm    Pointer to UART regs.
 */
__STATIC_INLINE void I2CM_DrainRX(mxc_i2cm_regs_t *i2cm)
{
    i2cm->ctrl &= ~(MXC_F_I2CM_CTRL_RX_FIFO_EN);
    i2cm->ctrl |= MXC_F_I2CM_CTRL_RX_FIFO_EN;
}

/**
 * @brief   Drain all of the data in the TXFIFO.
 * @param   i2cm    Pointer to UART regs.
 */
__STATIC_INLINE void I2CM_DrainTX(mxc_i2cm_regs_t *i2cm)
{
    i2cm->ctrl &= ~(MXC_F_I2CM_CTRL_TX_FIFO_EN);
    i2cm->ctrl |= MXC_F_I2CM_CTRL_TX_FIFO_EN;
}

/**
 * @brief   Clear interrupt flags.
 * @param   i2cm    Pointer to I2CM regs.
  * @param   mask    Mask of interrupts to clear.
 */
__STATIC_INLINE void I2CM_ClearFlags(mxc_i2cm_regs_t *i2cm, uint32_t mask)
{
    i2cm->intfl = mask;
}

/**
 * @brief   Get interrupt flags.
 * @param   i2cm    Pointer to I2CM regs.
 * @returns Mask of active flags.
 */
__STATIC_INLINE unsigned I2CM_GetFlags(mxc_i2cm_regs_t *i2cm)
{
    return(i2cm->intfl);
}

#ifdef __cplusplus
}
#endif

#endif /* _I2CM_H_ */
