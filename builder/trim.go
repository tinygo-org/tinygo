package builder

/*
#include <stdlib.h>
#ifdef __GLIBC__
#include <malloc.h>
static void tinygo_release_heap(void) { malloc_trim(0); }
#else
static void tinygo_release_heap(void) {}
#endif
*/
import "C"

// releaseCHeap returns freed LLVM memory to the OS so that external tools
// started afterwards, such as wasm-opt, do not compete with it.
func releaseCHeap() { C.tinygo_release_heap() }
