/* ******************************************************************************
 * Copyright (C) 2017 Maxim Integrated Products, Inc., All Rights Reserved.
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
 **************************************************************************** */

#ifdef USBCON

#ifndef UsbCdcAcm_h
#define UsbCdcAcm_h

#include <stdint.h>
#include "Arduino.h"
#include "usb.h"
#include "usb_event.h"
#include "enumerate.h"
#include "cdc_acm.h"
#include "HardwareSerial.h"

class UsbCdcAcm : public HardwareSerial
{
public:
    UsbCdcAcm();
    ~UsbCdcAcm(){}
    void begin(unsigned long baud) {}
    void begin(unsigned long, uint8_t) {}
    void end() {}
    int available(void);
    int peek(void);
    int read(void);
    int availableForWrite(void);
    void flush(void){}
    size_t write(uint8_t n);
    operator bool();

private:
    static int configured;
    static int suspended;
    static int remote_wake_en;
    static int peeked;

    static acm_cfg_t acm_cfg;
    static int setconfig_callback(usb_setup_pkt *sud, void *cbdata);
    static int setfeature_callback(usb_setup_pkt *sud, void *cbdata);
    static int clrfeature_callback(usb_setup_pkt *sud, void *cbdata);
    static int event_callback(maxusb_event_t evt, void *data);
};

extern UsbCdcAcm Serial_USB;
extern void usbCdcAcmEventRun(void);

#endif // UsbCdcAcm_h

#endif // USBCON
