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

#include <mxc_sys.h>
#include <spim_regs.h>
#include <variant.h>
#include <SPI.h>

// When interrupt number is not GPIO Interrupt
#define NOT_GPIO_IRQn_MASK (1<<5)

SPIClass::SPIClass(uint32_t index):
    initialized(false),
    idx(index)
{
    spim = MXC_SPIM_GET_SPIM(idx);
}

void SPIClass::setClockDivider(uint8_t divider)
{
    // Changing the divider only, would leave the HI_CLK and LO_CLK cycles unchanged
    modifyClk(SYS_GetFreq(divider));
}

void SPIClass::setDataMode(uint8_t mode)
{
    // Update the SPI data tranfer mode
    uint32_t temp = spim->mstr_cfg & ~MXC_F_SPIM_MSTR_CFG_SPI_MODE;
    temp |= (mode << MXC_F_SPIM_MSTR_CFG_SPI_MODE_POS);
    spim->mstr_cfg = temp;
}

void SPIClass::setBitOrder(uint8_t bo)
{
    bitOrder = bo;
}

void SPIClass::begin()
{
    sys_cfg_spim_t sys_cfg;
    spim_cfg_t cfg;

    if (!initialized) {
        // Default Peripheral Config
        cfg.mode = SPI_MODE0;
        cfg.ssel_pol = 0;
        cfg.baud = currentSPIClk;

        // Configure the SPI Master
        switch(idx) {
            case 0: sys_cfg.io_cfg = (ioman_cfg_t)IOMAN_SPIM0(1, 0, 0, 0, 0, 0, 0, 1); break;
            case 1: sys_cfg.io_cfg = (ioman_cfg_t)IOMAN_SPIM1(1, 0, 0, 0, 0, 1); break;

            #if defined(MAX32620) || defined(MAX32630)
            case 2: sys_cfg.io_cfg = (ioman_cfg_t)IOMAN_SPIM2(IOMAN_MAP_B, 1, 0, 0, 0, 0, 0, 0, 1); break;
            #endif

            #ifdef MAX32625
            case 2: sys_cfg.io_cfg = (ioman_cfg_t)IOMAN_SPIM2(1, 0, 0, 0, 0, 0, 1); break;
            #endif
        }

        sys_cfg.clk_scale = CLKMAN_SCALE_AUTO;

        SPIM_Init(spim, &cfg, &sys_cfg);
        initialized = true;
    }
}

void SPIClass::end()
{
    SPIM_Shutdown(spim);
    initialized = false;
}

void SPIClass::beginTransaction(SPISettings settings)
{
    // Disable interrupt(s)
    if (intPortMask) {
        // Disable all interrupts
        if (intPortMask & NOT_GPIO_IRQn_MASK) {
            noInterrupts();
        } else {
            // Disable individual GPIO port interrupt
            if (intPortMask & (1 << PORT_0)) NVIC_DisableIRQ(GPIO_P0_IRQn);
            if (intPortMask & (1 << PORT_1)) NVIC_DisableIRQ(GPIO_P1_IRQn);
            if (intPortMask & (1 << PORT_2)) NVIC_DisableIRQ(GPIO_P2_IRQn);
            if (intPortMask & (1 << PORT_3)) NVIC_DisableIRQ(GPIO_P3_IRQn);
            if (intPortMask & (1 << PORT_4)) NVIC_DisableIRQ(GPIO_P4_IRQn);
        }
    }

    // Set SPI clock frequency
    if (currentSPIClk != settings.clk) {
        modifyClk(settings.clk);
        currentSPIClk = settings.clk;
    }

    // Set the SPI mode
    setDataMode(settings.transferMode);

    // Set the bit order
    bitOrder = settings.s_bitOrder;
}

int SPIClass::modifyClk(uint32_t baud)
{
    uint8_t clk_scale, clocks;
    uint32_t min_baud, spim_clk;
    // Calculate and update the clock divider for the given baud rate
    clk_scale = CLKMAN_SCALE_DISABLED;
    do {
        // Dividing by 2 times maximum possible clocks of HI_CLK and LO_CLK
        min_baud = ((SystemCoreClock >> clk_scale++) / (2 *
        ((MXC_F_SPIM_MSTR_CFG_SCK_HI_CLK >> MXC_F_SPIM_MSTR_CFG_SCK_HI_CLK_POS)+1)));
    } while (baud < min_baud && clk_scale < CLKMAN_SCALE_AUTO);

    // Check if new baud rate is too low to reach
    if (baud < min_baud)
        return E_BAD_STATE;

    CLKMAN_SetClkScale((clkman_clk_t)(CLKMAN_CLK_SPIM0 + idx), (clkman_scale_t)clk_scale);

    // Calculate and update SPI clock frequency
    spim_clk = SYS_SPIM_GetFreq(spim); // Get frequency with updated clock scale
    if (spim_clk == 0)
        return E_BAD_PARAM;

    // Calculate the clocks for HI_CLK and LO_CLK
    clocks = (spim_clk / (2*baud)); // 2, considering Hi and Low clock

    if (clocks == 0 || clocks > 16) // 16 max clock for Hi or Low
        return E_BAD_PARAM;

    // 16 = 0, in the 4 bit field for HI_CLK and LO_CLK
    if (clocks == 16)
        clocks = 0;

    // Set baud
    uint32_t temp = spim->mstr_cfg &
            ~( MXC_F_SPIM_MSTR_CFG_SCK_HI_CLK | MXC_F_SPIM_MSTR_CFG_SCK_LO_CLK);
    temp |= ((clocks << MXC_F_SPIM_MSTR_CFG_SCK_HI_CLK_POS) |
             (clocks << MXC_F_SPIM_MSTR_CFG_SCK_LO_CLK_POS));
    spim->mstr_cfg = temp;
}

