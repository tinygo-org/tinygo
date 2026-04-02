
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
// Entire handler is inline — no jump, no stack access, no memory loads.
// Just disable all CPU interrupts via INTENABLE and return.
// Buffered() re-enables INTENABLE after draining the hardware FIFO.
// -----------------------------------------------------------------------
    .org _vector_table + 0x340
    .global _level1_vector
_level1_vector:
    wsr     a0, EXCSAVE1        // save a0
    movi    a0, 0
    wsr     a0, INTENABLE       // disable ALL CPU interrupts
    rsr     a0, EXCSAVE1        // restore a0
    rfe                         // return from exception

// -----------------------------------------------------------------------
// Offset 0x3C0 — Double exception (stub — loops forever)
// -----------------------------------------------------------------------
    .org _vector_table + 0x3C0
_double_vector:
    j _double_vector
