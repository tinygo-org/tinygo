/* Linker script for the ESP32 */

MEMORY
{
    /* Data RAM. Allows byte access.
     * There are various data RAM regions:
     *   SRAM1: 0x3FC8_8000..0x3FCE_FFFF (416K)
     *   SRAM2: 0x3FCF_0000..0x3FCF_FFFF (64K)
     * This gives us 480k of contiguous RAM, which is the largest span possible.
     */
    DRAM  (rw) : ORIGIN = 0x3FC88000, LENGTH = 416K /* Internal SRAM 1 + 2 */

    /* Instruction RAM. */
    IRAM  (x)  : ORIGIN = 0x40370000, LENGTH = 32K /* Internal SRAM 0 */
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
