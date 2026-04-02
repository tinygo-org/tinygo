
// Xtensa interrupt/exception vector table for the ESP32-S3.
//
// The ESP32-S3 uses an Xtensa LX7 core with the windowed register ABI.
// Interrupt vectors are placed at fixed offsets from the VECBASE special
// register.  We only handle level-1 (user) interrupts for now.
//
// Vector offsets (from ESP32-S3 core-isa.h XCHAL definitions):
//   0x000  Window overflow  4
//   0x040  Window underflow 4
//   0x080  Window overflow  8
//   0x0C0  Window underflow 8
//   0x100  Window overflow  12
//   0x140  Window underflow 12
//   0x180  Level-2 interrupt
//   0x1C0  Level-3 interrupt
//   0x200  Level-4 interrupt
//   0x240  Level-5 interrupt
//   0x280  Debug exception (level-6)
//   0x2C0  NMI (level-7)
//   0x300  Kernel exception
//   0x340  User exception (level-1 interrupt)
//   0x3C0  Double exception

// PS register field definitions.
#define PS_WOE           0x00040000
#define PS_EXCM          0x00000010
#define PS_INTLEVEL_MASK 0x0000000F

// -----------------------------------------------------------------------
// Vector table — must be aligned to 0x400 (1024 bytes).
// -----------------------------------------------------------------------
    .section .text.exception_vectors,"ax"
    .global _vector_table
    .balign 0x400
_vector_table:

// -----------------------------------------------------------------------
// Offset 0x000 — Window overflow 4
// -----------------------------------------------------------------------
    .org _vector_table + 0x000
_window_overflow4:
    s32e    a0, a5, -16
    s32e    a1, a5, -12
    s32e    a2, a5, -8
    s32e    a3, a5, -4
    rfwo

// -----------------------------------------------------------------------
// Offset 0x040 — Window underflow 4
// -----------------------------------------------------------------------
    .org _vector_table + 0x040
_window_underflow4:
    l32e    a0, a5, -16
    l32e    a1, a5, -12
    l32e    a2, a5, -8
    l32e    a3, a5, -4
    rfwu

// -----------------------------------------------------------------------
// Offset 0x080 — Window overflow 8
// -----------------------------------------------------------------------
    .org _vector_table + 0x080
_window_overflow8:
    s32e    a0, a9, -16
    l32e    a0, a1, -12
    s32e    a1, a9, -12
    s32e    a2, a9, -8
    s32e    a3, a9, -4
    s32e    a4, a0, -32
    s32e    a5, a0, -28
    s32e    a6, a0, -24
    s32e    a7, a0, -20
    rfwo

// -----------------------------------------------------------------------
// Offset 0x0C0 — Window underflow 8
// -----------------------------------------------------------------------
    .org _vector_table + 0x0C0
_window_underflow8:
    l32e    a0, a9, -16
    l32e    a1, a9, -12
    l32e    a2, a9, -8
    l32e    a7, a1, -12
    l32e    a3, a9, -4
    l32e    a4, a7, -32
    l32e    a5, a7, -28
    l32e    a6, a7, -24
    l32e    a7, a7, -20
    rfwu

// -----------------------------------------------------------------------
// Offset 0x100 — Window overflow 12
// -----------------------------------------------------------------------
    .org _vector_table + 0x100
_window_overflow12:
    s32e    a0,  a13, -16
    l32e    a0,  a1,  -12
    s32e    a1,  a13, -12
    s32e    a2,  a13, -8
    s32e    a3,  a13, -4
    s32e    a4,  a0,  -48
    s32e    a5,  a0,  -44
    s32e    a6,  a0,  -40
    s32e    a7,  a0,  -36
    s32e    a8,  a0,  -32
    s32e    a9,  a0,  -28
    s32e    a10, a0,  -24
    s32e    a11, a0,  -20
    rfwo

// -----------------------------------------------------------------------
// Offset 0x140 — Window underflow 12
// -----------------------------------------------------------------------
    .org _vector_table + 0x140
