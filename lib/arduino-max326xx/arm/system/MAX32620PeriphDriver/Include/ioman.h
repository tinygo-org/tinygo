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
 * $Date: 2017-05-19 15:53:14 -0500 (Fri, 19 May 2017) $
 * $Revision: 28074 $
 * 
 ******************************************************************************/

/**
 * @file    ioman.h
 * @brief   IOMAN provides IO Management to the device. The functions in this
 * API enable requesting port pin assignment and release for all peripherals
 * with external I/O. Port pin mapping support is included for peripherals
 * that can support more than one pin mapping in a package.
 */

#ifndef _IOMAN_H_
#define _IOMAN_H_

#include "mxc_config.h"
#include "ioman_regs.h"

#ifdef __cplusplus
extern "C" {
#endif

/***** Definitions *****/

/** @brief Aliases for IOMAN package mapping field values. Refer to the
  * User's Guide for pinouts for each mapping.
  */
typedef enum {
    IOMAN_MAP_UNUSED = 0,   /**< Pin is not used */
    IOMAN_MAP_A = 0,        /**< Pin Mapping A */
    IOMAN_MAP_B = 1,        /**< Pin Mapping B */
    IOMAN_MAP_C = 2,        /**< Pin Mapping C */
    IOMAN_MAP_D = 3,        /**< Pin Mapping D */
    IOMAN_MAP_E = 4,        /**< Pin Mapping E */
    IOMAN_MAP_F = 5,        /**< Pin Mapping F */
    IOMAN_MAP_G = 6         /**< Pin Mapping G */
}
ioman_map_t;

/** @brief Typing of IOMAN Req and Ack register fields */
typedef union {
    uint32_t                value;
    mxc_ioman_spix_req_t    spix;   /**< SPIX IOMAN configuration struct */
    mxc_ioman_uart0_req_t   uart;   /**< UART IOMAN configuration struct, see mxc_ioman_uart0_req_t */
    mxc_ioman_i2cm0_req_t   i2cm0;   /**< I2C Master 0 IOMAN configuration struct, see mxc_ioman_i2cm0_req_t */
    mxc_ioman_i2cm1_req_t   i2cm1;   /**< I2C Master 1 IOMAN configuration struct, see mxc_ioman_i2cm1_req_t */
    mxc_ioman_i2cm2_req_t   i2cm2;   /**< I2C Master 2 IOMAN configuration struct, see mxc_ioman_i2cm2_req_t */    
    mxc_ioman_i2cs_req_t    i2cs;   /**< I2C Slave IOMAN configuration struct, see mxc_ioman_i2cs_req_t */
    mxc_ioman_spim0_req_t   spim0;  /**< SPI Master 0 IOMAN configuration struct, see mxc_ioman_spim0_req_t */
    mxc_ioman_spim1_req_t   spim1;  /**< SPI Master 1 IOMAN configuration struct, see mxc_ioman_spim1_req_t */
    mxc_ioman_spim2_req_t   spim2;  /**< SPI Master 2 IOMAN configuration struct, see mxc_ioman_spim1_req_t */
    mxc_ioman_spis_req_t    spis;   /**< SPI Slave IOMAN configuration struct, see mxc_ioman_spis_req_t */
    mxc_ioman_owm_req_t     owm;    /**< 1-Wire Master IOMAN configuration struct, see mxc_ioman_owm_req_t */
} ioman_req_t;

/** @brief IOMAN configuration object */
typedef struct {
    volatile uint32_t *req_reg;     /** Pointer to an IOMAN request register */
    volatile uint32_t *ack_reg;     /** Pointer to an IOMAN acknowledge register */
    ioman_req_t req_val;            /** IOMAN request register value, see ioman_req_t */
} ioman_cfg_t;


/***** Function Prototypes *****/

/**
 * @brief Configure the IO Manager using the specified configuration object.
 * @param cfg       IOMAN configuration object
 * @returns         E_NO_ERROR Configuration successful
 */
int IOMAN_Config(const ioman_cfg_t *cfg);

/**
 * @brief Create an IOMAN configuration object for the SPI XIP module. Call IOMAN_Config with this object.
 * @param           core            Request (1) or release (0) SPIX core external pins
 * @param           ss0             Request (1) or release (0) slave select 0 active out
 * @param           ss1             Request (1) or release (0) slave select 1 active out
 * @param           ss2             Request (1) or release (0) slave select 2 active out
 * @param           quad            Request (1) or release (0) quad IO
 * @param           fast            Request (1) or release (0) fast mode
 * @returns         io_man_cfg_t    IOMAN configuration object for the SPI XIP module.
 */
ioman_cfg_t IOMAN_SPIX(int core, int ss0, int ss1, int ss2, int quad, int fast);

/**
 * @brief Create an IOMAN configuration object for a UART module. Call IOMAN_Config with this object.
 * @param           idx             Index of the UART module
 * @param           io_map          Set the pin mapping for RX/TX pins, see ioman_map_t
 * @param           cts_map         Set the pin mapping for CTS pin, see ioman_map_t
 * @param           rts_map         Set the pin mapping for RTS pin, see ioman_map_t
 * @param           io_en           Request (1) or release (0) RX and TX pins
 * @param           cts_en          Request (1) or release (0) CTS pin
 * @param           rts_en          Request (1) or release (0) RTS pin
 * @returns         ioman_cfg_t     IOMAN configuration object for the UART module
 */
ioman_cfg_t IOMAN_UART(int idx, ioman_map_t io_map, ioman_map_t cts_map, ioman_map_t rts_map, int io_en, int cts_en, int rts_en);

/**
 * @brief Create an IOMAN configuration object for the I2CM0 module. Call IOMAN_Config with this object.
 * @param           core            Request (1) or release (0) I2CM0 core external pins
 * @param           push_pull       (1) Use push/pull driver (0) us open drain driver for the I2CM0 module
 * @returns         ioman_cfg_t     IOMAN configuration object for the I2CM0 module.
 */
ioman_cfg_t IOMAN_I2CM0(int core, int push_pull);

/**
 * @brief Create an IOMAN configuration object for the I2CM1 module. Call IOMAN_Config with this object.
 * @param           core            Request (1) or release (0) I2CM1 core external pins
 * @param           push_pull       (1) Use push/pull driver (0) us open drain driver for the I2CM1 module
 * @returns         ioman_cfg_t     IOMAN configuration object for the I2CM1 module.
 */
ioman_cfg_t IOMAN_I2CM1(int core, int push_pull);

/**
 * @brief Create an IOMAN configuration object for the I2CM2 module. Call IOMAN_Config with this object.
 * @param           core            Request (1) or release (0) I2CM2 core external pins
 * @param           push_pull       (1) Use push/pull driver (0) us open drain driver for the I2CM2 module
 * @returns         ioman_cfg_t     IOMAN configuration object for the I2CM2 module.
 */
ioman_cfg_t IOMAN_I2CM2(int core, int push_pull);

/**
 * @brief Create an IOMAN configuration object for an I2C slave module. Call IOMAN_Config with this object.
 * @param           map             Select the pin mapping for all configured pins, see ioman_map_t
 * @param           io_en           Request (1) or release (0) the I/O for this module
 * @returns         ioman_cfg_t     IOMAN configuration object for the I2CS module
 */
ioman_cfg_t IOMAN_I2CS(ioman_map_t map, int io_en);

/**
 * @brief Create an IOMAN configuration object for a SPI Master (SPIM) module. Call IOMAN_Config with this object.
 * @param           io_en           Request (1) or release (0) the core IO for the module
 * @param           ss0             Request (1) or release (0) slave select 0
 * @param           ss1             Request (1) or release (0) slave select 1
 * @param           ss2             Request (1) or release (0) slave select 2
 * @param           ss3             Request (1) or release (0) slave select 3
 * @param           ss4             Request (1) or release (0) slave select 4
 * @param           quad            Request (1) or release (0) quad IO
 * @param           fast            Request (1) or release (0) fast mode
 * @returns         ioman_cfg_t     IOMAN configuration object for an SPIM0 module
 */
ioman_cfg_t IOMAN_SPIM0(int io_en, int ss0, int ss1, int ss2, int ss3, int ss4, int quad, int fast);

/**
 * @brief Create an IOMAN configuration object for a SPIM module. Call IOMAN_Config with this object.
 * @param           io_en           Request (1) or release (0) the core IO for the module
 * @param           ss0             Request (1) or release (0) slave select 0
 * @param           ss1             Request (1) or release (0) slave select 1
 * @param           ss2             Request (1) or release (0) slave select 2
 * @param           quad            Request (1) or release (0) quad IO
 * @param           fast            Request (1) or release (0) fast mode
 * @returns         ioman_cfg_t     IOMAN configuration object for the SPIM1 module.
 */
ioman_cfg_t IOMAN_SPIM1(int io_en, int ss0, int ss1, int ss2, int quad, int fast);

/**
 * @brief Create an IOMAN configuration object for a SPI module. Call IOMAN_Config with this object.
 * @param           map             Select the pin mapping, see ioman_map_t
 * @param           io_en           Request (1) or release (0) the core IO for the module
 * @param           ss0             Request (1) or release (0) slave select 0
 * @param           ss1             Request (1) or release (0) slave select 1
 * @param           ss2             Request (1) or release (0) slave select 2
 * @param           sr0             Request (1) or release (0) slave ready 0
 * @param           sr1             Request (1) or release (0) slave ready 1
 * @param           quad            Request (1) or release (0) quad IO
 * @param           fast            Request (1) or release (0) fast mode
 * @returns         ioman_cfg_t     IOMAN configuration object for the SPIM2 module
 */
ioman_cfg_t IOMAN_SPIM2(ioman_map_t map, int io_en, int ss0, int ss1, int ss2, int sr0, int sr1, int quad, int fast);

/**
 * @brief Create an IOMAN configuration object for a SPI module. Call IOMAN_Config with this object.
 * @param           io_en           Request (1) or release (0) the core IO for the module
 * @param           quad            Request (1) or release (0) quad IO
 * @param           fast            Request (1) or release (0) fast mode
 * @returns         ioman_cfg_t     IOMAN configuration object for the SPIM2 module
 */
ioman_cfg_t IOMAN_SPIS(int io_en, int quad, int fast);

/**
 * @brief Create an IOMAN configuration object for the 1-Wire Master module. Call IOMAN_Config with this object.
 * @param           io_en           Request (1) or release (0) the core IO for the module
 * @param           epu             Request (1) or release (0) external pullup
 * @returns         ioman_cfg_t     IOMAN configuration object for the OWM module
 */
ioman_cfg_t IOMAN_OWM(int io_en, int epu);

/**
 * @}
 */

/******************************************************************************/
/* All the function prototypes above are implemented as macros below. The
 * above prototypes are for simplicity in doxygen.
 */
#define IOMAN_SPIX(c, ss0, ss1, ss2, q, f) {                                        \
        .req_reg = &MXC_IOMAN->spix_req,                                            \
        .ack_reg = &MXC_IOMAN->spix_ack,                                            \
        .req_val = { .spix =  { .core_io_req = c,                                   \
                           .ss0_io_req = ss0,                                       \
                           .ss1_io_req = ss1,                                       \
                           .ss2_io_req = ss2,                                       \
                           .quad_io_req = q,                                        \
                           .fast_mode = f } } }

