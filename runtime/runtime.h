
#pragma once

#include <stdint.h>

typedef struct {
	uint32_t length; // TODO: size_t
	uint8_t *data;
} string_t;

void __go_printstring(char *str);
void __go_printnl();
