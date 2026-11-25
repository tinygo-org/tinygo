// Startup code for the ESP32-S3 (Xtensa LX7, windowed ABI).
//
// The ROM bootloader loads IRAM/DRAM segments into SRAM but does NOT
// configure flash cache/MMU.  We must:
//   1. Set up the windowed-ABI register file and stack pointer.
//   2. Set VECBASE and clear PS.EXCM (needed for callx4 window overflows).
//   3. Disable watchdog timers.
//   4. Configure the flash cache and MMU so IROM/DROM are accessible.
//   5. Jump to runtime.main (in IROM).
//
// Cache/MMU init sequence (from NuttX esp_loader.c / ESP-IDF bootloader / esp-hal):
//   Phase A — configure cache modes:
//     a. rom_config_instruction_cache_mode(16KB, 8-way, 32B)
//     b. rom_Cache_Suspend_DCache()
//     c. rom_config_data_cache_mode(32KB, 8-way, 32B)
//     d. Cache_Resume_DCache(0)
//   Phase B — map flash pages:
//     e. Disable caches
//     f. Cache_MMU_Init() — reset all MMU entries to invalid
//     g. Cache_Set_IDROM_MMU_Size() — set IROM/DROM entry split
//     h. Write MMU entries mapping flash pages for IROM, DROM and machine.Flash
//     i. Clear bus-shut bits
//     j. Enable caches + isync

#define PS_WOE   0x00040000

// -----------------------------------------------------------------------
// Boot entry point — placed in IRAM by the linker.
// -----------------------------------------------------------------------
.section .text.call_start_cpu0
    .literal_position
    .align 4
.Lstack_top_addr:
    .long _stack_top
.Lmain_addr:
    .long main
.Lvector_table_addr:
    .long _vector_table
// WDT register addresses
.Lwdt_key:
    .long 0x50D83AA1
.Lrtc_wdt_protect:
    .long 0x600080B0
.Lrtc_wdt_config0:
    .long 0x60008098
.Ltimg0_wdt_protect:
    .long 0x6001F064
.Ltimg0_wdt_config0:
    .long 0x6001F048
.Ltimg1_wdt_protect:
    .long 0x60020064
.Ltimg1_wdt_config0:
    .long 0x60020048
.Lswd_protect:
    .long 0x600080B8
.Lswd_key:
    .long 0x8F1D312A
.Lswd_conf:
    .long 0x600080B4
.Lswd_disable:
    .long 0x40000000
// ROM function addresses (from ESP-IDF esp32s3.rom.ld)
.Lrom_config_icache:
    .long 0x40001a1c
.Lrom_config_dcache:
    .long 0x40001a28
.Lrom_suspend_dcache:
    .long 0x400018b4
.LCache_Resume_DCache:
    .long 0x400018c0
.LCache_Disable_ICache:
    .long 0x4000186c
.LCache_Disable_DCache:
    .long 0x40001884
.LCache_MMU_Init:
    .long 0x40001998
.LCache_Set_IDROM_MMU_Size:
    .long 0x40001914
.LCache_Enable_ICache:
    .long 0x40001878
.LCache_Enable_DCache:
    .long 0x40001890
// Cache/MMU register addresses
.Lmmu_table_base:
    .long 0x600C5000
.Licache_ctrl1_reg:
    .long 0x600C4064
.Ldcache_ctrl1_reg:
    .long 0x600C4004
// Flash MMU layout from targets/esp32s3.ld. The builder patches the physical
// IROM and DROM offsets into _irom_flash_addr and _drom_flash_addr.
.Lirom_flash_addr:
    .long _irom_flash_addr
.Lirom_mmu_pages:
    .long _irom_mmu_pages
.Ldrom_flash_addr:
    .long _drom_flash_addr
.Ldrom_mmu_index:
    .long _drom_mmu_index
.Ldrom_mmu_pages:
    .long _drom_mmu_pages
