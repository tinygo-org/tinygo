/**
 * @file
 * @brief   registers, bit masks and bit positions for the Flash
 *          Controller (FLC) peripheral module.
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
 * trademarks, Maskwork rights, or any other form of intellectual
 * property whatsoever. Maxim Integrated Products, Inc. retains all
 * ownership rights.
 *
 * $Date: 2017-02-14 18:12:18 -0600 (Tue, 14 Feb 2017) $
 * $Revision: 26422 $
 *
 *************************************************************************** */

/* Define to prevent redundant inclusion */
#ifndef _MXC_FLC_REGS_H_
#define _MXC_FLC_REGS_H_

/* **** Includes **** */
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

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
/**
 * @ingroup     flc
 * @defgroup    flc_registers Registers
 * @brief       Register interface definitions
 * @{
 */  
/* **** Definitions **** */
/**
 * @ingroup     flc_registers
 * @defgroup    flc_special_codes   Flash Controller Codes/Keys.
 * @brief       Required values to pass to the flash controller to perform restricted
 *              operations.
 * @{
 */
#define MXC_V_FLC_ERASE_CODE_PAGE_ERASE   ((uint8_t)0x55)           /**< Page Erase Code required to perform a page erase operation */
#define MXC_V_FLC_ERASE_CODE_MASS_ERASE   ((uint8_t)0xAA)           /**< Mass Erase Code required to perform a page erase operation */
#define MXC_V_FLC_FLSH_UNLOCK_KEY         ((uint8_t)0x2)            /**< Unlock Code required to unlock the flash for erase and write functions */
/**@} end of flc_special_codes */


/*
   Typedefed structure(s) for module registers (per instance or section) with direct 32-bit
   access to each register in module.
*/

/**
 * @ingroup    flc_registers
 * @brief      Structure type to access the Flash Controller registers with
 *             direct 32-bit access to each.
 */
