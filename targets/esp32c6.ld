/* Linker script for the ESP32-C6
 *
 * The ESP32-C6 has a simpler memory layout than the ESP32-C3:
 *   - It has 512kB of HP-SRAM. Unlike the C3, IRAM and DRAM share the same
 *     address space at 0x40800000, so there is no need for separate DRAM/IRAM
 *     regions.
 *   - It has 16kB of LP-SRAM at 0x50000000 for low-power / RTC use.
 *   - DROM and IROM are in separate, non-overlapping 8MB address ranges:
 *     IROM at 0x42000000 and DROM at 0x42800000.
 *   - The MMU works in pages of 64kB, which means the bottom 16 bits of the
 *     address in flash and the address in DROM/IROM need to match.
 *   - Memory in SRAM is loaded at reset by the ROM bootloader.
 *
 * The firmware image segments are sorted by virtual address (by esp.go):
 *   1. .data  (SRAM 0x408...)   - loaded by ROM bootloader
 *   2. .iram  (SRAM 0x408...)   - loaded by ROM bootloader
 *   3. .text  (IROM 0x420...)   - flash-mapped via MMU
 *   4. .rodata (DROM 0x428...)  - flash-mapped via MMU
 *
 * IMPORTANT: SRAM sections (.data, .iram) are defined BEFORE the dummy
 * sections so that SIZEOF() references are backward (not forward). lld may
 * not correctly resolve forward SIZEOF() references, which would produce
 * wrong dummy sizes and corrupt the firmware image.
 */

MEMORY
{
    /* HP-SRAM: unified IRAM/DRAM address space (512K).
     * Unlike the C3, IRAM and DRAM share the same address space, so we use
     * a single memory region to avoid linker overlap errors.
     */
    SRAM (rwx) : ORIGIN = 0x40800000, LENGTH = 512K

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
     * Since IRAM and DRAM share the same address space on ESP32-C6, this is
     * placed sequentially after .data in the same SRAM region.
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

/* ROM function addresses for ESP32-C6.
 * Source: ESP-IDF components/esp_rom/esp32c6/ld/esp32c6.rom.ld
 */
PROVIDE( ets_delay_us = 0x40000040 );

/* Cache ROM functions */
Cache_Get_ICache_Line_Size = 0x40000628;
Cache_Get_Mode = 0x4000062c;
Cache_Address_Through_Cache = 0x40000630;
ROM_Boot_Cache_Init = 0x40000634;
MMU_Set_Page_Mode = 0x40000638;
MMU_Get_Page_Mode = 0x4000063c;
Cache_Invalidate_ICache_Items = 0x40000640;
Cache_Op_Addr = 0x40000644;
Cache_Invalidate_Addr = 0x40000648;
Cache_Invalidate_ICache_All = 0x4000064c;
Cache_Mask_All = 0x40000650;
Cache_UnMask_Dram0 = 0x40000654;
Cache_Suspend_ICache_Autoload = 0x40000658;
Cache_Resume_ICache_Autoload = 0x4000065c;
Cache_Start_ICache_Preload = 0x40000660;
Cache_ICache_Preload_Done = 0x40000664;
Cache_End_ICache_Preload = 0x40000668;
Cache_Config_ICache_Autoload = 0x4000066c;
Cache_Enable_ICache_Autoload = 0x40000670;
Cache_Disable_ICache_Autoload = 0x40000674;
Cache_Enable_ICache_PreLock = 0x40000678;
Cache_Disable_ICache_PreLock = 0x4000067c;
Cache_Lock_ICache_Items = 0x40000680;
Cache_Unlock_ICache_Items = 0x40000684;
Cache_Lock_Addr = 0x40000688;
Cache_Unlock_Addr = 0x4000068c;
Cache_Disable_ICache = 0x40000690;
Cache_Enable_ICache = 0x40000694;
Cache_Suspend_ICache = 0x40000698;
Cache_Resume_ICache = 0x4000069c;
Cache_Freeze_ICache_Enable = 0x400006a0;
Cache_Freeze_ICache_Disable = 0x400006a4;
Cache_Set_IDROM_MMU_Size = 0x400006a8;
Cache_Get_IROM_MMU_End = 0x400006ac;
Cache_Get_DROM_MMU_End = 0x400006b0;
Cache_MMU_Init = 0x400006b4;
Cache_MSPI_MMU_Set = 0x400006b8;
Cache_Travel_Tag_Memory = 0x400006bc;
Cache_Get_Virtual_Addr = 0x400006c0;
