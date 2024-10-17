//go:build none

// Ignore the //go:build above. This file is manually included on Linux and
// MacOS to provide os/signal support.

#include <stdint.h>
#include <signal.h>
#include <time.h>
#include <unistd.h>

// Signal handler in the runtime.
void tinygo_signal_handler(int sig);

// Enable a signal from the runtime.
void tinygo_signal_enable(uint32_t sig) {
    struct sigaction act = { 0 };
    act.sa_handler = &tinygo_signal_handler;
    sigaction(sig, &act, NULL);
}

void tinygo_signal_ignore(uint32_t sig) {
    struct sigaction act = { 0 };
    act.sa_handler = SIG_IGN;
    sigaction(sig, &act, NULL);
}

void tinygo_signal_disable(uint32_t sig) {
    struct sigaction act = { 0 };
    act.sa_handler = SIG_DFL;
    sigaction(sig, &act, NULL);
}

// Implement waitForEvents and sleep with signals.
// Warning: sigprocmask is not defined in a multithreaded program so will need
// to be replaced with something else once we implement threading on POSIX.

// Signals active before a call to tinygo_wfi_mask.
static sigset_t active_signals;

static void tinygo_set_signals(sigset_t *mask, uint32_t signals) {
    sigemptyset(mask);
    for (int i=0; i<32; i++) {
        if ((signals & (1<<i)) != 0) {
            sigaddset(mask, i);
        }
    }
}

// Mask the given signals.
// This function must always restore the previous signals using
// tinygo_wfi_unmask, to create a critical section.
void tinygo_wfi_mask(uint32_t active) {
    sigset_t mask;
    tinygo_set_signals(&mask, active);

    sigprocmask(SIG_BLOCK, &mask, &active_signals);
}

// Wait until a signal becomes pending (or is already pending), and return the
// signal.
#if !defined(__APPLE__)
int tinygo_wfi_sleep(uint32_t active, uint64_t timeout) {
    sigset_t active_set;
    tinygo_set_signals(&active_set, active);

    struct timespec ts = {0};
    ts.tv_sec  = timeout / 1000000000;
    ts.tv_nsec = timeout % 1000000000;

    int result = sigtimedwait(&active_set, NULL, &ts);
    return result;
}
#endif

// Wait until any of the active signals becomes pending (or returns immediately
// if one is already pending).
int tinygo_wfi_wait(uint32_t active) {
    sigset_t active_set;
    tinygo_set_signals(&active_set, active);

    int sig = 0;
    sigwait(&active_set, &sig);
    return sig;
}

// Restore previous signal mask.
void tinygo_wfi_unmask(void) {
    sigprocmask(SIG_SETMASK, &active_signals, NULL);
}
