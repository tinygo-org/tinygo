/* Linker script for the ESP32-H2.
 * The memory layout is the same as the ESP32-C6. Only the SRAM size (320kB)
 * and the ROM function addresses are different.
 * See ESP-IDF components/soc/esp32h2/include/soc/soc.h.
 *
 * Code and read-only data share one 16MB flash window at 0x42000000.
 * See ESP32-H2 Technical Reference Manual, chapter System and Memory,
 * section External Memory.
 *
 * The ROM bootloader only loads .data and .iram. .text and .rodata follow in
 * flash and esp32c6.S maps them 1:1. See targets/esp32c6.ld.
 *
 * Keep .data and .iram before the dummy sections. lld can calculate a wrong
 * value for a forward SIZEOF() reference, and then the image is corrupt.
 */
MEMORY
{
    /* HP-SRAM. IRAM and DRAM use the same address space. */
    SRAM (rwx) : ORIGIN = 0x40800000, LENGTH = 320K

    /* Flash window, shared by code and read-only data. */
    FLASH (rx) : ORIGIN = 0x42000000, LENGTH = 16M
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
        *(COMMON)
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
        *(.text.iram.*)
        *(.text.handleInterruptASM) /* must be in SRAM, within JAL range of _vector_table */
        . = ALIGN(256);
        *(.text.exception_vectors)
        . = ALIGN(4);
    } > SRAM

    /* Heap starts after IRAM. Goroutine stacks come from the heap and need
     * 16 byte alignment (RISC-V ELF psABI, integer calling convention). */
    . = ALIGN(16);
    _heap_start = ABSOLUTE(.);
    _heap_end = ORIGIN(SRAM) + LENGTH(SRAM);

    /* === FLASH sections (flash-mapped code and read-only data) === */

    /* Size of the RAM part of the image. Header, segments, padding to 16 bytes
     * and the SHA-256 digest, as written by builder/esp.go. */
    _ram_image_payload_size =
        0x18 +
        SIZEOF(.data) + ((SIZEOF(.data) != 0) ? 0x8 : 0) +
        SIZEOF(.iram) + 0x8;

    _ram_image_size =
        _ram_image_payload_size +
        16 - (_ram_image_payload_size % 16) +
        0x20;

    /* The ROM only loads the RAM part of the image. .text and .rodata follow
     * it in flash, so each offset in the flash window is its flash offset. */
    .flash_dummy (NOLOAD) : ALIGN(4)
    {
        . += _ram_image_size;
    } > FLASH

    /* Code stored in flash. This is the main program code. */
    .text : ALIGN(4)
    {
        *(.text .text.*)
    } > FLASH

    /* Constant global variables, stored in flash. */
    .rodata : ALIGN(4)
    {
        *(.rodata*)
        . = ALIGN (4);
    } > FLASH

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
