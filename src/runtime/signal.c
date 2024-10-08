//go:build none

// Ignore the //go:build above. This file is manually included on Linux and
// MacOS to provide os/signal support.

#include <stdint.h>
#include <signal.h>
#include <unistd.h>

// Signal handler in the runtime.
void tinygo_signal_handler(int sig);

// Enable a signal from the runtime.
void tinygo_signal_enable(uint32_t sig)
{
    struct sigaction act = {0};
    act.sa_handler = &tinygo_signal_handler;
    sigaction(sig, &act, NULL);
}

void tinygo_signal_ignore(uint32_t sig)
{
    struct sigaction act = {0};
    act.sa_handler = SIG_IGN;
    sigaction(sig, &act, NULL);
}

void tinygo_signal_disable(uint32_t sig)
{
    struct sigaction act = {0};
    act.sa_handler = SIG_DFL;
    sigaction(sig, &act, NULL);
}

// https://www.cipht.net/2023/11/30/perils-of-pause.html#text-3
void tinygo_mask_signals(uint32_t mask)
{
    sigset_t set, prev;
    sigemptyset(&set);
    for (int i = 0; i < 32; i++)
    {
        if (mask & (1 << i))
        {
            sigaddset(&set, i);
        }
    }
    sigprocmask(SIG_BLOCK, &set, &prev);
}

void tinygo_unmask_signals(uint32_t mask)
{
    sigset_t set, prev;
    sigemptyset(&set);
    for (int i = 0; i < 32; i++)
    {
        if (mask & (1 << i))
        {
            sigaddset(&set, i);
        }
    }
    sigprocmask(SIG_UNBLOCK, &set, &prev);
}

void tinygo_sigsuspend(uint32_t mask)
{
    sigset_t set;
    sigemptyset(&set);
    for (int i = 0; i < 32; i++)
    {
        if (mask & (1 << i))
        {
            sigaddset(&set, i);
        }
    }
    sigsuspend(&set);
}
