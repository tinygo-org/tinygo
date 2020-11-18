/**
 * @file
 * @brief      This file contains the function implementations for the Flash
 *             Controller (FLC) peripheral module.
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
 * $Date: 2016-09-09 11:48:21 -0500 (Fri, 09 Sep 2016) $
 * $Revision: 24338 $
 *
 *************************************************************************** */

/* **** Includes **** */
#include "mxc_config.h"
#include "flc.h"


/**
 * @ingroup flc
 * @{
 */
    
/* **** Definitions **** */

/* **** Globals **** */

/* **** Functions **** */

/* ************************************************************************* */
#if defined ( __GNUC__ )
#undef IAR_PRAGMAS //Make sure this is not defined for GCC
#endif

#if IAR_PRAGMAS
// IAR memory section declaration for the in-system flash programming functions to be loaded in RAM.
#pragma section=".flashprog"
#endif
#if defined ( __GNUC__ )
__attribute__ ((section(".flashprog")))
#endif
/**
 * @brief      Return the status of the busy state of the flash controller. 
 *
 * @return     0 Flash Controller is idle.
 * @return     Non-zero indicates the flash controller is performing an 
 *             erase or write request. 
 */
__STATIC_INLINE int FLC_Busy(void)
{
    return (MXC_FLC->ctrl & (MXC_F_FLC_CTRL_WRITE | MXC_F_FLC_CTRL_MASS_ERASE | MXC_F_FLC_CTRL_PAGE_ERASE));
}

/* ************************************************************************* */
#if IAR_PRAGMAS
// IAR memory section declaration for the in-system flash programming functions to be loaded in RAM.
#pragma section=".flashprog"
#endif
#if defined ( __GNUC__ )
__attribute__ ((section(".flashprog")))
#endif
int FLC_Init(void)
{
    /* Check if the flash controller is busy */
    if (FLC_Busy()) {
        return E_BUSY;
    }

    /* Enable automatic calculation of the clock divider to generate a 1MHz clock from the APB clock */
    MXC_FLC->perform |= MXC_F_FLC_PERFORM_AUTO_CLKDIV;

    /* The flash controller will stall any reads while flash operations are in
     * progress. Disable the legacy failure detection logic that would flag reads
     * during flash operations as errors.
     */
    MXC_FLC->perform |= MXC_F_FLC_PERFORM_EN_PREVENT_FAIL;

    return E_NO_ERROR;
}

/* ************************************************************************* */
#if IAR_PRAGMAS
// IAR memory section declaration for the in-system flash programming functions to be loaded in RAM.
#pragma section=".flashprog"
#endif
#if defined ( __GNUC__ )
__attribute__ ((section(".flashprog")))
#endif
int FLC_PageErase(uint32_t address, uint8_t erase_code, uint8_t unlock_key)
{
    /* Check if the flash controller is busy */
    if (FLC_Busy()) {
        return E_BUSY;
    }

    /* Clear stale errors. Interrupt flags can only be written to zero, so this is safe */
    MXC_FLC->intr &= ~MXC_F_FLC_INTR_FAILED_IF;

    /* Unlock flash */
    MXC_FLC->ctrl = (MXC_FLC->ctrl & ~MXC_F_FLC_CTRL_FLSH_UNLOCK) |
                    ((unlock_key << MXC_F_FLC_CTRL_FLSH_UNLOCK_POS) & MXC_F_FLC_CTRL_FLSH_UNLOCK);

    /* Write the Erase Code */
    MXC_FLC->ctrl = (MXC_FLC->ctrl & ~MXC_F_FLC_CTRL_ERASE_CODE) |
                    ((erase_code << MXC_F_FLC_CTRL_ERASE_CODE_POS) & MXC_F_FLC_CTRL_ERASE_CODE);

    /* Erase the request page */
    MXC_FLC->faddr = address;
    MXC_FLC->ctrl |= MXC_F_FLC_CTRL_PAGE_ERASE;

    /* Wait until flash operation is complete */
    while (FLC_Busy());

    /* Lock flash */
    MXC_FLC->ctrl &= ~(MXC_F_FLC_CTRL_FLSH_UNLOCK | MXC_F_FLC_CTRL_ERASE_CODE);

    /* Check for failures */
    if (MXC_FLC->intr & MXC_F_FLC_INTR_FAILED_IF) {
        /* Interrupt flags can only be written to zero, so this is safe */
        MXC_FLC->intr &= ~MXC_F_FLC_INTR_FAILED_IF;
        return E_UNKNOWN;
    }

    return E_NO_ERROR;
}

