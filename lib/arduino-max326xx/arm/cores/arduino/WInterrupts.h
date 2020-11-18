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

#ifndef _WIRING_INTERRUPTS_
#define _WIRING_INTERRUPTS_

#include "gpio.h"

// Maps GPIO mode to Maxim compatible GPIO mode
#define GET_GPIO_INT_MODE(m) (gpio_int_mode_t)( (m) == LOW      ? GPIO_INT_LOW_LEVEL    :               \
                                                (m) == CHANGE   ? GPIO_INT_ANY_EDGE     :               \
                                                (m) == RISING   ? GPIO_INT_RISING_EDGE  :               \
                                                (m) == FALLING  ? GPIO_INT_FALLING_EDGE : GPIO_INT_DISABLE)

#ifdef __cplusplus
extern "C" {
#endif

uint32_t digitalPinToInterrupt(uint32_t);
void attachInterrupt(uint32_t, void (*callback)(void), uint32_t);
void detachInterrupt(uint32_t);

#ifdef __cplusplus
}
#endif

#endif /* _WIRING_INTERRUPTS_ */
