//go:build none

// Ignore the //go:build above. This file is included on macOS for CPU
// profiling. It samples threads with Mach calls like sample(1) does.

#include <stdint.h>
#include <string.h>
#include <time.h>
#include <pthread.h>
#include <sys/mman.h>
#include <sys/resource.h>

#define CPUPROF_MAX_DEPTH 64
#define CPUPROF_SLOT_WORDS (CPUPROF_MAX_DEPTH + 2)

// User space addresses fit in 47 bits. Masking also strips pointer
// authentication bits.
#define CPUPROF_ADDR_MASK ((((uintptr_t)1) << 47) - 1)

// The minimal SDK does not have <mach/mach.h>. These match the xnu headers in
// https://github.com/apple-oss-distributions/xnu/blob/main/osfmk/mach
typedef uint32_t cpuprof_port_t;
extern cpuprof_port_t mach_task_self_;
cpuprof_port_t mach_thread_self(void);
int mach_port_deallocate(cpuprof_port_t task, cpuprof_port_t name);
int vm_deallocate(cpuprof_port_t task, uintptr_t address, uintptr_t size);
int task_threads(cpuprof_port_t task, cpuprof_port_t **threads, uint32_t *count);
int thread_suspend(cpuprof_port_t thread);
int thread_resume(cpuprof_port_t thread);
int thread_get_state(cpuprof_port_t thread, int flavor, uint32_t *state, uint32_t *count);
int thread_info(cpuprof_port_t thread, uint32_t flavor, int32_t *info, uint32_t *count);

struct cpuprof_time_value {
    int32_t seconds, microseconds;
};
struct cpuprof_thread_basic_info {
    struct cpuprof_time_value user_time, system_time;
    int32_t cpu_usage, policy, run_state, flags, suspend_count, sleep_time;
};
#define CPUPROF_THREAD_BASIC_INFO 3
struct cpuprof_thread_identifier_info {
    uint64_t thread_id, thread_handle, dispatch_qaddr;
};
#define CPUPROF_THREAD_IDENTIFIER_INFO 4

#if defined(__aarch64__)
struct cpuprof_thread_state {
    uint64_t x[29], fp, lr, sp, pc;
    uint32_t cpsr, flags;
};
#define CPUPROF_THREAD_STATE 6 // ARM_THREAD_STATE64
#define CPUPROF_PC(s) ((s).pc)
#define CPUPROF_FP(s) ((s).fp)
#define CPUPROF_SP(s) ((s).sp)
#elif defined(__x86_64__)
struct cpuprof_thread_state {
    uint64_t rax, rbx, rcx, rdx, rdi, rsi, rbp, rsp;
    uint64_t r8, r9, r10, r11, r12, r13, r14, r15;
    uint64_t rip, rflags, cs, fs, gs;
};
#define CPUPROF_THREAD_STATE 4 // x86_THREAD_STATE64
#define CPUPROF_PC(s) ((s).rip)
#define CPUPROF_FP(s) ((s).rbp)
#define CPUPROF_SP(s) ((s).rsp)
#else
#error "unsupported architecture"
#endif

// Each sample is the number of PCs, the CPU time in nanoseconds and the PCs,
// leaf first. The sampler thread and exiting threads write them.
static uintptr_t *cpuprof_mem;
static size_t cpuprof_mem_bytes;
static uint32_t cpuprof_slots;
static uint32_t cpuprof_next;
static uint32_t cpuprof_dropped;

// CPU time of each thread at the previous wakeup. Each wakeup fills the other
// table with the live threads, so exited threads drop out.
#define CPUPROF_MAX_THREADS 4096
struct cpuprof_entry {
    uint64_t id;
    uint64_t cpu_ns;
};
static struct cpuprof_entry cpuprof_tables[2][CPUPROF_MAX_THREADS];
static uint16_t cpuprof_used[2][CPUPROF_MAX_THREADS];
static uint32_t cpuprof_nused[2];
static int cpuprof_prev;
// Guards the tables, which exiting threads also use. The sampler never holds
// it while a thread is suspended.
static pthread_mutex_t cpuprof_table_lock = PTHREAD_MUTEX_INITIALIZER;

