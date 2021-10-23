/* Linker script for the ESP32-C3
 *
 * The ESP32-C3 has a rather funky memory layout, more like its Xtensa
 * predecessors than like other RISC-V chips:
 *   - It has 384kB of regular RAM. This RAM can be used both as data RAM and
 *     instruction RAM, but needs to be accessed via a different address space
 *     (DRAM/IRAM).
 *   - It has another 16kB of RAM, that could be used as regular RAM but is
 *     normally used by the flash cache.
 *   - It has 8MB of address space for the DROM and IROM, but for some reason
 *     this address space is shared. So it isn't possible to map all DROM at
 *     0x3C000000 and all DRAM at 0x42000000: they would overlap.
 *   - The MMU works in pages of 64kB, which means the bottom 16 bits of the
 *     address in flash and the address in DROM/IROM need to match.
 *   - Memory in DRAM and IRAM is loaded at reset by the ROM bootloader.
 *     Luckily, this doesn't have significant alignment requirements.
 *
 * This linker script has been written to carefully work around (or with) these
 * limitations:
 *   - It adds dummy sections so that the bottom 16 bits of the virtual address
 *     and the physical address (in the generated firmware image) match.
 *   - It also offsets sections that share an address space using those same
 *     dummy sections.
 *   - It sorts the sections by load address, to avoid surprises as esptool.py
 *     also does it.
 * This way, it's possible to create a very small firmware image that still
 * conforms to the expectations of esptool.py and the ROM bootloader.
 */

MEMORY
{
    /* Note: DRAM and IRAM below are actually in the same 384K address space. */
    DRAM (rw) : ORIGIN = 0x3FC80000, LENGTH = 384K /* Internal SRAM 1 (data bus) */
    IRAM (x)  : ORIGIN = 0x40380000, LENGTH = 384K /* Internal SRAM 1 (instruction bus) */

    /* Note: DROM and IROM below are actually in the same 8M address space. */
    DROM (r)  : ORIGIN = 0x3C000000, LENGTH = 8M /* Data bus (read-only) */
    IROM (rx) : ORIGIN = 0x42000000, LENGTH = 8M /* Instruction bus */
}

/* The entry point. It is set in the image flashed to the chip, so must be
 * defined.
 */
ENTRY(call_start_cpu0)

SECTIONS
{
    /* Dummy section to make sure the .rodata section starts exactly behind the
     * image header.
     */
    .rodata_dummy (NOLOAD): ALIGN(4)
    {
        . += 0x18; /* image header at start of flash: esp_image_header_t */
        . += 0x8;  /* DROM segment header (8 bytes) */
    } > DROM

    /* Constant global variables, stored in DROM.
     */
    .rodata : ALIGN(4)
    {
        *(.rodata .rodata.*)
        . = ALIGN (4);
    } >DROM

    /* Put the stack at the bottom of DRAM, so that the application will
     * crash on stack overflow instead of silently corrupting memory.
     * See: http://blog.japaric.io/stack-overflow-protection/
     * TODO: this might not actually work because memory protection hasn't been set up.
     */
    .stack (NOLOAD) :
    {
        . = ALIGN(16);
        . += _stack_size;
        _stack_top = .;
    } >DRAM

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
    } >DRAM

    /* Mutable global variables. This data (in the DRAM segment) is initialized
     * by the ROM bootloader.
     */
    .data : ALIGN(4)
    {
        . = ALIGN (4);
        _sdata = ABSOLUTE(.);
        *(.sdata)
        *(.data .data.*)
        *(.dram1 .dram1.*)
        . = ALIGN (4);
        _edata = ABSOLUTE(.);
    } >DRAM

    /* Dummy section to make sure the .init section (in the IRAM segment) is just
     * behind the DRAM segment. For IRAM and DRAM, we luckily don't have to
     * worry about 64kB pages or image headers as they're loaded in RAM by the
     * bootloader (not mapped from flash).
     */
    .iram_dummy (NOLOAD): ALIGN(4)
    {
        . += SIZEOF(.stack);
        . += SIZEOF(.bss);
        . += SIZEOF(.data);
    } > IRAM

    /* Initialization code is loaded into IRAM. This memory area is also used by
     * the heap, so no RAM is wasted.
     */
    .init : ALIGN(4)
    {
        *(.init)
        *(.iram1 .iram1.*)
        *(.wifi0iram .wifi0iram.*)
        *(.wifirxiram .wifirxiram.*)
        *(.wifislprxiram .wifislprxiram.*)
    } >IRAM

    /* Dummy section to put the IROM segment exactly behind the IRAM segment.
     * This has to follow the app image format exactly.
     */
    .text_dummy (NOLOAD): ALIGN(4)
    {
        /* Note: DRAM and DROM are not always present so the header should only
         * be inserted if it actually exists.
         */
        . += 0x18;                                                 /* esp_image_header_t */
        . += SIZEOF(.rodata) + ((SIZEOF(.rodata) != 0) ? 0x8 : 0); /* DROM segment (optional) */
        . += SIZEOF(.data)   + ((SIZEOF(.data)   != 0) ? 0x8 : 0); /* DRAM segment (optional) */
        . += SIZEOF(.init)   + 0x8;                                /* IRAM segment */
        . += 0x8;                                                  /* IROM segment header */
    } > IROM

    /* IROM segment. This contains all the actual code and is placed right after
     * the DROM segment.
     */
    .text : ALIGN(4)
    {
        . = ALIGN (256);
        *(.text.exception_vectors)
        . = ALIGN (4);
        *(.text .text.*)
        *(.wifislpiram .wifislpiram.*)
    } >IROM

    /DISCARD/ :
    {
        *(.eh_frame)       /* causes 'no memory region specified' error in lld */
    }

    /* Check that the boot ROM stack (for the APP CPU) does not overlap with the
     * data that is loaded by the boot ROM. This is unlikely to happen in
     * practice. 
     * The magic value comes from here:
     * https://github.com/espressif/esp-idf/blob/61299f879e/components/bootloader/subproject/main/ld/esp32c3/bootloader.ld#L191
     */
    ASSERT((_edata + SIZEOF(.init)) < 0x3FCDE710, "the .init section overlaps with the stack used by the boot ROM, possibly causing corruption at startup")
}

