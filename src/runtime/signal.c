//go:build none

// Ignore the //go:build above. This file is manually included on Linux and
// MacOS to provide os/signal support.

#include <stdint.h>
#include <signal.h>
#include <time.h>
#include <unistd.h>

// Signal handler in the runtime.
void tinygo_signal_handler(int sig);

static uint32_t tinygo_handling_sig;

// Enable a signal from the runtime.
void tinygo_signal_enable(uint32_t sig) {
    struct sigaction act = { 0 };
    act.sa_handler = &tinygo_signal_handler;
    act.sa_flags = SA_RESTART;
    __atomic_fetch_or(&tinygo_handling_sig, 1U << sig, __ATOMIC_RELAXED);
    sigaction(sig, &act, NULL);
}

void tinygo_signal_ignore(uint32_t sig) {
    struct sigaction act = { 0 };
    act.sa_handler = SIG_IGN;
    __atomic_fetch_and(&tinygo_handling_sig, ~(1U << sig), __ATOMIC_RELAXED);
    sigaction(sig, &act, NULL);
}

void tinygo_signal_disable(uint32_t sig) {
    struct sigaction act = { 0 };
    act.sa_handler = SIG_DFL;
    __atomic_fetch_and(&tinygo_handling_sig, ~(1U << sig), __ATOMIC_RELAXED);
    sigaction(sig, &act, NULL);
}

void tinygo_signal_clear_handlers(void) {
    uint32_t mask = __atomic_load_n(&tinygo_handling_sig, __ATOMIC_RELAXED);
    struct sigaction act = { 0 };
    act.sa_handler = SIG_DFL;
    for (uint32_t sig = 1; sig < 32; sig++) {
        if (mask & (1U << sig)) {
            sigaction(sig, &act, NULL);
        }
    }
}
