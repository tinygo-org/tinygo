//go:build tinygo && rp2350

package machine

import (
	"runtime/interrupt"
	"unsafe"
)

/*
typedef unsigned char uint8_t;
typedef unsigned short uint16_t;
typedef unsigned long uint32_t;
typedef unsigned long size_t;
typedef unsigned long uintptr_t;
typedef long int intptr_t;

typedef const volatile uint16_t io_ro_16;
typedef const volatile uint32_t io_ro_32;
typedef volatile uint16_t io_rw_16;
typedef volatile uint32_t io_rw_32;
typedef volatile uint32_t io_wo_32;

#define false 0
#define true 1
typedef int bool;

#define ram_func __attribute__((section(".ramfuncs"),noinline))

typedef void (*flash_exit_xip_fn)(void);
typedef void (*flash_flush_cache_fn)(void);
typedef void (*flash_connect_internal_fn)(void);
typedef void (*flash_range_erase_fn)(uint32_t, size_t, uint32_t, uint16_t);
typedef void (*flash_range_program_fn)(uint32_t, const uint8_t*, size_t);
static inline __attribute__((always_inline)) void __compiler_memory_barrier(void) {
    __asm__ volatile ("" : : : "memory");
}

// https://datasheets.raspberrypi.com/rp2350/rp2350-datasheet.pdf
// 13.9. Predefined OTP Data Locations
// OTP_DATA: FLASH_DEVINFO Register

#define OTP_DATA_FLASH_DEVINFO_CS0_SIZE_BITS 0x0F00
#define OTP_DATA_FLASH_DEVINFO_CS0_SIZE_LSB  8
#define OTP_DATA_FLASH_DEVINFO_CS1_SIZE_BITS 0xF000
#define OTP_DATA_FLASH_DEVINFO_CS1_SIZE_LSB  12


// https://github.com/raspberrypi/pico-sdk
// src/rp2350/hardware_regs/include/hardware/regs/addressmap.h

#define REG_ALIAS_RW_BITS  (0x0 << 12)
#define REG_ALIAS_XOR_BITS (0x1 << 12)
#define REG_ALIAS_SET_BITS (0x2 << 12)
#define REG_ALIAS_CLR_BITS (0x3 << 12)

#define XIP_BASE     0x10000000
#define XIP_QMI_BASE 0x400d0000
#define IO_QSPI_BASE 0x40030000
#define BOOTRAM_BASE 0x400e0000


// https://github.com/raspberrypi/pico-sdk
// src/rp2_common/hardware_base/include/hardware/address_mapped.h

#define hw_alias_check_addr(addr) ((uintptr_t)(addr))
#define hw_set_alias_untyped(addr) ((void *)(REG_ALIAS_SET_BITS + hw_alias_check_addr(addr)))
#define hw_clear_alias_untyped(addr) ((void *)(REG_ALIAS_CLR_BITS + hw_alias_check_addr(addr)))
#define hw_xor_alias_untyped(addr) ((void *)(REG_ALIAS_XOR_BITS + hw_alias_check_addr(addr)))

__attribute__((always_inline))
static void hw_set_bits(io_rw_32 *addr, uint32_t mask) {
    *(io_rw_32 *) hw_set_alias_untyped((volatile void *) addr) = mask;
}

__attribute__((always_inline))
static void hw_clear_bits(io_rw_32 *addr, uint32_t mask) {
    *(io_rw_32 *) hw_clear_alias_untyped((volatile void *) addr) = mask;
}

__attribute__((always_inline))
static void hw_xor_bits(io_rw_32 *addr, uint32_t mask) {
    *(io_rw_32 *) hw_xor_alias_untyped((volatile void *) addr) = mask;
}

__attribute__((always_inline))
static void hw_write_masked(io_rw_32 *addr, uint32_t values, uint32_t write_mask) {
    hw_xor_bits(addr, (*addr ^ values) & write_mask);
}


// https://github.com/raspberrypi/pico-sdk
// src/rp2_common/pico_platform_compiler/include/pico/platform/compiler.h

#define pico_default_asm_volatile(...) __asm volatile (".syntax unified\n" __VA_ARGS__)


// https://github.com/raspberrypi/pico-sdk
// src/rp2350/pico_platform/include/pico/platform.h

static bool pico_processor_state_is_nonsecure(void) {
//    // todo add a define to disable NS checking at all?
//    // IDAU-Exempt addresses return S=1 when tested in the Secure state,
//    // whereas executing a tt in the NonSecure state will always return S=0.
//    uint32_t tt;
//    pico_default_asm_volatile (
//        "movs %0, #0\n"
//        "tt %0, %0\n"
//        : "=r" (tt) : : "cc"
//    );
//    return !(tt & (1u << 22));

    return false;
}


// https://github.com/raspberrypi/pico-sdk
// src/rp2_common/pico_bootrom/include/pico/bootrom_constants.h

// RP2040 & RP2350
#define ROM_DATA_SOFTWARE_GIT_REVISION          ROM_TABLE_CODE('G', 'R')
#define ROM_FUNC_FLASH_ENTER_CMD_XIP            ROM_TABLE_CODE('C', 'X')
#define ROM_FUNC_FLASH_EXIT_XIP                 ROM_TABLE_CODE('E', 'X')
#define ROM_FUNC_FLASH_FLUSH_CACHE              ROM_TABLE_CODE('F', 'C')
#define ROM_FUNC_CONNECT_INTERNAL_FLASH         ROM_TABLE_CODE('I', 'F')
#define ROM_FUNC_FLASH_RANGE_ERASE              ROM_TABLE_CODE('R', 'E')
#define ROM_FUNC_FLASH_RANGE_PROGRAM            ROM_TABLE_CODE('R', 'P')

// RP2350 only
#define ROM_FUNC_PICK_AB_PARTITION              ROM_TABLE_CODE('A', 'B')
#define ROM_FUNC_CHAIN_IMAGE                    ROM_TABLE_CODE('C', 'I')
#define ROM_FUNC_EXPLICIT_BUY                   ROM_TABLE_CODE('E', 'B')
#define ROM_FUNC_FLASH_RUNTIME_TO_STORAGE_ADDR  ROM_TABLE_CODE('F', 'A')
#define ROM_DATA_FLASH_DEVINFO16_PTR            ROM_TABLE_CODE('F', 'D')
#define ROM_FUNC_FLASH_OP                       ROM_TABLE_CODE('F', 'O')
#define ROM_FUNC_GET_B_PARTITION                ROM_TABLE_CODE('G', 'B')
#define ROM_FUNC_GET_PARTITION_TABLE_INFO       ROM_TABLE_CODE('G', 'P')
#define ROM_FUNC_GET_SYS_INFO                   ROM_TABLE_CODE('G', 'S')
#define ROM_FUNC_GET_UF2_TARGET_PARTITION       ROM_TABLE_CODE('G', 'U')
#define ROM_FUNC_LOAD_PARTITION_TABLE           ROM_TABLE_CODE('L', 'P')
#define ROM_FUNC_OTP_ACCESS                     ROM_TABLE_CODE('O', 'A')
#define ROM_DATA_PARTITION_TABLE_PTR            ROM_TABLE_CODE('P', 'T')
#define ROM_FUNC_FLASH_RESET_ADDRESS_TRANS      ROM_TABLE_CODE('R', 'A')
#define ROM_FUNC_REBOOT                         ROM_TABLE_CODE('R', 'B')
#define ROM_FUNC_SET_ROM_CALLBACK               ROM_TABLE_CODE('R', 'C')
#define ROM_FUNC_SECURE_CALL                    ROM_TABLE_CODE('S', 'C')
#define ROM_FUNC_SET_NS_API_PERMISSION          ROM_TABLE_CODE('S', 'P')
#define ROM_FUNC_BOOTROM_STATE_RESET            ROM_TABLE_CODE('S', 'R')
#define ROM_FUNC_SET_BOOTROM_STACK              ROM_TABLE_CODE('S', 'S')
#define ROM_DATA_SAVED_XIP_SETUP_FUNC_PTR       ROM_TABLE_CODE('X', 'F')
#define ROM_FUNC_FLASH_SELECT_XIP_READ_MODE     ROM_TABLE_CODE('X', 'M')
#define ROM_FUNC_VALIDATE_NS_BUFFER             ROM_TABLE_CODE('V', 'B')

#define BOOTSEL_FLAG_GPIO_PIN_SPECIFIED         0x20

#define BOOTROM_FUNC_TABLE_OFFSET 0x14

// todo remove this (or #ifdef it for A1/A2)
#define BOOTROM_IS_A2() ((*(volatile uint8_t *)0x13) == 2)
#define BOOTROM_WELL_KNOWN_PTR_SIZE (BOOTROM_IS_A2() ? 2 : 4)

#define BOOTROM_VTABLE_OFFSET 0x00
#define BOOTROM_TABLE_LOOKUP_OFFSET     (BOOTROM_FUNC_TABLE_OFFSET + BOOTROM_WELL_KNOWN_PTR_SIZE)


// https://github.com/raspberrypi/pico-sdk
// src/common/boot_picoboot_headers/include/boot/picoboot_constants.h

// values 0-7 are secure/non-secure
#define REBOOT2_FLAG_REBOOT_TYPE_NORMAL       0x0 // param0 = diagnostic partition
#define REBOOT2_FLAG_REBOOT_TYPE_BOOTSEL      0x2 // param0 = bootsel_flags, param1 = gpio_config
#define REBOOT2_FLAG_REBOOT_TYPE_RAM_IMAGE    0x3 // param0 = image_base, param1 = image_end
#define REBOOT2_FLAG_REBOOT_TYPE_FLASH_UPDATE 0x4 // param0 = update_base

#define REBOOT2_FLAG_NO_RETURN_ON_SUCCESS    0x100

#define RT_FLAG_FUNC_ARM_SEC    0x0004
#define RT_FLAG_FUNC_ARM_NONSEC 0x0010
#define RT_FLAG_DATA            0x0040


// https://github.com/raspberrypi/pico-sdk
// src/rp2_common/pico_bootrom/include/pico/bootrom.h

#define ROM_TABLE_CODE(c1, c2) ((c1) | ((c2) << 8))

typedef void *(*rom_table_lookup_fn)(uint32_t code, uint32_t mask);

__attribute__((always_inline))
static void *rom_func_lookup_inline(uint32_t code) {
    rom_table_lookup_fn rom_table_lookup = (rom_table_lookup_fn) (uintptr_t)*(uint16_t*)(BOOTROM_TABLE_LOOKUP_OFFSET);
    if (pico_processor_state_is_nonsecure()) {
        return rom_table_lookup(code, RT_FLAG_FUNC_ARM_NONSEC);
    } else {
        return rom_table_lookup(code, RT_FLAG_FUNC_ARM_SEC);
    }
}

__attribute__((always_inline))
static void *rom_data_lookup_inline(uint32_t code) {
    rom_table_lookup_fn rom_table_lookup = (rom_table_lookup_fn) (uintptr_t)*(uint16_t*)(BOOTROM_TABLE_LOOKUP_OFFSET);
    return rom_table_lookup(code, RT_FLAG_DATA);
}

typedef int (*rom_reboot_fn)(uint32_t flags, uint32_t delay_ms, uint32_t p0, uint32_t p1);

__attribute__((always_inline))
int rom_reboot(uint32_t flags, uint32_t delay_ms, uint32_t p0, uint32_t p1) {
    rom_reboot_fn func = (rom_reboot_fn) rom_func_lookup_inline(ROM_FUNC_REBOOT);
    return func(flags, delay_ms, p0, p1);
}


// https://github.com/raspberrypi/pico-sdk
// src/rp2_common/pico_bootrom/bootrom.c

void reset_usb_boot(uint32_t usb_activity_gpio_pin_mask, uint32_t disable_interface_mask) {
    uint32_t flags = disable_interface_mask;
    if (usb_activity_gpio_pin_mask) {
        flags |= BOOTSEL_FLAG_GPIO_PIN_SPECIFIED;
        // the parameter is actually the gpio number, but we only care if BOOTSEL_FLAG_GPIO_PIN_SPECIFIED
        usb_activity_gpio_pin_mask = (uint32_t)__builtin_ctz(usb_activity_gpio_pin_mask);
    }
    rom_reboot(REBOOT2_FLAG_REBOOT_TYPE_BOOTSEL | REBOOT2_FLAG_NO_RETURN_ON_SUCCESS, 10, flags, usb_activity_gpio_pin_mask);
    __builtin_unreachable();
}


// https://github.com/raspberrypi/pico-sdk
// src/rp2350/hardware_regs/include/hardware/regs/qmi.h

#define QMI_DIRECT_CSR_EN_BITS      0x00000001
#define QMI_DIRECT_CSR_RXEMPTY_BITS 0x00010000
#define QMI_DIRECT_CSR_TXFULL_BITS  0x00000400
#define QMI_M1_WFMT_RESET           0x00001000
#define QMI_M1_WCMD_RESET           0x0000a002


// https://github.com/raspberrypi/pico-sdk
// src/rp2350/hardware_regs/include/hardware/regs/io_qspi.h

#define IO_QSPI_GPIO_QSPI_SS_CTRL_OUTOVER_BITS       0x00003000
#define IO_QSPI_GPIO_QSPI_SS_CTRL_OUTOVER_LSB        12
#define IO_QSPI_GPIO_QSPI_SS_CTRL_OUTOVER_VALUE_LOW  0x2
#define IO_QSPI_GPIO_QSPI_SS_CTRL_OUTOVER_VALUE_HIGH 0x3


// https://github.com/raspberrypi/pico-sdk
// src/rp2350/hardware_structs/include/hardware/structs/io_qspi.h

typedef struct {
    io_rw_32 inte; // IO_QSPI_PROC0_INTE
    io_rw_32 intf; // IO_QSPI_PROC0_INTF
    io_ro_32 ints; // IO_QSPI_PROC0_INTS
} io_qspi_irq_ctrl_hw_t;

typedef struct {
    io_ro_32 status; // IO_QSPI_GPIO_QSPI_SCLK_STATUS
    io_rw_32 ctrl;   // IO_QSPI_GPIO_QSPI_SCLK_CTRL
} io_qspi_status_ctrl_hw_t;

typedef struct {
    io_ro_32 usbphy_dp_status;                  // IO_QSPI_USBPHY_DP_STATUS
    io_rw_32 usbphy_dp_ctrl;                    // IO_QSPI_USBPHY_DP_CTRL
    io_ro_32 usbphy_dm_status;                  // IO_QSPI_USBPHY_DM_STATUS
    io_rw_32 usbphy_dm_ctrl;                    // IO_QSPI_USBPHY_DM_CTRL
    io_qspi_status_ctrl_hw_t io[6];
    uint32_t _pad0[112];
    io_ro_32 irqsummary_proc0_secure;           // IO_QSPI_IRQSUMMARY_PROC0_SECURE
    io_ro_32 irqsummary_proc0_nonsecure;        // IO_QSPI_IRQSUMMARY_PROC0_NONSECURE
    io_ro_32 irqsummary_proc1_secure;           // IO_QSPI_IRQSUMMARY_PROC1_SECURE
    io_ro_32 irqsummary_proc1_nonsecure;        // IO_QSPI_IRQSUMMARY_PROC1_NONSECURE
    io_ro_32 irqsummary_dormant_wake_secure;    // IO_QSPI_IRQSUMMARY_DORMANT_WAKE_SECURE
    io_ro_32 irqsummary_dormant_wake_nonsecure; // IO_QSPI_IRQSUMMARY_DORMANT_WAKE_NONSECURE
    io_rw_32 intr;                              // IO_QSPI_INTR

    union {
        struct {
            io_qspi_irq_ctrl_hw_t proc0_irq_ctrl;
            io_qspi_irq_ctrl_hw_t proc1_irq_ctrl;
            io_qspi_irq_ctrl_hw_t dormant_wake_irq_ctrl;
        };
        io_qspi_irq_ctrl_hw_t irq_ctrl[3];
    };
} io_qspi_hw_t;

#define io_qspi_hw ((io_qspi_hw_t *)IO_QSPI_BASE)


// https://github.com/raspberrypi/pico-sdk
// src/rp2350/hardware_structs/include/hardware/structs/qmi.h

typedef struct {
    io_rw_32 timing; // QMI_M0_TIMING
    io_rw_32 rfmt;   // QMI_M0_RFMT
    io_rw_32 rcmd;   // QMI_M0_RCMD
    io_rw_32 wfmt;   // QMI_M0_WFMT
    io_rw_32 wcmd;   // QMI_M0_WCMD
} qmi_mem_hw_t;

typedef struct {
    io_rw_32 direct_csr; // QMI_DIRECT_CSR
    io_wo_32 direct_tx;  // QMI_DIRECT_TX
    io_ro_32 direct_rx;  // QMI_DIRECT_RX
    qmi_mem_hw_t m[2];
    io_rw_32 atrans[8];  // QMI_ATRANS0
} qmi_hw_t;

#define qmi_hw ((qmi_hw_t *)XIP_QMI_BASE)


// https://github.com/raspberrypi/pico-sdk
// src/rp2_common/hardware_xip_cache/include/hardware/xip_cache.h

// Noop unless using XIP Cache-as-SRAM
// Non-noop version in src/rp2_common/hardware_xip_cache/xip_cache.c
static inline void xip_cache_clean_all(void) {}


// https://github.com/raspberrypi/pico-sdk
// src/rp2_common/hardware_flash/include/hardware/flash.h

#define FLASH_PAGE_SIZE (1u << 8)
#define FLASH_SECTOR_SIZE (1u << 12)
#define FLASH_BLOCK_SIZE (1u << 16)


// https://github.com/raspberrypi/pico-sdk
// src/rp2_common/hardware_flash/flash.c

#define BOOT2_SIZE_WORDS 64
#define FLASH_BLOCK_ERASE_CMD 0xd8

static uint32_t boot2_copyout[BOOT2_SIZE_WORDS];
static bool boot2_copyout_valid = false;

static ram_func void flash_init_boot2_copyout(void) {
    if (boot2_copyout_valid)
        return;
    for (int i = 0; i < BOOT2_SIZE_WORDS; ++i)
		boot2_copyout[i] = ((uint32_t *)BOOTRAM_BASE)[i];
    __compiler_memory_barrier();
    boot2_copyout_valid = true;
}

static ram_func void flash_enable_xip_via_boot2(void) {
    ((void (*)(void))((intptr_t)boot2_copyout+1))();
}

// This is a static symbol because the layout of FLASH_DEVINFO is liable to change from device to
// device, so fields must have getters/setters.
static io_rw_16 * ram_func flash_devinfo_ptr(void) {
    // Note the lookup returns a pointer to a 32-bit pointer literal in the ROM
    io_rw_16 **p = (io_rw_16 **) rom_data_lookup_inline(ROM_DATA_FLASH_DEVINFO16_PTR);
    return *p;
}

// This is a RAM function because may be called during flash programming to enable save/restore of
// QMI window 1 registers on RP2350:
uint8_t ram_func flash_devinfo_get_cs_size(uint8_t cs) {
    io_ro_16 *devinfo = (io_ro_16 *) flash_devinfo_ptr();
    if (cs == 0u) {
        return (uint8_t) (
            (*devinfo & OTP_DATA_FLASH_DEVINFO_CS0_SIZE_BITS) >> OTP_DATA_FLASH_DEVINFO_CS0_SIZE_LSB
        );
    } else {
        return (uint8_t) (
            (*devinfo & OTP_DATA_FLASH_DEVINFO_CS1_SIZE_BITS) >> OTP_DATA_FLASH_DEVINFO_CS1_SIZE_LSB
        );
    }
}

// This is specifically for saving/restoring the registers modified by RP2350
// flash_exit_xip() ROM func, not the entirety of the QMI window state.
typedef struct flash_rp2350_qmi_save_state {
    uint32_t timing;
    uint32_t rcmd;
    uint32_t rfmt;
} flash_rp2350_qmi_save_state_t;

static ram_func void flash_rp2350_save_qmi_cs1(flash_rp2350_qmi_save_state_t *state) {
    state->timing = qmi_hw->m[1].timing;
    state->rcmd = qmi_hw->m[1].rcmd;
    state->rfmt = qmi_hw->m[1].rfmt;
}

static ram_func void flash_rp2350_restore_qmi_cs1(const flash_rp2350_qmi_save_state_t *state) {
    if (flash_devinfo_get_cs_size(1) == 0) {
        // Case 1: The RP2350 ROM sets QMI to a clean (03h read) configuration
        // during flash_exit_xip(), even though when CS1 is not enabled via
        // FLASH_DEVINFO it does not issue an XIP exit sequence to CS1. In
        // this case, restore the original register config for CS1 as it is
        // still the correct config.
        qmi_hw->m[1].timing = state->timing;
        qmi_hw->m[1].rcmd = state->rcmd;
        qmi_hw->m[1].rfmt = state->rfmt;
    } else {
        // Case 2: If RAM is attached to CS1, and the ROM has issued an XIP
        // exit sequence to it, then the ROM re-initialisation of the QMI
        // registers has actually not gone far enough. The old XIP write mode
        // is no longer valid when the QSPI RAM is returned to a serial
        // command state. Restore the default 02h serial write command config.
        qmi_hw->m[1].wfmt = QMI_M1_WFMT_RESET;
        qmi_hw->m[1].wcmd = QMI_M1_WCMD_RESET;
    }
}

void ram_func flash_cs_force(bool high) {
    uint32_t field_val = high ?
        IO_QSPI_GPIO_QSPI_SS_CTRL_OUTOVER_VALUE_HIGH :
        IO_QSPI_GPIO_QSPI_SS_CTRL_OUTOVER_VALUE_LOW;
    hw_write_masked(&io_qspi_hw->io[1].ctrl,
        field_val << IO_QSPI_GPIO_QSPI_SS_CTRL_OUTOVER_LSB,
        IO_QSPI_GPIO_QSPI_SS_CTRL_OUTOVER_BITS
    );
}

// Adapted from flash_range_program()
void ram_func flash_range_write(uint32_t offset, const uint8_t *data, size_t count) {
    flash_connect_internal_fn flash_connect_internal_func = (flash_connect_internal_fn)rom_func_lookup_inline(ROM_FUNC_CONNECT_INTERNAL_FLASH);
    flash_exit_xip_fn flash_exit_xip_func = (flash_exit_xip_fn)rom_func_lookup_inline(ROM_FUNC_FLASH_EXIT_XIP);
    flash_range_program_fn flash_range_program_func = (flash_range_program_fn)rom_func_lookup_inline(ROM_FUNC_FLASH_RANGE_PROGRAM);
    flash_flush_cache_fn flash_flush_cache_func = (flash_flush_cache_fn)rom_func_lookup_inline(ROM_FUNC_FLASH_FLUSH_CACHE);
    flash_init_boot2_copyout();
    xip_cache_clean_all();
    flash_rp2350_qmi_save_state_t qmi_save;
    flash_rp2350_save_qmi_cs1(&qmi_save);

    __compiler_memory_barrier();

    flash_connect_internal_func();
    flash_exit_xip_func();
    flash_range_program_func(offset, data, count);
    flash_flush_cache_func(); // Note this is needed to remove CSn IO force as well as cache flushing
    flash_enable_xip_via_boot2();
    flash_rp2350_restore_qmi_cs1(&qmi_save);
}

// Adapted from flash_range_erase()
void ram_func flash_erase_blocks(uint32_t offset, size_t count) {
    flash_connect_internal_fn flash_connect_internal_func = (flash_connect_internal_fn)rom_func_lookup_inline(ROM_FUNC_CONNECT_INTERNAL_FLASH);
    flash_exit_xip_fn flash_exit_xip_func = (flash_exit_xip_fn)rom_func_lookup_inline(ROM_FUNC_FLASH_EXIT_XIP);
    flash_range_erase_fn flash_range_erase_func = (flash_range_erase_fn)rom_func_lookup_inline(ROM_FUNC_FLASH_RANGE_ERASE);
    flash_flush_cache_fn flash_flush_cache_func = (flash_flush_cache_fn)rom_func_lookup_inline(ROM_FUNC_FLASH_FLUSH_CACHE);
    flash_init_boot2_copyout();
    // Commit any pending writes to external RAM, to avoid losing them in the subsequent flush:
    xip_cache_clean_all();
    flash_rp2350_qmi_save_state_t qmi_save;
    flash_rp2350_save_qmi_cs1(&qmi_save);

    // No flash accesses after this point
    __compiler_memory_barrier();

    flash_connect_internal_func();
    flash_exit_xip_func();
    flash_range_erase_func(offset, count, FLASH_BLOCK_SIZE, FLASH_BLOCK_ERASE_CMD);
    flash_flush_cache_func(); // Note this is needed to remove CSn IO force as well as cache flushing
    flash_enable_xip_via_boot2();
    flash_rp2350_restore_qmi_cs1(&qmi_save);
}

void ram_func flash_do_cmd(const uint8_t *txbuf, uint8_t *rxbuf, size_t count) {
    flash_connect_internal_fn flash_connect_internal_func = (flash_connect_internal_fn)rom_func_lookup_inline(ROM_FUNC_CONNECT_INTERNAL_FLASH);
    flash_exit_xip_fn flash_exit_xip_func = (flash_exit_xip_fn)rom_func_lookup_inline(ROM_FUNC_FLASH_EXIT_XIP);
    flash_flush_cache_fn flash_flush_cache_func = (flash_flush_cache_fn)rom_func_lookup_inline(ROM_FUNC_FLASH_FLUSH_CACHE);
    flash_init_boot2_copyout();
    xip_cache_clean_all();

    flash_rp2350_qmi_save_state_t qmi_save;
    flash_rp2350_save_qmi_cs1(&qmi_save);

    __compiler_memory_barrier();
    flash_connect_internal_func();
    flash_exit_xip_func();

    flash_cs_force(0);
    size_t tx_remaining = count;
    size_t rx_remaining = count;

    // QMI version -- no need to bound FIFO contents as QMI stalls on full DIRECT_RX.
    hw_set_bits(&qmi_hw->direct_csr, QMI_DIRECT_CSR_EN_BITS);
    while (tx_remaining || rx_remaining) {
        uint32_t flags = qmi_hw->direct_csr;
        bool can_put = !(flags & QMI_DIRECT_CSR_TXFULL_BITS);
        bool can_get = !(flags & QMI_DIRECT_CSR_RXEMPTY_BITS);
        if (can_put && tx_remaining) {
            qmi_hw->direct_tx = *txbuf++;
            --tx_remaining;
        }
        if (can_get && rx_remaining) {
            *rxbuf++ = (uint8_t)qmi_hw->direct_rx;
            --rx_remaining;
        }
    }
    hw_clear_bits(&qmi_hw->direct_csr, QMI_DIRECT_CSR_EN_BITS);

    flash_cs_force(1);

    flash_flush_cache_func();
    flash_enable_xip_via_boot2();
    flash_rp2350_restore_qmi_cs1(&qmi_save);
}

*/
import "C"

func enterBootloader() {
	C.reset_usb_boot(0, 0)
}

func doFlashCommand(tx []byte, rx []byte) error {
	if len(tx) != len(rx) {
		return errFlashInvalidWriteLength
	}

	C.flash_do_cmd(
		(*C.uint8_t)(unsafe.Pointer(&tx[0])),
		(*C.uint8_t)(unsafe.Pointer(&rx[0])),
		C.ulong(len(tx)))

	return nil
}

// Flash related code
const memoryStart = C.XIP_BASE // memory start for purpose of erase

func (f flashBlockDevice) writeAt(p []byte, off int64) (n int, err error) {
	if writeAddress(off)+uintptr(C.XIP_BASE) > FlashDataEnd() {
		return 0, errFlashCannotWritePastEOF
	}

	state := interrupt.Disable()
	defer interrupt.Restore(state)

	// rp2350 writes to offset, not actual address
	// e.g. real address 0x10003000 is written to at
	// 0x00003000
	address := writeAddress(off)
	padded := flashPad(p, int(f.WriteBlockSize()))

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
