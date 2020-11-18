/*******************************************************************************
 * Copyright (C) 2017 Maxim Integrated Products, Inc., All Rights Reserved.
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
 ******************************************************************************/

#include "mxc_lock.h"
#include "mxc_errors.h"
#include "mxc_sys.h"
#include "Wire.h"
#include "i2cm.h"

TwoWire::TwoWire(uint32_t index) :
  idx(index)
{
}

void TwoWire::begin()
{
    rxBufferIndex = 0;
    rxBufferLength = 0;
    txBufferIndex = 0;
    txBufferLength = 0;

    i2cm = MXC_I2CM_GET_I2CM(idx);
    fifo = MXC_I2CM_GET_FIFO(idx);

    sys_cfg_i2cm_t i2cm_sys_cfg;

#ifdef MAX32620
    switch(idx)
    {
        case 0: i2cm_sys_cfg.io_cfg = (ioman_cfg_t)IOMAN_I2CM0(1,0); break;
        case 1: i2cm_sys_cfg.io_cfg = (ioman_cfg_t)IOMAN_I2CM1(1,0); break;
        case 2: i2cm_sys_cfg.io_cfg = (ioman_cfg_t)IOMAN_I2CM2(1,0); break;
    }
#endif

#ifdef MAX32625
    i2cm_sys_cfg.io_cfg = (ioman_cfg_t)IOMAN_I2CM(idx, 1, 1);
#endif

#ifdef MAX32630
    switch (idx) {
        case 0: i2cm_sys_cfg.io_cfg = (ioman_cfg_t)IOMAN_I2CM0(IOMAN_MAP_A, 1); break;
        case 1: i2cm_sys_cfg.io_cfg = (ioman_cfg_t)IOMAN_I2CM1(IOMAN_MAP_A, 1); break;
        case 2: i2cm_sys_cfg.io_cfg = (ioman_cfg_t)IOMAN_I2CM2(IOMAN_MAP_A, 1); break;
    }
#endif

    i2cm_sys_cfg.clk_scale = CLKMAN_SCALE_DIV_1;
    I2CM_Init(i2cm, &i2cm_sys_cfg, I2CM_SPEED_400KHZ);
}

void TwoWire::begin(uint8_t addr)
{
    // TODO implement slave
}

void TwoWire::begin(int addr)
{
    begin((uint8_t)addr);
}

void TwoWire::end(void)
{
    I2CM_Shutdown(i2cm);
}

void TwoWire::setClock(uint32_t clock)
{
    // Default speed
    i2cm_speed_t speed = I2CM_SPEED_400KHZ;


    // Compute clock array index
    int clki = ((SYS_I2CM_GetFreq(i2cm) / 12000000) - 1);

    if (clock == 400000) {
        speed = I2CM_SPEED_400KHZ;
        // Get clock divider settings from lookup table
        if (clk_div_table[I2CM_SPEED_400KHZ][clki] > 0) {
            i2cm->fs_clk_div = clk_div_table[I2CM_SPEED_400KHZ][clki];
        }
    } else if (clock == 100000) {
        speed = I2CM_SPEED_100KHZ;
        if (clk_div_table[I2CM_SPEED_100KHZ][clki] > 0) {
            i2cm->fs_clk_div = clk_div_table[I2CM_SPEED_100KHZ][clki];
        }
    }

    // Update current_speed if clock is valid, or set to default speed
    current_speed = speed;

}

