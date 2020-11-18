/**
 * @file
 * @brief      Advanced Encryption Standard (AES) function prototypes and data
 *             types.
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
 * $Date: 2016-10-10 16:51:05 -0500 (Mon, 10 Oct 2016) $
 * $Revision: 24655 $
 *
  *************************************************************************** */

/* Define to prevent redundant inclusion */
#ifndef _AES_H
#define _AES_H
/* **** Includes **** */
#include <stdint.h>
#include "aes_regs.h"

#ifdef __cplusplus
extern "C" {
#endif

/**
 * @ingroup    periphlibs
 * @defgroup   aes Advanced Encryption Standard (AES)
 * @brief      High-level API for AES encryption engine 
 */

/**
 * @ingroup aes
 * @defgroup aes_overview Overview and Usage
 * @brief Advanced Encryption Standard API public include file.
 * @details
 * <b>Key/data format in memory</b>
 * The API functions require that key and plain/ciphertext will be stored as a 
 * byte array in LSB .. MSB format.
 * @par
 * As an example, given the key @a 0x139A35422F1D61DE3C91787FE0507AFD, the proper storage order is:
 * ~~~~~
 * uint8_t key[16] = { 0xFD, 0x7A, 0x50, 0xE0,
 *                     0x7F, 0x78, 0x91, 0x3C,
 *                     0xDE, 0x61, 0x1D, 0x2F,
 *                     0x42, 0x35, 0x9A, 0x13 };
 * ~~~~~
 * This is the same order expected by the underlying hardware.
 */

/* **** Definitions **** */
/**
 * @ingroup aes
 * @{
 */
#define MXC_AES_DATA_LEN (128 / 8)                      /**< Number of bytes in an AES plaintext or cyphertext block, which are always 128-bits long. */
#define MXC_AES_KEY_128_LEN (128 / 8)                   /**< Number of bytes in a AES-128 key. */
#define MXC_AES_KEY_192_LEN (192 / 8)                   /**< Number of bytes in a AES-192 key. */
#define MXC_AES_KEY_256_LEN (256 / 8)                   /**< Number of bytes in a AES-256 key. */

/**
 * Enumeration type for AES key size selection (bits).
 */
typedef enum {
    MXC_E_AES_MODE_128 = MXC_V_AES_CTRL_KEY_SIZE_128,   /**< 128-bit key. */
    MXC_E_AES_MODE_192 = MXC_V_AES_CTRL_KEY_SIZE_192,   /**< 192-bit key. */
    MXC_E_AES_MODE_256 = MXC_V_AES_CTRL_KEY_SIZE_256    /**< 256-bit key. */
} mxc_aes_mode_t;

/**
 * Enumeration type for specifying encryption/decrytion and asynchronous or blocking behavior.  
 */
typedef enum {
    MXC_E_AES_ENCRYPT       = 0,                        /**< Encrypt (synchronous/blocking).         */
    MXC_E_AES_ENCRYPT_ASYNC = 1,                        /**< Encrypt (aynchronous/interrupt-driven). */
    MXC_E_AES_DECRYPT       = 2,                        /**< Decrypt (synchronous/blocking).         */
    MXC_E_AES_DECRYPT_ASYNC = 3                         /**< Decrypt (aynchronous/interrupt-driven). */
} mxc_aes_dir_t;

/* **** Function Prototypes **** */

/**
 * @brief Configure AES block with keying material
 *
 * @param       key             128, 192, or 256 bit keying material
 * @param       mode            The key length, see #mxc_aes_mode_t for supported lengths.
 * 
 * @return      #E_BAD_PARAM    Specified @a mode is invalid, see #mxc_aes_mode_t.
 * @return      #E_NULL_PTR     Invalid/Null pointer for parameter @a key.
 * @return      #E_SUCCESS      Key and mode set up correctly.
 */
int AES_SetKey(const uint8_t *key, mxc_aes_mode_t mode);


/**
 * @brief       Encrypt/decrypt an input block with the loaded AES key.
 * @note        The parameters @a in and @a out must be 16 bytes. 
 *
 * @param       in              Pointer to input array of 16 bytes.
 * @param       out             Pointer to output array of 16 bytes.
 * @param       mode            AES key size to use for the transaction, see #mxc_aes_mode_t for supported key sizes. 
 * @param       dir             Operation to perform, see #mxc_aes_dir_t for supported operations.
 * 
 * @return      #E_SUCCESS      Operation completed successfully, output data is stored in @a *out.
 * @return      ErrorCode       An @ref MXC_Error_Codes "Error Code" if an error occured.
 */
int AES_ECBOp(const uint8_t *in, uint8_t *out, mxc_aes_mode_t mode, mxc_aes_dir_t dir);

/**
 * @brief      Read the AES output memory, used for asynchronous encryption, and
 *             clears interrupt flag.
 * @note       The parameter @a out must always be 16 bytes.
 *
 * @param      out   Pointer to a 16-byte array to store the output from the AES operation. 
 *
 * @return     #E_SUCCESS      Output data was written to the location pointed
 *             to by @a *out.
 * @return     A @ref MXC_Error_Codes "Error Code" indicating the error that
 *             occured.
 */
int AES_GetOutput(uint8_t *out);

/**
 * @def AES_ECBEncrypt(ptxt, ctxt, mode)
 * @brief      Encrypt a block of plaintext with the loaded AES key, blocks
 *             until complete.
 * @hideinitializer
 *
 * @param      ptxt  Pointer to plaintext input array (always 16 bytes)
 * @param      ctxt  Pointer to ciphertext output array (always 16 bytes)
 * @param      mode  Selects key length, valid modes found in mxc_aes_mode_t
 */
#define AES_ECBEncrypt(ptxt, ctxt, mode) AES_ECBOp(ptxt, ctxt, mode, MXC_E_AES_ENCRYPT)


/**
 * @def AES_ECBDecrypt(ctxt, ptxt, mode)
 * @hideinitializer
 * @brief      Decrypt a block of ciphertext with the loaded AES key, blocks
 *             until complete.
 *
 * @param      ctxt  Pointer to ciphertext output array (always 16 bytes)
 * @param      ptxt  Pointer to plaintext input array (always 16 bytes)
 * @param      mode  Selects key length, valid modes found in mxc_aes_mode_t
 */
#define AES_ECBDecrypt(ctxt, ptxt, mode) AES_ECBOp(ctxt, ptxt, mode, MXC_E_AES_DECRYPT)

/**
 * @def AES_ECBEncryptAsync(ptxt, mode)
 * @hideinitializer
 * @brief      Starts encryption of a block, enables interrupt, and returns
 *             immediately. Use AES_GetOuput() to retrieve result after
 *             interrupt fires
 *
 *
 * @param      ptxt  Pointer to plaintext input array (always 16 bytes)
 * @param      mode  Selects key length, valid modes found in mxc_aes_mode_t
 */
#define AES_ECBEncryptAsync(ptxt, mode) AES_ECBOp(ptxt, NULL, mode, MXC_E_AES_ENCRYPT_ASYNC)

/**
 * @def AES_ECBDecryptAsync(ctxt, mode)
 * @hideinitializer
 * @brief      Starts encryption of a block, enables interrupt, and returns
 *             immediately. Use AES_GetOuput() to retrieve result after
 *             interrupt fires
 *
 * @param      ctxt  Pointer to ciphertext output array (always 16 bytes)
 * @param      mode  Selects key length, valid modes found in mxc_aes_mode_t
 */
#define AES_ECBDecryptAsync(ctxt, mode) AES_ECBOp(ctxt, NULL, mode, MXC_E_AES_DECRYPT_ASYNC)

/**@} end of group aes*/

#ifdef __cplusplus
}
#endif

#endif
