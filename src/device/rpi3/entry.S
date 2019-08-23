.macro  kernel_entry
sub     sp, sp, #256
stp     x0, x1, [sp, #16 * 0]
stp     x2, x3, [sp, #16 * 1]
stp     x4, x5, [sp, #16 * 2]
stp     x6, x7, [sp, #16 * 3]
stp     x8, x9, [sp, #16 * 4]
stp     x10, x11, [sp, #16 * 5]
stp     x12, x13, [sp, #16 * 6]
stp     x14, x15, [sp, #16 * 7]
stp     x16, x17, [sp, #16 * 8]
stp     x18, x19, [sp, #16 * 9]
stp     x20, x21, [sp, #16 * 10]
stp     x22, x23, [sp, #16 * 11]
stp     x24, x25, [sp, #16 * 12]
stp     x26, x27, [sp, #16 * 13]
stp     x28, x29, [sp, #16 * 14]
str     x30, [sp, #16 * 15]
.endm

.macro  kernel_exit
ldp     x0, x1, [sp, #16 * 0]
ldp     x2, x3, [sp, #16 * 1]
ldp     x4, x5, [sp, #16 * 2]
ldp     x6, x7, [sp, #16 * 3]
ldp     x8, x9, [sp, #16 * 4]
ldp     x10, x11, [sp, #16 * 5]
ldp     x12, x13, [sp, #16 * 6]
ldp     x14, x15, [sp, #16 * 7]
ldp     x16, x17, [sp, #16 * 8]
ldp     x18, x19, [sp, #16 * 9]
ldp     x20, x21, [sp, #16 * 10]
ldp     x22, x23, [sp, #16 * 11]
ldp     x24, x25, [sp, #16 * 12]
ldp     x26, x27, [sp, #16 * 13]
ldp     x28, x29, [sp, #16 * 14]
ldr     x30, [sp, #16 * 15]
add     sp, sp, #256
eret
.endm

.macro xxxold_handle_invalid_entry type
kernel_entry
mov	x0, #\type
mrs	x1, esr_el1
mrs	x2, elr_el1
bl	unexpectedInterrupt
b abort
.endm

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
