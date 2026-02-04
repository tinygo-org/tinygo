// uefiTicks returns the ARM64 generic timer counter value (CNTVCT_EL0).
// This is the ARM64 equivalent of x86 RDTSC.
// Go: func uefiTicks() uint64
// Returns: x0 = 64-bit counter value
//
.section .text.uefiTicks,"ax"
.global uefiTicks
uefiTicks:
    mrs     x0, CNTVCT_EL0
    ret

// uefiCpuPause executes the YIELD instruction for spin-wait loops.
// This is the ARM64 equivalent of x86 PAUSE.
// Go: func uefiCpuPause()
//
.section .text.uefiCpuPause,"ax"
.global uefiCpuPause
uefiCpuPause:
    yield
    ret

// uefiReadRandom generates a 64-bit random number using the RNDR register.
// Requires ARMv8.5-A RNG extension. Check HasRNGSupport() first.
// Go: func uefiReadRandom() (value uint64, ok bool)
//
// Returns:
//   x0 = random value (0 if failed)
//   x1 = 1 if success, 0 if failed
//
// The RNDR instruction sets NZCV flags: if ZF=1, the read failed
// (entropy exhausted).
//
.section .text.uefiReadRandom,"ax"
.global uefiReadRandom
uefiReadRandom:
    mov     x0, #0              // Default: value = 0
    mov     x1, #0              // Default: ok = false
    mrs     x0, S3_3_C2_C4_0   // RNDR register (encoded as S3_3_C2_C4_0)
    cset    x1, ne              // ok = (ZF == 0), i.e. success
    // If ZF was set (failure), x0 is undefined; clear it.
    csel    x0, x0, xzr, ne    // x0 = (success) ? x0 : 0
    ret

// uefiHasRNG checks if the RNDR instruction is supported.
// Go: func uefiHasRNG() bool
//
// Returns:
//   x0 = 1 if RNDR is supported, 0 otherwise
//
// Reads ID_AA64ISAR0_EL1 and checks bits [63:60] for the RNDR field.
// Value >= 1 means RNDR is supported.
//
.section .text.uefiHasRNG,"ax"
.global uefiHasRNG
uefiHasRNG:
    mrs     x0, ID_AA64ISAR0_EL1
    ubfx    x0, x0, #60, #4    // Extract bits [63:60]
    cmp     x0, #0
    cset    x0, ne              // x0 = (field != 0) ? 1 : 0
    ret
