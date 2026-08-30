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
        *(.sbss.*)
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
        *(.sdata.*)
        *(.data .data.*)
        *(.dram*)
        . = ALIGN (4);
        _edata = ABSOLUTE(.);
        /* Keep the image segment a multiple of 8; see the MMU note below. */
        . = ALIGN (8);
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

        /* WiFi/BLE/coex blob code that must run from IRAM. The Espressif
         * binary libraries place their hot paths and ISR code in these
         * sections. Without collecting them here they fall through to .text
         * in flash, which crashes when they are called from an interrupt or
         * while the flash cache is disabled.
         */
        *(.wifi0iram .wifi0iram.*)
        *(.wifirxiram .wifirxiram.*)
        *(.wifislpiram .wifislpiram.*)
        *(.wifislprxiram .wifislprxiram.*)
        *(.wifiextrairam .wifiextrairam.*)
        *(.wifiorslpiram .wifiorslpiram.*)
        *(.coexiram .coexiram.*)
        *(.iram1 .iram1.*)

        *(.text.handleInterruptASM) /* must be in SRAM, within JAL range of _vector_table */
        . = ALIGN(256);
        *(.text.exception_vectors)
        . = ALIGN(4);
        /* Keep the image segment a multiple of 8; see the MMU note below. */
        . = ALIGN(8);
    } > SRAM

    /* Heap starts after IRAM.
     *
     * The heap must NOT run to the end of SRAM. The mask ROM keeps its stack
     * and its data/bss at the top of HP-SRAM: SOC_ROM_STACK_START is
     * 0x4087e610 with a 0x2000 stack below it, and ROM variables live above
     * that -- including phy_param_rom at 0x4087fce8, which the PHY blob reads.
     * Letting the Go heap grow into that region corrupts the radio.
     * ESP-IDF and esp-hal both treat everything above 0x4087e610 as reserved.
     */
    . = ALIGN(4);
    _heap_start = ABSOLUTE(.);
    _heap_end = 0x4087C610;

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
        /* Keep the image segment a multiple of 8; see the MMU note below. */
        . = ALIGN(8);
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

    /* Constant global variables, stored in DROM.
     *
     * The flash MMU maps 64kB pages, so a flash-mapped section is only read
     * correctly when its virtual address and its offset within the firmware
     * image agree modulo 64kB. The dummy section above reproduces the image
     * offset by accumulating the preceding segment sizes and headers, but that
     * only works if the linker does not then insert alignment padding between
     * the dummy and this section: the padding moves the virtual address without
     * moving the image offset. lld gives .rodata 8-byte alignment, so every
     * preceding segment is padded to a multiple of 8 above to keep the two in
     * step. Getting this wrong shifts the whole of .rodata by a few bytes, which
     * is silent and looks like arbitrary data corruption rather than a mapping
     * bug -- it was found via a WiFi blob rejecting a struct whose first field
     * read back as a code pointer. The ASSERT after this section catches it.
     */
    .rodata : ALIGN(4)
    {
        *(.rodata*)
        *(.srodata*)
        . = ALIGN (4);
    } >DROM

    /* The flash MMU constraint described above, enforced at link time for both
     * flash-mapped sections. */
    ASSERT(((ADDR(.text) - ORIGIN(IROM)) & 0xffff) ==
           ((0x18 + ((SIZEOF(.data) > 0) ? (SIZEOF(.data) + 0x8) : 0)
                  + SIZEOF(.iram) + 0x8 + 0x8) & 0xffff),
           ".text virtual address and flash offset disagree modulo 64kB")
    ASSERT(((ADDR(.rodata) - ORIGIN(DROM)) & 0xffff) ==
           ((0x18 + ((SIZEOF(.data) > 0) ? (SIZEOF(.data) + 0x8) : 0)
                  + SIZEOF(.iram) + 0x8 + SIZEOF(.text) + 0x8 + 0x8) & 0xffff),
           ".rodata virtual address and flash offset disagree modulo 64kB")

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
Cache_Invalidate_ICache_All = 0x4000064c;
Cache_Suspend_ICache = 0x40000698;
Cache_Resume_ICache = 0x4000069c;
Cache_MMU_Init = 0x400006b4;
Cache_MSPI_MMU_Set = 0x400006b8;

/* PHY ROM data */
phy_param_rom = 0x4087fce8;

/* ROM function and data addresses for the ESP32-C6.
 *
 * The Espressif WiFi/PHY binary libraries resolve a large part of their
 * internals against the mask ROM rather than shipping the code, so these
 * tables are required to link espradio. They are the ESP-IDF generated
 * interface files, vendored via esp-hal at esp-rom-sys/ld/esp32c6/rom/*.ld
 * (generated from target/esp32c6/interface-esp32c6.yml,
 * md5 06c13e133e0743d09b87aba30d3e213b).
 *
 * ROM *functions* (0x4000xxxx) use PROVIDE(), so a real definition in the
 * program (picolibc, compiler-rt, or a vendored library) takes precedence over
 * the ROM copy.
 *
 * ROM *data* (0x408xxxxx) is assigned unconditionally instead, matching
 * esp32c3.ld. These are variables in RAM that mask ROM code reads and writes at
 * a fixed address, so they are shared storage rather than a fallback: with
 * PROVIDE(), a weak definition elsewhere in the program wins and the program
 * and the ROM then each use a different location for the same variable.
 */

