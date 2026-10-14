MEMORY
{
  ITCM      (rwx): ORIGIN = 0x00000000, LENGTH = 0x00080000 /*  512 Kib */
  DTCM      (rwx): ORIGIN = 0x20000000, LENGTH = 0x00080000 /*  512 Kib */
  RAM       (rwx): ORIGIN = 0x20200000, LENGTH = 0x0007F000 /*  512 KiB minus 4 KiB reserved for USB DMA descriptors/buffers */
  /* The last 64 KiB of the flash are reserved for machine.Flash data. */
  FLASH      (rx): ORIGIN = 0x60000000, LENGTH = 0x001F0000 /* 1984 KiB */
}

ENTRY(Reset_Handler);

_stack_size = 4K;

SECTIONS
{
  .text : ALIGN(8) {

    FILL(0xFFFFFFFF);

    /* place flash config at beginning of flash device */
    KEEP(*(.flash_config));

    /* IVT must be located at +4 Kbyte offset from base address of flash. */
    . = ORIGIN(FLASH) + 0x1000;
    KEEP(*(.ivt));

    . = ORIGIN(FLASH) + 0x1020;
    KEEP(*(.boot_data));

    . = ORIGIN(FLASH) + 0x2000;

    _svectors = ABSOLUTE(.);
    KEEP(*(.isr_vector));
    . = ALIGN(8);

    *(.text.Reset_Handler);
    . = ALIGN(8);

    _stext = .;
    *(.text*);
    *(.rodata* .constdata*);
    . = ALIGN(8);
    _etext = .;

  } > FLASH

  .tinygo_stacksizes : ALIGN(8) {

    *(.tinygo_stacksizes);
    . = ALIGN(8);

  } > FLASH

  /* Code that runs from OCRAM2. The flash gives no data during an erase
     or a program operation. See machine_mimxrt1062_flash.go. */
  .ramfuncs : ALIGN(8) {

    _sramfuncs = .;
    *(.ramfuncs .ramfuncs.*);
    . = ALIGN(16);
    _eramfuncs = .;

  } > RAM AT > FLASH

  _framfuncs = LOADADDR(.ramfuncs);

  .text.padding (NOLOAD) : {

    . = ALIGN(32768);

  } > ITCM

  .stack (NOLOAD) : {

    . = ALIGN(8);
    . += _stack_size;
    _stack_top = .;

  } > DTCM

  .data : ALIGN(8) {

    FILL(0xFFFFFFFF);

    _sdata = .;
    *(.data*);
    . = ALIGN(8);
    _edata = .;

  } > DTCM AT > FLASH

  .bss : ALIGN(8) {

    _sbss = .;
    *(.bss*);
    *(COMMON);
    . = ALIGN(8);
    _ebss = .;

  } > DTCM AT > DTCM

  /DISCARD/ : {

    *(.ARM.exidx*); /* causes spurious 'undefined reference' errors */

  }

  _sidata = LOADADDR(.data);

  _heap_start = _eramfuncs;
  _heap_end = ORIGIN(RAM) + LENGTH(RAM);

  /* USB DMA region, the 4 KiB reserved above RAM. */
  _usb_dma_start = ORIGIN(RAM) + LENGTH(RAM);
  ASSERT((_usb_dma_start & 0x7FF) == 0, "_usb_dma_start must be 2 KiB aligned for ENDPTLISTADDR")

  /* Fixed data area for machine.Flash. The Teensy bootloader does not
     erase it, see Teensyduino cores/teensy4/eeprom.c FLASH_BASEADDR. */
  __flash_data_start = ORIGIN(FLASH) + LENGTH(FLASH);
  __flash_data_end = 0x601FF000;

  _globals_start = _sdata;
  _globals_end = _ebss;

  _image_size = SIZEOF(.text) + SIZEOF(.tinygo_stacksizes) + SIZEOF(.ramfuncs) + SIZEOF(.data);

  /* TODO: link .text to ITCM */
  _itcm_blocks = (0 + 0x7FFF) >> 15;
  _flexram_cfg = 0xAAAAAAAA | ((1 << (_itcm_blocks * 2)) - 1);
}
