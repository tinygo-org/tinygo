#pragma once

// Note: in FreeRTOS, SemaphoreHandle_t is an alias for QueueHandle_t.
typedef struct SemaphoreDefinition * SemaphoreHandle_t;

SemaphoreHandle_t xSemaphoreCreateRecursiveMutex(void);

void vSemaphoreDelete(SemaphoreHandle_t xSemaphore);

// Note: these two functions are macros in FreeRTOS.
BaseType_t xSemaphoreTakeRecursive(SemaphoreHandle_t xMutex, TickType_t xTicksToWait);
BaseType_t xSemaphoreGiveRecursive(SemaphoreHandle_t xMutex);
