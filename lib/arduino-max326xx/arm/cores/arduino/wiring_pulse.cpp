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
#include "wiring_private.h"
#include "variant.h"

/* Measures the length (in microseconds) of a pulse on the pin; state is HIGH
 * or LOW, the type of pulse to measure.  Works on pulses from 2-3 microseconds
 * to 3 minutes in length, but must be called at least a few dozen microseconds
 * before the start of the pulse.
 *
 * ATTENTION:
 * This function performs better with short pulses in noInterrupt() context
 */
 

uint32_t pulseIn(uint32_t pin, uint32_t state, uint32_t timeout)
{   
    uint32_t port       = GET_PIN_PORT(pin);
    uint32_t bit        = GET_PIN_MASK(pin);
    uint32_t stateMask  = state ? bit : 0;
    
    uint32_t maxLoops = microsecondsToClockCycles(timeout) / 12;

   	// convert the timeout from microseconds to a number of times through
	// the initial loop; it takes (roughly) 18 clock cycles per iteration.
    uint32_t width = countPulseASM( &MXC_GPIO->in_val[port], bit, stateMask, maxLoops );

	// convert the reading to microseconds. The loop has been determined
	// to be 12(9 for loop + 3 differences in speeds of APB & AHB buses)
    // clock cycles long and have about 8 clocks between the edge
	// and the start of the loop. There will be some error introduced by
	// the interrupt handlers.
	if (width)
		return clockCyclesToMicroseconds(width * 12 + 8);
	else
		return 0;
}

/* Measures the length (in microseconds) of a pulse on the pin; state is HIGH
 * or LOW, the type of pulse to measure.  Works on pulses from 2-3 microseconds
 * to 3 minutes in length, but must be called at least a few dozen microseconds
 * before the start of the pulse.
 *
 * ATTENTION:
 * this function relies on micros() so cannot be used in noInterrupt() context
 */
uint32_t pulseInLong(uint32_t pin, uint32_t state, uint32_t timeout)
{	
    // cache the port and bit of the pin in order to speed up the
	// pulse width measuring loop and achieve finer resolution.  calling
	// digitalRead() instead yields much coarser resolution.
	
    uint32_t port       = GET_PIN_PORT(pin);
	uint32_t bit        = GET_PIN_MASK(pin);
	uint32_t stateMask  = state ? bit : 0;
	
    uint32_t startMicros = micros();

	// wait for any previous pulse to end
	while ((MXC_GPIO->in_val[port] & bit) == stateMask) {
		if (micros() - startMicros > timeout)
			return 0;
	}

	// wait for the pulse to start
	while ((MXC_GPIO->in_val[port] & bit) != stateMask) {
		if (micros() - startMicros > timeout)
			return 0;
	}

	uint32_t start = micros();
	// wait for the pulse to stop
	while ((MXC_GPIO->in_val[port] & bit) == stateMask) {
		if (micros() - startMicros > timeout)
			return 0;
	}
	return micros() - start;
}
