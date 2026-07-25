/* Linker script for the ESP32 (flash XIP)
 *
 * The ESP32 has 520KB of internal SRAM:
 *   - SRAM0 (192KB): 0x40070000-0x4009FFFF — upper 64KB used by ICache when
 *     flash XIP is active. Usable IRAM: 0x40080000 (128KB).
 *   - SRAM1 (128KB): 0x3FFE0000-0x3FFFFFFF — data only.
 *   - SRAM2 (200KB): 0x3FFAE000-0x3FFDFFFF — data only.
 *
 * Flash is memory-mapped via the cache:
 *   - DROM (read-only data): 0x3F400000, up to 4MB (64KB pages via MMU)
 *   - IROM (executable code): 0x400D0000, up to 4MB (64KB pages via MMU)
 * The startup code in IRAM configures the MMU to map flash pages.
 * The TinyGo builder places DROM/IROM content at page-aligned flash offsets.
 */

MEMORY
{
    /* SRAM2 usable DRAM (0x3FFB0000-0x3FFE0000, 192K): reserve the first 8KB
     * (0x3FFAE000-0x3FFB0000) for ROM PHY data (rom_phyFuns at 0x3FFAE0C0),
     * then use the rest for the stack, initialized .data, .bss, and the Go
     * GC heap. */
    DRAM  (rw) : ORIGIN = 0x3FFB0000, LENGTH = 192K

    /* Top of SRAM1 (0x3FFF0000-0x40000000, 64K): pool 7 + pool 6. This is the
     * ONLY part of SRAM1 that is safe for us to use:
     *   - 0x3FFE0000-0x3FFE0440 : ROM PRO data (ROM writes it) — avoid
     *   - 0x3FFE0440-0x3FFE8000 : ROM boot-stack area — NOT free during WiFi;
     *                             putting the arena here corrupts DMA and breaks
     *                             association ("auth expired") — avoid
     *   - 0x3FFE8000-0x3FFF0000 : pool 8, "used for MAC dump" by the WiFi MAC —
     *                             corrupts anything placed here — avoid
     *   - 0x3FFF0000-0x40000000 : pool 7/6 — safe (verified: arena + .bss both
     *                             work here). Used for the WiFi arena.
     * The arena (DMA buffers) owns this whole 64K block; .bss stays in SRAM2. */
    DRAM1 (rw) : ORIGIN = 0x3FFF0000, LENGTH = 64K

    IRAM  (x)  : ORIGIN = 0x40080000, LENGTH = 128K /* SRAM0 (usable portion) */

    DROM (r)  : ORIGIN = 0x3F400000, LENGTH = 4M /* Flash data bus (page-aligned) */
    IROM (rx) : ORIGIN = 0x400D0000, LENGTH = 4M /* Flash instruction bus (page-aligned) */
}

ENTRY(call_start_cpu0)

