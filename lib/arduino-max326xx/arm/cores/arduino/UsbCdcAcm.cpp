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

#include <stddef.h>
#include "mxc_config.h"
#include "mxc_sys.h"
#include "pwrman_regs.h"
#include "lp.h"
#include "usb.h"
#include "usb_event.h"
#include "enumerate.h"
#include "cdc_acm.h"

#include "UsbCdcAcm.h"
#include "usb_descriptors.h"

#define EVENT_ENUM_COMP     MAXUSB_NUM_EVENTS
#define EVENT_REMOTE_WAKE   (EVENT_ENUM_COMP + 1)

int UsbCdcAcm::configured = 0;
int UsbCdcAcm::suspended = 0;
int UsbCdcAcm::remote_wake_en = 0;
int UsbCdcAcm::peeked = -1;

/* This EP assignment must match the Configuration Descriptor */
acm_cfg_t UsbCdcAcm::acm_cfg = {
        1,                  /* EP OUT */
        MXC_USB_MAX_PACKET, /* OUT max packet size */
        2,                  /* EP IN */
        MXC_USB_MAX_PACKET, /* IN max packet size */
        3,                  /* EP Notify */
        MXC_USB_MAX_PACKET, /* Notify max packet size */
};

/* ************************************************************************** */
UsbCdcAcm::UsbCdcAcm()
{
    /* Enable the USB clock and power */
    SYS_USB_Enable(1);

    /* Initialize the usb module */
    if (usb_init(NULL) != 0) {
        return;
    }

    /* Initialize the enumeration module */
    if (enum_init() != 0) {
        return;
    }

    /* Register enumeration data */
    enum_register_descriptor(ENUM_DESC_DEVICE, (uint8_t*)&device_descriptor, 0);
    enum_register_descriptor(ENUM_DESC_CONFIG, (uint8_t*)&config_descriptor, 0);
    enum_register_descriptor(ENUM_DESC_STRING, lang_id_desc, 0);
    enum_register_descriptor(ENUM_DESC_STRING, mfg_id_desc, 1);
    enum_register_descriptor(ENUM_DESC_STRING, prod_id_desc, 2);

    /* Handle configuration */
    enum_register_callback(ENUM_SETCONFIG, setconfig_callback, NULL);

    /* Handle feature set/clear */
    enum_register_callback(ENUM_SETFEATURE, setfeature_callback, NULL);
    enum_register_callback(ENUM_CLRFEATURE, clrfeature_callback, NULL);

    /* Initialize the class driver */
    if (acm_init() != 0) {
        return;
    }

    /* Register callbacks */
    usb_event_enable(MAXUSB_EVENT_NOVBUS, event_callback, NULL);
    usb_event_enable(MAXUSB_EVENT_VBUS, event_callback, NULL);

    NVIC_EnableIRQ(USB_IRQn);
}

/* ************************************************************************** */
int UsbCdcAcm::available(void)
{
   return acm_canread();
}

/* ************************************************************************** */
int UsbCdcAcm::peek(void)
{
    if (acm_present() && acm_canread()) {
        uint8_t peekChar;

        if (peeked == -1) {
            if (acm_read(&peekChar, 1) != 1) {
                return -1;
            }
            peeked = (int)peekChar;
            return peeked;
        }
    }
    return -1;
}

/* ************************************************************************** */
int UsbCdcAcm::read(void)
{
    // Return peeked character, if available
    if (peeked != -1) {
        int tmp = peeked;
        // Reset for next peek
        peeked = -1;
        return tmp;
    }

    if (acm_present() && acm_canread()) {
        uint8_t data;
        if (acm_read(&data, 1) != 1) {
            return -1;
        }
        return (int)data;
    }
    return -1;
}

/* ************************************************************************** */
int UsbCdcAcm::availableForWrite(void)
{
    if (acm_present()) {
        return acm_canwrite();
    }
    return 0;
}

/* ************************************************************************** */
size_t UsbCdcAcm::write(uint8_t n)
{
    if (acm_present()) {
        acm_write(&n, 1);
    }
    return 1;
}

/* ************************************************************************** */
int UsbCdcAcm::setconfig_callback(usb_setup_pkt *sud, void *cbdata)
{
    /* Confirm the configuration value */
    if (sud->wValue == config_descriptor.config_descriptor.bConfigurationValue) {
        configured = 1;
        return acm_configure(&acm_cfg); /* Configure the device class */
    } else if (sud->wValue == 0) {
        configured = 0;
        return acm_deconfigure();
    }

    return -1;
}

/* ************************************************************************** */
int UsbCdcAcm::setfeature_callback(usb_setup_pkt *sud, void *cbdata)
{
    if (sud->wValue == FEAT_REMOTE_WAKE) {
        remote_wake_en = 1;
    } else {
        // Unknown callback
        return -1;
    }

    return 0;
}

/* ************************************************************************** */
int UsbCdcAcm::clrfeature_callback(usb_setup_pkt *sud, void *cbdata)
{
    if (sud->wValue == FEAT_REMOTE_WAKE) {
        remote_wake_en = 0;
    } else {
        // Unknown callback
        return -1;
    }

    return 0;
}

/* ************************************************************************** */
int UsbCdcAcm::event_callback(maxusb_event_t evt, void *data)
{
    switch (evt) {
        case MAXUSB_EVENT_NOVBUS:
            usb_event_disable(MAXUSB_EVENT_BRST);
            usb_event_disable(MAXUSB_EVENT_SUSP);
            usb_event_disable(MAXUSB_EVENT_DPACT);
            usb_disconnect();
            configured = 0;
            enum_clearconfig();
            acm_deconfigure();
            break;
        case MAXUSB_EVENT_VBUS:
            usb_event_clear(MAXUSB_EVENT_BRST);
            usb_event_enable(MAXUSB_EVENT_BRST, event_callback, NULL);
            usb_event_clear(MAXUSB_EVENT_SUSP);
            usb_event_enable(MAXUSB_EVENT_SUSP, event_callback, NULL);
            usb_connect();
            break;
        case MAXUSB_EVENT_BRST:
            enum_clearconfig();
            acm_deconfigure();
            configured = 0;
            suspended = 0;
            break;
        case MAXUSB_EVENT_SUSP:
            break;
        case MAXUSB_EVENT_DPACT:
            break;
        default:
            break;
    }
    return 0;
}

/* ************************************************************************** */
UsbCdcAcm::operator bool()
{
    return !!acm_present();
}

/* ************************************************************************** */
extern "C" void USB_IRQHandler(void) { usb_event_handler(); }

/* ************************************************************************** */
void usbCdcAcmEventRun(void)
{
    if (Serial.available()) serialEvent();
}

UsbCdcAcm Serial_USB;
HardwareSerial &Serial = Serial_USB;

#endif // USBCON