typedef struct {
    __IO uint32_t faddr;                   /**<  <b><tt>0x0000:</tt></b> FLC_FADDR Register - Flash Operation Address                                          */
    __IO uint32_t fckdiv;                  /**<  <b><tt>0x0004:</tt></b> FLC_FCKDIV Register - Flash Clock Pulse Divisor                                       */
    __IO uint32_t ctrl;                    /**<  <b><tt>0x0008:</tt></b> FLC_CTRL Register - Flash Control Register                                            */
    __R  uint32_t rsv00C[6];               /**<  <b><tt>0x000C-0x0020:</tt></b> RESERVED \warning Do Not Modify Reserved Locations!                                   */
    __IO uint32_t intr;                    /**<  <b><tt>0x0024:</tt></b> FLC_INTR Register - Flash Controller Interrupt Flags and Enable/Disable 0             */
    __R  uint32_t rsv028[2];               /**<  <b><tt>0x0028-0x002C:</tt></b> RESERVED                                                                              */
    __IO uint32_t fdata;                   /**<  <b><tt>0x0030:</tt></b> FLC_FDATA Register - Flash Operation Data Register                                    */
    __R  uint32_t rsv034[7];               /**<  <b><tt>0x0034-0x004C:</tt></b> RESERVED \warning Do Not Modify Reserved Locations!                                   */
    __IO uint32_t perform;                 /**<  <b><tt>0x0050:</tt></b> FLC_PERFORM Register - Flash Performance Settings                                     */
    __IO uint32_t tacc;                    /**<  <b><tt>0x0054:</tt></b> FLC_TACC Register - Flash Read Cycle Config                                           */
    __IO uint32_t tprog;                   /**<  <b><tt>0x0058:</tt></b> FLC_TPROG Register - Flash Write Cycle Config                                         */
    __R  uint32_t rsv05C[9];               /**<  <b><tt>0x005C-0x007C:</tt></b> RESERVED \warning Do Not Modify Reserved Locations!                                   */
    __IO uint32_t status;                  /**<  <b><tt>0x0080:</tt></b> FLC_STATUS Register - Security Status Flags                                           */
    __R  uint32_t rsv084;                  /**<  <b><tt>0x0084:</tt></b> RESERVED \warning Do Not Modify Reserved Locations!                                   */
    __IO uint32_t security;                /**<  <b><tt>0x0088:</tt></b> FLC_SECURITY Register - Flash Controller Security Settings                            */
    __R  uint32_t rsv08C[4];               /**<  <b><tt>0x008C-0x0098:</tt></b> RESERVED \warning Do Not Modify Reserved Locations!                                   */
    __IO uint32_t bypass;                  /**<  <b><tt>0x009C:</tt></b> FLC_BYPASS Register - Status Flags for DSB Operations                                 */
    __R  uint32_t rsv0A0[24];              /**<  <b><tt>0x00A0-0x00FC:</tt></b> RESERVED \warning Do Not Modify Reserved Locations!                                   */
    __IO uint32_t user_option;             /**<  <b><tt>0x0100:</tt></b> FLC_USER_OPTION Register - Used to set DSB Access code and Auto-Lock in info block    */
    __R  uint32_t rsv104[15];              /**<  <b><tt>0x0104-0x013C:</tt></b> RESERVED \warning Do Not Modify Reserved Locations!                                   */
    __IO uint32_t ctrl2;                   /**<  <b><tt>0x0140:</tt></b> FLC_CTRL2 Register - Flash Control Register 2                                         */
    __IO uint32_t intfl1;                  /**<  <b><tt>0x0144:</tt></b> FLC_INTFL1 Register - Interrupt Flags Register 1                                      */
    __IO uint32_t inten1;                  /**<  <b><tt>0x0148:</tt></b> FLC_INTEN1 Register - Interrupt Enable/Disable Register 1                             */
    __R  uint32_t rsv14C[9];               /**<  <b><tt>0x014C-0x016C:</tt></b> RESERVED \warning Do Not Modify Reserved Locations!                                   */
    __IO uint32_t bl_ctrl;                 /**<  <b><tt>0x0170:</tt></b> FLC_BL_CTRL Register - Bootloader Control Register                                    */
    __IO uint32_t twk;                     /**<  <b><tt>0x0174:</tt></b> FLC_TWK Register - PDM33 Register                                                     */
    __R  uint32_t rsv178;                  /**<  <b><tt>0x0178:</tt></b> RESERVED \warning Do Not Modify Reserved Locations!                                   */
    __IO uint32_t slm;                     /**<  <b><tt>0x017C:</tt></b> FLC_SLM Register - Sleep Mode Register                                                */
    __R  uint32_t rsv180[32];              /**<  <b><tt>0x0180-0x01FC:</tt></b> RESERVED \warning Do Not Modify Reserved Locations!                                   */
    __IO uint32_t disable_xr0;             /**<  <b><tt>0x0200:</tt></b> FLC_DISABLE_XR0 Register - Disable Flash Page Exec/Read Register 0                    */
    __IO uint32_t disable_xr1;             /**<  <b><tt>0x0204:</tt></b> FLC_DISABLE_XR1 Register - Disable Flash Page Exec/Read Register 1                    */
    __IO uint32_t disable_xr2;             /**<  <b><tt>0x0208:</tt></b> FLC_DISABLE_XR2 Register - Disable Flash Page Exec/Read Register 2                    */
    __IO uint32_t disable_xr3;             /**<  <b><tt>0x020C:</tt></b> FLC_DISABLE_XR3 Register - Disable Flash Page Exec/Read Register 3                    */
    __IO uint32_t disable_xr4;             /**<  <b><tt>0x0210:</tt></b> FLC_DISABLE_XR4 Register - Disable Flash Page Exec/Read Register 4                    */
    __IO uint32_t disable_xr5;             /**<  <b><tt>0x0214:</tt></b> FLC_DISABLE_XR5 Register - Disable Flash Page Exec/Read Register 5                    */
    __IO uint32_t disable_xr6;             /**<  <b><tt>0x0218:</tt></b> FLC_DISABLE_XR6 Register - Disable Flash Page Exec/Read Register 6                    */
    __IO uint32_t disable_xr7;             /**<  <b><tt>0x021C:</tt></b> FLC_DISABLE_XR7 Register - Disable Flash Page Exec/Read Register 7                    */
    __R  uint32_t rsv220[56];              /**<  <b><tt>0x0220-0x02FC:</tt></b> RESERVED \warning Do Not Modify Reserved Locations!                                   */
    __IO uint32_t disable_we0;             /**<  <b><tt>0x0300:</tt></b> FLC_DISABLE_WE0 Register - Disable Flash Page Write/Erase Register 0                  */
    __IO uint32_t disable_we1;             /**<  <b><tt>0x0304:</tt></b> FLC_DISABLE_WE1 Register - Disable Flash Page Write/Erase Register 1                  */
    __IO uint32_t disable_we2;             /**<  <b><tt>0x0308:</tt></b> FLC_DISABLE_WE2 Register - Disable Flash Page Write/Erase Register 2                  */
    __IO uint32_t disable_we3;             /**<  <b><tt>0x030C:</tt></b> FLC_DISABLE_WE3 Register - Disable Flash Page Write/Erase Register 3                  */
    __IO uint32_t disable_we4;             /**<  <b><tt>0x0310:</tt></b> FLC_DISABLE_WE4 Register - Disable Flash Page Write/Erase Register 4                  */
    __IO uint32_t disable_we5;             /**<  <b><tt>0x0314:</tt></b> FLC_DISABLE_WE5 Register - Disable Flash Page Write/Erase Register 5                  */
    __IO uint32_t disable_we6;             /**<  <b><tt>0x0318:</tt></b> FLC_DISABLE_WE6 Register - Disable Flash Page Write/Erase Register 6                  */
    __IO uint32_t disable_we7;             /**<  <b><tt>0x031C:</tt></b> FLC_DISABLE_WE7 Register - Disable Flash Page Write/Erase Register 7                  */
} mxc_flc_regs_t;
/*
   Register offsets for module FLC.
*/

/**
 * @ingroup    flc_registers
 * @defgroup   FLC_Register_Offsets Register Offsets
 * @brief      Flash Controller Register Offsets from the FLC Base Peripheral Address. 
 * @{
 */
