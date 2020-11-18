/**
 * @file
 * @brief   CRC peripheral module function prototypes and data types.
 */

/* ****************************************************************************
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
 * $Date: 2016-10-10 16:57:56 -0500 (Mon, 10 Oct 2016) $
 * $Revision: 24657 $
 *
 **************************************************************************** */

/* Define to prevent redundant inclusion */
#ifndef _CRC_H_
#define _CRC_H_

#ifdef __cplusplus
extern "C" {
#endif

/**
 * @ingroup periphlibs
 * @defgroup crc Cyclic Redundancy Check (CRC)
 * @brief High-level API for CRC Peripheral Module
 * @{
 */

/**
 * @page crc_overview Overview and Usage
 * @parblock
 *  - Initialize the CRC for usage by calling either the CRC16_Init() function or the CRC32_Init() funciton
 *    + For CRC16, CCITT mode is enabled if the parameter is set
 *    + Select the Endian of the data for calculation
 *      - passing a 1 sets the module to little endian
 *      - passing a 0 sets the module to big endian
 *  - Set the intial CRC seed by calling CRC16_Reseed() or CRC32_Reseed()
 *    + This is the initial value of the CRC remainder to be used when the data is passed to the module
 *  - Pass data to the CRC engine using the methods:
 *    + CRC16_AddData()
 *    + CRC16_AddDataArray()
 *    + CRC32_AddData()
 *    + CRC32_AddDataArray()
 *      @note any data passed to the add methods of the peripheral is padded with zeros if it is less than 
 *      32-bits.
 * @endparblock
*/


/* **** Includes **** */
#include "mxc_config.h"
#include <string.h>
#include "crc_regs.h"

/* **** Definitions **** */

/* **** Globals **** */

/* **** Function Prototypes **** */

/**
 * @brief Initialize CRC clock and select CRC16 mode and byte order.
 *
 * @param CCITT_TRUE    CRC16-CCITT-TRUE = 1, CRC16-CCITT-FALSE = 0
 * @param lilEndian     byte order, little endian = 1, big endian = 0
 */
void CRC16_Init(uint8_t CCITT_TRUE, uint8_t lilEndian);

/**
 * @brief Initialize CRC clock and select byte order for CRC32.
 * 
 * @param lilEndian  byte order, little endian = 1, big endian = 0
 */
void CRC32_Init(uint8_t lilEndian);

/**
 * @brief Initialize CRC16 calculation.
 *
 * @param initData intial remainder to start the CRC16 calculation with
 */
void CRC16_Reseed(uint16_t initData);

/**
 * @brief Initialize CRC32 calculation.
 *
 * @param initData intial remainder to start the CRC32 calculation with
 */
void CRC32_Reseed(uint32_t initData);

/**
 * @brief Add data to the CRC16 calculation.
 *
 * @param data    data to add to the CRC16 calculation
 * @note  data is padded with zeros if less than 32bits.
 */
__STATIC_INLINE void CRC16_AddData(uint32_t data)
{
    MXC_CRC_DATA->value16[0] = data;
}

/**
 * @brief Add data to the CRC32 calculation
 *
 * @param data    data to add to the CRC32 calculation
 * @note  data is padded with zeros if less than 32bits
 */
__STATIC_INLINE void CRC32_AddData(uint32_t data)
{
    MXC_CRC_DATA->value32[0] = data;
}

/**
 * @brief Add an array of data to the CRC16 calculation
 *
 * @param data          pointer to array of data
 * @note  data is padded with zeros if less than 32bits
 * 
 * @param arrayLength   number of elements in array
 */
__STATIC_INLINE void CRC16_AddDataArray(uint32_t *data, uint32_t arrayLength)
{
    memcpy((void *)(&(MXC_CRC_DATA->value16)), (void *)data, arrayLength * sizeof(data[0]));
}

/**
 * @brief Add an array of data to the CRC32 calculation
 *
 * @param data          pointer to array of data
 * @note  data is padded with zeros if less than 32bits
 * @param arrayLength   number of elements in array
 */
__STATIC_INLINE void CRC32_AddDataArray(uint32_t *data, uint32_t arrayLength)
{
    memcpy((void *)(&(MXC_CRC_DATA->value32)), (void *)data, arrayLength * sizeof(data[0]));
}

/**
 * @brief Get the calculated CRC16 value
 *
 * @return CRC16 value
 */
__STATIC_INLINE uint32_t CRC16_GetCRC()
{
    return MXC_CRC_DATA->value16[0];
}

/**
 * @brief Get the calculated CRC32 value
 *
 * @return CRC32 value
 */
__STATIC_INLINE uint32_t CRC32_GetCRC()
{
    return MXC_CRC_DATA->value32[0];
}

/**@} end of crc group */

#ifdef __cplusplus
}
#endif

#endif /* _CRC_H_ */
