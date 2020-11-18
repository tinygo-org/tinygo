/**
 * @file
 * @brief      1-Wire Master (OWM) API Function Implementations.
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
 * $Date: 2016-03-14 10:08:53 -0500 (Mon, 14 Mar 2016) $
 * $Revision: 21855 $
 *
 **************************************************************************** */

/* **** Includes **** */
#include <string.h>
#include "mxc_assert.h"
#include "mxc_sys.h"
#include "owm.h"

/**
 * @ingroup    owm 
 * @{
 */ 
///@cond
/* **** Definitions **** */
#define OWM_CLK_FREQ    1000000 //1-Wire requires 1MHz clock

/* **** Globals **** */
int LastDiscrepancy;
int LastDeviceFlag;

/* **** Functions **** */
static uint8_t CalculateCRC8(uint8_t* data, int len);
static uint8_t update_crc8(uint8_t crc, uint8_t value);
///@endcond


/* ************************************************************************* */
int OWM_Init(mxc_owm_regs_t *owm, const owm_cfg_t *cfg, const sys_cfg_owm_t *sys_cfg)
{
    int err = 0;
    uint32_t owm_clk, clk_div = 0;
    uint32_t ext_pu_mode = 0;
    uint32_t ext_pu_polarity = 0;

    // Check the OWM register pointer is valid
    MXC_ASSERT(MXC_OWM_GET_IDX(owm) >= 0);

    if(cfg == NULL) {
        return E_NULL_PTR;
    }

    // Set system level configurations
    if ((err = SYS_OWM_Init(owm, sys_cfg)) != E_NO_ERROR) {
        return err;
    }

    // Configure clk divisor to get 1MHz OWM clk
    owm_clk = SYS_OWM_GetFreq(owm);

    if(owm_clk == 0) {
        return E_UNINITIALIZED;
    }

    // Return error if clk doesn't divide evenly to 1MHz
    if(owm_clk % OWM_CLK_FREQ) {
        return E_NOT_SUPPORTED;
    }

    clk_div = (owm_clk / (OWM_CLK_FREQ));

    // Can not support lower frequencies
    if(clk_div == 0) {
        return E_NOT_SUPPORTED;
    }

    // Select the PU mode and polarity based on cfg input
    switch(cfg->ext_pu_mode)
    {
        case OWM_EXT_PU_ACT_HIGH:
            ext_pu_mode = MXC_V_OWM_CFG_EXT_PULLUP_MODE_USED;
            ext_pu_polarity = MXC_V_OWM_CTRL_STAT_EXT_PULLUP_POL_ACT_HIGH;
            break;
        case OWM_EXT_PU_ACT_LOW:
            ext_pu_mode = MXC_V_OWM_CFG_EXT_PULLUP_MODE_USED;
            ext_pu_polarity = MXC_V_OWM_CTRL_STAT_EXT_PULLUP_POL_ACT_LOW;
            break;
        case OWM_EXT_PU_UNUSED:
            ext_pu_mode = MXC_V_OWM_CFG_EXT_PULLUP_MODE_UNUSED;
            ext_pu_polarity = MXC_V_OWM_CTRL_STAT_EXT_PULLUP_POL_ACT_HIGH;
            break;
        default:
            return E_BAD_PARAM;
    }

    // Set clk divisor
    owm->clk_div_1us = (clk_div << MXC_F_OWM_CLK_DIV_1US_DIVISOR_POS) & MXC_F_OWM_CLK_DIV_1US_DIVISOR;

    // Set configuration
    owm->cfg = (((cfg->int_pu_en << MXC_F_OWM_CFG_INT_PULLUP_ENABLE_POS) & MXC_F_OWM_CFG_INT_PULLUP_ENABLE) |
               ((ext_pu_mode << MXC_F_OWM_CFG_EXT_PULLUP_MODE_POS) & MXC_F_OWM_CFG_EXT_PULLUP_MODE) |
               ((cfg->long_line_mode << MXC_F_OWM_CFG_LONG_LINE_MODE) & MXC_F_OWM_CFG_LONG_LINE_MODE_POS));

    owm->ctrl_stat = (((ext_pu_polarity << MXC_F_OWM_CTRL_STAT_EXT_PULLUP_POL_POS) & MXC_F_OWM_CTRL_STAT_EXT_PULLUP_POL) |
                     ((cfg->overdrive_spec << MXC_F_OWM_CTRL_STAT_OD_SPEC_MODE_POS) & MXC_F_OWM_CTRL_STAT_OD_SPEC_MODE));

    // Clear all interrupt flags
    owm->intfl = owm->intfl;

    return E_NO_ERROR;
}

