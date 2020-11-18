/**
 * @file 
 * @brief      This file contains the implementations of the vendor
 *             defined NVIC function.
 *             
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
 * $Date: 2016-09-28 10:27:00 -0500 (Wed, 28 Sep 2016) $
 * $Revision: 24512 $
 *
 *************************************************************************** */

/**
 * @ingroup mxc_nvic
 * @{
 */

/* **** Includes **** */
#include "mxc_config.h"
#include <string.h>
#include "nvic_table.h"

/* RAM vector_table needs to be aligned with the size of the vector table */
/* TODO: Use MXC_IRQ_COUNT to automatically set this alignment per DUI0553A 4.3.4 */
/*       Vector Table Offset which requires the next-power-of-two alignment. This */
/*       can be calculated by 4*pow(2,ceil(log2(MXC_IRQ_COUNT))) */
#if defined ( __ICCARM__ )
#pragma data_alignment = 512
#define __isr_vector __vector_table
#else
__attribute__ ((aligned (512)))
#endif
static void (*ramVectorTable[MXC_IRQ_COUNT])(void);

/* **** Definitions **** */

/* **** Globals **** */

/* **** Functions **** */

/* ************************************************************************* */
void NVIC_SetRAM(void)
{
    /* should be defined in starup_<device>.S */
    extern uint32_t __isr_vector;

    memcpy(&ramVectorTable, &__isr_vector, sizeof(ramVectorTable));
    SCB->VTOR = (uint32_t)&ramVectorTable;
}

/* ************************************************************************* */
int NVIC_SetVector(IRQn_Type irqn, void(*irq_handler)(void))
{
    int index = irqn + 16;  /* offset for externals */

    /* If not copied, do copy */
    if(SCB->VTOR != (uint32_t)&ramVectorTable) {
        NVIC_SetRAM();
    }

    ramVectorTable[index] = irq_handler;
    NVIC_EnableIRQ(irqn);

    return 0;
}
/**@} end of group mxc_nvic */
