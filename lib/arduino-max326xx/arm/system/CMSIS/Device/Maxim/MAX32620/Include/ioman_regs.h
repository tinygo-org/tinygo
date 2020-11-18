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
 * $Date: 2016-06-03 16:02:37 -0500 (Fri, 03 Jun 2016) $
 * $Revision: 23203 $
 *
 ******************************************************************************/

#ifndef _MXC_IOMAN_REGS_H_
#define _MXC_IOMAN_REGS_H_

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
    Bitfield structs for registers in this module
*/

typedef struct {
    uint32_t wud_req_p0 : 8;
    uint32_t wud_req_p1 : 8;
    uint32_t wud_req_p2 : 8;
    uint32_t wud_req_p3 : 8;
} mxc_ioman_wud_req0_t;

typedef struct {
    uint32_t wud_req_p4 : 8;
    uint32_t wud_req_p5 : 8;
    uint32_t wud_req_p6 : 8;
    uint32_t wud_req_p7 : 8;
} mxc_ioman_wud_req1_t;

typedef struct {
    uint32_t wud_ack_p0 : 8;
    uint32_t wud_ack_p1 : 8;
    uint32_t wud_ack_p2 : 8;
    uint32_t wud_ack_p3 : 8;
} mxc_ioman_wud_ack0_t;

typedef struct {
    uint32_t wud_ack_p4 : 8;
    uint32_t wud_ack_p5 : 8;
    uint32_t wud_ack_p6 : 8;
    uint32_t wud_ack_p7 : 8;
} mxc_ioman_wud_ack1_t;

typedef struct {
    uint32_t ali_req_p0 : 8;
    uint32_t ali_req_p1 : 8;
    uint32_t ali_req_p2 : 8;
    uint32_t ali_req_p3 : 8;
} mxc_ioman_ali_req0_t;

typedef struct {
    uint32_t ali_req_p4 : 8;
    uint32_t ali_req_p5 : 8;
    uint32_t ali_req_p6 : 8;
    uint32_t ali_req_p7 : 8;
} mxc_ioman_ali_req1_t;

typedef struct {
    uint32_t ali_ack_p0 : 8;
    uint32_t ali_ack_p1 : 8;
    uint32_t ali_ack_p2 : 8;
    uint32_t ali_ack_p3 : 8;
} mxc_ioman_ali_ack0_t;

typedef struct {
    uint32_t ali_ack_p4 : 8;
    uint32_t ali_ack_p5 : 8;
    uint32_t ali_ack_p6 : 8;
    uint32_t ali_ack_p7 : 8;
} mxc_ioman_ali_ack1_t;

typedef struct {
    uint32_t             : 4;
    uint32_t core_io_req : 1;
    uint32_t             : 3;
    uint32_t ss0_io_req  : 1;
    uint32_t ss1_io_req  : 1;
    uint32_t ss2_io_req  : 1;
    uint32_t             : 1;
    uint32_t quad_io_req : 1;
    uint32_t             : 3;
    uint32_t fast_mode   : 1;
    uint32_t             : 15;
} mxc_ioman_spix_req_t;

typedef struct {
    uint32_t             : 4;
    uint32_t core_io_ack : 1;
    uint32_t             : 3;
    uint32_t ss0_io_ack  : 1;
    uint32_t ss1_io_ack  : 1;
    uint32_t ss2_io_ack  : 1;
    uint32_t             : 1;
    uint32_t quad_io_ack : 1;
    uint32_t             : 3;
    uint32_t fast_mode   : 1;
    uint32_t             : 15;
} mxc_ioman_spix_ack_t;

typedef struct {
    uint32_t io_map     : 1;
    uint32_t cts_map    : 1;
    uint32_t rts_map    : 1;
    uint32_t            : 1;
    uint32_t io_req     : 1;
    uint32_t cts_io_req : 1;
    uint32_t rts_io_req : 1;
    uint32_t            : 25;
} mxc_ioman_uart0_req_t;

typedef struct {
    uint32_t io_map     : 1;
    uint32_t cts_map    : 1;
    uint32_t rts_map    : 1;
    uint32_t            : 1;
    uint32_t io_ack     : 1;
    uint32_t cts_io_ack : 1;
    uint32_t rts_io_ack : 1;
    uint32_t            : 25;
} mxc_ioman_uart0_ack_t;

typedef struct {
    uint32_t io_map     : 1;
    uint32_t cts_map    : 1;
    uint32_t rts_map    : 1;
    uint32_t            : 1;
    uint32_t io_req     : 1;
    uint32_t cts_io_req : 1;
    uint32_t rts_io_req : 1;
    uint32_t            : 25;
} mxc_ioman_uart1_req_t;

typedef struct {
    uint32_t io_map     : 1;
    uint32_t cts_map    : 1;
    uint32_t rts_map    : 1;
    uint32_t            : 1;
    uint32_t io_ack     : 1;
    uint32_t cts_io_ack : 1;
    uint32_t rts_io_ack : 1;
    uint32_t            : 25;
} mxc_ioman_uart1_ack_t;

typedef struct {
    uint32_t io_map     : 1;
    uint32_t cts_map    : 1;
    uint32_t rts_map    : 1;
    uint32_t            : 1;
    uint32_t io_req     : 1;
    uint32_t cts_io_req : 1;
    uint32_t rts_io_req : 1;
    uint32_t            : 25;
} mxc_ioman_uart2_req_t;

typedef struct {
    uint32_t io_map     : 1;
    uint32_t cts_map    : 1;
    uint32_t rts_map    : 1;
    uint32_t            : 1;
    uint32_t io_ack     : 1;
    uint32_t cts_io_ack : 1;
    uint32_t rts_io_ack : 1;
    uint32_t            : 25;
} mxc_ioman_uart2_ack_t;

typedef struct {
    uint32_t io_map     : 1;
    uint32_t cts_map    : 1;
    uint32_t rts_map    : 1;
    uint32_t            : 1;
    uint32_t io_req     : 1;
    uint32_t cts_io_req : 1;
    uint32_t rts_io_req : 1;
    uint32_t            : 25;
} mxc_ioman_uart3_req_t;

typedef struct {
    uint32_t io_map     : 1;
    uint32_t cts_map    : 1;
    uint32_t rts_map    : 1;
    uint32_t            : 1;
    uint32_t io_ack     : 1;
    uint32_t cts_io_ack : 1;
    uint32_t rts_io_ack : 1;
    uint32_t            : 25;
} mxc_ioman_uart3_ack_t;

typedef struct {
    uint32_t             : 4;
    uint32_t core_io_req : 1;
    uint32_t push_pull   : 1;
    uint32_t             : 26;
} mxc_ioman_i2cm0_req_t;

typedef struct {
    uint32_t             : 4;
    uint32_t core_io_ack : 1;
    uint32_t             : 27;
} mxc_ioman_i2cm0_ack_t;

typedef struct {
    uint32_t             : 4;
    uint32_t core_io_req : 1;
    uint32_t push_pull   : 1;
    uint32_t             : 26;
} mxc_ioman_i2cm1_req_t;

typedef struct {
    uint32_t             : 4;
    uint32_t core_io_ack : 1;
    uint32_t             : 27;
} mxc_ioman_i2cm1_ack_t;

typedef struct {
    uint32_t             : 4;
    uint32_t core_io_req : 1;
    uint32_t push_pull   : 1;
    uint32_t             : 26;
} mxc_ioman_i2cm2_req_t;

typedef struct {
    uint32_t             : 4;
    uint32_t core_io_ack : 1;
    uint32_t             : 27;
} mxc_ioman_i2cm2_ack_t;

typedef struct {
    uint32_t mapping_req : 2;
    uint32_t             : 2;
    uint32_t core_io_req : 1;
    uint32_t             : 27;
} mxc_ioman_i2cs_req_t;

typedef struct {
    uint32_t mapping_ack : 2;
    uint32_t             : 2;
    uint32_t core_io_ack : 1;
    uint32_t             : 27;
} mxc_ioman_i2cs_acl_t;

typedef struct {
    uint32_t             : 4;
    uint32_t core_io_req : 1;
    uint32_t             : 3;
    uint32_t ss0_io_req  : 1;
    uint32_t ss1_io_req  : 1;
    uint32_t ss2_io_req  : 1;
    uint32_t ss3_io_req  : 1;
    uint32_t ss4_io_req  : 1;
    uint32_t             : 7;
    uint32_t quad_io_req : 1;
    uint32_t             : 3;
    uint32_t fast_mode   : 1;
    uint32_t             : 7;
} mxc_ioman_spim0_req_t;

typedef struct {
    uint32_t             : 4;
    uint32_t core_io_ack : 1;
    uint32_t             : 3;
    uint32_t ss0_io_ack  : 1;
    uint32_t ss1_io_ack  : 1;
    uint32_t ss2_io_ack  : 1;
    uint32_t ss3_io_ack  : 1;
    uint32_t ss4_io_ack  : 1;
    uint32_t             : 7;
    uint32_t quad_io_ack : 1;
    uint32_t             : 3;
    uint32_t fast_mode   : 1;
    uint32_t             : 7;
} mxc_ioman_spim0_ack_t;

