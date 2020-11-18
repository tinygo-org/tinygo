/**
 * @file
 * @brief    Exclusive access lock utility functions.
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
 * $Date: 2017-02-14 18:16:40 -0600 (Tue, 14 Feb 2017) $
 * $Revision: 26426 $
 *
 *************************************************************************** */

/* Define to prevent redundant inclusion */
#ifndef _MXC_LOCK_H_
#define _MXC_LOCK_H_

/* **** Includes **** */
#include "mxc_config.h"
#include "mxc_lock.h"

#ifdef __cplusplus
extern "C" {
#endif

/* ************************************************************************** */
int mxc_get_lock(uint32_t *lock, uint32_t value)
{
    do {

        // Return if the lock is taken by a different thread
        if(__LDREXW((volatile unsigned long *)lock) != 0) {
            return E_BUSY;
        }

        // Attempt to take the lock
    } while(__STREXW(value, (volatile unsigned long *)lock) != 0);

    // Do not start any other memory access until memory barrier is complete
    __DMB();

    return E_NO_ERROR;
}

/* ************************************************************************** */
void mxc_free_lock(uint32_t *lock)
{
    // Ensure memory operations complete before releasing lock
    __DMB();
    *lock = 0;
}

#ifdef __cplusplus
}
#endif
#endif /* _MXC_LOCK_H_ */
