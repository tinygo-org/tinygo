/**
 * @file
 * @brief   Serial Peripheral Interface Master (SPIM) Function Implementations. 
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
 * $Date: 2016-09-08 17:32:38 -0500 (Thu, 08 Sep 2016) $
 * $Revision: 24324 $
 *
 **************************************************************************** */

/* **** Includes **** */
#include <stddef.h>
#include <string.h>
#include "mxc_config.h"
#include "mxc_assert.h"
#include "mxc_sys.h"
#include "mxc_lock.h"
#include "spim.h"

/**
 * @ingroup spim 
 * @{
 */ 
/* **** Definitions **** */
#define SPIM_MAX_BYTE_LEN     32
#define SPIM_MAX_PAGE_LEN     32

/* **** Globals **** */

// Saves the state of the non-blocking requests
typedef struct {
    spim_req_t *req;
    unsigned head_rem;
} spim_req_head_t;

static spim_req_head_t states[MXC_CFG_SPIM_INSTANCES];

/* **** Local Function Prototypes **** */

static unsigned SPIM_ReadRXFIFO(mxc_spim_regs_t *spim, mxc_spim_fifo_regs_t *fifo,
                           uint8_t *data, unsigned len);

static uint32_t SPIM_TransHandler(mxc_spim_regs_t *spim, spim_req_t *req, int spim_num);

/* ************************************************************************* */
int SPIM_Init(mxc_spim_regs_t *spim, const spim_cfg_t *cfg, const sys_cfg_spim_t *sys_cfg)
{
    int err, spim_num;
    uint32_t spim_clk, clocks;

    spim_num = MXC_SPIM_GET_IDX(spim);
    MXC_ASSERT(spim_num >= 0);

    // Check the input parameters
    if(cfg == NULL)
        return E_NULL_PTR;

    if(cfg->baud == 0)
        return E_BAD_PARAM;

    // Set system level configurations
    if ((err = SYS_SPIM_Init(spim, cfg, sys_cfg)) != E_NO_ERROR) {
        return err;
    }

    /*  Configure the baud, make sure the SPIM clk is enabled and the baud
        is less than the maximum */
    spim_clk = SYS_SPIM_GetFreq(spim);
    if((spim_clk == 0) || ((spim_clk == SystemCoreClock) && ((spim_clk/2) < cfg->baud))) {
        return E_BAD_PARAM;
    }

    // Initialize state pointers
    states[spim_num].req = NULL;
    states[spim_num].head_rem = 0;

    // Drain the FIFOs, enable SPIM, enable SCK Feedback
    spim->gen_ctrl = 0;
    spim->gen_ctrl = (MXC_F_SPIM_GEN_CTRL_SPI_MSTR_EN | MXC_F_SPIM_GEN_CTRL_TX_FIFO_EN |
        MXC_F_SPIM_GEN_CTRL_RX_FIFO_EN | MXC_F_SPIM_GEN_CTRL_ENABLE_SCK_FB_MODE);

    // Set mode and page size
    spim->mstr_cfg = (((cfg->mode << MXC_F_SPIM_MSTR_CFG_SPI_MODE_POS) & MXC_F_SPIM_MSTR_CFG_SPI_MODE) |
        MXC_S_SPIM_MSTR_CFG_PAGE_32B | (0x2 << MXC_F_SPIM_MSTR_CFG_ACT_DELAY_POS));

    // Configure the SSEL polarity
    spim->ss_sr_polarity = ((cfg->ssel_pol << MXC_F_SPIM_SS_SR_POLARITY_SS_POLARITY_POS) & 
        MXC_F_SPIM_SS_SR_POLARITY_SS_POLARITY);

#if(MXC_SPIM_REV == 0)
    // Disable the feedback clock in modes 1 and 2
    if((cfg->mode == 1) || (cfg->mode == 2)) {
        spim->gen_ctrl &= ~MXC_F_SPIM_GEN_CTRL_ENABLE_SCK_FB_MODE;  
        spim->mstr_cfg |= (0x1 << MXC_F_SPIM_MSTR_CFG_SDIO_SAMPLE_POINT_POS);    
    }
#else
    // Increase the RX FIFO margin
    MXC_SPIM1->spcl_ctrl = ((MXC_SPIM1->spcl_ctrl & ~(MXC_F_SPIM_SPCL_CTRL_RX_FIFO_MARGIN)) | 
    (0x3 << MXC_F_SPIM_SPCL_CTRL_RX_FIFO_MARGIN_POS));
#endif

    // Calculate the hi/lo clock setting
    if(spim_clk/2 > cfg->baud) {

        /*  Disable the feedback mode and use the sample mode with an appropriate hi/lo clk
            to achieve the lower baud rate */
        spim->gen_ctrl &= ~MXC_F_SPIM_GEN_CTRL_ENABLE_SCK_FB_MODE;

        clocks = (spim_clk / (2*cfg->baud));

        if(clocks == 0 || clocks > 0x10) {
            return E_BAD_PARAM;            
        }

        // 0 => 16 in the 4 bit field for HI_CLK and LO_CLK
        if(clocks == 0x10) {
            clocks = 0;
        }

    } else {
        // Continue to use feedback mode and set hi/lo clk to 1
        clocks = 1;
    }

    spim->mstr_cfg |= (((clocks << MXC_F_SPIM_MSTR_CFG_SCK_HI_CLK_POS) & MXC_F_SPIM_MSTR_CFG_SCK_HI_CLK) |
        ((clocks << MXC_F_SPIM_MSTR_CFG_SCK_LO_CLK_POS) & MXC_F_SPIM_MSTR_CFG_SCK_LO_CLK));

    return E_NO_ERROR;
}

