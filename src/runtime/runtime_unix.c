//go:build none

// This file is included on Darwin and Linux (despite the //go:build line above).

#define _GNU_SOURCE
#define _XOPEN_SOURCE
#include <signal.h>
#include <unistd.h>
#include <stdint.h>
#include <ucontext.h>
#include <string.h>

void tinygo_handle_fatal_signal(int sig, uintptr_t addr);

static void signal_handler(int sig, siginfo_t *info, void *context) {
	ucontext_t* uctx = context;
	uintptr_t addr = 0;
	#if __APPLE__
		#if __arm64__
			addr = uctx->uc_mcontext->__ss.__pc;
		#elif __x86_64__
			addr = uctx->uc_mcontext->__ss.__rip;
		#else
			#error unknown architecture
		#endif
	#elif __linux__
		// Note: this can probably be simplified using the MC_PC macro in musl,
		// but this works for now.
		#if __arm__
			addr = uctx->uc_mcontext.arm_pc;
		#elif __i386__
			addr = uctx->uc_mcontext.gregs[REG_EIP];
		#elif __x86_64__
			addr = uctx->uc_mcontext.gregs[REG_RIP];
		#else // aarch64, mips, maybe others
			addr = uctx->uc_mcontext.pc;
		#endif
	#else
		#error unknown platform
	#endif
	tinygo_handle_fatal_signal(sig, addr);
}

void tinygo_register_fatal_signals(void) {
	struct sigaction act = { 0 };
	// SA_SIGINFO:   we want the 2 extra parameters
	// SA_RESETHAND: only catch the signal once (the handler will re-raise the signal)
	act.sa_flags = SA_SIGINFO | SA_RESETHAND;
	act.sa_sigaction = &signal_handler;

	// Register the signal handler for common issues. There are more signals,
	// which can be added if needed.
	sigaction(SIGBUS, &act, NULL);
	sigaction(SIGILL, &act, NULL);
	sigaction(SIGSEGV, &act, NULL);
}
