/**
 * @file
 * @brief    List of common error return codes for Maxim Integrated libraries. 
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
#ifndef _ERRORS_H_
#define _ERRORS_H_

/**
 * @ingroup syscfg
 * @defgroup MXC_Error_Codes Error Codes
 * @brief      A list of common error codes used by the API.
 * @note       A Negative Error Convention is used to avoid conflict with
 *             positive, Non-Error, returns. 
 * @{
 */ 

/** No Error */
#define		E_NO_ERROR		0
/** No Error, success */
#define		E_SUCCESS		0
/** Pointer is NULL */ 
#define		E_NULL_PTR		-1
/** No such device */
#define		E_NO_DEVICE		-2
/** Parameter not acceptable */
#define		E_BAD_PARAM		-3
/** Value not valid or allowed */
#define		E_INVALID		-4
/** Module not initialized */
#define		E_UNINITIALIZED	-5
/** Busy now, try again later */
#define		E_BUSY			-6
/** Operation not allowed in current state */
#define		E_BAD_STATE		-7
/** Generic error */
#define		E_UNKNOWN		-8
/** General communications error */
#define		E_COMM_ERR		-9
/** Operation timed out */
#define		E_TIME_OUT		-10
/** Expected response did not occur */
#define		E_NO_RESPONSE	-11
/** Operations resulted in unexpected overflow */
#define		E_OVERFLOW		-12
/** Operations resulted in unexpected underflow */
#define     E_UNDERFLOW     -13
/** Data or resource not available at this time */
#define		E_NONE_AVAIL	-14
/** Event was shutdown */
#define		E_SHUTDOWN		-15
/** Event was aborted */
#define     E_ABORT         -16
/** The requested operation is not supported */
#define		E_NOT_SUPPORTED	-17
/**@} end of MXC_Error_Codes group */
 
#endif /* _ERRORS_H_ */
