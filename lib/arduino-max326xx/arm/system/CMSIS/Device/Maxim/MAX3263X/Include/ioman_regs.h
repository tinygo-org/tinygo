/**
 * @file   
 * @brief   IOMAN hardware register definitions.
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
 * $Date: 2017-02-16 12:53:53 -0600 (Thu, 16 Feb 2017) $
 * $Revision: 26473 $
 *
 **************************************************************************** */

/* Define to prevent redundant inclusion. */
#ifndef _MXC_IOMAN_REGS_H_
#define _MXC_IOMAN_REGS_H_

/* **** Includes **** */
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

/*
    If types are not defined elsewhere (CMSIS) define them here
*/
///@cond
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
/**
 * @ingroup ioman
 * @defgroup ioman_registers Registers
 * @brief IOMAN Registers
 */
/* **** Definitions **** */
/**
 * @ingroup ioman_registers
 * @defgroup wud_req_ports WUD_REQ_Px
 * Wakeup requests for each port
 */
/**
 * @ingroup ioman_registers
 * @defgroup wud_ack_ports WUD_ACK_Px
 * Wakeup acknowledgement for each port
 */    
/**
 * @ingroup wud_req_ports
 * Structure type for wakeup detection @b request for port 4 to port 7.  
 */
typedef struct {
    uint32_t wud_req_p4 : 8;        /**< Port 4 wake-up detection @b request. */
    uint32_t wud_req_p5 : 8;        /**< Port 5 wake-up detection @b request. */
    uint32_t wud_req_p6 : 8;        /**< Port 6 wake-up detection @b request. */
    uint32_t wud_req_p7 : 8;        /**< Port 7 wake-up detection @b request. */
} mxc_ioman_wud_req1_t;
/**
 * @ingroup wud_ack_ports
 * Structure type for wakeup detection @b acknowledgement for port 0 to port 3.  
 */
typedef struct {
    uint32_t wud_ack_p0 : 8;        /**< Port 0 wake-up detection @b acknowledgement. */
    uint32_t wud_ack_p1 : 8;        /**< Port 1 wake-up detection @b acknowledgement. */
    uint32_t wud_ack_p2 : 8;        /**< Port 2 wake-up detection @b acknowledgement. */
    uint32_t wud_ack_p3 : 8;        /**< Port 3 wake-up detection @b acknowledgement. */
} mxc_ioman_wud_ack0_t;

/**
 * @ingroup wud_req_ports
 * Structure type for wakeup detection @b request for port 0 to port 3.  
 */ 
typedef struct {
    uint32_t wud_req_p0 : 8;        /**< Port 0 wake-up detection @b request. */
    uint32_t wud_req_p1 : 8;        /**< Port 1 wake-up detection @b request. */
    uint32_t wud_req_p2 : 8;        /**< Port 2 wake-up detection @b request. */
    uint32_t wud_req_p3 : 8;        /**< Port 3 wake-up detection @b request. */
} mxc_ioman_wud_req0_t;
/**
 * @ingroup wud_ack_ports
 * Structure type for wakeup detection @b acknowledgement for port 4 to port 7.  
 */
typedef struct {
    uint32_t wud_ack_p4 : 8;        /**< Port 4 wake-up detection @b acknowledgement. */
    uint32_t wud_ack_p5 : 8;        /**< Port 5 wake-up detection @b acknowledgement. */
    uint32_t wud_ack_p6 : 8;        /**< Port 6 wake-up detection @b acknowledgement. */
    uint32_t wud_ack_p7 : 8;        /**< Port 7 wake-up detection @b acknowledgement. */
} mxc_ioman_wud_ack1_t;
/**
 * @ingroup ioman_registers
 * @defgroup ali_req_ports ALI_REQ_Px
 * Analog Input Request per port
 */
/**
 * @ingroup ioman_registers
 * @defgroup ali_ack_ports ALI_ACK_Px
 * Analog Input Acknowledgment per port
 */
/**
 * @ingroup ali_req_ports
 * Structure type for analog input @b request for port 0 to port 3.  
 */
typedef struct {
    uint32_t ali_req_p0 : 8;        /**< Port 0 analog input @b request.  */
    uint32_t ali_req_p1 : 8;        /**< Port 1 analog input @b request.  */
    uint32_t ali_req_p2 : 8;        /**< Port 2 analog input @b request.  */
    uint32_t ali_req_p3 : 8;        /**< Port 3 analog input @b request.  */
} mxc_ioman_ali_req0_t;
/**
 * @ingroup ali_req_ports
 * Structure type for analog input @b request for port 4 to port 7.  
 */
typedef struct {
    uint32_t ali_req_p4 : 8;        /**< Port 4 analog input @b request.  */
    uint32_t ali_req_p5 : 8;        /**< Port 5 analog input @b request.  */
    uint32_t ali_req_p6 : 8;        /**< Port 6 analog input @b request.  */
    uint32_t ali_req_p7 : 8;        /**< Port 7 analog input @b request.  */
} mxc_ioman_ali_req1_t;
/**
 * @ingroup ali_ack_ports
 * Structure type for analog input @b acknowledgement for port 0 to port 3.  
 */
typedef struct {
    uint32_t ali_ack_p0 : 8;        /**< Port 0 analog input @b acknowledgement.  */
    uint32_t ali_ack_p1 : 8;        /**< Port 1 analog input @b acknowledgement.  */
    uint32_t ali_ack_p2 : 8;        /**< Port 2 analog input @b acknowledgement.  */
    uint32_t ali_ack_p3 : 8;        /**< Port 3 analog input @b acknowledgement.  */
} mxc_ioman_ali_ack0_t;
/**
 * @ingroup ali_ack_ports
 * Structure type for analog input @b acknowledgement for port 4 to port 7.  
 */
typedef struct {
    uint32_t ali_ack_p4 : 8;        /**< Port 4 analog input @b acknowledgement.  */
    uint32_t ali_ack_p5 : 8;        /**< Port 5 analog input @b acknowledgement.  */
    uint32_t ali_ack_p6 : 8;        /**< Port 6 analog input @b acknowledgement.  */
    uint32_t ali_ack_p7 : 8;        /**< Port 7 analog input @b acknowledgement.  */
} mxc_ioman_ali_ack1_t;
/**
 * @ingroup ioman_registers
 * @defgroup ioman_periph_config Peripheral Configuration
 * Peripheral I/O Configuration Request and Acknowledment Objects
 * @{ 
 */
/**
 * Structure type for SPI XIP configuration @b requests. 
 */
