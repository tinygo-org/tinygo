//go:build none

// Ignore the //go:build above. This file is included on Linux for CPU
// profiling. Each thread gets a CPU time timer that sends it SIGPROF.

#define _GNU_SOURCE
#include <stdint.h>
#include <string.h>
#include <stdlib.h>
#include <signal.h>
#include <time.h>
#include <unistd.h>
#include <fcntl.h>
#include <pthread.h>
#include <sched.h>
#include <sys/syscall.h>
#include <ucontext.h>
#include <sys/mman.h>
#include <sys/resource.h>

#if defined(__aarch64__) || defined(__x86_64__)

// glibc does not name the thread ID field of struct sigevent, only the kernel
// headers do.
#ifndef sigev_notify_thread_id
#define sigev_notify_thread_id _sigev_un._tid
#endif

#define CPUPROF_MAX_DEPTH 64
#define CPUPROF_SLOT_WORDS (CPUPROF_MAX_DEPTH + 2)

// User space addresses fit in 48 bits. Masking also strips pointer
// authentication bits.
#define CPUPROF_ADDR_MASK ((((uintptr_t)1) << 48) - 1)

// Each sample is the number of PCs, the CPU time in nanoseconds and the PCs.
// The count is stored last, so zero means the sample is not complete.
static uintptr_t *cpuprof_buf;
// Reused by later profiles.
static uintptr_t *cpuprof_mem;
static size_t cpuprof_mem_bytes;
static uint32_t cpuprof_slots;
static uint32_t cpuprof_next;
static uint32_t cpuprof_dropped;
// Handlers and exiting threads that may be writing a sample.
static uint32_t cpuprof_writers;
static uintptr_t cpuprof_period_ns;
static uint64_t cpuprof_start_ns;
static uint64_t cpuprof_stop_ns;
static uint64_t cpuprof_sampled_ns;

// Highest address of the current thread's stack, or 0 if not known.
static __thread uintptr_t cpuprof_stack_hi;

// The timer of each thread, while profiling.
#define CPUPROF_MAX_TIMERS 4096
static pthread_mutex_t cpuprof_lock = PTHREAD_MUTEX_INITIALIZER;
static int cpuprof_active;
// Timers are made with system calls, as TinyGo's musl does not build
// timer_create, and use kernel timer IDs.
static struct {
    pid_t tid;
    int timer;
    uint64_t base_ns; // thread CPU time when the timer was made
} cpuprof_timers[CPUPROF_MAX_TIMERS];
static int cpuprof_ntimers;

// Claim a sample slot, or return NULL. A writer is registered before the
// buffer is checked, so stop either waits for it or it sees profiling stopped.
static uintptr_t *cpuprof_claim(void) {
    __atomic_fetch_add(&cpuprof_writers, 1, __ATOMIC_SEQ_CST);
    uintptr_t *buf = __atomic_load_n(&cpuprof_buf, __ATOMIC_SEQ_CST);
    uint32_t slot = buf != NULL ? __atomic_fetch_add(&cpuprof_next, 1, __ATOMIC_RELAXED) : 0;
    if (buf != NULL && slot >= cpuprof_slots) {
        __atomic_fetch_add(&cpuprof_dropped, 1, __ATOMIC_RELAXED);
        buf = NULL;
    }
    if (buf == NULL) {
        __atomic_fetch_sub(&cpuprof_writers, 1, __ATOMIC_RELEASE);
        return NULL;
    }
    return buf + (uintptr_t)slot * CPUPROF_SLOT_WORDS;
}

// Publish a sample claimed with cpuprof_claim.
static void cpuprof_publish(uintptr_t *rec, uintptr_t n, uint64_t ns) {
    rec[1] = ns;
    __atomic_fetch_add(&cpuprof_sampled_ns, ns, __ATOMIC_RELAXED);
    __atomic_store_n(&rec[0], n, __ATOMIC_RELEASE);
    __atomic_fetch_sub(&cpuprof_writers, 1, __ATOMIC_RELEASE);
}

static void cpuprof_handler(int sig, siginfo_t *info, void *context) {
    uintptr_t *rec = cpuprof_claim();
    if (rec == NULL) {
        return;
    }

    ucontext_t *uc = context;
#if defined(__aarch64__)
    uintptr_t pc = uc->uc_mcontext.pc;
    uintptr_t fp = uc->uc_mcontext.regs[29];
    uintptr_t sp = uc->uc_mcontext.sp;
#elif defined(__x86_64__)
    uintptr_t pc = uc->uc_mcontext.gregs[REG_RIP];
    uintptr_t fp = uc->uc_mcontext.gregs[REG_RBP];
    uintptr_t sp = uc->uc_mcontext.gregs[REG_RSP];
#else
#error "unsupported architecture"
#endif

    // Only follow frame records between the stack pointer and the stack top.
    uintptr_t hi = cpuprof_stack_hi;
    uintptr_t *pcs = rec + 2;
    uintptr_t n = 0;
    pcs[n++] = pc & CPUPROF_ADDR_MASK;
    while (n < CPUPROF_MAX_DEPTH && fp >= sp && fp + 2 * sizeof(uintptr_t) <= hi &&
           (fp & (sizeof(uintptr_t) - 1)) == 0) {
        uintptr_t next = ((uintptr_t *)fp)[0];
        uintptr_t ret = ((uintptr_t *)fp)[1] & CPUPROF_ADDR_MASK;
        if (ret == 0) {
            break;
        }
        pcs[n++] = ret;
        if (next <= fp) {
            break;
        }
        fp = next;
    }
    cpuprof_publish(rec, n, cpuprof_period_ns);
}

