//go:build rp2040
// +build rp2040

package machine

/*
// https://github.com/raspberrypi/pico-sdk
// src/rp2_common/pico_bootrom/include/pico/bootrom.h

#define ROM_FUNC_POPCOUNT32             ROM_TABLE_CODE('P', '3')
#define ROM_FUNC_REVERSE32              ROM_TABLE_CODE('R', '3')
#define ROM_FUNC_CLZ32                  ROM_TABLE_CODE('L', '3')
#define ROM_FUNC_CTZ32                  ROM_TABLE_CODE('T', '3')
#define ROM_FUNC_MEMSET                 ROM_TABLE_CODE('M', 'S')
#define ROM_FUNC_MEMSET4                ROM_TABLE_CODE('S', '4')
#define ROM_FUNC_MEMCPY                 ROM_TABLE_CODE('M', 'C')
#define ROM_FUNC_MEMCPY44               ROM_TABLE_CODE('C', '4')
#define ROM_FUNC_RESET_USB_BOOT         ROM_TABLE_CODE('U', 'B')
#define ROM_FUNC_CONNECT_INTERNAL_FLASH ROM_TABLE_CODE('I', 'F')
#define ROM_FUNC_FLASH_EXIT_XIP         ROM_TABLE_CODE('E', 'X')
#define ROM_FUNC_FLASH_RANGE_ERASE      ROM_TABLE_CODE('R', 'E')
#define ROM_FUNC_FLASH_RANGE_PROGRAM    ROM_TABLE_CODE('R', 'P')
#define ROM_FUNC_FLASH_FLUSH_CACHE      ROM_TABLE_CODE('F', 'C')
#define ROM_FUNC_FLASH_ENTER_CMD_XIP    ROM_TABLE_CODE('C', 'X')

#define ROM_TABLE_CODE(c1, c2) ((c1) | ((c2) << 8))

typedef unsigned short uint16_t;
typedef unsigned long uint32_t;
typedef unsigned long uintptr_t;

typedef void *(*rom_table_lookup_fn)(uint16_t *table, uint32_t code);
typedef void __attribute__((noreturn)) (*rom_reset_usb_boot_fn)(uint32_t, uint32_t);
#define rom_hword_as_ptr(rom_address) (void *)(uintptr_t)(*(uint16_t *)(uintptr_t)(rom_address))

void *rom_func_lookup(uint32_t code) {
  rom_table_lookup_fn rom_table_lookup = (rom_table_lookup_fn) rom_hword_as_ptr(0x18);
  uint16_t *func_table = (uint16_t *) rom_hword_as_ptr(0x14);
  return rom_table_lookup(func_table, code);
}

void reset_usb_boot(uint32_t usb_activity_gpio_pin_mask, uint32_t disable_interface_mask) {
    rom_reset_usb_boot_fn func = (rom_reset_usb_boot_fn) rom_func_lookup(ROM_FUNC_RESET_USB_BOOT);
    func(usb_activity_gpio_pin_mask, disable_interface_mask);
}
*/
import "C"

// EnterBootloader should perform a system reset in preparation
// to switch to the bootloader to flash new firmware.
func EnterBootloader() {
	C.reset_usb_boot(0, 0)
}
