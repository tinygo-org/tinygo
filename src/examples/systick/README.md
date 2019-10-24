# TinyGo ARM SysTick example

This example uses the ARM System Timer to blink an LED.  The timer fires
an interrupt 10 times per second.  The interrupt handler toggles the LED on
and off.

Many ARM-based chips have this timer feature.  If you run the example and the
LED blinks, then you have one.

The System Timer runs from a cycle counter.  The more cycles, the slower the
LED will blink.  This counter is 24 bits wide, which places an upper bound on
the number of cycles, and the slowness of the blinking.
