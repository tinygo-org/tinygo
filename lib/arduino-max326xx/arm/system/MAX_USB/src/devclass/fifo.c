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
 * $Date: 2017-02-27 18:16:06 -0600 (Mon, 27 Feb 2017) $ 
 * $Revision: 26712 $
 *
 ******************************************************************************/
 
#include <stdint.h>
#include "fifo.h"
#include "usb_hwopt.h"

#if defined ( __ICCARM__ ) || defined( __GNUC__ ) || defined ( __CC_ARM )
#include <mxc_config.h>
#else
#include <inmaxq.h>
#endif

/****************************************************************************/
void fifo_init(fifo_t * fifo, void * mem, unsigned int length)
{
  // atomic FIFO access
  MAXUSB_ENTER_CRITICAL();

  fifo->rindex = 0;
  fifo->windex = 0;
  fifo->data = mem;
  fifo->length = length;

  MAXUSB_EXIT_CRITICAL();
}

/****************************************************************************/
int fifo_put8(fifo_t * fifo, uint8_t element)
{
  // Check if FIFO is full
  if ( (fifo->windex == (fifo->rindex - 1)) ||
       ((fifo->rindex == 0) && (fifo->windex == (fifo->length - 1))) ) {
    return -1;
  }

  // atomic FIFO access
  MAXUSB_ENTER_CRITICAL();

  // Put data into FIFO
  ((uint8_t*)(fifo->data))[fifo->windex] = element;

  // Increment pointer
  fifo->windex++;
  if (fifo->windex == fifo->length) {
    fifo->windex = 0;
  }

  MAXUSB_EXIT_CRITICAL();
    
  return 0;
}

/****************************************************************************/
int fifo_get8(fifo_t * fifo, uint8_t * element)
{
  // Check if FIFO is empty
  if (fifo->rindex == fifo->windex)
    return -1;

  // atomic FIFO access
  MAXUSB_ENTER_CRITICAL();

   // Get data from FIFO
  *element = ((uint8_t*)(fifo->data))[fifo->rindex];

  // Increment pointer
  fifo->rindex++;
  if (fifo->rindex == fifo->length) {
    fifo->rindex = 0;
  }

  MAXUSB_EXIT_CRITICAL();

  return 0;
}

/****************************************************************************/
int fifo_put16(fifo_t * fifo, uint16_t element)
{
  // Check if FIFO is full
  if ( (fifo->windex == (fifo->rindex - 1)) ||
       ((fifo->rindex == 0) && (fifo->windex == (fifo->length - 1))) ) {
    return -1;
  }

  // atomic FIFO access
  MAXUSB_ENTER_CRITICAL();

  // Put data into FIFO
  ((uint16_t*)(fifo->data))[fifo->windex] = element;

  // Increment pointer
  fifo->windex++;
  if (fifo->windex == fifo->length) {
    fifo->windex = 0;
  }

  MAXUSB_EXIT_CRITICAL();
    
  return 0;
}

/****************************************************************************/
int fifo_get16(fifo_t * fifo, uint16_t * element)
{
  // Check if FIFO is empty
  if (fifo->rindex == fifo->windex)
    return -1;

  // atomic FIFO access
  MAXUSB_ENTER_CRITICAL();

  // Get data from FIFO
  *element = ((uint16_t*)(fifo->data))[fifo->rindex];

  // Increment pointer
  fifo->rindex++;
  if (fifo->rindex == fifo->length) {
    fifo->rindex = 0;
  }

  MAXUSB_EXIT_CRITICAL();
    
  return 0;
}

/****************************************************************************/
void fifo_clear(fifo_t * fifo)
{
  // atomic FIFO access
  MAXUSB_ENTER_CRITICAL();

  fifo->rindex = 0;
  fifo->windex = 0;

  MAXUSB_EXIT_CRITICAL();
}

/****************************************************************************/
int fifo_empty(fifo_t * fifo)
{
  return (fifo->rindex == fifo->windex);
}

/****************************************************************************/
int fifo_full(fifo_t * fifo)
{
  int retval;

  // atomic FIFO access
  MAXUSB_ENTER_CRITICAL();

  retval = ( (fifo->windex == (fifo->rindex - 1)) || ((fifo->rindex == 0) && (fifo->windex == (fifo->length - 1))) );

  MAXUSB_EXIT_CRITICAL();

  return retval;
}

/****************************************************************************/
unsigned int fifo_level(fifo_t * fifo)
{
  uint16_t value;

  // atomic FIFO access
  MAXUSB_ENTER_CRITICAL();

  if (fifo->windex >= fifo->rindex) {
    value = fifo->windex - fifo->rindex;
  } else {
    value = fifo->length - fifo->rindex + fifo->windex;
  }

  MAXUSB_EXIT_CRITICAL();

  return value;
}

/****************************************************************************/
unsigned int fifo_remaining(fifo_t * fifo)
{
  uint16_t value;

  // atomic FIFO access
  MAXUSB_ENTER_CRITICAL();

  if (fifo->rindex > fifo->windex) {
    value = fifo->rindex - fifo->windex - 1;
  } else {
    value = fifo->length - fifo->windex + fifo->rindex - 1;
  }

  MAXUSB_EXIT_CRITICAL();

  return value;
}