#define MXC_R_FLC_OFFS_FADDR                                ((uint32_t)0x00000000UL)            /**<  Offset from FLC Base Address: <b><tt>0x0000</tt></b> */
#define MXC_R_FLC_OFFS_FCKDIV                               ((uint32_t)0x00000004UL)            /**<  Offset from FLC Base Address: <b><tt>0x0004</tt></b> */
#define MXC_R_FLC_OFFS_CTRL                                 ((uint32_t)0x00000008UL)            /**<  Offset from FLC Base Address: <b><tt>0x0008</tt></b> */
#define MXC_R_FLC_OFFS_INTR                                 ((uint32_t)0x00000024UL)            /**<  Offset from FLC Base Address: <b><tt>0x0024</tt></b> */
#define MXC_R_FLC_OFFS_FDATA                                ((uint32_t)0x00000030UL)            /**<  Offset from FLC Base Address: <b><tt>0x0030</tt></b> */
#define MXC_R_FLC_OFFS_PERFORM                              ((uint32_t)0x00000050UL)            /**<  Offset from FLC Base Address: <b><tt>0x0050</tt></b> */
#define MXC_R_FLC_OFFS_TACC                                 ((uint32_t)0x00000054UL)            /**<  Offset from FLC Base Address: <b><tt>0x0054</tt></b> */
#define MXC_R_FLC_OFFS_TPROG                                ((uint32_t)0x00000058UL)            /**<  Offset from FLC Base Address: <b><tt>0x0058</tt></b> */
#define MXC_R_FLC_OFFS_STATUS                               ((uint32_t)0x00000080UL)            /**<  Offset from FLC Base Address: <b><tt>0x0080</tt></b> */
#define MXC_R_FLC_OFFS_SECURITY                             ((uint32_t)0x00000088UL)            /**<  Offset from FLC Base Address: <b><tt>0x0088</tt></b> */
#define MXC_R_FLC_OFFS_BYPASS                               ((uint32_t)0x0000009CUL)            /**<  Offset from FLC Base Address: <b><tt>0x009C</tt></b> */
#define MXC_R_FLC_OFFS_USER_OPTION                          ((uint32_t)0x00000100UL)            /**<  Offset from FLC Base Address: <b><tt>0x0100</tt></b> */
#define MXC_R_FLC_OFFS_CTRL2                                ((uint32_t)0x00000140UL)            /**<  Offset from FLC Base Address: <b><tt>0x0140</tt></b> */
#define MXC_R_FLC_OFFS_INTFL1                               ((uint32_t)0x00000144UL)            /**<  Offset from FLC Base Address: <b><tt>0x0144</tt></b> */
#define MXC_R_FLC_OFFS_INTEN1                               ((uint32_t)0x00000148UL)            /**<  Offset from FLC Base Address: <b><tt>0x0148</tt></b> */
#define MXC_R_FLC_OFFS_BL_CTRL                              ((uint32_t)0x00000170UL)            /**<  Offset from FLC Base Address: <b><tt>0x0170</tt></b> */
#define MXC_R_FLC_OFFS_TWK                                  ((uint32_t)0x00000174UL)            /**<  Offset from FLC Base Address: <b><tt>0x0174</tt></b> */
#define MXC_R_FLC_OFFS_SLM                                  ((uint32_t)0x0000017CUL)            /**<  Offset from FLC Base Address: <b><tt>0x017C</tt></b> */
#define MXC_R_FLC_OFFS_DISABLE_XR0                          ((uint32_t)0x00000200UL)            /**<  Offset from FLC Base Address: <b><tt>0x0200</tt></b> */
#define MXC_R_FLC_OFFS_DISABLE_XR1                          ((uint32_t)0x00000204UL)            /**<  Offset from FLC Base Address: <b><tt>0x0204</tt></b> */
#define MXC_R_FLC_OFFS_DISABLE_XR2                          ((uint32_t)0x00000208UL)            /**<  Offset from FLC Base Address: <b><tt>0x0208</tt></b> */
#define MXC_R_FLC_OFFS_DISABLE_XR3                          ((uint32_t)0x0000020CUL)            /**<  Offset from FLC Base Address: <b><tt>0x020C</tt></b> */
#define MXC_R_FLC_OFFS_DISABLE_XR4                          ((uint32_t)0x00000210UL)            /**<  Offset from FLC Base Address: <b><tt>0x0210</tt></b> */
#define MXC_R_FLC_OFFS_DISABLE_XR5                          ((uint32_t)0x00000214UL)            /**<  Offset from FLC Base Address: <b><tt>0x0214</tt></b> */
#define MXC_R_FLC_OFFS_DISABLE_XR6                          ((uint32_t)0x00000218UL)            /**<  Offset from FLC Base Address: <b><tt>0x0218</tt></b> */
#define MXC_R_FLC_OFFS_DISABLE_XR7                          ((uint32_t)0x0000021CUL)            /**<  Offset from FLC Base Address: <b><tt>0x021C</tt></b> */
#define MXC_R_FLC_OFFS_DISABLE_WE0                          ((uint32_t)0x00000300UL)            /**<  Offset from FLC Base Address: <b><tt>0x0300</tt></b> */
#define MXC_R_FLC_OFFS_DISABLE_WE1                          ((uint32_t)0x00000304UL)            /**<  Offset from FLC Base Address: <b><tt>0x0304</tt></b> */
#define MXC_R_FLC_OFFS_DISABLE_WE2                          ((uint32_t)0x00000308UL)            /**<  Offset from FLC Base Address: <b><tt>0x0308</tt></b> */
#define MXC_R_FLC_OFFS_DISABLE_WE3                          ((uint32_t)0x0000030CUL)            /**<  Offset from FLC Base Address: <b><tt>0x030C</tt></b> */
#define MXC_R_FLC_OFFS_DISABLE_WE4                          ((uint32_t)0x00000310UL)            /**<  Offset from FLC Base Address: <b><tt>0x0310</tt></b> */
#define MXC_R_FLC_OFFS_DISABLE_WE5                          ((uint32_t)0x00000314UL)            /**<  Offset from FLC Base Address: <b><tt>0x0314</tt></b> */
#define MXC_R_FLC_OFFS_DISABLE_WE6                          ((uint32_t)0x00000318UL)            /**<  Offset from FLC Base Address: <b><tt>0x0318</tt></b> */
#define MXC_R_FLC_OFFS_DISABLE_WE7                          ((uint32_t)0x0000031CUL)            /**<  Offset from FLC Base Address: <b><tt>0x031C</tt></b> */
/**@} end of group FLC_Register_Offsets */   

