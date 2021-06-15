MEMORY
{
  FLASH   (rx) : ORIGIN = 0x08040000, LENGTH = 768K /* M7 code + bootloader */
  ITCM   (rwx) : ORIGIN = 0x00000000, LENGTH =  64K /* M7 VTOR/ISRs */
  DTCM   (rwx) : ORIGIN = 0x20000000, LENGTH = 128K /* M7 stack/heap */
  SRAMD1 (rwx) : ORIGIN = 0x24000000, LENGTH = 512K /* M7 globals */
  SRAMD2 (rwx) : ORIGIN = 0x30000000, LENGTH = 288K /* M4 stack/heap/globals */
  SRAMD3 (rwx) : ORIGIN = 0x38000000, LENGTH =  64K /* M7+M4 shared */
}

ENTRY(Reset_Handler);

_stack_size = 4K;

SECTIONS
{
  /* Program code and read-only data goes to FLASH. */
  .text : {

    _svectors = ABSOLUTE(.);
    /* Copied to start of ITCM during init */
    KEEP(*(.isr_vector));
    . = ALIGN(32);
    _evectors = ABSOLUTE(.);
    *(.text*);
    *(.rodata*);
    . = ALIGN(8);

  } > FLASH

  .tinygo_stacksizes : {

    *(.tinygo_stacksizes)

  } > FLASH

  .critical (NOLOAD) : {

    _svtor = ABSOLUTE(.);
    /* Reserve space for VTOR at start of ITCM */
    . += _evectors - _svectors;
    . = ALIGN(32);
    /* User-defined ISRs placed in ITCM */
    *(.fastrun);
    *(.isr);

  } > ITCM

  .stack (NOLOAD) : {

    . = ALIGN(8);
    . += _stack_size;
    _stack_top = .;

  } > DTCM

  .data : ALIGN(8) {

    _sdata = .;
    *(.data*);
    . = ALIGN(8);
    _edata = .;

  } > SRAMD1 AT > FLASH

  .bss : ALIGN(8) {

    _sbss = .;
    *(.bss*);
    *(COMMON);
    . = ALIGN(8);
    _ebss = .;

  } > SRAMD1

  /DISCARD/ : {

    *(.ARM.exidx*) /* causes spurious linker errors */

  }

  _sidata = LOADADDR(.data); /* Used for initialization */

  /* Remaining items are for the TinyGo memory allocator. */

  _heap_start = _stack_top;
  _heap_end = ORIGIN(DTCM) + LENGTH(DTCM);

  _globals_start = _sdata;
  _globals_end = _ebss;

}
