// uefiTicks returns the x86 Time Stamp Counter value using RDTSC instruction
// Go: func uefiTicks() uint64
// Returns: RAX = full 64-bit TSC value
//
.section .text.uefiTicks,"ax"
.global uefiTicks
uefiTicks:
    rdtsc
    // RDTSC returns TSC in EDX:EAX (high 32 bits in EDX, low 32 bits in EAX)
    // Combine into 64-bit value: RAX = (RDX << 32) | EAX
    shlq $32, %rdx
    orq %rdx, %rax
    retq

// uefiCpuPause executes the PAUSE instruction for spin-wait loops
// Go: func uefiCpuPause()
//
// The PAUSE instruction is a hint to the CPU that we're in a spin-wait loop.
// It improves performance on hyperthreaded CPUs by:
// - Reducing power consumption
// - Avoiding memory order violations
// - Allowing the other hyperthread to use resources
//
.section .text.uefiCpuPause,"ax"
.global uefiCpuPause
uefiCpuPause:
    pause
    retq

// uefiReadRandom generates a 64-bit random number using the RDRAND instruction
// Go: func uefiReadRandom() (value uint64, ok bool)
//
// Returns:
//   RAX = random value (undefined if CF=0)
//   RDX = 1 if success (CF=1), 0 if failed (CF=0)
//
// The RDRAND instruction uses the CPU's hardware random number generator.
// It may fail if the entropy pool is exhausted, so callers should retry.
//
.section .text.uefiReadRandom,"ax"
.global uefiReadRandom
uefiReadRandom:
    xorq %rax, %rax         // Clear RAX (return 0 on failure)
    xorq %rdx, %rdx         // Clear RDX (ok = false)
    rdrand %rax             // Try to get random number into RAX
    jnc 1f                  // Jump if carry flag not set (failure)
    movq $1, %rdx           // ok = true
1:
    retq

// uefiHasRNG checks if RDRAND instruction is supported
// Go: func uefiHasRNG() bool
//
// Returns:
//   RAX = 1 if RDRAND is supported, 0 otherwise
//
// Uses CPUID leaf 1 to check ECX bit 30 (RDRAND feature flag)
//
.section .text.uefiHasRNG,"ax"
.global uefiHasRNG
uefiHasRNG:
    pushq %rbx              // Save RBX (CPUID clobbers it)

    movl $1, %eax           // CPUID leaf 1
    cpuid

    xorq %rax, %rax         // Clear result
    testl $(1 << 30), %ecx  // Check RDRAND bit (bit 30 of ECX)
    jz 1f                   // Jump if not set
    movq $1, %rax           // RDRAND supported
1:
    popq %rbx               // Restore RBX
    retq