/* --- esp32c6.rom.ld --- */
PROVIDE( rtc_get_reset_reason = 0x40000018 );
PROVIDE( analog_super_wdt_reset_happened = 0x4000001c );
PROVIDE( rtc_get_wakeup_cause = 0x40000020 );
PROVIDE( rtc_unhold_all_pads = 0x40000024 );
PROVIDE( ets_printf = 0x40000028 );
PROVIDE( ets_install_putc1 = 0x4000002c );
PROVIDE( ets_install_putc2 = 0x40000030 );
PROVIDE( ets_install_uart_printf = 0x40000034 );
PROVIDE( ets_install_usb_printf = 0x40000038 );
PROVIDE( ets_get_printf_channel = 0x4000003c );
PROVIDE( ets_get_cpu_frequency = 0x40000044 );
PROVIDE( ets_update_cpu_frequency = 0x40000048 );
PROVIDE( ets_install_lock = 0x4000004c );
PROVIDE( UartRxString = 0x40000050 );
PROVIDE( UartGetCmdLn = 0x40000054 );
PROVIDE( uart_tx_one_char = 0x40000058 );
PROVIDE( uart_tx_one_char2 = 0x4000005c );
PROVIDE( uart_rx_one_char = 0x40000060 );
PROVIDE( uart_rx_one_char_block = 0x40000064 );
PROVIDE( uart_rx_intr_handler = 0x40000068 );
PROVIDE( uart_rx_readbuff = 0x4000006c );
PROVIDE( uartAttach = 0x40000070 );
PROVIDE( uart_tx_flush = 0x40000074 );
PROVIDE( uart_tx_wait_idle = 0x40000078 );
PROVIDE( uart_div_modify = 0x4000007c );
PROVIDE( ets_write_char_uart = 0x40000080 );
PROVIDE( uart_tx_switch = 0x40000084 );
PROVIDE( roundup2 = 0x40000088 );
PROVIDE( multofup = 0x4000008c );
PROVIDE( software_reset = 0x40000090 );
PROVIDE( software_reset_cpu = 0x40000094 );
PROVIDE( ets_clk_assist_debug_clock_enable = 0x40000098 );
PROVIDE( clear_super_wdt_reset_flag = 0x4000009c );
PROVIDE( disable_default_watchdog = 0x400000a0 );
PROVIDE( esp_rom_set_rtc_wake_addr = 0x400000a4 );
PROVIDE( esp_rom_get_rtc_wake_addr = 0x400000a8 );
PROVIDE( send_packet = 0x400000ac );
PROVIDE( recv_packet = 0x400000b0 );
PROVIDE( GetUartDevice = 0x400000b4 );
PROVIDE( UartDwnLdProc = 0x400000b8 );
PROVIDE( GetSecurityInfoProc = 0x400000bc );
PROVIDE( Uart_Init = 0x400000c0 );
PROVIDE( ets_set_user_start = 0x400000c4 );
PROVIDE( ets_rom_layout_p = 0x4004fffc );
ets_ops_table_ptr = 0x4087fff8;
g_saved_pc = 0x4087fffc;
PROVIDE( mz_adler32 = 0x400000c8 );
PROVIDE( mz_free = 0x400000cc );
PROVIDE( tdefl_compress = 0x400000d0 );
PROVIDE( tdefl_compress_buffer = 0x400000d4 );
PROVIDE( tdefl_compress_mem_to_heap = 0x400000d8 );
PROVIDE( tdefl_compress_mem_to_mem = 0x400000dc );
PROVIDE( tdefl_compress_mem_to_output = 0x400000e0 );
PROVIDE( tdefl_get_adler32 = 0x400000e4 );
PROVIDE( tdefl_get_prev_return_status = 0x400000e8 );
PROVIDE( tdefl_init = 0x400000ec );
PROVIDE( tdefl_write_image_to_png_file_in_memory = 0x400000f0 );
PROVIDE( tdefl_write_image_to_png_file_in_memory_ex = 0x400000f4 );
PROVIDE( tinfl_decompress = 0x400000f8 );
PROVIDE( tinfl_decompress_mem_to_callback = 0x400000fc );
PROVIDE( tinfl_decompress_mem_to_heap = 0x40000100 );
PROVIDE( tinfl_decompress_mem_to_mem = 0x40000104 );
PROVIDE( jd_prepare = 0x40000108 );
PROVIDE( jd_decomp = 0x4000010c );
PROVIDE( esp_rom_spiflash_wait_idle = 0x40000110 );
PROVIDE( esp_rom_spiflash_write_encrypted = 0x40000114 );
PROVIDE( esp_rom_spiflash_write_encrypted_dest = 0x40000118 );
PROVIDE( esp_rom_spiflash_write_encrypted_enable = 0x4000011c );
PROVIDE( esp_rom_spiflash_write_encrypted_disable = 0x40000120 );
PROVIDE( esp_rom_spiflash_erase_chip = 0x40000124 );
PROVIDE( _esp_rom_spiflash_erase_sector = 0x40000128 );
PROVIDE( _esp_rom_spiflash_erase_block = 0x4000012c );
PROVIDE( _esp_rom_spiflash_write = 0x40000130 );
PROVIDE( _esp_rom_spiflash_read = 0x40000134 );
PROVIDE( _esp_rom_spiflash_unlock = 0x40000138 );
PROVIDE( _SPIEraseArea = 0x4000013c );
PROVIDE( _SPI_write_enable = 0x40000140 );
PROVIDE( esp_rom_spiflash_erase_sector = 0x40000144 );
PROVIDE( esp_rom_spiflash_erase_block = 0x40000148 );
PROVIDE( esp_rom_spiflash_write = 0x4000014c );
PROVIDE( esp_rom_spiflash_read = 0x40000150 );
PROVIDE( esp_rom_spiflash_unlock = 0x40000154 );
PROVIDE( SPIEraseArea = 0x40000158 );
PROVIDE( SPI_write_enable = 0x4000015c );
PROVIDE( esp_rom_spiflash_config_param = 0x40000160 );
PROVIDE( esp_rom_spiflash_read_user_cmd = 0x40000164 );
PROVIDE( esp_rom_spiflash_select_qio_pins = 0x40000168 );
PROVIDE( esp_rom_spi_flash_auto_sus_res = 0x4000016c );
PROVIDE( esp_rom_spi_flash_send_resume = 0x40000170 );
PROVIDE( esp_rom_spi_flash_update_id = 0x40000174 );
PROVIDE( esp_rom_spiflash_config_clk = 0x40000178 );
PROVIDE( esp_rom_spiflash_config_readmode = 0x4000017c );
PROVIDE( esp_rom_spiflash_read_status = 0x40000180 );
PROVIDE( esp_rom_spiflash_read_statushigh = 0x40000184 );
PROVIDE( esp_rom_spiflash_write_status = 0x40000188 );
PROVIDE( spi_cache_mode_switch = 0x4000018c );
PROVIDE( spi_common_set_dummy_output = 0x40000190 );
PROVIDE( spi_common_set_flash_cs_timing = 0x40000194 );
PROVIDE( esp_rom_spi_set_address_bit_len = 0x40000198 );
PROVIDE( SPILock = 0x4000019c );
PROVIDE( SPIMasterReadModeCnfig = 0x400001a0 );
PROVIDE( SPI_Common_Command = 0x400001a4 );
PROVIDE( SPI_WakeUp = 0x400001a8 );
PROVIDE( SPI_block_erase = 0x400001ac );
PROVIDE( SPI_chip_erase = 0x400001b0 );
PROVIDE( SPI_init = 0x400001b4 );
PROVIDE( SPI_page_program = 0x400001b8 );
PROVIDE( SPI_read_data = 0x400001bc );
PROVIDE( SPI_sector_erase = 0x400001c0 );
PROVIDE( SelectSpiFunction = 0x400001c4 );
PROVIDE( SetSpiDrvs = 0x400001c8 );
PROVIDE( Wait_SPI_Idle = 0x400001cc );
PROVIDE( spi_dummy_len_fix = 0x400001d0 );
PROVIDE( Disable_QMode = 0x400001d4 );
PROVIDE( Enable_QMode = 0x400001d8 );
PROVIDE( spi_flash_attach = 0x400001dc );
PROVIDE( spi_flash_get_chip_size = 0x400001e0 );
PROVIDE( spi_flash_guard_set = 0x400001e4 );
PROVIDE( spi_flash_guard_get = 0x400001e8 );
PROVIDE( spi_flash_read_encrypted = 0x400001ec );
rom_spiflash_legacy_funcs = 0x4087fff0;
rom_spiflash_legacy_data = 0x4087ffec;
g_flash_guard_ops = 0x4087fff4;
PROVIDE( esp_rom_spiflash_write_disable = 0x40000278 );
PROVIDE( Cache_Get_ICache_Line_Size = 0x40000628 );
PROVIDE( Cache_Get_Mode = 0x4000062c );
PROVIDE( Cache_Address_Through_Cache = 0x40000630 );
PROVIDE( ROM_Boot_Cache_Init = 0x40000634 );
PROVIDE( MMU_Set_Page_Mode = 0x40000638 );
PROVIDE( MMU_Get_Page_Mode = 0x4000063c );
PROVIDE( Cache_Invalidate_ICache_Items = 0x40000640 );
PROVIDE( Cache_Op_Addr = 0x40000644 );
PROVIDE( Cache_Invalidate_Addr = 0x40000648 );
PROVIDE( Cache_Mask_All = 0x40000650 );
PROVIDE( Cache_UnMask_Dram0 = 0x40000654 );
PROVIDE( Cache_Suspend_ICache_Autoload = 0x40000658 );
PROVIDE( Cache_Resume_ICache_Autoload = 0x4000065c );
PROVIDE( Cache_Start_ICache_Preload = 0x40000660 );
PROVIDE( Cache_ICache_Preload_Done = 0x40000664 );
PROVIDE( Cache_End_ICache_Preload = 0x40000668 );
PROVIDE( Cache_Config_ICache_Autoload = 0x4000066c );
PROVIDE( Cache_Enable_ICache_Autoload = 0x40000670 );
PROVIDE( Cache_Disable_ICache_Autoload = 0x40000674 );
PROVIDE( Cache_Enable_ICache_PreLock = 0x40000678 );
PROVIDE( Cache_Disable_ICache_PreLock = 0x4000067c );
PROVIDE( Cache_Lock_ICache_Items = 0x40000680 );
PROVIDE( Cache_Unlock_ICache_Items = 0x40000684 );
PROVIDE( Cache_Lock_Addr = 0x40000688 );
PROVIDE( Cache_Unlock_Addr = 0x4000068c );
PROVIDE( Cache_Disable_ICache = 0x40000690 );
PROVIDE( Cache_Enable_ICache = 0x40000694 );
PROVIDE( Cache_Freeze_ICache_Enable = 0x400006a0 );
PROVIDE( Cache_Freeze_ICache_Disable = 0x400006a4 );
PROVIDE( Cache_Set_IDROM_MMU_Size = 0x400006a8 );
PROVIDE( Cache_Get_IROM_MMU_End = 0x400006ac );
PROVIDE( Cache_Get_DROM_MMU_End = 0x400006b0 );
PROVIDE( Cache_Travel_Tag_Memory = 0x400006bc );
PROVIDE( Cache_Get_Virtual_Addr = 0x400006c0 );
rom_cache_op_cb = 0x4087ffcc;
rom_cache_internal_table_ptr = 0x4087ffc8;
PROVIDE( ets_clk_get_xtal_freq = 0x400006c4 );
PROVIDE( ets_clk_get_cpu_freq = 0x400006c8 );
PROVIDE( ets_clk_apb_wait_ready = 0x400006cc );
PROVIDE( ets_clk_mspi_apb_wait_ready = 0x400006d0 );
PROVIDE( gpio_input_get = 0x400006d4 );
PROVIDE( gpio_matrix_in = 0x400006d8 );
PROVIDE( gpio_matrix_out = 0x400006dc );
PROVIDE( gpio_output_set = 0x400006e8 );
PROVIDE( gpio_pad_hold = 0x400006ec );
PROVIDE( gpio_pad_input_disable = 0x400006f0 );
PROVIDE( gpio_pad_input_enable = 0x400006f4 );
PROVIDE( gpio_pad_pulldown = 0x400006f8 );
PROVIDE( gpio_pad_pullup = 0x400006fc );
PROVIDE( gpio_pad_select_gpio = 0x40000700 );
PROVIDE( gpio_pad_set_drv = 0x40000704 );
PROVIDE( gpio_pad_unhold = 0x40000708 );
PROVIDE( gpio_pin_wakeup_disable = 0x4000070c );
PROVIDE( gpio_pin_wakeup_enable = 0x40000710 );
PROVIDE( gpio_bypass_matrix_in = 0x40000714 );
PROVIDE( esprv_intc_int_set_priority = 0x40000718 );
PROVIDE( esprv_intc_int_set_threshold = 0x4000071c );
PROVIDE( esprv_intc_int_enable = 0x40000720 );
PROVIDE( esprv_intc_int_disable = 0x40000724 );
PROVIDE( esprv_intc_int_set_type = 0x40000728 );
PROVIDE( intr_handler_set = 0x4000072c );
PROVIDE( intr_matrix_set = 0x40000730 );
PROVIDE( ets_intr_lock = 0x40000734 );
PROVIDE( ets_intr_unlock = 0x40000738 );
PROVIDE( ets_isr_attach = 0x4000073c );
PROVIDE( ets_isr_mask = 0x40000740 );
PROVIDE( ets_isr_unmask = 0x40000744 );
PROVIDE( md5_vector = 0x40000748 );
PROVIDE( MD5Init = 0x4000074c );
PROVIDE( MD5Update = 0x40000750 );
PROVIDE( MD5Final = 0x40000754 );
PROVIDE( crc32_le = 0x40000758 );
PROVIDE( crc16_le = 0x4000075c );
PROVIDE( crc8_le = 0x40000760 );
PROVIDE( crc32_be = 0x40000764 );
PROVIDE( crc16_be = 0x40000768 );
PROVIDE( crc8_be = 0x4000076c );
PROVIDE( esp_crc8 = 0x40000770 );
PROVIDE( ets_sha_enable = 0x40000774 );
PROVIDE( ets_sha_disable = 0x40000778 );
PROVIDE( ets_sha_get_state = 0x4000077c );
PROVIDE( ets_sha_init = 0x40000780 );
PROVIDE( ets_sha_process = 0x40000784 );
PROVIDE( ets_sha_starts = 0x40000788 );
PROVIDE( ets_sha_update = 0x4000078c );
PROVIDE( ets_sha_finish = 0x40000790 );
PROVIDE( ets_sha_clone = 0x40000794 );
PROVIDE( ets_hmac_enable = 0x40000798 );
PROVIDE( ets_hmac_disable = 0x4000079c );
PROVIDE( ets_hmac_calculate_message = 0x400007a0 );
PROVIDE( ets_hmac_calculate_downstream = 0x400007a4 );
PROVIDE( ets_hmac_invalidate_downstream = 0x400007a8 );
PROVIDE( ets_jtag_enable_temporarily = 0x400007ac );
PROVIDE( ets_aes_enable = 0x400007b0 );
PROVIDE( ets_aes_disable = 0x400007b4 );
PROVIDE( ets_aes_setkey = 0x400007b8 );
PROVIDE( ets_aes_block = 0x400007bc );
PROVIDE( ets_aes_setkey_dec = 0x400007c0 );
PROVIDE( ets_aes_setkey_enc = 0x400007c4 );
PROVIDE( ets_bigint_enable = 0x400007c8 );
PROVIDE( ets_bigint_disable = 0x400007cc );
PROVIDE( ets_bigint_multiply = 0x400007d0 );
PROVIDE( ets_bigint_modmult = 0x400007d4 );
PROVIDE( ets_bigint_modexp = 0x400007d8 );
PROVIDE( ets_bigint_wait_finish = 0x400007dc );
PROVIDE( ets_bigint_getz = 0x400007e0 );
PROVIDE( ets_ds_enable = 0x400007e4 );
PROVIDE( ets_ds_disable = 0x400007e8 );
PROVIDE( ets_ds_start_sign = 0x400007ec );
PROVIDE( ets_ds_is_busy = 0x400007f0 );
PROVIDE( ets_ds_finish_sign = 0x400007f4 );
PROVIDE( ets_ds_encrypt_params = 0x400007f8 );
PROVIDE( ets_mgf1_sha256 = 0x400007fc );
PROVIDE( crc32_le_table_ptr = 0x4004fff8 );
PROVIDE( crc16_le_table_ptr = 0x4004fff4 );
PROVIDE( crc8_le_table_ptr = 0x4004fff0 );
PROVIDE( crc32_be_table_ptr = 0x4004ffec );
PROVIDE( crc16_be_table_ptr = 0x4004ffe8 );
PROVIDE( crc8_be_table_ptr = 0x4004ffe4 );
PROVIDE( ets_efuse_read = 0x40000800 );
PROVIDE( ets_efuse_program = 0x40000804 );
PROVIDE( ets_efuse_clear_program_registers = 0x40000808 );
PROVIDE( ets_efuse_write_key = 0x4000080c );
PROVIDE( ets_efuse_get_read_register_address = 0x40000810 );
PROVIDE( ets_efuse_get_key_purpose = 0x40000814 );
PROVIDE( ets_efuse_key_block_unused = 0x40000818 );
PROVIDE( ets_efuse_find_unused_key_block = 0x4000081c );
PROVIDE( ets_efuse_rs_calculate = 0x40000820 );
PROVIDE( ets_efuse_count_unused_key_blocks = 0x40000824 );
PROVIDE( ets_efuse_secure_boot_enabled = 0x40000828 );
PROVIDE( ets_efuse_secure_boot_aggressive_revoke_enabled = 0x4000082c );
PROVIDE( ets_efuse_cache_encryption_enabled = 0x40000830 );
PROVIDE( ets_efuse_download_modes_disabled = 0x40000834 );
PROVIDE( ets_efuse_find_purpose = 0x40000838 );
PROVIDE( ets_efuse_force_send_resume = 0x4000083c );
PROVIDE( ets_efuse_get_flash_delay_us = 0x40000840 );
PROVIDE( ets_efuse_get_mac = 0x40000844 );
PROVIDE( ets_efuse_get_uart_print_control = 0x40000848 );
PROVIDE( ets_efuse_direct_boot_mode_disabled = 0x4000084c );
PROVIDE( ets_efuse_security_download_modes_enabled = 0x40000850 );
PROVIDE( ets_efuse_set_timing = 0x40000854 );
PROVIDE( ets_efuse_jtag_disabled = 0x40000858 );
PROVIDE( ets_efuse_usb_print_is_disabled = 0x4000085c );
PROVIDE( ets_efuse_usb_download_mode_disabled = 0x40000860 );
PROVIDE( ets_efuse_usb_device_disabled = 0x40000864 );
PROVIDE( ets_efuse_secure_boot_fast_wake_enabled = 0x40000868 );
PROVIDE( ets_emsa_pss_verify = 0x4000086c );
PROVIDE( ets_rsa_pss_verify = 0x40000870 );
PROVIDE( ets_secure_boot_verify_bootloader_with_keys = 0x40000874 );
PROVIDE( ets_secure_boot_verify_signature = 0x40000878 );
PROVIDE( ets_secure_boot_read_key_digests = 0x4000087c );
PROVIDE( ets_secure_boot_revoke_public_key_digest = 0x40000880 );
PROVIDE( usb_serial_device_rx_one_char = 0x40000a80 );
PROVIDE( usb_serial_device_rx_one_char_block = 0x40000a84 );
PROVIDE( usb_serial_device_tx_flush = 0x40000a88 );
PROVIDE( usb_serial_device_tx_one_char = 0x40000a8c );
PROVIDE( lldesc_build_chain = 0x40000a90 );
PROVIDE( sip_after_tx_complete = 0x40000a94 );
PROVIDE( sip_alloc_to_host_evt = 0x40000a98 );
PROVIDE( sip_download_begin = 0x40000a9c );
PROVIDE( sip_get_ptr = 0x40000aa0 );
PROVIDE( sip_get_state = 0x40000aa4 );
PROVIDE( sip_init_attach = 0x40000aa8 );
PROVIDE( sip_install_rx_ctrl_cb = 0x40000aac );
PROVIDE( sip_install_rx_data_cb = 0x40000ab0 );
PROVIDE( sip_is_active = 0x40000ab4 );
PROVIDE( sip_post_init = 0x40000ab8 );
PROVIDE( sip_reclaim_from_host_cmd = 0x40000abc );
PROVIDE( sip_reclaim_tx_data_pkt = 0x40000ac0 );
PROVIDE( sip_send = 0x40000ac4 );
PROVIDE( sip_to_host_chain_append = 0x40000ac8 );
PROVIDE( sip_to_host_evt_send_done = 0x40000acc );
PROVIDE( slc_add_credits = 0x40000ad0 );
PROVIDE( slc_enable = 0x40000ad4 );
PROVIDE( slc_from_host_chain_fetch = 0x40000ad8 );
PROVIDE( slc_from_host_chain_recycle = 0x40000adc );
PROVIDE( slc_has_pkt_to_host = 0x40000ae0 );
PROVIDE( slc_init_attach = 0x40000ae4 );
PROVIDE( slc_init_credit = 0x40000ae8 );
PROVIDE( slc_reattach = 0x40000aec );
PROVIDE( slc_send_to_host_chain = 0x40000af0 );
PROVIDE( slc_set_host_io_max_window = 0x40000af4 );
PROVIDE( slc_to_host_chain_recycle = 0x40000af8 );