/* ************************************************************************* */
int SPIM_Shutdown(mxc_spim_regs_t *spim)
{
    int spim_num, err;
    spim_req_t *temp_req;

    // Disable and clear interrupts
    spim->inten = 0;
    spim->intfl = spim->intfl;

    // Disable SPIM and FIFOS
    spim->gen_ctrl &= ~(MXC_F_SPIM_GEN_CTRL_SPI_MSTR_EN | MXC_F_SPIM_GEN_CTRL_TX_FIFO_EN |
                        MXC_F_SPIM_GEN_CTRL_RX_FIFO_EN);

    // Call all of the pending callbacks for this SPIM
    spim_num = MXC_SPIM_GET_IDX(spim);
    if(states[spim_num].req != NULL) {

        // Save the request
        temp_req = states[spim_num].req;

        // Unlock this SPIM
        mxc_free_lock((uint32_t*)&states[spim_num].req);

        // Callback if not NULL
        if(temp_req->callback != NULL) {
            temp_req->callback(temp_req, E_SHUTDOWN);
        }
    }

    // Clear system level configurations
    if ((err = SYS_SPIM_Shutdown(spim)) != E_NO_ERROR) {
        return err;
    }

    return E_NO_ERROR;
}

/* ************************************************************************* */
int SPIM_Clocks(mxc_spim_regs_t *spim, uint32_t len, uint8_t ssel, uint8_t deass)
{
    int spim_num;
    mxc_spim_fifo_regs_t *fifo;
    uint16_t header = 0x1;
    uint32_t num = len;

    // Make sure the SPIM has been initialized
    if((spim->gen_ctrl & MXC_F_SPIM_GEN_CTRL_SPI_MSTR_EN) == 0)
        return E_UNINITIALIZED;

    if(!(len > 0)) {
        return E_NO_ERROR;
    }

    // Check the previous transaction if we're switching the slave select
    if((ssel != ((spim->mstr_cfg & MXC_F_SPIM_MSTR_CFG_SLAVE_SEL) >> 
        MXC_F_SPIM_MSTR_CFG_SLAVE_SEL_POS)) && (spim->gen_ctrl & MXC_F_SPIM_GEN_CTRL_BB_SS_IN_OUT)) {

        // Return E_BUSY if the slave select is still asserted
        return E_BUSY;
    }

    // Attempt to lock this SPIM
    spim_num = MXC_SPIM_GET_IDX(spim);
    if(mxc_get_lock((uint32_t*)&states[spim_num].req, 1) != E_NO_ERROR) {
        return E_BUSY;
    }

    // Set which slave select we are using
    spim->mstr_cfg = ((spim->mstr_cfg & ~MXC_F_SPIM_MSTR_CFG_SLAVE_SEL) |
        ((ssel << MXC_F_SPIM_MSTR_CFG_SLAVE_SEL_POS) & MXC_F_SPIM_MSTR_CFG_SLAVE_SEL));

    //force deass to a 1 or 0
    deass = !!deass;

#if(MXC_SPIM_REV == 0)
    // Wait for all of the data to transmit
    while(spim->fifo_ctrl & MXC_F_SPIM_FIFO_CTRL_TX_FIFO_USED) {}

    // Disable the feedback clock, save state
    uint32_t gen_ctrl = spim->gen_ctrl;
    spim->gen_ctrl &= ~MXC_F_SPIM_GEN_CTRL_ENABLE_SCK_FB_MODE;
#endif

    // Get the TX and RX FIFO for this SPIM
    fifo = MXC_SPIM_GET_SPIM_FIFO(spim_num);

    // Send the headers to transmit the clocks
    while(len > 32) {
        fifo->trans_16[0] = header;
        fifo->trans_16[0] = 0xF000;
        fifo->trans_16[0] = 0xF000;
        len -= 32;
    }

    if(len) {
        if(len < 32) {
            header |= (len << 4);
        }
        header |= (deass << 13);

        fifo->trans_16[0] = header;

        if(len > 16) {
            fifo->trans_16[0] = 0xF000;
        }
        fifo->trans_16[0] = 0xF000;
    }

#if(MXC_SPIM_REV == 0)
    // Wait for all of the data to transmit
    while(spim->fifo_ctrl & MXC_F_SPIM_FIFO_CTRL_TX_FIFO_USED) {}

    // Restore feedback clock setting
    spim->gen_ctrl |= (gen_ctrl & MXC_F_SPIM_GEN_CTRL_ENABLE_SCK_FB_MODE);
#endif

    // Unlock this SPIM
    mxc_free_lock((uint32_t*)&states[spim_num].req);

    return num;
}