/* ************************************************************************* */
int OWM_Shutdown(mxc_owm_regs_t *owm)
{
    int err;

    // Disable and clear interrupts
    owm->inten = 0;
    owm->intfl = owm->intfl;

    // Release IO pins and disable clk
    if ((err = SYS_OWM_Shutdown(owm)) != E_NO_ERROR) {
        return err;
    }

    return E_NO_ERROR;
}

/* ************************************************************************* */
int OWM_Reset(mxc_owm_regs_t *owm)
{
    owm->intfl = MXC_F_OWM_INTFL_OW_RESET_DONE;                 // Clear the reset flag
    owm->ctrl_stat |= MXC_F_OWM_CTRL_STAT_START_OW_RESET;       // Generate a reset pulse
    while((owm->intfl & MXC_F_OWM_INTFL_OW_RESET_DONE) == 0);   // Wait for reset time slot to complete

    return (!!(owm->ctrl_stat & MXC_F_OWM_CTRL_STAT_PRESENCE_DETECT)); // Return presence pulse detect status
}

/* ************************************************************************* */
int OWM_TouchByte(mxc_owm_regs_t *owm, uint8_t data)
{
    owm->cfg &= ~MXC_F_OWM_CFG_SINGLE_BIT_MODE;                                     // Set to 8 bit mode
    owm->intfl = (MXC_F_OWM_INTFL_TX_DATA_EMPTY | MXC_F_OWM_INTFL_RX_DATA_READY);   // Clear the flags
    owm->data = (data << MXC_F_OWM_DATA_TX_RX_POS) & MXC_F_OWM_DATA_TX_RX;          // Write data
    while((owm->intfl & MXC_F_OWM_INTFL_TX_DATA_EMPTY) == 0);   // Wait for data to be sent
    while((owm->intfl & MXC_F_OWM_INTFL_RX_DATA_READY) == 0);   // Wait for data to be read

    return (owm->data >> MXC_F_OWM_DATA_TX_RX_POS) & 0xFF; // Return the data read
}

/* ************************************************************************* */
int OWM_WriteByte(mxc_owm_regs_t *owm, uint8_t data)
{
    // Send one byte of data and verify the data sent = data parameter
    return (OWM_TouchByte(owm, data) == data) ? E_NO_ERROR : E_COMM_ERR;
}

/* ************************************************************************* */
int OWM_ReadByte(mxc_owm_regs_t *owm)
{
    // Read one byte of data
    return OWM_TouchByte(owm, 0xFF);
}

/* ************************************************************************* */
int OWM_TouchBit(mxc_owm_regs_t *owm, uint8_t bit)
{
    MXC_OWM->cfg |= MXC_F_OWM_CFG_SINGLE_BIT_MODE;                                  // Set to 1 bit mode
    owm->intfl = (MXC_F_OWM_INTFL_TX_DATA_EMPTY | MXC_F_OWM_INTFL_RX_DATA_READY);   // Clear the flags
    owm->data = (bit << MXC_F_OWM_DATA_TX_RX_POS) & MXC_F_OWM_DATA_TX_RX;           // Write data
    while((owm->intfl & MXC_F_OWM_INTFL_TX_DATA_EMPTY) == 0);   // Wait for data to be sent
    while((owm->intfl & MXC_F_OWM_INTFL_RX_DATA_READY) == 0);   // Wait for data to be read

    return (owm->data >> MXC_F_OWM_DATA_TX_RX_POS) & 0x1; // Return the bit read
}

/* ************************************************************************* */
int OWM_WriteBit(mxc_owm_regs_t *owm, uint8_t bit)
{
    // Send a bit and verify the bit sent = bit parameter
    return (OWM_TouchBit(owm, bit) == bit) ? E_NO_ERROR : E_COMM_ERR;
}

/* ************************************************************************* */
int OWM_ReadBit(mxc_owm_regs_t *owm)
{
    // Read a bit
    return OWM_TouchBit(owm, 1);
}