/**
 * @ingroup    flc_registers
 * @defgroup   FLC_FADDR_Register FLC_FADDR
 * @brief      Field Positions and Bit Masks for the FLC_FADDR register. 
 * @{
 */
#define MXC_F_FLC_FADDR_FADDR_POS                           0                                                                                       /**< FADDR Position                 */
#define MXC_F_FLC_FADDR_FADDR                               ((uint32_t)(0x003FFFFFUL << MXC_F_FLC_FADDR_FADDR_POS))                                 /**< FADDR Mask                     */
/**@} end of group FLC_FADDR */
/**
 * @ingroup    flc_registers
 * @defgroup   FLC_FCKDIV_Register FLC_FCKDIV
 * @brief      Field Positions and Bit Masks for the FLC_FCKDIV register. 
 * @{
 */
#define MXC_F_FLC_FCKDIV_FCKDIV_POS                                                                                                                 /**< FCKDIV Position                */
#define MXC_F_FLC_FCKDIV_FCKDIV                             ((uint32_t)(0x0000007FUL << MXC_F_FLC_FCKDIV_FCKDIV_POS))                               /**< FCKDIV Mask                    */
#define MXC_F_FLC_FCKDIV_AUTO_FCKDIV_RESULT_POS             16                                                                                      /**< AUTO_FCKDIV_RESULT Position    */
#define MXC_F_FLC_FCKDIV_AUTO_FCKDIV_RESULT                 ((uint32_t)(0x0000FFFFUL << MXC_F_FLC_FCKDIV_AUTO_FCKDIV_RESULT_POS))                   /**< AUTO_FCKDIV_RESULT Mask        */
/**@} end of group FLC_FCKDIV */
/**
 * @ingroup    flc_registers
 * @defgroup   FLC_CTRL_Register FLC_CTRL
 * @brief      Field Positions and Bit Masks for the FLC_CTRL register. 
 * @{
 */
#define MXC_F_FLC_CTRL_WRITE_POS                            0                                                                                       /**< WRITE Position                     */
#define MXC_F_FLC_CTRL_WRITE                                ((uint32_t)(0x00000001UL << MXC_F_FLC_CTRL_WRITE_POS))                                  /**< WRITE Mask                         */ 
#define MXC_F_FLC_CTRL_MASS_ERASE_POS                       1                                                                                       /**< MASS_ERASE Position                */
#define MXC_F_FLC_CTRL_MASS_ERASE                           ((uint32_t)(0x00000001UL << MXC_F_FLC_CTRL_MASS_ERASE_POS))                             /**< MASS_ERASE Mask                    */ 
#define MXC_F_FLC_CTRL_PAGE_ERASE_POS                       2                                                                                       /**< PAGE_ERASE Position                */
#define MXC_F_FLC_CTRL_PAGE_ERASE                           ((uint32_t)(0x00000001UL << MXC_F_FLC_CTRL_PAGE_ERASE_POS))                             /**< PAGE_ERASE Mask                    */ 
#define MXC_F_FLC_CTRL_ERASE_CODE_POS                       8                                                                                       /**< ERASE_CODE Position                */
#define MXC_F_FLC_CTRL_ERASE_CODE                           ((uint32_t)(0x000000FFUL << MXC_F_FLC_CTRL_ERASE_CODE_POS))                             /**< ERASE_CODE Mask                    */ 
#define MXC_F_FLC_CTRL_INFO_BLOCK_UNLOCK_POS                16                                                                                      /**< INFO_BLOCK_UNLOCK Position         */
#define MXC_F_FLC_CTRL_INFO_BLOCK_UNLOCK                    ((uint32_t)(0x00000001UL << MXC_F_FLC_CTRL_INFO_BLOCK_UNLOCK_POS))                      /**< INFO_BLOCK_UNLOCK Mask             */ 
#define MXC_F_FLC_CTRL_WRITE_ENABLE_POS                     17                                                                                      /**< WRITE_ENABLE Position              */
#define MXC_F_FLC_CTRL_WRITE_ENABLE                         ((uint32_t)(0x00000001UL << MXC_F_FLC_CTRL_WRITE_ENABLE_POS))                           /**< WRITE_ENABLE Mask                  */ 
#define MXC_F_FLC_CTRL_PENDING_POS                          24                                                                                      /**< PENDING Position                   */
#define MXC_F_FLC_CTRL_PENDING                              ((uint32_t)(0x00000001UL << MXC_F_FLC_CTRL_PENDING_POS))                                /**< PENDING Mask                       */ 
#define MXC_F_FLC_CTRL_INFO_BLOCK_VALID_POS                 25                                                                                      /**< INFO_BLOCK_VALID Position          */
#define MXC_F_FLC_CTRL_INFO_BLOCK_VALID                     ((uint32_t)(0x00000001UL << MXC_F_FLC_CTRL_INFO_BLOCK_VALID_POS))                       /**< INFO_BLOCK_VALID Mask              */ 
#define MXC_F_FLC_CTRL_AUTO_INCRE_MODE_POS                  27                                                                                      /**< AUTO_INCRE_MODE Position           */
#define MXC_F_FLC_CTRL_AUTO_INCRE_MODE                      ((uint32_t)(0x00000001UL << MXC_F_FLC_CTRL_AUTO_INCRE_MODE_POS))                        /**< AUTO_INCRE_MODE Mask               */ 
#define MXC_F_FLC_CTRL_FLSH_UNLOCK_POS                      28                                                                                      /**< FLSH_UNLOCK Position               */
#define MXC_F_FLC_CTRL_FLSH_UNLOCK                          ((uint32_t)(0x0000000FUL << MXC_F_FLC_CTRL_FLSH_UNLOCK_POS))                            /**< FLSH_UNLOCK Mask                   */ 
/**@} end of group FLC_CTRL */
/**
 * @ingroup    flc_registers
 * @defgroup   FLC_INTR_Register FLC_INTR
 * @brief      Field Positions and Bit Masks for the FLC_INTR register. 
 * @{
 */