/* ************************************************************************* */
int SPIM_Trans(mxc_spim_regs_t *spim, spim_req_t *req)
{
    int spim_num;

    // Make sure the SPIM has been initialized
    if((spim->gen_ctrl & MXC_F_SPIM_GEN_CTRL_SPI_MSTR_EN) == 0)
        return E_UNINITIALIZED;

    // Check the input parameters
    if(req == NULL)
        return E_NULL_PTR;

    if((req->rx_data == NULL) && (req->tx_data == NULL))
        return E_NULL_PTR;

    if(!(req->len > 0)) {
        return E_NO_ERROR;
    }

    // Check the previous transaction if we're switching the slave select
    if((req->ssel != ((spim->mstr_cfg & MXC_F_SPIM_MSTR_CFG_SLAVE_SEL) >> 
        MXC_F_SPIM_MSTR_CFG_SLAVE_SEL_POS)) && (spim->gen_ctrl & MXC_F_SPIM_GEN_CTRL_BB_SS_IN_OUT)) {

        // Return E_BUSY if the slave select is still asserted
        return E_BUSY;
    }

    // Attempt to register this write request
    spim_num = MXC_SPIM_GET_IDX(spim);
    if(mxc_get_lock((uint32_t*)&states[spim_num].req, (uint32_t)req) != E_NO_ERROR) {
        return E_BUSY;
    }

    // Set which slave select we are using
    spim->mstr_cfg = ((spim->mstr_cfg & ~MXC_F_SPIM_MSTR_CFG_SLAVE_SEL) |
        ((req->ssel << MXC_F_SPIM_MSTR_CFG_SLAVE_SEL_POS) & MXC_F_SPIM_MSTR_CFG_SLAVE_SEL));

    //force deass to a 1 or 0
    req->deass = !!req->deass;

    // Clear the number of bytes counter
    req->read_num = 0;
    req->write_num = 0;
    req->callback = NULL;
    states[spim_num].head_rem = 0;

    // Start the transaction, keep calling the handler until complete
    while(SPIM_TransHandler(spim, req, spim_num) != 0);

    if(req->tx_data == NULL) {
        return req->read_num;
    }
    return req->write_num;
}