// Profiling states. Exiting threads only record samples once the sampler has
// noted the CPU time of the threads that were already running.
#define CPUPROF_OFF 0
#define CPUPROF_STARTING 1
#define CPUPROF_RECORDING 2

static pthread_t cpuprof_sampler;
static int cpuprof_state;
// Exiting threads that may be writing samples or tables.
static uint32_t cpuprof_writers;
static int32_t cpuprof_hz;

// Process CPU time at start and stop, CPU time in samples, and CPU time of
// the sampler thread, in nanoseconds.
static uint64_t cpuprof_start_ns;
static uint64_t cpuprof_sampled_ns;
static uint64_t cpuprof_sampler_ns;
static uint64_t cpuprof_stop_ns;

static struct cpuprof_entry *cpuprof_lookup(int table, uint64_t id, int insert) {
    struct cpuprof_entry *t = cpuprof_tables[table];
    uint32_t i = (uint32_t)(id * 0x9e3779b97f4a7c15ull >> 52) % CPUPROF_MAX_THREADS;
    for (uint32_t probe = 0; probe < CPUPROF_MAX_THREADS; probe++) {
        if (t[i].id == id) {
            return &t[i];
        }
        if (t[i].id == 0) {
            if (!insert) {
                return NULL;
            }
            t[i].id = id;
            t[i].cpu_ns = 0;
            cpuprof_used[table][cpuprof_nused[table]++] = i;
            return &t[i];
        }
        i = (i + 1) % CPUPROF_MAX_THREADS;
    }
    return NULL;
}

static void cpuprof_clear_table(int table) {
    for (uint32_t i = 0; i < cpuprof_nused[table]; i++) {
        cpuprof_tables[table][cpuprof_used[table][i]].id = 0;
    }
    cpuprof_nused[table] = 0;
}

// The newest CPU time recorded for a thread, or 0 if it was not seen.
static uint64_t cpuprof_last_locked(uint64_t id) {
    struct cpuprof_entry *e = cpuprof_lookup(1 - cpuprof_prev, id, 0);
    if (e == NULL) {
        e = cpuprof_lookup(cpuprof_prev, id, 0);
    }
    return e != NULL ? e->cpu_ns : 0;
}

// Record the CPU time of a thread and return the CPU time since it was last
// recorded, or 0 if the table is full. A thread not seen before is new, so all
// its time counts.
static uint64_t cpuprof_update(uint64_t id, uint64_t cpu_ns) {
    pthread_mutex_lock(&cpuprof_table_lock);
    uint64_t last = cpuprof_last_locked(id);
    struct cpuprof_entry *now = cpuprof_lookup(1 - cpuprof_prev, id, 1);
    if (now != NULL) {
        // Another thread may have stored a newer time since cpu_ns was read.
        now->cpu_ns = cpu_ns > last ? cpu_ns : last;
    }
    pthread_mutex_unlock(&cpuprof_table_lock);
    return now != NULL && cpu_ns > last ? cpu_ns - last : 0;
}

// Claim a sample slot, or return NULL. Exiting threads write samples too.
static uintptr_t *cpuprof_claim(void) {
    uint32_t slot = __atomic_fetch_add(&cpuprof_next, 1, __ATOMIC_RELAXED);
    if (slot >= cpuprof_slots) {
        __atomic_fetch_add(&cpuprof_dropped, 1, __ATOMIC_RELAXED);
        return NULL;
    }
    // The buffer is reused, so clear the length in case the sample is not
    // completed.
    uintptr_t *rec = cpuprof_mem + (uintptr_t)slot * CPUPROF_SLOT_WORDS;
    rec[0] = 0;
    return rec;
}

static uint64_t cpuprof_time_ns(struct cpuprof_time_value user, struct cpuprof_time_value sys) {
    return ((uint64_t)user.seconds + sys.seconds) * 1000000000ull +
           ((uint64_t)user.microseconds + sys.microseconds) * 1000ull;
}

static uint64_t cpuprof_process_ns(void) {
    struct rusage ru;
    if (getrusage(RUSAGE_SELF, &ru) != 0) {
        return 0;
    }
    return ((uint64_t)ru.ru_utime.tv_sec + ru.ru_stime.tv_sec) * 1000000000ull +
           ((uint64_t)ru.ru_utime.tv_usec + ru.ru_stime.tv_usec) * 1000ull;
}

