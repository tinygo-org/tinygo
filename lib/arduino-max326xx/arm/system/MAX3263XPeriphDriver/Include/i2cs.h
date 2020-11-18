/**
 * @file
 * @brief   I2CS (Inter-Integrated Circuit Slave) function prototypes and data types.
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
 * $Date: 2017-02-16 12:07:30 -0600 (Thu, 16 Feb 2017) $ 
 * $Revision: 26468 $
 *
 *************************************************************************** */

/* Define to prevent redundant inclusion */
#ifndef _I2CS_H_
#define _I2CS_H_

/* **** Includes **** */
#include "mxc_config.h"
#include "mxc_sys.h"
#include "mxc_assert.h"
#include "i2cs_regs.h"

#ifdef __cplusplus
extern "C" {
#endif

/**
 * @ingroup periphlibs
 * @defgroup i2cs I2C Slave
 * @brief I2C Slave Peripheral API
 * @{
 */

/* **** Definitions **** */
/**
 * Internal buffer size for storing I2C Slave Messages 
 */
#define I2CS_BUFFER_SIZE    32          

/**
 * Enumeration type to select supported I2CS frequencies. 
 */
typedef enum {
    I2CS_SPEED_100KHZ = 0,              /**< Use to select a bus communication speed of 100 kHz. */
    I2CS_SPEED_400KHZ = 1               /**< Use to select a bus communication speed of 400 kHz. */
} i2cs_speed_t;

/**
 * Enumeration type to select the I2CS addressing mode. 
 */
typedef enum {
    I2CS_ADDR_8 = 0,                                    /**< Sets the slave address mode to 8-bits (7-bits address plus read/write bit). */
    I2CS_ADDR_10 = MXC_F_I2CS_DEV_ID_TEN_BIT_ID_MODE    /**< Sets the slave address mode to 10-bits. */
} i2cs_addr_t;

/**
 * Type alias for an I2CS callback function that will be called when a given byte is updated by the Master, see I2CS_RegisterCallback(mxc_i2cs_regs_t *i2cs, uint8_t addr, i2cs_callback_fn callback).
 * @details The function prototype for implementing callback_fn is:
 * @code
 * void func(uint8_t addr);
 * @endcode         
 */
typedef void (*i2cs_callback_fn)(uint8_t error_code);
/* **** Globals **** */

/* **** Function Prototypes **** */

/**
 * @brief      Initialize I2CS module.
 * @param      i2cs      Pointer to I2CS regs.
 * @param      sys_cfg   Pointer to I2CS system configuration, see
 *                       #sys_cfg_i2cs_t.
 * @param      speed     I2CS frequency.
 * @param      address   I2CS address.
 * @param      addr_len  I2CS address length.
 * @return     #E_NO_ERROR if everything is successful or an
 *             @ref MXC_Error_Codes "error code" if unsuccessful.
 *             
 */
int I2CS_Init(mxc_i2cs_regs_t *i2cs, const sys_cfg_i2cs_t *sys_cfg, i2cs_speed_t speed, uint16_t address, i2cs_addr_t addr_len);

/**
 * @brief      Shutdown I2CS module.
 * @param      i2cs  Pointer to I2CS regs.
 * @return     #E_NO_ERROR if everything is successful or an 
 *             @ref MXC_Error_Codes "error code" if unsuccessful.
 */
int I2CS_Shutdown(mxc_i2cs_regs_t *i2cs);

/**
 * @brief      I2CS interrupt handler.
 * @details    This function should be called by the application from the
 *             interrupt handler if I2CS interrupts are enabled. Alternately,
 *             this function can be periodically called by the application if
 *             I2CS interrupts are disabled.
 *             
 * @param      i2cs  Pointer to I2CS regs.
 */
void I2CS_Handler(mxc_i2cs_regs_t *i2cs);

/**
 * @brief      Register a callback that is triggered by an update of a specified
 *             byte.
 * @details    Registering a callback causes the slave to interrupt when the
 *             master has updated a specified byte.
 *
 * @param      i2cs      Pointer to the I2CS register structure, see
 *                       #mxc_i2cs_regs_t.
 * @param      addr      Index to trigger a call to the #i2cs_callback_fn.
 * @param      callback  callback function of type #i2cs_callback_fn to be called
 *                       when the addr being written by the master matches \c addr.
 */
void I2CS_RegisterCallback(mxc_i2cs_regs_t *i2cs, uint8_t addr, i2cs_callback_fn callback);

/**
 * @brief      Write I2CS data to a given byte.
 * @details    The slave has a buffer of registers that the external master can
 *             read. Use this function to write data into a specified
 *             address/index.
 *
 * @param      i2cs  Pointer to I2CS regs.
 * @param      addr  Address/Index to write.
 * @param      data  Data to be written.
 */
__STATIC_INLINE void I2CS_Write(mxc_i2cs_regs_t *i2cs, uint8_t addr, uint8_t data)
{
    // Make sure we don't overflow
    MXC_ASSERT(addr < MXC_CFG_I2CS_BUFFER_SIZE);
    i2cs->data_byte[addr] = ((i2cs->data_byte[addr] & ~MXC_F_I2CS_DATA_BYTE_DATA_FIELD) |
        (data << MXC_F_I2CS_DATA_BYTE_DATA_FIELD_POS));
}

/**
 * @brief      Read I2CS data from a given address .
 * @details    The slave has a buffer of registers that the external master can
 *             read. Use this function to read the data from the registers.
 *
 * @param      i2cs  Pointer to I2CS regs.
 * @param      addr  Address/Index to read from.
 *
 * @return     Data contained in requested @c addr register.
 */
__STATIC_INLINE uint8_t I2CS_Read(mxc_i2cs_regs_t *i2cs, uint8_t addr)
{
    // Make sure we don't overflow
    MXC_ASSERT(addr < MXC_CFG_I2CS_BUFFER_SIZE);
    return ((i2cs->data_byte[addr] & MXC_F_I2CS_DATA_BYTE_DATA_FIELD) >> 
        MXC_F_I2CS_DATA_BYTE_DATA_FIELD_POS);
}

/**
 * @brief      Set the given index to read only (RO).
 * @details    This index will be flagged as read only. The slave will NACK the
 *             master if it attempts to write this location. Multiple calls with
 *             different index/address values will yield multiple read-only
 *             locations within the slave register set.
 *
 * @param      i2cs  Pointer to I2CS regs.
 * @param      addr  Address/Index of the byte to set to RO.
 */
__STATIC_INLINE void I2CS_SetRO(mxc_i2cs_regs_t *i2cs, uint8_t addr)
{
    // Make sure we don't overflow
    MXC_ASSERT(addr < MXC_CFG_I2CS_BUFFER_SIZE);
    i2cs->data_byte[addr] |= MXC_F_I2CS_DATA_BYTE_READ_ONLY_FL;
}

/**
 * @brief      Sets the given address to R/W.
 * @param      i2cs  Pointer to I2CS regs.
 * @param      addr  Index to start clearing RO flag.
 */
__STATIC_INLINE void I2CS_ClearRO(mxc_i2cs_regs_t *i2cs, uint8_t addr)
{
    // Make sure we don't overflow
    MXC_ASSERT(addr < MXC_CFG_I2CS_BUFFER_SIZE);
    i2cs->data_byte[addr] &= ~MXC_F_I2CS_DATA_BYTE_READ_ONLY_FL;
}

/**@} end of group i2cs */

#ifdef __cplusplus
}
#endif

#endif /* _I2CS_H_ */
