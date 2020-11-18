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
 * $Date: 2015-11-02 13:19:39 -0600 (Mon, 02 Nov 2015) $
 * $Revision: 19838 $
 *
 ******************************************************************************/

#ifndef _MXC_PRNG_REGS_H_
#define _MXC_PRNG_REGS_H_

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
    __IO uint32_t user_entropy;								/*  0x0000          PRNG User Entropy and Status                                                 */
    __IO uint32_t rnd_num;									/*  0x0004          PRNG Seed Output                                                             */
} mxc_prng_regs_t;

/*
   Register offsets for module PRNG.
*/

#define MXC_R_PRNG_OFFS_USER_ENTROPY						((uint32_t)0x00000000UL)
#define MXC_R_PRNG_OFFS_RND_NUM								((uint32_t)0x00000004UL)

/*
   Field positions and masks for module PRNG.
*/

#define MXC_F_PRNG_USER_ENTROPY_VALUE_POS               0
#define MXC_F_PRNG_USER_ENTROPY_VALUE                   ((uint32_t)(0x000000FFUL << MXC_F_PRNG_USER_ENTROPY_VALUE_POS))
#define MXC_F_PRNG_USER_ENTROPY_RND_NUM_READY_POS       8
#define MXC_F_PRNG_USER_ENTROPY_RND_NUM_READY           ((uint32_t)(0x00000001UL << MXC_F_PRNG_USER_ENTROPY_RND_NUM_READY_POS))



#ifdef __cplusplus
}
#endif

#endif   /* _MXC_PRNG_REGS_H_ */