typedef struct {
    uint32_t             : 4;       /**< Reserved: Do Not Modify.                       */
    uint32_t core_io_req : 1;       /**< Set to request the SPIX core external pins.    */
    uint32_t             : 3;       /**< Reserved: Do Not Modify.                       */
    uint32_t ss0_io_req  : 1;       /**< Set to request slave select 0 active out.      */
    uint32_t ss1_io_req  : 1;       /**< Set to request slave select 1 active out.      */
    uint32_t ss2_io_req  : 1;       /**< Set to request slave select 2 active out.      */
    uint32_t             : 1;       /**< Reserved: Do Not Modify.                       */
    uint32_t quad_io_req : 1;       /**< Set to request quad I/O operation.             */
    uint32_t             : 3;       /**< Reserved: Do Not Modify.                       */
    uint32_t fast_mode   : 1;       /**< Set to request fast mode operation.            */
    uint32_t             : 15;      /**< Reserved: Do Not Modify.                       */
} mxc_ioman_spix_req_t;
/**
 * Structure type for SPI XIP configuration @b acknowledgements. 
 */
typedef struct {
    uint32_t             : 4;       /**< Reserved: Do Not Modify.                                           */
    uint32_t core_io_ack : 1;       /**< Is set if the request for the SPIX core external pins succeeded.   */
    uint32_t             : 3;       /**< Reserved: Do Not Modify.                                           */
    uint32_t ss0_io_ack  : 1;       /**< Is set if the request for the slave select 0 active out succeeded. */
    uint32_t ss1_io_ack  : 1;       /**< Is set if the request for the slave select 1 active out succeeded. */
    uint32_t ss2_io_ack  : 1;       /**< Is set if the request for the slave select 2 active out succeeded. */
    uint32_t             : 1;       /**< Reserved: Do Not Modify.                                           */
    uint32_t quad_io_ack : 1;       /**< Is set if the request for the quad I/O operation succeeded.        */
    uint32_t             : 3;       /**< Reserved: Do Not Modify.                                           */
    uint32_t fast_mode   : 1;       /**< Is set if the request for fast mode operation succeeded.           */
    uint32_t             : 15;      /**< Reserved: Do Not Modify.                                           */
} mxc_ioman_spix_ack_t;
/**
 * Structure type for UART0 configuration @b requests. 
 */
typedef struct {
    uint32_t io_map     : 1;        /**< Value for the desired pin mapping for the RX/TX pins.          */
    uint32_t cts_map    : 1;        /**< Value for the desired pin mapping for the CTS pin.             */
    uint32_t rts_map    : 1;        /**< Value for the desired pin mapping for the RTS pin.             */
    uint32_t            : 1;        /**< Reserved: Do No Modify.                                        */
    uint32_t io_req     : 1;        /**< RX/TX pin mapping, set to @p 1 to request or @p 0 to release.  */
    uint32_t cts_io_req : 1;        /**< CTS pin mapping, set to @p 1 to request or @p 0 to release.    */
    uint32_t rts_io_req : 1;        /**< RTS pin mapping, set to @p 1 to request or @p 0 to release.    */
    uint32_t            : 25;       /**< Reserved: Do No Modify.                                        */
} mxc_ioman_uart0_req_t;
/**
 * Structure type for UART0 configuration @b acknowledgements. 
 */
typedef struct {
    uint32_t io_map     : 1;        /**< Is set to @p 1 if request successful.  */
    uint32_t cts_map    : 1;        /**< Is set to @p 1 if request successful.  */
    uint32_t rts_map    : 1;        /**< Is set to @p 1 if request successful.  */
    uint32_t            : 1;        /**< Reserved: Do No Modify.                */
    uint32_t io_ack     : 1;        /**< Is set to @p 1 if request successful.  */
    uint32_t cts_io_ack : 1;        /**< Is set to @p 1 if request successful.  */
    uint32_t rts_io_ack : 1;        /**< Is set to @p 1 if request successful.  */
    uint32_t            : 25;       /**< Reserved: Do No Modify.                */
} mxc_ioman_uart0_ack_t;
/**
 * Structure type for UART1 configuration @b requests.  
 */
typedef struct {
    uint32_t io_map     : 1;        /**< Value for the desired pin mapping for the RX/TX pins.          */
    uint32_t cts_map    : 1;        /**< Value for the desired pin mapping for the CTS pin.             */
    uint32_t rts_map    : 1;        /**< Value for the desired pin mapping for the RTS pin.             */
    uint32_t            : 1;        /**< Reserved: Do No Modify.                                        */
    uint32_t io_req     : 1;        /**< RX/TX pin mapping, set to @p 1 to request or @p 0 to release.  */
    uint32_t cts_io_req : 1;        /**< CTS pin mapping, set to @p 1 to request or @p 0 to release.    */
    uint32_t rts_io_req : 1;        /**< RTS pin mapping, set to @p 1 to request or @p 0 to release.    */
    uint32_t            : 25;       /**< Reserved: Do No Modify.                                        */
} mxc_ioman_uart1_req_t;
/**
 * Structure type for UART1 configuration @b acknowledgements. 
 */
typedef struct {
    uint32_t io_map     : 1;        /**< Is set to @p 1 if request successful.  */
    uint32_t cts_map    : 1;        /**< Is set to @p 1 if request successful.  */
    uint32_t rts_map    : 1;        /**< Is set to @p 1 if request successful.  */
    uint32_t            : 1;        /**< Reserved: Do No Modify.                */
    uint32_t io_ack     : 1;        /**< Is set to @p 1 if request successful.  */
    uint32_t cts_io_ack : 1;        /**< Is set to @p 1 if request successful.  */
    uint32_t rts_io_ack : 1;        /**< Is set to @p 1 if request successful.  */
    uint32_t            : 25;       /**< Reserved: Do No Modify.                */
} mxc_ioman_uart1_ack_t;
/**
 * Structure type for UART2 configuration @b requests.  
 */
typedef struct {
    uint32_t io_map     : 1;        /**< Value for the desired pin mapping for the RX/TX pins.          */
    uint32_t cts_map    : 1;        /**< Value for the desired pin mapping for the CTS pin.             */
    uint32_t rts_map    : 1;        /**< Value for the desired pin mapping for the RTS pin.             */
    uint32_t            : 1;        /**< Reserved: Do No Modify.                                        */
    uint32_t io_req     : 1;        /**< RX/TX pin mapping, set to @p 1 to request or @p 0 to release.  */
    uint32_t cts_io_req : 1;        /**< CTS pin mapping, set to @p 1 to request or @p 0 to release.    */
    uint32_t rts_io_req : 1;        /**< RTS pin mapping, set to @p 1 to request or @p 0 to release.    */
    uint32_t            : 25;       /**< Reserved: Do No Modify.                                        */
} mxc_ioman_uart2_req_t;
/**
 * Structure type for UART2 configuration @b acknowledgements. 
 */
