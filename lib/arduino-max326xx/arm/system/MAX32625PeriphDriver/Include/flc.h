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
 * $Date: 2016-03-21 14:44:55 -0500 (Mon, 21 Mar 2016) $
 * $Revision: 22017 $
 *
 ******************************************************************************/

/**
 * @file  flc.h
 * @addtogroup flash Flash Controller
 * @{
 * @brief Internal flash controller module API
 */

#ifndef _FLC_H_
#define _FLC_H_

#ifdef __cplusplus
extern "C" {
#endif

/**
 * @brief Prepares the Flash Controller for flash operations. This function only
 *        needs to be called once after reset.
 *
 * @returns E_NO_ERROR if everything is successful, error if unsuccessful.
 */
int FLC_Init(void);

/**
 * @brief This function will erase a single page of flash.  Keys needed for
 * flash are in the hardware specific register file "flc_regs.h"
 *
 * @param address     Address of the page to be erased.
 * @param erase_code  Flash erase code; defined as 'MXC_V_FLC_ERASE_CODE_PAGE_ERASE' for page erase
 * @param unlock_key  Key necessary for accessing flash; defined as 'MXC_V_FLC_FLSH_UNLOCK_KEY'
 *
 * @returns E_NO_ERROR if everything is successful, error if unsuccessful.
 */
int FLC_PageErase(uint32_t address, uint8_t erase_code, uint8_t unlock_key);

/**
 * @brief This function writes data to the flash device through flash controller
 *
 * @param address     Start address in flash to be written. Must be 32-bit word aligned
 * @param data        Pointer to the buffer containing data to write.
 * @param length      Size of the data to write in bytes. Must be a 32-bit word multiple.
 * @param unlock_key  Key necessary for accessing flash; defined as 'MXC_V_FLC_FLSH_UNLOCK_KEY'
 *
 * @returns E_NO_ERROR if everything is successful, error if unsuccessful.
 */
int FLC_Write(uint32_t address, const void *data, uint32_t length, uint8_t unlock_key);

/**
 * @}
 */

#ifdef __cplusplus
}
#endif

#endif /* _FLC_H_ */
