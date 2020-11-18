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

#ifndef _MXC_FLC_REGS_H_
#define _MXC_FLC_REGS_H_

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


#define MXC_V_FLC_ERASE_CODE_PAGE_ERASE   ((uint8_t)0x55)
#define MXC_V_FLC_ERASE_CODE_MASS_ERASE   ((uint8_t)0xAA)
#define MXC_V_FLC_FLSH_UNLOCK_KEY         ((uint8_t)0x2)

/*
   Typedefed structure(s) for module registers (per instance or section) with direct 32-bit
   access to each register in module.
*/

/*                                                          Offset          Register Description
                                                            =============   ============================================================================ */
typedef struct {
    __IO uint32_t faddr;                                /*  0x0000          Flash Operation Address                                                      */
    __IO uint32_t fckdiv;                               /*  0x0004          Flash Clock Pulse Divisor                                                    */
    __IO uint32_t ctrl;                                 /*  0x0008          Flash Control Register                                                       */
    __R  uint32_t rsv00C[6];                            /*  0x000C-0x0020                                                                                */
    __IO uint32_t intr;                                 /*  0x0024          Flash Controller Interrupt Flags and Enable/Disable 0                        */
    __R  uint32_t rsv028[2];                            /*  0x0028-0x002C                                                                                */
    __IO uint32_t fdata;                                /*  0x0030          Flash Operation Data Register                                                */
    __R  uint32_t rsv034[7];                            /*  0x0034-0x004C                                                                                */
    __IO uint32_t perform;                              /*  0x0050          Flash Performance Settings                                                   */
    __IO uint32_t tacc;                                 /*  0x0054          Flash Read Cycle Config                                                      */
    __IO uint32_t tprog;                                /*  0x0058          Flash Write Cycle Config                                                     */
    __R  uint32_t rsv05C[9];                            /*  0x005C-0x007C                                                                                */
    __IO uint32_t status;                               /*  0x0080          Security Status Flags                                                        */
    __R  uint32_t rsv084;                               /*  0x0084                                                                                       */
    __IO uint32_t security;                             /*  0x0088          Flash Controller Security Settings                                           */
    __R  uint32_t rsv08C[4];                            /*  0x008C-0x0098                                                                                */
    __IO uint32_t bypass;                               /*  0x009C          Status Flags for DSB Operations                                              */
    __R  uint32_t rsv0A0[24];                           /*  0x00A0-0x00FC                                                                                */
    __IO uint32_t user_option;                          /*  0x0100          Used to set DSB Access code and Auto-Lock in info block                      */
    __R  uint32_t rsv104[15];                           /*  0x0104-0x013C                                                                                */
    __IO uint32_t ctrl2;                                /*  0x0140          Flash Control Register 2                                                     */
    __IO uint32_t intfl1;                               /*  0x0144          Interrupt Flags Register 1                                                   */
    __IO uint32_t inten1;                               /*  0x0148          Interrupt Enable/Disable Register 1                                          */
    __R  uint32_t rsv14C[9];                            /*  0x014C-0x016C                                                                                */
    __IO uint32_t bl_ctrl;                              /*  0x0170          Bootloader Control Register                                                  */
    __IO uint32_t twk;                                  /*  0x0174          PDM33 Register                                                               */
    __R  uint32_t rsv178;                               /*  0x0178                                                                                       */
    __IO uint32_t slm;                                  /*  0x017C          Sleep Mode Register                                                          */
    __R  uint32_t rsv180[32];                           /*  0x0180-0x01FC                                                                                */
    __IO uint32_t disable_xr0;                          /*  0x0200          Disable Flash Page Exec/Read Register 0                                      */
    __IO uint32_t disable_xr1;                          /*  0x0204          Disable Flash Page Exec/Read Register 1                                      */
    __IO uint32_t disable_xr2;                          /*  0x0208          Disable Flash Page Exec/Read Register 2                                      */
    __IO uint32_t disable_xr3;                          /*  0x020C          Disable Flash Page Exec/Read Register 3                                      */
    __IO uint32_t disable_xr4;                          /*  0x0210          Disable Flash Page Exec/Read Register 4                                      */
    __IO uint32_t disable_xr5;                          /*  0x0214          Disable Flash Page Exec/Read Register 5                                      */
    __IO uint32_t disable_xr6;                          /*  0x0218          Disable Flash Page Exec/Read Register 6                                      */
    __IO uint32_t disable_xr7;                          /*  0x021C          Disable Flash Page Exec/Read Register 7                                      */
    __R  uint32_t rsv220[56];                           /*  0x0220-0x02FC                                                                                */
    __IO uint32_t disable_we0;                          /*  0x0300          Disable Flash Page Write/Erase Register 0                                    */
    __IO uint32_t disable_we1;                          /*  0x0304          Disable Flash Page Write/Erase Register 1                                    */
    __IO uint32_t disable_we2;                          /*  0x0308          Disable Flash Page Write/Erase Register 2                                    */
    __IO uint32_t disable_we3;                          /*  0x030C          Disable Flash Page Write/Erase Register 3                                    */
    __IO uint32_t disable_we4;                          /*  0x0310          Disable Flash Page Write/Erase Register 4                                    */
    __IO uint32_t disable_we5;                          /*  0x0314          Disable Flash Page Write/Erase Register 5                                    */
    __IO uint32_t disable_we6;                          /*  0x0318          Disable Flash Page Write/Erase Register 6                                    */
    __IO uint32_t disable_we7;                          /*  0x031C          Disable Flash Page Write/Erase Register 7                                    */
} mxc_flc_regs_t;


