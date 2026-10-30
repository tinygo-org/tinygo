//go:build tinygo && rp2350

package machine

/*
typedef unsigned long uint32_t;

#define ram_func __attribute__((section(".ramfuncs"),noinline))

#define IO_QSPI_BASE 0x40030000
#define IO_QSPI_GPIO_QSPI_SS_CTRL_OFFSET 0x0000001c
#define IO_QSPI_GPIO_QSPI_SS_CTRL_OEOVER_BITS 0x0000c000
#define IO_QSPI_GPIO_QSPI_SS_CTRL_OEOVER_LSB 14

#define SIO_BASE 0xd0000000
#define SIO_GPIO_HI_IN_OFFSET 0x00000008
#define SIO_GPIO_HI_IN_QSPI_CSN_BITS 0x08000000

#define GPIO_OVERRIDE_LOW 0x2

uint32_t ram_func read_bootsel_button(void) {
	volatile uint32_t *ctrl = (volatile uint32_t *)(IO_QSPI_BASE + IO_QSPI_GPIO_QSPI_SS_CTRL_OFFSET);
	uint32_t saved_ctrl = *ctrl;

	*ctrl = (saved_ctrl & ~IO_QSPI_GPIO_QSPI_SS_CTRL_OEOVER_BITS) |
		(GPIO_OVERRIDE_LOW << IO_QSPI_GPIO_QSPI_SS_CTRL_OEOVER_LSB);

	// Same delay loop as the Raspberry Pi Pico example:
	// https://github.com/raspberrypi/pico-examples/blob/master/picoboard/button/button.c
	for (volatile int i = 0; i < 1000; ++i) {
	}

	volatile uint32_t *gpio_hi_in = (volatile uint32_t *)(SIO_BASE + SIO_GPIO_HI_IN_OFFSET);
	uint32_t pressed = !(*gpio_hi_in & SIO_GPIO_HI_IN_QSPI_CSN_BITS);

	*ctrl = saved_ctrl;
	return pressed;
}
*/
import "C"

// BootselButton reports whether the BOOTSEL button is pressed on the RP2350.
//
// With scheduler=cores, BootselButton must not be called from an interrupt
// handler or while interrupts are disabled because it must synchronize with
// the other core.
func BootselButton() bool {
	state, multicore := rp2EnterFlashSafeSection()
	res := C.read_bootsel_button()
	rp2ExitFlashSafeSection(state, multicore)
	return res != 0
}
