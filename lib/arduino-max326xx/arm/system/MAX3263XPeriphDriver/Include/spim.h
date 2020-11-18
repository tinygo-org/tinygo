/**
 * @file
 * @brief Registers, Bit Masks and Bit Positions for the SPI Master module.
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
 * $Date: 2017-02-16 12:03:47 -0600 (Thu, 16 Feb 2017) $
 * $Revision: 26464 $
 *
 **************************************************************************** */

/* **** Includes **** */ 
#include "mxc_config.h"
#include "mxc_sys.h"
#include "spim_regs.h"

/* Define to prevent redundant inclusion */
#ifndef _SPIM_H_
#define _SPIM_H_
 
#ifdef __cplusplus
extern "C" {
#endif

/**
 * @ingroup    periphlibs
 * @defgroup   spim SPI Master
 * @brief      Serial Peripheral Interface Master (SPIM) API.
 * @{
 */

/* **** Definitions **** */

/** 
 * Enumeration type for selecting the active levels for the SPI Master Slave Select (SS) lines. 
 */
typedef enum {
    SPIM_SSEL0_HIGH  =    (0x1 << 0), /**<  Slave Select 0 High. */
    SPIM_SSEL0_LOW   =    0,          /**<  Slave Select 0 Low.  */
    SPIM_SSEL1_HIGH  =    (0x1 << 1), /**<  Slave Select 1 High. */
    SPIM_SSEL1_LOW   =    0,          /**<  Slave Select 1 Low.  */
    SPIM_SSEL2_HIGH  =    (0x1 << 2), /**<  Slave Select 2 High. */
    SPIM_SSEL2_LOW   =    0,          /**<  Slave Select 2 Low.  */
    SPIM_SSEL3_HIGH  =    (0x1 << 3), /**<  Slave Select 3 High. */
    SPIM_SSEL3_LOW   =    0,          /**<  Slave Select 3 Low.  */
    SPIM_SSEL4_HIGH  =    (0x1 << 4), /**<  Slave Select 4 High. */
    SPIM_SSEL4_LOW   =    0           /**<  Slave Select 4 Low.  */
}
spim_ssel_t;

/** 
 * Enumeration type for setting the number data lines to use for communication. 
 */
typedef enum {
    SPIM_WIDTH_1 = 0,  /**< 1 Data Line.                        */
    SPIM_WIDTH_2 = 1,  /**< 2 Data Lines (x2).                  */
    SPIM_WIDTH_4 = 2   /**< 4 Data Lines (x4).                  */
} spim_width_t;

/** 
 * Structure type for configuring a SPIM port. 
 */
typedef struct {
    uint8_t     mode;       /**< SPIM mode selection, 0 to 3.                                           */
    uint32_t    ssel_pol;   /**< Mask of active levels for the slave select signals, see #spim_ssel_t.  */
    uint32_t    baud;       /**< Baud rate in Hz.                                                       */
} spim_cfg_t;

/** 
 * Structure type representing a SPI Master Transaction request.
 */
typedef struct spim_req spim_req_t;

/**
 * @brief Callback function type used in asynchromous SPIM communications requests.
 * @details The function declaration for the SPIM callback is:
 * @code 
 * void callback(spim_req_t * req, int error_code);
 * @endcode
 * |        |                                            |
 * | -----: | :----------------------------------------- |
 * | \p req |  Pointer to a #spim_req object representing the active SPIM active transaction. |
 * | \p error_code | An error code if the active transaction had a failure or #E_NO_ERROR if successful. |
 * @addtogroup spim_async
 */
typedef void (*spim_callback_fn)(spim_req_t * req, int error_code);

/**
 * @brief      Structure definition for an SPI Master Transaction request.
 * @note       When using this structure for an asynchronous operation, the
 *             structure must remain allocated until the callback is completed.
 * @addtogroup spim_async
 */
struct spim_req {
    uint8_t             ssel;       /**< Number of the Slave Select to use. */
    uint8_t             deass;      /**< Set to de-assert slave select at the completions of the transaction.*/
    const uint8_t       *tx_data;   /**< Pointer to a buffer to transmit data from. */
    uint8_t             *rx_data;   /**< Pointer to a buffer to store data received. */
    spim_width_t        width;      /**< Number of data lines to use, see #spim_width_t. */
    unsigned            len;        /**< Number of bytes to send from the \p tx_data buffer. */
    unsigned            read_num;   /**< Number of bytes read and stored in \p rx_data buffer. */
    unsigned            write_num;  /**< Number of bytes sent from the \p tx_data buffer, this will be filled by the driver after up to \p len bytes have been transmitted. */
    spim_callback_fn    callback;   /**< Function pointer to a callback function if desired, NULL otherwise */
};

/* **** Globals **** */

/* **** Function Prototypes **** */

/**
 * @brief      Initialize the SPIM peripheral module.
 *
 * @param      spim     Pointer to the SPIM register structure.
 * @param      cfg      Pointer to an SPIM configuration object.
 * @param      sys_cfg  Pointer to a system configuration object to select the
 *                      peripheral clock rate and assign the requested GPIO.
 *
 * @return     #E_NO_ERROR if the SPIM port is initialized successfully, @ref MXC_Error_Codes
 *             "error" if unsuccessful.
 */
int SPIM_Init(mxc_spim_regs_t *spim, const spim_cfg_t *cfg, const sys_cfg_spim_t *sys_cfg);

/**
 * @brief      Shutdown the SPIM peripheral module instance represented by the
 *             @p spim parameter.
 *
 * @param      spim  Pointer to the SPIM register structure.
 *
 * @return     #E_NO_ERROR if the SPIM is shutdown successfully, @ref
 *             MXC_Error_Codes "error" if unsuccessful.
 */
int SPIM_Shutdown(mxc_spim_regs_t *spim);

/**
 * @brief      Send Clock cycles on SCK without reading or writing.
 *
 * @param      spim   Pointer to the SPIM register structure.
 * @param      len    Number of clock cycles to send.
 * @param      ssel   Slave select number.
 * @param      deass  De-assert slave select at the end of the transaction.
 *
 * @return     Cycles transacted if everything is successful, @ref
 *             MXC_Error_Codes "error" if unsuccessful.
 */
int SPIM_Clocks(mxc_spim_regs_t *spim, uint32_t len, uint8_t ssel, uint8_t deass);

/**
 * @brief      Read/write SPIM data. This function will block until the
 *             transaction is complete.
 *
 * @param      spim  Pointer to the SPIM register structure.
 * @param      req   Request for a SPIM transaction.
 * @note       If a callback function is registered it will not be called when using a blocking function.
 *
 * @return     Bytes transacted if everything is successful, error if
 *             unsuccessful.
 */
int SPIM_Trans(mxc_spim_regs_t *spim, spim_req_t *req);
/**
 * @defgroup spim_async SPIM Asynchrous Functions
 * @{ 
 */
/**
 * @brief      Asynchronously read/write SPIM data.
 *
 * @param      spim  Pointer to the SPIM register structure.
 * @param      req   Request for a SPIM transaction.
 * @note       Request struct must remain allocated until callback.
 *
 * @return     #E_NO_ERROR if everything is successful, @ref MXC_Error_Codes
 *             "error" if unsuccessful.
 */
int SPIM_TransAsync(mxc_spim_regs_t *spim, spim_req_t *req);

/**
 * @brief      Abort asynchronous request.
 *
 * @param      req   Pointer to a request structure for a SPIM transaction.
 *
 * @return     #E_NO_ERROR if request aborted, , @ref MXC_Error_Codes "error" if
 *             unsuccessful.
 */
int SPIM_AbortAsync(spim_req_t *req);

/**
 * @brief      SPIM interrupt handler.
 * @details    This function should be called by the application from the
 *             interrupt handler if SPIM interrupts are enabled. Alternately,
 *             this function can be periodically polled by the application if
 *             SPIM interrupts are disabled.
 *
 * @param      spim  Base address of the SPIM module.
 */
void SPIM_Handler(mxc_spim_regs_t *spim);

/**
 * @brief      Check the SPIM to see if it's busy.
 *
 * @param      spim         Pointer to the SPIM register structure.
 *
 * @retval     #E_NO_ERROR  if idle.
 * @retval     #E_BUSY      if in use.
 */
int SPIM_Busy(mxc_spim_regs_t *spim);
/**@} end of spim_async define group */

/**
 * @brief      Attempts to prepare the SPIM for Low Power Sleep Modes.
 * @details    Checks for any ongoing transactions. Disables interrupts if the
 *             SPIM is idle.
 *
 * @param      spim  The spim
 *
 * @return     #E_NO_ERROR if ready to sleep.
 * @return     #E_BUSY if not able to sleep at this time.
 */
int SPIM_PrepForSleep(mxc_spim_regs_t *spim);

/**
 * @brief      Enables the SPIM without overwriting the existing configuration.
 *
 * @param      spim  Pointer to the SPIM register structure.
 */
__STATIC_INLINE void SPIM_Enable(mxc_spim_regs_t *spim)
{
    spim->gen_ctrl |= (MXC_F_SPIM_GEN_CTRL_SPI_MSTR_EN |
                       MXC_F_SPIM_GEN_CTRL_TX_FIFO_EN | MXC_F_SPIM_GEN_CTRL_RX_FIFO_EN);
}

/**
 * @brief      Drains/empties the data in the RX FIFO.
 *
 * @param      spim  Pointer to the SPIM register structure.
 */
__STATIC_INLINE void SPIM_DrainRX(mxc_spim_regs_t *spim)
{
    uint32_t ctrl_save = spim->gen_ctrl;
    spim->gen_ctrl = (ctrl_save & ~MXC_F_SPIM_GEN_CTRL_RX_FIFO_EN);
    spim->gen_ctrl = ctrl_save;
}

/**
 * @brief      Drains/empties the data in the TX FIFO.
 *
 * @param      spim  Pointer to the SPIM register structure.
 */
__STATIC_INLINE void SPIM_DrainTX(mxc_spim_regs_t *spim)
{
    uint32_t ctrl_save = spim->gen_ctrl;
    spim->gen_ctrl = (ctrl_save & ~MXC_F_SPIM_GEN_CTRL_TX_FIFO_EN);
    spim->gen_ctrl = ctrl_save;
}

/**
 * @brief      Returns the number of bytes free in the TX FIFO.
 *
 * @param      spim  Pointer to the SPIM register structure.
 *
 * @return     Number of bytes free in Transmit FIFO.
 */
__STATIC_INLINE unsigned SPIM_NumWriteAvail(mxc_spim_regs_t *spim)
{
    return (MXC_CFG_SPIM_FIFO_DEPTH - ((spim->fifo_ctrl &
                                        MXC_F_SPIM_FIFO_CTRL_TX_FIFO_USED) >> MXC_F_SPIM_FIFO_CTRL_TX_FIFO_USED_POS));
}

/**
 * @brief      Returns the number of bytes available to read in the RX FIFO.
 *
 * @param      spim  Pointer to the SPIM register structure.
 *
 * @return     Number of bytes in RX FIFO.
 */
__STATIC_INLINE unsigned SPIM_NumReadAvail(mxc_spim_regs_t *spim)
{
    return ((spim->fifo_ctrl & MXC_F_SPIM_FIFO_CTRL_RX_FIFO_USED) >>
            MXC_F_SPIM_FIFO_CTRL_RX_FIFO_USED_POS);
}

/**
 * @brief      Clear the SPIM interrupt flags.
 *
 * @param      spim  Pointer to the SPIM register structure.
 * @param      mask  Mask of the SPIM interrupt flags to clear, see @ref
 *                   SPIM_INTFL_Register Register for the SPIM interrupt flag
 *                   bit masks.
 */
__STATIC_INLINE void SPIM_ClearFlags(mxc_spim_regs_t *spim, uint32_t mask)
{
    spim->intfl = mask;
}

/**
 * @brief      Read the current SPIM interrupt flags.
 *
 * @param      spim  Pointer to the SPIM register structure.
 *
 * @return     Mask of currently set SPIM interrupt flags, see @ref
 *             SPIM_INTFL_Register Register for the SPIM interrupt flag bit
 *             masks.
 */
__STATIC_INLINE unsigned SPIM_GetFlags(mxc_spim_regs_t *spim)
{
    return (spim->intfl);
}

/**@} end of group spim_comm */
#ifdef __cplusplus
}
#endif

#endif /* _SPIM_H_ */