#define IOMAN_UART(i, im, cm, rm, ien, cen, ren) {                                                   \
        .req_reg = (uint32_t*)((unsigned int)(&MXC_IOMAN->uart0_req) + (i * 2*sizeof(uint32_t))),    \
        .ack_reg = (uint32_t*)((unsigned int)(&MXC_IOMAN->uart0_ack) + (i * 2*sizeof(uint32_t))),    \
        .req_val = { .uart = { .io_map = im,                                                         \
                               .cts_map = cm,                                                        \
                               .rts_map = rm,                                                        \
                               .io_req = ien,                                                        \
                               .cts_io_req = cen,                                                    \
                               .rts_io_req = ren } } }

#define IOMAN_I2CM0(c, p) {                                                         \
        .req_reg = &MXC_IOMAN->i2cm0_req,                                           \
        .ack_reg = &MXC_IOMAN->i2cm0_ack,                                           \
        .req_val = { .i2cm0 = { .core_io_req = c,                                   \
                                .push_pull = p } } }

#define IOMAN_I2CM1(c, p) {                                                         \
        .req_reg = &MXC_IOMAN->i2cm1_req,                                           \
        .ack_reg = &MXC_IOMAN->i2cm1_ack,                                           \
        .req_val = { .i2cm1 = { .core_io_req = c,                                   \
                                .push_pull = p } } }

