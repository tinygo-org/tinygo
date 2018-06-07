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

    ; Zero .bss
clear_bss:
    ldi  xl, lo8(_sbss)
    ldi  xh, hi8(_sbss)
clear_bss_loop:
    ldi  yl, lo8(_ebss)
    ldi  yh, hi8(_ebss)
    cp   xl, yl         ; if x == y
    cpc  xh, yh
    breq clear_bss_end
    st   x+, r1         ; zero byte in *x
    rjmp clear_bss_loop
clear_bss_end:

    ; Set up the stack pointer.
    ldi  xl, lo8(_stack_top)
    ldi  xh, hi8(_stack_top)
    out  0x3d, xl; SPL
    out  0x3e, xh; SPH

    ; Enable interrupts.
    ; TODO: make sure interrupts are started after all initializers have run.
    sei

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
