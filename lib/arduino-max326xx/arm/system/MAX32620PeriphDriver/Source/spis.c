/**
 * @file
 * @brief   SPI Slave API Function Implementations.
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
 * $Date: 2016-09-08 17:42:39 -0500 (Thu, 08 Sep 2016) $
 * $Revision: 24325 $
 **************************************************************************** */

/* **** Includes **** */
#include <stddef.h>
#include <string.h>
#include "mxc_config.h"
#include "mxc_assert.h"
#include "mxc_lock.h"
#include "spis.h"

/**
 * @ingroup spis 
 * @{
 */ 
/* **** Definitions **** */
#define SPIS_FIFO_BUFFER        6

/* **** Globals **** */

static spis_req_t *states[MXC_CFG_SPIS_INSTANCES];

/* **** Functions **** */

static uint32_t SPIS_TransHandler(mxc_spis_regs_t *spis, spis_req_t *req, int spis_num);

/* ************************************************************************* */
int SPIS_Init(mxc_spis_regs_t *spis, uint8_t mode, const sys_cfg_spis_t *sys_cfg)
{
    int err, spis_num;

    spis_num = MXC_SPIS_GET_IDX(spis);
    MXC_ASSERT(spis_num >= 0);

    // Set system level configurations
    if ((err = SYS_SPIS_Init(sys_cfg)) != E_NO_ERROR) {
        return err;
    }

    // Initialize state pointers
    states[spis_num] = NULL;

    // Drain the FIFOs, enable SPIS
    spis->gen_ctrl = 0;
    spis->gen_ctrl = (MXC_F_SPIS_GEN_CTRL_SPI_SLAVE_EN | MXC_F_SPIS_GEN_CTRL_TX_FIFO_EN |
        MXC_F_SPIS_GEN_CTRL_RX_FIFO_EN);

    // Set the TX FIFO almost empty level
    spis->fifo_ctrl = ((spis->fifo_ctrl & ~MXC_F_SPIS_FIFO_CTRL_TX_FIFO_AE_LVL) |
        ((MXC_CFG_SPIS_FIFO_DEPTH - SPIS_FIFO_BUFFER) << MXC_F_SPIS_FIFO_CTRL_TX_FIFO_AE_LVL_POS));

    return E_NO_ERROR;
}

/* ************************************************************************* */
int SPIS_Shutdown(mxc_spis_regs_t *spis)
{
    int spis_num, err;
    spis_req_t *temp_req;

    // Disable and clear interrupts
    spis->inten = 0;
    spis->intfl = spis->intfl;

    // Disable SPIS and FIFOS
    spis->gen_ctrl &= ~(MXC_F_SPIS_GEN_CTRL_SPI_SLAVE_EN | MXC_F_SPIS_GEN_CTRL_TX_FIFO_EN |
        MXC_F_SPIS_GEN_CTRL_RX_FIFO_EN);

    // Call all of the pending callbacks for this SPIS
    spis_num = MXC_SPIS_GET_IDX(spis);
    if(states[spis_num] != NULL) {

        // Save the request
        temp_req = states[spis_num];

        // Unlock this SPIS
        mxc_free_lock((uint32_t*)&states[spis_num]);

        // Callback if not NULL
        if(temp_req->callback != NULL) {
            temp_req->callback(temp_req, E_SHUTDOWN);
        }
    }

    // Clear system level configurations
    if ((err = SYS_SPIS_Shutdown(spis)) != E_NO_ERROR) {
        return err;
    }

    return E_NO_ERROR;
}

/* ************************************************************************* */
int SPIS_Trans(mxc_spis_regs_t *spis, spis_req_t *req)
{
    int spis_num;

    // Make sure the SPIS has been initialized
    if((spis->gen_ctrl & MXC_F_SPIS_GEN_CTRL_SPI_SLAVE_EN) == 0)
        return E_UNINITIALIZED;

    // Check the input parameters
    if(req == NULL)
        return E_NULL_PTR;

    if((req->rx_data == NULL) && (req->tx_data == NULL))
        return E_NULL_PTR;

    if(!(req->len > 0)) {
        return E_NO_ERROR;
    }

    // Attempt to register this write request
    spis_num = MXC_SPIS_GET_IDX(spis);
    if(mxc_get_lock((uint32_t*)&states[spis_num], (uint32_t)req) != E_NO_ERROR) {
        return E_BUSY;
    }

    //force deass to a 1 or 0
    req->deass = !!req->deass;

    // Clear the number of bytes counter
    req->read_num = 0;
    req->write_num = 0;
    req->callback = NULL;

    // Start the transaction, keep calling the handler until complete
    spis->intfl = (MXC_F_SPIS_INTFL_SS_DEASSERTED | MXC_F_SPIS_INTFL_TX_UNDERFLOW);
    while(SPIS_TransHandler(spis, req, spis_num) & (MXC_F_SPIS_INTEN_RX_FIFO_AF | 
        MXC_F_SPIS_INTEN_TX_FIFO_AE)) {

        if((req->tx_data != NULL) && (spis->intfl & MXC_F_SPIS_INTFL_TX_UNDERFLOW)) {
            return E_UNDERFLOW;
        }

        if((req->rx_data != NULL) && (spis->intfl & MXC_F_SPIS_INTFL_RX_LOST_DATA)) {
            return E_OVERFLOW;
        }

        if((req->deass) && (spis->intfl & MXC_F_SPIS_INTFL_SS_DEASSERTED)) {
            if(((req->rx_data != NULL) && ((req->read_num + SPIS_NumReadAvail(spis)) < req->len)) || 
                ((req->tx_data != NULL) && (req->write_num < req->len))) {

                return E_COMM_ERR;
            }
        }
    }

    if(req->tx_data == NULL) {
        return req->read_num;
    }
    return req->write_num;
}

