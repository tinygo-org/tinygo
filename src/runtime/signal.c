//go:build none

// Ignore the //go:build above. This file is manually included on Linux and
// MacOS to provide os/signal support.

// see man (7) feature_test_macros
// _XOPEN_SOURCE is needed  to control the definitions that are exposed by
// system header files when a program is compiled
#define _XOPEN_SOURCE 700
#include <stdint.h>
#include <signal.h>
#include <unistd.h>

// Signal handler in the runtime.
void tinygo_signal_handler(int sig);

// Enable a signal from the runtime.
void tinygo_signal_enable(uint32_t sig)
{
    struct sigaction act = {0};
    act.sa_flags = SA_SIGINFO;
    act.sa_handler = &tinygo_signal_handler;
    sigaction(sig, &act, NULL);
}

// Disable a signal from the runtime.
// Defer the signal handler to the default handler for now.
void tinygo_signal_disable(uint32_t sig)
{
    struct sigaction act = {0};
    act.sa_handler = SIG_DFL;
    sigaction(sig, &act, NULL);
}

// Ignore a signal from the runtime.
void tinygo_signal_ignore(uint32_t sig)
{
    struct sigaction act = {0};
    act.sa_handler = SIG_IGN;
    sigaction(sig, &act, NULL);
}