.Lflash_data_phys:
    .long __flash_data_phys
.Lflash_data_mmu_index:
    .long _flash_data_mmu_index
.Lflash_data_mmu_pages:
    .long _flash_data_mmu_pages

.global call_start_cpu0
call_start_cpu0:

    // ---- 1. Windowed-ABI register file setup ----

    // Disable WOE (bit 18 of PS).
    // Avoid large movi constants to prevent auto-generated .literal section
    // entries, which cause l32r offset miscalculation in LLVM 22 / lld.
    rsr.ps a2
    movi a3, 1
    slli a3, a3, 18         // a3 = PS_WOE (0x40000)
    and  a3, a2, a3         // isolate WOE bit
    xor  a2, a2, a3         // clear WOE bit
    wsr.ps a2
    rsync

    // Set WINDOWSTART to 1 << WINDOWBASE (mark only current window as valid).
    rsr.windowbase  a2
    ssl  a2
    movi a2, 1
    sll  a2, a2
    wsr.windowstart a2
    rsync

    // Load stack pointer.
    l32r a1, .Lstack_top_addr

    // Re-enable WOE.
    rsr.ps a2
    movi a3, 1
    slli a3, a3, 18         // a3 = PS_WOE (0x40000)
    or a2, a2, a3
    wsr.ps a2
    rsync

    // Enable FPU (coprocessor 0).
    movi a2, 1
    wsr.cpenable a2
    rsync

    // ---- 2. Disable all watchdog timers (IMMEDIATELY, before any delay) ----
    l32r a3, .Lwdt_key
    movi a4, 0

    // RTC WDT
    l32r a2, .Lrtc_wdt_protect
    memw
    s32i a3, a2, 0
    l32r a5, .Lrtc_wdt_config0
    memw
    s32i a4, a5, 0
    memw
    s32i a4, a2, 0

    // TIMG0 WDT
    l32r a2, .Ltimg0_wdt_protect
    memw
    s32i a3, a2, 0
    l32r a5, .Ltimg0_wdt_config0
    memw
    s32i a4, a5, 0
    memw
    s32i a4, a2, 0

    // TIMG1 WDT
    l32r a2, .Ltimg1_wdt_protect
    memw
    s32i a3, a2, 0
    l32r a5, .Ltimg1_wdt_config0
    memw
    s32i a4, a5, 0
    memw
    s32i a4, a2, 0

    // Super WDT
    l32r a2, .Lswd_protect
    l32r a3, .Lswd_key
    memw
    s32i a3, a2, 0
    l32r a5, .Lswd_conf
    l32r a6, .Lswd_disable
    memw
    s32i a6, a5, 0
    memw
    s32i a4, a2, 0

    // ---- 3. Set VECBASE and clear PS.EXCM ----
    // VECBASE must be set before any callx4 so that window overflow
    // exceptions (triggered by register window rotation) route to our
    // handlers in IRAM, not the ROM's default vectors.
    l32r a8, .Lvector_table_addr
    wsr.vecbase a8
    rsync

    // Clear PS.EXCM (bit 4) and PS.INTLEVEL (bits 0-3).
    // The ROM bootloader may leave EXCM=1; with EXCM set any callx4
    // window overflow would become a double exception.
    // Set PS.UM (bit 5) so level-1 exceptions route to User vector.
    rsr.ps a2
    movi a3, ~0x1F
    and a2, a2, a3
    movi a3, 0x20
    or a2, a2, a3
    wsr.ps a2
    rsync

    // ---- 4. Configure flash cache and MMU ----
    //
    // ROM function calls use callx4 (windowed ABI):
    //   a4 = target address (overwritten with return addr by call mechanism)
    //   a5 = stack pointer for callee (becomes callee's a1 via entry)
    //   a6 = first argument  (becomes callee's a2)
    //   a7 = second argument (becomes callee's a3)
    //   a8 = third argument  (becomes callee's a4)
    // Registers a0-a3 are preserved across callx4; a4-a11 may be clobbered.

    // Phase A: Configure cache modes (required for cache hardware to function).
    // Without this, the cache doesn't know its size/associativity/line-size
    // and cannot service flash accesses.

    // 4a. Configure ICache mode: 16KB, 8-way, 32-byte line
    // Use movi+slli to avoid auto-literal generation (lld l32r bug).
    movi a6, 1
    slli a6, a6, 14         // a6 = 0x4000 = 16KB
    movi a7, 8              // ways = 8
    movi a8, 32             // line_size = 32
    mov  a5, a1
    l32r a4, .Lrom_config_icache
    callx4 a4

    // 4b. Suspend DCache before configuring it
    mov  a5, a1
    l32r a4, .Lrom_suspend_dcache
    callx4 a4

    // 4c. Configure DCache mode: 32KB, 8-way, 32-byte line
    movi a6, 1
    slli a6, a6, 15         // a6 = 0x8000 = 32KB
    movi a7, 8              // ways = 8
    movi a8, 32             // line_size = 32
    mov  a5, a1
    l32r a4, .Lrom_config_dcache
    callx4 a4

    // 4d. Resume DCache
    movi a6, 0
    mov  a5, a1
    l32r a4, .LCache_Resume_DCache
    callx4 a4

    // Phase B: Map flash pages into MMU.

    // 4e. Disable ICache
    mov  a5, a1
    l32r a4, .LCache_Disable_ICache
    callx4 a4

    // 4f. Disable DCache
    mov  a5, a1
    l32r a4, .LCache_Disable_DCache
    callx4 a4

    // 4g. Initialize MMU (resets all 512 entries to invalid = 0x4000)
    mov  a5, a1
    l32r a4, .LCache_MMU_Init
    callx4 a4

    // 4h. IROM and DROM share one table of 512 entries. IROM uses the
    //     entries before DROM, see mmu_ll_get_entry_id in esp-idf mmu_ll.h.
    l32r a6, .Ldrom_mmu_index
    slli a6, a6, 2                     // a6 = IROM table size in bytes
    movi a7, 0x7FF
    addi a7, a7, 1
    sub  a7, a7, a6                    // a7 = DROM table size in bytes
    mov  a5, a1
    l32r a4, .LCache_Set_IDROM_MMU_Size
    callx4 a4

    // 4i. Map IROM, DROM and the machine.Flash area. Each loop writes a2
    //     entries from a10 with physical pages counting up from a9.
    l32r a8, .Lmmu_table_base

    mov  a10, a8
    l32r a9, .Lirom_flash_addr
    l32i a9, a9, 0
    srli a9, a9, 16
    l32r a2, .Lirom_mmu_pages