uint8_t TwoWire::requestFrom(uint8_t address, uint8_t quantity, uint8_t sendStop)
{
    // This method returns 0 incase of any error
    // To match the Arduino API compatibility, return type is uint8_t

    int i2cm_num;
    int error = E_NO_ERROR;
    mxc_i2cm_fifo_regs_t *fifo;

    if (quantity > BUFFER_LENGTH) {
        quantity = BUFFER_LENGTH;
    }

    rxBufferIndex = 0;

    if (rxBuffer == NULL) {
        // Null Pointer error
        return 0;
    }

    // Make sure the I2CM has been initialized
    if (i2cm->ctrl == 0) {
        // Uninitialized error
        return 0;
    }

    if (!(quantity > 0)) {
        return E_NO_ERROR;
    }

    // Lock this I2CM
    i2cm_num = MXC_I2CM_GET_IDX(i2cm);
    while (mxc_get_lock((uint32_t*)&i2cm_states[i2cm_num].req,1) != E_NO_ERROR);

    // Get the FIFO pointer for this I2CM
    fifo = MXC_I2CM_GET_FIFO(i2cm_num);

    // Disable and clear the interrupts
	i2cm->inten = 0;
    i2cm->intfl = i2cm->intfl;

    // Write the address to the TXFIFO
    error = I2CM_WriteTxFifo(i2cm, fifo, (MXC_S_I2CM_TRANS_TAG_START |
                                        (address << 1) | I2CM_READ_BIT));

    uint8_t i = quantity;
    // Write to the TXFIFO the number of bytes we want to read
    while ((i > 256) && !error) {
        error = I2CM_WriteTxFifo(i2cm, fifo, (MXC_S_I2CM_TRANS_TAG_RXDATA_COUNT | 255));
        i -= 256;
    }

    if ((i > 1) && !error) {
        error = I2CM_WriteTxFifo(i2cm, fifo, (MXC_S_I2CM_TRANS_TAG_RXDATA_COUNT | (i-2)));
    }

    if (!error) {
        // Start the transaction if it is not currently ongoing
        if (!(i2cm->trans & MXC_F_I2CM_TRANS_TX_IN_PROGRESS)) {
            i2cm->trans |= MXC_F_I2CM_TRANS_TX_START;
        }

        // NACK the last read byte
        error = I2CM_WriteTxFifo(i2cm, fifo, MXC_S_I2CM_TRANS_TAG_RXDATA_NACK);

        // Send the stop condition
        if (sendStop & !error) {
            error = I2CM_WriteTxFifo(i2cm, fifo, MXC_S_I2CM_TRANS_TAG_STOP);
        }
    }

    // Get the data from the RX FIFO
    int32_t timeout;
    i = 0;
    while ((i < quantity) && (!error)) {
        // Wait for there to be data in the RX FIFO
        timeout = MXC_I2CM_RX_TIMEOUT;
        while (!(i2cm->intfl & MXC_F_I2CM_INTFL_RX_FIFO_NOT_EMPTY) &&
            ((i2cm->bb & MXC_F_I2CM_BB_RX_FIFO_CNT) == 0) &&
            (!error)) {

            if (timeout-- < 0) {
                error = E_TIME_OUT;
            }

            if (i2cm->trans & (MXC_F_I2CM_TRANS_TX_LOST_ARBITR | MXC_F_I2CM_TRANS_TX_NACKED) &&
                (!error)) {
                error = E_COMM_ERR;
            }
        }
        i2cm->intfl = MXC_F_I2CM_INTFL_RX_FIFO_NOT_EMPTY;

        // Save the data from the RX FIFO
        uint16_t tmp;
        tmp = fifo->rx;
        if (tmp & MXC_S_I2CM_RSTLS_TAG_EMPTY) {
            continue;
        }
        rxBuffer[i++] = (uint8_t)tmp;
    }

    // Wait for the transaction to complete
    if (sendStop && !error) {
        error = I2CM_TxInProgress(i2cm);
    }

    // Unlock this I2CM
    mxc_free_lock((uint32_t*)&i2cm_states[i2cm_num].req);

    if (error != E_NO_ERROR) {
        return 0;
    }

    rxBufferLength = quantity;
    return rxBufferLength;
}

uint8_t TwoWire::requestFrom(uint8_t address, uint8_t quantity)
{
    return requestFrom(address, quantity, (uint8_t)true);
}

