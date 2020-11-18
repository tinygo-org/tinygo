/**
 * @file
 * @brief   Flash Controller (FLC) function prototypes and data types.
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
 * $Date: 2017-02-16 12:09:00 -0600 (Thu, 16 Feb 2017) $
 * $Revision: 26470 $
 *
 *************************************************************************** */

/* Define to prevent redundant inclusion */
#ifndef _FLC_H_
#define _FLC_H_

/* **** Includes **** */
#include "flc_regs.h"
 
#ifdef __cplusplus
extern "C" {
#endif

/**
 * @ingroup periphlibs
 * @defgroup flc Flash Controller 
 * @brief Internal Flash Controller (FLC) API. 
 * @{
 */

/* **** Definitions **** */

/* **** Globals **** */

/* **** Function Prototypes **** */

/**
 * @brief      Prepares the Flash Controller for in-application flash operations. This function
 *             only needs to be called one time after a reset event.
 *
 * @return     #E_NO_ERROR if flash controller initialized correctly, error if
 *             unsuccessful.
 */
int FLC_Init(void);

/**
 * @brief      This function will erase a single page of flash.
 *
 * @param      address     Address of the page to be erased.
 * @param      erase_code  Flash erase code; defined as
 *                         #MXC_V_FLC_ERASE_CODE_PAGE_ERASE for page erase
 * @param      unlock_key  Unlock key, #MXC_V_FLC_FLSH_UNLOCK_KEY.
 *
 * @returns    #E_NO_ERROR if page erase successful, error if unsuccessful.
 */
int FLC_PageErase(uint32_t address, uint8_t erase_code, uint8_t unlock_key);

/**
 * @brief      This function writes data to the flash device through the flash
 *             controller interface
 *
 * @param      address     Start address for desired write. @note This address 
 * 						   must be 32-bit word aligned                        
 * @param      data        A pointer to the buffer containing the data to write.
 * @param      length      Size of the data to write in bytes. @note The length  
 * 						   must be in 32-bit multiples.  
 * @param      unlock_key  Unlock key, #MXC_V_FLC_FLSH_UNLOCK_KEY.
 *
 * @returns    #E_NO_ERROR if data written successfully, error if unsuccessful.
 */
int FLC_Write(uint32_t address, const void *data, uint32_t length, uint8_t unlock_key);

/**
 * @brief      This function will mass erase the flash.
 *
 * @param      erase_code  Flash erase code, #MXC_V_FLC_ERASE_CODE_MASS_ERASE.
 * @param      unlock_key  Unlock key, #MXC_V_FLC_FLSH_UNLOCK_KEY.
 *
 * @returns    #E_NO_ERROR if device mass erase successful, error if unsuccessful.
 */
int FLC_MassErase(uint8_t erase_code, uint8_t unlock_key);

/**@} end of group flc */

#ifdef __cplusplus
}
#endif

#endif /* _FLC_H_ */
