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
 * $Date: 2016-04-27 11:55:43 -0500 (Wed, 27 Apr 2016) $ 
 * $Revision: 22541 $
 *
 ******************************************************************************/
 
/**
 * @file    i2cs.h
 * @brief   I2C Slave driver header file.
 */

#ifndef _I2CS_H_
#define _I2CS_H_

/***** Includes *****/
#include "mxc_config.h"
#include "mxc_sys.h"
#include "mxc_assert.h"
#include "i2cs_regs.h"

#ifdef __cplusplus
 extern "C" {
#endif

/***** Definitions *****/

#define I2CS_BUFFER_SIZE    32

/// @brief I2CS frequencies.
typedef enum {
    I2CS_SPEED_100KHZ = 0,
    I2CS_SPEED_400KHZ = 1
} i2cs_speed_t;

/// @brief I2CS addressing modes.
typedef enum {
    I2CS_ADDR_8 = 0,
    I2CS_ADDR_10 = MXC_F_I2CS_DEV_ID_TEN_BIT_ID_MODE
} i2cs_addr_t;

/***** Globals *****/

/***** Function Prototypes *****/

/**
 * @brief   Initialize I2CS module.
 * @param   i2cs        Pointer to I2CS regs.
 * @param   sys_cfg     Pointer to I2CS system configuration.
 * @param   speed       I2CS frequency.        
 * @param   address     I2CS address.        
 * @param   addr_len    I2CS address length.        
 * @returns #E_NO_ERROR if everything is successful, error if unsuccessful.
 */
int I2CS_Init(mxc_i2cs_regs_t *i2cs, const sys_cfg_i2cs_t *sys_cfg, i2cs_speed_t speed, 
    uint16_t address, i2cs_addr_t addr_len);

/**
 * @brief   Shutdown I2CS module.
 * @param   i2cs    Pointer to I2CS regs.
 * @returns #E_NO_ERROR if everything is successful, error if unsuccessful.
 */
int I2CS_Shutdown(mxc_i2cs_regs_t *i2cs);

/**
 * @brief   I2CS interrupt handler.
 * @details This function should be called by the application from the interrupt
 *          handler if I2CS interrupts are enabled. Alternately, this function
 *          can be periodically called by the application if I2CS interrupts are
 *          disabled.
 * @param   i2cs        Pointer to I2CS regs.
 */
void I2CS_Handler(mxc_i2cs_regs_t *i2cs);

/**
 * @brief   Register a callback for a given range of bytes.
 * @details Registering a callback here will cause the slave to interrupt when 
 *          the master has updated a specified byte.
 * @param   i2cs        Pointer to I2CS regs.
 * @param   addr        Index to start assigning this callback.
 * @param   len         Number of bytes to assign to this callback.
 * @param   callback    callback function to be called.
 */
void I2CS_RegisterCallback(mxc_i2cs_regs_t *i2cs, uint8_t addr, void (*callback)(uint8_t));

/**
 * @brief   Write I2CS data to a given byte.
 * @details The slave has a buffer of registers that the external master can read.
 *          Use this function to write the data into the registers.
 * @param   i2cs        Pointer to I2CS regs.
 * @param   addr        Index to write to.
 * @param   data        Data to be written.
 */
__STATIC_INLINE void I2CS_Write(mxc_i2cs_regs_t *i2cs, uint8_t addr, uint8_t data)
{
    // Make sure we don't overflow
    MXC_ASSERT(addr < MXC_CFG_I2CS_BUFFER_SIZE);
    i2cs->data_byte[addr] = ((i2cs->data_byte[addr] & ~MXC_F_I2CS_DATA_BYTE_DATA_FIELD) |
        (data << MXC_F_I2CS_DATA_BYTE_DATA_FIELD_POS));
}

/**
 * @brief   Read I2CS data from a given byte.
 * @details The slave has a buffer of registers that the external master can read.
 *          Use this function to read the data from the registers.
 * @param   i2cs        Pointer to I2CS regs.
 * @param   addr        Index to read from.
 * @returns Data contained in given addr register.
 */
__STATIC_INLINE uint8_t I2CS_Read(mxc_i2cs_regs_t *i2cs, uint8_t addr)
{
    // Make sure we don't overflow
    MXC_ASSERT(addr < MXC_CFG_I2CS_BUFFER_SIZE);
    return ((i2cs->data_byte[addr] & MXC_F_I2CS_DATA_BYTE_DATA_FIELD) >> 
        MXC_F_I2CS_DATA_BYTE_DATA_FIELD_POS);
}

/**
 * @brief   Set the given range of bytes to read only.
 * @details These bytes will be flagged as read only. The slave will NACK to the
 *          master if it attempts to write these bits.
 * @param   i2cs        Pointer to I2CS regs.
 * @param   addr        Index to start setting RO flag.
 * @param   len         Number of bytes to assign RO flag.
 */
__STATIC_INLINE void I2CS_SetRO(mxc_i2cs_regs_t *i2cs, uint8_t addr)
{
    // Make sure we don't overflow
    MXC_ASSERT(addr < MXC_CFG_I2CS_BUFFER_SIZE);
    i2cs->data_byte[addr] |= MXC_F_I2CS_DATA_BYTE_READ_ONLY_FL;
}

/**
 * @brief   Clear the given range of bytes from read only.
 * @param   i2cs        Pointer to I2CS regs.
 * @param   addr        Index to start clearing RO flag.
 * @param   len         Number of bytes to clear RO flag.
 */
__STATIC_INLINE void I2CS_ClearRO(mxc_i2cs_regs_t *i2cs, uint8_t addr)
{
    // Make sure we don't overflow
    MXC_ASSERT(addr < MXC_CFG_I2CS_BUFFER_SIZE);
    i2cs->data_byte[addr] &= ~MXC_F_I2CS_DATA_BYTE_READ_ONLY_FL;
}

#ifdef __cplusplus
}
#endif

#endif /* _I2CS_H_ */
