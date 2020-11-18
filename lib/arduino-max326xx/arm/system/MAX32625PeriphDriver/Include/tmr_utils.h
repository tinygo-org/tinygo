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
 * $Date: 2016-03-21 09:04:59 -0500 (Mon, 21 Mar 2016) $
 * $Revision: 22006 $
 *
 ******************************************************************************/

/**
 * @file    tmr_utils.h
 * @brief   Timer utility functions.
 */

#ifndef _TMR_UTILS_H
#define _TMR_UTILS_H

/***** Includes *****/
#include "mxc_config.h"
#include "tmr_regs.h"
#include "mxc_sys.h"

#ifdef __cplusplus
extern "C" {
#endif

/***** Definitions *****/
#define SEC(s)            (((unsigned long)s) * 1000000UL)
#define MSEC(ms)          (ms * 1000UL)
#define USEC(us)          (us)

/***** Globals *****/

/***** Function Prototypes *****/

/**
 * @brief   Delays for the specified number of microseconds.
 * @param   tmr     TMR module to operate on
 * @param   us      Number of microseconds to delay.
 */
void TMR_Delay(mxc_tmr_regs_t* tmr, unsigned long us);

/**
 * @brief   Start the timeout time for the specified number of microseconds.
 * @param   tmr     TMR module to operate on
 * @param   us      Number of microseconds in the timeout.
 */
void TMR_TO_Start(mxc_tmr_regs_t* tmr, unsigned long us);

/**
 * @brief   Check if the timeout has occured.
 * @param   tmr     TMR module to operate on
 * @returns E_NO_ERROR if the timeout has not occurred, E_TIME_OUT if it has.
 */
int TMR_TO_Check(mxc_tmr_regs_t* tmr);

/**
 * @brief   Stops the timer for the timeout.
 * @param   tmr     TMR module to operate on
 */
void TMR_TO_Stop(mxc_tmr_regs_t* tmr);

/**
 * @brief   Clears the timeout flag.
 * @param   tmr     TMR module to operate on
 */
void TMR_TO_Clear(mxc_tmr_regs_t* tmr);

/**
 * @brief   Get the number of microseconds elapsed since to_start().
 * @param   tmr     TMR module to operate on
 * @returns Number of microseconds since to_start().
 */
unsigned TMR_TO_Elapsed(mxc_tmr_regs_t* tmr);

/**
 * @brief   Get the number of microseconds remaining in the timeout.
 * @param   tmr     TMR module to operate on
 * @returns Number of microseconds since to_start().
 */
unsigned TMR_TO_Remaining(mxc_tmr_regs_t* tmr);

/**
 * @brief   Start the stopwatch.
 */
void TMR_SW_Start(mxc_tmr_regs_t* tmr);

/**
 * @brief   Stop the stopwatch and return the number of microseconds that have elapsed.
 * @param   tmr     TMR module to operate on
 * @returns Number of microseconds since sw_start().
 */
unsigned TMR_SW_Stop(mxc_tmr_regs_t* tmr);

#ifdef __cplusplus
}
#endif

#endif /* _TMR_UTILS_H */