/* ************************************************************************* */
int SPIS_TransAsync(mxc_spis_regs_t *spis, spis_req_t *req)
{
    int spis_num;

    // Make sure the SPIS has been initialized
    if((spis->gen_ctrl & MXC_F_SPIS_GEN_CTRL_SPI_SLAVE_EN) == 0)
        return E_UNINITIALIZED;

    // Check the input parameters
    if(req == NULL)
        return E_NULL_PTR;

    if((req->rx_data == NULL) && (req->tx_data == NULL))
        return E_NULL_PTR;

    if(!(req->len > 0)) {
        return E_NO_ERROR;
    }

    // Attempt to register this write request
    spis_num = MXC_SPIS_GET_IDX(spis);
    if(mxc_get_lock((uint32_t*)&states[spis_num], (uint32_t)req) != E_NO_ERROR) {
        return E_BUSY;
    }

    //force deass to a 1 or 0
    req->deass = !!req->deass;

    // Clear the number of bytes counter
    req->read_num = 0;
    req->write_num = 0;

    // Start the transaction, enable the interrupts
    spis->intfl = MXC_F_SPIS_INTFL_SS_DEASSERTED;
    spis->inten = SPIS_TransHandler(spis, req, spis_num);

    if(spis->intfl & MXC_F_SPIS_INTFL_SS_DEASSERTED) {
        return E_COMM_ERR;
    }

    return E_NO_ERROR;
}

