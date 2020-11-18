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
 * @file    spim.h
 * @brief   SPI Master driver header file.
 */

#include "mxc_config.h"
#include "mxc_sys.h"
#include "spim_regs.h"

#ifndef _SPIM_H_
#define _SPIM_H_

#ifdef __cplusplus
extern "C" {
#endif

/***** Definitions *****/

/// @brief Active levels for slave select lines.
typedef enum {
    SPIM_SSEL0_HIGH  =    (0x1 << 0),
    SPIM_SSEL0_LOW   =    0,
    SPIM_SSEL1_HIGH  =    (0x1 << 1),
    SPIM_SSEL1_LOW   =    0,
    SPIM_SSEL2_HIGH  =    (0x1 << 2),
    SPIM_SSEL2_LOW   =    0,
    SPIM_SSEL3_HIGH  =    (0x1 << 3),
    SPIM_SSEL3_LOW   =    0,
    SPIM_SSEL4_HIGH  =    (0x1 << 4),
    SPIM_SSEL4_LOW   =    0
}
spim_ssel_t;

/// @brief Number of data lines to use.
typedef enum {
    SPIM_WIDTH_1 = 0,
    SPIM_WIDTH_2 = 1,
    SPIM_WIDTH_4 = 2
} spim_width_t;

/// @brief SPIM configuration type.
typedef struct {
    uint8_t     mode;       ///< SPIM mode to use, 0-3.
    uint32_t    ssel_pol;   ///< Mask of active levels for slave select signals, use spim_ssel_t.
    uint32_t    baud;       ///< Baud rate in Hz.
} spim_cfg_t;

/// @brief SPIM Transaction request.
typedef struct spim_req spim_req_t;
struct spim_req {
    uint8_t             ssel;       ///< Slave select number.
    uint8_t             deass;      ///< De-assert slave select at the end of the transaction.
    const uint8_t       *tx_data;   ///< TX buffer.
    uint8_t             *rx_data;   ///< RX buffer.
    spim_width_t        width;      ///< Number of data lines to use
    unsigned            len;        ///< Number of bytes to send.
    unsigned            read_num;   ///< Number of bytes read.
    unsigned            write_num;  ///< Number of bytes written.

    /**
     * @brief   Callback for asynchronous request.
     * @param   spim_req_t*  Pointer to the transaction request.
     * @param   int         Error code.
     */
    void (*callback)(spim_req_t*, int);
};

/***** Globals *****/

/***** Function Prototypes *****/

/**
 * @brief   Initialize SPIM module.
 * @param   spim        Pointer to SPIM regs.
 * @param   cfg         Pointer to SPIM configuration.
 * @param   sys_cfg     Pointer to system configuration object
 * @returns #E_NO_ERROR if everything is successful
 */
int SPIM_Init(mxc_spim_regs_t *spim, const spim_cfg_t *cfg, const sys_cfg_spim_t *sys_cfg);

/**
 * @brief   Shutdown SPIM module.
 * @param   spim    Pointer to SPIM regs.
 * @returns #E_NO_ERROR if everything is successful
 */
int SPIM_Shutdown(mxc_spim_regs_t *spim);

/**
 * @brief   Send Clock cycles on SCK without reading or writing.
 * @param   spim    Pointer to SPIM regs.
 * @param   len     Number of clock cycles to send.
 * @param   ssel    Slave select number.
 * @param   deass   De-assert slave select at the end of the transaction.
 * @returns Cycles transacted if everything is successful, error if unsuccessful.
 */
int SPIM_Clocks(mxc_spim_regs_t *spim, uint32_t len, uint8_t ssel, uint8_t deass);

/**
 * @brief   Read/write SPIM data. Will block until transaction is complete.
 * @param   spim    Pointer to SPIM regs.
 * @param   req     Request for a SPIM transaction.
 * @note    Callback is ignored.
 * @returns Bytes transacted if everything is successful, error if unsuccessful.
 */
int SPIM_Trans(mxc_spim_regs_t *spim, spim_req_t *req);

/**
 * @brief   Asynchronously read/write SPIM data.
 * @param   spim    Pointer to SPIM regs.
 * @param   req     Request for a SPIM transaction.
 * @note    Request struct must remain allocated until callback.
 * @returns #E_NO_ERROR if everything is successful, error if unsuccessful.
 */
int SPIM_TransAsync(mxc_spim_regs_t *spim, spim_req_t *req);

/**
 * @brief   Abort asynchronous request.
 * @param   req     Pointer to request for a SPIM transaction.
 * @returns #E_NO_ERROR if request aborted, error if unsuccessful.
 */
int SPIM_AbortAsync(spim_req_t *req);

/**
 * @brief   SPIM interrupt handler.
 * @details This function should be called by the application from the interrupt
 *          handler if SPIM interrupts are enabled. Alternately, this function
 *          can be periodically called by the application if SPIM interrupts are
 *          disabled.
 * @param   spim    Base address of the SPIM module.
 */
void SPIM_Handler(mxc_spim_regs_t *spim);

/**
 * @brief   Check the SPIM to see if it's busy.
 * @param   spim    Pointer to SPIM regs.
 * @returns #E_NO_ERROR if idle, #E_BUSY if in use.
 */
int SPIM_Busy(mxc_spim_regs_t *spim);

/**
 * @brief   Attempt to prepare the SPIM for sleep.
 * @param   spim    Pointer to SPIM regs.
* @details  Checks for any ongoing transactions. Disables interrupts if the SPIM
            is idle.
 * @returns #E_NO_ERROR if ready to sleep, #E_BUSY if not ready for sleep.
 */
int SPIM_PrepForSleep(mxc_spim_regs_t *spim);

/**
 * @brief   Enables the SPIM without overwriting existing configuration.
 * @param   spim    Pointer to SPIM regs.
 */
__STATIC_INLINE void SPIM_Enable(mxc_spim_regs_t *spim)
{
    spim->gen_ctrl |= (MXC_F_SPIM_GEN_CTRL_SPI_MSTR_EN |
                       MXC_F_SPIM_GEN_CTRL_TX_FIFO_EN | MXC_F_SPIM_GEN_CTRL_RX_FIFO_EN);
}

/**
 * @brief   Drain all of the data in the RXFIFO.
 * @param   spim    Pointer to UART regs.
 */
__STATIC_INLINE void SPIM_DrainRX(mxc_spim_regs_t *spim)
{
    uint32_t ctrl_save = spim->gen_ctrl;
    spim->gen_ctrl = (ctrl_save & ~MXC_F_SPIM_GEN_CTRL_RX_FIFO_EN);
    spim->gen_ctrl = ctrl_save;
}

/**
 * @brief   Drain all of the data in the TXFIFO.
 * @param   spim    Pointer to UART regs.
 */
__STATIC_INLINE void SPIM_DrainTX(mxc_spim_regs_t *spim)
{
    uint32_t ctrl_save = spim->gen_ctrl;
    spim->gen_ctrl = (ctrl_save & ~MXC_F_SPIM_GEN_CTRL_TX_FIFO_EN);
    spim->gen_ctrl = ctrl_save;
}

/**
 * @brief   TX FIFO availability.
 * @param   spim    Pointer to UART regs.
 * @returns Number of empty bytes available in write FIFO.
 */
__STATIC_INLINE unsigned SPIM_NumWriteAvail(mxc_spim_regs_t *spim)
{
    return (MXC_CFG_SPIM_FIFO_DEPTH - ((spim->fifo_ctrl &
                                        MXC_F_SPIM_FIFO_CTRL_TX_FIFO_USED) >> MXC_F_SPIM_FIFO_CTRL_TX_FIFO_USED_POS));
}

/**
 * @brief   RX FIFO availability.
 * @param   spim    Pointer to UART regs.
 * @returns Number of bytes in read FIFO.
 */
__STATIC_INLINE unsigned SPIM_NumReadAvail(mxc_spim_regs_t *spim)
{
    return ((spim->fifo_ctrl & MXC_F_SPIM_FIFO_CTRL_RX_FIFO_USED) >>
            MXC_F_SPIM_FIFO_CTRL_RX_FIFO_USED_POS);
}

/**
 * @brief   Clear interrupt flags.
 * @param   spim    Pointer to SPIM regs.
 * @param   mask    Mask of interrupts to clear.
 */
__STATIC_INLINE void SPIM_ClearFlags(mxc_spim_regs_t *spim, uint32_t mask)
{
    spim->intfl = mask;
}

/**
 * @brief   Get interrupt flags.
 * @param   spim    Pointer to SPIM regs.
 * @returns Mask of active flags.
 */
__STATIC_INLINE unsigned SPIM_GetFlags(mxc_spim_regs_t *spim)
{
    return (spim->intfl);
}

#ifdef __cplusplus
}
#endif

#endif /* _SPIM_H_ */
