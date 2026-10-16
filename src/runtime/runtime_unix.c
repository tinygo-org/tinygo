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

// tinygo_sigpanic is defined in Go. It turns a signal into a Go panic
// that can be recovered with recover(). The signal number is passed via
// the tinygo_caught_signal global.
void tinygo_sigpanic(void);

// Set by the signal handler before redirecting to tinygo_sigpanic.
int tinygo_caught_signal;

// Whether sigpanic-based recovery is supported for the current
// architecture. Set to 0 on architectures where we can't reliably
// modify the ucontext to redirect execution.
static int can_sigpanic(void) {
#if defined(__x86_64__) || defined(__i386__) || defined(__aarch64__) || defined(__arm64__) || defined(__arm__) || defined(__mips__)
	return 1;
#else
	return 0;
#endif
}

// Try to redirect execution from the signal handler to tinygo_sigpanic.
// Returns 1 on success, 0 if the architecture doesn't support it.
static int redirect_to_sigpanic(int sig, ucontext_t *uctx) {
	if (!can_sigpanic()) {
		return 0;
	}
	tinygo_caught_signal = sig;

#if __APPLE__
	#if __arm64__
		// ARM64: set LR to the faulting PC (as return address), set PC to sigpanic
		uctx->uc_mcontext->__ss.__lr = uctx->uc_mcontext->__ss.__pc;
		uctx->uc_mcontext->__ss.__pc = (uint64_t)&tinygo_sigpanic;
	#elif __x86_64__
		// x86_64: push the faulting PC onto the stack, set RIP to sigpanic
		uintptr_t sp = uctx->uc_mcontext->__ss.__rsp;
		sp -= sizeof(uintptr_t);
		*(uintptr_t *)sp = uctx->uc_mcontext->__ss.__rip;
		uctx->uc_mcontext->__ss.__rsp = sp;
		uctx->uc_mcontext->__ss.__rip = (uint64_t)&tinygo_sigpanic;
	#else
		return 0;
	#endif
#elif __linux__
	#if __x86_64__
		uintptr_t sp = uctx->uc_mcontext.gregs[REG_RSP];
		sp -= sizeof(uintptr_t);
		*(uintptr_t *)sp = uctx->uc_mcontext.gregs[REG_RIP];
		uctx->uc_mcontext.gregs[REG_RSP] = sp;
		uctx->uc_mcontext.gregs[REG_RIP] = (uintptr_t)&tinygo_sigpanic;
	#elif __i386__
		uintptr_t sp = uctx->uc_mcontext.gregs[REG_ESP];
		sp -= sizeof(uintptr_t);
		*(uintptr_t *)sp = uctx->uc_mcontext.gregs[REG_EIP];
		uctx->uc_mcontext.gregs[REG_ESP] = sp;
		uctx->uc_mcontext.gregs[REG_EIP] = (uintptr_t)&tinygo_sigpanic;
	#elif __aarch64__
		uctx->uc_mcontext.regs[30] = uctx->uc_mcontext.pc; // LR = faulting PC
		uctx->uc_mcontext.pc = (uintptr_t)&tinygo_sigpanic;
	#elif __arm__
		uctx->uc_mcontext.arm_lr = uctx->uc_mcontext.arm_pc;
		uctx->uc_mcontext.arm_pc = (uintptr_t)&tinygo_sigpanic;
	#elif defined(__mips__)
		// MIPS: set RA (gregs[31]) to the faulting PC, set PC to sigpanic.
		uctx->uc_mcontext.gregs[31] = uctx->uc_mcontext.pc;
		uctx->uc_mcontext.pc = (uintptr_t)&tinygo_sigpanic;
	#else
		return 0;
	#endif
#else
	return 0;
#endif

	return 1;
}

static void signal_handler(int sig, siginfo_t *info, void *context) {
	ucontext_t* uctx = context;

	// Try to redirect to sigpanic for a recoverable panic.
	if (redirect_to_sigpanic(sig, uctx)) {
		// Re-register the signal handler since SA_RESETHAND cleared it.
		// We need it active in case the sigpanic itself faults (e.g.,
		// stack overflow during panic).
		struct sigaction act;
		memset(&act, 0, sizeof(act));
		act.sa_flags = SA_SIGINFO | SA_RESETHAND;
		act.sa_sigaction = &signal_handler;
		sigaction(sig, &act, NULL);
		return; // return from signal handler; execution resumes at sigpanic
	}

	// Fallback: extract the faulting address and call the fatal handler.
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
	sigaction(SIGFPE, &act, NULL);
	sigaction(SIGILL, &act, NULL);
	sigaction(SIGSEGV, &act, NULL);
}

#ifdef __linux__
#include <poll.h>
#include <sys/eventfd.h>
#include <sys/syscall.h>
#include <time.h>

int32_t tinygo_poll_eventfd(void) {
	return (int32_t)eventfd(0, EFD_NONBLOCK | EFD_CLOEXEC);
}

void tinygo_poll_wake(int32_t wakefd) {
	uint64_t val = 1;
	write(wakefd, &val, sizeof(val));
}

void tinygo_poll_close(int32_t wakefd) {
	close(wakefd);
}

// Returns: 0 = fd ready, 1 = woken via wakefd, 2 = timeout, -1 = error/EINTR
int32_t tinygo_poll_wait(int32_t fd, int32_t wakefd, int32_t mode, int64_t timeout_ns) {
	struct pollfd fds[2];
	fds[0].fd = fd;
	fds[0].events = (mode == 'w') ? POLLOUT : (POLLIN | POLLRDHUP);
	fds[0].revents = 0;
	fds[1].fd = wakefd;
	fds[1].events = POLLIN;
	fds[1].revents = 0;

	struct timespec ts;
	struct timespec *pts = NULL;
	if (timeout_ns >= 0) {
		ts.tv_sec = timeout_ns / 1000000000LL;
		ts.tv_nsec = timeout_ns % 1000000000LL;
		pts = &ts;
	}
#ifdef SYS_ppoll_time64
	long r = syscall(SYS_ppoll_time64, fds, 2, pts, NULL, _NSIG / 8);
#else
	long r = syscall(SYS_ppoll, fds, 2, pts, NULL, _NSIG / 8);
#endif
	if (r < 0) {
		return -1;
	}
	if (r == 0) {
		return 2;
	}
	if (fds[1].revents != 0) {
		uint64_t val;
		while (read(wakefd, &val, sizeof(val)) > 0) {
		}
		return 1;
	}
	if (fds[0].revents != 0) {
		return 0;
	}
	return -1;
}
#endif