/* ************************************************************************* */
int OWM_Write(mxc_owm_regs_t *owm, uint8_t* data, int len)
{
    int num = 0;

    owm->cfg &= ~MXC_F_OWM_CFG_SINGLE_BIT_MODE; // Set to 8 bit mode

    while(num < len) // Loop for number of bytes to write
    {
        owm->intfl = (MXC_F_OWM_INTFL_TX_DATA_EMPTY | MXC_F_OWM_INTFL_RX_DATA_READY | MXC_F_OWM_INTEN_LINE_SHORT);    // Clear the flags
        owm->data = (data[num] << MXC_F_OWM_DATA_TX_RX_POS) & MXC_F_OWM_DATA_TX_RX;   // Write data
        while((owm->intfl & MXC_F_OWM_INTFL_TX_DATA_EMPTY) == 0);                     // Wait for data to be sent
        while((owm->intfl & MXC_F_OWM_INTFL_RX_DATA_READY) == 0);                     // Wait for data to be read

        // Verify data sent is correct
        if(owm->data != data[num]) {
            return E_COMM_ERR;
        }

        // Check error flag
        if(owm->intfl & MXC_F_OWM_INTEN_LINE_SHORT) {
            return E_COMM_ERR; // Wire was low before transaction
        }

        num++; // Keep track of how many bytes written
    }

    return num; // Return number of bytes written
}

/* ************************************************************************* */
int OWM_Read(mxc_owm_regs_t *owm, uint8_t* data, int len)
{
    int num = 0;

    owm->cfg &= ~MXC_F_OWM_CFG_SINGLE_BIT_MODE; // Set to 8 bit mode

    while(num < len) // Loop for number of bytes to read
    {
        owm->intfl = (MXC_F_OWM_INTFL_TX_DATA_EMPTY | MXC_F_OWM_INTFL_RX_DATA_READY | MXC_F_OWM_INTEN_LINE_SHORT);   // Clear the flags
        owm->data = 0xFF;                                                            // Write 0xFF for a read
        while((owm->intfl & MXC_F_OWM_INTFL_TX_DATA_EMPTY) == 0);                    // Wait for data to be sent
        while((owm->intfl & MXC_F_OWM_INTFL_RX_DATA_READY) == 0);                    // Wait for data to be read

        // Check error flag
        if(owm->intfl & MXC_F_OWM_INTEN_LINE_SHORT) {
            return E_COMM_ERR; // Wire was low before transaction
        }

        // Store read data into buffer
        data[num] = (owm->data >> MXC_F_OWM_DATA_TX_RX_POS) & MXC_F_OWM_DATA_TX_RX;

        num++; // Keep track of how many bytes read
    }

    return num; // Return number of bytes read
}

/* ************************************************************************* */
int OWM_ReadROM(mxc_owm_regs_t *owm, uint8_t* ROMCode)
{
    int num_read = 0;

    // Send reset and wait for presence pulse
    if(OWM_Reset(owm))
    {
        // Send Read ROM command code
        if(OWM_WriteByte(owm, READ_ROM_COMMAND) == E_NO_ERROR)
        {
            // Read 8 bytes and store in buffer
            num_read = OWM_Read(owm, ROMCode, 8);

            // Check the number of bytes read
            if(num_read != 8) {
                return E_COMM_ERR;
            }
        }
        else
        {
            // Write failed
            return E_COMM_ERR;
        }
    }
    else
    {
        // No presence pulse
        return E_COMM_ERR;
    }

    return E_NO_ERROR;
}

/* ************************************************************************* */
int OWM_MatchROM(mxc_owm_regs_t *owm, uint8_t* ROMCode)
{
    int num_wrote = 0;

    // Send reset and wait for presence pulse
    if(OWM_Reset(owm))
    {
        // Send match ROM command code
        if(OWM_WriteByte(owm, MATCH_ROM_COMMAND) == E_NO_ERROR)
        {
            // Write 8 bytes in ROMCode buffer
            num_wrote = OWM_Write(owm, ROMCode, 8);

            // Check the number of bytes written
            if(num_wrote != 8) {
                return E_COMM_ERR;
            }
        }
        else
        {
            // Write failed
            return E_COMM_ERR;
        }
    }
    else
    {
        // No presence pulse
        return E_COMM_ERR;
    }

    return E_NO_ERROR;
}

