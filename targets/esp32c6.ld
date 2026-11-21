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
 * The ROM bootloader rejects IROM or DROM segments of 1 MiB or more, so it
 * only loads .data and .iram. .text and then .rodata follow the RAM part of
 * the image in flash, at the offset that matches their address modulo 8MB.
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
    SRAM (rwx) : ORIGIN = 0x40800000, LENGTH = 0x7E610 /* ROM .bss/.data from 0x4087E610, used by WiFi */

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
        /* Keep .rodata at a 16 byte aligned address and flash offset. */
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
        *(.sbss .sbss.*)
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
        *(.sdata .sdata.*)
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
        *(.text.iram.*)
        *(.wifislprxiram*)
        *(.wifiextrairam*)
        *(.wifi0iram*)
        *(.wifislpiram*)
        *(.wifirxiram*)
        *(.wifiorslpiram*)
        *(.coexiram*)
        *(.high_perf_code_iram1*)
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

    /* === IROM sections (flash-mapped code) === */

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

    /* The ROM only loads the RAM part of the image. esp32c6.S maps flash 1:1
     * at IROM and DROM, so the VMA offset of .text is its flash offset. */
    .irom_dummy (NOLOAD) : ALIGN(0x10000)
    {
        . += _ram_image_size;
    } > IROM

    /* Code stored in IROM (flash-mapped).
     * This is the main program code.
     */
    .text : ALIGN(4)
    {
        *(.text .text.*)
        /* Keep .rodata at a 16 byte aligned address and flash offset. */
        . = ALIGN(16);
    } > IROM

    /* === DROM sections (flash-mapped read-only data) === */

    /* DROM follows IROM in flash. */
    .rodata_dummy (NOLOAD): ALIGN(4)
    {
        . += _ram_image_size + SIZEOF(.text);
    } > DROM

    /* Constant global variables, stored in DROM. */
    .rodata : ALIGN(4)
    {
        *(.rodata*)
        *(.srodata*)
        . = ALIGN (4);
    } >DROM

    ASSERT(ADDR(.text) == ADDR(.irom_dummy) + SIZEOF(.irom_dummy), ".text address does not match its flash offset")
    ASSERT(ADDR(.rodata) == ADDR(.rodata_dummy) + SIZEOF(.rodata_dummy), ".rodata address does not match its flash offset")

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

/* From ESP-IDF:
 * components/esp_rom/esp32c6/ld/esp32c6.rom.newlib.ld
 * The WiFi libraries copy keys into MMIO registers with memcpy, which needs the
 * word copies of the ROM version.
 */
memset  = 0x400004a8;
memcpy  = 0x400004ac;
memmove = 0x400004b0;

/* From ESP-IDF:
 * components/esp_rom/esp32c6/ld/esp32c6.rom.ld
 * components/esp_rom/esp32c6/ld/esp32c6.rom.coexist.ld
 * components/esp_rom/esp32c6/ld/esp32c6.rom.net80211.ld
 * components/esp_rom/esp32c6/ld/esp32c6.rom.phy.ld
 * components/esp_rom/esp32c6/ld/esp32c6.rom.pp.ld
 * These are needed for WiFi support and are available on the Apache 2.0
 * license. Entries that the WiFi libraries replace are commented out, as in
 * esp-hal (esp-rom-sys/ld/esp32c6/rom).
 */

/*
 * SPDX-FileCopyrightText: 2022-2024 Espressif Systems (Shanghai) CO LTD
 *
 * SPDX-License-Identifier: Apache-2.0
 */
/* ROM function interface esp32c6.rom.ld for esp32c6
 *
 *
 * Generated from ./target/esp32c6/interface-esp32c6.yml md5sum 06c13e133e0743d09b87aba30d3e213b
 *
 * Compatible with ROM where ECO version equal or greater to 0.
 *
 * THIS FILE WAS AUTOMATICALLY GENERATED. DO NOT EDIT.
 */

/***************************************
 Group common
 ***************************************/

/* Functions */
rtc_get_reset_reason = 0x40000018;
analog_super_wdt_reset_happened = 0x4000001c;
rtc_get_wakeup_cause = 0x40000020;
rtc_unhold_all_pads = 0x40000024;
ets_printf = 0x40000028;
ets_install_putc1 = 0x4000002c;
ets_install_putc2 = 0x40000030;
ets_install_uart_printf = 0x40000034;
ets_install_usb_printf = 0x40000038;
ets_get_printf_channel = 0x4000003c;
ets_delay_us = 0x40000040;
ets_get_cpu_frequency = 0x40000044;
ets_update_cpu_frequency = 0x40000048;
ets_install_lock = 0x4000004c;
UartRxString = 0x40000050;
UartGetCmdLn = 0x40000054;
uart_tx_one_char = 0x40000058;
uart_tx_one_char2 = 0x4000005c;
uart_rx_one_char = 0x40000060;
uart_rx_one_char_block = 0x40000064;
uart_rx_intr_handler = 0x40000068;
uart_rx_readbuff = 0x4000006c;
uartAttach = 0x40000070;
uart_tx_flush = 0x40000074;
uart_tx_wait_idle = 0x40000078;
uart_div_modify = 0x4000007c;
ets_write_char_uart = 0x40000080;
uart_tx_switch = 0x40000084;
roundup2 = 0x40000088;
multofup = 0x4000008c;
software_reset = 0x40000090;
software_reset_cpu = 0x40000094;
ets_clk_assist_debug_clock_enable = 0x40000098;
clear_super_wdt_reset_flag = 0x4000009c;
disable_default_watchdog = 0x400000a0;
esp_rom_set_rtc_wake_addr = 0x400000a4;
esp_rom_get_rtc_wake_addr = 0x400000a8;
send_packet = 0x400000ac;
recv_packet = 0x400000b0;
GetUartDevice = 0x400000b4;
UartDwnLdProc = 0x400000b8;
GetSecurityInfoProc = 0x400000bc;
Uart_Init = 0x400000c0;
ets_set_user_start = 0x400000c4;
/* Data (.data, .bss, .rodata) */
ets_rom_layout_p = 0x4004fffc;
ets_ops_table_ptr = 0x4087fff8;
g_saved_pc = 0x4087fffc;


/***************************************
 Group miniz
 ***************************************/

/* Functions */
mz_adler32 = 0x400000c8;
mz_free = 0x400000cc;
tdefl_compress = 0x400000d0;
tdefl_compress_buffer = 0x400000d4;
tdefl_compress_mem_to_heap = 0x400000d8;
tdefl_compress_mem_to_mem = 0x400000dc;
tdefl_compress_mem_to_output = 0x400000e0;
tdefl_get_adler32 = 0x400000e4;
tdefl_get_prev_return_status = 0x400000e8;
tdefl_init = 0x400000ec;
tdefl_write_image_to_png_file_in_memory = 0x400000f0;
tdefl_write_image_to_png_file_in_memory_ex = 0x400000f4;
tinfl_decompress = 0x400000f8;
tinfl_decompress_mem_to_callback = 0x400000fc;
tinfl_decompress_mem_to_heap = 0x40000100;
tinfl_decompress_mem_to_mem = 0x40000104;


/***************************************
 Group tjpgd
 ***************************************/

/* Functions */
jd_prepare = 0x40000108;
jd_decomp = 0x4000010c;


/***************************************
 Group spiflash_legacy
 ***************************************/