typedef struct {
    uint32_t             : 4;
    uint32_t core_io_req : 1;
    uint32_t             : 3;
    uint32_t ss0_io_req  : 1;
    uint32_t ss1_io_req  : 1;
    uint32_t ss2_io_req  : 1;
    uint32_t             : 9;
    uint32_t quad_io_req : 1;
    uint32_t             : 3;
    uint32_t fast_mode   : 1;
    uint32_t             : 7;
} mxc_ioman_spim1_req_t;

typedef struct {
    uint32_t             : 4;
    uint32_t core_io_ack : 1;
    uint32_t             : 3;
    uint32_t ss0_io_ack  : 1;
    uint32_t ss1_io_ack  : 1;
    uint32_t ss2_io_ack  : 1;
    uint32_t             : 9;
    uint32_t quad_io_ack : 1;
    uint32_t             : 3;
    uint32_t fast_mode   : 1;
    uint32_t             : 7;
} mxc_ioman_spim1_ack_t;

typedef struct {
    uint32_t mapping_req : 2;
    uint32_t             : 2;
    uint32_t core_io_req : 1;
    uint32_t             : 3;
    uint32_t ss0_io_req  : 1;
    uint32_t ss1_io_req  : 1;
    uint32_t ss2_io_req  : 1;
    uint32_t             : 5;
    uint32_t sr0_io_req  : 1;
    uint32_t sr1_io_req  : 1;
    uint32_t             : 2;
    uint32_t quad_io_req : 1;
    uint32_t             : 3;
    uint32_t fast_mode   : 1;
    uint32_t             : 7;
} mxc_ioman_spim2_req_t;

typedef struct {
    uint32_t mapping_ack : 2;
    uint32_t             : 2;
    uint32_t core_io_ack : 1;
    uint32_t             : 3;
    uint32_t ss0_io_ack  : 1;
    uint32_t ss1_io_ack  : 1;
    uint32_t ss2_io_ack  : 1;
    uint32_t             : 5;
    uint32_t sr0_io_req  : 1;
    uint32_t sr1_io_req  : 1;
    uint32_t             : 2;
    uint32_t quad_io_ack : 1;
    uint32_t             : 3;
    uint32_t fast_mode   : 1;
    uint32_t             : 7;
} mxc_ioman_spim2_ack_t;

typedef struct {
    uint32_t             : 4;
    uint32_t core_io_req : 1;
    uint32_t             : 3;
    uint32_t quad_io_req : 1;
    uint32_t             : 3;
    uint32_t fast_mode   : 1;
    uint32_t             : 19;
} mxc_ioman_spib_req_t;

typedef struct {
    uint32_t             : 4;
    uint32_t core_io_ack : 1;
    uint32_t             : 3;
    uint32_t quad_io_ack : 1;
    uint32_t             : 3;
    uint32_t fast_mode   : 1;
    uint32_t             : 19;
} mxc_ioman_spib_ack_t;

typedef struct {
    uint32_t             : 4;
    uint32_t mapping_req : 1;
    uint32_t epu_io_req  : 1;
    uint32_t             : 26;
} mxc_ioman_owm_req_t;

typedef struct {
    uint32_t             : 4;
    uint32_t mapping_ack : 1;
    uint32_t epu_io_ack  : 1;
    uint32_t             : 26;
} mxc_ioman_owm_ack_t;

typedef struct {
    uint32_t mapping_req : 2;
    uint32_t             : 2;
    uint32_t core_io_req : 1;
    uint32_t             : 3;
    uint32_t quad_io_req : 1;
    uint32_t             : 3;
    uint32_t fast_mode   : 1;
    uint32_t             : 19;
} mxc_ioman_spis_req_t;

typedef struct {
    uint32_t mapping_ack : 2;
    uint32_t             : 2;
    uint32_t core_io_ack : 1;
    uint32_t             : 3;
    uint32_t quad_io_ack : 1;
    uint32_t             : 3;
    uint32_t fast_mode   : 1;
    uint32_t             : 19;
} mxc_ioman_spis_ack_t;

typedef struct {
    uint32_t slow_mode     : 1;
    uint32_t alt_rcvr_mode : 1;
    uint32_t               : 30;
} mxc_ioman_pad_mode_t;

typedef struct {
    uint32_t wud_req_p8 : 2;
    uint32_t            : 30;
} mxc_ioman_wud_req2_t;

typedef struct {
    uint32_t wud_ack_p8 : 2;
    uint32_t            : 30;
} mxc_ioman_wud_ack2_t;

typedef struct {
    uint32_t ali_req_p8 : 2;
    uint32_t            : 30;
} mxc_ioman_ali_req2_t;

typedef struct {
    uint32_t ali_ack_p8 : 2;
    uint32_t            : 30;
} mxc_ioman_ali_ack2_t;


/*
   Typedefed structure(s) for module registers (per instance or section) with direct 32-bit
   access to each register in module.
*/

/*                                                          Offset          Register Description
                                                            =============   ============================================================================ */
typedef struct {
    __IO uint32_t wud_req0;                             /*  0x0000          Wakeup Detect Mode Request Register 0 (P0/P1/P2/P3)                          */
    __IO uint32_t wud_req1;                             /*  0x0004          Wakeup Detect Mode Request Register 1 (P4/P5/P6/P7)                          */
    __IO uint32_t wud_ack0;                             /*  0x0008          Wakeup Detect Mode Acknowledge Register 0 (P0/P1/P2/P3)                      */
    __IO uint32_t wud_ack1;                             /*  0x000C          Wakeup Detect Mode Acknowledge Register 1 (P4/P5/P6/P7)                      */
    __IO uint32_t ali_req0;                             /*  0x0010          Analog Input Request Register 0 (P0/P1/P2/P3)                                */
    __IO uint32_t ali_req1;                             /*  0x0014          Analog Input Request Register 1 (P4/P5/P6/P7)                                */
    __IO uint32_t ali_ack0;                             /*  0x0018          Analog Input Acknowledge Register 0 (P0/P1/P2/P3)                            */
    __IO uint32_t ali_ack1;                             /*  0x001C          Analog Input Acknowledge Register 1 (P4/P5/P6/P7)                            */
    __IO uint32_t ali_connect0;                         /*  0x0020          Analog I/O Connection Control Register 0                                     */
    __IO uint32_t ali_connect1;                         /*  0x0024          Analog I/O Connection Control Register 1                                     */
    __IO uint32_t spix_req;                             /*  0x0028          SPIX I/O Mode Request                                                        */
    __IO uint32_t spix_ack;                             /*  0x002C          SPIX I/O Mode Acknowledge                                                    */
    __IO uint32_t uart0_req;                            /*  0x0030          UART0 I/O Mode Request                                                       */
    __IO uint32_t uart0_ack;                            /*  0x0034          UART0 I/O Mode Acknowledge                                                   */
    __IO uint32_t uart1_req;                            /*  0x0038          UART1 I/O Mode Request                                                       */
    __IO uint32_t uart1_ack;                            /*  0x003C          UART1 I/O Mode Acknowledge                                                   */
    __IO uint32_t uart2_req;                            /*  0x0040          UART2 I/O Mode Request                                                       */
    __IO uint32_t uart2_ack;                            /*  0x0044          UART2 I/O Mode Acknowledge                                                   */
    __IO uint32_t uart3_req;                            /*  0x0048          UART3 I/O Mode Request                                                       */
    __IO uint32_t uart3_ack;                            /*  0x004C          UART3 I/O Mode Acknowledge                                                   */
    __IO uint32_t i2cm0_req;                            /*  0x0050          I2C Master 0 I/O Request                                                     */
    __IO uint32_t i2cm0_ack;                            /*  0x0054          I2C Master 0 I/O Acknowledge                                                 */
    __IO uint32_t i2cm1_req;                            /*  0x0058          I2C Master 1 I/O Request                                                     */
    __IO uint32_t i2cm1_ack;                            /*  0x005C          I2C Master 1 I/O Acknowledge                                                 */
    __IO uint32_t i2cm2_req;                            /*  0x0060          I2C Master 2 I/O Request                                                     */
    __IO uint32_t i2cm2_ack;                            /*  0x0064          I2C Master 2 I/O Acknowledge                                                 */
    __IO uint32_t i2cs_req;                             /*  0x0068          I2C Slave I/O Request                                                        */
    __IO uint32_t i2cs_ack;                             /*  0x006C          I2C Slave I/O Acknowledge                                                    */
    __IO uint32_t spim0_req;                            /*  0x0070          SPI Master 0 I/O Mode Request                                                */
    __IO uint32_t spim0_ack;                            /*  0x0074          SPI Master 0 I/O Mode Acknowledge                                            */
    __IO uint32_t spim1_req;                            /*  0x0078          SPI Master 1 I/O Mode Request                                                */
    __IO uint32_t spim1_ack;                            /*  0x007C          SPI Master 1 I/O Mode Acknowledge                                            */
    __IO uint32_t spim2_req;                            /*  0x0080          SPI Master 2 I/O Mode Request                                                */
    __IO uint32_t spim2_ack;                            /*  0x0084          SPI Master 2 I/O Mode Acknowledge                                            */
    __IO uint32_t spib_req;                             /*  0x0088          SPI Bridge I/O Mode Request                                                  */
    __IO uint32_t spib_ack;                             /*  0x008C          SPI Bridge I/O Mode Acknowledge                                              */
    __IO uint32_t owm_req;                              /*  0x0090          1-Wire Master I/O Mode Request                                               */
    __IO uint32_t owm_ack;                              /*  0x0094          1-Wire Master I/O Mode Acknowledge                                           */
    __IO uint32_t spis_req;                             /*  0x0098          SPI Slave I/O Mode Request                                                   */
    __IO uint32_t spis_ack;                             /*  0x009C          SPI Slave I/O Mode Acknowledge                                               */
    __R  uint32_t rsv0A0[24];                           /*  0x00A0-0x00FC                                                                                */
    __IO uint32_t use_vddioh_0;                         /*  0x0100          Enable VDDIOH Register 0                                                     */
    __IO uint32_t use_vddioh_1;                         /*  0x0104          Enable VDDIOH Register 1                                                     */
    __IO uint32_t use_vddioh_2;                         /*  0x0108          Enable VDDIOH Register 2                                                     */
    __R  uint32_t rsv10C;                               /*  0x010C                                                                                       */
    __IO uint32_t pad_mode;                             /*  0x0110          Pad Mode Control Register                                                    */
    __R  uint32_t rsv114[27];                           /*  0x0114-0x017C                                                                                */
    __IO uint32_t wud_req2;                             /*  0x0180          Wakeup Detect Mode Request Register 2 (P8)                                   */
    __R  uint32_t rsv184;                               /*  0x0184                                                                                       */
    __IO uint32_t wud_ack2;                             /*  0x0188          Wakeup Detect Mode Acknowledge Register 2 (P8)                               */
    __R  uint32_t rsv18C;                               /*  0x018C                                                                                       */
    __IO uint32_t ali_req2;                             /*  0x0190          Analog Input Request Register 2 (P8)                                         */
    __R  uint32_t rsv194;                               /*  0x0194                                                                                       */
    __IO uint32_t ali_ack2;                             /*  0x0198          Analog Input Acknowledge Register 2 (P8)                                     */
    __R  uint32_t rsv19C;                               /*  0x019C                                                                                       */
    __IO uint32_t ali_connect2;                         /*  0x01A0          Analog I/O Connection Control Register 2                                     */
} mxc_ioman_regs_t;