/*
   Register offsets for module FLC.
*/

#define MXC_R_FLC_OFFS_FADDR                                ((uint32_t)0x00000000UL)
#define MXC_R_FLC_OFFS_FCKDIV                               ((uint32_t)0x00000004UL)
#define MXC_R_FLC_OFFS_CTRL                                 ((uint32_t)0x00000008UL)
#define MXC_R_FLC_OFFS_INTR                                 ((uint32_t)0x00000024UL)
#define MXC_R_FLC_OFFS_FDATA                                ((uint32_t)0x00000030UL)
#define MXC_R_FLC_OFFS_PERFORM                              ((uint32_t)0x00000050UL)
#define MXC_R_FLC_OFFS_TACC                                 ((uint32_t)0x00000054UL)
#define MXC_R_FLC_OFFS_TPROG                                ((uint32_t)0x00000058UL)
#define MXC_R_FLC_OFFS_STATUS                               ((uint32_t)0x00000080UL)
#define MXC_R_FLC_OFFS_SECURITY                             ((uint32_t)0x00000088UL)
#define MXC_R_FLC_OFFS_BYPASS                               ((uint32_t)0x0000009CUL)
#define MXC_R_FLC_OFFS_USER_OPTION                          ((uint32_t)0x00000100UL)
#define MXC_R_FLC_OFFS_CTRL2                                ((uint32_t)0x00000140UL)
#define MXC_R_FLC_OFFS_INTFL1                               ((uint32_t)0x00000144UL)
#define MXC_R_FLC_OFFS_INTEN1                               ((uint32_t)0x00000148UL)
#define MXC_R_FLC_OFFS_BL_CTRL                              ((uint32_t)0x00000170UL)
#define MXC_R_FLC_OFFS_TWK                                  ((uint32_t)0x00000174UL)
#define MXC_R_FLC_OFFS_SLM                                  ((uint32_t)0x0000017CUL)
#define MXC_R_FLC_OFFS_DISABLE_XR0                          ((uint32_t)0x00000200UL)
#define MXC_R_FLC_OFFS_DISABLE_XR1                          ((uint32_t)0x00000204UL)
#define MXC_R_FLC_OFFS_DISABLE_XR2                          ((uint32_t)0x00000208UL)
#define MXC_R_FLC_OFFS_DISABLE_XR3                          ((uint32_t)0x0000020CUL)
#define MXC_R_FLC_OFFS_DISABLE_XR4                          ((uint32_t)0x00000210UL)
#define MXC_R_FLC_OFFS_DISABLE_XR5                          ((uint32_t)0x00000214UL)
#define MXC_R_FLC_OFFS_DISABLE_XR6                          ((uint32_t)0x00000218UL)
#define MXC_R_FLC_OFFS_DISABLE_XR7                          ((uint32_t)0x0000021CUL)
#define MXC_R_FLC_OFFS_DISABLE_WE0                          ((uint32_t)0x00000300UL)
#define MXC_R_FLC_OFFS_DISABLE_WE1                          ((uint32_t)0x00000304UL)
#define MXC_R_FLC_OFFS_DISABLE_WE2                          ((uint32_t)0x00000308UL)
#define MXC_R_FLC_OFFS_DISABLE_WE3                          ((uint32_t)0x0000030CUL)
#define MXC_R_FLC_OFFS_DISABLE_WE4                          ((uint32_t)0x00000310UL)
#define MXC_R_FLC_OFFS_DISABLE_WE5                          ((uint32_t)0x00000314UL)
#define MXC_R_FLC_OFFS_DISABLE_WE6                          ((uint32_t)0x00000318UL)
#define MXC_R_FLC_OFFS_DISABLE_WE7                          ((uint32_t)0x0000031CUL)


