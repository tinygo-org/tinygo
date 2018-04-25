
#pragma once

#include <stdint.h>

void uart_init(uint32_t pin_tx);
void uart_send(uint8_t c);