/* Functions */
esp_rom_spiflash_wait_idle = 0x40000110;
esp_rom_spiflash_write_encrypted = 0x40000114;
esp_rom_spiflash_write_encrypted_dest = 0x40000118;
esp_rom_spiflash_write_encrypted_enable = 0x4000011c;
esp_rom_spiflash_write_encrypted_disable = 0x40000120;
esp_rom_spiflash_erase_chip = 0x40000124;
_esp_rom_spiflash_erase_sector = 0x40000128;
_esp_rom_spiflash_erase_block = 0x4000012c;
_esp_rom_spiflash_write = 0x40000130;
_esp_rom_spiflash_read = 0x40000134;
_esp_rom_spiflash_unlock = 0x40000138;
_SPIEraseArea = 0x4000013c;
_SPI_write_enable = 0x40000140;
esp_rom_spiflash_erase_sector = 0x40000144;
esp_rom_spiflash_erase_block = 0x40000148;
esp_rom_spiflash_write = 0x4000014c;
esp_rom_spiflash_read = 0x40000150;
esp_rom_spiflash_unlock = 0x40000154;
SPIEraseArea = 0x40000158;
SPI_write_enable = 0x4000015c;
esp_rom_spiflash_config_param = 0x40000160;
esp_rom_spiflash_read_user_cmd = 0x40000164;
esp_rom_spiflash_select_qio_pins = 0x40000168;
esp_rom_spi_flash_auto_sus_res = 0x4000016c;
esp_rom_spi_flash_send_resume = 0x40000170;
esp_rom_spi_flash_update_id = 0x40000174;
esp_rom_spiflash_config_clk = 0x40000178;
esp_rom_spiflash_config_readmode = 0x4000017c;
esp_rom_spiflash_read_status = 0x40000180;
esp_rom_spiflash_read_statushigh = 0x40000184;
esp_rom_spiflash_write_status = 0x40000188;
spi_cache_mode_switch = 0x4000018c;
spi_common_set_dummy_output = 0x40000190;
spi_common_set_flash_cs_timing = 0x40000194;
esp_rom_spi_set_address_bit_len = 0x40000198;
SPILock = 0x4000019c;
SPIMasterReadModeCnfig = 0x400001a0;
SPI_Common_Command = 0x400001a4;
SPI_WakeUp = 0x400001a8;
SPI_block_erase = 0x400001ac;
SPI_chip_erase = 0x400001b0;
SPI_init = 0x400001b4;
SPI_page_program = 0x400001b8;
SPI_read_data = 0x400001bc;
SPI_sector_erase = 0x400001c0;
SelectSpiFunction = 0x400001c4;
SetSpiDrvs = 0x400001c8;
Wait_SPI_Idle = 0x400001cc;
spi_dummy_len_fix = 0x400001d0;
Disable_QMode = 0x400001d4;
Enable_QMode = 0x400001d8;
spi_flash_attach = 0x400001dc;
spi_flash_get_chip_size = 0x400001e0;
spi_flash_guard_set = 0x400001e4;
spi_flash_guard_get = 0x400001e8;
spi_flash_read_encrypted = 0x400001ec;
/* Data (.data, .bss, .rodata) */
rom_spiflash_legacy_funcs = 0x4087fff0;
rom_spiflash_legacy_data = 0x4087ffec;
g_flash_guard_ops = 0x4087fff4;

/* Note: esp_rom_spiflash_write_disable was moved from esp32c6.rom.spiflash.ld */
esp_rom_spiflash_write_disable = 0x40000278;


/***************************************
 Group cache
 ***************************************/

/* Functions */
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
/* Data (.data, .bss, .rodata) */
rom_cache_op_cb = 0x4087ffcc;
rom_cache_internal_table_ptr = 0x4087ffc8;


/***************************************
 Group clock
 ***************************************/

/* Functions */
ets_clk_get_xtal_freq = 0x400006c4;
ets_clk_get_cpu_freq = 0x400006c8;
ets_clk_apb_wait_ready = 0x400006cc;
ets_clk_mspi_apb_wait_ready = 0x400006d0;


/***************************************
 Group gpio
 ***************************************/

/* Functions */
gpio_input_get = 0x400006d4;
gpio_matrix_in = 0x400006d8;
gpio_matrix_out = 0x400006dc;
/* gpio_output_disable = 0x400006e0; */
/* gpio_output_enable = 0x400006e4; */
gpio_output_set = 0x400006e8;
gpio_pad_hold = 0x400006ec;
gpio_pad_input_disable = 0x400006f0;
gpio_pad_input_enable = 0x400006f4;
gpio_pad_pulldown = 0x400006f8;
gpio_pad_pullup = 0x400006fc;
gpio_pad_select_gpio = 0x40000700;
gpio_pad_set_drv = 0x40000704;
gpio_pad_unhold = 0x40000708;
gpio_pin_wakeup_disable = 0x4000070c;
gpio_pin_wakeup_enable = 0x40000710;
gpio_bypass_matrix_in = 0x40000714;


/***************************************
 Group interrupts
 ***************************************/

/* Functions */
PROVIDE( esprv_intc_int_set_priority = 0x40000718 );
PROVIDE( esprv_intc_int_set_threshold = 0x4000071c );
PROVIDE( esprv_intc_int_enable = 0x40000720 );
PROVIDE( esprv_intc_int_disable = 0x40000724 );
PROVIDE( esprv_intc_int_set_type = 0x40000728 );
PROVIDE( intr_handler_set = 0x4000072c );
intr_matrix_set = 0x40000730;
ets_intr_lock = 0x40000734;
ets_intr_unlock = 0x40000738;
ets_isr_attach = 0x4000073c;
ets_isr_mask = 0x40000740;
ets_isr_unmask = 0x40000744;


/***************************************
 Group crypto
 ***************************************/

/* Functions */
md5_vector = 0x40000748;
MD5Init = 0x4000074c;
MD5Update = 0x40000750;
MD5Final = 0x40000754;
crc32_le = 0x40000758;
crc16_le = 0x4000075c;
crc8_le = 0x40000760;
crc32_be = 0x40000764;
crc16_be = 0x40000768;
crc8_be = 0x4000076c;
esp_crc8 = 0x40000770;
ets_sha_enable = 0x40000774;
ets_sha_disable = 0x40000778;
ets_sha_get_state = 0x4000077c;
ets_sha_init = 0x40000780;
ets_sha_process = 0x40000784;
ets_sha_starts = 0x40000788;
ets_sha_update = 0x4000078c;
ets_sha_finish = 0x40000790;
ets_sha_clone = 0x40000794;
ets_hmac_enable = 0x40000798;
ets_hmac_disable = 0x4000079c;
ets_hmac_calculate_message = 0x400007a0;
ets_hmac_calculate_downstream = 0x400007a4;
ets_hmac_invalidate_downstream = 0x400007a8;
ets_jtag_enable_temporarily = 0x400007ac;
ets_aes_enable = 0x400007b0;
ets_aes_disable = 0x400007b4;
ets_aes_setkey = 0x400007b8;
ets_aes_block = 0x400007bc;
ets_aes_setkey_dec = 0x400007c0;
ets_aes_setkey_enc = 0x400007c4;
ets_bigint_enable = 0x400007c8;
ets_bigint_disable = 0x400007cc;
ets_bigint_multiply = 0x400007d0;
ets_bigint_modmult = 0x400007d4;
ets_bigint_modexp = 0x400007d8;
ets_bigint_wait_finish = 0x400007dc;
ets_bigint_getz = 0x400007e0;
ets_ds_enable = 0x400007e4;
ets_ds_disable = 0x400007e8;
ets_ds_start_sign = 0x400007ec;
ets_ds_is_busy = 0x400007f0;
ets_ds_finish_sign = 0x400007f4;
ets_ds_encrypt_params = 0x400007f8;
ets_mgf1_sha256 = 0x400007fc;
/* Data (.data, .bss, .rodata) */
crc32_le_table_ptr = 0x4004fff8;
crc16_le_table_ptr = 0x4004fff4;
crc8_le_table_ptr = 0x4004fff0;
crc32_be_table_ptr = 0x4004ffec;
crc16_be_table_ptr = 0x4004ffe8;
crc8_be_table_ptr = 0x4004ffe4;


