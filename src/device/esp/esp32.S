
// The following definitions were copied from:
// esp-idf/components/xtensa/include/xtensa/corebits.h
#define PS_WOE_MASK                0x00040000
#define PS_OWB_MASK                0x00000F00
#define PS_CALLINC_MASK            0x00030000
#define PS_WOE                     PS_WOE_MASK

// Only calling it call_start_cpu0 for consistency with ESP-IDF.
.section .text.call_start_cpu0
1:
    .long _stack_top
.Lmain_addr:
    .long main
.Lrom_mmu_init:
    .long 0x400095a4         // mmu_init(int cpu_no)
.Lrom_cache_flash_mmu_set:
    .long 0x400095e0         // cache_flash_mmu_set(cpu, pid, vaddr, paddr, pgsz, pgcnt)
.Lrom_Cache_Read_Enable:
    .long 0x40009a84         // Cache_Read_Enable(int cpu_no)
.Lrom_Cache_Read_Disable:
    .long 0x40009ab8         // Cache_Read_Disable(int cpu_no)
.Lrom_Cache_Flush:
    .long 0x40009a14         // Cache_Flush(int cpu_no)
.Lrodata_start:
    .long _rodata_start
.Lrodata_end:
    .long _rodata_end
.Ltext_start:
    .long _text_start
.Ltext_end:
    .long _text_end
.Ldport_pro_cache_ctrl1:
    .long 0x3FF00044         // DPORT_PRO_CACHE_CTRL1_REG
.Ldrom_paddr_ptr:
    .long _drom_flash_addr   // pointer to builder-patched DROM flash offset
.Ldrom_vaddr:
    .long 0x3F400000         // DROM virtual base address
.Lirom_vaddr:
    .long 0x400D0000         // IROM virtual base address
.Lmmu_table_base:
    .long 0x3FF10000         // PRO CPU Flash MMU table
.Lrtc_wdt_protect:
    .long 0x3FF480A4         // RTC_CNTL_WDTWPROTECT_REG
.Lrtc_wdt_key:
    .long 0x50D83AA1         // WDT write-protect key
.Lrtc_wdt_config0:
    .long 0x3FF4808C         // RTC_CNTL_WDTCONFIG0_REG
.Ltimg0_wdt_protect:
    .long 0x3FF5F064         // TIMG0_WDTWPROTECT_REG
.Ltimg0_wdt_config0:
    .long 0x3FF5F048         // TIMG0_WDTCONFIG0_REG
.Lvector_table:
    .long _vector_table

.global call_start_cpu0
call_start_cpu0:
    // We need to set the stack pointer to a different value. This is somewhat
    // complicated in the Xtensa architecture. The code below is a modified
    // version of the following code:
    // https://github.com/espressif/esp-idf/blob/c77c4ccf/components/xtensa/include/xt_instr_macros.h#L47

    // Disable WOE (bit 18 of PS).
    // Avoid large movi constants to prevent auto-generated .literal section
    // entries, which cause l32r offset miscalculation in LLVM 22 / lld.
    rsr.ps a2
    movi a3, 1
    slli a3, a3, 18         // a3 = PS_WOE_MASK (0x40000)
    and  a3, a2, a3         // a3 = a2 & WOE_MASK (isolate WOE bit)
    xor  a2, a2, a3         // clear WOE bit
    wsr.ps a2
    rsync

    // Set WINDOWSTART to 1 << WINDOWBASE.
    rsr.windowbase  a2
    ssl  a2
    movi a2, 1
    sll  a2, a2
    wsr.windowstart a2
    rsync

    // Load new stack pointer.
    l32r sp, 1b

    // Re-enable WOE.
    rsr.ps a2
    movi a3, 1
    slli a3, a3, 18         // a3 = PS_WOE (0x40000)
    or a2, a2, a3
    wsr.ps a2
    rsync

    // Enable the FPU (coprocessor 0 so the lowest bit).
    movi a2, 1
    wsr.cpenable a2
    rsync

    // Disable the RTC and TIMG0 watchdogs before configuring the flash cache.
    // The ROM bootloader leaves them running; a fault during cache setup would
    // otherwise reset the chip. The Go runtime re-disables them once it starts.
    l32r a2, .Lrtc_wdt_protect
    l32r a3, .Lrtc_wdt_key
    s32i a3, a2, 0          // unlock WDT write-protect
    memw
    l32r a2, .Lrtc_wdt_config0
    movi a3, 0
    s32i a3, a2, 0          // disable WDT (write 0 to config0)
    memw

    // Disable TG0 WDT (Timer Group 0 Main Watchdog).
    // TIMG0_WDTWPROTECT_REG = 0x3FF5F064, TIMG0_WDTCONFIG0_REG = 0x3FF5F048
    l32r a2, .Ltimg0_wdt_protect
    l32r a3, .Lrtc_wdt_key  // same unlock key 0x50D83AA1
    s32i a3, a2, 0
    memw
    l32r a2, .Ltimg0_wdt_config0
    movi a3, 0
    s32i a3, a2, 0
    memw

    // Set VECBASE to our vector table. Must happen before any callx4 so that
    // register-window overflow exceptions route to our handlers.
    l32r a2, .Lvector_table
    wsr.vecbase a2
    rsync

    // Clear PS.EXCM so window overflow exceptions work properly.
    rsr.ps a2
    movi a3, ~0x1F
    and a2, a2, a3
    movi a3, 0x20          // PS.UM = 1
    or a2, a2, a3
    wsr.ps a2
    rsync

    // ---- Configure flash cache and MMU ----
    movi a6, 0
    mov  a5, a1
    l32r a4, .Lrom_Cache_Read_Disable
    callx4 a4

    movi a6, 0
    mov  a5, a1
    l32r a4, .Lrom_Cache_Flush
    callx4 a4

    movi a6, 0
    mov  a5, a1
    l32r a4, .Lrom_mmu_init
    callx4 a4

    l32r a2, .Lrodata_end
    l32r a3, .Lrodata_start
    sub  a2, a2, a3
    beqz a2, .Lskip_drom
    addi a2, a2, -1
    srli a2, a2, 16
    addi a2, a2, 1

    movi a6, 0
    movi a7, 0
    l32r a8, .Ldrom_vaddr
    l32r a9, .Ldrom_paddr_ptr
    l32i a9, a9, 0
    movi a10, 64
    mov  a11, a2
    mov  a5, a1
    l32r a4, .Lrom_cache_flash_mmu_set
    callx4 a4
