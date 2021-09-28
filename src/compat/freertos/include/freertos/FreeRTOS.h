#pragma once

#include <stdint.h>

typedef uint32_t     TickType_t;
typedef int          BaseType_t;
typedef unsigned int UBaseType_t;

#define portMAX_DELAY        (TickType_t)0xffffffffUL
#define portTICK_PERIOD_MS   (10)
#define configMAX_PRIORITIES (25)
