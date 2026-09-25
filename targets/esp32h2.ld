/* Linker script for the ESP32-H2.
 * The memory layout is the same as the ESP32-C6. Only the SRAM size (320kB)
 * and the ROM function addresses are different.
 * See ESP-IDF components/soc/esp32h2/include/soc/soc.h.
 *
 * Keep .data and .iram before the dummy sections. lld can calculate a wrong
 * value for a forward SIZEOF() reference, and then the image is corrupt.
 */
MEMORY
{
    /* HP-SRAM. IRAM and DRAM use the same address space. */
    SRAM (rwx) : ORIGIN = 0x40800000, LENGTH = 320K

    /* DROM and IROM are in separate, non-overlapping address ranges. */
    DROM (r)  : ORIGIN = 0x42800000, LENGTH = 8M /* Data bus (read-only, flash-mapped) */
    IROM (rx) : ORIGIN = 0x42000000, LENGTH = 8M /* Instruction bus (flash-mapped) */
}

/* The entry point. It is set in the image flashed to the chip, so must be
 * defined.
 */
ENTRY(call_start_cpu0)

SECTIONS
{
    /* === SRAM sections (loaded by ROM bootloader) === */

    /* Put the stack at the bottom of SRAM, so that the application will
     * crash on stack overflow instead of silently corrupting memory.
     * See: http://blog.japaric.io/stack-overflow-protection/
     */
    .stack (NOLOAD) :
    {
        . = ALIGN(16);
        . += _stack_size;
        _stack_top = .;
    } >SRAM

    /* Global variables that are mutable and zero-initialized.
     * These must be zeroed at startup (unlike data, which is loaded by the
     * bootloader).
     */
    .bss (NOLOAD) : ALIGN(4)
    {
        . = ALIGN (4);
        _sbss = ABSOLUTE(.);
        *(.sbss)
        *(.bss .bss.*)
        . = ALIGN (4);
        _ebss = ABSOLUTE(.);
    } >SRAM

    /* Mutable global variables. This data (in the SRAM segment) is initialized
     * by the ROM bootloader.
     */
    .data : ALIGN(4)
    {
        . = ALIGN (4);
        _sdata = ABSOLUTE(.);
        *(.sdata)
        *(.data .data.*)
        *(.dram*)
        . = ALIGN (4);
        _edata = ABSOLUTE(.);
    } >SRAM

    /* Code that must run from IRAM (not flash-mapped), for example interrupt
     * vectors and code that runs before the flash has been mapped.
     * Since IRAM and DRAM share the same address space, this is placed
     * sequentially after .data in the same SRAM region.
     */
    .iram : ALIGN(4)
    {
        *(.init)      /* call_start_cpu0 (before flash is mapped) */
        *(.iram*)     /* code that must run from IRAM */
        *(.text.handleInterruptASM) /* must be in SRAM, within JAL range of _vector_table */
        . = ALIGN(256);
        *(.text.exception_vectors)
        . = ALIGN(4);
    } > SRAM

    /* Heap starts after IRAM. */
    . = ALIGN(4);
    _heap_start = ABSOLUTE(.);
    _heap_end = ORIGIN(SRAM) + LENGTH(SRAM);

    /* === IROM sections (flash-mapped code) === */

    /* Dummy section to align .text virtual address with its flash offset.
     * In the firmware image, .data and .iram (SRAM) come before .text (IROM).
     * SIZEOF(.data) and SIZEOF(.iram) are backward references (defined above).
     * .data may be empty (esp.go skips empty sections), so use a ternary.
     */
    .irom_dummy (NOLOAD) : ALIGN(0x10000)
    {
        . += 0x18;                                              /* image header */
        . += (SIZEOF(.data) > 0) ? (SIZEOF(.data) + 0x8) : 0; /* data segment + header (if non-empty) */
        . += SIZEOF(.iram) + 0x8;                               /* iram segment + header */
        . += 0x8;                                               /* text segment header */
    } > IROM

    /* Code stored in IROM (flash-mapped).
     * This is the main program code.
     */
    .text : ALIGN(4)
    {
        *(.text .text.*)
        /* Put the .rodata data at a 16-byte aligned flash offset, the same as its address. */
        . = ALIGN(. + 0x8, 16) - 0x8;
    } > IROM

    /* === DROM sections (flash-mapped read-only data) === */

    /* Dummy section to align .rodata virtual address with its flash offset.
     * Segments in the image are sorted by address: SRAM (.data, .iram),
     * IROM (.text), then DROM (.rodata). So .rodata comes last, and we must
     * account for all preceding segments.
     * All SIZEOF() references are backward (sections defined above).
     */
    .rodata_dummy (NOLOAD): ALIGN(4)
    {
        . += 0x18;                                              /* image header (24 bytes) */
        . += (SIZEOF(.data) > 0) ? (SIZEOF(.data) + 0x8) : 0; /* data segment + header (if non-empty) */
        . += SIZEOF(.iram) + 0x8;                               /* iram segment + header */
        . += SIZEOF(.text) + 0x8;                               /* text segment + header */
        . += 0x8;                                               /* rodata segment header */
    } > DROM

    /* Constant global variables, stored in DROM. */
    .rodata : ALIGN(4)
    {
        *(.rodata*)
        . = ALIGN (4);
    } >DROM

    ASSERT(ADDR(.text) == ADDR(.irom_dummy) + SIZEOF(.irom_dummy), ".text address does not match its flash offset")
    ASSERT(ADDR(.rodata) == ADDR(.rodata_dummy) + SIZEOF(.rodata_dummy), ".rodata address does not match its flash offset")

    /DISCARD/ :
    {
        *(.eh_frame)     /* we don't do exception handling in C */
    }

    /* For the garbage collector. */
    .tinygo_stacksizes (INFO) :
    {
        *(.tinygo_stacksizes)
    }
    .tinygo_arm_entries (INFO) :
    {
        *(.tinygo_arm_entries)
    }
}

/* For the GC and scheduler. */
_globals_start = _sbss;
_globals_end = _edata;

/* Default stack size. */
_stack_size = 4K;

/* ROM function addresses for ESP32-H2.
 * See ESP-IDF components/esp_rom/esp32h2/ld/esp32h2.rom.ld.
 */
PROVIDE( ets_delay_us = 0x40000040 );

/* Cache ROM functions */
Cache_Invalidate_ICache_All = 0x40000620;
Cache_Suspend_ICache = 0x4000066c;
Cache_Resume_ICache = 0x40000670;
Cache_MMU_Init = 0x40000688;
Cache_MSPI_MMU_Set = 0x4000068c;
