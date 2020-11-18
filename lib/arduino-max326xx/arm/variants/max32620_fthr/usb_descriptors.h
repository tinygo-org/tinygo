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
 * Description: Communications Device Class ACM (Serial Port) over USB
 *
 *******************************************************************************
 */

#ifdef USBCON

#ifndef _DESCRIPTORS_H_
#define _DESCRIPTORS_H_

#include <stdint.h>
#include "usb.h"

usb_device_descriptor_t __attribute__((aligned(4))) device_descriptor = {
    0x12,         /* bLength = 18                     */
    0x01,         /* bDescriptorType = Device         */
    0x0110,       /* bcdUSB USB spec rev (BCD)        */
    0x02,         /* bDeviceClass = comm class (2)    */
    0x00,         /* bDeviceSubClass                  */
    0x00,         /* bDeviceProtocol                  */
    0x40,         /* bMaxPacketSize0 is 64 bytes      */
    0x0B6A,       /* idVendor (Maxim Integrated)      */
    0x003C,       /* idProduct                        */
    0x0100,       /* bcdDevice                        */
    0x01,         /* iManufacturer Descriptor ID      */
    0x02,         /* iProduct Descriptor ID           */
    0x00,         /* iSerialNumber = (0) No string    */
    0x01          /* bNumConfigurations               */
};

__attribute__((aligned(4)))
struct __attribute__((packed)) {
    usb_configuration_descriptor_t  config_descriptor;
    usb_interface_descriptor_t      comm_interface_descriptor;
    uint8_t                         header_functional_descriptor[5];
    uint8_t                         call_management_descriptor[5];
    uint8_t                         acm_functional_descriptor[4];
    uint8_t                         union_functional_descriptor[5];
    usb_endpoint_descriptor_t       endpoint_descriptor_3;
    usb_interface_descriptor_t      data_interface_descriptor;
    usb_endpoint_descriptor_t       endpoint_descriptor_1;
    usb_endpoint_descriptor_t       endpoint_descriptor_2;
} config_descriptor =
{
    {
        0x09,       /*  bLength = 9                     */
        0x02,       /*  bDescriptorType = Config (2)    */
        0x0043,     /*  wTotalLength(L/H)               */
        0x02,       /*  bNumInterfaces                  */
        0x01,       /*  bConfigValue                    */
        0x00,       /*  iConfiguration                  */
        0xE0,       /*  bmAttributes (self-powered, remote wakeup) */
        0x01,       /*  MaxPower is 2ma (units are 2ma/bit) */
    },
    { /*  First Interface Descriptor For Comm Class Interface */
        0x09,       /*  bLength = 9                     */
        0x04,       /*  bDescriptorType = Interface (4) */
        0x00,       /*  bInterfaceNumber                */
        0x00,       /*  bAlternateSetting               */
        0x01,       /*  bNumEndpoints (one for OUT)     */
        0x02,       /*  bInterfaceClass = Communications Interface Class (2) */
        0x02,       /*  bInterfaceSubClass = Abstract Control Model (2) */
        0x01,       /*  bInterfaceProtocol = Common "AT" commands (1), no class specific protocol (0) */
        0x00,       /*  iInterface                      */
    },
    { /*  Header Functional Descriptor */
        0x05,         /*  bFunctionalLength = 5           */
        0x24,         /*  bDescriptorType                 */
        0x00,         /*  bDescriptorSubtype              */
        0x10, 0x01,   /*  bcdCDC                          */
    },
    { /*  Call Management Descriptor */
        0x05,         /*  bFunctionalLength = 5           */
        0x24,         /*  bDescriptorType                 */
        0x01,         /*  bDescriptorSubtype              */
        0x03,         /*  bmCapabilities = Device handles call management itself (0x01), management over data class (0x02) */
        0x01,         /*  bmDataInterface                 */
    },
    { /*  Abstract Control Management Functional Descriptor */
        0x04,         /*  bFunctionalLength = 4           */
        0x24,         /*  bDescriptorType                 */
        0x02,         /*  bDescriptorSubtype              */
        0x02,         /*  bmCapabilities                  */
    },
    { /*  Union Functional Descriptor */
        0x05,         /*  bFunctionalLength = 5           */
        0x24,         /*  bDescriptorType                 */
        0x06,         /*  bDescriptorSubtype              */
        0x00,         /*  bmMasterInterface               */
        0x01,         /*  bmSlaveInterface0               */
    },
    { /*  IN Endpoint 3 (Descriptor #1) */
        0x07,         /*  bLength                          */
        0x05,         /*  bDescriptorType (Endpoint)       */
        0x83,         /*  bEndpointAddress (EP3-IN)        */
        0x03,         /*  bmAttributes (interrupt)         */
        0x0040,       /*  wMaxPacketSize                   */
        0xff,         /*  bInterval (milliseconds)         */
    },
    { /*  Second Interface Descriptor For Data Interface */
        0x09,         /*  bLength                          */
        0x04,         /*  bDescriptorType (Interface)      */
        0x01,         /*  bInterfaceNumber                 */
        0x00,         /*  bAlternateSetting                */
        0x02,         /*  bNumEndpoints                    */
        0x0a,         /*  bInterfaceClass = Data Interface (10) */
        0x00,         /*  bInterfaceSubClass = none (0)    */
        0x00,         /*  bInterfaceProtocol = No class specific protocol (0) */
        0x00,         /*  biInterface = No Text String (0) */
    },
    { /*  OUT Endpoint 1 (Descriptor #2) */
        0x07,         /*  bLength                          */
        0x05,         /*  bDescriptorType (Endpoint)       */
        0x01,         /*  bEndpointAddress (EP1-OUT)       */
        0x02,         /*  bmAttributes (bulk)              */
        0x0040,       /*  wMaxPacketSize                   */
        0x00,         /*  bInterval (N/A)                  */
    },
    { /*  IN Endpoint 2 (Descriptor #3) */
        0x07,         /*  bLength                          */
        0x05,         /*  bDescriptorType (Endpoint)       */
        0x82,         /*  bEndpointAddress (EP2-IN)        */
        0x02,         /*  bmAttributes (bulk)              */
        0x0040,       /*  wMaxPacketSize                   */
        0x00          /*  bInterval (N/A)                  */
    }
};

