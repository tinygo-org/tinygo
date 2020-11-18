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
*******************************************************************************
*/

/**
 * @file  maa.h
 * @addtogroup maa MAA
 * @{
 * @brief This is the high level API for the Modular Math Accelerator (MAA)
 *
 * @details The API only supports synchronous operations due to the sensitive 
 *          nature of the input and output data. 
 *        
 */

#ifndef _MAA_H
#define _MAA_H

#include <stdint.h>

#include "maa_regs.h"

#ifdef __cplusplus
extern "C" {
#endif

/**
 * @brief Maximum MAA register size on this device (128 bytes or 1024 bits)
 */
#define MXC_MAA_REG_SIZE      0x80
#define MXC_MAA_REG_SIZE_BITS (MXC_MAA_REG_SIZE << 3)

/**
 * @brief Sub-register ("half size"), allowing 2x more operands in MAA at a time when MAWS <= MAX_SIZE/2
 */
#define MXC_MAA_HALF_SIZE (MXC_MAA_REG_SIZE/2)

/**
 * @brief Flags for MAA_Load and MAA_Unload
 */
#define MXC_MAA_F_MEM_VERBATIM                        0
#define MXC_MAA_F_MEM_REVERSE                         1

/**
 * @brief (Sub-)Segment select
 */
/* Warning: Do not change the assigned numbers/ordering without associated changes to UMAA_REGFILE_TO_ADDR(x) */
typedef enum {
  /* Register names when MAWS > 512 */ 
  MXC_E_REG_0 = 0,
  MXC_E_REG_1 = 2,
  MXC_E_REG_2 = 4,
  MXC_E_REG_3 = 6,
  MXC_E_REG_4 = 8,
  MXC_E_REG_5 = 10,
  /* Register names when MAWS < 512 */
  MXC_E_REG_00 = 0,
  MXC_E_REG_01 = 1,
  MXC_E_REG_10 = 2,
  MXC_E_REG_11 = 3,
  MXC_E_REG_20 = 4,
  MXC_E_REG_21 = 5,
  MXC_E_REG_30 = 6,
  MXC_E_REG_31 = 7,
  MXC_E_REG_40 = 8,
  MXC_E_REG_41 = 9,
  MXC_E_REG_50 = 10,
  MXC_E_REG_51 = 11
} mxc_maa_reg_select_t;

/**
 * @brief Operation select
 */
typedef enum {
  /** Exponentiate */
  MXC_E_MAA_EXP = 0,
  /** Square */
  MXC_E_MAA_SQR = 1,
  /** Multiply */
  MXC_E_MAA_MUL = 2,
  /** Square followed by Multiply */
  MXC_E_MAA_SQRMUL = 3,
  /** Addition */
  MXC_E_MAA_ADD = 4,
  /** Subtraction */
  MXC_E_MAA_SUB = 5
} mxc_maa_operation_t;

/**
 * @brief Special flags for loading & unloading data
 */
typedef enum {
  /** Copy bytes without reversal and right-justification */
  MXC_E_MAA_VERBATIM = 0,
  /** Reverse bytes and right-justify (bytes are loaded at the highest address, then descending) */
  MXC_E_MAA_REVERSE
} mxc_maa_endian_select_t;

/**
 * @brief Standardized return codes for the MAA module
 */
typedef enum {
  /** Error */
  MXC_E_MAA_ERR = -1,
  /** No error */
  MXC_E_MAA_OK = 0,
  /** Engine busy, try again later */
  MXC_E_MAA_BUSY
} mxc_maa_ret_t;

/**
 * @brief Initialize required clocks and enable MAA module
 *
 */
mxc_maa_ret_t MAA_Init(void);

/**
 * @brief Erase all MAA register RAM
 *
 */
mxc_maa_ret_t MAA_WipeRAM(void);


/**
 * @brief Load the selected MAA register
 *
 * @param regfile              Selects register to load
 * @param data                 Data to load into register
 * @param size                 Size of data to load
 * @param reverse              Reverse the data so that it will load properly on little endian machines
 *
 * @return MXC_E_MAA_ERR if any parameter out of range, MXC_E_MAA_BUSY if MAA registers are not accessible, otherwise MXC_E_MAA_OK
 */
mxc_maa_ret_t MAA_Load(mxc_maa_reg_select_t regfile, const uint8_t *data, unsigned int size, mxc_maa_endian_select_t flag);

/**
 * @brief Unload (copy from) the selected MAA register
 *
 * @param regfile              Selects register to unload
 * @param data                 Buffer in which to copy data
 * @param size                 Size of data to unload
 * @param reverse              Reverse the data so that it will unload properly on little endian machines
 *
 * @return MXC_E_MAA_ERR if any parameter out of range, MXC_E_MAA_BUSY if MAA registers are not accessible, otherwise MXC_E_MAA_OK
 */
mxc_maa_ret_t MAA_Unload(mxc_maa_reg_select_t regfile, uint8_t *data, unsigned int size, mxc_maa_endian_select_t flag);

/**
 * @brief Select and commence a MAA operation
 *
 * @param op                   Operation to perform
 * @param al                   Segment to use for operand A
 * @param bl                   Segment to use for operand B
 * @param rl                   Segment which will hold result R after operation is complete
 * @param tl                   Segment to use for temporary storage T
 *
 * @return MXC_E_MAA_ERR if any parameter out of range, MXC_E_MAA_BUSY if MAA already running, otherwise MXC_E_MAA_OK
 */
mxc_maa_ret_t MAA_Run(mxc_maa_operation_t op, \
		      mxc_maa_reg_select_t al, mxc_maa_reg_select_t bl, \
		      mxc_maa_reg_select_t rl, mxc_maa_reg_select_t tl);

/**
 * @brief Set the modular bit length
 *
 * @param len                  Modulus size in bits (ie. ln2(modulus))
 *
 * @return MXC_E_MAA_ERR if any parameter out of range, MXC_E_MAA_BUSY if MAA already running, otherwise MXC_E_MAA_OK
 */
mxc_maa_ret_t MAA_SetWordSize(unsigned int len);

/**
 * @}
 */

#ifdef __cplusplus
}
#endif

#endif