// The CPU time clock of a thread, as glibc's pthread_getcpuclockid makes it.
// https://github.com/torvalds/linux/blob/master/include/linux/posix-timers_types.h
static clockid_t cpuprof_thread_clock(pid_t tid) {
    return (clockid_t)((~(uint32_t)tid << 3) | 6);
}

// Give a thread a timer that sends it SIGPROF as it uses CPU time.
static void cpuprof_add_timer_locked(pid_t tid) {
    for (int i = 0; i < cpuprof_ntimers; i++) {
        if (cpuprof_timers[i].tid == tid) {
            return;
        }
    }
    if (cpuprof_ntimers >= CPUPROF_MAX_TIMERS) {
        return;
    }
    struct sigevent sev;
    memset(&sev, 0, sizeof(sev));
    sev.sigev_notify = SIGEV_THREAD_ID;
    sev.sigev_signo = SIGPROF;
    sev.sigev_notify_thread_id = tid;
    int timer;
    if (syscall(SYS_timer_create, cpuprof_thread_clock(tid), &sev, &timer) != 0) {
        return;
    }
    struct timespec cpu;
    if (clock_gettime(cpuprof_thread_clock(tid), &cpu) != 0) {
        syscall(SYS_timer_delete, timer);
        return;
    }
    struct itimerspec its;
    memset(&its, 0, sizeof(its));
    its.it_interval.tv_nsec = cpuprof_period_ns;
    its.it_value.tv_nsec = cpuprof_period_ns;
    if (syscall(SYS_timer_settime, timer, 0, &its, NULL) != 0) {
        syscall(SYS_timer_delete, timer);
        return;
    }
    cpuprof_timers[cpuprof_ntimers].tid = tid;
    cpuprof_timers[cpuprof_ntimers].timer = timer;
    cpuprof_timers[cpuprof_ntimers].base_ns = (uint64_t)cpu.tv_sec * 1000000000ull + cpu.tv_nsec;
    cpuprof_ntimers++;
}

// Give every thread of the process a timer. TinyGo's musl does not build
// opendir, so read the directory with getdents64. See getdents(2).
static void cpuprof_add_timers_locked(void) {
    int fd = open("/proc/self/task", O_RDONLY | O_DIRECTORY);
    if (fd < 0) {
        return;
    }
    char buf[4096];
    long n;
    while ((n = syscall(SYS_getdents64, fd, buf, sizeof(buf))) > 0) {
        for (long pos = 0; pos < n;) {
            // struct linux_dirent64 is d_ino, d_off, d_reclen, d_type, d_name.
            unsigned short reclen;
            memcpy(&reclen, buf + pos + 16, sizeof(reclen));
            pid_t tid = atoi(buf + pos + 19);
            if (tid > 0) {
                cpuprof_add_timer_locked(tid);
            }
            pos += reclen;
        }
    }
    close(fd);
}

static uint64_t cpuprof_process_ns(void) {
    struct rusage ru;
    if (getrusage(RUSAGE_SELF, &ru) != 0) {
        return 0;
    }
    return ((uint64_t)ru.ru_utime.tv_sec + ru.ru_stime.tv_sec) * 1000000000ull +
           ((uint64_t)ru.ru_utime.tv_usec + ru.ru_stime.tv_usec) * 1000ull;
}

// Called by every goroutine thread when it starts, and for the main thread.
void tinygo_cpuprof_thread_start(void) {
    pthread_attr_t attr;
    if (pthread_getattr_np(pthread_self(), &attr) == 0) {
        void *addr;
        size_t size;
        if (pthread_attr_getstack(&attr, &addr, &size) == 0) {
            cpuprof_stack_hi = (uintptr_t)addr + size;
        }
        pthread_attr_destroy(&attr);
    }
    pthread_mutex_lock(&cpuprof_lock);
    if (cpuprof_active) {
        cpuprof_add_timer_locked(gettid());
    }
    pthread_mutex_unlock(&cpuprof_lock);
}

