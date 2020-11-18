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
 * @file    prng.h
 * @brief   Pseudo-random number generator(PRNG) driver header file.
 * @note  The PRNG hardware does not produce true random numbers. The output
 *        should be used as a seed to an approved random-number algorithm, per
 *        a certifying authority such as NIST or PCI. The approved algorithm
 *        will output random numbers which are cerfitied for use in encryption
 *        and authentication algorithms.
 */

#ifndef _PRNG_H_
#define _PRNG_H_

/***** Includes *****/
#include "prng_regs.h"

#ifdef __cplusplus
extern "C" {
#endif

/***** Definitions *****/

/***** Globals *****/

/***** Function Prototypes *****/

/**
 * @brief Initialize required clocks and enable PRNG module
 * @note  Function will set divisors to /1 if they are found disabled.
 *        Otherwise, it will not change the divisor.
 */
void PRNG_Init(void);

/**
 * @brief   Returns ready bit to indicates that the PRNG_GetSeed can be called without
 *          being held off. Only needs to be called after POR.
 *
 * @return  0 - PRNG not ready, non-zero = PRNG ready to read
 */
__STATIC_INLINE uint16_t PRNG_Ready(void)
{
    return (MXC_PRNG->user_entropy & MXC_F_PRNG_USER_ENTROPY_RND_NUM_READY);
}

/**
 * @brief Retrieve a seed value from the PRNG
 * @note  The PRNG hardware does not produce true random numbers. The output
 *        should be used as a seed to an approved random-number algorithm, per
 *        a certifying authority such as NIST or PCI. The approved algorithm
 *        will output random numbers which are cerfitied for use in encryption
 *        and authentication algorithms.
 *
 * @return This function returns a 16-bit seed value
 */
__STATIC_INLINE uint16_t PRNG_GetSeed(void)
{
    return MXC_PRNG->rnd_num;
}

/**
 * @brief Add user entropy to the PRNG entropy source
 *
 * @param entropy    This value will be mixed into the PRNG entropy source
 */
__STATIC_INLINE void PRNG_AddUserEntropy(uint8_t entropy)
{
    MXC_PRNG->user_entropy = (uint32_t)entropy;
}

#ifdef __cplusplus
}
#endif

#endif /* _PRNG_H_ */
