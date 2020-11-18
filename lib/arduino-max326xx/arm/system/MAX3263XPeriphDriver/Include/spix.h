/**
 * @file
 * @brief      This is the high level API for the SPI Execute in Place (SPIX)
 *             module.
 * @note       If using this SPIX with IAR Embedded Workbench for ARM, it is
 *             required to define <tt>IAR_SPIX_PRAGMA=1</tt>. This should be
 *             done under Project->Options-> C/C++ Compiler->Preprocessor in the
 *             Defined Symbols input box. See the IAR documentation for
 *             additional information on how to set a preprocessor define in a
 *             project.
 */
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
 * $Date: 2017-02-16 12:03:11 -0600 (Thu, 16 Feb 2017) $
 * $Revision: 26463 $
 *
 **************************************************************************** */



#include "mxc_sys.h"
#include "spix_regs.h"

/* Define to prevent redundant inclusion */
#ifndef _SPIX_H_
#define _SPIX_H_

#ifdef __cplusplus
extern "C" {
#endif

/**
 * @ingroup periphlibs
 * @defgroup spix SPIX
 * @brief SPI Execute In Place.
 * @{
 */

/* **** Definitions **** */
/**
 * Enumeration type for the number of I/O pins to use during each fetch stage for SPIX
 */
typedef enum {
    SPIX_SINGLE_IO = MXC_V_SPIX_FETCH_CTRL_CMD_WIDTH_SINGLE, /**< 1 I/O pin for SPIX Fetch Cmd Data line  */
    SPIX_DUAL_IO = MXC_V_SPIX_FETCH_CTRL_CMD_WIDTH_DUAL_IO,  /**< 2 I/O pins for SPIX Fetch Cmd Data line */
    SPIX_QUAD_IO = MXC_V_SPIX_FETCH_CTRL_CMD_WIDTH_QUAD_IO   /**< 4 I/O pins for SPIX Fetch Cmd Data line */
} spix_width_t;

/**
 * Enumeration type for the number of address bytes to use during an SPIX fetch
 */
typedef enum {
    SPIX_3BYTE_FETCH_ADDR = 0,              ///< Fetch Address of 3 Bytes
    SPIX_4BYTE_FETCH_ADDR = 1               ///< Fetch Address of 4 Bytes
} spix_addr_size_t;

/**
 * Structure type for SPIX fetch configuration
 */
typedef struct {
    spix_width_t cmd_width;                 ///< Number of I/O lines used for command SPI transaction.
    spix_width_t addr_width;                ///< Number of I/O lines used for address SPI transaction.
    spix_width_t data_width;                ///< Number of I/O lines used for data SPI transaction.
    spix_addr_size_t addr_size;             ///< Use 3 or 4 byte addresses for fetches.
    uint8_t cmd;                            ///< Command value to initiate fetch.
    uint8_t mode_clocks;                    ///< Number of SPI clocks required during mode phase of fetch.
    uint8_t no_cmd_mode;                    ///< Read command sent only once.
    uint16_t mode_data;                     ///< Data sent with mode clocks.
} spix_fetch_t;

/* **** Globals **** */

/* **** Function Prototypes **** */

 /**
  * @brief      Configure SPI execute in place clocking.
  * @param      sys_cfg  Pointer to system level configuration structure.
  * @param      baud     Frequency in hertz to set the clock to. May not be able
  *                      to achieve with the given clock divider.
  * @param      sample   Number of SPIX clocks to delay the sampling of the SDIO
  *                      lines. Will use feedback mode if set to 0.
  * @return     #E_NO_ERROR if everything is successful
  */
int SPIX_ConfigClock(const sys_cfg_spix_t *sys_cfg, uint32_t baud, uint8_t sample);

/**
 * @brief      Configure SPI execute in place slave select.
 * @param      ssel         Index of which slave select line to use.
 * @param      pol          Polarity of slave select (0 for active low, 1 for
 *                          active high).
 * @param      act_delay    SPIX clocks between slave select assert and active
 *                          SPI clock.
 * @param      inact_delay  SPIX clocks between active SPI clock and slave
 *                          select deassert.
 */
void SPIX_ConfigSlave(uint8_t ssel, uint8_t pol, uint8_t act_delay, uint8_t inact_delay);

/**
 * @brief      Configure how the SPIX fetches data.
 * @param      fetch  Pointer to configuration struct that describes how to
 *                    fetch data.
 */
void SPIX_ConfigFetch(const spix_fetch_t *fetch);

/**
 * @brief      Shutdown SPIX module.
 * @param      spix  Pointer to SPIX regs.
 * @return     #E_NO_ERROR if everything is successful
 */
int SPIX_Shutdown(mxc_spix_regs_t *spix);

/**@} end of group spix */
#ifdef __cplusplus
}
#endif

#endif /* _SPIX_H */
