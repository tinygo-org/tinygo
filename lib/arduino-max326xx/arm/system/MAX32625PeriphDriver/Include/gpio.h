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

/**
 * @file    gpio.h
 * @brief   GPIO driver header file.
 */

#ifndef _GPIO_H_
#define _GPIO_H_

#include "mxc_config.h"
#include "gpio_regs.h"

#ifdef __cplusplus
extern "C" {
#endif

/***** Definitions *****/
#define PORT_0      (0)
#define PORT_1      (1)
#define PORT_2      (2)
#define PORT_3      (3)
#define PORT_4      (4)
#define PORT_5      (5)
#define PORT_6      (6)
#define PORT_7      (7)
#define PORT_8      (8)
#define PORT_9      (9)
#define PORT_10     (10)
#define PORT_11     (11)
#define PORT_12     (12)
#define PORT_13     (13)
#define PORT_14     (14)
#define PORT_15     (15)

#define PIN_0       (1 << 0)
#define PIN_1       (1 << 1)
#define PIN_2       (1 << 2)
#define PIN_3       (1 << 3)
#define PIN_4       (1 << 4)
#define PIN_5       (1 << 5)
#define PIN_6       (1 << 6)
#define PIN_7       (1 << 7)

/// \brief GPIO Function Types
typedef enum {
    GPIO_FUNC_GPIO  = MXC_V_GPIO_FUNC_SEL_MODE_GPIO,
    GPIO_FUNC_PT    = MXC_V_GPIO_FUNC_SEL_MODE_PT,
    GPIO_FUNC_TMR   = MXC_V_GPIO_FUNC_SEL_MODE_TMR
} gpio_func_t;

/// \brief GPIO Pad Types
typedef enum {
    GPIO_PAD_INPUT_PULLUP           = MXC_V_GPIO_OUT_MODE_HIGH_Z_WEAK_PULLUP,
    GPIO_PAD_OPEN_DRAIN             = MXC_V_GPIO_OUT_MODE_OPEN_DRAIN,
    GPIO_PAD_OPEN_DRAIN_PULLUP      = MXC_V_GPIO_OUT_MODE_OPEN_DRAIN_WEAK_PULLUP,
    GPIO_PAD_INPUT                  = MXC_V_GPIO_OUT_MODE_NORMAL_HIGH_Z,
    GPIO_PAD_NORMAL                 = MXC_V_GPIO_OUT_MODE_NORMAL,
    GPIO_PAD_SLOW                   = MXC_V_GPIO_OUT_MODE_SLOW_DRIVE,
    GPIO_PAD_FAST                   = MXC_V_GPIO_OUT_MODE_FAST_DRIVE,
    GPIO_PAD_INPUT_PULLDOWN         = MXC_V_GPIO_OUT_MODE_HIGH_Z_WEAK_PULLDOWN,
    GPIO_PAD_OPEN_SOURCE            = MXC_V_GPIO_OUT_MODE_OPEN_SOURCE,
    GPIO_PAD_OPEN_SOURCE_PULLDOWN   = MXC_V_GPIO_OUT_MODE_OPEN_SOURCE_WEAK_PULLDOWN,
} gpio_pad_t;

/// \brief GPIO Configuration Structure
typedef struct {
    uint32_t port;      /// Index of GPIO port
    uint32_t mask;      /// Pin mask. Multiple bits can be set.
    gpio_func_t func;   /// Function type
    gpio_pad_t pad;     /// Pad type
} gpio_cfg_t;

/// \brief GPIO Mode Structure
typedef enum {
    GPIO_INT_DISABLE        = MXC_V_GPIO_INT_MODE_DISABLE,
    GPIO_INT_FALLING_EDGE   = MXC_V_GPIO_INT_MODE_FALLING_EDGE,
    GPIO_INT_RISING_EDGE    = MXC_V_GPIO_INT_MODE_RISING_EDGE,
    GPIO_INT_ANY_EDGE       = MXC_V_GPIO_INT_MODE_ANY_EDGE,
    GPIO_INT_LOW_LEVEL      = MXC_V_GPIO_INT_MODE_LOW_LVL,
    GPIO_INT_HIGH_LEVEL     = MXC_V_GPIO_INT_MODE_HIGH_LVL,
} gpio_int_mode_t;

/***** Function Prototypes *****/

/**
 * \brief   Configure GPIO pin(s)
 * \param   cfg   Pointer to configuration structure describing the pin
 * \returns #E_NO_ERROR if everything is successful
 */
int GPIO_Config(const gpio_cfg_t *cfg);

/**
 * \brief   Gets the pin(s) input state
 * \param   cfg   Pointer to configuration structure describing the pin
 * \returns The requested pin state
 */
__STATIC_INLINE uint32_t GPIO_InGet(const gpio_cfg_t *cfg)
{
    return (MXC_GPIO->in_val[cfg->port] & cfg->mask);
}

/**
 * \brief   Sets the pin(s) to a high level output
 * \param   cfg   Pointer to configuration structure describing the pin
 */
__STATIC_INLINE void GPIO_OutSet(const gpio_cfg_t *cfg)
{
    MXC_GPIO->out_val[cfg->port] |= cfg->mask;
}

/**
 * \brief   Clears the pin(s) to a low level output
 * \param   cfg   Pointer to configuration structure describing the pin
 */
__STATIC_INLINE void GPIO_OutClr(const gpio_cfg_t *cfg)
{
    MXC_GPIO->out_val[cfg->port] &= ~(cfg->mask);
}

/**
 * \brief   Gets the pin(s) output state
 * \param   cfg   Pointer to configuration structure describing the pin
 * \returns The requested pin state
 */
__STATIC_INLINE uint32_t GPIO_OutGet(const gpio_cfg_t *cfg)
{
    return (MXC_GPIO->out_val[cfg->port] & cfg->mask);
}

/**
 * \brief   Write the pin(s) to a desired output level
 * \param   cfg   Pointer to configuration structure describing the pin
 * \param   val   Desired output level of the pin(s). This will be masked with the configuration mask.
 */
__STATIC_INLINE void GPIO_OutPut(const gpio_cfg_t *cfg, uint32_t val)
{
    MXC_GPIO->out_val[cfg->port] = (MXC_GPIO->out_val[cfg->port] & ~cfg->mask) | (val & cfg->mask);
}

/**
 * \brief   Toggles the the pin(s) output level
 * \param   cfg   Pointer to configuration structure describing the pin
 */
__STATIC_INLINE void GPIO_OutToggle(const gpio_cfg_t *cfg)
{
    MXC_GPIO->out_val[cfg->port] ^= cfg->mask;
}

/**
 * \brief   Configure GPIO interrupt(s)
 * \param   cfg     Pointer to configuration structure describing the pin
 * \param   mode    Requested interrupt mode
 */
void GPIO_IntConfig(const gpio_cfg_t *cfg, gpio_int_mode_t mode);

/**
 * \brief   Enables the specified GPIO interrupt
 * \param   cfg   Pointer to configuration structure describing the pin
 */
__STATIC_INLINE void GPIO_IntEnable(const gpio_cfg_t *cfg)
{
    MXC_GPIO->inten[cfg->port] |= cfg->mask;
}

/**
 * \brief   Disables the specified GPIO interrupt
 * \param   cfg   Pointer to configuration structure describing the pin
 */
__STATIC_INLINE void GPIO_IntDisable(const gpio_cfg_t *cfg)
{
    MXC_GPIO->inten[cfg->port] &= ~cfg->mask;
}

/**
 * \brief   Gets the interrupt(s) status
 * \param   cfg   Pointer to configuration structure describing the pin
 * \returns The requested interrupt status
 */
__STATIC_INLINE uint32_t GPIO_IntStatus(const gpio_cfg_t *cfg)
{
    return (MXC_GPIO->intfl[cfg->port] & cfg->mask);
}

/**
 * \brief   Clears the interrupt(s) status
 * \param   cfg   Pointer to configuration structure describing the pin
 */
__STATIC_INLINE void GPIO_IntClr(const gpio_cfg_t *cfg)
{
    MXC_GPIO->intfl[cfg->port] = cfg->mask;
}

/**
 * \brief   Registers a callback for the interrupt
 * \param   cfg       Pointer to configuration structure describing the pin
 * \param   callback  The function to be called
 * \param   cbdata    The parameter to be passed to the callback function
 */
void GPIO_RegisterCallback(const gpio_cfg_t *cfg, void (*callback)(void *cbdata), void *cbdata);

/**
 * \brief   Handle GPIO interrupts can call registered callbacks
 * \param   gpio  Pointer to base address of GPIO module
 */
void GPIO_Handler(unsigned int port);

#ifdef __cplusplus
}
#endif

#endif /* _GPIO_H_ */
