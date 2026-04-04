/* Linker script for the ESP32-S3 (flash XIP)
 *
 * The ESP32-S3 has 512KB of internal SRAM:
 *   - SRAM0 (32KB): 0x40370000-0x40377FFF — used by ICache when flash XIP is
 *     active, so we MUST NOT place code here.
 *   - SRAM1 (416KB): dual-mapped as IRAM 0x40378000-0x403DFFFF and
 *     DRAM 0x3FC88000-0x3FCEFFFF.
 *
 * Flash is memory-mapped via the cache:
 *   - DROM (read-only data): 0x3C000000, up to 32MB
 *   - IROM (executable code): 0x42000000, up to 32MB
 * The MMU uses 64KB pages, so the bottom 16 bits of the virtual address
 * and the flash offset must match. Dummy sections handle this alignment.
 */

MEMORY
{
    DRAM (rw) : ORIGIN = 0x3FC88000, LENGTH = 416K
    IRAM (x)  : ORIGIN = 0x40378000, LENGTH = 416K /* SRAM1 only (SRAM0 used by ICache) */

    DROM (r)  : ORIGIN = 0x3C000000, LENGTH = 32M /* Flash data bus (read-only) */
    IROM (rx) : ORIGIN = 0x42000000, LENGTH = 32M /* Flash instruction bus */
}

/* The entry point. It is set in the image flashed to the chip, so must be
 * defined.
 */
ENTRY(call_start_cpu0)

SECTIONS
{
    /* Dummy section so that .rodata starts right after the image header
     * and DROM segment header in the flash image.
     */
    .rodata_dummy (NOLOAD): ALIGN(4)
    {
        . += 0x18; /* esp_image_header_t at start of flash */
        . += 0x8;  /* DROM segment header (8 bytes) */
    } > DROM

    /* Constant global variables, stored in flash (DROM). */
    .rodata : ALIGN(4)
    {
        *(.rodata*)
        . = ALIGN (4);
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

    /* Global variables that are mutable and zero-initialized. */
    .bss (NOLOAD) : ALIGN(4)
    {
        . = ALIGN (4);
        _sbss = ABSOLUTE(.);
        *(.bss .bss.*)
        . = ALIGN (4);
        _ebss = ABSOLUTE(.);
    } >DRAM

    /* Mutable global variables, initialized by the ROM bootloader. */
    .data : ALIGN(4)
    {
        . = ALIGN (4);
        _sdata = ABSOLUTE(.);
        *(.data .data.*)
        *(.dram*)
        . = ALIGN (4);
        _edata = ABSOLUTE(.);
    } >DRAM

    /* Dummy section to skip past stack+bss+data in IRAM (dual-mapped with DRAM). */
    .iram_dummy (NOLOAD): ALIGN(4)
    {
        . += SIZEOF(.stack);
        . += SIZEOF(.bss);
        . += SIZEOF(.data);
    } > IRAM

    /* IRAM segment: boot code, interrupt vectors, and any code that must
     * run from RAM. Loaded into SRAM by the ROM bootloader.
     */
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

        /* Init code — reclaimed for heap after startup */
        __init_start = .;
        *(.init)
        __init_end = .;
        . = ALIGN(4);
        _iram_end = .;
    } >IRAM

    /* Dummy section to put the IROM segment at the correct flash offset. */
    .text_dummy (NOLOAD): ALIGN(4)
    {
        . += 0x18;                                                  /* esp_image_header_t */
        . += SIZEOF(.rodata) + ((SIZEOF(.rodata) != 0) ? 0x8 : 0); /* DROM segment (optional) */
        . += SIZEOF(.data)   + ((SIZEOF(.data)   != 0) ? 0x8 : 0); /* DRAM segment (optional) */
        . += SIZEOF(.iram)   + 0x8;                                 /* IRAM segment */
        . += 0x8;                                                   /* IROM segment header */
    } > IROM

    /* IROM segment: main code executed from flash via cache. */
    .text : ALIGN(4)
    {
        *(.literal .text)
        *(.literal.* .text.*)
        _irom_end = .;
    } >IROM

    /DISCARD/ :
    {
        *(.eh_frame)
    }
}

/* For the garbage collector.
 * _heap_start must be after the DRAM shadow of the IRAM section.
 * IRAM and DRAM share the same physical SRAM1, with addresses offset by
 * (IRAM_ORIGIN - DRAM_ORIGIN) = 0x6F0000. We use _iram_end (the actual
 * end of .iram in IRAM space) converted to DRAM to avoid counting any
 * alignment gaps between .iram_dummy and .iram.
 * Init code at the end of .iram can be reclaimed for heap.
 */