// Called by every goroutine thread before it exits. Its CPU time since the
// last expiry goes to its entry function, so short threads are still counted.
// The main thread passes 0 and only has its timer removed.
void tinygo_cpuprof_thread_exit(uintptr_t fn) {
    pid_t tid = gettid();
    struct timespec cpu;
    clock_gettime(CLOCK_THREAD_CPUTIME_ID, &cpu);
    uint64_t now = (uint64_t)cpu.tv_sec * 1000000000ull + cpu.tv_nsec;
    pthread_mutex_lock(&cpuprof_lock);
    for (int i = 0; i < cpuprof_ntimers; i++) {
        if (cpuprof_timers[i].tid == tid) {
            syscall(SYS_timer_delete, cpuprof_timers[i].timer);
            uint64_t rest = (now - cpuprof_timers[i].base_ns) % cpuprof_period_ns;
            cpuprof_timers[i] = cpuprof_timers[--cpuprof_ntimers];
            uintptr_t *rec = rest > 0 && fn != 0 ? cpuprof_claim() : NULL;
            if (rec != NULL) {
                rec[2] = fn;
                cpuprof_publish(rec, 1, rest);
            }
            break;
        }
    }
    pthread_mutex_unlock(&cpuprof_lock);
}

// Start sampling hz times per second of CPU time with room for slots samples.
// Returns 0 on failure.
int tinygo_cpuprof_start(int32_t hz, uint32_t slots) {
    size_t bytes = (size_t)slots * CPUPROF_SLOT_WORDS * sizeof(uintptr_t);
    if (cpuprof_mem_bytes < bytes) {
        void *p = mmap(NULL, bytes, PROT_READ | PROT_WRITE, MAP_PRIVATE | MAP_ANONYMOUS, -1, 0);
        if (p == MAP_FAILED) {
            return 0;
        }
        cpuprof_mem = p;
        cpuprof_mem_bytes = bytes;
    }
    memset(cpuprof_mem, 0, bytes);
    cpuprof_slots = slots;
    cpuprof_period_ns = 1000000000 / hz;
    __atomic_store_n(&cpuprof_next, 0, __ATOMIC_RELAXED);
    __atomic_store_n(&cpuprof_dropped, 0, __ATOMIC_RELAXED);
    __atomic_store_n(&cpuprof_buf, cpuprof_mem, __ATOMIC_RELEASE);

    struct sigaction act;
    memset(&act, 0, sizeof(act));
    act.sa_sigaction = cpuprof_handler;
    act.sa_flags = SA_RESTART | SA_SIGINFO;
    sigemptyset(&act.sa_mask);
    if (sigaction(SIGPROF, &act, NULL) != 0) {
        __atomic_store_n(&cpuprof_buf, NULL, __ATOMIC_RELEASE);
        return 0;
    }

    cpuprof_start_ns = cpuprof_process_ns();
    __atomic_store_n(&cpuprof_sampled_ns, 0, __ATOMIC_RELAXED);
    pthread_mutex_lock(&cpuprof_lock);
    cpuprof_active = 1;
    cpuprof_add_timers_locked();
    pthread_mutex_unlock(&cpuprof_lock);
    return 1;
}

void tinygo_cpuprof_stop(void) {
    pthread_mutex_lock(&cpuprof_lock);
    cpuprof_active = 0;
    for (int i = 0; i < cpuprof_ntimers; i++) {
        syscall(SYS_timer_delete, cpuprof_timers[i].timer);
    }
    cpuprof_ntimers = 0;
    pthread_mutex_unlock(&cpuprof_lock);
    cpuprof_stop_ns = cpuprof_process_ns();
    __atomic_store_n(&cpuprof_buf, NULL, __ATOMIC_SEQ_CST);
    // Wait for handlers that claimed a slot before the buffer was cleared.
    while (__atomic_load_n(&cpuprof_writers, __ATOMIC_ACQUIRE) != 0) {
        sched_yield();
    }
    // A SIGPROF that is still pending would otherwise end the process.
    struct sigaction act;
    memset(&act, 0, sizeof(act));
    act.sa_handler = SIG_IGN;
    sigaction(SIGPROF, &act, NULL);
}

uintptr_t *tinygo_cpuprof_samples(uint32_t *used, uint32_t *dropped, uint32_t *slotWords) {
    uint32_t next = __atomic_load_n(&cpuprof_next, __ATOMIC_RELAXED);
    *used = next < cpuprof_slots ? next : cpuprof_slots;
    *dropped = __atomic_load_n(&cpuprof_dropped, __ATOMIC_RELAXED);
    *slotWords = CPUPROF_SLOT_WORDS;
    return cpuprof_mem;
}

// Return CPU time no sample covers, such as threads without a timer.
void tinygo_cpuprof_unsampled(uint64_t *unattributed_ns, uint64_t *sampler_ns) {
    uint64_t total = cpuprof_stop_ns > cpuprof_start_ns ? cpuprof_stop_ns - cpuprof_start_ns : 0;
    uint64_t sampled = __atomic_load_n(&cpuprof_sampled_ns, __ATOMIC_RELAXED);
    *unattributed_ns = total > sampled ? total - sampled : 0;
    *sampler_ns = 0;
}

#else

// CPU profiling is only supported on amd64 and arm64. The task code calls
// these hooks on every architecture.
void tinygo_cpuprof_thread_start(void) {}
void tinygo_cpuprof_thread_exit(uintptr_t fn) {}

#endif
