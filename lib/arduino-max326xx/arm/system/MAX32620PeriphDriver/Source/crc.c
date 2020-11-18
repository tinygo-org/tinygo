/**
 * @file
 * @brief   This file contains the function implementations for the Cyclic 
 *          Redundency Check (CRC) peripheral module. 
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
 * $Date: 2016-09-09 11:49:04 -0500 (Fri, 09 Sep 2016) $
 * $Revision: 24339 $
 *
 *************************************************************************** */

/* **** Includes **** */
#include "crc.h"

/**
 * @ingroup crc
 * @{
 */

/* **** Definitions **** */

/* **** Globals **** */

/* **** Functions **** */

/* ************************************************************************* */
void CRC16_Init(uint8_t CCITT_TRUE, uint8_t lilEndian)
{
    if(CCITT_TRUE)
        MXC_CRC->reseed |= MXC_F_CRC_RESEED_CCITT_MODE;
    else
        MXC_CRC->reseed &= ~MXC_F_CRC_RESEED_CCITT_MODE;

    if(lilEndian)
        MXC_CRC->reseed |= MXC_F_CRC_RESEED_REV_ENDIAN16;
    else
        MXC_CRC->reseed &= ~MXC_F_CRC_RESEED_REV_ENDIAN16;
}

/* ************************************************************************* */
void CRC32_Init(uint8_t lilEndian)
{
    if(lilEndian)
        MXC_CRC->reseed |= MXC_F_CRC_RESEED_REV_ENDIAN32;
    else
        MXC_CRC->reseed &= ~MXC_F_CRC_RESEED_REV_ENDIAN32;
}

/* ************************************************************************* */

void CRC16_Reseed(uint16_t initData)
{
    //set initial value
    MXC_CRC->seed16 = initData;

    //reseed the CRC16 generator
    MXC_CRC->reseed |= MXC_F_CRC_RESEED_CRC16;

    //wait for reseed to clear itself
    while(MXC_CRC->reseed & MXC_F_CRC_RESEED_CRC16);

}

/* ************************************************************************* */
void CRC32_Reseed(uint32_t initData)
{
    //set initial value
    MXC_CRC->seed32 = initData;

    //reseed the CRC16 generator
    MXC_CRC->reseed |= MXC_F_CRC_RESEED_CRC32;

    //wait for reseed to clear itself
    while(MXC_CRC->reseed & MXC_F_CRC_RESEED_CRC32);
}

/**@} end of group crc */
