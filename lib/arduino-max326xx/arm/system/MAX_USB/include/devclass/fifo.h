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
 * $Date: 2016-03-11 11:46:37 -0600 (Fri, 11 Mar 2016) $ 
 * $Revision: 21839 $
 *
 ******************************************************************************/

/**
* \file   fifo.h
* \brief  Driver for management of a software FIFO consisting of 8- or 16-bit elements.
*
*         This file defines the driver API including all types and function prototypes.
*/

#ifndef _FIFO_H_
#define _FIFO_H_

#include <stdint.h>

/***** Definitions *****/

/// Structure used for FIFO management
typedef struct {
  unsigned int length;  ///< FIFO size (number of elements)
  void * data;          ///< pointer to the FIFO buffer
  unsigned int rindex;  ///< current FIFO read index
  unsigned int windex;  ///< current FIFO write index
} fifo_t;

/// Function alias
/// \sa fifo_put8
#define fifo_put   fifo_put8

/// Function alias
/// \sa fifo_get8
#define fifo_get   fifo_get8


/***** Function Prototypes *****/

/// Initializes the specified FIFO
/**
*   \param    fifo     FIFO on which to perform the operation
*   \param    mem      memory buffer to use for FIFO element storage
*   \param    length   number of elements that the memory buffer can contain
*   \returns  0 if successful, -1 upon failure
*/
void fifo_init(fifo_t * fifo, void * mem, unsigned int length);

/**
*   \brief    Adds and 8-bit element to the FIFO
*   \param    fifo     FIFO on which to perform the operation
*   \param    element  element to add to the FIFO
*   \returns  0 if successful, -1 upon failure
*/
int fifo_put8(fifo_t * fifo, uint8_t element);

/**
*   \brief    Gets the next 8-bit element to the FIFO
*   \param    fifo     FIFO on which to perform the operation
*   \param    element  pointer to where to store the element from the FIFO
*   \returns  0 if successful, -1 upon failure
*/
int fifo_get8(fifo_t * fifo, uint8_t * element);

/**
*   \brief    Adds the next 16-bit element to the FIFO
*   \param    fifo     FIFO on which to perform the operation
*   \param    element  element to add to the FIFO
*   \returns  0 if successful, -1 upon failure
*/
int fifo_put16(fifo_t * fifo, uint16_t element);

/**
*   \brief    Gets the next 16-bit element to the FIFO
*   \param    fifo     FIFO on which to perform the operation
*   \param    element  pointer to where to store the element from the FIFO
*   \returns  0 if successful, -1 upon failure
*/
int fifo_get16(fifo_t * fifo, uint16_t * element);

/**
*   \brief    Immediately resets the FIFO to the empty state
*   \param    fifo   FIFO on which to perform the operation
*/
void fifo_clear(fifo_t * fifo);

/**
*   \brief    Determines if the FIFO is empty
*   \param    fifo   FIFO on which to perform the operation
*   \returns  #TRUE if FIFO is empty, #FALSE otherwise
*/
int fifo_empty(fifo_t * fifo);

/**
*   \brief    FIFO status function
*   \param    fifo   FIFO on which to perform the operation
*   \returns  #TRUE if FIFO is full, #FALSE otherwise
*/
int fifo_full(fifo_t * fifo);

/**
*   \brief    FIFO status function
*   \param    fifo   FIFO on which to perform the operation
*   \returns  the number of elements currently in the FIFO
*/
unsigned int fifo_level(fifo_t * fifo);

/**
*   \brief    FIFO status function
*   \param    fifo   FIFO on which to perform the operation
*   \returns  the remaining elements that can be added to the FIFO
*/
unsigned int fifo_remaining(fifo_t * fifo);

#endif /* _FIFO_H_ */