SECTIONS
{
    /* Constant global variables, stored in flash (DROM).
     * The builder places this at a page-aligned flash offset.
     */
    .rodata : ALIGN(4)
    {
        _rodata_start = ABSOLUTE(.);
        *(.rodata*)
        . = ALIGN (4);
        _rodata_end = ABSOLUTE(.);
    } >DROM

    /* Put the stack at the bottom of DRAM, so that the application will
     * crash on stack overflow instead of silently corrupting memory.
     */
    .stack (NOLOAD) :
    {
        . = ALIGN(16);
        . += _stack_size;
        _stack_top = .;
    } >DRAM

    /* Global variables that are mutable and zero-initialized. Kept in SRAM2
     * (DRAM): the WiFi blob's .bss holds structures the MAC/DMA touches, and
     * SRAM1 has no room for both .bss and an adequately-sized arena in its only
     * safe region (pool 7/6, 64K) — so .bss stays in SRAM2 and the arena owns
     * the SRAM1 block. */
    .bss (NOLOAD) : ALIGN(4)
    {
        . = ALIGN (4);
        _sbss = ABSOLUTE(.);
        *(.bss .bss.*)
        . = ALIGN (4);
        _ebss = ABSOLUTE(.);
    } >DRAM

    /* Reserved zero-initialized storage for WiFi blob global variables.
     * On ESP32-S3 these live at fixed ROM DRAM addresses; on ESP32 the blob
     * doesn't export them so we allocate space here. */
    .wifi_bss (NOLOAD) : ALIGN(4)
    {
        _wifi_bss_start = ABSOLUTE(.);
        /* Function pointers patched by netif.c (6 × 4 bytes) */
        PROVIDE(g_config_func = ABSOLUTE(.));             . += 4;
        PROVIDE(g_net80211_tx_func = ABSOLUTE(.));        . += 4;
        PROVIDE(g_timer_func = ABSOLUTE(.));              . += 4;
        PROVIDE(s_michael_mic_failure_cb = ABSOLUTE(.));  . += 4;
        PROVIDE(g_tx_done_cb_func = ABSOLUTE(.));         . += 4;
        PROVIDE(s_encap_amsdu_func = ABSOLUTE(.));        . += 4;
        /* Pointer-sized globals used by pp/lmac (3 × 4 bytes) */
        PROVIDE(pp_wdev_funcs = ABSOLUTE(.));             . += 4;
        PROVIDE(our_tx_eb = ABSOLUTE(.));                 . += 4;
        PROVIDE(our_wait_eb = ABSOLUTE(.));               . += 4;
        PROVIDE(lmacConfMib_ptr = ABSOLUTE(.));           . += 4;
        . = ALIGN(4);
        _wifi_bss_end = ABSOLUTE(.);
    } >DRAM

    /* RX ring buffer: placed in DRAM1 (ahead of the arena) to free ~9.6 KB
     * of SRAM2 for the Go GC heap. espradio's netif.c tags the RX ring
     * with __attribute__((section(".rxring"))) on ESP32. */
    .rxring (NOLOAD) : ALIGN(16)
    {
        *(.rxring .rxring.*)
    } >DRAM1

    /* WiFi-only BSS: ISR tables, timer slots, and ISR ring buffers moved
     * from SRAM2 to DRAM1 to free ~2.8 KB for the Go GC heap. */
    .wifibss (NOLOAD) : ALIGN(4)
    {
        *(.wifibss .wifibss.*)
    } >DRAM1

    /* WiFi blob arena (DMA buffers): owns the remainder of DRAM1 (SRAM1
     * pool 7/6) after the RX ring. Outside the Go GC heap; DMA-capable;
     * isolated from GC objects. */
    .arena (NOLOAD) : ALIGN(16)
    {
        _arena_start = ABSOLUTE(.);
    } >DRAM1
    _arena_end = ORIGIN(DRAM1) + LENGTH(DRAM1);
    ASSERT((_arena_end - _arena_start) >= 32K, "arena region < 32K")

    /* Mutable global variables, initialized by the ROM bootloader. */
    .data : ALIGN(4)
    {
        . = ALIGN (4);
        _sdata = ABSOLUTE(.);
        *(.data .data.*)
        *(.dram*)
        . = ALIGN (4);
        /* DROM flash physical address, patched by the TinyGo ESP image builder
         * with the flash offset where the .rodata (DROM) segment was placed.
         * The startup code reads this to program the flash cache MMU. */
        _drom_flash_addr = ABSOLUTE(.);
        LONG(0x10000);
        . = ALIGN (4);
        _edata = ABSOLUTE(.);
    } >DRAM

    /* Check that the boot ROM stack (for the APP CPU) does not overlap with the
     * data that is loaded by the boot ROM.
     */
    ASSERT(_edata < 0x3ffe1320, "the .data section overlaps with the stack used by the boot ROM, possibly causing corruption at startup")

    /* IRAM segment: boot code, interrupt vectors, and WiFi/BLE blob code
     * that must run from RAM. */
    .iram : ALIGN(4)
    {
        /* Boot entry point and its literals */
        *(.literal.call_start_cpu0)
        *(.text.call_start_cpu0)

        /* Xtensa exception/interrupt vector table — must be 0x400-aligned */
        . = ALIGN(0x400);
        *(.text.exception_vectors)

        /* Level-1 interrupt handler */
        *(.literal._handle_level1)
        *(.text._handle_level1)

        /* WiFi/BLE blob IRAM sections */
        *(.iram*)
        *(.wifislprxiram*)
        *(.wifiextrairam*)
        *(.wifi0iram*)
        *(.wifislpiram*)
        *(.wifirxiram*)
        *(.wifiorslpiram*)
        *(.iram1*)
        *(.coexiram*)

        . = ALIGN(4);
        _iram_end = .;
    } >IRAM

    /* IROM segment: main code executed from flash via MMU cache.
     * The builder places this at a page-aligned flash offset.
     */
    .text : ALIGN(4)
    {
        _text_start = ABSOLUTE(.);
        *(.literal .text)
        *(.literal.* .text.*)
        *(.phyiram*)
        . = ALIGN(4);
        _text_end = ABSOLUTE(.);
    } >IROM

    /DISCARD/ :
    {
        *(.eh_frame)
    }
}