/* --- esp32c6.rom.libgcc.ld --- */
PROVIDE( __absvdi2 = 0x40000884 );
PROVIDE( __absvsi2 = 0x40000888 );
PROVIDE( __adddf3 = 0x4000088c );
PROVIDE( __addsf3 = 0x40000890 );
PROVIDE( __addvdi3 = 0x40000894 );
PROVIDE( __addvsi3 = 0x40000898 );
PROVIDE( __ashldi3 = 0x4000089c );
PROVIDE( __ashrdi3 = 0x400008a0 );
PROVIDE( __bswapdi2 = 0x400008a4 );
PROVIDE( __bswapsi2 = 0x400008a8 );
PROVIDE( __clear_cache = 0x400008ac );
PROVIDE( __clrsbdi2 = 0x400008b0 );
PROVIDE( __clrsbsi2 = 0x400008b4 );
PROVIDE( __clzdi2 = 0x400008b8 );
PROVIDE( __clzsi2 = 0x400008bc );
PROVIDE( __cmpdi2 = 0x400008c0 );
PROVIDE( __ctzdi2 = 0x400008c4 );
PROVIDE( __ctzsi2 = 0x400008c8 );
PROVIDE( __divdc3 = 0x400008cc );
PROVIDE( __divdf3 = 0x400008d0 );
PROVIDE( __divdi3 = 0x400008d4 );
PROVIDE( __divsc3 = 0x400008d8 );
PROVIDE( __divsf3 = 0x400008dc );
PROVIDE( __divsi3 = 0x400008e0 );
PROVIDE( __eqdf2 = 0x400008e4 );
PROVIDE( __eqsf2 = 0x400008e8 );
PROVIDE( __extendsfdf2 = 0x400008ec );
PROVIDE( __ffsdi2 = 0x400008f0 );
PROVIDE( __ffssi2 = 0x400008f4 );
PROVIDE( __fixdfdi = 0x400008f8 );
PROVIDE( __fixdfsi = 0x400008fc );
PROVIDE( __fixsfdi = 0x40000900 );
PROVIDE( __fixsfsi = 0x40000904 );
PROVIDE( __fixunsdfsi = 0x40000908 );
PROVIDE( __fixunssfdi = 0x4000090c );
PROVIDE( __fixunssfsi = 0x40000910 );
PROVIDE( __floatdidf = 0x40000914 );
PROVIDE( __floatdisf = 0x40000918 );
PROVIDE( __floatsidf = 0x4000091c );
PROVIDE( __floatsisf = 0x40000920 );
PROVIDE( __floatundidf = 0x40000924 );
PROVIDE( __floatundisf = 0x40000928 );
PROVIDE( __floatunsidf = 0x4000092c );
PROVIDE( __floatunsisf = 0x40000930 );
PROVIDE( __gcc_bcmp = 0x40000934 );
PROVIDE( __gedf2 = 0x40000938 );
PROVIDE( __gesf2 = 0x4000093c );
PROVIDE( __gtdf2 = 0x40000940 );
PROVIDE( __gtsf2 = 0x40000944 );
PROVIDE( __ledf2 = 0x40000948 );
PROVIDE( __lesf2 = 0x4000094c );
PROVIDE( __lshrdi3 = 0x40000950 );
PROVIDE( __ltdf2 = 0x40000954 );
PROVIDE( __ltsf2 = 0x40000958 );
PROVIDE( __moddi3 = 0x4000095c );
PROVIDE( __modsi3 = 0x40000960 );
PROVIDE( __muldc3 = 0x40000964 );
PROVIDE( __muldf3 = 0x40000968 );
PROVIDE( __muldi3 = 0x4000096c );
PROVIDE( __mulsc3 = 0x40000970 );
PROVIDE( __mulsf3 = 0x40000974 );
PROVIDE( __mulsi3 = 0x40000978 );
PROVIDE( __mulvdi3 = 0x4000097c );
PROVIDE( __mulvsi3 = 0x40000980 );
PROVIDE( __nedf2 = 0x40000984 );
PROVIDE( __negdf2 = 0x40000988 );
PROVIDE( __negdi2 = 0x4000098c );
PROVIDE( __negsf2 = 0x40000990 );
PROVIDE( __negvdi2 = 0x40000994 );
PROVIDE( __negvsi2 = 0x40000998 );
PROVIDE( __nesf2 = 0x4000099c );
PROVIDE( __paritysi2 = 0x400009a0 );
PROVIDE( __popcountdi2 = 0x400009a4 );
PROVIDE( __popcountsi2 = 0x400009a8 );
PROVIDE( __powidf2 = 0x400009ac );
PROVIDE( __powisf2 = 0x400009b0 );
PROVIDE( __subdf3 = 0x400009b4 );
PROVIDE( __subsf3 = 0x400009b8 );
PROVIDE( __subvdi3 = 0x400009bc );
PROVIDE( __subvsi3 = 0x400009c0 );
PROVIDE( __truncdfsf2 = 0x400009c4 );
PROVIDE( __ucmpdi2 = 0x400009c8 );
PROVIDE( __udivdi3 = 0x400009cc );
PROVIDE( __udivmoddi4 = 0x400009d0 );
PROVIDE( __udivsi3 = 0x400009d4 );
PROVIDE( __udiv_w_sdiv = 0x400009d8 );
PROVIDE( __umoddi3 = 0x400009dc );
PROVIDE( __umodsi3 = 0x400009e0 );
PROVIDE( __unorddf2 = 0x400009e4 );
PROVIDE( __unordsf2 = 0x400009e8 );
PROVIDE( __extenddftf2 = 0x400009ec );
PROVIDE( __trunctfdf2 = 0x400009f0 );

