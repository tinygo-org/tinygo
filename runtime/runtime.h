
#pragma once

#include <stdint.h>

typedef struct {
	uint32_t len;   // TODO: size_t (or, let max string size depend on target/flag)
	uint8_t  buf[]; // variable size
} string_t;

void __go_printstring(string_t *str);
void __go_printnl();
