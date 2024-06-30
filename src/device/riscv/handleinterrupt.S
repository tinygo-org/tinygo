// We don't support RV32E at the moment.
#if !defined(__riscv_32e)

#ifdef __riscv_flen
#define NREG 48
#define LFREG flw
#define SFREG fsw
#else
#define NREG 16
#endif

#if __riscv_xlen==64
#define REGSIZE 8
#define SREG sd
#define LREG ld
#else
#define REGSIZE 4
#define SREG sw
#define LREG lw
#endif

.section .text.handleInterruptASM
.global handleInterruptASM
.type handleInterruptASM,@function
handleInterruptASM:
    // Save and restore all registers, because the hardware only saves/restores
    // the pc.
    // Note: we have to do this in assembly because the "interrupt"="machine"
    // attribute is broken in LLVM: https://bugs.llvm.org/show_bug.cgi?id=42984
    addi    sp, sp, -NREG*REGSIZE
    SREG    ra, 0*REGSIZE(sp)
    SREG    t0, 1*REGSIZE(sp)
    SREG    t1, 2*REGSIZE(sp)
    SREG    t2, 3*REGSIZE(sp)
    SREG    a0, 4*REGSIZE(sp)
    SREG    a1, 5*REGSIZE(sp)
    SREG    a2, 6*REGSIZE(sp)
    SREG    a3, 7*REGSIZE(sp)
    SREG    a4, 8*REGSIZE(sp)
    SREG    a5, 9*REGSIZE(sp)
    SREG    a6, 10*REGSIZE(sp)
    SREG    a7, 11*REGSIZE(sp)
    SREG    t3, 12*REGSIZE(sp)
    SREG    t4, 13*REGSIZE(sp)
    SREG    t5, 14*REGSIZE(sp)
    SREG    t6, 15*REGSIZE(sp)
#ifdef __riscv_flen
    SFREG   f0, (0  + 16)*REGSIZE(sp)
    SFREG   f1, (1  + 16)*REGSIZE(sp)
    SFREG   f2, (2  + 16)*REGSIZE(sp)
    SFREG   f3, (3  + 16)*REGSIZE(sp)
    SFREG   f4, (4  + 16)*REGSIZE(sp)
    SFREG   f5, (5  + 16)*REGSIZE(sp)
    SFREG   f6, (6  + 16)*REGSIZE(sp)
    SFREG   f7, (7  + 16)*REGSIZE(sp)
    SFREG   f8, (8  + 16)*REGSIZE(sp)
    SFREG   f9, (9  + 16)*REGSIZE(sp)
    SFREG   f10,(10 + 16)*REGSIZE(sp)
    SFREG   f11,(11 + 16)*REGSIZE(sp)
    SFREG   f12,(12 + 16)*REGSIZE(sp)
    SFREG   f13,(13 + 16)*REGSIZE(sp)
    SFREG   f14,(14 + 16)*REGSIZE(sp)
    SFREG   f15,(15 + 16)*REGSIZE(sp)
    SFREG   f16,(16 + 16)*REGSIZE(sp)
    SFREG   f17,(17 + 16)*REGSIZE(sp)
    SFREG   f18,(18 + 16)*REGSIZE(sp)
    SFREG   f19,(19 + 16)*REGSIZE(sp)
    SFREG   f20,(20 + 16)*REGSIZE(sp)
    SFREG   f21,(21 + 16)*REGSIZE(sp)
    SFREG   f22,(22 + 16)*REGSIZE(sp)
    SFREG   f23,(23 + 16)*REGSIZE(sp)
    SFREG   f24,(24 + 16)*REGSIZE(sp)
    SFREG   f25,(25 + 16)*REGSIZE(sp)
    SFREG   f26,(26 + 16)*REGSIZE(sp)
    SFREG   f27,(27 + 16)*REGSIZE(sp)
    SFREG   f28,(28 + 16)*REGSIZE(sp)
    SFREG   f29,(29 + 16)*REGSIZE(sp)
    SFREG   f30,(30 + 16)*REGSIZE(sp)
    SFREG   f31,(31 + 16)*REGSIZE(sp)
#endif
    call    handleInterrupt
#ifdef __riscv_flen
    LFREG   f0, (31 + 16)*REGSIZE(sp)
    LFREG   f1, (30 + 16)*REGSIZE(sp)
    LFREG   f2, (29 + 16)*REGSIZE(sp)
    LFREG   f3, (28 + 16)*REGSIZE(sp)
    LFREG   f4, (27 + 16)*REGSIZE(sp)
    LFREG   f5, (26 + 16)*REGSIZE(sp)
    LFREG   f6, (25 + 16)*REGSIZE(sp)
    LFREG   f7, (24 + 16)*REGSIZE(sp)
    LFREG   f8, (23 + 16)*REGSIZE(sp)
    LFREG   f9, (22 + 16)*REGSIZE(sp)
    LFREG   f10,(21 + 16)*REGSIZE(sp)
    LFREG   f11,(20 + 16)*REGSIZE(sp)
    LFREG   f12,(19 + 16)*REGSIZE(sp)
    LFREG   f13,(18 + 16)*REGSIZE(sp)
    LFREG   f14,(17 + 16)*REGSIZE(sp)
    LFREG   f15,(16 + 16)*REGSIZE(sp)
    LFREG   f16,(15 + 16)*REGSIZE(sp)
    LFREG   f17,(14 + 16)*REGSIZE(sp)
    LFREG   f18,(13 + 16)*REGSIZE(sp)
    LFREG   f19,(12 + 16)*REGSIZE(sp)
    LFREG   f20,(11 + 16)*REGSIZE(sp)
    LFREG   f21,(10 + 16)*REGSIZE(sp)
    LFREG   f22,(9 + 16)*REGSIZE(sp)
    LFREG   f23,(8 + 16)*REGSIZE(sp)
    LFREG   f24,(7 + 16)*REGSIZE(sp)
    LFREG   f25,(6 + 16)*REGSIZE(sp)
    LFREG   f26,(5 + 16)*REGSIZE(sp)
    LFREG   f27,(4 + 16)*REGSIZE(sp)
    LFREG   f28,(3 + 16)*REGSIZE(sp)
    LFREG   f29,(2 + 16)*REGSIZE(sp)
    LFREG   f30,(1 + 16)*REGSIZE(sp)
    LFREG   f31,(0 + 16)*REGSIZE(sp)
#endif
    LREG    t6, 15*REGSIZE(sp)
    LREG    t5, 14*REGSIZE(sp)
    LREG    t4, 13*REGSIZE(sp)
    LREG    t3, 12*REGSIZE(sp)
    LREG    a7, 11*REGSIZE(sp)
    LREG    a6, 10*REGSIZE(sp)
    LREG    a5, 9*REGSIZE(sp)
    LREG    a4, 8*REGSIZE(sp)
    LREG    a3, 7*REGSIZE(sp)
    LREG    a2, 6*REGSIZE(sp)
    LREG    a1, 5*REGSIZE(sp)
    LREG    a0, 4*REGSIZE(sp)
    LREG    t2, 3*REGSIZE(sp)
    LREG    t1, 2*REGSIZE(sp)
    LREG    t0, 1*REGSIZE(sp)
    LREG    ra, 0*REGSIZE(sp)
    addi    sp, sp, NREG*REGSIZE
    mret

#endif