/* --- esp32c6.rom.spiflash.ld --- */
PROVIDE( spi_flash_disable_cache = 0x400001f0 );
PROVIDE( spi_flash_restore_cache = 0x400001f4 );
PROVIDE( spi_flash_cache_enabled = 0x400001f8 );
PROVIDE( esp_enable_cache_flash_wrap = 0x40000200 );
PROVIDE( spi_flash_mmap_os_func_set = 0x40000204 );
PROVIDE( spi_flash_mmap_page_num_init = 0x40000208 );
PROVIDE( spi_flash_mmap = 0x4000020c );
PROVIDE( spi_flash_mmap_pages = 0x40000210 );
PROVIDE( spi_flash_munmap = 0x40000214 );
PROVIDE( spi_flash_mmap_dump = 0x40000218 );
PROVIDE( spi_flash_check_and_flush_cache = 0x4000021c );
PROVIDE( spi_flash_mmap_get_free_pages = 0x40000220 );
PROVIDE( spi_flash_cache2phys = 0x40000224 );
PROVIDE( spi_flash_phys2cache = 0x40000228 );
PROVIDE( esp_flash_chip_driver_initialized = 0x4000022c );
PROVIDE( esp_flash_read_id = 0x40000230 );
PROVIDE( esp_flash_get_size = 0x40000234 );
PROVIDE( esp_flash_erase_chip = 0x40000238 );
PROVIDE( rom_esp_flash_erase_region = 0x4000023c );
PROVIDE( esp_flash_get_chip_write_protect = 0x40000240 );
PROVIDE( esp_flash_set_chip_write_protect = 0x40000244 );
PROVIDE( esp_flash_get_protectable_regions = 0x40000248 );
PROVIDE( esp_flash_get_protected_region = 0x4000024c );
PROVIDE( esp_flash_set_protected_region = 0x40000250 );
PROVIDE( esp_flash_read = 0x40000254 );
PROVIDE( rom_esp_flash_write = 0x40000258 );
PROVIDE( rom_esp_flash_write_encrypted = 0x4000025c );
PROVIDE( esp_flash_read_encrypted = 0x40000260 );
PROVIDE( esp_flash_get_io_mode = 0x40000264 );
PROVIDE( esp_flash_set_io_mode = 0x40000268 );
PROVIDE( spi_flash_boot_attach = 0x4000026c );
PROVIDE( esp_flash_read_chip_id = 0x40000270 );
PROVIDE( detect_spi_flash_chip = 0x40000274 );
PROVIDE( esp_flash_suspend_cmd_init = 0x4000027c );
esp_flash_default_chip = 0x4087ffe8;
esp_flash_api_funcs = 0x4087ffe4;
PROVIDE( spi_flash_chip_generic_probe = 0x40000280 );
PROVIDE( spi_flash_chip_generic_detect_size = 0x40000284 );
PROVIDE( spi_flash_chip_generic_write = 0x40000288 );
PROVIDE( spi_flash_chip_generic_write_encrypted = 0x4000028c );
PROVIDE( spi_flash_chip_generic_set_write_protect = 0x40000290 );
PROVIDE( spi_flash_common_write_status_16b_wrsr = 0x40000294 );
PROVIDE( spi_flash_chip_generic_reset = 0x40000298 );
PROVIDE( spi_flash_chip_generic_erase_chip = 0x4000029c );
PROVIDE( spi_flash_chip_generic_erase_sector = 0x400002a0 );
PROVIDE( spi_flash_chip_generic_erase_block = 0x400002a4 );
PROVIDE( spi_flash_chip_generic_page_program = 0x400002a8 );
PROVIDE( spi_flash_chip_generic_get_write_protect = 0x400002ac );
PROVIDE( spi_flash_common_read_status_16b_rdsr_rdsr2 = 0x400002b0 );
PROVIDE( spi_flash_chip_generic_read_reg = 0x400002b4 );
PROVIDE( spi_flash_chip_generic_yield = 0x400002b8 );
PROVIDE( spi_flash_generic_wait_host_idle = 0x400002bc );
PROVIDE( spi_flash_chip_generic_wait_idle = 0x400002c0 );
PROVIDE( spi_flash_chip_generic_config_host_io_mode = 0x400002c4 );
PROVIDE( spi_flash_chip_generic_read = 0x400002c8 );
PROVIDE( spi_flash_common_read_status_8b_rdsr2 = 0x400002cc );
PROVIDE( spi_flash_chip_generic_get_io_mode = 0x400002d0 );
PROVIDE( spi_flash_common_read_status_8b_rdsr = 0x400002d4 );
PROVIDE( spi_flash_common_write_status_8b_wrsr = 0x400002d8 );
PROVIDE( spi_flash_common_write_status_8b_wrsr2 = 0x400002dc );
PROVIDE( spi_flash_common_set_io_mode = 0x400002e0 );
PROVIDE( spi_flash_chip_generic_set_io_mode = 0x400002e4 );
PROVIDE( spi_flash_chip_generic_read_unique_id = 0x400002e8 );
PROVIDE( spi_flash_chip_generic_get_caps = 0x400002ec );
PROVIDE( spi_flash_chip_generic_suspend_cmd_conf = 0x400002f0 );
PROVIDE( spi_flash_chip_gd_get_io_mode = 0x400002f4 );
PROVIDE( spi_flash_chip_gd_probe = 0x400002f8 );
PROVIDE( spi_flash_chip_gd_set_io_mode = 0x400002fc );
spi_flash_chip_generic_config_data = 0x4087ffe0;
spi_flash_encryption = 0x4087ffdc;
PROVIDE( memspi_host_read_id_hs = 0x40000300 );
PROVIDE( memspi_host_read_status_hs = 0x40000304 );
PROVIDE( memspi_host_flush_cache = 0x40000308 );
PROVIDE( memspi_host_erase_chip = 0x4000030c );
PROVIDE( memspi_host_erase_sector = 0x40000310 );
PROVIDE( memspi_host_erase_block = 0x40000314 );
PROVIDE( memspi_host_program_page = 0x40000318 );
PROVIDE( memspi_host_read = 0x4000031c );
PROVIDE( memspi_host_set_write_protect = 0x40000320 );
PROVIDE( memspi_host_set_max_read_len = 0x40000324 );
PROVIDE( memspi_host_read_data_slicer = 0x40000328 );
PROVIDE( memspi_host_write_data_slicer = 0x4000032c );
PROVIDE( spi_flash_hal_poll_cmd_done = 0x40000330 );
PROVIDE( spi_flash_hal_device_config = 0x40000334 );
PROVIDE( spi_flash_hal_configure_host_io_mode = 0x40000338 );
PROVIDE( spi_flash_hal_common_command = 0x4000033c );
PROVIDE( spi_flash_hal_read = 0x40000340 );
PROVIDE( spi_flash_hal_erase_chip = 0x40000344 );
PROVIDE( spi_flash_hal_erase_sector = 0x40000348 );
PROVIDE( spi_flash_hal_erase_block = 0x4000034c );
PROVIDE( spi_flash_hal_program_page = 0x40000350 );
PROVIDE( spi_flash_hal_set_write_protect = 0x40000354 );
PROVIDE( spi_flash_hal_host_idle = 0x40000358 );
PROVIDE( spi_flash_hal_check_status = 0x4000035c );
PROVIDE( spi_flash_hal_setup_read_suspend = 0x40000360 );
PROVIDE( spi_flash_hal_setup_auto_suspend_mode = 0x40000364 );
PROVIDE( spi_flash_hal_setup_auto_resume_mode = 0x40000368 );
PROVIDE( spi_flash_hal_disable_auto_suspend_mode = 0x4000036c );
PROVIDE( spi_flash_hal_disable_auto_resume_mode = 0x40000370 );
PROVIDE( spi_flash_hal_resume = 0x40000374 );
PROVIDE( spi_flash_hal_suspend = 0x40000378 );
PROVIDE( spi_flash_encryption_hal_enable = 0x4000037c );
PROVIDE( spi_flash_encryption_hal_disable = 0x40000380 );
PROVIDE( spi_flash_encryption_hal_prepare = 0x40000384 );
PROVIDE( spi_flash_encryption_hal_done = 0x40000388 );
PROVIDE( spi_flash_encryption_hal_destroy = 0x4000038c );
PROVIDE( spi_flash_encryption_hal_check = 0x40000390 );

/* --- esp32c6.rom.heap.ld --- */
PROVIDE( tlsf_create = 0x400003fc );
PROVIDE( tlsf_create_with_pool = 0x40000400 );
PROVIDE( tlsf_get_pool = 0x40000404 );
PROVIDE( tlsf_add_pool = 0x40000408 );
PROVIDE( tlsf_remove_pool = 0x4000040c );
PROVIDE( tlsf_malloc = 0x40000410 );
PROVIDE( tlsf_memalign = 0x40000414 );
PROVIDE( tlsf_memalign_offs = 0x40000418 );
PROVIDE( tlsf_realloc = 0x4000041c );
PROVIDE( tlsf_free = 0x40000420 );
PROVIDE( tlsf_block_size = 0x40000424 );
PROVIDE( tlsf_size = 0x40000428 );
PROVIDE( tlsf_align_size = 0x4000042c );
PROVIDE( tlsf_block_size_min = 0x40000430 );
PROVIDE( tlsf_block_size_max = 0x40000434 );
PROVIDE( tlsf_pool_overhead = 0x40000438 );
PROVIDE( tlsf_alloc_overhead = 0x4000043c );
PROVIDE( tlsf_walk_pool = 0x40000440 );
PROVIDE( tlsf_check = 0x40000444 );
PROVIDE( tlsf_poison_fill_pfunc_set = 0x4000044c );
PROVIDE( tlsf_poison_check_pfunc_set = 0x40000450 );
PROVIDE( multi_heap_get_block_address_impl = 0x40000454 );
PROVIDE( multi_heap_get_allocated_size_impl = 0x40000458 );
PROVIDE( multi_heap_register_impl = 0x4000045c );
PROVIDE( multi_heap_set_lock = 0x40000460 );
PROVIDE( multi_heap_mutex_init = 0x40000464 );
PROVIDE( multi_heap_internal_lock = 0x40000468 );
PROVIDE( multi_heap_internal_unlock = 0x4000046c );
PROVIDE( multi_heap_get_first_block = 0x40000470 );
PROVIDE( multi_heap_get_next_block = 0x40000474 );
PROVIDE( multi_heap_is_free = 0x40000478 );
PROVIDE( multi_heap_malloc_impl = 0x4000047c );
PROVIDE( multi_heap_free_impl = 0x40000480 );
PROVIDE( multi_heap_realloc_impl = 0x40000484 );
PROVIDE( multi_heap_aligned_alloc_impl_offs = 0x40000488 );
PROVIDE( multi_heap_aligned_alloc_impl = 0x4000048c );
PROVIDE( multi_heap_check = 0x40000490 );
PROVIDE( multi_heap_dump = 0x40000494 );
PROVIDE( multi_heap_free_size_impl = 0x40000498 );
PROVIDE( multi_heap_minimum_free_size_impl = 0x4000049c );
PROVIDE( multi_heap_get_info_impl = 0x400004a0 );
heap_tlsf_table_ptr = 0x4087ffd8;

