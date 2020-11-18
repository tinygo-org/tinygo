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
 * $Date: 2016-05-18 16:27:43 -0500 (Wed, 18 May 2016) $
 * $Revision: 22908 $
 *
 ******************************************************************************/

/**
 * @file    spis.s
 * @brief   SPI Slave driver header file.
 */

#include "mxc_config.h"
#include "mxc_sys.h"
#include "spis_regs.h"

#ifndef _SPIS_H_
#define _SPIS_H_

#ifdef __cplusplus
extern "C" {
#endif

/***** Definitions *****/

/// @brief Number of data lines to use.
typedef enum {
    SPIS_WIDTH_1 = 0,
    SPIS_WIDTH_2 = 1,
    SPIS_WIDTH_4 = 2
} spis_width_t;

/// @brief SPIS Transaction request.
typedef struct spis_req spis_req_t;
struct spis_req {
    uint8_t             deass;      ///< End the transaction when SS is deasserted.
    const uint8_t       *tx_data;   ///< TX buffer.
    uint8_t             *rx_data;   ///< RX buffer.
    spis_width_t        width;      ///< Number of data lines to use. 
    unsigned            len;        ///< Number of bytes to send.
    unsigned            read_num;   ///< Number of bytes transacted.
    unsigned            write_num;  ///< Number of bytes transacted.

    /**
     * @brief   Callback for asynchronous request.
     * @param   spis_req_t*     Pointer to the transaction request.
     * @param   int             Error code.
     */
    void (*callback)(spis_req_t*, int);
};

/***** Globals *****/

/***** Function Prototypes *****/

/**
 * @brief   Initialize SPIS module.
 * @param   spis        Pointer to SPIS regs.
 * @param   cfg         Pointer to SPIS configuration.
 * @param   mode        SPI Mode to configure the slave.
 * @param   sys_cfg     Pointer to system configuration object.
 * @returns #E_NO_ERROR if everything is successful.
 */
int SPIS_Init(mxc_spis_regs_t *spis, uint8_t mode, const sys_cfg_spis_t *sys_cfg);


/**
 * @brief   Shutdown SPIS module.
 * @param   spis    Pointer to SPIS regs.
 * @returns #E_NO_ERROR if everything is successful
 */
int SPIS_Shutdown(mxc_spis_regs_t *spis);

/**
 * @brief   Read/write SPIS data. Will block until transaction is complete.
 * @param   spis    Pointer to SPIS regs.
 * @param   req     Request for a SPIS transaction.
 * @note    Callback is ignored.
 * @returns Bytes transacted if everything is successful, error if unsuccessful.
 */
int SPIS_Trans(mxc_spis_regs_t *spis, spis_req_t *req);

/**
 * @brief   Asynchronously read/write SPIS data.
 * @param   spis    Pointer to SPIS regs.
 * @param   req     Request for a SPIS transaction.
 * @note    Request struct must remain allocated until callback.
 * @returns #E_NO_ERROR if everything is successful, error if unsuccessful.
 */
int SPIS_TransAsync(mxc_spis_regs_t *spis, spis_req_t *req);

/**
 * @brief   Abort asynchronous request.
 * @param   req     Pointer to request for a SPIS transaction.
 * @returns #E_NO_ERROR if request aborted, error if unsuccessful.
 */
int SPIS_AbortAsync(spis_req_t *req);

/**
 * @brief   SPIS interrupt handler.
 * @details This function should be called by the application from the interrupt
 *          handler if SPIS interrupts are enabled. Alternately, this function
 *          can be periodically called by the application if SPIS interrupts are
 *          disabled.
 * @param   spis    Base address of the SPIS module.
 */
void SPIS_Handler(mxc_spis_regs_t *spis);

/**
 * @brief   Check the SPIS to see if it's busy.
 * @param   spis    Pointer to SPIS regs.
 * @returns #E_NO_ERROR if idle, #E_BUSY if in use.
 */
int SPIS_Busy(mxc_spis_regs_t *spis);

/**
 * @brief   Attempt to prepare the SPIS for sleep.
 * @param   spis    Pointer to SPIS regs.
* @details  Checks for any ongoing transactions. Disables interrupts if the SPIS
            is idle.
 * @returns #E_NO_ERROR if ready to sleep, #E_BUSY if not ready for sleep.
 */
int SPIS_PrepForSleep(mxc_spis_regs_t *spis);

/**
 * @brief   Enables the SPIS without overwriting existing configuration.
 * @param   spis    Pointer to SPIS regs.
 */
__STATIC_INLINE void SPIS_Enable(mxc_spis_regs_t *spis)
{
    spis->gen_ctrl |= (MXC_F_SPIS_GEN_CTRL_SPI_SLAVE_EN |
                       MXC_F_SPIS_GEN_CTRL_TX_FIFO_EN | MXC_F_SPIS_GEN_CTRL_RX_FIFO_EN);
}

/**
 * @brief   Drain all of the data in the RXFIFO.
 * @param   spis    Pointer to UART regs.
 */
__STATIC_INLINE void SPIS_DrainRX(mxc_spis_regs_t *spis)
{
    uint32_t ctrl_save = spis->gen_ctrl;
    spis->gen_ctrl = (ctrl_save & ~MXC_F_SPIS_GEN_CTRL_RX_FIFO_EN);
    spis->gen_ctrl = ctrl_save;
}

/**
 * @brief   Drain all of the data in the TXFIFO.
 * @param   spis    Pointer to UART regs.
 */
__STATIC_INLINE void SPIS_DrainTX(mxc_spis_regs_t *spis)
{
    uint32_t ctrl_save = spis->gen_ctrl;
    spis->gen_ctrl = (ctrl_save & ~MXC_F_SPIS_GEN_CTRL_TX_FIFO_EN);
    spis->gen_ctrl = ctrl_save;
}

/**
 * @brief   TX FIFO availability.
 * @param   spis    Pointer to UART regs.
 * @returns Number of empty bytes available in write FIFO.
 */
__STATIC_INLINE unsigned SPIS_NumWriteAvail(mxc_spis_regs_t *spis)
{
    return (MXC_CFG_SPIS_FIFO_DEPTH - ((spis->fifo_stat &
        MXC_F_SPIS_FIFO_STAT_TX_FIFO_USED) >> MXC_F_SPIS_FIFO_STAT_TX_FIFO_USED_POS));
}

/**
 * @brief   RX FIFO availability.
 * @param   spis    Pointer to UART regs.
 * @returns Number of bytes in read FIFO.
 */
__STATIC_INLINE unsigned SPIS_NumReadAvail(mxc_spis_regs_t *spis)
{
    return ((spis->fifo_stat & MXC_F_SPIS_FIFO_STAT_RX_FIFO_USED) >>
        MXC_F_SPIS_FIFO_STAT_RX_FIFO_USED_POS);
}

/**
 * @brief   Clear interrupt flags.
 * @param   spis    Pointer to SPIS regs.
 * @param   mask    Mask of interrupts to clear.
 */
__STATIC_INLINE void SPIS_ClearFlags(mxc_spis_regs_t *spis, uint32_t mask)
{
    spis->intfl = mask;
}

/**
 * @brief   Get interrupt flags.
 * @param   spis    Pointer to SPIS regs.
 * @returns Mask of active flags.
 */
__STATIC_INLINE unsigned SPIS_GetFlags(mxc_spis_regs_t *spis)
{
    return (spis->intfl);
}

#ifdef __cplusplus
}
#endif

#endif /* _SPIS_H_ */
