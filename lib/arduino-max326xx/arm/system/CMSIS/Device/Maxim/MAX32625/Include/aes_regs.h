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

#ifndef _MXC_AES_REGS_H_
#define _MXC_AES_REGS_H_

#ifdef __cplusplus
extern "C" {
#endif

#include <stdint.h>

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


/*
   Typedefed structure(s) for module registers (per instance or section) with direct 32-bit
   access to each register in module.
*/

/*                                                          Offset          Register Description
                                                            =============   ============================================================================ */
typedef struct {
    __IO uint32_t ctrl;                                 /*  0x0000          AES Control and Status                                                       */
    __R  uint32_t rsv004;                               /*  0x0004                                                                                       */
    __IO uint32_t erase_all;                            /*  0x0008          Write to Trigger AES Memory Erase                                            */
} mxc_aes_regs_t;


/*                                                          Offset          Register Description
                                                            =============   ============================================================================ */
typedef struct {
    __IO uint32_t inp[4];                               /*  0x0000-0x000C   AES Input (128 bits)                                                         */
    __IO uint32_t key[8];                               /*  0x0010-0x002C   AES Symmetric Key (up to 256 bits)                                           */
    __IO uint32_t out[4];                               /*  0x0030-0x003C   AES Output Data (128 bits)                                                   */
    __IO uint32_t expkey[8];                            /*  0x0040-0x005C   AES Expanded Key Data (256 bits)                                             */
} mxc_aes_mem_regs_t;


/*
   Register offsets for module AES.
*/

#define MXC_R_AES_OFFS_CTRL                                 ((uint32_t)0x00000000UL)
#define MXC_R_AES_OFFS_ERASE_ALL                            ((uint32_t)0x00000008UL)
#define MXC_R_AES_MEM_OFFS_INP0                             ((uint32_t)0x00000000UL)
#define MXC_R_AES_MEM_OFFS_INP1                             ((uint32_t)0x00000004UL)
#define MXC_R_AES_MEM_OFFS_INP2                             ((uint32_t)0x00000008UL)
#define MXC_R_AES_MEM_OFFS_INP3                             ((uint32_t)0x0000000CUL)
#define MXC_R_AES_MEM_OFFS_KEY0                             ((uint32_t)0x00000010UL)
#define MXC_R_AES_MEM_OFFS_KEY1                             ((uint32_t)0x00000014UL)
#define MXC_R_AES_MEM_OFFS_KEY2                             ((uint32_t)0x00000018UL)
#define MXC_R_AES_MEM_OFFS_KEY3                             ((uint32_t)0x0000001CUL)
#define MXC_R_AES_MEM_OFFS_KEY4                             ((uint32_t)0x00000020UL)
#define MXC_R_AES_MEM_OFFS_KEY5                             ((uint32_t)0x00000024UL)
#define MXC_R_AES_MEM_OFFS_KEY6                             ((uint32_t)0x00000028UL)
#define MXC_R_AES_MEM_OFFS_KEY7                             ((uint32_t)0x0000002CUL)
#define MXC_R_AES_MEM_OFFS_OUT0                             ((uint32_t)0x00000030UL)
#define MXC_R_AES_MEM_OFFS_OUT1                             ((uint32_t)0x00000034UL)
#define MXC_R_AES_MEM_OFFS_OUT2                             ((uint32_t)0x00000038UL)
#define MXC_R_AES_MEM_OFFS_OUT3                             ((uint32_t)0x0000003CUL)
#define MXC_R_AES_MEM_OFFS_EXPKEY0                          ((uint32_t)0x00000040UL)
#define MXC_R_AES_MEM_OFFS_EXPKEY1                          ((uint32_t)0x00000044UL)
#define MXC_R_AES_MEM_OFFS_EXPKEY2                          ((uint32_t)0x00000048UL)
#define MXC_R_AES_MEM_OFFS_EXPKEY3                          ((uint32_t)0x0000004CUL)
#define MXC_R_AES_MEM_OFFS_EXPKEY4                          ((uint32_t)0x00000050UL)
#define MXC_R_AES_MEM_OFFS_EXPKEY5                          ((uint32_t)0x00000054UL)
#define MXC_R_AES_MEM_OFFS_EXPKEY6                          ((uint32_t)0x00000058UL)
#define MXC_R_AES_MEM_OFFS_EXPKEY7                          ((uint32_t)0x0000005CUL)


/*
   Field positions and masks for module AES.
*/

#define MXC_F_AES_CTRL_START_POS                            0
#define MXC_F_AES_CTRL_START                                ((uint32_t)(0x00000001UL << MXC_F_AES_CTRL_START_POS))
#define MXC_F_AES_CTRL_CRYPT_MODE_POS                       1
#define MXC_F_AES_CTRL_CRYPT_MODE                           ((uint32_t)(0x00000001UL << MXC_F_AES_CTRL_CRYPT_MODE_POS))
#define MXC_F_AES_CTRL_EXP_KEY_MODE_POS                     2
#define MXC_F_AES_CTRL_EXP_KEY_MODE                         ((uint32_t)(0x00000001UL << MXC_F_AES_CTRL_EXP_KEY_MODE_POS))
#define MXC_F_AES_CTRL_KEY_SIZE_POS                         3
#define MXC_F_AES_CTRL_KEY_SIZE                             ((uint32_t)(0x00000003UL << MXC_F_AES_CTRL_KEY_SIZE_POS))
#define MXC_F_AES_CTRL_INTEN_POS                            5
#define MXC_F_AES_CTRL_INTEN                                ((uint32_t)(0x00000001UL << MXC_F_AES_CTRL_INTEN_POS))
#define MXC_F_AES_CTRL_INTFL_POS                            6
#define MXC_F_AES_CTRL_INTFL                                ((uint32_t)(0x00000001UL << MXC_F_AES_CTRL_INTFL_POS))
#define MXC_F_AES_CTRL_LOAD_HW_KEY_POS                      7
#define MXC_F_AES_CTRL_LOAD_HW_KEY                          ((uint32_t)(0x00000001UL << MXC_F_AES_CTRL_LOAD_HW_KEY_POS))



/*
   Field values and shifted values for module AES.
*/

#define MXC_V_AES_CTRL_ENCRYPT_MODE                                             ((uint32_t)(0x00000000UL))
#define MXC_V_AES_CTRL_DECRYPT_MODE                                             ((uint32_t)(0x00000001UL))

#define MXC_S_AES_CTRL_ENCRYPT_MODE                                             ((uint32_t)(MXC_V_AES_CTRL_ENCRYPT_MODE   << MXC_F_AES_CTRL_CRYPT_MODE_POS))
#define MXC_S_AES_CTRL_DECRYPT_MODE                                             ((uint32_t)(MXC_V_AES_CTRL_DECRYPT_MODE   << MXC_F_AES_CTRL_CRYPT_MODE_POS))

#define MXC_V_AES_CTRL_CALC_NEW_EXP_KEY                                         ((uint32_t)(0x00000000UL))
#define MXC_V_AES_CTRL_USE_LAST_EXP_KEY                                         ((uint32_t)(0x00000001UL))

#define MXC_S_AES_CTRL_CALC_NEW_EXP_KEY                                         ((uint32_t)(MXC_V_AES_CTRL_CALC_NEW_EXP_KEY   << MXC_F_AES_CTRL_EXP_KEY_MODE_POS))
#define MXC_S_AES_CTRL_USE_LAST_EXP_KEY                                         ((uint32_t)(MXC_V_AES_CTRL_USE_LAST_EXP_KEY   << MXC_F_AES_CTRL_EXP_KEY_MODE_POS))

#define MXC_V_AES_CTRL_KEY_SIZE_128                                             ((uint32_t)(0x00000000UL))
#define MXC_V_AES_CTRL_KEY_SIZE_192                                             ((uint32_t)(0x00000001UL))
#define MXC_V_AES_CTRL_KEY_SIZE_256                                             ((uint32_t)(0x00000002UL))

#define MXC_S_AES_CTRL_KEY_SIZE_128                                             ((uint32_t)(MXC_V_AES_CTRL_KEY_SIZE_128   << MXC_F_AES_CTRL_KEY_SIZE_POS))
#define MXC_S_AES_CTRL_KEY_SIZE_192                                             ((uint32_t)(MXC_V_AES_CTRL_KEY_SIZE_192   << MXC_F_AES_CTRL_KEY_SIZE_POS))
#define MXC_S_AES_CTRL_KEY_SIZE_256                                             ((uint32_t)(MXC_V_AES_CTRL_KEY_SIZE_256   << MXC_F_AES_CTRL_KEY_SIZE_POS))



#ifdef __cplusplus
}
#endif

#endif   /* _MXC_AES_REGS_H_ */

