
#include "hal/nrf_gpio.h"
#include "hal/nrf_uart.h"
#include "nrf.h"
#include "runtime.h"
#include "runtime_nrf.h"
#include <string.h>

static volatile bool rtc_wakeup;

void rtc_sleep(uint32_t ticks) {
	NRF_RTC0->INTENSET = RTC_INTENSET_COMPARE0_Msk;
	rtc_wakeup = false;
	if (ticks == 1) {
		// Race condition (even in hardware) at ticks == 1.
		// TODO: fix this in a better way by detecting it, like the manual
		// describes.
		ticks = 2;
	}
	NRF_RTC0->CC[0] = (NRF_RTC0->COUNTER + ticks) & 0x00ffffff;
	while (!rtc_wakeup) {
		__WFI();
	}
}

void RTC0_IRQHandler() {
	NRF_RTC0->INTENCLR = RTC_INTENSET_COMPARE0_Msk;
	NRF_RTC0->EVENTS_COMPARE[0] = 0;
	rtc_wakeup = true;
}

void _start() {
	main();
}

__attribute__((weak))
void __aeabi_unwind_cpp_pr0() {
	// dummy, not actually used
}

__attribute__((weak))
void __aeabi_memclr(uint8_t *dest, size_t n) {
	// TODO: link with compiler-rt for a better implementation.
	// For now, use a simple memory zeroer.
	memset(dest, 0, n);
}

__attribute__((weak))
void __aeabi_memclr4(uint8_t *dest, size_t n) {
	__aeabi_memclr(dest, n);
}