/***************************************
 Group efuse
 ***************************************/

/* Functions */
ets_efuse_read = 0x40000800;
ets_efuse_program = 0x40000804;
ets_efuse_clear_program_registers = 0x40000808;
ets_efuse_write_key = 0x4000080c;
ets_efuse_get_read_register_address = 0x40000810;
ets_efuse_get_key_purpose = 0x40000814;
ets_efuse_key_block_unused = 0x40000818;
ets_efuse_find_unused_key_block = 0x4000081c;
ets_efuse_rs_calculate = 0x40000820;
ets_efuse_count_unused_key_blocks = 0x40000824;
ets_efuse_secure_boot_enabled = 0x40000828;
ets_efuse_secure_boot_aggressive_revoke_enabled = 0x4000082c;
ets_efuse_cache_encryption_enabled = 0x40000830;
ets_efuse_download_modes_disabled = 0x40000834;
ets_efuse_find_purpose = 0x40000838;
ets_efuse_force_send_resume = 0x4000083c;
ets_efuse_get_flash_delay_us = 0x40000840;
ets_efuse_get_mac = 0x40000844;
ets_efuse_get_uart_print_control = 0x40000848;
ets_efuse_direct_boot_mode_disabled = 0x4000084c;
ets_efuse_security_download_modes_enabled = 0x40000850;
ets_efuse_set_timing = 0x40000854;
ets_efuse_jtag_disabled = 0x40000858;
ets_efuse_usb_print_is_disabled = 0x4000085c;
ets_efuse_usb_download_mode_disabled = 0x40000860;
ets_efuse_usb_device_disabled = 0x40000864;
ets_efuse_secure_boot_fast_wake_enabled = 0x40000868;


/***************************************
 Group secureboot
 ***************************************/

/* Functions */
ets_emsa_pss_verify = 0x4000086c;
ets_rsa_pss_verify = 0x40000870;
ets_secure_boot_verify_bootloader_with_keys = 0x40000874;
ets_secure_boot_verify_signature = 0x40000878;
ets_secure_boot_read_key_digests = 0x4000087c;
ets_secure_boot_revoke_public_key_digest = 0x40000880;


/***************************************
 Group usb_device_uart
 ***************************************/

/* Functions */
usb_serial_device_rx_one_char = 0x40000a80;
usb_serial_device_rx_one_char_block = 0x40000a84;
usb_serial_device_tx_flush = 0x40000a88;
usb_serial_device_tx_one_char = 0x40000a8c;


/***************************************
 Group lldesc
 ***************************************/

/* Functions */
lldesc_build_chain = 0x40000a90;


/***************************************
 Group sip
 ***************************************/

/* Functions */
sip_after_tx_complete = 0x40000a94;
sip_alloc_to_host_evt = 0x40000a98;
sip_download_begin = 0x40000a9c;
sip_get_ptr = 0x40000aa0;
sip_get_state = 0x40000aa4;
sip_init_attach = 0x40000aa8;
sip_install_rx_ctrl_cb = 0x40000aac;
sip_install_rx_data_cb = 0x40000ab0;
sip_is_active = 0x40000ab4;
sip_post_init = 0x40000ab8;
sip_reclaim_from_host_cmd = 0x40000abc;
sip_reclaim_tx_data_pkt = 0x40000ac0;
sip_send = 0x40000ac4;
sip_to_host_chain_append = 0x40000ac8;
sip_to_host_evt_send_done = 0x40000acc;


/***************************************
 Group slc
 ***************************************/

/* Functions */
slc_add_credits = 0x40000ad0;
slc_enable = 0x40000ad4;
slc_from_host_chain_fetch = 0x40000ad8;
slc_from_host_chain_recycle = 0x40000adc;
slc_has_pkt_to_host = 0x40000ae0;
slc_init_attach = 0x40000ae4;
slc_init_credit = 0x40000ae8;
slc_reattach = 0x40000aec;
slc_send_to_host_chain = 0x40000af0;
slc_set_host_io_max_window = 0x40000af4;
slc_to_host_chain_recycle = 0x40000af8;

/*
 * SPDX-FileCopyrightText: 2022 Espressif Systems (Shanghai) CO LTD
 *
 * SPDX-License-Identifier: Apache-2.0
 */
/* ROM function interface esp32c6.rom.coexist.ld for esp32c6
 *
 *
 * Generated from ./target/esp32c6/interface-esp32c6.yml md5sum 06c13e133e0743d09b87aba30d3e213b
 *
 * Compatible with ROM where ECO version equal or greater to 0.
 *
 * THIS FILE WAS AUTOMATICALLY GENERATED. DO NOT EDIT.
 */

/***************************************
 Group rom_coexist
 ***************************************/

/* Functions */
esp_coex_rom_version_get = 0x40000afc;
coex_bt_release = 0x40000b00;
coex_bt_request = 0x40000b04;
coex_core_ble_conn_dyn_prio_get = 0x40000b08;
/*coex_core_event_duration_get = 0x40000b0c;*/
coex_core_pti_get = 0x40000b10;
coex_core_release = 0x40000b14;
coex_core_request = 0x40000b18;
coex_core_status_get = 0x40000b1c;
/*coex_core_timer_idx_get = 0x40000b20;*/
coex_event_duration_get = 0x40000b24;
coex_hw_timer_disable = 0x40000b28;
coex_hw_timer_enable = 0x40000b2c;
coex_hw_timer_set = 0x40000b30;
coex_schm_interval_set = 0x40000b34;
coex_schm_lock = 0x40000b38;
coex_schm_unlock = 0x40000b3c;
/*coex_status_get = 0x40000b40;*/
coex_wifi_release = 0x40000b44;
esp_coex_ble_conn_dynamic_prio_get = 0x40000b48;
/* Data (.data, .bss, .rodata) */
coex_env_ptr = 0x4087ffc4;
coex_pti_tab_ptr = 0x4087ffc0;
coex_schm_env_ptr = 0x4087ffbc;
coexist_funcs = 0x4087ffb8;
g_coa_funcs_p = 0x4087ffb4;
g_coex_param_ptr = 0x4087ffb0;

/*
 * SPDX-FileCopyrightText: 2022 Espressif Systems (Shanghai) CO LTD
 *
 * SPDX-License-Identifier: Apache-2.0
 */
/* ROM function interface esp32c6.rom.net80211.ld for esp32c6
 *
 *
 * Generated from ./target/esp32c6/interface-esp32c6.yml md5sum 06c13e133e0743d09b87aba30d3e213b
 *
 * Compatible with ROM where ECO version equal or greater to 0.
 *
 * THIS FILE WAS AUTOMATICALLY GENERATED. DO NOT EDIT.
 */

/***************************************
 Group rom_net80211
 ***************************************/