/* ************************************************************************* */
int SPIM_TransAsync(mxc_spim_regs_t *spim, spim_req_t *req)
{
    int spim_num;

    // Make sure the SPIM has been initialized
    if((spim->gen_ctrl & MXC_F_SPIM_GEN_CTRL_SPI_MSTR_EN) == 0)
        return E_UNINITIALIZED;

    // Check the input parameters
    if(req == NULL)
        return E_NULL_PTR;

    if((req->rx_data == NULL) && (req->tx_data == NULL))
        return E_NULL_PTR;

    if(!(req->len > 0)) {
        return E_NO_ERROR;
    }


    // Check the previous transaction if we're switching the slave select
    if((req->ssel != ((spim->mstr_cfg & MXC_F_SPIM_MSTR_CFG_SLAVE_SEL) >> 
        MXC_F_SPIM_MSTR_CFG_SLAVE_SEL_POS)) && (spim->gen_ctrl & MXC_F_SPIM_GEN_CTRL_BB_SS_IN_OUT)) {

        // Return E_BUSY if the slave select is still asserted
        return E_BUSY;
    }

    // Attempt to register this write request
    spim_num = MXC_SPIM_GET_IDX(spim);
    if(mxc_get_lock((uint32_t*)&states[spim_num].req, (uint32_t)req) != E_NO_ERROR) {
        return E_BUSY;
    }

    // Set which slave select we are using
    spim->mstr_cfg = ((spim->mstr_cfg & ~MXC_F_SPIM_MSTR_CFG_SLAVE_SEL) |
        ((req->ssel << MXC_F_SPIM_MSTR_CFG_SLAVE_SEL_POS) & MXC_F_SPIM_MSTR_CFG_SLAVE_SEL));

    //force deass to a 1 or 0
    req->deass = !!req->deass;

    // Clear the number of bytes counter
    req->read_num = 0;
    req->write_num = 0;
    states[spim_num].head_rem = 0;

    // Start the transaction, enable the interrupts
    spim->inten = SPIM_TransHandler(spim, req, spim_num);

    return E_NO_ERROR;
}