/* For the garbage collector.
 * Note that _heap_start starts after _edata (without caring for the .init
 * section), because the .init section isn't necessary anymore after startup and
 * can thus be overwritten by the heap.
 */
_globals_start = _sbss;
_globals_end = _edata;
_heap_start = _edata;
_heap_end = ORIGIN(DRAM) + LENGTH(DRAM);

_stack_size = 4K;

/* ROM functions used for setting up the flash mapping.
 */
Cache_Invalidate_ICache_All = 0x400004d8;
Cache_Suspend_ICache        = 0x40000524;
Cache_Resume_ICache         = 0x40000528;
Cache_MMU_Init              = 0x4000055c;
Cache_Dbus_MMU_Set          = 0x40000564;

/* From ESP-IDF:
 * components/esp_rom/esp32c3/ld/esp32c3.rom.libgcc.ld
 * These are called from LLVM during codegen. The original license is Apache
 * 2.0.
 */
__absvdi2     = 0x40000764;
__absvsi2     = 0x40000768;
__adddf3      = 0x4000076c;
__addsf3      = 0x40000770;
__addvdi3     = 0x40000774;
__addvsi3     = 0x40000778;
__ashldi3     = 0x4000077c;
__ashrdi3     = 0x40000780;
__bswapdi2    = 0x40000784;
__bswapsi2    = 0x40000788;
__clear_cache = 0x4000078c;
__clrsbdi2    = 0x40000790;
__clrsbsi2    = 0x40000794;
__clzdi2      = 0x40000798;
__clzsi2      = 0x4000079c;
__cmpdi2      = 0x400007a0;
__ctzdi2      = 0x400007a4;
__ctzsi2      = 0x400007a8;
__divdc3      = 0x400007ac;
__divdf3      = 0x400007b0;
__divdi3      = 0x400007b4;
__divsc3      = 0x400007b8;
__divsf3      = 0x400007bc;
__divsi3      = 0x400007c0;
__eqdf2       = 0x400007c4;
__eqsf2       = 0x400007c8;
__extendsfdf2 = 0x400007cc;
__ffsdi2      = 0x400007d0;
__ffssi2      = 0x400007d4;
__fixdfdi     = 0x400007d8;
__fixdfsi     = 0x400007dc;
__fixsfdi     = 0x400007e0;
__fixsfsi     = 0x400007e4;
__fixunsdfsi  = 0x400007e8;
__fixunssfdi  = 0x400007ec;
__fixunssfsi  = 0x400007f0;
__floatdidf   = 0x400007f4;
__floatdisf   = 0x400007f8;
__floatsidf   = 0x400007fc;
__floatsisf   = 0x40000800;
__floatundidf = 0x40000804;
__floatundisf = 0x40000808;
__floatunsidf = 0x4000080c;
__floatunsisf = 0x40000810;
__gcc_bcmp    = 0x40000814;
__gedf2       = 0x40000818;
__gesf2       = 0x4000081c;
__gtdf2       = 0x40000820;
__gtsf2       = 0x40000824;
__ledf2       = 0x40000828;
__lesf2       = 0x4000082c;
__lshrdi3     = 0x40000830;
__ltdf2       = 0x40000834;
__ltsf2       = 0x40000838;
__moddi3      = 0x4000083c;
__modsi3      = 0x40000840;
__muldc3      = 0x40000844;
__muldf3      = 0x40000848;
__muldi3      = 0x4000084c;
__mulsc3      = 0x40000850;
__mulsf3      = 0x40000854;
__mulsi3      = 0x40000858;
__mulvdi3     = 0x4000085c;
__mulvsi3     = 0x40000860;
__nedf2       = 0x40000864;
__negdf2      = 0x40000868;
__negdi2      = 0x4000086c;
__negsf2      = 0x40000870;
__negvdi2     = 0x40000874;
__negvsi2     = 0x40000878;
__nesf2       = 0x4000087c;
__paritysi2   = 0x40000880;
__popcountdi2 = 0x40000884;
__popcountsi2 = 0x40000888;
__powidf2     = 0x4000088c;
__powisf2     = 0x40000890;
__subdf3      = 0x40000894;
__subsf3      = 0x40000898;
__subvdi3     = 0x4000089c;
__subvsi3     = 0x400008a0;
__truncdfsf2  = 0x400008a4;
__ucmpdi2     = 0x400008a8;
__udivdi3     = 0x400008ac;
__udivmoddi4  = 0x400008b0;
__udivsi3     = 0x400008b4;
__udiv_w_sdiv = 0x400008b8;
__umoddi3     = 0x400008bc;
__umodsi3     = 0x400008c0;
__unorddf2    = 0x400008c4;
__unordsf2    = 0x400008c8;

/* From ESP-IDF:
 * components/esp_rom/esp32c3/ld/esp32c3.rom.newlib.ld
 * These are called during codegen and thus it's a good idea to make them always
 * available. ROM functions may also be faster than functions in IROM (that go
 * through the flash cache) and are always available in interrupts.
 */
memset  = 0x40000354;
memcpy  = 0x40000358;
memmove = 0x4000035c;
