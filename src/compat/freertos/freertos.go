// Package freertos provides a compatibility layer on top of the TinyGo
// scheduler for C code that wants to call FreeRTOS functions. One example is
// the ESP-IDF, which expects there to be a FreeRTOS-like RTOS.
package freertos

// #include <freertos/FreeRTOS.h>
// #include <freertos/queue.h>
// #include <freertos/semphr.h>
// #include <freertos/task.h>
// void freertos_callFunction(void (*function)(void *), void *parameter);
import "C"
import (
	"sync"
	"time"
	"unsafe"

	"internal/task"
)

//export xTaskGetCurrentTaskHandle
func xTaskGetCurrentTaskHandle() C.TaskHandle_t {
	return C.TaskHandle_t(task.Current())
}

//export xTaskCreate
func xTaskCreate(pvTaskCode C.TaskFunction_t, pcName *C.char, usStackDepth uintptr, pvParameters unsafe.Pointer, uxPriority C.UBaseType_t, pxCreatedTask *C.TaskHandle_t) C.BaseType_t {
	go func() {
		C.freertos_callFunction(pvTaskCode, pvParameters)
	}()
	if pxCreatedTask != nil {
		// Code expectes there to be *something*.
		var tcb int
		*pxCreatedTask = unsafe.Pointer(&tcb)
	}
	return 1 // pdPASS
}

//export vTaskDelay
func vTaskDelay(xTicksToDelay C.TickType_t) {
	// The default tick rate appears to be 100Hz (10ms per tick).
	time.Sleep(time.Duration(xTicksToDelay) * C.portTICK_PERIOD_MS * time.Millisecond)
}

type Semaphore struct {
	lock  sync.Mutex // the lock itself
	task  *task.Task // the task currently locking this semaphore
	count uint32     // how many times this semaphore is locked
}

//export xSemaphoreCreateCounting
func xSemaphoreCreateCounting(uxMaxCount C.UBaseType_t, uxInitialCount C.UBaseType_t) C.SemaphoreHandle_t {
	if uxMaxCount != 1 || uxInitialCount != 0 {
		println("TODO: xSemaphoreCreateCounting that's not a mutex")
		return nil
	}
	mutex := Semaphore{}
	return C.SemaphoreHandle_t(&mutex)
}

//export xSemaphoreCreateRecursiveMutex
func xSemaphoreCreateRecursiveMutex() C.SemaphoreHandle_t {
	var mutex Semaphore
	return (C.SemaphoreHandle_t)(&mutex)
}

//export vSemaphoreDelete
func vSemaphoreDelete(xSemaphore C.SemaphoreHandle_t) {
	mutex := (*Semaphore)(xSemaphore)
	if mutex.task != nil {
		panic("vSemaphoreDelete: still locked")
	}
}

//export xSemaphoreTake
func xSemaphoreTake(xSemaphore C.QueueHandle_t, xTicksToWait C.TickType_t) C.BaseType_t {
	mutex := (*Semaphore)(xSemaphore)
	mutex.lock.Lock()
	return 1 // pdTRUE
}

//export xSemaphoreTakeRecursive
func xSemaphoreTakeRecursive(xMutex C.SemaphoreHandle_t, xTicksToWait C.TickType_t) C.BaseType_t {
	// TODO: implement xTickToWait, or at least when xTicksToWait equals 0.
	mutex := (*Semaphore)(xMutex)
	if mutex.task == task.Current() {
		// Already locked.
		mutex.count++
		return 1 // pdTRUE
	}
	// Not yet locked.
	mutex.lock.Lock()
	mutex.task = task.Current()
	return 1 // pdTRUE
}

//export xSemaphoreGive
func xSemaphoreGive(xSemaphore C.QueueHandle_t) C.BaseType_t {
	mutex := (*Semaphore)(xSemaphore)
	mutex.lock.Unlock()
	return 1 // pdTRUE
}

//export xSemaphoreGiveRecursive
func xSemaphoreGiveRecursive(xMutex C.SemaphoreHandle_t) C.BaseType_t {
	mutex := (*Semaphore)(xMutex)
	if mutex.task == task.Current() {
		// Already locked.
		mutex.count--
		if mutex.count == 0 {
			mutex.lock.Unlock()
		}
		return 1 // pdTRUE
	}
	panic("xSemaphoreGiveRecursive: not locked by this task")
}

//export xQueueCreate
func xQueueCreate(uxQueueLength C.UBaseType_t, uxItemSize C.UBaseType_t) C.QueueHandle_t {
	return chanMakeUnsafePointer(uintptr(uxItemSize), uintptr(uxQueueLength))
}

//export vQueueDelete
func vQueueDelete(xQueue C.QueueHandle_t) {
	// TODO: close the channel
}

//export xQueueReceive
func xQueueReceive(xQueue C.QueueHandle_t, pvBuffer unsafe.Pointer, xTicksToWait C.TickType_t) C.BaseType_t {
	// Note: xTicksToWait is ignored.
	chanRecvUnsafePointer(xQueue, pvBuffer)
	return 1 // pdTRUE
}

//export xQueueSend
func xQueueSend(xQueue C.QueueHandle_t, pvBuffer unsafe.Pointer, xTicksToWait C.TickType_t) C.BaseType_t {
	// Note: xTicksToWait is ignored.
	chanSendUnsafePointer(xQueue, pvBuffer)
	return 1 // pdTRUE
}

//export uxQueueMessagesWaiting
func uxQueueMessagesWaiting(xQueue C.QueueHandle_t) C.UBaseType_t {
	return C.UBaseType_t(chanLenUnsafePointer(xQueue))
}

//go:linkname chanMakeUnsafePointer runtime.chanMakeUnsafePointer
func chanMakeUnsafePointer(elementSize uintptr, bufSize uintptr) unsafe.Pointer

//go:linkname chanLenUnsafePointer runtime.chanLenUnsafePointer
func chanLenUnsafePointer(ch unsafe.Pointer) int

//go:linkname chanSendUnsafePointer runtime.chanSendUnsafePointer
func chanSendUnsafePointer(ch, value unsafe.Pointer)

//go:linkname chanRecvUnsafePointer runtime.chanRecvUnsafePointer
func chanRecvUnsafePointer(ch, value unsafe.Pointer)
