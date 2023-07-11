//go:build avr && attiny85

package avr

// The attiny85 have only TIMSK when other AVR boards have TIMSK0 and TIMSK1, etc.
// Create an alias of TIMSK, TIMSK_OCIE0A, TIMSK_OCIE0B and TIMSK_TOIE0 so we can use
// common code to mask interrupts.

var TIMSK0 = TIMSK

const (
	TIMSK0_OCIE0A = TIMSK_OCIE0A
	TIMSK0_OCIE0B = TIMSK_OCIE0B
	TIMSK0_TOIE0  = TIMSK_TOIE0
)