uint8_t TwoWire::requestFrom(int address, int quantity)
{
    return requestFrom((uint8_t)address, (uint8_t)quantity, (uint8_t)true);
}

uint8_t TwoWire::requestFrom(int address, int quantity, int sendStop)
{
    return requestFrom((uint8_t)address, (uint8_t)quantity, (uint8_t)sendStop);
}

void TwoWire::beginTransmission(uint8_t addr)
{
    // Clear previous Index and Length if any
    txBufferIndex = 0;
    txBufferLength = 0;
    targetAddr = addr;
    transmitting = 1;
}

void TwoWire::beginTransmission(int addr)
{
    beginTransmission((uint8_t)addr);
}

uint8_t TwoWire::endTransmission(uint8_t stop)
{
    int error;

    // Check if data is too long to fit in transmit buffer
    if (getWriteError()) {
        return DATA_TOO_LONG;
    }

    // Disable interrupts and clear interrupt flag from previous transaction
    i2cm->inten = 0;
    i2cm->intfl = i2cm->intfl;

    if (stop) {
        error = I2CM_Write(i2cm, targetAddr, NULL, 0, (txBufferLength) ? txBuffer : NULL, txBufferLength);

        // Keep error unchanged if it is not equal to txBufferLength
        error = (error == txBufferLength) ? E_NO_ERROR : error;

        return map_to_arduino_err(error);
    }

    error = I2CM_Tx(i2cm, fifo, targetAddr, txBuffer, txBufferLength, 0);
    return map_to_arduino_err(error);
}

uint8_t TwoWire::endTransmission(void)
{
    return endTransmission(true);
}

size_t TwoWire::write(uint8_t data)
{
    if (transmitting) {
        if (txBufferLength >= BUFFER_LENGTH) {
            setWriteError();
            return 0;
        }
        txBuffer[txBufferIndex++] = data;
        txBufferLength = txBufferIndex;
    }
    return 1;
}

size_t TwoWire::write(const uint8_t *data, size_t len)
{
    size_t i = 0;

    if (transmitting) {
        for (i = 0; (i < len) && (write(*data++) == 1); i++)
            ;
    }
    return i;
}

int TwoWire::available(void)
{
    return rxBufferLength - rxBufferIndex;
}

int TwoWire::read(void)
{
    if (rxBufferIndex < rxBufferLength) {
        return rxBuffer[rxBufferIndex++];
    } else {
        return -1;
    }
}

int TwoWire::peek(void)
{
    if (rxBufferIndex < rxBufferLength) {
        return rxBuffer[rxBufferIndex];
    } else {
        return -1;
    }
}

void TwoWire::flush(void)
{
    endTransmission(true);
}

inline uint8_t TwoWire::map_to_arduino_err(int err)
{
    switch (err) {
        case E_NO_ERROR:
            return TX_SUCCESS; break;
        case E_NACK_ON_ADDR_ARDUINO:
            return NACK_TX_ADDR; break;
        case E_NACK_ON_DATA_ARDUINO:
            return NACK_TX_DATA; break;
        default:
            return OTHER_ERROR; break;
    }
}

TwoWire Wire0 = TwoWire(0);
#if (MXC_CFG_I2CM_INSTANCES > 1)
TwoWire Wire1 = TwoWire(1);
#endif
#if (MXC_CFG_I2CM_INSTANCES > 2)
TwoWire Wire2 = TwoWire(2);
#endif
#if (MXC_CFG_I2CM_INSTANCES > 3)
TwoWire Wire3 = TwoWire(3);
#endif
#if (MXC_CFG_I2CM_INSTANCES > 4)
TwoWire Wire4 = TwoWire(4);
#endif
#if (MXC_CFG_I2CM_INSTANCES > 5)
TwoWire Wire5 = TwoWire(5);
#endif

TwoWire& Wire = CONCAT(Wire, DEFAULT_I2CM_PORT);