/* ************************************************************************* */
#if IAR_PRAGMAS
// IAR memory section declaration for the in-system flash programming functions to be loaded in RAM.
#pragma section=".flashprog"
#endif
#if defined ( __GNUC__ )
__attribute__ ((section(".flashprog")))
#endif
int FLC_Write(uint32_t address, const void *data, uint32_t length, uint8_t unlock_key)
{
    uint32_t *ptr = (uint32_t*)data;

    /* Can only write in full word units */
    if ((address & 3) || (length & 3)) {
        return E_BAD_PARAM;
    }

    if (length == 0) {
        /* Nothing to do */
        return E_NO_ERROR;
    }

    /* Check if the flash controller is busy */
    if (FLC_Busy()) {
        return E_BUSY;
    }

    /* Clear stale errors. Interrupt flags can only be written to zero, so this is safe */
    MXC_FLC->intr &= ~MXC_F_FLC_INTR_FAILED_IF;

    /* Unlock flash */
    MXC_FLC->ctrl = (MXC_FLC->ctrl & ~MXC_F_FLC_CTRL_FLSH_UNLOCK) |
                    ((unlock_key << MXC_F_FLC_CTRL_FLSH_UNLOCK_POS) & MXC_F_FLC_CTRL_FLSH_UNLOCK);

    /* Set the address to write and enable auto increment */
    MXC_FLC->faddr = address;
    MXC_FLC->ctrl |= MXC_F_FLC_CTRL_AUTO_INCRE_MODE;
    uint32_t write_cmd = MXC_FLC->ctrl | MXC_F_FLC_CTRL_WRITE;

    for (; length > 0; length -= 4) {
        /* Perform the write */
        MXC_FLC->fdata = *ptr++;
        MXC_FLC->ctrl = write_cmd;
        while (FLC_Busy());
    }

    /* Lock flash */
    MXC_FLC->ctrl &= ~MXC_F_FLC_CTRL_FLSH_UNLOCK;

    /* Check for failures */
    if (MXC_FLC->intr & MXC_F_FLC_INTR_FAILED_IF) {
        /* Interrupt flags can only be written to zero, so this is safe */
        MXC_FLC->intr &= ~MXC_F_FLC_INTR_FAILED_IF;
        return E_UNKNOWN;
    }

    return E_NO_ERROR;
}

/* ************************************************************************* */
#if IAR_PRAGMAS
// IAR memory section declaration for the in-system flash programming functions to be loaded in RAM.
#pragma section=".flashprog"
#endif
#if defined ( __GNUC__ )
__attribute__ ((section(".flashprog")))
#endif
int FLC_MassErase(uint8_t erase_code, uint8_t unlock_key)
{
    /* Check if the flash controller is busy */
    if (FLC_Busy()) {
        return E_BUSY;
    }

    /* Clear stale errors. Interrupt flags can only be written to zero, so this is safe */
    MXC_FLC->intr &= ~MXC_F_FLC_INTR_FAILED_IF;

    /* Unlock flash */
    MXC_FLC->ctrl = (MXC_FLC->ctrl & ~MXC_F_FLC_CTRL_FLSH_UNLOCK) |
                    ((unlock_key << MXC_F_FLC_CTRL_FLSH_UNLOCK_POS) & MXC_F_FLC_CTRL_FLSH_UNLOCK);

    /* Write the Erase Code */
    MXC_FLC->ctrl = (MXC_FLC->ctrl & ~MXC_F_FLC_CTRL_ERASE_CODE) |
                    ((erase_code << MXC_F_FLC_CTRL_ERASE_CODE_POS) & MXC_F_FLC_CTRL_ERASE_CODE);

    /* Start the mass erase */
    MXC_FLC->ctrl |= MXC_F_FLC_CTRL_MASS_ERASE;

    /* Wait until flash operation is complete */
    while (FLC_Busy());

    /* Lock flash */
    MXC_FLC->ctrl &= ~(MXC_F_FLC_CTRL_FLSH_UNLOCK | MXC_F_FLC_CTRL_ERASE_CODE);

    /* Check for failures */
    if (MXC_FLC->intr & MXC_F_FLC_INTR_FAILED_IF) {
        /* Interrupt flags can only be written to zero, so this is safe */
        MXC_FLC->intr &= ~MXC_F_FLC_INTR_FAILED_IF;
        return E_UNKNOWN;
    }

    return E_NO_ERROR;
}

/**@} end of group flc */
