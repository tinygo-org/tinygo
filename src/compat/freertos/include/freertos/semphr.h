#pragma once

// Note: in FreeRTOS, SemaphoreHandle_t is an alias for QueueHandle_t.
typedef void * SemaphoreHandle_t;

SemaphoreHandle_t xSemaphoreCreateCounting(UBaseType_t uxMaxCount, UBaseType_t uxInitialCount);
SemaphoreHandle_t xSemaphoreCreateRecursiveMutex(void);

void vSemaphoreDelete(SemaphoreHandle_t xSemaphore);

// Note: these two functions are macros in FreeRTOS.
BaseType_t xSemaphoreTakeRecursive(SemaphoreHandle_t xMutex, TickType_t xTicksToWait);
BaseType_t xSemaphoreGiveRecursive(SemaphoreHandle_t xMutex);

// Note: these functions are macros in FreeRTOS.
BaseType_t xSemaphoreTake(QueueHandle_t xSemaphore, TickType_t xTicksToWait);
BaseType_t xSemaphoreGive(QueueHandle_t xSemaphore);
