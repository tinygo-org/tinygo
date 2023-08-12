//go:build rp2040

package machine

import (
	"runtime/interrupt"
	"unsafe"
)

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

typedef unsigned char uint8_t;
typedef unsigned short uint16_t;
typedef unsigned long uint32_t;
typedef unsigned long size_t;
typedef unsigned long uintptr_t;

#define false 0
#define true 1
typedef int bool;

#define ram_func __attribute__((section(".ramfuncs"),noinline))

typedef void *(*rom_table_lookup_fn)(uint16_t *table, uint32_t code);
typedef void __attribute__((noreturn)) (*rom_reset_usb_boot_fn)(uint32_t, uint32_t);
typedef void (*flash_init_boot2_copyout_fn)(void);
typedef void (*flash_enable_xip_via_boot2_fn)(void);
typedef void (*flash_exit_xip_fn)(void);
typedef void (*flash_flush_cache_fn)(void);
typedef void (*flash_connect_internal_fn)(void);
typedef void (*flash_range_erase_fn)(uint32_t, size_t, uint32_t, uint16_t);
typedef void (*flash_range_program_fn)(uint32_t, const uint8_t*, size_t);

static inline __attribute__((always_inline)) void __compiler_memory_barrier(void) {
    __asm__ volatile ("" : : : "memory");
}

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

#define FLASH_BLOCK_ERASE_CMD 0xd8

#define FLASH_PAGE_SIZE (1u << 8)
#define FLASH_SECTOR_SIZE (1u << 12)
#define FLASH_BLOCK_SIZE (1u << 16)

#define BOOT2_SIZE_WORDS 64
#define XIP_BASE 0x10000000

static uint32_t boot2_copyout[BOOT2_SIZE_WORDS];
static bool boot2_copyout_valid = false;

static ram_func void flash_init_boot2_copyout() {
	if (boot2_copyout_valid)
			return;
	for (int i = 0; i < BOOT2_SIZE_WORDS; ++i)
			boot2_copyout[i] = ((uint32_t *)XIP_BASE)[i];
	__compiler_memory_barrier();
	boot2_copyout_valid = true;
}

static ram_func void flash_enable_xip_via_boot2() {
	((void (*)(void))boot2_copyout+1)();
}

// See https://github.com/raspberrypi/pico-sdk/blob/master/src/rp2_common/hardware_flash/flash.c#L86
void ram_func flash_range_write(uint32_t offset, const uint8_t *data, size_t count)
{
	flash_range_program_fn flash_range_program_func = (flash_range_program_fn) rom_func_lookup(ROM_FUNC_FLASH_RANGE_PROGRAM);
	flash_connect_internal_fn flash_connect_internal_func = (flash_connect_internal_fn) rom_func_lookup(ROM_FUNC_CONNECT_INTERNAL_FLASH);
	flash_exit_xip_fn flash_exit_xip_func = (flash_exit_xip_fn) rom_func_lookup(ROM_FUNC_FLASH_EXIT_XIP);
	flash_flush_cache_fn flash_flush_cache_func = (flash_flush_cache_fn) rom_func_lookup(ROM_FUNC_FLASH_FLUSH_CACHE);

	flash_init_boot2_copyout();

	__compiler_memory_barrier();

	flash_connect_internal_func();
	flash_exit_xip_func();

	flash_range_program_func(offset, data, count);
	flash_flush_cache_func();
	flash_enable_xip_via_boot2();
}

void ram_func flash_erase_blocks(uint32_t offset, size_t count)
{
	flash_range_erase_fn flash_range_erase_func = (flash_range_erase_fn) rom_func_lookup(ROM_FUNC_FLASH_RANGE_ERASE);
	flash_connect_internal_fn flash_connect_internal_func = (flash_connect_internal_fn) rom_func_lookup(ROM_FUNC_CONNECT_INTERNAL_FLASH);
	flash_exit_xip_fn flash_exit_xip_func = (flash_exit_xip_fn) rom_func_lookup(ROM_FUNC_FLASH_EXIT_XIP);
	flash_flush_cache_fn flash_flush_cache_func = (flash_flush_cache_fn) rom_func_lookup(ROM_FUNC_FLASH_FLUSH_CACHE);

	flash_init_boot2_copyout();

	__compiler_memory_barrier();

	flash_connect_internal_func();
	flash_exit_xip_func();

	flash_range_erase_func(offset, count, FLASH_BLOCK_SIZE, FLASH_BLOCK_ERASE_CMD);
	flash_flush_cache_func();
	flash_enable_xip_via_boot2();
}

*/
import "C"

func enterBootloader() {
	C.reset_usb_boot(0, 0)
}

// Flash related code
const memoryStart = C.XIP_BASE // memory start for purpose of erase

func (f flashBlockDevice) writeAt(p []byte, off int64) (n int, err error) {
	if writeAddress(off)+uintptr(C.XIP_BASE) > FlashDataEnd() {
		return 0, errFlashCannotWritePastEOF
	}

	state := interrupt.Disable()
	defer interrupt.Restore(state)

	// rp2040 writes to offset, not actual address
	// e.g. real address 0x10003000 is written to at
	// 0x00003000
	address := writeAddress(off)
	padded := f.pad(p)

	C.flash_range_write(C.uint32_t(address),
		(*C.uint8_t)(unsafe.Pointer(&padded[0])),
		C.ulong(len(padded)))

	return len(padded), nil
}

func (f flashBlockDevice) eraseBlocks(start, length int64) error {
	address := writeAddress(start * f.EraseBlockSize())
	if address+uintptr(C.XIP_BASE) > FlashDataEnd() {
		return errFlashCannotErasePastEOF
	}

	state := interrupt.Disable()
	defer interrupt.Restore(state)

	C.flash_erase_blocks(C.uint32_t(address), C.ulong(length*f.EraseBlockSize()))

	return nil
}
