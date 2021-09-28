#pragma once

typedef void * QueueHandle_t;

QueueHandle_t xQueueCreate(UBaseType_t uxQueueLength, UBaseType_t uxItemSize);
void vQueueDelete(QueueHandle_t xQueue);

BaseType_t xQueueReceive(QueueHandle_t xQueue, void *pvBuffer, TickType_t xTicksToWait);
BaseType_t xQueueSend(QueueHandle_t xQueue, void *pvBuffer, TickType_t xTicksToWait);
UBaseType_t uxQueueMessagesWaiting(QueueHandle_t xQueue);