/* Functions */
esp_net80211_rom_version_get = 0x40000b4c;
ampdu_dispatch = 0x40000b50;
ampdu_dispatch_all = 0x40000b54;
ampdu_dispatch_as_many_as_possible = 0x40000b58;
ampdu_dispatch_movement = 0x40000b5c;
ampdu_dispatch_upto = 0x40000b60;
chm_is_at_home_channel = 0x40000b64;
cnx_node_is_existing = 0x40000b68;
cnx_node_search = 0x40000b6c;
ic_ebuf_recycle_rx = 0x40000b70;
ic_ebuf_recycle_tx = 0x40000b74;
ic_reset_rx_ba = 0x40000b78;
ieee80211_align_eb = 0x40000b7c;
/*ieee80211_ampdu_reorder = 0x40000b80;*/
ieee80211_ampdu_start_age_timer = 0x40000b84;
/*ieee80211_encap_esfbuf = 0x40000b88;*/
ieee80211_is_tx_allowed = 0x40000b8c;
ieee80211_output_pending_eb = 0x40000b90;
/*ieee80211_output_process = 0x40000b94;*/
/* ieee80211_set_tx_desc = 0x40000b98; */
/* sta_input = 0x40000b9c; */
wifi_get_macaddr = 0x40000ba0;
wifi_rf_phy_disable = 0x40000ba4;
wifi_rf_phy_enable = 0x40000ba8;
ic_ebuf_alloc = 0x40000bac;
/*ieee80211_classify = 0x40000bb0;*/
ieee80211_copy_eb_header = 0x40000bb4;
ieee80211_recycle_cache_eb = 0x40000bb8;
ieee80211_search_node = 0x40000bbc;
ieee80211_crypto_encap = 0x40000bc0;
/* ieee80211_crypto_decap = 0x40000bc4; */
ieee80211_decap = 0x40000bc8;
wifi_is_started = 0x40000bcc;
ieee80211_gettid = 0x40000bd0;
/* ieee80211_encap_esfbuf_htc = 0x40000bd4; */
/* Data (.data, .bss, .rodata) */
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

/*
 * SPDX-FileCopyrightText: 2022 Espressif Systems (Shanghai) CO LTD
 *
 * SPDX-License-Identifier: Apache-2.0
 */
/* ROM function interface esp32c6.rom.phy.ld for esp32c6
 *
 *
 * Generated from ./target/esp32c6/interface-esp32c6.yml md5sum 06c13e133e0743d09b87aba30d3e213b
 *
 * Compatible with ROM where ECO version equal or greater to 0.
 *
 * THIS FILE WAS AUTOMATICALLY GENERATED. DO NOT EDIT.
 */

/***************************************
 Group rom_phy
 ***************************************/

