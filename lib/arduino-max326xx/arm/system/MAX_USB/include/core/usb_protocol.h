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
 * $Date: 2017-01-16 14:05:15 -0600 (Mon, 16 Jan 2017) $
 * $Revision: 25899 $
 *
 ******************************************************************************/

#ifndef _USB_PROTOCOL_H_
#define _USB_PROTOCOL_H_

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

/* SETUP message byte offsets */
#define SETUP_bmRequestType   0
#define	SETUP_bRequest        1
#define SETUP_wValueL         2
#define SETUP_wValueH         3
#define SETUP_wIndexL         4
#define SETUP_wIndexH         5
#define SETUP_wLengthL        6
#define SETUP_wLengthH        7

typedef struct {
  uint8_t  bmRequestType;
  uint8_t  bRequest;
  uint16_t wValue;
  uint16_t wIndex;
  uint16_t wLength;
} usb_setup_pkt;

/* Bitmasks for the bit-field bmRequestType */
#define RT_DEV_TO_HOST            0x80

#define RT_TYPE_MASK              0x60
#define RT_TYPE_STD               0x00
#define RT_TYPE_CLASS             0x20
#define RT_TYPE_VENDOR            0x40

#define RT_RECIP_MASK             0x1f
#define RT_RECIP_DEVICE           0x00
#define RT_RECIP_IFACE            0x01
#define RT_RECIP_ENDP             0x02
#define RT_RECIP_OTHER            0x03

/* Standard Device Requests for bRequest */
#define SDR_GET_STATUS            0x00
#define SDR_CLEAR_FEATURE         0x01
#define SDR_SET_FEATURE           0x03
#define SDR_SET_ADDRESS           0x05
#define SDR_GET_DESCRIPTOR        0x06
#define SDR_SET_DESCRIPTOR        0x07
#define SDR_GET_CONFIG            0x08
#define SDR_SET_CONFIG            0x09
#define SDR_GET_INTERFACE         0x0a
#define SDR_SET_INTERFACE         0x0b
#define SDR_SYNCH_FRAME           0x0c

/* Descriptor types for *_DESCRIPTOR */
#define DESC_DEVICE               1
#define DESC_CONFIG               2
#define DESC_STRING               3
#define DESC_INTERFACE            4
#define DESC_ENDPOINT             5
#define DESC_DEVICE_QUAL          6
#define DESC_OTHER_SPEED          7
#define DESC_IFACE_PWR            8

/* Feature types for *_FEATURE */
#define FEAT_ENDPOINT_HALT        0
#define FEAT_REMOTE_WAKE          1
#define FEAT_TEST_MODE            2

/* Get Status bit positions */
#define STATUS_EP_HALT            0x1
#define STATUS_DEV_SELF_POWERED   0x1
#define STATUS_DEV_REMOTE_WAKE    0x2

/* bmAttributes bit positions */
#define BMATT_REMOTE_WAKE         0x20
#define BMATT_SELF_POWERED        0x40

#if defined(__GNUC__)
typedef struct __attribute__((packed)) {
#else
typedef __packed struct {
#endif
  uint8_t  bLength;
  uint8_t  bDescriptorType;
  uint16_t bcdUSB;
  uint8_t  bDeviceClass;
  uint8_t  bDeviceSubClass;
  uint8_t  bDeviceProtocol;
  uint8_t  bMaxPacketSize;
  uint16_t idVendor;
  uint16_t idProduct;
  uint16_t bcdDevice;
  uint8_t  iManufacturer;
  uint8_t  iProduct;
  uint8_t  iSerialNumber;
  uint8_t  bNumConfigurations;
} usb_device_descriptor_t;

#if defined(__GNUC__)
typedef struct __attribute__((packed)) {
#else
typedef __packed struct {
#endif
  uint8_t  bLength;
  uint8_t  bDescriptorType;
  uint16_t wTotalLength;
  uint8_t  bNumInterfaces;
  uint8_t  bConfigurationValue;
  uint8_t  iConfiguration;
  uint8_t  bmAttributes;
  uint8_t  bMaxPower;
} usb_configuration_descriptor_t;

#if defined(__GNUC__)
typedef struct __attribute__((packed)) {
#else
typedef __packed struct {
#endif
  uint8_t  bLength;
  uint8_t  bDescriptorType;
  uint8_t  bInterfaceNumber;
  uint8_t  bAlternateSetting;
  uint8_t  bNumEndpoints;
  uint8_t  bInterfaceClass;
  uint8_t  bInterfaceSubClass;
  uint8_t  bInterfaceProtocol;
  uint8_t  iInterface;
} usb_interface_descriptor_t;

#define USB_EP_NUM_MASK   0x0F
#define USB_EP_DIR_MASK   0x80

#if defined(__GNUC__)
typedef struct __attribute__((packed)) {
#else
typedef __packed struct {
#endif
  uint8_t  bLength;
  uint8_t  bDescriptorType;
  uint8_t  bEndpointAddress;
  uint8_t  bmAttributes;
  uint16_t wMaxPacketSize;
  uint8_t  bInterval;
} usb_endpoint_descriptor_t;

#if defined(__GNUC__)
typedef struct __attribute__((packed)) {
#else
typedef __packed struct {
#endif
  uint8_t  bLength;
  uint8_t  bDescriptorType;
  uint16_t bcdUSB;
  uint8_t  bDeviceClass;
  uint8_t  bDeviceSubClass;
  uint8_t  bDeviceProtocol;
  uint8_t  bMaxPacketSize;
  uint8_t  bNumConfigurations;
  uint8_t  bReserved;
} usb_device_qualifier_descriptor_t;

#if defined(__GNUC__)
typedef struct __attribute__((packed)) {
#else
typedef __packed struct {
#endif
  uint8_t  bLength;
  uint8_t  bDescriptorType;
  uint16_t wTotalLength;
  uint8_t  bNumInterfaces;
  uint8_t  bConfigurationValue;
  uint8_t  iConfiguration;
  uint8_t  bmAttributes;
  uint8_t  bMaxPower;
} usb_other_speed_configuration_descriptor_t;


#ifdef __cplusplus
}
#endif

#endif
