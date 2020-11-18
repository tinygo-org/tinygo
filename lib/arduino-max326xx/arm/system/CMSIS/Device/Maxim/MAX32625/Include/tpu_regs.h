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
 * $Date: 2016-03-11 12:50:27 -0600 (Fri, 11 Mar 2016) $
 * $Revision: 21840 $
 *
 ******************************************************************************/

#ifndef _MXC_TPU_REGS_H_
#define _MXC_TPU_REGS_H_

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
    __R  uint32_t rsv000[4];                            /*  0x0000-0x000C                                                                                */
    __IO uint32_t sks0;                                 /*  0x0010          TPU Secure Key Storage Register 0 (Cleared on Tamper Detect)                 */
    __IO uint32_t sks1;                                 /*  0x0014          TPU Secure Key Storage Register 1 (Cleared on Tamper Detect)                 */
    __IO uint32_t sks2;                                 /*  0x0018          TPU Secure Key Storage Register 2 (Cleared on Tamper Detect)                 */
    __IO uint32_t sks3;                                 /*  0x001C          TPU Secure Key Storage Register 3 (Cleared on Tamper Detect)                 */
} mxc_tpu_tsr_regs_t;


/*
   Register offsets for module TPU.
*/

#define MXC_R_TPU_TSR_OFFS_SKS0                             ((uint32_t)0x00000010UL)
#define MXC_R_TPU_TSR_OFFS_SKS1                             ((uint32_t)0x00000014UL)
#define MXC_R_TPU_TSR_OFFS_SKS2                             ((uint32_t)0x00000018UL)
#define MXC_R_TPU_TSR_OFFS_SKS3                             ((uint32_t)0x0000001CUL)



#ifdef __cplusplus
}
#endif

#endif   /* _MXC_TPU_REGS_H_ */

