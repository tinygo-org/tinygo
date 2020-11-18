/**
 * @file
 * @brief   Registers, Bit Masks and Bit Positions for the CRC Peripheral Module.
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
 * $Date: 2017-02-14 18:11:20 -0600 (Tue, 14 Feb 2017) $
 * $Revision: 26421 $
 *
 *************************************************************************** */

/* Define to prevent redundant inclusion */
#ifndef _MXC_CRC_REGS_H_
#define _MXC_CRC_REGS_H_
/* **** Includes **** */
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

///@cond
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
///@endcond

/* **** Definitions **** */

/** 
 * @ingroup  crc
 * @defgroup crc_registers Registers
 * @brief      Registers, Bit Masks and Bit Positions for the CRC Peripheral Module.
 * @{
 */ 

/**
 * Structure type for the CRC peripheral registers for reseeding and seeding the CRC16/32 
 */
typedef struct {
    __IO uint32_t reseed;                               /**< <b><tt>0x0000:</tt></b> CRC_RESEED Register */
    __IO uint32_t seed16;                               /**< <b><tt>0x0004:</tt></b> CRC_SEED16 Register */
    __IO uint32_t seed32;                               /**< <b><tt>0x0008:</tt></b> CRC_SEED32 Register */
} mxc_crc_regs_t;

/**
 * Structure type for the CRC Data Values.
 */
typedef struct {
    __IO uint32_t value16[512];                         /**< <b><tt>0x0000:</tt></b> CRC16_DATA Register */
    __IO uint32_t value32[512];                         /**< <b><tt>0x8000:</tt></b> CRC32_DATA Register */
} mxc_crc_data_regs_t;
/**@} end of group crc_registers */

/* Register offsets for module CRC. */
/**
 * @ingroup    crc_registers
 * @defgroup   CRC_Register_Offsets Register Offsets
 * @brief      CRC Peripheral Module Register Offsets from the CRC Base Peripheral Address. 
 * @{
 */
#define MXC_R_CRC_OFFS_RESEED           ((uint32_t)0x00000000UL)                    /**< Offset from CRC Base Address: <b><tt>0x0000</tt></b> */
#define MXC_R_CRC_OFFS_SEED16           ((uint32_t)0x00000004UL)                    /**< Offset from CRC Base Address: <b><tt>0x0004</tt></b> */
#define MXC_R_CRC_OFFS_SEED32           ((uint32_t)0x00000008UL)                    /**< Offset from CRC Base Address: <b><tt>0x0008</tt></b> */
#define MXC_R_CRC_DATA_OFFS_VALUE16     ((uint32_t)0x00000000UL)                    /**< Offset from CRC DATA Base Address: <b><tt>0x0000</tt></b> */
#define MXC_R_CRC_DATA_OFFS_VALUE32     ((uint32_t)0x00000800UL)                    /**< Offset from CRC DATA Base Address: <b><tt>0x8000</tt></b> */
/**@} end of group CRC_Register_offsets */

/**
 * @ingroup  crc_registers
 * @defgroup CRC_RESEED_Register CRC_RESEED
 * @brief    Field Positions and Bit Masks for the CRC_RESEED register
 * @{
 */
#define MXC_F_CRC_RESEED_CRC16_POS                          0                                                                   /**< CRC16 Position */  
#define MXC_F_CRC_RESEED_CRC16                              ((uint32_t)(0x00000001UL << MXC_F_CRC_RESEED_CRC16_POS))            /**< CRC16 Mask */  
#define MXC_F_CRC_RESEED_CRC32_POS                          1                                                                   /**< CRC32 Position */  
#define MXC_F_CRC_RESEED_CRC32                              ((uint32_t)(0x00000001UL << MXC_F_CRC_RESEED_CRC32_POS))            /**< CRC32 Mask */
#define MXC_F_CRC_RESEED_REV_ENDIAN16_POS                   4                                                                   /**< REV_ENDIAN16 Position */
#define MXC_F_CRC_RESEED_REV_ENDIAN16                       ((uint32_t)(0x00000001UL << MXC_F_CRC_RESEED_REV_ENDIAN16_POS))     /**< REV_ENDIAN16 Mask */
#define MXC_F_CRC_RESEED_REV_ENDIAN32_POS                   5                                                                   /**< REV_ENDIAN32 Position */
#define MXC_F_CRC_RESEED_REV_ENDIAN32                       ((uint32_t)(0x00000001UL << MXC_F_CRC_RESEED_REV_ENDIAN32_POS))     /**< REV_ENDIAN32 Mask */
#define MXC_F_CRC_RESEED_CCITT_MODE_POS                     8                                                                   /**< CCITT_MODE Position */
#define MXC_F_CRC_RESEED_CCITT_MODE                         ((uint32_t)(0x00000001UL << MXC_F_CRC_RESEED_CCITT_MODE_POS))       /**< CCITT_MODE Mask */
/**@} end of CRC_RESEED_Fields */

#ifdef __cplusplus
}
#endif

#endif   /* _MXC_CRC_REGS_H_ */
