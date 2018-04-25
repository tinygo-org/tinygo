
#include "hal/nrf_uart.h"
#include "nrf.h"
#include "runtime.h"

void uart_init(uint32_t pin_tx) {
	NRF_UART0->ENABLE        = UART_ENABLE_ENABLE_Enabled;
	NRF_UART0->BAUDRATE      = UART_BAUDRATE_BAUDRATE_Baud115200;
	NRF_UART0->TASKS_STARTTX = 1;
	NRF_UART0->PSELTXD       = 6;
}

void uart_send(uint8_t c) {
	NRF_UART0->TXD = c;
	while (NRF_UART0->EVENTS_TXDRDY != 1) {}
	NRF_UART0->EVENTS_TXDRDY = 0;
}

void _start() {
	uart_init(6); // pin_tx = 6, for NRF52840-DK
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
	for (size_t i = 0; i < n; i++) {
		dest[i] = 0;
	}
}

__attribute__((weak))
void __aeabi_memclr4(uint8_t *dest, size_t n) {
	__aeabi_memclr(dest, n);
}
