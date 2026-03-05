
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
// -----------------------------------------------------------------------
    .org _vector_table + 0x340
    .global _level1_vector
_level1_vector:
    wsr     a0, EXCSAVE1        // save a0 in EXCSAVE1
    call0   _handle_level1      // jump to handler (must be in reach)
    rfe                         // return from exception

// -----------------------------------------------------------------------
// Offset 0x3C0 — Double exception (stub — loops forever)
// -----------------------------------------------------------------------
    .org _vector_table + 0x3C0
_double_vector:
    j _double_vector

// -----------------------------------------------------------------------
// Level-1 interrupt / exception handler (full, outside vector).
//
// The vector stub at offset 0x180 saved a0 to EXCSAVE1 and did a
// call0 here.  On entry:
//   a0 = return address into the vector stub (for ret.n → rfe)
//   EXCSAVE1 = original a0
//   PS.EXCM = 1  (interrupts masked at level ≤ EXCMLEVEL)
//
// We save all live a0–a15 on the interrupted task's stack and call
// the Go handleInterrupt() function with call4 (windowed).
// -----------------------------------------------------------------------
    .section .text._handle_level1,"ax",@progbits
    .global _handle_level1
    .type   _handle_level1, %function
    .balign 4
_handle_level1:
    // We were called with call0, so a0 is the return address into the
    // vector stub.  Stash it so we can return later.
    mov     a2, a0              // a2 = return-to-vector address

    rsr     a0, EXCSAVE1        // restore original a0

    // Allocate 80 bytes on the interrupted stack:
    //   16 words * 4 = 64 bytes for a0–a15
    //   + 16 bytes standard ABI base-area
    addi    sp, sp, -80

    // Save all general-purpose registers.
    s32i    a0,  sp, 16         // original a0 (was in EXCSAVE1)
    s32i    a1,  sp, 20         // (sp before our addi, will fixup)
    s32i    a2,  sp, 24         // holds return-address, save it
    s32i    a3,  sp, 28
    s32i    a4,  sp, 32
    s32i    a5,  sp, 36
    s32i    a6,  sp, 40
    s32i    a7,  sp, 44
    s32i    a8,  sp, 48
    s32i    a9,  sp, 52
    s32i    a10, sp, 56
    s32i    a11, sp, 60
    s32i    a12, sp, 64
    s32i    a13, sp, 68
    s32i    a14, sp, 72
    s32i    a15, sp, 76

    // Fix the saved sp to be the value before we adjusted it.
    addi    a3, sp, 80
    s32i    a3, sp, 20

    // Save special registers we'll need to restore.
    rsr     a3, SAR
    s32i    a3, sp, 0           // SAR at offset 0
    rsr     a3, EPC1
    s32i    a3, sp, 4           // EPC1 at offset 4
    rsr     a3, EXCCAUSE
    s32i    a3, sp, 8           // EXCCAUSE at offset 8
    rsr     a3, EXCSAVE1
    s32i    a3, sp, 12          // EXCSAVE1 at offset 12

    // Clear PS.EXCM so the windowed ABI works (entry/retw), and set
    // PS.INTLEVEL = 1 to keep level-1 interrupts masked until we're done.
    // Keep PS.WOE = 1.
    rsr     a3, PS
    movi    a4, ~(PS_WOE | PS_EXCM | PS_INTLEVEL_MASK)
    and     a3, a3, a4          // clear WOE, EXCM, INTLEVEL
    movi    a4, PS_WOE | 1      // WOE=1, INTLEVEL=1, EXCM=0
    or      a3, a3, a4
    wsr     a3, PS
    rsync

    // Call the Go-level interrupt handler using the windowed call ABI.
    // handleInterrupt is exported from the interrupt_esp32s3.go file.
    call4   handleInterrupt

    // Re-enter exception mode: set PS.EXCM = 1, PS.INTLEVEL = 0 (EXCM
    // will mask level ≤ EXCMLEVEL which is level 1 on ESP32-S3).
    // This prevents window exceptions while we restore registers.
    rsr     a3, PS
    movi    a4, ~PS_INTLEVEL_MASK
    and     a3, a3, a4          // clear INTLEVEL
    movi    a4, 0x00000010      // PS.EXCM bit
    or      a3, a3, a4
    wsr     a3, PS
    rsync

    // Disable windowed ABI (clear PS.WOE) so register restores work.
    rsr     a3, PS
    movi    a4, ~PS_WOE
    and     a3, a3, a4
    wsr     a3, PS
    rsync

    // Restore special registers.
    l32i    a3, sp, 0
    wsr     a3, SAR
    l32i    a3, sp, 4
    wsr     a3, EPC1
    l32i    a3, sp, 8
    wsr     a3, EXCCAUSE
    l32i    a3, sp, 12
    wsr     a3, EXCSAVE1

    // Restore general-purpose registers.
    l32i    a15, sp, 76
    l32i    a14, sp, 72
    l32i    a13, sp, 68
    l32i    a12, sp, 64
    l32i    a11, sp, 60
    l32i    a10, sp, 56
    l32i    a9,  sp, 52
    l32i    a8,  sp, 48
    l32i    a7,  sp, 44
    l32i    a6,  sp, 40
    l32i    a5,  sp, 36
    l32i    a4,  sp, 32
    l32i    a3,  sp, 28
    l32i    a2,  sp, 24         // this is the return address
    l32i    a0,  sp, 16         // original a0

    // Restore sp.
    l32i    sp,  sp, 20

    // Re-enable PS.WOE (the rfe will restore PS from EPS1 anyway,
    // but we need WOE working for the rfe itself if underflow occurs).
    // Actually, rfe restores PS from the saved EPS1, so we just need
    // a0 = return address into the vector stub for the rfe instruction.
    // The vector stub does: rfe
    // But we did call0 from the vector, so a0 still holds the return
    // address (we saved it as a2, now restored to a2).
    mov     a0, a2              // a0 = return to vector stub (for rfe)
    ret.n                       // return to vector stub → rfe
