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
 * $Date: 2016-03-21 09:04:59 -0500 (Mon, 21 Mar 2016) $
 * $Revision: 22006 $
 *
 ******************************************************************************/

/**
 * @file  aes.h
 * @addtogroup aes AES
 * @{
 * @brief High-level API for AES encryption engine
 *
 * -- Key/data format in memory --
 *
 * These functions expect that key and plain/ciphertext will be stored as a byte array in LSB .. MSB format.
 *
 * As an example, given the key 0x139a35422f1d61de3c91787fe0507afd, the proper storage order is:
 *
 * uint8_t key[16] = {0xfd, 0x7a, 0x50, 0xe0,
 *                    0x7f, 0x78, 0x91, 0x3c,
 *                    0xde, 0x61, 0x1d, 0x2f,
 *                    0x42, 0x35, 0x9a, 0x13};
 *
 * This is the same order expected by the underlying hardware.
 *
 */

#ifndef _AES_H
#define _AES_H

#include <stdint.h>
#include "aes_regs.h"

#ifdef __cplusplus
extern "C" {
#endif

/**
 * @brief Number of bytes in a AES plaintext or cyphertext block (always 128-bits)
 */
#define MXC_AES_DATA_LEN (128 / 8)

/**
 * @brief Number of bytes in a AES key, based on size of key
 */
#define MXC_AES_KEY_128_LEN (128 / 8)
#define MXC_AES_KEY_192_LEN (192 / 8)
#define MXC_AES_KEY_256_LEN (256 / 8)

/**
 * @brief Key size selection (bits)
 */
typedef enum {
    /** 128-bit key */
    MXC_E_AES_MODE_128 = MXC_V_AES_CTRL_KEY_SIZE_128,
    /** 192-bit key */
    MXC_E_AES_MODE_192 = MXC_V_AES_CTRL_KEY_SIZE_192,
    /** 256-bit key */
    MXC_E_AES_MODE_256 = MXC_V_AES_CTRL_KEY_SIZE_256
} mxc_aes_mode_t;

/**
 * @brief Direction select
 */
typedef enum {
    /** Encrypt (blocking) */
    MXC_E_AES_ENCRYPT = 0,
    /** Encrypt (interrupt-driven) */
    MXC_E_AES_ENCRYPT_ASYNC,
    /** Decrypt (blocking) */
    MXC_E_AES_DECRYPT,
    /** Decrypt (interrupt-driven) */
    MXC_E_AES_DECRYPT_ASYNC
} mxc_aes_dir_t;

/**
 * @brief Configure AES block with keying material
 *
 * @param key                 128, 192, or 256 bit keying material
 * @param mode                Selects key length, valid modes found in mxc_aes_mode_t
 */
int AES_SetKey(const uint8_t *key, mxc_aes_mode_t mode);


/**
 * @brief Encrypt/decrypt an input block with the loaded AES key
 *
 * @param in                  Pointer to input array (always 16 bytes)
 * @param out                 Pointer to output array (always 16 bytes)
 */
int AES_ECBOp(const uint8_t *in, uint8_t *out, mxc_aes_mode_t mode, mxc_aes_dir_t dir);

/**
 * @brief Read the AES output memory, used for asynchronous encryption, and clears interrupt flag
 *
 * @param out                 Pointer to output array (always 16 bytes)
 */
int AES_GetOutput(uint8_t *out);

/**
 * @brief Encrypt a block of plaintext with the loaded AES key, blocks until complete
 *
 * @param ptxt                Pointer to plaintext input array (always 16 bytes)
 * @param ctxt                Pointer to ciphertext output array (always 16 bytes)
 * @param mode                Selects key length, valid modes found in mxc_aes_mode_t
 */
#define AES_ECBEncrypt(ptxt, ctxt, mode) AES_ECBOp(ptxt, ctxt, mode, MXC_E_AES_ENCRYPT)


/**
 * @brief Decrypt a block of ciphertext with the loaded AES key, blocks until complete
 *
 * @param ctxt                Pointer to ciphertext output array (always 16 bytes)
 * @param ptxt                Pointer to plaintext input array (always 16 bytes)
 * @param mode                Selects key length, valid modes found in mxc_aes_mode_t
 */
#define AES_ECBDecrypt(ctxt, ptxt, mode) AES_ECBOp(ctxt, ptxt, mode, MXC_E_AES_DECRYPT)

/**
 * @brief Starts encryption of a block, enables interrupt, and returns immediately.
 *        Use AES_GetOuput() to retrieve result after interrupt fires
 *
 *
 * @param ptxt                Pointer to plaintext input array (always 16 bytes)
 * @param mode                Selects key length, valid modes found in mxc_aes_mode_t
 */
#define AES_ECBEncryptAsync(ptxt, mode) AES_ECBOp(ptxt, NULL, mode, MXC_E_AES_ENCRYPT_ASYNC)

/**
 * @brief Starts encryption of a block, enables interrupt, and returns immediately.
 *        Use AES_GetOuput() to retrieve result after interrupt fires
 *
 * @param ctxt                Pointer to ciphertext output array (always 16 bytes)
 * @param mode                Selects key length, valid modes found in mxc_aes_mode_t
 */
#define AES_ECBDecryptAsync(ctxt, mode) AES_ECBOp(ctxt, NULL, mode, MXC_E_AES_DECRYPT_ASYNC)

/**
 * @}
 */

#ifdef __cplusplus
}
#endif

#endif