_globals_start = _sbss;
_globals_end = _edata;
_heap_start = _iram_end - 0x6F0000 - (__init_end - __init_start);
_heap_end = ORIGIN(DRAM) + LENGTH(DRAM);

_stack_size = 4K;

/* From ESP-IDF:
 * components/esp_rom/esp32/ld/esp32.rom.newlib-funcs.ld
 * This is the subset that is sometimes used by LLVM during codegen, and thus
 * must always be present.
 */
memset = 0x400011e8;
memcpy = 0x400011f4;
memmove = 0x40001200;
memcmp = 0x4000120c;

/* From ESP-IDF:
 * components/esp_rom/esp32/ld/esp32.rom.libgcc.ld
 * These are called from LLVM during codegen. The original license is Apache
 * 2.0, but I believe that a list of function names and addresses can't really
 * be copyrighted.
 */
__absvdi2 = 0x4000216c;
__absvsi2 = 0x40002178;
__adddf3 = 0x40002184;
__addsf3 = 0x40002190;
__addvdi3 = 0x4000219c;
__addvsi3 = 0x400021a8;
__ashldi3 = 0x400021b4;
__ashrdi3 = 0x400021c0;
__bswapdi2 = 0x400021cc;
__bswapsi2 = 0x400021d8;
__clear_cache = 0x400021e4;
__clrsbdi2 = 0x400021f0;
__clrsbsi2 = 0x400021fc;
__clzdi2 = 0x40002208;
__clzsi2 = 0x40002214;
__cmpdi2 = 0x40002220;
__ctzdi2 = 0x4000222c;
__ctzsi2 = 0x40002238;
__divdc3 = 0x40002244;
__divdf3 = 0x40002250;
__divdi3 = 0x4000225c;
__divsc3 = 0x40002268;
__divsf3 = 0x40002274;
__divsi3 = 0x40002280;
__eqdf2 = 0x4000228c;
__eqsf2 = 0x40002298;
__extendsfdf2 = 0x400022a4;
__ffsdi2 = 0x400022b0;
__ffssi2 = 0x400022bc;
__fixdfdi = 0x400022c8;
__fixdfsi = 0x400022d4;
__fixsfdi = 0x400022e0;
__fixsfsi = 0x400022ec;
__fixunsdfsi = 0x400022f8;
__fixunssfdi = 0x40002304;
__fixunssfsi = 0x40002310;
__floatdidf = 0x4000231c;
__floatdisf = 0x40002328;
__floatsidf = 0x40002334;
__floatsisf = 0x40002340;
__floatundidf = 0x4000234c;
__floatundisf = 0x40002358;
__floatunsidf = 0x40002364;
__floatunsisf = 0x40002370;
__gcc_bcmp = 0x4000237c;
__gedf2 = 0x40002388;
__gesf2 = 0x40002394;
__gtdf2 = 0x400023a0;
__gtsf2 = 0x400023ac;
__ledf2 = 0x400023b8;
__lesf2 = 0x400023c4;
__lshrdi3 = 0x400023d0;
__ltdf2 = 0x400023dc;
__ltsf2 = 0x400023e8;
__moddi3 = 0x400023f4;
__modsi3 = 0x40002400;
__muldc3 = 0x4000240c;
__muldf3 = 0x40002418;
__muldi3 = 0x40002424;
__mulsc3 = 0x40002430;
__mulsf3 = 0x4000243c;
__mulsi3 = 0x40002448;
__mulvdi3 = 0x40002454;
__mulvsi3 = 0x40002460;
__nedf2 = 0x4000246c;
__negdf2 = 0x40002478;
__negdi2 = 0x40002484;
__negsf2 = 0x40002490;
__negvdi2 = 0x4000249c;
__negvsi2 = 0x400024a8;
__nesf2 = 0x400024b4;
__paritysi2 = 0x400024c0;
__popcountdi2 = 0x400024cc;
__popcountsi2 = 0x400024d8;
__powidf2 = 0x400024e4;
__powisf2 = 0x400024f0;
__subdf3 = 0x400024fc;
__subsf3 = 0x40002508;
__subvdi3 = 0x40002514;
__subvsi3 = 0x40002520;
__truncdfsf2 = 0x4000252c;
__ucmpdi2 = 0x40002538;
__udivdi3 = 0x40002544;
__udivmoddi4 = 0x40002550;
__udivsi3 = 0x4000255c;
__udiv_w_sdiv = 0x40002568;
__umoddi3 = 0x40002574;
__umodsi3 = 0x40002580;
__unorddf2 = 0x4000258c;
__unordsf2 = 0x40002598;
