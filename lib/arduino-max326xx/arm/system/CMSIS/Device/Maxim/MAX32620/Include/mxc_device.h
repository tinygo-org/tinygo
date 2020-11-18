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
 * $Date: 2016-03-23 13:28:53 -0700 (Wed, 23 Mar 2016) $
 * $Revision: 22067 $
 *
 ******************************************************************************/

#ifndef _MXC_DEVICE_H_
#define _MXC_DEVICE_H_

#include "max32620.h"

#ifndef TARGET
#error TARGET NOT DEFINED
#endif

// Create a string definition for the TARGET
#define STRING_ARG(arg) #arg
#define STRING_NAME(name) STRING_ARG(name)
#define TARGET_NAME STRING_NAME(TARGET)

// Define which revisions of the IP we are using
#ifndef TARGET_REV
#error TARGET_REV NOT DEFINED
#endif

#if(TARGET_REV == 0x4332)
// C2
#define MXC_ADC_REV         1
#define MXC_AES_REV         0
#define MXC_CRC_REV         0
#define MXC_FLC_REV         0
#define MXC_GPIO_REV        0
#define MXC_I2CM_REV        0
#define MXC_I2CS_REV        0
#define MXC_ICC_REV         0
#define MXC_MAA_REV         0
#define MXC_OWM_REV         0
#define MXC_PMU_REV         1
#define MXC_PRNG_REV        0
#define MXC_PT_REV          0
#define MXC_RTC_REV         0
#define MXC_SPIM_REV        1
#define MXC_SPIS_REV        0
#define MXC_SPIX_REV        1
#define MXC_TMR_REV         0
#define MXC_UART_REV        1
#define MXC_USB_REV         0
#define MXC_WDT2_REV        0
#define MXC_WDT_REV         0
#else

#error TARGET_REV NOT SUPPORTED

#endif /* if(TARGET_REV == 0x4332) */

#endif  /* _MXC_DEVICE_H_ */
