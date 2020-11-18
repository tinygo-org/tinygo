/*
  Copyright (c) 2011-2012 Arduino.  All right reserved.

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
#include "gpio.h"

// Interrupt Handler for each port
void GPIO_P0_IRQHandler(void){ GPIO_Handler(PORT_0); }
void GPIO_P1_IRQHandler(void){ GPIO_Handler(PORT_1); }
void GPIO_P2_IRQHandler(void){ GPIO_Handler(PORT_2); }
void GPIO_P3_IRQHandler(void){ GPIO_Handler(PORT_3); }
void GPIO_P4_IRQHandler(void){ GPIO_Handler(PORT_4); }
void GPIO_P5_IRQHandler(void){ GPIO_Handler(PORT_5); }
void GPIO_P6_IRQHandler(void){ GPIO_Handler(PORT_6); }

// Returns the interrupt number associated with the pin
uint32_t digitalPinToInterrupt(uint32_t pin)
{
    if (IS_DIGITAL(pin)) {
        return GET_PIN_IRQ(pin);
    }
    return NOT_AN_INTERRUPT;
}

void attachInterrupt(uint32_t pin, void (*callback)(void), uint32_t mode)
{
    if (IS_DIGITAL(pin)) {
        // Set the callback function
        GPIO_RegisterCallback(GET_PIN_CFG(pin), (void (*)(void *))callback, NULL);

        // Configure and Enable interrupt
        GPIO_IntConfig(GET_PIN_CFG(pin),GET_GPIO_INT_MODE(mode));
        GPIO_IntEnable(GET_PIN_CFG(pin));
        NVIC_EnableIRQ(GET_PIN_IRQ(pin));
    }
}

void detachInterrupt(uint32_t pin)
{
    GPIO_IntDisable(GET_PIN_CFG(pin));
}