__attribute__((aligned(4)))
uint8_t lang_id_desc[] = {
    0x04,         /* bLength */
    0x03,         /* bDescriptorType */
    0x09, 0x04    /* bString = wLANGID (see usb_20.pdf 9.6.7 String) */
};

__attribute__((aligned(4)))
uint8_t mfg_id_desc[] = {
    0x22,         /* bLength */
    0x03,         /* bDescriptorType */
    'M', 0,
    'a', 0,
    'x', 0,
    'i', 0,
    'm', 0,
    ' ', 0,
    'I', 0,
    'n', 0,
    't', 0,
    'e', 0,
    'g', 0,
    'r', 0,
    'a', 0,
    't', 0,
    'e', 0,
    'd', 0
};

__attribute__((aligned(4)))
uint8_t prod_id_desc[] = {
    0x1A,         /* bLength */
    0x03,         /* bDescriptorType */
    'M', 0,
    'A', 0,
    'X', 0,
    '3', 0,
    '2', 0,
    '6', 0,
    '2', 0,
    '0', 0,
    'F', 0,
    'T', 0,
    'H', 0,
    'R', 0
};

/* Not currently used (see device descriptor), but could be enabled if desired */
__attribute__((aligned(4)))
uint8_t serial_id_desc[] = {
    0x14,         /* bLength */
    0x03,         /* bDescriptorType */
    '0', 0,
    '0', 0,
    '0', 0,
    '0', 0,
    '0', 0,
    '0', 0,
    '0', 0,
    '0', 0,
    '1', 0
};

#endif /* _DESCRIPTORS_H_ */

#endif // USBCON
