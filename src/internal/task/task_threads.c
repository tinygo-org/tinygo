//go:build none

#define _GNU_SOURCE
#include <pthread.h>
#include <signal.h>
#include <stdint.h>
#include <stdio.h>
#include <unistd.h>

#ifdef __linux__
#include <semaphore.h>

// BDWGC also uses SIGRTMIN+6 on Linux, which seems like a reasonable choice.
#define taskPauseSignal (SIGRTMIN + 6)

#elif __APPLE__
#include <dispatch/dispatch.h>
// SIGIO is for interrupt-driven I/O.
// I don't think anybody should be using this nowadays, so I think we can
// repurpose it as a signal for GC.
// BDWGC uses a special way to pause/resume other threads on MacOS, which may be
// better but needs more work. Using signal keeps the code similar between Linux
// and MacOS.
#define taskPauseSignal SIGIO

#endif // __linux__, __APPLE__

// Pointer to the current task.Task structure.
// Ideally the entire task.Task structure would be a thread-local variable but
// this also works.
static __thread void *current_task;

struct state_pass {
    void      *(*start)(void*);
    void      *args;
    void      *task;
    uintptr_t *stackTop;
    #if __APPLE__
    dispatch_semaphore_t startlock;
    #else
    sem_t     startlock;
    #endif
};

// Handle the GC pause in Go.
void tinygo_task_gc_pause(int sig);

// Initialize the main thread.
void tinygo_task_init(void *mainTask, pthread_t *thread, int *numCPU, void *context) {
    // Make sure the current task pointer is set correctly for the main
    // goroutine as well.
    current_task = mainTask;

    // Store the thread ID of the main thread.
    *thread = pthread_self();

    // Register the "GC pause" signal for the entire process.
    // Using pthread_kill, we can still send the signal to a specific thread.
    struct sigaction act = { 0 };
    act.sa_handler = tinygo_task_gc_pause;
    act.sa_flags = SA_RESTART;
    sigaction(taskPauseSignal, &act, NULL);

    // Obtain the number of CPUs available on program start (for NumCPU).
    int num = sysconf(_SC_NPROCESSORS_ONLN);
    if (num <= 0) {
        // Fallback in case there is an error.
        num = 1;
    }
    *numCPU = num;
}

void tinygo_task_exited(void*);

// Helper to start a goroutine while also storing the 'task' structure.
static void* start_wrapper(void *arg) {
    struct state_pass *state = arg;
    void *(*start)(void*) = state->start;
    void *args = state->args;
    current_task = state->task;

    // Save the current stack pointer in the goroutine state, for the GC.
    int stackAddr;
    *(state->stackTop) = (uintptr_t)(&stackAddr);

    // Notify the caller that the thread has successfully started and
    // initialized.
    #if __APPLE__
    dispatch_semaphore_signal(state->startlock);
    #else
    sem_post(&state->startlock);
    #endif

    // Run the goroutine function.
    start(args);

    // Notify the Go side this thread will exit.
    tinygo_task_exited(current_task);

    return NULL;
};

// Start a new goroutine in an OS thread.
int tinygo_task_start(uintptr_t fn, void *args, void *task, pthread_t *thread, uintptr_t *stackTop, uintptr_t stackSize, void *context) {
    // Sanity check. Should get optimized away.
    if (sizeof(pthread_t) != sizeof(void*)) {
        __builtin_trap();
    }

    struct state_pass state = {
        .start     = (void*)fn,
        .args      = args,
        .task      = task,
        .stackTop  = stackTop,
    };
    #if __APPLE__
    state.startlock = dispatch_semaphore_create(0);
    #else
    sem_init(&state.startlock, 0, 0);
    #endif
    pthread_attr_t attrs;
    pthread_attr_init(&attrs);
	pthread_attr_setdetachstate(&attrs, PTHREAD_CREATE_DETACHED);
    pthread_attr_setstacksize(&attrs, stackSize);
    int result = pthread_create(thread, &attrs, &start_wrapper, &state);
    pthread_attr_destroy(&attrs);
	if (result != 0) {
		return result;
	}

    // Wait until the thread has been created and read all state_pass variables.
    #if __APPLE__
    dispatch_semaphore_wait(state.startlock, DISPATCH_TIME_FOREVER);
    dispatch_release(state.startlock);
    #else
    sem_wait(&state.startlock);
    sem_destroy(&state.startlock);
    #endif

    return result;
}

// Return the current task (for task.Current()).
void* tinygo_task_current(void) {
    return current_task;
}

// Send a signal to cause the task to pause for the GC mark phase.
void tinygo_task_send_gc_signal(pthread_t thread) {
    pthread_kill(thread, taskPauseSignal);
}
