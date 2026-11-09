//go:build none

// This file is manually included, to avoid CGo which would cause a circular
// import.

// Futex for the cooperative threads of wasip3. It does not use the futex in
// wasi-libc for waiting without timeout, because that allocates memory. The
// memory allocator itself uses futexes when it is locked.

#include <stdbool.h>
#include <stdint.h>
#include <time.h>
#include <wasi/api.h>

// Internal function of the cooperative threads implementation in wasi-libc.
void __wake(volatile void *addr, int cnt, int priv);
int __timedwait(volatile int *addr, int val, clockid_t clk, const struct timespec *at, int priv);

struct waiter {
    volatile uint32_t *addr;
    uint32_t           tid;
    bool               woken;
    struct waiter     *next;
    struct waiter     *prev;
};

// Threads only switch inside the calls below, so no locking is needed.
static struct waiter *waiters;

static void remove_waiter(struct waiter *w) {
    if (w->prev) {
        w->prev->next = w->next;
    } else {
        waiters = w->next;
    }
    if (w->next) {
        w->next->prev = w->prev;
    }
}

void tinygo_futex_wait(uint32_t *addr, uint32_t cmp) {
    if (*addr != cmp) {
        return;
    }
    struct waiter w = {
        .addr = addr,
        .tid  = wasip3_thread_index(),
        .next = waiters,
    };
    if (waiters) {
        waiters->prev = &w;
    }
    waiters = &w;
    wasip3_thread_suspend();
    __atomic_signal_fence(__ATOMIC_SEQ_CST);
    if (!w.woken) {
        // Spurious wakeup.
        remove_waiter(&w);
    }
}

void tinygo_futex_wait_timeout(uint32_t *addr, uint32_t cmp, uint64_t timeout) {
    // This allocates memory. But, timed waits are not used for the locks of
    // the memory allocator.
    struct timespec at;
    clock_gettime(CLOCK_MONOTONIC, &at);
    uint64_t nsec = (uint64_t)at.tv_nsec + timeout % 1000000000;
    at.tv_sec += timeout / 1000000000 + nsec / 1000000000;
    at.tv_nsec = nsec % 1000000000;
    __timedwait((volatile int*)addr, cmp, CLOCK_MONOTONIC, &at, 1);
}

static void wake(uint32_t *addr, bool all) {
    struct waiter *w = waiters;
    while (w) {
        struct waiter *next = w->next;
        if (w->addr == addr) {
            remove_waiter(w);
            w->woken = true;
            wasip3_thread_resume_later(w->tid);
            if (!all) {
                return;
            }
        }
        w = next;
    }
}

// Waiters with a timeout are in the list of wasi-libc, so wake those as well.
void tinygo_futex_wake(uint32_t *addr) {
    wake(addr, false);
    __wake(addr, 1, 1);
}

void tinygo_futex_wake_all(uint32_t *addr) {
    wake(addr, true);
    __wake(addr, -1, 1);
}