typedef struct {
    uint32_t io_map     : 1;        /**< Is set to @p 1 if request successful.  */
    uint32_t cts_map    : 1;        /**< Is set to @p 1 if request successful.  */
    uint32_t rts_map    : 1;        /**< Is set to @p 1 if request successful.  */
    uint32_t            : 1;        /**< Reserved: Do No Modify.                */
    uint32_t io_ack     : 1;        /**< Is set to @p 1 if request successful.  */
    uint32_t cts_io_ack : 1;        /**< Is set to @p 1 if request successful.  */
    uint32_t rts_io_ack : 1;        /**< Is set to @p 1 if request successful.  */
    uint32_t            : 25;       /**< Reserved: Do No Modify.                */
} mxc_ioman_uart2_ack_t;
/**
 * Structure type for UART3 configuration @b requests.  
 */
typedef struct {
    uint32_t io_map     : 1;        /**< Value for the desired pin mapping for the RX/TX pins.          */
    uint32_t cts_map    : 1;        /**< Value for the desired pin mapping for the CTS pin.             */
    uint32_t rts_map    : 1;        /**< Value for the desired pin mapping for the RTS pin.             */
    uint32_t            : 1;        /**< Reserved: Do No Modify.                                        */
    uint32_t io_req     : 1;        /**< RX/TX pin mapping, set to @p 1 to request or @p 0 to release.  */
    uint32_t cts_io_req : 1;        /**< CTS pin mapping, set to @p 1 to request or @p 0 to release.    */
    uint32_t rts_io_req : 1;        /**< RTS pin mapping, set to @p 1 to request or @p 0 to release.    */
    uint32_t            : 25;       /**< Reserved: Do No Modify.                */
} mxc_ioman_uart3_req_t;
/**
 * Structure type for UART3 configuration @b acknowledgements. 
 */
typedef struct {
    uint32_t io_map     : 1;        /**< Is set to @p 1 if request successful.  */
    uint32_t cts_map    : 1;        /**< Is set to @p 1 if request successful.  */
    uint32_t rts_map    : 1;        /**< Is set to @p 1 if request successful.  */
    uint32_t            : 1;        /**< Reserved: Do No Modify.                */
    uint32_t io_ack     : 1;        /**< Is set to @p 1 if request successful.  */
    uint32_t cts_io_ack : 1;        /**< Is set to @p 1 if request successful.  */
    uint32_t rts_io_ack : 1;        /**< Is set to @p 1 if request successful.  */
    uint32_t            : 25;       /**< Reserved: Do No Modify.                */
} mxc_ioman_uart3_ack_t;
/**
 * Structure type for I2C Master 0 configuration @b requests. 
 */
typedef struct {
    uint32_t             : 4;       /**< Reserved: Do No Modify.                                */
    uint32_t mapping_req : 1;       /**< Value for the desired pin mapping for the I2CM0 pins.  */       
    uint32_t             : 27;      /**< Reserved: Do No Modify.                                */
} mxc_ioman_i2cm0_req_t;
/**
 * Structure type for I2C Master 0 configuration @b acknowledgements. 
 */
typedef struct {
    uint32_t             : 4;       /**< Reserved: Do No Modify.                                */
    uint32_t mapping_ack : 1;       /**< Is set to @p 1 if request successful.                  */
    uint32_t             : 27;      /**< Reserved: Do No Modify.                                */
} mxc_ioman_i2cm0_ack_t;
/**
 * Structure type for I2C Master 1 configuration @b requests. 
 */
typedef struct {
    uint32_t io_sel      : 2;      /**< Value for the desired pin mapping for the I2CM1 CLK and Data pins.  */
    uint32_t             : 2;      /**< Reserved: Do No Modify.                                             */
    uint32_t mapping_req : 1;      /**< Value for the desired pin mapping for the I2CM1 pins.               */
    uint32_t             : 27;     /**< Reserved: Do No Modify.                                             */
} mxc_ioman_i2cm1_req_t;
/**
 * Structure type for I2C Master 1 configuration @b acknowledgements. 
 */
typedef struct {
    uint32_t io_sel      : 2;      /**< Non-zero if mapping request successful.                */
    uint32_t             : 2;      /**< Reserved: Do No Modify.                                */
    uint32_t mapping_ack : 1;      /**< Is set to @p 1 if request successful.                  */
    uint32_t             : 27;     /**< Reserved: Do No Modify.                                */
} mxc_ioman_i2cm1_ack_t;
/**
 * Structure type for I2C Master 2 configuration @b requests. 
 */
typedef struct {
    uint32_t io_sel      : 2;      /**< Value for the desired pin mapping for the I2CM2 CLK and Data pins.  */
    uint32_t             : 2;      /**< Reserved: Do No Modify.                                             */
    uint32_t mapping_req : 1;      /**< Value for the desired pin mapping for the I2CM2 pins.               */
    uint32_t             : 27;     /**< Reserved: Do No Modify.                                             */
} mxc_ioman_i2cm2_req_t;
/**
 * Structure type for I2C Master 2 configuration @b acknowledgements. 
 */
typedef struct {
    uint32_t io_sel      : 2;      /**< Non-zero if mapping request successful.                */
    uint32_t             : 2;      /**< Reserved: Do No Modify.                                */
    uint32_t mapping_ack : 1;      /**< Is set to @p 1 if request successful.                  */
    uint32_t             : 27;     /**< Reserved: Do No Modify.                                */
} mxc_ioman_i2cm2_ack_t;
/**
 * Structure type for I2C Slave 0 configuration @b requests. 
 */
typedef struct {
    uint32_t io_sel      : 3;      /**< Value for the desired pin mapping for the I2CS0 CLK and Data pins.  */
    uint32_t             : 1;      /**< Reserved: Do No Modify.                                             */
    uint32_t mapping_req : 1;      /**< Value for the desired pin mapping for the I2CS0 pins.               */
    uint32_t             : 27;     /**< Reserved: Do No Modify.                                             */
} mxc_ioman_i2cs_req_t;
/**
 * Structure type for I2C Slave 0 configuration @b acknowledgements. 
 */
typedef struct {
    uint32_t io_sel      : 3;      /**< Non-zero if mapping request successful.                */
    uint32_t             : 1;      /**< Reserved: Do No Modify.                                */
    uint32_t mapping_ack : 1;      /**< Is set to @p 1 if request successful.                  */
    uint32_t             : 27;     /**< Reserved: Do No Modify.                                */
} mxc_ioman_i2cs_ack_t;
/**
 * Structure type for SPI Master 0 configuration @b requests. 
 */