/* Functions */
phy_param_addr = 0x40001104;
phy_get_romfuncs = 0x40001108;
chip761_phyrom_version = 0x4000110c;
chip761_phyrom_version_num = 0x40001110;
get_rc_dout = 0x40001114;
rc_cal = 0x40001118;
rom_enter_critical_phy = 0x4000111c;
rom_exit_critical_phy = 0x40001120;
rom_set_chan_cal_interp = 0x40001124;
rom_loopback_mode_en = 0x40001128;
rom_bb_bss_cbw40 = 0x4000112c;
abs_temp = 0x40001130;
get_data_sat = 0x40001134;
phy_byte_to_word = 0x40001138;
set_chan_reg = 0x4000113c;
i2c_master_reset = 0x40001140;
rom_set_chan_freq_sw_start = 0x40001144;
freq_module_resetn = 0x40001148;
freq_chan_en_sw = 0x4000114c;
write_chan_freq = 0x40001150;
get_freq_mem_param = 0x40001154;
get_freq_mem_addr = 0x40001158;
bt_txpwr_freq = 0x4000115c;
wr_rf_freq_mem = 0x40001160;
read_rf_freq_mem = 0x40001164;
freq_i2c_mem_write = 0x40001168;
freq_num_get_data = 0x4000116c;
freq_i2c_num_addr = 0x40001170;
freq_i2c_write_set = 0x40001174;
pll_dac_mem_update = 0x40001178;
pll_cap_mem_update = 0x4000117c;
get_rf_freq_cap = 0x40001180;
get_rf_freq_init = 0x40001184;
phy_en_hw_set_freq = 0x40001188;
phy_dis_hw_set_freq = 0x4000118c;
rom_pwdet_sar2_init = 0x40001190;
rom_en_pwdet = 0x40001194;
rom_get_sar_sig_ref = 0x40001198;
rom_pwdet_tone_start = 0x4000119c;
rom_pwdet_wait_idle = 0x400011a0;
rom_read_sar_dout = 0x400011a4;
get_tone_sar_dout = 0x400011a8;
get_fm_sar_dout = 0x400011ac;
txtone_linear_pwr = 0x400011b0;
linear_to_db = 0x400011b4;
get_power_db = 0x400011b8;
meas_tone_pwr_db = 0x400011bc;
pkdet_vol_start = 0x400011c0;
read_sar2_code = 0x400011c4;
get_sar2_vol = 0x400011c8;
get_pll_vol = 0x400011cc;
tx_pwctrl_bg_init = 0x400011d0;
phy_pwdet_always_en = 0x400011d4;
phy_pwdet_onetime_en = 0x400011d8;
esp_tx_state_out_rom = 0x400011dc;
ant_dft_cfg_rom = 0x400011e0;
ant_wifitx_cfg_rom = 0x400011e4;
ant_wifirx_cfg_rom = 0x400011e8;
ant_bttx_cfg_rom = 0x400011ec;
ant_btrx_cfg_rom = 0x400011f0;
phy_chan_dump_cfg_rom = 0x400011f4;
phy_enable_low_rate = 0x400011f8;
phy_disable_low_rate = 0x400011fc;
phy_is_low_rate_enabled = 0x40001200;
phy_dig_reg_backup_rom = 0x40001204;
phy_chan_filt_set_rom = 0x40001208;
phy_rx11blr_cfg = 0x4000120c;
set_cca_rom = 0x40001210;
set_rx_sense_rom = 0x40001214;
rx_gain_force_rom = 0x40001218;
rom_rfpll_set_freq = 0x4000121c;
mhz2ieee = 0x40001220;
chan_to_freq = 0x40001224;
restart_cal = 0x40001228;
write_rfpll_sdm = 0x4000122c;
wait_rfpll_cal_end = 0x40001230;
set_rf_freq_offset = 0x40001234;
set_rfpll_freq = 0x40001238;
set_channel_rfpll_freq = 0x4000123c;
rfpll_cap_correct = 0x40001240;
rfpll_cap_init_cal = 0x40001244;
write_pll_cap = 0x40001248;
read_pll_cap = 0x4000124c;
chip_v7_set_chan_ana = 0x40001250;
freq_set_reg = 0x40001254;
gen_rx_gain_table = 0x40001258;
bt_txdc_cal = 0x4000125c;
bt_txiq_cal = 0x40001260;
txiq_cal_init = 0x40001264;
txdc_cal_init = 0x40001268;
txdc_cal = 0x4000126c;
txiq_get_mis_pwr = 0x40001270;
txiq_cover = 0x40001274;
rfcal_txiq = 0x40001278;
get_power_atten = 0x4000127c;
pwdet_ref_code = 0x40001280;
pwdet_code_cal = 0x40001284;
rfcal_txcap = 0x40001288;
tx_cap_init = 0x4000128c;
rfcal_pwrctrl = 0x40001290;
tx_pwctrl_init_cal = 0x40001294;
tx_pwctrl_init = 0x40001298;
bt_tx_pwctrl_init = 0x4000129c;
rom_i2c_enter_critical = 0x400012a0;
rom_i2c_exit_critical = 0x400012a4;
rom_get_i2c_read_mask = 0x400012a8;
rom_get_i2c_mst0_mask = 0x400012ac;
rom_get_i2c_hostid = 0x400012b0;
rom_chip_i2c_readReg_org = 0x400012b4;
rom_chip_i2c_readReg = 0x400012b8;
rom_chip_i2c_writeReg = 0x400012c0;
rom_set_txcap_reg = 0x400012d0;
i2c_paral_set_mst0 = 0x400012d4;
i2c_paral_set_read = 0x400012d8;
i2c_paral_read = 0x400012dc;
i2c_paral_write = 0x400012e0;
i2c_paral_write_num = 0x400012e4;
i2c_paral_write_mask = 0x400012e8;
i2c_sar2_init_code = 0x400012ec;
rom_pbus_force_mode = 0x400012f0;
rom_pbus_rd_addr = 0x400012f4;
rom_pbus_rd_shift = 0x400012f8;
rom_pbus_force_test = 0x400012fc;
rom_pbus_rd = 0x40001300;
rom_pbus_set_rxgain = 0x40001304;
rom_pbus_xpd_rx_off = 0x40001308;
rom_pbus_xpd_rx_on = 0x4000130c;
rom_pbus_xpd_tx_off = 0x40001310;
rom_pbus_xpd_tx_on = 0x40001314;
rom_set_loopback_gain = 0x40001318;
rom_txcal_debuge_mode = 0x4000131c;
pbus_debugmode = 0x40001320;
pbus_workmode = 0x40001324;
pbus_set_dco = 0x40001328;
txcal_work_mode = 0x4000132c;
rom_start_tx_tone_step = 0x40001330;
rom_stop_tx_tone = 0x40001334;
disable_agc = 0x40001338;
enable_agc = 0x4000133c;
phy_disable_cca = 0x40001340;
phy_enable_cca = 0x40001344;
write_gain_mem = 0x40001348;
bb_bss_cbw40_dig = 0x4000134c;
cbw2040_cfg = 0x40001350;
mac_tx_chan_offset = 0x40001354;
tx_paon_set = 0x40001358;
pwdet_reg_init = 0x4000135c;
i2cmst_reg_init = 0x40001360;
bt_gain_offset = 0x40001364;
fe_reg_init = 0x40001368;
mac_enable_bb = 0x4000136c;
bb_wdg_cfg = 0x40001370;
fe_txrx_reset = 0x40001374;
set_rx_comp = 0x40001378;
agc_reg_init = 0x4000137c;
bb_reg_init = 0x40001380;
open_i2c_xpd = 0x40001384;
txiq_set_reg = 0x40001388;
rxiq_set_reg = 0x4000138c;
set_txclk_en = 0x40001390;
set_rxclk_en = 0x40001394;
bb_wdg_test_en = 0x40001398;
noise_floor_auto_set = 0x4000139c;
read_hw_noisefloor = 0x400013a0;
iq_corr_enable = 0x400013a4;
wifi_agc_sat_gain = 0x400013a8;
phy_bbpll_cal = 0x400013ac;
phy_ant_init = 0x400013b0;
phy_set_bbfreq_init = 0x400013b4;
wifi_fbw_sel = 0x400013b8;
bt_filter_reg = 0x400013bc;
phy_rx_sense_set = 0x400013c0;
tx_state_set = 0x400013c4;
phy_close_pa = 0x400013c8;
phy_freq_correct = 0x400013cc;
set_pbus_reg = 0x400013d0;
wifi_rifs_mode_en = 0x400013d4;
nrx_freq_set = 0x400013d8;
fe_adc_on = 0x400013dc;
phy_force_pwr_index = 0x400013e0;
rom_iq_est_enable = 0x400013e4;
rom_iq_est_disable = 0x400013e8;
rom_bb_gain_index = 0x400013ec;
rom_rfrx_gain_index = 0x400013f0;
dc_iq_est = 0x400013f4;
set_cal_rxdc = 0x400013f8;
rxiq_get_mis = 0x400013fc;
rxiq_cover_mg_mp = 0x40001400;
rfcal_rxiq = 0x40001404;
get_rfcal_rxiq_data = 0x40001408;
get_dco_comp = 0x4000140c;
pbus_rx_dco_cal = 0x40001410;
rxdc_est_min = 0x40001414;
pbus_rx_dco_cal_1step = 0x40001418;
set_lb_txiq = 0x4000141c;
set_rx_gain_cal_iq = 0x40001420;
set_rx_gain_cal_dc = 0x40001424;
spur_reg_write_one_tone = 0x40001428;
spur_cal = 0x4000142c;
spur_coef_cfg = 0x40001430;
tsens_power_up = 0x40001434;
tsens_read_init = 0x40001438;
code_to_temp = 0x4000143c;
tsens_index_to_dac = 0x40001440;
tsens_index_to_offset = 0x40001444;
tsens_dac_cal = 0x40001448;
tsens_code_read = 0x4000144c;
tsens_temp_read = 0x40001450;
temp_to_power = 0x40001454;
get_temp_init = 0x40001458;
txbbgain_to_index = 0x4000145c;
index_to_txbbgain = 0x40001460;
bt_index_to_bb = 0x40001464;
bt_bb_to_index = 0x40001468;
bt_get_tx_gain = 0x4000146c;
dig_gain_check = 0x40001470;
wifi_get_tx_gain = 0x40001474;
wifi_11g_rate_chg = 0x40001478;
bt_chan_pwr_interp = 0x4000147c;
get_rate_fcc_index = 0x40001480;
get_chan_target_power = 0x40001484;
get_tx_gain_value = 0x40001488;
wifi_get_target_power = 0x4000148c;
/* Data (.data, .bss, .rodata) */
phy_param_rom = 0x4087fce8;

/*
 * SPDX-FileCopyrightText: 2022-2024 Espressif Systems (Shanghai) CO LTD
 *
 * SPDX-License-Identifier: Apache-2.0
 */
/* ROM function interface esp32c6.rom.pp.ld for esp32c6
 *
 *
 * Generated from ./target/esp32c6/interface-esp32c6.yml md5sum 06c13e133e0743d09b87aba30d3e213b
 *
 * Compatible with ROM where ECO version equal or greater to 0.
 *
 * THIS FILE WAS AUTOMATICALLY GENERATED. DO NOT EDIT.
 */

/***************************************
 Group rom_pp
 ***************************************/