/* ************************************************************************* */
int SPIS_AbortAsync(spis_req_t *req)
{
    int spis_num;

    // Check the input parameters
    if(req == NULL) {
        return E_BAD_PARAM;
    }

    // Find the request, set to NULL
    for(spis_num = 0; spis_num < MXC_CFG_SPIS_INSTANCES; spis_num++) {
        if(req == states[spis_num]) {

            // Disable interrupts, clear the flags
            MXC_SPIS_GET_SPIS(spis_num)->inten = 0;
            MXC_SPIS_GET_SPIS(spis_num)->intfl = MXC_SPIS_GET_SPIS(spis_num)->intfl;

            // Unlock this SPIS
            mxc_free_lock((uint32_t*)&states[spis_num]);

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
void SPIS_Handler(mxc_spis_regs_t *spis)
{
    int spis_num;
    uint32_t flags;
    spis_req_t *req;

    // Clear the interrupt flags
    spis->inten = 0;
    flags = spis->intfl;
    spis->intfl = flags;

    spis_num = MXC_SPIS_GET_IDX(spis);
    req = states[spis_num];

    // Check for errors
    if((flags & MXC_F_SPIS_INTFL_TX_UNDERFLOW) && (req->tx_data != NULL)) {
        // Unlock this SPIS
        mxc_free_lock((uint32_t*)&states[spis_num]);

        // Callback if not NULL
        if(req->callback != NULL) {
            req->callback(req, E_UNDERFLOW);
        }
        return;
    }
    if((flags & MXC_F_SPIS_INTFL_RX_LOST_DATA) && (req->rx_data != NULL)) {
        // Unlock this SPIS
        mxc_free_lock((uint32_t*)&states[spis_num]);

        // Callback if not NULL
        if(req->callback != NULL) {
            req->callback(req, E_OVERFLOW);
        }
        return;
    }

    // Check for deassert
    if((flags & MXC_F_SPIS_INTFL_SS_DEASSERTED) && (req != NULL) &&
        (req->deass)) {

        if(((req->rx_data != NULL) && ((req->read_num + SPIS_NumReadAvail(spis)) < req->len)) || 
            ((req->tx_data != NULL) && (req->write_num < req->len))) {

            // Unlock this SPIS
            mxc_free_lock((uint32_t*)&states[spis_num]);

            // Callback if not NULL
            if(req->callback != NULL) {
                req->callback(states[spis_num], E_COMM_ERR);
            }

            return;
        }
    }

    // Figure out if this SPIS has an active request
    if(flags && (req != NULL)) {

        spis->inten = SPIS_TransHandler(spis, req, spis_num);
    }
}

/* ************************************************************************* */
int SPIS_Busy(mxc_spis_regs_t *spis)
{
    // Check to see if there are any ongoing transactions
    if(states[MXC_SPIS_GET_IDX(spis)] == NULL) {
        return E_NO_ERROR;
    }

    return E_BUSY;
}

/* ************************************************************************* */
int SPIS_PrepForSleep(mxc_spis_regs_t *spis)
{
    if(SPIS_Busy(spis) != E_NO_ERROR) {
        return E_BUSY;
    }

    // Disable interrupts
    spis->inten = 0;
    return E_NO_ERROR;
}

/* ************************************************************************* */
static uint32_t SPIS_TransHandler(mxc_spis_regs_t *spis, spis_req_t *req, int spis_num)
{
    uint8_t read, write;
    uint32_t inten;
	unsigned remain, avail, temp_len;
    mxc_spis_fifo_regs_t *fifo;

    inten = 0;

    // Get the FIFOS for this UART
    fifo = MXC_SPIS_GET_SPIS_FIFO(spis_num);

    // Figure out if we're reading
    if(req->rx_data != NULL) {
        read = 1;
    } else {
        read = 0;
    }

    // Figure out if we're writing
    if(req->tx_data != NULL) {
        write = 1;
    } else {
        write = 0;
    }

    // Put data into the FIFO if we are writing
    if(write) {

        avail = SPIS_NumWriteAvail(spis);
        remain = req->len - req->write_num;

        if(remain > avail) {
            temp_len = avail;
        } else {
            temp_len = remain;
        }

        memcpy((void*)fifo->tx_32, &(req->tx_data[req->write_num]), temp_len);
        spis->intfl = MXC_F_SPIS_INTFL_TX_FIFO_AE;
        req->write_num += temp_len;
        remain = req->len - req->write_num;

        // Set the TX interrupts
        if(remain) {
            inten |= (MXC_F_SPIS_INTEN_TX_FIFO_AE | MXC_F_SPIS_INTFL_TX_UNDERFLOW);
        }
    }

    // Read from the FIFO if we are reading
    if(read) {

        avail = SPIS_NumReadAvail(MXC_SPIS);
        remain = req->len - req->read_num;

        if(remain > avail) {
            temp_len = avail;
        } else {
            temp_len = remain;
        }

        memcpy((void*)&req->rx_data[req->read_num], (void*)&(fifo->rx_8[0]), temp_len);
        spis->intfl = MXC_F_SPIS_INTFL_RX_FIFO_AF;
        req->read_num += temp_len;
        remain = req->len - req->read_num;

        // Set the RX interrupts
        if(remain) {

            // Adjust the almost full threshold
            if (remain > (MXC_CFG_SPIS_FIFO_DEPTH - SPIS_FIFO_BUFFER)) {
                spis->fifo_ctrl = ((spis->fifo_ctrl & ~MXC_F_SPIS_FIFO_CTRL_RX_FIFO_AF_LVL) |
                    ((MXC_CFG_SPIS_FIFO_DEPTH - SPIS_FIFO_BUFFER) << MXC_F_SPIS_FIFO_CTRL_RX_FIFO_AF_LVL_POS));

            } else {
                spis->fifo_ctrl = ((spis->fifo_ctrl & ~MXC_F_SPIS_FIFO_CTRL_RX_FIFO_AF_LVL) |
                    ((remain-1) << MXC_F_SPIS_FIFO_CTRL_RX_FIFO_AF_LVL_POS));
            }

            inten |= (MXC_F_SPIS_INTEN_RX_FIFO_AF | MXC_F_SPIS_INTFL_RX_LOST_DATA);
        }
    }

    // Check to see if we've finished reading and writing
    if(((read && (req->read_num == req->len)) || !read) &&
        ((req->write_num == req->len) || !write)) {

        // Unlock this SPIS
        mxc_free_lock((uint32_t*)&states[spis_num]);

        // Callback if not NULL
        if(req->callback != NULL) {
            req->callback(req, E_NO_ERROR);
        }

        return 0;
    }

    // Enable deassert interrupt
    if(req->deass) {
        inten |= MXC_F_SPIS_INTEN_SS_DEASSERTED;
    }

    return inten;
}
/**@} end of ingroup spis*/
