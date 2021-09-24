// Package freertos provides a compatibility layer on top of the TinyGo
// scheduler for C code that wants to call FreeRTOS functions. One example is
// the ESP-IDF, which expects there to be a FreeRTOS-like RTOS.
package freertos

// #include <freertos/FreeRTOS.h>
// #include <freertos/semphr.h>
// #include <freertos/task.h>
import "C"
import (
	"sync"
	"unsafe"

	"internal/task"
)

//export xTaskGetCurrentTaskHandle
func xTaskGetCurrentTaskHandle() C.TaskHandle_t {
	return C.TaskHandle_t(task.Current())
}

type Semaphore struct {
	lock  sync.Mutex // the lock itself
	task  *task.Task // the task currently locking this semaphore
	count uint32     // how many times this semaphore is locked
}

//export xSemaphoreCreateRecursiveMutex
func xSemaphoreCreateRecursiveMutex() C.SemaphoreHandle_t {
	var mutex Semaphore
	return (C.SemaphoreHandle_t)(unsafe.Pointer(&mutex))
}

//export vSemaphoreDelete
func vSemaphoreDelete(xSemaphore C.SemaphoreHandle_t) {
	mutex := (*Semaphore)(unsafe.Pointer(xSemaphore))
	if mutex.task != nil {
		panic("vSemaphoreDelete: still locked")
	}
}

//export xSemaphoreTakeRecursive
func xSemaphoreTakeRecursive(xMutex C.SemaphoreHandle_t, xTicksToWait C.TickType_t) C.BaseType_t {
	// TODO: implement xTickToWait, or at least when xTicksToWait equals 0.
	mutex := (*Semaphore)(unsafe.Pointer(xMutex))
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

//export xSemaphoreGiveRecursive
func xSemaphoreGiveRecursive(xMutex C.SemaphoreHandle_t) C.BaseType_t {
	mutex := (*Semaphore)(unsafe.Pointer(xMutex))
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