#define MXC_F_FLC_INTR_FINISHED_IF_POS                      0                                                                                       /**< FINISHED_IF Position               */
#define MXC_F_FLC_INTR_FINISHED_IF                          ((uint32_t)(0x00000001UL << MXC_F_FLC_INTR_FINISHED_IF_POS))                            /**< FINISHED_IF Mask                   */ 
#define MXC_F_FLC_INTR_FAILED_IF_POS                        1                                                                                       /**< FAILED_IF Position                 */
#define MXC_F_FLC_INTR_FAILED_IF                            ((uint32_t)(0x00000001UL << MXC_F_FLC_INTR_FAILED_IF_POS))                              /**< FAILED_IF Mask                     */ 
#define MXC_F_FLC_INTR_FINISHED_IE_POS                      8                                                                                       /**< FINISHED_IE Position               */
#define MXC_F_FLC_INTR_FINISHED_IE                          ((uint32_t)(0x00000001UL << MXC_F_FLC_INTR_FINISHED_IE_POS))                            /**< FINISHED_IE Mask                   */ 
#define MXC_F_FLC_INTR_FAILED_IE_POS                        9                                                                                       /**< FAILED_IE Position                 */
#define MXC_F_FLC_INTR_FAILED_IE                            ((uint32_t)(0x00000001UL << MXC_F_FLC_INTR_FAILED_IE_POS))                              /**< FAILED_IE Mask                     */ 
#define MXC_F_FLC_INTR_FAIL_FLAGS_POS                       16                                                                                      /**< FAIL_FLAGS Position                */
#define MXC_F_FLC_INTR_FAIL_FLAGS                           ((uint32_t)(0x0000FFFFUL << MXC_F_FLC_INTR_FAIL_FLAGS_POS))                             /**< FAIL_FLAGS Mask                    */ 
/**@} end of group FLC_INTR */
/**
 * @ingroup    flc_registers
 * @defgroup   FLC_PERFORM_Register FLC_PERFORM
 * @brief      Field Positions and Bit Masks for the FLC_PERFORM register. 
 * @{
 */
#define MXC_F_FLC_PERFORM_DELAY_SE_EN_POS                   0                                                                                       /**< DELAY_SE_EN Position               */
#define MXC_F_FLC_PERFORM_DELAY_SE_EN                       ((uint32_t)(0x00000001UL << MXC_F_FLC_PERFORM_DELAY_SE_EN_POS))                         /**< DELAY_SE_EN Mask                   */ 
#define MXC_F_FLC_PERFORM_FAST_READ_MODE_EN_POS             8                                                                                       /**< FAST_READ_MODE_EN Position         */
#define MXC_F_FLC_PERFORM_FAST_READ_MODE_EN                 ((uint32_t)(0x00000001UL << MXC_F_FLC_PERFORM_FAST_READ_MODE_EN_POS))                   /**< FAST_READ_MODE_EN Mask             */ 
#define MXC_F_FLC_PERFORM_EN_PREVENT_FAIL_POS               12                                                                                      /**< EN_PREVENT_FAIL Position           */
#define MXC_F_FLC_PERFORM_EN_PREVENT_FAIL                   ((uint32_t)(0x00000001UL << MXC_F_FLC_PERFORM_EN_PREVENT_FAIL_POS))                     /**< EN_PREVENT_FAIL Mask               */ 
#define MXC_F_FLC_PERFORM_EN_BACK2BACK_RDS_POS              16                                                                                      /**< EN_BACK2BACK_RDS Position          */
#define MXC_F_FLC_PERFORM_EN_BACK2BACK_RDS                  ((uint32_t)(0x00000001UL << MXC_F_FLC_PERFORM_EN_BACK2BACK_RDS_POS))                    /**< EN_BACK2BACK_RDS Mask              */ 
#define MXC_F_FLC_PERFORM_EN_BACK2BACK_WRS_POS              20                                                                                      /**< EN_BACK2BACK_WRS Position          */
#define MXC_F_FLC_PERFORM_EN_BACK2BACK_WRS                  ((uint32_t)(0x00000001UL << MXC_F_FLC_PERFORM_EN_BACK2BACK_WRS_POS))                    /**< EN_BACK2BACK_WRS Mask              */ 
#define MXC_F_FLC_PERFORM_EN_MERGE_GRAB_GNT_POS             24                                                                                      /**< EN_MERGE_GRAB_GNT Position         */
#define MXC_F_FLC_PERFORM_EN_MERGE_GRAB_GNT                 ((uint32_t)(0x00000001UL << MXC_F_FLC_PERFORM_EN_MERGE_GRAB_GNT_POS))                   /**< EN_MERGE_GRAB_GNT Mask             */ 
#define MXC_F_FLC_PERFORM_AUTO_TACC_POS                     28                                                                                      /**< AUTO_TACC Position                 */
#define MXC_F_FLC_PERFORM_AUTO_TACC                         ((uint32_t)(0x00000001UL << MXC_F_FLC_PERFORM_AUTO_TACC_POS))                           /**< AUTO_TACC Mask                     */ 
#define MXC_F_FLC_PERFORM_AUTO_CLKDIV_POS                   29                                                                                      /**< AUTO_CLKDIV Position               */
#define MXC_F_FLC_PERFORM_AUTO_CLKDIV                       ((uint32_t)(0x00000001UL << MXC_F_FLC_PERFORM_AUTO_CLKDIV_POS))                         /**< AUTO_CLKDIV Mask                   */   
/**@} end of group FLC_PERFORM */
/**
 * @ingroup    flc_registers
 * @defgroup   FLC_STATUS_Register FLC_STATUS
 * @brief      Field Positions and Bit Masks for the FLC_STATUS register. 
 * @{
 */