typedef struct {
    uint32_t             : 4;       /**< Reserved: Do Not Modify.                       */
    uint32_t core_io_req : 1;       /**< Set to request the SPIM0 core external pins.   */
    uint32_t             : 3;       /**< Reserved: Do Not Modify.                       */
    uint32_t ss0_io_req  : 1;       /**< Set to request slave select 0 active out.      */
    uint32_t ss1_io_req  : 1;       /**< Set to request slave select 1 active out.      */
    uint32_t ss2_io_req  : 1;       /**< Set to request slave select 2 active out.      */
    uint32_t ss3_io_req  : 1;       /**< Set to request slave select 3 active out.      */
    uint32_t ss4_io_req  : 1;       /**< Set to request slave select 4 active out.      */
    uint32_t             : 7;       /**< Reserved: Do Not Modify.                       */
    uint32_t quad_io_req : 1;       /**< Set to 1 to request Quad I/O for SPIM0.        */
    uint32_t             : 3;       /**< Reserved: Do Not Modify.                       */
    uint32_t fast_mode   : 1;       /**< Set to request fast mode operation for SPIM0.  */
    uint32_t             : 7;       /**< Reserved: Do Not Modify.                       */
} mxc_ioman_spim0_req_t;
/**
 * Structure type for SPI Master 0 configuration @b acknowledgements. 
 */
typedef struct {
    uint32_t             : 4;       /**< Reserved: Do Not Modify.                                           */
    uint32_t core_io_ack : 1;       /**< Is set if the request for the SPIM0 core external pins succeeded.  */
    uint32_t             : 3;       /**< Reserved: Do Not Modify.                                           */
    uint32_t ss0_io_ack  : 1;       /**< Is set if the request for the slave select 0 active out succeeded. */
    uint32_t ss1_io_ack  : 1;       /**< Is set if the request for the slave select 1 active out succeeded. */
    uint32_t ss2_io_ack  : 1;       /**< Is set if the request for the slave select 2 active out succeeded. */
    uint32_t ss3_io_ack  : 1;       /**< Is set if the request for the slave select 3 active out succeeded. */
    uint32_t ss4_io_ack  : 1;       /**< Is set if the request for the slave select 4 active out succeeded. */
    uint32_t             : 7;       /**< Reserved: Do Not Modify.                                           */
    uint32_t quad_io_ack : 1;       /**< Is set if the request for the quad I/O operation succeeded.        */
    uint32_t             : 3;       /**< Reserved: Do Not Modify.                                           */
    uint32_t fast_mode   : 1;       /**< Is set if the request for fast mode operation succeeded.           */
    uint32_t             : 7;       /**< Reserved: Do Not Modify.                                           */
} mxc_ioman_spim0_ack_t;
/**
 * Structure type for SPI Master 1 configuration @b requests. 
 */
typedef struct {
    uint32_t             : 4;      /**< Reserved: Do Not Modify.                       */
    uint32_t core_io_req : 1;      /**< Set to request the SPIM1 core external pins.   */
    uint32_t             : 3;      /**< Reserved: Do Not Modify.                       */
    uint32_t ss0_io_req  : 1;      /**< Set to request slave select 0 active out.      */
    uint32_t ss1_io_req  : 1;      /**< Set to request slave select 1 active out.      */
    uint32_t ss2_io_req  : 1;      /**< Set to request slave select 2 active out.      */
    uint32_t             : 9;      /**< Reserved: Do Not Modify.                       */
    uint32_t quad_io_req : 1;      /**< Set to request quad I/O operation.             */
    uint32_t             : 3;      /**< Reserved: Do Not Modify.                       */
    uint32_t fast_mode   : 1;      /**< Set to request fast mode operation.            */
    uint32_t             : 7;      /**< Reserved: Do Not Modify.                       */
} mxc_ioman_spim1_req_t;
/**
 * Structure type for SPI Master 1 configuration @b acknowledgements. 
 */
typedef struct {
    uint32_t             : 4;      /**< Reserved: Do Not Modify.                                           */
    uint32_t core_io_ack : 1;      /**< Is set if the request for the SPIM1 core external pins succeeded.  */
    uint32_t             : 3;      /**< Reserved: Do Not Modify.                                           */
    uint32_t ss0_io_ack  : 1;      /**< Is set if the request for the slave select 0 active out succeeded. */
    uint32_t ss1_io_ack  : 1;      /**< Is set if the request for the slave select 1 active out succeeded. */
    uint32_t ss2_io_ack  : 1;      /**< Is set if the request for the slave select 2 active out succeeded. */
    uint32_t             : 9;      /**< Reserved: Do Not Modify.                                           */
    uint32_t quad_io_ack : 1;      /**< Is set if the request for the quad I/O operation succeeded.        */
    uint32_t             : 3;      /**< Reserved: Do Not Modify.                                           */
    uint32_t fast_mode   : 1;      /**< Is set if the request for fast mode operation succeeded.           */
    uint32_t             : 7;      /**< Reserved: Do Not Modify.                                           */
} mxc_ioman_spim1_ack_t;
/**
 * Structure type for SPI Master 2 configuration @b requests. 
 */
typedef struct {
    uint32_t mapping_req : 2;      /**< Set to the desired port pin mapping for the SPIM2.                  */
    uint32_t             : 2;      /**< Reserved: Do Not Modify.                                            */
    uint32_t core_io_req : 1;      /**< Set to request the SPIM2 core external pins.                        */
    uint32_t             : 3;      /**< Reserved: Do Not Modify.                                            */
    uint32_t ss0_io_req  : 1;      /**< Set to request slave select 0 active out.                           */
    uint32_t ss1_io_req  : 1;      /**< Set to request slave select 1 active out.                           */
    uint32_t ss2_io_req  : 1;      /**< Set to request slave select 2 active out.                           */
    uint32_t             : 5;      /**< Reserved: Do Not Modify.                                            */
    uint32_t sr0_io_req  : 1;      /**< Set to 1 to request slave ready 0 input.                            */
    uint32_t sr1_io_req  : 1;      /**< Set to 1 to request slave ready 1 input.                            */
    uint32_t             : 2;      /**< Reserved: Do Not Modify.                                            */
    uint32_t quad_io_req : 1;      /**< Set to request quad I/O operation.                                  */
    uint32_t             : 3;      /**< Reserved: Do Not Modify.                                            */
    uint32_t fast_mode   : 1;      /**< Set to request fast mode operation.                                 */
    uint32_t             : 7;      /**< Reserved: Do Not Modify.                                            */
} mxc_ioman_spim2_req_t;
/**
 * Structure type for SPI Master 2 configuration @b acknowledgements. 
 */
