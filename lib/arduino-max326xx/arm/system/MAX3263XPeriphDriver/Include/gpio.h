/**
 * @file
 * @brief   General-Purpose Input/Output (GPIO) function prototypes and data types.
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
 * $Date: 2016-10-10 18:56:06 -0500 (Mon, 10 Oct 2016) $
 * $Revision: 24659 $
 *
 *************************************************************************** */

/* Define to prevent redundant inclusion */
#ifndef _GPIO_H_
#define _GPIO_H_

/* **** Includes **** */
#include "mxc_config.h"
#include "gpio_regs.h"

#ifdef __cplusplus
extern "C" {
#endif

// Doxy group definition for this peripheral module
/**
 * @ingroup periphlibs
 * @defgroup gpio General-Purpose Input/Output (GPIO)
 * @{
 */

/* **** Definitions **** */
/**
 * @defgroup gpio_port_pin Port and Pin Definitions
 * @ingroup gpio
 * @{
 * @defgroup gpio_port Port Definitions
 * @ingroup gpio_port_pin
 * @{
 */
#define PORT_0      (0)             /**< Port 0  Define*/
#define PORT_1      (1)             /**< Port 1  Define*/
#define PORT_2      (2)             /**< Port 2  Define*/
#define PORT_3      (3)             /**< Port 3  Define*/
#define PORT_4      (4)             /**< Port 4  Define*/
#define PORT_5      (5)             /**< Port 5  Define*/
#define PORT_6      (6)             /**< Port 6  Define*/
#define PORT_7      (7)             /**< Port 7  Define*/
#define PORT_8      (8)             /**< Port 8  Define*/
#define PORT_9      (9)             /**< Port 9  Define*/
#define PORT_10     (10)            /**< Port 10  Define*/
#define PORT_11     (11)            /**< Port 11  Define*/
#define PORT_12     (12)            /**< Port 12  Define*/
#define PORT_13     (13)            /**< Port 13  Define*/
#define PORT_14     (14)            /**< Port 14  Define*/
#define PORT_15     (15)            /**< Port 15  Define*/
/**@} end of gpio_port group*/
/**
 * @defgroup gpio_pin Pin Definitions
 * @ingroup gpio_port_pin
 * @{
 */
#define PIN_0       (1 << 0)        /**< Pin 0 Define */
#define PIN_1       (1 << 1)        /**< Pin 1 Define */
#define PIN_2       (1 << 2)        /**< Pin 2 Define */
#define PIN_3       (1 << 3)        /**< Pin 3 Define */
#define PIN_4       (1 << 4)        /**< Pin 4 Define */
#define PIN_5       (1 << 5)        /**< Pin 5 Define */
#define PIN_6       (1 << 6)        /**< Pin 6 Define */
#define PIN_7       (1 << 7)        /**< Pin 7 Define */
/**@} end of gpio_pin group */
/**@} end of gpio_port_pin group */
    
/**
 * Enumeration type for the GPIO Function Type
 */
typedef enum {
    GPIO_FUNC_GPIO  = MXC_V_GPIO_FUNC_SEL_MODE_GPIO,    /**< GPIO Function Selection */
    GPIO_FUNC_PT    = MXC_V_GPIO_FUNC_SEL_MODE_PT,      /**< Pulse Train Function Selection */
    GPIO_FUNC_TMR   = MXC_V_GPIO_FUNC_SEL_MODE_TMR      /**< Timer Function Selection */
}
gpio_func_t;

/**
 * Enumeration type for the type of GPIO pad on a given pin.  
 */
typedef enum {
    GPIO_PAD_INPUT_PULLUP           = MXC_V_GPIO_OUT_MODE_HIGH_Z_WEAK_PULLUP,       /**< Set pad to high impedance, weak pull-up */
    GPIO_PAD_OPEN_DRAIN             = MXC_V_GPIO_OUT_MODE_OPEN_DRAIN,               /**< Set pad to open-drain with high impedance with input buffer */
    GPIO_PAD_OPEN_DRAIN_PULLUP      = MXC_V_GPIO_OUT_MODE_OPEN_DRAIN_WEAK_PULLUP,   /**< Set pad to open-drain with weak pull-up */
    GPIO_PAD_INPUT                  = MXC_V_GPIO_OUT_MODE_NORMAL_HIGH_Z,            /**< Set pad to high impednace, input buffer enabled */
    GPIO_PAD_NORMAL                 = MXC_V_GPIO_OUT_MODE_NORMAL,                   /**< Set pad to normal drive mode for high an low output */
    GPIO_PAD_SLOW                   = MXC_V_GPIO_OUT_MODE_SLOW_DRIVE,               /**< Set pad to slow drive mode, which is normal mode with negative feedback to slow edge transitions */
    GPIO_PAD_FAST                   = MXC_V_GPIO_OUT_MODE_FAST_DRIVE,               /**< Set pad to fash drive mode, which is normal mode with a transistor drive to drive fast high and low */
    GPIO_PAD_INPUT_PULLDOWN         = MXC_V_GPIO_OUT_MODE_HIGH_Z_WEAK_PULLDOWN,     /**< Set pad to weak pulldown mode */ 
    GPIO_PAD_OPEN_SOURCE            = MXC_V_GPIO_OUT_MODE_OPEN_SOURCE,              /**< Set pad to open source mode, transistor drive to high */
    GPIO_PAD_OPEN_SOURCE_PULLDOWN   = MXC_V_GPIO_OUT_MODE_OPEN_SOURCE_WEAK_PULLDOWN /**< Set pad to open source with weak pulldown mode, transistor drive to high, weak pulldown to GND for low */
} gpio_pad_t;

/**
 * Structure type for configuring a GPIO port.
 */
typedef struct {
    uint32_t port;      /// Index of GPIO port
    uint32_t mask;      /// Pin mask. Multiple bits can be set.
    gpio_func_t func;   /// Function type
    gpio_pad_t pad;     /// Pad type
} gpio_cfg_t;

/**
 * Enumeration type for the interrupt type on a GPIO port. 
 */
typedef enum {
    GPIO_INT_DISABLE        = MXC_V_GPIO_INT_MODE_DISABLE,          /**< Disable interrupts */
    GPIO_INT_FALLING_EDGE   = MXC_V_GPIO_INT_MODE_FALLING_EDGE,     /**< Interrupt on Falling Edge */
    GPIO_INT_RISING_EDGE    = MXC_V_GPIO_INT_MODE_RISING_EDGE,      /**< Interrupt on Rising Edge */
    GPIO_INT_ANY_EDGE       = MXC_V_GPIO_INT_MODE_ANY_EDGE,         /**< Interrupt on Falling or Rising Edge */
    GPIO_INT_LOW_LEVEL      = MXC_V_GPIO_INT_MODE_LOW_LVL,          /**< Interrupt on a low level input detection */
    GPIO_INT_HIGH_LEVEL     = MXC_V_GPIO_INT_MODE_HIGH_LVL          /**< Interrupt on a high level input detection */
} gpio_int_mode_t;

/* **** Function Prototypes **** */

/**
 * @brief      Configure GPIO pin(s).
 * @param      cfg   Pointer to configuration structure describing the pin.
 * 
 * @return     #E_NO_ERROR if everything is successful.
 * 
 */
int GPIO_Config(const gpio_cfg_t *cfg);

/**
 * @brief      Gets the pin(s) input state.
 * @param      cfg   Pointer to configuration structure describing the pin.
 * 
 * @return     The requested pin state.
 * 
 */
__STATIC_INLINE uint32_t GPIO_InGet(const gpio_cfg_t *cfg)
{
    return (MXC_GPIO->in_val[cfg->port] & cfg->mask);
}

/**
 * @brief      Sets the pin(s) to a high level output.
 * @param      cfg   Pointer to configuration structure describing the pin.
 * 
 */
__STATIC_INLINE void GPIO_OutSet(const gpio_cfg_t *cfg)
{
    MXC_GPIO->out_val[cfg->port] |= cfg->mask;
}

/**
 * @brief      Clears the pin(s) to a low level output.
 * @param      cfg   Pointer to configuration structure describing the pin.
 * 
 */
__STATIC_INLINE void GPIO_OutClr(const gpio_cfg_t *cfg)
{
    MXC_GPIO->out_val[cfg->port] &= ~(cfg->mask);
}

/**
 * @brief      Gets the pin(s) output state.
 * @param      cfg   Pointer to configuration structure describing the pin.
 * 
 * @return     The state of the requested pin. 
 * 
 */
__STATIC_INLINE uint32_t GPIO_OutGet(const gpio_cfg_t *cfg)
{
    return (MXC_GPIO->out_val[cfg->port] & cfg->mask);
}

/**
 * @brief      Write the pin(s) to a desired output level.
 * @param      cfg   Pointer to configuration structure describing the pin.
 * @param      val   Desired output level of the pin(s). This will be masked
 *                   with the configuration mask.
 *                   
 */
__STATIC_INLINE void GPIO_OutPut(const gpio_cfg_t *cfg, uint32_t val)
{
    MXC_GPIO->out_val[cfg->port] = (MXC_GPIO->out_val[cfg->port] & ~cfg->mask) | (val & cfg->mask);
}

/**
 * @brief      Toggles the the pin(s) output level.
 * @param      cfg   Pointer to configuration structure describing the pin.
 * 
 */
__STATIC_INLINE void GPIO_OutToggle(const gpio_cfg_t *cfg)
{
    MXC_GPIO->out_val[cfg->port] ^= cfg->mask;
}

/**
 * @brief      Configure GPIO interrupt(s)
 * @param      cfg   Pointer to configuration structure describing the pin.
 * @param      mode  Requested interrupt mode.
 * 
 */
void GPIO_IntConfig(const gpio_cfg_t *cfg, gpio_int_mode_t mode);

/**
 * @brief      Enables the specified GPIO interrupt
 * @param      cfg   Pointer to configuration structure describing the pin.
 * 
 */
__STATIC_INLINE void GPIO_IntEnable(const gpio_cfg_t *cfg)
{
    MXC_GPIO->inten[cfg->port] |= cfg->mask;
}

/**
 * @brief      Disables the specified GPIO interrupt.
 * @param      cfg   Pointer to configuration structure describing the pin.
 * 
 */
__STATIC_INLINE void GPIO_IntDisable(const gpio_cfg_t *cfg)
{
    MXC_GPIO->inten[cfg->port] &= ~cfg->mask;
}

/**
 * @brief      Gets the interrupt(s) status on a GPIO pin.
 * @param      cfg   Pointer to configuration structure describing the pin
 *                   for which the status is being requested. 
 * 
 * @return     The requested interrupt status.
 * 
 */
__STATIC_INLINE uint32_t GPIO_IntStatus(const gpio_cfg_t *cfg)
{
    return (MXC_GPIO->intfl[cfg->port] & cfg->mask);
}

/**
 * @brief      Clears the interrupt(s) status on a GPIO pin.
 * @param      cfg   Pointer to configuration structure describing the pin
 *                   to clear the interrupt state of. 
 * 
 */
__STATIC_INLINE void GPIO_IntClr(const gpio_cfg_t *cfg)
{
    MXC_GPIO->intfl[cfg->port] = cfg->mask;
}

/**
 * @brief      Type alias for a GPIO callback function with prototype:
 * @code
 *      void callback_fn(void *cbdata);
 * @endcode
 * @param      cbdata  A void pointer to the data type as registered when
 *                     @c GPIO_RegisterCallback() was called.
 *                     
 */
typedef void (*gpio_callback_fn)(void *cbdata);

/**
 * @brief      Registers a callback for the interrupt on a given port and pin. 
 * @param      cfg       Pointer to configuration structure describing the pin
 * @param      callback  A pointer to a function of type #gpio_callback_fn.
 * @param      cbdata    The parameter to be passed to the callback function, #gpio_callback_fn, when an interrupt occurs. 
 * 
 */
void GPIO_RegisterCallback(const gpio_cfg_t *cfg, gpio_callback_fn callback, void *cbdata);

/**
 * @brief      GPIO IRQ Handler. @note If a callback is registered for a given
 *             interrupt, the callback function will be called. 
 *             
 * @param      port number of the port that generated the interrupt service routine.  
 * 
 */
void GPIO_Handler(unsigned int port);

/**@} end of group gpio */

#ifdef __cplusplus
}
#endif

#endif /* _GPIO_H_ */