#define MXC_F_FLC_STATUS_JTAG_LOCK_WINDOW_POS               0                                                                                       /**< JTAG_LOCK_WINDOW Position          */
#define MXC_F_FLC_STATUS_JTAG_LOCK_WINDOW                   ((uint32_t)(0x00000001UL << MXC_F_FLC_STATUS_JTAG_LOCK_WINDOW_POS))                     /**< JTAG_LOCK_WINDOW Mask              */ 
#define MXC_F_FLC_STATUS_JTAG_LOCK_STATIC_POS               1                                                                                       /**< JTAG_LOCK_STATIC Position          */
#define MXC_F_FLC_STATUS_JTAG_LOCK_STATIC                   ((uint32_t)(0x00000001UL << MXC_F_FLC_STATUS_JTAG_LOCK_STATIC_POS))                     /**< JTAG_LOCK_STATIC Mask              */
#define MXC_F_FLC_STATUS_AUTO_LOCK_POS                      3                                                                                       /**< AUTO_LOCK Position                 */
#define MXC_F_FLC_STATUS_AUTO_LOCK                          ((uint32_t)(0x00000001UL << MXC_F_FLC_STATUS_AUTO_LOCK_POS))                            /**< AUTO_LOCK Mask                     */ 
#define MXC_F_FLC_STATUS_TRIM_UPDATE_DONE_POS               29                                                                                      /**< TRIM_UPDATE_DONE Position          */
#define MXC_F_FLC_STATUS_TRIM_UPDATE_DONE                   ((uint32_t)(0x00000001UL << MXC_F_FLC_STATUS_TRIM_UPDATE_DONE_POS))                     /**< TRIM_UPDATE_DONE Mask              */ 
#define MXC_F_FLC_STATUS_INFO_BLOCK_VALID_POS               30                                                                                      /**< INFO_BLOCK_VALID Position          */
#define MXC_F_FLC_STATUS_INFO_BLOCK_VALID                   ((uint32_t)(0x00000001UL << MXC_F_FLC_STATUS_INFO_BLOCK_VALID_POS))                     /**< INFO_BLOCK_VALID Mask              */ 
/**@} end of group FLC_STATUS*/
/**
 * @ingroup    flc_registers
 * @defgroup   FLC_SECURITY_Register FLC_SECURITY
 * @brief      Field Positions and Bit Masks for the FLC_SECURITY register. 
 * @{
 */
#define MXC_F_FLC_SECURITY_DEBUG_DISABLE_POS                0                                                                                       /**< DEBUG_DISABLE Position             */
#define MXC_F_FLC_SECURITY_DEBUG_DISABLE                    ((uint32_t)(0x000000FFUL << MXC_F_FLC_SECURITY_DEBUG_DISABLE_POS))                      /**< DEBUG_DISABLE Mask                 */ 
#define MXC_F_FLC_SECURITY_MASS_ERASE_LOCK_POS              8                                                                                       /**< MASS_ERASE_LOCK Position           */
#define MXC_F_FLC_SECURITY_MASS_ERASE_LOCK                  ((uint32_t)(0x0000000FUL << MXC_F_FLC_SECURITY_MASS_ERASE_LOCK_POS))                    /**< MASS_ERASE_LOCK Mask               */ 
#define MXC_F_FLC_SECURITY_DISABLE_AHB_WR_POS               16                                                                                      /**< DISABLE_AHB_WR Position            */
#define MXC_F_FLC_SECURITY_DISABLE_AHB_WR                   ((uint32_t)(0x0000000FUL << MXC_F_FLC_SECURITY_DISABLE_AHB_WR_POS))                     /**< DISABLE_AHB_WR Mask                */ 
#define MXC_F_FLC_SECURITY_FLC_SETTINGS_LOCK_POS            24                                                                                      /**< FLC_SETTINGS_LOCK Position         */
#define MXC_F_FLC_SECURITY_FLC_SETTINGS_LOCK                ((uint32_t)(0x0000000FUL << MXC_F_FLC_SECURITY_FLC_SETTINGS_LOCK_POS))                  /**< FLC_SETTINGS_LOCK Mask             */ 
#define MXC_F_FLC_SECURITY_SECURITY_LOCK_POS                28                                                                                      /**< SECURITY_LOCK Position             */
#define MXC_F_FLC_SECURITY_SECURITY_LOCK                    ((uint32_t)(0x0000000FUL << MXC_F_FLC_SECURITY_SECURITY_LOCK_POS))                      /**< SECURITY_LOCK Mask                 */ 
/**@} end of group FLC_SECURITY */
/**
 * @ingroup    flc_registers
 * @defgroup   FLC_BYPASS_Register FLC_BYPASS
 * @brief      Field Positions and Bit Masks for the FLC_BYPASS register. 
 * @{
 */