.Lskip_drom:

    l32r a2, .Ltext_end
    l32r a3, .Ltext_start
    sub  a2, a2, a3
    beqz a2, .Lskip_irom
    addi a2, a2, -1
    srli a2, a2, 16
    addi a2, a2, 1

    l32r a9, .Lrodata_end
    l32r a3, .Lrodata_start
    sub  a9, a9, a3
    l32r a3, .Ldrom_paddr_ptr
    l32i a3, a3, 0
    beqz a9, .Lirom_paddr_ready
    addi a9, a9, -1
    srli a9, a9, 16
    addi a9, a9, 1
    slli a9, a9, 16
    add  a3, a3, a9
.Lirom_paddr_ready:

    movi a6, 0
    movi a7, 0
    l32r a8, .Lirom_vaddr
    mov  a9, a3
    movi a10, 64
    mov  a11, a2
    mov  a5, a1
    l32r a4, .Lrom_cache_flash_mmu_set
    callx4 a4
.Lskip_irom:

    l32r a2, .Ldport_pro_cache_ctrl1
    l32i a3, a2, 0
    movi a4, ~0x11
    and  a3, a3, a4
    s32i a3, a2, 0
    memw

    movi a6, 0
    mov  a5, a1
    l32r a4, .Lrom_Cache_Read_Enable
    callx4 a4

    isync

    // ---- Jump to main (in IROM/flash, now accessible) ----
    mov  a5, a1
    l32r a4, .Lmain_addr
    callx4 a4

    // If main returns, loop forever.
1:  j 1b

// -----------------------------------------------------------------------
// tinygo_scanCurrentStack — Spill all Xtensa register windows to the
// stack, then call tinygo_scanstack(sp) so the conservative GC can
// discover live heap pointers that are currently in physical registers.
//
// On RISC-V / ARM the equivalent function pushes callee-saved registers
// before the call.  On Xtensa windowed ABI the same effect is achieved
// by forcing hardware window-overflow for every occupied pane: each
// overflow saves the four registers in that pane to the stack frame
// pointed to by the pane's a1 (sp).  After all panes are flushed, a
// scan from the current sp to stackTop covers every live value.
//
// Without this spill the conservative GC misses heap pointers held only
// in physical registers, frees live objects, and later crashes jumping
// through a freed/garbage function pointer (e.g. a goroutine trampoline).
// -----------------------------------------------------------------------
.section .text.tinygo_scanCurrentStack

.global tinygo_scanCurrentStack
tinygo_scanCurrentStack:
    entry a1, 48

    // Disable interrupts while flushing register windows.
    rsr   a4, PS
    s32i  a4, a1, 0         // save PS for later restore
    rsil  a4, 3             // XCHAL_EXCM_LEVEL

    // Flush all register windows using recursive call4.
    // For NAREG=64 (16 panes), 15 recursive levels cover all panes
    // except the current one (which is kept active).
    movi a6, 15
    call4 .Lscan_spill

    // Restore interrupts.
    l32i  a4, a1, 0
    wsr.ps a4
    rsync

    // Pass current sp to tinygo_scanstack.
    // call4 maps caller's a5→callee's a1 (stack ptr for callee's entry)
    // and caller's a6→callee's a2 (first argument = sp).
    mov  a5, a1             // callee's a1 = valid stack pointer
    mov  a6, a1             // callee's a2 = sp argument
    call4 tinygo_scanstack

    retw

    .balign 4
.Lscan_spill:
    entry a1, 16
    beqz  a2, .Lscan_spill_done
    addi  a2, a2, -1
    mov   a6, a2
    call4 .Lscan_spill
.Lscan_spill_done:
    retw