#define IOMAN_I2CM2(c, p) {                                                         \
        .req_reg = &MXC_IOMAN->i2cm2_req,                                           \
        .ack_reg = &MXC_IOMAN->i2cm2_ack,                                           \
        .req_val = { .i2cm2 = { .core_io_req = c,                                   \
                                .push_pull = p } } }

#define IOMAN_I2CS(m, ien) {                                                        \
        .req_reg = &MXC_IOMAN->i2cs_req,                                            \
        .ack_reg = &MXC_IOMAN->i2cs_ack,                                            \
        .req_val = { .i2cs = { .mapping_req = m,                                    \
                               .core_io_req = ien } } }

#define IOMAN_SPIM0(io, ss0, ss1, ss2, ss3, ss4, q, f) {                            \
        .req_reg = &MXC_IOMAN->spim0_req,                                           \
        .ack_reg = &MXC_IOMAN->spim0_ack,                                           \
        .req_val = { .spim0 = { .core_io_req = io,                                  \
                                .ss0_io_req = ss0,                                  \
                                .ss1_io_req = ss1,                                  \
                                .ss2_io_req = ss2,                                  \
                                .ss3_io_req = ss3,                                  \
                                .ss4_io_req = ss4,                                  \
                                .quad_io_req = q,                                   \
                                .fast_mode = f } } }