/* --- esp32c6.rom.phy.ld --- */
PROVIDE( phy_param_addr = 0x40001104 );
PROVIDE( phy_get_romfuncs = 0x40001108 );
PROVIDE( chip761_phyrom_version = 0x4000110c );
PROVIDE( chip761_phyrom_version_num = 0x40001110 );
PROVIDE( get_rc_dout = 0x40001114 );
PROVIDE( rc_cal = 0x40001118 );
PROVIDE( rom_enter_critical_phy = 0x4000111c );
PROVIDE( rom_exit_critical_phy = 0x40001120 );
PROVIDE( rom_set_chan_cal_interp = 0x40001124 );
PROVIDE( rom_loopback_mode_en = 0x40001128 );
PROVIDE( rom_bb_bss_cbw40 = 0x4000112c );
PROVIDE( abs_temp = 0x40001130 );
PROVIDE( get_data_sat = 0x40001134 );
PROVIDE( phy_byte_to_word = 0x40001138 );
PROVIDE( set_chan_reg = 0x4000113c );
PROVIDE( i2c_master_reset = 0x40001140 );
PROVIDE( rom_set_chan_freq_sw_start = 0x40001144 );
PROVIDE( freq_module_resetn = 0x40001148 );
PROVIDE( freq_chan_en_sw = 0x4000114c );
PROVIDE( write_chan_freq = 0x40001150 );
PROVIDE( get_freq_mem_param = 0x40001154 );
PROVIDE( get_freq_mem_addr = 0x40001158 );
PROVIDE( bt_txpwr_freq = 0x4000115c );
PROVIDE( wr_rf_freq_mem = 0x40001160 );
PROVIDE( read_rf_freq_mem = 0x40001164 );
PROVIDE( freq_i2c_mem_write = 0x40001168 );
PROVIDE( freq_num_get_data = 0x4000116c );
PROVIDE( freq_i2c_num_addr = 0x40001170 );
PROVIDE( freq_i2c_write_set = 0x40001174 );
PROVIDE( pll_dac_mem_update = 0x40001178 );
PROVIDE( pll_cap_mem_update = 0x4000117c );
PROVIDE( get_rf_freq_cap = 0x40001180 );
PROVIDE( get_rf_freq_init = 0x40001184 );
PROVIDE( phy_en_hw_set_freq = 0x40001188 );
PROVIDE( phy_dis_hw_set_freq = 0x4000118c );
PROVIDE( rom_pwdet_sar2_init = 0x40001190 );
PROVIDE( rom_en_pwdet = 0x40001194 );
PROVIDE( rom_get_sar_sig_ref = 0x40001198 );
PROVIDE( rom_pwdet_tone_start = 0x4000119c );
PROVIDE( rom_pwdet_wait_idle = 0x400011a0 );
PROVIDE( rom_read_sar_dout = 0x400011a4 );
PROVIDE( get_tone_sar_dout = 0x400011a8 );
PROVIDE( get_fm_sar_dout = 0x400011ac );
PROVIDE( txtone_linear_pwr = 0x400011b0 );
PROVIDE( linear_to_db = 0x400011b4 );
PROVIDE( get_power_db = 0x400011b8 );
PROVIDE( meas_tone_pwr_db = 0x400011bc );
PROVIDE( pkdet_vol_start = 0x400011c0 );
PROVIDE( read_sar2_code = 0x400011c4 );
PROVIDE( get_sar2_vol = 0x400011c8 );
PROVIDE( get_pll_vol = 0x400011cc );
PROVIDE( tx_pwctrl_bg_init = 0x400011d0 );
PROVIDE( phy_pwdet_always_en = 0x400011d4 );
PROVIDE( phy_pwdet_onetime_en = 0x400011d8 );
PROVIDE( esp_tx_state_out_rom = 0x400011dc );
PROVIDE( ant_dft_cfg_rom = 0x400011e0 );
PROVIDE( ant_wifitx_cfg_rom = 0x400011e4 );
PROVIDE( ant_wifirx_cfg_rom = 0x400011e8 );
PROVIDE( ant_bttx_cfg_rom = 0x400011ec );
PROVIDE( ant_btrx_cfg_rom = 0x400011f0 );
PROVIDE( phy_chan_dump_cfg_rom = 0x400011f4 );
PROVIDE( phy_enable_low_rate = 0x400011f8 );
PROVIDE( phy_disable_low_rate = 0x400011fc );
PROVIDE( phy_is_low_rate_enabled = 0x40001200 );
PROVIDE( phy_dig_reg_backup_rom = 0x40001204 );
PROVIDE( phy_chan_filt_set_rom = 0x40001208 );
PROVIDE( phy_rx11blr_cfg = 0x4000120c );
PROVIDE( set_cca_rom = 0x40001210 );
PROVIDE( set_rx_sense_rom = 0x40001214 );
PROVIDE( rx_gain_force_rom = 0x40001218 );
PROVIDE( rom_rfpll_set_freq = 0x4000121c );
PROVIDE( mhz2ieee = 0x40001220 );
PROVIDE( chan_to_freq = 0x40001224 );
PROVIDE( restart_cal = 0x40001228 );
PROVIDE( write_rfpll_sdm = 0x4000122c );
PROVIDE( wait_rfpll_cal_end = 0x40001230 );
PROVIDE( set_rf_freq_offset = 0x40001234 );
PROVIDE( set_rfpll_freq = 0x40001238 );
PROVIDE( set_channel_rfpll_freq = 0x4000123c );
PROVIDE( rfpll_cap_correct = 0x40001240 );
PROVIDE( rfpll_cap_init_cal = 0x40001244 );
PROVIDE( write_pll_cap = 0x40001248 );
PROVIDE( read_pll_cap = 0x4000124c );
PROVIDE( chip_v7_set_chan_ana = 0x40001250 );
PROVIDE( freq_set_reg = 0x40001254 );
PROVIDE( gen_rx_gain_table = 0x40001258 );
PROVIDE( bt_txdc_cal = 0x4000125c );
PROVIDE( bt_txiq_cal = 0x40001260 );
PROVIDE( txiq_cal_init = 0x40001264 );
PROVIDE( txdc_cal_init = 0x40001268 );
PROVIDE( txdc_cal = 0x4000126c );
PROVIDE( txiq_get_mis_pwr = 0x40001270 );
PROVIDE( txiq_cover = 0x40001274 );
PROVIDE( rfcal_txiq = 0x40001278 );
PROVIDE( get_power_atten = 0x4000127c );
PROVIDE( pwdet_ref_code = 0x40001280 );
PROVIDE( pwdet_code_cal = 0x40001284 );
PROVIDE( rfcal_txcap = 0x40001288 );
PROVIDE( tx_cap_init = 0x4000128c );
PROVIDE( rfcal_pwrctrl = 0x40001290 );
PROVIDE( tx_pwctrl_init_cal = 0x40001294 );
PROVIDE( tx_pwctrl_init = 0x40001298 );
PROVIDE( bt_tx_pwctrl_init = 0x4000129c );
PROVIDE( rom_i2c_enter_critical = 0x400012a0 );
PROVIDE( rom_i2c_exit_critical = 0x400012a4 );
PROVIDE( rom_get_i2c_read_mask = 0x400012a8 );
PROVIDE( rom_get_i2c_mst0_mask = 0x400012ac );
PROVIDE( rom_get_i2c_hostid = 0x400012b0 );
PROVIDE( rom_chip_i2c_readReg_org = 0x400012b4 );
PROVIDE( rom_chip_i2c_readReg = 0x400012b8 );
PROVIDE( rom_chip_i2c_writeReg = 0x400012c0 );
PROVIDE( rom_set_txcap_reg = 0x400012d0 );
PROVIDE( i2c_paral_set_mst0 = 0x400012d4 );
PROVIDE( i2c_paral_set_read = 0x400012d8 );
PROVIDE( i2c_paral_read = 0x400012dc );
PROVIDE( i2c_paral_write = 0x400012e0 );
PROVIDE( i2c_paral_write_num = 0x400012e4 );
PROVIDE( i2c_paral_write_mask = 0x400012e8 );
PROVIDE( i2c_sar2_init_code = 0x400012ec );
PROVIDE( rom_pbus_force_mode = 0x400012f0 );
PROVIDE( rom_pbus_rd_addr = 0x400012f4 );
PROVIDE( rom_pbus_rd_shift = 0x400012f8 );
PROVIDE( rom_pbus_force_test = 0x400012fc );
PROVIDE( rom_pbus_rd = 0x40001300 );
PROVIDE( rom_pbus_set_rxgain = 0x40001304 );
PROVIDE( rom_pbus_xpd_rx_off = 0x40001308 );
PROVIDE( rom_pbus_xpd_rx_on = 0x4000130c );
PROVIDE( rom_pbus_xpd_tx_off = 0x40001310 );
PROVIDE( rom_pbus_xpd_tx_on = 0x40001314 );
PROVIDE( rom_set_loopback_gain = 0x40001318 );
PROVIDE( rom_txcal_debuge_mode = 0x4000131c );
PROVIDE( pbus_debugmode = 0x40001320 );
PROVIDE( pbus_workmode = 0x40001324 );
PROVIDE( pbus_set_dco = 0x40001328 );
PROVIDE( txcal_work_mode = 0x4000132c );
PROVIDE( rom_start_tx_tone_step = 0x40001330 );
PROVIDE( rom_stop_tx_tone = 0x40001334 );
PROVIDE( disable_agc = 0x40001338 );
PROVIDE( enable_agc = 0x4000133c );
PROVIDE( phy_disable_cca = 0x40001340 );
PROVIDE( phy_enable_cca = 0x40001344 );
PROVIDE( write_gain_mem = 0x40001348 );
PROVIDE( bb_bss_cbw40_dig = 0x4000134c );
PROVIDE( cbw2040_cfg = 0x40001350 );
PROVIDE( mac_tx_chan_offset = 0x40001354 );
PROVIDE( tx_paon_set = 0x40001358 );
PROVIDE( pwdet_reg_init = 0x4000135c );
PROVIDE( i2cmst_reg_init = 0x40001360 );
PROVIDE( bt_gain_offset = 0x40001364 );
PROVIDE( fe_reg_init = 0x40001368 );
PROVIDE( mac_enable_bb = 0x4000136c );
PROVIDE( bb_wdg_cfg = 0x40001370 );
PROVIDE( fe_txrx_reset = 0x40001374 );
PROVIDE( set_rx_comp = 0x40001378 );
PROVIDE( agc_reg_init = 0x4000137c );
PROVIDE( bb_reg_init = 0x40001380 );
PROVIDE( open_i2c_xpd = 0x40001384 );
PROVIDE( txiq_set_reg = 0x40001388 );
PROVIDE( rxiq_set_reg = 0x4000138c );
PROVIDE( set_txclk_en = 0x40001390 );
PROVIDE( set_rxclk_en = 0x40001394 );
PROVIDE( bb_wdg_test_en = 0x40001398 );
PROVIDE( noise_floor_auto_set = 0x4000139c );
PROVIDE( read_hw_noisefloor = 0x400013a0 );
PROVIDE( iq_corr_enable = 0x400013a4 );
PROVIDE( wifi_agc_sat_gain = 0x400013a8 );
PROVIDE( phy_bbpll_cal = 0x400013ac );
PROVIDE( phy_ant_init = 0x400013b0 );
PROVIDE( phy_set_bbfreq_init = 0x400013b4 );
PROVIDE( wifi_fbw_sel = 0x400013b8 );
PROVIDE( bt_filter_reg = 0x400013bc );
PROVIDE( phy_rx_sense_set = 0x400013c0 );
PROVIDE( tx_state_set = 0x400013c4 );
PROVIDE( phy_close_pa = 0x400013c8 );
PROVIDE( phy_freq_correct = 0x400013cc );
PROVIDE( set_pbus_reg = 0x400013d0 );
PROVIDE( wifi_rifs_mode_en = 0x400013d4 );
PROVIDE( nrx_freq_set = 0x400013d8 );
PROVIDE( fe_adc_on = 0x400013dc );
PROVIDE( phy_force_pwr_index = 0x400013e0 );
PROVIDE( rom_iq_est_enable = 0x400013e4 );
PROVIDE( rom_iq_est_disable = 0x400013e8 );
PROVIDE( rom_bb_gain_index = 0x400013ec );
PROVIDE( rom_rfrx_gain_index = 0x400013f0 );
PROVIDE( dc_iq_est = 0x400013f4 );
PROVIDE( set_cal_rxdc = 0x400013f8 );
PROVIDE( rxiq_get_mis = 0x400013fc );
PROVIDE( rxiq_cover_mg_mp = 0x40001400 );
PROVIDE( rfcal_rxiq = 0x40001404 );
PROVIDE( get_rfcal_rxiq_data = 0x40001408 );
PROVIDE( get_dco_comp = 0x4000140c );
PROVIDE( pbus_rx_dco_cal = 0x40001410 );
PROVIDE( rxdc_est_min = 0x40001414 );
PROVIDE( pbus_rx_dco_cal_1step = 0x40001418 );
PROVIDE( set_lb_txiq = 0x4000141c );
PROVIDE( set_rx_gain_cal_iq = 0x40001420 );
PROVIDE( set_rx_gain_cal_dc = 0x40001424 );
PROVIDE( spur_reg_write_one_tone = 0x40001428 );
PROVIDE( spur_cal = 0x4000142c );
PROVIDE( spur_coef_cfg = 0x40001430 );
PROVIDE( tsens_power_up = 0x40001434 );
PROVIDE( tsens_read_init = 0x40001438 );
PROVIDE( code_to_temp = 0x4000143c );
PROVIDE( tsens_index_to_dac = 0x40001440 );
PROVIDE( tsens_index_to_offset = 0x40001444 );
PROVIDE( tsens_dac_cal = 0x40001448 );
PROVIDE( tsens_code_read = 0x4000144c );
PROVIDE( tsens_temp_read = 0x40001450 );
PROVIDE( temp_to_power = 0x40001454 );
PROVIDE( get_temp_init = 0x40001458 );
PROVIDE( txbbgain_to_index = 0x4000145c );
PROVIDE( index_to_txbbgain = 0x40001460 );
PROVIDE( bt_index_to_bb = 0x40001464 );
PROVIDE( bt_bb_to_index = 0x40001468 );
PROVIDE( bt_get_tx_gain = 0x4000146c );
PROVIDE( dig_gain_check = 0x40001470 );
PROVIDE( wifi_get_tx_gain = 0x40001474 );
PROVIDE( wifi_11g_rate_chg = 0x40001478 );
PROVIDE( bt_chan_pwr_interp = 0x4000147c );
PROVIDE( get_rate_fcc_index = 0x40001480 );
PROVIDE( get_chan_target_power = 0x40001484 );
PROVIDE( get_tx_gain_value = 0x40001488 );
PROVIDE( wifi_get_target_power = 0x4000148c );

