/*
  Copyright (c) 2011 Arduino.  All right reserved.

  This library is free software; you can redistribute it and/or
  modify it under the terms of the GNU Lesser General Public
  License as published by the Free Software Foundation; either
  version 2.1 of the License, or (at your option) any later version.

  This library is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
  See the GNU Lesser General Public License for more details.

  You should have received a copy of the GNU Lesser General Public
  License along with this library; if not, write to the Free Software
  Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA

  Modified 2017 by Maxim Integrated for MAX326xx
*/

#include "Arduino.h"
#include "gpio_regs.h"

#define USE_SYSTEM_CLK 1
#define TICK_RATE_HZ 1000
#define SYS_CLK_HZ RO_FREQ
#define SYSTICK_PERIOD_SYS_CLK (SYS_CLK_HZ/TICK_RATE_HZ)

#ifdef __cplusplus
extern "C" {
#endif

/** Tick Counter in ms */
static volatile uint32_t tickCount = 0;

void init(void)
{
    SYS_SysTick_Config(SYSTICK_PERIOD_SYS_CLK, USE_SYSTEM_CLK);

    // Set all pins to Normal Input Monitoring Mode
    uint8_t port;
    for (port = 0; port < MXC_GPIO_NUM_PORTS; port++) {
        MXC_GPIO->in_mode[port] = 0UL;
    }
}

uint32_t millis( void )
{
    return tickCount;
}

// Interrupt-compatible version of micros
// Theory: repeatedly take readings of SysTick counter, millis counter and SysTick interrupt pending flag.
// When it appears that millis counter and pending is stable and SysTick hasn't rolled over, use these
// values to calculate micros. If there is a pending SysTick, add one to the millis counter in the calculation.
uint32_t micros( void )
{
    uint32_t ticks, ticks2;
    uint32_t pend, pend2;
    uint32_t count, count2;

    ticks2  = SysTick->VAL;
    pend2   = !!((SCB->ICSR & SCB_ICSR_PENDSTSET_Msk) || ((SCB->SHCSR & SCB_SHCSR_SYSTICKACT_Msk)));
    count2  = tickCount;

    do {
        ticks   = ticks2;
        pend    = pend2;
        count   = count2;
        ticks2  = SysTick->VAL;
        pend2   = !!((SCB->ICSR & SCB_ICSR_PENDSTSET_Msk) || ((SCB->SHCSR & SCB_SHCSR_SYSTICKACT_Msk)));
        count2  = tickCount;
    } while ((pend != pend2) || (count != count2) || (ticks < ticks2));

    return ((count + pend) * 1000) + (((SysTick->LOAD - ticks) * (1048576 / (F_CPU / 1000000))) >> 20);
    // this is an optimization to turn a runtime division into two compile-time divisions and
    // a runtime multiplication and shift, saving a few cycles
}

void delay( uint32_t ms )
{
    if (ms == 0)
        return;

    uint32_t start = tickCount;

    while ((tickCount - start) < ms) {
        __WFI();
    }
}

void SysTick_Handler(void)
{
	tickCount++;
}

#ifdef __cplusplus
}
#endif