/*
   Field positions and masks for module FLC.
*/

#define MXC_F_FLC_FADDR_FADDR_POS                           0
#define MXC_F_FLC_FADDR_FADDR                               ((uint32_t)(0x003FFFFFUL << MXC_F_FLC_FADDR_FADDR_POS))

#define MXC_F_FLC_FCKDIV_FCKDIV_POS                         0
#define MXC_F_FLC_FCKDIV_FCKDIV                             ((uint32_t)(0x0000007FUL << MXC_F_FLC_FCKDIV_FCKDIV_POS))
#define MXC_F_FLC_FCKDIV_AUTO_FCKDIV_RESULT_POS             16
#define MXC_F_FLC_FCKDIV_AUTO_FCKDIV_RESULT                 ((uint32_t)(0x0000FFFFUL << MXC_F_FLC_FCKDIV_AUTO_FCKDIV_RESULT_POS))

#define MXC_F_FLC_CTRL_WRITE_POS                            0
#define MXC_F_FLC_CTRL_WRITE                                ((uint32_t)(0x00000001UL << MXC_F_FLC_CTRL_WRITE_POS))
#define MXC_F_FLC_CTRL_MASS_ERASE_POS                       1
#define MXC_F_FLC_CTRL_MASS_ERASE                           ((uint32_t)(0x00000001UL << MXC_F_FLC_CTRL_MASS_ERASE_POS))
#define MXC_F_FLC_CTRL_PAGE_ERASE_POS                       2
#define MXC_F_FLC_CTRL_PAGE_ERASE                           ((uint32_t)(0x00000001UL << MXC_F_FLC_CTRL_PAGE_ERASE_POS))
#define MXC_F_FLC_CTRL_ERASE_CODE_POS                       8
#define MXC_F_FLC_CTRL_ERASE_CODE                           ((uint32_t)(0x000000FFUL << MXC_F_FLC_CTRL_ERASE_CODE_POS))
#define MXC_F_FLC_CTRL_INFO_BLOCK_UNLOCK_POS                16
#define MXC_F_FLC_CTRL_INFO_BLOCK_UNLOCK                    ((uint32_t)(0x00000001UL << MXC_F_FLC_CTRL_INFO_BLOCK_UNLOCK_POS))
#define MXC_F_FLC_CTRL_WRITE_ENABLE_POS                     17
#define MXC_F_FLC_CTRL_WRITE_ENABLE                         ((uint32_t)(0x00000001UL << MXC_F_FLC_CTRL_WRITE_ENABLE_POS))
#define MXC_F_FLC_CTRL_PENDING_POS                          24
#define MXC_F_FLC_CTRL_PENDING                              ((uint32_t)(0x00000001UL << MXC_F_FLC_CTRL_PENDING_POS))
#define MXC_F_FLC_CTRL_INFO_BLOCK_VALID_POS                 25
#define MXC_F_FLC_CTRL_INFO_BLOCK_VALID                     ((uint32_t)(0x00000001UL << MXC_F_FLC_CTRL_INFO_BLOCK_VALID_POS))
#define MXC_F_FLC_CTRL_AUTO_INCRE_MODE_POS                  27
#define MXC_F_FLC_CTRL_AUTO_INCRE_MODE                      ((uint32_t)(0x00000001UL << MXC_F_FLC_CTRL_AUTO_INCRE_MODE_POS))
#define MXC_F_FLC_CTRL_FLSH_UNLOCK_POS                      28
#define MXC_F_FLC_CTRL_FLSH_UNLOCK                          ((uint32_t)(0x0000000FUL << MXC_F_FLC_CTRL_FLSH_UNLOCK_POS))