void SPIClass::endTransaction(void)
{
    if (intPortMask) {
        // Enable all interrupts which were disabled in beginTransaction
        if (intPortMask & NOT_GPIO_IRQn_MASK) {
            interrupts();
        } else {
            // Enable individual GPIO port interrupt
            if (intPortMask & (1 << PORT_0)) NVIC_EnableIRQ(GPIO_P0_IRQn);
            if (intPortMask & (1 << PORT_1)) NVIC_EnableIRQ(GPIO_P1_IRQn);
            if (intPortMask & (1 << PORT_2)) NVIC_EnableIRQ(GPIO_P2_IRQn);
            if (intPortMask & (1 << PORT_3)) NVIC_EnableIRQ(GPIO_P3_IRQn);
            if (intPortMask & (1 << PORT_4)) NVIC_EnableIRQ(GPIO_P4_IRQn);
        }
    }
}

void SPIClass::usingInterrupt(uint8_t intNum)
{
    switch (intNum) {
        // Individual GPIO port interrupt
        case GPIO_P0_IRQn: intPortMask |= (1 << PORT_0); break;
        case GPIO_P1_IRQn: intPortMask |= (1 << PORT_1); break;
        case GPIO_P2_IRQn: intPortMask |= (1 << PORT_2); break;
        case GPIO_P3_IRQn: intPortMask |= (1 << PORT_3); break;
        case GPIO_P4_IRQn: intPortMask |= (1 << PORT_4); break;
        // default, all other interrupts
        default:intPortMask |= NOT_GPIO_IRQn_MASK;
                sumIntNum += intNum;
                break;
    }
}

void SPIClass::notUsingInterrupt(uint8_t intNum)
{
    if (intPortMask & NOT_GPIO_IRQn_MASK) {
        sumIntNum -= intNum;
        if (!sumIntNum) intPortMask &= ~NOT_GPIO_IRQn_MASK;
    } else {
        switch (intNum) {
            // Individual GPIO port interrupt
            case GPIO_P0_IRQn: intPortMask &= ~(1 << PORT_0); break;
            case GPIO_P1_IRQn: intPortMask &= ~(1 << PORT_1); break;
            case GPIO_P2_IRQn: intPortMask &= ~(1 << PORT_2); break;
            case GPIO_P3_IRQn: intPortMask &= ~(1 << PORT_3); break;
            case GPIO_P4_IRQn: intPortMask &= ~(1 << PORT_4); break;
        }
    }
}

uint8_t SPIClass::transfer(uint8_t value)
{
    spim_req_t req;

    req.tx_data = &value;
    req.rx_data = &value;
    req.ssel = 0;
    req.deass = 0;
    req.width = SPIM_WIDTH_1;
    req.len = sizeof(value);
    req.callback = NULL;

    if (bitOrder == LSBFIRST) {
        value = __RBIT(value) >> 24;
    }

    SPIM_Trans(spim, &req);

    return value;   // Received data
}

uint16_t SPIClass::transfer16(uint16_t value)
{
    uint8_t temp;

    // Least Significant Byte is sent 1st to ensure correct endianness
    // Send LSB
    temp = transfer(value);
    value = (value & 0XFF00) | temp;

    // Send MSB
    temp = transfer(value >> 8);
    value = (temp << 8) | (value & 0x00FF);

    return value;
}

void SPIClass::transfer(void *buf, size_t count)
{
    spim_req_t req;
    uint8_t n = 1;
    uint8_t *ptr = (uint8_t *)buf;
    uint32_t num;

    req.tx_data = ptr;
    req.rx_data = ptr;
    req.ssel = 0;
    req.deass = 0;
    req.width = SPIM_WIDTH_1;
    req.len = count;
    req.callback = NULL;

    // Changing the bit order of individual byte
    if (bitOrder == LSBFIRST) {
        while (n <= count) {
            num = *ptr;
            *ptr++ = __RBIT(num) >> 24;
            n++;
        }
    }
// TODO disable interrupt
    SPIM_Trans(spim, &req);
// TODO enable interrupts

}

SPIClass SPI0 = SPIClass(0);
#if (MXC_CFG_SPIM_INSTANCES > 1)
SPIClass SPI1 = SPIClass(1);
#endif
#if (MXC_CFG_SPIM_INSTANCES > 2)
SPIClass SPI2 = SPIClass(2);
#endif
#if (MXC_CFG_SPIM_INSTANCES > 3)
SPIClass SPI3 = SPIClass(3);
#endif
#if (MXC_CFG_SPIM_INSTANCES > 4)
SPIClass SPI4 = SPIClass(4);
#endif
#if (MXC_CFG_SPIM_INSTANCES > 5)
SPIClass SPI5 = SPIClass(5);
#endif
SPIClass& SPI = CONCAT(SPI, DEFAULT_SPIM_PORT);
