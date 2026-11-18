// Minimal bootloader for the ESP32-P4. It initializes the FPU, sets up the
// flash cache/MMU, and then continues with the generic RISC-V initialization
// code which in turn calls runtime.main.
//
// Written in assembly to ensure it is entirely in IRAM and does not
// accidentally call functions stored in IROM (flash).
//
// The ESP32-P4 has:
//   - Unified I/D SRAM (L2MEM) at 0x4FF00000
//   - Unified flash mapping at 0x40000000 (both IROM and DROM)
//   - L1 I-cache + L1 D-cache + L2 cache hierarchy
//   - Hardware FPU (RV32IMAFC)
//
// Based on the ESP32-C6 startup (esp32c6.S) adapted for P4 cache API.

.section .init
.global call_start_cpu0
.type call_start_cpu0,@function
call_start_cpu0:
    // At this point:
    // - The ROM bootloader is finished and has jumped to here.
    // - We're running from IRAM: both IRAM and DRAM segments have been loaded
    //   by the ROM bootloader.
    // - We have a usable stack (but not the one we would like to use).
    // - Flash MMU IS already set up by the ROM bootloader (it maps 0x40000000).
    // - MWDT0 (TIMG0) and LP_WDT are in flashboot mode with a short timeout.

    // Disable MWDT0 (TIMG0) flashboot watchdog.
    // The ROM bootloader enables flashboot mode (bit 14 of WDTCONFIG0) and
    // starts a countdown. Must be cleared before the timeout fires.
    li t0, 0x500c2000     // TIMG0 base address
    li t1, 0x50D83AA1     // WDT write-protect key (disables write protection)
    sw t1, 0x64(t0)       // TIMG0.WDTWPROTECT = key  (unlock)
    sw zero, 0x48(t0)     // TIMG0.WDTCONFIG0  = 0    (disable + clear flashboot mode)

    // Disable LP_WDT (RWDT) flashboot watchdog.
    li t0, 0x50116000     // LP_WDT base address
    sw t1, 0x18(t0)       // LP_WDT.WPROTECT   = key  (unlock)
    sw zero, 0x0(t0)      // LP_WDT.CONFIG0    = 0    (disable + clear flashboot mode)

    // Disable LP_WDT super watchdog (SWD).
    // The SWD persists across HP system resets, so it must also be disabled.
    sw t1, 0x20(t0)       // LP_WDT.SWD_WPROTECT = key  (unlock)
    lw t2, 0x1c(t0)       // t2 = LP_WDT.SWD_CONFIG
    lui t3, 0x40000       // t3 = (1 << 30) = SWD_DISABLE bit
    or t2, t2, t3
    sw t2, 0x1c(t0)       // LP_WDT.SWD_CONFIG |= SWD_DISABLE

    // Enable FPU: set MSTATUS.FS = Initial (01)
    // Must be done before any float instructions execute.
    li t0, (1 << 13)
    csrs mstatus, t0

    // Enable UART0 clock chain so ROM UART keeps working and TinyGo can use it.
    // HP_SYS_CLKRST base = 0x500E6000.
    // All RMW done here in assembly to guarantee lui+lw/sw inline immediates;
    // TinyGo emits GP-relative BSS loads (address in BSS zero'd by ROM → crash).

    // t0 = HP_SYS_CLKRST base 0x500E6000
    lui  t0, 0x500E6

    // SOC_CLK_CTRL1 (0x500E6018): uart0_sys_clk_en = bit 18 (mask 0x40000)
    lw   t1, 0x18(t0)
    lui  t2, 0x40           // lui 0x40 → 0x40000 (bit 18)
    or   t1, t1, t2
    sw   t1, 0x18(t0)

    // SOC_CLK_CTRL2 (0x500E601C): uart0_apb_clk_en = bit 7 (mask 0x80)
    lw   t1, 0x1c(t0)
    ori  t1, t1, 0x80
    sw   t1, 0x1c(t0)

    // PERI_CLK_CTRL110 (0x500E6068): uart0_clk_src_sel=bits[25:24]=0 (XTAL),
    //                                 uart0_clk_en=bit26 (mask 0x4000000)
    lw   t1, 0x68(t0)
    lui  t2, 0x3000         // lui 0x3000 → 0x3000000 = bits[25:24] mask
    not  t2, t2
    and  t1, t1, t2         // clear src_sel bits
    lui  t2, 0x4000         // lui 0x4000 → 0x4000000 = bit 26 (clk_en)
    or   t1, t1, t2
    sw   t1, 0x68(t0)

    // UART0.CLK_CONF (0x500CA054): tx_sclk_en = bit 24
    lui  t0, 0x500CA         // t0 = 0x500CA000
    lw   t1, 0x54(t0)
    lui  t2, 0x1000         // lui 0x1000 → 0x1000000 = bit 24
    or   t1, t1, t2
    sw   t1, 0x54(t0)

    // Short NOP spin so new SCLK propagates before FIFO write.
    li   t3, 10000
