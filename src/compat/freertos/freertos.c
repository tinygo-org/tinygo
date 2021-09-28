#include "freertos/FreeRTOS.h"
#include "freertos/task.h"

void freertos_callFunction(void (*function)(void *), void *parameter) {
    function(parameter);
}