#define MXC_F_FLC_INTR_FINISHED_IF_POS                      0
#define MXC_F_FLC_INTR_FINISHED_IF                          ((uint32_t)(0x00000001UL << MXC_F_FLC_INTR_FINISHED_IF_POS))
#define MXC_F_FLC_INTR_FAILED_IF_POS                        1
#define MXC_F_FLC_INTR_FAILED_IF                            ((uint32_t)(0x00000001UL << MXC_F_FLC_INTR_FAILED_IF_POS))
#define MXC_F_FLC_INTR_FINISHED_IE_POS                      8
#define MXC_F_FLC_INTR_FINISHED_IE                          ((uint32_t)(0x00000001UL << MXC_F_FLC_INTR_FINISHED_IE_POS))
#define MXC_F_FLC_INTR_FAILED_IE_POS                        9
#define MXC_F_FLC_INTR_FAILED_IE                            ((uint32_t)(0x00000001UL << MXC_F_FLC_INTR_FAILED_IE_POS))
#define MXC_F_FLC_INTR_FAIL_FLAGS_POS                       16
#define MXC_F_FLC_INTR_FAIL_FLAGS                           ((uint32_t)(0x0000FFFFUL << MXC_F_FLC_INTR_FAIL_FLAGS_POS))

#define MXC_F_FLC_PERFORM_DELAY_SE_EN_POS                   0
#define MXC_F_FLC_PERFORM_DELAY_SE_EN                       ((uint32_t)(0x00000001UL << MXC_F_FLC_PERFORM_DELAY_SE_EN_POS))
#define MXC_F_FLC_PERFORM_FAST_READ_MODE_EN_POS             8
#define MXC_F_FLC_PERFORM_FAST_READ_MODE_EN                 ((uint32_t)(0x00000001UL << MXC_F_FLC_PERFORM_FAST_READ_MODE_EN_POS))
#define MXC_F_FLC_PERFORM_EN_PREVENT_FAIL_POS               12
#define MXC_F_FLC_PERFORM_EN_PREVENT_FAIL                   ((uint32_t)(0x00000001UL << MXC_F_FLC_PERFORM_EN_PREVENT_FAIL_POS))
#define MXC_F_FLC_PERFORM_EN_BACK2BACK_RDS_POS              16
#define MXC_F_FLC_PERFORM_EN_BACK2BACK_RDS                  ((uint32_t)(0x00000001UL << MXC_F_FLC_PERFORM_EN_BACK2BACK_RDS_POS))
#define MXC_F_FLC_PERFORM_EN_BACK2BACK_WRS_POS              20
#define MXC_F_FLC_PERFORM_EN_BACK2BACK_WRS                  ((uint32_t)(0x00000001UL << MXC_F_FLC_PERFORM_EN_BACK2BACK_WRS_POS))
#define MXC_F_FLC_PERFORM_EN_MERGE_GRAB_GNT_POS             24
#define MXC_F_FLC_PERFORM_EN_MERGE_GRAB_GNT                 ((uint32_t)(0x00000001UL << MXC_F_FLC_PERFORM_EN_MERGE_GRAB_GNT_POS))
#define MXC_F_FLC_PERFORM_AUTO_TACC_POS                     28
#define MXC_F_FLC_PERFORM_AUTO_TACC                         ((uint32_t)(0x00000001UL << MXC_F_FLC_PERFORM_AUTO_TACC_POS))
#define MXC_F_FLC_PERFORM_AUTO_CLKDIV_POS                   29
#define MXC_F_FLC_PERFORM_AUTO_CLKDIV                       ((uint32_t)(0x00000001UL << MXC_F_FLC_PERFORM_AUTO_CLKDIV_POS))

#define MXC_F_FLC_STATUS_JTAG_LOCK_WINDOW_POS               0
#define MXC_F_FLC_STATUS_JTAG_LOCK_WINDOW                   ((uint32_t)(0x00000001UL << MXC_F_FLC_STATUS_JTAG_LOCK_WINDOW_POS))
#define MXC_F_FLC_STATUS_JTAG_LOCK_STATIC_POS               1
#define MXC_F_FLC_STATUS_JTAG_LOCK_STATIC                   ((uint32_t)(0x00000001UL << MXC_F_FLC_STATUS_JTAG_LOCK_STATIC_POS))
#define MXC_F_FLC_STATUS_AUTO_LOCK_POS                      3
#define MXC_F_FLC_STATUS_AUTO_LOCK                          ((uint32_t)(0x00000001UL << MXC_F_FLC_STATUS_AUTO_LOCK_POS))
#define MXC_F_FLC_STATUS_TRIM_UPDATE_DONE_POS               29
#define MXC_F_FLC_STATUS_TRIM_UPDATE_DONE                   ((uint32_t)(0x00000001UL << MXC_F_FLC_STATUS_TRIM_UPDATE_DONE_POS))
#define MXC_F_FLC_STATUS_INFO_BLOCK_VALID_POS               30
#define MXC_F_FLC_STATUS_INFO_BLOCK_VALID                   ((uint32_t)(0x00000001UL << MXC_F_FLC_STATUS_INFO_BLOCK_VALID_POS))

