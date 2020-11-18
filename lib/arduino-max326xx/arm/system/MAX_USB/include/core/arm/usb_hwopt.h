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
 * $Date: 2017-01-16 13:33:16 -0600 (Mon, 16 Jan 2017) $ 
 * $Revision: 25896 $
 *
 ******************************************************************************/
 
#ifndef _USBIO_HWOPT_H_
#define _USBIO_HWOPT_H_

#include "mxc_config.h"
#include "usb_regs.h"

/* There are no configuration options for this target */
typedef void maxusb_cfg_options_t;

#define MAXUSB_ENTER_CRITICAL() __disable_irq()
#define MAXUSB_EXIT_CRITICAL() __enable_irq()

/** 
 * @brief Put the transceiver into a low power state.
 */
static inline void usb_sleep(void)
{
    MXC_USB->dev_cn |= MXC_F_USB_DEV_CN_ULPM;
}

/** 
 * @brief Power up the USB transceiver, must be called once the device wakes from sleep.
 */
static inline void usb_wakeup(void)
{
    MXC_USB->dev_cn &= ~MXC_F_USB_DEV_CN_ULPM;
}

/** 
 * @brief Send a remote wakeup signal to the host.
 */
static inline void usb_remote_wakeup(void)
{
    MXC_USB->dev_cn |= MXC_F_USB_DEV_CN_SIGRWU;
}

#endif /* _USBIO_HWOPT_H_ */