.Lirom_map_loop:
    beqz a2, .Lirom_map_done
    s32i a9, a10, 0
    addi a9, a9, 1
    addi a10, a10, 4
    addi a2, a2, -1
    j    .Lirom_map_loop
.Lirom_map_done:

    l32r a10, .Ldrom_mmu_index
    addx4 a10, a10, a8
    l32r a9, .Ldrom_flash_addr
    l32i a9, a9, 0
    srli a9, a9, 16
    l32r a2, .Ldrom_mmu_pages
.Ldrom_map_loop:
    beqz a2, .Ldrom_map_done
    s32i a9, a10, 0
    addi a9, a9, 1
    addi a10, a10, 4
    addi a2, a2, -1
    j    .Ldrom_map_loop
.Ldrom_map_done:

    l32r a10, .Lflash_data_mmu_index
    addx4 a10, a10, a8
    l32r a9, .Lflash_data_phys
    srli a9, a9, 16
    l32r a2, .Lflash_data_mmu_pages
.Lflash_data_map_loop:
    beqz a2, .Lflash_data_map_done
    s32i a9, a10, 0
    addi a9, a9, 1
    addi a10, a10, 4
    addi a2, a2, -1
    j    .Lflash_data_map_loop
.Lflash_data_map_done:
    memw

    // 4j. Clear bus-shut bits so core 0 can access ICache and DCache buses.
    //     This also clears the core 1 bits, which call_start_cpu1 relies on.
    l32r a8, .Licache_ctrl1_reg         // 0x600C4064
    movi a9, 0
    s32i a9, a8, 0                      // Clear all ICACHE_CTRL1 shut bits
    l32r a8, .Ldcache_ctrl1_reg         // 0x600C4004
    s32i a9, a8, 0                      // Clear all DCACHE_CTRL1 shut bits
    memw

    // 4k. Enable ICache (arg: autoload = 0)
    movi a6, 0
    mov  a5, a1
    l32r a4, .LCache_Enable_ICache
    callx4 a4

    // 4l. Enable DCache (arg: autoload = 0)
    movi a6, 0
    mov  a5, a1
    l32r a4, .LCache_Enable_DCache
    callx4 a4

    // Flush instruction pipeline so new cache/MMU config takes effect.
    isync

    // ---- 5. Jump to main (in IROM) ----
    // Re-clear PS.EXCM in case ROM calls changed processor state.
    rsr.ps a2
    movi a3, ~0x1F
    and a2, a2, a3
    movi a3, 0x20
    or a2, a2, a3
    wsr.ps a2
    rsync

    mov  a5, a1
    l32r a4, .Lmain_addr
    callx4 a4

    // If main returns, loop forever.