/* For the garbage collector. */
_globals_start = _sbss;
_globals_end = _edata;
_heap_start = _edata;
/* Go GC heap = the rest of SRAM2 above .data (~113KB), a single safe contiguous
 * block. The WiFi arena lives in DRAM1 (SRAM1 pool 7/6), so its ~64KB does not
 * consume Go heap and its DMA is isolated from GC objects. The heap never
 * touches SRAM1's ROM/MAC-dump regions (0x3FFE0000-0x3FFF0000, left unused). */
_heap_end = 0x3FFE0000;

_stack_size = 4K;

/* ESP32 ROM-owned PHY DRAM symbols from ESP-IDF esp32.rom.ld.
 * Exposed as linker symbols so firmware code can reference them without
 * hardcoding addresses in C sources. */
PROVIDE ( rom_phyFuns = 0x3ffae0c0 );
PROVIDE ( g_phyFuns_instance = 0x3ffae0c4 );

/* From ESP-IDF:
 * components/esp_rom/esp32/ld/esp32.rom.newlib-funcs.ld
 * This is the subset that is sometimes used by LLVM during codegen, and thus
 * must always be present.
 */
memcpy  = 0x4000c2c8;
memmove = 0x4000c3c0;
memset  = 0x4000c44c;

/* From ESP-IDF:
 * components/esp_rom/esp32/ld/esp32.rom.libgcc.ld
 * These are called from LLVM during codegen. The original license is Apache
 * 2.0, but I believe that a list of function names and addresses can't really
 * be copyrighted.
 */