/*
   Register offsets for module IOMAN.
*/

#define MXC_R_IOMAN_OFFS_WUD_REQ0                           ((uint32_t)0x00000000UL)
#define MXC_R_IOMAN_OFFS_WUD_REQ1                           ((uint32_t)0x00000004UL)
#define MXC_R_IOMAN_OFFS_WUD_ACK0                           ((uint32_t)0x00000008UL)
#define MXC_R_IOMAN_OFFS_WUD_ACK1                           ((uint32_t)0x0000000CUL)
#define MXC_R_IOMAN_OFFS_ALI_REQ0                           ((uint32_t)0x00000010UL)
#define MXC_R_IOMAN_OFFS_ALI_REQ1                           ((uint32_t)0x00000014UL)
#define MXC_R_IOMAN_OFFS_ALI_ACK0                           ((uint32_t)0x00000018UL)
#define MXC_R_IOMAN_OFFS_ALI_ACK1                           ((uint32_t)0x0000001CUL)
#define MXC_R_IOMAN_OFFS_ALI_CONNECT0                       ((uint32_t)0x00000020UL)
#define MXC_R_IOMAN_OFFS_ALI_CONNECT1                       ((uint32_t)0x00000024UL)
#define MXC_R_IOMAN_OFFS_SPIX_REQ                           ((uint32_t)0x00000028UL)
#define MXC_R_IOMAN_OFFS_SPIX_ACK                           ((uint32_t)0x0000002CUL)
#define MXC_R_IOMAN_OFFS_UART0_REQ                          ((uint32_t)0x00000030UL)
#define MXC_R_IOMAN_OFFS_UART0_ACK                          ((uint32_t)0x00000034UL)
#define MXC_R_IOMAN_OFFS_UART1_REQ                          ((uint32_t)0x00000038UL)
#define MXC_R_IOMAN_OFFS_UART1_ACK                          ((uint32_t)0x0000003CUL)
#define MXC_R_IOMAN_OFFS_UART2_REQ                          ((uint32_t)0x00000040UL)
#define MXC_R_IOMAN_OFFS_UART2_ACK                          ((uint32_t)0x00000044UL)
#define MXC_R_IOMAN_OFFS_UART3_REQ                          ((uint32_t)0x00000048UL)
#define MXC_R_IOMAN_OFFS_UART3_ACK                          ((uint32_t)0x0000004CUL)
#define MXC_R_IOMAN_OFFS_I2CM0_REQ                          ((uint32_t)0x00000050UL)
#define MXC_R_IOMAN_OFFS_I2CM0_ACK                          ((uint32_t)0x00000054UL)
#define MXC_R_IOMAN_OFFS_I2CM1_REQ                          ((uint32_t)0x00000058UL)
#define MXC_R_IOMAN_OFFS_I2CM1_ACK                          ((uint32_t)0x0000005CUL)
#define MXC_R_IOMAN_OFFS_I2CM2_REQ                          ((uint32_t)0x00000060UL)
#define MXC_R_IOMAN_OFFS_I2CM2_ACK                          ((uint32_t)0x00000064UL)
#define MXC_R_IOMAN_OFFS_I2CS_REQ                           ((uint32_t)0x00000068UL)
#define MXC_R_IOMAN_OFFS_I2CS_ACK                           ((uint32_t)0x0000006CUL)
#define MXC_R_IOMAN_OFFS_SPIM0_REQ                          ((uint32_t)0x00000070UL)
#define MXC_R_IOMAN_OFFS_SPIM0_ACK                          ((uint32_t)0x00000074UL)
#define MXC_R_IOMAN_OFFS_SPIM1_REQ                          ((uint32_t)0x00000078UL)
#define MXC_R_IOMAN_OFFS_SPIM1_ACK                          ((uint32_t)0x0000007CUL)
#define MXC_R_IOMAN_OFFS_SPIM2_REQ                          ((uint32_t)0x00000080UL)
#define MXC_R_IOMAN_OFFS_SPIM2_ACK                          ((uint32_t)0x00000084UL)
#define MXC_R_IOMAN_OFFS_SPIB_REQ                           ((uint32_t)0x00000088UL)
#define MXC_R_IOMAN_OFFS_SPIB_ACK                           ((uint32_t)0x0000008CUL)
#define MXC_R_IOMAN_OFFS_OWM_REQ                            ((uint32_t)0x00000090UL)
#define MXC_R_IOMAN_OFFS_OWM_ACK                            ((uint32_t)0x00000094UL)
#define MXC_R_IOMAN_OFFS_SPIS_REQ                           ((uint32_t)0x00000098UL)
#define MXC_R_IOMAN_OFFS_SPIS_ACK                           ((uint32_t)0x0000009CUL)
#define MXC_R_IOMAN_OFFS_USE_VDDIOH_0                       ((uint32_t)0x00000100UL)
#define MXC_R_IOMAN_OFFS_USE_VDDIOH_1                       ((uint32_t)0x00000104UL)
#define MXC_R_IOMAN_OFFS_USE_VDDIOH_2                       ((uint32_t)0x00000108UL)
#define MXC_R_IOMAN_OFFS_PAD_MODE                           ((uint32_t)0x00000110UL)
#define MXC_R_IOMAN_OFFS_WUD_REQ2                           ((uint32_t)0x00000180UL)
#define MXC_R_IOMAN_OFFS_WUD_ACK2                           ((uint32_t)0x00000188UL)
#define MXC_R_IOMAN_OFFS_ALI_REQ2                           ((uint32_t)0x00000190UL)
#define MXC_R_IOMAN_OFFS_ALI_ACK2                           ((uint32_t)0x00000198UL)
#define MXC_R_IOMAN_OFFS_ALI_CONNECT2                       ((uint32_t)0x000001A0UL)


/*
   Field positions and masks for module IOMAN.
*/

#define MXC_F_IOMAN_WUD_REQ0_WUD_REQ_P0_POS                 0
#define MXC_F_IOMAN_WUD_REQ0_WUD_REQ_P0                     ((uint32_t)(0x000000FFUL << MXC_F_IOMAN_WUD_REQ0_WUD_REQ_P0_POS))
#define MXC_F_IOMAN_WUD_REQ0_WUD_REQ_P1_POS                 8
#define MXC_F_IOMAN_WUD_REQ0_WUD_REQ_P1                     ((uint32_t)(0x000000FFUL << MXC_F_IOMAN_WUD_REQ0_WUD_REQ_P1_POS))
#define MXC_F_IOMAN_WUD_REQ0_WUD_REQ_P2_POS                 16
#define MXC_F_IOMAN_WUD_REQ0_WUD_REQ_P2                     ((uint32_t)(0x000000FFUL << MXC_F_IOMAN_WUD_REQ0_WUD_REQ_P2_POS))
#define MXC_F_IOMAN_WUD_REQ0_WUD_REQ_P3_POS                 24
#define MXC_F_IOMAN_WUD_REQ0_WUD_REQ_P3                     ((uint32_t)(0x000000FFUL << MXC_F_IOMAN_WUD_REQ0_WUD_REQ_P3_POS))