1:  j 1b

// -----------------------------------------------------------------------
// Core 1 entry point, set with ets_set_appcpu_boot_addr. Placed in IRAM by
// the linker. Only the per core setup of call_start_cpu0 is done here.
// -----------------------------------------------------------------------
.section .text.call_start_cpu1
    .literal_position
    .align 4
.Lstack1_top_addr:
    .long _stack1_top
.LrunCore1_addr:
    .long tinygo_runCore1
.Lvector_table_addr1:
    .long _vector_table

.global call_start_cpu1
call_start_cpu1:
    // Disable WOE.
    rsr.ps a2
    movi a3, 1
    slli a3, a3, 18         // a3 = PS_WOE (0x40000)
    and  a3, a2, a3
    xor  a2, a2, a3
    wsr.ps a2
    rsync

    // Set WINDOWSTART to 1 << WINDOWBASE.
    rsr.windowbase  a2
    ssl  a2
    movi a2, 1
    sll  a2, a2
    wsr.windowstart a2
    rsync

    l32r a1, .Lstack1_top_addr

    // Re-enable WOE.
    rsr.ps a2
    movi a3, 1
    slli a3, a3, 18         // a3 = PS_WOE (0x40000)
    or a2, a2, a3
    wsr.ps a2
    rsync

    // Enable FPU (coprocessor 0).
    movi a2, 1
    wsr.cpenable a2
    rsync

    // VECBASE must be set before the first callx4.
    l32r a8, .Lvector_table_addr1
    wsr.vecbase a8
    rsync

    // Clear PS.EXCM and PS.INTLEVEL, set PS.UM.
    rsr.ps a2
    movi a3, ~0x1F
    and a2, a2, a3
    movi a3, 0x20
    or a2, a2, a3
    wsr.ps a2
    rsync

    mov  a5, a1
    l32r a4, .LrunCore1_addr
    callx4 a4

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

// Park this core in IRAM while the other core erases or writes flash.
// a2 points to the state: set 1, wait for 2, then set 0.
.section .iram1.tinygo_flashPark, "ax"
.global tinygo_flashPark
    .balign 4
tinygo_flashPark:
    entry a1, 16
    movi  a3, 1
    s32i  a3, a2, 0
    memw
.Lflash_park_wait:
    memw
    l32i  a3, a2, 0
    bnei  a3, 2, .Lflash_park_wait
    movi  a3, 0
    s32i  a3, a2, 0
    memw
    retw
