//go:build tkey

// Hand written file based on https://github.com/tillitis/tkey-libs/blob/main/include/tkey/tk1_mem.h

package tkey

import (
	"runtime/volatile"
	"unsafe"
)

// Peripherals
var (
	TRNG = (*TRNG_Type)(unsafe.Pointer(TK1_MMIO_TRNG_BASE))

	TIMER = (*TIMER_Type)(unsafe.Pointer(TK1_MMIO_TIMER_BASE))

	UDS = (*UDS_Type)(unsafe.Pointer(TK1_MMIO_UDS_BASE))

	UART = (*UART_Type)(unsafe.Pointer(TK1_MMIO_UART_BASE))

	TOUCH = (*TOUCH_Type)(unsafe.Pointer(TK1_MMIO_TOUCH_BASE))

	TK1 = (*TK1_Type)(unsafe.Pointer(TK1_MMIO_TK1_BASE))
)

// Memory sections
const (
	TK1_ROM_BASE uintptr = 0x00000000

	TK1_RAM_BASE uintptr = 0x40000000

	TK1_MMIO_BASE uintptr = 0xc0000000

	TK1_MMIO_TRNG_BASE uintptr = 0xc0000000

	TK1_MMIO_TIMER_BASE uintptr = 0xc1000000

	TK1_MMIO_UDS_BASE uintptr = 0xc2000000

	TK1_MMIO_UART_BASE uintptr = 0xc3000000

	TK1_MMIO_TOUCH_BASE uintptr = 0xc4000000

	TK1_MMIO_FW_RAM_BASE uintptr = 0xd0000000

	TK1_MMIO_TK1_BASE uintptr = 0xff000000
)

// Memory section sizes
const (
	TK1_RAM_SIZE uintptr = 0x20000

	TK1_MMIO_SIZE uintptr = 0x3fffffff
)

type TRNG_Type struct {
	_       [36]byte
	STATUS  volatile.Register32
	_       [88]byte
	ENTROPY volatile.Register32
}

type TIMER_Type struct {
	_         [32]byte
	CTRL      volatile.Register32
	STATUS    volatile.Register32
	PRESCALER volatile.Register32
	TIMER     volatile.Register32
}

type UDS_Type struct {
	_    [64]byte
	DATA [8]volatile.Register32
}

type UART_Type struct {
	_         [128]byte
	RX_STATUS volatile.Register32
	RX_DATA   volatile.Register32
	RX_BYTES  volatile.Register32
	_         [116]byte
	TX_STATUS volatile.Register32
	TX_DATA   volatile.Register32
}

type TOUCH_Type struct {
	_      [36]byte
	STATUS volatile.Register32
}

type TK1_Type struct {
	NAME0         volatile.Register32
	NAME1         volatile.Register32
	VERSION       volatile.Register32
	_             [16]byte
	SWITCH_APP    volatile.Register32
	_             [4]byte
	LED           volatile.Register32
	GPIO          volatile.Register16
	APP_ADDR      volatile.Register32
	APP_SIZE      volatile.Register32
	BLAKE2S       volatile.Register32
	_             [72]byte
	CDI_FIRST     [8]volatile.Register32
	_             [32]byte
	UDI_FIRST     [2]volatile.Register32
	_             [62]byte
	RAM_ADDR_RAND volatile.Register16
	_             [2]byte
	RAM_DATA_RAND volatile.Register16
	_             [126]byte
	CPU_MON_CTRL  volatile.Register16
	_             [2]byte
	CPU_MON_FIRST volatile.Register32
	CPU_MON_LAST  volatile.Register32
	_             [60]byte
	SYSTEM_RESET  volatile.Register16
	_             [66]byte
	SPI_EN        volatile.Register32
	SPI_XFER      volatile.Register32
	SPI_DATA      volatile.Register32
}

const (
	TK1_MMIO_TIMER_CTRL_START_BIT = 0
	TK1_MMIO_TIMER_CTRL_STOP_BIT  = 1
	TK1_MMIO_TIMER_CTRL_START     = 1 << TK1_MMIO_TIMER_CTRL_START_BIT
	TK1_MMIO_TIMER_CTRL_STOP      = 1 << TK1_MMIO_TIMER_CTRL_STOP_BIT

	TK1_MMIO_TK1_LED_R_BIT = 2
	TK1_MMIO_TK1_LED_G_BIT = 1
	TK1_MMIO_TK1_LED_B_BIT = 0

	TK1_MMIO_TK1_GPIO1_BIT = 0
	TK1_MMIO_TK1_GPIO2_BIT = 1
	TK1_MMIO_TK1_GPIO3_BIT = 2
	TK1_MMIO_TK1_GPIO4_BIT = 3
)
