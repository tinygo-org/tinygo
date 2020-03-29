# TinyGo ARM SysTick example w/ condition variable

This example uses the ARM System Timer to awake a goroutine.
That goroutine sends to a channel, and another goroutine toggles an LED on every notification.

Many ARM-based chips have this timer feature.  If you run the example and the
LED blinks, then you have one.

The System Timer runs from a cycle counter.  The more cycles, the slower the
LED will blink.  This counter is 24 bits wide, which places an upper bound on
the number of cycles, and the slowness of the blinking.