_window_underflow12:
    l32e    a0,  a13, -16
    l32e    a1,  a13, -12
    l32e    a2,  a13, -8
    l32e    a11, a1,  -12
    l32e    a3,  a13, -4
    l32e    a4,  a11, -48
    l32e    a5,  a11, -44
    l32e    a6,  a11, -40
    l32e    a7,  a11, -36
    l32e    a8,  a11, -32
    l32e    a9,  a11, -28
    l32e    a10, a11, -24
    l32e    a11, a11, -20
    rfwu

// -----------------------------------------------------------------------
// Offset 0x180 — Level-2 interrupt (stub — loops forever)
// -----------------------------------------------------------------------
    .org _vector_table + 0x180
_level2_vector:
    j _level2_vector

// -----------------------------------------------------------------------
// Offset 0x1C0 — Level-3 interrupt (stub — loops forever)
// -----------------------------------------------------------------------
    .org _vector_table + 0x1C0
_level3_vector:
    j _level3_vector

// -----------------------------------------------------------------------
// Offset 0x200 — Level-4 interrupt (stub — loops forever)
// -----------------------------------------------------------------------
    .org _vector_table + 0x200
_level4_vector:
    j _level4_vector

// -----------------------------------------------------------------------
// Offset 0x240 — Level-5 interrupt (stub — loops forever)
// -----------------------------------------------------------------------
    .org _vector_table + 0x240
_level5_vector:
    j _level5_vector

// -----------------------------------------------------------------------
// Offset 0x280 — Debug exception / level-6 (stub — loops forever)
// -----------------------------------------------------------------------
    .org _vector_table + 0x280
_debug_vector:
    j _debug_vector

// -----------------------------------------------------------------------
// Offset 0x2C0 — NMI / level-7 (stub — loops forever)
// -----------------------------------------------------------------------
    .org _vector_table + 0x2C0
_nmi_vector:
    j _nmi_vector

// -----------------------------------------------------------------------
// Offset 0x300 — Kernel exception (stub — loops forever)
// -----------------------------------------------------------------------
    .org _vector_table + 0x300
_kernel_vector:
    j _kernel_vector

// -----------------------------------------------------------------------
// Offset 0x340 — User exception / level-1 interrupt
//
// Save a0 and jump to the full handler below the vector table.
// -----------------------------------------------------------------------
    .org _vector_table + 0x340
    .global _level1_vector
_level1_vector:
    wsr     a0, EXCSAVE1        // save a0 — only scratch register available
    j       _handle_level1      // jump to full handler (PC-relative, no literal pool)

// -----------------------------------------------------------------------
// Offset 0x3C0 — Double exception (stub — loops forever)
// -----------------------------------------------------------------------
    .org _vector_table + 0x3C0
_double_vector:
    j _double_vector

// -----------------------------------------------------------------------
// Level-1 interrupt handler — lives outside the vector table so there
// is no 64-byte size constraint.
//
// Saves the interrupted context on the current stack, clears PS.EXCM
// (so window overflow/underflow work), calls the Go handleInterrupt
// dispatcher, restores context, and returns via rfe.
//
// We call handleInterrupt via callx4 (window rotation by 4).  This is
// required because:
//   - callx0 does not set PS.CALLINC, so the Go function's "entry"
//     instruction would use whatever CALLINC the interrupted code left,
//     causing incorrect window rotation and a garbage stack pointer.
//   - callx0 puts the return address in a0 with the raw PC (0x42xxx for
//     flash), whose top 2 bits (01) cause retw to decrement WindowBase
//     by 1 even though nothing was incremented.
//
// With callx4, CALLINC is explicitly set to 1 and the return address
// in a4 has the top 2 bits set to 01 — matching the window rotation
// that entry performs.  After retw, WindowBase is correctly restored.
// Our a0..a3 (including a1, the frame pointer) are NOT in the callee's
// register window (callee uses physical regs +4..+19), so a1 is
// preserved across the call without needing EXCSAVE1.
// -----------------------------------------------------------------------
// Literal data for l32r (must be at a lower address than the l32r).
    .balign 4
