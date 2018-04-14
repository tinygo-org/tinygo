
#pragma once

#include <stdint.h>

typedef struct {
	uint32_t len;  // TODO: size_t (or, let max string size depend on target/flag)
	uint8_t  *buf; // points to string buffer itself
} string_t;

typedef int32_t intgo_t; // may be 64-bit
