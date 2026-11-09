//go:build none

// Threads for wasip3, using the cooperative threads of wasi-libc. This is the
// equivalent of task_threads.c, but without signals.

#include <pthread.h>
#include <semaphore.h>
#include <stdint.h>

// Pointer to the current task.Task structure.
static __thread void *current_task;

struct state_pass {
    void      (*start)(void*);
    void      *args;
    void      *task;
    uintptr_t *stackTop;
    sem_t     startlock;
};

// Initialize the main thread.
void tinygo_task_init(void *mainTask, pthread_t *thread, int *numCPU) {
    current_task = mainTask;
    *thread = pthread_self();

    // Threads are cooperative, so there is no parallelism.
    *numCPU = 1;
}

void tinygo_task_exited(void*);

// Helper to start a goroutine while also storing the 'task' structure.
static void* start_wrapper(void *arg) {
    struct state_pass *state = arg;
    void (*start)(void*) = state->start;
    void *args = state->args;
    current_task = state->task;

    // Save the stack top in the goroutine state, for the GC.
    int stackAddr;
    *(state->stackTop) = (uintptr_t)(&stackAddr);

    // Notify the caller that the thread has successfully started and
    // initialized.
    sem_post(&state->startlock);

    // Run the goroutine function.
    start(args);

    // Notify the Go side this thread will exit.
    tinygo_task_exited(current_task);

    return NULL;
};

// Start a new goroutine in a thread.
int tinygo_task_start(uintptr_t fn, void *args, void *task, pthread_t *thread, uintptr_t *stackTop, uintptr_t stackSize) {
    struct state_pass state = {
        .start     = (void*)fn,
        .args      = args,
        .task      = task,
        .stackTop  = stackTop,
    };
    sem_init(&state.startlock, 0, 0);
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
    sem_wait(&state.startlock);
    sem_destroy(&state.startlock);

    return result;
}

// Return the current task (for task.Current()).
void* tinygo_task_current(void) {
    return current_task;
}

// Exit the current thread. Cooperative threads in wasi-libc cannot end a thread
// before the start function returns (there is no thread.exit), so park the
// thread forever instead. The task is already removed from the task list.
void tinygo_task_exit_thread(void) {
    sem_t forever;
    sem_init(&forever, 0, 0);
    for (;;) {
        sem_wait(&forever);
    }
}