/* --- esp32c6.rom.pp.ld --- */
PROVIDE( esp_pp_rom_version_get = 0x40000bd8 );
PROVIDE( ppCalTxopRTSThreshold = 0x40000bdc );
PROVIDE( RC_GetBlockAckTime = 0x40000be0 );
PROVIDE( ebuf_list_remove = 0x40000be4 );
PROVIDE( GetAccess = 0x40000bf4 );
PROVIDE( hal_mac_is_low_rate_enabled = 0x40000bf8 );
PROVIDE( hal_mac_tx_get_blockack = 0x40000bfc );
PROVIDE( ic_get_trc = 0x40000c04 );
PROVIDE( ic_interface_enabled = 0x40000c10 );
PROVIDE( is_lmac_idle = 0x40000c14 );
PROVIDE( lmacDiscardAgedMSDU = 0x40000c1c );
PROVIDE( lmacIsIdle = 0x40000c28 );
PROVIDE( lmacIsLongFrame = 0x40000c2c );
PROVIDE( lmacPostTxComplete = 0x40000c34 );
PROVIDE( lmacProcessAllTxTimeout = 0x40000c38 );
PROVIDE( lmacProcessCollisions = 0x40000c3c );
PROVIDE( lmacReachLongLimit = 0x40000c44 );
PROVIDE( lmacReachShortLimit = 0x40000c48 );
PROVIDE( lmacRecycleMPDU = 0x40000c4c );
PROVIDE( lmacRxDone = 0x40000c50 );
PROVIDE( mac_tx_set_duration = 0x40000c60 );
PROVIDE( mac_tx_set_plcp2 = 0x40000c6c );
PROVIDE( pm_disable_sleep_delay_timer = 0x40000c78 );
PROVIDE( pm_mac_wakeup = 0x40000c80 );
PROVIDE( pm_mac_sleep = 0x40000c84 );
PROVIDE( pm_enable_sleep_delay_timer = 0x40000c8c );
PROVIDE( pm_local_tsf_process = 0x40000c90 );
PROVIDE( pm_is_waked = 0x40000c9c );
PROVIDE( pm_on_data_rx = 0x40000ca8 );
PROVIDE( pm_sleep_for = 0x40000cc4 );
PROVIDE( ppAMPDU2Normal = 0x40000ccc );
PROVIDE( ppCalFrameTimes = 0x40000cd4 );
PROVIDE( ppCalSubFrameLength = 0x40000cd8 );
PROVIDE( ppCheckTxAMPDUlength = 0x40000ce0 );
PROVIDE( ppDequeueRxq_Locked = 0x40000ce4 );
PROVIDE( ppDequeueTxQ = 0x40000ce8 );
PROVIDE( ppEmptyDelimiterLength = 0x40000cec );
PROVIDE( ppEnqueueRxq = 0x40000cf0 );
PROVIDE( ppEnqueueTxDone = 0x40000cf4 );
PROVIDE( ppGetTxframe = 0x40000cf8 );
PROVIDE( ppProcessRxPktHdr = 0x40000d04 );
PROVIDE( ppRecordBarRRC = 0x40000d0c );
PROVIDE( ppRecycleAmpdu = 0x40000d10 );
PROVIDE( ppRecycleRxPkt = 0x40000d14 );
PROVIDE( ppResumeTxAMPDU = 0x40000d1c );
PROVIDE( ppSearchTxQueue = 0x40000d2c );
PROVIDE( ppSearchTxframe = 0x40000d30 );
PROVIDE( ppSelectNextQueue = 0x40000d34 );
PROVIDE( ppSubFromAMPDU = 0x40000d38 );
PROVIDE( ppTxProtoProc = 0x40000d44 );
PROVIDE( ppTxqUpdateBitmap = 0x40000d48 );
PROVIDE( pp_hdrsize = 0x40000d50 );
PROVIDE( pp_post = 0x40000d54 );
PROVIDE( pp_process_hmac_waiting_txq = 0x40000d58 );
PROVIDE( rcGetAmpduSched = 0x40000d5c );
PROVIDE( rcUpdateRxDone = 0x40000d60 );
PROVIDE( rc_get_trc = 0x40000d64 );
PROVIDE( rc_get_trc_by_index = 0x40000d68 );
PROVIDE( rcAmpduLowerRate = 0x40000d6c );
PROVIDE( rcampduuprate = 0x40000d70 );
PROVIDE( rcClearCurAMPDUSched = 0x40000d74 );
PROVIDE( rcClearCurSched = 0x40000d78 );
PROVIDE( rcClearCurStat = 0x40000d7c );
PROVIDE( rcLowerSched = 0x40000d84 );
PROVIDE( rcSetTxAmpduLimit = 0x40000d88 );
PROVIDE( rcTxUpdatePer = 0x40000d8c );
PROVIDE( rcUpdateAckSnr = 0x40000d90 );
PROVIDE( rcUpSched = 0x40000da0 );
PROVIDE( rssi_margin = 0x40000da4 );
PROVIDE( rx11NRate2AMPDULimit = 0x40000da8 );
PROVIDE( TRC_AMPDU_PER_DOWN_THRESHOLD = 0x40000dac );
PROVIDE( TRC_AMPDU_PER_UP_THRESHOLD = 0x40000db0 );
PROVIDE( trc_calc_duration = 0x40000db4 );
PROVIDE( trc_isTxAmpduOperational = 0x40000db8 );
PROVIDE( trc_onAmpduOp = 0x40000dbc );
PROVIDE( TRC_PER_IS_GOOD = 0x40000dc0 );
PROVIDE( trc_SetTxAmpduState = 0x40000dc4 );
PROVIDE( trc_tid_isTxAmpduOperational = 0x40000dc8 );
PROVIDE( trcAmpduSetState = 0x40000dcc );
PROVIDE( wDev_DiscardFrame = 0x40000dd8 );
PROVIDE( wDev_GetNoiseFloor = 0x40000ddc );
PROVIDE( wDev_IndicateAmpdu = 0x40000de0 );
PROVIDE( wdev_mac_reg_load = 0x40000de8 );
PROVIDE( wdev_mac_reg_store = 0x40000dec );
PROVIDE( wdev_mac_special_reg_load = 0x40000df0 );
PROVIDE( wdev_mac_special_reg_store = 0x40000df4 );
PROVIDE( wdev_mac_wakeup = 0x40000df8 );
PROVIDE( wdev_mac_sleep = 0x40000dfc );
PROVIDE( hal_mac_is_dma_enable = 0x40000e00 );
PROVIDE( wdev_csi_len_align = 0x40000e10 );
PROVIDE( ppDequeueTxDone_Locked = 0x40000e14 );
PROVIDE( config_is_cache_tx_buf_enabled = 0x40000e20 );
PROVIDE( ppProcessWaitingQueue = 0x40000e28 );
PROVIDE( ppDisableQueue = 0x40000e2c );
PROVIDE( pm_allow_tx = 0x40000e30 );
PROVIDE( ppProcTxCallback = 0x40000e38 );
PROVIDE( ppCalPreFecPaddingFactor = 0x40000e40 );
PROVIDE( hal_get_tsf_timer = 0x40000e4c );
PROVIDE( ppTxPktForceWaked = 0x40000e50 );
PROVIDE( lmacProcessLongFrameSuccess = 0x40000e54 );
PROVIDE( lmacProcessShortFrameSuccess = 0x40000e58 );
PROVIDE( lmacProcessTBSuccess = 0x40000e60 );
PROVIDE( lmacProcessAckTimeout = 0x40000e68 );
PROVIDE( get_estimated_batime = 0x40000e74 );
PROVIDE( is_use_muedca = 0x40000e78 );
PROVIDE( hal_mac_clr_txq_state = 0x40000e84 );
PROVIDE( hal_mac_get_txq_complete = 0x40000e88 );
PROVIDE( ht_get_min_subframe_len = 0x40000e8c );
PROVIDE( rx11ACRate2AMPDULimit = 0x40000e90 );
PROVIDE( pwr_hal_clear_intr_status = 0x40000e94 );
PROVIDE( pwr_hal_clear_mac_modem_beacon_miss_intr_filter = 0x40000e98 );
PROVIDE( pwr_hal_clear_mac_modem_rx_beacon_info = 0x40000e9c );
PROVIDE( pwr_hal_clear_mac_modem_rx_beacon_miss_counter = 0x40000ea0 );
PROVIDE( pwr_hal_clear_mac_modem_rx_beacon_sleep_counter = 0x40000ea4 );
PROVIDE( pwr_hal_clear_mac_modem_state_wakeup_protect_signal = 0x40000ea8 );
PROVIDE( pwr_hal_get_intr_raw_signal = 0x40000eac );
PROVIDE( pwr_hal_get_intr_status = 0x40000eb0 );
PROVIDE( pwr_hal_get_mac_modem_beacon_miss_limit_exceeded_status = 0x40000eb4 );
PROVIDE( pwr_hal_get_mac_modem_rx_beacon_location_state = 0x40000eb8 );
PROVIDE( pwr_hal_get_mac_modem_rx_beacon_valid_state = 0x40000ebc );
PROVIDE( pwr_hal_get_mac_modem_state_sleep_limit_exceeded_status = 0x40000ec0 );
PROVIDE( pwr_hal_set_beacon_filter_abort_disable = 0x40000ec4 );
PROVIDE( pwr_hal_set_beacon_filter_abort_enable = 0x40000ec8 );
PROVIDE( pwr_hal_set_beacon_filter_abort_length = 0x40000ecc );
PROVIDE( pwr_hal_set_beacon_filter_disable = 0x40000ed8 );
PROVIDE( pwr_hal_set_beacon_filter_enable = 0x40000edc );
PROVIDE( pwr_hal_set_beacon_filter_force_dump_disable = 0x40000ee0 );
PROVIDE( pwr_hal_set_beacon_filter_force_dump_enable = 0x40000ee4 );
PROVIDE( pwr_hal_set_beacon_filter_force_dump_limit = 0x40000ee8 );
PROVIDE( pwr_hal_set_beacon_filter_force_sync_disable = 0x40000eec );
PROVIDE( pwr_hal_set_beacon_filter_force_sync_enable = 0x40000ef0 );
PROVIDE( pwr_hal_set_beacon_filter_force_sync_limit = 0x40000ef4 );
PROVIDE( pwr_hal_set_beacon_filter_frame_crc_state = 0x40000ef8 );
PROVIDE( pwr_hal_set_beacon_filter_soc_wakeup_and_intr_disable = 0x40000efc );
PROVIDE( pwr_hal_set_beacon_filter_soc_wakeup_and_intr_enable = 0x40000f00 );
PROVIDE( pwr_hal_set_beacon_filter_unicast_wakeup_disable = 0x40000f04 );
PROVIDE( pwr_hal_set_beacon_filter_unicast_wakeup_enable = 0x40000f08 );
PROVIDE( pwr_hal_set_lpclk_cycle_time = 0x40000f0c );
PROVIDE( pwr_hal_set_lpclk_sync_disable = 0x40000f10 );
PROVIDE( pwr_hal_set_lpclk_sync_enable = 0x40000f14 );
PROVIDE( pwr_hal_set_mac_modem_beacon_miss_intr_disable = 0x40000f18 );
PROVIDE( pwr_hal_set_mac_modem_beacon_miss_intr_enable = 0x40000f1c );
PROVIDE( pwr_hal_set_mac_modem_beacon_miss_limit = 0x40000f20 );
PROVIDE( pwr_hal_set_mac_modem_beacon_miss_limit_exceeded_wakeup_disable = 0x40000f24 );
PROVIDE( pwr_hal_set_mac_modem_beacon_miss_limit_exceeded_wakeup_enable = 0x40000f28 );
PROVIDE( pwr_hal_set_mac_modem_beacon_miss_timeout = 0x40000f2c );
PROVIDE( pwr_hal_set_mac_modem_state_sleep_limit = 0x40000f30 );
PROVIDE( pwr_hal_set_mac_modem_state_sleep_limit_exceeded_wakeup_disable = 0x40000f34 );
PROVIDE( pwr_hal_set_mac_modem_state_sleep_limit_exceeded_wakeup_enable = 0x40000f38 );
PROVIDE( pwr_hal_set_mac_modem_state_wakeup_protect_disable = 0x40000f3c );
PROVIDE( pwr_hal_set_mac_modem_state_wakeup_protect_early_time = 0x40000f40 );
PROVIDE( pwr_hal_set_mac_modem_state_wakeup_protect_enable = 0x40000f44 );
PROVIDE( pwr_hal_set_mac_modem_tbtt_auto_period_disable = 0x40000f48 );
PROVIDE( pwr_hal_set_mac_modem_tbtt_auto_period_enable = 0x40000f4c );
PROVIDE( pwr_hal_set_mac_modem_tbtt_auto_period_interval = 0x40000f50 );
PROVIDE( pwr_hal_set_modem_state_interface = 0x40000f54 );
PROVIDE( hal_tsf_clear_soc_wakeup_request = 0x40000f58 );
PROVIDE( tsf_hal_clear_mac_modem_rf_power_state = 0x40000f5c );
PROVIDE( tsf_hal_clear_soc_wakeup_request = 0x40000f60 );
PROVIDE( tsf_hal_get_counter_value = 0x40000f64 );
PROVIDE( tsf_hal_get_mac_modem_rf_power_state = 0x40000f68 );
PROVIDE( tsf_hal_get_tbtt_interval = 0x40000f6c );
PROVIDE( tsf_hal_get_time = 0x40000f70 );
PROVIDE( tsf_hal_get_timer_target = 0x40000f74 );
PROVIDE( tsf_hal_is_tsf_enabled = 0x40000f78 );
PROVIDE( tsf_hal_map_tbtt_target_to_rx_frame = 0x40000f7c );
PROVIDE( tsf_hal_map_tsf_to_bssid = 0x40000f80 );
PROVIDE( tsf_hal_set_counter_value = 0x40000f84 );
PROVIDE( tsf_hal_set_modem_wakeup_early_time = 0x40000f88 );
PROVIDE( tsf_hal_set_rx_beacon_abort_tsf_time_deviation_sync_disable = 0x40000f8c );
PROVIDE( tsf_hal_set_rx_beacon_abort_tsf_time_deviation_sync_enable = 0x40000f90 );
PROVIDE( tsf_hal_set_rx_beacon_fail_tsf_time_deviation_sync_disable = 0x40000f94 );
PROVIDE( tsf_hal_set_rx_beacon_fail_tsf_time_deviation_sync_enable = 0x40000f98 );
PROVIDE( tsf_hal_set_rx_beacon_success_tsf_time_deviation_sync_disable = 0x40000f9c );
PROVIDE( tsf_hal_set_rx_beacon_success_tsf_time_deviation_sync_enable = 0x40000fa0 );
PROVIDE( tsf_hal_set_tbtt_disable = 0x40000fa4 );
PROVIDE( tsf_hal_set_tbtt_early_time = 0x40000fa8 );
PROVIDE( tsf_hal_set_tbtt_enable = 0x40000fac );
PROVIDE( tsf_hal_set_tbtt_interval = 0x40000fb0 );
PROVIDE( tsf_hal_set_tbtt_intr_disable = 0x40000fb4 );
PROVIDE( tsf_hal_set_tbtt_intr_enable = 0x40000fb8 );
PROVIDE( tsf_hal_set_tbtt_modem_wakeup_disable = 0x40000fbc );
PROVIDE( tsf_hal_set_tbtt_modem_wakeup_enable = 0x40000fc0 );
PROVIDE( tsf_hal_set_tbtt_rf_ctrl_disable = 0x40000fc4 );
PROVIDE( tsf_hal_set_tbtt_rf_ctrl_enable = 0x40000fc8 );
PROVIDE( tsf_hal_set_tbtt_rf_ctrl_wait_cycles = 0x40000fcc );
PROVIDE( tsf_hal_set_tbtt_soc_wakeup_disable = 0x40000fd0 );
PROVIDE( tsf_hal_set_tbtt_soc_wakeup_enable = 0x40000fd4 );
PROVIDE( tsf_hal_set_time = 0x40000fdc );
PROVIDE( tsf_hal_set_timer_disable = 0x40000fe0 );
PROVIDE( tsf_hal_set_timer_enable = 0x40000fe4 );
PROVIDE( tsf_hal_set_timer_intr_disable = 0x40000fe8 );
PROVIDE( tsf_hal_set_timer_intr_enable = 0x40000fec );
PROVIDE( tsf_hal_set_timer_modem_wakeup_disable = 0x40000ff0 );
PROVIDE( tsf_hal_set_timer_modem_wakeup_enable = 0x40000ff4 );
PROVIDE( tsf_hal_set_timer_rf_ctrl_disable = 0x40000ff8 );
PROVIDE( tsf_hal_set_timer_rf_ctrl_enable = 0x40000ffc );
PROVIDE( tsf_hal_set_timer_rf_ctrl_wait_cycles = 0x40001000 );
PROVIDE( tsf_hal_set_timer_soc_wakeup_disable = 0x40001004 );
PROVIDE( tsf_hal_set_timer_soc_wakeup_enable = 0x40001008 );
PROVIDE( tsf_hal_set_timer_target = 0x4000100c );
PROVIDE( tsf_hal_set_tsf_disable = 0x40001010 );
PROVIDE( tsf_hal_set_tsf_enable = 0x40001014 );
PROVIDE( tsf_hal_set_tsf_time_deviation = 0x40001018 );
PROVIDE( tsf_hal_set_tsf_time_deviation_sync_disable = 0x4000101c );
PROVIDE( tsf_hal_set_tsf_time_deviation_sync_enable = 0x40001020 );
PROVIDE( tsf_hal_unmap_tbtt_target_to_rx_frame = 0x40001024 );
PROVIDE( rcGetDCMMaxRate = 0x40001040 );
PROVIDE( ppDirectRecycleAmpdu = 0x40001048 );
PROVIDE( ppAdd2AMPDUTail = 0x4000105c );
PROVIDE( esp_test_disable_tx_statistics = 0x40001060 );
PROVIDE( esp_test_enable_tx_statistics = 0x40001064 );
PROVIDE( esp_test_clr_tx_statistics = 0x40001068 );
PROVIDE( esp_test_get_tx_statistics = 0x4000106c );
PROVIDE( esp_test_clr_tx_tb_statistics = 0x40001070 );
PROVIDE( esp_test_get_tx_tb_statistics = 0x40001074 );
PROVIDE( test_tx_fail_statistics = 0x40001078 );
PROVIDE( esp_test_tx_enab_statistics = 0x40001088 );
PROVIDE( esp_test_tx_tb_complete = 0x4000108c );
PROVIDE( esp_test_tx_count_retry = 0x40001090 );
PROVIDE( esp_test_tx_count_collision = 0x40001094 );
PROVIDE( esp_test_tx_count_timeout = 0x40001098 );
PROVIDE( hal_enable_tx_statistics = 0x4000109c );
PROVIDE( test_rx_process_complete_noeb = 0x400010a0 );
PROVIDE( test_rx_process_complete_retry = 0x400010a4 );
PROVIDE( esp_test_rx_process_complete = 0x400010a8 );
PROVIDE( esp_test_clr_rx_statistics = 0x400010ac );
PROVIDE( esp_test_get_rx_statistics = 0x400010b0 );
PROVIDE( test_free_rx_statistics = 0x400010b4 );
PROVIDE( esp_test_set_rx_error_occurs = 0x400010b8 );
PROVIDE( esp_test_get_rx_error_occurs = 0x400010bc );
PROVIDE( esp_test_clr_rx_error_occurs = 0x400010c0 );
PROVIDE( esp_test_disable_rx_statistics = 0x400010c4 );
PROVIDE( esp_test_enable_rx_statistics = 0x400010c8 );
PROVIDE( hal_enable_rx_statistics = 0x400010cc );
PROVIDE( get_user_num = 0x400010d0 );
PROVIDE( mumimo_spatial_cfg_get_nsts = 0x400010d4 );
PROVIDE( mumimo_spatial_cfg_get_nsts_tot = 0x400010d8 );
PROVIDE( test_mumimo_get_heltf_num = 0x400010dc );
PROVIDE( test_mimo_update_user_info = 0x400010e0 );
PROVIDE( test_parse_rx_mu_mimo = 0x400010e4 );
PROVIDE( test_nonmimo_update_user_info = 0x400010e8 );
PROVIDE( test_parse_rx_mu_nonmimo = 0x400010ec );
PROVIDE( esp_test_rx_parse_mu = 0x400010f0 );
PROVIDE( esp_test_get_rx_mu_statistics = 0x400010f4 );
PROVIDE( esp_test_clr_rx_mu_statistics = 0x400010f8 );
PROVIDE( esp_test_enable_rx_mu_statistics = 0x400010fc );
PROVIDE( esp_test_disable_rx_mu_statistics = 0x40001100 );
PROVIDE( our_instances_ptr = 0x4004ffe0 );
pTxRx = 0x4087ff80;
lmacConfMib_ptr = 0x4087ff7c;
our_wait_eb = 0x4087ff78;
our_tx_eb = 0x4087ff74;
pp_wdev_funcs = 0x4087ff70;
g_osi_funcs_p = 0x4087ff6c;
wDevCtrl_ptr = 0x4087ff68;
PROVIDE( g_wdev_last_desc_reset_ptr = 0x4004ffdc );
wDevMacSleep_ptr = 0x4087ff64;
g_lmac_cnt_ptr = 0x4087ff60;
PROVIDE( our_controls_ptr = 0x4004ffd8 );
pp_sig_cnt_ptr = 0x4087ff5c;
g_eb_list_desc_ptr = 0x4087ff58;
s_fragment_ptr = 0x4087ff54;
if_ctrl_ptr = 0x4087ff50;
g_intr_lock_mux = 0x4087ff4c;
g_wifi_global_lock = 0x4087ff48;
s_wifi_queue = 0x4087ff44;
pp_task_hdl = 0x4087ff40;
s_pp_task_create_sem = 0x4087ff3c;
s_pp_task_del_sem = 0x4087ff38;
g_wifi_menuconfig_ptr = 0x4087ff34;
xphyQueue = 0x4087ff30;
ap_no_lr_ptr = 0x4087ff2c;
rc11BSchedTbl_ptr = 0x4087ff28;
rc11NSchedTbl_ptr = 0x4087ff24;
rcLoRaSchedTbl_ptr = 0x4087ff20;
BasicOFDMSched_ptr = 0x4087ff1c;
trc_ctl_ptr = 0x4087ff18;
g_pm_cnt_ptr = 0x4087ff14;
g_pm_ptr = 0x4087ff10;
g_pm_cfg_ptr = 0x4087ff0c;
g_esp_mesh_quick_funcs_ptr = 0x4087ff08;
g_txop_queue_status_ptr = 0x4087ff04;
g_mac_sleep_en_ptr = 0x4087ff00;
g_mesh_is_root_ptr = 0x4087fefc;
g_mesh_topology_ptr = 0x4087fef8;
g_mesh_init_ps_type_ptr = 0x4087fef4;
g_mesh_is_started_ptr = 0x4087fef0;
g_config_func = 0x4087feec;
g_net80211_tx_func = 0x4087fee8;
g_timer_func = 0x4087fee4;
s_michael_mic_failure_cb = 0x4087fee0;
wifi_sta_rx_probe_req = 0x4087fedc;
g_tx_done_cb_func = 0x4087fed8;
g_per_conn_trc = 0x4087fe8c;
s_encap_amsdu_func = 0x4087fe88;
rx_beacon_count = 0x4087fe84;
rx_beacon_sw_parse = 0x4087fe80;
rx_beacon_hw_parse = 0x4087fe7c;
rx_beacon_tim_count = 0x4087fe78;
rx_beacon_tim_udata = 0x4087fe74;
rx_beacon_tim_udata_bitmap = 0x4087fe70;
rx_beacon_tim_bdata = 0x4087fe6c;
rx_beacon_tim_bdata_bitmapctl = 0x4087fe68;
rx_beacon_tim_bdata_bitmap_trans = 0x4087fe64;
rx_beacon_tim_bdata_bitmap_mbssid_self = 0x4087fe60;
rx_beacon_tim_bdata_bitmap_mbssid_other = 0x4087fe5c;
rx_beacon_dtim_tim = 0x4087fe58;
rx_beacon_dtim_tim_mcast = 0x4087fe54;
amdpu_delay_time_ms = 0x4087fd08;
ampdu_delay_packet = 0x4087fd04;
ampdu_delay = 0x4087fe51;
first_ampdu = 0x4087fe50;
s_ht_ampdu_density_us = 0x4087fd02;
s_ht_ampdu_density = 0x4087fd01;
s_running_phy_type = 0x4087fd00;
complete_ena_tb_seqno = 0x4087fe4c;
complete_ena_tb_final = 0x4087fe48;
complete_ena_tb_count = 0x4087fe44;
s_itwt_state = 0x4087fe40;
g_dbg_interp_tsf = 0x4087fe3c;
g_dbg_interp_tsf_end = 0x4087fe38;
g_dbg_closrf_tsf = 0x4087fe34;
g_dbg_closrf_idx = 0x4087fe30;
g_dbg_closrf_blk = 0x4087fe2c;
s_he_min_len_bytes = 0x4087fdf0;
s_he_dcm_min_len_bytes = 0x4087fdd0;
s_mplen_low_bitmap = 0x4087fdc0;
s_mplen_high_bitmap = 0x4087fdb0;
s_mplen_vi_bitmap = 0x4087fdac;
s_mplen_bk_bitmap = 0x4087fda8;
esp_wifi_cert_tx_mcs = 0x4087fcfc;
esp_wifi_cert_tx_bcc = 0x4087fcf8;
esp_wifi_cert_tx_nss = 0x4087fcec;
esp_test_tx_statistics_aci_bitmap = 0x4087fda4;
esp_test_tx_statistics = 0x4087fd94;
esp_test_tx_tb_statistics = 0x4087fd84;
esp_test_tx_fail_statistics = 0x4087fd24;
esp_test_rx_statistics = 0x4087fd1c;
esp_test_rx_mu_statistics = 0x4087fd18;
esp_test_mu_print_ru_allocation = 0x4087fd14;
PROVIDE( sigb_ru_allocation_user_num = 0x4004ffc8 );
PROVIDE( sigb_common_ru_allocation = 0x4004ff38 );
PROVIDE( mu_mimo_special_cfg_user_num_2 = 0x4004fee8 );
PROVIDE( mu_mimo_special_cfg_user_num_3 = 0x4004fe80 );
PROVIDE( mu_mimo_special_cfg_user_num_4 = 0x4004fe28 );
PROVIDE( mu_mimo_special_cfg_user_num_5 = 0x4004fdf0 );
PROVIDE( mu_mimo_special_cfg_user_num_6 = 0x4004fdd0 );
PROVIDE( mu_mimo_special_cfg_user_num_7 = 0x4004fdc0 );
PROVIDE( mu_mimo_special_cfg_user_num_8 = 0x4004fdb8 );
esp_test_rx_error_occurs = 0x4087fd10;
g_pp_tx_pkt_num = 0x4087fd0c;
PROVIDE( he_max_apep_length = 0x4004fd40 );