typedef struct {
    uint32_t mapping_ack : 2;      /**< Non-zero if mapping request successful.                             */
    uint32_t             : 2;      /**< Reserved: Do Not Modify.                                            */
    uint32_t core_io_ack : 1;      /**< Is set if the request for the SPIM2 core external pins succeeded.   */
    uint32_t             : 3;      /**< Reserved: Do Not Modify.                                            */
    uint32_t ss0_io_ack  : 1;      /**< Is set if the request for the slave select 0 active out succeeded.  */
    uint32_t ss1_io_ack  : 1;      /**< Is set if the request for the slave select 1 active out succeeded.  */
    uint32_t ss2_io_ack  : 1;      /**< Is set if the request for the slave select 2 active out succeeded.  */
    uint32_t             : 5;      /**< Reserved: Do Not Modify.                                            */
    uint32_t sr0_io_req  : 1;      /**< Is set if the request for the slave ready 0 active input succeeded. */
    uint32_t sr1_io_req  : 1;      /**< Is set if the request for the slave ready 1 active input succeeded. */
    uint32_t             : 2;      /**< Reserved: Do Not Modify.                                            */
    uint32_t quad_io_ack : 1;      /**< Is set if the request for the quad I/O operation succeeded.        */
    uint32_t             : 3;      /**< Reserved: Do Not Modify.                                           */
    uint32_t fast_mode   : 1;      /**< Is set if the request for fast mode operation succeeded.           */
    uint32_t             : 7;      /**< Reserved: Do Not Modify.                                            */
} mxc_ioman_spim2_ack_t;
/**
 * Structure type for SPI Bridge configuration @b requests. 
 */
typedef struct {
    uint32_t             : 4;      /**< Reserved: Do Not Modify.                                            */
    uint32_t core_io_req : 1;      /**< Set to request the SPIB core external pins.                         */
    uint32_t             : 3;      /**< Reserved: Do Not Modify.                                            */
    uint32_t quad_io_req : 1;      /**< Set to request quad I/O operation.                                  */
    uint32_t             : 3;      /**< Reserved: Do Not Modify.                                            */
    uint32_t fast_mode   : 1;      /**< Set to request fast mode operation.                                 */
    uint32_t             : 19;     /**< Reserved: Do Not Modify.                                            */
} mxc_ioman_spib_req_t;
/**
 * Structure type for SPI Bridge configuration @b acknowledgements. 
 */
typedef struct {
    uint32_t             : 4;      /**< Reserved: Do Not Modify.                                            */
    uint32_t core_io_ack : 1;      /**< Non-zero if mapping request successful.                             */
    uint32_t             : 3;      /**< Reserved: Do Not Modify.                                            */
    uint32_t quad_io_ack : 1;      /**< Is set if the request for the quad I/O operation succeeded.         */
    uint32_t             : 3;      /**< Reserved: Do Not Modify.                                            */
    uint32_t fast_mode   : 1;      /**< Is set if the request for fast mode operation succeeded.            */
    uint32_t             : 19;     /**< Reserved: Do Not Modify.                                            */
} mxc_ioman_spib_ack_t;
/**
 * Structure type for 1-Wire Master (OWM) configuration @b requests. 
 */
typedef struct {
    uint32_t             : 4;      /**< Reserved: Do Not Modify.                                            */
    uint32_t mapping_req : 1;      /**< Set to the desired port pin mapping for the 1-Wire Master.          */
    uint32_t epu_io_req  : 1;      /**< Set to 1 to request External Pull-up for the 1-Wire Master.         */
    uint32_t             : 26;     /**< Reserved: Do Not Modify.                                            */
} mxc_ioman_owm_req_t;
/**
 * Structure type for 1-Wire Master configuration @b acknowledgements. 
 */
typedef struct {
    uint32_t             : 4;      /**< Reserved: Do Not Modify.                                            */
    uint32_t mapping_ack : 1;      /**< Non-zero if mapping request successful.                             */
    uint32_t epu_io_ack  : 1;      /**< Non-zero if external pull-up request successful.                    */
    uint32_t             : 26;     /**< Reserved: Do Not Modify.                                            */
} mxc_ioman_owm_ack_t;
/**
 * Structure type for SPI Slave configuration @b requests. 
 */
typedef struct {
    uint32_t mapping_req : 2;      /**< Set to desired port pin mapping for the SPIS peripheral.            */
    uint32_t             : 2;      /**< Reserved: Do Not Modify.                                            */
    uint32_t core_io_req : 1;      /**< Set to 1 to request the I/O be assigned to the SPIS.                */
    uint32_t             : 3;      /**< Reserved: Do Not Modify.                                            */
    uint32_t quad_io_req : 1;      /**< Set to request quad I/O operation.                                  */
    uint32_t             : 3;      /**< Reserved: Do Not Modify.                                            */
    uint32_t fast_mode   : 1;      /**< Set to request fast mode operation.                                 */
    uint32_t             : 19;     /**< Reserved: Do Not Modify.                                            */
} mxc_ioman_spis_req_t;
/**
 * Structure type for SPI Slave configuration @b acknowledgements. 
 */
typedef struct {
    uint32_t mapping_ack : 2;      /**< Non-zero if mapping request successful.                             */
    uint32_t             : 2;      /**< Reserved: Do Not Modify.                                            */
    uint32_t core_io_ack : 1;      /**< Non-zero if core io request successful.                             */
    uint32_t             : 3;      /**< Reserved: Do Not Modify.                                            */
    uint32_t quad_io_ack : 1;      /**< Is set if the request for the quad I/O operation succeeded.         */
    uint32_t             : 3;      /**< Reserved: Do Not Modify.                                            */
    uint32_t fast_mode   : 1;      /**< Is set if the request for fast mode operation succeeded.            */
    uint32_t             : 19;     /**< Reserved: Do Not Modify.                                            */
} mxc_ioman_spis_ack_t;
/**@} end of defgroup ioman_periph_config */
/**
 * @ingroup ioman_registers
 * @defgroup ioman_pad_mode I/O Pad Modes
 * Configuration object to set the I/O pad operation for an I/O pin
 */
/**
 * @ingroup ioman_pad_mode
 * Structure type to configure the I/O pad mode options.
 */
typedef struct {
    uint32_t slow_mode     : 1;     /**< Slow mode I/O operation */
    uint32_t alt_rcvr_mode : 1;     /**< Alternative receive mode. */
    uint32_t               : 30;    /**< Reserved: Do not modify. */
} mxc_ioman_pad_mode_t;
/**
 * @ingroup wud_req_ports
 * Structure type for Wake-Up Detect (WUD) configuration @b requests for Port 8. 
 */
typedef struct {
    uint32_t wud_req_p8 : 2;        /**< Request bits for Wakeup Detect Mode requests for port P8. */
    uint32_t            : 30;       /**< Reserved: Do not modify. */
} mxc_ioman_wud_req2_t;
/**
 * @ingroup wud_ack_ports
 * Structure type for Wake-Up Detect (WUD) configuration @b acknowledgements for Port 8. 
 */
typedef struct {
    uint32_t wud_ack_p8 : 2;    /**< Acknowledgement bits for Wakeup Detect Mode requests for port P8. */
    uint32_t            : 30;   /**< Reserved: Do not modify. */
} mxc_ioman_wud_ack2_t;
/**
 * @ingroup ali_req_ports
 * Structure type for analog input @b request for port 8.  
 */