#define MXC_F_FLC_BYPASS_DESTRUCT_BYPASS_ERASE_POS          0                                                                                       /**< DESTRUCT_BYPASS_ERASE Position     */
#define MXC_F_FLC_BYPASS_DESTRUCT_BYPASS_ERASE              ((uint32_t)(0x00000001UL << MXC_F_FLC_BYPASS_DESTRUCT_BYPASS_ERASE_POS))                /**< DESTRUCT_BYPASS_ERASE Mask         */ 
#define MXC_F_FLC_BYPASS_SUPERWIPE_ERASE_POS                1                                                                                       /**< SUPERWIPE_ERASE Position           */
#define MXC_F_FLC_BYPASS_SUPERWIPE_ERASE                    ((uint32_t)(0x00000001UL << MXC_F_FLC_BYPASS_SUPERWIPE_ERASE_POS))                      /**< SUPERWIPE_ERASE Mask               */ 
#define MXC_F_FLC_BYPASS_DESTRUCT_BYPASS_COMPLETE_POS       2                                                                                       /**< DESTRUCT_BYPASS_COMPLETE Position  */
#define MXC_F_FLC_BYPASS_DESTRUCT_BYPASS_COMPLETE           ((uint32_t)(0x00000001UL << MXC_F_FLC_BYPASS_DESTRUCT_BYPASS_COMPLETE_POS))             /**< DESTRUCT_BYPASS_COMPLETE Mask      */ 
#define MXC_F_FLC_BYPASS_SUPERWIPE_COMPLETE_POS             3                                                                                       /**< SUPERWIPE_COMPLETE Position        */
#define MXC_F_FLC_BYPASS_SUPERWIPE_COMPLETE                 ((uint32_t)(0x00000001UL << MXC_F_FLC_BYPASS_SUPERWIPE_COMPLETE_POS))                   /**< SUPERWIPE_COMPLETE Mask            */ 
/**@} end of group FLC_BYPASS*/
/**
 * @ingroup    flc_registers
 * @defgroup   FLC_CTRL2_Register FLC_CTRL2
 * @brief      Field Positions and Bit Masks for the FLC_CTRL2 register. 
 * @{
 */
#define MXC_F_FLC_CTRL2_FLASH_LVE_POS                       0                                                                                       /**< FLASH_LVE Position                 */
#define MXC_F_FLC_CTRL2_FLASH_LVE                           ((uint32_t)(0x00000001UL << MXC_F_FLC_CTRL2_FLASH_LVE_POS))                             /**< FLASH_LVE Mask                     */ 
#define MXC_F_FLC_CTRL2_FRC_FCLK1_ON_POS                    1                                                                                       /**< FRC_FCLK1_ON Position              */
#define MXC_F_FLC_CTRL2_FRC_FCLK1_ON                        ((uint32_t)(0x00000001UL << MXC_F_FLC_CTRL2_FRC_FCLK1_ON_POS))                          /**< FRC_FCLK1_ON Mask                  */ 
#define MXC_F_FLC_CTRL2_EN_WRITE_ALL_ZEROES_POS             3                                                                                       /**< EN_WRITE_ALL_ZEROES Position       */
#define MXC_F_FLC_CTRL2_EN_WRITE_ALL_ZEROES                 ((uint32_t)(0x00000001UL << MXC_F_FLC_CTRL2_EN_WRITE_ALL_ZEROES_POS))                   /**< EN_WRITE_ALL_ZEROES Mask           */ 
#define MXC_F_FLC_CTRL2_EN_CHANGE_POS                       4                                                                                       /**< EN_CHANGE Position                 */
#define MXC_F_FLC_CTRL2_EN_CHANGE                           ((uint32_t)(0x00000001UL << MXC_F_FLC_CTRL2_EN_CHANGE_POS))                             /**< EN_CHANGE Mask                     */ 
#define MXC_F_FLC_CTRL2_SLOW_CLK_POS                        5                                                                                       /**< SLOW_CLK Position                  */
#define MXC_F_FLC_CTRL2_SLOW_CLK                            ((uint32_t)(0x00000001UL << MXC_F_FLC_CTRL2_SLOW_CLK_POS))                              /**< SLOW_CLK Mask                      */ 
#define MXC_F_FLC_CTRL2_ENABLE_RAM_HRESP_POS                6                                                                                       /**< ENABLE_RAM_HRESP Position          */
#define MXC_F_FLC_CTRL2_ENABLE_RAM_HRESP                    ((uint32_t)(0x00000001UL << MXC_F_FLC_CTRL2_ENABLE_RAM_HRESP_POS))                      /**< ENABLE_RAM_HRESP Mask              */ 
#define MXC_F_FLC_CTRL2_BYPASS_AHB_FAIL_POS                 8                                                                                       /**< BYPASS_AHB_FAIL Position           */
#define MXC_F_FLC_CTRL2_BYPASS_AHB_FAIL                     ((uint32_t)(0x000000FFUL << MXC_F_FLC_CTRL2_BYPASS_AHB_FAIL_POS))                       /**< BYPASS_AHB_FAIL Mask               */ 
/**@} end of group FLC_CTRL2*/
   /**
 * @ingroup    flc_registers
 * @defgroup   FLC_INTFL1_Register FLC_INTFL1
 * @brief      Field Positions and Bit Masks for the FLC_INTFL1 register. 
 * @{
 */
