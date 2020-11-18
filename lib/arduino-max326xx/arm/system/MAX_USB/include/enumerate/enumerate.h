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
 * $Date: 2017-04-05 04:24:37 -0500 (Wed, 05 Apr 2017) $ 
 * $Revision: 27405 $
 *
 ******************************************************************************/
 
#ifndef _ENUMERATE_H_
#define _ENUMERATE_H_

/**
 * @file  enumerate.h
 * @brief USB Device Enumeration Routines
 */

#include <usb_protocol.h>

#ifdef __cplusplus
extern "C" {
#endif

/// User can register callbacks for various control endpoint requests
typedef enum {
  ENUM_CLASS_REQ,
  ENUM_VENDOR_REQ,
  ENUM_SETCONFIG,
  ENUM_SETINTERFACE,
  ENUM_GETINTERFACE,
  ENUM_SETFEATURE,
  ENUM_CLRFEATURE,
  ENUM_NUM_CALLBACKS
} enum_callback_t;

/// User also can register device, config, and string descriptors
typedef enum {
  ENUM_DESC_DEVICE = 0, /// index qualifier ignored
  ENUM_DESC_CONFIG = 1, /// index qualifier ignored
  ENUM_DESC_OTHER = 2,  /// other speed qualifier
  ENUM_DESC_QUAL = 3,   /// device qualifier
  ENUM_DESC_STRING = 4, /// index is used to futher qualify string descriptor
  ENUM_NUM_DESCRIPTORS
} enum_descriptor_t;

/**
 *  \brief    Initialize the enumeration module
 *  \details  Initialize the enumeration module
 *  \return   Zero (0) for success, non-zero for failure
 */
int enum_init(void);

/**
 *  \brief    Register a descriptor
 *  \details  Register a descriptor
 *  \param    type    type of descriptor being registered
 *  \param    desc    pointer to the descriptor
 *  \param    index   index of the string descriptor. ignored for other descriptor types
 *  \return   Zero (0) for success, non-zero for failure
 */
int enum_register_descriptor(enum_descriptor_t type, const uint8_t *desc, uint8_t index);

/**
 *  \brief    Register an enumeration event callback
 *  \details  Register an enumeration event callback
 *  \param    type    event upon which the callback will occur
 *  \param    func    function to be called
 *  \param    cbdata   data to be passed to the callback function
 *  \return   Zero (0) for success, non-zero for failure
 */
int enum_register_callback(enum_callback_t type, int (*func)(usb_setup_pkt *sup, void *cbdata), void *cbdata);

/**
 *  \brief    Register a handler for device class descriptors
 *  \details  Register a handler for devuce class descriptors. The handler is used to respond to device class
 *            get descriptor requests from the host. The handler function shall update desc and desclen with
 *            a pointer to the descriptor and its length.
 *  \param    func    function to be called
 *  \return   Zero (0) for success, non-zero for failure
 */
int enum_register_getdescriptor(void (*func)(usb_setup_pkt *sup, const uint8_t **desc, uint16_t *desclen));

/**
 *  \brief    Gets the current configuration value.
 *  \details  Gets the current configuration value.
 *  \return   The current configuration value.
 */
uint8_t enum_getconfig(void);

/**
 *  \brief    Clears the configuration value.
 *  \details  Clears the configuration value. This function should be called when a host disconnect is detected.
 */
void enum_clearconfig(void);

#ifdef __cplusplus
}
#endif

#endif
