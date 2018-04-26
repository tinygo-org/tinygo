
#include "runtime.h"
#include <string.h>

void go_main() __asm__("main.main");

int main() {
	go_main();

	return 0;
}

__attribute__((weak))
void * memset(void *s, int c, size_t n) {
	for (size_t i = 0; i < n; i++) {
		((uint8_t*)s)[i] = c;
	}
	return s;
}