/* --- esp32c6.rom.net80211.ld --- */
PROVIDE( esp_net80211_rom_version_get = 0x40000b4c );
PROVIDE( ampdu_dispatch = 0x40000b50 );
PROVIDE( ampdu_dispatch_all = 0x40000b54 );
PROVIDE( ampdu_dispatch_as_many_as_possible = 0x40000b58 );
PROVIDE( ampdu_dispatch_movement = 0x40000b5c );
PROVIDE( ampdu_dispatch_upto = 0x40000b60 );
PROVIDE( chm_is_at_home_channel = 0x40000b64 );
PROVIDE( cnx_node_is_existing = 0x40000b68 );
PROVIDE( cnx_node_search = 0x40000b6c );
PROVIDE( ic_ebuf_recycle_rx = 0x40000b70 );
PROVIDE( ic_ebuf_recycle_tx = 0x40000b74 );
PROVIDE( ic_reset_rx_ba = 0x40000b78 );
PROVIDE( ieee80211_align_eb = 0x40000b7c );
PROVIDE( ieee80211_ampdu_start_age_timer = 0x40000b84 );
PROVIDE( ieee80211_is_tx_allowed = 0x40000b8c );
PROVIDE( ieee80211_output_pending_eb = 0x40000b90 );
PROVIDE( wifi_get_macaddr = 0x40000ba0 );
PROVIDE( wifi_rf_phy_disable = 0x40000ba4 );
PROVIDE( wifi_rf_phy_enable = 0x40000ba8 );
PROVIDE( ic_ebuf_alloc = 0x40000bac );
PROVIDE( ieee80211_copy_eb_header = 0x40000bb4 );
PROVIDE( ieee80211_recycle_cache_eb = 0x40000bb8 );
PROVIDE( ieee80211_search_node = 0x40000bbc );
PROVIDE( ieee80211_crypto_encap = 0x40000bc0 );
PROVIDE( ieee80211_decap = 0x40000bc8 );
PROVIDE( wifi_is_started = 0x40000bcc );
PROVIDE( ieee80211_gettid = 0x40000bd0 );
net80211_funcs = 0x4087ffac;
g_scan = 0x4087ffa8;
g_chm = 0x4087ffa4;
g_ic_ptr = 0x4087ffa0;
g_hmac_cnt_ptr = 0x4087ff9c;
g_tx_cacheq_ptr = 0x4087ff98;
s_netstack_free = 0x4087ff94;
mesh_rxcb = 0x4087ff90;
sta_rxcb = 0x4087ff8c;
g_itwt_fid = 0x4087ff88;