1:
    addi t3, t3, -1
    bnez t3, 1b

    // Enable TIMG0 APB clock.
    // SOC_CLK_CTRL2 (0x500E601C): timergrp0_apb_clk_en = bit 21 (mask 0x200000)
    lui  t0, 0x500E6
    lw   t1, 0x1c(t0)
    lui  t2, 0x200          // lui 0x200 → 0x200000 = bit 21
    or   t1, t1, t2
    sw   t1, 0x1c(t0)

    // PERI_CLK_CTRL20 (0x500E6094): timergrp0_t0_src_sel=bits[23:22]=0 (XTAL),
    //                                timergrp0_t0_clk_en=bit24 (mask 0x1000000)
    lw   t1, 0x94(t0)
    lui  t2, 0xC00          // lui 0xC00 → 0xC00000 = bits[23:22] mask
    not  t2, t2
    and  t1, t1, t2         // clear src_sel bits
    lui  t2, 0x1000         // lui 0x1000 → 0x1000000 = bit 24 (clk_en)
    or   t1, t1, t2
    sw   t1, 0x94(t0)

    // Program the flash MMU to map all 256 64 KB pages of flash (16 MB) from
    // physical offset 0 to the unified VMA base 0x40000000.
    //
    // The ROM bootloader only maps the pages it accesses while loading SRAM
    // segments (the image lives at flash[0x2000], within page 0).  Our .text
    // section is placed at flash[0x10000] (page 1) and .rodata follows it; both
    // fall in page 1 and beyond and are therefore NOT mapped by the ROM.
    // Any instruction fetch from an unmapped flash page triggers an access fault
    // which the ROM exception handler reports as "Illegal instruction".
    //
    // Cache_FLASH_MMU_Set(sensitive, vaddr, paddr, psize_kb, num_pages, fixed)
    //   sensitive = 0   – no encryption/ECC check
    //   vaddr     = 0x40000000 – start of the unified flash VMA
    //   paddr     = 0          – flash physical byte offset 0
    //   psize     = 64         – 64 KB pages (the only supported size on P4)
    //   num       = 256        – map 256 pages = 16 MB (full flash)
    //   fixed     = 0          – sequential: VMA[n] → flash[paddr + n*64KB]
    //
    // This call is safe from IRAM: no flash fetches are in flight, so
    // reprogramming the MMU cannot deadlock the cache.
    addi sp, sp, -16
    sw   ra, 12(sp)
    li   a0, 0              // sensitive = 0
    lui  a1, 0x40000        // vaddr = 0x40000000
    li   a2, 0              // paddr = 0
    li   a3, 64             // psize = 64 KB
    li   a4, 256            // num = 256 pages
    li   a5, 0              // fixed = 0
    li   t0, 0x4fc00518     // Cache_FLASH_MMU_Set
    jalr ra, 0(t0)
    lw   ra, 12(sp)
    addi sp, sp, 16

    // Jump to generic RISC-V initialization, which initializes the stack
    // pointer and globals register. It should not return.
    // Note: Cache_Invalidate_All(0x11) was attempted here to flush stale L1
    // cache entries, but it causes a chip reset when called from IRAM with
    // the L1 I-cache active.  Power-on resets (rst:0x1) clear L1 caches
    // automatically, so the call is not needed for normal esptool flashing.
    j _start

.section .text.exception_vectors
.global _vector_table
.type _vector_table,@function

_vector_table:

    .option push
    .option norvc

    .rept 32
    j handleInterruptASM            /* interrupt handler */
    .endr

    .option pop

.size _vector_table, .-_vector_table
