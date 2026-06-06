//go:build none

// This file is included on Windows (despite the //go:build line above).

#include <windows.h>
#include <stdint.h>

void tinygo_sigpanic_windows(int32_t exception_code);

static LONG WINAPI tinygo_exception_handler(EXCEPTION_POINTERS *info) {
	DWORD code = info->ExceptionRecord->ExceptionCode;
	switch (code) {
	case EXCEPTION_ACCESS_VIOLATION:
	case EXCEPTION_IN_PAGE_ERROR:
	case EXCEPTION_INT_DIVIDE_BY_ZERO:
	case EXCEPTION_INT_OVERFLOW:
		tinygo_sigpanic_windows((int32_t)code);
		// If runtimePanic triggers longjmp, we never reach here.
		// If it doesn't (no defer frame), it will abort and we also
		// never reach here.
		return EXCEPTION_CONTINUE_SEARCH;
	default:
		return EXCEPTION_CONTINUE_SEARCH;
	}
}

void tinygo_init_exception_handler(void) {
	AddVectoredExceptionHandler(1, tinygo_exception_handler);
}