typedef struct {
    uint32_t ali_req_p8 : 2;    /**< Request bits for Analog Wakeup Detect Mode requests for port P8. */
    uint32_t            : 30;   /**< Reserved: Do not modify. */
} mxc_ioman_ali_req2_t;
/**
 * @ingroup ali_ack_ports
 * Structure type for Analog Input Acknowledgement for Port 8. 
 */
typedef struct {
    uint32_t ali_ack_p8 : 2;   /**< Acknowledgement bits for Analog Wakeup Detect Mode requests for port P8. */
    uint32_t            : 30;  /**< Reserved: Do not modify. */
} mxc_ioman_ali_ack2_t;


/**
 * @ingroup ioman_registers
 * Structure type for the IOMAN Register Interface.
 * The table below shows the IOMAN Regsiter Offsets from the Base IOMAN Peripheral Address #MXC_BASE_IOMAN.
 */
typedef struct {
    __IO uint32_t wud_req0;                             /**< Wakeup Detect Mode Request Register 0 (P0/P1/P2/P3)        */
    __IO uint32_t wud_req1;                             /**< Wakeup Detect Mode Request Register 1 (P4/P5/P6/P7)        */
    __IO uint32_t wud_ack0;                             /**< Wakeup Detect Mode Acknowledge Register 0 (P0/P1/P2/P3)    */
    __IO uint32_t wud_ack1;                             /**< Wakeup Detect Mode Acknowledge Register 1 (P4/P5/P6/P7)    */
    __IO uint32_t ali_req0;                             /**< Analog Input Request Register 0 (P0/P1/P2/P3)              */
    __IO uint32_t ali_req1;                             /**< Analog Input Request Register 1 (P4/P5/P6/P7)              */
    __IO uint32_t ali_ack0;                             /**< Analog Input Acknowledge Register 0 (P0/P1/P2/P3)          */
    __IO uint32_t ali_ack1;                             /**< Analog Input Acknowledge Register 1 (P4/P5/P6/P7)          */
    __IO uint32_t ali_connect0;                         /**< Analog I/O Connection Control Register 0                   */
    __IO uint32_t ali_connect1;                         /**< Analog I/O Connection Control Register 1                   */
    __IO uint32_t spix_req;                             /**< SPIX I/O Mode Request                                      */
    __IO uint32_t spix_ack;                             /**< SPIX I/O Mode Acknowledge                                  */
    __IO uint32_t uart0_req;                            /**< UART0 I/O Mode Request                                     */
    __IO uint32_t uart0_ack;                            /**< UART0 I/O Mode Acknowledge                                 */
    __IO uint32_t uart1_req;                            /**< UART1 I/O Mode Request                                     */
    __IO uint32_t uart1_ack;                            /**< UART1 I/O Mode Acknowledge                                 */
    __IO uint32_t uart2_req;                            /**< UART2 I/O Mode Request                                     */
    __IO uint32_t uart2_ack;                            /**< UART2 I/O Mode Acknowledge                                 */
    __IO uint32_t uart3_req;                            /**< UART3 I/O Mode Request                                     */
    __IO uint32_t uart3_ack;                            /**< UART3 I/O Mode Acknowledge                                 */
    __IO uint32_t i2cm0_req;                            /**< I2C Master 0 I/O Request                                   */
    __IO uint32_t i2cm0_ack;                            /**< I2C Master 0 I/O Acknowledge                               */
    __IO uint32_t i2cm1_req;                            /**< I2C Master 1 I/O Request                                   */
    __IO uint32_t i2cm1_ack;                            /**< I2C Master 1 I/O Acknowledge                               */
    __IO uint32_t i2cm2_req;                            /**< I2C Master 2 I/O Request                                   */
    __IO uint32_t i2cm2_ack;                            /**< I2C Master 2 I/O Acknowledge                               */
    __IO uint32_t i2cs_req;                             /**< I2C Slave I/O Request                                      */
    __IO uint32_t i2cs_ack;                             /**< I2C Slave I/O Acknowledge                                  */
    __IO uint32_t spim0_req;                            /**< SPI Master 0 I/O Mode Request                              */
    __IO uint32_t spim0_ack;                            /**< SPI Master 0 I/O Mode Acknowledge                          */
    __IO uint32_t spim1_req;                            /**< SPI Master 1 I/O Mode Request                              */
    __IO uint32_t spim1_ack;                            /**< SPI Master 1 I/O Mode Acknowledge                          */
    __IO uint32_t spim2_req;                            /**< SPI Master 2 I/O Mode Request                              */
    __IO uint32_t spim2_ack;                            /**< SPI Master 2 I/O Mode Acknowledge                          */
    __IO uint32_t spib_req;                             /**< SPI Bridge I/O Mode Request                                */
    __IO uint32_t spib_ack;                             /**< SPI Bridge I/O Mode Acknowledge                            */
    __IO uint32_t owm_req;                              /**< 1-Wire Master I/O Mode Request                             */
    __IO uint32_t owm_ack;                              /**< 1-Wire Master I/O Mode Acknowledge                         */
    __IO uint32_t spis_req;                             /**< SPI Slave I/O Mode Request                                 */
    __IO uint32_t spis_ack;                             /**< SPI Slave I/O Mode Acknowledge                             */
    __R  uint32_t rsv0A0[24];                           /**< RESERVED: DO NOT MODIFY                                    */
    __IO uint32_t use_vddioh_0;                         /**< Enable VDDIOH Register 0                                   */
    __IO uint32_t use_vddioh_1;                         /**< Enable VDDIOH Register 1                                   */
    __IO uint32_t use_vddioh_2;                         /**< Enable VDDIOH Register 2                                   */
    __R  uint32_t rsv10C;                               /**< RESERVED: DO NOT MODIFY                                    */
    __IO uint32_t pad_mode;                             /**< Pad Mode Control Register                                  */
    __R  uint32_t rsv114[27];                           /**< RESERVED: DO NOT MODIFY                                    */
    __IO uint32_t wud_req2;                             /**< Wakeup Detect Mode Request Register 2 (Port 8)             */
    __R  uint32_t rsv184;                               /**< RESERVED: DO NOT MODIFY                                    */
    __IO uint32_t wud_ack2;                             /**< Wakeup Detect Mode Acknowledge Register 2 (Port 8)         */
    __R  uint32_t rsv18C;                               /**< RESERVED: DO NOT MODIFY                                    */
    __IO uint32_t ali_req2;                             /**< Analog Input Request Register 2 (Port 8)                   */
    __R  uint32_t rsv194;                               /**< RESERVED: DO NOT MODIFY                                    */
    __IO uint32_t ali_ack2;                             /**< Analog Input Acknowledge Register 2 (Port 8)               */
    __R  uint32_t rsv19C;                               /**< RESERVED: DO NOT MODIFY                                    */
    __IO uint32_t ali_connect2;                         /**< Analog I/O Connection Control Register 2                   */
} mxc_ioman_regs_t;
/**@}*/

