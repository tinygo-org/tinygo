/**
 * @file
 * @brief   Registers, Bit Masks and Bit Positions for the AES Peripheral Module.
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
 * $Date: 2017-02-14 18:06:23 -0600 (Tue, 14 Feb 2017) $
 * $Revision: 26418 $
 *
  *************************************************************************** */

/* Define to prevent redundant inclusion */
#ifndef _MXC_AES_REGS_H_
#define _MXC_AES_REGS_H_

#ifdef __cplusplus
extern "C" {
#endif

/* **** Includes **** */
#include <stdint.h>

/// @cond
/*
    If types are not defined elsewhere (CMSIS) define them here
*/
#ifndef __IO
#define __IO volatile
#endif
#ifndef __I
#define __I  volatile const
#endif
#ifndef __O
#define __O  volatile
#endif
#ifndef __R
#define __R  volatile const
#endif
/// @endcond

/* **** Definitions **** */

/**
 * @ingroup     aes
 * @defgroup    aes_registers Registers
 * @brief       Registers, Bit Masks and Bit Positions for the AES Peripheral Module.
 * @{
 */
 
/**
 * Structure type to access the AES Registers.
 */
typedef struct {
    __IO uint32_t ctrl;                                 /**<  <b><tt>0x0000:</tt></b> AES_CTRL Register                                                               */
    __R  uint32_t rsv004;                               /**<  <b><tt>0x0004:</tt></b> RESERVED                                                                        */
    __IO uint32_t erase_all;                            /**<  <b><tt>0x0008:</tt></b> AES_ERASE_ALL Register - A write to this register will trigger AES Memory Erase */
} mxc_aes_regs_t;

/**
 * Structure type to access the AES Memory Registers.
 */
typedef struct {
    __IO uint32_t inp[4];                               /**<  <b><tt>0x0000-0x000C:</tt></b> AES Input (128 bits)                                                         */
    __IO uint32_t key[8];                               /**<  <b><tt>0x0010-0x002C:</tt></b> AES Symmetric Key (up to 256 bits)                                           */
    __IO uint32_t out[4];                               /**<  <b><tt>0x0030-0x003C:</tt></b> AES Output Data (128 bits)                                                   */
    __IO uint32_t expkey[8];                            /**<  <b><tt>0x0040-0x005C:</tt></b> AES Expanded Key Data (256 bits)                                             */
} mxc_aes_mem_regs_t;
/**@} end of group aes_registers */

 /**
  * @ingroup    aes_registers
  * @defgroup   AES_Register_Offsets Register Offsets
  * @brief      AES Register Offsets from the AES Base Peripheral Address. 
  * @{
  */
/**
 * AES Register offsets from the AES base peripheral address.
 */
#define MXC_R_AES_OFFS_CTRL                                 ((uint32_t)0x00000000UL)        /**<  Offset from the AES Base Peripheral Address: <b><tt>0x0000</tt></b> */
#define MXC_R_AES_OFFS_ERASE_ALL                            ((uint32_t)0x00000008UL)        /**<  Offset from the AES Base Peripheral Address: <b><tt>0x0008</tt></b> */
#define MXC_R_AES_MEM_OFFS_INP0                             ((uint32_t)0x00000000UL)        /**<  Offset from the AES Base Peripheral Address: <b><tt>0x0000</tt></b> */
#define MXC_R_AES_MEM_OFFS_INP1                             ((uint32_t)0x00000004UL)        /**<  Offset from the AES Base Peripheral Address: <b><tt>0x0004</tt></b> */
#define MXC_R_AES_MEM_OFFS_INP2                             ((uint32_t)0x00000008UL)        /**<  Offset from the AES Base Peripheral Address: <b><tt>0x0008</tt></b> */
#define MXC_R_AES_MEM_OFFS_INP3                             ((uint32_t)0x0000000CUL)        /**<  Offset from the AES Base Peripheral Address: <b><tt>0x000C</tt></b> */
#define MXC_R_AES_MEM_OFFS_KEY0                             ((uint32_t)0x00000010UL)        /**<  Offset from the AES Base Peripheral Address: <b><tt>0x0010</tt></b> */
#define MXC_R_AES_MEM_OFFS_KEY1                             ((uint32_t)0x00000014UL)        /**<  Offset from the AES Base Peripheral Address: <b><tt>0x0014</tt></b> */
#define MXC_R_AES_MEM_OFFS_KEY2                             ((uint32_t)0x00000018UL)        /**<  Offset from the AES Base Peripheral Address: <b><tt>0x0018</tt></b> */
#define MXC_R_AES_MEM_OFFS_KEY3                             ((uint32_t)0x0000001CUL)        /**<  Offset from the AES Base Peripheral Address: <b><tt>0x001C</tt></b> */
#define MXC_R_AES_MEM_OFFS_KEY4                             ((uint32_t)0x00000020UL)        /**<  Offset from the AES Base Peripheral Address: <b><tt>0x0020</tt></b> */
#define MXC_R_AES_MEM_OFFS_KEY5                             ((uint32_t)0x00000024UL)        /**<  Offset from the AES Base Peripheral Address: <b><tt>0x0024</tt></b> */
#define MXC_R_AES_MEM_OFFS_KEY6                             ((uint32_t)0x00000028UL)        /**<  Offset from the AES Base Peripheral Address: <b><tt>0x0028</tt></b> */
#define MXC_R_AES_MEM_OFFS_KEY7                             ((uint32_t)0x0000002CUL)        /**<  Offset from the AES Base Peripheral Address: <b><tt>0x002C</tt></b> */
#define MXC_R_AES_MEM_OFFS_OUT0                             ((uint32_t)0x00000030UL)        /**<  Offset from the AES Base Peripheral Address: <b><tt>0x0030</tt></b> */
#define MXC_R_AES_MEM_OFFS_OUT1                             ((uint32_t)0x00000034UL)        /**<  Offset from the AES Base Peripheral Address: <b><tt>0x0034</tt></b> */
#define MXC_R_AES_MEM_OFFS_OUT2                             ((uint32_t)0x00000038UL)        /**<  Offset from the AES Base Peripheral Address: <b><tt>0x0038</tt></b> */
#define MXC_R_AES_MEM_OFFS_OUT3                             ((uint32_t)0x0000003CUL)        /**<  Offset from the AES Base Peripheral Address: <b><tt>0x003C</tt></b> */
#define MXC_R_AES_MEM_OFFS_EXPKEY0                          ((uint32_t)0x00000040UL)        /**<  Offset from the AES Base Peripheral Address: <b><tt>0x0040</tt></b> */
#define MXC_R_AES_MEM_OFFS_EXPKEY1                          ((uint32_t)0x00000044UL)        /**<  Offset from the AES Base Peripheral Address: <b><tt>0x0044</tt></b> */
#define MXC_R_AES_MEM_OFFS_EXPKEY2                          ((uint32_t)0x00000048UL)        /**<  Offset from the AES Base Peripheral Address: <b><tt>0x0048</tt></b> */
#define MXC_R_AES_MEM_OFFS_EXPKEY3                          ((uint32_t)0x0000004CUL)        /**<  Offset from the AES Base Peripheral Address: <b><tt>0x004C</tt></b> */
#define MXC_R_AES_MEM_OFFS_EXPKEY4                          ((uint32_t)0x00000050UL)        /**<  Offset from the AES Base Peripheral Address: <b><tt>0x0050</tt></b> */
#define MXC_R_AES_MEM_OFFS_EXPKEY5                          ((uint32_t)0x00000054UL)        /**<  Offset from the AES Base Peripheral Address: <b><tt>0x0054</tt></b> */
#define MXC_R_AES_MEM_OFFS_EXPKEY6                          ((uint32_t)0x00000058UL)        /**<  Offset from the AES Base Peripheral Address: <b><tt>0x0058</tt></b> */
#define MXC_R_AES_MEM_OFFS_EXPKEY7                          ((uint32_t)0x0000005CUL)        /**<  Offset from the AES Base Peripheral Address: <b><tt>0x005C</tt></b> */
/**@} end of group AES_Register_Offsets */

/**
 * @ingroup  aes_registers
 * @defgroup AES_CTRL_Register AES_CTRL
 * @brief    Field Positions and Bit Masks for the AES_CTRL register
 * @{
 */
#define MXC_F_AES_CTRL_START_POS                            0                                                                   /**< AES_CTRL START Position        */
#define MXC_F_AES_CTRL_START                                ((uint32_t)(0x00000001UL << MXC_F_AES_CTRL_START_POS))              /**< AES_CTRL START Mask            */
#define MXC_F_AES_CTRL_CRYPT_MODE_POS                       1                                                                   /**< AES_CTRL CRYPT_MODE Position   */
#define MXC_F_AES_CTRL_CRYPT_MODE                           ((uint32_t)(0x00000001UL << MXC_F_AES_CTRL_CRYPT_MODE_POS))         /**< AES_CTRL CRYPT_MODE Mask       */
#define MXC_F_AES_CTRL_EXP_KEY_MODE_POS                     2                                                                   /**< AES_CTRL EXP_KEY_MODE Position */   
#define MXC_F_AES_CTRL_EXP_KEY_MODE                         ((uint32_t)(0x00000001UL << MXC_F_AES_CTRL_EXP_KEY_MODE_POS))       /**< AES_CTRL EXP_KEY_MODE  Mask    */     
#define MXC_F_AES_CTRL_KEY_SIZE_POS                         3                                                                   /**< AES_CTRL KEY_SIZE Position     */  
#define MXC_F_AES_CTRL_KEY_SIZE                             ((uint32_t)(0x00000003UL << MXC_F_AES_CTRL_KEY_SIZE_POS))           /**< AES_CTRL KEY_SIZE  Mask        */
#define MXC_F_AES_CTRL_INTEN_POS                            5                                                                   /**< AES_CTRL INTEN Position        */  
#define MXC_F_AES_CTRL_INTEN                                ((uint32_t)(0x00000001UL << MXC_F_AES_CTRL_INTEN_POS))              /**< AES_CTRL INTEN  Mask           */   
#define MXC_F_AES_CTRL_INTFL_POS                            6                                                                   /**< AES_CTRL INTFL Position        */      
#define MXC_F_AES_CTRL_INTFL                                ((uint32_t)(0x00000001UL << MXC_F_AES_CTRL_INTFL_POS))              /**< AES_CTRL INTFL Mask            */
#define MXC_F_AES_CTRL_LOAD_HW_KEY_POS                      7                                                                   /**< AES_CTRL LOAD_HW_KEY Position  */  
#define MXC_F_AES_CTRL_LOAD_HW_KEY                          ((uint32_t)(0x00000001UL << MXC_F_AES_CTRL_LOAD_HW_KEY_POS))        /**< AES_CTRL LOAD_HW_KEY Mask      */   
/**@} end of aes_registers group */   

/*
   Field values and shifted values for module AES.
*/
/**
 * @defgroup aes_ctrl_field_values AES_CTRL Field Values
 * @ingroup AES_CTRL_Register
 * @{
 */
#define MXC_V_AES_CTRL_ENCRYPT_MODE                                             ((uint32_t)(0x00000000UL))                                                          /**< CRYPT_MODE Field: Encryption Mode value */                                               
#define MXC_V_AES_CTRL_DECRYPT_MODE                                             ((uint32_t)(0x00000001UL))                                                          /**< CRYPT_MODE Field: Decryption Mode value */                                              

#define MXC_S_AES_CTRL_ENCRYPT_MODE                                             ((uint32_t)(MXC_V_AES_CTRL_ENCRYPT_MODE   << MXC_F_AES_CTRL_CRYPT_MODE_POS))        /**< CRYPT_MODE Field: Encryption Mode Shifted Value*/
#define MXC_S_AES_CTRL_DECRYPT_MODE                                             ((uint32_t)(MXC_V_AES_CTRL_DECRYPT_MODE   << MXC_F_AES_CTRL_CRYPT_MODE_POS))        /**< CRYPT_MODE Field: Decryption Mode Shifted Value*/

#define MXC_V_AES_CTRL_CALC_NEW_EXP_KEY                                         ((uint32_t)(0x00000000UL))                                                          /**< EXP_KEY_MODE Field: Calculate New Exp Key value */
#define MXC_V_AES_CTRL_USE_LAST_EXP_KEY                                         ((uint32_t)(0x00000001UL))                                                          /**< EXP_KEY_MODE Field: Use previous Exp Key value */

#define MXC_S_AES_CTRL_CALC_NEW_EXP_KEY                                         ((uint32_t)(MXC_V_AES_CTRL_CALC_NEW_EXP_KEY   << MXC_F_AES_CTRL_EXP_KEY_MODE_POS))  /**< EXP_KEY_MODE Field: Calculate New Exp Key Shifted Value*/
#define MXC_S_AES_CTRL_USE_LAST_EXP_KEY                                         ((uint32_t)(MXC_V_AES_CTRL_USE_LAST_EXP_KEY   << MXC_F_AES_CTRL_EXP_KEY_MODE_POS))  /**< EXP_KEY_MODE Field: Use previous Exp Key Shifted Value*/ 

#define MXC_V_AES_CTRL_KEY_SIZE_128                                             ((uint32_t)(0x00000000UL))                                                          /**< KEY_SIZE 128-bit setting value */
#define MXC_V_AES_CTRL_KEY_SIZE_192                                             ((uint32_t)(0x00000001UL))                                                          /**< KEY_SIZE 192-bit setting value */
#define MXC_V_AES_CTRL_KEY_SIZE_256                                             ((uint32_t)(0x00000002UL))                                                          /**< KEY_SIZE 256-bit setting value */

#define MXC_S_AES_CTRL_KEY_SIZE_128                                             ((uint32_t)(MXC_V_AES_CTRL_KEY_SIZE_128   << MXC_F_AES_CTRL_KEY_SIZE_POS))          /**< KEY_SIZE 128-bit Shifted Value */
#define MXC_S_AES_CTRL_KEY_SIZE_192                                             ((uint32_t)(MXC_V_AES_CTRL_KEY_SIZE_192   << MXC_F_AES_CTRL_KEY_SIZE_POS))          /**< KEY_SIZE 192-bit Shifted Value */
#define MXC_S_AES_CTRL_KEY_SIZE_256                                             ((uint32_t)(MXC_V_AES_CTRL_KEY_SIZE_256   << MXC_F_AES_CTRL_KEY_SIZE_POS))          /**< KEY_SIZE 256-bit Shifted Value */
/**@}*/
#ifdef __cplusplus
}
#endif

#endif   /* _MXC_AES_REGS_H_ */