/* ************************************************************************* */
int SPIM_AbortAsync(spim_req_t *req)
{
    int spim_num;
    mxc_spim_regs_t *spim;

    // Check the input parameters
    if(req == NULL) {
        return E_BAD_PARAM;
    }

    // Find the request, set to NULL
    for(spim_num = 0; spim_num < MXC_CFG_SPIM_INSTANCES; spim_num++) {
        if(req == states[spim_num].req) {

            spim = MXC_SPIM_GET_SPIM(spim_num);

            // Disable interrupts, clear the flags
            spim->inten = 0;
            spim->intfl = spim->intfl;

            // Reset the SPIM to cancel the on ongoing transaction
            spim->gen_ctrl &= ~(MXC_F_SPIM_GEN_CTRL_SPI_MSTR_EN);
            spim->gen_ctrl |= (MXC_F_SPIM_GEN_CTRL_SPI_MSTR_EN);

            // Unlock this SPIM
            mxc_free_lock((uint32_t*)&states[spim_num].req);

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
void SPIM_Handler(mxc_spim_regs_t *spim)
{
    int spim_num;
    uint32_t flags;

    // Clear the interrupt flags
    spim->inten = 0;
    flags = spim->intfl;
    spim->intfl = flags;

    spim_num = MXC_SPIM_GET_IDX(spim);

    // Figure out if this SPIM has an active request
    if((states[spim_num].req != NULL) && (flags)) {
        spim->inten = SPIM_TransHandler(spim, states[spim_num].req, spim_num);
    }
}

/* ************************************************************************* */
int SPIM_Busy(mxc_spim_regs_t *spim)
{
    // Check to see if there are any ongoing transactions
    if((states[MXC_SPIM_GET_IDX(spim)].req == NULL) &&
        !(spim->fifo_ctrl & MXC_F_SPIM_FIFO_CTRL_TX_FIFO_USED)) {

        return E_NO_ERROR;
    }

    return E_BUSY;
}

/* ************************************************************************* */
int SPIM_PrepForSleep(mxc_spim_regs_t *spim)
{
    if(SPIM_Busy(spim) != E_NO_ERROR) {
        return E_BUSY;
    }

    // Disable interrupts
    spim->inten = 0;
    return E_NO_ERROR;
}

/* ************************************************************************* */
static unsigned SPIM_ReadRXFIFO(mxc_spim_regs_t *spim, mxc_spim_fifo_regs_t *fifo,
                           uint8_t *data, unsigned len)
{
    unsigned num = 0;
    unsigned avail = ((spim->fifo_ctrl & MXC_F_SPIM_FIFO_CTRL_RX_FIFO_USED) >> 
        MXC_F_SPIM_FIFO_CTRL_RX_FIFO_USED_POS);

    // Get data from the RXFIFO
    while(avail && (len - num)) {

        if((avail >= 4) && ((len-num) >= 4)) {
            // Save data from the RXFIFO
            uint32_t temp = fifo->rslts_32[0];
            data[num+0] = ((temp & 0x000000FF) >> 0);
            data[num+1] = ((temp & 0x0000FF00) >> 8);
            data[num+2] = ((temp & 0x00FF0000) >> 16);
            data[num+3] = ((temp & 0xFF000000) >> 24);
            num+=4;
            avail-=4;
        } else if ((avail >= 2) && ((len-num) >= 2)) {
            // Save data from the RXFIFO
            uint16_t temp = fifo->rslts_16[0];
            data[num+0] = ((temp & 0x00FF) >> 0);
            data[num+1] = ((temp & 0xFF00) >> 8);
            num+=2;
            avail-=2;
        } else {
            // Save data from the RXFIFO
            data[num] = fifo->rslts_8[0];
            num+=1;
            avail-=1;
        }

        // Check to see if there is more data in the FIFO
        if(avail == 0) {
            avail = ((spim->fifo_ctrl & MXC_F_SPIM_FIFO_CTRL_RX_FIFO_USED) >> 
                MXC_F_SPIM_FIFO_CTRL_RX_FIFO_USED_POS);
        }
    }

    return num;
}

uint16_t header_save;


/* ************************************************************************* */
static uint32_t SPIM_TransHandler(mxc_spim_regs_t *spim, spim_req_t *req, int spim_num)
{
    uint8_t read, write;
    uint16_t header;
    uint32_t pages, bytes, inten;
	unsigned remain, bytes_read, head_rem_temp, avail;
    mxc_spim_fifo_regs_t *fifo;

    inten = 0;

    // Get the FIFOS for this UART
    fifo = MXC_SPIM_GET_SPIM_FIFO(spim_num);

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

    // Read byte from the FIFO if we are reading
    if(read) {

        // Read all of the data in the RXFIFO, or until we don't need anymore
        bytes_read = SPIM_ReadRXFIFO(spim, fifo, &req->rx_data[req->read_num],
                                         (req->len - req->read_num));

        req->read_num += bytes_read;

        // Adjust head_rem if we are only reading
        if(!write && (states[spim_num].head_rem > 0)) {
            states[spim_num].head_rem -= bytes_read;
        }

        // Figure out how many byte we have left to read
        if(states[spim_num].head_rem > 0) {
            remain = states[spim_num].head_rem;
        } else {
            remain = req->len - req->read_num;
        }

        if(remain) {

            // Set the RX interrupts
            if (remain > MXC_CFG_SPIM_FIFO_DEPTH) {
                spim->fifo_ctrl = ((spim->fifo_ctrl & ~MXC_F_SPIM_FIFO_CTRL_RX_FIFO_AF_LVL) |
                                   ((MXC_CFG_SPIM_FIFO_DEPTH - 2) <<
                                    MXC_F_SPIM_FIFO_CTRL_RX_FIFO_AF_LVL_POS));

            } else {
                spim->fifo_ctrl = ((spim->fifo_ctrl & ~MXC_F_SPIM_FIFO_CTRL_RX_FIFO_AF_LVL) |
                                   ((remain - 1) << MXC_F_SPIM_FIFO_CTRL_RX_FIFO_AF_LVL_POS));
            }

            inten |= MXC_F_SPIM_INTEN_RX_FIFO_AF;
        }
    }

    // Figure out how many bytes we have left to send headers for
    if(write) {
        remain = req->len - req->write_num;
    } else {
        remain = req->len - req->read_num;
    }

    // See if we need to send a new header
    if(states[spim_num].head_rem <= 0 && remain) {

        // Set the transaction configuration in the header
        header = ((write << 0) | (read << 1) | (req->width << 9));
        
        if(remain >= SPIM_MAX_BYTE_LEN) {

            // Send a 32 byte header
            if(remain == SPIM_MAX_BYTE_LEN) {

                header |= ((0x1 << 2) | (req->deass << 13));

                // Save the number of bytes we need to write to the FIFO
                bytes = SPIM_MAX_BYTE_LEN;

            } else {
                // Send in increments of 32 byte pages
                header |= (0x2 << 2);
                pages = remain / SPIM_MAX_PAGE_LEN;
                
                if(pages >= 32) {
                    // 0 maps to 32 in the header
                    bytes = 32 * SPIM_MAX_PAGE_LEN;
                } else {
                    header |= (pages << 4);
                    bytes = pages * SPIM_MAX_PAGE_LEN;
                }

                // Check if this is the last header we will send
                if((remain - bytes) == 0) {
                    header |= (req->deass << 13);
                }
            }   

            header_save = header;
            fifo->trans_16[0] = header;

            // Save the number of bytes we need to write to the FIFO
            states[spim_num].head_rem = bytes;

            } else {

                // Send final header with the number of bytes remaining and if
                // we want to de-assert the SS at the end of the transaction
                header |= ((0x1 << 2) | (remain << 4) | (req->deass << 13));
                fifo->trans_16[0] = header;
                states[spim_num].head_rem = remain;
            }
    }

    // Put data into the FIFO if we are writing
    remain = req->len - req->write_num;
    head_rem_temp = states[spim_num].head_rem;
    if(write && head_rem_temp) {

        // Fill the FIFO
        avail = (MXC_CFG_SPIM_FIFO_DEPTH - ((spim->fifo_ctrl & MXC_F_SPIM_FIFO_CTRL_TX_FIFO_USED) >> 
            MXC_F_SPIM_FIFO_CTRL_TX_FIFO_USED_POS));

        // Use memcpy for everything except the last byte in odd length transactions
        while((avail >= 2) && (head_rem_temp >= 2)) {

            unsigned length;
            if(head_rem_temp < avail) {
                length = head_rem_temp;
            } else {
                length = avail;
            }

            // Only memcpy even numbers
            length = ((length / 2) * 2); 

            memcpy((void*)fifo->trans_32, &(req->tx_data[req->write_num]), length);

            head_rem_temp -= length;
            req->write_num += length;

            avail = (MXC_CFG_SPIM_FIFO_DEPTH - ((spim->fifo_ctrl & MXC_F_SPIM_FIFO_CTRL_TX_FIFO_USED) >> 
                MXC_F_SPIM_FIFO_CTRL_TX_FIFO_USED_POS));
        }

        // Copy the last byte and pad with 0xF0 to not get confused as header
        if((avail >= 1) && (head_rem_temp == 1)) {

            // Write the last byte
            fifo->trans_16[0] = (0xF000 | req->tx_data[req->write_num]);

            avail -= 1;
            req->write_num += 1;
            head_rem_temp -= 1;
        }

        states[spim_num].head_rem = head_rem_temp;
        remain = req->len - req->write_num;

        // Set the TX interrupts
        if(remain) {

            // Set the TX FIFO almost empty interrupt if we have to refill
            spim->fifo_ctrl = ((spim->fifo_ctrl & ~MXC_F_SPIM_FIFO_CTRL_TX_FIFO_AE_LVL) |
                ((MXC_CFG_SPIM_FIFO_DEPTH - 2) << MXC_F_SPIM_FIFO_CTRL_TX_FIFO_AE_LVL_POS));

            inten |= MXC_F_SPIM_INTEN_TX_FIFO_AE;

        }
    }

    // Check to see if we've finished reading and writing
    if(((read && (req->read_num == req->len)) || !read) &&
            ((req->write_num == req->len) || !write)) {

        // Disable interrupts
        spim->inten = 0;

        // Unlock this SPIM
        mxc_free_lock((uint32_t*)&states[spim_num].req);

        // Callback if not NULL
        if(req->callback != NULL) {
            req->callback(req, E_NO_ERROR);
        }
    }

    // Enable the SPIM interrupts
    return inten;
}
/**@} end of ingroup spim */