#define MXC_F_IOMAN_WUD_REQ1_WUD_REQ_P4_POS                 0
#define MXC_F_IOMAN_WUD_REQ1_WUD_REQ_P4                     ((uint32_t)(0x000000FFUL << MXC_F_IOMAN_WUD_REQ1_WUD_REQ_P4_POS))
#define MXC_F_IOMAN_WUD_REQ1_WUD_REQ_P5_POS                 8
#define MXC_F_IOMAN_WUD_REQ1_WUD_REQ_P5                     ((uint32_t)(0x000000FFUL << MXC_F_IOMAN_WUD_REQ1_WUD_REQ_P5_POS))
#define MXC_F_IOMAN_WUD_REQ1_WUD_REQ_P6_POS                 16
#define MXC_F_IOMAN_WUD_REQ1_WUD_REQ_P6                     ((uint32_t)(0x000000FFUL << MXC_F_IOMAN_WUD_REQ1_WUD_REQ_P6_POS))
#define MXC_F_IOMAN_WUD_REQ1_WUD_REQ_P7_POS                 24
#define MXC_F_IOMAN_WUD_REQ1_WUD_REQ_P7                     ((uint32_t)(0x000000FFUL << MXC_F_IOMAN_WUD_REQ1_WUD_REQ_P7_POS))

#define MXC_F_IOMAN_WUD_ACK0_WUD_ACK_P0_POS                 0
#define MXC_F_IOMAN_WUD_ACK0_WUD_ACK_P0                     ((uint32_t)(0x000000FFUL << MXC_F_IOMAN_WUD_ACK0_WUD_ACK_P0_POS))
#define MXC_F_IOMAN_WUD_ACK0_WUD_ACK_P1_POS                 8
#define MXC_F_IOMAN_WUD_ACK0_WUD_ACK_P1                     ((uint32_t)(0x000000FFUL << MXC_F_IOMAN_WUD_ACK0_WUD_ACK_P1_POS))
#define MXC_F_IOMAN_WUD_ACK0_WUD_ACK_P2_POS                 16
#define MXC_F_IOMAN_WUD_ACK0_WUD_ACK_P2                     ((uint32_t)(0x000000FFUL << MXC_F_IOMAN_WUD_ACK0_WUD_ACK_P2_POS))
#define MXC_F_IOMAN_WUD_ACK0_WUD_ACK_P3_POS                 24
#define MXC_F_IOMAN_WUD_ACK0_WUD_ACK_P3                     ((uint32_t)(0x000000FFUL << MXC_F_IOMAN_WUD_ACK0_WUD_ACK_P3_POS))

#define MXC_F_IOMAN_WUD_ACK1_WUD_ACK_P4_POS                 0
#define MXC_F_IOMAN_WUD_ACK1_WUD_ACK_P4                     ((uint32_t)(0x000000FFUL << MXC_F_IOMAN_WUD_ACK1_WUD_ACK_P4_POS))
#define MXC_F_IOMAN_WUD_ACK1_WUD_ACK_P5_POS                 8
#define MXC_F_IOMAN_WUD_ACK1_WUD_ACK_P5                     ((uint32_t)(0x000000FFUL << MXC_F_IOMAN_WUD_ACK1_WUD_ACK_P5_POS))
#define MXC_F_IOMAN_WUD_ACK1_WUD_ACK_P6_POS                 16
#define MXC_F_IOMAN_WUD_ACK1_WUD_ACK_P6                     ((uint32_t)(0x000000FFUL << MXC_F_IOMAN_WUD_ACK1_WUD_ACK_P6_POS))
#define MXC_F_IOMAN_WUD_ACK1_WUD_ACK_P7_POS                 24
#define MXC_F_IOMAN_WUD_ACK1_WUD_ACK_P7                     ((uint32_t)(0x000000FFUL << MXC_F_IOMAN_WUD_ACK1_WUD_ACK_P7_POS))

#define MXC_F_IOMAN_ALI_REQ0_ALI_REQ_P0_POS                 0
#define MXC_F_IOMAN_ALI_REQ0_ALI_REQ_P0                     ((uint32_t)(0x000000FFUL << MXC_F_IOMAN_ALI_REQ0_ALI_REQ_P0_POS))
#define MXC_F_IOMAN_ALI_REQ0_ALI_REQ_P1_POS                 8
#define MXC_F_IOMAN_ALI_REQ0_ALI_REQ_P1                     ((uint32_t)(0x000000FFUL << MXC_F_IOMAN_ALI_REQ0_ALI_REQ_P1_POS))
#define MXC_F_IOMAN_ALI_REQ0_ALI_REQ_P2_POS                 16
#define MXC_F_IOMAN_ALI_REQ0_ALI_REQ_P2                     ((uint32_t)(0x000000FFUL << MXC_F_IOMAN_ALI_REQ0_ALI_REQ_P2_POS))
#define MXC_F_IOMAN_ALI_REQ0_ALI_REQ_P3_POS                 24
#define MXC_F_IOMAN_ALI_REQ0_ALI_REQ_P3                     ((uint32_t)(0x000000FFUL << MXC_F_IOMAN_ALI_REQ0_ALI_REQ_P3_POS))

#define MXC_F_IOMAN_ALI_REQ1_ALI_REQ_P4_POS                 0
#define MXC_F_IOMAN_ALI_REQ1_ALI_REQ_P4                     ((uint32_t)(0x000000FFUL << MXC_F_IOMAN_ALI_REQ1_ALI_REQ_P4_POS))
#define MXC_F_IOMAN_ALI_REQ1_ALI_REQ_P5_POS                 8
#define MXC_F_IOMAN_ALI_REQ1_ALI_REQ_P5                     ((uint32_t)(0x000000FFUL << MXC_F_IOMAN_ALI_REQ1_ALI_REQ_P5_POS))
#define MXC_F_IOMAN_ALI_REQ1_ALI_REQ_P6_POS                 16
#define MXC_F_IOMAN_ALI_REQ1_ALI_REQ_P6                     ((uint32_t)(0x000000FFUL << MXC_F_IOMAN_ALI_REQ1_ALI_REQ_P6_POS))
#define MXC_F_IOMAN_ALI_REQ1_ALI_REQ_P7_POS                 24
#define MXC_F_IOMAN_ALI_REQ1_ALI_REQ_P7                     ((uint32_t)(0x000000FFUL << MXC_F_IOMAN_ALI_REQ1_ALI_REQ_P7_POS))

#define MXC_F_IOMAN_ALI_ACK0_ALI_ACK_P0_POS                 0
#define MXC_F_IOMAN_ALI_ACK0_ALI_ACK_P0                     ((uint32_t)(0x000000FFUL << MXC_F_IOMAN_ALI_ACK0_ALI_ACK_P0_POS))
#define MXC_F_IOMAN_ALI_ACK0_ALI_ACK_P1_POS                 8
#define MXC_F_IOMAN_ALI_ACK0_ALI_ACK_P1                     ((uint32_t)(0x000000FFUL << MXC_F_IOMAN_ALI_ACK0_ALI_ACK_P1_POS))
#define MXC_F_IOMAN_ALI_ACK0_ALI_ACK_P2_POS                 16
#define MXC_F_IOMAN_ALI_ACK0_ALI_ACK_P2                     ((uint32_t)(0x000000FFUL << MXC_F_IOMAN_ALI_ACK0_ALI_ACK_P2_POS))
#define MXC_F_IOMAN_ALI_ACK0_ALI_ACK_P3_POS                 24
#define MXC_F_IOMAN_ALI_ACK0_ALI_ACK_P3                     ((uint32_t)(0x000000FFUL << MXC_F_IOMAN_ALI_ACK0_ALI_ACK_P3_POS))

#define MXC_F_IOMAN_ALI_ACK1_ALI_ACK_P4_POS                 0
#define MXC_F_IOMAN_ALI_ACK1_ALI_ACK_P4                     ((uint32_t)(0x000000FFUL << MXC_F_IOMAN_ALI_ACK1_ALI_ACK_P4_POS))
#define MXC_F_IOMAN_ALI_ACK1_ALI_ACK_P5_POS                 8
#define MXC_F_IOMAN_ALI_ACK1_ALI_ACK_P5                     ((uint32_t)(0x000000FFUL << MXC_F_IOMAN_ALI_ACK1_ALI_ACK_P5_POS))
#define MXC_F_IOMAN_ALI_ACK1_ALI_ACK_P6_POS                 16
#define MXC_F_IOMAN_ALI_ACK1_ALI_ACK_P6                     ((uint32_t)(0x000000FFUL << MXC_F_IOMAN_ALI_ACK1_ALI_ACK_P6_POS))
#define MXC_F_IOMAN_ALI_ACK1_ALI_ACK_P7_POS                 24
#define MXC_F_IOMAN_ALI_ACK1_ALI_ACK_P7                     ((uint32_t)(0x000000FFUL << MXC_F_IOMAN_ALI_ACK1_ALI_ACK_P7_POS))

