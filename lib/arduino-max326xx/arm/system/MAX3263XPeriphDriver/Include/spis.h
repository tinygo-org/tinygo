/**
 * @file
 * @brief   SPI Slave High Level API.
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
 * $Date: 2016-10-03 14:23:55 -0500 (Mon, 03 Oct 2016) $
 * $Revision: 24550 $
 *
 **************************************************************************** */

/* Define to prevent redundant inclusion */
#ifndef _SPIS_H_
#define _SPIS_H_

/* **** Includes **** */ 
#include "mxc_config.h"
#include "mxc_sys.h"
#include "spis_regs.h"

#ifdef __cplusplus
extern "C" {
#endif

/**
 * @defgroup spis SPI Slave
 * @ingroup spi_comm
 * @brief SPI Slave (SPIS) Communications Interface API.
 * @{
 */

/* **** Definitions **** */

/** 
 * Enumeration type for setting the number data lines to use for communication. 
 */
typedef enum {
    SPIS_WIDTH_1 = 0,   /**< 1 Data Line */
    SPIS_WIDTH_2 = 1,   /**< 2 Data Lines (x2) */
    SPIS_WIDTH_4 = 2    /**< 4 Data Lines (x4) */
} spis_width_t;

/**
 * Structure type for an SPI Slave transaction request. 
 */
typedef struct spis_req spis_req_t;

/**
 * Callback function type used in asynchromous SPIS communications requests.
 * The function declaration for the SPI Slave callback is:
 * \code 
 * void callback(spis_req_t * req, int error_code);
 * \endcode
 * |        |                                            |
 * | -----: | :----------------------------------------- |
 * | \p req |  Pointer to an #spis_req object representing the active SPIS active transaction. |
 * | \p error_code | An error code if the active transaction had a failure or #E_NO_ERROR if successful. |
 * @addtogroup spis_async
 * @{
 */
typedef void (*callback_fn)(spis_req_t*, int);
/**@} end of async group for this prototype */


/**
 * Struture for the SPI Slave transaction request.
 */
struct spis_req {
    uint8_t             deass;      /**< End the transaction when SS is deasserted. */
    const uint8_t       *tx_data;   /**< TX buffer. */
    uint8_t             *rx_data;   /**< RX buffer. */
    spis_width_t        width;      /**< Number of data lines to use.  */
    unsigned            len;        /**< Number of bytes to send. */
    unsigned            read_num;   /**< Number of bytes transacted. */
    unsigned            write_num;  /**< Number of bytes transacted. */
    callback_fn         callback;   /**< Function pointer to a callback function if desired, NULL otherwise */
};

/* **** Globals **** */

/* **** Function Prototypes **** */

/**
 * @brief      Initialize SPIS module.
 * @param      spis     Pointer to the SPIS register structure.
 * @param      mode     SPI Mode to configure the slave, 0 to 3.
 * @param      sys_cfg  Pointer to system configuration object, see #sys_cfg_spis_t.
 * @return     #E_NO_ERROR if everything is successful.
 */
int SPIS_Init(mxc_spis_regs_t *spis, uint8_t mode, const sys_cfg_spis_t *sys_cfg);


/**
 * @brief      Shutdown SPIS module.
 * @param      spis  Pointer to SPIS regs.
 * @return     #E_NO_ERROR if everything is successful
 */
int SPIS_Shutdown(mxc_spis_regs_t *spis);

/**
 * @brief      Read/write SPIS data. Will block until transaction is complete.
 * @note       Callback is ignored.
 * 
 * @param      spis  Pointer to the SPIS register structure.
 * @param      req   Pointer to a request structure for a SPI Slave transaction.
 * @return     Bytes transacted if everything is successful, @ref
 *             MXC_Error_Codes "error" if unsuccessful.
 */
int SPIS_Trans(mxc_spis_regs_t *spis, spis_req_t *req);

/**
 * @defgroup spis_async SPI Slave Asynchrous Functions
 * @{ 
 */

/**
 * @brief      Asynchronously read/write SPIS data.
 * @note       Request struct, @p req must remain allocated until callback.
 * @param      spis  Pointer to the SPIS register structure.
 * @param      req   Pointer to a request structure for a SPI Slave transaction.
 * @return     #E_NO_ERROR if everything is successful, @ref
 *             MXC_Error_Codes "error" if unsuccessful.
 */
int SPIS_TransAsync(mxc_spis_regs_t *spis, spis_req_t *req);

/**
 * @brief   Abort asynchronous request.
 * @param   req     Pointer to request for a SPIS transaction.
 * @returns #E_NO_ERROR if request aborted, error if unsuccessful.
 */
int SPIS_AbortAsync(spis_req_t *req);

/**
 * @brief   SPI Slave interrupt handler.
 * @details This function should be called by the application from the interrupt
 *          handler if SPIS interrupts are enabled. Alternately, this function
 *          can be periodically called by the application if SPIS interrupts are
 *          disabled.
 * @param   spis    Pointer to the base address of the SPIS[n] module.
 */
void SPIS_Handler(mxc_spis_regs_t *spis);

/**
 * @brief      Check the SPIS to see if it's busy.
 * @param      spis  Pointer to the SPIS register structure.
 * @return     #E_NO_ERROR if idle.
 * @return     #E_BUSY if in use.
 */
int SPIS_Busy(mxc_spis_regs_t *spis);
/**@} end of group spis_async*/

/**
 * @brief      Attempt to prepare the SPI Slave for sleep.
 * @param      spis  Pointer to the SPIS register structure.
 * @details    Checks for any ongoing transactions. Disables interrupts if the
 *             SPI Slave is idle.
 * @return     #E_NO_ERROR if ready to sleep.
 * @return     #E_BUSY if not ready for sleep.
 */
int SPIS_PrepForSleep(mxc_spis_regs_t *spis);

/**
 * @brief      Inline function that enables the SPI Slave without overwriting the existing configuration.
 * @param      spis  Pointer to the SPIS register structure.
 */
__STATIC_INLINE void SPIS_Enable(mxc_spis_regs_t *spis)
{
    spis->gen_ctrl |= (MXC_F_SPIS_GEN_CTRL_SPI_SLAVE_EN |
                       MXC_F_SPIS_GEN_CTRL_TX_FIFO_EN | MXC_F_SPIS_GEN_CTRL_RX_FIFO_EN);
}

/**
 * @brief      Inline function that drains all of the data in the Receive FIFO.
 * @param      spis  Pointer to the SPIS register structure.
 */
__STATIC_INLINE void SPIS_DrainRX(mxc_spis_regs_t *spis)
{
    uint32_t ctrl_save = spis->gen_ctrl;
    spis->gen_ctrl = (ctrl_save & ~MXC_F_SPIS_GEN_CTRL_RX_FIFO_EN);
    spis->gen_ctrl = ctrl_save;
}

/**
 * @brief      Inline function that drains all of the data in the Transmit FIFO.
 * @param      spis  Pointer to the SPIS register structure.
 */
__STATIC_INLINE void SPIS_DrainTX(mxc_spis_regs_t *spis)
{
    uint32_t ctrl_save = spis->gen_ctrl;
    spis->gen_ctrl = (ctrl_save & ~MXC_F_SPIS_GEN_CTRL_TX_FIFO_EN);
    spis->gen_ctrl = ctrl_save;
}

/**
 * @brief      Inline function that returns the Transmit FIFO availability.
 * @param      spis  Pointer to the SPIS register structure.
 * @return     Number of empty bytes available in write FIFO.
 */
__STATIC_INLINE unsigned SPIS_NumWriteAvail(mxc_spis_regs_t *spis)
{
    return (MXC_CFG_SPIS_FIFO_DEPTH - ((spis->fifo_stat &
        MXC_F_SPIS_FIFO_STAT_TX_FIFO_USED) >> MXC_F_SPIS_FIFO_STAT_TX_FIFO_USED_POS));
}

/**
 * @brief      Inline function that returns the Receive FIFO availability.
 * @param      spis  Pointer to the SPIS register structure.
 * @return     Number of bytes in read FIFO.
 */
__STATIC_INLINE unsigned SPIS_NumReadAvail(mxc_spis_regs_t *spis)
{
    return ((spis->fifo_stat & MXC_F_SPIS_FIFO_STAT_RX_FIFO_USED) >>
        MXC_F_SPIS_FIFO_STAT_RX_FIFO_USED_POS);
}

/**
 * @brief      Inline function that clears the specified interrupt flags.
 * @param      spis  Pointer to the SPIS register structure.
 * @param      mask  Mask of interrupts to clear, see the @ref SPIS_INTFL_Register Register for details.
 */
__STATIC_INLINE void SPIS_ClearFlags(mxc_spis_regs_t *spis, uint32_t mask)
{
    spis->intfl = mask;
}

/**
 * @brief   Inline function to get the currently interrupt flags.
 * @param      spis  Pointer to the SPIS register structure.
 * @returns Mask of active flags, see the @ref SPIS_INTFL_Register Register for details.
 */
__STATIC_INLINE unsigned SPIS_GetFlags(mxc_spis_regs_t *spis)
{
    return (spis->intfl);
}
/**@} end of group spis */
#ifdef __cplusplus
}
#endif

#endif /* _SPIS_H_ */