// Sample a thread if it used CPU time since the last wakeup. With record set
// to 0 only note its CPU time.
static void cpuprof_sample_thread(cpuprof_port_t thread, int record) {
    struct cpuprof_thread_identifier_info ident;
    uint32_t count = sizeof(ident) / sizeof(uint32_t);
    if (thread_info(thread, CPUPROF_THREAD_IDENTIFIER_INFO, (int32_t *)&ident, &count) != 0) {
        return;
    }
    struct cpuprof_thread_basic_info basic;
    count = sizeof(basic) / sizeof(uint32_t);
    if (thread_info(thread, CPUPROF_THREAD_BASIC_INFO, (int32_t *)&basic, &count) != 0) {
        return;
    }
    uint64_t delta = cpuprof_update(ident.thread_id, cpuprof_time_ns(basic.user_time, basic.system_time));
    if (!record || delta == 0) {
        return;
    }
    uintptr_t *rec = cpuprof_claim();
    if (rec == NULL) {
        return;
    }

    // This takes the pthread list lock, so do it before suspending. Nothing
    // that could take a lock the thread holds is called while it is suspended.
    uintptr_t lo = 0, hi = 0;
    pthread_t pt = pthread_from_mach_thread_np(thread);
    if (pt != NULL) {
        hi = (uintptr_t)pthread_get_stackaddr_np(pt);
        lo = hi - pthread_get_stacksize_np(pt);
    }

    if (thread_suspend(thread) != 0) {
        return;
    }
    struct cpuprof_thread_state state;
    count = sizeof(state) / sizeof(uint32_t);
    if (thread_get_state(thread, CPUPROF_THREAD_STATE, (uint32_t *)&state, &count) != 0) {
        thread_resume(thread);
        return;
    }
    uintptr_t *pcs = rec + 2;
    uintptr_t n = 0;
    uintptr_t fp = CPUPROF_FP(state);
    uintptr_t sp = CPUPROF_SP(state);
    if (sp > lo) {
        lo = sp;
    }
    pcs[n++] = CPUPROF_PC(state) & CPUPROF_ADDR_MASK;
    while (n < CPUPROF_MAX_DEPTH && fp >= lo && fp + 2 * sizeof(uintptr_t) <= hi &&
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
    thread_resume(thread);
    rec[1] = delta;
    __atomic_fetch_add(&cpuprof_sampled_ns, delta, __ATOMIC_RELAXED);
    __atomic_store_n(&rec[0], n, __ATOMIC_RELEASE);
}

static void cpuprof_record_exit(uintptr_t fn) {
    cpuprof_port_t self = mach_thread_self();
    struct cpuprof_thread_identifier_info ident;
    struct cpuprof_thread_basic_info basic;
    uint32_t count = sizeof(ident) / sizeof(uint32_t);
    int ok = thread_info(self, CPUPROF_THREAD_IDENTIFIER_INFO, (int32_t *)&ident, &count) == 0;
    count = sizeof(basic) / sizeof(uint32_t);
    ok = ok && thread_info(self, CPUPROF_THREAD_BASIC_INFO, (int32_t *)&basic, &count) == 0;
    mach_port_deallocate(mach_task_self_, self);
    if (!ok) {
        return;
    }
    uint64_t delta = cpuprof_update(ident.thread_id, cpuprof_time_ns(basic.user_time, basic.system_time));
    uintptr_t *rec = delta > 0 ? cpuprof_claim() : NULL;
    if (rec == NULL) {
        return;
    }
    rec[2] = fn;
    rec[1] = delta;
    __atomic_fetch_add(&cpuprof_sampled_ns, delta, __ATOMIC_RELAXED);
    __atomic_store_n(&rec[0], 1, __ATOMIC_RELEASE);
}

// Called by every goroutine thread before it exits. Its CPU time since the
// last wakeup goes to its entry function, so short threads are still counted.
// The main thread passes 0 and is not recorded.
void tinygo_cpuprof_thread_exit(uintptr_t fn) {
    if (fn == 0) {
        return;
    }
    // Register as a writer before checking the state, so that stop either
    // sees this thread or this thread sees that profiling stopped.
    __atomic_fetch_add(&cpuprof_writers, 1, __ATOMIC_SEQ_CST);
    if (__atomic_load_n(&cpuprof_state, __ATOMIC_SEQ_CST) == CPUPROF_RECORDING) {
        cpuprof_record_exit(fn);
    }
    __atomic_fetch_sub(&cpuprof_writers, 1, __ATOMIC_RELEASE);
}

static void cpuprof_sample_all(int record) {
    cpuprof_port_t task = mach_task_self_;
    cpuprof_port_t self = mach_thread_self();
    cpuprof_port_t *threads;
    uint32_t count;
    if (task_threads(task, &threads, &count) == 0) {
        for (uint32_t i = 0; i < count; i++) {
            if (threads[i] != self) {
                cpuprof_sample_thread(threads[i], record);
            }
            mach_port_deallocate(task, threads[i]);
        }
        vm_deallocate(task, (uintptr_t)threads, count * sizeof(*threads));
    }
    pthread_mutex_lock(&cpuprof_table_lock);
    cpuprof_clear_table(cpuprof_prev);
    cpuprof_prev = 1 - cpuprof_prev;
    pthread_mutex_unlock(&cpuprof_table_lock);

    struct cpuprof_thread_basic_info basic;
    uint32_t n = sizeof(basic) / sizeof(uint32_t);
    if (thread_info(self, CPUPROF_THREAD_BASIC_INFO, (int32_t *)&basic, &n) == 0) {
        cpuprof_sampler_ns = cpuprof_time_ns(basic.user_time, basic.system_time);
    }
    mach_port_deallocate(task, self);
}

static void *cpuprof_loop(void *arg) {
    cpuprof_sample_all(0);
    int starting = CPUPROF_STARTING;
    __atomic_compare_exchange_n(&cpuprof_state, &starting, CPUPROF_RECORDING, 0,
                                __ATOMIC_SEQ_CST, __ATOMIC_SEQ_CST);
    struct timespec interval = { 0, 1000000000 / cpuprof_hz };
    while (__atomic_load_n(&cpuprof_state, __ATOMIC_ACQUIRE) != CPUPROF_OFF) {
        nanosleep(&interval, NULL);
        cpuprof_sample_all(1);
    }
    return NULL;
}

// Start sampling hz times per second with room for slots samples. Returns 0
// on failure.
int tinygo_cpuprof_start(int32_t hz, uint32_t slots) {
    size_t bytes = (size_t)slots * CPUPROF_SLOT_WORDS * sizeof(uintptr_t);
    if (cpuprof_mem_bytes < bytes) {
        if (cpuprof_mem != NULL) {
            munmap(cpuprof_mem, cpuprof_mem_bytes);
        }
        void *p = mmap(NULL, bytes, PROT_READ | PROT_WRITE, MAP_PRIVATE | MAP_ANON, -1, 0);
        if (p == MAP_FAILED) {
            cpuprof_mem = NULL;
            cpuprof_mem_bytes = 0;
            return 0;
        }
        cpuprof_mem = p;
        cpuprof_mem_bytes = bytes;
    }
    cpuprof_slots = slots;
    __atomic_store_n(&cpuprof_next, 0, __ATOMIC_RELAXED);
    __atomic_store_n(&cpuprof_dropped, 0, __ATOMIC_RELAXED);
    cpuprof_hz = hz;
    cpuprof_clear_table(0);
    cpuprof_clear_table(1);
    __atomic_store_n(&cpuprof_sampled_ns, 0, __ATOMIC_RELAXED);
    cpuprof_sampler_ns = 0;
    cpuprof_start_ns = cpuprof_process_ns();
    __atomic_store_n(&cpuprof_state, CPUPROF_STARTING, __ATOMIC_SEQ_CST);
    if (pthread_create(&cpuprof_sampler, NULL, cpuprof_loop, NULL) != 0) {
        __atomic_store_n(&cpuprof_state, CPUPROF_OFF, __ATOMIC_SEQ_CST);
        return 0;
    }
    return 1;
}

void tinygo_cpuprof_stop(void) {
    __atomic_store_n(&cpuprof_state, CPUPROF_OFF, __ATOMIC_SEQ_CST);
    pthread_join(cpuprof_sampler, NULL);
    // Wait for exiting threads that saw profiling still running.
    while (__atomic_load_n(&cpuprof_writers, __ATOMIC_ACQUIRE) != 0) {
        sched_yield();
    }
    cpuprof_stop_ns = cpuprof_process_ns();
}

uintptr_t *tinygo_cpuprof_samples(uint32_t *used, uint32_t *dropped, uint32_t *slotWords) {
    uint32_t next = __atomic_load_n(&cpuprof_next, __ATOMIC_RELAXED);
    *used = next < cpuprof_slots ? next : cpuprof_slots;
    *dropped = __atomic_load_n(&cpuprof_dropped, __ATOMIC_RELAXED);
    *slotWords = CPUPROF_SLOT_WORDS;
    return cpuprof_mem;
}

// Return CPU time no sample covers, such as threads that ran between two
// wakeups, and the CPU time of the sampler thread.
void tinygo_cpuprof_unsampled(uint64_t *unattributed_ns, uint64_t *sampler_ns) {
    uint64_t total = cpuprof_stop_ns > cpuprof_start_ns ? cpuprof_stop_ns - cpuprof_start_ns : 0;
    uint64_t covered = __atomic_load_n(&cpuprof_sampled_ns, __ATOMIC_RELAXED) + cpuprof_sampler_ns;
    *unattributed_ns = total > covered ? total - covered : 0;
    *sampler_ns = cpuprof_sampler_ns;
}

// From <mach-o/loader.h>, which the minimal SDK does not have.
// https://github.com/apple-oss-distributions/xnu/blob/main/EXTERNAL_HEADERS/mach-o/loader.h
struct cpuprof_mach_header_64 {
    uint32_t magic, cputype, cpusubtype, filetype, ncmds, sizeofcmds, flags, reserved;
};
struct cpuprof_load_command {
    uint32_t cmd, cmdsize;
};
struct cpuprof_segment_command_64 {
    uint32_t cmd, cmdsize;
    char segname[16];
    uint64_t vmaddr, vmsize;
};
#define CPUPROF_LC_SEGMENT_64 0x19

// From <mach-o/dyld.h>, which the minimal SDK does not have.
int _NSGetExecutablePath(char *buf, uint32_t *bufsize);

// Defined by the linker at the start of the executable's __TEXT segment.
extern const struct cpuprof_mach_header_64 _mh_execute_header;

// Return the address range of the executable's __TEXT segment after ASLR.
void tinygo_cpuprof_text(uintptr_t *start, uintptr_t *end) {
    const struct cpuprof_mach_header_64 *mh = &_mh_execute_header;
    *start = (uintptr_t)mh;
    *end = *start;
    const char *p = (const char *)(mh + 1);
    for (uint32_t i = 0; i < mh->ncmds; i++) {
        const struct cpuprof_load_command *lc = (const void *)p;
        if (lc->cmd == CPUPROF_LC_SEGMENT_64) {
            const struct cpuprof_segment_command_64 *seg = (const void *)p;
            if (strncmp(seg->segname, "__TEXT", 16) == 0) {
                *end = *start + seg->vmsize;
                return;
            }
        }
        p += lc->cmdsize;
    }
}

// Write the path of the executable to buf. Returns its length or 0.
uint32_t tinygo_cpuprof_executable(char *buf, uint32_t size) {
    if (_NSGetExecutablePath(buf, &size) != 0) {
        return 0;
    }
    return strlen(buf);
}

// From <dlfcn.h>, which the minimal SDK does not have.
typedef struct {
    const char *dli_fname;
    void *dli_fbase;
    const char *dli_sname;
    void *dli_saddr;
} cpuprof_dl_info;
int dladdr(const void *addr, cpuprof_dl_info *info);

// Look up the library and symbol containing pc. Returns 0 if not found.
int tinygo_cpuprof_dladdr(uintptr_t pc, const char **lib, uintptr_t *base, const char **sym) {
    cpuprof_dl_info info = { 0 };
    if (dladdr((const void *)pc, &info) == 0) {
        return 0;
    }
    *lib = info.dli_fname;
    *base = (uintptr_t)info.dli_fbase;
    *sym = info.dli_sname;
    return 1;
}