#define IOMAN_SPIM1(io, ss0, ss1, ss2, q, f) {                                      \
        .req_reg = &MXC_IOMAN->spim1_req,                                           \
        .ack_reg = &MXC_IOMAN->spim1_ack,                                           \
        .req_val = { .spim1 = { .core_io_req = io,                                  \
                                .ss0_io_req = ss0,                                  \
                                .ss1_io_req = ss1,                                  \
                                .ss2_io_req = ss2,                                  \
                                .quad_io_req = q,                                   \
                                .fast_mode = f } } }

#define IOMAN_SPIM2(m, io, ss0, ss1, ss2, sr0, sr1, q, f) {                         \
        .req_reg = &MXC_IOMAN->spim2_req,                                           \
        .ack_reg = &MXC_IOMAN->spim2_ack,                                           \
        .req_val = { .spim2 = { .mapping_req = m,                                   \
                                .core_io_req = io,                                  \
                                .ss0_io_req = ss0,                                  \
                                .ss1_io_req = ss1,                                  \
                                .ss2_io_req = ss2,                                  \
                                .sr0_io_req = sr0,                                  \
                                .sr1_io_req = sr1,                                  \
                                .quad_io_req = q,                                   \
                                .fast_mode = f } } }

#define IOMAN_SPIS(io, q, f) {                                                      \
        .req_reg = &MXC_IOMAN->spis_req,                                            \
        .ack_reg = &MXC_IOMAN->spis_ack,                                            \
        .req_val = { .spis = {  .core_io_req = io,                                  \
                                .quad_io_req = q,                                   \
                                .fast_mode = f } } }

#define IOMAN_OWM(io, p) {                                                          \
        .req_reg = &MXC_IOMAN->owm_req,                                             \
        .ack_reg = &MXC_IOMAN->owm_ack,                                             \
        .req_val = { .owm =  {  .mapping_req = io,                                  \
                                .epu_io_req = p } } }

#ifdef __cplusplus
}
#endif

#endif /* _IOMAN_H_ */
