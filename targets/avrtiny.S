; TODO: remove these in LLVM 15
#define __tmp_reg__ r16
#define __zero_reg__ r17

; Startup code
.section .text.__vector_RESET
.global  __vector_RESET
__vector_RESET:
    clr  __zero_reg__ ; this register is expected to be 0 by the C calling convention

    ; Keep the stack pointer at the default location, which is RAMEND.

; Initialize .data section.
.section .text.__do_copy_data,"ax",@progbits
.global __do_copy_data
__do_copy_data:
    ldi  xl, lo8(__data_start)
    ldi  xh, hi8(__data_start)
    ldi  yl, lo8(__data_end)
    ldi  yh, hi8(__data_end)
    ldi  zl, lo8(__data_load_start)
    ldi  zh, hi8(__data_load_start)
1: ; loop
    cp   xl, yl         ; if x == y
    cpc  xh, yh
    breq 2f             ; goto end
    ld   r16, Z+        ; r0 = *(z++)
    st   X+, r16        ; *(x++) = r0
    rjmp 1b             ; goto loop
2: ; end

; Initialize .bss section.
.section .text.__do_clear_bss,"ax",@progbits
.global __do_clear_bss
__do_clear_bss:
    ldi  xl, lo8(__bss_start)
    ldi  xh, hi8(__bss_start)
    ldi  yl, lo8(__bss_end)
1: ; loop
    cp   xl, yl           ; if x == y
    breq 2f               ; goto end
    st   X+, __zero_reg__ ; *(x++) = 0
    rjmp 1b               ; goto loop
2: ; end
