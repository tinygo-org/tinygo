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

#ifndef _WIRING_ANALOG_
#define _WIRING_ANALOG_

#include "adc.h"

#ifdef __cplusplus
extern "C" {
#endif

#define READ_DEFAULT_RESOLUTION 10
#define READ_MAX_RESOLUTION     16

#define WRITE_DEFAULT_RESOLUTION 8
#define WRITE_MAX_RESOLUTION    16

#define ADC_FULL_SCALE 0x3FFU

#define PWM_FREQUENCY_HZ 16000
#define PWM_RESOLUTION 8

/*
 * \brief Change channel A0 to A4.
 */
__STATIC_INLINE void changeA0_A4()
{
    pinLut[A0].mask = ADC_CH_0_DIV_5;
}

/*
 * \brief Change channel A4 to A0.
 */
__STATIC_INLINE void changeA4_A0()
{
    pinLut[A0].mask = ADC_CH_0;
}

/*
 * \brief Change channel A1 to A5.
 */
__STATIC_INLINE void changeA1_A5()
{
    pinLut[A1].mask = ADC_CH_1_DIV_5;
}

/*
 * \brief Change channel A5 to A1.
 */
__STATIC_INLINE void changeA5_A1()
{
    pinLut[A1].mask = ADC_CH_1;
}

/*
 * \brief Configures the reference voltage used for analog input (i.e. the value used as the top of the input range).
 *
 * \param mode
 */
extern void analogReference(int);

/*
 * \brief Set the resolution of analogRead return values. Default is 10 bits (range from 0 to 1023).
 *
 * \param newRes
 */
extern void analogReadResolution(int);

/*
 * \brief Reads the value from the specified analog pin.
 *
 * \param pin
 *
 * \return Read value from selected pin, if no error.
 */
extern uint32_t analogRead(uint32_t);

/*
 * \brief Set the resolution of analogWrite parameters. Default is 8 bits (range from 0 to 255).
 *
 * \param newRes
 */
extern void analogWriteResolution(int);

/*
 * \brief Writes an analog value (PWM wave) to a pin.
 *
 * \param pin
 * \param value
 */
extern void analogWrite(uint32_t, uint32_t);

#ifdef __cplusplus
}
#endif

#endif /* _WIRING_ANALOG_ */