__absvdi2      = 0x4006387c;
__absvsi2      = 0x40063868;
__adddf3       = 0x40002590;
__addsf3       = 0x400020e8;
__addvdi3      = 0x40002cbc;
__addvsi3      = 0x40002c98;
__ashldi3      = 0x4000c818;
__ashrdi3      = 0x4000c830;
__bswapdi2     = 0x40064b08;
__bswapsi2     = 0x40064ae0;
__clrsbdi2     = 0x40064b7c;
__clrsbsi2     = 0x40064b64;
__clzdi2       = 0x4000ca50;
__clzsi2       = 0x4000c7e8;
__cmpdi2       = 0x40063820;
__ctzdi2       = 0x4000ca64;
__ctzsi2       = 0x4000c7f0;
__divdc3       = 0x400645a4;
__divdf3       = 0x40002954;
__divdi3       = 0x4000ca84;
__divsi3       = 0x4000c7b8;
__eqdf2        = 0x400636a8;
__eqsf2        = 0x40063374;
__extendsfdf2  = 0x40002c34;
__ffsdi2       = 0x4000ca2c;
__ffssi2       = 0x4000c804;
__fixdfdi      = 0x40002ac4;
__fixdfsi      = 0x40002a78;
__fixsfdi      = 0x4000244c;
__fixsfsi      = 0x4000240c;
__fixunsdfsi   = 0x40002b30;
__fixunssfdi   = 0x40002504;
__fixunssfsi   = 0x400024ac;
__floatdidf    = 0x4000c988;
__floatdisf    = 0x4000c8c0;
__floatsidf    = 0x4000c944;
__floatsisf    = 0x4000c870;
__floatundidf  = 0x4000c978;
__floatundisf  = 0x4000c8b0;
__floatunsidf  = 0x4000c938;
__floatunsisf  = 0x4000c864;
__gcc_bcmp     = 0x40064a70;
__gedf2        = 0x40063768;
__gesf2        = 0x4006340c;
__gtdf2        = 0x400636dc;
__gtsf2        = 0x400633a0;
__ledf2        = 0x40063704;
__lesf2        = 0x400633c0;
__lshrdi3      = 0x4000c84c;
__ltdf2        = 0x40063790;
__ltsf2        = 0x4006342c;
__moddi3       = 0x4000cd4c;
__modsi3       = 0x4000c7c0;
__muldc3       = 0x40063c90;
__muldf3       = 0x4006358c;
__muldi3       = 0x4000c9fc;
__mulsf3       = 0x400632c8;
__mulsi3       = 0x4000c7b0;
__mulvdi3      = 0x40002d78;
__mulvsi3      = 0x40002d60;
__nedf2        = 0x400636a8;
__negdf2       = 0x400634a0;
__negdi2       = 0x4000ca14;
__negsf2       = 0x400020c0;
__negvdi2      = 0x40002e98;
__negvsi2      = 0x40002e78;
__nesf2        = 0x40063374;
__nsau_data    = 0x3ff96544;
__paritysi2    = 0x40002f3c;
__popcount_tab = 0x3ff96544;
__popcountdi2  = 0x40002ef8;
__popcountsi2  = 0x40002ed0;
__powidf2      = 0x400638e4;
__subdf3       = 0x400026e4;
__subsf3       = 0x400021d0;
__subvdi3      = 0x40002d20;
__subvsi3      = 0x40002cf8;
__truncdfsf2   = 0x40002b90;
__ucmpdi2      = 0x40063840;
__udiv_w_sdiv  = 0x40064bec;
__udivdi3      = 0x4000cff8;
__udivmoddi4   = 0x40064bf4;
__udivsi3      = 0x4000c7c8;
__umoddi3      = 0x4000d280;
__umodsi3      = 0x4000c7d0;
__umulsidi3    = 0x4000c7d8;
__unorddf2     = 0x400637f4;
__unordsf2     = 0x40063478;

/* From ESP-IDF:
 * components/esp_rom/esp32/ld/esp32.rom.ld
 * ROM functions needed for WiFi/PHY support.
 */
PROVIDE(ets_delay_us = 0x40008534);
PROVIDE(ets_printf = 0x40007d54);
PROVIDE(ets_install_putc1 = 0x40007d18);
PROVIDE(ets_install_uart_printf = 0x40007d28);
PROVIDE(intr_matrix_set = 0x4000681c);
PROVIDE(rtc_get_reset_reason = 0x400081d4);
PROVIDE(phy_get_romfuncs = 0x40004100);
PROVIDE(uart_tx_one_char = 0x40009200);
PROVIDE(uart_rx_one_char = 0x400092d0);
PROVIDE(ets_intr_lock = 0x400067b0);
PROVIDE(ets_intr_unlock = 0x400067c4);
PROVIDE(ets_isr_attach = 0x400067ec);
PROVIDE(ets_isr_mask = 0x400067fc);
PROVIDE(ets_isr_unmask = 0x40006808);
PROVIDE(roundup2 = 0x4000ab7c);