#define MXC_F_FLC_SECURITY_DEBUG_DISABLE_POS                0
#define MXC_F_FLC_SECURITY_DEBUG_DISABLE                    ((uint32_t)(0x000000FFUL << MXC_F_FLC_SECURITY_DEBUG_DISABLE_POS))
#define MXC_F_FLC_SECURITY_MASS_ERASE_LOCK_POS              8
#define MXC_F_FLC_SECURITY_MASS_ERASE_LOCK                  ((uint32_t)(0x0000000FUL << MXC_F_FLC_SECURITY_MASS_ERASE_LOCK_POS))
#define MXC_F_FLC_SECURITY_DISABLE_AHB_WR_POS               16
#define MXC_F_FLC_SECURITY_DISABLE_AHB_WR                   ((uint32_t)(0x0000000FUL << MXC_F_FLC_SECURITY_DISABLE_AHB_WR_POS))
#define MXC_F_FLC_SECURITY_FLC_SETTINGS_LOCK_POS            24
#define MXC_F_FLC_SECURITY_FLC_SETTINGS_LOCK                ((uint32_t)(0x0000000FUL << MXC_F_FLC_SECURITY_FLC_SETTINGS_LOCK_POS))
#define MXC_F_FLC_SECURITY_SECURITY_LOCK_POS                28
#define MXC_F_FLC_SECURITY_SECURITY_LOCK                    ((uint32_t)(0x0000000FUL << MXC_F_FLC_SECURITY_SECURITY_LOCK_POS))

#define MXC_F_FLC_BYPASS_DESTRUCT_BYPASS_ERASE_POS          0
#define MXC_F_FLC_BYPASS_DESTRUCT_BYPASS_ERASE              ((uint32_t)(0x00000001UL << MXC_F_FLC_BYPASS_DESTRUCT_BYPASS_ERASE_POS))
#define MXC_F_FLC_BYPASS_SUPERWIPE_ERASE_POS                1
#define MXC_F_FLC_BYPASS_SUPERWIPE_ERASE                    ((uint32_t)(0x00000001UL << MXC_F_FLC_BYPASS_SUPERWIPE_ERASE_POS))
#define MXC_F_FLC_BYPASS_DESTRUCT_BYPASS_COMPLETE_POS       2
#define MXC_F_FLC_BYPASS_DESTRUCT_BYPASS_COMPLETE           ((uint32_t)(0x00000001UL << MXC_F_FLC_BYPASS_DESTRUCT_BYPASS_COMPLETE_POS))
#define MXC_F_FLC_BYPASS_SUPERWIPE_COMPLETE_POS             3
#define MXC_F_FLC_BYPASS_SUPERWIPE_COMPLETE                 ((uint32_t)(0x00000001UL << MXC_F_FLC_BYPASS_SUPERWIPE_COMPLETE_POS))

#define MXC_F_FLC_CTRL2_FLASH_LVE_POS                       0
#define MXC_F_FLC_CTRL2_FLASH_LVE                           ((uint32_t)(0x00000001UL << MXC_F_FLC_CTRL2_FLASH_LVE_POS))
#define MXC_F_FLC_CTRL2_FRC_FCLK1_ON_POS                    1
#define MXC_F_FLC_CTRL2_FRC_FCLK1_ON                        ((uint32_t)(0x00000001UL << MXC_F_FLC_CTRL2_FRC_FCLK1_ON_POS))
#define MXC_F_FLC_CTRL2_EN_WRITE_ALL_ZEROES_POS             3
#define MXC_F_FLC_CTRL2_EN_WRITE_ALL_ZEROES                 ((uint32_t)(0x00000001UL << MXC_F_FLC_CTRL2_EN_WRITE_ALL_ZEROES_POS))
#define MXC_F_FLC_CTRL2_EN_CHANGE_POS                       4
#define MXC_F_FLC_CTRL2_EN_CHANGE                           ((uint32_t)(0x00000001UL << MXC_F_FLC_CTRL2_EN_CHANGE_POS))
#define MXC_F_FLC_CTRL2_SLOW_CLK_POS                        5
#define MXC_F_FLC_CTRL2_SLOW_CLK                            ((uint32_t)(0x00000001UL << MXC_F_FLC_CTRL2_SLOW_CLK_POS))
#define MXC_F_FLC_CTRL2_ENABLE_RAM_HRESP_POS                6
#define MXC_F_FLC_CTRL2_ENABLE_RAM_HRESP                    ((uint32_t)(0x00000001UL << MXC_F_FLC_CTRL2_ENABLE_RAM_HRESP_POS))
#define MXC_F_FLC_CTRL2_BYPASS_AHB_FAIL_POS                 8
#define MXC_F_FLC_CTRL2_BYPASS_AHB_FAIL                     ((uint32_t)(0x000000FFUL << MXC_F_FLC_CTRL2_BYPASS_AHB_FAIL_POS))

