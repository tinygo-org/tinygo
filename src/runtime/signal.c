//go:build none

// Ignore the //go:build above. This file is manually included on Linux and
// MacOS to provide os/signal support.

#include <stdint.h>
#include <signal.h>
#include <time.h>
#include <unistd.h>

// Signal handler in the runtime.
void tinygo_signal_handler(int sig);

// sigpipe_wanted is set while the program has asked for SIGPIPE through
// os/signal, by Notify or Ignore.
static volatile sig_atomic_t sigpipe_wanted;

// Enable a signal from the runtime.
void tinygo_signal_enable(uint32_t sig) {
    if (sig == SIGPIPE) sigpipe_wanted = 1;
    struct sigaction act = { 0 };
    act.sa_handler = &tinygo_signal_handler;
    act.sa_flags = SA_RESTART;
    sigaction(sig, &act, NULL);
}

void tinygo_signal_ignore(uint32_t sig) {
    if (sig == SIGPIPE) sigpipe_wanted = 1;
    struct sigaction act = { 0 };
    act.sa_handler = SIG_IGN;
    sigaction(sig, &act, NULL);
}

void tinygo_signal_disable(uint32_t sig) {
    if (sig == SIGPIPE) sigpipe_wanted = 0;
    struct sigaction act = { 0 };
    // SIGPIPE goes back to ignored, which is its state at start.
    act.sa_handler = sig == SIGPIPE ? SIG_IGN : SIG_DFL;
    sigaction(sig, &act, NULL);
}

// Die by SIGPIPE after a failed write to stdout or stderr, as Go does, unless
// the program asked for the signal.
void tinygo_sigpipe(void) {
    if (sigpipe_wanted) return;
    struct sigaction act = { 0 };
    act.sa_handler = SIG_DFL;
    sigaction(SIGPIPE, &act, NULL);
    sigset_t set;
    sigemptyset(&set);
    sigaddset(&set, SIGPIPE);
    sigprocmask(SIG_UNBLOCK, &set, NULL);
    raise(SIGPIPE);
}
