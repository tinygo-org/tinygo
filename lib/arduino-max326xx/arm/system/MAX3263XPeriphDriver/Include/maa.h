/**
 * @file
 * @brief      Registers, Bit Masks and Bit Positions for the Modular Math
 *             Accelerator (MAA) module.
 */
/* *****************************************************************************
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
 * $Date: 2017-02-16 12:05:53 -0600 (Thu, 16 Feb 2017) $
 * $Revision: 26466 $
 *
 **************************************************************************** */

/* Define to prevent redundant inclusion */
#ifndef _MAA_H
#define _MAA_H

/* **** Includes **** */
#include <stdint.h>
/**
 * @defgroup maa MAA
 * @ingroup periphlibs
 * @details The API only supports synchronous operations due to the sensitive 
 *          nature of the input and output data. 
 * @{     
 */
#include "maa_regs.h"

#ifdef __cplusplus
extern "C" {
#endif



/* **** Definitions **** */

/**
 * Definition for the maximum MAA register size on this device in bytes, 128.
 */
#define MXC_MAA_REG_SIZE      0x80
/**
 * Definition for the maximum MAA register size on this device in bits, 1024.
 */
#define MXC_MAA_REG_SIZE_BITS (MXC_MAA_REG_SIZE << 3)

/**
 * Definition for a sub-register ("half size"), allowing 2x more operands in MAA at a time when \f$ MAWS \leq \fraq{MAX\_SIZE}{2} \f$
 */
#define MXC_MAA_HALF_SIZE (MXC_MAA_REG_SIZE/2)

/** Flags for MAA_Load() and MAA_Unload() */
#define MXC_MAA_F_MEM_VERBATIM                        0 
/** Flags for MAA_Load() and MAA_Unload() */
#define MXC_MAA_F_MEM_REVERSE                         1 

/**
 * Enumeration type for Segment and Sub-segment selection
 */
typedef enum {
    /* Warning: Do not change the assigned numbers/ordering without associated changes to UMAA_REGFILE_TO_ADDR(x) */
    /* Register names when MAWS > 512 */ 
    MXC_E_REG_0 = 0,      /**< Register MXC_E_REG_0:  If MAA_MAWS > 512 use this register name. */
    MXC_E_REG_1 = 2,      /**< Register MXC_E_REG_1:  If MAA_MAWS > 512 use this register name. */
    MXC_E_REG_2 = 4,      /**< Register MXC_E_REG_2:  If MAA_MAWS > 512 use this register name. */
    MXC_E_REG_3 = 6,      /**< Register MXC_E_REG_3:  If MAA_MAWS > 512 use this register name. */
    MXC_E_REG_4 = 8,      /**< Register MXC_E_REG_4:  If MAA_MAWS > 512 use this register name. */
    MXC_E_REG_5 = 10,     /**< Register MXC_E_REG_5:  If MAA_MAWS > 512 use this register name. */
    /* Register names when MAWS < 512 */ 
    MXC_E_REG_00 = 0,     /**< Register MXC_E_REG_00: If MAA_MAWS < 512 this is the register name. */
    MXC_E_REG_01 = 1,     /**< Register MXC_E_REG_01: If MAA_MAWS < 512 this is the register name. */
    MXC_E_REG_10 = 2,     /**< Register MXC_E_REG_10: If MAA_MAWS < 512 this is the register name. */
    MXC_E_REG_11 = 3,     /**< Register MXC_E_REG_11: If MAA_MAWS < 512 this is the register name. */
    MXC_E_REG_20 = 4,     /**< Register MXC_E_REG_20: If MAA_MAWS < 512 this is the register name. */
    MXC_E_REG_21 = 5,     /**< Register MXC_E_REG_21: If MAA_MAWS < 512 this is the register name. */
    MXC_E_REG_30 = 6,     /**< Register MXC_E_REG_30: If MAA_MAWS < 512 this is the register name. */
    MXC_E_REG_31 = 7,     /**< Register MXC_E_REG_31: If MAA_MAWS < 512 this is the register name. */
    MXC_E_REG_40 = 8,     /**< Register MXC_E_REG_40: If MAA_MAWS < 512 this is the register name. */
    MXC_E_REG_41 = 9,     /**< Register MXC_E_REG_41: If MAA_MAWS < 512 this is the register name. */
    MXC_E_REG_50 = 10,    /**< Register MXC_E_REG_50: If MAA_MAWS < 512 this is the register name. */ 
    MXC_E_REG_51 = 11     /**< Register MXC_E_REG_51: If MAA_MAWS < 512 this is the register name. */
} mxc_maa_reg_select_t;

/**
 * Enumeration type for MAA operation selection.
 */
typedef enum {
    MXC_E_MAA_EXP = 0,              /**< Exponentiate */
    MXC_E_MAA_SQR = 1,              /**< Square */
    MXC_E_MAA_MUL = 2,              /**< Multiply */
    MXC_E_MAA_SQRMUL = 3,           /**< Square followed by Multiply */
    MXC_E_MAA_ADD = 4,              /**< Addition */
    MXC_E_MAA_SUB = 5               /**< Subtraction */
} mxc_maa_operation_t;

/**
 * Enumeration type to set special flags for loading & unloading data.
 */
typedef enum {
  
  MXC_E_MAA_VERBATIM = 0,   /**< Copy bytes without reversal and right-justification */
  MXC_E_MAA_REVERSE         /**< Reverse bytes and right-justify (bytes are loaded at the highest address, then descending) */
} mxc_maa_endian_select_t;

/**
 * Enumeration type for MAA module specific return codes.
 */
typedef enum {
  MXC_E_MAA_ERR = -1,     /**< Error */
  MXC_E_MAA_OK = 0,       /**< No Error */
  MXC_E_MAA_BUSY          /**< MAA engine busy, try again later */
} mxc_maa_ret_t;

/**
 * @brief      Initialize the required clocks and enable the MAA peripheral
 *             module.
 * @retval     #MXC_E_MAA_ERR on error. 
 * @retval     #MXC_E_MAA_BUSY if the MAA is busy. 
 * @retval     #MXC_E_MAA_OK if the MAA is initialized successfully. 
 */
mxc_maa_ret_t MAA_Init(void);

/**
 * @brief      Erase all MAA register RAM
 * @retval     #MXC_E_MAA_ERR on error. 
 * @retval     #MXC_E_MAA_BUSY if the MAA is busy. 
 * @retval     #MXC_E_MAA_OK if the MAA is initialized successfully. 
 */
mxc_maa_ret_t MAA_WipeRAM(void);


/**
 * @brief      Load the selected MAA register.
 *
 * @param      regfile  Selects the register to load.
 * @param      data     Pointer to a data buffer to load into the register.
 * @param      size     Size of the data to load.
 * @param      flag     Reverse the data so that it will unload properly on
 *                      little endian machines, see #mxc_maa_endian_select_t.
 *
 * @return     #MXC_E_MAA_ERR if any parameter out of range.
 * @return     #MXC_E_MAA_BUSY if MAA registers are not currently accessible. 
 * @return     #MXC_E_MAA_OK if the selected register is loaded correctly.
 */
mxc_maa_ret_t MAA_Load(mxc_maa_reg_select_t regfile, const uint8_t *data, unsigned int size, mxc_maa_endian_select_t flag);

/**
 * @brief      Unload (copy from) the selected MAA register
 *
 * @param      regfile  Selects the register to unload.
 * @param      data     Pointer to a buffer to store the unloaded data.
 * @param      size     Maximum size of the data to unload.
 * @param      flag     Reverse the data so that it will unload properly on
 *                      little endian machines, see #mxc_maa_endian_select_t.
 * @return     #MXC_E_MAA_ERR if any parameter out of range.
 * @return     #MXC_E_MAA_BUSY if MAA registers are not currently accessible.
 * @return     #MXC_E_MAA_OK if the requested register data is copied correctly
 *             to @p data.
 */
mxc_maa_ret_t MAA_Unload(mxc_maa_reg_select_t regfile, uint8_t *data, unsigned int size, mxc_maa_endian_select_t flag);

/**
 * @brief      Execute an MAA operation specified.
 *
 * @param      op    Operation to perform, see #mxc_maa_operation_t.
 * @param      al    Segment to use for operand A, see #mxc_maa_reg_select_t.
 * @param      bl    Segment to use for operand B.
 * @param      rl    Segment which will hold result R after the operation is
 *                   complete.
 * @param      tl    Segment to use for temporary storage T.
 * 
 * @return     #MXC_E_MAA_ERR if any parameter out of range.
 * @return     #MXC_E_MAA_BUSY if MAA registers are not currently accessible.
 * @return     #MXC_E_MAA_OK if the operation completed. 
 */
mxc_maa_ret_t MAA_Run(mxc_maa_operation_t op, \
          mxc_maa_reg_select_t al, mxc_maa_reg_select_t bl, \
          mxc_maa_reg_select_t rl, mxc_maa_reg_select_t tl);

/**
 * @brief      Set the bit length of the modulus.
 *
 * @param      len   Modulus size in bits (ie. \f$ ln_2(modulus) \f$ )
 *
 * @return     #MXC_E_MAA_ERR if any parameter out of range.
 * @return     #MXC_E_MAA_BUSY if MAA registers are not currently accessible.
 * @return     #MXC_E_MAA_OK if the length is set as requested. 
 */
mxc_maa_ret_t MAA_SetWordSize(unsigned int len);

/**@} end of group maa*/

#ifdef __cplusplus
}
#endif

#endif
