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
*/

#include "Arduino.h"

#ifdef __cplusplus
extern "C"{
#endif

uint32_t shiftIn( uint32_t ulDataPin, uint32_t ulClockPin, uint32_t ulBitOrder )
{
	uint8_t value = 0 ;
	uint8_t i ;

	for ( i=0 ; i < 8 ; ++i )
    {
		digitalWrite( ulClockPin, HIGH ) ;

		if ( ulBitOrder == LSBFIRST )
        {
			value |= digitalRead( ulDataPin ) << i ;
        }
		else
        {
			value |= digitalRead( ulDataPin ) << (7 - i) ;
        }

		digitalWrite( ulClockPin, LOW ) ;
	}

	return value ;
}

void shiftOut( uint32_t ulDataPin, uint32_t ulClockPin, uint32_t ulBitOrder, uint32_t ulVal )
{
	uint8_t i ;

	for ( i=0 ; i < 8 ; i++ )
    {
		if ( ulBitOrder == LSBFIRST )
        {
			digitalWrite( ulDataPin, !!(ulVal & (1 << i)) ) ;
        }
		else	
        {
			digitalWrite( ulDataPin, !!(ulVal & (1 << (7 - i))) ) ;
        }

		digitalWrite( ulClockPin, HIGH ) ;
		digitalWrite( ulClockPin, LOW ) ;		
	}
}

#ifdef __cplusplus
} // extern "C"
#endif
