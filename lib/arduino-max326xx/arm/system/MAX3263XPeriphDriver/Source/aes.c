/**
 * @file
 * @brief      This file contains the function implementations for the Advanced 
 *             Encryption Standard (AES) peripheral module.
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
 * $Date: 2016-09-09 12:50:17 -0500 (Fri, 09 Sep 2016) $
 * $Revision: 24348 $
 *
 *************************************************************************** */

/* **** Includes **** */
#include <string.h>    /* Included for memcpy() & #includes stddef for NULL */

#include "mxc_config.h"
#include "aes.h"
#include "nvic_table.h"

/**
 * @ingroup aes
 * @{
 */

/* **** Definitions **** */

/* **** Globals **** */

/* **** Local Function Prototypes **** */
static int aes_memcpy32(uint32_t *out, uint32_t *in, unsigned int count);

/* **** Functions **** */

/* ************************************************************************* */
int AES_SetKey(const uint8_t *key, mxc_aes_mode_t mode)
{
    unsigned int len;

    /* Erase any existing key */
    MXC_AES_MEM->key[7] = MXC_AES_MEM->key[6] = MXC_AES_MEM->key[5] = MXC_AES_MEM->key[4] \
                          = MXC_AES_MEM->key[3] = MXC_AES_MEM->key[2] = MXC_AES_MEM->key[1] = MXC_AES_MEM->key[0] \
                                  = 0x00000000;

    /* Determine length of key */
    if (mode == MXC_E_AES_MODE_256) {
        len = MXC_AES_KEY_256_LEN;
    } else if (mode == MXC_E_AES_MODE_192) {
        len = MXC_AES_KEY_192_LEN;
    } else if (mode == MXC_E_AES_MODE_128) {
        len = MXC_AES_KEY_128_LEN;
    } else {
        return E_BAD_PARAM;
    }

    /* Load new key, based on key mode */
    if (aes_memcpy32((uint32_t *)MXC_AES_MEM->key, (uint32_t *)key, len / sizeof(uint32_t)) < 0) {
        return E_NULL_PTR;
    }

    return E_SUCCESS;
}

/* ************************************************************************* */
int AES_ECBOp(const uint8_t *in, uint8_t *out, mxc_aes_mode_t mode, mxc_aes_dir_t dir)
{
    /* Output array can't be a NULL, unless we are in _ASYNC mode */
    if ((out == NULL)
            && ((dir != MXC_E_AES_ENCRYPT_ASYNC) && (dir != MXC_E_AES_DECRYPT_ASYNC))) {
        return E_NULL_PTR;
    }

    /* Another encryption is already in progress */
    if (MXC_AES->ctrl & MXC_F_AES_CTRL_START) {
        return E_BUSY;
    }

    /* Clear interrupt flag and any existing configuration*/
    MXC_AES->ctrl = MXC_F_AES_CTRL_INTFL;

    /* Select key size & direction
     *
     * Note: This is done first to detect argument errors, before sensitive data
     *  is loaded into AES_MEM block
     *
     */
    switch (mode) {
        case MXC_E_AES_MODE_128:
            MXC_AES->ctrl |= MXC_S_AES_CTRL_KEY_SIZE_128;
            break;

        case MXC_E_AES_MODE_192:
            MXC_AES->ctrl |= MXC_S_AES_CTRL_KEY_SIZE_192;
            break;

        case MXC_E_AES_MODE_256:
            MXC_AES->ctrl |= MXC_S_AES_CTRL_KEY_SIZE_256;
            break;

        default:
            return E_BAD_PARAM;
    }

    switch (dir) {
        case MXC_E_AES_ENCRYPT:
        case MXC_E_AES_ENCRYPT_ASYNC:
            MXC_AES->ctrl |= MXC_S_AES_CTRL_ENCRYPT_MODE;
            break;

        case MXC_E_AES_DECRYPT:
        case MXC_E_AES_DECRYPT_ASYNC:
            MXC_AES->ctrl |= MXC_S_AES_CTRL_DECRYPT_MODE;
            break;

        default:
            return E_BAD_PARAM;
    }

    /* If non-blocking mode has been selected, interrupts are automatically enabled */
    if ((dir == MXC_E_AES_ENCRYPT_ASYNC) ||
            (dir == MXC_E_AES_DECRYPT_ASYNC)) {
        MXC_AES->ctrl |= MXC_F_AES_CTRL_INTEN;
    }

    /* Load input into engine */
    if (aes_memcpy32((uint32_t *)MXC_AES_MEM->inp, (uint32_t *)in, MXC_AES_DATA_LEN / sizeof(uint32_t)) < 0) {
        return E_NULL_PTR;
    }

    /* Start operation */
    MXC_AES->ctrl |= MXC_F_AES_CTRL_START;

    /* Block, waiting on engine to complete, or fall through if non-blocking */
    if ((dir != MXC_E_AES_ENCRYPT_ASYNC) &&
            (dir != MXC_E_AES_DECRYPT_ASYNC)) {
        while (MXC_AES->ctrl & MXC_F_AES_CTRL_START) {
            /* Ensure that this wait loop is not optimized out */
            __NOP();
        }

        /* Get output from engine */
        return AES_GetOutput(out);
    }

    return E_SUCCESS;
}

/* ************************************************************************* */
int AES_GetOutput(uint8_t *out)
{
    /* Don't read it out of the AES memory unless engine is idle */
    if (MXC_AES->ctrl & MXC_F_AES_CTRL_START) {
        return E_BUSY;
    }

    /* Pull out result */
    if (aes_memcpy32((uint32_t *)out, (uint32_t *)MXC_AES_MEM->out, MXC_AES_DATA_LEN / sizeof(uint32_t)) < 0) {
        return E_NULL_PTR;
    }

    /* Clear interrupt flag, write 1 to clear */
    MXC_AES->ctrl |= MXC_F_AES_CTRL_INTFL;

    return E_SUCCESS;
}

/**
 * @internal This memory copy is used only by the AES module to avoid data leakage by the standard C library.
 * Copy count number of 32-bit locations from in to out 
 */
static int aes_memcpy32(uint32_t *out, uint32_t *in, unsigned int count)
{
    if ((out == NULL) || (in == NULL)) {
        /* Invalid arguments, but is internal-only so don't use error codes */
        return -1;
    }

    while (count--) {
        *out++ = *in++;
    }

    return 0;
}

/**@} end of group aes */
