
#pragma once

#include <stdint.h>
#include <stdbool.h>

void uart_init(uint32_t pin_tx);
void uart_send(uint8_t c);

void rtc_init();
void rtc_sleep(uint32_t ticks);

typedef enum {
	GPIO_INPUT,
	GPIO_OUTPUT,
} gpio_mode_t;

void gpio_cfg(uint32_t pin, gpio_mode_t mode);
void gpio_set(uint32_t pin, bool high);
