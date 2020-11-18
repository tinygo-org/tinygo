/*******************************************************************************
 * Copyright (C) 2017 Maxim Integrated Products, Inc., All Rights Reserved.
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
 ******************************************************************************/

#ifndef _SPI_H_INCLUDED
#define _SPI_H_INCLUDED

#include <Arduino.h>
#include <spim.h>

// SPI_HAS_TRANSACTION means SPI has
//   - beginTransaction()
//   - endTransaction()
//   - usingInterrupt()
//   - SPISetting(clock, bitOrder, dataMode)
#define SPI_HAS_TRANSACTION 1

// SPI_HAS_NOTUSINGINTERRUPT means that SPI has notUsingInterrupt() method
#define SPI_HAS_NOTUSINGINTERRUPT 1

// Transfer Modes
#define SPI_MODE0 0x00
#define SPI_MODE1 0x01
#define SPI_MODE2 0x02
#define SPI_MODE3 0x03

// Clock Dividers
typedef enum {
    SPI_CLOCK_DIV2      = CLKMAN_SCALE_DIV_2,
    SPI_CLOCK_DIV4      = CLKMAN_SCALE_DIV_4,
    SPI_CLOCK_DIV8      = CLKMAN_SCALE_DIV_8,
    SPI_CLOCK_DIV16     = CLKMAN_SCALE_DIV_16,
    SPI_CLOCK_DIV32     = CLKMAN_SCALE_DIV_32,
    SPI_CLOCK_DIV64     = CLKMAN_SCALE_DIV_64,
    SPI_CLOCK_DIV128    = CLKMAN_SCALE_DIV_128,
    SPI_CLOCK_DIV256    = CLKMAN_SCALE_DIV_256,
};

class SPISettings {
public:
    SPISettings(uint32_t clock = 1000000, uint8_t bo = MSBFIRST, uint8_t mode = SPI_MODE0):
        transferMode(mode),
        s_bitOrder(bo),
        clk(clock)
    { /* This is intentionally left blank */ }

private:
    uint32_t clk;
    uint8_t s_bitOrder;
    uint8_t transferMode;   // Data transfer mode, 4 SPI modes
    
    friend class SPIClass;
};

class SPIClass {
public:
    SPIClass(uint32_t index);

    // Configuration Functions
    void setClockDivider(uint8_t divider);
    void setDataMode(uint8_t mode);
    void setBitOrder(uint8_t bo);
    
    // Transaction Functions
    void begin(void);
    void end(void);
    void beginTransaction(SPISettings);
    void endTransaction(void);
    void usingInterrupt(uint8_t);
    void notUsingInterrupt(uint8_t);

    // Transfer Functions
    uint8_t transfer(uint8_t);
    uint16_t transfer16(uint16_t);
    void transfer(void *buf, size_t count);

private:
    bool initialized;
    volatile uint8_t bitOrder;
    uint8_t intPortMask = 0;
    uint32_t sumIntNum = 0;
    uint32_t currentSPIClk = 5000000;
    uint32_t idx;
    
    mxc_spim_regs_t *spim;
    
    int modifyClk(uint32_t);
};

extern SPIClass SPI0;
#if (MXC_CFG_SPIM_INSTANCES > 1)
extern SPIClass SPI1;
#endif
#if (MXC_CFG_SPIM_INSTANCES > 2)
extern SPIClass SPI2;
#endif
#if (MXC_CFG_SPIM_INSTANCES > 3)
extern SPIClass SPI3;
#endif
#if (MXC_CFG_SPIM_INSTANCES > 4)
extern SPIClass SPI4;
#endif
#if (MXC_CFG_SPIM_INSTANCES > 5)
extern SPIClass SPI5;
#endif
extern SPIClass& SPI;

#endif // _SPI_H_INCLUDED
