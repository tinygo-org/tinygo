/**
 * @file
 * @brief      This file contains the function implementations for the 
 *             General-Purpose Input/Output (GPIO) peripheral module.
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
 * $Date: 2016-09-09 11:41:02 -0500 (Fri, 09 Sep 2016) $
 * $Revision: 24337 $
 *
 *************************************************************************** */

/* **** Includes **** */
#include "mxc_config.h"
#include "mxc_assert.h"
#include "mxc_sys.h"
#include "gpio.h"
#include "clkman_regs.h"

/**
 * @ingroup gpio
 * @{
 */

/* **** Definitions **** */

/* **** Globals **** */

/* ************************************************************************* */
static void (*callbacks[MXC_GPIO_NUM_PORTS][MXC_GPIO_MAX_PINS_PER_PORT])(void *);
static void *cbparam[MXC_GPIO_NUM_PORTS][MXC_GPIO_MAX_PINS_PER_PORT];

/* **** Functions **** */

/* ************************************************************************* */
static int PinConfig(unsigned int port, unsigned int pin, gpio_func_t func, gpio_pad_t pad)
{
    /* Check if available */
    if (!(MXC_GPIO->free[port] & (1 << pin))) {
        return E_BUSY;
    }

    /* Set function */
    uint32_t func_sel = MXC_GPIO->func_sel[port];
    func_sel &= ~(0xF << (4 * pin));
    func_sel |=  (func << (4 * pin));
    MXC_GPIO->func_sel[port] = func_sel;

    /* Normal input is always enabled */
    MXC_GPIO->in_mode[port] &= ~(0xF << (4 * pin));

    /* Set requested output mode */
    uint32_t out_mode = MXC_GPIO->out_mode[port];
    out_mode &= ~(0xF << (4 * pin));
    out_mode |=  (pad << (4 * pin));
    MXC_GPIO->out_mode[port] = out_mode;

    /* Enable the pull up/down if necessary */
    if (pad == MXC_V_GPIO_OUT_MODE_HIGH_Z_WEAK_PULLUP) {
        MXC_GPIO->out_val[port] |= (1 << pin);
    } else if (pad == MXC_V_GPIO_OUT_MODE_HIGH_Z_WEAK_PULLDOWN) {
        MXC_GPIO->out_val[port] &= ~(1 << pin);
    }

    return E_NO_ERROR;
}

/* ************************************************************************* */
int GPIO_Config(const gpio_cfg_t *cfg)
{
    unsigned int pin;
    int err = E_NO_ERROR;

    MXC_ASSERT(cfg);
    MXC_ASSERT(cfg->port < MXC_GPIO_NUM_PORTS);

    // Set system level configurations
    if ((err = SYS_GPIO_Init()) != E_NO_ERROR) {
        return err;
    }

    // Configure each pin in the mask
    for (pin = 0; pin < MXC_GPIO_MAX_PINS_PER_PORT; pin++) {
        if (cfg->mask & (1 << pin)) {
            if (PinConfig(cfg->port, pin, cfg->func, cfg->pad) != E_NO_ERROR) {
                err = E_BUSY;
            }
        }
    }

    return err;
}

/* ************************************************************************* */
static void IntConfig(unsigned int port, unsigned int pin, gpio_int_mode_t mode)
{
    uint32_t int_mode = MXC_GPIO->int_mode[port];
    int_mode &= ~(0xF << (pin*4));
    int_mode |= (mode << (pin*4));
    MXC_GPIO->int_mode[port] = int_mode;
}

/* ************************************************************************* */
void GPIO_IntConfig(const gpio_cfg_t *cfg, gpio_int_mode_t mode)
{
    unsigned int pin;

    MXC_ASSERT(cfg);
    MXC_ASSERT(cfg->port < MXC_GPIO_NUM_PORTS);

    // Configure each pin in the mask
    for (pin = 0; pin < MXC_GPIO_MAX_PINS_PER_PORT; pin++) {
        if (cfg->mask & (1 << pin)) {
            IntConfig(cfg->port, pin, mode);
        }
    }
}

/* ************************************************************************* */
void GPIO_RegisterCallback(const gpio_cfg_t *cfg, gpio_callback_fn func, void *cbdata)
{
    unsigned int pin;

    MXC_ASSERT(cfg);
    MXC_ASSERT(cfg->port < MXC_GPIO_NUM_PORTS);

    for (pin = 0; pin < MXC_GPIO_MAX_PINS_PER_PORT; pin++) {
        if (cfg->mask & (1 << pin)) {
            callbacks[cfg->port][pin] = func;
            cbparam[cfg->port][pin] = cbdata;
        }
    }
}

/* ************************************************************************* */
void GPIO_Handler(unsigned int port)
{
    uint8_t intfl;
    unsigned int pin;

    MXC_ASSERT(port < MXC_GPIO_NUM_PORTS);

    // Read and clear enabled interrupts.
    intfl = MXC_GPIO->intfl[port];
    intfl &= MXC_GPIO->inten[port];
    MXC_GPIO->intfl[port] = intfl;

    // Process each pins' interrupt
    for (pin = 0; pin < MXC_GPIO_MAX_PINS_PER_PORT; pin++) {
        if ((intfl & (1 << pin)) && callbacks[port][pin]) {
            callbacks[port][pin](cbparam[port][pin]);
        }
    }
}

/**@} end of group gpio */