/* Functions */
esp_pp_rom_version_get = 0x40000bd8;
ppCalTxopRTSThreshold = 0x40000bdc;
RC_GetBlockAckTime = 0x40000be0;
ebuf_list_remove = 0x40000be4;
/* esf_buf_alloc = 0x40000be8; */
/* esf_buf_alloc_dynamic = 0x40000bec; */
/* esf_buf_recycle = 0x40000bf0; */
GetAccess = 0x40000bf4;
hal_mac_is_low_rate_enabled = 0x40000bf8;
hal_mac_tx_get_blockack = 0x40000bfc;
/* hal_mac_tx_set_ppdu = 0x40000c00; */
ic_get_trc = 0x40000c04;
/* ic_mac_deinit = 0x40000c08; */
/* ic_mac_init = 0x40000c0c; */
ic_interface_enabled = 0x40000c10;
is_lmac_idle = 0x40000c14;
/*lmacAdjustTimestamp = 0x40000c18;*/
lmacDiscardAgedMSDU = 0x40000c1c;
/*lmacDiscardMSDU = 0x40000c20;*/
/*lmacEndFrameExchangeSequence = 0x40000c24;*/
lmacIsIdle = 0x40000c28;
lmacIsLongFrame = 0x40000c2c;
/*lmacMSDUAged = 0x40000c30;*/
lmacPostTxComplete = 0x40000c34;
lmacProcessAllTxTimeout = 0x40000c38;
lmacProcessCollisions = 0x40000c3c;
/* lmacProcessRxSucData = 0x40000c40; */
lmacReachLongLimit = 0x40000c44;
lmacReachShortLimit = 0x40000c48;
lmacRecycleMPDU = 0x40000c4c;
lmacRxDone = 0x40000c50;
/* lmacSetTxFrame = 0x40000c54; */
/* lmacTxDone = 0x40000c58; */
/* lmacTxFrame = 0x40000c5c; */
mac_tx_set_duration = 0x40000c60;
/* mac_tx_set_plcp0 = 0x40000c64; */
/* mac_tx_set_plcp1 = 0x40000c68; */
mac_tx_set_plcp2 = 0x40000c6c;
/* pm_check_state = 0x40000c70; */
/* pm_disable_dream_timer = 0x40000c74; */
pm_disable_sleep_delay_timer = 0x40000c78;
/*pm_dream = 0x40000c7c;*/
pm_mac_wakeup = 0x40000c80;
pm_mac_sleep = 0x40000c84;
/* pm_enable_active_timer = 0x40000c88; */
pm_enable_sleep_delay_timer = 0x40000c8c;
pm_local_tsf_process = 0x40000c90;
/* pm_set_beacon_filter = 0x40000c94; */
/*pm_is_in_wifi_slice_threshold = 0x40000c98;*/
pm_is_waked = 0x40000c9c;
/* pm_keep_alive = 0x40000ca0; */
/* pm_on_beacon_rx = 0x40000ca4; */
pm_on_data_rx = 0x40000ca8;
/* pm_on_tbtt = 0x40000cac; */
/* pm_parse_beacon = 0x40000cb0; */
/* pm_process_tim = 0x40000cb4; */
/* pm_rx_beacon_process = 0x40000cb8; */
/* pm_rx_data_process = 0x40000cbc; */
/* pm_sleep = 0x40000cc0; */
pm_sleep_for = 0x40000cc4;
/* pm_tbtt_process = 0x40000cc8; */
ppAMPDU2Normal = 0x40000ccc;
/* ppAssembleAMPDU = 0x40000cd0; */
ppCalFrameTimes = 0x40000cd4;
ppCalSubFrameLength = 0x40000cd8;
/* ppCalTxAMPDULength = 0x40000cdc; */
ppCheckTxAMPDUlength = 0x40000ce0;
ppDequeueRxq_Locked = 0x40000ce4;
ppDequeueTxQ = 0x40000ce8;
ppEmptyDelimiterLength = 0x40000cec;
ppEnqueueRxq = 0x40000cf0;
ppEnqueueTxDone = 0x40000cf4;
ppGetTxframe = 0x40000cf8;
/* ppMapTxQueue = 0x40000cfc; */
/* ppProcTxSecFrame = 0x40000d00; */
ppProcessRxPktHdr = 0x40000d04;
/* ppProcessTxQ = 0x40000d08; */
ppRecordBarRRC = 0x40000d0c;
ppRecycleAmpdu = 0x40000d10;
ppRecycleRxPkt = 0x40000d14;
/* ppResortTxAMPDU = 0x40000d18; */
ppResumeTxAMPDU = 0x40000d1c;
/*ppRxFragmentProc = 0x40000d20;*/
/* ppRxPkt = 0x40000d24; */
/* ppRxProtoProc = 0x40000d28; */
ppSearchTxQueue = 0x40000d2c;
ppSearchTxframe = 0x40000d30;
ppSelectNextQueue = 0x40000d34;
ppSubFromAMPDU = 0x40000d38;
/* ppTask = 0x40000d3c; */
/* ppTxPkt = 0x40000d40; */
ppTxProtoProc = 0x40000d44;
ppTxqUpdateBitmap = 0x40000d48;
/*pp_coex_tx_request = 0x40000d4c;*/
pp_hdrsize = 0x40000d50;
pp_post = 0x40000d54;
pp_process_hmac_waiting_txq = 0x40000d58;
rcGetAmpduSched = 0x40000d5c;
rcUpdateRxDone = 0x40000d60;
rc_get_trc = 0x40000d64;
rc_get_trc_by_index = 0x40000d68;
rcAmpduLowerRate = 0x40000d6c;
rcampduuprate = 0x40000d70;
rcClearCurAMPDUSched = 0x40000d74;
rcClearCurSched = 0x40000d78;
rcClearCurStat = 0x40000d7c;
/*rcGetSched = 0x40000d80;*/
rcLowerSched = 0x40000d84;
rcSetTxAmpduLimit = 0x40000d88;
rcTxUpdatePer = 0x40000d8c;
rcUpdateAckSnr = 0x40000d90;
/*rcUpdateRate = 0x40000d94;*/
/*rcUpdateTxDone = 0x40000d98;*/
/*rcUpdateTxDoneAmpdu2 = 0x40000d9c;*/
rcUpSched = 0x40000da0;
rssi_margin = 0x40000da4;
rx11NRate2AMPDULimit = 0x40000da8;
TRC_AMPDU_PER_DOWN_THRESHOLD = 0x40000dac;
TRC_AMPDU_PER_UP_THRESHOLD = 0x40000db0;
trc_calc_duration = 0x40000db4;
trc_isTxAmpduOperational = 0x40000db8;
trc_onAmpduOp = 0x40000dbc;
TRC_PER_IS_GOOD = 0x40000dc0;
trc_SetTxAmpduState = 0x40000dc4;
trc_tid_isTxAmpduOperational = 0x40000dc8;
trcAmpduSetState = 0x40000dcc;
/* wDevCheckBlockError = 0x40000dd0; */
/*wDev_AppendRxBlocks = 0x40000dd4;*/
wDev_DiscardFrame = 0x40000dd8;
wDev_GetNoiseFloor = 0x40000ddc;
wDev_IndicateAmpdu = 0x40000de0;
/* wDev_IndicateFrame = 0x40000de4; */
wdev_mac_reg_load = 0x40000de8;
wdev_mac_reg_store = 0x40000dec;
wdev_mac_special_reg_load = 0x40000df0;
wdev_mac_special_reg_store = 0x40000df4;
wdev_mac_wakeup = 0x40000df8;
wdev_mac_sleep = 0x40000dfc;
hal_mac_is_dma_enable = 0x40000e00;
/* wDev_ProcessFiq = 0x40000e04; */
/* wDev_ProcessRxSucData = 0x40000e08; */
/* wdevProcessRxSucDataAll = 0x40000e0c; */
wdev_csi_len_align = 0x40000e10;
ppDequeueTxDone_Locked = 0x40000e14;
/* ppProcTxDone = 0x40000e18; */
/* pm_tx_data_done_process = 0x40000e1c; */
config_is_cache_tx_buf_enabled = 0x40000e20;
/* ppMapWaitTxq = 0x40000e24; */
ppProcessWaitingQueue = 0x40000e28;
ppDisableQueue = 0x40000e2c;
pm_allow_tx = 0x40000e30;
/* wdev_is_data_in_rxlist = 0x40000e34; */
ppProcTxCallback = 0x40000e38;
/* mac_tx_set_hesig = 0x40000e3c; */
ppCalPreFecPaddingFactor = 0x40000e40;
/* mac_tx_set_tb = 0x40000e44; */
/* mac_tx_set_mplen = 0x40000e48; */
hal_get_tsf_timer = 0x40000e4c;
ppTxPktForceWaked = 0x40000e50;
lmacProcessLongFrameSuccess = 0x40000e54;
lmacProcessShortFrameSuccess = 0x40000e58;
/* lmacDiscardFrameExchangeSequence = 0x40000e5c; */
lmacProcessTBSuccess = 0x40000e60;
/*lmacProcessTxSuccess = 0x40000e64;*/
lmacProcessAckTimeout = 0x40000e68;
/* lmacProcessTxComplete = 0x40000e6c; */
/* ppRemoveHTC = 0x40000e70; */
get_estimated_batime = 0x40000e74;
is_use_muedca = 0x40000e78;
/* hal_mac_tx_clr_mplen = 0x40000e7c; */
/* hal_mac_get_txq_state = 0x40000e80; */
hal_mac_clr_txq_state = 0x40000e84;
hal_mac_get_txq_complete = 0x40000e88;
ht_get_min_subframe_len = 0x40000e8c;
rx11ACRate2AMPDULimit = 0x40000e90;
pwr_hal_clear_intr_status = 0x40000e94;
pwr_hal_clear_mac_modem_beacon_miss_intr_filter = 0x40000e98;
pwr_hal_clear_mac_modem_rx_beacon_info = 0x40000e9c;
pwr_hal_clear_mac_modem_rx_beacon_miss_counter = 0x40000ea0;
pwr_hal_clear_mac_modem_rx_beacon_sleep_counter = 0x40000ea4;
pwr_hal_clear_mac_modem_state_wakeup_protect_signal = 0x40000ea8;
pwr_hal_get_intr_raw_signal = 0x40000eac;
pwr_hal_get_intr_status = 0x40000eb0;
pwr_hal_get_mac_modem_beacon_miss_limit_exceeded_status = 0x40000eb4;
pwr_hal_get_mac_modem_rx_beacon_location_state = 0x40000eb8;
pwr_hal_get_mac_modem_rx_beacon_valid_state = 0x40000ebc;
pwr_hal_get_mac_modem_state_sleep_limit_exceeded_status = 0x40000ec0;
pwr_hal_set_beacon_filter_abort_disable = 0x40000ec4;
pwr_hal_set_beacon_filter_abort_enable = 0x40000ec8;
pwr_hal_set_beacon_filter_abort_length = 0x40000ecc;
/* pwr_hal_set_beacon_filter_broadcast_wakeup_disable = 0x40000ed0; */
/* pwr_hal_set_beacon_filter_broadcast_wakeup_enable = 0x40000ed4; */
pwr_hal_set_beacon_filter_disable = 0x40000ed8;
pwr_hal_set_beacon_filter_enable = 0x40000edc;
pwr_hal_set_beacon_filter_force_dump_disable = 0x40000ee0;
pwr_hal_set_beacon_filter_force_dump_enable = 0x40000ee4;
pwr_hal_set_beacon_filter_force_dump_limit = 0x40000ee8;
pwr_hal_set_beacon_filter_force_sync_disable = 0x40000eec;
pwr_hal_set_beacon_filter_force_sync_enable = 0x40000ef0;
pwr_hal_set_beacon_filter_force_sync_limit = 0x40000ef4;
pwr_hal_set_beacon_filter_frame_crc_state = 0x40000ef8;
pwr_hal_set_beacon_filter_soc_wakeup_and_intr_disable = 0x40000efc;
pwr_hal_set_beacon_filter_soc_wakeup_and_intr_enable = 0x40000f00;
pwr_hal_set_beacon_filter_unicast_wakeup_disable = 0x40000f04;
pwr_hal_set_beacon_filter_unicast_wakeup_enable = 0x40000f08;
pwr_hal_set_lpclk_cycle_time = 0x40000f0c;
pwr_hal_set_lpclk_sync_disable = 0x40000f10;
pwr_hal_set_lpclk_sync_enable = 0x40000f14;
pwr_hal_set_mac_modem_beacon_miss_intr_disable = 0x40000f18;
pwr_hal_set_mac_modem_beacon_miss_intr_enable = 0x40000f1c;
pwr_hal_set_mac_modem_beacon_miss_limit = 0x40000f20;
pwr_hal_set_mac_modem_beacon_miss_limit_exceeded_wakeup_disable = 0x40000f24;
pwr_hal_set_mac_modem_beacon_miss_limit_exceeded_wakeup_enable = 0x40000f28;
pwr_hal_set_mac_modem_beacon_miss_timeout = 0x40000f2c;
pwr_hal_set_mac_modem_state_sleep_limit = 0x40000f30;
pwr_hal_set_mac_modem_state_sleep_limit_exceeded_wakeup_disable = 0x40000f34;
pwr_hal_set_mac_modem_state_sleep_limit_exceeded_wakeup_enable = 0x40000f38;
pwr_hal_set_mac_modem_state_wakeup_protect_disable = 0x40000f3c;
pwr_hal_set_mac_modem_state_wakeup_protect_early_time = 0x40000f40;
pwr_hal_set_mac_modem_state_wakeup_protect_enable = 0x40000f44;
pwr_hal_set_mac_modem_tbtt_auto_period_disable = 0x40000f48;
pwr_hal_set_mac_modem_tbtt_auto_period_enable = 0x40000f4c;
pwr_hal_set_mac_modem_tbtt_auto_period_interval = 0x40000f50;
pwr_hal_set_modem_state_interface = 0x40000f54;
hal_tsf_clear_soc_wakeup_request = 0x40000f58;
tsf_hal_clear_mac_modem_rf_power_state = 0x40000f5c;
tsf_hal_clear_soc_wakeup_request = 0x40000f60;
tsf_hal_get_counter_value = 0x40000f64;
tsf_hal_get_mac_modem_rf_power_state = 0x40000f68;
tsf_hal_get_tbtt_interval = 0x40000f6c;
tsf_hal_get_time = 0x40000f70;
tsf_hal_get_timer_target = 0x40000f74;
tsf_hal_is_tsf_enabled = 0x40000f78;
tsf_hal_map_tbtt_target_to_rx_frame = 0x40000f7c;
tsf_hal_map_tsf_to_bssid = 0x40000f80;
tsf_hal_set_counter_value = 0x40000f84;
tsf_hal_set_modem_wakeup_early_time = 0x40000f88;
tsf_hal_set_rx_beacon_abort_tsf_time_deviation_sync_disable = 0x40000f8c;
tsf_hal_set_rx_beacon_abort_tsf_time_deviation_sync_enable = 0x40000f90;
tsf_hal_set_rx_beacon_fail_tsf_time_deviation_sync_disable = 0x40000f94;
tsf_hal_set_rx_beacon_fail_tsf_time_deviation_sync_enable = 0x40000f98;
tsf_hal_set_rx_beacon_success_tsf_time_deviation_sync_disable = 0x40000f9c;
tsf_hal_set_rx_beacon_success_tsf_time_deviation_sync_enable = 0x40000fa0;
tsf_hal_set_tbtt_disable = 0x40000fa4;
tsf_hal_set_tbtt_early_time = 0x40000fa8;
tsf_hal_set_tbtt_enable = 0x40000fac;
tsf_hal_set_tbtt_interval = 0x40000fb0;
tsf_hal_set_tbtt_intr_disable = 0x40000fb4;
tsf_hal_set_tbtt_intr_enable = 0x40000fb8;
tsf_hal_set_tbtt_modem_wakeup_disable = 0x40000fbc;
tsf_hal_set_tbtt_modem_wakeup_enable = 0x40000fc0;
tsf_hal_set_tbtt_rf_ctrl_disable = 0x40000fc4;
tsf_hal_set_tbtt_rf_ctrl_enable = 0x40000fc8;
tsf_hal_set_tbtt_rf_ctrl_wait_cycles = 0x40000fcc;
tsf_hal_set_tbtt_soc_wakeup_disable = 0x40000fd0;
tsf_hal_set_tbtt_soc_wakeup_enable = 0x40000fd4;
/*tsf_hal_set_tbtt_start_time = 0x40000fd8;*/
tsf_hal_set_time = 0x40000fdc;
tsf_hal_set_timer_disable = 0x40000fe0;
tsf_hal_set_timer_enable = 0x40000fe4;
tsf_hal_set_timer_intr_disable = 0x40000fe8;
tsf_hal_set_timer_intr_enable = 0x40000fec;
tsf_hal_set_timer_modem_wakeup_disable = 0x40000ff0;
tsf_hal_set_timer_modem_wakeup_enable = 0x40000ff4;
tsf_hal_set_timer_rf_ctrl_disable = 0x40000ff8;
tsf_hal_set_timer_rf_ctrl_enable = 0x40000ffc;
tsf_hal_set_timer_rf_ctrl_wait_cycles = 0x40001000;
tsf_hal_set_timer_soc_wakeup_disable = 0x40001004;
tsf_hal_set_timer_soc_wakeup_enable = 0x40001008;
tsf_hal_set_timer_target = 0x4000100c;
tsf_hal_set_tsf_disable = 0x40001010;
tsf_hal_set_tsf_enable = 0x40001014;
tsf_hal_set_tsf_time_deviation = 0x40001018;
tsf_hal_set_tsf_time_deviation_sync_disable = 0x4000101c;
tsf_hal_set_tsf_time_deviation_sync_enable = 0x40001020;
tsf_hal_unmap_tbtt_target_to_rx_frame = 0x40001024;
/* ppSelectTxFormat = 0x40001028; */
/* ppCertSetRate = 0x4000102c; */
/* ppHEAMPDU2Normal = 0x40001030; */
/* ppCalTxHEAMPDULength = 0x40001034; */
/* ppCalTxHESMPDULength = 0x40001038; */
/* rcGetRate = 0x4000103c; */
rcGetDCMMaxRate = 0x40001040;
/* rcGetSMPDURate = 0x40001044; */
ppDirectRecycleAmpdu = 0x40001048;
/* ppCheckTxHEAMPDUlength = 0x4000104c; */
/* rx11AXRate2AMPDULimit = 0x40001050; */
/* ppRegressAmpdu = 0x40001054; */
/* ppCalDeliNum = 0x40001058; */
ppAdd2AMPDUTail = 0x4000105c;
esp_test_disable_tx_statistics = 0x40001060;
esp_test_enable_tx_statistics = 0x40001064;
esp_test_clr_tx_statistics = 0x40001068;
esp_test_get_tx_statistics = 0x4000106c;
esp_test_clr_tx_tb_statistics = 0x40001070;
esp_test_get_tx_tb_statistics = 0x40001074;
test_tx_fail_statistics = 0x40001078;
/* test_tx_succ_statistics = 0x4000107c; */
/* esp_test_tx_process_complete = 0x40001080; */
/* esp_test_tx_process_txq_state = 0x40001084; */
esp_test_tx_enab_statistics = 0x40001088;
esp_test_tx_tb_complete = 0x4000108c;
esp_test_tx_count_retry = 0x40001090;
esp_test_tx_count_collision = 0x40001094;
esp_test_tx_count_timeout = 0x40001098;
hal_enable_tx_statistics = 0x4000109c;
test_rx_process_complete_noeb = 0x400010a0;
test_rx_process_complete_retry = 0x400010a4;
esp_test_rx_process_complete = 0x400010a8;
esp_test_clr_rx_statistics = 0x400010ac;
esp_test_get_rx_statistics = 0x400010b0;
test_free_rx_statistics = 0x400010b4;
esp_test_set_rx_error_occurs = 0x400010b8;
esp_test_get_rx_error_occurs = 0x400010bc;
esp_test_clr_rx_error_occurs = 0x400010c0;
esp_test_disable_rx_statistics = 0x400010c4;
esp_test_enable_rx_statistics = 0x400010c8;
hal_enable_rx_statistics = 0x400010cc;
get_user_num = 0x400010d0;
mumimo_spatial_cfg_get_nsts = 0x400010d4;
mumimo_spatial_cfg_get_nsts_tot = 0x400010d8;
test_mumimo_get_heltf_num = 0x400010dc;
test_mimo_update_user_info = 0x400010e0;
test_parse_rx_mu_mimo = 0x400010e4;
test_nonmimo_update_user_info = 0x400010e8;
test_parse_rx_mu_nonmimo = 0x400010ec;
esp_test_rx_parse_mu = 0x400010f0;
esp_test_get_rx_mu_statistics = 0x400010f4;
esp_test_clr_rx_mu_statistics = 0x400010f8;
esp_test_enable_rx_mu_statistics = 0x400010fc;
esp_test_disable_rx_mu_statistics = 0x40001100;
/* Data (.data, .bss, .rodata) */
our_instances_ptr = 0x4004ffe0;
pTxRx = 0x4087ff80;
lmacConfMib_ptr = 0x4087ff7c;
our_wait_eb = 0x4087ff78;
our_tx_eb = 0x4087ff74;
pp_wdev_funcs = 0x4087ff70;
g_osi_funcs_p = 0x4087ff6c;
wDevCtrl_ptr = 0x4087ff68;
g_wdev_last_desc_reset_ptr = 0x4004ffdc;
wDevMacSleep_ptr = 0x4087ff64;
g_lmac_cnt_ptr = 0x4087ff60;
our_controls_ptr = 0x4004ffd8;
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
/* esp_wifi_cert_tx_ltf = 0x4087fcf4; */
/* esp_wifi_cert_tx_gi = 0x4087fcf0; */
esp_wifi_cert_tx_nss = 0x4087fcec;
esp_test_tx_statistics_aci_bitmap = 0x4087fda4;
esp_test_tx_statistics = 0x4087fd94;
esp_test_tx_tb_statistics = 0x4087fd84;
esp_test_tx_fail_statistics = 0x4087fd24;
esp_test_rx_statistics = 0x4087fd1c;
esp_test_rx_mu_statistics = 0x4087fd18;
esp_test_mu_print_ru_allocation = 0x4087fd14;
sigb_ru_allocation_user_num = 0x4004ffc8;
sigb_common_ru_allocation = 0x4004ff38;
mu_mimo_special_cfg_user_num_2 = 0x4004fee8;
mu_mimo_special_cfg_user_num_3 = 0x4004fe80;
mu_mimo_special_cfg_user_num_4 = 0x4004fe28;
mu_mimo_special_cfg_user_num_5 = 0x4004fdf0;
mu_mimo_special_cfg_user_num_6 = 0x4004fdd0;
mu_mimo_special_cfg_user_num_7 = 0x4004fdc0;
mu_mimo_special_cfg_user_num_8 = 0x4004fdb8;
esp_test_rx_error_occurs = 0x4087fd10;
g_pp_tx_pkt_num = 0x4087fd0c;
he_max_apep_length = 0x4004fd40;
