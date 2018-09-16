.section .isr
isr:
    rjmp reset

.org 0x18 ; WDT
    rjmp wdt

; Startup code
.section .reset
.org 26
reset:
    clr  r1          ; r1 is expected to be 0 by the C calling convention

    ; Set up the stack pointer.
    ldi  xl, lo8(_stack_top)
    ldi  xh, hi8(_stack_top)
    out  0x3d, xl; SPL
    out  0x3e, xh; SPH

    ; main will be placed right after here by the linker script so there's no
    ; need to jump.


; The only thing this WDT handler really does is disable itself, to get out of
; sleep mode.
.section .text.wdt
wdt:
    push r16

    clr  r16
    wdr            ; Reset watchdog
    out  0x34, r16 ; Clear reset reason (MCUSR)

    ; part 1: set WDCE and WDE to enable editing WDTCSR
    lds  r16, 0x60 ; r16 = WDTCSR
    ori  r16, 0x18 ; r16 |= WDCE | WDE
    sts  0x60, r16 ; WDTCSR = r16

    ; part 2: within 4 clock cycles, set the new value for WDTCSR
    clr  r16
    sts  0x60, r16 ; WDTCSR = 0

    pop  r16
    reti
