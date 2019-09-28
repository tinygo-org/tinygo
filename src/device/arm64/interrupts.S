//
// This sets up all 16 exception vectors and maps all of them to the symbol
// handle_invalid_entry which will be called when the particular exception
// happens. This call will have three parameters, the number of the exception
// (see list below), the exception syndrome register value, and the exception
// link register value.  You must clear the exception before calling eret.
//
// This assumes you are running in EL1 as would be the case of used start.S.
//
// The values in this file do not matter unless you enable interrupts.
//
// Each handler in this file, e.g. irq_el1h_handler, is defined as a weak
// symbol.  By defining such a symbol in your program, you can force your
// code to be called instead of handle_invalid_entry and with the same
// three parameters as above.

.macro handle_invalid_entry type
mov	x0, #\type
mrs	x1, esr_el1
mrs	x2, elr_el1
bl	unexpectedInterrupt
b abort
.endm

.macro	ventry	label symname
.align	7
.weak \symname
.set \symname, \label
b \symname
.endm

/*
* Exception vectors.
*/
.align	11
.globl vectors
vectors:
ventry	sync_invalid_el1t	sync_el1t_handler		// Synchronous EL1t
ventry	irq_invalid_el1t	irq_el1t_handler		// IRQ EL1t
ventry	fiq_invalid_el1t	fiq_el1t_handler		// FIQ EL1t
ventry	error_invalid_el1t error_el1t_handler			// Error EL1t

ventry	sync_invalid_el1h sync_el1h_handler			// Synchronous EL1h
ventry	irq_invalid_el1h	irq_el1h_handler		// IRQ EL1h
ventry	fiq_invalid_el1h	fiq_el1h_handler		// FIQ EL1h
ventry	error_invalid_el1h error_el1h_handler			// Error EL1h

ventry	sync_invalid_el0_64 sync_el0_64_handler			// Synchronous 64-bit EL0
ventry	irq_invalid_el0_64	irq_el0_64_handler		// IRQ 64-bit EL0
ventry	fiq_invalid_el0_64	fiq_el0_64_handler		// FIQ 64-bit EL0
ventry	error_invalid_el0_64 error_el0_64_handler			// Error 64-bit EL0

ventry	sync_invalid_el0_32	sync_el0_32_handler		// Synchronous 32-bit EL0
ventry	irq_invalid_el0_32 irq_el0_32_handler		// IRQ 32-bit EL0
ventry	fiq_invalid_el0_32	fiq_el0_32_handler		// FIQ 32-bit EL0
ventry	error_invalid_el0_32	error_el0_32_handler		// Error 32-bit EL0

sync_invalid_el1t:
handle_invalid_entry  0

irq_invalid_el1t:
handle_invalid_entry  1

fiq_invalid_el1t:
handle_invalid_entry  2

error_invalid_el1t:
handle_invalid_entry  3

sync_invalid_el1h:
handle_invalid_entry  4

irq_invalid_el1h:
handle_invalid_entry  5

fiq_invalid_el1h:
handle_invalid_entry  6

error_invalid_el1h:
handle_invalid_entry  7

sync_invalid_el0_64:
handle_invalid_entry  8

irq_invalid_el0_64:
handle_invalid_entry  9

fiq_invalid_el0_64:
handle_invalid_entry  10

error_invalid_el0_64:
handle_invalid_entry  11

sync_invalid_el0_32:
handle_invalid_entry  12

irq_invalid_el0_32:
handle_invalid_entry  13

fiq_invalid_el0_32:
handle_invalid_entry  14

error_invalid_el0_32:
handle_invalid_entry  15