#define MXC_F_IOMAN_SPIX_REQ_CORE_IO_REQ_POS                4
#define MXC_F_IOMAN_SPIX_REQ_CORE_IO_REQ                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIX_REQ_CORE_IO_REQ_POS))
#define MXC_F_IOMAN_SPIX_REQ_SS0_IO_REQ_POS                 8
#define MXC_F_IOMAN_SPIX_REQ_SS0_IO_REQ                     ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIX_REQ_SS0_IO_REQ_POS))
#define MXC_F_IOMAN_SPIX_REQ_SS1_IO_REQ_POS                 9
#define MXC_F_IOMAN_SPIX_REQ_SS1_IO_REQ                     ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIX_REQ_SS1_IO_REQ_POS))
#define MXC_F_IOMAN_SPIX_REQ_SS2_IO_REQ_POS                 10
#define MXC_F_IOMAN_SPIX_REQ_SS2_IO_REQ                     ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIX_REQ_SS2_IO_REQ_POS))
#define MXC_F_IOMAN_SPIX_REQ_QUAD_IO_REQ_POS                12
#define MXC_F_IOMAN_SPIX_REQ_QUAD_IO_REQ                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIX_REQ_QUAD_IO_REQ_POS))
#define MXC_F_IOMAN_SPIX_REQ_FAST_MODE_POS                  16
#define MXC_F_IOMAN_SPIX_REQ_FAST_MODE                      ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIX_REQ_FAST_MODE_POS))

#define MXC_F_IOMAN_SPIX_ACK_CORE_IO_ACK_POS                4
#define MXC_F_IOMAN_SPIX_ACK_CORE_IO_ACK                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIX_ACK_CORE_IO_ACK_POS))
#define MXC_F_IOMAN_SPIX_ACK_SS0_IO_ACK_POS                 8
#define MXC_F_IOMAN_SPIX_ACK_SS0_IO_ACK                     ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIX_ACK_SS0_IO_ACK_POS))
#define MXC_F_IOMAN_SPIX_ACK_SS1_IO_ACK_POS                 9
#define MXC_F_IOMAN_SPIX_ACK_SS1_IO_ACK                     ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIX_ACK_SS1_IO_ACK_POS))
#define MXC_F_IOMAN_SPIX_ACK_SS2_IO_ACK_POS                 10
#define MXC_F_IOMAN_SPIX_ACK_SS2_IO_ACK                     ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIX_ACK_SS2_IO_ACK_POS))
#define MXC_F_IOMAN_SPIX_ACK_QUAD_IO_ACK_POS                12
#define MXC_F_IOMAN_SPIX_ACK_QUAD_IO_ACK                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIX_ACK_QUAD_IO_ACK_POS))
#define MXC_F_IOMAN_SPIX_ACK_FAST_MODE_POS                  16
#define MXC_F_IOMAN_SPIX_ACK_FAST_MODE                      ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIX_ACK_FAST_MODE_POS))

#define MXC_F_IOMAN_UART0_REQ_IO_MAP_POS                    0
#define MXC_F_IOMAN_UART0_REQ_IO_MAP                        ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART0_REQ_IO_MAP_POS))
#define MXC_F_IOMAN_UART0_REQ_CTS_MAP_POS                   1
#define MXC_F_IOMAN_UART0_REQ_CTS_MAP                       ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART0_REQ_CTS_MAP_POS))
#define MXC_F_IOMAN_UART0_REQ_RTS_MAP_POS                   2
#define MXC_F_IOMAN_UART0_REQ_RTS_MAP                       ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART0_REQ_RTS_MAP_POS))
#define MXC_F_IOMAN_UART0_REQ_IO_REQ_POS                    4
#define MXC_F_IOMAN_UART0_REQ_IO_REQ                        ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART0_REQ_IO_REQ_POS))
#define MXC_F_IOMAN_UART0_REQ_CTS_IO_REQ_POS                5
#define MXC_F_IOMAN_UART0_REQ_CTS_IO_REQ                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART0_REQ_CTS_IO_REQ_POS))
#define MXC_F_IOMAN_UART0_REQ_RTS_IO_REQ_POS                6
#define MXC_F_IOMAN_UART0_REQ_RTS_IO_REQ                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART0_REQ_RTS_IO_REQ_POS))

#define MXC_F_IOMAN_UART0_ACK_IO_MAP_POS                    0
#define MXC_F_IOMAN_UART0_ACK_IO_MAP                        ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART0_ACK_IO_MAP_POS))
#define MXC_F_IOMAN_UART0_ACK_CTS_MAP_POS                   1
#define MXC_F_IOMAN_UART0_ACK_CTS_MAP                       ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART0_ACK_CTS_MAP_POS))
#define MXC_F_IOMAN_UART0_ACK_RTS_MAP_POS                   2
#define MXC_F_IOMAN_UART0_ACK_RTS_MAP                       ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART0_ACK_RTS_MAP_POS))
#define MXC_F_IOMAN_UART0_ACK_IO_ACK_POS                    4
#define MXC_F_IOMAN_UART0_ACK_IO_ACK                        ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART0_ACK_IO_ACK_POS))
#define MXC_F_IOMAN_UART0_ACK_CTS_IO_ACK_POS                5
#define MXC_F_IOMAN_UART0_ACK_CTS_IO_ACK                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART0_ACK_CTS_IO_ACK_POS))
#define MXC_F_IOMAN_UART0_ACK_RTS_IO_ACK_POS                6
#define MXC_F_IOMAN_UART0_ACK_RTS_IO_ACK                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART0_ACK_RTS_IO_ACK_POS))

#define MXC_F_IOMAN_UART1_REQ_IO_MAP_POS                    0
#define MXC_F_IOMAN_UART1_REQ_IO_MAP                        ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART1_REQ_IO_MAP_POS))
#define MXC_F_IOMAN_UART1_REQ_CTS_MAP_POS                   1
#define MXC_F_IOMAN_UART1_REQ_CTS_MAP                       ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART1_REQ_CTS_MAP_POS))
#define MXC_F_IOMAN_UART1_REQ_RTS_MAP_POS                   2
#define MXC_F_IOMAN_UART1_REQ_RTS_MAP                       ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART1_REQ_RTS_MAP_POS))
#define MXC_F_IOMAN_UART1_REQ_IO_REQ_POS                    4
#define MXC_F_IOMAN_UART1_REQ_IO_REQ                        ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART1_REQ_IO_REQ_POS))
#define MXC_F_IOMAN_UART1_REQ_CTS_IO_REQ_POS                5
#define MXC_F_IOMAN_UART1_REQ_CTS_IO_REQ                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART1_REQ_CTS_IO_REQ_POS))
#define MXC_F_IOMAN_UART1_REQ_RTS_IO_REQ_POS                6
#define MXC_F_IOMAN_UART1_REQ_RTS_IO_REQ                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART1_REQ_RTS_IO_REQ_POS))

#define MXC_F_IOMAN_UART1_ACK_IO_MAP_POS                    0
#define MXC_F_IOMAN_UART1_ACK_IO_MAP                        ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART1_ACK_IO_MAP_POS))
#define MXC_F_IOMAN_UART1_ACK_CTS_MAP_POS                   1
#define MXC_F_IOMAN_UART1_ACK_CTS_MAP                       ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART1_ACK_CTS_MAP_POS))
#define MXC_F_IOMAN_UART1_ACK_RTS_MAP_POS                   2
#define MXC_F_IOMAN_UART1_ACK_RTS_MAP                       ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART1_ACK_RTS_MAP_POS))
#define MXC_F_IOMAN_UART1_ACK_IO_ACK_POS                    4
#define MXC_F_IOMAN_UART1_ACK_IO_ACK                        ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART1_ACK_IO_ACK_POS))
#define MXC_F_IOMAN_UART1_ACK_CTS_IO_ACK_POS                5
#define MXC_F_IOMAN_UART1_ACK_CTS_IO_ACK                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART1_ACK_CTS_IO_ACK_POS))
#define MXC_F_IOMAN_UART1_ACK_RTS_IO_ACK_POS                6
#define MXC_F_IOMAN_UART1_ACK_RTS_IO_ACK                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART1_ACK_RTS_IO_ACK_POS))

