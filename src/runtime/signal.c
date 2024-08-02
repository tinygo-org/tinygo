//go:build none

// Ignore the //go:build above. This file is manually included on Linux and
// MacOS to provide os/signal support.

#include <stdint.h>
#include <signal.h>
#include <unistd.h>

// Signal handler in the runtime.
void tinygo_signal_handler(int sig);

// Enable a signal from the runtime.
void tinygo_signal_enable(uint32_t sig) {
    struct sigaction act = { 0 };
    act.sa_flags = SA_SIGINFO;
    act.sa_handler = &tinygo_signal_handler;
    sigaction(sig, &act, NULL);
}
