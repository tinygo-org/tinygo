/* Linker script for the ESP32-P4
 *
 * The ESP32-P4 memory layout:
 *   - It has 768KB of HP L2MEM (internal SRAM) with a unified I/D address
 *     space at 0x4FF00000. The first 128KB is reserved for L2 cache, so
 *     usable SRAM starts at 0x4FF20000 (640KB).
 *   - It has 32KB of LP-SRAM at 0x50108000 for low-power / RTC use.
 *   - Flash is mapped via a unified SPI0 MMU at 0x40000000 for both code
 *     (IROM) and read-only data (DROM).
 *   - The MMU works in pages of 64KB, which means the bottom 16 bits of the
 *     address in flash and the address in the mapped region need to match.
 *   - Memory in SRAM is loaded at reset by the ROM bootloader.
 *
 * Flash image layout (image flashed at flash offset 0x2000):
 *   [0x2000..]: SRAM image — ROM image header + .data/.iram segments + footer
 *   [0x10000..]: .text data  (VMA 0x40010000, MMU-mapped)
 *   [...]:       .rodata data (VMA 0x40010000 + sizeof(.text), MMU-mapped)
 *
 * The ROM bootloader loads only SRAM segments (.data, .iram) into HP-SRAM.
 * Flash-mapped sections (.text, .rodata) are NOT ROM image segments; they are
 * appended at flash offset 0x10000 by esp.go. The L2 cache MMU (set up by the
 * ROM before calling the application entry point) maps flash[0x10000] to
 * VMA 0x40010000, making .text/.rodata directly accessible.
 *
 * This avoids the ROM's watchdog-triggering behaviour of memcpy-ing large
 * flash-mapped segments to their VMA in the flash cache address space.
 *
 * Like the ESP32-C6, the ESP32-P4 uses one address space for IROM and DROM,
 * here at 0x40000000. We use a single FLASH memory region to avoid lld
 * virtual-address-overlap errors.
 */

MEMORY
{
    /* HP L2MEM: unified IRAM/DRAM address space.
     * Total: 768 KB (0x4FF00000 – 0x4FFC0000)
     * L2 cache uses the first 128 KB, so usable starts at 0x4FF20000.
     * ESP-IDF limits application SRAM to 0x4FFADFC0 on rev3. On rev<3 the
     * ROM uses 0x4FF2BBD0–0x4FF40000 for BSS, data and stack, so start above
     * that range to avoid placing the stack or heap on ROM data.
     */
    SRAM  (rwx) : ORIGIN = 0x4FF40000, LENGTH = 0x6DFC0

    /* Flash-mapped via unified SPI0 MMU.
     * A single region covers both IROM (code) and DROM (rodata).
     */
    FLASH (rx)  : ORIGIN = 0x40000000, LENGTH = 64M
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

    /* Mutable global variables. This data (in the SRAM segment) is initialized
     * by the ROM bootloader.  Placed before .bss so that .iram (startup code)
     * always lands at a low, stable SRAM address regardless of .bss size.
     */
    .data : ALIGN(4)
    {
        . = ALIGN (4);
        PROVIDE( __global_pointer$ = . + (4K / 2) );
        _sdata = ABSOLUTE(.);
        *(.sdata)
        *(.data .data.*)
        *(.dram*)
        . = ALIGN (4);
        _edata = ABSOLUTE(.);
    } >SRAM

    /* Code that must run from IRAM (not flash-mapped), for example interrupt
     * vectors and code that runs before the flash has been mapped.
     * Placed before .bss so that a large .bss (e.g. a display framebuffer)
     * cannot push .iram beyond the ROM bootloader's effective load range.
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

    /* Global variables that are mutable and zero-initialized.
     * Placed after .data/.iram so that growing this section (e.g. a large
     * display framebuffer) does not displace the startup code toward the
     * SRAM ceiling where the ROM bootloader may fail to load it.
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

    /* Heap starts after IRAM. */
    . = ALIGN(4);
    _heap_start = ABSOLUTE(.);
    _heap_end = ORIGIN(SRAM) + LENGTH(SRAM);

    /* === FLASH sections (flash-mapped via MMU) === */

    /* Dummy section to align .text virtual address with its flash file offset.
     *
     * The ESP32-P4 builder (esp.go) separates flash-mapped segments (.text,
     * .rodata) from SRAM segments (.data, .iram). Only SRAM segments are
     * written as ROM image segments. Flash-mapped code is appended at flash
     * physical offset 0x10000 (the second 64KB page, i.e. image offset 0xE000).
     *
     * The L2 cache MMU maps flash[0x10000] → VMA 0x40010000, so .text must
     * start at VMA 0x40010000.
     *
     * Image layout (flash offset 0x2000 = image base):
     *   [0x2000..0x2xxx]: SRAM image (header + .data/.iram segments + footer)
     *   [0x10000..]:      .text data  (VMA 0x40010000)
     *   [...]:            .rodata data (VMA 0x40010000 + sizeof(.text))
     */
    .flash_dummy (NOLOAD) :
    {
        . = ORIGIN(FLASH) + 0x10000; /* flash code starts at 0x10000, VMA 0x40010000 */
    } > FLASH

    /* Code stored in FLASH (flash-mapped).
     * This is the main program code.
     */
    .text : ALIGN(4)
    {
        *(.text .text.*)
    } > FLASH

    /* Constant global variables, stored in FLASH (read-only data).
     * Follows .text immediately — no dummy needed since there is no segment
     * header between .text and .rodata in the flash image.
     */
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

/* For the GC and scheduler.
 * Must span all mutable global data: start of .data through end of .bss.
 * (.iram lies between them and contains only code; scanning it is harmless
 *  for the conservative GC.)
 */
_globals_start = _sdata;
_globals_end = _ebss;

/* Default stack size.  go-mp3 and other decoders need well over 4 KB. */
_stack_size = 32K;

/* ROM function addresses for ESP32-P4.
 * Source: ESP-IDF components/esp_rom/esp32p4/ld/esp32p4.rom.ld
 */
PROVIDE( ets_delay_us = 0x4fc0003c );

/* Cache / MMU ROM functions called from call_start_cpu0.
 * Source: ESP-IDF components/esp_rom/esp32p4/ld/esp32p4.rom.ld
 */
Cache_Invalidate_All        = 0x4fc00404;
Cache_FLASH_MMU_Set         = 0x4fc00518;
