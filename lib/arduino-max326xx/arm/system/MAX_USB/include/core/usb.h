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
 
#ifndef _USB_H_
#define _USB_H_

#include "usb_hwopt.h"
#include "usb_protocol.h"

#ifdef __cplusplus
extern "C" {
#endif

/**
 * @file usb.h
 * @brief Defines the API used to abstract USB hardware away from upper layers.
 *
 */

/******************************** Definitions *********************************/

/*
 * Endpoint types, additional types may be supported in the future.
 * 
 * On non-configurable hardware, an error will be returned if the type selected
 * disagrees with hardware capability on that endpoint (ie. can't make an IN
 * into an OUT).
 *
 */
typedef enum {
  MAXUSB_EP_TYPE_DISABLED = 0,
  MAXUSB_EP_TYPE_OUT      = 1,
  MAXUSB_EP_TYPE_IN       = 2,
  MAXUSB_EP_TYPE_CONTROL  = 3
} maxusb_ep_type_t;

/*
 * USB events. Register for callbacks with usb_register_callback().
 */
typedef enum {
  MAXUSB_EVENT_DPACT = 0, /* D+ Activity */
  MAXUSB_EVENT_RWUDN,     /* Remote Wake-Up Signaling Done */
  MAXUSB_EVENT_BACT,      /* Bus Active */
  MAXUSB_EVENT_BRST,      /* Bus Reset */
  MAXUSB_EVENT_SUSP,      /* Suspend */
  MAXUSB_EVENT_NOVBUS,    /* No VBUS - VBUSDET signal makes 0 -> 1 transition i.e. VBUS not present */
  MAXUSB_EVENT_VBUS,      /* VBUS present */
  MAXUSB_EVENT_BRSTDN,    /* Bus Reset Done */
  MAXUSB_EVENT_SUDAV,     /* Setup Data Available */
  MAXUSB_NUM_EVENTS
} maxusb_event_t;

/*
 * USB events flags.
 */
typedef struct {
  /* Non-endpoint events */
  unsigned int dpact  : 1;
  unsigned int rwudn  : 1;
  unsigned int bact   : 1;
  unsigned int brst   : 1;
  unsigned int susp   : 1;
  unsigned int novbus : 1;
  unsigned int vbus   : 1;
  unsigned int brstdn : 1;
  unsigned int sudav  : 1;
} maxusb_usbio_events_t;

/*
 * USB Request Type
 */
typedef enum {
  MAXUSB_TYPE_TRANS = 0,  /* The request will complete once the requested amount
                           * of data has been received, or when a packet is
                           * received containing less than max packet.
                           */
  MAXUSB_TYPE_PKT         /* The request will complete each time a packet is
                           * received. The caller is responsible for zero-packet
                           * handling
                           */
} maxusb_req_type_t;

/*
 * Object for requesting an endpoint read or write. The object is updated with
 * the transaction status and can be observed when the callback is called.
 */
typedef struct {
  unsigned int ep;
  uint8_t *data;
  unsigned int reqlen; // requested / max length
  unsigned int actlen; // actual length transacted
  int error_code;
  void (*callback)(void *);
  void *cbdata;
  maxusb_req_type_t type;
  void *driver_xtra; /* driver-specific data, do not modify */
} usb_req_t;


/**************************** Function Prototypes *****************************/

/**
 * @brief Initialize the USB hardware to a non-connected, "ready" state
 *
 * @param options    Hardware-specific options which are in each chip's usb_hwopt.h
 * @return This function returns zero (0) for success, non-zero for failure
 * 
 */
int usb_init(maxusb_cfg_options_t *options);

/** 
 * @brief Shut down the USB peripheral block
 *
 *  This function will shut down the USB IP. Once called, you must call
 *  usb_init() to bring the block back into a working condition. No
 *  state persists after this call, including endpoint configuration,
 *  pending reads/writes, etc. All pending and outstanding events will be 
 *  quashed at the driver layer.
 * 
 * @return This function returns zero (0) for success, non-zero for failure
 *
 */
int usb_shutdown(void);

/** 
 * @brief Connect to the USB bus
 * @return This function returns zero (0) for success, non-zero for failure
 *
 */
int usb_connect(void);

/** 
 * @brief Disconnect from the USB bus
 * @return This function returns zero (0) for success, non-zero for failure
 *
 */
int usb_disconnect(void);

/** 
 * @brief Endpoint configuration function
 *
 * Endpoints can be Disabled, CONTROL, BULK/INTERRUPT IN, or BULK/INTERRUPT OUT. 
 *  No hardware support for ISOCHRONOUS exists currently, but may appear in the future.
 *
 * An endpoint has a configured size, which should match that advertised to the host in
 * the Device Descriptor.
 *
 * @param ep    endpoint number
 * @param type  endpoint type
 * @param size  endpoint size
 * @return This function returns zero (0) for success, non-zero for failure
 *
 */
int usb_config_ep(unsigned int ep, maxusb_ep_type_t type, unsigned int size);

/** 
 * @brief Query the configuration status of the selected endpoint
 * @param ep   endpoint number
 * @return This function returns 1 if the endpoint is configured, 0 if it is not, and < 0 for error
 *
 */
int usb_is_configured(unsigned int ep);

/** 
 * @brief Stall the selected endpoint
 *
 * If the endpoint is the CONTROL endpoint, then both the IN and OUT pipes are stalled.
 * In this case, the hardware will also stall the Status stage of the transfer.
 *
 * @param ep   endpoint number
 * @return This function returns zero (0) for success, non-zero for failure
 *
 */
int usb_stall(unsigned int ep);

/** 
 * @brief Unstall the selected endpoint
 * 
 * If this endpoint is the CONTROL endpoint, the IN, OUT, and Status stage stall bits are cleared.
 * This is not normally needed, as hardware should clear these bits upon reception of the
 * next SETUP packet.
 *
 * @param ep   endpoint number
 * @return This function returns zero (0) for success, non-zero for failure
 *
 */
int usb_unstall(unsigned int ep);

/** 
 * @brief Query the stalled/unstalled status of the selected endpoint
 * @param ep   endpoint number
 * @return This function returns 0 if the endpoint is not stalled, 1 if it is, and < 0 for error
 *
 */
int usb_is_stalled(unsigned int ep);

/** 
 * @brief Reset state and clear the data toggle on the selected endpoint
 * @param ep   endpoint number
 * @return This function returns zero (0) for success, non-zero for failure
 *
 */
int usb_reset_ep(unsigned int ep);

/** 
 * @brief Arm the hardware to ACK the Status stage of a SETUP transfer. Only valid for CONTROL endpoints.
 * @param ep   endpoint number
 * @return This function returns zero (0) for success, non-zero for failure
 *
 */
int usb_ackstat(unsigned int ep);

/** 
 * @brief Enable the specified event interrupt in hardware
 * This function is called by the event layer through usb_event_enable() and
 * should not be called directly from the application.
 * @param event   event number
 * @return This function returns zero (0) for success, non-zero for failure
 *
 */
int usb_irq_enable(maxusb_event_t event);

/** 
 * @brief Enable the specified event interrupt in hardware
 * This function is called by the event layer through usb_event_disable() and
 * should not be called directly from the application.
 * @param event   event number
 * @return This function returns zero (0) for success, non-zero for failure
 *
 */
int usb_irq_disable(maxusb_event_t event);

/** 
 * @brief Clear the specified interrupt flag in hardware
 * This function is called by the event layer through usb_event_clear() and
 * should not be called directly from the application.
 * @param event   event number
 * @return This function returns zero (0) for success, non-zero for failure
 *
 */
int usb_irq_clear(maxusb_event_t event);

/** 
 * @brief Interrupt handler. 
 * This function is called by the event layer through usb_event_handler() and
 * should not be called directly from the application.
 * This function will read the interrupt flags, handle any outstanding
 * I/O in a chip-specific way (DMA, Programmed I/O, Bus Master, etc.)
 * The event structure is returned to the upper layer so that it may react to 
 * bus conditions.
 * @param evt   structure of event flags
 */
void usb_irq_handler(maxusb_usbio_events_t *events);

/** 
 * @brief Read the SETUP data from hardware
 * @param setup_pkt   Pointer to setup data structure
 * @return This function returns zero (0) for success, non-zero for failure
 *
 */
int usb_get_setup(usb_setup_pkt *setup_pkt);

/** 
 * @brief Change the function address, in response to a Host request
 * @param addr The 7-bit address in the SET_ADDRESS request
 * @return This function returns zero (0) for success, non-zero for failure
 * @note Some hardware does this automatically; In that case, this function will always return 0.
 *
 */
int usb_set_func_addr(unsigned int addr);

/** 
 * @brief Returns a pointer to the request queued for the specified endpoint
 * @param ep Endpoint which to query
 * @return Pointer to request structure, or NULL if none queued
 *
 */
usb_req_t *usb_get_request(unsigned int ep);

/** 
 * @brief Send data to the host via the selected endpoint
 * 
 * This asynchronous function allows the caller to specify a buffer for outgoing data.
 * The buffer may have any length, and device-specific code will handle breaking the data 
 * into endpoint-sized chunks.  The request object and data buffer passed to this function must
 * remain "owned" by the USB stack until the callback function is called indicating completion.
 * It will handle a zero-length length, as that is a valid message on the USB to indicate 
 * success during various phases of data transfer.
 *
 * Once called, the next IN transaction processed by hardware on the selected endpoint
 * will cause the IN DATAx payload to be sent from the provided buffer. The driver will
 * keep track of how many bytes have been sent and continue sending additional chunks of
 * data until all data has been sent to the host. The driver will send a zero-length packet 
 * if the data to be sent is a whole multiple (no remainder after division) of the endpoint size.
 *
 * Upon completion of the request, the request object's error_code and actlen fields are
 * updated to reflect the result of the transaction and the function specified in the request
 * object is be called.
 * 
 * Only one outstanding buffer is allowed to exist in the current implementation. This function
 * will return an error to the caller if it finds that there is an outstanding buffer already
 * configured. This will not affect the outstanding buffer.
 *
 * A special case exists for this call: If the data pointer is NULL, then any existing
 * outstanding buffer is removed from the driver. This allows for a "disarming" mechanism
 * without shutting down the entire stack and re-starting.
 *
 * @param req   Initialized request object
 * 
 * @return This function returns zero (0) for success, non-zero for failure
 *
 */
int usb_write_endpoint(usb_req_t *req);

/** 
 * @brief Arm the selected endpoint to receive data from the host
 * 
 * This asynchronous function allows the caller to specify a buffer for incoming data.
 * The driver will read from the endpoint at most len bytes into the provided buffer.
 *
 * Once called, the next OUT transaction processed by hardware on the selected endpoint
 * will cause the OUT DATAx payload to be loaded into the provided buffer. Additional OUT
 * DATAx payloads will be concatenated to the buffer until 1) len bytes have been read, or
 * 2) a DATAx payload of length less than the maximum endpoint size has been read. The 
 * latter case signifies the end of a USB transaction. If case #1 is reached before the end
 * of the USB transfer, any additional data is thrown away by the driver layer.
 *
 * Upon completion of the request, the request object's error_code and actlen fields are
 * updated to reflect the result of the transaction and the function specified in the request
 * object is be called.
 * 
 * Only one outstanding buffer is allowed to exist in the current implementation. This function
 * will return an error to the caller if it finds that there is an outstanding buffer already
 * configured. This will not affect the outstanding buffer.
 *
 * A special case exists for this call: If the data pointer is NULL, then any existing
 * outstanding buffer is removed from the driver. This allows for a "disarming" mechanism
 * without shutting down the entire stack and re-starting.
 *
 * @param req   Initialized request object
 * 
 * @return This function returns zero (0) for success, non-zero for failure
 *
 */
int usb_read_endpoint(usb_req_t *req);

#ifdef __cplusplus
}
#endif

#endif /* _USB_H_ */
