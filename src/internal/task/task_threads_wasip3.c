//go:build none

// Threads for wasip3, using the cooperative threads of wasi-libc. This is the
// equivalent of task_threads.c, but without signals.
//
// Threads are kept in a pool. A thread that finishes its goroutine waits for
// the next goroutine instead of exiting, which saves the cost of creating a
// thread. Cooperative threads in wasi-libc cannot end a thread before its start
// function returns, so a goroutine that ends with Goexit also turns its thread
// into a pool thread, on top of the frames that Goexit left.

#include <pthread.h>
#include <semaphore.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdlib.h>

// Futex that does not allocate, see src/internal/futex/futex_wasip3.c.
void tinygo_futex_wait(uint32_t *addr, uint32_t cmp);
void tinygo_futex_wake(uint32_t *addr);

// The most idle threads to keep. Threads that finish when more are idle exit.
#define MAX_IDLE_WORKERS 32

// How many times a thread can restart after Goexit. Each time leaves a few
// frames on the stack of the thread. After this, the thread waits forever.
#define MAX_GOEXIT_NESTING 8

// Pointer to the current task.Task structure.
static __thread void *current_task;

struct worker {
    // The goroutine to run.
    void      (*start)(void*);
    void      *args;
    void      *task;

    // The highest address of the stack of the thread.
    uintptr_t stackTop;

    // Futex that is nonzero when the worker has a goroutine to run.
    uint32_t  has_job;

    // Number of times that Goexit restarted this thread.
    int       nesting;

    // Next worker in the list of idle workers.
    struct worker *next;
};

// Idle workers. Threads only switch at blocking calls, so no lock is needed.
static struct worker *idle_workers;
static int idle_count;

// The worker of the current thread, or NULL for the main thread.
static __thread struct worker *self_worker;

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

static void run_job(struct worker *w) {
    current_task = w->task;
    w->start(w->args);

    // Notify the Go side this goroutine ended.
    tinygo_task_exited(current_task);
    current_task = NULL;
}

// Run goroutines until too many workers are idle. If force_idle is true, this
// does not stop for that reason and does not return.
static void worker_loop(struct worker *w, bool force_idle) {
    for (;;) {
        if (!force_idle && idle_count >= MAX_IDLE_WORKERS) {
            return;
        }
        w->has_job = 0;
        w->next = idle_workers;
        idle_workers = w;
        idle_count++;
        while (!w->has_job) {
            tinygo_futex_wait(&w->has_job, 0);
        }
        run_job(w);
    }
}

// Entry of a new thread.
static void* worker_main(void *arg) {
    struct state_pass *state = arg;
    struct worker *w = calloc(1, sizeof(struct worker));
    if (!w) {
        __builtin_trap();
    }
    w->start = state->start;
    w->args = state->args;
    w->task = state->task;
    self_worker = w;

    // Save the stack top in the goroutine state, for the GC.
    int stackAddr;
    w->stackTop = (uintptr_t)(&stackAddr);
    *(state->stackTop) = w->stackTop;

    // Notify the caller that the thread has successfully started and
    // initialized.
    sem_post(&state->startlock);

    run_job(w);
    worker_loop(w, false);

    self_worker = NULL;
    free(w);
    return NULL;
}

// Start a new goroutine in a thread.
int tinygo_task_start(uintptr_t fn, void *args, void *task, pthread_t *thread, uintptr_t *stackTop, uintptr_t stackSize) {
    // Reuse an idle thread.
    struct worker *w = idle_workers;
    if (w) {
        idle_workers = w->next;
        idle_count--;
        w->next = NULL;
        w->start = (void*)fn;
        w->args = args;
        w->task = task;
        *stackTop = w->stackTop;
        w->has_job = 1;
        tinygo_futex_wake(&w->has_job);
        return 0;
    }

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
    int result = pthread_create(thread, &attrs, &worker_main, &state);
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

// Called when a goroutine ends with Goexit, which cannot return to the start
// function of the thread. The task is already removed from the task list.
void tinygo_task_exit_thread(void) {
    struct worker *w = self_worker;
    if (w && ++w->nesting <= MAX_GOEXIT_NESTING) {
        current_task = NULL;
        worker_loop(w, true);
    }

    // The main thread, or a thread that already restarted too many times.
    sem_t forever;
    sem_init(&forever, 0, 0);
    for (;;) {
        sem_wait(&forever);
    }
}
