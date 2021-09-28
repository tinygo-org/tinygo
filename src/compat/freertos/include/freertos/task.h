#pragma once

typedef void * TaskHandle_t;
typedef void (*TaskFunction_t)(void *);

TaskHandle_t xTaskGetCurrentTaskHandle(void);

BaseType_t xTaskCreate(TaskFunction_t pvTaskCode, const char * const pcName, uintptr_t usStackDepth, void *pvParameters, UBaseType_t uxPriority, TaskHandle_t *pxCreatedTask);

void vTaskDelay(const TickType_t xTicksToDelay);
