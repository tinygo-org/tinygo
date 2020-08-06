/* Linker script for the ESP32 */

MEMORY
{
    /* Data RAM. Allows byte access.
     * There are various data RAM regions:
     *   SRAM2: 0x3FFA_E000..0x3FFD_FFFF (72 + 128 = 200K)
     *   SRAM1: 0x3FFE_0000..0x3FFF_FFFF (128K)
     * This gives us 328K of contiguous RAM, which is the largest span possible.
     * SRAM1 has other addresses as well but the datasheet seems to indicate
     * these are aliases.
     */
    DRAM  (rw) : ORIGIN = 0x3FFAE000, LENGTH = 200K + 128K /* Internal SRAM 1 + 2 */

    /* Instruction RAM. */
    IRAM  (x)  : ORIGIN = 0x40080000, LENGTH = 128K /* Internal SRAM 0 */
}

/* The entry point. It is set in the image flashed to the chip, so must be
 * defined.
 */
ENTRY(call_start_cpu0)

SECTIONS
{
    /* Constant literals and code. Loaded into IRAM for now. Eventually, most
     * code should be executed directly from flash.
     * Note that literals must be before code for the l32r instruction to work.
     */
    .text : ALIGN(4)
    {
        *(.literal.text.call_start_cpu0)
        *(.text.call_start_cpu0)
        *(.literal .text)
        *(.literal.* .text.*)
    } >IRAM

    /* Put the stack at the bottom of DRAM, so that the application will
     * crash on stack overflow instead of silently corrupting memory.
     * See: http://blog.japaric.io/stack-overflow-protection/ */
    .stack (NOLOAD) :
    {
        . = ALIGN(16);
        . += _stack_size;
        _stack_top = .;
    } >DRAM

    /* Constant global variables.
     * They are loaded in DRAM for ease of use. Eventually they should be stored
     * in flash and loaded directly from there but they're kept in RAM to make
     * sure they can always be accessed (even in interrupts).
     */
    .rodata : ALIGN(4)
    {
        *(.rodata)
        *(.rodata.*)
    } >DRAM

    /* Mutable global variables.
     */
    .data : ALIGN(4)
    {
        _sdata = ABSOLUTE(.);
        *(.data)
        *(.data.*)
        _edata = ABSOLUTE(.);
    } >DRAM

    /* Check that the boot ROM stack (for the APP CPU) does not overlap with the
     * data that is loaded by the boot ROM. There may be ways to avoid this
     * issue if it occurs in practice.
     * The magic value here is _stack_sentry in the boot ROM ELF file.
     */
    ASSERT(_edata < 0x3ffe1320, "the .data section overlaps with the stack used by the boot ROM, possibly causing corruption at startup")

    /* Global variables that are mutable and zero-initialized.
     * These must be zeroed at startup (unlike data, which is loaded by the
     * bootloader).
     */
    .bss (NOLOAD) : ALIGN(4)
    {
        . = ALIGN (4);
        _sbss = ABSOLUTE(.);
        *(.bss)
        *(.bss.*)
        . = ALIGN (4);
        _ebss = ABSOLUTE(.);
    } >DRAM
}

/* For the garbage collector.
 */
_globals_start = _sdata;
_globals_end = _ebss;
_heap_start = _ebss;
_heap_end = ORIGIN(DRAM) + LENGTH(DRAM);

_stack_size = 4K;

/* From ESP-IDF, included here as long as picolibc doesn't compile.
 */
memcpy    = 0x4000c2c8;
memset    = 0x4000c44c;
__udivdi3 = 0x4000cff8;