/*
   Register offsets for module IOMAN.
*/
/**
 * @ingroup    ioman_registers
 * @defgroup   ioman_reg_offs IOMAN Register Offsets
 * @{
 * @details    The @ref IOMAN_REGS_OFFS_TABLE "IOMAN Register Offset Table"
 *             shows the register offsets for the IOMAN registers from the base
 *             IOMAN peripheral address, #MXC_BASE_IOMAN.
 * @anchor IOMAN_REGS_OFFS_TABLE
 * | Register     | Offset |
 * | :----------- | ------:|
 * | WUD_REQ0     | 0x0000 |
 * | WUD_REQ1     | 0x0004 |
 * | WUD_ACK0     | 0x0008 |
 * | WUD_ACK1     | 0x000C |
 * | ALI_REQ0     | 0x0010 |
 * | ALI_REQ1     | 0x0014 |
 * | ALI_ACK0     | 0x0018 |
 * | ALI_ACK1     | 0x001C |
 * | ALI_CONNECT0 | 0x0020 |
 * | ALI_CONNECT1 | 0x0024 |
 * | SPIX_REQ     | 0x0028 |
 * | SPIX_ACK     | 0x002C |
 * | UART0_REQ    | 0x0030 |
 * | UART0_ACK    | 0x0034 |
 * | UART1_REQ    | 0x0038 |
 * | UART1_ACK    | 0x003C |
 * | UART2_REQ    | 0x0040 |
 * | UART2_ACK    | 0x0044 |
 * | UART3_REQ    | 0x0048 |
 * | UART3_ACK    | 0x004C |
 * | I2CM0_REQ    | 0x0050 |
 * | I2CM0_ACK    | 0x0054 |
 * | I2CM1_REQ    | 0x0058 |
 * | I2CM1_ACK    | 0x005C |
 * | I2CM2_REQ    | 0x0060 |
 * | I2CM2_ACK    | 0x0064 |
 * | I2CS_REQ     | 0x0068 |
 * | I2CS_ACK     | 0x006C |
 * | SPIM0_REQ    | 0x0070 |
 * | SPIM0_ACK    | 0x0074 |
 * | SPIM1_REQ    | 0x0078 |
 * | SPIM1_ACK    | 0x007C |
 * | SPIM2_REQ    | 0x0080 |
 * | SPIM2_ACK    | 0x0084 |
 * | SPIB_REQ     | 0x0088 |
 * | SPIB_ACK     | 0x008C |
 * | OWM_REQ      | 0x0090 |
 * | OWM_ACK      | 0x0094 |
 * | SPIS_REQ     | 0x0098 |
 * | SPIS_ACK     | 0x009C |
 * | USE_VDDIOH_0 | 0x0100 |
 * | USE_VDDIOH_1 | 0x0104 |
 * | USE_VDDIOH_2 | 0x0108 |
 * | PAD_MODE     | 0x0110 |
 * | WUD_REQ2     | 0x0180 |
 * | WUD_ACK2     | 0x0188 |
 * | ALI_REQ2     | 0x0190 |
 * | ALI_ACK2     | 0x0198 |
 * | ALI_CONNECT2 | 0x01A0 |
 */
