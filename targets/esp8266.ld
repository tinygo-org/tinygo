/* Linker script for the ESP8266 */

MEMORY
{
    /* Data RAM. Allows byte access. */
    DRAM  (rw) : ORIGIN = 0x3FFE8000, LENGTH = 80K
    /* Instruction RAM. */
    IRAM  (x)  : ORIGIN = 0x40100000, LENGTH = 32K
}

/* The entry point. It is set in the image flashed to the chip, so must be
 * defined.
 */
ENTRY(main)

SECTIONS
{
    /* Mutable global variables.
     */
    .data : ALIGN(4)
    {
        _sdata = ABSOLUTE(.);
        *(.data)
        *(.data.*)
    } >DRAM

    /* Constant global variables.
     * Note that they still need to be loaded in RAM because the ESP8266 doesn't
     * allow byte access to flash.
     */
    .rodata : ALIGN(4)
    {
        *(.rodata)
        *(.rodata.*)
    } >DRAM

    /* Global variables that are mutable and zero-initialized.
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

    /* Constant literals and code. Loaded into IRAM for now. Eventually, most
     * code should be executed directly from flash.
     * Note that literals must be before code for the l32r instruction to work.
     */
    .text : ALIGN(4)
    {
        *(.literal .text)
        *(.literal.* .text.*)
    } >IRAM
}

_globals_start = _sdata;
_globals_end = _ebss;
_heap_start = _ebss;
_heap_end = ORIGIN(DRAM) + LENGTH(DRAM);

/* It appears that the stack is set to 0x3ffffff0 when main is called.
 * Be conservative and scan all the way up to the end of the RAM.
 */
_stack_top = 0x40000000;

/* This is the stack bottom according to this forum post:
 * https://bbs.espressif.com/viewtopic.php?t=8879#p19038
 */
_stack_bottom = 0x3fffeb2c;

/* Functions normally provided by a libc.
 * Source:
 * https://github.com/espressif/ESP8266_NONOS_SDK/blob/master/ld/eagle.rom.addr.v6.ld
 */
memcpy  = 0x4000df48;
memmove = 0x4000e04c;
memset  = 0x4000e190;

/* Compiler runtime functions provided by the ROM.
 * Source:
 * https://github.com/espressif/ESP8266_NONOS_SDK/blob/master/ld/eagle.rom.addr.v6.ld
 */
__adddf3      = 0x4000c538;
__addsf3      = 0x4000c180;
__divdf3      = 0x4000cb94;
__divdi3      = 0x4000ce60;
__divsi3      = 0x4000dc88;
__extendsfdf2 = 0x4000cdfc;
__fixdfsi     = 0x4000ccb8;
__fixunsdfsi  = 0x4000cd00;
__fixunssfsi  = 0x4000c4c4;
__floatsidf   = 0x4000e2f0;
__floatsisf   = 0x4000e2ac;
__floatunsidf = 0x4000e2e8;
__floatunsisf = 0x4000e2a4;
__muldf3      = 0x4000c8f0;
__muldi3      = 0x40000650;
__mulsf3      = 0x4000c3dc;
__subdf3      = 0x4000c688;
__subsf3      = 0x4000c268;
__truncdfsf2  = 0x4000cd5c;
__udivdi3     = 0x4000d310;
__udivsi3     = 0x4000e21c;
__umoddi3     = 0x4000d770;
__umodsi3     = 0x4000e268;
__umulsidi3   = 0x4000dcf0;

/* Proprietary ROM function needed for proper clock configuration.
 */
rom_i2c_writeReg = 0x400072d8;