#define MXC_F_FLC_INTFL1_SRAM_ADDR_WRAPPED_POS              0
#define MXC_F_FLC_INTFL1_SRAM_ADDR_WRAPPED                  ((uint32_t)(0x00000001UL << MXC_F_FLC_INTFL1_SRAM_ADDR_WRAPPED_POS))
#define MXC_F_FLC_INTFL1_INVALID_FLASH_ADDR_POS             1
#define MXC_F_FLC_INTFL1_INVALID_FLASH_ADDR                 ((uint32_t)(0x00000001UL << MXC_F_FLC_INTFL1_INVALID_FLASH_ADDR_POS))
#define MXC_F_FLC_INTFL1_FLASH_READ_LOCKED_POS              2
#define MXC_F_FLC_INTFL1_FLASH_READ_LOCKED                  ((uint32_t)(0x00000001UL << MXC_F_FLC_INTFL1_FLASH_READ_LOCKED_POS))
#define MXC_F_FLC_INTFL1_TRIM_UPDATE_DONE_POS               3
#define MXC_F_FLC_INTFL1_TRIM_UPDATE_DONE                   ((uint32_t)(0x00000001UL << MXC_F_FLC_INTFL1_TRIM_UPDATE_DONE_POS))
#define MXC_F_FLC_INTFL1_FLC_STATE_DONE_POS                 4
#define MXC_F_FLC_INTFL1_FLC_STATE_DONE                     ((uint32_t)(0x00000001UL << MXC_F_FLC_INTFL1_FLC_STATE_DONE_POS))
#define MXC_F_FLC_INTFL1_FLC_PROG_COMPLETE_POS              5
#define MXC_F_FLC_INTFL1_FLC_PROG_COMPLETE                  ((uint32_t)(0x00000001UL << MXC_F_FLC_INTFL1_FLC_PROG_COMPLETE_POS))

#define MXC_F_FLC_INTEN1_SRAM_ADDR_WRAPPED_POS              0
#define MXC_F_FLC_INTEN1_SRAM_ADDR_WRAPPED                  ((uint32_t)(0x00000001UL << MXC_F_FLC_INTEN1_SRAM_ADDR_WRAPPED_POS))
#define MXC_F_FLC_INTEN1_INVALID_FLASH_ADDR_POS             1
#define MXC_F_FLC_INTEN1_INVALID_FLASH_ADDR                 ((uint32_t)(0x00000001UL << MXC_F_FLC_INTEN1_INVALID_FLASH_ADDR_POS))
#define MXC_F_FLC_INTEN1_FLASH_READ_LOCKED_POS              2
#define MXC_F_FLC_INTEN1_FLASH_READ_LOCKED                  ((uint32_t)(0x00000001UL << MXC_F_FLC_INTEN1_FLASH_READ_LOCKED_POS))
#define MXC_F_FLC_INTEN1_TRIM_UPDATE_DONE_POS               3
#define MXC_F_FLC_INTEN1_TRIM_UPDATE_DONE                   ((uint32_t)(0x00000001UL << MXC_F_FLC_INTEN1_TRIM_UPDATE_DONE_POS))
#define MXC_F_FLC_INTEN1_FLC_STATE_DONE_POS                 4
#define MXC_F_FLC_INTEN1_FLC_STATE_DONE                     ((uint32_t)(0x00000001UL << MXC_F_FLC_INTEN1_FLC_STATE_DONE_POS))
#define MXC_F_FLC_INTEN1_FLC_PROG_COMPLETE_POS              5
#define MXC_F_FLC_INTEN1_FLC_PROG_COMPLETE                  ((uint32_t)(0x00000001UL << MXC_F_FLC_INTEN1_FLC_PROG_COMPLETE_POS))



#ifdef __cplusplus
}
#endif

#endif   /* _MXC_FLC_REGS_H_ */