#define MXC_F_IOMAN_UART2_REQ_IO_MAP_POS                    0
#define MXC_F_IOMAN_UART2_REQ_IO_MAP                        ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART2_REQ_IO_MAP_POS))
#define MXC_F_IOMAN_UART2_REQ_CTS_MAP_POS                   1
#define MXC_F_IOMAN_UART2_REQ_CTS_MAP                       ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART2_REQ_CTS_MAP_POS))
#define MXC_F_IOMAN_UART2_REQ_RTS_MAP_POS                   2
#define MXC_F_IOMAN_UART2_REQ_RTS_MAP                       ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART2_REQ_RTS_MAP_POS))
#define MXC_F_IOMAN_UART2_REQ_IO_REQ_POS                    4
#define MXC_F_IOMAN_UART2_REQ_IO_REQ                        ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART2_REQ_IO_REQ_POS))
#define MXC_F_IOMAN_UART2_REQ_CTS_IO_REQ_POS                5
#define MXC_F_IOMAN_UART2_REQ_CTS_IO_REQ                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART2_REQ_CTS_IO_REQ_POS))
#define MXC_F_IOMAN_UART2_REQ_RTS_IO_REQ_POS                6
#define MXC_F_IOMAN_UART2_REQ_RTS_IO_REQ                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART2_REQ_RTS_IO_REQ_POS))

#define MXC_F_IOMAN_UART2_ACK_IO_MAP_POS                    0
#define MXC_F_IOMAN_UART2_ACK_IO_MAP                        ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART2_ACK_IO_MAP_POS))
#define MXC_F_IOMAN_UART2_ACK_CTS_MAP_POS                   1
#define MXC_F_IOMAN_UART2_ACK_CTS_MAP                       ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART2_ACK_CTS_MAP_POS))
#define MXC_F_IOMAN_UART2_ACK_RTS_MAP_POS                   2
#define MXC_F_IOMAN_UART2_ACK_RTS_MAP                       ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART2_ACK_RTS_MAP_POS))
#define MXC_F_IOMAN_UART2_ACK_IO_ACK_POS                    4
#define MXC_F_IOMAN_UART2_ACK_IO_ACK                        ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART2_ACK_IO_ACK_POS))
#define MXC_F_IOMAN_UART2_ACK_CTS_IO_ACK_POS                5
#define MXC_F_IOMAN_UART2_ACK_CTS_IO_ACK                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART2_ACK_CTS_IO_ACK_POS))
#define MXC_F_IOMAN_UART2_ACK_RTS_IO_ACK_POS                6
#define MXC_F_IOMAN_UART2_ACK_RTS_IO_ACK                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART2_ACK_RTS_IO_ACK_POS))

#define MXC_F_IOMAN_UART3_REQ_IO_MAP_POS                    0
#define MXC_F_IOMAN_UART3_REQ_IO_MAP                        ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART3_REQ_IO_MAP_POS))
#define MXC_F_IOMAN_UART3_REQ_CTS_MAP_POS                   1
#define MXC_F_IOMAN_UART3_REQ_CTS_MAP                       ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART3_REQ_CTS_MAP_POS))
#define MXC_F_IOMAN_UART3_REQ_RTS_MAP_POS                   2
#define MXC_F_IOMAN_UART3_REQ_RTS_MAP                       ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART3_REQ_RTS_MAP_POS))
#define MXC_F_IOMAN_UART3_REQ_IO_REQ_POS                    4
#define MXC_F_IOMAN_UART3_REQ_IO_REQ                        ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART3_REQ_IO_REQ_POS))
#define MXC_F_IOMAN_UART3_REQ_CTS_IO_REQ_POS                5
#define MXC_F_IOMAN_UART3_REQ_CTS_IO_REQ                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART3_REQ_CTS_IO_REQ_POS))
#define MXC_F_IOMAN_UART3_REQ_RTS_IO_REQ_POS                6
#define MXC_F_IOMAN_UART3_REQ_RTS_IO_REQ                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART3_REQ_RTS_IO_REQ_POS))

#define MXC_F_IOMAN_UART3_ACK_IO_MAP_POS                    0
#define MXC_F_IOMAN_UART3_ACK_IO_MAP                        ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART3_ACK_IO_MAP_POS))
#define MXC_F_IOMAN_UART3_ACK_CTS_MAP_POS                   1
#define MXC_F_IOMAN_UART3_ACK_CTS_MAP                       ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART3_ACK_CTS_MAP_POS))
#define MXC_F_IOMAN_UART3_ACK_RTS_MAP_POS                   2
#define MXC_F_IOMAN_UART3_ACK_RTS_MAP                       ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART3_ACK_RTS_MAP_POS))
#define MXC_F_IOMAN_UART3_ACK_IO_ACK_POS                    4
#define MXC_F_IOMAN_UART3_ACK_IO_ACK                        ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART3_ACK_IO_ACK_POS))
#define MXC_F_IOMAN_UART3_ACK_CTS_IO_ACK_POS                5
#define MXC_F_IOMAN_UART3_ACK_CTS_IO_ACK                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART3_ACK_CTS_IO_ACK_POS))
#define MXC_F_IOMAN_UART3_ACK_RTS_IO_ACK_POS                6
#define MXC_F_IOMAN_UART3_ACK_RTS_IO_ACK                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_UART3_ACK_RTS_IO_ACK_POS))

#define MXC_F_IOMAN_I2CM0_REQ_MAPPING_REQ_POS               4
#define MXC_F_IOMAN_I2CM0_REQ_MAPPING_REQ                   ((uint32_t)(0x00000001UL << MXC_F_IOMAN_I2CM0_REQ_MAPPING_REQ_POS))

#define MXC_F_IOMAN_I2CM0_ACK_MAPPING_ACK_POS               4
#define MXC_F_IOMAN_I2CM0_ACK_MAPPING_ACK                   ((uint32_t)(0x00000001UL << MXC_F_IOMAN_I2CM0_ACK_MAPPING_ACK_POS))

#define MXC_F_IOMAN_I2CM1_REQ_IO_SEL_POS                    0
#define MXC_F_IOMAN_I2CM1_REQ_IO_SEL                        ((uint32_t)(0x00000003UL << MXC_F_IOMAN_I2CM1_REQ_IO_SEL_POS))
#define MXC_F_IOMAN_I2CM1_REQ_MAPPING_REQ_POS               4
#define MXC_F_IOMAN_I2CM1_REQ_MAPPING_REQ                   ((uint32_t)(0x00000001UL << MXC_F_IOMAN_I2CM1_REQ_MAPPING_REQ_POS))

#define MXC_F_IOMAN_I2CM1_ACK_IO_SEL_POS                    0
#define MXC_F_IOMAN_I2CM1_ACK_IO_SEL                        ((uint32_t)(0x00000003UL << MXC_F_IOMAN_I2CM1_ACK_IO_SEL_POS))
#define MXC_F_IOMAN_I2CM1_ACK_MAPPING_ACK_POS               4
#define MXC_F_IOMAN_I2CM1_ACK_MAPPING_ACK                   ((uint32_t)(0x00000001UL << MXC_F_IOMAN_I2CM1_ACK_MAPPING_ACK_POS))

#define MXC_F_IOMAN_I2CM2_REQ_IO_SEL_POS                    0
#define MXC_F_IOMAN_I2CM2_REQ_IO_SEL                        ((uint32_t)(0x00000003UL << MXC_F_IOMAN_I2CM2_REQ_IO_SEL_POS))
#define MXC_F_IOMAN_I2CM2_REQ_MAPPING_REQ_POS               4
#define MXC_F_IOMAN_I2CM2_REQ_MAPPING_REQ                   ((uint32_t)(0x00000001UL << MXC_F_IOMAN_I2CM2_REQ_MAPPING_REQ_POS))

#define MXC_F_IOMAN_I2CM2_ACK_IO_SEL_POS                    0
#define MXC_F_IOMAN_I2CM2_ACK_IO_SEL                        ((uint32_t)(0x00000003UL << MXC_F_IOMAN_I2CM2_ACK_IO_SEL_POS))
#define MXC_F_IOMAN_I2CM2_ACK_MAPPING_ACK_POS               4
#define MXC_F_IOMAN_I2CM2_ACK_MAPPING_ACK                   ((uint32_t)(0x00000001UL << MXC_F_IOMAN_I2CM2_ACK_MAPPING_ACK_POS))

#define MXC_F_IOMAN_I2CS_REQ_IO_SEL_POS                     0
#define MXC_F_IOMAN_I2CS_REQ_IO_SEL                         ((uint32_t)(0x00000007UL << MXC_F_IOMAN_I2CS_REQ_IO_SEL_POS))
#define MXC_F_IOMAN_I2CS_REQ_MAPPING_REQ_POS                4
#define MXC_F_IOMAN_I2CS_REQ_MAPPING_REQ                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_I2CS_REQ_MAPPING_REQ_POS))

#define MXC_F_IOMAN_I2CS_ACK_IO_SEL_POS                     0
#define MXC_F_IOMAN_I2CS_ACK_IO_SEL                         ((uint32_t)(0x00000007UL << MXC_F_IOMAN_I2CS_ACK_IO_SEL_POS))
#define MXC_F_IOMAN_I2CS_ACK_MAPPING_ACK_POS                4
#define MXC_F_IOMAN_I2CS_ACK_MAPPING_ACK                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_I2CS_ACK_MAPPING_ACK_POS))