/* ************************************************************************* */
int OWM_ODMatchROM(mxc_owm_regs_t *owm, uint8_t* ROMCode)
{
    int num_wrote = 0;

    // Set to standard speed
    owm->cfg &= ~(MXC_F_OWM_CFG_OVERDRIVE);

    // Send reset and wait for presence pulse
    if(OWM_Reset(owm))
    {
        // Send Overdrive match ROM command code
        if(OWM_WriteByte(owm, OD_MATCH_ROM_COMMAND) == E_NO_ERROR)
        {
            // Set overdrive
            owm->cfg |= MXC_F_OWM_CFG_OVERDRIVE;

            // Write 8 bytes in ROMCode buffer
            num_wrote = OWM_Write(owm, ROMCode, 8);

            // Check the number of bytes written
            if(num_wrote != 8) {
                return E_COMM_ERR;
            }
        }
        else
        {
            // Write failed
            return E_COMM_ERR;
        }
    }
    else
    {
        // No presence pulse
        return E_COMM_ERR;
    }

    return E_NO_ERROR;
}

/* ************************************************************************* */
int OWM_SkipROM(mxc_owm_regs_t *owm)
{
    // Send reset and wait for presence pulse
    if(OWM_Reset(owm))
    {
        // Send skip ROM command code
        return OWM_WriteByte(owm, SKIP_ROM_COMMAND);
    }
    else
    {
        // No presence pulse
        return E_COMM_ERR;
    }
}

/* ************************************************************************* */
int OWM_ODSkipROM(mxc_owm_regs_t *owm)
{
    // Set to standard speed
    owm->cfg &= ~(MXC_F_OWM_CFG_OVERDRIVE);

    // Send reset and wait for presence pulse
    if(OWM_Reset(owm))
    {
        // Send Overdrive skip ROM command code
        if(OWM_WriteByte(owm, OD_SKIP_ROM_COMMAND) == E_NO_ERROR)
        {
            // Set overdrive speed
            owm->cfg |= MXC_F_OWM_CFG_OVERDRIVE;

            return E_NO_ERROR;
        }
        else
        {
            // Write failed
            return E_COMM_ERR;
        }
    }
    else
    {
        // No presence pulse
        return E_COMM_ERR;
    }
}

/* ************************************************************************* */
int OWM_Resume(mxc_owm_regs_t *owm)
{
    // Send reset and wait for presence pulse
    if(OWM_Reset(owm))
    {
        // Send resume command code
        return OWM_WriteByte(owm, RESUME_COMMAND);
    }
    else
    {
        // No presence pulse
        return E_COMM_ERR;
    }
}