/* --- esp32c6.rom.coexist.ld --- */
PROVIDE( esp_coex_rom_version_get = 0x40000afc );
PROVIDE( coex_bt_release = 0x40000b00 );
PROVIDE( coex_bt_request = 0x40000b04 );
PROVIDE( coex_core_ble_conn_dyn_prio_get = 0x40000b08 );
PROVIDE( coex_core_pti_get = 0x40000b10 );
PROVIDE( coex_core_release = 0x40000b14 );
PROVIDE( coex_core_request = 0x40000b18 );
PROVIDE( coex_core_status_get = 0x40000b1c );
PROVIDE( coex_event_duration_get = 0x40000b24 );
PROVIDE( coex_hw_timer_disable = 0x40000b28 );
PROVIDE( coex_hw_timer_enable = 0x40000b2c );
PROVIDE( coex_hw_timer_set = 0x40000b30 );
PROVIDE( coex_schm_interval_set = 0x40000b34 );
PROVIDE( coex_schm_lock = 0x40000b38 );
PROVIDE( coex_schm_unlock = 0x40000b3c );
PROVIDE( coex_wifi_release = 0x40000b44 );
PROVIDE( esp_coex_ble_conn_dynamic_prio_get = 0x40000b48 );
coex_env_ptr = 0x4087ffc4;
coex_pti_tab_ptr = 0x4087ffc0;
coex_schm_env_ptr = 0x4087ffbc;
coexist_funcs = 0x4087ffb8;
g_coa_funcs_p = 0x4087ffb4;
g_coex_param_ptr = 0x4087ffb0;

/* --- esp32c6.rom.version.ld --- */
PROVIDE( _rom_chip_id = 0x40000010 );
PROVIDE( _rom_eco_version = 0x40000014 );

/* --- additional.ld --- */
PROVIDE( memset = 0x400004a8 );
PROVIDE( memcpy = 0x400004ac );
PROVIDE( memmove = 0x400004b0 );
PROVIDE( memcmp = 0x400004b4 );
PROVIDE( strncmp = 0x400004c4 );
PROVIDE( strncpy = 0x400004bc );
PROVIDE( strcpy = 0x400004b8 );
PROVIDE( abs = 0x40000578 );
PROVIDE( strcat = 0x4000052c );
PROVIDE( strcmp = 0x400004c0 );
PROVIDE( strchr = 0x40000534 );
PROVIDE( strlcpy = 0x40000544 );
PROVIDE( strstr = 0x400004cc );
PROVIDE( strcasecmp = 0x40000524 );
syscall_table_ptr = 0x4087ffd4;