#define MXC_F_IOMAN_SPIM0_REQ_CORE_IO_REQ_POS               4
#define MXC_F_IOMAN_SPIM0_REQ_CORE_IO_REQ                   ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM0_REQ_CORE_IO_REQ_POS))
#define MXC_F_IOMAN_SPIM0_REQ_SS0_IO_REQ_POS                8
#define MXC_F_IOMAN_SPIM0_REQ_SS0_IO_REQ                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM0_REQ_SS0_IO_REQ_POS))
#define MXC_F_IOMAN_SPIM0_REQ_SS1_IO_REQ_POS                9
#define MXC_F_IOMAN_SPIM0_REQ_SS1_IO_REQ                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM0_REQ_SS1_IO_REQ_POS))
#define MXC_F_IOMAN_SPIM0_REQ_SS2_IO_REQ_POS                10
#define MXC_F_IOMAN_SPIM0_REQ_SS2_IO_REQ                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM0_REQ_SS2_IO_REQ_POS))
#define MXC_F_IOMAN_SPIM0_REQ_SS3_IO_REQ_POS                11
#define MXC_F_IOMAN_SPIM0_REQ_SS3_IO_REQ                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM0_REQ_SS3_IO_REQ_POS))
#define MXC_F_IOMAN_SPIM0_REQ_SS4_IO_REQ_POS                12
#define MXC_F_IOMAN_SPIM0_REQ_SS4_IO_REQ                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM0_REQ_SS4_IO_REQ_POS))
#define MXC_F_IOMAN_SPIM0_REQ_QUAD_IO_REQ_POS               20
#define MXC_F_IOMAN_SPIM0_REQ_QUAD_IO_REQ                   ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM0_REQ_QUAD_IO_REQ_POS))
#define MXC_F_IOMAN_SPIM0_REQ_FAST_MODE_POS                 24
#define MXC_F_IOMAN_SPIM0_REQ_FAST_MODE                     ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM0_REQ_FAST_MODE_POS))

#define MXC_F_IOMAN_SPIM0_ACK_CORE_IO_ACK_POS               4
#define MXC_F_IOMAN_SPIM0_ACK_CORE_IO_ACK                   ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM0_ACK_CORE_IO_ACK_POS))
#define MXC_F_IOMAN_SPIM0_ACK_SS0_IO_ACK_POS                8
#define MXC_F_IOMAN_SPIM0_ACK_SS0_IO_ACK                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM0_ACK_SS0_IO_ACK_POS))
#define MXC_F_IOMAN_SPIM0_ACK_SS1_IO_ACK_POS                9
#define MXC_F_IOMAN_SPIM0_ACK_SS1_IO_ACK                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM0_ACK_SS1_IO_ACK_POS))
#define MXC_F_IOMAN_SPIM0_ACK_SS2_IO_ACK_POS                10
#define MXC_F_IOMAN_SPIM0_ACK_SS2_IO_ACK                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM0_ACK_SS2_IO_ACK_POS))
#define MXC_F_IOMAN_SPIM0_ACK_SS3_IO_ACK_POS                11
#define MXC_F_IOMAN_SPIM0_ACK_SS3_IO_ACK                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM0_ACK_SS3_IO_ACK_POS))
#define MXC_F_IOMAN_SPIM0_ACK_SS4_IO_ACK_POS                12
#define MXC_F_IOMAN_SPIM0_ACK_SS4_IO_ACK                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM0_ACK_SS4_IO_ACK_POS))
#define MXC_F_IOMAN_SPIM0_ACK_QUAD_IO_ACK_POS               20
#define MXC_F_IOMAN_SPIM0_ACK_QUAD_IO_ACK                   ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM0_ACK_QUAD_IO_ACK_POS))
#define MXC_F_IOMAN_SPIM0_ACK_FAST_MODE_POS                 24
#define MXC_F_IOMAN_SPIM0_ACK_FAST_MODE                     ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM0_ACK_FAST_MODE_POS))

#define MXC_F_IOMAN_SPIM1_REQ_CORE_IO_REQ_POS               4
#define MXC_F_IOMAN_SPIM1_REQ_CORE_IO_REQ                   ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM1_REQ_CORE_IO_REQ_POS))
#define MXC_F_IOMAN_SPIM1_REQ_SS0_IO_REQ_POS                8
#define MXC_F_IOMAN_SPIM1_REQ_SS0_IO_REQ                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM1_REQ_SS0_IO_REQ_POS))
#define MXC_F_IOMAN_SPIM1_REQ_SS1_IO_REQ_POS                9
#define MXC_F_IOMAN_SPIM1_REQ_SS1_IO_REQ                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM1_REQ_SS1_IO_REQ_POS))
#define MXC_F_IOMAN_SPIM1_REQ_SS2_IO_REQ_POS                10
#define MXC_F_IOMAN_SPIM1_REQ_SS2_IO_REQ                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM1_REQ_SS2_IO_REQ_POS))
#define MXC_F_IOMAN_SPIM1_REQ_QUAD_IO_REQ_POS               20
#define MXC_F_IOMAN_SPIM1_REQ_QUAD_IO_REQ                   ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM1_REQ_QUAD_IO_REQ_POS))
#define MXC_F_IOMAN_SPIM1_REQ_FAST_MODE_POS                 24
#define MXC_F_IOMAN_SPIM1_REQ_FAST_MODE                     ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM1_REQ_FAST_MODE_POS))

#define MXC_F_IOMAN_SPIM1_ACK_CORE_IO_ACK_POS               4
#define MXC_F_IOMAN_SPIM1_ACK_CORE_IO_ACK                   ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM1_ACK_CORE_IO_ACK_POS))
#define MXC_F_IOMAN_SPIM1_ACK_SS0_IO_ACK_POS                8
#define MXC_F_IOMAN_SPIM1_ACK_SS0_IO_ACK                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM1_ACK_SS0_IO_ACK_POS))
#define MXC_F_IOMAN_SPIM1_ACK_SS1_IO_ACK_POS                9
#define MXC_F_IOMAN_SPIM1_ACK_SS1_IO_ACK                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM1_ACK_SS1_IO_ACK_POS))
#define MXC_F_IOMAN_SPIM1_ACK_SS2_IO_ACK_POS                10
#define MXC_F_IOMAN_SPIM1_ACK_SS2_IO_ACK                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM1_ACK_SS2_IO_ACK_POS))
#define MXC_F_IOMAN_SPIM1_ACK_QUAD_IO_ACK_POS               20
#define MXC_F_IOMAN_SPIM1_ACK_QUAD_IO_ACK                   ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM1_ACK_QUAD_IO_ACK_POS))
#define MXC_F_IOMAN_SPIM1_ACK_FAST_MODE_POS                 24
#define MXC_F_IOMAN_SPIM1_ACK_FAST_MODE                     ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM1_ACK_FAST_MODE_POS))

#define MXC_F_IOMAN_SPIM2_REQ_MAPPING_REQ_POS               0
#define MXC_F_IOMAN_SPIM2_REQ_MAPPING_REQ                   ((uint32_t)(0x00000003UL << MXC_F_IOMAN_SPIM2_REQ_MAPPING_REQ_POS))
#define MXC_F_IOMAN_SPIM2_REQ_CORE_IO_REQ_POS               4
#define MXC_F_IOMAN_SPIM2_REQ_CORE_IO_REQ                   ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM2_REQ_CORE_IO_REQ_POS))
#define MXC_F_IOMAN_SPIM2_REQ_SS0_IO_REQ_POS                8
#define MXC_F_IOMAN_SPIM2_REQ_SS0_IO_REQ                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM2_REQ_SS0_IO_REQ_POS))
#define MXC_F_IOMAN_SPIM2_REQ_SS1_IO_REQ_POS                9
#define MXC_F_IOMAN_SPIM2_REQ_SS1_IO_REQ                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM2_REQ_SS1_IO_REQ_POS))
#define MXC_F_IOMAN_SPIM2_REQ_SS2_IO_REQ_POS                10
#define MXC_F_IOMAN_SPIM2_REQ_SS2_IO_REQ                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM2_REQ_SS2_IO_REQ_POS))
#define MXC_F_IOMAN_SPIM2_REQ_SR0_IO_REQ_POS                16
#define MXC_F_IOMAN_SPIM2_REQ_SR0_IO_REQ                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM2_REQ_SR0_IO_REQ_POS))
#define MXC_F_IOMAN_SPIM2_REQ_SR1_IO_REQ_POS                17
#define MXC_F_IOMAN_SPIM2_REQ_SR1_IO_REQ                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM2_REQ_SR1_IO_REQ_POS))
#define MXC_F_IOMAN_SPIM2_REQ_QUAD_IO_REQ_POS               20
#define MXC_F_IOMAN_SPIM2_REQ_QUAD_IO_REQ                   ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM2_REQ_QUAD_IO_REQ_POS))
#define MXC_F_IOMAN_SPIM2_REQ_FAST_MODE_POS                 24
#define MXC_F_IOMAN_SPIM2_REQ_FAST_MODE                     ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM2_REQ_FAST_MODE_POS))