/* ************************************************************************* */
int OWM_SearchROM(mxc_owm_regs_t *owm, int newSearch, uint8_t* ROMCode)
{
    int nibble_start_bit = 1;
    int rom_byte_number = 0;
    uint8_t rom_nibble_mask = 0x0F;
    uint8_t search_direction = 0;
    int readValue = 0;
    int sentBits = 0;
    int discrepancy = 0;
    int bit_position = 0;
    int discrepancy_mask = 0;
    int last_zero = 0;
    uint8_t crc8 = 0;
    int search_result = 0;

    // Clear ROM array
    memset(ROMCode, 0x0, 8);

    if(newSearch)
    {
        // Reset all global variables to start search from begining
        LastDiscrepancy = 0;
        LastDeviceFlag = 0;
    }

    // Check if the last call was the last device
    if(LastDeviceFlag)
    {
        // Reset the search
        LastDiscrepancy = 0;
        LastDeviceFlag = 0;
        return 0;
    }

    // Send reset and wait for presence pulse
    if (OWM_Reset(owm))
    {
        // Send the search command
        OWM_WriteByte(owm, SEARCH_ROM_COMMAND);

        // Set search ROM accelerator bit
        owm->ctrl_stat |= MXC_F_OWM_CTRL_STAT_SRA_MODE;

        // Loop until through all ROM bytes 0-7 (this loops 2 times per byte)
        while(rom_byte_number < 8)
        {
            // Each loop finds the discrepancy bits and finds 4 bits (nibble) of the ROM

            // Set the search direction the same as last time for the nibble masked
            search_direction = ROMCode[rom_byte_number] & rom_nibble_mask;

            // If the upper nibble is the mask then shift bits to lower nibble
            if(rom_nibble_mask > 0x0F) {
                search_direction  = search_direction >> 4;
            }

            // Get the last discrepancy bit position relative to the nibble start bit
            bit_position = LastDiscrepancy - nibble_start_bit;

            // Check if last discrepancy is witin this nibble
            if( (bit_position >= 0) && (bit_position < 4) )
            {
                // Last discrepancy is within this nibble
                // Set the bit of the last discrepancy bit
                search_direction |=  (1 << (bit_position));
            }

            // Performs two read bits and a write bit for 4 bits of the ROM
            readValue = OWM_TouchByte(owm, search_direction);
            // Get discrepancy flags
            discrepancy = readValue & 0xF;
            // Get the 4 bits sent to select the ROM
            sentBits = (readValue >> 4) & 0xF;

            // Store the bit location of the MSB discrepancy with sentbit = 0
            if(discrepancy)
            {
                // Initialize bit_position to MSB of nibble
                bit_position = 3;

                while(bit_position >= 0)
                {
                    // Get discrepancy flag of the current bit position
                    discrepancy_mask =  discrepancy & (1 << bit_position);

                    // If there is a discrepancy and the sent bit is 0 save this bit position
                    if( (discrepancy_mask)  && !(sentBits & discrepancy_mask))
                    {
                        last_zero = nibble_start_bit + bit_position;
                        break;
                    }

                    bit_position--;
                }
            }

            // Clear the nibble
            ROMCode[rom_byte_number] &= ~rom_nibble_mask;

            // Store the sentBits in the ROMCode
            if(rom_nibble_mask > 0x0F) {
                ROMCode[rom_byte_number] |= (sentBits << 4);
            }
            else {
                ROMCode[rom_byte_number] |= sentBits;
            }

            // Increment the nibble start bit and shift mask
            nibble_start_bit += 4;
            rom_nibble_mask <<= 4;

            // If the mask is 0 then go to new ROM byte rom_byte_number and reset mask
            if (rom_nibble_mask == 0)
            {
                rom_byte_number++;
                rom_nibble_mask = 0x0F;
            }

        } // End while(rom_byte_number < 8)

        // Clear search ROM accelerator
        owm->ctrl_stat &= ~(MXC_F_OWM_CTRL_STAT_SRA_MODE);

        // Calculate CRC to verify ROM code is correct
        crc8 = CalculateCRC8(ROMCode, 7);

        // If the search was successful then
        if ((nibble_start_bit >= 65) && (crc8 == ROMCode[7]))
        {
             // Search successful so set LastDiscrepancy,LastDeviceFlag,search_result
             LastDiscrepancy = last_zero;

             // Check for last device
             if (LastDiscrepancy == 0) {
                LastDeviceFlag = 1;
             }

             search_result = 1;
        }
    } // End if (OWM_Reset)

    // If no device found then reset counters so next 'search' will be like a first
    if (!search_result || !ROMCode[0])
    {
        LastDiscrepancy = 0;
        LastDeviceFlag = 0;
        search_result = 0;
    }

    return search_result;
}

/*
 * Calcualate CRC8 of the buffer of data provided
 */
uint8_t CalculateCRC8(uint8_t* data, int len)
{
    int i;
    uint8_t crc = 0;

    for(i = 0; i < len; i++)
    {
        crc = update_crc8(crc, data[i]);
    }

    return crc;
}

/*
 * Calculate the CRC8 of the byte value provided with the current crc value
 * provided Returns updated crc value
 */
uint8_t update_crc8(uint8_t crc, uint8_t val)
{
    uint8_t inc, tmp;

    for (inc = 0; inc < 8; inc++)
    {
        tmp = (uint8_t)(crc << 7); // Save X7 bit value
        crc >>= 1; // Shift crc
        if (((tmp >> 7) ^ (val & 0x01)) == 1) // If X7 xor X8 (input data)
        {
            crc ^= 0x8c; // XOR crc with X4 and X5, X1 = X7^X8
            crc |= 0x80; // Carry
        }
        val >>= 1;
    }

    return crc;
}

/**@} end of group owm */