#define MXC_R_IOMAN_OFFS_WUD_REQ0       ((uint32_t)0x00000000UL)    /**< WUD_REQ0 Register Offset from base IOMAN Peripheral Address.       */
#define MXC_R_IOMAN_OFFS_WUD_REQ1       ((uint32_t)0x00000004UL)    /**< WUD_REQ1 Register Offset from base IOMAN Peripheral Address.       */
#define MXC_R_IOMAN_OFFS_WUD_ACK0       ((uint32_t)0x00000008UL)    /**< WUD_ACK0 Register Offset from base IOMAN Peripheral Address.       */
#define MXC_R_IOMAN_OFFS_WUD_ACK1       ((uint32_t)0x0000000CUL)    /**< WUD_ACK1 Register Offset from base IOMAN Peripheral Address.       */
#define MXC_R_IOMAN_OFFS_ALI_REQ0       ((uint32_t)0x00000010UL)    /**< ALI_REQ0 Register Offset from base IOMAN Peripheral Address.       */
#define MXC_R_IOMAN_OFFS_ALI_REQ1       ((uint32_t)0x00000014UL)    /**< ALI_REQ1 Register Offset from base IOMAN Peripheral Address.       */
#define MXC_R_IOMAN_OFFS_ALI_ACK0       ((uint32_t)0x00000018UL)    /**< ALI_ACK0 Register Offset from base IOMAN Peripheral Address.       */
#define MXC_R_IOMAN_OFFS_ALI_ACK1       ((uint32_t)0x0000001CUL)    /**< ALI_ACK1 Register Offset from base IOMAN Peripheral Address.       */
#define MXC_R_IOMAN_OFFS_ALI_CONNECT0   ((uint32_t)0x00000020UL)    /**< ALI_CONNECT0 Register Offset from base IOMAN Peripheral Address.   */
#define MXC_R_IOMAN_OFFS_ALI_CONNECT1   ((uint32_t)0x00000024UL)    /**< ALI_CONNECT1 Register Offset from base IOMAN Peripheral Address.   */
#define MXC_R_IOMAN_OFFS_SPIX_REQ       ((uint32_t)0x00000028UL)    /**< SPIX_REQ Register Offset from base IOMAN Peripheral Address.       */
#define MXC_R_IOMAN_OFFS_SPIX_ACK       ((uint32_t)0x0000002CUL)    /**< SPIX_ACK Register Offset from base IOMAN Peripheral Address.       */
#define MXC_R_IOMAN_OFFS_UART0_REQ      ((uint32_t)0x00000030UL)    /**< UART0_REQ Register Offset from base IOMAN Peripheral Address.      */
#define MXC_R_IOMAN_OFFS_UART0_ACK      ((uint32_t)0x00000034UL)    /**< UART0_ACK Register Offset from base IOMAN Peripheral Address.      */
#define MXC_R_IOMAN_OFFS_UART1_REQ      ((uint32_t)0x00000038UL)    /**< UART1_REQ Register Offset from base IOMAN Peripheral Address.      */
#define MXC_R_IOMAN_OFFS_UART1_ACK      ((uint32_t)0x0000003CUL)    /**< UART1_ACK Register Offset from base IOMAN Peripheral Address.      */
#define MXC_R_IOMAN_OFFS_UART2_REQ      ((uint32_t)0x00000040UL)    /**< UART2_REQ Register Offset from base IOMAN Peripheral Address.      */
#define MXC_R_IOMAN_OFFS_UART2_ACK      ((uint32_t)0x00000044UL)    /**< UART2_ACK Register Offset from base IOMAN Peripheral Address.      */
#define MXC_R_IOMAN_OFFS_UART3_REQ      ((uint32_t)0x00000048UL)    /**< UART3_REQ Register Offset from base IOMAN Peripheral Address.      */
#define MXC_R_IOMAN_OFFS_UART3_ACK      ((uint32_t)0x0000004CUL)    /**< UART3_ACK Register Offset from base IOMAN Peripheral Address.      */
#define MXC_R_IOMAN_OFFS_I2CM0_REQ      ((uint32_t)0x00000050UL)    /**< I2CM0_REQ Register Offset from base IOMAN Peripheral Address.      */
#define MXC_R_IOMAN_OFFS_I2CM0_ACK      ((uint32_t)0x00000054UL)    /**< I2CM0_ACK Register Offset from base IOMAN Peripheral Address.      */
#define MXC_R_IOMAN_OFFS_I2CM1_REQ      ((uint32_t)0x00000058UL)    /**< I2CM1_REQ Register Offset from base IOMAN Peripheral Address.      */
#define MXC_R_IOMAN_OFFS_I2CM1_ACK      ((uint32_t)0x0000005CUL)    /**< I2CM1_ACK Register Offset from base IOMAN Peripheral Address.      */
#define MXC_R_IOMAN_OFFS_I2CM2_REQ      ((uint32_t)0x00000060UL)    /**< I2CM2_REQ Register Offset from base IOMAN Peripheral Address.      */
#define MXC_R_IOMAN_OFFS_I2CM2_ACK      ((uint32_t)0x00000064UL)    /**< I2CM2_ACK Register Offset from base IOMAN Peripheral Address.      */
#define MXC_R_IOMAN_OFFS_I2CS_REQ       ((uint32_t)0x00000068UL)    /**< I2CS_REQ Register Offset from base IOMAN Peripheral Address.       */
#define MXC_R_IOMAN_OFFS_I2CS_ACK       ((uint32_t)0x0000006CUL)    /**< I2CS_ACK Register Offset from base IOMAN Peripheral Address.       */
#define MXC_R_IOMAN_OFFS_SPIM0_REQ      ((uint32_t)0x00000070UL)    /**< SPIM0_REQ Register Offset from base IOMAN Peripheral Address.      */
#define MXC_R_IOMAN_OFFS_SPIM0_ACK      ((uint32_t)0x00000074UL)    /**< SPIM0_ACK Register Offset from base IOMAN Peripheral Address.      */
#define MXC_R_IOMAN_OFFS_SPIM1_REQ      ((uint32_t)0x00000078UL)    /**< SPIM1_REQ Register Offset from base IOMAN Peripheral Address.      */
#define MXC_R_IOMAN_OFFS_SPIM1_ACK      ((uint32_t)0x0000007CUL)    /**< SPIM1_ACK Register Offset from base IOMAN Peripheral Address.      */
#define MXC_R_IOMAN_OFFS_SPIM2_REQ      ((uint32_t)0x00000080UL)    /**< SPIM2_REQ Register Offset from base IOMAN Peripheral Address.      */
#define MXC_R_IOMAN_OFFS_SPIM2_ACK      ((uint32_t)0x00000084UL)    /**< SPIM2_ACK Register Offset from base IOMAN Peripheral Address.      */
#define MXC_R_IOMAN_OFFS_SPIB_REQ       ((uint32_t)0x00000088UL)    /**< SPIB_REQ Register Offset from base IOMAN Peripheral Address.       */
#define MXC_R_IOMAN_OFFS_SPIB_ACK       ((uint32_t)0x0000008CUL)    /**< SPIB_ACK Register Offset from base IOMAN Peripheral Address.       */
#define MXC_R_IOMAN_OFFS_OWM_REQ        ((uint32_t)0x00000090UL)    /**< OWM_REQ Register Offset from base IOMAN Peripheral Address.        */
#define MXC_R_IOMAN_OFFS_OWM_ACK        ((uint32_t)0x00000094UL)    /**< OWM_ACK Register Offset from base IOMAN Peripheral Address.        */
#define MXC_R_IOMAN_OFFS_SPIS_REQ       ((uint32_t)0x00000098UL)    /**< SPIS_REQ Register Offset from base IOMAN Peripheral Address.       */
#define MXC_R_IOMAN_OFFS_SPIS_ACK       ((uint32_t)0x0000009CUL)    /**< SPIS_ACK Register Offset from base IOMAN Peripheral Address.       */
#define MXC_R_IOMAN_OFFS_USE_VDDIOH_0   ((uint32_t)0x00000100UL)    /**< USE_VDDIOH_0 Register Offset from base IOMAN Peripheral Address.   */
#define MXC_R_IOMAN_OFFS_USE_VDDIOH_1   ((uint32_t)0x00000104UL)    /**< USE_VDDIOH_1 Register Offset from base IOMAN Peripheral Address.   */
#define MXC_R_IOMAN_OFFS_USE_VDDIOH_2   ((uint32_t)0x00000108UL)    /**< USE_VDDIOH_2 Register Offset from base IOMAN Peripheral Address.   */
#define MXC_R_IOMAN_OFFS_PAD_MODE       ((uint32_t)0x00000110UL)    /**< PAD_MODE Register Offset from base IOMAN Peripheral Address.       */
#define MXC_R_IOMAN_OFFS_WUD_REQ2       ((uint32_t)0x00000180UL)    /**< WUD_REQ2 Register Offset from base IOMAN Peripheral Address.       */
#define MXC_R_IOMAN_OFFS_WUD_ACK2       ((uint32_t)0x00000188UL)    /**< WUD_ACK2 Register Offset from base IOMAN Peripheral Address.       */
#define MXC_R_IOMAN_OFFS_ALI_REQ2       ((uint32_t)0x00000190UL)    /**< ALI_REQ2 Register Offset from base IOMAN Peripheral Address.       */
#define MXC_R_IOMAN_OFFS_ALI_ACK2       ((uint32_t)0x00000198UL)    /**< ALI_ACK2 Register Offset from base IOMAN Peripheral Address.       */
#define MXC_R_IOMAN_OFFS_ALI_CONNECT2   ((uint32_t)0x000001A0UL)    /**< ALI_CONNECT2 Register Offset from base IOMAN Peripheral Address.   */
/**@} end of ingroup ioman_regs_offs*/

/*
   Field positions and masks for module IOMAN.
*/
/**
 * @ingroup ioman_registers
 * @defgroup IOMAN_WUD_Req_Regs IOMAN_WUD_REQx
 * @brief IOMAN Wakeup Detect Register Fields Positions and Masks 
 * @{
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
/**@}*/

#ifdef __cplusplus
}
#endif

#endif   /* _MXC_IOMAN_REGS_H_ */