#define MXC_F_IOMAN_SPIM2_ACK_MAPPING_ACK_POS               0
#define MXC_F_IOMAN_SPIM2_ACK_MAPPING_ACK                   ((uint32_t)(0x00000003UL << MXC_F_IOMAN_SPIM2_ACK_MAPPING_ACK_POS))
#define MXC_F_IOMAN_SPIM2_ACK_CORE_IO_ACK_POS               4
#define MXC_F_IOMAN_SPIM2_ACK_CORE_IO_ACK                   ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM2_ACK_CORE_IO_ACK_POS))
#define MXC_F_IOMAN_SPIM2_ACK_SS0_IO_ACK_POS                8
#define MXC_F_IOMAN_SPIM2_ACK_SS0_IO_ACK                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM2_ACK_SS0_IO_ACK_POS))
#define MXC_F_IOMAN_SPIM2_ACK_SS1_IO_ACK_POS                9
#define MXC_F_IOMAN_SPIM2_ACK_SS1_IO_ACK                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM2_ACK_SS1_IO_ACK_POS))
#define MXC_F_IOMAN_SPIM2_ACK_SS2_IO_ACK_POS                10
#define MXC_F_IOMAN_SPIM2_ACK_SS2_IO_ACK                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM2_ACK_SS2_IO_ACK_POS))
#define MXC_F_IOMAN_SPIM2_ACK_SR0_IO_REQ_POS                16
#define MXC_F_IOMAN_SPIM2_ACK_SR0_IO_REQ                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM2_ACK_SR0_IO_REQ_POS))
#define MXC_F_IOMAN_SPIM2_ACK_SR1_IO_REQ_POS                17
#define MXC_F_IOMAN_SPIM2_ACK_SR1_IO_REQ                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM2_ACK_SR1_IO_REQ_POS))
#define MXC_F_IOMAN_SPIM2_ACK_QUAD_IO_ACK_POS               20
#define MXC_F_IOMAN_SPIM2_ACK_QUAD_IO_ACK                   ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM2_ACK_QUAD_IO_ACK_POS))
#define MXC_F_IOMAN_SPIM2_ACK_FAST_MODE_POS                 24
#define MXC_F_IOMAN_SPIM2_ACK_FAST_MODE                     ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIM2_ACK_FAST_MODE_POS))

#define MXC_F_IOMAN_SPIB_REQ_CORE_IO_REQ_POS                4
#define MXC_F_IOMAN_SPIB_REQ_CORE_IO_REQ                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIB_REQ_CORE_IO_REQ_POS))
#define MXC_F_IOMAN_SPIB_REQ_QUAD_IO_REQ_POS                8
#define MXC_F_IOMAN_SPIB_REQ_QUAD_IO_REQ                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIB_REQ_QUAD_IO_REQ_POS))
#define MXC_F_IOMAN_SPIB_REQ_FAST_MODE_POS                  12
#define MXC_F_IOMAN_SPIB_REQ_FAST_MODE                      ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIB_REQ_FAST_MODE_POS))

#define MXC_F_IOMAN_SPIB_ACK_CORE_IO_ACK_POS                4
#define MXC_F_IOMAN_SPIB_ACK_CORE_IO_ACK                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIB_ACK_CORE_IO_ACK_POS))
#define MXC_F_IOMAN_SPIB_ACK_QUAD_IO_ACK_POS                8
#define MXC_F_IOMAN_SPIB_ACK_QUAD_IO_ACK                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIB_ACK_QUAD_IO_ACK_POS))
#define MXC_F_IOMAN_SPIB_ACK_FAST_MODE_POS                  12
#define MXC_F_IOMAN_SPIB_ACK_FAST_MODE                      ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIB_ACK_FAST_MODE_POS))

#define MXC_F_IOMAN_OWM_REQ_MAPPING_REQ_POS                 4
#define MXC_F_IOMAN_OWM_REQ_MAPPING_REQ                     ((uint32_t)(0x00000001UL << MXC_F_IOMAN_OWM_REQ_MAPPING_REQ_POS))
#define MXC_F_IOMAN_OWM_REQ_EPU_IO_REQ_POS                  5
#define MXC_F_IOMAN_OWM_REQ_EPU_IO_REQ                      ((uint32_t)(0x00000001UL << MXC_F_IOMAN_OWM_REQ_EPU_IO_REQ_POS))

#define MXC_F_IOMAN_OWM_ACK_MAPPING_ACK_POS                 4
#define MXC_F_IOMAN_OWM_ACK_MAPPING_ACK                     ((uint32_t)(0x00000001UL << MXC_F_IOMAN_OWM_ACK_MAPPING_ACK_POS))
#define MXC_F_IOMAN_OWM_ACK_EPU_IO_ACK_POS                  5
#define MXC_F_IOMAN_OWM_ACK_EPU_IO_ACK                      ((uint32_t)(0x00000001UL << MXC_F_IOMAN_OWM_ACK_EPU_IO_ACK_POS))

#define MXC_F_IOMAN_SPIS_REQ_MAPPING_REQ_POS                0
#define MXC_F_IOMAN_SPIS_REQ_MAPPING_REQ                    ((uint32_t)(0x00000003UL << MXC_F_IOMAN_SPIS_REQ_MAPPING_REQ_POS))
#define MXC_F_IOMAN_SPIS_REQ_CORE_IO_REQ_POS                4
#define MXC_F_IOMAN_SPIS_REQ_CORE_IO_REQ                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIS_REQ_CORE_IO_REQ_POS))
#define MXC_F_IOMAN_SPIS_REQ_QUAD_IO_REQ_POS                8
#define MXC_F_IOMAN_SPIS_REQ_QUAD_IO_REQ                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIS_REQ_QUAD_IO_REQ_POS))
#define MXC_F_IOMAN_SPIS_REQ_FAST_MODE_POS                  12
#define MXC_F_IOMAN_SPIS_REQ_FAST_MODE                      ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIS_REQ_FAST_MODE_POS))

#define MXC_F_IOMAN_SPIS_ACK_MAPPING_ACK_POS                0
#define MXC_F_IOMAN_SPIS_ACK_MAPPING_ACK                    ((uint32_t)(0x00000003UL << MXC_F_IOMAN_SPIS_ACK_MAPPING_ACK_POS))
#define MXC_F_IOMAN_SPIS_ACK_CORE_IO_ACK_POS                4
#define MXC_F_IOMAN_SPIS_ACK_CORE_IO_ACK                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIS_ACK_CORE_IO_ACK_POS))
#define MXC_F_IOMAN_SPIS_ACK_QUAD_IO_ACK_POS                8
#define MXC_F_IOMAN_SPIS_ACK_QUAD_IO_ACK                    ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIS_ACK_QUAD_IO_ACK_POS))
#define MXC_F_IOMAN_SPIS_ACK_FAST_MODE_POS                  12
#define MXC_F_IOMAN_SPIS_ACK_FAST_MODE                      ((uint32_t)(0x00000001UL << MXC_F_IOMAN_SPIS_ACK_FAST_MODE_POS))

#define MXC_F_IOMAN_PAD_MODE_SLOW_MODE_POS                  0
#define MXC_F_IOMAN_PAD_MODE_SLOW_MODE                      ((uint32_t)(0x00000001UL << MXC_F_IOMAN_PAD_MODE_SLOW_MODE_POS))
#define MXC_F_IOMAN_PAD_MODE_ALT_RCVR_MODE_POS              1
#define MXC_F_IOMAN_PAD_MODE_ALT_RCVR_MODE                  ((uint32_t)(0x00000001UL << MXC_F_IOMAN_PAD_MODE_ALT_RCVR_MODE_POS))

#define MXC_F_IOMAN_WUD_REQ2_WUD_REQ_P8_POS                 0
#define MXC_F_IOMAN_WUD_REQ2_WUD_REQ_P8                     ((uint32_t)(0x00000003UL << MXC_F_IOMAN_WUD_REQ2_WUD_REQ_P8_POS))

#define MXC_F_IOMAN_WUD_ACK2_WUD_ACK_P8_POS                 0
#define MXC_F_IOMAN_WUD_ACK2_WUD_ACK_P8                     ((uint32_t)(0x00000003UL << MXC_F_IOMAN_WUD_ACK2_WUD_ACK_P8_POS))

#define MXC_F_IOMAN_ALI_REQ2_ALI_REQ_P8_POS                 0
#define MXC_F_IOMAN_ALI_REQ2_ALI_REQ_P8                     ((uint32_t)(0x00000003UL << MXC_F_IOMAN_ALI_REQ2_ALI_REQ_P8_POS))

#define MXC_F_IOMAN_ALI_ACK2_ALI_ACK_P8_POS                 0
#define MXC_F_IOMAN_ALI_ACK2_ALI_ACK_P8                     ((uint32_t)(0x00000003UL << MXC_F_IOMAN_ALI_ACK2_ALI_ACK_P8_POS))



#ifdef __cplusplus
}
#endif

#endif   /* _MXC_IOMAN_REGS_H_ */