.LhandleInterrupt_addr:
    .word   handleInterrupt

    .global _handle_level1
_handle_level1:
    // --- allocate 96-byte exception frame on the interrupted stack ---
    // Layout (offsets from a1 after adjustment):
    //   0: a0   4: a1(orig)  8: a2  12: a3  16: a4  20: a5
    //  24: a6  28: a7  32: a8  36: a9  40: a10 44: a11
    //  48: a12 52: a13 56: a14 60: a15
    //  64: SAR 68: EPC1  72: PS
    addi    a0, a1, -96         // a0 = new frame pointer
    s32i    a1, a0, 4           // save original a1 (SP)
    mov     a1, a0              // a1 = frame pointer

    rsr     a0, EXCSAVE1        // recover original a0
    s32i    a0, a1, 0           // save original a0

    // Save general registers a2..a15.
    s32i    a2,  a1, 8
    s32i    a3,  a1, 12
    s32i    a4,  a1, 16
    s32i    a5,  a1, 20
    s32i    a6,  a1, 24
    s32i    a7,  a1, 28
    s32i    a8,  a1, 32
    s32i    a9,  a1, 36
    s32i    a10, a1, 40
    s32i    a11, a1, 44
    s32i    a12, a1, 48
    s32i    a13, a1, 52
    s32i    a14, a1, 56
    s32i    a15, a1, 60

    // Save special registers.
    rsr     a2, SAR
    s32i    a2, a1, 64
    rsr     a2, EPC1
    s32i    a2, a1, 68

    // Clear PS.EXCM (bit 4) so window overflow/underflow exceptions work
    // during the Go call. Set PS.INTLEVEL=1 to prevent re-entry of
    // level-1 interrupts.
    rsr     a2, PS
    s32i    a2, a1, 72          // save PS (with EXCM=1 set by hardware)
    movi    a3, ~0x1F           // mask: clear INTLEVEL (bits 0-3) + EXCM (bit 4)
    and     a2, a2, a3
    movi    a3, 1               // INTLEVEL = 1
    or      a2, a2, a3
    wsr     a2, PS
    rsync

    // Call the Go interrupt dispatcher via callx4.
    // callx4 explicitly sets PS.CALLINC=1 and puts the return address
    // (with top 2 bits = 01) in a4.  After entry rotates the window by
    // 4, the callee sees:  a0 = our a4 (return addr), a1 = our a5 - N.
    // We set a5 = our frame pointer so the callee gets a valid stack.
    mov     a5, a1
    l32r    a2, .LhandleInterrupt_addr
    callx4  a2
    // After retw, WindowBase is restored.  a0..a3 are preserved because
    // they are outside the callee's register window.

    // --- restore context ---

    // Restore PS (restores EXCM=1).
    l32i    a2, a1, 72
    wsr     a2, PS
    rsync

    // Restore special registers.
    l32i    a2, a1, 64
    wsr     a2, SAR
    l32i    a2, a1, 68
    wsr     a2, EPC1

    // Restore general registers a15..a2.
    l32i    a15, a1, 60
    l32i    a14, a1, 56
    l32i    a13, a1, 52
    l32i    a12, a1, 48
    l32i    a11, a1, 44
    l32i    a10, a1, 40
    l32i    a9,  a1, 36
    l32i    a8,  a1, 32
    l32i    a7,  a1, 28
    l32i    a6,  a1, 24
    l32i    a5,  a1, 20
    l32i    a4,  a1, 16
    l32i    a3,  a1, 12
    l32i    a2,  a1, 8

    // Restore a0 and a1 (a1 must be last since it is the frame pointer).
    l32i    a0, a1, 0
    l32i    a1, a1, 4           // restores original SP (deallocates frame)

    rfe
