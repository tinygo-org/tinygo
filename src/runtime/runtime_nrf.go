
// +build nrf

package runtime

// #include "runtime_nrf.h"
import "C"

import (
	"device/arm"
	"device/nrf"
)

const Microsecond = 1

func init() {
	initUART()
	initLFCLK()
	initRTC()
}

func initUART() {
	nrf.UART0.ENABLE        = nrf.UART0_ENABLE_ENABLE_Enabled
	nrf.UART0.BAUDRATE      = nrf.UART0_BAUDRATE_BAUDRATE_Baud115200
	nrf.UART0.TASKS_STARTTX = 1
	nrf.UART0.PSELTXD       = 6 // pin 6 for NRF52840-DK
}

func initLFCLK() {
	nrf.CLOCK.LFCLKSRC = nrf.CLOCK_LFCLKSTAT_SRC_Xtal
	nrf.CLOCK.TASKS_LFCLKSTART = 1
	for nrf.CLOCK.EVENTS_LFCLKSTARTED == 0 {}
	nrf.CLOCK.EVENTS_LFCLKSTARTED = 0
}

func initRTC() {
	nrf.RTC0.TASKS_START = 1
	// TODO: set priority
	arm.EnableIRQ(nrf.IRQ_RTC0)
}

func putchar(c byte) {
	nrf.UART0.TXD = nrf.RegValue(c)
	for nrf.UART0.EVENTS_TXDRDY == 0 {}
	nrf.UART0.EVENTS_TXDRDY = 0
}

func Sleep(d Duration) {
	C.rtc_sleep(C.uint32_t(d / 32)) // TODO: not accurate (must be d / 30.5175...)
}

func abort() {
	// TODO: wfi
	for {
	}
}

// Align on word boundary.
func align(ptr uintptr) uintptr {
	return (ptr + 3) &^ 3
}