#define MXC_F_FLC_INTFL1_SRAM_ADDR_WRAPPED_POS              0                                                                                       /**< SRAM_ADDR_WRAPPED Position         */
#define MXC_F_FLC_INTFL1_SRAM_ADDR_WRAPPED                  ((uint32_t)(0x00000001UL << MXC_F_FLC_INTFL1_SRAM_ADDR_WRAPPED_POS))                    /**< SRAM_ADDR_WRAPPED Mask             */ 
#define MXC_F_FLC_INTFL1_INVALID_FLASH_ADDR_POS             1                                                                                       /**< INVALID_FLASH_ADDR Position        */
#define MXC_F_FLC_INTFL1_INVALID_FLASH_ADDR                 ((uint32_t)(0x00000001UL << MXC_F_FLC_INTFL1_INVALID_FLASH_ADDR_POS))                   /**< INVALID_FLASH_ADDR Mask            */ 
#define MXC_F_FLC_INTFL1_FLASH_READ_LOCKED_POS              2                                                                                       /**< FLASH_READ_LOCKED Position         */
#define MXC_F_FLC_INTFL1_FLASH_READ_LOCKED                  ((uint32_t)(0x00000001UL << MXC_F_FLC_INTFL1_FLASH_READ_LOCKED_POS))                    /**< FLASH_READ_LOCKED Mask             */ 
#define MXC_F_FLC_INTFL1_TRIM_UPDATE_DONE_POS               3                                                                                       /**< TRIM_UPDATE_DONE Position          */
#define MXC_F_FLC_INTFL1_TRIM_UPDATE_DONE                   ((uint32_t)(0x00000001UL << MXC_F_FLC_INTFL1_TRIM_UPDATE_DONE_POS))                     /**< TRIM_UPDATE_DONE Mask              */ 
#define MXC_F_FLC_INTFL1_FLC_STATE_DONE_POS                 4                                                                                       /**< FLC_STATE_DONE Position            */
#define MXC_F_FLC_INTFL1_FLC_STATE_DONE                     ((uint32_t)(0x00000001UL << MXC_F_FLC_INTFL1_FLC_STATE_DONE_POS))                       /**< FLC_STATE_DONE Mask                */ 
#define MXC_F_FLC_INTFL1_FLC_PROG_COMPLETE_POS              5                                                                                       /**< FLC_PROG_COMPLETE Position         */
#define MXC_F_FLC_INTFL1_FLC_PROG_COMPLETE                  ((uint32_t)(0x00000001UL << MXC_F_FLC_INTFL1_FLC_PROG_COMPLETE_POS))                    /**< FLC_PROG_COMPLETE Mask             */ 
/**@} end of group FLC_INTFL1 */
/**
 * @ingroup    flc_registers
 * @defgroup   FLC_INTEN1_Register FLC_INTEN1
 * @brief      Field Positions and Bit Masks for the FLC_INTEN1 register. 
 * @{
 */
#define MXC_F_FLC_INTEN1_SRAM_ADDR_WRAPPED_POS              0                                                                                       /**< SRAM_ADDR_WRAPPED Position         */
#define MXC_F_FLC_INTEN1_SRAM_ADDR_WRAPPED                  ((uint32_t)(0x00000001UL << MXC_F_FLC_INTEN1_SRAM_ADDR_WRAPPED_POS))                    /**< SRAM_ADDR_WRAPPED Mask             */ 
#define MXC_F_FLC_INTEN1_INVALID_FLASH_ADDR_POS             1                                                                                       /**< INVALID_FLASH_ADDR Position        */
#define MXC_F_FLC_INTEN1_INVALID_FLASH_ADDR                 ((uint32_t)(0x00000001UL << MXC_F_FLC_INTEN1_INVALID_FLASH_ADDR_POS))                   /**< INVALID_FLASH_ADDR Mask            */ 
#define MXC_F_FLC_INTEN1_FLASH_READ_LOCKED_POS              2                                                                                       /**< FLASH_READ_LOCKED Position         */
#define MXC_F_FLC_INTEN1_FLASH_READ_LOCKED                  ((uint32_t)(0x00000001UL << MXC_F_FLC_INTEN1_FLASH_READ_LOCKED_POS))                    /**< FLASH_READ_LOCKED Mask             */ 
#define MXC_F_FLC_INTEN1_TRIM_UPDATE_DONE_POS               3                                                                                       /**< TRIM_UPDATE_DONE Position          */
#define MXC_F_FLC_INTEN1_TRIM_UPDATE_DONE                   ((uint32_t)(0x00000001UL << MXC_F_FLC_INTEN1_TRIM_UPDATE_DONE_POS))                     /**< TRIM_UPDATE_DONE Mask              */ 
#define MXC_F_FLC_INTEN1_FLC_STATE_DONE_POS                 4                                                                                       /**< FLC_STATE_DONE Position            */
#define MXC_F_FLC_INTEN1_FLC_STATE_DONE                     ((uint32_t)(0x00000001UL << MXC_F_FLC_INTEN1_FLC_STATE_DONE_POS))                       /**< FLC_STATE_DONE Mask                */ 
#define MXC_F_FLC_INTEN1_FLC_PROG_COMPLETE_POS              5                                                                                       /**< FLC_PROG_COMPLETE Position         */
#define MXC_F_FLC_INTEN1_FLC_PROG_COMPLETE                  ((uint32_t)(0x00000001UL << MXC_F_FLC_INTEN1_FLC_PROG_COMPLETE_POS))                    /**< FLC_PROG_COMPLETE Mask             */ 
/**@} end of group FLC_INTEN1*/
/**@} end of ingroup flc_registers */
#ifdef __cplusplus
}
#endif

#endif   /* _MXC_FLC_REGS_H_ */

