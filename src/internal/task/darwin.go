//go:build darwin

package task

import "unsafe"

// MacOS uses a pointer so unsafe.Pointer should be fine:
//
//	typedef struct _opaque_pthread_t *__darwin_pthread_t;
//	typedef __darwin_pthread_t pthread_t;
type threadID unsafe.Pointer
